# Invoice discount and total reconciliation

## Corrected conclusion after Sweep 2

The reported `-23` is SAP `INV1.DiscPrcnt` on `DocEntry=80306`, `LineNum=5`; it is a signed percentage, not a target currency amount. The source row is `DocNum=2022025735`, `ItemCode=INV04287`, `Quantity=0.325`, `PriceBefDi=16.96`, `Price=20.86`, `DiscPrcnt=-23`, `LineTotal=6.78`, `VatSum=1.02`, and `GTotal=7.80`.

Sweep 2 corrected an important preliminary conclusion: the migration does select `OINV.DiscSum` and maps it to the target header `invoices.discount_amount`. Source and target header discount aggregates match exactly. The omission is at line level: `INV1.DiscPrcnt` and `PriceBefDi` are not selected or persisted, and every restored target `invoice_lines.discount_amount` is `0.00`. Source line currency, cancellation/type semantics, and line UOM/reference fields are separate gaps.

## Runtime path

```text
SQL Server Qadsiya_Dev
  -> packages/sap/schema/tables.go (OINV/INV1 SELECTs)
  -> packages/sap/mappings/models.go (SAP -> canonical mapping)
  -> packages/core/usecase/sap_migration.go (PostgreSQL inserts)
  -> invoices / invoice_lines / invoice_payments
```

The relevant source queries are `packages/sap/schema/tables.go:405-439`. The canonical header carries `DiscSum` (`packages/sap/mappings/models.go:723-865`), and the target invoice insert supplies `discount_amount` (`packages/core/usecase/sap_migration.go:747-760`). The canonical line has no discount field and the line insert does not supply `discount_amount`, `uom_id`, `tax_category_id`, or `tax_rate` (`packages/core/usecase/sap_migration.go:768-787`).

## Target semantics and exact restored row

The target trigger calculates invoice subtotal from `SUM(line_total - tax_amount)` and total as subtotal plus tax, less the header `discount_amount`. A negative monetary header discount would therefore increase the target total; SAP signed values need an approved business policy before any new derived field is introduced.

For source invoice `80306`, the restored target row is:

| Field | Source | Target |
|---|---:|---:|
| document number | `2022025735` | `INV-SAP-2022025735` |
| date | `2022-02-11` | `2022-02-11` |
| currency | `SAR` | `USD` |
| header subtotal | derived `76.78` | `76.78` |
| header discount | `0.00` | `0.00` |
| tax | `11.52` | `11.52` |
| total | `88.30` | `88.30` |
| line 5 quantity | `0.325` | `0.325` |
| line 5 unit price | `20.86` | `20.86` |
| line 5 discount | raw `-23%` | `0.00` |
| line 5 tax | `1.02` | `1.02` |
| line 5 total | `7.80` gross | `7.80` |

The target line total is intentionally the SAP gross `GTotal` semantic (`LineTotal + VatSum`), not the net `LineTotal` alone. The exact row therefore is not evidence of a bad total; it is evidence that the signed line percentage was omitted.

## Source distributions

`OINV` contains 969,665 rows dated 2021-09-08 through 2025-05-08. Header `DiscSum` is positive on 23,551 rows, zero on 911,017, and negative on 35,097; min `-0.44`, max `511.14`, sum `-725.45`. Header `DiscPrcnt` is positive on 20,725 rows, zero on 920,152, and negative on 28,788.

`INV1` contains 5,472,488 rows, of which 5,472,485 join to an invoice header and are eligible for target-line comparison. Among eligible rows, `DiscPrcnt` is positive on 814,610, zero on 3,280,140, and negative on 1,377,735; range `-4,685` to `100`. The three orphan source lines have zero `DiscPrcnt`.

Source header currencies are all `SAR`. Header `DocType` is `I` for 969,634 rows and `S` for 31; `CANCELED` is `N` for 964,321, `Y` for 2,672, and `C` for 2,672.

## Target reconciliation results

- Target invoices: `969,665`; target invoice lines: `5,472,485`.
- All target invoice lines have unique `(invoice_id,line_number)` and unique `(sap_doc_entry,sap_line_num)` metadata pairs.
- All `5,472,485` target line discounts are zero; source has `2,192,345` eligible lines with nonzero `DiscPrcnt`.
- Target line totals sum to `65,447,356.09`, matching eligible source `LineTotal + VatSum`; target tax sums to `8,536,900.76`, and target net reconstruction sums to `56,910,455.33`.
- Across all 969,665 target invoices, subtotal, tax, and total differences against the target trigger formula are zero; max absolute difference is `0.00`.
- Target header discount distribution and weighted aggregates match source exactly: positive `23,551`, zero `911,017`, negative `35,097`, null `0`, min `-0.44`, max `511.14`, sum `-725.45`; weighted sum `234,975,426.96`; total sum `65,448,081.54`; tax sum `8,536,900.76`; subtotal sum `56,910,455.33`.
- Target currency is `USD` for all 969,665 invoices although `SAR` exists in `public.currencies` and is the source currency.
- Target `uom_id`, `tax_category_id`, and `tax_rate` are NULL on all target invoice lines. Source `UnitMsr` is nonblank on 5,472,454 eligible lines.

## Arithmetic interpretation of the reported `-23`

For the exact source line, `0.325 * 16.96 = 5.512`. Applying the signed source percentage gives `16.96 * (1 - (-23/100)) = 20.8608`, stored as `20.86`; `0.325 * 20.86 = 6.7795`, stored as `6.78`. Under a gross-minus-discount convention, the signed adjustment is approximately `5.512 - 6.78 = -1.268`, which behaves as a surcharge. It must not be copied blindly into a positive target discount field.

## Classification

- **Confirmed data mismatch/code defect:** line discount evidence is omitted. The target has no nonzero line discounts despite 2,192,345 eligible source lines with nonzero percentages.
- **Confirmed correct:** header `DiscSum` is mapped and matches source aggregates exactly; target invoice totals and tax reconcile to stored line values.
- **Confirmed data mismatch/code defect:** source `SAR` invoice currency is not mapped; all target invoice rows are `USD`.
- **Confirmed data mismatch/code defect:** line UOM and tax-reference fields are omitted, although calculated tax amounts are correct.
- **Confirmed source condition:** signed negative header and line values occur at scale and reconcile with SAP arithmetic; sign handling requires business policy.
- **Out of scope:** `ORIN/RIN1` credit notes/returns are populated (`1,387`/`7,704`) but have no implemented invoice migration path.

The developer action register and the SELECT-only regression script contain the repeatable checks and the required remediation decisions.
