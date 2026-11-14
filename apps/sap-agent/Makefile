.PHONY: help build build-cli build-all run dev clean test

VERSION ?= $(shell git tag --sort=-v:refname 2>/dev/null | grep "^sap-v" | head -n 1 | sed 's/^sap-//' || echo "v1.0.0")
ifeq ($(strip $(VERSION)),)
	VERSION := v1.0.0
endif

CLOUD_URL ?= https://nembus.nashrms.com
APP_UPDATE_KEY ?= nembus-pos-update-client-key-default
LDFLAGS := -s -w -X main.Version=$(VERSION) -X main.AppUpdateKey=$(APP_UPDATE_KEY) -X main.CloudURL=$(CLOUD_URL)

help:
	@echo "Nembus SAP Agent - Build Targets:"
	@echo "  make build       - Build the primary SAP Agent service (bin/sap-agent.exe)"
	@echo "  make build-cli   - Build the CLI runner (bin/sap-agent-cli.exe)"
	@echo "  make build-all   - Build both agent service and CLI runner"
	@echo "  make run         - Run the primary SAP Agent service"
	@echo "  make dev         - Run the CLI daemon in development mode"
	@echo "  make test        - Run tests in sap-agent"
	@echo "  make clean       - Remove built binaries from bin/"

build:
	@echo "==> Building SAP Agent service ($(VERSION))..."
	@go build -ldflags "$(LDFLAGS)" -o bin/sap-agent.exe cmd/agent/main.go
	@echo "Built: bin/sap-agent.exe"

build-cli:
	@echo "==> Building SAP CLI runner ($(VERSION))..."
	@go build -ldflags "-s -w -X main.version=$(VERSION)" -o bin/sap-agent-cli.exe cmd/main.go
	@echo "Built: bin/sap-agent-cli.exe"

build-all: build build-cli

run:
	@go run -ldflags "$(LDFLAGS)" cmd/agent/main.go

dev:
	@go run cmd/main.go daemon

test:
	@go test ./...

clean:
	@rm -rf bin/sap-agent.exe bin/sap-agent-cli.exe

