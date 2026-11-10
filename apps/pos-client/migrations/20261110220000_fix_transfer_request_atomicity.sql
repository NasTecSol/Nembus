-- +goose Up
-- Modify "fn_approve_transfer_request" function
CREATE OR REPLACE FUNCTION fn_approve_transfer_request(
    p_transfer_request_id INTEGER,
    p_approved_by INTEGER
)
RETURNS TABLE (
    success BOOLEAN,
    message TEXT
) AS $BODY$
DECLARE
    v_req RECORD;
    v_item RECORD;
    v_need RECORD;
    v_available DECIMAL(15,3);
    v_qty DECIMAL(15,3);
    v_uom_code VARCHAR;
    v_base_qty DECIMAL(15,3);
    v_is_reserved BOOLEAN;
    v_expires_at TIMESTAMP;
BEGIN
    SELECT * INTO v_req FROM transfer_requests WHERE id = p_transfer_request_id FOR UPDATE;
    IF v_req IS NULL THEN
        RETURN QUERY SELECT false, 'Transfer request not found.';
        RETURN;
    END IF;

    IF v_req.status NOT IN ('draft', 'pending_approval') THEN
        RETURN QUERY SELECT false, 'Transfer request can only be approved from draft or pending_approval state.';
        RETURN;
    END IF;

    v_is_reserved := COALESCE(v_req.is_stock_reserved, false);

    -- Expiry timestamp: expected_delivery_date end of day if present, else midnight tonight (CURRENT_DATE + 1 day)
    IF v_req.expected_delivery_date IS NOT NULL THEN
        v_expires_at := (v_req.expected_delivery_date::TIMESTAMP + INTERVAL '1 day');
    ELSE
        v_expires_at := (CURRENT_DATE + INTERVAL '1 day');
    END IF;

    IF v_is_reserved THEN
        -- Upfront validation pass: lock all items and verify source stock sufficiency before mutating anything
        PERFORM 1 FROM transfer_request_items WHERE transfer_request_id = p_transfer_request_id FOR UPDATE;

        FOR v_need IN
            SELECT i.product_id, i.product_variant_id,
                   SUM(COALESCE(fn_convert_uom_quantity(i.product_id, u.code, q.qty), q.qty)) AS base_qty
            FROM transfer_request_items i
            LEFT JOIN units_of_measure u ON u.id = i.uom_id
            CROSS JOIN LATERAL (SELECT CASE WHEN i.approved_quantity > 0 THEN i.approved_quantity ELSE i.requested_quantity END AS qty) q
            WHERE i.transfer_request_id = p_transfer_request_id AND q.qty > 0
            GROUP BY i.product_id, i.product_variant_id
            ORDER BY i.product_id, i.product_variant_id NULLS FIRST
        LOOP
            SELECT quantity_available INTO v_available
            FROM inventory_stock
            WHERE product_id = v_need.product_id
              AND (product_variant_id = v_need.product_variant_id OR (product_variant_id IS NULL AND v_need.product_variant_id IS NULL))
              AND store_id = v_req.from_store_id
            FOR UPDATE;

            IF v_available IS NULL OR v_available < v_need.base_qty THEN
                RETURN QUERY SELECT false, format('Insufficient stock to reserve product ID %s at source store.', v_need.product_id);
                RETURN;
            END IF;
        END LOOP;

        -- Mutation pass: all stock verified, apply allocations and reservations
        FOR v_item IN SELECT * FROM transfer_request_items WHERE transfer_request_id = p_transfer_request_id FOR UPDATE LOOP
            v_qty := CASE WHEN v_item.approved_quantity > 0 THEN v_item.approved_quantity ELSE v_item.requested_quantity END;
            IF v_qty <= 0 THEN
                CONTINUE;
            END IF;

            SELECT code INTO v_uom_code FROM units_of_measure WHERE id = v_item.uom_id;
            v_base_qty := fn_convert_uom_quantity(v_item.product_id, v_uom_code, v_qty);
            IF v_base_qty IS NULL THEN
                v_base_qty := v_qty;
            END IF;

            SELECT quantity_available INTO v_available
            FROM inventory_stock
            WHERE product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND store_id = v_req.from_store_id
            FOR UPDATE;

            IF v_available IS NULL OR v_available < v_base_qty THEN
                RAISE EXCEPTION 'Insufficient stock to reserve product ID % at source store.', v_item.product_id;
            END IF;

            -- Allocate stock at source store
            UPDATE inventory_stock
            SET quantity_allocated = quantity_allocated + v_base_qty,
                quantity_available = quantity_available - v_base_qty,
                updated_at = CURRENT_TIMESTAMP
            WHERE product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND store_id = v_req.from_store_id;

            -- Log reservation row with expires_at
            INSERT INTO stock_reservations (
                reservation_number, product_id, product_variant_id, store_id,
                reference_type, reference_id, quantity_reserved, expires_at, status, reserved_by
            ) VALUES (
                'TR-RES-' || p_transfer_request_id || '-' || v_item.id || '-' || FLOOR(EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)),
                v_item.product_id, v_item.product_variant_id, v_req.from_store_id,
                'transfer', p_transfer_request_id::text, v_base_qty, v_expires_at, 'active', p_approved_by
            );
        END LOOP;
    END IF;

    UPDATE transfer_requests
    SET status = 'approved',
        approved_by = p_approved_by,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_transfer_request_id;

    RETURN QUERY SELECT true, 'Transfer request approved successfully.';
END;
$BODY$ LANGUAGE plpgsql;

-- Modify "fn_allocate_transfer_request_reservation" function
CREATE OR REPLACE FUNCTION fn_allocate_transfer_request_reservation(
    p_transfer_request_id INTEGER
)
RETURNS TABLE (
    success BOOLEAN,
    message TEXT
) AS $BODY$
DECLARE
    v_req RECORD;
    v_item RECORD;
    v_need RECORD;
    v_available DECIMAL(15,3);
    v_qty DECIMAL(15,3);
    v_uom_code VARCHAR;
    v_base_qty DECIMAL(15,3);
    v_expires_at TIMESTAMP;
BEGIN
    SELECT * INTO v_req FROM transfer_requests WHERE id = p_transfer_request_id FOR UPDATE;
    IF v_req IS NULL THEN
        RETURN QUERY SELECT false, 'Transfer request not found.';
        RETURN;
    END IF;

    IF v_req.is_stock_reserved AND v_req.status = 'approved' THEN
        IF v_req.expected_delivery_date IS NOT NULL THEN
            v_expires_at := (v_req.expected_delivery_date::TIMESTAMP + INTERVAL '1 day');
        ELSE
            v_expires_at := (CURRENT_DATE + INTERVAL '1 day');
        END IF;

        -- Upfront validation pass: lock items and check stock sufficiency before mutating anything
        PERFORM 1 FROM transfer_request_items WHERE transfer_request_id = p_transfer_request_id FOR UPDATE;

        FOR v_need IN
            SELECT i.product_id, i.product_variant_id,
                   SUM(COALESCE(fn_convert_uom_quantity(i.product_id, u.code, q.qty), q.qty)) AS base_qty
            FROM transfer_request_items i
            LEFT JOIN units_of_measure u ON u.id = i.uom_id
            CROSS JOIN LATERAL (SELECT CASE WHEN i.approved_quantity > 0 THEN i.approved_quantity ELSE i.requested_quantity END AS qty) q
            WHERE i.transfer_request_id = p_transfer_request_id AND q.qty > 0
            GROUP BY i.product_id, i.product_variant_id
            ORDER BY i.product_id, i.product_variant_id NULLS FIRST
        LOOP
            SELECT quantity_available INTO v_available
            FROM inventory_stock
            WHERE product_id = v_need.product_id
              AND (product_variant_id = v_need.product_variant_id OR (product_variant_id IS NULL AND v_need.product_variant_id IS NULL))
              AND store_id = v_req.from_store_id
            FOR UPDATE;

            IF v_available IS NULL OR v_available < v_need.base_qty THEN
                RETURN QUERY SELECT false, format('Insufficient stock to reserve product ID %s at source store.', v_need.product_id);
                RETURN;
            END IF;
        END LOOP;

        -- Mutation pass: apply allocations and reservations
        FOR v_item IN SELECT * FROM transfer_request_items WHERE transfer_request_id = p_transfer_request_id FOR UPDATE LOOP
            v_qty := CASE WHEN v_item.approved_quantity > 0 THEN v_item.approved_quantity ELSE v_item.requested_quantity END;
            IF v_qty <= 0 THEN
                CONTINUE;
            END IF;

            SELECT code INTO v_uom_code FROM units_of_measure WHERE id = v_item.uom_id;
            v_base_qty := fn_convert_uom_quantity(v_item.product_id, v_uom_code, v_qty);
            IF v_base_qty IS NULL THEN
                v_base_qty := v_qty;
            END IF;

            SELECT quantity_available INTO v_available
            FROM inventory_stock
            WHERE product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND store_id = v_req.from_store_id
            FOR UPDATE;

            IF v_available IS NULL OR v_available < v_base_qty THEN
                RAISE EXCEPTION 'Insufficient stock to reserve product ID % at source store.', v_item.product_id;
            END IF;

            -- Allocate stock at source store
            UPDATE inventory_stock
            SET quantity_allocated = quantity_allocated + v_base_qty,
                quantity_available = quantity_available - v_base_qty,
                updated_at = CURRENT_TIMESTAMP
            WHERE product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND store_id = v_req.from_store_id;

            -- Log reservation row with expires_at
            INSERT INTO stock_reservations (
                reservation_number, product_id, product_variant_id, store_id,
                reference_type, reference_id, quantity_reserved, expires_at, status, reserved_by
            ) VALUES (
                'TR-RES-' || p_transfer_request_id || '-' || v_item.id || '-' || FLOOR(EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)),
                v_item.product_id, v_item.product_variant_id, v_req.from_store_id,
                'transfer', p_transfer_request_id::text, v_base_qty, v_expires_at, 'active', v_req.approved_by
            );
        END LOOP;
    END IF;

    RETURN QUERY SELECT true, 'Reservation allocated successfully.';
END;
$BODY$ LANGUAGE plpgsql;

-- Modify "fn_ship_transfer_request" function
CREATE OR REPLACE FUNCTION fn_ship_transfer_request(
    p_transfer_request_id INTEGER,
    p_shipped_by INTEGER
)
RETURNS TABLE (
    success BOOLEAN,
    message TEXT
) AS $BODY$
DECLARE
    v_req RECORD;
    v_item RECORD;
    v_need RECORD;
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

    -- If stock is not pre-reserved, run an upfront validation pass across all items first
    IF NOT COALESCE(v_req.is_stock_reserved, false) THEN
        PERFORM 1 FROM transfer_request_items WHERE transfer_request_id = p_transfer_request_id FOR UPDATE;

        FOR v_need IN
            SELECT i.product_id, i.product_variant_id,
                   SUM(COALESCE(fn_convert_uom_quantity(i.product_id, u.code, q.qty), q.qty)) AS base_qty
            FROM transfer_request_items i
            LEFT JOIN units_of_measure u ON u.id = i.uom_id
            CROSS JOIN LATERAL (SELECT CASE WHEN i.approved_quantity > 0 THEN i.approved_quantity ELSE i.requested_quantity END AS qty) q
            WHERE i.transfer_request_id = p_transfer_request_id AND q.qty > 0
            GROUP BY i.product_id, i.product_variant_id
            ORDER BY i.product_id, i.product_variant_id NULLS FIRST
        LOOP
            SELECT quantity_available INTO v_available
            FROM inventory_stock
            WHERE product_id = v_need.product_id
              AND (product_variant_id = v_need.product_variant_id OR (product_variant_id IS NULL AND v_need.product_variant_id IS NULL))
              AND store_id = v_req.from_store_id
            FOR UPDATE;

            IF v_available IS NULL OR v_available < v_need.base_qty THEN
                RETURN QUERY SELECT false, format('Insufficient stock for product ID %s at source store.', v_need.product_id);
                RETURN;
            END IF;
        END LOOP;
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

        IF v_req.is_stock_reserved THEN
            -- Reserved mode: clear allocated and deduct on_hand
            UPDATE inventory_stock
            SET quantity_allocated = GREATEST(0, quantity_allocated - v_base_qty),
                quantity_on_hand = quantity_on_hand - v_base_qty,
                updated_at = CURRENT_TIMESTAMP
            WHERE product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND store_id = v_req.from_store_id;

            UPDATE stock_reservations
            SET status = 'fulfilled',
                updated_at = CURRENT_TIMESTAMP
            WHERE reference_type = 'transfer'
              AND reference_id = p_transfer_request_id::text
              AND product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND status = 'active';
        ELSE
            -- Default unreserved mode: check available and deduct both on_hand & available
            SELECT quantity_available INTO v_available
            FROM inventory_stock
            WHERE product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND store_id = v_req.from_store_id
            FOR UPDATE;

            IF v_available IS NULL OR v_available < v_base_qty THEN
                RAISE EXCEPTION 'Insufficient stock for product ID % at source store.', v_item.product_id;
            END IF;

            UPDATE inventory_stock
            SET quantity_on_hand = quantity_on_hand - v_base_qty,
                quantity_available = quantity_available - v_base_qty,
                updated_at = CURRENT_TIMESTAMP
            WHERE product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND store_id = v_req.from_store_id;
        END IF;

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
$BODY$ LANGUAGE plpgsql;