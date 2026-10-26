# Nembus Database Schema & Core Synchronization Guide

This guide establishes the canonical database schema architecture, migration workflow, code generation pipeline, and synchronization procedures across the Nembus monorepo (`packages/core`, `apps/cloud-server`, `apps/pos-client`, and `apps/sap-agent`).

---

## 1. Architecture & Single Source of Truth

To ensure zero schema drift between Cloud ERP, POS desktop clients, and background integration agents, the Nembus codebase enforces a single source of truth:

### 1.1. Canonical Schema & Migrations (`packages/core`)
- **Canonical Schema**: `packages/core/db/schema/*.sql` and `packages/core/db/views_functions/*.sql`.
  All domain entities, tables, relationships, constraints, views, and functions are authored here.
- **Atlas Migrations**: `packages/core/db/migrations/`.
  Timestamped declarative SQL migrations managed and validated by Atlas CLI.
- **Integrity Checksums**: `packages/core/db/migrations/atlas.sum`.
  Cryptographic checksum manifest recording hash sums for all core migrations.

### 1.2. Downstream Targets
- **Cloud Server Migrations**: `apps/cloud-server/migrations/`.
  A direct downstream mirror of `packages/core/db/migrations/` (including `atlas.sum`).
- **POS Client Migrations**: `apps/pos-client/migrations/`.
  Downstream migrations formatted for Goose (`-- +goose Up` and `-- +goose StatementBegin`/`StatementEnd` for PL/pgSQL functions).

### 1.3. Client-Only Schema Extensions
- **Client Extensions File**: `apps/pos-client/migrations/99999999999999_pos_extensions.sql`.
  Contains client-local tables, hardware configs, and offline outbox tracking:
  - `local_printer_configs`: POS receipt and ticket printer hardware configuration.
  - `sync_queue`: Transactional outbox table for queuing offline sales and delta sync events.
  - `local_device_config`: Local device ID, store/terminal bindings, and ZATCA sync watermarks.
  - `trg_enqueue_sync_outbox()`: Outbox triggers that queue mutations for upstream replication.
  *Client-only tables must never be added to `packages/core`.*

### 1.4. Shared Repository Layer & SQLC
- **Queries**: SQL query definitions live in `packages/core/queries/*.sql` (and `apps/cloud-server/queries/*.sql`).
- **Repository Code**: `packages/core/repository/` and `apps/cloud-server/internal/repository/`.
  Strictly generated via `sqlc generate`. Never hand-edit repository Go files.

---

## 2. Downstream Synchronization Workflow

Whenever schema files in `packages/core/db/schema/`, core migrations, or SQL queries are modified, synchronize all downstream targets using the unified sync script or Makefile:

### Executing Monorepo Sync
From the repository root:
```powershell
powershell -ExecutionPolicy Bypass -File scripts/sync_migrations.ps1
```
Or via make:
```powershell
make sync-core
```

### What `sync_migrations.ps1` Does:
1. **Ensures Target Directories**: Creates `apps/cloud-server/migrations` and `apps/pos-client/migrations` if they do not exist.
2. **Normalizes Core Migrations**: Enforces Unix LF (`\n`) line endings and UTF-8 encoding without BOM across all SQL files in `packages/core/db/migrations/`.
3. **Mirrors to Cloud Server**: Copies all migration files from `packages/core/db/migrations/` directly to `apps/cloud-server/migrations/`.
4. **Transforms for POS Client (Goose)**:
   - Reads each migration file from `packages/core/db/migrations/`.
   - Normalizes line endings to Unix LF (`\n`).
   - Wraps any PL/pgSQL function declaration (`CREATE [OR REPLACE] FUNCTION ...`) with `-- +goose StatementBegin` and `-- +goose StatementEnd`.
   - Prepends `-- +goose Up` header if not present.
   - Writes to `apps/pos-client/migrations/` with UTF-8 encoding without BOM and LF line endings.
   - Normalizes client extensions (`99999999999999_pos_extensions.sql`) to LF without BOM while preserving all local tables and triggers.
5. **Recomputes Atlas Hash**:
   - Executes `atlas migrate hash --env local` in `packages/core`.
   - Normalizes `packages/core/db/migrations/atlas.sum` to Unix LF and mirrors it to `apps/cloud-server/migrations/atlas.sum`.
6. **Executes SQLC Code Generation**:
   - Runs `sqlc generate` in `packages/core/` to refresh `packages/core/repository/`.
   - Runs `sqlc generate` in `apps/cloud-server/` if `sqlc.yaml` exists.

---

## 3. Step-by-Step Workflow for Schema Changes

Follow these steps when adding a new table, modifying columns, adding indexes, or updating SQL queries:

### Step 1: Edit Canonical Schema or Queries
- Modify definitions in `packages/core/db/schema/*.sql` (or `packages/core/db/views_functions/*.sql`).
- Add or modify query definitions in `packages/core/queries/*.sql`.

### Step 2: Generate Migration Diff
From repo root, generate a versioned migration using Atlas:
```powershell
atlas migrate diff <migration_name> --env local
```
Or via make:
```powershell
make db-diff name=<migration_name>
```

### Step 3: Validate Migration Integrity
Validate the migration syntax and directory checksums:
```powershell
atlas migrate validate --dir "file://packages/core/db/migrations"
```
Or via make:
```powershell
make db-validate
```

If manual edits were made to a migration file, recalculate the directory hash immediately:
```powershell
atlas migrate hash --dir "file://packages/core/db/migrations"
```
Or via make:
```powershell
make db-hash
```

### Step 4: Run Monorepo Sync
Propagate changes across the monorepo:
```powershell
powershell -ExecutionPolicy Bypass -File scripts/sync_migrations.ps1
```
*(Or `make sync-core`)*

### Step 5: Verify Build & Tests
Confirm that all affected Go modules compile and pass test suites:
```powershell
go test -v ./packages/core/...
go test -v ./apps/cloud-server/...
go test -v ./apps/pos-client/...
go test -v ./apps/sap-agent/...
```

---

## 4. Strict Monorepo Rules & Integrity Mandates

1. **Single Source of Truth**:
   - The canonical database schema lives exclusively in `packages/core/db/schema/*.sql`.
   - **NEVER** edit files in `apps/cloud-server/migrations/` or `apps/pos-client/migrations/` directly for shared domain tables. They are downstream targets generated by `scripts/sync_migrations.ps1`.
   - Client-only tables (`local_printer_configs`, `sync_queue`, `local_device_config`) must reside solely in `apps/pos-client/migrations/99999999999999_pos_extensions.sql` and must never pollute `packages/core`.

2. **Migration Integrity & Checksums**:
   - **NEVER** hand-edit `packages/core/db/migrations/atlas.sum` or `apps/cloud-server/migrations/atlas.sum`.
   - Always let Atlas recompute checksums via `atlas migrate hash` or `scripts/sync_migrations.ps1`.

3. **Line Endings & Encoding**:
   - All SQL migration files MUST use Unix LF line endings (`\n`) and UTF-8 encoding without BOM.
   - CRLF (`\r\n`) line endings cause checksum mismatches in GitHub Actions CI (`db-validate.yml`).

4. **Generated Repository Code**:
   - Never hand-edit code in `packages/core/repository/` or `apps/cloud-server/internal/repository/`.
   - These are regenerated automatically by `sqlc generate`.
