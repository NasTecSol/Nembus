<#
.SYNOPSIS
    Cleans build cache and temporary folders to start fresh for Wails development.
.DESCRIPTION
    Stops running Wails/Nembus processes, deletes front-end and back-end build caches,
    and optionally performs deep cleaning of dependencies or local database resets.
.PARAMETER Deep
    Performs a deep clean. Deletes 'node_modules' in the frontend and clears the Go module cache.
.PARAMETER ResetDb
    Resets the local embedded Postgres database by deleting its data directory.
.EXAMPLE
    .\clean.ps1
    Runs the standard clean (recommended before starting wails dev).
.EXAMPLE
    .\clean.ps1 -Deep -ResetDb
    Performs a full deep clean of dependencies and database.
#>
[CmdletBinding()]
param(
    [switch]$Deep,
    [switch]$ResetDb
)

$ErrorActionPreference = "Continue"

Write-Host "=== Starting Wails Dev Fresh Clean ===" -ForegroundColor Blue

# 1. Stop active processes to prevent file locking
Write-Host "Stopping any running Nembus, Wails, or PostgreSQL processes..." -ForegroundColor Yellow
Stop-Process -Name "NEMBUS" -ErrorAction SilentlyContinue
Stop-Process -Name "wails" -ErrorAction SilentlyContinue
Stop-Process -Name "postgres" -ErrorAction SilentlyContinue
Start-Sleep -Seconds 2

# Helper function to remove a path safely
function Remove-DirectorySafe {
    param (
        [string]$Path
    )
    if (Test-Path $Path) {
        Write-Host "Removing: $Path" -ForegroundColor Cyan
        try {
            Remove-Item -Path $Path -Recurse -Force -ErrorAction Stop
            Write-Host "[OK] Removed $Path" -ForegroundColor Green
        } catch {
            Write-Warning "[WARN] Could not fully remove '$Path'. It might be locked by another process: $_"
        }
    }
}

# 2. Clean frontend build outputs and caches
Write-Host ""
Write-Host "Cleaning frontend build and cache folders..." -ForegroundColor Yellow
Remove-DirectorySafe -Path "$PSScriptRoot\frontend\dist"
Remove-DirectorySafe -Path "$PSScriptRoot\frontend\.angular"
Remove-DirectorySafe -Path "$PSScriptRoot\frontend\wailsjs"
Remove-DirectorySafe -Path "$PSScriptRoot\frontend\node_modules\.cache"

# 3. Clean wails/Go build outputs and local runtime caches
Write-Host ""
Write-Host "Cleaning backend build outputs and runtime caches..." -ForegroundColor Yellow
Remove-DirectorySafe -Path "$PSScriptRoot\build\bin"

# Delete embedded postgres runtime dir if exists (often causes startup lock issues)
$postgresRuntime = Join-Path $Home ".nembus\data\runtime"
Remove-DirectorySafe -Path $postgresRuntime

# 4. Run Go build cache clean
Write-Host ""
Write-Host "Running go clean..." -ForegroundColor Yellow
try {
    go clean -cache -testcache
    Write-Host "[OK] Go build cache cleaned successfully" -ForegroundColor Green
} catch {
    Write-Warning "[WARN] Failed to clean Go build cache: $_"
}

# 5. Deep Clean Options
if ($Deep) {
    Write-Host ""
    Write-Host "=== Deep Clean Requested ===" -ForegroundColor Magenta
    
    # Delete node_modules
    Write-Host "Removing frontend/node_modules..." -ForegroundColor Yellow
    Remove-DirectorySafe -Path "$PSScriptRoot\frontend\node_modules"
    
    # Go mod cache clean
    Write-Host "Cleaning Go module cache (this may take a moment)..." -ForegroundColor Yellow
    try {
        go clean -modcache
        Write-Host "[OK] Go module cache cleaned successfully" -ForegroundColor Green
    } catch {
        Write-Warning "[WARN] Failed to clean Go module cache: $_"
    }
}

# 6. Reset Database Option
if ($ResetDb) {
    Write-Host ""
    Write-Host "=== Database & Setup Reset Requested ===" -ForegroundColor Magenta
    
    # Remove Postgres data directory
    $postgresData = Join-Path $Home ".nembus\data"
    Write-Host "Removing local database data directory: $postgresData" -ForegroundColor Yellow
    Remove-DirectorySafe -Path $postgresData

    # Remove Setup Done marker to force setup wizard on launch
    $setupMarker = Join-Path $Home ".nembus\.setup_done"
    if (Test-Path $setupMarker) {
        Write-Host "Removing setup marker file: $setupMarker" -ForegroundColor Yellow
        try {
            Remove-Item -Path $setupMarker -Force -ErrorAction Stop
            Write-Host "[OK] Removed setup marker file" -ForegroundColor Green
        } catch {
            Write-Warning "[WARN] Could not remove setup marker file: $_"
        }
    }

    # Remove Device Config to ensure a completely fresh setup experience
    $deviceConfig = Join-Path $Home ".nembus\device_config.json"
    if (Test-Path $deviceConfig) {
        Write-Host "Removing device configuration file: $deviceConfig" -ForegroundColor Yellow
        try {
            Remove-Item -Path $deviceConfig -Force -ErrorAction Stop
            Write-Host "[OK] Removed device configuration file" -ForegroundColor Green
        } catch {
            Write-Warning "[WARN] Could not remove device configuration file: $_"
        }
    }
}

Write-Host ""
Write-Host "=== Clean Complete! Ready for a fresh start ===" -ForegroundColor Green
Write-Host "To start development, run: wails dev" -ForegroundColor Cyan
