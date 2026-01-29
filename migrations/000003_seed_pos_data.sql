-- +goose Up
-- =====================================================
-- SEED DATA - Saudi Market (Retail & Wholesale)
-- =====================================================

DO $$
DECLARE
    v_org_id INTEGER;
    v_store_riyadh_id INTEGER;
    v_store_jeddah_id INTEGER;
    v_cat_beverages_id INTEGER;
    v_cat_dairy_id INTEGER;
    v_cat_snacks_id INTEGER;
    v_uom_each_id INTEGER;
    v_uom_box_id INTEGER;
    v_tax_vat15_id INTEGER;
    v_price_retail_id INTEGER;
    v_price_promo_id INTEGER;
    v_price_wholesale_id INTEGER;
    v_prod_coke_id INTEGER;
    v_prod_milk_id INTEGER;
    v_prod_chips_id INTEGER;
BEGIN
    -- 1. Organization
    INSERT INTO organizations (name, code, legal_name, tax_id, currency_code)
    VALUES ('Nasar Trading Group', 'NASAR-GROUP', 'Nasar Trading Company LLC', '300012345600003', 'SAR')
    ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO v_org_id;

    -- 2. Stores
    INSERT INTO stores (organization_id, name, code, store_type, is_pos_enabled, timezone)
    VALUES
        (v_org_id, 'Riyadh Main Store', 'RYD-001', 'Retail', true, 'Asia/Riyadh'),
        (v_org_id, 'Jeddah Wholesale Branch', 'JED-001', 'Wholesale', true, 'Asia/Riyadh')
    ON CONFLICT (organization_id, code) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO v_store_riyadh_id; -- Note: simplified, only captures last or needs loop

    SELECT id INTO v_store_riyadh_id FROM stores WHERE code = 'RYD-001' AND organization_id = v_org_id;
    SELECT id INTO v_store_jeddah_id FROM stores WHERE code = 'JED-001' AND organization_id = v_org_id;

    -- 3. Units of Measure
    INSERT INTO units_of_measure (code, name, uom_type, decimal_places)
    VALUES
        ('EA', 'Each', 'Quantity', 0),
        ('BOX', 'Box of 24', 'Quantity', 0)
    ON CONFLICT (code) DO NOTHING;

    SELECT id INTO v_uom_each_id FROM units_of_measure WHERE code = 'EA';
    SELECT id INTO v_uom_box_id FROM units_of_measure WHERE code = 'BOX';

    -- 4. Tax Categories (Saudi VAT 15%)
    INSERT INTO tax_categories (name, code, tax_rate, is_inclusive)
    VALUES ('VAT 15%', 'VAT15', 15.00, true)
    ON CONFLICT (code) DO UPDATE SET tax_rate = EXCLUDED.tax_rate
    RETURNING id INTO v_tax_vat15_id;

    -- 5. Price Lists
    INSERT INTO price_lists (name, code, price_list_type, currency_code, is_default, is_active)
    VALUES
        ('Retail SAR Price List', 'RETAIL_SAR', 'Retail', 'SAR', true, true),
        ('Promotion SAR Price List', 'PROMO_SAR', 'Promotion', 'SAR', false, true),
        ('Wholesale SAR Price List', 'WHOLESALE_SAR', 'Wholesale', 'SAR', false, true)
    ON CONFLICT (code) DO UPDATE SET is_active = EXCLUDED.is_active
    RETURNING id INTO v_price_retail_id; -- same here, simplified

    SELECT id INTO v_price_retail_id FROM price_lists WHERE code = 'RETAIL_SAR';
    SELECT id INTO v_price_promo_id FROM price_lists WHERE code = 'PROMO_SAR';
    SELECT id INTO v_price_wholesale_id FROM price_lists WHERE code = 'WHOLESALE_SAR';

    -- 6. Product Categories
    INSERT INTO product_categories (name, code, category_level)
    VALUES
        ('Beverages', 'BEV', 1),
        ('Dairy', 'DAIRY', 1),
        ('Snacks', 'SNACKS', 1)
    ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name;

    SELECT id INTO v_cat_beverages_id FROM product_categories WHERE code = 'BEV';
    SELECT id INTO v_cat_dairy_id FROM product_categories WHERE code = 'DAIRY';
    SELECT id INTO v_cat_snacks_id FROM product_categories WHERE code = 'SNACKS';

    -- 7. Products
    -- Coca-Cola
    INSERT INTO products (organization_id, sku, name, description, category_id, base_uom_id, product_type, tax_category_id)
    VALUES (v_org_id, 'BEV-001', 'Coca-Cola 330ml', 'Classic Coca-Cola Canned drink', v_cat_beverages_id, v_uom_each_id, 'Stockable', v_tax_vat15_id)
    ON CONFLICT (organization_id, sku) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO v_prod_coke_id;

    -- Almarai Milk
    INSERT INTO products (organization_id, sku, name, description, category_id, base_uom_id, product_type, tax_category_id)
    VALUES (v_org_id, 'DAIRY-001', 'Almarai Fresh Milk 1L', 'Fresh Cow Milk', v_cat_dairy_id, v_uom_each_id, 'Stockable', v_tax_vat15_id)
    ON CONFLICT (organization_id, sku) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO v_prod_milk_id;

    -- Lays Chips
    INSERT INTO products (organization_id, sku, name, description, category_id, base_uom_id, product_type, tax_category_id)
    VALUES (v_org_id, 'SNACK-001', 'Lays Salted Chips 50g', 'Classic Salted Potato Chips', v_cat_snacks_id, v_uom_each_id, 'Stockable', v_tax_vat15_id)
    ON CONFLICT (organization_id, sku) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO v_prod_chips_id;

    -- 8. Barcodes
    INSERT INTO product_barcodes (product_id, barcode, barcode_type, is_primary)
    VALUES
        (v_prod_coke_id, '5449000000996', 'EAN13', true),
        (v_prod_milk_id, '6281000001011', 'EAN13', true),
        (v_prod_chips_id, '6281007025010', 'EAN13', true)
    ON CONFLICT (barcode) DO NOTHING;

    -- 9. Prices
    -- Coke Prices
    INSERT INTO product_prices (product_id, price_list_id, uom_id, price) VALUES
        (v_prod_coke_id, v_price_retail_id, v_uom_each_id, 2.50),
        (v_prod_coke_id, v_price_promo_id, v_uom_each_id, 2.00),
        (v_prod_coke_id, v_price_wholesale_id, v_uom_box_id, 45.00); -- Box price

    -- Milk Prices
    INSERT INTO product_prices (product_id, price_list_id, uom_id, price) VALUES
        (v_prod_milk_id, v_price_retail_id, v_uom_each_id, 6.00);

    -- Chips Prices
    INSERT INTO product_prices (product_id, price_list_id, uom_id, price) VALUES
        (v_prod_chips_id, v_price_retail_id, v_uom_each_id, 1.50),
        (v_prod_chips_id, v_price_promo_id, v_uom_each_id, 1.00);

    -- 10. Inventory Stock
    INSERT INTO inventory_stock (product_id, store_id, quantity_on_hand, quantity_available, reorder_level)
    VALUES
        (v_prod_coke_id, v_store_riyadh_id, 100, 100, 20),
        (v_prod_coke_id, v_store_jeddah_id, 500, 500, 100),
        (v_prod_milk_id, v_store_riyadh_id, 50, 50, 10),
        (v_prod_chips_id, v_store_riyadh_id, 200, 200, 30)
    ON CONFLICT DO NOTHING;

END $$;

-- +goose Down
-- In a real scenario, you might want to delete the specific seeded data
-- but often down migrations for seeds are left empty or handles with care.
DELETE FROM inventory_stock WHERE product_id IN (SELECT id FROM products WHERE sku LIKE 'BEV-%' OR sku LIKE 'DAIRY-%' OR sku LIKE 'SNACK-%');
DELETE FROM product_prices WHERE product_id IN (SELECT id FROM products WHERE sku LIKE 'BEV-%' OR sku LIKE 'DAIRY-%' OR sku LIKE 'SNACK-%');
DELETE FROM product_barcodes WHERE product_id IN (SELECT id FROM products WHERE sku LIKE 'BEV-%' OR sku LIKE 'DAIRY-%' OR sku LIKE 'SNACK-%');
DELETE FROM products WHERE sku LIKE 'BEV-%' OR sku LIKE 'DAIRY-%' OR sku LIKE 'SNACK-%';
DELETE FROM product_categories WHERE code IN ('BEV', 'DAIRY', 'SNACKS');
DELETE FROM price_lists WHERE code IN ('RETAIL_SAR', 'PROMO_SAR', 'WHOLESALE_SAR');
DELETE FROM tax_categories WHERE code = 'VAT15';
DELETE FROM stores WHERE code IN ('RYD-001', 'JED-001');
DELETE FROM organizations WHERE code = 'NASAR-GROUP';
