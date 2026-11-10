/**
 * Generates NEMBUS Postman Collection with all routes from internal/routing and packages/core/routing.
 * Run: node generate-collection.js
 * Uses string URL so Postman displays the URL in the request bar.
 */
const fs = require('fs');
const path = require('path');

const H = (key, value) => ({ key, value: value || `{{${key}}}`, type: 'text' });
const tenantHeader = [H('x-tenant-id', '{{tenant_id}}')];
const authNoAuth = { type: 'noauth' };

function req(name, method, urlPath, opts = {}) {
  const url = (urlPath.startsWith('http') || urlPath.startsWith('{{')) ? urlPath : `{{base_url}}${urlPath}`;
  const request = {
    method,
    header: opts.header || tenantHeader,
    url: url  // string URL so Postman shows it in the bar
  };
  if (opts.auth) request.auth = opts.auth;
  if (opts.body) request.body = { mode: 'raw', raw: typeof opts.body === 'string' ? opts.body : JSON.stringify(opts.body, null, 2) };
  if (opts.description) request.description = opts.description;
  const item = { name, request };
  if (opts.event) item.event = opts.event;
  return item;
}

function folder(name, items) {
  return { name, item: items };
}

const collection = {
  info: {
    _postman_id: 'a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d',
    name: 'NEMBUS API',
    description: 'NEMBUS Backend API & Local Biometric Hardware Bridge. All routes from routing + JWT/Tenant middleware. Use x-tenant-id and Bearer token. Select NEMBUS Local environment.',
    schema: 'https://schema.getpostman.com/json/collection/v2.1.0/collection.json'
  },
  auth: { type: 'bearer', bearer: [{ key: 'token', value: '{{token}}', type: 'string' }] },
  variable: [
    { key: 'base_url', value: 'http://localhost:8080' },
    { key: 'biometric_bridge_url', value: 'http://127.0.0.1:5050' },
    { key: 'tenant_id', value: 'default' },
    { key: 'token', value: '' },
    { key: 'm2m_api_key', value: '' },
    { key: 'store_id', value: '1' },
    { key: 'cart_id', value: '' },
    { key: 'order_id', value: '' },
    { key: 'transaction_id', value: '1' },
    { key: 'cashier_id', value: '1' },
    { key: 'session_id', value: '1' },
    { key: 'id', value: '1' },
    { key: 'user_id', value: '1' },
    { key: 'role_id', value: '1' },
    { key: 'product_id', value: '1' },
    { key: 'backup_job_id', value: '' },
    { key: 'authorization_token', value: '' },
    { key: 'fmd_base64', value: '' }
  ],
  item: [
    // --- Auth (no JWT) ---
    folder('Auth', [
      req('Login', 'POST', '/api/auth/login', {
        auth: authNoAuth,
        header: [H('x-tenant-id', '{{tenant_id}}'), H('Content-Type', 'application/json')],
        body: '{"user_login":"admin","password":"your_password"}'
      }),
      req('Login with Biometric (1:1 or 1:N)', 'POST', '/api/auth/login/biometric', {
        auth: authNoAuth,
        header: [H('x-tenant-id', '{{tenant_id}}'), H('Content-Type', 'application/json')],
        body: {
          fmd_base64: '{{fmd_base64}}',
          user_login: 'admin'
        },
        description: 'Authenticates using an ISO/IEC 19794-2:2005 probe FMD captured from the fingerprint reader. Omit `user_login` for 1:N touch-only identification across active enrolled users.',
        event: [
          {
            listen: 'test',
            script: {
              type: 'text/javascript',
              exec: [
                'const res = pm.response.json();',
                'if (res && res.data && res.data.token) {',
                '  pm.collectionVariables.set("token", res.data.token);',
                '  pm.environment.set("token", res.data.token);',
                '  console.log("✅ Biometric login JWT saved");',
                '}'
              ]
            }
          }
        ]
      }),
      // Supervisor override: authenticate a supervisor and obtain a short-lived
      // authorization token scoped to a specific permission code.
      // The returned authorization_token is auto-saved to the environment variable.
      {
        name: 'Authorize Action (Supervisor Override)',
        request: {
          method: 'POST',
          auth: authNoAuth,
          header: [
            H('x-tenant-id', '{{tenant_id}}'),
            H('Content-Type', 'application/json')
          ],
          url: '{{base_url}}/api/auth/authorize-action',
          body: {
            mode: 'raw',
            raw: JSON.stringify({
              supervisor_login: 'supervisor_jane',
              supervisor_password: 'your_password',
              permission_code: 'sales_return.process',
              cashier_id: 1
            }, null, 2)
          },
          description: 'Authenticates a supervisor and returns a short-lived authorization token (5-min TTL) scoped to the given permission_code. The POS uses this token in the X-Authorization-Token header when calling a restricted endpoint (e.g. POST /api/pos/returns).\n\nThe test script below auto-saves the token to the `authorization_token` environment/collection variable so the POS Returns request can use it immediately.'
        },
        event: [
          {
            listen: 'test',
            script: {
              type: 'text/javascript',
              exec: [
                'const res = pm.response.json();',
                'if (res && res.data && res.data.authorization_token) {',
                '  pm.collectionVariables.set("authorization_token", res.data.authorization_token);',
                '  pm.environment.set("authorization_token", res.data.authorization_token);',
                '  console.log("✅ authorization_token saved:", res.data.permission_code, "expires:", res.data.expires_at);',
                '} else {',
                '  console.error("❌ No authorization_token in response", JSON.stringify(res));',
                '}'
              ]
            }
          }
        ]
      },
      req('Authorize Action with Biometric (Supervisor Fingerprint Override)', 'POST', '/api/auth/authorize-action/biometric', {
        header: [
          H('x-tenant-id', '{{tenant_id}}'),
          H('Content-Type', 'application/json')
        ],
        body: {
          fmd_base64: '{{fmd_base64}}',
          supervisor_login: 'supervisor_jane',
          permission_code: 'biometrics.enroll',
          cashier_id: 1
        },
        description: 'Identifies a supervisor by fingerprint (1:1 if `supervisor_login` provided, else 1:N across active enrolled users) and mints a 5-minute `X-Authorization-Token` when the matched user holds `permission_code`. Auto-saves `authorization_token` to environment/collection variables.',
        event: [
          {
            listen: 'test',
            script: {
              type: 'text/javascript',
              exec: [
                'const res = pm.response.json();',
                'if (res && res.data && res.data.authorization_token) {',
                '  pm.collectionVariables.set("authorization_token", res.data.authorization_token);',
                '  pm.environment.set("authorization_token", res.data.authorization_token);',
                '  console.log("✅ Biometric authorization_token saved:", res.data.permission_code, "expires:", res.data.expires_at);',
                '} else {',
                '  console.error("❌ No authorization_token in response", JSON.stringify(res));',
                '}'
              ]
            }
          }
        ]
      })
    ]),
    // --- Biometrics (Local Hardware Bridge - Port 5050) ---
    folder('Biometrics (Local Hardware Bridge)', [
      req('Check Connected Fingerprint Device', 'GET', '{{biometric_bridge_url}}/api/v1/device', {
        auth: authNoAuth,
        header: [],
        description: 'Checks whether a physical DigitalPersona U.are.U fingerprint reader is connected and ready on `http://127.0.0.1:5050/api/v1/device`.'
      }),
      req('Get Connected Device Info', 'GET', '{{biometric_bridge_url}}/api/v1/device/info', {
        auth: authNoAuth,
        header: [],
        description: 'Returns metadata about the connected fingerprint reader (device name, vendor/product name, serial number, resolution DPI).'
      }),
      req('Capture Single Fingerprint (BMP + FMD)', 'POST', '{{biometric_bridge_url}}/api/v1/capture', {
        auth: authNoAuth,
        header: [H('Content-Type', 'application/json')],
        body: { timeout: 15000 },
        description: 'Blocks until a finger is placed on the scanner, captures the image, checks NFIQ quality, extracts an ISO/IEC 19794-2:2005 FMD template, and auto-saves `fmd_base64` to the Postman environment/collection variable.',
        event: [
          {
            listen: 'test',
            script: {
              type: 'text/javascript',
              exec: [
                'const res = pm.response.json();',
                'const fmd = (res && res.fmd_base64) || (res && res.data && res.data.fmd_base64) || (res && res.fmd);',
                'if (fmd) {',
                '  pm.collectionVariables.set("fmd_base64", fmd);',
                '  pm.environment.set("fmd_base64", fmd);',
                '  console.log("✅ Captured fmd_base64 saved to variable");',
                '}'
              ]
            }
          }
        ]
      }),
      req('Cancel Active Capture', 'DELETE', '{{biometric_bridge_url}}/api/v1/capture', {
        auth: authNoAuth,
        header: [],
        description: 'Cancels an in-progress blocking fingerprint capture session.'
      }),
      req('Multi-Sample Enrollment Capture (4 Scans -> Consolidated FMD)', 'POST', '{{biometric_bridge_url}}/api/v1/enrollment', {
        auth: authNoAuth,
        header: [H('Content-Type', 'application/json')],
        body: {
          count: 4,
          finger_index: 2,
          session_id: 'enroll-session-01'
        },
        description: 'Captures `count` (default 4) sequential scans of the same finger, automatically rejecting poor NFIQ scans (>= 4), consolidates them into an ISO/IEC 19794-2:2005 enrollment FMD, and auto-saves `consolidated_fmd` into `{{fmd_base64}}` for immediate use in `POST /api/users/{{user_id}}/biometrics/enroll`.',
        event: [
          {
            listen: 'test',
            script: {
              type: 'text/javascript',
              exec: [
                'const res = pm.response.json();',
                'const fmd = (res && res.consolidated_fmd) || (res && res.data && res.data.consolidated_fmd) || (res && res.fmd);',
                'if (fmd) {',
                '  pm.collectionVariables.set("fmd_base64", fmd);',
                '  pm.environment.set("fmd_base64", fmd);',
                '  console.log("✅ Consolidated enrollment fmd_base64 saved to variable");',
                '}'
              ]
            }
          }
        ]
      }),
      req('Cancel Active Enrollment Session', 'DELETE', '{{biometric_bridge_url}}/api/v1/enrollment', {
        auth: authNoAuth,
        header: [],
        description: 'Cancels an active enrollment capture session and clears collected views.'
      }),
      req('Interactive Enrollment: 1. Start 4-Step Session', 'POST', '{{biometric_bridge_url}}/api/v1/enrollment/start', {
        auth: authNoAuth,
        header: [],
        description: 'Resets the embedded bridge enrollment state machine for a 4-scan interactive enrollment.'
      }),
      req('Interactive Enrollment: 2. Capture Step (Repeat 4x)', 'POST', '{{biometric_bridge_url}}/api/v1/enrollment/capture-step', {
        auth: authNoAuth,
        header: [],
        description: 'Captures one scan step for the interactive enrollment session. Call 4 times; on step 4 (`needs_more_data: false`), `consolidated_fmd` is returned and auto-saved to `{{fmd_base64}}`.',
        event: [
          {
            listen: 'test',
            script: {
              type: 'text/javascript',
              exec: [
                'const res = pm.response.json();',
                'if (res && res.consolidated_fmd) {',
                '  pm.collectionVariables.set("fmd_base64", res.consolidated_fmd);',
                '  pm.environment.set("fmd_base64", res.consolidated_fmd);',
                '  console.log("✅ Step 4 complete: consolidated_fmd saved to fmd_base64");',
                '}'
              ]
            }
          }
        ]
      })
    ]),
    // --- Biometrics (Cloud / Backend API) ---
    folder('Biometrics', [
      req('List User Enrolled Biometrics', 'GET', '/api/users/{{user_id}}/biometrics', {
        description: 'Returns active enrolled finger indices (1-10) and metadata for a user without exposing raw FMD bytes.'
      }),
      req('Enroll User Fingerprint (by User Path ID)', 'POST', '/api/users/{{user_id}}/biometrics/enroll', {
        header: [
          H('x-tenant-id', '{{tenant_id}}'),
          H('Content-Type', 'application/json'),
          H('X-Authorization-Token', '{{authorization_token}}')
        ],
        body: {
          user_id: 1,
          finger_index: 2,
          fmd_base64: '{{fmd_base64}}',
          quality_score: 90,
          reader_serial: '5000216000000000'
        },
        description: 'Stores a consolidated ISO/IEC 19794-2:2005 FMD after an organization-wide 1:N duplicate check. Requires `biometrics.enroll` permission or an `X-Authorization-Token` scoped to `biometrics.enroll`.\n\n**NIST `finger_index` values:**\n- `1`–`5`: Right Thumb (1), Right Index (2), Right Middle (3), Right Ring (4), Right Little (5)\n- `6`–`10`: Left Thumb (6), Left Index (7), Left Middle (8), Left Ring (9), Left Little (10)'
      }),
      req('Enroll User Fingerprint (Direct Endpoint)', 'POST', '/api/biometrics/enroll', {
        header: [
          H('x-tenant-id', '{{tenant_id}}'),
          H('Content-Type', 'application/json'),
          H('X-Authorization-Token', '{{authorization_token}}')
        ],
        body: {
          user_id: 1,
          finger_index: 2,
          fmd_base64: '{{fmd_base64}}',
          quality_score: 90,
          reader_serial: '5000216000000000'
        },
        description: 'Same enrollment endpoint without `:id` in the URL path (`user_id` is read from the JSON body). You may also pass `view_fmds_base64: ["..."]` to have the backend consolidate raw scan views.'
      }),
      req('Deactivate User Fingerprint Template', 'DELETE', '/api/users/{{user_id}}/biometrics/2', {
        header: [
          H('x-tenant-id', '{{tenant_id}}'),
          H('X-Authorization-Token', '{{authorization_token}}')
        ],
        description: 'Soft-deactivates (`is_active = false`) the specified finger template (`fingerIndex` 1–10) for sync propagation. Requires `biometrics.enroll` permission.'
      })
    ]),
    // --- Dev (dev only) ---
    folder('Dev', [
      req('Get Dev Token', 'GET', '{{base_url}}/api/dev/token', { auth: authNoAuth })
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
      req('Update Item Quantity', 'PATCH', '/api/cart-items/1/quantity', { body: { delta_quantity: '1.00' } }),
      req('Delete Cart Item', 'DELETE', '/api/cart-items/1')
    ]),
    // --- POS ---
    folder('POS', [
      req('Get POS Categories', 'GET', '/api/pos/categories'),
      req('Add POS Product', 'POST', '/api/pos/products', { body: { organization_id: 1, sku: 'SKU1', name: 'Product', category_id: 1, base_uom_id: 1 } }),
      req('Process Payment', 'POST', '/api/pos/payments', { body: { transaction_id: 1, payment_method: 'card', amount: '50.00', reference_number: 'ref-1' } }),
      req('Get Payment', 'GET', '/api/pos/payments/1'),
      req('Update Payment', 'PUT', '/api/pos/payments/1', { body: {} }),
      req('Delete Payment', 'DELETE', '/api/pos/payments/1'),
      req('List Store Products', 'GET', '/api/pos/stores/{{store_id}}/products'),
      req('Search Product', 'GET', '/api/pos/stores/{{store_id}}/products/search?q=test'),
      req('Products by Category', 'GET', '/api/pos/stores/{{store_id}}/products/category/1'),
      req('List Store Transactions', 'GET', '/api/pos/stores/{{store_id}}/transactions'),
      req('Create Transaction', 'POST', '/api/pos/transactions', { body: {} }),
      req('List Transactions by Cashier Session', 'GET', '/api/pos/transactions/by-cashier-session?session_id={{session_id}}'),
      req('List Transactions by Customer', 'GET', '/api/pos/transactions/customer/1'),
      req('List Transactions by Terminal', 'GET', '/api/pos/transactions/terminal/1'),
      req('List Terminal Transactions (Alt)', 'GET', '/api/pos/terminals/1/transactions'),
      req('Get Transaction', 'GET', '/api/pos/transactions/{{transaction_id}}'),
      req('Get Transaction Full', 'GET', '/api/pos/transactions/{{transaction_id}}/full'),
      req('Get Transaction Payments', 'GET', '/api/pos/transactions/{{transaction_id}}/payments'),
      req('Get Payment Summary', 'GET', '/api/pos/transactions/{{transaction_id}}/payment-summary'),
      req('Void Transaction', 'POST', '/api/pos/transactions/{{transaction_id}}/void', { body: { voided_by: 1, reason: 'test' } })
    ]),
    // --- Geidea Payments ---
    folder('Geidea Payments', [
      req('Geidea Purchase', 'POST', '/api/pos/payments/geidea/purchase', {
        body: { terminal_id: 1, amount: '100.00', ecr_reference: 'POS-TXN-001' }
      }),
      req('Geidea Resolve Uncertain Payment', 'POST', '/api/pos/payments/geidea/resolve-uncertain', {
        body: { terminal_id: 1, ecr_reference: 'POS-TXN-001' }
      }),
      req('Geidea Ping Terminal', 'POST', '/api/pos/payments/geidea/ping', {
        body: { terminal_id: 1 }
      })
    ]),
    folder('POS Returns', [
      // Step 1: Call POST /api/auth/authorize-action to get an authorization_token first.
      // Step 2: Use that token in the X-Authorization-Token header here.
      req('Process Return (requires supervisor auth)', 'POST', '/api/pos/returns', {
        header: [
          H('x-tenant-id', '{{tenant_id}}'),
          H('Content-Type', 'application/json'),
          H('X-Authorization-Token', '{{authorization_token}}')
        ],
        body: {
          store_id: 1,
          cashier_id: 1,
          session_id: 1,
          original_transaction_id: 1,
          return_reason: 'Defective product',
          subtotal: '50.00',
          tax_amount: '5.00',
          total_refund_amount: '55.00',
          refund_method: 'cash',
          lines: [
            {
              product_id: 1,
              quantity: '1',
              unit_price: '50.00',
              refund_amount: '55.00',
              return_to_stock: true,
              condition: 'good'
            }
          ]
        },
        description: 'Process a sales return. Requires a valid supervisor authorization token in X-Authorization-Token header.\n\n**Flow:**\n1. Run `Authorize Action (Supervisor Override)` or `Authorize Action with Biometric` first — the test script will auto-save the token.\n2. Run this request — the `{{authorization_token}}` variable is used automatically.\n\nThe `authorized_by_user_id` of the approving supervisor is stored in the return metadata for audit.'
      })
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
      req('Get Session Summary', 'GET', '/api/cashier-sessions/{{session_id}}/summary'),
      req('List Sessions by Cashier', 'GET', '/api/cashier-sessions/list/{{cashier_id}}'),
      req('Get Z-Report', 'GET', '/api/reports/z-report?store_id={{store_id}}&date=2026-10-31')
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
    // --- Invoices ---
    folder('Invoices', [
      req('List Invoices', 'GET', '/api/invoices'),
      req('Create Invoice', 'POST', '/api/invoices', { body: {} }),
      req('Get Invoice Stats', 'GET', '/api/invoices/stats'),
      req('List Overdue Invoices', 'GET', '/api/invoices/overdue'),
      req('List Customer Invoices', 'GET', '/api/invoices/customer/1'),
      req('Get Invoice by Number', 'GET', '/api/invoices/by-number/INV-001'),
      req('Create Invoice from Order', 'POST', '/api/invoices/from-order/{{order_id}}', { body: {} }),
      req('Get Invoice', 'GET', '/api/invoices/1'),
      req('Update Invoice', 'PUT', '/api/invoices/1', { body: {} }),
      req('Update Invoice Status', 'PATCH', '/api/invoices/1/status', { body: { status: 'paid' } }),
      req('Delete Invoice', 'DELETE', '/api/invoices/1'),
      req('List Invoice Lines', 'GET', '/api/invoices/1/lines'),
      req('Create Invoice Line', 'POST', '/api/invoices/1/lines', { body: {} }),
      req('Delete Invoice Line', 'DELETE', '/api/invoices/lines/1'),
      req('List Invoice Payments', 'GET', '/api/invoices/1/payments'),
      req('Create Invoice Payment', 'POST', '/api/invoices/1/payments', { body: {} })
    ]),
    // --- Customers ---
    folder('Customers', [
      req('Create Customer', 'POST', '/api/customers', { body: { organization_id: 1, name: 'John Doe', code: 'CUST-001', customer_type: 'retail' } }),
      req('List Customers', 'GET', '/api/customers'),
      req('List Active Customers', 'GET', '/api/customers/active'),
      req('Search Customers', 'GET', '/api/customers/search?q=John'),
      req('Get Customers with Outstanding Balance', 'GET', '/api/customers/outstanding'),
      req('Get Customer by ID', 'GET', '/api/customers/1'),
      req('Get Customer by Code', 'GET', '/api/customers/code/CUST-001'),
      req('List Customers by Type', 'GET', '/api/customers/type/retail'),
      req('Get Customer Credit Status', 'GET', '/api/customers/1/credit-status'),
      req('Update Customer', 'PUT', '/api/customers/1', { body: {} }),
      req('Toggle Customer Active', 'PATCH', '/api/customers/1/active', { body: { is_active: true } }),
      req('Update Customer Balance', 'PATCH', '/api/customers/1/balance', { body: { amount: '100.00' } }),
      req('Delete Customer', 'DELETE', '/api/customers/1')
    ]),
    // --- Business Partners ---
    folder('Business Partners', [
      req('Create Business Partner', 'POST', '/api/business-partners', { body: { organization_id: 1, code: 'BP-001', name: 'Acme Supplies', partner_type: 'supplier' } }),
      req('List Business Partners', 'GET', '/api/business-partners'),
      req('Search Business Partners', 'GET', '/api/business-partners/search?q=Acme'),
      req('Get Business Partner', 'GET', '/api/business-partners/1'),
      req('Update Business Partner', 'PUT', '/api/business-partners/1', { body: {} }),
      req('Delete Business Partner', 'DELETE', '/api/business-partners/1'),
      req('Toggle Business Partner Active', 'PATCH', '/api/business-partners/1/toggle', { body: {} }),
      req('Add Partner Address', 'POST', '/api/business-partners/1/addresses', { body: {} }),
      req('List Partner Addresses', 'GET', '/api/business-partners/1/addresses'),
      req('Get Partner Address', 'GET', '/api/business-partners/1/addresses/1'),
      req('Update Partner Address', 'PUT', '/api/business-partners/1/addresses/1', { body: {} }),
      req('Delete Partner Address', 'DELETE', '/api/business-partners/1/addresses/1'),
      req('Set Default Partner Address', 'PATCH', '/api/business-partners/1/addresses/1/default', { body: {} }),
      req('Add Partner Contact', 'POST', '/api/business-partners/1/contacts', { body: {} }),
      req('List Partner Contacts', 'GET', '/api/business-partners/1/contacts'),
      req('Get Partner Contact', 'GET', '/api/business-partners/1/contacts/1'),
      req('Update Partner Contact', 'PUT', '/api/business-partners/1/contacts/1', { body: {} }),
      req('Delete Partner Contact', 'DELETE', '/api/business-partners/1/contacts/1'),
      req('Set Primary Partner Contact', 'PATCH', '/api/business-partners/1/contacts/1/primary', { body: {} }),
      req('Create Partner Address (Standalone)', 'POST', '/api/partner-addresses', { body: {} }),
      req('List Partner Addresses (Standalone)', 'GET', '/api/partner-addresses'),
      req('Get Partner Address by ID', 'GET', '/api/partner-addresses/1'),
      req('Update Partner Address by ID', 'PUT', '/api/partner-addresses/1', { body: {} }),
      req('Delete Partner Address by ID', 'DELETE', '/api/partner-addresses/1'),
      req('Set Default Partner Address by ID', 'PATCH', '/api/partner-addresses/1/default', { body: {} }),
      req('Create Partner Contact (Standalone)', 'POST', '/api/partner-contacts', { body: {} }),
      req('List Partner Contacts (Standalone)', 'GET', '/api/partner-contacts'),
      req('Get Partner Contact by ID', 'GET', '/api/partner-contacts/1'),
      req('Update Partner Contact by ID', 'PUT', '/api/partner-contacts/1', { body: {} }),
      req('Delete Partner Contact by ID', 'DELETE', '/api/partner-contacts/1'),
      req('Set Primary Partner Contact by ID', 'PATCH', '/api/partner-contacts/1/primary', { body: {} })
    ]),
    // --- BP Price Contracts ---
    folder('BP Price Contracts', [
      req('Create BP Price Contract', 'POST', '/api/bp-price-contracts', { body: {} }),
      req('List BP Price Contracts', 'GET', '/api/bp-price-contracts'),
      req('Get Effective BP Price Contract', 'GET', '/api/bp-price-contracts/effective?partner_id=1&product_id=1'),
      req('List Contracts by Partner', 'GET', '/api/bp-price-contracts/partner/1'),
      req('Get BP Price Contract', 'GET', '/api/bp-price-contracts/1'),
      req('Update BP Price Contract', 'PUT', '/api/bp-price-contracts/1', { body: {} }),
      req('Delete BP Price Contract', 'DELETE', '/api/bp-price-contracts/1'),
      req('Toggle BP Price Contract Active', 'PATCH', '/api/bp-price-contracts/1/toggle', { body: {} })
    ]),
    // --- Payment Terms ---
    folder('Payment Terms', [
      req('Create Payment Term', 'POST', '/api/payment-terms', { body: { organization_id: 1, code: 'NET30', name: 'Net 30 Days', due_days: 30 } }),
      req('List Payment Terms', 'GET', '/api/payment-terms'),
      req('Get Payment Term', 'GET', '/api/payment-terms/1'),
      req('Update Payment Term', 'PUT', '/api/payment-terms/1', { body: {} }),
      req('Delete Payment Term', 'DELETE', '/api/payment-terms/1')
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
      req('Get Organization ZATCA Settings', 'GET', '/api/organizations/1/zatca'),
      req('Update Organization ZATCA Settings', 'PUT', '/api/organizations/1/zatca', { body: {} }),
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
      req('Delete User', 'DELETE', '/api/users/1'),
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
    // --- Product Catalog & Categories ---
    folder('Product Catalog', [
      req('List Products with Variants', 'GET', '/api/products/catalog'),
      req('Get Master Product Catalog', 'GET', '/api/products/master-catalog'),
      req('Get Master Product Stock Catalog', 'GET', '/api/products/master-catalog/stock?store_id={{store_id}}'),
      req('Get Master Catalog by Category', 'GET', '/api/products/master-catalog/category/1'),
      req('Search Master Product Catalog', 'GET', '/api/products/master-catalog/search?q=test'),
      req('Get Master Catalog by Product ID', 'GET', '/api/products/master-catalog/product/{{product_id}}')
    ]),
    folder('Product Categories', [
      req('Create Product Category', 'POST', '/api/product-categories', { body: { organization_id: 1, name: 'Beverages', code: 'BEV' } }),
      req('List Product Categories', 'GET', '/api/product-categories'),
      req('Get Category Hierarchy', 'GET', '/api/product-categories/hierarchy'),
      req('Get Product Category', 'GET', '/api/product-categories/1'),
      req('Get Category by Code', 'GET', '/api/product-categories/code/BEV'),
      req('List Category Children', 'GET', '/api/product-categories/1/children'),
      req('Update Product Category', 'PUT', '/api/product-categories/1', { body: {} }),
      req('Delete Product Category', 'DELETE', '/api/product-categories/1')
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
    // --- UOM Packaging Templates ---
    folder('UOM Packaging Templates', [
      req('Create Template', 'POST', '/api/uom-packaging-templates', { body: {} }),
      req('Create Template Pipeline', 'POST', '/api/uom-packaging-templates/pipeline', { body: {} }),
      req('List Templates', 'GET', '/api/uom-packaging-templates'),
      req('Get Templates by UOM ID', 'GET', '/api/uom-packaging-templates/by-uom/1'),
      req('Get Template', 'GET', '/api/uom-packaging-templates/1'),
      req('Get Template with Levels', 'GET', '/api/uom-packaging-templates/1/full'),
      req('Update Template', 'PUT', '/api/uom-packaging-templates/1', { body: {} }),
      req('Delete Template', 'DELETE', '/api/uom-packaging-templates/1'),
      req('Create Level', 'POST', '/api/uom-packaging-templates/1/levels', { body: {} }),
      req('List Levels', 'GET', '/api/uom-packaging-templates/1/levels'),
      req('Update Level', 'PUT', '/api/uom-packaging-template-levels/1', { body: {} }),
      req('Delete Level', 'DELETE', '/api/uom-packaging-template-levels/1')
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
    // --- Stock Movements ---
    folder('Stock Movements', [
      req('Create Stock Movement', 'POST', '/api/stock-movements', { body: {} }),
      req('List Stock Movements', 'GET', '/api/stock-movements'),
      req('List Reference Types', 'GET', '/api/stock-movements/reference-types'),
      req('List by Reference Type', 'GET', '/api/stock-movements/by-reference-type?reference_type=purchase_order'),
      req('Get Stock Movement', 'GET', '/api/stock-movements/1'),
      req('Update Stock Movement Status', 'PATCH', '/api/stock-movements/1/status', { body: { status: 'completed' } }),
      req('List by Product', 'GET', '/api/stock-movements/product/{{product_id}}'),
      req('List by Product Date Range', 'GET', '/api/stock-movements/product/{{product_id}}/daterange?start_date=2026-01-01&end_date=2026-12-31'),
      req('List by Date Range', 'GET', '/api/stock-movements/daterange?start_date=2026-01-01&end_date=2026-12-31')
    ]),
    // --- Stock Counts ---
    folder('Stock Counts', [
      req('Create Stock Count', 'POST', '/api/stock-counts', { body: {} }),
      req('List Stock Counts', 'GET', '/api/stock-counts'),
      req('Get Stock Count', 'GET', '/api/stock-counts/1'),
      req('Update Stock Count', 'PUT', '/api/stock-counts/1', { body: {} }),
      req('Delete Stock Count', 'DELETE', '/api/stock-counts/1'),
      req('Start Stock Count', 'POST', '/api/stock-counts/1/start', { body: {} }),
      req('Complete Stock Count', 'POST', '/api/stock-counts/1/complete', { body: {} }),
      req('Approve Stock Count', 'POST', '/api/stock-counts/1/approve', { body: {} }),
      req('Reconcile Stock Count', 'POST', '/api/stock-counts/1/reconcile', { body: {} }),
      req('Get Stock Count Summary', 'GET', '/api/stock-counts/1/summary'),
      req('Add Stock Count Line', 'POST', '/api/stock-counts/1/lines', { body: {} }),
      req('List Stock Count Lines', 'GET', '/api/stock-counts/1/lines'),
      req('Get Stock Count Line', 'GET', '/api/stock-counts/1/lines/1'),
      req('Update Stock Count Line', 'PUT', '/api/stock-counts/1/lines/1', { body: {} }),
      req('Delete Stock Count Line', 'DELETE', '/api/stock-counts/1/lines/1'),
      req('Bulk Update Stock Count Lines', 'PUT', '/api/stock-counts/1/lines', { body: {} })
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
    ]),
    // --- Promotions & Coupons ---
    folder('Promotions', [
      req('Create Promotion', 'POST', '/api/promotions', {
        body: {
          organization_id: 1,
          code: 'PROMO-SUMMER20',
          name: 'Summer Sale 20%',
          description: '20% off all summer items',
          promotion_type: 'percentage_discount',
          discount_value: '20.00',
          coupon_code: 'SUMMER20',
          valid_from: '2026-06-01T00:00:00Z',
          valid_to: '2026-08-31T23:59:59Z',
          applies_to: 'all',
          min_order_amount: '100.00',
          usage_limit: 100,
          is_active: true,
          is_stackable: false
        }
      }),
      req('List All Promotions', 'GET', '/api/promotions?organization_id=1&limit=50&offset=0'),
      req('List Active Promotions', 'GET', '/api/promotions/active?organization_id=1'),
      req('Get Promotion by ID', 'GET', '/api/promotions/1'),
      req('Get Promotion by Code', 'GET', '/api/promotions/code/PROMO-SUMMER20?organization_id=1'),
      req('Update Promotion', 'PUT', '/api/promotions/1', {
        body: {
          name: 'Summer Sale 25%',
          discount_value: '25.00',
          coupon_code: 'SUMMER25',
          valid_to: '2026-09-30T23:59:59Z'
        }
      }),
      req('Update Promotion Status', 'PATCH', '/api/promotions/1/status', {
        body: { is_active: true }
      }),
      req('Delete Promotion', 'DELETE', '/api/promotions/1'),
      req('Apply Coupon to Cart', 'POST', '/api/promotions/apply-coupon', {
        body: {
          cart_id: '{{cart_id}}',
          coupon_code: 'SUMMER20',
          organization_id: 1
        },
        description: 'Apply a coupon to a cart. Supports: percentage_discount, fixed_discount, buy_x_get_y, happy_hour, points_multiplier, bundle_price, free_item.'
      }),
      req('Validate Coupon (dry-run)', 'POST', '/api/promotions/validate-coupon', {
        body: {
          cart_id: '{{cart_id}}',
          coupon_code: 'SUMMER20',
          organization_id: 1
        },
        description: 'Check if a coupon is valid without applying it. Returns errors and estimated discount.'
      }),
      // ── Example payloads for each promotion type ──
      req('[Example] Create buy_x_get_y', 'POST', '/api/promotions', {
        body: {
          organization_id: 1,
          code: 'PROMO-BUY2GET1',
          name: 'Buy 2 Get 1 Free',
          promotion_type: 'buy_x_get_y',
          action_metadata: { buy_id: 101, buy_qty: 2, get_id: 205, get_qty: 1, discount_percent: 100 },
          applies_to: 'product',
          target_product_ids: [101, 205],
          coupon_code: 'BUY2GET1',
          is_active: true
        }
      }),
      req('[Example] Create happy_hour', 'POST', '/api/promotions', {
        body: {
          organization_id: 1,
          code: 'PROMO-HAPPY-HOUR',
          name: 'Happy Hour 20% Off',
          promotion_type: 'happy_hour',
          discount_value: '20.00',
          schedule_json: { days: ['Monday', 'Tuesday', 'Wednesday', 'Thursday'], start_time: '20:00', end_time: '22:00' },
          applies_to: 'all',
          coupon_code: 'HAPPY20',
          is_active: true
        }
      }),
      req('[Example] Create points_multiplier', 'POST', '/api/promotions', {
        body: {
          organization_id: 1,
          code: 'PROMO-2X-POINTS',
          name: '2x Loyalty Points',
          promotion_type: 'points_multiplier',
          action_metadata: { multiplier: 2 },
          applies_to: 'all',
          coupon_code: 'DOUBLE',
          is_active: true
        }
      }),
      req('[Example] Create fixed_discount + min_spend', 'POST', '/api/promotions', {
        body: {
          organization_id: 1,
          code: 'PROMO-50OFF500',
          name: 'Spend 500, Save 50',
          promotion_type: 'fixed_discount',
          discount_value: '50.00',
          min_order_amount: '500.00',
          applies_to: 'all',
          coupon_code: 'SAVE50',
          is_active: true
        }
      })
    ]),
    // --- Loyalty Redemption Rules ---
    folder('Loyalty Rules', [
      req('Create Loyalty Rule', 'POST', '/api/loyalty-rules', {
        body: {
          organization_id: 1,
          rule_name: 'Standard Loyalty',
          points_earning_rate: '1.0000',
          points_redemption_rate: '0.0100',
          min_points_to_redeem: '100.00',
          max_points_per_txn: '500.00',
          max_redemption_percent: '20.00',
          eligible_product_types: [],
          expiry_days: 365,
          is_active: true,
          valid_from: '2026-01-01',
          valid_to: '2026-12-31'
        }
      }),
      req('List Loyalty Rules', 'GET', '/api/loyalty-rules?organization_id=1'),
      req('Get Active Loyalty Rule', 'GET', '/api/loyalty-rules/active?organization_id=1',
        { description: 'Returns the current active rule used at POS checkout to compute point value.' }),
      req('Get Loyalty Rule by ID', 'GET', '/api/loyalty-rules/1'),
      req('Update Loyalty Rule', 'PUT', '/api/loyalty-rules/1', {
        body: {
          rule_name: 'Premium Loyalty',
          points_earning_rate: '1.5000',
          min_points_to_redeem: '50.00',
          max_redemption_percent: '30.00',
          is_active: true
        }
      }),
      req('Toggle Loyalty Rule Active', 'PATCH', '/api/loyalty-rules/1/active', {
        body: { is_active: false }
      }),
      req('Delete Loyalty Rule', 'DELETE', '/api/loyalty-rules/1'),
      // ── Customer loyalty point endpoints ──────────────────────────────────
      req('Get Customer Loyalty Balance', 'GET', '/api/customers/1/loyalty-balance',
        { description: 'Lightweight POS check before redemption. Returns id, name, and loyalty_points.' }),
      req('Adjust Customer Loyalty Points', 'PATCH', '/api/customers/1/loyalty-points', {
        body: { points: '-50.00' },
        description: 'Pass a negative value to redeem/deduct. Pass a positive value to grant bonus points.'
      })
    ]),
    // --- M2M Tokens ---
    folder('M2M Tokens', [
      req('Create M2M Token', 'POST', '/api/m2m/tokens', {
        auth: authNoAuth,
        header: [...tenantHeader, H('Content-Type', 'application/json'), H('x-m2m-key', '{{m2m_api_key}}')],
        description: 'Generate and save an M2M token using the shared middleware key. No user JWT required.',
        body: {
          client_id: 'pos-App',
          client_name: 'pos-App',
          scopes: ['products:read', 'orders:write'],
          years: 5
        }
      }),
      req('List M2M Tokens', 'GET', '/api/m2m/tokens', {
        auth: authNoAuth,
        header: [...tenantHeader, H('x-m2m-key', '{{m2m_api_key}}')],
        description: 'List saved tenant M2M tokens using the shared middleware key. No JWT required.'
      })
    ]),
    // --- Purchase Orders & RFQs ---
    folder('Purchase Orders', [
      req('Create Purchase Order', 'POST', '/api/purchase-orders', { body: {} }),
      req('List Purchase Orders', 'GET', '/api/purchase-orders'),
      req('Get Purchase Order', 'GET', '/api/purchase-orders/1'),
      req('Update Purchase Order', 'PUT', '/api/purchase-orders/1', { body: {} }),
      req('Update Purchase Order Status', 'PATCH', '/api/purchase-orders/1/status', { body: { status: 'approved' } }),
      req('Approve Purchase Order', 'POST', '/api/purchase-orders/1/approve', { body: {} }),
      req('Delete Purchase Order', 'DELETE', '/api/purchase-orders/1'),
      req('List Purchase Order Lines', 'GET', '/api/purchase-orders/1/lines'),
      req('Add Purchase Order Line', 'POST', '/api/purchase-orders/1/lines', { body: {} }),
      req('Get Purchase Order Line', 'GET', '/api/purchase-orders/1/lines/1'),
      req('Update Purchase Order Line', 'PUT', '/api/purchase-orders/1/lines/1', { body: {} }),
      req('Delete Purchase Order Line', 'DELETE', '/api/purchase-orders/1/lines/1'),
      req('List All Purchase Order Lines (Standalone)', 'GET', '/api/purchase-order-lines'),
      req('Get Purchase Order Line by ID (Standalone)', 'GET', '/api/purchase-order-lines/1'),
      req('Update Purchase Order Line by ID (Standalone)', 'PUT', '/api/purchase-order-lines/1', { body: {} }),
      req('Delete Purchase Order Line by ID (Standalone)', 'DELETE', '/api/purchase-order-lines/1')
    ]),
    folder('RFQs', [
      req('List RFQs', 'GET', '/api/rfqs'),
      req('Create RFQ', 'POST', '/api/rfqs', { body: {} }),
      req('Get RFQ', 'GET', '/api/rfqs/1'),
      req('Publish RFQ', 'POST', '/api/rfqs/1/publish', { body: {} }),
      req('Close RFQ', 'POST', '/api/rfqs/1/close', { body: {} }),
      req('List RFQ Responses', 'GET', '/api/rfqs/1/responses'),
      req('Public Get RFQ by Token', 'GET', '/api/public/rfqs/SAMPLE-TOKEN', { auth: authNoAuth }),
      req('Public Submit RFQ Response', 'POST', '/api/public/rfqs/SAMPLE-TOKEN/responses', { auth: authNoAuth, body: {} })
    ]),
    // --- Warehouse Transfer Requests ---
    folder('Transfer Requests', [
      req('Create Transfer Request', 'POST', '/api/transfer-requests', {
        body: {
          organization_id: 1,
          transfer_number: 'TR-20260727-001',
          from_store_id: 1,
          to_store_id: 2,
          requested_by: 1,
          expected_delivery_date: '2026-07-30',
          notes: 'Urgent inter-store restock',
          items: [
            {
              product_id: 1,
              requested_quantity: 50,
              uom_id: 1,
              notes: 'Restock front shelf'
            }
          ]
        }
      }),
      req('List Transfer Requests', 'GET', '/api/transfer-requests?organization_id=1&limit=50&offset=0'),
      req('Get Transfer Request by ID', 'GET', '/api/transfer-requests/1'),
      req('Update Transfer Request', 'PUT', '/api/transfer-requests/1', { body: {} }),
      req('Approve Transfer Request', 'POST', '/api/transfer-requests/1/approve', {
        body: { approved_by: 1 },
        description: 'State transition: Draft/Pending -> Approved'
      }),
      req('Ship Transfer Request', 'POST', '/api/transfer-requests/1/ship', {
        body: { shipped_by: 1 },
        description: 'Deducts stock at source, increments quantity_in_transit at destination'
      }),
      req('Receive Transfer Request', 'POST', '/api/transfer-requests/1/receive', {
        body: { received_by: 1 },
        description: 'Clears in-transit stock, increments quantity_on_hand and available at destination'
      }),
      req('Cancel Transfer Request', 'POST', '/api/transfer-requests/1/cancel', {
        body: { cancelled_by: 1, reason: 'No longer needed' }
      })
    ]),
    // --- Goods Receipt Notes (GRN) ---
    folder('Goods Receipt Notes (GRN)', [
      req('List Goods Receipt Notes', 'GET', '/api/goods-receipt-notes'),
      req('Create Goods Receipt Note', 'POST', '/api/goods-receipt-notes', {
        body: {
          organization_id: 1,
          grn_number: 'GRN-20260727-001',
          purchase_order_id: 1,
          supplier_id: 1,
          store_id: 1,
          received_by: 1,
          delivery_note_number: 'DN-998822',
          notes: 'Received delivery at warehouse dock',
          items: [
            {
              purchase_order_line_id: 1,
              product_id: 1,
              quantity_received: 100,
              quantity_rejected: 0,
              unit_cost: 15.50,
              batch_number: 'BATCH-2026A',
              expiry_date: '2027-12-31'
            }
          ]
        }
      }),
      req('Get Goods Receipt Note by ID', 'GET', '/api/goods-receipt-notes/1'),
      req('Update Goods Receipt Note', 'PUT', '/api/goods-receipt-notes/1', { body: {} }),
      req('Delete Goods Receipt Note', 'DELETE', '/api/goods-receipt-notes/1'),
      req('Post / Process Goods Receipt Note', 'POST', '/api/goods-receipt-notes/1/post', {
        description: 'Updates PO lines received qty, updates PO status, increments inventory stock, logs stock movement'
      })
    ]),
    // --- Approvals ---
    folder('Approvals', [
      req('List Approval Policies', 'GET', '/api/approvals/policies'),
      req('Get Approval Policy by Domain', 'GET', '/api/approvals/policies/goods_receipt_note'),
      req('Save Approval Policy', 'PUT', '/api/approvals/policies/goods_receipt_note', { body: {} }),
      req('Delete Approval Policy', 'DELETE', '/api/approvals/policies/goods_receipt_note'),
      req('Submit Approval Request', 'POST', '/api/approvals/submit', { body: {} }),
      req('List Approvals', 'GET', '/api/approvals'),
      req('Get Approval by ID', 'GET', '/api/approvals/1'),
      req('Delete Approval', 'DELETE', '/api/approvals/1'),
      req('Approve Request', 'POST', '/api/approvals/1/approve', { body: { comment: 'Approved' } }),
      req('Reject Request', 'POST', '/api/approvals/1/reject', { body: { comment: 'Rejected' } }),
      req('Cancel Request', 'POST', '/api/approvals/1/cancel', { body: {} })
    ]),
    // --- Notifications ---
    folder('Notifications', [
      req('List User Notifications', 'GET', '/api/notifications'),
      req('Get Unread Notification Count', 'GET', '/api/notifications/unread-count'),
      req('Mark Notification as Read', 'POST', '/api/notifications/1/read', { body: {} }),
      req('Mark All Notifications as Read', 'POST', '/api/notifications/read-all', { body: {} }),
      req('List Notification Channels', 'GET', '/api/notifications/channels'),
      req('Get Notification Channel by Type', 'GET', '/api/notifications/channels/whatsapp'),
      req('Upsert Notification Channel', 'PUT', '/api/notifications/channels/whatsapp', { body: { is_enabled: true, config: {} } }),
      req('List Notification Templates', 'GET', '/api/notifications/templates'),
      req('Upsert Notification Template', 'POST', '/api/notifications/templates', { body: {} }),
      req('Delete Notification Template', 'DELETE', '/api/notifications/templates/1'),
      req('List Notification Routing Rules', 'GET', '/api/notifications/rules'),
      req('Create Notification Routing Rule', 'POST', '/api/notifications/rules', { body: {} }),
      req('Update Notification Routing Rule', 'PUT', '/api/notifications/rules/1', { body: {} }),
      req('Delete Notification Routing Rule', 'DELETE', '/api/notifications/rules/1')
    ]),
    // --- Backup ---
    folder('Backup', [
      req('Get Backup Streaming Status', 'GET', '/api/backup/status/{{tenant_id}}'),
      req('Get gRPC Backup Info', 'GET', '/api/backup/grpc-info/{{tenant_id}}'),
      req('Trigger Server-Side Backup', 'POST', '/api/backup/trigger', { body: { tenant: '{{tenant_id}}' } }),
      req('List Backup Jobs', 'GET', '/api/backup/jobs'),
      req('Get Backup Job Status', 'GET', '/api/backup/jobs/{{backup_job_id}}')
    ]),
    // --- Releases & Auto-Update ---
    folder('Releases & Auto-Updates', [
      req('Upload Release Artifact', 'POST', '/api/v1/releases/upload', {
        auth: authNoAuth,
        header: [H('X-Release-Secret', 'your_release_secret')],
        description: 'Multipart form-data upload for CI/CD release binaries.'
      }),
      req('Check Latest App Update', 'GET', '/api/v1/updates/pos-client/latest?platform=windows&arch=amd64', {
        auth: authNoAuth,
        header: [H('X-Nembus-App-Key', 'your_app_key')]
      }),
      req('Download App Release', 'GET', '/api/v1/updates/pos-client/download/1', {
        auth: authNoAuth,
        header: [H('X-Nembus-App-Key', 'your_app_key')]
      })
    ]),
    // --- SAP Migration ---
    folder('SAP Migration', [
      req('Ingest SAP Migration Batch', 'POST', '/api/v1/migration/batch', {
        auth: authNoAuth,
        header: [H('x-tenant-id', '{{tenant_id}}'), H('X-Agent-API-Key', 'your_agent_api_key'), H('Content-Type', 'application/json')],
        body: { entity_type: 'items', records: [] }
      })
    ]),
    // --- ZATCA Phase 2 & Sync ---
    folder('ZATCA Phase 2 & Sync', [
      req('Get ZATCA Status', 'GET', '/api/zatca/status', {
        description: 'Returns ZATCA enabled flag, environment mode (sandbox/production), base URL, and Org VAT ID'
      }),
      req('List ZATCA Documents', 'GET', '/api/zatca/documents?store_id={{store_id}}'),
      req('Get Receipt QR Code', 'GET', '/api/zatca/receipt-qr?transaction_id={{transaction_id}}'),
      req('Get ZATCA Config Deltas (Pull Sync)', 'GET', '/api/zatca/configs?store_id=1&since=2026-01-01T00:00:00Z', {
        description: 'Delta-fetch ZATCA device configs, active CSID certificates, and revocation status for POS terminals'
      }),
      req('Receive Push Sync Payload (Outbox Push)', 'POST', '/api/zatca/sync/push', {
        body: {
          store_id: 1,
          items: [
            {
              id: 101,
              entity_type: 'pos_transaction',
              entity_id: 5001,
              action: 'INSERT',
              payload: {
                transaction_number: 'POS-2026-0001',
                total_amount: 115.00,
                vat_amount: 15.00,
                zatca_signed_xml: '<Invoice>...</Invoice>',
                qr_code_base64: 'AQZTZWxsZXIS...'
              },
              created_at: '2026-07-29T01:00:00Z'
            }
          ]
        },
        description: 'Ingests offline signed transactions and operational outbox payloads pushed from POS terminals'
      }),
      req('Onboard ZATCA Device (CSID)', 'POST', '/api/zatca/onboard', {
        body: { terminal_id: 1, otp: '123456' }
      }),
      req('Complete Device Onboarding', 'POST', '/api/zatca/devices/1/complete-onboarding', {
        body: { otp: '123456' }
      }),
      req('Submit POS Transaction to ZATCA', 'POST', '/api/zatca/transactions/{{transaction_id}}/submit', { body: {} }),
      req('Check POS Transaction ZATCA Compliance', 'POST', '/api/zatca/transactions/{{transaction_id}}/check', { body: {} }),
      req('Generate ZATCA Document for Sales Order', 'POST', '/api/zatca/orders/{{order_id}}/generate', { body: {} }),
      req('Get Sales Order ZATCA Document Status', 'GET', '/api/zatca/orders/{{order_id}}/status'),
      req('Check Sales Order ZATCA Compliance', 'POST', '/api/zatca/orders/{{order_id}}/check', { body: {} }),
      req('Submit Sales Order to ZATCA', 'POST', '/api/zatca/orders/{{order_id}}/submit', { body: {} }),
      req('Get Terminal ZATCA Config', 'GET', '/api/zatca/terminals/1')
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
console.log('API folders included:', collection.item.length);
