package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func main() {
	url := "postgres://postgres:root1234@localhost:5432/nembus_migration_audit_postfix_20260929?sslmode=disable"
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Printf("Connect error: %v\n", err)
		return
	}
	defer conn.Close(context.Background())

	tx, err := conn.Begin(context.Background())
	if err != nil {
		fmt.Printf("Begin error: %v\n", err)
		return
	}
	var id int
	err = tx.QueryRow(context.Background(), "SELECT id FROM products WHERE sku = $1", "bad\x00sku").Scan(&id)
	fmt.Printf("Scan error: %v\n", err)

	err = tx.Commit(context.Background())
	fmt.Printf("Commit error: %v\n", err)
}
