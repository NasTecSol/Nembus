# Database Migrations & Schema Management Guide

This project uses [Atlas](https://atlasgo.io/) for declarative schema management and versioned migrations in a multi-tenant PostgreSQL architecture, integrated with [sqlc](https://sqlc.dev/) for type-safe Go code generation.

---

## 🏛️ Architecture & Single Source of Truth

```
Desired Schema State                 sqlc Type-Safe Go
(packages/core/db/schema/)          (packages/core/queries/)
          │                                   │
          ├───────────────────────────────────┤
          ▼                                   ▼
    Atlas Diff Engine                   sqlc generate
          ▼                                   ▼
Versioned Migration Files           packages/core/repository/
(packages/core/db/migrations/)
          ▼
PostgreSQL (Master & Tenants)
```

1. **Modular Desired Schema** (`packages/core/db/schema/`):
   * `00_extensions.sql` (PostgreSQL extensions)
   * `10_identity_rbac.sql` (Organizations, tenants, modules, menus, permissions, roles, UI settings)
   * `20_stores_terminals.sql` (Stores, storage locations, users, cashiers, terminals, sessions)
   * `30_catalog.sql` (Categories, brands, UOM, packaging templates, price lists, taxes, products, variants, barcodes, pricing, batches)
   * `40_inventory.sql` (Stock, movements, stock counts)
   * `50_purchasing_suppliers.sql` (Suppliers, customers, purchase orders, transfers, GRN)
   * `60_sales_pos.sql` (Sales orders, returns, POS transactions, payments, cart, checkout)
   * `70_restaurant.sql` (Tables, restaurant orders, menu items, recipes, waste logs)
   * `80_promotions_loyalty.sql` (Promotions, loyalty rules, points, rewards)
   * `85_zatca.sql` (ZATCA e-invoicing tables, device configs, document chain, watermarks)
   * `90_views_functions.sql` (Stored procedures, triggers, views, indexes)

2. **Immutable Migration History** (`packages/core/db/migrations/`):
   * Versioned SQL files (`YYYYMMDDHHMMSS_name.sql`) tracked with cryptographically signed `atlas.sum`.

---

## 🛠️ Prerequisites

1. **Atlas CLI**:
   * **Windows**: `Invoke-WebRequest -Uri https://release.ariga.io/atlas/atlas-windows-amd64-latest.exe -OutFile $env:USERPROFILE\go\bin\atlas.exe`
   * **macOS / Linux**: `curl -sSf https://atlasgo.io/install.sh | sh`
   * Verify: `atlas version`

2. **sqlc**:
   * `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
   * Verify: `sqlc version`

3. **Environment Setup**:
   Ensure `MASTER_DB_URL` is configured in `apps/cloud-server/.env.dev` or `.env`:
   ```env
   MASTER_DB_URL=postgres://postgres:postgres@localhost:5432/masterDB?sslmode=disable
   ```

---

## 🚀 Developer Workflow for Schema Changes

When you need to modify or add database tables, columns, or indexes:

### Step 1: Update the Modular Schema
Edit the appropriate domain SQL file in `packages/core/db/schema/` (e.g. `30_catalog.sql`).

### Step 2: Generate Migration Diff
From the monorepo root:
```bash
make db-diff name=add_product_metadata
```
Atlas automatically calculates the diff between your desired `db/schema/` and the latest migration state, writing a new versioned file to `packages/core/db/migrations/` and updating `atlas.sum`.

### Step 3: Review and Apply Locally
Review the generated SQL file in `packages/core/db/migrations/`, then apply:
```bash
make db-migrate
```
This runs migrations across both `masterDB` and all active tenant databases.

### Step 4: Update SQL Queries & Regenerate sqlc
Add or edit query definitions in `packages/core/queries/`, then run:
```bash
make sqlc
```
This regenerates `packages/core/repository/` with type-safe Go structs and query methods.

### Step 5: Test & Verify
```bash
make verify
```
Runs migration checksum validation, `sqlc generate`, and unit/integration test suites.

---

## 🏢 Multi-Tenant Database Migrations

Tenant migration is managed by `apps/cloud-server/cmd/migrate-tenants/main.go`:

```bash
# Migrate Master DB and all active Tenant DBs
go run cmd/migrate-tenants/main.go

# Migrate only active Tenant DBs
go run cmd/migrate-tenants/main.go -master=false

# Inspect migration status across all databases
go run cmd/migrate-tenants/main.go -status=true
```

### Baseline for Existing (Legacy) Databases

`migrate-tenants` **auto-detects the baseline per database** — no flags needed:

| Database state | Baseline used |
|---|---|
| Fresh / empty | none — all migrations applied from scratch |
| Has `atlas_schema_revisions` table | none — only pending migrations applied |
| Has schema but **no** revisions table (legacy / restored) | first migration version (currently `20260813124500`) — the existing schema is treated as already applied |

The `-baseline` flag exists only to force a fixed baseline for every database,
e.g. `go run cmd/migrate-tenants/main.go -baseline 20260813124500`. Migrations
with version <= baseline are treated as already applied. ⚠️ Never pass a
baseline for fresh/empty databases — Atlas would skip the initial schema
migration. Also, the baseline must be an existing migration version in
`packages/core/db/migrations/`, otherwise Atlas errors with
`baseline version "..." not found`.

### Tenant Connection Strings (Docker Hostnames)

Tenant `db_conn_str` values stored in the master `tenants` table often use the
Docker **service name** as the host (e.g. `host=postgres user=... dbname=qitaf`),
which only resolves inside the compose network. Because `migrate-tenants` runs on
the host, it automatically:

1. Converts any pgx-accepted connection string (keyword DSN **or** URL) into a
   `postgres://` URL, which the Atlas CLI requires (keyword DSNs are rejected
   with `missing driver`).
2. Rewrites the configured host via the `-host-override from=to` flag
   (default `postgres=localhost`) so `host=postgres` becomes `localhost` — valid
   because the Postgres container's port is published to the host. Disable with
   `-host-override ""`.

Notes:
- The `atlas_schema_revisions` lookup searches **all** schemas (not only
  `public`) — revisions tables created in a `$user` schema (search_path quirk)
  are still detected, so such databases are not wrongly baselined.
- Tenants whose database **does not exist** on the server (SQLSTATE `3D000`)
  are skipped with a warning and counted separately in the summary — they do
  not fail the run. Create the database or set `is_active = false` for the
  tenant to resolve.

---

## ⚡ Summary of Makefile Commands

| Command | Action |
|---------|--------|
| `make db-diff name=<name>` | Diff `db/schema/` vs migrations and create new migration file |
| `make db-migrate` | Run Atlas migrations on master and all tenant databases |
| `make db-status` | Check Atlas migration status |
| `make db-hash` | Recalculate and verify `atlas.sum` |
| `make db-lint` | Lint migrations for safety and backward compatibility |
| `make sqlc` | Regenerate Go repository code with sqlc |
| `make test` | Run Go test suites |
| `make verify` | Run end-to-end verification (hash, sqlc, tests) |
