package nembus_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/internal/nembus"
	"github.com/google/uuid"
)

func testEntityUUID(legacy int64) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("nembus-test-entity-%d", legacy)))
}

func TestSalesReturnStructSerialization(t *testing.T) {
	origTxID := testEntityUUID(50)
	cashierID := int32(1)
	customerID := testEntityUUID(10)
	reason := "Customer return"

	ret := nembus.SalesReturn{
		ID:                    testEntityUUID(100),
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

	if unmarshaled.ID != testEntityUUID(100) {
		t.Errorf("expected ID %s, got %s", testEntityUUID(100), unmarshaled.ID)
	}
	if unmarshaled.ReturnNumber != "RET-100" {
		t.Errorf("expected ReturnNumber RET-100, got %s", unmarshaled.ReturnNumber)
	}
	if unmarshaled.OriginalTransactionID == nil || *unmarshaled.OriginalTransactionID != testEntityUUID(50) {
		t.Errorf("expected OriginalTransactionID %s, got %v", testEntityUUID(50), unmarshaled.OriginalTransactionID)
	}
	if unmarshaled.TotalRefundAmount != 287.5 {
		t.Errorf("expected TotalRefundAmount 287.5, got %f", unmarshaled.TotalRefundAmount)
	}
}

func TestSalesReturnLineSerialization(t *testing.T) {
	origLineID := testEntityUUID(501)

	line := nembus.SalesReturnLine{
		ID:             testEntityUUID(1),
		ReturnID:       testEntityUUID(100),
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