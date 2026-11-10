-- restaurant_views_functions.sql

-- name: ListActiveRestaurantOrdersView :many
SELECT * FROM vw_active_restaurant_orders
WHERE store_id = $1
ORDER BY ordered_at;

-- name: ListRestaurantMenuView :many
SELECT vm.menu_item_id, vm.store_id, vm.item_name, vm.short_name, vm.description,
       vm.image_url, vm.base_price, vm.cost_price, vm.preparation_time_min,
       vm.is_available, vm.is_active, vm.display_order, vm.item_type,
       vm.item_metadata, vm.category_id, vm.category_name, vm.category_code,
       vm.parent_category_id, vm.category_display_order, vm.category_image_url,
       vm.parent_category_name, vm.tax_category_id, vm.tax_rate,
       vm.tax_is_inclusive, vm.recipe_id, vm.recipe_name, vm.recipe_yield,
       vm.product_id, vm.product_variant_id, vm.product_sku,
       vm.active_modifier_count, vm.margin_percent,
       promo.promo_price,
       COALESCE(promo.promo_price, vm.base_price) AS effective_price,
       (promo.promotion_name IS NOT NULL) AS has_promotion,
       COALESCE(promo.promotion_name, '') AS promotion_name,
       promo.discount_percent, promo.promo_min_quantity
FROM vw_restaurant_menu vm
LEFT JOIN LATERAL (
    SELECT
        CASE
            WHEN pr.promotion_type = 'percentage_discount' AND vm.base_price IS NOT NULL
                THEN ROUND(vm.base_price * (1 - pr.discount_value / 100), 2)
            WHEN pr.promotion_type = 'fixed_discount' AND vm.base_price IS NOT NULL
                THEN GREATEST(0, vm.base_price - pr.discount_value)
            ELSE vm.base_price
        END AS promo_price,
        pr.name AS promotion_name,
        CASE WHEN pr.promotion_type = 'percentage_discount'
             THEN CONCAT(TRIM(TRAILING '.' FROM TRIM(TRAILING '0' FROM pr.discount_value::text)), '%')
             ELSE NULL END AS discount_percent,
        pr.min_quantity AS promo_min_quantity
    FROM (
        SELECT rp.id, rp.name, rp.promotion_type, rp.discount_value,
               rp.min_quantity, rp.valid_from
        FROM restaurant_promotions rp
        WHERE rp.store_id = vm.store_id
          AND rp.is_active = true
          AND (rp.valid_from IS NULL OR rp.valid_from <= CURRENT_TIMESTAMP)
          AND (rp.valid_to IS NULL OR rp.valid_to >= CURRENT_TIMESTAMP)
          AND (rp.usage_limit IS NULL OR rp.usage_count < rp.usage_limit)
          AND (
              COALESCE(rp.applies_to, 'all') = 'all'
              OR (rp.applies_to = 'menu_item' AND vm.menu_item_id = ANY(rp.target_menu_item_ids))
              OR (rp.applies_to = 'menu_category' AND vm.category_id = ANY(rp.target_menu_category_ids))
          )
    ) pr
    ORDER BY pr.valid_from DESC NULLS LAST, pr.id DESC
    LIMIT 1
) promo ON true
WHERE vm.store_id = $1
ORDER BY vm.category_display_order, vm.display_order;

-- name: ListRecipeBomView :many
SELECT * FROM vw_recipe_bom
WHERE recipe_id = $1
ORDER BY line_number;

-- name: GetWasteDailySummaryView :many
SELECT * FROM vw_waste_daily_summary
WHERE store_id = $1
ORDER BY waste_date DESC;
