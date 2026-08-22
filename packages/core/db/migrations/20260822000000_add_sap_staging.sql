-- Migration: restore the canonical SAP staging objects to Atlas history.
-- The same forward migration is applied to master and tenant databases.

CREATE SCHEMA IF NOT EXISTS staging;

CREATE TABLE IF NOT EXISTS staging.sap_migration_batches (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) UNIQUE NOT NULL,
    run_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    domain VARCHAR(50) NOT NULL,
    record_count INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(30) DEFAULT 'staged',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS staging.sap_stores (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    store_type VARCHAR(50),
    is_warehouse BOOLEAN DEFAULT true,
    is_pos_enabled BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS staging.sap_products (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_code VARCHAR(50),
    brand_code VARCHAR(50),
    uom_code VARCHAR(20),
    product_type VARCHAR(50) DEFAULT 'standard',
    is_serialized BOOLEAN DEFAULT false,
    is_batch_managed BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    is_sellable BOOLEAN DEFAULT true,
    is_purchasable BOOLEAN DEFAULT true,
    track_inventory BOOLEAN DEFAULT true,
    primary_barcode VARCHAR(100),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS staging.sap_inventory (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    product_sku VARCHAR(100) NOT NULL,
    store_code VARCHAR(50) NOT NULL,
    quantity_on_hand DECIMAL(15,3) DEFAULT 0,
    quantity_allocated DECIMAL(15,3) DEFAULT 0,
    quantity_available DECIMAL(15,3) DEFAULT 0,
    quantity_on_order DECIMAL(15,3) DEFAULT 0,
    reorder_level DECIMAL(15,3),
    max_stock_level DECIMAL(15,3),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Existing staging objects are accepted only when required canonical
-- columns/defaults and constraints remain compatible. No reconciliation or
-- data mutation is attempted.
DO $$
DECLARE
    table_name TEXT;
    column_spec JSONB;
    constraint_spec JSONB;
    actual_type TEXT;
    actual_default TEXT;
    actual_not_null BOOLEAN;
    default_kind TEXT;
    expected_columns JSONB := $schema$
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
    $schema$::JSONB;
    expected_constraints JSONB := $constraints$
    [
      {"table":"sap_migration_batches","type":"p","key":[1]},
      {"table":"sap_migration_batches","type":"u","key":[2]},
      {"table":"sap_migration_batches","type":"f","key":[4],"ref":"organizations","ref_key":[1],"on_delete":"c"},
      {"table":"sap_stores","type":"p","key":[1]},
      {"table":"sap_products","type":"p","key":[1]},
      {"table":"sap_inventory","type":"p","key":[1]}
    ]
    $constraints$::JSONB;
BEGIN
    FOR table_name IN SELECT jsonb_object_keys(expected_columns) LOOP
        IF to_regclass(format('staging.%I', table_name)) IS NULL THEN
            RAISE EXCEPTION 'canonical SAP staging table staging.% is missing', table_name;
        END IF;

        FOR column_spec IN SELECT value FROM jsonb_array_elements(expected_columns -> table_name) LOOP
            SELECT format_type(a.atttypid, a.atttypmod), a.attnotnull,
                   pg_get_expr(d.adbin, d.adrelid)
              INTO actual_type, actual_not_null, actual_default
              FROM pg_attribute a
              LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
             WHERE a.attrelid = format('staging.%I', table_name)::regclass
               AND a.attnum > 0
               AND NOT a.attisdropped
               AND a.attname = column_spec ->> 'name';

            default_kind := column_spec ->> 'default';
            IF NOT FOUND
               OR actual_type <> (column_spec ->> 'type')
               OR actual_not_null <> (column_spec ->> 'not_null')::BOOLEAN
               OR NOT COALESCE(CASE default_kind
                    WHEN 'none' THEN actual_default IS NULL
                    WHEN 'serial' THEN actual_default LIKE 'nextval(%'
                    WHEN 'zero' THEN regexp_replace(actual_default, '\s+', '', 'g') = '0'
                    WHEN 'staged' THEN actual_default LIKE '%staged%'
                    WHEN 'standard' THEN actual_default LIKE '%standard%'
                    WHEN 'true' THEN lower(actual_default) LIKE 'true%'
                    WHEN 'false' THEN lower(actual_default) LIKE 'false%'
                    WHEN 'jsonb_empty' THEN actual_default LIKE '%{}%'
                    WHEN 'current_timestamp' THEN upper(actual_default) LIKE '%CURRENT_TIMESTAMP%'
                    ELSE false
                  END, false)
            THEN
                RAISE EXCEPTION 'staging.% column % is incompatible with canonical SAP staging schema', table_name, column_spec ->> 'name';
            END IF;
        END LOOP;
    END LOOP;

    FOR constraint_spec IN SELECT value FROM jsonb_array_elements(expected_constraints) LOOP
        IF NOT EXISTS (
            SELECT 1
              FROM pg_constraint c
             WHERE c.conrelid = format('staging.%I', constraint_spec ->> 'table')::regclass
               AND c.contype = (constraint_spec ->> 'type')
               AND c.conkey = ARRAY(SELECT value::SMALLINT FROM jsonb_array_elements_text(constraint_spec -> 'key'))
               AND ((constraint_spec ->> 'type') <> 'f' OR (
                    c.confrelid = CASE WHEN constraint_spec ->> 'type' = 'f'
                                       THEN format('%I', constraint_spec ->> 'ref')::regclass END
                AND c.confkey = CASE WHEN constraint_spec ->> 'type' = 'f'
                                     THEN ARRAY(SELECT value::SMALLINT FROM jsonb_array_elements_text(constraint_spec -> 'ref_key')) END
                AND c.confdeltype = (constraint_spec ->> 'on_delete')
               ))
        ) THEN
            RAISE EXCEPTION 'staging.% is missing a canonical constraint', constraint_spec ->> 'table';
        END IF;
    END LOOP;
END
$$;
