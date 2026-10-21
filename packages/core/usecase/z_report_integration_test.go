package usecase

import (
	"context"
	"os"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5"
)

// These temporary tables shadow application tables and disappear on rollback.
// Use a dedicated PostgreSQL database, never an application connection string.
func TestZReportPostgres(t *testing.T) {
	dsn := os.Getenv("NEMBUS_REPORT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_REPORT_TEST_DATABASE_URL to run PostgreSQL report tests")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
CREATE TEMP TABLE organizations (id int, currency_code text);
CREATE TEMP TABLE stores (id int, organization_id int, name text, timezone text);
CREATE TEMP TABLE users (id int, organization_id int, first_name text, last_name text, is_active boolean);
CREATE TEMP TABLE user_store_access (user_id int, store_id int);
CREATE TEMP TABLE cashiers (id int, user_id int, store_id int, cashier_code text);
CREATE TEMP TABLE pos_terminals (id int, store_id int, terminal_code text, terminal_name text);
CREATE TEMP TABLE cashier_sessions (id int, cashier_id int, pos_terminal_id int, session_number text,
 opening_time timestamp, closing_time timestamp, opening_balance numeric(15,2), closing_balance numeric(15,2),
 expected_balance numeric(15,2), status text);
CREATE TEMP TABLE pos_transactions (id int PRIMARY KEY, cashier_session_id int, store_id int, cashier_id int,
 pos_terminal_id int, transaction_type text, status text, total_amount numeric(15,2), tax_amount numeric(15,2),
 discount_amount numeric(15,2), change_given numeric(15,2), metadata jsonb);
CREATE TEMP TABLE pos_payments (transaction_id int, payment_method text, amount numeric(15,2), metadata jsonb);
CREATE TEMP TABLE sales_returns (cashier_session_id int, store_id int, cashier_id int, status text,
 total_refund_amount numeric(15,2), tax_amount numeric(15,2), refund_method text);
INSERT INTO organizations VALUES (1, 'SAR');
INSERT INTO stores VALUES (3,1,'Main','Asia/Riyadh');
INSERT INTO users VALUES (10,1,'Ahmed','Ali',true), (11,1,'No','Access',true), (12,2,'Other','Org',true);
INSERT INTO user_store_access VALUES (10,3), (12,3);
INSERT INTO cashiers VALUES (12,10,3,'CASH-012');
INSERT INTO pos_terminals VALUES (7,3,'POS-02','Counter 2');
INSERT INTO cashier_sessions VALUES
 (845,12,7,'SESSION-845','2026-10-19 09:00',NULL,200,NULL,965,'open'),
 (846,12,7,'SESSION-846','2026-10-19 09:00',NULL,100,NULL,100,'open');
INSERT INTO pos_transactions VALUES
 (1,845,3,12,7,'sale','completed',1150,150,100,50,'{}'),
 (2,845,3,12,7,'sale','refunded',230,30,0,0,'{}'),
 (3,845,3,12,7,'sale','partially_refunded',115,15,0,0,'{}'),
 (4,845,3,12,7,'sale','voided',80,0,0,0,'{}'),
 (5,845,3,12,7,'sale','draft',999,0,0,0,'{}'),
 (6,845,3,99,7,'sale','completed',35,0,0,0,'{}');
INSERT INTO pos_payments VALUES
 (1,' CASH ',700,'{}'), (1,'card',500,'{}'),
 (2,'cash',230,'{"change_given":"20"}'), (3,'card',115,'{}'),
 (4,'cash',80,'{}'), (5,'cash',999,'{}');
INSERT INTO sales_returns VALUES
 (845,3,12,'completed',115,15,'cash'), (845,3,12,'completed',23,3,'card'),
 (845,3,12,'pending',999,0,'cash'), (846,3,12,'cancelled',100,0,'cash');
`)
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.New(tx)
	filters := repository.GetCashierSessionZReportParams{CashierSessionID: 845, OrganizationID: 1, UserID: 10}
	calculate := func(t *testing.T, filters repository.GetCashierSessionZReportParams) ZReport {
		t.Helper()
		resp := CalculateCashierSessionZReport(ctx, repo, filters)
		if resp.StatusCode != 200 {
			t.Fatalf("report failed: %+v", resp)
		}
		return resp.Data.(ZReport)
	}
	assertMoney := func(label, got, want string) {
		t.Helper()
		if got != want {
			t.Errorf("%s: got %s, want %s", label, got, want)
		}
	}
	hasWarning := func(report ZReport, code string) bool {
		for _, warning := range report.Reconciliation.Warnings {
			if warning.Code == code {
				return true
			}
		}
		return false
	}
	t.Run("split payments refunds and applied cash", func(t *testing.T) {
		r := calculate(t, filters)
		if r.Finalized || r.Session.ClosedAt != nil || r.Drawer.CountedClosingCash != nil || r.Drawer.Variance != nil {
			t.Fatal("open session must not have finalized/count/variance")
		}
		if r.Sales.TransactionCount != 3 || r.Voids.TransactionCount != 1 || r.Returns.ReturnCount != 2 {
			t.Fatalf("incorrect counts: %+v %+v %+v", r.Sales, r.Voids, r.Returns)
		}
		assertMoney("gross", r.Sales.BeforeDiscountExcludingTax, "1400.00")
		assertMoney("sales", r.Sales.TotalIncludingTax, "1495.00")
		assertMoney("net", r.Net.TotalIncludingTax, "1357.00")
		assertMoney("net tax", r.Net.Tax, "177.00")
		assertMoney("received", r.PaymentTotals.Received, "1565.00")
		assertMoney("change", r.PaymentTotals.ChangeGiven, "70.00")
		assertMoney("payment net", r.PaymentTotals.Net, "1357.00")
		assertMoney("expected cash", r.Drawer.CalculatedExpectedCash, "965.00")
		assertMoney("payment difference", r.Reconciliation.SalesPaymentDifference, "0.00")
		if hasWarning(r, "PAYMENT_MISMATCH") || !hasWarning(r, "INCONSISTENT_SESSION_RECORDS") {
			t.Fatal("incorrect warnings")
		}
	})
	t.Run("closed session and fresh calculation", func(t *testing.T) {
		_, err := tx.Exec(ctx, `UPDATE cashier_sessions SET status='closed', closing_time='2026-10-19 17:00', closing_balance=960 WHERE id=845`)
		if err != nil {
			t.Fatal(err)
		}
		r := calculate(t, filters)
		if !r.Finalized || r.Drawer.Variance == nil {
			t.Fatal("closed session must be finalized with variance")
		}
		assertMoney("variance", *r.Drawer.Variance, "-5.00")
		_, err = tx.Exec(ctx, `INSERT INTO sales_returns VALUES (845,3,12,'completed',11.50,1.50,'cash')`)
		if err != nil {
			t.Fatal(err)
		}
		r = calculate(t, filters)
		assertMoney("fresh net", r.Net.TotalIncludingTax, "1345.50")
		assertMoney("fresh variance", *r.Drawer.Variance, "6.50")
		if !hasWarning(r, "EXPECTED_CASH_MISMATCH") {
			t.Fatal("expected mismatch warning")
		}
	})
	t.Run("empty session", func(t *testing.T) {
		f := filters
		f.CashierSessionID = 846
		r := calculate(t, f)
		if r.Payments == nil || len(r.Payments) != 0 || r.Sales.TransactionCount != 0 {
			t.Fatal("empty session arrays/counts")
		}
		assertMoney("empty net", r.Net.TotalIncludingTax, "0.00")
		assertMoney("empty cash", r.Drawer.CalculatedExpectedCash, "100.00")
	})
	t.Run("missing payments and unknown refund", func(t *testing.T) {
		_, err := tx.Exec(ctx, `INSERT INTO pos_transactions VALUES (7,846,3,12,7,'sale','completed',10,0,0,0,'{}');
INSERT INTO sales_returns VALUES (846,3,12,'completed',2,0,NULL)`)
		if err != nil {
			t.Fatal(err)
		}
		f := filters
		f.CashierSessionID = 846
		r := calculate(t, f)
		if !hasWarning(r, "MISSING_PAYMENTS") || !hasWarning(r, "UNKNOWN_PAYMENT_METHOD") || !hasWarning(r, "PAYMENT_MISMATCH") {
			t.Fatal("missing data warnings")
		}
		assertMoney("unknown refund is not assumed cash", r.Drawer.CashRefunded, "0.00")
		assertMoney("missing payment difference", r.Reconciliation.SalesPaymentDifference, "10.00")
	})
	t.Run("access and filters", func(t *testing.T) {
		for _, change := range []func(*repository.GetCashierSessionZReportParams){
			func(f *repository.GetCashierSessionZReportParams) { f.OrganizationID = 2 },
			func(f *repository.GetCashierSessionZReportParams) { f.UserID = 11 },
			func(f *repository.GetCashierSessionZReportParams) { f.UserID = 12 },
			func(f *repository.GetCashierSessionZReportParams) { f.StoreID = 99 },
			func(f *repository.GetCashierSessionZReportParams) { f.PosTerminalID = 99 },
			func(f *repository.GetCashierSessionZReportParams) { f.CashierID = 99 },
			func(f *repository.GetCashierSessionZReportParams) { f.CashierSessionID = 99 },
		} {
			f := filters
			change(&f)
			if resp := CalculateCashierSessionZReport(ctx, repo, f); resp.StatusCode != 404 {
				t.Fatalf("scope leaked for %+v: %+v", f, resp)
			}
		}
		f := filters
		f.StoreID = 3
		f.PosTerminalID = 7
		f.CashierID = 12
		calculate(t, f)
	})
}
