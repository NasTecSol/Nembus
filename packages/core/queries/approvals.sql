-- name: CreateApprovalRequest :one
INSERT INTO approval_requests (organization_id, domain, action, entity_type, entity_id, requested_by, required_permission, summary, payload, submission_notes, metadata)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING *;

-- name: GetApprovalRequest :one
SELECT * FROM approval_requests WHERE id = $1 AND organization_id = $2;

-- name: ListApprovalRequests :many
SELECT * FROM approval_requests
WHERE organization_id = $1
  AND (sqlc.narg(domain)::text IS NULL OR domain = sqlc.narg(domain))
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status))
  AND (sqlc.narg(requested_by)::int IS NULL OR requested_by = sqlc.narg(requested_by))
ORDER BY created_at DESC LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: DecideApprovalRequest :one
UPDATE approval_requests
SET status = $3, decision_notes = $4, reviewed_by = $5, reviewed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2 AND status = 'processing'
RETURNING *;

-- name: ClaimApprovalRequest :execrows
UPDATE approval_requests SET status = 'processing', updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2 AND status = 'pending';

-- name: ReleaseApprovalRequest :exec
UPDATE approval_requests SET status = 'pending', updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2 AND status = 'processing';

-- name: CancelApprovalRequest :one
UPDATE approval_requests SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2 AND requested_by = $3 AND status = 'pending'
RETURNING *;

-- name: CreateApprovalRequestLog :one
INSERT INTO approval_request_logs (approval_request_id, user_id, action, notes, metadata)
VALUES ($1,$2,$3,$4,$5) RETURNING *;

-- name: ListApprovalRequestLogs :many
SELECT * FROM approval_request_logs WHERE approval_request_id = $1 ORDER BY created_at ASC, id ASC;

-- name: GetApprovalPolicy :one
SELECT * FROM approval_policies WHERE organization_id = $1 AND domain = $2 AND action = $3;

-- name: UpsertApprovalPolicy :one
INSERT INTO approval_policies (organization_id, domain, action, is_enabled, requires_approval, required_permission, rules)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (organization_id, domain, action) DO UPDATE SET
 is_enabled=EXCLUDED.is_enabled, requires_approval=EXCLUDED.requires_approval,
 required_permission=EXCLUDED.required_permission, rules=EXCLUDED.rules, updated_at=CURRENT_TIMESTAMP
RETURNING *;

-- name: ListApprovalPolicies :many
SELECT * FROM approval_policies WHERE organization_id = $1 ORDER BY domain, action;

-- name: DeleteApprovalPolicy :execrows
DELETE FROM approval_policies
WHERE organization_id = $1 AND domain = $2 AND action = $3;

-- name: DeleteApprovalRequest :execrows
DELETE FROM approval_requests
WHERE organization_id = $1 AND id = $2;

-- name: FindPendingApprovalRequestByEntity :one
SELECT * FROM approval_requests
WHERE organization_id = $1
  AND domain = $2
  AND action = $3
  AND entity_type = $4
  AND entity_id = $5
  AND status = 'pending'
ORDER BY id DESC
LIMIT 1;

-- name: UpdateApprovalRequestPayload :one
UPDATE approval_requests
SET payload = $3, summary = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2 AND status = 'pending'
RETURNING *;
