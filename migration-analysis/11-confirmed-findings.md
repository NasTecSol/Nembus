# Sweep 2 confirmed findings

## Restore disposition

The authorized isolated restore succeeded on the existing PostgreSQL 18 service. The new database `nembus_migration_audit_20260922` is present and database-level read-only. All findings below are based on read-only target queries plus the previously captured read-only SAP source results.

## Confirmed data mismatches and code defects

| ID | Severity | Finding | Evidence | Action |
|---|---|---|---|---|
| S2-F001 | High | Invoice line discount evidence is omitted | Source has 2,192,345 eligible lines with nonzero `DiscPrcnt`; target has 5,472,485 lines and zero nonzero `discount_amount` values. The exact `-23%` row is target `0.00`. | Preserve raw percentage/price inputs and implement an approved SAP signed/rounding contract. |
| S2-F002 | High | Invoice currency is defaulted incorrectly | All 969,665 source invoice headers are `SAR`; all 969,665 target invoice headers are `USD`. `SAR` exists in target currency master. | Map `DocCur` explicitly and validate totals by currency. |
| S2-F003 | High | Cancellation semantics are lost | Source has `CANCELED=Y` 2,672 and `CANCELED=C` 2,672; target has zero `cancelled` invoices. | Define load/filter/status policy and apply source cancellation fields. |
| S2-F004 | Medium | Service-document semantics are lost | Source has 31 `DocType=S` headers; target invoice type is `standard` for all 969,665 and the 31 service lines are `item_type=product` with NULL product IDs. | Carry source document type or explicitly document/quarantine service handling. |
| S2-F005 | High | Payment detail is absent | Source has 961,745 positive AR invoice allocations across 1,329 payment documents; target `invoice_payments` is empty and no incoming-payment staging domain is present. | Include the payment domain or explicitly separate header paid-to-date from payment-detail scope. |
| S2-F006 | Medium | Invoice line UOM is omitted | Source `UnitMsr` is nonblank on 5,472,454 eligible lines; target `uom_id` is NULL on all 5,472,485 lines. | Resolve line UOM to target units or retain an auditable source UOM field. |
| S2-F007 | High | Inventory batches are not idempotent | Source eligible OITW has 7,203 product/store keys; target has 158,466 rows, exactly 22 rows per key; excess rows are 151,263. | Add a stable source-key uniqueness/upsert policy before replay. |
| S2-F008 | Medium | GRN-to-PO links are not resolved | Source has 15 PO-linked GRN headers and 248 linked lines (`BaseType=22`); target has NULL `purchase_order_id` on all 7,128 GRNs and NULL `purchase_order_line_id` on all 73,639 items. | Use the same PO key on both purchase and receipt paths; quarantine unresolved base links. |
| S2-F009 | High | Stock movement batches are not idempotent | Source eligible OINM slice has 2,163,651 rows; target has 2,710,151 rows and 2,163,651 distinct `sap_trans_num` keys, leaving 546,500 excess duplicates. | Add a stable movement-key uniqueness/upsert policy and replay test. |
| S2-F010 | Low | Partner address is not loaded | Source CRD1 has one address row; target `partner_addresses` has zero rows and no address staging domain is present. | Include the address domain or record it as an explicit scope exclusion. |

## Documentation and scope gaps

| ID | Severity | Finding | Evidence |
|---|---|---|---|
| S2-F011 | Medium | Historical run configuration is not reproducible | The current config says `invoice_start_date=2023-01-01`, but the restored artifact contains 315,715 invoices dated before 2023. The actual historical run configuration is not recorded. |
| S2-F012 | Medium | Failed invoice batch diagnostics are insufficient | Staging shows 366 failed invoice batches with `record_count_sum=182,686` and empty error text. Target invoice count still equals source invoice count, so this is an auditability gap rather than proof of lost rows. |
| S2-F013 | Medium | Tax reference/master migration is undocumented | Target calculated tax amounts reconcile exactly, but `tax_categories` is empty and all line `tax_category_id`/`tax_rate` values are NULL. |

## Confirmed correct or source-conditioned areas

- Header `OINV.DiscSum` is selected, mapped, and matches source/target distributions and weighted aggregates exactly.
- Invoice subtotal, tax, and total reconciliations are exact across all 969,665 target invoices; maximum target trigger difference is `0.00`.
- Products (16,955), categories (27), customers (14), suppliers (387), stores (2), purchase orders (99/1,755), goods receipts (7,128/73,639), and sales orders (1/1) match their loaded source slices by count and source keys unless a specific relationship issue is listed above.
- Barcode duplicates are normalized to 27,445 distinct target barcodes; source raw OBCD has 27,591 rows and 27,445 distinct codes.
- The absence of brands is source-conditioned: OMRG is absent, and the target has zero brands.
- Returns/credit notes (`ORIN/RIN1` = 1,387/7,704) remain outside the implemented migration path.
- Signed negative `DiscSum` and `DiscPrcnt` values are source conditions requiring policy; they are not evidence that values should be absolute-valued or blindly sign-flipped.

## Key implementation locations

- Invoice source fields: `packages/sap/schema/tables.go:405-439`.
- Header `DiscSum` mapping: `packages/sap/mappings/models.go:723-865`.
- Invoice insert and line insert: `packages/core/usecase/sap_migration.go:747-787`.
- Inventory insert without conflict key: `packages/core/usecase/sap_migration.go:412-429`.
- Stock movement insert without conflict key: `packages/core/usecase/sap_migration.go:664-682`.
- Purchase/GRN key construction and inserts: `packages/sap/mappings/models.go:1153,1258-1276`; `packages/core/usecase/sap_migration.go:543-655`.

The detailed row-level evidence is in `08-source-target-parity-summary.csv`, `09-invoice-discount-reconciliation.csv`, and `10-currency-reconciliation.csv`.
