# ZATCA from sales_orders_v2

ZATCA now generates simplified tax invoices from persisted `sales_orders_v2` and `sales_order_lines_v2`. No new order, payment, or fulfillment status is required. An order header with all lines and consistent final amounts is sufficient. Restaurant-specific settlement is outside this change.

## Run in simulation

Deploy the updated cloud and POS binaries, then enable the organization using `PUT /api/organizations/ID/zatca` with `{"enabled":true}`. The selected device's saved `zatca_env` controls its gateway; use `GET /api/zatca/status?device_id=3` to confirm simulation. See [device environment routing](ZATCA_DEVICE_ENVIRONMENT.md). Allow the POS pull sync to bring down the organization metadata and device row, then confirm the POS local tenant database has the enabled setting, production CSID, matching secret, and signing key before checkout.

Use `Authorization: Bearer <token>` and `x-tenant-id: stgin` on each request. Import `../postman-collection/ZATCA-Sales-Order-Test.postman_collection.json` and set `sales_order_id` to the UUID from `sales_orders_v2.id` (not the order number, POS transaction ID, or device ID).

Find a candidate in the tenant DB:

```sql
SELECT o.id, o.order_number, o.organization_id, o.store_id, o.pos_terminal_id,
       o.subtotal, o.discount_amount, o.tax_amount, o.total_amount,
       COUNT(l.id) AS lines
FROM sales_orders_v2 o
LEFT JOIN sales_order_lines_v2 l ON l.sales_order_id = o.id
GROUP BY o.id
ORDER BY o.created_at DESC
LIMIT 20;
```

1. `GET /api/orders/ORDER_UUID` — inspect the saved header and lines.
2. On the POS local API (`localhost:8080`), save every line, then `PUT /api/orders/ORDER_UUID/totals` as the final direct-order step. This call attempts generation and immediate reporting when the complete order is valid, the organization switch is enabled, and the selected device is reporting ready. Line creation alone never issues an invoice. Cart and POS checkout also attempt reporting after saving their order. Inspect the top-level `zatca.reported` and `zatca.zatca_response.reportingStatus` in the final save/checkout response.
3. `POST /api/zatca/orders/ORDER_UUID/submit` — **manual retry only**. No body. The POS local route retries from the local document chain and local device credentials. The cloud route remains available for documents present in the cloud tenant database. A successful response has `data.zatca_response.reportingStatus: REPORTED`. Repeating a successful submit does not report it again.
4. `GET /api/zatca/receipt-qr?sales_order_id=ORDER_UUID` — retrieves the same saved QR even if no POS transaction exists. Receipt rendering encodes this exact Base64 TLV text. Read the `zatca.reported` result before treating the QR as proof of submission.

`POST /api/zatca/orders/ORDER_UUID/generate` remains available for local signing without gateway reporting. `POST /api/zatca/orders/ORDER_UUID/check` uses compliance credentials and does not report the sale. A fully-onboarded device may receive ZATCA's `Submitted before` / already-completed compliance response. Both `/check` and `/submit` generate the document first if needed.

## Automatic generation

Cart conversion generates the document after all order lines are saved. POS transaction checkout already creates a sales order and signs that persisted order instead of transaction copies. The POS local API reports immediately on final totals save, cart checkout, or POS checkout and saves the gateway result in local `zatca_document_chain`. The cloud handlers no longer trigger automatic reporting. The existing `data` shape is preserved and a top-level `zatca` object reports `generated`, `reported`, QR/UUID, the saved gateway response, or an `error`. A successful business save remains HTTP 200/201 even if ZATCA reporting fails; the frontend must inspect `zatca.reported`. Fix the order if generation failed, or retry the same order through the POS `/submit` route if the gateway failed. Never create a second sale to retry reporting. Cloud syncing the locally saved document is separate work.

Direct SQL inserts do not execute Go hooks. Once SQL has inserted the header and complete lines, `/generate`, `/check`, or `/submit` can generate the document. No extra status needs to be changed.

## Source mapping and preservation

The invoice number uses `order_number`, the lines use their saved product names, quantities, prices, discounts and taxes, and the invoice issue timestamp is captured once when signing. Seller data still comes from organization/store metadata and the VAT number from the organization. Device selection uses the order terminal and organization, with the existing cloud-device fallback. Tenant selection remains request-scoped.

Order and device rows are locked inside a database transaction for document creation. The device's ICV/PIH continues from its existing chain. Retries return the existing XML, hash, UUID and QR unchanged, even if source rows have subsequently changed. Business corrections must follow the credit/debit-note process rather than regenerating an issued document. This change does not introduce a blanket restriction on order editing APIs.

Existing transaction-based signed documents remain intact. If the order is linked to such a transaction, the order endpoints reuse the historical document, preserving its original invoice number, UUID, QR, status and gateway response. They do not issue a second invoice. Historical `/transactions/:id/check` and `/submit` remain available. Returns linked through POS transactions now resolve the original number from the actual issued XML, including order-based invoices.

No schema migration or chain reset is needed: `entity_type` already accepts text and `entity_id` already stores text/UUIDs. Keep old chain entries and device credentials as they are.

## Current invoice scope

This migration preserves the tested **SAR simplified invoice / standard 15% VAT** scope. `line_total` must include VAT and equal quantity × price − line discount + line VAT. Order subtotal minus order discount plus VAT must equal total. Additional header discounts are allocated over lines and validated against the final stored tax/total; line discounts are not applied twice.

Nonzero shipping, adjustments, promotional credits, unsupported tax rates or a non-SAR organization currency return explicit validation errors. These require dedicated charge/tax mappings; they are not silently omitted. B2B clearance expansion and restaurant-specific sales-order creation are separate work.

## Verification

Local tests cover order/line mapping, absence of a status gate, signed XML/hash consistency, discounts, invalid/partial data without chain insertion, immutable retries, historical document reuse, original return references, order check/report credential selection, reporting persistence/idempotency, and receipt lookup. Tests use local fixtures/mock gateways; actual new order XML still needs the user's simulation test.
