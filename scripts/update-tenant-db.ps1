<#
.SYNOPSIS
    Updates or inserts a tenant in the Master Database to route requests to Supabase (or any DB URL).

.PARAMETER MasterDbUrl
    Connection string to the Master database.
    Defaults to local master DB from cloud-server/.env.dev.

.PARAMETER TenantSlug
    The slug of the tenant to update or register (e.g., "e2e-enrichment", "qitaf", "default").

.PARAMETER TenantName
    Display name for the tenant (only used if creating a new tenant).

.PARAMETER DbUrl
    Target database connection string (e.g. Supabase connection string).

.EXAMPLE
    .\scripts\update-tenant-db.ps1 -TenantSlug "e2e-enrichment" -DbUrl "postgresql://postgres.gktqsuecolwccvhpmuqw:Nembus%402026@aws-0-ap-northeast-2.pooler.supabase.com:5432/postgres?sslmode=require"
#>
param(
    [string]$MasterDbUrl = "postgres://postgres:root1234@localhost:5432/nembus_e2e_master?sslmode=disable",
    [Parameter(Mandatory = $true)]
    [string]$TenantSlug,
    [string]$TenantName = "Supabase Tenant",
    [Parameter(Mandatory = $true)]
    [string]$DbUrl
)

$ErrorActionPreference = "Stop"

# Auto-detect psql
if (-not (Get-Command psql -ErrorAction SilentlyContinue)) {
    $commonPgPaths = @(
        "C:\Program Files\PostgreSQL\18\bin",
        "C:\Program Files\PostgreSQL\17\bin",
        "C:\Program Files\PostgreSQL\16\bin",
        "C:\Program Files\PostgreSQL\15\bin"
    )
    foreach ($p in $commonPgPaths) {
        if (Test-Path (Join-Path $p "psql.exe")) {
            $env:PATH = "$p;$env:PATH"
            break
        }
    }
}

Write-Host "Updating tenant '$TenantSlug' in Master DB..." -ForegroundColor Cyan

# SQL script to upsert tenant
$sql = @"
INSERT INTO public.tenants (tenant_name, slug, db_conn_str, is_active, updated_at)
VALUES ('$TenantName', '$TenantSlug', '$DbUrl', true, CURRENT_TIMESTAMP)
ON CONFLICT (slug) DO UPDATE
SET db_conn_str = EXCLUDED.db_conn_str,
    is_active = true,
    updated_at = CURRENT_TIMESTAMP;

SELECT id, tenant_name, slug, is_active, SUBSTRING(db_conn_str, 1, 40) || '...' as target_db
FROM public.tenants
WHERE slug = '$TenantSlug';
"@

$tmpSql = Join-Path $env:TEMP "upsert_tenant.sql"
Set-Content -Path $tmpSql -Value $sql -Encoding UTF8

try {
    & psql -d $MasterDbUrl -f $tmpSql
    Write-Host "Tenant '$TenantSlug' successfully configured to point to target database!" -ForegroundColor Green
} finally {
    if (Test-Path $tmpSql) {
        Remove-Item $tmpSql -Force
    }
}
