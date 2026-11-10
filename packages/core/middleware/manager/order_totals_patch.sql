-- Apply to existing cloned tenants when their connection pool opens.
DO $patch$
BEGIN
    IF to_regclass('sales_orders_v2') IS NULL OR to_regclass('sales_order_lines_v2') IS NULL THEN
        RETURN;
    END IF;

    EXECUTE $definition$
CREATE OR REPLACE FUNCTION calculate_order_totals()
RETURNS TRIGGER AS $$
DECLARE
    v_subtotal DECIMAL(15,2);
    v_tax_amount DECIMAL(15,2);
BEGIN
    -- Calculate subtotal and tax from order lines
    SELECT 
        COALESCE(SUM(line_total - COALESCE(tax_amount, 0)), 0),
        COALESCE(SUM(tax_amount), 0)
    INTO v_subtotal, v_tax_amount
    FROM sales_order_lines_v2
    WHERE sales_order_id = COALESCE(NEW.sales_order_id, OLD.sales_order_id);
    
    -- Update the order
    UPDATE sales_orders_v2
    SET 
        subtotal = v_subtotal,
        tax_amount = v_tax_amount,
        total_amount = v_subtotal + v_tax_amount + COALESCE(shipping_amount, 0) + COALESCE(adjustment_amount, 0) - COALESCE(discount_amount, 0),
        balance_due = (v_subtotal + v_tax_amount + COALESCE(shipping_amount, 0) + COALESCE(adjustment_amount, 0) - COALESCE(discount_amount, 0)) - COALESCE(paid_amount, 0)
    WHERE id = COALESCE(NEW.sales_order_id, OLD.sales_order_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
    $definition$;

    -- Repair only already-failing jobs whose headers match the old NULL-tax formula.
    -- Existing signed documents and invoiced orders must remain unchanged.
    IF to_regclass('zatca_document_chain') IS NULL OR
       to_regclass('pos_transactions') IS NULL OR to_regclass('invoices') IS NULL OR
       to_regclass('zatca_order_jobs') IS NULL THEN
        RETURN;
    END IF;

    WITH line_totals AS (
        SELECT sales_order_id,
               COALESCE(SUM(line_total - tax_amount), 0) AS old_subtotal,
               COALESCE(SUM(line_total - COALESCE(tax_amount, 0)), 0) AS subtotal,
               COALESCE(SUM(tax_amount), 0) AS tax_amount
        FROM sales_order_lines_v2
        WHERE sales_order_id IN (
            SELECT order_id FROM zatca_order_jobs
            WHERE last_error = 'unexplained difference between order and line totals'
        )
        GROUP BY sales_order_id
        HAVING BOOL_OR(tax_amount IS NULL)
    )
    UPDATE sales_orders_v2 AS o
    SET subtotal = t.subtotal,
        tax_amount = t.tax_amount,
        total_amount = t.subtotal + t.tax_amount + COALESCE(o.shipping_amount, 0)
                     + COALESCE(o.adjustment_amount, 0) - COALESCE(o.discount_amount, 0),
        balance_due = t.subtotal + t.tax_amount + COALESCE(o.shipping_amount, 0)
                    + COALESCE(o.adjustment_amount, 0) - COALESCE(o.discount_amount, 0)
                    - COALESCE(o.paid_amount, 0),
        updated_at = NOW()
    FROM line_totals t
    WHERE o.id = t.sales_order_id
      AND t.subtotal > t.old_subtotal
      AND o.subtotal = t.old_subtotal
      AND COALESCE(o.tax_amount, 0) = t.tax_amount
      AND o.total_amount = t.old_subtotal + t.tax_amount + COALESCE(o.shipping_amount, 0)
                         + COALESCE(o.adjustment_amount, 0) - COALESCE(o.discount_amount, 0)
      AND NOT EXISTS (SELECT 1 FROM invoices i WHERE i.sales_order_id = o.id)
      AND NOT EXISTS (
          SELECT 1 FROM zatca_document_chain c
          WHERE (c.entity_type = 'sales_order_v2' AND c.entity_id = o.id::text)
             OR (c.entity_type = 'pos_transaction' AND EXISTS (
                 SELECT 1 FROM pos_transactions p
                 WHERE p.sales_order_id = o.id AND p.id::text = c.entity_id
             ))
      );
END;
$patch$;
