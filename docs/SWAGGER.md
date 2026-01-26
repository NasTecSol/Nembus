# Swagger API Documentation

NEMBUS API uses Swagger (OpenAPI) for interactive API documentation.

## Setup

### 1. Install Swagger CLI

```bash
make install-swagger
# or
go install github.com/swaggo/swag/cmd/swag@latest
```

### 2. Generate Swagger Documentation

```bash
make swagger
# or
swag init -g main.go -o docs/swagger
```

This will generate the Swagger documentation files in `docs/swagger/`.

## Accessing Swagger UI

1. Start the server:
   ```bash
   make dev
   # or
   make stg
   ```

2. Open your browser and navigate to:
   ```
   http://localhost:8080/swagger/index.html
   ```

## Updating Documentation

After adding or modifying Swagger annotations in your handlers:

1. Regenerate the documentation:
   ```bash
   make swagger
   ```

2. Restart the server to see the changes

## Swagger Annotations

Swagger annotations are added to handler functions using comments. Example:

```go
// @Summary      Create a new employee
// @Description  Create a new employee with optional login credentials
// @Tags         employees
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        employee      body      object  true  "Employee data"
// @Success      201  {object}  map[string]string
// @Router       /api/employees [post]
func (h *EmployeeHandler) CreateEmployee(c *gin.Context) {
    // ...
}
```

## Testing with Swagger UI

1. Get a dev token (development mode only):
   - Navigate to `/dev/token` endpoint in Swagger UI
   - Copy the token from the response

2. Click "Authorize" button in Swagger UI
3. Enter: `Bearer <your-token>`
4. Now you can test protected endpoints directly from Swagger UI

## Available Endpoints

- **Health Check**: `GET /health`
- **Dev Token** (dev only): `GET /dev/token`
- **Login**: `POST /api/auth/login`
- **Employees**:
  - `GET /api/employees` - List all employees
  - `GET /api/employees/{id}` - Get employee by ID
  - `POST /api/employees` - Create new employee
