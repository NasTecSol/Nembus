package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/nembus"
	"github.com/NasTecSol/nembus-sap-agent/internal/sap"
)

type SAPClient interface {
	PostInvoice(ctx context.Context, invoice *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error)
	PostIncomingPayment(ctx context.Context, payment *sap.SAPPaymentRequest) (*sap.SAPPaymentResponse, error)
	PostCreditMemo(ctx context.Context, memo *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error)
	CheckItemWarehouseStock(ctx context.Context, itemCode, warehouseCode string) (float64, error)
}

type NembusClient interface {
	FetchUnpostedPOSTransactions(ctx context.Context, limit int) ([]nembus.POSTransaction, error)
	FetchTransactionLines(ctx context.Context, transactionID int64) ([]nembus.POSTransactionLine, error)
	FetchTransactionPayments(ctx context.Context, transactionID int64) ([]nembus.POSPayment, error)
	MarkTransactionPostedToSAP(ctx context.Context, transactionID int64, sapDocEntry, sapDocNum int) error
	FetchCustomerSAPCode(ctx context.Context, customerID int32) (string, error)
	FetchUnpostedSalesReturns(ctx context.Context, limit int) ([]nembus.SalesReturn, error)
	FetchSalesReturnLines(ctx context.Context, returnID int64) ([]nembus.SalesReturnLine, error)
	MarkSalesReturnPostedToSAP(ctx context.Context, returnID int64, sapDocEntry, sapDocNum int) error
	FetchOriginalSAPDocEntry(ctx context.Context, originalTxID int32) (int, error)
	HoldTransactionForStockReview(ctx context.Context, txID int64, txNum string, storeID int32, whCode string, totalAmount float64, shortageItems []nembus.ShortageItem, summary string) error
}

type UpstreamSync struct {
	cfg          *config.Config
	sapClient    SAPClient
	nembusClient NembusClient
}

func NewUpstreamSync(cfg *config.Config, sapClient SAPClient, nembusClient NembusClient) *UpstreamSync {
	return &UpstreamSync{
		cfg:          cfg,
		sapClient:    sapClient,
		nembusClient: nembusClient,
	}
}

var ErrTransactionHeldForReview = errors.New("transaction held for stock review")

type UpstreamStats struct {
	TransactionsPosted int           `json:"transactions_posted"`
	TransactionsHeld   int           `json:"transactions_held,omitempty"`
	ReturnsPosted      int           `json:"returns_posted"`
	PaymentsPosted     int           `json:"payments_posted,omitempty"`
	Errors             int           `json:"errors"`
	Duration           time.Duration `json:"duration"`
}

func (s *UpstreamSync) ProcessOutbox(ctx context.Context) (*UpstreamStats, error) {
	start := time.Now()
	stats := &UpstreamStats{}

	limit := s.cfg.BatchSize
	if limit <= 0 {
		limit = 50
	}

	// 1. Fetch unposted completed POS transactions from Nembus
	txs, err := s.nembusClient.FetchUnpostedPOSTransactions(ctx, limit)
	if err != nil {
		return stats, fmt.Errorf("failed to fetch unposted transactions: %w", err)
	}

	if len(txs) > 0 {
		log.Printf("⬆️ [Nembus POS -> SAP] Found %d unposted transaction(s). Processing...", len(txs))
		for _, tx := range txs {
			if err := s.postTransactionToSAP(ctx, &tx); err != nil {
				if errors.Is(err, ErrTransactionHeldForReview) {
					stats.TransactionsHeld++
					continue
				}
				log.Printf("❌ Failed to post transaction %s (ID: %d) to SAP: %v", tx.TransactionNumber, tx.ID, err)
				stats.Errors++
			} else if !s.cfg.DryRun {
				stats.TransactionsPosted++
			}
		}
	}

	// 2. Fetch unposted completed sales returns from Nembus
	returns, rErr := s.nembusClient.FetchUnpostedSalesReturns(ctx, limit)
	if rErr != nil {
		log.Printf("⚠️ Failed to fetch unposted sales returns: %v", rErr)
		stats.Errors++
	} else if len(returns) > 0 {
		log.Printf("⬆️ [Nembus POS -> SAP] Found %d unposted return(s). Processing...", len(returns))
		for _, ret := range returns {
			if err := s.postSalesReturnToSAP(ctx, &ret); err != nil {
				log.Printf("❌ Failed to post sales return %s (ID: %d) to SAP: %v", ret.ReturnNumber, ret.ID, err)
				stats.Errors++
			} else if !s.cfg.DryRun {
				stats.ReturnsPosted++
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

	whCode := s.cfg.NembusDefaultWarehouse

	// 2. Pre-flight stock check against SAP to avoid overselling
	var shortages []nembus.ShortageItem
	for _, line := range lines {
		onHand, err := s.sapClient.CheckItemWarehouseStock(ctx, line.ProductSKU, whCode)
		if err == nil && onHand < line.Quantity {
			shortages = append(shortages, nembus.ShortageItem{
				ProductID:   line.ProductID,
				ProductSKU:  line.ProductSKU,
				ProductName: line.ProductName,
				Quantity:    line.Quantity,
				SAPOnHand:   onHand,
			})
		}
	}

	if len(shortages) > 0 {
		summary := fmt.Sprintf("SAP Stock Shortage on %s: %d item(s) below required quantity in warehouse %s", tx.TransactionNumber, len(shortages), whCode)
		log.Printf("⚠️ %s. Routing to Approval Review Queue...", summary)
		if err := s.nembusClient.HoldTransactionForStockReview(ctx, tx.ID, tx.TransactionNumber, tx.StoreID, whCode, tx.TotalAmount, shortages, summary); err != nil {
			return fmt.Errorf("failed holding transaction for review: %w", err)
		}
		return ErrTransactionHeldForReview
	}

	if s.cfg.DryRun {
		log.Printf("[DRY-RUN] Would post invoice for %s ($%.2f) to SAP", tx.TransactionNumber, tx.TotalAmount)
		return nil
	}

	// 4. Post Invoice to SAP (with reactive negative inventory interception)
	sapInv, err := s.sapClient.PostInvoice(ctx, invReq)
	if err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "fall below zero") || strings.Contains(errStr, "negative inventory") || strings.Contains(errStr, "code -10") || strings.Contains(errStr, "code 10") {
			summary := fmt.Sprintf("SAP Negative Inventory Rejection on %s: %v", tx.TransactionNumber, err)
			log.Printf("⚠️ %s. Intercepted and routing to Approval Review Queue...", summary)
			var fallbackShortages []nembus.ShortageItem
			for _, l := range lines {
				fallbackShortages = append(fallbackShortages, nembus.ShortageItem{
					ProductID:   l.ProductID,
					ProductSKU:  l.ProductSKU,
					ProductName: l.ProductName,
					Quantity:    l.Quantity,
					SAPOnHand:   0,
				})
			}
			_ = s.nembusClient.HoldTransactionForStockReview(ctx, tx.ID, tx.TransactionNumber, tx.StoreID, whCode, tx.TotalAmount, fallbackShortages, summary)
		}
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
	if err != nil || strings.TrimSpace(code) == "" {
		if s.cfg.NembusDefaultCustomer != "" {
			return s.cfg.NembusDefaultCustomer
		}
		return "C000001"
	}
	return strings.TrimSpace(code)
}

func (s *UpstreamSync) postSalesReturnToSAP(ctx context.Context, ret *nembus.SalesReturn) error {
	lines, err := s.nembusClient.FetchSalesReturnLines(ctx, ret.ID)
	if err != nil {
		return fmt.Errorf("failed fetching return lines: %w", err)
	}
	if len(lines) == 0 {
		return fmt.Errorf("sales return %s has no line items", ret.ReturnNumber)
	}

	// Lookup original invoice DocEntry in SAP if linked
	var baseDocEntry *int
	if ret.OriginalTransactionID != nil && *ret.OriginalTransactionID > 0 {
		if entry, err := s.nembusClient.FetchOriginalSAPDocEntry(ctx, *ret.OriginalTransactionID); err == nil && entry > 0 {
			baseDocEntry = &entry
		} else {
			log.Printf("⚠️ [SalesReturn %s] Original transaction ID %d has no SAP DocEntry (err: %v); posting as standalone credit memo", ret.ReturnNumber, *ret.OriginalTransactionID, err)
		}
	}

	var docLines []sap.SAPDocumentLine
	for idx, line := range lines {
		if line.Quantity <= 0 {
			return fmt.Errorf("sales return %s line %d has invalid quantity: %.3f", ret.ReturnNumber, idx, line.Quantity)
		}
		if line.RefundAmount < 0 {
			return fmt.Errorf("sales return %s line %d has negative refund amount: %.2f", ret.ReturnNumber, idx, line.RefundAmount)
		}
		lineNum := idx
		priceAfterVAT := line.RefundAmount / line.Quantity
		docLine := sap.SAPDocumentLine{
			LineNum:         &lineNum,
			ItemCode:        line.ProductSKU,
			ItemDescription: line.ProductName,
			Quantity:        line.Quantity,
			UnitPrice:       line.UnitPrice,
			PriceAfterVAT:   priceAfterVAT,
			WarehouseCode:   s.cfg.NembusDefaultWarehouse,
		}
		if baseDocEntry != nil {
			baseType := 13 // 13 = A/R Invoice
			docLine.BaseEntry = baseDocEntry
			docLine.BaseType = &baseType
		}
		docLines = append(docLines, docLine)
	}

	cardCode := s.resolveCardCode(ctx, ret.CustomerID)
	retDate := ret.ReturnDate.Format("2006-01-02")
	comments := fmt.Sprintf("Nembus POS Return: %s", ret.ReturnNumber)
	if ret.ReturnReason != nil && *ret.ReturnReason != "" {
		comments = fmt.Sprintf("%s (Reason: %s)", comments, *ret.ReturnReason)
	}

	memoReq := &sap.SAPInvoiceRequest{
		CardCode:      cardCode,
		DocDate:       retDate,
		DocDueDate:    retDate,
		TaxDate:       retDate,
		NumAtCard:     ret.ReturnNumber,
		Reference2:    ret.ReturnNumber,
		Comments:      comments,
		DocumentLines: docLines,
	}

	if s.cfg.DryRun {
		log.Printf("[DRY-RUN] Would post Credit Note for %s ($%.2f) to SAP", ret.ReturnNumber, ret.TotalRefundAmount)
		return nil
	}

	sapMemo, err := s.sapClient.PostCreditMemo(ctx, memoReq)
	if err != nil {
		return err
	}

	log.Printf(" Credit Note created in SAP B1! DocEntry: %d, DocNum: %d (Ref: %s)",
		sapMemo.DocEntry, sapMemo.DocNum, ret.ReturnNumber)

	return s.nembusClient.MarkSalesReturnPostedToSAP(ctx, ret.ID, sapMemo.DocEntry, sapMemo.DocNum)
}

