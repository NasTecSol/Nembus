# Script to generate consolidated Markdown/Text source files for NotebookLM upload

$outDir = "notebook_sources"
if (!(Test-Path $outDir)) {
    New-Item -ItemType Directory -Path $outDir | Out-Null
}

Write-Host "Generating NotebookLM context sources in '$outDir'..."

# 1. Database Schema and SQLC Configuration
$dbContent = @()
$dbContent += "# 1. CANONICAL BASE SCHEMA (packages/core/migrations/000001_base_schema.sql)"
if (Test-Path "packages/core/migrations/000001_base_schema.sql") {
    $dbContent += Get-Content "packages/core/migrations/000001_base_schema.sql" -Raw
}

$dbContent += "`n`n# 2. POS EXTENSIONS SCHEMA (apps/pos-client/migrations/000002_pos_extensions.sql)"
if (Test-Path "apps/pos-client/migrations/000002_pos_extensions.sql") {
    $dbContent += Get-Content "apps/pos-client/migrations/000002_pos_extensions.sql" -Raw
}

$dbContent += "`n`n# 3. SQLC CONFIGURATION (packages/core/sqlc.yaml)"
if (Test-Path "packages/core/sqlc.yaml") {
    $dbContent += Get-Content "packages/core/sqlc.yaml" -Raw
}

$dbContent | Out-File -FilePath "$outDir/01_database_schemas_and_config.md" -Encoding utf8
Write-Host "Created 01_database_schemas_and_config.md"

# 2. All 60 SQL Queries Bundle
$queryContent = @()
$queryFiles = Get-ChildItem -Path "packages/core/queries" -Filter "*.sql"
foreach ($f in $queryFiles) {
    $queryContent += "`n`n--------------------------------------------------"
    $queryContent += "# QUERY FILE: packages/core/queries/$($f.Name)"
    $queryContent += "--------------------------------------------------`n"
    $queryContent += Get-Content $f.FullName -Raw
}
$queryContent | Out-File -FilePath "$outDir/02_all_sql_queries.md" -Encoding utf8
Write-Host "Created 02_all_sql_queries.md (60 query files)"

# 3. Core Business Usecases Bundle
$ucContent = @()
$ucFiles = Get-ChildItem -Path "packages/core/usecase" -Filter "*.go"
foreach ($f in $ucFiles) {
    $ucContent += "`n`n--------------------------------------------------"
    $ucContent += "# USECASE FILE: packages/core/usecase/$($f.Name)"
    $ucContent += "--------------------------------------------------`n"
    $ucContent += Get-Content $f.FullName -Raw
}
$ucContent | Out-File -FilePath "$outDir/03_core_business_usecases.md" -Encoding utf8
Write-Host "Created 03_core_business_usecases.md (39 usecase files)"

# 4. Core Middleware, Routing, & DTOs Bundle
$hrmContent = @()
$hrmFiles = Get-ChildItem -Path "packages/core/middleware","packages/core/routing" -Recurse -Filter "*.go"
foreach ($f in $hrmFiles) {
    $relPath = $f.FullName.Replace((Get-Location).Path + "\", "")
    $hrmContent += "`n`n--------------------------------------------------"
    $hrmContent += "# FILE: $relPath"
    $hrmContent += "--------------------------------------------------`n"
    $hrmContent += Get-Content $f.FullName -Raw
}
$dtoFile = "packages/core/handler/dto.go"
if (Test-Path $dtoFile) {
    $hrmContent += "`n`n--------------------------------------------------"
    $hrmContent += "# HANDLER DTOS: packages/core/handler/dto.go"
    $hrmContent += "--------------------------------------------------`n"
    $hrmContent += Get-Content $dtoFile -Raw
}
$hrmContent | Out-File -FilePath "$outDir/04_core_middleware_routing_dtos.md" -Encoding utf8
Write-Host "Created 04_core_middleware_routing_dtos.md"

# 5. Application Entrypoints, Integrations & Tooling Bundle
$appsContent = @()
$appFileList = @(
    "go.work",
    "Makefile",
    "apps/pos-client/scripts/sync_core.ps1",
    "apps/pos-client/main.go",
    "apps/pos-client/app.go",
    "apps/pos-client/internal/db/db_manager.go",
    "apps/cloud-server/main.go",
    "apps/cloud-server/internal/grpc/backup_server.go"
)
foreach ($fPath in $appFileList) {
    if (Test-Path $fPath) {
        $appsContent += "`n`n--------------------------------------------------"
        $appsContent += "# APP FILE: $fPath"
        $appsContent += "--------------------------------------------------`n"
        $appsContent += Get-Content $fPath -Raw
    }
}
$appsContent | Out-File -FilePath "$outDir/05_apps_entrypoints_integrations_automation.md" -Encoding utf8
Write-Host "Created 05_apps_entrypoints_integrations_automation.md"

# 6. Architecture & Onboarding Guides Bundle
$onbContent = @()
$onbFiles = Get-ChildItem -Path "docs/onboarding" -Filter "*.md"
foreach ($f in $onbFiles) {
    $onbContent += "`n`n--------------------------------------------------"
    $onbContent += "# ONBOARDING GUIDE: docs/onboarding/$($f.Name)"
    $onbContent += "--------------------------------------------------`n"
    $onbContent += Get-Content $f.FullName -Raw
}
$onbContent | Out-File -FilePath "$outDir/06_architecture_and_onboarding_guides.md" -Encoding utf8
Write-Host "Created 06_architecture_and_onboarding_guides.md"

# 7. Gap Analysis & Feature Completeness Bundle
$gapContent = @()
if (Test-Path "gaps") {
    $gapFiles = Get-ChildItem -Path "gaps" -Filter "*.md"
    foreach ($f in $gapFiles) {
        $gapContent += "`n`n--------------------------------------------------"
        $gapContent += "# GAP ANALYSIS SPEC: gaps/$($f.Name)"
        $gapContent += "--------------------------------------------------`n"
        $gapContent += Get-Content $f.FullName -Raw
    }
    $gapContent | Out-File -FilePath "$outDir/07_gap_analysis_and_completeness_plan.md" -Encoding utf8
    Write-Host "Created 07_gap_analysis_and_completeness_plan.md"
}

Write-Host "`nAll NotebookLM source bundles successfully created under 'notebook_sources/'!"

