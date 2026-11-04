package grpc

import "testing"

func TestPosReceiptDecision(t *testing.T) {
	id, number, _, _, customer, phone, ok := posReceiptDecision([]byte(`{
		"id": 9,
		"transaction_number": "POS-9",
		"status": "completed",
		"store_id": 3,
		"customer_id": 4,
		"metadata": {"customer_phone": "+923216061056"},
		"voided_at": null
	}`))
	if !ok || id != 9 || number != "POS-9" || customer == nil || *customer != 4 || phone != "+923216061056" {
		t.Fatalf("completed sale: id=%d number=%s ok=%v customer=%v phone=%s", id, number, ok, customer, phone)
	}

	if _, _, _, _, _, _, ok := posReceiptDecision([]byte(`{"id":9,"transaction_number":"POS-9","status":"completed","voided_at":"2026-11-03T00:00:00Z"}`)); ok {
		t.Fatal("voided sale was enqueued")
	}
	if _, _, _, _, _, _, ok := posReceiptDecision([]byte(`{"id":9,"transaction_number":"POS-9","status":"voided"}`)); ok {
		t.Fatal("void status was enqueued")
	}
	if _, _, _, _, _, _, ok := posReceiptDecision([]byte(`{"id":9,"status":"completed"}`)); ok {
		t.Fatal("sale without a receipt number was enqueued")
	}
}

func TestResolveReceiptPhone(t *testing.T) {
	if got := resolveReceiptPhone("+1", "+2", "+3", "+4"); got != "+1" {
		t.Fatalf("metadata wins: %s", got)
	}
	if got := resolveReceiptPhone("", "+2", "+3", "+4"); got != "+2" {
		t.Fatalf("order wins: %s", got)
	}
	if got := resolveReceiptPhone("", "", "+3", "+4"); got != "+3" {
		t.Fatalf("cart guest wins: %s", got)
	}
	if got := resolveReceiptPhone("", "", "", "+4"); got != "+4" {
		t.Fatalf("customer wins: %s", got)
	}
	if got := resolveReceiptPhone("", "", "", ""); got != "" {
		t.Fatalf("expected empty, got %s", got)
	}
}
