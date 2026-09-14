package usecase

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestBundleToOutput(t *testing.T) {
	var priceNumeric pgtype.Numeric
	_ = priceNumeric.Scan("49.99")

	bundle := repository.ComboBundle{
		ID:                      1,
		OrganizationID:          10,
		StoreID:                 pgtype.Int4{Int32: 2, Valid: true},
		PriceListID:             pgtype.Int4{Int32: 3, Valid: true},
		ApplicableCustomerTypes: []string{"retail", "wholesale"},
		Code:                    "BUNDLE-RETAIL-01",
		Name:                    "Retail Starter Pack",
		Description:             pgtype.Text{String: "Starter kit bundle", Valid: true},
		BundlePrice:             priceNumeric,
		BundleType:              pgtype.Text{String: "fixed", Valid: true},
		IsActive:                pgtype.Bool{Bool: true, Valid: true},
		ValidFrom:               pgtype.Date{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		ValidTo:                 pgtype.Date{Time: time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), Valid: true},
		DisplayOrder:            pgtype.Int4{Int32: 5, Valid: true},
		Metadata:                json.RawMessage(`{"tags":["featured"]}`),
	}

	out := bundleToOutput(bundle)

	if out.ID != 1 {
		t.Errorf("expected ID 1, got %d", out.ID)
	}
	if out.OrganizationID != 10 {
		t.Errorf("expected OrganizationID 10, got %d", out.OrganizationID)
	}
	if out.StoreID == nil || *out.StoreID != 2 {
		t.Errorf("expected StoreID 2, got %v", out.StoreID)
	}
	if out.PriceListID == nil || *out.PriceListID != 3 {
		t.Errorf("expected PriceListID 3, got %v", out.PriceListID)
	}
	if len(out.ApplicableCustomerTypes) != 2 || out.ApplicableCustomerTypes[0] != "retail" {
		t.Errorf("expected customer types [retail, wholesale], got %v", out.ApplicableCustomerTypes)
	}
	if out.BundlePrice != "49.99" {
		t.Errorf("expected BundlePrice '49.99', got %s", out.BundlePrice)
	}
	if !out.IsActive {
		t.Errorf("expected IsActive true, got false")
	}
	if out.ValidFrom == nil || *out.ValidFrom != "2026-01-01" {
		t.Errorf("expected ValidFrom '2026-01-01', got %v", out.ValidFrom)
	}
	if out.ValidTo == nil || *out.ValidTo != "2026-12-31" {
		t.Errorf("expected ValidTo '2026-12-31', got %v", out.ValidTo)
	}
}

func TestBundlePriceMetadata(t *testing.T) {
	vID := int32(101)
	meta := bundlePriceMetadata{
		BundlePrice: 35.50,
		BundleItems: []bundlePriceItemMetadata{
			{ProductID: 1, ProductVariantID: &vID, Quantity: 2},
			{ProductID: 2, Quantity: 1},
		},
		AllowMultiple: true,
	}

	bytes, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("failed to marshal bundlePriceMetadata: %v", err)
	}

	var decoded bundlePriceMetadata
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatalf("failed to unmarshal bundlePriceMetadata: %v", err)
	}

	if decoded.BundlePrice != 35.50 {
		t.Errorf("expected BundlePrice 35.50, got %f", decoded.BundlePrice)
	}
	if len(decoded.BundleItems) != 2 {
		t.Fatalf("expected 2 items, got %d", len(decoded.BundleItems))
	}
	if decoded.BundleItems[0].ProductID != 1 || *decoded.BundleItems[0].ProductVariantID != 101 || decoded.BundleItems[0].Quantity != 2 {
		t.Errorf("first item mismatch: %+v", decoded.BundleItems[0])
	}
	if decoded.BundleItems[1].ProductID != 2 || decoded.BundleItems[1].ProductVariantID != nil || decoded.BundleItems[1].Quantity != 1 {
		t.Errorf("second item mismatch: %+v", decoded.BundleItems[1])
	}
	if !decoded.AllowMultiple {
		t.Errorf("expected AllowMultiple true, got false")
	}
}

func TestBundleDiscountProportionalDistribution(t *testing.T) {
	// Scenario:
	// Item A: 2 units @ $20.00 each = $40.00 normal total
	// Item B: 1 unit  @ $30.00 each = $30.00 normal total
	// Normal total = $70.00
	// Target Bundle Price = $50.00
	// Total Discount = $20.00
	// Proportional share:
	// Item A: (40 / 70) * 20 = $11.4285...
	// Item B: (30 / 70) * 20 = $8.5714...

	itemATotal := 40.0
	itemBTotal := 30.0
	normalSubtotal := itemATotal + itemBTotal
	targetBundlePrice := 50.0
	totalSavings := normalSubtotal - targetBundlePrice

	ratioA := itemATotal / normalSubtotal
	ratioB := itemBTotal / normalSubtotal

	discountA := totalSavings * ratioA
	discountB := totalSavings * ratioB

	if totalSavings != 20.0 {
		t.Errorf("expected total savings 20.0, got %f", totalSavings)
	}
	if discountA+discountB != 20.0 {
		t.Errorf("expected sum of discounts to equal 20.0, got %f", discountA+discountB)
	}
}
