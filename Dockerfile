FROM golang:1.25-alpine AS build

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

# Copy workspace configuration and module dependencies
COPY go.work go.work.sum* ./
COPY packages/core/go.mod packages/core/go.sum* ./packages/core/
COPY apps/cloud-server/go.mod apps/cloud-server/go.sum* ./apps/cloud-server/
COPY apps/pos-client/go.mod apps/pos-client/go.sum* ./apps/pos-client/

# Download dependencies
RUN cd packages/core && go mod download || true
RUN cd apps/cloud-server && go mod download || true

# Copy full source tree
COPY . .

# Generate Swagger docs inside apps/cloud-server
WORKDIR /app/apps/cloud-server
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN $(go env GOPATH)/bin/swag init -g main.go -o docs/swagger --parseDependency --parseInternal

# Build static binary for Linux x86_64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o server main.go

# Production runtime container
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata wget postgresql-client

COPY --from=build /app/apps/cloud-server/server /app/server

RUN mkdir -p /app/images

EXPOSE 3000
EXPOSE 50051

CMD ["/app/server"]
