package usecase

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
)

type mockSAPInvoicePoster struct {
	postCalled bool
	lastTxID   int64
	lastForce  bool
}

func (m *mockSAPInvoicePoster) PostHeldInvoiceToSAP(ctx context.Context, transactionID int64, force bool) (int, int, error) {
	m.postCalled = true
	m.lastTxID = transactionID
	m.lastForce = force
	return 101, 5001, nil
}

func TestCatalogPriceAccommodator_Rejection(t *testing.T) {
	acc := NewCatalogPriceAccommodator()
	if acc == nil {
		t.Fatal("expected non-nil CatalogPriceAccommodator")
	}

	payload := CatalogPriceChangePayload{
		ProductID:   1,
		ProductSKU:  "SKU-TEST",
		PriceListID: 10,
	}
	rawPayload, _ := json.Marshal(payload)
	req := repository.ApprovalRequest{
		ID:      1,
		Payload: rawPayload,
	}

	// Rejection should return nil without mutating anything
	if err := acc.Reject(context.Background(), nil, req, "Rejected: Price too low"); err != nil {
		t.Fatalf("unexpected error on reject: %v", err)
	}
}

func TestCatalogPriceAccommodator_Validation(t *testing.T) {
	acc := NewCatalogPriceAccommodator()

	// 1. Nil repo
	err := acc.Accommodate(context.Background(), nil, repository.ApprovalRequest{}, 1)
	if err == nil || err.Error() != "repository is not configured" {
		t.Fatalf("expected 'repository is not configured', got %v", err)
	}

	// 2. Invalid JSON payload
	// Since repo is checked first, create a mock-like or check unmarshal
	// Next, valid payload checks
	newPrice := 99.99
	payload := CatalogPriceChangePayload{
		ProductID:   0, // Missing product ID
		PriceListID: 1,
		NewPrice:    &newPrice,
	}
	rawPayload, _ := json.Marshal(payload)
	req := repository.ApprovalRequest{Payload: rawPayload}

	var dummyRepo repository.Queries
	err = acc.Accommodate(context.Background(), &dummyRepo, req, 1)
	if err == nil || err.Error() != "product_id is required in payload" {
		t.Fatalf("expected 'product_id is required in payload', got %v", err)
	}
}

func TestSAPInvoiceStockAccommodator_Poster(t *testing.T) {
	mockPoster := &mockSAPInvoicePoster{}
	acc := NewSAPInvoiceStockAccommodator(mockPoster)
	if acc == nil {
		t.Fatal("expected non-nil SAPInvoiceStockAccommodator")
	}

	payload := SAPInvoiceStockPayload{
		TransactionID:     42,
		TransactionNumber: "TX-42",
		ShortageItems: []ShortageItemInfo{
			{ProductID: 10, ProductSKU: "SKU-1", Quantity: 2, SAPOnHand: 0},
		},
	}
	rawPayload, _ := json.Marshal(payload)
	req := repository.ApprovalRequest{
		ID:      77,
		Payload: rawPayload,
	}

	// Accommodate with nil repo
	err := acc.Accommodate(context.Background(), nil, req, 5)
	if err == nil || err.Error() != "repository is not configured" {
		t.Fatalf("expected 'repository is not configured', got %v", err)
	}

	// Reject with nil repo
	err = acc.Reject(context.Background(), nil, req, "Rejected: create adjustment")
	if err == nil || err.Error() != "repository is not configured" {
		t.Fatalf("expected 'repository is not configured', got %v", err)
	}

	// Accommodate blocks approval when SAPOnHand < Quantity
	var dummyRepo repository.Queries
	err = acc.Accommodate(context.Background(), &dummyRepo, req, 5)
	if err == nil || !strings.Contains(err.Error(), "still has low stock in SAP") {
		t.Fatalf("expected low stock error, got %v", err)
	}
}

func TestApprovalsUseCase_AccommodatorRegistration(t *testing.T) {
	uc := NewApprovalsUseCase()
	catAcc := NewCatalogPriceAccommodator()
	sapAcc := NewSAPInvoiceStockAccommodator()

	uc.RegisterAccommodator("catalog", catAcc)
	uc.RegisterAccommodator("sap_invoice", sapAcc)

	if len(uc.accommodators) != 2 {
		t.Fatalf("expected 2 registered accommodators, got %d", len(uc.accommodators))
	}
}

func TestUpdateStockHoldPayloadOnHand(t *testing.T) {
	raw, _ := json.Marshal(SAPInvoiceStockPayload{
		TransactionID: 653872,
		StoreID:       3,
		ShortageItems: []ShortageItemInfo{
			{ProductID: 27512, ProductSKU: "INV10473", Quantity: 1, SAPOnHand: 0},
		},
	})

	// Unchanged on-hand returns false
	if _, changed := updateStockHoldPayloadOnHand(raw, 27512, "INV10473", 0); changed {
		t.Fatal("expected changed=false when sap_on_hand is already 0")
	}

	// Restocked to 5 updates sap_on_hand
	updated, changed := updateStockHoldPayloadOnHand(raw, 27512, "INV10473", 5)
	if !changed {
		t.Fatal("expected changed=true when sap_on_hand increases to 5")
	}
	var out SAPInvoiceStockPayload
	if err := json.Unmarshal(updated, &out); err != nil {
		t.Fatalf("Unmarshal updated payload: %v", err)
	}
	if out.ShortageItems[0].SAPOnHand != 5 {
		t.Fatalf("expected sap_on_hand=5, got %v", out.ShortageItems[0].SAPOnHand)
	}
}
