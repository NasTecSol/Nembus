-- Stage 2D tenant-local review support. No product or taxonomy mutation.

-- name: GetProductForEnrichmentReview :one
SELECT * FROM products
WHERE organization_id = $1
  AND id = $2
LIMIT 1;

-- name: InsertProductEnrichmentReviewAudit :exec
INSERT INTO audit_logs (
    organization_id,
    table_name,
    record_id,
    action,
    old_values,
    new_values,
    changed_fields,
    performed_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);
