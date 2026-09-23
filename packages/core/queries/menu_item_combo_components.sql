-- menu_item_combo_components.sql

-- name: CreateMenuItemComboComponent :one
INSERT INTO menu_item_combo_components (
    parent_menu_item_id, component_menu_item_id, group_name, min_selection, max_selection, price_adjustment, display_order
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: DeleteMenuItemComboComponentsByParentID :exec
DELETE FROM menu_item_combo_components
WHERE parent_menu_item_id = $1;

-- name: ListMenuItemComboComponentsByParentID :many
SELECT 
    cc.id,
    cc.parent_menu_item_id,
    cc.component_menu_item_id,
    cc.group_name,
    cc.min_selection,
    cc.max_selection,
    cc.price_adjustment,
    cc.display_order,
    cc.created_at,
    mi.name AS component_name,
    mi.base_price AS component_base_price,
    mi.image_url AS component_image_url,
    mi.is_available AS component_is_available
FROM menu_item_combo_components cc
JOIN menu_items mi ON cc.component_menu_item_id = mi.id
WHERE cc.parent_menu_item_id = $1
ORDER BY cc.group_name, cc.display_order;
