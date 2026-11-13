package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"encoding/json"

	"github.com/NasTecSol/nembus-sap-agent/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap-agent/internal/nembus"
	"github.com/NasTecSol/nembus-sap-agent/internal/sap"
	"github.com/google/uuid"
)

type SAPClient interface {
	PostInvoice(ctx context.Context, invoice *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error)
	PostIncomingPayment(ctx context.Context, payment *sap.SAPPaymentRequest) (*sap.SAPPaymentResponse, error)
	PostCreditMemo(ctx context.Context, memo *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error)
	CheckItemWarehouseStock(ctx context.Context, itemCode, warehouseCode string) (float64, error)
}

type NembusClient interface {
	FetchApprovedPendingPOSTransactions(ctx context.Context, limit int) ([]nembus.POSTransaction, error)
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
	FetchApprovedPendingSalesOrdersV2(ctx context.Context, limit int) ([]nembus.SalesOrderV2, error)
	FetchUnpostedSalesOrdersV2(ctx context.Context, limit int) ([]nembus.SalesOrderV2, error)
	FetchSalesOrderLinesV2(ctx context.Context, orderID uuid.UUID) ([]nembus.SalesOrderLineV2, error)
	MarkSalesOrderPostedToSAP(ctx context.Context, orderID uuid.UUID, sapDocEntry, sapDocNum int) error
	HoldSalesOrderForStockReview(ctx context.Context, orderID uuid.UUID, orderNum string, storeID int32, whCode string, totalAmount float64, shortageItems []nembus.ShortageItem, summary string) error
	GetApprovalPolicyRules(ctx context.Context, domain, action string) (map[string]any, bool, error)
}

type stockHoldPolicy struct {
	requiresApproval bool
	holdOnZeroStock  bool
}

func (s *UpstreamSync) loadStockHoldPolicy(ctx context.Context) stockHoldPolicy {
	rules, requiresApproval, err := s.nembusClient.GetApprovalPolicyRules(ctx, "sap_invoice", "stock_hold")
	if err != nil {
		log.Printf("⚠️ Warning: could not load sap_invoice.stock_hold policy: %v", err)
		return stockHoldPolicy{}
	}
	if !requiresApproval {
		return stockHoldPolicy{}
	}
	holdOnZeroStock, _ := rules["hold_on_zero_stock"].(bool)
	return stockHoldPolicy{
		requiresApproval: true,
		holdOnZeroStock:  holdOnZeroStock,
	}
}

func shouldHoldShortage(policy stockHoldPolicy, onHand, qty float64) bool {
	if !policy.requiresApproval || onHand >= qty {
		return false
	}
	if onHand <= 0 && !policy.holdOnZeroStock {
		return false
	}
	return true
}

type UpstreamSync struct {
	cfg          *config.Config
	sapClient    SAPClient
	nembusClient NembusClient
	sqlite       *db.SQLiteStore
	logFunc      func(domain, level, msg string)
}

func NewUpstreamSync(cfg *config.Config, sapClient SAPClient, nembusClient NembusClient, sqliteStore ...*db.SQLiteStore) *UpstreamSync {
	var store *db.SQLiteStore
	if len(sqliteStore) > 0 {
		store = sqliteStore[0]
	}
	return &UpstreamSync{
		cfg:          cfg,
		sapClient:    sapClient,
		nembusClient: nembusClient,
		sqlite:       store,
	}
}

func (s *UpstreamSync) SetSQLite(store *db.SQLiteStore) {
	s.sqlite = store
}

func (s *UpstreamSync) SetLogger(fn func(domain, level, msg string)) {
	s.logFunc = fn
}

func (s *UpstreamSync) emitLog(domain, level, msg string) {
	if s.logFunc != nil {
		s.logFunc(domain, level, msg)
	}
}

var (
	ErrTransactionHeldForReview = errors.New("transaction held for stock review")
	ErrOrderHeldForReview       = errors.New("sales order held for stock review")
	errNoLineItems              = errors.New("no line items")
)

func (s *UpstreamSync) recordSkippedDispatch(ctx context.Context, docNum string, docDate time.Time, amount float64, reason string) {
	if s.sqlite != nil {
		_ = s.sqlite.RecordUpstreamDispatch(ctx, docNum, docDate, amount, nil, nil, "skipped", reason)
	}
}

type UpstreamStats struct {
	TransactionsPosted int           `json:"transactions_posted"`
	TransactionsHeld   int           `json:"transactions_held,omitempty"`
	OrdersPosted       int           `json:"orders_posted,omitempty"`
	OrdersHeld         int           `json:"orders_held,omitempty"`
	ReturnsPosted      int           `json:"returns_posted"`
	PaymentsPosted     int           `json:"payments_posted,omitempty"`
	Errors             int           `json:"errors"`
	Duration           time.Duration `json:"duration"`
}

func skipSAPStockPreflight(metadata json.RawMessage) bool {
	if len(metadata) == 0 {
		return false
	}
	var m map[string]any
	if err := json.Unmarshal(metadata, &m); err != nil {
		return false
	}
	state, _ := m["sap_sync_state"].(string)
	return state == "approved_pending_upload"
}

func beforeWatermark(ts, since time.Time) bool {
	return !since.IsZero() && !ts.IsZero() && ts.Before(since)
}

func (s *UpstreamSync) resolveUpstreamSince(ctx context.Context, now time.Time) time.Time {
	if s.sqlite == nil {
		return time.Time{}
	}
	var since time.Time
	if wm, err := s.sqlite.GetUpstreamWatermark(ctx, "upstream"); err == nil && wm != nil {
		since = *wm
	}
	if migEnd, err := s.sqlite.GetLatestCompletedFullMigrationEndTime(ctx); err == nil && migEnd != nil {
		if since.IsZero() || migEnd.Truncate(time.Second).After(since.Truncate(time.Second)) {
			return *migEnd
		}
	}
	if since.IsZero() {
		since = now.Truncate(24 * time.Hour)
		_ = s.sqlite.SetUpstreamWatermark(ctx, "upstream", since, 0)
	}
	return since
}

func (s *UpstreamSync) ProcessOutbox(ctx context.Context) (*UpstreamStats, error) {
	start := time.Now()
	stats := &UpstreamStats{}
	since := s.resolveUpstreamSince(ctx, start)

	s.emitLog("pipeline", "INFO", "[Upstream] Checking Nembus Cloud for unposted POS transactions, sales orders, and returns...")

	limit := s.cfg.BatchSize
	if limit <= 0 {
		limit = 50
	}

	approvedPOS, err := s.nembusClient.FetchApprovedPendingPOSTransactions(ctx, limit)
	if err != nil {
		return stats, fmt.Errorf("failed to fetch approved pending POS transactions: %w", err)
	}
	for _, tx := range approvedPOS {
		if err := s.postTransactionToSAP(ctx, &tx, stockHoldPolicy{}); err != nil {
			log.Printf("❌ Failed to post approved transaction %s (ID: %d) to SAP: %v", tx.TransactionNumber, tx.ID, err)
			s.emitLog("pos_transactions", "ERROR", fmt.Sprintf("[Upstream] pos_transactions: failed to post approved %s: %v", tx.TransactionNumber, err))
			stats.Errors++
		} else if !s.cfg.DryRun {
			stats.TransactionsPosted++
		}
	}

	approvedOrders, err := s.nembusClient.FetchApprovedPendingSalesOrdersV2(ctx, limit)
	if err != nil {
		return stats, fmt.Errorf("failed to fetch approved pending sales orders: %w", err)
	}
	for _, order := range approvedOrders {
		if err := s.postSalesOrderToSAP(ctx, &order, stockHoldPolicy{}); err != nil {
			log.Printf("❌ Failed to post approved sales order %s to SAP: %v", order.OrderNumber, err)
			s.emitLog("sales_orders", "ERROR", fmt.Sprintf("[Upstream] sales_orders: failed to post approved %s: %v", order.OrderNumber, err))
			stats.Errors++
		} else if !s.cfg.DryRun {
			stats.OrdersPosted++
		}
	}

	var stockPolicy *stockHoldPolicy
	getStockPolicy := func() stockHoldPolicy {
		if stockPolicy == nil {
			p := s.loadStockHoldPolicy(ctx)
			stockPolicy = &p
		}
		return *stockPolicy
	}

	txs, err := s.nembusClient.FetchUnpostedPOSTransactions(ctx, limit)
	if err != nil {
		return stats, fmt.Errorf("failed to fetch unposted transactions: %w", err)
	}

	if len(txs) > 0 {
		policy := getStockPolicy()
		for _, tx := range txs {
			if beforeWatermark(tx.TransactionDate, since) {
				continue
			}
			if err := s.postTransactionToSAP(ctx, &tx, policy); err != nil {
				if errors.Is(err, ErrTransactionHeldForReview) {
					stats.TransactionsHeld++
					continue
				}
				if errors.Is(err, errNoLineItems) {
					s.emitLog("pos_transactions", "WARN", fmt.Sprintf("[Upstream] pos_transactions: skipped %s (no line items)", tx.TransactionNumber))
					continue
				}
				log.Printf("❌ Failed to post transaction %s (ID: %d) to SAP: %v", tx.TransactionNumber, tx.ID, err)
				s.emitLog("pos_transactions", "ERROR", fmt.Sprintf("[Upstream] pos_transactions: failed to post %s: %v", tx.TransactionNumber, err))
				stats.Errors++
			} else if !s.cfg.DryRun {
				stats.TransactionsPosted++
			}
		}
	}
	if stats.TransactionsPosted == 0 && stats.TransactionsHeld == 0 {
		s.emitLog("pos_transactions", "INFO", "[Upstream] pos_transactions: up to date (0 unposted transactions)")
	} else {
		s.emitLog("pos_transactions", "INFO", fmt.Sprintf("[Upstream] pos_transactions: %d posted, %d held", stats.TransactionsPosted, stats.TransactionsHeld))
	}

	salesOrders, oErr := s.nembusClient.FetchUnpostedSalesOrdersV2(ctx, limit)
	if oErr != nil {
		log.Printf("⚠️ Failed to fetch unposted sales orders: %v", oErr)
		stats.Errors++
	} else if len(salesOrders) > 0 {
		policy := getStockPolicy()
		for _, order := range salesOrders {
			if beforeWatermark(order.OrderDate, since) {
				continue
			}
			if err := s.postSalesOrderToSAP(ctx, &order, policy); err != nil {
				if errors.Is(err, ErrOrderHeldForReview) {
					stats.OrdersHeld++
					continue
				}
				if errors.Is(err, errNoLineItems) {
					s.emitLog("sales_orders", "WARN", fmt.Sprintf("[Upstream] sales_orders: skipped %s (no line items)", order.OrderNumber))
					continue
				}
				log.Printf("❌ Failed to post sales order %s to SAP: %v", order.OrderNumber, err)
				s.emitLog("sales_orders", "ERROR", fmt.Sprintf("[Upstream] sales_orders: failed to post %s: %v", order.OrderNumber, err))
				stats.Errors++
			} else if !s.cfg.DryRun {
				stats.OrdersPosted++
			}
		}
	}
	if stats.OrdersPosted == 0 && stats.OrdersHeld == 0 {
		s.emitLog("sales_orders", "INFO", "[Upstream] sales_orders: up to date (0 unposted orders)")
	} else {
		s.emitLog("sales_orders", "INFO", fmt.Sprintf("[Upstream] sales_orders: %d posted, %d held", stats.OrdersPosted, stats.OrdersHeld))
	}

	// Fetch unposted completed sales returns from Nembus
	returns, rErr := s.nembusClient.FetchUnpostedSalesReturns(ctx, limit)
	if rErr != nil {
		log.Printf("⚠️ Failed to fetch unposted sales returns: %v", rErr)
		stats.Errors++
	} else if len(returns) > 0 {
		for _, ret := range returns {
			if beforeWatermark(ret.ReturnDate, since) {
				continue
			}
			if err := s.postSalesReturnToSAP(ctx, &ret); err != nil {
				if errors.Is(err, errNoLineItems) {
					log.Printf("⚠️ Skipping sales return %s (ID: %d): %v", ret.ReturnNumber, ret.ID, err)
					s.emitLog("sales_returns", "WARN", fmt.Sprintf("[Upstream] sales_returns: skipped %s: %v", ret.ReturnNumber, err))
					continue
				}
				log.Printf("❌ Failed to post sales return %s (ID: %d) to SAP: %v", ret.ReturnNumber, ret.ID, err)
				s.emitLog("sales_returns", "ERROR", fmt.Sprintf("[Upstream] sales_returns: failed to post %s: %v", ret.ReturnNumber, err))
				stats.Errors++
			} else if !s.cfg.DryRun {
				stats.ReturnsPosted++
			}
		}
	}
	if stats.ReturnsPosted == 0 {
		s.emitLog("sales_returns", "INFO", "[Upstream] sales_returns: up to date (0 unposted returns)")
	} else {
		s.emitLog("sales_returns", "INFO", fmt.Sprintf("[Upstream] sales_returns: %d posted", stats.ReturnsPosted))
	}

	if s.sqlite != nil && stats.Errors == 0 && !s.cfg.DryRun {
		_ = s.sqlite.SetUpstreamWatermark(ctx, "upstream", start, stats.TransactionsPosted+stats.OrdersPosted+stats.ReturnsPosted)
	}

	stats.Duration = time.Since(start)
	s.emitLog("pipeline", "INFO", fmt.Sprintf("Upstream sync complete in %v: %d transactions posted, %d orders posted, %d returns posted, %d held",
		stats.Duration.Round(time.Millisecond), stats.TransactionsPosted, stats.OrdersPosted, stats.ReturnsPosted, stats.TransactionsHeld+stats.OrdersHeld))
	return stats, nil
}

func (s *UpstreamSync) postTransactionToSAP(ctx context.Context, tx *nembus.POSTransaction, policy stockHoldPolicy) error {
	// 1. Fetch lines
	lines, err := s.nembusClient.FetchTransactionLines(ctx, tx.ID)
	if err != nil {
		return fmt.Errorf("failed fetching lines: %w", err)
	}

	if len(lines) == 0 {
		s.recordSkippedDispatch(ctx, tx.TransactionNumber, tx.TransactionDate, tx.TotalAmount, errNoLineItems.Error())
		return fmt.Errorf("transaction %s: %w", tx.TransactionNumber, errNoLineItems)
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

	var shortages []nembus.ShortageItem
	if policy.requiresApproval && !skipSAPStockPreflight(tx.Metadata) {
		for _, line := range lines {
			onHand, err := s.sapClient.CheckItemWarehouseStock(ctx, line.ProductSKU, whCode)
			if err == nil && shouldHoldShortage(policy, onHand, line.Quantity) {
				shortages = append(shortages, nembus.ShortageItem{
					ProductID:   line.ProductID,
					ProductSKU:  line.ProductSKU,
					ProductName: line.ProductName,
					Quantity:    line.Quantity,
					SAPOnHand:   onHand,
				})
			}
		}
	}

	if len(shortages) > 0 {
		summary := fmt.Sprintf("SAP Stock Shortage on %s: %d item(s) below required quantity in warehouse %s", tx.TransactionNumber, len(shortages), whCode)
		log.Printf("⚠️ %s. Routing to Approval Review Queue...", summary)
		if err := s.nembusClient.HoldTransactionForStockReview(ctx, tx.ID, tx.TransactionNumber, tx.StoreID, whCode, tx.TotalAmount, shortages, summary); err != nil {
			return fmt.Errorf("failed holding transaction for review: %w", err)
		}
		if s.sqlite != nil {
			_ = s.sqlite.RecordUpstreamDispatch(ctx, tx.TransactionNumber, tx.TransactionDate, tx.TotalAmount, nil, nil, "held", summary)
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
		if policy.requiresApproval && !skipSAPStockPreflight(tx.Metadata) && (strings.Contains(errStr, "fall below zero") || strings.Contains(errStr, "negative inventory") || strings.Contains(errStr, "code -10") || strings.Contains(errStr, "code 10")) {
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
		if s.sqlite != nil {
			_ = s.sqlite.RecordUpstreamDispatch(ctx, tx.TransactionNumber, tx.TransactionDate, tx.TotalAmount, nil, nil, "failed", err.Error())
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
	err = s.nembusClient.MarkTransactionPostedToSAP(ctx, tx.ID, sapInv.DocEntry, sapInv.DocNum)
	if err == nil && s.sqlite != nil {
		_ = s.sqlite.RecordUpstreamDispatch(ctx, tx.TransactionNumber, tx.TransactionDate, tx.TotalAmount, &sapInv.DocEntry, &sapInv.DocNum, "posted", "")
	}
	return err
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
		s.recordSkippedDispatch(ctx, ret.ReturnNumber, ret.ReturnDate, ret.TotalRefundAmount, errNoLineItems.Error())
		return fmt.Errorf("sales return %s: %w", ret.ReturnNumber, errNoLineItems)
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
		if s.sqlite != nil {
			_ = s.sqlite.RecordUpstreamDispatch(ctx, ret.ReturnNumber, ret.ReturnDate, ret.TotalRefundAmount, nil, nil, "failed", err.Error())
		}
		return err
	}

	log.Printf(" Credit Note created in SAP B1! DocEntry: %d, DocNum: %d (Ref: %s)",
		sapMemo.DocEntry, sapMemo.DocNum, ret.ReturnNumber)

	err = s.nembusClient.MarkSalesReturnPostedToSAP(ctx, ret.ID, sapMemo.DocEntry, sapMemo.DocNum)
	if err == nil && s.sqlite != nil {
		_ = s.sqlite.RecordUpstreamDispatch(ctx, ret.ReturnNumber, ret.ReturnDate, ret.TotalRefundAmount, &sapMemo.DocEntry, &sapMemo.DocNum, "posted", "")
	}
	return err
}

func (s *UpstreamSync) postSalesOrderToSAP(ctx context.Context, order *nembus.SalesOrderV2, policy stockHoldPolicy) error {
	lines, err := s.nembusClient.FetchSalesOrderLinesV2(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("failed fetching sales order lines: %w", err)
	}
	if len(lines) == 0 {
		s.recordSkippedDispatch(ctx, order.OrderNumber, order.OrderDate, order.TotalAmount, errNoLineItems.Error())
		return fmt.Errorf("sales order %s: %w", order.OrderNumber, errNoLineItems)
	}

	var docLines []sap.SAPDocumentLine
	for idx, line := range lines {
		if line.Quantity <= 0 {
			continue
		}
		unitPriceNet := line.UnitPrice
		if line.TaxAmount > 0 && line.Quantity > 0 {
			unitPriceNet = (line.LineTotal - line.TaxAmount) / line.Quantity
			if unitPriceNet <= 0 {
				unitPriceNet = line.UnitPrice
			}
		}
		lineNum := idx
		priceAfterVAT := line.LineTotal / line.Quantity
		docLines = append(docLines, sap.SAPDocumentLine{
			LineNum:         &lineNum,
			ItemCode:        line.ProductSKU,
			ItemDescription: line.ProductName,
			Quantity:        line.Quantity,
			UnitPrice:       unitPriceNet,
			PriceAfterVAT:   priceAfterVAT,
			WarehouseCode:   s.cfg.NembusDefaultWarehouse,
		})
	}
	if len(docLines) == 0 {
		return fmt.Errorf("sales order %s has no billable lines", order.OrderNumber)
	}

	cardCode := s.resolveCardCode(ctx, order.CustomerID)
	invDate := order.OrderDate.Format("2006-01-02")
	invReq := &sap.SAPInvoiceRequest{
		CardCode:      cardCode,
		DocDate:       invDate,
		DocDueDate:    invDate,
		TaxDate:       invDate,
		NumAtCard:     order.OrderNumber,
		Reference2:    order.OrderNumber,
		Comments:      fmt.Sprintf("Nembus Sales Order: %s", order.OrderNumber),
		DocumentLines: docLines,
	}

	whCode := s.cfg.NembusDefaultWarehouse
	var shortages []nembus.ShortageItem
	if policy.requiresApproval && !skipSAPStockPreflight(order.Metadata) {
		for _, line := range lines {
			if line.Quantity <= 0 {
				continue
			}
			onHand, err := s.sapClient.CheckItemWarehouseStock(ctx, line.ProductSKU, whCode)
			if err == nil && shouldHoldShortage(policy, onHand, line.Quantity) {
				shortages = append(shortages, nembus.ShortageItem{
					ProductID:   line.ProductID,
					ProductSKU:  line.ProductSKU,
					ProductName: line.ProductName,
					Quantity:    line.Quantity,
					SAPOnHand:   onHand,
				})
			}
		}
	}

	if len(shortages) > 0 {
		summary := fmt.Sprintf("SAP Stock Shortage on order %s: %d item(s) below required quantity in warehouse %s", order.OrderNumber, len(shortages), whCode)
		log.Printf("⚠️ %s. Routing to Approval Review Queue...", summary)
		if err := s.nembusClient.HoldSalesOrderForStockReview(ctx, order.ID, order.OrderNumber, order.StoreID, whCode, order.TotalAmount, shortages, summary); err != nil {
			return fmt.Errorf("failed holding sales order for review: %w", err)
		}
		return ErrOrderHeldForReview
	}

	if s.cfg.DryRun {
		log.Printf("[DRY-RUN] Would post OINV for sales order %s ($%.2f) to SAP", order.OrderNumber, order.TotalAmount)
		return nil
	}

	sapInv, err := s.sapClient.PostInvoice(ctx, invReq)
	if err != nil {
		errStr := strings.ToLower(err.Error())
		if policy.requiresApproval && !skipSAPStockPreflight(order.Metadata) && (strings.Contains(errStr, "fall below zero") || strings.Contains(errStr, "negative inventory") || strings.Contains(errStr, "code -10") || strings.Contains(errStr, "code 10")) {
			summary := fmt.Sprintf("SAP Negative Inventory Rejection on order %s: %v", order.OrderNumber, err)
			var fallback []nembus.ShortageItem
			for _, l := range lines {
				fallback = append(fallback, nembus.ShortageItem{
					ProductID:   l.ProductID,
					ProductSKU:  l.ProductSKU,
					ProductName: l.ProductName,
					Quantity:    l.Quantity,
					SAPOnHand:   0,
				})
			}
			_ = s.nembusClient.HoldSalesOrderForStockReview(ctx, order.ID, order.OrderNumber, order.StoreID, whCode, order.TotalAmount, fallback, summary)
		}
		return err
	}

	log.Printf(" Invoice created in SAP B1 from sales order! DocEntry: %d, DocNum: %d (Ref: %s)",
		sapInv.DocEntry, sapInv.DocNum, order.OrderNumber)

	return s.nembusClient.MarkSalesOrderPostedToSAP(ctx, order.ID, sapInv.DocEntry, sapInv.DocNum)
}

