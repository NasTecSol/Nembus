-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1 AND is_active = true;

-- name: GetTenantBySlugAny :one
SELECT * FROM tenants WHERE slug = $1;

-- name: ListActiveTenants :many
SELECT * FROM tenants WHERE is_active = true ORDER BY tenant_name;

-- name: ListAllTenants :many
SELECT * FROM tenants ORDER BY tenant_name;

-- name: CreateTenant :one
INSERT INTO tenants (tenant_name, slug, db_conn_str, is_active, settings)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateTenant :one
UPDATE tenants
SET 
    tenant_name = COALESCE($2, tenant_name),
    slug = COALESCE($3, slug),
    db_conn_str = COALESCE($4, db_conn_str),
    is_active = COALESCE($5, is_active),
    settings = COALESCE($6, settings),
    updated_at = NOW()
WHERE id = $1
RETURNING *;
