[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$diagnosticRoot = 'D:\nastecsol\Nembus'
$diagnosticURLVariable = 'NEMBUS_CATEGORY_DIAGNOSTIC_DATABASE_URL'
$databaseHost = '127.0.0.1'
$databasePort = 5432
$databaseName = 'nembus_e2e_tenant'
$databaseUser = 'nembus_e2e_user'
$securePassword = Read-Host -AsSecureString 'Password for nembus_e2e_user on local nembus_e2e_tenant'
$passwordBSTR = [IntPtr]::Zero
$plainPassword = $null

try {
    $passwordBSTR = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($securePassword)
    $plainPassword = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($passwordBSTR)
    $encodedPassword = [Uri]::EscapeDataString($plainPassword)
    Set-Item -Path "Env:$diagnosticURLVariable" -Value "postgresql://${databaseUser}:${encodedPassword}@${databaseHost}:${databasePort}/${databaseName}?sslmode=disable"

    Set-Location -LiteralPath $diagnosticRoot
    & go run .\local-category-store-diagnostic.go
    exit $LASTEXITCODE
}
finally {
    if ($passwordBSTR -ne [IntPtr]::Zero) {
        [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($passwordBSTR)
    }
    Remove-Item -Path "Env:$diagnosticURLVariable" -ErrorAction SilentlyContinue
    $plainPassword = $null
    $securePassword = $null
}
