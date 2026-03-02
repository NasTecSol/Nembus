package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"strconv"
	"time"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CartUseCase struct {
	repo *repository.Queries
}

func NewCartUseCase() *CartUseCase {
	return &CartUseCase{}
}

func (uc *CartUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *CartUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// CartOutput is the response shape for cart APIs with JSONB fields marshaled as JSON.
type CartOutput struct {
	ID                 uuid.UUID             `json:"id"`
	CartNumber         string                `json:"cart_number"`
	OrganizationID     int32                 `json:"organization_id"`
	StoreID            pgtype.Int4           `json:"store_id"`
	CustomerID         pgtype.Int4           `json:"customer_id"`
	GuestIdentifier    pgtype.Text           `json:"guest_identifier"`
	GuestEmail         pgtype.Text           `json:"guest_email"`
	GuestPhone         pgtype.Text           `json:"guest_phone"`
	CartStatus         repository.CartStatus `json:"cart_status"`
	CartType           repository.CartType   `json:"cart_type"`
	Channel            pgtype.Text           `json:"channel"`
	PaymentMethod      pgtype.Text           `json:"payment_method"`
	PaymentGateway     pgtype.Text           `json:"payment_gateway"`
	DeviceInfo         json.RawMessage       `json:"device_info"`
	CreatedByUserID    pgtype.Int4           `json:"created_by_user_id"`
	CashierID          pgtype.Int4           `json:"cashier_id"`
	PosTerminalID      pgtype.Int4           `json:"pos_terminal_id"`
	Subtotal           pgtype.Numeric        `json:"subtotal"`
	DiscountAmount     pgtype.Numeric        `json:"discount_amount"`
	TaxAmount          pgtype.Numeric        `json:"tax_amount"`
	ShippingAmount     pgtype.Numeric        `json:"shipping_amount"`
	TotalAmount        pgtype.Numeric        `json:"total_amount"`
	CouponCode         pgtype.Text           `json:"coupon_code"`
	DiscountCode       pgtype.Text           `json:"discount_code"`
	PromotionalCredits pgtype.Numeric        `json:"promotional_credits"`
	ShippingAddress    json.RawMessage       `json:"shipping_address"`
	BillingAddress     json.RawMessage       `json:"billing_address"`
	ShippingMethod     pgtype.Text           `json:"shipping_method"`
	ConvertedToOrderID pgtype.UUID           `json:"converted_to_order_id"`
	ConvertedAt        pgtype.Timestamp      `json:"converted_at"`
	LastActivityAt     pgtype.Timestamp      `json:"last_activity_at"`
	ExpiresAt          pgtype.Timestamp      `json:"expires_at"`
	CreatedAt          pgtype.Timestamp      `json:"created_at"`
	UpdatedAt          pgtype.Timestamp      `json:"updated_at"`
	Metadata           json.RawMessage       `json:"metadata"`
	Notes              pgtype.Text           `json:"notes"`
}

type CartItemOutput struct {
	ID                   uuid.UUID        `json:"id"`
	CartID               uuid.UUID        `json:"cart_id"`
	OrganizationID       int32            `json:"organization_id"`
	ProductID            int32            `json:"product_id"`
	ProductVariantID     pgtype.Int4      `json:"product_variant_id"`
	Quantity             pgtype.Numeric   `json:"quantity"`
	UomID                pgtype.Int4      `json:"uom_id"`
	UnitPrice            pgtype.Numeric   `json:"unit_price"`
	DiscountAmount       pgtype.Numeric   `json:"discount_amount"`
	TaxAmount            pgtype.Numeric   `json:"tax_amount"`
	LineTotal            pgtype.Numeric   `json:"line_total"`
	PriceListID          pgtype.Int4      `json:"price_list_id"`
	TaxCategoryID        pgtype.Int4      `json:"tax_category_id"`
	BatchNumber          pgtype.Text      `json:"batch_number"`
	SerialNumber         pgtype.Text      `json:"serial_number"`
	CustomizationDetails json.RawMessage  `json:"customization_details"`
	Notes                pgtype.Text      `json:"notes"`
	Metadata             json.RawMessage  `json:"metadata"`
	AddedAt              pgtype.Timestamp `json:"added_at"`
	UpdatedAt            pgtype.Timestamp `json:"updated_at"`
}

type ListCartItemsOutput struct {
	ID                   uuid.UUID        `json:"id"`
	CartID               uuid.UUID        `json:"cart_id"`
	OrganizationID       int32            `json:"organization_id"`
	ProductID            int32            `json:"product_id"`
	ProductVariantID     pgtype.Int4      `json:"product_variant_id"`
	Quantity             pgtype.Numeric   `json:"quantity"`
	UomID                pgtype.Int4      `json:"uom_id"`
	UnitPrice            pgtype.Numeric   `json:"unit_price"`
	DiscountAmount       pgtype.Numeric   `json:"discount_amount"`
	TaxAmount            pgtype.Numeric   `json:"tax_amount"`
	LineTotal            pgtype.Numeric   `json:"line_total"`
	PriceListID          pgtype.Int4      `json:"price_list_id"`
	TaxCategoryID        pgtype.Int4      `json:"tax_category_id"`
	BatchNumber          pgtype.Text      `json:"batch_number"`
	SerialNumber         pgtype.Text      `json:"serial_number"`
	CustomizationDetails json.RawMessage  `json:"customization_details"`
	Notes                pgtype.Text      `json:"notes"`
	Metadata             json.RawMessage  `json:"metadata"`
	AddedAt              pgtype.Timestamp `json:"added_at"`
	UpdatedAt            pgtype.Timestamp `json:"updated_at"`
	ProductName          string           `json:"product_name"`
	ProductSku           string           `json:"product_sku"`
}

type CartActivityLogOutput struct {
	ID                int64            `json:"id"`
	CartID            uuid.UUID        `json:"cart_id"`
	OrganizationID    int32            `json:"organization_id"`
	ActivityType      string           `json:"activity_type"`
	Description       pgtype.Text      `json:"description"`
	PerformedByUserID pgtype.Int4      `json:"performed_by_user_id"`
	IpAddress         *netip.Addr      `json:"ip_address"`
	UserAgent         pgtype.Text      `json:"user_agent"`
	OldValue          json.RawMessage  `json:"old_value"`
	NewValue          json.RawMessage  `json:"new_value"`
	CreatedAt         pgtype.Timestamp `json:"created_at"`
}

type ListAbandonedCartOutput struct {
	ID                 uuid.UUID             `json:"id"`
	CartNumber         string                `json:"cart_number"`
	OrganizationID     int32                 `json:"organization_id"`
	StoreID            pgtype.Int4           `json:"store_id"`
	CustomerID         pgtype.Int4           `json:"customer_id"`
	GuestIdentifier    pgtype.Text           `json:"guest_identifier"`
	GuestEmail         pgtype.Text           `json:"guest_email"`
	GuestPhone         pgtype.Text           `json:"guest_phone"`
	CartStatus         repository.CartStatus `json:"cart_status"`
	CartType           repository.CartType   `json:"cart_type"`
	Channel            pgtype.Text           `json:"channel"`
	PaymentMethod      pgtype.Text           `json:"payment_method"`
	PaymentGateway     pgtype.Text           `json:"payment_gateway"`
	DeviceInfo         json.RawMessage       `json:"device_info"`
	CreatedByUserID    pgtype.Int4           `json:"created_by_user_id"`
	CashierID          pgtype.Int4           `json:"cashier_id"`
	PosTerminalID      pgtype.Int4           `json:"pos_terminal_id"`
	Subtotal           pgtype.Numeric        `json:"subtotal"`
	DiscountAmount     pgtype.Numeric        `json:"discount_amount"`
	TaxAmount          pgtype.Numeric        `json:"tax_amount"`
	ShippingAmount     pgtype.Numeric        `json:"shipping_amount"`
	TotalAmount        pgtype.Numeric        `json:"total_amount"`
	CouponCode         pgtype.Text           `json:"coupon_code"`
	DiscountCode       pgtype.Text           `json:"discount_code"`
	PromotionalCredits pgtype.Numeric        `json:"promotional_credits"`
	ShippingAddress    json.RawMessage       `json:"shipping_address"`
	BillingAddress     json.RawMessage       `json:"billing_address"`
	ShippingMethod     pgtype.Text           `json:"shipping_method"`
	ConvertedToOrderID pgtype.UUID           `json:"converted_to_order_id"`
	ConvertedAt        pgtype.Timestamp      `json:"converted_at"`
	LastActivityAt     pgtype.Timestamp      `json:"last_activity_at"`
	ExpiresAt          pgtype.Timestamp      `json:"expires_at"`
	CreatedAt          pgtype.Timestamp      `json:"created_at"`
	UpdatedAt          pgtype.Timestamp      `json:"updated_at"`
	Metadata           json.RawMessage       `json:"metadata"`
	Notes              pgtype.Text           `json:"notes"`
	ItemCount          int64                 `json:"item_count"`
	CartValue          int64                 `json:"cart_value"`
}

func cartToOutput(c repository.Cart) CartOutput {
	return CartOutput{
		ID:                 c.ID,
		CartNumber:         c.CartNumber,
		OrganizationID:     c.OrganizationID,
		StoreID:            c.StoreID,
		CustomerID:         c.CustomerID,
		GuestIdentifier:    c.GuestIdentifier,
		GuestEmail:         c.GuestEmail,
		GuestPhone:         c.GuestPhone,
		CartStatus:         c.CartStatus,
		CartType:           c.CartType,
		Channel:            c.Channel,
		PaymentMethod:      c.PaymentMethod,
		PaymentGateway:     c.PaymentGateway,
		DeviceInfo:         utils.BytesToJSONRawMessage(c.DeviceInfo),
		CreatedByUserID:    c.CreatedByUserID,
		CashierID:          c.CashierID,
		PosTerminalID:      c.PosTerminalID,
		Subtotal:           c.Subtotal,
		DiscountAmount:     c.DiscountAmount,
		TaxAmount:          c.TaxAmount,
		ShippingAmount:     c.ShippingAmount,
		TotalAmount:        c.TotalAmount,
		CouponCode:         c.CouponCode,
		DiscountCode:       c.DiscountCode,
		PromotionalCredits: c.PromotionalCredits,
		ShippingAddress:    utils.BytesToJSONRawMessage(c.ShippingAddress),
		BillingAddress:     utils.BytesToJSONRawMessage(c.BillingAddress),
		ShippingMethod:     c.ShippingMethod,
		ConvertedToOrderID: c.ConvertedToOrderID,
		ConvertedAt:        c.ConvertedAt,
		LastActivityAt:     c.LastActivityAt,
		ExpiresAt:          c.ExpiresAt,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
		Metadata:           utils.BytesToJSONRawMessage(c.Metadata),
		Notes:              c.Notes,
	}
}

func cartItemToOutput(i repository.CartItem) CartItemOutput {
	return CartItemOutput{
		ID:                   i.ID,
		CartID:               i.CartID,
		OrganizationID:       i.OrganizationID,
		ProductID:            i.ProductID,
		ProductVariantID:     i.ProductVariantID,
		Quantity:             i.Quantity,
		UomID:                i.UomID,
		UnitPrice:            i.UnitPrice,
		DiscountAmount:       i.DiscountAmount,
		TaxAmount:            i.TaxAmount,
		LineTotal:            i.LineTotal,
		PriceListID:          i.PriceListID,
		TaxCategoryID:        i.TaxCategoryID,
		BatchNumber:          i.BatchNumber,
		SerialNumber:         i.SerialNumber,
		CustomizationDetails: utils.BytesToJSONRawMessage(i.CustomizationDetails),
		Notes:                i.Notes,
		Metadata:             utils.BytesToJSONRawMessage(i.Metadata),
		AddedAt:              i.AddedAt,
		UpdatedAt:            i.UpdatedAt,
	}
}

func listCartItemsToOutput(rows []repository.ListCartItemsRow) []ListCartItemsOutput {
	out := make([]ListCartItemsOutput, 0, len(rows))
	for _, r := range rows {
		out = append(out, ListCartItemsOutput{
			ID:                   r.ID,
			CartID:               r.CartID,
			OrganizationID:       r.OrganizationID,
			ProductID:            r.ProductID,
			ProductVariantID:     r.ProductVariantID,
			Quantity:             r.Quantity,
			UomID:                r.UomID,
			UnitPrice:            r.UnitPrice,
			DiscountAmount:       r.DiscountAmount,
			TaxAmount:            r.TaxAmount,
			LineTotal:            r.LineTotal,
			PriceListID:          r.PriceListID,
			TaxCategoryID:        r.TaxCategoryID,
			BatchNumber:          r.BatchNumber,
			SerialNumber:         r.SerialNumber,
			CustomizationDetails: utils.BytesToJSONRawMessage(r.CustomizationDetails),
			Notes:                r.Notes,
			Metadata:             utils.BytesToJSONRawMessage(r.Metadata),
			AddedAt:              r.AddedAt,
			UpdatedAt:            r.UpdatedAt,
			ProductName:          r.ProductName,
			ProductSku:           r.ProductSku,
		})
	}
	return out
}

func cartActivityToOutput(a repository.CartActivityLog) CartActivityLogOutput {
	return CartActivityLogOutput{
		ID:                a.ID,
		CartID:            a.CartID,
		OrganizationID:    a.OrganizationID,
		ActivityType:      a.ActivityType,
		Description:       a.Description,
		PerformedByUserID: a.PerformedByUserID,
		IpAddress:         a.IpAddress,
		UserAgent:         a.UserAgent,
		OldValue:          utils.BytesToJSONRawMessage(a.OldValue),
		NewValue:          utils.BytesToJSONRawMessage(a.NewValue),
		CreatedAt:         a.CreatedAt,
	}
}

func abandonedCartToOutput(c repository.ListAbandonedCartsRow) ListAbandonedCartOutput {
	return ListAbandonedCartOutput{
		ID:                 c.ID,
		CartNumber:         c.CartNumber,
		OrganizationID:     c.OrganizationID,
		StoreID:            c.StoreID,
		CustomerID:         c.CustomerID,
		GuestIdentifier:    c.GuestIdentifier,
		GuestEmail:         c.GuestEmail,
		GuestPhone:         c.GuestPhone,
		CartStatus:         c.CartStatus,
		CartType:           c.CartType,
		Channel:            c.Channel,
		PaymentMethod:      c.PaymentMethod,
		PaymentGateway:     c.PaymentGateway,
		DeviceInfo:         utils.BytesToJSONRawMessage(c.DeviceInfo),
		CreatedByUserID:    c.CreatedByUserID,
		CashierID:          c.CashierID,
		PosTerminalID:      c.PosTerminalID,
		Subtotal:           c.Subtotal,
		DiscountAmount:     c.DiscountAmount,
		TaxAmount:          c.TaxAmount,
		ShippingAmount:     c.ShippingAmount,
		TotalAmount:        c.TotalAmount,
		CouponCode:         c.CouponCode,
		DiscountCode:       c.DiscountCode,
		PromotionalCredits: c.PromotionalCredits,
		ShippingAddress:    utils.BytesToJSONRawMessage(c.ShippingAddress),
		BillingAddress:     utils.BytesToJSONRawMessage(c.BillingAddress),
		ShippingMethod:     c.ShippingMethod,
		ConvertedToOrderID: c.ConvertedToOrderID,
		ConvertedAt:        c.ConvertedAt,
		LastActivityAt:     c.LastActivityAt,
		ExpiresAt:          c.ExpiresAt,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
		Metadata:           utils.BytesToJSONRawMessage(c.Metadata),
		Notes:              c.Notes,
		ItemCount:          c.ItemCount,
		CartValue:          c.CartValue,
	}
}

func cartsToOutput(carts []repository.Cart) []CartOutput {
	out := make([]CartOutput, 0, len(carts))
	for _, c := range carts {
		out = append(out, cartToOutput(c))
	}
	return out
}

func abandonedCartsToOutput(carts []repository.ListAbandonedCartsRow) []ListAbandonedCartOutput {
	out := make([]ListAbandonedCartOutput, 0, len(carts))
	for _, c := range carts {
		out = append(out, abandonedCartToOutput(c))
	}
	return out
}

func cartActivitiesToOutput(rows []repository.CartActivityLog) []CartActivityLogOutput {
	out := make([]CartActivityLogOutput, 0, len(rows))
	for _, r := range rows {
		out = append(out, cartActivityToOutput(r))
	}
	return out
}

// GetCart retrieves a cart by its ID.
func (uc *CartUseCase) GetCart(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	cart, err := uc.repo.GetCartByID(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cart not found", nil)
	}

	items, err := uc.repo.ListCartItems(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch cart items", nil)
	}

	result := map[string]interface{}{
		"cart":  cartToOutput(cart),
		"items": listCartItemsToOutput(items),
	}

	return utils.NewResponse(utils.CodeOK, "cart fetched successfully", result)
}

func (uc *CartUseCase) CreateCart(ctx context.Context, arg repository.CreateCartParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.CreateCart(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create cart", err.Error())
	}
	return utils.NewResponse(utils.CodeCreated, "cart created successfully", cartToOutput(cart))
}

// CreateNewCart is a convenience helper that builds CreateCartParams from
// common inputs, generates a cart number, and calls the repository CreateCart.
func (uc *CartUseCase) CreateNewCart(ctx context.Context, organizationID int32, storeID pgtype.Int4, customerID pgtype.Int4, guestIdentifier string, guestEmail string, guestPhone string, paymentMethod string, paymentGateway string, createdByUserID pgtype.Int4, cashierID pgtype.Int4, posTerminalID pgtype.Int4, metadata []byte, notes string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	arg := repository.CreateCartParams{
		CartNumber:      fmt.Sprintf("CART-%s", uuid.NewString()),
		OrganizationID:  organizationID,
		StoreID:         storeID,
		CustomerID:      customerID,
		GuestIdentifier: pgtype.Text{String: guestIdentifier, Valid: guestIdentifier != ""},
		GuestEmail:      pgtype.Text{String: guestEmail, Valid: guestEmail != ""},
		GuestPhone:      pgtype.Text{String: guestPhone, Valid: guestPhone != ""},
		CartStatus:      repository.CartStatusActive,
		CartType:        repository.CartTypeStandard,
		Channel:         pgtype.Text{Valid: false},
		PaymentMethod:   pgtype.Text{String: paymentMethod, Valid: paymentMethod != ""},
		PaymentGateway:  pgtype.Text{String: paymentGateway, Valid: paymentGateway != ""},
		DeviceInfo:      nil,
		CreatedByUserID: createdByUserID,
		CashierID:       cashierID,
		PosTerminalID:   posTerminalID,
		ShippingAddress: nil,
		BillingAddress:  nil,
		ShippingMethod:  pgtype.Text{Valid: false},
		CouponCode:      pgtype.Text{Valid: false},
		DiscountCode:    pgtype.Text{Valid: false},
		ExpiresAt:       pgtype.Timestamp{},
		Notes:           pgtype.Text{String: notes, Valid: notes != ""},
		Metadata:        metadata,
	}

	cart, err := uc.repo.CreateCart(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create cart", err.Error())
	}
	return utils.NewResponse(utils.CodeCreated, "cart created successfully", cartToOutput(cart))
}

func (uc *CartUseCase) GetCartByNumber(ctx context.Context, cartNumber string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.GetCartByNumber(ctx, cartNumber)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cart not found", nil)
	}
	items, err := uc.repo.ListCartItems(ctx, cart.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch cart items", nil)
	}
	return utils.NewResponse(utils.CodeOK, "cart fetched successfully", map[string]interface{}{
		"cart":  cartToOutput(cart),
		"items": listCartItemsToOutput(items),
	})
}

func (uc *CartUseCase) GetActiveCartByCustomer(ctx context.Context, arg repository.GetCartByCustomerParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.GetCartByCustomer(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "active cart not found", nil)
	}
	items, err := uc.repo.ListCartItems(ctx, cart.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch cart items", nil)
	}
	return utils.NewResponse(utils.CodeOK, "cart fetched successfully", map[string]interface{}{
		"cart":  cartToOutput(cart),
		"items": listCartItemsToOutput(items),
	})
}

func (uc *CartUseCase) GetActiveCartByGuestIdentifier(ctx context.Context, arg repository.GetCartByGuestIdentifierParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.GetCartByGuestIdentifier(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "active cart not found", nil)
	}
	items, err := uc.repo.ListCartItems(ctx, cart.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch cart items", nil)
	}
	return utils.NewResponse(utils.CodeOK, "cart fetched successfully", map[string]interface{}{
		"cart":  cartToOutput(cart),
		"items": listCartItemsToOutput(items),
	})
}

func (uc *CartUseCase) ListActiveCarts(ctx context.Context, arg repository.ListActiveCartsParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	carts, err := uc.repo.ListActiveCarts(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list active carts", nil)
	}
	return utils.NewResponse(utils.CodeOK, "active carts listed", cartsToOutput(carts))
}

func (uc *CartUseCase) ListAbandonedCarts(ctx context.Context, arg repository.ListAbandonedCartsParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	carts, err := uc.repo.ListAbandonedCarts(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list abandoned carts", nil)
	}
	return utils.NewResponse(utils.CodeOK, "abandoned carts listed", abandonedCartsToOutput(carts))
}

func (uc *CartUseCase) UpdateCart(ctx context.Context, arg repository.UpdateCartParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.UpdateCart(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update cart", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart updated successfully", cartToOutput(cart))
}

func (uc *CartUseCase) UpdateCartStatus(ctx context.Context, arg repository.UpdateCartStatusParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.UpdateCartStatus(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update cart status", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart status updated", cartToOutput(cart))
}

func (uc *CartUseCase) UpdateCartCustomer(ctx context.Context, arg repository.UpdateCartCustomerParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.UpdateCartCustomer(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update cart customer", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart customer updated", cartToOutput(cart))
}

func (uc *CartUseCase) DeleteCart(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if err := uc.repo.DeleteCart(ctx, cartID); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to delete cart", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart deleted", nil)
}

func (uc *CartUseCase) ExpireAbandonedCarts(ctx context.Context, storeID pgtype.Int4) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if err := uc.repo.ExpireAbandonedCarts(ctx, storeID); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to expire carts", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "carts expired", nil)
}

func (uc *CartUseCase) ListCartItems(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	items, err := uc.repo.ListCartItems(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list cart items", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart items listed", listCartItemsToOutput(items))
}

func (uc *CartUseCase) CreateCartItem(ctx context.Context, arg repository.CreateCartItemParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	item, err := uc.repo.CreateCartItem(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create cart item", err.Error())
	}
	_, _ = uc.repo.RecalculateCartTotals(ctx, arg.CartID)
	return utils.NewResponse(utils.CodeCreated, "cart item created", cartItemToOutput(item))
}

func (uc *CartUseCase) GetCartItem(ctx context.Context, itemID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	item, err := uc.repo.GetCartItem(ctx, itemID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cart item not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "cart item fetched", cartItemToOutput(item))
}

func (uc *CartUseCase) GetCartItemByProduct(ctx context.Context, arg repository.GetCartItemByProductParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	item, err := uc.repo.GetCartItemByProduct(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cart item not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "cart item fetched", cartItemToOutput(item))
}

func (uc *CartUseCase) UpdateCartItem(ctx context.Context, arg repository.UpdateCartItemParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	item, err := uc.repo.UpdateCartItem(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update cart item", err.Error())
	}
	_, _ = uc.repo.RecalculateCartTotals(ctx, item.CartID)
	return utils.NewResponse(utils.CodeOK, "cart item updated", cartItemToOutput(item))
}

func (uc *CartUseCase) UpdateCartItemQuantity(ctx context.Context, arg repository.UpdateCartItemQuantityParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	item, err := uc.repo.UpdateCartItemQuantity(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update cart item quantity", err.Error())
	}
	_, _ = uc.repo.RecalculateCartTotals(ctx, item.CartID)
	return utils.NewResponse(utils.CodeOK, "cart item quantity updated", cartItemToOutput(item))
}

func (uc *CartUseCase) DeleteCartItem(ctx context.Context, itemID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	item, err := uc.repo.GetCartItem(ctx, itemID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cart item not found", nil)
	}
	if err := uc.repo.DeleteCartItem(ctx, itemID); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to delete cart item", err.Error())
	}
	_, _ = uc.repo.RecalculateCartTotals(ctx, item.CartID)
	return utils.NewResponse(utils.CodeOK, "cart item deleted", nil)
}

func (uc *CartUseCase) ClearCartItems(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if err := uc.repo.ClearCartItems(ctx, cartID); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to clear cart items", err.Error())
	}
	_, _ = uc.repo.RecalculateCartTotals(ctx, cartID)
	return utils.NewResponse(utils.CodeOK, "cart items cleared", nil)
}

func (uc *CartUseCase) GetCartItemCount(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	row, err := uc.repo.GetCartItemCount(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to get cart item count", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart item count fetched", row)
}

func (uc *CartUseCase) GetCartTotals(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	row, err := uc.repo.GetCartTotals(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to get cart totals", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart totals fetched", row)
}

func (uc *CartUseCase) CreateCartActivity(ctx context.Context, arg repository.CreateCartActivityParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	logRow, err := uc.repo.CreateCartActivity(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create cart activity", err.Error())
	}
	return utils.NewResponse(utils.CodeCreated, "cart activity created", cartActivityToOutput(logRow))
}

func (uc *CartUseCase) ListCartActivities(ctx context.Context, arg repository.ListCartActivitiesParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	rows, err := uc.repo.ListCartActivities(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list cart activities", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart activities listed", cartActivitiesToOutput(rows))
}

func (uc *CartUseCase) ApplyCouponToCart(ctx context.Context, arg repository.ApplyCouponToCartParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.ApplyCouponToCart(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to apply coupon", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "coupon applied", cartToOutput(cart))
}

func (uc *CartUseCase) RecalculateCartTotals(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.RecalculateCartTotals(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to recalculate totals", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart totals recalculated", cartToOutput(cart))
}

func (uc *CartUseCase) MergeGuestCartToCustomer(ctx context.Context, arg repository.MergeGuestCartToCustomerParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if err := uc.repo.MergeGuestCartToCustomer(ctx, arg); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to merge carts", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "carts merged", nil)
}

// AddToCart adds an item to the cart.
func (uc *CartUseCase) AddToCart(ctx context.Context, cartID uuid.UUID, orgID int32, productID int32, productVariantID *int32, qty float64, uomID int32, priceListID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	// 1. Check if product exists
	product, err := uc.repo.GetProduct(ctx, productID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "product not found", nil)
	}

	// 2. Convert quantity to pgtype.Numeric for query
	qtyNumeric := pgtype.Numeric{}
	if err := qtyNumeric.Scan(fmt.Sprintf("%.2f", qty)); err != nil {
		return utils.NewResponse(utils.CodeError, "invalid quantity", nil)
	}

	// Prepare variant parameter
	var variantParam pgtype.Int4

	if productVariantID != nil {
		variantParam = pgtype.Int4{
			Int32: *productVariantID,
			Valid: true,
		}
	} else {
		variantParam = pgtype.Int4{
			Valid: false,
		}
	}

	// 3. Get price using GetProductPriceForList with uom_id and price_list_id
	productPrice, err := uc.repo.GetProductPriceForList(ctx, repository.GetProductPriceForListParams{
		ProductID:        productID,
		PriceListID:      priceListID,
		UomID:            pgtype.Int4{Int32: uomID, Valid: true},
		ProductVariantID: variantParam,
		Quantity:         qtyNumeric,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("product price not found for the specified price list and UOM: %v", err), nil)
	}

	// 4. Calculate line total (quantity * unit price)
	// Convert Numeric to string, parse as float64, multiply, convert back
	var priceFloat, qtyFloat float64

	// Get price value
	if val, err := productPrice.Price.Value(); err == nil {
		if str, ok := val.(string); ok {
			priceFloat, _ = strconv.ParseFloat(str, 64)
		} else {
			priceFloat, _ = strconv.ParseFloat(fmt.Sprintf("%v", val), 64)
		}
	} else {
		// Fallback: try to get as string directly
		priceStr := fmt.Sprintf("%v", productPrice.Price)
		priceFloat, _ = strconv.ParseFloat(priceStr, 64)
	}

	// Get quantity value
	if val, err := qtyNumeric.Value(); err == nil {
		if str, ok := val.(string); ok {
			qtyFloat, _ = strconv.ParseFloat(str, 64)
		} else {
			qtyFloat, _ = strconv.ParseFloat(fmt.Sprintf("%v", val), 64)
		}
	} else {
		// Fallback: use the original qty float64
		qtyFloat = qty
	}

	lineTotalFloat := priceFloat * qtyFloat

	// 5. Calculate tax amount if tax category exists
	var taxAmountFloat float64
	var taxAmount pgtype.Numeric
	taxAmount = pgtype.Numeric{Valid: false} // Default to no tax

	if product.TaxCategoryID.Valid {
		// Get tax category to get tax rate
		taxCategory, err := uc.repo.GetTaxCategory(ctx, product.TaxCategoryID.Int32)
		if err == nil && taxCategory.IsActive.Bool {
			// Get tax rate as float64
			var taxRateFloat float64
			if val, err := taxCategory.TaxRate.Value(); err == nil {
				if str, ok := val.(string); ok {
					taxRateFloat, _ = strconv.ParseFloat(str, 64)
				} else {
					taxRateFloat, _ = strconv.ParseFloat(fmt.Sprintf("%v", val), 64)
				}
			}

			// Calculate tax amount based on whether tax is inclusive or exclusive
			if taxCategory.IsInclusive.Bool {
				// Tax is included in price: tax = line_total * (tax_rate / (100 + tax_rate))
				taxAmountFloat = lineTotalFloat * (taxRateFloat / (100 + taxRateFloat))
			} else {
				// Tax is added on top: tax = line_total * (tax_rate / 100)
				taxAmountFloat = lineTotalFloat * (taxRateFloat / 100)
				// Update line total to include tax
				lineTotalFloat = lineTotalFloat + taxAmountFloat
			}

			// Convert tax amount to pgtype.Numeric
			taxAmount = pgtype.Numeric{}
			if err := taxAmount.Scan(fmt.Sprintf("%.2f", taxAmountFloat)); err != nil {
				return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to convert tax amount: %v", err), nil)
			}
		}
	}

	// Convert final line total to pgtype.Numeric
	lineTotal := pgtype.Numeric{}
	if err := lineTotal.Scan(fmt.Sprintf("%.2f", lineTotalFloat)); err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to convert line total: %v", err), nil)
	}

	cartItem, err := uc.repo.CreateCartItem(ctx, repository.CreateCartItemParams{
		CartID:           cartID,
		OrganizationID:   orgID,
		ProductID:        productID,
		ProductVariantID: productPrice.ProductVariantID, // already pgtype.Int4
		Quantity:         qtyNumeric,
		UomID:            productPrice.UomID, // already pgtype.Int4
		UnitPrice:        productPrice.Price,
		TaxAmount:        taxAmount,
		LineTotal:        lineTotal,
		PriceListID:      pgtype.Int4{Int32: productPrice.PriceListID, Valid: true},
		TaxCategoryID:    pgtype.Int4{Int32: product.TaxCategoryID.Int32, Valid: product.TaxCategoryID.Valid},
	})

	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to add item: %v", err), nil)
	}

	// 6. Recalculate cart totals
	_, _ = uc.repo.RecalculateCartTotals(ctx, cartID)

	return utils.NewResponse(utils.CodeOK, "item added to cart", map[string]interface{}{
		"item_id":            cartItem.ID,
		"product_name":       product.Name,
		"quantity":           qty,
		"uom_id":             cartItem.UomID.Int32,            // UOM used
		"price_list_id":      cartItem.PriceListID.Int32,      // price list used
		"unit_price":         cartItem.UnitPrice,              // unit price
		"tax_amount":         cartItem.TaxAmount,              // tax amount
		"line_total":         cartItem.LineTotal,              // line total (includes tax if exclusive)
		"product_variant_id": cartItem.ProductVariantID.Int32, // variant if any
	})

}

// ConvertToOrder converts a cart to an order.
func (uc *CartUseCase) ConvertToOrder(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	// 1. Get Cart using generated method
	cart, err := uc.repo.GetCartByID(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cart not found", nil)
	}

	if cart.CartStatus == repository.CartStatusConverted {
		return utils.NewResponse(utils.CodeBadReq, "cart already converted", nil)
	}

	// 2. Fetch items using generated method
	items, err := uc.repo.ListCartItems(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch cart items", nil)
	}

	// 3. Convert Cart to Order Header using the new generated business logic query
	orderNumber := fmt.Sprintf("ORD-%s", time.Now().Format("20060102150405"))
	convertedCart, err := uc.repo.ConvertCartToOrder(ctx, repository.ConvertCartToOrderParams{
		ID:          cartID,
		OrderNumber: orderNumber,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, "conversion failed: "+err.Error(), nil)
	}

	orderID := convertedCart.ConvertedToOrderID

	// 4. Create Order Lines with all fields from cart_items
	for i, item := range items {
		// Calculate discount percentage if discount exists
		var discountPercentage pgtype.Numeric
		discountPercentage = pgtype.Numeric{Valid: false}
		if item.DiscountAmount.Valid {
			// Get unit price and quantity as floats for calculation
			var unitPriceFloat, qtyFloat, discountFloat float64
			if val, err := item.UnitPrice.Value(); err == nil {
				if str, ok := val.(string); ok {
					unitPriceFloat, _ = strconv.ParseFloat(str, 64)
				}
			}
			if val, err := item.Quantity.Value(); err == nil {
				if str, ok := val.(string); ok {
					qtyFloat, _ = strconv.ParseFloat(str, 64)
				}
			}
			if val, err := item.DiscountAmount.Value(); err == nil {
				if str, ok := val.(string); ok {
					discountFloat, _ = strconv.ParseFloat(str, 64)
				}
			}

			if unitPriceFloat > 0 && qtyFloat > 0 {
				subtotalBeforeDiscount := unitPriceFloat * qtyFloat
				if subtotalBeforeDiscount > 0 {
					percent := (discountFloat / subtotalBeforeDiscount) * 100
					discountPercentage = pgtype.Numeric{}
					if err := discountPercentage.Scan(fmt.Sprintf("%.2f", percent)); err != nil {
						discountPercentage = pgtype.Numeric{Valid: false}
					} else {
						discountPercentage.Valid = true
					}
				}
			}
		}

		// Get tax rate from tax category if available
		var taxRate pgtype.Numeric
		taxRate = pgtype.Numeric{Valid: false}
		if item.TaxCategoryID.Valid {
			taxCategory, err := uc.repo.GetTaxCategory(ctx, item.TaxCategoryID.Int32)
			if err == nil {
				taxRate = taxCategory.TaxRate
			}
		}

		// Convert serial_number to serial_numbers array
		var serialNumbers []string
		if item.SerialNumber.Valid && item.SerialNumber.String != "" {
			serialNumbers = []string{item.SerialNumber.String}
		}

		// Prepare batch number
		var batchNumber pgtype.Text
		if item.BatchNumber.Valid {
			batchNumber = item.BatchNumber
		} else {
			batchNumber = pgtype.Text{Valid: false}
		}

		arg := repository.CreateSalesOrderLineV2Params{
			SalesOrderID:         orderID,
			OrganizationID:       cart.OrganizationID,
			LineNumber:           int32(i + 1),
			ProductID:            item.ProductID,
			ProductVariantID:     item.ProductVariantID,
			ProductName:          item.ProductName, // From ListCartItems JOIN
			ProductSku:           pgtype.Text{String: item.ProductSku, Valid: true},
			QuantityOrdered:      item.Quantity,
			UomID:                item.UomID,
			UnitPrice:            item.UnitPrice,
			DiscountAmount:       item.DiscountAmount,
			DiscountPercentage:   discountPercentage,
			TaxAmount:            item.TaxAmount,
			LineTotal:            item.LineTotal,
			TaxCategoryID:        item.TaxCategoryID,
			TaxRate:              taxRate,
			BatchNumber:          batchNumber,
			SerialNumbers:        serialNumbers,
			CustomizationDetails: item.CustomizationDetails,
			LineStatus:           pgtype.Text{String: "pending", Valid: true},
			Notes:                item.Notes,
			Metadata:             item.Metadata,
		}

		_, err = uc.repo.CreateSalesOrderLineV2(ctx, arg)
		if err != nil {
			// In a real app, we might use a transaction and roll back
			fmt.Printf("Warning: failed to create order line: %s\n", err.Error())
		}
	}

	// 5. If sales_channel is 'pos', sync to POS transaction and record payment when cart was settled
	if cart.Channel.Valid && cart.Channel.String == "pos" {
		posTxn, err := uc.repo.CreateTransactionFromOrder(ctx, orderID)
		if err != nil {
			fmt.Printf("Warning: failed to create POS transaction: %s\n", err.Error())
		} else {
			err = uc.repo.SyncTransactionLinesFromOrder(ctx, repository.SyncTransactionLinesFromOrderParams{
				TransactionID: posTxn.ID,
				SalesOrderID:  orderID,
			})
			if err != nil {
				fmt.Printf("Warning: failed to sync POS transaction lines: %s\n", err.Error())
			}
			// If the cart had payment method/gateway set (settled), insert pos_payment for reconciliation
			order, err := uc.repo.GetSalesOrderV2(ctx, orderID)
			if err == nil && order.TotalAmount.Valid {
				paymentMethod := "cash"
				if order.PaymentMethod.Valid && order.PaymentMethod.String != "" {
					paymentMethod = order.PaymentMethod.String
				}
				var gateway pgtype.Text
				if order.PaymentGateway.Valid {
					gateway = order.PaymentGateway
				}
				_ = uc.repo.AddPaymentToTransaction(ctx, repository.AddPaymentToTransactionParams{
					TransactionID:   posTxn.ID,
					PaymentMethod:   paymentMethod,
					PaymentGateway:  gateway,
					Amount:          order.TotalAmount,
					ReferenceNumber: pgtype.Text{},
					Metadata:        order.Metadata,
				})
				// Update drawer expected_balance for this sale
				_ = uc.repo.UpdateSessionExpectedBalance(ctx, repository.UpdateSessionExpectedBalanceParams{
					ID:              posTxn.CashierSessionID,
					ExpectedBalance: order.TotalAmount,
				})
			}
		}
	}

	return utils.NewResponse(utils.CodeCreated, "order created from cart", map[string]interface{}{
		"order_id":     orderID,
		"order_number": orderNumber,
	})
}
