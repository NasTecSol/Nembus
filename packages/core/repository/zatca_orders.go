package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// InZatcaTransaction keeps document creation, source locking, and device-chain
// allocation on one connection. Nested callers use pgx savepoints.
func (q *Queries) InZatcaTransaction(ctx context.Context, fn func(*Queries) error) error {
	beginner, ok := q.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return fmt.Errorf("ZATCA document generation requires a transactional repository")
	}
	tx, err := beginner.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(q.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (q *Queries) LockZatcaOrder(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, `/* LockZatcaOrder */ SELECT id FROM sales_orders_v2 WHERE id=$1 FOR UPDATE`, id)
	return err
}

func (q *Queries) LockZatcaDevice(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, `/* LockZatcaDevice */ SELECT id FROM zatca_device_configs WHERE id=$1 FOR UPDATE`, id)
	return err
}

// Serialize reporting attempts for one saved invoice across requests and
// workers. The caller must hold the transaction through the gateway response.
func (q *Queries) LockZatcaChainEntry(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, `/* LockZatcaChainEntry */ SELECT id FROM zatca_document_chain WHERE id=$1 FOR UPDATE`, id)
	return err
}

// Find an older issued POS invoice for this order. Never issue it again under a
// new entity type, even if it has not yet been reported.
func (q *Queries) GetLegacyZatcaOrderTransaction(ctx context.Context, id uuid.UUID) (int32, error) {
	var transactionID int32
	err := q.db.QueryRow(ctx, `/* GetLegacyZatcaOrderTransaction */
 SELECT t.id FROM pos_transactions t JOIN zatca_document_chain c
 ON c.entity_type='pos_transaction' AND c.entity_id=t.id::text
 WHERE t.sales_order_id=$1 ORDER BY c.id LIMIT 1`, id).Scan(&transactionID)
	return transactionID, err
}
