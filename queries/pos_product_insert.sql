-- name: CreateProductFull :one
WITH inserted_product AS (
    INSERT INTO products (
        organization_id, sku, name, description, category_id,
        brand_id, base_uom_id, product_type, tax_category_id,
        is_serialized, is_batch_managed, is_active, is_sellable,
        is_purchasable, allow_decimal_quantity, track_inventory, metadata
    ) VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
    ) RETURNING id
),
inserted_barcodes AS (
    INSERT INTO product_barcodes (product_id, barcode, barcode_type, is_primary, metadata)
    SELECT (SELECT id FROM inserted_product),
           (x->>'barcode')::TEXT,
           (x->>'barcode_type')::TEXT,
           COALESCE((x->>'is_primary')::BOOLEAN, false),
           COALESCE((x->>'metadata')::JSONB, '{}'::JSONB)
    FROM jsonb_array_elements(sqlc.arg(barcodes)::jsonb) AS x
),
inserted_prices AS (
    INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, max_quantity, valid_from, valid_to, is_active, metadata)
    SELECT (SELECT id FROM inserted_product),
           (x->>'price_list_id')::INTEGER,
           (x->>'uom_id')::INTEGER,
           (x->>'price')::NUMERIC,
           COALESCE((x->>'min_quantity')::NUMERIC, 1),
           (x->>'max_quantity')::NUMERIC,
           (x->>'valid_from')::DATE,
           (x->>'valid_to')::DATE,
           COALESCE((x->>'is_active')::BOOLEAN, true),
           COALESCE((x->>'metadata')::JSONB, '{}'::JSONB)
    FROM jsonb_array_elements(sqlc.arg(prices)::jsonb) AS x
),
inserted_conversions AS (
    INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata)
    SELECT (SELECT id FROM inserted_product),
           (x->>'from_uom_id')::INTEGER,
           (x->>'to_uom_id')::INTEGER,
           (x->>'conversion_factor')::NUMERIC,
           COALESCE((x->>'is_default')::BOOLEAN, false),
           COALESCE((x->>'metadata')::JSONB, '{}'::JSONB)
    FROM jsonb_array_elements(sqlc.arg(conversions)::jsonb) AS x
)
SELECT id FROM inserted_product;
