# migrate.ps1 - Atlas migration runner for Cloud Server
$ErrorActionPreference = "Stop"

# Auto-detect or run through Go tenant migration orchestrator
go run cmd/migrate-tenants/main.go
