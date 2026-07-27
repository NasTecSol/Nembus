<#
.SYNOPSIS
    Synchronizes the base schema from the core package into POS client and Cloud server.
    Run this script whenever packages/core/migrations or packages/core/queries are updated.
#>

$ScriptDir = $PSScriptRoot
$RepoRoot = (Get-Item "$ScriptDir\..\..\..").FullName
$CoreDir = Join-Path $RepoRoot "packages\core"
$CoreMigrations = Join-Path $CoreDir "migrations"
$ClientMigrations = Join-Path $RepoRoot "apps\pos-client\migrations"
$ServerMigrations = Join-Path $RepoRoot "apps\cloud-server\migrations"

Write-Host "Syncing base schema & migrations from core to pos-client..." -ForegroundColor Cyan
Get-ChildItem -Path "$CoreMigrations" -Filter "*.sql" | ForEach-Object {
    Copy-Item $_.FullName "$ClientMigrations\$($_.Name)" -Force
}

Write-Host "Syncing base schema & migrations from core to cloud-server..." -ForegroundColor Cyan
if (Test-Path "$ServerMigrations") {
    Copy-Item "$CoreMigrations\000001_base_schema.sql" "$ServerMigrations\000001_init_schema.sql" -Force
    Get-ChildItem -Path "$CoreMigrations" -Filter "*.sql" | Where-Object { $_.Name -ne "000001_base_schema.sql" } | ForEach-Object {
        Copy-Item $_.FullName "$ServerMigrations\$($_.Name)" -Force
    }
}

Write-Host "Running sqlc generate in core..." -ForegroundColor Cyan
Push-Location "$CoreDir"
sqlc generate
Pop-Location

Write-Host "Core sync complete. Review any changes and commit." -ForegroundColor Green
