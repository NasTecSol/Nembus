package nembus

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	legacyconfig "github.com/NasTecSol/nembus-sap-agent/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/google/uuid"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func testJWT(orgID int, userID string) string {
	hdr := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claimsBytes, _ := json.Marshal(map[string]any{
		"organization_id": orgID,
		"user_id":         userID,
	})
	claims := base64.RawURLEncoding.EncodeToString(claimsBytes)
	return hdr + "." + claims + ".sig"
}

func TestHoldForStockReview_SubmitsRequiredFieldsAndChecksStatus(t *testing.T) {
	ctx := context.Background()
	store, err := db.NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	var submitStatus int = http.StatusCreated
	var capturedBody map[string]any
	token := testJWT(7, "42")

	client := NewHTTPClient(&legacyconfig.Config{
		NembusAPIURL:         "https://nembus.example.test",
		NembusOrganizationID: 7,
	}, store)
	client.httpClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			switch {
			case r.Method == http.MethodPost && r.URL.Path == "/api/auth/login":
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"token":"` + token + `"}`)),
					Header:     make(http.Header),
				}, nil
			case r.Method == http.MethodPost && r.URL.Path == "/api/approvals/submit":
				raw, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(raw, &capturedBody)
				body := `{"id":99,"status":"pending"}`
				if submitStatus >= 400 {
					body = `{"error":"Key: 'ApprovalInput.OrganizationID' Error"}`
				}
				return &http.Response{
					StatusCode: submitStatus,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     make(http.Header),
				}, nil
			default:
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader(`{}`)),
					Header:     make(http.Header),
				}, nil
			}
		}),
	}

	shortages := []ShortageItem{{
		ProductID:   10,
		ProductSKU:  "SKU-10",
		ProductName: "Widget",
		Quantity:    5,
		SAPOnHand:   1,
	}}
	if err := client.HoldTransactionForStockReview(ctx, 123, "INV-123", 3, "01", 99.5, shortages, "shortage"); err != nil {
		t.Fatalf("HoldTransactionForStockReview: %v", err)
	}

	if got := int(capturedBody["organization_id"].(float64)); got != 7 {
		t.Fatalf("expected organization_id=7, got %v", capturedBody["organization_id"])
	}
	if got := int(capturedBody["requested_by"].(float64)); got != 42 {
		t.Fatalf("expected requested_by=42, got %v", capturedBody["requested_by"])
	}
	if capturedBody["domain"] != "sap_invoice" || capturedBody["action"] != "stock_hold" {
		t.Fatalf("unexpected domain/action: %v / %v", capturedBody["domain"], capturedBody["action"])
	}
	if capturedBody["entity_type"] != "pos_transaction" || capturedBody["entity_id"] != "123" {
		t.Fatalf("unexpected entity: %v / %v", capturedBody["entity_type"], capturedBody["entity_id"])
	}

	// Non-2xx response must return an error and NOT record held dispatch on sales orders.
	submitStatus = http.StatusBadRequest
	orderID := uuid.New()
	if err := client.HoldSalesOrderForStockReview(ctx, orderID, "SO-FAIL", 3, "01", 50, shortages, "shortage"); err == nil || !strings.Contains(err.Error(), "status 400") {
		t.Fatalf("expected status 400 error, got %v", err)
	}
	dispatches, err := store.GetUpstreamDispatches(ctx, 10)
	if err != nil {
		t.Fatalf("GetUpstreamDispatches: %v", err)
	}
	if len(dispatches) != 0 {
		t.Fatalf("expected 0 dispatches after failed approval submit, got %+v", dispatches)
	}

	// 201 response succeeds and records held dispatch for sales orders.
	submitStatus = http.StatusCreated
	if err := client.HoldSalesOrderForStockReview(ctx, orderID, "SO-OK", 3, "01", 50, shortages, "shortage"); err != nil {
		t.Fatalf("HoldSalesOrderForStockReview 201: %v", err)
	}
	dispatches, err = store.GetUpstreamDispatches(ctx, 10)
	if err != nil {
		t.Fatalf("GetUpstreamDispatches: %v", err)
	}
	if len(dispatches) != 1 || dispatches[0].TransactionNumber != "SO-OK" || dispatches[0].Status != "held" {
		t.Fatalf("expected SO-OK held dispatch, got %+v", dispatches)
	}
}

func TestFetchApprovedPendingPOSTransactions_TopLevelArrayResponse(t *testing.T) {
	ctx := context.Background()
	store, err := db.NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	if err := store.RecordUpstreamDispatch(ctx, "TXN-HELD-1", time.Now(), 50, nil, nil, "held", "shortage"); err != nil {
		t.Fatalf("RecordUpstreamDispatch: %v", err)
	}

	token := testJWT(1, "1")
	client := NewHTTPClient(&legacyconfig.Config{
		NembusAPIURL:         "https://nembus.example.test",
		NembusOrganizationID: 1,
		NembusStoreID:        3,
	}, store)
	client.httpClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			switch {
			case r.Method == http.MethodPost && r.URL.Path == "/api/auth/login":
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"token":"` + token + `"}`)),
					Header:     make(http.Header),
				}, nil
			case r.Method == http.MethodGet && r.URL.Path == "/api/pos/stores/3/transactions":
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"data":{"data":[{"id":55,"transaction_number":"TXN-HELD-1","store_id":3,"total_amount":50,"amount_paid":50}]}}`)),
					Header:     make(http.Header),
				}, nil
			case r.Method == http.MethodGet && r.URL.Path == "/api/approvals":
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`[{"entity_type":"pos_transaction","entity_id":"55","status":"approved","payload":{"transaction_number":"TXN-HELD-1"}}]`)),
					Header:     make(http.Header),
				}, nil
			default:
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader(`{}`)),
					Header:     make(http.Header),
				}, nil
			}
		}),
	}

	got, err := client.FetchApprovedPendingPOSTransactions(ctx, 50)
	if err != nil {
		t.Fatalf("FetchApprovedPendingPOSTransactions: %v", err)
	}
	if len(got) != 1 || got[0].TransactionNumber != "TXN-HELD-1" {
		t.Fatalf("expected 1 approved pending tx TXN-HELD-1, got %+v", got)
	}
}

func TestHTTPClient_UnpostedTransactionsLinesAndSapSync(t *testing.T) {
	ctx := context.Background()
	store, err := db.NewSQLiteStore(filepath.Join(t.TempDir(), "http_test.db"))
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	// Seed 505 dispatches so TXN-OLD-1 falls beyond the old LIMIT 500 window.
	for i := 1; i <= 505; i++ {
		txNum := fmt.Sprintf("TXN-OLD-%d", i)
		if err := store.RecordUpstreamDispatch(ctx, txNum, time.Now(), 10, nil, nil, "posted", ""); err != nil {
			t.Fatalf("RecordUpstreamDispatch %d: %v", i, err)
		}
	}

	token := testJWT(1, "1")
	var syncPayload map[string]int
	var syncPath string

	client := NewHTTPClient(&legacyconfig.Config{
		NembusAPIURL:         "https://nembus.example.test",
		NembusOrganizationID: 1,
		NembusStoreID:        2,
	}, store)
	client.httpClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			switch {
			case r.Method == http.MethodPost && r.URL.Path == "/api/auth/login":
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"token":"` + token + `"}`)), Header: make(http.Header)}, nil
			case r.Method == http.MethodGet && r.URL.Path == "/api/pos/stores/2/transactions/unposted":
				body := `{"data":[
					{"id":1,"store_id":2,"cashier_id":3,"cashier_session_id":7,"customer_id":474,"customer_sap_code":"C474","transaction_number":"TXN-OLD-1","transaction_date":"2026-11-19T10:00:00Z","total_amount":"100.00","amount_paid":"100.00","status":"completed"},
					{"id":2,"store_id":2,"cashier_id":3,"cashier_session_id":7,"customer_id":474,"customer_sap_code":"C474","transaction_number":"TXN-NEW-2","transaction_date":"2026-11-19T11:00:00Z","total_amount":"345.00","amount_paid":"0.00","status":"completed","metadata":{"source":"pos"}}
				]}`
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			case r.Method == http.MethodGet && r.URL.Path == "/api/pos/transactions/2/full":
				body := `{"data":[
					{"line_id":11,"product_id":99,"uom_id":5,"sku":"INV00683","product_name":"Espresso","quantity":"1.000","unit_price":"300.0000","discount_amount_2":"0.00","line_tax_amount":"45.00","line_subtotal":"300.00","line_total":"345.00","sap_uom_entry":"2","sap_ugp_entry":"-1","uom_conversion_factor":"2","is_base_uom":false,"sap_tax_code":"O15"}
				]}`
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			case r.Method == http.MethodPost && r.URL.Path == "/api/pos/transactions/2/sap-sync":
				syncPath = r.URL.Path
				raw, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(raw, &syncPayload)
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"data":{"id":2}}`)), Header: make(http.Header)}, nil
			default:
				return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(`{}`)), Header: make(http.Header)}, nil
			}
		}),
	}

	txs, err := client.FetchUnpostedPOSTransactions(ctx, 50)
	if err != nil {
		t.Fatalf("FetchUnpostedPOSTransactions: %v", err)
	}
	if len(txs) != 1 || txs[0].TransactionNumber != "TXN-NEW-2" {
		t.Fatalf("expected only TXN-NEW-2 (TXN-OLD-1 skipped beyond 500 cap), got %+v", txs)
	}
	if txs[0].CustomerID == nil || *txs[0].CustomerID != 474 || txs[0].CustomerSAPCode != "C474" || txs[0].CashierSessionID != 7 {
		t.Fatalf("unexpected parsed customer/session fields: %+v", txs[0])
	}

	lines, err := client.FetchTransactionLines(ctx, 2)
	if err != nil {
		t.Fatalf("FetchTransactionLines: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	l := lines[0]
	if l.SAPUoMEntry == nil || *l.SAPUoMEntry != 2 || l.SAPUgpEntry == nil || *l.SAPUgpEntry != -1 || l.UoMConversionFactor != 2 || l.IsBaseUoM || l.TaxAmount != 45 || l.SAPTaxCode != "O15" {
		t.Fatalf("unexpected parsed line metadata: %+v", l)
	}

	if err := client.MarkTransactionPostedToSAP(ctx, 2, 20025, 2600018); err != nil {
		t.Fatalf("MarkTransactionPostedToSAP: %v", err)
	}
	if syncPath != "/api/pos/transactions/2/sap-sync" || syncPayload["sap_doc_entry"] != 20025 || syncPayload["sap_doc_num"] != 2600018 {
		t.Fatalf("unexpected sap-sync call path=%s payload=%+v", syncPath, syncPayload)
	}
}

func TestHTTPClient_SalesOrdersFallbackSkipsPOSAndPropagates500(t *testing.T) {
	ctx := context.Background()
	store, err := db.NewSQLiteStore(filepath.Join(t.TempDir(), "orders_test.db"))
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	token := testJWT(1, "1")
	ordersStatus := http.StatusOK
	posOrderID := uuid.New().String()
	webOrderID := uuid.New().String()

	client := NewHTTPClient(&legacyconfig.Config{
		NembusAPIURL:         "https://nembus.example.test",
		NembusOrganizationID: 1,
	}, store)
	client.httpClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			switch {
			case r.Method == http.MethodPost && r.URL.Path == "/api/auth/login":
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"token":"` + token + `"}`)), Header: make(http.Header)}, nil
			case r.Method == http.MethodGet && r.URL.Path == "/api/orders/unposted-sap":
				return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(`{}`)), Header: make(http.Header)}, nil
			case r.Method == http.MethodGet && r.URL.Path == "/api/orders":
				if ordersStatus != http.StatusOK {
					return &http.Response{StatusCode: ordersStatus, Body: io.NopCloser(strings.NewReader(`{"error":"db fail"}`)), Header: make(http.Header)}, nil
				}
				body := fmt.Sprintf(`{"data":[
					{"id":%q,"order_number":"ORD-POS-1","sales_channel":"pos","order_status":"confirmed","fulfillment_status":"fulfilled","total_amount":"100.00"},
					{"id":%q,"order_number":"ORD-WEB-1","sales_channel":"online","order_status":"confirmed","fulfillment_status":"fulfilled","total_amount":"250.00"}
				]}`, posOrderID, webOrderID)
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			default:
				return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(`{}`)), Header: make(http.Header)}, nil
			}
		}),
	}

	orders, err := client.FetchUnpostedSalesOrdersV2(ctx, 50)
	if err != nil {
		t.Fatalf("FetchUnpostedSalesOrdersV2 fallback: %v", err)
	}
	if len(orders) != 1 || orders[0].OrderNumber != "ORD-WEB-1" {
		t.Fatalf("expected only ORD-WEB-1 (POS channel skipped), got %+v", orders)
	}

	ordersStatus = http.StatusInternalServerError
	if _, err := client.FetchUnpostedSalesOrdersV2(ctx, 50); err == nil {
		t.Fatalf("expected HTTP 500 from /api/orders fallback to return an error")
	}
}
