# Nembus SAP Sync Agent (`apps/sap-agent`)

The **SAP Sync Agent** is a bidirectional synchronization service between **SAP Business One (SAP B1)** and the **Nembus Ecosystem** (Cloud ERP & POS Desktop Client). It supersedes the legacy `SAP_To_Vtec` .NET application.

---

## Capabilities

1. **Downstream Pipeline (SAP B1 $\rightarrow$ Nembus Master Data)**
   - **Categories & Groups**: Synchronizes SAP Item Groups (`OITB`) to Nembus `product_categories`.
   - **Units of Measure**: Synchronizes SAP UOMs (`OUOM`) to Nembus `units_of_measure`.
   - **Products**: Synchronizes SAP Items (`OITM`) to Nembus `products` (SKU, names, Arabic foreign names, scale/weight flags).
   - **Barcodes**: Synchronizes primary and multi-barcodes (`OBCD`) to Nembus `product_barcodes`.
   - **Price Lists**: Synchronizes price lists (`ITM1`) to Nembus `product_prices`.
   - **Inventory**: Syncs warehouse stock levels to Nembus `inventory_stock`.

2. **Upstream Pipeline (Nembus POS $\rightarrow$ SAP B1 Invoices & Payments)**
   - **Outbox Watcher**: Monitors completed POS transactions and outbox items (`sync_queue`).
   - **AR Invoices (`OINV`)**: Automatically creates official SAP AR Invoices with line items, VAT calculations, and POS transaction references.
   - **Incoming Payments (`ORCT`)**: Posts cash and card payments linked directly to the created SAP AR Invoice.
   - **Idempotency**: Tags Nembus transactions with SAP `DocEntry` and `DocNum` to prevent duplicate postings.

---

## Configuration

Copy `.env.example` to `.env` in `apps/sap-agent/`:

```bash
cp .env.example .env
```

Key environment variables:
- `SAP_SL_URL`: SAP Service Layer endpoint (e.g. `https://192.168.1.90:50000/b1s/v1`)
- `SAP_COMPANY_DB`: Target SAP company database (e.g. `QITAFALAELAHLTD`)
- `SAP_USERNAME` & `SAP_PASSWORD`: SAP B1 credentials
- `NEMBUS_DB_URL`: Local POS database (port 5433) or Cloud server (port 5432)
- `NEMBUS_ORGANIZATION_ID`: Target tenant/organization ID (default: `1`)
- `DOWNSTREAM_INTERVAL_SECONDS`: Frequency of master data refresh (default: `300s` / 5m)
- `UPSTREAM_OUTBOX_POLL_MS`: Frequency of transaction outbox check (default: `3000ms` / 3s)

---

## Quick Commands

From the monorepo root:

```bash
# Build the binary
make build-agent

# Run in background daemon mode
make dev-agent
```

Direct CLI commands (inside `apps/sap-agent`):

```bash
# Test connectivity to SAP and Nembus
go run cmd/main.go test-connection

# Run one-shot master data sync (SAP -> Nembus)
go run cmd/main.go sync-downstream

# Run one-shot transaction posting (Nembus -> SAP)
go run cmd/main.go sync-upstream

# Run continuous background daemon
go run cmd/main.go daemon
```
