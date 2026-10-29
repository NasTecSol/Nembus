-- =====================================================
-- USER BIOMETRIC TEMPLATES (ISO/IEC 19794-2:2005 FMD)
-- =====================================================

-- name: UpsertUserBiometricTemplate :one
INSERT INTO user_biometric_templates (
    organization_id,
    user_id,
    finger_index,
    fmd_format,
    fmd_base64,
    quality_score,
    reader_serial,
    enrolled_by_user_id,
    is_active,
    metadata,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, true, COALESCE($9, '{}'::jsonb), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
ON CONFLICT (user_id, finger_index) DO UPDATE SET
    organization_id = EXCLUDED.organization_id,
    fmd_format = EXCLUDED.fmd_format,
    fmd_base64 = EXCLUDED.fmd_base64,
    quality_score = EXCLUDED.quality_score,
    reader_serial = EXCLUDED.reader_serial,
    enrolled_by_user_id = EXCLUDED.enrolled_by_user_id,
    is_active = true,
    metadata = EXCLUDED.metadata,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: ListUserBiometricTemplates :many
SELECT
    id,
    organization_id,
    user_id,
    finger_index,
    fmd_format,
    quality_score,
    reader_serial,
    enrolled_by_user_id,
    is_active,
    metadata,
    created_at,
    updated_at
FROM user_biometric_templates
WHERE user_id = $1 AND is_active = true
ORDER BY finger_index ASC;

-- name: GetActiveUserBiometricTemplatesWithFMD :many
SELECT *
FROM user_biometric_templates
WHERE user_id = $1 AND is_active = true
ORDER BY finger_index ASC;

-- name: ListActiveBiometricTemplatesByOrg :many
SELECT
    ubt.*,
    u.username,
    u.email,
    u.first_name,
    u.last_name,
    u.is_active AS user_is_active
FROM user_biometric_templates ubt
JOIN users u ON u.id = ubt.user_id
WHERE ubt.organization_id = $1
  AND ubt.is_active = true
  AND u.is_active = true
ORDER BY ubt.id ASC;

-- name: ListAllActiveBiometricTemplates :many
SELECT
    ubt.*,
    u.username,
    u.email,
    u.first_name,
    u.last_name,
    u.is_active AS user_is_active
FROM user_biometric_templates ubt
JOIN users u ON u.id = ubt.user_id
WHERE ubt.is_active = true
  AND u.is_active = true
ORDER BY ubt.id ASC;

-- name: DeactivateUserBiometricTemplate :one
UPDATE user_biometric_templates
SET
    is_active = false,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND finger_index = $2 AND is_active = true
RETURNING *;

-- name: SetCashierSessionOpenedSupervisor :exec
UPDATE cashier_sessions
SET
    opened_by_supervisor_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: SetCashierSessionClosedSupervisor :exec
UPDATE cashier_sessions
SET
    closed_by_supervisor_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;
