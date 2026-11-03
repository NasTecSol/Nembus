package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// posReceiptDecision reports whether a synced pos_transactions payload should
// enqueue a customer receipt. Lines and the ZATCA QR are loaded when the
// notification is sent.
func posReceiptDecision(payload []byte) (id int64, number string, storeID int32, orgID int32, customerID *int32, custPhone string, enqueue bool) {
	if len(payload) == 0 {
		return 0, "", 0, 0, nil, "", false
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		return 0, "", 0, 0, nil, "", false
	}
	id = jsonInt64(raw["id"])
	number = jsonString(raw["transaction_number"])
	storeID = int32(jsonInt64(raw["store_id"]))
	orgID = int32(jsonInt64(raw["organization_id"]))
	if n := jsonInt64(raw["customer_id"]); n > 0 {
		v := int32(n)
		customerID = &v
	}
	if len(raw["metadata"]) > 0 {
		var meta map[string]interface{}
		if json.Unmarshal(raw["metadata"], &meta) == nil {
			if p, ok := meta["customer_phone"].(string); ok {
				custPhone = strings.TrimSpace(p)
			}
		}
	}
	status := strings.ToLower(strings.TrimSpace(jsonString(raw["status"])))
	if status != "" && status != "completed" {
		return id, number, storeID, orgID, customerID, custPhone, false
	}
	if jsonPresent(raw["voided_at"]) || jsonPresent(raw["voided_by"]) {
		return id, number, storeID, orgID, customerID, custPhone, false
	}
	if id == 0 || number == "" {
		return id, number, storeID, orgID, customerID, custPhone, false
	}
	return id, number, storeID, orgID, customerID, custPhone, true
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

func resolveReceiptPhone(metadataPhone, orderPhone, cartGuestPhone, customerPhone string) string {
	for _, phone := range []string{metadataPhone, orderPhone, cartGuestPhone, customerPhone} {
		if strings.TrimSpace(phone) != "" {
			return strings.TrimSpace(phone)
		}
	}
	return ""
}

// resolveSyncedTransactionPhone loads contact from cloud rows after cart/order/txn upsert.
func resolveSyncedTransactionPhone(ctx context.Context, pool *pgxpool.Pool, txnID int64, payloadMetaPhone string) (string, error) {
	var metaPhone, orderPhone, cartPhone, customerPhone string
	err := pool.QueryRow(ctx, `
		SELECT
			COALESCE(pt.metadata->>'customer_phone', ''),
			COALESCE(so.customer_phone, ''),
			COALESCE(cart.guest_phone, ''),
			COALESCE(cu.phone, '')
		FROM pos_transactions pt
		LEFT JOIN sales_orders_v2 so ON so.id = pt.sales_order_id
		LEFT JOIN carts cart ON cart.id = pt.source_cart_id
		LEFT JOIN customers cu ON cu.id = pt.customer_id
		WHERE pt.id = $1
	`, txnID).Scan(&metaPhone, &orderPhone, &cartPhone, &customerPhone)
	if err != nil {
		return "", err
	}
	return resolveReceiptPhone(payloadMetaPhone, orderPhone, cartPhone, customerPhone), nil
}

// EnqueuePosReceiptNotification inserts one pending pos.receipt event for a
// completed sale when a customer phone is available on cloud. The idempotency
// key ignores repeated sync upserts.
func EnqueuePosReceiptNotification(ctx context.Context, pool *pgxpool.Pool, payload []byte) error {
	id, number, storeID, orgID, customerID, custPhone, ok := posReceiptDecision(payload)
	if !ok || pool == nil {
		return nil
	}
	if orgID == 0 && storeID > 0 {
		if err := pool.QueryRow(ctx, `SELECT organization_id FROM stores WHERE id = $1`, storeID).Scan(&orgID); err != nil {
			return fmt.Errorf("receipt organization for store %d: %w", storeID, err)
		}
	}
	if orgID == 0 {
		return fmt.Errorf("receipt organization missing for transaction %d", id)
	}
	phone, err := resolveSyncedTransactionPhone(ctx, pool, id, custPhone)
	if err != nil {
		return fmt.Errorf("receipt contact for transaction %d: %w", id, err)
	}
	if phone == "" {
		return nil
	}
	body := map[string]any{
		"transaction_id":     id,
		"transaction_number": number,
		"organization_id":    orgID,
		"customer_phone":     phone,
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
		EntityID:       fmt.Sprintf("%d", id),
		IdempotencyKey: fmt.Sprintf("pos.receipt:%d:%s", id, number),
		Payload:        encoded,
	})
	return err
}

func enqueuePosReceiptAfterSync(ctx context.Context, pool *pgxpool.Pool, payload []byte) {
	if err := EnqueuePosReceiptNotification(ctx, pool, payload); err != nil {
		log.Printf("[gRPC SyncServer] pos receipt notification enqueue: %v", err)
	}
}
