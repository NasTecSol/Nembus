-- =====================================================
-- REQUEST FOR QUOTATIONS (RFQ) QUERIES (SQLC)
-- =====================================================

-- name: ListRFQs :many
SELECT 
    r.id,
    r.organization_id,
    r.store_id,
    r.rfq_number,
    r.vendor_name,
    r.title,
    r.description,
    r.status,
    r.token,
    r.token_valid_from,
    r.token_expires_at,
    r.invited_partner_id,
    r.created_by_user_id,
    r.metadata,
    r.created_at,
    r.updated_at,
    -- 1. Active timeline/expiry status (informational only; expiry does NOT mutate actual status)
    CASE 
        WHEN r.token_valid_from > CURRENT_TIMESTAMP THEN 'pending'
        WHEN r.token_expires_at < CURRENT_TIMESTAMP THEN 'expired'
        ELSE 'active'
    END AS timeline_status,
    (r.token_expires_at < CURRENT_TIMESTAMP) AS is_token_expired,
    GREATEST(0, EXTRACT(EPOCH FROM (r.token_expires_at - CURRENT_TIMESTAMP))::bigint) AS time_remaining_seconds,
    -- 2. Whether a response has been received
    (COUNT(resp.id) > 0) AS has_response,
    COUNT(resp.id)::int AS response_count,
    -- 3. Supplier classification (registered vs unregistered)
    CASE 
        WHEN COUNT(resp.id) = 0 THEN 'none'
        WHEN bool_or(resp.is_registered_supplier = true OR resp.partner_id IS NOT NULL) THEN 'registered'
        ELSE 'unregistered'
    END AS supplier_type,
    MAX(resp.submitted_at) AS latest_response_date
FROM rfqs r
LEFT JOIN rfq_responses resp ON r.id = resp.rfq_id
WHERE r.organization_id = $1
  AND (NULLIF($2::text, '') IS NULL OR r.status = $2)
  AND (NULLIF($3::text, '') IS NULL OR (
      CASE 
          WHEN r.token_valid_from > CURRENT_TIMESTAMP THEN 'pending'
          WHEN r.token_expires_at < CURRENT_TIMESTAMP THEN 'expired'
          ELSE 'active'
      END = $3
  ))
  AND (NULLIF($4::text, '') IS NULL OR r.rfq_number ILIKE '%' || $4 || '%' OR r.title ILIKE '%' || $4 || '%')
GROUP BY r.id
HAVING (NULLIF($5::text, '') IS NULL OR (COUNT(resp.id) > 0) = ($5::text)::boolean)
   AND (NULLIF($6::text, '') IS NULL OR (
       CASE 
           WHEN COUNT(resp.id) = 0 THEN 'none'
           WHEN bool_or(resp.is_registered_supplier = true OR resp.partner_id IS NOT NULL) THEN 'registered'
           ELSE 'unregistered'
       END = $6
   ))
ORDER BY r.created_at DESC
LIMIT $7 OFFSET $8;

-- name: CountRFQs :one
WITH filtered_rfqs AS (
    SELECT 
        r.id
    FROM rfqs r
    LEFT JOIN rfq_responses resp ON r.id = resp.rfq_id
    WHERE r.organization_id = $1
      AND (NULLIF($2::text, '') IS NULL OR r.status = $2)
      AND (NULLIF($3::text, '') IS NULL OR (
          CASE 
              WHEN r.token_valid_from > CURRENT_TIMESTAMP THEN 'pending'
              WHEN r.token_expires_at < CURRENT_TIMESTAMP THEN 'expired'
              ELSE 'active'
          END = $3
      ))
      AND (NULLIF($4::text, '') IS NULL OR r.rfq_number ILIKE '%' || $4 || '%' OR r.title ILIKE '%' || $4 || '%')
    GROUP BY r.id
    HAVING (NULLIF($5::text, '') IS NULL OR (COUNT(resp.id) > 0) = ($5::text)::boolean)
       AND (NULLIF($6::text, '') IS NULL OR (
           CASE 
               WHEN COUNT(resp.id) = 0 THEN 'none'
               WHEN bool_or(resp.is_registered_supplier = true OR resp.partner_id IS NOT NULL) THEN 'registered'
               ELSE 'unregistered'
           END = $6
       ))
)
SELECT COUNT(*) FROM filtered_rfqs;

-- name: GetRFQByID :one
SELECT 
    r.id,
    r.organization_id,
    r.store_id,
    r.rfq_number,
    r.vendor_name,
    r.title,
    r.description,
    r.status,
    r.token,
    r.token_valid_from,
    r.token_expires_at,
    r.invited_partner_id,
    r.created_by_user_id,
    r.metadata,
    r.created_at,
    r.updated_at,
    bp.name AS invited_partner_name,
    bp.code AS invited_partner_code,
    CASE 
        WHEN r.token_valid_from > CURRENT_TIMESTAMP THEN 'pending'
        WHEN r.token_expires_at < CURRENT_TIMESTAMP THEN 'expired'
        ELSE 'active'
    END AS timeline_status,
    (r.token_expires_at < CURRENT_TIMESTAMP) AS is_token_expired,
    GREATEST(0, EXTRACT(EPOCH FROM (r.token_expires_at - CURRENT_TIMESTAMP))::bigint) AS time_remaining_seconds,
    (COUNT(resp.id) > 0) AS has_response,
    COUNT(resp.id)::int AS response_count,
    CASE 
        WHEN COUNT(resp.id) = 0 THEN 'none'
        WHEN bool_or(resp.is_registered_supplier = true OR resp.partner_id IS NOT NULL) THEN 'registered'
        ELSE 'unregistered'
    END AS supplier_type,
    MAX(resp.submitted_at) AS latest_response_date
FROM rfqs r
LEFT JOIN business_partners bp ON r.invited_partner_id = bp.id
LEFT JOIN rfq_responses resp ON r.id = resp.rfq_id
WHERE r.id = $1
GROUP BY r.id, bp.name, bp.code;

-- name: GetRFQByToken :one
SELECT 
    r.id,
    r.organization_id,
    r.store_id,
    r.rfq_number,
    r.vendor_name,
    r.title,
    r.description,
    r.status,
    r.token,
    r.token_valid_from,
    r.token_expires_at,
    r.invited_partner_id,
    r.metadata,
    r.created_at,
    CASE 
        WHEN r.token_valid_from > CURRENT_TIMESTAMP THEN 'pending'
        WHEN r.token_expires_at < CURRENT_TIMESTAMP THEN 'expired'
        ELSE 'active'
    END AS timeline_status,
    (r.token_expires_at < CURRENT_TIMESTAMP) AS is_token_expired,
    GREATEST(0, EXTRACT(EPOCH FROM (r.token_expires_at - CURRENT_TIMESTAMP))::bigint) AS time_remaining_seconds
FROM rfqs r
WHERE r.token = $1;

-- name: CreateRFQWithItems :one
WITH input_items AS (
    SELECT item, item_ord
    FROM jsonb_array_elements($12::jsonb) WITH ORDINALITY AS e(item, item_ord)
), invalid_items AS (
    SELECT 1
    FROM input_items e
    LEFT JOIN products p ON p.id = (e.item->>'product_id')::integer
    WHERE p.id IS NULL
       OR p.organization_id <> $1
       OR p.is_purchasable IS NOT TRUE
       OR NOT COALESCE((
           p.base_uom_id = (e.item->>'uom_id')::integer
           OR EXISTS (
               SELECT 1 FROM product_uom_conversions puc
               WHERE puc.product_id = p.id
                 AND ((puc.from_uom_id = (e.item->>'uom_id')::integer)
                   OR (puc.to_uom_id = (e.item->>'uom_id')::integer))
           )
       ), false)
       OR (NULLIF(e.item->>'product_variant_id', '') IS NOT NULL AND NOT EXISTS (
           SELECT 1 FROM product_variants pv
           WHERE pv.id = (e.item->>'product_variant_id')::integer AND pv.product_id = p.id
       ))
), created_rfq AS (
    INSERT INTO rfqs (organization_id, store_id, rfq_number, vendor_name, title, description, status, token,
                      token_valid_from, token_expires_at, invited_partner_id, created_by_user_id, metadata)
    SELECT $1, $2, $3, $6, $4, $5, 'draft', $7, $8, $9, $10, $11, '{}'::jsonb
    WHERE NOT EXISTS (SELECT 1 FROM invalid_items)
    RETURNING *
), created_items AS (
    INSERT INTO rfq_items (rfq_id, product_id, product_variant_id, item_name, item_description,
                           quantity, uom_id, target_price, line_number, notes)
    SELECT created_rfq.id,
           NULLIF(item->>'product_id', '')::integer,
           NULLIF(item->>'product_variant_id', '')::integer,
           item->>'item_name', item->>'item_description',
           (item->>'quantity')::numeric,
           NULLIF(item->>'uom_id', '')::integer,
           NULLIF(item->>'target_price', '')::numeric,
           item_ord::integer,
           item->>'notes'
    FROM created_rfq, input_items e
    RETURNING id
)
SELECT created_rfq.* FROM created_rfq;

-- name: PublishRFQ :one
UPDATE rfqs
SET status = 'published', token = $3, token_valid_from = CURRENT_TIMESTAMP,
    token_expires_at = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2 AND status = 'draft'
RETURNING *;

-- name: CloseRFQ :one
UPDATE rfqs
SET status = 'closed', updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2 AND status = 'published'
RETURNING *;

-- name: SubmitRFQResponse :one
WITH matched_supplier AS (
    SELECT bp.id
    FROM business_partners bp
    JOIN partner_contacts pc ON pc.partner_id = bp.id
    WHERE bp.organization_id = $2
      AND bp.partner_role IN ('supplier', 'vendor')
      AND bp.is_active = true
      AND lower(pc.email) = lower($4)
    ORDER BY pc.is_primary DESC, bp.id
    LIMIT 1
), new_response AS (
    INSERT INTO rfq_responses (rfq_id, organization_id, partner_id, is_registered_supplier,
        supplier_name, supplier_email, supplier_phone, response_status, total_quoted_amount,
        valid_until, notes, metadata)
    SELECT $1, $2, matched_supplier.id, (matched_supplier.id IS NOT NULL),
        $3, $4, $5, 'submitted', $6, $7, $8, '{}'::jsonb
    FROM (SELECT 1 WHERE EXISTS (SELECT 1 FROM rfqs r WHERE r.id = $1 AND r.status = 'published')) AS source
    LEFT JOIN matched_supplier ON true
    RETURNING *
), new_items AS (
    INSERT INTO rfq_response_items (rfq_response_id, rfq_item_id, uom_id, quantity, unit_price, tax_amount,
        line_total, delivery_lead_time_days, notes)
    SELECT nr.id, ri.id, ri.uom_id, (item->>'quantity')::numeric, (item->>'unit_price')::numeric,
        COALESCE(NULLIF(item->>'tax_amount', '')::numeric, 0),
        ((item->>'quantity')::numeric * (item->>'unit_price')::numeric) + COALESCE(NULLIF(item->>'tax_amount', '')::numeric, 0),
        NULLIF(item->>'delivery_lead_time_days', '')::integer, item->>'notes'
    FROM new_response nr, jsonb_array_elements($9::jsonb) AS e(item)
    JOIN rfq_items ri ON ri.id = (item->>'rfq_item_id')::uuid
        AND ri.rfq_id = $1 AND ri.product_id = (item->>'product_id')::integer
        AND ri.uom_id = (item->>'uom_id')::integer
    RETURNING id
)
SELECT new_response.* FROM new_response;

-- name: CreateRFQ :one
INSERT INTO rfqs (
    organization_id,
    store_id,
    rfq_number,
    title,
    description,
    status,
    token,
    token_valid_from,
    token_expires_at,
    invited_partner_id,
    created_by_user_id,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: UpdateRFQ :one
UPDATE rfqs
SET title = $2,
    description = $3,
    store_id = $4,
    token_expires_at = $5,
    invited_partner_id = $6,
    metadata = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateRFQStatus :one
UPDATE rfqs
SET status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteRFQ :exec
DELETE FROM rfqs
WHERE id = $1;

-- name: CreateRFQItem :one
INSERT INTO rfq_items (
    rfq_id,
    product_id,
    product_variant_id,
    item_name,
    item_description,
    quantity,
    uom_id,
    target_price,
    line_number,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: ListRFQItems :many
SELECT 
    ri.*,
    p.name AS product_name,
    p.sku AS product_sku,
    u.name AS uom_name
FROM rfq_items ri
LEFT JOIN products p ON ri.product_id = p.id
LEFT JOIN units_of_measure u ON ri.uom_id = u.id
WHERE ri.rfq_id = $1
ORDER BY ri.line_number ASC;

-- name: DeleteRFQItems :exec
DELETE FROM rfq_items
WHERE rfq_id = $1;

-- name: CreateRFQResponse :one
INSERT INTO rfq_responses (
    rfq_id,
    organization_id,
    partner_id,
    is_registered_supplier,
    supplier_name,
    supplier_email,
    supplier_phone,
    response_status,
    total_quoted_amount,
    valid_until,
    notes,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: ListRFQResponses :many
SELECT 
    resp.*,
    bp.code AS partner_code,
    bp.partner_role
FROM rfq_responses resp
LEFT JOIN business_partners bp ON resp.partner_id = bp.id
WHERE resp.rfq_id = $1
ORDER BY resp.submitted_at DESC;

-- name: GetRFQResponse :one
SELECT 
    resp.*,
    bp.code AS partner_code,
    bp.partner_role
FROM rfq_responses resp
LEFT JOIN business_partners bp ON resp.partner_id = bp.id
WHERE resp.id = $1;

-- name: CreateRFQResponseItem :one
INSERT INTO rfq_response_items (
    rfq_response_id,
    rfq_item_id,
    uom_id,
    quantity,
    unit_price,
    tax_amount,
    line_total,
    delivery_lead_time_days,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: ListRFQResponseItems :many
SELECT 
    rri.*,
    ri.item_name,
    ri.line_number,
    u.name AS uom_name,
    ri.product_id
FROM rfq_response_items rri
LEFT JOIN rfq_items ri ON rri.rfq_item_id = ri.id
LEFT JOIN units_of_measure u ON u.id = rri.uom_id
WHERE rri.rfq_response_id = $1
ORDER BY rri.created_at ASC;
