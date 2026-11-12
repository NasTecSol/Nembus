package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// TryInRestaurantTx runs fn inside a PostgreSQL transaction when q is backed
// by a pool, connection, or transaction supporting Begin. Returns handled=false
// when q.db cannot begin transactions (e.g., stub DBTX in unit tests).
func (q *Queries) TryInRestaurantTx(ctx context.Context, fn func(*Queries) error) (handled bool, err error) {
	beginner, ok := q.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return false, nil
	}
	tx, err := beginner.Begin(ctx)
	if err != nil {
		return true, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(q.WithTx(tx)); err != nil {
		return true, err
	}
	return true, tx.Commit(ctx)
}

// LockRestaurantCart acquires a row-level FOR UPDATE lock on the restaurant_carts
// row so concurrent mutations and checkout serialize on the same cart lock.
func (q *Queries) LockRestaurantCart(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var lockedID uuid.UUID
	err := q.db.QueryRow(ctx, `/* LockRestaurantCart */ SELECT id FROM restaurant_carts WHERE id = $1 FOR UPDATE`, id).Scan(&lockedID)
	return lockedID, err
}

// LockRestaurantOrder acquires a row-level FOR UPDATE lock on the restaurant_orders
// row so concurrent status changes, stock reversals, and settlements serialize.
func (q *Queries) LockRestaurantOrder(ctx context.Context, id int32) (int32, error) {
	var lockedID int32
	err := q.db.QueryRow(ctx, `/* LockRestaurantOrder */ SELECT id FROM restaurant_orders WHERE id = $1 FOR UPDATE`, id).Scan(&lockedID)
	return lockedID, err
}

// AcquireRestaurantCartScopeLock serializes lookup-or-create for active carts
// keyed by store and table/customer/guest when no cart_id is supplied.
func (q *Queries) AcquireRestaurantCartScopeLock(ctx context.Context, scopeKey string) error {
	if _, ok := q.db.(pgx.Tx); !ok || scopeKey == "" {
		return nil
	}
	_, err := q.db.Exec(ctx, `/* AcquireRestaurantCartScopeLock */ SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, scopeKey)
	return err
}
