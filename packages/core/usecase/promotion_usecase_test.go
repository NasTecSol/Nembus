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
