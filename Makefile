.PHONY: sync-core generate-core generate-client build-server build-client dev-server dev-client tidy

sync-core:
	@echo "==> Syncing core schema & queries into pos-client..."
	@cd apps/pos-client && powershell -ExecutionPolicy Bypass -File scripts/sync_core.ps1

generate-core:
	@echo "==> Running sqlc generate in packages/core..."
	@cd packages/core && sqlc generate

generate-client:
	@echo "==> Running sqlc generate in apps/pos-client..."
	@cd apps/pos-client && sqlc generate

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


