# On-demand cashier session Z-report

```http
GET /api/reports/z-report?cashier_session_id=845
Authorization: Bearer <user token>
x-tenant-id: <tenant>
```

Optional positive integer parameters: `store_id`, `pos_terminal_id`, `cashier_id`.
These must match the session; they do not trim the report's transaction set.
The authenticated user must belong to the organization and have a
`user_store_access` entry for the session's store. Organization and user identity
come from authentication, never query parameters. M2M tokens without user and
organization identity are not supported by this endpoint.

Both the local POS and cloud API register this route through the shared cashier
session routes. Each call reads the serving database in one SQL statement. It
does not close the session, update balances, persist a report, or print anything.
Cloud results include only records already synchronized to that database.

`finalized` is true when session status is `closed` and `closing_time` exists.
It means the session is closed, not that this response is an immutable snapshot.
Subsequent corrections/synchronization can change the next response. Open or
inconsistently closed sessions return false and null counted cash/variance.

## Response

The standard envelope is `{ "statusCode": 200, "message": "session Z-report
calculated successfully", "data": { ... } }`. `data` has this shape:

```json
{
  "report_type": "Z",
  "finalized": true,
  "generated_at": "2026-10-19T14:10:00+00:00",
  "calculation_version": 1,
  "currency": "SAR",
  "store": { "id": 3, "name": "Main Store", "timezone": "Asia/Riyadh" },
  "terminal": { "id": 7, "code": "POS-02", "name": "Counter 2" },
  "cashier": { "id": 12, "code": "CASH-012", "name": "Ahmed" },
  "session": {
    "id": 845,
    "number": "SESSION-845",
    "status": "closed",
    "opened_at": "2026-10-19T09:00:00",
    "closed_at": "2026-10-19T17:00:00"
  },
  "sales": {
    "transaction_count": 40,
    "before_discount_excluding_tax": "1100.00",
    "discounts": "100.00",
    "after_discount_excluding_tax": "1000.00",
    "tax": "150.00",
    "total_including_tax": "1150.00"
  },
  "voids": { "transaction_count": 2, "total_including_tax": "80.00" },
  "returns": {
    "return_count": 1,
    "excluding_tax": "100.00",
    "tax": "15.00",
    "total_refunded": "115.00"
  },
  "net": {
    "sales_excluding_tax": "900.00",
    "tax": "135.00",
    "total_including_tax": "1035.00"
  },
  "payments": [
    { "method": "cash", "payment_count": 25, "received": "700.00", "change_given": "50.00", "refunded": "115.00", "net": "535.00" },
    { "method": "card", "payment_count": 18, "received": "500.00", "change_given": "0.00", "refunded": "0.00", "net": "500.00" }
  ],
  "payment_totals": {
    "received": "1200.00",
    "change_given": "50.00",
    "refunded": "115.00",
    "net": "1035.00"
  },
  "drawer": {
    "opening_cash": "200.00",
    "cash_received": "700.00",
    "change_given": "50.00",
    "cash_refunded": "115.00",
    "manual_movements_included": false,
    "calculated_expected_cash": "735.00",
    "recorded_expected_cash": "735.00",
    "counted_closing_cash": "730.00",
    "variance": "-5.00"
  },
  "reconciliation": {
    "sales_payment_difference": "0.00",
    "expected_cash_difference": "0.00",
    "warnings": [
      { "code": "MANUAL_CASH_MOVEMENTS_NOT_INCLUDED", "message": "Manual cash additions and removals are not included." }
    ]
  }
}
```

All money is calculated with PostgreSQL numeric arithmetic and returned as
two-place decimal strings. Counts are JSON integers; `payments` is an empty
array for sessions without payments/refunds. Currency comes from the organization
and can be null if unconfigured. Names may be null. Session timestamps preserve
the existing database's timestamp-without-time-zone values; no timezone conversion
or invented UTC offset is applied. `generated_at` is a timestamp with offset.

## Calculation rules

- Include `sale` transactions with status `completed`, `partially_refunded`, or
  `refunded`. Missing transaction type is treated as sale for legacy data. Voided
  transactions are reported separately. Draft/pending transactions are excluded.
- Derive sales excluding tax from recorded `total_amount - tax_amount`. Add the
  recorded discount to obtain the amount before discount. This avoids relying
  on the different subtotal conventions of existing writers. Do not recalculate
  prices or tax from current product configuration.
- Aggregate sales and payments independently. Count each sale once even when
  it has multiple tenders. Payment counts count payment records, not unique sales.
- Include completed refunds under the session that processed them, regardless
  of the original sale's session. Preserve original sale amounts when its status
  becomes refunded. Include refund-only payment methods in the breakdown.
- Normalize payment/refund methods with trim/lowercase. Missing methods become
  `unknown`; do not assume cash or use transaction metadata as a payment ledger.
- Change comes from nonzero transaction `change_given`, otherwise the sum of
  cash payment metadata `change_given`, otherwise transaction metadata. Metadata
  accepts nonnegative plain decimal amounts; malformed values are ignored.
- Existing writers store either tendered or applied payment amounts. If payment
  totals equal the sale total and there is cash, reconstruct received cash by
  adding recorded change, then deduct change once. Otherwise treat amounts as
  tendered. Flag transactions whose resulting payment total differs from the sale.
- Expected drawer cash = opening balance + cash received - change - cash refunds.
  Variance = counted closing balance - calculated expected cash. Recorded expected
  cash is returned separately, with a warning if it differs. Manual drawer
  additions/removals are not supported by the current schema and always generate
  a warning. No fields imply those movements were recorded as zero.
- Reject sessions whose cashier and terminal stores disagree. Exclude transaction
  or return rows with inconsistent store/cashier/terminal references and report
  `INCONSISTENT_SESSION_RECORDS`. A null legacy transaction terminal is resolved
  through its session; a null return cashier is resolved through its session.

## Errors and verification

- 400: missing, duplicate, nonpositive, malformed, or overflowing ID filters.
- 401: missing/invalid authenticated organization or user.
- 404: missing/inaccessible session or a supplied filter does not match. This
  avoids disclosing sessions outside the caller's scope.
- 500: repository/database/report decoding failure.

PostgreSQL integration tests create only temporary fixture tables in a transaction
and roll them back. Run against a dedicated test database:

```powershell
$env:NEMBUS_REPORT_TEST_DATABASE_URL = 'postgres://postgres@localhost:55439/postgres?sslmode=disable'
cd packages/core
go test ./handler ./usecase ./repository ./routing -count=1
```

The query source is `queries/z_report.sql`; its repository adapter is generated
by sqlc. No database migrations are required.
