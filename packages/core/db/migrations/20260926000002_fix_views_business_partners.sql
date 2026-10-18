-- =====================================================================
-- Migration: Fix Views for Business Partners & Barcode Join
-- 1. vw_stock_movement_ledger: joins product_barcodes instead of p.barcode
-- 2. vw_vendor_performance_ppv: replaces suppliers with business_partners
-- 3. vw_supplier_aging_report: replaces suppliers with business_partners
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. Stock Movement Running Ledger & Inventory Valuation View
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
    pb.barcode,
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
LEFT JOIN product_barcodes pb ON pb.product_id = p.id AND pb.is_primary = true
LEFT JOIN product_categories pc ON pc.id = p.category_id
LEFT JOIN stores s ON s.id = COALESCE(sm.to_store_id, sm.from_store_id);

-- ---------------------------------------------------------------------
-- 2. Vendor Fulfillment & Purchase Price Variance (PPV) View
-- ---------------------------------------------------------------------
CREATE OR REPLACE VIEW vw_vendor_performance_ppv AS
SELECT
    bp.id AS supplier_id,
    bp.code AS supplier_code,
    bp.name AS supplier_name,
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
JOIN business_partners bp ON bp.id = grn.partners_id
JOIN products p ON p.id = grni.product_id;

-- ---------------------------------------------------------------------
-- 3. Supplier Accounts Payable Aging Report View
-- ---------------------------------------------------------------------
CREATE OR REPLACE VIEW vw_supplier_aging_report AS
WITH po_balances AS (
    SELECT
        po.id AS po_id,
        po.partners_id,
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
    bp.id AS supplier_id,
    bp.code AS supplier_code,
    bp.name AS supplier_name,
    bp.credit_limit,
    pb.organization_id,
    COALESCE(SUM(CASE WHEN pb.due_date >= CURRENT_DATE THEN pb.balance_due ELSE 0 END), 0) AS current_amount,
    COALESCE(SUM(CASE WHEN pb.due_date < CURRENT_DATE AND CURRENT_DATE - pb.due_date <= 30 THEN pb.balance_due ELSE 0 END), 0) AS overdue_1_30,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - pb.due_date BETWEEN 31 AND 60 THEN pb.balance_due ELSE 0 END), 0) AS overdue_31_60,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - pb.due_date BETWEEN 61 AND 90 THEN pb.balance_due ELSE 0 END), 0) AS overdue_61_90,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - pb.due_date > 90 THEN pb.balance_due ELSE 0 END), 0) AS overdue_over_90,
    COALESCE(SUM(CASE WHEN pb.balance_due > 0 THEN pb.balance_due ELSE 0 END), 0) AS total_outstanding,
    COUNT(CASE WHEN pb.due_date < CURRENT_DATE AND pb.balance_due > 0 THEN 1 END)::INTEGER AS overdue_po_count,
    MAX(pb.due_date) AS latest_due_date
FROM business_partners bp
LEFT JOIN po_balances pb
    ON pb.partners_id = bp.id
    AND pb.balance_due > 0
WHERE bp.is_active = true AND bp.partner_role = 'supplier'
GROUP BY
    bp.id, bp.code, bp.name,
    bp.credit_limit, pb.organization_id
ORDER BY total_outstanding DESC;
