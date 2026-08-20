module github.com/NasTecSol/nembus-sap-agent

go 1.25.0

require (
<<<<<<< Updated upstream
	github.com/NasTecSol/nembus-core v0.0.0
	github.com/NasTecSol/nembus-sap v0.0.0
	github.com/gin-gonic/gin v1.11.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/kardianos/service v1.2.2
	github.com/microsoft/go-mssqldb v1.8.0
	modernc.org/sqlite v1.35.0
)

replace github.com/NasTecSol/nembus-sap => ../../packages/sap
replace github.com/NasTecSol/nembus-core => ../../packages/core
replace github.com/NasTecSol/nembus-client => ../pos-client
=======
	github.com/jackc/pgx/v5 v5.8.0
	github.com/joho/godotenv v1.5.1
	github.com/microsoft/go-mssqldb v1.8.0
)

require (
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/text v0.29.0 // indirect
)

replace github.com/NasTecSol/nembus-core => ../../packages/core
>>>>>>> Stashed changes
