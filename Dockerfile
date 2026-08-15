FROM golang:1.25-alpine AS build

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

# Copy workspace configuration and module dependencies.
COPY go.work go.work.sum* ./
COPY packages/core/go.mod packages/core/go.sum* ./packages/core/
COPY apps/cloud-server/go.mod apps/cloud-server/go.sum* ./apps/cloud-server/
COPY apps/pos-client/go.mod apps/pos-client/go.sum* ./apps/pos-client/

# Download dependencies.
RUN cd packages/core && go mod download || true
RUN cd apps/cloud-server && go mod download || true

# Copy the full monorepo.
COPY . .

WORKDIR /app/apps/cloud-server

# Generate Swagger documentation.
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN $(go env GOPATH)/bin/swag init \
    -g main.go \
    -d ./,/app/packages/core/handler \
    -o docs/swagger \
    --parseDependency \
    --parseInternal

# Build the API server.
RUN mkdir -p /out && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /out/server main.go

# Build the tenant migration runner.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /out/migrate-tenants ./cmd/migrate-tenants

# Download Atlas CLI binary.
RUN wget -qO /out/atlas https://release.ariga.io/atlas/atlas-linux-amd64-latest && \
    chmod +x /out/atlas

# Production runtime container.
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget \
    postgresql-client

# Application and migration executables.
COPY --from=build /out/server /app/server
COPY --from=build /out/migrate-tenants /app/migrate-tenants
COPY --from=build /out/atlas /usr/local/bin/atlas

# Atlas configuration and migrations.
COPY --from=build /app/packages/core/db/migrations /app/migrations
COPY --from=build /app/atlas.hcl /app/atlas.hcl

RUN mkdir -p /app/images

EXPOSE 3000
EXPOSE 50051

CMD ["/app/server"]
