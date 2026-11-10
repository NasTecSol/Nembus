# POS gRPC sync (hardened)

Cloud ↔ POS synchronization uses gRPC **StreamPush** (outbox), **StreamChanges** (monotonic `change_log` pull), and **StreamSession** (PostgreSQL `LISTEN` wakeups). Legacy **StreamPull** is used only when the cloud binary does not implement `StreamChanges`.

## Deploy order

1. Apply migration `20261108140000_pos_sync_hardening.sql` on cloud tenant databases.
2. Deploy updated **cloud-server** (gRPC + change_log triggers).
3. Deploy updated **pos-client** (change_log consumer + optional session listener).

## Pull: change_log watermark

The POS stores progress in `sync_watermarks` under entity type `__change_log__` with `metadata.last_seq` (not per-entity timestamps for catalog pull).

```sql
SELECT entity_type, store_id, metadata->>'last_seq' AS last_seq, metadata->>'status' AS status
FROM sync_watermarks
WHERE entity_type = '__change_log__';
```

If the cloud returns `Unimplemented` for `StreamChanges`, the client falls back to timestamp-based `StreamPull` and per-entity watermarks described in historical notes below.

## Push: outbox resilience

`sync_queue` supports `error_class`, `next_attempt_at`, `quarantined_at`, and `outbox_uuid`. Permanent failures (validation / SQLSTATE 22xxx, 23xxx) are **quarantined** until an operator retries from the POS header. Transient errors use exponential backoff.

Inspect quarantine:

```sql
SELECT id, entity_type, entity_id, status, error_class, last_error, quarantined_at
FROM sync_queue
WHERE status IN ('quarantined', 'failed')
ORDER BY quarantined_at DESC NULLS LAST, id DESC;
```

## Terminal sync token (optional)

When cloud runs with `NEMBUS_REQUIRE_SYNC_TOKEN=true`, each stream must send metadata:

- `x-tenant-id` (tenant slug)
- `x-nembus-sync-store-id`
- `x-nembus-sync-device-id`
- `x-nembus-sync-token` (64-char hex raw token)

Issue or rotate a token:

```http
POST /api/pos/terminals/{id}/issue-sync-token
Authorization: Bearer <m2m-or-user-jwt>
x-tenant-id: <tenant-slug>
```

The response includes `sync_token` once; only SHA-256 is stored in `pos_terminals.sync_token_hash`. POS setup calls this during device registration and persists the token in `local_device_config` / `device_config.json`.

## Session stream (Phase 4)

The POS opens `StreamSession` primarily to receive `ChangesAvailableNotification` and trigger `SyncNow()`. Outbox drain still uses `StreamPush`. A 15s timer remains as fallback if the session RPC is unavailable.

## Legacy timestamp pull (fallback only)

When change_log is unavailable, each entity still has its own `sync_watermarks.last_sync_at` row. See prior checkpoint semantics: 200-row pages, cursor in metadata, max five pages per entity per cycle.

Known limitations (unchanged): hard deletes may not propagate on timestamp fallback; catalog FK ordering; some entities lack `updated_at`.
