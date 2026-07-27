-- =====================================================
-- 1. RESTAURANT CORE MASTER DATA
-- =====================================================

-- Create Restaurant Store
INSERT INTO stores (organization_id, name, code, store_type, is_warehouse, is_pos_enabled, timezone, is_active, metadata) VALUES
(1, 'NasaR Cafe & Restaurant', 'REST-001', 'restaurant', false, true, 'Asia/Riyadh', true, 
    '{"address": "King Fahd Road, Tabuk", "city": "Tabuk", "phone": "+966-14-1234567", "manager": "NasaR", "seating_capacity": 50}');

-- =====================================================
-- 2. RESTAURANT PRODUCT CATEGORIES
-- =====================================================

-- Ingredients Category (Top Level)
INSERT INTO product_categories (name, code, description, category_level, is_active, parent_category_id) VALUES
('Restaurant Ingredients', 'REST_ING', 'Raw materials and ingredients for kitchen', 1, true, NULL);

-- Sub Categories for Ingredients
INSERT INTO product_categories (name, code, description, category_level, is_active, parent_category_id) VALUES
('Proteins', 'ING_PROTEIN', 'Meat, poultry, and fish', 2, true, (SELECT id FROM product_categories WHERE code = 'REST_ING')),
('Vegetables', 'ING_VEG', 'Fresh vegetables', 2, true, (SELECT id FROM product_categories WHERE code = 'REST_ING')),
('Dairy & Pantry', 'ING_DAIRY', 'Milk, eggs, flour, oil', 2, true, (SELECT id FROM product_categories WHERE code = 'REST_ING')),
('Spices & Seasoning', 'ING_SPICE', 'Herbs, spices, and sauces', 2, true, (SELECT id FROM product_categories WHERE code = 'REST_ING'));

-- Prepared / Finished Food Categories
INSERT INTO product_categories (name, code, description, category_level, is_active, parent_category_id) VALUES
('Menu Items (Prepared)', 'MENU_FIN', 'Finished dishes served to customers', 1, true, NULL);

-- =====================================================
-- 3. RESTAURANT PRODUCTS (INGREDIENTS)
-- =====================================================

-- Proteins
INSERT INTO products (organization_id, sku, name, description, category_id, base_uom_id, product_type, tax_category_id, is_active, is_sellable, is_purchasable, track_inventory) VALUES
(1, 'ING-CHICKEN-BREAST', 'Chicken Breast (Fresh)', 'Fresh chicken breast per kg', 
    (SELECT id FROM product_categories WHERE code = 'ING_PROTEIN'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    'raw_material', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, false, true, true),
(1, 'ING-BEEF-MINCED', 'Minced Beef', 'Fresh minced beef per kg', 
    (SELECT id FROM product_categories WHERE code = 'ING_PROTEIN'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    'raw_material', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, false, true, true);

-- Vegetables
INSERT INTO products (organization_id, sku, name, description, category_id, base_uom_id, product_type, tax_category_id, is_active, is_sellable, is_purchasable, track_inventory) VALUES
(1, 'ING-TOMATO', 'Tomato (Fresh)', 'Fresh local tomatoes per kg', 
    (SELECT id FROM product_categories WHERE code = 'ING_VEG'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    'raw_material', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, false, true, true),
(1, 'ING-ONION', 'Onion (Red)', 'Fresh red onions per kg', 
    (SELECT id FROM product_categories WHERE code = 'ING_VEG'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    'raw_material', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, false, true, true),
(1, 'ING-LETTUCE', 'Lettuce', 'Fresh romaine lettuce per kg', 
    (SELECT id FROM product_categories WHERE code = 'ING_VEG'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    'raw_material', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, false, true, true);

-- Dairy & Pantry
INSERT INTO products (organization_id, sku, name, description, category_id, base_uom_id, product_type, tax_category_id, is_active, is_sellable, is_purchasable, track_inventory) VALUES
(1, 'ING-COOKING-OIL', 'Cooking Oil (Gallon)', 'Vegetable cooking oil 5L', 
    (SELECT id FROM product_categories WHERE code = 'ING_DAIRY'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    'raw_material', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, false, true, true),
(1, 'ING-SALT', 'Cooking Salt', 'Industrial size cooking salt 10kg', 
    (SELECT id FROM product_categories WHERE code = 'ING_SPICE'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    'raw_material', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, false, true, true);

-- Coffee & Beverages Ingredients
INSERT INTO products (organization_id, sku, name, description, category_id, base_uom_id, product_type, tax_category_id, is_active, is_sellable, is_purchasable, track_inventory) VALUES
(1, 'ING-COFFEE-BEANS', 'Espresso Beans (Arabica)', 'Premium Arabica coffee beans per kg', 
    (SELECT id FROM product_categories WHERE code = 'ING_DAIRY'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    'raw_material', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, false, true, true),
(1, 'ING-MILK-REST', 'Milk (Cafe Grade)', 'Full cream milk for beverages', 
    (SELECT id FROM product_categories WHERE code = 'ING_DAIRY'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    'raw_material', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, false, true, true);

-- =====================================================
-- 4. MENU CATEGORIES (RESTAURANT SPECIFIC)
-- =====================================================

INSERT INTO menu_categories (store_id, name, code, description, category_level, display_order, is_active) VALUES
((SELECT id FROM stores WHERE code = 'REST-001'), 'Breakfast', 'CAT-BRK', 'Morning delights', 1, 1, true),
((SELECT id FROM stores WHERE code = 'REST-001'), 'Appetizers', 'CAT-APP', 'Small bites and starters', 1, 2, true),
((SELECT id FROM stores WHERE code = 'REST-001'), 'Salads', 'CAT-SAL', 'Fresh garden salads', 1, 3, true),
((SELECT id FROM stores WHERE code = 'REST-001'), 'Main Courses', 'CAT-MAIN', 'Hearty meals and grills', 1, 4, true),
((SELECT id FROM stores WHERE code = 'REST-001'), 'Desserts', 'CAT-DES', 'Sweet treats', 1, 5, true),
((SELECT id FROM stores WHERE code = 'REST-001'), 'Hot Beverages', 'CAT-BEV-H', 'Coffee and tea', 1, 6, true),
((SELECT id FROM stores WHERE code = 'REST-001'), 'Cold Beverages', 'CAT-BEV-C', 'Smoothies and soft drinks', 1, 7, true);

-- =====================================================
-- 5. RECIPES
-- =====================================================

-- Espresso Recipe
INSERT INTO recipes (organization_id, recipe_code, recipe_name, description, yield_quantity, yield_uom_id, preparation_time_min, cooking_time_min, is_active) VALUES
(1, 'REC-ESPRESSO', 'Double Espresso', 'Standard double shot espresso', 1, (SELECT id FROM units_of_measure WHERE code = 'PCS'), 1, 1, true);

INSERT INTO recipe_ingredients (recipe_id, product_id, quantity, uom_id) VALUES
((SELECT id FROM recipes WHERE recipe_code = 'REC-ESPRESSO'), (SELECT id FROM products WHERE sku = 'ING-COFFEE-BEANS'), 0.018, (SELECT id FROM units_of_measure WHERE code = 'KG'));

-- Club Sandwich Recipe
INSERT INTO recipes (organization_id, recipe_code, recipe_name, description, yield_quantity, yield_uom_id, preparation_time_min, cooking_time_min, is_active) VALUES
(1, 'REC-CLUB-SANDWICH', 'Classic Club Sandwich', 'Three layers of chicken, egg, and veg', 1, (SELECT id FROM units_of_measure WHERE code = 'PCS'), 5, 10, true);

INSERT INTO recipe_ingredients (recipe_id, product_id, quantity, uom_id) VALUES
((SELECT id FROM recipes WHERE recipe_code = 'REC-CLUB-SANDWICH'), (SELECT id FROM products WHERE sku = 'ING-CHICKEN-BREAST'), 0.150, (SELECT id FROM units_of_measure WHERE code = 'KG')),
((SELECT id FROM recipes WHERE recipe_code = 'REC-CLUB-SANDWICH'), (SELECT id FROM products WHERE sku = 'ING-TOMATO'), 0.050, (SELECT id FROM units_of_measure WHERE code = 'KG')),
((SELECT id FROM recipes WHERE recipe_code = 'REC-CLUB-SANDWICH'), (SELECT id FROM products WHERE sku = 'ING-LETTUCE'), 0.030, (SELECT id FROM units_of_measure WHERE code = 'KG'));

-- =====================================================
-- 6. MENU ITEMS
-- =====================================================

INSERT INTO menu_items (store_id, menu_category_id, recipe_id, name, short_name, description, base_price, cost_price, preparation_time_min, tax_category_id, is_available, is_active, display_order) VALUES
((SELECT id FROM stores WHERE code = 'REST-001'), 
    (SELECT id FROM menu_categories WHERE code = 'CAT-BEV-H'),
    (SELECT id FROM recipes WHERE recipe_code = 'REC-ESPRESSO'),
    'Double Espresso', 'Espresso', 'Premium Arabica double shot', 12.00, 2.50, 2, (SELECT id FROM tax_categories WHERE code = 'VAT_15'), true, true, 1),

((SELECT id FROM stores WHERE code = 'REST-001'), 
    (SELECT id FROM menu_categories WHERE code = 'CAT-BRK'),
    (SELECT id FROM recipes WHERE recipe_code = 'REC-CLUB-SANDWICH'),
    'Classic Club Sandwich', 'Club Sandwich', 'Toasted triple-decker sandwich', 35.00, 12.00, 15, (SELECT id FROM tax_categories WHERE code = 'VAT_15'), true, true, 2);

-- Simple items without detailed recipes for now
INSERT INTO menu_items (store_id, menu_category_id, name, short_name, description, base_price, cost_price, preparation_time_min, tax_category_id, is_available, is_active, display_order) VALUES
((SELECT id FROM stores WHERE code = 'REST-001'), 
    (SELECT id FROM menu_categories WHERE code = 'CAT-BEV-C'),
    'Fresh Orange Juice', 'Orange Juice', '100% freshly squeezed', 18.00, 5.00, 5, (SELECT id FROM tax_categories WHERE code = 'VAT_15'), true, true, 3),
((SELECT id FROM stores WHERE code = 'REST-001'), 
    (SELECT id FROM menu_categories WHERE code = 'CAT-MAIN'),
    'Grilled Chicken Breast', 'Grilled Chicken', 'Marinated chicken with sides', 45.00, 15.00, 20, (SELECT id FROM tax_categories WHERE code = 'VAT_15'), true, true, 4);

-- =====================================================
-- 7. RESTAURANT TABLES
-- =====================================================

INSERT INTO restaurant_tables (store_id, table_number, table_name, section, capacity, is_active) VALUES
((SELECT id FROM stores WHERE code = 'REST-001'), 'T-01', 'Window Table 1', 'Indoor', 2, true),
((SELECT id FROM stores WHERE code = 'REST-001'), 'T-02', 'Window Table 2', 'Indoor', 2, true),
((SELECT id FROM stores WHERE code = 'REST-001'), 'T-03', 'Family Table 1', 'Indoor', 6, true),
((SELECT id FROM stores WHERE code = 'REST-001'), 'T-04', 'Family Table 2', 'Indoor', 6, true),
((SELECT id FROM stores WHERE code = 'REST-001'), 'T-05', 'Booth 1', 'Indoor', 4, true),
((SELECT id FROM stores WHERE code = 'REST-001'), 'T-06', 'Booth 2', 'Indoor', 4, true),
((SELECT id FROM stores WHERE code = 'REST-001'), 'T-07', 'Terrace 1', 'Outdoor', 4, true),
((SELECT id FROM stores WHERE code = 'REST-001'), 'T-08', 'Terrace 2', 'Outdoor', 4, true),
((SELECT id FROM stores WHERE code = 'REST-001'), 'T-VIP', 'VIP Suite 1', 'VIP', 8, true);

-- =====================================================
-- 8. RESTAURANT MODULE, MENUS & PERMISSIONS
-- =====================================================

-- New Permissions
INSERT INTO permissions (name, code, description) VALUES
('View Restaurant', 'restaurant:view', 'Can view restaurant operations'),
('Manage Menu', 'restaurant:menu_manage', 'Can manage menu categories and items'),
('Manage Recipes', 'restaurant:recipe_manage', 'Can manage recipes and ingredients'),
('Manage Tables', 'restaurant:table_manage', 'Can manage restaurant floor plan and tables'),
('Process Orders', 'restaurant:process_orders', 'Can take and process restaurant orders'),
('View Kitchen Display', 'restaurant:kitchen_view', 'Can view and manage kitchen display'),
('Restaurant Settings', 'restaurant:settings', 'Can configure restaurant specific settings');

-- New Module
INSERT INTO modules (name, code, description, icon, display_order, is_active) VALUES
('Restaurant Management', 'restaurant', 'Comprehensive restaurant and cafe management', 'coffee', 16, true);

-- Menus for Restaurant Module
INSERT INTO menus (module_id, parent_menu_id, name, code, route_path, icon, display_order, is_active) VALUES
(16, NULL, 'Restaurant Operations', 'rest_ops', '/restaurant/ops', 'clipboard', 1, true),
(16, NULL, 'Kitchen & Service', 'rest_kitchen', '/restaurant/kitchen', 'chef-hat', 2, true),
(16, NULL, 'Menu & Recipes', 'rest_catalog', '/restaurant/catalog', 'book-open', 3, true),
(16, NULL, 'Restaurant Reports', 'rest_reports', '/restaurant/reports', 'bar-chart', 4, true);

-- Submenus for Restaurant Module
INSERT INTO submenus (menu_id, parent_submenu_id, name, code, route_path, icon, display_order, is_active) VALUES
-- Operations Submenus
(32, NULL, 'Order Entry', 'rest_order_entry', '/restaurant/ops/orders', 'plus-circle', 1, true),
(32, NULL, 'Active Orders', 'rest_active_orders', '/restaurant/ops/active', 'clock', 2, true),
(32, NULL, 'Table Status', 'rest_table_status', '/restaurant/ops/tables', 'grid', 3, true),

-- Kitchen Submenus
(33, NULL, 'Kitchen Display (KDS)', 'rest_kds', '/restaurant/kitchen/kds', 'monitor', 1, true),
(33, NULL, 'Preparation Queue', 'rest_prep_queue', '/restaurant/kitchen/prep', 'activity', 2, true),

-- Menu & Recipe Submenus
(34, NULL, 'Menu Categories', 'rest_menu_cats', '/restaurant/catalog/categories', 'list', 1, true),
(34, NULL, 'Menu Items', 'rest_menu_items', '/restaurant/catalog/items', 'coffee', 2, true),
(34, NULL, 'Recipes', 'rest_recipes', '/restaurant/catalog/recipes', 'book', 3, true),

-- Reports Submenus
(35, NULL, 'Sales by Category', 'rest_sales_cat', '/restaurant/reports/sales-category', 'pie-chart', 1, true),
(35, NULL, 'Item Popularity', 'rest_item_rank', '/restaurant/reports/popularity', 'trending-up', 2, true),
(35, NULL, 'Wastage Report', 'rest_wastage', '/reports/restaurant/wastage', 'trash-2', 3, true);

-- Mapping Permissions to Module
INSERT INTO module_permissions (module_id, permission_id) VALUES
(16, (SELECT id FROM permissions WHERE code = 'restaurant:view')),
(16, (SELECT id FROM permissions WHERE code = 'restaurant:menu_manage')),
(16, (SELECT id FROM permissions WHERE code = 'restaurant:recipe_manage')),
(16, (SELECT id FROM permissions WHERE code = 'restaurant:table_manage')),
(16, (SELECT id FROM permissions WHERE code = 'restaurant:process_orders')),
(16, (SELECT id FROM permissions WHERE code = 'restaurant:kitchen_view')),
(16, (SELECT id FROM permissions WHERE code = 'restaurant:settings'));

-- Mapping Permissions to Menus
INSERT INTO menu_permissions (menu_id, permission_id) VALUES
((SELECT id FROM menus WHERE code = 'rest_ops'), (SELECT id FROM permissions WHERE code = 'restaurant:process_orders')),
((SELECT id FROM menus WHERE code = 'rest_kitchen'), (SELECT id FROM permissions WHERE code = 'restaurant:kitchen_view')),
((SELECT id FROM menus WHERE code = 'rest_catalog'), (SELECT id FROM permissions WHERE code = 'restaurant:menu_manage')),
((SELECT id FROM menus WHERE code = 'rest_catalog'), (SELECT id FROM permissions WHERE code = 'restaurant:recipe_manage')),
((SELECT id FROM menus WHERE code = 'rest_reports'), (SELECT id FROM permissions WHERE code = 'restaurant:view'));

-- Mapping Permissions to Submenus
INSERT INTO submenu_permissions (submenu_id, permission_id) VALUES
((SELECT id FROM submenus WHERE code = 'rest_order_entry'), (SELECT id FROM permissions WHERE code = 'restaurant:process_orders')),
((SELECT id FROM submenus WHERE code = 'rest_active_orders'), (SELECT id FROM permissions WHERE code = 'restaurant:view')),
((SELECT id FROM submenus WHERE code = 'rest_table_status'), (SELECT id FROM permissions WHERE code = 'restaurant:table_manage')),
((SELECT id FROM submenus WHERE code = 'rest_kds'), (SELECT id FROM permissions WHERE code = 'restaurant:kitchen_view')),
((SELECT id FROM submenus WHERE code = 'rest_prep_queue'), (SELECT id FROM permissions WHERE code = 'restaurant:kitchen_view')),
((SELECT id FROM submenus WHERE code = 'rest_menu_cats'), (SELECT id FROM permissions WHERE code = 'restaurant:menu_manage')),
((SELECT id FROM submenus WHERE code = 'rest_menu_items'), (SELECT id FROM permissions WHERE code = 'restaurant:menu_manage')),
((SELECT id FROM submenus WHERE code = 'rest_recipes'), (SELECT id FROM permissions WHERE code = 'restaurant:recipe_manage'));

-- =====================================================
-- 9. PRODUCT BARCODES & PRICES FOR RESTAURANT
-- =====================================================

-- Prices for ingredients (Purchasable)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active) VALUES
((SELECT id FROM products WHERE sku = 'ING-CHICKEN-BREAST'), (SELECT id FROM price_lists WHERE code = 'RETAIL_SAR'), (SELECT id FROM units_of_measure WHERE code = 'KG'), 25.00, 1, true),
((SELECT id FROM products WHERE sku = 'ING-BEEF-MINCED'), (SELECT id FROM price_lists WHERE code = 'RETAIL_SAR'), (SELECT id FROM units_of_measure WHERE code = 'KG'), 45.00, 1, true),
((SELECT id FROM products WHERE sku = 'ING-COFFEE-BEANS'), (SELECT id FROM price_lists WHERE code = 'RETAIL_SAR'), (SELECT id FROM units_of_measure WHERE code = 'KG'), 120.00, 1, true);

-- UOM Conversions (e.g., Coffee Beans KG to Grams for recipes)
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default) VALUES
((SELECT id FROM products WHERE sku = 'ING-COFFEE-BEANS'), (SELECT id FROM units_of_measure WHERE code = 'KG'), (SELECT id FROM units_of_measure WHERE code = 'GM'), 1000.00, true);
