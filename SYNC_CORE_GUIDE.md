# Core Schema & SQL Sync Guide

This guide explains how database migrations and SQL queries are managed across the Nembus Monorepo, ensuring zero schema drift between the Cloud ERP server and POS desktop client applications.

---

## Architecture & Source of Truth

To eliminate code duplication and database schema mismatches:

1. **`packages/core` is the Single Source of Truth**:
   - `packages/core/migrations/000001_base_schema.sql`: Contains the canonical ERP database schema.
   - `packages/core/queries/`: Contains all 60 `.sql` query files used by `sqlc` to generate the repository layer (`packages/core/repository/`).

2. **`apps/pos-client` Schema Extension Strategy**:
   - `apps/pos-client/migrations/000001_base_schema.sql`: Exact copy of the base schema synced from `packages/core`.
   - `apps/pos-client/migrations/000002_pos_extensions.sql`: Client-only schema extensions (e.g. `local_printer_configs`, `sync_queue`, `local_device_config`).
   - Goose migration runner in `apps/pos-client/app.go` executes `000001` and `000002` sequentially on startup.

---

## How to Sync Base Schema & Queries

Whenever base database tables or SQL query files in `packages/core` are updated, run the core sync tool.

### Method 1: Via Makefile (Recommended)
From the monorepo root:
```powershell
make sync-core
```

### Method 2: Via PowerShell Script Directly
From `apps/pos-client/`:
```powershell
cd apps/pos-client
powershell -ExecutionPolicy Bypass -File scripts/sync_core.ps1
```

### What `sync-core` Does:
1. Overwrites `apps/pos-client/migrations/000001_base_schema.sql` with `packages/core/migrations/000001_base_schema.sql`.
2. Copies all updated query files from `packages/core/queries/` into `apps/pos-client/queries/`.
3. Runs `sqlc generate` inside `packages/core/` to refresh the shared `packages/core/repository/` Go code.

---

## Step-by-Step Workflow for Adding a New Schema Feature

When adding a new table or modifying a query for shared domain models:

1. **Modify Base Schema in Core**:
   Update `packages/core/migrations/000001_base_schema.sql`.
2. **Modify Query Files in Core**:
   Add or update queries in `packages/core/queries/*.sql`.
3. **Regenerate & Sync**:
   ```powershell
   make generate-core
   make sync-core
   ```
4. **Compile & Test**:
   ```powershell
   make build-server
   make build-client
   ```
5. **Commit Changes**:
   Commit the updated files in `packages/core` and `apps/pos-client`.
