# Database Queries

This directory contains the raw SQL query definitions used by [SQLC](https://sqlc.dev/) to generate type-safe Go code for database interactions.

## 📂 Directory Contents

- **`*.sql`**: Each file contains SQL queries for a specific module or entity (e.g., `users.sql`, `products.sql`, `orders.sql`).
- Queries are annotated with SQLC-specific comments to define function names and return types.

## ✍️ Writing Queries

Queries should follow this format:

```sql
-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY name;

-- name: CreateUser :one
INSERT INTO users (name, email) VALUES ($1, $2) RETURNING *;
```

### Supported Annotations:
- `:one`: Returns a single record (and an error if not found).
- `:many`: Returns a slice of records.
- `:exec`: Executes a statement without returning records.
- `:execresult`: Executes a statement and returns the `sql.Result`.

## 🔄 Code Generation

Whenever you add or modify a query in this directory, you must regenerate the Go code:

```bash
# Using Makefile
make sqlc

# Or using sqlc directly
sqlc generate
```

The generated code will be located in `internal/repository/`.

## ⚠️ Important Notes

1. **Schema Sync**: Ensure that your queries match the current database schema defined in the `migrations/` directory.
2. **Type Safety**: SQLC uses the database schema to determine Go types. Be explicit with type casting in SQL if necessary.
3. **No Business Logic**: Keep queries focused on data retrieval and persistence. Complex business rules should reside in the `usecase/` layer.
