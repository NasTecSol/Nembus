# `stg-backup` inspection and isolated restore plan (superseded by executed Sweep 2)

> This plan was written before authenticated target access. The authorized restore was subsequently executed successfully using the existing PostgreSQL service into `nembus_migration_audit_20260922`. See `06-restore-execution.md` for the executed procedure and proof.

## Inspection result

| Property | Result |
|---|---|
| Exact path | `D:\stg-backup` |
| Extension | None |
| Size | 611,826,724 bytes |
| Last modified | 2026-09-21 12:55:22 (local time) |
| File signature | `PGDMP` |
| Format | PostgreSQL custom archive |
| Compression | gzip |
| TOC entries | 1,642 |
| Dump source | PostgreSQL 18.1, dumped by `pg_dump` 18.1 |
| Installed inspection tool | `pg_restore` 18.4 |
| Restore status at time of this plan | **NOT EXECUTED**; subsequently completed successfully in Sweep 2 |

`pg_restore --list D:\stg-backup` was used only to inspect archive metadata and the table of contents. The archive includes `public` and `staging` schemas, `uuid-ossp`, target types/functions, and the invoice totals function. No archive content was restored or modified.

## Proposed safe restore

This was the plan for a separately authorized follow-up. Its proposed database name was not used; the executed database name is:

```text
nembus_stg_audit_20260922_01
```

```text
Executed: nembus_migration_audit_20260922
```

The name is intentionally separate from application databases and does not use `--create` or `--clean`.

### Preflight checks

1. Confirm the PostgreSQL server/cluster and the intended maintenance connection are the local audit instance, not a production or shared application instance.
2. Confirm the proposed database name does not exist, using a read-only catalog query such as:

   ```sql
   SELECT datname
   FROM pg_database
   WHERE datname = 'nembus_stg_audit_20260922_01';
   ```

   The expected result is zero rows. If a row exists, stop and choose a new unique name; do not drop or overwrite it.
3. Confirm the PostgreSQL major version is compatible with the archive. The archive is from PostgreSQL 18.1 and the installed `pg_restore` is 18.4.
4. Confirm the archive checksum/ownership with the backup owner if an external checksum is available. No checksum was invented from the current inspection.
5. Confirm free space immediately before restore. The archive is approximately 0.57 GiB compressed; restored relation/index size may be materially larger. At inspection time the D: volume had approximately 27.8 GiB free.
6. Confirm that the restore role can create only the new database and required objects, and that ownership/privilege replay is intentionally disabled.

### Proposed commands

Use the PostgreSQL 18 binaries already installed under `C:\Program Files\PostgreSQL\18\bin`. After the name check and explicit authorization, the conceptual sequence is:

```powershell
$pg = 'C:\Program Files\PostgreSQL\18\bin'
$db = 'nembus_stg_audit_20260922_01'

& "$pg\createdb.exe" $db
& "$pg\pg_restore.exe" `
  --dbname=$db `
  --exit-on-error `
  --no-owner `
  --no-privileges `
  --jobs=4 `
  'D:\stg-backup'
```

The command intentionally omits `--clean`, `--create`, and any existing database target. If `createdb` reports that the name already exists, stop; do not force the operation.

### Read-only validation after restore

Run only read-only catalog and aggregate queries against the new database:

- confirm the expected `public` and `staging` schemas and relation list;
- count `invoices`, `invoice_lines`, `invoice_payments`, and staging batch rows;
- inspect target invoice currencies, statuses, discount columns, and null/default rates;
- locate the target row corresponding to source `OINV.DocEntry=80306` and line `INV1.LineNum=5`;
- compare source and target counts by document status, currency, date window, and representative discount sign;
- recompute target invoice totals from target lines and compare them with stored totals;
- inspect uniqueness and duplicate rates for source external keys;
- record restore warnings/errors without modifying the restored database.

No validation query should update statistics, create temporary/permanent tables, or alter settings. Use bounded samples and aggregates first.

### Cleanup plan, not executed

After validation is complete and the owner approves cleanup, confirm the database name one more time and remove only `nembus_stg_audit_20260922_01` from the isolated audit instance. Do not use a wildcard, `--all`, or a broad data-directory operation. The current audit did not create this database, so no cleanup action was necessary.

## Why restore was required for Sweep 2

The archive was available and structurally inspectable, but its contents were not target rows until restored. Sweep 2 completed that step and replaced the prior `TARGET_DATA_REQUIRED` conclusions with live target evidence. The executed results are in artifacts `06` through `13`.
