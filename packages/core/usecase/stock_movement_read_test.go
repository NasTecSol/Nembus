package usecase

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestStockMovementReadPreservesResponse(t *testing.T) {
	var quantity pgtype.Numeric
	if err := quantity.Scan("12.345"); err != nil {
		t.Fatal(err)
	}
	legacy := repository.StockMovement{
		ID: 42, MovementType: "sale", ReferenceType: pgtype.Text{String: "invoice", Valid: true},
		ReferenceID: pgtype.Int4{Int32: 123, Valid: true}, ProductID: 7,
		ProductVariantID: pgtype.Int4{Int32: 8, Valid: true},
		FromStoreID:      pgtype.Int4{Int32: 9, Valid: true}, ToStoreID: pgtype.Int4{Int32: 10, Valid: true},
		FromLocationID: pgtype.Int4{Int32: 11, Valid: true}, ToLocationID: pgtype.Int4{Int32: 12, Valid: true},
		Quantity: quantity, UomID: pgtype.Int4{Int32: 13, Valid: true},
		BatchNumber: pgtype.Text{String: "batch", Valid: true}, SerialNumber: pgtype.Text{String: "serial", Valid: true},
		MovementDate: pgtype.Timestamp{Time: time.Date(2026, 10, 23, 1, 2, 3, 0, time.UTC), Valid: true},
		PostedBy:     pgtype.Int4{Int32: 14, Valid: true}, Status: pgtype.Text{String: "completed", Valid: true},
		CostPerUnit: quantity, TotalValue: quantity, Metadata: json.RawMessage(`{"source":"test"}`),
		CreatedAt: pgtype.Timestamp{Time: time.Date(2026, 10, 22, 1, 2, 3, 0, time.UTC), Valid: true},
	}
	for _, tc := range []struct{ name, reference, want string }{
		{"uuid", "c469e140-a36f-4074-90e1-f1f265690086", `"c469e140-a36f-4074-90e1-f1f265690086"`},
		{"integer", "123", "123"},
		{"null", "", "null"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var row repository.ListStockMovementsRow
			// Copy every unchanged field to verify the new mapping preserves the existing response.
			from, to := reflect.ValueOf(legacy), reflect.ValueOf(&row).Elem()
			for i := 0; i < from.NumField(); i++ {
				name := from.Type().Field(i).Name
				if name != "ReferenceID" {
					to.FieldByName(name).Set(from.Field(i))
				}
			}
			if err := pgtype.NewMap().Scan(pgtype.TextOID, pgtype.TextFormatCode, []byte(tc.reference), &row.ReferenceID); err != nil {
				t.Fatal(err)
			}
			// Both generated list result types must support exactly the same mapping.
			filtered := repository.ListStockMovementsByReferenceTypeRow(row)
			gotBytes, err := json.Marshal(stockMovementReadToOutput(repository.ListStockMovementsRow(filtered)))
			if err != nil {
				t.Fatal(err)
			}
			wantBytes, err := json.Marshal(stockMovementBasicToOutput(legacy))
			if err != nil {
				t.Fatal(err)
			}
			var got, want map[string]json.RawMessage
			if err := json.Unmarshal(gotBytes, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(wantBytes, &want); err != nil {
				t.Fatal(err)
			}
			want["reference_id"] = json.RawMessage(tc.want)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("response changed: got %s; want fields %v", gotBytes, want)
			}
		})
	}
}
