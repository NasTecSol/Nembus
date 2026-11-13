package nembus

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
