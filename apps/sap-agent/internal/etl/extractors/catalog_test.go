package extractors

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-sap-agent/internal/db"
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
