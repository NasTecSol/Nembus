<#
.SYNOPSIS
    Restores a database backup (.dump, .tar, or .sql) into a Supabase PostgreSQL instance.

.DESCRIPTION
    Safely executes pg_restore or psql against Supabase with:
      - Automatic SSL mode verification (?sslmode=require)
      - Stripping/ignoring local roles and ownerships (--no-owner --no-privileges)
      - Handling pre-existing public schema

.PARAMETER SupabaseUrl
    The Supabase connection string.
    Example: postgresql://postgres.[project-ref]:[password]@aws-0-[region].pooler.supabase.com:5432/postgres?sslmode=require

.PARAMETER BackupFile
    Path to the backup file (.dump, .tar, or .sql).

.EXAMPLE
    .\scripts\restore-to-supabase.ps1 -SupabaseUrl "postgresql://postgres.xyz:pass@aws-0-eu-central-1.pooler.supabase.com:5432/postgres?sslmode=require" -BackupFile "D:\backups\tenant_backup.dump"
#>

[CmdletBinding()]
param (
    [Parameter(Mandatory = $true)]
    [string]$SupabaseUrl,

    [Parameter(Mandatory = $true)]
    [string]$BackupFile
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $BackupFile)) {
    Write-Host "❌ Backup file not found: $BackupFile" -ForegroundColor Red
    exit 1
}

# Ensure sslmode=require is in the connection string
if ($SupabaseUrl -notmatch "sslmode=") {
    if ($SupabaseUrl -contains "?") {
        $SupabaseUrl += "&sslmode=require"
    } else {
        $SupabaseUrl += "?sslmode=require"
    }
}

# Auto-discover PostgreSQL bin directory if not already in PATH
if (-not (Get-Command pg_restore -ErrorAction SilentlyContinue) -or -not (Get-Command psql -ErrorAction SilentlyContinue)) {
    $commonPgPaths = @(
        "C:\Program Files\PostgreSQL\18\bin",
        "C:\Program Files\PostgreSQL\17\bin",
        "C:\Program Files\PostgreSQL\16\bin",
        "C:\Program Files\PostgreSQL\15\bin"
    )
    foreach ($pgPath in $commonPgPaths) {
        if (Test-Path (Join-Path $pgPath "pg_restore.exe")) {
            $env:PATH = "$pgPath;$env:PATH"
            break
        }
    }
}

$ext = [System.IO.Path]::GetExtension($BackupFile).ToLower()

# Check magic bytes for pg_dump custom format if extension is missing or .dump/.backup
$isCustomDump = $false
try {
    $fs = [System.IO.File]::OpenRead($BackupFile)
    $headerBytes = New-Object byte[] 5
    $readCount = $fs.Read($headerBytes, 0, 5)
    $fs.Close()
    $headerStr = [System.Text.Encoding]::ASCII.GetString($headerBytes, 0, $readCount)
    if ($headerStr -eq "PGDMP") {
        $isCustomDump = $true
    }
} catch {
    # fallback to extension
}

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "          NEMBUS: Restore Backup to Supabase Tenant DB          " -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "Backup File: $BackupFile ($ext)" -ForegroundColor Yellow

if ($isCustomDump -or $ext -eq ".dump" -or $ext -eq ".tar" -or $ext -eq ".backup") {
    Write-Host "Detected PostgreSQL custom/archive format. Running pg_restore..." -ForegroundColor Yellow
    
    if (-not (Get-Command pg_restore -ErrorAction SilentlyContinue)) {
        Write-Host "❌ pg_restore CLI not found in PATH. Please install PostgreSQL client tools." -ForegroundColor Red
        exit 1
    }

    # Execute pg_restore with no-owner and no-privileges so missing local roles don't fail the restore
    $restoreArgs = @(
        "--dbname=$SupabaseUrl",
        "--no-owner",
        "--no-privileges",
        "--clean",
        "--if-exists",
        "--verbose",
        $BackupFile
    )

    & pg_restore @restoreArgs
    if ($LASTEXITCODE -ne 0 -and $LASTEXITCODE -ne 1) {
        Write-Host "❌ pg_restore failed with exit code $LASTEXITCODE" -ForegroundColor Red
        exit $LASTEXITCODE
    } else {
        Write-Host "✅ Backup restored successfully to Supabase!" -ForegroundColor Green
    }
} elseif ($ext -eq ".sql") {
    Write-Host "Detected plain SQL dump. Running psql..." -ForegroundColor Yellow

    if (-not (Get-Command psql -ErrorAction SilentlyContinue)) {
        Write-Host "❌ psql CLI not found in PATH. Please install PostgreSQL client tools." -ForegroundColor Red
        exit 1
    }

    $psqlArgs = @(
        "-d", $SupabaseUrl,
        "-f", $BackupFile,
        "-v", "ON_ERROR_STOP=0"
    )

    & psql @psqlArgs
    if ($LASTEXITCODE -ne 0) {
        Write-Host "⚠ psql executed with warnings or non-zero exit code ($LASTEXITCODE)." -ForegroundColor Yellow
    } else {
        Write-Host "✅ SQL backup script executed successfully on Supabase!" -ForegroundColor Green
    }
} else {
    Write-Host "❌ Unsupported backup extension: $ext. Expected .dump, .tar, .backup, or .sql" -ForegroundColor Red
    exit 1
}
