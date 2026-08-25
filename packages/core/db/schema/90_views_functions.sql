-- =====================================================
-- TRIGGERS FOR ORDER FINANCIAL CALCULATIONS
-- =====================================================

-- Calculate order totals
CREATE OR REPLACE FUNCTION calculate_order_totals()
RETURNS TRIGGER AS $$
DECLARE
    v_subtotal DECIMAL(15,2);
    v_tax_amount DECIMAL(15,2);
BEGIN
    -- Calculate subtotal and tax from order lines
    SELECT 
        COALESCE(SUM(line_total - tax_amount), 0),
        COALESCE(SUM(tax_amount), 0)
    INTO v_subtotal, v_tax_amount
    FROM sales_order_lines_v2
    WHERE sales_order_id = COALESCE(NEW.sales_order_id, OLD.sales_order_id);
    
    -- Update the order
    UPDATE sales_orders_v2
    SET 
        subtotal = v_subtotal,
        tax_amount = v_tax_amount,
        total_amount = v_subtotal + v_tax_amount + COALESCE(shipping_amount, 0) + COALESCE(adjustment_amount, 0) - COALESCE(discount_amount, 0),
        balance_due = (v_subtotal + v_tax_amount + COALESCE(shipping_amount, 0) + COALESCE(adjustment_amount, 0) - COALESCE(discount_amount, 0)) - COALESCE(paid_amount, 0)
    WHERE id = COALESCE(NEW.sales_order_id, OLD.sales_order_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER calculate_order_totals_trigger
    AFTER INSERT OR UPDATE OR DELETE ON sales_order_lines_v2
    FOR EACH ROW
    EXECUTE FUNCTION calculate_order_totals();

-- =====================================================
-- TRIGGERS FOR INVOICE CALCULATIONS
-- =====================================================

-- Calculate invoice totals
CREATE OR REPLACE FUNCTION calculate_invoice_totals()
RETURNS TRIGGER AS $$
DECLARE
    v_subtotal DECIMAL(15,2);
    v_tax_amount DECIMAL(15,2);
BEGIN
    -- Calculate from invoice lines
    SELECT 
        COALESCE(SUM(line_total - tax_amount), 0),
        COALESCE(SUM(tax_amount), 0)
    INTO v_subtotal, v_tax_amount
    FROM invoice_lines
    WHERE invoice_id = COALESCE(NEW.invoice_id, OLD.invoice_id);
    
    -- Update the invoice
    UPDATE invoices
    SET 
        subtotal = v_subtotal,
        tax_amount = v_tax_amount,
        total_amount = v_subtotal + v_tax_amount + COALESCE(shipping_amount, 0) + COALESCE(adjustment_amount, 0) - COALESCE(discount_amount, 0),
        balance_due = (v_subtotal + v_tax_amount + COALESCE(shipping_amount, 0) + COALESCE(adjustment_amount, 0) - COALESCE(discount_amount, 0)) - COALESCE(paid_amount, 0) - COALESCE(credit_applied, 0)
    WHERE id = COALESCE(NEW.invoice_id, OLD.invoice_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER calculate_invoice_totals_trigger
    AFTER INSERT OR UPDATE OR DELETE ON invoice_lines
    FOR EACH ROW
    EXECUTE FUNCTION calculate_invoice_totals();

-- Update invoice paid amount when payment received
CREATE OR REPLACE FUNCTION update_invoice_payment()
RETURNS TRIGGER AS $$
DECLARE
    v_total_paid DECIMAL(15,2);
BEGIN
    -- Calculate total paid
    SELECT COALESCE(SUM(payment_amount), 0)
    INTO v_total_paid
    FROM invoice_payments
    WHERE invoice_id = COALESCE(NEW.invoice_id, OLD.invoice_id);
    
    -- Update invoice
    UPDATE invoices
    SET 
        paid_amount = v_total_paid,
        balance_due = total_amount - v_total_paid - COALESCE(credit_applied, 0),
        invoice_status = CASE
            WHEN v_total_paid = 0 THEN 'sent'::invoice_status
            WHEN v_total_paid >= total_amount - COALESCE(credit_applied, 0) THEN 'paid'::invoice_status
            ELSE 'partially_paid'::invoice_status
        END,
        paid_date = CASE 
            WHEN v_total_paid >= total_amount - COALESCE(credit_applied, 0) THEN CURRENT_DATE
            ELSE NULL
        END
    WHERE id = COALESCE(NEW.invoice_id, OLD.invoice_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_invoice_payment_trigger
    AFTER INSERT OR UPDATE OR DELETE ON invoice_payments
    FOR EACH ROW
    EXECUTE FUNCTION update_invoice_payment();

-- =====================================================
-- COMMENTS FOR DOCUMENTATION
-- =====================================================

COMMENT ON TABLE carts IS 'Shopping carts for online and POS channels, supporting both registered customers and guests';
COMMENT ON TABLE cart_items IS 'Line items in shopping carts with pricing and customization details';
COMMENT ON TABLE cart_activity_log IS 'Audit trail of all cart activities and changes';
COMMENT ON TABLE draft_cart_templates IS 'Saved carts and wishlists for quick reordering';
COMMENT ON TABLE sales_orders_v2 IS 'Enhanced order management with comprehensive tracking across all sales channels';
COMMENT ON TABLE sales_order_lines_v2 IS 'Order line items with fulfillment tracking';
COMMENT ON TABLE order_fulfillments IS 'Shipment and fulfillment tracking for orders';
COMMENT ON TABLE invoices IS 'Customer invoices with payment tracking and recurring billing support';
COMMENT ON TABLE invoice_payments IS 'Payment records against invoices';
COMMENT ON TABLE quotes IS 'Sales quotations with approval workflow';

-- =====================================================
-- FOREIGN KEY CONSTRAINTS (DEFERRED)
-- =====================================================

-- Sales Analytics
ALTER TABLE sales_analytics 
    ADD CONSTRAINT fk_sales_analytics_store 
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE SET NULL;

ALTER TABLE sales_analytics 
    ADD CONSTRAINT fk_sales_analytics_product 
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL;

ALTER TABLE sales_analytics 
    ADD CONSTRAINT fk_sales_analytics_category 
    FOREIGN KEY (category_id) REFERENCES product_categories(id) ON DELETE SET NULL;

ALTER TABLE sales_analytics 
    ADD CONSTRAINT fk_sales_analytics_customer 
    FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE SET NULL;

-- Purchase Analytics
ALTER TABLE purchase_analytics 
    ADD CONSTRAINT fk_purchase_analytics_store 
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE SET NULL;

ALTER TABLE purchase_analytics 
    ADD CONSTRAINT fk_purchase_analytics_supplier 
    FOREIGN KEY (supplier_id) REFERENCES business_partners(id) ON DELETE SET NULL;

ALTER TABLE purchase_analytics 
    ADD CONSTRAINT fk_purchase_analytics_product 
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL;

ALTER TABLE purchase_analytics
    ADD CONSTRAINT fk_purchase_analytics_category
    FOREIGN KEY (category_id) REFERENCES product_categories(id) ON DELETE SET NULL;

-- Inventory Analytics
ALTER TABLE inventory_analytics 
    ADD CONSTRAINT fk_inventory_analytics_store 
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE SET NULL;

ALTER TABLE inventory_analytics 
    ADD CONSTRAINT fk_inventory_analytics_product 
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL;

ALTER TABLE inventory_analytics
    ADD CONSTRAINT fk_inventory_analytics_category
    FOREIGN KEY (category_id) REFERENCES product_categories(id) ON DELETE SET NULL;

-- Profit Loss Analytics
ALTER TABLE profit_loss_analytics 
    ADD CONSTRAINT fk_profit_loss_analytics_store 
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE SET NULL;

-- Discount Analytics
ALTER TABLE discount_analytics 
    ADD CONSTRAINT fk_discount_analytics_store 
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE SET NULL;

ALTER TABLE discount_analytics 
    ADD CONSTRAINT fk_discount_analytics_cashier 
    FOREIGN KEY (cashier_id) REFERENCES cashiers(id) ON DELETE SET NULL;

ALTER TABLE discount_analytics 
    ADD CONSTRAINT fk_discount_analytics_product 
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL;

-- =====================================================
-- TRIGGERS FOR UPDATED_AT
-- =====================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $func$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$func$ LANGUAGE plpgsql;

-- Apply triggers to all tables with updated_at
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;
CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_modules_updated_at ON modules;
CREATE TRIGGER update_modules_updated_at BEFORE UPDATE ON modules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_menus_updated_at ON menus;
CREATE TRIGGER update_menus_updated_at BEFORE UPDATE ON menus FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_submenus_updated_at ON submenus;
CREATE TRIGGER update_submenus_updated_at BEFORE UPDATE ON submenus FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_ui_settings_updated_at ON ui_settings;
CREATE TRIGGER update_ui_settings_updated_at BEFORE UPDATE ON ui_settings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_role_ui_customizations_updated_at ON role_ui_customizations;
CREATE TRIGGER update_role_ui_customizations_updated_at BEFORE UPDATE ON role_ui_customizations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_stores_updated_at ON stores;
CREATE TRIGGER update_stores_updated_at BEFORE UPDATE ON stores FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_pos_terminals_updated_at ON pos_terminals;
CREATE TRIGGER update_pos_terminals_updated_at BEFORE UPDATE ON pos_terminals FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_product_categories_updated_at ON product_categories;
CREATE TRIGGER update_product_categories_updated_at BEFORE UPDATE ON product_categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_brands_updated_at ON brands;
CREATE TRIGGER update_brands_updated_at BEFORE UPDATE ON brands FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_price_lists_updated_at ON price_lists;
CREATE TRIGGER update_price_lists_updated_at BEFORE UPDATE ON price_lists FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_product_variants_updated_at ON product_variants;
CREATE TRIGGER update_product_variants_updated_at BEFORE UPDATE ON product_variants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_product_prices_updated_at ON product_prices;
CREATE TRIGGER update_product_prices_updated_at BEFORE UPDATE ON product_prices FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_product_serial_numbers_updated_at ON product_serial_numbers;
CREATE TRIGGER update_product_serial_numbers_updated_at BEFORE UPDATE ON product_serial_numbers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_product_batches_updated_at ON product_batches;
CREATE TRIGGER update_product_batches_updated_at BEFORE UPDATE ON product_batches FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_inventory_stock_updated_at ON inventory_stock;
CREATE TRIGGER update_inventory_stock_updated_at BEFORE UPDATE ON inventory_stock FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_business_partners_updated_at ON business_partners;
CREATE TRIGGER update_business_partners_updated_at BEFORE UPDATE ON business_partners FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_customers_updated_at ON customers;
CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_purchase_orders_updated_at ON purchase_orders;
CREATE TRIGGER update_purchase_orders_updated_at BEFORE UPDATE ON purchase_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_transfer_requests_updated_at ON transfer_requests;
CREATE TRIGGER update_transfer_requests_updated_at BEFORE UPDATE ON transfer_requests FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_goods_receipt_notes_updated_at ON goods_receipt_notes;
CREATE TRIGGER update_goods_receipt_notes_updated_at BEFORE UPDATE ON goods_receipt_notes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_sales_orders_updated_at ON sales_orders;
CREATE TRIGGER update_sales_orders_updated_at BEFORE UPDATE ON sales_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_sales_analytics_updated_at ON sales_analytics;
CREATE TRIGGER update_sales_analytics_updated_at BEFORE UPDATE ON sales_analytics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_purchase_analytics_updated_at ON purchase_analytics;
CREATE TRIGGER update_purchase_analytics_updated_at BEFORE UPDATE ON purchase_analytics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_inventory_analytics_updated_at ON inventory_analytics;
CREATE TRIGGER update_inventory_analytics_updated_at BEFORE UPDATE ON inventory_analytics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_profit_loss_analytics_updated_at ON profit_loss_analytics;
CREATE TRIGGER update_profit_loss_analytics_updated_at BEFORE UPDATE ON profit_loss_analytics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_discount_analytics_updated_at ON discount_analytics;
CREATE TRIGGER update_discount_analytics_updated_at BEFORE UPDATE ON discount_analytics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS trg_business_partners_updated_at ON business_partners;
CREATE TRIGGER trg_business_partners_updated_at BEFORE UPDATE ON business_partners FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS trg_bp_price_contracts_updated_at ON bp_price_contracts;
CREATE TRIGGER trg_bp_price_contracts_updated_at BEFORE UPDATE ON bp_price_contracts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Restaurant triggers
DROP TRIGGER IF EXISTS trg_restaurant_tables_updated_at ON restaurant_tables;
CREATE TRIGGER trg_restaurant_tables_updated_at BEFORE UPDATE ON restaurant_tables FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS trg_menu_categories_updated_at ON menu_categories;
CREATE TRIGGER trg_menu_categories_updated_at BEFORE UPDATE ON menu_categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS trg_menu_items_updated_at ON menu_items;
CREATE TRIGGER trg_menu_items_updated_at BEFORE UPDATE ON menu_items FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS trg_recipes_updated_at ON recipes;
CREATE TRIGGER trg_recipes_updated_at BEFORE UPDATE ON recipes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS trg_restaurant_orders_updated_at ON restaurant_orders;
CREATE TRIGGER trg_restaurant_orders_updated_at BEFORE UPDATE ON restaurant_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS trg_restaurant_order_items_updated_at ON restaurant_order_items;
CREATE TRIGGER trg_restaurant_order_items_updated_at BEFORE UPDATE ON restaurant_order_items FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- INDEXES FOR PERFORMANCE
-- =====================================================

-- Organizations
CREATE INDEX idx_organizations_code ON organizations(code);
CREATE INDEX idx_organizations_is_active ON organizations(is_active);

-- Tenants
CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_is_active ON tenants(is_active);

-- Modules
CREATE INDEX idx_modules_code ON modules(code);
CREATE INDEX idx_modules_is_active ON modules(is_active);
CREATE INDEX idx_modules_display_order ON modules(display_order);

-- Menus
CREATE INDEX idx_menus_module_id ON menus(module_id);
CREATE INDEX idx_menus_parent_menu_id ON menus(parent_menu_id);
CREATE INDEX idx_menus_is_active ON menus(is_active);
CREATE INDEX idx_menus_display_order ON menus(display_order);

-- Submenus
CREATE INDEX idx_submenus_menu_id ON submenus(menu_id);
CREATE INDEX idx_submenus_parent_submenu_id ON submenus(parent_submenu_id);
CREATE INDEX idx_submenus_is_active ON submenus(is_active);
CREATE INDEX idx_submenus_display_order ON submenus(display_order);

-- Permissions
CREATE INDEX idx_permissions_code ON permissions(code);

-- Roles
CREATE INDEX idx_roles_code ON roles(code);
CREATE INDEX idx_roles_is_active ON roles(is_active);

-- Role Permissions
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- Stores
CREATE INDEX idx_stores_organization_id ON stores(organization_id);
CREATE INDEX idx_stores_parent_store_id ON stores(parent_store_id);
CREATE INDEX idx_stores_code ON stores(code);
CREATE INDEX idx_stores_is_active ON stores(is_active);
CREATE INDEX idx_stores_store_type ON stores(store_type);

-- Storage Locations
CREATE INDEX idx_storage_locations_store_id ON storage_locations(store_id);
CREATE INDEX idx_storage_locations_parent_location_id ON storage_locations(parent_location_id);
CREATE INDEX idx_storage_locations_code ON storage_locations(code);

-- Users
CREATE INDEX idx_users_organization_id ON users(organization_id);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_employee_code ON users(employee_code);
CREATE INDEX idx_users_is_active ON users(is_active);

-- User Roles
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- User Store Access
CREATE INDEX idx_user_store_access_user_id ON user_store_access(user_id);
CREATE INDEX idx_user_store_access_store_id ON user_store_access(store_id);

-- Cashiers
CREATE INDEX idx_cashiers_user_id ON cashiers(user_id);
CREATE INDEX idx_cashiers_store_id ON cashiers(store_id);
CREATE INDEX idx_cashiers_is_active ON cashiers(is_active);

-- POS Terminals
CREATE INDEX idx_pos_terminals_store_id ON pos_terminals(store_id);
CREATE INDEX idx_pos_terminals_is_active ON pos_terminals(is_active);

-- Cashier Sessions
CREATE INDEX idx_cashier_sessions_cashier_id ON cashier_sessions(cashier_id);
CREATE INDEX idx_cashier_sessions_pos_terminal_id ON cashier_sessions(pos_terminal_id);
CREATE INDEX idx_cashier_sessions_status ON cashier_sessions(status);
CREATE INDEX idx_cashier_sessions_opening_time ON cashier_sessions(opening_time);

-- Product Categories
CREATE INDEX idx_product_categories_parent_category_id ON product_categories(parent_category_id);
CREATE INDEX idx_product_categories_code ON product_categories(code);
CREATE INDEX idx_product_categories_is_active ON product_categories(is_active);

-- Brands
CREATE INDEX idx_brands_code ON brands(code);
CREATE INDEX idx_brands_is_active ON brands(is_active);

-- Units of Measure
CREATE INDEX idx_units_of_measure_code ON units_of_measure(code);
CREATE INDEX idx_units_of_measure_uom_type ON units_of_measure(uom_type);

-- UOM Packaging Templates
CREATE INDEX idx_uom_packaging_templates_organization_id ON uom_packaging_templates(organization_id);
CREATE INDEX idx_uom_packaging_templates_code ON uom_packaging_templates(code);
CREATE INDEX idx_uom_pkg_template_levels_template_id ON uom_packaging_template_levels(template_id);
CREATE INDEX idx_uom_pkg_template_levels_uom_id ON uom_packaging_template_levels(uom_id);

-- Price Lists
CREATE INDEX idx_price_lists_code ON price_lists(code);
CREATE INDEX idx_price_lists_is_active ON price_lists(is_active);
CREATE INDEX idx_price_lists_valid_from ON price_lists(valid_from);
CREATE INDEX idx_price_lists_valid_to ON price_lists(valid_to);

-- Tax Categories
CREATE INDEX idx_tax_categories_code ON tax_categories(code);
CREATE INDEX idx_tax_categories_is_active ON tax_categories(is_active);

-- Products
CREATE INDEX idx_products_organization_id ON products(organization_id);
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_brand_id ON products(brand_id);
CREATE INDEX idx_products_is_active ON products(is_active);
CREATE INDEX idx_products_is_sellable ON products(is_sellable);
CREATE INDEX idx_products_is_purchasable ON products(is_purchasable);
CREATE INDEX idx_products_product_type ON products(product_type);

-- Product Variants
CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX idx_product_variants_variant_sku ON product_variants(variant_sku);
CREATE INDEX idx_product_variants_is_active ON product_variants(is_active);

-- Product Barcodes
CREATE INDEX idx_product_barcodes_product_id ON product_barcodes(product_id);
CREATE INDEX idx_product_barcodes_product_variant_id ON product_barcodes(product_variant_id);
CREATE INDEX idx_product_barcodes_barcode ON product_barcodes(barcode);

-- Product Prices
CREATE INDEX idx_product_prices_product_id ON product_prices(product_id);
CREATE INDEX idx_product_prices_product_variant_id ON product_prices(product_variant_id);
CREATE INDEX idx_product_prices_price_list_id ON product_prices(price_list_id);
CREATE INDEX idx_product_prices_is_active ON product_prices(is_active);

-- Product Serial Numbers
CREATE INDEX idx_product_serial_numbers_product_id ON product_serial_numbers(product_id);
CREATE INDEX idx_product_serial_numbers_serial_number ON product_serial_numbers(serial_number);
CREATE INDEX idx_product_serial_numbers_status ON product_serial_numbers(status);
CREATE INDEX idx_product_serial_numbers_current_store_id ON product_serial_numbers(current_store_id);

-- Product Batches
CREATE INDEX idx_product_batches_product_id ON product_batches(product_id);
CREATE INDEX idx_product_batches_batch_number ON product_batches(batch_number);
CREATE INDEX idx_product_batches_store_id ON product_batches(store_id);
CREATE INDEX idx_product_batches_status ON product_batches(status);
CREATE INDEX idx_product_batches_expiry_date ON product_batches(expiry_date);

-- Inventory Stock
CREATE INDEX idx_inventory_stock_product_id ON inventory_stock(product_id);
CREATE INDEX idx_inventory_stock_product_variant_id ON inventory_stock(product_variant_id);
CREATE INDEX idx_inventory_stock_store_id ON inventory_stock(store_id);
CREATE INDEX idx_inventory_stock_storage_location_id ON inventory_stock(storage_location_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_inventory_stock_unique_product_variant_store ON inventory_stock(product_id, COALESCE(product_variant_id, -1), store_id);

-- Stock Movements
CREATE INDEX idx_stock_movements_product_id ON stock_movements(product_id);
CREATE INDEX idx_stock_movements_from_store_id ON stock_movements(from_store_id);
CREATE INDEX idx_stock_movements_to_store_id ON stock_movements(to_store_id);
CREATE INDEX idx_stock_movements_movement_type ON stock_movements(movement_type);
CREATE INDEX idx_stock_movements_movement_date ON stock_movements(movement_date);
CREATE INDEX idx_stock_movements_reference_type_id ON stock_movements(reference_type, reference_id);

-- Stock Counts
CREATE INDEX idx_stock_counts_store_id ON stock_counts(store_id);
CREATE INDEX idx_stock_counts_status ON stock_counts(status);
CREATE INDEX idx_stock_counts_count_number ON stock_counts(count_number);

-- Stock Count Lines
CREATE INDEX idx_stock_count_lines_stock_count_id ON stock_count_lines(stock_count_id);
CREATE INDEX idx_stock_count_lines_product_id ON stock_count_lines(product_id);

-- Suppliers
-- Indexes for suppliers table removed (replaced by business_partners in 50_purchasing_suppliers.sql)

-- Customers
CREATE INDEX idx_customers_organization_id ON customers(organization_id);
CREATE INDEX idx_customers_customer_code ON customers(customer_code);
CREATE INDEX idx_customers_is_active ON customers(is_active);
CREATE INDEX idx_customers_customer_type ON customers(customer_type);

-- Purchase Orders
CREATE INDEX idx_purchase_orders_organization_id ON purchase_orders(organization_id);
CREATE INDEX idx_purchase_orders_partners_id ON purchase_orders(partners_id);
CREATE INDEX idx_purchase_orders_store_id ON purchase_orders(store_id);
CREATE INDEX idx_purchase_orders_po_number ON purchase_orders(po_number);
CREATE INDEX idx_purchase_orders_status ON purchase_orders(status);
CREATE INDEX idx_purchase_orders_po_date ON purchase_orders(po_date);

-- Purchase Order Lines
CREATE INDEX idx_purchase_order_lines_purchase_order_id ON purchase_order_lines(purchase_order_id);
CREATE INDEX idx_purchase_order_lines_product_id ON purchase_order_lines(product_id);

-- Sales Orders
CREATE INDEX idx_sales_orders_organization_id ON sales_orders(organization_id);
CREATE INDEX idx_sales_orders_customer_id ON sales_orders(customer_id);
CREATE INDEX idx_sales_orders_store_id ON sales_orders(store_id);
CREATE INDEX idx_sales_orders_order_number ON sales_orders(order_number);
CREATE INDEX idx_sales_orders_status ON sales_orders(status);
CREATE INDEX idx_sales_orders_order_date ON sales_orders(order_date);

-- Sales Order Lines
CREATE INDEX idx_sales_order_lines_sales_order_id ON sales_order_lines(sales_order_id);
CREATE INDEX idx_sales_order_lines_product_id ON sales_order_lines(product_id);

-- POS Transactions
CREATE INDEX idx_pos_transactions_store_id ON pos_transactions(store_id);
CREATE INDEX idx_pos_transactions_cashier_id ON pos_transactions(cashier_id);
CREATE INDEX idx_pos_transactions_cashier_session_id ON pos_transactions(cashier_session_id);
CREATE INDEX idx_pos_transactions_customer_id ON pos_transactions(customer_id);
CREATE INDEX idx_pos_transactions_transaction_number ON pos_transactions(transaction_number);
CREATE INDEX idx_pos_transactions_transaction_date ON pos_transactions(transaction_date);
CREATE INDEX idx_pos_transactions_status ON pos_transactions(status);

-- POS Transaction Lines
CREATE INDEX idx_pos_transaction_lines_transaction_id ON pos_transaction_lines(transaction_id);
CREATE INDEX idx_pos_transaction_lines_product_id ON pos_transaction_lines(product_id);

-- FIX #7: returns indexes
CREATE INDEX idx_sales_returns_store_id              ON sales_returns(store_id);
CREATE INDEX idx_sales_returns_original_transaction  ON sales_returns(original_transaction_id);
CREATE INDEX idx_sales_returns_status                ON sales_returns(status);
CREATE INDEX idx_sales_return_lines_return_id        ON sales_return_lines(return_id);

-- POS Payments
CREATE INDEX idx_pos_payments_transaction_id ON pos_payments(transaction_id);
CREATE INDEX idx_pos_payments_payment_method ON pos_payments(payment_method);

-- Sales Analytics
CREATE INDEX idx_sales_analytics_organization_id ON sales_analytics(organization_id);
CREATE INDEX idx_sales_analytics_store_id ON sales_analytics(store_id);
CREATE INDEX idx_sales_analytics_product_id ON sales_analytics(product_id);
CREATE INDEX idx_sales_analytics_category_id ON sales_analytics(category_id);
CREATE INDEX idx_sales_analytics_customer_id ON sales_analytics(customer_id);
CREATE INDEX idx_sales_analytics_date ON sales_analytics(date);
CREATE INDEX idx_sales_analytics_year_month ON sales_analytics(year, month);

-- Purchase Analytics
CREATE INDEX idx_purchase_analytics_organization_id ON purchase_analytics(organization_id);
CREATE INDEX idx_purchase_analytics_store_id ON purchase_analytics(store_id);
CREATE INDEX idx_purchase_analytics_supplier_id ON purchase_analytics(supplier_id);
CREATE INDEX idx_purchase_analytics_product_id ON purchase_analytics(product_id);
CREATE INDEX idx_purchase_analytics_date ON purchase_analytics(date);

-- Inventory Analytics
CREATE INDEX idx_inventory_analytics_organization_id ON inventory_analytics(organization_id);
CREATE INDEX idx_inventory_analytics_store_id ON inventory_analytics(store_id);
CREATE INDEX idx_inventory_analytics_product_id ON inventory_analytics(product_id);
CREATE INDEX idx_inventory_analytics_date ON inventory_analytics(date);

-- FIX #1: stock reservations indexes
CREATE INDEX idx_stock_reservations_product_id         ON stock_reservations(product_id);
CREATE INDEX idx_stock_reservations_product_variant_id ON stock_reservations(product_variant_id);
CREATE INDEX idx_stock_reservations_store_id           ON stock_reservations(store_id);
CREATE INDEX idx_stock_reservations_reference          ON stock_reservations(reference_type, reference_id);
CREATE INDEX idx_stock_reservations_status             ON stock_reservations(status);
CREATE INDEX idx_stock_reservations_expires_at         ON stock_reservations(expires_at);

-- Profit Loss Analytics
CREATE INDEX idx_profit_loss_analytics_organization_id ON profit_loss_analytics(organization_id);
CREATE INDEX idx_profit_loss_analytics_store_id ON profit_loss_analytics(store_id);
CREATE INDEX idx_profit_loss_analytics_date ON profit_loss_analytics(date);
CREATE INDEX idx_profit_loss_analytics_period_type ON profit_loss_analytics(period_type);

-- Discount Analytics
CREATE INDEX idx_discount_analytics_organization_id ON discount_analytics(organization_id);
CREATE INDEX idx_discount_analytics_store_id ON discount_analytics(store_id);
CREATE INDEX idx_discount_analytics_cashier_id ON discount_analytics(cashier_id);
CREATE INDEX idx_discount_analytics_date ON discount_analytics(date);

-- Restaurant Module Indexes
CREATE INDEX idx_restaurant_tables_store_id         ON restaurant_tables(store_id);
CREATE INDEX idx_restaurant_tables_is_active        ON restaurant_tables(is_active);
CREATE INDEX idx_restaurant_tables_section          ON restaurant_tables(section);

CREATE INDEX idx_menu_categories_store_id           ON menu_categories(store_id);
CREATE INDEX idx_menu_categories_parent_id          ON menu_categories(parent_category_id);
CREATE INDEX idx_menu_categories_is_active          ON menu_categories(is_active);
CREATE INDEX idx_menu_categories_display_order      ON menu_categories(display_order);

CREATE INDEX idx_menu_items_store_id                ON menu_items(store_id);
CREATE INDEX idx_menu_items_category_id             ON menu_items(menu_category_id);
CREATE INDEX idx_menu_items_product_id              ON menu_items(product_id);
CREATE INDEX idx_menu_items_recipe_id               ON menu_items(recipe_id);
CREATE INDEX idx_menu_items_is_active               ON menu_items(is_active);
CREATE INDEX idx_menu_items_is_available            ON menu_items(is_available);
CREATE INDEX idx_menu_items_display_order           ON menu_items(display_order);

CREATE INDEX idx_menu_item_modifiers_item_id        ON menu_item_modifiers(menu_item_id);
CREATE INDEX idx_menu_item_modifiers_is_active      ON menu_item_modifiers(is_active);

CREATE INDEX idx_recipes_organization_id            ON recipes(organization_id);
CREATE INDEX idx_recipes_finished_product_id        ON recipes(finished_product_id);
CREATE INDEX idx_recipes_is_active                  ON recipes(is_active);
CREATE INDEX idx_recipes_code                       ON recipes(recipe_code);

CREATE INDEX idx_recipe_ingredients_recipe_id       ON recipe_ingredients(recipe_id);
CREATE INDEX idx_recipe_ingredients_product_id      ON recipe_ingredients(product_id);
CREATE INDEX idx_recipe_ingredients_variant_id      ON recipe_ingredients(product_variant_id);

CREATE INDEX idx_restaurant_orders_store_id         ON restaurant_orders(store_id);
CREATE INDEX idx_restaurant_orders_table_id         ON restaurant_orders(table_id);
CREATE INDEX idx_restaurant_orders_cashier_id       ON restaurant_orders(cashier_id);
CREATE INDEX idx_restaurant_orders_session_id       ON restaurant_orders(cashier_session_id);
CREATE INDEX idx_restaurant_orders_customer_id      ON restaurant_orders(customer_id);
CREATE INDEX idx_restaurant_orders_status           ON restaurant_orders(status);
CREATE INDEX idx_restaurant_orders_source           ON restaurant_orders(order_source);
CREATE INDEX idx_restaurant_orders_ordered_at       ON restaurant_orders(ordered_at);
CREATE INDEX idx_restaurant_orders_pos_txn_id       ON restaurant_orders(pos_transaction_id);
CREATE INDEX idx_restaurant_orders_store_status_time ON restaurant_orders(store_id, status, ordered_at);

CREATE INDEX idx_restaurant_order_items_order_id    ON restaurant_order_items(order_id);
CREATE INDEX idx_restaurant_order_items_menu_item   ON restaurant_order_items(menu_item_id);
CREATE INDEX idx_restaurant_order_items_status      ON restaurant_order_items(status);

CREATE INDEX idx_waste_logs_store_id                ON waste_logs(store_id);
CREATE INDEX idx_waste_logs_product_id              ON waste_logs(product_id);
CREATE INDEX idx_waste_logs_menu_item_id            ON waste_logs(menu_item_id);
CREATE INDEX idx_waste_logs_recipe_id               ON waste_logs(recipe_id);
CREATE INDEX idx_waste_logs_waste_source            ON waste_logs(waste_source);
CREATE INDEX idx_waste_logs_wasted_at               ON waste_logs(wasted_at);
CREATE INDEX idx_waste_logs_order_id                ON waste_logs(order_id);
CREATE INDEX idx_waste_logs_store_source_date       ON waste_logs(store_id, waste_source, wasted_at);

CREATE INDEX idx_kiosk_sessions_terminal_id         ON kiosk_sessions(pos_terminal_id);
CREATE INDEX idx_kiosk_sessions_store_id            ON kiosk_sessions(store_id);
CREATE INDEX idx_kiosk_sessions_status              ON kiosk_sessions(status);
CREATE INDEX idx_kiosk_sessions_token               ON kiosk_sessions(session_token);


-- FIX #11 index
-- CREATE INDEX idx_menu_availability_menu_item_id ON menu_item_availability_schedules(menu_item_id);
-- CREATE INDEX idx_menu_availability_day          ON menu_item_availability_schedules(day_of_week);
-- CREATE INDEX idx_recipes_organization_id        ON recipes(organization_id);
-- CREATE INDEX idx_recipes_finished_product_id    ON recipes(finished_product_id);
-- CREATE INDEX idx_recipes_is_active              ON recipes(is_active);
-- CREATE INDEX idx_recipes_code                   ON recipes(recipe_code);
-- CREATE INDEX idx_recipe_ingredients_recipe_id   ON recipe_ingredients(recipe_id);
-- CREATE INDEX idx_recipe_ingredients_product_id  ON recipe_ingredients(product_id);
-- CREATE INDEX idx_restaurant_orders_store_id     ON restaurant_orders(store_id);
-- CREATE INDEX idx_restaurant_orders_table_id     ON restaurant_orders(table_id);
-- CREATE INDEX idx_restaurant_orders_cashier_id   ON restaurant_orders(cashier_id);
-- CREATE INDEX idx_restaurant_orders_status       ON restaurant_orders(status);
-- CREATE INDEX idx_restaurant_orders_ordered_at   ON restaurant_orders(ordered_at);
-- CREATE INDEX idx_restaurant_order_items_order_id ON restaurant_order_items(order_id);
-- CREATE INDEX idx_restaurant_order_items_menu_item ON restaurant_order_items(menu_item_id);
-- CREATE INDEX idx_restaurant_order_items_status   ON restaurant_order_items(status);
-- FIX #13 indexes
CREATE INDEX idx_combo_bundles_store_id  ON combo_bundles(store_id);
CREATE INDEX idx_combo_bundles_is_active ON combo_bundles(is_active);
CREATE INDEX idx_combo_bundle_items_bundle_id ON combo_bundle_items(combo_bundle_id);
-- CREATE INDEX idx_waste_logs_store_id     ON waste_logs(store_id);
-- CREATE INDEX idx_waste_logs_wasted_at    ON waste_logs(wasted_at);
-- CREATE INDEX idx_kiosk_sessions_token    ON kiosk_sessions(session_token);
-- Additional POS Indexes
CREATE INDEX IF NOT EXISTS idx_product_barcodes_barcode_lookup 
ON product_barcodes(barcode) WHERE is_primary = true;

CREATE INDEX IF NOT EXISTS idx_products_sku_varchar_pattern 
ON products(sku varchar_pattern_ops);

CREATE INDEX IF NOT EXISTS idx_inventory_stock_store_product_qty 
ON inventory_stock(store_id, product_id, quantity_available);

CREATE INDEX IF NOT EXISTS idx_products_active_sellable 
ON products(is_active, is_sellable) WHERE is_active = true AND is_sellable = true;

-- =====================================================
-- POS VIEWS AND FUNCTIONS (with Type Fixes)
-- =====================================================

CREATE OR REPLACE VIEW vw_pos_product_catalog AS
SELECT
    -- Identity
    p.id                            AS product_id,
    pv.id                           AS product_variant_id,
    p.sku                           AS base_sku,
    COALESCE(pv.variant_sku, p.sku) AS sku,
    p.name                          AS base_product_name,
    COALESCE(pv.variant_name, p.name) AS product_name,
    pv.variant_attributes,
    p.description,
    p.product_type,
    -- Category
    pc.id                           AS category_id,
    pc.name                         AS category_name,
    pc.code                         AS category_code,
    pc_parent.id                    AS parent_category_id,
    pc_parent.name                  AS parent_category_name,
    -- Brand
    b.id                            AS brand_id,
    b.name                          AS brand_name,
    -- UOM
    uom.id                          AS uom_id,
    uom.code                        AS uom_code,
    uom.name                        AS uom_name,
    uom.decimal_places,
    -- Barcode: prefer variant barcode, fall back to base product barcode
    COALESCE(pb_variant.barcode, pb_base.barcode) AS barcode,
    COALESCE(pb_variant.barcode_type, pb_base.barcode_type) AS barcode_type,
    -- Tax
    tc.id                           AS tax_category_id,
    tc.name                         AS tax_category_name,
    tc.tax_rate,
    tc.is_inclusive                 AS tax_is_inclusive,
    -- Prices: prefer variant-level prices, fall back to product-level
    COALESCE(pp_retail_v.price, pp_retail.price)   AS retail_price,
    COALESCE(pp_retail_v.id,    pp_retail.id)       AS retail_price_id,
    COALESCE(pp_promo_v.price,  pp_promo.price)     AS promo_price,
    COALESCE(pp_promo_v.id,     pp_promo.id)        AS promo_price_id,
    COALESCE(pp_promo_v.min_quantity, pp_promo.min_quantity) AS promo_min_quantity,
    COALESCE(pp_promo_v.valid_from,   pp_promo.valid_from)   AS promo_valid_from,
    COALESCE(pp_promo_v.valid_to,     pp_promo.valid_to)     AS promo_valid_to,
    COALESCE(pp_promo_v.metadata->>'promotion_name', pp_promo.metadata->>'promotion_name') AS promotion_name,
    COALESCE(pp_promo_v.metadata->>'discount_percent', pp_promo.metadata->>'discount_percent') AS discount_percent,
    -- Effective price
    CASE
        WHEN COALESCE(pp_promo_v.id, pp_promo.id) IS NOT NULL
             AND COALESCE(pp_promo_v.is_active, pp_promo.is_active) = true
             AND COALESCE(pp_promo_v.valid_from, pp_promo.valid_from) <= CURRENT_DATE
             AND (COALESCE(pp_promo_v.valid_to, pp_promo.valid_to) IS NULL
                  OR COALESCE(pp_promo_v.valid_to, pp_promo.valid_to) >= CURRENT_DATE)
        THEN COALESCE(pp_promo_v.price, pp_promo.price)
        ELSE COALESCE(pp_retail_v.price, pp_retail.price)
    END AS effective_price,
    -- Active promotion flag
    CASE
        WHEN COALESCE(pp_promo_v.id, pp_promo.id) IS NOT NULL
             AND COALESCE(pp_promo_v.is_active, pp_promo.is_active) = true
             AND COALESCE(pp_promo_v.valid_from, pp_promo.valid_from) <= CURRENT_DATE
             AND (COALESCE(pp_promo_v.valid_to, pp_promo.valid_to) IS NULL
                  OR COALESCE(pp_promo_v.valid_to, pp_promo.valid_to) >= CURRENT_DATE)
        THEN true
        ELSE false
    END AS has_active_promotion,
    p.is_active,
    p.is_sellable,
    p.is_serialized,
    p.is_batch_managed,
    p.allow_decimal_quantity,
    p.track_inventory,
    p.metadata AS product_metadata
FROM products p
LEFT JOIN product_variants      pv           ON pv.product_id = p.id AND pv.is_active = true
LEFT JOIN product_categories    pc           ON p.category_id = pc.id
LEFT JOIN product_categories    pc_parent    ON pc.parent_category_id = pc_parent.id
LEFT JOIN brands                b            ON p.brand_id = b.id
LEFT JOIN units_of_measure      uom          ON p.base_uom_id = uom.id
-- Base product barcodes
LEFT JOIN product_barcodes      pb_base      ON p.id = pb_base.product_id AND pb_base.product_variant_id IS NULL AND pb_base.is_primary = true
-- Variant barcodes (FIX 9.1 / 9.4)
LEFT JOIN product_barcodes      pb_variant   ON pv.id = pb_variant.product_variant_id AND pb_variant.is_primary = true
LEFT JOIN tax_categories        tc           ON p.tax_category_id = tc.id
-- Retail price: base product
LEFT JOIN product_prices pp_retail
    ON p.id = pp_retail.product_id
    AND pp_retail.product_variant_id IS NULL
    AND pp_retail.price_list_id = (SELECT id FROM price_lists WHERE code = 'RETAIL_SAR' AND is_active = true LIMIT 1)
    AND pp_retail.is_active = true
-- Retail price: variant (FIX 9.2 / 9.3)
LEFT JOIN product_prices pp_retail_v
    ON p.id = pp_retail_v.product_id
    AND pp_retail_v.product_variant_id = pv.id
    AND pp_retail_v.price_list_id = (SELECT id FROM price_lists WHERE code = 'RETAIL_SAR' AND is_active = true LIMIT 1)
    AND pp_retail_v.is_active = true
-- Promo price: base product
LEFT JOIN product_prices pp_promo
    ON p.id = pp_promo.product_id
    AND pp_promo.product_variant_id IS NULL
    AND pp_promo.price_list_id = (SELECT id FROM price_lists WHERE code = 'PROMO_SAR' AND is_active = true LIMIT 1)
    AND pp_promo.is_active = true
-- Promo price: variant
LEFT JOIN product_prices pp_promo_v
    ON p.id = pp_promo_v.product_id
    AND pp_promo_v.product_variant_id = pv.id
    AND pp_promo_v.price_list_id = (SELECT id FROM price_lists WHERE code = 'PROMO_SAR' AND is_active = true LIMIT 1)
    AND pp_promo_v.is_active = true
WHERE p.is_active = true
  AND p.is_sellable = true
ORDER BY pc.name, p.name, pv.variant_name;
-- =====================================================
-- FIX #9 (continued): Variant-aware POS functions
-- =====================================================

CREATE OR REPLACE FUNCTION fn_pos_get_products_with_stock(
    p_store_id              INTEGER,
    p_category_id           INTEGER  DEFAULT NULL,
    p_search_term           VARCHAR  DEFAULT NULL,
    p_include_out_of_stock  BOOLEAN  DEFAULT false
)
RETURNS TABLE (
    product_id          INTEGER,
    product_variant_id  INTEGER,
    sku                 VARCHAR,
    product_name        VARCHAR,
    variant_attributes  JSONB,
    description         TEXT,
    category_id         INTEGER,
    category_name       VARCHAR,
    brand_name          VARCHAR,
    barcode             VARCHAR,
    uom_code            VARCHAR,
    decimal_places      INTEGER,
    retail_price        NUMERIC,
    promo_price         NUMERIC,
    effective_price     NUMERIC,
    has_promotion       BOOLEAN,
    promotion_name      VARCHAR,
    discount_percent    VARCHAR,
    promo_min_quantity  NUMERIC,
    tax_rate            NUMERIC,
    tax_is_inclusive    BOOLEAN,
    quantity_available  NUMERIC,
    quantity_on_hand    NUMERIC,
    quantity_allocated  NUMERIC,
    is_in_stock         BOOLEAN,
    is_low_stock        BOOLEAN,
    reorder_level       NUMERIC,
    allow_decimal_quantity BOOLEAN,
    is_serialized       BOOLEAN,
    is_batch_managed    BOOLEAN,
    product_metadata    JSONB,
    product_variants    JSONB,
    package_n_price     JSONB,
    product_uom_conversions JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        cat.product_id,
        cat.product_variant_id,
        cat.sku::VARCHAR,
        cat.product_name::VARCHAR,
        cat.variant_attributes,
        cat.description,
        cat.category_id,
        cat.category_name::VARCHAR,
        cat.brand_name::VARCHAR,
        cat.barcode::VARCHAR,
        cat.uom_code::VARCHAR,
        cat.decimal_places::INTEGER,
        cat.retail_price,
        COALESCE(cat.promo_price, promo_rule.calculated_promo_price) AS promo_price,
        COALESCE(cat.promo_price, promo_rule.calculated_promo_price, cat.retail_price) AS effective_price,
        (cat.has_active_promotion OR (promo_rule.promo_name IS NOT NULL)) AS has_promotion,
        COALESCE(cat.promotion_name, promo_rule.promo_name)::VARCHAR AS promotion_name,
        COALESCE(cat.discount_percent, promo_rule.calc_discount_percent)::VARCHAR AS discount_percent,
        COALESCE(cat.promo_min_quantity, promo_rule.promo_min_qty) AS promo_min_quantity,
        cat.tax_rate,
        cat.tax_is_inclusive,
        -- FIX 9.3: Join on both product_id AND product_variant_id
        COALESCE(inv.quantity_available, 0)::NUMERIC,
        COALESCE(inv.quantity_on_hand,   0)::NUMERIC,
        COALESCE(inv.quantity_allocated, 0)::NUMERIC,
        (COALESCE(inv.quantity_available, 0) > 0),
        (COALESCE(inv.quantity_available, 0) <= COALESCE(inv.reorder_level, 0) AND COALESCE(inv.quantity_available, 0) > 0),
        COALESCE(inv.reorder_level, 0)::NUMERIC,
        cat.allow_decimal_quantity,
        cat.is_serialized,
        cat.is_batch_managed,
        cat.product_metadata,
        (SELECT COALESCE(jsonb_agg(
            jsonb_build_object(
                'id', pv.id,
                'product_id', pv.product_id,
                'variant_sku', pv.variant_sku,
                'variant_name', pv.variant_name,
                'variant_attributes', pv.variant_attributes,
                'price', (
                    SELECT ppv.price
                    FROM product_prices ppv
                    INNER JOIN price_lists plv ON ppv.price_list_id = plv.id AND plv.is_active = true
                    WHERE ppv.product_id = pv.product_id
                      AND ppv.product_variant_id = pv.id
                      AND ppv.is_active = true
                      AND (ppv.valid_from IS NULL OR ppv.valid_from <= CURRENT_DATE)
                      AND (ppv.valid_to   IS NULL OR ppv.valid_to   >= CURRENT_DATE)
                    ORDER BY ppv.valid_from DESC NULLS LAST, ppv.id DESC
                    LIMIT 1
                ),
                'is_active', pv.is_active,
                'metadata', COALESCE(pv.metadata, '{}'::jsonb),
                'created_at', pv.created_at,
                'updated_at', pv.updated_at
            )
            ORDER BY pv.id
        ), '[]'::jsonb)
         FROM product_variants pv
         WHERE pv.product_id = cat.product_id),
        -- FIX 9.2: Include variant_id and variant_attributes in package_n_price JSON
        (SELECT COALESCE(jsonb_agg(s.rec ORDER BY s.pl_code, s.uom_code), '[]'::jsonb)
         FROM (
             SELECT pl.code AS pl_code, uom.code AS uom_code,
                    jsonb_build_object(
                        'product_name',        cat.product_name,
                        'price_list_id',       pl.id,
                        'price_list_code',     pl.code,
                        'price_list_name',     pl.name,
                        'price_list',          pl.name,
                        'price_list_type',     pl.price_list_type,
                        'currency_code',       pl.currency_code,
                        'uom_id',              uom.id,
                        'uom_code',            uom.code,
                        'uom',                 uom.code,
                        'uom_name',            uom.name,
                        'decimal_places',      uom.decimal_places,
                        'price',               pp.price,
                        'min_quantity',        pp.min_quantity,
                        'max_quantity',        pp.max_quantity,
                        'valid_from',          pp.valid_from,
                        'valid_to',            pp.valid_to,
                        'metadata',            COALESCE(pp.metadata, '{}'::jsonb),
                        'barcodes',            (SELECT COALESCE(jsonb_agg(pb.barcode), '[]'::jsonb)
                                                FROM product_barcodes pb
                                                WHERE pb.product_id = pp.product_id
                                                  AND (pb.product_variant_id = cat.product_variant_id
                                                       OR (cat.product_variant_id IS NULL AND pb.product_variant_id IS NULL)))
                    ) AS rec
             FROM product_prices pp
             INNER JOIN price_lists pl ON pp.price_list_id = pl.id AND pl.is_active = true
             LEFT JOIN units_of_measure uom ON pp.uom_id = uom.id
             WHERE pp.product_id = cat.product_id
               AND (pp.product_variant_id = cat.product_variant_id OR pp.product_variant_id IS NULL)
               AND pp.is_active = true
               AND (pp.valid_from IS NULL OR pp.valid_from <= CURRENT_DATE)
               AND (pp.valid_to   IS NULL OR pp.valid_to   >= CURRENT_DATE)
         ) AS s),
        (SELECT COALESCE(jsonb_agg(conv.cv ORDER BY conv.fu_code, conv.tu_code), '[]'::jsonb)
         FROM (
             SELECT fu.code AS fu_code, tu.code AS tu_code,
                    jsonb_build_object(
                        'from_uom_id', fu.id, 'from_uom_code', fu.code, 'from_uom_name', fu.name,
                        'to_uom_id',   tu.id, 'to_uom_code',   tu.code, 'to_uom_name',   tu.name,
                        'conversion_factor', puc.conversion_factor
                    ) AS cv
             FROM product_uom_conversions puc
             JOIN units_of_measure fu ON puc.from_uom_id = fu.id
             JOIN units_of_measure tu ON puc.to_uom_id   = tu.id
             WHERE puc.product_id = cat.product_id
         ) AS conv)
    FROM vw_pos_product_catalog cat
    LEFT JOIN LATERAL (
        SELECT 
            pr.name AS promo_name,
            pr.min_quantity AS promo_min_qty,
            pr.discount_value,
            pr.promotion_type,
            CASE 
                WHEN pr.promotion_type = 'percentage_discount' AND cat.retail_price IS NOT NULL THEN
                    ROUND(cat.retail_price * (1.0 - (pr.discount_value / 100.0)), 2)
                WHEN pr.promotion_type = 'fixed_discount' AND cat.retail_price IS NOT NULL THEN
                    GREATEST(0.00, cat.retail_price - pr.discount_value)
                ELSE cat.retail_price
            END AS calculated_promo_price,
            CASE
                WHEN pr.promotion_type = 'percentage_discount' AND pr.discount_value IS NOT NULL THEN
                    CONCAT(TRIM(TRAILING '.' FROM TRIM(TRAILING '0' FROM pr.discount_value::text)), '%')
                ELSE NULL
            END AS calc_discount_percent
        FROM promotions pr
        WHERE pr.is_active = true
          AND (pr.valid_from IS NULL OR pr.valid_from <= CURRENT_TIMESTAMP)
          AND (pr.valid_to IS NULL OR pr.valid_to >= CURRENT_TIMESTAMP)
          AND (cardinality(pr.store_ids) = 0 OR p_store_id = ANY(pr.store_ids))
          AND (
              pr.applies_to = 'all'
              OR (pr.applies_to = 'product' AND cat.product_id = ANY(pr.target_product_ids))
              OR (pr.applies_to = 'category' AND cat.category_id = ANY(pr.target_category_ids))
          )
          AND (
              NOT (pr.metadata ? 'target_uoms')
              OR NOT (pr.metadata->'target_uoms' ? cat.product_id::text)
              OR (
                  jsonb_typeof(pr.metadata->'target_uoms'->(cat.product_id::text)) = 'array'
                  AND (pr.metadata->'target_uoms'->(cat.product_id::text)) @> cat.uom_id::text::jsonb
              )
              OR (pr.metadata->'target_uoms'->>(cat.product_id::text)) = cat.uom_id::text
          )
          AND (
              NOT (pr.metadata ? 'target_variants')
              OR NOT (pr.metadata->'target_variants' ? cat.product_id::text)
              OR cat.product_variant_id IS NULL
              OR (
                  jsonb_typeof(pr.metadata->'target_variants'->(cat.product_id::text)) = 'array'
                  AND (pr.metadata->'target_variants'->(cat.product_id::text)) @> cat.product_variant_id::text::jsonb
              )
              OR (pr.metadata->'target_variants'->>(cat.product_id::text)) = cat.product_variant_id::text
          )
        ORDER BY pr.created_at DESC
        LIMIT 1
    ) promo_rule ON true
    -- FIX 9.3: correct variant-aware inventory join
    LEFT JOIN inventory_stock inv
        ON inv.product_id = cat.product_id
        AND inv.store_id = p_store_id
        AND (inv.product_variant_id = cat.product_variant_id
             OR (cat.product_variant_id IS NULL AND inv.product_variant_id IS NULL))
    WHERE
        (p_category_id IS NULL OR cat.category_id = p_category_id)
        AND (p_search_term IS NULL
             OR cat.product_name ILIKE '%' || p_search_term || '%'
             OR cat.sku         ILIKE '%' || p_search_term || '%'
             OR cat.barcode     ILIKE '%' || p_search_term || '%')
        AND (p_include_out_of_stock = true OR COALESCE(inv.quantity_available, 0) > 0)
    ORDER BY cat.category_name, cat.product_name;
END;
$$ LANGUAGE plpgsql;


-- =====================================================
-- FIX #19 (P1): Low stock alert view
-- =====================================================

CREATE OR REPLACE VIEW vw_low_stock_alerts AS
SELECT
    ist.id              AS inventory_stock_id,
    s.id                AS store_id,
    s.name              AS store_name,
    s.code              AS store_code,
    p.id                AS product_id,
    p.sku,
    COALESCE(pv.variant_sku, p.sku) AS effective_sku,
    p.name              AS product_name,
    COALESCE(pv.variant_name, '')   AS variant_name,
    pv.id               AS product_variant_id,
    pc.name             AS category_name,
    ist.quantity_on_hand,
    ist.quantity_available,
    ist.quantity_allocated,
    ist.reorder_level,
    ist.reorder_quantity,
    CASE
        WHEN ist.quantity_available <= 0                                              THEN 'out_of_stock'
        WHEN ist.quantity_available <= COALESCE(ist.reorder_level, 0)               THEN 'low_stock'
        WHEN ist.quantity_available <= COALESCE(ist.reorder_level, 0) * 1.5         THEN 'near_reorder'
        ELSE 'adequate'
    END AS stock_status,
    (ist.quantity_available <= 0)                                                    AS is_out_of_stock,
    (ist.quantity_available > 0 AND ist.quantity_available <= COALESCE(ist.reorder_level, 0)) AS is_low_stock,
    ist.last_counted_at
FROM inventory_stock ist
JOIN stores s   ON s.id = ist.store_id AND s.is_active = true
JOIN products p ON p.id = ist.product_id AND p.is_active = true AND p.track_inventory = true
LEFT JOIN product_variants pv ON pv.id = ist.product_variant_id
LEFT JOIN product_categories pc ON pc.id = p.category_id
WHERE ist.quantity_available <= COALESCE(ist.reorder_level, 0)
   OR ist.quantity_available <= 0
ORDER BY s.name, CASE WHEN ist.quantity_available <= 0 THEN 0 ELSE 1 END, p.name;

-- =====================================================
-- FIX #20 (P1): Pending / overdue purchase orders view
-- =====================================================

CREATE OR REPLACE VIEW vw_pending_purchase_orders AS
SELECT
    po.id                   AS po_id,
    po.po_number,
    po.po_date,
    po.expected_delivery_date,
    po.status,
    CURRENT_DATE - po.expected_delivery_date AS days_overdue,
    (po.expected_delivery_date < CURRENT_DATE AND po.status NOT IN ('received','cancelled','closed')) AS is_overdue,
    s.id    AS store_id,
    s.name  AS store_name,
    sup.id  AS supplier_id,
    sup.name AS supplier_name,
    NULL::TEXT AS contact_person,
    NULL::TEXT AS supplier_email,
    po.subtotal,
    po.discount_amount,
    po.tax_amount,
    po.total_amount,
    (SELECT COALESCE(SUM(pol.quantity - pol.received_quantity), 0)
     FROM purchase_order_lines pol WHERE pol.purchase_order_id = po.id) AS outstanding_quantity,
    u_created.username AS created_by_username,
    u_approved.username AS approved_by_username,
    po.created_at
FROM purchase_orders po
JOIN stores s    ON s.id = po.store_id
JOIN business_partners sup ON sup.id = po.partners_id
LEFT JOIN users u_created  ON u_created.id  = po.created_by
LEFT JOIN users u_approved ON u_approved.id = po.approved_by
WHERE po.status NOT IN ('received','cancelled','closed')
ORDER BY is_overdue DESC, days_overdue DESC NULLS LAST, po.expected_delivery_date;

-- =====================================================
-- FIX #21 (P1): Customer aging report
-- =====================================================

CREATE OR REPLACE VIEW vw_customer_aging_report AS
SELECT
    c.id                    AS customer_id,
    c.customer_code,
    c.name                  AS customer_name,
    c.email,
    c.phone,
    c.customer_type,
    c.credit_limit,
    c.outstanding_balance,
    i.organization_id,
    COALESCE(SUM(CASE WHEN i.due_date >= CURRENT_DATE THEN i.balance_due ELSE 0 END), 0) AS current_amount,
    COALESCE(SUM(CASE WHEN i.due_date < CURRENT_DATE AND CURRENT_DATE - i.due_date <= 30 THEN i.balance_due ELSE 0 END), 0) AS overdue_1_30,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - i.due_date BETWEEN 31 AND 60 THEN i.balance_due ELSE 0 END), 0) AS overdue_31_60,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - i.due_date BETWEEN 61 AND 90 THEN i.balance_due ELSE 0 END), 0) AS overdue_61_90,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - i.due_date > 90 THEN i.balance_due ELSE 0 END), 0) AS overdue_over_90,
    COALESCE(SUM(CASE WHEN i.balance_due > 0 THEN i.balance_due ELSE 0 END), 0) AS total_outstanding,
    COUNT(CASE WHEN i.invoice_status = 'overdue' THEN 1 END)::INTEGER AS overdue_invoice_count,
    MAX(i.due_date)         AS latest_due_date,
    c.loyalty_points
FROM customers c
LEFT JOIN invoices i
    ON i.customer_id = c.id
    AND i.invoice_status NOT IN ('cancelled','draft')
    AND i.balance_due > 0
LEFT JOIN organizations o 
    ON o.id = i.organization_id
WHERE c.is_active = true
GROUP BY 
    c.id, c.customer_code, c.name, c.email, c.phone, 
    c.customer_type, c.credit_limit, c.outstanding_balance, 
    c.loyalty_points, i.organization_id
ORDER BY total_outstanding DESC;

-- =====================================================
-- FIX #22 (P1): Accounts payable overview
-- =====================================================

CREATE OR REPLACE VIEW vw_accounts_payable AS
SELECT
    po.id                       AS po_id,
    po.po_number,
    po.organization_id,
    org.name                    AS organization_name,
    sup.id                      AS supplier_id,
    sup.name                    AS supplier_name,
    NULL::TEXT                  AS contact_person,
    NULL::TEXT                  AS email,
    sup.payment_terms_id        AS supplier_payment_terms,
    s.name                      AS store_name,
    po.po_date,
    po.expected_delivery_date,
    po.status,
    po.total_amount             AS po_total,
    po.discount_amount,
    po.tax_amount,
    -- Amount outstanding: items received but not yet paid
    (SELECT COALESCE(SUM(pol.received_quantity * pol.unit_price), 0)
     FROM purchase_order_lines pol WHERE pol.purchase_order_id = po.id) AS received_amount,
    po.metadata->>'amount_paid' AS amount_paid_str,
    po.created_at
FROM purchase_orders po
JOIN organizations org ON org.id = po.organization_id
JOIN business_partners sup ON sup.id = po.partners_id
JOIN stores        s   ON s.id   = po.store_id
WHERE po.status IN ('partially_received','received','approved')
ORDER BY po.po_date;

-- =====================================================
-- FIX #24 (P2): Profit margin analysis view
-- =====================================================

CREATE OR REPLACE VIEW vw_profit_margin_analysis AS
SELECT
    ptl.product_id,
    ptl.product_variant_id,
    p.sku,
    p.name AS product_name,
    pc.name AS category_name,
    pt.store_id,
    s.name AS store_name,
    DATE_TRUNC('month', pt.transaction_date)::DATE AS period_month,
    SUM(ptl.quantity) AS units_sold,
    SUM(ptl.line_total) AS total_revenue,
    SUM(ptl.cost_price * ptl.quantity) AS total_cost,
    SUM(ptl.line_total) - SUM(ptl.cost_price * ptl.quantity) AS gross_profit,
    CASE
        WHEN SUM(ptl.line_total) > 0
        THEN ROUND(
            (SUM(ptl.line_total) - SUM(ptl.cost_price * ptl.quantity))
            / SUM(ptl.line_total) * 100, 2
        )
        ELSE 0
    END AS gross_margin_pct,
    SUM(ptl.discount_amount) AS total_discounts,
    SUM(ptl.tax_amount) AS total_taxes
FROM pos_transaction_lines ptl
JOIN pos_transactions pt 
    ON pt.id = ptl.transaction_id 
    AND pt.status = 'completed'
JOIN products p 
    ON p.id = ptl.product_id
LEFT JOIN product_variants pv_id 
    ON pv_id.id = ptl.product_variant_id
LEFT JOIN product_categories pc 
    ON pc.id = p.category_id
JOIN stores s 
    ON s.id = pt.store_id
GROUP BY 
    ptl.product_id, 
    ptl.product_variant_id,
    p.sku, 
    p.name, 
    pc.name,
    pt.store_id, 
    s.name, 
    DATE_TRUNC('month', pt.transaction_date)
ORDER BY 
    period_month DESC, 
    gross_profit DESC;

-- =====================================================
-- FIX #25 (P1): Flattened user effective permissions view
-- =====================================================

CREATE OR REPLACE VIEW vw_user_effective_permissions AS
SELECT DISTINCT
    u.id            AS user_id,
    u.username,
    u.email,
    u.organization_id,
    r.id            AS role_id,
    r.name          AS role_name,
    r.code          AS role_code,
    per.id          AS permission_id,
    per.name        AS permission_name,
    per.code        AS permission_code,
    rp.scope,
    usa.store_id    AS accessible_store_id
FROM users u
JOIN user_roles    ur  ON ur.user_id = u.id
JOIN roles         r   ON r.id = ur.role_id AND r.is_active = true
JOIN role_permissions rp ON rp.role_id = r.id
JOIN permissions   per ON per.id = rp.permission_id
LEFT JOIN user_store_access usa ON usa.user_id = u.id
WHERE u.is_active = true
ORDER BY u.username, r.name, per.code;

CREATE OR REPLACE FUNCTION fn_pos_get_product_by_barcode(p_barcode VARCHAR, p_store_id INTEGER)
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
    product_metadata JSONB,
    package_n_price JSONB,
    product_uom_conversions JSONB
) AS $$
DECLARE
    v_product_id INT;
    v_variant_id INT;
BEGIN
    -- 1. Find product / variant by matching barcode
    SELECT pb.product_id, pb.product_variant_id INTO v_product_id, v_variant_id
    FROM product_barcodes pb
    WHERE pb.barcode = p_barcode
    LIMIT 1;

    IF v_product_id IS NULL THEN
        RETURN;
    END IF;

    -- 2. Return catalog query matching product ID and variant ID
    RETURN QUERY
    SELECT 
        cat.product_id,
        cat.sku::VARCHAR,
        cat.product_name::VARCHAR,
        cat.description,
        cat.category_name::VARCHAR,
        cat.brand_name::VARCHAR,
        p_barcode::VARCHAR AS barcode, -- Return scanned barcode
        cat.uom_code::VARCHAR,
        (cat.decimal_places)::INTEGER,
        cat.retail_price,
        COALESCE(cat.promo_price, promo_rule.calculated_promo_price) AS promo_price,
        COALESCE(cat.promo_price, promo_rule.calculated_promo_price, cat.retail_price) AS effective_price,
        (cat.has_active_promotion OR (promo_rule.promo_name IS NOT NULL)) AS has_promotion,
        COALESCE(cat.promotion_name, promo_rule.promo_name)::VARCHAR AS promotion_name,
        COALESCE(cat.promo_min_quantity, promo_rule.promo_min_qty) AS promo_min_quantity,
        cat.tax_rate,
        cat.tax_is_inclusive,
        COALESCE(inv.quantity_available, 0)::NUMERIC,
        (COALESCE(inv.quantity_available, 0) > 0),
        cat.allow_decimal_quantity,
        cat.is_serialized,
        cat.is_batch_managed,
        cat.product_metadata,
        (SELECT COALESCE(jsonb_agg(s.rec ORDER BY s.pl_code, s.uom_code), '[]'::jsonb)
         FROM (
             SELECT 
                 pl.code AS pl_code,
                 uom.code AS uom_code,
                 jsonb_build_object(
                     'product_name', p.name,
                     'price_list_id', pl.id,
                     'price_list_code', pl.code,
                     'price_list_name', pl.name,
                     'price_list', pl.name,
                     'price_list_type', pl.price_list_type,
                     'currency_code', pl.currency_code,
                     'uom_id', uom.id,
                     'uom_code', uom.code,
                     'uom', uom.code,
                     'uom_name', uom.name,
                     'decimal_places', uom.decimal_places,
                     'price', pp.price,
                     'min_quantity', pp.min_quantity,
                     'max_quantity', pp.max_quantity,
                     'valid_from', pp.valid_from,
                     'valid_to', pp.valid_to,
                     'metadata', COALESCE(pp.metadata, '{}'::jsonb),
                     'barcodes', (SELECT COALESCE(jsonb_agg(pb.barcode), '[]'::jsonb) FROM product_barcodes pb WHERE pb.product_id = pp.product_id)
                 ) AS rec
             FROM product_prices pp
             INNER JOIN products p ON pp.product_id = p.id
             INNER JOIN price_lists pl ON pp.price_list_id = pl.id AND pl.is_active = true
             LEFT JOIN units_of_measure uom ON pp.uom_id = uom.id
             WHERE pp.product_id = cat.product_id
               AND pp.is_active = true
               AND (pp.valid_from IS NULL OR pp.valid_from <= CURRENT_DATE)
               AND (pp.valid_to IS NULL OR pp.valid_to >= CURRENT_DATE)
         ) AS s),
        (SELECT COALESCE(jsonb_agg(conv.cv ORDER BY conv.fu_code, conv.tu_code), '[]'::jsonb)
         FROM (
             SELECT fu.code AS fu_code, tu.code AS tu_code,
                    jsonb_build_object(
                         'from_uom_id', fu.id, 'from_uom_code', fu.code, 'from_uom_name', fu.name,
                         'to_uom_id', tu.id, 'to_uom_code', tu.code, 'to_uom_name', tu.name,
                         'conversion_factor', puc.conversion_factor
                    ) AS cv
             FROM product_uom_conversions puc
             JOIN units_of_measure fu ON puc.from_uom_id = fu.id
             JOIN units_of_measure tu ON puc.to_uom_id = tu.id
             WHERE puc.product_id = cat.product_id
         ) AS conv)
    FROM vw_pos_product_catalog cat
    LEFT JOIN LATERAL (
        SELECT 
            pr.name AS promo_name,
            pr.min_quantity AS promo_min_qty,
            pr.discount_value,
            pr.promotion_type,
            CASE 
                WHEN pr.promotion_type = 'percentage_discount' AND cat.retail_price IS NOT NULL THEN
                    ROUND(cat.retail_price * (1.0 - (pr.discount_value / 100.0)), 2)
                WHEN pr.promotion_type = 'fixed_discount' AND cat.retail_price IS NOT NULL THEN
                    GREATEST(0.00, cat.retail_price - pr.discount_value)
                ELSE cat.retail_price
            END AS calculated_promo_price
        FROM promotions pr
        WHERE pr.is_active = true
          AND (pr.valid_from IS NULL OR pr.valid_from <= CURRENT_TIMESTAMP)
          AND (pr.valid_to IS NULL OR pr.valid_to >= CURRENT_TIMESTAMP)
          AND (cardinality(pr.store_ids) = 0 OR p_store_id = ANY(pr.store_ids))
          AND (
              pr.applies_to = 'all'
              OR (pr.applies_to = 'product' AND cat.product_id = ANY(pr.target_product_ids))
              OR (pr.applies_to = 'category' AND cat.category_id = ANY(pr.target_category_ids))
          )
          AND (
              NOT (pr.metadata ? 'target_uoms')
              OR NOT (pr.metadata->'target_uoms' ? cat.product_id::text)
              OR (
                  jsonb_typeof(pr.metadata->'target_uoms'->(cat.product_id::text)) = 'array'
                  AND (pr.metadata->'target_uoms'->(cat.product_id::text)) @> cat.uom_id::text::jsonb
              )
              OR (pr.metadata->'target_uoms'->>(cat.product_id::text)) = cat.uom_id::text
          )
          AND (
              NOT (pr.metadata ? 'target_variants')
              OR NOT (pr.metadata->'target_variants' ? cat.product_id::text)
              OR cat.product_variant_id IS NULL
              OR (
                  jsonb_typeof(pr.metadata->'target_variants'->(cat.product_id::text)) = 'array'
                  AND (pr.metadata->'target_variants'->(cat.product_id::text)) @> cat.product_variant_id::text::jsonb
              )
              OR (pr.metadata->'target_variants'->>(cat.product_id::text)) = cat.product_variant_id::text
          )
        ORDER BY pr.created_at DESC
        LIMIT 1
    ) promo_rule ON true
    LEFT JOIN inventory_stock inv ON cat.product_id = inv.product_id AND inv.store_id = p_store_id
    WHERE cat.product_id = v_product_id
      AND (cat.product_variant_id = v_variant_id OR (cat.product_variant_id IS NULL AND v_variant_id IS NULL))
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

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
    is_in_stock BOOLEAN,
    package_n_price JSONB,
    product_uom_conversions JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cat.product_id,
        cat.sku::VARCHAR,
        cat.product_name::VARCHAR,
        cat.category_name::VARCHAR,
        cat.brand_name::VARCHAR,
        cat.barcode::VARCHAR,
        cat.effective_price,
        cat.has_active_promotion,
        cat.promotion_name::VARCHAR,
        COALESCE(inv.quantity_available, 0)::NUMERIC,
        (COALESCE(inv.quantity_available, 0) > 0),
        (SELECT COALESCE(jsonb_agg(s.rec ORDER BY s.pl_code, s.uom_code), '[]'::jsonb)
         FROM (
             SELECT 
                 pl.code AS pl_code,
                 uom.code AS uom_code,
                 jsonb_build_object(
                     'product_name', p.name,
                     'price_list_id', pl.id,
                     'price_list_code', pl.code,
                     'price_list_name', pl.name,
                     'price_list', pl.name,
                     'price_list_type', pl.price_list_type,
                     'currency_code', pl.currency_code,
                     'uom_id', uom.id,
                     'uom_code', uom.code,
                     'uom', uom.code,
                     'uom_name', uom.name,
                     'decimal_places', uom.decimal_places,
                     'price', pp.price,
                     'min_quantity', pp.min_quantity,
                     'max_quantity', pp.max_quantity,
                     'valid_from', pp.valid_from,
                     'valid_to', pp.valid_to,
                     'metadata', COALESCE(pp.metadata, '{}'::jsonb),
                     'barcodes', (SELECT COALESCE(jsonb_agg(pb.barcode), '[]'::jsonb) FROM product_barcodes pb WHERE pb.product_id = pp.product_id)
                 ) AS rec
             FROM product_prices pp
             INNER JOIN products p ON pp.product_id = p.id
             INNER JOIN price_lists pl ON pp.price_list_id = pl.id AND pl.is_active = true
             LEFT JOIN units_of_measure uom ON pp.uom_id = uom.id
             WHERE pp.product_id = cat.product_id
               AND pp.is_active = true
               AND (pp.valid_from IS NULL OR pp.valid_from <= CURRENT_DATE)
               AND (pp.valid_to IS NULL OR pp.valid_to >= CURRENT_DATE)
         ) AS s),
        (SELECT COALESCE(jsonb_agg(conv.cv ORDER BY conv.fu_code, conv.tu_code), '[]'::jsonb)
         FROM (
             SELECT fu.code AS fu_code, tu.code AS tu_code,
                    jsonb_build_object(
                        'from_uom_id', fu.id, 'from_uom_code', fu.code, 'from_uom_name', fu.name,
                        'to_uom_id', tu.id, 'to_uom_code', tu.code, 'to_uom_name', tu.name,
                        'conversion_factor', puc.conversion_factor
                    ) AS cv
             FROM product_uom_conversions puc
             JOIN units_of_measure fu ON puc.from_uom_id = fu.id
             JOIN units_of_measure tu ON puc.to_uom_id = tu.id
             WHERE puc.product_id = cat.product_id
         ) AS conv)
    FROM vw_pos_product_catalog cat
    LEFT JOIN inventory_stock inv ON cat.product_id = inv.product_id AND inv.store_id = p_store_id
    WHERE 
        (cat.category_id = p_category_id 
         OR (p_include_subcategories = true AND cat.parent_category_id = p_category_id))
        AND COALESCE(inv.quantity_available, 0) > 0
    ORDER BY cat.product_name;
END;
$$ LANGUAGE plpgsql;

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
    relevance_score INTEGER,
    package_n_price JSONB,
    product_uom_conversions JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cat.product_id,
        cat.sku::VARCHAR,
        cat.product_name::VARCHAR,
        cat.category_name::VARCHAR,
        cat.brand_name::VARCHAR,
        cat.barcode::VARCHAR,
        cat.effective_price,
        cat.has_active_promotion,
        COALESCE(inv.quantity_available, 0)::NUMERIC,
        (COALESCE(inv.quantity_available, 0) > 0),
        (CASE 
            WHEN cat.sku ILIKE p_search_term THEN 100
            WHEN cat.product_name ILIKE p_search_term THEN 90
            WHEN cat.barcode = p_search_term THEN 95
            WHEN cat.sku ILIKE p_search_term || '%' THEN 80
            WHEN cat.product_name ILIKE p_search_term || '%' THEN 70
            WHEN cat.sku ILIKE '%' || p_search_term || '%' THEN 60
            WHEN cat.product_name ILIKE '%' || p_search_term || '%' THEN 50
            ELSE 40
        END)::INTEGER,
        (SELECT COALESCE(jsonb_agg(s.rec ORDER BY s.pl_code, s.uom_code), '[]'::jsonb)
         FROM (
             SELECT 
                 pl.code AS pl_code,
                 uom.code AS uom_code,
                 jsonb_build_object(
                     'product_name', p.name,
                     'price_list_id', pl.id,
                     'price_list_code', pl.code,
                     'price_list_name', pl.name,
                     'price_list', pl.name,
                     'price_list_type', pl.price_list_type,
                     'currency_code', pl.currency_code,
                     'uom_id', uom.id,
                     'uom_code', uom.code,
                     'uom', uom.code,
                     'uom_name', uom.name,
                     'decimal_places', uom.decimal_places,
                     'price', pp.price,
                     'min_quantity', pp.min_quantity,
                     'max_quantity', pp.max_quantity,
                     'valid_from', pp.valid_from,
                     'valid_to', pp.valid_to,
                     'metadata', COALESCE(pp.metadata, '{}'::jsonb),
                     'barcodes', (SELECT COALESCE(jsonb_agg(pb.barcode), '[]'::jsonb) FROM product_barcodes pb WHERE pb.product_id = pp.product_id)
                 ) AS rec
             FROM product_prices pp
             INNER JOIN products p ON pp.product_id = p.id
             INNER JOIN price_lists pl ON pp.price_list_id = pl.id AND pl.is_active = true
             LEFT JOIN units_of_measure uom ON pp.uom_id = uom.id
             WHERE pp.product_id = cat.product_id
               AND pp.is_active = true
               AND (pp.valid_from IS NULL OR pp.valid_from <= CURRENT_DATE)
               AND (pp.valid_to IS NULL OR pp.valid_to >= CURRENT_DATE)
         ) AS s),
        (SELECT COALESCE(jsonb_agg(conv.cv ORDER BY conv.fu_code, conv.tu_code), '[]'::jsonb)
         FROM (
             SELECT fu.code AS fu_code, tu.code AS tu_code,
                    jsonb_build_object(
                        'from_uom_id', fu.id, 'from_uom_code', fu.code, 'from_uom_name', fu.name,
                        'to_uom_id', tu.id, 'to_uom_code', tu.code, 'to_uom_name', tu.name,
                        'conversion_factor', puc.conversion_factor
                    ) AS cv
             FROM product_uom_conversions puc
             JOIN units_of_measure fu ON puc.from_uom_id = fu.id
             JOIN units_of_measure tu ON puc.to_uom_id = tu.id
             WHERE puc.product_id = cat.product_id
         ) AS conv)
    FROM vw_pos_product_catalog cat
    LEFT JOIN inventory_stock inv ON cat.product_id = inv.product_id AND inv.store_id = p_store_id
    WHERE 
        cat.product_name ILIKE '%' || p_search_term || '%'
        OR cat.sku ILIKE '%' || p_search_term || '%'
        OR cat.barcode ILIKE '%' || p_search_term || '%'
    ORDER BY 11 DESC, cat.product_name
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE VIEW vw_pos_categories AS
SELECT
    pc.id               AS category_id,
    pc.code             AS category_code,
    pc.name             AS category_name,
    pc.parent_category_id,
    pc_parent.name      AS parent_category_name,
    COUNT(DISTINCT p.id)::INTEGER AS product_count,
    COUNT(DISTINCT CASE 
        WHEN inv.quantity_available > 0 THEN p.id 
    END)::INTEGER AS in_stock_count,
    pc.metadata         AS category_metadata
FROM product_categories pc
LEFT JOIN product_categories pc_parent 
    ON pc.parent_category_id = pc_parent.id
LEFT JOIN products p 
    ON pc.id = p.category_id
LEFT JOIN inventory_stock inv 
    ON p.id = inv.product_id
GROUP BY 
    pc.id, 
    pc.code, 
    pc.name, 
    pc.parent_category_id, 
    pc_parent.name, 
    pc.metadata
ORDER BY 
    pc_parent.name NULLS FIRST, 
    pc.name;

-- =====================================================
-- RESTAURANT MODULE VIEWS
-- =====================================================

CREATE OR REPLACE VIEW vw_restaurant_menu AS
SELECT
    mi.id                       AS menu_item_id,
    mi.store_id,
    mi.name                     AS item_name,
    mi.short_name,
    mi.description,
    mi.image_url,
    mi.base_price,
    mi.cost_price,
    mi.preparation_time_min,
    mi.is_available,
    mi.is_active,
    mi.display_order,
    mi.metadata                 AS item_metadata,
    mc.id                       AS category_id,
    mc.name                     AS category_name,
    mc.code                     AS category_code,
    mc.parent_category_id,
    mc.display_order            AS category_display_order,
    mc.image_url                AS category_image_url,
    mc_parent.name              AS parent_category_name,
    tc.id                       AS tax_category_id,
    tc.tax_rate,
    tc.is_inclusive             AS tax_is_inclusive,
    mi.recipe_id,
    r.recipe_name,
    r.yield_quantity            AS recipe_yield,
    mi.product_id,
    p.sku                       AS product_sku,
    (SELECT COUNT(*) FROM menu_item_modifiers m WHERE m.menu_item_id = mi.id AND m.is_active = true)::INTEGER
                                AS active_modifier_count,
    CASE
        WHEN mi.base_price > 0 AND mi.cost_price > 0
        THEN ROUND(((mi.base_price - mi.cost_price) / mi.base_price) * 100, 2)
        ELSE NULL
    END                         AS margin_percent
FROM menu_items mi
JOIN menu_categories mc         ON mi.menu_category_id = mc.id
LEFT JOIN menu_categories mc_parent ON mc.parent_category_id = mc_parent.id
LEFT JOIN tax_categories tc     ON mi.tax_category_id = tc.id
LEFT JOIN recipes r             ON mi.recipe_id = r.id
LEFT JOIN products p            ON mi.product_id = p.id
WHERE mi.is_active = true;

CREATE OR REPLACE VIEW vw_recipe_bom AS
SELECT
    r.id                        AS recipe_id,
    r.recipe_code,
    r.recipe_name,
    r.yield_quantity,
    r.organization_id,
    ri.id                       AS ingredient_line_id,
    ri.line_number,
    ri.quantity                 AS ingredient_qty,
    ri.is_optional,
    ri.is_byproduct,
    p.id                        AS product_id,
    p.sku,
    p.name                      AS product_name,
    pv.id                       AS variant_id,
    pv.variant_name,
    uom.id                      AS uom_id,
    uom.code                    AS uom_code,
    uom.name                    AS uom_name,
    pp.price                    AS unit_cost_estimate,
    ROUND(ri.quantity * COALESCE(pp.price, 0), 4) AS line_cost_estimate
FROM recipes r
JOIN recipe_ingredients ri      ON r.id = ri.recipe_id
JOIN products p                 ON ri.product_id = p.id
LEFT JOIN product_variants pv   ON ri.product_variant_id = pv.id
LEFT JOIN units_of_measure uom  ON ri.uom_id = uom.id
LEFT JOIN product_prices pp     ON p.id = pp.product_id
    AND pp.price_list_id        = (SELECT id FROM price_lists WHERE code = 'RETAIL_SAR' AND is_active = true LIMIT 1)
    AND pp.is_active            = true
WHERE r.is_active = true;

CREATE OR REPLACE VIEW vw_active_restaurant_orders AS
SELECT
    ro.id                       AS order_id,
    ro.order_number,
    ro.store_id,
    ro.order_source,
    ro.status                   AS order_status,
    ro.subtotal,
    ro.tax_amount,
    ro.total_amount,
    ro.notes,
    ro.ordered_at,
    ro.confirmed_at,
    rt.id                       AS table_id,
    rt.table_number,
    rt.table_name,
    rt.section                  AS table_section,
    c.id                        AS cashier_id,
    u.first_name || ' ' || u.last_name AS waiter_name,
    ro.customer_id,
    cust.name                   AS customer_name,
    EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - ro.ordered_at)) / 60.0 AS minutes_since_ordered
FROM restaurant_orders ro
LEFT JOIN restaurant_tables rt  ON ro.table_id = rt.id
LEFT JOIN cashiers c            ON ro.cashier_id = c.id
LEFT JOIN users u               ON c.user_id = u.id
LEFT JOIN customers cust        ON ro.customer_id = cust.id
WHERE ro.status NOT IN ('paid', 'voided');

CREATE OR REPLACE VIEW vw_waste_daily_summary AS
SELECT
    wl.store_id,
    DATE(wl.wasted_at)          AS waste_date,
    wl.waste_source,
    COUNT(*)                    AS waste_entries,
    SUM(wl.quantity)            AS total_quantity_wasted,
    SUM(wl.total_cost)          AS total_cost_wasted,
    AVG(wl.total_cost)          AS avg_cost_per_entry
FROM waste_logs wl
GROUP BY wl.store_id, DATE(wl.wasted_at), wl.waste_source;

-- =====================================================
-- RESTAURANT MODULE FUNCTIONS
-- =====================================================

CREATE OR REPLACE FUNCTION fn_get_restaurant_menu(
    p_store_id          INTEGER,
    p_category_id       INTEGER  DEFAULT NULL,
    p_include_unavail   BOOLEAN  DEFAULT false
)
RETURNS TABLE (
    menu_item_id            INTEGER,
    item_name               VARCHAR,
    short_name              VARCHAR,
    description             TEXT,
    image_url               TEXT,
    base_price              NUMERIC,
    preparation_time_min    INTEGER,
    is_available            BOOLEAN,
    category_id             INTEGER,
    category_name           VARCHAR,
    parent_category_name    VARCHAR,
    tax_rate                NUMERIC,
    tax_is_inclusive        BOOLEAN,
    recipe_id               INTEGER,
    product_id              INTEGER,
    active_modifier_count   INTEGER,
    margin_percent          NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        vm.menu_item_id,
        vm.item_name::VARCHAR,
        vm.short_name::VARCHAR,
        vm.description,
        vm.image_url,
        vm.base_price,
        vm.preparation_time_min,
        vm.is_available,
        vm.category_id,
        vm.category_name::VARCHAR,
        vm.parent_category_name::VARCHAR,
        vm.tax_rate,
        vm.tax_is_inclusive,
        vm.recipe_id,
        vm.product_id,
        vm.active_modifier_count,
        vm.margin_percent
    FROM vw_restaurant_menu vm
    WHERE vm.store_id = p_store_id
      AND (p_category_id IS NULL OR vm.category_id = p_category_id)
      AND (p_include_unavail = true OR vm.is_available = true)
    ORDER BY vm.category_display_order, vm.display_order;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_get_item_modifiers(
    p_menu_item_id INTEGER
)
RETURNS TABLE (
    modifier_id         INTEGER,
    modifier_name       VARCHAR,
    modifier_type       VARCHAR,
    price_adjustment    NUMERIC,
    display_order       INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        m.id,
        m.modifier_name::VARCHAR,
        m.modifier_type::VARCHAR,
        m.price_adjustment,
        m.display_order
    FROM menu_item_modifiers m
    WHERE m.menu_item_id = p_menu_item_id
      AND m.is_active    = true
    ORDER BY m.display_order;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- FIX #2 (P0): Atomic inter-store / warehouse stock transfer function
-- =====================================================

CREATE OR REPLACE FUNCTION fn_process_stock_transfer(
    p_from_store_id     INTEGER,
    p_to_store_id       INTEGER,
    p_product_id        INTEGER,
    p_product_variant_id INTEGER,
    p_quantity          DECIMAL(15,3),
    p_from_location_id  INTEGER DEFAULT NULL,
    p_to_location_id    INTEGER DEFAULT NULL,
    p_batch_number      VARCHAR DEFAULT NULL,
    p_performed_by      INTEGER DEFAULT NULL,
    p_notes             TEXT    DEFAULT NULL
)
RETURNS TABLE (
    success          BOOLEAN,
    message          TEXT,
    movement_id      INTEGER
) AS $$
DECLARE
    v_available  DECIMAL(15,3);
    v_movement_id INTEGER;
    v_ref_num    VARCHAR(50);
BEGIN
    -- Validate same store check
    IF p_from_store_id = p_to_store_id THEN
        RETURN QUERY SELECT false, 'Source and destination stores must differ.', NULL::INTEGER;
        RETURN;
    END IF;

    -- Lock and read available stock at source
    SELECT quantity_available INTO v_available
    FROM inventory_stock
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_from_store_id
    FOR UPDATE;

    IF v_available IS NULL OR v_available < p_quantity THEN
        RETURN QUERY SELECT false,
            format('Insufficient stock. Available: %s, Requested: %s', COALESCE(v_available, 0), p_quantity),
            NULL::INTEGER;
        RETURN;
    END IF;

    v_ref_num := 'TRF-' || to_char(CURRENT_TIMESTAMP, 'YYYYMMDDHH24MISS') || '-' || p_from_store_id || '-' || p_to_store_id;

    -- Deduct from source
    UPDATE inventory_stock
    SET quantity_on_hand   = quantity_on_hand   - p_quantity,
        quantity_available = quantity_available - p_quantity,
        quantity_in_transit = quantity_in_transit + p_quantity,
        updated_at = CURRENT_TIMESTAMP
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_from_store_id;

    -- Add to destination
    UPDATE inventory_stock
    SET quantity_on_hand   = quantity_on_hand   + p_quantity,
        quantity_available = quantity_available + p_quantity,
        updated_at         = CURRENT_TIMESTAMP
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_to_store_id;

    IF NOT FOUND THEN
        INSERT INTO inventory_stock (product_id, product_variant_id, store_id, storage_location_id,
            quantity_on_hand, quantity_available, quantity_in_transit)
        VALUES (p_product_id, p_product_variant_id, p_to_store_id, p_to_location_id,
                p_quantity, p_quantity, 0);
    END IF;

    -- Clear in-transit at source
    UPDATE inventory_stock
    SET quantity_in_transit = GREATEST(0, quantity_in_transit - p_quantity),
        updated_at = CURRENT_TIMESTAMP
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_from_store_id;

    -- Record movement
    INSERT INTO stock_movements (movement_type, reference_type, product_id, product_variant_id,
        from_store_id, to_store_id, from_location_id, to_location_id,
        quantity, batch_number, posted_by, status,
        metadata)
    VALUES ('transfer', 'manual', p_product_id, p_product_variant_id,
            p_from_store_id, p_to_store_id, p_from_location_id, p_to_location_id,
            p_quantity, p_batch_number, p_performed_by, 'completed',
            jsonb_build_object('reference_number', v_ref_num, 'notes', p_notes))
    RETURNING id INTO v_movement_id;

    RETURN QUERY SELECT true, 'Transfer completed successfully. Ref: ' || v_ref_num, v_movement_id;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- LOGISTICS & IN-TRANSIT WORKFLOW FUNCTIONS
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
            conversion_factor::NUMERIC,
            1 as level
        FROM product_uom_conversions
        WHERE product_id = p_product_id
            AND from_uom_id = v_from_uom_id
        
        UNION ALL
        
        -- Recursive case: chain conversions
        SELECT 
            puc.from_uom_id,
            puc.to_uom_id,
            (up.conversion_factor * puc.conversion_factor)::NUMERIC,
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

CREATE OR REPLACE FUNCTION fn_approve_transfer_request(
    p_transfer_request_id INTEGER,
    p_approved_by INTEGER
)
RETURNS TABLE (
    success BOOLEAN,
    message TEXT
) AS $$
DECLARE
    v_req RECORD;
BEGIN
    SELECT * INTO v_req FROM transfer_requests WHERE id = p_transfer_request_id FOR UPDATE;
    IF v_req IS NULL THEN
        RETURN QUERY SELECT false, 'Transfer request not found.';
        RETURN;
    END IF;

    IF v_req.status NOT IN ('draft', 'pending_approval') THEN
        RETURN QUERY SELECT false, 'Transfer request can only be approved from draft or pending_approval state.';
        RETURN;
    END IF;

    UPDATE transfer_requests
    SET status = 'approved',
        approved_by = p_approved_by,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_transfer_request_id;

    RETURN QUERY SELECT true, 'Transfer request approved successfully.';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_ship_transfer_request(
    p_transfer_request_id INTEGER,
    p_shipped_by INTEGER
)
RETURNS TABLE (
    success BOOLEAN,
    message TEXT
) AS $$
DECLARE
    v_req RECORD;
    v_item RECORD;
    v_available DECIMAL(15,3);
    v_qty DECIMAL(15,3);
    v_uom_code VARCHAR;
    v_base_qty DECIMAL(15,3);
BEGIN
    SELECT * INTO v_req FROM transfer_requests WHERE id = p_transfer_request_id FOR UPDATE;
    IF v_req IS NULL THEN
        RETURN QUERY SELECT false, 'Transfer request not found.';
        RETURN;
    END IF;

    IF v_req.status NOT IN ('approved', 'pending_approval', 'draft') THEN
        RETURN QUERY SELECT false, 'Transfer request must be approved to ship.';
        RETURN;
    END IF;

    FOR v_item IN SELECT * FROM transfer_request_items WHERE transfer_request_id = p_transfer_request_id FOR UPDATE LOOP
        v_qty := CASE WHEN v_item.approved_quantity > 0 THEN v_item.approved_quantity ELSE v_item.requested_quantity END;
        IF v_qty <= 0 THEN
            CONTINUE;
        END IF;

        -- Retrieve UOM code for conversion
        SELECT code INTO v_uom_code FROM units_of_measure WHERE id = v_item.uom_id;
        
        -- Convert requested qty to base unit qty
        v_base_qty := fn_convert_uom_quantity(v_item.product_id, v_uom_code, v_qty);
        IF v_base_qty IS NULL THEN
            v_base_qty := v_qty;
        END IF;

        -- Check available stock at source using v_base_qty
        SELECT quantity_available INTO v_available
        FROM inventory_stock
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.from_store_id
        FOR UPDATE;

        IF v_available IS NULL OR v_available < v_base_qty THEN
            RETURN QUERY SELECT false, format('Insufficient stock for product ID %s at source store.', v_item.product_id);
            RETURN;
        END IF;

        -- Deduct from source store using v_base_qty
        UPDATE inventory_stock
        SET quantity_on_hand = quantity_on_hand - v_base_qty,
            quantity_available = quantity_available - v_base_qty,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.from_store_id;

        -- Increment quantity_in_transit at destination store using v_base_qty
        UPDATE inventory_stock
        SET quantity_in_transit = quantity_in_transit + v_base_qty,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.to_store_id;

        IF NOT FOUND THEN
            INSERT INTO inventory_stock (product_id, product_variant_id, store_id, storage_location_id,
                quantity_on_hand, quantity_available, quantity_in_transit)
            VALUES (v_item.product_id, v_item.product_variant_id, v_req.to_store_id, v_item.to_location_id,
                    0, 0, v_base_qty);
        END IF;

        -- Update item shipped_quantity
        UPDATE transfer_request_items
        SET shipped_quantity = v_qty,
            approved_quantity = v_qty
        WHERE id = v_item.id;

        -- Record stock movement (transfer_out / shipped) using v_base_qty
        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            from_store_id, to_store_id, from_location_id, to_location_id,
            quantity, uom_id, batch_number, posted_by, status, metadata
        ) VALUES (
            'transfer_out', 'transfer_request', p_transfer_request_id, v_item.product_id, v_item.product_variant_id,
            v_req.from_store_id, v_req.to_store_id, v_item.from_location_id, v_item.to_location_id,
            v_base_qty, v_item.uom_id, v_item.batch_number, p_shipped_by, 'shipped',
            jsonb_build_object('transfer_number', v_req.transfer_number)
        );
    END LOOP;

    -- Update header
    UPDATE transfer_requests
    SET status = 'shipped',
        shipped_by = p_shipped_by,
        shipped_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_transfer_request_id;

    RETURN QUERY SELECT true, 'Transfer request shipped successfully.';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_receive_transfer_request(
    p_transfer_request_id INTEGER,
    p_received_by INTEGER
)
RETURNS TABLE (
    success BOOLEAN,
    message TEXT
) AS $$
DECLARE
    v_req RECORD;
    v_item RECORD;
    v_qty DECIMAL(15,3);
    v_uom_code VARCHAR;
    v_base_qty DECIMAL(15,3);
BEGIN
    SELECT * INTO v_req FROM transfer_requests WHERE id = p_transfer_request_id FOR UPDATE;
    IF v_req IS NULL THEN
        RETURN QUERY SELECT false, 'Transfer request not found.';
        RETURN;
    END IF;

    IF v_req.status NOT IN ('shipped', 'partially_received') THEN
        RETURN QUERY SELECT false, 'Transfer request must be shipped or partially_received to receive.';
        RETURN;
    END IF;

    FOR v_item IN SELECT * FROM transfer_request_items WHERE transfer_request_id = p_transfer_request_id FOR UPDATE LOOP
        v_qty := CASE WHEN v_item.shipped_quantity > 0 THEN v_item.shipped_quantity ELSE v_item.requested_quantity END;
        IF v_qty <= 0 THEN
            CONTINUE;
        END IF;

        -- Retrieve UOM code for conversion
        SELECT code INTO v_uom_code FROM units_of_measure WHERE id = v_item.uom_id;
        
        -- Convert requested qty to base unit qty
        v_base_qty := fn_convert_uom_quantity(v_item.product_id, v_uom_code, v_qty);
        IF v_base_qty IS NULL THEN
            v_base_qty := v_qty;
        END IF;

        -- Decrement in-transit and increment on_hand & available at destination store using v_base_qty
        UPDATE inventory_stock
        SET quantity_in_transit = GREATEST(0, quantity_in_transit - v_base_qty),
            quantity_on_hand = quantity_on_hand + v_base_qty,
            quantity_available = quantity_available + v_base_qty,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.to_store_id;

        IF NOT FOUND THEN
            INSERT INTO inventory_stock (product_id, product_variant_id, store_id, storage_location_id,
                quantity_on_hand, quantity_available, quantity_in_transit)
            VALUES (v_item.product_id, v_item.product_variant_id, v_req.to_store_id, v_item.to_location_id,
                    v_base_qty, v_base_qty, 0);
        END IF;

        -- Update item received_quantity
        UPDATE transfer_request_items
        SET received_quantity = v_qty
        WHERE id = v_item.id;

        -- Record stock movement (transfer_in / completed) using v_base_qty
        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            from_store_id, to_store_id, from_location_id, to_location_id,
            quantity, uom_id, batch_number, posted_by, status, metadata
        ) VALUES (
            'transfer_in', 'transfer_request', p_transfer_request_id, v_item.product_id, v_item.product_variant_id,
            v_req.from_store_id, v_req.to_store_id, v_item.from_location_id, v_item.to_location_id,
            v_base_qty, v_item.uom_id, v_item.batch_number, p_received_by, 'completed',
            jsonb_build_object('transfer_number', v_req.transfer_number)
        );
    END LOOP;

    -- Update header
    UPDATE transfer_requests
    SET status = 'received',
        received_by = p_received_by,
        received_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_transfer_request_id;

    RETURN QUERY SELECT true, 'Transfer request received successfully.';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_process_goods_receipt(
    p_grn_id INTEGER
)
RETURNS TABLE (
    success BOOLEAN,
    message TEXT
) AS $$
DECLARE
    v_grn RECORD;
    v_item RECORD;
    v_all_po_received BOOLEAN := true;
    v_any_po_received BOOLEAN := false;
    v_po_line RECORD;
BEGIN
    SELECT * INTO v_grn FROM goods_receipt_notes WHERE id = p_grn_id FOR UPDATE;
    IF v_grn IS NULL THEN
        RETURN QUERY SELECT false, 'Goods receipt note not found.';
        RETURN;
    END IF;

    IF v_grn.status = 'completed' THEN
        RETURN QUERY SELECT false, 'Goods receipt note is already completed.';
        RETURN;
    END IF;

    FOR v_item IN SELECT * FROM goods_receipt_note_items WHERE grn_id = p_grn_id FOR UPDATE LOOP
        IF v_item.quantity_received <= 0 THEN
            CONTINUE;
        END IF;

        -- Update PO Line if associated
        IF v_item.purchase_order_line_id IS NOT NULL THEN
            UPDATE purchase_order_lines
            SET received_quantity = received_quantity + v_item.quantity_received
            WHERE id = v_item.purchase_order_line_id;
        END IF;

        -- Increment inventory stock at store
        UPDATE inventory_stock
        SET quantity_on_hand = quantity_on_hand + v_item.quantity_received,
            quantity_available = quantity_available + v_item.quantity_received,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_grn.store_id;

        IF NOT FOUND THEN
            INSERT INTO inventory_stock (
                product_id, product_variant_id, store_id, storage_location_id,
                quantity_on_hand, quantity_available
            ) VALUES (
                v_item.product_id, v_item.product_variant_id, v_grn.store_id, v_item.storage_location_id,
                v_item.quantity_received, v_item.quantity_received
            );
        END IF;

        -- Update/Insert product batch if batch number is present
        IF v_item.batch_number IS NOT NULL AND v_item.batch_number <> '' THEN
            UPDATE product_batches
            SET quantity_available = quantity_available + v_item.quantity_received,
                expiry_date = COALESCE(v_item.expiry_date, expiry_date),
                updated_at = CURRENT_TIMESTAMP
            WHERE product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND store_id = v_grn.store_id
              AND batch_number = v_item.batch_number;

            IF NOT FOUND THEN
                INSERT INTO product_batches (
                    product_id, product_variant_id, batch_number,
                    expiry_date, store_id, quantity_available, status
                ) VALUES (
                    v_item.product_id, v_item.product_variant_id, v_item.batch_number,
                    v_item.expiry_date, v_grn.store_id, v_item.quantity_received, 'active'
                );
            END IF;
        END IF;

        -- Insert stock movement (purchase_receipt)
        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            to_store_id, to_location_id, quantity, uom_id, batch_number,
            posted_by, status, cost_per_unit, total_value, metadata
        ) VALUES (
            'purchase_receipt', 'goods_receipt_note', p_grn_id, v_item.product_id, v_item.product_variant_id,
            v_grn.store_id, v_item.storage_location_id, v_item.quantity_received, v_item.uom_id, COALESCE(v_item.batch_number, ''),
            v_grn.received_by, 'completed', COALESCE(v_item.unit_cost, 0), (COALESCE(v_item.unit_cost, 0) * v_item.quantity_received),
            jsonb_build_object('grn_number', v_grn.grn_number, 'delivery_note_number', COALESCE(v_grn.delivery_note_number, ''))
        );
    END LOOP;

    -- Update GRN status
    UPDATE goods_receipt_notes
    SET status = 'completed',
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_grn_id;

    -- Update Purchase Order status if PO ID present
    IF v_grn.purchase_order_id IS NOT NULL THEN
        FOR v_po_line IN SELECT quantity, received_quantity FROM purchase_order_lines WHERE purchase_order_id = v_grn.purchase_order_id LOOP
            IF v_po_line.received_quantity < v_po_line.quantity THEN
                v_all_po_received := false;
            END IF;
            IF v_po_line.received_quantity > 0 THEN
                v_any_po_received := true;
            END IF;
        END LOOP;

        IF v_all_po_received THEN
            UPDATE purchase_orders SET status = 'received', updated_at = CURRENT_TIMESTAMP WHERE id = v_grn.purchase_order_id;
        ELSIF v_any_po_received THEN
            UPDATE purchase_orders SET status = 'partially_received', updated_at = CURRENT_TIMESTAMP WHERE id = v_grn.purchase_order_id;
        END IF;
    END IF;

    RETURN QUERY SELECT true, 'Goods receipt processed successfully.';
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- FIX #3 (P0): Auto stock adjustment after physical count
-- =====================================================

CREATE OR REPLACE FUNCTION fn_reconcile_stock_count(p_count_id INTEGER)
RETURNS TABLE (
    success       BOOLEAN,
    message       TEXT,
    lines_updated INTEGER
) AS $$
DECLARE
    v_count        RECORD;
    v_line         RECORD;
    v_lines_updated INTEGER := 0;
BEGIN
    SELECT * INTO v_count FROM stock_counts WHERE id = p_count_id;

    IF NOT FOUND THEN
        RETURN QUERY SELECT false, 'Stock count not found.', 0;
        RETURN;
    END IF;
    IF v_count.status <> 'approved' THEN
        RETURN QUERY SELECT false, 'Stock count must be in ''approved'' status to reconcile.', 0;
        RETURN;
    END IF;

    FOR v_line IN
        SELECT * FROM stock_count_lines WHERE stock_count_id = p_count_id
    LOOP
        -- Update inventory_stock to match counted quantity
        UPDATE inventory_stock
        SET quantity_on_hand   = v_line.counted_quantity,
            quantity_available = GREATEST(0, v_line.counted_quantity - COALESCE(quantity_allocated, 0)),
            last_counted_at    = CURRENT_TIMESTAMP,
            updated_at         = CURRENT_TIMESTAMP
        WHERE product_id = v_line.product_id
          AND (product_variant_id = v_line.product_variant_id
               OR (product_variant_id IS NULL AND v_line.product_variant_id IS NULL))
          AND store_id = v_count.store_id;

        -- Record the adjustment movement
        INSERT INTO stock_movements (movement_type, reference_type, reference_id, product_id,
            product_variant_id, to_store_id, quantity, batch_number, serial_number,
            posted_by, status, metadata)
        VALUES ('count_adjustment', 'stock_count', p_count_id, v_line.product_id,
                v_line.product_variant_id, v_count.store_id,
                v_line.counted_quantity - v_line.system_quantity,
                v_line.batch_number, v_line.serial_number,
                v_count.approved_by, 'completed',
                jsonb_build_object('count_id', p_count_id, 'variance', v_line.variance));

        v_lines_updated := v_lines_updated + 1;
    END LOOP;

    -- Mark count as reconciled
    UPDATE stock_counts SET status = 'approved', updated_at = CURRENT_TIMESTAMP
    WHERE id = p_count_id;

    RETURN QUERY SELECT true, format('Reconciliation complete. %s lines adjusted.', v_lines_updated), v_lines_updated;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- FIX #27 (P0): Auto stock deduction on order fulfillment
-- =====================================================

CREATE OR REPLACE FUNCTION fn_trigger_deduct_inventory_on_fulfillment()
RETURNS TRIGGER AS $$
DECLARE
    v_order_line RECORD;
    v_fulfilled_qty DECIMAL(15,3);
    v_reservation RECORD;
BEGIN
    -- Only process when order status changes to 'fulfilled'
    IF NOT (OLD.order_status IS DISTINCT FROM NEW.order_status AND NEW.order_status = 'fulfilled') THEN
        RETURN NEW;
    END IF;
        
    -- Ensure store_id is present
    IF NEW.store_id IS NULL THEN
        RAISE WARNING 'Order % has no store_id, skipping inventory deduction', NEW.id;
        RETURN NEW;
    END IF;

    -- Loop through all order lines
    FOR v_order_line IN
        SELECT 
            id,
            product_id,
            product_variant_id,
            quantity_ordered,
            quantity_fulfilled,
            uom_id
        FROM sales_order_lines_v2
        WHERE sales_order_id = NEW.id
    LOOP
        -- Determine the fulfilled quantity
        IF v_order_line.quantity_fulfilled IS NOT NULL AND v_order_line.quantity_fulfilled > 0 THEN
            v_fulfilled_qty := v_order_line.quantity_fulfilled;
        ELSE
            v_fulfilled_qty := v_order_line.quantity_ordered;
        END IF;
        
        IF v_fulfilled_qty <= 0 THEN
            CONTINUE;
        END IF;

        -- FULFILLMENT: Deduct from on-hand and reduce allocated
        UPDATE inventory_stock
        SET 
            quantity_on_hand = quantity_on_hand - v_fulfilled_qty,
            quantity_allocated = GREATEST(0, quantity_allocated - v_fulfilled_qty),
            quantity_available = GREATEST(0, 
                (quantity_on_hand - v_fulfilled_qty) - 
                GREATEST(0, quantity_allocated - v_fulfilled_qty)
            ),
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_order_line.product_id
          AND (product_variant_id = v_order_line.product_variant_id 
               OR (product_variant_id IS NULL AND v_order_line.product_variant_id IS NULL))
          AND store_id = NEW.store_id;

        IF NOT FOUND THEN
            RAISE WARNING 'No inventory_stock record found for product_id=%, product_variant_id=%, store_id=%. Movement recorded but stock not updated.',
                v_order_line.product_id, v_order_line.product_variant_id, NEW.store_id;
        END IF;

        -- Update product batch if batch number is present
        IF v_order_line.batch_number IS NOT NULL AND v_order_line.batch_number <> '' THEN
            UPDATE product_batches
            SET quantity_available = GREATEST(0, quantity_available - v_fulfilled_qty),
                updated_at = CURRENT_TIMESTAMP
            WHERE product_id = v_order_line.product_id
              AND (product_variant_id = v_order_line.product_variant_id 
                   OR (product_variant_id IS NULL AND v_order_line.product_variant_id IS NULL))
              AND store_id = NEW.store_id
              AND batch_number = v_order_line.batch_number;
        END IF;

        -- Record the stock movement for auditing
        INSERT INTO stock_movements (
            movement_type,
            reference_type,
            reference_id,
            product_id,
            product_variant_id,
            from_store_id,
            quantity,
            uom_id,
            status,
            metadata
        )
        VALUES (
            'sale',
            'sales_order',
            NULL,
            v_order_line.product_id,
            v_order_line.product_variant_id,
            NEW.store_id,
            v_fulfilled_qty,
            v_order_line.uom_id,
            'completed',
            jsonb_build_object(
                'sales_order_id', NEW.id::TEXT,
                'sales_order_number', NEW.order_number,
                'order_line_id', v_order_line.id::TEXT,
                'order_status', NEW.order_status
            )
        );

        -- Mark active reservations as 'fulfilled' when order is fulfilled
        FOR v_reservation IN
            SELECT id, quantity_reserved
            FROM stock_reservations
            WHERE reference_type = 'sales_order'
              AND reference_id = NEW.id::TEXT
              AND product_id = v_order_line.product_id
              AND (product_variant_id = v_order_line.product_variant_id 
                   OR (product_variant_id IS NULL AND v_order_line.product_variant_id IS NULL))
              AND store_id = NEW.store_id
              AND status = 'active'
        LOOP
            UPDATE stock_reservations
            SET 
                status = 'fulfilled',
                updated_at = CURRENT_TIMESTAMP
            WHERE id = v_reservation.id;
        END LOOP;

    END LOOP;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for UPDATE: fires when order status changes to 'fulfilled'
CREATE TRIGGER trg_deduct_inventory_on_fulfillment
    AFTER UPDATE ON sales_orders_v2
    FOR EACH ROW
    WHEN (OLD.order_status IS DISTINCT FROM NEW.order_status AND NEW.order_status = 'fulfilled')
    EXECUTE FUNCTION fn_trigger_deduct_inventory_on_fulfillment();

-- =====================================================
-- Trigger for POS Transaction Line insertion: deduct stock automatically
-- =====================================================

CREATE OR REPLACE FUNCTION fn_trigger_deduct_inventory_on_pos_transaction()
RETURNS TRIGGER AS $$
DECLARE
    v_store_id INTEGER;
    v_status VARCHAR(50);
BEGIN
    -- Get the store ID and status from the parent transaction
    SELECT store_id, status INTO v_store_id, v_status FROM pos_transactions WHERE id = NEW.transaction_id;

    IF v_store_id IS NULL OR v_status = 'voided' THEN
        RETURN NEW;
    END IF;

    -- 1. Deduct from general inventory stock
    UPDATE inventory_stock
    SET quantity_on_hand = quantity_on_hand - NEW.quantity,
        quantity_available = GREATEST(0, quantity_available - NEW.quantity),
        updated_at = CURRENT_TIMESTAMP
    WHERE product_id = NEW.product_id
      AND (product_variant_id = NEW.product_variant_id OR (product_variant_id IS NULL AND NEW.product_variant_id IS NULL))
      AND store_id = v_store_id;

    -- 2. Deduct from product batch if batch number is present
    IF NEW.batch_number IS NOT NULL AND NEW.batch_number <> '' THEN
        UPDATE product_batches
        SET quantity_available = GREATEST(0, quantity_available - NEW.quantity),
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = NEW.product_id
          AND (product_variant_id = NEW.product_variant_id OR (product_variant_id IS NULL AND NEW.product_variant_id IS NULL))
          AND store_id = v_store_id
          AND batch_number = NEW.batch_number;
    END IF;

    -- 3. Record a stock movement for POS transaction line
    INSERT INTO stock_movements (
        movement_type, reference_type, reference_id, product_id, product_variant_id,
        from_store_id, quantity, uom_id, status, metadata
    )
    VALUES (
        'sale', 'pos_transaction', NEW.transaction_id, NEW.product_id, NEW.product_variant_id,
        v_store_id, NEW.quantity, NEW.uom_id, 'completed',
        jsonb_build_object('pos_line_id', NEW.id, 'batch_number', COALESCE(NEW.batch_number, ''))
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_deduct_inventory_on_pos_transaction ON pos_transaction_lines;
CREATE TRIGGER trg_deduct_inventory_on_pos_transaction
    AFTER INSERT ON pos_transaction_lines
    FOR EACH ROW
    EXECUTE FUNCTION fn_trigger_deduct_inventory_on_pos_transaction();

-- =====================================================
-- Trigger for order line insertion: allocate stock when order is pending/confirmed
-- =====================================================

CREATE OR REPLACE FUNCTION fn_trigger_allocate_inventory_on_order_line()
RETURNS TRIGGER AS $$
DECLARE
    v_order RECORD;
    v_quantity DECIMAL(15,3);
BEGIN
    -- Get the parent order
    SELECT order_status, store_id INTO v_order
    FROM sales_orders_v2
    WHERE id = NEW.sales_order_id;

    -- Only process if order status is 'pending' or 'confirmed'
    IF v_order.order_status IN ('pending', 'confirmed') AND v_order.store_id IS NOT NULL THEN
        v_quantity := NEW.quantity_ordered;
        
        IF v_quantity > 0 THEN
            -- Allocate stock (increase allocated, decrease available)
            UPDATE inventory_stock
            SET 
                quantity_allocated = quantity_allocated + v_quantity,
                quantity_available = GREATEST(0, quantity_on_hand - (quantity_allocated + v_quantity)),
                updated_at = CURRENT_TIMESTAMP
            WHERE product_id = NEW.product_id
              AND (product_variant_id = NEW.product_variant_id 
                   OR (product_variant_id IS NULL AND NEW.product_variant_id IS NULL))
              AND store_id = v_order.store_id;

            -- Record the allocation movement
            INSERT INTO stock_movements (
                movement_type,
                reference_type,
                reference_id,
                product_id,
                product_variant_id,
                from_store_id,
                quantity,
                uom_id,
                status,
                metadata
            )
            VALUES (
                'allocation',
                'sales_order',
                NULL,
                NEW.product_id,
                NEW.product_variant_id,
                v_order.store_id,
                v_quantity,
                NEW.uom_id,
                'completed',
                jsonb_build_object(
                    'sales_order_id', NEW.sales_order_id::TEXT,
                    'order_line_id', NEW.id::TEXT,
                    'order_status', v_order.order_status
                )
            );
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_allocate_inventory_on_order_line_insert
    AFTER INSERT ON sales_order_lines_v2
    FOR EACH ROW
    EXECUTE FUNCTION fn_trigger_allocate_inventory_on_order_line();

-- =====================================================
-- FIX #26 (P1): Loyalty points earning calculation
-- =====================================================

CREATE OR REPLACE FUNCTION fn_calculate_loyalty_earned(p_transaction_id INTEGER)
RETURNS TABLE (
    points_earned  DECIMAL(15,2),
    rule_applied   VARCHAR(255),
    customer_id    INTEGER
) AS $$
DECLARE
    v_txn        RECORD;
    v_rule       RECORD;
    v_points     DECIMAL(15,2) := 0;
BEGIN
    SELECT pt.*, pt.total_amount, pt.customer_id AS cust_id
    INTO v_txn
    FROM pos_transactions pt
    WHERE pt.id = p_transaction_id;

    IF NOT FOUND OR v_txn.cust_id IS NULL THEN
        RETURN QUERY SELECT 0::DECIMAL(15,2), 'No customer on transaction'::VARCHAR(255), NULL::INTEGER;
        RETURN;
    END IF;

    SELECT * INTO v_rule
    FROM loyalty_redemption_rules
    WHERE is_active = true
      AND (valid_from IS NULL OR valid_from <= CURRENT_DATE)
      AND (valid_to   IS NULL OR valid_to   >= CURRENT_DATE)
    ORDER BY id DESC LIMIT 1;

    IF NOT FOUND THEN
        RETURN QUERY SELECT 0::DECIMAL(15,2), 'No active loyalty rule found'::VARCHAR(255), v_txn.cust_id;
        RETURN;
    END IF;

    v_points := FLOOR(v_txn.total_amount * v_rule.points_earning_rate);

    -- Update customer loyalty_points balance
    UPDATE customers
    SET loyalty_points = loyalty_points + v_points,
        updated_at     = CURRENT_TIMESTAMP
    WHERE id = v_txn.cust_id;

    RETURN QUERY SELECT v_points, v_rule.rule_name::VARCHAR(255), v_txn.cust_id;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- FIX #23 (P1): Daily analytics refresh function
-- =====================================================

CREATE OR REPLACE FUNCTION fn_refresh_daily_analytics(p_date DATE DEFAULT CURRENT_DATE)
RETURNS VOID AS $$
BEGIN
    -- Refresh sales_analytics from pos_transactions and sales_orders_v2
    INSERT INTO sales_analytics (
        organization_id, store_id, product_id, category_id,
        date, hour, day_of_week, month, quarter, year,
        units_sold, revenue, discounts, taxes, net_revenue, transactions,
        payment_method, payment_gateway
    )
    SELECT
        organization_id, store_id, product_id, category_id,
        date, hour, day_of_week, month, quarter, year,
        SUM(units_sold), SUM(revenue), SUM(discounts), SUM(taxes), SUM(net_revenue), SUM(transactions),
        payment_method, payment_gateway
    FROM (
        -- Data from POS transactions
        SELECT
            s.organization_id,
            pt.store_id,
            ptl.product_id,
            p.category_id,
            p_date AS date,
            EXTRACT(HOUR FROM pt.transaction_date)::INTEGER AS hour,
            EXTRACT(DOW  FROM pt.transaction_date)::INTEGER AS day_of_week,
            EXTRACT(MONTH FROM p_date)::INTEGER AS month,
            EXTRACT(QUARTER FROM p_date)::INTEGER AS quarter,
            EXTRACT(YEAR FROM p_date)::INTEGER AS year,
            ptl.quantity AS units_sold,
            ptl.line_total AS revenue,
            ptl.discount_amount AS discounts,
            ptl.tax_amount AS taxes,
            (ptl.line_total - ptl.tax_amount) AS net_revenue,
            1 AS transactions,
            pp.payment_method,
            pp.payment_gateway
        FROM pos_transactions pt
        JOIN pos_transaction_lines ptl ON ptl.transaction_id = pt.id
        LEFT JOIN pos_payments pp ON pp.transaction_id = pt.id
        JOIN products p ON p.id = ptl.product_id
        JOIN stores s ON s.id = pt.store_id
        WHERE DATE(pt.transaction_date) = p_date
          AND pt.status = 'completed'

        UNION ALL

        -- Data from sales_orders_v2 (excluding those that were synced to POS to avoid double counting)
        SELECT
            o.organization_id,
            o.store_id,
            ol.product_id,
            p.category_id,
            p_date AS date,
            EXTRACT(HOUR FROM o.order_date)::INTEGER AS hour,
            EXTRACT(DOW  FROM o.order_date)::INTEGER AS day_of_week,
            EXTRACT(MONTH FROM p_date)::INTEGER AS month,
            EXTRACT(QUARTER FROM p_date)::INTEGER AS quarter,
            EXTRACT(YEAR FROM p_date)::INTEGER AS year,
            ol.quantity_ordered::DECIMAL(15,3) AS units_sold,
            ol.line_total AS revenue,
            COALESCE(ol.discount_amount, 0) AS discounts,
            COALESCE(ol.tax_amount, 0) AS taxes,
            (ol.line_total - COALESCE(ol.tax_amount, 0)) AS net_revenue,
            1 AS transactions,
            o.payment_method,
            o.payment_gateway
        FROM sales_orders_v2 o
        JOIN sales_order_lines_v2 ol ON ol.sales_order_id = o.id
        JOIN products p ON p.id = ol.product_id
        WHERE DATE(o.order_date) = p_date
          AND o.order_status IN ('confirmed', 'processing', 'partially_fulfilled', 'fulfilled', 'shipped', 'delivered')
          AND NOT EXISTS (SELECT 1 FROM pos_transactions WHERE sales_order_id = o.id)
    ) aggregated_sales
    GROUP BY organization_id, store_id, product_id, category_id,
             date, hour, day_of_week, month, quarter, year,
             payment_method, payment_gateway
    ON CONFLICT (organization_id, store_id, product_id, date, hour, payment_method, payment_gateway) 
    DO UPDATE SET
        units_sold = EXCLUDED.units_sold,
        revenue = EXCLUDED.revenue,
        discounts = EXCLUDED.discounts,
        taxes = EXCLUDED.taxes,
        net_revenue = EXCLUDED.net_revenue,
        transactions = EXCLUDED.transactions,
        updated_at = CURRENT_TIMESTAMP;

    -- Refresh profit_loss_analytics
    INSERT INTO profit_loss_analytics (
        organization_id, store_id, date, period_type, month, quarter, year,
        gross_revenue, sales_discounts, net_revenue, cogs
    )
    SELECT
        s.organization_id,
        pt.store_id,
        p_date,
        'daily',
        EXTRACT(MONTH FROM p_date)::INTEGER,
        EXTRACT(QUARTER FROM p_date)::INTEGER,
        EXTRACT(YEAR FROM p_date)::INTEGER,
        SUM(pt.total_amount),
        SUM(pt.discount_amount),
        SUM(pt.total_amount - pt.discount_amount),
        SUM(pt.total_cost)
    FROM pos_transactions pt
    JOIN stores s ON s.id = pt.store_id
    WHERE DATE(pt.transaction_date) = p_date
      AND pt.status = 'completed'
    GROUP BY s.organization_id, pt.store_id
    ON CONFLICT DO NOTHING;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_calculate_recipe_cost(
    p_recipe_id INTEGER
)
RETURNS NUMERIC AS $$
DECLARE
    v_total_cost NUMERIC := 0;
BEGIN
    SELECT COALESCE(SUM(vb.line_cost_estimate), 0)
      INTO v_total_cost
    FROM vw_recipe_bom vb
    WHERE vb.recipe_id    = p_recipe_id
      AND vb.is_byproduct = false
      AND vb.is_optional  = false;

    RETURN v_total_cost;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_get_waste_report(
    p_store_id      INTEGER,
    p_from_date     DATE,
    p_to_date       DATE,
    p_waste_source  VARCHAR DEFAULT NULL
)
RETURNS TABLE (
    waste_date          DATE,
    waste_source        VARCHAR,
    product_id          INTEGER,
    product_name        VARCHAR,
    menu_item_id        INTEGER,
    menu_item_name      VARCHAR,
    quantity            NUMERIC,
    uom_code            VARCHAR,
    total_cost          NUMERIC,
    reason              TEXT,
    logged_by_name      VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        DATE(wl.wasted_at),
        wl.waste_source::VARCHAR,
        wl.product_id,
        p.name::VARCHAR,
        wl.menu_item_id,
        mi.name::VARCHAR,
        wl.quantity,
        uom.code::VARCHAR,
        wl.total_cost,
        wl.reason,
        (u.first_name || ' ' || u.last_name)::VARCHAR
    FROM waste_logs wl
    LEFT JOIN products p            ON wl.product_id   = p.id
    LEFT JOIN menu_items mi         ON wl.menu_item_id = mi.id
    LEFT JOIN units_of_measure uom  ON wl.uom_id       = uom.id
    LEFT JOIN users u               ON wl.logged_by    = u.id
    WHERE wl.store_id           = p_store_id
      AND DATE(wl.wasted_at)    BETWEEN p_from_date AND p_to_date
      AND (p_waste_source IS NULL OR wl.waste_source = p_waste_source)
    ORDER BY wl.wasted_at DESC;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_get_kds_orders(
    p_store_id      INTEGER,
    p_statuses      VARCHAR[] DEFAULT ARRAY['pending','confirmed','preparing']
)
RETURNS TABLE (
    order_id            INTEGER,
    order_number        VARCHAR,
    table_number        VARCHAR,
    waiter_name         VARCHAR,
    order_status        VARCHAR,
    ordered_at          TIMESTAMP,
    minutes_elapsed     NUMERIC,
    item_id             INTEGER,
    item_name           VARCHAR,
    item_short_name     VARCHAR,
    item_qty            NUMERIC,
    item_notes          TEXT,
    item_modifiers      JSONB,
    item_status         VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ro.id,
        ro.order_number::VARCHAR,
        rt.table_number::VARCHAR,
        (u.first_name || ' ' || u.last_name)::VARCHAR,
        ro.status::VARCHAR,
        ro.ordered_at,
        ROUND(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - ro.ordered_at)) / 60.0, 1),
        roi.id,
        mi.name::VARCHAR,
        mi.short_name::VARCHAR,
        roi.quantity,
        roi.notes,
        roi.modifiers_snapshot,
        roi.status::VARCHAR
    FROM restaurant_orders ro
    LEFT JOIN restaurant_tables rt      ON ro.table_id = rt.id
    LEFT JOIN cashiers c                ON ro.cashier_id = c.id
    LEFT JOIN users u                   ON c.user_id = u.id
    JOIN  restaurant_order_items roi    ON ro.id = roi.order_id
    JOIN  menu_items mi                 ON roi.menu_item_id = mi.id
    WHERE ro.store_id = p_store_id
      AND ro.status = ANY(p_statuses)
    ORDER BY ro.ordered_at, roi.line_number;
END;
$$ LANGUAGE plpgsql;


-- =====================================================

CREATE OR REPLACE FUNCTION fn_log_transfer_request_history()
RETURNS TRIGGER AS $$
DECLARE
    v_history_entry JSONB;
    v_items_array JSONB := '[]'::jsonb;
    v_item RECORD;
    v_user_id INTEGER;
    v_user_name VARCHAR;
    v_from_store_name VARCHAR;
    v_to_store_name VARCHAR;
    
    -- Quantities
    v_qty_moved NUMERIC;
    v_base_qty_moved NUMERIC;
    v_base_uom_code VARCHAR;
    
    -- Stock tracking variables
    v_from_before_on_hand NUMERIC := 0;
    v_from_after_on_hand NUMERIC := 0;
    v_to_before_on_hand NUMERIC := 0;
    v_to_after_on_hand NUMERIC := 0;
    v_to_before_transit NUMERIC := 0;
    v_to_after_transit NUMERIC := 0;
BEGIN
    -- If it's an UPDATE, and status did not change, skip history logging
    IF TG_OP = 'UPDATE' AND OLD.status = NEW.status THEN
        RETURN NEW;
    END IF;

    -- Ensure metadata block is initialized as a JSON object (avoids "cannot set path in scalar")
    IF NEW.metadata IS NULL OR jsonb_typeof(NEW.metadata) != 'object' THEN
        NEW.metadata := '{}'::jsonb;
    END IF;

    -- Determine who changed the status based on current state
    IF NEW.status = 'draft' THEN v_user_id := NEW.requested_by;
    ELSIF NEW.status = 'approved' THEN v_user_id := NEW.approved_by;
    ELSIF NEW.status = 'shipped' THEN v_user_id := NEW.shipped_by;
    ELSIF NEW.status = 'received' THEN v_user_id := NEW.received_by;
    ELSE v_user_id := COALESCE(NEW.approved_by, NEW.requested_by);
    END IF;

    -- Fetch user name and store names for auditing
    SELECT username INTO v_user_name FROM users WHERE id = v_user_id;
    SELECT name INTO v_from_store_name FROM stores WHERE id = NEW.from_store_id;
    SELECT name INTO v_to_store_name FROM stores WHERE id = NEW.to_store_id;

    -- Loop through all products/items linked to this transfer request
    FOR v_item IN (
        SELECT tri.*, p.sku, p.name AS prod_name, u.code AS req_uom_code, p.base_uom_id
        FROM transfer_request_items tri
        JOIN products p ON tri.product_id = p.id
        LEFT JOIN units_of_measure u ON tri.uom_id = u.id
        WHERE tri.transfer_request_id = NEW.id
    ) LOOP
        -- Retrieve base UOM code
        SELECT code INTO v_base_uom_code FROM units_of_measure WHERE id = v_item.base_uom_id;

        -- Get current live stock at sending Store (Source)
        SELECT COALESCE(quantity_on_hand, 0)
        INTO v_from_after_on_hand
        FROM inventory_stock
        WHERE product_id = v_item.product_id 
          AND store_id = NEW.from_store_id 
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL));

        -- Get current live stock at receiving Store (Destination)
        SELECT COALESCE(quantity_on_hand, 0), COALESCE(quantity_in_transit, 0)
        INTO v_to_after_on_hand, v_to_after_transit
        FROM inventory_stock
        WHERE product_id = v_item.product_id 
          AND store_id = NEW.to_store_id 
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL));

        -- Resolve the quantity moved in the current state transition
        IF NEW.status = 'shipped' THEN
            v_qty_moved := COALESCE(v_item.shipped_quantity, v_item.requested_quantity);
        ELSIF NEW.status = 'received' THEN
            v_qty_moved := COALESCE(v_item.received_quantity, v_item.shipped_quantity);
        ELSE
            v_qty_moved := v_item.requested_quantity;
        END IF;

        -- Convert to base unit quantity for inventory calculations
        v_base_qty_moved := fn_convert_uom_quantity(v_item.product_id, COALESCE(v_item.req_uom_code, ''), v_qty_moved);
        IF v_base_qty_moved IS NULL THEN
            v_base_qty_moved := v_qty_moved;
        END IF;

        -- Calculate predicted stock outcomes based on transition states using base quantities
        IF NEW.status = 'shipped' THEN
            -- Shipped: Stock has been deducted from source, and added to transit at destination
            v_from_before_on_hand := v_from_after_on_hand + v_base_qty_moved;
            
            v_to_before_transit   := v_to_after_transit - v_base_qty_moved;
            v_to_before_on_hand   := v_to_after_on_hand;
        ELSIF NEW.status = 'received' THEN
            -- Received: Stock was moved from transit to on_hand at destination
            v_from_before_on_hand := v_from_after_on_hand;
            v_from_after_on_hand  := v_from_after_on_hand;
            
            v_to_before_transit   := v_to_after_transit + v_base_qty_moved;
            v_to_before_on_hand   := v_to_after_on_hand - v_base_qty_moved;
        ELSE
            -- Default/Approval state: Physical stock levels unchanged
            v_from_before_on_hand := v_from_after_on_hand;
            v_to_before_transit   := v_to_after_transit;
            v_to_before_on_hand   := v_to_after_on_hand;
        END IF;

        -- Append item audit blocks to array
        v_items_array := v_items_array || jsonb_build_array(
            jsonb_build_object(
                'product_id', v_item.product_id,
                'product_variant_id', v_item.product_variant_id,
                'sku', v_item.sku,
                'product_name', v_item.prod_name,
                'requested_quantity', v_item.requested_quantity,
                'shipped_quantity', v_item.shipped_quantity,
                'received_quantity', v_item.received_quantity,
                'uom', COALESCE(v_item.req_uom_code, ''),
                'converted_base_quantity', v_base_qty_moved,
                'base_uom', COALESCE(v_base_uom_code, ''),
                'inventory_snapshot', jsonb_build_object(
                    'source_store', jsonb_build_object(
                        'store_name', v_from_store_name,
                        'before_on_hand', COALESCE(v_from_before_on_hand, 0),
                        'after_on_hand', COALESCE(v_from_after_on_hand, 0),
                        'deducted', CASE WHEN NEW.status = 'shipped' THEN v_base_qty_moved ELSE 0 END
                    ),
                    'destination_store', jsonb_build_object(
                        'store_name', v_to_store_name,
                        'before_on_hand', COALESCE(v_to_before_on_hand, 0),
                        'after_on_hand', COALESCE(v_to_after_on_hand, 0),
                        'before_in_transit', COALESCE(v_to_before_transit, 0),
                        'after_in_transit', COALESCE(v_to_after_transit, 0),
                        'added_received', CASE WHEN NEW.status = 'received' THEN v_base_qty_moved ELSE 0 END
                    )
                )
            )
        );
    END LOOP;

    -- Build the history state entry
    v_history_entry := jsonb_build_object(
        'status', NEW.status,
        'changed_at', CURRENT_TIMESTAMP,
        'user_details', jsonb_build_object('id', v_user_id, 'username', COALESCE(v_user_name, 'system')),
        'notes', COALESCE(NEW.notes, ''),
        'transfer_items_snapshot', v_items_array
    );

    -- Build nested history array key
    IF NOT (NEW.metadata ? 'history') THEN
        NEW.metadata := jsonb_set(NEW.metadata, '{history}', '[]'::jsonb);
    END IF;

    -- Append the state audit to the history queue
    NEW.metadata := jsonb_set(
        NEW.metadata, 
        '{history}', 
        (NEW.metadata->'history') || jsonb_build_array(v_history_entry)
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_transfer_request_history ON transfer_requests;
CREATE TRIGGER trg_transfer_request_history
BEFORE INSERT OR UPDATE ON transfer_requests
FOR EACH ROW
EXECUTE FUNCTION fn_log_transfer_request_history();

CREATE OR REPLACE FUNCTION fn_sync_promotion_to_product_prices()
RETURNS TRIGGER AS $$
DECLARE
    v_promo_pl_id INTEGER;
    v_target_product_id INTEGER;
    v_retail_pp RECORD;
    v_calculated_price NUMERIC(15,2);
    v_discount_percent_str VARCHAR;
    v_variant_id INTEGER;
BEGIN
    -- Locate existing PROMO price list
    SELECT id INTO v_promo_pl_id
    FROM price_lists
    WHERE (code = 'PROMO' OR price_list_type = 'promotional')
      AND is_active = true
    ORDER BY id ASC
    LIMIT 1;

    -- If no PROMO price list exists, dynamically create one
    IF v_promo_pl_id IS NULL AND (TG_OP = 'INSERT' OR TG_OP = 'UPDATE') THEN
        INSERT INTO price_lists (code, name, price_list_type, currency_code, is_active)
        VALUES (
            'PROMO', 
            'Promotional Price List', 
            'promotional', 
            'SAR', 
            true
        )
        ON CONFLICT (code) DO UPDATE SET is_active = true
        RETURNING id INTO v_promo_pl_id;
    END IF;

    -- If DELETING or DEACTIVATING promotion, remove/deactivate promo package prices
    IF TG_OP = 'DELETE' OR (TG_OP = 'UPDATE' AND NEW.is_active = false) THEN
        DELETE FROM product_prices
        WHERE price_list_id = v_promo_pl_id
          AND metadata->>'promotion_id' = COALESCE(OLD.id, NEW.id)::text;
        
        IF TG_OP = 'DELETE' THEN
            RETURN OLD;
        END IF;
        RETURN NEW;
    END IF;

    -- If INSERTING or UPDATING active promotion, generate promotional package prices
    IF NEW.is_active = true AND v_promo_pl_id IS NOT NULL THEN
        -- Clean up existing promotional prices for this promotion on UPDATE to ensure non-matching UOMs are purged
        IF TG_OP = 'UPDATE' THEN
            DELETE FROM product_prices
            WHERE price_list_id = v_promo_pl_id
              AND metadata->>'promotion_id' = NEW.id::text;
        END IF;

        -- Format discount percent string for metadata
        IF NEW.promotion_type = 'percentage_discount' AND NEW.discount_value IS NOT NULL THEN
            v_discount_percent_str := CONCAT(TRIM(TRAILING '.' FROM TRIM(TRAILING '0' FROM NEW.discount_value::text)), '%');
        ELSE
            v_discount_percent_str := NULL;
        END IF;

        -- Find all target products matching the promotion criteria
        FOR v_target_product_id IN (
            SELECT p.id
            FROM products p
            WHERE p.organization_id = NEW.organization_id
              AND p.is_active = true
              AND (
                  NEW.applies_to = 'all'
                  OR (NEW.applies_to = 'product' AND p.id = ANY(NEW.target_product_ids))
                  OR (NEW.applies_to = 'category' AND p.category_id = ANY(NEW.target_category_ids))
              )
        ) LOOP
            -- Loop through existing retail package prices for this product to compute promotional price per UOM
            FOR v_retail_pp IN (
                SELECT pp.*
                FROM product_prices pp
                JOIN price_lists pl ON pp.price_list_id = pl.id
                WHERE pp.product_id = v_target_product_id
                  AND pl.price_list_type = 'retail'
                  AND pp.is_active = true
                  AND (
                      NOT (NEW.metadata ? 'target_uoms')
                      OR NOT (NEW.metadata->'target_uoms' ? v_target_product_id::text)
                      OR (
                          jsonb_typeof(NEW.metadata->'target_uoms'->(v_target_product_id::text)) = 'array'
                          AND (NEW.metadata->'target_uoms'->(v_target_product_id::text)) @> pp.uom_id::text::jsonb
                      )
                      OR (NEW.metadata->'target_uoms'->>(v_target_product_id::text)) = pp.uom_id::text
                  )
                  AND (
                      NOT (NEW.metadata ? 'target_variants')
                      OR NOT (NEW.metadata->'target_variants' ? v_target_product_id::text)
                      OR pp.product_variant_id IS NULL
                      OR (
                          jsonb_typeof(NEW.metadata->'target_variants'->(v_target_product_id::text)) = 'array'
                          AND (NEW.metadata->'target_variants'->(v_target_product_id::text)) @> pp.product_variant_id::text::jsonb
                      )
                      OR (NEW.metadata->'target_variants'->>(v_target_product_id::text)) = pp.product_variant_id::text
                  )
            ) LOOP
                -- Calculate promo price
                IF NEW.promotion_type = 'percentage_discount' AND NEW.discount_value IS NOT NULL THEN
                    v_calculated_price := ROUND(v_retail_pp.price * (1.0 - (NEW.discount_value / 100.0)), 2);
                ELSIF NEW.promotion_type = 'fixed_discount' AND NEW.discount_value IS NOT NULL THEN
                    v_calculated_price := GREATEST(0.00, v_retail_pp.price - NEW.discount_value);
                ELSE
                    v_calculated_price := v_retail_pp.price;
                END IF;

                -- If retail price was at base product level (product_variant_id IS NULL), but promotion targets specific variants, loop over those variants
                IF v_retail_pp.product_variant_id IS NULL 
                   AND (NEW.metadata ? 'target_variants') 
                   AND (NEW.metadata->'target_variants' ? v_target_product_id::text) THEN
                    FOR v_variant_id IN (
                        SELECT (jsonb_array_elements_text(
                            CASE 
                                WHEN jsonb_typeof(NEW.metadata->'target_variants'->(v_target_product_id::text)) = 'array'
                                THEN NEW.metadata->'target_variants'->(v_target_product_id::text)
                                ELSE jsonb_build_array(NEW.metadata->'target_variants'->>(v_target_product_id::text))
                            END
                        ))::integer
                    ) LOOP
                        DELETE FROM product_prices
                        WHERE product_id = v_target_product_id
                          AND price_list_id = v_promo_pl_id
                          AND product_variant_id = v_variant_id
                          AND (uom_id = v_retail_pp.uom_id OR (uom_id IS NULL AND v_retail_pp.uom_id IS NULL));

                        INSERT INTO product_prices (
                            product_id,
                            product_variant_id,
                            price_list_id,
                            uom_id,
                            price,
                            min_quantity,
                            max_quantity,
                            valid_from,
                            valid_to,
                            is_active,
                            metadata
                        ) VALUES (
                            v_target_product_id,
                            v_variant_id,
                            v_promo_pl_id,
                            v_retail_pp.uom_id,
                            v_calculated_price,
                            v_retail_pp.min_quantity,
                            v_retail_pp.max_quantity,
                            NEW.valid_from,
                            NEW.valid_to,
                            true,
                            jsonb_build_object(
                                'promotion_id', NEW.id,
                                'promotion_name', NEW.name,
                                'discount_percent', v_discount_percent_str
                            )
                        );
                    END LOOP;
                ELSE
                    -- Remove previous promo price entry for this product + UOM + PROMO price list if it exists
                    DELETE FROM product_prices
                    WHERE product_id = v_target_product_id
                      AND price_list_id = v_promo_pl_id
                      AND (product_variant_id = v_retail_pp.product_variant_id OR (product_variant_id IS NULL AND v_retail_pp.product_variant_id IS NULL))
                      AND (uom_id = v_retail_pp.uom_id OR (uom_id IS NULL AND v_retail_pp.uom_id IS NULL));

                    -- Insert new promotional package price into product_prices
                    INSERT INTO product_prices (
                        product_id,
                        product_variant_id,
                        price_list_id,
                        uom_id,
                        price,
                        min_quantity,
                        max_quantity,
                        valid_from,
                        valid_to,
                        is_active,
                        metadata
                    ) VALUES (
                        v_target_product_id,
                        v_retail_pp.product_variant_id,
                        v_promo_pl_id,
                        v_retail_pp.uom_id,
                        v_calculated_price,
                        v_retail_pp.min_quantity,
                        v_retail_pp.max_quantity,
                        NEW.valid_from,
                        NEW.valid_to,
                        true,
                        jsonb_build_object(
                            'promotion_id', NEW.id,
                            'promotion_name', NEW.name,
                            'discount_percent', v_discount_percent_str
                        )
                    );
                END IF;
            END LOOP;
        END LOOP;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sync_promotion_to_product_prices ON promotions;
CREATE TRIGGER trg_sync_promotion_to_product_prices
AFTER INSERT OR UPDATE OR DELETE ON promotions
FOR EACH ROW
EXECUTE FUNCTION fn_sync_promotion_to_product_prices();





-- =====================================================
-- CART TRIGGERS & FUNCTIONS
-- =====================================================

CREATE OR REPLACE FUNCTION log_cart_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.cart_status IS DISTINCT FROM NEW.cart_status THEN
        INSERT INTO cart_activity_log (
            cart_id, 
            organization_id, 
            activity_type, 
            description, 
            old_value, 
            new_value
        )
        VALUES (
            NEW.id,
            NEW.organization_id,
            'status_changed',
            'Cart status changed from ' || OLD.cart_status || ' to ' || NEW.cart_status,
            jsonb_build_object('status', OLD.cart_status),
            jsonb_build_object('status', NEW.cart_status)
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_cart_activity()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE carts 
    SET last_activity_at = CURRENT_TIMESTAMP
    WHERE id = COALESCE(NEW.cart_id, OLD.cart_id);
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;


-- =====================================================
-- MASTER PRODUCT CATALOG FETCH FUNCTION
-- =====================================================

CREATE OR REPLACE VIEW v_master_product_catalog AS
SELECT 
    p.id AS product_id,
    p.organization_id,
    p.sku,
    p.name,
    p.description,
    p.product_type,
    p.is_serialized,
    p.is_batch_managed,
    p.is_active,
    p.is_sellable,
    p.is_purchasable,
    p.allow_decimal_quantity,
    p.track_inventory,
    p.metadata,
    p.created_at,
    p.updated_at,
    p.category_id,
    pc.name AS category_name,
    pc.code AS category_code,
    p.brand_id,
    b.name AS brand_name,
    b.code AS brand_code,
    p.tax_category_id,
    tc.name AS tax_category_name,
    tc.tax_rate AS tax_rate,
    tc.is_inclusive AS tax_inclusive,
    p.base_uom_id,
    uom.code AS base_uom_code,
    uom.name AS base_uom_name,
    -- Conversions
    COALESCE(
        (
            SELECT jsonb_agg(
                jsonb_build_object(
                    'id', puc.id,
                    'from_uom_id', puc.from_uom_id,
                    'from_uom_code', fuom.code,
                    'from_uom_name', fuom.name,
                    'to_uom_id', puc.to_uom_id,
                    'to_uom_code', tuom.code,
                    'to_uom_name', tuom.name,
                    'conversion_factor', puc.conversion_factor,
                    'is_default', puc.is_default
                )
            )
            FROM product_uom_conversions puc
            JOIN units_of_measure fuom ON puc.from_uom_id = fuom.id
            JOIN units_of_measure tuom ON puc.to_uom_id = tuom.id
            WHERE puc.product_id = p.id
        ),
        '[]'::jsonb
    ) AS uom_conversions,
    -- Prices
    COALESCE(
        (
            SELECT jsonb_agg(
                jsonb_build_object(
                    'id', pp.id,
                    'product_variant_id', pp.product_variant_id,
                    'price_list_id', pp.price_list_id,
                    'price_list_name', pl.name,
                    'price_list_code', pl.code,
                    'uom_id', pp.uom_id,
                    'uom_code', puom.code,
                    'uom_name', puom.name,
                    'price', pp.price,
                    'min_quantity', pp.min_quantity,
                    'max_quantity', pp.max_quantity,
                    'valid_from', pp.valid_from,
                    'valid_to', pp.valid_to,
                    'is_active', pp.is_active
                )
            )
            FROM product_prices pp
            JOIN price_lists pl ON pp.price_list_id = pl.id
            LEFT JOIN units_of_measure puom ON pp.uom_id = puom.id
            WHERE pp.product_id = p.id
        ),
        '[]'::jsonb
    ) AS prices,
    -- Variants
    COALESCE(
        (
            SELECT jsonb_agg(
                jsonb_build_object(
                    'id', pv.id,
                    'variant_sku', pv.variant_sku,
                    'variant_name', pv.variant_name,
                    'variant_attributes', pv.variant_attributes,
                    'is_active', pv.is_active,
                    'metadata', pv.metadata
                )
            )
            FROM product_variants pv
            WHERE pv.product_id = p.id
        ),
        '[]'::jsonb
    ) AS variants,
    -- Barcodes
    COALESCE(
        (
            SELECT jsonb_agg(
                jsonb_build_object(
                    'id', pb.id,
                    'product_variant_id', pb.product_variant_id,
                    'barcode', pb.barcode,
                    'barcode_type', pb.barcode_type,
                    'is_primary', pb.is_primary
                )
            )
            FROM product_barcodes pb
            WHERE pb.product_id = p.id
        ),
        '[]'::jsonb
    ) AS barcodes,
    -- Inventory
    COALESCE(
        (
            SELECT jsonb_agg(
                jsonb_build_object(
                    'stock_id', ist.id,
                    'store_id', ist.store_id,
                    'storage_location_id', ist.storage_location_id,
                    'storage_location_name', sl.name,
                    'storage_location_code', sl.code,
                    'product_variant_id', ist.product_variant_id,
                    'quantity_on_hand', ist.quantity_on_hand,
                    'quantity_available', ist.quantity_available,
                    'quantity_allocated', ist.quantity_allocated,
                    'quantity_on_order', ist.quantity_on_order,
                    'reorder_level', ist.reorder_level,
                    'reorder_quantity', ist.reorder_quantity,
                    'max_stock_level', ist.max_stock_level
                )
            )
            FROM inventory_stock ist
            LEFT JOIN storage_locations sl ON ist.storage_location_id = sl.id
            WHERE ist.product_id = p.id
        ),
        '[]'::jsonb
    ) AS inventory
FROM products p
LEFT JOIN product_categories pc ON p.category_id = pc.id
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN units_of_measure uom ON p.base_uom_id = uom.id
LEFT JOIN tax_categories tc ON p.tax_category_id = tc.id;


CREATE OR REPLACE FUNCTION get_master_product_catalog(p_organization_id INT)
RETURNS TABLE (
    product_id INT,
    sku VARCHAR(100),
    name VARCHAR(255),
    description TEXT,
    product_type VARCHAR(50),
    is_serialized BOOLEAN,
    is_batch_managed BOOLEAN,
    is_active BOOLEAN,
    is_sellable BOOLEAN,
    is_purchasable BOOLEAN,
    allow_decimal_quantity BOOLEAN,
    track_inventory BOOLEAN,
    metadata JSONB,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    category_id INT,
    category_name VARCHAR(255),
    category_code VARCHAR(50),
    brand_id INT,
    brand_name VARCHAR(255),
    brand_code VARCHAR(50),
    tax_category_id INT,
    tax_category_name VARCHAR(100),
    tax_rate DECIMAL(5,2),
    tax_inclusive BOOLEAN,
    base_uom_id INT,
    base_uom_code VARCHAR(20),
    base_uom_name VARCHAR(50),
    uom_conversions JSONB,
    prices JSONB,
    variants JSONB,
    barcodes JSONB,
    inventory JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        v.product_id,
        v.sku,
        v.name,
        v.description,
        v.product_type,
        v.is_serialized,
        v.is_batch_managed,
        v.is_active,
        v.is_sellable,
        v.is_purchasable,
        v.allow_decimal_quantity,
        v.track_inventory,
        v.metadata,
        v.created_at,
        v.updated_at,
        v.category_id,
        v.category_name,
        v.category_code,
        v.brand_id,
        v.brand_name,
        v.brand_code,
        v.tax_category_id,
        v.tax_category_name,
        v.tax_rate,
        v.tax_inclusive,
        v.base_uom_id,
        v.base_uom_code,
        v.base_uom_name,
        v.uom_conversions,
        v.prices,
        v.variants,
        v.barcodes,
        v.inventory
    FROM v_master_product_catalog v
    WHERE v.organization_id = p_organization_id;
END;
$$ LANGUAGE plpgsql;
