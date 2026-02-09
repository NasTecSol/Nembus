# Complete Route Paths List

This document provides a comprehensive list of all route paths mapped according to the navigation structure.

## Route Paths by Module

### Module 1: Dashboard (module_code: dashboard)

#### Overview Menu
- `/dashboard` → AdminDashboardComponent (submenu_id: 1)
- `/dashboard-stores` → StoreDashboardComponent (submenu_id: 2)

#### Analytics Menu
- `/dashboard/analytics/sales` → SalesAnalyticsComponent (submenu_id: 3)
- `/dashboard/analytics/inventory` → InventoryAnalyticsComponent (submenu_id: 4)

### Module 2: Tenant Management (module_code: tenants)

- `/admin/tenants/list` → TenantListComponent (submenu_id: 5)
- `/admin/tenants/new` → AddTenantComponent (submenu_id: 6)
- `/admin/tenants/config` → TenantConfigComponent (submenu_id: 7)

### Module 3: Organization Setup (module_code: organizations)

- `/admin/organizations/list` → OrgListComponent (submenu_id: 8)
- `/admin/organizations/new` → AddOrgComponent (submenu_id: 9)

### Module 4: User Management (module_code: users)

#### Users Menu
- `/admin/users/list` → UserListComponent (submenu_id: 10) ✅
- `/admin/users/new` → AddUserComponent (submenu_id: 11) ✅
- `/admin/users/activity` → UserActivityComponent (submenu_id: 12)

#### Roles & Permissions Menu
- `/admin/roles/list` → RoleListComponent (submenu_id: 13)
- `/admin/roles/new` → AddRoleComponent (submenu_id: 14)
- `/admin/roles/permissions` → PermissionMatrixComponent (submenu_id: 15)

### Module 5: Store Management (module_code: stores)

#### Stores Menu
- `/stores/list` → StoreslistComponent (submenu_id: 16) ✅
- `/stores/new` → AddStoreComponent (submenu_id: 17) ✅
- `/stores/config` → StoreConfigComponent (submenu_id: 18)

#### Storage Locations Menu
- `/stores/locations/list` → LocationListComponent (submenu_id: 19)
- `/stores/locations/new` → AddLocationComponent (submenu_id: 20)

### Module 6: Point of Sale (module_code: pos)

#### POS Transactions Menu
- `/pos/transactions/list` → TransactionListComponent (submenu_id: 21)
- `/pos/transactions/new` → ProcessSaleComponent (submenu_id: 22)
- `/pos/transactions/void` → VoidTransactionComponent (submenu_id: 23)

#### POS Terminals Menu
- `/pos/terminals/list` → TerminalListComponent (submenu_id: 24)
- `/pos/terminals/new` → AddTerminalComponent (submenu_id: 25)

#### POS Reports Menu
- `/pos/reports/daily` → DailySalesComponent (submenu_id: 26)
- `/pos/reports/cashier` → CashierPerformanceComponent (submenu_id: 27)

### Module 7: Cashier Operations (module_code: cashiers)

#### Cashiers Menu
- `/cashiers/list` → CashierListComponent (submenu_id: 28)
- `/cashiers/new` → AddCashierComponent (submenu_id: 29)

#### Cashier Sessions Menu
- `/cashiers/sessions/active` → ActiveSessionsComponent (submenu_id: 30)
- `/cashiers/sessions/history` → SessionHistoryComponent (submenu_id: 31)
- `/cashiers/sessions/open` → OpenSessionComponent (submenu_id: 32)
- `/cashiers/sessions/close` → CloseSessionComponent (submenu_id: 33)

### Module 8: Inventory Management (module_code: inventory)

#### Stock Overview Menu
- `/inventory/overview/levels` → StockLevelsComponent (submenu_id: 34)
- `/inventory/overview/low-stock` → LowStockComponent (submenu_id: 35)

#### Stock Movements Menu
- `/inventory/movements/history` → MovementHistoryComponent (submenu_id: 36)
- `/inventory/movements/new` → RecordMovementComponent (submenu_id: 37)

#### Stock Counts Menu
- `/inventory/counts/list` → StockCountListComponent (submenu_id: 38)
- `/inventory/counts/new` → CreateCountComponent (submenu_id: 39)

### Module 9: Product Catalog (module_code: products)

#### Products Menu
- `/products/list` → ProductListComponent (submenu_id: 40) ✅ (itemslist)
- `/products/new` → AddProductComponent (submenu_id: 41) ✅
- `/products/import` → ProductImportComponent (submenu_id: 42)

#### Categories Menu
- `/products/categories/list` → CategoryListComponent (submenu_id: 43)
- `/products/categories/new` → AddCategoryComponent (submenu_id: 44)

#### Brands Menu
- `/products/brands/list` → BrandListComponent (submenu_id: 45)
- `/products/brands/new` → AddBrandComponent (submenu_id: 46)

#### Price Lists Menu
- `/products/price-lists/list` → PriceListMgmtComponent (submenu_id: 47)
- `/products/price-lists/new` → AddPriceListComponent (submenu_id: 48)

### Module 10: Customer Management (module_code: customers)

- `/customers/list` → CustomerListComponent (submenu_id: 49)
- `/customers/new` → AddCustomerComponent (submenu_id: 50)
- `/customers/history` → CustomerHistoryComponent (submenu_id: 51)

### Module 11: Supplier Management (module_code: suppliers)

- `/suppliers/list` → SupplierListComponent (submenu_id: 52) ✅
- `/suppliers/new` → AddSupplierComponent (submenu_id: 53) ✅

### Module 12: Purchase Orders (module_code: purchase_orders)

- `/purchase-orders/list` → PoListComponent (submenu_id: 54)
- `/purchase-orders/new` → CreatePoComponent (submenu_id: 55)
- `/purchase-orders/approve` → ApprovePoComponent (submenu_id: 56)

### Module 13: Sales Orders (module_code: sales_orders)

- `/sales-orders/list` → SoListComponent (submenu_id: 57)
- `/sales-orders/new` → CreateSoComponent (submenu_id: 58)

### Module 14: Reports & Analytics (module_code: reports)

#### Sales Reports Menu
- `/reports/sales/daily` → DailySalesReportComponent (submenu_id: 59)
- `/reports/sales/monthly` → MonthlySalesReportComponent (submenu_id: 60)
- `/reports/sales/products` → ProductPerformanceComponent (submenu_id: 61)

#### Purchase Reports Menu
- `/reports/purchases/summary` → PurchaseSummaryComponent (submenu_id: 62)
- `/reports/purchases/suppliers` → SupplierAnalysisComponent (submenu_id: 63)

#### Inventory Reports Menu
- `/reports/inventory/valuation` → StockValuationComponent (submenu_id: 64)
- `/reports/inventory/turnover` → InventoryTurnoverComponent (submenu_id: 65)

#### Financial Reports Menu
- `/reports/financial/pl` → ProfitLossComponent (submenu_id: 66)
- `/reports/financial/discounts` → DiscountAnalysisComponent (submenu_id: 67)

### Module 15: System Administration (module_code: admin)

#### UI Modules Menu
- `/admin/ui-modules/list` → ModuleListComponent (submenu_id: 68)
- `/admin/ui-modules/menus` → MenuManagementComponent (submenu_id: 69)
- `/admin/ui-modules/permissions` → PermissionManagementComponent (submenu_id: 70)

#### System Settings Menu
- `/admin/settings/general` → GeneralSettingsComponent (submenu_id: 71) ✅
- `/admin/settings/tax` → TaxConfigComponent (submenu_id: 72)

#### Audit Logs Menu
- `/admin/audit-logs/view` → ViewAuditLogsComponent (submenu_id: 73)

## Summary

- **Total Routes**: 73 submenu routes
- **Total Modules**: 15 modules
- **Existing Components**: ~15 components already implemented
- **Components to Create**: ~58 components need to be created

## Notes

- Routes marked with ✅ indicate components that already exist
- All routes follow the pattern: `/module/menu/submenu` or `/admin/module/menu/submenu`
- Components are lazy-loaded where possible for better performance
- Each submenu has a unique submenu_id and submenu_code for identification
