# Nembus Monorepo

Welcome to the **Nembus Monorepo**, uniting the Cloud ERP Backend service and the POS Desktop client application into a single, high-performance Go Workspace.

---

## Workspace Architecture

```
nembus-monorepo/
├── go.work                         # Go Workspace definition
├── Makefile                        # Root automation task runner
├── Dockerfile                      # Production build container for cloud deployment
├── .gitmodules                     # Submodule configuration for POS Angular frontend
│
├── packages/
│   └── core/                       # Shared domain module: github.com/NasTecSol/nembus-core
│       ├── migrations/             # Canonical 000001_base_schema.sql
│       ├── queries/                # 60 SQL query files (sqlc source)
│       ├── repository/             # SQLC-generated models & queries
│       ├── config/                 # Base configuration struct
│       ├── usecase/                # 39 domain business logic usecases
│       ├── handler/                # 37 REST API handlers
│       ├── middleware/             # JWT, Tenant, Logger middleware
│       ├── routing/                # Router initialization helpers
│       ├── printing/               # Thermal printer ESC/POS engine
│       └── grpc/backuppb/          # gRPC Protobuf definitions
│
└── apps/
    ├── cloud-server/               # Cloud ERP Server: github.com/NasTecSol/nembus-server
    │   ├── main.go                 # Cloud server entry point (HTTP + gRPC)
    │   ├── internal/grpc/          # gRPC backup server service
    │   ├── docs/                   # Swagger OpenAPI definitions
    │   └── cmd/                    # Cloud CLI utilities (backup-client, etc.)
    │
    └── pos-client/                 # POS Desktop App: github.com/NasTecSol/nembus-client
        ├── main.go                 # Wails desktop app entry point
        ├── app.go                  # Embedded PostgreSQL & sync engine lifecycle
        ├── wails.json              # Wails desktop project configuration
        ├── migrations/             # 000001_base_schema.sql + 000002_pos_extensions.sql
        ├── frontend/               # Angular GUI (Git Submodule: NasTecSol/NPOS-Bofc)
        └── scripts/sync_core.ps1   # Core schema sync utility
```

---

## Quick Reference Commands

All commands can be run via `make` from the monorepo root or using `go` / `wails` directly:

| Action | Makefile Command | Direct Command |
|:---|:---|:---|
| **Build Cloud Server** | `make build-server` | `cd apps/cloud-server && go build -o bin/nembus-server main.go` |
| **Build POS Client (Exe)** | `make build-client` | `cd apps/pos-client && wails build` |
| **Run Cloud Server (Dev)** | `make dev-server` | `cd apps/cloud-server && go run main.go dev` |
| **Run POS Client (Dev)** | `make dev-client` | `cd apps/pos-client && wails dev` |
| **Sync Core Schema** | `make sync-core` | `cd apps/pos-client && powershell -ExecutionPolicy Bypass -File scripts/sync_core.ps1` |
| **Tidy Workspace** | `make tidy` | `go work sync` |

---

## Documentation Links

- **[Developer Setup Guide](file:///C:/Users/pc/Development/NASHR/BackEnd/nembus-monorepo/SETUP_GUIDE.md)**: Complete guide to installing dependencies, building the Wails desktop app on Windows, and running local development environments.
- **[Core Schema Sync Guide](file:///C:/Users/pc/Development/NASHR/BackEnd/nembus-monorepo/SYNC_CORE_GUIDE.md)**: Detailed instructions on propagating core schema and query updates from `packages/core` to `apps/pos-client`.
