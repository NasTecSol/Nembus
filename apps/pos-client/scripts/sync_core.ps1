<#
.SYNOPSIS
    Synchronizes the base schema from the core package into POS client and Cloud server.
    Run this script whenever packages/core/migrations or packages/core/queries are updated.
#>

$CoreMigrations = "..\..\packages\core\migrations"
$ClientMigrations = ".\migrations"
$ServerMigrations = "..\cloud-server\migrations"

Write-Host "Syncing base schema from core to pos-client..." -ForegroundColor Cyan
Copy-Item "$CoreMigrations\000001_base_schema.sql" "$ClientMigrations\000001_base_schema.sql" -Force

Write-Host "Syncing base schema from core to cloud-server..." -ForegroundColor Cyan
if (Test-Path "$ServerMigrations") {
    Copy-Item "$CoreMigrations\000001_base_schema.sql" "$ServerMigrations\000001_init_schema.sql" -Force
}

Write-Host "Running sqlc generate in core..." -ForegroundColor Cyan
Set-Location "..\..\packages\core"
sqlc generate
Set-Location "..\..\apps\pos-client"

Write-Host "Core sync complete. Review any changes and commit." -ForegroundColor Green
