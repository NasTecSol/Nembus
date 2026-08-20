-- Stage 2C durable execution operations. Generated SQLC companions are
-- intentionally not hand-edited; run sqlc v1.30.0 to generate them.

-- name: ListDueProductEnrichmentSuggestions :many
SELECT id, organization_id, product_id, source_item_code, source_data_fingerprint,
       status, attempt_count
FROM product_enrichment_suggestions
WHERE status IN ('pending', 'retryable')
  AND (next_attempt_at IS NULL OR next_attempt_at <= sqlc.arg(now))
ORDER BY COALESCE(next_attempt_at, created_at), id
LIMIT sqlc.arg('batch_limit');

-- name: ClaimProductEnrichmentSuggestion :one
UPDATE product_enrichment_suggestions
SET status = 'processing',
    attempt_count = attempt_count + 1,
    next_attempt_at = NULL,
    last_error_code = NULL,
    reviewer_id = NULL,
    reviewed_at = NULL,
    applied_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND status IN ('pending', 'retryable')
RETURNING id, organization_id, product_id, source_item_code, source_data_fingerprint,
          status, attempt_count;

-- name: CompleteProductEnrichmentSuggestion :one
UPDATE product_enrichment_suggestions
SET proposed_brand = sqlc.arg(proposed_brand),
    proposed_category = sqlc.arg(proposed_category),
    proposed_description = sqlc.arg(proposed_description),
    unsupported_semantics = sqlc.arg(unsupported_semantics),
    provider = sqlc.arg(provider),
    model = sqlc.arg(model),
    model_version = sqlc.narg(model_version),
    status = 'in_review',
    reviewer_id = NULL,
    reviewed_at = NULL,
    applied_at = NULL,
    next_attempt_at = NULL,
    last_error_code = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND status = 'processing'
RETURNING id, organization_id, product_id, source_item_code, source_data_fingerprint,
          status, attempt_count;

-- name: RetryProductEnrichmentSuggestion :one
UPDATE product_enrichment_suggestions
SET status = 'retryable',
    next_attempt_at = sqlc.arg(next_attempt_at),
    last_error_code = sqlc.arg(last_error_code),
    reviewer_id = NULL,
    reviewed_at = NULL,
    applied_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND status = 'processing'
RETURNING id, organization_id, product_id, source_item_code, source_data_fingerprint,
          status, attempt_count;

-- name: FailProductEnrichmentSuggestion :one
UPDATE product_enrichment_suggestions
SET status = 'failed',
    next_attempt_at = NULL,
    last_error_code = sqlc.arg(last_error_code),
    reviewer_id = NULL,
    reviewed_at = NULL,
    applied_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND status = 'processing'
RETURNING id, organization_id, product_id, source_item_code, source_data_fingerprint,
          status, attempt_count;
