<#
.SYNOPSIS
    Nembus E2E Hardening Verification Runner (4-Tier Acceptance Suite).
.DESCRIPTION
    Executes requirement-driven verification in strict DAG order:
      Stage 0: Pre-flight CLI checks & DB URL safety guards
      Stage 1: Schema, LF/no-BOM, Atlas checksum/validate, Goose tags & SQLC freshness
      Stage 2: Angular frontend typecheck, production build (dist/) & specs
      Stage 3: Go vet, build & unit/integration tests across all workspace modules
      Stage 4: PostgreSQL deterministic UUID backfill & upgrade-path tests
      Stage 5: Structured summary table & deterministic exit codes
#>

[CmdletBinding()]
param(
    [string]$TestDatabaseUrl = $env:NEMBUS_SYNC_TEST_DATABASE_URL,
    [string]$BackfillDatabaseUrl = $env:NEMBUS_BACKFILL_TEST_DATABASE_URL,
    [string]$AtlasDevUrl = $env:ATLAS_DEV_URL,
    [switch]$StrictE2E,
    [switch]$SkipSyncMigrations,
    [switch]$SkipFrontend,
    [switch]$SkipBackfill,
    [switch]$RunRace,
    [switch]$RunKarma
)

$ErrorActionPreference = "Stop"
$RepoRoot = (Get-Item (Join-Path $PSScriptRoot "..")).FullName

$StageResults = New-Object System.Collections.Generic.List[object]
$FinalExitCode = 0

function Add-StageResult {
    param(
        [string]$Stage,
        [string]$Status,
        [double]$DurationSeconds,
        [string]$Details = ""
    )
    $StageResults.Add([PSCustomObject]@{
        Stage   = $Stage
        Status  = $Status
        Seconds = [math]::Round($DurationSeconds, 2)
        Details = $Details
    })
}

function Fail-WithCode {
    param(
        [int]$Code,
        [string]$Message
    )
    Write-Host "ERROR (exit $Code): $Message" -ForegroundColor Red
    $script:FinalExitCode = $Code
    throw [System.Exception]::new($Message)
}

function Invoke-Native {
    param(
        [string]$Command,
        [string[]]$Arguments,
        [string]$WorkingDirectory = $RepoRoot,
        [switch]$StreamOutput
    )
    Push-Location $WorkingDirectory
    $prevEAP = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        if ($StreamOutput) {
            & $Command @Arguments
            return [PSCustomObject]@{
                ExitCode = $LASTEXITCODE
                Lines    = @()
            }
        } else {
            $raw = & $Command @Arguments 2>&1
            $code = $LASTEXITCODE
            $lines = @()
            foreach ($item in $raw) {
                $lines += $item.ToString()
            }
            return [PSCustomObject]@{
                ExitCode = $code
                Lines    = $lines
            }
        }
    } finally {
        $ErrorActionPreference = $prevEAP
        Pop-Location
    }
}

function Test-SafeDatabaseUrl {
    param(
        [string]$RawUrl,
        [string]$Label,
        [System.Diagnostics.Stopwatch]$Stopwatch
    )
    if ([string]::IsNullOrWhiteSpace($RawUrl)) {
        return
    }

    $dbHost = ""
    $dbName = ""

    $trimmed = $RawUrl.Trim()
    if ($trimmed -match '^(postgres|postgresql)://') {
        try {
            $uri = [System.Uri]$trimmed
            $dbHost = $uri.Host
            $dbName = $uri.AbsolutePath.TrimStart('/')
        } catch {
            if ($null -ne $Stopwatch) { Add-StageResult "Stage 0: Pre-Flight" "FAIL" $Stopwatch.Elapsed.TotalSeconds "Invalid URL ($Label)" }
            Fail-WithCode 3 "$Label is not a valid PostgreSQL URL."
        }
    } else {
        if ($trimmed -match '(?i)\bhost=([^\s]+)') { $dbHost = $Matches[1] }
        if ($trimmed -match '(?i)\bdbname=([^\s]+)') { $dbName = $Matches[1] }
    }

    $lowerDb = $dbName.ToLowerInvariant()
    $forbidden = @("masterdb", "qitaf", "prod", "production", "staging")
    foreach ($word in $forbidden) {
        if ($lowerDb -eq $word -or $lowerDb.Contains($word)) {
            if ($null -ne $Stopwatch) { Add-StageResult "Stage 0: Pre-Flight" "FAIL" $Stopwatch.Elapsed.TotalSeconds "Forbidden DB name '$dbName'" }
            Fail-WithCode 3 "$Label targets forbidden database name '$dbName' (matched '$word')."
        }
    }

    if (-not [string]::IsNullOrWhiteSpace($dbHost)) {
        $lowerHost = $dbHost.ToLowerInvariant()
        $localHosts = @("localhost", "127.0.0.1", "::1", "postgres")
        if ($localHosts -notcontains $lowerHost -and $env:NEMBUS_ALLOW_REMOTE_TEST_DB -ne "1") {
            if ($null -ne $Stopwatch) { Add-StageResult "Stage 0: Pre-Flight" "FAIL" $Stopwatch.Elapsed.TotalSeconds "Remote host '$dbHost'" }
            Fail-WithCode 3 "$Label targets non-local host '$dbHost'. Set NEMBUS_ALLOW_REMOTE_TEST_DB=1 to override."
        }
    }
}

$EnvNamesToSave = @(
    "GOPRIVATE",
    "GONOSUMCHECK",
    "NEMBUS_SYNC_TEST_DATABASE_URL",
    "NEMBUS_ZATCA_SYNC_TEST_DATABASE_URL",
    "NEMBUS_TEST_DATABASE_URL",
    "NEMBUS_ORDER_TOTALS_TEST_DATABASE_URL",
    "NEMBUS_INVENTORY_TOTALS_TEST_DATABASE_URL",
    "NEMBUS_STOCK_TEST_DATABASE_URL",
    "NEMBUS_ZATCA_TEST_DATABASE_URL",
    "NEMBUS_ZATCA_PRINT_TEST_DATABASE_URL",
    "NEMBUS_REPORT_TEST_DATABASE_URL",
    "NEMBUS_ZATCA_ORDER_TEST_DATABASE_URL",
    "NEMBUS_ZATCA_SANDBOX_TEST",
    "NEMBUS_RESTORE_BACKUP_FILE",
    "NEMBUS_BACKFILL_TEST_DATABASE_URL",
    "NEMBUS_REQUIRE_UUID_MIGRATION",
    "ATLAS_DEV_URL"
)

$SavedEnv = @{}
foreach ($name in $EnvNamesToSave) {
    $SavedEnv[$name] = [Environment]::GetEnvironmentVariable($name, "Process")
}

try {
    # =========================================================================
    # Stage 0: Pre-Flight & Safety Guards
    # =========================================================================
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    Write-Host "=== Stage 0: Pre-Flight & Safety Guards ===" -ForegroundColor Cyan

    $requiredTools = @("go", "atlas", "sqlc", "node", "npm", "npx")
    foreach ($tool in $requiredTools) {
        if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) {
            Add-StageResult "Stage 0: Pre-Flight" "FAIL" $sw.Elapsed.TotalSeconds "Missing CLI: $tool"
            Fail-WithCode 2 "Required CLI tool '$tool' was not found in PATH."
        }
    }

    Test-SafeDatabaseUrl -RawUrl $TestDatabaseUrl -Label "TestDatabaseUrl" -Stopwatch $sw
    Test-SafeDatabaseUrl -RawUrl $BackfillDatabaseUrl -Label "BackfillDatabaseUrl" -Stopwatch $sw
    Test-SafeDatabaseUrl -RawUrl $AtlasDevUrl -Label "AtlasDevUrl" -Stopwatch $sw

    if ([string]::IsNullOrWhiteSpace($BackfillDatabaseUrl) -and -not [string]::IsNullOrWhiteSpace($TestDatabaseUrl)) {
        $BackfillDatabaseUrl = $TestDatabaseUrl
        Write-Host "Info: Reusing TestDatabaseUrl as admin DSN for ephemeral Backfill DBs." -ForegroundColor DarkGray
    }

    if ($StrictE2E -and ([string]::IsNullOrWhiteSpace($TestDatabaseUrl) -or (-not $SkipBackfill -and [string]::IsNullOrWhiteSpace($BackfillDatabaseUrl)))) {
        Add-StageResult "Stage 0: Pre-Flight" "FAIL" $sw.Elapsed.TotalSeconds "StrictE2E requires DB URL"
        Fail-WithCode 4 "-StrictE2E requires -TestDatabaseUrl (and -BackfillDatabaseUrl unless derived or -SkipBackfill)."
    }

    $sw.Stop()
    Add-StageResult "Stage 0: Pre-Flight" "PASS" $sw.Elapsed.TotalSeconds "Tools & DB safety guards OK"

    # =========================================================================
    # Stage 1: Schema & Migration Integrity
    # =========================================================================
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    Write-Host "=== Stage 1: Schema & Migration Integrity ===" -ForegroundColor Cyan

    if (-not $SkipSyncMigrations) {
        $syncScript = Join-Path $RepoRoot "scripts\sync_migrations.ps1"
        $res = Invoke-Native -Command "powershell" -Arguments @("-ExecutionPolicy", "Bypass", "-File", $syncScript) -StreamOutput
        if ($res.ExitCode -ne 0) {
            Add-StageResult "Stage 1: Schema & SQLC" "FAIL" $sw.Elapsed.TotalSeconds "sync_migrations.ps1 failed"
            Fail-WithCode 10 "scripts/sync_migrations.ps1 exited with code $($res.ExitCode)."
        }
    }

    $scanDirs = @(
        "packages\core\db\migrations",
        "apps\cloud-server\migrations",
        "apps\pos-client\migrations",
        "packages\core\db\patches",
        "packages\core\db\views_functions"
    )
    foreach ($relDir in $scanDirs) {
        $fullDir = Join-Path $RepoRoot $relDir
        if (-not (Test-Path $fullDir)) { continue }
        $files = Get-ChildItem -Path $fullDir -File | Where-Object { $_.Extension -eq ".sql" -or $_.Name -eq "atlas.sum" }
        foreach ($f in $files) {
            $bytes = [System.IO.File]::ReadAllBytes($f.FullName)
            if ($bytes.Length -ge 3 -and $bytes[0] -eq 0xEF -and $bytes[1] -eq 0xBB -and $bytes[2] -eq 0xBF) {
                Add-StageResult "Stage 1: Schema & SQLC" "FAIL" $sw.Elapsed.TotalSeconds "UTF-8 BOM in $($f.Name)"
                Fail-WithCode 10 "UTF-8 BOM detected in $($f.FullName)."
            }
            if ([Array]::IndexOf($bytes, [byte]0x0D) -ge 0) {
                Add-StageResult "Stage 1: Schema & SQLC" "FAIL" $sw.Elapsed.TotalSeconds "CRLF in $($f.Name)"
                Fail-WithCode 10 "CR (0x0D) line ending detected in $($f.FullName)."
            }
        }
    }

    $coreSumPath = Join-Path $RepoRoot "packages\core\db\migrations\atlas.sum"
    $serverSumPath = Join-Path $RepoRoot "apps\cloud-server\migrations\atlas.sum"
    $coreHash = (Get-FileHash -Path $coreSumPath -Algorithm SHA256).Hash
    $serverHash = (Get-FileHash -Path $serverSumPath -Algorithm SHA256).Hash
    if ($coreHash -ne $serverHash) {
        Add-StageResult "Stage 1: Schema & SQLC" "FAIL" $sw.Elapsed.TotalSeconds "atlas.sum parity mismatch"
        Fail-WithCode 10 "atlas.sum mismatch between packages/core/db/migrations ($coreHash) and apps/cloud-server/migrations ($serverHash)."
    }

    foreach ($migRel in @("packages/core/db/migrations", "apps/cloud-server/migrations", "apps/pos-client/migrations")) {
        $valRes = Invoke-Native -Command "atlas" -Arguments @("migrate", "validate", "--dir", "file://$migRel")
        if ($valRes.ExitCode -ne 0) {
            $valRes.Lines | ForEach-Object { Write-Host $_ }
            Add-StageResult "Stage 1: Schema & SQLC" "FAIL" $sw.Elapsed.TotalSeconds "atlas validate failed: $migRel"
            Fail-WithCode 10 "atlas migrate validate failed for $migRel."
        }
    }

    # Goose integrity check on apps/pos-client/migrations/*.sql
    $posMigDir = Join-Path $RepoRoot "apps\pos-client\migrations"
    Get-ChildItem -Path $posMigDir -Filter "*.sql" | ForEach-Object {
        $content = [System.IO.File]::ReadAllText($_.FullName)
        if (-not $content.StartsWith("-- +goose Up")) {
            Add-StageResult "Stage 1: Schema & SQLC" "FAIL" $sw.Elapsed.TotalSeconds "Missing -- +goose Up in $($_.Name)"
            Fail-WithCode 10 "Goose migration $($_.Name) does not start with '-- +goose Up'."
        }
        $upOnly = $content
        $downIdx = $upOnly.IndexOf("-- +goose Down")
        if ($downIdx -ge 0) {
            $upOnly = $upOnly.Substring(0, $downIdx)
        }
        $beginCount = ([regex]::Matches($upOnly, '(?m)^--\s*\+goose\s+StatementBegin\b')).Count
        $endCount = ([regex]::Matches($upOnly, '(?m)^--\s*\+goose\s+StatementEnd\b')).Count
        if ($beginCount -ne $endCount) {
            Add-StageResult "Stage 1: Schema & SQLC" "FAIL" $sw.Elapsed.TotalSeconds "Unbalanced StatementBegin/End in $($_.Name)"
            Fail-WithCode 10 "Unbalanced StatementBegin ($beginCount) vs StatementEnd ($endCount) in $($_.Name)."
        }
        $fnCount = ([regex]::Matches($upOnly, '(?mi)^\s*CREATE\s+(?:OR\s+REPLACE\s+)?FUNCTION\b')).Count
        $doCount = ([regex]::Matches($upOnly, '(?mi)^\s*DO\s+\$\$')).Count
        if (($fnCount + $doCount) -gt 0 -and $beginCount -lt 1) {
            Add-StageResult "Stage 1: Schema & SQLC" "FAIL" $sw.Elapsed.TotalSeconds "Missing StatementBegin for PL/pgSQL in $($_.Name)"
            Fail-WithCode 10 "Found CREATE FUNCTION ($fnCount) or DO `$$ ($doCount) without StatementBegin/End in $($_.Name)."
        }
    }

    # SQLC vet, generate (unless -SkipSyncMigrations), and read-only freshness diff
    $sqlcDirs = @(
        @{ Dir = (Join-Path $RepoRoot "packages\core"); RepoRel = "packages/core/repository" },
        @{ Dir = (Join-Path $RepoRoot "apps\cloud-server"); RepoRel = "apps/cloud-server/internal/repository" }
    )
    foreach ($entry in $sqlcDirs) {
        if (-not (Test-Path (Join-Path $entry.Dir "sqlc.yaml"))) { continue }

        $vetRes = Invoke-Native -Command "sqlc" -Arguments @("vet") -WorkingDirectory $entry.Dir
        if ($vetRes.ExitCode -ne 0) {
            $vetRes.Lines | ForEach-Object { Write-Host $_ }
            Add-StageResult "Stage 1: Schema & SQLC" "FAIL" $sw.Elapsed.TotalSeconds "sqlc vet failed in $($entry.Dir)"
            Fail-WithCode 10 "sqlc vet failed in $($entry.Dir)."
        }

        if (-not $SkipSyncMigrations) {
            $genRes = Invoke-Native -Command "sqlc" -Arguments @("generate") -WorkingDirectory $entry.Dir
            if ($genRes.ExitCode -ne 0) {
                $genRes.Lines | ForEach-Object { Write-Host $_ }
                Add-StageResult "Stage 1: Schema & SQLC" "FAIL" $sw.Elapsed.TotalSeconds "sqlc generate failed in $($entry.Dir)"
                Fail-WithCode 10 "sqlc generate failed in $($entry.Dir)."
            }
        }

        $diffRes = Invoke-Native -Command "sqlc" -Arguments @("diff") -WorkingDirectory $entry.Dir
        if ($diffRes.ExitCode -ne 0) {
            $diffRes.Lines | ForEach-Object { Write-Host $_ }
            Add-StageResult "Stage 1: Schema & SQLC" "FAIL" $sw.Elapsed.TotalSeconds "Stale sqlc output in $($entry.RepoRel)"
            Fail-WithCode 10 "SQLC generated code in $($entry.RepoRel) is out of date (sqlc diff returned non-zero)."
        }
    }

    $sw.Stop()
    Add-StageResult "Stage 1: Schema & SQLC" "PASS" $sw.Elapsed.TotalSeconds "LF, no-BOM, Atlas & SQLC clean"

    # =========================================================================
    # Stage 2: Angular Frontend Type-Check, Build & Specs
    # =========================================================================
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    Write-Host "=== Stage 2: Angular Frontend Build & Specs ===" -ForegroundColor Cyan

    $frontendDir = Join-Path $RepoRoot "apps\pos-client\frontend"
    $distDir = Join-Path $frontendDir "dist"

    if ($SkipFrontend) {
        if (-not (Test-Path $distDir) -or (Get-ChildItem -Path $distDir -Recurse -File | Measure-Object).Count -eq 0) {
            Add-StageResult "Stage 2: Frontend" "FAIL" $sw.Elapsed.TotalSeconds "Missing frontend/dist"
            Fail-WithCode 2 "-SkipFrontend passed, but $distDir is missing or empty (required by go:embed)."
        }
        $sw.Stop()
        Add-StageResult "Stage 2: Frontend" "SKIPPED" $sw.Elapsed.TotalSeconds "dist/ exists"
    } else {
        if (-not (Test-Path (Join-Path $frontendDir "node_modules"))) {
            $ciRes = Invoke-Native -Command "npm" -Arguments @("ci") -WorkingDirectory $frontendDir -StreamOutput
            if ($ciRes.ExitCode -ne 0) {
                Add-StageResult "Stage 2: Frontend" "FAIL" $sw.Elapsed.TotalSeconds "npm ci failed"
                Fail-WithCode 20 "npm ci failed in $frontendDir."
            }
        }

        $tscRes = Invoke-Native -Command "npx" -Arguments @("tsc", "--noEmit", "-p", "tsconfig.app.json") -WorkingDirectory $frontendDir -StreamOutput
        if ($tscRes.ExitCode -ne 0) {
            Add-StageResult "Stage 2: Frontend" "FAIL" $sw.Elapsed.TotalSeconds "tsc failed"
            Fail-WithCode 20 "TypeScript typecheck failed in $frontendDir."
        }

        $buildRes = Invoke-Native -Command "npm" -Arguments @("run", "build") -WorkingDirectory $frontendDir -StreamOutput
        if ($buildRes.ExitCode -ne 0) {
            Add-StageResult "Stage 2: Frontend" "FAIL" $sw.Elapsed.TotalSeconds "npm run build failed"
            Fail-WithCode 20 "Angular production build failed in $frontendDir."
        }

        $specRes = Invoke-Native -Command "npm" -Arguments @("run", "test:zatca-dashboard") -WorkingDirectory $frontendDir -StreamOutput
        if ($specRes.ExitCode -ne 0) {
            Add-StageResult "Stage 2: Frontend" "FAIL" $sw.Elapsed.TotalSeconds "test:zatca-dashboard failed"
            Fail-WithCode 20 "ZATCA dashboard specs failed in $frontendDir."
        }

        if ($RunKarma) {
            $karmaRes = Invoke-Native -Command "npm" -Arguments @("test", "--", "--watch=false", "--browsers=ChromeHeadless") -WorkingDirectory $frontendDir -StreamOutput
            if ($karmaRes.ExitCode -ne 0) {
                Add-StageResult "Stage 2: Frontend" "FAIL" $sw.Elapsed.TotalSeconds "Karma tests failed"
                Fail-WithCode 20 "Karma unit tests failed in $frontendDir."
            }
        }

        $sw.Stop()
        Add-StageResult "Stage 2: Frontend" "PASS" $sw.Elapsed.TotalSeconds "tsc, ng build & specs OK"
    }

    # =========================================================================
    # Stage 3: Go Build, Vet & Unit/Integration Tests (DAG Order)
    # =========================================================================
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    Write-Host "=== Stage 3: Go Vet, Build & Tests ===" -ForegroundColor Cyan

    $env:GOPRIVATE = "github.com/NasTecSol/*"
    $env:GONOSUMCHECK = "github.com/NasTecSol/*"
    Remove-Item Env:NEMBUS_ZATCA_SANDBOX_TEST -ErrorAction SilentlyContinue
    Remove-Item Env:NEMBUS_RESTORE_BACKUP_FILE -ErrorAction SilentlyContinue

    if (-not [string]::IsNullOrWhiteSpace($TestDatabaseUrl)) {
        $env:NEMBUS_SYNC_TEST_DATABASE_URL = $TestDatabaseUrl
        $env:NEMBUS_ZATCA_SYNC_TEST_DATABASE_URL = $TestDatabaseUrl
        $env:NEMBUS_TEST_DATABASE_URL = $TestDatabaseUrl
        $env:NEMBUS_ORDER_TOTALS_TEST_DATABASE_URL = $TestDatabaseUrl
        $env:NEMBUS_INVENTORY_TOTALS_TEST_DATABASE_URL = $TestDatabaseUrl
        $env:NEMBUS_STOCK_TEST_DATABASE_URL = $TestDatabaseUrl
        $env:NEMBUS_ZATCA_TEST_DATABASE_URL = $TestDatabaseUrl
        $env:NEMBUS_ZATCA_PRINT_TEST_DATABASE_URL = $TestDatabaseUrl
        $env:NEMBUS_REPORT_TEST_DATABASE_URL = $TestDatabaseUrl
        $env:NEMBUS_ZATCA_ORDER_TEST_DATABASE_URL = $TestDatabaseUrl
    }
    if ($StrictE2E) {
        $env:NEMBUS_BACKFILL_TEST_DATABASE_URL = $BackfillDatabaseUrl
        $env:NEMBUS_REQUIRE_UUID_MIGRATION = "1"
        if (-not [string]::IsNullOrWhiteSpace($AtlasDevUrl)) {
            $env:ATLAS_DEV_URL = $AtlasDevUrl
        }
    }

    $requiredStrictTests = @(
        "TestMultiTerminal",
        "TestCrossTerminal",
        "TestCheckoutParentsBeforePOS",
        "TestCustomerCreationSyncPreservesIdentity",
        "TestSyncServerUpsertReturnsAndMovements",
        "TestCheckoutInventoryExactlyOnce",
        "TestBoundedDeltaPagesWithTimestampTies",
        "TestEnqueuePosReceiptIdempotent",
        "TestTransactionParentsForExistingQueueEntry",
        "TestDeltaDataAndWatermarkCommitTogether",
        "TestCustomerPullDoesNotOverwriteUnsentLocalCreation",
        "TestZatcaOutboxCloudAndCloneRecoveryPostgres",
        "TestZReportPostgres",
        "TestUUIDBackfillAndUpgradePaths"
    )

    $goModules = @(
        "packages\sap",
        "packages\core",
        "apps\cloud-server",
        "apps\pos-client",
        "apps\sap-agent",
        "apps\biometric-agent"
    )

    foreach ($modRel in $goModules) {
        $modDir = Join-Path $RepoRoot $modRel
        if (-not (Test-Path (Join-Path $modDir "go.mod"))) { continue }
        Write-Host "--- Go Module: $modRel ---" -ForegroundColor Yellow

        $vetRes = Invoke-Native -Command "go" -Arguments @("vet", "./...") -WorkingDirectory $modDir -StreamOutput
        if ($vetRes.ExitCode -ne 0) {
            Add-StageResult "Stage 3: Go Modules" "FAIL" $sw.Elapsed.TotalSeconds "go vet failed in $modRel"
            Fail-WithCode 30 "go vet ./... failed in $modRel."
        }

        $buildRes = Invoke-Native -Command "go" -Arguments @("build", "./...") -WorkingDirectory $modDir -StreamOutput
        if ($buildRes.ExitCode -ne 0) {
            Add-StageResult "Stage 3: Go Modules" "FAIL" $sw.Elapsed.TotalSeconds "go build failed in $modRel"
            Fail-WithCode 30 "go build ./... failed in $modRel."
        }

        if ($StrictE2E) {
            $testRes = Invoke-Native -Command "go" -Arguments @("test", "-json", "-count=1", "./...") -WorkingDirectory $modDir
            $skippedRequired = New-Object System.Collections.Generic.List[string]
            foreach ($line in $testRes.Lines) {
                if ($line.StartsWith("{")) {
                    try {
                        $ev = $line | ConvertFrom-Json
                        if ($null -ne $ev.Output) {
                            Write-Host ($ev.Output.TrimEnd("`r", "`n"))
                        }
                        if ($ev.Action -eq "skip" -and -not [string]::IsNullOrWhiteSpace($ev.Test)) {
                            foreach ($req in $requiredStrictTests) {
                                if ($ev.Test -like "*$req*") {
                                    $skippedRequired.Add("$($ev.Package)::$($ev.Test)")
                                }
                            }
                        }
                    } catch {
                        Write-Host $line
                    }
                } else {
                    Write-Host $line
                }
            }
            if ($testRes.ExitCode -ne 0) {
                Add-StageResult "Stage 3: Go Modules" "FAIL" $sw.Elapsed.TotalSeconds "go test failed in $modRel"
                Fail-WithCode 30 "go test ./... failed in $modRel."
            }
            if ($skippedRequired.Count -gt 0) {
                $joined = ($skippedRequired | Select-Object -Unique) -join ", "
                Add-StageResult "Stage 3: Go Modules" "FAIL" $sw.Elapsed.TotalSeconds "StrictE2E skipped required tests"
                Fail-WithCode 4 "StrictE2E detected skipped required DB tests: $joined"
            }
        } else {
            $testRes = Invoke-Native -Command "go" -Arguments @("test", "-count=1", "./...") -WorkingDirectory $modDir -StreamOutput
            if ($testRes.ExitCode -ne 0) {
                Add-StageResult "Stage 3: Go Modules" "FAIL" $sw.Elapsed.TotalSeconds "go test failed in $modRel"
                Fail-WithCode 30 "go test ./... failed in $modRel."
            }
        }
    }

    if ($RunRace) {
        $raceTargets = @(
            @{ Dir = (Join-Path $RepoRoot "apps\cloud-server"); Pkg = "./internal/grpc/..." },
            @{ Dir = (Join-Path $RepoRoot "apps\pos-client"); Pkg = "./internal/sync/..." }
        )
        foreach ($rt in $raceTargets) {
            Write-Host "--- Race Check: $($rt.Dir) $($rt.Pkg) ---" -ForegroundColor Yellow
            $raceRes = Invoke-Native -Command "go" -Arguments @("test", "-race", "-count=1", $rt.Pkg) -WorkingDirectory $rt.Dir -StreamOutput
            if ($raceRes.ExitCode -ne 0) {
                Add-StageResult "Stage 3: Go Modules" "FAIL" $sw.Elapsed.TotalSeconds "go test -race failed in $($rt.Pkg)"
                Fail-WithCode 30 "go test -race failed in $($rt.Dir) ($($rt.Pkg))."
            }
        }
    }

    $sw.Stop()
    Add-StageResult "Stage 3: Go Modules" "PASS" $sw.Elapsed.TotalSeconds "All Go modules vet/build/test OK"

    # =========================================================================
    # Stage 4: Deterministic Backfill & Upgrade Paths PG18 Test
    # =========================================================================
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    Write-Host "=== Stage 4: Deterministic UUID Backfill PG18 Test ===" -ForegroundColor Cyan

    if ($SkipBackfill) {
        $sw.Stop()
        Add-StageResult "Stage 4: Backfill PG18" "SKIPPED" $sw.Elapsed.TotalSeconds "-SkipBackfill passed"
    } elseif ([string]::IsNullOrWhiteSpace($BackfillDatabaseUrl)) {
        if ($StrictE2E) {
            Add-StageResult "Stage 4: Backfill PG18" "FAIL" $sw.Elapsed.TotalSeconds "No BackfillDatabaseUrl"
            Fail-WithCode 4 "-StrictE2E requires -BackfillDatabaseUrl (or -TestDatabaseUrl)."
        }
        Write-Host "WARN: Backfill DB test skipped (pass -BackfillDatabaseUrl or set NEMBUS_BACKFILL_TEST_DATABASE_URL)." -ForegroundColor Yellow
        $sw.Stop()
        Add-StageResult "Stage 4: Backfill PG18" "SKIPPED" $sw.Elapsed.TotalSeconds "No DB URL provided"
    } else {
        $env:NEMBUS_BACKFILL_TEST_DATABASE_URL = $BackfillDatabaseUrl
        if (-not [string]::IsNullOrWhiteSpace($AtlasDevUrl)) {
            $env:ATLAS_DEV_URL = $AtlasDevUrl
        }
        if ($StrictE2E) {
            $env:NEMBUS_REQUIRE_UUID_MIGRATION = "1"
        }

        $cloudDir = Join-Path $RepoRoot "apps\cloud-server"
        $bfRes = Invoke-Native -Command "go" -Arguments @("test", "-json", "-v", "-count=1", "./cmd/migrate-tenants", "-run", "TestUUIDBackfillAndUpgradePaths") -WorkingDirectory $cloudDir
        $bfSkipped = $false
        foreach ($line in $bfRes.Lines) {
            if ($line.StartsWith("{")) {
                try {
                    $ev = $line | ConvertFrom-Json
                    if ($null -ne $ev.Output) {
                        Write-Host ($ev.Output.TrimEnd("`r", "`n"))
                    }
                    if ($ev.Action -eq "skip" -and $ev.Test -like "TestUUIDBackfillAndUpgradePaths*") {
                        $bfSkipped = $true
                    }
                } catch {
                    Write-Host $line
                }
            } else {
                Write-Host $line
            }
        }
        if ($bfRes.ExitCode -ne 0) {
            Add-StageResult "Stage 4: Backfill PG18" "FAIL" $sw.Elapsed.TotalSeconds "Backfill test failed"
            Fail-WithCode 40 "TestUUIDBackfillAndUpgradePaths failed."
        }
        if ($StrictE2E -and $bfSkipped) {
            Add-StageResult "Stage 4: Backfill PG18" "FAIL" $sw.Elapsed.TotalSeconds "Backfill test skipped in StrictE2E"
            Fail-WithCode 4 "TestUUIDBackfillAndUpgradePaths reported SKIP under -StrictE2E."
        }

        $sw.Stop()
        if ($bfSkipped) {
            Add-StageResult "Stage 4: Backfill PG18" "SKIPPED" $sw.Elapsed.TotalSeconds "Pending M1 migration"
        } else {
            Add-StageResult "Stage 4: Backfill PG18" "PASS" $sw.Elapsed.TotalSeconds "Paths A, B, C verified"
        }
    }

} catch {
    if ($FinalExitCode -eq 0) {
        Write-Host "Unexpected error: $_" -ForegroundColor Red
        $FinalExitCode = 1
    }
} finally {
    foreach ($name in $EnvNamesToSave) {
        $orig = $SavedEnv[$name]
        if ($null -eq $orig) {
            Remove-Item "Env:$name" -ErrorAction SilentlyContinue
        } else {
            [Environment]::SetEnvironmentVariable($name, $orig, "Process")
        }
    }

    Write-Host ""
    Write-Host "=== Stage 5: E2E Verification Summary ===" -ForegroundColor Cyan
    if ($StageResults.Count -gt 0) {
        $StageResults | Format-Table -AutoSize | Out-String | Write-Host
    }
    Write-Host "Final Exit Code: $FinalExitCode" -ForegroundColor $(if ($FinalExitCode -eq 0) { "Green" } else { "Red" })
}

exit $FinalExitCode
