-- restaurant_carts.sql
-- SQLC queries for Restaurant Cart -> Order flow (matching retail Carts API parity)

-- =====================================================
-- RESTAURANT CART MANAGEMENT
-- =====================================================

-- name: CreateRestaurantCart :one
INSERT INTO restaurant_carts (
    cart_number,
    organization_id,
    store_id,
    table_id,
    order_type,
    customer_id,
    guest_identifier,
    guest_name,
    guest_email,
    guest_phone,
    cart_status,
    channel,
    payment_method,
    payment_gateway,
    device_info,
    created_by_user_id,
    cashier_id,
    pos_terminal_id,
    kiosk_session_id,
    coupon_code,
    discount_code,
    delivery_address,
    expires_at,
    notes,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25
) RETURNING *;

-- name: GetRestaurantCartByID :one
SELECT * FROM restaurant_carts
WHERE id = $1 LIMIT 1;

-- name: GetRestaurantCartByNumber :one
SELECT * FROM restaurant_carts
WHERE cart_number = $1 LIMIT 1;

-- name: GetRestaurantCartByCustomer :one
SELECT * FROM restaurant_carts
WHERE customer_id = $1
  AND cart_status = 'active'
  AND store_id = $2
ORDER BY last_activity_at DESC
LIMIT 1;

-- name: GetRestaurantCartByGuestIdentifier :one
SELECT * FROM restaurant_carts
WHERE guest_identifier = $1
  AND cart_status = 'active'
  AND store_id = $2
ORDER BY last_activity_at DESC
LIMIT 1;

-- name: GetRestaurantCartByTable :one
SELECT * FROM restaurant_carts
WHERE table_id = $1
  AND cart_status = 'active'
  AND store_id = $2
ORDER BY last_activity_at DESC
LIMIT 1;

-- name: ListActiveRestaurantCarts :many
SELECT * FROM restaurant_carts
WHERE store_id = $1
  AND cart_status IN ('active', 'draft')
  AND last_activity_at > NOW() - INTERVAL '24 hours'
ORDER BY last_activity_at DESC
LIMIT $2 OFFSET $3;

-- name: ListDraftRestaurantCartsByCashier :many
SELECT * FROM restaurant_carts
WHERE cashier_id = $1
  AND cart_status = 'draft'
  AND (sqlc.narg(store_id)::int IS NULL OR store_id = sqlc.narg(store_id)::int)
ORDER BY last_activity_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateRestaurantCart :one
UPDATE restaurant_carts
SET
    table_id = COALESCE($2, table_id),
    order_type = COALESCE(NULLIF(TRIM(sqlc.narg(order_type)::text), ''), order_type),
    customer_id = COALESCE($3, customer_id),
    guest_identifier = COALESCE($4, guest_identifier),
    guest_name = COALESCE($5, guest_name),
    guest_email = COALESCE($6, guest_email),
    guest_phone = COALESCE($7, guest_phone),
    channel = COALESCE($8, channel),
    payment_method = COALESCE($9, payment_method),
    payment_gateway = COALESCE($10, payment_gateway),
    coupon_code = COALESCE($11, coupon_code),
    discount_code = COALESCE($12, discount_code),
    delivery_address = COALESCE($13, delivery_address),
    notes = COALESCE($14, notes),
    metadata = COALESCE($15, metadata),
    last_activity_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateRestaurantCartStatus :one
UPDATE restaurant_carts
SET
    cart_status = $2,
    last_activity_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateRestaurantCartCustomer :one
UPDATE restaurant_carts
SET
    customer_id = $2,
    guest_identifier = NULL,
    last_activity_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateRestaurantCartTable :one
UPDATE restaurant_carts
SET
    table_id = $2,
    last_activity_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteRestaurantCart :exec
DELETE FROM restaurant_carts
WHERE id = $1;

-- name: ExpireAbandonedRestaurantCarts :exec
UPDATE restaurant_carts
SET cart_status = 'expired', updated_at = CURRENT_TIMESTAMP
WHERE cart_status IN ('active', 'draft')
  AND expires_at IS NOT NULL
  AND expires_at < CURRENT_TIMESTAMP;

-- =====================================================
-- RESTAURANT CART ITEMS
-- =====================================================

-- name: CreateRestaurantCartItem :one
INSERT INTO restaurant_cart_items (
    cart_id,
    organization_id,
    menu_item_id,
    quantity,
    unit_price,
    modifiers_snapshot,
    modifiers_total,
    combo_components_snapshot,
    combo_total,
    discount_amount,
    tax_category_id,
    tax_rate,
    tax_amount,
    line_total,
    customization_details,
    notes,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17
) RETURNING *;

-- name: GetRestaurantCartItem :one
SELECT * FROM restaurant_cart_items
WHERE id = $1 LIMIT 1;

-- name: ListRestaurantCartItems :many
SELECT
    rci.*,
    mi.name AS menu_item_name,
    mi.short_name AS menu_item_short_name,
    mi.image_url AS menu_item_image_url,
    mi.item_type AS menu_item_type,
    mi.is_available AS menu_item_is_available,
    mi.is_active AS menu_item_is_active,
    mi.preparation_time_min,
    mc.name AS menu_category_name,
    tc.name AS tax_category_name,
    tc.is_inclusive AS tax_is_inclusive
FROM restaurant_cart_items rci
JOIN menu_items mi ON rci.menu_item_id = mi.id
LEFT JOIN menu_categories mc ON mi.menu_category_id = mc.id
LEFT JOIN tax_categories tc ON rci.tax_category_id = tc.id
WHERE rci.cart_id = $1
ORDER BY rci.added_at ASC;

-- name: FindMatchingRestaurantCartItem :one
SELECT * FROM restaurant_cart_items
WHERE cart_id = $1
  AND menu_item_id = $2
  AND COALESCE(modifiers_snapshot, '[]'::jsonb) = COALESCE($3::jsonb, '[]'::jsonb)
  AND COALESCE(combo_components_snapshot, '[]'::jsonb) = COALESCE($4::jsonb, '[]'::jsonb)
  AND COALESCE(notes, '') = COALESCE($5::text, '')
  AND unit_price = $6
  AND COALESCE(customization_details, '{}'::jsonb) = COALESCE($7::jsonb, '{}'::jsonb)
LIMIT 1;

-- name: UpdateRestaurantCartItem :one
UPDATE restaurant_cart_items
SET
    quantity = $2,
    unit_price = $3,
    modifiers_snapshot = $4,
    modifiers_total = $5,
    combo_components_snapshot = $6,
    combo_total = $7,
    discount_amount = $8,
    tax_category_id = $9,
    tax_rate = $10,
    tax_amount = $11,
    line_total = $12,
    customization_details = $13,
    notes = $14,
    metadata = $15,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateRestaurantCartItemQuantityAndTotals :one
UPDATE restaurant_cart_items
SET
    quantity = $2,
    tax_amount = $3,
    line_total = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteRestaurantCartItem :exec
DELETE FROM restaurant_cart_items
WHERE id = $1;

-- name: ClearRestaurantCartItems :exec
DELETE FROM restaurant_cart_items
WHERE cart_id = $1;

-- name: MergeGuestRestaurantCartItems :exec
UPDATE restaurant_cart_items
SET cart_id = $2, updated_at = CURRENT_TIMESTAMP
WHERE cart_id = $1;

-- name: GetRestaurantCartItemCount :one
SELECT COUNT(*) AS item_count, COALESCE(SUM(quantity), 0)::numeric(15,3) AS total_quantity
FROM restaurant_cart_items
WHERE cart_id = $1;

-- name: GetRestaurantCartTotals :one
SELECT
    COUNT(*) AS item_count,
    COALESCE(SUM(quantity), 0)::numeric(15,3) AS total_quantity,
    COALESCE(SUM(line_total - tax_amount), 0)::numeric(15,2) AS subtotal,
    COALESCE(SUM(discount_amount), 0)::numeric(15,2) AS total_discount,
    COALESCE(SUM(tax_amount), 0)::numeric(15,2) AS total_tax,
    COALESCE(SUM(line_total), 0)::numeric(15,2) AS total_amount
FROM restaurant_cart_items
WHERE cart_id = $1;

-- name: RecalculateRestaurantCartTotals :one
UPDATE restaurant_carts rc
SET
    subtotal = totals.subtotal,
    discount_amount = totals.total_discount,
    tax_amount = totals.total_tax,
    total_amount = totals.total_amount,
    last_activity_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
FROM (
    SELECT
        COALESCE(SUM(line_total - tax_amount), 0)::numeric(15,2) AS subtotal,
        COALESCE(SUM(discount_amount), 0)::numeric(15,2) AS total_discount,
        COALESCE(SUM(tax_amount), 0)::numeric(15,2) AS total_tax,
        COALESCE(SUM(line_total), 0)::numeric(15,2) AS total_amount
    FROM restaurant_cart_items
    WHERE cart_id = $1
) AS totals
WHERE rc.id = $1
RETURNING rc.*;

-- name: ConvertRestaurantCartToOrder :one
UPDATE restaurant_carts
SET
    cart_status = 'converted',
    converted_to_order_id = $2,
    converted_at = CURRENT_TIMESTAMP,
    last_activity_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND cart_status IN ('active', 'draft')
RETURNING *;

-- name: ReopenRestaurantCart :one
UPDATE restaurant_carts
SET
    cart_status = 'active',
    converted_to_order_id = NULL,
    converted_at = NULL,
    last_activity_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE converted_to_order_id = $1
  AND cart_status = 'converted'
RETURNING *;

-- name: ReopenRestaurantCartByID :one
UPDATE restaurant_carts
SET
    cart_status = 'active',
    converted_to_order_id = NULL,
    converted_at = NULL,
    last_activity_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ApplyCouponToRestaurantCart :one
UPDATE restaurant_carts
SET
    coupon_code = $2,
    discount_code = $3,
    last_activity_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ClearCouponFromRestaurantCart :one
UPDATE restaurant_carts
SET
    coupon_code = NULL,
    discount_code = NULL,
    last_activity_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateRestaurantCartItemDiscountAndTotals :one
UPDATE restaurant_cart_items
SET
    discount_amount = $2,
    tax_amount = $3,
    line_total = $4,
    metadata = COALESCE($5, metadata),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
