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

# Use the official PostgreSQL repository so pg_dump supports newer tenant databases.
ARG PG_CLIENT_MAJOR=18
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata wget \
    && mkdir -p /usr/share/postgresql-common/pgdg \
    && wget -qO /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc \
        https://www.postgresql.org/media/keys/ACCC4CF8.asc \
    && printf '%s\n' 'deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] https://apt.postgresql.org/pub/repos/apt bookworm-pgdg main' \
        > /etc/apt/sources.list.d/pgdg.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends postgresql-client-${PG_CLIENT_MAJOR} \
    && /usr/lib/postgresql/${PG_CLIENT_MAJOR}/bin/pg_dump --version \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /var/cache/nembus/Nembus/biometrics/v3.4.0/linux_amd64 && \
    # Root-owned and not group/world writable: these libraries are dlopen()ed and registered
    # with ldconfig, so only the server's user (root) may modify them.
    chown -R root:root /var/cache/nembus && \
    chmod -R u=rwX,go=rX /var/cache/nembus && \
    echo "/var/cache/nembus/Nembus/biometrics/v3.4.0/linux_amd64" > /etc/ld.so.conf.d/nembus_biometrics.conf && \
    ldconfig
ENV XDG_CACHE_HOME=/var/cache/nembus
ENV LD_LIBRARY_PATH=/var/cache/nembus/Nembus/biometrics/v3.4.0/linux_amd64:${LD_LIBRARY_PATH}

COPY --from=build /app/server /app/server

RUN mkdir -p /app/images

EXPOSE 3000
EXPOSE 50051

CMD ["/app/server"]
