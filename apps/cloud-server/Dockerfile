FROM golang:1.25-alpine AS build

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate Swagger docs before building.
# main.go imports NEMBUS/docs/swagger, so this package must exist.
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN $(go env GOPATH)/bin/swag init -g main.go -d ./,../../packages/core/handler -o docs/swagger --parseDependency --parseInternal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o server .

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata wget postgresql-client

COPY --from=build /app/server /app/server

RUN mkdir -p /app/images

EXPOSE 3000
EXPOSE 50051

CMD ["/app/server"]
