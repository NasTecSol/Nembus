-- restaurant_views_functions.sql

-- name: ListActiveRestaurantOrdersView :many
SELECT * FROM vw_active_restaurant_orders
WHERE store_id = $1
ORDER BY ordered_at;

-- name: ListRestaurantMenuView :many
SELECT * FROM vw_restaurant_menu
WHERE store_id = $1
ORDER BY category_display_order, display_order;

-- name: ListRecipeBomView :many
SELECT * FROM vw_recipe_bom
WHERE recipe_id = $1
ORDER BY line_number;

-- name: GetWasteDailySummaryView :many
SELECT * FROM vw_waste_daily_summary
WHERE store_id = $1
ORDER BY waste_date DESC;
