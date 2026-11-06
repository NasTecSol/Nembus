package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestLocalCustomerCreationQueuesExistingOutbox(t *testing.T) {
	dsn := os.Getenv("NEMBUS_SYNC_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_SYNC_TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("pos_customer_handler_%d", time.Now().UnixNano())
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE") }()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, `
		CREATE TABLE local_device_config (id serial PRIMARY KEY, pos_terminal_id integer);
		INSERT INTO local_device_config(pos_terminal_id) VALUES(3);
		CREATE TABLE customers (
			id serial PRIMARY KEY, organization_id integer NOT NULL,
			customer_code varchar(50) NOT NULL, name varchar(255) NOT NULL,
			email text, phone text, address text, customer_type text,
			price_list_id integer, credit_limit numeric DEFAULT 0,
			outstanding_balance numeric DEFAULT 0, loyalty_points numeric DEFAULT 0,
			loyalty_tier text DEFAULT 'bronze', is_active boolean DEFAULT true,
			metadata jsonb DEFAULT '{}', created_at timestamp DEFAULT now(),
			updated_at timestamp DEFAULT now(), UNIQUE(organization_id,customer_code));
		CREATE TABLE sync_queue (entity_type text, entity_id bigint, action text,
			payload jsonb, status text, priority integer);`); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	h := NewCustomerHandler(usecase.NewCustomerUseCase())
	h.EnableLocalCreationSync()
	r := gin.New()
	r.POST("/customers", func(c *gin.Context) {
		requestCtx := context.WithValue(c.Request.Context(), middleware.RepoKey, repository.New(pool))
		requestCtx = context.WithValue(requestCtx, middleware.PoolKey, pool)
		c.Request = c.Request.WithContext(requestCtx)
		h.CreateCustomer(c)
	})
	create := func(code string) int {
		t.Helper()
		body, _ := json.Marshal(map[string]any{"organization_id": 1, "customer_code": code, "name": "POS customer"})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/customers", bytes.NewReader(body)))
		return w.Code
	}
	if got := create("POS-ONE"); got != http.StatusCreated {
		t.Fatalf("first customer HTTP status: %d", got)
	}
	var id int32
	if err := pool.QueryRow(ctx, `SELECT id FROM customers WHERE customer_code='POS-ONE'`).Scan(&id); err != nil || id != 1000300001 {
		t.Fatalf("first customer ID: %d, err=%v", id, err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM sync_queue WHERE entity_type='customers'
		AND entity_id=$1 AND action='INSERT' AND status='pending' AND payload->>'customer_code'='POS-ONE'`, id).Scan(&count); err != nil || count != 1 {
		t.Fatalf("customer outbox event: count=%d err=%v", count, err)
	}
	if _, err := pool.Exec(ctx, `DROP TABLE sync_queue`); err != nil {
		t.Fatal(err)
	}
	if got := create("POS-TWO"); got != http.StatusInternalServerError {
		t.Fatalf("failed outbox should fail creation: HTTP %d", got)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM customers WHERE customer_code='POS-TWO'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("failed outbox left a customer row: count=%d err=%v", count, err)
	}
}
