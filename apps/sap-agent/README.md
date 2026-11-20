# Nembus SAP Sync Agent (`apps/sap-agent`)

The **SAP Sync Agent** is a bidirectional synchronization service between **SAP Business One (SAP B1)** and the **Nembus Ecosystem** (Cloud ERP & POS Desktop Client). It supersedes the legacy `SAP_To_Vtec` .NET application.

---

## Architecture & Entry Points

1. **Agent Service (`cmd/agent/main.go`)**:
   - Primary production daemon with embedded Web UI (Gin + HTML/JS/CSS).
   - Runs interactively (opens browser dashboard on port `17890`) or as a native **Windows Service** (`-service install|start|stop|uninstall`).
   - SQLite checkpoint store for watermarks, MSSQL extractor client, cloud governance transport, and auto-updater.

2. **CLI Runner (`cmd/main.go`)**:
   - Command-line utility for one-off master data syncs, transaction posting, connectivity diagnostics, and headless pipeline execution.

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
   - **Outbox Watcher**: Monitors completed POS transactions and sales returns.
   - **AR Invoices (`OINV`)**: Automatically creates official SAP AR Invoices with line items, VAT calculations, and POS transaction references.
   - **AR Credit Memos (`ORIN`)**: Posts sales returns back to SAP B1 with links to original invoices (`BaseEntry`, `BaseType = 13`).
   - **Incoming Payments (`ORCT`)**: Posts cash and card payments linked directly to the created SAP AR Invoice.
   - **Idempotency**: Tags Nembus transactions and returns with SAP `DocEntry` and `DocNum` to prevent duplicate postings.

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

## Building the Agent

### 1. PowerShell Build Script (Recommended on Windows)

From the monorepo root:

```powershell
# Build both the Agent Service and CLI Runner (outputs to build/bin/ and apps/sap-agent/bin/)
powershell -ExecutionPolicy Bypass -File scripts/build_sap_agent.ps1

# Build only the primary Agent Service
powershell -ExecutionPolicy Bypass -File scripts/build_sap_agent.ps1 -Target agent

# Build with custom version and Cloud URL
powershell -ExecutionPolicy Bypass -File scripts/build_sap_agent.ps1 -Version "v1.1.0" -CloudURL "https://nembus.nashrms.com"
```

### 2. Make Targets

From the monorepo root:

```bash
# Build primary Agent Service
make build-agent

# Build CLI Runner
make build-agent-cli

# Build both binaries
make build-agent-all
```

Or from inside `apps/sap-agent`:

```bash
make build       # Builds bin/sap-agent.exe
make build-cli   # Builds bin/sap-agent-cli.exe
make build-all   # Builds both
```

---

## Running the Agent

### Interactive Web UI Mode (Default)

```powershell
# Launches embedded UI dashboard at http://127.0.0.1:17890
& 'build/bin/sap-agent.exe'
```

### Windows Service Management

```powershell
# Install as a Windows Service
& 'build/bin/sap-agent.exe' -service install

# Start the Windows Service
& 'build/bin/sap-agent.exe' -service start

# Check Status / Stop / Uninstall
& 'build/bin/sap-agent.exe' -service status
& 'build/bin/sap-agent.exe' -service stop
& 'build/bin/sap-agent.exe' -service uninstall
```

### Direct CLI Commands

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
