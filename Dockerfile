FROM golang:1.25-alpine AS build

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o server .

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata wget postgresql-client

COPY --from=build /app/server /app/server

RUN mkdir -p /app/images

EXPOSE 3000
EXPOSE 50051

CMD ["/app/server"]
