package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const receiptTestTransactionUUID = "00000000-0000-4000-8000-000000000006"
const receiptTestMissingTransactionUUID = "00000000-0000-4000-8000-000000000099"

type receiptQRDB struct {
	repository.DBTX
	qr    string
	calls int
	txID  string
	order string
	err   error
}
type receiptQRRow struct {
	qr  string
	err error
}

func (r receiptQRRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	*dest[0].(*string) = r.qr
	return nil
}
func (db *receiptQRDB) QueryRow(_ context.Context, _ string, args ...interface{}) pgx.Row {
	db.calls++
	if len(args) == 1 {
		db.order = args[0].(string)
	} else if len(args) >= 1 {
		if u, ok := args[0].(pgtype.UUID); ok && u.Valid {
			db.txID = uuid.UUID(u.Bytes).String()
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				db.order = s
			}
		}
	}
	return receiptQRRow{db.qr, db.err}
}

func TestReceiptQRLoadsOnlySavedPayloadFromRequestTenant(t *testing.T) {
	for _, test := range []struct {
		name, query string
		code        int
		calls       int
		missing     bool
	}{
		{"sale", "transaction_id=" + receiptTestTransactionUUID, 200, 1, false},
		{"order", "sales_order_id=00000000-0000-4000-8000-000000000006", 200, 1, false},
		{"bad sale", "transaction_id=not-a-uuid", 400, 0, false},
		{"bad order", "sales_order_id=invalid", 400, 0, false},
		{"missing sale", "transaction_id=" + receiptTestMissingTransactionUUID, 404, 1, true},
		{"order without signed QR", "sales_order_id=00000000-0000-4000-8000-000000000007", 404, 1, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			db := &receiptQRDB{qr: "exact-signed-base64-payload"}
			if test.name == "order without signed QR" {
				db.qr = ""
			}
			if test.missing {
				db.err = pgx.ErrNoRows
			}
			engine := gin.New()
			engine.Use(func(c *gin.Context) {
				c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), middleware.RepoKey, repository.New(db)))
			})
			engine.GET("/receipt-qr", NewZatcaHandler(nil).GetReceiptQR)
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, httptest.NewRequest("GET", "/receipt-qr?"+test.query, nil))
			if recorder.Code != test.code || db.calls != test.calls {
				t.Fatalf("code=%d calls=%d body=%s", recorder.Code, db.calls, recorder.Body)
			}
			if test.code == 200 {
				var body struct {
					Data struct {
						QR string `json:"qr_code_base64"`
					} `json:"data"`
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Data.QR != db.qr {
					t.Fatal("saved QR changed")
				}
				if test.name == "sale" && db.txID != receiptTestTransactionUUID {
					t.Fatalf("wrong transaction lookup: got %q", db.txID)
				}
				if test.name == "order" && (db.txID != "" || db.order == "") {
					t.Fatal("wrong order lookup")
				}
			}
		})
	}
}