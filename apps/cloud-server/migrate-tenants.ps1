# PowerShell script to migrate all tenant databases
# Usage: .\migrate-tenants.ps1 [-Down] [-MigrationsDir "./migrations"]

param(
    [switch]$Down,
    [string]$MigrationsDir = "./migrations"
)

# Load environment variables from .env file if it exists
if (Test-Path .env) {
    Get-Content .env | ForEach-Object {
        if ($_ -match '^([^=]+)=(.*)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim()
            [Environment]::SetEnvironmentVariable($name, $value, "Process")
        }
    }
}

# Check if MASTER_DB_URL is set
$masterDbUrl = $env:MASTER_DB_URL
if (-not $masterDbUrl) {
    Write-Host "Error: MASTER_DB_URL is not set" -ForegroundColor Red
    exit 1
}

Write-Host "Running migrations on all active tenant databases..." -ForegroundColor Green
Write-Host "Master DB URL: $masterDbUrl" -ForegroundColor Cyan

# Run the Go script
$downFlag = if ($Down) { "-down" } else { "" }
$dirFlag = "-dir", $MigrationsDir

$args = @()
if ($downFlag) { $args += $downFlag }
$args += $dirFlag

go run cmd/migrate-tenants/main.go $args

if ($LASTEXITCODE -ne 0) {
    Write-Host "Migration failed!" -ForegroundColor Red
    exit $LASTEXITCODE
}
