# Grafana Alloy → Grafana Cloud Observability

Nembus uses **Grafana Cloud** for traces (Tempo), metrics (Prometheus/Mimir), and logs (Loki), fed via a single lightweight **Grafana Alloy** gateway (`deploy/alloy/`).

- **Local development**: Apps export to Alloy on `http://127.0.0.1:4318` (configured in `apps/cloud-server/.env.dev` and `apps/pos-client/.env.dev`), and Alloy forwards batches using gzip compression to your Grafana Cloud OTLP endpoint.
- **EC2 Cloud Server (`pos-api`)**: Exports internally over `nembus-observability-net` to `http://alloy:4318`.
- **Edge Windows Apps (`pos-client`, `sap-agent`, `biometric-agent`)**: Export over HTTPS to `https://otel.nashrms.com` (Nginx validates `OTEL_INGEST_TOKEN` and proxies to Alloy on `127.0.0.1:4318`) with a bounded 50 MB FIFO disk spool for offline resilience.

---

## 1. Collect Credentials from Grafana Cloud

1. Sign in at [grafana.com](https://grafana.com/) → **My Account** → launch your Cloud stack.
2. Open **Connections → Add new connection → OpenTelemetry** (or the **OpenTelemetry / OTLP Configure** tile on your stack home).
3. Copy:
   - **OTLP Endpoint** (e.g. `https://otlp-gateway-prod-us-east-0.grafana.net/otlp`)
   - **Instance ID** (numeric username)
   - Create an **Access Policy Token** with scopes: `metrics:write`, `logs:write`, `traces:write`.

---

## 2. Local Testing on Windows (Before EC2 Deploy)

### A. Smoke-test Alloy → Grafana Cloud

1. Edit `deploy/alloy/.env` and replace the placeholder values with your real `GRAFANA_CLOUD_OTLP_ENDPOINT`, `GRAFANA_CLOUD_INSTANCE_ID`, and `GRAFANA_CLOUD_API_KEY`.
2. Run the smoke-test script from repo root:
   ```powershell
   powershell -ExecutionPolicy Bypass -File deploy\alloy\scripts\test-local.ps1
   ```
   The script starts `grafana/alloy:v1.5.1`, injects synthetic OTLP trace, log, and metric payloads, waits for the 5s batch flush, and checks `http://127.0.0.1:12345/metrics` to verify zero export failures and prints the `traceId` to look up in Grafana Cloud **Explore → Tempo**.

### B. Run Nembus Apps Against Local Alloy

Both `apps/cloud-server/.env.dev` and `apps/pos-client/.env.dev` already include:
```env
OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4318
OTEL_METRIC_EXPORT_INTERVAL=15s
```

Start `cloud-server` (pointing `MASTER_DB_URL` to your local Postgres if not running Docker Compose):
```powershell
cd apps\cloud-server
$env:MASTER_DB_URL = "postgres://root:YOUR_PASS@localhost:5432/masterDB?sslmode=disable"
$env:OTEL_EXPORTER_OTLP_ENDPOINT = "http://127.0.0.1:4318"
$env:OTEL_METRIC_EXPORT_INTERVAL = "15s"
go run main.go dev
```

Hit `http://localhost:8080/health`, then inspect Grafana Cloud:
- **Traces (Tempo)**: `service.name="nembus-cloud-server"` (includes PGX SQL child spans + HTTP server spans)
- **Logs (Loki)**: `{service_name="nembus-cloud-server"}`
- **Metrics (Prometheus)**: `http_server_requests_total` or `nembus_pos_outbox_pending_rows`

> Note: Production binaries default to a **60s** metric export interval to optimize Grafana Cloud active-series/DPM usage; set `OTEL_METRIC_EXPORT_INTERVAL=15s` locally for faster feedback.

---

## 3. Importing Nembus Dashboards into Grafana Cloud

Five Grafana dashboards live in `deploy/alloy/dashboards/`:
- `00-nembus-ecosystem-overview.json`
- `01-nembus-cloud-server.json`
- `02-nembus-pos-client.json`
- `03-nembus-sap-agent.json`
- `04-nembus-biometric-agent.json`

Import them in Grafana Cloud via **Dashboards → New → Import** and select your stack's Prometheus, Loki, and Tempo datasources in the `$ds_prometheus`, `$ds_loki`, and `$ds_tempo` dropdowns.

---

## 4. EC2 Production Setup (`/opt/apps/alloy`)

`deploy-pos-api.yml` automatically syncs `deploy/alloy/` to `/opt/apps/alloy` on EC2. Before the first deploy starts Alloy in production:

1. Create `/opt/apps/alloy/.env` on EC2 with your Grafana Cloud credentials (`GRAFANA_CLOUD_OTLP_ENDPOINT`, `GRAFANA_CLOUD_INSTANCE_ID`, `GRAFANA_CLOUD_API_KEY`).
2. Configure Nginx for edge ingestion (`otel.nashrms.com`) using `deploy/alloy/nginx/otel.conf`:
   ```bash
   sudo mkdir -p /etc/nginx/snippets
   echo 'set $nembus_otel_expected "<STRONG_RANDOM_TOKEN>";' | sudo tee /etc/nginx/snippets/nembus-otel-token.conf
   sudo chmod 640 /etc/nginx/snippets/nembus-otel-token.conf
   sudo cp /opt/apps/alloy/nginx/otel.conf /etc/nginx/sites-available/otel.conf
   sudo ln -sf /etc/nginx/sites-available/otel.conf /etc/nginx/sites-enabled/otel.conf
   sudo nginx -t && sudo systemctl reload nginx
   sudo certbot --nginx -d otel.nashrms.com
   ```
3. Set GitHub Repository Secrets for edge agent releases (`pos-client`, `sap-agent`, `biometric-agent`):
   - `OTEL_ENDPOINT`: `https://otel.nashrms.com`
   - `OTEL_INGEST_TOKEN`: `<STRONG_RANDOM_TOKEN>` (matching `/etc/nginx/snippets/nembus-otel-token.conf`)
4. In `/opt/apps/pos/docker-compose.yml` on EC2, ensure `pos-api` passes `OTEL_EXPORTER_OTLP_ENDPOINT: http://alloy:4318` and `OTEL_SERVICE_NAME: nembus-cloud-server` on network `nembus-observability-net`.
