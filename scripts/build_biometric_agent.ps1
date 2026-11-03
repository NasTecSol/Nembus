param(
    [switch]$Console,
    [string]$OutputDir = "build/bin"
)

$ErrorActionPreference = "Stop"

Write-Host "======================================================" -ForegroundColor Cyan
Write-Host "  Building Nembus Biometric Agent (DigitalPersona)   " -ForegroundColor Cyan
Write-Host "======================================================" -ForegroundColor Cyan

if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}

$ExeName = "nembus-biometric-agent.exe"
$TargetPath = Join-Path $OutputDir $ExeName

# Flags: strip debug symbols (-s -w). If not console mode, hide Windows cmd window (-H=windowsgui)
$LdFlags = "-s -w"
if (-not $Console) {
    $LdFlags += " -H=windowsgui"
    Write-Host "[Mode] Production GUI (Console hidden for tray)" -ForegroundColor Green
} else {
    Write-Host "[Mode] Debug Console (Console window visible for logging)" -ForegroundColor Yellow
}

Write-Host "[Build] Compiling ./apps/biometric-agent -> $TargetPath..." -ForegroundColor Gray
go build -ldflags $LdFlags -o $TargetPath ./apps/biometric-agent

if ($LASTEXITCODE -eq 0 -and (Test-Path $TargetPath)) {
    $fileInfo = Get-Item $TargetPath
    $sizeMB = [math]::Round($fileInfo.Length / 1MB, 2)
    Write-Host "SUCCESS: Built $TargetPath ($sizeMB MB)" -ForegroundColor Green
    Write-Host ""
    Write-Host "Quick launch options:" -ForegroundColor Cyan
    Write-Host "  Tray Mode (Silent):     & '$TargetPath'"
    Write-Host "  Device Setup UI:        & '$TargetPath' --setup"
    Write-Host "  Driver Installer (UAC): & '$TargetPath' --install-drivers"
    Write-Host "  JSON Status Dump:       & '$TargetPath' --status"
} else {
    Write-Host "ERROR: Build failed with exit code $LASTEXITCODE" -ForegroundColor Red
    exit 1
}

