/**
 * Generates NEMBUS Postman Collection with all routes from internal/routing.
 * Run: node generate-collection.js
 * Uses string URL so Postman displays the URL in the request bar.
 */
const fs = require('fs');
const path = require('path');

const H = (key, value) => ({ key, value: value || `{{${key}}}`, type: 'text' });
const tenantHeader = [H('x-tenant-id', '{{tenant_id}}')];
const authNoAuth = { type: 'noauth' };

function req(name, method, urlPath, opts = {}) {
  const url = urlPath.startsWith('http') ? urlPath : `{{base_url}}${urlPath}`;
  const request = {
    method,
    header: opts.header || tenantHeader,
    url: url  // string URL so Postman shows it in the bar
  };
  if (opts.auth) request.auth = opts.auth;
  if (opts.body) request.body = { mode: 'raw', raw: typeof opts.body === 'string' ? opts.body : JSON.stringify(opts.body, null, 0) };
  if (opts.description) request.description = opts.description;
  return { name, request };
}

function folder(name, items) {
  return { name, item: items };
}

const collection = {
  info: {
    _postman_id: 'a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d',
    name: 'NEMBUS API',
    description: 'NEMBUS Backend API. All routes from routing + JWT/Tenant middleware. Use x-tenant-id and Bearer token. Select NEMBUS Local environment.',
    schema: 'https://schema.getpostman.com/json/collection/v2.1.0/collection.json'
  },
  auth: { type: 'bearer', bearer: [{ key: 'token', value: '{{token}}', type: 'string' }] },
  variable: [
    { key: 'base_url', value: 'http://localhost:8080' },
    { key: 'tenant_id', value: 'default' },
    { key: 'token', value: '' },
    { key: 'store_id', value: '1' },
    { key: 'cart_id', value: '' },
    { key: 'order_id', value: '' },
    { key: 'transaction_id', value: '1' },
    { key: 'cashier_id', value: '1' },
    { key: 'session_id', value: '1' },
    { key: 'id', value: '1' },
    { key: 'user_id', value: '1' },
    { key: 'role_id', value: '1' },
    { key: 'product_id', value: '1' }
  ],
  item: [
    // --- Auth (no JWT) ---
    folder('Auth', [
      req('Login', 'POST', '/api/auth/login', {
        auth: authNoAuth,
        header: [H('x-tenant-id', '{{tenant_id}}'), H('Content-Type', 'application/json')],
        body: '{"user_login":"admin","password":"your_password"}'
      })
    ]),
    // --- Dev (dev only) ---
    folder('Dev', [
      req('Get Dev Token', 'GET', '{{base_url}}/dev/token', { auth: authNoAuth })
    ]),
    // --- Public Tenants (no JWT in main, but same base) ---
    folder('Tenants (public)', [
      req('Get Tenant by Slug', 'GET', '/api/tenants/{{tenant_id}}'),
      req('List Active Tenants', 'GET', '/api/tenants/active')
    ]),
    // --- Carts ---
    folder('Carts', [
      req('Create Cart', 'POST', '/api/carts', { body: { organization_id: 1, store_id: 1, cart_type: 'standard', channel: 'pos' } }),
      req('Create New Cart', 'POST', '/api/carts/new', { body: { organization_id: 1, store_id: 1 } }),
      req('List Active Carts', 'GET', '/api/carts'),
      req('Get Cart', 'GET', '/api/carts/{{cart_id}}'),
      req('Get Cart by Number', 'GET', '/api/carts/by-number/CART-001'),
      req('Get by Customer', 'GET', '/api/carts/by-customer'),
      req('Get by Guest', 'GET', '/api/carts/by-guest'),
      req('Update Cart', 'PUT', '/api/carts/{{cart_id}}', { body: {} }),
      req('Update Cart Status', 'PUT', '/api/carts/{{cart_id}}/status', { body: {} }),
      req('Update Cart Customer', 'PUT', '/api/carts/{{cart_id}}/customer', { body: {} }),
      req('Delete Cart', 'DELETE', '/api/carts/{{cart_id}}'),
      req('Expire Abandoned', 'POST', '/api/carts/expire', { body: {} }),
      req('List Cart Items', 'GET', '/api/carts/{{cart_id}}/items'),
      req('Add to Cart', 'POST', '/api/carts/{{cart_id}}/items', { body: { organization_id: 1, product_id: 1, quantity: 2, uom_id: 1, price_list_id: 1, unit_price: '10', line_total: '20' } }),
      req('Create Cart Item Raw', 'POST', '/api/carts/{{cart_id}}/items/raw', { body: {} }),
      req('Get Cart Item by Product', 'GET', '/api/carts/{{cart_id}}/items/by-product?product_id=1'),
      req('Clear Cart Items', 'DELETE', '/api/carts/{{cart_id}}/items'),
      req('Get Cart Item Count', 'GET', '/api/carts/{{cart_id}}/items/count'),
      req('Get Cart Totals', 'GET', '/api/carts/{{cart_id}}/totals'),
      req('Apply Coupon', 'POST', '/api/carts/{{cart_id}}/coupon', { body: { code: 'PROMO1' } }),
      req('Recalculate Totals', 'POST', '/api/carts/{{cart_id}}/recalculate'),
      req('Checkout (Convert to Order)', 'POST', '/api/carts/{{cart_id}}/checkout'),
      req('Merge Guest to Customer', 'POST', '/api/carts/{{cart_id}}/merge', { body: {} }),
      req('Create Cart Activity', 'POST', '/api/carts/{{cart_id}}/activities', { body: {} }),
      req('List Cart Activities', 'GET', '/api/carts/{{cart_id}}/activities')
    ]),
    folder('Cart Items', [
      req('Get Cart Item', 'GET', '/api/cart-items/1'),
      req('Update Cart Item', 'PUT', '/api/cart-items/1', { body: {} }),
      req('Update Item Quantity', 'PATCH', '/api/cart-items/1/quantity', { body: { quantity: 2 } }),
      req('Delete Cart Item', 'DELETE', '/api/cart-items/1')
    ]),
    // --- POS ---
    folder('POS', [
      req('Get POS Categories', 'GET', '/api/pos/categories'),
      req('Add POS Product', 'POST', '/api/pos/products', { body: { organization_id: 1, sku: 'SKU1', name: 'Product', category_id: 1, base_uom_id: 1 } }),
      req('Process Payment', 'POST', '/api/pos/payments', { body: { transaction_id: 1, payment_method: 'card', amount: '50.00', reference_number: 'ref-1' } }),
      req('List Store Products', 'GET', '/api/pos/stores/{{store_id}}/products'),
      req('Search Product', 'GET', '/api/pos/stores/{{store_id}}/products/search?q=test'),
      req('Products by Category', 'GET', '/api/pos/stores/{{store_id}}/products/category/1'),
      req('List Store Transactions', 'GET', '/api/pos/stores/{{store_id}}/transactions'),
      req('Get Transaction', 'GET', '/api/pos/transactions/{{transaction_id}}'),
      req('Get Transaction Full', 'GET', '/api/pos/transactions/{{transaction_id}}/full'),
      req('Get Transaction Payments', 'GET', '/api/pos/transactions/{{transaction_id}}/payments'),
      req('Get Payment Summary', 'GET', '/api/pos/transactions/{{transaction_id}}/payment-summary'),
      req('Void Transaction', 'POST', '/api/pos/transactions/{{transaction_id}}/void', { body: { voided_by: 1, reason: 'test' } })
    ]),
    folder('POS Returns', [
      req('Process Return', 'POST', '/api/pos/returns', { body: { store_id: 1, session_id: 1, original_transaction_id: 1, return_reason: 'Defective', subtotal: '50', tax_amount: '5', total_refund_amount: '55', refund_method: 'cash', lines: [] } })
    ]),
    folder('POS Terminals', [
      req('List Terminals', 'GET', '/api/pos/terminals'),
      req('Create Terminal', 'POST', '/api/pos/terminals', { body: {} }),
      req('Get Terminal', 'GET', '/api/pos/terminals/1'),
      req('Update Terminal', 'PUT', '/api/pos/terminals/1', { body: {} }),
      req('Delete Terminal', 'DELETE', '/api/pos/terminals/1'),
      req('Toggle Terminal Active', 'PATCH', '/api/pos/terminals/1/active', { body: {} }),
      req('List Store Terminals', 'GET', '/api/pos/stores/{{store_id}}/terminals'),
      req('List Active Store Terminals', 'GET', '/api/pos/stores/{{store_id}}/terminals/active'),
      req('Get Terminal by Code', 'GET', '/api/pos/stores/{{store_id}}/terminals/code/T1')
    ]),
    // --- Cashier Sessions ---
    folder('Cashier Sessions', [
      req('Open Session', 'POST', '/api/cashier-sessions', { body: { cashier_id: 1, pos_terminal_id: 1, session_number: 'SES-001', opening_balance: '100.00' } }),
      req('Get Active Session', 'GET', '/api/cashier-sessions/active/{{cashier_id}}'),
      req('Get Session', 'GET', '/api/cashier-sessions/{{session_id}}'),
      req('Close Session', 'PUT', '/api/cashier-sessions/{{session_id}}/close', { body: { closing_balance: '200', closing_note: 'OK', closed_by: 1 } }),
      req('Get Session Summary', 'GET', '/api/cashier-sessions/{{session_id}}/summary')
    ]),
    // --- Cashiers ---
    folder('Cashiers', [
      req('Create Cashier', 'POST', '/api/cashiers', { body: {} }),
      req('Create with Defaults', 'POST', '/api/cashiers/with-defaults', { body: {} }),
      req('List All', 'GET', '/api/cashiers/all'),
      req('List Active', 'GET', '/api/cashiers/active'),
      req('List Cashiers', 'GET', '/api/cashiers'),
      req('Count', 'GET', '/api/cashiers/count'),
      req('Count Active', 'GET', '/api/cashiers/count/active'),
      req('Count by Store', 'GET', '/api/cashiers/count/store/{{store_id}}'),
      req('List by Store', 'GET', '/api/cashiers/store/{{store_id}}'),
      req('List Active by Store', 'GET', '/api/cashiers/store/{{store_id}}/active'),
      req('List Active with Sessions', 'GET', '/api/cashiers/store/{{store_id}}/active-with-sessions'),
      req('Get Cashier', 'GET', '/api/cashiers/{{cashier_id}}'),
      req('Get by Code', 'GET', '/api/cashiers/code/CODE1'),
      req('Get by User', 'GET', '/api/cashiers/user/{{user_id}}'),
      req('Exists', 'GET', '/api/cashiers/1/exists'),
      req('Code Exists', 'GET', '/api/cashiers/code/CODE1/exists'),
      req('Get Limits', 'GET', '/api/cashiers/1/limits'),
      req('Update Cashier', 'PUT', '/api/cashiers/1', { body: {} }),
      req('Update Limits', 'PATCH', '/api/cashiers/1/limits', { body: {} }),
      req('Update Drawer Limit', 'PATCH', '/api/cashiers/1/drawer-limit', { body: {} }),
      req('Update Discount Limit', 'PATCH', '/api/cashiers/1/discount-limit', { body: {} }),
      req('Update Metadata', 'PATCH', '/api/cashiers/1/metadata', { body: {} }),
      req('Activate', 'PATCH', '/api/cashiers/1/activate', { body: {} }),
      req('Deactivate', 'PATCH', '/api/cashiers/1/deactivate', { body: {} }),
      req('Delete', 'DELETE', '/api/cashiers/1'),
      req('Soft Delete', 'DELETE', '/api/cashiers/1/soft')
    ]),
    // --- Orders ---
    folder('Orders', [
      req('List Orders', 'GET', '/api/orders'),
      req('Create Order', 'POST', '/api/orders', { body: {} }),
      req('Get by Number', 'GET', '/api/orders/by-number/ORD-001'),
      req('Get Order', 'GET', '/api/orders/{{order_id}}'),
      req('Update Order', 'PUT', '/api/orders/{{order_id}}', { body: {} }),
      req('Update Status', 'PUT', '/api/orders/{{order_id}}/status', { body: {} }),
      req('Update Payment Status', 'PUT', '/api/orders/{{order_id}}/payment-status', { body: {} }),
      req('Update Fulfillment Status', 'PUT', '/api/orders/{{order_id}}/fulfillment-status', { body: {} }),
      req('Update Totals', 'PUT', '/api/orders/{{order_id}}/totals', { body: {} }),
      req('Update Delivery', 'PUT', '/api/orders/{{order_id}}/delivery', { body: {} }),
      req('Assign Order', 'PUT', '/api/orders/{{order_id}}/assign', { body: {} }),
      req('Cancel Order', 'POST', '/api/orders/{{order_id}}/cancel'),
      req('Delete Order', 'DELETE', '/api/orders/{{order_id}}'),
      req('List Order Lines', 'GET', '/api/orders/{{order_id}}/lines'),
      req('Create Order Line', 'POST', '/api/orders/{{order_id}}/lines', { body: {} }),
      req('Get Line Totals', 'GET', '/api/orders/{{order_id}}/lines/totals'),
      req('Get Line Margin', 'GET', '/api/orders/{{order_id}}/lines/margin'),
      req('Create Status History', 'POST', '/api/orders/{{order_id}}/status-history', { body: {} }),
      req('List Status History', 'GET', '/api/orders/{{order_id}}/status-history')
    ]),
    folder('Order Lines', [
      req('Get Order Line', 'GET', '/api/order-lines/1'),
      req('Update Order Line', 'PUT', '/api/order-lines/1', { body: {} }),
      req('Update Fulfillment', 'PATCH', '/api/order-lines/1/fulfillment', { body: {} }),
      req('Update Status', 'PATCH', '/api/order-lines/1/status', { body: {} }),
      req('Delete Order Line', 'DELETE', '/api/order-lines/1')
    ]),
    folder('Order Fulfillments', [
      req('Get Fulfillment', 'GET', '/api/order-fulfillments/1'),
      req('Get by Number', 'GET', '/api/order-fulfillments/by-number/FUL-001'),
      req('Update Fulfillment', 'PUT', '/api/order-fulfillments/1', { body: {} }),
      req('Update Shipment', 'PUT', '/api/order-fulfillments/1/shipment', { body: {} }),
      req('Update Pick Pack', 'PUT', '/api/order-fulfillments/1/pick-pack', { body: {} }),
      req('Delete Fulfillment', 'DELETE', '/api/order-fulfillments/1'),
      req('Create Fulfillment Item', 'POST', '/api/order-fulfillments/1/items', { body: {} }),
      req('List Fulfillment Items', 'GET', '/api/order-fulfillments/1/items')
    ]),
    folder('Order Fulfillment Items', [
      req('Delete Fulfillment Item', 'DELETE', '/api/order-fulfillment-items/1')
    ]),
    // --- Stores ---
    folder('Stores', [
      req('Create Store', 'POST', '/api/stores', { body: { organization_id: 1, name: 'Store', code: 'STORE1', store_type: 'retail', is_warehouse: false, is_pos_enabled: true, timezone: 'Asia/Riyadh' } }),
      req('List Stores', 'GET', '/api/stores'),
      req('List POS Enabled', 'GET', '/api/stores/pos-enabled'),
      req('List Warehouses', 'GET', '/api/stores/warehouse'),
      req('List by Parent', 'GET', '/api/stores/by-parent/1'),
      req('Get Store', 'GET', '/api/stores/{{store_id}}'),
      req('Get Hierarchy', 'GET', '/api/stores/{{store_id}}/hierarchy'),
      req('Update Store', 'PATCH', '/api/stores/{{store_id}}', { body: {} }),
      req('Delete Store', 'DELETE', '/api/stores/{{store_id}}')
    ]),
    // --- Organizations ---
    folder('Organizations', [
      req('Create Organization', 'POST', '/api/organizations', { body: { name: 'Org', code: 'ORG1' } }),
      req('List Organizations', 'GET', '/api/organizations'),
      req('Get by Code', 'GET', '/api/organizations/code/ORG1'),
      req('Get Organization', 'GET', '/api/organizations/1'),
      req('Update Organization', 'PUT', '/api/organizations/1', { body: {} }),
      req('Delete Organization', 'DELETE', '/api/organizations/1')
    ]),
    // --- Tenants (protected) ---
    folder('Tenants', [
      req('Create Tenant', 'POST', '/api/tenants', { body: {} }),
      req('List All Tenants', 'GET', '/api/tenants/all'),
      req('Get by Slug Any', 'GET', '/api/tenants/default/any'),
      req('Update Tenant', 'PUT', '/api/tenants/1', { body: {} }),
      req('Deactivate Tenant', 'PUT', '/api/tenants/deactivate/default')
    ]),
    // --- Users ---
    folder('Users', [
      req('Create User', 'POST', '/api/users', { body: {} }),
      req('List Users', 'GET', '/api/users'),
      req('Get User', 'GET', '/api/users/1'),
      req('Update User', 'PATCH', '/api/users/1', { body: {} }),
      req('Update Password', 'PUT', '/api/users/1/password', { body: { password: 'newpass' } }),
      req('Assign Role', 'POST', '/api/users/addUserRoles/1', { body: { role_id: 1 } }),
      req('Get by Role', 'GET', '/api/users/role/1'),
      req('Revoke Role', 'DELETE', '/api/users/revokeRole/1/1'),
      req('Revoke All Roles', 'DELETE', '/api/users/revokeAllRoles/1'),
      req('Grant Store', 'POST', '/api/users/grantStore/1', { body: {} }),
      req('Set Primary Store', 'PUT', '/api/users/1/stores/primary', { body: { store_id: 1 } }),
      req('Revoke Store', 'DELETE', '/api/users/revokeStore/1/{{store_id}}'),
      req('Revoke All Stores', 'DELETE', '/api/users/revokeAllStores/1'),
      req('Get User Details', 'GET', '/api/users/details/1'),
      req('List Users Details', 'GET', '/api/users/details'),
      req('Search Users', 'GET', '/api/users/search?q=admin'),
      req('Get Store Users', 'GET', '/api/users/store/{{store_id}}'),
      req('Get User Stores', 'GET', '/api/users/1/stores'),
      req('Get User Primary Store', 'GET', '/api/users/1/primaryStore')
    ]),
    // --- Roles ---
    folder('Roles', [
      req('Create Role', 'POST', '/api/roles', { body: {} }),
      req('List Roles', 'GET', '/api/roles'),
      req('Get Role', 'GET', '/api/roles/1'),
      req('Get by Code', 'GET', '/api/roles/code/ADMIN'),
      req('Update Role', 'PUT', '/api/roles/1', { body: {} }),
      req('Delete Role', 'DELETE', '/api/roles/1'),
      req('List Active', 'GET', '/api/roles/active'),
      req('List Non-System', 'GET', '/api/roles/non-system'),
      req('Toggle Active', 'PATCH', '/api/roles/1/active', { body: {} }),
      req('Assign Permission', 'POST', '/api/roles/1/permissions', { body: {} }),
      req('Get Role Permissions', 'GET', '/api/roles/1/permissions'),
      req('Remove Permission', 'DELETE', '/api/roles/1/permissions/1'),
      req('Check Permission', 'GET', '/api/roles/1/permissions/1/check')
    ]),
    // --- Permissions ---
    folder('Permissions', [
      req('Create Permission', 'POST', '/api/permissions', { body: {} }),
      req('List Permissions', 'GET', '/api/permissions'),
      req('Get Permission', 'GET', '/api/permissions/1'),
      req('Get by Code', 'GET', '/api/permissions/code/READ'),
      req('Update Permission', 'PUT', '/api/permissions/1', { body: {} }),
      req('Delete Permission', 'DELETE', '/api/permissions/1'),
      req('Get by Module', 'GET', '/api/permissions/module/1'),
      req('Assign to Module', 'POST', '/api/permissions/module/1', { body: {} }),
      req('Revoke from Module', 'DELETE', '/api/permissions/module/1/permission/1'),
      req('Get by Menu', 'GET', '/api/permissions/menu/1'),
      req('Assign to Menu', 'POST', '/api/permissions/menu/1', { body: {} }),
      req('Revoke from Menu', 'DELETE', '/api/permissions/menu/1/permission/1'),
      req('Get by Submenu', 'GET', '/api/permissions/submenu/1'),
      req('Assign to Submenu', 'POST', '/api/permissions/submenu/1', { body: {} }),
      req('Revoke from Submenu', 'DELETE', '/api/permissions/submenu/1/permission/1'),
      req('Get Role Permissions', 'GET', '/api/permissions/role/1'),
      req('Revoke from Role', 'DELETE', '/api/permissions/role/1/permission/1'),
      req('Update Role Scope', 'PUT', '/api/permissions/role/1/permission/1/scope', { body: {} }),
      req('Check User Submenu', 'GET', '/api/permissions/user/1/submenu/MENU1'),
      req('Get User Permissions', 'GET', '/api/permissions/user/1'),
      req('Get User with Scope', 'GET', '/api/permissions/user/1/with-scope'),
      req('Check User Permission', 'GET', '/api/permissions/user/1/check/READ'),
      req('Get User Modules', 'GET', '/api/permissions/user/1/modules'),
      req('Get User Menus', 'GET', '/api/permissions/user/1/menus'),
      req('Get User Submenus', 'GET', '/api/permissions/user/1/submenus')
    ]),
    // --- Modules ---
    folder('Modules', [
      req('Create Module', 'POST', '/api/modules', { body: {} }),
      req('List Modules', 'GET', '/api/modules'),
      req('Get Module', 'GET', '/api/modules/1'),
      req('Get by Code', 'GET', '/api/modules/code/POS'),
      req('Update Module', 'PUT', '/api/modules/1', { body: {} }),
      req('Delete Module', 'DELETE', '/api/modules/1'),
      req('Get Navigation', 'GET', '/api/modules/navigation')
    ]),
    // --- Menus ---
    folder('Menus', [
      req('Create Menu', 'POST', '/api/menus', { body: {} }),
      req('List Menus', 'GET', '/api/menus'),
      req('Get Menu', 'GET', '/api/menus/1'),
      req('List by Module', 'GET', '/api/menus/module/1'),
      req('List by Parent', 'GET', '/api/menus/parent/1'),
      req('Toggle Active', 'PATCH', '/api/menus/1/toggle-active', { body: {} }),
      req('Update Menu', 'PUT', '/api/menus/1', { body: {} }),
      req('Delete Menu', 'DELETE', '/api/menus/1')
    ]),
    // --- Submenus ---
    folder('Submenus', [
      req('Create Submenu', 'POST', '/api/submenus', { body: {} }),
      req('List Submenus', 'GET', '/api/submenus'),
      req('Get Submenu', 'GET', '/api/submenus/1'),
      req('List by Menu', 'GET', '/api/submenus/by-menu/1'),
      req('List Active by Menu', 'GET', '/api/submenus/active/1'),
      req('List by Parent', 'GET', '/api/submenus/parent/1'),
      req('Get by Code', 'GET', '/api/submenus/by-code?menu_id=1&code=SUB1'),
      req('Toggle Active', 'PATCH', '/api/submenus/1/toggle', { body: {} }),
      req('Update Submenu', 'PUT', '/api/submenus/1', { body: {} }),
      req('Delete Submenu', 'DELETE', '/api/submenus/1')
    ]),
    // --- Navigation ---
    folder('Navigation', [
      req('Get User Navigation', 'GET', '/api/navigation/user/{{user_id}}'),
      req('Get Roles with User Counts', 'GET', '/api/navigation/rolesWithUserCounts/ADMIN')
    ]),
    // --- Images ---
    folder('Images', [
      req('Upload Image', 'POST', '/api/images/uploadImage/products', { body: {}, description: 'Use form-data file upload; module: products, categories, brands, etc.' })
    ]),
    // --- Brands ---
    folder('Brands', [
      req('Create Brand', 'POST', '/api/brands', { body: {} }),
      req('Create with Defaults', 'POST', '/api/brands/with-defaults', { body: {} }),
      req('List All', 'GET', '/api/brands/all'),
      req('List Active', 'GET', '/api/brands/active'),
      req('List Brands', 'GET', '/api/brands'),
      req('List Active Paginated', 'GET', '/api/brands/active/paginated'),
      req('Count', 'GET', '/api/brands/count'),
      req('Count Active', 'GET', '/api/brands/count/active'),
      req('Search', 'GET', '/api/brands/search?q=test'),
      req('Search Active', 'GET', '/api/brands/search/active?q=test'),
      req('Get Brand', 'GET', '/api/brands/1'),
      req('Get by Code', 'GET', '/api/brands/code/BRAND1'),
      req('Exists', 'GET', '/api/brands/1/exists'),
      req('Code Exists', 'GET', '/api/brands/code/BRAND1/exists'),
      req('Update Brand', 'PUT', '/api/brands/1', { body: {} }),
      req('Update Name', 'PATCH', '/api/brands/1/name', { body: {} }),
      req('Update Code', 'PATCH', '/api/brands/1/code', { body: {} }),
      req('Update Metadata', 'PATCH', '/api/brands/1/metadata', { body: {} }),
      req('Activate', 'PATCH', '/api/brands/1/activate', { body: {} }),
      req('Deactivate', 'PATCH', '/api/brands/1/deactivate', { body: {} }),
      req('Toggle Status', 'PATCH', '/api/brands/1/toggle-status', { body: {} }),
      req('Delete', 'DELETE', '/api/brands/1'),
      req('Delete by Code', 'DELETE', '/api/brands/code/BRAND1'),
      req('Soft Delete', 'DELETE', '/api/brands/1/soft'),
      req('Get with Product Count', 'GET', '/api/brands/1/products/count'),
      req('List with Counts', 'GET', '/api/brands/products/counts'),
      req('List Active with Counts', 'GET', '/api/brands/active/products/counts'),
      req('Get Top Brands', 'GET', '/api/brands/top'),
      req('Get No Products', 'GET', '/api/brands/no-products'),
      req('Get Inactive with Active Products', 'GET', '/api/brands/inactive/active-products'),
      req('Bulk Activate', 'POST', '/api/brands/bulk/activate', { body: {} }),
      req('Bulk Deactivate', 'POST', '/api/brands/bulk/deactivate', { body: {} }),
      req('Bulk Delete', 'POST', '/api/brands/bulk/delete', { body: {} }),
      req('Recent Created', 'GET', '/api/brands/recent/created'),
      req('Recent Updated', 'GET', '/api/brands/recent/updated'),
      req('By Date', 'GET', '/api/brands/by-date'),
      req('Get Metadata Key', 'GET', '/api/brands/1/metadata/key1'),
      req('List with Stats', 'GET', '/api/brands/stats')
    ]),
    // --- Price Lists ---
    folder('Price Lists', [
      req('Create Price List', 'POST', '/api/price-lists', { body: {} }),
      req('List', 'GET', '/api/price-lists'),
      req('List Active', 'GET', '/api/price-lists/active'),
      req('List Valid', 'GET', '/api/price-lists/valid'),
      req('Get Default', 'GET', '/api/price-lists/default'),
      req('Get by Code', 'GET', '/api/price-lists/code/PL1'),
      req('Get Price List', 'GET', '/api/price-lists/1'),
      req('Update', 'PUT', '/api/price-lists/1', { body: {} }),
      req('Delete', 'DELETE', '/api/price-lists/1'),
      req('Set Default', 'POST', '/api/price-lists/1/set-default'),
      req('Toggle Active', 'PATCH', '/api/price-lists/1/active', { body: {} })
    ]),
    // --- Tax Categories ---
    folder('Tax Categories', [
      req('Create', 'POST', '/api/tax-categories', { body: {} }),
      req('List', 'GET', '/api/tax-categories'),
      req('List Active', 'GET', '/api/tax-categories/active'),
      req('Get by Code', 'GET', '/api/tax-categories/code/VAT15'),
      req('Get', 'GET', '/api/tax-categories/1'),
      req('Update', 'PUT', '/api/tax-categories/1', { body: {} }),
      req('Delete', 'DELETE', '/api/tax-categories/1'),
      req('Toggle Active', 'PATCH', '/api/tax-categories/1/active', { body: {} })
    ]),
    // --- UOM ---
    folder('UOMs', [
      req('Create', 'POST', '/api/uoms', { body: {} }),
      req('List', 'GET', '/api/uoms'),
      req('List Active', 'GET', '/api/uoms/active'),
      req('List by Type', 'GET', '/api/uoms/by-type?uom_type=quantity'),
      req('Get by Code', 'GET', '/api/uoms/code/PCS'),
      req('Get', 'GET', '/api/uoms/1'),
      req('Update', 'PUT', '/api/uoms/1', { body: {} }),
      req('Delete', 'DELETE', '/api/uoms/1'),
      req('Create Product Conversion', 'POST', '/api/products/{{product_id}}/uom-conversions', { body: {} }),
      req('List Product Conversions', 'GET', '/api/products/{{product_id}}/uom-conversions'),
      req('Get Conversion Lookup', 'GET', '/api/products/{{product_id}}/uom-conversions/lookup'),
      req('Update Conversion', 'PUT', '/api/uom-conversions/1', { body: {} }),
      req('Delete Conversion', 'DELETE', '/api/uom-conversions/1')
    ]),
    // --- Product Pricing ---
    folder('Product Prices', [
      req('Create', 'POST', '/api/product-prices', { body: {} }),
      req('Get', 'GET', '/api/product-prices/1'),
      req('Update', 'PUT', '/api/product-prices/1', { body: {} }),
      req('Delete', 'DELETE', '/api/product-prices/1'),
      req('List by Product', 'GET', '/api/product-prices/product/{{product_id}}'),
      req('Get Product with Pricing', 'GET', '/api/product-prices/product/{{product_id}}/with-pricing'),
      req('List by Price List', 'GET', '/api/product-prices/price-list/1'),
      req('Bulk Update', 'POST', '/api/product-prices/price-list/1/bulk-update', { body: {} }),
      req('Expire Prices', 'POST', '/api/product-prices/price-list/1/expire', { body: {} }),
      req('Get Effective', 'GET', '/api/product-prices/effective?store_id=1&price_list_id=1'),
      req('Get for List', 'GET', '/api/product-prices/price-list'),
      req('Get Comparison', 'GET', '/api/product-prices/comparison/{{product_id}}'),
      req('Search', 'GET', '/api/product-prices/search')
    ]),
    // --- Product Variants ---
    folder('Product Variants', [
      req('Create', 'POST', '/api/product-variants', { body: {} }),
      req('List', 'GET', '/api/product-variants'),
      req('Search', 'GET', '/api/product-variants/search'),
      req('List by Product', 'GET', '/api/product-variants/product/{{product_id}}'),
      req('List Active by Product', 'GET', '/api/product-variants/active/{{product_id}}'),
      req('Get by SKU', 'GET', '/api/product-variants/by-sku'),
      req('Get Variant', 'GET', '/api/product-variants/variant/1'),
      req('Update Variant', 'PUT', '/api/product-variants/variant/1', { body: {} }),
      req('Delete Variant', 'DELETE', '/api/product-variants/variant/1'),
      req('Toggle Active', 'PATCH', '/api/product-variants/variant/1/toggle-active', { body: {} })
    ]),
    // --- Product Barcodes ---
    folder('Product Barcodes', [
      req('Create', 'POST', '/api/product-barcodes', { body: {} }),
      req('List', 'GET', '/api/product-barcodes'),
      req('Lookup', 'GET', '/api/product-barcodes/lookup/BARCODE123'),
      req('Get', 'GET', '/api/product-barcodes/1'),
      req('Update', 'PUT', '/api/product-barcodes/1', { body: {} }),
      req('Delete', 'DELETE', '/api/product-barcodes/1'),
      req('List by Product', 'GET', '/api/products/{{product_id}}/barcodes'),
      req('Get Primary', 'GET', '/api/products/{{product_id}}/barcodes/primary'),
      req('Set Primary', 'PUT', '/api/products/{{product_id}}/barcodes/primary', { body: {} }),
      req('List by Variant', 'GET', '/api/product-variants/1/barcodes')
    ]),
    // --- Inventory Stock ---
    folder('Inventory Stock', [
      req('Create', 'POST', '/api/inventory-stock', { body: {} }),
      req('List', 'GET', '/api/inventory-stock'),
      req('Get', 'GET', '/api/inventory-stock/1'),
      req('Update', 'PUT', '/api/inventory-stock/1', { body: {} }),
      req('Delete', 'DELETE', '/api/inventory-stock/1'),
      req('Upsert', 'POST', '/api/inventory-stock/upsert', { body: {} }),
      req('Adjust', 'POST', '/api/inventory-stock/1/adjust', { body: {} }),
      req('Adjust by Product Store', 'POST', '/api/inventory-stock/adjust', { body: {} }),
      req('Get by Product Store', 'GET', '/api/inventory-stock/product-store'),
      req('List by Store', 'GET', '/api/inventory-stock/store/{{store_id}}'),
      req('List by Store and Location', 'GET', '/api/inventory-stock/store/{{store_id}}/location'),
      req('Get Store Summary', 'GET', '/api/inventory-stock/store/{{store_id}}/summary'),
      req('List by Product', 'GET', '/api/inventory-stock/product/{{product_id}}'),
      req('List by Storage Location', 'GET', '/api/inventory-stock/storage-location/1')
    ]),
    // --- Storage Locations ---
    folder('Storage Locations', [
      req('List', 'GET', '/api/storage-locations'),
      req('Create', 'POST', '/api/storage-locations', { body: {} }),
      req('List by Parent', 'GET', '/api/storage-locations/by-parent'),
      req('Get', 'GET', '/api/storage-locations/1'),
      req('Update', 'PUT', '/api/storage-locations/1', { body: {} }),
      req('Delete', 'DELETE', '/api/storage-locations/1'),
      req('Toggle Active', 'PATCH', '/api/storage-locations/1/active', { body: {} }),
      req('List by Store', 'GET', '/api/stores/{{store_id}}/storage-locations'),
      req('List Active by Store', 'GET', '/api/stores/{{store_id}}/storage-locations/active'),
      req('Get by Code', 'GET', '/api/stores/{{store_id}}/storage-locations/code/WH1'),
      req('List by Type', 'GET', '/api/stores/{{store_id}}/storage-locations/type/warehouse')
    ]),
    // --- Restaurant ---
    folder('Restaurant', [
      req('Create Table', 'POST', '/api/restaurant/tables', { body: {} }),
      req('Get Table', 'GET', '/api/restaurant/tables/1'),
      req('Update Table', 'PUT', '/api/restaurant/tables/1', { body: {} }),
      req('Delete Table', 'DELETE', '/api/restaurant/tables/1'),
      req('Create Menu Category', 'POST', '/api/restaurant/menu-categories', { body: {} }),
      req('Get Menu Category', 'GET', '/api/restaurant/menu-categories/1'),
      req('Update Menu Category', 'PUT', '/api/restaurant/menu-categories/1', { body: {} }),
      req('Delete Menu Category', 'DELETE', '/api/restaurant/menu-categories/1'),
      req('List Menu Items', 'GET', '/api/restaurant/menu-categories/1/items'),
      req('Create Menu Item', 'POST', '/api/restaurant/menu-items', { body: {} }),
      req('Get Menu Item', 'GET', '/api/restaurant/menu-items/1'),
      req('Update Menu Item', 'PUT', '/api/restaurant/menu-items/1', { body: {} }),
      req('Delete Menu Item', 'DELETE', '/api/restaurant/menu-items/1'),
      req('List Modifiers', 'GET', '/api/restaurant/menu-items/1/modifiers'),
      req('Create Modifier', 'POST', '/api/restaurant/modifiers', { body: {} }),
      req('Get Modifier', 'GET', '/api/restaurant/modifiers/1'),
      req('Update Modifier', 'PUT', '/api/restaurant/modifiers/1', { body: {} }),
      req('Delete Modifier', 'DELETE', '/api/restaurant/modifiers/1'),
      req('Create Order', 'POST', '/api/restaurant/orders', { body: {} }),
      req('Create Online Order', 'POST', '/api/restaurant/orders/online', { body: {} }),
      req('Get Order', 'GET', '/api/restaurant/orders/1'),
      req('Update Order', 'PUT', '/api/restaurant/orders/1', { body: {} }),
      req('Delete Order', 'DELETE', '/api/restaurant/orders/1')
    ])
  ]
};

// Ensure URL is always a string for display in Postman
collection.item.forEach(folder => {
  if (folder.item)
    folder.item.forEach(it => {
      if (it.request && it.request.url && typeof it.request.url !== 'string')
        it.request.url = it.request.url.raw || it.request.url;
    });
});

const outPath = path.join(__dirname, 'Nembus-API.postman_collection.json');
fs.writeFileSync(outPath, JSON.stringify(collection, null, 2), 'utf8');
console.log('Written:', outPath);
