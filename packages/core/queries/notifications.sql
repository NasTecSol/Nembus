-- =====================================================
-- NOTIFICATION ENGINE & OUTBOX QUERIES
-- =====================================================

-- -----------------------------------------------------
-- 1. Transactional Outbox (Events)
-- -----------------------------------------------------

-- name: EnqueueNotificationEvent :one
INSERT INTO notification_events (
    organization_id,
    event_type_id,
    entity_type,
    entity_id,
    idempotency_key,
    payload,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, 'pending'
)
ON CONFLICT (idempotency_key) DO UPDATE
    SET status = notification_events.status
RETURNING *;

-- name: GetPendingNotificationEventsForUpdate :many
SELECT * FROM notification_events
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: UpdateNotificationEventStatus :one
UPDATE notification_events
SET status = $2,
    retry_count = retry_count + 1,
    dispatched_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: MarkNotificationEventFailed :one
UPDATE notification_events
SET status = 'failed',
    retry_count = retry_count + 1
WHERE id = $1
RETURNING *;

-- name: GetNotificationEventByID :one
SELECT * FROM notification_events
WHERE id = $1;

-- -----------------------------------------------------
-- 2. Delivery Audit Log
-- -----------------------------------------------------

-- name: CreateNotificationDelivery :one
INSERT INTO notification_deliveries (
    event_id,
    channel_type,
    recipient,
    status,
    provider_message_id,
    error_message,
    sent_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: ListNotificationDeliveriesByEvent :many
SELECT * FROM notification_deliveries
WHERE event_id = $1
ORDER BY created_at ASC;

-- -----------------------------------------------------
-- 3. In-App User Notifications (Inbox & Bell)
-- -----------------------------------------------------

-- name: CreateUserNotification :one
INSERT INTO user_notifications (
    organization_id,
    user_id,
    event_id,
    title,
    message,
    action_url
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: ListUserNotifications :many
SELECT * FROM user_notifications
WHERE organization_id = $1
  AND user_id = $2
  AND (sqlc.narg('is_read')::boolean IS NULL OR is_read = sqlc.narg('is_read'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: GetUnreadNotificationCount :one
SELECT COUNT(*) FROM user_notifications
WHERE organization_id = $1
  AND user_id = $2
  AND is_read = false;

-- name: MarkUserNotificationAsRead :one
UPDATE user_notifications
SET is_read = true,
    read_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND user_id = $2
RETURNING *;

-- name: MarkAllUserNotificationsAsRead :exec
UPDATE user_notifications
SET is_read = true,
    read_at = CURRENT_TIMESTAMP
WHERE organization_id = $1
  AND user_id = $2
  AND is_read = false;

-- -----------------------------------------------------
-- 4. Channel Configurations (AWS SES, Meta WhatsApp)
-- -----------------------------------------------------

-- name: GetNotificationChannel :one
SELECT * FROM notification_channels
WHERE organization_id = $1
  AND channel_type = $2;

-- name: ListNotificationChannels :many
SELECT * FROM notification_channels
WHERE organization_id = $1
ORDER BY channel_type;

-- name: UpsertNotificationChannel :one
INSERT INTO notification_channels (
    organization_id,
    channel_type,
    is_active,
    config
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (organization_id, channel_type) DO UPDATE SET
    is_active = EXCLUDED.is_active,
    config = EXCLUDED.config,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- -----------------------------------------------------
-- 5. Notification Templates
-- -----------------------------------------------------

-- name: GetNotificationTemplate :one
SELECT * FROM notification_templates
WHERE organization_id = $1
  AND event_type_id = $2
  AND channel_type = $3
  AND locale = $4;

-- name: ListNotificationTemplates :many
SELECT * FROM notification_templates
WHERE organization_id = $1
  AND (sqlc.narg('event_type_id')::text IS NULL OR event_type_id = sqlc.narg('event_type_id'))
ORDER BY event_type_id, channel_type, locale;

-- name: UpsertNotificationTemplate :one
INSERT INTO notification_templates (
    organization_id,
    event_type_id,
    channel_type,
    locale,
    subject_template,
    body_template,
    meta_template_name
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (organization_id, event_type_id, channel_type, locale) DO UPDATE SET
    subject_template = EXCLUDED.subject_template,
    body_template = EXCLUDED.body_template,
    meta_template_name = EXCLUDED.meta_template_name,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: DeleteNotificationTemplate :execrows
DELETE FROM notification_templates
WHERE organization_id = $1
  AND id = $2;

-- -----------------------------------------------------
-- 6. Routing Rules
-- -----------------------------------------------------

-- name: ListNotificationRulesByEvent :many
SELECT * FROM notification_rules
WHERE organization_id = $1
  AND event_type_id = $2
  AND is_active = true;

-- name: ListAllNotificationRules :many
SELECT * FROM notification_rules
WHERE organization_id = $1
ORDER BY event_type_id, id;

-- name: CreateNotificationRule :one
INSERT INTO notification_rules (
    organization_id,
    event_type_id,
    audience_type,
    target_role_id,
    channels,
    conditions,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: UpdateNotificationRule :one
UPDATE notification_rules
SET audience_type = $3,
    target_role_id = $4,
    channels = $5,
    conditions = $6,
    is_active = $7
WHERE id = $1
  AND organization_id = $2
RETURNING *;

-- name: DeleteNotificationRule :execrows
DELETE FROM notification_rules
WHERE id = $1
  AND organization_id = $2;

-- -----------------------------------------------------
-- 7. Internal Recipient Lookup
-- -----------------------------------------------------

-- name: GetUsersByRole :many
SELECT u.id, u.email, u.first_name, u.last_name, u.is_active
FROM users u
JOIN user_roles ur ON ur.user_id = u.id
WHERE u.organization_id = $1
  AND ur.role_id = $2
  AND u.is_active = true;
