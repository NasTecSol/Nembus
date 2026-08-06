

-- =====================================================
-- 1. CORE MASTER DATA
-- =====================================================

-- Create Default Organization
INSERT INTO organizations (name, code, legal_name, tax_id, currency_code, fiscal_year_variant, is_active, metadata) VALUES
('Qitaf Group', 'ORG001', 'Qitaf Group LLC', 'TAX123456789', 'SAR', 'CALENDAR', true, '{"industry": "retail", "established": "2020"}')

select * from organizations
-- =====================================================
-- CREATE DEMO STORES
-- =====================================================
select * from organisation;
-- Main Store (Flagship Retail)
INSERT INTO stores (organization_id, name, code, store_type, is_warehouse, is_pos_enabled, timezone, is_active, metadata) VALUES
(1, 'Qitaf al Ayela', 'RYD-001', 'retail', false, true, 'Asia/Riyadh', true, 
    '{"address": "Saudia/Tabuk", "city": "Tabuk", "phone": "+966-11-1234567", "manager": "NasaR"}'),

(1, 'Qitaf al Qadsyia', 'JED-001', 'retail', false, true, 'Asia/Riyadh', true,
    '{"address": "Saudia/Tabuk", "city": "Tabuk", "phone": "+966-11-1234567", "manager": "NasaR"}'),

(1, 'Qitaf al Tamaouz', 'DMM-001', 'retail', false, true, 'Asia/Riyadh', true,
    '{"address": "King Saud Road, Dammam", "city": "Dammam", "phone": "+966-13-3456789", "manager": "Omar Al-Otaibi"}'),

(1, 'Qitaf Warehouse', 'WH-RYD-001', 'warehouse', true, true, 'Asia/Riyadh', true,
    '{"address": "Industrial Area, Tabuk", "city": "Tabuk", "phone": "+966-11-9876543", "manager": "Hassan Al-Mutairi"}'),

(1, 'Wholesale Center Riyadh', 'WHSL-RYD-001', 'wholesale', false, true, 'Asia/Riyadh', true,
    '{"address": "Industrial Area, Tabuk", "city": "Tabuk", "phone": "+966-11-9876543", "manager": "Hassan Al-Mutairi"}');

select * from stores;

-- =====================================================
-- SAUDI ARABIA MARKET - PRODUCT SEED DATA
-- For Retail/Wholesale Chain Stores
-- =====================================================

-- Assumptions:
-- organization_id = 1 (adjust as needed)
-- Base UOM IDs: 1=PCS, 2=KG, 3=LTR, 4=BOX, 5=CTN (will be created first)

-- =====================================================
-- UNITS OF MEASURE
-- =====================================================
INSERT INTO units_of_measure (code, name, uom_type, decimal_places, is_active) VALUES
('PCS', 'Pieces', 'quantity', 0, true),
('KG', 'Kilogram', 'weight', 3, true),
('LTR', 'Liter', 'volume', 3, true),
('BOX', 'Box', 'packaging', 0, true),
('CTN', 'Carton', 'packaging', 0, true),
('PKT', 'Packet', 'packaging', 0, true),
('BTL', 'Bottle', 'packaging', 0, true),
('CAN', 'Can', 'packaging', 0, true),
('BAG', 'Bag', 'packaging', 0, true),
('GM', 'Gram', 'weight', 0, true),
('ML', 'Milliliter', 'volume', 0, true),
('DZN', 'Dozen', 'quantity', 0, true)
ON CONFLICT (code) DO NOTHING;

-- =====================================================
-- BRANDS (Saudi & International brands popular in KSA)
-- =====================================================
INSERT INTO brands (name, code, is_active, metadata) VALUES
-- Food & Beverages
('Almarai', 'ALMARAI', true, '{"country": "Saudi Arabia", "category": "Dairy"}'),
('Nadec', 'NADEC', true, '{"country": "Saudi Arabia", "category": "Dairy"}'),
('Al-Safi Danone', 'ALSAFI', true, '{"country": "Saudi Arabia", "category": "Dairy"}'),
('Americana', 'AMERICANA', true, '{"country": "Kuwait", "category": "Food"}'),
('Nestlé', 'NESTLE', true, '{"country": "International", "category": "Food & Beverages"}'),
('Coca-Cola', 'COCACOLA', true, '{"country": "International", "category": "Beverages"}'),
('PepsiCo', 'PEPSI', true, '{"country": "International", "category": "Beverages"}'),
('Al-Watania', 'WATANIA', true, '{"country": "Saudi Arabia", "category": "Poultry"}'),
('Sunbulah', 'SUNBULAH', true, '{"country": "Saudi Arabia", "category": "Frozen Foods"}'),
('California Garden', 'CALGARDEN', true, '{"country": "Saudi Arabia", "category": "Canned Foods"}'),
('Rabea', 'RABEA', true, '{"country": "Saudi Arabia", "category": "Tea"}'),
('Lipton', 'LIPTON', true, '{"country": "International", "category": "Tea"}'),
('Nescafe', 'NESCAFE', true, '{"country": "International", "category": "Coffee"}'),
('Lavazza', 'LAVAZZA', true, '{"country": "Italy", "category": "Coffee"}'),

-- Personal Care & Household
('Dettol', 'DETTOL', true, '{"country": "International", "category": "Personal Care"}'),
('Tide', 'TIDE', true, '{"country": "International", "category": "Household"}'),
('Ariel', 'ARIEL', true, '{"country": "International", "category": "Household"}'),
('Persil', 'PERSIL', true, '{"country": "International", "category": "Household"}'),
('Dove', 'DOVE', true, '{"country": "International", "category": "Personal Care"}'),
('Lux', 'LUX', true, '{"country": "International", "category": "Personal Care"}'),
('Palmolive', 'PALMOLIVE', true, '{"country": "International", "category": "Personal Care"}'),
('Finish', 'FINISH', true, '{"country": "International", "category": "Household"}'),

-- Electronics & Home Appliances
('Samsung', 'SAMSUNG', true, '{"country": "South Korea", "category": "Electronics"}'),
('LG', 'LG', true, '{"country": "South Korea", "category": "Electronics"}'),
('Panasonic', 'PANASONIC', true, '{"country": "Japan", "category": "Electronics"}'),
('Philips', 'PHILIPS', true, '{"country": "Netherlands", "category": "Electronics"}')
ON CONFLICT (code) DO NOTHING;

-- =====================================================
-- PRODUCT CATEGORIES (Hierarchical)
-- =====================================================

-- Level 1: Main Categories
INSERT INTO product_categories (name, code, description, category_level, is_active, parent_category_id) VALUES
('Food & Groceries', 'FOOD', 'Food and grocery items', 1, true, NULL),
('Beverages', 'BEVERAGES', 'All types of beverages', 1, true, NULL),
('Dairy & Eggs', 'DAIRY', 'Dairy products and eggs', 1, true, NULL),
('Personal Care', 'PERSONAL_CARE', 'Personal hygiene and care products', 1, true, NULL),
('Household & Cleaning', 'HOUSEHOLD', 'Household and cleaning supplies', 1, true, NULL),
('Electronics', 'ELECTRONICS', 'Electronic devices and appliances', 1, true, NULL),
('Bakery', 'BAKERY', 'Bread and bakery items', 1, true, NULL),
('Frozen Foods', 'FROZEN', 'Frozen food products', 1, true, NULL)
ON CONFLICT (code) DO NOTHING;

-- Level 2: Sub Categories (Food & Groceries)
INSERT INTO product_categories (name, code, description, category_level, is_active, parent_category_id) VALUES
('Rice & Grains', 'RICE_GRAINS', 'Rice, wheat, and grain products', 2, true, (SELECT id FROM product_categories WHERE code = 'FOOD')),
('Cooking Oil', 'COOKING_OIL', 'Cooking and frying oils', 2, true, (SELECT id FROM product_categories WHERE code = 'FOOD')),
('Canned Foods', 'CANNED', 'Canned vegetables, fruits, and ready meals', 2, true, (SELECT id FROM product_categories WHERE code = 'FOOD')),
('Spices & Seasonings', 'SPICES', 'Spices and cooking seasonings', 2, true, (SELECT id FROM product_categories WHERE code = 'FOOD')),
('Pasta & Noodles', 'PASTA', 'Pasta, noodles, and macaroni', 2, true, (SELECT id FROM product_categories WHERE code = 'FOOD')),
('Sugar & Salt', 'SUGAR_SALT', 'Sugar, salt, and sweeteners', 2, true, (SELECT id FROM product_categories WHERE code = 'FOOD'))
ON CONFLICT (code) DO NOTHING;

-- Level 2: Sub Categories (Beverages)
INSERT INTO product_categories (name, code, description, category_level, is_active, parent_category_id) VALUES
('Soft Drinks', 'SOFT_DRINKS', 'Carbonated and non-carbonated soft drinks', 2, true, (SELECT id FROM product_categories WHERE code = 'BEVERAGES')),
('Juices', 'JUICES', 'Fresh and packaged juices', 2, true, (SELECT id FROM product_categories WHERE code = 'BEVERAGES')),
('Water', 'WATER', 'Bottled water and mineral water', 2, true, (SELECT id FROM product_categories WHERE code = 'BEVERAGES')),
('Tea & Coffee', 'TEA_COFFEE', 'Tea, coffee, and hot beverages', 2, true, (SELECT id FROM product_categories WHERE code = 'BEVERAGES')),
('Energy Drinks', 'ENERGY_DRINKS', 'Energy and sports drinks', 2, true, (SELECT id FROM product_categories WHERE code = 'BEVERAGES'))
ON CONFLICT (code) DO NOTHING;

-- Level 2: Sub Categories (Dairy)
INSERT INTO product_categories (name, code, description, category_level, is_active, parent_category_id) VALUES
('Fresh Milk', 'FRESH_MILK', 'Fresh and UHT milk', 2, true, (SELECT id FROM product_categories WHERE code = 'DAIRY')),
('Yogurt & Laban', 'YOGURT', 'Yogurt, laban, and cultured dairy', 2, true, (SELECT id FROM product_categories WHERE code = 'DAIRY')),
('Cheese', 'CHEESE', 'All types of cheese', 2, true, (SELECT id FROM product_categories WHERE code = 'DAIRY')),
('Butter & Ghee', 'BUTTER_GHEE', 'Butter, ghee, and spreads', 2, true, (SELECT id FROM product_categories WHERE code = 'DAIRY')),
('Eggs', 'EGGS', 'Fresh eggs', 2, true, (SELECT id FROM product_categories WHERE code = 'DAIRY'))
ON CONFLICT (code) DO NOTHING;

-- Level 2: Sub Categories (Personal Care)
INSERT INTO product_categories (name, code, description, category_level, is_active, parent_category_id) VALUES
('Bath & Shower', 'BATH_SHOWER', 'Soaps, shower gels, and body wash', 2, true, (SELECT id FROM product_categories WHERE code = 'PERSONAL_CARE')),
('Hair Care', 'HAIR_CARE', 'Shampoo, conditioner, and hair products', 2, true, (SELECT id FROM product_categories WHERE code = 'PERSONAL_CARE')),
('Oral Care', 'ORAL_CARE', 'Toothpaste, toothbrush, and mouthwash', 2, true, (SELECT id FROM product_categories WHERE code = 'PERSONAL_CARE')),
('Skin Care', 'SKIN_CARE', 'Lotions, creams, and skin care', 2, true, (SELECT id FROM product_categories WHERE code = 'PERSONAL_CARE')),
('Deodorants', 'DEODORANTS', 'Deodorants and antiperspirants', 2, true, (SELECT id FROM product_categories WHERE code = 'PERSONAL_CARE'))
ON CONFLICT (code) DO NOTHING;

-- Level 2: Sub Categories (Household)
INSERT INTO product_categories (name, code, description, category_level, is_active, parent_category_id) VALUES
('Laundry Detergents', 'LAUNDRY', 'Washing powders and liquid detergents', 2, true, (SELECT id FROM product_categories WHERE code = 'HOUSEHOLD')),
('Dishwashing', 'DISHWASHING', 'Dishwashing liquid and tablets', 2, true, (SELECT id FROM product_categories WHERE code = 'HOUSEHOLD')),
('Surface Cleaners', 'SURFACE_CLEAN', 'Floor and surface cleaning products', 2, true, (SELECT id FROM product_categories WHERE code = 'HOUSEHOLD')),
('Paper Products', 'PAPER_PRODUCTS', 'Tissues, toilet paper, and paper towels', 2, true, (SELECT id FROM product_categories WHERE code = 'HOUSEHOLD')),
('Air Fresheners', 'AIR_FRESH', 'Air fresheners and deodorizers', 2, true, (SELECT id FROM product_categories WHERE code = 'HOUSEHOLD'))
ON CONFLICT (code) DO NOTHING;

-- Level 2: Sub Categories (Frozen)
INSERT INTO product_categories (name, code, description, category_level, is_active, parent_category_id) VALUES
('Frozen Vegetables', 'FROZEN_VEG', 'Frozen vegetables and mixed vegetables', 2, true, (SELECT id FROM product_categories WHERE code = 'FROZEN')),
('Frozen Chicken', 'FROZEN_CHICKEN', 'Frozen chicken and poultry', 2, true, (SELECT id FROM product_categories WHERE code = 'FROZEN')),
('Frozen Snacks', 'FROZEN_SNACKS', 'Frozen snacks and appetizers', 2, true, (SELECT id FROM product_categories WHERE code = 'FROZEN')),
('Ice Cream', 'ICE_CREAM', 'Ice cream and frozen desserts', 2, true, (SELECT id FROM product_categories WHERE code = 'FROZEN'))
ON CONFLICT (code) DO NOTHING;

-- =====================================================
-- TAX CATEGORIES (Saudi VAT is 15%)
-- =====================================================
INSERT INTO tax_categories (name, code, tax_rate, is_inclusive, is_active) VALUES
('Standard VAT', 'VAT_15', 15.00, false, true),
('Zero Rated', 'VAT_0', 0.00, false, true),
('Exempt', 'VAT_EXEMPT', 0.00, false, true)
ON CONFLICT (code) DO NOTHING;

-- =====================================================
-- PRICE LISTS
-- =====================================================
INSERT INTO price_lists (name, code, price_list_type, currency_code, valid_from, is_default, is_active) VALUES
('Retail Price List', 'RETAIL', 'retail', 'SAR', '2024-01-01', true, true),
('Wholesale Price List', 'WHOLESALE', 'wholesale', 'SAR', '2024-01-01', false, true),
('Promotional Price List', 'PROMO', 'promotional', 'SAR', '2024-01-01', false, true)
ON CONFLICT (code) DO NOTHING;

-- =====================================================
-- PRODUCTS - DAIRY & EGGS
-- =====================================================

-- Fresh Milk
INSERT INTO products (organization_id, sku, name, description, category_id, brand_id, base_uom_id, product_type, tax_category_id, is_active, is_sellable, is_purchasable, track_inventory) VALUES
(1, 'ALMARAI-MILK-FW-1L', 'Almarai Fresh Milk Full Fat 1L', 'Full cream fresh milk', 
    (SELECT id FROM product_categories WHERE code = 'FRESH_MILK'),
    (SELECT id FROM brands WHERE code = 'ALMARAI'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'ALMARAI-MILK-LF-1L', 'Almarai Low Fat Milk 1L', 'Low fat fresh milk', 
    (SELECT id FROM product_categories WHERE code = 'FRESH_MILK'),
    (SELECT id FROM brands WHERE code = 'ALMARAI'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'NADEC-MILK-FW-2L', 'Nadec Full Cream Milk 2L', 'Full cream UHT milk 2 liters', 
    (SELECT id FROM product_categories WHERE code = 'FRESH_MILK'),
    (SELECT id FROM brands WHERE code = 'NADEC'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

-- Yogurt & Laban
(1, 'ALMARAI-LABAN-1L', 'Almarai Laban Full Fat 1L', 'Traditional laban drink', 
    (SELECT id FROM product_categories WHERE code = 'YOGURT'),
    (SELECT id FROM brands WHERE code = 'ALMARAI'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'ALSAFI-YOGURT-170G', 'Al-Safi Greek Yogurt 170g', 'Greek style yogurt', 
    (SELECT id FROM product_categories WHERE code = 'YOGURT'),
    (SELECT id FROM brands WHERE code = 'ALSAFI'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

-- Cheese
(1, 'ALMARAI-CHEESE-SLICE-200G', 'Almarai Cheese Slices 200g', 'Processed cheese slices', 
    (SELECT id FROM product_categories WHERE code = 'CHEESE'),
    (SELECT id FROM brands WHERE code = 'ALMARAI'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'ALMARAI-FETA-CHEESE-400G', 'Almarai Feta Cheese 400g', 'White feta cheese', 
    (SELECT id FROM product_categories WHERE code = 'CHEESE'),
    (SELECT id FROM brands WHERE code = 'ALMARAI'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

-- Eggs
(1, 'EGGS-WHITE-30PCS', 'Fresh White Eggs 30 Pieces', 'Medium size white eggs tray', 
    (SELECT id FROM product_categories WHERE code = 'EGGS'),
    NULL,
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true);

-- =====================================================
-- PRODUCTS - BEVERAGES
-- =====================================================

-- Soft Drinks
INSERT INTO products (organization_id, sku, name, description, category_id, brand_id, base_uom_id, product_type, tax_category_id, is_active, is_sellable, is_purchasable, track_inventory) VALUES
(1, 'COCA-COLA-330ML', 'Coca-Cola 330ml Can', 'Coca-Cola regular can', 
    (SELECT id FROM product_categories WHERE code = 'SOFT_DRINKS'),
    (SELECT id FROM brands WHERE code = 'COCACOLA'),
    (SELECT id FROM units_of_measure WHERE code = 'CAN'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'PEPSI-330ML', 'Pepsi 330ml Can', 'Pepsi regular can', 
    (SELECT id FROM product_categories WHERE code = 'SOFT_DRINKS'),
    (SELECT id FROM brands WHERE code = 'PEPSI'),
    (SELECT id FROM units_of_measure WHERE code = 'CAN'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'COCA-COLA-2L', 'Coca-Cola 2L Bottle', 'Coca-Cola 2 liter bottle', 
    (SELECT id FROM product_categories WHERE code = 'SOFT_DRINKS'),
    (SELECT id FROM brands WHERE code = 'COCACOLA'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

-- Water
(1, 'WATER-600ML', 'Bottled Water 600ml', 'Purified drinking water', 
    (SELECT id FROM product_categories WHERE code = 'WATER'),
    NULL,
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'WATER-1.5L', 'Bottled Water 1.5L', 'Purified drinking water 1.5 liters', 
    (SELECT id FROM product_categories WHERE code = 'WATER'),
    NULL,
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

-- Juices
(1, 'ALMARAI-ORANGE-1L', 'Almarai Orange Juice 1L', '100% pure orange juice', 
    (SELECT id FROM product_categories WHERE code = 'JUICES'),
    (SELECT id FROM brands WHERE code = 'ALMARAI'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'ALMARAI-MIXED-1L', 'Almarai Mixed Fruit Juice 1L', 'Mixed fruit juice', 
    (SELECT id FROM product_categories WHERE code = 'JUICES'),
    (SELECT id FROM brands WHERE code = 'ALMARAI'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

-- Tea & Coffee
(1, 'RABEA-TEA-100BAG', 'Rabea Tea 100 Bags', 'Premium black tea bags', 
    (SELECT id FROM product_categories WHERE code = 'TEA_COFFEE'),
    (SELECT id FROM brands WHERE code = 'RABEA'),
    (SELECT id FROM units_of_measure WHERE code = 'BOX'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'LIPTON-TEA-100BAG', 'Lipton Yellow Label Tea 100 Bags', 'Yellow label tea', 
    (SELECT id FROM product_categories WHERE code = 'TEA_COFFEE'),
    (SELECT id FROM brands WHERE code = 'LIPTON'),
    (SELECT id FROM units_of_measure WHERE code = 'BOX'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'NESCAFE-CLASSIC-200G', 'Nescafe Classic 200g', 'Instant coffee', 
    (SELECT id FROM product_categories WHERE code = 'TEA_COFFEE'),
    (SELECT id FROM brands WHERE code = 'NESCAFE'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'NESCAFE-ARABIAN-200G', 'Nescafe Arabian Coffee 200g', 'Arabian style instant coffee', 
    (SELECT id FROM product_categories WHERE code = 'TEA_COFFEE'),
    (SELECT id FROM brands WHERE code = 'NESCAFE'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true);

-- =====================================================
-- PRODUCTS - FOOD & GROCERIES
-- =====================================================

-- Rice & Grains
INSERT INTO products (organization_id, sku, name, description, category_id, brand_id, base_uom_id, product_type, tax_category_id, is_active, is_sellable, is_purchasable, track_inventory) VALUES
(1, 'RICE-BASMATI-5KG', 'Basmati Rice 5kg', 'Premium basmati rice', 
    (SELECT id FROM product_categories WHERE code = 'RICE_GRAINS'),
    NULL,
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'RICE-BASMATI-10KG', 'Basmati Rice 10kg', 'Premium basmati rice 10kg bag', 
    (SELECT id FROM product_categories WHERE code = 'RICE_GRAINS'),
    NULL,
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

-- Cooking Oil
(1, 'OIL-SUNFLOWER-1.8L', 'Sunflower Cooking Oil 1.8L', 'Pure sunflower oil', 
    (SELECT id FROM product_categories WHERE code = 'COOKING_OIL'),
    NULL,
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'OIL-CORN-1.8L', 'Corn Oil 1.8L', 'Pure corn oil', 
    (SELECT id FROM product_categories WHERE code = 'COOKING_OIL'),
    NULL,
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

-- Pasta & Noodles
(1, 'PASTA-PENNE-500G', 'Penne Pasta 500g', 'Durum wheat penne pasta', 
    (SELECT id FROM product_categories WHERE code = 'PASTA'),
    NULL,
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'PASTA-SPAGHETTI-500G', 'Spaghetti Pasta 500g', 'Durum wheat spaghetti', 
    (SELECT id FROM product_categories WHERE code = 'PASTA'),
    NULL,
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

-- Canned Foods
(1, 'CALGARDEN-BEANS-400G', 'California Garden Baked Beans 400g', 'Baked beans in tomato sauce', 
    (SELECT id FROM product_categories WHERE code = 'CANNED'),
    (SELECT id FROM brands WHERE code = 'CALGARDEN'),
    (SELECT id FROM units_of_measure WHERE code = 'CAN'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'CALGARDEN-TUNA-185G', 'California Garden Tuna Chunks 185g', 'Tuna chunks in water', 
    (SELECT id FROM product_categories WHERE code = 'CANNED'),
    (SELECT id FROM brands WHERE code = 'CALGARDEN'),
    (SELECT id FROM units_of_measure WHERE code = 'CAN'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

-- Sugar & Salt
(1, 'SUGAR-WHITE-1KG', 'White Sugar 1kg', 'Refined white sugar', 
    (SELECT id FROM product_categories WHERE code = 'SUGAR_SALT'),
    NULL,
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'SALT-TABLE-1KG', 'Table Salt 1kg', 'Iodized table salt', 
    (SELECT id FROM product_categories WHERE code = 'SUGAR_SALT'),
    NULL,
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true);

-- =====================================================
-- PRODUCTS - FROZEN FOODS
-- =====================================================

INSERT INTO products (organization_id, sku, name, description, category_id, brand_id, base_uom_id, product_type, tax_category_id, is_active, is_sellable, is_purchasable, track_inventory) VALUES
(1, 'SUNBULAH-FRIES-1KG', 'Sunbulah French Fries 1kg', 'Frozen french fries', 
    (SELECT id FROM product_categories WHERE code = 'FROZEN_SNACKS'),
    (SELECT id FROM brands WHERE code = 'SUNBULAH'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'SUNBULAH-VEGETABLES-450G', 'Sunbulah Mixed Vegetables 450g', 'Frozen mixed vegetables', 
    (SELECT id FROM product_categories WHERE code = 'FROZEN_VEG'),
    (SELECT id FROM brands WHERE code = 'SUNBULAH'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'WATANIA-CHICKEN-1KG', 'Al-Watania Frozen Chicken 1kg', 'Whole frozen chicken', 
    (SELECT id FROM product_categories WHERE code = 'FROZEN_CHICKEN'),
    (SELECT id FROM brands WHERE code = 'WATANIA'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true);

-- =====================================================
-- PRODUCTS - PERSONAL CARE
-- =====================================================

INSERT INTO products (organization_id, sku, name, description, category_id, brand_id, base_uom_id, product_type, tax_category_id, is_active, is_sellable, is_purchasable, track_inventory) VALUES
-- Bath & Shower
(1, 'DETTOL-SOAP-125G', 'Dettol Original Soap 125g', 'Antibacterial soap bar', 
    (SELECT id FROM product_categories WHERE code = 'BATH_SHOWER'),
    (SELECT id FROM brands WHERE code = 'DETTOL'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'DOVE-BODYWASH-500ML', 'Dove Body Wash 500ml', 'Moisturizing body wash', 
    (SELECT id FROM product_categories WHERE code = 'BATH_SHOWER'),
    (SELECT id FROM brands WHERE code = 'DOVE'),
    (SELECT id FROM units_of_measure WHERE code = 'ML'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'LUX-SOAP-120G', 'Lux Beauty Soap 120g', 'Beauty soap bar', 
    (SELECT id FROM product_categories WHERE code = 'BATH_SHOWER'),
    (SELECT id FROM brands WHERE code = 'LUX'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

-- Oral Care
(1, 'PALMOLIVE-TOOTHPASTE-100ML', 'Palmolive Toothpaste 100ml', 'Fresh mint toothpaste', 
    (SELECT id FROM product_categories WHERE code = 'ORAL_CARE'),
    (SELECT id FROM brands WHERE code = 'PALMOLIVE'),
    (SELECT id FROM units_of_measure WHERE code = 'ML'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true);

-- =====================================================
-- PRODUCTS - HOUSEHOLD & CLEANING
-- =====================================================

INSERT INTO products (organization_id, sku, name, description, category_id, brand_id, base_uom_id, product_type, tax_category_id, is_active, is_sellable, is_purchasable, track_inventory) VALUES
-- Laundry
(1, 'TIDE-POWDER-3KG', 'Tide Washing Powder 3kg', 'Automatic washing powder', 
    (SELECT id FROM product_categories WHERE code = 'LAUNDRY'),
    (SELECT id FROM brands WHERE code = 'TIDE'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'ARIEL-POWDER-2.5KG', 'Ariel Washing Powder 2.5kg', 'Automatic washing powder', 
    (SELECT id FROM product_categories WHERE code = 'LAUNDRY'),
    (SELECT id FROM brands WHERE code = 'ARIEL'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'PERSIL-LIQUID-3L', 'Persil Liquid Detergent 3L', 'Liquid laundry detergent', 
    (SELECT id FROM product_categories WHERE code = 'LAUNDRY'),
    (SELECT id FROM brands WHERE code = 'PERSIL'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

-- Dishwashing
(1, 'FINISH-TABS-40PCS', 'Finish Dishwasher Tablets 40pcs', 'Dishwasher tablets', 
    (SELECT id FROM product_categories WHERE code = 'DISHWASHING'),
    (SELECT id FROM brands WHERE code = 'FINISH'),
    (SELECT id FROM units_of_measure WHERE code = 'BOX'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true),

(1, 'PALMOLIVE-DISH-750ML', 'Palmolive Dishwashing Liquid 750ml', 'Dishwashing liquid', 
    (SELECT id FROM product_categories WHERE code = 'DISHWASHING'),
    (SELECT id FROM brands WHERE code = 'PALMOLIVE'),
    (SELECT id FROM units_of_measure WHERE code = 'ML'),
    'finished_good', 
    (SELECT id FROM tax_categories WHERE code = 'VAT_15'),
    true, true, true, true);

-- =====================================================
-- PRODUCT BARCODES (EAN-13 format - Saudi standard)
-- =====================================================

-- Dairy Barcodes
INSERT INTO product_barcodes (product_id, barcode, barcode_type, is_primary) VALUES
((SELECT id FROM products WHERE sku = 'ALMARAI-MILK-FW-1L'), '6281000001011', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'ALMARAI-MILK-LF-1L'), '6281000001028', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'NADEC-MILK-FW-2L'), '6281030000019', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'ALMARAI-LABAN-1L'), '6281000001035', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'ALSAFI-YOGURT-170G'), '6281007000024', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'ALMARAI-CHEESE-SLICE-200G'), '6281000001042', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'ALMARAI-FETA-CHEESE-400G'), '6281000001059', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'EGGS-WHITE-30PCS'), '6281000002011', 'EAN13', true);

-- Beverages Barcodes
INSERT INTO product_barcodes (product_id, barcode, barcode_type, is_primary) VALUES
((SELECT id FROM products WHERE sku = 'COCA-COLA-330ML'), '6281055001017', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'PEPSI-330ML'), '6281055001024', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'COCA-COLA-2L'), '6281055001031', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'WATER-600ML'), '6281000003011', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'WATER-1.5L'), '6281000003028', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'ALMARAI-ORANGE-1L'), '6281000001066', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'ALMARAI-MIXED-1L'), '6281000001073', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'RABEA-TEA-100BAG'), '6281000004011', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'LIPTON-TEA-100BAG'), '8722700001089', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'NESCAFE-CLASSIC-200G'), '7613035814196', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'NESCAFE-ARABIAN-200G'), '7613035814202', 'EAN13', true);

-- Food & Groceries Barcodes
INSERT INTO product_barcodes (product_id, barcode, barcode_type, is_primary) VALUES
((SELECT id FROM products WHERE sku = 'RICE-BASMATI-5KG'), '6281000005011', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'RICE-BASMATI-10KG'), '6281000005028', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'OIL-SUNFLOWER-1.8L'), '6281000006011', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'OIL-CORN-1.8L'), '6281000006028', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'PASTA-PENNE-500G'), '6281000007011', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'PASTA-SPAGHETTI-500G'), '6281000007028', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'CALGARDEN-BEANS-400G'), '6281000008011', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'CALGARDEN-TUNA-185G'), '6281000008028', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'SUGAR-WHITE-1KG'), '6281000009011', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'SALT-TABLE-1KG'), '6281000009028', 'EAN13', true);

-- Frozen Foods Barcodes
INSERT INTO product_barcodes (product_id, barcode, barcode_type, is_primary) VALUES
((SELECT id FROM products WHERE sku = 'SUNBULAH-FRIES-1KG'), '6281006000019', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'SUNBULAH-VEGETABLES-450G'), '6281006000026', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'WATANIA-CHICKEN-1KG'), '6281000010011', 'EAN13', true);

-- Personal Care Barcodes
INSERT INTO product_barcodes (product_id, barcode, barcode_type, is_primary) VALUES
((SELECT id FROM products WHERE sku = 'DETTOL-SOAP-125G'), '6281001001017', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'DOVE-BODYWASH-500ML'), '8710908231445', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'LUX-SOAP-120G'), '8901030672408', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'PALMOLIVE-TOOTHPASTE-100ML'), '8714789730127', 'EAN13', true);

-- Household Barcodes
INSERT INTO product_barcodes (product_id, barcode, barcode_type, is_primary) VALUES
((SELECT id FROM products WHERE sku = 'TIDE-POWDER-3KG'), '8001841501130', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'ARIEL-POWDER-2.5KG'), '8001841501147', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'PERSIL-LIQUID-3L'), '9000101326291', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'FINISH-TABS-40PCS'), '5900627064087', 'EAN13', true),
((SELECT id FROM products WHERE sku = 'PALMOLIVE-DISH-750ML'), '8714789730158', 'EAN13', true);

-- =====================================================
-- PRODUCT PRICES - RETAIL
-- =====================================================

-- Dairy Prices (Retail)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active) VALUES
-- Milk & Dairy
((SELECT id FROM products WHERE sku = 'ALMARAI-MILK-FW-1L'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'LTR'), 8.50, 1, true),

((SELECT id FROM products WHERE sku = 'ALMARAI-MILK-LF-1L'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'LTR'), 8.50, 1, true),

((SELECT id FROM products WHERE sku = 'NADEC-MILK-FW-2L'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'LTR'), 14.95, 1, true),

((SELECT id FROM products WHERE sku = 'ALMARAI-LABAN-1L'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'LTR'), 6.95, 1, true),

((SELECT id FROM products WHERE sku = 'ALSAFI-YOGURT-170G'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'GM'), 3.50, 1, true),

((SELECT id FROM products WHERE sku = 'ALMARAI-CHEESE-SLICE-200G'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'GM'), 12.95, 1, true),

((SELECT id FROM products WHERE sku = 'ALMARAI-FETA-CHEESE-400G'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'GM'), 15.50, 1, true),

((SELECT id FROM products WHERE sku = 'EGGS-WHITE-30PCS'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'PCS'), 18.00, 1, true);

-- Beverages Prices (Retail)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active) VALUES
((SELECT id FROM products WHERE sku = 'COCA-COLA-330ML'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'CAN'), 2.00, 1, true),

((SELECT id FROM products WHERE sku = 'PEPSI-330ML'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'CAN'), 2.00, 1, true),

((SELECT id FROM products WHERE sku = 'COCA-COLA-2L'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'BTL'), 6.50, 1, true),

((SELECT id FROM products WHERE sku = 'WATER-600ML'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'BTL'), 1.00, 1, true),

((SELECT id FROM products WHERE sku = 'WATER-1.5L'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'BTL'), 1.50, 1, true),

((SELECT id FROM products WHERE sku = 'ALMARAI-ORANGE-1L'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'LTR'), 7.95, 1, true),

((SELECT id FROM products WHERE sku = 'ALMARAI-MIXED-1L'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'LTR'), 7.95, 1, true),

((SELECT id FROM products WHERE sku = 'RABEA-TEA-100BAG'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'BOX'), 12.50, 1, true),

((SELECT id FROM products WHERE sku = 'LIPTON-TEA-100BAG'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'BOX'), 14.95, 1, true),

((SELECT id FROM products WHERE sku = 'NESCAFE-CLASSIC-200G'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'GM'), 28.50, 1, true),

((SELECT id FROM products WHERE sku = 'NESCAFE-ARABIAN-200G'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'GM'), 32.00, 1, true);

-- Food & Groceries Prices (Retail)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active) VALUES
((SELECT id FROM products WHERE sku = 'RICE-BASMATI-5KG'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'KG'), 45.00, 1, true),

((SELECT id FROM products WHERE sku = 'RICE-BASMATI-10KG'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'KG'), 85.00, 1, true),

((SELECT id FROM products WHERE sku = 'OIL-SUNFLOWER-1.8L'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'LTR'), 18.95, 1, true),

((SELECT id FROM products WHERE sku = 'OIL-CORN-1.8L'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'LTR'), 19.95, 1, true),

((SELECT id FROM products WHERE sku = 'PASTA-PENNE-500G'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'GM'), 6.50, 1, true),

((SELECT id FROM products WHERE sku = 'PASTA-SPAGHETTI-500G'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'GM'), 6.50, 1, true),

((SELECT id FROM products WHERE sku = 'CALGARDEN-BEANS-400G'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'CAN'), 4.50, 1, true),

((SELECT id FROM products WHERE sku = 'CALGARDEN-TUNA-185G'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'CAN'), 8.95, 1, true),

((SELECT id FROM products WHERE sku = 'SUGAR-WHITE-1KG'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'KG'), 5.50, 1, true),

((SELECT id FROM products WHERE sku = 'SALT-TABLE-1KG'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'KG'), 3.00, 1, true);

-- Frozen Foods Prices (Retail)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active) VALUES
((SELECT id FROM products WHERE sku = 'SUNBULAH-FRIES-1KG'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'KG'), 12.95, 1, true),

((SELECT id FROM products WHERE sku = 'SUNBULAH-VEGETABLES-450G'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'GM'), 9.50, 1, true),

((SELECT id FROM products WHERE sku = 'WATANIA-CHICKEN-1KG'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'KG'), 22.00, 1, true);

-- Personal Care Prices (Retail)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active) VALUES
((SELECT id FROM products WHERE sku = 'DETTOL-SOAP-125G'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'GM'), 4.50, 1, true),

((SELECT id FROM products WHERE sku = 'DOVE-BODYWASH-500ML'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'ML'), 24.95, 1, true),

((SELECT id FROM products WHERE sku = 'LUX-SOAP-120G'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'GM'), 3.50, 1, true),

((SELECT id FROM products WHERE sku = 'PALMOLIVE-TOOTHPASTE-100ML'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'ML'), 8.95, 1, true);

-- Household Prices (Retail)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active) VALUES
((SELECT id FROM products WHERE sku = 'TIDE-POWDER-3KG'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'KG'), 39.95, 1, true),

((SELECT id FROM products WHERE sku = 'ARIEL-POWDER-2.5KG'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'KG'), 35.95, 1, true),

((SELECT id FROM products WHERE sku = 'PERSIL-LIQUID-3L'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'LTR'), 42.00, 1, true),

((SELECT id FROM products WHERE sku = 'FINISH-TABS-40PCS'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'BOX'), 48.50, 1, true),

((SELECT id FROM products WHERE sku = 'PALMOLIVE-DISH-750ML'), 
    (SELECT id FROM price_lists WHERE code = 'RETAIL'), 
    (SELECT id FROM units_of_measure WHERE code = 'ML'), 12.95, 1, true);

-- =====================================================
-- PRODUCT PRICES - WHOLESALE (15-20% discount)
-- =====================================================

-- Dairy Prices (Wholesale)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active) VALUES
((SELECT id FROM products WHERE sku = 'ALMARAI-MILK-FW-1L'), 
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'), 
    (SELECT id FROM units_of_measure WHERE code = 'LTR'), 7.25, 12, true),

((SELECT id FROM products WHERE sku = 'NADEC-MILK-FW-2L'), 
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'), 
    (SELECT id FROM units_of_measure WHERE code = 'LTR'), 12.75, 6, true),

((SELECT id FROM products WHERE sku = 'ALMARAI-LABAN-1L'), 
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'), 
    (SELECT id FROM units_of_measure WHERE code = 'LTR'), 5.95, 12, true),

((SELECT id FROM products WHERE sku = 'EGGS-WHITE-30PCS'), 
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'), 
    (SELECT id FROM units_of_measure WHERE code = 'PCS'), 15.50, 10, true),

-- Beverages (Wholesale - Carton pricing)
((SELECT id FROM products WHERE sku = 'COCA-COLA-330ML'), 
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'), 
    (SELECT id FROM units_of_measure WHERE code = 'CAN'), 1.70, 24, true),

((SELECT id FROM products WHERE sku = 'PEPSI-330ML'), 
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'), 
    (SELECT id FROM units_of_measure WHERE code = 'CAN'), 1.70, 24, true),

((SELECT id FROM products WHERE sku = 'WATER-600ML'), 
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'), 
    (SELECT id FROM units_of_measure WHERE code = 'BTL'), 0.75, 24, true),

((SELECT id FROM products WHERE sku = 'RABEA-TEA-100BAG'), 
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'), 
    (SELECT id FROM units_of_measure WHERE code = 'BOX'), 10.50, 12, true),

((SELECT id FROM products WHERE sku = 'NESCAFE-CLASSIC-200G'), 
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'), 
    (SELECT id FROM units_of_measure WHERE code = 'GM'), 24.50, 12, true),

-- Food (Wholesale)
((SELECT id FROM products WHERE sku = 'RICE-BASMATI-10KG'), 
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'), 
    (SELECT id FROM units_of_measure WHERE code = 'KG'), 75.00, 5, true),

((SELECT id FROM products WHERE sku = 'OIL-SUNFLOWER-1.8L'), 
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'), 
    (SELECT id FROM units_of_measure WHERE code = 'LTR'), 16.50, 6, true),

((SELECT id FROM products WHERE sku = 'SUGAR-WHITE-1KG'), 
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'), 
    (SELECT id FROM units_of_measure WHERE code = 'KG'), 4.75, 10, true),

-- Household (Wholesale)
((SELECT id FROM products WHERE sku = 'TIDE-POWDER-3KG'), 
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'), 
    (SELECT id FROM units_of_measure WHERE code = 'KG'), 34.50, 6, true),

((SELECT id FROM products WHERE sku = 'ARIEL-POWDER-2.5KG'), 
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'), 
    (SELECT id FROM units_of_measure WHERE code = 'KG'), 30.50, 6, true),

((SELECT id FROM products WHERE sku = 'FINISH-TABS-40PCS'), 
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'), 
    (SELECT id FROM units_of_measure WHERE code = 'BOX'), 42.00, 6, true);

-- =====================================================
-- SUMMARY QUERY
-- =====================================================

-- Query to verify the data
SELECT 
    'Categories' as data_type,
    COUNT(*) as count
FROM product_categories
UNION ALL
SELECT 
    'Brands' as data_type,
    COUNT(*) as count
FROM brands
UNION ALL
SELECT 
    'Products' as data_type,
    COUNT(*) as count
FROM products
UNION ALL
SELECT 
    'Barcodes' as data_type,
    COUNT(*) as count
FROM product_barcodes
UNION ALL
SELECT 
    'Retail Prices' as data_type,
    COUNT(*) as count
FROM product_prices
WHERE price_list_id = (SELECT id FROM price_lists WHERE code = 'RETAIL')
UNION ALL
SELECT 
    'Wholesale Prices' as data_type,
    COUNT(*) as count
FROM product_prices
WHERE price_list_id = (SELECT id FROM price_lists WHERE code = 'WHOLESALE');

-- =====================================================
-- DETAILED PRODUCT LIST WITH CATEGORIES AND PRICES
-- =====================================================

SELECT 
    pc.name as category,
    b.name as brand,
    p.sku,
    p.name as product_name,
    pb.barcode,
    uom.code as unit,
    pp_retail.price as retail_price,
    pp_wholesale.price as wholesale_price,
    tc.name as tax_category,
    tc.tax_rate
FROM products p
LEFT JOIN product_categories pc ON p.category_id = pc.id
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN units_of_measure uom ON p.base_uom_id = uom.id
LEFT JOIN product_barcodes pb ON p.id = pb.product_id AND pb.is_primary = true
LEFT JOIN product_prices pp_retail ON p.id = pp_retail.product_id 
    AND pp_retail.price_list_id = (SELECT id FROM price_lists WHERE code = 'RETAIL')
LEFT JOIN product_prices pp_wholesale ON p.id = pp_wholesale.product_id 
    AND pp_wholesale.price_list_id = (SELECT id FROM price_lists WHERE code = 'WHOLESALE')
LEFT JOIN tax_categories tc ON p.tax_category_id = tc.id
WHERE p.organization_id = 1
ORDER BY pc.name, b.name, p.name;

-- =====================================================
-- STORAGE LOCATIONS FOR STORES
-- =====================================================

-- Riyadh Flagship Store Locations
INSERT INTO storage_locations (store_id, code, name, location_type, is_active) VALUES
((SELECT id FROM stores WHERE code = 'RYD-001'), 'RYD-DAIRY', 'Dairy Section', 'retail_floor', true),
((SELECT id FROM stores WHERE code = 'RYD-001'), 'RYD-BEVERAGE', 'Beverage Section', 'retail_floor', true),
((SELECT id FROM stores WHERE code = 'RYD-001'), 'RYD-FOOD', 'Food & Groceries', 'retail_floor', true),
((SELECT id FROM stores WHERE code = 'RYD-001'), 'RYD-FROZEN', 'Frozen Section', 'retail_floor', true),
((SELECT id FROM stores WHERE code = 'RYD-001'), 'RYD-PERSONAL', 'Personal Care', 'retail_floor', true),
((SELECT id FROM stores WHERE code = 'RYD-001'), 'RYD-HOUSEHOLD', 'Household Products', 'retail_floor', true),
((SELECT id FROM stores WHERE code = 'RYD-001'), 'RYD-BACK', 'Back Storage', 'backroom', true);

-- Warehouse Locations
INSERT INTO storage_locations (store_id, code, name, location_type, is_active) VALUES
((SELECT id FROM stores WHERE code = 'WH-RYD-001'), 'WH-ZONE-A', 'Zone A - Dry Goods', 'warehouse_zone', true),
((SELECT id FROM stores WHERE code = 'WH-RYD-001'), 'WH-ZONE-B', 'Zone B - Beverages', 'warehouse_zone', true),
((SELECT id FROM stores WHERE code = 'WH-RYD-001'), 'WH-ZONE-C', 'Zone C - Cold Storage', 'warehouse_zone', true),
((SELECT id FROM stores WHERE code = 'WH-RYD-001'), 'WH-ZONE-D', 'Zone D - Frozen', 'warehouse_zone', true);

-- =====================================================
-- INITIALIZE INVENTORY STOCK FOR RIYADH FLAGSHIP STORE
-- =====================================================

-- Helper function to get product and store IDs
DO $$
DECLARE
    v_store_id INTEGER;
    v_product_id INTEGER;
    v_location_id INTEGER;
BEGIN
    -- Get Riyadh store ID
    SELECT id INTO v_store_id FROM stores WHERE code = 'RYD-001';

    -- Dairy Products Stock
    INSERT INTO inventory_stock (product_id, store_id, storage_location_id, quantity_on_hand, quantity_available, reorder_level, reorder_quantity, max_stock_level) VALUES
    ((SELECT id FROM products WHERE sku = 'ALMARAI-MILK-FW-1L'), v_store_id, 
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-DAIRY'), 
        120, 120, 30, 100, 200),
    ((SELECT id FROM products WHERE sku = 'ALMARAI-MILK-LF-1L'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-DAIRY'),
        80, 80, 20, 80, 150),
    ((SELECT id FROM products WHERE sku = 'NADEC-MILK-FW-2L'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-DAIRY'),
        60, 60, 15, 60, 120),
    ((SELECT id FROM products WHERE sku = 'ALMARAI-LABAN-1L'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-DAIRY'),
        100, 100, 25, 80, 180),
    ((SELECT id FROM products WHERE sku = 'ALSAFI-YOGURT-170G'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-DAIRY'),
        150, 150, 40, 120, 250),
    ((SELECT id FROM products WHERE sku = 'ALMARAI-CHEESE-SLICE-200G'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-DAIRY'),
        90, 90, 20, 70, 150),
    ((SELECT id FROM products WHERE sku = 'ALMARAI-FETA-CHEESE-400G'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-DAIRY'),
        75, 75, 15, 60, 120),
    ((SELECT id FROM products WHERE sku = 'EGGS-WHITE-30PCS'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-DAIRY'),
        50, 50, 10, 40, 100);

    -- Beverages Stock
    INSERT INTO inventory_stock (product_id, store_id, storage_location_id, quantity_on_hand, quantity_available, reorder_level, reorder_quantity, max_stock_level) VALUES
    ((SELECT id FROM products WHERE sku = 'COCA-COLA-330ML'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-BEVERAGE'),
        288, 288, 72, 240, 500),
    ((SELECT id FROM products WHERE sku = 'PEPSI-330ML'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-BEVERAGE'),
        240, 240, 60, 200, 450),
    ((SELECT id FROM products WHERE sku = 'COCA-COLA-2L'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-BEVERAGE'),
        72, 72, 20, 60, 120),
    ((SELECT id FROM products WHERE sku = 'WATER-600ML'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-BEVERAGE'),
        480, 480, 120, 400, 800),
    ((SELECT id FROM products WHERE sku = 'WATER-1.5L'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-BEVERAGE'),
        360, 360, 90, 300, 600),
    ((SELECT id FROM products WHERE sku = 'ALMARAI-ORANGE-1L'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-BEVERAGE'),
        100, 100, 25, 80, 180),
    ((SELECT id FROM products WHERE sku = 'ALMARAI-MIXED-1L'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-BEVERAGE'),
        85, 85, 20, 70, 150),
    ((SELECT id FROM products WHERE sku = 'RABEA-TEA-100BAG'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-BEVERAGE'),
        60, 60, 15, 50, 100),
    ((SELECT id FROM products WHERE sku = 'LIPTON-TEA-100BAG'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-BEVERAGE'),
        55, 55, 12, 45, 90),
    ((SELECT id FROM products WHERE sku = 'NESCAFE-CLASSIC-200G'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-BEVERAGE'),
        70, 70, 18, 60, 120),
    ((SELECT id FROM products WHERE sku = 'NESCAFE-ARABIAN-200G'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-BEVERAGE'),
        65, 65, 16, 55, 110);

    -- Food & Groceries Stock
    INSERT INTO inventory_stock (product_id, store_id, storage_location_id, quantity_on_hand, quantity_available, reorder_level, reorder_quantity, max_stock_level) VALUES
    ((SELECT id FROM products WHERE sku = 'RICE-BASMATI-5KG'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-FOOD'),
        80, 80, 20, 60, 150),
    ((SELECT id FROM products WHERE sku = 'RICE-BASMATI-10KG'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-FOOD'),
        45, 45, 10, 40, 80),
    ((SELECT id FROM products WHERE sku = 'OIL-SUNFLOWER-1.8L'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-FOOD'),
        90, 90, 25, 70, 150),
    ((SELECT id FROM products WHERE sku = 'OIL-CORN-1.8L'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-FOOD'),
        75, 75, 20, 60, 130),
    ((SELECT id FROM products WHERE sku = 'PASTA-PENNE-500G'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-FOOD'),
        120, 120, 30, 100, 200),
    ((SELECT id FROM products WHERE sku = 'PASTA-SPAGHETTI-500G'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-FOOD'),
        110, 110, 28, 90, 180),
    ((SELECT id FROM products WHERE sku = 'CALGARDEN-BEANS-400G'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-FOOD'),
        95, 95, 24, 80, 160),
    ((SELECT id FROM products WHERE sku = 'CALGARDEN-TUNA-185G'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-FOOD'),
        85, 85, 22, 70, 140),
    ((SELECT id FROM products WHERE sku = 'SUGAR-WHITE-1KG'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-FOOD'),
        100, 100, 25, 80, 180),
    ((SELECT id FROM products WHERE sku = 'SALT-TABLE-1KG'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-FOOD'),
        80, 80, 20, 65, 140);

    -- Frozen Foods Stock
    INSERT INTO inventory_stock (product_id, store_id, storage_location_id, quantity_on_hand, quantity_available, reorder_level, reorder_quantity, max_stock_level) VALUES
    ((SELECT id FROM products WHERE sku = 'SUNBULAH-FRIES-1KG'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-FROZEN'),
        70, 70, 18, 60, 120),
    ((SELECT id FROM products WHERE sku = 'SUNBULAH-VEGETABLES-450G'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-FROZEN'),
        65, 65, 16, 55, 110),
    ((SELECT id FROM products WHERE sku = 'WATANIA-CHICKEN-1KG'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-FROZEN'),
        50, 50, 12, 45, 90);

    -- Personal Care Stock
    INSERT INTO inventory_stock (product_id, store_id, storage_location_id, quantity_on_hand, quantity_available, reorder_level, reorder_quantity, max_stock_level) VALUES
    ((SELECT id FROM products WHERE sku = 'DETTOL-SOAP-125G'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-PERSONAL'),
        150, 150, 40, 120, 250),
    ((SELECT id FROM products WHERE sku = 'DOVE-BODYWASH-500ML'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-PERSONAL'),
        80, 80, 20, 65, 130),
    ((SELECT id FROM products WHERE sku = 'LUX-SOAP-120G'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-PERSONAL'),
        140, 140, 35, 110, 230),
    ((SELECT id FROM products WHERE sku = 'PALMOLIVE-TOOTHPASTE-100ML'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-PERSONAL'),
        100, 100, 25, 85, 170);

    -- Household Stock
    INSERT INTO inventory_stock (product_id, store_id, storage_location_id, quantity_on_hand, quantity_available, reorder_level, reorder_quantity, max_stock_level) VALUES
    ((SELECT id FROM products WHERE sku = 'TIDE-POWDER-3KG'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-HOUSEHOLD'),
        60, 60, 15, 50, 100),
    ((SELECT id FROM products WHERE sku = 'ARIEL-POWDER-2.5KG'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-HOUSEHOLD'),
        55, 55, 14, 45, 90),
    ((SELECT id FROM products WHERE sku = 'PERSIL-LIQUID-3L'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-HOUSEHOLD'),
        50, 50, 12, 40, 85),
    ((SELECT id FROM products WHERE sku = 'FINISH-TABS-40PCS'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-HOUSEHOLD'),
        45, 45, 11, 38, 75),
    ((SELECT id FROM products WHERE sku = 'PALMOLIVE-DISH-750ML'), v_store_id,
        (SELECT id FROM storage_locations WHERE store_id = v_store_id AND code = 'RYD-HOUSEHOLD'),
        85, 85, 22, 70, 140);

END $$;

-- =====================================================
-- INITIALIZE STOCK FOR OTHER STORES (Lower quantities)
-- =====================================================

-- Jeddah Store (60% of Riyadh stock)
INSERT INTO inventory_stock (product_id, store_id, quantity_on_hand, quantity_available, reorder_level, reorder_quantity)
SELECT 
    product_id,
    (SELECT id FROM stores WHERE code = 'JED-001'),
    ROUND(quantity_on_hand * 0.6),
    ROUND(quantity_available * 0.6),
    ROUND(reorder_level * 0.6),
    ROUND(reorder_quantity * 0.6)
FROM inventory_stock
WHERE store_id = (SELECT id FROM stores WHERE code = 'RYD-001');

-- Dammam Store (40% of Riyadh stock)
INSERT INTO inventory_stock (product_id, store_id, quantity_on_hand, quantity_available, reorder_level, reorder_quantity)
SELECT 
    product_id,
    (SELECT id FROM stores WHERE code = 'DMM-001'),
    ROUND(quantity_on_hand * 0.4),
    ROUND(quantity_available * 0.4),
    ROUND(reorder_level * 0.4),
    ROUND(reorder_quantity * 0.4)
FROM inventory_stock
WHERE store_id = (SELECT id FROM stores WHERE code = 'RYD-001');

-- Warehouse (200% of Riyadh stock)
INSERT INTO inventory_stock (product_id, store_id, quantity_on_hand, quantity_available, reorder_level, reorder_quantity)
SELECT 
    product_id,
    (SELECT id FROM stores WHERE code = 'WH-RYD-001'),
    ROUND(quantity_on_hand * 2),
    ROUND(quantity_available * 2),
    ROUND(reorder_level * 2),
    ROUND(reorder_quantity * 2)
FROM inventory_stock
WHERE store_id = (SELECT id FROM stores WHERE code = 'RYD-001');

-- Wholesale Center (150% of Riyadh stock)
INSERT INTO inventory_stock (product_id, store_id, quantity_on_hand, quantity_available, reorder_level, reorder_quantity)
SELECT 
    product_id,
    (SELECT id FROM stores WHERE code = 'WHSL-RYD-001'),
    ROUND(quantity_on_hand * 1.5),
    ROUND(quantity_available * 1.5),
    ROUND(reorder_level * 1.5),
    ROUND(reorder_quantity * 1.5)
FROM inventory_stock
WHERE store_id = (SELECT id FROM stores WHERE code = 'RYD-001');

-- =====================================================
-- CREATE POS TERMINALS
-- =====================================================

-- Riyadh Flagship Store Terminals
INSERT INTO pos_terminals (store_id, terminal_code, terminal_name, device_id, is_active, metadata) VALUES
((SELECT id FROM stores WHERE code = 'RYD-001'), 'POS-RYD-01', 'Checkout Counter 1', 'DEVICE-RYD-001', true, '{"location": "Front"}'),
((SELECT id FROM stores WHERE code = 'RYD-001'), 'POS-RYD-02', 'Checkout Counter 2', 'DEVICE-RYD-002', true, '{"location": "Front"}'),
((SELECT id FROM stores WHERE code = 'RYD-001'), 'POS-RYD-03', 'Checkout Counter 3', 'DEVICE-RYD-003', true, '{"location": "Front"}'),
((SELECT id FROM stores WHERE code = 'RYD-001'), 'POS-RYD-04', 'Express Checkout', 'DEVICE-RYD-004', true, '{"location": "Express Lane"}');

-- Jeddah Store Terminals
INSERT INTO pos_terminals (store_id, terminal_code, terminal_name, device_id, is_active, metadata) VALUES
((SELECT id FROM stores WHERE code = 'JED-001'), 'POS-JED-01', 'Checkout Counter 1', 'DEVICE-JED-001', true, '{"location": "Front"}'),
((SELECT id FROM stores WHERE code = 'JED-001'), 'POS-JED-02', 'Checkout Counter 2', 'DEVICE-JED-002', true, '{"location": "Front"}');

-- Wholesale Center Terminal
INSERT INTO pos_terminals (store_id, terminal_code, terminal_name, device_id, is_active, metadata) VALUES
((SELECT id FROM stores WHERE code = 'WHSL-RYD-001'), 'POS-WHSL-01', 'Wholesale Counter', 'DEVICE-WHSL-001', true, '{"location": "Main Counter"}');

-- =====================================================
-- CREATE SAMPLE PROMOTIONAL PRICES (Offers)
-- =====================================================

-- Update valid dates for promotional price list
UPDATE price_lists 
SET valid_from = CURRENT_DATE, 
    valid_to = CURRENT_DATE + INTERVAL '30 days'
WHERE code = 'PROMO';

-- Add promotional prices (10-30% discount on selected items)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, valid_from, valid_to, is_active, metadata) VALUES
-- Dairy Promotions
((SELECT id FROM products WHERE sku = 'ALMARAI-MILK-FW-1L'), 
    (SELECT id FROM price_lists WHERE code = 'PROMO'), 
    (SELECT id FROM units_of_measure WHERE code = 'LTR'), 
    6.99, 1, CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days', true,
    '{"promotion_name": "Weekly Special", "discount_percent": 18}'),

((SELECT id FROM products WHERE sku = 'ALMARAI-LABAN-1L'), 
    (SELECT id FROM price_lists WHERE code = 'PROMO'), 
    (SELECT id FROM units_of_measure WHERE code = 'LTR'), 
    5.50, 2, CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days', true,
    '{"promotion_name": "Buy 2 Get Discount", "discount_percent": 21}'),

-- Beverage Promotions
((SELECT id FROM products WHERE sku = 'COCA-COLA-330ML'), 
    (SELECT id FROM price_lists WHERE code = 'PROMO'), 
    (SELECT id FROM units_of_measure WHERE code = 'CAN'), 
    1.50, 6, CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days', true,
    '{"promotion_name": "6-Pack Deal", "discount_percent": 25}'),

((SELECT id FROM products WHERE sku = 'WATER-600ML'), 
    (SELECT id FROM price_lists WHERE code = 'PROMO'), 
    (SELECT id FROM units_of_measure WHERE code = 'BTL'), 
    0.75, 12, CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days', true,
    '{"promotion_name": "12-Pack Deal", "discount_percent": 25}'),

((SELECT id FROM products WHERE sku = 'NESCAFE-CLASSIC-200G'), 
    (SELECT id FROM price_lists WHERE code = 'PROMO'), 
    (SELECT id FROM units_of_measure WHERE code = 'GM'), 
    24.99, 1, CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days', true,
    '{"promotion_name": "Coffee Week", "discount_percent": 12}'),

-- Food Promotions
((SELECT id FROM products WHERE sku = 'RICE-BASMATI-10KG'), 
    (SELECT id FROM price_lists WHERE code = 'PROMO'), 
    (SELECT id FROM units_of_measure WHERE code = 'KG'), 
    69.99, 1, CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days', true,
    '{"promotion_name": "Rice Festival", "discount_percent": 18}'),

((SELECT id FROM products WHERE sku = 'SUGAR-WHITE-1KG'), 
    (SELECT id FROM price_lists WHERE code = 'PROMO'), 
    (SELECT id FROM units_of_measure WHERE code = 'KG'), 
    4.99, 2, CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days', true,
    '{"promotion_name": "Multi-buy Deal", "discount_percent": 9}'),

-- Household Promotions
((SELECT id FROM products WHERE sku = 'TIDE-POWDER-3KG'), 
    (SELECT id FROM price_lists WHERE code = 'PROMO'), 
    (SELECT id FROM units_of_measure WHERE code = 'KG'), 
    34.99, 1, CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days', true,
    '{"promotion_name": "Cleaning Month", "discount_percent": 12}'),

((SELECT id FROM products WHERE sku = 'FINISH-TABS-40PCS'), 
    (SELECT id FROM price_lists WHERE code = 'PROMO'), 
    (SELECT id FROM units_of_measure WHERE code = 'BOX'), 
    39.99, 1, CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days', true,
    '{"promotion_name": "Cleaning Month", "discount_percent": 18}');


-- =====================================================
-- 2. PERMISSIONS
-- =====================================================

INSERT INTO permissions (name, code, description) VALUES
-- Dashboard Permissions
('View Dashboard', 'dashboard:view', 'Can view dashboard and analytics'),
('Manage Dashboard', 'dashboard:manage', 'Can customize and manage dashboard layouts'),
('Export Dashboard', 'dashboard:export', 'Can export dashboard reports'),

-- Tenant Management Permissions
('View Tenants', 'tenants:view', 'Can view tenant information'),
('Manage Tenants', 'tenants:manage', 'Can create and edit tenants'),
('Delete Tenants', 'tenants:delete', 'Can delete tenants'),
('Configure Tenants', 'tenants:configure', 'Can configure tenant settings and features'),

-- Organization Management
('View Organizations', 'organizations:view', 'Can view organization details'),
('Manage Organizations', 'organizations:manage', 'Can create and edit organizations'),
('Delete Organizations', 'organizations:delete', 'Can delete organizations'),

-- User Management Permissions
('View Users', 'users:view', 'Can view user list and details'),
('Manage Users', 'users:manage', 'Can create and edit users'),
('Delete Users', 'users:delete', 'Can delete users'),
('Reset User Password', 'users:reset_password', 'Can reset user passwords'),

-- Role Management
('View Roles', 'roles:view', 'Can view roles list'),
('Manage Roles', 'roles:manage', 'Can create and edit roles'),
('Delete Roles', 'roles:delete', 'Can delete roles'),
('Assign Roles', 'roles:assign', 'Can assign roles to users'),

-- UI Module Management
('View UI Modules', 'ui_modules:view', 'Can view UI modules and menus'),
('Manage UI Modules', 'ui_modules:manage', 'Can create and edit UI modules'),
('Delete UI Modules', 'ui_modules:delete', 'Can delete UI modules'),

-- Permission Management
('View Permissions', 'permissions:view', 'Can view permissions'),
('Manage Permissions', 'permissions:manage', 'Can create and edit permissions'),
('Delete Permissions', 'permissions:delete', 'Can delete permissions'),

-- Store Management
('View Stores', 'stores:view', 'Can view store information'),
('Manage Stores', 'stores:manage', 'Can create and edit stores'),
('Delete Stores', 'stores:delete', 'Can delete stores'),
('Configure Stores', 'stores:configure', 'Can configure store settings'),

-- POS Management
('View POS', 'pos:view', 'Can view POS transactions and terminals'),
('Manage POS', 'pos:manage', 'Can configure POS terminals and settings'),
('Process Sales', 'pos:process_sales', 'Can process sales transactions'),
('Void Transactions', 'pos:void_transactions', 'Can void POS transactions'),
('Apply Discounts', 'pos:apply_discounts', 'Can apply discounts to transactions'),
('Process Returns', 'pos:process_returns', 'Can process return transactions'),
('View POS Reports', 'pos:view_reports', 'Can view POS reports and analytics'),

-- Cashier Management
('View Cashiers', 'cashiers:view', 'Can view cashier information'),
('Manage Cashiers', 'cashiers:manage', 'Can create and edit cashiers'),
('Delete Cashiers', 'cashiers:delete', 'Can delete cashiers'),
('Manage Sessions', 'cashiers:manage_sessions', 'Can open/close cashier sessions'),
('View Sessions', 'cashiers:view_sessions', 'Can view cashier session history'),

-- Inventory Management
('View Inventory', 'inventory:view', 'Can view inventory levels and stock'),
('Manage Inventory', 'inventory:manage', 'Can adjust inventory and manage stock'),
('Transfer Inventory', 'inventory:transfer', 'Can transfer stock between locations'),
('Conduct Stock Count', 'inventory:stock_count', 'Can perform stock counts'),
('View Inventory Reports', 'inventory:view_reports', 'Can view inventory reports'),

-- Product Management
('View Products', 'products:view', 'Can view product catalog'),
('Manage Products', 'products:manage', 'Can create and edit products'),
('Delete Products', 'products:delete', 'Can delete products'),
('Manage Pricing', 'products:manage_pricing', 'Can manage product prices'),
('View Cost Prices', 'products:view_cost', 'Can view product cost prices'),

-- Customer Management
('View Customers', 'customers:view', 'Can view customer information'),
('Manage Customers', 'customers:manage', 'Can create and edit customers'),
('Delete Customers', 'customers:delete', 'Can delete customers'),
('View Customer History', 'customers:view_history', 'Can view customer purchase history'),

-- Supplier Management
('View Suppliers', 'suppliers:view', 'Can view supplier information'),
('Manage Suppliers', 'suppliers:manage', 'Can create and edit suppliers'),
('Delete Suppliers', 'suppliers:delete', 'Can delete suppliers'),

-- Purchase Order Management
('View Purchase Orders', 'purchase_orders:view', 'Can view purchase orders'),
('Manage Purchase Orders', 'purchase_orders:manage', 'Can create and edit purchase orders'),
('Delete Purchase Orders', 'purchase_orders:delete', 'Can delete purchase orders'),
('Approve Purchase Orders', 'purchase_orders:approve', 'Can approve purchase orders'),

-- Sales Order Management
('View Sales Orders', 'sales_orders:view', 'Can view sales orders'),
('Manage Sales Orders', 'sales_orders:manage', 'Can create and edit sales orders'),
('Delete Sales Orders', 'sales_orders:delete', 'Can delete sales orders'),
('Approve Sales Orders', 'sales_orders:approve', 'Can approve sales orders'),

-- Reports & Analytics
('View Sales Reports', 'reports:sales', 'Can view sales reports and analytics'),
('View Purchase Reports', 'reports:purchases', 'Can view purchase reports'),
('View Inventory Reports', 'reports:inventory', 'Can view inventory reports'),
('View Financial Reports', 'reports:financial', 'Can view P&L and financial reports'),
('Export Reports', 'reports:export', 'Can export reports to various formats'),

-- System Settings
('View Settings', 'settings:view', 'Can view system settings'),
('Manage Settings', 'settings:manage', 'Can modify system settings'),
('View Audit Logs', 'settings:audit_logs', 'Can view system audit logs');

-- =====================================================
-- 3. MODULES
-- =====================================================

INSERT INTO modules (name, code, description, icon, display_order, is_active) VALUES
('Dashboard', 'dashboard', 'Main dashboard and overview with analytics', 'dashboard', 1, true),
('Tenant Management', 'tenants', 'Multi-tenant configuration and management', 'building', 2, true),
('Organization Setup', 'organizations', 'Organization and company structure management', 'briefcase', 3, true),
('User Management', 'users', 'User accounts and authentication', 'users', 4, true),
('Store Management', 'stores', 'Store locations and configuration', 'store', 5, true),
('Point of Sale', 'pos', 'POS transactions and terminal management', 'shopping-cart', 6, true),
('Cashier Operations', 'cashiers', 'Cashier management and session control', 'user-check', 7, true),
('Inventory Management', 'inventory', 'Stock control and warehouse management', 'package', 8, true),
('Product Catalog', 'products', 'Product master data and catalog', 'box', 9, true),
('Customer Management', 'customers', 'Customer database and relationship management', 'user-circle', 10, true),
('Supplier Management', 'suppliers', 'Supplier database and procurement', 'truck', 11, true),
('Purchase Orders', 'purchase_orders', 'Purchase order processing', 'file-text', 12, true),
('Sales Orders', 'sales_orders', 'Sales order management', 'shopping-bag', 13, true),
('Reports & Analytics', 'reports', 'Business intelligence and reporting', 'bar-chart', 14, true),
('System Administration', 'admin', 'System settings and configuration', 'settings', 15, true);

-- =====================================================
-- 4. MENUS
-- =====================================================

INSERT INTO menus (module_id, parent_menu_id, name, code, route_path, icon, display_order, is_active) VALUES
-- Dashboard Module
(1, NULL, 'Overview', 'overview', '/dashboard/overview', 'home', 1, true),
(1, NULL, 'Analytics', 'analytics', '/dashboard/analytics', 'trending-up', 2, true),

-- Tenant Management Module
(2, NULL, 'Tenants', 'tenants', '/admin/tenants', 'building', 1, true),

-- Organization Setup Module
(3, NULL, 'Organizations', 'organizations', '/admin/organizations', 'briefcase', 1, true),

-- User Management Module
(4, NULL, 'Users', 'users', '/admin/users', 'users', 1, true),
(4, NULL, 'Roles & Permissions', 'roles_permissions', '/admin/roles', 'shield', 2, true),

-- Store Management Module
(5, NULL, 'Stores', 'stores', '/stores/list', 'store', 1, true),
(5, NULL, 'Storage Locations', 'storage_locations', '/stores/locations', 'map-pin', 2, true),

-- Point of Sale Module
(6, NULL, 'POS Transactions', 'pos_transactions', '/pos/transactions', 'credit-card', 1, true),
(6, NULL, 'POS Terminals', 'pos_terminals', '/pos/terminals', 'monitor', 2, true),
(6, NULL, 'POS Reports', 'pos_reports', '/pos/reports', 'file-text', 3, true),

-- Cashier Operations Module
(7, NULL, 'Cashiers', 'cashiers', '/cashiers/list', 'user-check', 1, true),
(7, NULL, 'Cashier Sessions', 'cashier_sessions', '/cashiers/sessions', 'clock', 2, true),

-- Inventory Management Module
(8, NULL, 'Stock Overview', 'stock_overview', '/inventory/overview', 'package', 1, true),
(8, NULL, 'Stock Movements', 'stock_movements', '/inventory/movements', 'arrow-right', 2, true),
(8, NULL, 'Stock Counts', 'stock_counts', '/inventory/counts', 'clipboard', 3, true),

-- Product Catalog Module
(9, NULL, 'Products', 'products', '/products/list', 'box', 1, true),
(9, NULL, 'Categories', 'categories', '/products/categories', 'grid', 2, true),
(9, NULL, 'Brands', 'brands', '/products/brands', 'tag', 3, true),
(9, NULL, 'Price Lists', 'price_lists', '/products/price-lists', 'dollar-sign', 4, true),

-- Customer Management Module
(10, NULL, 'Customers', 'customers', '/customers/list', 'user-circle', 1, true),

-- Supplier Management Module
(11, NULL, 'Suppliers', 'suppliers', '/suppliers/list', 'truck', 1, true),

-- Purchase Orders Module
(12, NULL, 'Purchase Orders', 'purchase_orders', '/purchase-orders/list', 'file-text', 1, true),

-- Sales Orders Module
(13, NULL, 'Sales Orders', 'sales_orders', '/sales-orders/list', 'shopping-bag', 1, true),

-- Reports & Analytics Module
(14, NULL, 'Sales Reports', 'sales_reports', '/reports/sales', 'bar-chart', 1, true),
(14, NULL, 'Purchase Reports', 'purchase_reports', '/reports/purchases', 'file-text', 2, true),
(14, NULL, 'Inventory Reports', 'inventory_reports', '/reports/inventory', 'package', 3, true),
(14, NULL, 'Financial Reports', 'financial_reports', '/reports/financial', 'dollar-sign', 4, true),

-- System Administration Module
(15, NULL, 'UI Modules', 'ui_modules', '/admin/ui-modules', 'layout', 1, true),
(15, NULL, 'System Settings', 'system_settings', '/admin/settings', 'settings', 2, true),
(15, NULL, 'Audit Logs', 'audit_logs', '/admin/audit-logs', 'file-text', 3, true);

-- =====================================================
-- 5. SUBMENUS
-- =====================================================

INSERT INTO submenus (menu_id, parent_submenu_id, name, code, route_path, icon, display_order, is_active) VALUES
-- Dashboard Submenus
(1, NULL, 'Admin Dashboard', 'admin_dashboard', '/dashboard/admin', 'layout', 1, true),
(1, NULL, 'Store Dashboard', 'store_dashboard', '/dashboard/store', 'store', 2, true),
(2, NULL, 'Sales Analytics', 'sales_analytics', '/dashboard/analytics/sales', 'trending-up', 1, true),
(2, NULL, 'Inventory Analytics', 'inventory_analytics', '/dashboard/analytics/inventory', 'package', 2, true),

-- Tenant Submenus
(3, NULL, 'Tenant List', 'tenant_list', '/admin/tenants/list', 'list', 1, true),
(3, NULL, 'Add Tenant', 'add_tenant', '/admin/tenants/new', 'plus', 2, true),
(3, NULL, 'Tenant Configuration', 'tenant_config', '/admin/tenants/config', 'settings', 3, true),

-- Organization Submenus
(4, NULL, 'Organization List', 'org_list', '/admin/organizations/list', 'list', 1, true),
(4, NULL, 'Add Organization', 'add_org', '/admin/organizations/new', 'plus', 2, true),

-- User Management Submenus
(5, NULL, 'User List', 'user_list', '/admin/users/list', 'list', 1, true),
(5, NULL, 'Add User', 'add_user', '/admin/users/new', 'user-plus', 2, true),
(5, NULL, 'User Activity', 'user_activity', '/admin/users/activity', 'activity', 3, true),
(6, NULL, 'Role List', 'role_list', '/admin/roles/list', 'shield', 1, true),
(6, NULL, 'Add Role', 'add_role', '/admin/roles/new', 'plus', 2, true),
(6, NULL, 'Permission Matrix', 'permission_matrix', '/admin/roles/permissions', 'grid', 3, true),

-- Store Management Submenus
(7, NULL, 'Store List', 'store_list', '/stores/list', 'list', 1, true),
(7, NULL, 'Add Store', 'add_store', '/stores/new', 'plus', 2, true),
(7, NULL, 'Store Configuration', 'store_config', '/stores/config', 'settings', 3, true),
(8, NULL, 'Location List', 'location_list', '/stores/locations/list', 'list', 1, true),
(8, NULL, 'Add Location', 'add_location', '/stores/locations/new', 'plus', 2, true),

-- POS Submenus
(9, NULL, 'Transaction List', 'transaction_list', '/pos/transactions/list', 'list', 1, true),
(9, NULL, 'Process Sale', 'process_sale', '/pos/transactions/new', 'shopping-cart', 2, true),
(9, NULL, 'Void Transaction', 'void_transaction', '/pos/transactions/void', 'x-circle', 3, true),
(10, NULL, 'Terminal List', 'terminal_list', '/pos/terminals/list', 'list', 1, true),
(10, NULL, 'Add Terminal', 'add_terminal', '/pos/terminals/new', 'plus', 2, true),
(11, NULL, 'Daily Sales Report', 'daily_sales', '/pos/reports/daily', 'calendar', 1, true),
(11, NULL, 'Cashier Performance', 'cashier_performance', '/pos/reports/cashier', 'award', 2, true),

-- Cashier Submenus
(12, NULL, 'Cashier List', 'cashier_list', '/cashiers/list', 'list', 1, true),
(12, NULL, 'Add Cashier', 'add_cashier', '/cashiers/new', 'user-plus', 2, true),
(13, NULL, 'Active Sessions', 'active_sessions', '/cashiers/sessions/active', 'clock', 1, true),
(13, NULL, 'Session History', 'session_history', '/cashiers/sessions/history', 'history', 2, true),
(13, NULL, 'Open Session', 'open_session', '/cashiers/sessions/open', 'unlock', 3, true),
(13, NULL, 'Close Session', 'close_session', '/cashiers/sessions/close', 'lock', 4, true),

-- Inventory Submenus
(14, NULL, 'Stock Levels', 'stock_levels', '/inventory/overview/levels', 'package', 1, true),
(14, NULL, 'Low Stock Alert', 'low_stock', '/inventory/overview/low-stock', 'alert-triangle', 2, true),
(15, NULL, 'Movement History', 'movement_history', '/inventory/movements/history', 'list', 1, true),
(15, NULL, 'Record Movement', 'record_movement', '/inventory/movements/new', 'arrow-right', 2, true),
(16, NULL, 'Stock Count List', 'stock_count_list', '/inventory/counts/list', 'list', 1, true),
(16, NULL, 'Create Count', 'create_count', '/inventory/counts/new', 'plus', 2, true),

-- Product Submenus
(17, NULL, 'Product List', 'product_list', '/products/list', 'list', 1, true),
(17, NULL, 'Add Product', 'add_product', '/products/new', 'plus', 2, true),
(17, NULL, 'Product Import', 'product_import', '/products/import', 'upload', 3, true),
(18, NULL, 'Category List', 'category_list', '/products/categories/list', 'list', 1, true),
(18, NULL, 'Add Category', 'add_category', '/products/categories/new', 'plus', 2, true),
(19, NULL, 'Brand List', 'brand_list', '/products/brands/list', 'list', 1, true),
(19, NULL, 'Add Brand', 'add_brand', '/products/brands/new', 'plus', 2, true),
(20, NULL, 'Price List Management', 'price_list_mgmt', '/products/price-lists/list', 'list', 1, true),
(20, NULL, 'Add Price List', 'add_price_list', '/products/price-lists/new', 'plus', 2, true),

-- Customer Submenus
(21, NULL, 'Customer List', 'customer_list', '/customers/list', 'list', 1, true),
(21, NULL, 'Add Customer', 'add_customer', '/customers/new', 'user-plus', 2, true),
(21, NULL, 'Customer History', 'customer_history', '/customers/history', 'history', 3, true),

-- Supplier Submenus
(22, NULL, 'Supplier List', 'supplier_list', '/suppliers/list', 'list', 1, true),
(22, NULL, 'Add Supplier', 'add_supplier', '/suppliers/new', 'plus', 2, true),

-- Purchase Order Submenus
(23, NULL, 'PO List', 'po_list', '/purchase-orders/list', 'list', 1, true),
(23, NULL, 'Create PO', 'create_po', '/purchase-orders/new', 'plus', 2, true),
(23, NULL, 'Approve PO', 'approve_po', '/purchase-orders/approve', 'check-circle', 3, true),

-- Sales Order Submenus
(24, NULL, 'SO List', 'so_list', '/sales-orders/list', 'list', 1, true),
(24, NULL, 'Create SO', 'create_so', '/sales-orders/new', 'plus', 2, true),

-- Report Submenus
(25, NULL, 'Daily Sales', 'daily_sales_report', '/reports/sales/daily', 'calendar', 1, true),
(25, NULL, 'Monthly Sales', 'monthly_sales_report', '/reports/sales/monthly', 'calendar', 2, true),
(25, NULL, 'Product Performance', 'product_performance', '/reports/sales/products', 'trending-up', 3, true),
(26, NULL, 'Purchase Summary', 'purchase_summary', '/reports/purchases/summary', 'file-text', 1, true),
(26, NULL, 'Supplier Analysis', 'supplier_analysis', '/reports/purchases/suppliers', 'truck', 2, true),
(27, NULL, 'Stock Valuation', 'stock_valuation', '/reports/inventory/valuation', 'dollar-sign', 1, true),
(27, NULL, 'Inventory Turnover', 'inventory_turnover', '/reports/inventory/turnover', 'refresh-cw', 2, true),
(28, NULL, 'Profit & Loss', 'profit_loss', '/reports/financial/pl', 'trending-up', 1, true),
(28, NULL, 'Discount Analysis', 'discount_analysis', '/reports/financial/discounts', 'percent', 2, true),

-- Admin Submenus
(29, NULL, 'Module List', 'module_list', '/admin/ui-modules/list', 'list', 1, true),
(29, NULL, 'Menu Management', 'menu_management', '/admin/ui-modules/menus', 'menu', 2, true),
(29, NULL, 'Permission Management', 'permission_management', '/admin/ui-modules/permissions', 'lock', 3, true),
(30, NULL, 'General Settings', 'general_settings', '/admin/settings/general', 'settings', 1, true),
(30, NULL, 'Tax Configuration', 'tax_config', '/admin/settings/tax', 'file-text', 2, true),
(31, NULL, 'View Audit Logs', 'view_audit_logs', '/admin/audit-logs/view', 'eye', 1, true);

-- =====================================================
-- 6. MODULE PERMISSIONS
-- =====================================================

INSERT INTO module_permissions (module_id, permission_id) VALUES
-- Dashboard Module
(1, (SELECT id FROM permissions WHERE code = 'dashboard:view')),
(1, (SELECT id FROM permissions WHERE code = 'dashboard:manage')),
(1, (SELECT id FROM permissions WHERE code = 'dashboard:export')),

-- Tenant Management Module
(2, (SELECT id FROM permissions WHERE code = 'tenants:view')),
(2, (SELECT id FROM permissions WHERE code = 'tenants:manage')),
(2, (SELECT id FROM permissions WHERE code = 'tenants:delete')),
(2, (SELECT id FROM permissions WHERE code = 'tenants:configure')),

-- Organization Module
(3, (SELECT id FROM permissions WHERE code = 'organizations:view')),
(3, (SELECT id FROM permissions WHERE code = 'organizations:manage')),
(3, (SELECT id FROM permissions WHERE code = 'organizations:delete')),

-- User Management Module
(4, (SELECT id FROM permissions WHERE code = 'users:view')),
(4, (SELECT id FROM permissions WHERE code = 'users:manage')),
(4, (SELECT id FROM permissions WHERE code = 'users:delete')),
(4, (SELECT id FROM permissions WHERE code = 'users:reset_password')),
(4, (SELECT id FROM permissions WHERE code = 'roles:view')),
(4, (SELECT id FROM permissions WHERE code = 'roles:manage')),
(4, (SELECT id FROM permissions WHERE code = 'roles:delete')),
(4, (SELECT id FROM permissions WHERE code = 'roles:assign')),

-- Store Management Module
(5, (SELECT id FROM permissions WHERE code = 'stores:view')),
(5, (SELECT id FROM permissions WHERE code = 'stores:manage')),
(5, (SELECT id FROM permissions WHERE code = 'stores:delete')),
(5, (SELECT id FROM permissions WHERE code = 'stores:configure')),

-- POS Module
(6, (SELECT id FROM permissions WHERE code = 'pos:view')),
(6, (SELECT id FROM permissions WHERE code = 'pos:manage')),
(6, (SELECT id FROM permissions WHERE code = 'pos:process_sales')),
(6, (SELECT id FROM permissions WHERE code = 'pos:void_transactions')),
(6, (SELECT id FROM permissions WHERE code = 'pos:apply_discounts')),
(6, (SELECT id FROM permissions WHERE code = 'pos:process_returns')),
(6, (SELECT id FROM permissions WHERE code = 'pos:view_reports')),

-- Cashier Module
(7, (SELECT id FROM permissions WHERE code = 'cashiers:view')),
(7, (SELECT id FROM permissions WHERE code = 'cashiers:manage')),
(7, (SELECT id FROM permissions WHERE code = 'cashiers:delete')),
(7, (SELECT id FROM permissions WHERE code = 'cashiers:manage_sessions')),
(7, (SELECT id FROM permissions WHERE code = 'cashiers:view_sessions')),

-- Inventory Module
(8, (SELECT id FROM permissions WHERE code = 'inventory:view')),
(8, (SELECT id FROM permissions WHERE code = 'inventory:manage')),
(8, (SELECT id FROM permissions WHERE code = 'inventory:transfer')),
(8, (SELECT id FROM permissions WHERE code = 'inventory:stock_count')),
(8, (SELECT id FROM permissions WHERE code = 'inventory:view_reports')),

-- Product Module
(9, (SELECT id FROM permissions WHERE code = 'products:view')),
(9, (SELECT id FROM permissions WHERE code = 'products:manage')),
(9, (SELECT id FROM permissions WHERE code = 'products:delete')),
(9, (SELECT id FROM permissions WHERE code = 'products:manage_pricing')),
(9, (SELECT id FROM permissions WHERE code = 'products:view_cost')),

-- Customer Module
(10, (SELECT id FROM permissions WHERE code = 'customers:view')),
(10, (SELECT id FROM permissions WHERE code = 'customers:manage')),
(10, (SELECT id FROM permissions WHERE code = 'customers:delete')),
(10, (SELECT id FROM permissions WHERE code = 'customers:view_history')),

-- Supplier Module
(11, (SELECT id FROM permissions WHERE code = 'suppliers:view')),
(11, (SELECT id FROM permissions WHERE code = 'suppliers:manage')),
(11, (SELECT id FROM permissions WHERE code = 'suppliers:delete')),

-- Purchase Order Module
(12, (SELECT id FROM permissions WHERE code = 'purchase_orders:view')),
(12, (SELECT id FROM permissions WHERE code = 'purchase_orders:manage')),
(12, (SELECT id FROM permissions WHERE code = 'purchase_orders:delete')),
(12, (SELECT id FROM permissions WHERE code = 'purchase_orders:approve')),

-- Sales Order Module
(13, (SELECT id FROM permissions WHERE code = 'sales_orders:view')),
(13, (SELECT id FROM permissions WHERE code = 'sales_orders:manage')),
(13, (SELECT id FROM permissions WHERE code = 'sales_orders:delete')),
(13, (SELECT id FROM permissions WHERE code = 'sales_orders:approve')),

-- Reports Module
(14, (SELECT id FROM permissions WHERE code = 'reports:sales')),
(14, (SELECT id FROM permissions WHERE code = 'reports:purchases')),
(14, (SELECT id FROM permissions WHERE code = 'reports:inventory')),
(14, (SELECT id FROM permissions WHERE code = 'reports:financial')),
(14, (SELECT id FROM permissions WHERE code = 'reports:export')),

-- Admin Module
(15, (SELECT id FROM permissions WHERE code = 'ui_modules:view')),
(15, (SELECT id FROM permissions WHERE code = 'ui_modules:manage')),
(15, (SELECT id FROM permissions WHERE code = 'ui_modules:delete')),
(15, (SELECT id FROM permissions WHERE code = 'permissions:view')),
(15, (SELECT id FROM permissions WHERE code = 'permissions:manage')),
(15, (SELECT id FROM permissions WHERE code = 'permissions:delete')),
(15, (SELECT id FROM permissions WHERE code = 'settings:view')),
(15, (SELECT id FROM permissions WHERE code = 'settings:manage')),
(15, (SELECT id FROM permissions WHERE code = 'settings:audit_logs'));

-- =====================================================
-- 7. MENU PERMISSIONS
-- =====================================================

INSERT INTO menu_permissions (menu_id, permission_id) 
SELECT m.id, mp.permission_id
FROM menus m
JOIN module_permissions mp ON m.module_id = mp.module_id;

-- =====================================================
-- 8. SUBMENU PERMISSIONS
-- =====================================================

-- Dashboard Submenus
INSERT INTO submenu_permissions (submenu_id, permission_id) VALUES
(1, (SELECT id FROM permissions WHERE code = 'dashboard:view')),
(1, (SELECT id FROM permissions WHERE code = 'dashboard:manage')),
(2, (SELECT id FROM permissions WHERE code = 'dashboard:view')),
(3, (SELECT id FROM permissions WHERE code = 'dashboard:view')),
(4, (SELECT id FROM permissions WHERE code = 'dashboard:view')),

-- Tenant Submenus
(5, (SELECT id FROM permissions WHERE code = 'tenants:view')),
(6, (SELECT id FROM permissions WHERE code = 'tenants:manage')),
(7, (SELECT id FROM permissions WHERE code = 'tenants:configure')),

-- Organization Submenus
(8, (SELECT id FROM permissions WHERE code = 'organizations:view')),
(9, (SELECT id FROM permissions WHERE code = 'organizations:manage')),

-- User Submenus
(10, (SELECT id FROM permissions WHERE code = 'users:view')),
(11, (SELECT id FROM permissions WHERE code = 'users:manage')),
(12, (SELECT id FROM permissions WHERE code = 'users:view')),
(13, (SELECT id FROM permissions WHERE code = 'roles:view')),
(14, (SELECT id FROM permissions WHERE code = 'roles:manage')),
(15, (SELECT id FROM permissions WHERE code = 'roles:view')),
(15, (SELECT id FROM permissions WHERE code = 'permissions:view')),

-- Store Submenus
(16, (SELECT id FROM permissions WHERE code = 'stores:view')),
(17, (SELECT id FROM permissions WHERE code = 'stores:manage')),
(18, (SELECT id FROM permissions WHERE code = 'stores:configure')),
(19, (SELECT id FROM permissions WHERE code = 'stores:view')),
(20, (SELECT id FROM permissions WHERE code = 'stores:manage')),

-- POS Submenus
(21, (SELECT id FROM permissions WHERE code = 'pos:view')),
(22, (SELECT id FROM permissions WHERE code = 'pos:process_sales')),
(23, (SELECT id FROM permissions WHERE code = 'pos:void_transactions')),
(24, (SELECT id FROM permissions WHERE code = 'pos:view')),
(25, (SELECT id FROM permissions WHERE code = 'pos:manage')),
(26, (SELECT id FROM permissions WHERE code = 'pos:view_reports')),
(27, (SELECT id FROM permissions WHERE code = 'pos:view_reports')),

-- Cashier Submenus
(28, (SELECT id FROM permissions WHERE code = 'cashiers:view')),
(29, (SELECT id FROM permissions WHERE code = 'cashiers:manage')),
(30, (SELECT id FROM permissions WHERE code = 'cashiers:view_sessions')),
(31, (SELECT id FROM permissions WHERE code = 'cashiers:view_sessions')),
(32, (SELECT id FROM permissions WHERE code = 'cashiers:manage_sessions')),
(33, (SELECT id FROM permissions WHERE code = 'cashiers:manage_sessions')),

-- Inventory Submenus
(34, (SELECT id FROM permissions WHERE code = 'inventory:view')),
(35, (SELECT id FROM permissions WHERE code = 'inventory:view')),
(36, (SELECT id FROM permissions WHERE code = 'inventory:view')),
(37, (SELECT id FROM permissions WHERE code = 'inventory:manage')),
(38, (SELECT id FROM permissions WHERE code = 'inventory:stock_count')),
(39, (SELECT id FROM permissions WHERE code = 'inventory:stock_count')),

-- Product Submenus
(40, (SELECT id FROM permissions WHERE code = 'products:view')),
(41, (SELECT id FROM permissions WHERE code = 'products:manage')),
(42, (SELECT id FROM permissions WHERE code = 'products:manage')),
(43, (SELECT id FROM permissions WHERE code = 'products:view')),
(44, (SELECT id FROM permissions WHERE code = 'products:manage')),
(45, (SELECT id FROM permissions WHERE code = 'products:view')),
(46, (SELECT id FROM permissions WHERE code = 'products:manage')),
(47, (SELECT id FROM permissions WHERE code = 'products:manage_pricing')),
(48, (SELECT id FROM permissions WHERE code = 'products:manage_pricing')),

-- Customer Submenus
(49, (SELECT id FROM permissions WHERE code = 'customers:view')),
(50, (SELECT id FROM permissions WHERE code = 'customers:manage')),
(51, (SELECT id FROM permissions WHERE code = 'customers:view_history')),

-- Supplier Submenus
(52, (SELECT id FROM permissions WHERE code = 'suppliers:view')),
(53, (SELECT id FROM permissions WHERE code = 'suppliers:manage')),

-- Purchase Order Submenus
(54, (SELECT id FROM permissions WHERE code = 'purchase_orders:view')),
(55, (SELECT id FROM permissions WHERE code = 'purchase_orders:manage')),
(56, (SELECT id FROM permissions WHERE code = 'purchase_orders:approve')),

-- Sales Order Submenus
(57, (SELECT id FROM permissions WHERE code = 'sales_orders:view')),
(58, (SELECT id FROM permissions WHERE code = 'sales_orders:manage')),

-- Report Submenus
(59, (SELECT id FROM permissions WHERE code = 'reports:sales')),
(60, (SELECT id FROM permissions WHERE code = 'reports:sales')),
(61, (SELECT id FROM permissions WHERE code = 'reports:sales')),
(62, (SELECT id FROM permissions WHERE code = 'reports:purchases')),
(63, (SELECT id FROM permissions WHERE code = 'reports:purchases')),
(64, (SELECT id FROM permissions WHERE code = 'reports:inventory')),
(65, (SELECT id FROM permissions WHERE code = 'reports:inventory')),
(66, (SELECT id FROM permissions WHERE code = 'reports:financial')),
(67, (SELECT id FROM permissions WHERE code = 'reports:financial')),

-- Admin Submenus
(68, (SELECT id FROM permissions WHERE code = 'ui_modules:view')),
(69, (SELECT id FROM permissions WHERE code = 'ui_modules:manage')),
(70, (SELECT id FROM permissions WHERE code = 'permissions:view')),
(71, (SELECT id FROM permissions WHERE code = 'settings:view')),
(72, (SELECT id FROM permissions WHERE code = 'settings:manage')),
(73, (SELECT id FROM permissions WHERE code = 'settings:audit_logs'));

-- =====================================================
-- 9. ROLES
-- =====================================================

INSERT INTO roles (name, code, description, is_system_role, is_active,metadata) VALUES
('Super Administrator', 'super_admin', 'Full system access with all permissions including tenant management', true, true,'{"scope":"all"}'),
('Owner', 'owner', 'Organization owner with full access except tenant and UI module management', true, true,'{"scope":"all"}'),
('Store Manager', 'store_manager', 'Manages store operations, inventory, sales, and staff', false, true,'{"scope":"all"}'),
('Cashier', 'cashier', 'Processes sales transactions at POS', false, true,'{"scope":"own"}'),
('Inventory Manager', 'inventory_manager', 'Manages inventory, stock counts, and transfers', false, true,'{"scope":"own"}'),
('Sales Executive', 'sales_executive', 'Manages customers and sales orders', false, true,'{"scope":"own"}'),
('Purchase Manager', 'purchase_manager', 'Manages suppliers and purchase orders', false, true,'{"scope":"own"}'),
('Accountant', 'accountant', 'Access to financial reports and analytics', false, true,'{"scope":"own"}');

-- =====================================================
-- 10. ROLE PERMISSIONS
-- =====================================================

-- Super Administrator - All permissions
INSERT INTO role_permissions (role_id, permission_id, scope)
SELECT 
    (SELECT id FROM roles WHERE code = 'super_admin'),
    id,
    'all'
FROM permissions;

-- Owner - All permissions EXCEPT tenant and UI module management
INSERT INTO role_permissions (role_id, permission_id, scope)
SELECT 
    (SELECT id FROM roles WHERE code = 'owner'),
    id,
    'all'
FROM permissions
WHERE code NOT IN (
    'tenants:view', 'tenants:manage', 'tenants:delete', 'tenants:configure',
    'ui_modules:view', 'ui_modules:manage', 'ui_modules:delete'
);

-- Store Manager Permissions
INSERT INTO role_permissions (role_id, permission_id, scope)
SELECT 
    (SELECT id FROM roles WHERE code = 'store_manager'),
    id,
    'store'
FROM permissions
WHERE code IN (
    'dashboard:view', 'dashboard:manage', 'dashboard:export',
    'users:view', 'users:manage', 'roles:view', 'roles:assign',
    'stores:view', 'stores:configure',
    'pos:view', 'pos:process_sales', 'pos:void_transactions', 'pos:apply_discounts', 
    'pos:process_returns', 'pos:view_reports', 'pos:manage',
    'cashiers:view', 'cashiers:manage', 'cashiers:manage_sessions', 'cashiers:view_sessions',
    'inventory:view', 'inventory:manage', 'inventory:transfer', 'inventory:stock_count', 'inventory:view_reports',
    'products:view', 'products:manage', 'products:manage_pricing', 'products:view_cost',
    'customers:view', 'customers:manage', 'customers:view_history',
    'suppliers:view',
    'sales_orders:view', 'sales_orders:manage', 'sales_orders:approve',
    'purchase_orders:view',
    'reports:sales', 'reports:inventory', 'reports:financial', 'reports:export'
);

-- Cashier Permissions
INSERT INTO role_permissions (role_id, permission_id, scope)
SELECT 
    (SELECT id FROM roles WHERE code = 'cashier'),
    id,
    'own'
FROM permissions
WHERE code IN (
    'dashboard:view',
    'pos:view', 'pos:process_sales', 'pos:apply_discounts', 'pos:process_returns',
    'cashiers:view_sessions', 'cashiers:manage_sessions',
    'inventory:view',
    'products:view',
    'customers:view', 'customers:manage'
);


INSERT INTO uom_packaging_templates (organization_id, name, code, is_active)
VALUES
(1, 'Beverage Standard Pattern', '1-24-12', true),
(1, 'Snack Box Pattern', '1-12-6', true),
(1, 'Warehouse Bulk Pattern', '1-50-10', true),
(1, 'Retail Small Pattern', '1-6-4', true),
(1, 'Pharma Packaging Pattern', '1-100-10', true);


INSERT INTO uom_packaging_template_levels (template_id, level_order, uom_id, multiplier)
VALUES
-- Beverage Standard Pattern
(1, 1, 1, 1),
(1, 2, 2, 24),
(1, 3, 3, 12),

-- Snack Box Pattern
(2, 1, 1, 1),
(2, 2, 2, 12),
(2, 3, 3, 6),

-- Warehouse Bulk Pattern
(3, 1, 1, 1),
(3, 2, 4, 50),
(3, 3, 5, 10),

-- Retail Small Pattern
(4, 1, 1, 1),
(4, 2, 6, 6),
(4, 3, 7, 4),

-- Pharma Packaging Pattern
(5, 1, 1, 1),
(5, 2, 8, 100),
(5, 3, 9, 10);

-- Inventory Manager Permissions
INSERT INTO role_permissions (role_id, permission_id, scope)
SELECT 
    (SELECT id FROM roles WHERE code = 'inventory_manager'),
    id,
    'all'
FROM permissions
WHERE code IN (
    'dashboard:view',
    'inventory:view', 'inventory:manage', 'inventory:transfer', 'inventory:stock_count', 'inventory:view_reports',
    'products:view', 'products:manage', 'products:view_cost',
    'stores:view',
    'purchase_orders:view', 'purchase_orders:manage',
    'suppliers:view', 'suppliers:manage',
    'reports:inventory', 'reports:purchases', 'reports:export'
);

-- Sales Executive Permissions
INSERT INTO role_permissions (role_id, permission_id, scope)
SELECT 
    (SELECT id FROM roles WHERE code = 'sales_executive'),
    id,
    'all'
FROM permissions
WHERE code IN (
    'dashboard:view',
    'customers:view', 'customers:manage', 'customers:view_history',
    'sales_orders:view', 'sales_orders:manage',
    'products:view',
    'inventory:view',
    'reports:sales', 'reports:export'
);

-- Purchase Manager Permissions
INSERT INTO role_permissions (role_id, permission_id, scope)
SELECT 
    (SELECT id FROM roles WHERE code = 'purchase_manager'),
    id,
    'all'
FROM permissions
WHERE code IN (
    'dashboard:view',
    'suppliers:view', 'suppliers:manage', 'suppliers:delete',
    'purchase_orders:view', 'purchase_orders:manage', 'purchase_orders:approve',
    'products:view', 'products:manage',
    'inventory:view',
    'reports:purchases', 'reports:inventory', 'reports:export'
);

-- Accountant Permissions
INSERT INTO role_permissions (role_id, permission_id, scope)
SELECT 
    (SELECT id FROM roles WHERE code = 'accountant'),
    id,
    'all'
FROM permissions
WHERE code IN (
    'dashboard:view', 'dashboard:export',
    'reports:sales', 'reports:purchases', 'reports:inventory', 'reports:financial', 'reports:export',
    'pos:view', 'pos:view_reports',
    'customers:view',
    'suppliers:view',
    'products:view', 'products:view_cost',
    'inventory:view'
);

-- =====================================================
-- 11. CREATE DEMO USERS
-- =====================================================

INSERT INTO users (organization_id, username, email, password_hash, first_name, last_name, employee_code, is_active) VALUES
(1, 'admin', 'admin@democorp.com', '$2a$10$dcbOmMwFBzWWMJWhSy3iW.HasRcquRCFz5nRmcIMK36V6VCxoIlCC', 'Admin', 'User', 'EMP001', true),
(1, 'owner', 'owner@democorp.com', '$2a$10$dcbOmMwFBzWWMJWhSy3iW.HasRcquRCFz5nRmcIMK36V6VCxoIlCC', 'Business', 'Owner', 'EMP002', true),
(1, 'manager', 'manager@democorp.com', '$2a$10$dcbOmMwFBzWWMJWhSy3iW.HasRcquRCFz5nRmcIMK36V6VCxoIlCC', 'Store', 'Manager', 'EMP003', true),
(1, 'cashier1', 'cashier1@democorp.com', '$2a$10$dcbOmMwFBzWWMJWhSy3iW.HasRcquRCFz5nRmcIMK36V6VCxoIlCC', 'John', 'Cashier', 'EMP004', true),
(1, 'inventory', 'inventory@democorp.com', '$2a$10$dcbOmMwFBzWWMJWhSy3iW.HasRcquRCFz5nRmcIMK36V6VCxoIlCC', 'Inventory', 'Manager', 'EMP005', true);

-- =====================================================
-- 12. ASSIGN USER ROLES
-- =====================================================

INSERT INTO user_roles (user_id, role_id) VALUES
(1, (SELECT id FROM roles WHERE code = 'super_admin')),
(2, (SELECT id FROM roles WHERE code = 'owner')),
(3, (SELECT id FROM roles WHERE code = 'store_manager')),
(4, (SELECT id FROM roles WHERE code = 'cashier')),
(5, (SELECT id FROM roles WHERE code = 'inventory_manager'));

-- =====================================================
-- 13. ASSIGN USER STORE ACCESS
-- =====================================================

INSERT INTO user_store_access (user_id, store_id, is_primary) VALUES
-- Admin has access to all stores
(1, 1, true),
(1, 2, false),
(1, 3, false),
-- Owner has access to all stores
(2, 1, true),
(2, 2, false),
(2, 3, false),
-- Manager has access to main store
(3, 1, true),
-- Cashier has access to main store
(4, 1, true),
-- Inventory manager has access to warehouse
(5, 3, true),
(5, 1, false);

-- =====================================================
-- 14. CREATE POS TERMINALS
-- =====================================================

INSERT INTO pos_terminals (store_id, terminal_code, terminal_name, device_id, is_active) VALUES
(1, 'POS-001', 'Main Counter Terminal 1', 'DEVICE-MAIN-001', true),
(1, 'POS-002', 'Main Counter Terminal 2', 'DEVICE-MAIN-002', true),
(2, 'POS-003', 'Branch Counter Terminal 1', 'DEVICE-BRANCH-001', true);

-- =====================================================
-- 15. CREATE CASHIERS
-- =====================================================

INSERT INTO cashiers (user_id, store_id, cashier_code, drawer_limit, discount_limit, is_active) VALUES
(4, 1, 'CASH-001', 5000.00, 10.00, true);

-- =====================================================
-- 16. UI SETTINGS
-- =====================================================

INSERT INTO ui_settings (submenu_id, setting_key, setting_value, description) VALUES
(10, 'table_columns', '{"columns": ["username", "email", "first_name", "last_name", "is_active", "created_at"], "default_sort": "created_at", "default_order": "desc"}'::jsonb, 'User list table configuration'),
(10, 'pagination', '{"enabled": true, "default_page_size": 25, "page_size_options": [10, 25, 50, 100]}'::jsonb, 'User list pagination settings'),
(21, 'table_columns', '{"columns": ["transaction_number", "cashier_name", "customer_name", "total_amount", "status", "transaction_date"], "default_sort": "transaction_date", "default_order": "desc"}'::jsonb, 'POS transaction list configuration'),
(34, 'alert_threshold', '{"low_stock_threshold": 10, "show_alerts": true}'::jsonb, 'Inventory alert settings'),
(40, 'display_mode', '{"view": "grid", "items_per_page": 20}'::jsonb, 'Product list display settings');

-- =====================================================
-- 17. ROLE UI CUSTOMIZATIONS
-- =====================================================

INSERT INTO role_ui_customizations (role_id, submenu_id, customization_data) VALUES
-- Super Admin Dashboard Customization
(
    (SELECT id FROM roles WHERE code = 'super_admin'),
    (SELECT id FROM submenus WHERE code = 'admin_dashboard'),
    '{
        "widgets": ["sales_overview", "inventory_status", "recent_transactions", "low_stock_alerts", "top_products"],
        "layout": "grid",
        "refresh_interval": 30
    }'::jsonb
),
-- Owner Dashboard Customization
(
    (SELECT id FROM roles WHERE code = 'owner'),
    (SELECT id FROM submenus WHERE code = 'admin_dashboard'),
    '{
        "widgets": ["sales_overview", "profit_analysis", "store_performance", "inventory_value"],
        "layout": "grid",
        "refresh_interval": 60
    }'::jsonb
),
-- Store Manager Dashboard
(
    (SELECT id FROM roles WHERE code = 'store_manager'),
    (SELECT id FROM submenus WHERE code = 'store_dashboard'),
    '{
        "widgets": ["daily_sales", "active_cashiers", "inventory_alerts", "pending_orders"],
        "layout": "list",
        "refresh_interval": 30
    }'::jsonb
),
-- Cashier POS Customization
(
    (SELECT id FROM roles WHERE code = 'cashier'),
    (SELECT id FROM submenus WHERE code = 'process_sale'),
    '{
        "quick_access_categories": true,
        "barcode_scanner": true,
        "discount_requires_approval": true,
        "max_discount_percent": 10
    }'::jsonb
);

-- =====================================================
-- 18. SUCCESS MESSAGE
-- =====================================================
DO $$ 
BEGIN
    SELECT * FROM organizations;
	    SELECT * FROM stores;
		 SELECT * FROM modules;
		 SELECT * FROM permissions;
		  SELECT * FROM roles;
		   SELECT * FROM users;
		 SELECT * FROM submenus;
		  SELECT * FROM roles;
		   SELECT * FROM pos_terminals;
		 SELECT * FROM cashiers;
		  SELECT * FROM role_ui_customizations;
SELECT * FROM ui_settings;
		  
		 
    (SELECT COUNT(*) FROM stores),
    (SELECT COUNT(*) FROM modules),
    (SELECT COUNT(*) FROM permissions),
    (SELECT COUNT(*) FROM roles),
    (SELECT COUNT(*) FROM users),
    (SELECT COUNT(*) FROM pos_terminals),
    (SELECT COUNT(*) FROM cashiers);
END $$;

-- =====================================================
-- 19. VERIFICATION QUERIES (Optional - Uncomment to run)
-- =====================================================

-- Uncomment below to see the data after initialization:

/*
SELECT 'Organizations' as entity, COUNT(*) as count FROM organizations
UNION ALL
SELECT 'Tenants', COUNT(*) FROM tenants
UNION ALL
SELECT 'Stores', COUNT(*) FROM stores
UNION ALL
SELECT 'Modules', COUNT(*) FROM modules
UNION ALL
SELECT 'Menus', COUNT(*) FROM menus
UNION ALL
SELECT 'Submenus', COUNT(*) FROM submenus
UNION ALL
SELECT 'Permissions', COUNT(*) FROM permissions
UNION ALL
SELECT 'Roles', COUNT(*) FROM roles
UNION ALL
SELECT 'Users', COUNT(*) FROM users
UNION ALL
SELECT 'Role Permissions', COUNT(*) FROM role_permissions
UNION ALL
SELECT 'POS Terminals', COUNT(*) FROM pos_terminals
UNION ALL
SELECT 'Cashiers', COUNT(*) FROM cashiers;
*/

-- =====================================================
-- END OF INITIALIZATION SCRIPT
-- =====================================================

select * from users
select * from user_roles
select * from roles

select * from stores
delete from gosse_d













-- =====================================================
-- ADDITIONAL UNITS OF MEASURE FOR PACKAGING HIERARCHY
-- =====================================================

INSERT INTO units_of_measure (code, name, uom_type, decimal_places, is_active, metadata) VALUES
('TRAY', 'Tray', 'packaging', 0, true, '{"description": "Plastic or cardboard tray"}'),
('BUNDLE', 'Bundle', 'packaging', 0, true, '{"description": "Bundle of items"}'),
('PACK', 'Pack', 'packaging', 0, true, '{"description": "Small pack"}'),
('SACK', 'Sack', 'packaging', 0, true, '{"description": "Large sack for bulk items"}'),
('PALLET', 'Pallet', 'packaging', 0, true, '{"description": "Full pallet"}'),
('CASE', 'Case', 'packaging', 0, true, '{"description": "Case or container"}')
ON CONFLICT (code) DO NOTHING;

-- =====================================================
-- PRODUCT UOM CONVERSIONS
-- Defines how different UOMs relate to each other for each product
-- =====================================================

-- =====================================================
-- DAIRY PRODUCTS UOM CONVERSIONS
-- =====================================================

-- Almarai Fresh Milk 1L
-- Retail: 1 Piece = 1 Liter
-- Wholesale: 1 Carton = 12 Pieces = 12 Liters
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'ALMARAI-MILK-FW-1L'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    12.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 12 bottles of 1L milk"}'
),
((SELECT id FROM products WHERE sku = 'ALMARAI-MILK-FW-1L'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    1.000000,
    true,
    '{"packaging_type": "bottle", "description": "1 Bottle = 1 Liter"}'
),
((SELECT id FROM products WHERE sku = 'ALMARAI-MILK-FW-1L'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    (SELECT id FROM units_of_measure WHERE code = 'ML'),
    1000.000000,
    false,
    '{"description": "1 Liter = 1000 Milliliters"}'
);

-- Almarai Low Fat Milk 1L
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'ALMARAI-MILK-LF-1L'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    12.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 12 bottles"}'
),
((SELECT id FROM products WHERE sku = 'ALMARAI-MILK-LF-1L'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    1.000000,
    true,
    '{"packaging_type": "bottle", "description": "1 Bottle = 1 Liter"}'
);

-- Nadec Milk 2L
-- 1 Carton = 6 bottles (2L each)
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'NADEC-MILK-FW-2L'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    6.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 6 bottles of 2L milk"}'
),
((SELECT id FROM products WHERE sku = 'NADEC-MILK-FW-2L'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    2.000000,
    true,
    '{"packaging_type": "bottle", "description": "1 Bottle = 2 Liters"}'
);

-- Almarai Laban 1L
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'ALMARAI-LABAN-1L'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    12.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 12 bottles"}'
),
((SELECT id FROM products WHERE sku = 'ALMARAI-LABAN-1L'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    1.000000,
    true,
    '{"packaging_type": "bottle"}'
);

-- Al-Safi Yogurt 170g
-- 1 Tray = 6 cups, 1 Carton = 4 trays = 24 cups
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'ALSAFI-YOGURT-170G'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'TRAY'),
    4.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 4 trays"}'
),
((SELECT id FROM products WHERE sku = 'ALSAFI-YOGURT-170G'),
    (SELECT id FROM units_of_measure WHERE code = 'TRAY'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    6.000000,
    true,
    '{"packaging_type": "tray", "description": "1 Tray = 6 yogurt cups"}'
),
((SELECT id FROM products WHERE sku = 'ALSAFI-YOGURT-170G'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    170.000000,
    true,
    '{"packaging_type": "cup", "description": "1 Cup = 170 grams"}'
);

-- Eggs 30 Pieces
-- 1 Tray = 30 pieces
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'EGGS-WHITE-30PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'TRAY'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    30.000000,
    true,
    '{"packaging_type": "tray", "description": "1 Tray = 30 eggs"}'
),
((SELECT id FROM products WHERE sku = 'EGGS-WHITE-30PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'TRAY'),
    12.000000,
    false,
    '{"packaging_type": "carton", "description": "1 Carton = 12 trays = 360 eggs"}'
);

-- =====================================================
-- BEVERAGES UOM CONVERSIONS
-- =====================================================

-- Coca Cola 330ml Can
-- 1 Pack = 6 cans, 1 Carton = 4 packs = 24 cans
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'COCA-COLA-330ML'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PACK'),
    4.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 4 packs = 24 cans"}'
),
((SELECT id FROM products WHERE sku = 'COCA-COLA-330ML'),
    (SELECT id FROM units_of_measure WHERE code = 'PACK'),
    (SELECT id FROM units_of_measure WHERE code = 'CAN'),
    6.000000,
    true,
    '{"packaging_type": "pack", "description": "1 Pack = 6 cans"}'
),
((SELECT id FROM products WHERE sku = 'COCA-COLA-330ML'),
    (SELECT id FROM units_of_measure WHERE code = 'CAN'),
    (SELECT id FROM units_of_measure WHERE code = 'ML'),
    330.000000,
    true,
    '{"packaging_type": "can", "description": "1 Can = 330ml"}'
);

-- Pepsi 330ml Can
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'PEPSI-330ML'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PACK'),
    4.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 4 packs = 24 cans"}'
),
((SELECT id FROM products WHERE sku = 'PEPSI-330ML'),
    (SELECT id FROM units_of_measure WHERE code = 'PACK'),
    (SELECT id FROM units_of_measure WHERE code = 'CAN'),
    6.000000,
    true,
    '{"packaging_type": "pack", "description": "1 Pack = 6 cans"}'
),
((SELECT id FROM products WHERE sku = 'PEPSI-330ML'),
    (SELECT id FROM units_of_measure WHERE code = 'CAN'),
    (SELECT id FROM units_of_measure WHERE code = 'ML'),
    330.000000,
    true,
    '{"packaging_type": "can"}'
);

-- Coca Cola 2L Bottle
-- 1 Carton = 8 bottles
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'COCA-COLA-2L'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    8.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 8 bottles of 2L"}'
),
((SELECT id FROM products WHERE sku = 'COCA-COLA-2L'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    2.000000,
    true,
    '{"packaging_type": "bottle", "description": "1 Bottle = 2 Liters"}'
);

-- Water 600ml
-- 1 Pack = 12 bottles, 1 Carton = 2 packs = 24 bottles
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'WATER-600ML'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PACK'),
    2.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 2 packs = 24 bottles"}'
),
((SELECT id FROM products WHERE sku = 'WATER-600ML'),
    (SELECT id FROM units_of_measure WHERE code = 'PACK'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    12.000000,
    true,
    '{"packaging_type": "pack", "description": "1 Pack = 12 bottles"}'
),
((SELECT id FROM products WHERE sku = 'WATER-600ML'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    (SELECT id FROM units_of_measure WHERE code = 'ML'),
    600.000000,
    true,
    '{"packaging_type": "bottle", "description": "1 Bottle = 600ml"}'
);

-- Water 1.5L
-- 1 Pack = 6 bottles, 1 Carton = 2 packs = 12 bottles
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'WATER-1.5L'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PACK'),
    2.000000,
    true,
    '{"packaging_type": "carton"}'
),
((SELECT id FROM products WHERE sku = 'WATER-1.5L'),
    (SELECT id FROM units_of_measure WHERE code = 'PACK'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    6.000000,
    true,
    '{"packaging_type": "pack", "description": "1 Pack = 6 bottles"}'
),
((SELECT id FROM products WHERE sku = 'WATER-1.5L'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    1.500000,
    true,
    '{"packaging_type": "bottle", "description": "1 Bottle = 1.5 Liters"}'
);

-- Almarai Orange Juice 1L
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'ALMARAI-ORANGE-1L'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    12.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 12 bottles"}'
),
((SELECT id FROM products WHERE sku = 'ALMARAI-ORANGE-1L'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    1.000000,
    true,
    '{"packaging_type": "bottle"}'
);

-- Almarai Mixed Fruit Juice 1L
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'ALMARAI-MIXED-1L'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    12.000000,
    true,
    '{"packaging_type": "carton"}'
),
((SELECT id FROM products WHERE sku = 'ALMARAI-MIXED-1L'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    1.000000,
    true,
    '{"packaging_type": "bottle"}'
);

-- Rabea Tea 100 Bags
-- 1 Box = 100 tea bags, 1 Carton = 12 boxes
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'RABEA-TEA-100BAG'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'BOX'),
    12.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 12 boxes"}'
),
((SELECT id FROM products WHERE sku = 'RABEA-TEA-100BAG'),
    (SELECT id FROM units_of_measure WHERE code = 'BOX'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    100.000000,
    true,
    '{"packaging_type": "box", "description": "1 Box = 100 tea bags"}'
);

-- Lipton Tea 100 Bags
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'LIPTON-TEA-100BAG'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'BOX'),
    12.000000,
    true,
    '{"packaging_type": "carton"}'
),
((SELECT id FROM products WHERE sku = 'LIPTON-TEA-100BAG'),
    (SELECT id FROM units_of_measure WHERE code = 'BOX'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    100.000000,
    true,
    '{"packaging_type": "box"}'
);

-- Nescafe Classic 200g
-- 1 Carton = 24 jars
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'NESCAFE-CLASSIC-200G'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    24.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 24 jars"}'
),
((SELECT id FROM products WHERE sku = 'NESCAFE-CLASSIC-200G'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    200.000000,
    true,
    '{"packaging_type": "jar", "description": "1 Jar = 200g"}'
);

-- Nescafe Arabian 200g
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'NESCAFE-ARABIAN-200G'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    24.000000,
    true,
    '{"packaging_type": "carton"}'
),
((SELECT id FROM products WHERE sku = 'NESCAFE-ARABIAN-200G'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    200.000000,
    true,
    '{"packaging_type": "jar"}'
);

-- Continue with part 2...

-- =====================================================
-- FOOD & GROCERIES UOM CONVERSIONS
-- =====================================================

-- Rice Basmati 5kg
-- 1 Bag = 5kg, 1 Sack = 4 bags = 20kg
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'RICE-BASMATI-5KG'),
    (SELECT id FROM units_of_measure WHERE code = 'SACK'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    4.000000,
    true,
    '{"packaging_type": "sack", "description": "1 Sack = 4 bags of 5kg each = 20kg total"}'
),
((SELECT id FROM products WHERE sku = 'RICE-BASMATI-5KG'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    5.000000,
    true,
    '{"packaging_type": "bag", "description": "1 Bag = 5kg"}'
),
((SELECT id FROM products WHERE sku = 'RICE-BASMATI-5KG'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    1000.000000,
    false,
    '{"description": "1 Kilogram = 1000 grams"}'
);

-- Rice Basmati 10kg
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'RICE-BASMATI-10KG'),
    (SELECT id FROM units_of_measure WHERE code = 'SACK'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    2.000000,
    true,
    '{"packaging_type": "sack", "description": "1 Sack = 2 bags of 10kg each"}'
),
((SELECT id FROM products WHERE sku = 'RICE-BASMATI-10KG'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    10.000000,
    true,
    '{"packaging_type": "bag", "description": "1 Bag = 10kg"}'
);

-- Sunflower Oil 1.8L
-- 1 Carton = 6 bottles
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'OIL-SUNFLOWER-1.8L'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    6.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 6 bottles"}'
),
((SELECT id FROM products WHERE sku = 'OIL-SUNFLOWER-1.8L'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    1.800000,
    true,
    '{"packaging_type": "bottle", "description": "1 Bottle = 1.8L"}'
);

-- Corn Oil 1.8L
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'OIL-CORN-1.8L'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    6.000000,
    true,
    '{"packaging_type": "carton"}'
),
((SELECT id FROM products WHERE sku = 'OIL-CORN-1.8L'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    1.800000,
    true,
    '{"packaging_type": "bottle"}'
);

-- Pasta Penne 500g
-- 1 Carton = 20 packs
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'PASTA-PENNE-500G'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PKT'),
    20.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 20 packets"}'
),
((SELECT id FROM products WHERE sku = 'PASTA-PENNE-500G'),
    (SELECT id FROM units_of_measure WHERE code = 'PKT'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    500.000000,
    true,
    '{"packaging_type": "packet", "description": "1 Packet = 500g"}'
);

-- Pasta Spaghetti 500g
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'PASTA-SPAGHETTI-500G'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PKT'),
    20.000000,
    true,
    '{"packaging_type": "carton"}'
),
((SELECT id FROM products WHERE sku = 'PASTA-SPAGHETTI-500G'),
    (SELECT id FROM units_of_measure WHERE code = 'PKT'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    500.000000,
    true,
    '{"packaging_type": "packet"}'
);

-- California Garden Beans 400g Can
-- 1 Carton = 24 cans
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'CALGARDEN-BEANS-400G'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'CAN'),
    24.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 24 cans"}'
),
((SELECT id FROM products WHERE sku = 'CALGARDEN-BEANS-400G'),
    (SELECT id FROM units_of_measure WHERE code = 'CAN'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    400.000000,
    true,
    '{"packaging_type": "can", "description": "1 Can = 400g"}'
);

-- California Garden Tuna 185g
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'CALGARDEN-TUNA-185G'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'CAN'),
    48.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 48 small cans"}'
),
((SELECT id FROM products WHERE sku = 'CALGARDEN-TUNA-185G'),
    (SELECT id FROM units_of_measure WHERE code = 'CAN'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    185.000000,
    true,
    '{"packaging_type": "can", "description": "1 Can = 185g"}'
);

-- Sugar 1kg
-- 1 Carton = 10 bags, 1 Sack = 50kg (bulk)
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'SUGAR-WHITE-1KG'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    10.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 10 bags of 1kg"}'
),
((SELECT id FROM products WHERE sku = 'SUGAR-WHITE-1KG'),
    (SELECT id FROM units_of_measure WHERE code = 'SACK'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    50.000000,
    false,
    '{"packaging_type": "sack", "description": "1 Sack = 50kg bulk sugar"}'
),
((SELECT id FROM products WHERE sku = 'SUGAR-WHITE-1KG'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    1.000000,
    true,
    '{"packaging_type": "bag", "description": "1 Bag = 1kg"}'
);

-- Salt 1kg
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'SALT-TABLE-1KG'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    20.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 20 bags"}'
),
((SELECT id FROM products WHERE sku = 'SALT-TABLE-1KG'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    1.000000,
    true,
    '{"packaging_type": "bag"}'
);

-- =====================================================
-- FROZEN FOODS UOM CONVERSIONS
-- =====================================================

-- Sunbulah French Fries 1kg
-- 1 Carton = 10 bags
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'SUNBULAH-FRIES-1KG'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    10.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 10 bags of frozen fries"}'
),
((SELECT id FROM products WHERE sku = 'SUNBULAH-FRIES-1KG'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    1.000000,
    true,
    '{"packaging_type": "bag", "description": "1 Bag = 1kg"}'
);

-- Sunbulah Mixed Vegetables 450g
-- 1 Carton = 20 bags
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'SUNBULAH-VEGETABLES-450G'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    20.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 20 bags"}'
),
((SELECT id FROM products WHERE sku = 'SUNBULAH-VEGETABLES-450G'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    450.000000,
    true,
    '{"packaging_type": "bag", "description": "1 Bag = 450g"}'
);

-- Watania Chicken 1kg
-- 1 Carton = 12 pieces (whole chickens)
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'WATANIA-CHICKEN-1KG'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    12.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 12 whole chickens"}'
),
((SELECT id FROM products WHERE sku = 'WATANIA-CHICKEN-1KG'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    1.000000,
    true,
    '{"packaging_type": "piece", "description": "1 Chicken ≈ 1kg"}'
);

-- =====================================================
-- PERSONAL CARE UOM CONVERSIONS
-- =====================================================

-- Dettol Soap 125g
-- 1 Box = 48 pieces (individual soaps)
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'DETTOL-SOAP-125G'),
    (SELECT id FROM units_of_measure WHERE code = 'BOX'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    48.000000,
    true,
    '{"packaging_type": "box", "description": "1 Box = 48 soap bars"}'
),
((SELECT id FROM products WHERE sku = 'DETTOL-SOAP-125G'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    125.000000,
    true,
    '{"packaging_type": "piece", "description": "1 Bar = 125g"}'
);

-- Dove Body Wash 500ml
-- 1 Carton = 12 bottles
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'DOVE-BODYWASH-500ML'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    12.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 12 bottles"}'
),
((SELECT id FROM products WHERE sku = 'DOVE-BODYWASH-500ML'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    (SELECT id FROM units_of_measure WHERE code = 'ML'),
    500.000000,
    true,
    '{"packaging_type": "bottle", "description": "1 Bottle = 500ml"}'
);

-- Lux Soap 120g
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'LUX-SOAP-120G'),
    (SELECT id FROM units_of_measure WHERE code = 'BOX'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    48.000000,
    true,
    '{"packaging_type": "box"}'
),
((SELECT id FROM products WHERE sku = 'LUX-SOAP-120G'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'GM'),
    120.000000,
    true,
    '{"packaging_type": "piece"}'
);

-- Palmolive Toothpaste 100ml
-- 1 Carton = 24 tubes
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'PALMOLIVE-TOOTHPASTE-100ML'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    24.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 24 tubes"}'
),
((SELECT id FROM products WHERE sku = 'PALMOLIVE-TOOTHPASTE-100ML'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'ML'),
    100.000000,
    true,
    '{"packaging_type": "tube", "description": "1 Tube = 100ml"}'
);

-- =====================================================
-- HOUSEHOLD UOM CONVERSIONS
-- =====================================================

-- Tide Powder 3kg
-- 1 Carton = 6 bags
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'TIDE-POWDER-3KG'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    6.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 6 bags of 3kg"}'
),
((SELECT id FROM products WHERE sku = 'TIDE-POWDER-3KG'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    3.000000,
    true,
    '{"packaging_type": "bag", "description": "1 Bag = 3kg"}'
);

-- Ariel Powder 2.5kg
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'ARIEL-POWDER-2.5KG'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    8.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 8 bags"}'
),
((SELECT id FROM products WHERE sku = 'ARIEL-POWDER-2.5KG'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    (SELECT id FROM units_of_measure WHERE code = 'KG'),
    2.500000,
    true,
    '{"packaging_type": "bag", "description": "1 Bag = 2.5kg"}'
);

-- Persil Liquid 3L
-- 1 Carton = 4 bottles
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'PERSIL-LIQUID-3L'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    4.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 4 bottles of 3L"}'
),
((SELECT id FROM products WHERE sku = 'PERSIL-LIQUID-3L'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    (SELECT id FROM units_of_measure WHERE code = 'LTR'),
    3.000000,
    true,
    '{"packaging_type": "bottle", "description": "1 Bottle = 3L"}'
);

-- Finish Dishwasher Tablets 40pcs
-- 1 Box = 40 tablets, 1 Carton = 6 boxes
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'FINISH-TABS-40PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'BOX'),
    6.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 6 boxes"}'
),
((SELECT id FROM products WHERE sku = 'FINISH-TABS-40PCS'),
    (SELECT id FROM units_of_measure WHERE code = 'BOX'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    40.000000,
    true,
    '{"packaging_type": "box", "description": "1 Box = 40 tablets"}'
);

-- Palmolive Dishwashing Liquid 750ml
-- 1 Carton = 12 bottles
INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata) VALUES
((SELECT id FROM products WHERE sku = 'PALMOLIVE-DISH-750ML'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    12.000000,
    true,
    '{"packaging_type": "carton", "description": "1 Carton = 12 bottles"}'
),
((SELECT id FROM products WHERE sku = 'PALMOLIVE-DISH-750ML'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    (SELECT id FROM units_of_measure WHERE code = 'ML'),
    750.000000,
    true,
    '{"packaging_type": "bottle", "description": "1 Bottle = 750ml"}'
);

-- =====================================================
-- MULTI-LEVEL PRICING WITH UOM CONVERSIONS
-- Wholesale and Bulk Pricing
-- =====================================================

-- WHOLESALE PRICES FOR MILK (Example: Almarai Milk)
-- Piece level (retail customer buying individual bottles)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active, metadata) VALUES
((SELECT id FROM products WHERE sku = 'ALMARAI-MILK-FW-1L'),
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    7.50,  -- Lower than retail (8.50)
    1,
    true,
    '{"level": "piece", "discount_percent": 11.76}'
);

-- Carton level (wholesale customer buying cartons of 12)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active, metadata) VALUES
((SELECT id FROM products WHERE sku = 'ALMARAI-MILK-FW-1L'),
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    85.00,  -- 7.08 per piece (12 pieces per carton)
    1,
    true,
    '{"level": "carton", "price_per_piece": 7.08, "discount_percent": 16.7}'
);

-- COCA COLA MULTI-LEVEL PRICING
-- Single can
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active, metadata) VALUES
((SELECT id FROM products WHERE sku = 'COCA-COLA-330ML'),
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'),
    (SELECT id FROM units_of_measure WHERE code = 'CAN'),
    1.75,  -- Retail is 2.00
    1,
    true,
    '{"level": "can", "discount_percent": 12.5}'
);

-- 6-pack
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active, metadata) VALUES
((SELECT id FROM products WHERE sku = 'COCA-COLA-330ML'),
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'),
    (SELECT id FROM units_of_measure WHERE code = 'PACK'),
    9.90,  -- 1.65 per can (6 cans)
    1,
    true,
    '{"level": "pack", "price_per_can": 1.65, "discount_percent": 17.5}'
);

-- Full carton (24 cans = 4 packs)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active, metadata) VALUES
((SELECT id FROM products WHERE sku = 'COCA-COLA-330ML'),
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    36.00,  -- 1.50 per can (24 cans)
    1,
    true,
    '{"level": "carton", "price_per_can": 1.50, "discount_percent": 25}'
);

-- WATER 600ML MULTI-LEVEL PRICING
-- Single bottle
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active, metadata) VALUES
((SELECT id FROM products WHERE sku = 'WATER-600ML'),
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'),
    (SELECT id FROM units_of_measure WHERE code = 'BTL'),
    0.85,
    1,
    true,
    '{"level": "bottle"}'
);

-- 12-pack
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active, metadata) VALUES
((SELECT id FROM products WHERE sku = 'WATER-600ML'),
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'),
    (SELECT id FROM units_of_measure WHERE code = 'PACK'),
    9.00,  -- 0.75 per bottle
    1,
    true,
    '{"level": "pack", "price_per_bottle": 0.75}'
);

-- Full carton (24 bottles)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active, metadata) VALUES
((SELECT id FROM products WHERE sku = 'WATER-600ML'),
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    16.80,  -- 0.70 per bottle
    1,
    true,
    '{"level": "carton", "price_per_bottle": 0.70}'
);

-- EGGS MULTI-LEVEL PRICING
-- Single tray (30 eggs)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active, metadata) VALUES
((SELECT id FROM products WHERE sku = 'EGGS-WHITE-30PCS'),
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'),
    (SELECT id FROM units_of_measure WHERE code = 'TRAY'),
    16.50,  -- Retail is 18.00
    1,
    true,
    '{"level": "tray", "price_per_egg": 0.55}'
);

-- Full carton (360 eggs = 12 trays)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active, metadata) VALUES
((SELECT id FROM products WHERE sku = 'EGGS-WHITE-30PCS'),
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    180.00,  -- 15.00 per tray, 0.50 per egg
    1,
    true,
    '{"level": "carton", "price_per_tray": 15.00, "price_per_egg": 0.50}'
);

-- RICE MULTI-LEVEL PRICING (5kg bags)
-- Single bag
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active, metadata) VALUES
((SELECT id FROM products WHERE sku = 'RICE-BASMATI-5KG'),
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'),
    (SELECT id FROM units_of_measure WHERE code = 'BAG'),
    42.00,
    1,
    true,
    '{"level": "bag", "price_per_kg": 8.40}'
);

-- Sack (4 bags = 20kg)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active, metadata) VALUES
((SELECT id FROM products WHERE sku = 'RICE-BASMATI-5KG'),
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'),
    (SELECT id FROM units_of_measure WHERE code = 'SACK'),
    160.00,  -- 40.00 per 5kg bag, 8.00 per kg
    1,
    true,
    '{"level": "sack", "price_per_bag": 40.00, "price_per_kg": 8.00}'
);

-- NESCAFE MULTI-LEVEL PRICING
-- Single jar (200g)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active, metadata) VALUES
((SELECT id FROM products WHERE sku = 'NESCAFE-CLASSIC-200G'),
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'),
    (SELECT id FROM units_of_measure WHERE code = 'PCS'),
    26.00,
    1,
    true,
    '{"level": "jar"}'
);

-- Full carton (24 jars)
INSERT INTO product_prices (product_id, price_list_id, uom_id, price, min_quantity, is_active, metadata) VALUES
((SELECT id FROM products WHERE sku = 'NESCAFE-CLASSIC-200G'),
    (SELECT id FROM price_lists WHERE code = 'WHOLESALE'),
    (SELECT id FROM units_of_measure WHERE code = 'CTN'),
    576.00,  -- 24.00 per jar
    1,
    true,
    '{"level": "carton", "price_per_jar": 24.00}'
);

-- =====================================================
-- HELPER FUNCTION: Calculate total quantity in base UOM
-- =====================================================

CREATE OR REPLACE FUNCTION fn_convert_uom_quantity(
    p_product_id INTEGER,
    p_from_uom_code VARCHAR,
    p_quantity NUMERIC
)
RETURNS NUMERIC AS $$
DECLARE
    v_base_uom_id INTEGER;
    v_from_uom_id INTEGER;
    v_base_quantity NUMERIC;
BEGIN
    -- Get base UOM for product
    SELECT base_uom_id INTO v_base_uom_id
    FROM products
    WHERE id = p_product_id;
    
    -- Get from UOM ID
    SELECT id INTO v_from_uom_id
    FROM units_of_measure
    WHERE code = p_from_uom_code;
    
    -- If from_uom is already base_uom, return as is
    IF v_from_uom_id = v_base_uom_id THEN
        RETURN p_quantity;
    END IF;
    
    -- Calculate conversion
    WITH RECURSIVE uom_path AS (
        -- Base case: direct conversion
        SELECT 
            from_uom_id,
            to_uom_id,
            conversion_factor,
            1 as level
        FROM product_uom_conversions
        WHERE product_id = p_product_id
            AND from_uom_id = v_from_uom_id
        
        UNION ALL
        
        -- Recursive case: chain conversions
        SELECT 
            puc.from_uom_id,
            puc.to_uom_id,
            up.conversion_factor * puc.conversion_factor,
            up.level + 1
        FROM product_uom_conversions puc
        JOIN uom_path up ON puc.from_uom_id = up.to_uom_id
        WHERE puc.product_id = p_product_id
            AND up.level < 10  -- Prevent infinite loops
    )
    SELECT p_quantity * conversion_factor INTO v_base_quantity
    FROM uom_path
    WHERE to_uom_id = v_base_uom_id
    ORDER BY level
    LIMIT 1;
    
    RETURN COALESCE(v_base_quantity, p_quantity);
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION fn_convert_uom_quantity IS 
'Converts a quantity from one UOM to the product''s base UOM.
Follows the UOM conversion chain defined in product_uom_conversions table.
Example: Convert 2 cartons of Coca Cola to number of cans
Usage: SELECT fn_convert_uom_quantity(product_id, ''CTN'', 2);';

-- =====================================================
-- VERIFICATION QUERIES
-- =====================================================


SELECT 
    p.name as product_name,
    from_uom.code as from_uom,
    to_uom.code as to_uom,
    puc.conversion_factor,
    puc.metadata->>'description' as description
FROM product_uom_conversions puc
JOIN products p ON puc.product_id = p.id
JOIN units_of_measure from_uom ON puc.from_uom_id = from_uom.id
JOIN units_of_measure to_uom ON puc.to_uom_id = to_uom.id
WHERE p.sku = 'COCA-COLA-330ML'
ORDER BY puc.id;


-- View all prices with their UOMs for a product

SELECT 
    p.name as product_name,
    pl.name as price_list,
    uom.code as uom,
    pp.price,
    pp.min_quantity,
    pp.metadata
FROM product_prices pp
JOIN products p ON pp.product_id = p.id
JOIN price_lists pl ON pp.price_list_id = pl.id
JOIN units_of_measure uom ON pp.uom_id = uom.id
WHERE p.sku = 'COCA-COLA-330ML'
ORDER BY pl.code, uom.code;


-- Calculate effective price per base unit for all UOMs
/*
WITH price_conversions AS (
    SELECT 
        p.id as product_id,
        p.name as product_name,
        p.sku,
        pl.code as price_list_code,
        uom.code as uom_code,
        pp.price,
        base_uom.code as base_uom_code,
        COALESCE(
            (SELECT conversion_factor 
             FROM product_uom_conversions 
             WHERE product_id = p.id 
               AND from_uom_id = pp.uom_id 
               AND to_uom_id = p.base_uom_id
             LIMIT 1),
            1
        ) as uom_to_base_factor
    FROM product_prices pp
    JOIN products p ON pp.product_id = p.id
    JOIN price_lists pl ON pp.price_list_id = pl.id
    JOIN units_of_measure uom ON pp.uom_id = uom.id
    JOIN units_of_measure base_uom ON p.base_uom_id = base_uom.id
    WHERE pl.code = 'WHOLESALE'
)
SELECT 
    product_name,
    sku,
    price_list_code,
    uom_code,
    price as price_per_uom,
    base_uom_code,
    uom_to_base_factor,
    ROUND(price / uom_to_base_factor, 4) as price_per_base_unit
FROM price_conversions
WHERE sku IN ('COCA-COLA-330ML', 'ALMARAI-MILK-FW-1L', 'WATER-600ML')
ORDER BY sku, price_per_base_unit;
*/
