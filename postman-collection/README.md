# NEMBUS API – Postman Collection

Postman collection and environment for testing the **NEMBUS** backend API. All routes from `internal/routing` (with JWT + Tenant middleware) are included. Uses a **unified base URL**, **tenant**, and **token**.

## Setup

1. **Import in Postman**  
   Use [Importing data into Postman](https://learning.postman.com/docs/getting-started/importing-and-exporting/importing-data/):
   - In Postman: **Import** (or drag and drop).
   - Select **Collection:** `Nembus-API.postman_collection.json` and **Environment:** `Nembus-Local.postman_environment.json`.
   - Click **Import**.

2. **Select environment**
   - In the top-right dropdown, choose **NEMBUS Local**.  
   - This ensures **URLs show** in the request bar (e.g. `{{base_url}}/api/...` resolves to `http://localhost:8080/api/...`).

3. **Get a token**
   - Run **Auth → Login**.
   - Use `x-tenant-id` as your tenant (e.g. `default`; must match a tenant in your DB).
   - Body example: `{ "user_login": "admin", "password": "your_password" }`.
   - From the response, copy the JWT into the environment variable **token** (or set the collection variable **token**). All other requests use **Authorization: Bearer {{token}}** and **x-tenant-id: {{tenant_id}}**.

4. **Start the API**
   - Ensure the backend is running (e.g. `http://localhost:8080`). Change **base_url** in the environment if you use a different host/port.

## Variables (unified)

| Variable          | Description                                      | Example / note                          |
|-------------------|--------------------------------------------------|-----------------------------------------|
| `base_url`        | API base URL                                     | `http://localhost:8080`                 |
| `tenant_id`       | Tenant identifier (slug) for `x-tenant-id`       | `default` (must exist in DB)            |
| `token`           | JWT from **Auth → Login**                        | Set after login                          |
| `store_id`        | Store ID for store-scoped requests              | `1` (see sample data below)              |
| `cart_id`         | Cart ID (set after creating a cart)             | —                                       |
| `order_id`        | Order ID                                         | —                                       |
| `transaction_id`  | POS transaction ID                               | `1`                                     |
| `cashier_id`      | Cashier ID                                       | `1`                                     |
| `session_id`      | Cashier session ID                               | `1`                                     |

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

**Note:** If your DB has no tenants or users, create at least one tenant and user (or add them in the init scripts) so **Login** can succeed. Use that tenant slug as `tenant_id` and the user’s credentials in the login body.

## Collection structure (all routing modules)

Every module registered under `api := r.Group("/api")` with JWT + Tenant middleware is included:

- **Auth** – Login (no Bearer).
- **Dev** – Get Dev Token (`GET /dev/token`, dev only).
- **Tenants (public)** – Get by slug, List active.
- **Carts** – Full CRUD, items, totals, checkout, activities.
- **Cart Items** – Get, update, quantity, delete.
- **POS** – Categories, products, payments, store products/transactions, void.
- **POS Returns** – Process return.
- **POS Terminals** – List, create, get, update, delete, by store.
- **Cashier Sessions** – Open, get, close, summary.
- **Cashiers** – Full CRUD, limits, activate/deactivate, by store.
- **Orders** – Full CRUD, lines, status, payment status, fulfillments.
- **Order Lines** – Get, update, fulfillment, status, delete.
- **Order Fulfillments** – CRUD, shipment, pick-pack, items.
- **Stores** – CRUD, POS-enabled, warehouses, by parent, hierarchy.
- **Organizations** – CRUD, by code.
- **Tenants** – Create, list all, update, deactivate.
- **Users** – CRUD, roles, stores, details, search.
- **Roles** – CRUD, active, permissions.
- **Permissions** – CRUD, module/menu/submenu/role assign, user check.
- **Modules** – CRUD, navigation.
- **Menus** – CRUD, by module, by parent, toggle active.
- **Submenus** – CRUD, by menu, by parent, by code, toggle.
- **Navigation** – User navigation, roles with counts.
- **Images** – Upload (by module).
- **Brands** – Full CRUD, search, bulk, stats.
- **Price Lists** – CRUD, default, active, valid.
- **Tax Categories** – CRUD, active.
- **UOMs** – CRUD, by type, product conversions.
- **Product Prices** – CRUD, effective, comparison, search.
- **Product Variants** – CRUD, by product, by SKU, toggle active.
- **Product Barcodes** – CRUD, lookup, by product/variant, primary.
- **Inventory Stock** – CRUD, upsert, adjust, by store/product/location.
- **Storage Locations** – CRUD, by store, by parent, by type.
- **Restaurant** – Tables, menu categories, menu items, modifiers, orders.

For request/response shapes, see **Swagger** at `http://localhost:8080/swagger/index.html` when the API is running.

## Regenerating the collection

To regenerate the collection from the script (e.g. after adding routes):  
`node postman-collection/generate-collection.js`
