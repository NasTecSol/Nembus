package usecase

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// workflowMockDBTX intercepts SQL queries executed by repository.Queries
// to verify inventory adjustments, stock movements, and sales_orders_v2 linkage.
type workflowMockDBTX struct {
	organizationMetadata json.RawMessage
	order                repository.SalesOrdersV2
	orderLines           []repository.ListSalesOrderLinesV2Row
	commits, rollbacks   int
	legacyTransactionID  int32

	// Recorded AdjustInventoryStock calls
	adjustedStocks []repository.AdjustInventoryStockByProductAndStoreParams
	// Recorded CreateStockMovement calls
	stockMovements []repository.CreateStockMovementParams
	// Recorded sales order v2 created
	createdOrdersV2 []repository.CreateSalesOrderV2Params
	// Recorded sales order lines v2 created
	createdOrderLinesV2 []repository.CreateSalesOrderLineV2Params
	// Recorded links between pos_transactions and sales_orders_v2
	linkedOrders []repository.LinkPosTransactionToSalesOrderParams
	// Recorded ZATCA chain entries
	chainEntries []repository.InsertChainEntryParams

	// Recorded metadata updates
	updatedPosMetadata    []repository.UpdatePosTransactionMetadataParams
	updatedReturnMetadata []repository.UpdateSalesReturnMetadataParams

	// Keypair for device signing
	privateKeyPEM       string
	certificate         string
	createdTransactions int
	latestICV           int64
}

type mockDashboardRows struct {
	orders []repository.ListSalesOrdersDashboardRow
	index  int
}

func (m *mockDashboardRows) Close()                                       {}
func (m *mockDashboardRows) Err() error                                   { return nil }
func (m *mockDashboardRows) CommandTag() pgconn.CommandTag                { return pgconn.NewCommandTag("SELECT 1") }
func (m *mockDashboardRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *mockDashboardRows) Next() bool {
	if m.index < len(m.orders) {
		m.index++
		return true
	}
	return false
}
func (m *mockDashboardRows) Scan(dest ...interface{}) error {
	row := m.orders[m.index-1]
	if len(dest) >= 10 {
		if idPtr, ok := dest[0].(*uuid.UUID); ok {
			*idPtr = row.ID
		}
		if numPtr, ok := dest[1].(*string); ok {
			*numPtr = row.OrderNumber
		}
		if datePtr, ok := dest[2].(*pgtype.Timestamp); ok {
			*datePtr = row.OrderDate
		}
		if statusPtr, ok := dest[3].(*string); ok {
			*statusPtr = row.Status
		}
		if totPtr, ok := dest[4].(*pgtype.Numeric); ok {
			*totPtr = row.TotalAmount
		}
		if chPtr, ok := dest[5].(*pgtype.Text); ok {
			*chPtr = row.SalesChannel
		}
		if cNamePtr, ok := dest[6].(*string); ok {
			*cNamePtr = row.CustomerName
		}
		if sNamePtr, ok := dest[7].(*string); ok {
			*sNamePtr = row.StoreName
		}
		if lCountPtr, ok := dest[8].(*int64); ok {
			*lCountPtr = row.LineCount
		}
		if itemsPtr, ok := dest[9].(*pgtype.Numeric); ok {
			*itemsPtr = row.TotalItems
		}
	}
	return nil
}
func (m *mockDashboardRows) Values() ([]interface{}, error) { return nil, nil }
func (m *mockDashboardRows) RawValues() [][]byte            { return nil }
func (m *mockDashboardRows) Conn() *pgx.Conn                { return nil }

func (m *workflowMockDBTX) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	if strings.Contains(sql, "UPDATE pos_transactions") && strings.Contains(sql, "sales_order_id") {
		if len(args) >= 2 {
			id := args[0].(int32)
			soID := args[1].(pgtype.UUID)
			m.linkedOrders = append(m.linkedOrders, repository.LinkPosTransactionToSalesOrderParams{
				ID:           id,
				SalesOrderID: soID,
			})
		}
	}
	if strings.Contains(sql, "UPDATE pos_transactions") && strings.Contains(sql, "metadata") {
		if len(args) >= 2 {
			id := args[0].(int32)
			meta, _ := args[1].(json.RawMessage)
			m.updatedPosMetadata = append(m.updatedPosMetadata, repository.UpdatePosTransactionMetadataParams{
				ID:       id,
				Metadata: meta,
			})
		}
	}
	if strings.Contains(sql, "UPDATE sales_returns") && strings.Contains(sql, "metadata") {
		if len(args) >= 2 {
			id := args[0].(int32)
			meta, _ := args[1].(json.RawMessage)
			m.updatedReturnMetadata = append(m.updatedReturnMetadata, repository.UpdateSalesReturnMetadataParams{
				ID:       id,
				Metadata: meta,
			})
		}
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (m *workflowMockDBTX) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if strings.Contains(sql, "ListSalesOrderLinesV2") {
		return &zatcaOrderRows{values: m.orderLines}, nil
	}

	if strings.Contains(sql, "ListSalesOrdersDashboard") || strings.Contains(sql, "FROM sales_orders_v2") {
		rows := make([]repository.ListSalesOrdersDashboardRow, 0, len(m.createdOrdersV2))
		for _, o := range m.createdOrdersV2 {
			rows = append(rows, repository.ListSalesOrdersDashboardRow{
				OrderNumber:  o.OrderNumber,
				SalesChannel: o.SalesChannel,
				Status:       "fulfilled",
			})
		}
		return &mockDashboardRows{orders: rows}, nil
	}
	return nil, nil
}

func (m *workflowMockDBTX) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return &workflowMockRow{mock: m, sql: sql, args: args}
}

func (m *workflowMockDBTX) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

type workflowMockRow struct {
	mock *workflowMockDBTX
	sql  string
	args []interface{}
}

func (r *workflowMockRow) Scan(dest ...interface{}) error {
	s := r.sql
	switch {
	case strings.Contains(s, "CountActiveZatcaDevices"):
		*dest[0].(*int64) = 1
		return nil
	case strings.Contains(s, "GetSalesOrderV2"):
		return scanOrderTestValue(r.mock.order, dest...)
	case strings.Contains(s, "GetChainEntryByEntity"):
		for _, c := range r.mock.chainEntries {
			if c.EntityType == r.args[0].(string) && c.EntityID == r.args[1].(string) {
				return scanOrderTestValue(orderTestChain(c), dest...)
			}
		}
		return pgx.ErrNoRows
	case strings.Contains(s, "GetLegacyZatcaOrderTransaction"):
		if r.mock.legacyTransactionID > 0 {
			*dest[0].(*int32) = r.mock.legacyTransactionID
			return nil
		}
		return pgx.ErrNoRows
	case strings.Contains(s, "GetStore") || strings.Contains(s, "FROM stores"):
		if len(dest) >= 3 {
			// Scan into repository.Store
			if idPtr, ok := dest[0].(*int32); ok {
				*idPtr = 1
			}
			if orgPtr, ok := dest[1].(*int32); ok {
				*orgPtr = 10
			}
			if namePtr, ok := dest[2].(*string); ok {
				*namePtr = "Downtown Store"
			}
		}
		return nil

	case strings.Contains(s, "GetOrganization") || strings.Contains(s, "FROM organizations"):
		if len(dest) >= 4 {
			// Scan into repository.Organization
			if idPtr, ok := dest[0].(*int32); ok {
				*idPtr = 10
			}
			if namePtr, ok := dest[1].(*string); ok {
				*namePtr = "Nembus Retail LLC"
			}
			if codePtr, ok := dest[2].(*string); ok {
				*codePtr = "ORG-10"
			}
			if taxIDPtr, ok := dest[4].(*pgtype.Text); ok {
				*taxIDPtr = pgtype.Text{String: "310123456700003", Valid: true}
			}
		}
		if len(dest) > 8 {
			metadata := r.mock.organizationMetadata
			if metadata == nil {
				metadata = json.RawMessage(testZatcaSellerMetadata)
			}
			*dest[8].(*json.RawMessage) = metadata
		}
		return nil

	case strings.Contains(s, "GetActivePosDevice") || strings.Contains(s, "GetActiveCloudDevice"):
		if len(dest) >= 8 {
			if idPtr, ok := dest[0].(*int32); ok {
				*idPtr = 1
			}
			if orgPtr, ok := dest[1].(*int32); ok {
				*orgPtr = 10
			}
			if pkPtr, ok := dest[7].(*string); ok {
				*pkPtr = r.mock.privateKeyPEM
			}
		}
		if len(dest) > 9 {
			*dest[8].(*pgtype.Text) = pgtype.Text{String: r.mock.certificate, Valid: true}
			*dest[11].(*pgtype.Text) = pgtype.Text{String: "sandbox", Valid: true}
			*dest[12].(*pgtype.Bool) = pgtype.Bool{Bool: true, Valid: true}
		}
		return nil

	case strings.Contains(s, "GetNextICV"):
		if len(dest) > 0 {
			if icvPtr, ok := dest[0].(*int32); ok {
				*icvPtr = int32(r.mock.latestICV + 1)
			} else if icvPtr, ok := dest[0].(*int64); ok {
				*icvPtr = r.mock.latestICV + 1
			}
		}
		return nil

	case strings.Contains(s, "::numeric") || strings.Contains(s, "SELECT $1"):
		if len(dest) > 0 && len(r.args) > 0 {
			if numPtr, ok := dest[0].(*pgtype.Numeric); ok {
				if str, ok := r.args[0].(string); ok {
					_ = numPtr.Scan(str)
				}
			}
		}
		return nil

	case strings.Contains(s, "GetLatestChainEntry"):
		if r.mock.latestICV > 0 {
			if len(dest) >= 8 {
				if icvPtr, ok := dest[6].(*int64); ok {
					*icvPtr = r.mock.latestICV
				}
				if pihPtr, ok := dest[7].(*string); ok {
					*pihPtr = "previous-hash"
				}
			}
			return nil
		}
		return pgx.ErrNoRows

	case strings.Contains(s, "InsertChainEntry") || strings.Contains(s, "INSERT INTO zatca_document_chain"):
		if len(r.args) >= 9 {
			r.mock.chainEntries = append(r.mock.chainEntries, repository.InsertChainEntryParams{
				EntityType:     r.args[0].(string),
				EntityID:       r.args[1].(string),
				DeviceConfigID: r.args[2].(int32),
				OrganizationID: r.args[3].(int32),
				ZatcaUuid:      r.args[4].(uuid.UUID),
				Icv:            r.args[5].(int64),
				Pih:            r.args[6].(string),
				XmlHash:        r.args[7].(string),
				QrCodeBase64:   r.args[9].(pgtype.Text), SignedXml: r.args[10].(pgtype.Text), ZatcaStatus: r.args[8].(repository.NullZatcaDocStatus),
			})
		}
		return scanOrderTestValue(orderTestChain(r.mock.chainEntries[len(r.mock.chainEntries)-1]), dest...)

	case strings.Contains(s, "CreatePosTransaction :one") || strings.Contains(s, "INSERT INTO pos_transactions ("):
		r.mock.createdTransactions++
		if len(dest) >= 7 {
			if idPtr, ok := dest[0].(*int32); ok {
				*idPtr = 77
			}
			if numPtr, ok := dest[6].(*string); ok {
				*numPtr = "POS-TXN-0077"
			}
			if totalPtr, ok := dest[12].(*pgtype.Numeric); ok {
				_ = totalPtr.Scan("115.00")
			}
		}
		if len(dest) > 12 && len(r.args) > 12 {
			*dest[1].(*int32) = r.args[1].(int32)
			*dest[9].(*pgtype.Numeric) = r.args[9].(pgtype.Numeric)
			*dest[10].(*pgtype.Numeric) = r.args[11].(pgtype.Numeric)
			*dest[11].(*pgtype.Numeric) = r.args[10].(pgtype.Numeric)
			*dest[12].(*pgtype.Numeric) = r.args[12].(pgtype.Numeric)
		}
		return nil

	case strings.Contains(s, "CreatePosTransactionLine") || strings.Contains(s, "INSERT INTO pos_transaction_lines"):
		if len(dest) >= 7 {
			if idPtr, ok := dest[0].(*int32); ok {
				*idPtr = 101
			}
			if txIDPtr, ok := dest[1].(*int32); ok {
				*txIDPtr = 77
			}
			if pIDPtr, ok := dest[2].(*int32); ok {
				if len(r.args) >= 3 {
					*pIDPtr = r.args[2].(int32)
				}
			}
			if qtyPtr, ok := dest[6].(*pgtype.Numeric); ok {
				if len(r.args) >= 7 {
					*qtyPtr = r.args[6].(pgtype.Numeric)
				}
			}
		}
		if len(dest) > 12 && len(r.args) > 13 {
			for _, pair := range [][2]int{{4, 6}, {6, 8}, {7, 9}, {8, 10}, {9, 11}, {10, 12}, {11, 13}} {
				*dest[pair[0]].(*pgtype.Numeric) = r.args[pair[1]].(pgtype.Numeric)
			}
			*dest[12].(*pgtype.Int4) = r.args[1].(pgtype.Int4)
		}
		return nil

	case strings.Contains(s, "AdjustInventoryStockByProductAndStore"):
		if len(r.args) >= 5 {
			r.mock.adjustedStocks = append(r.mock.adjustedStocks, repository.AdjustInventoryStockByProductAndStoreParams{
				ProductID:         r.args[0].(int32),
				ProductVariantID:  r.args[1].(pgtype.Int4),
				StoreID:           r.args[2].(int32),
				QuantityOnHand:    r.args[3].(pgtype.Numeric),
				QuantityAvailable: r.args[4].(pgtype.Numeric),
			})
		}
		return nil

	case strings.Contains(s, "CreateStockMovement") || strings.Contains(s, "INSERT INTO stock_movements"):
		var costPerUnit, totalVal pgtype.Numeric
		if len(r.args) >= 18 {
			if cpu, ok := r.args[16].(pgtype.Numeric); ok {
				costPerUnit = cpu
			}
			if tv, ok := r.args[17].(pgtype.Numeric); ok {
				totalVal = tv
			}
		}
		if len(r.args) >= 10 {
			r.mock.stockMovements = append(r.mock.stockMovements, repository.CreateStockMovementParams{
				MovementType:  r.args[0].(string),
				ReferenceType: r.args[1].(pgtype.Text),
				ReferenceID:   r.args[2].(pgtype.Int4),
				ProductID:     r.args[3].(int32),
				Quantity:      r.args[9].(pgtype.Numeric),
				CostPerUnit:   costPerUnit,
				TotalValue:    totalVal,
			})
		}
		return nil

	case strings.Contains(s, "CreateSalesOrderV2") || strings.Contains(s, "INSERT INTO sales_orders_v2"):
		var shipAddr, billAddr json.RawMessage
		if len(r.args) >= 21 {
			if sa, ok := r.args[19].(json.RawMessage); ok {
				shipAddr = sa
			}
			if ba, ok := r.args[20].(json.RawMessage); ok {
				billAddr = ba
			}
		}
		if len(r.args) >= 12 {
			r.mock.createdOrdersV2 = append(r.mock.createdOrdersV2, repository.CreateSalesOrderV2Params{
				OrderNumber:     r.args[0].(string),
				SalesChannel:    r.args[11].(pgtype.Text),
				ShippingAddress: shipAddr,
				BillingAddress:  billAddr,
			})
		}
		r.mock.order = repository.SalesOrdersV2{ID: uuid.New(), OrderNumber: r.args[0].(string), OrganizationID: r.args[1].(int32), StoreID: r.args[2].(pgtype.Int4), PosTerminalID: r.args[26].(pgtype.Int4)}
		return scanOrderTestValue(r.mock.order, dest...)

	case strings.Contains(s, "UpdateOrderTotals") || strings.Contains(s, "UpdateSalesOrderTotals"):
		r.mock.order.Subtotal = r.args[1].(pgtype.Numeric)
		r.mock.order.DiscountAmount = r.args[2].(pgtype.Numeric)
		r.mock.order.TaxAmount = r.args[3].(pgtype.Numeric)
		r.mock.order.TotalAmount = r.args[6].(pgtype.Numeric)
		return scanOrderTestValue(r.mock.order, dest...)

	case strings.Contains(s, "GetProduct") || strings.Contains(s, "FROM products"):
		if len(dest) >= 18 {
			if idPtr, ok := dest[0].(*int32); ok {
				*idPtr = 401
			}
			if costPtr, ok := dest[17].(*pgtype.Numeric); ok {
				_ = costPtr.Scan("35.00")
			}
		}
		return nil

	case strings.Contains(s, "GetPosTransactionLineByID") || strings.Contains(s, "FROM pos_transaction_lines"):
		if len(dest) >= 5 {
			if idPtr, ok := dest[0].(*int32); ok {
				*idPtr = 101
			}
			if qtyPtr, ok := dest[4].(*pgtype.Numeric); ok {
				_ = qtyPtr.Scan("5.000")
			}
		}
		return nil

	case strings.Contains(s, "GetReturnedQuantityByLineID"):
		if len(dest) > 0 {
			if totPtr, ok := dest[0].(*pgtype.Numeric); ok {
				_ = totPtr.Scan("1.000")
			}
		}
		return nil

	case strings.Contains(s, "CreateSalesOrderLineV2"):
		r.mock.orderLines = append(r.mock.orderLines, repository.ListSalesOrderLinesV2Row{SalesOrderID: r.args[0].(uuid.UUID), OrganizationID: r.args[1].(int32), LineNumber: r.args[2].(int32), ProductName: r.args[5].(string), QuantityOrdered: r.args[7].(pgtype.Numeric), UnitPrice: r.args[9].(pgtype.Numeric), DiscountAmount: r.args[10].(pgtype.Numeric), TaxAmount: r.args[12].(pgtype.Numeric), LineTotal: r.args[13].(pgtype.Numeric), TaxRate: r.args[15].(pgtype.Numeric)})
		if len(r.args) >= 8 {
			r.mock.createdOrderLinesV2 = append(r.mock.createdOrderLinesV2, repository.CreateSalesOrderLineV2Params{
				SalesOrderID:    r.args[0].(uuid.UUID),
				ProductID:       r.args[3].(int32),
				QuantityOrdered: r.args[7].(pgtype.Numeric),
			})
		}
		return nil

	case strings.Contains(s, "CreateSalesReturn :one") || strings.Contains(s, "INSERT INTO sales_returns ("):
		if len(dest) >= 8 {
			if idPtr, ok := dest[0].(*int32); ok {
				*idPtr = 55
			}
			if numPtr, ok := dest[1].(*string); ok {
				*numPtr = "RET-2026-0055"
			}
			if subPtr, ok := dest[6].(*pgtype.Numeric); ok {
				_ = subPtr.Scan("100.00")
			}
			if taxPtr, ok := dest[7].(*pgtype.Numeric); ok {
				_ = taxPtr.Scan("15.00")
			}
		}
		if len(dest) > 12 && len(r.args) > 11 {
			*dest[2].(*int32) = r.args[1].(int32)
			for _, pair := range [][2]int{{10, 9}, {11, 10}, {12, 11}} {
				*dest[pair[0]].(*pgtype.Numeric) = r.args[pair[1]].(pgtype.Numeric)
			}
		}
		return nil

	case strings.Contains(s, "CreateSalesReturnLine") || strings.Contains(s, "INSERT INTO sales_return_lines"):
		if len(dest) >= 5 {
			if idPtr, ok := dest[0].(*int32); ok {
				*idPtr = 201
			}
			if retIDPtr, ok := dest[1].(*int32); ok {
				*retIDPtr = 55
			}
			if pIDPtr, ok := dest[2].(*int32); ok {
				if len(r.args) >= 2 {
					*pIDPtr = r.args[1].(int32)
				}
			}
		}
		if len(dest) > 7 && len(r.args) > 6 {
			for _, pair := range [][2]int{{5, 4}, {6, 5}, {7, 6}} {
				*dest[pair[0]].(*pgtype.Numeric) = r.args[pair[1]].(pgtype.Numeric)
			}
		}
		return nil

	case strings.Contains(s, "GetPosTransaction"):
		if len(dest) >= 7 {
			if idPtr, ok := dest[0].(*int32); ok {
				*idPtr = 77
			}
			if numPtr, ok := dest[6].(*string); ok {
				*numPtr = "POS-TXN-0077"
			}
			if totPtr, ok := dest[12].(*pgtype.Numeric); ok {
				_ = totPtr.Scan("115.00")
			}
		}
		return nil

	case strings.Contains(s, "UpdatePOSTransactionStatus"):
		return nil

	case strings.Contains(s, "UpdateSessionExpectedBalance"):
		return nil
	}
	return nil
}

// TestPosTransactionInventoryDeduction verifies that CreateTransaction:
// 1. Decrements inventory_stock by negative quantity
// 2. Logs stock_movements with movement_type 'pos_sale' and reference_type 'pos_transaction'
func TestPosTransactionInventoryDeduction(t *testing.T) {
	mockDB := &workflowMockDBTX{}
	repo := repository.New(mockDB)

	posUC := NewPosUseCase()
	posUC.SetRepository(repo)

	var qty1, qty2 pgtype.Numeric
	_ = qty1.Scan("3.000")
	_ = qty2.Scan("1.500")

	input := &PosCreateTransactionInput{
		StoreID:           1,
		CashierID:         10,
		CashierSessionID:  88,
		PosTerminalID:     2,
		TransactionNumber: "POS-TXN-0077",
		Lines: []PosCreateTransactionLineInput{
			{
				ProductID: 201,
				Quantity:  qty1,
				UnitPrice: pgtype.Numeric{},
				LineTotal: pgtype.Numeric{},
			},
			{
				ProductID: 202,
				Quantity:  qty2,
				UnitPrice: pgtype.Numeric{},
				LineTotal: pgtype.Numeric{},
			},
		},
	}

	resp := posUC.CreateTransaction(context.Background(), input)
	if resp.StatusCode != 201 {
		t.Fatalf("CreateTransaction failed with code %d: %s", resp.StatusCode, resp.Message)
	}

	// 1. Verify inventory stock was decremented for both products
	if len(mockDB.adjustedStocks) != 2 {
		t.Fatalf("Expected 2 AdjustInventoryStock calls, got %d", len(mockDB.adjustedStocks))
	}

	// Verify product 201 adjustment was negative (-3)
	adj1 := mockDB.adjustedStocks[0]
	if adj1.ProductID != 201 {
		t.Errorf("Expected product 201, got %d", adj1.ProductID)
	}
	val1, err := adj1.QuantityOnHand.Value()
	if err != nil || val1 == nil {
		t.Fatalf("Failed to read QuantityOnHand: %v", err)
	}
	if !strings.HasPrefix(val1.(string), "-3") {
		t.Errorf("Expected negative quantity -3, got %v", val1)
	}

	// 2. Verify stock movements were created with movement_type = 'pos_sale'
	if len(mockDB.stockMovements) != 2 {
		t.Fatalf("Expected 2 stock movements, got %d", len(mockDB.stockMovements))
	}

	mov1 := mockDB.stockMovements[0]
	if mov1.MovementType != "pos_sale" {
		t.Errorf("Expected movement_type 'pos_sale', got '%s'", mov1.MovementType)
	}
	if mov1.ReferenceType.String != "pos_transaction" {
		t.Errorf("Expected reference_type 'pos_transaction', got '%s'", mov1.ReferenceType.String)
	}
	if mov1.ReferenceID.Int32 != 77 {
		t.Errorf("Expected reference_id 77, got %d", mov1.ReferenceID.Int32)
	}
}

// TestPosTransactionZatcaSigningAndChaining verifies that when ZATCA is enabled:
// 1. SignPosTransaction is called during CreateTransaction
// 2. Cryptographic TLV QR code is populated in response and metadata
// 3. Document is chained in zatca_document_chain
func TestPosTransactionZatcaSigningAndChaining(t *testing.T) {
	key, err := GenerateECDSAKeyPair()
	if err != nil {
		t.Fatalf("Key generation failed: %v", err)
	}
	pemStr, err := EncodePrivateKeyPEM(key)
	if err != nil {
		t.Fatalf("Key PEM encoding failed: %v", err)
	}

	mockDB := &workflowMockDBTX{
		privateKeyPEM: pemStr,
		certificate:   testZatcaCertificate(t, key),
		latestICV:     5,
	}
	repo := repository.New(mockDB)

	zatcaCfg := &ZatcaConfig{
		Enabled:  true,
		Env:      "sandbox",
		OrgVATID: "310123456700003",
	}
	zatcaUC := NewZatcaUseCase(zatcaCfg)
	zatcaUC.SetRepository(repo)

	posUC := NewPosUseCase()
	posUC.SetRepository(repo)
	posUC.SetZatcaUseCase(zatcaUC)

	var qty pgtype.Numeric
	_ = qty.Scan("1.000")
	var price, tax, total pgtype.Numeric
	_ = price.Scan("100.00")
	_ = tax.Scan("15.00")
	_ = total.Scan("115.00")

	input := &PosCreateTransactionInput{
		StoreID:           1,
		CashierID:         10,
		CashierSessionID:  88,
		PosTerminalID:     2,
		TransactionNumber: "POS-TXN-0077",
		Subtotal:          price, TaxAmount: tax, TotalAmount: total,
		Lines: []PosCreateTransactionLineInput{
			{
				ProductID: 301,
				Quantity:  qty,
				UnitPrice: price, Subtotal: price, TaxAmount: tax, LineTotal: total,
			},
		},
	}

	resp := posUC.CreateTransaction(context.Background(), input)
	if resp.StatusCode != 201 {
		t.Fatalf("CreateTransaction failed: %s", resp.Message)
	}

	out, ok := resp.Data.(CreateTransactionOutput)
	if !ok {
		t.Fatalf("Expected resp.Data of type CreateTransactionOutput, got %T", resp.Data)
	}

	// 1. Verify TLV QR code is returned
	if out.QRCode == "" {
		t.Error("Expected non-empty QRCode in CreateTransactionOutput")
	}
	if out.ZatcaUUID == "" {
		t.Error("Expected non-empty ZatcaUUID in CreateTransactionOutput")
	}

	// 2. Verify metadata contains ZATCA fields
	var metaMap map[string]interface{}
	err = json.Unmarshal(out.Metadata, &metaMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}
	if metaMap["qr_code_base64"] == nil || metaMap["qr_code_base64"] == "" {
		t.Error("Expected qr_code_base64 in metadata")
	}
	if metaMap["zatca_uuid"] == nil {
		t.Error("Expected zatca_uuid in metadata")
	}

	// 3. Verify zatca_document_chain entry was inserted
	if len(mockDB.chainEntries) == 0 {
		t.Fatal("Expected entry inserted into zatca_document_chain")
	}
	chain := mockDB.chainEntries[0]
	if chain.EntityType != ZatcaSalesOrderEntity {
		t.Errorf("expected order entity, got %s", chain.EntityType)
	}
	if chain.EntityID != mockDB.order.ID.String() {
		t.Errorf("expected sales order UUID, got %s", chain.EntityID)
	}
	if chain.Icv != 6 {
		t.Errorf("Expected ICV 6, got %d", chain.Icv)
	}

	// 4. Verify pos_transactions.metadata was persisted to DB
	if len(mockDB.updatedPosMetadata) == 0 {
		t.Fatal("Expected pos_transactions.metadata to be updated in database")
	}
	if mockDB.updatedPosMetadata[0].ID != 77 {
		t.Errorf("Expected updated pos_transaction ID 77, got %d", mockDB.updatedPosMetadata[0].ID)
	}
}

// TestSalesReturnRestockingAndCreditNote verifies:
// 1. Lines with ReturnToStock = true replenish inventory_stock and write stock_movement
// 2. Lines with ReturnToStock = false do not adjust inventory
// 3. ZATCA Credit Note (381) is signed and attached
func TestSalesReturnRestockingAndCreditNote(t *testing.T) {
	key, err := GenerateECDSAKeyPair()
	if err != nil {
		t.Fatalf("Key generation failed: %v", err)
	}
	pemStr, err := EncodePrivateKeyPEM(key)
	if err != nil {
		t.Fatalf("Key PEM encoding failed: %v", err)
	}

	mockDB := &workflowMockDBTX{
		privateKeyPEM: pemStr,
		certificate:   testZatcaCertificate(t, key),
		latestICV:     10,
	}
	repo := repository.New(mockDB)

	zatcaCfg := &ZatcaConfig{
		Enabled:  true,
		Env:      "sandbox",
		OrgVATID: "310123456700003",
	}
	zatcaUC := NewZatcaUseCase(zatcaCfg)
	zatcaUC.SetRepository(repo)

	salesReturnUC := NewSalesReturnUseCase()
	salesReturnUC.SetRepository(repo)
	salesReturnUC.SetZatcaUseCase(zatcaUC)

	origTxnID := int32(77)
	input := ProcessReturnInput{
		StoreID:               1,
		OriginalTransactionID: &origTxnID,
		ReturnReason:          "Defective unit exchange",
		Subtotal:              "140.00",
		TaxAmount:             "21.00",
		TotalRefundAmount:     "161.00",
		RefundMethod:          "card",
		Lines: []ProcessReturnLineInput{
			{
				ProductID:     401,
				Quantity:      "2.000",
				UnitPrice:     "50.00",
				CostPrice:     "35.00",
				RefundAmount:  "100.00",
				ReturnToStock: true, // Should replenish stock
				Condition:     "good",
			},
			{
				ProductID:     402,
				Quantity:      "1.000",
				UnitPrice:     "15.00",
				RefundAmount:  "15.00",
				ReturnToStock: false, // Damaged, do not restock
				Condition:     "damaged",
			},
			{
				ProductID:     403,
				Quantity:      "1.000",
				UnitPrice:     "25.00",
				RefundAmount:  "25.00",
				ReturnToStock: true, // ReturnToStock is true, BUT condition is damaged -> MUST NOT replenish sellable inventory
				Condition:     "damaged",
			},
		},
	}

	resp := salesReturnUC.ProcessSalesReturn(context.Background(), input)
	if resp.StatusCode != 201 {
		t.Fatalf("ProcessSalesReturn failed: %s", resp.Message)
	}

	// 1. Verify ONLY product 401 was restocked (402 has ReturnToStock=false; 403 has Condition='damaged')
	if len(mockDB.adjustedStocks) != 1 {
		t.Fatalf("Expected exactly 1 inventory adjustment for ReturnToStock=true and good condition, got %d", len(mockDB.adjustedStocks))
	}
	adj := mockDB.adjustedStocks[0]
	if adj.ProductID != 401 {
		t.Errorf("Expected product 401 restocked, got %d", adj.ProductID)
	}
	val, _ := adj.QuantityOnHand.Value()
	if !strings.HasPrefix(val.(string), "2") {
		t.Errorf("Expected positive replenishment quantity 2, got %v", val)
	}

	// 2. Verify stock movement was created for product 401 with movement_type 'return' and unit cost valuation
	if len(mockDB.stockMovements) != 1 {
		t.Fatalf("Expected exactly 1 stock movement for restocked line, got %d", len(mockDB.stockMovements))
	}
	mov := mockDB.stockMovements[0]
	if mov.MovementType != "return" {
		t.Errorf("Expected movement_type 'return', got '%s'", mov.MovementType)
	}
	if mov.ReferenceType.String != "sales_return" {
		t.Errorf("Expected reference_type 'sales_return', got '%s'", mov.ReferenceType.String)
	}
	// Verify movement used CostPrice 35.00 rather than selling price 50.00
	costVal, _ := mov.CostPerUnit.Value()
	if !strings.HasPrefix(costVal.(string), "35") {
		t.Errorf("Expected unit cost 35.00, got %v", costVal)
	}

	// 3. Verify ZATCA Credit Note was generated and returned
	out, ok := resp.Data.(ProcessSalesReturnOutput)
	if !ok {
		t.Fatalf("Expected resp.Data of type ProcessSalesReturnOutput, got %T", resp.Data)
	}
	if out.QRCode == "" {
		t.Error("Expected non-empty Credit Note QRCode in response")
	}

	// 4. Verify chain entry recorded with entity_type 'sales_return'
	if len(mockDB.chainEntries) == 0 {
		t.Fatal("Expected ZATCA chain entry recorded for credit note")
	}
	chain := mockDB.chainEntries[0]
	if chain.EntityType != "sales_return" {
		t.Errorf("Expected chain EntityType 'sales_return', got '%s'", chain.EntityType)
	}

	// 5. Verify sales_returns.metadata was persisted to DB
	if len(mockDB.updatedReturnMetadata) == 0 {
		t.Fatal("Expected sales_returns.metadata to be updated in database")
	}
	if mockDB.updatedReturnMetadata[0].ID != 55 {
		t.Errorf("Expected updated sales_return ID 55, got %d", mockDB.updatedReturnMetadata[0].ID)
	}
}

// TestDualOrderHarmonization verifies that POS transactions link to sales_orders_v2
// with sales_channel = 'pos', satisfy NOT NULL constraints, and appear in ListSalesOrdersDashboard.
func TestDualOrderHarmonization(t *testing.T) {
	mockDB := &workflowMockDBTX{}
	repo := repository.New(mockDB)

	posUC := NewPosUseCase()
	posUC.SetRepository(repo)

	var qty pgtype.Numeric
	_ = qty.Scan("1.000")

	input := &PosCreateTransactionInput{
		StoreID:           1,
		CashierID:         10,
		CashierSessionID:  88,
		PosTerminalID:     2,
		TransactionNumber: "POS-TXN-0077",
		Lines: []PosCreateTransactionLineInput{
			{
				ProductID: 501,
				Quantity:  qty,
			},
		},
	}

	resp := posUC.CreateTransaction(context.Background(), input)
	if resp.StatusCode != 201 {
		t.Fatalf("CreateTransaction failed: %s", resp.Message)
	}

	// 1. Verify sales_orders_v2 was created with sales_channel 'pos' and non-nil addresses
	if len(mockDB.createdOrdersV2) != 1 {
		t.Fatalf("Expected 1 sales_orders_v2 created, got %d", len(mockDB.createdOrdersV2))
	}
	so := mockDB.createdOrdersV2[0]
	if so.SalesChannel.String != "pos" {
		t.Errorf("Expected sales_channel 'pos', got '%s'", so.SalesChannel.String)
	}
	if len(so.ShippingAddress) == 0 || string(so.ShippingAddress) == "null" {
		t.Error("Expected non-null ShippingAddress to satisfy PostgreSQL NOT NULL constraint")
	}
	if len(so.BillingAddress) == 0 || string(so.BillingAddress) == "null" {
		t.Error("Expected non-null BillingAddress to satisfy PostgreSQL NOT NULL constraint")
	}

	// 2. Verify pos_transactions.sales_order_id was linked
	if len(mockDB.linkedOrders) != 1 {
		t.Fatalf("Expected pos_transactions to be linked to sales_orders_v2, got %d links", len(mockDB.linkedOrders))
	}
	link := mockDB.linkedOrders[0]
	if link.ID != 77 {
		t.Errorf("Expected linked pos_transaction ID 77, got %d", link.ID)
	}
	if !link.SalesOrderID.Valid {
		t.Error("Expected valid SalesOrderID UUID linked to pos_transactions")
	}

	// 3. Verify order appears in ListSalesOrdersDashboard
	store, _ := repo.GetStore(context.Background(), 1)
	dashboardRows, err := repo.ListSalesOrdersDashboard(context.Background(), repository.ListSalesOrdersDashboardParams{
		OrgID:        store.OrganizationID,
		SalesChannel: "pos",
	})
	if err != nil {
		t.Fatalf("ListSalesOrdersDashboard failed: %v", err)
	}
	if len(dashboardRows) != 1 {
		t.Fatalf("Expected 1 order in dashboard, got %d", len(dashboardRows))
	}
	if dashboardRows[0].SalesChannel.String != "pos" {
		t.Errorf("Expected sales_channel 'pos' in dashboard row, got '%s'", dashboardRows[0].SalesChannel.String)
	}
}

// Transactional test adapter for order-based issuance.
type workflowZatcaTx struct {
	pgx.Tx
	db *workflowMockDBTX
}

func (db *workflowMockDBTX) Begin(context.Context) (pgx.Tx, error) {
	return &workflowZatcaTx{db: db}, nil
}
func (tx *workflowZatcaTx) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return tx.db.Exec(ctx, sql, args...)
}
func (tx *workflowZatcaTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return tx.db.Query(ctx, sql, args...)
}
func (tx *workflowZatcaTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return tx.db.QueryRow(ctx, sql, args...)
}
func (tx *workflowZatcaTx) Commit(context.Context) error   { tx.db.commits++; return nil }
func (tx *workflowZatcaTx) Rollback(context.Context) error { tx.db.rollbacks++; return nil }
func scanOrderTestValue(value interface{}, dest ...interface{}) error {
	v := reflect.ValueOf(value)
	for i, d := range dest {
		reflect.ValueOf(d).Elem().Set(v.Field(i))
	}
	return nil
}
func orderTestChain(c repository.InsertChainEntryParams) repository.ZatcaDocumentChain {
	return repository.ZatcaDocumentChain{ID: 1, EntityType: c.EntityType, EntityID: c.EntityID, OrganizationID: c.OrganizationID, DeviceConfigID: c.DeviceConfigID, ZatcaUuid: c.ZatcaUuid, Icv: c.Icv, Pih: c.Pih, XmlHash: c.XmlHash, QrCodeBase64: c.QrCodeBase64, SignedXml: c.SignedXml, ZatcaStatus: c.ZatcaStatus}
}

type zatcaOrderRows struct {
	pgx.Rows
	values []repository.ListSalesOrderLinesV2Row
	index  int
}

func (r *zatcaOrderRows) Next() bool {
	if r.index >= len(r.values) {
		return false
	}
	r.index++
	return true
}
func (r *zatcaOrderRows) Scan(dest ...interface{}) error {
	return scanOrderTestValue(r.values[r.index-1], dest...)
}
func (r *zatcaOrderRows) Close()     {}
func (r *zatcaOrderRows) Err() error { return nil }
