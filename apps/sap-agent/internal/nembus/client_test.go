package nembus_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/internal/nembus"
)

func TestSalesReturnStructSerialization(t *testing.T) {
	origTxID := int32(50)
	cashierID := int32(1)
	customerID := int32(10)
	reason := "Customer return"

	ret := nembus.SalesReturn{
		ID:                    100,
		ReturnNumber:          "RET-100",
		StoreID:               2,
		CashierID:             &cashierID,
		CustomerID:            &customerID,
		OriginalTransactionID: &origTxID,
		ReturnDate:            time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC),
		ReturnReason:          &reason,
		Status:                "completed",
		Subtotal:              250.0,
		TaxAmount:             37.5,
		TotalRefundAmount:     287.5,
		Metadata:              json.RawMessage(`{"source":"pos"}`),
	}

	data, err := json.Marshal(ret)
	if err != nil {
		t.Fatalf("failed to marshal SalesReturn: %v", err)
	}

	var unmarshaled nembus.SalesReturn
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal SalesReturn: %v", err)
	}

	if unmarshaled.ID != 100 {
		t.Errorf("expected ID 100, got %d", unmarshaled.ID)
	}
	if unmarshaled.ReturnNumber != "RET-100" {
		t.Errorf("expected ReturnNumber RET-100, got %s", unmarshaled.ReturnNumber)
	}
	if unmarshaled.OriginalTransactionID == nil || *unmarshaled.OriginalTransactionID != 50 {
		t.Errorf("expected OriginalTransactionID 50, got %v", unmarshaled.OriginalTransactionID)
	}
	if unmarshaled.TotalRefundAmount != 287.5 {
		t.Errorf("expected TotalRefundAmount 287.5, got %f", unmarshaled.TotalRefundAmount)
	}
}

func TestSalesReturnLineSerialization(t *testing.T) {
	origLineID := int32(501)

	line := nembus.SalesReturnLine{
		ID:             1,
		ReturnID:       100,
		ProductID:      25,
		ProductSKU:     "SKU-RET-TEST",
		ProductName:    "Test Item",
		OriginalLineID: &origLineID,
		Quantity:       3.0,
		UnitPrice:      50.0,
		RefundAmount:   172.5,
		ReturnToStock:  true,
		LineNumber:     1,
	}

	data, err := json.Marshal(line)
	if err != nil {
		t.Fatalf("failed to marshal SalesReturnLine: %v", err)
	}

	var unmarshaled nembus.SalesReturnLine
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal SalesReturnLine: %v", err)
	}

	if unmarshaled.ProductSKU != "SKU-RET-TEST" {
		t.Errorf("expected SKU-RET-TEST, got %s", unmarshaled.ProductSKU)
	}
	if unmarshaled.Quantity != 3.0 {
		t.Errorf("expected quantity 3.0, got %f", unmarshaled.Quantity)
	}
	if !unmarshaled.ReturnToStock {
		t.Errorf("expected ReturnToStock true")
	}
}
