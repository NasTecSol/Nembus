-- =====================================================
-- SQLC QUERIES FOR PROMOTIONS
-- Coupon / Discount Promotion Management
-- =====================================================

-- name: CreatePromotion :one
INSERT INTO promotions (
    organization_id, code, name, description, promotion_type,
    action_metadata, valid_from, valid_to, schedule_json,
    applies_to, target_product_ids, target_category_ids,
    min_order_amount, min_quantity, coupon_code,
    usage_limit, discount_value, is_stackable, is_active,
    store_ids, created_by, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
) RETURNING *;

-- name: GetPromotion :one
SELECT * FROM promotions
WHERE id = $1;

-- name: GetPromotionByCode :one
SELECT * FROM promotions
WHERE code = $1
  AND organization_id = $2;

-- name: ListActivePromotions :many
SELECT * FROM promotions
WHERE organization_id = $1
  AND is_active = true
ORDER BY created_at DESC;

-- name: ListAllPromotions :many
SELECT * FROM promotions
WHERE organization_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdatePromotion :one
UPDATE promotions
SET name = $2,
    description = $3,
    action_metadata = $4,
    valid_from = $5,
    valid_to = $6,
    schedule_json = $7,
    applies_to = $8,
    target_product_ids = $9,
    target_category_ids = $10,
    min_order_amount = $11,
    min_quantity = $12,
    usage_limit = $13,
    discount_value = $14,
    is_stackable = $15,
    is_active = $16,
    store_ids = $17,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdatePromotionStatus :one
UPDATE promotions
SET is_active = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeletePromotion :exec
DELETE FROM promotions
WHERE id = $1;

-- name: GetActivePromotionByCouponCode :one
-- Fetches a coupon that is currently valid and within usage limits.
-- Safe to call before applying a coupon at checkout.
SELECT * FROM promotions
WHERE coupon_code = $1
  AND organization_id = $2
  AND is_active = true
  AND (valid_from IS NULL OR valid_from <= CURRENT_TIMESTAMP)
  AND (valid_to IS NULL OR valid_to >= CURRENT_TIMESTAMP)
  AND (usage_limit IS NULL OR usage_count < usage_limit);

-- name: IncrementPromotionUsage :one
-- Atomically increments usage_count. Call this during ConvertCartToOrder.
UPDATE promotions
SET usage_count = usage_count + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
