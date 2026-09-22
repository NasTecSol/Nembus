# Sweep 2 restore execution

**Status: `SUCCESS` — isolated Path A restore completed and target was left read-only.**

## Scope and safety

- Source backup: `D:\stg-backup` (PostgreSQL custom format, gzip, PGDMP; TOC entries `1642`; dumped by PostgreSQL 18.1).
- Existing service: `postgresql-x64-18`, `127.0.0.1:5432`, existing data directory `C:\Program Files\PostgreSQL\18\data`.
- New audit database: `nembus_migration_audit_20260922`.
- No existing database was dropped, overwritten, restored into, or altered.
- The application, migration service, source SAP database, backup file, and existing databases were not changed.
- The only database-level writes were creation of the new audit database and its database-level `default_transaction_read_only` setting. No business-table writes were issued after restore.

## Authentication and preflight

The two user-authorized password candidates were tried once each using process-scoped environment state. One candidate authenticated successfully; the other did not. No other passwords were attempted. The process-scoped secret was cleared immediately and was not written to any report or log.

Before creation, the target name was absent, active sessions for that name were zero, and the existing database list was:

```text
nembus_e2e_master | 16408
nembus_e2e_tenant | 16409
postgres          | 5
rms_db            | 16389
template0         | 4
template1         | 1
```

The target database name was therefore collision-free. Approximate free space before creation was 24.7 GB on `C:` and 29.9 GB on `D:`.

## Creation and restore

The first `createdb` invocation used an unsupported short option and failed before creating anything. A corrected invocation then created only the new database from `template0` with UTF-8 encoding. No destructive retry or cleanup was required.

The successful restore used the existing service and an explicit target database:

```text
pg_restore -h 127.0.0.1 -p 5432 -U postgres -w \
  --dbname=nembus_migration_audit_20260922 \
  --exit-on-error --no-owner --no-privileges --verbose D:\stg-backup
```

The command did not use `--clean` or `--create`. Restore completed successfully. The captured verbose log is [06-restore.log](D:\nastecsol\Nembus\migration-analysis\06-restore.log), 121,656 bytes. It contains the target connection, schema/object creation, data load, and final foreign-key creation; no `error`, `failed`, or `warning` tokens were found in the captured log.

## Post-restore proof

The new database was configured with:

```sql
ALTER DATABASE nembus_migration_audit_20260922
  SET default_transaction_read_only = on;
```

A fresh connection to the audit database returned `transaction_read_only=on` and `default_transaction_read_only=on`. Read-only checks returned `public.invoices=969,665` and `public.invoice_lines=5,472,485`. The new database remains present for audit reproducibility and was not stopped or deleted.

## Initial target inventory

The restored non-system application schemas are `atlas_schema_revisions`, `public`, and `staging`. The catalog contains 114 base tables including the revision and staging tables, 12 views, 98 sequences, and 38 public functions. The detailed inventory, parity, reconciliation, and regression artifacts are the companion files `07` through `13`.

## Existing service state

The existing PostgreSQL service was reused and left running. No temporary PostgreSQL 18 cluster was needed because the authorized credential succeeded and the new database could be isolated safely on the existing service.
