# proto.ps1 — Regenerate gRPC bindings from proto/backup.proto
# Windows equivalent of: make proto
#
# Usage: .\scripts\proto.ps1

$ErrorActionPreference = "Stop"
$gobin = (go env GOPATH) + "\bin"
$env:PATH = "$gobin;$env:PATH"

Write-Host "🔧 Checking tools..." -ForegroundColor Cyan

# Ensure protoc exists
if (-not (Get-Command protoc -ErrorAction SilentlyContinue)) {
    Write-Error "protoc not found. It should be at $gobin\protoc.exe. Re-run the install step."
    exit 1
}

# Install plugins if missing
if (-not (Get-Command protoc-gen-go -ErrorAction SilentlyContinue)) {
    Write-Host "  Installing protoc-gen-go..." -ForegroundColor Yellow
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
}
if (-not (Get-Command protoc-gen-go-grpc -ErrorAction SilentlyContinue)) {
    Write-Host "  Installing protoc-gen-go-grpc..." -ForegroundColor Yellow
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
}

Write-Host "⚙️  Running protoc on proto/backup.proto..." -ForegroundColor Cyan

protoc `
    --go_out=. `
    --go-grpc_out=. `
    --go_opt=paths=source_relative `
    --go-grpc_opt=paths=source_relative `
    proto/backup.proto

# Move generated files to the correct package location
if (Test-Path "proto\backup.pb.go") {
    Move-Item "proto\backup.pb.go" "internal\grpc\backuppb\backup.pb.go" -Force
}
if (Test-Path "proto\backup_grpc.pb.go") {
    Move-Item "proto\backup_grpc.pb.go" "internal\grpc\backuppb\backup_grpc.pb.go" -Force
}

Write-Host "✅ Done! Generated files in internal/grpc/backuppb/" -ForegroundColor Green
Write-Host "   backup.pb.go      — message types"
Write-Host "   backup_grpc.pb.go — service/client/server stubs"
