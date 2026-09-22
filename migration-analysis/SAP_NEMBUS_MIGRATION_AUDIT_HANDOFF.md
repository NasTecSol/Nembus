# SAP → NEMBUS Migration Audit — Authoritative Handoff

**Document type:** Authoritative, self-contained project handoff and revalidation manual  
**Audit execution date:** 2026-09-22  
**Repository path:** `D:\nastecsol\Nembus`  
**Current branch:** `agent-v1`  
**Exact commit (HEAD):** `c10b338948311646afaaa8181300b989a8caf7f7`  
**Audit mode:** Investigation, authorized isolated target restoration, and read-only validation only  
**Remediation ownership:** Migration developer (implementation) and business/finance owners (policy decisions)  

---

## 1. Continuity directive

This document is the **sole authoritative source of truth** for the completed SAP → NEMBUS migration audit. A new chat, auditor, or engineer must read this file first and treat all facts, counts, code references, and findings recorded here as definitive.

### Core operating rules for subsequent chats:
1. **Treat this handoff as authoritative:** Supporting reports in `migration-analysis/` provide supplementary evidence, logs, and breakdowns, but they must not override the verified conclusions recorded in this document without fresh, reproducible proof.
2. **Do not ask the user to repeat information:** All requirements, environmental facts, database credentials status, backup locations, architecture paths, findings, affected counts, and representative records have already been discovered, verified, and recorded here.
3. **Investigation and validation only:** This audit project is strictly read-only analysis and post-fix validation. **Do not implement fixes or modify application/migration code** in this worktree. Code modifications are exclusively the responsibility of the migration developer.
4. **Developer owns implementation; audit owns revalidation:** Once the migration developer delivers code changes, the audit responsibility is to inspect the developer's commit/diff, verify that only the intended mappings changed, and execute the read-only regression workflow documented in Section 16 against a newly generated, isolated target database.
5. **Strict read-only database constraint:** Keep all database interactions against SQL Server `Qadsiya_Dev` and PostgreSQL `nembus_migration_audit_20260922` strictly read-only (SELECT queries only). No DDL, DML, temporary table creation, statistics updates, or configuration modifications are permitted.
6. **Do not rerun or restore `D:\stg-backup` again:** The audit database `nembus_migration_audit_20260922` was already safely restored, verified, and configured with `default_transaction_read_only = on` on the local PostgreSQL 18 service. It serves as an immutable evidence baseline.
7. **Preserve repository and worktree state:** Do not commit, stage, push, pull, merge, rebase, reset, or switch branches. Do not alter the pre-existing uncommitted worktree paths (`apps/pos-client/frontend` and `packages/core/db.zip`).

---

## 2. Executive summary

### Objective and context
A comprehensive, end-to-end migration audit was conducted to validate data parity, transformation accuracy, business semantics, and architectural integrity between an existing SAP Business One source system and the modern NEMBUS PostgreSQL platform. The audit verified the extraction, mapping, transport, and ingestion pipeline against real source data and a fully restored staging backup.

### Systems audited
- **Source system:** SAP Business One running on Microsoft SQL Server (`localhost`), database `Qadsiya_Dev` (2,503 base tables).
- **Target system:** NEMBUS platform running on PostgreSQL 18 (`127.0.0.1:5432`), audit database `nembus_migration_audit_20260922` (114 base tables across schemas `atlas_schema_revisions`, `public`, and `staging`).

### Prominent audit numbers (verified against live source and target)
The following key figures represent the verified baseline of the audit:

| Metric / Domain | SAP Source (`Qadsiya_Dev`) | NEMBUS Target (`nembus_migration_audit_20260922`) | Parity Status / Root Cause |
|---|---:|---:|---|
| **Invoice headers** | **969,665** | **969,665** | **Exact match (1:1)** by `metadata.sap_doc_entry`. |
| **Invoice lines (raw)** | **5,472,488** | **5,472,485** | **Target short by 3 lines.** (Identities proven: orphan lines). |
| **Eligible invoice lines** | **5,472,485** | **5,472,485** | **Exact match** for inner-join eligible population. |
| **Nonzero line discount count** | **2,192,345** | **0** | **100% loss of line discount data** (`DiscPrcnt` omitted in code). |
| **Audited invoice currency** | **SAR (969,665)** | **USD (969,665)** | **100% currency mismatch** (defaults to USD; SAR master exists). |
| **Header discount sum** | **-725.45 SAR** | **-725.45** | **Exact match** (`OINV.DiscSum` correctly mapped and verified). |
| **Invoice total reconciliation** | Reconstructed | Reconstructed | **0.00 difference** across all 969,665 target invoices. |
| **Payment allocations** | **961,745** | **0** | **100% loss of payment detail** (`invoice_payments` is empty). |
| **Payment documents / Invoices** | **1,329 / 955,788** | **0 / 0** | Totaling **64,438,195.83 SAR**; domain was never delivered. |
| **Inventory snapshot keys** | **7,203** | **158,466** | **151,263 excess rows** (each pair repeated exactly 22 times). |
| **Stock movements** | **2,163,651** | **2,710,151** | **546,500 excess rows** (non-idempotent replay duplicates). |
| **PO-linked GRNs / Lines** | **15 / 248** | **0 / 0 linked** | **100% PO link failure** (DocNum vs BaseEntry mismatch). |
| **Canceled documents (Y/C)** | **5,344** | **0 cancelled** | Cancellation flags ignored; all derived as paid/sent. |
| **Service documents (DocType S)**| **31** | **31 standard** | Service lines treated as products with NULL `product_id`. |

### Finding classifications summary
Reconciling Sweep 1 (provisional code audit) and Sweep 2 (restored target validation):
- **Confirmed code & data defects:** **10** (Line discount omission, currency default, payment detail omission, inventory non-idempotency, stock-movement non-idempotency, GRN-to-PO foreign key link failure, partner address omission, line UOM omission, cancellation semantics loss, service document misclassification).
- **Confirmed data mismatches:** **10** (Direct target data divergence resulting from the code defects above).
- **Confirmed source data conditions:** **2** (Negative signed values in header `DiscSum` and line `DiscPrcnt`; absence of brands due to missing `OMRG` in source).
- **Suspicious mappings needing telemetry proof:** **1** (Arbitrary supplier/store fallback in PO/GRN ingestion; code is unsafe, though loaded key mismatch is currently 0).
- **Documentation & provenance gaps:** **3** (Historical run configuration discrepancy with 315,715 pre-2023 invoices; 366 failed invoice batches with empty error strings; undocumented tax master/reference scope).
- **Out of scope / intentionally excluded:** **2** (A/R credit notes & returns `ORIN`/`RIN1` [1,387/7,704]; promotions).
- **Confirmed correct mappings:** **11** (Header `DiscSum`, invoice counts, line uniqueness, mathematical total reconstructions, product catalog, categories, customers, suppliers, stores, positive price-list slice, barcode deduplication).
- **Blocked items:** **0** (All previous authentication and data blockers resolved).

### Highest impact findings
1. **Invoice line discounts completely omitted:** `INV1.DiscPrcnt` and `PriceBefDi` are omitted from extraction queries and target schemas. Over 2.19 million lines lost commercial discount/surcharge auditability.
2. **Systemic currency mismatch:** Every invoice is mislabeled as `USD` instead of `SAR`, misrepresenting all financial, tax, and ZATCA reporting.
3. **Total absence of payment allocations:** 961,745 payment allocations across 955,788 invoices totaling 64.44M SAR were completely omitted from the target database.
4. **Massive inventory & stock-movement duplication:** 151,263 duplicate inventory rows (22x multiplication) and 546,500 duplicate stock movements occurred due to a lack of replay-safe upsert constraints.
5. **Broken PO-to-GRN procurement links:** Goods receipt notes construct PO numbers using SAP `BaseEntry` (internal `DocEntry`), while purchase orders use `DocNum`. All target foreign keys are consequently NULL.

### Recommended next action
The migration developer must implement targeted code corrections in `packages/sap` and `packages/core`, guided by the Developer Action Register (Section 15) and business policy decisions on line discount arithmetic, currency handling, and cancellation status.

---

## 3. Scope and ownership

### Roles and responsibilities
- **Audit team (pair programmer / assistant):** Responsible for forensic discovery, proving defects with exact SQL and code evidence, establishing parity baselines, maintaining safety constraints, and running read-only regression verifications against newly delivered targets.
- **Migration developer:** Responsible for modifying source extraction queries, mapping structures, canonical models, and target insertion/upsert logic, as well as providing clean migration commits.
- **Business / Finance stakeholders:** Responsible for formalizing business policies regarding signed surcharge interpretations, currency fallback rules, cancellation status handling, and historical date horizons.

### Strict non-modification boundaries
During this audit:
- **Zero source data modified:** SQL Server `Qadsiya_Dev` was accessed strictly via read-only queries.
- **Zero target data modified:** No rows in `nembus_migration_audit_20260922` were updated, inserted, or deleted to "simulate" fixes.
- **Zero application/migration code modified:** No files in `apps/` or `packages/` were edited.
- **Zero git mutations:** No staging, committing, pushing, or branch switching took place.

### Validated vs unvalidated domains

| Domain | Source Tables | Target Tables | Audit Validation Status | Notes |
|---|---|---|---|---|
| **Stores / Warehouses** | `OWHS` | `public.stores` | **Fully validated (Correct)** | Count 2 = 2. |
| **Storage Locations** | `OBIN` | `public.storage_locations` | **Fully validated (Source condition)** | Count 0 = 0. |
| **Units of Measure** | `OUOM`, `OUGP`, `UGP1`| `public.units_of_measure` | **Fully validated (Correct)** | 37 source keys + 4 defaults = 41 target. |
| **Categories** | `OITB` | `public.product_categories`| **Fully validated (Correct)** | Count 27 = 27. |
| **Brands** | `OMRG` | `public.brands` | **Fully validated (Source condition)** | Source table absent; target 0. |
| **Products** | `OITM` | `public.products` | **Fully validated (Correct)** | 16,955 distinct SAP items match. |
| **Barcodes** | `OBCD` | `public.product_barcodes` | **Fully validated (Correct)** | 27,591 raw normalize to 27,445 distinct. |
| **Price Lists & Prices** | `OPLN`, `ITM1` | `public.price_lists`, `product_prices` | **Fully validated (Correct)** | 10 lists; 16,686 positive prices match. |
| **Business Partners** | `OCRD` | `public.customers`, `suppliers` | **Fully validated (Correct)** | 14 customers, 387 suppliers match. |
| **Sales Orders (v2)** | `ORDR`, `RDR1` | `public.sales_orders_v2`, `_lines_v2` | **Fully validated (Correct)** | 1 order / 1 line matches. |
| **Invoices (Headers)** | `OINV` | `public.invoices` | **Fully validated (Defects found)** | 969,665 rows; currency & status defects. |
| **Invoice Lines** | `INV1` | `public.invoice_lines` | **Fully validated (Defects found)** | 5,472,485 rows; discount & UOM omitted. |
| **Invoice Payments** | `ORCT`, `RCT2` | `public.invoice_payments` | **Fully validated (Defect found)** | 0 target rows vs 961,745 source allocations. |
| **Inventory Stock** | `OITW` | `public.inventory_stock` | **Fully validated (Defect found)** | 151,263 duplicate excess rows. |
| **Stock Movements** | `OINM` | `public.stock_movements` | **Fully validated (Defect found)** | 546,500 duplicate excess rows. |
| **Purchase Orders** | `OPOR`, `POR1` | `public.purchase_orders`, `_lines` | **Partially validated** | Loaded 2024-2025 slice matches (99/1,755). |
| **Goods Receipts** | `OPDN`, `PDN1` | `public.goods_receipt_notes`, `_items`| **Partially validated (Defect found)** | Loaded slice matches; PO links 100% broken. |
| **Partner Addresses** | `CRD1` | `public.partner_addresses` | **Fully validated (Defect found)** | 1 source row vs 0 target rows. |
| **A/R Credit Notes** | `ORIN`, `RIN1` | `public.sales_returns`, `_lines` | **Unvalidated (Out of scope)** | No migration path implemented. |
| **Promotions** | None | `public.promotions` | **Unvalidated (Out of scope)** | No SAP source path exists. |
| **Users / Cashiers** | `OUSR`, `OSLP` | `public.users`, `cashiers` | **Partially validated (Doc gap)** | Target baseline users prevent raw count parity. |

---

## 4. Repository and worktree state

### Repository metadata
- **Repository directory:** `D:\nastecsol\Nembus`
- **Active branch:** `agent-v1`
- **Exact commit hash (HEAD):** `c10b338948311646afaaa8181300b989a8caf7f7`
- **Configured remotes:**
  - `origin` -> `https://github.com/NasTecSol/Nembus.git` (fetch & push)
- **Go workspace:** Configured via `go.work` in root.

### Initial and final worktree status
Running `git status --short` confirms the worktree was kept pristine with no unapproved changes:
```text
 M apps/pos-client/frontend
?? migration-analysis/
?? packages/core/db.zip
```
- `apps/pos-client/frontend`: Pre-existing modification present prior to the audit; untouched.
- `packages/core/db.zip`: Pre-existing untracked zip archive; untouched.
- `migration-analysis/`: Audit output directory containing documentation, inventories, and SQL checks.

### Key architecture files and entry points
- **Orchestration & ETL domain runner:** `apps/sap-agent/internal/etl/engine.go` (Domain lists at lines 20–41; invoice date-window extraction at lines 792–864).
- **SAP source SQL extraction queries:** `packages/sap/schema/tables.go`.
- **SAP source models & canonical transformations:** `packages/sap/mappings/models.go`.
- **HTTP batch transport client:** `apps/sap-agent/internal/transport/client.go` (Sends gzip-compressed JSON payload to `/api/v1/migration/batch`).
- **Target migration HTTP handler:** `packages/core/handler/sap_migration.go` (Decompresses gzip, dispatches to usecase).
- **Target transactional ingestion usecase:** `packages/core/usecase/sap_migration.go` (Contains domain-by-domain upsert and insert transactions).
- **Target schema definitions:**
  - `packages/core/db/schema/85_zatca.sql` (Invoices, lines, payments, sales returns).
  - `packages/core/db/schema/90_views_functions.sql` (Reconstructive total triggers & payment balance functions).
  - `packages/core/db/schema/95_sap_staging.sql` (Migration batch audit tracking).

### Exact code locations for confirmed defects

| Defect / Omission | File & Line Reference | Function / Symbol | Code Excerpt / Search Phrase |
|---|---|---|---|
| **Line discount query omission** | `packages/sap/schema/tables.go:425-440` | `const QueryInvoiceLines` | Omits `DiscPrcnt`, `PriceBefDi`, `GTotal`. Only selects `Quantity, Price, LineTotal, VatSum, WhsCode, unitMsr`. |
| **Line discount model omission** | `packages/sap/mappings/models.go:786-827` | `type SAPInvoiceLine`, `CanonicalInvoiceLine` | Struct lacks any discount field; maps `LineSubtotal` directly from `LineTotal`. |
| **Line discount target insert omission** | `packages/core/usecase/sap_migration.go:768-787` | `DomainInvoices` line insert | `INSERT INTO invoice_lines` omits `discount_amount`, `uom_id`, `tax_category_id`, `tax_rate`; hardcodes `item_type = 'product'`. |
| **Invoice currency query omission** | `packages/sap/schema/tables.go:405-423` | `const QueryInvoicesHeader` | Omits `DocCur`, `CANCELED`, `DocType`. Only selects `DocEntry, DocNum, DocDate, DocDueDate, CardCode, CardName, DocTotal, PaidToDate, VatSum, DiscSum, DocStatus, SlpCode, Comments`. |
| **Invoice currency insert omission** | `packages/core/usecase/sap_migration.go:747-760` | `DomainInvoices` header insert | `INSERT INTO invoices` omits `currency_code`, allowing database schema default `'USD'` to apply. |
| **Cancellation & status derivation gap** | `packages/sap/mappings/models.go:846-895` | `SAPInvoice.ToCanonical` | `invoice_status` derived solely from `PaidToDate` vs `DocTotal`; ignores source `CANCELED` and `DocStatus`. |
| **Service invoice type hardcoding** | `packages/sap/mappings/models.go:882-883` | `SAPInvoice.ToCanonical` | Hardcodes `InvoiceType: "standard"`; does not inspect `DocType == 'S'`. |
| **Inventory replay duplication** | `packages/core/usecase/sap_migration.go:412-429` | `DomainInventory` insert | `INSERT INTO inventory_stock (...) SELECT ... LIMIT 1;` lacks `ON CONFLICT` or pre-check. |
| **Stock-movement replay duplication** | `packages/core/usecase/sap_migration.go:660-682` | `DomainStockMovements` insert | `INSERT INTO stock_movements (...) SELECT ...` lacks conflict target on `sap_trans_num`. |
| **GRN-to-PO link mismatch** | `packages/sap/mappings/models.go:1153, 1251-1276` | `SAPPurchaseOrder`, `SAPGoodsReceipt` | PO builds `PO-%d` using `po.DocNum`; GRN builds `PO-%d` using `l.BaseEntry` (which is SAP `DocEntry`). |
| **Arbitrary supplier/store fallback** | `packages/core/usecase/sap_migration.go:549-552, 616-619` | `DomainPurchaseOrders`, `DomainGoodsReceipts` | `COALESCE((SELECT id ... WHERE code = $3), (SELECT id ... LIMIT 1))` silently falls back to arbitrary entity. |

---

## 5. Environment and database inventory

### Microsoft SQL Server (Source)
- **Engine:** Microsoft SQL Server (x64), localhost instance.
- **Authentication mode:** Windows Integrated Authentication (`-E`), strictly credential-free in configuration and reports.
- **Source database:** `Qadsiya_Dev`.
- **Operational state:** Online, verified reachable, queried read-only.
- **Catalog size:** 2,503 base tables under schema `dbo`.
- **Audited invoice volume:** 969,665 headers (`OINV`), 5,472,488 raw lines (`INV1`).
- **Relevant SAP Business One tables:**
  - Administrative & Master: `OWHS` (Warehouses), `OBIN` (Bins), `OUSR` (Users), `OSLP` (Salespersons/Cashiers), `OUOM` (UOM), `OUGP`/`UGP1` (UOM Groups & Conversions), `OITB` (Categories), `OMRG` (Brands - absent), `OITM` (Items), `OBCD` (Barcodes), `OPLN`/`ITM1` (Price Lists & Prices), `OITW` (Item Stock Balances), `OCRD` (Business Partners), `CRD1` (Partner Addresses).
  - Transactions: `ORDR`/`RDR1` (Sales Orders), `OINV`/`INV1` (A/R Invoices), `ORCT`/`RCT2` (Incoming Payments), `OPOR`/`POR1` (Purchase Orders), `OPDN`/`PDN1` (Goods Receipt POs), `OINM` (Warehouse Journal / Stock Movements), `ORIN`/`RIN1` (Credit Notes / Returns).
- **Safety enforcement:** Only bounded `SELECT`, `COUNT`, `MIN/MAX`, and aggregation queries were executed.

### PostgreSQL (Target)
- **Client & server tools:** PostgreSQL 18.4 client binaries (`C:\Program Files\PostgreSQL\18\bin`), server major version 18.
- **Windows service:** `postgresql-x64-18` (Status: Running, Startup: Automatic).
- **Audit database name:** `nembus_migration_audit_20260922`.
- **Endpoint:** `127.0.0.1:5432` (Approved local authentication succeeded; no credentials disclosed).
- **Proof of database existence & isolation:** Restored from `D:\stg-backup` into this newly created database on 2026-09-22. Existing databases (`nembus_e2e_master`, `nembus_e2e_tenant`, `postgres`, `rms_db`, `template0`, `template1`) were completely untouched.
- **Proof of read-only enforcement:**
  The audit database was locked with:
  ```sql
  ALTER DATABASE nembus_migration_audit_20260922 SET default_transaction_read_only = on;
  ```
  Fresh connection verification query:
  ```sql
  SELECT current_setting('transaction_read_only'), current_setting('default_transaction_read_only');
  -- Returns: 'on', 'on'
  ```
- **Target schema inventory:**
  - `atlas_schema_revisions`: 1 base table.
  - `public`: 109 base tables, 12 views, 94 sequences, 38 public functions.
  - `staging`: 4 base tables (`sap_migration_batches`, `sap_stores`, `sap_products`, `sap_inventory`), 4 sequences, 5 indexes.
  - Total: 114 base tables.
- **Warning to future chats:** Always verify `SELECT current_database();` equals `nembus_migration_audit_20260922` before executing any SQL queries.

---

## 6. Backup provenance and restore history

### Backup files audited
- **SAP source backup:** `OsturatAlebtekar_backup_2025_01_01_000002_6419114` (Restored into `Qadsiya_Dev` prior to this audit).
- **PostgreSQL staging backup:** `D:\stg-backup` (File size: `611,826,724` bytes; last modified 2026-09-21 12:55:22).
  - Format: PostgreSQL custom archive (`PGDMP` header, gzip compressed).
  - Dump source: PostgreSQL 18.1 via `pg_dump`.
  - TOC entries: 1,642 entries.

### Chronological restore sequence and safety stop
1. **Initial discovery & safe stop (Sweep 1):**
   `pg_restore --list D:\stg-backup` was used to inspect archive headers safely. A subsequent connection probe stopped safely with:
   `fe_sendauth: no password supplied`.
   The agent correctly terminated the operation without attempting brute-force logins, creating databases, or overwriting files.
2. **Authorized preflight (Sweep 2):**
   User provided two authorized local password candidates. Exactly two attempts were made using process-scoped environment variables. One succeeded; one failed. The secret was immediately purged from process memory.
   Preflight verified that `nembus_migration_audit_20260922` did not exist, had zero active sessions, and adequate disk space existed (24.7 GB free on `C:`, 29.9 GB free on `D:`).
3. **Creation & restore execution:**
   The new database was created cleanly from `template0` with UTF-8 encoding. The restore command was executed:
   ```cmd
   pg_restore.exe -h 127.0.0.1 -p 5432 -U postgres -w ^
     --dbname=nembus_migration_audit_20260922 ^
     --exit-on-error --no-owner --no-privileges --verbose D:\stg-backup
   ```
   *Flags explained:* `--clean` and `--create` were strictly omitted to prevent any accidental dropping of objects.
4. **Restore log verification:**
   Execution logged to `migration-analysis/06-restore.log` (121,656 bytes, timestamp `2026-09-22 12:06:41`).
   - Recorded SHA-256: `2A7F49E43F48A16425A7724AE7A83A0D10DC087FB4EE975F0B558292BF615633`.
   - Result: Successful completion through final foreign key creation. Zero error, failure, or warning tokens.
5. **Read-only lockdown:**
   `default_transaction_read_only = on` was applied immediately. The existing PostgreSQL service remained running.

---

## 7. Migration architecture

### End-to-end data flow

```text
+------------------------------------+
|  SAP SQL Server (Qadsiya_Dev)      |
|  Tables: OINV, INV1, OITM, etc.   |
+-----------------+------------------+
                  |
                  | 1. SQL extraction queries (packages/sap/schema/tables.go)
                  v
+------------------------------------+
|  sap-agent ETL Extractor           |
|  (apps/sap-agent/internal/etl)     |
+-----------------+------------------+
                  |
                  | 2. SAP -> Canonical Go struct transformation (packages/sap/mappings/models.go)
                  v
+------------------------------------+
|  HTTP Migration Client             |
|  (apps/sap-agent/internal/transport|
+-----------------+------------------+
                  |
                  | 3. HTTP POST gzip JSON (/api/v1/migration/batch)
                  v
+------------------------------------+
|  NEMBUS Cloud Server Handler       |
|  (packages/core/handler)           |
+-----------------+------------------+
                  |
                  | 4. Decompress, unmarshal, start pgx transaction
                  v
+------------------------------------+
|  Migration Ingestion Usecase       |
|  (packages/core/usecase)           |
+--------+-------------------+-------+
         |                   |
         | 5a. Staging log   | 5b. Upsert / Insert business tables
         v                   v
+-----------------+ +-----------------------------------------------+
| staging.        | | public.invoices, public.invoice_lines,        |
| sap_migration_  | | public.products, public.inventory_stock, etc. |
| batches         | | (PostgreSQL: nembus_migration_audit_20260922) |
+-----------------+ +-----------------------------------------------+
```

### Complete entity-by-entity mapping inventory (All implemented domains)

| Domain | SAP Source Table & Key | Source Fields Extracted | Go Code Reference | Transformation Logic | Target Table & Key | Status | Confirmed Issues |
|---|---|---|---|---|---|---|---|
| **Stores** | `OWHS` (`WhsCode`) | `WhsCode, WhsName` | `tables.go:35-48`, `models.go:59-88` | Maps code/name, generates target UUID | `public.stores` (`id`) | `CONFIRMED_CORRECT` | None. Exact count match (2). |
| **Storage Bins** | `OBIN` (`AbsEntry`) | `AbsEntry, BinCode, WhsCode` | `models.go:108-132` | Maps bins to store IDs | `public.storage_locations` (`id`) | `CONFIRMED_SOURCE_DATA_CONDITION` | Source has 0 rows; target 0. |
| **Users** | `OUSR` (`USERID`) | `USERID, USER_CODE, U_NAME` | `models.go:158-186` | Maps user credentials & names | `public.users` (`id`) | `DOCUMENTATION_GAP` | Target has baseline users; count parity unasserted. |
| **Cashiers** | `OSLP` (`SlpCode`) | `SlpCode, SlpName` | `models.go:214-242` | Maps salesperson reference | `public.cashiers` (`id`) | `CONFIRMED_CORRECT` | Count match (1 = 1). |
| **UOM Master** | `OUOM` (`UomEntry`) | `UomEntry, UomCode, UomName` | `tables.go:72-86`, `models.go:264-280` | Maps code/name, retains source entry in metadata | `public.units_of_measure` (`id`)| `CONFIRMED_CORRECT` | 37 source keys + 4 defaults = 41 target rows. |
| **UOM Groups** | `OUGP`/`UGP1` (`UgpEntry+UomEntry`) | `UgpEntry, UgpCode, UomEntry, BaseQty` | `models.go:421-529` | Maps conversion ratios | `public.uom_groups`, `_lines` | `CONFIRMED_CORRECT` | Conversion structures present in target. |
| **Categories** | `OITB` (`ItmsGrpCod`) | `ItmsGrpCod, ItmsGrpNam` | `models.go:333-344` | Maps item group to category | `public.product_categories` (`id`) | `CONFIRMED_CORRECT` | Exact count match (27 = 27). |
| **Brands** | `OMRG` (`FirmCode`) | `FirmCode, FirmName` | `models.go:358-367` | Catches query error; returns empty | `public.brands` (`id`) | `CONFIRMED_SOURCE_DATA_CONDITION` | `OMRG` missing in source; target brands = 0. |
| **Products** | `OITM` (`ItemCode`) | `ItemCode, ItemName, ItmsGrpCod, VatGourpSa` | `tables.go:156-197`, `models.go:421-529`| Resolves category/tax/UOM; brand NULL | `public.products` (`id`) | `CONFIRMED_CORRECT` | 16,955 distinct items match. Brand NULL is source-conditioned. |
| **Barcodes** | `OBCD` (`ItemCode+BcdCode`) | `ItemCode, BcdCode, UomEntry` | `models.go:539-568` | Resolves product; collapses duplicates | `public.product_barcodes` (`id`)| `CONFIRMED_CORRECT` | 27,591 source rows normalize to 27,445 target. |
| **Price Lists** | `OPLN`/`ITM1` (`ListNum+ItemCode`) | `ListNum, ItemCode, Price, Currency` | `tables.go:224-253`, `models.go:925-969`| Maps positive prices for price lists | `public.product_prices` (`id`) | `CONFIRMED_CORRECT` | 10 lists; 16,686 positive prices match. |
| **Inventory Stock**| `OITW` (`ItemCode+WhsCode`) | `OnHand, IsCommited, OnOrder` | `models.go:596-613`, `usecase.go:412-429`| Plain INSERT without unique conflict key | `public.inventory_stock` (`id`)| `CONFIRMED_DATA_MISMATCH` | **151,263 excess rows** (7,203 keys x 22 runs). |
| **Partners** | `OCRD` (`CardCode`) | `CardCode, CardName, CardType, Balance` | `models.go:644-665`, `usecase.go:431-476`| Splits into customers and suppliers | `public.customers`, `suppliers` | `CONFIRMED_CORRECT` | 14 customers, 387 suppliers match exactly. |
| **Addresses** | `CRD1` (`CardCode+Address`) | `Address, AdresType, Street, City` | `models.go:1004-1028`, `usecase:477-524` | Target write path exists; not delivered | `public.partner_addresses` | `CONFIRMED_DATA_MISMATCH` | 1 source row vs 0 target rows. |
| **Sales Orders**| `ORDR`/`RDR1` (`DocEntry+LineNum`)| `DocEntry, DocNum, CardCode, LineTotal` | `models.go:731-778`, `usecase:685-732` | Maps order header & line to v2 relations| `public.sales_orders_v2` | `CONFIRMED_CORRECT` | 1 header / 1 line matches exactly. |
| **Invoices** | `OINV` (`DocEntry`) | `DocEntry, DocNum, DocTotal, DiscSum, Vat`| `tables.go:405-423`, `usecase:747-760` | Header mapped; currency, cancel omitted | `public.invoices` (`id`) | `CONFIRMED_DATA_MISMATCH` | Currency USD default; cancellation lost. |
| **Invoice Lines**| `INV1` (`DocEntry+LineNum`)| `ItemCode, Quantity, Price, LineTotal` | `tables.go:425-439`, `usecase:768-787` | Omits `DiscPrcnt`, `PriceBefDi`, UOM | `public.invoice_lines` (`id`) | `CONFIRMED_DATA_MISMATCH` | 2,192,345 line discounts omitted; UOM NULL. |
| **Payments** | `ORCT`/`RCT2` (`DocEntry+LineNum`)| `DocNum, InvType=13, SumApplied>0` | `models.go:1419-1515`, `usecase:792-865`| Mapping code exists; domain not run | `public.invoice_payments` (`id`)| `CONFIRMED_DATA_MISMATCH` | 961,745 allocations omitted (0 target rows). |
| **Purchase Orders**| `OPOR`/`POR1` (`DocEntry+LineNum`)| `DocNum, CardCode, ItemCode, Quantity` | `models.go:1097-1165`, `usecase:530-595`| Generates `PO-{DocNum}`; fallback code | `public.purchase_orders` (`id`) | `CONFIRMED_CORRECT` | 99 POs / 1,755 lines match loaded slice. |
| **Goods Receipts**| `OPDN`/`PDN1` (`DocEntry+LineNum`)| `BaseEntry, BaseLine, BaseType=22` | `models.go:1237-1281`, `usecase:596-658`| Generates `PO-{BaseEntry}` for link | `public.goods_receipt_notes` | `CONFIRMED_DATA_MISMATCH` | All PO links NULL (DocNum vs BaseEntry). |
| **Stock Movements**| `OINM` (`TransNum`) | `TransNum, ItemCode, Warehouse, In/Out` | `models.go:1327-1413`, `usecase:660-682`| Plain INSERT without conflict key | `public.stock_movements` (`id`)| `CONFIRMED_DATA_MISMATCH` | **546,500 excess rows** due to replayed batches. |
| **Credit Notes** | `ORIN`/`RIN1` (`DocEntry`) | `DocEntry, DocNum, LineTotal, DiscSum` | None | No extractor or target mapping exists | `public.sales_returns` | `OUT_OF_SCOPE_OR_INTENTIONALLY_IGNORED` | 1,387 headers / 7,704 lines not migrated. |
| **Promotions** | None | None | None | No extractor or target mapping exists | `public.promotions` | `OUT_OF_SCOPE_OR_INTENTIONALLY_IGNORED` | Unused in migration. |

---

## 8. Invoice-discount investigation

### Background and history of the issue
Initial analysis observed a mysterious `-23` value associated with invoice discounts, leading to hypotheses of sign flips, data corruption, or broken arithmetic. Forensic source inspection and target reconciliation completely resolved this question.

### Representative record proof
- **SAP source location:** Table `INV1`, `DocEntry = 80306`, `LineNum = 5`.
- **Header context:** Table `OINV`, `DocEntry = 80306`, `DocNum = 2022025735`, `DocDate = 2022-02-11`.
- **Item code:** `INV04287`.
- **Quantity:** `0.325`.
- **Pre-discount price (`PriceBefDi`):** `16.96 SAR`.
- **Stored unit price (`Price`):** `20.86 SAR`.
- **Stored line percentage (`DiscPrcnt`):** `-23.0%`.
- **Line net total (`LineTotal`):** `6.78 SAR`.
- **Line tax (`VatSum`):** `1.02 SAR`.
- **Line gross total (`GTotal`):** `7.80 SAR`.
- **Header discount (`DiscSum`):** `0.00 SAR`.
- **Header total (`DocTotal`):** `88.30 SAR`.

### Mathematical verification of SAP arithmetic
Applying SAP Business One document arithmetic:
1. Base amount before adjustment:
   $$\text{Quantity} \times \text{PriceBefDi} = 0.325 \times 16.96 = 5.512\text{ SAR}$$
2. Unit price adjustment via negative discount:
   $$\text{PriceBefDi} \times \left(1 - \frac{\text{DiscPrcnt}}{100}\right) = 16.96 \times (1 - (-0.23)) = 16.96 \times 1.23 = 20.8608 \rightarrow \text{Stored as } 20.86\text{ SAR}$$
3. Line net total:
   $$\text{Quantity} \times \text{Stored Price} = 0.325 \times 20.86 = 6.7795 \rightarrow \text{Stored as } 6.78\text{ SAR}$$
4. Line gross total:
   $$\text{LineTotal} + \text{VatSum} = 6.78 + 1.02 = 7.80\text{ SAR}$$
5. Effective monetary adjustment:
   $$\text{Base Amount} - \text{Line Total} = 5.512 - 6.780 = -1.268\text{ SAR (approx. } -1.27\text{ SAR)}$$

### Crucial conclusions on line discount:
1. **`-23` is a percentage, not a monetary currency amount:** Blindly mapping `-23` into a monetary `discount_amount` column subtracts 23 currency units from the line, which completely corrupts invoice totals.
2. **Negative percentage represents a commercial surcharge:** The price increased from 16.96 to 20.86.
3. **NEMBUS target line total is gross:** Target `invoice_lines.line_total` is mapped as `LineTotal + VatSum = 7.80`, which matches `GTotal`.
4. **Target line discount was completely omitted:**
   In `packages/sap/schema/tables.go:425-439`, `QueryInvoiceLines` never selected `DiscPrcnt` or `PriceBefDi`.
   In `packages/core/usecase/sap_migration.go:768-787`, `discount_amount` was omitted from `INSERT INTO invoice_lines`. Consequently, all 5,472,485 target lines defaulted to `0.00`.
   Across the 5,472,485 eligible source lines, **2,192,345 lines have nonzero `DiscPrcnt`** (814,610 positive, 1,377,735 negative). All of them were loaded with `0.00` discount in the target.

### Crucial distinction: Header discount is correct
Earlier provisional hypotheses suspected that header discounts were also broken. Sweep 2 proved that **`OINV.DiscSum` is correctly mapped** to `invoices.discount_amount`!
- Source `OINV.DiscSum` distribution across 969,665 invoices: Positive: 23,551; Zero: 911,017; Negative: 35,097; Sum: `-725.45 SAR`.
- Target `invoices.discount_amount` distribution: Positive: 23,551; Zero: 911,017; Negative: 35,097; Sum: `-725.45`.
- Target reconstructive trigger check: Recomputing `total_amount = subtotal + tax_amount - discount_amount` produced **zero mismatches** and a maximum absolute difference of `0.00` across all 969,665 invoices.

---

## 9. Currency investigation

### Source vs target evidence
- **Source currency distribution:**
  In `Qadsiya_Dev`, `SELECT DISTINCT DocCur FROM OINV;` returns exclusively **`SAR`** across all **969,665** invoice headers.
  Line-level `SELECT DISTINCT Currency FROM INV1;` returns exclusively **`SAR`** across all **5,472,485** eligible lines.
- **Target currency distribution:**
  In `nembus_migration_audit_20260922`, `SELECT DISTINCT currency_code, COUNT(*) FROM public.invoices GROUP BY currency_code;` returns:
  `USD`: **969,665 (100%)**.
  Target `invoice_lines` has no currency column.

### Root cause in migration code
1. **Query omission:** `packages/sap/schema/tables.go:405-423` (`QueryInvoicesHeader`) omits `DocCur`.
2. **Model omission:** `packages/sap/mappings/models.go:799-895` (`SAPInvoice` / `CanonicalInvoice`) contains no currency field.
3. **Target insert omission:** `packages/core/usecase/sap_migration.go:747-760` does not supply `currency_code` in the `INSERT INTO invoices` statement.
4. **Target schema default:** `packages/core/db/schema/85_zatca.sql:26` defines `currency_code VARCHAR(3) NOT NULL DEFAULT 'USD'`.
5. **Master data verification:** `SELECT * FROM public.currencies WHERE code = 'SAR';` proves that `SAR` exists in the target master table. The failure is entirely due to code omission and schema defaulting.

### Business impact
All monetary reporting, tax reporting, and ZATCA electronic invoicing compliance files generated from target invoices will be mislabeled as US Dollars instead of Saudi Riyals.

### Developer correction concept
Do not hardcode `'SAR'` in Go code. Explicitly extract `OINV.DocCur`, map it through `CanonicalInvoice`, and supply it to `INSERT INTO invoices (..., currency_code) VALUES (..., $n)`. Implement an approved fallback/policy for genuine foreign currency documents (`DocTotalFC`, `SysCur`).

---

## 10. Payments

### Source allocation evidence
In SAP Business One, incoming customer payments are stored in `ORCT` (headers) and allocated to invoices in `RCT2`.
- `ORCT` payment headers: **1,486**.
- `RCT2` total allocation rows: **971,586**.
- Positive A/R invoice allocations (`InvType = 13` and `SumApplied > 0`): **961,745**.
- Distinct payment documents represented: **1,329**.
- Distinct source invoices referenced: **955,788** (verified via SQL Server query: `SELECT COUNT(DISTINCT DocEntry) FROM RCT2 WHERE InvType = 13 AND SumApplied > 0;`).
- Total monetary value allocated: **64,438,195.83 SAR**.

### Target evidence and code status
- Target table `public.invoice_payments` contains **0 rows**.
- Staging table `staging.sap_migration_batches` contains **0 batches** for domain `incoming_payments`.
- **Code inspection:** A complete transformation exists in `packages/sap/mappings/models.go:1419-1515` (`SAPIncomingPayment.ToCanonicalList`), and an ingestion handler exists in `packages/core/usecase/sap_migration.go:792-865` (`DomainIncomingPayments`).
- **Conclusion:** The payment migration domain was omitted from the migration run configuration. While invoice headers show `paid_amount` and `balance_due`, the underlying transaction history, payment method breakdowns (cash/card/transfer/check), and allocation dates are completely missing from the target database.

---

## 11. Inventory duplicates

### Business grain and root cause
- **Source business grain:** `OITW` represents current inventory balances per item and warehouse (`ItemCode` + `WhsCode`). Filtering for active stock (`OnHand <> 0 OR IsCommited <> 0 OR OnOrder <> 0`) yields exactly **7,203 distinct product/store pairs**.
- **Target table:** `public.inventory_stock` contains **158,466 rows**.
- **Distinct target pairs:** Querying `metadata->>'sap_item_code'` and `metadata->>'sap_whs_code'` yields exactly **7,203 distinct pairs**.
- **Excess duplicate rows:**
  $$158,466 - 7,203 = 151,263\text{ excess rows}$$
- **Multiplication factor:** Every single source pair occurs **exactly 22 times** in the target table.
- **Representative duplicate keys:**
  - `INV00007` at warehouse `01`: 22 rows.
  - `INV00008` at warehouse `01`: 22 rows.
  - `INV00012` at warehouse `01`: 22 rows.
- **Mechanism in code:** `packages/core/usecase/sap_migration.go:412-429` executes a raw `INSERT INTO inventory_stock (...) SELECT ... LIMIT 1;` without an `ON CONFLICT` clause or unique constraint on `(product_id, store_id)`. Staging records show 330 inventory batches; repeated execution multiplied the rows without deduplication.

### Developer correction concept
Add a unique constraint on `(organization_id, product_id, store_id)` in PostgreSQL, and alter the ingestion usecase to use:
```sql
INSERT INTO inventory_stock (...) VALUES (...)
ON CONFLICT (organization_id, product_id, store_id) DO UPDATE SET
  quantity_on_hand = excluded.quantity_on_hand,
  quantity_allocated = excluded.quantity_allocated, ...;
```

---

## 12. Stock-movement duplicates

### Business grain and root cause
- **Source business grain:** `OINM` records inventory movement journal entries. Filtering for the audited date window (`DocDate >= 2024-01-01 AND DocDate < 2026-01-01`) yields **2,163,651 eligible movement records**, uniquely identified by `TransNum`.
- **Target table:** `public.stock_movements` contains **2,710,151 rows**.
- **Distinct target movement keys:** Querying `COUNT(DISTINCT metadata->>'sap_trans_num')` yields exactly **2,163,651**.
- **Excess duplicate rows:**
  $$2,710,151 - 2,163,651 = 546,500\text{ excess duplicate rows}$$
- **Duplicate distribution:**
  - Keys appearing exactly 1 time: **1,796,151**.
  - Keys appearing exactly 2 times: **188,500** (377,000 rows).
  - Keys appearing exactly 3 times: **179,000** (537,000 rows).
- **Representative duplicate keys:**
  - `sap_trans_num = 3681569`: 3 rows in target (dated 2024-01-02).
  - `sap_trans_num = 3681570`: 3 rows in target (dated 2024-01-02).
  - `sap_trans_num = 3681571`: 3 rows in target (dated 2024-01-02).
- **Mechanism in code:** `packages/core/usecase/sap_migration.go:660-682` performs plain `INSERT INTO stock_movements` without a unique constraint or conflict target on the source transaction number. Replaying failed or split batches resulted in duplicate rows.

### Developer correction concept
Add an index or unique constraint on `(organization_id, (metadata->>'sap_trans_num'))` or a dedicated `external_id` column, and implement replay-safe idempotent insertion (`ON CONFLICT DO NOTHING` or `DO UPDATE`).

---

## 13. Remaining confirmed and suspicious findings

This section contains a comprehensive inventory of all findings reconciled across `04-findings.csv`, `11-confirmed-findings.md`, and `12-developer-action-register.csv`.

```
====================================================================================================
FINDING F001 / S2-F001 / A01: Invoice Line Discount Omission
====================================================================================================
Classification: CONFIRMED_CODE_DEFECT / CONFIRMED_DATA_MISMATCH
Severity: High
Domain: Invoices (Line items)
Source Table & Columns: INV1 (DiscPrcnt, PriceBefDi, Price, LineTotal)
Target Table & Columns: public.invoice_lines (discount_amount, unit_price, line_total)
Code Location: packages/sap/schema/tables.go:425-439; packages/sap/mappings/models.go:786-827;
               packages/core/usecase/sap_migration.go:768-787
Expected Behavior: Line discount percentages and raw pricing inputs must be preserved in metadata
                   and converted to an auditable signed monetary adjustment under approved finance rules.
Implemented Behavior: Query omits DiscPrcnt and PriceBefDi; line insert omits discount_amount;
                      target defaults all discount_amount values to 0.00.
Restored Target Evidence: All 5,472,485 target lines have discount_amount = 0.00;
                          Source has 2,192,345 eligible lines with nonzero DiscPrcnt.
Affected Count: 2,192,345 lines.
Representative Key: DocEntry = 80306, LineNum = 5 (DiscPrcnt = -23.0%, target discount = 0.00).
Business Impact: Loss of commercial discount/surcharge auditability; line-level reporting is distorted.
Developer Recommendation: Extract DiscPrcnt & PriceBefDi; store in line metadata; calculate signed
                          monetary discount under agreed formula; populate invoice_lines.discount_amount.
Validation Query: Section 13 query in 13-regression-verification.sql.
Unresolved Policy Question: Should negative percentages be stored as negative monetary discounts
                            (surcharges) or separated into a dedicated adjustment column?

====================================================================================================
FINDING F002 / S2-F002 / A02: Invoice Header Currency Mislabeling
====================================================================================================
Classification: CONFIRMED_CODE_DEFECT / CONFIRMED_DATA_MISMATCH
Severity: High
Domain: Invoices (Header)
Source Table & Column: OINV (DocCur)
Target Table & Column: public.invoices (currency_code)
Code Location: packages/sap/schema/tables.go:405-423; packages/sap/mappings/models.go:799-895;
               packages/core/usecase/sap_migration.go:747-760; packages/core/db/schema/85_zatca.sql:26
Expected Behavior: Invoices must carry the source document currency (SAR).
Implemented Behavior: Query omits DocCur; model lacks currency field; insert omits currency_code;
                      schema defaults to 'USD'.
Restored Target Evidence: All 969,665 target invoices have currency_code = 'USD';
                          All 969,665 source invoices have DocCur = 'SAR'. SAR exists in public.currencies.
Affected Count: 969,665 invoice headers.
Representative Key: DocEntry = 80306 (DocNum = 2022025735).
Business Impact: Severe financial misreporting; currency mismatch breaks ZATCA phase 2 XML generation.
Developer Recommendation: Extract OINV.DocCur; pass through CanonicalInvoice; explicitly insert
                          into public.invoices.currency_code.
Validation Query: SELECT currency_code, COUNT(*) FROM public.invoices GROUP BY currency_code;
Unresolved Policy Question: How should multi-currency transactions or exchange rates (SysCur/DocRate)
                            be preserved if non-SAR invoices appear in future companies?

====================================================================================================
FINDING F003 / S2-F003 / A03: Cancellation Semantics Loss
====================================================================================================
Classification: CONFIRMED_DATA_MISMATCH / CONFIRMED_CODE_DEFECT
Severity: High
Domain: Invoices (Header Status)
Source Table & Columns: OINV (CANCELED, DocStatus)
Target Table & Column: public.invoices (invoice_status)
Code Location: packages/sap/schema/tables.go:405-423; packages/sap/mappings/models.go:846-895
Expected Behavior: Canceled SAP invoices (CANCELED = 'Y' or 'C') should be mapped to
                   invoice_status = 'cancelled' or quarantined per policy.
Implemented Behavior: Status is derived solely from PaidToDate vs DocTotal; CANCELED is not queried;
                      target contains 0 cancelled invoices.
Restored Target Evidence: Target status: paid = 969,612; sent = 52; partially_paid = 1; cancelled = 0.
                          Source OINV has CANCELED='Y' (2,672) and CANCELED='C' (2,672) = 5,344 total.
Affected Count: 5,344 canceled invoice documents.
Representative Key: Any OINV record where CANCELED IN ('Y', 'C').
Business Impact: Canceled invoices appear active/paid, inflating historical sales totals and customer balances.
Developer Recommendation: Query CANCELED and DocStatus; map CANCELED IN ('Y', 'C') to 'cancelled'.
Validation Query: SELECT COUNT(*) FROM public.invoices WHERE invoice_status = 'cancelled';
Unresolved Policy Question: Should canceled documents be migrated as status 'cancelled', or filtered
                            out entirely prior to target insertion?

====================================================================================================
FINDING F004 / S2-F004 / A04: Service Invoice Misclassification
====================================================================================================
Classification: CONFIRMED_CODE_DEFECT / CONFIRMED_DATA_MISMATCH
Severity: Medium
Domain: Invoices (Service vs Item)
Source Table & Column: OINV (DocType)
Target Table & Columns: public.invoices (invoice_type), public.invoice_lines (item_type, product_id)
Code Location: packages/sap/mappings/models.go:882-883; packages/core/usecase/sap_migration.go:753, 769
Expected Behavior: Service invoices (DocType = 'S') should be mapped to a dedicated service invoice type
                   or handled with service-type line items without requiring a physical product SKU.
Implemented Behavior: Query omits DocType; invoice_type is hardcoded to 'standard'; line item_type is
                      hardcoded to 'product'; lines fail product lookup and have NULL product_id.
Restored Target Evidence: Source OINV has 31 headers with DocType = 'S' (and 969,634 with DocType = 'I').
                          Target has 969,665 'standard' invoices; the 31 service lines have product_id NULL.
Affected Count: 31 service invoice headers and their associated lines.
Representative Key: Any OINV record where DocType = 'S'.
Business Impact: Service sales are mislabeled as inventory/product sales; product_id foreign keys are NULL.
Developer Recommendation: Select DocType; map DocType = 'S' to a service invoice representation, or
                          set line item_type = 'service' when ItemCode is blank.
Validation Query: SELECT COUNT(*) FROM public.invoice_lines WHERE product_id IS NULL;
Unresolved Policy Question: Does NEMBUS support service lines natively, or should service lines map to a
                            generic placeholder service product?

====================================================================================================
FINDING F005 / A14: Signed Negative Header Discounts (Confirmed Source Condition)
====================================================================================================
Classification: CONFIRMED_SOURCE_DATA_CONDITION (Validated No Defect)
Severity: Informational / Medium
Domain: Invoices (Header Discounts)
Source Table & Column: OINV (DiscSum)
Target Table & Column: public.invoices (discount_amount)
Code Location: packages/sap/schema/tables.go:416; packages/sap/mappings/models.go:850;
               packages/core/usecase/sap_migration.go:748
Observed Reality: Source OINV has 35,097 rows with negative DiscSum (summing to -725.45 SAR).
                  Mapping correctly carries DiscSum into target invoices.discount_amount.
                  Target trigger subtracts discount_amount, perfectly reconstructing total_amount.
Restored Target Evidence: Target discount_amount sum = -725.45. Zero mathematical discrepancies.
Affected Count: 35,097 invoices with negative header discount.
Representative Key: DocEntry = 89366 (DocNum = 2022034795, DiscSum = -0.44).
Developer Recommendation: DO NOT ALTER this mapping. Protect existing OINV.DiscSum mapping.
                          Document the finance policy that negative header discounts represent surcharges.

====================================================================================================
FINDING F006 / A01: Negative Line Discount Percentage Semantics
====================================================================================================
Classification: CONFIRMED_SOURCE_DATA_CONDITION
Severity: Medium
Domain: Invoice Lines (Discounts)
Source Table & Column: INV1 (DiscPrcnt)
Observed Reality: 1,377,735 eligible lines have negative DiscPrcnt (ranging down to -4,685%).
                  Forensic audit proves these are legitimate SAP surcharge percentages that increase
                  unit price (PriceBefDi -> Price).
Developer Recommendation: Do not apply ABS() or negate negative percentages. Implement signed
                          monetary derivation conforming to SAP formula.

====================================================================================================
FINDING F008 / A15: Omission of A/R Credit Notes and Returns
====================================================================================================
Classification: OUT_OF_SCOPE_OR_INTENTIONALLY_IGNORED
Severity: Informational
Domain: Sales Returns / Credit Notes
Source Tables: ORIN (Headers - 1,387 rows), RIN1 (Lines - 7,704 rows)
Target Tables: public.sales_returns, public.sales_return_lines (0 rows)
Observed Reality: Migration pipeline contains zero extractors, models, or usecase handlers for ORIN/RIN1.
Business Impact: Full sales ledger cannot be reconciled against general ledger without credit notes.
Developer Recommendation: Formalize return scope; if required, implement dedicated ORIN/RIN1 migration.

====================================================================================================
FINDING F009: Missing Brands Table (OMRG)
====================================================================================================
Classification: CONFIRMED_SOURCE_DATA_CONDITION
Severity: Low
Domain: Products (Brands)
Source Table: OMRG (FirmCode, FirmName)
Observed Reality: Table dbo.OMRG does not exist in Qadsiya_Dev. The extractor catches the SQL Server
                  error and returns an empty list. Target products have brand_id = NULL.
Developer Recommendation: Confirm brand omission as an expected source condition; do not treat as a bug.

====================================================================================================
FINDING F010 / S2-F008 / A08: Goods Receipt to Purchase Order Link Failure
====================================================================================================
Classification: CONFIRMED_CODE_DEFECT / CONFIRMED_DATA_MISMATCH
Severity: Medium
Domain: Procurement (GRN to PO Links)
Source Tables & Columns: OPDN/PDN1 (BaseEntry, BaseLine, BaseType), OPOR/POR1 (DocEntry, DocNum)
Target Tables & Columns: public.goods_receipt_notes (purchase_order_id),
                         public.goods_receipt_note_items (purchase_order_line_id)
Code Location: packages/sap/mappings/models.go:1153, 1251-1276;
               packages/core/usecase/sap_migration.go:614-615, 646
Expected Behavior: When PDN1 has BaseType = 22, the GRN must link to the target PO and PO line.
Implemented Behavior: Purchase order mapper creates PO number as 'PO-' + po.DocNum.
                      GRN mapper creates referenced PO number as 'PO-' + l.BaseEntry.
                      In SAP, BaseEntry is OPOR.DocEntry, NOT OPOR.DocNum!
                      The lookup fails; all target GRN PO links are NULL.
Restored Target Evidence: All 7,128 GRNs have purchase_order_id = NULL; all 73,639 items have
                          purchase_order_line_id = NULL. Source has 15 linked GRNs and 248 linked lines.
Affected Count: 15 PO-linked GRN headers and 248 lines in the loaded 2024-2025 slice.
Representative Key: GRN DocEntry = 4158 (DocNum = 2024002818, BaseEntry = 11675, BaseType = 22).
                    PO DocEntry = 11675 has DocNum = 2024000032. GRN looked for 'PO-11675' instead of 'PO-2024000032'.
Business Impact: Procurement history is disconnected; three-way matching cannot be performed.
Developer Recommendation: Map PO numbers consistently using a shared source key (or resolve BaseEntry
                          via OPOR.DocEntry -> DocNum lookup before constructing the foreign key).
Validation Query: SELECT COUNT(*) FROM public.goods_receipt_notes WHERE purchase_order_id IS NOT NULL;

====================================================================================================
FINDING SUSPICIOUS-01: Arbitrary Supplier and Store Fallback in PO / GRN Ingestion
====================================================================================================
Classification: SUSPICIOUS_MAPPING (Code Risk)
Severity: Medium
Domain: Procurement (Supplier / Store Resolution)
Code Location: packages/core/usecase/sap_migration.go:549-552, 616-619
Observed Reality: When a supplier or store code is not found, the query executes:
                  COALESCE((SELECT id FROM suppliers WHERE code = $3), (SELECT id FROM suppliers LIMIT 1))
                  This silently attaches purchase orders and GRNs to an arbitrary supplier or store!
Target Evidence: In the loaded 2024-2025 slice, supplier/store code mismatches were 0, so arbitrary
                 fallback was not triggered for loaded rows. However, the code remains inherently unsafe.
Developer Recommendation: Remove the fallback (SELECT id ... LIMIT 1); fail the row or quarantine it.

====================================================================================================
FINDING S2-F006 / A06: Invoice Line Unit of Measure Omission
====================================================================================================
Classification: CONFIRMED_DATA_MISMATCH / CONFIRMED_CODE_DEFECT
Severity: Medium
Domain: Invoice Lines (UOM)
Source Table & Column: INV1 (unitMsr)
Target Table & Column: public.invoice_lines (uom_id)
Code Location: packages/sap/schema/tables.go:436; packages/core/usecase/sap_migration.go:768-787
Expected Behavior: Source line UOM should resolve to public.units_of_measure.id.
Implemented Behavior: Query selects unitMsr, but CanonicalInvoiceLine stores it only in metadata;
                      usecase omits uom_id from INSERT; target uom_id is NULL for all lines.
Restored Target Evidence: 5,472,454 eligible source lines have nonblank unitMsr (e.g., 'kg', 'UNIT').
                          All 5,472,485 target invoice lines have uom_id = NULL.
Affected Count: 5,472,454 invoice lines.
Representative Key: DocEntry = 80306, LineNum = 5 (unitMsr = 'kg', target uom_id = NULL).
Business Impact: Quantity reporting lacks physical units; commercial printing and invoicing distorted.
Developer Recommendation: Resolve line unitMsr against public.units_of_measure and populate uom_id.

====================================================================================================
FINDING S2-F010 / A10: Missing Partner Address Domain
====================================================================================================
Classification: CONFIRMED_DATA_MISMATCH / CONFIRMED_CODE_DEFECT
Severity: Low
Domain: Business Partners (Addresses)
Source Table: CRD1 (CardCode, Address, AdresType, Street, City, Country)
Target Table: public.partner_addresses (0 rows)
Code Location: packages/core/usecase/sap_migration.go:477-524
Observed Reality: Source CRD1 contains 1 address row (CardCode = 'V00059').
                  Target partner_addresses has 0 rows; no partner_addresses staging batch exists.
Developer Recommendation: Include address domain in migration run or explicitly document exclusion.

====================================================================================================
FINDING S2-F011 / A11: Historical Run Configuration Discrepancy
====================================================================================================
Classification: DOCUMENTATION_GAP
Severity: Medium
Domain: Migration Configuration / Provenance
Observed Reality: Config in repository specifies invoice_start_date = '2023-01-01'.
                  However, the restored target backup contains 315,715 invoices dated between 2021
                  and 2022. The effective historical run configuration was not recorded.
Developer Recommendation: Persist the exact effective configuration (start/end dates, filters) in
                          staging.sap_migration_batches with each migration execution.

====================================================================================================
FINDING S2-F012 / A12: Failed Invoice Batch Auditability Gap
====================================================================================================
Classification: DOCUMENTATION_GAP
Severity: Medium
Domain: Staging & Observability
Observed Reality: Staging table staging.sap_migration_batches contains 366 failed invoice batches
                  with record_count_sum = 182,686 and empty error_message text.
                  Target invoice header count (969,665) equals source count, proving no rows were
                  ultimately lost, but the empty error message prevents diagnosing what failed.
Developer Recommendation: Ensure batch execution logs capture structured error strings upon failure.

====================================================================================================
FINDING S2-F013 / A13: Tax Master and Reference Omission
====================================================================================================
Classification: DOCUMENTATION_GAP
Severity: Medium
Domain: Taxes (Master Data & Line References)
Target Tables & Columns: public.tax_categories (0 rows);
                         public.invoice_lines (tax_category_id, tax_rate - all NULL)
Observed Reality: Calculated tax amounts (tax_amount = 8,536,900.76) match source VatSum exactly.
                  However, tax_categories is empty, and line tax_category_id and tax_rate are NULL.
Developer Recommendation: Formalize whether tax reference migration is in scope or if calculated VAT
                          amounts alone satisfy requirements.

====================================================================================================
FINDING MISSING-LINES: Forensic Proof of the Three Missing Invoice Lines
====================================================================================================
Classification: CONFIRMED_SOURCE_DATA_CONDITION (Orphan Records in Source)
Severity: Low / Informational
Domain: Invoices (Line Parity)
Observed Reality: Raw source INV1 = 5,472,488 lines; restored target = 5,472,485 lines (Difference = 3).
Forensic SQL Proof: Querying Qadsiya_Dev:
  SELECT l.DocEntry, l.LineNum, l.ItemCode, l.Quantity, l.Price, l.LineTotal, l.DiscPrcnt
  FROM INV1 l LEFT JOIN OINV h ON l.DocEntry = h.DocEntry WHERE h.DocEntry IS NULL;
Returns EXACTLY 3 orphan records:
  1. DocEntry = 1, LineNum = 0, ItemCode = 'INV00001', Qty = 38.0, Price = 360.0, LineTotal = 13680.0
  2. DocEntry = 2, LineNum = 0, ItemCode = 'INV00001', Qty = 10.0, Price = 360.0, LineTotal = 3600.0
  3. DocEntry = 3, LineNum = 0, ItemCode = 'INV00001', Qty = 2.0,  Price = 360.0, LineTotal = 720.0
Cause: Table OINV in Qadsiya_Dev has MIN(DocEntry) = 4 and MAX(DocEntry) = 969668 (Total = 969,665).
       Headers 1, 2, and 3 do not exist in OINV. The migration extracts lines by inner joining or
       querying WHERE DocEntry IN (...) for existing headers. These 3 orphan lines were correctly
       filtered out by relational integrity.
Conclusion: TARGET COUNT OF 5,472,485 IS 100% CORRECT FOR ALL ELIGIBLE PARENTED INVOICE LINES.
```

---

## 14. Confirmed-correct behavior

To ensure the migration developer does not inadvertently "fix" or regress valid code, the following behaviors are **proven correct** by verified mathematical aggregates and relational parity checks:

1. **Header discount mapping (`OINV.DiscSum` -> `invoices.discount_amount`):**
   - *Evidence:* Exact match across all 969,665 invoices. Positive count: 23,551; zero count: 911,017; negative count: 35,097; sum: `-725.45 SAR`; weighted sum: `234,975,426.96`; squared sum: `283,748,598,101,569.80`.
   - *Coverage limitation:* Applies to document header level; does not compensate for omitted line discounts.
2. **Invoice header population & key mapping:**
   - *Evidence:* 969,665 of 969,665 source invoices present in target; mapped via standard prefix `INV-SAP-{DocNum}`; `metadata.sap_doc_entry` is 100% populated and unique.
3. **Eligible invoice line count & relational integrity:**
   - *Evidence:* 5,472,485 lines match the inner-join eligible source population exactly. Uniqueness of `(invoice_id, line_number)` and metadata `(sap_doc_entry, sap_line_num)` is 100% verified.
4. **Mathematical invoice total reconstructions:**
   - *Evidence:* Evaluated against target trigger formula `subtotal + tax_amount + shipping - discount_amount`. Subtotal mismatch count = 0, total mismatch count = 0; max absolute difference = `0.00` across all 969,665 target rows.
5. **Product master catalog:**
   - *Evidence:* All 16,955 source items in `OITM` exist in `public.products` with matching `sap_item_code`.
6. **Product categories:**
   - *Evidence:* All 27 source item groups in `OITB` exist in `public.product_categories` with matching codes.
7. **Business partners (Customers & Suppliers):**
   - *Evidence:* All 14 customer cards and 387 supplier cards in `OCRD` reconcile to `public.customers` and `public.suppliers`.
8. **Warehouses / Stores:**
   - *Evidence:* All 2 warehouses in `OWHS` match `public.stores`.
9. **Sales orders (v2):**
   - *Evidence:* The 1 source order and 1 line in `ORDR`/`RDR1` match `public.sales_orders_v2` and `_lines_v2`.
10. **Barcode deduplication & normalization:**
    - *Evidence:* 27,591 raw source rows in `OBCD` contain duplicate entries; target deduplicates them to exactly 27,445 unique barcodes, matching distinct source barcode values.
11. **Price list & price slice:**
    - *Evidence:* 10 price lists and 16,686 positive-price product prices match the validated source slice.
12. **Source-conditioned absence of brands and storage bins:**
    - *Evidence:* Table `OMRG` is absent and `OBIN` has 0 rows in `Qadsiya_Dev`; target having 0 brands and 0 storage locations is confirmed correct.

---

## 15. Developer action register

This register provides a prioritized, actionable roadmap for the migration developer.

### Categorized Action Table

| Priority | Action ID | Finding Ref | Severity | Domain | Required Conceptual Correction | Affected Population | Prerequisite Policy Decision | Regression Test | Status |
|---:|---|---|---|---|---|---:|---|---|---|
| **1** | **A01** | F001 / S2-F001 | High | Line Discounts | Select `DiscPrcnt` & `PriceBefDi`; compute signed monetary discount; populate `invoice_lines.discount_amount`. Never copy `-23` as currency. | 2,192,345 lines | Finance sign & surcharge policy | Exact DocEntry 80306 Line 5 check + all-line discount nonzero count | `REMEDIATION_REQUIRED` |
| **2** | **A02** | F002 / S2-F002 | High | Invoice Currency | Select `OINV.DocCur`; map through canonical model; explicitly insert into `invoices.currency_code`. | 969,665 invoices | Currency fallback & exchange rate rules | `SELECT DISTINCT currency_code` returns SAR | `REMEDIATION_REQUIRED` |
| **3** | **A05** | S2-F005 | High | Payment Allocations| Deliver incoming-payments migration domain; link allocations to target invoices; ensure idempotent ingestion. | 961,745 allocations (64.44M SAR) | Payment-detail scope & link contract | `invoice_payments` row count = 961,745; sum matches | `REMEDIATION_REQUIRED` |
| **4** | **A07** | S2-F007 | High | Inventory Stock | Add unique constraint on `(organization_id, product_id, store_id)`; implement upsert (`ON CONFLICT DO UPDATE`). | 151,263 excess rows | Current snapshot vs historical grain | Run batch twice; assert count = 7,203 | `REMEDIATION_REQUIRED` |
| **5** | **A09** | S2-F009 | High | Stock Movements | Add unique constraint on `(organization_id, metadata->>'sap_trans_num')`; implement replay-safe ingestion. | 546,500 excess rows | TransNum uniqueness scope | Run batch twice; assert count = 2,163,651 | `REMEDIATION_REQUIRED` |
| **6** | **A08** | F010 / S2-F008 | Medium | GRN PO Links | Build PO link in GRN mapper using `OPOR.DocNum` (or resolve `BaseEntry` to `DocNum`); populate `purchase_order_id`. | 15 GRNs / 248 lines | BaseType=22 contract | `purchase_order_id IS NOT NULL` on linked GRNs | `REMEDIATION_REQUIRED` |
| **7** | **A03** | F003 / S2-F003 | High | Cancellation | Select `CANCELED` & `DocStatus`; map `CANCELED IN ('Y','C')` to `invoice_status = 'cancelled'`. | 5,344 invoices | Exclude vs load as cancelled | Cross-tab of source CANCELED vs target status | `POLICY_REQUIRED` |
| **8** | **A04** | F004 / S2-F004 | Medium | Service Invoices | Select `DocType`; preserve service invoice representation or set line `item_type = 'service'`. | 31 invoices | Service document representation | Check 31 service keys in target | `POLICY_REQUIRED` |
| **9** | **A06** | S2-F006 | Medium | Line UOM | Map `INV1.unitMsr` to `public.units_of_measure.id`; populate `invoice_lines.uom_id`. | 5,472,454 lines | UOM mapping fallback rules | `COUNT(*) FILTER (WHERE uom_id IS NULL)` | `REMEDIATION_REQUIRED` |
| **10** | **A10** | S2-F010 | Low | Partner Address | Deliver partner address migration domain or explicitly document exclusion. | 1 address | Address scope decision | `COUNT(*)` in `partner_addresses` = 1 | `POLICY_REQUIRED` |
| **11** | **A11** | S2-F011 | Medium | Run Config | Record effective date window and parameters in `staging.sap_migration_batches`. | 315,715 pre-2023 invoices | Configuration audit policy | Inspect batch payload metadata | `DOCUMENTATION_REQUIRED` |
| **12** | **A12** | S2-F012 | Medium | Batch Diagnostics | Capture structured error strings in `staging.sap_migration_batches.error_message`. | 366 failed batches | Observability contract | Verify error text populated on failure | `DOCUMENTATION_REQUIRED` |
| **13** | **A13** | S2-F013 | Medium | Tax References | Formalize whether tax category master & line tax reference IDs are required. | 5,472,485 lines | Tax compliance policy | `tax_categories` count & line references | `POLICY_REQUIRED` |
| **14** | **A15** | F008 | Informational| Sales Returns | Formalize whether A/R credit notes (`ORIN`/`RIN1`) are in scope. | 1,387 credit notes | Return ledger scope | Return count check | `POLICY_REQUIRED` |
| **15** | **A16** | SUSPICIOUS-01 | Medium | Fallback Logic | Remove `LIMIT 1` arbitrary fallback for missing suppliers/stores; fail or quarantine unresolved records. | Code risk | Error handling contract | Code diff review | `REMEDIATION_REQUIRED` |

---

## 16. Regression and revalidation procedure

When the migration developer delivers code changes, the audit team must execute the following step-by-step verification protocol:

### Step-by-Step Procedure
1. **Record the developer commit and diff:**
   Obtain the exact commit hash and review the git diff:
   `git show --stat <commit_hash>`
2. **Review intended changes only:**
   Ensure changes are confined to mapping logic, extraction queries, and ingestion upserts. Verify no unrelated files or schemas were inadvertently altered.
3. **DO NOT reuse or mutate `nembus_migration_audit_20260922`:**
   The existing audit database is an immutable before-state snapshot. Never run a new migration into it.
4. **Create a newly authorized post-fix target database:**
   Create a fresh, isolated database (e.g., `nembus_migration_audit_postfix`) and execute the corrected migration pipeline into it.
5. **Lock the new target database read-only:**
   `ALTER DATABASE nembus_migration_audit_postfix SET default_transaction_read_only = on;`
6. **Execute regression SQL script:**
   Run `migration-analysis/13-regression-verification.sql` in a strictly read-only session against the post-fix database.
7. **Verify specific defect corrections:**
   - *Line discounts:* Run Section 4 & 5 of regression SQL; verify nonzero line discounts exist and exact invoice 80306 line 5 has derived discount matching approved formula.
   - *Currency:* Verify `SELECT DISTINCT currency_code FROM public.invoices;` returns exclusively `SAR`.
   - *Payments:* Verify `SELECT COUNT(*) FROM public.invoice_payments;` equals `961,745`.
   - *Inventory duplicates:* Verify `inventory_stock` row count equals `7,203` (excess duplicate count = 0).
   - *Stock-movement duplicates:* Verify `stock_movements` row count equals `2,163,651` (excess duplicate count = 0).
   - *GRN links:* Verify `SELECT COUNT(*) FROM public.goods_receipt_notes WHERE purchase_order_id IS NOT NULL;` equals `15`.
   - *Cancellations:* Verify `SELECT COUNT(*) FROM public.invoices WHERE invoice_status = 'cancelled';` matches policy.
8. **Verify non-regression of confirmed-correct areas:**
   Verify that header discount sum is still `-725.45`, total invoice count is still `969,665`, eligible line count is still `5,472,485`, and total reconstruction difference remains `0.00`.
9. **Update action register & audit artifacts:**
   Update finding statuses from `REMEDIATION_REQUIRED` to `VERIFIED_CLOSED` with concrete SQL evidence.
10. **Publish final closure report:**
    Issue a concise verification summary comparing before-and-after metrics.

### Quantitative success criteria

| Verification Metric | Target Audit Baseline (Before) | Post-Fix Success Criterion |
|---|---|---|
| **Invoice line nonzero discount count** | `0` | **`2,192,345`** (or approved filtered count) |
| **Exact line 5 discount (`DocEntry 80306`)** | `0.00` | Derived signed adjustment (approx `-1.27 SAR`) |
| **Invoice currency distribution** | `USD: 969,665` | **`SAR: 969,665`** (0 USD) |
| **Payment allocation row count** | `0` | **`961,745`** rows |
| **Total payment amount allocated** | `0.00` | **`64,438,195.83 SAR`** |
| **Inventory stock row count** | `158,466` (22x duplicated) | **`7,203`** rows (0 excess duplicates) |
| **Stock movement row count** | `2,710,151` (duplicated) | **`2,163,651`** rows (0 excess duplicates) |
| **PO-linked GRN headers** | `0` linked | **`15`** linked headers |
| **PO-linked GRN item lines** | `0` linked | **`248`** linked lines |
| **Canceled invoice count** | `0` cancelled | **`5,344`** cancelled (or documented filtered count) |
| **Header discount sum** | `-725.45` | **`-725.45`** (Must remain identical) |
| **Reconstructive total mismatch count** | `0` (diff = 0.00) | **`0`** (diff = 0.00) |

---

## 17. Read-only query and artifact index

### Artifact inventory (`migration-analysis/`)

| File Name | Nature / Status | Superseded By | Description & Important Contents |
|---|---|---|---|
| `01-repository-and-environment.md` | Baseline Report | Partially superseded by `06` and this handoff | Documents repository preflight, git metadata, source counts, and initial backup inspection. |
| `02-mapping-inventory.csv` | Comprehensive Inventory | Authoritative for field mappings | Detailed CSV inventory covering all 22 domains, transformations, code locations, and status. |
| `03-invoice-discount-analysis.md`| Forensic Analysis | Authoritative for discount arithmetic | Detailed mathematical proof of SAP line discounts, `-23%` surcharge, and header DiscSum. |
| `04-findings.csv` | Finding Register | Cross-referenced with `11` and this handoff | Tabular register of findings F001–F012 from Sweep 1. |
| `05-stg-backup-restore-plan.md` | Historical Plan | Superseded by `06-restore-execution.md` | Pre-restore execution plan written before authenticated target access. |
| `06-restore-execution.md` | Restore Proof | Authoritative for restore history | Documents successful Path A restore, preflight checks, database isolation, and read-only proof. |
| `06-restore.log` | Raw System Log | Primary Evidence | 121,656-byte verbose `pg_restore` log showing clean execution and foreign key completion. |
| `07-target-schema-inventory.csv` | Catalog Inventory | Authoritative for target catalog | Comprehensive relation inventory of 114 target tables, row counts, and schema distributions. |
| `08-source-target-parity-summary.csv` | Parity Summary | Authoritative for domain parity | High-level comparison table of source vs target counts, missing rows, and duplicates. |
| `09-invoice-discount-reconciliation.csv` | Field Reconciliation | Authoritative for exact record | Field-by-field reconciliation of representative invoice 80306 and line 5. |
| `10-currency-reconciliation.csv` | Currency Reconciliation| Authoritative for currency | Breakdown of source SAR vs target USD across headers, lines, and currency masters. |
| `11-confirmed-findings.md` | Finding Register | Authoritative for Sweep 2 findings | Confirmed findings S2-F001 through S2-F013 based on restored target evidence. |
| `12-developer-action-register.csv` | Action Register | Authoritative for actions | Prioritized action items A01 through A15 for developer remediation. |
| `13-regression-verification.sql` | SQL Test Suite | Authoritative validation script | 282 lines of pure SELECT statements for complete target regression verification. |
| `analysis-summary.json` | Machine Summary | Authoritative JSON | Machine-readable summary of repository, source counts, and revised discount conclusion. |
| `sweep-2-summary.json` | Machine Summary | Authoritative JSON | Machine-readable summary of Sweep 2 restore, counts, reconciliations, and findings. |
| `SAP_NEMBUS_MIGRATION_AUDIT_HANDOFF.md` | Final Handoff | None (This document) | Consolidated, self-contained authoritative project handoff. |

### Query index for `13-regression-verification.sql`
- **Lines 4–6:** Verify database identity and enforce read-only settings (`transaction_read_only = on`).
- **Lines 8–27:** Inventory all tables, views, and sequences across application schemas.
- **Lines 29–49:** Invoice header status, currency distribution, and `sap_doc_entry` metadata coverage.
- **Lines 50–70:** Invoice line discount null/zero/nonzero rates, line UOM coverage, and metadata duplicate checks.
- **Lines 71–111:** Exact target lookup for invoice `80306` (`INV-SAP-2022025735`) and line 5.
- **Lines 112–143:** Reconstructive total verification across all 969,665 invoices using the target trigger formula.
- **Lines 144–159:** Header discount sum, weighted sum, squared sum, and tax/subtotal aggregates.
- **Lines 160–179:** Target orphan line checks, cancellation counts, and status breakdowns.
- **Lines 180–201:** Product catalog source-key coverage, category/brand NULL checks, and tax category inventory.
- **Lines 202–217:** Inventory stock duplicate grain proof (reports 22x repetition factor).
- **Lines 218–236:** Stock movement duplicate grain proof (reports 546,500 excess rows).
- **Lines 237–250:** PO and GRN counts, supplier/store link checks, and broken PO foreign key checks.
- **Lines 251–275:** Payment row counts and staging migration batch audit counts (including failed batches).
- **Lines 276–282:** Sales orders (v2), sales returns, and partner address inventory counts.

---

## 18. Safety state and prohibited actions

The following boundaries must be strictly observed by all subsequent agents, engineers, and chats:

```text
+--------------------------------------------------------------------------------------------------+
|                                    STRICT SAFETY CONSTRAINTS                                     |
+--------------------------------------------------------------------------------------------------+
| 1. SQL Server database Qadsiya_Dev must remain strictly READ-ONLY.                               |
| 2. PostgreSQL database nembus_migration_audit_20260922 must remain strictly READ-ONLY.            |
| 3. DO NOT rerun the migration pipeline into nembus_migration_audit_20260922.                     |
| 4. DO NOT use nembus_migration_audit_20260922 as a development database.                         |
| 5. DO NOT restore D:\stg-backup again; the target audit database is already complete.            |
| 6. DO NOT modify target rows to "simulate" fixes or corrections.                                 |
| 7. DO NOT alter, move, or rename backup file D:\stg-backup.                                      |
| 8. DO NOT modify SAP source data or production databases.                                        |
| 9. DO NOT modify application or migration code in this worktree (Investigation scope only).      |
| 10. DO NOT touch unrelated git paths (apps/pos-client/frontend, packages/core/db.zip).           |
| 11. DO NOT commit, stage, push, pull, merge, reset, or switch git branches.                      |
| 12. DO NOT record, print, or disclose PostgreSQL passwords, tokens, or full connection strings.  |
| 13. DO NOT stop, restart, or reconfigure the running postgresql-x64-18 Windows service.         |
+--------------------------------------------------------------------------------------------------+
```

---

## 19. Exact next-chat operating instructions

### Immediate instructions for a new chat session
1. **Load this handoff:** Read `migration-analysis/SAP_NEMBUS_MIGRATION_AUDIT_HANDOFF.md` first. Treat it as the authoritative baseline.
2. **Do not repeat completed work:** Do not perform environment discovery, do not run table row-count scripts from scratch, and do not attempt to restore the backup.
3. **Do not ask the user for credentials:** Database connections on localhost have already been established and verified.
4. **Do not modify code:** If the user asks you to implement fixes, remind them that this workspace is reserved for investigation and revalidation, and that implementation belongs to the migration developer.
5. **If the developer has not supplied fixes:** Assist the user in preparing formal bug tickets, acceptance matrices, or stakeholder policy briefs based on Section 13 and Section 15.
6. **If the developer has supplied fixes:** Follow Section 16 to inspect the commit diff and execute read-only regression checks against a fresh, separately authorized target database.

### Compact system checkpoint

```text
====================================================================================================
CURRENT SYSTEM CHECKPOINT
====================================================================================================
Repository Path:           D:\nastecsol\Nembus
Git Branch:                agent-v1
HEAD Commit:               c10b338948311646afaaa8181300b989a8caf7f7
Source Database:           SQL Server Qadsiya_Dev (localhost, Windows Integrated, Read-Only)
Target Audit Database:     PostgreSQL nembus_migration_audit_20260922 (127.0.0.1:5432, Read-Only)
Source Invoice Population: 969,665 headers | 5,472,488 raw lines (5,472,485 eligible parented lines)
Target Invoice Population: 969,665 headers | 5,472,485 lines (Exact 1:1 match on eligible lines)
Missing Lines Status:      3 raw lines are orphan INV1 records (DocEntry 1, 2, 3; headers absent in OINV)
Major Confirmed Defects:   1. Line discounts omitted (2,192,345 lines default to 0.00; -23% is surcharge)
                           2. Invoice currency mislabeled (100% USD instead of source SAR)
                           3. Payment detail omitted (961,745 allocations / 64.44M SAR missing)
                           4. Inventory stock duplicated (151,263 excess rows; 22x repetition)
                           5. Stock movements duplicated (546,500 excess duplicate rows)
                           6. GRN PO links 100% broken (DocNum vs BaseEntry string mismatch)
                           7. Cancellation flags lost (5,344 canceled documents mapped as active)
Confirmed Correct Areas:   Header DiscSum (-725.45 SAR), Reconstructive invoice totals (0.00 diff),
                           Products (16,955), Categories (27), Customers (14), Suppliers (387), Stores (2)
Current Action Owner:      Migration Developer (Implementation) & Business Stakeholders (Policy)
Next Authorized Activity:  Review developer fix commit / diff and execute Section 16 revalidation protocol
====================================================================================================
```
