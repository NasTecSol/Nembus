package main

import (
	"context"
	"fmt"
	"log"

	"github.com/NasTecSol/nembus-sap-agent/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.NembusDBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	var posCount, posCompletedCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM pos_transactions").Scan(&posCount)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM pos_transactions WHERE status = 'completed'").Scan(&posCompletedCount)

	var invCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM invoices").Scan(&invCount)
	
	var soCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM sales_orders_v2").Scan(&soCount)

	fmt.Printf("Total pos_transactions: %d (Completed: %d)\n", posCount, posCompletedCount)
	fmt.Printf("Total invoices: %d\n", invCount)
	fmt.Printf("Total sales_orders_v2: %d\n", soCount)
}
