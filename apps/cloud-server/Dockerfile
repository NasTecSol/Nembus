FROM golang:1.25-bookworm AS build

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    git ca-certificates tzdata build-essential \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate Swagger docs before building.
# main.go imports NEMBUS/docs/swagger, so this package must exist.
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN $(go env GOPATH)/bin/swag init -g main.go -d ./,../../packages/core/handler -o docs/swagger --parseDependency --parseInternal

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o server .

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata wget postgresql-client \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /var/cache/nembus/Nembus/biometrics/v3.4.0/linux_amd64 && \
    chmod -R 777 /var/cache/nembus && \
    echo "/var/cache/nembus/Nembus/biometrics/v3.4.0/linux_amd64" > /etc/ld.so.conf.d/nembus_biometrics.conf && \
    ldconfig
ENV XDG_CACHE_HOME=/var/cache/nembus
ENV LD_LIBRARY_PATH=/var/cache/nembus/Nembus/biometrics/v3.4.0/linux_amd64:${LD_LIBRARY_PATH}

COPY --from=build /app/server /app/server

RUN mkdir -p /app/images

EXPOSE 3000
EXPOSE 50051

CMD ["/app/server"]
