# NEMBUS Backend  
**Nasar Entity-driven Modular Business Unified System**

A scalable, clean-architecture backend built with **Go**, **PostgreSQL**, **SQLC**, and **Gin**, designed for enterprise-grade business systems such as ERP, Retail, and **Restaurant/Cafe** platforms.

---

## 🚀 Tech Stack
| Layer | Technology |
|-----|-----------|
| Language | Go (Golang) |
| Web Framework | Gin |
| Database | PostgreSQL |
| Query Layer | SQLC |
| Architecture | Clean Architecture (Handler -> Usecase -> Repository) |
| Config | Environment Variables |
| API Style | REST (JSON) |
| Auth | JWT / Middleware based |
| Migrations | Goose (Multi-tenant support) |

---

## 🧠 Architectural Philosophy

This project follows **Clean Architecture principles** with a focus on:
- **Separation of concerns**: Each layer has a specific responsibility.
- **Dependency Inversion**: High-level modules do not depend on low-level modules.
- **Testability**: Logic is decoupled from infrastructure.
- **Multi-tenancy**: First-class support for isolated tenant databases.

---

## 📂 Project Structure

```
.
├── cmd/                     # CLI tools (e.g., tenant migration tools)
├── docs/                    # Documentation and Swagger definitions
├── internal/                # Private application code
│   ├── config/              # Configuration loading logic
│   ├── handler/             # HTTP handlers (Gin) - Entry point for requests
│   ├── middleware/          # HTTP middlewares (Auth, Tenant-detection, etc.)
│   ├── repository/          # Data access layer (SQLC generated code & wrappers)
│   ├── routing/             # API route definitions
│   └── usecase/             # Business logic layer (Core domain logic)
├── migrations/              # Database migration files (Goose format)
├── queries/                 # SQLC query definitions (.sql files)
├── scripts/                 # Utility scripts for development/deployment
│   ├── init-Data-Dump.sql            # Core retail/wholesale master data
│   └── init-Restaurant-Data-Dump.sql # Restaurant-specific master data (menus, ingredients, tables)
├── utils/                   # Shared utility functions
├── .env.dev                 # Local development environment variables
├── ENVIRONMENTS.md          # Guide for environment setup
├── MIGRATIONS.md            # Guide for database migrations
├── Makefile                 # Automation commands (build, run, migrate)
├── main.go                  # Application entry point
└── sqlc.yaml                # SQLC configuration
```

---

## 🔄 Request Flow

```
HTTP Request
↓
Gin Router (internal/routing)
↓
Middleware (Auth / Tenant Selection)
↓
Handler (internal/handler)
↓
UseCase (internal/usecase)
↓
Repository (internal/repository - SQLC)
↓
PostgreSQL (Tenant-specific DB)
```

---

## 🗄️ Database & SQLC

### Why SQLC?
- Type-safe Go code from raw SQL.
- No heavy ORM overhead.
- Compile-time SQL validation.

### Workflow
1. Define schema in migrations.
2. Write SQL queries in `queries/`.
3. Run `make sqlc` to generate code in `internal/repository/`.

---

## ⚙️ Environment Configuration

NEMBUS uses environment variables for configuration.

### Setup
1. Copy `.env.dev` if you need a template or use it directly for local development.
2. Update `MASTER_DB_URL` and `JWT_SECRET`.

See [ENVIRONMENTS.md](ENVIRONMENTS.md) for a detailed guide.

### Makefile Commands
```bash
make dev              # Run in development mode
make build            # Build application
make swagger          # Generate API documentation
make migrate-all      # Run all migrations (Master + Tenants)
make sqlc             # Generate SQLC code
```

---

## ▶️ Running the Application

### Quick Start
```bash
# Install dependencies
go mod tidy

# Run migrations
make migrate-all

# Start server
make dev
```

Server starts at: `http://localhost:8080`

### API Documentation (Swagger)
Open: `http://localhost:8080/swagger/index.html` (after running `make dev`)

---

## 🧪 Testing
Run tests using:
```bash
make test
```

---

## 📜 License
MIT License © Nasar Tech
