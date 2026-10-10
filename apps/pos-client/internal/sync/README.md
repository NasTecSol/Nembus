# Delta sync checkpoints

The POS tenant database stores one `sync_watermarks` row per entity and store.
Pull requests read that entity's `last_sync_at`, rather than the old shared
`local_device_config.last_zatca_sync_at`. Missing checkpoints start at Unix epoch
to recover updates that older clients may have skipped.

`last_sync_at` is the latest cloud row timestamp successfully applied locally.
It is not the wall-clock time of the last check. An empty check preserves that
cursor and records `metadata.last_success_at` instead.

Inspect progress in the **local POS tenant database**:

```sql
SELECT entity_type, store_id, last_sync_at,
       metadata->>'status' AS status,
       metadata->>'last_success_at' AS last_success_at,
       metadata->>'records_applied' AS records_applied,
       metadata->>'last_entity_id' AS last_entity_id,
       metadata->>'last_error' AS last_error
FROM sync_watermarks
WHERE store_id = 1
ORDER BY entity_type;
```

Status is `syncing` while pages commit, `synced` after the final empty page, or
`error` on failure. `records_applied` counts committed records in the current
attempt; `last_entity_id` identifies the final row of the latest committed page.
Failed pages roll back their data and checkpoint together. Other entities can
continue. Updated clients use a timestamp and row-ID cursor (`metadata.cursor_id`)
with a strict 200-record page limit, including when thousands of rows share a
timestamp. The cursor travels in gRPC metadata; the server confirms support in
its response headers. Clients refuse to checkpoint against an older server.
Legacy clients retain the previous timestamp-only behavior.
Deploy the updated cloud server before the updated POS client.

Cloning stops the local background worker until backup download and restore
finish. Replacing a worker cancels and waits for the old one instead of leaving
it running. A failed clone resumes the previous tenant worker.

The cloud's existing watermark writes describe upstream push ingestion; they
do not prove a terminal applied a pull. Pull status is recorded locally.

This remains timestamp-based replication: hard deletes are not propagated,
catalog dependencies outside the pull list may cause foreign-key failures,
and the current availability-schedule schema has no `updated_at` (reported as
an entity error). Promotion price-generation triggers still run locally and
can regenerate price row identities. Stock reconciliation business logic is
unchanged. This work does not add store filtering or a change-log protocol.
