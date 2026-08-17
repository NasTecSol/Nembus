<#
.SYNOPSIS
    Unified Monorepo Database Sync Script.
    Synchronizes Atlas core migrations from packages/core/db/migrations to:
      - apps/cloud-server/migrations
      - apps/pos-client/migrations (adding -- +goose Up & StatementBegin/StatementEnd tags for PL/pgSQL functions)
    Also runs sqlc generate in packages/core.
#>

$ErrorActionPreference = "Stop"
$ScriptDir = $PSScriptRoot
$RepoRoot = (Get-Item "$ScriptDir\..").FullName

$CoreDir          = Join-Path $RepoRoot "packages\core"
$CoreMigrations   = Join-Path $CoreDir "db\migrations"
$ServerMigrations = Join-Path $RepoRoot "apps\cloud-server\migrations"
$ClientMigrations = Join-Path $RepoRoot "apps\pos-client\migrations"

$Utf8NoBom = New-Object System.Text.UTF8Encoding $false

Write-Host "Starting Unified Monorepo Database Migration Sync..." -ForegroundColor Cyan

# 1. Ensure target directories exist
if (-not (Test-Path $ServerMigrations)) {
    New-Item -ItemType Directory -Path $ServerMigrations -Force | Out-Null
}
if (-not (Test-Path $ClientMigrations)) {
    New-Item -ItemType Directory -Path $ClientMigrations -Force | Out-Null
}

# 2. Sync Core migrations to apps/cloud-server/migrations (Atlas format)
if (Test-Path $CoreMigrations) {
    Write-Host "Syncing Core migrations -> apps/cloud-server/migrations..." -ForegroundColor Yellow
    Get-ChildItem -Path $CoreMigrations | ForEach-Object {
        Copy-Item -Path $_.FullName -Destination $ServerMigrations -Force
    }
}

# 3. Sync Core migrations to apps/pos-client/migrations (Goose format with -- +goose Up and StatementBegin/End)
if (Test-Path $CoreMigrations) {
    Write-Host "Syncing Core migrations -> apps/pos-client/migrations..." -ForegroundColor Yellow
    Get-ChildItem -Path $CoreMigrations -Filter "*.sql" | ForEach-Object {
        $fileName = $_.Name
        
        # Ensure filename has an underscore description for Goose parser (e.g., 20260813124500_core_baseline.sql)
        if ($fileName -notlike "*_*") {
            $baseName = $_.BaseName
            $targetName = "${baseName}_core_baseline.sql"
        } else {
            $targetName = $fileName
        }
        
        $targetPath = Join-Path $ClientMigrations $targetName
        $content = [System.IO.File]::ReadAllText($_.FullName)

        # Wrap PL/pgSQL CREATE FUNCTION blocks with -- +goose StatementBegin and -- +goose StatementEnd
        $funcPattern = '(?ms)(CREATE (?:OR REPLACE )?FUNCTION "public"\."[^"]+" \([^\)]*\) [^\$]*?\$\$.*?\$\$;)'
        $content = [System.Text.RegularExpressions.Regex]::Replace($content, $funcPattern, "-- +goose StatementBegin`n`$1`n-- +goose StatementEnd")

        # Add Goose Up header if not present
        if (-not ($content.StartsWith("-- +goose Up"))) {
            $content = "-- +goose Up`n" + $content
        }

        # Write without UTF-8 BOM so Goose parses '-- +goose Up' cleanly
        [System.IO.File]::WriteAllText($targetPath, $content, $Utf8NoBom)
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
