# M2M token issuance

Set `M2M_API_KEY` in the server environment or ignored `.env.local` in the
cloud-server working directory. The existing local key is unchanged. Restart
when changing configuration. `JWT_SECRET` must be configured to sign tokens.

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
missing tenant or invalid payload returns HTTP 400. An unconfigured shared key
returns HTTP 503. Responses disable caching.

GET `/api/m2m/tokens` lists existing tokens for the supplied tenant and requires
JWT authentication again. It does not accept the shared key as authentication.

A key embedded in frontend code is visible to users and allows token issuance
for any supplied tenant ID. Keep it out of source control and published docs,
use HTTPS, and distribute it only to trusted clients.
