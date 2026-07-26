# Database Migrations Guide

This project uses [goose](https://github.com/pressly/goose) for database migrations in a multi-tenant architecture.

## Prerequisites

1. Install goose:
   ```bash
   go install github.com/pressly/goose/v3/cmd/goose@latest
   ```

2. Set environment variable:
   ```bash
   export MASTER_DB_URL="postgresql://user:password@localhost:5432/master_db?sslmode=disable"
   ```
   
   Or create a `.env` file:
   ```
   MASTER_DB_URL=postgresql://user:password@localhost:5432/master_db?sslmode=disable
   ```

## Makefile Commands

### Migrate Master Database
```bash
make migrate-master
```
Runs migrations on the master database (where the `tenants` table is stored).

### Migrate All Tenant Databases
```bash
make migrate-tenants
```
Runs migrations on all active tenant databases. The script:
1. Connects to the master database
2. Retrieves all active tenants
3. Runs migrations on each tenant's database

### Migrate Everything
```bash
make migrate-all
```
Runs migrations on both master and all tenant databases.

### Rollback Commands

```bash
make migrate-down-master    # Rollback master database
make migrate-down-tenants    # Rollback all tenant databases
make migrate-down-all        # Rollback master and all tenants
```

## Manual Usage

### Using the Go Script Directly

```bash
# Migrate all tenants (up)
go run cmd/migrate-tenants/main.go

# Rollback all tenants (down)
go run cmd/migrate-tenants/main.go -down

# Custom migrations directory
go run cmd/migrate-tenants/main.go -dir ./custom-migrations
```

### Using Goose Directly

```bash
# Master database
GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgresql://..." goose -dir ./migrations up

# Specific tenant database
GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgresql://..." goose -dir ./migrations up
```

## Migration Files

Migrations are stored in the `./migrations` directory. Each migration file should follow the naming convention:
- `XXXXX_description.sql` (e.g., `00001_initialise_db.sql`)

Migration files use goose directives:
- `-- +goose Up` - Migration to apply
- `-- +goose Down` - Migration to rollback

## Architecture

```
Master Database (tenants table)
    ├── Tenant 1 Database (acme-corp)
    ├── Tenant 2 Database (techstart)
    └── Tenant 3 Database (startupxyz)
```

Each tenant has its own database with the same schema. Migrations are applied to:
1. Master database (contains tenant metadata)
2. Each tenant database (contains tenant-specific data)

## Troubleshooting

### "MASTER_DB_URL is not set"
Set the environment variable or create a `.env` file with `MASTER_DB_URL`.

### "goose: command not found"
Install goose: `go install github.com/pressly/goose/v3/cmd/goose@latest`

### Migration fails on a specific tenant
The script will continue with other tenants and report failures at the end. Check the tenant's database connection string in the master database.

### Tenant database doesn't exist
Create the tenant database before running migrations, or ensure the connection string in the `tenants` table is correct.

## Example Workflow

1. Create a new migration file:
   ```bash
   touch migrations/00002_add_new_table.sql
   ```

2. Add migration SQL:
   ```sql
   -- +goose Up
   CREATE TABLE new_table (...);
   
   -- +goose Down
   DROP TABLE IF EXISTS new_table;
   ```

3. Run migrations:
   ```bash
   make migrate-all
   ```

4. Verify:
   - Check master database
   - Check each tenant database
   - Verify tables exist in all databases
