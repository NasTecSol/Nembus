# Test the ZATCA checkout pipeline

**Current source:** ZATCA now reads `sales_orders_v2` and `sales_order_lines_v2`. See [the sales-order test guide](ZATCA_SALES_ORDER_TEST.md) and the new sales-order Postman collection. The transaction endpoints below apply only to historical signed transaction documents; new checkout invoices use `/api/zatca/orders/ORDER_UUID/...`.

Use the local **cloud-server** on port 8081 with `x-tenant-id` and your login bearer token. Enable `organizations.metadata.zatca.enabled` through the organization ZATCA endpoint and use the saved device environment and enable ZATCA in the organization metadata. Restart after rebuilding.

## Test in the FATOORA simulation portal

The environment is now selected per device. Onboard with `"zatca_env":"simulation"`, then use `GET /api/zatca/status?device_id=ID` to confirm its environment and gateway. `ZATCA_ENV`/`ZATCA_BASE_URL` no longer route device requests; enable the organization through `PUT /api/organizations/ID/zatca` with `{"enabled":true}`. See [device environment routing](ZATCA_DEVICE_ENVIRONMENT.md).

Obtain a fresh OTP from the **simulation** portal and onboard a simulation device through `/api/zatca/onboard`. Keep the earlier sandbox registration separate, for example by using a distinct EGS serial `SIM-32164879` for this simulation unit. Do not relabel an existing sandbox certificate's `zatca_env` in SQL. Create a fresh checkout after simulation onboarding and use its `/check` endpoint.

The CSR uses `PREZATCA-Code-Signing` for simulation. Certificates and secrets must come from that same environment. A production CSID issued by `/simulation/production/csids` is still a **simulation credential**. All requests remain under `/simulation`; a device onboarded with `zatca_env: production` uses the live `/core` gateway instead.

The configured CSR invoice type `1100` covers standard and simplified documents. Checking one simplified sale does not complete the required invoice/credit-note/debit-note compliance suite. `POST /api/zatca/onboard` now performs the entire sequence in a single HTTP call: obtain and save compliance credentials, run all six sample checks, then obtain/save the reporting CSID and its matching secret. It returns HTTP 200 only when `reporting_ready: true`. No separate completion call is needed on success. Allow up to eight sequential gateway requests (30 seconds each); use a Postman request timeout of 300000 ms or 0 (unlimited).

### Resume an incomplete saved device (recovery only)

If the main onboarding call fails after saving compliance credentials, its HTTP 502 response contains the saved `device_id`, `onboarding_stage`, `compliance_checks`, and `completion_endpoint`. Resume that device without a new OTP. For example, for device configuration ID `3`:

```http
POST http://localhost:8081/api/zatca/devices/3/complete-onboarding
Authorization: Bearer <your login token>
x-tenant-id: stgin
```

No body or new OTP is required. Use the **device configuration ID** from `/onboard`, not a POS terminal or transaction ID. Set Postman's request timeout long enough for seven sequential gateway requests (six checks and one certificate exchange; each has a 30-second timeout).

The endpoint loads that tenant's saved compliance certificate, secret, request ID, key and seller profile. It generates six isolated onboarding samples (standard/simplified invoice, credit note and debit note), with an explicit sample buyer for standard documents. Samples use the normal signing pipeline with the compliance certificate and are sent only to `/compliance/invoices`. They never enter business sales, stock, payment or document-chain tables. The current CSR advertises `1100`; all six sample types are therefore checked.

Each check requires a successful HTTP response, validation status `PASS`/`WARNING`, and no validation errors. If a sample fails, HTTP 502 includes `data.compliance_checks` with ZATCA's responses; the endpoint does not request or save reporting credentials. Fix the cause and retry the same completion endpoint. Retrying an incomplete device generates a new sample suite; these are compliance tests, not duplicate sales.

If an earlier `/check` or completion request already completed a sample type, ZATCA can return HTTP 406 with code `Submitted before`, category `Compliance-Check`, and `Compliance check already completed for <type>.`. The completion flow recognizes only this specific response for the requested type and continues, returning `already_completed: true` alongside the unmodified gateway response. It does not bypass other HTTP 406s, mixed errors, or validation failures. ZATCA must still authorize the final reporting-CSID exchange.

After all six checks pass, it calls the same environment's `/production/csids` using the saved compliance request ID and credential pair. It validates the returned certificate against the existing key, reads its real expiry, then atomically saves `production_csid`, `metadata.production_secret`, the legacy active secret, `metadata.onboarding_stage`, completion time and validation responses. Other metadata and compliance credentials are preserved. Success returns `reporting_ready: true` and `onboarding_stage: production` (still `zatca_env: simulation` when testing). No secrets or private keys are returned. Repeating a completed request reuses the saved reporting configuration without gateway calls.

Then create a **new checkout** so its XML/QR uses the reporting certificate, and call its `/submit` endpoint. A sale generated before completion retains its original compliance-certificate stamp; this endpoint does not alter existing signed sale documents. Complete the specific device referenced by the transaction; another device's reporting credentials cannot repair it.

## Configure the seller first

Organizations and stores have no dedicated national-address columns. The signer reads `metadata.zatca.seller`. Non-empty store fields override organization fields. Legal name and VAT number still come from the organization (`legal_name` / `name` and `tax_id`). Organization `code` is **not** a commercial registration number.

The following is a configuration template, not real company data. Replace **every** placeholder with the seller's actual details before running it in the tenant database:

```sql
UPDATE organizations
SET metadata = jsonb_set(
    COALESCE(metadata, '{}'::jsonb),
    '{zatca}',
    COALESCE(metadata->'zatca', '{}'::jsonb) || jsonb_build_object(
        'seller', jsonb_build_object(
            'commercial_registration_number', '<10-digit CRN>',
            'street', '<actual street>',
            'building_number', '<4 digits>',
            'district', '<actual district>',
            'city', '<actual city>',
            'postal_code', '<5 digits>',
            'country_code', 'SA'
        )
    ),
    true
), updated_at = CURRENT_TIMESTAMP
WHERE id = 1;
```

This preserves other organization metadata and other settings under `zatca`. It replaces that organization's `zatca.seller` profile. For a branch-specific address, set the same profile under `stores.metadata` for that store instead.

When details are missing, checkout returns HTTP 400 listing missing fields **before** creating the sale. It also validates the certificate/key pair and amounts before writing the sale. Local signing requires a real compliance or production certificate matching the private key, even in sandbox; simulated tokens cannot create a stamped invoice.

## Create a new checkout

Import `../postman-collection/ZATCA-Checkout-Test.postman_collection.json`. Fill in its token, tenant, product, cashier and open-session variables. The checkout request creates a new sale and captures its numeric transaction ID.

For product `INV00519`, use the product's numeric database ID. For the example net amount of 100 and VAT of 15, either of these is consistent:

| Quantity | Unit price | Line subtotal | Line VAT | Line total |
|---|---|---|---|---|
| 1 | 100 | 100 | 15 | 115 |
| 2 | 50 | 100 | 15 | 115 |

Quantity 2 at unit price 100 cannot have a subtotal of 100. Amount arithmetic uses decimal fractions, not floating-point approximations. The current POS mapper uses the standard 15% VAT category; other rates/categories need their own mapping.

On success, check `data.qr_code`, `data.zatca_uuid` and the absence of `data.zatca_error`.

The signer saves **final stamped XML** to `zatca_document_chain.signed_xml`. Its API hash is recalculated from that final XML after excluding UBL extensions, the signature reference and QR reference. It fails if stamping changes the hash. The QR is embedded in the XML and contains all nine simplified-invoice tags, including the signing certificate's public key and CA signature. XML signature and QR fields use the same invoice digest and ECDSA signature. Anonymous simplified customers no longer generate empty country and VAT identifiers.

## Check compliance, then report when ready

For a device with compliance credentials:

```http
POST /api/zatca/orders/SALES_ORDER_UUID/check
```

No body is needed. Inspect `data.validationResults.errorMessages`, `warningMessages` and `status`. This check does not report the sale, change its reporting status, or complete all sample-document onboarding checks automatically.

For a device with genuine production credentials and their matching production secret:

```http
POST /api/zatca/orders/SALES_ORDER_UUID/submit
```

Only `data.zatca_response.reportingStatus = "REPORTED"` confirms reporting. Local QR generation and HTTP 201 checkout success do not establish gateway acceptance.

Existing documents generated by the old signer retain their original XML and hashes. Create a fresh checkout for this test; do not edit signed XML, reset the ICV, or change a document-chain hash manually.

## Receipt QR printing

The signer stores the QR text in `zatca_document_chain.qr_code_base64` and transaction metadata. This is Base64-encoded TLV **text**, not PNG bytes. Receipts encode that exact saved text as the QR symbol; they do not rebuild the seller, timestamp, amounts, or signature.

`GET /api/zatca/receipt-qr?transaction_id=6` retrieves the saved payload with the usual tenant/auth headers. Order-based checkout can use `sales_order_id=<UUID>` instead. This read-only endpoint returns no credentials and does not create an invoice or report it. It prefers the document-chain QR and falls back to transaction metadata.

The POS printer service loads this saved payload for retail order receipts, restaurant receipts, bill reprints and backoffice browser reprints. Sales-return receipts pass the **credit note's** `qr_code` from the return response, never the original sale's QR. It renders a local PNG with a four-module white quiet zone for preview/browser printing and passes `receipt.qr_code_base64` to the local thermal API. The Go receipt builder emits native ESC/POS QR Model 2 commands; the transaction barcode remains separate. Session summaries do not gain invoice QR codes.

Rebuild/restart the backend and frontend, open the completed transaction in Bills, then reprint or choose receipt preview when the printer is unavailable. Scan the QR and verify its text matches the saved Base64 payload. The physical printer must support ESC/POS QR commands. Command-byte tests and frontend compilation do not replace a physical print-and-scan check.

Receipt rendering uses existing signed QR data. Cart conversion now generates the order-based ZATCA document after its lines are saved. Older unsigned orders can use the order `/generate` endpoint once their final data validates; the receipt GET lookup never creates documents.

## Verification results

Offline tests cover C14N ordering/escaping and excluded nodes, malformed XML, final XML/API hash equality, certificate-key matching, ECDSA verification, all nine QR tags, one-byte TLV lengths, UTC timestamps, amount consistency, seller configuration and failure before sale insertion. Existing checkout and credit-note workflow tests now use certificate-backed fixtures.

An optional test `TestZatcaDeveloperSandboxCompliance` uses a new ephemeral test key, the developer portal's test OTP and synthetic sample seller data. It never loads tenant credentials or calls production. Enable it explicitly with `NEMBUS_ZATCA_SANDBOX_TEST=1`. Live sandbox execution was declined during this task, and the official SDK was unavailable locally; local tests are not a claim of ZATCA or SDK acceptance.

References: [Security features standard](https://www.zatca.gov.sa/ar/E-Invoicing/SystemsDevelopers/Documents/20230519_ZATCA_Electronic_Invoice_Security_Features_Implementation_Standards_vF.pdf), [ZATCA SignedProperties guidance](https://zatca1.discourse.group/t/what-do-signed-properties-look-like-when-hashing/717), [ZATCA signing steps](https://zatca.gov.sa/en/E-Invoicing/Introduction/Guidelines/Documents/E-invoicing-Detailed-Technical-Guideline.pdf).
