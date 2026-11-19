package manager

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed order_totals_patch.sql
var orderTotalsPatchSQL string

// RepairNullTaxOrderTotals upgrades existing tenants and repairs only unissued
// order headers with an existing failed ZATCA job matching the old calculation.
func RepairNullTaxOrderTotals(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return nil
	}
	if _, err := pool.Exec(ctx, orderTotalsPatchSQL); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42501" {
			// Runtime tenant credentials may update data without owning existing
			// trigger functions. This optional compatibility patch must not make
			// the tenant unavailable; migrations should apply it as the owner.
			log.Printf("[tenant patch] Skipping NULL-tax order-total patch: database role does not own the existing function")
			return nil
		}
		return fmt.Errorf("patch NULL-tax order totals: %w", err)
	}
	return nil
}
