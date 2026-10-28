# Device-owned ZATCA environment

`zatca_device_configs.zatca_env` now selects the gateway for an onboarded device. The process-wide `ZATCA_ENV` and `ZATCA_BASE_URL` settings do not route device requests. The organization switch is `organizations.metadata.zatca.enabled`; `ZATCA_ENABLED` is ignored. See [organization settings](ZATCA_ORGANIZATION_SETTINGS.md). No new column or migration is required.

## Single-call onboarding

```http
POST http://localhost:8081/api/zatca/onboard
Authorization: Bearer <login token>
x-tenant-id: stgin
Content-Type: application/json
```

```json
{
  "organization_id": 1,
  "store_id": 3,
  "pos_terminal_id": 1,
  "device_serial": "POS-01-SIM",
  "device_type": "pos",
  "otp": "<OTP from the matching portal>",
  "zatca_env": "simulation"
}
```

`zatca_env` is required and must be `sandbox`, `simulation`, or `production`. Unknown/missing environments return HTTP 400 before contacting ZATCA. For production use `production`, a production-portal OTP and a separate device registration. The environment is stored with the compliance credentials before the six checks run. The CSR template, all six checks, and reporting-CSID exchange use the same requested/saved environment. Normal success requires `reporting_ready: true`. Allow up to 240 seconds for the eight sequential gateway requests; set Postman's timeout to 300000 ms or 0.

| Saved environment | Gateway base URL |
|---|---|
| `sandbox` | `https://gw-fatoora.zatca.gov.sa/e-invoicing/developer-portal` |
| `simulation` | `https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation` |
| `production` | `https://gw-fatoora.zatca.gov.sa/e-invoicing/core` |

These URLs are allowlisted in code. Database values cannot supply arbitrary gateway URLs. Each call creates/selects its client without changing shared service configuration, so devices in different environments can use the same backend.

## Saved credentials and retries

Compliance checks use the selected device's compliance CSID and compliance secret. Reporting uses its reporting CSID and matching production secret. Completion/recovery and renewal read the saved environment; they do not accept an environment override. Renewal saves the rotated certificate, matching secret and real certificate expiry together.

Do not change an existing row's environment to move it between simulation and production. New onboarding stores `metadata.credential_environment` as a binding; an environment-only edit is rejected. Old registrations without this marker remain supported using their saved environment, but their credentials are still valid only in their original environment. Onboard separate credentials to switch environments.

Failure after compliance credentials are saved returns the device ID and `/api/zatca/devices/ID/complete-onboarding`. Retry that saved device without another OTP. An invoice already signed stays attached to its original `device_config_id` on check/report/retry; activating a different device cannot reroute its saved XML. Return credit notes also resolve their original invoice's device/environment.

## Choosing the signing device

If there is one active POS device for the order's terminal, use it. If there is none, use one active organization cloud device. Multiple active candidates return an explicit ambiguity error; the server does not choose simulation versus production by creation time.

For terminals with multiple registrations, set the intended device in the order metadata:

```json
{
  "zatca_device_config_id": 3
}
```

The device must belong to the same organization and, for POS devices, the same terminal, and must be active/not revoked. The frontend saves the actual environment/ID returned by onboarding. New POS carts copy that selected reporting-ready device ID into their metadata, which cart conversion carries into the sales order. An existing cart retains its saved choice; creating a fresh cart after switching devices makes the new choice explicit. Manually inserted/API-created orders can provide the same metadata field.

## Status and frontend

`GET /api/zatca/status` reports the global switch and `environment_source: device_config`. It deliberately has no global `env` or gateway URL.

```http
GET /api/zatca/status?device_id=3
```

Device-specific status includes that device's `env`, canonical `base_url`, active flag and reporting readiness. It does not expose private keys/secrets. Update integrations that previously assumed a global environment in `/status`.

Both frontend onboarding selectors now include simulation and send `zatca_env`. The frontend reads the returned environment instead of assuming the requested/global environment, only treats `reporting_ready: true` as successful onboarding, and never fabricates sandbox credentials after an HTTP error.

## Verification

Local mock-gateway tests cover simulation and production onboarding while the old global setting deliberately differs, document/device-based reporting, fixed gateway allowlisting, invalid/missing environment rejection, explicit versus ambiguous device selection, environment-binding mismatch, device-specific status and renewal with atomic secret rotation. These tests do not contact live ZATCA. Validate a fresh simulation order after rebuilding before production rollout.
