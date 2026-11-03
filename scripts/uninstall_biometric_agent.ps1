<#
.SYNOPSIS
    Uninstalls the Nembus Biometric Agent, stops running processes, cleans up
    extracted DLLs/runtime caches, removes auto-start registry entries, and
    optionally uninstalls the DigitalPersona kernel USB driver.
.PARAMETER RemoveDrivers
    If specified, removes DigitalPersona driver packages (oem*.inf) from the
    Windows Driver Store using pnputil (Requires Administrator).
#>
param(
    [switch]$RemoveDrivers
)

$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

Write-Host "======================================================" -ForegroundColor Yellow
Write-Host "  Uninstalling Nembus Biometric Agent & Runtime       " -ForegroundColor Yellow
Write-Host "======================================================" -ForegroundColor Yellow

# 1. Stop running biometric agent processes
Write-Host "[1/4] Terminating running biometric agent processes..." -ForegroundColor Gray
$processes = Get-Process -Name "nembus-biometric-agent" -ErrorAction SilentlyContinue
if ($processes) {
    $processes | Stop-Process -Force
    Write-Host "      Stopped $($processes.Count) running process(es)." -ForegroundColor Green
} else {
    Write-Host "      No active process found." -ForegroundColor DarkGray
}

# 2. Remove Windows Startup Registry Key (if present)
Write-Host "[2/5] Checking Windows Auto-start Registry..." -ForegroundColor Gray
$regPath = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run"
if (Get-ItemProperty -Path $regPath -Name "NembusBiometricAgent" -ErrorAction SilentlyContinue) {
    Remove-ItemProperty -Path $regPath -Name "NembusBiometricAgent" -Force
    Write-Host "      Removed auto-start registry key." -ForegroundColor Green
} else {
    Write-Host "      No auto-start registry entry found." -ForegroundColor DarkGray
}

# 2b. Remove Desktop / Start Menu shortcuts
Write-Host "[3/5] Removing Desktop and Start Menu shortcuts..." -ForegroundColor Gray
$shortcuts = @(
    (Join-Path $env:USERPROFILE "Desktop\Nembus Biometric Agent.lnk"),
    (Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs\Nembus Biometric Agent.lnk")
)
foreach ($lnk in $shortcuts) {
    if (Test-Path $lnk) {
        Remove-Item -LiteralPath $lnk -Force
        Write-Host "      Deleted: $lnk" -ForegroundColor Green
    }
}

$installDir = Join-Path $env:LOCALAPPDATA "Programs\Nembus"
if (Test-Path $installDir) {
    Remove-Item -Path $installDir -Recurse -Force
    Write-Host "      Deleted install folder: $installDir" -ForegroundColor Green
}

# 3. Clean up extracted runtime DLLs and caches
Write-Host "[4/5] Removing extracted runtime files and DLL caches..." -ForegroundColor Gray
$runtimeDir = Join-Path $env:LOCALAPPDATA "Nembus\biometrics"
if (Test-Path $runtimeDir) {
    Remove-Item -Path $runtimeDir -Recurse -Force
    Write-Host "      Deleted: $runtimeDir" -ForegroundColor Green
} else {
    Write-Host "      No runtime cache folder found at: $runtimeDir" -ForegroundColor DarkGray
}

# 4. Optional / Elevated: Remove Drivers from Windows Driver Store
if ($RemoveDrivers) {
    Write-Host "[5/5] Removing DigitalPersona kernel drivers from Driver Store..." -ForegroundColor Gray
    if (-not $isAdmin) {
        Write-Warning "Driver removal requires an Administrator PowerShell terminal. Rerun with admin rights."
    } else {
        $drivers = pnputil.exe /enum-drivers
        $oemList = @()
        $currentOem = ""
        foreach ($line in ($drivers -split "`r?`n")) {
            if ($line -match "Published Name:\s+(oem\d+\.inf)") {
                $currentOem = $matches[1]
            } elseif ($line -match "dpersona" -and $currentOem) {
                $oemList += $currentOem
                $currentOem = ""
            }
        }
        $oemList = $oemList | Select-Object -Unique
        if ($oemList.Count -gt 0) {
            foreach ($oem in $oemList) {
                Write-Host "      Deleting driver package: $oem..." -ForegroundColor Yellow
                pnputil.exe /delete-driver $oem /uninstall /force
            }
            Write-Host "      Driver packages uninstalled." -ForegroundColor Green
        } else {
            Write-Host "      No DigitalPersona OEM drivers found in Driver Store." -ForegroundColor DarkGray
        }
    }
} else {
    Write-Host "[5/5] Driver packages retained in Driver Store (pass -RemoveDrivers to wipe kernel drivers)." -ForegroundColor DarkGray
}

Write-Host "======================================================" -ForegroundColor Green
Write-Host "  Nembus Biometric Agent uninstalled successfully.    " -ForegroundColor Green
Write-Host "======================================================" -ForegroundColor Green

