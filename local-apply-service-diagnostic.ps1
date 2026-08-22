[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$diagnosticRoot = 'D:\nastecsol\Nembus'
$diagnosticURLVariable = 'NEMBUS_DIRECT_APPLY_DIAGNOSTIC_DATABASE_URL'
$databaseHost = '127.0.0.1'
$databasePort = 5432
$databaseName = 'nembus_e2e_tenant'
$databaseUser = 'nembus_e2e_user'

if ($databaseHost -ne '127.0.0.1' -or
    $databasePort -ne 5432 -or
    $databaseName -ne 'nembus_e2e_tenant' -or
    $databaseUser -ne 'nembus_e2e_user') {
    throw 'Safety gate rejected a database target other than 127.0.0.1:5432 / nembus_e2e_tenant / nembus_e2e_user.'
}

$securePassword = Read-Host -AsSecureString 'Password for nembus_e2e_user on local nembus_e2e_tenant'
$passwordBSTR = [IntPtr]::Zero
$plainPassword = $null

try {
    $passwordBSTR = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($securePassword)
    $plainPassword = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($passwordBSTR)
    if ([string]::IsNullOrEmpty($plainPassword)) {
        throw 'A PostgreSQL password is required.'
    }

    $encodedPassword = [Uri]::EscapeDataString($plainPassword)
    Set-Item -Path "Env:$diagnosticURLVariable" -Value "postgresql://${databaseUser}:${encodedPassword}@${databaseHost}:${databasePort}/${databaseName}?sslmode=disable"

    Set-Location -LiteralPath $diagnosticRoot
    & go run .\local-apply-service-diagnostic.go
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
