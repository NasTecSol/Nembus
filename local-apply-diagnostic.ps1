Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# Targeted, one-request local diagnostic for the disposable E2E tenant. This
# script deliberately does not create/approve/reject suggestions, run SAP, or
# start any enrichment provider/worker. Its only non-auth business mutation is
# the one real apply endpoint call, if the selected approved E2E suggestion can
# be applied successfully.

$repoRoot = if (-not [string]::IsNullOrWhiteSpace($PSScriptRoot)) { $PSScriptRoot } else { Split-Path -Parent $MyInvocation.MyCommand.Path }
if ([string]::IsNullOrWhiteSpace($repoRoot)) { throw 'Could not resolve repository root.' }
Set-Location -LiteralPath $repoRoot

$localHost = '127.0.0.1'
$localPort = 5432
$roleName = 'nembus_e2e_user'
$masterDatabase = 'nembus_e2e_master'
$tenantDatabase = 'nembus_e2e_tenant'
$tenantSlug = 'e2e-enrichment'
$fixtureMarker = 'local-e2e-runner'
$organizationCode = 'E2E-ENRICHMENT-ORG'
$roleCode = 'E2E-ENRICHMENT-REVIEWER'
$reviewerUsername = 'e2e_enrichment_reviewer'
$reviewerEmail = 'e2e-enrichment-reviewer@nembus.invalid'
$httpPort = 18080
$grpcPort = 19051
$runID = "apply-diagnostic-{0}-{1}" -f [DateTime]::UtcNow.ToString('yyyyMMddHHmmss'), [Guid]::NewGuid().ToString('N').Substring(0, 8)

$phase = 'INITIALIZATION'
$dbPasswordSecure = $null
$jwtSecret = $null
$fixturePassword = $null
$userToken = $null
$dbPasswordPlainForServer = $null
$psqlExe = $null
$goExe = $null
$serverProcess = $null
$runtimeRoot = $null
$stdoutLog = $null
$stderrLog = $null
$cloudServerBinary = $null
$bcryptHelperBinary = $null
$organizationID = 0
$reviewerID = 0
$suggestion = $null
$preApplySnapshot = $null
$postApplySnapshot = $null
$applyResponse = $null
$applyRequestCount = 0
$logStartLines = @{}
$originalEnvironment = @{}

function Convert-SecureStringToPlainText {
    param([Parameter(Mandatory = $true)][Security.SecureString]$SecureString)
    $bstr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($SecureString)
    try { return [Runtime.InteropServices.Marshal]::PtrToStringBSTR($bstr) }
    finally { [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr) }
}

function ConvertTo-PgSqlLiteral {
    param([Parameter(Mandatory = $true)][AllowEmptyString()][string]$Value)
    return "'" + $Value.Replace("'", "''") + "'"
}

function Resolve-Tool {
    param([Parameter(Mandatory = $true)][string]$Name)
    $command = Get-Command $Name -ErrorAction SilentlyContinue
    if ($null -eq $command -or $command.CommandType -ne 'Application') { throw "Required local tool '$Name' was not found." }
    return $command.Source
}

function Save-ProcessEnvironment {
    param([Parameter(Mandatory = $true)][string[]]$Names)
    foreach ($name in $Names) {
        $value = [Environment]::GetEnvironmentVariable($name, 'Process')
        if ($null -ne $value) { $originalEnvironment[$name] = $value }
    }
}

function Restore-ProcessEnvironment {
    param([Parameter(Mandatory = $true)][string[]]$Names)
    foreach ($name in $Names) {
        if ($originalEnvironment.ContainsKey($name)) { [Environment]::SetEnvironmentVariable($name, $originalEnvironment[$name], 'Process') }
        else { [Environment]::SetEnvironmentVariable($name, $null, 'Process') }
    }
}

function New-DatabaseUrl {
    param([Parameter(Mandatory = $true)][string]$Database, [Parameter(Mandatory = $true)][string]$Password)
    $encodedPassword = [Uri]::EscapeDataString($Password)
    return "postgres://${roleName}:$encodedPassword@${localHost}:${localPort}/${Database}?sslmode=disable"
}

function Test-DisposableDsn {
    param([Parameter(Mandatory = $true)][string]$Dsn, [Parameter(Mandatory = $true)][string]$ExpectedDatabase)
    try {
        $uri = [Uri]$Dsn
        $userInfo = $uri.UserInfo.Split(':')[0]
        $database = $uri.AbsolutePath.Trim('/')
        return ($uri.Host -eq $localHost -and $uri.Port -eq $localPort -and $userInfo -eq $roleName -and $database -eq $ExpectedDatabase)
    }
    catch { return $false }
}

function Assert-LocalSafetyGate {
    if ($localHost -ne '127.0.0.1' -or $localPort -ne 5432 -or $masterDatabase -ne 'nembus_e2e_master' -or $tenantDatabase -ne 'nembus_e2e_tenant' -or $tenantSlug -ne 'e2e-enrichment' -or $roleName -ne 'nembus_e2e_user') {
        throw 'Disposable diagnostic constants were changed; refusing to continue.'
    }
}

function Invoke-Psql {
    param(
        [Parameter(Mandatory = $true)][string]$Database,
        [Parameter(Mandatory = $true)][Security.SecureString]$Password,
        [Parameter(Mandatory = $true)][string]$Sql
    )
    $plainPassword = $null
    try {
        $plainPassword = Convert-SecureStringToPlainText -SecureString $Password
        [Environment]::SetEnvironmentVariable('PGPASSWORD', $plainPassword, 'Process')
        $arguments = @('-X', '-w', '-q', '-t', '-A', '-v', 'ON_ERROR_STOP=1', '-h', $localHost, '-p', [string]$localPort, '-U', $roleName, '-d', $Database)
        $output = $Sql | & $psqlExe @arguments 2>&1
        if ($LASTEXITCODE -ne 0) { throw "psql failed for disposable database '$Database' with exit code $LASTEXITCODE." }
        return (($output | ForEach-Object { $_.ToString() }) -join "`n").Trim()
    }
    finally {
        $plainPassword = $null
        Remove-Item -LiteralPath 'Env:PGPASSWORD' -ErrorAction SilentlyContinue
    }
}

function Get-JsonSql {
    param([Parameter(Mandatory = $true)][string]$Database, [Parameter(Mandatory = $true)][string]$Sql)
    $raw = Invoke-Psql -Database $Database -Password $dbPasswordSecure -Sql $Sql
    if ([string]::IsNullOrWhiteSpace($raw)) { return $null }
    return $raw | ConvertFrom-Json
}

function New-LocalJwtSecret {
    $bytes = New-Object 'System.Byte[]' 32
    $rng = [Security.Cryptography.RandomNumberGenerator]::Create()
    try { $rng.GetBytes($bytes) } finally { $rng.Dispose() }
    return [Convert]::ToBase64String($bytes)
}

function Get-SanitizedText {
    param([AllowNull()][string]$Text)
    if ($null -eq $Text) { return '' }
    $sanitized = $Text -replace '(?i)Bearer\s+\S+', 'Bearer [redacted]'
    $sanitized = $sanitized -replace '(?i)postgres(?:ql)?://\S+', 'postgres://[redacted]'
    $sanitized = $sanitized -replace '(?i)\b(password|passwd|pwd|jwt[_ -]?secret|api[_ -]?key|authorization)\s*[:=]\s*("[^"]*"|''[^'']*''|\S+)', '$1=[redacted]'
    $sanitized = $sanitized -replace '(?<![A-Za-z0-9_-])[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}(?![A-Za-z0-9_-])', '[redacted-jwt]'
    foreach ($secret in @($jwtSecret, $fixturePassword, $userToken, $dbPasswordPlainForServer)) {
        if (-not [string]::IsNullOrWhiteSpace($secret)) { $sanitized = $sanitized.Replace($secret, '[redacted]') }
    }
    return $sanitized
}

function Get-ApiResponse {
    param(
        [Parameter(Mandatory = $true)][string]$Method,
        [Parameter(Mandatory = $true)][string]$Path,
        [hashtable]$Headers = @{},
        [AllowNull()][string]$Body = $null
    )
    $requestUri = "http://127.0.0.1:${httpPort}${Path}"
    $readResponse = {
        param([Parameter(Mandatory = $true)][System.Net.HttpWebResponse]$Response)
        $responseBody = ''
        $responseStream = $null
        $responseReader = $null
        try {
            $responseStream = $Response.GetResponseStream()
            if ($null -ne $responseStream) {
                $responseReader = [System.IO.StreamReader]::new($responseStream, [Text.Encoding]::UTF8, $true)
                $responseBody = $responseReader.ReadToEnd()
            }
        }
        finally {
            if ($null -ne $responseReader) { $responseReader.Dispose() }
            elseif ($null -ne $responseStream) { $responseStream.Dispose() }
        }
        $json = $null
        if (-not [string]::IsNullOrWhiteSpace($responseBody)) { try { $json = $responseBody | ConvertFrom-Json } catch { } }
        return [pscustomobject]@{ StatusCode = [int]$Response.StatusCode; Body = $responseBody; Json = $json }
    }

    $request = [System.Net.HttpWebRequest][System.Net.WebRequest]::Create($requestUri)
    $request.Proxy = $null
    $request.Method = $Method
    $request.Timeout = 20000
    $request.ReadWriteTimeout = 20000
    try {
        foreach ($key in $Headers.Keys) { $request.Headers.Set([string]$key, [string]$Headers[$key]) }
        if (-not [string]::IsNullOrEmpty($Body)) {
            $bodyBytes = [Text.Encoding]::UTF8.GetBytes($Body)
            $request.ContentType = 'application/json'
            $request.ContentLength = $bodyBytes.Length
            $requestStream = $null
            try { $requestStream = $request.GetRequestStream(); $requestStream.Write($bodyBytes, 0, $bodyBytes.Length) }
            finally { if ($null -ne $requestStream) { $requestStream.Dispose() } }
        }
        $response = [System.Net.HttpWebResponse]$request.GetResponse()
        try { return & $readResponse $response } finally { $response.Dispose() }
    }
    catch [System.Net.WebException] {
        if ($null -ne $_.Exception.Response) {
            $response = [System.Net.HttpWebResponse]$_.Exception.Response
            try { return & $readResponse $response } finally { $response.Dispose() }
        }
        throw ('HTTP transport failure: ' + (Get-SanitizedText $_.Exception.Message))
    }
}

function Wait-Health {
    for ($attempt = 1; $attempt -le 45; $attempt++) {
        if ($null -ne $serverProcess -and $serverProcess.HasExited) { throw 'cloud-server exited before health became ready.' }
        try {
            $response = Get-ApiResponse -Method 'GET' -Path '/health'
            if ($response.StatusCode -eq 200 -and $null -ne $response.Json -and $response.Json.status -eq 'OK') { return }
        }
        catch { }
        Start-Sleep -Seconds 1
    }
    throw 'cloud-server health timed out on localhost port 18080.'
}

function Stop-CloudServer {
    if ($null -ne $serverProcess) {
        try {
            if (-not $serverProcess.HasExited) { Stop-Process -Id $serverProcess.Id -ErrorAction Stop }
        }
        catch { }
        $serverProcess = $null
    }
}

function Get-LogLineCount {
    param([AllowEmptyString()][string]$Path)
    if ([string]::IsNullOrWhiteSpace($Path) -or -not (Test-Path -LiteralPath $Path)) { return 0 }
    return @(Get-Content -LiteralPath $Path -ErrorAction Stop).Count
}

function Get-ApplyLogWindow {
    $lines = New-Object 'System.Collections.Generic.List[string]'
    foreach ($entry in @(@{ Name = 'STDOUT'; Path = $stdoutLog }, @{ Name = 'STDERR'; Path = $stderrLog })) {
        $skip = if ($logStartLines.ContainsKey($entry.Name)) { [int]$logStartLines[$entry.Name] } else { 0 }
        if (-not (Test-Path -LiteralPath $entry.Path)) { continue }
        $newLines = @(Get-Content -LiteralPath $entry.Path -ErrorAction Stop | Select-Object -Skip $skip)
        foreach ($line in $newLines) { [void]$lines.Add(("{0}: {1}" -f $entry.Name, (Get-SanitizedText ([string]$line)))) }
    }
    return @($lines)
}

function Get-ApplyFailureLayer {
    param([Parameter(Mandatory = $true)][string]$ErrorText)
    if ($ErrorText -match '(?i)lock enrichment suggestion|LockProductEnrichmentSuggestion') { return 'SUGGESTION_LOCK' }
    if ($ErrorText -match '(?i)lock current enrichment product|LockProductForEnrichmentApplication') { return 'PRODUCT_LOCK' }
    if ($ErrorText -match '(?i)source is stale|source context changed|FingerprintSnapshot') { return 'STALE_FINGERPRINT_RECHECK' }
    if ($ErrorText -match '(?i)brand target|ResolveBrandApplicationTarget|LockBrandForEnrichment') { return 'BRAND_REVALIDATION' }
    if ($ErrorText -match '(?i)category target|ResolveCategoryApplicationTarget|LockCategoryForEnrichment') { return 'CATEGORY_REVALIDATION' }
    if ($ErrorText -match '(?i)apply narrow product enrichment fields|ApplyProductEnrichmentFields') { return 'PRODUCT_UPDATE' }
    if ($ErrorText -match '(?i)mark enrichment suggestion applied|MarkProductEnrichmentSuggestionApplied') { return 'MARK_APPLIED' }
    if ($ErrorText -match '(?i)write enrichment application audit|InsertProductEnrichmentReviewAudit') { return 'AUDIT_INSERT' }
    if ($ErrorText -match '(?i)commit enrichment application|commit already-applied') { return 'TRANSACTION_COMMIT' }
    if ($ErrorText -match '(?i)scan|cannot scan|SQLSTATE 22') { return 'SQL_SCAN' }
    if ($ErrorText -match '(?i)foreign key|SQLSTATE 23503') { return 'FOREIGN_KEY' }
    if ($ErrorText -match '(?i)not null|SQLSTATE 23502') { return 'NOT_NULL' }
    return 'OTHER_LOGGED_ERROR'
}

function New-BcryptHash {
    param([Parameter(Mandatory = $true)][string]$Password)
    $helper = Join-Path $repoRoot 'local-apply-diagnostic.go'
    if (-not (Test-Path -LiteralPath $helper)) { throw 'The required local bcrypt helper is missing.' }
    $old = [Environment]::GetEnvironmentVariable('NEMBUS_APPLY_DIAGNOSTIC_PASSWORD', 'Process')
    try {
        [Environment]::SetEnvironmentVariable('NEMBUS_APPLY_DIAGNOSTIC_PASSWORD', $Password, 'Process')
        Push-Location -LiteralPath (Join-Path $repoRoot 'apps/cloud-server')
        try {
            $hashOutput = & $goExe run $helper 2>&1
            if ($LASTEXITCODE -ne 0) { throw 'Unable to generate the disposable E2E bcrypt fixture hash.' }
        }
        finally { Pop-Location }
        return (($hashOutput | ForEach-Object { $_.ToString() }) -join '').Trim()
    }
    finally {
        if ($null -ne $old) { [Environment]::SetEnvironmentVariable('NEMBUS_APPLY_DIAGNOSTIC_PASSWORD', $old, 'Process') }
        else { [Environment]::SetEnvironmentVariable('NEMBUS_APPLY_DIAGNOSTIC_PASSWORD', $null, 'Process') }
    }
}

function Get-SelectedSuggestion {
    $organizationCodeLiteral = ConvertTo-PgSqlLiteral $organizationCode
    $markerLiteral = ConvertTo-PgSqlLiteral $fixtureMarker
    return Get-JsonSql -Database $tenantDatabase -Sql @"
SELECT jsonb_build_object(
  'suggestion_id', s.id,
  'product_id', p.id,
  'source_item_code', s.source_item_code,
  'status', s.status
)::text
FROM product_enrichment_suggestions s
JOIN products p ON p.id = s.product_id AND p.organization_id = s.organization_id
JOIN organizations o ON o.id = s.organization_id
WHERE s.organization_id = $organizationID
  AND o.code = $organizationCodeLiteral
  AND o.metadata->>'e2e_runner' = $markerLiteral
  AND s.source_item_code LIKE 'E2E-PANTENE-%'
  AND s.status = 'approved'
  AND s.applied_at IS NULL
  AND NULLIF(btrim(s.contract_version), '') IS NOT NULL
  AND p.product_type IN ('standard', 'raw_material', 'fixed_asset', 'finished_good')
  AND btrim(s.source_item_name) <> ''
  AND (s.proposed_brand IS NOT NULL OR s.proposed_category IS NOT NULL OR s.proposed_description IS NOT NULL)
ORDER BY s.updated_at DESC NULLS LAST, s.created_at DESC NULLS LAST, s.id DESC
LIMIT 1;
"@
}

function Get-ApplySnapshot {
    param([Parameter(Mandatory = $true)][int]$SuggestionID)
    return Get-JsonSql -Database $tenantDatabase -Sql @"
SELECT jsonb_build_object(
  'suggestion_status', s.status,
  'product', jsonb_build_object('brand_id', p.brand_id, 'category_id', p.category_id, 'description', p.description),
  'apply_audit_count', (
    SELECT count(*) FROM audit_logs a
    WHERE a.organization_id = s.organization_id
      AND a.table_name = 'product_enrichment_suggestions'
      AND a.record_id = s.id::text
      AND a.new_values->>'event' = 'product_enrichment.applied'
  )
)::text
FROM product_enrichment_suggestions s
JOIN products p ON p.id = s.product_id AND p.organization_id = s.organization_id
WHERE s.organization_id = $organizationID AND s.id = $SuggestionID;
"@
}

function Assert-ApplyPermission {
    $permission = Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql @"
SELECT EXISTS (
  SELECT 1
  FROM users u
  JOIN user_roles ur ON ur.user_id = u.id
  JOIN role_permissions rp ON rp.role_id = ur.role_id
  JOIN permissions p ON p.id = rp.permission_id
  WHERE u.id = $reviewerID AND u.organization_id = $organizationID AND p.code = 'product_enrichment:apply'
);
"@
    if ($permission.Trim().ToLowerInvariant() -ne 't') { throw 'Disposable reviewer does not have product_enrichment:apply.' }
}

try {
    Save-ProcessEnvironment @('PGPASSWORD', 'JWT_SECRET', 'MASTER_DB_URL', 'ENV', 'PORT', 'GRPC_PORT', 'ZATCA_ENABLED', 'ENRICHMENT_ENABLED', 'ENRICHMENT_PROVIDER', 'DEEPSEEK_API_KEY', 'DEEPSEEK_MODEL', 'DEEPSEEK_BASE_URL', 'OPENAI_API_KEY', 'OPENAI_ENRICHMENT_MODEL', 'OPENAI_ENRICHMENT_TIMEOUT', 'OPENAI_ENRICHMENT_WORKER_INTERVAL', 'OPENAI_ENRICHMENT_BATCH_SIZE', 'OPENAI_ENRICHMENT_MAX_RETRIES', 'NEMBUS_APPLY_DIAGNOSTIC_PASSWORD')
    $phase = 'LOCAL_SAFETY_GATE'
    Assert-LocalSafetyGate
    $psqlExe = Resolve-Tool 'psql'
    $goExe = Resolve-Tool 'go'

    $dbPasswordSecure = Read-Host 'nembus_e2e_user PostgreSQL password' -AsSecureString
    if ($null -eq $dbPasswordSecure) { throw 'The disposable PostgreSQL password is required.' }

    $phase = 'DATABASE_SAFETY_VERIFICATION'
    $dbPasswordPlainForServer = Convert-SecureStringToPlainText $dbPasswordSecure
    $masterDsn = New-DatabaseUrl -Database $masterDatabase -Password $dbPasswordPlainForServer
    $tenantDsn = New-DatabaseUrl -Database $tenantDatabase -Password $dbPasswordPlainForServer
    if (-not (Test-DisposableDsn -Dsn $masterDsn -ExpectedDatabase $masterDatabase) -or -not (Test-DisposableDsn -Dsn $tenantDsn -ExpectedDatabase $tenantDatabase)) { throw 'Constructed database DSN failed the exact disposable loopback gate.' }
    if ((Invoke-Psql -Database $masterDatabase -Password $dbPasswordSecure -Sql 'SELECT 1;') -ne '1') { throw 'Master disposable database authentication failed.' }
    if ((Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql 'SELECT 1;') -ne '1') { throw 'Tenant disposable database authentication failed.' }
    $registeredDsn = Invoke-Psql -Database $masterDatabase -Password $dbPasswordSecure -Sql "SELECT db_conn_str FROM tenants WHERE slug = $(ConvertTo-PgSqlLiteral $tenantSlug) AND is_active = true;"
    if ([string]::IsNullOrWhiteSpace($registeredDsn) -or -not (Test-DisposableDsn -Dsn $registeredDsn -ExpectedDatabase $tenantDatabase)) { throw 'Tenant registry does not bind e2e-enrichment to the required disposable tenant database.' }
    Write-Host 'LOCAL_APPLY_SAFETY_GATE=PASSED'

    $phase = 'SELECT_APPROVED_E2E_SUGGESTION'
    $organizationID = [int](Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT id FROM organizations WHERE code = $(ConvertTo-PgSqlLiteral $organizationCode) AND metadata->>'e2e_runner' = $(ConvertTo-PgSqlLiteral $fixtureMarker);").Trim()
    if ($organizationID -le 0) { throw 'The expected disposable E2E organization is missing or not owned by the E2E runner.' }
    $suggestion = Get-SelectedSuggestion
    if ($null -eq $suggestion) {
        Write-Host 'APPLY_DIAGNOSTIC_RESULT=NO_APPROVED_E2E_SUGGESTION'
        exit 0
    }
    $suggestionID = [int]$suggestion.suggestion_id
    Write-Host ("SELECTED_SUGGESTION_ID={0}" -f $suggestionID)
    Write-Host ("SELECTED_SOURCE_ITEM_CODE={0}" -f $suggestion.source_item_code)
    Write-Host 'SELECTED_STATUS=approved'

    $phase = 'DISPOSABLE_APPLY_AUTH'
    $permissionExists = Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT EXISTS (SELECT 1 FROM permissions WHERE code = 'product_enrichment:apply');"
    if ($permissionExists.Trim().ToLowerInvariant() -ne 't') { throw 'Apply permission is missing from the disposable tenant.' }
    $roleID = [int](Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT id FROM roles WHERE code = $(ConvertTo-PgSqlLiteral $roleCode) AND metadata->>'e2e_runner' = $(ConvertTo-PgSqlLiteral $fixtureMarker);").Trim()
    if ($roleID -le 0) { throw 'The existing disposable E2E reviewer role is missing or not owned by the E2E runner.' }
    Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "INSERT INTO role_permissions (role_id, permission_id, scope) SELECT $roleID, p.id, 'all' FROM permissions p WHERE p.code IN ('product_enrichment:review', 'product_enrichment:apply') ON CONFLICT (role_id, permission_id) DO NOTHING;" | Out-Null
    $randomPasswordBytes = New-Object 'System.Byte[]' 24
    $rng = [Security.Cryptography.RandomNumberGenerator]::Create()
    try { $rng.GetBytes($randomPasswordBytes) } finally { $rng.Dispose() }
    $fixturePassword = 'E2E-APPLY-' + [Convert]::ToBase64String($randomPasswordBytes)
    $randomPasswordBytes = $null
    $passwordHash = New-BcryptHash -Password $fixturePassword
    $usernameLiteral = ConvertTo-PgSqlLiteral $reviewerUsername
    $emailLiteral = ConvertTo-PgSqlLiteral $reviewerEmail
    $markerLiteral = ConvertTo-PgSqlLiteral $fixtureMarker
    $userCollision = Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT CASE WHEN EXISTS (SELECT 1 FROM users WHERE (username = $usernameLiteral OR email = $emailLiteral) AND metadata->>'e2e_runner' IS DISTINCT FROM $markerLiteral) THEN 'collision' ELSE 'ok' END;"
    if ($userCollision -ne 'ok') { throw 'Disposable reviewer username/email collides with non-E2E data.' }
    Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "INSERT INTO users (organization_id, username, email, password_hash, first_name, last_name, employee_code, is_active, metadata) VALUES ($organizationID, $usernameLiteral, $emailLiteral, $(ConvertTo-PgSqlLiteral $passwordHash), 'E2E', 'Reviewer', 'E2E-REVIEWER', true, jsonb_build_object('e2e_runner', $markerLiteral)) ON CONFLICT (username) DO UPDATE SET organization_id = EXCLUDED.organization_id, email = EXCLUDED.email, password_hash = EXCLUDED.password_hash, is_active = true, metadata = EXCLUDED.metadata WHERE users.metadata->>'e2e_runner' = $markerLiteral;" | Out-Null
    $reviewerID = [int](Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT id FROM users WHERE username = $usernameLiteral AND organization_id = $organizationID AND metadata->>'e2e_runner' = $markerLiteral;").Trim()
    if ($reviewerID -le 0) { throw 'Could not resolve the disposable E2E reviewer.' }
    Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "INSERT INTO user_roles (user_id, role_id, metadata) VALUES ($reviewerID, $roleID, jsonb_build_object('e2e_runner', $markerLiteral)) ON CONFLICT (user_id, role_id) DO NOTHING;" | Out-Null
    Assert-ApplyPermission
    Write-Host 'APPLY_PERMISSION_PRESENT=PASS'

    $phase = 'CLOUD_SERVER_START'
    $runtimeRoot = Join-Path ([IO.Path]::GetTempPath()) ("nembus-local-apply-diagnostic-" + $runID)
    New-Item -ItemType Directory -Path $runtimeRoot -Force | Out-Null
    $stdoutLog = Join-Path $runtimeRoot 'cloud-server.stdout.log'
    $stderrLog = Join-Path $runtimeRoot 'cloud-server.stderr.log'
    $cloudServerBinary = Join-Path $runtimeRoot 'cloud-server.exe'
    Write-Host ("CLOUD_STDOUT_LOG={0}" -f $stdoutLog)
    Write-Host ("CLOUD_STDERR_LOG={0}" -f $stderrLog)
    $jwtSecret = New-LocalJwtSecret
    Push-Location -LiteralPath $repoRoot
    try {
        $buildOutput = & $goExe build -o $cloudServerBinary (Join-Path $repoRoot 'apps/cloud-server/main.go') 2>&1
        if ($LASTEXITCODE -ne 0) { throw 'Unable to build the current-worktree cloud-server binary.' }
    }
    finally { $buildOutput = $null; Pop-Location }
    [Environment]::SetEnvironmentVariable('MASTER_DB_URL', $masterDsn, 'Process')
    [Environment]::SetEnvironmentVariable('JWT_SECRET', $jwtSecret, 'Process')
    [Environment]::SetEnvironmentVariable('ENV', 'development', 'Process')
    [Environment]::SetEnvironmentVariable('PORT', [string]$httpPort, 'Process')
    [Environment]::SetEnvironmentVariable('GRPC_PORT', [string]$grpcPort, 'Process')
    [Environment]::SetEnvironmentVariable('ZATCA_ENABLED', 'false', 'Process')
    [Environment]::SetEnvironmentVariable('ENRICHMENT_ENABLED', 'false', 'Process')
    [Environment]::SetEnvironmentVariable('ENRICHMENT_PROVIDER', $null, 'Process')
    [Environment]::SetEnvironmentVariable('DEEPSEEK_API_KEY', $null, 'Process')
    [Environment]::SetEnvironmentVariable('DEEPSEEK_MODEL', $null, 'Process')
    [Environment]::SetEnvironmentVariable('DEEPSEEK_BASE_URL', $null, 'Process')
    [Environment]::SetEnvironmentVariable('OPENAI_API_KEY', $null, 'Process')
    [Environment]::SetEnvironmentVariable('OPENAI_ENRICHMENT_MODEL', $null, 'Process')
    $serverProcess = Start-Process -FilePath $cloudServerBinary -WorkingDirectory $runtimeRoot -ArgumentList @('dev') -RedirectStandardOutput $stdoutLog -RedirectStandardError $stderrLog -WindowStyle Hidden -PassThru
    Wait-Health

    $phase = 'REAL_USER_LOGIN'
    $loginBody = @{ user_login = $reviewerUsername; password = $fixturePassword } | ConvertTo-Json -Compress
    $loginResponse = Get-ApiResponse -Method 'POST' -Path '/api/auth/login' -Headers @{ 'x-tenant-id' = $tenantSlug } -Body $loginBody
    if ($loginResponse.StatusCode -lt 200 -or $loginResponse.StatusCode -ge 300 -or $null -eq $loginResponse.Json -or [string]::IsNullOrWhiteSpace([string]$loginResponse.Json.data)) { throw ("Disposable reviewer login failed with HTTP {0}." -f $loginResponse.StatusCode) }
    $userToken = [string]$loginResponse.Json.data
    Write-Host 'USER_TOKEN_CREATED=PASS'
    $userHeaders = @{ Authorization = "Bearer $userToken"; 'x-tenant-id' = $tenantSlug }

    $phase = 'PRE_APPLY_SNAPSHOT'
    $preApplySnapshot = Get-ApplySnapshot -SuggestionID $suggestionID
    if ($null -eq $preApplySnapshot -or [string]$preApplySnapshot.suggestion_status -ne 'approved') { throw 'Selected suggestion is no longer approved before the one allowed apply request.' }
    Write-Host 'PRE_APPLY_SNAPSHOT=PASS'

    $logStartLines['STDOUT'] = Get-LogLineCount $stdoutLog
    $logStartLines['STDERR'] = Get-LogLineCount $stderrLog
    $phase = 'SINGLE_APPLY_REQUEST'
    $applyRequestCount = 1
    $applyResponse = Get-ApiResponse -Method 'POST' -Path ("/api/product-enrichment/suggestions/{0}/apply" -f $suggestionID) -Headers $userHeaders -Body '{}'
    Write-Host 'APPLY_REQUEST_COUNT=1'
    Write-Host ("APPLY_HTTP_STATUS={0}" -f $applyResponse.StatusCode)
    Write-Host ("APPLY_RESPONSE={0}" -f (Get-SanitizedText $applyResponse.Body))

    $phase = 'APPLY_RUNTIME_LOG_CAPTURE'
    Start-Sleep -Milliseconds 750
    $applyLogWindow = @(Get-ApplyLogWindow)
    if ($applyLogWindow.Count -gt 0) {
        Write-Host 'APPLY_LOG_WINDOW_BEGIN'
        foreach ($line in $applyLogWindow) { Write-Host $line }
        Write-Host 'APPLY_LOG_WINDOW_END'
    }
    $underlyingErrors = @($applyLogWindow | Where-Object { $_ -match '(?i)(panic|fatal|SQLSTATE|pq:|pgx:|error:|failed:|ERROR)' })
    if ($underlyingErrors.Count -gt 0) {
        $exactError = $underlyingErrors -join ' | '
        Write-Host 'APPLY_RUNTIME_ERROR_AVAILABLE=true'
        Write-Host ("APPLY_SANITIZED_ERROR={0}" -f $exactError)
        Write-Host ("APPLY_FAILURE_LAYER={0}" -f (Get-ApplyFailureLayer -ErrorText $exactError))
    }
    else {
        Write-Host 'APPLY_RUNTIME_ERROR_AVAILABLE=false'
        Write-Host 'APPLY_FAILURE_LAYER=UNKNOWN'
    }

    $phase = 'POST_APPLY_SNAPSHOT'
    $postApplySnapshot = Get-ApplySnapshot -SuggestionID $suggestionID
    if ($null -eq $postApplySnapshot) { throw 'Selected suggestion disappeared after the apply request.' }
    if ($applyResponse.StatusCode -eq 500) {
        $sameStatus = [string]$postApplySnapshot.suggestion_status -eq 'approved'
        $sameBrand = [string]$postApplySnapshot.product.brand_id -eq [string]$preApplySnapshot.product.brand_id
        $sameCategory = [string]$postApplySnapshot.product.category_id -eq [string]$preApplySnapshot.product.category_id
        $sameDescription = [string]$postApplySnapshot.product.description -eq [string]$preApplySnapshot.product.description
        $sameAuditCount = [int64]$postApplySnapshot.apply_audit_count -eq [int64]$preApplySnapshot.apply_audit_count
        Write-Host ("TRANSACTION_ROLLBACK_CONFIRMED={0}" -f (($sameStatus -and $sameBrand -and $sameCategory -and $sameDescription -and $sameAuditCount).ToString().ToLowerInvariant()))
    }
    elseif ($applyResponse.StatusCode -ge 200 -and $applyResponse.StatusCode -lt 300) {
        $changedFields = @()
        if ($null -ne $applyResponse.Json -and $null -ne $applyResponse.Json.data) { $changedFields = @($applyResponse.Json.data.changed_fields | ForEach-Object { [string]$_ }) }
        Write-Host ("POST_APPLY_STATUS={0}" -f $postApplySnapshot.suggestion_status)
        Write-Host ("POST_APPLY_CHANGED_FIELDS={0}" -f ($changedFields -join ','))
        Write-Host ("POST_APPLY_AUDIT_COUNT={0}" -f $postApplySnapshot.apply_audit_count)
    }
    else {
        Write-Host 'TRANSACTION_ROLLBACK_CONFIRMED=UNKNOWN'
    }
    Write-Host 'APPLY_DIAGNOSTIC_RESULT=COMPLETED'
}
catch {
    Write-Host ("APPLY_DIAGNOSTIC_RESULT=FAILED")
    Write-Host ("FAILURE_PHASE={0}" -f $phase)
    Write-Host ("FAILURE_REASON={0}" -f (Get-SanitizedText $_.Exception.Message))
    if ($applyRequestCount -eq 1 -and $null -eq $applyResponse) { Write-Host 'APPLY_REQUEST_COUNT=1' }
}
finally {
    Stop-CloudServer
    Restore-ProcessEnvironment @('PGPASSWORD', 'JWT_SECRET', 'MASTER_DB_URL', 'ENV', 'PORT', 'GRPC_PORT', 'ZATCA_ENABLED', 'ENRICHMENT_ENABLED', 'ENRICHMENT_PROVIDER', 'DEEPSEEK_API_KEY', 'DEEPSEEK_MODEL', 'DEEPSEEK_BASE_URL', 'OPENAI_API_KEY', 'OPENAI_ENRICHMENT_MODEL', 'OPENAI_ENRICHMENT_TIMEOUT', 'OPENAI_ENRICHMENT_WORKER_INTERVAL', 'OPENAI_ENRICHMENT_BATCH_SIZE', 'OPENAI_ENRICHMENT_MAX_RETRIES', 'NEMBUS_APPLY_DIAGNOSTIC_PASSWORD')
    $dbPasswordPlainForServer = $null
    $fixturePassword = $null
    $userToken = $null
    $jwtSecret = $null
    if ($null -ne $dbPasswordSecure) { $dbPasswordSecure.Dispose() }
    if (-not [string]::IsNullOrWhiteSpace($stdoutLog)) { Write-Host ("SERVER_LOGS_PRESERVED=true") }
}
