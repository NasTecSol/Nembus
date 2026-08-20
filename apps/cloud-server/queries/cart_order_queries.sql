-- =====================================================
-- SQLC QUERIES FOR ENHANCED SCHEMA
-- Carts, Orders, Invoices, Quotes, and Related Modules
-- =====================================================

-- =====================================================
-- CART MANAGEMENT
-- =====================================================

-- name: CreateCart :one
INSERT INTO carts (
    cart_number,
    organization_id,
    store_id,
    customer_id,
    guest_identifier,
    guest_email,
    guest_phone,
    cart_status,
    cart_type,
    channel,
    payment_method,
    payment_gateway,
    device_info,
    created_by_user_id,
    cashier_id,
    pos_terminal_id,
    shipping_address,
    billing_address,
    shipping_method,
    coupon_code,
    discount_code,
    expires_at,
    notes,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
) RETURNING *;

-- name: GetCartByID :one
SELECT * FROM carts
WHERE id = $1;

-- name: GetCartByNumber :one
SELECT * FROM carts
WHERE cart_number = $1;

-- name: GetCartByCustomer :one
SELECT * FROM carts
WHERE customer_id = $1
  AND cart_status = 'active'
  AND store_id = $2
ORDER BY last_activity_at DESC
LIMIT 1;

-- name: GetCartByGuestIdentifier :one
SELECT * FROM carts
WHERE guest_identifier = $1
  AND cart_status = 'active'
  AND store_id = $2
ORDER BY last_activity_at DESC
LIMIT 1;

-- name: ListActiveCarts :many
SELECT * FROM carts
WHERE store_id = $1
  AND cart_status IN ('active', 'draft')
  AND last_activity_at > NOW() - INTERVAL '24 hours'
ORDER BY last_activity_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAbandonedCarts :many
SELECT c.*, COUNT(ci.id) as item_count, COALESCE(SUM(ci.line_total), 0)::numeric as cart_value
FROM carts c
LEFT JOIN cart_items ci ON c.id = ci.cart_id
LEFT JOIN customers cu ON c.customer_id = cu.id
WHERE c.store_id = $1
  AND c.cart_status = 'active'
  AND c.last_activity_at < NOW() - INTERVAL '24 hours'
  AND (cu.email IS NOT NULL OR c.guest_email IS NOT NULL)
GROUP BY c.id
HAVING COALESCE(SUM(ci.line_total), 0) > $2
ORDER BY c.last_activity_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateCart :one
UPDATE carts
SET subtotal = $2,
    discount_amount = $3,
    tax_amount = $4,
    shipping_amount = $5,
    total_amount = $6,
    coupon_code = $7,
    discount_code = $8,
    promotional_credits = $9,
    shipping_address = $10,
    billing_address = $11,
    shipping_method = $12,
    payment_method = $13,
    payment_gateway = $14,
    last_activity_at = NOW(),
    notes = $15,
    metadata = $16,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCartStatus :one
UPDATE carts
SET cart_status = $2,
    converted_to_order_id = $3,
    converted_at = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCartCustomer :one
UPDATE carts
SET customer_id = $2,
    guest_identifier = NULL,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCart :exec
DELETE FROM carts
WHERE id = $1;

-- name: ExpireAbandonedCarts :exec
UPDATE carts
SET cart_status = 'expired',
    updated_at = NOW()
WHERE cart_status = 'active'
  AND last_activity_at < NOW() - INTERVAL '7 days'
  AND store_id = $1;

-- =====================================================
-- CART ITEMS
-- =====================================================

-- name: CreateCartItem :one
INSERT INTO cart_items (
    cart_id,
    organization_id,
    product_id,
    product_variant_id,
    quantity,
    uom_id,
    unit_price,
    discount_amount,
    tax_amount,
    line_total,
    price_list_id,
    tax_category_id,
    batch_number,
    serial_number,
    customization_details,
    notes,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17
) RETURNING *;

-- name: GetCartItem :one
SELECT * FROM cart_items
WHERE id = $1;

-- name: ListCartItems :many
SELECT ci.*, p.name as product_name, p.sku as product_sku
FROM cart_items ci
JOIN products p ON ci.product_id = p.id
WHERE ci.cart_id = $1
ORDER BY ci.added_at;

-- name: GetCartItemByProduct :one
SELECT * FROM cart_items
WHERE cart_id = $1
  AND product_id = $2
  AND COALESCE(product_variant_id, 0) = COALESCE($3, 0)
  AND COALESCE(batch_number, '') = COALESCE($4, '')
  AND COALESCE(serial_number, '') = COALESCE($5, '');

-- name: UpdateCartItem :one
UPDATE cart_items
SET quantity = $2,
    unit_price = $3,
    discount_amount = $4,
    tax_amount = $5,
    line_total = $6,
    notes = $7,
    metadata = $8,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCartItemQuantity :one
UPDATE cart_items
SET discount_amount = CASE WHEN quantity > 0 THEN COALESCE(discount_amount, 0) / quantity * (quantity + $2) ELSE COALESCE(discount_amount, 0) END,
    tax_amount = CASE WHEN quantity > 0 THEN COALESCE(tax_amount, 0) / quantity * (quantity + $2) ELSE COALESCE(tax_amount, 0) END,
    quantity = quantity + $2,
    line_total = COALESCE(unit_price, 0) * (quantity + $2)
                 - CASE WHEN quantity > 0 THEN COALESCE(discount_amount, 0) / quantity * (quantity + $2) ELSE COALESCE(discount_amount, 0) END
                 + CASE WHEN quantity > 0 THEN COALESCE(tax_amount, 0) / quantity * (quantity + $2) ELSE COALESCE(tax_amount, 0) END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCartItem :exec
DELETE FROM cart_items
WHERE id = $1;

-- name: ClearCartItems :exec
DELETE FROM cart_items
WHERE cart_id = $1;

-- name: GetCartItemCount :one
SELECT COUNT(*) as item_count, COALESCE(SUM(quantity), 0) as total_quantity
FROM cart_items
WHERE cart_id = $1;

-- name: GetCartTotals :one
SELECT 
    COALESCE(SUM(line_total - tax_amount), 0) as subtotal,
    COALESCE(SUM(discount_amount), 0) as total_discount,
    COALESCE(SUM(tax_amount), 0) as total_tax,
    COALESCE(SUM(line_total), 0) as total
FROM cart_items
WHERE cart_id = $1;

-- =====================================================
-- CART ACTIVITY LOG
-- =====================================================

-- name: CreateCartActivity :one
INSERT INTO cart_activity_log (
    cart_id,
    organization_id,
    activity_type,
    description,
    performed_by_user_id,
    ip_address,
    user_agent,
    old_value,
    new_value
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: ListCartActivities :many
SELECT * FROM cart_activity_log
WHERE cart_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- =====================================================
-- DRAFT CART TEMPLATES
-- =====================================================

-- name: CreateDraftCartTemplate :one
INSERT INTO draft_cart_templates (
    organization_id,
    customer_id,
    template_name,
    description,
    template_type,
    is_favorite,
    auto_reorder_enabled,
    reorder_frequency_days,
    next_reorder_date,
    notes,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetDraftCartTemplate :one
SELECT * FROM draft_cart_templates
WHERE id = $1;

-- name: ListDraftCartTemplates :many
SELECT dct.*, COUNT(dcti.id) as item_count
FROM draft_cart_templates dct
LEFT JOIN draft_cart_template_items dcti ON dct.id = dcti.template_id
WHERE dct.customer_id = $1
  AND dct.organization_id = $2
GROUP BY dct.id
ORDER BY dct.is_favorite DESC, dct.updated_at DESC
LIMIT $3 OFFSET $4;

-- name: ListFavoriteDraftTemplates :many
SELECT * FROM draft_cart_templates
WHERE customer_id = $1
  AND organization_id = $2
  AND is_favorite = true
ORDER BY updated_at DESC;

-- name: ListAutoReorderTemplates :many
SELECT * FROM draft_cart_templates
WHERE organization_id = $1
  AND auto_reorder_enabled = true
  AND next_reorder_date <= $2
ORDER BY next_reorder_date;

-- name: UpdateDraftCartTemplate :one
UPDATE draft_cart_templates
SET template_name = $2,
    description = $3,
    template_type = $4,
    is_favorite = $5,
    auto_reorder_enabled = $6,
    reorder_frequency_days = $7,
    next_reorder_date = $8,
    notes = $9,
    metadata = $10,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateTemplateStats :one
UPDATE draft_cart_templates
SET total_items = $2,
    estimated_total = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteDraftCartTemplate :exec
DELETE FROM draft_cart_templates
WHERE id = $1;

-- =====================================================
-- DRAFT CART TEMPLATE ITEMS
-- =====================================================

-- name: CreateDraftCartTemplateItem :one
INSERT INTO draft_cart_template_items (
    template_id,
    organization_id,
    product_id,
    product_variant_id,
    quantity,
    uom_id,
    last_known_price,
    priority,
    notes,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetDraftCartTemplateItem :one
SELECT * FROM draft_cart_template_items
WHERE id = $1;

-- name: ListDraftCartTemplateItems :many
SELECT dcti.*, p.name as product_name, p.sku as product_sku, pp.price as current_price
FROM draft_cart_template_items dcti
JOIN products p ON dcti.product_id = p.id
LEFT JOIN product_prices pp ON dcti.product_id = pp.product_id 
    AND pp.is_active = true
WHERE dcti.template_id = $1
ORDER BY dcti.priority DESC, dcti.created_at;

-- name: UpdateDraftCartTemplateItem :one
UPDATE draft_cart_template_items
SET quantity = $2,
    last_known_price = $3,
    priority = $4,
    notes = $5,
    metadata = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteDraftCartTemplateItem :exec
DELETE FROM draft_cart_template_items
WHERE id = $1;

-- =====================================================
-- SALES ORDERS V2 (ENHANCED)
-- =====================================================

-- name: CreateSalesOrderV2 :one
INSERT INTO sales_orders_v2 (
    order_number,
    organization_id,
    store_id,
    customer_id,
    customer_name,
    customer_email,
    customer_phone,
    order_type,
    order_status,
    payment_status,
    fulfillment_status,
    sales_channel,
    order_source,
    referral_source,
    source_cart_id,
    created_by_user_id,
    assigned_to_user_id,
    order_date,
    expected_delivery_date,
    shipping_address,
    billing_address,
    shipping_method,
    payment_method,
    payment_gateway,
    payment_terms,
    payment_due_date,
    pos_terminal_id,
    cashier_id,
    is_gift,
    gift_message,
    special_instructions,
    internal_notes,
    tags,
    priority,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35
) RETURNING *;

-- name: GetSalesOrderV2 :one
SELECT * FROM sales_orders_v2
WHERE id = $1;

-- name: GetSalesOrderV2ByNumber :one
SELECT * FROM sales_orders_v2
WHERE order_number = $1;

-- name: ListSalesOrdersV2 :many
SELECT * FROM sales_orders_v2
WHERE organization_id = $1
  AND ($2::int IS NULL OR store_id = $2)
  AND ($3::order_status_v2 IS NULL OR order_status = $3)
  AND ($4::timestamp IS NULL OR order_date >= $4)
  AND ($5::timestamp IS NULL OR order_date <= $5)
ORDER BY order_date DESC, created_at DESC
LIMIT $6 OFFSET $7;

-- name: ListCustomerOrders :many
SELECT * FROM sales_orders_v2
WHERE customer_id = $1
  AND organization_id = $2
ORDER BY order_date DESC
LIMIT $3 OFFSET $4;

-- name: ListOrdersByStatus :many
SELECT so.*, c.name as customer_name_full, c.email as customer_email_full
FROM sales_orders_v2 so
LEFT JOIN customers c ON so.customer_id = c.id
WHERE so.store_id = $1
  AND so.order_status = $2
ORDER BY so.order_date DESC
LIMIT $3 OFFSET $4;

-- name: ListPendingOrders :many
SELECT * FROM sales_orders_v2
WHERE store_id = $1
  AND order_status IN ('pending', 'confirmed', 'processing')
  AND payment_status != 'paid'
ORDER BY order_date
LIMIT $2 OFFSET $3;

-- name: UpdateSalesOrderV2 :one
UPDATE sales_orders_v2
SET customer_id = $2,
    customer_name = $3,
    customer_email = $4,
    customer_phone = $5,
    expected_delivery_date = $6,
    shipping_address = $7,
    billing_address = $8,
    shipping_method = $9,
    payment_method = $10,
    payment_gateway = $11,
    special_instructions = $12,
    internal_notes = $13,
    tags = $14,
    priority = $15,
    metadata = $16,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateOrderStatus :one
UPDATE sales_orders_v2
SET order_status = $2,
    confirmed_date = CASE 
        WHEN order_status = 'confirmed'::order_status_v2 
        THEN NOW() 
        ELSE confirmed_date 
    END,
    cancelled_date = CASE 
        WHEN order_status = 'cancelled'::order_status_v2 
        THEN NOW() 
        ELSE cancelled_date 
    END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
-- name: UpdateOrderPaymentStatus :one
UPDATE sales_orders_v2
SET payment_status = $2,
    paid_amount = $3,
    payment_method = COALESCE($4, payment_method),
    payment_gateway = COALESCE($5, payment_gateway),
    balance_due = total_amount - $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateOrderFulfillmentStatus :one
UPDATE sales_orders_v2
SET fulfillment_status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateOrderTotals :one
UPDATE sales_orders_v2
SET subtotal = $2,
    discount_amount = $3,
    tax_amount = $4,
    shipping_amount = $5,
    adjustment_amount = $6,
    total_amount = $7,
    balance_due = $7 - COALESCE(paid_amount, 0),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateOrderDelivery :one
UPDATE sales_orders_v2
SET shipping_carrier = $2,
    tracking_number = $3,
    tracking_url = $4,
    actual_delivery_date = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: AssignOrder :one
UPDATE sales_orders_v2
SET assigned_to_user_id = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CancelOrder :one
UPDATE sales_orders_v2
SET order_status = 'cancelled',
    cancelled_date = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteSalesOrderV2 :exec
DELETE FROM sales_orders_v2
WHERE id = $1;

-- name: GetOrderStats :one
SELECT 
    COUNT(*) as total_orders,
    COUNT(*) FILTER (WHERE order_status = 'pending') as pending_orders,
    COUNT(*) FILTER (WHERE order_status = 'confirmed') as confirmed_orders,
    COUNT(*) FILTER (WHERE order_status = 'processing') as processing_orders,
    COUNT(*) FILTER (WHERE order_status = 'delivered') as delivered_orders,
    COUNT(*) FILTER (WHERE payment_status = 'unpaid') as unpaid_orders,
    COALESCE(SUM(total_amount), 0) as total_revenue,
    COALESCE(SUM(total_amount) FILTER (WHERE payment_status = 'paid'), 0) as paid_revenue,
    COALESCE(SUM(balance_due), 0) as outstanding_balance
FROM sales_orders_v2
WHERE store_id = $1
  AND order_date >= $2
  AND order_date <= $3;

-- =====================================================
-- SALES ORDER LINES V2
-- =====================================================

-- name: CreateSalesOrderLineV2 :one
INSERT INTO sales_order_lines_v2 (
    sales_order_id,
    organization_id,
    line_number,
    product_id,
    product_variant_id,
    product_name,
    product_sku,
    quantity_ordered,
    uom_id,
    unit_price,
    discount_amount,
    discount_percentage,
    tax_amount,
    line_total,
    tax_category_id,
    tax_rate,
    batch_number,
    serial_numbers,
    expiry_date,
    line_status,
    customization_details,
    unit_cost,
    notes,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24
) RETURNING *;

-- name: GetSalesOrderLineV2 :one
SELECT * FROM sales_order_lines_v2
WHERE id = $1;

-- name: ListSalesOrderLinesV2 :many
SELECT sol.*, p.name as product_name_current, p.sku as product_sku_current
FROM sales_order_lines_v2 sol
LEFT JOIN products p ON sol.product_id = p.id
WHERE sol.sales_order_id = $1
ORDER BY sol.line_number;

-- name: UpdateSalesOrderLineV2 :one
UPDATE sales_order_lines_v2
SET quantity_ordered = $2,
    unit_price = $3,
    discount_amount = $4,
    discount_percentage = $5,
    tax_amount = $6,
    line_total = $7,
    notes = $8,
    metadata = $9,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateOrderLineFulfillment :one
UPDATE sales_order_lines_v2
SET quantity_fulfilled = $2,
    line_status = CASE 
        WHEN $2 >= quantity_ordered THEN 'fulfilled'
        WHEN $2 > 0 THEN 'partially_fulfilled'
        ELSE line_status
    END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateOrderLineStatus :one
UPDATE sales_order_lines_v2
SET line_status = $2,
    quantity_cancelled = CASE WHEN $2 = 'cancelled' THEN quantity_ordered - quantity_fulfilled ELSE quantity_cancelled END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteSalesOrderLineV2 :exec
DELETE FROM sales_order_lines_v2
WHERE id = $1;

-- name: GetOrderLineTotals :one
SELECT 
    COALESCE(SUM(line_total - tax_amount), 0) as subtotal,
    COALESCE(SUM(discount_amount), 0) as total_discount,
    COALESCE(SUM(tax_amount), 0) as total_tax,
    COALESCE(SUM(line_total), 0) as total
FROM sales_order_lines_v2
WHERE sales_order_id = $1;

-- name: GetOrderLineMargin :one
SELECT 
    COALESCE(SUM(line_total), 0) as revenue,
    COALESCE(SUM(unit_cost * quantity_ordered), 0) as cost,
    COALESCE(SUM(line_total) - SUM(unit_cost * quantity_ordered), 0) as profit,
    CASE 
        WHEN SUM(line_total) > 0 THEN 
            ((SUM(line_total) - SUM(unit_cost * quantity_ordered)) / SUM(line_total)) * 100
        ELSE 0
    END as margin_percentage
FROM sales_order_lines_v2
WHERE sales_order_id = $1;

-- =====================================================
-- ORDER STATUS HISTORY
-- =====================================================

-- name: CreateOrderStatusHistory :one
INSERT INTO order_status_history (
    sales_order_id,
    organization_id,
    from_status,
    to_status,
    reason,
    notes,
    changed_by_user_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: ListOrderStatusHistory :many
SELECT osh.*, (u.first_name || ' ' || u.last_name) as changed_by_name
FROM order_status_history osh
LEFT JOIN users u ON osh.changed_by_user_id = u.id
WHERE osh.sales_order_id = $1
ORDER BY osh.changed_at DESC;

-- =====================================================
-- ORDER FULFILLMENTS
-- =====================================================

-- name: CreateOrderFulfillment :one
INSERT INTO order_fulfillments (
    sales_order_id,
    organization_id,
    fulfillment_number,
    fulfillment_status,
    shipment_status,
    fulfillment_store_id,
    shipping_carrier,
    shipping_method,
    tracking_number,
    tracking_url,
    estimated_delivery_date,
    notes,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetOrderFulfillment :one
SELECT * FROM order_fulfillments
WHERE id = $1;

-- name: GetOrderFulfillmentByNumber :one
SELECT * FROM order_fulfillments
WHERE fulfillment_number = $1;

-- name: ListOrderFulfillments :many
SELECT * FROM order_fulfillments
WHERE sales_order_id = $1
ORDER BY created_at DESC;

-- name: UpdateOrderFulfillment :one
UPDATE order_fulfillments
SET fulfillment_status = $2,
    shipment_status = $3,
    shipping_carrier = $4,
    tracking_number = $5,
    tracking_url = $6,
    estimated_delivery_date = $7,
    actual_delivery_date = $8,
    notes = $9,
    metadata = $10,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateFulfillmentShipment :one
UPDATE order_fulfillments
SET shipment_status = $2,
    shipped_at = CASE WHEN $2 = 'shipped' THEN NOW() ELSE shipped_at END,
    actual_delivery_date = CASE WHEN $2 = 'delivered' THEN NOW() ELSE actual_delivery_date END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateFulfillmentPickPack :one
UPDATE order_fulfillments
SET picked_at = $2,
    packed_at = $3,
    picked_by_user_id = $4,
    packed_by_user_id = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteOrderFulfillment :exec
DELETE FROM order_fulfillments
WHERE id = $1;

-- =====================================================
-- ORDER FULFILLMENT ITEMS
-- =====================================================

-- name: CreateOrderFulfillmentItem :one
INSERT INTO order_fulfillment_items (
    fulfillment_id,
    order_line_id,
    organization_id,
    quantity_fulfilled,
    batch_number,
    serial_numbers
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: ListOrderFulfillmentItems :many
SELECT ofi.*, sol.product_name, sol.product_sku
FROM order_fulfillment_items ofi
JOIN sales_order_lines_v2 sol ON ofi.order_line_id = sol.id
WHERE ofi.fulfillment_id = $1
ORDER BY ofi.created_at;

-- name: DeleteOrderFulfillmentItem :exec
DELETE FROM order_fulfillment_items
WHERE id = $1;

-- =====================================================
-- INVOICES
-- =====================================================

-- name: CreateInvoice :one
INSERT INTO invoices (
    invoice_number,
    organization_id,
    store_id,
    customer_id,
    customer_name,
    customer_email,
    customer_phone,
    customer_tax_id,
    invoice_type,
    invoice_status,
    sales_order_id,
    related_invoice_id,
    invoice_date,
    due_date,
    payment_terms,
    currency_code,
    exchange_rate,
    billing_address,
    shipping_address,
    is_recurring,
    recurrence_pattern,
    next_invoice_date,
    notes,
    internal_notes,
    reference_number,
    created_by_user_id,
    tags,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25, $26, $27, $28
) RETURNING *;

-- name: GetInvoice :one
SELECT * FROM invoices
WHERE id = $1;

-- name: GetInvoiceByNumber :one
SELECT * FROM invoices
WHERE invoice_number = $1;

-- name: ListInvoices :many
SELECT * FROM invoices
WHERE organization_id = $1
  AND ($2::int IS NULL OR store_id = $2)
  AND ($3::invoice_status IS NULL OR invoice_status = $3)
  AND ($4::date IS NULL OR invoice_date >= $4)
  AND ($5::date IS NULL OR invoice_date <= $5)
ORDER BY invoice_date DESC, created_at DESC
LIMIT $6 OFFSET $7;

-- name: ListCustomerInvoices :many
SELECT * FROM invoices
WHERE customer_id = $1
  AND organization_id = $2
ORDER BY invoice_date DESC
LIMIT $3 OFFSET $4;

-- name: ListOverdueInvoices :many
SELECT i.*, c.name as customer_name_full, c.email as customer_email_full
FROM invoices i
LEFT JOIN customers c ON i.customer_id = c.id
WHERE i.store_id = $1
  AND i.invoice_status IN ('sent', 'viewed', 'partially_paid')
  AND i.due_date < CURRENT_DATE
  AND i.balance_due > 0
ORDER BY i.due_date
LIMIT $2 OFFSET $3;

-- name: ListRecurringInvoicesDue :many
SELECT * FROM invoices
WHERE organization_id = $1
  AND is_recurring = true
  AND next_invoice_date <= $2
ORDER BY next_invoice_date;

-- name: UpdateInvoice :one
UPDATE invoices
SET customer_name = $2,
    customer_email = $3,
    customer_phone = $4,
    due_date = $5,
    payment_terms = $6,
    billing_address = $7,
    shipping_address = $8,
    notes = $9,
    internal_notes = $10,
    reference_number = $11,
    tags = $12,
    metadata = $13,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateInvoiceStatus :one
UPDATE invoices
SET invoice_status = $2,
    sent_date = CASE WHEN $2 = 'sent' THEN CURRENT_DATE ELSE sent_date END,
    paid_date = CASE WHEN $2 = 'paid' THEN CURRENT_DATE ELSE paid_date END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateInvoiceTotals :one
UPDATE invoices
SET subtotal = $2,
    discount_amount = $3,
    tax_amount = $4,
    shipping_amount = $5,
    adjustment_amount = $6,
    total_amount = $7,
    balance_due = $7 - COALESCE(paid_amount, 0) - COALESCE(credit_applied, 0),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateInvoicePaymentSummary :one
UPDATE invoices
SET paid_amount = $2,
    balance_due = total_amount - $2 - COALESCE(credit_applied, 0),
    invoice_status = CASE 
        WHEN $2 >= (total_amount - COALESCE(credit_applied, 0)) THEN 'paid'::invoice_status
        WHEN $2 > 0 THEN 'partially_paid'::invoice_status
        ELSE invoice_status
    END,
    paid_date = CASE 
        WHEN $2 >= (total_amount - COALESCE(credit_applied, 0)) THEN CURRENT_DATE
        ELSE paid_date
    END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateInvoiceReminder :one
UPDATE invoices
SET reminder_sent_count = reminder_sent_count + 1,
    last_reminder_sent_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateRecurringInvoice :one
UPDATE invoices
SET next_invoice_date = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteInvoice :exec
DELETE FROM invoices
WHERE id = $1;

-- name: GetInvoiceStats :one
SELECT 
    COUNT(*) as total_invoices,
    COUNT(*) FILTER (WHERE invoice_status = 'draft') as draft_count,
    COUNT(*) FILTER (WHERE invoice_status = 'sent') as sent_count,
    COUNT(*) FILTER (WHERE invoice_status = 'overdue') as overdue_count,
    COUNT(*) FILTER (WHERE invoice_status = 'paid') as paid_count,
    COALESCE(SUM(total_amount), 0) as total_billed,
    COALESCE(SUM(paid_amount), 0) as total_paid,
    COALESCE(SUM(balance_due), 0) as total_outstanding
FROM invoices
WHERE store_id = $1
  AND invoice_date >= $2
  AND invoice_date <= $3;

-- =====================================================
-- INVOICE LINES
-- =====================================================

-- name: CreateInvoiceLine :one
INSERT INTO invoice_lines (
    invoice_id,
    organization_id,
    line_number,
    description,
    item_type,
    product_id,
    product_variant_id,
    product_sku,
    order_line_id,
    quantity,
    unit_price,
    discount_amount,
    tax_amount,
    line_total,
    tax_category_id,
    tax_rate,
    uom_id,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18
) RETURNING *;

-- name: GetInvoiceLine :one
SELECT * FROM invoice_lines
WHERE id = $1;

-- name: ListInvoiceLines :many
SELECT il.*, p.name as product_name
FROM invoice_lines il
LEFT JOIN products p ON il.product_id = p.id
WHERE il.invoice_id = $1
ORDER BY il.line_number;

-- name: UpdateInvoiceLine :one
UPDATE invoice_lines
SET description = $2,
    quantity = $3,
    unit_price = $4,
    discount_amount = $5,
    tax_amount = $6,
    line_total = $7,
    metadata = $8,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteInvoiceLine :exec
DELETE FROM invoice_lines
WHERE id = $1;

-- name: GetInvoiceLineTotals :one
SELECT 
    COALESCE(SUM(line_total - tax_amount), 0) as subtotal,
    COALESCE(SUM(discount_amount), 0) as total_discount,
    COALESCE(SUM(tax_amount), 0) as total_tax,
    COALESCE(SUM(line_total), 0) as total
FROM invoice_lines
WHERE invoice_id = $1;

-- =====================================================
-- INVOICE PAYMENTS
-- =====================================================

-- name: CreateInvoicePayment :one
INSERT INTO invoice_payments (
    invoice_id,
    organization_id,
    payment_number,
    payment_date,
    payment_amount,
    payment_method,
    payment_gateway,
    payment_reference,
    currency_code,
    exchange_rate,
    bank_account_id,
    reconciled,
    reconciled_date,
    notes,
    received_by_user_id,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16
) RETURNING *;

-- name: GetInvoicePayment :one
SELECT * FROM invoice_payments
WHERE id = $1;

-- name: GetInvoicePaymentByNumber :one
SELECT * FROM invoice_payments
WHERE payment_number = $1;

-- name: ListInvoicePayments :many
SELECT ip.*, (u.first_name || ' ' || u.last_name) as received_by_name
FROM invoice_payments ip
LEFT JOIN users u ON ip.received_by_user_id = u.id
WHERE ip.invoice_id = $1
ORDER BY ip.payment_date DESC;

-- name: ListUnreconciledPayments :many
SELECT ip.*, i.invoice_number, c.name as customer_name
FROM invoice_payments ip
JOIN invoices i ON ip.invoice_id = i.id
LEFT JOIN customers c ON i.customer_id = c.id
WHERE ip.organization_id = $1
  AND ip.reconciled = false
ORDER BY ip.payment_date
LIMIT $2 OFFSET $3;

-- name: UpdateInvoicePayment :one
UPDATE invoice_payments
SET payment_amount = $2,
    payment_method = $3,
    payment_reference = $4,
    notes = $5,
    metadata = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ReconcilePayment :one
UPDATE invoice_payments
SET reconciled = true,
    reconciled_date = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteInvoicePayment :exec
DELETE FROM invoice_payments
WHERE id = $1;

-- name: GetTotalPaymentsForInvoice :one
SELECT COALESCE(SUM(payment_amount), 0) as total_paid
FROM invoice_payments
WHERE invoice_id = $1;

-- =====================================================
-- INVOICE STATUS HISTORY
-- =====================================================

-- name: CreateInvoiceStatusHistory :one
INSERT INTO invoice_status_history (
    invoice_id,
    organization_id,
    from_status,
    to_status,
    reason,
    notes,
    changed_by_user_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: ListInvoiceStatusHistory :many
SELECT ish.*, (u.first_name || ' ' || u.last_name) as changed_by_name
FROM invoice_status_history ish
LEFT JOIN users u ON ish.changed_by_user_id = u.id
WHERE ish.invoice_id = $1
ORDER BY ish.changed_at DESC;

-- =====================================================
-- QUOTES
-- =====================================================

-- name: CreateQuote :one
INSERT INTO quotes (
    quote_number,
    organization_id,
    store_id,
    customer_id,
    customer_name,
    customer_email,
    customer_phone,
    quote_status,
    quote_date,
    valid_until,
    payment_terms,
    delivery_terms,
    terms_and_conditions,
    notes,
    internal_notes,
    created_by_user_id,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17
) RETURNING *;

-- name: GetQuote :one
SELECT * FROM quotes
WHERE id = $1;

-- name: GetQuoteByNumber :one
SELECT * FROM quotes
WHERE quote_number = $1;

-- name: ListQuotes :many
SELECT * FROM quotes
WHERE organization_id = $1
  AND ($2::int IS NULL OR store_id = $2)
  AND ($3::quote_status IS NULL OR quote_status = $3)
  AND ($4::date IS NULL OR quote_date >= $4)
  AND ($5::date IS NULL OR quote_date <= $5)
ORDER BY quote_date DESC, created_at DESC
LIMIT $6 OFFSET $7;

-- name: ListCustomerQuotes :many
SELECT * FROM quotes
WHERE customer_id = $1
  AND organization_id = $2
ORDER BY quote_date DESC
LIMIT $3 OFFSET $4;

-- name: ListExpiredQuotes :many
SELECT * FROM quotes
WHERE organization_id = $1
  AND quote_status NOT IN ('accepted', 'declined', 'converted')
  AND valid_until < CURRENT_DATE
ORDER BY valid_until;

-- name: UpdateQuote :one
UPDATE quotes
SET customer_name = $2,
    customer_email = $3,
    customer_phone = $4,
    valid_until = $5,
    payment_terms = $6,
    delivery_terms = $7,
    terms_and_conditions = $8,
    notes = $9,
    internal_notes = $10,
    metadata = $11,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateQuoteStatus :one
UPDATE quotes
SET quote_status = $2,
    sent_date = CASE WHEN $2 = 'sent' THEN CURRENT_DATE ELSE sent_date END,
    accepted_date = CASE WHEN $2 = 'accepted' THEN CURRENT_DATE ELSE accepted_date END,
    converted_date = CASE WHEN $2 = 'converted' THEN CURRENT_DATE ELSE converted_date END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateQuoteTotals :one
UPDATE quotes
SET subtotal = $2,
    discount_amount = $3,
    tax_amount = $4,
    total_amount = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ConvertQuoteToOrder :one
UPDATE quotes
SET quote_status = 'converted',
    converted_to_order_id = $2,
    converted_date = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteQuote :exec
DELETE FROM quotes
WHERE id = $1;

-- =====================================================
-- QUOTE LINES
-- =====================================================

-- name: CreateQuoteLine :one
INSERT INTO quote_lines (
    quote_id,
    organization_id,
    line_number,
    product_id,
    product_variant_id,
    description,
    quantity,
    unit_price,
    discount_amount,
    tax_amount,
    line_total,
    uom_id,
    notes,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14
) RETURNING *;

-- name: GetQuoteLine :one
SELECT * FROM quote_lines
WHERE id = $1;

-- name: ListQuoteLines :many
SELECT ql.*, p.name as product_name, p.sku as product_sku
FROM quote_lines ql
LEFT JOIN products p ON ql.product_id = p.id
WHERE ql.quote_id = $1
ORDER BY ql.line_number;

-- name: UpdateQuoteLine :one
UPDATE quote_lines
SET description = $2,
    quantity = $3,
    unit_price = $4,
    discount_amount = $5,
    tax_amount = $6,
    line_total = $7,
    notes = $8,
    metadata = $9,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteQuoteLine :exec
DELETE FROM quote_lines
WHERE id = $1;

-- name: GetQuoteLineTotals :one
SELECT 
    COALESCE(SUM(line_total - tax_amount), 0) as subtotal,
    COALESCE(SUM(discount_amount), 0) as total_discount,
    COALESCE(SUM(tax_amount), 0) as total_tax,
    COALESCE(SUM(line_total), 0) as total
FROM quote_lines
WHERE quote_id = $1;

-- =====================================================
-- BUSINESS USE CASES - CART OPERATIONS
-- =====================================================

-- name: ConvertCartToOrder :one
-- Convert cart to sales order (business logic)
WITH cart_data AS (
    SELECT c.*, cu.name as customer_name, cu.email as customer_email, cu.phone as customer_phone
    FROM carts c
    LEFT JOIN customers cu ON c.customer_id = cu.id
    WHERE c.id = $1
),
new_order AS (
    INSERT INTO sales_orders_v2 (
        order_number, organization_id, store_id, customer_id,
        customer_name, customer_email, customer_phone,
        order_type, order_status, payment_status, fulfillment_status,
        sales_channel, source_cart_id, created_by_user_id,
        order_date, shipping_address, billing_address,
        shipping_method, payment_method, payment_gateway,
        pos_terminal_id, cashier_id,
        subtotal, discount_amount, tax_amount, shipping_amount, total_amount,
        coupon_code, discount_codes, promotional_credits,
        special_instructions, metadata
    )
    SELECT 
        $2, organization_id, store_id, customer_id,
        COALESCE(cart_data.customer_name, guest_email), COALESCE(cart_data.customer_email, guest_email), COALESCE(cart_data.customer_phone, guest_phone),
        'standard', 'pending', 'unpaid', 'unfulfilled',
        channel, id, created_by_user_id,
        NOW(), shipping_address, billing_address,
        shipping_method, payment_method, payment_gateway,
        pos_terminal_id, cashier_id,
        subtotal, discount_amount, tax_amount, shipping_amount, total_amount,
        coupon_code,
        CASE 
            WHEN discount_code IS NOT NULL AND discount_code != '' 
            THEN ARRAY[discount_code]::TEXT[]
            ELSE ARRAY[]::TEXT[]
        END,
        promotional_credits,
        notes,
        metadata
    FROM cart_data
    RETURNING *
)
UPDATE carts 
SET cart_status = 'converted',
    converted_to_order_id = (SELECT id FROM new_order),
    converted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE carts.id = $1
RETURNING id, (SELECT id FROM new_order) as converted_to_order_id;

-- name: MergeGuestCartToCustomer :exec
-- Merge guest cart into customer cart when user logs in
WITH guest_items AS (
    SELECT organization_id, product_id, product_variant_id,
           quantity, uom_id, unit_price, discount_amount, tax_amount,
           line_total, price_list_id, tax_category_id, batch_number,
           serial_number, customization_details, notes, metadata
    FROM cart_items WHERE cart_items.cart_id = $1
)
INSERT INTO cart_items (
    cart_id, organization_id, product_id, product_variant_id,
    quantity, uom_id, unit_price, discount_amount, tax_amount,
    line_total, price_list_id, tax_category_id, batch_number,
    serial_number, customization_details, notes, metadata
)
SELECT 
    $2, gi.organization_id, gi.product_id, gi.product_variant_id,
    gi.quantity, gi.uom_id, gi.unit_price, gi.discount_amount, gi.tax_amount,
    gi.line_total, gi.price_list_id, gi.tax_category_id, gi.batch_number,
    gi.serial_number, gi.customization_details, gi.notes, gi.metadata
FROM guest_items gi
ON CONFLICT (cart_id, product_id, product_variant_id, batch_number, serial_number)
DO UPDATE SET
    quantity = cart_items.quantity + EXCLUDED.quantity,
    line_total = (cart_items.quantity + EXCLUDED.quantity) * cart_items.unit_price,
    updated_at = NOW();

-- name: ApplyCouponToCart :one
UPDATE carts
SET coupon_code = $2,
    discount_amount = $3,
    total_amount = subtotal + tax_amount + shipping_amount - $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: RecalculateCartTotals :one
WITH totals AS (
    SELECT 
        COALESCE(SUM(line_total - COALESCE(tax_amount, 0)), 0) as subtotal,
        COALESCE(SUM(COALESCE(tax_amount, 0)), 0) as tax,
        COALESCE(SUM(COALESCE(discount_amount, 0)), 0) as discount
    FROM cart_items
    WHERE cart_id = $1
)
UPDATE carts c
SET subtotal = t.subtotal,
    tax_amount = t.tax,
    discount_amount = COALESCE(t.discount, 0),
    total_amount = t.subtotal + t.tax + COALESCE(c.shipping_amount, 0) - COALESCE(t.discount, 0),
    updated_at = NOW()
FROM totals t
WHERE c.id = $1
RETURNING c.*;

-- =====================================================
-- BUSINESS USE CASES - ORDER OPERATIONS
-- =====================================================

-- name: CreateOrderFromQuote :one
-- Convert accepted quote to sales order
WITH quote_data AS (
    SELECT * FROM quotes WHERE id = $1
),
new_order AS (
    INSERT INTO sales_orders_v2 (
        order_number, organization_id, store_id, customer_id,
        customer_name, customer_email, customer_phone,
        order_type, order_status, payment_status, fulfillment_status,
        sales_channel, order_date,
        shipping_address, billing_address, payment_terms,
        subtotal, discount_amount, tax_amount, total_amount,
        notes, internal_notes, created_by_user_id, metadata
    )
    SELECT 
        $2, organization_id, store_id, customer_id,
        customer_name, customer_email, customer_phone,
        'standard', 'pending', 'unpaid', 'unfulfilled',
        'quote_conversion', NOW(),
        billing_address, billing_address, payment_terms,
        subtotal, discount_amount, tax_amount, total_amount,
        notes, internal_notes, created_by_user_id, metadata
    FROM quote_data
    RETURNING *
)
UPDATE quotes
SET quote_status = 'converted',
    converted_to_order_id = (SELECT id FROM new_order),
    converted_date = NOW()
WHERE quotes.id = $1
RETURNING *;

-- name: CreateInvoiceFromOrder :one
-- Generate invoice from completed order
WITH order_data AS (
    SELECT id, organization_id, store_id, customer_id,
           customer_name, customer_email, customer_phone,
           payment_terms, subtotal, discount_amount, tax_amount, 
           shipping_amount, total_amount, billing_address, shipping_address, 
           special_instructions, created_by_user_id, metadata
    FROM sales_orders_v2 WHERE sales_orders_v2.id = $1
)
INSERT INTO invoices (
    invoice_number, organization_id, store_id, customer_id,
    customer_name, customer_email, customer_phone,
    invoice_type, invoice_status, sales_order_id,
    invoice_date, due_date, payment_terms,
    subtotal, discount_amount, tax_amount, shipping_amount, total_amount,
    billing_address, shipping_address, notes, created_by_user_id, metadata
)
SELECT 
    $2, od.organization_id, od.store_id, od.customer_id,
    od.customer_name, od.customer_email, od.customer_phone,
    'standard', 'draft', od.id,
    CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days', od.payment_terms,
    od.subtotal, od.discount_amount, od.tax_amount, od.shipping_amount, od.total_amount,
    od.billing_address, od.shipping_address, od.special_instructions, od.created_by_user_id, od.metadata
FROM order_data od
RETURNING *;

-- name: ProcessOrderPayment :one
UPDATE sales_orders_v2
SET paid_amount = paid_amount + $2,
    balance_due = total_amount - (paid_amount + $2),
    payment_status = CASE 
        WHEN (paid_amount + $2) >= total_amount THEN 'paid'::payment_status
        WHEN (paid_amount + $2) > 0 THEN 'partially_paid'::payment_status
        ELSE payment_status
    END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CheckLowStock :many
-- Find products that need reordering
SELECT 
    p.id,
    p.sku,
    p.name,
    ist.store_id,
    ist.quantity_available,
    ist.reorder_level,
    ist.reorder_quantity
FROM inventory_stock ist
JOIN products p ON ist.product_id = p.id
WHERE ist.store_id = $1
  AND ist.quantity_available <= ist.reorder_level
  AND p.is_active = true
ORDER BY (ist.reorder_level - ist.quantity_available) DESC;

-- name: GetCustomerLifetimeValue :one
SELECT 
    c.id,
    c.name,
    COUNT(DISTINCT so.id) as total_orders,
    COALESCE(SUM(so.total_amount), 0) as lifetime_value,
    COALESCE(SUM(so.total_amount) FILTER (WHERE so.order_date >= CURRENT_DATE - INTERVAL '12 months'), 0) as last_12_months,
    COALESCE(AVG(so.total_amount), 0) as average_order_value
FROM customers c
LEFT JOIN sales_orders_v2 so ON c.id = so.customer_id
WHERE c.id = $1
GROUP BY c.id, c.name;

-- name: GetTopSellingProducts :many
SELECT 
    p.id,
    p.sku,
    p.name,
    COUNT(DISTINCT sol.sales_order_id) as order_count,
    SUM(sol.quantity_ordered) as total_quantity,
    SUM(sol.line_total) as total_revenue
FROM sales_order_lines_v2 sol
JOIN products p ON sol.product_id = p.id
JOIN sales_orders_v2 so ON sol.sales_order_id = so.id
WHERE so.store_id = $1
  AND so.order_date >= $2
  AND so.order_date <= $3
  AND so.order_status NOT IN ('cancelled')
GROUP BY p.id, p.sku, p.name
ORDER BY total_revenue DESC
LIMIT $4;

-- name: GetOrderFulfillmentRate :one
SELECT 
    COUNT(*) as total_orders,
    COUNT(*) FILTER (WHERE fulfillment_status = 'fulfilled') as fulfilled_orders,
    COUNT(*) FILTER (WHERE fulfillment_status = 'partially_fulfilled') as partial_orders,
    COUNT(*) FILTER (WHERE fulfillment_status = 'unfulfilled') as unfulfilled_orders,
    CASE 
        WHEN COUNT(*) > 0 THEN 
            (COUNT(*) FILTER (WHERE fulfillment_status = 'fulfilled')::DECIMAL / COUNT(*) * 100)
        ELSE 0 
    END as fulfillment_rate
FROM sales_orders_v2
WHERE store_id = $1
  AND order_date >= $2
  AND order_date <= $3;

-- =====================================================
-- PROMOTIONS: LINE-ITEM DISCOUNT APPLICATION
-- =====================================================

-- name: ApplyDiscountToCartItem :one
-- Applies a line-item discount for product-targeted promotions
-- (e.g. buy_x_get_y, specific product percentage).
-- Updates the line_total to reflect the discount.
UPDATE cart_items
SET discount_amount = $2,
    line_total = (unit_price * quantity) - $2 + tax_amount,
    metadata = jsonb_set(COALESCE(metadata, '{}'), '{applied_promotion}', to_jsonb($3::text)),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ReopenCart :one
WITH order_to_delete AS (
    SELECT id, store_id FROM sales_orders_v2 
    WHERE source_cart_id = $1 
      AND order_status IN ('draft', 'pending')
),
lines_to_delete AS (
    SELECT id, product_id, product_variant_id, quantity_ordered
    FROM sales_order_lines_v2
    WHERE sales_order_id IN (SELECT id FROM order_to_delete)
),
deallocate_stock AS (
    UPDATE inventory_stock
    SET quantity_allocated = GREATEST(0, quantity_allocated - l.quantity_ordered),
        quantity_available = quantity_on_hand - GREATEST(0, quantity_allocated - l.quantity_ordered),
        updated_at = NOW()
    FROM lines_to_delete l
    CROSS JOIN order_to_delete o
    WHERE inventory_stock.product_id = l.product_id
      AND COALESCE(inventory_stock.product_variant_id, 0) = COALESCE(l.product_variant_id, 0)
      AND inventory_stock.store_id = o.store_id
    RETURNING inventory_stock.id
),
deleted_lines AS (
    DELETE FROM sales_order_lines_v2 
    WHERE sales_order_id IN (SELECT id FROM order_to_delete)
    RETURNING id
),
deleted_order AS (
    DELETE FROM sales_orders_v2 
    WHERE id IN (SELECT id FROM order_to_delete)
    RETURNING id
)
UPDATE carts
SET cart_status = 'active',
    converted_to_order_id = NULL,
    converted_at = NULL,
    updated_at = NOW()
WHERE carts.id = $1
  AND EXISTS (SELECT 1 FROM deleted_order)
RETURNING *;
