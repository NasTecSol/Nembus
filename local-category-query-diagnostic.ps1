Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# This diagnostic intentionally has no parameters or environment-based target overrides.
# It can connect only to the local E2E tenant named below.
$targetHost = '127.0.0.1'
$targetPort = 5432
$targetDatabase = 'nembus_e2e_tenant'
$targetUser = 'nembus_e2e_user'

if ($targetHost -ne '127.0.0.1' -or
    $targetPort -ne 5432 -or
    $targetDatabase -ne 'nembus_e2e_tenant' -or
    $targetUser -ne 'nembus_e2e_user') {
    throw 'Safety gate rejected a database target other than 127.0.0.1:5432 / nembus_e2e_tenant / nembus_e2e_user.'
}

function Get-SanitizedDiagnostic {
    param([AllowEmptyString()][string]$Text)

    $sanitized = $Text -replace '(?i)(postgres(?:ql)?://[^\s:@/]+):[^@\s]+@', '$1:[REDACTED]@'
    $sanitized = $sanitized -replace '(?im)((?:password|pass|pwd)\s*=\s*)\S+', '$1[REDACTED]'
    $sanitized = $sanitized.Trim()

    if ([string]::IsNullOrWhiteSpace($sanitized)) {
        return 'psql returned no PostgreSQL diagnostic text.'
    }
    if ($sanitized.Length -gt 12000) {
        return $sanitized.Substring(0, 12000) + "`n[diagnostic truncated]"
    }
    return $sanitized
}

$psqlCommand = Get-Command psql -ErrorAction SilentlyContinue
if ($null -eq $psqlCommand -or $psqlCommand.CommandType -ne 'Application') {
    throw 'psql was not found. Install PostgreSQL client tools or add the PostgreSQL bin directory to PATH, then rerun this script.'
}
$psqlPath = $psqlCommand.Source

$temporaryDirectory = Join-Path ([System.IO.Path]::GetTempPath()) ('nembus-category-query-' + [Guid]::NewGuid().ToString('N'))
$securePassword = $null
$passwordBstr = [IntPtr]::Zero

function Invoke-LocalE2EPsql {
    param(
        [Parameter(Mandatory = $true)][string]$Sql,
        [Parameter(Mandatory = $true)][string]$Label
    )

    $sqlPath = Join-Path $temporaryDirectory ($Label + '.sql')
    $stdoutPath = Join-Path $temporaryDirectory ($Label + '.stdout')
    $stderrPath = Join-Path $temporaryDirectory ($Label + '.stderr')
    [System.IO.File]::WriteAllText($sqlPath, $Sql, [System.Text.UTF8Encoding]::new($false))

    $arguments = @(
        '--no-psqlrc',
        '--quiet',
        '--tuples-only',
        '--no-align',
        '--set', 'ON_ERROR_STOP=1',
        '--host', $targetHost,
        '--port', $targetPort.ToString(),
        '--username', $targetUser,
        '--dbname', $targetDatabase,
        '--file', $sqlPath
    )

    & $psqlPath @arguments 1> $stdoutPath 2> $stderrPath
    $exitCode = $LASTEXITCODE
    $stdout = if (Test-Path -LiteralPath $stdoutPath) { [System.IO.File]::ReadAllText($stdoutPath) } else { '' }
    $stderr = if (Test-Path -LiteralPath $stderrPath) { [System.IO.File]::ReadAllText($stderrPath) } else { '' }

    return [PSCustomObject]@{
        ExitCode = $exitCode
        StdOut = $stdout
        StdErr = $stderr
    }
}

try {
    New-Item -ItemType Directory -Path $temporaryDirectory -ErrorAction Stop | Out-Null

    $securePassword = Read-Host -AsSecureString -Prompt 'nembus_e2e_user PostgreSQL password'
    $passwordBstr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($securePassword)
    [Environment]::SetEnvironmentVariable('PGPASSWORD', [Runtime.InteropServices.Marshal]::PtrToStringBSTR($passwordBstr), 'Process')

    Write-Output 'DATABASE_TARGET=127.0.0.1:5432 / nembus_e2e_tenant only'

    $login = Invoke-LocalE2EPsql -Label 'login' -Sql @'
BEGIN READ ONLY;
SELECT 1;
COMMIT;
'@
    if ($login.ExitCode -ne 0) {
        Write-Output 'DATABASE_LOGIN=FAILED'
        Write-Output 'CATEGORY_QUERY_RESULT=FAILED'
        Write-Output 'POSTGRES_ERROR_BEGIN'
        Write-Output (Get-SanitizedDiagnostic -Text $login.StdErr)
        Write-Output 'POSTGRES_ERROR_END'
        exit 1
    }
    Write-Output 'DATABASE_LOGIN=PASS'

    $activeCount = Invoke-LocalE2EPsql -Label 'active-category-count' -Sql @'
BEGIN READ ONLY;
SELECT count(*) FROM product_categories WHERE is_active = true;
COMMIT;
'@
    if ($activeCount.ExitCode -eq 0 -and $activeCount.StdOut.Trim() -match '^\d+$') {
        Write-Output ('E2E_ACTIVE_CATEGORY_COUNT=' + $activeCount.StdOut.Trim())
    }
    else {
        Write-Output 'E2E_ACTIVE_CATEGORY_COUNT=UNAVAILABLE'
    }

    # This is the SQLC-generated GetCategoryHierarchy CTE. The ProductEnrichmentStore
    # call GetCategoryHierarchy(ctx, true) binds $1 to boolean true; this diagnostic
    # substitutes that identical literal. The outer count prevents category data dumps.
    $categoryQuery = @'
BEGIN READ ONLY;
WITH RECURSIVE category_tree AS (
    SELECT
        id, parent_category_id, name, code, description,
        category_level, is_active, metadata,
        1 as level,
        ARRAY[id] as path,
        name::text as full_path
    FROM product_categories
    WHERE parent_category_id IS NULL

    UNION ALL

    SELECT
        pc.id, pc.parent_category_id, pc.name, pc.code, pc.description,
        pc.category_level, pc.is_active, pc.metadata,
        ct.level + 1,
        ct.path || pc.id,
        ct.full_path || ' > ' || pc.name
    FROM product_categories pc
    INNER JOIN category_tree ct ON pc.parent_category_id = ct.id
)
SELECT count(*) AS category_rows_returned
FROM (
    SELECT id, parent_category_id, name, code, description, category_level, is_active, metadata, level, path, full_path FROM category_tree ct
    WHERE CASE
        WHEN true IS NULL THEN true
        ELSE ct.is_active = true
    END
    ORDER BY ct.path
) AS get_category_hierarchy;
COMMIT;
'@

    $categoryResult = Invoke-LocalE2EPsql -Label 'get-category-hierarchy' -Sql $categoryQuery
    if ($categoryResult.ExitCode -ne 0) {
        $diagnostic = if ([string]::IsNullOrWhiteSpace($categoryResult.StdErr)) { $categoryResult.StdOut } else { $categoryResult.StdErr }
        Write-Output 'CATEGORY_QUERY_RESULT=FAILED'
        Write-Output 'POSTGRES_ERROR_BEGIN'
        Write-Output (Get-SanitizedDiagnostic -Text $diagnostic)
        Write-Output 'POSTGRES_ERROR_END'
        exit 1
    }

    $rowsReturned = $categoryResult.StdOut.Trim()
    if ($rowsReturned -notmatch '^\d+$') {
        Write-Output 'CATEGORY_QUERY_RESULT=FAILED'
        Write-Output 'POSTGRES_ERROR_BEGIN'
        Write-Output 'The category query completed, but psql returned an unexpected row-count format.'
        Write-Output 'POSTGRES_ERROR_END'
        exit 1
    }

    Write-Output 'CATEGORY_QUERY_RESULT=PASS'
    Write-Output ('CATEGORY_ROWS_RETURNED=' + $rowsReturned)
}
finally {
    [Environment]::SetEnvironmentVariable('PGPASSWORD', $null, 'Process')
    if ($passwordBstr -ne [IntPtr]::Zero) {
        [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($passwordBstr)
    }
    if ($null -ne $securePassword) {
        $securePassword.Dispose()
    }
    if (Test-Path -LiteralPath $temporaryDirectory) {
        Remove-Item -LiteralPath $temporaryDirectory -Recurse -Force
    }
}
