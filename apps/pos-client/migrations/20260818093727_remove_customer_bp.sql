-- +goose Up
-- Modify "customers" table
ALTER TABLE "public"."customers" DROP COLUMN "business_partner_id";
-- Create index "idx_inventory_stock_unique_product_variant_store" to table: "inventory_stock"
CREATE UNIQUE INDEX "idx_inventory_stock_unique_product_variant_store" ON "public"."inventory_stock" ("product_id", (COALESCE(product_variant_id, '-1'::integer)), "store_id");
-- Modify "fn_process_goods_receipt" function
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION "public"."fn_process_goods_receipt" ("p_grn_id" integer) RETURNS TABLE ("success" boolean, "message" text) LANGUAGE plpgsql AS $$
DECLARE
    v_grn RECORD;
    v_item RECORD;
    v_all_po_received BOOLEAN := true;
    v_any_po_received BOOLEAN := false;
    v_po_line RECORD;
BEGIN
    SELECT * INTO v_grn FROM goods_receipt_notes WHERE id = p_grn_id FOR UPDATE;
    IF v_grn IS NULL THEN
        RETURN QUERY SELECT false, 'Goods receipt note not found.';
        RETURN;
    END IF;

    IF v_grn.status = 'completed' THEN
        RETURN QUERY SELECT false, 'Goods receipt note is already completed.';
        RETURN;
    END IF;

    FOR v_item IN SELECT * FROM goods_receipt_note_items WHERE grn_id = p_grn_id FOR UPDATE LOOP
        IF v_item.quantity_received <= 0 THEN
            CONTINUE;
        END IF;

        -- Update PO Line if associated
        IF v_item.purchase_order_line_id IS NOT NULL THEN
            UPDATE purchase_order_lines
            SET received_quantity = received_quantity + v_item.quantity_received
            WHERE id = v_item.purchase_order_line_id;
        END IF;

        -- Increment inventory stock at store
        UPDATE inventory_stock
        SET quantity_on_hand = quantity_on_hand + v_item.quantity_received,
            quantity_available = quantity_available + v_item.quantity_received,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_grn.store_id;

        IF NOT FOUND THEN
            INSERT INTO inventory_stock (
                product_id, product_variant_id, store_id, storage_location_id,
                quantity_on_hand, quantity_available
            ) VALUES (
                v_item.product_id, v_item.product_variant_id, v_grn.store_id, v_item.storage_location_id,
                v_item.quantity_received, v_item.quantity_received
            );
        END IF;

        -- Insert stock movement (purchase_receipt)
        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            to_store_id, to_location_id, quantity, uom_id, batch_number,
            posted_by, status, cost_per_unit, total_value, metadata
        ) VALUES (
            'purchase_receipt', 'goods_receipt_note', p_grn_id, v_item.product_id, v_item.product_variant_id,
            v_grn.store_id, v_item.storage_location_id, v_item.quantity_received, v_item.uom_id, COALESCE(v_item.batch_number, ''),
            v_grn.received_by, 'completed', COALESCE(v_item.unit_cost, 0), (COALESCE(v_item.unit_cost, 0) * v_item.quantity_received),
            jsonb_build_object('grn_number', v_grn.grn_number, 'delivery_note_number', COALESCE(v_grn.delivery_note_number, ''))
        );
    END LOOP;

    -- Update GRN status
    UPDATE goods_receipt_notes
    SET status = 'completed',
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_grn_id;

    -- Update Purchase Order status if PO ID present
    IF v_grn.purchase_order_id IS NOT NULL THEN
        FOR v_po_line IN SELECT quantity, received_quantity FROM purchase_order_lines WHERE purchase_order_id = v_grn.purchase_order_id LOOP
            IF v_po_line.received_quantity < v_po_line.quantity THEN
                v_all_po_received := false;
            END IF;
            IF v_po_line.received_quantity > 0 THEN
                v_any_po_received := true;
            END IF;
        END LOOP;

        IF v_all_po_received THEN
            UPDATE purchase_orders SET status = 'received', updated_at = CURRENT_TIMESTAMP WHERE id = v_grn.purchase_order_id;
        ELSIF v_any_po_received THEN
            UPDATE purchase_orders SET status = 'partially_received', updated_at = CURRENT_TIMESTAMP WHERE id = v_grn.purchase_order_id;
        END IF;
    END IF;

    RETURN QUERY SELECT true, 'Goods receipt processed successfully.';
END;
$$;
-- +goose StatementEnd
-- Modify "fn_process_stock_transfer" function
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION "public"."fn_process_stock_transfer" ("p_from_store_id" integer, "p_to_store_id" integer, "p_product_id" integer, "p_product_variant_id" integer, "p_quantity" numeric, "p_from_location_id" integer DEFAULT NULL::integer, "p_to_location_id" integer DEFAULT NULL::integer, "p_batch_number" character varying DEFAULT NULL::character varying, "p_performed_by" integer DEFAULT NULL::integer, "p_notes" text DEFAULT NULL::text) RETURNS TABLE ("success" boolean, "message" text, "movement_id" integer) LANGUAGE plpgsql AS $$
DECLARE
    v_available  DECIMAL(15,3);
    v_movement_id INTEGER;
    v_ref_num    VARCHAR(50);
BEGIN
    -- Validate same store check
    IF p_from_store_id = p_to_store_id THEN
        RETURN QUERY SELECT false, 'Source and destination stores must differ.', NULL::INTEGER;
        RETURN;
    END IF;

    -- Lock and read available stock at source
    SELECT quantity_available INTO v_available
    FROM inventory_stock
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_from_store_id
    FOR UPDATE;

    IF v_available IS NULL OR v_available < p_quantity THEN
        RETURN QUERY SELECT false,
            format('Insufficient stock. Available: %s, Requested: %s', COALESCE(v_available, 0), p_quantity),
            NULL::INTEGER;
        RETURN;
    END IF;

    v_ref_num := 'TRF-' || to_char(CURRENT_TIMESTAMP, 'YYYYMMDDHH24MISS') || '-' || p_from_store_id || '-' || p_to_store_id;

    -- Deduct from source
    UPDATE inventory_stock
    SET quantity_on_hand   = quantity_on_hand   - p_quantity,
        quantity_available = quantity_available - p_quantity,
        quantity_in_transit = quantity_in_transit + p_quantity,
        updated_at = CURRENT_TIMESTAMP
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_from_store_id;

    -- Add to destination
    UPDATE inventory_stock
    SET quantity_on_hand   = quantity_on_hand   + p_quantity,
        quantity_available = quantity_available + p_quantity,
        updated_at         = CURRENT_TIMESTAMP
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_to_store_id;

    IF NOT FOUND THEN
        INSERT INTO inventory_stock (product_id, product_variant_id, store_id, storage_location_id,
            quantity_on_hand, quantity_available, quantity_in_transit)
        VALUES (p_product_id, p_product_variant_id, p_to_store_id, p_to_location_id,
                p_quantity, p_quantity, 0);
    END IF;

    -- Clear in-transit at source
    UPDATE inventory_stock
    SET quantity_in_transit = GREATEST(0, quantity_in_transit - p_quantity),
        updated_at = CURRENT_TIMESTAMP
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_from_store_id;

    -- Record movement
    INSERT INTO stock_movements (movement_type, reference_type, product_id, product_variant_id,
        from_store_id, to_store_id, from_location_id, to_location_id,
        quantity, batch_number, posted_by, status,
        metadata)
    VALUES ('transfer', 'manual', p_product_id, p_product_variant_id,
            p_from_store_id, p_to_store_id, p_from_location_id, p_to_location_id,
            p_quantity, p_batch_number, p_performed_by, 'completed',
            jsonb_build_object('reference_number', v_ref_num, 'notes', p_notes))
    RETURNING id INTO v_movement_id;

    RETURN QUERY SELECT true, 'Transfer completed successfully. Ref: ' || v_ref_num, v_movement_id;
END;
$$;
-- +goose StatementEnd
-- Modify "fn_receive_transfer_request" function
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION "public"."fn_receive_transfer_request" ("p_transfer_request_id" integer, "p_received_by" integer) RETURNS TABLE ("success" boolean, "message" text) LANGUAGE plpgsql AS $$
DECLARE
    v_req RECORD;
    v_item RECORD;
    v_qty DECIMAL(15,3);
    v_uom_code VARCHAR;
    v_base_qty DECIMAL(15,3);
BEGIN
    SELECT * INTO v_req FROM transfer_requests WHERE id = p_transfer_request_id FOR UPDATE;
    IF v_req IS NULL THEN
        RETURN QUERY SELECT false, 'Transfer request not found.';
        RETURN;
    END IF;

    IF v_req.status NOT IN ('shipped', 'partially_received') THEN
        RETURN QUERY SELECT false, 'Transfer request must be shipped or partially_received to receive.';
        RETURN;
    END IF;

    FOR v_item IN SELECT * FROM transfer_request_items WHERE transfer_request_id = p_transfer_request_id FOR UPDATE LOOP
        v_qty := CASE WHEN v_item.shipped_quantity > 0 THEN v_item.shipped_quantity ELSE v_item.requested_quantity END;
        IF v_qty <= 0 THEN
            CONTINUE;
        END IF;

        -- Retrieve UOM code for conversion
        SELECT code INTO v_uom_code FROM units_of_measure WHERE id = v_item.uom_id;
        
        -- Convert requested qty to base unit qty
        v_base_qty := fn_convert_uom_quantity(v_item.product_id, v_uom_code, v_qty);
        IF v_base_qty IS NULL THEN
            v_base_qty := v_qty;
        END IF;

        -- Decrement in-transit and increment on_hand & available at destination store using v_base_qty
        UPDATE inventory_stock
        SET quantity_in_transit = GREATEST(0, quantity_in_transit - v_base_qty),
            quantity_on_hand = quantity_on_hand + v_base_qty,
            quantity_available = quantity_available + v_base_qty,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.to_store_id;

        IF NOT FOUND THEN
            INSERT INTO inventory_stock (product_id, product_variant_id, store_id, storage_location_id,
                quantity_on_hand, quantity_available, quantity_in_transit)
            VALUES (v_item.product_id, v_item.product_variant_id, v_req.to_store_id, v_item.to_location_id,
                    v_base_qty, v_base_qty, 0);
        END IF;

        -- Update item received_quantity
        UPDATE transfer_request_items
        SET received_quantity = v_qty
        WHERE id = v_item.id;

        -- Record stock movement (transfer_in / completed) using v_base_qty
        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            from_store_id, to_store_id, from_location_id, to_location_id,
            quantity, uom_id, batch_number, posted_by, status, metadata
        ) VALUES (
            'transfer_in', 'transfer_request', p_transfer_request_id, v_item.product_id, v_item.product_variant_id,
            v_req.from_store_id, v_req.to_store_id, v_item.from_location_id, v_item.to_location_id,
            v_base_qty, v_item.uom_id, v_item.batch_number, p_received_by, 'completed',
            jsonb_build_object('transfer_number', v_req.transfer_number)
        );
    END LOOP;

    -- Update header
    UPDATE transfer_requests
    SET status = 'received',
        received_by = p_received_by,
        received_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_transfer_request_id;

    RETURN QUERY SELECT true, 'Transfer request received successfully.';
END;
$$;
-- +goose StatementEnd
-- Modify "fn_ship_transfer_request" function
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION "public"."fn_ship_transfer_request" ("p_transfer_request_id" integer, "p_shipped_by" integer) RETURNS TABLE ("success" boolean, "message" text) LANGUAGE plpgsql AS $$
DECLARE
    v_req RECORD;
    v_item RECORD;
    v_available DECIMAL(15,3);
    v_qty DECIMAL(15,3);
    v_uom_code VARCHAR;
    v_base_qty DECIMAL(15,3);
BEGIN
    SELECT * INTO v_req FROM transfer_requests WHERE id = p_transfer_request_id FOR UPDATE;
    IF v_req IS NULL THEN
        RETURN QUERY SELECT false, 'Transfer request not found.';
        RETURN;
    END IF;

    IF v_req.status NOT IN ('approved', 'pending_approval', 'draft') THEN
        RETURN QUERY SELECT false, 'Transfer request must be approved to ship.';
        RETURN;
    END IF;

    FOR v_item IN SELECT * FROM transfer_request_items WHERE transfer_request_id = p_transfer_request_id FOR UPDATE LOOP
        v_qty := CASE WHEN v_item.approved_quantity > 0 THEN v_item.approved_quantity ELSE v_item.requested_quantity END;
        IF v_qty <= 0 THEN
            CONTINUE;
        END IF;

        -- Retrieve UOM code for conversion
        SELECT code INTO v_uom_code FROM units_of_measure WHERE id = v_item.uom_id;
        
        -- Convert requested qty to base unit qty
        v_base_qty := fn_convert_uom_quantity(v_item.product_id, v_uom_code, v_qty);
        IF v_base_qty IS NULL THEN
            v_base_qty := v_qty;
        END IF;

        -- Check available stock at source using v_base_qty
        SELECT quantity_available INTO v_available
        FROM inventory_stock
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.from_store_id
        FOR UPDATE;

        IF v_available IS NULL OR v_available < v_base_qty THEN
            RETURN QUERY SELECT false, format('Insufficient stock for product ID %s at source store.', v_item.product_id);
            RETURN;
        END IF;

        -- Deduct from source store using v_base_qty
        UPDATE inventory_stock
        SET quantity_on_hand = quantity_on_hand - v_base_qty,
            quantity_available = quantity_available - v_base_qty,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.from_store_id;

        -- Increment quantity_in_transit at destination store using v_base_qty
        UPDATE inventory_stock
        SET quantity_in_transit = quantity_in_transit + v_base_qty,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.to_store_id;

        IF NOT FOUND THEN
            INSERT INTO inventory_stock (product_id, product_variant_id, store_id, storage_location_id,
                quantity_on_hand, quantity_available, quantity_in_transit)
            VALUES (v_item.product_id, v_item.product_variant_id, v_req.to_store_id, v_item.to_location_id,
                    0, 0, v_base_qty);
        END IF;

        -- Update item shipped_quantity
        UPDATE transfer_request_items
        SET shipped_quantity = v_qty,
            approved_quantity = v_qty
        WHERE id = v_item.id;

        -- Record stock movement (transfer_out / shipped) using v_base_qty
        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            from_store_id, to_store_id, from_location_id, to_location_id,
            quantity, uom_id, batch_number, posted_by, status, metadata
        ) VALUES (
            'transfer_out', 'transfer_request', p_transfer_request_id, v_item.product_id, v_item.product_variant_id,
            v_req.from_store_id, v_req.to_store_id, v_item.from_location_id, v_item.to_location_id,
            v_base_qty, v_item.uom_id, v_item.batch_number, p_shipped_by, 'shipped',
            jsonb_build_object('transfer_number', v_req.transfer_number)
        );
    END LOOP;

    -- Update header
    UPDATE transfer_requests
    SET status = 'shipped',
        shipped_by = p_shipped_by,
        shipped_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_transfer_request_id;

    RETURN QUERY SELECT true, 'Transfer request shipped successfully.';
END;
$$;
-- +goose StatementEnd
-- Modify "vw_accounts_payable" view
CREATE OR REPLACE VIEW "public"."vw_accounts_payable" (
  "po_id",
  "po_number",
  "organization_id",
  "organization_name",
  "supplier_id",
  "supplier_name",
  "contact_person",
  "email",
  "supplier_payment_terms",
  "store_name",
  "po_date",
  "expected_delivery_date",
  "status",
  "po_total",
  "discount_amount",
  "tax_amount",
  "received_amount",
  "amount_paid_str",
  "created_at"
) AS SELECT po.id AS po_id,
    po.po_number,
    po.organization_id,
    org.name AS organization_name,
    sup.id AS supplier_id,
    sup.name AS supplier_name,
    sup.contact_person,
    sup.email,
    sup.payment_terms AS supplier_payment_terms,
    s.name AS store_name,
    po.po_date,
    po.expected_delivery_date,
    po.status,
    po.total_amount AS po_total,
    po.discount_amount,
    po.tax_amount,
    ( SELECT COALESCE(sum(pol.received_quantity * pol.unit_price), 0::numeric) AS "coalesce"
           FROM public.purchase_order_lines pol
          WHERE pol.purchase_order_id = po.id) AS received_amount,
    po.metadata ->> 'amount_paid'::text AS amount_paid_str,
    po.created_at
   FROM public.purchase_orders po
     JOIN public.organizations org ON org.id = po.organization_id
     JOIN public.suppliers sup ON sup.id = po.supplier_id
     JOIN public.stores s ON s.id = po.store_id
  WHERE po.status::text = ANY (ARRAY['partially_received'::character varying, 'received'::character varying, 'approved'::character varying]::text[])
  ORDER BY po.po_date;
-- Modify "vw_active_restaurant_orders" view
CREATE OR REPLACE VIEW "public"."vw_active_restaurant_orders" (
  "order_id",
  "order_number",
  "store_id",
  "order_source",
  "order_status",
  "subtotal",
  "tax_amount",
  "total_amount",
  "notes",
  "ordered_at",
  "confirmed_at",
  "table_id",
  "table_number",
  "table_name",
  "table_section",
  "cashier_id",
  "waiter_name",
  "customer_id",
  "customer_name",
  "minutes_since_ordered"
) AS SELECT ro.id AS order_id,
    ro.order_number,
    ro.store_id,
    ro.order_source,
    ro.status AS order_status,
    ro.subtotal,
    ro.tax_amount,
    ro.total_amount,
    ro.notes,
    ro.ordered_at,
    ro.confirmed_at,
    rt.id AS table_id,
    rt.table_number,
    rt.table_name,
    rt.section AS table_section,
    c.id AS cashier_id,
    (u.first_name::text || ' '::text) || u.last_name::text AS waiter_name,
    ro.customer_id,
    cust.name AS customer_name,
    EXTRACT(epoch FROM CURRENT_TIMESTAMP - ro.ordered_at::timestamp with time zone) / 60.0 AS minutes_since_ordered
   FROM public.restaurant_orders ro
     LEFT JOIN public.restaurant_tables rt ON ro.table_id = rt.id
     LEFT JOIN public.cashiers c ON ro.cashier_id = c.id
     LEFT JOIN public.users u ON c.user_id = u.id
     LEFT JOIN public.customers cust ON ro.customer_id = cust.id
  WHERE ro.status::text <> ALL (ARRAY['paid'::character varying, 'voided'::character varying]::text[]);
-- Modify "vw_pending_purchase_orders" view
CREATE OR REPLACE VIEW "public"."vw_pending_purchase_orders" (
  "po_id",
  "po_number",
  "po_date",
  "expected_delivery_date",
  "status",
  "days_overdue",
  "is_overdue",
  "store_id",
  "store_name",
  "supplier_id",
  "supplier_name",
  "contact_person",
  "supplier_email",
  "subtotal",
  "discount_amount",
  "tax_amount",
  "total_amount",
  "outstanding_quantity",
  "created_by_username",
  "approved_by_username",
  "created_at"
) AS SELECT po.id AS po_id,
    po.po_number,
    po.po_date,
    po.expected_delivery_date,
    po.status,
    CURRENT_DATE - po.expected_delivery_date AS days_overdue,
    po.expected_delivery_date < CURRENT_DATE AND (po.status::text <> ALL (ARRAY['received'::character varying, 'cancelled'::character varying, 'closed'::character varying]::text[])) AS is_overdue,
    s.id AS store_id,
    s.name AS store_name,
    sup.id AS supplier_id,
    sup.name AS supplier_name,
    sup.contact_person,
    sup.email AS supplier_email,
    po.subtotal,
    po.discount_amount,
    po.tax_amount,
    po.total_amount,
    ( SELECT COALESCE(sum(pol.quantity - pol.received_quantity), 0::numeric) AS "coalesce"
           FROM public.purchase_order_lines pol
          WHERE pol.purchase_order_id = po.id) AS outstanding_quantity,
    u_created.username AS created_by_username,
    u_approved.username AS approved_by_username,
    po.created_at
   FROM public.purchase_orders po
     JOIN public.stores s ON s.id = po.store_id
     JOIN public.suppliers sup ON sup.id = po.supplier_id
     LEFT JOIN public.users u_created ON u_created.id = po.created_by
     LEFT JOIN public.users u_approved ON u_approved.id = po.approved_by
  WHERE po.status::text <> ALL (ARRAY['received'::character varying, 'cancelled'::character varying, 'closed'::character varying]::text[])
  ORDER BY (po.expected_delivery_date < CURRENT_DATE AND (po.status::text <> ALL (ARRAY['received'::character varying, 'cancelled'::character varying, 'closed'::character varying]::text[]))) DESC, (CURRENT_DATE - po.expected_delivery_date) DESC NULLS LAST, po.expected_delivery_date;
