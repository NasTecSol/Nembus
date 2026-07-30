-- ZATCA Device Configuration Queries
-- ====================================

-- name: GetZatcaDeviceConfig :one
SELECT * FROM zatca_device_configs WHERE id = $1;

-- name: GetZatcaDeviceConfigBySerial :one
SELECT * FROM zatca_device_configs
WHERE organization_id = $1 AND device_serial = $2;

-- name: ListZatcaDeviceConfigsByOrg :many
SELECT * FROM zatca_device_configs
WHERE organization_id = $1
ORDER BY created_at DESC;

-- name: ListZatcaDeviceConfigsByStore :many
SELECT * FROM zatca_device_configs
WHERE store_id = $1
ORDER BY created_at DESC;

-- name: GetActiveCloudDevice :one
SELECT * FROM zatca_device_configs
WHERE organization_id = $1
  AND device_type = 'cloud'
  AND is_active = true
  AND is_revoked = false
ORDER BY created_at DESC
LIMIT 1;

-- name: GetActivePosDevice :one
SELECT * FROM zatca_device_configs
WHERE organization_id = $1
  AND pos_terminal_id = $2
  AND device_type = 'pos'
  AND is_active = true
  AND is_revoked = false
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateZatcaDeviceConfig :one
INSERT INTO zatca_device_configs (
    organization_id, store_id, pos_terminal_id,
    device_serial, device_type,
    csr_pem, private_key_pem,
    compliance_csid, production_csid, csid_expiry,
    zatca_env, is_active, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: UpdateZatcaDeviceCSID :exec
UPDATE zatca_device_configs
SET compliance_csid = $2,
    production_csid = $3,
    csid_expiry     = $4,
    updated_at      = NOW()
WHERE id = $1;

-- name: RevokeZatcaDevice :exec
UPDATE zatca_device_configs
SET is_revoked     = true,
    is_active      = false,
    revoked_at     = NOW(),
    revoked_reason = $2,
    updated_at     = NOW()
WHERE id = $1;

-- name: GetZatcaConfigsDelta :many
-- Delta-fetch: returns configs modified since a given timestamp for a store
SELECT * FROM zatca_device_configs
WHERE store_id = $1 AND updated_at > $2
ORDER BY updated_at ASC;

-- name: GetZatcaConfigsDeltaByOrg :many
-- Delta-fetch by org: returns configs modified since a given timestamp
SELECT * FROM zatca_device_configs
WHERE organization_id = $1 AND updated_at > $2
ORDER BY updated_at ASC;


-- ZATCA Document Chain Queries
-- ====================================

-- name: GetLatestChainEntry :one
-- Returns the most recent chain entry for a specific device (for PIH lookup)
SELECT * FROM zatca_document_chain
WHERE device_config_id = $1
ORDER BY icv DESC
LIMIT 1;

-- name: GetNextICV :one
-- Returns MAX(icv) + 1 for a device, or 1 if no entries exist
SELECT COALESCE(MAX(icv), 0) + 1 AS next_icv
FROM zatca_document_chain
WHERE device_config_id = $1;

-- name: InsertChainEntry :one
INSERT INTO zatca_document_chain (
    entity_type, entity_id,
    device_config_id, organization_id,
    zatca_uuid, icv, pih, xml_hash,
    zatca_status, qr_code_base64, signed_xml
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: UpdateChainEntryStatus :exec
UPDATE zatca_document_chain
SET zatca_status   = $2,
    zatca_response = $3,
    submitted_at   = $4,
    cleared_at     = $5
WHERE id = $1;

-- name: UpdateChainEntryQRCode :exec
UPDATE zatca_document_chain
SET qr_code_base64 = $2
WHERE id = $1;

-- name: GetChainEntryByEntity :one
SELECT * FROM zatca_document_chain
WHERE entity_type = $1 AND entity_id = $2;

-- name: ListPendingChainEntries :many
-- For reporting worker: entries awaiting submission to ZATCA
SELECT * FROM zatca_document_chain
WHERE zatca_status IN ('pending', 'failed')
ORDER BY created_at ASC
LIMIT $1;

-- name: ListChainEntriesByDevice :many
SELECT * FROM zatca_document_chain
WHERE device_config_id = $1
ORDER BY icv DESC
LIMIT $2 OFFSET $3;


-- Sync Watermark Queries
-- ====================================

-- name: GetSyncWatermark :one
SELECT * FROM sync_watermarks
WHERE entity_type = $1 AND store_id = $2;

-- name: UpsertSyncWatermark :exec
INSERT INTO sync_watermarks (entity_type, store_id, last_sync_at)
VALUES ($1, $2, $3)
ON CONFLICT (entity_type, store_id) DO UPDATE
SET last_sync_at = EXCLUDED.last_sync_at;
