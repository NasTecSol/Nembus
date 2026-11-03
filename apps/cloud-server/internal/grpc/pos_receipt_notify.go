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
func posReceiptDecision(payload []byte) (id int64, number string, storeID int32, orgID int32, customerID *int32, enqueue bool) {
	if len(payload) == 0 {
		return 0, "", 0, 0, nil, false
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		return 0, "", 0, 0, nil, false
	}
	id = jsonInt64(raw["id"])
	number = jsonString(raw["transaction_number"])
	storeID = int32(jsonInt64(raw["store_id"]))
	orgID = int32(jsonInt64(raw["organization_id"]))
	if n := jsonInt64(raw["customer_id"]); n > 0 {
		v := int32(n)
		customerID = &v
	}
	status := strings.ToLower(strings.TrimSpace(jsonString(raw["status"])))
	if status != "" && status != "completed" {
		return id, number, storeID, orgID, customerID, false
	}
	if jsonPresent(raw["voided_at"]) || jsonPresent(raw["voided_by"]) {
		return id, number, storeID, orgID, customerID, false
	}
	if id == 0 || number == "" {
		return id, number, storeID, orgID, customerID, false
	}
	return id, number, storeID, orgID, customerID, true
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

// EnqueuePosReceiptNotification inserts one pending pos.receipt event for a
// completed sale. The idempotency key ignores repeated sync upserts.
func EnqueuePosReceiptNotification(ctx context.Context, pool *pgxpool.Pool, payload []byte) error {
	id, number, storeID, orgID, customerID, ok := posReceiptDecision(payload)
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
	body := map[string]any{
		"transaction_id":     id,
		"transaction_number": number,
		"organization_id":    orgID,
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
