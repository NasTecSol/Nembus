-- name: CreateLoyaltyRule :one
INSERT INTO loyalty_redemption_rules (
    organization_id,
    rule_name,
    points_earning_rate,
    points_redemption_rate,
    min_points_to_redeem,
    max_points_per_txn,
    max_redemption_percent,
    eligible_product_types,
    expiry_days,
    is_active,
    valid_from,
    valid_to,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetLoyaltyRule :one
SELECT * FROM loyalty_redemption_rules
WHERE id = $1;

-- name: GetActiveLoyaltyRule :one
-- Returns the current active rule for the organization (used at checkout to compute points value)
SELECT * FROM loyalty_redemption_rules
WHERE organization_id = $1
  AND is_active = true
  AND (valid_from IS NULL OR valid_from <= CURRENT_DATE)
  AND (valid_to IS NULL OR valid_to >= CURRENT_DATE)
ORDER BY created_at DESC
LIMIT 1;

-- name: ListLoyaltyRules :many
SELECT * FROM loyalty_redemption_rules
WHERE organization_id = $1
ORDER BY created_at DESC;

-- name: UpdateLoyaltyRule :one
UPDATE loyalty_redemption_rules
SET
    rule_name              = $2,
    points_earning_rate    = $3,
    points_redemption_rate = $4,
    min_points_to_redeem   = $5,
    max_points_per_txn     = $6,
    max_redemption_percent = $7,
    is_active              = $8,
    valid_from             = $9,
    valid_to               = $10,
    metadata               = $11,
    updated_at             = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ToggleLoyaltyRuleActive :one
UPDATE loyalty_redemption_rules
SET is_active = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteLoyaltyRule :exec
DELETE FROM loyalty_redemption_rules
WHERE id = $1;
