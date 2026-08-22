[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$goFile = Join-Path $root 'local-stage2b-contract-diagnostic.go'

if (-not (Test-Path -LiteralPath $goFile -PathType Leaf)) {
    throw "Diagnostic Go source was not found: $goFile"
}

$databasePassword = Read-Host -Prompt 'nembus_e2e_user password' -AsSecureString
$deepSeekAPIKey = Read-Host -Prompt 'DeepSeek API key' -AsSecureString
$dbBSTR = [IntPtr]::Zero
$apiBSTR = [IntPtr]::Zero

try {
    $dbBSTR = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($databasePassword)
    $apiBSTR = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($deepSeekAPIKey)
    $env:NEMBUS_STAGE2B_DB_PASSWORD = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($dbBSTR)
    $env:NEMBUS_STAGE2B_DEEPSEEK_API_KEY = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($apiBSTR)

    Push-Location (Join-Path $root 'packages/core')
    try {
        # The Go program only opens the fixed disposable E2E database in
        # read-only mode and makes one DeepSeek provider request.
        go run ../../local-stage2b-contract-diagnostic.go
    }
    finally {
        Pop-Location
    }
}
finally {
    Remove-Item Env:NEMBUS_STAGE2B_DB_PASSWORD -ErrorAction SilentlyContinue
    Remove-Item Env:NEMBUS_STAGE2B_DEEPSEEK_API_KEY -ErrorAction SilentlyContinue
    if ($dbBSTR -ne [IntPtr]::Zero) {
        [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($dbBSTR)
    }
    if ($apiBSTR -ne [IntPtr]::Zero) {
        [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($apiBSTR)
    }
    $databasePassword = $null
    $deepSeekAPIKey = $null
}
