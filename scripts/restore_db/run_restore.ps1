<#
.SYNOPSIS
    Interactive runner for the Nembus Postgres Backup Restorer (Go).
#>

param (
    [string]$KeyPath = "",
    [string]$SqlFile = "",
    [string]$DbPassword = "",
    [string]$DbUser = "nembus_admin",
    [string]$DbName = "sap"
)

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$exePath = Join-Path $scriptDir "restore_db.exe"

if (-not (Test-Path $exePath)) {
    Write-Host "Compiling restore_db.exe..." -ForegroundColor Yellow
    $env:GOWORK="off"
    Push-Location $scriptDir
    go build -o restore_db.exe .
    Pop-Location
}

$argsList = @()
if ($KeyPath) { $argsList += "-key=`"$KeyPath`"" }
if ($SqlFile) { $argsList += "-file=`"$SqlFile`"" }
if ($DbPassword) { $argsList += "-db-password=`"$DbPassword`"" }
if ($DbUser) { $argsList += "-db-user=`"$DbUser`"" }
if ($DbName) { $argsList += "-db-name=`"$DbName`"" }

& $exePath $argsList
