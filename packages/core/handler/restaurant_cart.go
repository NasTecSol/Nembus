package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *RestaurantHandler) getCartUseCase(c *gin.Context) *usecase.RestaurantCartUseCase {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return nil
	}
	uc := usecase.NewRestaurantCartUseCase()
	uc.SetRepository(repo)
	if h.cartPool != nil {
		uc.SetPool(h.cartPool)
	}
	return uc
}

func resolveCreatedByUserIDFromContext(c *gin.Context, provided *int32) *int32 {
	if provided != nil && *provided > 0 {
		return provided
	}
	if uidStr, ok := middleware.GetUserIDFromContext(c); ok && strings.TrimSpace(uidStr) != "" {
		if parsed, err := strconv.ParseInt(strings.TrimSpace(uidStr), 10, 32); err == nil && parsed > 0 {
			v := int32(parsed)
			return &v
		}
	}
	return provided
}

func toUseCaseLineItems(items []RestaurantLineItemRequest) []usecase.RestaurantLineItemInput {
	out := make([]usecase.RestaurantLineItemInput, 0, len(items))
	for _, it := range items {
		out = append(out, usecase.RestaurantLineItemInput{
			MenuItemID:           it.MenuItemID,
			Quantity:             it.Quantity,
			UnitPrice:            it.UnitPrice,
			ModifierIDs:          it.ModifierIDs,
			Modifiers:            it.Modifiers,
			ComboComponentIDs:    it.ComboComponentIDs,
			ComboComponents:      it.ComboComponents,
			DiscountAmount:       it.DiscountAmount,
			CustomizationDetails: it.CustomizationDetails,
			Notes:                it.Notes,
			Status:               it.Status,
			Metadata:             it.Metadata,
		})
	}
	return out
}

func parseOptionalFloatString(s string) float64 {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 0
	}
	f, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0
	}
	return f
}

// === Restaurant Cart Handlers ===

// CreateNewRestaurantCart handles POST /api/restaurant/carts/new and POST /api/restaurant/carts
// @Summary      Create restaurant cart
// @Description  Creates a new restaurant cart with optional initial items and coupon code.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string                       true  "Tenant identifier"
// @Param        body         body      RestaurantCreateCartRequest  true  "Restaurant cart payload"
// @Success      201          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/carts/new [post]
// @Router       /api/restaurant/carts [post]
func (h *RestaurantHandler) CreateNewRestaurantCart(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	var req RestaurantCreateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := uc.CreateCart(c.Request.Context(), usecase.RestaurantCreateCartInput{
		CartNumber:      req.CartNumber,
		OrganizationID:  req.OrganizationID,
		StoreID:         req.StoreID,
		TableID:         req.TableID,
		OrderType:       req.OrderType,
		CustomerID:      req.CustomerID,
		GuestIdentifier: req.GuestIdentifier,
		GuestName:       req.GuestName,
		GuestEmail:      req.GuestEmail,
		GuestPhone:      req.GuestPhone,
		CartStatus:      req.CartStatus,
		Channel:         req.Channel,
		PaymentMethod:   req.PaymentMethod,
		PaymentGateway:  req.PaymentGateway,
		DeviceInfo:      req.DeviceInfo,
		CreatedByUserID: resolveCreatedByUserIDFromContext(c, req.CreatedByUserID),
		CashierID:       req.CashierID,
		PosTerminalID:   req.PosTerminalID,
		KioskSessionID:  req.KioskSessionID,
		CouponCode:      req.CouponCode,
		DiscountCode:    req.DiscountCode,
		DeliveryAddress: req.DeliveryAddress,
		ExpiresAt:       req.ExpiresAt,
		Notes:           req.Notes,
		Metadata:        req.Metadata,
		Items:           toUseCaseLineItems(req.Items),
	})
	c.JSON(resp.StatusCode, resp)
}

// GetRestaurantCart handles GET /api/restaurant/carts/:id
// @Summary      Get restaurant cart
// @Description  Returns a restaurant cart and its items by cart UUID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        id           path      string  true  "Cart UUID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Router       /api/restaurant/carts/{id} [get]
func (h *RestaurantHandler) GetRestaurantCart(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	resp := uc.GetCart(c.Request.Context(), cartID)
	c.JSON(resp.StatusCode, resp)
}

// GetActiveRestaurantCart handles GET /api/restaurant/carts?store_id=...&(customer_id|guest_identifier|table_id)=...
// @Summary      Get active restaurant cart
// @Description  Looks up an active restaurant cart by store_id and customer_id, guest_identifier, or table_id.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id       header    string  true   "Tenant identifier"
// @Param        store_id          query     int     true   "Store ID"
// @Param        customer_id       query     int     false  "Customer ID"
// @Param        table_id          query     int     false  "Table ID"
// @Param        guest_identifier  query     string  false  "Guest identifier"
// @Success      200               {object}  SuccessResponse
// @Failure      400               {object}  ErrorResponse
// @Failure      401               {object}  ErrorResponse
// @Failure      404               {object}  ErrorResponse
// @Router       /api/restaurant/carts [get]
func (h *RestaurantHandler) GetActiveRestaurantCart(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	storeID64, err := strconv.ParseInt(c.Query("store_id"), 10, 32)
	if err != nil || storeID64 <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "valid store_id query parameter is required", nil))
		return
	}

	var customerID *int32
	if cidStr := strings.TrimSpace(c.Query("customer_id")); cidStr != "" {
		if cid, err := strconv.ParseInt(cidStr, 10, 32); err == nil && cid > 0 {
			v := int32(cid)
			customerID = &v
		}
	}

	var tableID *int32
	if tidStr := strings.TrimSpace(c.Query("table_id")); tidStr != "" {
		if tid, err := strconv.ParseInt(tidStr, 10, 32); err == nil && tid > 0 {
			v := int32(tid)
			tableID = &v
		}
	}

	guestID := c.Query("guest_identifier")
	resp := uc.GetActiveCart(c.Request.Context(), int32(storeID64), customerID, guestID, tableID)
	c.JSON(resp.StatusCode, resp)
}

// ListActiveRestaurantCarts handles GET /api/restaurant/carts/active?store_id=...
// @Summary      List active restaurant carts
// @Description  Lists active restaurant carts for a store.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        store_id     query     int     true   "Store ID"
// @Param        limit        query     int     false  "Page limit (default 50)"
// @Param        offset       query     int     false  "Page offset (default 0)"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /api/restaurant/carts/active [get]
func (h *RestaurantHandler) ListActiveRestaurantCarts(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	storeID64, err := strconv.ParseInt(c.Query("store_id"), 10, 32)
	if err != nil || storeID64 <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "valid store_id query parameter is required", nil))
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	resp := uc.ListActiveCarts(c.Request.Context(), int32(storeID64), int32(limit), int32(offset))
	c.JSON(resp.StatusCode, resp)
}

// ListDraftRestaurantCarts handles GET /api/restaurant/carts/drafts?cashier_id=...&store_id=...
// @Summary      List draft/held restaurant carts by cashier
// @Description  Lists held/draft restaurant carts assigned to a cashier.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        cashier_id   query     int     true   "Cashier ID"
// @Param        store_id     query     int     false  "Store ID"
// @Param        limit        query     int     false  "Page limit (default 50)"
// @Param        offset       query     int     false  "Page offset (default 0)"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /api/restaurant/carts/drafts [get]
func (h *RestaurantHandler) ListDraftRestaurantCarts(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	cashierID64, err := strconv.ParseInt(c.Query("cashier_id"), 10, 32)
	if err != nil || cashierID64 <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "valid cashier_id query parameter is required", nil))
		return
	}
	var storeID *int32
	if sStr := strings.TrimSpace(c.Query("store_id")); sStr != "" {
		if sVal, err := strconv.ParseInt(sStr, 10, 32); err == nil && sVal > 0 {
			v := int32(sVal)
			storeID = &v
		}
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	resp := uc.ListDraftCartsByCashier(c.Request.Context(), int32(cashierID64), storeID, int32(limit), int32(offset))
	c.JSON(resp.StatusCode, resp)
}

// UpdateRestaurantCart handles PUT /api/restaurant/carts/:id
// @Summary      Update restaurant cart
// @Description  Updates restaurant cart metadata, customer, table, channel, or notes.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string                       true  "Tenant identifier"
// @Param        id           path      string                       true  "Cart UUID"
// @Param        body         body      RestaurantUpdateCartRequest  true  "Cart update payload"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Router       /api/restaurant/carts/{id} [put]
func (h *RestaurantHandler) UpdateRestaurantCart(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	var req RestaurantUpdateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	optText := func(s string) pgtype.Text {
		if strings.TrimSpace(s) == "" {
			return pgtype.Text{Valid: false}
		}
		return pgtype.Text{String: s, Valid: true}
	}

	resp := uc.UpdateCart(c.Request.Context(), repository.UpdateRestaurantCartParams{
		ID:              cartID,
		TableID:         utils.Int32ToPgInt4(req.TableID),
		OrderType:       optText(req.OrderType),
		CustomerID:      utils.Int32ToPgInt4(req.CustomerID),
		GuestIdentifier: optText(req.GuestIdentifier),
		GuestName:       optText(req.GuestName),
		GuestEmail:      optText(req.GuestEmail),
		GuestPhone:      optText(req.GuestPhone),
		Channel:         optText(req.Channel),
		PaymentMethod:   optText(req.PaymentMethod),
		PaymentGateway:  optText(req.PaymentGateway),
		CouponCode:      optText(req.CouponCode),
		DiscountCode:    optText(req.DiscountCode),
		DeliveryAddress: req.DeliveryAddress,
		Notes:           optText(req.Notes),
		Metadata:        req.Metadata,
	})
	c.JSON(resp.StatusCode, resp)
}

// AssignRestaurantCartTable handles PATCH /api/restaurant/carts/:id/table
// @Summary      Assign table to restaurant cart
// @Description  Assigns or clears a restaurant table on a cart.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string                        true  "Tenant identifier"
// @Param        id           path      string                        true  "Cart UUID"
// @Param        body         body      RestaurantAssignTableRequest  true  "Table assignment payload"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /api/restaurant/carts/{id}/table [patch]
func (h *RestaurantHandler) AssignRestaurantCartTable(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	var req RestaurantAssignTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := uc.AssignCartTable(c.Request.Context(), cartID, req.TableID)
	c.JSON(resp.StatusCode, resp)
}

// HoldRestaurantCart handles POST /api/restaurant/carts/:id/hold
// @Summary      Hold restaurant cart
// @Description  Marks an active restaurant cart as held/draft.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        id           path      string  true  "Cart UUID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /api/restaurant/carts/{id}/hold [post]
func (h *RestaurantHandler) HoldRestaurantCart(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	resp := uc.HoldCart(c.Request.Context(), cartID)
	c.JSON(resp.StatusCode, resp)
}

// DeleteRestaurantCart handles DELETE /api/restaurant/carts/:id
// @Summary      Delete restaurant cart
// @Description  Deletes a restaurant cart by UUID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        id           path      string  true  "Cart UUID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /api/restaurant/carts/{id} [delete]
func (h *RestaurantHandler) DeleteRestaurantCart(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	resp := uc.DeleteCart(c.Request.Context(), cartID)
	c.JSON(resp.StatusCode, resp)
}

// AddToRestaurantCart handles POST /api/restaurant/carts/add and POST /api/restaurant/carts/:id/items
// @Summary      Add menu item to restaurant cart
// @Description  Adds a menu item (with modifiers/combo components) to an existing cart or creates one automatically.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string                      true   "Tenant identifier"
// @Param        id           path      string                      false  "Cart UUID (optional when using /carts/add)"
// @Param        body         body      RestaurantAddToCartRequest  true   "Add to cart payload"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /api/restaurant/carts/add [post]
// @Router       /api/restaurant/carts/{id}/items [post]
func (h *RestaurantHandler) AddToRestaurantCart(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	var req RestaurantAddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	var cartUUID *uuid.UUID
	rawCartID := strings.TrimSpace(c.Param("id"))
	if rawCartID == "" {
		rawCartID = strings.TrimSpace(req.CartID)
	}
	if rawCartID != "" {
		parsed, err := uuid.Parse(rawCartID)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart_id", nil))
			return
		}
		cartUUID = &parsed
	}

	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}

	resp := uc.AddToCart(c.Request.Context(), usecase.RestaurantAddToCartInput{
		CartID:               cartUUID,
		OrganizationID:       req.OrganizationID,
		StoreID:              req.StoreID,
		TableID:              req.TableID,
		OrderType:            req.OrderType,
		CustomerID:           req.CustomerID,
		GuestIdentifier:      req.GuestIdentifier,
		GuestName:            req.GuestName,
		GuestEmail:           req.GuestEmail,
		GuestPhone:           req.GuestPhone,
		Channel:              req.Channel,
		CashierID:            req.CashierID,
		PosTerminalID:        req.PosTerminalID,
		KioskSessionID:       req.KioskSessionID,
		CreatedByUserID:      resolveCreatedByUserIDFromContext(c, req.CreatedByUserID),
		MenuItemID:           req.MenuItemID,
		Quantity:             qty,
		UnitPrice:            req.UnitPrice,
		ModifierIDs:          req.ModifierIDs,
		Modifiers:            req.Modifiers,
		ComboComponentIDs:    req.ComboComponentIDs,
		ComboComponents:      req.ComboComponents,
		DiscountAmount:       req.DiscountAmount,
		CustomizationDetails: req.CustomizationDetails,
		Notes:                req.Notes,
		Metadata:             req.Metadata,
	})
	c.JSON(resp.StatusCode, resp)
}

// UpdateRestaurantCartItem handles PUT /api/restaurant/cart-items/:item_id
// @Summary      Update restaurant cart item
// @Description  Updates quantity, modifiers, combo components, discount, or notes on a restaurant cart item.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string                           true  "Tenant identifier"
// @Param        item_id      path      string                           true  "Cart Item UUID"
// @Param        body         body      RestaurantUpdateCartItemRequest  true  "Cart item update payload"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /api/restaurant/cart-items/{item_id} [put]
func (h *RestaurantHandler) UpdateRestaurantCartItem(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid item_id", nil))
		return
	}

	var req RestaurantUpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := uc.UpdateCartItemQuantity(c.Request.Context(), usecase.RestaurantUpdateCartItemInput{
		ItemID:               itemID,
		Quantity:             req.Quantity,
		ModifierIDs:          req.ModifierIDs,
		Modifiers:            req.Modifiers,
		ComboComponentIDs:    req.ComboComponentIDs,
		ComboComponents:      req.ComboComponents,
		HasModifiersUpdate:   req.ModifierIDs != nil || req.Modifiers != nil,
		HasCombosUpdate:      req.ComboComponentIDs != nil || req.ComboComponents != nil,
		DiscountAmount:       req.DiscountAmount,
		CustomizationDetails: req.CustomizationDetails,
		Notes:                req.Notes,
		Metadata:             req.Metadata,
	})
	c.JSON(resp.StatusCode, resp)
}

// RemoveRestaurantCartItem handles DELETE /api/restaurant/cart-items/:item_id
// @Summary      Remove restaurant cart item
// @Description  Removes an item from a restaurant cart and recalculates cart totals.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        item_id      path      string  true  "Cart Item UUID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /api/restaurant/cart-items/{item_id} [delete]
func (h *RestaurantHandler) RemoveRestaurantCartItem(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid item_id", nil))
		return
	}

	resp := uc.RemoveCartItem(c.Request.Context(), itemID)
	c.JSON(resp.StatusCode, resp)
}

// ClearRestaurantCart handles DELETE /api/restaurant/carts/:id/items
// @Summary      Clear restaurant cart items
// @Description  Removes all items from a restaurant cart.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        id           path      string  true  "Cart UUID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /api/restaurant/carts/{id}/items [delete]
func (h *RestaurantHandler) ClearRestaurantCart(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	resp := uc.ClearCart(c.Request.Context(), cartID)
	c.JSON(resp.StatusCode, resp)
}

// MergeGuestRestaurantCart handles POST /api/restaurant/carts/merge
// @Summary      Merge guest restaurant cart into customer cart
// @Description  Merges a guest restaurant cart into the customer's active cart for the store.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string                      true  "Tenant identifier"
// @Param        body         body      RestaurantMergeCartRequest  true  "Merge cart payload"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /api/restaurant/carts/merge [post]
func (h *RestaurantHandler) MergeGuestRestaurantCart(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	var req RestaurantMergeCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := uc.MergeGuestCart(c.Request.Context(), req.StoreID, req.GuestIdentifier, req.CustomerID)
	c.JSON(resp.StatusCode, resp)
}

// ConvertRestaurantCartToOrder handles POST /api/restaurant/carts/:id/checkout
// @Summary      Checkout restaurant cart into order
// @Description  Converts a restaurant cart into a restaurant_orders row, applies coupon if provided, logs initial status history, deducts recipe ingredients, and creates a POS transaction when paid/pos channel.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string                     true   "Tenant identifier"
// @Param        id           path      string                     true   "Cart UUID"
// @Param        body         body      RestaurantCheckoutRequest  false  "Checkout payload"
// @Success      201          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Router       /api/restaurant/carts/{id}/checkout [post]
func (h *RestaurantHandler) ConvertRestaurantCartToOrder(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	var req RestaurantCheckoutRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
			return
		}
	}

	resp := uc.ConvertToOrder(c.Request.Context(), usecase.RestaurantCheckoutInput{
		CartID:           cartID,
		OrderNumber:      req.OrderNumber,
		OrderSource:      req.OrderSource,
		OrderType:        req.OrderType,
		TableID:          req.TableID,
		CashierID:        req.CashierID,
		CashierSessionID: req.CashierSessionID,
		CustomerID:       req.CustomerID,
		PosTerminalID:    req.PosTerminalID,
		KioskSessionID:   req.KioskSessionID,
		CreatedByUserID:  resolveCreatedByUserIDFromContext(c, req.CreatedByUserID),
		CustomerName:     req.CustomerName,
		CustomerEmail:    req.CustomerEmail,
		CustomerPhone:    req.CustomerPhone,
		DeliveryAddress:  req.DeliveryAddress,
		PaymentStatus:    req.PaymentStatus,
		PaymentMethod:    req.PaymentMethod,
		PaymentGateway:   req.PaymentGateway,
		ReferenceNumber:  req.ReferenceNumber,
		AmountPaid:       req.AmountPaid,
		PosTransactionID: req.PosTransactionID,
		CouponCode:       req.CouponCode,
		DiscountCode:     req.DiscountCode,
		Notes:            req.Notes,
		Status:           req.Status,
		Metadata:         req.Metadata,
	})
	c.JSON(resp.StatusCode, resp)
}

// ReopenRestaurantCart handles POST /api/restaurant/carts/:id/reopen
// @Summary      Reopen restaurant cart
// @Description  Reopens a held or abandoned restaurant cart back to active status.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        id           path      string  true  "Cart UUID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /api/restaurant/carts/{id}/reopen [post]
func (h *RestaurantHandler) ReopenRestaurantCart(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	resp := uc.ReopenCartByCartID(c.Request.Context(), cartID)
	c.JSON(resp.StatusCode, resp)
}

// ReopenRestaurantOrder handles POST /api/restaurant/orders/:order_id/reopen
// @Summary      Reopen unpaid restaurant order to cart
// @Description  Reverses recipe stock deduction, voids the unpaid order with status history, and returns an active cart.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        order_id     path      int     true  "Order ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Router       /api/restaurant/orders/{order_id}/reopen [post]
func (h *RestaurantHandler) ReopenRestaurantOrder(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	orderID, err := strconv.ParseInt(c.Param("order_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order_id", nil))
		return
	}

	resp := uc.ReopenOrderToCart(c.Request.Context(), int32(orderID))
	c.JSON(resp.StatusCode, resp)
}

// AddRestaurantOrderItem handles POST /api/restaurant/orders/:order_id/items
// @Summary      Add item to existing restaurant order
// @Description  Appends a new menu item to an unpaid restaurant order, recalculates order totals, and deducts recipe ingredients.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string                     true  "Tenant identifier"
// @Param        order_id     path      int                        true  "Order ID"
// @Param        body         body      RestaurantLineItemRequest  true  "Order line item payload"
// @Success      201          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /api/restaurant/orders/{order_id}/items [post]
func (h *RestaurantHandler) AddRestaurantOrderItem(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	orderID, err := strconv.ParseInt(c.Param("order_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order_id", nil))
		return
	}

	var req RestaurantLineItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := uc.AddOrderItem(c.Request.Context(), int32(orderID), usecase.RestaurantLineItemInput{
		MenuItemID:           req.MenuItemID,
		Quantity:             req.Quantity,
		UnitPrice:            req.UnitPrice,
		ModifierIDs:          req.ModifierIDs,
		Modifiers:            req.Modifiers,
		ComboComponentIDs:    req.ComboComponentIDs,
		ComboComponents:      req.ComboComponents,
		DiscountAmount:       req.DiscountAmount,
		CustomizationDetails: req.CustomizationDetails,
		Notes:                req.Notes,
		Status:               req.Status,
		Metadata:             req.Metadata,
	})
	c.JSON(resp.StatusCode, resp)
}

// UpdateRestaurantOrderPaymentStatus handles PATCH /api/restaurant/orders/:order_id/payment-status
// @Summary      Update restaurant order payment status
// @Description  Updates payment status, amount paid, and change given on a restaurant order.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string                                     true  "Tenant identifier"
// @Param        order_id     path      int                                        true  "Order ID"
// @Param        body         body      UpdateRestaurantOrderPaymentStatusRequest  true  "Payment status payload"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /api/restaurant/orders/{order_id}/payment-status [patch]
func (h *RestaurantHandler) UpdateRestaurantOrderPaymentStatus(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	orderID, err := strconv.ParseInt(c.Param("order_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order_id", nil))
		return
	}

	var req UpdateRestaurantOrderPaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := uc.UpdateOrderPaymentStatus(c.Request.Context(), int32(orderID), req.PaymentStatus, req.AmountPaid, req.ChangeGiven)
	c.JSON(resp.StatusCode, resp)
}

// ApplyRestaurantCoupon handles POST /api/restaurant/carts/:id/coupon
// @Summary      Apply coupon code to restaurant cart
// @Description  Validates a restaurant promotion coupon code, applies prorated discounts to eligible cart items, and recalculates VAT and cart totals.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string                        true  "Tenant identifier"
// @Param        id           path      string                        true  "Cart UUID"
// @Param        body         body      ApplyRestaurantCouponRequest  true  "Coupon code payload"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Router       /api/restaurant/carts/{id}/coupon [post]
func (h *RestaurantHandler) ApplyRestaurantCoupon(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	var req ApplyRestaurantCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := uc.ApplyCouponToCart(c.Request.Context(), cartID, req.CouponCode)
	c.JSON(resp.StatusCode, resp)
}

// RemoveRestaurantCoupon handles DELETE /api/restaurant/carts/:id/coupon
// @Summary      Remove coupon code from restaurant cart
// @Description  Clears the applied restaurant coupon discount from cart items and recalculates VAT and cart totals.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        id           path      string  true  "Cart UUID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Router       /api/restaurant/carts/{id}/coupon [delete]
func (h *RestaurantHandler) RemoveRestaurantCoupon(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	resp := uc.RemoveCouponFromCart(c.Request.Context(), cartID)
	c.JSON(resp.StatusCode, resp)
}

// GetRestaurantOrderStatusHistory handles GET /api/restaurant/orders/:order_id/status-history
// @Summary      Get restaurant order status history
// @Description  Returns the chronological status transition audit trail for a restaurant order.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        order_id     path      int     true  "Order ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Router       /api/restaurant/orders/{order_id}/status-history [get]
func (h *RestaurantHandler) GetRestaurantOrderStatusHistory(c *gin.Context) {
	uc := h.getCartUseCase(c)
	if uc == nil {
		return
	}

	orderID, err := strconv.ParseInt(c.Param("order_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order_id", nil))
		return
	}

	resp := uc.GetOrderStatusHistory(c.Request.Context(), int32(orderID))
	c.JSON(resp.StatusCode, resp)
}
