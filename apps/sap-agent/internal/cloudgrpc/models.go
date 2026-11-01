package cloudgrpc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/NasTecSol/nembus-sap/contracts"
)

type POSTransaction struct {
	ID                int64           `json:"id"`
	StoreID           int32           `json:"store_id"`
	CashierID         int32           `json:"cashier_id"`
	CashierSessionID  int32           `json:"cashier_session_id"`
	CustomerID        *int32          `json:"customer_id"`
	TransactionNumber string          `json:"transaction_number"`
	TransactionDate   time.Time       `json:"transaction_date"`
	Subtotal          float64         `json:"subtotal"`
	DiscountAmount    float64         `json:"discount_amount"`
	TaxAmount         float64         `json:"tax_amount"`
	TotalAmount       float64         `json:"total_amount"`
	AmountPaid        float64         `json:"amount_paid"`
	Status            string          `json:"status"`
	Metadata          json.RawMessage `json:"metadata"`
}

type POSTransactionLine struct {
	ID             int64   `json:"id"`
	TransactionID  int64   `json:"transaction_id"`
	ProductID      int32   `json:"product_id"`
	ProductSKU     string  `json:"product_sku"`
	ProductName    string  `json:"product_name"`
	Quantity       float64 `json:"quantity"`
	UoMID          *int32  `json:"uom_id"`
	UnitPrice      float64 `json:"unit_price"`
	DiscountAmount float64 `json:"discount_amount"`
	TaxAmount      float64 `json:"tax_amount"`
	Subtotal       float64 `json:"subtotal"`
	LineTotal      float64 `json:"line_total"`
	LineNumber     int     `json:"line_number"`
}

type POSPayment struct {
	ID            int64     `json:"id"`
	TransactionID int64     `json:"transaction_id"`
	PaymentMethod string    `json:"payment_method"`
	Amount        float64   `json:"amount"`
	Reference     *string   `json:"reference"`
	PaymentDate   time.Time `json:"payment_date"`
}

type SalesReturn struct {
	ID                    int64           `json:"id"`
	ReturnNumber          string          `json:"return_number"`
	StoreID               int32           `json:"store_id"`
	CashierID             *int32          `json:"cashier_id"`
	CustomerID            *int32          `json:"customer_id"`
	OriginalTransactionID *int32          `json:"original_transaction_id"`
	ReturnDate            time.Time       `json:"return_date"`
	ReturnReason          *string         `json:"return_reason"`
	Status                string          `json:"status"`
	Subtotal              float64         `json:"subtotal"`
	TaxAmount             float64         `json:"tax_amount"`
	TotalRefundAmount     float64         `json:"total_refund_amount"`
	Metadata              json.RawMessage `json:"metadata"`
}

type SalesReturnLine struct {
	ID             int64   `json:"id"`
	ReturnID       int64   `json:"return_id"`
	ProductID      int32   `json:"product_id"`
	ProductSKU     string  `json:"product_sku"`
	ProductName    string  `json:"product_name"`
	OriginalLineID *int32  `json:"original_line_id"`
	Quantity       float64 `json:"quantity"`
	UnitPrice      float64 `json:"unit_price"`
	RefundAmount   float64 `json:"refund_amount"`
	ReturnToStock  bool    `json:"return_to_stock"`
	LineNumber     int     `json:"line_number"`
}

type ShortageItem struct {
	ProductID   int32   `json:"product_id"`
	ProductSKU  string  `json:"product_sku"`
	ProductName string  `json:"product_name"`
	Quantity    float64 `json:"quantity"`
	SAPOnHand   float64 `json:"sap_on_hand"`
}

type TenantResolveResult struct {
	TenantSlug     string
	TenantName     string
	OrganizationID int32
	DefaultStoreID int32
}

// CloudGateway is implemented by the gRPC cloud client (production path).
type CloudGateway interface {
	Close() error
	TestConnection(ctx context.Context) error
	ResolveTenant(ctx context.Context, slug string) (*TenantResolveResult, error)
	SendBatchWithRetry(ctx context.Context, payload *contracts.MigrationBatchPayload) (*contracts.MigrationBatchResponse, error)
	UpsertCategory(ctx context.Context, code, name string) (int32, error)
	UpsertUoM(ctx context.Context, code, name string) (int32, error)
	UpsertProduct(ctx context.Context, sku, name, foreignName string, categoryID, uomID, taxCategoryID *int32, isWeighted bool, costPrice float64, skipCostUpdate bool) (int32, error)
	LookupProductBySKU(ctx context.Context, sku string) (productID int32, cost float64, found bool, err error)
	UpsertBarcode(ctx context.Context, productID int32, barcode string, isPrimary bool) error
	UpsertProductPrice(ctx context.Context, productID, priceListID int32, uomID *int32, price float64) error
	FetchUnpostedPOSTransactions(ctx context.Context, limit int) ([]POSTransaction, error)
	FetchTransactionLines(ctx context.Context, transactionID int64) ([]POSTransactionLine, error)
	FetchTransactionPayments(ctx context.Context, transactionID int64) ([]POSPayment, error)
	MarkTransactionPostedToSAP(ctx context.Context, transactionID int64, sapDocEntry, sapDocNum int) error
	FetchCustomerSAPCode(ctx context.Context, customerID int32) (string, error)
	FetchUnpostedSalesReturns(ctx context.Context, limit int) ([]SalesReturn, error)
	FetchSalesReturnLines(ctx context.Context, returnID int64) ([]SalesReturnLine, error)
	MarkSalesReturnPostedToSAP(ctx context.Context, returnID int64, sapDocEntry, sapDocNum int) error
	FetchOriginalSAPDocEntry(ctx context.Context, originalTxID int32) (int, error)
	HoldTransactionForStockReview(ctx context.Context, txID int64, txNum string, storeID int32, whCode string, totalAmount float64, shortageItems []ShortageItem, summary string) error
	GetActiveProductPrice(ctx context.Context, productID, priceListID int32) (float64, bool, error)
	GetProductCostPrice(ctx context.Context, productID int32) (float64, error)
	GetApprovalPolicyRules(ctx context.Context, domain, action string) (map[string]any, bool, error)
	SubmitPriceApprovalRequest(ctx context.Context, productID int32, sku string, priceListID int32, uomID *int32, oldPrice, newPrice, oldCost, newCost *float64, reason string) error
	GetReconciliationSummary(ctx context.Context, runID string, orgID int) ([]byte, error)
}
