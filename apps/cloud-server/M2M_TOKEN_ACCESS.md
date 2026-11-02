# M2M token issuance

The shared key is defined by `M2MAPIKey` in
`packages/core/middleware/m2m_key.go`. Rebuild and restart the backend after
changing it. `JWT_SECRET` must still be configured to sign tokens.

```javascript
const response = await fetch(`${baseUrl}/api/m2m/tokens`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'x-m2m-key': m2mApiKey,
    'x-tenant-id': 'stg',
  },
  body: JSON.stringify({
    client_id: 'pos-app-stg',
    client_name: 'Staging POS',
    scopes: ['products:read', 'orders:write'],
    years: 5,
  }),
});
if (!response.ok) throw new Error(`M2M issuance failed: ${response.status}`);
const client = await response.json();
// Use client.token as Authorization: Bearer <token> on subsequent API calls.
```

POST validates the shared key, generates a signed M2M JWT, saves the client to
`config/m2m_clients.json`, and returns the client object including `token`.
No user JWT or manual creation step is needed. `client_id` and `client_name`
are required; `scopes` is optional and `years` defaults to 5. Posting an existing
client ID updates its saved record. Missing or incorrect keys return HTTP 401;
missing tenant or invalid payload returns HTTP 400. Responses disable caching.

GET `/api/m2m/tokens` lists existing tokens for the supplied tenant. Send
`x-tenant-id` and `x-m2m-key` with no JWT and no request body. It returns an
array of saved clients including their tokens, or `[]` when none match.

A key embedded in frontend code is visible to users and allows token issuance
for any supplied tenant ID. The key is stored in source code; use HTTPS and distribute it only to trusted clients.
