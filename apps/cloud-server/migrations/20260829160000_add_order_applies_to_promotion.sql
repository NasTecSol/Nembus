-- Modify constraint on promotions table to allow 'order'
ALTER TABLE "public"."promotions" DROP CONSTRAINT IF EXISTS "promotions_applies_to_check";
ALTER TABLE "public"."promotions" ADD CONSTRAINT "promotions_applies_to_check" CHECK ((applies_to)::text = ANY ((ARRAY['all'::character varying, 'order'::character varying, 'category'::character varying, 'product'::character varying, 'customer_type'::character varying, 'price_list'::character varying])::text[]));

-- Update fn_sync_promotion_to_product_prices to skip 'order' promotions
CREATE OR REPLACE FUNCTION "public"."fn_sync_promotion_to_product_prices" () RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    v_promo_pl_id INTEGER;
    v_target_product_id INTEGER;
    v_retail_pp RECORD;
    v_calculated_price NUMERIC(15,2);
    v_discount_percent_str VARCHAR;
    v_variant_id INTEGER;
BEGIN
    -- If applies_to is 'order', skip generating item-level promotional prices
    IF NEW.applies_to = 'order' THEN
        RETURN NEW;
    END IF;

    -- Locate existing PROMO price list
    SELECT id INTO v_promo_pl_id
    FROM price_lists
    WHERE (code = 'PROMO' OR price_list_type = 'promotional')
      AND is_active = true
    ORDER BY id ASC
    LIMIT 1;

    -- If no PROMO price list exists, dynamically create one
    IF v_promo_pl_id IS NULL AND (TG_OP = 'INSERT' OR TG_OP = 'UPDATE') THEN
        INSERT INTO price_lists (code, name, price_list_type, currency_code, is_active)
        VALUES (
            'PROMO', 
            'Promotional Price List', 
            'promotional', 
            'SAR', 
            true
        )
        ON CONFLICT (code) DO UPDATE SET is_active = true
        RETURNING id INTO v_promo_pl_id;
    END IF;

    -- If DELETING or DEACTIVATING promotion, remove/deactivate promo package prices
    IF TG_OP = 'DELETE' OR (TG_OP = 'UPDATE' AND NEW.is_active = false) THEN
        DELETE FROM product_prices
        WHERE price_list_id = v_promo_pl_id
          AND metadata->>'promotion_id' = COALESCE(OLD.id, NEW.id)::text;
        
        IF TG_OP = 'DELETE' THEN
            RETURN OLD;
        END IF;
        RETURN NEW;
    END IF;

    -- If INSERTING or UPDATING active promotion, generate promotional package prices
    IF NEW.is_active = true AND v_promo_pl_id IS NOT NULL THEN
        -- Clean up existing promotional prices for this promotion on UPDATE to ensure non-matching UOMs are purged
        IF TG_OP = 'UPDATE' THEN
            DELETE FROM product_prices
            WHERE price_list_id = v_promo_pl_id
              AND metadata->>'promotion_id' = NEW.id::text;
        END IF;

        -- Format discount percent string for metadata
        IF NEW.promotion_type = 'percentage_discount' AND NEW.discount_value IS NOT NULL THEN
            v_discount_percent_str := CONCAT(TRIM(TRAILING '.' FROM TRIM(TRAILING '0' FROM NEW.discount_value::text)), '%');
        ELSE
            v_discount_percent_str := NULL;
        END IF;

        -- Find all target products matching the promotion criteria
        FOR v_target_product_id IN (
            SELECT p.id
            FROM products p
            WHERE p.organization_id = NEW.organization_id
              AND p.is_active = true
              AND (
                  NEW.applies_to = 'all'
                  OR (NEW.applies_to = 'product' AND p.id = ANY(NEW.target_product_ids))
                  OR (NEW.applies_to = 'category' AND p.category_id = ANY(NEW.target_category_ids))
              )
        ) LOOP
            -- Loop through existing retail package prices for this product to compute promotional price per UOM
            FOR v_retail_pp IN (
                SELECT pp.*
                FROM product_prices pp
                JOIN price_lists pl ON pp.price_list_id = pl.id
                WHERE pp.product_id = v_target_product_id
                  AND pl.price_list_type = 'retail'
                  AND pp.is_active = true
                  AND (
                      NOT (NEW.metadata ? 'target_uoms')
                      OR NOT (NEW.metadata->'target_uoms' ? v_target_product_id::text)
                      OR (
                          jsonb_typeof(NEW.metadata->'target_uoms'->(v_target_product_id::text)) = 'array'
                          AND (NEW.metadata->'target_uoms'->(v_target_product_id::text)) @> pp.uom_id::text::jsonb
                      )
                      OR (NEW.metadata->'target_uoms'->>(v_target_product_id::text)) = pp.uom_id::text
                  )
                  AND (
                      NOT (NEW.metadata ? 'target_variants')
                      OR NOT (NEW.metadata->'target_variants' ? v_target_product_id::text)
                      OR pp.product_variant_id IS NULL
                      OR (
                          jsonb_typeof(NEW.metadata->'target_variants'->(v_target_product_id::text)) = 'array'
                          AND (NEW.metadata->'target_variants'->(v_target_product_id::text)) @> pp.product_variant_id::text::jsonb
                      )
                      OR (NEW.metadata->'target_variants'->>(v_target_product_id::text)) = pp.product_variant_id::text
                  )
            ) LOOP
                -- Calculate promo price
                IF NEW.promotion_type = 'percentage_discount' AND NEW.discount_value IS NOT NULL THEN
                    v_calculated_price := ROUND(v_retail_pp.price * (1.0 - (NEW.discount_value / 100.0)), 2);
                ELSIF NEW.promotion_type = 'fixed_discount' AND NEW.discount_value IS NOT NULL THEN
                    v_calculated_price := GREATEST(0.00, v_retail_pp.price - NEW.discount_value);
                ELSE
                    v_calculated_price := v_retail_pp.price;
                END IF;

                IF v_retail_pp.product_variant_id IS NOT NULL THEN
                    INSERT INTO product_prices (
                        price_list_id, product_id, product_variant_id, uom_id, price, is_active, metadata
                    )
                    VALUES (
                        v_promo_pl_id,
                        v_target_product_id,
                        v_retail_pp.product_variant_id,
                        v_retail_pp.uom_id,
                        v_calculated_price,
                        true,
                        jsonb_build_object(
                            'promotion_id', NEW.id,
                            'promotion_name', NEW.name,
                            'promotion_code', NEW.code,
                            'discount_percent', v_discount_percent_str,
                            'original_price', v_retail_pp.price
                        )
                    )
                    ON CONFLICT (price_list_id, product_id, COALESCE(product_variant_id, 0), uom_id)
                    DO UPDATE SET
                        price = EXCLUDED.price,
                        is_active = EXCLUDED.is_active,
                        metadata = EXCLUDED.metadata,
                        updated_at = CURRENT_TIMESTAMP;
                ELSIF NEW.metadata ? 'target_variants' AND NEW.metadata->'target_variants' ? v_target_product_id::text THEN
                    FOR v_variant_id IN (
                        SELECT (jsonb_array_elements_text(
                            CASE 
                                WHEN jsonb_typeof(NEW.metadata->'target_variants'->(v_target_product_id::text)) = 'array' 
                                THEN NEW.metadata->'target_variants'->(v_target_product_id::text)
                                ELSE jsonb_build_array(NEW.metadata->'target_variants'->>(v_target_product_id::text))
                            END
                        ))::INTEGER
                    ) LOOP
                        INSERT INTO product_prices (
                            price_list_id, product_id, product_variant_id, uom_id, price, is_active, metadata
                        )
                        VALUES (
                            v_promo_pl_id,
                            v_target_product_id,
                            v_variant_id,
                            v_retail_pp.uom_id,
                            v_calculated_price,
                            true,
                            jsonb_build_object(
                                'promotion_id', NEW.id,
                                'promotion_name', NEW.name,
                                'promotion_code', NEW.code,
                                'discount_percent', v_discount_percent_str,
                                'original_price', v_retail_pp.price
                            )
                        )
                        ON CONFLICT (price_list_id, product_id, COALESCE(product_variant_id, 0), uom_id)
                        DO UPDATE SET
                            price = EXCLUDED.price,
                            is_active = EXCLUDED.is_active,
                            metadata = EXCLUDED.metadata,
                            updated_at = CURRENT_TIMESTAMP;
                    END LOOP;
                ELSE
                    INSERT INTO product_prices (
                        price_list_id, product_id, product_variant_id, uom_id, price, is_active, metadata
                    )
                    VALUES (
                        v_promo_pl_id,
                        v_target_product_id,
                        NULL,
                        v_retail_pp.uom_id,
                        v_calculated_price,
                        true,
                        jsonb_build_object(
                            'promotion_id', NEW.id,
                            'promotion_name', NEW.name,
                            'promotion_code', NEW.code,
                            'discount_percent', v_discount_percent_str,
                            'original_price', v_retail_pp.price
                        )
                    )
                    ON CONFLICT (price_list_id, product_id, COALESCE(product_variant_id, 0), uom_id)
                    DO UPDATE SET
                        price = EXCLUDED.price,
                        is_active = EXCLUDED.is_active,
                        metadata = EXCLUDED.metadata,
                        updated_at = CURRENT_TIMESTAMP;
                END IF;
            END LOOP;
        END LOOP;
    END IF;

    RETURN NEW;
END;
$$;
