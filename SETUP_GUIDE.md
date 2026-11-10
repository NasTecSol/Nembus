# Developer Setup Guide

This guide details how to set up your environment, build the Cloud Server, run the Cloud Server locally without containers, and build the Wails v2 Windows Desktop application within the Nembus Monorepo.

---

## 1. Prerequisites

Ensure you have the following installed on your machine:

- **Go**: Version `1.25.0` or higher ([golang.org](https://go.dev/dl/))
- **PostgreSQL**: Local PostgreSQL server (v14+) installed and running on `localhost:5432` ([postgresql.org](https://www.postgresql.org/download/windows/))
- **Node.js**: Version `18.x` or higher + `npm` ([nodejs.org](https://nodejs.org/))
- **Wails CLI v2**: Install globally via Go:
  ```powershell
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
- **C++ Build Tools (Windows)**: Required by Wails & CGO for Windows desktop packaging.
  - Install **Visual Studio Build Tools** with the **"Desktop development with C++"** workload selected.
- **SQLC** (Optional for schema generation):
  ```powershell
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
  ```

---

## 2. Running Nembus Cloud Server Locally (Without Container)

To run the Cloud Server directly on your local machine using your local PostgreSQL server:

### Step 2.1: Create Local Master Database
Open PostgreSQL CLI (`psql`) or PGAdmin and create the master database:
```sql
CREATE DATABASE "masterDB";
```

### Step 2.2: Apply Base Schema Migration
Run the canonical base schema migration onto your local `masterDB`:
```powershell
psql -U postgres -d masterDB -f packages/core/migrations/000001_base_schema.sql
```

### Step 2.3: Configure `.env.dev` in `apps/cloud-server`
Open `apps/cloud-server/.env.dev` and ensure `MASTER_DB_URL` points to `localhost`:
```env
ENV=development
PORT=8080
GRPC_PORT=50051
MASTER_DB_URL=postgres://postgres:your_pg_password@localhost:5432/masterDB?sslmode=disable
JWT_SECRET=nastecsol
PG_DUMP_PATH=C:\Program Files\PostgreSQL\18\bin\pg_dump.exe
```
*(Replace `postgres:your_pg_password` with your local PostgreSQL user and password).*

### Step 2.4: Start Local Cloud Server
From the monorepo root:
```powershell
make dev-server
```

Or directly inside `apps/cloud-server`:
```powershell
cd apps/cloud-server
go run main.go dev
```

### Step 2.5: Verify Local Server
- **Health Check**: `http://localhost:8080/health`
- **Swagger Documentation**: `http://localhost:8080/swagger/index.html`
- **Generate Dev Token**: `http://localhost:8080/dev/token`

---

## 3. Cloning & Git Submodule Initialization

The POS client frontend (`apps/pos-client/frontend`) is tracked as a Git submodule (`NasTecSol/NPOS-Bofc`).

When cloning or pulling the repository:

```powershell
# If cloning fresh
git clone --recursive https://github.com/NasTecSol/nembus-monorepo.git

# If already cloned, initialize submodules:
git submodule update --init --recursive
```

---

## 4. Building the Wails Windows Desktop App (`apps/pos-client`)

The POS client desktop app uses **Wails v2** to package the Go backend and Angular frontend into a single native Windows `.exe` installer/executable.

### Step 4.1: Verify Wails Doctor
Run `wails doctor` to ensure all system dependencies are detected:
```powershell
wails doctor
```
Ensure `C++ Compiler` and `npm` are marked with green checkmarks.

### Step 4.2: Frontend Dependencies Setup
Ensure npm dependencies are installed inside the frontend submodule:
```powershell
cd apps/pos-client/frontend
npm install
```

### Step 4.3: Live Development Mode (`wails dev`)
Run live dev mode with hot reloading:
```powershell
cd apps/pos-client
wails dev
```
*In dev mode, Wails compiles the Go backend, starts the embedded PostgreSQL database manager, builds the Angular frontend, and opens the native application window.*

### Step 4.4: Windows Production Executable Build (`wails build`)
To compile a standalone Windows executable (`nembus-client.exe`):

```powershell
cd apps/pos-client
wails build
```

The compiled binary will be placed at:
`apps/pos-client/build/bin/nembus-client.exe`

#### Additional Wails Build Flags:
- **Clean build**: `wails build -clean`
- **Compressed executable (UPX)**: `wails build -upx`
- **Debug build with devtools enabled**: `wails build -debug`

---

## 5. Troubleshooting Windows Wails Builds

1. **`gcc` / C++ Compiler Errors**:
   - Ensure Visual Studio Build Tools with C++ workload is installed.
   - Alternatively, install `MinGW-w64` via `choco install mingw` and ensure `gcc` is in your system `PATH`.
2. **WebView2 Runtime Missing**:
   - Windows 10/11 usually has Microsoft Edge WebView2 pre-installed. If missing, download the evergreen bootstrapper from Microsoft.
3. **Frontend Build Failures**:
   - Clear Angular build cache: `cd apps/pos-client/frontend && rm -rf .angular dist`.
   - Run `npm run build` manually inside `frontend/` to check for TypeScript errors.
