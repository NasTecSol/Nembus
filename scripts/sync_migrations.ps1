<#
.SYNOPSIS
    Unified Monorepo Database Sync Script.
    Synchronizes Atlas core migrations from packages/core/db/migrations to:
      - apps/cloud-server/migrations
      - apps/pos-client/migrations (preserving 99999999999999_pos_extensions.sql)
    Also runs sqlc generate in packages/core.
#>

$ErrorActionPreference = "Stop"
$ScriptDir = $PSScriptRoot
$RepoRoot = (Get-Item "$ScriptDir\..").FullName

$CoreDir          = Join-Path $RepoRoot "packages\core"
$CoreMigrations   = Join-Path $CoreDir "db\migrations"
$ServerMigrations = Join-Path $RepoRoot "apps\cloud-server\migrations"
$ClientMigrations = Join-Path $RepoRoot "apps\pos-client\migrations"

Write-Host "Starting Unified Monorepo Database Migration Sync..." -ForegroundColor Cyan

# 1. Ensure target directories exist
if (-not (Test-Path $ServerMigrations)) {
    New-Item -ItemType Directory -Path $ServerMigrations -Force | Out-Null
}
if (-not (Test-Path $ClientMigrations)) {
    New-Item -ItemType Directory -Path $ClientMigrations -Force | Out-Null
}

# 2. Sync Core migrations to apps/cloud-server/migrations
if (Test-Path $CoreMigrations) {
    Write-Host "Syncing Core migrations -> apps/cloud-server/migrations..." -ForegroundColor Yellow
    Get-ChildItem -Path $CoreMigrations | ForEach-Object {
        Copy-Item -Path $_.FullName -Destination $ServerMigrations -Force
    }
}

# 3. Sync Core migrations to apps/pos-client/migrations (preserving 99999999999999_pos_extensions.sql)
if (Test-Path $CoreMigrations) {
    Write-Host "Syncing Core migrations -> apps/pos-client/migrations..." -ForegroundColor Yellow
    Get-ChildItem -Path $CoreMigrations | ForEach-Object {
        Copy-Item -Path $_.FullName -Destination $ClientMigrations -Force
    }
}

# 4. Recalculate Atlas hashes if atlas CLI is installed
if (Get-Command atlas -ErrorAction SilentlyContinue) {
    Write-Host "Recalculating Atlas migration hash sums..." -ForegroundColor Yellow
    if (Test-Path "$CoreDir\atlas.hcl") {
        Push-Location "$CoreDir"
        atlas migrate hash | Out-Null
        Pop-Location
    }
}

# 5. Run sqlc generate in packages/core if sqlc is installed
if (Get-Command sqlc -ErrorAction SilentlyContinue) {
    Write-Host "Running sqlc generate in packages/core..." -ForegroundColor Yellow
    Push-Location "$CoreDir"
    sqlc generate
    Pop-Location
} else {
    Write-Host "sqlc command not found in PATH, skipping sqlc generate step." -ForegroundColor Gray
}

Write-Host "Unified Monorepo Database Sync Complete!" -ForegroundColor Green
