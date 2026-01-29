# NEMBUS Backend  
**Nasar Entity-driven Modular Business Unified System**

A scalable, clean-architecture backend built with **Go**, **PostgreSQL**, **SQLC**, and **Gin**, designed for enterprise-grade business systems such as ERP, HR, POS, and IAM platforms.

---

## 🚀 Tech Stack
| Layer | Technology |
|-----|-----------|
| Language | Go (Golang) |
| Web Framework | Gin |
| Database | PostgreSQL |
| Query Layer | SQLC |
| Architecture | Clean Architecture / Repository Pattern |
| Config | Environment Variables |
| API Style | REST (JSON) |
| Auth (optional) | JWT / Middleware based |
| Migrations | golang-migrate / goose (optional) |

---

## 🧠 Architectural Philosophy

This project follows **Clean Architecture principles** inspired by:
- Robert C. Martin (Uncle Bob)
- Domain-Driven Design (DDD)
- Enterprise Application Architecture patterns

### Key Goals:
- **Separation of concerns**
- **Database-agnostic business logic**
- **Testability**
- **Scalability**
- **Long-term maintainability**

---

## 📂 Project Structure

```

.
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
│
├── internal/
│   ├── config/              # Environment & config loading
│   │   └── config.go
│   │
│   ├── database/
│   │   ├── migrations/      # SQL migrations
│   │   └── postgres.go
│   │
│   ├── domain/              # Core business entities
│   │   └── user.go
│   │
│   ├── repository/          # Data access layer (SQLC wrappers)
│   │   └── user_repository.go
│   │
│   ├── service/             # Business logic layer
│   │   └── user_service.go
│   │
│   ├── handler/             # HTTP handlers (Gin)
│   │   └── user_handler.go
│   │
│   ├── middleware/          # Auth, logging, recovery
│   │   └── auth.go
│   │
│   └── router/              # Route definitions
│       └── router.go
│
├── sql/
│   ├── queries/             # SQLC queries
│   │   └── user.sql
│   └── schema/              # SQL schema
│       └── user.sql
│
├── sqlc.yaml                # SQLC configuration
├── go.mod
├── go.sum
├── .env.example
└── README.md

```

---

## 🔄 Request Flow (High Level)

```

HTTP Request
↓
Gin Router
↓
Middleware (Auth / Logging)
↓
Handler (HTTP layer)
↓
Service (Business logic)
↓
Repository (SQLC)
↓
PostgreSQL

````

---

## 🗄️ Database & SQLC

### Why SQLC?
- Compile-time safety for SQL
- No ORM magic
- Full control over queries
- Excellent performance

### Example SQLC Query

```sql
-- name: GetUserByID :one
SELECT id, email, name
FROM users
WHERE id = $1;
````

SQLC generates **type-safe Go code** automatically.

---

## ⚙️ Environment Configuration

NEMBUS supports multiple environments (development, staging, production) with environment-specific configuration files.

### Setup

1. **Development**: Create `configs/.env.dev`
2. **Staging**: Create `configs/.env.stg`
3. **Production**: Use system environment variables or secure config management

### Configuration Files

See [ENVIRONMENTS.md](ENVIRONMENTS.md) for detailed environment configuration guide.

### Quick Example

```env
ENV=development
PORT=8080
MASTER_DB_URL=postgres://user:pass@localhost:5432/dbname?sslmode=disable
JWT_SECRET=your-secret-key-minimum-32-characters-long
```

### Makefile Commands

```bash
make dev              # Run in development mode
make stg              # Run in staging mode
make build            # Build application
make swagger          # Generate API documentation
make migrate-all      # Run all migrations
```

See `make help` for all available commands.

---

## ▶️ Running the Application

### Quick Start

```bash
# Install dependencies
make deps

# Run in development mode
make dev

# Or run in staging mode
make stg
```

### Detailed Setup

1. **Install Dependencies**
   ```bash
   make deps
   # or
   go mod tidy
   ```

2. **Setup Environment Configuration**
   - Copy `configs/.env.example` to `configs/.env.dev` for development
   - Update database connection strings and secrets

3. **Run Database Migrations**
   ```bash
   make migrate-master    # Master database
   make migrate-tenants   # Tenant databases
   # or
   make migrate-all       # All databases
   ```

4. **Generate SQLC Code** (if needed)
   ```bash
   make sqlc
   # or
   sqlc generate
   ```

5. **Start the Server**
   ```bash
   make dev    # Development mode
   make stg    # Staging mode
   make run    # Using .env file
   ```

Server will start at: `http://localhost:8080`

### API Documentation (Swagger)

1. **Generate Swagger Documentation**
   ```bash
   make swagger
   # or
   make install-swagger  # Install Swagger CLI first
   swag init -g main.go -o docs/swagger
   ```

2. **Access Swagger UI**
   - Start the server: `make dev`
   - Open browser: `http://localhost:8080/swagger/index.html`

See [docs/SWAGGER.md](docs/SWAGGER.md) for more details.

---

## 🔐 Middleware

Supported middleware pattern:

* JWT authentication
* Request logging
* Panic recovery
* Role-based access control (RBAC)

Example:

```go
router.Use(middleware.JWTAuth())
```

---

## 🧪 Testing Strategy

* **Unit tests** for services
* **Repository tests** with test database
* **Handler tests** using `httptest`
* SQLC enables mocking DB logic cleanly

---

## 📦 Use Cases

This backend architecture is suitable for:

* ERP Systems
* HR Management Systems
* POS Systems
* Inventory & Procurement
* IAM / Access Control Systems
* SaaS Multi-Tenant Platforms

---

## 🛣️ Roadmap

* [ ] Multi-tenancy support
* [ ] Role & permission engine
* [ ] Event-driven modules
* [ ] Audit logs
* [ ] API versioning
* [ ] GraphQL gateway (optional)

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit with clear messages
4. Submit a Pull Request

---

## 📜 License

MIT License © Nasar Tech

```

---

If you want, next we can:
- Align this README **exactly** with `go-clean-template`
- Add **API versioning conventions**
- Add **RBAC + policy engine section**
- Design a **mono-repo vs multi-repo strategy**
- Create **Makefile + Docker setup**

Just say the word 🚀
```
