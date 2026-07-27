# Environment Configuration Guide

NEMBUS supports multiple environments: **development**, **staging**, and **production**.

## Environment Setup

### Development Environment

1. Create `configs/.env.dev` (or copy from `configs/.env.example`):
```env
ENV=development
PORT=8080
MASTER_DB_URL=postgres://postgres:postgres@localhost:5432/nembus_master?sslmode=disable
JWT_SECRET=dev-secret-key-change-in-production-min-32-chars
DEV_USER_ID=00000000-0000-0000-0000-000000000000
DEV_USER_LOGIN=dev_user
LOG_LEVEL=debug
```

2. Run in development mode:
```bash
make dev
```

### Staging Environment

1. Create `configs/.env.stg`:
```env
ENV=staging
PORT=8080
MASTER_DB_URL=postgres://postgres:postgres@localhost:5432/nembus_master_stg?sslmode=disable
JWT_SECRET=staging-secret-key-change-in-production-min-32-chars
LOG_LEVEL=info
```

2. Run in staging mode:
```bash
make stg
```

### Production Environment

For production, use environment variables directly or a secure configuration management system. Do not commit production secrets.

## Makefile Commands

```bash
# Development
make dev              # Run in development mode
make stg              # Run in staging mode
make run              # Run using .env file

# Build
make build            # Build the application

# Swagger
make swagger          # Generate Swagger documentation
make swagger-serve    # Generate and serve Swagger docs
make install-swagger  # Install Swagger CLI tool

# Database
make migrate-master   # Run migrations on master DB
make migrate-tenants  # Run migrations on tenant DBs
make migrate-all      # Run all migrations

# Utilities
make clean            # Clean build artifacts
make test             # Run tests
make sqlc             # Generate SQLC code
make deps             # Install dependencies
```

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `ENV` | Environment name (development/staging/production) | No | development |
| `PORT` | Server port | No | 8080 |
| `MASTER_DB_URL` | Master database connection string | Yes | - |
| `JWT_SECRET` | JWT signing secret (min 32 chars) | Yes | - |
| `DEV_USER_ID` | Dev token user ID | No | 00000000-0000-0000-0000-000000000000 |
| `DEV_USER_LOGIN` | Dev token username | No | dev_user |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | No | info |

## Configuration Loading Order

1. Environment-specific config file (`configs/.env.dev` or `configs/.env.stg`)
2. Root `.env` file (fallback)
3. System environment variables (highest priority)

System environment variables always override file-based configuration.
