package usecase

import (
	"context"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func orderNumber(value string) pgtype.Numeric { var n pgtype.Numeric; _ = n.Scan(value); return n }

func orderZatcaFixture(t *testing.T) (*workflowMockDBTX, *ZatcaUseCase) {
	t.Helper()
	key, _ := GenerateECDSAKeyPair()
	pem, _ := EncodePrivateKeyPEM(key)
	id := uuid.New()
	db := &workflowMockDBTX{privateKeyPEM: pem, certificate: testZatcaCertificate(t, key), order: repository.SalesOrdersV2{
		ID: id, OrderNumber: "SO-TEST-ORDER", OrganizationID: 10, StoreID: pgtype.Int4{Int32: 1, Valid: true}, PosTerminalID: pgtype.Int4{Int32: 2, Valid: true},
		OrderStatus: repository.OrderStatusV2Draft, Subtotal: orderNumber("100.00"), TaxAmount: orderNumber("15.00"), TotalAmount: orderNumber("115.00"),
	}, orderLines: []repository.ListSalesOrderLinesV2Row{{SalesOrderID: id, OrganizationID: 10, LineNumber: 1, ProductName: "Actual order product", QuantityOrdered: orderNumber("2"), UnitPrice: orderNumber("50"), TaxAmount: orderNumber("15"), TaxRate: orderNumber("15"), LineTotal: orderNumber("115")}}}
	uc := NewZatcaUseCase(&ZatcaConfig{Enabled: true, Env: "sandbox"})
	uc.SetRepository(repository.New(db))
	return db, uc
}

func TestZatcaSalesOrderSourceAndImmutableRetryWithoutStatusGate(t *testing.T) {
	db, uc := orderZatcaFixture(t)
	entry, err := uc.EnsureSalesOrderDocument(context.Background(), db.order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if entry.EntityType != ZatcaSalesOrderEntity || entry.EntityID != db.order.ID.String() {
		t.Fatal("document did not link to sales order UUID")
	}
	if !strings.Contains(entry.SignedXml.String, "SO-TEST-ORDER") || !strings.Contains(entry.SignedXml.String, "Actual order product") {
		t.Fatal("XML did not use persisted order data")
	}
	canonical, err := CanonicalizeXML([]byte(entry.SignedXml.String))
	if err != nil || HashSHA256Base64(canonical) != entry.XmlHash {
		t.Fatal("order XML hash mismatch")
	}
	// Once issued, retries must not regenerate using changed source data.
	db.order.OrderNumber = "MODIFIED"
	db.order.TotalAmount = orderNumber("999")
	db.orderLines = nil
	again, err := uc.EnsureSalesOrderDocument(context.Background(), db.order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(db.chainEntries) != 1 || again.ZatcaUuid != entry.ZatcaUuid || again.XmlHash != entry.XmlHash || again.SignedXml != entry.SignedXml {
		t.Fatal("retry regenerated an issued invoice")
	}
	if db.commits != 2 {
		t.Fatal("order transactions were not committed")
	}
}

func TestZatcaSalesOrderReusesLegacyTransactionDocument(t *testing.T) {
	db, uc := orderZatcaFixture(t)
	entry, err := uc.EnsureSalesOrderDocument(context.Background(), db.order.ID)
	if err != nil {
		t.Fatal(err)
	}
	db.chainEntries[0].EntityType = "pos_transaction"
	db.chainEntries[0].EntityID = "77"
	db.legacyTransactionID = 77
	again, err := uc.EnsureSalesOrderDocument(context.Background(), db.order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(db.chainEntries) != 1 || again.EntityType != "pos_transaction" || again.XmlHash != entry.XmlHash {
		t.Fatal("historical sale was issued twice")
	}
	transaction := &repository.PosTransaction{ID: 77, SalesOrderID: pgtype.UUID{Bytes: db.order.ID, Valid: true}}
	original, err := uc.OriginalIssuedInvoiceNumber(context.Background(), transaction)
	if err != nil || original != "SO-TEST-ORDER" {
		t.Fatal("return did not reference saved original invoice", original, err)
	}
}

func TestZatcaSalesOrderInvalidDataDoesNotConsumeChainEntry(t *testing.T) {
	for _, scenario := range []string{"no-lines", "totals", "line-total", "wrong-organization", "unsupported-rate", "shipping"} {
		t.Run(scenario, func(t *testing.T) {
			db, uc := orderZatcaFixture(t)
			switch scenario {
			case "no-lines":
				db.orderLines = nil
			case "totals":
				db.order.TotalAmount = orderNumber("999")
			case "line-total":
				db.orderLines[0].LineTotal = orderNumber("100")
			case "wrong-organization":
				db.orderLines[0].OrganizationID = 11
			case "unsupported-rate":
				db.orderLines[0].TaxRate = orderNumber("5")
			case "shipping":
				db.order.ShippingAmount = orderNumber("10")
			}
			if _, err := uc.EnsureSalesOrderDocument(context.Background(), db.order.ID); err == nil {
				t.Fatal("invalid order was signed")
			}
			if len(db.chainEntries) != 0 || db.commits != 0 {
				t.Fatal("invalid order consumed an invoice/chain number")
			}
		})
	}
}

func TestZatcaSalesOrderAllocatesFinalDiscountOnce(t *testing.T) {
	db, uc := orderZatcaFixture(t)
	db.order.DiscountAmount = orderNumber("10")
	db.order.TaxAmount = orderNumber("13.50")
	db.order.TotalAmount = orderNumber("103.50")
	entry, err := uc.EnsureSalesOrderDocument(context.Background(), db.order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(entry.SignedXml.String, `>103.50<`) || !strings.Contains(entry.SignedXml.String, `>13.50<`) {
		t.Fatal("discount/VAT totals were not mapped")
	}
	txn := &repository.PosTransaction{ID: 77, SalesOrderID: pgtype.UUID{Bytes: db.order.ID, Valid: true}}
	name, err := uc.OriginalIssuedInvoiceNumber(context.Background(), txn)
	if err != nil || name != db.order.OrderNumber {
		t.Fatal("new order return reference is incorrect", name, err)
	}
}
