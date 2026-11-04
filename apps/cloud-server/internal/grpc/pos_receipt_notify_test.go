package grpc

import "testing"

func TestPosReceiptDecision(t *testing.T) {
	id, number, _, _, customer, ok := posReceiptDecision([]byte(`{
		"id": 9,
		"transaction_number": "POS-9",
		"status": "completed",
		"store_id": 3,
		"customer_id": 4,
		"voided_at": null
	}`))
	if !ok || id != 9 || number != "POS-9" || customer == nil || *customer != 4 {
		t.Fatalf("completed sale not enqueued: id=%d number=%s ok=%v customer=%v", id, number, ok, customer)
	}

	if _, _, _, _, _, ok := posReceiptDecision([]byte(`{"id":9,"transaction_number":"POS-9","status":"completed","voided_at":"2026-11-03T00:00:00Z"}`)); ok {
		t.Fatal("voided sale was enqueued")
	}
	if _, _, _, _, _, ok := posReceiptDecision([]byte(`{"id":9,"transaction_number":"POS-9","status":"voided"}`)); ok {
		t.Fatal("void status was enqueued")
	}
	if _, _, _, _, _, ok := posReceiptDecision([]byte(`{"id":9,"status":"completed"}`)); ok {
		t.Fatal("sale without a receipt number was enqueued")
	}
}
