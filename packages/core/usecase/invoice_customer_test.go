package usecase

import (
	"encoding/json"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestInvoiceCustomerNullScanAndResponse(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source []byte
		want   string
	}{
		{"no customer", nil, "null"},
		{"existing customer", []byte("42"), "42"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var invoice repository.Invoice
			// Exercise the PostgreSQL driver's NULL scanning path used by ListInvoices.
			if err := pgtype.NewMap().Scan(pgtype.Int4OID, pgtype.TextFormatCode, tc.source, &invoice.CustomerID); err != nil {
				t.Fatalf("scan invoice customer: %v", err)
			}
			output := invoiceToOutput(invoice, nil, nil)
			encoded, err := json.Marshal(output.CustomerID)
			if err != nil {
				t.Fatal(err)
			}
			if string(encoded) != tc.want {
				t.Fatalf("customer_id JSON = %s, want %s", encoded, tc.want)
			}
		})
	}
}
