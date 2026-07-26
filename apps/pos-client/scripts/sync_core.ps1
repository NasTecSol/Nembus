<#
.SYNOPSIS
    Synchronizes the base schema and SQL queries from the core package into the POS client.
    Run this script whenever packages/core/migrations or packages/core/queries are updated.
#>

$CoreMigrations = "..\..\packages\core\migrations"
$CoreQueries    = "..\..\packages\core\queries"
$ClientMigrations = ".\migrations"
$ClientQueries    = ".\queries"

Write-Host "Syncing base schema from core..." -ForegroundColor Cyan
Copy-Item "$CoreMigrations\000001_base_schema.sql" "$ClientMigrations\000001_base_schema.sql" -Force

Write-Host "Syncing SQL queries from core..." -ForegroundColor Cyan
Copy-Item "$CoreQueries\*" "$ClientQueries\" -Recurse -Force

Write-Host "Running sqlc generate for client..." -ForegroundColor Cyan
sqlc generate

Write-Host "Core sync complete. Review any changes and commit." -ForegroundColor Green
