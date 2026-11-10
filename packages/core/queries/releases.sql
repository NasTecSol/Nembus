-- name: GetLatestActiveRelease :one
SELECT * FROM app_releases 
WHERE app_name = $1 AND channel = $2 AND is_active = true 
ORDER BY created_at DESC 
LIMIT 1;

-- name: GetReleaseByID :one
SELECT * FROM app_releases 
WHERE id = $1;

-- name: CreateRelease :one
INSERT INTO app_releases (
    app_name,
    version,
    channel,
    release_notes,
    file_name,
    file_path,
    file_size_bytes,
    checksum_sha256,
    is_mandatory,
    min_supported_version,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: ListReleasesForApp :many
SELECT * FROM app_releases 
WHERE app_name = $1 AND channel = $2 
ORDER BY created_at DESC;

-- name: DeactivateOlderReleases :exec
UPDATE app_releases 
SET is_active = false, updated_at = NOW() 
WHERE app_name = $1 AND channel = $2 AND id != $3;

-- name: GetOlderReleasesToPrune :many
SELECT * FROM app_releases 
WHERE app_name = $1 AND channel = $2 AND id != $3 
ORDER BY created_at DESC;

-- name: DeleteRelease :exec
DELETE FROM app_releases 
WHERE id = $1;
