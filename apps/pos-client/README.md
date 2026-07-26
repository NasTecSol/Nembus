# nembus-client

> Wails v2 desktop POS client — part of the NasTecSol NEMBUS ecosystem.

## Repository Map

| Repo | Role | Branch |
|------|------|--------|
| [NasTecSol/nembus-client](https://github.com/NasTecSol/nembus-client) | **This repo** — Wails desktop POS shell | `main` |
| [NasTecSol/Nembus](https://github.com/NasTecSol/Nembus) | Core API / Cloud ERP | `development` (server), `nembus-client` (desktop bridge) |
| [NasTecSol/NPOS-Bofc](https://github.com/NasTecSol/NPOS-Bofc) | Angular frontend | `nembus-client` |

## What This Repo Contains

This repo is the **Wails desktop application shell** for the NEMBUS POS client.

```
nembus-client/
├── main.go              ← Wails entry point (wails.Run)
├── app.go               ← Wails App struct — DB lifecycle, embedded postgres startup
├── wails.json           ← Wails project config
├── printing.go          ← ESC/POS receipt printing helper
├── internal/
│   ├── db/              ← Embedded postgres manager (client-only)
│   ├── sync/            ← DB cloner — syncs cloud → local postgres
│   └── printing/        ← ESC/POS encoder, printer interface, receipt templates
├── client/              ← Backup stub helpers
├── proto/               ← Sync protobuf definitions
├── migrations/          ← SQL schema for the embedded local database
├── frontend/            ← Git submodule → NasTecSol/NPOS-Bofc@nembus-client
└── scripts/             ← Dev utilities
```

Shared backend logic (handlers, usecases, repository, middleware) lives in
[NasTecSol/Nembus](https://github.com/NasTecSol/Nembus) and is referenced via
the `go.mod` replace directive during local development.

## Prerequisites

- **Go** 1.25+
- **Wails v2**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Node.js** 18+ (for Angular frontend)
- **Git** with submodule support

## First-Time Setup

```powershell
# Clone with submodules (gets the Angular frontend too)
git clone --recurse-submodules https://github.com/NasTecSol/nembus-client.git
cd nembus-client

# Also clone the Nembus core (required for local build — see go.work)
git clone --branch nembus-client https://github.com/NasTecSol/Nembus.git ../../nembus-client/Nembus

# Install frontend dependencies
cd frontend
npm install --legacy-peer-deps
cd ..

# Set up environment
Copy-Item .env.dev .env
# Edit .env with your local values

# Verify Wails setup
wails doctor
```

## Development

```powershell
# Run in dev mode (hot reload)
wails dev

# Build production binary
wails build
```

## Updating the Frontend Submodule

```powershell
# Pull latest from NPOS-Bofc@nembus-client
git submodule update --remote --merge frontend
git add frontend
git commit -m "chore: update frontend submodule"
```

## Pulling Core API Updates

```powershell
cd ../../nembus-client/Nembus
git pull origin nembus-client
cd -
```

## Architecture Notes

### Local Dev Bridge
`go.work` (gitignored) makes Go treat both this repo and the Nembus monorepo as
a single workspace. This means imports like `"NEMBUS/internal/handler"` resolve
directly to the local Nembus source — no publish/tag cycle needed during development.

### Future nembus-core Extraction
When `NasTecSol/Nembus` is split into a proper `nembus-core` library repo:
1. Remove the `replace NEMBUS => ...` line from `go.mod`
2. Add `require github.com/NasTecSol/nembus-core v0.x.x`
3. Update import paths from `"NEMBUS/internal/..."` to `"github.com/NasTecSol/nembus-core/..."`

### Frontend Submodule
`frontend/` is a git submodule pointing to `NasTecSol/NPOS-Bofc@nembus-client`.
Wails embeds the compiled Angular output (`frontend/dist/`) into the binary.

## Environment Variables

Copy `.env.dev` to `.env` and configure:

```env
ENV=development
PORT=8080
MASTER_DB_URL=postgresql://postgres:password@localhost:5432/nembus
```
