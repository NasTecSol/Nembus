package handler

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const posCustomerIDBase int64 = 1_000_000_000
const posCustomerIDsPerTerminal int64 = 100_000
const maxCustomerID int64 = 2_147_483_647

// prepareLocalCustomerID reserves a terminal-specific part of the existing
// customers serial sequence. No table, trigger, or migration is required.
func prepareLocalCustomerID(ctx context.Context, tx pgx.Tx) error {
	var terminalID int64
	if err := tx.QueryRow(ctx, `SELECT pos_terminal_id FROM local_device_config
		WHERE pos_terminal_id > 0 ORDER BY id DESC LIMIT 1`).Scan(&terminalID); err != nil {
		return fmt.Errorf("POS terminal must be configured before creating a customer: %w", err)
	}
	base := posCustomerIDBase + terminalID*posCustomerIDsPerTerminal
	if base+posCustomerIDsPerTerminal > maxCustomerID {
		return fmt.Errorf("POS terminal ID %d is outside the customer ID range", terminalID)
	}
	// Serialize local customer creation and recover the next ID from saved rows
	// after a clone or terminal change, without changing the database schema.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(723067123)`); err != nil {
		return fmt.Errorf("lock POS customer allocation: %w", err)
	}
	if _, err := tx.Exec(ctx, `SELECT setval(
		pg_get_serial_sequence('customers', 'id'),
		GREATEST($1::bigint, COALESCE((SELECT MAX(id)::bigint FROM customers
			WHERE id > $1 AND id <= $2), $1::bigint)), true)`, base, base+posCustomerIDsPerTerminal); err != nil {
		return fmt.Errorf("reserve POS customer ID: %w", err)
	}
	return nil
}
