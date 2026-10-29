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

## Biometric verification endpoints

After `make swagger`, the **auth** and **biometrics** tags include:

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/login/biometric` | 1:1 or 1:N fingerprint login (`fmd_base64`, optional `user_login`) |
| POST | `/api/auth/authorize-action/biometric` | Supervisor override via fingerprint + `permission_code` |
| GET | `/api/users/{id}/biometrics` | List enrolled fingers (metadata only) |
| POST | `/api/users/{id}/biometrics/enroll` | Enroll consolidated ISO FMD (`biometrics.enroll` or matching `X-Authorization-Token`) |
| POST | `/api/biometrics/enroll` | Same enroll body without path user id |
| DELETE | `/api/users/{id}/biometrics/{fingerIndex}` | Soft-deactivate a finger template |

Local hardware capture (POS / enrollment station) uses the embedded bridge at `http://127.0.0.1:5050/api/v1` (`/device`, `/capture`, `/enrollment/*`); that bridge is not part of the cloud OpenAPI surface.

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
   - Navigate to `/api/dev/token` endpoint in Swagger UI
   - Copy the token from the response

2. Click "Authorize" button in Swagger UI
3. Enter: `Bearer <your-token>`
4. Now you can test protected endpoints directly from Swagger UI

## Available Endpoints

- **Health Check**: `GET /health`
- **Dev Token** (dev only): `GET /api/dev/token`
- **Login**: `POST /api/auth/login`
- **Employees**:
  - `GET /api/employees` - List all employees
  - `GET /api/employees/{id}` - Get employee by ID
  - `POST /api/employees` - Create new employee
- **Restaurant**:
  - `GET /api/restaurant/tables` - List tables
  - `GET /api/restaurant/orders/{order_id}` - Get order by specific ID
  - `POST /api/restaurant/orders` - Create restaurant order
- **Notification Center & Outbox**:
  - `GET /api/notifications` - Paginated in-app user notifications
  - `GET /api/notifications/unread-count` - Total unread notifications count
  - `POST /api/notifications/{id}/read` - Mark specific notification as read
  - `POST /api/notifications/read-all` - Mark all user notifications as read
  - `GET /api/notifications/channels` - List notification channels
  - `GET /api/notifications/channels/{type}` - Get channel credentials (whatsapp, email)
  - `PUT /api/notifications/channels/{type}` - Upsert channel configuration
  - `GET /api/notifications/templates` - List notification templates
  - `POST /api/notifications/templates` - Upsert notification template
  - `DELETE /api/notifications/templates/{id}` - Delete notification template
  - `GET /api/notifications/rules` - List event routing rules
  - `POST /api/notifications/rules` - Create routing rule
  - `PUT /api/notifications/rules/{id}` - Update routing rule
  - `DELETE /api/notifications/rules/{id}` - Delete routing rule
  - `GET /api/ws/notifications` - Live WebSocket stream for real-time alerts

