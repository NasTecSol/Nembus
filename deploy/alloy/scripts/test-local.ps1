#Requires -Version 5.1
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
$EnvFile = Join-Path $Root ".env"

if (-not (Test-Path $EnvFile)) {
    Write-Host "Missing $EnvFile. Copy .env.example to .env and fill your Grafana Cloud credentials first." -ForegroundColor Red
    exit 1
}

$envMap = @{}
Get-Content $EnvFile | ForEach-Object {
    $line = $_.Trim()
    if ($line -and -not $line.StartsWith("#") -and $line -match "^([A-Za-z_][A-Za-z0-9_]*)=(.*)$") {
        $envMap[$matches[1]] = $matches[2].Trim()
    }
}

$hasPlaceholders = $false
if (-not $envMap["GRAFANA_CLOUD_OTLP_ENDPOINT"] -or -not $envMap["GRAFANA_CLOUD_INSTANCE_ID"] -or -not $envMap["GRAFANA_CLOUD_API_KEY"]) {
    Write-Host "ERROR: GRAFANA_CLOUD_OTLP_ENDPOINT, GRAFANA_CLOUD_INSTANCE_ID, or GRAFANA_CLOUD_API_KEY is empty in $EnvFile" -ForegroundColor Red
    exit 1
}
if ($envMap["GRAFANA_CLOUD_INSTANCE_ID"] -eq "123456" -or $envMap["GRAFANA_CLOUD_API_KEY"].StartsWith("glc_eyJv...")) {
    $hasPlaceholders = $true
    Write-Host "WARNING: $EnvFile still contains example placeholder credentials." -ForegroundColor Yellow
    Write-Host "         Alloy will accept local OTLP telemetry, but export to Grafana Cloud will fail 401 until replaced." -ForegroundColor Yellow
}

Push-Location $Root
try {
    Write-Host "[1/4] Starting Grafana Alloy container..." -ForegroundColor Cyan
    docker compose up -d
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

    Write-Host "[2/4] Waiting for Alloy readiness at http://127.0.0.1:12345/-/ready ..." -ForegroundColor Cyan
    $ready = $false
    for ($i = 0; $i -lt 20; $i++) {
        try {
            $resp = Invoke-WebRequest -Uri "http://127.0.0.1:12345/-/ready" -UseBasicParsing -TimeoutSec 2
            if ($resp.StatusCode -eq 200) { $ready = $true; break }
        } catch {
            Start-Sleep -Seconds 1
        }
    }
    if (-not $ready) {
        Write-Host "Alloy did not become ready in time. Recent logs:" -ForegroundColor Red
        docker compose logs --tail 30 alloy
        exit 1
    }

    Write-Host "[3/4] Sending synthetic OTLP Trace, Log, and Metric to http://127.0.0.1:4318 ..." -ForegroundColor Cyan
    $traceId = -join ((1..16) | ForEach-Object { "{0:x2}" -f (Get-Random -Maximum 256) })
    $spanId  = -join ((1..8)  | ForEach-Object { "{0:x2}" -f (Get-Random -Maximum 256) })
    $nowNs   = [string]([DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds() * 1000000L)
    $endNs   = [string](([DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds() + 25) * 1000000L)

    $tracePayload = @{
        resourceSpans = @(@{
            resource = @{
                attributes = @(
                    @{ key = "service.name"; value = @{ stringValue = "nembus-smoke-test" } },
                    @{ key = "service.namespace"; value = @{ stringValue = "nembus" } },
                    @{ key = "deployment.environment"; value = @{ stringValue = "development" } },
                    @{ key = "deployment.environment.name"; value = @{ stringValue = "development" } }
                )
            }
            scopeSpans = @(@{
                scope = @{ name = "nembus/smoke" }
                spans = @(@{
                    traceId = $traceId
                    spanId  = $spanId
                    name    = "GET /smoke-test"
                    kind    = 2
                    startTimeUnixNano = $nowNs
                    endTimeUnixNano   = $endNs
                })
            })
        })
    } | ConvertTo-Json -Depth 12 -Compress

    $logPayload = @{
        resourceLogs = @(@{
            resource = @{
                attributes = @(
                    @{ key = "service.name"; value = @{ stringValue = "nembus-smoke-test" } },
                    @{ key = "service.namespace"; value = @{ stringValue = "nembus" } },
                    @{ key = "deployment.environment.name"; value = @{ stringValue = "development" } }
                )
            }
            scopeLogs = @(@{
                scope = @{ name = "nembus/smoke" }
                logRecords = @(@{
                    timeUnixNano   = $nowNs
                    severityNumber = 9
                    severityText   = "INFO"
                    body           = @{ stringValue = "Nembus Alloy smoke test log record" }
                    traceId        = $traceId
                    spanId         = $spanId
                })
            })
        })
    } | ConvertTo-Json -Depth 12 -Compress

    $metricPayload = @{
        resourceMetrics = @(@{
            resource = @{
                attributes = @(
                    @{ key = "service.name"; value = @{ stringValue = "nembus-smoke-test" } },
                    @{ key = "service.namespace"; value = @{ stringValue = "nembus" } },
                    @{ key = "deployment.environment.name"; value = @{ stringValue = "development" } }
                )
            }
            scopeMetrics = @(@{
                scope = @{ name = "nembus/smoke" }
                metrics = @(@{
                    name  = "nembus.smoke.check"
                    gauge = @{
                        dataPoints = @(@{
                            timeUnixNano = $nowNs
                            asInt        = "1"
                        })
                    }
                })
            })
        })
    } | ConvertTo-Json -Depth 12 -Compress

    $rTrace  = Invoke-WebRequest -Method Post -Uri "http://127.0.0.1:4318/v1/traces"  -ContentType "application/json" -Body $tracePayload  -UseBasicParsing
    $rLog    = Invoke-WebRequest -Method Post -Uri "http://127.0.0.1:4318/v1/logs"    -ContentType "application/json" -Body $logPayload    -UseBasicParsing
    $rMetric = Invoke-WebRequest -Method Post -Uri "http://127.0.0.1:4318/v1/metrics" -ContentType "application/json" -Body $metricPayload -UseBasicParsing
    Write-Host ("   Ingest HTTP codes -> traces={0}, logs={1}, metrics={2}, trace_id={3}" -f $rTrace.StatusCode, $rLog.StatusCode, $rMetric.StatusCode, $traceId) -ForegroundColor Green

    Write-Host "[4/4] Waiting 7s for Alloy 5s batch flush and checking Alloy telemetry metrics..." -ForegroundColor Cyan
    Start-Sleep -Seconds 7

    $metricsText = (Invoke-WebRequest -Uri "http://127.0.0.1:12345/metrics" -UseBasicParsing).Content
    $lines = $metricsText -split "`n" | Where-Object {
        $_ -match "^otelcol_receiver_accepted_(spans|log_records|metric_points)" -or
        $_ -match "^otelcol_exporter_sent_(spans|log_records|metric_points)" -or
        $_ -match "^otelcol_exporter_send_failed_(spans|log_records|metric_points)"
    }
    foreach ($line in $lines) {
        Write-Host "   $line"
    }

    $sentSpans = 0.0
    $failedSpans = 0.0
    foreach ($line in $lines) {
        if ($line -match "^otelcol_exporter_sent_spans[^\s]*\s+([0-9eE\.\+\-]+)") { $sentSpans += [double]$matches[1] }
        if ($line -match "^otelcol_exporter_send_failed_spans[^\s]*\s+([0-9eE\.\+\-]+)") { $failedSpans += [double]$matches[1] }
    }

    Write-Host ""
    if ($hasPlaceholders) {
        Write-Host "Local OTLP ingestion into Alloy succeeded, but Grafana Cloud credentials are placeholders." -ForegroundColor Yellow
        Write-Host "Update $EnvFile and re-run this script to verify live Grafana Cloud upload." -ForegroundColor Yellow
    } elseif ($sentSpans -ge 1 -and $failedSpans -eq 0) {
        Write-Host "PASS: Telemetry was accepted by Alloy AND forwarded cleanly to Grafana Cloud!" -ForegroundColor Green
        Write-Host "      Search Tempo in Grafana Cloud for traceId: $traceId (or service.name=nembus-smoke-test)" -ForegroundColor Green
    } else {
        Write-Host "WARN: Check Alloy logs (`docker compose logs alloy`) - sent_spans=$sentSpans, failed_spans=$failedSpans." -ForegroundColor Red
        exit 1
    }
}
finally {
    Pop-Location
}
