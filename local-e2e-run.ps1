Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# Temporary local-only full E2E runner for the disposable Nembus environment.
# This file intentionally does not contain credentials. It prompts for the
# database password and provider key at runtime, keeps them in process memory,
# and removes its temporary runtime directory during cleanup.

$repoRoot = if (-not [string]::IsNullOrWhiteSpace($PSScriptRoot)) { $PSScriptRoot } else { Split-Path -Parent $MyInvocation.MyCommand.Path }
if ([string]::IsNullOrWhiteSpace($repoRoot)) { throw 'Could not resolve the local E2E runner repository root.' }
Set-Location -LiteralPath $repoRoot

$localHost = '127.0.0.1'
$localPort = 5432
$roleName = 'nembus_e2e_user'
$masterDatabase = 'nembus_e2e_master'
$tenantDatabase = 'nembus_e2e_tenant'
$tenantSlug = 'e2e-enrichment'
$deepSeekModel = 'deepseek-v4-flash'
$deepSeekBaseUrl = 'https://api.deepseek.com'
$runID = "{0}-{1}" -f [DateTime]::UtcNow.ToString('yyyyMMddHHmmss'), [Guid]::NewGuid().ToString('N').Substring(0, 8)
$fixtureMarker = 'local-e2e-runner'
$organizationCode = 'E2E-ENRICHMENT-ORG'
$organizationName = 'Nembus Local E2E Enrichment Organization'
$categoryCode = 'E2E-ENRICHMENT-CATEGORY'
$brandCode = 'E2E-ENRICHMENT-BRAND'
$roleCode = 'E2E-ENRICHMENT-REVIEWER'
$reviewerUsername = 'e2e_enrichment_reviewer'
$reviewerEmail = 'e2e-enrichment-reviewer@nembus.invalid'
$completeSku = "E2E-COMPLETE-$runID"
$incompleteSku = "E2E-PANTENE-$runID"
$disabledSku = "E2E-DISABLED-$runID"
$completeBatchID = "E2E-$runID-COMPLETE"
$incompleteBatchID = "E2E-$runID-INCOMPLETE"
$repeatBatchID = "E2E-$runID-REPEAT"
$disabledBatchID = "E2E-$runID-DISABLED"

$report = [ordered]@{
    LOCAL_E2E_SAFETY_GATE = 'NOT_RUN'
    MASTER_DB_AUTH = 'NOT_RUN'
    TENANT_DB_AUTH = 'NOT_RUN'
    TENANT_REGISTRY_CHECK = 'NOT_RUN'
    E2E_FIXTURES_READY = 'NOT_RUN'
    M2M_TOKEN_CREATED = 'NOT_RUN'
    USER_TOKEN_CREATED = 'NOT_RUN'
    CLOUD_ENRICHMENT_ENABLED_BOOT = 'NOT_RUN'
    HEALTH_ENABLED = 'NOT_RUN'
    COMPLETE_PRODUCT_SAP_SYNC = 'NOT_RUN'
    COMPLETE_PRODUCT_SUGGESTIONS_CREATED = 'NOT_RUN'
    COMPLETE_PRODUCT_PROVIDER_CALLS = 'NOT_RUN'
    INCOMPLETE_PRODUCT_SAP_SYNC = 'NOT_RUN'
    STAGE2A_SUGGESTION_CREATED = 'NOT_RUN'
    STAGE2A_PENDING_OBSERVED = 'NOT_RUN'
    CURRENT_RUN_SUGGESTION_ISOLATION = 'NOT_RUN'
    DEEPSEEK_PROVIDER_CALL = 'NOT_RUN'
    STAGE2B_RESULT = 'NOT_RUN'
    SUGGESTION_STATUS = 'NOT_RUN'
    AI_DIRECT_PRODUCT_MUTATION = 'NOT_RUN'
    PROHIBITED_FIELD_MUTATION = 'NOT_RUN'
    REVIEW_LIST_API = 'NOT_RUN'
    REVIEW_DETAIL_API = 'NOT_RUN'
    APPROVE_API = 'NOT_RUN'
    APPROVAL_AUTO_APPLY = 'false'
    PRODUCT_CHANGED_ON_APPROVAL = 'NOT_RUN'
    APPLY_API = 'NOT_RUN'
    APPLY_AUDIT_CREATED = 'NOT_RUN'
    APPLY_CHANGED_FIELDS = 'NOT_RUN'
    SECOND_APPLY_IDEMPOTENT = 'NOT_RUN'
    REPEAT_SAP_SYNC = 'NOT_RUN'
    SAP_EMPTY_RESETS_AI_BRAND = 'NOT_RUN'
    SAP_EMPTY_RESETS_AI_DESCRIPTION = 'NOT_RUN'
    NEW_SUGGESTIONS_AFTER_REPEAT = 'NOT_RUN'
    PROVIDER_CALLS_AFTER_REPEAT = 'NOT_RUN'
    FIVE_MINUTE_SYNC_ENRICHMENT_LOOP_RISK = 'NOT_RUN'
    OPERATIONAL_ONLY_CHANGE_TEST = 'SKIPPED'
    OPERATIONAL_CHANGE_TRIGGERED_AI = 'false'
    ENRICHMENT_DISABLED_SAP_SYNC_WORKS = 'NOT_RUN'
    ENRICHMENT_DISABLED_NEW_PENDING_SUGGESTIONS = 'NOT_RUN'
    ENRICHMENT_DISABLED_PROVIDER_CALLS = 'NOT_RUN'
    DISABLED_QUEUE_GROWTH_RISK = 'NOT_RUN'
    MASTER_BUSINESS_WRITES = 'NOT_RUN'
    PROVIDER_CALL_ACCOUNTING_METHOD = 'NOT_RUN'
    FINAL_VERDICT = 'LOCAL_FULL_E2E_FAILED'
}

$phase = 'INITIALIZATION'
$dbPasswordSecure = $null
$deepSeekKeySecure = $null
$deepSeekKeyPlain = $null
$fixturePassword = $null
$m2mToken = $null
$userToken = $null
$jwtSecret = $null
$serverProcess = $null
$runtimeRoot = $null
$runtimeConfig = $null
$logDirectory = $null
$stdoutLog = $null
$stderrLog = $null
$httpPort = $null
$grpcPort = $null
$psqlExe = $null
$goExe = $null
$m2mGeneratorBinary = $null
$script:cloudServerBinary = $null
$organizationID = 0
$categoryID = 0
$brandID = 0
$roleID = 0
$reviewerID = 0
$suggestionID = 0
$incompleteProductID = 0
$baselineTenantCounts = $null
$baselineMasterCounts = $null
$productBeforeAI = $null
$productBeforeApproval = $null
$productAfterApproval = $null
$protectedBeforeApply = $null
$auditBeforeApply = 0
$trace = New-Object 'System.Collections.Generic.List[string]'
$originalEnvironment = @{}

function Set-ReportValue {
    param([Parameter(Mandatory = $true)][string]$Name, [Parameter(Mandatory = $true)][AllowEmptyString()][string]$Value)
    $report[$Name] = $Value
}

function Add-Trace {
    param([Parameter(Mandatory = $true)][string]$Text)
    [void]$trace.Add($Text)
}

function Convert-SecureStringToPlainText {
    param([Parameter(Mandatory = $true)][Security.SecureString]$SecureString)
    $bstr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($SecureString)
    try {
        return [Runtime.InteropServices.Marshal]::PtrToStringBSTR($bstr)
    }
    finally {
        [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)
    }
}

function ConvertTo-PgSqlLiteral {
    param([Parameter(Mandatory = $true)][AllowEmptyString()][string]$Value)
    return "'" + $Value.Replace("'", "''") + "'"
}

function Resolve-Tool {
    param([Parameter(Mandatory = $true)][string]$Name)
    $command = Get-Command $Name -ErrorAction SilentlyContinue
    if ($null -eq $command -or $command.CommandType -ne 'Application') {
        throw "Required local tool '$Name' was not found."
    }
    return $command.Source
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
        $exitCode = $LASTEXITCODE
        if ($exitCode -ne 0) {
            throw "psql failed for disposable database '$Database' with exit code ${exitCode}."
        }
        return (($output | ForEach-Object { $_.ToString() }) -join "`n").Trim()
    }
    finally {
        $plainPassword = $null
        Remove-Item -LiteralPath 'Env:PGPASSWORD' -ErrorAction SilentlyContinue
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
    catch {
        return $false
    }
}

function Assert-LocalSafetyGate {
    if ($localHost -ne '127.0.0.1' -or $localPort -ne 5432 -or $masterDatabase -ne 'nembus_e2e_master' -or $tenantDatabase -ne 'nembus_e2e_tenant' -or $tenantSlug -ne 'e2e-enrichment' -or $roleName -ne 'nembus_e2e_user') {
        Set-ReportValue 'LOCAL_E2E_SAFETY_GATE' 'FAILED'
        Write-Host 'LOCAL_E2E_SAFETY_GATE=FAILED' -ForegroundColor Red
        throw 'The disposable local E2E constants were changed; refusing to continue.'
    }
    Set-ReportValue 'LOCAL_E2E_SAFETY_GATE' 'PASSED'
}

function Test-Table {
    param([string]$Database, [string]$Schema, [string]$Table)
    $schemaLiteral = ConvertTo-PgSqlLiteral $Schema
    $tableLiteral = ConvertTo-PgSqlLiteral $Table
    $value = Invoke-Psql -Database $Database -Password $dbPasswordSecure -Sql "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = $schemaLiteral AND table_name = $tableLiteral);"
    return $value.Trim().ToLowerInvariant() -eq 't'
}

function Get-JsonSql {
    param([string]$Database, [string]$Sql)
    $raw = Invoke-Psql -Database $Database -Password $dbPasswordSecure -Sql $Sql
    if ([string]::IsNullOrWhiteSpace($raw)) { throw "Expected JSON query output from disposable database '$Database'." }
    return $raw | ConvertFrom-Json
}

function Get-Count {
    param([string]$Database, [string]$Sql)
    return [int64](Invoke-Psql -Database $Database -Password $dbPasswordSecure -Sql $Sql).Trim()
}

function Get-TenantCounts {
    return Get-JsonSql -Database $tenantDatabase -Sql @"
SELECT jsonb_build_object(
  'products', (SELECT count(*) FROM products),
  'suggestions', (SELECT count(*) FROM product_enrichment_suggestions),
  'audit_logs', (SELECT count(*) FROM audit_logs)
)::text;
"@
}

function Get-MasterBusinessCounts {
    return Get-JsonSql -Database $masterDatabase -Sql @"
SELECT jsonb_build_object(
  'organizations', (SELECT count(*) FROM organizations),
  'products', (SELECT count(*) FROM products),
  'suggestions', (SELECT count(*) FROM product_enrichment_suggestions),
  'audit_logs', (SELECT count(*) FROM audit_logs)
)::text;
"@
}

function Get-ProductSnapshot {
    param([string]$Sku)
    $skuLiteral = ConvertTo-PgSqlLiteral $Sku
    return Get-JsonSql -Database $tenantDatabase -Sql @"
SELECT jsonb_build_object(
  'product', to_jsonb(p),
  'barcodes', COALESCE((SELECT jsonb_agg(to_jsonb(b) ORDER BY b.id) FROM product_barcodes b WHERE b.product_id = p.id), '[]'::jsonb),
  'uom_conversions', COALESCE((SELECT jsonb_agg(to_jsonb(c) ORDER BY c.id) FROM product_uom_conversions c WHERE c.product_id = p.id), '[]'::jsonb),
  'inventory', COALESCE((SELECT jsonb_agg(to_jsonb(i) ORDER BY i.id) FROM inventory_stock i WHERE i.product_id = p.id), '[]'::jsonb),
  'prices', COALESCE((SELECT jsonb_agg(to_jsonb(pr) ORDER BY pr.id) FROM product_prices pr WHERE pr.product_id = p.id), '[]'::jsonb),
  'supplier_contracts', COALESCE((SELECT jsonb_agg(to_jsonb(pc) ORDER BY pc.id) FROM bp_price_contracts pc WHERE pc.product_id = p.id), '[]'::jsonb)
)::text
FROM products p
WHERE p.organization_id = $organizationID AND p.sku = $skuLiteral
LIMIT 1;
"@
}

function Get-ProductMutableState {
    param([string]$Sku)
    $skuLiteral = ConvertTo-PgSqlLiteral $Sku
    return Get-JsonSql -Database $tenantDatabase -Sql "SELECT jsonb_build_object('brand_id', brand_id, 'category_id', category_id, 'description', description)::text FROM products WHERE organization_id = $organizationID AND sku = $skuLiteral;"
}

function Get-SuggestionCountForSourceIdentity {
    param([string]$SourceItemCode)
    $sourceItemCodeLiteral = ConvertTo-PgSqlLiteral $SourceItemCode
    return Get-Count -Database $tenantDatabase -Sql "SELECT count(*) FROM product_enrichment_suggestions s JOIN products p ON p.id = s.product_id AND p.organization_id = s.organization_id WHERE s.organization_id = $organizationID AND p.sku = $sourceItemCodeLiteral AND s.source_item_code = $sourceItemCodeLiteral;"
}

function Get-SuggestionCountForProductIdentity {
    param([int]$ProductID, [string]$SourceItemCode)
    $sourceItemCodeLiteral = ConvertTo-PgSqlLiteral $SourceItemCode
    return Get-Count -Database $tenantDatabase -Sql "SELECT count(*) FROM product_enrichment_suggestions s JOIN products p ON p.id = s.product_id AND p.organization_id = s.organization_id WHERE s.organization_id = $organizationID AND s.product_id = $ProductID AND s.source_item_code = $sourceItemCodeLiteral AND p.sku = $sourceItemCodeLiteral;"
}

function Get-SuggestionForProductIdentity {
    param([int]$ProductID, [string]$SourceItemCode)
    $sourceItemCodeLiteral = ConvertTo-PgSqlLiteral $SourceItemCode
    $matchingSuggestionCount = Get-SuggestionCountForProductIdentity -ProductID $ProductID -SourceItemCode $SourceItemCode
    if ($matchingSuggestionCount -eq 0) { return $null }
    if ($matchingSuggestionCount -ne 1) { throw "Current-run suggestion identity is ambiguous: found ${matchingSuggestionCount} rows." }
    $raw = Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql @"
SELECT jsonb_build_object(
  'id', s.id, 'organization_id', s.organization_id, 'product_id', s.product_id, 'source_item_code', s.source_item_code, 'contract_version', s.contract_version, 'status', s.status,
  'provider', s.provider, 'model', s.model, 'model_version', s.model_version,
  'attempt_count', s.attempt_count, 'last_error_code', s.last_error_code,
  'source_data_fingerprint', s.source_data_fingerprint,
  'proposed_brand_present', s.proposed_brand IS NOT NULL,
  'proposed_category_present', s.proposed_category IS NOT NULL,
  'proposed_description_present', s.proposed_description IS NOT NULL,
  'unsupported_semantics_present', s.unsupported_semantics IS NOT NULL
)::text
FROM product_enrichment_suggestions s
JOIN products p ON p.id = s.product_id AND p.organization_id = s.organization_id
WHERE s.organization_id = $organizationID AND s.product_id = $ProductID AND s.source_item_code = $sourceItemCodeLiteral AND p.sku = $sourceItemCodeLiteral;
"@
    if ([string]::IsNullOrWhiteSpace($raw)) { return $null }
    return $raw | ConvertFrom-Json
}

function Get-SuggestionByCurrentRunIdentity {
    param([int]$SuggestionID, [int]$ProductID, [string]$SourceItemCode)
    $row = Get-SuggestionForProductIdentity -ProductID $ProductID -SourceItemCode $SourceItemCode
    if ($null -eq $row) { return $null }
    if ([int]$row.id -ne $SuggestionID -or [int]$row.organization_id -ne $organizationID -or [int]$row.product_id -ne $ProductID -or [string]$row.source_item_code -cne $SourceItemCode) {
        throw 'Current-run suggestion identity did not match its organization, product, and source item code.'
    }
    return $row
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
        if (-not [string]::IsNullOrWhiteSpace($responseBody)) {
            try { $json = $responseBody | ConvertFrom-Json } catch { $json = $null }
        }
        return [pscustomobject]@{ StatusCode = [int]$Response.StatusCode; Body = $responseBody; Json = $json }
    }

    $request = [System.Net.HttpWebRequest][System.Net.WebRequest]::Create($requestUri)
    $request.Proxy = $null
    $request.Method = $Method
    $request.Timeout = 20000
    $request.ReadWriteTimeout = 20000
    $hasBody = -not [string]::IsNullOrEmpty($Body)
    try {
        foreach ($key in $Headers.Keys) { $request.Headers.Set([string]$key, [string]$Headers[$key]) }
        if ($hasBody -and @('GET', 'HEAD') -contains $Method.ToUpperInvariant()) {
            throw "HTTP method $Method does not permit a request body."
        }
        if ($hasBody) {
            $bodyBytes = [Text.Encoding]::UTF8.GetBytes($Body)
            $request.ContentType = 'application/json'
            $request.ContentLength = $bodyBytes.Length
            $requestStream = $null
            try {
                $requestStream = $request.GetRequestStream()
                $requestStream.Write($bodyBytes, 0, $bodyBytes.Length)
            }
            finally {
                if ($null -ne $requestStream) { $requestStream.Dispose() }
            }
        }
        $response = [System.Net.HttpWebResponse]$request.GetResponse()
        try {
            return & $readResponse $response
        }
        finally {
            $response.Dispose()
        }
    }
    catch [System.Net.WebException] {
        if ($null -ne $_.Exception.Response) {
            $response = [System.Net.HttpWebResponse]$_.Exception.Response
            try {
                return & $readResponse $response
            }
            finally {
                $response.Dispose()
            }
        }
        $transportMessage = Get-SanitizedFailureMessage -Message $_.Exception.Message -Secrets @($jwtSecret, $deepSeekKeyPlain, $m2mToken, $userToken, $fixturePassword)
        throw "HTTP transport failure: $transportMessage"
    }
}

function Assert-ApiSuccess {
    param([Parameter(Mandatory = $true)]$Response, [Parameter(Mandatory = $true)][string]$Operation)
    if ($Response.StatusCode -lt 200 -or $Response.StatusCode -ge 300) {
        throw "$Operation returned HTTP $($Response.StatusCode)."
    }
}

function New-LocalJwtSecret {
    $bytes = New-Object 'System.Byte[]' 32
    $rng = [Security.Cryptography.RandomNumberGenerator]::Create()
    try { $rng.GetBytes($bytes) } finally { $rng.Dispose() }
    return [Convert]::ToBase64String($bytes)
}

function Get-M2MTokenFromGeneratorOutput {
    param([Parameter(Mandatory = $true)][object[]]$Output)
    $text = (($Output | ForEach-Object { $_.ToString() }) -join "`n")
    $matches = [regex]::Matches($text, '(?m)^JWT ACCESS TOKEN \(Bearer Token\):\r?\n(?<token>[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+)\r?$')
    if ($matches.Count -ne 1) { throw 'Repository-supported M2M generator did not emit exactly one JWT access token.' }
    $token = $matches[0].Groups['token'].Value
    $segments = $token.Split('.')
    if ($segments.Count -ne 3 -or (@($segments | Where-Object { [string]::IsNullOrWhiteSpace($_) }).Count -ne 0)) {
        throw 'Repository-supported M2M generator emitted an invalid JWT structure.'
    }
    return $token
}

function Find-FreeLocalPort {
    param([int]$PreferredPort, [int]$MaximumPort)
    for ($candidate = $PreferredPort; $candidate -le $MaximumPort; $candidate++) {
        $listener = New-Object System.Net.Sockets.TcpListener([Net.IPAddress]::Parse($localHost), $candidate)
        try {
            $listener.Start()
            $listener.Stop()
            return $candidate
        }
        catch { $listener.Stop() }
    }
    throw "No safe localhost port was available in the requested range."
}

function Get-LastCloudLogLine {
    foreach ($path in @($stderrLog, $stdoutLog)) {
        if ([string]::IsNullOrWhiteSpace($path) -or -not (Test-Path -LiteralPath $path)) { continue }
        try {
            $line = @(Get-Content -LiteralPath $path -Tail 40 -ErrorAction Stop | Where-Object { -not [string]::IsNullOrWhiteSpace($_) } | Select-Object -Last 1)
            if ($line.Count -gt 0) { return [string]$line[0] }
        }
        catch { }
    }
    return $null
}

function Get-CloudProcessTreePids {
    param([Parameter(Mandatory = $true)][int]$RootProcessId)

    $allProcessIds = New-Object 'System.Collections.Generic.List[int]'
    $pendingProcessIds = New-Object 'System.Collections.Generic.Queue[int]'
    [void]$allProcessIds.Add($RootProcessId)
    $pendingProcessIds.Enqueue($RootProcessId)

    while ($pendingProcessIds.Count -gt 0) {
        $parentProcessId = $pendingProcessIds.Dequeue()
        try {
            $children = @(Get-CimInstance -ClassName Win32_Process -Filter "ParentProcessId = $parentProcessId" -ErrorAction Stop)
            foreach ($child in $children) {
                $childProcessId = [int]$child.ProcessId
                if (-not $allProcessIds.Contains($childProcessId)) {
                    [void]$allProcessIds.Add($childProcessId)
                    $pendingProcessIds.Enqueue($childProcessId)
                }
            }
        }
        catch { }
    }

    return @($allProcessIds)
}

function Write-CloudLogTail {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [AllowEmptyString()][string]$Path
    )

    Write-Host ("CLOUD_{0}_TAIL_BEGIN" -f $Name) -ForegroundColor Red
    if (-not [string]::IsNullOrWhiteSpace($Path) -and (Test-Path -LiteralPath $Path)) {
        try {
            $lines = @(Get-Content -LiteralPath $Path -Tail 80 -ErrorAction Stop | Where-Object { -not [string]::IsNullOrWhiteSpace($_) } | Select-Object -Last 20)
            foreach ($line in $lines) {
                Write-Host (Get-SanitizedFailureMessage -Message ([string]$line) -Secrets @($jwtSecret, $deepSeekKeyPlain, $m2mToken, $userToken, $fixturePassword)) -ForegroundColor Red
            }
        }
        catch {
            Write-Host 'cloud log tail unavailable' -ForegroundColor Red
        }
    }
    else {
        Write-Host 'cloud log unavailable' -ForegroundColor Red
    }
    Write-Host ("CLOUD_{0}_TAIL_END" -f $Name) -ForegroundColor Red
}

function Write-CloudTimeoutDiagnostic {
    Write-Host ("CLOUD_REQUESTED_HTTP_PORT={0}" -f $httpPort) -ForegroundColor Red

    if ($null -eq $script:serverProcess) {
        Write-Host 'CLOUD_ROOT_PID=NONE' -ForegroundColor Red
        Write-Host 'CLOUD_CHILD_PIDS=NONE' -ForegroundColor Red
        Write-Host 'CLOUD_LISTENING_PORTS=NONE' -ForegroundColor Red
    }
    else {
        $rootProcessId = [int]$script:serverProcess.Id
        $treeProcessIds = @(Get-CloudProcessTreePids -RootProcessId $rootProcessId)
        $childProcessIds = @($treeProcessIds | Where-Object { $_ -ne $rootProcessId } | Sort-Object -Unique)
        $listeningPorts = New-Object 'System.Collections.Generic.List[int]'

        foreach ($processId in $treeProcessIds) {
            try {
                $connections = @(Get-NetTCPConnection -State Listen -OwningProcess $processId -ErrorAction Stop)
                foreach ($connection in $connections) {
                    $port = [int]$connection.LocalPort
                    if (-not $listeningPorts.Contains($port)) { [void]$listeningPorts.Add($port) }
                }
            }
            catch { }
        }

        $sortedPorts = @($listeningPorts | Sort-Object -Unique)
        Write-Host ("CLOUD_ROOT_PID={0}" -f $rootProcessId) -ForegroundColor Red
        Write-Host ("CLOUD_CHILD_PIDS={0}" -f $(if ($childProcessIds.Count -gt 0) { $childProcessIds -join ',' } else { 'NONE' })) -ForegroundColor Red
        Write-Host ("CLOUD_LISTENING_PORTS={0}" -f $(if ($sortedPorts.Count -gt 0) { $sortedPorts -join ',' } else { 'NONE' })) -ForegroundColor Red
        if ($sortedPorts.Count -gt 0) {
            Write-Host ("CLOUD_ACTUAL_LISTENING_PORTS={0}" -f ($sortedPorts -join ',')) -ForegroundColor Red
        }
    }

    Write-CloudLogTail -Name 'STDOUT' -Path $stdoutLog
    Write-CloudLogTail -Name 'STDERR' -Path $stderrLog
}

function Write-CloudStartupFailureDiagnostic {
    param(
        [AllowNull()]$LastHttpStatus = $null,
        [AllowEmptyString()][string]$LastHealthBody = ''
    )
    $exited = $false
    $exitCode = 'unknown'
    if ($null -ne $script:serverProcess) {
        try {
            $script:serverProcess.Refresh()
            $exited = $script:serverProcess.HasExited
            if ($exited) { $exitCode = [string]$script:serverProcess.ExitCode }
        }
        catch { }
    }

    Write-Host ("CLOUD_PROCESS_EXITED={0}" -f $exited.ToString().ToLowerInvariant()) -ForegroundColor Red
    if ($exited) {
        Write-Host ("CLOUD_EXIT_CODE={0}" -f $exitCode) -ForegroundColor Red
        $line = Get-LastCloudLogLine
        if ([string]::IsNullOrWhiteSpace($line)) { $line = 'cloud startup log line unavailable' }
        Write-Host ("CLOUD_STARTUP_ERROR={0}" -f (Get-SanitizedFailureMessage -Message $line -Secrets @($jwtSecret, $deepSeekKeyPlain, $m2mToken, $userToken, $fixturePassword))) -ForegroundColor Red
        return $true
    }

    if ($null -ne $LastHttpStatus) {
        Write-Host ("HEALTH_HTTP_STATUS={0}" -f $LastHttpStatus) -ForegroundColor Red
        if (-not [string]::IsNullOrWhiteSpace($LastHealthBody)) {
            Write-Host ("HEALTH_HTTP_RESPONSE={0}" -f (Get-SanitizedFailureMessage -Message $LastHealthBody -Secrets @($jwtSecret, $deepSeekKeyPlain, $m2mToken, $userToken, $fixturePassword))) -ForegroundColor Red
        }
        else {
            Write-Host 'HEALTH_RESPONSE_INVALID=true' -ForegroundColor Red
        }
        return $false
    }

    Write-Host 'HEALTH_TIMEOUT=true' -ForegroundColor Red
    Write-CloudTimeoutDiagnostic
    return $false
}

function Wait-Health {
    $lastHttpStatus = $null
    $lastHealthBody = ''
    $lastTransportException = $null
    for ($attempt = 1; $attempt -le 60; $attempt++) {
        if ($null -ne $script:serverProcess -and $script:serverProcess.HasExited) {
            [void](Write-CloudStartupFailureDiagnostic -LastHttpStatus $lastHttpStatus -LastHealthBody $lastHealthBody)
            throw 'cloud-server exited before health became ready.'
        }
        try {
            $response = Get-ApiResponse -Method 'GET' -Path '/health'
            $lastHttpStatus = [int]$response.StatusCode
            $lastHealthBody = [string]$response.Body
            if ($response.StatusCode -eq 200 -and $null -ne $response.Json -and $response.Json.status -eq 'OK') { return }
        }
        catch {
            $lastTransportException = Get-SanitizedFailureMessage -Message $_.Exception.Message -Secrets @($jwtSecret, $deepSeekKeyPlain, $m2mToken, $userToken, $fixturePassword)
        }
        Start-Sleep -Seconds 1
    }
    if ([string]::IsNullOrWhiteSpace($lastTransportException)) { $lastTransportException = 'none recorded' }
    Write-Host ("HEALTH_LAST_TRANSPORT_ERROR={0}" -f $lastTransportException) -ForegroundColor Red
    [void](Write-CloudStartupFailureDiagnostic -LastHttpStatus $lastHttpStatus -LastHealthBody $lastHealthBody)
    if ($null -ne $lastHttpStatus) { throw "cloud-server health endpoint returned HTTP ${lastHttpStatus}." }
    throw "cloud-server health timed out on localhost port ${httpPort}."
}

function Stop-CloudServer {
    if ($null -ne $script:serverProcess) {
        try {
            if (-not $script:serverProcess.HasExited) {
                & taskkill.exe /PID $script:serverProcess.Id /T /F 2>$null | Out-Null
            }
        }
        catch { }
        $script:serverProcess = $null
    }
}

function Save-ProcessEnvironment {
    param([string[]]$Names)
    foreach ($name in $Names) {
        $value = [Environment]::GetEnvironmentVariable($name, 'Process')
        if ($null -ne $value) { $originalEnvironment[$name] = $value }
    }
}

function Restore-ProcessEnvironment {
    param([string[]]$Names)
    foreach ($name in $Names) {
        if ($originalEnvironment.ContainsKey($name)) { [Environment]::SetEnvironmentVariable($name, $originalEnvironment[$name], 'Process') }
        else { [Environment]::SetEnvironmentVariable($name, $null, 'Process') }
    }
}

function Start-CloudServer {
    param([Parameter(Mandatory = $true)][bool]$EnrichmentEnabled)
    Stop-CloudServer
    if ($EnrichmentEnabled) {
        Set-ReportValue 'CLOUD_ENRICHMENT_ENABLED_BOOT' 'true'
    }
    else {
        Set-ReportValue 'CLOUD_ENRICHMENT_ENABLED_BOOT' 'false'
    }
    $serverDsn = New-DatabaseUrl -Database $masterDatabase -Password $dbPasswordPlainForServer
    $envNames = @('MASTER_DB_URL', 'JWT_SECRET', 'ENV', 'PORT', 'GRPC_PORT', 'ZATCA_ENABLED', 'ENRICHMENT_ENABLED', 'ENRICHMENT_PROVIDER', 'DEEPSEEK_API_KEY', 'DEEPSEEK_MODEL', 'DEEPSEEK_BASE_URL', 'OPENAI_ENRICHMENT_TIMEOUT', 'OPENAI_ENRICHMENT_WORKER_INTERVAL', 'OPENAI_ENRICHMENT_BATCH_SIZE', 'OPENAI_ENRICHMENT_MAX_RETRIES')
    Save-ProcessEnvironment $envNames
    try {
        if ([string]::IsNullOrWhiteSpace($script:cloudServerBinary) -or -not (Test-Path -LiteralPath $script:cloudServerBinary)) {
            $script:cloudServerBinary = Join-Path $runtimeRoot 'cloud-server.exe'
            Push-Location -LiteralPath $repoRoot
            try {
                $buildOutput = & $goExe build -o $script:cloudServerBinary (Join-Path $repoRoot 'apps/cloud-server/main.go') 2>&1
                if ($LASTEXITCODE -ne 0) { throw 'Unable to build the repository cloud-server binary for the temporary local E2E runtime.' }
            }
            finally {
                $buildOutput = $null
                Pop-Location
            }
        }
        [Environment]::SetEnvironmentVariable('MASTER_DB_URL', $serverDsn, 'Process')
        [Environment]::SetEnvironmentVariable('JWT_SECRET', $jwtSecret, 'Process')
        [Environment]::SetEnvironmentVariable('ENV', 'development', 'Process')
        [Environment]::SetEnvironmentVariable('PORT', [string]$httpPort, 'Process')
        [Environment]::SetEnvironmentVariable('GRPC_PORT', [string]$grpcPort, 'Process')
        [Environment]::SetEnvironmentVariable('ZATCA_ENABLED', 'false', 'Process')
        [Environment]::SetEnvironmentVariable('ENRICHMENT_ENABLED', ($(if ($EnrichmentEnabled) { 'true' } else { 'false' })), 'Process')
        [Environment]::SetEnvironmentVariable('ENRICHMENT_PROVIDER', 'deepseek', 'Process')
        [Environment]::SetEnvironmentVariable('DEEPSEEK_MODEL', $deepSeekModel, 'Process')
        [Environment]::SetEnvironmentVariable('DEEPSEEK_BASE_URL', $deepSeekBaseUrl, 'Process')
        [Environment]::SetEnvironmentVariable('OPENAI_ENRICHMENT_TIMEOUT', '60s', 'Process')
        [Environment]::SetEnvironmentVariable('OPENAI_ENRICHMENT_WORKER_INTERVAL', '5s', 'Process')
        [Environment]::SetEnvironmentVariable('OPENAI_ENRICHMENT_BATCH_SIZE', '5', 'Process')
        [Environment]::SetEnvironmentVariable('OPENAI_ENRICHMENT_MAX_RETRIES', '1', 'Process')
        if ($EnrichmentEnabled) { [Environment]::SetEnvironmentVariable('DEEPSEEK_API_KEY', $deepSeekKeyPlain, 'Process') }
        else { [Environment]::SetEnvironmentVariable('DEEPSEEK_API_KEY', $null, 'Process') }
        $stdoutLog = Join-Path $logDirectory ("cloud-${httpPort}.stdout.log")
        $stderrLog = Join-Path $logDirectory ("cloud-${httpPort}.stderr.log")
        $script:serverProcess = Start-Process -FilePath $script:cloudServerBinary -WorkingDirectory $runtimeRoot -ArgumentList @('dev') -RedirectStandardOutput $stdoutLog -RedirectStandardError $stderrLog -WindowStyle Hidden -PassThru
    }
    finally {
        Restore-ProcessEnvironment $envNames
        $serverDsn = $null
    }
    Wait-Health
    if ([string]$report.HEALTH_ENABLED -eq 'NOT_RUN') { Set-ReportValue 'HEALTH_ENABLED' 'PASS' }
}

function New-BcryptHash {
    param([Parameter(Mandatory = $true)][string]$Password)
    $helper = Join-Path $runtimeRoot 'bcrypt-helper.go'
    $source = @'
package main

import (
    "fmt"
    "os"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    value := os.Getenv("NEMBUS_E2E_FIXTURE_PASSWORD")
    if value == "" { panic("fixture password missing") }
    hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
    if err != nil { panic(err) }
    fmt.Print(string(hash))
}
'@
    [IO.File]::WriteAllText($helper, $source, (New-Object Text.UTF8Encoding($false)))
    $old = [Environment]::GetEnvironmentVariable('NEMBUS_E2E_FIXTURE_PASSWORD', 'Process')
    try {
        [Environment]::SetEnvironmentVariable('NEMBUS_E2E_FIXTURE_PASSWORD', $Password, 'Process')
        $hashOutput = & $goExe run $helper 2>$null
        if ($LASTEXITCODE -ne 0) { throw 'Unable to generate the E2E bcrypt fixture hash using the repository Go toolchain.' }
        return (($hashOutput | ForEach-Object { $_.ToString() }) -join '').Trim()
    }
    finally {
        if ($null -ne $old) { [Environment]::SetEnvironmentVariable('NEMBUS_E2E_FIXTURE_PASSWORD', $old, 'Process') }
        else { [Environment]::SetEnvironmentVariable('NEMBUS_E2E_FIXTURE_PASSWORD', $null, 'Process') }
        Remove-Item -LiteralPath $helper -Force -ErrorAction SilentlyContinue
    }
}

function New-SapPayload {
    param([string]$Sku, [string]$Name, [string]$Description, [string]$Brand, [string]$BatchID)
    return [ordered]@{
        batch_id = $BatchID
        run_id = "E2E-RUN-$runID"
        organization_id = $organizationID
        domain = 'products'
        sequence_number = 1
        is_last_batch = $true
        timestamp = [DateTime]::UtcNow.ToString('o')
        products = @([ordered]@{
            sku = $Sku
            name = $Name
            description = $Description
            category_code = $categoryCode
            brand_code = $Brand
            uom_code = 'UNIT'
            base_uom_code = 'UNIT'
            sales_uom_code = 'UNIT'
            purchase_uom_code = 'UNIT'
            uom_group_code = ''
            sales_qty_per_base = 1.0
            purchase_qty_per_base = 1.0
            product_type = 'standard'
            is_serialized = $false
            is_batch_managed = $false
            is_active = $true
            is_sellable = $true
            is_purchasable = $true
            track_inventory = $true
            primary_barcode = ''
            uom_conversions = @()
            metadata = [ordered]@{ e2e_runner = $fixtureMarker; e2e_run_id = $runID; source = 'synthetic-local-e2e' }
        })
    }
}

function Invoke-SapMigration {
    param([System.Collections.IDictionary]$Payload)
    $body = $Payload | ConvertTo-Json -Depth 30 -Compress
    $headers = @{ Authorization = "Bearer $m2mToken"; 'x-tenant-id' = $tenantSlug; 'x-organization-id' = [string]$organizationID }
    $response = Get-ApiResponse -Method 'POST' -Path '/api/v1/migration/batch' -Headers $headers -Body $body
    Assert-ApiSuccess $response 'SAP migration batch'
    if ($null -eq $response.Json -or $response.Json.success -ne $true -or $response.Json.records_failed -ne 0) { throw 'SAP migration returned an unsuccessful batch response.' }
    return $response.Json
}

function Wait-Stage2B {
    param([int]$SuggestionID, [int]$ProductID, [string]$SourceItemCode)
    for ($attempt = 1; $attempt -le 36; $attempt++) {
        $row = Get-SuggestionByCurrentRunIdentity -SuggestionID $SuggestionID -ProductID $ProductID -SourceItemCode $SourceItemCode
        if ($null -eq $row) { throw 'Current-run Stage 2A suggestion disappeared before worker completion.' }
        $status = [string]$row.status
        switch ($status) {
            'in_review' { return $row }
            'pending' { Add-Trace 'Stage 2B waiting: current-run suggestion remains pending' }
            'processing' { Add-Trace 'Stage 2B waiting: current-run suggestion is processing' }
            'retryable' {
                Set-ReportValue 'SUGGESTION_STATUS' 'retryable'
                Add-Trace "Stage 2B waiting: current-run suggestion is retryable (attempt_count=$($row.attempt_count), error_code=$($row.last_error_code)); following finite retry policy"
            }
            'failed' {
                Set-ReportValue 'SUGGESTION_STATUS' 'failed'
                Set-ReportValue 'STAGE2B_RESULT' 'FAILED'
                Add-Trace "Stage 2B failed for the current-run suggestion (error_code=$($row.last_error_code))"
                throw "Stage 2B ended in status 'failed' with sanitized error code '$($row.last_error_code)'."
            }
            'approved' { throw 'Current-run isolation failed: a new suggestion was already approved before review.' }
            'applied' { throw 'Current-run isolation failed: a new suggestion was already applied before review.' }
            'rejected' { throw 'Current-run isolation failed: a new suggestion was already rejected before review.' }
            default { throw "Current-run suggestion entered unexpected status '$status'." }
        }
        Start-Sleep -Seconds 2
    }
    Set-ReportValue 'STAGE2B_RESULT' 'FAILED'
    throw 'Timed out waiting for the tenant-aware enrichment worker to complete Stage 2B under the finite retry policy.'
}

function Assert-DetailContract {
    param([Parameter(Mandatory = $true)]$Detail)
    $required = @('source_identity', 'inference_snapshot', 'current_authoritative_state', 'provider_context', 'review_state', 'safety')
    foreach ($name in $required) {
        if ($null -eq $Detail.PSObject.Properties[$name]) { throw "Review detail is missing '$name'." }
    }
    $state = $Detail.review_state
    if ($null -eq $state -or [string]$state.status -ne 'in_review') { throw 'Review detail did not report in_review state.' }
    $responseText = $Detail | ConvertTo-Json -Depth 30 -Compress
    if ($responseText.Contains($deepSeekKeyPlain) -or $responseText.Contains($m2mToken) -or $responseText.Contains($userToken)) { throw 'Review response contained a sensitive credential.' }
}

function Get-SanitizedFailureMessage {
    param(
        [Parameter(Mandatory = $true)][string]$Message,
        [string[]]$Secrets = @()
    )
    $sanitized = $Message -replace '(?i)Bearer\s+\S+', 'Bearer [redacted]'
    $sanitized = $sanitized -replace '(?i)postgres(?:ql)?://\S+', 'postgres://[redacted]'
    $sanitized = $sanitized -replace '(?i)\b(master[_ -]?db[_ -]?(?:url|dsn)|db[_ -]?conn[_ -]?str|tenant[_ -]?db[_ -]?(?:url|dsn))\s*[:=]\s*("[^"]*"|''[^'']*''|\S+)', '$1=[redacted]'
    $sanitized = $sanitized -replace '(?i)\b(password|passwd|pwd|jwt[_ -]?secret|api[_ -]?key|authorization)\s*[:=]\s*("[^"]*"|''[^'']*''|\S+)', '$1=[redacted]'
    $sanitized = $sanitized -replace '(?i)("(?:master[_ -]?db[_ -]?(?:url|dsn)|db[_ -]?conn[_ -]?str|tenant[_ -]?db[_ -]?(?:url|dsn)|password|passwd|pwd|jwt[_ -]?secret|api[_ -]?key|authorization)"\s*:\s*)"[^"]*"', '$1"[redacted]"'
    $sanitized = $sanitized -replace '(?<![A-Za-z0-9_-])[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}(?![A-Za-z0-9_-])', '[redacted-jwt]'
    foreach ($secret in $Secrets) {
        if (-not [string]::IsNullOrWhiteSpace($secret)) { $sanitized = $sanitized.Replace($secret, '[redacted]') }
    }
    if ($sanitized.Length -gt 300) { $sanitized = $sanitized.Substring(0, 300) }
    return $sanitized
}

try {
    Save-ProcessEnvironment @('PGPASSWORD', 'JWT_SECRET', 'MASTER_DB_URL', 'DEEPSEEK_API_KEY', 'NEMBUS_E2E_FIXTURE_PASSWORD')
    $phase = 'LOCAL_SAFETY_GATE'
    Assert-LocalSafetyGate
    $psqlExe = Resolve-Tool 'psql'
    $goExe = Resolve-Tool 'go'

    $dbPasswordSecure = Read-Host 'nembus_e2e_user PostgreSQL password' -AsSecureString
    $deepSeekKeySecure = Read-Host 'DeepSeek API key' -AsSecureString
    if ($null -eq $dbPasswordSecure -or $null -eq $deepSeekKeySecure) { throw 'Both secure prompts are required.' }

    $phase = 'DATABASE_READINESS'
    $masterDsnExpected = $null
    $tenantDsnExpected = $null
    $plainForRegistry = $null
    try {
        $plainForRegistry = Convert-SecureStringToPlainText $dbPasswordSecure
        $masterDsnExpected = New-DatabaseUrl -Database $masterDatabase -Password $plainForRegistry
        $tenantDsnExpected = New-DatabaseUrl -Database $tenantDatabase -Password $plainForRegistry
    }
    finally { $plainForRegistry = $null }
    if (-not (Test-DisposableDsn -Dsn $masterDsnExpected -ExpectedDatabase $masterDatabase) -or -not (Test-DisposableDsn -Dsn $tenantDsnExpected -ExpectedDatabase $tenantDatabase)) {
        Write-Host 'LOCAL_E2E_SAFETY_GATE=FAILED' -ForegroundColor Red
        throw 'Constructed database DSN failed the exact disposable loopback safety gate.'
    }
    if ((Invoke-Psql -Database $masterDatabase -Password $dbPasswordSecure -Sql 'SELECT 1;') -ne '1') { throw 'Master disposable database authentication failed.' }
    Set-ReportValue 'MASTER_DB_AUTH' 'PASSED'
    if ((Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql 'SELECT 1;') -ne '1') { throw 'Tenant disposable database authentication failed.' }
    Set-ReportValue 'TENANT_DB_AUTH' 'PASSED'
    $registryTenantLiteral = ConvertTo-PgSqlLiteral $tenantSlug
    $registeredDsn = Invoke-Psql -Database $masterDatabase -Password $dbPasswordSecure -Sql "SELECT db_conn_str FROM tenants WHERE slug = $registryTenantLiteral AND is_active = true;"
    if ([string]::IsNullOrWhiteSpace($registeredDsn) -or -not (Test-DisposableDsn -Dsn $registeredDsn -ExpectedDatabase $tenantDatabase)) {
        Write-Host 'LOCAL_E2E_SAFETY_GATE=FAILED' -ForegroundColor Red
        throw 'Tenant registry did not resolve e2e-enrichment to the expected local disposable tenant DSN.'
    }
    Set-ReportValue 'TENANT_REGISTRY_CHECK' 'PASSED'
    $requiredTables = @(
        @{ Schema = 'public'; Table = 'organizations' }, @{ Schema = 'public'; Table = 'users' },
        @{ Schema = 'public'; Table = 'roles' }, @{ Schema = 'public'; Table = 'permissions' },
        @{ Schema = 'public'; Table = 'products' }, @{ Schema = 'public'; Table = 'brands' },
        @{ Schema = 'public'; Table = 'product_categories' }, @{ Schema = 'public'; Table = 'product_enrichment_suggestions' },
        @{ Schema = 'public'; Table = 'audit_logs' }, @{ Schema = 'staging'; Table = 'sap_migration_batches' }
    )
    foreach ($requiredTable in $requiredTables) {
        if (-not (Test-Table -Database $tenantDatabase -Schema $requiredTable.Schema -Table $requiredTable.Table)) { throw "Required tenant table '$($requiredTable.Schema).$($requiredTable.Table)' is missing." }
    }

    $phase = 'FIXTURE_CREATION'
    $runtimeRoot = Join-Path ([IO.Path]::GetTempPath()) ("nembus-local-e2e-$runID")
    New-Item -ItemType Directory -Path $runtimeRoot -Force | Out-Null
    $baselineMasterCounts = Get-MasterBusinessCounts
    $baselineTenantCounts = Get-TenantCounts
    $organizationCodeLiteral = ConvertTo-PgSqlLiteral $organizationCode
    $organizationNameLiteral = ConvertTo-PgSqlLiteral $organizationName
    $markerLiteral = ConvertTo-PgSqlLiteral $fixtureMarker
    $orgState = Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT CASE WHEN NOT EXISTS (SELECT 1 FROM organizations WHERE code = $organizationCodeLiteral) THEN 'absent' WHEN EXISTS (SELECT 1 FROM organizations WHERE code = $organizationCodeLiteral AND metadata->>'e2e_runner' = $markerLiteral) THEN 'owned' ELSE 'collision' END;"
    if ($orgState -eq 'collision') { throw 'E2E organization code exists without the local E2E ownership marker.' }
    if ($orgState -eq 'absent') {
        Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "INSERT INTO organizations (name, code, legal_name, currency_code, is_active, metadata) VALUES ($organizationNameLiteral, $organizationCodeLiteral, $organizationNameLiteral, 'USD', true, jsonb_build_object('e2e_runner', $markerLiteral));" | Out-Null
    }
    Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "UPDATE organizations SET name = $organizationNameLiteral, is_active = true, metadata = jsonb_build_object('e2e_runner', $markerLiteral) WHERE code = $organizationCodeLiteral AND metadata->>'e2e_runner' = $markerLiteral;" | Out-Null
    $organizationID = [int](Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT id FROM organizations WHERE code = $organizationCodeLiteral AND metadata->>'e2e_runner' = $markerLiteral;").Trim()
    if ($organizationID -le 0) { throw 'Could not resolve the owned E2E organization ID.' }

    foreach ($taxonomy in @(@{ Code = $categoryCode; Name = 'E2E Hair Care Category'; Table = 'product_categories' }, @{ Code = $brandCode; Name = 'E2E Pantene-Like Brand'; Table = 'brands' })) {
        $codeLiteral = ConvertTo-PgSqlLiteral $taxonomy.Code
        $nameLiteral = ConvertTo-PgSqlLiteral $taxonomy.Name
        $state = Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT CASE WHEN NOT EXISTS (SELECT 1 FROM $($taxonomy.Table) WHERE code = $codeLiteral) THEN 'absent' WHEN EXISTS (SELECT 1 FROM $($taxonomy.Table) WHERE code = $codeLiteral AND metadata->>'e2e_runner' = $markerLiteral) THEN 'owned' ELSE 'collision' END;"
        if ($state -eq 'collision') { throw "E2E taxonomy code '$($taxonomy.Code)' exists without the local E2E ownership marker." }
        if ($state -eq 'absent') {
            if ($taxonomy.Table -eq 'product_categories') { Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "INSERT INTO product_categories (name, code, category_level, is_active, metadata) VALUES ($nameLiteral, $codeLiteral, 1, true, jsonb_build_object('e2e_runner', $markerLiteral));" | Out-Null }
            else { Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "INSERT INTO brands (name, code, is_active, metadata) VALUES ($nameLiteral, $codeLiteral, true, jsonb_build_object('e2e_runner', $markerLiteral));" | Out-Null }
        }
        if ($taxonomy.Table -eq 'product_categories') { $categoryID = [int](Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT id FROM product_categories WHERE code = $codeLiteral AND metadata->>'e2e_runner' = $markerLiteral;").Trim() }
        else { $brandID = [int](Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT id FROM brands WHERE code = $codeLiteral AND metadata->>'e2e_runner' = $markerLiteral;").Trim() }
    }
    if ((Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT EXISTS (SELECT 1 FROM permissions WHERE code = 'product_enrichment:review');").Trim().ToLowerInvariant() -ne 't') { throw 'Review permission is missing.' }
    if ((Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT EXISTS (SELECT 1 FROM permissions WHERE code = 'product_enrichment:apply');").Trim().ToLowerInvariant() -ne 't') { throw 'Apply permission is missing.' }
    $roleCodeLiteral = ConvertTo-PgSqlLiteral $roleCode
    $roleNameLiteral = ConvertTo-PgSqlLiteral 'E2E Enrichment Reviewer'
    $roleState = Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT CASE WHEN NOT EXISTS (SELECT 1 FROM roles WHERE code = $roleCodeLiteral) THEN 'absent' WHEN EXISTS (SELECT 1 FROM roles WHERE code = $roleCodeLiteral AND metadata->>'e2e_runner' = $markerLiteral) THEN 'owned' ELSE 'collision' END;"
    if ($roleState -eq 'collision') { throw 'E2E role code exists without the local E2E ownership marker.' }
    if ($roleState -eq 'absent') { Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "INSERT INTO roles (name, code, description, is_system_role, is_active, metadata) VALUES ($roleNameLiteral, $roleCodeLiteral, 'Temporary local E2E review/apply role', false, true, jsonb_build_object('e2e_runner', $markerLiteral));" | Out-Null }
    $roleID = [int](Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT id FROM roles WHERE code = $roleCodeLiteral AND metadata->>'e2e_runner' = $markerLiteral;").Trim()
    if ($roleID -le 0) { throw 'Could not resolve the E2E role ID.' }
    Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "INSERT INTO role_permissions (role_id, permission_id, scope) SELECT $roleID, p.id, 'all' FROM permissions p WHERE p.code IN ('product_enrichment:review', 'product_enrichment:apply') ON CONFLICT (role_id, permission_id) DO NOTHING;" | Out-Null

    $randomPasswordBytes = New-Object 'System.Byte[]' 24
    $randomPasswordRng = [Security.Cryptography.RandomNumberGenerator]::Create()
    try { $randomPasswordRng.GetBytes($randomPasswordBytes) } finally { $randomPasswordRng.Dispose() }
    $fixturePassword = [Convert]::ToBase64String($randomPasswordBytes)
    $randomPasswordBytes = $null
    $fixturePassword = "E2E-$runID-$fixturePassword"
    $passwordHash = New-BcryptHash $fixturePassword
    $usernameLiteral = ConvertTo-PgSqlLiteral $reviewerUsername
    $emailLiteral = ConvertTo-PgSqlLiteral $reviewerEmail
    $userCollision = Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT CASE WHEN EXISTS (SELECT 1 FROM users WHERE (username = $usernameLiteral OR email = $emailLiteral) AND metadata->>'e2e_runner' IS DISTINCT FROM $markerLiteral) THEN 'collision' ELSE 'ok' END;"
    if ($userCollision -ne 'ok') { throw 'E2E reviewer username/email exists without the local E2E ownership marker.' }
    $passwordHashLiteral = ConvertTo-PgSqlLiteral $passwordHash
    Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "INSERT INTO users (organization_id, username, email, password_hash, first_name, last_name, employee_code, is_active, metadata) VALUES ($organizationID, $usernameLiteral, $emailLiteral, $passwordHashLiteral, 'E2E', 'Reviewer', 'E2E-REVIEWER', true, jsonb_build_object('e2e_runner', $markerLiteral)) ON CONFLICT (username) DO UPDATE SET organization_id = EXCLUDED.organization_id, email = EXCLUDED.email, password_hash = EXCLUDED.password_hash, is_active = true, metadata = EXCLUDED.metadata WHERE users.metadata->>'e2e_runner' = $markerLiteral;" | Out-Null
    $reviewerID = [int](Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "SELECT id FROM users WHERE username = $usernameLiteral AND organization_id = $organizationID AND metadata->>'e2e_runner' = $markerLiteral;").Trim()
    if ($reviewerID -le 0) { throw 'Could not resolve the E2E reviewer ID.' }
    Invoke-Psql -Database $tenantDatabase -Password $dbPasswordSecure -Sql "INSERT INTO user_roles (user_id, role_id, metadata) VALUES ($reviewerID, $roleID, jsonb_build_object('e2e_runner', $markerLiteral)) ON CONFLICT (user_id, role_id) DO NOTHING;" | Out-Null
    Set-ReportValue 'E2E_FIXTURES_READY' 'PASSED'

    $phase = 'AUTHENTICATION_AND_SERVER_BOOT'
    $logDirectory = Join-Path $runtimeRoot 'logs'
    New-Item -ItemType Directory -Path (Join-Path $runtimeRoot 'config') -Force | Out-Null
    New-Item -ItemType Directory -Path $logDirectory -Force | Out-Null
    $runtimeConfig = Join-Path $runtimeRoot 'config/m2m_clients.json'
    $httpPort = Find-FreeLocalPort 18080 18120
    $grpcPort = Find-FreeLocalPort 19051 19090
    $jwtSecret = New-LocalJwtSecret
    $m2mGenerator = Join-Path $repoRoot 'apps/cloud-server/cmd/m2m-gen/main.go'
    $m2mGeneratorBinary = Join-Path $runtimeRoot 'm2m-gen.exe'
    $m2mClientID = "local-e2e-$runID"
    $m2mEnvNames = @('JWT_SECRET')
    Save-ProcessEnvironment $m2mEnvNames
    try {
        [Environment]::SetEnvironmentVariable('JWT_SECRET', $jwtSecret, 'Process')
        Push-Location -LiteralPath $repoRoot
        try {
            $buildOutput = & $goExe build -o $m2mGeneratorBinary $m2mGenerator 2>&1
            if ($LASTEXITCODE -ne 0) { throw 'Unable to build the repository M2M generator for the temporary local E2E runtime.' }
        }
        finally {
            $buildOutput = $null
            Pop-Location
        }
        Push-Location -LiteralPath $runtimeRoot
        try {
            $generatorOutput = & $m2mGeneratorBinary -client-id $m2mClientID -client-name 'Nembus Local E2E SAP Agent' -tenant-slug $tenantSlug -organization-id $organizationID -scopes 'sap:migration' -years 1 2>&1
            if ($LASTEXITCODE -ne 0) { throw 'Repository-supported M2M generator failed for the temporary local E2E runtime.' }
        }
        finally {
            Pop-Location
        }
        if (-not (Test-Path -LiteralPath $runtimeConfig)) { throw 'Repository-supported M2M generator did not create the temporary registry.' }
        $m2mToken = Get-M2MTokenFromGeneratorOutput $generatorOutput
    }
    finally {
        $generatorOutput = $null
        Restore-ProcessEnvironment $m2mEnvNames
    }
    $registry = Get-Content -Raw -LiteralPath $runtimeConfig | ConvertFrom-Json
    $m2mClient = @($registry.clients | Where-Object { $_.client_id -eq $m2mClientID })[0]
    if ($null -eq $m2mClient -or [string]::IsNullOrWhiteSpace($m2mClient.token) -or ([string]$m2mClient.token) -cne $m2mToken) { throw 'Temporary M2M registry did not contain the generated client token.' }
    if ([string]$m2mClient.tenant_slug -cne $tenantSlug -or [int]$m2mClient.organization_id -ne $organizationID -or $m2mClient.is_active -ne $true -or -not (@($m2mClient.scopes) -contains 'sap:migration')) {
        throw 'Temporary M2M registry did not preserve the required tenant, organization, active, and migration-scope binding.'
    }
    Set-ReportValue 'M2M_TOKEN_CREATED' 'PASSED'
    $dbPasswordPlainForServer = Convert-SecureStringToPlainText $dbPasswordSecure
    $deepSeekKeyPlain = Convert-SecureStringToPlainText $deepSeekKeySecure
    Start-CloudServer $true
    Add-Trace '1 SAP request path and tenant-aware cloud-server boot are ready'

    $phase = 'BASELINE_AND_COMPLETE_PRODUCT'
    $beforeCompleteSuggestionCount = Get-SuggestionCountForSourceIdentity $completeSku
    $completePayload = New-SapPayload $completeSku 'E2E Complete Shampoo' 'Complete E2E description' $brandCode $completeBatchID
    [void](Invoke-SapMigration $completePayload)
    $completeProduct = Get-ProductSnapshot $completeSku
    if ($null -eq $completeProduct) { throw 'Complete E2E product did not persist in the tenant database.' }
    $completeSuggestionCount = Get-SuggestionCountForSourceIdentity $completeSku
    if ($completeSuggestionCount -ne $beforeCompleteSuggestionCount) { throw 'Complete product unexpectedly created an enrichment suggestion.' }
    Set-ReportValue 'COMPLETE_PRODUCT_SAP_SYNC' 'PASS'
    Set-ReportValue 'COMPLETE_PRODUCT_SUGGESTIONS_CREATED' ([string]($completeSuggestionCount - $beforeCompleteSuggestionCount))
    Set-ReportValue 'COMPLETE_PRODUCT_PROVIDER_CALLS' '0'
    Add-Trace '2 SAP request accepted; 3 deterministic complete product persisted; 4 Stage 2A found no gap'

    $phase = 'INCOMPLETE_PRODUCT_AND_STAGE2B'
    $beforeIncompleteSuggestionCount = Get-SuggestionCountForSourceIdentity $incompleteSku
    if ($beforeIncompleteSuggestionCount -ne 0) { throw 'Current-run incomplete source identity already has a suggestion before migration.' }
    $incompletePayload = New-SapPayload $incompleteSku 'Pantene shampoo smooth care 400 ml' '' '' $incompleteBatchID
    [void](Invoke-SapMigration $incompletePayload)
    $productBeforeAI = Get-ProductSnapshot $incompleteSku
    if ($null -eq $productBeforeAI) { throw 'Incomplete E2E product did not persist in the tenant database.' }
    $incompleteProductID = [int]$productBeforeAI.product.id
    if ($incompleteProductID -le 0 -or [string]$productBeforeAI.product.sku -cne $incompleteSku -or [int]$productBeforeAI.product.organization_id -ne $organizationID) { throw 'Incomplete E2E product did not preserve the current-run organization and source identity.' }
    $currentRunSuggestionCount = Get-SuggestionCountForProductIdentity -ProductID $incompleteProductID -SourceItemCode $incompleteSku
    if ($currentRunSuggestionCount -ne ($beforeIncompleteSuggestionCount + 1)) { throw "Expected exactly one new current-run Stage 2A suggestion, found $($currentRunSuggestionCount - $beforeIncompleteSuggestionCount)." }
    $currentRunSuggestion = Get-SuggestionForProductIdentity -ProductID $incompleteProductID -SourceItemCode $incompleteSku
    if ($null -eq $currentRunSuggestion) { throw 'Stage 2A did not create a suggestion for the current-run product identity.' }
    $suggestionID = [int]$currentRunSuggestion.id
    $firstObservedStatus = [string]$currentRunSuggestion.status
    if ($firstObservedStatus -notin @('pending', 'processing', 'in_review', 'retryable', 'failed')) { throw "Current-run isolation failed: new suggestion first appeared in unexpected status '$firstObservedStatus'." }
    Set-ReportValue 'INCOMPLETE_PRODUCT_SAP_SYNC' 'PASS'
    Set-ReportValue 'STAGE2A_SUGGESTION_CREATED' 'PASS'
    Set-ReportValue 'STAGE2A_PENDING_OBSERVED' ($(if ($firstObservedStatus -eq 'pending') { 'true' } else { 'false' }))
    Set-ReportValue 'CURRENT_RUN_SUGGESTION_ISOLATION' 'true'
    Add-Trace "5 Stage 2A eligibility checked; 6 current-run suggestion exists (first observed status=$firstObservedStatus)"
    $stage2B = Wait-Stage2B -SuggestionID $suggestionID -ProductID $incompleteProductID -SourceItemCode $incompleteSku
    if ([string]$stage2B.provider -ne 'deepseek' -or [string]$stage2B.model -ne $deepSeekModel -or [int]$stage2B.attempt_count -lt 1) { throw 'Stage 2B provider metadata or attempt accounting did not match the configured DeepSeek cycle.' }
    if (-not $stage2B.proposed_brand_present -or -not $stage2B.proposed_category_present -or -not $stage2B.proposed_description_present) { throw 'Stage 2B did not persist the strict proposal fields.' }
    Set-ReportValue 'DEEPSEEK_PROVIDER_CALL' 'PASS'
    Set-ReportValue 'STAGE2B_RESULT' 'PASS'
    Set-ReportValue 'SUGGESTION_STATUS' 'in_review'
    Add-Trace '7 F2 claimed the suggestion; 8 DeepSeek called; 9 Stage 2B strict output entered in_review'
    $productAfterAI = Get-ProductSnapshot $incompleteSku
    $aiMutation = (($productBeforeAI | ConvertTo-Json -Depth 30 -Compress) -ne ($productAfterAI | ConvertTo-Json -Depth 30 -Compress))
    Set-ReportValue 'AI_DIRECT_PRODUCT_MUTATION' ($(if ($aiMutation) { 'true' } else { 'false' }))
    Set-ReportValue 'PROHIBITED_FIELD_MUTATION' ($(if ($aiMutation) { 'true' } else { 'false' }))
    if ($aiMutation) { throw 'Worker processing changed the product before explicit application.' }

    $phase = 'REVIEW_API'
    $userLoginBody = @{ user_login = $reviewerUsername; password = $fixturePassword } | ConvertTo-Json -Compress
    $loginResponse = Get-ApiResponse -Method 'POST' -Path '/api/auth/login' -Headers @{ 'x-tenant-id' = $tenantSlug } -Body $userLoginBody
    Assert-ApiSuccess $loginResponse 'user login'
    if ($null -eq $loginResponse.Json -or [string]::IsNullOrWhiteSpace([string]$loginResponse.Json.data)) { throw 'User login did not return the current response data token.' }
    $userToken = [string]$loginResponse.Json.data
    Set-ReportValue 'USER_TOKEN_CREATED' 'true'
    $userHeaders = @{ Authorization = "Bearer $userToken"; 'x-tenant-id' = $tenantSlug }
    $listResponse = Get-ApiResponse -Method 'GET' -Path '/api/product-enrichment/suggestions' -Headers $userHeaders
    Assert-ApiSuccess $listResponse 'review list API'
    $listItems = @($listResponse.Json.data.items)
    if ($null -eq ($listItems | Where-Object { [int]$_.suggestion_id -eq $suggestionID })) { throw 'Review list API did not return the E2E in_review suggestion.' }
    Set-ReportValue 'REVIEW_LIST_API' 'PASS'
    $detailResponse = Get-ApiResponse -Method 'GET' -Path ("/api/product-enrichment/suggestions/{0}" -f $suggestionID) -Headers $userHeaders
    Assert-ApiSuccess $detailResponse 'review detail API'
    Assert-DetailContract $detailResponse.Json.data
    if ($detailResponse.Json.data.safety.approvable -ne $true) {
        $blockingReasons = @($detailResponse.Json.data.safety.blocking_reasons | ForEach-Object { [string]$_ }) -join ','
        throw "Approval blocked by current deterministic review rules: $blockingReasons"
    }
    Set-ReportValue 'REVIEW_DETAIL_API' 'PASS'
    Add-Trace '10 review list API read; 11 review detail API returned source, proposals, provider, lifecycle, and safety'

    $phase = 'APPROVAL_AND_APPLICATION'
    $productBeforeApproval = Get-ProductSnapshot $incompleteSku
    $approveResponse = Get-ApiResponse -Method 'POST' -Path ("/api/product-enrichment/suggestions/{0}/approve" -f $suggestionID) -Headers $userHeaders -Body '{}'
    Assert-ApiSuccess $approveResponse 'approve API'
    $approvedState = Get-SuggestionByCurrentRunIdentity -SuggestionID $suggestionID -ProductID $incompleteProductID -SourceItemCode $incompleteSku
    if ($null -eq $approvedState -or [string]$approvedState.status -ne 'approved') { throw 'Approval did not transition the suggestion to approved.' }
    $productAfterApproval = Get-ProductSnapshot $incompleteSku
    $approvalChanged = (($productBeforeApproval | ConvertTo-Json -Depth 30 -Compress) -ne ($productAfterApproval | ConvertTo-Json -Depth 30 -Compress))
    Set-ReportValue 'APPROVE_API' 'PASS'
    Set-ReportValue 'PRODUCT_CHANGED_ON_APPROVAL' ($(if ($approvalChanged) { 'true' } else { 'false' }))
    if ($approvalChanged) { throw 'Approval changed the product before explicit apply.' }
    Add-Trace '12 suggestion approved with no product mutation'
    $protectedBeforeApply = Get-JsonSql -Database $tenantDatabase -Sql @"
SELECT jsonb_build_object(
  'product', to_jsonb(p) - 'brand_id' - 'category_id' - 'description' - 'updated_at',
  'barcodes', COALESCE((SELECT jsonb_agg(to_jsonb(b) ORDER BY b.id) FROM product_barcodes b WHERE b.product_id = p.id), '[]'::jsonb),
  'uom_conversions', COALESCE((SELECT jsonb_agg(to_jsonb(c) ORDER BY c.id) FROM product_uom_conversions c WHERE c.product_id = p.id), '[]'::jsonb),
  'inventory', COALESCE((SELECT jsonb_agg(to_jsonb(i) ORDER BY i.id) FROM inventory_stock i WHERE i.product_id = p.id), '[]'::jsonb),
  'prices', COALESCE((SELECT jsonb_agg(to_jsonb(pr) ORDER BY pr.id) FROM product_prices pr WHERE pr.product_id = p.id), '[]'::jsonb),
  'supplier_contracts', COALESCE((SELECT jsonb_agg(to_jsonb(pc) ORDER BY pc.id) FROM bp_price_contracts pc WHERE pc.product_id = p.id), '[]'::jsonb)
)::text FROM products p WHERE p.organization_id = $organizationID AND p.sku = $(ConvertTo-PgSqlLiteral $incompleteSku) LIMIT 1;
"@
    $auditBeforeApply = Get-Count -Database $tenantDatabase -Sql "SELECT count(*) FROM audit_logs WHERE organization_id = $organizationID AND table_name = 'product_enrichment_suggestions' AND record_id = '$suggestionID';"
    $applyResponse = Get-ApiResponse -Method 'POST' -Path ("/api/product-enrichment/suggestions/{0}/apply" -f $suggestionID) -Headers $userHeaders -Body '{}'
    Assert-ApiSuccess $applyResponse 'apply API'
    $applyData = $applyResponse.Json.data
    $appliedState = Get-SuggestionByCurrentRunIdentity -SuggestionID $suggestionID -ProductID $incompleteProductID -SourceItemCode $incompleteSku
    if ($null -eq $appliedState -or [string]$appliedState.status -ne 'applied') { throw 'Apply did not transition the suggestion to applied.' }
    $changedFields = @($applyData.changed_fields | ForEach-Object { [string]$_ })
    foreach ($field in $changedFields) { if (@('brand_id', 'category_id', 'description') -notcontains $field) { throw "Apply returned a prohibited changed field '$field'." } }
    $protectedAfterApply = Get-JsonSql -Database $tenantDatabase -Sql @"
SELECT jsonb_build_object(
  'product', to_jsonb(p) - 'brand_id' - 'category_id' - 'description' - 'updated_at',
  'barcodes', COALESCE((SELECT jsonb_agg(to_jsonb(b) ORDER BY b.id) FROM product_barcodes b WHERE b.product_id = p.id), '[]'::jsonb),
  'uom_conversions', COALESCE((SELECT jsonb_agg(to_jsonb(c) ORDER BY c.id) FROM product_uom_conversions c WHERE c.product_id = p.id), '[]'::jsonb),
  'inventory', COALESCE((SELECT jsonb_agg(to_jsonb(i) ORDER BY i.id) FROM inventory_stock i WHERE i.product_id = p.id), '[]'::jsonb),
  'prices', COALESCE((SELECT jsonb_agg(to_jsonb(pr) ORDER BY pr.id) FROM product_prices pr WHERE pr.product_id = p.id), '[]'::jsonb),
  'supplier_contracts', COALESCE((SELECT jsonb_agg(to_jsonb(pc) ORDER BY pc.id) FROM bp_price_contracts pc WHERE pc.product_id = p.id), '[]'::jsonb)
)::text FROM products p WHERE p.organization_id = $organizationID AND p.sku = $(ConvertTo-PgSqlLiteral $incompleteSku) LIMIT 1;
"@
    if (($protectedBeforeApply | ConvertTo-Json -Depth 30 -Compress) -ne ($protectedAfterApply | ConvertTo-Json -Depth 30 -Compress)) { throw 'Apply changed a field outside brand_id, category_id, or description.' }
    $auditAfterApply = Get-Count -Database $tenantDatabase -Sql "SELECT count(*) FROM audit_logs WHERE organization_id = $organizationID AND table_name = 'product_enrichment_suggestions' AND record_id = '$suggestionID';"
    if ($auditAfterApply -le $auditBeforeApply) { throw 'Apply did not create an audit record.' }
    Set-ReportValue 'APPLY_API' 'PASS'
    Set-ReportValue 'APPLY_AUDIT_CREATED' 'PASS'
    Set-ReportValue 'APPLY_CHANGED_FIELDS' (($changedFields -join ',') )
    Add-Trace '13 explicit apply revalidated the source; 14 narrow product fields changed; 15 audit created'
    $secondApplyResponse = Get-ApiResponse -Method 'POST' -Path ("/api/product-enrichment/suggestions/{0}/apply" -f $suggestionID) -Headers $userHeaders -Body '{}'
    Assert-ApiSuccess $secondApplyResponse 'second apply API'
    if ($null -eq $secondApplyResponse.Json.data -or $secondApplyResponse.Json.data.already_applied -ne $true) { throw 'Second apply did not return the current AlreadyApplied idempotent result.' }
    $auditAfterSecondApply = Get-Count -Database $tenantDatabase -Sql "SELECT count(*) FROM audit_logs WHERE organization_id = $organizationID AND table_name = 'product_enrichment_suggestions' AND record_id = '$suggestionID';"
    if ($auditAfterSecondApply -ne $auditAfterApply) { throw 'Second apply created a duplicate apply audit.' }
    Set-ReportValue 'SECOND_APPLY_IDEMPOTENT' 'PASS'

    $phase = 'REPEAT_SAP_SYNC'
    $suggestionsBeforeRepeat = Get-SuggestionCountForProductIdentity -ProductID $incompleteProductID -SourceItemCode $incompleteSku
    $attemptsBeforeRepeat = [int]$appliedState.attempt_count
    $repeatPayload = New-SapPayload $incompleteSku 'Pantene shampoo smooth care 400 ml' '' '' $repeatBatchID
    [void](Invoke-SapMigration $repeatPayload)
    $suggestionsAfterRepeat = Get-SuggestionCountForProductIdentity -ProductID $incompleteProductID -SourceItemCode $incompleteSku
    $repeatState = Get-SuggestionByCurrentRunIdentity -SuggestionID $suggestionID -ProductID $incompleteProductID -SourceItemCode $incompleteSku
    $repeatProduct = Get-ProductMutableState $incompleteSku
    if ($suggestionsAfterRepeat -ne $suggestionsBeforeRepeat -or [string]$repeatState.status -ne 'applied' -or [int]$repeatState.attempt_count -ne $attemptsBeforeRepeat) { throw 'Repeat SAP sync reset or duplicated the accepted enrichment lifecycle.' }
    if ([int]$repeatProduct.brand_id -ne [int]$productAfterApproval.brand_id -or [string]$repeatProduct.description -ne [string]$productAfterApproval.description) { throw 'Repeat SAP sync cleared an explicitly applied brand or description.' }
    Set-ReportValue 'REPEAT_SAP_SYNC' 'PASS'
    Set-ReportValue 'SAP_EMPTY_RESETS_AI_BRAND' 'false'
    Set-ReportValue 'SAP_EMPTY_RESETS_AI_DESCRIPTION' 'false'
    Set-ReportValue 'NEW_SUGGESTIONS_AFTER_REPEAT' '0'
    Set-ReportValue 'PROVIDER_CALLS_AFTER_REPEAT' '0'
    Set-ReportValue 'FIVE_MINUTE_SYNC_ENRICHMENT_LOOP_RISK' 'false'
    Add-Trace '16 repeated SAP poll caused no AI loop and preserved accepted brand/description'
    Set-ReportValue 'PROVIDER_CALL_ACCOUNTING_METHOD' 'tenant suggestion lifecycle, provider/model metadata, and attempt_count; no network interception'

    $phase = 'KILL_SWITCH'
    Set-ReportValue 'OPERATIONAL_ONLY_CHANGE_TEST' 'SKIPPED'
    Stop-CloudServer
    Start-CloudServer $false
    $disabledPayload = New-SapPayload $disabledSku 'E2E disabled enrichment shampoo 250 ml' '' '' $disabledBatchID
    [void](Invoke-SapMigration $disabledPayload)
    if ($null -eq (Get-ProductSnapshot $disabledSku)) { throw 'Enrichment-disabled SAP product did not persist.' }
    $disabledSuggestionCount = Get-SuggestionCountForSourceIdentity $disabledSku
    if ($disabledSuggestionCount -ne 0) { throw 'Enrichment-disabled SAP sync created a new suggestion.' }
    $masterAfterCounts = Get-MasterBusinessCounts
    foreach ($name in @('organizations', 'products', 'suggestions', 'audit_logs')) {
        if ([int64]$masterAfterCounts.$name -ne [int64]$baselineMasterCounts.$name) { throw "Master business count changed for '$name'." }
    }
    Set-ReportValue 'ENRICHMENT_DISABLED_SAP_SYNC_WORKS' 'true'
    Set-ReportValue 'ENRICHMENT_DISABLED_NEW_PENDING_SUGGESTIONS' '0'
    Set-ReportValue 'ENRICHMENT_DISABLED_PROVIDER_CALLS' '0'
    Set-ReportValue 'DISABLED_QUEUE_GROWTH_RISK' 'false'
    Set-ReportValue 'MASTER_BUSINESS_WRITES' '0'
    Add-Trace '17 enrichment-disabled server boot passed; 18 SAP sync still worked with no Stage 2A queue growth'
    $report.FINAL_VERDICT = 'LOCAL_FULL_E2E_PASS'
}
catch {
    Write-Host ("FAILURE_PHASE=" + $phase) -ForegroundColor Red
    Write-Host ("FAILURE_REASON=" + (Get-SanitizedFailureMessage $_.Exception.Message)) -ForegroundColor Red
    if ($phase -in @('REVIEW_API', 'APPROVAL_AND_APPLICATION', 'REPEAT_SAP_SYNC', 'KILL_SWITCH')) { $report.FINAL_VERDICT = 'LOCAL_FULL_E2E_PARTIAL' }
    else { $report.FINAL_VERDICT = 'LOCAL_FULL_E2E_FAILED' }
}
finally {
    Stop-CloudServer
    Restore-ProcessEnvironment @('PGPASSWORD', 'JWT_SECRET', 'MASTER_DB_URL', 'DEEPSEEK_API_KEY', 'NEMBUS_E2E_FIXTURE_PASSWORD')
    $dbPasswordPlainForServer = $null
    $deepSeekKeyPlain = $null
    $fixturePassword = $null
    $m2mToken = $null
    $userToken = $null
    $jwtSecret = $null
    if ($null -ne $deepSeekKeySecure) { $deepSeekKeySecure.Dispose() }
    if ($null -ne $dbPasswordSecure) { $dbPasswordSecure.Dispose() }
    if ($null -ne $runtimeRoot -and (Test-Path -LiteralPath $runtimeRoot)) { Remove-Item -LiteralPath $runtimeRoot -Recurse -Force -ErrorAction SilentlyContinue }
    Write-Host ''
    Write-Host '===== LOCAL FULL E2E SANITIZED REPORT ====='
    foreach ($key in $report.Keys) { Write-Host ("{0}={1}" -f $key, $report[$key]) }
    Write-Host 'FLOW_TRACE_BEGIN'
    $step = 1
    foreach ($item in $trace) { Write-Host ("{0} {1}" -f $step, $item); $step++ }
    Write-Host 'FLOW_TRACE_END'
}
