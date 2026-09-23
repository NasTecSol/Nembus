package mappings_test

import (
	"testing"
	"time"

	"github.com/NasTecSol/nembus-sap/mappings"
)

func TestSAPBool(t *testing.T) {
	tests := []struct {
		input    string
		fallback bool
		expected bool
	}{
		{"Y", false, true},
		{"y", false, true},
		{"1", false, true},
		{"N", true, false},
		{"n", true, false},
		{"0", true, false},
		{"", true, true},
		{"", false, false},
		{"invalid", true, true},
	}

	for _, tt := range tests {
		got := mappings.SAPBool(tt.input, tt.fallback)
		if got != tt.expected {
			t.Errorf("SAPBool(%q, %v) = %v, want %v", tt.input, tt.fallback, got, tt.expected)
		}
	}
}

func TestStoreMapping(t *testing.T) {
	sapStore := mappings.SAPStore{
		WhsCode: "01",
		WhsName: "Main Warehouse",
		Locked:  "N",
		Street:  "Industrial Ave 42",
		City:    "Riyadh",
		Country: "SA",
		ZipCode: "12345",
	}

	canonical := sapStore.ToCanonical()
	if canonical.Code != "01" {
		t.Errorf("expected code '01', got %s", canonical.Code)
	}
	if canonical.Name != "Main Warehouse" {
		t.Errorf("expected name 'Main Warehouse', got %s", canonical.Name)
	}
	if !canonical.IsActive {
		t.Errorf("expected is_active true for Locked='N'")
	}
	if canonical.Metadata["sap_city"] != "Riyadh" {
		t.Errorf("expected metadata city 'Riyadh'")
	}
}

func TestProductMapping(t *testing.T) {
	sapProd := mappings.SAPProduct{
		ItemCode:   "ITEM-001",
		ItemName:   "Premium Coffee Beans 1kg",
		UserText:   "Arabica roasted coffee",
		ItmsGrpCod: 105,
		FirmCode:   12,
		InvntItem:  "Y",
		SellItem:   "Y",
		PrchseItem: "Y",
		ValidFor:   "Y",
		CodeBars:   "6281000123456",
		SalUnitMsr: "KG",
		ManBtchNum: "Y",
		ManSerNum:  "N",
	}

	canonical := sapProd.ToCanonical()
	if canonical.SKU != "ITEM-001" {
		t.Errorf("expected SKU 'ITEM-001', got %s", canonical.SKU)
	}
	if canonical.CategoryCode != "CAT-105" {
		t.Errorf("expected CategoryCode 'CAT-105', got %s", canonical.CategoryCode)
	}
	if canonical.BrandCode != "BRD-12" {
		t.Errorf("expected BrandCode 'BRD-12', got %s", canonical.BrandCode)
	}
	if canonical.UOMCode != "KG" {
		t.Errorf("expected UOMCode 'KG', got %s", canonical.UOMCode)
	}
	if !canonical.IsBatchManaged {
		t.Errorf("expected IsBatchManaged true for ManBtchNum='Y'")
	}
	if canonical.IsSerialized {
		t.Errorf("expected IsSerialized false for ManSerNum='N'")
	}
	if canonical.PrimaryBarcode != "6281000123456" {
		t.Errorf("expected PrimaryBarcode '6281000123456', got %s", canonical.PrimaryBarcode)
	}
}

func TestInventoryStockCalculation(t *testing.T) {
	stock := mappings.SAPInventoryStock{
		ItemCode:   "ITEM-001",
		WhsCode:    "01",
		OnHand:     100.0,
		IsCommited: 25.0,
		OnOrder:    50.0,
		MinStock:   10.0,
		MaxStock:   200.0,
	}

	canonical := stock.ToCanonical()
	if canonical.QuantityOnHand != 100.0 {
		t.Errorf("expected on_hand 100, got %f", canonical.QuantityOnHand)
	}
	if canonical.QuantityAllocated != 25.0 {
		t.Errorf("expected allocated 25, got %f", canonical.QuantityAllocated)
	}
	if canonical.QuantityAvailable != 75.0 {
		t.Errorf("expected available 75, got %f", canonical.QuantityAvailable)
	}
}

func TestInvoiceMapping(t *testing.T) {
	now := time.Now()
	inv := mappings.SAPInvoice{
		DocEntry:   1001,
		DocNum:     200045,
		DocDate:    now,
		DocDueDate: now.AddDate(0, 0, 30),
		CardCode:   "C0001",
		CardName:   "Acme Corp",
		DocTotal:   575.00,
		PaidToDate: 200.00,
		VatSum:     75.00,
		DiscSum:    0.00,
		DocStatus:  "O",
		Lines: []mappings.SAPInvoiceLine{
			{
				DocEntry:   1001,
				LineNum:    0,
				ItemCode:   "ITEM-001",
				Dscription: "Premium Coffee Beans 1kg",
				Quantity:   5.0,
				Price:      100.0,
				LineTotal:  500.0,
				VatSum:     75.0,
				WhsCode:    "01",
				UnitMsr:    "KG",
			},
		},
	}

	canonical := inv.ToCanonical()
	if canonical.InvoiceNumber != "INV-SAP-200045" {
		t.Errorf("expected invoice number 'INV-SAP-200045', got %s", canonical.InvoiceNumber)
	}
	if canonical.InvoiceStatus != "partially_paid" {
		t.Errorf("expected invoice status 'partially_paid', got %s", canonical.InvoiceStatus)
	}
	if canonical.BalanceDue != 375.00 {
		t.Errorf("expected balance due 375.00, got %f", canonical.BalanceDue)
	}
	if len(canonical.Lines) != 1 {
		t.Fatalf("expected 1 line item, got %d", len(canonical.Lines))
	}
	if canonical.Lines[0].LineTotal != 575.00 {
		t.Errorf("expected line total 575.00, got %f", canonical.Lines[0].LineTotal)
	}
}

func TestUOMMapping(t *testing.T) {
	sapUom := mappings.SAPUOM{
		UomEntry: 5,
		UomCode:  "BOX",
		UomName:  "Box of 12",
		Locked:   "N",
	}

	canonical := sapUom.ToCanonical()
	if canonical.Code != "BOX" {
		t.Errorf("expected code 'BOX', got %s", canonical.Code)
	}
	if canonical.Name != "Box of 12" {
		t.Errorf("expected name 'Box of 12', got %s", canonical.Name)
	}
	if !canonical.IsActive {
		t.Errorf("expected is_active true for Locked='N'")
	}
	if canonical.Metadata["sap_uom_entry"] != int64(5) {
		t.Errorf("expected sap_uom_entry 5, got %v", canonical.Metadata["sap_uom_entry"])
	}
}

func TestProductMappingWithUOMConversions(t *testing.T) {
	sapProd := mappings.SAPProduct{
		ItemCode:   "BEV-001",
		ItemName:   "Sparkling Water 330ml",
		ItmsGrpCod: 201,
		FirmCode:   5,
		InvntItem:  "Y",
		SellItem:   "Y",
		PrchseItem: "Y",
		ValidFor:   "Y",
		CodeBars:   "1234567890123",
		InvntryUom: "PCS",
		SalUnitMsr: "BOX",
		BuyUnitMsr: "PALLET",
		NumInSale:  24.0,
		NumInBuy:   480.0,
		UgpEntry:   10,
	}

	canonical := sapProd.ToCanonical()
	if canonical.BaseUOMCode != "PCS" {
		t.Errorf("expected BaseUOMCode 'PCS', got %s", canonical.BaseUOMCode)
	}
	if canonical.SalesUOMCode != "BOX" {
		t.Errorf("expected SalesUOMCode 'BOX', got %s", canonical.SalesUOMCode)
	}
	if canonical.PurchaseUOMCode != "PALLET" {
		t.Errorf("expected PurchaseUOMCode 'PALLET', got %s", canonical.PurchaseUOMCode)
	}
	if canonical.UOMGroupCode != "UGP-10" {
		t.Errorf("expected UOMGroupCode 'UGP-10', got %s", canonical.UOMGroupCode)
	}
	if len(canonical.UOMConversions) != 2 {
		t.Fatalf("expected 2 UOM conversions, got %d", len(canonical.UOMConversions))
	}
	if canonical.UOMConversions[0].FromUOMCode != "BOX" || canonical.UOMConversions[0].ConversionFactor != 24.0 {
		t.Errorf("expected conversion BOX -> PCS with factor 24.0, got %+v", canonical.UOMConversions[0])
	}
	if canonical.UOMConversions[1].FromUOMCode != "PALLET" || canonical.UOMConversions[1].ConversionFactor != 480.0 {
		t.Errorf("expected conversion PALLET -> PCS with factor 480.0, got %+v", canonical.UOMConversions[1])
	}
}

func TestBarcodeAndPriceWithUOM(t *testing.T) {
	barcode := mappings.SAPBarcode{
		BcdEntry: 1,
		BcdCode:  "9876543210987",
		ItemCode: "BEV-001",
		UomEntry: 5,
		UomCode:  "BOX",
	}
	canonBarcode := barcode.ToCanonical(false)
	if canonBarcode.UOMCode != "BOX" {
		t.Errorf("expected UOMCode 'BOX', got %s", canonBarcode.UOMCode)
	}

	price := mappings.SAPPriceListItem{
		ItemCode:  "BEV-001",
		PriceList: 1,
		Price:     45.0,
		Currency:  "USD",
		UomEntry:  5,
		UomCode:   "BOX",
	}
	canonPrice := price.ToCanonical()
	if canonPrice.UOMCode != "BOX" {
		t.Errorf("expected UOMCode 'BOX', got %s", canonPrice.UOMCode)
	}
}

func TestPurchaseOrderMapping(t *testing.T) {
	po := mappings.SAPPurchaseOrder{
		DocEntry:   1001,
		DocNum:     5001,
		DocDate:    time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		DocDueDate: time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
		CardCode:   "SUP-001",
		CardName:   "Acme Roasters Ltd",
		DocTotal:   1150.0,
		VatSum:     150.0,
		DiscSum:    0.0,
		DocStatus:  "O",
		Lines: []mappings.SAPPurchaseOrderLine{
			{
				DocEntry:   1001,
				LineNum:    0,
				ItemCode:   "ITEM-001",
				Dscription: "Premium Coffee Beans 1kg",
				Quantity:   100,
				Price:      10.0,
				LineTotal:  1000.0,
				VatSum:     150.0,
				WhsCode:    "01",
				UnitMsr:    "KG",
				OpenQty:    50, // 50 already received -> partially_received
			},
		},
	}

	canon := po.ToCanonical()
	if canon.PONumber != "PO-5001" {
		t.Errorf("expected PONumber 'PO-5001', got %s", canon.PONumber)
	}
	if canon.SupplierCode != "SUP-001" {
		t.Errorf("expected SupplierCode 'SUP-001', got %s", canon.SupplierCode)
	}
	if canon.Status != "partially_received" {
		t.Errorf("expected Status 'partially_received', got %s", canon.Status)
	}
	if canon.Subtotal != 1000.0 {
		t.Errorf("expected Subtotal 1000.0, got %f", canon.Subtotal)
	}
	if len(canon.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(canon.Lines))
	}
	if canon.Lines[0].ReceivedQuantity != 50.0 {
		t.Errorf("expected ReceivedQuantity 50.0, got %f", canon.Lines[0].ReceivedQuantity)
	}
}

func TestGoodsReceiptMapping(t *testing.T) {
	gr := mappings.SAPGoodsReceipt{
		DocEntry:   2001,
		DocNum:     6001,
		DocDate:    time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
		CardCode:   "SUP-001",
		CardName:   "Acme Roasters Ltd",
		DocTotal:   575.0,
		VatSum:     75.0,
		DocStatus:  "C",
		Lines: []mappings.SAPGoodsReceiptLine{
			{
				DocEntry:   2001,
				LineNum:    0,
				ItemCode:   "ITEM-001",
				Dscription: "Premium Coffee Beans 1kg",
				Quantity:   50,
				Price:      10.0,
				LineTotal:  500.0,
				VatSum:     75.0,
				WhsCode:    "01",
				UnitMsr:    "KG",
				BaseEntry:  1001,
				BaseLine:   0,
				BaseType:   22, // PO
			},
		},
	}

	canon := gr.ToCanonical()
	if canon.GRNNumber != "GRN-6001" {
		t.Errorf("expected GRNNumber 'GRN-6001', got %s", canon.GRNNumber)
	}
	if canon.PONumber != "PO-1001" {
		t.Errorf("expected PONumber 'PO-1001', got %s", canon.PONumber)
	}
	if canon.Status != "posted" {
		t.Errorf("expected Status 'posted', got %s", canon.Status)
	}
	if len(canon.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(canon.Items))
	}
	if canon.Items[0].QuantityReceived != 50.0 {
		t.Errorf("expected QuantityReceived 50.0, got %f", canon.Items[0].QuantityReceived)
	}
}

func TestStockMovementMapping(t *testing.T) {
	sm := mappings.SAPStockMovement{
		TransNum:   9001,
		TransType:  20, // Goods Receipt PO
		CreatedBy:  2001,
		BaseRef:    "6001",
		DocDate:    time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
		ItemCode:   "ITEM-001",
		Warehouse:  "01",
		InQty:      50.0,
		OutQty:     0.0,
		Price:      10.0,
		TransValue: 500.0,
	}

	canon := sm.ToCanonical()
	if canon.MovementType != "purchase_receipt" {
		t.Errorf("expected MovementType 'purchase_receipt', got %s", canon.MovementType)
	}
	if canon.ReferenceType != "goods_receipt_note" {
		t.Errorf("expected ReferenceType 'goods_receipt_note', got %s", canon.ReferenceType)
	}
	if canon.ReferenceNumber != "6001" {
		t.Errorf("expected ReferenceNumber '6001', got %s", canon.ReferenceNumber)
	}
	if canon.Quantity != 50.0 {
		t.Errorf("expected Quantity 50.0, got %f", canon.Quantity)
	}
	if canon.ToStoreCode != "01" {
		t.Errorf("expected ToStoreCode '01', got %s", canon.ToStoreCode)
	}
}

func TestIncomingPaymentMapping(t *testing.T) {
	p := mappings.SAPIncomingPayment{
		DocEntry:  3001,
		DocNum:    7001,
		DocDate:   time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		CardCode:  "CUST-001",
		CardName:  "Retail Customer Inc",
		DocCurr:   "SAR",
		DocTotal:  230.0,
		CashSum:   230.0,
		TrsfrSum:  0.0,
		CheckSum:  0.0,
		CreditSum: 0.0,
		Comments:  "Cash payment for invoice",
		Invoices: []mappings.SAPIncomingPaymentInvoice{
			{
				DocNum:        3001,
				LineID:        0,
				InvoiceID:     501,
				InvType:       13,
				SumApplied:    230.0,
				InvoiceDocNum: 9001,
			},
		},
	}

	canons := p.ToCanonicalList()
	if len(canons) != 1 {
		t.Fatalf("expected 1 canonical payment, got %d", len(canons))
	}
	pay := canons[0]
	if pay.PaymentNumber != "PAY-SAP-7001-0" {
		t.Errorf("expected PaymentNumber 'PAY-SAP-7001-0', got %s", pay.PaymentNumber)
	}
	if pay.InvoiceNumber != "INV-SAP-9001" {
		t.Errorf("expected InvoiceNumber 'INV-SAP-9001', got %s", pay.InvoiceNumber)
	}
	if pay.SAPInvoiceDocEntry != 501 {
		t.Errorf("expected SAPInvoiceDocEntry 501, got %d", pay.SAPInvoiceDocEntry)
	}
	if pay.PaymentMethod != "cash" {
		t.Errorf("expected PaymentMethod 'cash', got %s", pay.PaymentMethod)
	}
	if pay.PaymentAmount != 230.0 {
		t.Errorf("expected PaymentAmount 230.0, got %f", pay.PaymentAmount)
	}
	if pay.CurrencyCode != "SAR" {
		t.Errorf("expected CurrencyCode 'SAR', got %s", pay.CurrencyCode)
	}
}



