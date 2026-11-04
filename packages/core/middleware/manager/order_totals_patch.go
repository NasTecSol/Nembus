package manager

import (
	"context"
	_ "embed"
	"fmt"

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
		return fmt.Errorf("patch NULL-tax order totals: %w", err)
	}
	return nil
}
