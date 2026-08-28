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
	if canonical.ProductType != "standard" {
		t.Errorf("expected ProductType 'standard', got %s", canonical.ProductType)
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

func TestSAPProductMappingLeavesMissingBrandAndDescriptionMissing(t *testing.T) {
	canonical := (&mappings.SAPProduct{
		ItemCode:   "E2E-PANTENE-001",
		ItemName:   "Pantene shampoo",
		UserText:   "   ",
		ItmsGrpCod: 201,
		ItmsGrpNam: "Finished Goods",
		FirmCode:   0,
	}).ToCanonical()

	if canonical.BrandCode != "" {
		t.Fatalf("missing SAP FirmCode produced brand %q", canonical.BrandCode)
	}
	if canonical.Description != "" {
		t.Fatalf("whitespace-only SAP description was not treated as empty: %q", canonical.Description)
	}
}

func TestSAPProductMappingUsesPopulatedSAPBrandAndDescription(t *testing.T) {
	canonical := (&mappings.SAPProduct{
		ItemCode:   "E2E-PANTENE-001",
		ItemName:   "Pantene shampoo",
		UserText:   "Authoritative SAP description",
		ItmsGrpCod: 201,
		ItmsGrpNam: "Finished Goods",
		FirmCode:   42,
	}).ToCanonical()

	if canonical.BrandCode != "BRD-42" {
		t.Fatalf("SAP FirmCode mapping = %q, want BRD-42", canonical.BrandCode)
	}
	if canonical.Description != "Authoritative SAP description" {
		t.Fatalf("SAP description = %q, want populated source value", canonical.Description)
	}
}

func TestSAPProductTypeMapping(t *testing.T) {
	tests := []struct {
		name         string
		groupName    string
		categoryCode string
		productType  string
	}{
		{name: "Fixed Asset", groupName: "Fixed Asset", productType: mappings.ProductTypeFixedAsset},
		{name: "Fixed Assets with normalization", groupName: "  fIxEd AsSeTs  ", productType: mappings.ProductTypeFixedAsset},
		{name: "Raw Material", groupName: "Raw Material", productType: mappings.ProductTypeRawMaterial},
		{name: "Raw Materials", groupName: "RAW MATERIALS", productType: mappings.ProductTypeRawMaterial},
		{name: "normal category", groupName: " Drinks ", categoryCode: "CAT-201", productType: "standard"},
		{name: "unknown group", groupName: "Seasonal Specials", categoryCode: "CAT-201", productType: "standard"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canonical := (&mappings.SAPProduct{
				ItemCode:   "ITEM-001",
				ItemName:   "Test item",
				ItmsGrpCod: 201,
				ItmsGrpNam: tt.groupName,
			}).ToCanonical()

			if canonical.ProductType != tt.productType {
				t.Errorf("ProductType = %q, want %q", canonical.ProductType, tt.productType)
			}
			if canonical.CategoryCode != tt.categoryCode {
				t.Errorf("CategoryCode = %q, want %q", canonical.CategoryCode, tt.categoryCode)
			}
		})
	}
}

func TestSAPProductTypeGroupClassificationForCategoryFiltering(t *testing.T) {
	for _, groupName := range []string{"Fixed Asset", "Fixed Assets", "Raw Material", "Raw Materials"} {
		if mappings.ClassifySAPProductType(groupName) == "" {
			t.Errorf("expected %q to be classified as a product type group", groupName)
		}
	}
	for _, groupName := range []string{"Drinks", "Unrecognized Group"} {
		if mappings.ClassifySAPProductType(groupName) != "" {
			t.Errorf("expected %q to remain a category group", groupName)
		}
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
	if canonical.QuantityAllocated != 0 {
		t.Errorf("expected physical-only allocated 0, got %f", canonical.QuantityAllocated)
	}
	if canonical.QuantityAvailable != 100.0 {
		t.Errorf("expected physical-only available 100, got %f", canonical.QuantityAvailable)
	}
	if canonical.QuantityOnOrder != 0 {
		t.Errorf("expected physical-only on-order 0, got %f", canonical.QuantityOnOrder)
	}
	if canonical.Metadata["sap_is_commited"] != 25.0 || canonical.Metadata["sap_on_order"] != 50.0 {
		t.Errorf("expected committed/on-order values to remain reference metadata, got %#v", canonical.Metadata)
	}
	if canonical.Metadata["inventory_uom_to_base_factor"] != 1.0 {
		t.Errorf("expected unmanaged factor 1, got %#v", canonical.Metadata["inventory_uom_to_base_factor"])
	}
}

func TestInventoryNormalizationUsesAuthoritativeSAPFactor(t *testing.T) {
	tests := []struct {
		name       string
		onHand     float64
		baseQty    float64
		altQty     float64
		wantOnHand float64
	}{
		{name: "factor one", onHand: 10, baseQty: 1, altQty: 1, wantOnHand: 10},
		{name: "factor greater than one", onHand: 10, baseQty: 24, altQty: 1, wantOnHand: 240},
		{name: "fractional factor", onHand: 10, baseQty: 1, altQty: 2, wantOnHand: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stock := mappings.SAPInventoryStock{
				ItemCode:            "ITEM-001",
				WhsCode:             "01",
				OnHand:              tt.onHand,
				UgpEntry:            10,
				IUoMEntry:           20,
				InventoryUom:        "BOX",
				BaseUomEntry:        1,
				BaseUomCode:         "PCS",
				InventoryUomBaseQty: tt.baseQty,
				InventoryUomAltQty:  tt.altQty,
			}
			canonical, err := stock.ToCanonicalChecked()
			if err != nil {
				t.Fatalf("ToCanonicalChecked() error = %v", err)
			}
			if canonical.QuantityOnHand != tt.wantOnHand || canonical.QuantityAvailable != tt.wantOnHand {
				t.Fatalf("normalized quantities = (%v, %v), want (%v, %v)", canonical.QuantityOnHand, canonical.QuantityAvailable, tt.wantOnHand, tt.wantOnHand)
			}
			if canonical.QuantityAllocated != 0 || canonical.QuantityOnOrder != 0 {
				t.Fatalf("PHYSICAL_ONLY operational quantities = allocated %v/on-order %v, want zeroes", canonical.QuantityAllocated, canonical.QuantityOnOrder)
			}
			if canonical.Metadata["source_on_hand"] != tt.onHand || canonical.Metadata["inventory_uom_to_base_factor"] != tt.wantOnHand/tt.onHand {
				t.Fatalf("raw/factor lineage = %#v, want source_on_hand %v and factor %v", canonical.Metadata, tt.onHand, tt.wantOnHand/tt.onHand)
			}
		})
	}
}

func TestInventoryNormalizationResolvesBaseUOMAsOne(t *testing.T) {
	stock := mappings.SAPInventoryStock{
		ItemCode:     "ITEM-BASE",
		UgpEntry:     10,
		IUoMEntry:    7,
		InventoryUom: "PCS",
		BaseUomEntry: 7,
		BaseUomCode:  "PCS",
		OnHand:       12,
	}

	canonical, err := stock.ToCanonicalChecked()
	if err != nil {
		t.Fatalf("base-UoM normalization error = %v", err)
	}
	if canonical.QuantityOnHand != 12 || canonical.Metadata["inventory_uom_to_base_factor"] != 1.0 {
		t.Fatalf("base-UoM result = quantity %v, factor %#v; want 12 and 1", canonical.QuantityOnHand, canonical.Metadata["inventory_uom_to_base_factor"])
	}
}

func TestInventoryNormalizationFailsClosedForManagedMissingOrInvalidFactor(t *testing.T) {
	tests := []struct {
		name    string
		baseQty float64
		altQty  float64
	}{
		{name: "missing conversion", baseQty: 0, altQty: 0},
		{name: "zero base quantity", baseQty: 0, altQty: 1},
		{name: "zero alternate quantity", baseQty: 1, altQty: 0},
		{name: "negative factor input", baseQty: -1, altQty: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stock := mappings.SAPInventoryStock{
				ItemCode:            "ITEM-INVALID",
				UgpEntry:            10,
				IUoMEntry:           20,
				InventoryUom:        "BOX",
				BaseUomEntry:        1,
				BaseUomCode:         "PCS",
				InventoryUomBaseQty: tt.baseQty,
				InventoryUomAltQty:  tt.altQty,
			}
			if _, err := stock.ToCanonicalChecked(); err == nil {
				t.Fatal("managed inventory with invalid conversion was accepted")
			}
		})
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

func TestProductMappingUsesUOMGroupBaseWhenAvailable(t *testing.T) {
	canonical := (&mappings.SAPProduct{
		ItemCode:     "ITEM-UOM-GROUP",
		ItemName:     "Boxed item",
		InvntryUom:   "BOX",
		BaseUomEntry: 1,
		BaseUomCode:  "PCS",
		UgpEntry:     10,
	}).ToCanonical()

	if canonical.BaseUOMCode != "PCS" || canonical.UOMCode != "PCS" {
		t.Fatalf("product base UoM = (%q, %q), want PCS", canonical.BaseUOMCode, canonical.UOMCode)
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

func TestPriceListHeaderAndItemCanonicalCodesMatch(t *testing.T) {
	header := (&mappings.SAPPriceList{ListNum: 10}).ToCanonical()
	item := (&mappings.SAPPriceListItem{PriceList: 10}).ToCanonical()
	if header.Code != item.PriceListCode {
		t.Fatalf("price-list identities differ: header=%q item=%q", header.Code, item.PriceListCode)
	}
}

func TestPriceListItemPreservesSourceCurrencyMetadata(t *testing.T) {
	item := (&mappings.SAPPriceListItem{PriceList: 10, Currency: " SAR "}).ToCanonical()
	if got := item.Metadata["sap_item_currency"]; got != "SAR" {
		t.Fatalf("sap_item_currency metadata = %#v, want SAR", got)
	}
}

func TestSAPBusinessPartnerCanonicalMapping(t *testing.T) {
	tests := []struct {
		name       string
		cardType   string
		validFor   string
		frozenFor  string
		wantType   string
		wantActive bool
	}{
		{name: "supplier", cardType: "S", validFor: "Y", frozenFor: "N", wantType: "supplier", wantActive: true},
		{name: "supplier frozen", cardType: "S", validFor: "Y", frozenFor: "Y", wantType: "supplier", wantActive: false},
		{name: "customer frozen", cardType: "C", validFor: "Y", frozenFor: "Y", wantType: "customer", wantActive: false},
		{name: "valid for inactive", cardType: "C", validFor: "N", frozenFor: "N", wantType: "customer", wantActive: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canonical := (&mappings.SAPBusinessPartner{
				CardCode:  "BP-001",
				CardName:  "Partner",
				CardType:  tt.cardType,
				ValidFor:  tt.validFor,
				FrozenFor: tt.frozenFor,
			}).ToCanonical()

			if canonical.PartnerType != tt.wantType {
				t.Fatalf("PartnerType = %q, want %q", canonical.PartnerType, tt.wantType)
			}
			if canonical.IsActive != tt.wantActive {
				t.Fatalf("IsActive = %t, want %t", canonical.IsActive, tt.wantActive)
			}
		})
	}
}
