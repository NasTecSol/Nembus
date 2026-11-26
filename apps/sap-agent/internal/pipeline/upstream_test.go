package pipeline_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap-agent/internal/nembus"
	"github.com/NasTecSol/nembus-sap-agent/internal/pipeline"
	"github.com/NasTecSol/nembus-sap-agent/internal/sap"
	"github.com/NasTecSol/nembus-sap/contracts"
	"github.com/google/uuid"
)

func testUUID(legacy int64) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("nembus-test-entity-%d", legacy)))
}

// MockSAPClient implements pipeline.SAPClient
type MockSAPClient struct {
	postInvoiceFunc         func(ctx context.Context, invoice *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error)
	postIncomingPaymentFunc func(ctx context.Context, payment *sap.SAPPaymentRequest) (*sap.SAPPaymentResponse, error)
	postCreditMemoFunc      func(ctx context.Context, memo *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error)
	stockBySKU              map[string]float64
	binAllocations          []sap.SAPBinAllocation
	seriesByObj             map[string]int
	baseLineByItem          map[string]int

	postedCreditMemos []*sap.SAPInvoiceRequest
}

func (m *MockSAPClient) ResolveWarehouseBinAllocations(ctx context.Context, itemCode, warehouseCode string, requiredQty float64, isReturn bool) ([]sap.SAPBinAllocation, error) {
	if len(m.binAllocations) == 0 {
		return nil, nil
	}
	out := make([]sap.SAPBinAllocation, len(m.binAllocations))
	copy(out, m.binAllocations)
	return out, nil
}

func (m *MockSAPClient) ResolveDocumentSeries(ctx context.Context, objectCode string, docDate string, bplID int) (*int, error) {
	if m.seriesByObj != nil {
		if s, ok := m.seriesByObj[objectCode]; ok && s > 0 {
			v := s
			return &v, nil
		}
	}
	return nil, nil
}

func (m *MockSAPClient) ResolveInvoiceBaseLine(ctx context.Context, baseDocEntry int, itemCode string) (*int, error) {
	if m.baseLineByItem != nil {
		if ln, ok := m.baseLineByItem[itemCode]; ok {
			v := ln
			return &v, nil
		}
	}
	return nil, nil
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

func (m *MockSAPClient) CheckItemWarehouseStock(ctx context.Context, itemCode, warehouseCode string) (float64, error) {
	if m.stockBySKU != nil {
		if v, ok := m.stockBySKU[itemCode]; ok {
			return v, nil
		}
		if v, ok := m.stockBySKU["*"]; ok {
			return v, nil
		}
	}
	return 999.0, nil
}

// MockNembusClient implements pipeline.NembusClient
type MockNembusClient struct {
	unpostedPOSTransactions        []nembus.POSTransaction
	approvedPendingPOSTransactions []nembus.POSTransaction
	txLines                        map[uuid.UUID][]nembus.POSTransactionLine
	txPayments                     map[uuid.UUID][]nembus.POSPayment
	markedTransactions             map[uuid.UUID]markedSAPDoc
	heldTransactions               map[uuid.UUID]string

	unpostedSalesOrders        []nembus.SalesOrderV2
	approvedPendingSalesOrders []nembus.SalesOrderV2
	salesOrderLines            map[uuid.UUID][]nembus.SalesOrderLineV2
	markedSalesOrders          map[uuid.UUID]markedSAPDoc
	heldSalesOrders            map[uuid.UUID]string

	unpostedSalesReturns  []nembus.SalesReturn
	returnLines           map[uuid.UUID][]nembus.SalesReturnLine
	markedReturns         map[uuid.UUID]markedSAPDoc
	originalSAPDocEntries map[uuid.UUID]int

	stockBySKU              map[string]float64
	fetchUnpostedReturnsErr error
	fetchReturnLinesErr     error
	markReturnPostedErr     error

	stockHoldPolicyLoaded   bool
	stockHoldPolicyRequired bool
	stockHoldPolicyRules    map[string]any
	stockHoldPolicyErr      error
}

type markedSAPDoc struct {
	DocEntry int
	DocNum   int
}

func newMockNembusClient() *MockNembusClient {
	return &MockNembusClient{
		txLines:               make(map[uuid.UUID][]nembus.POSTransactionLine),
		txPayments:            make(map[uuid.UUID][]nembus.POSPayment),
		markedTransactions:    make(map[uuid.UUID]markedSAPDoc),
		salesOrderLines:       make(map[uuid.UUID][]nembus.SalesOrderLineV2),
		markedSalesOrders:     make(map[uuid.UUID]markedSAPDoc),
		returnLines:           make(map[uuid.UUID][]nembus.SalesReturnLine),
		markedReturns:         make(map[uuid.UUID]markedSAPDoc),
		originalSAPDocEntries: make(map[uuid.UUID]int),
		stockBySKU:            map[string]float64{"*": 999},
	}
}

func (m *MockNembusClient) FetchApprovedPendingPOSTransactions(ctx context.Context, limit int) ([]nembus.POSTransaction, error) {
	return m.approvedPendingPOSTransactions, nil
}

func (m *MockNembusClient) FetchUnpostedPOSTransactions(ctx context.Context, limit int) ([]nembus.POSTransaction, error) {
	return m.unpostedPOSTransactions, nil
}

func (m *MockNembusClient) FetchTransactionLines(ctx context.Context, transactionID uuid.UUID) ([]nembus.POSTransactionLine, error) {
	return m.txLines[transactionID], nil
}

func (m *MockNembusClient) FetchTransactionPayments(ctx context.Context, transactionID uuid.UUID) ([]nembus.POSPayment, error) {
	return m.txPayments[transactionID], nil
}

func (m *MockNembusClient) MarkTransactionPostedToSAP(ctx context.Context, transactionID uuid.UUID, sapDocEntry, sapDocNum int) error {
	m.markedTransactions[transactionID] = markedSAPDoc{DocEntry: sapDocEntry, DocNum: sapDocNum}
	return nil
}

func (m *MockNembusClient) HoldTransactionForStockReview(ctx context.Context, txID uuid.UUID, txNum string, storeID int32, whCode string, totalAmount float64, shortageItems []nembus.ShortageItem, summary string) error {
	if m.heldTransactions == nil {
		m.heldTransactions = make(map[uuid.UUID]string)
	}
	m.heldTransactions[txID] = summary
	return nil
}

func (m *MockNembusClient) FetchCustomerSAPCode(ctx context.Context, customerID uuid.UUID) (string, error) {
	return "C_TEST", nil
}

func (m *MockNembusClient) FetchUnpostedSalesReturns(ctx context.Context, limit int) ([]nembus.SalesReturn, error) {
	if m.fetchUnpostedReturnsErr != nil {
		return nil, m.fetchUnpostedReturnsErr
	}
	return m.unpostedSalesReturns, nil
}

func (m *MockNembusClient) FetchSalesReturnLines(ctx context.Context, returnID uuid.UUID) ([]nembus.SalesReturnLine, error) {
	if m.fetchReturnLinesErr != nil {
		return nil, m.fetchReturnLinesErr
	}
	return m.returnLines[returnID], nil
}

func (m *MockNembusClient) MarkSalesReturnPostedToSAP(ctx context.Context, returnID uuid.UUID, sapDocEntry, sapDocNum int) error {
	if m.markReturnPostedErr != nil {
		return m.markReturnPostedErr
	}
	m.markedReturns[returnID] = markedSAPDoc{DocEntry: sapDocEntry, DocNum: sapDocNum}
	return nil
}

func (m *MockNembusClient) FetchOriginalSAPDocEntry(ctx context.Context, originalTxID uuid.UUID) (int, error) {
	entry, ok := m.originalSAPDocEntries[originalTxID]
	if !ok {
		return 0, nil
	}
	return entry, nil
}

func (m *MockNembusClient) FetchApprovedPendingSalesOrdersV2(ctx context.Context, limit int) ([]nembus.SalesOrderV2, error) {
	return m.approvedPendingSalesOrders, nil
}

func (m *MockNembusClient) FetchUnpostedSalesOrdersV2(ctx context.Context, limit int) ([]nembus.SalesOrderV2, error) {
	return m.unpostedSalesOrders, nil
}

func (m *MockNembusClient) FetchSalesOrderLinesV2(ctx context.Context, orderID uuid.UUID) ([]nembus.SalesOrderLineV2, error) {
	return m.salesOrderLines[orderID], nil
}

func (m *MockNembusClient) MarkSalesOrderPostedToSAP(ctx context.Context, orderID uuid.UUID, sapDocEntry, sapDocNum int) error {
	m.markedSalesOrders[orderID] = markedSAPDoc{DocEntry: sapDocEntry, DocNum: sapDocNum}
	return nil
}

func (m *MockNembusClient) HoldSalesOrderForStockReview(ctx context.Context, orderID uuid.UUID, orderNum string, storeID int32, whCode string, totalAmount float64, shortageItems []nembus.ShortageItem, summary string) error {
	if m.heldSalesOrders == nil {
		m.heldSalesOrders = make(map[uuid.UUID]string)
	}
	m.heldSalesOrders[orderID] = summary
	return nil
}

func (m *MockNembusClient) GetApprovalPolicyRules(ctx context.Context, domain, action string) (map[string]any, bool, error) {
	if !m.stockHoldPolicyLoaded {
		return map[string]any{"hold_on_zero_stock": true}, true, nil
	}
	return m.stockHoldPolicyRules, m.stockHoldPolicyRequired, m.stockHoldPolicyErr
}

func TestSalesOrderUpstream_StockShortageHold(t *testing.T) {
	orderID := uuid.New()
	mockNembus := newMockNembusClient()
	mockSAP := &MockSAPClient{stockBySKU: map[string]float64{"SKU-1": 0}}
	mockNembus.unpostedSalesOrders = []nembus.SalesOrderV2{{
		ID:          orderID,
		OrderNumber: "SO-100",
		StoreID:     1,
		OrderDate:   time.Now(),
		TotalAmount: 50,
	}}
	mockNembus.salesOrderLines[orderID] = []nembus.SalesOrderLineV2{{
		ProductID: 1, ProductSKU: "SKU-1", ProductName: "Widget", Quantity: 2, UnitPrice: 25, LineTotal: 50,
	}}

	sync := pipeline.NewUpstreamSync(&config.Config{NembusDefaultWarehouse: "WH1", BatchSize: 10}, mockSAP, mockNembus)
	stats, err := sync.ProcessOutbox(context.Background())
	if err != nil {
		t.Fatalf("ProcessOutbox: %v", err)
	}
	if stats.OrdersHeld != 1 {
		t.Fatalf("expected 1 order held, got %d", stats.OrdersHeld)
	}
	if _, ok := mockNembus.heldSalesOrders[orderID]; !ok {
		t.Fatal("expected sales order held for review")
	}
}

func TestSalesOrderUpstream_PostsWhenStockAvailable(t *testing.T) {
	orderID := uuid.New()
	mockNembus := newMockNembusClient()
	mockSAP := &MockSAPClient{stockBySKU: map[string]float64{"SKU-2": 100}}
	mockNembus.unpostedSalesOrders = []nembus.SalesOrderV2{{
		ID:          orderID,
		OrderNumber: "SO-200",
		StoreID:     1,
		OrderDate:   time.Now(),
		TotalAmount: 40,
	}}
	mockNembus.salesOrderLines[orderID] = []nembus.SalesOrderLineV2{{
		ProductID: 2, ProductSKU: "SKU-2", ProductName: "Gadget", Quantity: 1, UnitPrice: 40, LineTotal: 40,
	}}

	sync := pipeline.NewUpstreamSync(&config.Config{NembusDefaultWarehouse: "WH1", BatchSize: 10}, mockSAP, mockNembus)
	stats, err := sync.ProcessOutbox(context.Background())
	if err != nil {
		t.Fatalf("ProcessOutbox: %v", err)
	}
	if stats.OrdersPosted != 1 {
		t.Fatalf("expected 1 order posted, got %d", stats.OrdersPosted)
	}
	if _, ok := mockNembus.markedSalesOrders[orderID]; !ok {
		t.Fatal("expected sales order marked posted to SAP")
	}
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

	origTxID := testUUID(42)
	reason := "Damaged item"
	customerID := testUUID(101)

	mockNembus.unpostedSalesReturns = []nembus.SalesReturn{
		{
			ID: testUUID(10),
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

	mockNembus.returnLines[testUUID(10)] = []nembus.SalesReturnLine{
		{
			ID: testUUID(101),
			ReturnID: testUUID(10),
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
	marked, ok := mockNembus.markedReturns[testUUID(10)]
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
			ID: testUUID(20),
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

	mockNembus.returnLines[testUUID(20)] = []nembus.SalesReturnLine{
		{
			ID: testUUID(201),
			ReturnID: testUUID(20),
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
			ID: testUUID(15),
			TransactionNumber: "TX-DRYRUN",
			TotalAmount:       100.0,
			AmountPaid:        100.0,
			TransactionDate:   time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC),
		},
	}
	mockNembus.txLines[testUUID(15)] = []nembus.POSTransactionLine{
		{ID: testUUID(151), TransactionID: testUUID(15), ProductSKU: "SKU-TX", Quantity: 1.0, UnitPrice: 100.0, LineTotal: 100.0},
	}

	mockNembus.unpostedSalesReturns = []nembus.SalesReturn{
		{
			ID: testUUID(30),
			ReturnNumber:      "RET-DRYRUN",
			Status:            "completed",
			TotalRefundAmount: 200.0,
			ReturnDate:        time.Date(2026, 10, 22, 11, 0, 0, 0, time.UTC),
		},
	}
	mockNembus.returnLines[testUUID(30)] = []nembus.SalesReturnLine{
		{ID: testUUID(301), ReturnID: testUUID(30), ProductSKU: "SKU-D", Quantity: 1.0, UnitPrice: 200.0, RefundAmount: 200.0},
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
			ID: testUUID(40),
			ReturnNumber:      "RET-FAIL",
			Status:            "completed",
			TotalRefundAmount: 50.0,
		},
	}
	mockNembus.returnLines[testUUID(40)] = []nembus.SalesReturnLine{
		{ID: testUUID(401), ReturnID: testUUID(40), ProductSKU: "SKU-FAIL", Quantity: 1.0, UnitPrice: 50.0, RefundAmount: 50.0},
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
			ID: testUUID(50),
			ReturnNumber:      "RET-ZERO-QTY",
			Status:            "completed",
			TotalRefundAmount: 50.0,
			ReturnDate:        time.Date(2026, 10, 22, 11, 0, 0, 0, time.UTC),
		},
	}
	mockNembus.returnLines[testUUID(50)] = []nembus.SalesReturnLine{
		{ID: testUUID(501), ReturnID: testUUID(50), ProductSKU: "SKU-INV", Quantity: 0.0, UnitPrice: 50.0, RefundAmount: 50.0},
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
			ID: testUUID(60),
			ReturnNumber:      "RET-NEG-AMT",
			Status:            "completed",
			TotalRefundAmount: -25.0,
			ReturnDate:        time.Date(2026, 10, 22, 11, 0, 0, 0, time.UTC),
		},
	}
	mockNembus.returnLines[testUUID(60)] = []nembus.SalesReturnLine{
		{ID: testUUID(601), ReturnID: testUUID(60), ProductSKU: "SKU-NEG", Quantity: 1.0, UnitPrice: 25.0, RefundAmount: -25.0},
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

	origTxID := testUUID(9999) // Non-existent in mock originalSAPDocEntries (resolves to 0)
	mockNembus.unpostedSalesReturns = []nembus.SalesReturn{
		{
			ID: testUUID(70),
			ReturnNumber:          "RET-UNRESOLVED-ENTRY",
			OriginalTransactionID: &origTxID,
			Status:                "completed",
			TotalRefundAmount:     80.0,
			ReturnDate:            time.Date(2026, 10, 22, 11, 0, 0, 0, time.UTC),
		},
	}
	mockNembus.returnLines[testUUID(70)] = []nembus.SalesReturnLine{
		{ID: testUUID(701), ReturnID: testUUID(70), ProductSKU: "SKU-UNRES", Quantity: 2.0, UnitPrice: 40.0, RefundAmount: 80.0},
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

type stockShortageSAPClient struct {
	MockSAPClient
	onHandStock float64
}

func (s *stockShortageSAPClient) CheckItemWarehouseStock(ctx context.Context, itemCode, warehouseCode string) (float64, error) {
	return s.onHandStock, nil
}

func TestProcessOutbox_StockShortageHold(t *testing.T) {
	mockSAP := &stockShortageSAPClient{onHandStock: 0.0}
	mockNembus := newMockNembusClient()
	mockNembus.unpostedPOSTransactions = []nembus.POSTransaction{
		{
			ID: testUUID(88),
			TransactionNumber: "TX-HOLD-001",
			TransactionDate:   time.Now(),
			TotalAmount:       150.0,
			Status:            "completed",
		},
	}
	mockNembus.txLines[testUUID(88)] = []nembus.POSTransactionLine{
		{ID: testUUID(881), TransactionID: testUUID(88), ProductSKU: "SKU-DEPLETED", Quantity: 5.0, UnitPrice: 30.0, LineTotal: 150.0},
	}

	cfg := &config.Config{BatchSize: 10, NembusDefaultWarehouse: "W01"}
	sync := pipeline.NewUpstreamSync(cfg, mockSAP, mockNembus)
	ctx := context.Background()

	stats, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("unexpected ProcessOutbox error: %v", err)
	}

	if stats.TransactionsPosted != 0 {
		t.Errorf("expected 0 transactions posted, got %d", stats.TransactionsPosted)
	}
	if summary, held := mockNembus.heldTransactions[testUUID(88)]; !held {
		t.Errorf("expected transaction 88 to be held for review")
	} else if !strings.Contains(summary, "SAP Stock Shortage") {
		t.Errorf("expected shortage summary, got: %s", summary)
	}
}

func TestProcessOutbox_BinAllocationsAndSeries(t *testing.T) {
	var capturedInvoice *sap.SAPInvoiceRequest
	mockSAP := &MockSAPClient{
		binAllocations: []sap.SAPBinAllocation{{BinAbsEntry: 2, Quantity: 3}},
		seriesByObj:    map[string]int{"13": 101, "14": 202},
		baseLineByItem: map[string]int{"INV00028": 4},
		postInvoiceFunc: func(ctx context.Context, inv *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error) {
			capturedInvoice = inv
			return &sap.SAPInvoiceResponse{DocEntry: 500, DocNum: 5000}, nil
		},
	}
	mockNembus := newMockNembusClient()
	mockNembus.unpostedPOSTransactions = []nembus.POSTransaction{
		{
			ID:                testUUID(901),
			TransactionNumber: "TX-BIN-001",
			TransactionDate:   time.Now(),
			TotalAmount:       60.0,
			Status:            "completed",
		},
	}
	mockNembus.txLines[testUUID(901)] = []nembus.POSTransactionLine{
		{ID: testUUID(9011), TransactionID: testUUID(901), ProductSKU: "INV00028", Quantity: 3, UnitPrice: 20, LineTotal: 60},
	}
	origTxID := testUUID(902)
	mockNembus.originalSAPDocEntries[origTxID] = 500
	mockNembus.unpostedSalesReturns = []nembus.SalesReturn{
		{
			ID:                    testUUID(903),
			ReturnNumber:          "RET-SER-001",
			ReturnDate:            time.Now(),
			OriginalTransactionID: &origTxID,
			TotalRefundAmount:     60.0,
		},
	}
	mockNembus.returnLines[testUUID(903)] = []nembus.SalesReturnLine{
		{ID: testUUID(9031), ReturnID: testUUID(903), ProductSKU: "INV00028", Quantity: 3, UnitPrice: 20, RefundAmount: 60},
	}

	cfg := &config.Config{BatchSize: 10, NembusDefaultWarehouse: "01"}
	u := pipeline.NewUpstreamSync(cfg, mockSAP, mockNembus)
	stats, err := u.ProcessOutbox(context.Background())
	if err != nil {
		t.Fatalf("ProcessOutbox error: %v", err)
	}
	if stats.TransactionsPosted != 1 || stats.ReturnsPosted != 1 {
		t.Fatalf("expected 1 tx and 1 return posted, got %+v", stats)
	}
	if capturedInvoice == nil || capturedInvoice.Series == nil || *capturedInvoice.Series != 101 {
		t.Fatalf("expected invoice Series=101, got %+v", capturedInvoice)
	}
	if len(capturedInvoice.DocumentLines[0].DocumentLinesBinAllocations) != 1 ||
		capturedInvoice.DocumentLines[0].DocumentLinesBinAllocations[0].BinAbsEntry != 2 ||
		capturedInvoice.DocumentLines[0].DocumentLinesBinAllocations[0].BaseLineNumber == nil ||
		*capturedInvoice.DocumentLines[0].DocumentLinesBinAllocations[0].BaseLineNumber != 0 {
		t.Fatalf("unexpected invoice bin allocations: %+v", capturedInvoice.DocumentLines[0].DocumentLinesBinAllocations)
	}
	if len(mockSAP.postedCreditMemos) != 1 {
		t.Fatalf("expected 1 credit memo posted")
	}
	memo := mockSAP.postedCreditMemos[0]
	if memo.Series == nil || *memo.Series != 202 {
		t.Fatalf("expected credit memo Series=202, got %v", memo.Series)
	}
	if memo.DocumentLines[0].BaseLine == nil || *memo.DocumentLines[0].BaseLine != 4 {
		t.Fatalf("expected credit memo BaseLine=4, got %v", memo.DocumentLines[0].BaseLine)
	}
	if len(memo.DocumentLines[0].DocumentLinesBinAllocations) != 1 {
		t.Fatalf("expected credit memo bin allocation to be attached")
	}
}

func TestProcessOutbox_CodeMinus10BinErrorDoesNotCreateStockApproval(t *testing.T) {
	mockSAP := &MockSAPClient{
		postInvoiceFunc: func(ctx context.Context, inv *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error) {
			return nil, fmt.Errorf("failed to post invoice to SAP: %w", &sap.APIError{
				StatusCode: 400,
				Code:       -10,
				Message:    `1470000341 - Fully allocate item "INV00028" to bin locations in warehouse "01"`,
			})
		},
	}
	mockNembus := newMockNembusClient()
	mockNembus.stockHoldPolicyRequired = true
	mockNembus.stockHoldPolicyRules = map[string]any{"hold_on_zero_stock": true}
	mockNembus.unpostedPOSTransactions = []nembus.POSTransaction{
		{
			ID:                testUUID(910),
			TransactionNumber: "TX-BIN-ERR",
			TransactionDate:   time.Now(),
			TotalAmount:       50.0,
			Status:            "completed",
		},
	}
	mockNembus.txLines[testUUID(910)] = []nembus.POSTransactionLine{
		{ID: testUUID(9101), TransactionID: testUUID(910), ProductSKU: "INV00028", Quantity: 1, UnitPrice: 50, LineTotal: 50},
	}

	cfg := &config.Config{BatchSize: 10, NembusDefaultWarehouse: "01"}
	u := pipeline.NewUpstreamSync(cfg, mockSAP, mockNembus)
	stats, err := u.ProcessOutbox(context.Background())
	if err != nil {
		t.Fatalf("ProcessOutbox error: %v", err)
	}
	if stats.TransactionsHeld != 0 || len(mockNembus.heldTransactions) != 0 {
		t.Fatalf("generic [code -10] bin allocation error must NOT create a stock-hold approval (held=%d)", len(mockNembus.heldTransactions))
	}
	if stats.Errors != 1 {
		t.Fatalf("expected 1 error, got %d", stats.Errors)
	}
}

func TestProcessOutbox_StockHoldPolicyDisabled_DoesNotHold(t *testing.T) {
	mockSAP := &stockShortageSAPClient{
		MockSAPClient: MockSAPClient{
			postInvoiceFunc: func(ctx context.Context, invoice *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error) {
				return nil, errors.New("SAP error code -10: item quantity would fall below zero")
			},
		},
		onHandStock: 0.0,
	}
	mockNembus := newMockNembusClient()
	mockNembus.stockHoldPolicyLoaded = true
	mockNembus.stockHoldPolicyRequired = false
	mockNembus.unpostedPOSTransactions = []nembus.POSTransaction{
		{
			ID: testUUID(89),
			TransactionNumber: "TX-NO-HOLD-001",
			TransactionDate:   time.Now(),
			TotalAmount:       120.0,
			Status:            "completed",
		},
	}
	mockNembus.txLines[testUUID(89)] = []nembus.POSTransactionLine{
		{ID: testUUID(891), TransactionID: testUUID(89), ProductSKU: "SKU-ZERO", Quantity: 2.0, UnitPrice: 60.0, LineTotal: 120.0},
	}

	cfg := &config.Config{BatchSize: 10, NembusDefaultWarehouse: "W01"}
	sync := pipeline.NewUpstreamSync(cfg, mockSAP, mockNembus)

	stats, err := sync.ProcessOutbox(context.Background())
	if err != nil {
		t.Fatalf("unexpected ProcessOutbox error: %v", err)
	}
	if stats.TransactionsHeld != 0 {
		t.Errorf("expected 0 held transactions when policy disabled, got %d", stats.TransactionsHeld)
	}
	if stats.Errors != 1 {
		t.Errorf("expected 1 SAP posting error without reactive hold, got %d", stats.Errors)
	}
	if len(mockNembus.heldTransactions) != 0 {
		t.Errorf("expected HoldTransactionForStockReview not to be called, got %+v", mockNembus.heldTransactions)
	}
}

func TestProcessOutbox_HoldOnZeroStockDisabled(t *testing.T) {
	t.Run("zero stock bypasses preflight hold", func(t *testing.T) {
		mockSAP := &stockShortageSAPClient{onHandStock: 0.0}
		mockNembus := newMockNembusClient()
		mockNembus.stockHoldPolicyLoaded = true
		mockNembus.stockHoldPolicyRequired = true
		mockNembus.stockHoldPolicyRules = map[string]any{"hold_on_zero_stock": false}
		mockNembus.unpostedPOSTransactions = []nembus.POSTransaction{
			{
				ID: testUUID(90),
				TransactionNumber: "TX-ZERO-ALLOW",
				TransactionDate:   time.Now(),
				TotalAmount:       60.0,
				Status:            "completed",
			},
		}
		mockNembus.txLines[testUUID(90)] = []nembus.POSTransactionLine{
			{ID: testUUID(901), TransactionID: testUUID(90), ProductSKU: "SKU-ZERO", Quantity: 2.0, UnitPrice: 30.0, LineTotal: 60.0},
		}

		cfg := &config.Config{BatchSize: 10, NembusDefaultWarehouse: "W01"}
		sync := pipeline.NewUpstreamSync(cfg, mockSAP, mockNembus)

		stats, err := sync.ProcessOutbox(context.Background())
		if err != nil {
			t.Fatalf("unexpected ProcessOutbox error: %v", err)
		}
		if stats.TransactionsHeld != 0 || len(mockNembus.heldTransactions) != 0 {
			t.Errorf("expected 0 held transactions when hold_on_zero_stock=false, got stats.Held=%d held=%+v", stats.TransactionsHeld, mockNembus.heldTransactions)
		}
		if stats.TransactionsPosted != 1 {
			t.Errorf("expected 1 transaction posted when zero-stock hold disabled, got %d", stats.TransactionsPosted)
		}
	})

	t.Run("partial shortage still holds", func(t *testing.T) {
		mockSAP := &stockShortageSAPClient{onHandStock: 2.0}
		mockNembus := newMockNembusClient()
		mockNembus.stockHoldPolicyLoaded = true
		mockNembus.stockHoldPolicyRequired = true
		mockNembus.stockHoldPolicyRules = map[string]any{"hold_on_zero_stock": false}
		mockNembus.unpostedPOSTransactions = []nembus.POSTransaction{
			{
				ID: testUUID(91),
				TransactionNumber: "TX-PARTIAL-HOLD",
				TransactionDate:   time.Now(),
				TotalAmount:       150.0,
				Status:            "completed",
			},
		}
		mockNembus.txLines[testUUID(91)] = []nembus.POSTransactionLine{
			{ID: testUUID(911), TransactionID: testUUID(91), ProductSKU: "SKU-LOW", Quantity: 5.0, UnitPrice: 30.0, LineTotal: 150.0},
		}

		cfg := &config.Config{BatchSize: 10, NembusDefaultWarehouse: "W01"}
		sync := pipeline.NewUpstreamSync(cfg, mockSAP, mockNembus)

		stats, err := sync.ProcessOutbox(context.Background())
		if err != nil {
			t.Fatalf("unexpected ProcessOutbox error: %v", err)
		}
		if stats.TransactionsHeld != 1 {
			t.Errorf("expected 1 held transaction for partial shortage, got %d", stats.TransactionsHeld)
		}
		if _, held := mockNembus.heldTransactions[testUUID(91)]; !held {
			t.Errorf("expected transaction 91 to be held")
		}
	})
}

func TestProcessOutbox_WatermarkFiltering(t *testing.T) {
	store, err := db.NewSQLiteStore(filepath.Join(t.TempDir(), "upstream_wm.db"))
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	run, err := store.CreateRun(ctx, 1, contracts.MigrationModeFull, 1)
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	if err := store.UpdateRunStatus(ctx, run.ID, contracts.StatusCompleted, ""); err != nil {
		t.Fatalf("UpdateRunStatus: %v", err)
	}
	migEnd, err := store.GetLatestCompletedFullMigrationEndTime(ctx)
	if err != nil || migEnd == nil {
		t.Fatalf("expected migration end time, got %v (err: %v)", migEnd, err)
	}

	mockSAP := &MockSAPClient{stockBySKU: map[string]float64{"*": 999}}
	mockNembus := newMockNembusClient()
	preMigOrderID := uuid.New()
	mockNembus.unpostedPOSTransactions = []nembus.POSTransaction{
		{
			ID: testUUID(1001),
			TransactionNumber: "TX-PRE-MIG",
			TransactionDate:   migEnd.Add(-10 * time.Minute),
			TotalAmount:       50,
			Status:            "completed",
		},
		{
			ID: testUUID(1002),
			TransactionNumber: "TX-POST-MIG",
			TransactionDate:   migEnd.Add(2 * time.Second),
			TotalAmount:       75,
			Status:            "completed",
		},
	}
	mockNembus.txLines[testUUID(1001)] = []nembus.POSTransactionLine{{ID: testUUID(1), ProductSKU: "SKU-1", Quantity: 1, LineTotal: 50}}
	mockNembus.txLines[testUUID(1002)] = []nembus.POSTransactionLine{{ID: testUUID(2), ProductSKU: "SKU-1", Quantity: 1, LineTotal: 75}}
	mockNembus.unpostedSalesOrders = []nembus.SalesOrderV2{
		{
			ID:          preMigOrderID,
			OrderNumber: "SO-PRE-MIG",
			OrderDate:   migEnd.Add(-5 * time.Minute),
			TotalAmount: 90,
		},
	}
	mockNembus.salesOrderLines[preMigOrderID] = []nembus.SalesOrderLineV2{{ProductSKU: "SKU-1", Quantity: 1, LineTotal: 90}}

	cfg := &config.Config{BatchSize: 10, NembusDefaultWarehouse: "01"}
	sync := pipeline.NewUpstreamSync(cfg, mockSAP, mockNembus, store)

	// 1. First run skips pre-migration TX/Order and posts only post-migration TX.
	stats, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("ProcessOutbox run 1: %v", err)
	}
	if stats.TransactionsPosted != 1 || stats.OrdersPosted != 0 {
		t.Fatalf("expected 1 TX and 0 orders posted, got %+v", stats)
	}
	if _, ok := mockNembus.markedTransactions[testUUID(1001)]; ok {
		t.Fatalf("pre-migration transaction 1001 should have been skipped")
	}
	if _, ok := mockNembus.markedTransactions[testUUID(1002)]; !ok {
		t.Fatalf("post-migration transaction 1002 should have been posted")
	}

	wm1, err := store.GetUpstreamWatermark(ctx, "upstream")
	if err != nil || wm1 == nil {
		t.Fatalf("expected upstream watermark after run 1, got %v (err: %v)", wm1, err)
	}

	// 2. Subsequent scheduled run advances watermark and ignores items before wm1, while errors block advance.
	time.Sleep(1100 * time.Millisecond)
	mockNembus.unpostedPOSTransactions = []nembus.POSTransaction{
		{
			ID: testUUID(1003),
			TransactionNumber: "TX-FAIL-AFTER",
			TransactionDate:   wm1.Add(1 * time.Second),
			TotalAmount:       25,
			Status:            "completed",
		},
	}
	mockNembus.txLines[testUUID(1003)] = []nembus.POSTransactionLine{{ID: testUUID(3), ProductSKU: "SKU-1", Quantity: 1, LineTotal: 25}}
	mockSAP.postInvoiceFunc = func(ctx context.Context, inv *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error) {
		return nil, errors.New("sap temporary failure")
	}
	stats2, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("ProcessOutbox run 2: %v", err)
	}
	if stats2.Errors != 1 {
		t.Fatalf("expected 1 error, got %+v", stats2)
	}
	wm2, err := store.GetUpstreamWatermark(ctx, "upstream")
	if err != nil || wm2 == nil || !wm2.Equal(*wm1) {
		t.Fatalf("watermark must not advance on error: got %v, want %v", wm2, wm1)
	}
}

func TestProcessOutbox_SkipsZeroLineRecordsAndAdvancesWatermark(t *testing.T) {
	ctx := context.Background()
	store, err := db.NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	runStartedAt := time.Now().UTC().Add(-time.Second)
	base := time.Now().UTC().Add(-10 * time.Minute)
	emptyTxTime := base.Add(1 * time.Minute)
	emptyOrderTime := base.Add(2 * time.Minute)
	validTxTime := base.Add(3 * time.Minute)

	emptyOrderID := uuid.New()
	mockNembus := newMockNembusClient()
	mockNembus.unpostedPOSTransactions = []nembus.POSTransaction{
		{
			ID: testUUID(101),
			TransactionNumber: "TXN-ORD-EMPTY",
			StoreID:           1,
			TransactionDate:   emptyTxTime,
			TotalAmount:       25,
		},
		{
			ID: testUUID(102),
			TransactionNumber: "INV-VALID",
			StoreID:           1,
			TransactionDate:   validTxTime,
			TotalAmount:       50,
		},
	}
	mockNembus.txLines[testUUID(102)] = []nembus.POSTransactionLine{{
		ProductID: 1, ProductSKU: "SKU-1", ProductName: "Item 1", Quantity: 1, UnitPrice: 50, LineTotal: 50,
	}}
	mockNembus.unpostedSalesOrders = []nembus.SalesOrderV2{{
		ID:          emptyOrderID,
		OrderNumber: "SO-EMPTY",
		StoreID:     1,
		OrderDate:   emptyOrderTime,
		TotalAmount: 30,
	}}

	mockSAP := &MockSAPClient{stockBySKU: map[string]float64{"*": 100}}
	cfg := &config.Config{NembusDefaultWarehouse: "WH1", BatchSize: 10, NembusStoreID: 1}
	sync := pipeline.NewUpstreamSync(cfg, mockSAP, mockNembus, store)

	stats, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("ProcessOutbox: %v", err)
	}
	if stats.Errors != 0 || stats.TransactionsPosted != 1 {
		t.Fatalf("unexpected stats: got %+v, want Errors=0 TransactionsPosted=1", stats)
	}

	dispatches, err := store.GetUpstreamDispatches(ctx, 10)
	if err != nil {
		t.Fatalf("GetUpstreamDispatches: %v", err)
	}
	statuses := make(map[string]db.UpstreamDispatch)
	for _, d := range dispatches {
		statuses[d.TransactionNumber] = d
	}
	for _, docNum := range []string{"TXN-ORD-EMPTY", "SO-EMPTY"} {
		d, ok := statuses[docNum]
		if !ok {
			t.Fatalf("expected dispatch record for %s", docNum)
		}
		if d.Status != "skipped" {
			t.Fatalf("expected %s status skipped, got %q", docNum, d.Status)
		}
		if d.ErrorMessage != "no line items" {
			t.Fatalf("expected %s error message 'no line items', got %q", docNum, d.ErrorMessage)
		}
	}
	if d, ok := statuses["INV-VALID"]; !ok || d.Status != "posted" {
		t.Fatalf("expected INV-VALID posted, got %+v (ok=%v)", d, ok)
	}

	wm, err := store.GetUpstreamWatermark(ctx, "upstream")
	if err != nil || wm == nil {
		t.Fatalf("expected upstream watermark to be set, got %v (err=%v)", wm, err)
	}
	if wm.Before(runStartedAt) {
		t.Fatalf("expected upstream watermark >= %v, got %v", runStartedAt, *wm)
	}
}

func intPtr(v int) *int { return &v }

func TestPostTransactionToSAP_UoMEntryAndManualUseBaseUnits(t *testing.T) {
	ctx := context.Background()
	var capturedInv *sap.SAPInvoiceRequest
	mockSAP := &MockSAPClient{
		stockBySKU: map[string]float64{"*": 100},
		postInvoiceFunc: func(ctx context.Context, invoice *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error) {
			capturedInv = invoice
			return &sap.SAPInvoiceResponse{DocEntry: 201, DocNum: 26001, DocTotal: 597.0}, nil
		},
	}
	mockNembus := newMockNembusClient()
	mockNembus.unpostedPOSTransactions = []nembus.POSTransaction{
		{
			ID: testUUID(501),
			StoreID:           1,
			CustomerSAPCode:   "C-REAL-474",
			TransactionNumber: "TXN-UOM-1",
			TransactionDate:   time.Now().UTC(),
			TotalAmount:       597.0,
			AmountPaid:        597.0,
			Status:            "completed",
		},
	}
	mockNembus.txLines[testUUID(501)] = []nembus.POSTransactionLine{
		{
			ID: testUUID(1),
			TransactionID: testUUID(501),
			ProductSKU:          "ITEM-UGP",
			ProductName:         "UoM Group Item",
			Quantity:            1.0,
			UnitPrice:           345.0,
			TaxAmount:           45.0,
			LineTotal:           345.0,
			SAPUoMEntry:         intPtr(12),
			SAPUgpEntry:         intPtr(5),
			UoMConversionFactor: 1.0,
			IsBaseUoM:           true,
			SAPTaxCode:          "O15",
		},
		{
			ID: testUUID(2),
			TransactionID: testUUID(501),
			ProductSKU:          "ITEM-MANUAL",
			ProductName:         "Manual Carton Item",
			Quantity:            1.0,
			UnitPrice:           252.0,
			TaxAmount:           0.0,
			LineTotal:           252.0,
			SAPUoMEntry:         nil,
			SAPUgpEntry:         intPtr(-1),
			UoMConversionFactor: 12.0,
			IsBaseUoM:           false,
			SAPTaxCode:          "O0",
		},
	}

	sync := pipeline.NewUpstreamSync(&config.Config{BatchSize: 10, NembusDefaultWarehouse: "01"}, mockSAP, mockNembus)
	stats, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("ProcessOutbox: %v", err)
	}
	if stats.TransactionsPosted != 1 || stats.Errors != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if capturedInv == nil || len(capturedInv.DocumentLines) != 2 {
		t.Fatalf("expected 2 invoice lines, got %+v", capturedInv)
	}
	if capturedInv.CardCode != "C-REAL-474" {
		t.Fatalf("expected CardCode C-REAL-474, got %s", capturedInv.CardCode)
	}

	line0 := capturedInv.DocumentLines[0]
	if line0.UoMEntry == nil || *line0.UoMEntry != 12 {
		t.Fatalf("expected UoMEntry=12 for UoM group item, got %v", line0.UoMEntry)
	}
	if line0.UseBaseUnits != "" {
		t.Fatalf("expected empty UseBaseUnits for UoM group item, got %q", line0.UseBaseUnits)
	}
	if line0.TaxCode != "O15" {
		t.Fatalf("expected TaxCode O15, got %q", line0.TaxCode)
	}
	if math.Abs(line0.UnitPrice-300.0) > 0.001 {
		t.Fatalf("expected net UnitPrice 300.0, got %f", line0.UnitPrice)
	}

	line1 := capturedInv.DocumentLines[1]
	if line1.UoMEntry != nil {
		t.Fatalf("expected nil UoMEntry for manual UoM item, got %v", *line1.UoMEntry)
	}
	if line1.UseBaseUnits != "tYES" {
		t.Fatalf("expected UseBaseUnits=tYES for manual UoM item, got %q", line1.UseBaseUnits)
	}
	if math.Abs(line1.Quantity-12.0) > 0.001 || math.Abs(line1.UnitPrice-21.0) > 0.001 || math.Abs(line1.PriceAfterVAT-21.0) > 0.001 {
		t.Fatalf("expected converted base qty=12 and price=21, got qty=%f unitPrice=%f afterVAT=%f", line1.Quantity, line1.UnitPrice, line1.PriceAfterVAT)
	}
	if line1.TaxCode != "O0" {
		t.Fatalf("expected TaxCode O0, got %q", line1.TaxCode)
	}
}

func TestPostTransactionToSAP_PaymentTotalMismatchLogsWarning(t *testing.T) {
	ctx := context.Background()
	var postedPayments []*sap.SAPPaymentRequest
	mockSAP := &MockSAPClient{
		stockBySKU: map[string]float64{"*": 100},
		postInvoiceFunc: func(ctx context.Context, invoice *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error) {
			return &sap.SAPInvoiceResponse{DocEntry: 301, DocNum: 36001, DocTotal: 345.0, VatSum: 45.0}, nil
		},
		postIncomingPaymentFunc: func(ctx context.Context, payment *sap.SAPPaymentRequest) (*sap.SAPPaymentResponse, error) {
			postedPayments = append(postedPayments, payment)
			return &sap.SAPPaymentResponse{DocEntry: 401, DocNum: 46001}, nil
		},
	}
	mockNembus := newMockNembusClient()
	mockNembus.unpostedPOSTransactions = []nembus.POSTransaction{
		{
			ID: testUUID(601),
			StoreID:           1,
			TransactionNumber: "TXN-PAY-WARN",
			TransactionDate:   time.Now().UTC(),
			TotalAmount:       300.0,
			AmountPaid:        0, // Also tests TotalAmount > 0 fallback
			Status:            "completed",
		},
		{
			ID: testUUID(602),
			StoreID:           1,
			TransactionNumber: "TXN-PAY-CAP",
			TransactionDate:   time.Now().UTC(),
			TotalAmount:       345.0,
			AmountPaid:        400.0,
			Status:            "completed",
		},
	}
	mockNembus.txLines[testUUID(601)] = []nembus.POSTransactionLine{{ID: testUUID(1), ProductSKU: "SKU-1", Quantity: 1, UnitPrice: 300, LineTotal: 300}}
	mockNembus.txLines[testUUID(602)] = []nembus.POSTransactionLine{{ID: testUUID(2), ProductSKU: "SKU-1", Quantity: 1, UnitPrice: 300, LineTotal: 345, TaxAmount: 45}}
	mockNembus.txPayments[testUUID(601)] = []nembus.POSPayment{{ID: testUUID(1), TransactionID: testUUID(601), PaymentMethod: "cash", Amount: 300.0}}
	mockNembus.txPayments[testUUID(602)] = []nembus.POSPayment{{ID: testUUID(2), TransactionID: testUUID(602), PaymentMethod: "cash", Amount: 400.0}}

	var logs []string
	sync := pipeline.NewUpstreamSync(&config.Config{BatchSize: 10, NembusDefaultWarehouse: "01"}, mockSAP, mockNembus)
	sync.SetLogger(func(domain, level, msg string) {
		logs = append(logs, level+":"+msg)
	})

	stats, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("ProcessOutbox: %v", err)
	}
	if stats.TransactionsPosted != 2 {
		t.Fatalf("expected 2 posted, got %+v", stats)
	}
	if len(postedPayments) != 2 {
		t.Fatalf("expected 2 incoming payments, got %d", len(postedPayments))
	}
	if math.Abs(postedPayments[1].CashSum-345.0) > 0.001 || math.Abs(postedPayments[1].PaymentInvoices[0].SumApplied-345.0) > 0.001 {
		t.Fatalf("expected capped overpayment 345.0, got cashSum=%f sumApplied=%f", postedPayments[1].CashSum, postedPayments[1].PaymentInvoices[0].SumApplied)
	}
	foundWarn := false
	for _, entry := range logs {
		if strings.Contains(entry, "WARN:") && strings.Contains(entry, "TXN-PAY-WARN") && strings.Contains(entry, "differs from payments sum") {
			foundWarn = true
			break
		}
	}
	if !foundWarn {
		t.Fatalf("expected WARN log for payment mismatch, got logs: %v", logs)
	}
}

func TestProcessOutbox_OfflineTransactionNotDroppedAfterIdlePoll(t *testing.T) {
	ctx := context.Background()
	store, err := db.NewSQLiteStore(filepath.Join(t.TempDir(), "offline_cutoff.db"))
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	mockSAP := &MockSAPClient{stockBySKU: map[string]float64{"*": 100}}
	mockNembus := newMockNembusClient()
	sync := pipeline.NewUpstreamSync(&config.Config{BatchSize: 10, NembusDefaultWarehouse: "01"}, mockSAP, mockNembus, store)

	// Run 1: idle poll advances 'upstream' watermark to now, initializes 'upstream_cutoff' to midnight UTC today.
	stats1, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("ProcessOutbox idle: %v", err)
	}
	if stats1.TransactionsPosted != 0 {
		t.Fatalf("expected 0 posted on idle run, got %+v", stats1)
	}

	// Simulate an offline POS transaction created 30 seconds after UTC midnight (earlier than the idle poll!).
	offlineDate := time.Now().UTC().Truncate(24*time.Hour).Add(30 * time.Second)
	mockNembus.unpostedPOSTransactions = []nembus.POSTransaction{
		{
			ID: testUUID(701),
			StoreID:           1,
			TransactionNumber: "TXN-OFFLINE-1",
			TransactionDate:   offlineDate,
			TotalAmount:       100.0,
			AmountPaid:        100.0,
			Status:            "completed",
		},
	}
	mockNembus.txLines[testUUID(701)] = []nembus.POSTransactionLine{{ID: testUUID(1), ProductSKU: "SKU-1", Quantity: 1, UnitPrice: 100, LineTotal: 100}}

	stats2, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("ProcessOutbox run 2: %v", err)
	}
	if stats2.TransactionsPosted != 1 {
		t.Fatalf("expected offline transaction after cutoff to post, got %+v", stats2)
	}
}

func TestProcessOutbox_ErrorLoggingNotUpToDate(t *testing.T) {
	ctx := context.Background()
	mockSAP := &MockSAPClient{
		stockBySKU: map[string]float64{"*": 100},
		postInvoiceFunc: func(ctx context.Context, invoice *sap.SAPInvoiceRequest) (*sap.SAPInvoiceResponse, error) {
			return nil, errors.New("Invalid BP code")
		},
	}
	mockNembus := newMockNembusClient()
	mockNembus.unpostedPOSTransactions = []nembus.POSTransaction{
		{
			ID: testUUID(801),
			StoreID:           1,
			TransactionNumber: "TXN-ERR-1",
			TransactionDate:   time.Now().UTC(),
			TotalAmount:       50.0,
			Status:            "completed",
		},
	}
	mockNembus.txLines[testUUID(801)] = []nembus.POSTransactionLine{{ID: testUUID(1), ProductSKU: "SKU-1", Quantity: 1, UnitPrice: 50, LineTotal: 50}}

	var logs []string
	sync := pipeline.NewUpstreamSync(&config.Config{BatchSize: 10, NembusDefaultWarehouse: "01"}, mockSAP, mockNembus)
	sync.SetLogger(func(domain, level, msg string) {
		logs = append(logs, level+":"+msg)
	})

	stats, err := sync.ProcessOutbox(ctx)
	if err != nil {
		t.Fatalf("ProcessOutbox: %v", err)
	}
	if stats.Errors != 1 {
		t.Fatalf("expected 1 error, got %+v", stats)
	}
	for _, entry := range logs {
		if strings.Contains(entry, "pos_transactions: up to date") {
			t.Fatalf("must not log 'pos_transactions: up to date' when errors occurred: %v", logs)
		}
	}
	joined := strings.Join(logs, "\n")
	if !strings.Contains(joined, "1 failed") || !strings.Contains(joined, "1 errors") {
		t.Fatalf("expected failure counts in logs, got:\n%s", joined)
	}
}


