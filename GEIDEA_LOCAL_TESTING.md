# GEIDEA Cloud API: local test guide

## What was wired

The cloud server now mounts the GEIDEA payment endpoints under its authenticated tenant API. The handler takes the tenant-scoped repository from the request context, so payment and terminal records are read/written in the selected tenant database rather than in the master database. The versioned endpoint prefix is `/api/v1/...` when mounted by the cloud server.

The terminal driver uses the terminal record's metadata. Set `is_test_mode` to `true` to select the in-process simulator; no card terminal or real charge is involved. A normal simulated amount is approved, while an amount ending in `.99` is declined. The ping endpoint also uses a simulator and is not a physical-terminal connectivity check.

## Prerequisites

- Local PostgreSQL and the NEMBUS master/tenant databases are running and migrated.
- The cloud server development configuration is set up as described in [apps/cloud-server/ENVIRONMENTS.md](apps/cloud-server/ENVIRONMENTS.md).
- You know an active tenant slug, a valid store ID in that tenant, and an existing `pos_transactions.id` in that tenant. The GEIDEA purchase endpoint records payment against an existing transaction; it does not create the order/transaction.
- Run all commands from the repository root unless otherwise noted.

## 1. Start the cloud server

In PowerShell:

```powershell
$env:ENV = "development"
Push-Location .\apps\cloud-server
go run . dev
```

Leave that terminal running. The configured master DB must be reachable. The dev-token endpoint only works when the process environment has `ENV=development` or `ENV=dev`.

## 2. Get a local development token

In a second PowerShell terminal:

```powershell
$tokenResponse = Invoke-RestMethod -Method Get -Uri "http://localhost:8080/api/dev/token"
$token = $tokenResponse.token
$tenant = "YOUR_TENANT_SLUG"
$headers = @{
  Authorization = "Bearer $token"
  "x-tenant-id" = $tenant
}
```

Use a tenant slug that exists and is active in the master database. The dev JWT is for local development only.

## 3. Create a simulator terminal

Use an existing store ID in that tenant. This creates a terminal record configured for the simulator:

```powershell
$terminalBody = @{
  store_id = 1
  terminal_code = "GEIDEA-SIM-LOCAL"
  terminal_name = "GEIDEA Simulator"
  metadata = @{ is_test_mode = $true }
} | ConvertTo-Json -Depth 5

$terminalResult = Invoke-RestMethod -Method Post `
  -Uri "http://localhost:8080/api/pos/terminals" `
  -Headers $headers `
  -ContentType "application/json" `
  -Body $terminalBody
$terminalResult
```

Replace `store_id` with an actual store ID. If the terminal code already exists, retrieve it with the POS terminal API instead of creating a duplicate; confirm its metadata contains `is_test_mode: true`.

## 4. Check the registered API responds

```powershell
Invoke-RestMethod -Method Post `
  -Uri "http://localhost:8080/api/pos/payments/geidea/ping" `
  -Headers $headers
```

Expected: `status: online`. This checks the simulator ping handler and route/middleware path; it does **not** prove a physical terminal is online.

## 5. Simulate approval and decline

Choose an existing transaction ID in the same tenant database. Use a new ECR reference for every call because the database column is unique. Example approval:

```powershell
$purchaseBody = @{
  transaction_id = 123
  terminal_id = $terminalResult.data.id
  amount = 25.00
  ecr_reference_id = "LOCAL-GEIDEA-APPROVE-001"
} | ConvertTo-Json

$approved = Invoke-RestMethod -Method Post `
  -Uri "http://localhost:8080/api/pos/payments/geidea/purchase" `
  -Headers $headers `
  -ContentType "application/json" `
  -Body $purchaseBody
$approved | ConvertTo-Json -Depth 10
```

**The API call to hit is `POST http://localhost:8080/api/pos/payments/geidea/purchase`.** The PowerShell command above sends that request. Replace `transaction_id` with an existing ID, and use the terminal ID returned in step 3. The request body sent to the API is:

```json
{
  "transaction_id": 123,
  "terminal_id": 1,
  "amount": 25.00,
  "ecr_reference_id": "LOCAL-GEIDEA-APPROVE-001"
}
```

Expected: HTTP 200, `payment.status` is `approved`, and `terminal_response` includes simulated RRN/auth/TID/MID/card details. Copy the exact `ecr_reference_id` you sent (`LOCAL-GEIDEA-APPROVE-001` in this example), then run the SQL check below in pgAdmin connected to the same tenant database:

```sql
SELECT id, transaction_id, payment_gateway, amount, status,
       ecr_reference_id, rrn, auth_code, card_scheme, masked_pan, raw_response
FROM public.pos_payments
WHERE ecr_reference_id = 'LOCAL-GEIDEA-APPROVE-001';
```

You should see one saved row with `payment_gateway = 'geidea'` and `status = 'approved'`. For the decline test, hit the **same POST endpoint** again with `amount` set to `25.99` and a new reference such as `LOCAL-GEIDEA-DECLINE-001`; query the row using that new reference and expect `status = 'declined'`.

Replace `transaction_id` and `terminal_id` with real IDs. The response shape from terminal creation may wrap the terminal under `data`; inspect `$terminalResult` and set `$terminalId` from the returned record if the wrapper differs:

```powershell
$terminalId = $terminalResult.data.id
```

## Versioned path

The same endpoints are also mounted at `/api/v1/pos/payments/geidea/...`, for example:

`POST http://localhost:8080/api/v1/pos/payments/geidea/purchase`

## Automated checks

From the repository root:

```powershell
go -C .\packages\core test ./routing ./handler ./usecase ./payment/geidea
go -C .\apps\cloud-server test ./...
go -C .\apps\pos-client test ./internal/payment/geidea
```

The routing test verifies both path sets are registered and the simulator ping responds. The focused POS-client tests use a mock terminal for the separate local ECR protocol. These automated tests do not require real hardware.

## Scope and remaining boundary

This enables the cloud GEIDEA purchase API and simulator-backed persistence test. It does not connect the Angular checkout UI to that API, and the simulator path does not validate a physical GEIDEA terminal. Physical terminal testing still requires the authorized ECR device and its supported protocol/configuration. The POS desktop's ECR protocol and the backend driver's JSON TCP protocol remain separate implementations; do not claim they have been bridged by this change.
