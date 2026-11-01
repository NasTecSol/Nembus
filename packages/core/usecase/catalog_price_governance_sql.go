package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type sqlExecer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

// SubmitPendingCatalogPriceApproval upserts a pending catalog.price_change approval request.
func SubmitPendingCatalogPriceApproval(ctx context.Context, db sqlExecer, orgID int32, productID int32, sku string, payload []byte, summary string) error {
	if summary == "" {
		summary = fmt.Sprintf("Price/Cost revision for SKU %s", sku)
	}
	entityID := fmt.Sprintf("%d", productID)
	upd, err := db.Exec(ctx, `
		UPDATE approval_requests
		SET payload = $4, summary = $3, updated_at = NOW()
		WHERE organization_id = $1 AND domain = 'catalog' AND action = 'price_change'
		  AND entity_type = 'product' AND entity_id = $2 AND status = 'pending'
	`, orgID, entityID, summary, payload)
	if err != nil {
		return err
	}
	if upd.RowsAffected() == 0 {
		_, err = db.Exec(ctx, `
			INSERT INTO approval_requests (
				organization_id, domain, action, entity_type, entity_id,
				status, required_permission, summary, payload, metadata
			) VALUES (
				$1, 'catalog', 'price_change', 'product', $2,
				'pending', 'catalog:pricing:approve', $3, $4, '{}'::jsonb
			)`, orgID, entityID, summary, payload)
	}
	return err
}

// LoadCatalogPricePolicy reads catalog.price_change policy for an organization.
func LoadCatalogPricePolicy(ctx context.Context, tx pgx.Tx, orgID int32) CatalogPricePolicyRules {
	var requiresApproval bool
	var rawRules []byte
	err := tx.QueryRow(ctx, `
		SELECT requires_approval, rules FROM approval_policies
		WHERE organization_id = $1 AND domain = 'catalog' AND action = 'price_change' AND is_enabled = true
		LIMIT 1`, orgID).Scan(&requiresApproval, &rawRules)
	if err != nil {
		return ParseCatalogPricePolicy(nil, false)
	}
	rules := map[string]any{}
	if len(rawRules) > 0 {
		_ = json.Unmarshal(rawRules, &rules)
	}
	return ParseCatalogPricePolicy(rules, requiresApproval)
}
