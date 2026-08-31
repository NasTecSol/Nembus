package extractors

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap/mappings"
)

const priceListQueryErrorDriverName = "nembus-price-list-query-error"

func init() {
	sql.Register(priceListQueryErrorDriverName, priceListQueryErrorDriver{})
}

type priceListQueryErrorDriver struct{}

func (priceListQueryErrorDriver) Open(string) (driver.Conn, error) {
	return priceListQueryErrorConn{}, nil
}

type priceListQueryErrorConn struct{}

func (priceListQueryErrorConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not expected")
}
func (priceListQueryErrorConn) Close() error { return nil }
func (priceListQueryErrorConn) Begin() (driver.Tx, error) {
	return nil, errors.New("begin is not expected")
}

func (priceListQueryErrorConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return nil, errors.New("ITM1 is unavailable")
}

var _ driver.QueryerContext = priceListQueryErrorConn{}

func TestExtractPriceListItemsSurfacesITM1QueryError(t *testing.T) {
	sqlDB, err := sql.Open(priceListQueryErrorDriverName, "")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer sqlDB.Close()

	items, err := NewCatalogExtractor(&db.MSSQLClient{DB: sqlDB}).ExtractPriceListItems(context.Background())
	if err == nil || !strings.Contains(err.Error(), "failed to query ITM1") || !strings.Contains(err.Error(), "ITM1 is unavailable") {
		t.Fatalf("ExtractPriceListItems() error = %v, want surfaced ITM1 query error", err)
	}
	if items != nil {
		t.Fatalf("items = %#v, want nil on source query failure", items)
	}
}

var catalogBrandsTestDriverID atomic.Uint64

type catalogBrandsTestDriver struct {
	rows         [][]driver.Value
	queryErr     error
	iterationErr error
}

func (d catalogBrandsTestDriver) Open(string) (driver.Conn, error) {
	return &catalogBrandsTestConn{driver: d}, nil
}

type catalogBrandsTestConn struct {
	driver catalogBrandsTestDriver
}

func (c *catalogBrandsTestConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not expected")
}
func (c *catalogBrandsTestConn) Close() error { return nil }
func (c *catalogBrandsTestConn) Begin() (driver.Tx, error) {
	return nil, errors.New("begin is not expected")
}
func (c *catalogBrandsTestConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	if c.driver.queryErr != nil {
		return nil, c.driver.queryErr
	}
	return &catalogBrandsTestRows{
		rows:         c.driver.rows,
		iterationErr: c.driver.iterationErr,
	}, nil
}

var _ driver.QueryerContext = (*catalogBrandsTestConn)(nil)

type catalogBrandsTestRows struct {
	rows          [][]driver.Value
	index         int
	iterationErr  error
	iterationDone bool
}

func (*catalogBrandsTestRows) Columns() []string {
	return []string{"FirmCode", "FirmName"}
}
func (*catalogBrandsTestRows) Close() error { return nil }
func (r *catalogBrandsTestRows) Next(dest []driver.Value) error {
	if r.index < len(r.rows) {
		copy(dest, r.rows[r.index])
		r.index++
		return nil
	}
	if r.iterationErr != nil && !r.iterationDone {
		r.iterationDone = true
		return r.iterationErr
	}
	return io.EOF
}

func openCatalogBrandsTestDB(t *testing.T, testDriver catalogBrandsTestDriver) *sql.DB {
	t.Helper()
	driverName := fmt.Sprintf("nembus-catalog-brands-%d", catalogBrandsTestDriverID.Add(1))
	sql.Register(driverName, testDriver)
	sqlDB, err := sql.Open(driverName, "")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return sqlDB
}

func extractBrandsFromRows(t *testing.T, rows ...[]driver.Value) ([]mappings.CanonicalBrand, error) {
	t.Helper()
	sqlDB := openCatalogBrandsTestDB(t, catalogBrandsTestDriver{rows: rows})
	return NewCatalogExtractor(&db.MSSQLClient{DB: sqlDB}).ExtractBrands(context.Background())
}

func TestExtractBrandsExcludesNegativeFirmCodePlaceholder(t *testing.T) {
	brands, err := extractBrandsFromRows(t, []driver.Value{-1, "- No Manufacturer -"})
	if err != nil {
		t.Fatalf("ExtractBrands() error = %v", err)
	}
	if len(brands) != 0 {
		t.Fatalf("brands = %#v, want zero canonical brands", brands)
	}
}

func TestExtractBrandsExcludesZeroFirmCode(t *testing.T) {
	brands, err := extractBrandsFromRows(t, []driver.Value{0, "Manufacturer"})
	if err != nil {
		t.Fatalf("ExtractBrands() error = %v", err)
	}
	if len(brands) != 0 {
		t.Fatalf("brands = %#v, want zero canonical brands", brands)
	}
}

func TestExtractBrandsPreservesPositiveFirmCodeMapping(t *testing.T) {
	brands, err := extractBrandsFromRows(t, []driver.Value{42, "  Acme  "})
	if err != nil {
		t.Fatalf("ExtractBrands() error = %v", err)
	}
	if len(brands) != 1 {
		t.Fatalf("len(brands) = %d, want 1", len(brands))
	}
	if brands[0].Code != "BRD-42" || brands[0].Name != "Acme" || !brands[0].IsActive {
		t.Fatalf("brand = %#v, want existing canonical mapping for FirmCode 42", brands[0])
	}
}

func TestExtractBrandsExcludesPositiveFirmCodeWithBlankName(t *testing.T) {
	brands, err := extractBrandsFromRows(t, []driver.Value{42, " \t\n "})
	if err != nil {
		t.Fatalf("ExtractBrands() error = %v", err)
	}
	if len(brands) != 0 {
		t.Fatalf("brands = %#v, want zero canonical brands", brands)
	}
}

func TestExtractBrandsReturnsOnlyRealBrandsFromMixedRows(t *testing.T) {
	brands, err := extractBrandsFromRows(t,
		[]driver.Value{-1, "- No Manufacturer -"},
		[]driver.Value{0, "Unknown"},
		[]driver.Value{7, "  Acme  "},
	)
	if err != nil {
		t.Fatalf("ExtractBrands() error = %v", err)
	}
	if len(brands) != 1 || brands[0].Code != "BRD-7" || brands[0].Name != "Acme" {
		t.Fatalf("brands = %#v, want only canonical brand BRD-7 / Acme", brands)
	}
}

func TestExtractBrandsSurfacesQueryError(t *testing.T) {
	sqlDB := openCatalogBrandsTestDB(t, catalogBrandsTestDriver{queryErr: errors.New("OMRC is unavailable")})
	brands, err := NewCatalogExtractor(&db.MSSQLClient{DB: sqlDB}).ExtractBrands(context.Background())
	if err == nil || !strings.Contains(err.Error(), "failed to query OMRC") || !strings.Contains(err.Error(), "OMRC is unavailable") {
		t.Fatalf("ExtractBrands() error = %v, want surfaced OMRC query error", err)
	}
	if brands != nil {
		t.Fatalf("brands = %#v, want nil on source query failure", brands)
	}
}

func TestExtractBrandsSurfacesScanError(t *testing.T) {
	brands, err := extractBrandsFromRows(t, []driver.Value{"not-an-int", "Acme"})
	if err == nil || !strings.Contains(err.Error(), "failed to scan OMRC row") {
		t.Fatalf("ExtractBrands() error = %v, want surfaced OMRC scan error", err)
	}
	if brands != nil {
		t.Fatalf("brands = %#v, want nil on source scan failure", brands)
	}
}

func TestExtractBrandsSurfacesIterationError(t *testing.T) {
	sqlDB := openCatalogBrandsTestDB(t, catalogBrandsTestDriver{
		rows:         [][]driver.Value{{7, "Acme"}},
		iterationErr: errors.New("OMRC iteration failed"),
	})
	brands, err := NewCatalogExtractor(&db.MSSQLClient{DB: sqlDB}).ExtractBrands(context.Background())
	if err == nil || !strings.Contains(err.Error(), "failed to iterate OMRC rows") || !strings.Contains(err.Error(), "OMRC iteration failed") {
		t.Fatalf("ExtractBrands() error = %v, want surfaced OMRC iteration error", err)
	}
	if brands != nil {
		t.Fatalf("brands = %#v, want nil on source iteration failure", brands)
	}
}
