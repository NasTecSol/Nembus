-- recipes.sql

-- name: GetRecipe :one
SELECT * FROM recipes
WHERE id = $1 LIMIT 1;

-- name: ListRecipes :many
SELECT * FROM recipes
WHERE organization_id = $1
ORDER BY recipe_name;

-- name: CreateRecipe :one
INSERT INTO recipes (
    organization_id, recipe_code, recipe_name, description, finished_product_id, yield_quantity, yield_uom_id, preparation_steps, preparation_time_min, cooking_time_min, is_active, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: UpdateRecipe :one
UPDATE recipes
SET
    recipe_code = $2,
    recipe_name = $3,
    description = $4,
    finished_product_id = $5,
    yield_quantity = $6,
    yield_uom_id = $7,
    preparation_steps = $8,
    preparation_time_min = $9,
    cooking_time_min = $10,
    is_active = $11,
    metadata = $12,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteRecipe :exec
DELETE FROM recipes
WHERE id = $1;
