# migrate.ps1
$env:GOOSE_DRIVER="postgres"                  
$env:GOOSE_DBSTRING="postgres://root:nastecsol@localhost:5432/masterDB?sslmode=disable"
goose -s -dir .\migrations up
