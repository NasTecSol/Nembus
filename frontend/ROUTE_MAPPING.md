# Complete Route Mapping Document

This document maps all routes from the navigation structure to Angular components.

## Module 1: Dashboard (module_code: dashboard)

### Menu 1: Overview (menu_code: overview)
- `/dashboard` → AdminDashboardComponent (submenu_id: 1, submenu_code: admin_dashboard)
- `/dashboard-stores` → StoreDashboardComponent (submenu_id: 2, submenu_code: store_dashboard)

### Menu 2: Analytics (menu_code: analytics)
- `/dashboard/analytics/sales` → SalesAnalyticsComponent (submenu_id: 3, submenu_code: sales_analytics)
- `/dashboard/analytics/inventory` → InventoryAnalyticsComponent (submenu_id: 4, submenu_code: inventory_analytics)

## Module 2: Tenant Management (module_code: tenants)

### Menu 1: Tenants (menu_code: tenants)
- `/admin/tenants/list` → TenantListComponent (submenu_id: 5, submenu_code: tenant_list)
- `/admin/tenants/new` → AddTenantComponent (submenu_id: 6, submenu_code: add_tenant)
- `/admin/tenants/config` → TenantConfigComponent (submenu_id: 7, submenu_code: tenant_config)

## Module 3: Organization Setup (module_code: organizations)

### Menu 1: Organizations (menu_code: organizations)
- `/admin/organizations/list` → OrgListComponent (submenu_id: 8, submenu_code: org_list)
- `/admin/organizations/new` → AddOrgComponent (submenu_id: 9, submenu_code: add_org)

## Module 4: User Management (module_code: users)

### Menu 1: Users (menu_code: users)
- `/admin/users/list` → UserListComponent (submenu_id: 10, submenu_code: user_list) ✅ EXISTS
- `/admin/users/new` → AddUserComponent (submenu_id: 11, submenu_code: add_user) ✅ EXISTS
- `/admin/users/activity` → UserActivityComponent (submenu_id: 12, submenu_code: user_activity)

### Menu 2: Roles & Permissions (menu_code: roles_permissions)
- `/admin/roles/list` → RoleListComponent (submenu_id: 13, submenu_code: role_list)
- `/admin/roles/new` → AddRoleComponent (submenu_id: 14, submenu_code: add_role)
- `/admin/roles/permissions` → PermissionMatrixComponent (submenu_id: 15, submenu_code: permission_matrix)

## Module 5: Store Management (module_code: stores)

### Menu 1: Stores (menu_code: stores)
- `/stores/list` → StoreslistComponent (submenu_id: 16, submenu_code: store_list) ✅ EXISTS
- `/stores/new` → AddStoreComponent (submenu_id: 17, submenu_code: add_store) ✅ EXISTS
- `/stores/config` → StoreConfigComponent (submenu_id: 18, submenu_code: store_config)

### Menu 2: Storage Locations (menu_code: storage_locations)
- `/stores/locations/list` → LocationListComponent (submenu_id: 19, submenu_code: location_list)
- `/stores/locations/new` → AddLocationComponent (submenu_id: 20, submenu_code: add_location)

## Module 6: Point of Sale (module_code: pos)

### Menu 1: POS Transactions (menu_code: pos_transactions)
- `/pos/transactions/list` → TransactionListComponent (submenu_id: 21, submenu_code: transaction_list)
- `/pos/transactions/new` → ProcessSaleComponent (submenu_id: 22, submenu_code: process_sale)
- `/pos/transactions/void` → VoidTransactionComponent (submenu_id: 23, submenu_code: void_transaction)

### Menu 2: POS Terminals (menu_code: pos_terminals)
- `/pos/terminals/list` → TerminalListComponent (submenu_id: 24, submenu_code: terminal_list)
- `/pos/terminals/new` → AddTerminalComponent (submenu_id: 25, submenu_code: add_terminal)

### Menu 3: POS Reports (menu_code: pos_reports)
- `/pos/reports/daily` → DailySalesComponent (submenu_id: 26, submenu_code: daily_sales)
- `/pos/reports/cashier` → CashierPerformanceComponent (submenu_id: 27, submenu_code: cashier_performance)

## Module 7: Cashier Operations (module_code: cashiers)

### Menu 1: Cashiers (menu_code: cashiers)
- `/cashiers/list` → CashierListComponent (submenu_id: 28, submenu_code: cashier_list)
- `/cashiers/new` → AddCashierComponent (submenu_id: 29, submenu_code: add_cashier)

### Menu 2: Cashier Sessions (menu_code: cashier_sessions)
- `/cashiers/sessions/active` → ActiveSessionsComponent (submenu_id: 30, submenu_code: active_sessions)
- `/cashiers/sessions/history` → SessionHistoryComponent (submenu_id: 31, submenu_code: session_history)
- `/cashiers/sessions/open` → OpenSessionComponent (submenu_id: 32, submenu_code: open_session)
- `/cashiers/sessions/close` → CloseSessionComponent (submenu_id: 33, submenu_code: close_session)

## Module 8: Inventory Management (module_code: inventory)

### Menu 1: Stock Overview (menu_code: stock_overview)
- `/inventory/overview/levels` → StockLevelsComponent (submenu_id: 34, submenu_code: stock_levels)
- `/inventory/overview/low-stock` → LowStockComponent (submenu_id: 35, submenu_code: low_stock)

### Menu 2: Stock Movements (menu_code: stock_movements)
- `/inventory/movements/history` → MovementHistoryComponent (submenu_id: 36, submenu_code: movement_history)
- `/inventory/movements/new` → RecordMovementComponent (submenu_id: 37, submenu_code: record_movement)

### Menu 3: Stock Counts (menu_code: stock_counts)
- `/inventory/counts/list` → StockCountListComponent (submenu_id: 38, submenu_code: stock_count_list)
- `/inventory/counts/new` → CreateCountComponent (submenu_id: 39, submenu_code: create_count)

## Module 9: Product Catalog (module_code: products)

### Menu 1: Products (menu_code: products)
- `/products/list` → ProductListComponent (submenu_id: 40, submenu_code: product_list) ✅ EXISTS (itemslist)
- `/products/new` → AddProductComponent (submenu_id: 41, submenu_code: add_product) ✅ EXISTS
- `/products/import` → ProductImportComponent (submenu_id: 42, submenu_code: product_import)

### Menu 2: Categories (menu_code: categories)
- `/products/categories/list` → CategoryListComponent (submenu_id: 43, submenu_code: category_list)
- `/products/categories/new` → AddCategoryComponent (submenu_id: 44, submenu_code: add_category)

### Menu 3: Brands (menu_code: brands)
- `/products/brands/list` → BrandListComponent (submenu_id: 45, submenu_code: brand_list)
- `/products/brands/new` → AddBrandComponent (submenu_id: 46, submenu_code: add_brand)

### Menu 4: Price Lists (menu_code: price_lists)
- `/products/price-lists/list` → PriceListMgmtComponent (submenu_id: 47, submenu_code: price_list_mgmt)
- `/products/price-lists/new` → AddPriceListComponent (submenu_id: 48, submenu_code: add_price_list)

## Module 10: Customer Management (module_code: customers)

### Menu 1: Customers (menu_code: customers)
- `/customers/list` → CustomerListComponent (submenu_id: 49, submenu_code: customer_list)
- `/customers/new` → AddCustomerComponent (submenu_id: 50, submenu_code: add_customer)
- `/customers/history` → CustomerHistoryComponent (submenu_id: 51, submenu_code: customer_history)

## Module 11: Supplier Management (module_code: suppliers)

### Menu 1: Suppliers (menu_code: suppliers)
- `/suppliers/list` → SupplierListComponent (submenu_id: 52, submenu_code: supplier_list) ✅ EXISTS
- `/suppliers/new` → AddSupplierComponent (submenu_id: 53, submenu_code: add_supplier) ✅ EXISTS

## Module 12: Purchase Orders (module_code: purchase_orders)

### Menu 1: Purchase Orders (menu_code: purchase_orders)
- `/purchase-orders/list` → PoListComponent (submenu_id: 54, submenu_code: po_list)
- `/purchase-orders/new` → CreatePoComponent (submenu_id: 55, submenu_code: create_po)
- `/purchase-orders/approve` → ApprovePoComponent (submenu_id: 56, submenu_code: approve_po)

## Module 13: Sales Orders (module_code: sales_orders)

### Menu 1: Sales Orders (menu_code: sales_orders)
- `/sales-orders/list` → SoListComponent (submenu_id: 57, submenu_code: so_list)
- `/sales-orders/new` → CreateSoComponent (submenu_id: 58, submenu_code: create_so)

## Module 14: Reports & Analytics (module_code: reports)

### Menu 1: Sales Reports (menu_code: sales_reports)
- `/reports/sales/daily` → DailySalesReportComponent (submenu_id: 59, submenu_code: daily_sales_report)
- `/reports/sales/monthly` → MonthlySalesReportComponent (submenu_id: 60, submenu_code: monthly_sales_report)
- `/reports/sales/products` → ProductPerformanceComponent (submenu_id: 61, submenu_code: product_performance)

### Menu 2: Purchase Reports (menu_code: purchase_reports)
- `/reports/purchases/summary` → PurchaseSummaryComponent (submenu_id: 62, submenu_code: purchase_summary)
- `/reports/purchases/suppliers` → SupplierAnalysisComponent (submenu_id: 63, submenu_code: supplier_analysis)

### Menu 3: Inventory Reports (menu_code: inventory_reports)
- `/reports/inventory/valuation` → StockValuationComponent (submenu_id: 64, submenu_code: stock_valuation)
- `/reports/inventory/turnover` → InventoryTurnoverComponent (submenu_id: 65, submenu_code: inventory_turnover)

### Menu 4: Financial Reports (menu_code: financial_reports)
- `/reports/financial/pl` → ProfitLossComponent (submenu_id: 66, submenu_code: profit_loss)
- `/reports/financial/discounts` → DiscountAnalysisComponent (submenu_id: 67, submenu_code: discount_analysis)

## Module 15: System Administration (module_code: admin)

### Menu 1: UI Modules (menu_code: ui_modules)
- `/admin/ui-modules/list` → ModuleListComponent (submenu_id: 68, submenu_code: module_list)
- `/admin/ui-modules/menus` → MenuManagementComponent (submenu_id: 69, submenu_code: menu_management)
- `/admin/ui-modules/permissions` → PermissionManagementComponent (submenu_id: 70, submenu_code: permission_management)

### Menu 2: System Settings (menu_code: system_settings)
- `/admin/settings/general` → GeneralSettingsComponent (submenu_id: 71, submenu_code: general_settings) ✅ EXISTS
- `/admin/settings/tax` → TaxConfigComponent (submenu_id: 72, submenu_code: tax_config)

### Menu 3: Audit Logs (menu_code: audit_logs)
- `/admin/audit-logs/view` → ViewAuditLogsComponent (submenu_id: 73, submenu_code: view_audit_logs)

## Route Path Summary

Total Routes: 73 submenu routes across 15 modules
