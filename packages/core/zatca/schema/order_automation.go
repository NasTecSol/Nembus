// Package schema contains ZATCA database objects without depending on the
// application packages that install or use them. This keeps tenant pool
// initialization free of import cycles.
package schema

// DDL installs the delayed order submission queue and its change triggers.
const DDL = `
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
 -- Only finalized fulfillment can create an invoice submission job. Line
 -- edits after fulfillment still refresh the job so the final invoice is used.
 IF TG_TABLE_NAME='sales_orders_v2' THEN
  IF NEW.fulfillment_status IS DISTINCT FROM 'fulfilled' THEN RETURN NEW; END IF;
 ELSE
  IF NOT EXISTS(SELECT 1 FROM sales_orders_v2 WHERE id=order_key AND fulfillment_status='fulfilled') THEN RETURN NEW; END IF;
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
