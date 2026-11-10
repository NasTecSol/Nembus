# Local Development Setup

This file documents the steps for a new developer (or AI agent) to set up the full
local development environment for the `nembus-client` desktop POS app.

---

## Repository Structure

```
nembus-client/                         ← This repo (NasTecSol/nembus-client)
├── main.go                            ← Wails entry point
├── app.go                             ← Wails App struct + embedded postgres lifecycle
├── wails.json                         ← Wails project config
├── printing.go                        ← ESC/POS helper
├── internal/
│   ├── db/db_manager.go               ← Embedded postgres manager
│   ├── sync/{cloner,service}.go       ← Cloud → local DB sync service
│   └── printing/{escpos,printer,template}.go
├── client/backup_stub.go
├── proto/sync.proto
├── migrations/000001_init_schema.sql  ← Schema for local embedded postgres
├── frontend/                          ← git submodule → NasTecSol/NPOS-Bofc@nembus-client
├── go.mod                             ← module: github.com/NasTecSol/nembus-client
├── go.work                            ← GITIGNORED — created manually by each dev
└── README.md
```

## Step 1 — Clone this repo

```powershell
git clone --recurse-submodules https://github.com/NasTecSol/nembus-client.git
cd nembus-client
```

> The `--recurse-submodules` flag automatically clones `frontend/`
> (NasTecSol/NPOS-Bofc on branch `nembus-client`).

---

## Step 2 — Clone the Nembus core (required for local build)

The `go.mod` file has a `replace NEMBUS => ../../nembus-client/Nembus` directive.
This path assumes the Nembus monorepo is cloned **two directories up**, at:
```
c:\Users\pc\Development\nembus-client\Nembus\
```

If you are on a different machine, clone Nembus at a sibling path and update `go.work`:

```powershell
# Clone Nembus on the nembus-client branch (the desktop-compatible branch)
git clone --branch nembus-client https://github.com/NasTecSol/Nembus.git ..\Nembus
```

---

## Step 3 — Create go.work (NOT committed — each dev does this once)

```powershell
# From the nembus-client repo root:
@"
go 1.25

use (
    .
    ..\Nembus
)
"@ | Set-Content go.work
```

Adjust the path ``..\Nembus`` to match wherever you cloned the Nembus repo.

> **Why go.work?** The `replace NEMBUS => ...` in `go.mod` tells Go where to find the
> shared packages. `go.work` tells the Go workspace to treat both repos as one unit,
> so IDE autocomplete and `go build` work without publishing to GitHub.

---

## Step 4 — Install frontend dependencies

```powershell
cd frontend
npm install --legacy-peer-deps
cd ..
```

---

## Step 5 — Configure environment

```powershell
Copy-Item .env.dev .env
```

Edit `.env` with your local values:
```
ENV=development
PORT=8080
MASTER_DB_URL=postgresql://postgres:password@localhost:5432/nembus
```

---

## Step 6 — Verify Wails

```powershell
wails doctor
```

---

## Step 7 — Run in dev mode

```powershell
wails dev
```

### Starting Fresh / Cleaning Caches

If you face build caching issues (e.g., Angular templates not updating, frontend-to-backend Wails JS bindings being out of sync, or embedded PostgreSQL runtime locks), you can clean all build caches and temporary folders:

```powershell
.\clean.ps1
```

For a deeper clean (removing `node_modules`, cleaning Go module cache, and resetting the local embedded database completely to start fresh), run:

```powershell
.\clean.ps1 -Deep -ResetDb
```

---

## Updating the Frontend Submodule

When the Angular frontend (NPOS-Bofc) has new commits on `nembus-client`:

```powershell
git submodule update --remote --merge frontend
git add frontend
git commit -m "chore: update frontend submodule to latest nembus-client"
git push
```

---

## Pulling Core API Updates from Nembus

When the Nembus monorepo (`nembus-client` branch) has changes:

```powershell
cd ..\Nembus
git pull origin nembus-client
cd ..\nembus-client
go work sync
```

---

## Branch Policy

| Branch | Purpose |
|--------|---------|
| `main` | Stable desktop POS releases |
| `development` | Active feature development |
| `feature/*` | Individual feature branches |

Feature branches → PR to `development` → merge to `main` on release.

---

## Key Files Explained

| File | Why it's here (not in Nembus) |
|------|-------------------------------|
| `app.go` | Wails lifecycle hooks — not needed in cloud server |
| `internal/db/db_manager.go` | Embedded postgres — desktop-only dependency |
| `internal/sync/` | Clones cloud DB schema to local — desktop-only |
| `internal/printing/` | ESC/POS hardware — no meaning in cloud |
| `migrations/` | Local schema for embedded postgres |
| `proto/sync.proto` | Client-to-cloud sync protocol |
| `frontend/` | Submodule — Angular app compiled and embedded in Wails binary |
