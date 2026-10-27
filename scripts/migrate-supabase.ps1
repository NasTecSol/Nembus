<#
.SYNOPSIS
    Runs Atlas migrations or status checks against the Supabase database.

.PARAMETER Action
    Migration action: "status" (default), "apply", or "diff".

.PARAMETER SupabaseUrl
    The Supabase connection string.
    Defaults to the project's session pooler.

.EXAMPLE
    .\scripts\migrate-supabase.ps1 -Action status
    .\scripts\migrate-supabase.ps1 -Action apply
#>
param(
    [ValidateSet("status", "apply", "diff")]
    [string]$Action = "status",
    [string]$SupabaseUrl = "postgresql://postgres.gktqsuecolwccvhpmuqw:Nembus%402026@aws-0-ap-northeast-2.pooler.supabase.com:5432/postgres?sslmode=require"
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command atlas -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Atlas CLI not found in PATH. Please install Atlas." -ForegroundColor Red
    exit 1
}

$RepoRoot = Split-Path -Parent $PSScriptRoot
$CoreDir = Join-Path $RepoRoot "packages\core"

$env:SUPABASE_DB_URL = $SupabaseUrl

Write-Host "Running Atlas migrate $Action against Supabase..." -ForegroundColor Cyan

Push-Location $CoreDir
try {
    if ($Action -eq "status") {
        atlas migrate status --env supabase
    } elseif ($Action -eq "apply") {
        atlas migrate apply --env supabase
    } elseif ($Action -eq "diff") {
        atlas migrate diff --env supabase
    }
} finally {
    Pop-Location
}
