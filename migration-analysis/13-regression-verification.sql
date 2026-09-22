-- Sweep 2 regression verification for nembus_migration_audit_20260922.
-- Every statement is SELECT-only. Run in a read-only session.

SELECT current_database() AS database_name,
       current_setting('transaction_read_only') AS transaction_read_only,
       current_setting('default_transaction_read_only') AS default_transaction_read_only;

SELECT table_schema,
       table_name
FROM information_schema.tables
WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
ORDER BY table_schema, table_name;

SELECT n.nspname AS schema_name,
       c.relname AS relation_name,
       CASE c.relkind
         WHEN 'r' THEN 'table'
         WHEN 'p' THEN 'partitioned_table'
         WHEN 'v' THEN 'view'
         WHEN 'm' THEN 'materialized_view'
         WHEN 'S' THEN 'sequence'
         ELSE c.relkind::text
       END AS relation_kind
FROM pg_catalog.pg_class c
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname IN ('atlas_schema_revisions', 'public', 'staging')
ORDER BY n.nspname, c.relname;

-- Invoice header distribution and source-key coverage.
SELECT currency_code,
       invoice_status,
       invoice_type,
       COUNT(*) AS invoice_count,
       COUNT(*) FILTER (WHERE discount_amount IS NULL) AS null_discount_count,
       COUNT(*) FILTER (WHERE discount_amount = 0) AS zero_discount_count,
       SUM(COALESCE(discount_amount, 0)) AS discount_sum
FROM public.invoices
GROUP BY currency_code, invoice_status, invoice_type
ORDER BY currency_code, invoice_status, invoice_type;

SELECT COUNT(*) AS invoice_count,
       COUNT(*) FILTER (WHERE metadata ? 'sap_doc_entry') AS with_sap_doc_entry,
       COUNT(DISTINCT metadata->>'sap_doc_entry') AS distinct_sap_doc_entries,
       COUNT(*) FILTER (WHERE metadata->>'sap_doc_entry' IS NULL) AS without_sap_doc_entry,
       MIN((metadata->>'sap_doc_entry')::bigint) AS minimum_sap_doc_entry,
       MAX((metadata->>'sap_doc_entry')::bigint) AS maximum_sap_doc_entry
FROM public.invoices
WHERE metadata->>'sap_doc_entry' IS NOT NULL;

-- Invoice-line discount, UOM, tax-reference, and source-key population.
SELECT COUNT(*) AS line_count,
       COUNT(*) FILTER (WHERE discount_amount IS NULL) AS null_discount_count,
       COUNT(*) FILTER (WHERE discount_amount = 0) AS zero_discount_count,
       COUNT(*) FILTER (WHERE discount_amount <> 0) AS nonzero_discount_count,
       COUNT(*) FILTER (WHERE uom_id IS NULL) AS null_uom_count,
       COUNT(*) FILTER (WHERE tax_category_id IS NULL) AS null_tax_category_count,
       COUNT(*) FILTER (WHERE tax_rate IS NULL) AS null_tax_rate_count,
       COUNT(*) FILTER (WHERE metadata ? 'sap_doc_entry' AND metadata ? 'sap_line_num') AS with_source_pair
FROM public.invoice_lines;

SELECT metadata->>'sap_doc_entry' AS sap_doc_entry,
       metadata->>'sap_line_num' AS sap_line_num,
       COUNT(*) AS duplicate_count
FROM public.invoice_lines
WHERE metadata ? 'sap_doc_entry' AND metadata ? 'sap_line_num'
GROUP BY metadata->>'sap_doc_entry', metadata->>'sap_line_num'
HAVING COUNT(*) > 1
ORDER BY duplicate_count DESC, sap_doc_entry, sap_line_num
LIMIT 100;

-- Exact target lookup for SAP invoice DocEntry 80306 / line 5.
SELECT i.id,
       i.invoice_number,
       i.invoice_date,
       i.due_date,
       i.customer_id,
       i.customer_name,
       i.invoice_type,
       i.invoice_status,
       i.subtotal,
       i.discount_amount,
       i.tax_amount,
       i.total_amount,
       i.paid_amount,
       i.balance_due,
       i.currency_code,
       i.metadata
FROM public.invoices i
WHERE i.metadata->>'sap_doc_entry' = '80306'
   OR i.invoice_number = 'INV-SAP-2022025735';

SELECT l.id,
       l.invoice_id,
       l.line_number,
       l.product_id,
       l.product_sku,
       l.quantity,
       l.unit_price,
       l.discount_amount,
       l.tax_amount,
       l.line_total,
       l.tax_category_id,
       l.tax_rate,
       l.uom_id,
       l.item_type,
       l.metadata
FROM public.invoice_lines l
JOIN public.invoices i ON i.id = l.invoice_id
WHERE i.metadata->>'sap_doc_entry' = '80306'
ORDER BY l.line_number;

-- Target-side invoice total reconstruction for every loaded invoice.
WITH line_rollup AS (
  SELECT invoice_id,
         SUM(COALESCE(line_total, 0)) AS line_total_sum,
         SUM(COALESCE(tax_amount, 0)) AS line_tax_sum,
         COUNT(*) AS line_count
  FROM public.invoice_lines
  GROUP BY invoice_id
), reconstructed AS (
  SELECT i.id,
         i.invoice_number,
         i.subtotal,
         i.tax_amount,
         i.total_amount,
         r.line_count,
         (r.line_total_sum - r.line_tax_sum) AS reconstructed_subtotal,
         ((r.line_total_sum - r.line_tax_sum)
           + COALESCE(i.tax_amount, 0)
           + COALESCE(i.shipping_amount, 0)
           + COALESCE(i.adjustment_amount, 0)
           - COALESCE(i.discount_amount, 0)) AS reconstructed_total
  FROM public.invoices i
  JOIN line_rollup r ON r.invoice_id = i.id
)
SELECT COUNT(*) AS invoice_count,
       COUNT(*) FILTER (WHERE ROUND(subtotal - reconstructed_subtotal, 2) <> 0) AS subtotal_mismatch_count,
       COUNT(*) FILTER (WHERE ROUND(total_amount - reconstructed_total, 2) <> 0) AS total_mismatch_count,
       MAX(ABS(subtotal - reconstructed_subtotal)) AS max_subtotal_difference,
       MAX(ABS(total_amount - reconstructed_total)) AS max_total_difference,
       MIN(line_count) AS minimum_line_count
FROM reconstructed;

-- Exact header-discount aggregate proof.
SELECT COUNT(*) AS invoice_count,
       COUNT(*) FILTER (WHERE discount_amount > 0) AS positive_count,
       COUNT(*) FILTER (WHERE discount_amount = 0) AS zero_count,
       COUNT(*) FILTER (WHERE discount_amount < 0) AS negative_count,
       COUNT(*) FILTER (WHERE discount_amount IS NULL) AS null_count,
       MIN(discount_amount) AS minimum_discount,
       MAX(discount_amount) AS maximum_discount,
       SUM(discount_amount) AS discount_sum,
       SUM(discount_amount * paid_amount) AS weighted_discount_sum,
       SUM(discount_amount * discount_amount) AS squared_discount_sum,
       SUM(total_amount) AS total_sum,
       SUM(tax_amount) AS tax_sum,
       SUM(subtotal) AS subtotal_sum
FROM public.invoices;

-- Target-side orphan and duplicate checks.
SELECT COUNT(*) AS orphan_invoice_line_count
FROM public.invoice_lines l
LEFT JOIN public.invoices i ON i.id = l.invoice_id
WHERE i.id IS NULL;

SELECT invoice_status, COUNT(*) AS invoice_count
FROM public.invoices
GROUP BY invoice_status
ORDER BY invoice_status;

SELECT currency_code, COUNT(*) AS invoice_count
FROM public.invoices
GROUP BY currency_code
ORDER BY invoice_count DESC, currency_code;

SELECT COUNT(*) AS canceled_status_count
FROM public.invoices
WHERE invoice_status = 'cancelled';

-- Product, reference, and source-key coverage.
SELECT COUNT(*) AS product_count,
       COUNT(DISTINCT metadata->>'sap_item_code') AS distinct_sap_item_codes,
       COUNT(*) FILTER (WHERE metadata->>'sap_item_code' IS NULL) AS missing_sap_item_code,
       COUNT(*) FILTER (WHERE category_id IS NULL) AS missing_category,
       COUNT(*) FILTER (WHERE brand_id IS NULL) AS missing_brand,
       COUNT(*) FILTER (WHERE base_uom_id IS NULL) AS missing_base_uom
FROM public.products;

SELECT COUNT(*) AS duplicate_sap_item_code_groups
FROM (
  SELECT metadata->>'sap_item_code' AS sap_item_code
  FROM public.products
  WHERE metadata->>'sap_item_code' IS NOT NULL
  GROUP BY metadata->>'sap_item_code'
  HAVING COUNT(*) > 1
) d;

SELECT COUNT(*) AS tax_category_count,
       COUNT(*) FILTER (WHERE tax_rate IS NULL) AS null_tax_rate_count
FROM public.tax_categories;

-- Inventory duplicate proof.
SELECT COUNT(*) AS inventory_rows,
       COUNT(DISTINCT COALESCE(metadata->>'sap_item_code', '') || '|' || COALESCE(metadata->>'sap_whs_code', '')) AS distinct_source_pairs
FROM public.inventory_stock;

SELECT duplicate_count, COUNT(*) AS source_pair_groups, SUM(duplicate_count) AS rows_in_groups
FROM (
  SELECT metadata->>'sap_item_code' AS sap_item_code,
         metadata->>'sap_whs_code' AS sap_whs_code,
         COUNT(*) AS duplicate_count
  FROM public.inventory_stock
  GROUP BY metadata->>'sap_item_code', metadata->>'sap_whs_code'
) d
GROUP BY duplicate_count
ORDER BY duplicate_count;

-- Stock movement duplicate proof.
SELECT COUNT(*) AS movement_rows,
       COUNT(DISTINCT metadata->>'sap_trans_num') AS distinct_source_trans_nums,
       COUNT(*) - COUNT(DISTINCT metadata->>'sap_trans_num') AS excess_duplicate_rows,
       MIN(movement_date) AS minimum_movement_date,
       MAX(movement_date) AS maximum_movement_date
FROM public.stock_movements;

SELECT duplicate_count, COUNT(*) AS source_key_groups, SUM(duplicate_count) AS rows_in_groups
FROM (
  SELECT metadata->>'sap_trans_num' AS sap_trans_num,
         COUNT(*) AS duplicate_count
  FROM public.stock_movements
  WHERE metadata->>'sap_trans_num' IS NOT NULL
  GROUP BY metadata->>'sap_trans_num'
) d
GROUP BY duplicate_count
ORDER BY duplicate_count;

-- PO/GRN relationship checks.
SELECT COUNT(*) AS grn_count,
       COUNT(*) FILTER (WHERE purchase_order_id IS NULL) AS null_purchase_order_count
FROM public.goods_receipt_notes;

SELECT COUNT(*) AS grn_item_count,
       COUNT(*) FILTER (WHERE purchase_order_line_id IS NULL) AS null_purchase_order_line_count
FROM public.goods_receipt_note_items;

SELECT COUNT(*) AS purchase_order_count,
       COUNT(*) FILTER (WHERE supplier_id IS NULL) AS null_supplier_count,
       COUNT(*) FILTER (WHERE store_id IS NULL) AS null_store_count
FROM public.purchase_orders;

-- Payment and optional domain checks.
SELECT COUNT(*) AS payment_count
FROM public.invoice_payments;

SELECT table_name, SUM(row_count) AS row_count
FROM (
  SELECT 'sap_migration_batches' AS table_name, COUNT(*) AS row_count
  FROM staging.sap_migration_batches
  UNION ALL
  SELECT 'failed_invoice_batches', COUNT(*)
  FROM staging.sap_migration_batches
  WHERE domain = 'invoices' AND status = 'failed'
) s
GROUP BY table_name
ORDER BY table_name;

SELECT domain,
       status,
       COUNT(*) AS batch_count,
       SUM(record_count) AS record_count_sum,
       COUNT(*) FILTER (WHERE COALESCE(error_message, '') = '') AS empty_error_message_count
FROM staging.sap_migration_batches
GROUP BY domain, status
ORDER BY domain, status;

-- Sales, return, address, and staging-domain inventory.
SELECT COUNT(*) AS sales_order_count FROM public.sales_orders_v2;
SELECT COUNT(*) AS sales_order_line_count FROM public.sales_order_lines_v2;
SELECT COUNT(*) AS return_count FROM public.sales_returns;
SELECT COUNT(*) AS return_line_count FROM public.sales_return_lines;
SELECT COUNT(*) AS partner_address_count FROM public.partner_addresses;
