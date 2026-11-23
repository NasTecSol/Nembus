package handler

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// prepareLocalCustomerID validates that a POS terminal is configured before
// creating a customer locally. Because customers.id is now a UUID generated via
// gen_random_uuid(), serial sequence reservation is no longer needed.
func prepareLocalCustomerID(ctx context.Context, tx pgx.Tx) error {
	var terminalID int64
	if err := tx.QueryRow(ctx, `SELECT pos_terminal_id FROM local_device_config
		WHERE pos_terminal_id > 0 ORDER BY id DESC LIMIT 1`).Scan(&terminalID); err != nil {
		return fmt.Errorf("POS terminal must be configured before creating a customer: %w", err)
	}
	return nil
}

