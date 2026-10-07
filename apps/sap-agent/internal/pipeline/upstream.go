package pipeline

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/nembus"
	"github.com/NasTecSol/nembus-sap-agent/internal/sap"
)

type UpstreamSync struct {
	cfg          *config.Config
	sapClient    *sap.Client
	nembusClient *nembus.Client
}

func NewUpstreamSync(cfg *config.Config, sapClient *sap.Client, nembusClient *nembus.Client) *UpstreamSync {
	return &UpstreamSync{
		cfg:          cfg,
		sapClient:    sapClient,
		nembusClient: nembusClient,
	}
}

type UpstreamStats struct {
	TransactionsPosted int
	PaymentsPosted     int
	Errors             int
	Duration           time.Duration
}

func (s *UpstreamSync) ProcessOutbox(ctx context.Context) (*UpstreamStats, error) {
	start := time.Now()
	stats := &UpstreamStats{}

	// 1. Fetch unposted completed POS transactions from Nembus
	limit := s.cfg.BatchSize
	if limit <= 0 {
		limit = 50
	}

	txs, err := s.nembusClient.FetchUnpostedPOSTransactions(ctx, limit)
	if err != nil {
		return stats, fmt.Errorf("failed to fetch unposted transactions: %w", err)
	}

	if len(txs) == 0 {
		return stats, nil
	}

	log.Printf("⬆️ [Nembus POS -> SAP] Found %d unposted transaction(s). Processing...", len(txs))

	for _, tx := range txs {
		if err := s.postTransactionToSAP(ctx, &tx); err != nil {
			log.Printf("❌ Failed to post transaction %s (ID: %d) to SAP: %v", tx.TransactionNumber, tx.ID, err)
			stats.Errors++
		} else {
			stats.TransactionsPosted++
		}
	}

	// 2. Also check and resolve pending sync_queue records
	queueItems, qErr := s.nembusClient.FetchPendingSyncQueue(ctx, limit)
	if qErr == nil {
		for _, item := range queueItems {
			if item.ID > 0 {
				_ = s.nembusClient.UpdateSyncQueueStatus(ctx, item.ID, "synced", "")
			}
		}
	}

	stats.Duration = time.Since(start)
	return stats, nil
}

func (s *UpstreamSync) postTransactionToSAP(ctx context.Context, tx *nembus.POSTransaction) error {
	// 1. Fetch lines
	lines, err := s.nembusClient.FetchTransactionLines(ctx, tx.ID)
	if err != nil {
		return fmt.Errorf("failed fetching lines: %w", err)
	}

	if len(lines) == 0 {
		return fmt.Errorf("transaction %s has no line items", tx.TransactionNumber)
	}

	// 2. Build SAP Document Lines
	var docLines []sap.SAPDocumentLine
	for idx, line := range lines {
		// Calculate unit price net of tax if not already net
		unitPriceNet := line.UnitPrice
		if line.TaxAmount > 0 && line.Quantity > 0 {
			// Subtotal is net before tax
			unitPriceNet = line.Subtotal / line.Quantity
		}

		// Calculate discount percent from discount amount
		discountPct := 0.0
		if line.UnitPrice > 0 && line.Quantity > 0 && line.DiscountAmount > 0 {
			gross := line.UnitPrice * line.Quantity
			discountPct = (line.DiscountAmount / gross) * 100.0
		}

		lineNum := idx
		docLines = append(docLines, sap.SAPDocumentLine{
			LineNum:         &lineNum,
			ItemCode:        line.ProductSKU,
			ItemDescription: line.ProductName,
			Quantity:        line.Quantity,
			UnitPrice:       unitPriceNet,
			PriceAfterVAT:   line.LineTotal / line.Quantity,
			DiscountPercent: discountPct,
			WarehouseCode:   s.cfg.NembusDefaultWarehouse,
		})
	}

	// 3. Prepare Invoice Header
	cardCode := s.resolveCardCode(ctx, tx.CustomerID)

	invDate := tx.TransactionDate.Format("2006-01-02")
	invReq := &sap.SAPInvoiceRequest{
		CardCode:      cardCode,
		DocDate:       invDate,
		DocDueDate:    invDate,
		TaxDate:       invDate,
		NumAtCard:     tx.TransactionNumber,
		Reference2:    tx.TransactionNumber,
		Comments:      fmt.Sprintf("Nembus POS Tx: %s | Shift: %d | Cashier: %d", tx.TransactionNumber, tx.CashierSessionID, tx.CashierID),
		DocumentLines: docLines,
	}

	if s.cfg.DryRun {
		log.Printf("[DRY-RUN] Would post invoice for %s ($%.2f) to SAP", tx.TransactionNumber, tx.TotalAmount)
		return nil
	}

	// 4. Post Invoice to SAP
	sapInv, err := s.sapClient.PostInvoice(ctx, invReq)
	if err != nil {
		return err
	}

	log.Printf(" Invoice created in SAP B1! DocEntry: %d, DocNum: %d (Ref: %s)",
		sapInv.DocEntry, sapInv.DocNum, tx.TransactionNumber)

	// 5. Post matching payment in SAP if paid
	if tx.AmountPaid > 0 {
		payments, err := s.nembusClient.FetchTransactionPayments(ctx, tx.ID)
		if err == nil && len(payments) > 0 {
			for _, p := range payments {
				payReq := &sap.SAPPaymentRequest{
					CardCode: cardCode,
					DocDate:  invDate,
					DocType:  "rCustomer",
					PaymentInvoices: []sap.SAPPendingPaymentInvoice{
						{
							LineNum:     0,
							DocEntry:    sapInv.DocEntry,
							SumApplied:  p.Amount,
							InvoiceType: "it_Invoice",
						},
					},
					Remarks: fmt.Sprintf("POS Payment for %s (Method: %s)", tx.TransactionNumber, p.PaymentMethod),
				}

				if p.PaymentMethod == "cash" || p.PaymentMethod == "CASH" {
					payReq.CashSum = p.Amount
				} else {
					payReq.TransferSum = p.Amount
					if p.Reference != nil {
						payReq.TransferReference = *p.Reference
					}
				}

				if payResp, err := s.sapClient.PostIncomingPayment(ctx, payReq); err == nil {
					log.Printf("  Incoming Payment created in SAP! DocEntry: %d, Sum: %.2f",
						payResp.DocEntry, p.Amount)
				} else {
					log.Printf("⚠️ Warning: Failed to post payment for DocEntry %d: %v", sapInv.DocEntry, err)
				}
			}
		}
	}

	// 6. Mark transaction as synced in Nembus
	return s.nembusClient.MarkTransactionPostedToSAP(ctx, tx.ID, sapInv.DocEntry, sapInv.DocNum)
}

func (s *UpstreamSync) resolveCardCode(ctx context.Context, customerID *int32) string {
	if customerID == nil {
		if s.cfg.NembusDefaultCustomer != "" {
			return s.cfg.NembusDefaultCustomer
		}
		return "C000001"
	}
	code, err := s.nembusClient.FetchCustomerSAPCode(ctx, *customerID)
	if err == nil && strings.TrimSpace(code) != "" {
		return strings.TrimSpace(code)
	}

	// 1. Fetch Nembus customer details
	cust, err := s.nembusClient.FetchCustomerDetails(ctx, *customerID)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to fetch customer details for ID %d: %v", *customerID, err)
		if s.cfg.NembusDefaultCustomer != "" {
			return s.cfg.NembusDefaultCustomer
		}
		return "C000001"
	}

	// 2. Create BP in SAP
	newCode := fmt.Sprintf("NMB-C%d", cust.ID)
	bp := &sap.SAPBusinessPartner{
		CardCode: newCode,
		CardName: cust.Name,
		CardType: "cCustomer",
	}
	if cust.Phone != nil {
		bp.Phone1 = *cust.Phone
	}
	if cust.Email != nil {
		bp.EmailAddress = *cust.Email
	}

	if _, err := s.sapClient.PostBusinessPartner(ctx, bp); err != nil {
		log.Printf("⚠️ Warning: Failed to create SAP BP %s: %v", newCode, err)
		// Fallback to default
		if s.cfg.NembusDefaultCustomer != "" {
			return s.cfg.NembusDefaultCustomer
		}
		return "C000001"
	}

	// 3. Save back to Nembus
	if err := s.nembusClient.UpdateCustomerSAPCode(ctx, cust.ID, newCode); err != nil {
		log.Printf("⚠️ Warning: Failed to update customer %d with SAP code %s: %v", cust.ID, newCode, err)
	}

	return newCode
}

