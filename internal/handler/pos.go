package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// PosHandler holds the POS use case.
type PosHandler struct {
	useCase *usecase.PosUseCase
}

// NewPosHandler creates a new POS handler.
func NewPosHandler(uc *usecase.PosUseCase) *PosHandler {
	return &PosHandler{useCase: uc}
}

func (h *PosHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// ListProducts handles GET /api/pos/stores/:store_id/products
// @Summary      List POS products for store
// @Description  Returns products with stock for a store (categories, prices, barcode). Optional filters: category_id, search_term, include_out_of_stock.
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id           header    string  true   "Tenant identifier"
// @Param        Authorization         header    string  true   "Bearer token"
// @Param        store_id              path      int     true   "Store ID"
// @Param        category_id           query     int     false  "Filter by category ID"
// @Param        search_term           query     string  false  "Filter by name, SKU, or barcode"
// @Param        include_out_of_stock  query     bool    false  "Include out-of-stock products (default false)"
// @Success      200                   {object}  SuccessResponse
// @Failure      400                   {object}  ErrorResponse
// @Failure      401                   {object}  ErrorResponse
// @Failure      404                   {object}  ErrorResponse
// @Failure      500                   {object}  ErrorResponse
// @Router       /api/pos/stores/{store_id}/products [get]
func (h *PosHandler) ListProducts(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeID, err := strconv.ParseInt(c.Param("store_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	var categoryID *int32
	if s := c.Query("category_id"); s != "" {
		id, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid category_id", nil))
			return
		}
		catID := int32(id)
		categoryID = &catID
	}
	var searchTerm *string
	if s := c.Query("search_term"); s != "" {
		searchTerm = &s
	}
	includeOutOfStock := c.Query("include_out_of_stock") == "true" || c.Query("include_out_of_stock") == "1"

	resp := h.useCase.ListProductsForStore(c.Request.Context(), int32(storeID), categoryID, searchTerm, includeOutOfStock)
	c.JSON(resp.StatusCode, resp)
}

// GetProductsByCategory handles GET /api/pos/stores/:store_id/products/category/:category_id
// @Summary      Get POS products by category
// @Description  Returns products in a category (and optionally subcategories) for a store, with stock and pricing.
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id            header    string  true   "Tenant identifier"
// @Param        Authorization          header    string  true   "Bearer token"
// @Param        store_id               path      int     true   "Store ID"
// @Param        category_id            path      int     true   "Category ID"
// @Param        include_subcategories  query     bool    false  "Include subcategories (default true)"
// @Success      200                    {object}  SuccessResponse
// @Failure      400                    {object}  ErrorResponse
// @Failure      401                    {object}  ErrorResponse
// @Failure      404                    {object}  ErrorResponse
// @Failure      500                    {object}  ErrorResponse
// @Router       /api/pos/stores/{store_id}/products/category/{category_id} [get]
func (h *PosHandler) GetProductsByCategory(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeID, err := strconv.ParseInt(c.Param("store_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}
	categoryID, err := strconv.ParseInt(c.Param("category_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid category_id", nil))
		return
	}
	includeSubcategories := c.Query("include_subcategories") != "false" && c.Query("include_subcategories") != "0"

	resp := h.useCase.GetProductsByCategory(c.Request.Context(), int32(storeID), int32(categoryID), includeSubcategories)
	c.JSON(resp.StatusCode, resp)
}

// SearchProduct handles GET /api/pos/stores/:store_id/products/search
// @Summary      Search POS product by barcode, ID, or name
// @Description  Searches by barcode (exact), product ID (exact), or name/SKU (fuzzy). Returns single product or list of matches.
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        store_id      path      int     true   "Store ID"
// @Param        q             query     string  true   "Search term (barcode, product ID, or name/SKU)"
// @Param        limit         query     int     false  "Max results (default 50)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/stores/{store_id}/products/search [get]
func (h *PosHandler) SearchProduct(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeID, err := strconv.ParseInt(c.Param("store_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "query parameter 'q' required", nil))
		return
	}
	limit := int32(50)
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 32); err == nil && n > 0 {
			limit = int32(n)
		}
	}

	resp := h.useCase.SearchProduct(c.Request.Context(), int32(storeID), q, limit)
	c.JSON(resp.StatusCode, resp)
}

// GetCategories handles GET /api/pos/categories
// @Summary      Get POS categories
// @Description  Returns POS categories with product counts
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/categories [get]
func (h *PosHandler) GetCategories(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.GetCategories(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// AddProduct handles POST /api/pos/products
// @Summary      Add POS product
// @Description  Creates a product with optional barcode and retail price
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        body          body      AddProductRequest true "Product payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/products [post]
func (h *PosHandler) AddProduct(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req AddProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	input := &usecase.PosAddProductInput{
		OrganizationID:       req.OrganizationID,
		SKU:                  req.SKU,
		Name:                 req.Name,
		Description:          req.Description,
		CategoryID:           req.CategoryID,
		BrandID:              req.BrandID,
		BaseUomID:            req.BaseUomID,
		ProductType:          req.ProductType,
		TaxCategoryID:        req.TaxCategoryID,
		IsSerialized:         req.IsSerialized,
		IsBatchManaged:       req.IsBatchManaged,
		IsActive:             req.IsActive,
		IsSellable:           req.IsSellable,
		IsPurchasable:        req.IsPurchasable,
		AllowDecimalQuantity: req.AllowDecimalQuantity,
		TrackInventory:       req.TrackInventory,
		Barcode:              req.Barcode,
		RetailPrice:          req.RetailPrice,
	}

	resp := h.useCase.AddProduct(c.Request.Context(), input)
	c.JSON(resp.StatusCode, resp)
}

// AddPaymentToTransactionRequest captures payment from terminal/mobile clients.
// Metadata can hold gateway response: gateway_txn_id, masked_card, auth_code for auditing.
type AddPaymentToTransactionRequest struct {
	TransactionID   int32       `json:"transaction_id"`
	PaymentMethod   string      `json:"payment_method"`
	PaymentGateway  string      `json:"payment_gateway"`
	Amount          string      `json:"amount"`
	ReferenceNumber string      `json:"reference_number"`
	Metadata        interface{} `json:"metadata"` // Gateway payload: e.g. {"gateway_txn_id":"...","masked_card":"****1234","auth_code":"ABC123"}
}

// ProcessPayment handles POST /api/pos/payments
// @Summary      Process a POS payment
// @Description  Records a payment for a transaction and updates drawer balance
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        body          body      AddPaymentToTransactionRequest true "Payment payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/payments [post]
func (h *PosHandler) ProcessPayment(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req AddPaymentToTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	amount, err := repo.ParseNumeric(c.Request.Context(), req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid amount format", nil))
		return
	}

	metadataJSON, _ := json.Marshal(req.Metadata)
	if metadataJSON == nil {
		metadataJSON = []byte("{}")
	}
	arg := repository.AddPaymentToTransactionParams{
		TransactionID:   req.TransactionID,
		PaymentMethod:   req.PaymentMethod,
		PaymentGateway:  pgtype.Text{String: req.PaymentGateway, Valid: req.PaymentGateway != ""},
		Amount:          amount,
		ReferenceNumber: pgtype.Text{String: req.ReferenceNumber, Valid: req.ReferenceNumber != ""},
		Metadata:         metadataJSON,
	}

	resp := h.useCase.ProcessPOSPayment(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// ListTodayTransactions handles GET /api/pos/stores/:store_id/transactions
// @Summary      List today's POS transactions
// @Description  Returns today's completed POS transactions for a store
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      path      int     true  "Store ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/stores/{store_id}/transactions [get]
func (h *PosHandler) ListTodayTransactions(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)
	storeID, err := strconv.ParseInt(c.Param("store_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}
	resp := h.useCase.ListTodaysTransactions(c.Request.Context(), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// GetTransaction handles GET /api/pos/transactions/:id
// @Summary      Get POS transaction
// @Description  Returns a single POS transaction by ID
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Transaction ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/transactions/{id} [get]
func (h *PosHandler) GetTransaction(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid transaction id", nil))
		return
	}
	resp := h.useCase.GetTransaction(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// GetTransactionFull handles GET /api/pos/transactions/:id/full
// @Summary      Get POS transaction with lines
// @Description  Returns a POS transaction with full line details (products, quantities, prices)
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Transaction ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/transactions/{id}/full [get]
func (h *PosHandler) GetTransactionFull(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid transaction id", nil))
		return
	}
	resp := h.useCase.GetTransactionFull(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// VoidTransactionRequest is the body for voiding a transaction.
type VoidTransactionRequest struct {
	VoidedBy int32  `json:"voided_by" binding:"required"` // User ID performing the void
	Reason   string `json:"reason"`                        // Optional reason
}

// VoidTransaction handles POST /api/pos/transactions/:id/void
// @Summary      Void a POS transaction
// @Description  Voids a completed POS transaction
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Transaction ID"
// @Param        body          body      VoidTransactionRequest true "Void payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/transactions/{id}/void [post]
func (h *PosHandler) VoidTransaction(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid transaction id", nil))
		return
	}
	var req VoidTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}
	resp := h.useCase.VoidTransaction(c.Request.Context(), int32(id), req.VoidedBy, req.Reason)
	c.JSON(resp.StatusCode, resp)
}

// GetTransactionPayments handles GET /api/pos/transactions/:id/payments
// @Summary      Get payments for a POS transaction
// @Description  Returns all payment records for a POS transaction (method, gateway, amount, reference)
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Transaction ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/transactions/{id}/payments [get]
func (h *PosHandler) GetTransactionPayments(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid transaction id", nil))
		return
	}
	resp := h.useCase.GetTransactionPayments(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// GetTransactionPaymentSummary handles GET /api/pos/transactions/:id/payment-summary
// @Summary      Get payment summary for a POS transaction
// @Description  Returns total paid and payment breakdown for a POS transaction
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Transaction ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/transactions/{id}/payment-summary [get]
func (h *PosHandler) GetTransactionPaymentSummary(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid transaction id", nil))
		return
	}
	resp := h.useCase.GetTransactionPaymentSummary(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}
