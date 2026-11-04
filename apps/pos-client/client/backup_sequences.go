package client

import (
	"context"
	"fmt"
)

// repairRestoredSequences advances owned serial/identity sequences past restored
// rows. Keep counters that are already ahead, and leave empty tables unchanged.
// This runs only during clone restore, before the database is used by the POS.
func repairRestoredSequences(ctx context.Context, conn statementExecutor) error {
	_, err := conn.Exec(ctx, `
DO $repair_sequences$
DECLARE
    owned RECORD;
    highest_id BIGINT;
    sequence_value BIGINT;
    sequence_called BOOLEAN;
BEGIN
    FOR owned IN
        SELECT ns.nspname AS schema_name, tbl.relname AS table_name,
               col.attname AS column_name, seq.oid AS sequence_oid
        FROM pg_catalog.pg_class seq
        JOIN pg_catalog.pg_sequence settings ON settings.seqrelid = seq.oid
        JOIN pg_catalog.pg_depend dep
          ON dep.classid = 'pg_catalog.pg_class'::regclass
         AND dep.objid = seq.oid
         AND dep.refclassid = 'pg_catalog.pg_class'::regclass
         AND dep.deptype IN ('a', 'i')
        JOIN pg_catalog.pg_class tbl ON tbl.oid = dep.refobjid
        JOIN pg_catalog.pg_namespace ns ON ns.oid = tbl.relnamespace
        JOIN pg_catalog.pg_attribute col
          ON col.attrelid = tbl.oid AND col.attnum = dep.refobjsubid
        WHERE seq.relkind = 'S'
          AND ns.nspname IN ('public', 'staging')
          AND NOT col.attisdropped
          AND settings.seqincrement > 0
          AND NOT settings.seqcycle
        ORDER BY ns.nspname, tbl.relname, col.attnum
    LOOP
        EXECUTE format('LOCK TABLE %I.%I IN SHARE ROW EXCLUSIVE MODE',
                       owned.schema_name, owned.table_name);
        EXECUTE format('SELECT max(%I)::bigint FROM %I.%I',
                       owned.column_name, owned.schema_name, owned.table_name)
            INTO highest_id;
        EXECUTE format('SELECT last_value, is_called FROM %s',
                       owned.sequence_oid::regclass)
            INTO sequence_value, sequence_called;
        IF highest_id IS NOT NULL AND
           (highest_id > sequence_value OR
            (highest_id = sequence_value AND NOT sequence_called)) THEN
            PERFORM pg_catalog.setval(owned.sequence_oid::regclass, highest_id, true);
        END IF;
    END LOOP;
END;
$repair_sequences$;
`)
	if err != nil {
		return fmt.Errorf("repair restored ID sequences: %w", err)
	}
	return nil
}
