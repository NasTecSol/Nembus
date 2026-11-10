-- name: CreateUISetting :one
INSERT INTO ui_settings (
    submenu_id,
    setting_key,
    setting_value,
    description,
    metadata
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetUISetting :one
SELECT * FROM ui_settings
WHERE id = $1;

-- name: GetUISettingByKey :one
SELECT * FROM ui_settings
WHERE submenu_id = $1 AND setting_key = $2;

-- name: ListUISettings :many
SELECT * FROM ui_settings
ORDER BY submenu_id, setting_key;

-- name: ListUISettingsBySubmenu :many
SELECT * FROM ui_settings
WHERE submenu_id = $1
ORDER BY setting_key;

-- name: UpdateUISetting :one
UPDATE ui_settings
SET 
    setting_value = $2,
    description = $3,
    metadata = $4
WHERE id = $1
RETURNING *;

-- name: DeleteUISetting :exec
DELETE FROM ui_settings
WHERE id = $1;

-- name: CreateRoleUICustomization :one
INSERT INTO role_ui_customizations (
    role_id,
    submenu_id,
    customization_data,
    metadata
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetRoleUICustomization :one
SELECT * FROM role_ui_customizations
WHERE id = $1;

-- name: GetRoleUICustomizationByRoleAndSubmenu :one
SELECT * FROM role_ui_customizations
WHERE role_id = $1 AND submenu_id = $2;

-- name: ListRoleUICustomizations :many
SELECT * FROM role_ui_customizations
ORDER BY role_id, submenu_id;

-- name: ListRoleUICustomizationsByRole :many
SELECT * FROM role_ui_customizations
WHERE role_id = $1
ORDER BY submenu_id;

-- name: ListRoleUICustomizationsBySubmenu :many
SELECT * FROM role_ui_customizations
WHERE submenu_id = $1
ORDER BY role_id;

-- name: UpdateRoleUICustomization :one
UPDATE role_ui_customizations
SET 
    customization_data = $2,
    metadata = $3
WHERE id = $1
RETURNING *;

-- name: DeleteRoleUICustomization :exec
DELETE FROM role_ui_customizations
WHERE id = $1;
