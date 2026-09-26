-- +goose Up
-- Migration: 20260921000000_financial_inventory_reports_views
-- Description: Adds 6 core financial and inventory analytical views:
--   1. vw_realtime_pnl
--   2. vw_cashier_shift_reconciliation
--   3. vw_stock_movement_ledger
--   4. vw_vendor_performance_ppv
--   5. vw_supplier_aging_report
--   6. vw_tax_vat_summary

-- ---------------------------------------------------------------------
-- 1. Real-Time Operational P&L & COGS View
-- ---------------------------------------------------------------------
CREATE OR REPLACE VIEW vw_realtime_pnl AS
WITH daily_pos_sales AS (
    SELECT
        pt.store_id,
        s.organization_id,
        s.name AS store_name,
        pt.transaction_date::DATE AS period_date,
        COUNT(DISTINCT pt.id) AS transaction_count,
        COALESCE(SUM(ptl.line_total), 0) AS gross_sales,
        COALESCE(SUM(ptl.discount_amount), 0) AS total_discounts,
        COALESCE(SUM(ptl.tax_amount), 0) AS total_tax,
        COALESCE(SUM(ptl.cost_price * ptl.quantity), 0) AS total_cogs
    FROM pos_transactions pt
    JOIN stores s ON s.id = pt.store_id
    JOIN pos_transaction_lines ptl ON ptl.transaction_id = pt.id
    WHERE pt.status = 'completed'
    GROUP BY pt.store_id, s.organization_id, s.name, pt.transaction_date::DATE
),
daily_returns AS (
    SELECT
        sr.store_id,
        sr.return_date::DATE AS period_date,
        COUNT(DISTINCT sr.id) AS return_count,
        COALESCE(SUM(sr.total_refund_amount), 0) AS total_returns
    FROM sales_returns sr
    WHERE sr.status = 'completed'
    GROUP BY sr.store_id, sr.return_date::DATE
),
daily_invoices AS (
    SELECT
        inv.store_id,
        inv.organization_id,
        s.name AS store_name,
        inv.invoice_date AS period_date,
        COUNT(DISTINCT inv.id) AS invoice_count,
        COALESCE(SUM(inv.total_amount), 0) AS invoice_gross,
        COALESCE(SUM(inv.tax_amount), 0) AS invoice_tax,
        COALESCE(SUM(il_cost.total_cost), 0) AS invoice_cogs
    FROM invoices inv
    JOIN stores s ON s.id = inv.store_id
    LEFT JOIN (
        SELECT 
            il.invoice_id,
            SUM(COALESCE(p.cost_price, 0) * il.quantity) AS total_cost
        FROM invoice_lines il
        LEFT JOIN products p ON p.id = il.product_id
        GROUP BY il.invoice_id
    ) il_cost ON il_cost.invoice_id = inv.id
    WHERE inv.invoice_status NOT IN ('draft', 'cancelled')
    GROUP BY inv.store_id, inv.organization_id, s.name, inv.invoice_date
),
all_periods AS (
    SELECT store_id, organization_id, store_name, period_date FROM daily_pos_sales
    UNION
    SELECT store_id, organization_id, store_name, period_date FROM daily_invoices
)
SELECT
    ap.organization_id,
    ap.store_id,
    ap.store_name,
    ap.period_date,
    DATE_TRUNC('month', ap.period_date)::DATE AS period_month,
    COALESCE(ps.transaction_count, 0) + COALESCE(di.invoice_count, 0) AS total_transactions,
    COALESCE(ps.gross_sales, 0) + COALESCE(di.invoice_gross, 0) AS gross_revenue,
    COALESCE(ps.total_discounts, 0) AS total_discounts,
    COALESCE(ret.total_returns, 0) AS total_returns,
    (COALESCE(ps.gross_sales, 0) + COALESCE(di.invoice_gross, 0) - COALESCE(ps.total_discounts, 0) - COALESCE(ret.total_returns, 0)) AS net_revenue,
    COALESCE(ps.total_cogs, 0) + COALESCE(di.invoice_cogs, 0) AS cogs,
    (COALESCE(ps.gross_sales, 0) + COALESCE(di.invoice_gross, 0) - COALESCE(ps.total_discounts, 0) - COALESCE(ret.total_returns, 0))
        - (COALESCE(ps.total_cogs, 0) + COALESCE(di.invoice_cogs, 0)) AS gross_profit,
    CASE
        WHEN (COALESCE(ps.gross_sales, 0) + COALESCE(di.invoice_gross, 0) - COALESCE(ps.total_discounts, 0) - COALESCE(ret.total_returns, 0)) > 0
        THEN ROUND(
            ((COALESCE(ps.gross_sales, 0) + COALESCE(di.invoice_gross, 0) - COALESCE(ps.total_discounts, 0) - COALESCE(ret.total_returns, 0))
             - (COALESCE(ps.total_cogs, 0) + COALESCE(di.invoice_cogs, 0)))
            / (COALESCE(ps.gross_sales, 0) + COALESCE(di.invoice_gross, 0) - COALESCE(ps.total_discounts, 0) - COALESCE(ret.total_returns, 0)) * 100, 2
        )
        ELSE 0
    END AS gross_margin_pct,
    COALESCE(ps.total_tax, 0) + COALESCE(di.invoice_tax, 0) AS total_tax_collected
FROM all_periods ap
LEFT JOIN daily_pos_sales ps ON ps.store_id = ap.store_id AND ps.period_date = ap.period_date
LEFT JOIN daily_returns ret ON ret.store_id = ap.store_id AND ret.period_date = ap.period_date
LEFT JOIN daily_invoices di ON di.store_id = ap.store_id AND di.period_date = ap.period_date;

-- ---------------------------------------------------------------------
-- 2. Cashier Shift Drawer & Tender Reconciliation View
-- ---------------------------------------------------------------------
CREATE OR REPLACE VIEW vw_cashier_shift_reconciliation AS
WITH payments_summary AS (
    SELECT
        pt.cashier_session_id,
        COUNT(DISTINCT pt.id) AS transaction_count,
        COALESCE(SUM(pt.total_amount), 0) AS total_sales,
        COALESCE(SUM(CASE WHEN LOWER(p.payment_method) = 'cash' THEN p.amount ELSE 0 END), 0) AS cash_collected,
        COALESCE(SUM(CASE WHEN LOWER(p.payment_method) IN ('card', 'credit_card', 'debit_card') THEN p.amount ELSE 0 END), 0) AS card_collected,
        COALESCE(SUM(CASE WHEN LOWER(p.payment_method) NOT IN ('cash', 'card', 'credit_card', 'debit_card') THEN p.amount ELSE 0 END), 0) AS other_tender_collected
    FROM pos_transactions pt
    LEFT JOIN pos_payments p ON p.transaction_id = pt.id
    WHERE pt.status = 'completed'
      AND pt.cashier_session_id IS NOT NULL
    GROUP BY pt.cashier_session_id
),
refunds_summary AS (
    SELECT
        sr.cashier_session_id,
        COUNT(DISTINCT sr.id) AS return_count,
        COALESCE(SUM(CASE WHEN LOWER(COALESCE(sr.refund_method, 'cash')) = 'cash' THEN sr.total_refund_amount ELSE 0 END), 0) AS cash_refunds,
        COALESCE(SUM(sr.total_refund_amount), 0) AS total_refunds
    FROM sales_returns sr
    WHERE sr.status = 'completed'
      AND sr.cashier_session_id IS NOT NULL
    GROUP BY sr.cashier_session_id
)
SELECT
    cs.id AS session_id,
    cs.session_number,
    t.store_id,
    s.name AS store_name,
    cs.pos_terminal_id,
    t.terminal_code,
    t.terminal_name,
    cs.cashier_id,
    c.cashier_code,
    u.first_name || ' ' || u.last_name AS cashier_name,
    cs.opening_time,
    cs.closing_time,
    cs.status,
    COALESCE(cs.opening_balance, 0) AS opening_balance,
    COALESCE(ps.cash_collected, 0) AS cash_collected,
    COALESCE(ps.card_collected, 0) AS card_collected,
    COALESCE(ps.other_tender_collected, 0) AS other_tender_collected,
    COALESCE(rs.cash_refunds, 0) AS cash_refunds,
    (COALESCE(cs.opening_balance, 0) + COALESCE(ps.cash_collected, 0) - COALESCE(rs.cash_refunds, 0)) AS expected_closing_cash,
    cs.closing_balance AS actual_closing_cash,
    COALESCE(
        cs.variance,
        cs.closing_balance - (COALESCE(cs.opening_balance, 0) + COALESCE(ps.cash_collected, 0) - COALESCE(rs.cash_refunds, 0))
    ) AS over_short_variance,
    COALESCE(ps.total_sales, 0) AS total_sales,
    COALESCE(ps.transaction_count, 0) AS transaction_count,
    COALESCE(rs.total_refunds, 0) AS total_refunds,
    COALESCE(rs.return_count, 0) AS return_count
FROM cashier_sessions cs
JOIN pos_terminals t ON t.id = cs.pos_terminal_id
JOIN stores s ON s.id = t.store_id
JOIN cashiers c ON c.id = cs.cashier_id
JOIN users u ON u.id = c.user_id
LEFT JOIN payments_summary ps ON ps.cashier_session_id = cs.id
LEFT JOIN refunds_summary rs ON rs.cashier_session_id = cs.id;

-- ---------------------------------------------------------------------
-- 3. Stock Movement Running Ledger & Inventory Valuation View
-- ---------------------------------------------------------------------
CREATE OR REPLACE VIEW vw_stock_movement_ledger AS
SELECT
    sm.id AS movement_id,
    sm.movement_date,
    COALESCE(sm.to_store_id, sm.from_store_id) AS store_id,
    s.name AS store_name,
    s.organization_id,
    sm.product_id,
    p.sku,
    p.name AS product_name,
    p.barcode,
    pc.name AS category_name,
    sm.movement_type,
    sm.reference_type,
    sm.reference_id,
    sm.batch_number,
    sm.serial_number,
    CASE WHEN sm.quantity > 0 THEN sm.quantity ELSE 0 END AS quantity_in,
    CASE WHEN sm.quantity < 0 THEN ABS(sm.quantity) ELSE 0 END AS quantity_out,
    sm.quantity AS net_quantity,
    COALESCE(sm.cost_per_unit, p.cost_price, 0) AS unit_cost,
    COALESCE(sm.total_value, sm.quantity * COALESCE(sm.cost_per_unit, p.cost_price, 0)) AS movement_value,
    SUM(sm.quantity) OVER (
        PARTITION BY COALESCE(sm.to_store_id, sm.from_store_id), sm.product_id
        ORDER BY sm.movement_date, sm.id
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) AS running_quantity,
    SUM(sm.quantity) OVER (
        PARTITION BY COALESCE(sm.to_store_id, sm.from_store_id), sm.product_id
        ORDER BY sm.movement_date, sm.id
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) * COALESCE(sm.cost_per_unit, p.cost_price, 0) AS running_valuation,
    sm.status,
    sm.metadata
FROM stock_movements sm
JOIN products p ON p.id = sm.product_id
LEFT JOIN product_categories pc ON pc.id = p.category_id
LEFT JOIN stores s ON s.id = COALESCE(sm.to_store_id, sm.from_store_id);

-- ---------------------------------------------------------------------
-- 4. Vendor Fulfillment & Purchase Price Variance (PPV) View
-- ---------------------------------------------------------------------
CREATE OR REPLACE VIEW vw_vendor_performance_ppv AS
SELECT
    sup.id AS supplier_id,
    sup.code AS supplier_code,
    sup.name AS supplier_name,
    po.organization_id,
    po.id AS po_id,
    po.po_number,
    po.po_date,
    po.expected_delivery_date,
    grn.id AS grn_id,
    grn.grn_number,
    grn.receipt_date,
    p.id AS product_id,
    p.sku,
    p.name AS product_name,
    pol.quantity AS ordered_quantity,
    grni.quantity_received,
    grni.quantity_rejected,
    ROUND((grni.quantity_received / NULLIF(pol.quantity, 0)) * 100, 2) AS fill_rate_pct,
    CASE 
        WHEN po.expected_delivery_date IS NOT NULL AND grn.receipt_date::DATE <= po.expected_delivery_date::DATE THEN true
        WHEN po.expected_delivery_date IS NULL THEN true
        ELSE false
    END AS is_on_time,
    pol.unit_price AS po_unit_price,
    COALESCE(grni.unit_cost, pol.unit_price) AS grn_unit_cost,
    (COALESCE(grni.unit_cost, pol.unit_price) - pol.unit_price) AS unit_price_variance,
    ROUND((COALESCE(grni.unit_cost, pol.unit_price) - pol.unit_price) * grni.quantity_received, 2) AS total_price_variance,
    CASE
        WHEN (COALESCE(grni.unit_cost, pol.unit_price) - pol.unit_price) > 0 THEN 'unfavorable'
        WHEN (COALESCE(grni.unit_cost, pol.unit_price) - pol.unit_price) < 0 THEN 'favorable'
        ELSE 'at_cost'
    END AS variance_status
FROM goods_receipt_notes grn
JOIN goods_receipt_note_items grni ON grni.grn_id = grn.id
JOIN purchase_orders po ON po.id = grn.purchase_order_id
LEFT JOIN purchase_order_lines pol ON pol.id = grni.purchase_order_line_id
JOIN suppliers sup ON sup.id = grn.supplier_id
JOIN products p ON p.id = grni.product_id;

-- ---------------------------------------------------------------------
-- 5. Supplier Accounts Payable Aging Report View
-- ---------------------------------------------------------------------
CREATE OR REPLACE VIEW vw_supplier_aging_report AS
WITH po_balances AS (
    SELECT
        po.id AS po_id,
        po.supplier_id,
        po.organization_id,
        COALESCE(po.expected_delivery_date, (po.po_date + INTERVAL '30 days')::DATE) AS due_date,
        (
            SELECT COALESCE(SUM(pol.received_quantity * pol.unit_price), 0)
            FROM purchase_order_lines pol
            WHERE pol.purchase_order_id = po.id
        ) - COALESCE(NULLIF(po.metadata->>'amount_paid', '')::numeric, 0) AS balance_due,
        po.status
    FROM purchase_orders po
    WHERE po.status IN ('partially_received', 'received', 'approved')
)
SELECT
    sup.id AS supplier_id,
    sup.code AS supplier_code,
    sup.name AS supplier_name,
    sup.email,
    sup.phone,
    sup.payment_terms,
    sup.credit_limit,
    pb.organization_id,
    COALESCE(SUM(CASE WHEN pb.due_date >= CURRENT_DATE THEN pb.balance_due ELSE 0 END), 0) AS current_amount,
    COALESCE(SUM(CASE WHEN pb.due_date < CURRENT_DATE AND CURRENT_DATE - pb.due_date <= 30 THEN pb.balance_due ELSE 0 END), 0) AS overdue_1_30,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - pb.due_date BETWEEN 31 AND 60 THEN pb.balance_due ELSE 0 END), 0) AS overdue_31_60,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - pb.due_date BETWEEN 61 AND 90 THEN pb.balance_due ELSE 0 END), 0) AS overdue_61_90,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - pb.due_date > 90 THEN pb.balance_due ELSE 0 END), 0) AS overdue_over_90,
    COALESCE(SUM(CASE WHEN pb.balance_due > 0 THEN pb.balance_due ELSE 0 END), 0) AS total_outstanding,
    COUNT(CASE WHEN pb.due_date < CURRENT_DATE AND pb.balance_due > 0 THEN 1 END)::INTEGER AS overdue_po_count,
    MAX(pb.due_date) AS latest_due_date
FROM suppliers sup
LEFT JOIN po_balances pb
    ON pb.supplier_id = sup.id
    AND pb.balance_due > 0
WHERE sup.is_active = true
GROUP BY
    sup.id, sup.code, sup.name, sup.email, sup.phone,
    sup.payment_terms, sup.credit_limit, pb.organization_id
ORDER BY total_outstanding DESC;

-- ---------------------------------------------------------------------
-- 6. Tax / VAT Reconciliation Summary View
-- ---------------------------------------------------------------------
CREATE OR REPLACE VIEW vw_tax_vat_summary AS
WITH output_vat AS (
    SELECT
        s.organization_id,
        pt.store_id,
        DATE_TRUNC('month', pt.transaction_date)::DATE AS period_month,
        COALESCE(SUM(pt.subtotal), 0) AS taxable_sales,
        COALESCE(SUM(pt.tax_amount), 0) AS output_vat
    FROM pos_transactions pt
    JOIN stores s ON s.id = pt.store_id
    WHERE pt.status = 'completed'
    GROUP BY s.organization_id, pt.store_id, DATE_TRUNC('month', pt.transaction_date)::DATE

    UNION ALL

    SELECT
        inv.organization_id,
        inv.store_id,
        DATE_TRUNC('month', inv.invoice_date::timestamp)::DATE AS period_month,
        COALESCE(SUM(inv.subtotal), 0) AS taxable_sales,
        COALESCE(SUM(inv.tax_amount), 0) AS output_vat
    FROM invoices inv
    WHERE inv.invoice_status NOT IN ('draft', 'cancelled')
    GROUP BY inv.organization_id, inv.store_id, DATE_TRUNC('month', inv.invoice_date::timestamp)::DATE
),
input_vat AS (
    SELECT
        po.organization_id,
        po.store_id,
        DATE_TRUNC('month', po.po_date::timestamp)::DATE AS period_month,
        COALESCE(SUM(po.subtotal), 0) AS taxable_purchases,
        COALESCE(SUM(po.tax_amount), 0) AS input_vat
    FROM purchase_orders po
    WHERE po.status NOT IN ('draft', 'cancelled')
    GROUP BY po.organization_id, po.store_id, DATE_TRUNC('month', po.po_date::timestamp)::DATE
),
all_periods AS (
    SELECT organization_id, store_id, period_month FROM output_vat
    UNION
    SELECT organization_id, store_id, period_month FROM input_vat
)
SELECT
    ap.organization_id,
    ap.store_id,
    s.name AS store_name,
    ap.period_month,
    COALESCE(SUM(ov.taxable_sales), 0) AS total_taxable_sales,
    COALESCE(SUM(ov.output_vat), 0) AS total_output_vat,
    COALESCE(SUM(iv.taxable_purchases), 0) AS total_taxable_purchases,
    COALESCE(SUM(iv.input_vat), 0) AS total_input_vat,
    (COALESCE(SUM(ov.output_vat), 0) - COALESCE(SUM(iv.input_vat), 0)) AS net_vat_payable
FROM all_periods ap
JOIN stores s ON s.id = ap.store_id
LEFT JOIN output_vat ov ON ov.organization_id = ap.organization_id AND ov.store_id = ap.store_id AND ov.period_month = ap.period_month
LEFT JOIN input_vat iv ON iv.organization_id = ap.organization_id AND iv.store_id = ap.store_id AND iv.period_month = ap.period_month
GROUP BY ap.organization_id, ap.store_id, s.name, ap.period_month
ORDER BY ap.period_month DESC, s.name;
