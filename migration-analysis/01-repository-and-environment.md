# SAP → NEMBUS migration audit: repository and environment

**Audit mode:** read-only investigation and reporting only  
**Audit date:** 2026-09-22  
**Repository:** `D:\nastecsol\Nembus`

## Scope and preservation

No application code, migration code, SQL schema, database data, configuration, or backup was changed. The SAP migration service was not run. The PostgreSQL backup was inspected but not restored. The only new files are the six artifacts in `migration-analysis/`.

No repository `AGENTS.md` file was present. The supplied audit brief was the governing instruction set.

## Repository preflight

| Item | Result |
|---|---|
| Current branch | `agent-v1` |
| HEAD | `c10b338948311646afaaa8181300b989a8caf7f7` |
| Remote | `origin https://github.com/NasTecSol/Nembus.git` |
| Pre-existing modified path | `apps/pos-client/frontend` |
| Pre-existing untracked path | `packages/core/db.zip` |
| New audit output | `migration-analysis/` only |
| Go workspace | `go.work` |
| TypeScript/build verification | Not run; Angular build was outside the read-only audit scope |

The pre-existing worktree paths were preserved exactly. The final status is reported in the final response and can be compared with this baseline.

## Repository structure and runtime entry points

| Area | Observed role |
|---|---|
| `apps/sap-agent` | SAP extraction agent, CLI/service entry points under `cmd/`, ETL engine under `internal/etl`, SQL/source access under `internal/source`, transport under `internal/transport` |
| `packages/sap` | SAP table SQL, source records, and SAP-to-domain mapping functions |
| `packages/core` | PostgreSQL target schema, HTTP handlers, migration use case, repositories, and target models |
| `apps/cloud-server` | Cloud/server application wiring and HTTP runtime |
| `apps/pos-client` | POS client; not executed or changed during this audit |
| `packages/core/db/schema` | PostgreSQL schemas, invoice totals trigger, and SAP staging tables |
| `db` / `knowledge` material | Repository support material was inventoried; no files were changed |

The SAP agent's default ETL domain list is in `apps/sap-agent/internal/etl/engine.go:20-41`. It includes stores, users, units of measure, UOM groups, categories, brands, products, barcodes, price lists, inventory, business partners, partner addresses, purchase orders, goods receipts, stock movements, sales orders, invoices, and incoming payments.

The CLI/service entry points are `apps/sap-agent/cmd/main.go` and `apps/sap-agent/cmd/agent/main.go`. Invoice extraction is orchestrated in `apps/sap-agent/internal/etl/engine.go:792-864`, using three-month windows and the configured invoice start date. The inspected configuration contained `invoice_start_date: "2023-01-01"`; credentials were not copied into this report.

## Database engines and runtime flow

| Stage | Engine/transport | Evidence |
|---|---|---|
| Source | Microsoft SQL Server, Windows integrated authentication | `sqlcmd -S localhost -E -C -d Qadsiya_Dev` succeeded |
| Source access | Go SQL queries in `packages/sap/schema/tables.go` and source helpers | OINV/INV1 and other SAP tables are queried directly |
| Batch transport | HTTP POST with gzip-compressed JSON | `apps/sap-agent/internal/transport/client.go` posts `/api/v1/migration/batch` |
| Cloud handler | Gzip decompression and JSON unmarshal | `packages/core/handler/sap_migration.go:25-67` |
| Target | PostgreSQL through pgx transaction | `packages/core/usecase/sap_migration.go` |
| Staging | PostgreSQL staging schema exists | `packages/core/db/schema/95_sap_staging.sql` |

The target use case opens one PostgreSQL transaction, records a row in `staging.sap_migration_batches`, and then upserts/inserts target entities. The implementation references `staging.sap_migration_batches`; the additional `staging.sap_stores`, `staging.sap_products`, and `staging.sap_inventory` tables are defined in the schema but were not observed as write targets in the migration use case.

## Read-only source inspection

`Qadsiya_Dev` was reachable and contained 2,503 `dbo` base tables. The source inspection used only metadata, `SELECT`, CTE, and aggregate queries. Representative source counts included:

| Domain | Source count |
|---|---:|
| Warehouses / bins | `OWHS` 2 / `OBIN` 0 |
| Users / salespersons | `OUSR` 32 / `OSLP` 1 |
| UOM / UOM groups | `OUOM` 37 / `OUGP` 345 / `UGP1` 1,118 |
| Items / categories / barcodes | `OITM` 16,955 / `OITB` 27 / `OBCD` 27,591 |
| Item warehouses | `OITW` 33,910 |
| Business partners | customers 14 / suppliers 387 / `CRD1` 1 |
| Sales orders | `ORDR` 1 / `RDR1` 1 |
| A/R invoices | `OINV` 969,665 / `INV1` 5,472,488 |
| Purchase orders | `OPOR` 11,742 / `POR1` 134,214 |
| Goods receipts | `OPDN` 8,468 / `PDN1` 92,205 |
| Inventory movements | `OINM` 5,875,327 |
| Incoming payments | `ORCT` 1,486 / `RCT2` 971,586 |
| A/R credit notes | `ORIN` 1,387 / `RIN1` 7,704 |

All invoice header currency values observed were SAR. The invoice source date range was 2021-09-08 through 2025-05-08.

## Backup inspection

The likely target backup was found at `D:\stg-backup`. It is a PostgreSQL custom-format archive beginning with the `PGDMP` signature, 611,826,724 bytes, gzip-compressed, with 1,642 TOC entries. It was dumped from PostgreSQL 18.1 by `pg_dump` 18.1. The installed inspection tool was PostgreSQL 18.4, which is the compatible major version family. No restore or database creation was attempted.

## Audit limitations

The target database was unavailable because the backup was intentionally not restored. Therefore target row counts, target-side negative discount values, target foreign-key outcomes, and post-load parity could not be verified. Those items are classified as `TARGET_DATA_REQUIRED` or `SUSPICIOUS_MAPPING` where appropriate rather than as confirmed target-output defects.
