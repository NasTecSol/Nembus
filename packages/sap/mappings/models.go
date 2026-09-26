package mappings

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"
)

// Helper: Convert SAP 'Y' / 'N' strings to boolean
func SAPBool(val string, defaultVal bool) bool {
	clean := strings.ToUpper(strings.TrimSpace(val))
	if clean == "Y" || clean == "1" || clean == "TRUE" {
		return true
	}
	if clean == "N" || clean == "0" || clean == "FALSE" {
		return false
	}
	return defaultVal
}

// Helper: Generate a secure temporary random hex password hash / token
func GenerateRandomHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("tmp_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// ----------------------------------------------------
// Domain 1: Stores & Locations (OWHS, OBIN)
// ----------------------------------------------------

type SAPStore struct {
	WhsCode  string                 `json:"whs_code"`
	WhsName  string                 `json:"whs_name"`
	Locked   string                 `json:"locked"`
	Street   string                 `json:"street,omitempty"`
	City     string                 `json:"city,omitempty"`
	Country  string                 `json:"country,omitempty"`
	ZipCode  string                 `json:"zip_code,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type CanonicalStore struct {
	Code         string                 `json:"code"`
	Name         string                 `json:"name"`
	StoreType    string                 `json:"store_type"`
	IsWarehouse  bool                   `json:"is_warehouse"`
	IsPosEnabled bool                   `json:"is_pos_enabled"`
	IsActive     bool                   `json:"is_active"`
	Timezone     string                 `json:"timezone"`
	Metadata     map[string]interface{} `json:"metadata"`
}

func (s *SAPStore) ToCanonical() CanonicalStore {
	isActive := !SAPBool(s.Locked, false)
	meta := s.Metadata
	if meta == nil {
		meta = make(map[string]interface{})
	}
	if s.Street != "" {
		meta["sap_street"] = s.Street
	}
	if s.City != "" {
		meta["sap_city"] = s.City
	}
	if s.Country != "" {
		meta["sap_country"] = s.Country
	}
	if s.ZipCode != "" {
		meta["sap_zip_code"] = s.ZipCode
	}

	return CanonicalStore{
		Code:         strings.TrimSpace(s.WhsCode),
		Name:         strings.TrimSpace(s.WhsName),
		StoreType:    "retail",
		IsWarehouse:  true,
		IsPosEnabled: true,
		IsActive:     isActive,
		Timezone:     "UTC",
		Metadata:     meta,
	}
}

type SAPStorageLocation struct {
	AbsEntry int64                  `json:"abs_entry"`
	BinCode  string                 `json:"bin_code"`
	WhsCode  string                 `json:"whs_code"`
	Descr    string                 `json:"descr,omitempty"`
	Disabled string                 `json:"disabled"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type CanonicalStorageLocation struct {
	StoreCode    string                 `json:"store_code"`
	Code         string                 `json:"code"`
	Name         string                 `json:"name"`
	LocationType string                 `json:"location_type"`
	IsActive     bool                   `json:"is_active"`
	Metadata     map[string]interface{} `json:"metadata"`
}

func (b *SAPStorageLocation) ToCanonical() CanonicalStorageLocation {
	isActive := !SAPBool(b.Disabled, false)
	name := strings.TrimSpace(b.Descr)
	if name == "" {
		name = strings.TrimSpace(b.BinCode)
	}
	meta := b.Metadata
	if meta == nil {
		meta = make(map[string]interface{})
	}
	meta["sap_abs_entry"] = b.AbsEntry

	return CanonicalStorageLocation{
		StoreCode:    strings.TrimSpace(b.WhsCode),
		Code:         strings.TrimSpace(b.BinCode),
		Name:         name,
		LocationType: "standard",
		IsActive:     isActive,
		Metadata:     meta,
	}
}

// ----------------------------------------------------
// Domain 2: Users & Cashiers (OUSR, OSLP)
// ----------------------------------------------------

type SAPUser struct {
	UserID   int64  `json:"user_id"`
	UserCode string `json:"user_code"`
	UName    string `json:"u_name"`
	EMail    string `json:"e_mail"`
	Locked   string `json:"locked"`
}

type CanonicalUser struct {
	Username     string                 `json:"username"`
	Email        string                 `json:"email"`
	PasswordHash string                 `json:"password_hash"`
	FirstName    string                 `json:"first_name"`
	LastName     string                 `json:"last_name"`
	EmployeeCode string                 `json:"employee_code"`
	IsActive     bool                   `json:"is_active"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// SAPImportPasswordSentinel is stored as the password hash for all imported SAP users.
// The cloud server MUST treat this sentinel as a mandatory password-reset trigger —
// the value is NOT a valid bcrypt hash and cannot be used to authenticate.
const SAPImportPasswordSentinel = "{SAP_IMPORT_MUST_RESET}"

func (u *SAPUser) ToCanonical() CanonicalUser {
	isActive := !SAPBool(u.Locked, false)
	email := strings.TrimSpace(u.EMail)
	username := strings.ToLower(strings.TrimSpace(u.UserCode))
	if email == "" {
		email = fmt.Sprintf("%s@imported-sap.local", username)
	}

	parts := strings.SplitN(strings.TrimSpace(u.UName), " ", 2)
	firstName := parts[0]
	lastName := ""
	if len(parts) > 1 {
		lastName = parts[1]
	}

	return CanonicalUser{
		Username:     username,
		Email:        email,
		PasswordHash: SAPImportPasswordSentinel, // Sentinel — cloud must force password reset
		FirstName:    firstName,
		LastName:     lastName,
		EmployeeCode: fmt.Sprintf("SAP-%d", u.UserID),
		IsActive:     isActive,
		Metadata: map[string]interface{}{
			"sap_user_id":   u.UserID,
			"sap_user_code": u.UserCode,
		},
	}
}

type SAPCashier struct {
	SlpCode   int64  `json:"slp_code"`
	SlpName   string `json:"slp_name"`
	Memo      string `json:"memo,omitempty"`
	Active    string `json:"active"`
	Email     string `json:"email,omitempty"`
	Telephone string `json:"telephone,omitempty"`
}

type CanonicalCashier struct {
	CashierCode   string                 `json:"cashier_code"`
	Username      string                 `json:"username"`
	StoreCode     string                 `json:"store_code,omitempty"`
	DrawerLimit   float64                `json:"drawer_limit"`
	DiscountLimit float64                `json:"discount_limit"`
	IsActive      bool                   `json:"is_active"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// CashierDefaults holds org-configurable limits applied to all migrated cashiers.
type CashierDefaults struct {
	DefaultStoreCode string
	DrawerLimit      float64
	DiscountLimit    float64
}

func (c *SAPCashier) ToCanonical(defaults CashierDefaults) CanonicalCashier {
	isActive := SAPBool(c.Active, true)
	code := fmt.Sprintf("CSH-%d", c.SlpCode)
	username := fmt.Sprintf("slp_%d", c.SlpCode)

	drawerLimit := defaults.DrawerLimit
	if drawerLimit <= 0 {
		drawerLimit = 5000.0
	}
	discountLimit := defaults.DiscountLimit
	if discountLimit < 0 {
		discountLimit = 20.0
	}

	return CanonicalCashier{
		CashierCode:   code,
		Username:      username,
		StoreCode:     defaults.DefaultStoreCode,
		DrawerLimit:   drawerLimit,
		DiscountLimit: discountLimit,
		IsActive:      isActive,
		Metadata: map[string]interface{}{
			"sap_slp_code": c.SlpCode,
			"sap_slp_name": c.SlpName,
			"sap_memo":     c.Memo,
			"sap_phone":    c.Telephone,
		},
	}
}

// ----------------------------------------------------
// Domain 3: Catalog (OUOM, OUGP, UGP1, OITB, OMRG, OITM, OBCD)
// ----------------------------------------------------

type SAPUOM struct {
	UomEntry int64  `json:"uom_entry"`
	UomCode  string `json:"uom_code"`
	UomName  string `json:"uom_name"`
	Locked   string `json:"locked"`
}

type CanonicalUOM struct {
	Code          string                 `json:"code"`
	Name          string                 `json:"name"`
	UOMType       string                 `json:"uom_type"`
	DecimalPlaces int                    `json:"decimal_places"`
	IsActive      bool                   `json:"is_active"`
	Metadata      map[string]interface{} `json:"metadata"`
}

func (u *SAPUOM) ToCanonical() CanonicalUOM {
	code := strings.TrimSpace(u.UomCode)
	name := strings.TrimSpace(u.UomName)
	if name == "" {
		name = code
	}
	return CanonicalUOM{
		Code:          code,
		Name:          name,
		UOMType:       "unit",
		DecimalPlaces: 2,
		IsActive:      !SAPBool(u.Locked, false),
		Metadata: map[string]interface{}{
			"sap_uom_entry": u.UomEntry,
		},
	}
}

type SAPUOMGroupDetail struct {
	UgpEntry     int64   `json:"ugp_entry"`
	UgpCode      string  `json:"ugp_code"`
	UgpName      string  `json:"ugp_name"`
	BaseUomEntry int64   `json:"base_uom_entry"`
	BaseUomCode  string  `json:"base_uom_code"`
	AltUomEntry  int64   `json:"alt_uom_entry"`
	AltUomCode   string  `json:"alt_uom_code"`
	AltQty       float64 `json:"alt_qty"`
	BaseQty      float64 `json:"base_qty"`
}

type CanonicalUOMGroupLevel struct {
	LevelOrder       int     `json:"level_order"`
	UOMCode          string  `json:"uom_code"`
	Multiplier       float64 `json:"multiplier"`
	ConversionFactor float64 `json:"conversion_factor"` // BaseQty / AltQty
}

type CanonicalUOMGroup struct {
	Code        string                   `json:"code"`
	Name        string                   `json:"name"`
	BaseUOMCode string                   `json:"base_uom_code"`
	IsActive    bool                     `json:"is_active"`
	Levels      []CanonicalUOMGroupLevel `json:"levels"`
	Metadata    map[string]interface{}   `json:"metadata"`
}

type CanonicalProductUOMConversion struct {
	ProductSKU       string                 `json:"product_sku"`
	FromUOMCode      string                 `json:"from_uom_code"`
	ToUOMCode        string                 `json:"to_uom_code"`
	ConversionFactor float64                `json:"conversion_factor"`
	IsDefault        bool                   `json:"is_default"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type SAPCategory struct {
	ItmsGrpCod int64  `json:"itms_grp_cod"`
	ItmsGrpNam string `json:"itms_grp_nam"`
}

type CanonicalCategory struct {
	Code          string                 `json:"code"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	CategoryLevel int                    `json:"category_level"`
	IsActive      bool                   `json:"is_active"`
	Metadata      map[string]interface{} `json:"metadata"`
}

func (cat *SAPCategory) ToCanonical() CanonicalCategory {
	return CanonicalCategory{
		Code:          fmt.Sprintf("CAT-%d", cat.ItmsGrpCod),
		Name:          strings.TrimSpace(cat.ItmsGrpNam),
		Description:   fmt.Sprintf("Imported SAP Item Group %d", cat.ItmsGrpCod),
		CategoryLevel: 1,
		IsActive:      true,
		Metadata: map[string]interface{}{
			"sap_itms_grp_cod": cat.ItmsGrpCod,
		},
	}
}

type SAPBrand struct {
	FirmCode int64  `json:"firm_code"`
	FirmName string `json:"firm_name"`
}

type CanonicalBrand struct {
	Code     string                 `json:"code"`
	Name     string                 `json:"name"`
	IsActive bool                   `json:"is_active"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (b *SAPBrand) ToCanonical() CanonicalBrand {
	return CanonicalBrand{
		Code:     fmt.Sprintf("BRD-%d", b.FirmCode),
		Name:     strings.TrimSpace(b.FirmName),
		IsActive: true,
		Metadata: map[string]interface{}{
			"sap_firm_code": b.FirmCode,
		},
	}
}

type SAPProduct struct {
	ItemCode   string  `json:"item_code"`
	ItemName   string  `json:"item_name"`
	UserText   string  `json:"user_text,omitempty"`
	ItmsGrpCod int64   `json:"itms_grp_cod"`
	FirmCode   int64   `json:"firm_code"`
	InvntItem  string  `json:"invnt_item"`
	SellItem   string  `json:"sell_item"`
	PrchseItem string  `json:"prchse_item"`
	ValidFor   string  `json:"valid_for"`
	CodeBars   string  `json:"code_bars,omitempty"`
	BuyUnitMsr string  `json:"buy_unit_msr,omitempty"`
	SalUnitMsr string  `json:"sal_unit_msr,omitempty"`
	InvntryUom string  `json:"invntry_uom,omitempty"`
	NumInSale  float64 `json:"num_in_sale"`
	NumInBuy   float64 `json:"num_in_buy"`
	UgpEntry   int64   `json:"ugp_entry"`
	IUoMEntry  int64   `json:"i_uom_entry"`
	SUoMEntry  int64   `json:"s_uom_entry"`
	PUoMEntry  int64   `json:"p_uom_entry"`
	ManSerNum  string  `json:"man_ser_num"`
	ManBtchNum string  `json:"man_btch_num"`
	VatGourpSa string  `json:"vat_group_sa,omitempty"`
	AssetItem  string  `json:"asset_item,omitempty"`
	ItemType   string  `json:"item_type,omitempty"`
}

type CanonicalProduct struct {
	SKU                string                          `json:"sku"`
	Name               string                          `json:"name"`
	Description        string                          `json:"description,omitempty"`
	CategoryCode       string                          `json:"category_code,omitempty"`
	BrandCode          string                          `json:"brand_code,omitempty"`
	UOMCode            string                          `json:"uom_code,omitempty"`
	BaseUOMCode        string                          `json:"base_uom_code,omitempty"`
	SalesUOMCode       string                          `json:"sales_uom_code,omitempty"`
	PurchaseUOMCode    string                          `json:"purchase_uom_code,omitempty"`
	UOMGroupCode       string                          `json:"uom_group_code,omitempty"`
	SalesQtyPerBase    float64                         `json:"sales_qty_per_base"`
	PurchaseQtyPerBase float64                         `json:"purchase_qty_per_base"`
	ProductType        string                          `json:"product_type"`
	IsSerialized       bool                            `json:"is_serialized"`
	IsBatchManaged     bool                            `json:"is_batch_managed"`
	IsActive           bool                            `json:"is_active"`
	IsSellable         bool                            `json:"is_sellable"`
	IsPurchasable      bool                            `json:"is_purchasable"`
	TrackInventory     bool                            `json:"track_inventory"`
	PrimaryBarcode     string                          `json:"primary_barcode,omitempty"`
	UOMConversions     []CanonicalProductUOMConversion `json:"uom_conversions,omitempty"`
	Metadata           map[string]interface{}          `json:"metadata"`
}

func (p *SAPProduct) ToCanonical() CanonicalProduct {
	categoryCode := ""
	if p.ItmsGrpCod > 0 {
		categoryCode = fmt.Sprintf("CAT-%d", p.ItmsGrpCod)
	}
	brandCode := ""
	if p.FirmCode > 0 {
		brandCode = fmt.Sprintf("BRD-%d", p.FirmCode)
	}

	baseUom := strings.TrimSpace(p.InvntryUom)
	if baseUom == "" {
		baseUom = strings.TrimSpace(p.SalUnitMsr)
	}
	if baseUom == "" {
		baseUom = "UNIT"
	}

	salesUom := strings.TrimSpace(p.SalUnitMsr)
	if salesUom == "" {
		salesUom = baseUom
	}

	purchaseUom := strings.TrimSpace(p.BuyUnitMsr)
	if purchaseUom == "" {
		purchaseUom = baseUom
	}

	uomGroupCode := ""
	if p.UgpEntry > 0 {
		uomGroupCode = fmt.Sprintf("UGP-%d", p.UgpEntry)
	}

	numInSale := p.NumInSale
	if numInSale <= 0 {
		numInSale = 1.0
	}
	numInBuy := p.NumInBuy
	if numInBuy <= 0 {
		numInBuy = 1.0
	}

	var conversions []CanonicalProductUOMConversion
	sku := strings.TrimSpace(p.ItemCode)

	// If Sales UoM differs from Base UoM or has a conversion ratio
	if salesUom != baseUom || numInSale != 1.0 {
		conversions = append(conversions, CanonicalProductUOMConversion{
			ProductSKU:       sku,
			FromUOMCode:      salesUom,
			ToUOMCode:        baseUom,
			ConversionFactor: numInSale,
			IsDefault:        true,
			Metadata: map[string]interface{}{
				"source": "sap_sales_uom",
			},
		})
	}

	// If Purchase UoM differs from Base UoM and Sales UoM
	if (purchaseUom != baseUom || numInBuy != 1.0) && purchaseUom != salesUom {
		conversions = append(conversions, CanonicalProductUOMConversion{
			ProductSKU:       sku,
			FromUOMCode:      purchaseUom,
			ToUOMCode:        baseUom,
			ConversionFactor: numInBuy,
			IsDefault:        false,
			Metadata: map[string]interface{}{
				"source": "sap_purchase_uom",
			},
		})
	}

	return CanonicalProduct{
		SKU:                sku,
		Name:               strings.TrimSpace(p.ItemName),
		Description:        strings.TrimSpace(p.UserText),
		CategoryCode:       categoryCode,
		BrandCode:          brandCode,
		UOMCode:            baseUom,
		BaseUOMCode:        baseUom,
		SalesUOMCode:       salesUom,
		PurchaseUOMCode:    purchaseUom,
		UOMGroupCode:       uomGroupCode,
		SalesQtyPerBase:    numInSale,
		PurchaseQtyPerBase: numInBuy,
		ProductType: func() string {
			if SAPBool(p.AssetItem, false) || strings.ToUpper(strings.TrimSpace(p.ItemType)) == "F" {
				return "fixed_asset"
			}
			return "standard"
		}(),
		IsSerialized:   SAPBool(p.ManSerNum, false),
		IsBatchManaged: SAPBool(p.ManBtchNum, false),
		IsActive:       SAPBool(p.ValidFor, true),
		IsSellable:     SAPBool(p.SellItem, true),
		IsPurchasable:  SAPBool(p.PrchseItem, true),
		TrackInventory: SAPBool(p.InvntItem, true),
		PrimaryBarcode: strings.TrimSpace(p.CodeBars),
		UOMConversions: conversions,
		Metadata: map[string]interface{}{
			"sap_item_code":    p.ItemCode,
			"sap_vat_group":    p.VatGourpSa,
			"sap_asset_item":   p.AssetItem,
			"sap_item_type":    p.ItemType,
			"sap_num_in_sale":  p.NumInSale,
			"sap_num_in_buy":   p.NumInBuy,
			"sap_buy_unit_msr": p.BuyUnitMsr,
			"sap_sal_unit_msr": p.SalUnitMsr,
			"sap_invntry_uom":  p.InvntryUom,
			"sap_ugp_entry":    p.UgpEntry,
			"sap_i_uom_entry":  p.IUoMEntry,
			"sap_s_uom_entry":  p.SUoMEntry,
			"sap_p_uom_entry":  p.PUoMEntry,
		},
	}
}

type SAPBarcode struct {
	BcdEntry int64  `json:"bcd_entry"`
	BcdCode  string `json:"bcd_code"`
	ItemCode string `json:"item_code"`
	UomEntry int64  `json:"uom_entry"`
	UomCode  string `json:"uom_code,omitempty"`
}

type CanonicalBarcode struct {
	ProductSKU  string                 `json:"product_sku"`
	Barcode     string                 `json:"barcode"`
	BarcodeType string                 `json:"barcode_type"`
	UOMCode     string                 `json:"uom_code,omitempty"`
	IsPrimary   bool                   `json:"is_primary"`
	Metadata    map[string]interface{} `json:"metadata"`
}

func (b *SAPBarcode) ToCanonical(isPrimary bool) CanonicalBarcode {
	return CanonicalBarcode{
		ProductSKU:  strings.TrimSpace(b.ItemCode),
		Barcode:     strings.TrimSpace(b.BcdCode),
		BarcodeType: "EAN13",
		UOMCode:     strings.TrimSpace(b.UomCode),
		IsPrimary:   isPrimary,
		Metadata: map[string]interface{}{
			"sap_bcd_entry": b.BcdEntry,
			"sap_uom_entry": b.UomEntry,
		},
	}
}

// ----------------------------------------------------
// Domain 4: Inventory Balances (OITW)
// ----------------------------------------------------

type SAPInventoryStock struct {
	ItemCode   string  `json:"item_code"`
	WhsCode    string  `json:"whs_code"`
	OnHand     float64 `json:"on_hand"`
	IsCommited float64 `json:"is_commited"`
	OnOrder    float64 `json:"on_order"`
	MinStock   float64 `json:"min_stock"`
	MaxStock   float64 `json:"max_stock"`
}

type CanonicalInventoryStock struct {
	ProductSKU        string                 `json:"product_sku"`
	StoreCode         string                 `json:"store_code"`
	QuantityOnHand    float64                `json:"quantity_on_hand"`
	QuantityAllocated float64                `json:"quantity_allocated"`
	QuantityAvailable float64                `json:"quantity_available"`
	QuantityOnOrder   float64                `json:"quantity_on_order"`
	ReorderLevel      float64                `json:"reorder_level"`
	MaxStockLevel     float64                `json:"max_stock_level"`
	Metadata          map[string]interface{} `json:"metadata"`
}

func (inv *SAPInventoryStock) ToCanonical() CanonicalInventoryStock {
	available := math.Max(0, inv.OnHand-inv.IsCommited)

	return CanonicalInventoryStock{
		ProductSKU:        strings.TrimSpace(inv.ItemCode),
		StoreCode:         strings.TrimSpace(inv.WhsCode),
		QuantityOnHand:    inv.OnHand,
		QuantityAllocated: inv.IsCommited,
		QuantityAvailable: available,
		QuantityOnOrder:   inv.OnOrder,
		ReorderLevel:      inv.MinStock,
		MaxStockLevel:     inv.MaxStock,
		Metadata: map[string]interface{}{
			"sap_item_code": inv.ItemCode,
			"sap_whs_code":  inv.WhsCode,
		},
	}
}

// ----------------------------------------------------
// Domain 5: Business Partners (OCRD)
// ----------------------------------------------------

type SAPBusinessPartner struct {
	CardCode    string  `json:"card_code"`
	CardName    string  `json:"card_name"`
	CardType    string  `json:"card_type"` // 'C' = Customer, 'S' = Supplier, 'L' = Lead
	LicTradNum  string  `json:"lic_trad_num,omitempty"`
	Phone1      string  `json:"phone1,omitempty"`
	EMail       string  `json:"e_mail,omitempty"`
	Currency    string  `json:"currency,omitempty"`
	ValidFor    string  `json:"valid_for"`
	Balance     float64 `json:"balance"`
	CreditLimit float64 `json:"credit_limit"` // A19: feeds business_partners.credit_limit
	GroupNum    int64   `json:"group_num"`    // A23: OCRD.GroupNum → OCTG payment term
}

type CanonicalPartner struct {
	PartnerType     string                 `json:"partner_type"` // "customer", "supplier" or "lead"
	Code            string                 `json:"code"`
	Name            string                 `json:"name"`
	Email           string                 `json:"email,omitempty"`
	Phone           string                 `json:"phone,omitempty"`
	TaxID           string                 `json:"tax_id,omitempty"`
	CurrencyCode    string                 `json:"currency_code"`
	IsActive        bool                   `json:"is_active"`
	Balance         float64                `json:"balance"`
	CreditLimit     float64                `json:"credit_limit"`
	PaymentTermCode string                 `json:"payment_term_code,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
}

func (bp *SAPBusinessPartner) ToCanonical() CanonicalPartner {
	// A34.1: CardType 'L' (Lead) is preserved instead of defaulting to customer.
	partnerType := "customer"
	switch strings.ToUpper(bp.CardType) {
	case "S":
		partnerType = "supplier"
	case "L":
		partnerType = "lead"
	}

	// A23: OCRD.GroupNum resolves to payment_terms code "PT-{GroupNum}".
	paymentTermCode := ""
	if bp.GroupNum > 0 {
		paymentTermCode = fmt.Sprintf("PT-%d", bp.GroupNum)
	}

	return CanonicalPartner{
		PartnerType:     partnerType,
		Code:            strings.TrimSpace(bp.CardCode),
		Name:            strings.TrimSpace(bp.CardName),
		Email:           strings.TrimSpace(bp.EMail),
		Phone:           strings.TrimSpace(bp.Phone1),
		TaxID:           strings.TrimSpace(bp.LicTradNum),
		CurrencyCode:    strings.TrimSpace(bp.Currency),
		IsActive:        SAPBool(bp.ValidFor, true),
		Balance:         bp.Balance,
		CreditLimit:     bp.CreditLimit,
		PaymentTermCode: paymentTermCode,
		Metadata: map[string]interface{}{
			"sap_card_code": bp.CardCode,
			"sap_card_type": bp.CardType,
			"sap_group_num": bp.GroupNum,
		},
	}
}

// ----------------------------------------------------
// Domain 6: Sales Orders & Invoices (ORDR/RDR1, OINV/INV1)
// ----------------------------------------------------

type SAPSalesOrderLine struct {
	DocEntry   int64   `json:"doc_entry"`
	LineNum    int     `json:"line_num"`
	ItemCode   string  `json:"item_code"`
	Dscription string  `json:"dscription"`
	Quantity   float64 `json:"quantity"`
	Price      float64 `json:"price"`
	LineTotal  float64 `json:"line_total"`
	VatSum     float64 `json:"vat_sum"`
	WhsCode    string  `json:"whs_code"`
	UnitMsr    string  `json:"unit_msr"`
}

type SAPSalesOrder struct {
	DocEntry   int64               `json:"doc_entry"`
	DocNum     int64               `json:"doc_num"`
	DocDate    time.Time           `json:"doc_date"`
	DocDueDate time.Time           `json:"doc_due_date"`
	CardCode   string              `json:"card_code"`
	CardName   string              `json:"card_name"`
	DocTotal   float64             `json:"doc_total"`
	VatSum     float64             `json:"vat_sum"`
	DiscSum    float64             `json:"disc_sum"`
	DocStatus  string              `json:"doc_status"` // 'O'=Open, 'C'=Closed
	SlpCode    int64               `json:"slp_code"`
	Comments   string              `json:"comments,omitempty"`
	Lines      []SAPSalesOrderLine `json:"lines,omitempty"`
}

type CanonicalSalesOrderLine struct {
	LineNumber   int                    `json:"line_number"`
	ProductSKU   string                 `json:"product_sku"`
	ProductName  string                 `json:"product_name"`
	StoreCode    string                 `json:"store_code,omitempty"`
	Quantity     float64                `json:"quantity"`
	UnitPrice    float64                `json:"unit_price"`
	LineSubtotal float64                `json:"line_subtotal"`
	TaxAmount    float64                `json:"tax_amount"`
	LineTotal    float64                `json:"line_total"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type CanonicalSalesOrder struct {
	OrderNumber  string `json:"order_number"`
	CustomerCode string `json:"customer_code"`
	CustomerName string `json:"customer_name"`
	// A18: primary warehouse (first non-empty RDR1.WhsCode) feeding
	// sales_orders_v2.store_id.
	StoreCode         string                    `json:"store_code,omitempty"`
	OrderStatus       string                    `json:"order_status"`
	PaymentStatus     string                    `json:"payment_status"`
	FulfillmentStatus string                    `json:"fulfillment_status"`
	OrderDate         time.Time                 `json:"order_date"`
	ExpectedDate      time.Time                 `json:"expected_date"`
	Subtotal          float64                   `json:"subtotal"`
	DiscountAmount    float64                   `json:"discount_amount"`
	TaxAmount         float64                   `json:"tax_amount"`
	TotalAmount       float64                   `json:"total_amount"`
	Notes             string                    `json:"notes,omitempty"`
	Lines             []CanonicalSalesOrderLine `json:"lines"`
	Metadata          map[string]interface{}    `json:"metadata"`
}

func (so *SAPSalesOrder) ToCanonical() CanonicalSalesOrder {
	orderStatus := "fulfilled"
	paymentStatus := "paid"
	fulfillmentStatus := "fulfilled"
	if strings.ToUpper(so.DocStatus) == "O" {
		orderStatus = "confirmed"
		paymentStatus = "unpaid"
		fulfillmentStatus = "unfulfilled"
	}

	lines := make([]CanonicalSalesOrderLine, len(so.Lines))
	primaryStore := ""
	for i, l := range so.Lines {
		whs := strings.TrimSpace(l.WhsCode)
		if primaryStore == "" && whs != "" {
			primaryStore = whs
		}
		lines[i] = CanonicalSalesOrderLine{
			LineNumber:   l.LineNum,
			ProductSKU:   strings.TrimSpace(l.ItemCode),
			ProductName:  strings.TrimSpace(l.Dscription),
			StoreCode:    whs,
			Quantity:     l.Quantity,
			UnitPrice:    l.Price,
			LineSubtotal: l.LineTotal,
			TaxAmount:    l.VatSum,
			LineTotal:    l.LineTotal + l.VatSum,
			Metadata: map[string]interface{}{
				"sap_doc_entry": l.DocEntry,
				"sap_line_num":  l.LineNum,
				"sap_unit_msr":  l.UnitMsr,
				"sap_whs_code":  l.WhsCode,
			},
		}
	}

	subtotal := so.DocTotal - so.VatSum + so.DiscSum

	return CanonicalSalesOrder{
		OrderNumber:       fmt.Sprintf("SO-SAP-%d", so.DocNum),
		CustomerCode:      strings.TrimSpace(so.CardCode),
		CustomerName:      strings.TrimSpace(so.CardName),
		StoreCode:         primaryStore,
		OrderStatus:       orderStatus,
		PaymentStatus:     paymentStatus,
		FulfillmentStatus: fulfillmentStatus,
		OrderDate:         so.DocDate,
		ExpectedDate:      so.DocDueDate,
		Subtotal:          subtotal,
		DiscountAmount:    so.DiscSum,
		TaxAmount:         so.VatSum,
		TotalAmount:       so.DocTotal,
		Notes:             strings.TrimSpace(so.Comments),
		Lines:             lines,
		Metadata: map[string]interface{}{
			"sap_doc_entry": so.DocEntry,
			"sap_doc_num":   so.DocNum,
			"sap_slp_code":  so.SlpCode,
		},
	}
}

type SAPInvoiceLine struct {
	DocEntry   int64   `json:"doc_entry"`
	LineNum    int     `json:"line_num"`
	ItemCode   string  `json:"item_code"`
	Dscription string  `json:"dscription"`
	Quantity   float64 `json:"quantity"`
	Price      float64 `json:"price"`
	LineTotal  float64 `json:"line_total"`
	VatSum     float64 `json:"vat_sum"`
	WhsCode    string  `json:"whs_code"`
	UnitMsr    string  `json:"unit_msr"`
	// A01: line discount evidence
	DiscPrcnt  float64 `json:"disc_prcnt"`
	PriceBefDi float64 `json:"price_bef_di"`
}

type SAPInvoice struct {
	DocEntry   int64     `json:"doc_entry"`
	DocNum     int64     `json:"doc_num"`
	DocDate    time.Time `json:"doc_date"`
	DocDueDate time.Time `json:"doc_due_date"`
	CardCode   string    `json:"card_code"`
	CardName   string    `json:"card_name"`
	DocTotal   float64   `json:"doc_total"`
	PaidToDate float64   `json:"paid_to_date"`
	VatSum     float64   `json:"vat_sum"`
	DiscSum    float64   `json:"disc_sum"`
	DocStatus  string    `json:"doc_status"` // 'O'=Open, 'C'=Closed
	SlpCode    int64     `json:"slp_code"`
	Comments   string    `json:"comments,omitempty"`
	// A02: source currency
	DocCur  string  `json:"doc_cur"`
	DocRate float64 `json:"doc_rate"`
	// A03: cancellation flag
	Canceled string `json:"canceled"` // 'Y', 'C', or 'N'
	// A04: document type
	DocType string           `json:"doc_type"` // 'I'=standard, 'S'=service
	Lines   []SAPInvoiceLine `json:"lines,omitempty"`
}

type CanonicalInvoiceLine struct {
	LineNumber          int                    `json:"line_number"`
	ProductSKU          string                 `json:"product_sku"`
	ProductName         string                 `json:"product_name"`
	ItemType            string                 `json:"item_type"`            // A04: 'product' or 'service'
	StoreCode           string                 `json:"store_code,omitempty"` // A17: INV1.WhsCode
	UOMCode             string                 `json:"uom_code"`             // A06: source UOM code
	Quantity            float64                `json:"quantity"`
	UnitPrice           float64                `json:"unit_price"`
	DiscountPercent     float64                `json:"discount_percent"`      // A01: raw percentage
	DiscountAmount      float64                `json:"discount_amount"`       // A01: monetary discount
	PriceBeforeDiscount float64                `json:"price_before_discount"` // A01: pre-discount price
	LineSubtotal        float64                `json:"line_subtotal"`
	TaxAmount           float64                `json:"tax_amount"`
	LineTotal           float64                `json:"line_total"`
	Metadata            map[string]interface{} `json:"metadata"`
}

type CanonicalInvoice struct {
	InvoiceNumber string `json:"invoice_number"`
	CustomerCode  string `json:"customer_code"`
	CustomerName  string `json:"customer_name"`
	InvoiceType   string `json:"invoice_type"`
	InvoiceStatus string `json:"invoice_status"`
	// A17: primary warehouse for the invoice (first non-empty INV1.WhsCode);
	// feeds invoices.store_id so the migrated invoices appear in
	// vw_realtime_pnl and vw_tax_vat_summary.
	StoreCode string `json:"store_code,omitempty"`
	// A02: source currency
	CurrencyCode   string                 `json:"currency_code"`
	ExchangeRate   float64                `json:"exchange_rate"`
	InvoiceDate    time.Time              `json:"invoice_date"`
	DueDate        time.Time              `json:"due_date"`
	Subtotal       float64                `json:"subtotal"`
	DiscountAmount float64                `json:"discount_amount"`
	TaxAmount      float64                `json:"tax_amount"`
	TotalAmount    float64                `json:"total_amount"`
	PaidAmount     float64                `json:"paid_amount"`
	BalanceDue     float64                `json:"balance_due"`
	Lines          []CanonicalInvoiceLine `json:"lines"`
	Metadata       map[string]interface{} `json:"metadata"`
}

func (inv *SAPInvoice) ToCanonical() CanonicalInvoice {
	// A03: check CANCELED flag first — takes precedence over balance
	invoiceStatus := "paid"
	canceled := strings.ToUpper(strings.TrimSpace(inv.Canceled))
	if canceled == "Y" || canceled == "C" {
		invoiceStatus = "cancelled"
	} else {
		balanceDue := inv.DocTotal - inv.PaidToDate
		if balanceDue > 0.01 {
			if inv.PaidToDate > 0.01 {
				invoiceStatus = "partially_paid"
			} else {
				invoiceStatus = "sent"
			}
		}
	}

	// A04 + A26: invoice_type is an enum column ('standard','proforma','credit_note',
	// 'debit_note','recurring') — the source DocType is preserved in metadata and the
	// line-level item_type ('service') carries the service distinction. DocType 'S'
	// must NOT be emitted as InvoiceType "service" (invalid enum value).
	invoiceType := "standard"

	// A02: ensure currency is populated; fallback to SAR
	currencyCode := strings.TrimSpace(inv.DocCur)
	if currencyCode == "" {
		currencyCode = "SAR"
	}
	exchangeRate := inv.DocRate
	if exchangeRate == 0 {
		exchangeRate = 1.0
	}

	lines := make([]CanonicalInvoiceLine, len(inv.Lines))
	primaryStore := ""
	for i, l := range inv.Lines {
		// A04: service-line item_type
		itemType := "product"
		if strings.ToUpper(strings.TrimSpace(inv.DocType)) == "S" && strings.TrimSpace(l.ItemCode) == "" {
			itemType = "service"
		}
		// A17: per-line warehouse; first non-empty wins for the header store
		whs := strings.TrimSpace(l.WhsCode)
		if primaryStore == "" && whs != "" {
			primaryStore = whs
		}
		// A01: derive monetary discount (Quantity * PriceBefDi * DiscPrcnt / 100)
		// Stored unsigned; sign policy to be confirmed by finance.
		discountAmount := 0.0
		if l.DiscPrcnt != 0 && l.PriceBefDi != 0 {
			discountAmount = math.Round(l.Quantity*l.PriceBefDi*(l.DiscPrcnt/100.0)*100) / 100
		}
		lines[i] = CanonicalInvoiceLine{
			LineNumber:          l.LineNum,
			ProductSKU:          strings.TrimSpace(l.ItemCode),
			ProductName:         strings.TrimSpace(l.Dscription),
			ItemType:            itemType,
			StoreCode:           whs,
			UOMCode:             strings.TrimSpace(l.UnitMsr),
			Quantity:            l.Quantity,
			UnitPrice:           l.Price,
			DiscountPercent:     l.DiscPrcnt,
			DiscountAmount:      discountAmount,
			PriceBeforeDiscount: l.PriceBefDi,
			LineSubtotal:        l.LineTotal,
			TaxAmount:           l.VatSum,
			LineTotal:           l.LineTotal + l.VatSum,
			Metadata: map[string]interface{}{
				"sap_doc_entry":    l.DocEntry,
				"sap_line_num":     l.LineNum,
				"sap_unit_msr":     l.UnitMsr,
				"sap_whs_code":     l.WhsCode,
				"sap_disc_prcnt":   l.DiscPrcnt,
				"sap_price_bef_di": l.PriceBefDi,
			},
		}
	}

	balanceDue := inv.DocTotal - inv.PaidToDate
	subtotal := inv.DocTotal - inv.VatSum + inv.DiscSum

	return CanonicalInvoice{
		InvoiceNumber:  fmt.Sprintf("INV-SAP-%d", inv.DocNum),
		CustomerCode:   strings.TrimSpace(inv.CardCode),
		CustomerName:   strings.TrimSpace(inv.CardName),
		InvoiceType:    invoiceType,
		InvoiceStatus:  invoiceStatus,
		StoreCode:      primaryStore,
		CurrencyCode:   currencyCode,
		ExchangeRate:   exchangeRate,
		InvoiceDate:    inv.DocDate,
		DueDate:        inv.DocDueDate,
		Subtotal:       subtotal,
		DiscountAmount: inv.DiscSum,
		TaxAmount:      inv.VatSum,
		TotalAmount:    inv.DocTotal,
		PaidAmount:     inv.PaidToDate,
		BalanceDue:     math.Max(0, balanceDue),
		Lines:          lines,
		Metadata: map[string]interface{}{
			"sap_doc_entry": inv.DocEntry,
			"sap_doc_num":   inv.DocNum,
			"sap_doc_cur":   inv.DocCur,
			"sap_doc_rate":  inv.DocRate,
			"sap_canceled":  inv.Canceled,
			"sap_doc_type":  inv.DocType,
			"sap_slp_code":  inv.SlpCode,
			"sap_comments":  inv.Comments,
		},
	}
}

// ----------------------------------------------------
// Domain 7: Price Lists (OPLN, ITM1)
// ----------------------------------------------------

type SAPPriceList struct {
	ListNum  int64   `json:"list_num"`
	ListName string  `json:"list_name"`
	Currency string  `json:"currency"`
	Factor   float64 `json:"factor"`
	BasedOn  int64   `json:"based_on"`
	ValidFor string  `json:"valid_for"`
}

type CanonicalPriceList struct {
	Code         string                 `json:"code"`
	Name         string                 `json:"name"`
	CurrencyCode string                 `json:"currency_code"`
	Factor       float64                `json:"factor"`
	BasedOnCode  string                 `json:"based_on_code,omitempty"`
	IsActive     bool                   `json:"is_active"`
	Metadata     map[string]interface{} `json:"metadata"`
}

func (pl *SAPPriceList) ToCanonical() CanonicalPriceList {
	basedOnCode := ""
	if pl.BasedOn > 0 {
		basedOnCode = fmt.Sprintf("PL-%d", pl.BasedOn)
	}
	return CanonicalPriceList{
		Code:         fmt.Sprintf("PL-%d", pl.ListNum),
		Name:         strings.TrimSpace(pl.ListName),
		CurrencyCode: strings.TrimSpace(pl.Currency),
		Factor:       pl.Factor,
		BasedOnCode:  basedOnCode,
		IsActive:     SAPBool(pl.ValidFor, true),
		Metadata: map[string]interface{}{
			"sap_list_num": pl.ListNum,
		},
	}
}

type SAPPriceListItem struct {
	ItemCode  string  `json:"item_code"`
	PriceList int64   `json:"price_list"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	UomEntry  int64   `json:"uom_entry"`
	UomCode   string  `json:"uom_code,omitempty"`
}

type CanonicalPriceListItem struct {
	PriceListCode string                 `json:"price_list_code"`
	ProductSKU    string                 `json:"product_sku"`
	UOMCode       string                 `json:"uom_code,omitempty"`
	Price         float64                `json:"price"`
	CurrencyCode  string                 `json:"currency_code"`
	Metadata      map[string]interface{} `json:"metadata"`
}

func (item *SAPPriceListItem) ToCanonical() CanonicalPriceListItem {
	return CanonicalPriceListItem{
		PriceListCode: fmt.Sprintf("PL-%d", item.PriceList),
		ProductSKU:    strings.TrimSpace(item.ItemCode),
		UOMCode:       strings.TrimSpace(item.UomCode),
		Price:         item.Price,
		CurrencyCode:  strings.TrimSpace(item.Currency),
		Metadata: map[string]interface{}{
			"sap_uom_entry": item.UomEntry,
		},
	}
}

// ----------------------------------------------------
// Domain 8: Business Partner Addresses (CRD1)
// ----------------------------------------------------

type SAPBPAddress struct {
	CardCode  string `json:"card_code"`
	AdresType string `json:"adres_type"` // 'B'=Bill-To, 'S'=Ship-To
	Address   string `json:"address"`
	Street    string `json:"street"`
	City      string `json:"city"`
	Country   string `json:"country"`
	ZipCode   string `json:"zip_code"`
	State     string `json:"state"`
	Phone1    string `json:"phone1"`
	Phone2    string `json:"phone2"`
}

type CanonicalBPAddress struct {
	PartnerCode string                 `json:"partner_code"`
	AddressType string                 `json:"address_type"` // "billing" or "shipping"
	AddressLine string                 `json:"address_line"`
	Street      string                 `json:"street"`
	City        string                 `json:"city"`
	Country     string                 `json:"country"`
	PostalCode  string                 `json:"postal_code"`
	State       string                 `json:"state"`
	Phone       string                 `json:"phone"`
	Metadata    map[string]interface{} `json:"metadata"`
}

func (addr *SAPBPAddress) ToCanonical() CanonicalBPAddress {
	addrType := "shipping"
	if strings.ToUpper(addr.AdresType) == "B" {
		addrType = "billing"
	}
	phone := strings.TrimSpace(addr.Phone1)
	if phone == "" {
		phone = strings.TrimSpace(addr.Phone2)
	}
	return CanonicalBPAddress{
		PartnerCode: strings.TrimSpace(addr.CardCode),
		AddressType: addrType,
		AddressLine: strings.TrimSpace(addr.Address),
		Street:      strings.TrimSpace(addr.Street),
		City:        strings.TrimSpace(addr.City),
		Country:     strings.TrimSpace(addr.Country),
		PostalCode:  strings.TrimSpace(addr.ZipCode),
		State:       strings.TrimSpace(addr.State),
		Phone:       phone,
		Metadata: map[string]interface{}{
			"sap_card_code":  addr.CardCode,
			"sap_adres_type": addr.AdresType,
		},
	}
}

// ----------------------------------------------------
// Domain 8: Purchase Orders (OPOR / POR1)
// ----------------------------------------------------

type SAPPurchaseOrderLine struct {
	DocEntry   int64   `json:"doc_entry"`
	LineNum    int     `json:"line_num"`
	ItemCode   string  `json:"item_code"`
	Dscription string  `json:"dscription"`
	Quantity   float64 `json:"quantity"`
	Price      float64 `json:"price"`
	LineTotal  float64 `json:"line_total"`
	VatSum     float64 `json:"vat_sum"`
	WhsCode    string  `json:"whs_code"`
	UnitMsr    string  `json:"unit_msr"`
	OpenQty    float64 `json:"open_qty"`
}

type SAPPurchaseOrder struct {
	DocEntry   int64                  `json:"doc_entry"`
	DocNum     int64                  `json:"doc_num"`
	DocDate    time.Time              `json:"doc_date"`
	DocDueDate time.Time              `json:"doc_due_date"`
	CardCode   string                 `json:"card_code"`
	CardName   string                 `json:"card_name"`
	DocTotal   float64                `json:"doc_total"`
	VatSum     float64                `json:"vat_sum"`
	DiscSum    float64                `json:"disc_sum"`
	DocStatus  string                 `json:"doc_status"` // 'O' = Open, 'C' = Closed
	SlpCode    int64                  `json:"slp_code"`
	Comments   string                 `json:"comments,omitempty"`
	Lines      []SAPPurchaseOrderLine `json:"lines,omitempty"`
}

type CanonicalPurchaseOrderLine struct {
	LineNumber       int                    `json:"line_number"`
	ProductSKU       string                 `json:"product_sku"`
	ProductName      string                 `json:"product_name"`
	StoreCode        string                 `json:"store_code,omitempty"`
	Quantity         float64                `json:"quantity"`
	ReceivedQuantity float64                `json:"received_quantity"`
	UOMCode          string                 `json:"uom_code"`
	UnitPrice        float64                `json:"unit_price"`
	DiscountAmount   float64                `json:"discount_amount"`
	TaxAmount        float64                `json:"tax_amount"`
	Subtotal         float64                `json:"subtotal"`
	LineTotal        float64                `json:"line_total"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type CanonicalPurchaseOrder struct {
	PONumber             string                       `json:"po_number"`
	SupplierCode         string                       `json:"supplier_code"`
	SupplierName         string                       `json:"supplier_name"`
	StoreCode            string                       `json:"store_code"`
	PODate               time.Time                    `json:"po_date"`
	ExpectedDeliveryDate *time.Time                   `json:"expected_delivery_date,omitempty"`
	Status               string                       `json:"status"` // 'draft', 'sent', 'partially_received', 'received', 'cancelled'
	Subtotal             float64                      `json:"subtotal"`
	DiscountAmount       float64                      `json:"discount_amount"`
	TaxAmount            float64                      `json:"tax_amount"`
	TotalAmount          float64                      `json:"total_amount"`
	Notes                string                       `json:"notes,omitempty"`
	Lines                []CanonicalPurchaseOrderLine `json:"lines"`
	Metadata             map[string]interface{}       `json:"metadata"`
}

func (po *SAPPurchaseOrder) ToCanonical() CanonicalPurchaseOrder {
	status := "received"
	if strings.ToUpper(po.DocStatus) == "O" {
		status = "sent"
		// Check if any line has been partially received
		for _, l := range po.Lines {
			if l.OpenQty < l.Quantity && l.OpenQty > 0 {
				status = "partially_received"
				break
			}
		}
	}

	primaryStore := ""
	lines := make([]CanonicalPurchaseOrderLine, len(po.Lines))
	for i, l := range po.Lines {
		whs := strings.TrimSpace(l.WhsCode)
		if primaryStore == "" && whs != "" {
			primaryStore = whs
		}
		receivedQty := l.Quantity - l.OpenQty
		if receivedQty < 0 {
			receivedQty = 0
		}
		lines[i] = CanonicalPurchaseOrderLine{
			LineNumber:       l.LineNum,
			ProductSKU:       strings.TrimSpace(l.ItemCode),
			ProductName:      strings.TrimSpace(l.Dscription),
			StoreCode:        whs,
			Quantity:         l.Quantity,
			ReceivedQuantity: receivedQty,
			UOMCode:          strings.TrimSpace(l.UnitMsr),
			UnitPrice:        l.Price,
			DiscountAmount:   0,
			TaxAmount:        l.VatSum,
			Subtotal:         l.LineTotal,
			LineTotal:        l.LineTotal + l.VatSum,
			Metadata: map[string]interface{}{
				"sap_doc_entry": po.DocEntry,
				"sap_line_num":  l.LineNum,
				"sap_open_qty":  l.OpenQty,
			},
		}
	}

	subtotal := po.DocTotal - po.VatSum + po.DiscSum
	if subtotal < 0 {
		subtotal = 0
	}
	// A34.2: aligned with the invoice formula (DocTotal − VatSum + DiscSum =
	// pre-discount, pre-VAT net amount). The previous "− DiscSum" understated
	// the subtotal by 2×DiscSum whenever a header discount existed.

	var expDate *time.Time
	if !po.DocDueDate.IsZero() {
		expDate = &po.DocDueDate
	}

	return CanonicalPurchaseOrder{
		PONumber:             fmt.Sprintf("PO-%d", po.DocNum),
		SupplierCode:         strings.TrimSpace(po.CardCode),
		SupplierName:         strings.TrimSpace(po.CardName),
		StoreCode:            primaryStore,
		PODate:               po.DocDate,
		ExpectedDeliveryDate: expDate,
		Status:               status,
		Subtotal:             subtotal,
		DiscountAmount:       po.DiscSum,
		TaxAmount:            po.VatSum,
		TotalAmount:          po.DocTotal,
		Notes:                strings.TrimSpace(po.Comments),
		Lines:                lines,
		Metadata: map[string]interface{}{
			"sap_doc_entry": po.DocEntry,
			"sap_doc_num":   po.DocNum,
			"sap_card_code": po.CardCode,
			"sap_status":    po.DocStatus,
		},
	}
}

// ----------------------------------------------------
// Domain 9: Goods Receipt Notes (OPDN / PDN1)
// ----------------------------------------------------

type SAPGoodsReceiptLine struct {
	DocEntry   int64   `json:"doc_entry"`
	LineNum    int     `json:"line_num"`
	ItemCode   string  `json:"item_code"`
	Dscription string  `json:"dscription"`
	Quantity   float64 `json:"quantity"`
	Price      float64 `json:"price"`
	LineTotal  float64 `json:"line_total"`
	VatSum     float64 `json:"vat_sum"`
	WhsCode    string  `json:"whs_code"`
	UnitMsr    string  `json:"unit_msr"`
	BaseEntry  int64   `json:"base_entry"` // Source PO DocEntry
	BaseLine   int     `json:"base_line"`  // Source PO LineNum
	BaseType   int     `json:"base_type"`  // 22 = Purchase Order
}

type SAPGoodsReceipt struct {
	DocEntry   int64                 `json:"doc_entry"`
	DocNum     int64                 `json:"doc_num"`
	DocDate    time.Time             `json:"doc_date"`
	DocDueDate time.Time             `json:"doc_due_date"`
	CardCode   string                `json:"card_code"`
	CardName   string                `json:"card_name"`
	DocTotal   float64               `json:"doc_total"`
	VatSum     float64               `json:"vat_sum"`
	DocStatus  string                `json:"doc_status"` // 'O'=Open, 'C'=Closed
	Canceled   string                `json:"canceled"`   // A34: 'Y', 'C', or 'N'
	Comments   string                `json:"comments,omitempty"`
	Lines      []SAPGoodsReceiptLine `json:"lines,omitempty"`
}

type CanonicalGoodsReceiptItem struct {
	LineNumber       int                    `json:"line_number"`
	ProductSKU       string                 `json:"product_sku"`
	ProductName      string                 `json:"product_name"`
	StoreCode        string                 `json:"store_code,omitempty"`
	QuantityReceived float64                `json:"quantity_received"`
	QuantityRejected float64                `json:"quantity_rejected"`
	UOMCode          string                 `json:"uom_code"`
	UnitCost         float64                `json:"unit_cost"`
	SourcePOLineNum  int                    `json:"source_po_line_num,omitempty"`
	Notes            string                 `json:"notes,omitempty"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type CanonicalGoodsReceiptNote struct {
	GRNNumber          string                      `json:"grn_number"`
	PONumber           string                      `json:"po_number,omitempty"`
	SupplierCode       string                      `json:"supplier_code"`
	SupplierName       string                      `json:"supplier_name"`
	StoreCode          string                      `json:"store_code"`
	ReceiptDate        time.Time                   `json:"receipt_date"`
	DeliveryNoteNumber string                      `json:"delivery_note_number,omitempty"`
	Status             string                      `json:"status"` // 'draft', 'posted', 'cancelled'
	Notes              string                      `json:"notes,omitempty"`
	Items              []CanonicalGoodsReceiptItem `json:"items"`
	Metadata           map[string]interface{}      `json:"metadata"`
}

func (gr *SAPGoodsReceipt) ToCanonical() CanonicalGoodsReceiptNote {
	// A34.3: a canceled GRP (CANCELED 'Y'/'C') must not import as a posted receipt;
	// DocStatus 'C' (fully closed/processed) still maps to 'posted'.
	status := "posted"
	if gr.Canceled == "Y" || strings.ToUpper(strings.TrimSpace(gr.Canceled)) == "C" {
		status = "cancelled"
	}

	primaryStore := ""
	var sourcePONum string
	items := make([]CanonicalGoodsReceiptItem, len(gr.Lines))
	for i, l := range gr.Lines {
		whs := strings.TrimSpace(l.WhsCode)
		if primaryStore == "" && whs != "" {
			primaryStore = whs
		}
		if sourcePONum == "" && l.BaseEntry > 0 && l.BaseType == 22 {
			// A08: store the SAP DocEntry of the source PO, NOT the DocNum.
			// The PO lookup in sap_migration.go uses metadata->>'sap_doc_entry'
			// so we format as the raw DocEntry integer string, not "PO-{DocNum}".
			sourcePONum = fmt.Sprintf("%d", l.BaseEntry)
		}
		items[i] = CanonicalGoodsReceiptItem{
			LineNumber:       l.LineNum,
			ProductSKU:       strings.TrimSpace(l.ItemCode),
			ProductName:      strings.TrimSpace(l.Dscription),
			StoreCode:        whs,
			QuantityReceived: l.Quantity,
			QuantityRejected: 0,
			UOMCode:          strings.TrimSpace(l.UnitMsr),
			UnitCost:         l.Price,
			SourcePOLineNum:  l.BaseLine,
			Metadata: map[string]interface{}{
				"sap_doc_entry":  gr.DocEntry,
				"sap_line_num":   l.LineNum,
				"sap_base_entry": l.BaseEntry,
				"sap_base_line":  l.BaseLine,
				"sap_base_type":  l.BaseType,
			},
		}
	}

	return CanonicalGoodsReceiptNote{
		GRNNumber:          fmt.Sprintf("GRN-%d", gr.DocNum),
		PONumber:           sourcePONum,
		SupplierCode:       strings.TrimSpace(gr.CardCode),
		SupplierName:       strings.TrimSpace(gr.CardName),
		StoreCode:          primaryStore,
		ReceiptDate:        gr.DocDate,
		DeliveryNoteNumber: fmt.Sprintf("SAP-GRN-%d", gr.DocNum),
		Status:             status,
		Notes:              strings.TrimSpace(gr.Comments),
		Items:              items,
		Metadata: map[string]interface{}{
			"sap_doc_entry": gr.DocEntry,
			"sap_doc_num":   gr.DocNum,
			"sap_card_code": gr.CardCode,
			"sap_doc_total": gr.DocTotal,
			"sap_canceled":  gr.Canceled,
		},
	}
}

// ----------------------------------------------------
// Domain 10: Stock Movements (OINM Item Ledger)
// ----------------------------------------------------

type SAPStockMovement struct {
	TransNum   int64     `json:"trans_num"`
	TransType  int       `json:"trans_type"`
	CreatedBy  int64     `json:"created_by"`
	BaseRef    string    `json:"base_ref"`
	DocDate    time.Time `json:"doc_date"`
	ItemCode   string    `json:"item_code"`
	Warehouse  string    `json:"warehouse"`
	InQty      float64   `json:"in_qty"`
	OutQty     float64   `json:"out_qty"`
	Price      float64   `json:"price"`
	TransValue float64   `json:"trans_value"`
}

type CanonicalStockMovement struct {
	MovementType    string                 `json:"movement_type"`
	ReferenceType   string                 `json:"reference_type"`
	ReferenceNumber string                 `json:"reference_number"`
	ProductSKU      string                 `json:"product_sku"`
	FromStoreCode   string                 `json:"from_store_code,omitempty"`
	ToStoreCode     string                 `json:"to_store_code,omitempty"`
	Quantity        float64                `json:"quantity"`
	UOMCode         string                 `json:"uom_code,omitempty"`
	MovementDate    time.Time              `json:"movement_date"`
	CostPerUnit     float64                `json:"cost_per_unit"`
	TotalValue      float64                `json:"total_value"`
	Metadata        map[string]interface{} `json:"metadata"`
}

func (sm *SAPStockMovement) ToCanonical() CanonicalStockMovement {
	movementType := "adjustment_positive"
	referenceType := "stock_count"
	referenceNum := sm.BaseRef
	if referenceNum == "" && sm.CreatedBy > 0 {
		referenceNum = fmt.Sprintf("%d", sm.CreatedBy)
	}

	fromStore := ""
	toStore := ""
	whs := strings.TrimSpace(sm.Warehouse)

	switch sm.TransType {
	case 20: // Goods Receipt PO
		movementType = "purchase_receipt"
		referenceType = "goods_receipt_note"
		toStore = whs
	case 13: // A/R Invoice (ObjType 13)
		movementType = "sales_delivery"
		referenceType = "invoice"
		fromStore = whs
	case 14: // A/R Credit Memo (ObjType 14)
		movementType = "sales_return"
		referenceType = "sales_return"
		toStore = whs
	case 15: // Delivery Notes
		movementType = "sales_delivery"
		referenceType = "delivery_note"
		fromStore = whs
	case 16: // Returns (customer returns w/o credit memo)
		movementType = "sales_return"
		referenceType = "sales_return"
		toStore = whs
	case 18: // A/P Invoice (ObjType 18) — A28: was mislabeled "A/R Invoice"
		movementType = "purchase_receipt"
		referenceType = "purchase_invoice"
		toStore = whs
	case 19: // A/P Credit Memo (ObjType 19) — A28: was mislabeled "A/R Credit Memo"
		movementType = "purchase_return"
		referenceType = "purchase_credit_note"
		fromStore = whs
	case 21: // Goods Return to supplier
		movementType = "purchase_return"
		referenceType = "goods_return"
		fromStore = whs
	case 67: // Inventory Transfer
		if sm.InQty > 0 {
			movementType = "transfer_in"
			toStore = whs
		} else {
			movementType = "transfer_out"
			fromStore = whs
		}
		referenceType = "transfer_request"
	case 59: // Goods Receipt non-PO
		movementType = "adjustment_positive"
		referenceType = "stock_count"
		toStore = whs
	case 60: // Goods Issue
		movementType = "adjustment_negative"
		referenceType = "stock_count"
		fromStore = whs
	case 162: // Inventory Recount
		if sm.InQty > 0 {
			movementType = "adjustment_positive"
			toStore = whs
		} else {
			movementType = "adjustment_negative"
			fromStore = whs
		}
		referenceType = "stock_count"
	default:
		if sm.InQty > 0 {
			movementType = "adjustment_positive"
			toStore = whs
		} else {
			movementType = "adjustment_negative"
			fromStore = whs
		}
	}

	qty := sm.InQty
	if qty <= 0 {
		qty = sm.OutQty
	}

	return CanonicalStockMovement{
		MovementType:    movementType,
		ReferenceType:   referenceType,
		ReferenceNumber: referenceNum,
		ProductSKU:      strings.TrimSpace(sm.ItemCode),
		FromStoreCode:   fromStore,
		ToStoreCode:     toStore,
		Quantity:        qty,
		MovementDate:    sm.DocDate,
		CostPerUnit:     sm.Price,
		TotalValue:      sm.TransValue,
		Metadata: map[string]interface{}{
			"sap_trans_num":  sm.TransNum,
			"sap_trans_type": sm.TransType,
			"sap_created_by": sm.CreatedBy,
			"sap_warehouse":  sm.Warehouse,
			"sap_in_qty":     sm.InQty,
			"sap_out_qty":    sm.OutQty,
			// A27: OINM.BASE_REF (source document DocNum) must survive for the
			// movement ledger audit trail even when reference_id cannot resolve.
			"sap_base_ref": sm.BaseRef,
		},
	}
}

// ----------------------------------------------------
// Domain 11: Incoming Payments (ORCT / RCT2)
// ----------------------------------------------------

type SAPIncomingPaymentInvoice struct {
	DocNum        int64   `json:"doc_num"` // Payment DocEntry
	LineID        int     `json:"line_id"`
	InvoiceID     int64   `json:"invoice_id"`      // OINV.DocEntry
	InvType       int     `json:"inv_type"`        // 13 = A/R Invoice
	SumApplied    float64 `json:"sum_applied"`     // Amount paid on this invoice
	InvoiceDocNum int64   `json:"invoice_doc_num"` // OINV.DocNum
}

type SAPIncomingPayment struct {
	DocEntry  int64                       `json:"doc_entry"`
	DocNum    int64                       `json:"doc_num"`
	DocDate   time.Time                   `json:"doc_date"`
	CardCode  string                      `json:"card_code"`
	CardName  string                      `json:"card_name"`
	DocCurr   string                      `json:"doc_curr"`
	DocTotal  float64                     `json:"doc_total"`
	CashSum   float64                     `json:"cash_sum"`
	TrsfrSum  float64                     `json:"trsfr_sum"`
	TrsfrRef  string                      `json:"trsfr_ref"`
	CheckSum  float64                     `json:"check_sum"`
	CreditSum float64                     `json:"credit_sum"`
	Comments  string                      `json:"comments"`
	JrnlMemo  string                      `json:"jrnl_memo"`
	Invoices  []SAPIncomingPaymentInvoice `json:"invoices,omitempty"`
}

type CanonicalIncomingPayment struct {
	PaymentNumber      string `json:"payment_number"`
	InvoiceNumber      string `json:"invoice_number"`        // Links to invoices.invoice_number
	SAPInvoiceDocEntry int64  `json:"sap_invoice_doc_entry"` // Fallback lookup via metadata->>'sap_doc_entry'
	// A20: true when the payment has no invoice allocation (on-account) or the
	// allocation could not be matched; routed to customer_payments instead of
	// being silently dropped.
	IsOnAccount      bool                   `json:"is_on_account,omitempty"`
	CustomerCode     string                 `json:"customer_code"`
	CustomerName     string                 `json:"customer_name"`
	PaymentDate      time.Time              `json:"payment_date"`
	PaymentAmount    float64                `json:"payment_amount"`
	PaymentMethod    string                 `json:"payment_method"` // cash, card, bank_transfer, check, other
	PaymentReference string                 `json:"payment_reference"`
	CurrencyCode     string                 `json:"currency_code"`
	Notes            string                 `json:"notes,omitempty"`
	Metadata         map[string]interface{} `json:"metadata"`
}

func (p *SAPIncomingPayment) ToCanonicalList() []CanonicalIncomingPayment {
	paymentMethod := "other"
	if p.CashSum > 0 && p.CashSum >= p.CreditSum && p.CashSum >= p.TrsfrSum && p.CashSum >= p.CheckSum {
		paymentMethod = "cash"
	} else if p.CreditSum > 0 && p.CreditSum >= p.CashSum && p.CreditSum >= p.TrsfrSum && p.CreditSum >= p.CheckSum {
		paymentMethod = "card"
	} else if p.TrsfrSum > 0 && p.TrsfrSum >= p.CashSum && p.TrsfrSum >= p.CreditSum && p.TrsfrSum >= p.CheckSum {
		paymentMethod = "bank_transfer"
	} else if p.CheckSum > 0 {
		paymentMethod = "check"
	}

	payRef := p.TrsfrRef
	if payRef == "" && p.JrnlMemo != "" {
		payRef = p.JrnlMemo
	}

	// A33: SAR is the deployment default (98_seed_currencies.sql).
	curr := strings.TrimSpace(p.DocCurr)
	if curr == "" {
		curr = "SAR"
	}

	var results []CanonicalIncomingPayment
	var appliedTotal float64
	for _, inv := range p.Invoices {
		if inv.SumApplied <= 0 {
			continue
		}
		appliedTotal += inv.SumApplied
		invNum := ""
		if inv.InvoiceDocNum > 0 {
			invNum = fmt.Sprintf("INV-SAP-%d", inv.InvoiceDocNum)
		}
		payNum := fmt.Sprintf("PAY-SAP-%d-%d", p.DocNum, inv.LineID)

		results = append(results, CanonicalIncomingPayment{
			PaymentNumber:      payNum,
			InvoiceNumber:      invNum,
			SAPInvoiceDocEntry: inv.InvoiceID,
			CustomerCode:       strings.TrimSpace(p.CardCode),
			CustomerName:       strings.TrimSpace(p.CardName),
			PaymentDate:        p.DocDate,
			PaymentAmount:      inv.SumApplied,
			PaymentMethod:      paymentMethod,
			PaymentReference:   payRef,
			CurrencyCode:       curr,
			Notes:              strings.TrimSpace(p.Comments),
			Metadata: map[string]interface{}{
				"sap_payment_doc_entry": p.DocEntry,
				"sap_payment_doc_num":   p.DocNum,
				"sap_invoice_doc_entry": inv.InvoiceID,
				"sap_line_id":           inv.LineID,
				"sap_inv_type":          inv.InvType,
				"sap_cash_sum":          p.CashSum,
				"sap_credit_sum":        p.CreditSum,
				"sap_transfer_sum":      p.TrsfrSum,
				"sap_check_sum":         p.CheckSum,
			},
		})
	}

	// A20: on-account portion — allocations that are not A/R invoices
	// (InvType != 13, e.g. credit-note or journal allocations) or a remaining
	// unallocated balance. These must reach the cloud as customer_payments,
	// never silently dropped.
	unallocated := math.Round((p.DocTotal-appliedTotal)*100) / 100
	if unallocated > 0.01 {
		payNum := fmt.Sprintf("PAY-SAP-%d-OA", p.DocNum)
		results = append(results, CanonicalIncomingPayment{
			PaymentNumber:    payNum,
			IsOnAccount:      true,
			CustomerCode:     strings.TrimSpace(p.CardCode),
			CustomerName:     strings.TrimSpace(p.CardName),
			PaymentDate:      p.DocDate,
			PaymentAmount:    unallocated,
			PaymentMethod:    paymentMethod,
			PaymentReference: payRef,
			CurrencyCode:     curr,
			Notes:            strings.TrimSpace(p.Comments),
			Metadata: map[string]interface{}{
				"sap_payment_doc_entry": p.DocEntry,
				"sap_payment_doc_num":   p.DocNum,
				"sap_on_account":        true,
				"sap_doc_total":         p.DocTotal,
				"sap_applied_total":     appliedTotal,
			},
		})
	}

	return results
}

// ----------------------------------------------------
// Domain 12: Sales Returns / Credit Memos (ORIN/RIN1, ORDN/RDN1) — A21
// ----------------------------------------------------

type SAPSalesReturnLine struct {
	DocEntry   int64   `json:"doc_entry"`
	LineNum    int     `json:"line_num"`
	ItemCode   string  `json:"item_code"`
	Dscription string  `json:"dscription"`
	Quantity   float64 `json:"quantity"`
	Price      float64 `json:"price"`
	LineTotal  float64 `json:"line_total"`
	VatSum     float64 `json:"vat_sum"`
	WhsCode    string  `json:"whs_code"`
	UnitMsr    string  `json:"unit_msr"`
	BaseEntry  int64   `json:"base_entry"` // Original invoice DocEntry (when linked)
	BaseType   int     `json:"base_type"`  // 13 = A/R Invoice
}

type SAPSalesReturn struct {
	DocEntry  int64                `json:"doc_entry"`
	DocNum    int64                `json:"doc_num"`
	DocDate   time.Time            `json:"doc_date"`
	CardCode  string               `json:"card_code"`
	CardName  string               `json:"card_name"`
	DocTotal  float64              `json:"doc_total"`
	VatSum    float64              `json:"vat_sum"`
	DiscSum   float64              `json:"disc_sum"`
	Canceled  string               `json:"canceled"`   // 'Y', 'C', or 'N'
	DocSource string               `json:"doc_source"` // 'credit_memo' (ORIN) or 'return' (ORDN)
	Comments  string               `json:"comments,omitempty"`
	Lines     []SAPSalesReturnLine `json:"lines,omitempty"`
}

type CanonicalSalesReturnLine struct {
	LineNumber   int                    `json:"line_number"`
	ProductSKU   string                 `json:"product_sku"`
	ProductName  string                 `json:"product_name"`
	StoreCode    string                 `json:"store_code,omitempty"`
	UOMCode      string                 `json:"uom_code,omitempty"`
	Quantity     float64                `json:"quantity"`
	UnitPrice    float64                `json:"unit_price"`
	TaxAmount    float64                `json:"tax_amount"`
	RefundAmount float64                `json:"refund_amount"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type CanonicalSalesReturn struct {
	ReturnNumber      string                     `json:"return_number"`
	CustomerCode      string                     `json:"customer_code"`
	CustomerName      string                     `json:"customer_name"`
	StoreCode         string                     `json:"store_code,omitempty"`
	ReturnDate        time.Time                  `json:"return_date"`
	Reason            string                     `json:"reason,omitempty"`
	Status            string                     `json:"status"` // 'completed' or 'cancelled'
	Subtotal          float64                    `json:"subtotal"`
	TaxAmount         float64                    `json:"tax_amount"`
	TotalRefundAmount float64                    `json:"total_refund_amount"`
	DocSource         string                     `json:"doc_source"` // 'credit_memo' | 'return'
	Lines             []CanonicalSalesReturnLine `json:"lines"`
	Metadata          map[string]interface{}     `json:"metadata"`
}

func (sr *SAPSalesReturn) ToCanonical() CanonicalSalesReturn {
	// 'completed' so vw_realtime_pnl.daily_returns and
	// vw_cashier_shift_reconciliation.refunds_summary include the row;
	// canceled documents import as 'cancelled'.
	status := "completed"
	if strings.ToUpper(strings.TrimSpace(sr.Canceled)) == "Y" ||
		strings.ToUpper(strings.TrimSpace(sr.Canceled)) == "C" {
		status = "cancelled"
	}

	primaryStore := ""
	lines := make([]CanonicalSalesReturnLine, len(sr.Lines))
	for i, l := range sr.Lines {
		whs := strings.TrimSpace(l.WhsCode)
		if primaryStore == "" && whs != "" {
			primaryStore = whs
		}
		lines[i] = CanonicalSalesReturnLine{
			LineNumber:   l.LineNum,
			ProductSKU:   strings.TrimSpace(l.ItemCode),
			ProductName:  strings.TrimSpace(l.Dscription),
			StoreCode:    whs,
			UOMCode:      strings.TrimSpace(l.UnitMsr),
			Quantity:     l.Quantity,
			UnitPrice:    l.Price,
			TaxAmount:    l.VatSum,
			RefundAmount: l.LineTotal,
			Metadata: map[string]interface{}{
				"sap_doc_entry":  sr.DocEntry,
				"sap_line_num":   l.LineNum,
				"sap_base_entry": l.BaseEntry,
				"sap_base_type":  l.BaseType,
			},
		}
	}

	// ORIN → CN-SAP-{DocNum}, ORDN → RET-SAP-{DocNum}
	prefix := "RET"
	if sr.DocSource == "credit_memo" {
		prefix = "CN"
	}

	subtotal := sr.DocTotal - sr.VatSum + sr.DiscSum
	if subtotal < 0 {
		subtotal = 0
	}

	return CanonicalSalesReturn{
		ReturnNumber:      fmt.Sprintf("%s-SAP-%d", prefix, sr.DocNum),
		CustomerCode:      strings.TrimSpace(sr.CardCode),
		CustomerName:      strings.TrimSpace(sr.CardName),
		StoreCode:         primaryStore,
		ReturnDate:        sr.DocDate,
		Reason:            strings.TrimSpace(sr.Comments),
		Status:            status,
		Subtotal:          subtotal,
		TaxAmount:         sr.VatSum,
		TotalRefundAmount: sr.DocTotal,
		DocSource:         sr.DocSource,
		Lines:             lines,
		Metadata: map[string]interface{}{
			"sap_doc_entry":  sr.DocEntry,
			"sap_doc_num":    sr.DocNum,
			"sap_card_code":  sr.CardCode,
			"sap_doc_total":  sr.DocTotal,
			"sap_canceled":   sr.Canceled,
			"sap_doc_source": sr.DocSource,
		},
	}
}

// ----------------------------------------------------
// Domain 13: Inventory Transfers (OWTR / WTR1) — A22
// ----------------------------------------------------

type SAPTransferLine struct {
	DocEntry   int64   `json:"doc_entry"`
	LineNum    int     `json:"line_num"`
	ItemCode   string  `json:"item_code"`
	Dscription string  `json:"dscription"`
	Quantity   float64 `json:"quantity"`
	WhsCode    string  `json:"whs_code"` // receiving warehouse
	UnitMsr    string  `json:"unit_msr"`
}

type SAPTransfer struct {
	DocEntry    int64             `json:"doc_entry"`
	DocNum      int64             `json:"doc_num"`
	DocDate     time.Time         `json:"doc_date"`
	Filler      string            `json:"filler"`       // issuing warehouse
	ToWarehouse string            `json:"to_warehouse"` // receiving warehouse
	Comments    string            `json:"comments,omitempty"`
	Lines       []SAPTransferLine `json:"lines,omitempty"`
}

type CanonicalTransferLine struct {
	LineNumber  int                    `json:"line_number"`
	ProductSKU  string                 `json:"product_sku"`
	ProductName string                 `json:"product_name"`
	UOMCode     string                 `json:"uom_code,omitempty"`
	Quantity    float64                `json:"quantity"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type CanonicalTransfer struct {
	TransferNumber string                  `json:"transfer_number"`
	FromStoreCode  string                  `json:"from_store_code"`
	ToStoreCode    string                  `json:"to_store_code"`
	TransferDate   time.Time               `json:"transfer_date"`
	Status         string                  `json:"status"` // 'completed'
	Notes          string                  `json:"notes,omitempty"`
	Lines          []CanonicalTransferLine `json:"lines"`
	Metadata       map[string]interface{}  `json:"metadata"`
}

func (t *SAPTransfer) ToCanonical() CanonicalTransfer {
	lines := make([]CanonicalTransferLine, len(t.Lines))
	for i, l := range t.Lines {
		lines[i] = CanonicalTransferLine{
			LineNumber:  l.LineNum,
			ProductSKU:  strings.TrimSpace(l.ItemCode),
			ProductName: strings.TrimSpace(l.Dscription),
			UOMCode:     strings.TrimSpace(l.UnitMsr),
			Quantity:    l.Quantity,
			Metadata: map[string]interface{}{
				"sap_doc_entry": t.DocEntry,
				"sap_line_num":  l.LineNum,
				"sap_whs_code":  l.WhsCode,
			},
		}
	}

	return CanonicalTransfer{
		TransferNumber: fmt.Sprintf("TR-SAP-%d", t.DocNum),
		FromStoreCode:  strings.TrimSpace(t.Filler),
		ToStoreCode:    strings.TrimSpace(t.ToWarehouse),
		TransferDate:   t.DocDate,
		Status:         "completed",
		Notes:          strings.TrimSpace(t.Comments),
		Lines:          lines,
		Metadata: map[string]interface{}{
			"sap_doc_entry": t.DocEntry,
			"sap_doc_num":   t.DocNum,
		},
	}
}

// ----------------------------------------------------
// Domain 14: Payment Terms (OCTG) — A23
// ----------------------------------------------------

type SAPPaymentTerm struct {
	GroupNum    int64   `json:"group_num"`
	PymGroup    string  `json:"pym_group"`
	InstMonths  int     `json:"inst_months"`
	InstDays    int     `json:"inst_days"`
	ExtraMonths int     `json:"extra_months"`
	ExtraDays   int     `json:"extra_days"`
	DiscPrcnt   float64 `json:"disc_prcnt"`
	DiscDays    int     `json:"disc_days"`
}

type CanonicalPaymentTerm struct {
	Code               string                 `json:"code"`
	Name               string                 `json:"name"`
	DueDays            int                    `json:"due_days"`
	DiscountDays       int                    `json:"discount_days"`
	DiscountPercentage float64                `json:"discount_percentage"`
	Metadata           map[string]interface{} `json:"metadata"`
}

func (pt *SAPPaymentTerm) ToCanonical() CanonicalPaymentTerm {
	// due_days: instalment months/days plus end-of-month extra grace.
	dueDays := pt.InstMonths*30 + pt.InstDays + pt.ExtraMonths*30 + pt.ExtraDays

	return CanonicalPaymentTerm{
		Code:               fmt.Sprintf("PT-%d", pt.GroupNum),
		Name:               strings.TrimSpace(pt.PymGroup),
		DueDays:            dueDays,
		DiscountDays:       pt.DiscDays,
		DiscountPercentage: pt.DiscPrcnt,
		Metadata: map[string]interface{}{
			"sap_group_num":    pt.GroupNum,
			"sap_inst_months":  pt.InstMonths,
			"sap_inst_days":    pt.InstDays,
			"sap_extra_months": pt.ExtraMonths,
			"sap_extra_days":   pt.ExtraDays,
		},
	}
}

// ----------------------------------------------------
// Domain 15: Outgoing Payments (OVPM / VPM2) — A31
// ----------------------------------------------------

type SAPOutgoingPaymentDocument struct {
	DocNum     int64   `json:"doc_num"` // Payment DocEntry
	LineID     int     `json:"line_id"`
	ObjType    int     `json:"obj_type"`  // 22 = Purchase Order
	DocEntry   int64   `json:"doc_entry"` // Base document DocEntry (PO)
	SumApplied float64 `json:"sum_applied"`
}

type SAPOutgoingPayment struct {
	DocEntry  int64                        `json:"doc_entry"`
	DocNum    int64                        `json:"doc_num"`
	DocDate   time.Time                    `json:"doc_date"`
	CardCode  string                       `json:"card_code"`
	CardName  string                       `json:"card_name"`
	DocCurr   string                       `json:"doc_curr"`
	DocTotal  float64                      `json:"doc_total"`
	CashSum   float64                      `json:"cash_sum"`
	TrsfrSum  float64                      `json:"trsfr_sum"`
	TrsfrRef  string                       `json:"trsfr_ref"`
	CheckSum  float64                      `json:"check_sum"`
	CreditSum float64                      `json:"credit_sum"`
	Comments  string                       `json:"comments"`
	JrnlMemo  string                       `json:"jrnl_memo"`
	Documents []SAPOutgoingPaymentDocument `json:"documents,omitempty"`
}

type CanonicalOutgoingPayment struct {
	PaymentNumber    string                 `json:"payment_number"`
	SupplierCode     string                 `json:"supplier_code"`
	SupplierName     string                 `json:"supplier_name"`
	PODocEntry       int64                  `json:"po_doc_entry"` // 0 when not PO-allocated
	PaymentDate      time.Time              `json:"payment_date"`
	PaymentAmount    float64                `json:"payment_amount"`
	PaymentMethod    string                 `json:"payment_method"`
	PaymentReference string                 `json:"payment_reference"`
	CurrencyCode     string                 `json:"currency_code"`
	Notes            string                 `json:"notes,omitempty"`
	Metadata         map[string]interface{} `json:"metadata"`
}

func (p *SAPOutgoingPayment) ToCanonicalList() []CanonicalOutgoingPayment {
	// A31: one canonical payment per OVPM header; the first PO allocation
	// (VPM2.ObjType = 22) links the payment to its purchase order.
	if p.DocTotal <= 0 {
		return nil
	}

	paymentMethod := "other"
	if p.CashSum > 0 && p.CashSum >= p.CreditSum && p.CashSum >= p.TrsfrSum && p.CashSum >= p.CheckSum {
		paymentMethod = "cash"
	} else if p.TrsfrSum > 0 && p.TrsfrSum >= p.CashSum && p.TrsfrSum >= p.CreditSum && p.TrsfrSum >= p.CheckSum {
		paymentMethod = "bank_transfer"
	} else if p.CheckSum > 0 {
		paymentMethod = "check"
	}

	payRef := p.TrsfrRef
	if payRef == "" && p.JrnlMemo != "" {
		payRef = p.JrnlMemo
	}

	curr := strings.TrimSpace(p.DocCurr)
	if curr == "" {
		curr = "SAR"
	}

	return []CanonicalOutgoingPayment{
		{
			PaymentNumber:    fmt.Sprintf("VPM-SAP-%d", p.DocNum),
			SupplierCode:     strings.TrimSpace(p.CardCode),
			SupplierName:     strings.TrimSpace(p.CardName),
			PODocEntry:       poDocEntryFor(p.Documents),
			PaymentDate:      p.DocDate,
			PaymentAmount:    p.DocTotal,
			PaymentMethod:    paymentMethod,
			PaymentReference: payRef,
			CurrencyCode:     curr,
			Notes:            strings.TrimSpace(p.Comments),
			Metadata: map[string]interface{}{
				"sap_payment_doc_entry": p.DocEntry,
				"sap_payment_doc_num":   p.DocNum,
				"sap_doc_total":         p.DocTotal,
				"sap_cash_sum":          p.CashSum,
				"sap_check_sum":         p.CheckSum,
				"sap_transfer_sum":      p.TrsfrSum,
				"sap_credit_sum":        p.CreditSum,
				"sap_allocations":       len(p.Documents),
			},
		},
	}
}

// poDocEntryFor returns the first PO allocation (ObjType 22) DocEntry for a payment.
func poDocEntryFor(docs []SAPOutgoingPaymentDocument) int64 {
	for _, d := range docs {
		if d.ObjType == 22 {
			return d.DocEntry
		}
	}
	return 0
}
