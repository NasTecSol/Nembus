-- =====================================================
-- PRODUCT ENRICHMENT SUGGESTIONS (SQLC)
-- =====================================================
-- Stage 1 review queue only. Approval does not mutate products.

-- name: CreateOrGetProductEnrichmentSuggestion :one
INSERT INTO product_enrichment_suggestions (
    organization_id,
    product_id,
    source_item_code,
    source_item_name,
    source_data_fingerprint,
    contract_version,
    structured_current,
    proposed_brand,
    proposed_category,
    proposed_description,
    unsupported_semantics,
    provider,
    model,
    model_version
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
ON CONFLICT (organization_id, product_id, source_data_fingerprint, contract_version)
DO UPDATE SET updated_at = product_enrichment_suggestions.updated_at
RETURNING *;

-- name: GetProductEnrichmentSuggestionByID :one
SELECT * FROM product_enrichment_suggestions
WHERE organization_id = $1
  AND id = $2
LIMIT 1;

-- name: ListProductEnrichmentSuggestions :many
SELECT * FROM product_enrichment_suggestions
WHERE organization_id = $1
  AND status = COALESCE(sqlc.narg(status), status)
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3;

-- name: MarkProductEnrichmentSuggestionProcessing :one
UPDATE product_enrichment_suggestions
SET status = 'processing',
    reviewer_id = NULL,
    reviewed_at = NULL,
    applied_at = NULL,
    next_attempt_at = NULL,
    last_error_code = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = $1
  AND id = $2
  AND status IN ('pending', 'retryable')
RETURNING *;

-- name: MarkProductEnrichmentSuggestionInReview :one
UPDATE product_enrichment_suggestions
SET status = 'in_review',
    proposed_brand = $3,
    proposed_category = $4,
    proposed_description = $5,
    unsupported_semantics = $6,
    provider = $7,
    model = $8,
    model_version = $9,
    reviewer_id = NULL,
    reviewed_at = NULL,
    applied_at = NULL,
    next_attempt_at = NULL,
    last_error_code = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = $1
  AND id = $2
  AND status = 'processing'
RETURNING *;

-- name: ApproveProductEnrichmentSuggestion :one
UPDATE product_enrichment_suggestions
SET status = 'approved',
    reviewer_id = $3,
    reviewed_at = CURRENT_TIMESTAMP,
    applied_at = NULL,
    next_attempt_at = NULL,
    last_error_code = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = $1
  AND id = $2
  AND status = 'in_review'
RETURNING *;

-- name: RejectProductEnrichmentSuggestion :one
UPDATE product_enrichment_suggestions
SET status = 'rejected',
    reviewer_id = $3,
    reviewed_at = CURRENT_TIMESTAMP,
    applied_at = NULL,
    next_attempt_at = NULL,
    last_error_code = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = $1
  AND id = $2
  AND status = 'in_review'
RETURNING *;

-- name: MarkProductEnrichmentSuggestionFailed :one
UPDATE product_enrichment_suggestions
SET status = 'failed',
    reviewer_id = NULL,
    reviewed_at = NULL,
    applied_at = NULL,
    next_attempt_at = NULL,
    last_error_code = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = $1
  AND id = $2
  AND status = 'processing'
RETURNING *;

-- name: MarkProductEnrichmentSuggestionRetryable :one
UPDATE product_enrichment_suggestions
SET status = 'retryable',
    reviewer_id = NULL,
    reviewed_at = NULL,
    applied_at = NULL,
    next_attempt_at = NULL,
    last_error_code = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = $1
  AND id = $2
  AND status = 'processing'
RETURNING *;

-- name: MarkProductEnrichmentSuggestionApplied :one
UPDATE product_enrichment_suggestions
SET status = 'applied',
    applied_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = $1
  AND id = $2
  AND status = 'approved'
RETURNING *;
