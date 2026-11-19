package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// Repair only the known optional menu variant link during clone restore. Keep
// menu rows and valid links, and still create and validate the foreign key.
func executeRestoreStatement(ctx context.Context, conn statementExecutor, stmt string) error {
	_, err := conn.Exec(ctx, stmt)
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	normalized := strings.Join(strings.Fields(strings.ToUpper(stmt)), " ")
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" ||
		!strings.HasPrefix(normalized, "ALTER TABLE ") {
		return err
	}
	if pgErr.SchemaName == "public" && pgErr.TableName == "menu_items" &&
		pgErr.ConstraintName == "menu_items_product_variant_id_fkey" &&
		strings.Contains(normalized, " ON DELETE SET NULL") {
		repaired, repairErr := conn.Exec(ctx, `
UPDATE public.menu_items AS item
SET product_variant_id = NULL
WHERE item.product_variant_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM public.product_variants AS variant
      WHERE variant.id = item.product_variant_id
  );`)
		if repairErr != nil {
			return fmt.Errorf("repair orphaned menu variant links before validating foreign key: %w", repairErr)
		}
		log.Printf("[restore] Cleared %d missing product variant links in local menu items; menu rows preserved", repaired.RowsAffected())
		_, err = conn.Exec(ctx, stmt)
		return err
	}
	if pgErr.SchemaName == "public" && pgErr.TableName == "business_partners" &&
		pgErr.ConstraintName == "business_partners_currency_code_fkey" {
		repaired, repairErr := conn.Exec(ctx, `
UPDATE public.business_partners AS partner
SET currency_code = NULL
WHERE partner.currency_code IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM public.currencies AS currency
      WHERE currency.code = partner.currency_code
  );`)
		if repairErr != nil {
			return fmt.Errorf("repair orphaned partner currency codes before validating foreign key: %w", repairErr)
		}
		log.Printf("[restore] Cleared %d missing currency references in local business partners; partner rows preserved", repaired.RowsAffected())
		_, err = conn.Exec(ctx, stmt)
		return err
	}
	return err
}
