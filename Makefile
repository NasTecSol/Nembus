.PHONY: sync-core generate-core generate-client build-server build-client dev-server dev-client tidy swagger db-diff db-migrate db-status db-hash db-lint sqlc verify test

# Database / Atlas Commands
db-diff:
	@echo "==> Generating migration diff from db/schema/..."
	@atlas migrate diff $(name) --env local

db-migrate:
	@echo "==> Running Atlas migrations on master and tenant databases..."
	@cd apps/cloud-server && go run cmd/migrate-tenants/main.go

db-status:
	@echo "==> Checking Atlas migration status..."
	@atlas migrate status --env local

db-hash:
	@echo "==> Re-calculating Atlas migration directory hash..."
	@atlas migrate hash --dir "file://packages/core/db/migrations"

db-lint:
	@echo "==> Linting Atlas migrations against dev database..."
	@atlas migrate lint --env local --latest 1

# SQLC Code Generation
sqlc: generate-core

generate-core:
	@echo "==> Running sqlc generate in packages/core..."
	@cd packages/core && sqlc generate

sync-core:
	@echo "==> Syncing core schema & queries into pos-client..."
	@cd apps/pos-client && powershell -ExecutionPolicy Bypass -File scripts/sync_core.ps1

# Build & Run Commands
build-server:
	@echo "==> Building cloud server..."
	@cd apps/cloud-server && go build -o bin/nembus-server main.go

build-client:
	@echo "==> Building POS client (Wails)..."
	@cd apps/pos-client && wails build

dev-server:
	@cd apps/cloud-server && go run main.go dev

dev-client:
	@cd apps/pos-client && wails dev

sync-notebook:
	@echo "==> Bundling NotebookLM context sources..."
	@powershell -ExecutionPolicy Bypass -File scripts/build_notebook_sources.ps1

tidy:
	@go work sync
	@cd packages/core && go mod tidy
	@cd apps/cloud-server && go mod tidy
	@cd apps/pos-client && go mod tidy

swagger:
	@echo "==> Generating Swagger documentation..."
	@cd apps/cloud-server && swag init -g main.go --dir ./,../../packages/core/handler -o docs/swagger -p camelcase --parseInternal --parseDependency

test:
	@echo "==> Running tests across packages..."
	@cd packages/core && go test ./...
	@cd apps/cloud-server && go test ./...

db-validate:
	@echo "==> Validating Atlas migration integrity..."
	@atlas migrate validate --dir "file://packages/core/db/migrations"

verify: db-validate sqlc test
	@echo "==> All schema, migration, sqlc, and build verifications passed!"
