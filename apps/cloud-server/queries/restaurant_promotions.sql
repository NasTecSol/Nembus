-- =====================================================
-- SQLC QUERIES FOR RESTAURANT PROMOTIONS
-- Dedicated Restaurant Menu Item & Category Promotion Management
-- =====================================================

-- name: CreateRestaurantPromotion :one
INSERT INTO restaurant_promotions (
    store_id, code, name, description, promotion_type,
    action_metadata, valid_from, valid_to, schedule_json,
    applies_to, target_menu_item_ids, target_menu_category_ids,
    target_customer_types, target_customer_tiers,
    min_order_amount, min_quantity, coupon_code,
    usage_limit, usage_per_customer, discount_value, is_stackable, is_active,
    created_by, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
) RETURNING *;

-- name: GetRestaurantPromotion :one
SELECT * FROM restaurant_promotions
WHERE id = $1;

-- name: GetRestaurantPromotionByCode :one
SELECT * FROM restaurant_promotions
WHERE code = $1
  AND store_id = $2;

-- name: ListActiveRestaurantPromotions :many
SELECT * FROM restaurant_promotions
WHERE store_id = $1
  AND is_active = true
ORDER BY created_at DESC;

-- name: ListAllRestaurantPromotions :many
SELECT * FROM restaurant_promotions
WHERE store_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateRestaurantPromotion :one
UPDATE restaurant_promotions
SET name = $2,
    description = $3,
    action_metadata = $4,
    valid_from = $5,
    valid_to = $6,
    schedule_json = $7,
    applies_to = $8,
    target_menu_item_ids = $9,
    target_menu_category_ids = $10,
    target_customer_types = $11,
    target_customer_tiers = $12,
    min_order_amount = $13,
    min_quantity = $14,
    usage_limit = $15,
    usage_per_customer = $16,
    discount_value = $17,
    is_stackable = $18,
    is_active = $19,
    metadata = $20,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateRestaurantPromotionStatus :one
UPDATE restaurant_promotions
SET is_active = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteRestaurantPromotion :exec
DELETE FROM restaurant_promotions
WHERE id = $1;

-- name: GetActiveRestaurantPromotionByCouponCode :one
SELECT * FROM restaurant_promotions
WHERE coupon_code = $1
  AND store_id = $2
  AND is_active = true
  AND (valid_from IS NULL OR valid_from <= CURRENT_TIMESTAMP)
  AND (valid_to IS NULL OR valid_to >= CURRENT_TIMESTAMP)
  AND (usage_limit IS NULL OR usage_count < usage_limit);

-- name: IncrementRestaurantPromotionUsage :one
UPDATE restaurant_promotions
SET usage_count = usage_count + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
