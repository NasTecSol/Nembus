//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

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

	var storeID int32
	err = conn.QueryRow(ctx, `
		INSERT INTO stores (organization_id, code, name, store_type)
		VALUES (1, $1, 'Test Store', 'retail') RETURNING id
	`, fmt.Sprintf("ST-%d", time.Now().Unix())).Scan(&storeID)
	if err != nil { log.Fatalf("Failed to insert store: %v", err) }

	var userID int32
	err = conn.QueryRow(ctx, `
		INSERT INTO users (organization_id, username, email, password_hash, must_reset_password, sap_imported)
		VALUES (1, $1, $1, 'hash', false, false) RETURNING id
	`, fmt.Sprintf("user%d", time.Now().Unix())).Scan(&userID)
	if err != nil { log.Fatalf("Failed to insert user: %v", err) }

	var cashierID int32
	err = conn.QueryRow(ctx, `
		INSERT INTO cashiers (user_id, store_id, cashier_code)
		VALUES ($1, $2, $3) RETURNING id
	`, userID, storeID, fmt.Sprintf("CASH-%d", time.Now().Unix())).Scan(&cashierID)
	if err != nil { log.Fatalf("Failed to insert cashier: %v", err) }

	var terminalID int32
	err = conn.QueryRow(ctx, `
		INSERT INTO pos_terminals (store_id, terminal_code)
		VALUES ($1, $2) RETURNING id
	`, storeID, fmt.Sprintf("TERM-%d", time.Now().Unix())).Scan(&terminalID)
	if err != nil { log.Fatalf("Failed to insert terminal: %v", err) }

	var sessionID int32
	err = conn.QueryRow(ctx, `
		INSERT INTO cashier_sessions (cashier_id, pos_terminal_id, session_number, opening_time) 
		VALUES ($1, $2, $3, NOW()) RETURNING id
	`, cashierID, terminalID, fmt.Sprintf("SESS-%d", time.Now().Unix())).Scan(&sessionID)
	if err != nil { log.Fatalf("Failed to insert session: %v", err) }

	var customerID int32
	custCode := fmt.Sprintf("NEM-TEST-%d", time.Now().Unix())
	err = conn.QueryRow(ctx, `
		INSERT INTO customers (organization_id, customer_code, name, email, phone, metadata)
		VALUES (1, $1, 'Test Pos Customer', 'test@example.com', '555-1234', '{}'::jsonb)
		RETURNING id
	`, custCode).Scan(&customerID)
	if err != nil { log.Fatalf("Failed to insert customer: %v", err) }

	var productID int32
	var sku string
	err = conn.QueryRow(ctx, `SELECT id, sku FROM products LIMIT 1`).Scan(&productID, &sku)
	if err != nil { log.Fatalf("Failed to get a product: %v", err) }

	var txID int64
	txNum := fmt.Sprintf("POS-%d", time.Now().Unix())
	err = conn.QueryRow(ctx, `
		INSERT INTO pos_transactions (
			store_id, cashier_id, cashier_session_id, customer_id,
			transaction_number, transaction_date, subtotal, discount_amount,
			tax_amount, total_amount, amount_paid, status, metadata
		) VALUES (
			$1, $2, $3, $4,
			$5, NOW(), 100.0, 0.0, 15.0, 115.0, 115.0, 'completed', '{}'::jsonb
		) RETURNING id
	`, storeID, cashierID, sessionID, customerID, txNum).Scan(&txID)
	if err != nil { log.Fatalf("Failed to insert POS transaction: %v", err) }

	_, err = conn.Exec(ctx, `
		INSERT INTO pos_transaction_lines (
			transaction_id, product_id, quantity, unit_price,
			discount_amount, tax_amount, subtotal, line_total
		) VALUES (
			$1, $2, 1, 100.0, 0.0, 15.0, 100.0, 115.0
		)
	`, txID, productID)
	if err != nil { log.Fatalf("Failed to insert line: %v", err) }

	fmt.Printf("Successfully created POS transaction %s (ID %d) for new customer %d\n", txNum, txID, customerID)
}
