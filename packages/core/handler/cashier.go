package handler

import (
	"net/http"
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type CashierHandler struct {
	useCase *usecase.CashierUseCase
}

func NewCashierHandler(uc *usecase.CashierUseCase) *CashierHandler {
	return &CashierHandler{
		useCase: uc,
	}
}

// getRepositoryFromContext extracts repository from Gin context
func (h *CashierHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateCashier handles POST /cashiers
// @Summary      Create a new cashier
// @Description  Create a new cashier with required fields
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        cashier       body      CreateCashierRequest  true  "Cashier data"
// @Success      201           {object}  CashierResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashiers [post]
func (h *CashierHandler) CreateCashier(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateCashierRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	var metaBytes []byte
	if req.Metadata != nil {
		var err error
		metaBytes, err = bytesFromMap(req.Metadata)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
			return
		}
	}

	// Convert string amounts to pgtype.Numeric
	var drawerLimit *pgtype.Numeric
	if req.DrawerLimit != nil {
		dl, err := numericFromString(*req.DrawerLimit)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid drawer_limit", nil))
			return
		}
		drawerLimit = &dl
	}

	var discountLimit *pgtype.Numeric
	if req.DiscountLimit != nil {
		dl, err := numericFromString(*req.DiscountLimit)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid discount_limit", nil))
			return
		}
		discountLimit = &dl
	}

	resp := h.useCase.CreateCashier(c.Request.Context(), req.UserID, req.StoreID, req.CashierCode, drawerLimit, discountLimit, req.IsActive, metaBytes)
	c.JSON(resp.StatusCode, resp)
}

// CreateCashierWithDefaults handles POST /cashiers/with-defaults
// @Summary      Create a new cashier with defaults
// @Description  Create a new cashier with default values
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        cashier       body      CreateCashierWithDefaultsRequest  true  "Cashier data"
// @Success      201           {object}  CashierResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashiers/with-defaults [post]
func (h *CashierHandler) CreateCashierWithDefaults(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req struct {
		UserID      int32  `json:"user_id" binding:"required"`
		StoreID     int32  `json:"store_id" binding:"required"`
		CashierCode string `json:"cashier_code" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	resp := h.useCase.CreateCashierWithDefaults(c.Request.Context(), req.UserID, req.StoreID, req.CashierCode)
	c.JSON(resp.StatusCode, resp)
}

// GetCashierByID handles GET /cashiers/:id
// @Summary      Get cashier by ID
// @Description  Retrieve a specific cashier by its ID
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cashier ID"
// @Success      200           {object}  CashierResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/cashiers/{id} [get]
func (h *CashierHandler) GetCashierByID(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetCashierByID(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// GetCashierByCode handles GET /cashiers/code/:code
// @Summary      Get cashier by code
// @Description  Retrieve a specific cashier by its code and store ID
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        code          path      string  true  "Cashier code"
// @Param        store_id     query     int     true  "Store ID"
// @Success      200           {object}  CashierResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/cashiers/code/{code} [get]
func (h *CashierHandler) GetCashierByCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	code := c.Param("code")
	storeIDStr := c.Query("store_id")
	if storeIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "store_id is required", nil))
		return
	}

	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	resp := h.useCase.GetCashierByCode(c.Request.Context(), code, int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// GetCashierByUserID handles GET /cashiers/user/:user_id
// @Summary      Get cashier by user ID
// @Description  Retrieve a cashier by user ID and store ID
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        user_id       path      string  true  "User ID"
// @Param        store_id      query     int     true  "Store ID"
// @Success      200           {object}  CashierResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/cashiers/user/{user_id} [get]
func (h *CashierHandler) GetCashierByUserID(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userIDStr := c.Param("user_id")
	storeIDStr := c.Query("store_id")
	if storeIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "store_id is required", nil))
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user_id", nil))
		return
	}

	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	resp := h.useCase.GetCashierByUserID(c.Request.Context(), int32(userID), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// ListAllCashiers handles GET /cashiers/all
// @Summary      List all cashiers
// @Description  Retrieve a list of all cashiers without pagination
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {array}   CashierResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashiers/all [get]
func (h *CashierHandler) ListAllCashiers(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListAllCashiers(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListActiveCashiers handles GET /cashiers/active
// @Summary      List active cashiers
// @Description  Retrieve a list of only active cashiers
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {array}   CashierResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashiers/active [get]
func (h *CashierHandler) ListActiveCashiers(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListActiveCashiers(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListCashiersByStore handles GET /cashiers/store/:store_id
// @Summary      List cashiers by store
// @Description  Retrieve all cashiers for a specific store
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      path      string  true  "Store ID"
// @Success      200           {array}   CashierResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashiers/store/{store_id} [get]
func (h *CashierHandler) ListCashiersByStore(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeIDStr := c.Param("store_id")
	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	resp := h.useCase.ListCashiersByStore(c.Request.Context(), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// ListActiveCashiersByStore handles GET /cashiers/store/:store_id/active
// @Summary      List active cashiers by store
// @Description  Retrieve active cashiers for a specific store
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      path      string  true  "Store ID"
// @Success      200           {array}   CashierResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashiers/store/{store_id}/active [get]
func (h *CashierHandler) ListActiveCashiersByStore(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeIDStr := c.Param("store_id")
	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	resp := h.useCase.ListActiveCashiersByStore(c.Request.Context(), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// ListCashiers handles GET /cashiers
// @Summary      List cashiers with pagination
// @Description  Retrieve a list of cashiers with pagination
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        limit         query     int     false "Limit"
// @Param        offset        query     int     false "Offset"
// @Success      200           {array}   CashierResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashiers [get]
func (h *CashierHandler) ListCashiers(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 32)
	offset, _ := strconv.ParseInt(offsetStr, 10, 32)

	resp := h.useCase.ListCashiersWithPagination(c.Request.Context(), int32(limit), int32(offset))
	c.JSON(resp.StatusCode, resp)
}

// CountCashiers handles GET /cashiers/count
// @Summary      Count cashiers
// @Description  Get total number of cashiers
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  CountResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashiers/count [get]
func (h *CashierHandler) CountCashiers(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.CountCashiers(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// CountActiveCashiers handles GET /cashiers/count/active
// @Summary      Count active cashiers
// @Description  Get total number of active cashiers
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  CountResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashiers/count/active [get]
func (h *CashierHandler) CountActiveCashiers(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.CountActiveCashiers(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// CountCashiersByStore handles GET /cashiers/count/store/:store_id
// @Summary      Count cashiers by store
// @Description  Get total number of cashiers in a specific store
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      path      string  true  "Store ID"
// @Success      200           {object}  CountResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashiers/count/store/{store_id} [get]
func (h *CashierHandler) CountCashiersByStore(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeIDStr := c.Param("store_id")
	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	resp := h.useCase.CountCashiersByStore(c.Request.Context(), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// UpdateCashier handles PUT /cashiers/:id
// @Summary      Update a cashier
// @Description  Update an existing cashier by its ID
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cashier ID"
// @Param        cashier       body      UpdateCashierRequest  true  "Cashier data"
// @Success      200           {object}  CashierResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashiers/{id} [put]
func (h *CashierHandler) UpdateCashier(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req UpdateCashierRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	var metaBytes []byte
	if req.Metadata != nil {
		var err error
		metaBytes, err = bytesFromMap(req.Metadata)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
			return
		}
	}

	// Convert string amounts to pgtype.Numeric
	var drawerLimit *pgtype.Numeric
	if req.DrawerLimit != nil {
		dl, err := numericFromString(*req.DrawerLimit)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid drawer_limit", nil))
			return
		}
		drawerLimit = &dl
	}

	var discountLimit *pgtype.Numeric
	if req.DiscountLimit != nil {
		dl, err := numericFromString(*req.DiscountLimit)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid discount_limit", nil))
			return
		}
		discountLimit = &dl
	}

	resp := h.useCase.UpdateCashier(c.Request.Context(), id, req.UserID, req.StoreID, req.CashierCode, drawerLimit, discountLimit, req.IsActive, metaBytes)
	c.JSON(resp.StatusCode, resp)
}

// UpdateCashierLimits handles PATCH /cashiers/:id/limits
// @Summary      Update cashier limits
// @Description  Update cashier drawer and discount limits
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cashier ID"
// @Param        limits        body      UpdateCashierLimitsRequest  true  "Limits data"
// @Success      200           {object}  CashierResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/cashiers/{id}/limits [patch]
func (h *CashierHandler) UpdateCashierLimits(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req UpdateCashierLimitsRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	drawerLimit, err := numericFromString(req.DrawerLimit)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid drawer_limit", nil))
		return
	}

	discountLimit, err := numericFromString(req.DiscountLimit)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid discount_limit", nil))
		return
	}

	resp := h.useCase.UpdateCashierLimits(c.Request.Context(), id, drawerLimit, discountLimit)
	c.JSON(resp.StatusCode, resp)
}

// UpdateCashierDrawerLimit handles PATCH /cashiers/:id/drawer-limit
// @Summary      Update cashier drawer limit
// @Description  Update only the drawer limit
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cashier ID"
// @Param        limit         body      UpdateCashierDrawerLimitRequest  true  "Drawer limit"
// @Success      200           {object}  CashierResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/cashiers/{id}/drawer-limit [patch]
func (h *CashierHandler) UpdateCashierDrawerLimit(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req UpdateCashierDrawerLimitRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	drawerLimit, err := numericFromString(req.DrawerLimit)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid drawer_limit", nil))
		return
	}

	resp := h.useCase.UpdateCashierDrawerLimit(c.Request.Context(), id, drawerLimit)
	c.JSON(resp.StatusCode, resp)
}

// UpdateCashierDiscountLimit handles PATCH /cashiers/:id/discount-limit
// @Summary      Update cashier discount limit
// @Description  Update only the discount limit
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cashier ID"
// @Param        limit         body      UpdateCashierDiscountLimitRequest  true  "Discount limit"
// @Success      200           {object}  CashierResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/cashiers/{id}/discount-limit [patch]
func (h *CashierHandler) UpdateCashierDiscountLimit(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req UpdateCashierDiscountLimitRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	discountLimit, err := numericFromString(req.DiscountLimit)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid discount_limit", nil))
		return
	}

	resp := h.useCase.UpdateCashierDiscountLimit(c.Request.Context(), id, discountLimit)
	c.JSON(resp.StatusCode, resp)
}

// UpdateCashierMetadata handles PATCH /cashiers/:id/metadata
// @Summary      Update cashier metadata
// @Description  Update cashier metadata
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cashier ID"
// @Param        metadata      body      UpdateCashierMetadataRequest  true  "Metadata"
// @Success      200           {object}  CashierResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/cashiers/{id}/metadata [patch]
func (h *CashierHandler) UpdateCashierMetadata(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req UpdateCashierMetadataRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	resp := h.useCase.UpdateCashierMetadata(c.Request.Context(), id, metaBytes)
	c.JSON(resp.StatusCode, resp)
}

// ActivateCashier handles PATCH /cashiers/:id/activate
// @Summary      Activate a cashier
// @Description  Activate a cashier
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cashier ID"
// @Success      200           {object}  CashierResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/cashiers/{id}/activate [patch]
func (h *CashierHandler) ActivateCashier(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.ActivateCashier(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// DeactivateCashier handles PATCH /cashiers/:id/deactivate
// @Summary      Deactivate a cashier
// @Description  Deactivate a cashier
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cashier ID"
// @Success      200           {object}  CashierResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/cashiers/{id}/deactivate [patch]
func (h *CashierHandler) DeactivateCashier(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeactivateCashier(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// DeleteCashier handles DELETE /cashiers/:id
// @Summary      Delete a cashier
// @Description  Hard delete a cashier by its ID
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cashier ID"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashiers/{id} [delete]
func (h *CashierHandler) DeleteCashier(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeleteCashier(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// SoftDeleteCashier handles DELETE /cashiers/:id/soft
// @Summary      Soft delete a cashier
// @Description  Soft delete a cashier by deactivating it
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cashier ID"
// @Success      200           {object}  CashierResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/cashiers/{id}/soft [delete]
func (h *CashierHandler) SoftDeleteCashier(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.SoftDeleteCashier(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// CashierExists handles GET /cashiers/:id/exists
// @Summary      Check if cashier exists
// @Description  Check if a cashier exists by ID
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cashier ID"
// @Success      200           {object}  ExistsResponse
// @Failure      401           {object}  ErrorResponse
// @Router       /api/cashiers/{id}/exists [get]
func (h *CashierHandler) CashierExists(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.CashierExists(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// CashierCodeExists handles GET /cashiers/code/:code/exists
// @Summary      Check if cashier code exists
// @Description  Check if a cashier code already exists for a store
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        code          path      string  true  "Cashier code"
// @Param        store_id      query     int     true  "Store ID"
// @Success      200           {object}  ExistsResponse
// @Failure      401           {object}  ErrorResponse
// @Router       /api/cashiers/code/{code}/exists [get]
func (h *CashierHandler) CashierCodeExists(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	code := c.Param("code")
	storeIDStr := c.Query("store_id")
	if storeIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "store_id is required", nil))
		return
	}

	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	resp := h.useCase.CashierCodeExists(c.Request.Context(), code, int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// GetCashierWithLimits handles GET /cashiers/:id/limits
// @Summary      Get cashier with limits
// @Description  Get cashier with limits and user details
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Cashier ID"
// @Success      200           {object}  CashierWithLimitsResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/cashiers/{id}/limits [get]
func (h *CashierHandler) GetCashierWithLimits(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetCashierWithLimits(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// ListActiveCashiersInStore handles GET /cashiers/store/:store_id/active-with-sessions
// @Summary      List active cashiers in store with sessions
// @Description  List active cashiers in a store with session information
// @Tags         cashiers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      path      string  true  "Store ID"
// @Success      200           {array}   CashierWithSessionsResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashiers/store/{store_id}/active-with-sessions [get]
func (h *CashierHandler) ListActiveCashiersInStore(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeIDStr := c.Param("store_id")
	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	resp := h.useCase.ListActiveCashiersInStore(c.Request.Context(), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}
