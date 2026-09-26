package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://root:nastecsol@localhost:5432/stg?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer pool.Close()

	log.Println("================================================================")
	log.Println(" Nembus SAP Cash Invoices -> POS Transactions Projection Tool")
	log.Println("================================================================")

	// 1. Ensure Index on session_number
	_, err = pool.Exec(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS idx_cashier_sessions_session_number ON cashier_sessions(session_number);`)
	if err != nil {
		log.Fatalf("Failed to create index on cashier_sessions: %v", err)
	}

	// 2. Resolve default Store, Cashier, Terminal, and Customer
	var storeID, orgID int
	err = pool.QueryRow(ctx, `SELECT id, organization_id FROM stores WHERE code = '01' LIMIT 1`).Scan(&storeID, &orgID)
	if err != nil || storeID == 0 {
		err = pool.QueryRow(ctx, `SELECT id, organization_id FROM stores ORDER BY id ASC LIMIT 1`).Scan(&storeID, &orgID)
		if err != nil {
			log.Fatalf("No store found: %v", err)
		}
	}

	var terminalID int
	err = pool.QueryRow(ctx, `
		INSERT INTO pos_terminals (store_id, terminal_code, terminal_name, is_active)
		VALUES ($1, 'TERM-01', 'Main POS Counter 01', true)
		ON CONFLICT (store_id, terminal_code) DO UPDATE SET is_active = true
		RETURNING id;
	`, storeID).Scan(&terminalID)
	if err != nil {
		log.Fatalf("Failed to resolve pos_terminal: %v", err)
	}

	var userID int
	_ = pool.QueryRow(ctx, `SELECT id FROM users WHERE username IN ('manager', 'admin-1') ORDER BY id ASC LIMIT 1`).Scan(&userID)
	if userID == 0 {
		_ = pool.QueryRow(ctx, `SELECT id FROM users ORDER BY id ASC LIMIT 1`).Scan(&userID)
	}

	var cashierID int
	err = pool.QueryRow(ctx, `
		INSERT INTO cashiers (user_id, store_id, cashier_code, drawer_limit, discount_limit, is_active)
		VALUES ($1, $2, 'CASHIER-01', 50000, 10, true)
		ON CONFLICT (store_id, cashier_code) DO UPDATE SET is_active = true
		RETURNING id;
	`, userID, storeID).Scan(&cashierID)
	if err != nil {
		log.Fatalf("Failed to resolve cashier: %v", err)
	}

	var customerID int
	err = pool.QueryRow(ctx, `
		INSERT INTO customers (organization_id, customer_code, name, is_active)
		VALUES ($1, 'C00006', 'عميل نقدي', true)
		ON CONFLICT (organization_id, customer_code) DO UPDATE SET is_active = true
		RETURNING id;
	`, orgID).Scan(&customerID)
	if err != nil {
		log.Fatalf("Failed to resolve customer: %v", err)
	}

	var defaultProductID int
	_ = pool.QueryRow(ctx, `SELECT id FROM products ORDER BY id ASC LIMIT 1`).Scan(&defaultProductID)

	log.Printf("[CONFIG] Target Store: %d | Cashier: %d | Terminal: %d | Customer: %d | Fallback Product: %d\n",
		storeID, cashierID, terminalID, customerID, defaultProductID)

	// 3. Populate Closed Daily Cashier Sessions (Z-Reports)
	log.Println("[STEP 1/4] Creating Closed Daily Cashier Sessions (Z-Reports)...")
	sessQuery := `
	INSERT INTO cashier_sessions (
		cashier_id, pos_terminal_id, session_number,
		opening_time, closing_time,
		opening_balance, closing_balance, expected_balance, variance,
		status
	)
	SELECT 
		$1, $2,
		'SES-' || TO_CHAR(inv_date, 'YYYYMMDD'),
		inv_date + interval '8 hours',
		inv_date + interval '23 hours',
		500.00,
		500.00 + daily_total,
		500.00 + daily_total,
		0.00,
		'closed'
	FROM (
		SELECT invoice_date as inv_date, SUM(total_amount) as daily_total
		FROM invoices
		WHERE metadata->>'sap_customer_code' = 'C00006'
		GROUP BY invoice_date
	) d
	ON CONFLICT (session_number) DO NOTHING;
	`
	sessTag, err := pool.Exec(ctx, sessQuery, cashierID, terminalID)
	if err != nil {
		log.Fatalf("Failed to create cashier sessions: %v", err)
	}
	log.Printf("✓ Created / verified daily sessions (rows affected: %d)\n", sessTag.RowsAffected())

	// 4. Discover Months to Project
	type MonthJob struct {
		Start time.Time
		End   time.Time
		Count int64
	}

	rows, err := pool.Query(ctx, `
		SELECT 
			DATE_TRUNC('month', invoice_date)::DATE as m_start,
			(DATE_TRUNC('month', invoice_date) + interval '1 month')::DATE as m_end,
			count(*) as cnt
		FROM invoices
		WHERE metadata->>'sap_customer_code' = 'C00006'
		GROUP BY 1, 2
		ORDER BY 1 ASC;
	`)
	if err != nil {
		log.Fatalf("Failed to query months: %v", err)
	}
	defer rows.Close()

	var jobs []MonthJob
	var grandTotal int64
	for rows.Next() {
		var j MonthJob
		if err := rows.Scan(&j.Start, &j.End, &j.Count); err == nil {
			jobs = append(jobs, j)
			grandTotal += j.Count
		}
	}
	rows.Close()

	log.Printf("[STEP 2/4] Discovered %d months covering %d cash invoices to project.\n", len(jobs), grandTotal)

	// Disable stock deduction trigger during historical bulk copy to prevent redundant updates
	_, err = pool.Exec(ctx, `ALTER TABLE pos_transaction_lines DISABLE TRIGGER trg_deduct_inventory_on_pos_transaction;`)
	if err != nil {
		log.Printf("Notice: could not disable trigger trg_deduct_inventory_on_pos_transaction: %v", err)
	} else {
		log.Println("✓ Temporarily disabled trg_deduct_inventory_on_pos_transaction for high-speed projection.")
		defer func() {
			_, _ = pool.Exec(context.Background(), `ALTER TABLE pos_transaction_lines ENABLE TRIGGER trg_deduct_inventory_on_pos_transaction;`)
			log.Println("✓ Re-enabled trg_deduct_inventory_on_pos_transaction.")
		}()
	}

	// 5. Project each month
	var totalTransProjected int64
	startTime := time.Now()

	for idx, job := range jobs {
		monthStr := job.Start.Format("2006-01")
		mStart := time.Now()

		// A. Project Transactions
		txQuery := `
		INSERT INTO pos_transactions (
			store_id, cashier_id, cashier_session_id, customer_id, pos_terminal_id,
			transaction_number, transaction_date, transaction_type,
			subtotal, discount_amount, tax_amount, total_amount,
			amount_paid, change_given, status, metadata
		)
		SELECT 
			COALESCE(i.store_id, $1),
			$2,
			cs.id,
			COALESCE(i.customer_id, $3),
			$4,
			REPLACE(i.invoice_number, 'INV-SAP-', 'POS-'),
			i.invoice_date + interval '12 hours',
			'sale',
			i.subtotal,
			i.discount_amount,
			i.tax_amount,
			i.total_amount,
			i.paid_amount,
			0.00,
			'completed',
			jsonb_build_object('sap_invoice_id', i.id, 'sap_invoice_number', i.invoice_number)
		FROM invoices i
		JOIN cashier_sessions cs ON cs.session_number = 'SES-' || TO_CHAR(i.invoice_date, 'YYYYMMDD')
		WHERE i.metadata->>'sap_customer_code' = 'C00006'
		  AND i.invoice_date >= $5 AND i.invoice_date < $6
		ON CONFLICT (transaction_number) DO NOTHING;
		`
		tTag, err := pool.Exec(ctx, txQuery, storeID, cashierID, customerID, terminalID, job.Start, job.End)
		if err != nil {
			log.Fatalf("[%s] Failed to project transactions: %v", monthStr, err)
		}

		// B. Project Line Items
		linesQuery := `
		INSERT INTO pos_transaction_lines (
			transaction_id, product_id, quantity, unit_price,
			discount_amount, tax_amount, subtotal, line_total, line_number
		)
		SELECT 
			pt.id,
			COALESCE(il.product_id, $3),
			il.quantity,
			il.unit_price,
			COALESCE(il.discount_amount, 0),
			COALESCE(il.tax_amount, 0),
			COALESCE(il.line_total - il.tax_amount, il.unit_price * il.quantity),
			il.line_total,
			il.line_number
		FROM invoice_lines il
		JOIN invoices i ON il.invoice_id = i.id
		JOIN pos_transactions pt ON pt.transaction_number = REPLACE(i.invoice_number, 'INV-SAP-', 'POS-')
		WHERE i.metadata->>'sap_customer_code' = 'C00006'
		  AND i.invoice_date >= $1 AND i.invoice_date < $2
		  AND NOT EXISTS (
			  SELECT 1 FROM pos_transaction_lines ptl WHERE ptl.transaction_id = pt.id
		  );
		`
		lTag, err := pool.Exec(ctx, linesQuery, job.Start, job.End, defaultProductID)
		if err != nil {
			log.Fatalf("[%s] Failed to project transaction lines: %v", monthStr, err)
		}

		// C. Project Payments
		payQuery := `
		INSERT INTO pos_payments (
			transaction_id, payment_method, amount, payment_date
		)
		SELECT 
			pt.id,
			'cash',
			pt.total_amount,
			pt.transaction_date
		FROM pos_transactions pt
		WHERE pt.transaction_number LIKE 'POS-%'
		  AND pt.transaction_date >= $1 AND pt.transaction_date < $2
		  AND NOT EXISTS (
			  SELECT 1 FROM pos_payments pp WHERE pp.transaction_id = pt.id
		  );
		`
		pTag, err := pool.Exec(ctx, payQuery, job.Start, job.End)
		if err != nil {
			log.Fatalf("[%s] Failed to project payments: %v", monthStr, err)
		}

		totalTransProjected += tTag.RowsAffected()
		pct := float64(idx+1) / float64(len(jobs)) * 100.0
		log.Printf("[%2d/%d] Month %s: %d tickets, %d lines, %d payments projected in %v (%.1f%% overall)\n",
			idx+1, len(jobs), monthStr, tTag.RowsAffected(), lTag.RowsAffected(), pTag.RowsAffected(), time.Since(mStart).Round(time.Millisecond), pct)
	}

	log.Println("================================================================")
	log.Printf("🎉 Finished projecting %d transactions in %v!\n", totalTransProjected, time.Since(startTime).Round(time.Second))
	log.Println("================================================================")
}
