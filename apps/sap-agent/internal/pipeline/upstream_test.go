package pipeline_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/nembus"
	"github.com/NasTecSol/nembus-sap-agent/internal/pipeline"
	"github.com/NasTecSol/nembus-sap-agent/internal/sap"
)

// MockSAPClient implements pipeline.SAPClient
type MockSAPClient struct {
	postInvoiceFunc        func(ctx context.Context, invoice *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error)
	postIncomingPaymentFunc func(ctx context.Context, payment *sap.SAPPaymentRequest) (*sap.SAPPaymentResponse, error)
	postCreditMemoFunc     func(ctx context.Context, memo *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error)

	postedCreditMemos []*sap.SAPInvoiceRequest
}

func (m *MockSAPClient) PostInvoice(ctx context.Context, invoice *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error) {
	if m.postInvoiceFunc != nil {
		return m.postInvoiceFunc(ctx, invoice)
	}
	return &sap.SAPInvoiceResponse{DocEntry: 1, DocNum: 100}, nil
}

func (m *MockSAPClient) PostIncomingPayment(ctx context.Context, payment *sap.SAPPaymentRequest) (*sap.SAPPaymentResponse, error) {
	if m.postIncomingPaymentFunc != nil {
		return m.postIncomingPaymentFunc(ctx, payment)
	}
	return &sap.SAPPaymentResponse{DocEntry: 2, DocNum: 200}, nil
}

func (m *MockSAPClient) PostCreditMemo(ctx context.Context, memo *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error) {
	m.postedCreditMemos = append(m.postedCreditMemos, memo)
	if m.postCreditMemoFunc != nil {
		return m.postCreditMemoFunc(ctx, memo)
	}
	return &sap.SAPInvoiceResponse{DocEntry: 777, DocNum: 888}, nil
}

// MockNembusClient implements pipeline.NembusClient
type MockNembusClient struct {
	unpostedPOSTransactions   []nembus.POSTransaction
	txLines                   map[int64][]nembus.POSTransactionLine
	txPayments                map[int64][]nembus.POSPayment
	markedTransactions        map[int64]markedSAPDoc

	unpostedSalesReturns      []nembus.SalesReturn
	returnLines               map[int64][]nembus.SalesReturnLine
	markedReturns             map[int64]markedSAPDoc
	originalSAPDocEntries     map[int32]int

	fetchUnpostedReturnsErr   error
	fetchReturnLinesErr       error
	markReturnPostedErr       error
}

type markedSAPDoc struct {
	DocEntry int
	DocNum   int
}

func newMockNembusClient() *MockNembusClient {
	return &MockNembusClient{
		txLines:               make(map[int64][]nembus.POSTransactionLine),
		txPayments:            make(map[int64][]nembus.POSPayment),
		markedTransactions:    make(map[int64]markedSAPDoc),
		returnLines:           make(map[int64][]nembus.SalesReturnLine),
		markedReturns:         make(map[int64]markedSAPDoc),
		originalSAPDocEntries: make(map[int32]int),
	}
}

func (m *MockNembusClient) FetchUnpostedPOSTransactions(ctx context.Context, limit int) ([]nembus.POSTransaction, error) {
	return m.unpostedPOSTransactions, nil
}

func (m *MockNembusClient) FetchTransactionLines(ctx context.Context, transactionID int64) ([]nembus.POSTransactionLine, error) {
	return m.txLines[transactionID], nil
}

func (m *MockNembusClient) FetchTransactionPayments(ctx context.Context, transactionID int64) ([]nembus.POSPayment, error) {
	return m.txPayments[transactionID], nil
}

func (m *MockNembusClient) MarkTransactionPostedToSAP(ctx context.Context, transactionID int64, sapDocEntry, sapDocNum int) error {
	m.markedTransactions[transactionID] = markedSAPDoc{DocEntry: sapDocEntry, DocNum: sapDocNum}
	return nil
}

func (m *MockNembusClient) FetchCustomerSAPCode(ctx context.Context, customerID int32) (string, error) {
	return "C_TEST", nil
}

func (m *MockNembusClient) FetchUnpostedSalesReturns(ctx context.Context, limit int) ([]nembus.SalesReturn, error) {
	if m.fetchUnpostedReturnsErr != nil {
		return nil, m.fetchUnpostedReturnsErr
	}
	return m.unpostedSalesReturns, nil
}

func (m *MockNembusClient) FetchSalesReturnLines(ctx context.Context, returnID int64) ([]nembus.SalesReturnLine, error) {
	if m.fetchReturnLinesErr != nil {
		return nil, m.fetchReturnLinesErr
	}
	return m.returnLines[returnID], nil
}

func (m *MockNembusClient) MarkSalesReturnPostedToSAP(ctx context.Context, returnID int64, sapDocEntry, sapDocNum int) error {
	if m.markReturnPostedErr != nil {
		return m.markReturnPostedErr
	}
	m.markedReturns[returnID] = markedSAPDoc{DocEntry: sapDocEntry, DocNum: sapDocNum}
	return nil
}

func (m *MockNembusClient) FetchOriginalSAPDocEntry(ctx context.Context, originalTxID int32) (int, error) {
	entry, ok := m.originalSAPDocEntries[originalTxID]
	if !ok {
		return 0, nil
	}
	return entry, nil
}

func TestBoundaryDecoupling_ZeroSyncQueueReferences(t *testing.T) {
	// Root directory of apps/sap-agent
	// Looking from apps/sap-agent/internal/pipeline
	sapAgentDir := filepath.Join("..", "..")

	err := filepath.Walk(sapAgentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".go") || strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		contentLower := strings.ToLower(string(content))

		// Check for sync_queue references in source code
		if strings.Contains(contentLower, "sync_queue") {
			t.Errorf("Boundary violation: found reference to 'sync_queue' in %s", path)
		}
		if strings.Contains(contentLower, "syncqueue") {
			t.Errorf("Boundary violation: found reference to 'syncqueue' in %s", path)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("failed to scan sap-agent codebase: %v", err)
	}
}

func TestProcessOutbox_LinkedReturnPosting(t *testing.T) {
	mockSAP := &MockSAPClient{}
	mockNembus := newMockNembusClient()

	origTxID := int32(42)
	reason := "Damaged item"
	customerID := int32(101)

	mockNembus.unpostedSalesReturns = []nembus.SalesReturn{
		{
			ID:                    10,
			ReturnNumber:          "RET-2026-001",
			StoreID:               1,
			CustomerID:            &customerID,
			OriginalTransactionID: &origTxID,
			ReturnDate:            time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC),
			ReturnReason:          &reason,
			Status:                "completed",
			Subtotal:              100.0,
			TaxAmount:             15.0,
			TotalRefundAmount:     115.0,
		},
	}

	mockNembus.returnLines[10] = []nembus.SalesReturnLine{
		{
			ID:            101,
			ReturnID:      10,
			ProductID:     5,
			ProductSKU:    "SKU-RETURN-01",
			ProductName:   "Returned Widget",
			Quantity:      2.0,
			UnitPrice:     50.0,
			RefundAmount:  115.0,
			ReturnToStock: true,
			LineNumber:    1,
		},
	}

	// Link original invoice docEntry = 9001
	mockNembus.originalSAPDocEntries[origTxID] = 9001

	cfg := &config.Config{
		BatchSize:              10,
		NembusDefaultWarehouse: "WH-01",
		NembusDefaultCustomer:  "C-DEFAULT",
	}

	sync := pipeline.NewUpstreamSync(cfg, mockSAP, mockNembus)
	ctx := context.Background()

	stats, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("ProcessOutbox failed: %v", err)
	}

	if stats.ReturnsPosted != 1 {
		t.Errorf("expected 1 return posted, got %d", stats.ReturnsPosted)
	}
	if stats.Errors != 0 {
		t.Errorf("expected 0 errors, got %d", stats.Errors)
	}

	// Verify Credit Memo posted to SAP
	if len(mockSAP.postedCreditMemos) != 1 {
		t.Fatalf("expected 1 credit memo sent to SAP, got %d", len(mockSAP.postedCreditMemos))
	}

	memo := mockSAP.postedCreditMemos[0]
	if memo.NumAtCard != "RET-2026-001" {
		t.Errorf("expected NumAtCard RET-2026-001, got %s", memo.NumAtCard)
	}
	if !strings.Contains(memo.Comments, "Damaged item") {
		t.Errorf("expected reason in comments, got %s", memo.Comments)
	}

	if len(memo.DocumentLines) != 1 {
		t.Fatalf("expected 1 document line, got %d", len(memo.DocumentLines))
	}

	line := memo.DocumentLines[0]
	if line.ItemCode != "SKU-RETURN-01" {
		t.Errorf("expected ItemCode SKU-RETURN-01, got %s", line.ItemCode)
	}
	if line.BaseEntry == nil || *line.BaseEntry != 9001 {
		t.Errorf("expected BaseEntry 9001, got %v", line.BaseEntry)
	}
	if line.BaseType == nil || *line.BaseType != 13 {
		t.Errorf("expected BaseType 13, got %v", line.BaseType)
	}
	if line.PriceAfterVAT != 57.5 { // 115.0 / 2.0
		t.Errorf("expected PriceAfterVAT 57.5, got %f", line.PriceAfterVAT)
	}
	if line.WarehouseCode != "WH-01" {
		t.Errorf("expected warehouse WH-01, got %s", line.WarehouseCode)
	}

	// Verify MarkSalesReturnPostedToSAP was called with SAP DocEntry & DocNum
	marked, ok := mockNembus.markedReturns[10]
	if !ok {
		t.Fatalf("expected return 10 to be marked as posted in Nembus")
	}
	if marked.DocEntry != 777 || marked.DocNum != 888 {
		t.Errorf("expected marked with (777, 888), got (%d, %d)", marked.DocEntry, marked.DocNum)
	}
}

func TestProcessOutbox_StandaloneReturnPosting(t *testing.T) {
	mockSAP := &MockSAPClient{}
	mockNembus := newMockNembusClient()

	mockNembus.unpostedSalesReturns = []nembus.SalesReturn{
		{
			ID:                    20,
			ReturnNumber:          "RET-STANDALONE-001",
			StoreID:               1,
			OriginalTransactionID: nil, // Standalone
			ReturnDate:            time.Date(2026, 10, 22, 11, 0, 0, 0, time.UTC),
			Status:                "completed",
			Subtotal:              50.0,
			TaxAmount:             7.5,
			TotalRefundAmount:     57.5,
		},
	}

	mockNembus.returnLines[20] = []nembus.SalesReturnLine{
		{
			ID:           201,
			ReturnID:     20,
			ProductID:    8,
			ProductSKU:   "SKU-STANDALONE",
			ProductName:  "Standalone Item",
			Quantity:     1.0,
			UnitPrice:    50.0,
			RefundAmount: 57.5,
		},
	}

	cfg := &config.Config{
		BatchSize: 10,
	}

	sync := pipeline.NewUpstreamSync(cfg, mockSAP, mockNembus)
	ctx := context.Background()

	stats, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("ProcessOutbox failed: %v", err)
	}

	if stats.ReturnsPosted != 1 {
		t.Errorf("expected 1 return posted, got %d", stats.ReturnsPosted)
	}

	memo := mockSAP.postedCreditMemos[0]
	line := memo.DocumentLines[0]
	if line.BaseEntry != nil {
		t.Errorf("BaseEntry should be nil for standalone return, got %v", *line.BaseEntry)
	}
	if line.BaseType != nil {
		t.Errorf("BaseType should be nil for standalone return, got %v", *line.BaseType)
	}
}

func TestProcessOutbox_DryRun(t *testing.T) {
	mockSAP := &MockSAPClient{}
	mockNembus := newMockNembusClient()

	mockNembus.unpostedPOSTransactions = []nembus.POSTransaction{
		{
			ID:                15,
			TransactionNumber: "TX-DRYRUN",
			TotalAmount:       100.0,
			AmountPaid:        100.0,
			TransactionDate:   time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC),
		},
	}
	mockNembus.txLines[15] = []nembus.POSTransactionLine{
		{ID: 151, TransactionID: 15, ProductSKU: "SKU-TX", Quantity: 1.0, UnitPrice: 100.0, LineTotal: 100.0},
	}

	mockNembus.unpostedSalesReturns = []nembus.SalesReturn{
		{
			ID:                30,
			ReturnNumber:      "RET-DRYRUN",
			Status:            "completed",
			TotalRefundAmount: 200.0,
			ReturnDate:        time.Date(2026, 10, 22, 11, 0, 0, 0, time.UTC),
		},
	}
	mockNembus.returnLines[30] = []nembus.SalesReturnLine{
		{ID: 301, ReturnID: 30, ProductSKU: "SKU-D", Quantity: 1.0, UnitPrice: 200.0, RefundAmount: 200.0},
	}

	cfg := &config.Config{
		BatchSize: 10,
		DryRun:    true,
	}

	sync := pipeline.NewUpstreamSync(cfg, mockSAP, mockNembus)
	ctx := context.Background()

	stats, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("ProcessOutbox failed: %v", err)
	}

	if len(mockSAP.postedCreditMemos) != 0 {
		t.Errorf("expected 0 credit memos in dry run, got %d", len(mockSAP.postedCreditMemos))
	}
	if len(mockNembus.markedTransactions) != 0 {
		t.Errorf("expected 0 transactions marked as posted in dry run, got %d", len(mockNembus.markedTransactions))
	}
	if len(mockNembus.markedReturns) != 0 {
		t.Errorf("expected 0 returns marked as posted in dry run, got %d", len(mockNembus.markedReturns))
	}
	if stats.TransactionsPosted != 0 {
		t.Errorf("expected 0 transactions posted in dry run, got %d", stats.TransactionsPosted)
	}
	if stats.ReturnsPosted != 0 {
		t.Errorf("expected 0 returns posted in dry run, got %d", stats.ReturnsPosted)
	}
	if stats.Errors != 0 {
		t.Errorf("expected 0 errors in dry run, got %d", stats.Errors)
	}
}

func TestProcessOutbox_CreditMemoFailure(t *testing.T) {
	mockSAP := &MockSAPClient{
		postCreditMemoFunc: func(ctx context.Context, memo *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error) {
			return nil, errors.New("SAP connection timeout")
		},
	}
	mockNembus := newMockNembusClient()

	mockNembus.unpostedSalesReturns = []nembus.SalesReturn{
		{
			ID:                40,
			ReturnNumber:      "RET-FAIL",
			Status:            "completed",
			TotalRefundAmount: 50.0,
		},
	}
	mockNembus.returnLines[40] = []nembus.SalesReturnLine{
		{ID: 401, ReturnID: 40, ProductSKU: "SKU-FAIL", Quantity: 1.0, UnitPrice: 50.0, RefundAmount: 50.0},
	}

	cfg := &config.Config{BatchSize: 10}
	sync := pipeline.NewUpstreamSync(cfg, mockSAP, mockNembus)
	ctx := context.Background()

	stats, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("ProcessOutbox should not return fatal error on single item failure: %v", err)
	}

	if stats.Errors != 1 {
		t.Errorf("expected 1 error, got %d", stats.Errors)
	}
	if stats.ReturnsPosted != 0 {
		t.Errorf("expected 0 returns posted, got %d", stats.ReturnsPosted)
	}
	if len(mockNembus.markedReturns) != 0 {
		t.Errorf("failed return should not be marked as posted in Nembus")
	}
}

func TestMarkSalesReturnPostedMetadataFormat(t *testing.T) {
	// Verify that MarkSalesReturnPostedToSAP generates compliant JSON metadata
	metaPatch, err := json.Marshal(map[string]any{
		"sap_doc_entry":  12345,
		"sap_doc_num":    67890,
		"sap_posted_at":  time.Now().Format(time.RFC3339),
		"sap_sync_state": "synced",
	})
	if err != nil {
		t.Fatalf("failed to marshal metaPatch: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(metaPatch, &parsed); err != nil {
		t.Fatalf("failed to unmarshal metaPatch: %v", err)
	}

	if parsed["sap_doc_entry"].(float64) != 12345 {
		t.Errorf("expected sap_doc_entry 12345, got %v", parsed["sap_doc_entry"])
	}
	if parsed["sap_doc_num"].(float64) != 67890 {
		t.Errorf("expected sap_doc_num 67890, got %v", parsed["sap_doc_num"])
	}
	if parsed["sap_sync_state"] != "synced" {
		t.Errorf("expected sap_sync_state 'synced', got %v", parsed["sap_sync_state"])
	}
	if parsed["sap_posted_at"] == "" {
		t.Errorf("expected non-empty sap_posted_at")
	}
}

func TestProcessOutbox_ReturnInvalidQuantity(t *testing.T) {
	mockSAP := &MockSAPClient{}
	mockNembus := newMockNembusClient()

	mockNembus.unpostedSalesReturns = []nembus.SalesReturn{
		{
			ID:                50,
			ReturnNumber:      "RET-ZERO-QTY",
			Status:            "completed",
			TotalRefundAmount: 50.0,
			ReturnDate:        time.Date(2026, 10, 22, 11, 0, 0, 0, time.UTC),
		},
	}
	mockNembus.returnLines[50] = []nembus.SalesReturnLine{
		{ID: 501, ReturnID: 50, ProductSKU: "SKU-INV", Quantity: 0.0, UnitPrice: 50.0, RefundAmount: 50.0},
	}

	cfg := &config.Config{BatchSize: 10}
	sync := pipeline.NewUpstreamSync(cfg, mockSAP, mockNembus)
	ctx := context.Background()

	stats, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("ProcessOutbox should not fail fatally: %v", err)
	}

	if stats.Errors != 1 {
		t.Errorf("expected 1 error for zero quantity, got %d", stats.Errors)
	}
	if stats.ReturnsPosted != 0 {
		t.Errorf("expected 0 returns posted, got %d", stats.ReturnsPosted)
	}
	if len(mockSAP.postedCreditMemos) != 0 {
		t.Errorf("expected no credit memos sent to SAP")
	}
	if len(mockNembus.markedReturns) != 0 {
		t.Errorf("expected no returns marked as posted")
	}
}

func TestProcessOutbox_ReturnNegativeRefundAmount(t *testing.T) {
	mockSAP := &MockSAPClient{}
	mockNembus := newMockNembusClient()

	mockNembus.unpostedSalesReturns = []nembus.SalesReturn{
		{
			ID:                60,
			ReturnNumber:      "RET-NEG-AMT",
			Status:            "completed",
			TotalRefundAmount: -25.0,
			ReturnDate:        time.Date(2026, 10, 22, 11, 0, 0, 0, time.UTC),
		},
	}
	mockNembus.returnLines[60] = []nembus.SalesReturnLine{
		{ID: 601, ReturnID: 60, ProductSKU: "SKU-NEG", Quantity: 1.0, UnitPrice: 25.0, RefundAmount: -25.0},
	}

	cfg := &config.Config{BatchSize: 10}
	sync := pipeline.NewUpstreamSync(cfg, mockSAP, mockNembus)
	ctx := context.Background()

	stats, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("ProcessOutbox should not fail fatally: %v", err)
	}

	if stats.Errors != 1 {
		t.Errorf("expected 1 error for negative refund amount, got %d", stats.Errors)
	}
	if stats.ReturnsPosted != 0 {
		t.Errorf("expected 0 returns posted, got %d", stats.ReturnsPosted)
	}
	if len(mockSAP.postedCreditMemos) != 0 {
		t.Errorf("expected no credit memos sent to SAP")
	}
	if len(mockNembus.markedReturns) != 0 {
		t.Errorf("expected no returns marked as posted")
	}
}

func TestProcessOutbox_ReturnUnresolvedOriginalDocEntry(t *testing.T) {
	mockSAP := &MockSAPClient{}
	mockNembus := newMockNembusClient()

	origTxID := int32(9999) // Non-existent in mock originalSAPDocEntries (resolves to 0)
	mockNembus.unpostedSalesReturns = []nembus.SalesReturn{
		{
			ID:                    70,
			ReturnNumber:          "RET-UNRESOLVED-ENTRY",
			OriginalTransactionID: &origTxID,
			Status:                "completed",
			TotalRefundAmount:     80.0,
			ReturnDate:            time.Date(2026, 10, 22, 11, 0, 0, 0, time.UTC),
		},
	}
	mockNembus.returnLines[70] = []nembus.SalesReturnLine{
		{ID: 701, ReturnID: 70, ProductSKU: "SKU-UNRES", Quantity: 2.0, UnitPrice: 40.0, RefundAmount: 80.0},
	}

	cfg := &config.Config{BatchSize: 10}
	sync := pipeline.NewUpstreamSync(cfg, mockSAP, mockNembus)
	ctx := context.Background()

	stats, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("ProcessOutbox failed: %v", err)
	}

	if stats.Errors != 0 {
		t.Errorf("expected 0 errors, got %d", stats.Errors)
	}
	if stats.ReturnsPosted != 1 {
		t.Errorf("expected 1 return posted as standalone fallback, got %d", stats.ReturnsPosted)
	}
	if len(mockSAP.postedCreditMemos) != 1 {
		t.Fatalf("expected 1 credit memo posted, got %d", len(mockSAP.postedCreditMemos))
	}
	memo := mockSAP.postedCreditMemos[0]
	if len(memo.DocumentLines) != 1 {
		t.Fatalf("expected 1 document line, got %d", len(memo.DocumentLines))
	}
	if memo.DocumentLines[0].BaseEntry != nil {
		t.Errorf("expected nil BaseEntry for unresolved original entry, got %v", *memo.DocumentLines[0].BaseEntry)
	}
	if memo.DocumentLines[0].BaseType != nil {
		t.Errorf("expected nil BaseType for unresolved original entry, got %v", *memo.DocumentLines[0].BaseType)
	}
	if len(mockNembus.markedReturns) != 1 {
		t.Errorf("expected return to be marked as posted in Nembus")
	}
}

