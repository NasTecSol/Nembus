# Organization-wide ZATCA switch

ZATCA enablement is read from **`organizations.metadata.zatca.enabled`** in the request's tenant database. `ZATCA_ENABLED` no longer enables/disables the feature. Device environments still come from `zatca_device_configs.zatca_env`.

Missing flags mean **disabled**. After upgrading, enable each intended organization using the endpoint or the checkbox below. Existing seller metadata and certificates are preserved; no database migration or new OTP is required for this setting change.

## Read and update

Use your normal login bearer token and `x-tenant-id` header. The routes use the same authenticated tenant/organization boundary as organization administration; when the token has an organization ID, another organization cannot be targeted.

```http
GET http://localhost:8081/api/organizations/1/zatca
Authorization: Bearer <token>
x-tenant-id: stgin
```

```http
PUT http://localhost:8081/api/organizations/1/zatca
Authorization: Bearer <token>
x-tenant-id: stgin
Content-Type: application/json
```

Enable:

```json
{"enabled": true}
```

Disable using the same PUT endpoint:

```json
{"enabled": false}
```

Successful reads and writes return:

```json
{
  "statusCode": 200,
  "data": {
    "organization_id": 1,
    "enabled": true,
    "source": "organizations.metadata.zatca.enabled"
  }
}
```

The update atomically changes only the nested boolean and `updated_at`. It preserves `metadata.zatca.seller`, other ZATCA options, and unrelated metadata. Missing/null/non-boolean request values return HTTP 400; `false` is a valid update.

## Runtime behavior

* Enabled organizations can onboard/complete/renew devices, generate signed invoices and credit notes, check compliance and submit/report them.
* Disabled organizations can continue ordinary sales without automatic ZATCA generation. Explicit ZATCA generation, onboarding, checking, reporting and renewal are blocked. Public gateway operations return HTTP 409 for the disabled organization.
* Turning off the switch never deletes credentials, signed XML, QR codes, chain entries or reporting responses. Existing saved receipts can still retrieve their QR.
* Re-enabling resumes eligibility for pending signed documents. Retries reuse the same saved invoice UUID/hash/XML. It does not automatically issue invoices for all sales created while disabled.
* The reporting worker starts independently of the old environment flag and filters enabled organizations before its batch limit. Each dispatch rechecks the organization. Already-dispatched network requests are not cancelled retroactively.

Every operation reads the setting from the organization record; toggling does not require a server restart. Rebuild/restart once to install this code. For offline POS databases, the organization record still needs the application's normal synchronization to receive changes made in the cloud tenant database.

## Frontend

The ZATCA Settings page now has **Enable ZATCA for this organization**. It reads the dedicated GET endpoint and saves changes using PUT. The checkbox is organization-wide, not a per-terminal setting. Onboarding is disabled in the form until the organization setting has loaded and is enabled. Switching off preserves saved device credentials so switching back on does not require onboarding again.

Status can be read with either:

```http
GET /api/zatca/status?organization_id=1
GET /api/zatca/status?device_id=3
```

Both include `enabled_source: organization_metadata`. The device variant also includes its saved environment and gateway. Without an organization ID, device ID, or organization in the authenticated context, status returns `enabled: null` and `organization_required: true` instead of inventing a global setting.

Import `../postman-collection/ZATCA-Organization-Settings.postman_collection.json` for the read/enable/disable calls. No real gateway request is needed to test the switch itself.
