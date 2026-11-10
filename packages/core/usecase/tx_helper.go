package usecase

import (
	"context"
	"encoding/json"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func resolvePoolFromContext(ctx context.Context, fallback *pgxpool.Pool) *pgxpool.Pool {
	if fallback != nil {
		return fallback
	}
	if ctx == nil {
		return nil
	}
	if pool, ok := ctx.Value(middleware.PoolKey).(*pgxpool.Pool); ok && pool != nil {
		return pool
	}
	return nil
}

// runInSavepoint executes fn inside a PostgreSQL SAVEPOINT when tx is non-nil so
// that an optional or non-fatal SQL error does not leave tx in an aborted state (25P02).
func runInSavepoint(ctx context.Context, tx pgx.Tx, name string, fn func() error) error {
	if tx == nil {
		return fn()
	}
	if _, err := tx.Exec(ctx, "SAVEPOINT "+name); err != nil {
		return err
	}
	err := fn()
	if err != nil {
		_, _ = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+name)
		_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT "+name)
		return err
	}
	_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT "+name)
	return nil
}

// enqueuePOSCheckoutOutboxTx enqueues the synthesized carts and sales_orders_v2 (POS_CHECKOUT)
// outbox items inside the same local POS transaction and correlation group as the checkout.
func enqueuePOSCheckoutOutboxTx(ctx context.Context, tx pgx.Tx, corrID string, entityID int64, orderID uuid.UUID, cartID *uuid.UUID) {
	if tx == nil || corrID == "" {
		return
	}
	_ = runInSavepoint(ctx, tx, "sp_enqueue_pos_checkout_outbox", func() error {
		var hasSyncQueue bool
		if err := tx.QueryRow(ctx, `SELECT to_regclass('public.sync_queue') IS NOT NULL`).Scan(&hasSyncQueue); err != nil || !hasSyncQueue {
			return err
		}

		if cartID != nil && *cartID != uuid.Nil {
			var cartJSON []byte
			err := tx.QueryRow(ctx, `SELECT row_to_json(c)::text FROM carts c WHERE c.id = $1`, *cartID).Scan(&cartJSON)
			if err == nil && len(cartJSON) > 0 {
				var cartMap map[string]json.RawMessage
				if json.Unmarshal(cartJSON, &cartMap) == nil {
					cartMap["converted_to_order_id"] = json.RawMessage("null")
					cartMap["converted_at"] = json.RawMessage("null")
					if stripped, err := json.Marshal(cartMap); err == nil {
						_, _ = tx.Exec(ctx, `
							INSERT INTO sync_queue (entity_type, entity_id, action, payload, priority, status, correlation_id)
							VALUES ('carts', $1, 'INSERT', $2::jsonb, 1, 'pending', $3)
						`, entityID, string(stripped), corrID)
					}
				}
			}
		}

		if orderID != uuid.Nil {
			var orderPayload []byte
			err := tx.QueryRow(ctx, `
				SELECT json_build_object(
					'order', row_to_json(o),
					'lines', COALESCE((SELECT json_agg(l ORDER BY l.line_number)
						FROM sales_order_lines_v2 l WHERE l.sales_order_id = o.id), '[]'::json)
				)::text
				FROM sales_orders_v2 o
				WHERE o.id = $1
			`, orderID).Scan(&orderPayload)
			if err == nil && len(orderPayload) > 0 {
				_, _ = tx.Exec(ctx, `
					INSERT INTO sync_queue (entity_type, entity_id, action, payload, priority, status, correlation_id)
					VALUES ('sales_orders_v2', $1, 'POS_CHECKOUT', $2::jsonb, 1, 'pending', $3)
				`, entityID, string(orderPayload), corrID)
			}
		}

		if cartID != nil && *cartID != uuid.Nil {
			var cartJSON []byte
			err := tx.QueryRow(ctx, `SELECT row_to_json(c)::text FROM carts c WHERE c.id = $1`, *cartID).Scan(&cartJSON)
			if err == nil && len(cartJSON) > 0 {
				_, _ = tx.Exec(ctx, `
					INSERT INTO sync_queue (entity_type, entity_id, action, payload, priority, status, correlation_id)
					VALUES ('carts', $1, 'UPDATE', $2::jsonb, 1, 'pending', $3)
				`, entityID, string(cartJSON), corrID)
			}
		}

		return nil
	})
}
