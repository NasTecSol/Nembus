# Use Case Layer (Business Logic)

The Use Case layer represents the core business logic of the NEMBUS application. It implements the "Interactors" from Clean Architecture, orchestrating the flow of data to and from the entities and repositories.

## 🎯 Role & Responsibilities

- **Business Rules**: Implements specific business rules and use cases (e.g., "Assigning a role to a user").
- **Validation**: Performs business-level validation of input data.
- **Orchestration**: Coordinates multiple repository calls to fulfill a single business requirement.
- **Isolation**: Keeps business logic independent of external concerns like HTTP handlers or database implementation.

## 🏗️ Architecture

```
Handler (HTTP) -> Use Case (Business Logic) -> Repository (Data Access)
```

### Dependency Injection
Use Cases do not depend on specific database implementations. Instead, they interact with the data layer via the `repository.Queries` struct.

In this project's multi-tenant architecture, the repository instance is injected into the Use Case per request via the `SetRepository()` method. This ensures that the Use Case always operates on the correct tenant's database.

## 💻 Example Usage

```go
func (uc *UserUseCase) CreateUser(ctx context.Context, ...) *repository.Response {
    // 1. Business Validation
    if email == "" {
        return utils.NewResponse(utils.CodeBadReq, "email required", nil)
    }

    // 2. Data Interaction
    user, err := uc.repo.CreateUser(ctx, ...)
    if err != nil {
        return utils.NewResponse(utils.CodeError, err.Error(), nil)
    }

    // 3. Return Standard Response
    return utils.NewResponse(utils.CodeCreated, "success", user)
}
```

## 📂 Directory Structure

Each file in this directory should correspond to a logical domain:
- `user_usecase.go`: User management, registration, and role assignment.
- `auth_usecase.go`: Authentication and token generation logic.
- `pos_usecase.go`: Point of Sale specific business logic.
- ...and so on.

## ⚠️ Guidelines

1. **No HTTP Leakage**: Use cases should not know about Gin, HTTP status codes (except via the `utils.Response` wrapper), or request/response headers.
2. **Standardized Responses**: Always return a `*repository.Response` (or similar utility) to provide a consistent interface for handlers.
3. **Keep it Pure**: Use cases should be easily testable by providing a mock or a controlled repository instance.
