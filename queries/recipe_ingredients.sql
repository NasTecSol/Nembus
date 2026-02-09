-- recipe_ingredients.sql

-- name: GetRecipeIngredient :one
SELECT * FROM recipe_ingredients
WHERE id = $1 LIMIT 1;

-- name: ListRecipeIngredients :many
SELECT * FROM recipe_ingredients
WHERE recipe_id = $1
ORDER BY line_number;

-- name: CreateRecipeIngredient :one
INSERT INTO recipe_ingredients (
    recipe_id, product_id, product_variant_id, quantity, uom_id, is_optional, is_byproduct, line_number, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpdateRecipeIngredient :one
UPDATE recipe_ingredients
SET
    product_id = $2,
    product_variant_id = $3,
    quantity = $4,
    uom_id = $5,
    is_optional = $6,
    is_byproduct = $7,
    line_number = $8,
    metadata = $9
WHERE id = $1
RETURNING *;

-- name: DeleteRecipeIngredient :exec
DELETE FROM recipe_ingredients
WHERE id = $1;
