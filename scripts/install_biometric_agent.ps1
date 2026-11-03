# Installs Nembus Biometric Agent for the current user: copies binary, creates
# Desktop + Start Menu shortcuts with icon, and registers logon auto-start.
param(
    [string]$ExePath = "",
    [switch]$Launch
)

$ErrorActionPreference = "Stop"

if (-not $ExePath) {
    $ExePath = Join-Path $PSScriptRoot "..\build\bin\nembus-biometric-agent.exe"
    $ExePath = (Resolve-Path $ExePath -ErrorAction SilentlyContinue).Path
}
if (-not $ExePath -or -not (Test-Path $ExePath)) {
    Write-Host "Build the agent first: scripts\build_biometric_agent.ps1" -ForegroundColor Red
    Write-Host "Or pass -ExePath to nembus-biometric-agent.exe" -ForegroundColor Yellow
    exit 1
}

Write-Host "Installing Nembus Biometric Agent from: $ExePath" -ForegroundColor Cyan
& $ExePath --install
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

if ($Launch) {
    $installed = Join-Path $env:LOCALAPPDATA "Programs\Nembus\nembus-biometric-agent.exe"
    if (Test-Path $installed) {
        Start-Process -FilePath $installed -ArgumentList "--setup"
    } else {
        Start-Process -FilePath $ExePath -ArgumentList "--setup"
    }
}

Write-Host "Installation complete." -ForegroundColor Green
