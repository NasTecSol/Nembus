-- name: GetCashierSessionZReport :one
-- One statement gives every section the same PostgreSQL snapshot. Aggregate
-- transactions, tenders and refunds independently to avoid split-payment fanout.
WITH session_scope AS (
    SELECT cs.id, cs.cashier_id, cs.pos_terminal_id, cs.session_number,
           cs.opening_time, cs.closing_time, cs.opening_balance, cs.closing_balance,
           cs.expected_balance, cs.status, t.store_id, t.terminal_code, t.terminal_name,
           s.name AS store_name, s.timezone, o.currency_code,
           c.cashier_code, concat_ws(' ', u.first_name, u.last_name) AS cashier_name
    FROM cashier_sessions cs
    JOIN pos_terminals t ON t.id = cs.pos_terminal_id
    JOIN stores s ON s.id = t.store_id
    JOIN organizations o ON o.id = s.organization_id
    JOIN cashiers c ON c.id = cs.cashier_id AND c.store_id = s.id
    JOIN users u ON u.id = c.user_id
    WHERE cs.id = sqlc.arg('cashier_session_id')::int
      AND s.organization_id = sqlc.arg('organization_id')::int
      AND EXISTS (
          SELECT 1 FROM users viewer
          JOIN user_store_access access ON access.user_id = viewer.id
          WHERE viewer.id = sqlc.arg('user_id')::int
            AND viewer.organization_id = s.organization_id AND viewer.is_active
            AND access.store_id = s.id
      )
      AND (sqlc.arg('store_id')::int = 0 OR s.id = sqlc.arg('store_id')::int)
      AND (sqlc.arg('pos_terminal_id')::int = 0 OR t.id = sqlc.arg('pos_terminal_id')::int)
      AND (sqlc.arg('cashier_id')::int = 0 OR c.id = sqlc.arg('cashier_id')::int)
), session_transactions AS (
    SELECT pt.id, pt.status, pt.transaction_type, pt.total_amount, pt.tax_amount,
           pt.discount_amount, pt.change_given, pt.metadata
    FROM pos_transactions pt
    JOIN session_scope ss ON ss.id = pt.cashier_session_id
    WHERE pt.store_id = ss.store_id AND pt.cashier_id = ss.cashier_id
      AND (pt.pos_terminal_id IS NULL OR pt.pos_terminal_id = ss.pos_terminal_id)
), sales AS (
    SELECT * FROM session_transactions
    WHERE lower(status) IN ('completed', 'partially_refunded', 'refunded')
      AND lower(COALESCE(transaction_type, 'sale')) = 'sale'
), sales_totals AS (
    SELECT count(*) AS transaction_count,
           COALESCE(sum(total_amount - COALESCE(tax_amount, 0)), 0) AS excluding_tax,
           COALESCE(sum(discount_amount), 0) AS discounts,
           COALESCE(sum(tax_amount), 0) AS tax,
           COALESCE(sum(total_amount), 0) AS total
    FROM sales
), void_totals AS (
    SELECT count(*) AS transaction_count, COALESCE(sum(total_amount), 0) AS total
    FROM session_transactions WHERE lower(status) = 'voided'
), payment_rows AS (
    SELECT pp.transaction_id, COALESCE(NULLIF(lower(trim(pp.payment_method)), ''), 'unknown') AS method,
           pp.amount,
           CASE WHEN pp.metadata->>'change_given' ~ '^[0-9]+([.][0-9]+)?$'
                THEN (pp.metadata->>'change_given')::numeric ELSE 0 END AS metadata_change
    FROM pos_payments pp JOIN sales ON sales.id = pp.transaction_id
), transaction_payments AS (
    SELECT sales.id, COALESCE(sales.total_amount, 0) AS sale_total,
           count(pr.transaction_id) AS payment_count,
           COALESCE(sum(pr.amount), 0) AS paid,
           COALESCE(sum(pr.amount) FILTER (WHERE pr.method = 'cash'), 0) AS cash,
           CASE WHEN count(*) FILTER (WHERE pr.method = 'cash') > 0 THEN
             COALESCE(NULLIF(sales.change_given, 0),
               NULLIF(sum(pr.metadata_change) FILTER (WHERE pr.method = 'cash'), 0),
               CASE WHEN sales.metadata->>'change_given' ~ '^[0-9]+([.][0-9]+)?$'
                    THEN (sales.metadata->>'change_given')::numeric END, 0)
           ELSE 0 END AS change
    FROM sales LEFT JOIN payment_rows pr ON pr.transaction_id = sales.id
    GROUP BY sales.id, sales.total_amount, sales.change_given, sales.metadata
), cash_adjustments AS (
    -- Existing writers can store applied amounts or tendered amounts. When
    -- payments already equal the sale, reconstruct tendered cash before change.
    SELECT *, CASE WHEN paid = sale_total AND cash > 0 THEN change ELSE 0 END AS tender_adjustment
    FROM transaction_payments
), completed_returns AS (
    SELECT sr.total_refund_amount, sr.tax_amount,
           COALESCE(NULLIF(lower(trim(sr.refund_method)), ''), 'unknown') AS method
    FROM sales_returns sr JOIN session_scope ss ON ss.id = sr.cashier_session_id
    WHERE sr.status = 'completed' AND sr.store_id = ss.store_id
      AND (sr.cashier_id IS NULL OR sr.cashier_id = ss.cashier_id)
), return_totals AS (
    SELECT count(*) AS return_count,
           COALESCE(sum(total_refund_amount - COALESCE(tax_amount, 0)), 0) AS excluding_tax,
           COALESCE(sum(tax_amount), 0) AS tax,
           COALESCE(sum(total_refund_amount), 0) AS total
    FROM completed_returns
), methods AS (
    SELECT method FROM payment_rows UNION SELECT method FROM completed_returns
), payment_methods AS (
    SELECT m.method,
           (SELECT count(*) FROM payment_rows pr WHERE pr.method = m.method) AS payment_count,
           COALESCE((SELECT sum(amount) FROM payment_rows pr WHERE pr.method = m.method), 0)
             + CASE WHEN m.method = 'cash' THEN COALESCE((SELECT sum(tender_adjustment) FROM cash_adjustments), 0) ELSE 0 END AS received,
           CASE WHEN m.method = 'cash' THEN COALESCE((SELECT sum(change) FROM cash_adjustments), 0) ELSE 0 END AS change,
           COALESCE((SELECT sum(total_refund_amount) FROM completed_returns cr WHERE cr.method = m.method), 0) AS refunded
    FROM methods m
), payment_totals AS (
    SELECT COALESCE(sum(received), 0) AS received, COALESCE(sum(change), 0) AS change,
           COALESCE(sum(refunded), 0) AS refunded,
           COALESCE(sum(received - change - refunded), 0) AS net
    FROM payment_methods
), cash_totals AS (
    SELECT COALESCE(sum(received), 0) AS received, COALESCE(sum(change), 0) AS change,
           COALESCE(sum(refunded), 0) AS refunded
    FROM payment_methods WHERE method = 'cash'
), drawer AS (
    SELECT COALESCE(ss.opening_balance, 0) + ct.received - ct.change - ct.refunded AS expected
    FROM session_scope ss CROSS JOIN cash_totals ct
), warnings AS (
    SELECT 'MANUAL_CASH_MOVEMENTS_NOT_INCLUDED' AS code,
           'Manual cash additions and removals are not included.' AS message
    UNION ALL SELECT 'MISSING_PAYMENTS', 'Some posted sales have no payment records.'
      WHERE EXISTS (SELECT 1 FROM transaction_payments WHERE payment_count = 0)
    UNION ALL SELECT 'PAYMENT_MISMATCH', 'Some posted sales do not reconcile with their recorded payments.'
      WHERE EXISTS (SELECT 1 FROM cash_adjustments WHERE paid + tender_adjustment - change <> sale_total)
    UNION ALL SELECT 'UNKNOWN_PAYMENT_METHOD', 'Some payments or refunds have no recorded payment method.'
      WHERE EXISTS (SELECT 1 FROM methods WHERE method = 'unknown')
    UNION ALL SELECT 'EXPECTED_CASH_MISMATCH', 'Recorded session expected cash differs from the calculated amount.'
      WHERE EXISTS (SELECT 1 FROM session_scope ss CROSS JOIN drawer d WHERE ss.expected_balance <> d.expected)
    UNION ALL SELECT 'INCONSISTENT_SESSION_RECORDS', 'Some transactions or returns have inconsistent store, terminal or cashier references and were excluded.'
      WHERE EXISTS (
        SELECT 1 FROM pos_transactions pt JOIN session_scope ss ON ss.id = pt.cashier_session_id
        WHERE pt.store_id <> ss.store_id OR pt.cashier_id <> ss.cashier_id
          OR (pt.pos_terminal_id IS NOT NULL AND pt.pos_terminal_id <> ss.pos_terminal_id)
        UNION ALL
        SELECT 1 FROM sales_returns sr JOIN session_scope ss ON ss.id = sr.cashier_session_id
        WHERE sr.store_id <> ss.store_id OR (sr.cashier_id IS NOT NULL AND sr.cashier_id <> ss.cashier_id)
      )
)
SELECT jsonb_build_object(
    'report_type', 'Z', 'finalized', COALESCE(lower(ss.status) = 'closed' AND ss.closing_time IS NOT NULL, false),
    'generated_at', statement_timestamp(), 'calculation_version', 1, 'currency', ss.currency_code,
    'store', jsonb_build_object('id', ss.store_id, 'name', ss.store_name, 'timezone', ss.timezone),
    'terminal', jsonb_build_object('id', ss.pos_terminal_id, 'code', ss.terminal_code, 'name', ss.terminal_name),
    'cashier', jsonb_build_object('id', ss.cashier_id, 'code', ss.cashier_code, 'name', ss.cashier_name),
    'session', jsonb_build_object('id', ss.id, 'number', ss.session_number, 'status', ss.status,
        'opened_at', ss.opening_time, 'closed_at', ss.closing_time),
    'sales', jsonb_build_object('transaction_count', st.transaction_count,
        'before_discount_excluding_tax', (st.excluding_tax + st.discounts)::numeric(20,2)::text,
        'discounts', st.discounts::numeric(20,2)::text,
        'after_discount_excluding_tax', st.excluding_tax::numeric(20,2)::text,
        'tax', st.tax::numeric(20,2)::text, 'total_including_tax', st.total::numeric(20,2)::text),
    'voids', jsonb_build_object('transaction_count', vt.transaction_count, 'total_including_tax', vt.total::numeric(20,2)::text),
    'returns', jsonb_build_object('return_count', rt.return_count, 'excluding_tax', rt.excluding_tax::numeric(20,2)::text,
        'tax', rt.tax::numeric(20,2)::text, 'total_refunded', rt.total::numeric(20,2)::text),
    'net', jsonb_build_object('sales_excluding_tax', (st.excluding_tax - rt.excluding_tax)::numeric(20,2)::text,
        'tax', (st.tax - rt.tax)::numeric(20,2)::text, 'total_including_tax', (st.total - rt.total)::numeric(20,2)::text),
    'payments', COALESCE((SELECT jsonb_agg(jsonb_build_object(
        'method', pm.method, 'payment_count', pm.payment_count, 'received', pm.received::numeric(20,2)::text,
        'change_given', pm.change::numeric(20,2)::text, 'refunded', pm.refunded::numeric(20,2)::text,
        'net', (pm.received - pm.change - pm.refunded)::numeric(20,2)::text) ORDER BY pm.method) FROM payment_methods pm), '[]'::jsonb),
    'payment_totals', jsonb_build_object('received', pt.received::numeric(20,2)::text,
        'change_given', pt.change::numeric(20,2)::text, 'refunded', pt.refunded::numeric(20,2)::text, 'net', pt.net::numeric(20,2)::text),
    'drawer', jsonb_build_object('opening_cash', COALESCE(ss.opening_balance, 0)::numeric(20,2)::text,
        'cash_received', ct.received::numeric(20,2)::text, 'change_given', ct.change::numeric(20,2)::text,
        'cash_refunded', ct.refunded::numeric(20,2)::text, 'manual_movements_included', false,
        'calculated_expected_cash', d.expected::numeric(20,2)::text,
        'recorded_expected_cash', ss.expected_balance::numeric(20,2)::text,
        'counted_closing_cash', CASE WHEN lower(ss.status) = 'closed' AND ss.closing_time IS NOT NULL THEN ss.closing_balance::numeric(20,2)::text END,
        'variance', CASE WHEN lower(ss.status) = 'closed' AND ss.closing_time IS NOT NULL THEN (ss.closing_balance - d.expected)::numeric(20,2)::text END),
    'reconciliation', jsonb_build_object('sales_payment_difference', (st.total - rt.total - pt.net)::numeric(20,2)::text,
        'expected_cash_difference', (ss.expected_balance - d.expected)::numeric(20,2)::text,
        'warnings', (SELECT jsonb_agg(jsonb_build_object('code', w.code, 'message', w.message) ORDER BY w.code) FROM warnings w))
)::jsonb AS report
FROM session_scope ss CROSS JOIN sales_totals st CROSS JOIN void_totals vt CROSS JOIN return_totals rt
CROSS JOIN payment_totals pt CROSS JOIN cash_totals ct CROSS JOIN drawer d;
