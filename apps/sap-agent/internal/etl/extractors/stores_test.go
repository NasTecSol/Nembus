package extractors

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-sap-agent/internal/db"
)

const storesExtractorTestDriverName = "nembus-stores-extractor-test"

func init() {
	sql.Register(storesExtractorTestDriverName, storesExtractorTestDriver{})
}

type storesExtractorTestDriver struct{}

func (storesExtractorTestDriver) Open(name string) (driver.Conn, error) {
	return storesExtractorTestConn{scenario: name}, nil
}

type storesExtractorTestConn struct {
	scenario string
}

func (storesExtractorTestConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not expected")
}
func (storesExtractorTestConn) Close() error { return nil }
func (storesExtractorTestConn) Begin() (driver.Tx, error) {
	return nil, errors.New("begin is not expected")
}

func (c storesExtractorTestConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if strings.Contains(query, "FROM OWHS") {
		if c.scenario == "owhs-query-error" {
			return nil, errors.New("OWHS is unavailable")
		}
		rows := [][]driver.Value{
			{"WH-1", "Warehouse 1", "N", "", "", "", ""},
		}
		if c.scenario == "multiple-warehouses" {
			rows = append(rows, []driver.Value{"WH-2", "Warehouse 2", "N", "", "", "", ""})
		}
		var iterationErr error
		if c.scenario == "owhs-iteration-error" {
			iterationErr = errors.New("OWHS iteration failed")
		}
		return &storesExtractorTestRows{
			columns: []string{"WhsCode", "WhsName", "Locked", "Street", "City", "Country", "ZipCode"},
			values:  rows,
			err:     iterationErr,
		}, nil
	}

	if strings.Contains(query, "FROM OBIN") {
		if c.scenario == "obin-query-error" {
			return nil, errors.New("OBIN is unavailable")
		}
		var rows [][]driver.Value
		switch c.scenario {
		case "existing-main":
			rows = [][]driver.Value{{int64(1), "MAIN", "WH-1", "Main bin", "N"}}
		case "ordinary-bins":
			rows = [][]driver.Value{
				{int64(1), "A-01", "WH-1", "Bin A", "N"},
				{int64(2), "B-01", "WH-1", "Bin B", "N"},
			}
		case "multiple-warehouses":
			rows = [][]driver.Value{{int64(1), "A-01", "WH-1", "Bin A", "N"}}
		}
		return &storesExtractorTestRows{
			columns: []string{"AbsEntry", "BinCode", "WhsCode", "Descr", "Disabled"},
			values:  rows,
		}, nil
	}

	return nil, fmt.Errorf("unexpected query: %s", query)
}

var _ driver.QueryerContext = storesExtractorTestConn{}

type storesExtractorTestRows struct {
	columns []string
	values  [][]driver.Value
	index   int
	err     error
}

func (r *storesExtractorTestRows) Columns() []string { return r.columns }
func (r *storesExtractorTestRows) Close() error      { return nil }
func (r *storesExtractorTestRows) Next(dest []driver.Value) error {
	if r.index < len(r.values) {
		copy(dest, r.values[r.index])
		r.index++
		return nil
	}
	if r.err != nil {
		err := r.err
		r.err = nil
		return err
	}
	return io.EOF
}

func extractStoresForTest(t *testing.T, scenario string) ([]string, []struct {
	storeCode string
	code      string
	name      string
	typeName  string
	active    bool
}) {
	t.Helper()
	sqlDB, err := sql.Open(storesExtractorTestDriverName, scenario)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer sqlDB.Close()

	stores, locations, err := NewStoresExtractor(&db.MSSQLClient{DB: sqlDB}).ExtractStores(context.Background())
	if err != nil {
		t.Fatalf("ExtractStores() error = %v", err)
	}
	storeCodes := make([]string, 0, len(stores))
	for _, store := range stores {
		storeCodes = append(storeCodes, store.Code)
	}
	canonicalLocations := make([]struct {
		storeCode string
		code      string
		name      string
		typeName  string
		active    bool
	}, 0, len(locations))
	for _, location := range locations {
		canonicalLocations = append(canonicalLocations, struct {
			storeCode string
			code      string
			name      string
			typeName  string
			active    bool
		}{location.StoreCode, location.Code, location.Name, location.LocationType, location.IsActive})
	}
	return storeCodes, canonicalLocations
}

func TestExtractStoresAddsMainWhenOBINIsEmpty(t *testing.T) {
	stores, locations := extractStoresForTest(t, "no-obin")
	if len(stores) != 1 || len(locations) != 1 {
		t.Fatalf("stores = %v, locations = %#v, want one store and one MAIN location", stores, locations)
	}
	assertMainLocation(t, locations[0], "WH-1")
}

func TestExtractStoresAddsMainForEachWarehouse(t *testing.T) {
	stores, locations := extractStoresForTest(t, "multiple-warehouses")
	if len(stores) != 2 || len(locations) != 3 {
		t.Fatalf("stores = %v, locations = %#v, want two stores, one OBIN bin, and two MAIN locations", stores, locations)
	}
	for _, storeCode := range []string{"WH-1", "WH-2"} {
		found := false
		for _, location := range locations {
			if location.storeCode == storeCode && location.code == "MAIN" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("locations = %#v, want MAIN associated with %s", locations, storeCode)
		}
	}
}

func TestExtractStoresDoesNotDuplicateExistingMainBin(t *testing.T) {
	_, locations := extractStoresForTest(t, "existing-main")
	if len(locations) != 1 {
		t.Fatalf("locations = %#v, want one existing MAIN location", locations)
	}
	assertMainLocation(t, locations[0], "WH-1")
	if locations[0].name != "Main bin" {
		t.Fatalf("MAIN name = %q, want existing OBIN description", locations[0].name)
	}
}

func TestExtractStoresPreservesOrdinaryBins(t *testing.T) {
	_, locations := extractStoresForTest(t, "ordinary-bins")
	if len(locations) != 3 {
		t.Fatalf("locations = %#v, want two OBIN bins plus MAIN", locations)
	}
	seen := make(map[string]bool, len(locations))
	for _, location := range locations {
		seen[location.code] = true
	}
	for _, code := range []string{"A-01", "B-01", "MAIN"} {
		if !seen[code] {
			t.Errorf("locations = %#v, missing %s", locations, code)
		}
	}
}

func TestExtractStoresFailsClosedOnOWHSQueryError(t *testing.T) {
	sqlDB, err := sql.Open(storesExtractorTestDriverName, "owhs-query-error")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer sqlDB.Close()

	stores, locations, err := NewStoresExtractor(&db.MSSQLClient{DB: sqlDB}).ExtractStores(context.Background())
	if err == nil || !strings.Contains(err.Error(), "failed to query OWHS") {
		t.Fatalf("ExtractStores() error = %v, want OWHS query error", err)
	}
	if stores != nil || locations != nil {
		t.Fatalf("stores = %v, locations = %#v, want nil results on OWHS query failure", stores, locations)
	}
}

func TestExtractStoresFailsClosedOnOWHSIterationError(t *testing.T) {
	sqlDB, err := sql.Open(storesExtractorTestDriverName, "owhs-iteration-error")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer sqlDB.Close()

	stores, locations, err := NewStoresExtractor(&db.MSSQLClient{DB: sqlDB}).ExtractStores(context.Background())
	if err == nil || !strings.Contains(err.Error(), "failed to iterate OWHS rows") {
		t.Fatalf("ExtractStores() error = %v, want OWHS iteration error", err)
	}
	if stores != nil || locations != nil {
		t.Fatalf("stores = %v, locations = %#v, want nil results on OWHS iteration failure", stores, locations)
	}
}

func assertMainLocation(t *testing.T, location struct {
	storeCode string
	code      string
	name      string
	typeName  string
	active    bool
}, storeCode string) {
	t.Helper()
	if location.storeCode != storeCode || location.code != "MAIN" || location.name == "" || location.typeName != "standard" || !location.active {
		t.Fatalf("location = %#v, want active standard MAIN for store %s", location, storeCode)
	}
}
