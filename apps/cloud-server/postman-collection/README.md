# NEMBUS API – Postman Collection

Postman collection and environment for testing the **NEMBUS** backend API and the **Local Biometric Hardware Bridge**. All routes from `packages/core/routing` (with JWT + Tenant middleware) and `http://127.0.0.1:5050/api/v1` are included. Uses a **unified base URL**, **tenant**, and **token**.

## Setup

1. **Import in Postman**  
   Use [Importing data into Postman](https://learning.postman.com/docs/getting-started/importing-and-exporting/importing-data/):
   - In Postman: **Import** (or drag and drop).
   - Select **Collection:** `Nembus-API.postman_collection.json` and **Environment:** `Nembus-Local.postman_environment.json`.
   - Click **Import**.

2. **Select environment**
   - In the top-right dropdown, choose **NEMBUS Local**.  
   - This ensures **URLs show** in the request bar (e.g. `{{base_url}}/api/...` resolves to `http://localhost:8080/api/...`, and `{{biometric_bridge_url}}/api/v1/...` resolves to `http://127.0.0.1:5050/api/v1/...`).

3. **Get a token**
   - Run **Auth → Login** (or **Auth → Login with Biometric**).
   - Use `x-tenant-id` as your tenant (e.g. `default`; must match a tenant in your DB).
   - Body example: `{ "user_login": "admin", "password": "your_password" }`.
   - From the response, copy the JWT into the environment variable **token** (or set the collection variable **token**). All other requests use **Authorization: Bearer {{token}}** and **x-tenant-id: {{tenant_id}}**.

4. **Start the API**
   - Ensure the backend is running (e.g. `http://localhost:8080`). Change **base_url** in the environment if you use a different host/port.

## Variables (unified)

| Variable               | Description                                                 | Example / note                          |
|------------------------|-------------------------------------------------------------|-----------------------------------------|
| `base_url`             | Cloud / Backend API base URL                                | `http://localhost:8080`                 |
| `biometric_bridge_url` | Local Fingerprint Hardware Bridge base URL                  | `http://127.0.0.1:5050`                 |
| `tenant_id`            | Tenant identifier (slug) for `x-tenant-id`                  | `default` (must exist in DB)            |
| `token`                | JWT from **Auth → Login**                                   | Set after login                         |

## SAP Agent gRPC (sap-agent)

Cloud-server exposes **`sapagent.SapAgentService`** on the gRPC port (environment `GRPCPort`, often `50051`). Use **grpcurl** with reflection enabled:

```bash
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext -H "authorization: Bearer {{token}}" -H "x-tenant-id: {{tenant_id}}" \
  localhost:50051 sapagent.SapAgentService/ResolveTenant
```

RPCs: `ResolveTenant`, `StreamMigrationIngest`, `StreamMasterDataUpsert`, `StreamSapOutbox`, `AckSapOutbox`, `HoldSapOutbox`, `GetReconciliationSummary`. Proto: `packages/core/grpc/sapagentpb/sap_agent.proto`.
| `authorization_token`  | Supervisor override token (`X-Authorization-Token`)         | Auto-saved by Authorize Action requests |
| `fmd_base64`           | Captured/consolidated ISO 19794-2:2005 fingerprint template | Auto-saved by Bridge Capture/Enrollment |
| `store_id`             | Store ID for store-scoped requests                          | `1` (see sample data below)             |
| `cart_id`              | Cart ID (set after creating a cart)                         | —                                       |
| `order_id`             | Order ID                                                    | —                                       |
| `transaction_id`       | POS transaction ID                                          | `1`                                     |
| `cashier_id`           | Cashier ID                                                  | `1`                                     |
| `session_id`           | Cashier session ID                                          | `1`                                     |
| `user_id`              | User ID                                                     | `1`                                     |

## End-to-End Fingerprint Enrollment Flow in Postman

1. **Check Reader Connection:** Run **Biometrics (Local Hardware Bridge) → Check Connected Fingerprint Device** (`GET {{biometric_bridge_url}}/api/v1/device`) or **Get Connected Device Info** (`GET {{biometric_bridge_url}}/api/v1/device/info`).
2. **Capture & Consolidate 4 Scans:** Run **Biometrics (Local Hardware Bridge) → Multi-Sample Enrollment Capture (4 Scans -> Consolidated FMD)** (`POST {{biometric_bridge_url}}/api/v1/enrollment`), and scan the same finger 4 times on the DigitalPersona reader. The test script automatically saves `consolidated_fmd` into `{{fmd_base64}}`.
3. **Save Template to Backend:** Run **Biometrics → Enroll User Fingerprint (by User Path ID)** (`POST {{base_url}}/api/users/{{user_id}}/biometrics/enroll`) or **Enroll User Fingerprint (Direct Endpoint)** (`POST {{base_url}}/api/biometrics/enroll`), which automatically uses `{{fmd_base64}}`.

### Fingerprint capture troubleshooting (Windows)

If **Capture Single Fingerprint** returns `408` with *timed out waiting for finger* while the reader is connected:

1. **Stop Windows Biometric Service** so it does not hold the USB reader: open **Services** (`services.msc`), stop **Windows Biometric Service** (`WbioSrvc`), then retry capture.
2. **Start capture before touching the sensor** — click **Send** in Postman, wait for the reader to arm (LED), then place the **pad** of your finger flat on the glass.
3. **Rebuild/restart the POS app** after pulling biometric bridge changes so `NEMBUS-dev` loads the updated embedded bridge on port `5050`.
4. Clean the sensor (U.are.U 4500: gently dab with sticky tape; do not use solvents).

## Sample data reference

Use these scripts to seed data so the collection requests return meaningful results:

- **`scripts/init-Data-Dump.sql`** – Core master data:
  - **Organization:** id `1`, code `ORG001` (Qitaf Group).
  - **Stores:** e.g. `RYD-001`, `JED-001`, `DMM-001`, `WH-RYD-001`, `WHSL-RYD-001` (store ids 1–5); POS-enabled retail/warehouse/wholesale.
  - **Product categories**, **brands**, **products**, **price lists** (e.g. `RETAIL_SAR`, `WHOLESALE_SAR`), **tax** (`VAT_15`), **UOMs** (PCS, KG, LTR, BOX, etc.), **storage locations**, **inventory stock** for RYD-001.

- **`scripts/init-Restaurant-Data-Dump.sql`** – Restaurant data (run after init-Data-Dump.sql):
  - **Store:** `REST-001` (NasaR Cafe & Restaurant), restaurant type.
  - **Categories:** Restaurant Ingredients, Menu Items (Prepared), and subcategories (Proteins, Vegetables, etc.).
  - **Products:** Ingredients (chicken, beef, tomato, etc.) and menu items (e.g. Double Espresso, Club Sandwich).

## Collection structure (all routing modules)

Every module registered in `apps/cloud-server/main.go` and `packages/core/routing` is included:

- **Auth** – Login, Login with Biometric (1:1 / 1:N), Authorize Action (Supervisor Override), Authorize Action with Biometric.
- **Biometrics (Local Hardware Bridge)** – Device status, Device info, Single capture, Cancel capture, Multi-sample 4-scan enrollment, Interactive enrollment (`start` & `capture-step`).
- **Biometrics** – List user enrolled biometrics, Enroll user fingerprint (`/api/users/:id/biometrics/enroll` & `/api/biometrics/enroll`), Deactivate user fingerprint template.
- **Dev** – Get Dev Token (`GET /api/dev/token`, dev only).
- **Tenants (public)** – Get by slug, List active.
- **Carts & Cart Items** – Full CRUD, items, totals, checkout, activities.
- **POS, Geidea Payments, POS Returns & POS Terminals** – Categories, products, payments, Geidea ECR terminal purchase/resolve/ping, store products/transactions, returns, terminals.
- **Cashier Sessions & Cashiers** – Open, get, close, summary, list by cashier, Z-Report, cashier CRUD & limits.
- **Orders, Order Lines, Order Fulfillments & Invoices** – Full CRUD, lines, status, fulfillments, commercial invoices.
- **Customers, Business Partners, BP Price Contracts & Payment Terms** – Full CRUD, addresses, contacts, price contracts, payment terms.
- **Stores, Organizations & Tenants** – CRUD, ZATCA organization settings, hierarchy.
- **Users, Roles, Permissions, Modules, Menus, Submenus & Navigation** – Full RBAC & navigation management.
- **Images** – Upload (by module).
- **Product Catalog, Product Categories, Brands, Price Lists, Tax Categories, UOMs & UOM Packaging Templates** – Catalog, categories, brands, pricing, tax, UOMs, packaging levels.
- **Product Prices, Product Variants & Product Barcodes** – Effective prices, variants, barcodes.
- **Inventory Stock, Stock Movements, Stock Counts & Storage Locations** – Stock CRUD, movements, physical count workflows, locations.
- **Restaurant** – Tables, menu categories, menu items, modifiers, orders.
- **Promotions & Loyalty Rules** – Promotions, coupons, loyalty rules, customer loyalty balances.
- **M2M Tokens** – Generate and list M2M tokens.
- **Purchase Orders, RFQs, Transfer Requests & Goods Receipt Notes (GRN)** – Full procurement and warehouse logistics workflows.
- **Approvals & Notifications** – Approval policies/requests, in-app notification center, channels, templates, routing rules.
- **Backup, Releases & Auto-Updates, SAP Migration & ZATCA Phase 2 & Sync** – Backup streaming/jobs, release artifacts, SAP batch ingestion, ZATCA onboarding/compliance/sync.

For request/response shapes, see **Swagger** at `http://localhost:8080/swagger/index.html` when the API is running.

## Regenerating the collection

To regenerate the collection from the script (e.g. after adding routes):  
`node apps/cloud-server/postman-collection/generate-collection.js`
