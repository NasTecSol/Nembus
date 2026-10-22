package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// These rows exercise the generated repository scanners without a live database.
type transferTestRows struct {
	pgx.Rows
	values []any
	index  int
}

func (r *transferTestRows) Next() bool { return r.index < len(r.values) }
func (r *transferTestRows) Close()     {}
func (r *transferTestRows) Err() error { return nil }
func (r *transferTestRows) Scan(dest ...any) error {
	row := reflect.ValueOf(r.values[r.index])
	for i, d := range dest {
		reflect.ValueOf(d).Elem().Set(row.Field(i))
	}
	r.index++
	return nil
}

type transferTestDB struct {
	repository.DBTX
	headers   []any
	items     []any
	itemError error
	calls     int
	t         *testing.T
}

func (db *transferTestDB) Query(_ context.Context, query string, args ...any) (pgx.Rows, error) {
	db.calls++
	switch {
	case strings.Contains(query, "-- name: ListTransferRequestsByOrganization"):
		if !reflect.DeepEqual(args, []any{int32(1), int32(50), int32(0)}) {
			db.t.Fatalf("unexpected list parameters: %v", args)
		}
		return &transferTestRows{values: db.headers}, nil
	case strings.Contains(query, "-- name: ListTransferRequestItemsByRequestIDs"):
		ids := make([]int32, len(db.headers))
		for i, h := range db.headers {
			ids[i] = h.(repository.ListTransferRequestsByOrganizationRow).ID
		}
		if !reflect.DeepEqual(args, []any{ids}) {
			db.t.Fatalf("items must be restricted to the listed transfers: %v", args)
		}
		return &transferTestRows{values: db.items}, db.itemError
	default:
		db.t.Fatalf("unexpected query: %s", query)
		return nil, errors.New("unexpected query")
	}
}

func TestListTransferRequestsIncludesItems(t *testing.T) {
	db := &transferTestDB{
		t: t,
		headers: []any{
			repository.ListTransferRequestsByOrganizationRow{ID: 19, RequestedByName: pgtype.Text{String: "manager", Valid: true}, ApprovedByName: pgtype.Text{String: "approver", Valid: true}},
			repository.ListTransferRequestsByOrganizationRow{ID: 20},
			repository.ListTransferRequestsByOrganizationRow{ID: 21},
		},
		items: []any{
			repository.ListTransferRequestItemsByRequestIDsRow{ID: 1, TransferRequestID: 20, ProductID: 7, ProductName: "Milk", ProductSku: "MILK-001"},
			repository.ListTransferRequestItemsByRequestIDsRow{ID: 2, TransferRequestID: 19, ProductID: 8, ProductName: "Bread"},
			repository.ListTransferRequestItemsByRequestIDsRow{ID: 3, TransferRequestID: 19, ProductID: 9, ProductName: "Butter"},
		},
	}
	uc := NewTransferRequestsUseCase()
	uc.SetRepository(repository.New(db))
	resp := uc.ListTransferRequestsByOrganization(context.Background(), 1, 50, 0)
	if resp.StatusCode != 200 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	list := resp.Data.([]TransferRequestOutput)
	if len(list) != 3 || len(list[0].Items) != 2 || len(list[1].Items) != 1 || len(list[2].Items) != 0 {
		t.Fatalf("items grouped incorrectly: %+v", list)
	}
	if list[0].Items[0].ProductID != 8 || list[1].Items[0].ProductName.String != "Milk" || list[1].Items[0].ProductSKU.String != "MILK-001" {
		t.Fatalf("item details missing or assigned to wrong transfer: %+v", list)
	}
	if list[0].RequestedByName.String != "manager" || list[0].ApprovedByName.String != "approver" {
		t.Fatalf("user names missing: %+v", list[0])
	}
	encoded, err := json.Marshal(list)
	if err != nil {
		t.Fatal(err)
	}
	var payload []map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatal(err)
	}
	if string(payload[2]["items"]) != "[]" {
		t.Fatalf("empty items must serialize as [], got %s", encoded)
	}
	if db.calls != 2 {
		t.Fatalf("expected two batched queries, got %d", db.calls)
	}
}

func TestListTransferRequestsEmpty(t *testing.T) {
	db := &transferTestDB{t: t}
	uc := NewTransferRequestsUseCase()
	uc.SetRepository(repository.New(db))
	resp := uc.ListTransferRequestsByOrganization(context.Background(), 1, 50, 0)
	if resp.StatusCode != 200 || len(resp.Data.([]TransferRequestOutput)) != 0 || db.calls != 1 {
		t.Fatalf("expected empty list without an item query: %+v, calls=%d", resp, db.calls)
	}
}

func TestListTransferRequestsItemQueryError(t *testing.T) {
	db := &transferTestDB{
		t:         t,
		headers:   []any{repository.ListTransferRequestsByOrganizationRow{ID: 19}},
		itemError: errors.New("items unavailable"),
	}
	uc := NewTransferRequestsUseCase()
	uc.SetRepository(repository.New(db))
	resp := uc.ListTransferRequestsByOrganization(context.Background(), 1, 50, 0)
	if resp.StatusCode != 500 || resp.Data != nil {
		t.Fatalf("expected an error instead of incomplete data: %+v", resp)
	}
}
