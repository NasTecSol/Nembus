# ZATCA Phase 2 E-Invoicing & Sync Core Guide

This guide provides a comprehensive overview of the **ZATCA Phase 2 (FATOORA)** E-Invoicing integration, the **Transactional Outbox Pattern (Push)**, and the **Delta-Fetch Mechanism (Pull)** implemented across the NEMBUS monorepo.

---

## 1. Architecture Overview

```
                      +---------------------------------------+
                      |          ZATCA FATOORA Portal         |
                      +---------------------------------------+
                                  ^               ^
             B2B Clearance (Sync) |               | B2C Reporting (Async Worker)
                                  v               |
                      +---------------------------------------+
                      |             Cloud ERP Server          |
                      |          (apps/cloud-server)          |
                      +---------------------------------------+
                                  ^               ^
                Push Outbox Sync  |               | Pull Delta Configs
              (POST /zatca/sync/push)             | (GET /zatca/configs)
                                  |               |
                      +---------------------------------------+
                      |            Wails POS Client           |
                      |           (apps/pos-client)           |
                      +---------------------------------------+
```

- **Shared Core Foundation (`packages/core`)**: Contains cryptographic algorithms, UBL 2.1 XML builder, canonicalization (C14N11), TLV QR generator, base migrations, and SQLC queries.
- **Cloud ERP Server (`apps/cloud-server`)**: Handles real-time B2B clearance with ZATCA APIs, device onboarding, B2C background reporting worker, and sync ingestion.
- **Wails POS Client (`apps/pos-client`)**: Operates 100% offline, signs B2C tax invoices locally using embedded PCSID, generates TLV QR code for receipt printing, enqueues records into `sync_queue` outbox, and pulls certificate updates/revocations via delta-fetch.

---

## 2. Environment Configuration & Feature Toggles

ZATCA flow behavior is completely controlled via environment variables in `.env` (`apps/cloud-server/.env` and `apps/pos-client/.env`).

```env
# ── ZATCA Configuration ────────────────────────────
# Set to false for deployments outside Saudi Arabia (completely bypasses ZATCA logic)
ZATCA_ENABLED=true

# Environment: "sandbox" (local/dev) or "production" (live ZATCA portal)
ZATCA_ENV=sandbox

# Organization VAT Registration Number (15 digits)
ZATCA_ORG_VAT_ID=300000000000003
```

### Sandbox vs Production Environment Switching

| Setting | Base URL | CSR Template Name | OTP Source |
|---------|----------|-------------------|------------|
| **`ZATCA_ENV=sandbox`** | `https://gw-fatoora.zatca.gov.sa/e-invoicing/developer-portal` | `PREZATCA-Code-Signing` | Developer Portal Simulation |
| **`ZATCA_ENV=production`** | `https://gw-fatoora.zatca.gov.sa/e-invoicing/core` | `ZATCA-Code-Signing` | FATOORA Live Portal (1-hour OTP) |

---

## 3. Database Schema (`000001_base_schema.sql`)

### Dedicated ZATCA Tables

1. **`zatca_device_configs`**
   - Stores EGS unit identities, private keys (ECDSA secp256k1), Compliance CSID (CCSID), Production CSID (PCSID), certificate expiry dates, and revocation flags for both Cloud (B2B) and POS terminals (B2C).

2. **`zatca_document_chain`**
   - Sequential cryptographic ledger tracking Invoice Counter Value (ICV), Previous Invoice Hash (PIH), XML SHA-256 Hash, Base64 TLV QR code, and ZATCA response payload per device.
   - *Note:* The first document in a new device chain uses the exact ZATCA-mandated Genesis Base64 PIH:
     `NWZlY2ViNjZmZmM4NmYzOGQ5NTI3ODZjNmQ2OTZjNzljMmRiYzIzOWRkNGU5MWI0NjcyOWQ3M2EyN2ZiNTdlOQ==`

3. **`sync_queue`** (POS extension)
   - Priority outbox queue for storing offline transactions and operational actions for push synchronization.

4. **`sync_watermarks`**
   - Tracks store-level synchronization timestamps for delta-fetch polling.

---

## 4. Cryptographic Engine & TLV QR Code

All cryptographic functions reside in `packages/core/usecase`:

- **Key Generation & CSR (`zatca_crypto.go`)**:
  - ECDSA `secp256k1` keypair generation.
  - PKCS#10 CSR builder with ZATCA OID extensions (Certificate Template Name, SAN directoryName encoding InvoiceType `1100`, Location, Industry).

- **UBL 2.1 XML Builder (`zatca_xml.go`)**:
  - Builds standard UBL 2.1 XML documents (`388` Standard, `381` Credit Note, `383` Debit Note).
  - Subtype tags: `0100000` (B2B Clearance) or `0200000` (B2C Reporting).
  - Performs canonicalization (C14N 1.1) prior to hashing and signing.

- **TLV QR Code Generator (`GenerateTLVQRCode`)**:
  - Encodes the 9 mandatory ZATCA tags into UTF-8 Base64:
    1. Seller Name
    2. Seller VAT Registration Number
    3. Timestamp (ISO 8601)
    4. Invoice Total (with VAT)
    5. VAT Total
    6. XML SHA-256 Hash
    7. ECDSA Signature
    8. ECDSA Public Key
    9. ZATCA Cryptographic Stamp (if cleared)

---

## 5. Device Onboarding Flow

```
1. Generate Keypair -> 2. CSR Generation -> 3. POST /compliance (OTP) -> 4. 6 Test Invoices -> 5. POST /production/csids
```

1. Admin inputs a 1-hour OTP from the FATOORA Portal.
2. The Cloud server executes CSR generation and calls `POST /compliance` to obtain a Compliance CSID (CCSID).
3. 6 test documents are submitted to `POST /compliance/invoices`.
4. `POST /production/csids` exchanges the CCSID for a Production CSID (PCSID) and saves it to `zatca_device_configs`.

---

## 6. Real-Time B2B Clearance vs Offline B2C Reporting

### A. B2B Standard Invoices (Real-Time Clearance)
1. User issues a B2B invoice in Cloud ERP.
2. `ClearB2BInvoice` fetches latest PIH from `zatca_document_chain`.
3. Canonicalizes XML, signs hash with cloud PCSID.
4. Synchronously submits to ZATCA Clearance API (`POST /invoices/clearance/single`).
5. On success, extracts official ZATCA stamp and attaches QR code to `invoices.metadata`.

### B. B2C Simplified Tax Invoices (Offline Signing + 24h Reporting)
1. POS terminal executes sale offline.
2. POS queries local `zatca_document_chain` for PIH and signs transaction locally.
3. Instantly generates Base64 TLV QR code for receipt printing.
4. In the *same atomic transaction*, inserts transaction into `pos_transactions` and payload into `sync_queue`.
5. Cloud server background reporting worker (`StartReportingWorker`) routinely polls pending reports and submits them to ZATCA Reporting API (`POST /invoices/reporting/single`) within 24 hours.
6. *Chain Integrity:* Rejected invoices do **not** break the hash chain; subsequent invoices continue from the previous hash.

---

## 7. Sync Engine (Push & Pull)

### Push Mechanism (POS → Cloud)
The Wails background worker (`SyncService`) drains `sync_queue` every 2 minutes:
- Endpoint: `POST /api/zatca/sync/push`
- Sends batched JSON outbox items (transactions, cashier sessions, orders).
- Updates local `sync_queue` item status to `synced` upon acknowledgment.

### Pull / Delta-Fetch Mechanism (Cloud → POS)
The POS routinely polls for configuration deltas:
- Endpoint: `GET /api/zatca/configs?store_id=X&since=<timestamp>`
- Returns device config records updated after `since`.
- Automatically downloads **renewed PCSID certificates** or **revocation flags** (instantly halting compromised terminals offline).

---

## 8. API Reference & Swagger

API endpoints are documented in Swagger UI (`/swagger/index.html`) and included in `postman-collection/Nembus-API.postman_collection.json`:

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/zatca/status` | Get ZATCA feature toggle, env, base URL, and VAT ID |
| `GET` | `/api/zatca/configs` | Delta-fetch device configs (Pull Sync) |
| `POST` | `/api/zatca/sync/push` | Ingest offline POS outbox transactions (Push Sync) |
