# Nembus — SAP Benchmark Gap Analysis & Enhancement Roadmap

**Product under review:** Nembus (cloud ERP + POS monorepo: `packages/core` domain/backend, `apps/cloud-server` multi-tenant API, `apps/pos-client` offline-first desktop POS).
**Benchmarks:** SAP Business One (incl. SAP Business One on HANA) and SAP S/4HANA "standards".
**Lens:** functional parity, market trends, and engineering standards for SMB ERP/POS software in 2024–2026.
**Audience:** Nembus architects, product owners, and engineering leads planning the next schema/feature releases.

---

## Table of Contents

1. [Executive summary](#1-executive-summary)
2. [Scope, method & benchmarks](#2-scope-method-benchmarks)
3. [Nembus today — inventory & architecture](#3-nembus-today--inventory--architecture)
4. [Domain-by-domain gap matrix](#4-domain-by-domain-gap-matrix)
5. [Priority deep-dive A — Finance & accounting (GL, AR/AP, banking, fixed assets)](#5-priority-deep-dive-a-finance-accounting)
6. [Priority deep-dive B — Commerce & omnichannel](#6-priority-deep-dive-b-commerce-omnichannel)
7. [Priority deep-dive C — Operations (production/MRP, costing, landed costs)](#7-priority-deep-dive-c-operations)
8. [Platform standards & market trends (S/4HANA-derived)](#8-platform-standards-market-trends-s4hana-derived)
9. [Prioritized roadmap](#9-prioritized-roadmap)
10. [Risks & architectural considerations](#10-risks-architectural-considerations)
11. [Appendix A — Nembus schema inventory](#appendix-a-nembus-schema-inventory)
12. [Appendix B — SAP ↔ Nembus concept mapping](#appendix-b-sap-nembus-concept-mapping)
13. [Appendix C — References](#appendix-c-references)

---

## 1. Executive summary

Nembus is a genuinely modern, clean-architecture SMB ERP/POS system: Go + PostgreSQL + multi-tenant-by-database, offline-first POS with a transactional-outbox sync engine, an unusually complete omni-channel sales stack (carts, orders v2, fulfillments, invoices, quotes, returns), restaurant operations, recipe/BOM costing, promotions/loyalty, KSA ZATCA e-invoicing, and pre-aggregated analytics. On the commerce side it is arguably ahead of SAP Business One's default footprint.

However, measured against SAP Business One (the SMB ERP benchmark) and the data-model standards SAP S/4HANA established, Nembus has four structural gaps:

1. **No operational finance layer.** GL primitives exist (`chart_of_accounts`, `cost_centers`, `journal_entries/lines`, `gl_account_mappings`) but there is no posting engine, no posting periods / fiscal years, no document-numbering series, no double-entry postings generated from POS/GRN/orders, no AR/AP subledgers with payment allocation, and no financial statements. Every commercial document is schema-valid but never "posted to the books". This is the largest functional gap vs SAP Business One and the single biggest obstacle to being sold as an ERP rather than a POS-with-extras.
2. **Document-lifecycle mechanics are missing.** Approval procedures/workflows, centralized numbering series, draft-vs-posted document states, alerts, dunning/collections, and credit management are either ad hoc (`approved_by` columns) or absent.
3. **Costing, valuation and manufacturing are thin.** There is no costing-method engine (moving average / FIFO / standard / specific), no landed-cost allocation, no goods-issue / production-consumption pipeline, no production orders or BOM-driven backflush (recipes exist but are restaurant recipes), and therefore no true COGS per transaction or inventory valuation report.
4. **Platform/analytics positioning vs "on HANA".** Analytics are nightly pre-aggregated tables + SQL views. There is no semantic/OLAP layer, no self-service BI or KPI cockpit, no real-time (in-memory/columnar) reporting story, and only nascent AI. Underlying event data is largely there, but it is not exposed as a modern embedded-analytics or event-driven platform (webhooks, CDC/outbox to analytics, semantic model).

The recommended answer is not "build SAP B1": it is to adopt SAP's *underlying standards* — the Universal Journal idea (single append-only financial line store), Business Partner master with customer+vendor roles, clean posting periods, document numbering, approval workflows, user-defined fields, an item-valuation ledger — while keeping Nembus's modern commerce/omnichannel strengths. Sections 5–9 translate this into concrete schema/feature designs and a phased roadmap (P0 = financial backbone; P1 = operations & commerce depth; P2 = platform/analytics/AI).

---

## 2. Scope, method & benchmarks

### 2.1 What was compared

- **Nembus as built in this repository** — schema source of truth under `packages/core/db/schema/*.sql` (plus POS-client extensions), sqlc query layer, use-case/handler/routing surface, cloud-server and POS-client apps. Both implemented (API-backed) features and schema-only tables were catalogued, because schema-only tables are latent capability, not delivered capability.
- **SAP Business One (B1)** — SAP's ERP for small/mid-size companies; SMB functional baseline.
- **SAP Business One on HANA** — the in-memory database edition; platform/analytics baseline.
- **SAP S/4HANA** — used not as a feature checklist (enterprise scope, not realistic for SMB) but as the source of *data-model standards* and *product-direction signals* worth cherry-picking (Universal Journal, business partner CVI, posting periods, extensibility/clean-core, embedded analytics, AI).
- **Market trends & standards 2024–2026** — e-invoicing mandates, omnichannel order management, subscription/recurring revenue, AI copilots, event-driven integration, security/identity, sustainability reporting, etc.

### 2.2 Method

1. Full read of the canonical schema (`10` domain files, ~5.9k lines) plus POS-client/cloud-server migration extensions and the restaurant/ZATCA modules.
2. Cross-reference of schema objects against the implemented API surface (`packages/core/handler`, `usecase`, `routing`, sqlc queries) to classify **delivered** vs **schema-only** features.
3. Keyword/coverage scans for SAP-class concepts (landed cost, dunning, depreciation, posting periods, approval workflows, UDF, credit/debit memos, withholding, production orders, MRP, etc.) to make the gap evidence precise.
4. Desktop research on current SAP B1 / B1-on-HANA releases and S/4HANA standards and ERP market trends, with sources in [Appendix C](#appendix-c-references).
5. Scoring: each domain row gives a **gap severity** (None / Low / Medium / High / Critical) plus evidence and recommended direction. Severity weighs *delivered business value and adoption blockers for the SMB segment Nembus targets (retail / restaurant / wholesale / light distribution)*, not raw feature-count distance.

### 2.3 Reading the tables

Throughout the document:

- **Nembus today** names concrete schema tables and code surfaces (verified against the repo).
- **SAP B1** is the SMB functional baseline.
- **S/4HANA standard** names the underlying concept Nembus should borrow, even when full S/4HANA machinery is overkill.
- **Recommended Nembus action** is a concrete, conventionally-named schema/feature proposal.

---

## 3. Nembus today — inventory & architecture

### 3.1 Architecture (verified from repo)

| Layer | Nembus choice |
|---|---|
| Language / web | Go, Gin (REST/JSON) |
| Data | PostgreSQL; schema managed by **Atlas** versioned migrations (`packages/core/db/migrations`), POS client adds a **Goose** migration (`apps/pos-client/migrations`) |
| Query layer | sqlc-generated repositories |
| Business layer | Clean architecture: `handler → usecase → repository` |
| Multi-tenancy | Master DB holding `tenants` (each with `db_conn_str`); per-tenant databases with their own `organizations`/data; JWT + tenant middleware; `apps/cloud-server/cmd` tenant tooling (migrate/check/backup) |
| Clients | `apps/cloud-server` (REST + gRPC backup), `apps/pos-client` (Wails desktop + Angular frontend, embedded Postgres, offline-first) |
| Offline sync | POS-local transactional **outbox** (`sync_queue` with triggers on `pos_transactions/lines/payments/cashier_sessions`), delta sync w/ `sync_watermarks`, local device config |
| E-invoicing | KSA ZATCA Phase-2 device configs, document chain, XML + crypto signing in `usecase/zatca_*` |
| Analytics | Nightly pre-aggregated fact tables + SQL views/functions in `90_views_functions.sql` |

### 3.2 Domain coverage map

Canonical schema lives in `packages/core/db/schema/` (10 files, ~110 tables incl. joins) + POS extensions (`sync_queue`, `local_printer_configs`, `local_device_config`).

| Domain (schema file) | Key tables | Delivered API? |
|---|---|---|
| Identity & RBAC (`10_identity_rbac`) | `organizations`, `tenants`, `modules/menus/submenus`, `permissions`, `roles`, `role_permissions`, `ui_settings`, `role_ui_customizations` | ✅ org/user/auth/RBAC/navigation menus |
| Stores & users (`20_stores_terminals`) | `stores`, `storage_locations`, `users`, `user_roles`, `user_store_access`, `cashiers`, `pos_terminals`, `cashier_sessions` | ✅ stores, locations, users, cashiers, terminals, sessions |
| Catalog (`30_catalog`) | `product_categories`, `brands`, `units_of_measure`, `uom_packaging_templates(_levels)`, `price_lists`, `tax_categories`, `products`, `product_variants`, `product_barcodes`, `product_prices`, `product_uom_conversions`, `product_serial_numbers`, `product_batches`, `stock_reservations` | ✅ categories/brands/UoM/packaging/price lists/tax/products/catalog search/variants/barcodes/prices; serials/batches partially via product APIs |
| Inventory (`40_inventory`) | `inventory_stock`, `stock_movements`, `stock_counts(_lines)` | ✅ stock + movements; stock counting schema/views only (no REST surface found) |
| Purchasing & partners (`50_purchasing_suppliers`) | `currencies`, `exchange_rates`, `payment_terms`, `cost_centers`, `chart_of_accounts`, `gl_account_mappings`, `business_partners`, `partner_addresses/contacts`, `customers`, `purchase_orders(_lines)`, `transfer_requests(_items)`, `goods_receipt_notes(_items)`, `bp_price_contracts`, `journal_entries(_lines)`, `sales_orders(v1)` | ✅ BP/customer/GRN/transfer APIs; PO query layer but no dedicated REST surface found; **COA/GL/currencies/cost-centers/payment-terms have no API or posting engine** (generated sqlc models only) |
| Sales/OMS (`60_sales_pos`, `85_zatca`) | `carts`, `cart_items`, `cart_activity_log`, `draft_cart_templates(_items)`, `sales_orders_v2`, `sales_order_lines_v2`, `order_status_history`, `order_fulfillments(_items)`, `quotes(_lines)`, `invoices(_lines/payments/status_history)`, `sales_returns(_lines)` | ✅ carts & orders v2 + fulfillments & returns; invoices/quotes queries exist in `cart_order_queries.sql` (no dedicated route group found) |
| POS + restaurant (`70_restaurant`) | `pos_transactions(_lines)`, `pos_payments`, `restaurant_tables`, `menu_categories`, `menu_items`, `menu_item_modifiers`, `menu_modifier_groups`, `promotions`, `recipes(_ingredients)`, `combo_bundles(_items)`, `menu_item_availability_schedules`, `restaurant_orders(_items)`, `waste_logs`, `kiosk_sessions` | ✅ POS, restaurant, promotions, waste, kiosk |
| Loyalty & analytics (`80_promotions_loyalty`, part of `10`) | `sales_analytics`, `purchase_analytics`, `inventory_analytics`, `discount_analytics`, `profit_loss_analytics`, `loyalty_redemption_rules`, `audit_logs` | ✅ loyalty; analytics views/functions (some exposed via query endpoints) |
| Compliance (`85_zatca`) | `zatca_device_configs`, `zatca_document_chain`, `sync_watermarks`, invoice/quotes tables above | ✅ ZATCA |
| POS-client extensions | `local_printer_configs`, `sync_queue` (outbox), `local_device_config` | n/a (client-local) |

Legend: ✅ = routing/handler/usecase present at time of writing · 🔶 = schema (+ sometimes generated sqlc models/queries) with partial or no REST surface · ⬜ = schema only.

### 3.3 Feature-status highlights that matter for the gap analysis

- **Strong / ahead of typical SMB ERP baselines:** omni-channel order lifecycle (cart → order v2 → fulfillment w/ pick/pack/ship → invoice/return), offline-first POS with transactional outbox and delta sync, recipe-based BOM costing, promotions engine (BOGO, happy-hour, stacking, schedules), loyalty rules, ZATCA Phase-2 e-invoicing, configurable navigation/RBAC/UI customization, multi-store + multi-warehouse + stock reservations + transfer lifecycle functions.
- **Latent but not wired:** GL primitives (`chart_of_accounts`, `cost_centers`, `gl_account_mappings`, `journal_entries/lines`), `exchange_rates`, `payment_terms`, `invoices/quotes`, `stock_counts`. No code path posts POS/GRN/order events into the ledger; no financial report is served.
- **Repeated ad-hoc patterns that should become engines:** document numbering (each `*_number` is a free `VARCHAR(50) UNIQUE` filled by callers), approval (`approved_by` FK columns), audit (`audit_logs` table but no generic trigger/API layer), status enums duplicated across documents, currency (`VARCHAR(3)` default literals scattered) — the hallmarks of a system that has not yet had a "financial backbone" pass.

### 3.4 Schema conventions observed

- `SERIAL`/`BIGSERIAL` ids (v1 tables) mixed with `UUID` ids (v2 tables, carts/orders/invoices); v2 documents snapshot customer/product names for historical accuracy.
- Decimal money everywhere `DECIMAL(15,2)`, quantities `DECIMAL(15,3)`, cost `DECIMAL(15,4)`.
- `JSONB metadata` on nearly every table — an informal user-defined-field mechanism (see §8 extensibility recommendation).
- `organization_id` on tenant-scoped tables; `store_id` where location matters; soft-delete absent (mostly `is_active`).
- Domain triggers/functions: totals calculators, `updated_at` maintenance, stock transfer/GRN/count/reconcile procedures, inventory allocation on order lines, daily analytics refresh.

---

## 4. Domain-by-domain gap matrix

Severity scale used below: **Critical** (blocks selling as ERP), **High** (material adoption blocker or top user pain), **Medium** (competitive differentiator gap), **Low** (nice-to-have), **None** (parity/ahead). Priority tags P0/P1/P2 map to the roadmap in §9.

### 4.1 Business partners & master data

| | |
|---|---|
| **Nembus today** | `business_partners` (supplier/vendor/special_customer/corporate_group) with addresses/contacts; separate `customers`; `bp_price_contracts` (schema-only). No single "partner" identity shared across AR & AP; no BP groups/segments with properties; no partner balance ledger beyond `outstanding_balance`/`credit_limit` columns. |
| **SAP B1** | Business Partner master is one of the two cornerstones (with Items). CardTypes (customer/vendor/lead), BP groups, properties, contact persons + addresses (bill-to/ship-to), credit limits & warnings/blocks, balance + transactions per BP, per-BP currency & price list & payment terms, attachments. |
| **S/4HANA standard** | **Business Partner (CVI)** — one master, many roles (customer, vendor, contact). Customer-vendor integration; global data vs company/role data; centralized master governance. |
| **Gap severity / direction** | **High (P1).** Unify `customers` + `business_partners` behind one partner master (or a formal link table) with roles (`customer`, `supplier`, `both`), BP groups, properties/attributes, and a real balance ledger. Keep `customers` as a compatible view/table for the POS stack to avoid breaking existing queries. |

### 4.2 Items / product master & pricing

| | |
|---|---|
| **Nembus today** | Deep catalog: categories tree, brands, UoM + conversions + packaging templates, price lists (+ per-UoM, quantity-break, dated prices), tax categories, variants, barcodes, serial/batch flags, sellable/purchasable/track-inventory flags. Product API incl. master-catalog search view (`v_master_product_catalog`). No item *groups* with default settings, no item cost/valuation method per item, no supplier-item linkage (preferred vendor, vendor item code), no substitute/alternative items. |
| **SAP B1** | Items + item groups (with defaults), UoM groups, multiple warehouses/prices (price lists), serial & batch mgmt, **item management method** (moving avg / standard / FIFO / serial / batch), procurement & sales data per item (preferred vendor, min/max, lead time), substitutes & alternatives, item attributes/properties. |
| **S/4HANA standard** | Material master views (basic/sales/purchasing/plant), item hierarchy, batch + serial with documents; flexible UoM; price & condition technique. |
| **Gap severity / direction** | **Medium-High (P1/P2).** Add per-item costing method + valuation class and supplier links (preferred vendor, supplier SKU, lead time, min order qty) — needed by §7 costing and by PO flows. Item *group* concept can be layered as a `product_groups` table mirroring category defaults. |

### 4.3 Pricing, discounts & conditions

| | |
|---|---|
| **Nembus today** | Price lists, quantity breaks, dated validity, per-variant/UoM prices; BP price contracts (schema-only); promotions engine incl. BOGO/happy-hour/stacking; coupon codes on carts/orders. |
| **SAP B1** | Price lists with base-price + factor lists, BP-specific prices, **discount groups**, special prices per BP/item, period/volume discounts; (on HANA) price lists copied to analytic views. |
| **S/4HANA standard** | Condition technique (condition types, scales, validity, records), contract pricing. |
| **Gap severity / direction** | **Medium (P2).** A promotion/condition "engine" that resolves price = f(list price, partner group, quantity, channel, time, stacking rules) would beat B1's discount groups. Nembus is close; formalize a pricing-context resolver rather than new tables only. |

### 4.4 Sales, order management & POS

| | |
|---|---|
| **Nembus today** | **Ahead of B1 defaults**: carts (guest/registered, channels), `sales_orders_v2` with status/payment/fulfillment machines + history, fulfillments w/ pick-pack-ship, quotes, invoices (draft/sent/paid/overdue), returns/refunds, POS transactions/payments/cashier sessions, sales channels incl. online/POS/kiosk/phone; stock reservations on order lines; multi-store. |
| **SAP B1** | Sales quotations → orders → deliveries → AR invoices → returns/credit memos; down payments; documents tied to inventory & GL postings; document lifecycle (draft → posted → cancel) with numbering series. B1 is weaker on true omnichannel/POS (separate add-ons), richer on the **financial consequences** of sales. |
| **S/4HANA standard** | Order-to-cash with condition pricing, delivery & billing blocks, credit management, revenue recognition, and automatic GL posting per step. |
| **Gap severity / direction** | **None→Low on process coverage; High on consequences.** Nembus covers the retail commerce flow well; the gap is that orders/invoices/returns do **not post to a ledger, do not allocate stock costing, do not drive receivables aging** — see §5. Add: sales-quotation→order conversion semantics, down-payment / prepayment handling, and "delivery vs invoice" separation for wholesale. |

### 4.5 Purchasing, payables & supplier documents

| | |
|---|---|
| **Nembus today** | `business_partners` suppliers, `purchase_orders(_lines)` (schema+queries; REST surface not found), `transfer_requests` full lifecycle functions, `goods_receipt_notes(_items)` API w/ posting (receipt, rejections, batch/expiry capture). No purchase quotations, no AP invoices, no supplier returns/credit notes, no landed costs, no down payments, no purchase analytics per supplier contract. |
| **SAP B1** | Purchase quotations → purchase orders → goods receipt PO → **AP invoice** (matched to receipt) → outgoing payments; AP credit memos & returns; **landed costs** (freight/customs allocation to receipts); purchase down payments; item costing updated on GRN (moving average recalculation). |
| **S/4HANA standard** | Procurement with purchase requisitions, source lists, outline agreements, invoice verification w/ 3-way match, valuation via material ledger, logistics invoice verification. |
| **Gap severity / direction** | **High (P0/P1).** Missing AP-invoice document + 3-way match (PO line ↔ GRN line ↔ AP invoice), purchase quotation, and landed-cost allocation engine (see §7). GRN already triggers stock + can trigger GL once posting engine exists. |

### 4.6 Inventory, warehousing & stock valuation

| | |
|---|---|
| **Nembus today** | Multi-store/warehouse + `storage_locations` (tree), on-hand/allocated/available/on-order/in-transit per store, movements ledger w/ cost capture, stock counts (schema), reservations, batches/serials w/ store & status, transfers with approval/ship/receive functions, reorder levels on stock. Strong retail inventory. |
| **SAP B1** | Warehouses (incl. bin locations for managed warehouses), goods receipt/issue documents, inventory transfers + transfer postings, **item costing methods with automatic valuation updates**, inventory counting & revaluation, stock by warehouse/bin reports, cycle counting, stock taking w/ GL difference postings. |
| **S/4HANA standard** | Stock types (unrestricted/quality/in-transit), **material ledger for actual costing** (multi-level), valuation areas, batch/serial lifecycle, embedded EWM bin strategies. |
| **Gap severity / direction** | **High (P0/P1).** Movement types exist but there is no **valuation ledger** (costing method per item, moving-average recompute on receipt, COGS on issue/sale) and no goods-issue / goods-receipt-for-production documents. `stock_movements` already carries `cost_per_unit`/`total_value` — formalize into item valuation entries (see §7). Bins: Nembus locations ≈ B1 bin locations — extend to multiple stock per (item, bin) incl. quality status if needed later. |

### 4.7 Finance — general ledger, posting, periods & reporting  ⭐ highest-priority gap

| | |
|---|---|
| **Nembus today** | `chart_of_accounts` (asset/liability/equity/revenue/expense, self-referencing tree), `cost_centers` (dimension field), `gl_account_mappings`, `journal_entries/_lines` (debit/credit, reference_type/id), P&L/discount analytics tables, aging & AP views. **No posting engine**: nothing writes journal lines; no fiscal year / posting period management; no document numbering series; no COA groups & standard reports (balance sheet, trial balance, cash flow); no budget/scenarios; no multi-GAAP; no intercompany. |
| **SAP B1** | Full Financials module: COA with account groups & default reports, journal entries (manual + automatic from every document), **posting periods + fiscal years (open/close, period indicators)**, document numbering series per doc type, currencies: base/system/secondary + **revaluation**, cost accounting (cost/profit centers + distribution rules), **budget & scenarios**, financial reports incl. balance sheet/P&L/cash-flow/trial balance, internal reconciliation. |
| **S/4HANA standard** | **Universal Journal (ACDOCA)**: one append-only financial line store combining FI-GL, AP/AR, asset, CO, margin — every business event posts once; dimensions at line level; parallel ledgers/GAAP. |
| **Gap severity / direction** | **Critical (P0).** This is the defining ERP gap. Adopt a S/4HANA-inspired design on PostgreSQL: a `general_ledger_entries` line store with dimensions (account, cost center/profit center/segment, store, partner, document ref, currency) into which POS receipts, GRNs, invoices, returns, payments post **through one posting service/trigger**, plus `fiscal_years`/`posting_periods`, numbering series, and reporting views (trial balance, balance sheet, P&L, cash flow). Full design in §5. |

### 4.8 Receivables / payables subledgers, payments & credit management

| | |
|---|---|
| **Nembus today** | Invoices w/ payment status + `invoice_payments` (reconciled flag, bank_account_id placeholder), `payment_terms` (schema-only), customers' `credit_limit`/`outstanding_balance`, `vw_customer_aging_report`, `vw_accounts_payable`. **No open-item allocation** (payments not allocated to specific invoices/credit memos), no AR credit memos applying to invoices, no dunning, no credit-block workflow. |
| **SAP B1** | Incoming/outgoing payments **allocated to open documents** (invoice/credit memo/down payment), partial payments, discounts/rounding during allocation, **dunning (collection letters) with levels**, credit limit checks & blocks/warnings, aging reports, checks & postdated checks, deposits. |
| **S/4HANA standard** | FI-AR/AP as roles of Business Partner + open-item management, clearing, **dunning**, collections & dispute management (AI-enhanced), credit management (credit segments, checks on sales orders/deliveries). |
| **Gap severity / direction** | **Critical/High (P0).** Introduce open-item subledgers: payments allocate across invoices (invoice_allocations), credit/debit memos as invoice types with application, aging by open items (replace `outstanding_balance` recompute), dunning levels. Credit hold on order creation when partner exceeds limit. |

### 4.9 Banking, cash & reconciliation

| | |
|---|---|
| **Nembus today** | No bank accounts master; `invoice_payments.bank_account_id` placeholder; POS payment methods incl. cash/card/gateways; cashier sessions w/ opening/closing balances & variance; change computation. |
| **SAP B1** | Bank account master (house banks), deposits, checks received/issued & postdated checks, **internal bank reconciliation** (side-by-side bank/GL), external reconciliation from bank statements, outgoing/incoming payment batches. |
| **S/4HANA standard** | Bank Account Management, cash/bank journal, bank statement import + **automatic/manual clearing**, cash application incl. ML-based suggestions, liquidity/cash-flow planning. |
| **Gap severity / direction** | **High (P0/P1).** Add `bank_accounts`, check management, payment batches, and a **reconciliation workspace**: import statement lines → match open payments/invoices (auto-rules then manual) → clear. §5 includes the schema. |

### 4.10 Fixed assets & depreciation

| | |
|---|---|
| **Nembus today** | None. |
| **SAP B1** | Core B1 lacks a full asset module in all markets (often delivered by localization/add-ons); where present: asset master, capitalization, depreciation (several methods), manual & automatic posting. |
| **S/4HANA standard** | Asset Accounting: asset master w/ depreciation areas (book/tax/IFRS), acquisition/production, capitalization, depreciation run, retirement/scrapping; integrated in Universal Journal. |
| **Gap severity / direction** | **Medium (P2)** for core retail; valuable for growth to wholesale/distribution. A pragmatic scope: `fixed_assets` (master, depreciation method/rates, depreciation areas), acquisition via AP invoice or manual entry, monthly depreciation run → GL entries. Not P0; do not build B1-style localization breadth. |

### 4.11 Production, BOM & MRP

| | |
|---|---|
| **Nembus today** | Restaurant BOMs: `recipes`/`recipe_ingredients` with yield, optional/byproduct ingredients, cost calc (`fn_calculate_recipe_cost`); `combo_bundles` for sales kitting. **No manufacturing documents** (production order, issue to production, receipt from production), no operations/routings, no multi-level BOM, no MRP. |
| **SAP B1** | Production module: production orders, BOMs, issue for production, receipt from production, item costing at production receipt; **MRP wizard** (net requirements, generates purchase/production suggestions), planning calendar. |
| **S/4HANA standard** | MRP Live, MRP areas, BOM/routing, production orders w/ scheduling & capacity, PP-DS; predictive/replenishment AI. |
| **Gap severity / direction** | **Medium-High (P1, staged)** — the user prioritizes it. Pragmatic path: (1) generalize `recipes` → multi-level `bill_of_materials`; (2) `production_orders` + `production_order_issues/receipts` posting to stock & costing (§7); (3) MRP-lite: net-requirements engine over inventory/on-order/open-sales → purchase suggestions + production suggestions; (4) later: routing/operations & scrap. See §7. |

### 4.12 CRM — opportunities, activities, service, campaigns

| | |
|---|---|
| **Nembus today** | Customers/BP with contacts & addresses; sales order assignment (`assigned_to_user_id`); loyalty & coupons. No pipeline, no activities/journal, no service calls/contracts, no campaigns. |
| **SAP B1** | **Sales Opportunities** (stages, weighted pipeline, forecasts), **CRM activities** (calls/tasks/meetings w/ calendar), **Service module** (service calls, contracts, queue), marketing campaigns w/ segments + mail merge; opportunities convert to quotations/orders. |
| **S/4HANA standard** | Sales cloud / CPQ; service; AI lead scoring. (For SMB ERP, B1 scope is the benchmark.) |
| **Gap severity / direction** | **Medium (P1/P2).** Small-footprint CRM: `opportunities` (stage/amount/probability/owner) convertible to quote/order, `activities` timeline on partners, `service_calls` (+ contracts later). High perceived value for B2B/wholesale buyers of an ERP; not needed for pure retail POS. |

### 4.13 Human resources & employees

| | |
|---|---|
| **Nembus today** | `users` w/ `employee_code`, cashiers; no HR. |
| **SAP B1** | Employee master + payroll in selected localizations (weak, add-on driven). |
| **S/4HANA standard** | Full HCM is out of scope for Nembus; standard is only master-data integration. |
| **Gap severity / direction** | **Low (P2+).** Add `employees` linked to users (department, manager, hire date, role) sufficient for commission reporting & future payroll add-ons; do not attempt payroll engine. |

### 4.14 Restaurant & hospitality operations

| | |
|---|---|
| **Nembus today** | Very strong: tables/sections, menu mgmt with modifiers/min-max groups, availability schedules, KDS views/functions, recipes costing, waste logs, combo/bundles, kiosk sessions, order sources (counter/dine-in/etc.), promotions. |
| **SAP B1** | No hospitality depth — restaurant is served by B1 add-ons/partners (e.g., Fore, LS Retail / S4H Retail). |
| **S/4HANA** | Retail/food via S/4HANA Retail add-ons. |
| **Gap severity / direction** | **None (Nembus differentiator).** Keep investing here per market trends (delivery aggregator integrations, AI upsell, table-side payments, kitchen display optimization). |

### 4.15 Tax, e-invoicing & compliance

| | |
|---|---|
| **Nembus today** | `tax_categories` (rate, inclusive flag) on products/documents; **ZATCA Phase-2 e-invoicing** (device config, document chain, XML/crypto, sync watermark) — genuinely ahead for KSA. No multi-tax-code/tax-group engine, no withholding/non-deductible VAT, no tax reporting suite, no per-region localization framework beyond ZATCA. |
| **SAP B1** | Tax groups & tax codes per transaction; VAT/non-deductible/withholding handling; tax reports by localization (KSA ZATCA via localization); e-document compliance. |
| **S/4HANA standard** | Document & Reporting Compliance (PEPPOL, ZATCA, e-Fattura, CFDI, ViDA real-time reporting), tax determination & calculation mgmt, e-invoice archiving/audit. |
| **Gap severity / direction** | **Medium (P1).** Build a **tax engine** (tax groups → multiple tax lines per document incl. inclusive/exclusive, withholding, exempt) and a localization/e-invoice framework so ZATCA today can be joined by UAE/VAT-GCC, VAT in digital age (ViDA) and others tomorrow. See §8. |

### 4.16 Analytics, reporting & BI (the "HANA" benchmark)

| | |
|---|---|
| **Nembus today** | Pre-aggregated daily fact tables (`sales_analytics`, `purchase_analytics`, `inventory_analytics`, `discount_analytics`, `profit_loss_analytics`) refreshed by `fn_refresh_daily_analytics`; operational views (low stock, pending POs, aging, AP, margin, master catalog). No self-service BI, no KPI cockpit/dashboards app, no OLAP/semantic model, no real-time reporting, no forecasting/ML. |
| **SAP B1 on HANA** | Real-time in-memory analytics: dashboards & KPI cockpit, drag-and-relate queries, Excel integration, Crystal Reports; marketing positions HANA edition as "real-time insights". |
| **S/4HANA standard** | Embedded analytics via CDS/VDM — analytic queries run **on transactional data with no duplication**; Fiori KPI tiles; SAC for planning/BI; ML/AI on top. |
| **Gap severity / direction** | **High (P1/P2) but architectural.** Short term: expose existing data through a proper **analytics schema + read API** (semantic views, KPI endpoints). Medium: for "on-HANA-like" positioning, add a columnar/real-time layer (e.g., Timescale/PG partitioning or ClickHouse/DuckDB export via the existing outbox) feeding dashboards; adopt **event/CDC streaming from the outbox** instead of nightly batch. Long term: forecasting/AI (see §8). |

### 4.17 Platform mechanics — numbering, approvals, alerts, extensibility

| | |
|---|---|
| **Nembus today** | Free-text document numbers; ad-hoc `approved_by` FKs; `audit_logs` table; `ui_settings`/JSONB `metadata` as informal extensibility; `role_permissions` scoping; alerts absent. |
| **SAP B1** | **Document numbering series** (per doc type + branch), draft vs posted, period control; **approval procedures** (multi-stage templates w/ conditions); **alerts manager**; **user-defined fields & tables**; authorization hierarchy incl. row-level via branches/departments; audit trail per document. |
| **S/4HANA standard** | Number ranges w/ intervals; workflow (SAP Build) incl. approvals; **key-user extensibility (custom fields/custom logic)** on a clean core; eventing + API hub. |
| **Gap severity / direction** | **High (P0/P1).** These become *prerequisites* once finance (P0) exists (posting periods, numbering series, approvals for posting). Design: `document_series`, `approval_templates/steps/requests`, `alerts`/notification bus, formal **UDF registry** (`udf_definitions`, values in JSONB) to replace ad-hoc metadata; eventing via existing outbox → webhooks. See §8. |

### 4.18 Identity, security & multi-tenancy standards

| | |
|---|---|
| **Nembus today** | JWT auth, RBAC w/ role+store scoping, tenant-per-database isolation (strong), m2m tokens, audit table. No SSO (SAML/OIDC), no SCIM, no MFA, no row-level data permissions beyond store, no per-field/approval authority, no session/device management UI. |
| **SAP B1** | B1 superuser + per-user authorizations w/ granular permissions incl. row level; B1 on-prem auth; newer: integration with SAP IAS/cloud identity via Service Layer. |
| **S/4HANA / market standard** | OIDC/SAML SSO, MFA, SCIM provisioning, ABAC/fine-grained, audit & privacy (GDPR). |
| **Gap severity / direction** | **Medium (P1).** SSO/OIDC + MFA and SCIM provisioning are table stakes for enterprise sales; row-level permission layer (organization→store→record scope) for multi-branch chains. |

### 4.19 Integration, APIs & events

| | |
|---|---|
| **Nembus today** | REST/JSON (Gin) + OpenAPI/swagger docs; gRPC backup; POS offline sync via outbox + watermarks; m2m tokens. No OData, no public webhooks/event subscriptions, no API versioning/rate limiting/usage analytics, no iPaaS connectors, no CDC publishing for analytics. |
| **SAP B1** | Service Layer (OData v4 REST), DI API (COM/.NET), B1iF, add-on SDK; webhook/notification support in recent Service Layer versions. |
| **S/4HANA / market standard** | API-first (OData/REST, API Business Hub), **webhooks & event mesh**, outbox/CDC patterns, iPaaS connectors, OpenAPI + SDKs, sandbox/tenant provisioning. |
| **Gap severity / direction** | **Medium (P1).** Promote the POS outbox into a general **event outbox** with webhook subscriptions + retries; version the REST API; publish OpenAPI for the whole surface; expose OData option later for B1 Service Layer-compatible tooling (Excel, Power BI). |

### 4.20 AI & emerging-market trends (2024–2026)

| | |
|---|---|
| **Nembus today** | None beyond analytics tables and ZATCA crypto. |
| **Market baseline** | AI copilots (SAP Joule), demand forecasting, dynamic pricing, anomaly detection (sales/stock), intelligent cash application & collections, natural-language query over ERP data, ML-assisted catalog enrichment; plus subscription/recurring revenue engines, BNPL/wallets, sustainability/ESG reporting. |
| **Gap severity / direction** | **Medium (P2) but high narrative value.** Structure data now (events, finance line store, semantic views) so AI can be layered later; pick two 2025-visible bets: demand forecasting for replenishment, and a copilot/NL→SQL over the semantic layer. Details in §8. |

**Roll-up:** critical gaps = GL/posting/financial backbone (§4.7, §4.8); high gaps = AP + 3-way match, stock valuation/costing, banking & reconciliation, numbering/approvals/platform, analytics positioning. Everything else is Medium/Low or a Nembus strength.

---

## 5. Priority deep-dive A — Finance & accounting

> Goal: turn Nembus from "POS + orders + schema-only GL" into a system where **every business event has a financial consequence recorded in one audit-proof ledger**, following the S/4HANA Universal-Journal idea (single line store, dimensions at line level) at an SMB scale, with SAP-B1-style posting periods, document series and AR/AP subledgers.

### 5.1 Design principles (borrowed from SAP, sized for Nembus)

1. **Everything posts, once.** POS sale, cashier session close, invoice, payment, credit memo, GRN, AP invoice, stock adjustment, production receipt → the posting service writes balanced double-entry lines to a single append-only ledger line store. No separate subledgers to reconcile — the AR/AP invoice tables and the ledger are kept consistent by the same service (Universal-Journal idea).
2. **Ledger rows are immutable & reference-complete.** Each line stores: organization, ledger/period, posting date, document type + document number + source doc reference, account, partner (customer/supplier) when applicable, dimensions (cost center, profit center, segment, store, project), currency + rate + amount in doc currency, amount in base currency, debit/credit, memo. Unique constraint on (source_type, source_id, source_line) → **idempotent posting** (offline POS can post via sync retries without duplicates).
3. **Periods before postings.** A document may only post to an **open posting period**; closing a period blocks further postings (B1 period indicators), enabling period-end close and audit.
4. **Documents, not free numbers.** Every financial document number comes from a **document series** (prefix + per-period/per-year counters + draft vs posted), replacing free-text `VARCHAR(50) UNIQUE` numbers.
5. **Currency discipline.** Base (functional) currency per organization; every money document stores its own currency + rate; amounts are also recorded in base currency; FX revaluation of open items at period end (B1 revaluation).
6. **Posting happens cloud-side with authority.** POST/put "post" endpoints run in a transaction: validate → number → write ledger lines → update source doc status → enqueue events. For offline POS, the outbox replays financial events through the same idempotent posting service.

### 5.2 Core schema foundation (proposed; to be implemented as Atlas migrations)

> Names follow Nembus conventions (`snake_case`, `organization_id` on tenant tables, `created_at/updated_at`, JSONB `metadata`). Sketches are column lists, not final DDL — exact types/indexes decided at implementation.

**Fiscal calendar & posting control**

```
fiscal_years         (id, organization_id, year, name, start_date, end_date,
                      status open/closed, metadata)
posting_periods      (id, organization_id, fiscal_year_id, period_no, start_date, end_date,
                      period_name, status draft/open/closed/locked, closed_at, closed_by,
                      allow_retro_posting, metadata)
```

**Document series & numbering**

```
document_series      (id, organization_id, doc_type e.g. 'sales_invoice'|'ap_invoice'|'payment'|'grn'…,
                      series_code, name, prefix, suffix, start_no, next_no, increment,
                      is_default, per_fiscal_year bool, is_active, metadata)
```

Apply a series to every document table (add `series_id`, keep `*_number` as the rendered number) — or a shared `document_numbers` counter table; series table above is B1-style and clearer.

**Chart of accounts (evolve existing `chart_of_accounts`)**

Add: `account_group`/`account_type_detail` (e.g. B1-like groups under asset/liability/equity/revenue/expense), `is_control_account` (AR/AP/tax control accounts), `currency_code` (hard-currency accounts), `is_bank_account`, `is_active` already present, `external_code` (mapping to statutory/local codes). Optionally `account_aliases`. B1-style financial report templates reference groups, not accounts.

**Dimensions (evolve `cost_centers`)**

```
profit_centers       (id, organization_id, code, name, parent_id, is_active)
segments             (id, organization_id, code, name, …)   -- optional, S/4HANA-style
```

`cost_centers.dimension` exists — formalize a small `dimensions` dictionary table and allow N lines per posting to carry (cost_center_id, profit_center_id, store_id, project_id).

**The ledger (evolve `journal_entries`/`journal_lines` into a line store)**

Keep `journal_entries` as the posting header (document, period, posting date, status, memo, source reference, series) and make `journal_lines` carry **all dimensions + dual amounts**:

```
journal_entries      (…existing + fiscal_year_id, posting_period_id, series_id,
                      status draft/posted/void, base_currency_code, total_debit/total_credit
                      in base currency, reversal_of_id, created_by, posted_by, posted_at)
journal_lines        (…existing + line_no, account_id, partner_id (BP), customer/supplier,
                      cost_center_id, profit_center_id, store_id,
                      currency_code, exchange_rate, amount_doc_currency DECIMAL(15,2),
                      amount_base DECIMAL(15,2),  -- functional currency
                      debit_base/credit_base or signed amount,
                      reference_type/reference_id/reference_line (source doc + line),
                      statement/notes, metadata)
```

Recommended: store **signed `amount_base`** plus `debit/credit` for display OR keep debit/credit and add base-currency debit/credit — pick one convention at implementation; the essential S/4HANA property is *dimension-bearing, append-only, single store*.

Indexes: `(organization_id, posting_date)`, `(reference_type, reference_id)`, `(account_id, posting_period_id)`, `(partner_id)` partial for open items.

**GL mappings (evolve `gl_account_mappings`)**

Extend to a rule engine: mapping_type (e.g. `sales_revenue_store`, `cogs`, `inventory_store`, `cash_payment_method`, `tax_output`, `ar_control`, `ap_control`, `discount_given`, `rounding`), store/payment-method/tax-dimension selectors → GL account. One place the finance team maintains so postings need no code change per account.

**AR/AP open items & allocations**

```
invoice_allocations  (id, organization_id, invoice_id (or ap_invoice_id), payment_id (invoice_payments or ap_payments),
                      allocated_amount, allocated_amount_base, discount_taken, write_off,
                      allocation_date, created_by, metadata, UNIQUE(invoice_id, payment_id))
```

- Keep `invoices` as the AR document incl. `credit_applied`, `balance_due`; add `ap_invoices` (mirror) + `ap_payments` or reuse one `payments` ledger with direction/partner role — recommend **one `payments` table with partner role + direction** to avoid AR/AP table duplication (customer & supplier share `business_partners` master after §4.1).
- Dunning: `dunning_levels` + `dunning_runs`/`dunning_history` over open items.
- Credit mgmt: `credit_limit`/`balance` on partner; service-level check on order creation (warn/block) — B1 semantics.

**Banking & reconciliation**

```
bank_accounts       (id, organization_id, code, name, bank_name, branch, account_number,
                     iban, currency_code, gl_account_id, is_active, …)
bank_statements     (id, organization_id, bank_account_id, statement_number, statement_date,
                     opening_balance, closing_balance, import_file/format, status, …)
bank_statement_lines(id, bank_statement_id, line_no, value_date, posting_date, description,
                     reference, debit/credit amount, currency, matched_payment_id, match_status,
                     reconciliation_date, …)
check_registry      (id, organization_id, bank_account_id, check_number, partner_id, amount,
                     issue_date, due_date, status issued/deposited/cleared/bounced/void, …)
```

Reconciliation workspace: statement lines → suggested match (payment reference/amount/date, AI-assisted later) → user confirms → clears open items & posts bank/GL entries.

**Fixed assets (P2, scoped)**

```
fixed_assets         (id, organization_id, asset_code, name, category, partner_id (vendor),
                      acquisition_date, acquisition_cost, residual_value, useful_life_months,
                      depreciation_method, depreciation_rate,
                      gl_accounts (asset/depreciation/accumulated/cost), status, …)
depreciation_runs    (id, organization_id, period_id, run_date, status, …)
depreciation_lines   (id, run_id, asset_id, period_id, amount_base, posted_journal_id, …)
```

**Budgeting (P2)**

```
budgets              (id, organization_id, fiscal_year_id, account_id, cost_center_id/profit_center_id,
                      period_no or month, amount, scenario, …)   → actual-vs-budget report view
```

### 5.3 Posting map — every Nembus event that should hit the ledger (P0 scope)

| Source event (exists) | Proposed GL consequence | Needs new doc? |
|---|---|---|
| POS sale completed (`pos_transactions` + lines + `pos_payments`) | Dr Cash/Bank per payment method, Cr Sales revenue (per store mapping), Cr/Dr VAT output (inclusive/exclusive), Dr COGS (from item costing, §7) per line | Posting service + outbox consumer |
| Cashier session close (variance) | Dr/Cr Cash over/short | — |
| Sales invoice posted (`invoices`) | Dr AR control (customer), Cr Revenue/Tax; opens AR item | posting on invoice |
| AR payment / allocation | Dr Bank/Cash, Cr AR control; allocate open items; FX & discount handling | payments + allocations |
| Credit memo / AR return | Reverse revenue/tax/AR; stock-in + COGS reversal if restocked | existing `sales_returns`/invoice_type credit_note |
| GRN posted (`goods_receipt_notes`) | Dr Inventory (per store + item), Cr AP accrual / GR/IR clearing | existing fn_process_goods_receipt |
| AP invoice + 3-way match | Dr Inventory/Clearing, Cr AP control; opens AP item | **new `ap_invoices`** |
| AP payment & allocation | Dr AP control, Cr Bank | payments |
| Stock count variance (`fn_reconcile_stock_count`) | Dr/Cr Inventory, Cr/Dr Inventory variance/expense | — |
| Stock transfer | no GL (movement) or in-transit accounts later | — |
| Waste (`waste_logs`) | Dr COGS/Waste expense, Cr Inventory | — |
| Purchase order | no GL (commitment only, optional) | — |

### 5.4 Reporting views to deliver (P0/P1)

`vw_trial_balance`, `vw_general_ledger`, `vw_balance_sheet`, `vw_profit_and_loss` (replace manual `profit_loss_analytics` as source of truth), `vw_cash_flow` (indirect), `vw_ar_aging`/`vw_ap_aging` (from open items, not recomputed balances), `vw_budget_vs_actual` (P2), `vw_fixed_asset_register` (P2). B1-style: financial reports driven by account-group report templates so a finance user can add/remove lines without SQL.

### 5.5 Sequencing & risks (finance)

- **P0 (do first):** fiscal years/periods, document series, COA groups + mappings rule, ledger line store, posting for POS + invoices + payments + GRN + returns, trial balance + P&L + balance sheet views, AR/AP open items & allocations, basic dunning, bank accounts + internal reconciliation.
- **Watch-items:** idempotency of offline-POS posting (source-unique constraint); closing periods must not break in-flight POS sync (post to current open period, alert on period closed); money rounding per line vs per document (choose per-document rounding w/ a rounding account); performance of dimension-bearing line store at high POS volumes (partition by fiscal year; move to columnar export for BI later — §8); dual-write consistency between ledger and existing `*_analytics` tables (posting service should also feed/refresh analytics or mark them deprecated).
- **Testing:** double-entry invariants (`sum(debit)==sum(credit)` per entry), idempotent re-post, period close blocking, FX revaluation arithmetic, and an end-to-end "POS receipt → GL → P&L" golden test.

---

## 6. Priority deep-dive B — Commerce & omnichannel

> Nembus is already strong here; the goal is closing specific monetization/retention gaps and standardizing mechanisms so commerce features compose (promotions ↔ gift cards ↔ credit ↔ loyalty ↔ BNPL all apply in one pricing/payment context).

### 6.1 Store value & customer balance (Nembus gap today: none of these exist as ledgers)

- **Gift cards / store credit / loyalty-as-currency** currently only via `customers.loyalty_points` + `loyalty_redemption_rules` (earn/redeem rates, caps, expiry).
- **Proposed schema** (one generic value-store model beats three parallel tables):

```
customer_value_accounts (id, organization_id, customer_id, account_type gift_card|store_credit|loyalty|deposit,
                         balance, currency_code, status, expires_at, metadata)
customer_value_transactions (id, account_id, transaction_type issued|top_up|redeem|refund|expire|void|adjust,
                         amount, reference_type/reference_id (sale, return, order), store_id,
                         created_by, reason, metadata)
gift_card_codes        (id, organization_id, code, pin, account_id, initial_value, issued_at,
                        activated_at, status, …)   -- for physical/digital gift cards
```

All commerce value (gift card, store credit from returns, loyalty redemption, prepaid deposit) flows through `customer_value_accounts`; POS/order checkout can apply multiple value types + split tender in one payment context. This is what lets Nembus beat B1 add-ons.

### 6.2 Returns, refunds & RMA orchestration

- `sales_returns` exists w/ refund method/reference + `sales_return_lines`; extend into a **return lifecycle** that composes: restock-to-inventory (with batch/serial), value-store credit (if no original payment), original-payment reversal (with gateway refund token), exchange into a new order, RMA numbers + return reasons analytics, and multi-store return policy (buy-online-return-in-store: source store vs returning store → cross-store stock posting via `stock_movements`).
- Add `return_reasons` taxonomy + analytics to reduce return rate (market standard).

### 6.3 Subscriptions & recurring revenue (schema hints exist: `order_type subscription`, `invoices.is_recurring`)

- **P0/P1 gap:** no recurring-run engine (create invoice from subscription on schedule), no subscription plans/entitlements, no proration/cancel/upgrade semantics, no revenue recognition (IFRS 15/ASC 606 for deferred revenue).
- **Proposed:**

```
subscription_plans    (id, organization_id, name, billing_frequency daily/weekly/monthly/yearly,
                       billing_anchor, trial_days, auto_renew, …)
customer_subscriptions(id, organization_id, customer_id, plan_id, price_list_id, status active/paused/canceled/expired,
                       current_period_start/end, cancel_at, …)
subscription_line_items(id, subscription_id, product/service, quantity, unit_price, …)
recurring_invoice_runs(id, organization_id, run_date, status, …)  -- creates invoices from active subscriptions
deferred_revenue_schedules(id, organization_id, invoice_id, amount, period_from/to, recognized_amount, …)  -- P2
```

Revenue recognition (P2) can be a schedule table + period-end recognition journal (§5) — a compact ASC 606/IFRS 15 answer without S/4HANA RAR.

### 6.4 Loyalty 2.0

- Existing: earn/redemption rules table + points column. Gaps: **tiers** (status levels w/ multipliers & perks), **campaign-driven points** (bonus events), points expiry jobs, partner/affiliate programs, and loyalty *liability* accounting (accrued points = liability on the balance sheet; points redemption reverses it) — the finance tie-in is usually missed and is a great differentiator. Propose `loyalty_tiers`, `loyalty_tier_history`, `loyalty_campaigns`, extend `loyalty_redemption_rules` with tier eligibility, and add GL mappings for `loyalty_liability`.

### 6.5 Payments & checkout standards

- `pos_payments`/`invoice_payments` cover cash/card/gateway with reference numbers. Market standards to add: **split tender** (present as multiple rows — verify UI/flow), tips (restaurant), BNPL/instalments, wallets & tokenization, multi-acquirer routing, per-terminal settlement reconciliation vs `cashier_sessions`, digital receipts (email/SMS/whatsapp) w/ e-invoice links (ZATCA), and cash management (safe drops, petty cash, change floats) — the last is a real operations gap for chains.
- Proposed: extend `cashier_sessions` lifecycle w/ cash-drawer operations (`cash_drawer_operations`: float/open/drop/loan/pickup/close) — SAP B1 does not cover POS cash mgmt well, so this is a differentiator; standardize payment-method master (`payment_methods`) instead of free-text `payment_method` columns.

### 6.6 Omnichannel order management (OMS)

- Already strong (carts → orders v2 → fulfillments w/ pick/pack/ship; channels; `carts.channel`; order source/referral). Gaps vs market: **ship-from-store / store-fulfillment allocation** (fulfillment store exists; tie to `inventory_stock.quantity_allocated` + `stock_reservations` — partially there via triggers), **dropship** (vendor ships directly: PO from sales order line, vendor flag), **marketplace order ingestion** (normalize external order/shipment status; channel adapters as thin services over the order API), **buy-online-return-in-store** (see 6.2), order promises (availability → ETA), and a **cancellation/exception queue** (S4H order-management analog: order orchestration events).
- Standardize with an **order event log** (`order_events`: order_id, event, payload, occurred_at) built on the existing outbox pattern so every channel transition emits an event for webhooks/analytics (§8).

### 6.7 B2B/wholesale sales (needed to grow beyond retail)

- SAP B1's B2B strength: price lists per BP, special prices, **sales quotations→orders**, delivery vs invoice separation, payment terms, credit checks, partner-specific catalogs. Nembus has price lists + `bp_price_contracts` (schema-only) — **activate**: per-partner price & discount resolution, credit-limit enforcement at order time (with approval override), sales-quotation lifecycle, delivery (pick/pack ship exists) then periodic invoicing (consolidate multiple deliveries into one invoice) — mirrors B1 wholesale flow and reuses §5 posting.

**Commerce sequencing:** P0 = value-store (gift card/store credit) + return lifecycle + cash-drawer ops; P1 = subscriptions/recurring engine, OMS events + ship-from-store, B2B quotation/credit; P2 = revenue recognition, loyalty tiers/liability, BNPL/wallets integration framework.

---

## 7. Priority deep-dive C — Operations (costing, landed costs, production/MRP)

> Nembus tracks *quantities* well (stock, movements, transfers, counts, reservations) but not *values* with a costing method. Every operations gap below exists because there is no **item valuation engine**. SAP B1's differentiator for distributors/light-manufacturers is exactly this: documents change both quantity and value, and COGS lands in the P&L.

### 7.1 Item valuation & costing engine (foundation for everything else in this section)

**Per-item costing method** (B1's "item management method"; S/4HANA material-ledger idea scaled down):

- Add to `products`: `costing_method` (`moving_average` | `fifo` | `standard` | `specific` | `none`), `standard_cost`, plus store-level current average cost on `inventory_stock` (`avg_unit_cost`, `last_unit_cost`).
- **Valuation ledger**: every quantity movement that changes value writes a value entry; the ledger drives stock valuation reports and COGS. New table:

```
stock_valuation_entries (id BIGSERIAL, organization_id, store_id, product_id, product_variant_id,
                         batch_id, movement_type (grn|goods_issue|transfer_out|transfer_in|sale|return|adjust|count|prod_issue|prod_receipt|landed_cost|revaluation),
                         reference_type, reference_id, reference_line_id,
                         quantity DECIMAL(15,3), unit_cost DECIMAL(15,4), total_cost DECIMAL(15,2),
                         prev_avg_cost, new_avg_cost, posting_date, created_at, metadata,
                         UNIQUE(reference_type, reference_id, reference_line_id))   -- idempotent
```

- **Moving-average recompute** on each receipt: `new_avg = (old_qty*old_avg + received_qty*unit_cost(+landed)) / new_qty` (B1 semantics; store per store). FIFO: layers per batch/date; `specific`: batch/serial carries cost.
- **COGS on sale**: POS/order line already snapshots `cost_price`; costing engine replaces that snapshot with the method-derived cost at posting time, and (with §5) posts Dr COGS / Cr Inventory per line.
- Backward impact: existing `stock_movements.cost_per_unit/total_value` and `pos_transaction_lines.cost_price` columns are compatible; treat them as the pre-ledger approximation until the valuation ledger replaces them as the source of truth.

### 7.2 Landed costs (B1 purchasing feature; must-have for importers/distributors)

- **Proposed:** `landed_cost_types` (freight, customs duty, insurance, handling, …), `landed_cost_documents` (vendor invoice for freight, store/warehouse), and `landed_cost_allocations` tying a landed-cost document to a GRN with an **allocation basis** (by value | quantity | volume | weight | fixed) across GRN lines. Posting updates each line's unit cost (adds to the valuation ledger entry of the GRN) and re-computes moving average. GL side (with §5): Dr Inventory / Cr AP-accrual per landed-cost doc.

### 7.3 Goods issue / internal consumption & delivery notes

- Retail/wholesale already ships via `order_fulfillments`; missing are **non-sales stock-out documents**: internal use, marketing/sample, damage write-off, transfer to production, donation. **Proposed:**

```
goods_issues      (id, organization_id, store_id, issue_number, issue_type internal_use|sample|damage|donation|production,
                   issue_date, status, reason, approved_by, posted_journal_id, …)
goods_issue_items (id, goods_issue_id, product_id, variant_id, batch_id, quantity, unit_cost, total_cost, …)
```

Waste already exists (`waste_logs`) for restaurant — unify semantics so waste/damage/goods-issue all produce a valuation + GL entry (Dr expense / Cr inventory).

### 7.4 Production & BOM (staged; restaurant BOMs are the head start)

Stage 1 — **generalize BOMs** (recipes exist and are restaurant-grade):
- Reuse `recipes`/`recipe_ingredients` patterns but with multi-level support (`bill_of_materials` with `bom_lines`, line can reference a sub-BOM), scrap %, co/by-products (`recipe_ingredients.is_byproduct` already exists), and a **routing/operations** table (labor operations with cost rates — P2).

Stage 2 — **production documents**:

```
production_orders      (id, organization_id, store_id, po_number, product_id (finished), bom_id,
                        planned_qty, started_qty, completed_qty, scrap_qty,
                        status draft/released/in_progress/completed/cancelled, planned dates, …)
production_order_issues(id, production_order_id, product_id, variant_id, batch_id, quantity, unit_cost, …)
production_order_receipts(id, production_order_id, product_id, variant_id, quantity, unit_cost,
                        byproduct handling, posted_journal_id, …)
```

Flow (B1-like): release order → **issue to production** (Dr WIP / Cr Inventory per component — optionally backflushed automatically on completion) → **receipt from production** (Dr Inventory finished / Cr WIP) → costing variance (standard vs actual) posts to a variance account. `fn_calculate_recipe_cost` already computes expected component cost — extend into the production costing report.

Stage 3 — **MRP-lite** (B1 MRP wizard / S/4 MRP Live concept, deterministic, not APS):
- Net-requirements engine over: open sales orders + forecasts − current available stock − on-order (`inventory_stock.quantity_on_order`) → **purchase suggestions** and **production suggestions** honoring lead times & reorder/min quantities (columns already on `inventory_stock`). Output table `planning_suggestions` (type purchase/production/transfer, item, qty, date, source, status) that users release into POs/production orders/transfers.
- Inputs: `forecast_demand` table (from §8 analytics/AI or manual) — keep it a simple table so AI forecasting later just writes into it.

### 7.5 Margin & profitability reporting (CO-PA-lite)

With §5 ledger dimensions + §7 valuation: contribution-margin reports per product/location/channel/partner become views over journal/valuation lines (account-based margin analysis — the S/4HANA CO-PA principle [12][43]). This replaces/augments the manually maintained `profit_loss_analytics` and gives owners the standard "gross profit by item, store, hour" questions that sell an ERP.

**Operations sequencing:** P0 = costing engine + valuation ledger + COGS posting + goods issues; P1 = landed costs, BOM/production orders (issue/receipt), margin reports; P2 = routings/scrap, MRP-lite + purchase suggestions, consignment stock.

---

## 8. Platform standards & market trends (S/4HANA-derived)

What follows distills S/4HANA's *underlying standards* and 2024–2026 market expectations into concrete Nembus actions. Trend ratings: **mainstream** (must-have), **emerging** (adopt selectively), **hype** (watch). Verdicts summarize SMB applicability.

### 8.1 Extensibility & clean core — mainstream (the modern SaaS blueprint)

**SAP standard:** upgrade-stable key-user extensibility (custom fields/logic) + side-by-side extensions through released APIs; low-code workflows; everything REST/OData with a public API catalog [19][20][21][37].
**Nembus today:** JSONB `metadata` on nearly every table is an informal UDF mechanism; no formal registry, no tenant isolation of custom fields, no low-code workflow builder.
**Actions (P1):**
1. Formal **UDF registry**: `udf_definitions` (entity/table, field name, type, required, default, validation, UI hints) and values stored either in the row's `metadata` JSONB (fast to ship) or a typed `udf_values` table. Provide API + query filters over custom fields. This is the B1 "user-defined fields/tables" story without schema drift.
2. **Public, versioned API + catalog** (OpenAPI for the full surface, `/v1` prefixes, deprecation policy) and **webhook subscriptions** with the existing outbox as the event source (§8.3). Sandbox tenant provisioning for ISV/partner apps (marketplace later).
3. **Low-code workflow builder** (approvals/notifications) — ties to §8.2.

### 8.2 Approvals, alerts & workflow — mainstream

**SAP standard:** B1 approval procedures + alerts; S/4HANA workflows via SAP Build (conditions, multi-step, delegation) [21].
**Nembus today:** ad-hoc `approved_by` columns on POs, transfers, returns, stock counts; no workflow engine.
**Actions (P0/P1):** generic engine —
```
approval_templates (id, org, entity_type (po|transfer|return|ap_invoice|stock_count|price_change|…), name,
                    condition JSONB, step definitions JSONB, is_active)
approval_requests  (id, org, template_id, entity_type, entity_id, requester_id, status,
                    current_step, submitted_at, …)
approval_steps     (id, request_id, step_no, approver_role/user, status, decided_by, decided_at, comment, …)
```
with alerts/notifications (`notifications` table + email/push/webhook channels) and delegation. Every existing `approved_by` becomes an approval-request outcome.

### 8.3 Event-driven core, outbox & analytics plumbing — mainstream

**SAP standard:** event mesh, outbox/CDC, API hub; analytics on live data (embedded analytics / CDS virtual data model — no duplication) [17][18][20][21].
**Nembus today:** transactional outbox exists but only for POS→cloud sync; analytics are nightly batch aggregates.
**Actions (P1/P2):**
1. Generalize the outbox into a **domain event bus** (`event_outbox`: aggregate type/id, event type, payload, published_at, subscriber acks) — POS sync becomes one subscriber; webhooks, BI export, and third-party connectors become others.
2. **Virtual read models first**: S/4HANA-style — expose analytic/read views (aging, margin, stock value, sales by dimension) **directly over transactional tables via the API** (SQL views already exist; serve them through REST with filtering/pagination). Materialize only hot KPI aggregates.
3. Columnar/OLAP next (ClickHouse/DuckDB/Timescale [38] or partitioned PG) fed by the event stream — this is the "on-HANA-like" performance story; do **not** replicate SAP's appliance; PostgreSQL + columnar export + semantic views is the credible SMB answer.

### 8.4 AI & copilot (Joule-style) — emerging/mainstream; pick 2–3 bets

**SAP:** Joule copilot; AI in finance (intelligent cash application, collections), demand forecasting, anomaly detection, ML in inventory [22][23][30].
**Actions (P2):** (a) **NL→report copilot** over the semantic views of §8.3 (fastest value; secure via existing RBAC); (b) **demand forecasting** writing into the `forecast_demand` table of §7.4 (replenishment value); (c) **anomaly detection** on POS transactions (refund abuse, drawer variance) using existing analytics tables. Keep data-model readiness now (events + valuation + ledger dimensions), since all three need clean, dimensioned history rather than bespoke tables later.

### 8.5 Payments modernization — mainstream (commerce)

Tokenization/PCI, BNPL & wallets, multi-acquirer orchestration, unified commerce [33][34]. Actions fold into §6.5 (payment-method master, split tender, settlement reconciliation vs cashier sessions, digital receipts). No schema surprise here — but **payment orchestration behind one `payments` model** (instead of gateway-specific free-text) is the P1 enabler.

### 8.6 Subscription & recurring revenue — mainstream

Memberships and recurring billing + deferred revenue (ASC 606/IFRS 15 [10][11][35][36]) → §6.3 design. Note IFRS 18 (income-statement categories & management-defined measures) applies from FY2027 [39] — a reason the GL/reporting engine (§5) should be built with **grouped/definable reporting lines** from the start.

### 8.7 Compliance floor: e-invoicing, fiscalization, digital records — regulatory must

- KSA **ZATCA Phase 2** waves continue (24th group ≈ Oct 2025; 25th wave announced for 2026; SAR 375k threshold) [26][27] — Nembus is already positioned; keep wave onboarding/config tooling generic.
- **EU ViDA** (Directive (EU) 2025/516) phases digital reporting & e-invoicing through ~2030 [24][25][41]; **PEPPOL** dominates cross-border/public-procurement e-invoicing [41]. GCC (UAE etc.) and other Gulf mandates follow the ZATCA pattern.
- **Action (P1):** build a **tax & e-document engine** that generalizes ZATCA: tax-group determination → structured invoice output (UBL/PEPPOL/ZATCA XML) via format adapters, fiscal device/serialization hooks, and an append-only signed document store (`zatca_document_chain` generalizes). Never hard-code a country's format in business logic.

### 8.8 Identity & security standards — mainstream, non-negotiable for SaaS

OIDC/SAML SSO, MFA, SCIM provisioning, fine-grained RBAC→ABAC, audit & GDPR [40] → §4.18 actions (P1). Multi-tenant-database isolation is already a strength; missing pieces are protocol-level: OIDC integration, SCIM endpoints, per-customer data retention/privacy controls, and row-level authority beyond store scope.

### 8.9 Sustainability/ESG — hype at SMB (watch, keep data extensible)

CSRD phases SME applicability 2026–27 [28][29]. Action (P2): keep `metadata`/UDF + product/purchase fields able to carry emissions/energy/supplier-ESG data; ship a small "energy & waste" dashboard only if a flagship customer demands it (Nembus restaurant waste data is a natural seed).

### 8.10 Trend summary for Nembus (what to internalize now vs later)

| Trend | Rating | SMB relevance | Nembus action |
|---|---|---|---|
| Clean-core extensibility (UDF + public API + low-code) | mainstream | high | §8.1 (P1) |
| Approvals/alerts workflow | mainstream | high | §8.2 (P0/P1) |
| Event-driven + outbox + webhooks | mainstream | high | §8.3 (P1) |
| Embedded analytics / semantic read models | mainstream | high | §8.3 (P1) |
| Identity (OIDC/SCIM/MFA) | mainstream | high | §8.8 (P1) |
| e-invoicing/fiscalization | mainstream (regulatory floor) | high | §8.7 (P1) |
| Payments modernization | mainstream | high | §6.5 (P1) |
| Subscription & recurring revenue | mainstream | med-high | §6.3 (P1) |
| AI copilot / forecasting / anomaly | emerging | med (selective) | §8.4 (P2) |
| Columnar/OLAP analytics layer | emerging | med | §8.3 (P2) |
| ESG / carbon | hype (enterprise emerging) | low | §8.9 (P2+) |

---

## 9. Prioritized roadmap

Sequencing logic: **P0 builds the spine** (ledger, periods, series, approvals, COGS) because every later module (AP, assets, budgeting, AI margin analysis) hangs off it. P1 harvests the highest-value commerce/operations/platform features once the spine exists. P2 contains differentiators that need clean history first. Efforts are relative (S/M/L); each item lists the schema/feature work and an acceptance criterion.

### Phase P0 — Financial backbone & platform prerequisites (do first; highest business value)

| # | Item | Why | Key schema/features | Effort | Acceptance criterion |
|---|---|---|---|---|---|
| 1 | Posting periods & fiscal years | Finance cannot exist without period control | `fiscal_years`, `posting_periods`, open/close rules (§5.2) | S | Can open/close a period; postings to closed period rejected |
| 2 | Document numbering series | Replace free-text numbers; audit & period-based counters | `document_series` + wiring into doc create/post (§5.2) | S | Every financial doc number auto-assigned, gap-free, per series |
| 3 | Ledger line store + posting service | Universal-Journal spine | Extend `journal_entries/_lines` w/ dimensions, base amounts, period, series, immutable postings; posting service & idempotency keys (§5.2–5.3) | L | Double-entry invariant holds; re-post idempotent; every P0 doc posts |
| 4 | GL mappings rule engine | Finance team controls accounts w/o code | Extend `gl_account_mappings` (§5.2) | S | Change mapping → next posting uses new account |
| 5 | COA groups + financial report views | Balance sheet/P&L/trial balance | COA group fields; `vw_trial_balance`, `vw_balance_sheet`, `vw_profit_and_loss` (§5.4) | M | Views reconcile to posting totals; period-end P&L matches analytics |
| 6 | Posting map: POS, invoice, payment, GRN, return | Core docs hit the books | Posting for existing events incl. outbox replay (§5.3) | L | POS receipt → GL → P&L golden test green |
| 7 | AR/AP open items & allocations | Real receivables/payables | `invoice_allocations`, unified `payments`, aging from open items, basic dunning (§5.2) | L | Payment allocates across invoices; aging correct |
| 8 | Approval workflow engine | All approval columns become real workflows | `approval_templates/requests/steps` + notifications (§8.2) | M | PO/return/transfer approval via workflow, not ad-hoc flag |
| 9 | Costing engine + COGS posting | Operations & margins depend on it | costing method on product, valuation ledger, moving-avg recompute, COGS post (§7.1) | L | COGS per sale equals method-derived cost; stock value report correct |
| 10 | Goods issue documents | Non-sales stock-out + GL | `goods_issues(_items)` (§7.3) | M | Issue → stock down + expense posted |

**P0 exit criteria:** a trial balance, balance sheet and P&L produced from live POS/GRN/returns data; a payment allocated to invoices; a PO approved via workflow; finance user can close a month.

### Phase P1 — Commerce, operations & platform depth

| # | Item | Why | Key schema/features | Effort |
|---|---|---|---|---|
| 11 | Business Partner unification + groups | One AR+AP identity (CVI idea) | Partner roles, BP groups/properties, customer-compatible view (§4.1) | L |
| 12 | AP invoices + 3-way match | Procurement financial close | `ap_invoices`, match PO↔GRN↔AP, AP aging (§4.5) | L |
| 13 | Landed costs | Distributor/import must-have | landed-cost docs & allocations on GRN (§7.2) | M |
| 14 | Value store: gift cards & store credit | Retail monetization | `customer_value_accounts/_transactions`, `gift_card_codes` (§6.1) | M |
| 15 | Return lifecycle & RMA | Returns orchestration | extend `sales_returns` + cross-store restock + credit-to-value (§6.2) | M |
| 16 | Subscriptions & recurring billing | Memberships/recurring revenue | plans/customer subscriptions/recurring runs (§6.3) | L |
| 17 | B2B: quotations + credit management | Wholesale growth | quote lifecycle (schema exists), credit checks on order, delivery→periodic invoice (§6.7) | M |
| 18 | Cash-drawer operations & payment-method master | Chain cash mgmt | `payment_methods`, drawer ops, settlement vs cashier session (§6.5) | M |
| 19 | UDF registry + public API/versioning + webhooks | Clean-core extensibility | `udf_definitions`, `/v1`, event outbox → webhooks (§8.1, §8.3) | L |
| 20 | SSO/OIDC/SCIM/MFA | Enterprise sales floor | identity protocols (§4.18) | M |
| 21 | Tax engine & multi-region e-invoice framework | Regulatory floor (ViDA, GCC waves) | generalize ZATCA → tax groups + format adapters (§8.7) | L |
| 22 | Bank accounts, statements & reconciliation workspace | Cash visibility | `bank_accounts/statements/lines`, match & clear (§5.2) | M |
| 23 | Production: BOM generalization + production orders | Light manufacturing | multi-level BOM, `production_orders`, issue/receipt (§7.4) | L |
| 24 | Margin & profitability views (CO-PA-lite) | Sell the ERP | margin views over ledger/valuation (§7.5) | M |
| 25 | Embedded analytics read APIs | Dashboards in-app | serve analytic views via REST (§8.3) | M |

### Phase P2 — Analytics platform, AI & advanced modules

| # | Item | Why | Key schema/features | Effort |
|---|---|---|---|---|
| 26 | Columnar/OLAP layer fed by events | "On-HANA-like" real-time story | event-driven export to columnar store; KPI cubes (§8.3) | L |
| 27 | NL→report copilot | AI differentiator | over §25 semantic views, RBAC-scoped (§8.4) | M |
| 28 | Demand forecasting → replenishment | Inventory value | `forecast_demand` + AI models (§8.4, §7.4) | M |
| 29 | Anomaly detection (POS/refunds/drawers) | Fraud & ops | ML over analytics (§8.4) | M |
| 30 | MRP-lite + purchase suggestions | Beyond reorder points | net-requirements engine, `planning_suggestions` (§7.4) | M |
| 31 | Fixed assets & depreciation | Scale to distributors | register, depreciation runs → GL (§4.10, §5.2) | M |
| 32 | Budgeting & scenarios | Finance depth | `budgets` + variance views (§5.2) | M |
| 33 | Loyalty tiers & liability accounting | Loyalty 2.0 | tiers/history, points liability GL (§6.4) | M |
| 34 | Revenue recognition schedules | IFRS 15/ASC 606 for subs | deferred revenue schedules (§6.3) | M |
| 35 | ESG-lite & carbon fields | Watch-list compliance | extensible product/partner fields, waste/energy dashboard (§8.9) | S |

### Roadmap notes

- **Dependencies:** P1 items 12, 13, 22 depend on P0 ledger & allocations; 14–16 depend on unified payments/P0; 23 depends on 9 (costing). Most P2 AI/analytics items depend on 25 and clean history from P0.
- **Parallel tracks:** platform track (1,2,8,19,20) can proceed independently of finance item details; POS-client work (18, drawer ops; offline outbox generalization) is separable.
- **Measure success by:** (P0) finance-ready demo — real P&L from live POS; (P1) a "quotation→order→delivery→invoice→payment→reconciliation→P&L" customer journey and an ISV building on the public API; (P2) analytics latency, forecast accuracy, support tickets avoided via copilot.

---

## 10. Risks & architectural considerations

1. **Schema churn is the top risk.** The repo already mixes v1 (`sales_orders`, `purchase_orders`) and v2 (`sales_orders_v2`, `carts`) models, `SERIAL` vs `UUID` ids, and duplicated query/repository trees (`packages/core` vs `apps/cloud-server` vs POS client copies). Adding a ledger multiplies that unless **v2 becomes the canonical spine** and v1 tables are retired (views/migration) — otherwise every new feature pays a 3× codebase tax. Address duplication (single source for schema + queries, generated consumers) before or with P0.
2. **Ledger data volume & immutability.** Append-only, dimension-bearing lines at POS scale grow fast and must never be `UPDATE`d. Mitigations: partition by `fiscal_year`/month, index strategy in §5.2, and strict service-level write rules (no direct SQL writes in production paths). Consider columnar export (§8.3) before the OLTP ledger becomes the BI bottleneck.
3. **Backfilling & cut-over.** Existing documents have free-text numbers and no posting history. Define a one-time **opening-balance migration** (post opening entries per COA + open AR/AP items by aging from current `outstanding_balance`) and keep legacy numbering series readable. Do not attempt to reconstruct full retroactive ledgers beyond a stated cut-over date.
4. **Offline POS → finance correctness.** Posting must be **idempotent** (source-unique constraint) because offline POS replays via outbox; a period-close race (store syncing an old transaction into a closed period) needs a defined policy (post to current period with original posting date in memo, or reject & alert).
5. **Accounting expertise is a feature.** GL design, tax engine and financial reports need a domain person (or advisor) validating posting maps, tax treatment (inclusive/exclusive, withholding), and report templates; this is not a pure engineering backlog item.
6. **Scope discipline against the benchmarks.** SAP B1 and S/4HANA are enormous. The roadmap deliberately skips HR/payroll engines, full EWM, parallel ledgers/multi-GAAP breadth, RAR-scale revenue accounting, PP/DS scheduling and ESG platforms. Revisit only if a flagship customer pays for it.
7. **Multi-tenancy & per-tenant finance customization.** Each tenant DB gets its own COA/periods/series — good isolation, but migrations must handle N live tenants (rolling apply, no downtime), and tenant-specific COA/GL-mapping configs need backup/restore and clone tooling (the repo already has tenant tooling to extend).
8. **Money handling.** Standardize DECIMAL types and rounding policy per document (rounding account), define base-currency per organization (currently `currency_code` defaults are scattered: 'SAR'/'USD'), and enforce one place for FX rates (`exchange_rates` exists) with a revaluation job — before multi-currency invoices spread further.
9. **Feature creep through JSONB.** `metadata` on every table is convenient but unqueryable at scale and ungoverned; the UDF registry (§8.1) should own custom fields while internal code moves to typed columns. Decide when P0.
10. **Team & sequencing.** P0 is a multi-month, multi-person effort even at SMB scope; recommend a dedicated finance-platform squad, feature flags for posting (shadow-post in parallel to current analytics before switching sources of truth), and demo-driven milestones (finance-ready demo first).

---

## Appendix A — Nembus schema inventory

Complete object inventory derived from the canonical schema (`packages/core/db/schema/*.sql`), the POS-client migration, plus the main views/functions. This is the reference the gap analysis cites.

### A.1 Tables by domain

**Identity, RBAC & settings (`10_identity_rbac.sql`)** — `organizations`, `tenants`, `modules`, `menus`, `submenus`, `permissions`, `module_permissions`, `menu_permissions`, `submenu_permissions`, `roles`, `role_permissions`, `ui_settings`, `role_ui_customizations`; analytics snapshots `profit_loss_analytics`, `discount_analytics`.

**Stores, users, cash & terminals (`20_stores_terminals.sql`)** — `stores`, `storage_locations`, `users`, `user_roles`, `user_store_access`, `cashiers`, `pos_terminals`, `cashier_sessions`.

**Catalog (`30_catalog.sql`)** — `product_categories`, `brands`, `units_of_measure`, `uom_packaging_templates`, `uom_packaging_template_levels`, `price_lists`, `tax_categories`, `products`, `product_variants`, `product_barcodes`, `product_prices`, `product_uom_conversions`, `product_serial_numbers`, `product_batches`, `stock_reservations`.

**Inventory (`40_inventory.sql`)** — `inventory_stock`, `stock_movements`, `stock_counts`, `stock_count_lines`.

**Partners, purchasing, GL primitives & sales v1 (`50_purchasing_suppliers.sql`)** — `currencies`, `exchange_rates`, `payment_terms`, `cost_centers`, `chart_of_accounts`, `gl_account_mappings`, `business_partners`, `partner_addresses`, `partner_contacts`, `customers`, `purchase_orders`, `purchase_order_lines`, `transfer_requests`, `transfer_request_items`, `goods_receipt_notes`, `goods_receipt_note_items`, `sales_orders`, `sales_order_lines`, `carts`, `cart_items`, `cart_activity_log`, `bp_price_contracts`, `journal_entries`, `journal_lines`. (Enums: `order_type`, `order_status_v2`, `payment_status`, `fulfillment_status`, `cart_status`, `cart_type`.)

**Sales/OMS v2 (`60_sales_pos.sql`, `85_zatca.sql`)** — `draft_cart_templates`, `draft_cart_template_items`, `sales_orders_v2`, `sales_order_lines_v2`, `order_status_history`, `order_fulfillments`, `order_fulfillment_items`, `quotes`, `quote_lines`, `invoices`, `invoice_lines`, `invoice_payments`, `invoice_status_history`, `sales_returns`, `sales_return_lines`. (Enums: `invoice_type`, `invoice_status`, `quote_status`.)

**POS & restaurant (`70_restaurant.sql`)** — `pos_transactions`, `pos_transaction_lines`, `pos_payments`, `restaurant_tables`, `menu_categories`, `menu_items`, `menu_item_modifiers`, `menu_modifier_groups`, `promotions`, `recipes`, `recipe_ingredients`, `combo_bundles`, `combo_bundle_items`, `menu_item_availability_schedules`, `restaurant_orders`, `restaurant_order_items`, `waste_logs`, `kiosk_sessions`.

**Loyalty & audit (`80_promotions_loyalty.sql`)** — `sales_analytics`, `purchase_analytics`, `inventory_analytics`, `loyalty_redemption_rules`, `audit_logs`.

**ZATCA & sync (`85_zatca.sql`)** — `zatca_device_configs`, `zatca_document_chain`, `sync_watermarks`.

**POS-client local (`apps/pos-client/migrations/99999999999999_pos_extensions.sql`)** — `local_printer_configs`, `sync_queue` (transactional outbox), `local_device_config`.

### A.2 Main views & functions (`90_views_functions.sql` and schema-local)

- **Totals/triggers:** `calculate_order_totals`, `calculate_invoice_totals`, `update_invoice_payment`, `update_updated_at_column` (+ ~30 per-table `updated_at` triggers), `log_cart_status_change`, `update_cart_activity`.
- **Operational views:** `vw_pos_product_catalog`, `vw_low_stock_alerts`, `vw_pending_purchase_orders`, `vw_customer_aging_report`, `vw_accounts_payable`, `vw_profit_margin_analysis`, `vw_user_effective_permissions`, `vw_pos_categories`, `vw_restaurant_menu`, `vw_recipe_bom`, `vw_active_restaurant_orders`, `vw_waste_daily_summary`, `v_master_product_catalog`.
- **POS/functions:** `fn_pos_get_products_with_stock`, `fn_pos_get_product_by_barcode`, `fn_pos_get_products_by_category`, `fn_pos_search_products`, `fn_get_restaurant_menu`, `fn_get_item_modifiers`, `fn_get_kds_orders`, `get_master_product_catalog`.
- **Inventory/functions:** `fn_process_stock_transfer`, `fn_convert_uom_quantity`, `fn_approve_transfer_request`, `fn_ship_transfer_request`, `fn_receive_transfer_request`, `fn_process_goods_receipt`, `fn_reconcile_stock_count`, `fn_trigger_deduct_inventory_on_fulfillment`, `fn_trigger_allocate_inventory_on_order_line`, `fn_log_transfer_request_history`, `fn_sync_promotion_to_product_prices`.
- **Analytics/costing:** `fn_refresh_daily_analytics`, `fn_calculate_loyalty_earned`, `fn_calculate_recipe_cost`, `fn_get_waste_report`.

### A.3 Status legend used throughout this document

✅ delivered API (handler/usecase/routing) · 🔶 schema + queries, partial or no REST surface · ⬜ schema only. Verified at the time of writing by scanning `packages/core/routing/*` and `handler/*`; treat "not found in routing" as "verify before relying on it", not as proof of absence in another entry point.

---

## Appendix B — SAP ↔ Nembus concept mapping

| SAP B1 / S/4HANA concept | Nembus analog today | Status / gap |
|---|---|---|
| Business Partner (customer/vendor) [CVI] | `business_partners` + separate `customers` | 🔶 split identities; unify w/ roles (§4.1) |
| Item master / item group | `products`, `product_categories` | ✅ categories; item-group defaults missing |
| UoM groups | `units_of_measure` + `product_uom_conversions` + packaging templates | ✅ |
| Price lists / special prices / discount groups | `price_lists`, `product_prices`, `bp_price_contracts`, `promotions` | ✅ list engine; 🔶 contract/pricing-context resolver |
| Warehouses / bin locations | `stores`, `storage_locations` | ✅ |
| Serial & batch mgmt | `product_serial_numbers`, `product_batches` | ✅ |
| Item costing / valuation method | cost fields on movements/lines only | ⬜ costing engine missing (§7.1) |
| Goods receipt PO | `goods_receipt_notes(_items)` + `fn_process_goods_receipt` | ✅ |
| Goods issue / delivery note | `order_fulfillments` (sales ship); no internal issue | 🔶 add `goods_issues` (§7.3) |
| Inventory transfer | `transfer_requests` + lifecycle functions | ✅ |
| Cycle/stock counting | `stock_counts(_lines)` + reconcile fn | 🔶 no REST surface found |
| Landed costs | — | ⬜ (§7.2) |
| Purchase order / quotation / AP invoice | `purchase_orders` (schema); PO API absent; no AP invoice | 🔶 (§4.5) |
| Sales quotation → order → delivery → AR invoice → return | `quotes`, `sales_orders_v2`, `order_fulfillments`, `invoices`, `sales_returns` | ✅ process coverage; ⬜ financial consequences |
| AR/AP open items & payment allocation | `invoice_payments` (per-invoice), balances on partner | 🔶 allocations/clearing missing (§5.2) |
| Dunning / credit mgmt | `vw_customer_aging_report`; credit limit columns | 🔶 dunning & credit-block workflow missing |
| Bank accounts, checks, reconciliation | placeholder `bank_account_id` | ⬜ (§5.2) |
| Fixed assets & depreciation | — | ⬜ P2 (§4.10) |
| Production orders / BOM / MRP | `recipes`/`recipe_ingredients` (restaurant), `combo_bundles` | 🔶 generalize → manufacturing (§7.4) |
| CRM: opportunities/activities/service/campaigns | sales-order assignment only | ⬜ P2 (§4.12) |
| GL: COA groups, posting periods, series | `chart_of_accounts`, `cost_centers`, `journal_entries/_lines` | 🔶 schema; ⬜ engine (§5) |
| Budget & scenarios | — | ⬜ P2 (§5.2) |
| Universal Journal / dimensioned line store | `journal_entries/_lines` | ⬜ evolve to line store (§5.2) |
| Financial reports (BS/P&L/CF/trial balance) | P&L analytics table + margin views | 🔶 views only; engine missing (§5.4) |
| Multi-currency / revaluation | `currencies`, `exchange_rates`, per-doc currency | 🔶 no base-currency discipline/revaluation (§5.1, §10.8) |
| Tax codes/groups, withholding | `tax_categories` single rate | 🔶 tax engine (§8.7) |
| e-invoicing/fiscalization | ZATCA Phase-2 (device, chain, XML/crypto) | ✅ KSA; generalize (§8.7) |
| Document numbering series | free-text `*_number` columns | ⬜ (§5.2) |
| Approval procedures / workflows | `approved_by` FKs | ⬜ engine (§8.2) |
| Alerts & notifications | — | ⬜ (§8.2) |
| User-defined fields & tables | JSONB `metadata` | 🔶 formalize UDF registry (§8.1) |
| Authorizations/row-level | RBAC roles + store access | 🔶 extend to row/record scope (§4.18) |
| Audit trail | `audit_logs` (no triggers) | 🔶 wire generically |
| Service Layer / DI API (B1) | REST/JSON (Gin) + OpenAPI | ✅ style differs; OData/webhooks later (§8.3) |
| Embedded analytics / CDS (S/4) | SQL views + pre-aggregated tables | 🔶 virtual read models via API (§8.3) |
| Outbox / eventing (S/4 BTP) | POS `sync_queue` outbox | ✅ pattern; generalize to domain events (§8.3) |
| Clean-core extensibility | JSONB + m2m tokens | 🔶 UDF + public versioned API (§8.1) |

---

## Appendix C — References

Numbered inline citations in §7.5 and §8 refer to the S/4HANA & market-trend sources below (group C1); SAP Business One sources are added in group C2. Sources were gathered via web research (official SAP/tax-authority pages preferred, corroborated with reputable secondary sources) at the time of writing; help.sap.com pages are JS-rendered, so their titles were verified via search and content cross-checked against secondary sources.

### C1 — S/4HANA concepts & ERP market trends (2024–2026)

1. SAP Help — Universal Journal in SAP S/4HANA Cloud: https://help.sap.com/docs/SAP_S4HANA_CLOUD/0fa84c9d9c634132b7c4abb9ffdd8f06/523b8a55559ad007e10000000a44538d.html
2. SAP Help — Universal Journal (S/4HANA on-premise): https://help.sap.com/docs/SAP_S4HANA_ON-PREMISE/3eb1567cf97543c08087efb0936964e6/8b8e5695c4dc4749a706f9fa2f6bda92.html
3. SAP PRESS blog — What is SAP's Universal Journal: https://blog.sap-press.com/what-is-saps-universal-journal
4. SAP Activate — S/4HANA Customer Vendor Integration cookbook (PDF): https://support.sap.com/content/dam/SAAP/SAP_Activate/S4H.0781%20SAP%20S4HANA%20Cookbook%20Customer%20Vendor%20Integration.pdf
5. CDQ — SAP Customer-Vendor Integration (CVI) explained: https://www.cdq.com/blog/sap-customer-vendor-integration-cvi
6. SAP Help — Business Partner in S/4HANA Cloud: https://help.sap.com/docs/SAP_S4HANA_CLOUD/a376cd9ea00d476b96f18dea1247e6a5/0871bd534f22b44ce10000000a174cb4.html
7. SAP Help — Ledger approach / parallel ledgers: https://help.sap.com/docs/SAP_S4HANA_CLOUD/0fa84c9d9c634132b7c4abb9ffdd8f06/fbf38d5377a0ec23e10000000a174cb4.html
8. SAP Help — Asset Accounting (S/4HANA Cloud): https://help.sap.com/docs/SAP_S4HANA_CLOUD/3e5fcf2c768746049b5627bd5a42f720/733dde531ed3424de10000000a174cb4.html
9. SAP Best Practices — Asset Accounting, group ledger IFRS: https://help.sap.com/docs/s4hana-cloud-best-practices/asset-accounting-group-ledger-ifrs-1gb-vn/overview-table
10. Bramasol — Revenue Accounting and Reporting (RAR): https://www.bramasol.com/products/revenue-accounting-and-reporting-rar/
11. SAP PRESS blog — How SAP handles revenue recognition: https://blog.sap-press.com/how-does-sap-handle-revenue-recognition
12. SAP Help — Profitability Analysis (CO-PA): https://help.sap.com/docs/SAP_S4HANA_ON-PREMISE/5e23dc8fe9be4fd496f8ab556667ea05/f813d254fb26c50ae10000000a441470.html
13. SAP Help — MRP Live: https://help.sap.com/docs/SAP_S4HANA_ON-PREMISE/f899ce30af9044299d573ea30b533f1c/86e15c58eb021f60e10000000a44147b.html
14. Arion ERP — SAP manufacturing modules overview: https://www.arionerp.com/news/productivity/sap-manufacturing-modules.html
15. SAP Help — Extended Warehouse Management: https://help.sap.com/docs/SAP_EXTENDED_WAREHOUSE_MANAGEMENT/25cf88dfa94c49e4a440f3f1d752b8a1/2ececb53ad377114e10000000a174cb4.html
16. SAP Help — Warehouse Management (S/4HANA on-premise): https://help.sap.com/docs/SAP_S4HANA_ON-PREMISE/9832125c23154a179bfa1784cdc9577a/4cb4cf0d0c056642e10000000a15822b.html
17. SAP Learning — Describing embedded analytics in S/4HANA Cloud: https://learning.sap.com/courses/implementing-sap-s-4hana-cloud-public-edition/describing-embedded-analytics_bd89eeb8-5bb6-43a7-a8aa-906a4d32a3c4
18. SAP Help — Core Data Services (CDS) in S/4HANA: https://help.sap.com/docs/SAP_S4HANA_ON-PREMISE/8308e6d301d54584a33cd04a9861bc52/5418de55938d1d22e10000000a44147b.html
19. SAP — Clean core strategy (document): https://www.sap.com/suisse/documents/2024/09/20aece06-d87e-0010-bca6-c68f7e60039b.html
20. SAP Help — Extensibility in S/4HANA Cloud (key-user extensibility): https://help.sap.com/docs/SAP_S4HANA_CLOUD/0f69f8fb28ac4bf48d2b57b9637e81fa/c3a5da91691d4ebd89748d9f40af7a4c.html
21. SAP Developers — Integration Suite, advanced event mesh (mission): https://developers.sap.com/tutorials/mission-get-started-with-sap-integration-suite-advanced-event-mesh
22. SAP Learning — SAP Business AI for S/4HANA Cloud (AI use cases): https://learning.sap.com/courses/introducing-sap-business-ai-for-sap-s-4hana-cloud/exploring-ai-use-cases-in-sap-s-4hana-private-cloud
23. ERP Today — Five AI trends affecting enterprise scaling: https://erp.today/five-ai-trends-affecting-enterprise-scaling/
24. Agencia Tributaria (ES) — VAT Digital Age Directive (EU) 2025/516: https://sede.agenciatributaria.gob.es/Sede/en_gb/iva/novedades-iva/novedades-normativa-2025/directiva-2025-516-consejo-11-2025.html
25. EU Council — Directive (EU) 2025/516 (ViDA) text (PDF): https://data.consilium.europa.eu/doc/document/ST-12492-2025-INIT/en/pdf
26. KPMG — Saudi Arabia mandatory e-invoicing, 24th group: https://kpmg.com/us/en/taxnewsflash/news/2025/10/saudi-arabia-mandatory-e-invoicing-24th-group-taxpayers.html
27. EY — Saudi Arabia announces 25th wave of Phase 2 e-invoicing: https://taxnews.ey.com/news/2026-1705-saudi-arabia-announces-25th-wave-of-phase-2-e-invoicing-integration
28. SAP Learning — SAP Sustainability solutions: https://learning.sap.com/courses/positioning-sap-sustainability-solutions-ko/demonstrating-the-end-to-end-capabilities-of-sap-s-sustainability-solutions
29. BearingPoint — SAP Sustainability Control Tower: https://bearingpoint.services/emissions-calculator/en/services/co2-software-implementation/sap-sustainability-control-tower/
30. Gartner reprint — (AI trends) https://www.gartner.com/doc/reprints?id=1-2M3AUSGM&ct=251013&st=sb
31. Kibo Commerce — retail omnichannel: https://kibocommerce.com/solutions/retail/
32. Congruence Market Insights — headless commerce market: https://www.congruencemarketinsights.com/report/headless-commerce-market
33. Fexco — payment orchestration (PayUnite): https://www.fexco.com/news-and-insights/fexco-launches-payunite-independent-payment-orchestration-supporting-business-growth/
34. Juspay — payment orchestration, control plane for global commerce: https://juspay.io/blog/payment-orchestration-the-control-plane-for-global-commerce
35. Zuora — Subscription Economy Index: https://www.zuora.com/resource/subscription-economy-index/
36. FinTech Magazine — Zuora expands Workday tie-up for revenue management: https://fintechmagazine.com/news/zuora-expands-workday-tie-up-for-revenue-management
37. e3 Magazine — Analysis of clean-core approaches for S/4HANA migration: https://e3mag.com/en/analysis-of-clean-core-approaches-for-sap-s-4-hana-migration/
38. GrowthX tech — Choosing analytics DB: TimescaleDB / DuckDB / ClickHouse: https://tech.growthx.ai/posts/choosing-your-analytics-database-timescaledb-duckdb-or-clickhouse
39. FM Magazine — How companies can prepare for IFRS 18 adoption: https://www.fm-magazine.com/issues/2026/mar/how-companies-can-prepare-for-ifrs-18-adoption/
40. Microsoft Learn — Identity governance & app integration (OIDC/SCIM baseline): https://learn.microsoft.com/en-us/entra/id-governance/identity-governance-applications-integrate
41. Peppol.nl — Proposed legislation VAT in the Digital Age (ViDA) approved: https://www.peppol.nl/en/news/proposed-legislation-vat-digital-age-vida-approved
42. Sukisoft — How mobiles enable ERP access: https://sukisoft.com/blog/mobile-apps/how-mobiles-are-enabling-erp-access-for-modern-enterprises/
43. SAP PRESS blog — Account-based CO-PA in S/4HANA (margin analysis in the Universal Journal): https://blog.sap-press.com/account-based-co-pa-in-sap-s4hana-how-margin-analysis-works-in-the-universal-journal

### C2 — SAP Business One (incl. SAP Business One on HANA)

*(pending final research pass — inserted on completion)*
