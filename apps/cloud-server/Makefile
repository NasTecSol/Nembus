.PHONY: help dev stg build run clean swagger test migrate-master migrate-tenants migrate-all proto install-grpcurl

# Default target
help:
	@echo "NEMBUS - Available Commands:"
	@echo ""
	@echo "  Environment Commands:"
	@echo "    make dev              - Run in development mode"
	@echo "    make stg              - Run in staging mode"
	@echo "    make build            - Build the application"
	@echo "    make run              - Run the application (uses .env)"
	@echo ""
	@echo "  Swagger Commands:"
	@echo "    make swagger          - Generate Swagger documentation"
	@echo "    make swagger-serve    - Generate and view Swagger docs"
	@echo ""
	@echo "  Database Commands:"
	@echo "    make migrate-master   - Run migrations on master database"
	@echo "    make migrate-tenants  - Run migrations on all tenant databases"
	@echo "    make migrate-all      - Run migrations on master and all tenants"
	@echo ""
	@echo "  gRPC Commands:"
	@echo "    make proto            - Regenerate gRPC bindings from proto/backup.proto"
	@echo "    make install-grpcurl  - Install grpcurl for gRPC testing"
	@echo ""
	@echo "  Utility Commands:"
	@echo "    make clean            - Clean build artifacts"
	@echo "    make test            - Run tests"
	@echo "    make sqlc            - Generate SQLC code"

# Development environment
dev:
	@echo "Starting in DEVELOPMENT mode..."
	@ENV=development go run main.go

# Staging environment
stg:
	@echo "Starting in STAGING mode..."
	@ENV=staging go run main.go

# Build the application
build:
	@echo "Building application..."
	@go build -o bin/nembus main.go
	@echo "Build complete: bin/nembus"

# Run the application (uses .env file)
run:
	@echo "Starting application..."
	@go run main.go

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@which swag > /dev/null 2>&1 || (echo "Error: swag command not found. Install it with: make install-swagger" && exit 1)
	@swag init -g main.go -o docs/swagger
	@echo "Swagger docs generated in docs/swagger/"

# Generate and serve Swagger docs
swagger-serve: swagger
	@echo "Swagger documentation available at: http://localhost:8080/swagger/index.html"
	@echo "Starting server..."
	@ENV=development go run main.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf docs/swagger
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Generate SQLC code
sqlc:
	@echo "Generating SQLC code..."
	@sqlc generate
	@echo "SQLC code generated"

# Generate gRPC bindings from proto/backup.proto
proto:
	@echo "Generating gRPC bindings from proto/backup.proto..."
	@which protoc > /dev/null 2>&1 || (echo "Error: protoc not found. Download from https://github.com/protocolbuffers/protobuf/releases" && exit 1)
	@which protoc-gen-go > /dev/null 2>&1 || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@which protoc-gen-go-grpc > /dev/null 2>&1 || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@protoc --go_out=. --go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		proto/backup.proto
	@mv proto/backup.pb.go internal/grpc/backuppb/backup.pb.go 2>/dev/null || true
	@mv proto/backup_grpc.pb.go internal/grpc/backuppb/backup_grpc.pb.go 2>/dev/null || true
	@echo "Done. Files in internal/grpc/backuppb/"

# Install grpcurl for gRPC testing
install-grpcurl:
	@echo "Installing grpcurl..."
	@go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	@echo "grpcurl installed. Make sure $(shell go env GOPATH)/bin is in your PATH"

# Database migration commands
migrate-master:
	@echo "Running migrations on master database..."
	@ENV_FILE=$$([ -f .env.dev ] && echo .env.dev || echo .env); \
	if [ ! -f "$$ENV_FILE" ]; then \
		echo "Error: Neither .env.dev nor .env file found"; \
		exit 1; \
	fi; \
	MASTER_DB_URL=$$(grep -E '^MASTER_DB_URL=' "$$ENV_FILE" | cut -d '=' -f 2- | sed 's/^[[:space:]]*//;s/[[:space:]]*$$//'); \
	if [ -z "$$MASTER_DB_URL" ]; then \
		echo "Error: MASTER_DB_URL is not set in $$ENV_FILE"; \
		exit 1; \
	fi; \
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$$MASTER_DB_URL" go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations up

migrate-tenants:
	@echo "Running migrations on all active tenant databases..."
	@go run cmd/migrate-tenants/main.go

migrate-all: migrate-master migrate-tenants
	@echo "All migrations completed!"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Install Swagger CLI tool
install-swagger:
	@echo "Installing Swagger CLI..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Swagger CLI installed"
	@echo "Note: Make sure $(shell go env GOPATH)/bin is in your PATH"
	@echo "Add to your ~/.bashrc or ~/.zshrc: export PATH=\$$PATH:$(shell go env GOPATH)/bin"