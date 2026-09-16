package usecase

import (
	"testing"
)

func TestValidateBucketComboMetadata(t *testing.T) {
	tests := []struct {
		name    string
		meta    BucketComboMetadata
		wantErr bool
	}{
		{
			name: "Valid fixed_price bucket_combo",
			meta: BucketComboMetadata{
				ComboType:  "fixed_price",
				ComboPrice: floatPtr(50.0),
				BucketItems: []BucketItem{
					{ProductID: 10, Quantity: 2},
					{ProductID: 20, Quantity: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid percentage bucket_combo",
			meta: BucketComboMetadata{
				ComboType:     "percentage",
				DiscountValue: floatPtr(20.0),
				BucketItems: []BucketItem{
					{ProductID: 10, Quantity: 1},
					{ProductID: 20, Quantity: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid fixed_discount bucket_combo",
			meta: BucketComboMetadata{
				ComboType:     "fixed_discount",
				DiscountValue: floatPtr(15.0),
				BucketItems: []BucketItem{
					{ProductID: 10, Quantity: 1},
					{ProductID: 20, Quantity: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid - Less than 2 distinct products",
			meta: BucketComboMetadata{
				ComboType:  "fixed_price",
				ComboPrice: floatPtr(50.0),
				BucketItems: []BucketItem{
					{ProductID: 10, Quantity: 2},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid - Duplicate products",
			meta: BucketComboMetadata{
				ComboType:  "fixed_price",
				ComboPrice: floatPtr(50.0),
				BucketItems: []BucketItem{
					{ProductID: 10, Quantity: 2},
					{ProductID: 10, Quantity: 1},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid - Quantity less than 1",
			meta: BucketComboMetadata{
				ComboType:  "fixed_price",
				ComboPrice: floatPtr(50.0),
				BucketItems: []BucketItem{
					{ProductID: 10, Quantity: 0},
					{ProductID: 20, Quantity: 1},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid - Unknown combo_type",
			meta: BucketComboMetadata{
				ComboType: "invalid_type",
				BucketItems: []BucketItem{
					{ProductID: 10, Quantity: 1},
					{ProductID: 20, Quantity: 1},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBucketComboMetadata(tt.meta)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBucketComboMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func TestCalculateTierAndThreshold(t *testing.T) {
	tests := []struct {
		points            float64
		expectedTier      string
		expectedThreshold float64
	}{
		{points: 0, expectedTier: "bronze", expectedThreshold: 500.0},
		{points: 250, expectedTier: "bronze", expectedThreshold: 500.0},
		{points: 499.99, expectedTier: "bronze", expectedThreshold: 500.0},
		{points: 500, expectedTier: "silver", expectedThreshold: 1000.0},
		{points: 750, expectedTier: "silver", expectedThreshold: 1000.0},
		{points: 999.99, expectedTier: "silver", expectedThreshold: 1000.0},
		{points: 1000, expectedTier: "gold", expectedThreshold: 1000.0},
		{points: 2500, expectedTier: "gold", expectedThreshold: 1000.0},
	}

	for _, tt := range tests {
		tier, threshold := CalculateTierAndThreshold(tt.points)
		if tier != tt.expectedTier {
			t.Errorf("CalculateTierAndThreshold(%.2f) tier = %s, want %s", tt.points, tier, tt.expectedTier)
		}
		if threshold != tt.expectedThreshold {
			t.Errorf("CalculateTierAndThreshold(%.2f) threshold = %.2f, want %.2f", tt.points, threshold, tt.expectedThreshold)
		}
	}
}

func TestIsTierAllowed(t *testing.T) {
	tests := []struct {
		name         string
		targetTiers  []string
		customerTier string
		wantAllowed  bool
	}{
		{
			name:         "Empty target tiers allows any customer tier",
			targetTiers:  []string{},
			customerTier: "bronze",
			wantAllowed:  true,
		},
		{
			name:         "Gold target tier allows Gold customer",
			targetTiers:  []string{"gold"},
			customerTier: "gold",
			wantAllowed:  true,
		},
		{
			name:         "Gold target tier rejects Silver customer",
			targetTiers:  []string{"gold"},
			customerTier: "silver",
			wantAllowed:  false,
		},
		{
			name:         "Multiple target tiers match (case-insensitive)",
			targetTiers:  []string{"Silver", "Gold"},
			customerTier: "silver",
			wantAllowed:  true,
		},
		{
			name:         "Multiple target tiers reject Bronze customer",
			targetTiers:  []string{"silver", "gold"},
			customerTier: "bronze",
			wantAllowed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTierAllowed(tt.targetTiers, tt.customerTier)
			if got != tt.wantAllowed {
				t.Errorf("isTierAllowed(%v, %s) = %v, want %v", tt.targetTiers, tt.customerTier, got, tt.wantAllowed)
			}
		})
	}
}

