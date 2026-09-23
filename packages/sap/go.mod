module github.com/NasTecSol/nembus-sap

go 1.25.0

require (
	github.com/NasTecSol/nembus-core v0.0.0
	github.com/google/uuid v1.6.0
)

replace github.com/NasTecSol/nembus-core => ../core
replace github.com/NasTecSol/nembus-client => ../../apps/pos-client
