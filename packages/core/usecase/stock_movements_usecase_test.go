package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type stockMovementTestDB struct {
	repository.DBTX
	count     int64
	types     []string
	err       error
	countArgs []any
	listArgs  []any
}

func (db *stockMovementTestDB) QueryRow(_ context.Context, _ string, args ...any) pgx.Row {
	db.countArgs = args
	return stockMovementCountRow{count: db.count, err: db.err}
}

func (db *stockMovementTestDB) Query(_ context.Context, _ string, args ...any) (pgx.Rows, error) {
	db.listArgs = args
	return &stockMovementTypeRows{values: db.types}, db.err
}

type stockMovementCountRow struct {
	count int64
	err   error
}

func (row stockMovementCountRow) Scan(dest ...any) error {
	*dest[0].(*int64) = row.count
	return row.err
}

type stockMovementTypeRows struct {
	pgx.Rows
	values []string
	index  int
}

func (rows *stockMovementTypeRows) Close()     {}
func (rows *stockMovementTypeRows) Err() error { return nil }
func (rows *stockMovementTypeRows) Next() bool { return rows.index < len(rows.values) }
func (rows *stockMovementTypeRows) Scan(dest ...any) error {
	*dest[0].(*string) = rows.values[rows.index]
	rows.index++
	return nil
}

func TestStockMovementsReferenceTypePagination(t *testing.T) {
	for _, tc := range []struct {
		name                                       string
		page, limit                                int32
		count                                      int64
		wantPage, wantLimit, wantOffset, wantPages int32
	}{
		{"filtered totals", 3, 2, 5, 3, 2, 4, 3},
		{"unknown type", 1, 100, 0, 1, 100, 0, 0},
		{"defaults", 0, 0, 101, 1, 100, 0, 2},
		{"limit cap", 2, 1000, 501, 2, 500, 500, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &stockMovementTestDB{count: tc.count}
			uc := NewStockMovementsUseCase()
			uc.SetRepository(repository.New(db))
			resp := uc.ListStockMovementsByReferenceType(context.Background(), " sales_order ", tc.page, tc.limit)
			if resp.StatusCode != 200 {
				t.Fatalf("unexpected response: %+v", resp)
			}
			data := resp.Data.(StockMovementReadListResponse)
			if data.TotalCount != tc.count || data.Page != tc.wantPage || data.Limit != tc.wantLimit || data.TotalPages != tc.wantPages || data.Data == nil {
				t.Fatalf("unexpected pagination: %+v", data)
			}
			ref := pgtype.Text{String: "sales_order", Valid: true}
			if !reflect.DeepEqual(db.countArgs, []any{ref}) || !reflect.DeepEqual(db.listArgs, []any{ref, tc.wantLimit, tc.wantOffset}) {
				t.Fatalf("count/list filters differ: count=%v list=%v", db.countArgs, db.listArgs)
			}
		})
	}
}

func TestStockMovementReferenceTypes(t *testing.T) {
	for _, types := range [][]string{nil, {"purchase_order", "sales_order"}} {
		uc := NewStockMovementsUseCase()
		uc.SetRepository(repository.New(&stockMovementTestDB{types: types}))
		resp := uc.ListStockMovementReferenceTypes(context.Background())
		if resp.StatusCode != 200 {
			t.Fatalf("unexpected response: %+v", resp)
		}
		got := resp.Data.([]string)
		if got == nil || len(got) != len(types) {
			t.Fatalf("unexpected types: %v", got)
		}
		for i := range types {
			if got[i] != types[i] {
				t.Fatalf("types = %v, want %v", got, types)
			}
		}
	}
}

func TestStockMovementReferenceTypeErrors(t *testing.T) {
	uc := NewStockMovementsUseCase()
	uc.SetRepository(repository.New(&stockMovementTestDB{err: errors.New("database unavailable")}))
	if resp := uc.ListStockMovementReferenceTypes(context.Background()); resp.StatusCode != 500 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp := uc.ListStockMovementsByReferenceType(context.Background(), "sales_order", 1, 100); resp.StatusCode != 500 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp := uc.ListStockMovementsByReferenceType(context.Background(), " ", 1, 100); resp.StatusCode != 400 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
