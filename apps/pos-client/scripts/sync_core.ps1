<#
.SYNOPSIS
    Synchronizes schema, migrations, and generated code from packages/core into POS client and Cloud server.
    Run this script whenever packages/core/db/schema or packages/core/queries are updated.
#>

$ScriptDir = $PSScriptRoot
$RepoRoot = (Get-Item "$ScriptDir\..\..\..").FullName
$CoreDir = Join-Path $RepoRoot "packages\core"
$CoreMigrations = Join-Path $CoreDir "db\migrations"
$ClientMigrations = Join-Path $RepoRoot "apps\pos-client\migrations"
$ServerMigrations = Join-Path $RepoRoot "apps\cloud-server\migrations"

Write-Host "Syncing Atlas migrations from core to pos-client..." -ForegroundColor Cyan
if (Test-Path "$CoreMigrations") {
    if (-not (Test-Path "$ClientMigrations")) {
        New-Item -ItemType Directory -Path "$ClientMigrations" -Force | Out-Null
    }
    Get-ChildItem -Path "$CoreMigrations" | ForEach-Object {
        Copy-Item $_.FullName "$ClientMigrations\$($_.Name)" -Force
    }
}

Write-Host "Syncing Atlas migrations from core to cloud-server..." -ForegroundColor Cyan
if (Test-Path "$CoreMigrations") {
    if (-not (Test-Path "$ServerMigrations")) {
        New-Item -ItemType Directory -Path "$ServerMigrations" -Force | Out-Null
    }
    Get-ChildItem -Path "$CoreMigrations" | ForEach-Object {
        Copy-Item $_.FullName "$ServerMigrations\$($_.Name)" -Force
    }
}

Write-Host "Running sqlc generate in packages/core..." -ForegroundColor Cyan
Push-Location "$CoreDir"
sqlc generate
Pop-Location

Write-Host "Core sync complete. Review any changes and commit." -ForegroundColor Green
