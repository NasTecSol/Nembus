# migrate.ps1
$env:GOOSE_DRIVER="postgres"                  
$env:GOOSE_DBSTRING="postgres://nembus_admin_user:Nembus_Client2023@localhost:5432/masterDB?sslmode=disable"
goose -s -dir .\migrations up
