module github.com/NasTecSol/nembus-sap-agent

go 1.25.0

require (
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
