package usecase

import (
	"testing"
)

func TestAddressTypeValidation(t *testing.T) {
	tests := []struct {
		name        string
		addressType string
		wantValid   bool
	}{
		{name: "legacy bill_to", addressType: "bill_to", wantValid: true},
		{name: "legacy ship_to", addressType: "ship_to", wantValid: true},
		{name: "legacy both", addressType: "both", wantValid: true},
		{name: "sap/normalized billing", addressType: "billing", wantValid: true},
		{name: "sap/normalized shipping", addressType: "shipping", wantValid: true},
		{name: "invalid empty", addressType: "", wantValid: false},
		{name: "invalid random", addressType: "office", wantValid: false},
		{name: "invalid case sensitive", addressType: "BILLING", wantValid: false},
		{name: "invalid whitespace", addressType: "billing ", wantValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidAddressType(tt.addressType)
			if got != tt.wantValid {
				t.Errorf("isValidAddressType(%q) = %v; want %v", tt.addressType, got, tt.wantValid)
			}
		})
	}
}
