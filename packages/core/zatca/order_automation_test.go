package zatca

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type jobSubmitter struct {
	pool  *pgxpool.Pool
	calls int
	order uuid.UUID
}

func (s *jobSubmitter) SubmitSalesOrder(ctx context.Context, id uuid.UUID) (*repository.ZatcaDocumentChain, error) {
	if id != s.order || ctx.Value(middleware.RepoKey) == nil {
		return nil, fmt.Errorf("wrong order or missing tenant repository")
	}
	s.calls++
	switch s.calls {
	case 1:
		return nil, errors.New("order totals have not been saved")
	case 2:
		// A line/header change committed during submission must remain queued.
		if _, err := s.pool.Exec(ctx, `UPDATE sales_orders_v2 SET marker=1 WHERE id=$1`, id); err != nil {
			return nil, err
		}
	}
	return &repository.ZatcaDocumentChain{}, nil
}

func TestOrderAutomationPostgres(t *testing.T) {
	dsn := os.Getenv("NEMBUS_ZATCA_ORDER_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_ZATCA_ORDER_TEST_DATABASE_URL for isolated PostgreSQL job test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("zatca_order_job_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
		t.Fatal(err)
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		_, _ = admin.Exec(context.Background(), "DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE")
		admin.Close()
	})
	_, err = pool.Exec(ctx, `CREATE TABLE sales_orders_v2(id uuid PRIMARY KEY,marker int DEFAULT 0,organization_id int DEFAULT 1,pos_terminal_id int DEFAULT 8,sales_channel text DEFAULT 'pos');
CREATE TABLE sales_order_lines_v2(id uuid PRIMARY KEY,sales_order_id uuid NOT NULL REFERENCES sales_orders_v2);
CREATE TABLE zatca_device_configs(id serial PRIMARY KEY,organization_id int,pos_terminal_id int,device_type text,metadata jsonb);
CREATE TABLE zatca_document_chain(id bigserial PRIMARY KEY,entity_type text,entity_id text,device_config_id int REFERENCES zatca_device_configs,zatca_status text);
INSERT INTO sales_orders_v2(id) VALUES('00000000-0000-0000-0000-000000000001');`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, OrderAutomationDDL); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM zatca_order_jobs`).Scan(&count); err != nil || count != 0 {
		t.Fatal("historical orders were queued")
	}
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO sales_orders_v2(id) VALUES($1)`, id); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO sales_order_lines_v2(id,sales_order_id) VALUES($1,$2)`, uuid.New(), id); err != nil {
		t.Fatal(err)
	}
	var revision int64
	var due bool
	if err := pool.QueryRow(ctx, `SELECT revision,next_attempt_at<=now() FROM zatca_order_jobs WHERE order_id=$1`, id).Scan(&revision, &due); err != nil || revision != 2 || due {
		t.Fatalf("separate header/line changes were not debounced: revision=%d due=%t err=%v", revision, due, err)
	}
	submitter := &jobSubmitter{pool: pool, order: id}
	if err := ProcessOrderAutomationBatch(ctx, pool, submitter); err != nil || submitter.calls != 0 {
		t.Fatal("worker processed an incomplete write")
	}
	if _, err := pool.Exec(ctx, `UPDATE zatca_order_jobs SET next_attempt_at=now()-interval '1 second' WHERE order_id=$1`, id); err != nil {
		t.Fatal(err)
	}
	if err := ProcessOrderAutomationBatch(ctx, pool, submitter); err != nil || submitter.calls != 1 {
		t.Fatal("first attempt did not record a retry")
	}
	var status, lastError string
	if err := pool.QueryRow(ctx, `SELECT status,last_error FROM zatca_order_jobs WHERE order_id=$1`, id).Scan(&status, &lastError); err != nil || status != "pending" || lastError == "" {
		t.Fatal("failed order was lost")
	}
	if _, err := pool.Exec(ctx, `INSERT INTO sales_order_lines_v2(id,sales_order_id) VALUES($1,$2)`, uuid.New(), id); err != nil {
		t.Fatal(err)
	}
	for expected := 2; expected <= 3; expected++ {
		if _, err := pool.Exec(ctx, `UPDATE zatca_order_jobs SET next_attempt_at=now()-interval '1 second' WHERE order_id=$1`, id); err != nil {
			t.Fatal(err)
		}
		if err := ProcessOrderAutomationBatch(ctx, pool, submitter); err != nil || submitter.calls != expected {
			t.Fatalf("attempt %d failed: %v", expected, err)
		}
	}
	if err := pool.QueryRow(ctx, `SELECT status FROM zatca_order_jobs WHERE order_id=$1`, id).Scan(&status); err != nil || status != "done" {
		t.Fatal("latest revision was not acknowledged")
	}
	cloudOrder := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO sales_orders_v2(id) VALUES($1)`, cloudOrder); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE zatca_order_jobs SET next_attempt_at=now()-interval '1 second' WHERE order_id=$1`, cloudOrder); err != nil {
		t.Fatal(err)
	}
	if err := ProcessOrderAutomationBatch(ctx, pool, NewService(nil, &usecase.ZatcaConfig{})); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT status FROM zatca_order_jobs WHERE order_id=$1`, cloudOrder).Scan(&status); err != nil || status != "delegated" {
		t.Fatal("cloud attempted to sign a POS sale before its device registration arrived")
	}
}
