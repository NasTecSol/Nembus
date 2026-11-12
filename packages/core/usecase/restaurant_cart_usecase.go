package usecase

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RestaurantCartUseCase handles restaurant carts, cart items, cart-to-order conversion,
// reopening orders back into carts, and transactional restaurant order item calculation.
type RestaurantCartUseCase struct {
	repo *repository.Queries
	pool *pgxpool.Pool
}

func NewRestaurantCartUseCase() *RestaurantCartUseCase {
	return &RestaurantCartUseCase{}
}

func (uc *RestaurantCartUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *RestaurantCartUseCase) SetPool(pool *pgxpool.Pool) {
	uc.pool = pool
}

func (uc *RestaurantCartUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

var errRollbackRestaurantTx = errors.New("rollback restaurant transaction")

func unmarshalJSONObject(raw []byte) map[string]interface{} {
	m := map[string]interface{}{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &m)
	}
	if m == nil {
		m = map[string]interface{}{}
	}
	return m
}

func isInactiveOrderStatus(status string) bool {
	s := strings.ToLower(strings.TrimSpace(status))
	return s == "voided" || s == "cancelled"
}

func isRestaurantOrderPaidOrClosed(order repository.RestaurantOrder) bool {
	return strings.EqualFold(strings.TrimSpace(order.Status), "paid") ||
		isInactiveOrderStatus(order.Status) ||
		strings.EqualFold(strings.TrimSpace(order.PaymentStatus), "paid")
}

// withTx executes fn inside a database transaction when a pgxpool is available via
// uc.pool or context middleware.PoolKey, or when uc.repo supports Begin, falling back
// cleanly to uc.repo otherwise.
func (uc *RestaurantCartUseCase) withTx(ctx context.Context, fn func(q *repository.Queries) *repository.Response) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}
	return withRestaurantTx(ctx, uc.pool, uc.repo, fn)
}

func withRestaurantTx(ctx context.Context, explicitPool *pgxpool.Pool, repo *repository.Queries, fn func(q *repository.Queries) *repository.Response) *repository.Response {
	if repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	pool := resolvePoolFromContext(ctx, explicitPool)
	if pool != nil {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to begin transaction: "+err.Error(), nil)
		}
		defer func() {
			_ = tx.Rollback(ctx)
		}()

		qtx := repository.New(tx)
		resp := fn(qtx)
		if resp == nil || resp.StatusCode >= 400 {
			return resp
		}

		if err := tx.Commit(ctx); err != nil {
			return utils.NewResponse(utils.CodeError, "failed to commit transaction: "+err.Error(), nil)
		}
		return resp
	}

	var txResp *repository.Response
	handled, err := repo.TryInRestaurantTx(ctx, func(qtx *repository.Queries) error {
		txResp = fn(qtx)
		if txResp == nil || txResp.StatusCode >= 400 {
			return errRollbackRestaurantTx
		}
		return nil
	})
	if handled {
		if err != nil && !errors.Is(err, errRollbackRestaurantTx) {
			return utils.NewResponse(utils.CodeError, "failed to commit transaction: "+err.Error(), nil)
		}
		if txResp != nil {
			return txResp
		}
		return utils.NewResponse(utils.CodeError, "transaction failed", nil)
	}
	return fn(repo)
}

// lockAndGetRestaurantCartByID acquires the cart's FOR UPDATE row lock before
// reading its status or items so concurrent cart mutations and checkout serialize.
func lockAndGetRestaurantCartByID(ctx context.Context, q *repository.Queries, cartID uuid.UUID) (repository.RestaurantCart, error) {
	if _, err := q.LockRestaurantCart(ctx, cartID); err != nil {
		return repository.RestaurantCart{}, err
	}
	return q.GetRestaurantCartByID(ctx, cartID)
}

func restaurantCartNotFoundOrErr(err error, action string) *repository.Response {
	if errors.Is(err, pgx.ErrNoRows) {
		return utils.NewResponse(utils.CodeNotFound, "restaurant cart not found", nil)
	}
	return utils.NewResponse(utils.CodeError, "failed to "+action+": "+err.Error(), nil)
}

// lockAndGetRestaurantOrderByID acquires the order's FOR UPDATE row lock before
// reading its status, payments, or recipe movements.
func lockAndGetRestaurantOrderByID(ctx context.Context, q *repository.Queries, orderID int32) (repository.RestaurantOrder, error) {
	if _, err := q.LockRestaurantOrder(ctx, orderID); err != nil {
		return repository.RestaurantOrder{}, err
	}
	return q.GetRestaurantOrder(ctx, orderID)
}

func restaurantOrderNotFoundOrErr(err error, action string) *repository.Response {
	if errors.Is(err, pgx.ErrNoRows) {
		return utils.NewResponse(utils.CodeNotFound, "restaurant order not found", nil)
	}
	return utils.NewResponse(utils.CodeError, "failed to "+action+": "+err.Error(), nil)
}

// === Input Structs ===

type RestaurantModifierSnapshotItem struct {
	ModifierID      int32   `json:"modifier_id,omitempty" example:"1"`
	ModifierName    string  `json:"modifier_name" example:"Extra Cheese"`
	ModifierType    string  `json:"modifier_type,omitempty" example:"addon"`
	PriceAdjustment float64 `json:"price_adjustment" example:"4.00"`
}

type RestaurantComboComponentSnapshotItem struct {
	ComboComponentID    int32   `json:"combo_component_id,omitempty" example:"1"`
	ComponentMenuItemID int32   `json:"component_menu_item_id" example:"11"`
	ComponentName       string  `json:"component_name" example:"French Fries"`
	GroupName           string  `json:"group_name,omitempty" example:"Sides"`
	PriceAdjustment     float64 `json:"price_adjustment" example:"5.00"`
}

type RestaurantLineItemInput struct {
	MenuItemID           int32                                  `json:"menu_item_id"`
	Quantity             float64                                `json:"quantity"`
	UnitPrice            *float64                               `json:"unit_price,omitempty"`
	ModifierIDs          []int32                                `json:"modifier_ids,omitempty"`
	Modifiers            []RestaurantModifierSnapshotItem       `json:"modifiers,omitempty"`
	ComboComponentIDs    []int32                                `json:"combo_component_ids,omitempty"`
	ComboComponents      []RestaurantComboComponentSnapshotItem `json:"combo_components,omitempty"`
	DiscountAmount       float64                                `json:"discount_amount,omitempty"`
	CustomizationDetails json.RawMessage                        `json:"customization_details,omitempty"`
	Notes                string                                 `json:"notes,omitempty"`
	Status               string                                 `json:"status,omitempty"`
	Metadata             json.RawMessage                        `json:"metadata,omitempty"`
}

type RestaurantCreateCartInput struct {
	CartNumber      string                    `json:"cart_number"`
	OrganizationID  *int32                    `json:"organization_id"`
	StoreID         int32                     `json:"store_id"`
	TableID         *int32                    `json:"table_id"`
	OrderType       string                    `json:"order_type"`
	CustomerID      *int32                    `json:"customer_id"`
	GuestIdentifier string                    `json:"guest_identifier"`
	GuestName       string                    `json:"guest_name"`
	GuestEmail      string                    `json:"guest_email"`
	GuestPhone      string                    `json:"guest_phone"`
	CartStatus      string                    `json:"cart_status"`
	Channel         string                    `json:"channel"`
	PaymentMethod   string                    `json:"payment_method"`
	PaymentGateway  string                    `json:"payment_gateway"`
	DeviceInfo      json.RawMessage           `json:"device_info"`
	CreatedByUserID *int32                    `json:"created_by_user_id"`
	CashierID       *int32                    `json:"cashier_id"`
	PosTerminalID   *int32                    `json:"pos_terminal_id"`
	KioskSessionID  *int32                    `json:"kiosk_session_id"`
	CouponCode      string                    `json:"coupon_code"`
	DiscountCode    string                    `json:"discount_code"`
	DeliveryAddress json.RawMessage           `json:"delivery_address"`
	ExpiresAt       *time.Time                `json:"expires_at"`
	Notes           string                    `json:"notes"`
	Metadata        json.RawMessage           `json:"metadata"`
	Items           []RestaurantLineItemInput `json:"items"`
}

type RestaurantAddToCartInput struct {
	CartID               *uuid.UUID                             `json:"cart_id"`
	OrganizationID       *int32                                 `json:"organization_id"`
	StoreID              int32                                  `json:"store_id"`
	TableID              *int32                                 `json:"table_id"`
	OrderType            string                                 `json:"order_type"`
	CustomerID           *int32                                 `json:"customer_id"`
	GuestIdentifier      string                                 `json:"guest_identifier"`
	GuestName            string                                 `json:"guest_name"`
	GuestEmail           string                                 `json:"guest_email"`
	GuestPhone           string                                 `json:"guest_phone"`
	Channel              string                                 `json:"channel"`
	CashierID            *int32                                 `json:"cashier_id"`
	PosTerminalID        *int32                                 `json:"pos_terminal_id"`
	KioskSessionID       *int32                                 `json:"kiosk_session_id"`
	CreatedByUserID      *int32                                 `json:"created_by_user_id"`
	MenuItemID           int32                                  `json:"menu_item_id"`
	Quantity             float64                                `json:"quantity"`
	UnitPrice            *float64                               `json:"unit_price,omitempty"`
	ModifierIDs          []int32                                `json:"modifier_ids,omitempty"`
	Modifiers            []RestaurantModifierSnapshotItem       `json:"modifiers,omitempty"`
	ComboComponentIDs    []int32                                `json:"combo_component_ids,omitempty"`
	ComboComponents      []RestaurantComboComponentSnapshotItem `json:"combo_components,omitempty"`
	DiscountAmount       float64                                `json:"discount_amount,omitempty"`
	CustomizationDetails json.RawMessage                        `json:"customization_details,omitempty"`
	Notes                string                                 `json:"notes,omitempty"`
	Metadata             json.RawMessage                        `json:"metadata,omitempty"`
}

type RestaurantUpdateCartItemInput struct {
	ItemID               uuid.UUID                              `json:"item_id"`
	Quantity             float64                                `json:"quantity"`
	ModifierIDs          []int32                                `json:"modifier_ids,omitempty"`
	Modifiers            []RestaurantModifierSnapshotItem       `json:"modifiers,omitempty"`
	ComboComponentIDs    []int32                                `json:"combo_component_ids,omitempty"`
	ComboComponents      []RestaurantComboComponentSnapshotItem `json:"combo_components,omitempty"`
	HasModifiersUpdate   bool                                   `json:"-"`
	HasCombosUpdate      bool                                   `json:"-"`
	DiscountAmount       *float64                               `json:"discount_amount,omitempty"`
	CustomizationDetails json.RawMessage                        `json:"customization_details,omitempty"`
	Notes                *string                                `json:"notes,omitempty"`
	Metadata             json.RawMessage                        `json:"metadata,omitempty"`
}

type RestaurantCheckoutInput struct {
	CartID           uuid.UUID       `json:"cart_id"`
	OrderNumber      string          `json:"order_number"`
	OrderSource      string          `json:"order_source"`
	OrderType        string          `json:"order_type"`
	TableID          *int32          `json:"table_id"`
	CashierID        *int32          `json:"cashier_id"`
	CashierSessionID *int32          `json:"cashier_session_id"`
	CustomerID       *int32          `json:"customer_id"`
	PosTerminalID    *int32          `json:"pos_terminal_id"`
	KioskSessionID   *int32          `json:"kiosk_session_id"`
	CreatedByUserID  *int32          `json:"created_by_user_id"`
	CustomerName     string          `json:"customer_name"`
	CustomerEmail    string          `json:"customer_email"`
	CustomerPhone    string          `json:"customer_phone"`
	DeliveryAddress  json.RawMessage `json:"delivery_address"`
	PaymentStatus    string          `json:"payment_status"`
	PaymentMethod    string          `json:"payment_method"`
	PaymentGateway   string          `json:"payment_gateway"`
	ReferenceNumber  string          `json:"reference_number"`
	AmountPaid       float64         `json:"amount_paid"`
	PosTransactionID *int32          `json:"pos_transaction_id"`
	CouponCode       string          `json:"coupon_code"`
	DiscountCode     string          `json:"discount_code"`
	Notes            string          `json:"notes"`
	Status           string          `json:"status"`
	Metadata         json.RawMessage `json:"metadata"`
}

type RestaurantOrderStatusUpdateInput struct {
	Status          string          `json:"status"`
	Reason          string          `json:"reason"`
	Notes           string          `json:"notes"`
	ChangedByUserID *int32          `json:"changed_by_user_id"`
	Metadata        json.RawMessage `json:"metadata"`
}

type RestaurantSettleOrderInput struct {
	PosTransactionID *int32          `json:"pos_transaction_id"`
	CashierID        *int32          `json:"cashier_id"`
	CashierSessionID *int32          `json:"cashier_session_id"`
	PosTerminalID    *int32          `json:"pos_terminal_id"`
	PaymentMethod    string          `json:"payment_method"`
	PaymentGateway   string          `json:"payment_gateway"`
	AmountPaid       *float64        `json:"amount_paid"`
	ReferenceNumber  string          `json:"reference_number"`
	Metadata         json.RawMessage `json:"metadata"`
}

type RestaurantDirectOrderInput struct {
	OrganizationID   *int32                    `json:"organization_id"`
	StoreID          int32                     `json:"store_id"`
	TableID          *int32                    `json:"table_id"`
	CashierID        *int32                    `json:"cashier_id"`
	CashierSessionID *int32                    `json:"cashier_session_id"`
	CustomerID       *int32                    `json:"customer_id"`
	PosTerminalID    *int32                    `json:"pos_terminal_id"`
	KioskSessionID   *int32                    `json:"kiosk_session_id"`
	CreatedByUserID  *int32                    `json:"created_by_user_id"`
	SourceCartID     *uuid.UUID                `json:"source_cart_id"`
	OrderNumber      string                    `json:"order_number"`
	OrderSource      string                    `json:"order_source"`
	OrderType        string                    `json:"order_type"`
	Status           string                    `json:"status"`
	PaymentStatus    string                    `json:"payment_status"`
	CustomerName     string                    `json:"customer_name"`
	CustomerEmail    string                    `json:"customer_email"`
	CustomerPhone    string                    `json:"customer_phone"`
	DeliveryAddress  json.RawMessage           `json:"delivery_address"`
	Subtotal         float64                   `json:"subtotal"`
	DiscountAmount   float64                   `json:"discount_amount"`
	TaxAmount        float64                   `json:"tax_amount"`
	TotalAmount      float64                   `json:"total_amount"`
	AmountPaid       float64                   `json:"amount_paid"`
	ChangeGiven      float64                   `json:"change_given"`
	CouponCode       string                    `json:"coupon_code"`
	DiscountCode     string                    `json:"discount_code"`
	Notes            string                    `json:"notes"`
	PosTransactionID *int32                    `json:"pos_transaction_id"`
	Metadata         json.RawMessage           `json:"metadata"`
	Items            []RestaurantLineItemInput `json:"items"`
}

// === Clean Output Structs (matching CartOutput / ListCartItemsOutput parity) ===

type RestaurantCartOutput struct {
	ID                 uuid.UUID       `json:"id"`
	CartNumber         *string         `json:"cart_number"`
	OrganizationID     *int32          `json:"organization_id"`
	StoreID            int32           `json:"store_id"`
	TableID            *int32          `json:"table_id"`
	OrderType          string          `json:"order_type"`
	CustomerID         *int32          `json:"customer_id"`
	GuestIdentifier    *string         `json:"guest_identifier"`
	GuestName          *string         `json:"guest_name"`
	GuestEmail         *string         `json:"guest_email"`
	GuestPhone         *string         `json:"guest_phone"`
	CartStatus         string          `json:"cart_status"`
	Channel            *string         `json:"channel"`
	PaymentMethod      *string         `json:"payment_method"`
	PaymentGateway     *string         `json:"payment_gateway"`
	DeviceInfo         json.RawMessage `json:"device_info"`
	CreatedByUserID    *int32          `json:"created_by_user_id"`
	CashierID          *int32          `json:"cashier_id"`
	PosTerminalID      *int32          `json:"pos_terminal_id"`
	KioskSessionID     *int32          `json:"kiosk_session_id"`
	Subtotal           float64         `json:"subtotal"`
	DiscountAmount     float64         `json:"discount_amount"`
	TaxAmount          float64         `json:"tax_amount"`
	TotalAmount        float64         `json:"total_amount"`
	CouponCode         *string         `json:"coupon_code"`
	DiscountCode       *string         `json:"discount_code"`
	DeliveryAddress    json.RawMessage `json:"delivery_address"`
	ConvertedToOrderID *int32          `json:"converted_to_order_id"`
	ConvertedAt        *time.Time      `json:"converted_at"`
	LastActivityAt     *time.Time      `json:"last_activity_at"`
	ExpiresAt          *time.Time      `json:"expires_at"`
	Notes              *string         `json:"notes"`
	Metadata           json.RawMessage `json:"metadata"`
	CreatedAt          *time.Time      `json:"created_at"`
	UpdatedAt          *time.Time      `json:"updated_at"`
}

type RestaurantCartItemOutput struct {
	ID                      uuid.UUID       `json:"id"`
	CartID                  uuid.UUID       `json:"cart_id"`
	OrganizationID          *int32          `json:"organization_id"`
	MenuItemID              int32           `json:"menu_item_id"`
	MenuItemName            string          `json:"menu_item_name,omitempty"`
	MenuItemShortName       *string         `json:"menu_item_short_name,omitempty"`
	MenuItemImageURL        *string         `json:"menu_item_image_url,omitempty"`
	MenuItemType            string          `json:"menu_item_type,omitempty"`
	MenuCategoryName        *string         `json:"menu_category_name,omitempty"`
	Quantity                float64         `json:"quantity"`
	UnitPrice               float64         `json:"unit_price"`
	ModifiersSnapshot       json.RawMessage `json:"modifiers_snapshot"`
	ModifiersTotal          float64         `json:"modifiers_total"`
	ComboComponentsSnapshot json.RawMessage `json:"combo_components_snapshot"`
	ComboTotal              float64         `json:"combo_total"`
	DiscountAmount          float64         `json:"discount_amount"`
	TaxCategoryID           *int32          `json:"tax_category_id"`
	TaxRate                 float64         `json:"tax_rate"`
	TaxAmount               float64         `json:"tax_amount"`
	LineTotal               float64         `json:"line_total"`
	CustomizationDetails    json.RawMessage `json:"customization_details"`
	Notes                   *string         `json:"notes"`
	Metadata                json.RawMessage `json:"metadata"`
	AddedAt                 *time.Time      `json:"added_at"`
	UpdatedAt               *time.Time      `json:"updated_at"`
}

type RestaurantCartTotalsOutput struct {
	ItemCount     int64   `json:"item_count"`
	TotalQuantity float64 `json:"total_quantity"`
	Subtotal      float64 `json:"subtotal"`
	TotalDiscount float64 `json:"total_discount"`
	TotalTax      float64 `json:"total_tax"`
	TotalAmount   float64 `json:"total_amount"`
}

type RestaurantOrderWithItems struct {
	repository.RestaurantOrder
	Items []repository.RestaurantOrderItem `json:"items"`
}

// === Internal Calculation & Conversion Helpers ===

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

func pgTimestampToPtr(ts pgtype.Timestamp) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

func ensureJSONObject(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage("{}")
	}
	return raw
}

// mergeCartItemMetadataPreservingCoupon merges incoming item metadata on top of
// existing metadata while preserving coupon bookkeeping ("coupon_discount_amount",
// "applied_coupon_code") so subsequent coupon removal can undo the discount cleanly.
func mergeCartItemMetadataPreservingCoupon(existingRaw, incomingRaw json.RawMessage) json.RawMessage {
	merged := map[string]interface{}{}
	existingMap := map[string]interface{}{}
	if len(existingRaw) > 0 && json.Unmarshal(existingRaw, &existingMap) == nil {
		for k, v := range existingMap {
			merged[k] = v
		}
	}
	if len(incomingRaw) > 0 {
		incomingMap := map[string]interface{}{}
		if json.Unmarshal(incomingRaw, &incomingMap) == nil {
			for k, v := range incomingMap {
				merged[k] = v
			}
		}
	}
	if v, ok := existingMap["coupon_discount_amount"]; ok {
		merged["coupon_discount_amount"] = v
	}
	if v, ok := existingMap["applied_coupon_code"]; ok {
		merged["applied_coupon_code"] = v
	}
	if len(merged) == 0 {
		return json.RawMessage("{}")
	}
	encoded, err := json.Marshal(merged)
	if err != nil {
		return ensureJSONObject(existingRaw)
	}
	return encoded
}

// reconcileCartItemCouponMetadata keeps coupon bookkeeping consistent if an explicit
// discount override lowers or clears the line discount below the recorded coupon amount.
func reconcileCartItemCouponMetadata(raw json.RawMessage, discountOverride *float64) json.RawMessage {
	if discountOverride == nil {
		return ensureJSONObject(raw)
	}
	meta := map[string]interface{}{}
	if len(raw) == 0 || json.Unmarshal(raw, &meta) != nil {
		return ensureJSONObject(raw)
	}
	rawCouponDisc, hasCoupon := meta["coupon_discount_amount"]
	if !hasCoupon {
		return ensureJSONObject(raw)
	}
	couponDisc, _ := rawCouponDisc.(float64)
	overrideVal := roundMoney(*discountOverride)
	if overrideVal <= 0 {
		delete(meta, "coupon_discount_amount")
		delete(meta, "applied_coupon_code")
	} else if overrideVal < couponDisc {
		meta["coupon_discount_amount"] = overrideVal
	} else {
		return ensureJSONObject(raw)
	}
	if len(meta) == 0 {
		return json.RawMessage("{}")
	}
	encoded, err := json.Marshal(meta)
	if err != nil {
		return ensureJSONObject(raw)
	}
	return encoded
}

func ensureJSONArray(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage("[]")
	}
	return raw
}

func generateRestaurantRef(prefix string) string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s-%s", prefix, time.Now().Format("20060102-150405"), strings.ToUpper(hex.EncodeToString(b)))
}

func restaurantCartToOutput(c repository.RestaurantCart) RestaurantCartOutput {
	return RestaurantCartOutput{
		ID:                 c.ID,
		CartNumber:         utils.TextToPtr(c.CartNumber),
		OrganizationID:     utils.Int4ToPtr(c.OrganizationID),
		StoreID:            c.StoreID,
		TableID:            utils.Int4ToPtr(c.TableID),
		OrderType:          c.OrderType,
		CustomerID:         utils.Int4ToPtr(c.CustomerID),
		GuestIdentifier:    utils.TextToPtr(c.GuestIdentifier),
		GuestName:          utils.TextToPtr(c.GuestName),
		GuestEmail:         utils.TextToPtr(c.GuestEmail),
		GuestPhone:         utils.TextToPtr(c.GuestPhone),
		CartStatus:         string(c.CartStatus),
		Channel:            utils.TextToPtr(c.Channel),
		PaymentMethod:      utils.TextToPtr(c.PaymentMethod),
		PaymentGateway:     utils.TextToPtr(c.PaymentGateway),
		DeviceInfo:         ensureJSONObject(c.DeviceInfo),
		CreatedByUserID:    utils.Int4ToPtr(c.CreatedByUserID),
		CashierID:          utils.Int4ToPtr(c.CashierID),
		PosTerminalID:      utils.Int4ToPtr(c.PosTerminalID),
		KioskSessionID:     utils.Int4ToPtr(c.KioskSessionID),
		Subtotal:           roundMoney(numericToFloat(c.Subtotal)),
		DiscountAmount:     roundMoney(numericToFloat(c.DiscountAmount)),
		TaxAmount:          roundMoney(numericToFloat(c.TaxAmount)),
		TotalAmount:        roundMoney(numericToFloat(c.TotalAmount)),
		CouponCode:         utils.TextToPtr(c.CouponCode),
		DiscountCode:       utils.TextToPtr(c.DiscountCode),
		DeliveryAddress:    c.DeliveryAddress,
		ConvertedToOrderID: utils.Int4ToPtr(c.ConvertedToOrderID),
		ConvertedAt:        pgTimestampToPtr(c.ConvertedAt),
		LastActivityAt:     pgTimestampToPtr(c.LastActivityAt),
		ExpiresAt:          pgTimestampToPtr(c.ExpiresAt),
		Notes:              utils.TextToPtr(c.Notes),
		Metadata:           ensureJSONObject(c.Metadata),
		CreatedAt:          pgTimestampToPtr(c.CreatedAt),
		UpdatedAt:          pgTimestampToPtr(c.UpdatedAt),
	}
}

func restaurantCartsToOutput(carts []repository.RestaurantCart) []RestaurantCartOutput {
	out := make([]RestaurantCartOutput, len(carts))
	for i, c := range carts {
		out[i] = restaurantCartToOutput(c)
	}
	return out
}

func restaurantCartItemRowToOutput(r repository.ListRestaurantCartItemsRow) RestaurantCartItemOutput {
	return RestaurantCartItemOutput{
		ID:                      r.ID,
		CartID:                  r.CartID,
		OrganizationID:          utils.Int4ToPtr(r.OrganizationID),
		MenuItemID:              r.MenuItemID,
		MenuItemName:            r.MenuItemName,
		MenuItemShortName:       utils.TextToPtr(r.MenuItemShortName),
		MenuItemImageURL:        utils.TextToPtr(r.MenuItemImageUrl),
		MenuItemType:            r.MenuItemType,
		MenuCategoryName:        utils.TextToPtr(r.MenuCategoryName),
		Quantity:                numericToFloat(r.Quantity),
		UnitPrice:               numericToFloat(r.UnitPrice),
		ModifiersSnapshot:       ensureJSONArray(r.ModifiersSnapshot),
		ModifiersTotal:          roundMoney(numericToFloat(r.ModifiersTotal)),
		ComboComponentsSnapshot: ensureJSONArray(r.ComboComponentsSnapshot),
		ComboTotal:              roundMoney(numericToFloat(r.ComboTotal)),
		DiscountAmount:          roundMoney(numericToFloat(r.DiscountAmount)),
		TaxCategoryID:           utils.Int4ToPtr(r.TaxCategoryID),
		TaxRate:                 roundMoney(numericToFloat(r.TaxRate)),
		TaxAmount:               roundMoney(numericToFloat(r.TaxAmount)),
		LineTotal:               roundMoney(numericToFloat(r.LineTotal)),
		CustomizationDetails:    ensureJSONObject(r.CustomizationDetails),
		Notes:                   utils.TextToPtr(r.Notes),
		Metadata:                ensureJSONObject(r.Metadata),
		AddedAt:                 pgTimestampToPtr(r.AddedAt),
		UpdatedAt:               pgTimestampToPtr(r.UpdatedAt),
	}
}

func restaurantCartItemToOutput(i repository.RestaurantCartItem) RestaurantCartItemOutput {
	return RestaurantCartItemOutput{
		ID:                      i.ID,
		CartID:                  i.CartID,
		OrganizationID:          utils.Int4ToPtr(i.OrganizationID),
		MenuItemID:              i.MenuItemID,
		Quantity:                numericToFloat(i.Quantity),
		UnitPrice:               numericToFloat(i.UnitPrice),
		ModifiersSnapshot:       ensureJSONArray(i.ModifiersSnapshot),
		ModifiersTotal:          roundMoney(numericToFloat(i.ModifiersTotal)),
		ComboComponentsSnapshot: ensureJSONArray(i.ComboComponentsSnapshot),
		ComboTotal:              roundMoney(numericToFloat(i.ComboTotal)),
		DiscountAmount:          roundMoney(numericToFloat(i.DiscountAmount)),
		TaxCategoryID:           utils.Int4ToPtr(i.TaxCategoryID),
		TaxRate:                 roundMoney(numericToFloat(i.TaxRate)),
		TaxAmount:               roundMoney(numericToFloat(i.TaxAmount)),
		LineTotal:               roundMoney(numericToFloat(i.LineTotal)),
		CustomizationDetails:    ensureJSONObject(i.CustomizationDetails),
		Notes:                   utils.TextToPtr(i.Notes),
		Metadata:                ensureJSONObject(i.Metadata),
		AddedAt:                 pgTimestampToPtr(i.AddedAt),
		UpdatedAt:               pgTimestampToPtr(i.UpdatedAt),
	}
}

func restaurantCartItemRowsToOutput(rows []repository.ListRestaurantCartItemsRow) []RestaurantCartItemOutput {
	out := make([]RestaurantCartItemOutput, len(rows))
	for i, r := range rows {
		out[i] = restaurantCartItemRowToOutput(r)
	}
	return out
}

type resolvedRestaurantLine struct {
	MenuItem            repository.MenuItem
	Quantity            float64
	UnitPrice           float64
	ModifiersSnapshot   json.RawMessage
	ModifiersTotal      float64
	ComboSnapshot       json.RawMessage
	ComboTotal          float64
	DiscountAmount      float64
	TaxCategoryID       pgtype.Int4
	TaxRate             float64
	TaxAmount           float64
	Subtotal            float64
	LineTotal           float64
	CustomizationJSON   json.RawMessage
	Notes               string
	Status              string
	MetadataJSON        json.RawMessage
}

func clampLineDiscount(qty, unitPrice, modifiersTotal, comboTotal, discount float64) float64 {
	if discount < 0 {
		return 0
	}
	gross := qty * (unitPrice + modifiersTotal + comboTotal)
	if discount > gross {
		return gross
	}
	return discount
}

func computeLineTaxAndTotals(qty, unitPrice, modifiersTotal, comboTotal, discount float64, tc *repository.TaxCategory) (taxRate, taxAmount, subtotal, lineTotal float64) {
	effectiveUnit := unitPrice + modifiersTotal + comboTotal
	gross := qty * effectiveUnit
	discount = clampLineDiscount(qty, unitPrice, modifiersTotal, comboTotal, discount)
	netAfterDiscount := gross - discount

	if tc != nil {
		taxRate = numericToFloat(tc.TaxRate)
		if taxRate > 0 {
			if tc.IsInclusive.Valid && tc.IsInclusive.Bool {
				taxAmount = roundMoney(netAfterDiscount * (taxRate / (100.0 + taxRate)))
				lineTotal = roundMoney(netAfterDiscount)
				subtotal = roundMoney(lineTotal - taxAmount)
				return
			}
			taxAmount = roundMoney(netAfterDiscount * (taxRate / 100.0))
			lineTotal = roundMoney(netAfterDiscount + taxAmount)
			subtotal = roundMoney(netAfterDiscount)
			return
		}
	}

	lineTotal = roundMoney(netAfterDiscount)
	subtotal = lineTotal
	taxAmount = 0
	return
}

// resolveRestaurantLineItem validates a menu item, its modifiers, combo selections, and tax
// category against the database and computes authoritative monetary totals.
func resolveRestaurantLineItem(ctx context.Context, q *repository.Queries, storeID int32, item RestaurantLineItemInput) (*resolvedRestaurantLine, *repository.Response) {
	if item.MenuItemID <= 0 {
		return nil, utils.NewResponse(utils.CodeBadReq, "menu_item_id is required", nil)
	}
	if item.Quantity <= 0 {
		return nil, utils.NewResponse(utils.CodeBadReq, "quantity must be greater than 0", nil)
	}

	mi, err := q.GetMenuItem(ctx, item.MenuItemID)
	if err != nil {
		return nil, utils.NewResponse(utils.CodeNotFound, fmt.Sprintf("menu item %d not found", item.MenuItemID), nil)
	}
	if storeID > 0 && mi.StoreID != storeID {
		return nil, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("menu item %d does not belong to store %d", item.MenuItemID, storeID), nil)
	}
	if mi.IsActive.Valid && !mi.IsActive.Bool {
		return nil, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("menu item %q is inactive", mi.Name), nil)
	}
	if mi.IsAvailable.Valid && !mi.IsAvailable.Bool {
		return nil, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("menu item %q is currently unavailable", mi.Name), nil)
	}

	unitPrice := numericToFloat(mi.BasePrice)
	if item.UnitPrice != nil && *item.UnitPrice >= 0 {
		unitPrice = *item.UnitPrice
	}

	// Resolve modifiers against menu_item_modifiers
	var resolvedModifiers []RestaurantModifierSnapshotItem
	var modifiersTotal float64
	if len(item.ModifierIDs) > 0 || len(item.Modifiers) > 0 {
		availableMods, err := q.ListMenuItemModifiers(ctx, mi.ID)
		if err != nil {
			return nil, utils.NewResponse(utils.CodeError, "failed to load menu item modifiers: "+err.Error(), nil)
		}
		modByID := make(map[int32]repository.MenuItemModifier, len(availableMods))
		for _, m := range availableMods {
			modByID[m.ID] = m
		}

		for _, modID := range item.ModifierIDs {
			m, ok := modByID[modID]
			if !ok {
				return nil, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("modifier %d is not valid for menu item %d", modID, mi.ID), nil)
			}
			if m.IsActive.Valid && !m.IsActive.Bool {
				return nil, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("modifier %q is inactive", m.ModifierName), nil)
			}
			adj := numericToFloat(m.PriceAdjustment)
			modifiersTotal += adj
			resolvedModifiers = append(resolvedModifiers, RestaurantModifierSnapshotItem{
				ModifierID:      m.ID,
				ModifierName:    m.ModifierName,
				ModifierType:    m.ModifierType,
				PriceAdjustment: adj,
			})
		}

		for _, mInput := range item.Modifiers {
			if mInput.ModifierID > 0 {
				if m, ok := modByID[mInput.ModifierID]; ok {
					if m.IsActive.Valid && !m.IsActive.Bool {
						return nil, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("modifier %q is inactive", m.ModifierName), nil)
					}
					adj := numericToFloat(m.PriceAdjustment)
					modifiersTotal += adj
					resolvedModifiers = append(resolvedModifiers, RestaurantModifierSnapshotItem{
						ModifierID:      m.ID,
						ModifierName:    m.ModifierName,
						ModifierType:    m.ModifierType,
						PriceAdjustment: adj,
					})
					continue
				}
				return nil, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("modifier %d is not valid for menu item %d", mInput.ModifierID, mi.ID), nil)
			}
			modifiersTotal += mInput.PriceAdjustment
			resolvedModifiers = append(resolvedModifiers, mInput)
		}
	}
	if resolvedModifiers == nil {
		resolvedModifiers = []RestaurantModifierSnapshotItem{}
	}
	modJSON, _ := json.Marshal(resolvedModifiers)

	// Resolve combo components against menu_item_combo_components
	var resolvedCombos []RestaurantComboComponentSnapshotItem
	var comboTotal float64
	if len(item.ComboComponentIDs) > 0 || len(item.ComboComponents) > 0 {
		availableCombos, err := q.ListMenuItemComboComponentsByParentID(ctx, mi.ID)
		if err != nil {
			return nil, utils.NewResponse(utils.CodeError, "failed to load combo components: "+err.Error(), nil)
		}
		comboByID := make(map[int32]repository.ListMenuItemComboComponentsByParentIDRow, len(availableCombos))
		comboByComponentMenuID := make(map[int32]repository.ListMenuItemComboComponentsByParentIDRow, len(availableCombos))
		for _, c := range availableCombos {
			comboByID[c.ID] = c
			comboByComponentMenuID[c.ComponentMenuItemID] = c
		}

		resolveComboID := func(id int32) (repository.ListMenuItemComboComponentsByParentIDRow, bool) {
			if c, ok := comboByID[id]; ok {
				return c, true
			}
			c, ok := comboByComponentMenuID[id]
			return c, ok
		}

		for _, cID := range item.ComboComponentIDs {
			c, ok := resolveComboID(cID)
			if !ok {
				return nil, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("combo component %d is not valid for menu item %d", cID, mi.ID), nil)
			}
			adj := numericToFloat(c.PriceAdjustment)
			comboTotal += adj
			resolvedCombos = append(resolvedCombos, RestaurantComboComponentSnapshotItem{
				ComboComponentID:    c.ID,
				ComponentMenuItemID: c.ComponentMenuItemID,
				ComponentName:       c.ComponentName,
				GroupName:           c.GroupName,
				PriceAdjustment:     adj,
			})
		}

		for _, cInput := range item.ComboComponents {
			lookupID := cInput.ComboComponentID
			if lookupID == 0 {
				lookupID = cInput.ComponentMenuItemID
			}
			if c, ok := resolveComboID(lookupID); ok {
				adj := numericToFloat(c.PriceAdjustment)
				comboTotal += adj
				resolvedCombos = append(resolvedCombos, RestaurantComboComponentSnapshotItem{
					ComboComponentID:    c.ID,
					ComponentMenuItemID: c.ComponentMenuItemID,
					ComponentName:       c.ComponentName,
					GroupName:           c.GroupName,
					PriceAdjustment:     adj,
				})
				continue
			}
			return nil, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("combo component %d is not valid for menu item %d", lookupID, mi.ID), nil)
		}
	}
	if resolvedCombos == nil {
		resolvedCombos = []RestaurantComboComponentSnapshotItem{}
	}
	comboJSON, _ := json.Marshal(resolvedCombos)

	var tcPtr *repository.TaxCategory
	if mi.TaxCategoryID.Valid {
		tc, err := q.GetTaxCategory(ctx, mi.TaxCategoryID.Int32)
		if err == nil && (!tc.IsActive.Valid || tc.IsActive.Bool) {
			tcPtr = &tc
		}
	}

	cappedDiscount := roundMoney(clampLineDiscount(
		item.Quantity,
		unitPrice,
		roundMoney(modifiersTotal),
		roundMoney(comboTotal),
		item.DiscountAmount,
	))

	taxRate, taxAmount, subtotal, lineTotal := computeLineTaxAndTotals(
		item.Quantity,
		unitPrice,
		roundMoney(modifiersTotal),
		roundMoney(comboTotal),
		cappedDiscount,
		tcPtr,
	)

	status := strings.TrimSpace(item.Status)
	if status == "" {
		status = "pending"
	}

	return &resolvedRestaurantLine{
		MenuItem:          mi,
		Quantity:          item.Quantity,
		UnitPrice:         unitPrice,
		ModifiersSnapshot: modJSON,
		ModifiersTotal:    roundMoney(modifiersTotal),
		ComboSnapshot:     comboJSON,
		ComboTotal:        roundMoney(comboTotal),
		DiscountAmount:    cappedDiscount,
		TaxCategoryID:     mi.TaxCategoryID,
		TaxRate:           taxRate,
		TaxAmount:         taxAmount,
		Subtotal:          subtotal,
		LineTotal:         lineTotal,
		CustomizationJSON: ensureJSONObject(item.CustomizationDetails),
		Notes:             item.Notes,
		Status:            status,
		MetadataJSON:      ensureJSONObject(item.Metadata),
	}, nil
}

func resolveStoreOrganizationID(ctx context.Context, q *repository.Queries, storeID int32, providedOrgID *int32) pgtype.Int4 {
	if providedOrgID != nil && *providedOrgID > 0 {
		return pgtype.Int4{Int32: *providedOrgID, Valid: true}
	}
	if storeID > 0 {
		store, err := q.GetStore(ctx, storeID)
		if err == nil && store.OrganizationID > 0 {
			return pgtype.Int4{Int32: store.OrganizationID, Valid: true}
		}
	}
	return pgtype.Int4{Valid: false}
}

func buildCartDetailPayload(ctx context.Context, q *repository.Queries, cart repository.RestaurantCart) (map[string]interface{}, error) {
	itemRows, err := q.ListRestaurantCartItems(ctx, cart.ID)
	if err != nil {
		return nil, err
	}
	totalsRow, err := q.GetRestaurantCartTotals(ctx, cart.ID)
	var totals RestaurantCartTotalsOutput
	if err == nil {
		totals = RestaurantCartTotalsOutput{
			ItemCount:     totalsRow.ItemCount,
			TotalQuantity: numericToFloat(totalsRow.TotalQuantity),
			Subtotal:      roundMoney(numericToFloat(totalsRow.Subtotal)),
			TotalDiscount: roundMoney(numericToFloat(totalsRow.TotalDiscount)),
			TotalTax:      roundMoney(numericToFloat(totalsRow.TotalTax)),
			TotalAmount:   roundMoney(numericToFloat(totalsRow.TotalAmount)),
		}
	}
	return map[string]interface{}{
		"cart":   restaurantCartToOutput(cart),
		"items":  restaurantCartItemRowsToOutput(itemRows),
		"totals": totals,
	}, nil
}

// === Public Cart Lifecycle Methods ===

func (uc *RestaurantCartUseCase) CreateCart(ctx context.Context, input RestaurantCreateCartInput) *repository.Response {
	if input.StoreID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "store_id is required", nil)
	}

	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		cartNumber := strings.TrimSpace(input.CartNumber)
		if cartNumber == "" {
			cartNumber = generateRestaurantRef("RCART")
		}
		orderType := strings.TrimSpace(input.OrderType)
		if orderType == "" {
			if input.TableID != nil && *input.TableID > 0 {
				orderType = "dine_in"
			} else {
				orderType = "takeaway"
			}
		}
		cartStatus := repository.CartStatus(strings.TrimSpace(input.CartStatus))
		if cartStatus == "" {
			cartStatus = repository.CartStatusActive
		}
		channel := strings.TrimSpace(input.Channel)
		if channel == "" {
			channel = "pos"
		}

		orgID := resolveStoreOrganizationID(ctx, q, input.StoreID, input.OrganizationID)

		var expiresAt pgtype.Timestamp
		if input.ExpiresAt != nil {
			expiresAt = pgtype.Timestamp{Time: *input.ExpiresAt, Valid: true}
		}

		cart, err := q.CreateRestaurantCart(ctx, repository.CreateRestaurantCartParams{
			CartNumber:      pgtype.Text{String: cartNumber, Valid: true},
			OrganizationID:  orgID,
			StoreID:         input.StoreID,
			TableID:         utils.Int32ToPgInt4(input.TableID),
			OrderType:       orderType,
			CustomerID:      utils.Int32ToPgInt4(input.CustomerID),
			GuestIdentifier: nullableText(input.GuestIdentifier),
			GuestName:       nullableText(input.GuestName),
			GuestEmail:      nullableText(input.GuestEmail),
			GuestPhone:      nullableText(input.GuestPhone),
			CartStatus:      cartStatus,
			Channel:         pgtype.Text{String: channel, Valid: true},
			PaymentMethod:   nullableText(input.PaymentMethod),
			PaymentGateway:  nullableText(input.PaymentGateway),
			DeviceInfo:      ensureJSONObject(input.DeviceInfo),
			CreatedByUserID: utils.Int32ToPgInt4(input.CreatedByUserID),
			CashierID:       utils.Int32ToPgInt4(input.CashierID),
			PosTerminalID:   utils.Int32ToPgInt4(input.PosTerminalID),
			KioskSessionID:  utils.Int32ToPgInt4(input.KioskSessionID),
			CouponCode:      nullableText(input.CouponCode),
			DiscountCode:    nullableText(input.DiscountCode),
			DeliveryAddress: input.DeliveryAddress,
			ExpiresAt:       expiresAt,
			Notes:           nullableText(input.Notes),
			Metadata:        ensureJSONObject(input.Metadata),
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to create restaurant cart: "+err.Error(), nil)
		}

		for _, itemInput := range input.Items {
			resolved, errResp := resolveRestaurantLineItem(ctx, q, cart.StoreID, itemInput)
			if errResp != nil {
				return errResp
			}
			_, err := insertOrMergeCartItem(ctx, q, cart, resolved)
			if err != nil {
				return utils.NewResponse(utils.CodeError, "failed to add initial item to restaurant cart: "+err.Error(), nil)
			}
		}

		if len(input.Items) > 0 {
			updatedCart, err := q.RecalculateRestaurantCartTotals(ctx, cart.ID)
			if err != nil {
				return utils.NewResponse(utils.CodeError, "failed to recalculate cart totals: "+err.Error(), nil)
			}
			cart = updatedCart
			if strings.TrimSpace(input.CouponCode) != "" {
				promoCart, _, errResp := applyRestaurantCouponInTx(ctx, q, cart, input.CouponCode)
				if errResp != nil {
					return errResp
				}
				cart = promoCart
			}
		}

		payload, err := buildCartDetailPayload(ctx, q, cart)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to load restaurant cart details: "+err.Error(), nil)
		}
		return utils.NewResponse(utils.CodeCreated, "restaurant cart created successfully", payload)
	})
}

func (uc *RestaurantCartUseCase) GetCart(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}
	cart, err := uc.repo.GetRestaurantCartByID(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "restaurant cart not found", nil)
	}
	payload, err := buildCartDetailPayload(ctx, uc.repo, cart)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to load restaurant cart items: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "restaurant cart retrieved successfully", payload)
}

func (uc *RestaurantCartUseCase) GetActiveCart(ctx context.Context, storeID int32, customerID *int32, guestIdentifier string, tableID *int32) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}
	if storeID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "store_id is required", nil)
	}

	var cart repository.RestaurantCart
	var err error
	switch {
	case tableID != nil && *tableID > 0:
		cart, err = uc.repo.GetRestaurantCartByTable(ctx, repository.GetRestaurantCartByTableParams{
			TableID: pgtype.Int4{Int32: *tableID, Valid: true},
			StoreID: storeID,
		})
	case customerID != nil && *customerID > 0:
		cart, err = uc.repo.GetRestaurantCartByCustomer(ctx, repository.GetRestaurantCartByCustomerParams{
			CustomerID: pgtype.Int4{Int32: *customerID, Valid: true},
			StoreID:    storeID,
		})
	case strings.TrimSpace(guestIdentifier) != "":
		cart, err = uc.repo.GetRestaurantCartByGuestIdentifier(ctx, repository.GetRestaurantCartByGuestIdentifierParams{
			GuestIdentifier: pgtype.Text{String: strings.TrimSpace(guestIdentifier), Valid: true},
			StoreID:         storeID,
		})
	default:
		return utils.NewResponse(utils.CodeBadReq, "table_id, customer_id, or guest_identifier is required", nil)
	}

	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "active restaurant cart not found", nil)
	}

	payload, err := buildCartDetailPayload(ctx, uc.repo, cart)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to load restaurant cart items: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "active restaurant cart retrieved successfully", payload)
}

func (uc *RestaurantCartUseCase) ListActiveCarts(ctx context.Context, storeID, limit, offset int32) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}
	if storeID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "store_id is required", nil)
	}
	if limit <= 0 {
		limit = 50
	}
	carts, err := uc.repo.ListActiveRestaurantCarts(ctx, repository.ListActiveRestaurantCartsParams{
		StoreID: storeID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list active restaurant carts: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "active restaurant carts listed successfully", restaurantCartsToOutput(carts))
}

func (uc *RestaurantCartUseCase) ListDraftCartsByCashier(ctx context.Context, cashierID int32, storeID *int32, limit, offset int32) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}
	if limit <= 0 {
		limit = 50
	}
	carts, err := uc.repo.ListDraftRestaurantCartsByCashier(ctx, repository.ListDraftRestaurantCartsByCashierParams{
		CashierID: utils.Int32ToPgInt4(&cashierID),
		StoreID:   utils.Int32ToPgInt4(storeID),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list draft restaurant carts: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "draft restaurant carts listed successfully", restaurantCartsToOutput(carts))
}

func (uc *RestaurantCartUseCase) UpdateCart(ctx context.Context, arg repository.UpdateRestaurantCartParams) *repository.Response {
	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		existing, err := lockAndGetRestaurantCartByID(ctx, q, arg.ID)
		if err != nil {
			return restaurantCartNotFoundOrErr(err, "lock restaurant cart")
		}
		if existing.CartStatus == repository.CartStatusConverted {
			return utils.NewResponse(utils.CodeBadReq, "cannot modify an already converted cart", nil)
		}
		cart, err := q.UpdateRestaurantCart(ctx, arg)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to update restaurant cart: "+err.Error(), nil)
		}
		payload, err := buildCartDetailPayload(ctx, q, cart)
		if err != nil {
			return utils.NewResponse(utils.CodeOK, "restaurant cart updated successfully", restaurantCartToOutput(cart))
		}
		return utils.NewResponse(utils.CodeOK, "restaurant cart updated successfully", payload)
	})
}

func (uc *RestaurantCartUseCase) AssignCartTable(ctx context.Context, cartID uuid.UUID, tableID *int32) *repository.Response {
	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		existing, err := lockAndGetRestaurantCartByID(ctx, q, cartID)
		if err != nil {
			return restaurantCartNotFoundOrErr(err, "lock restaurant cart")
		}
		if existing.CartStatus == repository.CartStatusConverted {
			return utils.NewResponse(utils.CodeBadReq, "cannot assign table on an already converted cart", nil)
		}
		cart, err := q.UpdateRestaurantCartTable(ctx, repository.UpdateRestaurantCartTableParams{
			ID:      cartID,
			TableID: utils.Int32ToPgInt4(tableID),
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to update restaurant cart table: "+err.Error(), nil)
		}
		return utils.NewResponse(utils.CodeOK, "restaurant cart table updated", restaurantCartToOutput(cart))
	})
}

func (uc *RestaurantCartUseCase) HoldCart(ctx context.Context, cartID uuid.UUID) *repository.Response {
	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		existing, err := lockAndGetRestaurantCartByID(ctx, q, cartID)
		if err != nil {
			return restaurantCartNotFoundOrErr(err, "lock restaurant cart")
		}
		if existing.CartStatus == repository.CartStatusConverted {
			return utils.NewResponse(utils.CodeBadReq, "cannot hold an already converted cart", nil)
		}
		cart, err := q.UpdateRestaurantCartStatus(ctx, repository.UpdateRestaurantCartStatusParams{
			ID:         cartID,
			CartStatus: repository.CartStatusDraft,
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to hold restaurant cart: "+err.Error(), nil)
		}
		return utils.NewResponse(utils.CodeOK, "restaurant cart put on hold", restaurantCartToOutput(cart))
	})
}

func (uc *RestaurantCartUseCase) DeleteCart(ctx context.Context, cartID uuid.UUID) *repository.Response {
	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		if _, err := q.LockRestaurantCart(ctx, cartID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return utils.NewResponse(utils.CodeOK, "restaurant cart deleted successfully", nil)
			}
			return utils.NewResponse(utils.CodeError, "failed to lock restaurant cart: "+err.Error(), nil)
		}
		if err := q.DeleteRestaurantCart(ctx, cartID); err != nil {
			return utils.NewResponse(utils.CodeError, "failed to delete restaurant cart: "+err.Error(), nil)
		}
		return utils.NewResponse(utils.CodeOK, "restaurant cart deleted successfully", nil)
	})
}

func insertOrMergeCartItem(ctx context.Context, q *repository.Queries, cart repository.RestaurantCart, resolved *resolvedRestaurantLine) (repository.RestaurantCartItem, error) {
	existing, err := q.FindMatchingRestaurantCartItem(ctx, repository.FindMatchingRestaurantCartItemParams{
		CartID:     cart.ID,
		MenuItemID: resolved.MenuItem.ID,
		Column3:    resolved.ModifiersSnapshot,
		Column4:    resolved.ComboSnapshot,
		Column5:    resolved.Notes,
		UnitPrice:  utils.Float64ToPgNumeric(resolved.UnitPrice),
	})
	if err == nil {
		newQty := numericToFloat(existing.Quantity) + resolved.Quantity
		newDiscount := roundMoney(clampLineDiscount(
			newQty,
			numericToFloat(existing.UnitPrice),
			numericToFloat(existing.ModifiersTotal),
			numericToFloat(existing.ComboTotal),
			numericToFloat(existing.DiscountAmount)+resolved.DiscountAmount,
		))

		var tcPtr *repository.TaxCategory
		if resolved.TaxCategoryID.Valid {
			tc, tcErr := q.GetTaxCategory(ctx, resolved.TaxCategoryID.Int32)
			if tcErr == nil {
				tcPtr = &tc
			}
		}
		taxRate, taxAmount, _, lineTotal := computeLineTaxAndTotals(
			newQty,
			numericToFloat(existing.UnitPrice),
			numericToFloat(existing.ModifiersTotal),
			numericToFloat(existing.ComboTotal),
			newDiscount,
			tcPtr,
		)

		return q.UpdateRestaurantCartItem(ctx, repository.UpdateRestaurantCartItemParams{
			ID:                      existing.ID,
			Quantity:                utils.Float64ToPgNumeric(newQty),
			UnitPrice:               existing.UnitPrice,
			ModifiersSnapshot:       existing.ModifiersSnapshot,
			ModifiersTotal:          existing.ModifiersTotal,
			ComboComponentsSnapshot: existing.ComboComponentsSnapshot,
			ComboTotal:              existing.ComboTotal,
			DiscountAmount:          utils.Float64ToPgNumeric(newDiscount),
			TaxCategoryID:           resolved.TaxCategoryID,
			TaxRate:                 utils.Float64ToPgNumeric(taxRate),
			TaxAmount:               utils.Float64ToPgNumeric(taxAmount),
			LineTotal:               utils.Float64ToPgNumeric(lineTotal),
			CustomizationDetails:    resolved.CustomizationJSON,
			Notes:                   nullableText(resolved.Notes),
			Metadata:                mergeCartItemMetadataPreservingCoupon(existing.Metadata, resolved.MetadataJSON),
		})
	}

	return q.CreateRestaurantCartItem(ctx, repository.CreateRestaurantCartItemParams{
		CartID:                  cart.ID,
		OrganizationID:          cart.OrganizationID,
		MenuItemID:              resolved.MenuItem.ID,
		Quantity:                utils.Float64ToPgNumeric(resolved.Quantity),
		UnitPrice:               utils.Float64ToPgNumeric(resolved.UnitPrice),
		ModifiersSnapshot:       resolved.ModifiersSnapshot,
		ModifiersTotal:          utils.Float64ToPgNumeric(resolved.ModifiersTotal),
		ComboComponentsSnapshot: resolved.ComboSnapshot,
		ComboTotal:              utils.Float64ToPgNumeric(resolved.ComboTotal),
		DiscountAmount:          utils.Float64ToPgNumeric(resolved.DiscountAmount),
		TaxCategoryID:           resolved.TaxCategoryID,
		TaxRate:                 utils.Float64ToPgNumeric(resolved.TaxRate),
		TaxAmount:               utils.Float64ToPgNumeric(resolved.TaxAmount),
		LineTotal:               utils.Float64ToPgNumeric(resolved.LineTotal),
		CustomizationDetails:    resolved.CustomizationJSON,
		Notes:                   nullableText(resolved.Notes),
		Metadata:                resolved.MetadataJSON,
	})
}

// AddToCart adds a menu item (with modifiers/combos/notes) to an existing or newly created
// restaurant cart and recalculates cart totals atomically.
func (uc *RestaurantCartUseCase) AddToCart(ctx context.Context, input RestaurantAddToCartInput) *repository.Response {
	if input.MenuItemID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "menu_item_id is required", nil)
	}
	if input.Quantity <= 0 {
		input.Quantity = 1
	}

	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		var cart repository.RestaurantCart
		var err error

		if input.CartID != nil && *input.CartID != uuid.Nil {
			cart, err = lockAndGetRestaurantCartByID(ctx, q, *input.CartID)
			if err != nil {
				return restaurantCartNotFoundOrErr(err, "lock restaurant cart")
			}
			if cart.CartStatus == repository.CartStatusConverted {
				return utils.NewResponse(utils.CodeBadReq, "cannot add items to an already converted cart", nil)
			}
		} else {
			if input.StoreID <= 0 {
				return utils.NewResponse(utils.CodeBadReq, "store_id or cart_id is required", nil)
			}
			var scopeKey string
			switch {
			case input.TableID != nil && *input.TableID > 0:
				scopeKey = fmt.Sprintf("rest_cart:store:%d:table:%d", input.StoreID, *input.TableID)
			case input.CustomerID != nil && *input.CustomerID > 0:
				scopeKey = fmt.Sprintf("rest_cart:store:%d:customer:%d", input.StoreID, *input.CustomerID)
			case strings.TrimSpace(input.GuestIdentifier) != "":
				scopeKey = fmt.Sprintf("rest_cart:store:%d:guest:%s", input.StoreID, strings.TrimSpace(input.GuestIdentifier))
			}
			if scopeKey != "" {
				if lockErr := q.AcquireRestaurantCartScopeLock(ctx, scopeKey); lockErr != nil {
					return utils.NewResponse(utils.CodeError, "failed to lock restaurant cart scope: "+lockErr.Error(), nil)
				}
			}
			// Locate existing active cart by table -> customer -> guest_identifier and lock it before reading items
			found := false
			tryLockActive := func(candidate repository.RestaurantCart) {
				locked, lockErr := lockAndGetRestaurantCartByID(ctx, q, candidate.ID)
				if lockErr == nil && locked.CartStatus == repository.CartStatusActive {
					cart = locked
					found = true
				}
			}
			if input.TableID != nil && *input.TableID > 0 {
				if existing, findErr := q.GetRestaurantCartByTable(ctx, repository.GetRestaurantCartByTableParams{
					TableID: pgtype.Int4{Int32: *input.TableID, Valid: true},
					StoreID: input.StoreID,
				}); findErr == nil {
					tryLockActive(existing)
				}
			}
			if !found && input.CustomerID != nil && *input.CustomerID > 0 {
				if existing, findErr := q.GetRestaurantCartByCustomer(ctx, repository.GetRestaurantCartByCustomerParams{
					CustomerID: pgtype.Int4{Int32: *input.CustomerID, Valid: true},
					StoreID:    input.StoreID,
				}); findErr == nil {
					tryLockActive(existing)
				}
			}
			if !found && strings.TrimSpace(input.GuestIdentifier) != "" {
				if existing, findErr := q.GetRestaurantCartByGuestIdentifier(ctx, repository.GetRestaurantCartByGuestIdentifierParams{
					GuestIdentifier: pgtype.Text{String: strings.TrimSpace(input.GuestIdentifier), Valid: true},
					StoreID:         input.StoreID,
				}); findErr == nil {
					tryLockActive(existing)
				}
			}

			if !found {
				orderType := strings.TrimSpace(input.OrderType)
				if orderType == "" {
					if input.TableID != nil && *input.TableID > 0 {
						orderType = "dine_in"
					} else {
						orderType = "takeaway"
					}
				}
				channel := strings.TrimSpace(input.Channel)
				if channel == "" {
					channel = "pos"
				}
				orgID := resolveStoreOrganizationID(ctx, q, input.StoreID, input.OrganizationID)
				cart, err = q.CreateRestaurantCart(ctx, repository.CreateRestaurantCartParams{
					CartNumber:      pgtype.Text{String: generateRestaurantRef("RCART"), Valid: true},
					OrganizationID:  orgID,
					StoreID:         input.StoreID,
					TableID:         utils.Int32ToPgInt4(input.TableID),
					OrderType:       orderType,
					CustomerID:      utils.Int32ToPgInt4(input.CustomerID),
					GuestIdentifier: nullableText(input.GuestIdentifier),
					GuestName:       nullableText(input.GuestName),
					GuestEmail:      nullableText(input.GuestEmail),
					GuestPhone:      nullableText(input.GuestPhone),
					CartStatus:      repository.CartStatusActive,
					Channel:         pgtype.Text{String: channel, Valid: true},
					DeviceInfo:      json.RawMessage("{}"),
					CreatedByUserID: utils.Int32ToPgInt4(input.CreatedByUserID),
					CashierID:       utils.Int32ToPgInt4(input.CashierID),
					PosTerminalID:   utils.Int32ToPgInt4(input.PosTerminalID),
					KioskSessionID:  utils.Int32ToPgInt4(input.KioskSessionID),
					Metadata:        json.RawMessage("{}"),
				})
				if err != nil {
					return utils.NewResponse(utils.CodeError, "failed to create restaurant cart: "+err.Error(), nil)
				}
			}
		}

		resolved, errResp := resolveRestaurantLineItem(ctx, q, cart.StoreID, RestaurantLineItemInput{
			MenuItemID:           input.MenuItemID,
			Quantity:             input.Quantity,
			UnitPrice:            input.UnitPrice,
			ModifierIDs:          input.ModifierIDs,
			Modifiers:            input.Modifiers,
			ComboComponentIDs:    input.ComboComponentIDs,
			ComboComponents:      input.ComboComponents,
			DiscountAmount:       input.DiscountAmount,
			CustomizationDetails: input.CustomizationDetails,
			Notes:                input.Notes,
			Metadata:             input.Metadata,
		})
		if errResp != nil {
			return errResp
		}

		item, err := insertOrMergeCartItem(ctx, q, cart, resolved)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to add item to restaurant cart: "+err.Error(), nil)
		}

		updatedCart, err := q.RecalculateRestaurantCartTotals(ctx, cart.ID)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to recalculate restaurant cart totals: "+err.Error(), nil)
		}

		payload, err := buildCartDetailPayload(ctx, q, updatedCart)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to build cart response: "+err.Error(), nil)
		}
		payload["item"] = restaurantCartItemToOutput(item)
		return utils.NewResponse(utils.CodeOK, "item added to restaurant cart", payload)
	})
}

func (uc *RestaurantCartUseCase) UpdateCartItemQuantity(ctx context.Context, input RestaurantUpdateCartItemInput) *repository.Response {
	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		initialItem, err := q.GetRestaurantCartItem(ctx, input.ItemID)
		if err != nil {
			return utils.NewResponse(utils.CodeNotFound, "restaurant cart item not found", nil)
		}

		cart, err := lockAndGetRestaurantCartByID(ctx, q, initialItem.CartID)
		if err != nil {
			return restaurantCartNotFoundOrErr(err, "lock restaurant cart")
		}
		if cart.CartStatus == repository.CartStatusConverted {
			return utils.NewResponse(utils.CodeBadReq, "cannot modify items in a converted cart", nil)
		}

		existing, err := q.GetRestaurantCartItem(ctx, input.ItemID)
		if err != nil {
			return utils.NewResponse(utils.CodeNotFound, "restaurant cart item not found", nil)
		}

		if input.Quantity <= 0 {
			if err := q.DeleteRestaurantCartItem(ctx, input.ItemID); err != nil {
				return utils.NewResponse(utils.CodeError, "failed to delete cart item: "+err.Error(), nil)
			}
			updatedCart, err := q.RecalculateRestaurantCartTotals(ctx, cart.ID)
			if err != nil {
				return utils.NewResponse(utils.CodeError, "failed to recalculate cart totals: "+err.Error(), nil)
			}
			payload, err := buildCartDetailPayload(ctx, q, updatedCart)
			if err != nil {
				return utils.NewResponse(utils.CodeOK, "cart item removed", nil)
			}
			return utils.NewResponse(utils.CodeOK, "cart item removed", payload)
		}

		discount := numericToFloat(existing.DiscountAmount)
		if input.DiscountAmount != nil {
			discount = *input.DiscountAmount
		}
		notes := utils.DerefString(utils.TextToPtr(existing.Notes))
		if input.Notes != nil {
			notes = *input.Notes
		}
		mergedMeta := reconcileCartItemCouponMetadata(
			mergeCartItemMetadataPreservingCoupon(existing.Metadata, input.Metadata),
			input.DiscountAmount,
		)

		var updatedItem repository.RestaurantCartItem
		if input.HasModifiersUpdate || input.HasCombosUpdate {
			var modSnapshot []RestaurantModifierSnapshotItem
			var comboSnapshot []RestaurantComboComponentSnapshotItem
			if !input.HasModifiersUpdate {
				_ = json.Unmarshal(existing.ModifiersSnapshot, &modSnapshot)
			} else {
				modSnapshot = input.Modifiers
			}
			if !input.HasCombosUpdate {
				_ = json.Unmarshal(existing.ComboComponentsSnapshot, &comboSnapshot)
			} else {
				comboSnapshot = input.ComboComponents
			}

			savedUnitPrice := numericToFloat(existing.UnitPrice)
			resolved, errResp := resolveRestaurantLineItem(ctx, q, cart.StoreID, RestaurantLineItemInput{
				MenuItemID:           existing.MenuItemID,
				Quantity:             input.Quantity,
				UnitPrice:            &savedUnitPrice,
				ModifierIDs:          input.ModifierIDs,
				Modifiers:            modSnapshot,
				ComboComponentIDs:    input.ComboComponentIDs,
				ComboComponents:      comboSnapshot,
				DiscountAmount:       discount,
				CustomizationDetails: input.CustomizationDetails,
				Notes:                notes,
				Metadata:             mergedMeta,
			})
			if errResp != nil {
				return errResp
			}
			updatedItem, err = q.UpdateRestaurantCartItem(ctx, repository.UpdateRestaurantCartItemParams{
				ID:                      existing.ID,
				Quantity:                utils.Float64ToPgNumeric(resolved.Quantity),
				UnitPrice:               utils.Float64ToPgNumeric(resolved.UnitPrice),
				ModifiersSnapshot:       resolved.ModifiersSnapshot,
				ModifiersTotal:          utils.Float64ToPgNumeric(resolved.ModifiersTotal),
				ComboComponentsSnapshot: resolved.ComboSnapshot,
				ComboTotal:              utils.Float64ToPgNumeric(resolved.ComboTotal),
				DiscountAmount:          utils.Float64ToPgNumeric(resolved.DiscountAmount),
				TaxCategoryID:           resolved.TaxCategoryID,
				TaxRate:                 utils.Float64ToPgNumeric(resolved.TaxRate),
				TaxAmount:               utils.Float64ToPgNumeric(resolved.TaxAmount),
				LineTotal:               utils.Float64ToPgNumeric(resolved.LineTotal),
				CustomizationDetails:    resolved.CustomizationJSON,
				Notes:                   nullableText(resolved.Notes),
				Metadata:                mergedMeta,
			})
		} else {
			var tcPtr *repository.TaxCategory
			if existing.TaxCategoryID.Valid {
				if tc, tcErr := q.GetTaxCategory(ctx, existing.TaxCategoryID.Int32); tcErr == nil {
					tcPtr = &tc
				}
			}
			cappedDiscount := roundMoney(clampLineDiscount(
				input.Quantity,
				numericToFloat(existing.UnitPrice),
				numericToFloat(existing.ModifiersTotal),
				numericToFloat(existing.ComboTotal),
				discount,
			))
			taxRate, taxAmount, _, lineTotal := computeLineTaxAndTotals(
				input.Quantity,
				numericToFloat(existing.UnitPrice),
				numericToFloat(existing.ModifiersTotal),
				numericToFloat(existing.ComboTotal),
				cappedDiscount,
				tcPtr,
			)
			customDetails := existing.CustomizationDetails
			if len(input.CustomizationDetails) > 0 {
				customDetails = input.CustomizationDetails
			}
			updatedItem, err = q.UpdateRestaurantCartItem(ctx, repository.UpdateRestaurantCartItemParams{
				ID:                      existing.ID,
				Quantity:                utils.Float64ToPgNumeric(input.Quantity),
				UnitPrice:               existing.UnitPrice,
				ModifiersSnapshot:       existing.ModifiersSnapshot,
				ModifiersTotal:          existing.ModifiersTotal,
				ComboComponentsSnapshot: existing.ComboComponentsSnapshot,
				ComboTotal:              existing.ComboTotal,
				DiscountAmount:          utils.Float64ToPgNumeric(cappedDiscount),
				TaxCategoryID:           existing.TaxCategoryID,
				TaxRate:                 utils.Float64ToPgNumeric(taxRate),
				TaxAmount:               utils.Float64ToPgNumeric(taxAmount),
				LineTotal:               utils.Float64ToPgNumeric(lineTotal),
				CustomizationDetails:    ensureJSONObject(customDetails),
				Notes:                   nullableText(notes),
				Metadata:                mergedMeta,
			})
		}
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to update cart item: "+err.Error(), nil)
		}

		updatedCart, err := q.RecalculateRestaurantCartTotals(ctx, cart.ID)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to recalculate cart totals: "+err.Error(), nil)
		}
		payload, err := buildCartDetailPayload(ctx, q, updatedCart)
		if err != nil {
			return utils.NewResponse(utils.CodeOK, "cart item updated", restaurantCartItemToOutput(updatedItem))
		}
		payload["item"] = restaurantCartItemToOutput(updatedItem)
		return utils.NewResponse(utils.CodeOK, "cart item updated", payload)
	})
}

func (uc *RestaurantCartUseCase) RemoveCartItem(ctx context.Context, itemID uuid.UUID) *repository.Response {
	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		initialItem, err := q.GetRestaurantCartItem(ctx, itemID)
		if err != nil {
			return utils.NewResponse(utils.CodeNotFound, "restaurant cart item not found", nil)
		}
		cart, err := lockAndGetRestaurantCartByID(ctx, q, initialItem.CartID)
		if err != nil {
			return restaurantCartNotFoundOrErr(err, "lock restaurant cart")
		}
		if cart.CartStatus == repository.CartStatusConverted {
			return utils.NewResponse(utils.CodeBadReq, "cannot remove items from a converted cart", nil)
		}
		if _, err := q.GetRestaurantCartItem(ctx, itemID); err != nil {
			return utils.NewResponse(utils.CodeNotFound, "restaurant cart item not found", nil)
		}
		if err := q.DeleteRestaurantCartItem(ctx, itemID); err != nil {
			return utils.NewResponse(utils.CodeError, "failed to remove cart item: "+err.Error(), nil)
		}
		updatedCart, err := q.RecalculateRestaurantCartTotals(ctx, cart.ID)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to recalculate cart totals: "+err.Error(), nil)
		}
		payload, err := buildCartDetailPayload(ctx, q, updatedCart)
		if err != nil {
			return utils.NewResponse(utils.CodeOK, "cart item removed", nil)
		}
		return utils.NewResponse(utils.CodeOK, "cart item removed", payload)
	})
}

func (uc *RestaurantCartUseCase) ClearCart(ctx context.Context, cartID uuid.UUID) *repository.Response {
	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		cart, err := lockAndGetRestaurantCartByID(ctx, q, cartID)
		if err != nil {
			return restaurantCartNotFoundOrErr(err, "lock restaurant cart")
		}
		if cart.CartStatus == repository.CartStatusConverted {
			return utils.NewResponse(utils.CodeBadReq, "cannot clear a converted cart", nil)
		}
		if err := q.ClearRestaurantCartItems(ctx, cartID); err != nil {
			return utils.NewResponse(utils.CodeError, "failed to clear restaurant cart items: "+err.Error(), nil)
		}
		updatedCart, err := q.RecalculateRestaurantCartTotals(ctx, cartID)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to recalculate cart totals: "+err.Error(), nil)
		}
		payload, err := buildCartDetailPayload(ctx, q, updatedCart)
		if err != nil {
			return utils.NewResponse(utils.CodeOK, "restaurant cart cleared", nil)
		}
		return utils.NewResponse(utils.CodeOK, "restaurant cart cleared", payload)
	})
}

func (uc *RestaurantCartUseCase) MergeGuestCart(ctx context.Context, storeID int32, guestIdentifier string, customerID int32) *repository.Response {
	if storeID <= 0 || strings.TrimSpace(guestIdentifier) == "" || customerID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "store_id, guest_identifier, and customer_id are required", nil)
	}

	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		guestCart, err := q.GetRestaurantCartByGuestIdentifier(ctx, repository.GetRestaurantCartByGuestIdentifierParams{
			GuestIdentifier: pgtype.Text{String: strings.TrimSpace(guestIdentifier), Valid: true},
			StoreID:         storeID,
		})
		if err != nil {
			return utils.NewResponse(utils.CodeNotFound, "guest restaurant cart not found", nil)
		}

		custCart, custErr := q.GetRestaurantCartByCustomer(ctx, repository.GetRestaurantCartByCustomerParams{
			CustomerID: pgtype.Int4{Int32: customerID, Valid: true},
			StoreID:    storeID,
		})
		if custErr != nil {
			lockedGuest, lockErr := lockAndGetRestaurantCartByID(ctx, q, guestCart.ID)
			if lockErr != nil {
				return restaurantCartNotFoundOrErr(lockErr, "lock guest cart")
			}
			if lockedGuest.CartStatus == repository.CartStatusConverted {
				return utils.NewResponse(utils.CodeBadReq, "cannot merge an already converted cart", nil)
			}
			// Assign guest cart directly to customer
			updated, err := q.UpdateRestaurantCartCustomer(ctx, repository.UpdateRestaurantCartCustomerParams{
				ID:         lockedGuest.ID,
				CustomerID: pgtype.Int4{Int32: customerID, Valid: true},
			})
			if err != nil {
				return utils.NewResponse(utils.CodeError, "failed to assign guest cart to customer: "+err.Error(), nil)
			}
			payload, err := buildCartDetailPayload(ctx, q, updated)
			if err != nil {
				return utils.NewResponse(utils.CodeOK, "guest cart assigned to customer", restaurantCartToOutput(updated))
			}
			return utils.NewResponse(utils.CodeOK, "guest cart assigned to customer", payload)
		}

		if guestCart.ID == custCart.ID {
			lockedCust, lockErr := lockAndGetRestaurantCartByID(ctx, q, custCart.ID)
			if lockErr != nil {
				return restaurantCartNotFoundOrErr(lockErr, "lock customer cart")
			}
			payload, _ := buildCartDetailPayload(ctx, q, lockedCust)
			return utils.NewResponse(utils.CodeOK, "cart already belongs to customer", payload)
		}

		// Acquire cart locks in deterministic UUID order to prevent deadlocks.
		firstID, secondID := guestCart.ID, custCart.ID
		if bytes.Compare(secondID[:], firstID[:]) < 0 {
			firstID, secondID = secondID, firstID
		}
		if _, lockErr := q.LockRestaurantCart(ctx, firstID); lockErr != nil {
			return restaurantCartNotFoundOrErr(lockErr, "lock restaurant cart")
		}
		if _, lockErr := q.LockRestaurantCart(ctx, secondID); lockErr != nil {
			return restaurantCartNotFoundOrErr(lockErr, "lock restaurant cart")
		}
		lockedGuest, err := q.GetRestaurantCartByID(ctx, guestCart.ID)
		if err != nil {
			return restaurantCartNotFoundOrErr(err, "reload guest cart")
		}
		lockedCust, err := q.GetRestaurantCartByID(ctx, custCart.ID)
		if err != nil {
			return restaurantCartNotFoundOrErr(err, "reload customer cart")
		}
		if lockedGuest.CartStatus == repository.CartStatusConverted || lockedCust.CartStatus == repository.CartStatusConverted {
			return utils.NewResponse(utils.CodeBadReq, "cannot merge an already converted cart", nil)
		}

		guestCoupon := ""
		if lockedGuest.CouponCode.Valid {
			guestCoupon = strings.TrimSpace(lockedGuest.CouponCode.String)
		}
		custCoupon := ""
		if lockedCust.CouponCode.Valid {
			custCoupon = strings.TrimSpace(lockedCust.CouponCode.String)
		}

		if err := q.MergeGuestRestaurantCartItems(ctx, repository.MergeGuestRestaurantCartItemsParams{
			CartID:   lockedGuest.ID,
			CartID_2: lockedCust.ID,
		}); err != nil {
			return utils.NewResponse(utils.CodeError, "failed to merge guest cart items: "+err.Error(), nil)
		}
		_ = q.DeleteRestaurantCart(ctx, lockedGuest.ID)

		updatedCart, err := q.RecalculateRestaurantCartTotals(ctx, lockedCust.ID)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to recalculate merged cart totals: "+err.Error(), nil)
		}
		if custCoupon != "" {
			if promoCart, _, errResp := applyRestaurantCouponInTx(ctx, q, updatedCart, custCoupon); errResp == nil {
				updatedCart = promoCart
			} else if cleared, clrErr := clearCouponDiscountsFromCartItems(ctx, q, updatedCart); clrErr == nil {
				updatedCart = cleared
			}
		} else if guestCoupon != "" {
			if promoCart, _, errResp := applyRestaurantCouponInTx(ctx, q, updatedCart, guestCoupon); errResp == nil {
				updatedCart = promoCart
			} else if cleared, clrErr := clearCouponDiscountsFromCartItems(ctx, q, updatedCart); clrErr == nil {
				updatedCart = cleared
			}
		}
		payload, err := buildCartDetailPayload(ctx, q, updatedCart)
		if err != nil {
			return utils.NewResponse(utils.CodeOK, "carts merged successfully", restaurantCartToOutput(updatedCart))
		}
		return utils.NewResponse(utils.CodeOK, "carts merged successfully", payload)
	})
}

// ConvertToOrder converts an active or draft restaurant cart into a restaurant_order +
// restaurant_order_items inside a single atomic database transaction.
func (uc *RestaurantCartUseCase) ConvertToOrder(ctx context.Context, input RestaurantCheckoutInput) *repository.Response {
	if input.CartID == uuid.Nil {
		return utils.NewResponse(utils.CodeBadReq, "cart_id is required", nil)
	}

	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		cart, err := lockAndGetRestaurantCartByID(ctx, q, input.CartID)
		if err != nil {
			return restaurantCartNotFoundOrErr(err, "lock restaurant cart")
		}
		if cart.CartStatus == repository.CartStatusConverted {
			return utils.NewResponse(utils.CodeBadReq, "cart has already been converted to an order", nil)
		}
		if cart.CartStatus == repository.CartStatusAbandoned || cart.CartStatus == repository.CartStatusExpired {
			return utils.NewResponse(utils.CodeBadReq, "cannot checkout an abandoned or expired cart", nil)
		}

		cartItems, err := q.ListRestaurantCartItems(ctx, cart.ID)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to load cart items: "+err.Error(), nil)
		}
		if len(cartItems) == 0 {
			return utils.NewResponse(utils.CodeBadReq, "cannot checkout an empty cart", nil)
		}

		// Recompute totals from current cart items to guarantee header matches lines
		recalCart, err := q.RecalculateRestaurantCartTotals(ctx, cart.ID)
		if err == nil {
			cart = recalCart
		}

		customerID := cart.CustomerID
		if input.CustomerID != nil {
			customerID = utils.Int32ToPgInt4(input.CustomerID)
			cart.CustomerID = customerID
		}

		// Re-validate and apply either the checkout coupon_code or any coupon already saved on the cart
		couponToValidate := strings.TrimSpace(input.CouponCode)
		if couponToValidate == "" && cart.CouponCode.Valid {
			couponToValidate = strings.TrimSpace(cart.CouponCode.String)
		}
		if couponToValidate != "" {
			promoCart, _, errResp := applyRestaurantCouponInTx(ctx, q, cart, couponToValidate)
			if errResp != nil {
				return errResp
			}
			cart = promoCart
			cartItems, err = q.ListRestaurantCartItems(ctx, cart.ID)
			if err != nil {
				return utils.NewResponse(utils.CodeError, "failed to reload cart items after coupon: "+err.Error(), nil)
			}
		}

		orgID := cart.OrganizationID
		if !orgID.Valid {
			orgID = resolveStoreOrganizationID(ctx, q, cart.StoreID, nil)
		}

		tableID := cart.TableID
		if input.TableID != nil {
			tableID = utils.Int32ToPgInt4(input.TableID)
		}
		cashierID := cart.CashierID
		if input.CashierID != nil {
			cashierID = utils.Int32ToPgInt4(input.CashierID)
		}
		posTerminalID := cart.PosTerminalID
		if input.PosTerminalID != nil {
			posTerminalID = utils.Int32ToPgInt4(input.PosTerminalID)
		}
		kioskSessionID := cart.KioskSessionID
		if input.KioskSessionID != nil {
			kioskSessionID = utils.Int32ToPgInt4(input.KioskSessionID)
		}
		createdByUserID := cart.CreatedByUserID
		if input.CreatedByUserID != nil {
			createdByUserID = utils.Int32ToPgInt4(input.CreatedByUserID)
		}

		orderNumber := strings.TrimSpace(input.OrderNumber)
		if orderNumber == "" {
			orderNumber = generateRestaurantRef("RORD")
		}
		orderSource := strings.TrimSpace(input.OrderSource)
		if orderSource == "" {
			if cart.Channel.Valid && strings.TrimSpace(cart.Channel.String) != "" {
				orderSource = cart.Channel.String
			} else {
				orderSource = "counter"
			}
		}
		orderType := strings.TrimSpace(input.OrderType)
		if orderType == "" {
			orderType = cart.OrderType
		}
		if orderType == "" {
			orderType = "dine_in"
		}
		status := strings.TrimSpace(input.Status)
		if status == "" {
			status = "pending"
		}

		totalFloat := numericToFloat(cart.TotalAmount)
		amountPaid := input.AmountPaid
		if amountPaid < 0 {
			return utils.NewResponse(utils.CodeBadReq, "amount_paid cannot be negative", nil)
		}
		changeGiven := 0.0
		paymentStatus := strings.TrimSpace(input.PaymentStatus)
		if paymentStatus == "" {
			switch {
			case amountPaid >= totalFloat && totalFloat > 0:
				paymentStatus = "paid"
				changeGiven = roundMoney(amountPaid - totalFloat)
			case amountPaid > 0:
				paymentStatus = "partially_paid"
			default:
				paymentStatus = "unpaid"
			}
		} else if strings.EqualFold(paymentStatus, "paid") {
			if err := validateFullRestaurantPayment(amountPaid, totalFloat); err != nil {
				return utils.NewResponse(utils.CodeBadReq, err.Error(), nil)
			}
			if amountPaid > totalFloat {
				changeGiven = roundMoney(amountPaid - totalFloat)
			}
		}

		customerName := strings.TrimSpace(input.CustomerName)
		if customerName == "" && cart.GuestName.Valid {
			customerName = cart.GuestName.String
		}
		customerEmail := strings.TrimSpace(input.CustomerEmail)
		if customerEmail == "" && cart.GuestEmail.Valid {
			customerEmail = cart.GuestEmail.String
		}
		customerPhone := strings.TrimSpace(input.CustomerPhone)
		if customerPhone == "" && cart.GuestPhone.Valid {
			customerPhone = cart.GuestPhone.String
		}
		if customerName == "" && customerID.Valid {
			if cust, cErr := q.GetCustomer(ctx, customerID.Int32); cErr == nil {
				customerName = cust.Name
				if customerEmail == "" && cust.Email.Valid {
					customerEmail = cust.Email.String
				}
				if customerPhone == "" && cust.Phone.Valid {
					customerPhone = cust.Phone.String
				}
			}
		}

		deliveryAddr := input.DeliveryAddress
		if len(deliveryAddr) == 0 {
			deliveryAddr = cart.DeliveryAddress
		}
		notes := strings.TrimSpace(input.Notes)
		if notes == "" && cart.Notes.Valid {
			notes = cart.Notes.String
		}
		meta := input.Metadata
		if len(meta) == 0 {
			meta = cart.Metadata
		}

		order, err := q.CreateRestaurantOrderFull(ctx, repository.CreateRestaurantOrderFullParams{
			OrganizationID:   orgID,
			StoreID:          cart.StoreID,
			TableID:          tableID,
			CashierID:        cashierID,
			CashierSessionID: utils.Int32ToPgInt4(input.CashierSessionID),
			CustomerID:       customerID,
			PosTerminalID:    posTerminalID,
			KioskSessionID:   kioskSessionID,
			CreatedByUserID:  createdByUserID,
			SourceCartID:     pgtype.UUID{Bytes: cart.ID, Valid: true},
			OrderNumber:      orderNumber,
			OrderSource:      orderSource,
			OrderType:        orderType,
			Status:           status,
			PaymentStatus:    paymentStatus,
			CustomerName:     nullableText(customerName),
			CustomerEmail:    nullableText(customerEmail),
			CustomerPhone:    nullableText(customerPhone),
			DeliveryAddress:  deliveryAddr,
			Subtotal:         cart.Subtotal,
			DiscountAmount:   cart.DiscountAmount,
			TaxAmount:        cart.TaxAmount,
			TotalAmount:      cart.TotalAmount,
			AmountPaid:       utils.Float64ToPgNumeric(amountPaid),
			ChangeGiven:      utils.Float64ToPgNumeric(changeGiven),
			Notes:            nullableText(notes),
			PosTransactionID: utils.Int32ToPgInt4(input.PosTransactionID),
			Metadata:         ensureJSONObject(meta),
			OrderedAt:        pgtype.Timestamp{Time: time.Now(), Valid: true},
			CouponCode:       cart.CouponCode,
			DiscountCode:     cart.DiscountCode,
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to create restaurant order from cart: "+err.Error(), nil)
		}

		createdItems := make([]repository.RestaurantOrderItem, 0, len(cartItems))
		for idx, ci := range cartItems {
			itemSubtotal := numericToFloat(ci.LineTotal) - numericToFloat(ci.TaxAmount)
			oi, err := q.CreateRestaurantOrderItemFull(ctx, repository.CreateRestaurantOrderItemFullParams{
				OrderID:                 order.ID,
				MenuItemID:              ci.MenuItemID,
				MenuItemName:            pgtype.Text{String: ci.MenuItemName, Valid: ci.MenuItemName != ""},
				Quantity:                ci.Quantity,
				UnitPrice:               ci.UnitPrice,
				ModifiersSnapshot:       ensureJSONArray(ci.ModifiersSnapshot),
				ModifiersTotal:          ci.ModifiersTotal,
				ComboComponentsSnapshot: ensureJSONArray(ci.ComboComponentsSnapshot),
				ComboTotal:              ci.ComboTotal,
				DiscountAmount:          ci.DiscountAmount,
				TaxCategoryID:           ci.TaxCategoryID,
				TaxRate:                 ci.TaxRate,
				TaxAmount:               ci.TaxAmount,
				Subtotal:                utils.Float64ToPgNumeric(roundMoney(itemSubtotal)),
				LineTotal:               ci.LineTotal,
				LineNumber:              pgtype.Int4{Int32: int32(idx + 1), Valid: true},
				Notes:                   ci.Notes,
				Status:                  pgtype.Text{String: "pending", Valid: true},
				Metadata:                ensureJSONObject(ci.Metadata),
			})
			if err != nil {
				return utils.NewResponse(utils.CodeError, "failed to copy cart item into order: "+err.Error(), nil)
			}
			createdItems = append(createdItems, oi)
		}

		// Record initial order status history
		recordRestaurantOrderStatusHistory(ctx, q, order.ID, order.StoreID, nil, order.Status, "order_created", "Order created from cart checkout", createdByUserID, nil)

		// Deduct raw recipe ingredients from inventory_stock and record stock_movements
		if err := deductRecipeIngredientsForOrderItems(ctx, q, order, createdItems, createdByUserID); err != nil {
			return utils.NewResponse(utils.CodeError, "failed to deduct recipe ingredients: "+err.Error(), nil)
		}

		// Atomically claim a remaining promotion usage if a restaurant coupon was applied
		if cart.CouponCode.Valid && strings.TrimSpace(cart.CouponCode.String) != "" {
			promo, pErr := q.GetActiveRestaurantPromotionByCouponCodeInsensitive(ctx, repository.GetActiveRestaurantPromotionByCouponCodeInsensitiveParams{
				CouponCode: cart.CouponCode.String,
				StoreID:    cart.StoreID,
			})
			if pErr != nil {
				return utils.NewResponse(utils.CodeBadReq, "restaurant coupon is no longer valid or active", nil)
			}
			if _, claimErr := q.ClaimRestaurantPromotionUsage(ctx, promo.ID); claimErr != nil {
				return utils.NewResponse(utils.CodeBadReq, "restaurant promotion usage limit reached or expired", nil)
			}
		}

		convertedCart, err := q.ConvertRestaurantCartToOrder(ctx, repository.ConvertRestaurantCartToOrderParams{
			ID:                 cart.ID,
			ConvertedToOrderID: pgtype.Int4{Int32: order.ID, Valid: true},
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to mark restaurant cart as converted: "+err.Error(), nil)
		}

		var posTxnID *int32
		if order.PosTransactionID.Valid && order.PosTransactionID.Int32 > 0 {
			v := order.PosTransactionID.Int32
			posTxnID = &v
		} else if input.PosTransactionID == nil && strings.EqualFold(orderSource, "pos") && !strings.EqualFold(strings.TrimSpace(input.PaymentStatus), "unpaid") {
			pm := strings.TrimSpace(input.PaymentMethod)
			if pm == "" && cart.PaymentMethod.Valid {
				pm = cart.PaymentMethod.String
			}
			pg := strings.TrimSpace(input.PaymentGateway)
			if pg == "" && cart.PaymentGateway.Valid {
				pg = cart.PaymentGateway.String
			}
			var amtOverride *float64
			if input.AmountPaid > 0 {
				v := input.AmountPaid
				amtOverride = &v
			}
			updatedOrder, txn, txnErr := uc.createPosTransactionForRestaurantOrder(
				ctx,
				q,
				order,
				input.CashierID,
				input.CashierSessionID,
				input.PosTerminalID,
				pm,
				pg,
				input.ReferenceNumber,
				amtOverride,
				input.Metadata,
				false,
				order.Status,
				"paid",
			)
			if txnErr != nil {
				if errors.Is(txnErr, errInvalidRestaurantPayment) {
					return utils.NewResponse(utils.CodeBadReq, txnErr.Error(), nil)
				}
				return utils.NewResponse(utils.CodeError, "failed to create POS transaction for restaurant order: "+txnErr.Error(), nil)
			}
			if updatedOrder != nil && txn != nil {
				order = *updatedOrder
				v := txn.ID
				posTxnID = &v
			}
		}

		respData := map[string]interface{}{
			"cart":  restaurantCartToOutput(convertedCart),
			"order": order,
			"items": createdItems,
		}
		if posTxnID != nil {
			respData["pos_transaction_id"] = *posTxnID
			respData["transaction_id"] = *posTxnID
		}

		return utils.NewResponse(utils.CodeCreated, "restaurant cart converted to order successfully", respData)
	})
}

// ReopenOrderToCart reverts an unpaid/unfulfilled restaurant order back into an active
// restaurant cart so the cashier/customer can continue editing items.
func (uc *RestaurantCartUseCase) ReopenOrderToCart(ctx context.Context, orderID int32) *repository.Response {
	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		return reopenOrderToCartInTx(ctx, q, orderID)
	})
}

func reopenOrderToCartInTx(ctx context.Context, q *repository.Queries, orderID int32) *repository.Response {
	order, err := lockAndGetRestaurantOrderByID(ctx, q, orderID)
	if err != nil {
		return restaurantOrderNotFoundOrErr(err, "lock restaurant order")
	}
	if order.PaymentStatus == "paid" || order.Status == "paid" || order.Status == "completed" {
		return utils.NewResponse(utils.CodeBadReq, "cannot reopen a paid or completed restaurant order", nil)
	}
	if order.Status == "voided" || order.Status == "cancelled" {
		return utils.NewResponse(utils.CodeBadReq, "cannot reopen a voided or cancelled restaurant order", nil)
	}

	orgID := order.OrganizationID
	if !orgID.Valid {
		orgID = resolveStoreOrganizationID(ctx, q, order.StoreID, nil)
	}

	var reopenedCart repository.RestaurantCart
	if order.SourceCartID.Valid {
		cartUUID := uuid.UUID(order.SourceCartID.Bytes)
		if _, lockErr := q.LockRestaurantCart(ctx, cartUUID); lockErr == nil {
			if c, err := q.ReopenRestaurantCartByID(ctx, cartUUID); err == nil {
				if updatedCart, updErr := q.UpdateRestaurantCart(ctx, repository.UpdateRestaurantCartParams{
					ID:              c.ID,
					TableID:         order.TableID,
					CustomerID:      order.CustomerID,
					GuestName:       order.CustomerName,
					GuestEmail:      order.CustomerEmail,
					GuestPhone:      order.CustomerPhone,
					DeliveryAddress: order.DeliveryAddress,
					Notes:           order.Notes,
					OrderType:       pgtype.Text{String: order.OrderType, Valid: strings.TrimSpace(order.OrderType) != ""},
				}); updErr == nil {
					reopenedCart = updatedCart
				} else {
					reopenedCart = c
				}
				if err := q.ClearRestaurantCartItems(ctx, reopenedCart.ID); err != nil {
					return utils.NewResponse(utils.CodeError, "failed to clear stale reopened cart items: "+err.Error(), nil)
				}
			}
		}
	}

	if reopenedCart.ID == uuid.Nil {
		c, err := q.CreateRestaurantCart(ctx, repository.CreateRestaurantCartParams{
			CartNumber:      pgtype.Text{String: generateRestaurantRef("RCART"), Valid: true},
			OrganizationID:  orgID,
			StoreID:         order.StoreID,
			TableID:         order.TableID,
			OrderType:       order.OrderType,
			CustomerID:      order.CustomerID,
			GuestName:       order.CustomerName,
			GuestEmail:      order.CustomerEmail,
			GuestPhone:      order.CustomerPhone,
			CartStatus:      repository.CartStatusActive,
			Channel:         pgtype.Text{String: order.OrderSource, Valid: true},
			DeviceInfo:      json.RawMessage("{}"),
			CreatedByUserID: order.CreatedByUserID,
			CashierID:       order.CashierID,
			PosTerminalID:   order.PosTerminalID,
			KioskSessionID:  order.KioskSessionID,
			DeliveryAddress: order.DeliveryAddress,
			Notes:           order.Notes,
			Metadata:        ensureJSONObject(order.Metadata),
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to create reopened cart: "+err.Error(), nil)
		}
		reopenedCart = c
	}

	orderItems, err := q.ListRestaurantOrderItems(ctx, order.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to load order items for reopened cart: "+err.Error(), nil)
	}
	for _, oi := range orderItems {
		lineTotal := oi.LineTotal
		if !lineTotal.Valid || numericToFloat(lineTotal) == 0 {
			lineTotal = oi.Subtotal
		}
		if _, err := q.CreateRestaurantCartItem(ctx, repository.CreateRestaurantCartItemParams{
			CartID:                  reopenedCart.ID,
			OrganizationID:          orgID,
			MenuItemID:              oi.MenuItemID,
			Quantity:                oi.Quantity,
			UnitPrice:               oi.UnitPrice,
			ModifiersSnapshot:       ensureJSONArray(oi.ModifiersSnapshot),
			ModifiersTotal:          oi.ModifiersTotal,
			ComboComponentsSnapshot: ensureJSONArray(oi.ComboComponentsSnapshot),
			ComboTotal:              oi.ComboTotal,
			DiscountAmount:          oi.DiscountAmount,
			TaxCategoryID:           oi.TaxCategoryID,
			TaxRate:                 oi.TaxRate,
			TaxAmount:               oi.TaxAmount,
			LineTotal:               lineTotal,
			CustomizationDetails:    json.RawMessage("{}"),
			Notes:                   oi.Notes,
			Metadata:                ensureJSONObject(oi.Metadata),
		}); err != nil {
			return utils.NewResponse(utils.CodeError, "failed to copy order item to reopened cart: "+err.Error(), nil)
		}
	}
	reopenedCart, err = q.RecalculateRestaurantCartTotals(ctx, reopenedCart.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to recalculate reopened cart totals: "+err.Error(), nil)
	}

	if err := reverseRecipeIngredientsForOrder(ctx, q, order, order.CreatedByUserID); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to reverse recipe stock deduction: "+err.Error(), nil)
	}

	prevStatus := order.Status
	if _, err := q.UpdateRestaurantOrderStatusWithTimestamps(ctx, repository.UpdateRestaurantOrderStatusWithTimestampsParams{
		ID:     order.ID,
		Status: "voided",
	}); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to void reopened order: "+err.Error(), nil)
	}
	recordRestaurantOrderStatusHistory(ctx, q, order.ID, order.StoreID, &prevStatus, "voided", "reopened_to_cart", "Order reopened back into cart", order.CreatedByUserID, nil)

	payload, err := buildCartDetailPayload(ctx, q, reopenedCart)
	if err != nil {
		return utils.NewResponse(utils.CodeOK, "order reopened to cart", restaurantCartToOutput(reopenedCart))
	}
	return utils.NewResponse(utils.CodeOK, "order reopened to cart", payload)
}

func (uc *RestaurantCartUseCase) ReopenCartByCartID(ctx context.Context, cartID uuid.UUID) *repository.Response {
	var delegateOrderID int32
	resp := uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		peekCart, err := q.GetRestaurantCartByID(ctx, cartID)
		if err != nil {
			return utils.NewResponse(utils.CodeNotFound, "restaurant cart not found", nil)
		}
		if peekCart.ConvertedToOrderID.Valid {
			return reopenOrderToCartInTx(ctx, q, peekCart.ConvertedToOrderID.Int32)
		}
		cart, err := lockAndGetRestaurantCartByID(ctx, q, cartID)
		if err != nil {
			return restaurantCartNotFoundOrErr(err, "lock restaurant cart")
		}
		if cart.ConvertedToOrderID.Valid {
			delegateOrderID = cart.ConvertedToOrderID.Int32
			return nil
		}
		reopened, err := q.ReopenRestaurantCartByID(ctx, cartID)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to reopen restaurant cart: "+err.Error(), nil)
		}
		payload, err := buildCartDetailPayload(ctx, q, reopened)
		if err != nil {
			return utils.NewResponse(utils.CodeOK, "restaurant cart reopened", restaurantCartToOutput(reopened))
		}
		return utils.NewResponse(utils.CodeOK, "restaurant cart reopened", payload)
	})
	if delegateOrderID > 0 {
		return uc.ReopenOrderToCart(ctx, delegateOrderID)
	}
	return resp
}

// === Direct Order Creation / Update with Items & Server Calculation ===

func (uc *RestaurantCartUseCase) CreateOrderWithItems(ctx context.Context, input RestaurantDirectOrderInput) *repository.Response {
	if input.StoreID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "store_id is required", nil)
	}

	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		orgID := resolveStoreOrganizationID(ctx, q, input.StoreID, input.OrganizationID)

		orderNumber := strings.TrimSpace(input.OrderNumber)
		if orderNumber == "" {
			orderNumber = generateRestaurantRef("RORD")
		}
		orderSource := strings.TrimSpace(input.OrderSource)
		if orderSource == "" {
			orderSource = "counter"
		}
		orderType := strings.TrimSpace(input.OrderType)
		if orderType == "" {
			if orderSource == "online" || orderSource == "delivery" {
				orderType = "delivery"
			} else if input.TableID != nil && *input.TableID > 0 {
				orderType = "dine_in"
			} else {
				orderType = "takeaway"
			}
		}
		status := strings.TrimSpace(input.Status)
		if status == "" {
			status = "pending"
		}
		paymentStatus := strings.TrimSpace(input.PaymentStatus)
		if paymentStatus == "" {
			paymentStatus = "unpaid"
		}

		// Resolve all line items first so invalid menu_item_id/modifiers fail before inserting order
		resolvedLines := make([]*resolvedRestaurantLine, 0, len(input.Items))
		var calcSubtotal, calcDiscount, calcTax, calcTotal float64
		for _, itemInput := range input.Items {
			resolved, errResp := resolveRestaurantLineItem(ctx, q, input.StoreID, itemInput)
			if errResp != nil {
				return errResp
			}
			resolvedLines = append(resolvedLines, resolved)
			calcSubtotal += resolved.Subtotal
			calcDiscount += resolved.DiscountAmount
			calcTax += resolved.TaxAmount
			calcTotal += resolved.LineTotal
		}

		if len(resolvedLines) > 0 {
			calcSubtotal = roundMoney(calcSubtotal)
			calcDiscount = roundMoney(calcDiscount)
			calcTax = roundMoney(calcTax)
			calcTotal = roundMoney(calcTotal)
		} else {
			calcSubtotal = roundMoney(input.Subtotal)
			calcDiscount = roundMoney(input.DiscountAmount)
			calcTax = roundMoney(input.TaxAmount)
			calcTotal = roundMoney(input.TotalAmount)
		}

		var sourceCartUUID pgtype.UUID
		if input.SourceCartID != nil && *input.SourceCartID != uuid.Nil {
			sourceCartUUID = pgtype.UUID{Bytes: *input.SourceCartID, Valid: true}
		}

		order, err := q.CreateRestaurantOrderFull(ctx, repository.CreateRestaurantOrderFullParams{
			OrganizationID:   orgID,
			StoreID:          input.StoreID,
			TableID:          utils.Int32ToPgInt4(input.TableID),
			CashierID:        utils.Int32ToPgInt4(input.CashierID),
			CashierSessionID: utils.Int32ToPgInt4(input.CashierSessionID),
			CustomerID:       utils.Int32ToPgInt4(input.CustomerID),
			PosTerminalID:    utils.Int32ToPgInt4(input.PosTerminalID),
			KioskSessionID:   utils.Int32ToPgInt4(input.KioskSessionID),
			CreatedByUserID:  utils.Int32ToPgInt4(input.CreatedByUserID),
			SourceCartID:     sourceCartUUID,
			OrderNumber:      orderNumber,
			OrderSource:      orderSource,
			OrderType:        orderType,
			Status:           status,
			PaymentStatus:    paymentStatus,
			CustomerName:     nullableText(input.CustomerName),
			CustomerEmail:    nullableText(input.CustomerEmail),
			CustomerPhone:    nullableText(input.CustomerPhone),
			DeliveryAddress:  input.DeliveryAddress,
			Subtotal:         utils.Float64ToPgNumeric(calcSubtotal),
			DiscountAmount:   utils.Float64ToPgNumeric(calcDiscount),
			TaxAmount:        utils.Float64ToPgNumeric(calcTax),
			TotalAmount:      utils.Float64ToPgNumeric(calcTotal),
			AmountPaid:       utils.Float64ToPgNumeric(input.AmountPaid),
			ChangeGiven:      utils.Float64ToPgNumeric(input.ChangeGiven),
			Notes:            nullableText(input.Notes),
			PosTransactionID: utils.Int32ToPgInt4(input.PosTransactionID),
			Metadata:         ensureJSONObject(input.Metadata),
			OrderedAt:        pgtype.Timestamp{Time: time.Now(), Valid: true},
			CouponCode:       nullableText(input.CouponCode),
			DiscountCode:     nullableText(input.DiscountCode),
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to create restaurant order: "+err.Error(), nil)
		}

		createdItems := make([]repository.RestaurantOrderItem, 0, len(resolvedLines))
		for idx, r := range resolvedLines {
			oi, err := q.CreateRestaurantOrderItemFull(ctx, repository.CreateRestaurantOrderItemFullParams{
				OrderID:                 order.ID,
				MenuItemID:              r.MenuItem.ID,
				MenuItemName:            pgtype.Text{String: r.MenuItem.Name, Valid: true},
				Quantity:                utils.Float64ToPgNumeric(r.Quantity),
				UnitPrice:               utils.Float64ToPgNumeric(r.UnitPrice),
				ModifiersSnapshot:       r.ModifiersSnapshot,
				ModifiersTotal:          utils.Float64ToPgNumeric(r.ModifiersTotal),
				ComboComponentsSnapshot: r.ComboSnapshot,
				ComboTotal:              utils.Float64ToPgNumeric(r.ComboTotal),
				DiscountAmount:          utils.Float64ToPgNumeric(r.DiscountAmount),
				TaxCategoryID:           r.TaxCategoryID,
				TaxRate:                 utils.Float64ToPgNumeric(r.TaxRate),
				TaxAmount:               utils.Float64ToPgNumeric(r.TaxAmount),
				Subtotal:                utils.Float64ToPgNumeric(r.Subtotal),
				LineTotal:               utils.Float64ToPgNumeric(r.LineTotal),
				LineNumber:              pgtype.Int4{Int32: int32(idx + 1), Valid: true},
				Notes:                   nullableText(r.Notes),
				Status:                  pgtype.Text{String: r.Status, Valid: true},
				Metadata:                r.MetadataJSON,
			})
			if err != nil {
				return utils.NewResponse(utils.CodeError, "failed to create restaurant order item: "+err.Error(), nil)
			}
			createdItems = append(createdItems, oi)
		}

		recordRestaurantOrderStatusHistory(ctx, q, order.ID, order.StoreID, nil, order.Status, "order_created", "Direct restaurant order created", utils.Int32ToPgInt4(input.CreatedByUserID), nil)
		if err := deductRecipeIngredientsForOrderItems(ctx, q, order, createdItems, utils.Int32ToPgInt4(input.CreatedByUserID)); err != nil {
			return utils.NewResponse(utils.CodeError, "failed to deduct recipe ingredients: "+err.Error(), nil)
		}

		return utils.NewResponse(utils.CodeCreated, "restaurant order created successfully", RestaurantOrderWithItems{
			RestaurantOrder: order,
			Items:           createdItems,
		})
	})
}

func (uc *RestaurantCartUseCase) UpdateOrderWithItems(ctx context.Context, orderID int32, input RestaurantDirectOrderInput, replaceItems bool) *repository.Response {
	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		existing, err := lockAndGetRestaurantOrderByID(ctx, q, orderID)
		if err != nil {
			return restaurantOrderNotFoundOrErr(err, "lock restaurant order")
		}

		if replaceItems && isRestaurantOrderPaidOrClosed(existing) {
			return utils.NewResponse(utils.CodeBadReq, "cannot edit items on a paid, voided, or cancelled order", nil)
		}

		status := strings.TrimSpace(input.Status)
		if status == "" {
			status = existing.Status
		}
		if isInactiveOrderStatus(existing.Status) && !strings.EqualFold(status, existing.Status) {
			return utils.NewResponse(utils.CodeBadReq, "cannot transition order out of voided or cancelled status", nil)
		}

		resolvedLines := make([]*resolvedRestaurantLine, 0, len(input.Items))
		var calcSubtotal, calcDiscount, calcTax, calcTotal float64
		if replaceItems {
			for _, itemInput := range input.Items {
				resolved, errResp := resolveRestaurantLineItem(ctx, q, existing.StoreID, itemInput)
				if errResp != nil {
					return errResp
				}
				resolvedLines = append(resolvedLines, resolved)
				calcSubtotal += resolved.Subtotal
				calcDiscount += resolved.DiscountAmount
				calcTax += resolved.TaxAmount
				calcTotal += resolved.LineTotal
			}
			if err := q.DeleteRestaurantOrderItemsByOrderID(ctx, orderID); err != nil {
				return utils.NewResponse(utils.CodeError, "failed to replace restaurant order items: "+err.Error(), nil)
			}
		}
		orderType := strings.TrimSpace(input.OrderType)
		if orderType == "" {
			orderType = existing.OrderType
		}
		paymentStatus := strings.TrimSpace(input.PaymentStatus)
		if paymentStatus == "" {
			paymentStatus = existing.PaymentStatus
		}

		subtotalNum := existing.Subtotal
		discountNum := existing.DiscountAmount
		taxNum := existing.TaxAmount
		totalNum := existing.TotalAmount
		if replaceItems {
			subtotalNum = utils.Float64ToPgNumeric(roundMoney(calcSubtotal))
			discountNum = utils.Float64ToPgNumeric(roundMoney(calcDiscount))
			taxNum = utils.Float64ToPgNumeric(roundMoney(calcTax))
			totalNum = utils.Float64ToPgNumeric(roundMoney(calcTotal))
		} else {
			if input.Subtotal > 0 {
				subtotalNum = utils.Float64ToPgNumeric(input.Subtotal)
			}
			if input.DiscountAmount > 0 {
				discountNum = utils.Float64ToPgNumeric(input.DiscountAmount)
			}
			if input.TaxAmount > 0 {
				taxNum = utils.Float64ToPgNumeric(input.TaxAmount)
			}
			if input.TotalAmount > 0 {
				totalNum = utils.Float64ToPgNumeric(input.TotalAmount)
			}
		}

		amountPaidNum := existing.AmountPaid
		if input.AmountPaid > 0 {
			amountPaidNum = utils.Float64ToPgNumeric(input.AmountPaid)
		}
		changeNum := existing.ChangeGiven
		if input.ChangeGiven > 0 {
			changeNum = utils.Float64ToPgNumeric(input.ChangeGiven)
		}

		tableID := existing.TableID
		if input.TableID != nil {
			tableID = utils.Int32ToPgInt4(input.TableID)
		}
		cashierID := existing.CashierID
		if input.CashierID != nil {
			cashierID = utils.Int32ToPgInt4(input.CashierID)
		}
		cashierSessionID := existing.CashierSessionID
		if input.CashierSessionID != nil {
			cashierSessionID = utils.Int32ToPgInt4(input.CashierSessionID)
		}
		customerID := existing.CustomerID
		if input.CustomerID != nil {
			customerID = utils.Int32ToPgInt4(input.CustomerID)
		}
		posTxnID := existing.PosTransactionID
		if input.PosTransactionID != nil {
			posTxnID = utils.Int32ToPgInt4(input.PosTransactionID)
		}

		notes := existing.Notes
		if input.Notes != "" {
			notes = pgtype.Text{String: input.Notes, Valid: true}
		}
		custName := existing.CustomerName
		if input.CustomerName != "" {
			custName = pgtype.Text{String: input.CustomerName, Valid: true}
		}
		custEmail := existing.CustomerEmail
		if input.CustomerEmail != "" {
			custEmail = pgtype.Text{String: input.CustomerEmail, Valid: true}
		}
		custPhone := existing.CustomerPhone
		if input.CustomerPhone != "" {
			custPhone = pgtype.Text{String: input.CustomerPhone, Valid: true}
		}
		deliveryAddr := existing.DeliveryAddress
		if len(input.DeliveryAddress) > 0 {
			deliveryAddr = input.DeliveryAddress
		}
		meta := existing.Metadata
		if len(input.Metadata) > 0 {
			meta = input.Metadata
		}

		updatedOrder, err := q.UpdateRestaurantOrderFull(ctx, repository.UpdateRestaurantOrderFullParams{
			ID:               orderID,
			TableID:          tableID,
			CashierID:        cashierID,
			CashierSessionID: cashierSessionID,
			CustomerID:       customerID,
			OrderType:        orderType,
			Status:           status,
			PaymentStatus:    paymentStatus,
			CustomerName:     custName,
			CustomerEmail:    custEmail,
			CustomerPhone:    custPhone,
			DeliveryAddress:  deliveryAddr,
			Subtotal:         subtotalNum,
			DiscountAmount:   discountNum,
			TaxAmount:        taxNum,
			TotalAmount:      totalNum,
			AmountPaid:       amountPaidNum,
			ChangeGiven:      changeNum,
			Notes:            notes,
			PosTransactionID: posTxnID,
			ConfirmedAt:      existing.ConfirmedAt,
			ServedAt:         existing.ServedAt,
			PaidAt:           existing.PaidAt,
			Metadata:         ensureJSONObject(meta),
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to update restaurant order: "+err.Error(), nil)
		}

		if existing.Status != updatedOrder.Status {
			from := existing.Status
			recordRestaurantOrderStatusHistory(ctx, q, orderID, updatedOrder.StoreID, &from, updatedOrder.Status, "order_updated", "Order updated", updatedOrder.CreatedByUserID, nil)
		}

		var items []repository.RestaurantOrderItem
		if !isInactiveOrderStatus(existing.Status) && (replaceItems || isInactiveOrderStatus(updatedOrder.Status)) {
			if err := reverseRecipeIngredientsForOrder(ctx, q, existing, updatedOrder.CreatedByUserID); err != nil {
				return utils.NewResponse(utils.CodeError, "failed to reverse recipe stock deduction: "+err.Error(), nil)
			}
		}
		if replaceItems {
			items = make([]repository.RestaurantOrderItem, 0, len(resolvedLines))
			for idx, r := range resolvedLines {
				oi, err := q.CreateRestaurantOrderItemFull(ctx, repository.CreateRestaurantOrderItemFullParams{
					OrderID:                 orderID,
					MenuItemID:              r.MenuItem.ID,
					MenuItemName:            pgtype.Text{String: r.MenuItem.Name, Valid: true},
					Quantity:                utils.Float64ToPgNumeric(r.Quantity),
					UnitPrice:               utils.Float64ToPgNumeric(r.UnitPrice),
					ModifiersSnapshot:       r.ModifiersSnapshot,
					ModifiersTotal:          utils.Float64ToPgNumeric(r.ModifiersTotal),
					ComboComponentsSnapshot: r.ComboSnapshot,
					ComboTotal:              utils.Float64ToPgNumeric(r.ComboTotal),
					DiscountAmount:          utils.Float64ToPgNumeric(r.DiscountAmount),
					TaxCategoryID:           r.TaxCategoryID,
					TaxRate:                 utils.Float64ToPgNumeric(r.TaxRate),
					TaxAmount:               utils.Float64ToPgNumeric(r.TaxAmount),
					Subtotal:                utils.Float64ToPgNumeric(r.Subtotal),
					LineTotal:               utils.Float64ToPgNumeric(r.LineTotal),
					LineNumber:              pgtype.Int4{Int32: int32(idx + 1), Valid: true},
					Notes:                   nullableText(r.Notes),
					Status:                  pgtype.Text{String: r.Status, Valid: true},
					Metadata:                r.MetadataJSON,
				})
				if err != nil {
					return utils.NewResponse(utils.CodeError, "failed to insert order item: "+err.Error(), nil)
				}
				items = append(items, oi)
			}
			if !isInactiveOrderStatus(updatedOrder.Status) {
				if err := deductRecipeIngredientsForOrderItems(ctx, q, updatedOrder, items, updatedOrder.CreatedByUserID); err != nil {
					return utils.NewResponse(utils.CodeError, "failed to deduct recipe ingredients: "+err.Error(), nil)
				}
			}
		} else {
			items, _ = q.ListRestaurantOrderItems(ctx, orderID)
		}

		return utils.NewResponse(utils.CodeOK, "restaurant order updated successfully", RestaurantOrderWithItems{
			RestaurantOrder: updatedOrder,
			Items:           items,
		})
	})
}

func (uc *RestaurantCartUseCase) AddOrderItem(ctx context.Context, orderID int32, input RestaurantLineItemInput) *repository.Response {
	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		order, err := lockAndGetRestaurantOrderByID(ctx, q, orderID)
		if err != nil {
			return restaurantOrderNotFoundOrErr(err, "lock restaurant order")
		}
		if isInactiveOrderStatus(order.Status) {
			return utils.NewResponse(utils.CodeBadReq, "cannot settle a voided or cancelled order", nil)
		}
		if isRestaurantOrderPaidOrClosed(order) {
			return utils.NewResponse(utils.CodeBadReq, "cannot add items to a paid, voided, or cancelled order", nil)
		}

		resolved, errResp := resolveRestaurantLineItem(ctx, q, order.StoreID, input)
		if errResp != nil {
			return errResp
		}

		existingItems, _ := q.ListRestaurantOrderItems(ctx, orderID)
		lineNo := int32(len(existingItems) + 1)

		oi, err := q.CreateRestaurantOrderItemFull(ctx, repository.CreateRestaurantOrderItemFullParams{
			OrderID:                 orderID,
			MenuItemID:              resolved.MenuItem.ID,
			MenuItemName:            pgtype.Text{String: resolved.MenuItem.Name, Valid: true},
			Quantity:                utils.Float64ToPgNumeric(resolved.Quantity),
			UnitPrice:               utils.Float64ToPgNumeric(resolved.UnitPrice),
			ModifiersSnapshot:       resolved.ModifiersSnapshot,
			ModifiersTotal:          utils.Float64ToPgNumeric(resolved.ModifiersTotal),
			ComboComponentsSnapshot: resolved.ComboSnapshot,
			ComboTotal:              utils.Float64ToPgNumeric(resolved.ComboTotal),
			DiscountAmount:          utils.Float64ToPgNumeric(resolved.DiscountAmount),
			TaxCategoryID:           resolved.TaxCategoryID,
			TaxRate:                 utils.Float64ToPgNumeric(resolved.TaxRate),
			TaxAmount:               utils.Float64ToPgNumeric(resolved.TaxAmount),
			Subtotal:                utils.Float64ToPgNumeric(resolved.Subtotal),
			LineTotal:               utils.Float64ToPgNumeric(resolved.LineTotal),
			LineNumber:              pgtype.Int4{Int32: lineNo, Valid: true},
			Notes:                   nullableText(resolved.Notes),
			Status:                  pgtype.Text{String: resolved.Status, Valid: true},
			Metadata:                resolved.MetadataJSON,
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to create order item: "+err.Error(), nil)
		}

		updatedOrder, err := q.RecalculateRestaurantOrderTotals(ctx, orderID)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to recalculate order totals: "+err.Error(), nil)
		}

		if err := deductRecipeIngredientsForOrderItems(ctx, q, updatedOrder, []repository.RestaurantOrderItem{oi}, updatedOrder.CreatedByUserID); err != nil {
			return utils.NewResponse(utils.CodeError, "failed to deduct recipe ingredients for added item: "+err.Error(), nil)
		}

		items, _ := q.ListRestaurantOrderItems(ctx, orderID)
		return utils.NewResponse(utils.CodeCreated, "item added to restaurant order", map[string]interface{}{
			"item":  oi,
			"order": updatedOrder,
			"items": items,
		})
	})
}

var errInvalidRestaurantPayment = errors.New("invalid restaurant payment")

func moneyCents(v float64) int64 {
	return int64(math.Round(roundMoney(v) * 100))
}

func validateFullRestaurantPayment(paidFloat, totalFloat float64) error {
	if paidFloat < 0 {
		return fmt.Errorf("%w: amount_paid cannot be negative", errInvalidRestaurantPayment)
	}
	if moneyCents(paidFloat) < moneyCents(totalFloat) {
		return fmt.Errorf("%w: amount_paid (%.2f) is less than order total (%.2f)", errInvalidRestaurantPayment, roundMoney(paidFloat), roundMoney(totalFloat))
	}
	return nil
}

func (uc *RestaurantCartUseCase) UpdateOrderPaymentStatus(ctx context.Context, orderID int32, paymentStatus string, amountPaid, changeGiven *float64) *repository.Response {
	if strings.TrimSpace(paymentStatus) == "" {
		return utils.NewResponse(utils.CodeBadReq, "payment_status is required", nil)
	}
	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		order, err := lockAndGetRestaurantOrderByID(ctx, q, orderID)
		if err != nil {
			return restaurantOrderNotFoundOrErr(err, "lock restaurant order")
		}
		if isInactiveOrderStatus(order.Status) {
			return utils.NewResponse(utils.CodeBadReq, "cannot settle a voided or cancelled order", nil)
		}

		effectivePaid := numericToFloat(order.AmountPaid)
		if amountPaid != nil {
			if *amountPaid < 0 {
				return utils.NewResponse(utils.CodeBadReq, "amount_paid cannot be negative", nil)
			}
			effectivePaid = *amountPaid
		}
		if strings.EqualFold(strings.TrimSpace(paymentStatus), "paid") {
			if err := validateFullRestaurantPayment(effectivePaid, numericToFloat(order.TotalAmount)); err != nil {
				return utils.NewResponse(utils.CodeBadReq, err.Error(), nil)
			}
		}

		paidVal := order.AmountPaid
		if amountPaid != nil {
			paidVal = utils.Float64ToPgNumeric(*amountPaid)
		}
		changeVal := order.ChangeGiven
		if changeGiven != nil {
			changeVal = utils.Float64ToPgNumeric(*changeGiven)
		} else if amountPaid != nil {
			total := numericToFloat(order.TotalAmount)
			if *amountPaid > total {
				changeVal = utils.Float64ToPgNumeric(roundMoney(*amountPaid - total))
			}
		}

		updated, err := q.UpdateRestaurantOrderPaymentStatus(ctx, repository.UpdateRestaurantOrderPaymentStatusParams{
			ID:            orderID,
			PaymentStatus: paymentStatus,
			AmountPaid:    paidVal,
			ChangeGiven:   changeVal,
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to update payment status: "+err.Error(), nil)
		}
		return utils.NewResponse(utils.CodeOK, "order payment status updated", updated)
	})
}

func (uc *RestaurantCartUseCase) createPosTransactionForRestaurantOrder(
	ctx context.Context,
	q *repository.Queries,
	order repository.RestaurantOrder,
	explicitCashierID *int32,
	explicitSessionID *int32,
	explicitTerminalID *int32,
	paymentMethod string,
	paymentGateway string,
	referenceNumber string,
	amountPaidOverride *float64,
	extraMetadata json.RawMessage,
	requireSession bool,
	targetOrderStatus string,
	targetPaymentStatus string,
) (*repository.RestaurantOrder, *repository.PosTransaction, error) {
	cashierID := order.CashierID
	if explicitCashierID != nil {
		cashierID = utils.Int32ToPgInt4(explicitCashierID)
	}
	sessionID := order.CashierSessionID
	if explicitSessionID != nil {
		sessionID = utils.Int32ToPgInt4(explicitSessionID)
	}
	terminalID := order.PosTerminalID
	if explicitTerminalID != nil {
		terminalID = utils.Int32ToPgInt4(explicitTerminalID)
	}

	if amountPaidOverride != nil && *amountPaidOverride < 0 {
		return nil, nil, fmt.Errorf("%w: amount_paid cannot be negative", errInvalidRestaurantPayment)
	}

	totalFloat := roundMoney(numericToFloat(order.TotalAmount))
	paidFloat := roundMoney(numericToFloat(order.AmountPaid))
	if amountPaidOverride != nil {
		paidFloat = roundMoney(*amountPaidOverride)
	} else if paidFloat <= 0 && totalFloat > 0 {
		paidFloat = totalFloat
	}

	if err := validateFullRestaurantPayment(paidFloat, totalFloat); err != nil {
		return nil, nil, err
	}

	if sessionID.Valid {
		sess, err := q.ValidateRestaurantCashierSession(ctx, sessionID.Int32)
		if err != nil {
			return nil, nil, fmt.Errorf("cashier session %d not found", sessionID.Int32)
		}
		if !sess.Status.Valid || !strings.EqualFold(sess.Status.String, "open") || sess.ClosingTime.Valid {
			return nil, nil, fmt.Errorf("cashier session %d is closed", sessionID.Int32)
		}
		if sess.StoreID != order.StoreID {
			return nil, nil, fmt.Errorf("cashier session %d does not belong to store %d", sessionID.Int32, order.StoreID)
		}
		if cashierID.Valid && sess.CashierID != cashierID.Int32 {
			return nil, nil, fmt.Errorf("cashier session %d does not belong to cashier %d", sessionID.Int32, cashierID.Int32)
		}
		if terminalID.Valid && sess.PosTerminalID != terminalID.Int32 {
			return nil, nil, fmt.Errorf("cashier session %d does not belong to terminal %d", sessionID.Int32, terminalID.Int32)
		}
		if !cashierID.Valid {
			cashierID = pgtype.Int4{Int32: sess.CashierID, Valid: true}
		}
		if !terminalID.Valid && sess.PosTerminalID > 0 {
			terminalID = pgtype.Int4{Int32: sess.PosTerminalID, Valid: true}
		}
	} else {
		openSess, err := q.FindOpenCashierSessionForRestaurantStore(ctx, repository.FindOpenCashierSessionForRestaurantStoreParams{
			StoreID:       order.StoreID,
			CashierID:     cashierID,
			PosTerminalID: terminalID,
		})
		if err != nil {
			if requireSession {
				return nil, nil, fmt.Errorf("no open cashier session found for store %d; pass pos_transaction_id or open a cashier session", order.StoreID)
			}
			return nil, nil, nil
		}
		sessionID = pgtype.Int4{Int32: openSess.ID, Valid: true}
		if !cashierID.Valid {
			cashierID = pgtype.Int4{Int32: openSess.CashierID, Valid: true}
		}
		if !terminalID.Valid && openSess.PosTerminalID > 0 {
			terminalID = pgtype.Int4{Int32: openSess.PosTerminalID, Valid: true}
		}
	}

	pm := strings.ToLower(strings.TrimSpace(paymentMethod))
	if pm == "" {
		pm = "cash"
	}

	changeFloat := 0.0
	if paidFloat > totalFloat {
		changeFloat = roundMoney(paidFloat - totalFloat)
	}
	netPaid := paidFloat
	if changeFloat > 0 {
		netPaid = roundMoney(paidFloat - changeFloat)
	}

	metaMap := unmarshalJSONObject(order.Metadata)
	if len(extraMetadata) > 0 {
		for k, v := range unmarshalJSONObject(extraMetadata) {
			metaMap[k] = v
		}
	}
	metaMap["restaurant_order_id"] = order.ID
	metaMap["order_number"] = order.OrderNumber
	metaMap["order_type"] = order.OrderType
	metaMap["order_source"] = order.OrderSource
	if order.SourceCartID.Valid {
		metaMap["source_cart_id"] = uuid.UUID(order.SourceCartID.Bytes).String()
	}
	txnMeta, err := json.Marshal(metaMap)
	if err != nil {
		txnMeta = json.RawMessage("{}")
	}

	txn, err := q.CreatePosTransactionForRestaurantOrder(ctx, repository.CreatePosTransactionForRestaurantOrderParams{
		StoreID:           order.StoreID,
		CashierID:         cashierID.Int32,
		CashierSessionID:  sessionID.Int32,
		CustomerID:        order.CustomerID,
		PosTerminalID:     terminalID,
		TransactionNumber: "TXN-" + order.OrderNumber,
		TransactionType:   pgtype.Text{String: "restaurant_sale", Valid: true},
		Subtotal:          order.Subtotal,
		DiscountAmount:    order.DiscountAmount,
		TaxAmount:         order.TaxAmount,
		TotalAmount:       order.TotalAmount,
		AmountPaid:        utils.Float64ToPgNumeric(paidFloat),
		ChangeGiven:       utils.Float64ToPgNumeric(changeFloat),
		Status:            pgtype.Text{String: "completed", Valid: true},
		Metadata:          txnMeta,
	})
	if err != nil {
		return nil, nil, err
	}

	if err := q.AddPaymentToTransaction(ctx, repository.AddPaymentToTransactionParams{
		TransactionID:   txn.ID,
		PaymentMethod:   pm,
		PaymentGateway:  nullableText(paymentGateway),
		Amount:          utils.Float64ToPgNumeric(netPaid),
		ReferenceNumber: nullableText(referenceNumber),
		Metadata:        json.RawMessage("{}"),
	}); err != nil {
		return nil, nil, err
	}

	if pm == "cash" {
		if err := q.UpdateSessionExpectedBalance(ctx, repository.UpdateSessionExpectedBalanceParams{
			ID:              sessionID.Int32,
			ExpectedBalance: utils.Float64ToPgNumeric(netPaid),
		}); err != nil {
			return nil, nil, err
		}
	}

	if strings.TrimSpace(targetOrderStatus) == "" {
		targetOrderStatus = order.Status
	}
	if strings.TrimSpace(targetPaymentStatus) == "" {
		targetPaymentStatus = order.PaymentStatus
	}

	updatedOrder, err := q.LinkRestaurantOrderToPosTransaction(ctx, repository.LinkRestaurantOrderToPosTransactionParams{
		ID:               order.ID,
		PosTransactionID: pgtype.Int4{Int32: txn.ID, Valid: true},
		Status:           targetOrderStatus,
		PaymentStatus:    targetPaymentStatus,
		AmountPaid:       utils.Float64ToPgNumeric(paidFloat),
		ChangeGiven:      utils.Float64ToPgNumeric(changeFloat),
	})
	if err != nil {
		return nil, nil, err
	}

	return &updatedOrder, &txn, nil
}

// SettleOrderWithPosTransaction settles a restaurant order, creating a POS transaction,
// POS payment, and updating the active cashier session expected balance automatically when
// pos_transaction_id is omitted (or linking the provided pos_transaction_id).
func (uc *RestaurantCartUseCase) SettleOrderWithPosTransaction(ctx context.Context, orderID int32, input RestaurantSettleOrderInput) *repository.Response {
	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		order, err := lockAndGetRestaurantOrderByID(ctx, q, orderID)
		if err != nil {
			return restaurantOrderNotFoundOrErr(err, "lock restaurant order")
		}
		if isInactiveOrderStatus(order.Status) {
			return utils.NewResponse(utils.CodeBadReq, "cannot settle a voided or cancelled order", nil)
		}

		calcPaidAndChange := func(o repository.RestaurantOrder) (float64, float64, error) {
			if input.AmountPaid != nil && *input.AmountPaid < 0 {
				return 0, 0, fmt.Errorf("%w: amount_paid cannot be negative", errInvalidRestaurantPayment)
			}
			totalAmt := roundMoney(numericToFloat(o.TotalAmount))
			paidAmt := totalAmt
			if input.AmountPaid != nil {
				paidAmt = roundMoney(*input.AmountPaid)
			} else if numericToFloat(o.AmountPaid) > 0 {
				paidAmt = roundMoney(numericToFloat(o.AmountPaid))
			}
			if err := validateFullRestaurantPayment(paidAmt, totalAmt); err != nil {
				return 0, 0, err
			}
			changeAmt := 0.0
			if paidAmt > totalAmt {
				changeAmt = roundMoney(paidAmt - totalAmt)
			}
			return paidAmt, changeAmt, nil
		}

		if input.PosTransactionID != nil && *input.PosTransactionID > 0 {
			paidAmt, changeAmt, err := calcPaidAndChange(order)
			if err != nil {
				return utils.NewResponse(utils.CodeBadReq, err.Error(), nil)
			}
			updated, err := q.LinkRestaurantOrderToPosTransaction(ctx, repository.LinkRestaurantOrderToPosTransactionParams{
				ID:               orderID,
				PosTransactionID: pgtype.Int4{Int32: *input.PosTransactionID, Valid: true},
				Status:           "paid",
				PaymentStatus:    "paid",
				AmountPaid:       utils.Float64ToPgNumeric(paidAmt),
				ChangeGiven:      utils.Float64ToPgNumeric(changeAmt),
			})
			if err != nil {
				return utils.NewResponse(utils.CodeError, "failed to settle restaurant order: "+err.Error(), nil)
			}
			if order.Status != updated.Status {
				from := order.Status
				recordRestaurantOrderStatusHistory(ctx, q, orderID, updated.StoreID, &from, updated.Status, "order_settled", "Order settled with POS transaction", updated.CreatedByUserID, nil)
			}
			return utils.NewResponse(utils.CodeOK, "order settled successfully", map[string]interface{}{
				"order":              updated,
				"pos_transaction_id": *input.PosTransactionID,
				"transaction_id":     *input.PosTransactionID,
			})
		}

		if order.PosTransactionID.Valid && order.PosTransactionID.Int32 > 0 {
			paidAmt, changeAmt, err := calcPaidAndChange(order)
			if err != nil {
				return utils.NewResponse(utils.CodeBadReq, err.Error(), nil)
			}
			updated, err := q.LinkRestaurantOrderToPosTransaction(ctx, repository.LinkRestaurantOrderToPosTransactionParams{
				ID:               orderID,
				PosTransactionID: order.PosTransactionID,
				Status:           "paid",
				PaymentStatus:    "paid",
				AmountPaid:       utils.Float64ToPgNumeric(paidAmt),
				ChangeGiven:      utils.Float64ToPgNumeric(changeAmt),
			})
			if err != nil {
				return utils.NewResponse(utils.CodeError, "failed to settle restaurant order: "+err.Error(), nil)
			}
			if order.Status != updated.Status {
				from := order.Status
				recordRestaurantOrderStatusHistory(ctx, q, orderID, updated.StoreID, &from, updated.Status, "order_settled", "Order settled", updated.CreatedByUserID, nil)
			}
			return utils.NewResponse(utils.CodeOK, "order settled successfully", map[string]interface{}{
				"order":              updated,
				"pos_transaction_id": order.PosTransactionID.Int32,
				"transaction_id":     order.PosTransactionID.Int32,
			})
		}

		updatedOrder, txn, err := uc.createPosTransactionForRestaurantOrder(
			ctx,
			q,
			order,
			input.CashierID,
			input.CashierSessionID,
			input.PosTerminalID,
			input.PaymentMethod,
			input.PaymentGateway,
			input.ReferenceNumber,
			input.AmountPaid,
			input.Metadata,
			true,
			"paid",
			"paid",
		)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, err.Error(), nil)
		}
		if order.Status != updatedOrder.Status {
			from := order.Status
			recordRestaurantOrderStatusHistory(ctx, q, orderID, updatedOrder.StoreID, &from, updatedOrder.Status, "order_settled", "Order settled with new POS transaction", updatedOrder.CreatedByUserID, nil)
		}

		return utils.NewResponse(utils.CodeOK, "order settled successfully", map[string]interface{}{
			"order":              updatedOrder,
			"pos_transaction_id": txn.ID,
			"transaction_id":     txn.ID,
		})
	})
}

// === Restaurant Promotions / Coupons ===

type restaurantPromoActionMeta struct {
	DiscountPercent float64 `json:"discount_percent"`
	Percentage      float64 `json:"percentage"`
	DiscountAmount  float64 `json:"discount_amount"`
	Amount          float64 `json:"amount"`
	MaxDiscount     float64 `json:"max_discount_amount"`
	BuyQuantity     int     `json:"buy_quantity"`
	GetQuantity     int     `json:"get_quantity"`
	BundlePrice     float64 `json:"bundle_price"`
	TargetPrice     float64 `json:"target_price"`
}

func (uc *RestaurantCartUseCase) ApplyCouponToCart(ctx context.Context, cartID uuid.UUID, couponCode string) *repository.Response {
	if strings.TrimSpace(couponCode) == "" {
		return utils.NewResponse(utils.CodeBadReq, "coupon_code is required", nil)
	}
	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		cart, err := lockAndGetRestaurantCartByID(ctx, q, cartID)
		if err != nil {
			return restaurantCartNotFoundOrErr(err, "lock restaurant cart")
		}
		if cart.CartStatus == repository.CartStatusConverted {
			return utils.NewResponse(utils.CodeBadReq, "cannot modify coupon on a "+string(cart.CartStatus)+" cart", nil)
		}

		updatedCart, promo, errResp := applyRestaurantCouponInTx(ctx, q, cart, couponCode)
		if errResp != nil {
			return errResp
		}

		payload, err := buildCartDetailPayload(ctx, q, updatedCart)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to load restaurant cart details: "+err.Error(), nil)
		}
		payload["applied_promotion"] = promo
		return utils.NewResponse(utils.CodeOK, "restaurant coupon applied successfully", payload)
	})
}

func (uc *RestaurantCartUseCase) RemoveCouponFromCart(ctx context.Context, cartID uuid.UUID) *repository.Response {
	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		cart, err := lockAndGetRestaurantCartByID(ctx, q, cartID)
		if err != nil {
			return restaurantCartNotFoundOrErr(err, "lock restaurant cart")
		}
		if cart.CartStatus == repository.CartStatusConverted {
			return utils.NewResponse(utils.CodeBadReq, "cannot modify coupon on a "+string(cart.CartStatus)+" cart", nil)
		}

		clearedCart, err := clearCouponDiscountsFromCartItems(ctx, q, cart)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to remove coupon from cart: "+err.Error(), nil)
		}

		payload, err := buildCartDetailPayload(ctx, q, clearedCart)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to load restaurant cart details: "+err.Error(), nil)
		}
		return utils.NewResponse(utils.CodeOK, "restaurant coupon removed successfully", payload)
	})
}

func clearCouponDiscountsFromCartItems(ctx context.Context, q *repository.Queries, cart repository.RestaurantCart) (repository.RestaurantCart, error) {
	items, err := q.ListRestaurantCartItems(ctx, cart.ID)
	if err != nil {
		return cart, err
	}

	for _, ci := range items {
		metaMap := unmarshalJSONObject(ci.Metadata)
		rawCouponDisc, hasCouponDisc := metaMap["coupon_discount_amount"]
		if !hasCouponDisc {
			continue
		}
		couponDisc := 0.0
		switch v := rawCouponDisc.(type) {
		case float64:
			couponDisc = v
		}
		delete(metaMap, "coupon_discount_amount")
		delete(metaMap, "applied_coupon_code")

		newDiscount := roundMoney(numericToFloat(ci.DiscountAmount) - couponDisc)
		if newDiscount < 0 {
			newDiscount = 0
		}

		var tc *repository.TaxCategory
		if ci.TaxCategoryID.Valid {
			if fetched, tcErr := q.GetTaxCategory(ctx, ci.TaxCategoryID.Int32); tcErr == nil {
				tc = &fetched
			}
		}

		_, taxAmt, _, lineTotal := computeLineTaxAndTotals(
			numericToFloat(ci.Quantity),
			numericToFloat(ci.UnitPrice),
			numericToFloat(ci.ModifiersTotal),
			numericToFloat(ci.ComboTotal),
			newDiscount,
			tc,
		)

		metaBytes, _ := json.Marshal(metaMap)
		if _, updErr := q.UpdateRestaurantCartItemDiscountAndTotals(ctx, repository.UpdateRestaurantCartItemDiscountAndTotalsParams{
			ID:             ci.ID,
			DiscountAmount: utils.Float64ToPgNumeric(newDiscount),
			TaxAmount:      utils.Float64ToPgNumeric(taxAmt),
			LineTotal:      utils.Float64ToPgNumeric(lineTotal),
			Metadata:       ensureJSONObject(metaBytes),
		}); updErr != nil {
			return cart, updErr
		}
	}

	if _, err := q.ClearCouponFromRestaurantCart(ctx, cart.ID); err != nil {
		return cart, err
	}
	return q.RecalculateRestaurantCartTotals(ctx, cart.ID)
}

func applyRestaurantCouponInTx(ctx context.Context, q *repository.Queries, cart repository.RestaurantCart, rawCode string) (repository.RestaurantCart, repository.RestaurantPromotion, *repository.Response) {
	var emptyPromo repository.RestaurantPromotion
	code := strings.TrimSpace(rawCode)
	if code == "" {
		return cart, emptyPromo, utils.NewResponse(utils.CodeBadReq, "coupon_code is required", nil)
	}

	// Remove any previously applied coupon before evaluating the new code cleanly,
	// preserving any in-memory customer override (e.g. from checkout input).
	requestedCustomerID := cart.CustomerID
	if cart.CouponCode.Valid && strings.TrimSpace(cart.CouponCode.String) != "" {
		cleared, clrErr := clearCouponDiscountsFromCartItems(ctx, q, cart)
		if clrErr != nil {
			return cart, emptyPromo, utils.NewResponse(utils.CodeError, "failed to reset previous coupon: "+clrErr.Error(), nil)
		}
		cart = cleared
		cart.CustomerID = requestedCustomerID
	}

	promo, err := q.GetActiveRestaurantPromotionByCouponCodeInsensitive(ctx, repository.GetActiveRestaurantPromotionByCouponCodeInsensitiveParams{
		CouponCode: code,
		StoreID:    cart.StoreID,
	})
	if err != nil {
		return cart, emptyPromo, utils.NewResponse(utils.CodeNotFound, "invalid or expired restaurant coupon code", nil)
	}

	now := time.Now()
	if promo.ValidFrom.Valid && now.Before(promo.ValidFrom.Time) {
		return cart, emptyPromo, utils.NewResponse(utils.CodeBadReq, "restaurant promotion is not active yet", nil)
	}
	if promo.ValidTo.Valid && now.After(promo.ValidTo.Time) {
		return cart, emptyPromo, utils.NewResponse(utils.CodeBadReq, "restaurant promotion has expired", nil)
	}
	if promo.UsageLimit.Valid && promo.UsageLimit.Int32 > 0 {
		used := int32(0)
		if promo.UsageCount.Valid {
			used = promo.UsageCount.Int32
		}
		if used >= promo.UsageLimit.Int32 {
			return cart, emptyPromo, utils.NewResponse(utils.CodeBadReq, "restaurant promotion usage limit reached", nil)
		}
	}
	if promo.UsagePerCustomer.Valid && promo.UsagePerCustomer.Int32 > 0 {
		if !cart.CustomerID.Valid {
			return cart, emptyPromo, utils.NewResponse(utils.CodeBadReq, "customer_id is required for this restaurant promotion", nil)
		}
		if lockErr := q.LockCustomerRestaurantCouponUsage(ctx, repository.LockCustomerRestaurantCouponUsageParams{
			StoreID:    cart.StoreID,
			CustomerID: cart.CustomerID.Int32,
			CouponCode: code,
		}); lockErr != nil {
			return cart, emptyPromo, utils.NewResponse(utils.CodeError, "failed to lock customer coupon usage: "+lockErr.Error(), nil)
		}
		usedByCust, cntErr := q.CountCustomerRestaurantCouponUses(ctx, repository.CountCustomerRestaurantCouponUsesParams{
			StoreID:    cart.StoreID,
			CustomerID: cart.CustomerID,
			CouponCode: code,
		})
		if cntErr != nil {
			return cart, emptyPromo, utils.NewResponse(utils.CodeError, "failed to verify customer coupon usage: "+cntErr.Error(), nil)
		}
		if usedByCust >= promo.UsagePerCustomer.Int32 {
			return cart, emptyPromo, utils.NewResponse(utils.CodeBadReq, "customer coupon usage limit reached", nil)
		}
	}

	// Validate customer eligibility if restricted
	if len(promo.TargetCustomerTypes) > 0 || len(promo.TargetCustomerTiers) > 0 {
		if !cart.CustomerID.Valid {
			return cart, emptyPromo, utils.NewResponse(utils.CodeBadReq, "customer_id is required for this restaurant promotion", nil)
		}
		cust, custErr := q.GetCustomer(ctx, cart.CustomerID.Int32)
		if custErr != nil {
			return cart, emptyPromo, utils.NewResponse(utils.CodeBadReq, "customer not found for restaurant promotion", nil)
		}
		if len(promo.TargetCustomerTypes) > 0 && !containsIgnoreCase(promo.TargetCustomerTypes, cust.CustomerType.String) {
			return cart, emptyPromo, utils.NewResponse(utils.CodeBadReq, "customer type is not eligible for this promotion", nil)
		}
		if len(promo.TargetCustomerTiers) > 0 && !containsIgnoreCase(promo.TargetCustomerTiers, cust.LoyaltyTier.String) {
			return cart, emptyPromo, utils.NewResponse(utils.CodeBadReq, "customer loyalty tier is not eligible for this promotion", nil)
		}
	}

	cartItems, err := q.ListRestaurantCartItems(ctx, cart.ID)
	if err != nil {
		return cart, emptyPromo, utils.NewResponse(utils.CodeError, "failed to load cart items: "+err.Error(), nil)
	}
	if len(cartItems) == 0 {
		return cart, emptyPromo, utils.NewResponse(utils.CodeBadReq, "cannot apply coupon to an empty cart", nil)
	}

	type eligibleCartLine struct {
		row       repository.ListRestaurantCartItemsRow
		preTaxNet float64
		unitGross float64
		qty       float64
	}

	appliesTo := strings.ToLower(strings.TrimSpace(promo.AppliesTo.String))
	if appliesTo == "" {
		appliesTo = "all"
	}

	eligible := make([]eligibleCartLine, 0, len(cartItems))
	var cartGrossBeforeCoupon, eligibleNetTotal, eligibleQty float64

	for _, ci := range cartItems {
		qty := numericToFloat(ci.Quantity)
		unitGross := numericToFloat(ci.UnitPrice) + numericToFloat(ci.ModifiersTotal) + numericToFloat(ci.ComboTotal)
		lineGross := roundMoney(unitGross * qty)
		existingDisc := numericToFloat(ci.DiscountAmount)
		preTaxNet := roundMoney(math.Max(0, lineGross-existingDisc))
		cartGrossBeforeCoupon += preTaxNet

		matches := appliesTo == "all"
		if !matches {
			mi, miErr := q.GetMenuItem(ctx, ci.MenuItemID)
			if miErr == nil {
				switch appliesTo {
				case "menu_item":
					matches = len(promo.TargetMenuItemIds) == 0 || containsInt32(promo.TargetMenuItemIds, ci.MenuItemID)
				case "menu_category":
					matches = len(promo.TargetMenuCategoryIds) == 0 || (mi.MenuCategoryID > 0 && containsInt32(promo.TargetMenuCategoryIds, mi.MenuCategoryID))
				}
			}
		}

		if matches {
			eligible = append(eligible, eligibleCartLine{
				row:       ci,
				preTaxNet: preTaxNet,
				unitGross: unitGross,
				qty:       qty,
			})
			eligibleNetTotal += preTaxNet
			eligibleQty += qty
		}
	}

	if len(eligible) == 0 {
		return cart, emptyPromo, utils.NewResponse(utils.CodeBadReq, "no items in cart match this restaurant promotion", nil)
	}
	if promo.IsStackable.Valid && !promo.IsStackable.Bool {
		for _, item := range eligible {
			if numericToFloat(item.row.DiscountAmount) > 0 {
				return cart, emptyPromo, utils.NewResponse(utils.CodeBadReq, "promotion cannot be stacked with existing item discounts", nil)
			}
		}
	}

	if minOrder := numericToFloat(promo.MinOrderAmount); minOrder > 0 && roundMoney(cartGrossBeforeCoupon) < minOrder {
		return cart, emptyPromo, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("minimum order amount %.2f required for coupon", minOrder), nil)
	}
	if minQty := numericToFloat(promo.MinQuantity); minQty > 0 && eligibleQty < minQty {
		return cart, emptyPromo, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("minimum quantity %.0f of eligible items required for coupon", minQty), nil)
	}

	var meta restaurantPromoActionMeta
	if len(promo.ActionMetadata) > 0 {
		_ = json.Unmarshal(promo.ActionMetadata, &meta)
	}

	totalPromoDiscount := 0.0
	dv := numericToFloat(promo.DiscountValue)
	switch strings.ToLower(strings.TrimSpace(promo.PromotionType)) {
	case "percentage_discount", "happy_hour":
		pct := dv
		if pct <= 0 {
			if meta.DiscountPercent > 0 {
				pct = meta.DiscountPercent
			} else if meta.Percentage > 0 {
				pct = meta.Percentage
			}
		}
		if pct > 0 {
			totalPromoDiscount = eligibleNetTotal * (pct / 100.0)
			if meta.MaxDiscount > 0 && totalPromoDiscount > meta.MaxDiscount {
				totalPromoDiscount = meta.MaxDiscount
			}
		}
	case "fixed_discount":
		amt := dv
		if amt <= 0 {
			if meta.DiscountAmount > 0 {
				amt = meta.DiscountAmount
			} else if meta.Amount > 0 {
				amt = meta.Amount
			}
		}
		if amt > 0 {
			totalPromoDiscount = math.Min(amt, eligibleNetTotal)
		}
	case "buy_x_get_y", "bogo", "free_item":
		buyQty := meta.BuyQuantity
		if buyQty <= 0 {
			buyQty = 1
		}
		getQty := meta.GetQuantity
		if getQty <= 0 {
			getQty = 1
		}
		discPct := meta.DiscountPercent
		if discPct <= 0 {
			discPct = 100
		}
		freeCount := (int(eligibleQty) / (buyQty + getQty)) * getQty
		if freeCount > 0 {
			minUnit := eligible[0].unitGross
			for _, e := range eligible[1:] {
				if e.unitGross < minUnit {
					minUnit = e.unitGross
				}
			}
			totalPromoDiscount = float64(freeCount) * minUnit * (discPct / 100.0)
		}
	case "bundle_price", "bucket_combo":
		bundlePrice := dv
		if bundlePrice <= 0 {
			if meta.BundlePrice > 0 {
				bundlePrice = meta.BundlePrice
			} else if meta.TargetPrice > 0 {
				bundlePrice = meta.TargetPrice
			}
		}
		if bundlePrice > 0 && eligibleNetTotal > bundlePrice {
			totalPromoDiscount = eligibleNetTotal - bundlePrice
		}
	case "points_multiplier":
		totalPromoDiscount = 0
	default:
		if dv > 0 {
			totalPromoDiscount = math.Min(dv, eligibleNetTotal)
		}
	}

	totalPromoDiscount = roundMoney(math.Min(totalPromoDiscount, eligibleNetTotal))

	// Prorate totalPromoDiscount across eligible items
	allocated := 0.0
	for idx, item := range eligible {
		itemPromoDisc := 0.0
		if eligibleNetTotal > 0 && totalPromoDiscount > 0 {
			if idx == len(eligible)-1 {
				itemPromoDisc = roundMoney(totalPromoDiscount - allocated)
			} else {
				itemPromoDisc = roundMoney(totalPromoDiscount * (item.preTaxNet / eligibleNetTotal))
				allocated += itemPromoDisc
			}
			if itemPromoDisc < 0 {
				itemPromoDisc = 0
			}
			if itemPromoDisc > item.preTaxNet {
				itemPromoDisc = item.preTaxNet
			}
		}

		metaMap := unmarshalJSONObject(item.row.Metadata)
		metaMap["coupon_discount_amount"] = itemPromoDisc
		metaMap["applied_coupon_code"] = code

		newDiscount := roundMoney(numericToFloat(item.row.DiscountAmount) + itemPromoDisc)

		var tc *repository.TaxCategory
		if item.row.TaxCategoryID.Valid {
			if fetched, tcErr := q.GetTaxCategory(ctx, item.row.TaxCategoryID.Int32); tcErr == nil {
				tc = &fetched
			}
		}

		_, taxAmt, _, lineTotal := computeLineTaxAndTotals(
			numericToFloat(item.row.Quantity),
			numericToFloat(item.row.UnitPrice),
			numericToFloat(item.row.ModifiersTotal),
			numericToFloat(item.row.ComboTotal),
			newDiscount,
			tc,
		)

		metaBytes, _ := json.Marshal(metaMap)
		if _, updErr := q.UpdateRestaurantCartItemDiscountAndTotals(ctx, repository.UpdateRestaurantCartItemDiscountAndTotalsParams{
			ID:             item.row.ID,
			DiscountAmount: utils.Float64ToPgNumeric(newDiscount),
			TaxAmount:      utils.Float64ToPgNumeric(taxAmt),
			LineTotal:      utils.Float64ToPgNumeric(lineTotal),
			Metadata:       ensureJSONObject(metaBytes),
		}); updErr != nil {
			return cart, emptyPromo, utils.NewResponse(utils.CodeError, "failed to apply coupon discount to item: "+updErr.Error(), nil)
		}
	}

	if _, err := q.ApplyCouponToRestaurantCart(ctx, repository.ApplyCouponToRestaurantCartParams{
		ID:           cart.ID,
		CouponCode:   pgtype.Text{String: code, Valid: true},
		DiscountCode: pgtype.Text{String: promo.Code, Valid: true},
	}); err != nil {
		return cart, emptyPromo, utils.NewResponse(utils.CodeError, "failed to save coupon on cart: "+err.Error(), nil)
	}

	updatedCart, err := q.RecalculateRestaurantCartTotals(ctx, cart.ID)
	if err != nil {
		return cart, emptyPromo, utils.NewResponse(utils.CodeError, "failed to recalculate cart totals after coupon: "+err.Error(), nil)
	}
	return updatedCart, promo, nil
}

// === Restaurant Order Status History ===

func recordRestaurantOrderStatusHistory(
	ctx context.Context,
	q *repository.Queries,
	orderID, storeID int32,
	fromStatus *string,
	toStatus, reason, notes string,
	changedBy pgtype.Int4,
	metadata json.RawMessage,
) {
	var fromPg pgtype.Text
	if fromStatus != nil && strings.TrimSpace(*fromStatus) != "" {
		fromPg = pgtype.Text{String: strings.TrimSpace(*fromStatus), Valid: true}
	}
	_, _ = q.CreateRestaurantOrderStatusHistory(ctx, repository.CreateRestaurantOrderStatusHistoryParams{
		RestaurantOrderID: orderID,
		StoreID:           storeID,
		FromStatus:        fromPg,
		ToStatus:          strings.TrimSpace(toStatus),
		Reason:            nullableText(reason),
		Notes:             nullableText(notes),
		ChangedByUserID:   changedBy,
		Metadata:          ensureJSONObject(metadata),
	})
}

func (uc *RestaurantCartUseCase) UpdateOrderStatusWithHistory(ctx context.Context, orderID int32, input RestaurantOrderStatusUpdateInput) *repository.Response {
	toStatus := strings.TrimSpace(input.Status)
	if toStatus == "" {
		return utils.NewResponse(utils.CodeBadReq, "status is required", nil)
	}

	return uc.withTx(ctx, func(q *repository.Queries) *repository.Response {
		order, err := lockAndGetRestaurantOrderByID(ctx, q, orderID)
		if err != nil {
			return restaurantOrderNotFoundOrErr(err, "lock restaurant order")
		}

		if isInactiveOrderStatus(order.Status) && !strings.EqualFold(toStatus, order.Status) {
			return utils.NewResponse(utils.CodeBadReq, "cannot transition order out of voided or cancelled status", nil)
		}

		// Reverse recipe ingredients once if transitioning into voided/cancelled from an active status
		if isInactiveOrderStatus(toStatus) && !isInactiveOrderStatus(order.Status) {
			if revErr := reverseRecipeIngredientsForOrder(ctx, q, order, utils.Int32ToPgInt4(input.ChangedByUserID)); revErr != nil {
				return utils.NewResponse(utils.CodeError, "failed to reverse recipe stock deduction: "+revErr.Error(), nil)
			}
		}

		updated, err := q.UpdateRestaurantOrderStatusWithTimestamps(ctx, repository.UpdateRestaurantOrderStatusWithTimestampsParams{
			ID:     orderID,
			Status: toStatus,
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to update restaurant order status: "+err.Error(), nil)
		}

		if order.Status != updated.Status || strings.TrimSpace(input.Reason) != "" || strings.TrimSpace(input.Notes) != "" {
			from := order.Status
			recordRestaurantOrderStatusHistory(
				ctx,
				q,
				orderID,
				updated.StoreID,
				&from,
				updated.Status,
				input.Reason,
				input.Notes,
				utils.Int32ToPgInt4(input.ChangedByUserID),
				input.Metadata,
			)
		}

		return utils.NewResponse(utils.CodeOK, "order status updated successfully", updated)
	})
}

func (uc *RestaurantCartUseCase) GetOrderStatusHistory(ctx context.Context, orderID int32) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}
	if _, err := uc.repo.GetRestaurantOrder(ctx, orderID); err != nil {
		return utils.NewResponse(utils.CodeNotFound, "restaurant order not found", nil)
	}
	rows, err := uc.repo.ListRestaurantOrderStatusHistory(ctx, orderID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch order status history: "+err.Error(), nil)
	}
	if rows == nil {
		rows = []repository.ListRestaurantOrderStatusHistoryRow{}
	}
	return utils.NewResponse(utils.CodeOK, "order status history fetched successfully", rows)
}

// === Recipe Ingredient Stock Deduction & Reversal ===

func deductRecipeIngredientsForOrderItems(
	ctx context.Context,
	q *repository.Queries,
	order repository.RestaurantOrder,
	items []repository.RestaurantOrderItem,
	postedBy pgtype.Int4,
) error {
	type recipeDemand struct {
		recipeID   int32
		menuItemID int32
		orderItem  int32
		multiplier float64
	}

	demands := make([]recipeDemand, 0)
	for _, oi := range items {
		qty := numericToFloat(oi.Quantity)
		if qty <= 0 {
			continue
		}
		if mi, err := q.GetMenuItem(ctx, oi.MenuItemID); err == nil && mi.RecipeID.Valid {
			demands = append(demands, recipeDemand{
				recipeID:   mi.RecipeID.Int32,
				menuItemID: oi.MenuItemID,
				orderItem:  oi.ID,
				multiplier: qty,
			})
		}

		if len(oi.ComboComponentsSnapshot) > 0 {
			var comps []RestaurantComboComponentSnapshotItem
			if jsonErr := json.Unmarshal(oi.ComboComponentsSnapshot, &comps); jsonErr == nil {
				for _, comp := range comps {
					if comp.ComponentMenuItemID <= 0 {
						continue
					}
					if cmi, err := q.GetMenuItem(ctx, comp.ComponentMenuItemID); err == nil && cmi.RecipeID.Valid {
						demands = append(demands, recipeDemand{
							recipeID:   cmi.RecipeID.Int32,
							menuItemID: comp.ComponentMenuItemID,
							orderItem:  oi.ID,
							multiplier: qty,
						})
					}
				}
			}
		}
	}

	for _, d := range demands {
		recipe, err := q.GetRecipe(ctx, d.recipeID)
		if err != nil {
			continue
		}
		yieldQty := numericToFloat(recipe.YieldQuantity)
		if yieldQty <= 0 {
			yieldQty = 1
		}

		ingredients, err := q.ListRecipeIngredients(ctx, d.recipeID)
		if err != nil {
			return err
		}
		for _, ing := range ingredients {
			if ing.IsByproduct.Valid && ing.IsByproduct.Bool {
				continue
			}
			rawQty := (numericToFloat(ing.Quantity) / yieldQty) * d.multiplier
			if rawQty <= 0 {
				continue
			}
			reqQty := math.Round(rawQty*1000) / 1000
			baseNum, convErr := q.ResolveIngredientBaseQuantity(ctx, repository.ResolveIngredientBaseQuantityParams{
				UomID:     ing.UomID,
				ProductID: ing.ProductID,
				Quantity:  utils.Float64ToPgNumeric(reqQty),
			})
			if convErr != nil || !baseNum.Valid || numericToFloat(baseNum) <= 0 {
				baseNum = utils.Float64ToPgNumeric(reqQty)
			}

			deductedRows, err := q.DeductStoreInventoryStockForIngredient(ctx, repository.DeductStoreInventoryStockForIngredientParams{
				DeductQty:        baseNum,
				StoreID:          order.StoreID,
				ProductID:        ing.ProductID,
				ProductVariantID: ing.ProductVariantID,
			})
			if err != nil {
				return err
			}
			if len(deductedRows) == 0 {
				return fmt.Errorf("no inventory stock row found for store %d product %d", order.StoreID, ing.ProductID)
			}

			metaBytes, _ := json.Marshal(map[string]interface{}{
				"order_number":  order.OrderNumber,
				"recipe_id":     d.recipeID,
				"menu_item_id":  d.menuItemID,
				"order_item_id": d.orderItem,
			})
			if _, err := q.CreateRestaurantIngredientStockMovement(ctx, repository.CreateRestaurantIngredientStockMovementParams{
				MovementType:     "recipe_deduction",
				ReferenceID:      pgtype.Int4{Int32: order.ID, Valid: true},
				ProductID:        ing.ProductID,
				ProductVariantID: ing.ProductVariantID,
				FromStoreID:      pgtype.Int4{Int32: order.StoreID, Valid: true},
				Quantity:         baseNum,
				UomID:            ing.UomID,
				PostedBy:         postedBy,
				Metadata:         ensureJSONObject(metaBytes),
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func reverseRecipeIngredientsForOrder(
	ctx context.Context,
	q *repository.Queries,
	order repository.RestaurantOrder,
	postedBy pgtype.Int4,
) error {
	if _, err := q.LockRestaurantOrder(ctx, order.ID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	movements, err := q.ListRestaurantOrderRecipeMovements(ctx, pgtype.Int4{Int32: order.ID, Valid: true})
	if err != nil {
		return err
	}
	if len(movements) == 0 {
		return nil
	}

	type ingKey struct {
		productID int32
		variantID int32
		hasVar    bool
		uomID     int32
		hasUom    bool
	}
	orderKeys := make([]ingKey, 0)
	netQty := make(map[ingKey]float64)

	for _, mv := range movements {
		k := ingKey{
			productID: mv.ProductID,
			variantID: mv.ProductVariantID.Int32,
			hasVar:    mv.ProductVariantID.Valid,
			uomID:     mv.UomID.Int32,
			hasUom:    mv.UomID.Valid,
		}
		if _, exists := netQty[k]; !exists {
			orderKeys = append(orderKeys, k)
		}
		qty := numericToFloat(mv.Quantity)
		switch mv.MovementType {
		case "recipe_deduction":
			netQty[k] += qty
		case "recipe_reversal":
			netQty[k] -= qty
		}
	}

	for _, k := range orderKeys {
		rem := math.Round(netQty[k]*1000) / 1000
		if rem <= 0.0001 {
			continue
		}
		var varPg pgtype.Int4
		if k.hasVar {
			varPg = pgtype.Int4{Int32: k.variantID, Valid: true}
		}
		var uomPg pgtype.Int4
		if k.hasUom {
			uomPg = pgtype.Int4{Int32: k.uomID, Valid: true}
		}
		remNum := utils.Float64ToPgNumeric(rem)

		if _, err := q.RestoreStoreInventoryStockForIngredient(ctx, repository.RestoreStoreInventoryStockForIngredientParams{
			RestoreQty:       remNum,
			StoreID:          order.StoreID,
			ProductID:        k.productID,
			ProductVariantID: varPg,
		}); err != nil {
			return err
		}

		metaBytes, _ := json.Marshal(map[string]interface{}{
			"order_number": order.OrderNumber,
			"reason":       "order_reversal",
		})
		if _, err := q.CreateRestaurantIngredientStockMovement(ctx, repository.CreateRestaurantIngredientStockMovementParams{
			MovementType:     "recipe_reversal",
			ReferenceID:      pgtype.Int4{Int32: order.ID, Valid: true},
			ProductID:        k.productID,
			ProductVariantID: varPg,
			FromStoreID:      pgtype.Int4{Int32: order.StoreID, Valid: true},
			Quantity:         remNum,
			UomID:            uomPg,
			PostedBy:         postedBy,
			Metadata:         ensureJSONObject(metaBytes),
		}); err != nil {
			return err
		}
	}

	return nil
}

func containsInt32(slice []int32, val int32) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

func containsIgnoreCase(slice []string, val string) bool {
	v := strings.TrimSpace(val)
	for _, s := range slice {
		if strings.EqualFold(strings.TrimSpace(s), v) {
			return true
		}
	}
	return false
}

func nullableText(s string) pgtype.Text {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: trimmed, Valid: true}
}

// Keep pgx imported if needed by future extensions; compile check guard.
var _ pgx.Tx = (pgx.Tx)(nil)
