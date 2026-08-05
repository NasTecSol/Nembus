# Script to generate consolidated Markdown/Text source files for NotebookLM upload

$outDir = "notebook_sources"
if (!(Test-Path $outDir)) {
    New-Item -ItemType Directory -Path $outDir | Out-Null
}

Write-Host "Generating NotebookLM context sources in '$outDir'..." -ForegroundColor Cyan

# -----------------------------------------------------------------------------
# 1. Database Schemas, Migrations & SQLC Configuration Bundle
# -----------------------------------------------------------------------------
$dbContent = @()
$dbContent += "# 1. CANONICAL BASE SCHEMA (packages/core/migrations/000001_base_schema.sql)"
if (Test-Path "packages/core/migrations/000001_base_schema.sql") {
    $dbContent += Get-Content "packages/core/migrations/000001_base_schema.sql" -Raw
}

$dbContent += "`n`n# 2. POS CLIENT BASE MIGRATION (apps/pos-client/migrations/000001_base_schema.sql)"
if (Test-Path "apps/pos-client/migrations/000001_base_schema.sql") {
    $dbContent += Get-Content "apps/pos-client/migrations/000001_base_schema.sql" -Raw
}

$dbContent += "`n`n# 3. POS EXTENSIONS SCHEMA (apps/pos-client/migrations/000002_pos_extensions.sql)"
if (Test-Path "apps/pos-client/migrations/000002_pos_extensions.sql") {
    $dbContent += Get-Content "apps/pos-client/migrations/000002_pos_extensions.sql" -Raw
}

$dbContent += "`n`n# 4. CLOUD SERVER INITIAL SCHEMA (apps/cloud-server/migrations/000001_init_schema.sql)"
if (Test-Path "apps/cloud-server/migrations/000001_init_schema.sql") {
    $dbContent += Get-Content "apps/cloud-server/migrations/000001_init_schema.sql" -Raw
}

$dbContent += "`n`n# 5. SQLC CONFIGURATION (packages/core/sqlc.yaml)"
if (Test-Path "packages/core/sqlc.yaml") {
    $dbContent += Get-Content "packages/core/sqlc.yaml" -Raw
}

$dbContent | Out-File -FilePath "$outDir/01_database_schemas_and_config.md" -Encoding utf8
Write-Host "Created 01_database_schemas_and_config.md" -ForegroundColor Green

# -----------------------------------------------------------------------------
# 2. All SQL Queries Bundle (Core + Cloud Server)
# -----------------------------------------------------------------------------
$queryContent = @()
$queryPaths = @("packages/core/queries", "apps/cloud-server/queries")
$queryCount = 0
foreach ($qDir in $queryPaths) {
    if (Test-Path $qDir) {
        $qFiles = Get-ChildItem -Path $qDir -Filter "*.sql"
        foreach ($f in $qFiles) {
            $queryCount++
            $queryContent += "`n`n--------------------------------------------------"
            $queryContent += "# QUERY FILE: $qDir/$($f.Name)"
            $queryContent += "--------------------------------------------------`n"
            $queryContent += Get-Content $f.FullName -Raw
        }
    }
}
$queryContent | Out-File -FilePath "$outDir/02_all_sql_queries.md" -Encoding utf8
Write-Host "Created 02_all_sql_queries.md ($queryCount query files)" -ForegroundColor Green

# -----------------------------------------------------------------------------
# 3. Core Business Usecases Bundle
# -----------------------------------------------------------------------------
$ucContent = @()
if (Test-Path "packages/core/usecase") {
    $ucFiles = Get-ChildItem -Path "packages/core/usecase" -Filter "*.go"
    foreach ($f in $ucFiles) {
        $ucContent += "`n`n--------------------------------------------------"
        $ucContent += "# USECASE FILE: packages/core/usecase/$($f.Name)"
        $ucContent += "--------------------------------------------------`n"
        $ucContent += Get-Content $f.FullName -Raw
    }
}
$ucContent | Out-File -FilePath "$outDir/03_core_business_usecases.md" -Encoding utf8
Write-Host "Created 03_core_business_usecases.md ($($ucFiles.Count) usecase files)" -ForegroundColor Green

# -----------------------------------------------------------------------------
# 4. Core Middleware, Routing, Handlers & DTOs Bundle
# -----------------------------------------------------------------------------
$hrmContent = @()
$hrmDirs = @("packages/core/middleware", "packages/core/routing", "packages/core/handler")
foreach ($hDir in $hrmDirs) {
    if (Test-Path $hDir) {
        $hFiles = Get-ChildItem -Path $hDir -Recurse -Filter "*.go"
        foreach ($f in $hFiles) {
            $relPath = $f.FullName.Replace((Get-Location).Path + "\", "").Replace("\", "/")
            $hrmContent += "`n`n--------------------------------------------------"
            $hrmContent += "# FILE: $relPath"
            $hrmContent += "--------------------------------------------------`n"
            $hrmContent += Get-Content $f.FullName -Raw
        }
    }
}
$hrmContent | Out-File -FilePath "$outDir/04_core_middleware_routing_dtos.md" -Encoding utf8
Write-Host "Created 04_core_middleware_routing_dtos.md" -ForegroundColor Green

# -----------------------------------------------------------------------------
# 5. Push & Pull Sync Flows, gRPC Protocol & Data Synchronization Bundle
# -----------------------------------------------------------------------------
$syncContent = @()
$syncFileList = @(
    "packages/core/grpc/syncpb/sync.proto",
    "apps/cloud-server/internal/grpc/sync_server.go",
    "apps/cloud-server/internal/grpc/backup_server.go",
    "apps/pos-client/internal/sync/service.go",
    "apps/pos-client/internal/sync/cloner.go",
    "SYNC_CORE_GUIDE.md",
    "apps/cloud-server/scripts/GRPC_SYNC_GUIDE.md",
    "apps/pos-client/scripts/sync_core.ps1"
)
foreach ($fPath in $syncFileList) {
    if (Test-Path $fPath) {
        $syncContent += "`n`n--------------------------------------------------"
        $syncContent += "# SYNC FLOW / PROTO FILE: $fPath"
        $syncContent += "--------------------------------------------------`n"
        $syncContent += Get-Content $fPath -Raw
    }
}
$syncContent | Out-File -FilePath "$outDir/05_push_pull_sync_flows_and_proto.md" -Encoding utf8
Write-Host "Created 05_push_pull_sync_flows_and_proto.md" -ForegroundColor Green

# -----------------------------------------------------------------------------
# 6. Application Entrypoints, Integrations & Frontend Navigation Bundle
# -----------------------------------------------------------------------------
$appsContent = @()
$appFileList = @(
    "go.work",
    "Makefile",
    "Dockerfile",
    "apps/pos-client/main.go",
    "apps/pos-client/app.go",
    "apps/pos-client/internal/db/db_manager.go",
    "apps/pos-client/internal/updater/updater.go",
    "apps/cloud-server/main.go",
    "apps/cloud-server/internal/zatca/service.go",
    "apps/cloud-server/internal/zatca/client.go",
    "apps/pos-client/frontend/ROUTE_MAPPING.md",
    "apps/pos-client/frontend/COMPLETE_ROUTE_PATHS.md"
)
foreach ($fPath in $appFileList) {
    if (Test-Path $fPath) {
        $appsContent += "`n`n--------------------------------------------------"
        $appsContent += "# APP FILE: $fPath"
        $appsContent += "--------------------------------------------------`n"
        $appsContent += Get-Content $fPath -Raw
    }
}
$appsContent | Out-File -FilePath "$outDir/06_apps_entrypoints_integrations_automation.md" -Encoding utf8
Write-Host "Created 06_apps_entrypoints_integrations_automation.md" -ForegroundColor Green

# -----------------------------------------------------------------------------
# 7. Architecture, Compliance & Onboarding Guides Bundle
# -----------------------------------------------------------------------------
$onbContent = @()
$guideFileList = @(
    "README.md",
    "SETUP_GUIDE.md",
    "ZATCA_GUIDE.md",
    "apps/cloud-server/README.md",
    "apps/cloud-server/MIGRATIONS.md",
    "apps/cloud-server/ENVIRONMENTS.md",
    "apps/cloud-server/docs/SWAGGER.md",
    "apps/pos-client/README.md",
    "apps/pos-client/DEVELOPMENT.md"
)
foreach ($gPath in $guideFileList) {
    if (Test-Path $gPath) {
        $onbContent += "`n`n--------------------------------------------------"
        $onbContent += "# ARCHITECTURE / ONBOARDING GUIDE: $gPath"
        $onbContent += "--------------------------------------------------`n"
        $onbContent += Get-Content $gPath -Raw
    }
}
# Also collect docs/onboarding if present
if (Test-Path "docs/onboarding") {
    $onbFiles = Get-ChildItem -Path "docs/onboarding" -Filter "*.md"
    foreach ($f in $onbFiles) {
        $onbContent += "`n`n--------------------------------------------------"
        $onbContent += "# ONBOARDING GUIDE: docs/onboarding/$($f.Name)"
        $onbContent += "--------------------------------------------------`n"
        $onbContent += Get-Content $f.FullName -Raw
    }
}
$onbContent | Out-File -FilePath "$outDir/07_architecture_and_onboarding_guides.md" -Encoding utf8
Write-Host "Created 07_architecture_and_onboarding_guides.md" -ForegroundColor Green

# -----------------------------------------------------------------------------
# 8. Gap Analysis, Deficits & System Roadmap Bundle
# -----------------------------------------------------------------------------
$gapContent = @()
if (Test-Path "gaps") {
    $gapFiles = Get-ChildItem -Path "gaps" -Filter "*.md"
    foreach ($f in $gapFiles) {
        $gapContent += "`n`n--------------------------------------------------"
        $gapContent += "# GAP ANALYSIS & DEFICITS SPEC: gaps/$($f.Name)"
        $gapContent += "--------------------------------------------------`n"
        $gapContent += Get-Content $f.FullName -Raw
    }
}
$gapContent | Out-File -FilePath "$outDir/08_gaps_deficits_and_roadmap.md" -Encoding utf8
Write-Host "Created 08_gaps_deficits_and_roadmap.md" -ForegroundColor Green

Write-Host "`nAll 8 NotebookLM source bundles successfully updated under '$outDir/'!" -ForegroundColor Cyan
