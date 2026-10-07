//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	connStr := "postgres://postgres:root1234@localhost:5432/nembus_migration_audit_postfix_20260929?sslmode=disable"
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	tables := []string{"pos_terminals"}
	for _, t := range tables {
		rows, err := conn.Query(ctx, "SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = $1", t)
		if err != nil {
			log.Fatalf("query err: %v", err)
		}
		fmt.Printf("TABLE %s:\n", t)
		for rows.Next() {
			var name, dtype, nullable string
			rows.Scan(&name, &dtype, &nullable)
			fmt.Printf("  %s %s (nullable: %s)\n", name, dtype, nullable)
		}
		rows.Close()
	}
}
