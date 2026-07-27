module github.com/NasTecSol/nembus-server

go 1.25.0

require (
	github.com/NasTecSol/nembus-core v0.0.0
	github.com/gin-gonic/gin v1.11.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.8.0
	github.com/joho/godotenv v1.5.1
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.1
	github.com/swaggo/swag v1.16.6
	golang.org/x/crypto v0.47.0
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.11
)

replace github.com/NasTecSol/nembus-core => ../../packages/core
