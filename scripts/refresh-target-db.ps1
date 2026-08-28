<#
.SYNOPSIS
    Refreshes successfully migrated SAP domains from STG database to Target database.

.DESCRIPTION
    Executes high-speed domain migration & synchronization preserving foreign key
    dependency chains, sequences, and multi-tenant constraints.

.PARAMETER SourceUrl
    Source (STG) PostgreSQL connection string (defaults to STG_DATABASE_URL or DATABASE_URL).

.PARAMETER TargetUrl
    Target PostgreSQL connection string (defaults to TARGET_DB_URL or PROD_DATABASE_URL).

.PARAMETER OrgId
    Organization ID to sync (0 = sync all organizations).

.PARAMETER Domains
    Comma-separated list of domains to sync (or 'all').
    Domains: uom, categories, brands, stores, users, uom_groups, products, barcodes, price_lists, inventory, partners, bp_addresses, sales_orders

.PARAMETER Mode
    Sync mode: 'truncate_copy' (default, clean fast refresh) or 'upsert'.

.PARAMETER DryRun
    Simulate execution and print row comparisons without modifying Target DB.

.EXAMPLE
    .\scripts\refresh-target-db.ps1 -DryRun
    .\scripts\refresh-target-db.ps1 -SourceUrl "postgres://user:pass@localhost:5432/stg_db" -TargetUrl "postgres://user:pass@localhost:5432/target_db"
#>

[CmdletBinding()]
param (
    [string]$SourceUrl = "",
    [string]$TargetUrl = "",
    [int]$OrgId = 0,
    [string]$Domains = "all",
    [ValidateSet("truncate_copy", "upsert")]
    [string]$Mode = "truncate_copy",
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

$RepoRoot = Split-Path -Parent $PSScriptRoot
$CloudServerDir = Join-Path $RepoRoot "apps\cloud-server"

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "       NEMBUS: Refresh Migrated Domains -> Target Database      " -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

# Check for Go binary or source
Set-Location $CloudServerDir

$ArgsList = @(
    "run", "./cmd/sync-target-db"
)

if ($SourceUrl -ne "") {
    $ArgsList += "-source-url=$SourceUrl"
}

if ($TargetUrl -ne "") {
    $ArgsList += "-target-url=$TargetUrl"
}

if ($OrgId -gt 0) {
    $ArgsList += "-org-id=$OrgId"
}

if ($Domains -ne "" -and $Domains -ne "all") {
    $ArgsList += "-domains=$Domains"
}

$ArgsList += "-mode=$Mode"

if ($DryRun) {
    $ArgsList += "-dry-run=true"
}

Write-Host "Executing Go sync utility from $CloudServerDir..." -ForegroundColor Yellow
& go @ArgsList

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Database sync failed with exit code $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
} else {
    Write-Host "✅ Target database refresh completed successfully!" -ForegroundColor Green
}
