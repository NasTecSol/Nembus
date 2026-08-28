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

const inventoryIterationErrorDriverName = "nembus-inventory-iteration-error"

func init() {
	sql.Register(inventoryIterationErrorDriverName, inventoryIterationErrorDriver{})
}

type inventoryIterationErrorDriver struct{}

func (inventoryIterationErrorDriver) Open(string) (driver.Conn, error) {
	return inventoryIterationErrorConn{}, nil
}

type inventoryIterationErrorConn struct{}

func (inventoryIterationErrorConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not expected")
}
func (inventoryIterationErrorConn) Close() error { return nil }
func (inventoryIterationErrorConn) Begin() (driver.Tx, error) {
	return nil, errors.New("begin is not expected")
}
func (inventoryIterationErrorConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return &inventoryIterationErrorRows{returned: false}, nil
}

var _ driver.QueryerContext = inventoryIterationErrorConn{}

type inventoryIterationErrorRows struct {
	returned bool
}

func (*inventoryIterationErrorRows) Columns() []string {
	return []string{"ItemCode", "WhsCode", "OnHand", "IsCommited", "OnOrder", "MinStock", "MaxStock"}
}
func (r *inventoryIterationErrorRows) Close() error { return nil }
func (r *inventoryIterationErrorRows) Next(dest []driver.Value) error {
	if !r.returned {
		r.returned = true
		copy(dest, []driver.Value{"SKU-1", "WH-1", 10.0, 2.0, 3.0, 1.0, 20.0})
		return nil
	}
	return errors.New("OITW iteration failed")
}

func TestExtractInventorySurfacesIterationError(t *testing.T) {
	sqlDB, err := sql.Open(inventoryIterationErrorDriverName, "")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer sqlDB.Close()

	stock, err := NewInventoryExtractor(&db.MSSQLClient{DB: sqlDB}).ExtractInventory(context.Background())
	if err == nil || !strings.Contains(err.Error(), "failed to iterate OITW rows") || !strings.Contains(err.Error(), "OITW iteration failed") {
		t.Fatalf("ExtractInventory() error = %v, want surfaced OITW iteration error", err)
	}
	if stock != nil {
		t.Fatalf("stock = %#v, want nil on source iteration failure", stock)
	}
}
