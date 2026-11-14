param(
    [string]$Version = "",
    [string]$CloudURL = "https://nembus.nashrms.com",
    [string]$AppUpdateKey = "nembus-pos-update-client-key-default",
    [string]$OutputDir = "build/bin",
    [ValidateSet("agent", "cli", "all")]
    [string]$Target = "all",
    [switch]$Console
)

$ErrorActionPreference = "Stop"

Write-Host "======================================================" -ForegroundColor Cyan
Write-Host "       Building Nembus SAP Sync & Migration Agent     " -ForegroundColor Cyan
Write-Host "======================================================" -ForegroundColor Cyan

# Resolve monorepo root directory
$RepoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $RepoRoot

# Resolve Target Output Directory
if (-not [System.IO.Path]::IsPathRooted($OutputDir)) {
    $OutputDir = Join-Path $RepoRoot $OutputDir
}
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}

# Determine Version if not provided
if (-not $Version) {
    try {
        $tags = git tag --sort=-v:refname 2>$null | Where-Object { $_ -like "sap-v*" }
        $latestTag = $tags | Select-Object -First 1
        if ($latestTag) {
            $Version = $latestTag.Replace("sap-", "")
        } else {
            $Version = "v1.0.0"
        }
    } catch {
        $Version = "v1.0.0"
    }
}

Write-Host "[Version]      $Version" -ForegroundColor Yellow
Write-Host "[CloudURL]     $CloudURL" -ForegroundColor Gray
Write-Host "[AppUpdateKey] $AppUpdateKey" -ForegroundColor Gray
Write-Host "[OutputDir]    $OutputDir" -ForegroundColor Gray
Write-Host "[Target]       $Target" -ForegroundColor Gray
Write-Host ""

# Construct ldflags
$LdFlagsList = @(
    "-s", "-w",
    "-X main.Version=$Version",
    "-X main.AppUpdateKey=$AppUpdateKey",
    "-X main.CloudURL=$CloudURL"
)

if (-not $Console -and $Target -eq "agent") {
    # If explicitly building non-console GUI agent without console window:
    # Note: Default for agent leaves console open for logs unless -Console is omitted and user requests GUI.
    # $LdFlagsList += "-H=windowsgui"
}

$LdFlags = $LdFlagsList -join " "

# 1. Build Main Agent Service (cmd/agent/main.go)
if ($Target -eq "agent" -or $Target -eq "all") {
    $AgentExe = Join-Path $OutputDir "sap-agent.exe"
    Write-Host "[Build] Compiling SAP Agent Service (apps/sap-agent/cmd/agent) -> $AgentExe..." -ForegroundColor Cyan
    
    go build -ldflags "$LdFlags" -o $AgentExe ./apps/sap-agent/cmd/agent
    if ($LASTEXITCODE -ne 0 -or -not (Test-Path $AgentExe)) {
        Write-Host "ERROR: Failed to build sap-agent.exe" -ForegroundColor Red
        exit 1
    }

    $hash = (Get-FileHash -Path $AgentExe -Algorithm SHA256).Hash.ToLower()
    $fileInfo = Get-Item $AgentExe
    $sizeMB = [math]::Round($fileInfo.Length / 1MB, 2)
    Write-Host "SUCCESS: Built $AgentExe ($sizeMB MB)" -ForegroundColor Green
    Write-Host "  SHA-256: $hash" -ForegroundColor DarkGray
    Write-Host ""

    # Also mirror into apps/sap-agent/bin/ for local convenience
    $LocalBinDir = Join-Path $RepoRoot "apps/sap-agent/bin"
    if (-not (Test-Path $LocalBinDir)) {
        New-Item -ItemType Directory -Path $LocalBinDir -Force | Out-Null
    }
    Copy-Item -Path $AgentExe -Destination (Join-Path $LocalBinDir "sap-agent.exe") -Force
}

# 2. Build CLI Tool (cmd/main.go)
if ($Target -eq "cli" -or $Target -eq "all") {
    $CliExe = Join-Path $OutputDir "sap-agent-cli.exe"
    Write-Host "[Build] Compiling SAP CLI Runner (apps/sap-agent/cmd) -> $CliExe..." -ForegroundColor Cyan
    
    $CliLdFlags = "-s -w -X main.version=$Version"
    go build -ldflags "$CliLdFlags" -o $CliExe ./apps/sap-agent/cmd
    if ($LASTEXITCODE -ne 0 -or -not (Test-Path $CliExe)) {
        Write-Host "ERROR: Failed to build sap-agent-cli.exe" -ForegroundColor Red
        exit 1
    }

    $hash = (Get-FileHash -Path $CliExe -Algorithm SHA256).Hash.ToLower()
    $fileInfo = Get-Item $CliExe
    $sizeMB = [math]::Round($fileInfo.Length / 1MB, 2)
    Write-Host "SUCCESS: Built $CliExe ($sizeMB MB)" -ForegroundColor Green
    Write-Host "  SHA-256: $hash" -ForegroundColor DarkGray
    Write-Host ""

    # Also mirror into apps/sap-agent/bin/
    $LocalBinDir = Join-Path $RepoRoot "apps/sap-agent/bin"
    Copy-Item -Path $CliExe -Destination (Join-Path $LocalBinDir "sap-agent-cli.exe") -Force
}

Write-Host "======================================================" -ForegroundColor Green
Write-Host "             SAP Agent Build Succeeded!              " -ForegroundColor Green
Write-Host "======================================================" -ForegroundColor Green
Write-Host "Quick Launch Commands:" -ForegroundColor Cyan
Write-Host "  Interactive Web UI Mode:  & '$OutputDir\sap-agent.exe'"
Write-Host "  Install Windows Service:  & '$OutputDir\sap-agent.exe' -service install"
Write-Host "  Start Windows Service:    & '$OutputDir\sap-agent.exe' -service start"
Write-Host "  CLI Downstream Sync:      & '$OutputDir\sap-agent-cli.exe' sync-downstream"
Write-Host "  CLI Outbox Watcher:       & '$OutputDir\sap-agent-cli.exe' daemon"

