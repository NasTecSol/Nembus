-- =====================================================
-- POS QUERIES - Product Fetching for Point of Sale
-- =====================================================
-- This file contains optimized queries for POS operations
-- to fetch products by category, barcode, name, or SKU
-- =====================================================

-- =====================================================
-- QUERY 1: GET ALL PRODUCTS FOR POS (with stock and pricing)
-- Use this as the base product list for POS interface
-- Parameters: @store_id
-- =====================================================

CREATE OR REPLACE VIEW vw_pos_product_catalog AS
SELECT
    -- Product Core Info
    p.id as product_id,
    p.sku,
    p.name as product_name,
    p.description,
    p.product_type,

    -- Category Information
    pc.id as category_id,
    pc.name as category_name,
    pc.code as category_code,
    pc_parent.id as parent_category_id,
    pc_parent.name as parent_category_name,

    -- Brand Information
    b.id as brand_id,
    b.name as brand_name,

    -- UOM (Unit of Measure)
    uom.id as uom_id,
    uom.code as uom_code,
    uom.name as uom_name,
    uom.decimal_places,

    -- Primary Barcode
    pb.barcode,
    pb.barcode_type,

    -- Tax Information
    tc.id as tax_category_id,
    tc.name as tax_category_name,
    tc.tax_rate,
    tc.is_inclusive as tax_is_inclusive,

    -- Retail Price
    pp_retail.price as retail_price,
    pp_retail.id as retail_price_id,

    -- Promotional Price
    pp_promo.price as promo_price,
    pp_promo.id as promo_price_id,
    pp_promo.min_quantity as promo_min_quantity,
    pp_promo.valid_from as promo_valid_from,
    pp_promo.valid_to as promo_valid_to,
    (pp_promo.metadata->>'promotion_name')::text as promotion_name,
    (pp_promo.metadata->>'discount_percent')::text as discount_percent,

    -- Effective Price Logic
    (CASE
        WHEN pp_promo.id IS NOT NULL
             AND pp_promo.is_active = true
             AND pp_promo.valid_from <= CURRENT_DATE
             AND (pp_promo.valid_to IS NULL OR pp_promo.valid_to >= CURRENT_DATE)
        THEN pp_promo.price
        ELSE pp_retail.price
    END)::numeric as effective_price,

    -- Has Active Promotion Flag
    CASE
        WHEN pp_promo.id IS NOT NULL
             AND pp_promo.is_active = true
             AND pp_promo.valid_from <= CURRENT_DATE
             AND (pp_promo.valid_to IS NULL OR pp_promo.valid_to >= CURRENT_DATE)
        THEN true
        ELSE false
    END as has_active_promotion,

    -- Product Flags
    p.is_active,
    p.is_sellable,
    p.is_serialized,
    p.is_batch_managed,
    p.allow_decimal_quantity,
    p.track_inventory,

    -- Additional Info
    p.metadata as product_metadata

FROM products p
LEFT JOIN product_categories pc ON p.category_id = pc.id
LEFT JOIN product_categories pc_parent ON pc.parent_category_id = pc_parent.id
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN units_of_measure uom ON p.base_uom_id = uom.id
LEFT JOIN product_barcodes pb ON p.id = pb.product_id AND pb.is_primary = true
LEFT JOIN tax_categories tc ON p.tax_category_id = tc.id
LEFT JOIN product_prices pp_retail
    ON p.id = pp_retail.product_id
    AND pp_retail.price_list_id = (SELECT id FROM price_lists WHERE code = 'RETAIL_SAR' AND is_active = true)
    AND pp_retail.is_active = true
LEFT JOIN product_prices pp_promo
    ON p.id = pp_promo.product_id
    AND pp_promo.price_list_id = (SELECT id FROM price_lists WHERE code = 'PROMO_SAR' AND is_active = true)
    AND pp_promo.is_active = true
WHERE p.is_active = true
  AND p.is_sellable = true;

-- =====================================================
-- QUERY 2: GET PRODUCTS WITH STOCK FOR SPECIFIC STORE
-- This is the main POS query to fetch products with availability
-- Parameters: @store_id, @category_id (optional), @search_term (optional)
-- =====================================================

CREATE OR REPLACE FUNCTION fn_pos_get_products_with_stock(
    p_store_id INTEGER,
    p_category_id INTEGER DEFAULT NULL,
    p_search_term VARCHAR DEFAULT NULL,
    p_include_out_of_stock BOOLEAN DEFAULT false
)
RETURNS TABLE (
    product_id INTEGER,
    sku VARCHAR,
    product_name VARCHAR,
    description TEXT,
    category_id INTEGER,
    category_name VARCHAR,
    brand_name VARCHAR,
    barcode VARCHAR,
    uom_code VARCHAR,
    decimal_places INTEGER,
    retail_price NUMERIC,
    promo_price NUMERIC,
    effective_price NUMERIC,
    has_promotion BOOLEAN,
    promotion_name VARCHAR,
    discount_percent VARCHAR,
    promo_min_quantity NUMERIC,
    tax_rate NUMERIC,
    tax_is_inclusive BOOLEAN,
    quantity_available NUMERIC,
    quantity_on_hand NUMERIC,
    quantity_allocated NUMERIC,
    is_in_stock BOOLEAN,
    is_low_stock BOOLEAN,
    reorder_level NUMERIC,
    allow_decimal_quantity BOOLEAN,
    is_serialized BOOLEAN,
    is_batch_managed BOOLEAN,
    product_metadata JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        cat.product_id,
        cat.sku,
        cat.product_name,
        cat.description::text,
        cat.category_id,
        cat.category_name::varchar,
        cat.brand_name::varchar,
        cat.barcode::varchar,
        cat.uom_code::varchar,
        cat.decimal_places,
        cat.retail_price,
        cat.promo_price,
        cat.effective_price,
        cat.has_active_promotion as has_promotion,
        cat.promotion_name::varchar,
        cat.discount_percent::varchar,
        cat.promo_min_quantity,
        cat.tax_rate,
        cat.tax_is_inclusive,
        COALESCE(inv.quantity_available, 0)::numeric as quantity_available,
        COALESCE(inv.quantity_on_hand, 0)::numeric as quantity_on_hand,
        COALESCE(inv.quantity_allocated, 0)::numeric as quantity_allocated,
        CASE WHEN COALESCE(inv.quantity_available, 0) > 0 THEN true ELSE false END as is_in_stock,
        CASE
            WHEN COALESCE(inv.quantity_available, 0) <= COALESCE(inv.reorder_level, 0)
                 AND COALESCE(inv.quantity_available, 0) > 0
            THEN true
            ELSE false
        END as is_low_stock,
        COALESCE(inv.reorder_level, 0)::numeric as reorder_level,
        cat.allow_decimal_quantity,
        cat.is_serialized,
        cat.is_batch_managed,
        cat.product_metadata
    FROM vw_pos_product_catalog cat
    LEFT JOIN inventory_stock inv
        ON cat.product_id = inv.product_id
        AND inv.store_id = p_store_id
    WHERE
        (p_category_id IS NULL OR cat.category_id = p_category_id)
        AND (p_search_term IS NULL
             OR cat.product_name ILIKE '%' || p_search_term || '%'
             OR cat.sku ILIKE '%' || p_search_term || '%'
             OR cat.barcode ILIKE '%' || p_search_term || '%')
        AND (p_include_out_of_stock = true OR COALESCE(inv.quantity_available, 0) > 0)
    ORDER BY cat.category_name, cat.product_name;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- QUERY 3: SEARCH PRODUCT BY BARCODE (Fast Lookup)
-- Use this for barcode scanning at POS
-- Parameters: @barcode, @store_id
-- =====================================================

CREATE OR REPLACE FUNCTION fn_pos_get_product_by_barcode(
    p_barcode VARCHAR,
    p_store_id INTEGER
)
RETURNS TABLE (
    product_id INTEGER,
    sku VARCHAR,
    product_name VARCHAR,
    description TEXT,
    category_name VARCHAR,
    brand_name VARCHAR,
    barcode VARCHAR,
    uom_code VARCHAR,
    decimal_places INTEGER,
    retail_price NUMERIC,
    promo_price NUMERIC,
    effective_price NUMERIC,
    has_promotion BOOLEAN,
    promotion_name VARCHAR,
    promo_min_quantity NUMERIC,
    tax_rate NUMERIC,
    tax_is_inclusive BOOLEAN,
    quantity_available NUMERIC,
    is_in_stock BOOLEAN,
    allow_decimal_quantity BOOLEAN,
    is_serialized BOOLEAN,
    is_batch_managed BOOLEAN,
    product_metadata JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        cat.product_id,
        cat.sku,
        cat.product_name,
        cat.description::text,
        cat.category_name::varchar,
        cat.brand_name::varchar,
        cat.barcode::varchar,
        cat.uom_code::varchar,
        cat.decimal_places,
        cat.retail_price,
        cat.promo_price,
        cat.effective_price,
        cat.has_active_promotion as has_promotion,
        cat.promotion_name::varchar,
        cat.promo_min_quantity,
        cat.tax_rate,
        cat.tax_is_inclusive,
        COALESCE(inv.quantity_available, 0)::numeric as quantity_available,
        CASE WHEN COALESCE(inv.quantity_available, 0) > 0 THEN true ELSE false END as is_in_stock,
        cat.allow_decimal_quantity,
        cat.is_serialized,
        cat.is_batch_managed,
        cat.product_metadata
    FROM vw_pos_product_catalog cat
    LEFT JOIN inventory_stock inv
        ON cat.product_id = inv.product_id
        AND inv.store_id = p_store_id
    WHERE cat.barcode = p_barcode
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- QUERY 4: GET PRODUCTS BY CATEGORY
-- Parameters: @category_id, @store_id
-- =====================================================

CREATE OR REPLACE FUNCTION fn_pos_get_products_by_category(
    p_category_id INTEGER,
    p_store_id INTEGER,
    p_include_subcategories BOOLEAN DEFAULT true
)
RETURNS TABLE (
    product_id INTEGER,
    sku VARCHAR,
    product_name VARCHAR,
    category_name VARCHAR,
    brand_name VARCHAR,
    barcode VARCHAR,
    effective_price NUMERIC,
    has_promotion BOOLEAN,
    promotion_name VARCHAR,
    quantity_available NUMERIC,
    is_in_stock BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        cat.product_id,
        cat.sku,
        cat.product_name,
        cat.category_name::varchar,
        cat.brand_name::varchar,
        cat.barcode::varchar,
        cat.effective_price,
        cat.has_active_promotion as has_promotion,
        cat.promotion_name::varchar,
        COALESCE(inv.quantity_available, 0)::numeric as quantity_available,
        CASE WHEN COALESCE(inv.quantity_available, 0) > 0 THEN true ELSE false END as is_in_stock
    FROM vw_pos_product_catalog cat
    LEFT JOIN inventory_stock inv
        ON cat.product_id = inv.product_id
        AND inv.store_id = p_store_id
    WHERE
        (cat.category_id = p_category_id
         OR (p_include_subcategories = true AND cat.parent_category_id = p_category_id))
        AND COALESCE(inv.quantity_available, 0) > 0
    ORDER BY cat.product_name;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- QUERY 5: SEARCH PRODUCTS BY NAME/SKU (Fuzzy Search)
-- Parameters: @search_term, @store_id
-- =====================================================

CREATE OR REPLACE FUNCTION fn_pos_search_products(
    p_search_term VARCHAR,
    p_store_id INTEGER,
    p_limit INTEGER DEFAULT 50
)
RETURNS TABLE (
    product_id INTEGER,
    sku VARCHAR,
    product_name VARCHAR,
    category_name VARCHAR,
    brand_name VARCHAR,
    barcode VARCHAR,
    effective_price NUMERIC,
    has_promotion BOOLEAN,
    quantity_available NUMERIC,
    is_in_stock BOOLEAN,
    relevance_score INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        cat.product_id,
        cat.sku,
        cat.product_name,
        cat.category_name::varchar,
        cat.brand_name::varchar,
        cat.barcode::varchar,
        cat.effective_price,
        cat.has_active_promotion as has_promotion,
        COALESCE(inv.quantity_available, 0)::numeric as quantity_available,
        CASE WHEN COALESCE(inv.quantity_available, 0) > 0 THEN true ELSE false END as is_in_stock,
        -- Relevance scoring
        CASE
            WHEN cat.sku ILIKE p_search_term THEN 100
            WHEN cat.product_name ILIKE p_search_term THEN 90
            WHEN cat.barcode = p_search_term THEN 95
            WHEN cat.sku ILIKE p_search_term || '%' THEN 80
            WHEN cat.product_name ILIKE p_search_term || '%' THEN 70
            WHEN cat.sku ILIKE '%' || p_search_term || '%' THEN 60
            WHEN cat.product_name ILIKE '%' || p_search_term || '%' THEN 50
            ELSE 40
        END as relevance_score
    FROM vw_pos_product_catalog cat
    LEFT JOIN inventory_stock inv
        ON cat.product_id = inv.product_id
        AND inv.store_id = p_store_id
    WHERE
        (cat.product_name ILIKE '%' || p_search_term || '%'
         OR cat.sku ILIKE '%' || p_search_term || '%'
         OR cat.barcode ILIKE '%' || p_search_term || '%')
    ORDER BY relevance_score DESC, cat.product_name
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- QUERY 6: GET ALL CATEGORIES WITH PRODUCT COUNT
-- Use this for category navigation in POS
-- =====================================================

CREATE OR REPLACE VIEW vw_pos_categories AS
SELECT
    pc.id as category_id,
    pc.code as category_code,
    pc.name as category_name,
    pc.parent_category_id,
    pc_parent.name as parent_category_name,
    COUNT(DISTINCT p.id) as product_count,
    COUNT(DISTINCT CASE WHEN inv.quantity_available > 0 THEN p.id END) as in_stock_count,
    pc.metadata as category_metadata
FROM product_categories pc
LEFT JOIN product_categories pc_parent ON pc.parent_category_id = pc_parent.id
LEFT JOIN products p ON pc.id = p.category_id
    AND p.is_active = true
    AND p.is_sellable = true
LEFT JOIN inventory_stock inv ON p.id = inv.product_id
WHERE pc.is_active = true
GROUP BY pc.id, pc.code, pc.name, pc.parent_category_id, pc_parent.name, pc.metadata
HAVING COUNT(DISTINCT p.id) > 0
ORDER BY pc_parent.name NULLS FIRST, pc.name;

-- =====================================================
-- QUERY 7: GET PRODUCTS ON PROMOTION
-- Parameters: @store_id (optional)
-- =====================================================

CREATE OR REPLACE FUNCTION fn_pos_get_promoted_products(
    p_store_id INTEGER DEFAULT NULL
)
RETURNS TABLE (
    product_id INTEGER,
    sku VARCHAR,
    product_name VARCHAR,
    category_name VARCHAR,
    retail_price NUMERIC,
    promo_price NUMERIC,
    discount_amount NUMERIC,
    discount_percent VARCHAR,
    promotion_name VARCHAR,
    promo_valid_from DATE,
    promo_valid_to DATE,
    quantity_available NUMERIC,
    is_in_stock BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        cat.product_id,
        cat.sku,
        cat.product_name,
        cat.category_name::varchar,
        cat.retail_price,
        cat.promo_price,
        (cat.retail_price - cat.promo_price)::numeric as discount_amount,
        cat.discount_percent::varchar,
        cat.promotion_name::varchar,
        cat.promo_valid_from,
        cat.promo_valid_to,
        COALESCE(inv.quantity_available, 0)::numeric as quantity_available,
        CASE WHEN COALESCE(inv.quantity_available, 0) > 0 THEN true ELSE false END as is_in_stock
    FROM vw_pos_product_catalog cat
    LEFT JOIN inventory_stock inv
        ON cat.product_id = inv.product_id
        AND (p_store_id IS NULL OR inv.store_id = p_store_id)
    WHERE cat.has_active_promotion = true
      AND (p_store_id IS NULL OR COALESCE(inv.quantity_available, 0) > 0)
    ORDER BY discount_amount DESC;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- PERFORMANCE INDEXES FOR POS QUERIES
-- =====================================================

-- Index on product barcodes for fast scanning
CREATE INDEX IF NOT EXISTS idx_product_barcodes_barcode_lookup
ON product_barcodes(barcode) WHERE is_primary = true;

-- Index on product names for text search
CREATE INDEX IF NOT EXISTS idx_products_name_search
ON products USING gin(to_tsvector('english', name));

-- Index on product SKU for search
CREATE INDEX IF NOT EXISTS idx_products_sku_search
ON products(sku varchar_pattern_ops);

-- Composite index for inventory lookups
CREATE INDEX IF NOT EXISTS idx_inventory_stock_store_product_qty
ON inventory_stock(store_id, product_id, quantity_available);

-- Index for active sellable products
CREATE INDEX IF NOT EXISTS idx_products_active_sellable
ON products(is_active, is_sellable) WHERE is_active = true AND is_sellable = true;
