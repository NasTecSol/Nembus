package zatca

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrderAutomationDDL records committed order and line changes. A delayed job
// lets separately saved lines and totals settle before the invoice is signed.
// No existing orders are backfilled: historical sales require explicit review.
const OrderAutomationDDL = `
CREATE TABLE IF NOT EXISTS zatca_order_jobs (
 order_id uuid PRIMARY KEY REFERENCES sales_orders_v2(id) ON DELETE CASCADE,
 revision bigint NOT NULL DEFAULT 1,
 status text NOT NULL DEFAULT 'pending',
 attempts integer NOT NULL DEFAULT 0,
 next_attempt_at timestamptz NOT NULL DEFAULT now(),
 last_error text,
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_zatca_order_jobs_ready ON zatca_order_jobs(next_attempt_at) WHERE status='pending';
CREATE OR REPLACE FUNCTION enqueue_zatca_order_job() RETURNS trigger AS $$
DECLARE order_key uuid;
BEGIN
 IF TG_TABLE_NAME='sales_orders_v2' THEN
  IF TG_OP='DELETE' THEN RETURN OLD; END IF;
  order_key := NEW.id;
 ELSE
  IF TG_OP='DELETE' THEN order_key := OLD.sales_order_id; ELSE order_key := NEW.sales_order_id; END IF;
 END IF;
 IF current_setting('nembus.sync_origin',true)='cloud' THEN
  IF TG_OP='DELETE' THEN RETURN OLD; END IF;
  RETURN NEW;
 END IF;
 INSERT INTO zatca_order_jobs(order_id,next_attempt_at)
 SELECT order_key,now()+interval '5 seconds'
 WHERE EXISTS(SELECT 1 FROM sales_orders_v2 WHERE id=order_key)
 ON CONFLICT(order_id) DO UPDATE SET
  revision=zatca_order_jobs.revision+1,status='pending',attempts=0,
  next_attempt_at=now()+interval '5 seconds',last_error=NULL,updated_at=now();
 IF TG_OP='DELETE' THEN RETURN OLD; END IF;
 RETURN NEW;
END; $$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trg_zatca_order_job ON sales_orders_v2;
CREATE TRIGGER trg_zatca_order_job AFTER INSERT OR UPDATE ON sales_orders_v2 FOR EACH ROW EXECUTE FUNCTION enqueue_zatca_order_job();
DROP TRIGGER IF EXISTS trg_zatca_order_line_job ON sales_order_lines_v2;
CREATE TRIGGER trg_zatca_order_line_job AFTER INSERT OR UPDATE OR DELETE ON sales_order_lines_v2 FOR EACH ROW EXECUTE FUNCTION enqueue_zatca_order_job();
`

type OrderSubmitter interface {
	SubmitSalesOrder(context.Context, uuid.UUID) (*repository.ZatcaDocumentChain, error)
}

// ProcessOrderAutomationBatch serializes workers per tenant database. Failed
// attempts stay visible and retry later; a newer revision remains pending.
func ProcessOrderAutomationBatch(ctx context.Context, pool *pgxpool.Pool, submitter OrderSubmitter) error {
	if pool == nil || submitter == nil {
		return nil
	}
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	var locked bool
	if err := conn.QueryRow(ctx, `SELECT pg_try_advisory_lock(hashtext('nembus.zatca.order_jobs'))`).Scan(&locked); err != nil || !locked {
		return err
	}
	defer func() {
		_, _ = conn.Exec(context.Background(), `SELECT pg_advisory_unlock(hashtext('nembus.zatca.order_jobs'))`)
	}()
	rows, err := conn.Query(ctx, `SELECT order_id,revision FROM zatca_order_jobs WHERE status='pending' AND next_attempt_at<=now() ORDER BY next_attempt_at,order_id LIMIT 10`)
	if err != nil {
		return err
	}
	type job struct {
		id       uuid.UUID
		revision int64
	}
	jobs := make([]job, 0, 10)
	for rows.Next() {
		var current job
		if err := rows.Scan(&current.id, &current.revision); err != nil {
			rows.Close()
			return err
		}
		jobs = append(jobs, current)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, current := range jobs {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// POS checkout orders belong to the terminal even if their device
		// registration has not reached cloud yet. An existing POS-owned signed
		// document also stays with its signer.
		if _, cloudWorker := submitter.(*Service); cloudWorker {
			var posOwned bool
			if err := conn.QueryRow(ctx, `SELECT EXISTS(
 SELECT 1 FROM sales_orders_v2 o WHERE o.id=$1 AND o.sales_channel='pos'
 UNION ALL
 SELECT 1 FROM sales_orders_v2 o JOIN zatca_device_configs d ON d.organization_id=o.organization_id AND d.pos_terminal_id=o.pos_terminal_id
 WHERE o.id=$1 AND d.device_type='pos' AND d.metadata->>'signing_owner'='pos'
 UNION ALL
 SELECT 1 FROM zatca_document_chain c JOIN zatca_device_configs d ON d.id=c.device_config_id
 WHERE c.entity_type='sales_order_v2' AND c.entity_id=$1::text AND d.metadata->>'signing_owner'='pos')`, current.id).Scan(&posOwned); err != nil {
				return err
			}
			if posOwned {
				if _, err := conn.Exec(ctx, `UPDATE zatca_order_jobs SET status='delegated',last_error='Owning POS submits this order to ZATCA',updated_at=now() WHERE order_id=$1 AND revision=$2`, current.id, current.revision); err != nil {
					return err
				}
				continue
			}
		}
		var savedStatus string
		statusErr := conn.QueryRow(ctx, `SELECT zatca_status::text FROM zatca_document_chain WHERE entity_type='sales_order_v2' AND entity_id=$1::text ORDER BY id DESC LIMIT 1`, current.id).Scan(&savedStatus)
		if statusErr != nil && !errors.Is(statusErr, pgx.ErrNoRows) {
			return statusErr
		}
		if savedStatus == "reported" || savedStatus == "cleared" || savedStatus == "rejected" {
			if _, err := conn.Exec(ctx, `UPDATE zatca_order_jobs SET status=$3,last_error=NULL,updated_at=now() WHERE order_id=$1 AND revision=$2`, current.id, current.revision, savedStatus); err != nil {
				return err
			}
			continue
		}
		orderCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
		orderCtx = context.WithValue(orderCtx, middleware.RepoKey, repository.New(pool))
		entry, submitErr := submitter.SubmitSalesOrder(orderCtx, current.id)
		cancel()
		status, delay, message := "done", time.Duration(0), ""
		if errors.Is(submitErr, usecase.ErrZatcaDisabled) {
			status = "disabled"
		} else if entry != nil && entry.ZatcaStatus.Valid && entry.ZatcaStatus.ZatcaDocStatus == repository.ZatcaDocStatusRejected {
			status = "rejected"
			message = "ZATCA rejected the saved document; inspect zatca_document_chain.zatca_response"
		} else if submitErr != nil {
			status, delay, message = "pending", time.Minute, submitErr.Error()
			if strings.Contains(message, "owned by its POS") {
				status = "delegated"
			}
		}
		if len(message) > 500 {
			message = message[:500]
		}
		_, err := conn.Exec(ctx, `UPDATE zatca_order_jobs SET status=$3,attempts=attempts+1,next_attempt_at=now()+$4::interval,last_error=NULLIF($5,''),updated_at=now() WHERE order_id=$1 AND revision=$2`, current.id, current.revision, status, fmt.Sprintf("%d seconds", int(delay.Seconds())), message)
		if err != nil {
			return err
		}
	}
	return nil
}
