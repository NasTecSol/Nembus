package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// posReceiptDecision reports whether a synced pos_transactions payload should
// enqueue a customer receipt. Lines and the ZATCA QR are loaded when the
// notification is sent.
func posReceiptDecision(payload []byte) (id string, number string, storeID int32, orgID int32, customerID *string, metaPhone string, enqueue bool) {
	if len(payload) == 0 {
		return "", "", 0, 0, nil, "", false
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		return "", "", 0, 0, nil, "", false
	}
	id = jsonEntityIDString(raw["id"])
	number = jsonString(raw["transaction_number"])
	storeID = int32(jsonInt64(raw["store_id"]))
	orgID = int32(jsonInt64(raw["organization_id"]))
	metaPhone = metadataStringField(raw["metadata"], "customer_phone")
	if cid := jsonEntityIDString(raw["customer_id"]); cid != "" {
		customerID = &cid
	}
	status := strings.ToLower(strings.TrimSpace(jsonString(raw["status"])))
	if status != "" && status != "completed" {
		return id, number, storeID, orgID, customerID, metaPhone, false
	}
	if jsonPresent(raw["voided_at"]) || jsonPresent(raw["voided_by"]) {
		return id, number, storeID, orgID, customerID, metaPhone, false
	}
	if id == "" || number == "" {
		return id, number, storeID, orgID, customerID, metaPhone, false
	}
	return id, number, storeID, orgID, customerID, metaPhone, true
}

func jsonEntityIDString(raw json.RawMessage) string {
	if s := jsonString(raw); s != "" {
		return s
	}
	if n := jsonInt64(raw); n > 0 {
		return strconv.FormatInt(n, 10)
	}
	return ""
}

func jsonString(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	return ""
}

func jsonInt64(raw json.RawMessage) int64 {
	if len(raw) == 0 || string(raw) == "null" {
		return 0
	}
	var n int64
	if err := json.Unmarshal(raw, &n); err == nil {
		return n
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return int64(f)
	}
	return 0
}

func jsonPresent(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	text := strings.TrimSpace(string(raw))
	return text != "" && text != "null" && text != `""`
}

func metadataStringField(raw json.RawMessage, key string) string {
	if len(raw) == 0 {
		return ""
	}
	var meta map[string]interface{}
	if json.Unmarshal(raw, &meta) != nil {
		return ""
	}
	value, _ := meta[key].(string)
	return strings.TrimSpace(value)
}

func resolveReceiptPhone(metadataPhone, orderPhone, cartGuestPhone, customerPhone string) string {
	for _, phone := range []string{metadataPhone, orderPhone, cartGuestPhone, customerPhone} {
		if strings.TrimSpace(phone) != "" {
			return strings.TrimSpace(phone)
		}
	}
	return ""
}

type syncedReceiptContact struct {
	metaPhone     string
	metaEmail     string
	orderPhone    string
	orderEmail    string
	cartPhone     string
	cartEmail     string
	customerPhone string
	customerEmail string
	customerID    pgtype.Text
}

func loadSyncedReceiptContact(ctx context.Context, pool *pgxpool.Pool, txID string) (syncedReceiptContact, error) {
	var contact syncedReceiptContact
	err := pool.QueryRow(ctx, `
		SELECT
			COALESCE(pt.metadata->>'customer_phone', ''),
			COALESCE(pt.metadata->>'customer_email', ''),
			COALESCE(so.customer_phone, ''),
			COALESCE(so.customer_email, ''),
			COALESCE(cart.guest_phone, ''),
			COALESCE(cart.guest_email, ''),
			COALESCE(cu.phone, ''),
			COALESCE(cu.email, ''),
			pt.customer_id::text
		FROM pos_transactions pt
		LEFT JOIN sales_orders_v2 so ON so.id = pt.sales_order_id
		LEFT JOIN carts cart ON cart.id = pt.source_cart_id
		LEFT JOIN customers cu ON cu.id = pt.customer_id
		WHERE pt.id::text = $1
	`, txID).Scan(
		&contact.metaPhone, &contact.metaEmail,
		&contact.orderPhone, &contact.orderEmail,
		&contact.cartPhone, &contact.cartEmail,
		&contact.customerPhone, &contact.customerEmail,
		&contact.customerID,
	)
	return contact, err
}

// posReceiptContactEligible mirrors POS CreateTransaction: enqueue when the
// sale has a customer id or a reachable phone/email on the transaction, order,
// cart guest fields, or linked customer record.
func posReceiptContactEligible(ctx context.Context, pool *pgxpool.Pool, txID string, payloadCustomerID *string, payloadMetaPhone string) (bool, string, string) {
	if pool == nil || strings.TrimSpace(txID) == "" {
		return false, "", "transaction id missing"
	}
	if payloadCustomerID != nil && strings.TrimSpace(*payloadCustomerID) != "" {
		contact, err := loadSyncedReceiptContact(ctx, pool, txID)
		if err != nil {
			return true, resolveReceiptPhone(payloadMetaPhone, "", "", ""), ""
		}
		phone := resolveReceiptPhone(payloadMetaPhone, contact.orderPhone, contact.cartPhone, contact.customerPhone)
		return true, phone, ""
	}
	contact, err := loadSyncedReceiptContact(ctx, pool, txID)
	if err != nil {
		return false, "", "transaction not found"
	}
	if contact.customerID.Valid && strings.TrimSpace(contact.customerID.String) != "" {
		phone := resolveReceiptPhone(payloadMetaPhone, contact.orderPhone, contact.cartPhone, contact.customerPhone)
		return true, phone, ""
	}
	for _, value := range []string{
		payloadMetaPhone,
		contact.metaPhone, contact.metaEmail,
		contact.orderPhone, contact.orderEmail,
		contact.cartPhone, contact.cartEmail,
		contact.customerPhone, contact.customerEmail,
	} {
		if strings.TrimSpace(value) != "" {
			phone := resolveReceiptPhone(payloadMetaPhone, contact.orderPhone, contact.cartPhone, contact.customerPhone)
			return true, phone, ""
		}
	}
	return false, "", "no customer contact on transaction, order, cart, or customer"
}

// EnqueuePosReceiptNotification inserts one pending pos.receipt event for a
// completed sale. The idempotency key ignores repeated sync upserts.
func EnqueuePosReceiptNotification(ctx context.Context, pool *pgxpool.Pool, payload []byte) error {
	id, number, storeID, orgID, customerID, metaPhone, ok := posReceiptDecision(payload)
	if !ok || pool == nil {
		return nil
	}
	return enqueuePosReceiptForTransaction(ctx, pool, id, number, storeID, orgID, customerID, metaPhone)
}

func enqueuePosReceiptForTransaction(ctx context.Context, pool *pgxpool.Pool, id string, number string, storeID, orgID int32, customerID *string, payloadMetaPhone string) error {
	if orgID == 0 && storeID > 0 {
		if err := pool.QueryRow(ctx, `SELECT organization_id FROM stores WHERE id = $1`, storeID).Scan(&orgID); err != nil {
			return fmt.Errorf("receipt organization for store %d: %w", storeID, err)
		}
	}
	if orgID == 0 {
		return fmt.Errorf("receipt organization missing for transaction %s", id)
	}
	eligible, phone, reason := posReceiptContactEligible(ctx, pool, id, customerID, payloadMetaPhone)
	if !eligible {
		log.Printf("[gRPC SyncServer] skip pos.receipt for transaction %s (%s): %s", id, number, reason)
		return nil
	}
	body := map[string]any{
		"transaction_id":     id,
		"transaction_number": number,
		"organization_id":    orgID,
	}
	if phone != "" {
		body["customer_phone"] = phone
	}
	if customerID != nil {
		body["customer_id"] = *customerID
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return err
	}
	_, err = repository.New(pool).EnqueueNotificationEvent(ctx, repository.EnqueueNotificationEventParams{
		OrganizationID: orgID,
		EventTypeID:    "pos.receipt",
		EntityType:     "pos_transaction",
		EntityID:       id,
		IdempotencyKey: fmt.Sprintf("pos.receipt:%s:%s", id, number),
		Payload:        encoded,
	})
	return err
}

// EnqueuePosReceiptAfterLineSync retries receipt enqueue once lines land; the
// header sync may run before order joins are visible in the same push batch.
func EnqueuePosReceiptAfterLineSync(ctx context.Context, pool *pgxpool.Pool, linePayload []byte) {
	if pool == nil || len(linePayload) == 0 {
		return
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(linePayload, &raw); err != nil {
		return
	}
	txID := jsonEntityIDString(raw["transaction_id"])
	if txID == "" {
		return
	}
	var number string
	var storeID, orgID int32
	var customerID pgtype.Text
	err := pool.QueryRow(ctx, `
		SELECT transaction_number, store_id, COALESCE(
			(SELECT organization_id FROM stores s WHERE s.id = pt.store_id), 0
		), customer_id::text
		FROM pos_transactions pt
		WHERE pt.id::text = $1
	`, txID).Scan(&number, &storeID, &orgID, &customerID)
	if err != nil || strings.TrimSpace(number) == "" {
		return
	}
	var custPtr *string
	if customerID.Valid && strings.TrimSpace(customerID.String) != "" {
		val := strings.TrimSpace(customerID.String)
		custPtr = &val
	}
	if err := enqueuePosReceiptForTransaction(ctx, pool, txID, number, storeID, orgID, custPtr, ""); err != nil {
		log.Printf("[gRPC SyncServer] pos receipt notification enqueue (lines): %v", err)
	}
}

func enqueuePosReceiptAfterSync(ctx context.Context, pool *pgxpool.Pool, payload []byte) {
	if err := EnqueuePosReceiptNotification(ctx, pool, payload); err != nil {
		log.Printf("[gRPC SyncServer] pos receipt notification enqueue: %v", err)
	}
}
