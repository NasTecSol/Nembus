-- Validate the canonical SAP staging contract after provisioning.
-- This migration is intentionally fail-closed and non-destructive: it never
-- reconciles, repairs, or mutates an existing staging object.

DO $validate_sap_staging_contract$
DECLARE
    table_name TEXT;
    table_oid OID;
    column_spec JSONB;
    actual_type TEXT;
    actual_default TEXT;
    actual_not_null BOOLEAN;
    default_kind TEXT;
    sequence_name TEXT;
    default_matches BOOLEAN;
    expected_columns JSONB := $columns$
    {
      "sap_migration_batches": [
        {"name":"id","type":"integer","not_null":true,"default":"serial"},
        {"name":"batch_id","type":"character varying(100)","not_null":true,"default":"none"},
        {"name":"run_id","type":"character varying(100)","not_null":true,"default":"none"},
        {"name":"organization_id","type":"integer","not_null":true,"default":"none"},
        {"name":"domain","type":"character varying(50)","not_null":true,"default":"none"},
        {"name":"record_count","type":"integer","not_null":true,"default":"zero"},
        {"name":"status","type":"character varying(30)","not_null":false,"default":"staged"},
        {"name":"error_message","type":"text","not_null":false,"default":"none"},
        {"name":"created_at","type":"timestamp without time zone","not_null":false,"default":"current_timestamp"}
      ],
      "sap_stores": [
        {"name":"id","type":"integer","not_null":true,"default":"serial"},
        {"name":"batch_id","type":"character varying(100)","not_null":true,"default":"none"},
        {"name":"organization_id","type":"integer","not_null":true,"default":"none"},
        {"name":"code","type":"character varying(50)","not_null":true,"default":"none"},
        {"name":"name","type":"character varying(255)","not_null":true,"default":"none"},
        {"name":"store_type","type":"character varying(50)","not_null":false,"default":"none"},
        {"name":"is_warehouse","type":"boolean","not_null":false,"default":"true"},
        {"name":"is_pos_enabled","type":"boolean","not_null":false,"default":"true"},
        {"name":"is_active","type":"boolean","not_null":false,"default":"true"},
        {"name":"metadata","type":"jsonb","not_null":false,"default":"jsonb_empty"},
        {"name":"created_at","type":"timestamp without time zone","not_null":false,"default":"current_timestamp"}
      ],
      "sap_products": [
        {"name":"id","type":"integer","not_null":true,"default":"serial"},
        {"name":"batch_id","type":"character varying(100)","not_null":true,"default":"none"},
        {"name":"organization_id","type":"integer","not_null":true,"default":"none"},
        {"name":"sku","type":"character varying(100)","not_null":true,"default":"none"},
        {"name":"name","type":"character varying(255)","not_null":true,"default":"none"},
        {"name":"description","type":"text","not_null":false,"default":"none"},
        {"name":"category_code","type":"character varying(50)","not_null":false,"default":"none"},
        {"name":"brand_code","type":"character varying(50)","not_null":false,"default":"none"},
        {"name":"uom_code","type":"character varying(20)","not_null":false,"default":"none"},
        {"name":"product_type","type":"character varying(50)","not_null":false,"default":"standard"},
        {"name":"is_serialized","type":"boolean","not_null":false,"default":"false"},
        {"name":"is_batch_managed","type":"boolean","not_null":false,"default":"false"},
        {"name":"is_active","type":"boolean","not_null":false,"default":"true"},
        {"name":"is_sellable","type":"boolean","not_null":false,"default":"true"},
        {"name":"is_purchasable","type":"boolean","not_null":false,"default":"true"},
        {"name":"track_inventory","type":"boolean","not_null":false,"default":"true"},
        {"name":"primary_barcode","type":"character varying(100)","not_null":false,"default":"none"},
        {"name":"metadata","type":"jsonb","not_null":false,"default":"jsonb_empty"},
        {"name":"created_at","type":"timestamp without time zone","not_null":false,"default":"current_timestamp"}
      ],
      "sap_inventory": [
        {"name":"id","type":"integer","not_null":true,"default":"serial"},
        {"name":"batch_id","type":"character varying(100)","not_null":true,"default":"none"},
        {"name":"organization_id","type":"integer","not_null":true,"default":"none"},
        {"name":"product_sku","type":"character varying(100)","not_null":true,"default":"none"},
        {"name":"store_code","type":"character varying(50)","not_null":true,"default":"none"},
        {"name":"quantity_on_hand","type":"numeric(15,3)","not_null":false,"default":"zero"},
        {"name":"quantity_allocated","type":"numeric(15,3)","not_null":false,"default":"zero"},
        {"name":"quantity_available","type":"numeric(15,3)","not_null":false,"default":"zero"},
        {"name":"quantity_on_order","type":"numeric(15,3)","not_null":false,"default":"zero"},
        {"name":"reorder_level","type":"numeric(15,3)","not_null":false,"default":"none"},
        {"name":"max_stock_level","type":"numeric(15,3)","not_null":false,"default":"none"},
        {"name":"metadata","type":"jsonb","not_null":false,"default":"jsonb_empty"},
        {"name":"created_at","type":"timestamp without time zone","not_null":false,"default":"current_timestamp"}
      ]
    }
    $columns$::JSONB;
BEGIN
    FOR table_name IN SELECT jsonb_object_keys(expected_columns) LOOP
        table_oid := NULL;
        SELECT c.oid
          INTO table_oid
          FROM pg_class c
          JOIN pg_namespace n ON n.oid = c.relnamespace
         WHERE n.nspname = 'staging'
           AND c.relname = table_name
           AND c.relkind = 'r';

        IF NOT FOUND THEN
            RAISE EXCEPTION 'canonical SAP staging table staging.% is missing or is not a table', table_name;
        END IF;

        FOR column_spec IN
            SELECT value FROM jsonb_array_elements(expected_columns -> table_name)
        LOOP
            actual_type := NULL;
            actual_default := NULL;
            actual_not_null := NULL;

            SELECT format_type(a.atttypid, a.atttypmod),
                   a.attnotnull,
                   pg_get_expr(d.adbin, d.adrelid)
              INTO actual_type, actual_not_null, actual_default
              FROM pg_attribute a
              LEFT JOIN pg_attrdef d
                ON d.adrelid = a.attrelid
               AND d.adnum = a.attnum
             WHERE a.attrelid = table_oid
               AND a.attnum > 0
               AND NOT a.attisdropped
               AND a.attname = column_spec ->> 'name';

            IF NOT FOUND
               OR actual_type IS DISTINCT FROM (column_spec ->> 'type')
               OR actual_not_null IS DISTINCT FROM (column_spec ->> 'not_null')::BOOLEAN
            THEN
                RAISE EXCEPTION 'staging.% column % has incompatible type or nullability', table_name, column_spec ->> 'name';
            END IF;

            default_kind := column_spec ->> 'default';
            default_matches := FALSE;

            IF default_kind = 'none' THEN
                default_matches := actual_default IS NULL;
            ELSIF default_kind = 'serial' THEN
                SELECT pg_get_serial_sequence(format('staging.%I', table_name), column_spec ->> 'name')
                  INTO sequence_name;
                default_matches := sequence_name IS NOT NULL
                    AND actual_default = format('nextval(%L::regclass)', sequence_name);
            ELSE
                default_matches := COALESCE(
                    regexp_replace(lower(trim(COALESCE(actual_default, ''))), '\s+', '', 'g')
                        = CASE default_kind
                            WHEN 'zero' THEN '0'
                            WHEN 'staged' THEN quote_literal('staged') || '::charactervarying'
                            WHEN 'standard' THEN quote_literal('standard') || '::charactervarying'
                            WHEN 'true' THEN 'true'
                            WHEN 'false' THEN 'false'
                            WHEN 'jsonb_empty' THEN quote_literal('{}') || '::jsonb'
                            WHEN 'current_timestamp' THEN 'current_timestamp'
                            ELSE NULL
                          END,
                    FALSE
                );
            END IF;

            IF NOT default_matches THEN
                RAISE EXCEPTION 'staging.% column % has an incompatible canonical default', table_name, column_spec ->> 'name';
            END IF;
        END LOOP;
    END LOOP;

    IF NOT EXISTS (
        SELECT 1
          FROM pg_constraint c
         WHERE c.conrelid = 'staging.sap_migration_batches'::regclass
           AND c.contype = 'p'
           AND c.conkey = ARRAY[
               (SELECT a.attnum FROM pg_attribute a
                WHERE a.attrelid = c.conrelid AND a.attname = 'id')
           ]::SMALLINT[]
    ) THEN
        RAISE EXCEPTION 'staging.sap_migration_batches is missing the canonical primary key on id';
    END IF;

    IF NOT EXISTS (
        SELECT 1
          FROM pg_constraint c
         WHERE c.conrelid = 'staging.sap_migration_batches'::regclass
           AND c.contype = 'u'
           AND c.conkey = ARRAY[
               (SELECT a.attnum FROM pg_attribute a
                WHERE a.attrelid = c.conrelid AND a.attname = 'batch_id')
           ]::SMALLINT[]
    ) THEN
        RAISE EXCEPTION 'staging.sap_migration_batches is missing the canonical unique constraint on batch_id';
    END IF;

    FOREACH table_name IN ARRAY ARRAY['sap_stores', 'sap_products', 'sap_inventory'] LOOP
        IF NOT EXISTS (
            SELECT 1
              FROM pg_constraint c
             WHERE c.conrelid = format('staging.%I', table_name)::regclass
               AND c.contype = 'p'
               AND c.conkey = ARRAY[
                   (SELECT a.attnum FROM pg_attribute a
                    WHERE a.attrelid = c.conrelid AND a.attname = 'id')
               ]::SMALLINT[]
        ) THEN
            RAISE EXCEPTION 'staging.% is missing the canonical primary key on id', table_name;
        END IF;
    END LOOP;

    IF NOT EXISTS (
        SELECT 1
          FROM pg_constraint c
          JOIN pg_class referenced_table ON referenced_table.oid = c.confrelid
          JOIN pg_namespace referenced_schema ON referenced_schema.oid = referenced_table.relnamespace
         WHERE c.conrelid = 'staging.sap_migration_batches'::regclass
           AND c.contype = 'f'
           AND c.conkey = ARRAY[
               (SELECT a.attnum FROM pg_attribute a
                WHERE a.attrelid = c.conrelid AND a.attname = 'organization_id')
           ]::SMALLINT[]
           AND referenced_schema.nspname = 'public'
           AND referenced_table.relname = 'organizations'
           AND c.confkey = ARRAY[
               (SELECT a.attnum FROM pg_attribute a
                WHERE a.attrelid = c.confrelid AND a.attname = 'id')
           ]::SMALLINT[]
           AND c.confdeltype = 'c'
           AND c.confupdtype = 'a'
    ) THEN
        RAISE EXCEPTION 'staging.sap_migration_batches is missing the canonical organizations FK on organization_id with ON DELETE CASCADE and ON UPDATE NO ACTION';
    END IF;
END
$validate_sap_staging_contract$;
