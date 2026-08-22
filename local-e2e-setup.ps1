Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# Temporary, local-only preparation for the next Nembus enrichment E2E run.
# This script deliberately does not create business fixtures or call providers.

$repoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location -LiteralPath $repoRoot

$localHost = '127.0.0.1'
$localPort = 5432
$roleName = 'nembus_e2e_user'
$masterDatabase = 'nembus_e2e_master'
$tenantDatabase = 'nembus_e2e_tenant'
$tenantSlug = 'e2e-enrichment'
$databaseMarker = 'Nembus local E2E disposable database created by local-e2e-setup.ps1'
$roleMarker = 'Nembus local E2E disposable role created by local-e2e-setup.ps1'

$report = [ordered]@{
    LOCAL_DB_SAFETY_GATE = 'NOT_RUN'
    LOCAL_POSTGRES_ADMIN_AUTH = 'NOT_RUN'
    E2E_ROLE_READY = 'NOT_RUN'
    E2E_MASTER_DB_READY = 'NOT_RUN'
    E2E_TENANT_DB_READY = 'NOT_RUN'
    E2E_MASTER_AUTH = 'NOT_RUN'
    E2E_TENANT_AUTH = 'NOT_RUN'
    MASTER_MIGRATION_READY = 'NOT_RUN'
    E2E_TENANT_REGISTERED = 'NOT_RUN'
    TENANT_MIGRATION_READY = 'NOT_RUN'
    ENRICHMENT_SUGGESTION_TABLE_READY = 'NOT_RUN'
    REVIEW_PERMISSION_PRESENT = 'NOT_RUN'
    APPLY_PERMISSION_PRESENT = 'NOT_RUN'
    CLOUD_SERVER_LOCAL_DB_STARTUP = 'NOT_RUN'
    CLOUD_SERVER_TEST_PORT = 'NOT_SELECTED'
    HEALTH_ENDPOINT_PASS = 'NOT_RUN'
    AI_PROVIDER_REQUIRED_WHILE_DISABLED = 'false'
    FINAL_VERDICT = 'LOCAL_E2E_DATABASE_BLOCKED'
}

$collisionDetected = $false
$cloudProcess = $null
$cloudLogDirectory = $null
$cloudRawStdout = $null
$cloudRawStderr = $null
$cloudSafeStdout = $null
$cloudSafeStderr = $null
$httpPort = $null
$grpcPort = $null
$adminPassword = $null
$e2ePassword = $null
$atlasExe = $null
$goExe = $null
$psqlExe = $null
$pgIsReadyExe = $null

$environmentNamesToRestore = @(
    'ENV',
    'ENRICHMENT_ENABLED',
    'PORT',
    'GRPC_PORT',
    'ZATCA_ENABLED',
    'PG_DUMP_PATH'
)
$originalEnvironment = @{}
foreach ($environmentName in $environmentNamesToRestore) {
    $existingValue = [Environment]::GetEnvironmentVariable($environmentName, 'Process')
    if ($null -ne $existingValue) {
        $originalEnvironment[$environmentName] = $existingValue
    }
}

function Set-ReportValue {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string]$Value
    )

    $report[$Name] = $Value
}

function Resolve-LocalTool {
    param(
        [Parameter(Mandatory = $true)][string]$PreferredPath,
        [Parameter(Mandatory = $true)][string]$CommandName
    )

    if (Test-Path -LiteralPath $PreferredPath -PathType Leaf) {
        return (Get-Item -LiteralPath $PreferredPath).FullName
    }

    $command = Get-Command $CommandName -ErrorAction SilentlyContinue
    if ($null -ne $command -and $command.CommandType -eq 'Application') {
        return $command.Source
    }

    throw "Required local tool '$CommandName' was not found. Checked '$PreferredPath' and PATH."
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

function ConvertTo-PgIdentifier {
    param([Parameter(Mandatory = $true)][AllowEmptyString()][string]$Value)

    return '"' + $Value.Replace('"', '""') + '"'
}

function Invoke-Psql {
    param(
        [Parameter(Mandatory = $true)][string]$Database,
        [Parameter(Mandatory = $true)][string]$Username,
        [Parameter(Mandatory = $true)][Security.SecureString]$Password,
        [Parameter(Mandatory = $true)][string]$Sql
    )

    $plainPassword = $null
    $argumentList = @(
        '-X',
        '-w',
        '-q',
        '-t',
        '-A',
        '-v', 'ON_ERROR_STOP=1',
        '-h', $localHost,
        '-p', [string]$localPort,
        '-U', $Username,
        '-d', $Database
    )

    try {
        $plainPassword = Convert-SecureStringToPlainText -SecureString $Password
        [Environment]::SetEnvironmentVariable('PGPASSWORD', $plainPassword, 'Process')
        # SQL is supplied on stdin so neither SQL literals nor psql variables
        # can accidentally be exposed through the child process command line.
        $output = $Sql | & $psqlExe @argumentList 2>&1
        $exitCode = $LASTEXITCODE
        if ($exitCode -ne 0) {
            throw "psql failed for database '$Database' with exit code $exitCode."
        }

        return (($output | ForEach-Object { $_.ToString() }) -join "`n").Trim()
    }
    finally {
        $plainPassword = $null
        Remove-Item -LiteralPath 'Env:PGPASSWORD' -ErrorAction SilentlyContinue
    }
}

function Set-E2ERolePassword {
    param(
        [Parameter(Mandatory = $true)][string]$Database,
        [Parameter(Mandatory = $true)][string]$Username,
        [Parameter(Mandatory = $true)][Security.SecureString]$AdminPassword,
        [Parameter(Mandatory = $true)][Security.SecureString]$E2EPassword,
        [Parameter(Mandatory = $true)][string]$RoleName,
        [Parameter(Mandatory = $true)][bool]$Create
    )

    $plainPassword = $null
    $passwordLiteral = $null
    $roleIdentifier = $null
    $sql = $null
    try {
        $plainPassword = Convert-SecureStringToPlainText -SecureString $E2EPassword
        $passwordLiteral = ConvertTo-PgSqlLiteral -Value $plainPassword
        $roleIdentifier = ConvertTo-PgIdentifier -Value $RoleName
        $verb = if ($Create) { 'CREATE ROLE' } else { 'ALTER ROLE' }
        $sql = "$verb $roleIdentifier WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD $passwordLiteral;"
        Invoke-Psql -Database $Database -Username $Username -Password $AdminPassword -Sql $sql | Out-Null
    }
    finally {
        $plainPassword = $null
        $passwordLiteral = $null
        $roleIdentifier = $null
        $sql = $null
    }
}

function Set-SetupPhase {
    param([Parameter(Mandatory = $true)][string]$Phase)

    Write-Host "SETUP_PHASE=$Phase"
}

function New-DatabaseUrl {
    param(
        [Parameter(Mandatory = $true)][string]$Database,
        [Parameter(Mandatory = $true)][Security.SecureString]$Password
    )

    $plainPassword = $null
    try {
        $plainPassword = Convert-SecureStringToPlainText -SecureString $Password
        $encodedUser = [Uri]::EscapeDataString($roleName)
        $encodedPassword = [Uri]::EscapeDataString($plainPassword)
        return "postgres://$encodedUser`:$encodedPassword@$localHost`:$localPort/$Database`?sslmode=disable"
    }
    finally {
        $plainPassword = $null
    }
}

function Invoke-AtlasMigration {
    param(
        [Parameter(Mandatory = $true)][string]$Database,
        [Parameter(Mandatory = $true)][Security.SecureString]$Password,
        [Parameter(Mandatory = $true)][ValidateSet('MASTER', 'TENANT')][string]$Target
    )

    $databaseUrl = $null
    $previousMasterDbUrl = [Environment]::GetEnvironmentVariable('MASTER_DB_URL', 'Process')
    try {
        $databaseUrl = New-DatabaseUrl -Database $Database -Password $Password
        # Atlas reads this URL from packages/core/atlas.hcl. Keeping it in the
        # child environment avoids putting the password-bearing URL in argv.
        [Environment]::SetEnvironmentVariable('MASTER_DB_URL', $databaseUrl, 'Process')
        Push-Location (Join-Path $repoRoot 'packages/core')
        try {
            Write-Host "MIGRATION_TARGET=$Target"
            # The local Atlas environment resolves its migration directory as
            # file://db/migrations from packages/core. Fresh and partially
            # migrated disposable databases must apply the complete history;
            # no baseline is valid in this current migration directory.
            $arguments = @('migrate', 'apply', '--env', 'local')
            $output = & $atlasExe @arguments 2>&1
            $exitCode = $LASTEXITCODE
            if ($exitCode -ne 0) {
                throw "Atlas migration failed for $Target with exit code $exitCode."
            }
        }
        finally {
            Pop-Location
        }
    }
    finally {
        $databaseUrl = $null
        if ($null -eq $previousMasterDbUrl) {
            Remove-Item -LiteralPath 'Env:MASTER_DB_URL' -ErrorAction SilentlyContinue
        }
        else {
            [Environment]::SetEnvironmentVariable('MASTER_DB_URL', $previousMasterDbUrl, 'Process')
        }
    }
}

function Invoke-AtlasStatus {
    param(
        [Parameter(Mandatory = $true)][string]$Database,
        [Parameter(Mandatory = $true)][Security.SecureString]$Password
    )

    $databaseUrl = $null
    $previousMasterDbUrl = [Environment]::GetEnvironmentVariable('MASTER_DB_URL', 'Process')
    try {
        $databaseUrl = New-DatabaseUrl -Database $Database -Password $Password
        [Environment]::SetEnvironmentVariable('MASTER_DB_URL', $databaseUrl, 'Process')
        Push-Location (Join-Path $repoRoot 'packages/core')
        try {
            $arguments = @('migrate', 'status', '--env', 'local')
            $output = & $atlasExe @arguments 2>&1
            $exitCode = $LASTEXITCODE
            if ($exitCode -ne 0) {
                throw "Atlas migration status failed for '$Database' with exit code $exitCode."
            }
        }
        finally {
            Pop-Location
        }
    }
    finally {
        $databaseUrl = $null
        if ($null -eq $previousMasterDbUrl) {
            Remove-Item -LiteralPath 'Env:MASTER_DB_URL' -ErrorAction SilentlyContinue
        }
        else {
            [Environment]::SetEnvironmentVariable('MASTER_DB_URL', $previousMasterDbUrl, 'Process')
        }
    }
}

function Test-LocalPortFree {
    param([Parameter(Mandatory = $true)][int]$Port)

    $listener = $null
    try {
        $listener = [Net.Sockets.TcpListener]::new([Net.IPAddress]::Any, $Port)
        $listener.Start()
        return $true
    }
    catch {
        return $false
    }
    finally {
        if ($null -ne $listener) {
            $listener.Stop()
        }
    }
}

function Find-FreeLocalPort {
    param(
        [Parameter(Mandatory = $true)][int]$PreferredPort,
        [Parameter(Mandatory = $true)][int]$MaximumPort
    )

    for ($candidate = $PreferredPort; $candidate -le $MaximumPort; $candidate++) {
        if (Test-LocalPortFree -Port $candidate) {
            return $candidate
        }
    }

    throw "No free local TCP port was found between $PreferredPort and $MaximumPort."
}

function Test-PsqlBoolean {
    param(
        [Parameter(Mandatory = $true)][string]$Database,
        [Parameter(Mandatory = $true)][string]$Username,
        [Parameter(Mandatory = $true)][Security.SecureString]$Password,
        [Parameter(Mandatory = $true)][string]$Sql
    )

    $value = Invoke-Psql -Database $Database -Username $Username -Password $Password -Sql $Sql
    return $value.Trim().ToLowerInvariant() -eq 't'
}

function Test-DatabaseTable {
    param(
        [Parameter(Mandatory = $true)][string]$Database,
        [Parameter(Mandatory = $true)][Security.SecureString]$Password,
        [Parameter(Mandatory = $true)][string]$Schema,
        [Parameter(Mandatory = $true)][string]$Table
    )

    $qualifiedNameLiteral = ConvertTo-PgSqlLiteral -Value "${Schema}.${Table}"
    return Test-PsqlBoolean -Database $Database -Username $roleName -Password $Password -Sql "SELECT to_regclass($qualifiedNameLiteral) IS NOT NULL;"
}

function Get-ChildProcessIds {
    param([Parameter(Mandatory = $true)][int]$ParentId)

    $children = @(Get-CimInstance Win32_Process -Filter "ParentProcessId = $ParentId" -ErrorAction SilentlyContinue)
    $result = @()
    foreach ($child in $children) {
        $result += [int]$child.ProcessId
        $result += Get-ChildProcessIds -ParentId ([int]$child.ProcessId)
    }
    return $result
}

function Stop-CloudServer {
    if ($null -eq $cloudProcess) {
        return
    }

    try {
        $descendants = @(Get-ChildProcessIds -ParentId $cloudProcess.Id)
        foreach ($processId in ($descendants | Sort-Object -Descending)) {
            Stop-Process -Id $processId -Force -ErrorAction SilentlyContinue
        }
        if (-not $cloudProcess.HasExited) {
            Stop-Process -Id $cloudProcess.Id -Force -ErrorAction SilentlyContinue
        }
        $cloudProcess.WaitForExit(5000)
    }
    finally {
        $script:cloudProcess = $null
    }
}

function Sanitize-CloudLogs {
    if ($null -eq $cloudLogDirectory) {
        return
    }

    # Cloud logs should not contain either database password. The URL and
    # ephemeral JWT are redacted here without converting SecureString values.
    $secretValues = @([string]$env:MASTER_DB_URL, [string]$env:JWT_SECRET)

    foreach ($rawPath in @($cloudRawStdout, $cloudRawStderr)) {
        if ([string]::IsNullOrWhiteSpace($rawPath) -or -not (Test-Path -LiteralPath $rawPath)) {
            continue
        }

        $safePath = if ($rawPath -eq $cloudRawStdout) { $cloudSafeStdout } else { $cloudSafeStderr }
        try {
            [string]$content = Get-Content -LiteralPath $rawPath -Raw -ErrorAction SilentlyContinue
            foreach ($secretValue in $secretValues) {
                if (-not [string]::IsNullOrEmpty($secretValue)) {
                    $content = $content.Replace($secretValue, '<redacted>')
                }
            }
            $content = [Regex]::Replace($content, '(?i)postgres(?:ql)?://[^\s]+', 'postgres://<redacted>')
            $content = [Regex]::Replace($content, '(?i)(password\s*[=:]\s*)[^\s]+', '$1<redacted>')
            Set-Content -LiteralPath $safePath -Value $content -NoNewline
        }
        finally {
            Remove-Item -LiteralPath $rawPath -Force -ErrorAction SilentlyContinue
        }
    }
}

function Restore-ProcessEnvironment {
    foreach ($environmentName in $environmentNamesToRestore) {
        if ($originalEnvironment.ContainsKey($environmentName)) {
            [Environment]::SetEnvironmentVariable($environmentName, $originalEnvironment[$environmentName], 'Process')
        }
        else {
            Remove-Item -LiteralPath "Env:$environmentName" -ErrorAction SilentlyContinue
        }
    }

    # These are generated by this script and are never restored from a parent
    # process, because a pre-existing MASTER_DB_URL must never be reused.
    Remove-Item -LiteralPath 'Env:PGPASSWORD' -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath 'Env:MASTER_DB_URL' -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath 'Env:JWT_SECRET' -ErrorAction SilentlyContinue
}

try {
    $psqlExe = Resolve-LocalTool -PreferredPath 'C:\Program Files\PostgreSQL\18\bin\psql.exe' -CommandName 'psql.exe'
    $pgIsReadyExe = Resolve-LocalTool -PreferredPath 'C:\Program Files\PostgreSQL\18\bin\pg_isready.exe' -CommandName 'pg_isready.exe'
    $atlasExe = Resolve-LocalTool -PreferredPath 'C:\Program Files\Atlas\atlas.exe' -CommandName 'atlas.exe'
    $goExe = Resolve-LocalTool -PreferredPath 'C:\Program Files\Go\bin\go.exe' -CommandName 'go.exe'

    $pgReadyOutput = & $pgIsReadyExe '-h' $localHost '-p' ([string]$localPort) '-d' 'postgres' 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw 'PostgreSQL is not ready at the required local endpoint 127.0.0.1:5432.'
    }
    Set-ReportValue -Name 'LOCAL_DB_SAFETY_GATE' -Value 'PASSED'

    $adminUsername = Read-Host 'LOCAL PostgreSQL admin username'
    if ([string]::IsNullOrWhiteSpace($adminUsername)) {
        throw 'A local PostgreSQL administrator username is required.'
    }
    $adminPassword = Read-Host 'LOCAL PostgreSQL admin password' -AsSecureString

    try {
        $adminProbe = Invoke-Psql -Database 'postgres' -Username $adminUsername -Password $adminPassword -Sql 'SELECT current_user || ''@'' || current_database();'
        if ([string]::IsNullOrWhiteSpace($adminProbe)) {
            throw 'The local PostgreSQL admin validation query returned no result.'
        }
        Set-ReportValue -Name 'LOCAL_POSTGRES_ADMIN_AUTH' -Value 'PASSED'
    }
    catch {
        Set-ReportValue -Name 'LOCAL_POSTGRES_ADMIN_AUTH' -Value 'FAILED'
        throw 'LOCAL_POSTGRES_ADMIN_AUTH=FAILED. The supplied local administrator credentials were not accepted.'
    }

    $e2ePassword = Read-Host 'Choose a password for nembus_e2e_user' -AsSecureString

    $roleNameLiteral = ConvertTo-PgSqlLiteral -Value $roleName
    $roleMarkerLiteral = ConvertTo-PgSqlLiteral -Value $roleMarker
    $databaseMarkerLiteral = ConvertTo-PgSqlLiteral -Value $databaseMarker
    $roleStateSql = @"
SELECT CASE
    WHEN NOT EXISTS (SELECT 1 FROM pg_authid WHERE rolname = $roleNameLiteral) THEN 'absent'
    WHEN EXISTS (
        SELECT 1
        FROM pg_authid
        WHERE rolname = $roleNameLiteral
          AND rolcanlogin
          AND NOT rolsuper
          AND NOT rolcreatedb
          AND NOT rolcreaterole
          AND NOT rolreplication
          AND NOT rolbypassrls
          AND shobj_description(oid, 'pg_authid') = $roleMarkerLiteral
    )
    AND NOT EXISTS (
        SELECT 1
        FROM pg_auth_members m
        JOIN pg_authid r ON r.oid = m.roleid OR r.oid = m.member
        WHERE r.rolname = $roleNameLiteral
    ) THEN 'owned'
    ELSE 'collision'
END;
"@
    $roleState = (Invoke-Psql -Database 'postgres' -Username $adminUsername -Password $adminPassword -Sql $roleStateSql).Trim().ToLowerInvariant()
    if ($roleState -eq 'collision') {
        throw 'The role nembus_e2e_user exists but is not positively identified as this workflow''s disposable role.'
    }

    Set-SetupPhase -Phase 'CREATE_E2E_ROLE'
    Set-E2ERolePassword -Database 'postgres' -Username $adminUsername -AdminPassword $adminPassword -E2EPassword $e2ePassword -RoleName $roleName -Create ($roleState -eq 'absent')
    if ($roleState -eq 'absent') {
        $roleIdentifier = ConvertTo-PgIdentifier -Value $roleName
        Invoke-Psql -Database 'postgres' -Username $adminUsername -Password $adminPassword -Sql "COMMENT ON ROLE $roleIdentifier IS $roleMarkerLiteral;" | Out-Null
    }
    Set-ReportValue -Name 'E2E_ROLE_READY' -Value 'PASSED'

    $databaseStateSql = {
        param($databaseName)
        $databaseNameLiteral = ConvertTo-PgSqlLiteral -Value $databaseName
        $roleNameLiteral = ConvertTo-PgSqlLiteral -Value $roleName
        $databaseMarkerLiteral = ConvertTo-PgSqlLiteral -Value $databaseMarker
        return @"
SELECT CASE
    WHEN NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = $databaseNameLiteral) THEN 'absent'
    WHEN EXISTS (
        SELECT 1
        FROM pg_database
        WHERE datname = $databaseNameLiteral
          AND pg_get_userbyid(datdba) = $roleNameLiteral
          AND shobj_description(oid, 'pg_database') = $databaseMarkerLiteral
    ) THEN 'owned'
    ELSE 'collision'
END;
"@
    }

    $masterState = (Invoke-Psql -Database 'postgres' -Username $adminUsername -Password $adminPassword -Sql (& $databaseStateSql $masterDatabase)).Trim().ToLowerInvariant()
    $tenantState = (Invoke-Psql -Database 'postgres' -Username $adminUsername -Password $adminPassword -Sql (& $databaseStateSql $tenantDatabase)).Trim().ToLowerInvariant()
    Write-Host "E2E database metadata: master=$masterState tenant=$tenantState"
    if ($masterState -eq 'collision' -or $tenantState -eq 'collision') {
        $collisionDetected = $true
        throw 'E2E_DATABASE_NAME_COLLISION=true. An existing database is not positively identified as disposable E2E data.'
    }

    Set-SetupPhase -Phase 'CREATE_MASTER_DB'
    if ($masterState -eq 'absent') {
        $masterIdentifier = ConvertTo-PgIdentifier -Value $masterDatabase
        $roleIdentifier = ConvertTo-PgIdentifier -Value $roleName
        Invoke-Psql -Database 'postgres' -Username $adminUsername -Password $adminPassword -Sql "CREATE DATABASE $masterIdentifier OWNER $roleIdentifier TEMPLATE template0;" | Out-Null
        Invoke-Psql -Database 'postgres' -Username $adminUsername -Password $adminPassword -Sql "COMMENT ON DATABASE $masterIdentifier IS $databaseMarkerLiteral;" | Out-Null
    }
    if ($tenantState -eq 'absent') {
        $tenantIdentifier = ConvertTo-PgIdentifier -Value $tenantDatabase
        $roleIdentifier = ConvertTo-PgIdentifier -Value $roleName
        Invoke-Psql -Database 'postgres' -Username $adminUsername -Password $adminPassword -Sql "CREATE DATABASE $tenantIdentifier OWNER $roleIdentifier TEMPLATE template0;" | Out-Null
        Invoke-Psql -Database 'postgres' -Username $adminUsername -Password $adminPassword -Sql "COMMENT ON DATABASE $tenantIdentifier IS $databaseMarkerLiteral;" | Out-Null
    }
    Set-ReportValue -Name 'E2E_MASTER_DB_READY' -Value 'PASSED'
    Set-ReportValue -Name 'E2E_TENANT_DB_READY' -Value 'PASSED'

    $masterAuthProbe = Invoke-Psql -Database $masterDatabase -Username $roleName -Password $e2ePassword -Sql 'SELECT current_database();'
    if ($masterAuthProbe.Trim() -ne $masterDatabase) {
        throw 'The E2E role could not authenticate to the disposable master database.'
    }
    Set-ReportValue -Name 'E2E_MASTER_AUTH' -Value 'PASSED'

    $tenantAuthProbe = Invoke-Psql -Database $tenantDatabase -Username $roleName -Password $e2ePassword -Sql 'SELECT current_database();'
    if ($tenantAuthProbe.Trim() -ne $tenantDatabase) {
        throw 'The E2E role could not authenticate to the disposable tenant database.'
    }
    Set-ReportValue -Name 'E2E_TENANT_AUTH' -Value 'PASSED'

    # The repository's migrate-tenants command delegates to this exact Atlas
    # apply/status pair. Run it per database so only these disposable databases
    # are touched and no other active registry row can be migrated.
    Set-SetupPhase -Phase 'MASTER_MIGRATION'
    Invoke-AtlasMigration -Database $masterDatabase -Password $e2ePassword -Target 'MASTER'
    Invoke-AtlasStatus -Database $masterDatabase -Password $e2ePassword
    Set-ReportValue -Name 'MASTER_MIGRATION_READY' -Value 'PASSED'

    $tenantConnectionUrl = $null
    try {
        $tenantConnectionUrl = New-DatabaseUrl -Database $tenantDatabase -Password $e2ePassword
        $tenantSlugLiteral = ConvertTo-PgSqlLiteral -Value $tenantSlug
        $tenantNameLiteral = ConvertTo-PgSqlLiteral -Value 'Nembus E2E Enrichment'
        $tenantConnectionLiteral = ConvertTo-PgSqlLiteral -Value $tenantConnectionUrl
        $registryStateSql = @"
SELECT CASE
    WHEN NOT EXISTS (SELECT 1 FROM tenants WHERE slug = $tenantSlugLiteral) THEN 'absent'
    WHEN (
        SELECT count(*)
        FROM tenants
        WHERE slug = $tenantSlugLiteral
          AND tenant_name = $tenantNameLiteral
          AND is_active = true
          AND db_conn_str = $tenantConnectionLiteral
    ) = 1 THEN 'owned'
    ELSE 'collision'
END;
"@
        Set-SetupPhase -Phase 'TENANT_REGISTRATION'
        $registryState = (Invoke-Psql -Database $masterDatabase -Username $roleName -Password $e2ePassword -Sql $registryStateSql).Trim().ToLowerInvariant()
        if ($registryState -eq 'collision') {
            throw 'The e2e-enrichment tenant registry row exists with unexpected metadata.'
        }
        if ($registryState -eq 'absent') {
            Invoke-Psql -Database $masterDatabase -Username $roleName -Password $e2ePassword -Sql "INSERT INTO tenants (tenant_name, slug, db_conn_str, is_active, settings) VALUES ($tenantNameLiteral, $tenantSlugLiteral, $tenantConnectionLiteral, true, '{}'::jsonb);" | Out-Null
        }

        $unexpectedActiveTenants = Invoke-Psql -Database $masterDatabase -Username $roleName -Password $e2ePassword -Sql "SELECT count(*) FROM tenants WHERE is_active = true AND slug <> $tenantSlugLiteral;"
        if ([int]$unexpectedActiveTenants.Trim() -ne 0) {
            throw 'The disposable master registry contains an unexpected active tenant; refusing to expose it to migration.'
        }

        $registered = Test-PsqlBoolean -Database $masterDatabase -Username $roleName -Password $e2ePassword -Sql "SELECT count(*) = 1 FROM tenants WHERE slug = $tenantSlugLiteral AND is_active = true AND db_conn_str = $tenantConnectionLiteral;"
        if (-not $registered) {
            throw 'The e2e-enrichment tenant registry row could not be verified.'
        }
    }
    finally {
        $tenantConnectionUrl = $null
        $tenantConnectionLiteral = $null
    }
    Set-ReportValue -Name 'E2E_TENANT_REGISTERED' -Value 'PASSED'

    Set-SetupPhase -Phase 'TENANT_MIGRATION'
    Invoke-AtlasMigration -Database $tenantDatabase -Password $e2ePassword -Target 'TENANT'
    Invoke-AtlasStatus -Database $tenantDatabase -Password $e2ePassword
    Set-ReportValue -Name 'TENANT_MIGRATION_READY' -Value 'PASSED'

    $requiredTenantTables = @(
        @{ Schema = 'public'; Table = 'organizations' },
        @{ Schema = 'public'; Table = 'users' },
        @{ Schema = 'public'; Table = 'roles' },
        @{ Schema = 'public'; Table = 'permissions' },
        @{ Schema = 'public'; Table = 'products' },
        @{ Schema = 'public'; Table = 'brands' },
        @{ Schema = 'public'; Table = 'product_categories' },
        @{ Schema = 'public'; Table = 'product_enrichment_suggestions' },
        @{ Schema = 'public'; Table = 'audit_logs' },
        @{ Schema = 'staging'; Table = 'sap_migration_batches' },
        @{ Schema = 'staging'; Table = 'sap_stores' },
        @{ Schema = 'staging'; Table = 'sap_products' },
        @{ Schema = 'staging'; Table = 'sap_inventory' }
    )
    foreach ($requiredTable in $requiredTenantTables) {
        if (-not (Test-DatabaseTable -Database $tenantDatabase -Password $e2ePassword -Schema $requiredTable.Schema -Table $requiredTable.Table)) {
            throw "Required tenant table '$($requiredTable.Schema).$($requiredTable.Table)' is missing."
        }
    }
    Set-ReportValue -Name 'ENRICHMENT_SUGGESTION_TABLE_READY' -Value 'PASSED'

    if (-not (Test-PsqlBoolean -Database $tenantDatabase -Username $roleName -Password $e2ePassword -Sql "SELECT EXISTS (SELECT 1 FROM permissions WHERE code = 'product_enrichment:review');")) {
        throw 'The product_enrichment:review permission is missing.'
    }
    Set-ReportValue -Name 'REVIEW_PERMISSION_PRESENT' -Value 'PASSED'

    if (-not (Test-PsqlBoolean -Database $tenantDatabase -Username $roleName -Password $e2ePassword -Sql "SELECT EXISTS (SELECT 1 FROM permissions WHERE code = 'product_enrichment:apply');")) {
        throw 'The product_enrichment:apply permission is missing.'
    }
    Set-ReportValue -Name 'APPLY_PERMISSION_PRESENT' -Value 'PASSED'

    # The server is booted only against the disposable master registry. The
    # tenant database remains the business/RBAC/product/SAP/enrichment store.
    # Enrichment is disabled so setup never requires or calls an AI provider.
    $httpPort = Find-FreeLocalPort -PreferredPort 18080 -MaximumPort 18120
    $grpcPort = Find-FreeLocalPort -PreferredPort 15051 -MaximumPort 15090
    Set-ReportValue -Name 'CLOUD_SERVER_TEST_PORT' -Value ([string]$httpPort)
    $cloudLogDirectory = Join-Path ([IO.Path]::GetTempPath()) ("nembus-e2e-" + [Guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $cloudLogDirectory -Force | Out-Null
    $cloudRawStdout = Join-Path $cloudLogDirectory 'cloud.stdout.raw.log'
    $cloudRawStderr = Join-Path $cloudLogDirectory 'cloud.stderr.raw.log'
    $cloudSafeStdout = Join-Path $cloudLogDirectory 'cloud.stdout.log'
    $cloudSafeStderr = Join-Path $cloudLogDirectory 'cloud.stderr.log'

    $serverDatabaseUrl = $null
    try {
        $serverDatabaseUrl = New-DatabaseUrl -Database $masterDatabase -Password $e2ePassword
        [Environment]::SetEnvironmentVariable('MASTER_DB_URL', $serverDatabaseUrl, 'Process')
        [Environment]::SetEnvironmentVariable('ENRICHMENT_ENABLED', 'false', 'Process')
        [Environment]::SetEnvironmentVariable('PORT', [string]$httpPort, 'Process')
        [Environment]::SetEnvironmentVariable('GRPC_PORT', [string]$grpcPort, 'Process')
        [Environment]::SetEnvironmentVariable('ENV', 'development', 'Process')
        [Environment]::SetEnvironmentVariable('ZATCA_ENABLED', 'false', 'Process')
        $jwtBytes = New-Object -TypeName 'System.Byte[]' -ArgumentList 32
        $rng = [Security.Cryptography.RandomNumberGenerator]::Create()
        try {
            $rng.GetBytes($jwtBytes)
        }
        finally {
            $rng.Dispose()
        }
        [Environment]::SetEnvironmentVariable('JWT_SECRET', [Convert]::ToBase64String($jwtBytes), 'Process')
        $jwtBytes = $null

        $cloudProcess = Start-Process -FilePath $goExe -WorkingDirectory (Join-Path $repoRoot 'apps/cloud-server') -ArgumentList @('run', 'main.go') -RedirectStandardOutput $cloudRawStdout -RedirectStandardError $cloudRawStderr -WindowStyle Hidden -PassThru
        Start-Sleep -Seconds 2
        if ($cloudProcess.HasExited) {
            throw 'cloud-server exited before the health check could run.'
        }
        Set-ReportValue -Name 'CLOUD_SERVER_LOCAL_DB_STARTUP' -Value 'PASSED'

        $healthPassed = $false
        for ($attempt = 1; $attempt -le 45; $attempt++) {
            if ($cloudProcess.HasExited) {
                break
            }
            try {
                $healthResponse = Invoke-WebRequest -UseBasicParsing -Uri ("http://127.0.0.1:{0}/health" -f $httpPort) -TimeoutSec 2
                if ($healthResponse.StatusCode -eq 200) {
                    $healthPassed = $true
                    break
                }
            }
            catch {
                Start-Sleep -Seconds 1
            }
        }
        if (-not $healthPassed) {
            throw "cloud-server health check failed on local port $httpPort."
        }
        Set-ReportValue -Name 'HEALTH_ENDPOINT_PASS' -Value 'PASSED'
    }
    finally {
        Stop-CloudServer
        $serverDatabaseUrl = $null
        Sanitize-CloudLogs
    }

    $report.FINAL_VERDICT = 'LOCAL_E2E_DATABASE_READY'
}
catch {
    if ($collisionDetected) {
        Write-Host 'E2E_DATABASE_NAME_COLLISION=true' -ForegroundColor Red
    }
    Write-Host ("SETUP_BLOCKED=" + $_.Exception.Message) -ForegroundColor Red
}
finally {
    Stop-CloudServer
    Sanitize-CloudLogs
    Restore-ProcessEnvironment

    if ($null -ne $adminPassword) {
        $adminPassword.Dispose()
    }
    if ($null -ne $e2ePassword) {
        $e2ePassword.Dispose()
    }
    $adminPassword = $null
    $e2ePassword = $null
    $adminUsername = $null
}

Write-Host ''
Write-Host 'NEMBUS LOCAL E2E READINESS REPORT'
Write-Host 'nembus_e2e_master = tenant registry/control-plane; no test products are created there.'
Write-Host 'nembus_e2e_tenant = business/RBAC/product/SAP/enrichment/audit database.'
foreach ($reportName in $report.Keys) {
    Write-Host ("{0}={1}" -f $reportName, $report[$reportName])
}
if ($collisionDetected) {
    Write-Host 'E2E_DATABASE_NAME_COLLISION=true'
}

if ($report.FINAL_VERDICT -eq 'LOCAL_E2E_DATABASE_READY') {
    $global:LASTEXITCODE = 0
    return
}
$global:LASTEXITCODE = 1
return
