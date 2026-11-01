# Nembus SAP Agent (`apps/sap-agent`)

On-prem **SAP Business One** edge agent: bulk migration, incremental master-data sync, and POS outbox posting to SAP. All **Nembus tenant data** flows through **cloud gRPC** (`SapAgentService`) and HTTP onboarding — no direct PostgreSQL to the tenant DB from the agent.

## Architecture

- **Local**: `agent_config.json`, SQLite checkpoints, optional SAP company **MSSQL** for ETL extractors, SAP **Service Layer** for incremental sync and posting.
- **Cloud**: `SapAgentService` on cloud-server gRPC port (default `:50051`) — migration ingest, master-data upserts, SAP outbox pull/ack, reconciliation metrics, tenant resolve.
- **Auth**: `Authorization: Bearer <JWT>` or `x-api-key` + `x-tenant-id: <tenant_slug>` on gRPC metadata.

### gRPC (grpcurl example)

```bash
grpcurl -plaintext -H "authorization: Bearer $TOKEN" -H "x-tenant-id: stg" \
  localhost:50051 sapagent.SapAgentService/ResolveTenant
```

Regenerate stubs from [`packages/core/grpc/sapagentpb/sap_agent.proto`](../../packages/core/grpc/sapagentpb/sap_agent.proto):

```bash
cd packages/core/grpc/sapagentpb
protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative sap_agent.proto
```

HTTP `POST /api/v1/migration/batch` remains for legacy agents; set `SAP_MIGRATION_HTTP_ENABLED=false` on cloud-server to disable.

## Configuration

Primary file: **`agent_config.json`** (created on first run). Important fields:

| Section | Purpose |
|---------|---------|
| `tenant_slug` | Nembus tenant slug (same as POS `x-tenant-id`) |
| `cloud.base_url` | Cloud HTTP base (updates, tenant list, login) |
| `cloud.grpc_addr` | Optional gRPC override; default `{host}:50051` from `base_url` |
| `migrator_token` / `cloud.api_key` | gRPC and migration auth |
| `mssql` | SAP B1 company database (migration extractors) |
| `sap_service_layer` | Incremental downstream / upstream SAP SL |
| `sync` | Price list, warehouse, dry-run, upstream poll interval |

Onboarding in the local UI resolves **`organization_id`** via gRPC `ResolveTenant` after login.

## Build & run

```bash
cd apps/sap-agent
go build -o bin/sap-agent.exe cmd/agent/main.go
./bin/sap-agent.exe                    # Windows service + UI (port 17890)
./bin/sap-agent.exe sync-upstream      # One-shot outbox → SAP
./bin/sap-agent.exe test-connection    # SAP SL + cloud gRPC
```

Makefile target: `make build-agent` (builds `cmd/agent/main.go`).

## Capabilities

1. **Migration (SAP MSSQL → cloud)** — ETL domains via `StreamMigrationIngest` (gzip JSON batches).
2. **Downstream (SAP SL → cloud)** — master data via `StreamMasterDataUpsert`.
3. **Upstream (cloud → SAP)** — `StreamSapOutbox` + `AckSapOutbox` / `HoldSapOutbox`.
4. **Reconciliation** — SAP MSSQL counts vs cloud metrics from `GetReconciliationSummary`.

Legacy `.env` under `config/` is deprecated; use `agent_config.json` and CLI subcommands on the unified binary.
