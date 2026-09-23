# Phase-Wise Implementation Plan

**Purpose:** Remediation roadmap for the 13 developer actions (A01–A15) derived from the read-only audit findings. Each phase lists the code/schema locations, the concrete changes required, the regression verification query from `13-regression-verification.sql`, and the policy owner decisions needed before or during implementation.

**Grounding:** Every gap below was verified against the actual source code in this repository at commit `c10b338948311646afaaa8181300b989a8caf7f7`. Code references point to files on disk.

---

## Action Register Summary

| ID | Title | Severity | Domain | Code Locations (verified) | Status |
|---|---|---|---|---|---|
| A01 | Invoice line discount omission | High | invoice_lines | `tables.go:425-439`; `models.go:786-827`; `usecase/sap_migration.go:768-787` | REMEDIATION_REQUIRED |
| A02 | Invoice currency default to USD | High | invoices | `tables.go:405-423`; `models.go:799-814`; `usecase/sap_migration.go:747-760`; `85_zatca.sql:26,63` | REMEDIATION_REQUIRED |
| A03 | Cancellation semantics lost | High | invoices | `tables.go:405-423`; `models.go:846-855`; `usecase/sap_migration.go:747-760` | POLICY_REQUIRED |
| A04 | Service-document misclassification | Medium | invoices/lines | `tables.go:405-423`; `models.go:882-883`; `usecase/sap_migration.go:747-760,768-787` | POLICY_REQUIRED |
| A05 | Payment detail absent | High | invoice_payments | `tables.go:254-286`; `models.go:1419-1515`; `usecase/sap_migration.go:792-867`; `engine.go:20-41` | REMEDIATION_REQUIRED |
| A06 | Invoice line UOM omitted | Medium | invoice_lines | `tables.go:425-439`; `models.go:786-797`; `usecase/sap_migration.go:768-787` | REMEDIATION_REQUIRED |
| A07 | Inventory batches not idempotent | High | inventory_stock | `usecase/sap_migration.go:412-429`; `40_inventory.sql:5-23` | REMEDIATION_REQUIRED |
| A08 | GRN-to-PO links unresolved | Medium | goods_receipt_notes | `models.go:1153,1251-1253,1276`; `usecase/sap_migration.go:543-655` | REMEDIATION_REQUIRED |
| A09 | Stock movement batches not idempotent | High | stock_movements | `usecase/sap_migration.go:660-682`; `40_inventory.sql:25-47` | REMEDIATION_REQUIRED |
| A10 | Partner address not loaded | Low | partner_addresses | `engine.go:20-41`; `usecase/sap_migration.go:477-528`; `tables.go` (CRD1) | POLICY_REQUIRED |
| A11 | Run config not reproducible | Medium | documentation | `engine.go:792-864`; `config` files | DOCUMENTATION_REQUIRED |
| A12 | Failed batch diagnostics insufficient | Medium | staging/audit | `95_sap_staging.sql`; `usecase/sap_migration.go` error handling | DOCUMENTATION_REQUIRED |
| A13 | Tax reference/master undocumented | Medium | taxes | `85_zatca.sql`; `models.go:786-827`; `usecase/sap_migration.go:768-787` | POLICY_REQUIRED |
| A14 | Header discounts — no action needed | Informational | invoices | N/A | VALIDATED_NO_DEFECT |
| A15 | Returns/credit notes out of scope | Informational | returns | N/A | POLICY_REQUIRED |

---

## Phase 0 — Prerequisites: Policy & Contract Definition

**Owner:** Business / Finance stakeholders + Migration Developer
**Duration:** 1–2 days (parallel with code inspection)
**Goal:** Establish the business policy decisions that block remediation of A01, A02, A03, A04, A10, A13, A15. No code changes in this phase.

### Deliverables required before Phase 1 code changes:

1. **Line discount policy (A01, A06):**
   - Signed vs unsigned monetary representation for negative `DiscPrcnt` (surcharges).
   - Rounding contract: per-line or document-level, currency (SAR), decimal places (2 for amount, 4 for rate).
   - Decision: store raw `DiscPrcnt` and `PriceBefDi` in `invoice_lines.metadata` as auditable source evidence.
   - Decision: populate `discount_amount` as a signed monetary value derived via `Quantity × PriceBefDi × (DiscPrcnt/100)` rounded to 2 decimal places, OR leave `discount_amount = 0.00` and store the raw percentage only. **This decision directly changes the target `invoice_lines` insert.**

2. **Currency policy (A02):**
   - Map `OINV.DocCur` → `invoices.currency_code` explicitly.
   - Fallback behavior for any non-SAR invoices (currently 0, but documented per audit finding C004): error, default, or multi-currency support.
   - Exchange rate handling: `DocRate` / `SysCur` — store in metadata or populate `exchange_rate` column.

3. **Cancellation & status policy (A03):**
   - Decision: migrate `CANCELED = 'Y'/'C'` invoices as `invoice_status = 'cancelled'`, OR exclude them from the target entirely.
   - Decision: if excluded, record excluded count in staging audit metadata.

4. **Service-document policy (A04):**
   - Decision: map `DocType = 'S'` → `invoice_type = 'service'` (requires schema enum extension), OR keep as `standard` with `item_type = 'service'` on lines.
   - Decision: for service lines with no `ItemCode`, insert with `product_id = NULL` and `item_type = 'service'`.

5. **Payment scope decision (A05):**
   - Decision: include `incoming_payments` domain in the ETL run configuration, OR explicitly document payment-detail exclusion (header-level `paid_amount` only).
   - If included: confirm `ORCT`/`RCT2` extraction queries and `DomainIncomingPayments` usecase are activated in `engine.go` domain list.

6. **Partner address scope (A10):**
   - Decision: include `DomainBPAddresses` in ETL run, OR document as explicit scope exclusion.

7. **Tax reference scope (A13):**
   - Decision: populate `tax_categories` master from SAP `OSTC` / `OVTG`, OR document calculated-only behavior with `tax_category_id = NULL`.

8. **Returns scope (A15):**
   - Decision: implement `ORIN`/`RIN1` migration domain (future phase), OR formally accept exclusion from migration parity.

9. **Run configuration contract (A11):**
   - Decision: persist the exact effective `invoice_start_date`, `invoice_end_date`, and all ETL config parameters into `staging.sap_migration_batches` or a dedicated `staging.migration_run_config` table with every run.

10. **Batch error observability (A12):**
    - Decision: require non-empty `error_message` for all failed batches; reconcile `record_count_sum` against source/target row counts.

### Regression verification queries that depend on these policies:
- A01: `13-regression-verification.sql` lines 92–110 (exact line lookup for DocEntry 80306/LineNum 5)
- A02: `13-regression-verification.sql` lines 171–174 (currency code distribution)
- A03: `13-regression-verification.sql` lines 166–178 (cancelled status count)
- A04: `13-regression-verification.sql` lines 107–110 (product_id NULL count)
- A05: `13-regression-verification.sql` lines 252–253 (payment count)
- A06: `13-regression-verification.sql` lines 51–59 (uom_id null count)
- A10: `13-regression-verification.sql` line 281 (partner_address_count)
- A13: `13-regression-verification.sql` lines 198–200 (tax_category_count)

---

## Phase 1 — Invoice Header Corrections (A02, A03, A04)

**Owner:** Migration Developer
**Code locations to modify:**
- `packages/sap/schema/tables.go:405-423` — `QueryInvoicesHeader`
- `packages/sap/mappings/models.go:799-814` — `SAPInvoice` struct
- `packages/sap/mappings/models.go:846-900` — `SAPInvoice.ToCanonical()`
- `packages/core/usecase/sap_migration.go:747-760` — invoice header INSERT
- `packages/core/db/schema/85_zatca.sql:63` — `currency_code` column (already exists; remove reliance on default)

### Changes:

#### A02 — Map source invoice currency (High priority, 1 day)

1. **`tables.go:405-423`** — Add `DocCur` (and optionally `DocTotalFC`, `SysCur`, `DocRate`) to `QueryInvoicesHeader`:
   ```sql
   ISNULL(DocCur, 'SAR') AS DocCur,
   ISNULL(DocRate, 1.0) AS DocRate,
   ```

2. **`models.go:799-814`** — Add fields to `SAPInvoice` struct:
   ```go
   DocCur   string  `json:"doc_cur"`
   DocRate  float64 `json:"doc_rate"`
   ```

3. **`models.go:846-900`** — Add `CurrencyCode` field to `CanonicalInvoice` struct (line 828) and populate in `ToCanonical()`:
   ```go
   CurrencyCode: inv.DocCur,
   ```
   Also add to `Metadata`:
   ```go
   "sap_doc_cur":   inv.DocCur,
   "sap_doc_rate":  inv.DocRate,
   ```

4. **`usecase/sap_migration.go:747-760`** — Add `currency_code` to the `INSERT INTO invoices` column list and the parameter list on line 753 and 760:
   - Add `currency_code` column after `customer_name` or wherever appropriate in column list.
   - Add `$n` parameter binding for `inv.CurrencyCode`.
   - Add `exchange_rate` parameter if `DocRate` is mapped.

5. **`85_zatca.sql:63`** — Keep the `DEFAULT 'USD'` as a safety net but ensure the usecase always supplies the value. Consider changing default to `'SAR'` to match source.

### Regression:
```sql
-- From 13-regression-verification.sql lines 171-174
SELECT currency_code, COUNT(*) AS invoice_count
FROM public.invoices
GROUP BY currency_code
ORDER BY invoice_count DESC, currency_code;
-- Expected after fix: SAR = 969,665 (100%)
```

#### A03 — Preserve cancellation status (High priority, 0.5 day)

1. **`tables.go:405-423`** — Add `CANCELED` and `DocStatus` (DocStatus already present on line 417):
   ```sql
   ISNULL(CANCELED, 'N') AS CANCELED,
   ISNULL(DocStatus, 'C') AS DocStatus,
   ```

2. **`models.go:799-814`** — Add `Canceled string` field.

3. **`models.go:846-855`** — Modify `ToCanonical()` status derivation:
   ```go
   if strings.ToUpper(inv.Canceled) == "Y" || strings.ToUpper(inv.Canceled) == "C" {
       invoiceStatus = "cancelled"
   } else if balanceDue > 0.01 {
       // existing logic
   }
   ```

### Regression:
```sql
-- From 13-regression-verification.sql lines 176-178
SELECT COUNT(*) AS canceled_status_count
FROM public.invoices
WHERE invoice_status = 'cancelled';
-- Expected after fix: 5,344
```

#### A04 — Service document type preservation (Medium priority, 0.5 day)

1. **`tables.go:405-423`** — Add `DocType`:
   ```sql
   ISNULL(DocType, 'I') AS DocType,
   ```

2. **`models.go:799-814`** — Add `DocType string` field.

3. **`models.go:882-883`** — Replace hardcoded `"standard"`:
   ```go
   InvoiceType: map[string]string{"I": "standard", "S": "service"}[inv.DocType],
   // or if DocType not in map, default "standard"
   invoiceType := "standard"
   if strings.ToUpper(inv.DocType) == "S" {
       invoiceType = "service"
   }
   ```
   - **Requires schema change:** Add `'service'` to the `invoice_type` ENUM in `85_zatca.sql:17`.

4. **`usecase/sap_migration.go:768-787`** — In the line INSERT, change hardcoded `'product'` on line 775 to use `line.ItemType` from canonical model (add `ItemType` field to `CanonicalInvoiceLine`).

5. **`models.go:816-826`** — Add `ItemType string` to `CanonicalInvoiceLine` and populate in `ToCanonical()` lines 859-873:
   ```go
   ItemType: "service",  // if parent DocType == "S" and ItemCode is empty
   ```

### Regression:
```sql
-- From 13-regression-verification.sql lines 107-110 (via exact lookup)
SELECT l.id, l.item_type, l.product_id, l.metadata
FROM public.invoice_lines l
JOIN public.invoices i ON i.id = l.invoice_id
WHERE i.metadata->>'sap_doc_entry' = '80306'
ORDER BY l.line_number;
-- After fix: service invoices should have invoice_type='service' and appropriate item_type
```

---

## Phase 2 — Invoice Line Corrections (A01, A06, A13)

**Owner:** Migration Developer + Finance stakeholder (A01, A13)
**Dependencies:** Phase 0 policy decision on line discount arithmetic (A01) and tax scope (A13)
**Code locations:**
- `packages/sap/schema/tables.go:425-439` — `QueryInvoiceLines`
- `packages/sap/mappings/models.go:786-797` — `SAPInvoiceLine` struct
- `packages/sap/mappings/models.go:816-826` — `CanonicalInvoiceLine` struct
- `packages/sap/mappings/models.go:857-873` — line mapping in `ToCanonical()`
- `packages/core/usecase/sap_migration.go:768-787` — invoice line INSERT
- `packages/core/db/schema/85_zatca.sql:124-163` — `invoice_lines` table DDL

### Changes:

#### A01 — Line discount evidence preservation (High priority, 1 day)

1. **`tables.go:425-439`** — Add `DiscPrcnt` and `PriceBefDi` to `QueryInvoiceLines`:
   ```sql
   ISNULL(DiscPrcnt, 0.0) AS DiscPrcnt,
   ISNULL(PriceBefDi, Price) AS PriceBefDi,
   ISNULL(GTotal, 0.0) AS GTotal,
   ```

2. **`models.go:786-797`** — Add fields to `SAPInvoiceLine` struct:
   ```go
   DiscPrcnt  float64 `json:"disc_prcnt"`
   PriceBefDi float64 `json:"price_bef_di"`
   ```

3. **`models.go:816-826`** — Add fields to `CanonicalInvoiceLine`:
   ```go
   DiscountPercent float64 `json:"discount_percent"`
   DiscountAmount  float64 `json:"discount_amount"`
   PriceBeforeDiscount float64 `json:"price_before_discount"`
   ```

4. **`models.go:857-873`** — Populate in `ToCanonical()`:
   ```go
   DiscountPercent:   l.DiscPrcnt,
   PriceBeforeDiscount: l.PriceBefDi,
   // Per policy: either derive or leave zero
   DiscountAmount:    0.0,  // Placeholder; actual derivation per Phase 0 policy
   ```
   Also add to metadata:
   ```go
   "sap_disc_prcnt":  l.DiscPrcnt,
   "sap_price_bef_di": l.PriceBefDi,
   ```

5. **`usecase/sap_migration.go:768-787`** — Add `discount_amount` to INSERT column list (line 769) and parameter binding (line 777):
   ```sql
   -- Add discount_amount to column list
   quantity, unit_price, discount_amount, tax_amount, line_total, metadata
   -- Add parameter
   $6, $9, 10, 11, 12, 13
   ```
   Reorder parameters accordingly in `execWithSavepoint` call on line 787.

6. **Per policy decision** (from Phase 0): If finance approves signed monetary derivation:
   ```go
   // Derived monetary discount = Quantity * PriceBefDi * (DiscPrcnt / 100)
   lineDiscount := l.Quantity * l.PriceBefDi * (l.DiscPrcnt / 100.0)
   ```

### Regression:
```sql
-- From 13-regression-verification.sql lines 51-59
SELECT COUNT(*) AS line_count,
       COUNT(*) FILTER (WHERE discount_amount = 0) AS zero_discount_count,
       COUNT(*) FILTER (WHERE discount_amount <> 0) AS nonzero_discount_count
FROM public.invoice_lines;
-- Expected after fix: nonzero_discount_count > 0 (if monetary derivation approved)
--                  OR all metadata contains sap_disc_prcnt (if raw evidence only)
```

#### A06 — Line UOM resolution (Medium priority, 0.5 day)

1. **`tables.go:425-439`** — `unitMsr` is already selected (line 436). No query change needed.

2. **`models.go:786-797`** — `UnitMsr` is already in the struct (line 796). No model change needed.

3. **`models.go:816-826`** — Add `UOMCode string` to `CanonicalInvoiceLine`.

4. **`models.go:857-873`** — Populate:
   ```go
   UOMCode: strings.TrimSpace(l.UnitMsr),
   ```
   Already stored in metadata as `sap_unit_msr` (line 871). **The gap is that `uom_id` is not resolved** in the target INSERT.

5. **`usecase/sap_migration.go:768-787`** — Add UOM lookup subquery:
   ```sql
   INSERT INTO invoice_lines (
       invoice_id, organization_id, line_number, description, item_type, product_id, product_sku,
       quantity, unit_price, discount_amount, tax_amount, line_total, uom_id, metadata
   )
   SELECT
       inv.id, $2, $3,
       COALESCE(NULLIF($4, ''), $5),
       'product',
       (SELECT id FROM products WHERE organization_id = $2 AND sku = $5 LIMIT 1),
       $5, $6, $9, 10, $11, $12,
       (SELECT id FROM units_of_measure WHERE code = $13 LIMIT 1),
       $14
   FROM invoices inv
   WHERE inv.invoice_number = $1 AND inv.organization_id = $2
   ON CONFLICT(invoice_id, line_number) DO UPDATE SET
       quantity = excluded.quantity,
       unit_price = excluded.unit_price,
       discount_amount = excluded.discount_amount,
       tax_amount = excluded.tax_amount,
       line_total = excluded.line_total,
       uom_id = excluded.uom_id,
       metadata = excluded.metadata;
   ```
   Add `line.UOMCode` as a new parameter.

### Regression:
```sql
-- From 13-regression-verification.sql lines 51-59
SELECT COUNT(*) FILTER (WHERE uom_id IS NULL) AS null_uom_count
FROM public.invoice_lines;
-- Expected after fix: null_uom_count should drop significantly from 5,472,485
```

#### A13 — Tax reference/policy documentation (Medium priority, 0.5 day)

This is primarily a documentation gap. The schema column `tax_category_id` already exists (`85_zatca.sql:151`) and `tax_rate` exists (`85_zatca.sql:152`). The query already omits tax reference fields.

1. **`tables.go:425-439`** — Optionally add `VatGourpSa` and `TaxCode` from `INV1`:
   ```sql
   ISNULL(VatGourpSa, '') AS VatGourpSa,
   ISNULL(TaxCode, '') AS TaxCode,
   ```

2. **Documentation decision** (Phase 0): If tax master migration is out of scope, document calculated-only behavior. If in scope, add extraction of `OSTC` (tax codes) and `OVTG` (tax groups) in a new domain.

3. **`models.go:816-826`** — Add `TaxCategory string` and `TaxRate float64` to `CanonicalInvoiceLine` if extracting.

4. **`usecase/sap_migration.go:768-787`** — Add `tax_category_id` and `tax_rate` to INSERT if populated.

### Regression:
```sql
-- From 13-regression-verification.sql lines 198-200
SELECT COUNT(*) AS tax_category_count,
       COUNT(*) FILTER (WHERE tax_rate IS NULL) AS null_tax_rate_count
FROM public.tax_categories;
-- Target: populate if in scope, or document exclusion if out of scope
```

---

## Phase 3 — Idempotent Replay Fixes (A07, A09)

**Owner:** Migration Developer + DB Engineer
**Dependencies:** Schema migration permissions for unique constraints
**Code locations:**
- `packages/core/usecase/sap_migration.go:412-429` — inventory insert
- `packages/core/usecase/sap_migration.go:660-682` — stock movement insert
- `packages/core/db/schema/40_inventory.sql:5-23` — `inventory_stock` table
- `packages/core/db/schema/40_inventory.sql:25-47` — `stock_movements` table

### Changes:

#### A07 — Inventory idempotency (High priority, 1 day)

1. **`40_inventory.sql`** — Add a unique constraint or index on `(organization_id, product_id, store_id)`:
   ```sql
   ALTER TABLE inventory_stock
   ADD CONSTRAINT inventory_stock_uk_organization_product_store
   UNIQUE (organization_id, product_id, store_id);
   ```
   *Note: This is a schema migration that must be applied to target databases. The audit database `nembus_migration_audit_20260922` is read-only and cannot be tested against; a disposable target is required.*

2. **`usecase/sap_migration.go:412-429`** — Replace plain `INSERT ... LIMIT 1` with `INSERT ... ON CONFLICT`:
   ```sql
   INSERT INTO inventory_stock (product_id, store_id, quantity_on_hand, quantity_allocated, quantity_available, quantity_on_order, reorder_level, max_stock_level, metadata)
   SELECT p.id, s.id, $3, $4, $5, $6, $7, $8, $9
   FROM products p
   CROSS JOIN stores s
   WHERE p.organization_id = $1 AND p.sku = $2 AND s.organization_id = $1 AND s.code = $10
   ON CONFLICT (organization_id, product_id, store_id) DO UPDATE SET
       quantity_on_hand = excluded.quantity_on_hand,
       quantity_allocated = excluded.quantity_allocated,
       quantity_available = excluded.quantity_available,
       quantity_on_order = excluded.quantity_on_order,
       reorder_level = excluded.reorder_level,
       max_stock_level = excluded.max_stock_level,
       metadata = excluded.metadata;
   ```
   **Important:** The current `SELECT ... LIMIT 1` subquery approach will not support `ON CONFLICT` because the subquery doesn't lock the row. Refactor to use `VALUES (...)` instead of `SELECT` for the insert, with product/store IDs resolved as parameters first.

3. **Refactor approach:** Instead of `SELECT ... FROM products CROSS JOIN stores ... LIMIT 1`, resolve `product_id` and `store_id` as subqueries in Go first, then use `INSERT ... VALUES (...) ON CONFLICT ... DO UPDATE`.

### Regression:
```sql
-- From 13-regression-verification.sql lines 203-216
SELECT COUNT(*) AS inventory_rows,
       COUNT(DISTINCT ...) AS distinct_source_pairs
FROM public.inventory_stock;
-- Run inventory batch twice in a disposable target; assert:
-- 1. One row per source key after second run
-- 2. 151,263 excess rows = 0
```

#### A09 — Stock movement idempotency (High priority, 1 day)

1. **`40_inventory.sql`** — Add a unique index on the source transaction number:
   ```sql
   -- Option A: Use metadata JSON path (PostgreSQL 18 supports expression indexes)
   CREATE UNIQUE INDEX CONCURRENTLY stock_movements_sap_trans_num_uk
   ON stock_movements (organization_id, (metadata->>'sap_trans_num'))
   WHERE metadata ? 'sap_trans_num';
   
   -- Option B (preferred): Add a dedicated external_id column
   ALTER TABLE stock_movements ADD COLUMN external_id VARCHAR(100);
   CREATE UNIQUE INDEX CONCURRENTLY stock_movements_external_id_uk
   ON stock_movements (organization_id, external_id)
   WHERE external_id IS NOT NULL;
   ```

2. **`usecase/sap_migration.go:660-682`** — Add `ON CONFLICT DO NOTHING` (or `DO UPDATE`):
   ```sql
   INSERT INTO stock_movements (...)
   SELECT ...
   ON CONFLICT DO NOTHING;
   ```
   *Same refactor concern as inventory: the `SELECT ... WHERE ... IS NOT NULL` subquery pattern needs conversion to `VALUES` with pre-resolved IDs.*

### Regression:
```sql
-- From 13-regression-verification.sql lines 219-235
SELECT COUNT(*) AS movement_rows,
       COUNT(DISTINCT metadata->>'sap_trans_num') AS distinct_source_trans_nums,
       COUNT(*) - COUNT(DISTINCT metadata->>'sap_trans_num') AS excess_duplicate_rows
FROM public.stock_movements;
-- Run stock movement batch twice; assert excess_duplicate_rows = 0
```

---

## Phase 4 — Procurement Link Repair (A08)

**Owner:** Migration Developer
**Code locations:**
- `packages/sap/mappings/models.go:1153` — PO number construction (`PO-%d` with `DocNum`)
- `packages/sap/mappings/models.go:1251-1253` — GRN PO number construction (`PO-%d` with `BaseEntry`)
- `packages/core/usecase/sap_migration.go:614-615` — GRN PO lookup by `po_number`
- `packages/core/usecase/sap_migration.go:549-552` — PO insert (uses `po_number`)

### Root Cause:
- PO mapper (line 1153): `fmt.Sprintf("PO-%d", po.DocNum)` — uses SAP `DocNum` (e.g., `PO-12345`)
- GRN mapper (line 1252): `fmt.Sprintf("PO-%d", l.BaseEntry)` — uses SAP `BaseEntry` which is a `DocEntry` (internal ID), NOT a `DocNum`
- Result: GRN looks up `PO-{BaseEntry}` which never matches `PO-{DocNum}` → all `purchase_order_id` are NULL

### Changes:

#### A08 — Fix PO/GRN key consistency (Medium priority, 1 day)

1. **`models.go:1251-1253`** — The GRN mapper must resolve `BaseEntry` (a PO DocEntry) to the PO's `DocNum` before constructing the lookup key. Two approaches:

   **Approach A (recommended): Store the source PO DocEntry in GRN line metadata and look up by it.**
   
   a. In `SAPGoodsReceiptLine` struct (line 1179), `BaseEntry` and `BaseType` are already present (lines 1190-1192).
   
   b. In `ToCanonical()` (line 1237), store `BaseEntry` in item metadata (already done at line 1267: `"sap_base_entry": l.BaseEntry`).
   
   c. **Fix the PO number construction:** Instead of using `BaseEntry` as a `DocNum`, the GRN should pass the `BaseEntry` (PO DocEntry) and look up the PO by `metadata->>'sap_doc_entry'`:
   
   ```go
   // In CanonicalGoodsReceiptItem, add:
   SourceEntry int64 `json:"source_entry"`  // l.BaseEntry (PO DocEntry)
   ```

   d. **`usecase/sap_migration.go:614-615`** — Change the PO lookup from `po_number = $3` to a metadata-based lookup:
   ```sql
   (SELECT id FROM purchase_orders 
    WHERE organization_id = $1 AND metadata->>'sap_doc_entry' = $X LIMIT 1)
   ```
   where `$X` is the PO's `DocEntry` (which equals the GRN line's `BaseEntry` when `BaseType == 22`).

   e. **`usecase/sap_migration.go:646`** — Same change for line-level PO line lookup: instead of matching `po.po_number = $3 AND pol.line_number = $4`, match `po.metadata->>'sap_doc_entry' = $X AND pol.line_number = $4`.

2. **Alternative fix:** If the PO mapper stores `DocNum` in metadata, the GRN mapper can look up POs by `DocNum`:
   - Store `po.DocNum` in `CanonicalPurchaseOrder.Metadata` (check if already done).
   - GRN mapper resolves `BaseEntry` → PO `DocNum` via a source-side join during extraction.

### Regression:
```sql
-- From 13-regression-verification.sql lines 238-244
SELECT COUNT(*) AS grn_count,
       COUNT(*) FILTER (WHERE purchase_order_id IS NULL) AS null_purchase_order_count
FROM public.goods_receipt_notes;
-- Expected after fix: null_purchase_order_count for BaseType=22 GRNs = 0
-- (15 headers, 248 lines should resolve)
```

---

## Phase 5 — Payment Detail Delivery (A05)

**Owner:** Migration Developer
**Dependencies:** Phase 0 policy decision on payment scope
**Code locations:**
- `packages/sap/schema/tables.go:254-286` — payment extraction queries
- `packages/sap/mappings/models.go:1419-1515` — `SAPIncomingPayment` and `CanonicalIncomingPayment` structs
- `packages/core/usecase/sap_migration.go:792-867` — `DomainIncomingPayments` handler
- `apps/sap-agent/internal/etl/engine.go:35-41` — domain list (already includes `contracts.DomainIncomingPayments`)

### Changes:

#### A05 — Activate incoming payments domain (High priority, 1–2 days)

The code already exists (mapper, usecase, domain list). The gap is that the payments domain was not included in the historical migration run.

1. **`tables.go:254-286`** — Verify the `QueryIncomingPayments` and `QueryPaymentAllocations` queries are correct and select all required fields: `DocEntry, DocNum, DocDate, CardCode, CardName, DocCurr, DocTotal, CashSum, TrsfrSum, TrsfrRef, CheckSum, CreditSum, Comments, JrnlMemo` from `ORCT` and `DocEntry, DocNum, InvType, SumApplied, InvoiceDocNum, DocEntry` from `RCT2`.

2. **`models.go:1461-1515`** — Verify `ToCanonicalList()` correctly expands each payment into per-invoice allocation records. Check the `paymentMethod` derivation logic (cash/card/bank_transfer/check/credit).

3. **`usecase/sap_migration.go:792-867`** — Verify the INSERT and the invoice balance recalculation UPDATE. The `payment_number` uniqueness constraint needs verification — if multiple payments have the same `DocNum`, the `ON CONFLICT(payment_number)` on line 822 could merge allocations incorrectly.

4. **`engine.go:35-41`** — Confirm `contracts.DomainIncomingPayments` is in the `defaultDomains` list (line 40 confirms it is). The domain was enabled in code but not executed in the run that produced the audit backup. **This is a run-configuration gap, not a code gap.**

5. **Run configuration fix (A11):** Ensure `invoice_start_date` and other config parameters are recorded in `staging.sap_migration_batches` so the same run can be reproduced.

### Regression:
```sql
-- From 13-regression-verification.sql lines 252-253
SELECT COUNT(*) AS payment_count
FROM public.invoice_payments;
-- Expected after re-run with payments domain: > 0 (should be ~961,745 or deduplicated)
```

---

## Phase 6 — Partner Address & Documentation Gaps (A10, A11, A12)

**Owner:** Migration Developer + Operations
**Code locations:**
- `packages/sap/schema/tables.go` — CRD1 query (check if exists)
- `packages/sap/mappings/models.go:1004-1028` — `SAPBPAddress.ToCanonical()` (code exists)
- `packages/core/usecase/sap_migration.go:477-528` — `DomainBPAddresses` handler (code exists)
- `apps/sap-agent/internal/etl/engine.go:20-41` — domain list (check if `DomainBPAddresses` is listed)
- `apps/sap-agent/internal/etl/engine.go:792-864` — invoice date window config

### Changes:

#### A10 — Activate partner addresses (Low priority, 0.5 day)

The mapping and usecase code already exist. The domain may not be in the default ETL domain list.

1. Check `engine.go:20-41` for `contracts.DomainBPAddresses` in the domain list. If missing, add it.
2. Verify `tables.go` has a `QueryBPAddresses` query for `CRD1`. If not, add:
   ```sql
   SELECT CardCode, Address, AdresType, Street, City, State, Country, ZipCode, Phone1, Phone2
   FROM CRD1
   ```

### Regression:
```sql
-- From 13-regression-verification.sql line 281
SELECT COUNT(*) AS partner_address_count FROM public.partner_addresses;
-- Expected: 1 (matching source CRD1 row count)
```

#### A11 — Run configuration reproducibility (Medium priority, 0.5 day)

1. **`engine.go:792-864`** — Identify where `invoice_start_date` is consumed.
2. **`95_sap_staging.sql`** — Add a column to `sap_migration_batches` or create a `migration_run_config` table:
   ```sql
   ALTER TABLE staging.sap_migration_batches
   ADD COLUMN effective_config JSONB;
   ```
3. At the start of each migration run, write the effective config (start/end dates, domain list, batch size, etc.) into this column.

### Regression:
```sql
-- Verify all 315,715 pre-2023 invoices are accounted for in the recorded config
SELECT effective_config->>'invoice_start_date'
FROM staging.sap_migration_batches
WHERE domain = 'invoices'
LIMIT 1;
-- Cross-check with invoice date distribution query
```

#### A12 — Batch error observability (Medium priority, 0.5 day)

1. **`95_sap_staging.sql:8-19`** — Check `sap_migration_batches` table schema for `error_message` column.
2. **`usecase/sap_migration.go`** — Find where failed batches are recorded. Ensure `error_message` is populated with structured error details (not empty string) for all `status = 'failed'` batches.
3. Add batch-level error logging that captures the specific row errors (from `errs` slice) and persists them.

### Regression:
```sql
-- From 13-regression-verification.sql lines 267-274
SELECT domain, status, COUNT(*) AS batch_count,
       SUM(record_count) AS record_count_sum,
       COUNT(*) FILTER (WHERE COALESCE(error_message, '') = '') AS empty_error_message_count
FROM staging.sap_migration_batches
GROUP BY domain, status;
-- Expected after fix: empty_error_message_count = 0 for failed batches
```

---

## Phase 7 — Return Credit Notes (A15) — Future Work

**Owner:** Business stakeholder (scope decision) + Migration Developer (implementation)
**Status:** POLICY_REQUIRED — blocked on business scope decision

If returns/credit notes (`ORIN`/`RIN1`) are approved as in scope:
1. Add `DomainSalesReturns` to the ETL domain list in `engine.go`.
2. Create `QueryCreditNotes` in `tables.go` for `ORIN`/`RIN1`.
3. Add `SAPSalesReturn` struct and `CanonicalSalesReturn` in `models.go`.
4. Add `DomainSalesReturns` case in `sap_migration.go`.
5. Verify schema for `public.sales_returns` / `public.sales_return_lines` (`85_zatca.sql:100-122`).

If excluded: document in the migration contract and add a scope-exclusion marker in staging.

### Regression:
```sql
-- From 13-regression-verification.sql lines 279-280
SELECT COUNT(*) AS return_count FROM public.sales_returns;
SELECT COUNT(*) AS return_line_count FROM public.sales_return_lines;
-- Expected: > 0 if in scope; 0 with documented exclusion if out of scope
```

---

## Phase 8 — Validation & Regression

**Owner:** Audit team (read-only validation) + Migration Developer
**Dependencies:** All preceding phases

1. Generate a new isolated target database from a fresh migration run with all fixes applied.
2. Do NOT restore into `nembus_migration_audit_20260922` (immutable evidence baseline).
3. Run the full `13-regression-verification.sql` against the new target.
4. For idempotency fixes (A07, A09): run the affected batches twice in the new target and verify no excess rows.
5. For all data mismatch fixes: verify against the expected values documented in this plan.

### Regression execution matrix:

| Finding | Verification Query (from 13-regression-verification.sql) | Expected Result |
|---|---|---|
| A01 line discount | Lines 51-59 (nonzero discount count) | > 0 or metadata has sap_disc_prcnt |
| A02 currency | Lines 171-174 | SAR = 969,665 |
| A03 cancellation | Lines 166-178 | cancelled = 5,344 |
| A04 service docs | Lines 107-110 (exact lookup) | service line item_type preserved |
| A05 payments | Line 252 | > 0 |
| A06 UOM | Lines 51-59 (null_uom_count) | < 5,472,485 |
| A07 inventory | Lines 203-216 | distinct pairs = 7,203, rows = 7,203 |
| A08 GRN-PO links | Lines 238-244 | null PO links = 0 for BaseType=22 |
| A09 stock movements | Lines 219-235 | excess duplicates = 0 |
| A10 addresses | Line 281 | = 1 |
| A13 tax references | Lines 198-200 | populated if in scope, documented if not |

---

## Cross-reference to Action Register (12-developer-action-register.csv)

All 15 actions (A01–A15) are mapped to the phases above. A14 (header discounts) requires NO remediation — the mapping is correct and validated.

## Summary of Files to Modify

| File | Phases | Changes |
|---|---|---|
| `packages/sap/schema/tables.go` | P1, P2 | Add DocCur, CANCELED, DocType, DiscPrcnt, PriceBefDi to queries (lines 405-439) |
| `packages/sap/mappings/models.go` | P1, P2 | Add fields to SAPInvoice/SAPInvoiceLine/CanonicalInvoice/CanonicalInvoiceLine structs; update ToCanonical() methods |
| `packages/core/usecase/sap_migration.go` | P1, P2, P3, P4, P5 | Update INSERT statements for invoices, invoice_lines, inventory_stock, stock_movements, goods_receipt_notes |
| `packages/core/db/schema/85_zatca.sql` | P1 | Add 'service' to invoice_type ENUM |
| `packages/core/db/schema/40_inventory.sql` | P3 | Add unique constraints for idempotent upsert |
| `apps/sap-agent/internal/etl/engine.go` | P5, P6, A11 | Verify domain list (lines 20-41); record run config (lines 792-864) |
| `packages/core/db/schema/95_sap_staging.sql` | A11, A12 | Add effective_config column; ensure error_message is populated |
