package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

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
	ResolveWarehouseBinAllocations(ctx context.Context, itemCode, warehouseCode string, requiredQty float64, isReturn bool) ([]sap.SAPBinAllocation, error)
	ResolveDocumentSeries(ctx context.Context, objectCode string, docDate string, bplID int) (*int, error)
	ResolveInvoiceBaseLine(ctx context.Context, baseDocEntry int, itemCode string) (*int, error)
}

// isSAPNegativeInventoryError returns true when SAP explicitly rejected the
// document due to negative inventory, stock shortage, or bin-location quantity
// shortfall (1470000340) — never on unrelated [code -10] configuration errors
// such as missing bin allocation (1470000341) or closed posting periods.
func isSAPNegativeInventoryError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "fall below zero") ||
		strings.Contains(errStr, "falls into negative") ||
		strings.Contains(errStr, "negative inventory") ||
		strings.Contains(errStr, "10000023") ||
		strings.Contains(errStr, "1470000340") ||
		strings.Contains(errStr, "allocated quantity exceeds available quantity")
}

type NembusClient interface {
	FetchApprovedPendingPOSTransactions(ctx context.Context, limit int) ([]nembus.POSTransaction, error)
	FetchUnpostedPOSTransactions(ctx context.Context, limit int) ([]nembus.POSTransaction, error)
	FetchTransactionLines(ctx context.Context, transactionID uuid.UUID) ([]nembus.POSTransactionLine, error)
	FetchTransactionPayments(ctx context.Context, transactionID uuid.UUID) ([]nembus.POSPayment, error)
	MarkTransactionPostedToSAP(ctx context.Context, transactionID uuid.UUID, sapDocEntry, sapDocNum int) error
	FetchCustomerSAPCode(ctx context.Context, customerID uuid.UUID) (string, error)
	FetchUnpostedSalesReturns(ctx context.Context, limit int) ([]nembus.SalesReturn, error)
	FetchSalesReturnLines(ctx context.Context, returnID uuid.UUID) ([]nembus.SalesReturnLine, error)
	MarkSalesReturnPostedToSAP(ctx context.Context, returnID uuid.UUID, sapDocEntry, sapDocNum int) error
	FetchOriginalSAPDocEntry(ctx context.Context, originalTxID uuid.UUID) (int, error)
	HoldTransactionForStockReview(ctx context.Context, txID uuid.UUID, txNum string, storeID int32, whCode string, totalAmount float64, shortageItems []nembus.ShortageItem, summary string) error
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

const (
	maxPermanentRetryAttempts = 3
	permanentQuarantineDelay  = 6 * time.Hour
)

type retryEntry struct {
	attempts          int
	permanentAttempts int
	nextRetry         time.Time
	quarantined       bool
}

type UpstreamSync struct {
	cfg           *config.Config
	sapClient     SAPClient
	nembusClient  NembusClient
	sqlite        *db.SQLiteStore
	logFunc       func(domain, level, msg string)
	retryMu       sync.Mutex
	failedRetries map[string]retryEntry
}

func NewUpstreamSync(cfg *config.Config, sapClient SAPClient, nembusClient NembusClient, sqliteStore ...*db.SQLiteStore) *UpstreamSync {
	var store *db.SQLiteStore
	if len(sqliteStore) > 0 {
		store = sqliteStore[0]
	}
	return &UpstreamSync{
		cfg:           cfg,
		sapClient:     sapClient,
		nembusClient:  nembusClient,
		sqlite:        store,
		failedRetries: make(map[string]retryEntry),
	}
}

func (s *UpstreamSync) shouldBackoff(key string, now time.Time) (bool, time.Time, bool) {
	s.retryMu.Lock()
	defer s.retryMu.Unlock()
	if s.failedRetries == nil {
		s.failedRetries = make(map[string]retryEntry)
	}
	entry, ok := s.failedRetries[key]
	if !ok && s.sqlite != nil {
		if rec, err := s.sqlite.GetUpstreamRetry(context.Background(), key); err == nil && rec != nil {
			entry = retryEntry{
				attempts:          rec.Attempts,
				permanentAttempts: rec.PermanentAttempts,
				nextRetry:         rec.NextRetry,
				quarantined:       rec.Quarantined,
			}
			s.failedRetries[key] = entry
			ok = true
		}
	}
	if !ok || entry.attempts == 0 {
		return false, time.Time{}, false
	}
	if now.Before(entry.nextRetry) {
		return true, entry.nextRetry, entry.quarantined
	}
	return false, time.Time{}, false
}

func (s *UpstreamSync) persistRetryEntry(key string, entry retryEntry, err error) {
	if s.sqlite == nil {
		return
	}
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	_ = s.sqlite.UpsertUpstreamRetry(context.Background(), db.UpstreamRetryRecord{
		RetryKey:          key,
		Attempts:          entry.attempts,
		PermanentAttempts: entry.permanentAttempts,
		NextRetry:         entry.nextRetry,
		Quarantined:       entry.quarantined,
		LastError:         errMsg,
	})
}

func (s *UpstreamSync) recordRetryFailure(key string, now time.Time, err error) bool {
	s.retryMu.Lock()
	defer s.retryMu.Unlock()
	if s.failedRetries == nil {
		s.failedRetries = make(map[string]retryEntry)
	}
	entry := s.failedRetries[key]
	entry.attempts++

	if sap.IsPermanentBusinessError(err) {
		entry.permanentAttempts++
		if entry.permanentAttempts >= maxPermanentRetryAttempts {
			entry.quarantined = true
			entry.nextRetry = now.Add(permanentQuarantineDelay)
			s.failedRetries[key] = entry
			s.persistRetryEntry(key, entry, err)
			return true
		}
		if entry.permanentAttempts == 1 {
			entry.nextRetry = now.Add(time.Minute)
		} else {
			entry.nextRetry = now.Add(5 * time.Minute)
		}
		s.failedRetries[key] = entry
		s.persistRetryEntry(key, entry, err)
		return false
	}

	entry.quarantined = false
	if entry.attempts == 1 {
		entry.nextRetry = now
	} else {
		exp := entry.attempts - 1
		if exp > 4 {
			exp = 4
		}
		delay := time.Minute << exp
		if delay > 15*time.Minute {
			delay = 15 * time.Minute
		}
		entry.nextRetry = now.Add(delay)
	}
	s.failedRetries[key] = entry
	s.persistRetryEntry(key, entry, err)
	return false
}

func (s *UpstreamSync) recordFailedDispatch(ctx context.Context, docNum string, docDate time.Time, amount float64, reason string) {
	if s.sqlite != nil {
		_ = s.sqlite.RecordUpstreamDispatch(ctx, docNum, docDate, amount, nil, nil, "failed", reason)
	}
}

func (s *UpstreamSync) clearRetry(key string) {
	s.retryMu.Lock()
	defer s.retryMu.Unlock()
	if s.failedRetries != nil {
		delete(s.failedRetries, key)
	}
	if s.sqlite != nil {
		_ = s.sqlite.DeleteUpstreamRetry(context.Background(), key)
	}
}

// ClearRetryKey removes a document from in-memory and persisted quarantine/backoff.
func (s *UpstreamSync) ClearRetryKey(key string) {
	s.clearRetry(key)
}

func (s *UpstreamSync) SetSQLite(store *db.SQLiteStore) {
	s.sqlite = store
}

func (s *UpstreamSync) SetLogger(fn func(domain, level, msg string)) {
	s.logFunc = fn
}

// Close releases underlying SAP client resources (e.g., MSSQL pool) if supported.
func (s *UpstreamSync) Close() error {
	if closer, ok := s.sapClient.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

func (s *UpstreamSync) attachBinAllocations(ctx context.Context, docLine *sap.SAPDocumentLine, lineNum int, isReturn bool) error {
	if docLine == nil || docLine.ItemCode == "" || docLine.WarehouseCode == "" || docLine.Quantity <= 0 {
		return nil
	}
	allocs, err := s.sapClient.ResolveWarehouseBinAllocations(ctx, docLine.ItemCode, docLine.WarehouseCode, docLine.Quantity, isReturn)
	if err != nil {
		return fmt.Errorf("bin allocation lookup failed for item %s in warehouse %s: %w", docLine.ItemCode, docLine.WarehouseCode, err)
	}
	if len(allocs) == 0 {
		return nil
	}
	for i := range allocs {
		ln := lineNum
		allocs[i].BaseLineNumber = &ln
	}
	docLine.DocumentLinesBinAllocations = allocs
	return nil
}

func (s *UpstreamSync) applyDocumentHeaderMeta(ctx context.Context, req *sap.SAPInvoiceRequest, objectCode string) {
	if req == nil {
		return
	}
	bplID := 0
	if s.cfg != nil && s.cfg.SAPBranchID > 0 {
		bplID = s.cfg.SAPBranchID
		req.BPLIDAssignedToInvoice = &bplID
	}
	if series, err := s.sapClient.ResolveDocumentSeries(ctx, objectCode, req.DocDate, bplID); err == nil && series != nil && *series > 0 {
		req.Series = series
	}
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
	Quarantined        int           `json:"quarantined,omitempty"`
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

func (s *UpstreamSync) resolveUpstreamCutoff(ctx context.Context, now time.Time) time.Time {
	if s.sqlite == nil {
		return time.Time{}
	}
	if migEnd, err := s.sqlite.GetLatestCompletedFullMigrationEndTime(ctx); err == nil && migEnd != nil {
		return *migEnd
	}
	if cutoff, err := s.sqlite.GetUpstreamWatermark(ctx, "upstream_cutoff"); err == nil && cutoff != nil {
		return *cutoff
	}
	cutoff := now.UTC().Truncate(24 * time.Hour)
	_ = s.sqlite.SetUpstreamWatermark(ctx, "upstream_cutoff", cutoff, 0)
	return cutoff
}

func applySAPUoM(docLine *sap.SAPDocumentLine, sapUoMEntry, sapUgpEntry *int, convFactor float64, isBaseUoM bool) {
	if sapUgpEntry != nil && *sapUgpEntry > 0 {
		if sapUoMEntry != nil && *sapUoMEntry > 0 {
			docLine.UoMEntry = sapUoMEntry
		}
		return
	}
	if sapUgpEntry != nil && *sapUgpEntry <= 0 {
		docLine.UseBaseUnits = "tYES"
		if !isBaseUoM && convFactor > 0 && convFactor != 1 {
			docLine.Quantity = docLine.Quantity * convFactor
			docLine.UnitPrice = docLine.UnitPrice / convFactor
			docLine.PriceAfterVAT = docLine.PriceAfterVAT / convFactor
		}
		return
	}
	if sapUoMEntry != nil && *sapUoMEntry > 0 {
		docLine.UoMEntry = sapUoMEntry
	}
}

func (s *UpstreamSync) ProcessOutbox(ctx context.Context) (*UpstreamStats, error) {
	start := time.Now()
	stats := &UpstreamStats{}
	since := s.resolveUpstreamCutoff(ctx, start)

	s.emitLog("pipeline", "INFO", "[Upstream] Checking Nembus Cloud for unposted POS transactions, sales orders, and returns...")

	limit := s.cfg.BatchSize
	if limit <= 0 {
		limit = 50
	}

	var txErrors, orderErrors, returnErrors int

	approvedPOS, err := s.nembusClient.FetchApprovedPendingPOSTransactions(ctx, limit)
	if err != nil {
		return stats, fmt.Errorf("failed to fetch approved pending POS transactions: %w", err)
	}
	for _, tx := range approvedPOS {
		retryKey := "tx:" + tx.TransactionNumber
		if skip, nextAt, quarantined := s.shouldBackoff(retryKey, start); skip {
			if quarantined {
				stats.Quarantined++
				continue
			}
			s.emitLog("pos_transactions", "WARN", fmt.Sprintf("[Upstream] pos_transactions: backing off %s until %s", tx.TransactionNumber, nextAt.Format(time.RFC3339)))
			txErrors++
			stats.Errors++
			continue
		}
		if err := s.postTransactionToSAP(ctx, &tx, stockHoldPolicy{}); err != nil {
			quarantinedNow := s.recordRetryFailure(retryKey, start, err)
			if quarantinedNow {
				s.recordFailedDispatch(ctx, tx.TransactionNumber, tx.TransactionDate, tx.TotalAmount, err.Error())
				log.Printf("❌ Quarantined approved transaction %s (ID: %s) for %v after %d permanent SAP failures: %v", tx.TransactionNumber, tx.ID.String(), permanentQuarantineDelay, maxPermanentRetryAttempts, err)
			} else {
				log.Printf("❌ Failed to post approved transaction %s (ID: %s) to SAP: %v", tx.TransactionNumber, tx.ID.String(), err)
			}
			s.emitLog("pos_transactions", "ERROR", fmt.Sprintf("[Upstream] pos_transactions: failed to post approved %s: %v", tx.TransactionNumber, err))
			txErrors++
			stats.Errors++
		} else {
			s.clearRetry(retryKey)
			if !s.cfg.DryRun {
				stats.TransactionsPosted++
			}
		}
	}

	approvedOrders, err := s.nembusClient.FetchApprovedPendingSalesOrdersV2(ctx, limit)
	if err != nil {
		return stats, fmt.Errorf("failed to fetch approved pending sales orders: %w", err)
	}
	for _, order := range approvedOrders {
		retryKey := "order:" + order.OrderNumber
		if skip, nextAt, quarantined := s.shouldBackoff(retryKey, start); skip {
			if quarantined {
				stats.Quarantined++
				continue
			}
			s.emitLog("sales_orders", "WARN", fmt.Sprintf("[Upstream] sales_orders: backing off %s until %s", order.OrderNumber, nextAt.Format(time.RFC3339)))
			orderErrors++
			stats.Errors++
			continue
		}
		if err := s.postSalesOrderToSAP(ctx, &order, stockHoldPolicy{}); err != nil {
			quarantinedNow := s.recordRetryFailure(retryKey, start, err)
			if quarantinedNow {
				s.recordFailedDispatch(ctx, order.OrderNumber, order.OrderDate, order.TotalAmount, err.Error())
				log.Printf("❌ Quarantined approved sales order %s for %v after %d permanent SAP failures: %v", order.OrderNumber, permanentQuarantineDelay, maxPermanentRetryAttempts, err)
			} else {
				log.Printf("❌ Failed to post approved sales order %s to SAP: %v", order.OrderNumber, err)
			}
			s.emitLog("sales_orders", "ERROR", fmt.Sprintf("[Upstream] sales_orders: failed to post approved %s: %v", order.OrderNumber, err))
			orderErrors++
			stats.Errors++
		} else {
			s.clearRetry(retryKey)
			if !s.cfg.DryRun {
				stats.OrdersPosted++
			}
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
			retryKey := "tx:" + tx.TransactionNumber
			if skip, nextAt, quarantined := s.shouldBackoff(retryKey, start); skip {
				if quarantined {
					stats.Quarantined++
					continue
				}
				s.emitLog("pos_transactions", "WARN", fmt.Sprintf("[Upstream] pos_transactions: backing off %s until %s", tx.TransactionNumber, nextAt.Format(time.RFC3339)))
				txErrors++
				stats.Errors++
				continue
			}
			if err := s.postTransactionToSAP(ctx, &tx, policy); err != nil {
				if errors.Is(err, ErrTransactionHeldForReview) {
					s.clearRetry(retryKey)
					stats.TransactionsHeld++
					continue
				}
				if errors.Is(err, errNoLineItems) {
					s.clearRetry(retryKey)
					s.emitLog("pos_transactions", "WARN", fmt.Sprintf("[Upstream] pos_transactions: skipped %s (no line items)", tx.TransactionNumber))
					continue
				}
				quarantinedNow := s.recordRetryFailure(retryKey, start, err)
				if quarantinedNow {
					s.recordFailedDispatch(ctx, tx.TransactionNumber, tx.TransactionDate, tx.TotalAmount, err.Error())
					log.Printf("❌ Quarantined transaction %s (ID: %s) for %v after %d permanent SAP failures: %v", tx.TransactionNumber, tx.ID.String(), permanentQuarantineDelay, maxPermanentRetryAttempts, err)
				} else {
					log.Printf("❌ Failed to post transaction %s (ID: %s) to SAP: %v", tx.TransactionNumber, tx.ID.String(), err)
				}
				s.emitLog("pos_transactions", "ERROR", fmt.Sprintf("[Upstream] pos_transactions: failed to post %s: %v", tx.TransactionNumber, err))
				txErrors++
				stats.Errors++
			} else {
				s.clearRetry(retryKey)
				if !s.cfg.DryRun {
					stats.TransactionsPosted++
				}
			}
		}
	}
	if stats.TransactionsPosted == 0 && stats.TransactionsHeld == 0 && txErrors == 0 {
		s.emitLog("pos_transactions", "INFO", "[Upstream] pos_transactions: up to date (0 unposted transactions)")
	} else if txErrors > 0 {
		s.emitLog("pos_transactions", "WARN", fmt.Sprintf("[Upstream] pos_transactions: %d posted, %d held, %d failed", stats.TransactionsPosted, stats.TransactionsHeld, txErrors))
	} else {
		s.emitLog("pos_transactions", "INFO", fmt.Sprintf("[Upstream] pos_transactions: %d posted, %d held", stats.TransactionsPosted, stats.TransactionsHeld))
	}

	salesOrders, oErr := s.nembusClient.FetchUnpostedSalesOrdersV2(ctx, limit)
	if oErr != nil {
		log.Printf("⚠️ Failed to fetch unposted sales orders: %v", oErr)
		s.emitLog("sales_orders", "ERROR", fmt.Sprintf("[Upstream] sales_orders: failed to fetch unposted orders: %v", oErr))
		orderErrors++
		stats.Errors++
	} else if len(salesOrders) > 0 {
		policy := getStockPolicy()
		for _, order := range salesOrders {
			if beforeWatermark(order.OrderDate, since) {
				continue
			}
			retryKey := "order:" + order.OrderNumber
			if skip, nextAt, quarantined := s.shouldBackoff(retryKey, start); skip {
				if quarantined {
					stats.Quarantined++
					continue
				}
				s.emitLog("sales_orders", "WARN", fmt.Sprintf("[Upstream] sales_orders: backing off %s until %s", order.OrderNumber, nextAt.Format(time.RFC3339)))
				orderErrors++
				stats.Errors++
				continue
			}
			if err := s.postSalesOrderToSAP(ctx, &order, policy); err != nil {
				if errors.Is(err, ErrOrderHeldForReview) {
					s.clearRetry(retryKey)
					stats.OrdersHeld++
					continue
				}
				if errors.Is(err, errNoLineItems) {
					s.clearRetry(retryKey)
					s.emitLog("sales_orders", "WARN", fmt.Sprintf("[Upstream] sales_orders: skipped %s (no line items)", order.OrderNumber))
					continue
				}
				quarantinedNow := s.recordRetryFailure(retryKey, start, err)
				if quarantinedNow {
					s.recordFailedDispatch(ctx, order.OrderNumber, order.OrderDate, order.TotalAmount, err.Error())
					log.Printf("❌ Quarantined sales order %s for %v after %d permanent SAP failures: %v", order.OrderNumber, permanentQuarantineDelay, maxPermanentRetryAttempts, err)
				} else {
					log.Printf("❌ Failed to post sales order %s to SAP: %v", order.OrderNumber, err)
				}
				s.emitLog("sales_orders", "ERROR", fmt.Sprintf("[Upstream] sales_orders: failed to post %s: %v", order.OrderNumber, err))
				orderErrors++
				stats.Errors++
			} else {
				s.clearRetry(retryKey)
				if !s.cfg.DryRun {
					stats.OrdersPosted++
				}
			}
		}
	}
	if stats.OrdersPosted == 0 && stats.OrdersHeld == 0 && orderErrors == 0 {
		s.emitLog("sales_orders", "INFO", "[Upstream] sales_orders: up to date (0 unposted orders)")
	} else if orderErrors > 0 {
		s.emitLog("sales_orders", "WARN", fmt.Sprintf("[Upstream] sales_orders: %d posted, %d held, %d failed", stats.OrdersPosted, stats.OrdersHeld, orderErrors))
	} else {
		s.emitLog("sales_orders", "INFO", fmt.Sprintf("[Upstream] sales_orders: %d posted, %d held", stats.OrdersPosted, stats.OrdersHeld))
	}

	// Fetch unposted completed sales returns from Nembus
	returns, rErr := s.nembusClient.FetchUnpostedSalesReturns(ctx, limit)
	if rErr != nil {
		log.Printf("⚠️ Failed to fetch unposted sales returns: %v", rErr)
		s.emitLog("sales_returns", "ERROR", fmt.Sprintf("[Upstream] sales_returns: failed to fetch unposted returns: %v", rErr))
		returnErrors++
		stats.Errors++
	} else if len(returns) > 0 {
		for _, ret := range returns {
			if beforeWatermark(ret.ReturnDate, since) {
				continue
			}
			retryKey := "ret:" + ret.ReturnNumber
			if skip, nextAt, quarantined := s.shouldBackoff(retryKey, start); skip {
				if quarantined {
					stats.Quarantined++
					continue
				}
				s.emitLog("sales_returns", "WARN", fmt.Sprintf("[Upstream] sales_returns: backing off %s until %s", ret.ReturnNumber, nextAt.Format(time.RFC3339)))
				returnErrors++
				stats.Errors++
				continue
			}
			if err := s.postSalesReturnToSAP(ctx, &ret); err != nil {
				if errors.Is(err, errNoLineItems) {
					s.clearRetry(retryKey)
					log.Printf("⚠️ Skipping sales return %s (ID: %s): %v", ret.ReturnNumber, ret.ID.String(), err)
					s.emitLog("sales_returns", "WARN", fmt.Sprintf("[Upstream] sales_returns: skipped %s: %v", ret.ReturnNumber, err))
					continue
				}
				quarantinedNow := s.recordRetryFailure(retryKey, start, err)
				if quarantinedNow {
					s.recordFailedDispatch(ctx, ret.ReturnNumber, ret.ReturnDate, ret.TotalRefundAmount, err.Error())
					log.Printf("❌ Quarantined sales return %s (ID: %s) for %v after %d permanent SAP failures: %v", ret.ReturnNumber, ret.ID.String(), permanentQuarantineDelay, maxPermanentRetryAttempts, err)
				} else {
					log.Printf("❌ Failed to post sales return %s (ID: %s) to SAP: %v", ret.ReturnNumber, ret.ID.String(), err)
				}
				s.emitLog("sales_returns", "ERROR", fmt.Sprintf("[Upstream] sales_returns: failed to post %s: %v", ret.ReturnNumber, err))
				returnErrors++
				stats.Errors++
			} else {
				s.clearRetry(retryKey)
				if !s.cfg.DryRun {
					stats.ReturnsPosted++
				}
			}
		}
	}
	if stats.ReturnsPosted == 0 && returnErrors == 0 {
		s.emitLog("sales_returns", "INFO", "[Upstream] sales_returns: up to date (0 unposted returns)")
	} else if returnErrors > 0 {
		s.emitLog("sales_returns", "WARN", fmt.Sprintf("[Upstream] sales_returns: %d posted, %d failed", stats.ReturnsPosted, returnErrors))
	} else {
		s.emitLog("sales_returns", "INFO", fmt.Sprintf("[Upstream] sales_returns: %d posted", stats.ReturnsPosted))
	}

	if s.sqlite != nil && stats.Errors == 0 && !s.cfg.DryRun {
		_ = s.sqlite.SetUpstreamWatermark(ctx, "upstream", start, stats.TransactionsPosted+stats.OrdersPosted+stats.ReturnsPosted)
	}

	stats.Duration = time.Since(start)
	summaryLevel := "INFO"
	if stats.Errors > 0 {
		summaryLevel = "WARN"
	}
	s.emitLog("pipeline", summaryLevel, fmt.Sprintf("Upstream sync complete in %v: %d transactions posted, %d orders posted, %d returns posted, %d held, %d errors",
		stats.Duration.Round(time.Millisecond), stats.TransactionsPosted, stats.OrdersPosted, stats.ReturnsPosted, stats.TransactionsHeld+stats.OrdersHeld, stats.Errors))
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
		// Use PriceAfterVAT (gross unit price net of line discount) without UnitPrice so SAP
		// computes the exact gross LineTotal and back-calculates VAT cleanly without 1.15
		// rounding drift or double-discounting. Fall back to UnitPrice only when LineTotal <= 0.
		var unitPriceNet, priceAfterVAT float64
		if line.LineTotal > 0 && line.Quantity > 0 {
			priceAfterVAT = line.LineTotal / line.Quantity
		} else {
			unitPriceNet = line.UnitPrice
		}

		lineNum := idx
		docLine := sap.SAPDocumentLine{
			LineNum:         &lineNum,
			ItemCode:        line.ProductSKU,
			ItemDescription: line.ProductName,
			Quantity:        line.Quantity,
			UnitPrice:       unitPriceNet,
			PriceAfterVAT:   priceAfterVAT,
			TaxCode:         strings.TrimSpace(line.SAPTaxCode),
			WarehouseCode:   s.cfg.NembusDefaultWarehouse,
		}
		applySAPUoM(&docLine, line.SAPUoMEntry, line.SAPUgpEntry, line.UoMConversionFactor, line.IsBaseUoM)
		if err := s.attachBinAllocations(ctx, &docLine, lineNum, false); err != nil {
			return err
		}
		docLines = append(docLines, docLine)
	}

	// 3. Prepare Invoice Header
	cardCode, cardErr := s.resolveDocumentCardCode(ctx, tx.CustomerSAPCode, tx.CustomerID)
	if cardErr != nil {
		return cardErr
	}

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
	s.applyDocumentHeaderMeta(ctx, invReq, "13")

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
		if policy.requiresApproval && !skipSAPStockPreflight(tx.Metadata) && isSAPNegativeInventoryError(err) {
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
			if holdErr := s.nembusClient.HoldTransactionForStockReview(ctx, tx.ID, tx.TransactionNumber, tx.StoreID, whCode, tx.TotalAmount, fallbackShortages, summary); holdErr == nil {
				if s.sqlite != nil {
					_ = s.sqlite.RecordUpstreamDispatch(ctx, tx.TransactionNumber, tx.TransactionDate, tx.TotalAmount, nil, nil, "held", summary)
				}
				return ErrTransactionHeldForReview
			}
		}
		if s.sqlite != nil {
			_ = s.sqlite.RecordUpstreamDispatch(ctx, tx.TransactionNumber, tx.TransactionDate, tx.TotalAmount, nil, nil, "failed", err.Error())
		}
		return err
	}

	log.Printf(" Invoice created in SAP B1! DocEntry: %d, DocNum: %d (Ref: %s)",
		sapInv.DocEntry, sapInv.DocNum, tx.TransactionNumber)

	// Record local dispatch immediately after SAP PostInvoice succeeds to prevent duplicate posting
	// if incoming payment or cloud write-back encounters a transient failure.
	if s.sqlite != nil {
		_ = s.sqlite.RecordUpstreamDispatch(ctx, tx.TransactionNumber, tx.TransactionDate, tx.TotalAmount, &sapInv.DocEntry, &sapInv.DocNum, "posted", "")
	}

	// 5. Post matching payment in SAP if paid
	if tx.AmountPaid > 0 || tx.TotalAmount > 0 {
		payments, err := s.nembusClient.FetchTransactionPayments(ctx, tx.ID)
		if err == nil && len(payments) > 0 {
			var totalPaid float64
			remaining := sapInv.DocTotal
			for _, p := range payments {
				totalPaid += p.Amount
				sumApplied := p.Amount
				if sapInv.DocTotal > 0 {
					if remaining <= 0.001 {
						continue
					}
					if sumApplied > remaining {
						sumApplied = remaining
					}
					remaining -= sumApplied
				}
				payReq := &sap.SAPPaymentRequest{
					CardCode: cardCode,
					DocDate:  invDate,
					DocType:  "rCustomer",
					PaymentInvoices: []sap.SAPPendingPaymentInvoice{
						{
							LineNum:     0,
							DocEntry:    sapInv.DocEntry,
							SumApplied:  sumApplied,
							InvoiceType: "it_Invoice",
						},
					},
					Remarks: fmt.Sprintf("POS Payment for %s (Method: %s)", tx.TransactionNumber, p.PaymentMethod),
				}
				bplID := 0
				if s.cfg != nil && s.cfg.SAPBranchID > 0 {
					bplID = s.cfg.SAPBranchID
					payReq.BPLID = &bplID
				}
				if paySeries, sErr := s.sapClient.ResolveDocumentSeries(ctx, "24", invDate, bplID); sErr == nil && paySeries != nil && *paySeries > 0 {
					payReq.Series = paySeries
				}

				if strings.EqualFold(strings.TrimSpace(p.PaymentMethod), "cash") {
					payReq.CashSum = sumApplied
					if s.cfg != nil && strings.TrimSpace(s.cfg.SAPCashAccount) != "" {
						payReq.CashAccount = strings.TrimSpace(s.cfg.SAPCashAccount)
					}
				} else {
					payReq.TransferSum = sumApplied
					payReq.TransferDate = invDate
					transferAcct := ""
					if s.cfg != nil {
						transferAcct = strings.TrimSpace(s.cfg.SAPBankTransferAccount)
					}
					if transferAcct == "" {
						if resolver, ok := s.sapClient.(interface {
							ResolveDefaultTransferAccount(context.Context) (string, error)
						}); ok {
							if resolved, rErr := resolver.ResolveDefaultTransferAccount(ctx); rErr == nil {
								transferAcct = strings.TrimSpace(resolved)
							}
						}
					}
					if transferAcct != "" {
						payReq.TransferAccount = transferAcct
					}
					if p.Reference != nil && strings.TrimSpace(*p.Reference) != "" {
						payReq.TransferReference = strings.TrimSpace(*p.Reference)
					} else {
						payReq.TransferReference = tx.TransactionNumber
					}
				}

				if payResp, payErr := s.sapClient.PostIncomingPayment(ctx, payReq); payErr == nil {
					log.Printf("  Incoming Payment created in SAP! DocEntry: %d, Sum: %.2f",
						payResp.DocEntry, sumApplied)
				} else {
					log.Printf("⚠️ Warning: Failed to post payment for DocEntry %d: %v", sapInv.DocEntry, payErr)
					s.emitLog("pos_transactions", "WARN", fmt.Sprintf("[Upstream] pos_transactions: failed to post incoming payment for %s (DocEntry %d): %v", tx.TransactionNumber, sapInv.DocEntry, payErr))
				}
			}
			if sapInv.DocTotal > 0 && math.Abs(sapInv.DocTotal-totalPaid) > 0.01 {
				s.emitLog("pos_transactions", "WARN", fmt.Sprintf("[Upstream] pos_transactions: %s SAP DocTotal (%.2f, VatSum %.2f) differs from payments sum (%.2f)",
					tx.TransactionNumber, sapInv.DocTotal, sapInv.VatSum, totalPaid))
			}
		}
	}

	// 6. Mark transaction as synced in Nembus
	return s.nembusClient.MarkTransactionPostedToSAP(ctx, tx.ID, sapInv.DocEntry, sapInv.DocNum)
}

func (s *UpstreamSync) resolveDocumentCardCode(ctx context.Context, customerSAPCode string, customerID *uuid.UUID) (string, error) {
	if trimmed := strings.TrimSpace(customerSAPCode); trimmed != "" {
		return trimmed, nil
	}
	return s.resolveCardCode(ctx, customerID)
}

func (s *UpstreamSync) resolveCardCode(ctx context.Context, customerID *uuid.UUID) (string, error) {
	defaultCustomer := ""
	if s.cfg != nil {
		defaultCustomer = strings.TrimSpace(s.cfg.NembusDefaultCustomer)
	}
	if customerID == nil || *customerID == uuid.Nil {
		if defaultCustomer != "" {
			return defaultCustomer, nil
		}
		return "", fmt.Errorf("missing customer SAP CardCode and NEMBUS_DEFAULT_CUSTOMER is not configured")
	}
	code, err := s.nembusClient.FetchCustomerSAPCode(ctx, *customerID)
	if err != nil || strings.TrimSpace(code) == "" {
		if defaultCustomer != "" {
			return defaultCustomer, nil
		}
		return "", fmt.Errorf("customer %s has no SAP CardCode and NEMBUS_DEFAULT_CUSTOMER is not configured", customerID.String())
	}
	return strings.TrimSpace(code), nil
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
	if ret.OriginalSAPDocEntry > 0 {
		entry := ret.OriginalSAPDocEntry
		baseDocEntry = &entry
	} else if ret.OriginalTransactionID != nil && *ret.OriginalTransactionID != uuid.Nil {
		if entry, err := s.nembusClient.FetchOriginalSAPDocEntry(ctx, *ret.OriginalTransactionID); err == nil && entry > 0 {
			baseDocEntry = &entry
		} else {
			log.Printf("⚠️ [SalesReturn %s] Original transaction ID %s has no SAP DocEntry (err: %v); posting as standalone credit memo", ret.ReturnNumber, ret.OriginalTransactionID.String(), err)
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
			TaxCode:         strings.TrimSpace(line.SAPTaxCode),
			WarehouseCode:   s.cfg.NembusDefaultWarehouse,
		}
		if baseDocEntry != nil {
			baseType := 13 // 13 = A/R Invoice
			docLine.BaseEntry = baseDocEntry
			docLine.BaseType = &baseType
			if baseLine, blErr := s.sapClient.ResolveInvoiceBaseLine(ctx, *baseDocEntry, line.ProductSKU); blErr == nil && baseLine != nil {
				docLine.BaseLine = baseLine
			}
		}
		applySAPUoM(&docLine, line.SAPUoMEntry, line.SAPUgpEntry, line.UoMConversionFactor, line.IsBaseUoM)
		if err := s.attachBinAllocations(ctx, &docLine, lineNum, true); err != nil {
			return err
		}
		docLines = append(docLines, docLine)
	}

	cardCode, cardErr := s.resolveDocumentCardCode(ctx, ret.CustomerSAPCode, ret.CustomerID)
	if cardErr != nil {
		return cardErr
	}
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
	s.applyDocumentHeaderMeta(ctx, memoReq, "14")

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

	if s.sqlite != nil {
		_ = s.sqlite.RecordUpstreamDispatch(ctx, ret.ReturnNumber, ret.ReturnDate, ret.TotalRefundAmount, &sapMemo.DocEntry, &sapMemo.DocNum, "posted", "")
	}

	return s.nembusClient.MarkSalesReturnPostedToSAP(ctx, ret.ID, sapMemo.DocEntry, sapMemo.DocNum)
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
		var unitPriceNet, priceAfterVAT float64
		if line.LineTotal > 0 {
			priceAfterVAT = line.LineTotal / line.Quantity
		} else {
			unitPriceNet = line.UnitPrice
		}
		lineNum := idx
		docLine := sap.SAPDocumentLine{
			LineNum:         &lineNum,
			ItemCode:        line.ProductSKU,
			ItemDescription: line.ProductName,
			Quantity:        line.Quantity,
			UnitPrice:       unitPriceNet,
			PriceAfterVAT:   priceAfterVAT,
			TaxCode:         strings.TrimSpace(line.SAPTaxCode),
			WarehouseCode:   s.cfg.NembusDefaultWarehouse,
		}
		applySAPUoM(&docLine, line.SAPUoMEntry, line.SAPUgpEntry, line.UoMConversionFactor, line.IsBaseUoM)
		if err := s.attachBinAllocations(ctx, &docLine, lineNum, false); err != nil {
			return err
		}
		docLines = append(docLines, docLine)
	}
	if len(docLines) == 0 {
		return fmt.Errorf("sales order %s has no billable lines", order.OrderNumber)
	}

	cardCode, cardErr := s.resolveDocumentCardCode(ctx, order.CustomerSAPCode, order.CustomerID)
	if cardErr != nil {
		return cardErr
	}
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
	s.applyDocumentHeaderMeta(ctx, invReq, "13")

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
		if policy.requiresApproval && !skipSAPStockPreflight(order.Metadata) && isSAPNegativeInventoryError(err) {
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
			if holdErr := s.nembusClient.HoldSalesOrderForStockReview(ctx, order.ID, order.OrderNumber, order.StoreID, whCode, order.TotalAmount, fallback, summary); holdErr == nil {
				if s.sqlite != nil {
					_ = s.sqlite.RecordUpstreamDispatch(ctx, order.OrderNumber, order.OrderDate, order.TotalAmount, nil, nil, "held", summary)
				}
				return ErrOrderHeldForReview
			}
		}
		if s.sqlite != nil {
			_ = s.sqlite.RecordUpstreamDispatch(ctx, order.OrderNumber, order.OrderDate, order.TotalAmount, nil, nil, "failed", err.Error())
		}
		return err
	}

	log.Printf(" Invoice created in SAP B1 from sales order! DocEntry: %d, DocNum: %d (Ref: %s)",
		sapInv.DocEntry, sapInv.DocNum, order.OrderNumber)

	if s.sqlite != nil {
		_ = s.sqlite.RecordUpstreamDispatch(ctx, order.OrderNumber, order.OrderDate, order.TotalAmount, &sapInv.DocEntry, &sapInv.DocNum, "posted", "")
	}

	return s.nembusClient.MarkSalesOrderPostedToSAP(ctx, order.ID, sapInv.DocEntry, sapInv.DocNum)
}
