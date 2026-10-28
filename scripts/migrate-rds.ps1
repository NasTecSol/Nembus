<#
.SYNOPSIS
    Runs Atlas migrations or status checks against the AWS RDS PostgreSQL database.

.PARAMETER Action
    Migration action: "status" (default), "apply", or "diff".

.PARAMETER RdsUrl
    The AWS RDS connection string.

.EXAMPLE
    .\scripts\migrate-rds.ps1 -Action status
    .\scripts\migrate-rds.ps1 -Action apply
#>
param(
    [ValidateSet("status", "apply", "diff")]
    [string]$Action = "status",
    [string]$RdsUrl = "postgresql://postgres:Nembus-2026@nembus-pg-db.cf86sus6qqzy.us-east-2.rds.amazonaws.com:5432/postgres?sslmode=require"
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command atlas -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Atlas CLI not found in PATH. Please install Atlas." -ForegroundColor Red
    exit 1
}

$RepoRoot = Split-Path -Parent $PSScriptRoot
$CoreDir = Join-Path $RepoRoot "packages\core"

$env:AWS_RDS_DB_URL = $RdsUrl

Write-Host "Running Atlas migrate $Action against AWS RDS..." -ForegroundColor Cyan

Push-Location $CoreDir
try {
    if ($Action -eq "status") {
        atlas migrate status --env aws_rds
    } elseif ($Action -eq "apply") {
        atlas migrate apply --env aws_rds
    } elseif ($Action -eq "diff") {
        atlas migrate diff --env aws_rds
    }
} finally {
    Pop-Location
}
