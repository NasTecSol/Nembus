package handler

import (
	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CartHandler struct {
	useCase *usecase.CartUseCase
}

func NewCartHandler(uc *usecase.CartUseCase) *CartHandler {
	return &CartHandler{useCase: uc}
}

func (h *CartHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}
// GetCart handles GET /api/carts/:id
// @Summary      Get cart by ID
// @Description  Retrieve a specific cart and its items by cart ID
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cart ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id} [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	resp := h.useCase.GetCart(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// AddToCart handles POST /api/carts/:id/items
// @Summary      Add item to cart
// @Description  Add a product item to an existing cart
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string             true  "Bearer token"
// @Param        id            path      string             true  "Cart ID (UUID)"
// @Param        body          body      AddToCartRequest   true  "Cart item payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id}/items [post]
func (h *CartHandler) AddToCart(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	var req struct {
		OrganizationID int32   `json:"organization_id" binding:"required"`
		ProductID      int32   `json:"product_id" binding:"required"`
		Quantity       float64 `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.AddToCart(c.Request.Context(), id, req.OrganizationID, req.ProductID, req.Quantity)
	c.JSON(resp.StatusCode, resp)
}

// ConvertToOrder handles POST /api/carts/:id/checkout
// @Summary      Convert cart to order
// @Description  Convert an existing cart into a sales order
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cart ID (UUID)"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id}/checkout [post]
func (h *CartHandler) ConvertToOrder(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	resp := h.useCase.ConvertToOrder(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// CreateCart handles POST /api/carts
// @Summary      Create cart (full)
// @Description  Create a new cart with full payload (detailed fields)
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        body          body      CreateCartRequest true  "Full create cart payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts [post]
func (h *CartHandler) CreateCart(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	deviceInfo, err := bytesFromMap(req.DeviceInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid device_info", nil))
		return
	}
	shipAddr, err := bytesFromMap(req.ShippingAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid shipping_address", nil))
		return
	}
	billAddr, err := bytesFromMap(req.BillingAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid billing_address", nil))
		return
	}
	meta, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}
	expiresAt, err := timestampPtrFromRFC3339(req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid expires_at", nil))
		return
	}

	arg := repository.CreateCartParams{
		CartNumber:      req.CartNumber,
		OrganizationID:  req.OrganizationID,
		StoreID:         int4Ptr(req.StoreID),
		CustomerID:      int4Ptr(req.CustomerID),
		GuestIdentifier: textPtr(req.GuestIdentifier),
		GuestEmail:      textPtr(req.GuestEmail),
		GuestPhone:      textPtr(req.GuestPhone),
		CartStatus:      repository.CartStatus(req.CartStatus),
		CartType:        repository.CartType(req.CartType),
		Channel:         textPtr(req.Channel),
		DeviceInfo:      deviceInfo,
		CreatedByUserID: int4Ptr(req.CreatedByUserID),
		CashierID:       int4Ptr(req.CashierID),
		PosTerminalID:   int4Ptr(req.PosTerminalID),
		ShippingAddress: shipAddr,
		BillingAddress:  billAddr,
		ShippingMethod:  textPtr(req.ShippingMethod),
		CouponCode:      textPtr(req.CouponCode),
		DiscountCode:    textPtr(req.DiscountCode),
		ExpiresAt:       expiresAt,
		Notes:           textPtr(req.Notes),
		Metadata:        meta,
	}

	resp := h.useCase.CreateCart(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// CreateNewCart handles POST /api/carts/new
// @Summary      Create lightweight cart
// @Description  Create a new cart with minimal required fields (convenience endpoint)
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        cart          body      CreateNewCartRequest  true  "Lightweight create cart payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/new [post]
func (h *CartHandler) CreateNewCart(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)
	var req CreateNewCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	meta, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	storeID := int4Ptr(req.StoreID)
	customerID := int4Ptr(req.CustomerID)
	createdBy := int4Ptr(req.CreatedByUserID)
	cashierID := int4Ptr(req.CashierID)
	posID := int4Ptr(req.PosTerminalID)

	guestIdentifier := ""
	if req.GuestIdentifier != nil {
		guestIdentifier = *req.GuestIdentifier
	}
	guestEmail := ""
	if req.GuestEmail != nil {
		guestEmail = *req.GuestEmail
	}
	guestPhone := ""
	if req.GuestPhone != nil {
		guestPhone = *req.GuestPhone
	}

	notes := ""
	if req.Notes != nil {
		notes = *req.Notes
	}

	resp := h.useCase.CreateNewCart(c.Request.Context(), req.OrganizationID, storeID, customerID, guestIdentifier, guestEmail, guestPhone, createdBy, cashierID, posID, meta, notes)
	c.JSON(resp.StatusCode, resp)
}

// GetCartByNumber handles GET /api/carts/by-number/:cart_number
// @Summary      Get cart by cart number
// @Description  Retrieve a cart and its items by cart number
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        cart_number   path      string  true  "Cart number"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/by-number/{cart_number} [get]
func (h *CartHandler) GetCartByNumber(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cartNumber := c.Param("cart_number")
	if cartNumber == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "cart_number is required", nil))
		return
	}

	resp := h.useCase.GetCartByNumber(c.Request.Context(), cartNumber)
	c.JSON(resp.StatusCode, resp)
}

// GetActiveCartByCustomer handles GET /api/carts/by-customer?customer_id=&store_id=
// @Summary      Get active cart by customer
// @Description  Retrieve the active cart for a customer in a store
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        customer_id   query     int     true  "Customer ID"
// @Param        store_id      query     int     true  "Store ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/by-customer [get]
func (h *CartHandler) GetActiveCartByCustomer(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	customerIDStr := c.Query("customer_id")
	storeIDStr := c.Query("store_id")
	customerID64, err := strconv.ParseInt(customerIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid customer_id", nil))
		return
	}
	storeID64, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	arg := repository.GetCartByCustomerParams{
		CustomerID: pgtype.Int4{Int32: int32(customerID64), Valid: true},
		StoreID:    pgtype.Int4{Int32: int32(storeID64), Valid: true},
	}

	resp := h.useCase.GetActiveCartByCustomer(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// GetActiveCartByGuestIdentifier handles GET /api/carts/by-guest?guest_identifier=&store_id=
// @Summary      Get active cart by guest identifier
// @Description  Retrieve the active guest cart for a store
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true  "Tenant identifier"
// @Param        Authorization    header    string  true  "Bearer token"
// @Param        guest_identifier query     string  true  "Guest identifier"
// @Param        store_id         query     int     true  "Store ID"
// @Success      200              {object}  SuccessResponse
// @Failure      400              {object}  ErrorResponse
// @Failure      401              {object}  ErrorResponse
// @Failure      404              {object}  ErrorResponse
// @Failure      500              {object}  ErrorResponse
// @Router       /api/carts/by-guest [get]
func (h *CartHandler) GetActiveCartByGuestIdentifier(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	guestIdentifier := c.Query("guest_identifier")
	if guestIdentifier == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "guest_identifier is required", nil))
		return
	}
	storeIDStr := c.Query("store_id")
	storeID64, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	arg := repository.GetCartByGuestIdentifierParams{
		GuestIdentifier: pgtype.Text{String: guestIdentifier, Valid: true},
		StoreID:         pgtype.Int4{Int32: int32(storeID64), Valid: true},
	}

	resp := h.useCase.GetActiveCartByGuestIdentifier(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// ListActiveCarts handles GET /api/carts?store_id=&limit=&offset=
// @Summary      List active carts
// @Description  List active carts for a store with pagination
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      query     int     true  "Store ID"
// @Param        limit         query     int     false "Limit"
// @Param        offset        query     int     false "Offset"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts [get]
func (h *CartHandler) ListActiveCarts(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeIDStr := c.Query("store_id")
	storeID64, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	limit64, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 32)
	offset64, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)

	arg := repository.ListActiveCartsParams{
		StoreID: pgtype.Int4{Int32: int32(storeID64), Valid: true},
		Limit:   int32(limit64),
		Offset:  int32(offset64),
	}

	resp := h.useCase.ListActiveCarts(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// ListAbandonedCarts handles GET /api/carts/abandoned?store_id=&min_value=&limit=&offset=
// @Summary      List abandoned carts
// @Description  List abandoned carts filtered by store and minimum value
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      query     int     true  "Store ID"
// @Param        min_value     query     string  false "Minimum line total"
// @Param        limit         query     int     false "Limit"
// @Param        offset        query     int     false "Offset"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/abandoned [get]
func (h *CartHandler) ListAbandonedCarts(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeIDStr := c.Query("store_id")
	storeID64, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	minValue := c.DefaultQuery("min_value", "0")
	minNumeric, err := numericFromString(minValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid min_value", nil))
		return
	}

	limit64, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 32)
	offset64, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)

	arg := repository.ListAbandonedCartsParams{
		StoreID:   pgtype.Int4{Int32: int32(storeID64), Valid: true},
		LineTotal: minNumeric,
		Limit:     int32(limit64),
		Offset:    int32(offset64),
	}

	resp := h.useCase.ListAbandonedCarts(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdateCart handles PUT /api/carts/:id
// @Summary      Update cart totals and metadata
// @Description  Update cart monetary totals, shipping, notes and metadata
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cart ID (UUID)"
// @Param        body          body      UpdateCartRequest true  "Update cart payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id} [put]
func (h *CartHandler) UpdateCart(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	var req UpdateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	subtotal, err := numericFromString(req.Subtotal)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid subtotal", nil))
		return
	}
	discount, err := numericFromString(req.DiscountAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid discount_amount", nil))
		return
	}
	tax, err := numericFromString(req.TaxAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid tax_amount", nil))
		return
	}
	shipping, err := numericFromString(req.ShippingAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid shipping_amount", nil))
		return
	}
	total, err := numericFromString(req.TotalAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid total_amount", nil))
		return
	}

	promo := pgtype.Numeric{Valid: false}
	if req.PromotionalCreds != nil {
		promo, err = numericFromString(*req.PromotionalCreds)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid promotional_credits", nil))
			return
		}
	} else {
		promo, _ = numericFromString("0")
	}

	shipAddr, err := bytesFromMap(req.ShippingAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid shipping_address", nil))
		return
	}
	billAddr, err := bytesFromMap(req.BillingAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid billing_address", nil))
		return
	}
	meta, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	arg := repository.UpdateCartParams{
		ID:                 cartID,
		Subtotal:           subtotal,
		DiscountAmount:     discount,
		TaxAmount:          tax,
		ShippingAmount:     shipping,
		TotalAmount:        total,
		CouponCode:         textPtr(req.CouponCode),
		DiscountCode:       textPtr(req.DiscountCode),
		PromotionalCredits: promo,
		ShippingAddress:    shipAddr,
		BillingAddress:     billAddr,
		ShippingMethod:     textPtr(req.ShippingMethod),
		Notes:              textPtr(req.Notes),
		Metadata:           meta,
	}

	resp := h.useCase.UpdateCart(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdateCartStatus handles PUT /api/carts/:id/status
// @Summary      Update cart status
// @Description  Update cart status (e.g. active, converted) and optional conversion details
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cart ID (UUID)"
// @Param        body          body      UpdateCartStatusRequest true  "Status payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id}/status [put]
func (h *CartHandler) UpdateCartStatus(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	var req UpdateCartStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	convertedOrderID, err := uuidPtrFromString(req.ConvertedToOrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid converted_to_order_id", nil))
		return
	}
	convertedAt, err := timestampPtrFromRFC3339(req.ConvertedAtISO8601)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid converted_at", nil))
		return
	}

	arg := repository.UpdateCartStatusParams{
		ID:                 cartID,
		CartStatus:         repository.CartStatus(req.CartStatus),
		ConvertedToOrderID: convertedOrderID,
		ConvertedAt:        convertedAt,
	}

	resp := h.useCase.UpdateCartStatus(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdateCartCustomer handles PUT /api/carts/:id/customer
// @Summary      Update cart customer
// @Description  Associate a cart with a different customer
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cart ID (UUID)"
// @Param        body          body      UpdateCartCustomerRequest true  "Update customer payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id}/customer [put]
func (h *CartHandler) UpdateCartCustomer(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	var req UpdateCartCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	arg := repository.UpdateCartCustomerParams{
		ID:         cartID,
		CustomerID: pgtype.Int4{Int32: req.CustomerID, Valid: true},
	}

	resp := h.useCase.UpdateCartCustomer(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// DeleteCart handles DELETE /api/carts/:id
// @Summary      Delete cart
// @Description  Delete a cart and its items
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cart ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id} [delete]
func (h *CartHandler) DeleteCart(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	resp := h.useCase.DeleteCart(c.Request.Context(), cartID)
	c.JSON(resp.StatusCode, resp)
}

// ExpireAbandonedCarts handles POST /api/carts/expire?store_id=
// @Summary      Expire abandoned carts
// @Description  Expire abandoned carts for a given store
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      query     int     true  "Store ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/expire [post]
func (h *CartHandler) ExpireAbandonedCarts(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeIDStr := c.Query("store_id")
	storeID64, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	resp := h.useCase.ExpireAbandonedCarts(c.Request.Context(), pgtype.Int4{Int32: int32(storeID64), Valid: true})
	c.JSON(resp.StatusCode, resp)
}

// ListCartItems handles GET /api/carts/:id/items
// @Summary      List cart items
// @Description  List items for a cart
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cart ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id}/items [get]
func (h *CartHandler) ListCartItems(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	resp := h.useCase.ListCartItems(c.Request.Context(), cartID)
	c.JSON(resp.StatusCode, resp)
}

// CreateCartItemRaw handles POST /api/carts/:id/items/raw
// @Summary      Create cart item (raw)
// @Description  Create a cart item with full fields
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cart ID (UUID)"
// @Param        body          body      CreateCartItemRequest true  "Create cart item payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id}/items/raw [post]
func (h *CartHandler) CreateCartItemRaw(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	var req CreateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	qty, err := numericFromString(req.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid quantity", nil))
		return
	}
	unitPrice, err := numericFromString(req.UnitPrice)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid unit_price", nil))
		return
	}
	lineTotal, err := numericFromString(req.LineTotal)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid line_total", nil))
		return
	}

	discount := pgtype.Numeric{Valid: false}
	if req.DiscountAmount != nil {
		discount, err = numericFromString(*req.DiscountAmount)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid discount_amount", nil))
			return
		}
	} else {
		discount, _ = numericFromString("0")
	}

	tax := pgtype.Numeric{Valid: false}
	if req.TaxAmount != nil {
		tax, err = numericFromString(*req.TaxAmount)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid tax_amount", nil))
			return
		}
	} else {
		tax, _ = numericFromString("0")
	}

	custom, err := bytesFromMap(req.CustomizationDetails)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid customization_details", nil))
		return
	}
	meta, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	arg := repository.CreateCartItemParams{
		CartID:               cartID,
		OrganizationID:       req.OrganizationID,
		ProductID:            req.ProductID,
		ProductVariantID:     int4Ptr(req.ProductVariantID),
		Quantity:             qty,
		UomID:                int4Ptr(req.UomID),
		UnitPrice:            unitPrice,
		DiscountAmount:       discount,
		TaxAmount:            tax,
		LineTotal:            lineTotal,
		PriceListID:          int4Ptr(req.PriceListID),
		TaxCategoryID:        int4Ptr(req.TaxCategoryID),
		BatchNumber:          textPtr(req.BatchNumber),
		SerialNumber:         textPtr(req.SerialNumber),
		CustomizationDetails: custom,
		Notes:                textPtr(req.Notes),
		Metadata:             meta,
	}

	resp := h.useCase.CreateCartItem(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// GetCartItem handles GET /api/cart-items/:item_id
// @Summary      Get cart item by ID
// @Description  Retrieve a cart item by its UUID
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        item_id       path      string  true  "Cart item ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cart-items/{item_id} [get]
func (h *CartHandler) GetCartItem(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid item_id", nil))
		return
	}

	resp := h.useCase.GetCartItem(c.Request.Context(), itemID)
	c.JSON(resp.StatusCode, resp)
}

// GetCartItemByProduct handles GET /api/carts/:id/items/by-product?product_id=&product_variant_id=&batch_number=&serial_number=
// @Summary      Get cart item by product
// @Description  Find a cart item by product, variant, batch or serial
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id       header    string  true  "Tenant identifier"
// @Param        Authorization     header    string  true  "Bearer token"
// @Param        id                path      string  true  "Cart ID (UUID)"
// @Param        product_id        query     int     true  "Product ID"
// @Param        product_variant_id query    int     false "Product variant ID"
// @Param        batch_number      query     string  false "Batch number"
// @Param        serial_number     query     string  false "Serial number"
// @Success      200               {object}  SuccessResponse
// @Failure      400               {object}  ErrorResponse
// @Failure      401               {object}  ErrorResponse
// @Failure      404               {object}  ErrorResponse
// @Failure      500               {object}  ErrorResponse
// @Router       /api/carts/{id}/items/by-product [get]
func (h *CartHandler) GetCartItemByProduct(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	productIDStr := c.Query("product_id")
	productID64, err := strconv.ParseInt(productIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil))
		return
	}

	variantIDStr := c.Query("product_variant_id")
	var variantID *int32
	if variantIDStr != "" {
		v64, err := strconv.ParseInt(variantIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid product_variant_id", nil))
			return
		}
		v32 := int32(v64)
		variantID = &v32
	}

	batch := c.Query("batch_number")
	serial := c.Query("serial_number")

	arg := repository.GetCartItemByProductParams{
		CartID:           cartID,
		ProductID:        int32(productID64),
		ProductVariantID: int4Ptr(variantID),
		BatchNumber:      textPtr(&batch),
		SerialNumber:     textPtr(&serial),
	}

	resp := h.useCase.GetCartItemByProduct(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdateCartItem handles PUT /api/cart-items/:item_id
// @Summary      Update cart item
// @Description  Update quantity, pricing and metadata of a cart item
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        item_id       path      string  true  "Cart item ID (UUID)"
// @Param        body          body      UpdateCartItemRequest true  "Update cart item payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cart-items/{item_id} [put]
func (h *CartHandler) UpdateCartItem(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid item_id", nil))
		return
	}

	var req UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	qty, err := numericFromString(req.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid quantity", nil))
		return
	}
	unitPrice, err := numericFromString(req.UnitPrice)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid unit_price", nil))
		return
	}
	discount, err := numericFromString(req.DiscountAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid discount_amount", nil))
		return
	}
	tax, err := numericFromString(req.TaxAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid tax_amount", nil))
		return
	}
	lineTotal, err := numericFromString(req.LineTotal)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid line_total", nil))
		return
	}
	meta, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	arg := repository.UpdateCartItemParams{
		ID:             itemID,
		Quantity:       qty,
		UnitPrice:      unitPrice,
		DiscountAmount: discount,
		TaxAmount:      tax,
		LineTotal:      lineTotal,
		Notes:          textPtr(req.Notes),
		Metadata:       meta,
	}

	resp := h.useCase.UpdateCartItem(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdateCartItemQuantity handles PATCH /api/cart-items/:item_id/quantity
// @Summary      Update cart item quantity
// @Description  Adjust the quantity of a cart item by delta
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        item_id       path      string  true  "Cart item ID (UUID)"
// @Param        body          body      UpdateCartItemQuantityRequest true  "Delta quantity payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cart-items/{item_id}/quantity [patch]
func (h *CartHandler) UpdateCartItemQuantity(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid item_id", nil))
		return
	}

	var req UpdateCartItemQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	delta, err := numericFromString(req.DeltaQuantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid delta_quantity", nil))
		return
	}

	arg := repository.UpdateCartItemQuantityParams{
		ID:       itemID,
		Quantity: delta,
	}

	resp := h.useCase.UpdateCartItemQuantity(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// DeleteCartItem handles DELETE /api/cart-items/:item_id
// @Summary      Delete cart item
// @Description  Remove an item from a cart
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        item_id       path      string  true  "Cart item ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cart-items/{item_id} [delete]
func (h *CartHandler) DeleteCartItem(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid item_id", nil))
		return
	}

	resp := h.useCase.DeleteCartItem(c.Request.Context(), itemID)
	c.JSON(resp.StatusCode, resp)
}

// ClearCartItems handles DELETE /api/carts/:id/items
// @Summary      Clear cart items
// @Description  Remove all items from a cart
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cart ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id}/items [delete]
func (h *CartHandler) ClearCartItems(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	resp := h.useCase.ClearCartItems(c.Request.Context(), cartID)
	c.JSON(resp.StatusCode, resp)
}

// GetCartItemCount handles GET /api/carts/:id/items/count
// @Summary      Get cart item count
// @Description  Return number of items in a cart
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cart ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id}/items/count [get]
func (h *CartHandler) GetCartItemCount(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	resp := h.useCase.GetCartItemCount(c.Request.Context(), cartID)
	c.JSON(resp.StatusCode, resp)
}

// GetCartTotals handles GET /api/carts/:id/totals
// @Summary      Get cart totals
// @Description  Return subtotal, tax, shipping and total amounts for a cart
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cart ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id}/totals [get]
func (h *CartHandler) GetCartTotals(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	resp := h.useCase.GetCartTotals(c.Request.Context(), cartID)
	c.JSON(resp.StatusCode, resp)
}

// CreateCartActivity handles POST /api/carts/:id/activities
// @Summary      Create cart activity
// @Description  Log an activity for a cart
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cart ID (UUID)"
// @Param        body          body      CreateCartActivityRequest true  "Activity payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id}/activities [post]
func (h *CartHandler) CreateCartActivity(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	var req CreateCartActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	ipAddr, err := ipAddrPtr(req.IpAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid ip_address", nil))
		return
	}
	oldValue, err := bytesFromMap(req.OldValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid old_value", nil))
		return
	}
	newValue, err := bytesFromMap(req.NewValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid new_value", nil))
		return
	}

	arg := repository.CreateCartActivityParams{
		CartID:            cartID,
		OrganizationID:    req.OrganizationID,
		ActivityType:      req.ActivityType,
		Description:       textPtr(req.Description),
		PerformedByUserID: int4Ptr(req.PerformedByUserID),
		IpAddress:         ipAddr,
		UserAgent:         textPtr(req.UserAgent),
		OldValue:          oldValue,
		NewValue:          newValue,
	}

	resp := h.useCase.CreateCartActivity(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// ListCartActivities handles GET /api/carts/:id/activities?limit=&offset=
// @Summary      List cart activities
// @Description  List activity log entries for a cart
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cart ID (UUID)"
// @Param        limit         query     int     false "Limit (default 50)"
// @Param        offset        query     int     false "Offset (default 0)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id}/activities [get]
func (h *CartHandler) ListCartActivities(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	limit64, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 32)
	offset64, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)

	arg := repository.ListCartActivitiesParams{
		CartID: cartID,
		Limit:  int32(limit64),
		Offset: int32(offset64),
	}

	resp := h.useCase.ListCartActivities(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// ApplyCouponToCart handles POST /api/carts/:id/coupon
// @Summary      Apply coupon to cart
// @Description  Apply a coupon code and discount amount to a cart
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string             true  "Tenant identifier"
// @Param        Authorization header    string             true  "Bearer token"
// @Param        id            path      string             true  "Cart ID (UUID)"
// @Param        body          body      ApplyCouponRequest true  "Coupon payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id}/coupon [post]
func (h *CartHandler) ApplyCouponToCart(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	var req ApplyCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	discount, err := numericFromString(req.DiscountAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid discount_amount", nil))
		return
	}

	arg := repository.ApplyCouponToCartParams{
		ID:             cartID,
		CouponCode:     textPtr(&req.CouponCode),
		DiscountAmount: discount,
	}

	resp := h.useCase.ApplyCouponToCart(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// RecalculateCartTotals handles POST /api/carts/:id/recalculate
// @Summary      Recalculate cart totals
// @Description  Recalculate cart monetary totals based on current items
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cart ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id}/recalculate [post]
func (h *CartHandler) RecalculateCartTotals(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	resp := h.useCase.RecalculateCartTotals(c.Request.Context(), cartID)
	c.JSON(resp.StatusCode, resp)
}

// MergeGuestCartToCustomer handles POST /api/carts/:id/merge
// @Summary      Merge guest cart into customer cart
// @Description  Merge a guest cart (path id) into a target customer cart (body)
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string               true  "Tenant identifier"
// @Param        Authorization header    string               true  "Bearer token"
// @Param        id            path      string               true  "Guest cart ID (UUID)"
// @Param        body          body      MergeGuestCartRequest true  "Merge payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/carts/{id}/merge [post]
func (h *CartHandler) MergeGuestCartToCustomer(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	guestCartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart id", nil))
		return
	}

	var req MergeGuestCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	targetCartID, err := uuid.Parse(req.TargetCartID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid target_cart_id", nil))
		return
	}

	arg := repository.MergeGuestCartToCustomerParams{
		CartID:   guestCartID,
		CartID_2: targetCartID,
	}

	resp := h.useCase.MergeGuestCartToCustomer(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}
