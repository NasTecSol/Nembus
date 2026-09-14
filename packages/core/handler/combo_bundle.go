package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// ComboBundleHandler handles HTTP requests for combo deals & promotional bundles.
type ComboBundleHandler struct {
	useCase      *usecase.ComboBundleUseCase
	promoUseCase *usecase.PromotionUseCase
}

func NewComboBundleHandler(uc *usecase.ComboBundleUseCase, promoUC *usecase.PromotionUseCase) *ComboBundleHandler {
	return &ComboBundleHandler{
		useCase:      uc,
		promoUseCase: promoUC,
	}
}

func (h *ComboBundleHandler) getRepository(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// ─── Request DTOs ────────────────────────────────────────────────────────────

type CreateComboBundleRequest struct {
	OrganizationID          int32                    `json:"organization_id" binding:"required"`
	StoreID                 *int32                   `json:"store_id"`
	PriceListID             *int32                   `json:"price_list_id"`
	ApplicableCustomerTypes []string                 `json:"applicable_customer_types"`
	Code                    string                   `json:"code" binding:"required"`
	Name                    string                   `json:"name" binding:"required"`
	Description             string                   `json:"description"`
	BundlePrice             float64                  `json:"bundle_price" binding:"required"`
	BundleType              string                   `json:"bundle_type"` // fixed, build_your_own, meal_deal, bogo
	IsActive                *bool                    `json:"is_active"`
	ValidFrom               *string                  `json:"valid_from"` // YYYY-MM-DD
	ValidTo                 *string                  `json:"valid_to"`   // YYYY-MM-DD
	DisplayOrder            int32                    `json:"display_order"`
	Metadata                map[string]interface{}   `json:"metadata"`
	Items                   []ComboBundleItemRequest `json:"items"`
}

type UpdateComboBundleRequest struct {
	StoreID                 *int32                 `json:"store_id"`
	PriceListID             *int32                 `json:"price_list_id"`
	ApplicableCustomerTypes []string               `json:"applicable_customer_types"`
	Code                    string                 `json:"code" binding:"required"`
	Name                    string                 `json:"name" binding:"required"`
	Description             string                 `json:"description"`
	BundlePrice             float64                `json:"bundle_price" binding:"required"`
	BundleType              string                 `json:"bundle_type"`
	IsActive                *bool                  `json:"is_active"`
	ValidFrom               *string                `json:"valid_from"`
	ValidTo                 *string                `json:"valid_to"`
	DisplayOrder            int32                  `json:"display_order"`
	Metadata                map[string]interface{} `json:"metadata"`
}

type ToggleComboBundleActiveRequest struct {
	IsActive bool `json:"is_active"`
}

type ComboBundleItemRequest struct {
	MenuItemID       *int32                 `json:"menu_item_id"`
	ProductID        *int32                 `json:"product_id"`
	ProductVariantID *int32                 `json:"product_variant_id"`
	ItemType         string                 `json:"item_type"` // menu_item, product
	Quantity         float64                `json:"quantity"`
	IsRequired       *bool                  `json:"is_required"`
	GroupTag         string                 `json:"group_tag"`
	PriceOverride    *float64               `json:"price_override"`
	DisplayOrder     int32                  `json:"display_order"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type BatchSetComboBundleItemsRequest struct {
	Items []ComboBundleItemRequest `json:"items" binding:"required"`
}

type CreateBundlePromotionRequest struct {
	CouponCode string `json:"coupon_code"`
}

// ─── Bundle Handlers ─────────────────────────────────────────────────────────

// CreateComboBundle handles POST /api/combo-bundles
// @Summary      Create a combo bundle or promotional deal
// @Description  Creates a new sellable combo bundle or promotional kit for retail, wholesale, or restaurant
// @Tags         combo-bundles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                   true  "Tenant identifier"
// @Param        Authorization header    string                   true  "Bearer token"
// @Param        bundle        body      CreateComboBundleRequest true  "Combo bundle data"
// @Success      201           {object}  ComboBundleResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/combo-bundles [post]
func (h *ComboBundleHandler) CreateComboBundle(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateComboBundleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata json", nil))
		return
	}

	validFrom, err := datePtrFromYMD(req.ValidFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid valid_from date format", nil))
		return
	}

	validTo, err := datePtrFromYMD(req.ValidTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid valid_to date format", nil))
		return
	}

	priceNumeric, err := numericFromString(fmt.Sprintf("%.2f", req.BundlePrice))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid bundle_price", nil))
		return
	}

	bundleType := req.BundleType
	if bundleType == "" {
		bundleType = "fixed"
	}

	customerTypes := req.ApplicableCustomerTypes
	if len(customerTypes) == 0 {
		customerTypes = []string{"retail", "wholesale"}
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	arg := repository.CreateComboBundleParams{
		OrganizationID:          req.OrganizationID,
		StoreID:                 int4Ptr(req.StoreID),
		PriceListID:             int4Ptr(req.PriceListID),
		ApplicableCustomerTypes: customerTypes,
		Code:                    req.Code,
		Name:                    req.Name,
		Description:             textPtr(&req.Description),
		BundlePrice:             priceNumeric,
		BundleType:              textPtr(&bundleType),
		IsActive:                pgtype.Bool{Bool: isActive, Valid: true},
		ValidFrom:               validFrom,
		ValidTo:                 validTo,
		DisplayOrder:            pgtype.Int4{Int32: req.DisplayOrder, Valid: true},
		Metadata:                metaBytes,
	}

	resp := h.useCase.CreateComboBundle(c.Request.Context(), arg)
	if resp.StatusCode != http.StatusCreated {
		c.JSON(resp.StatusCode, resp)
		return
	}

	// If initial items provided, create them atomically
	if len(req.Items) > 0 {
		createdBundle, ok := resp.Data.(usecase.ComboBundleOutput)
		if ok {
			var itemParams []repository.CreateComboBundleItemParams
			for _, it := range req.Items {
				p, err := toCreateItemParam(createdBundle.ID, it)
				if err == nil {
					itemParams = append(itemParams, p)
				}
			}
			if len(itemParams) > 0 {
				_ = h.useCase.BatchSetComboBundleItems(c.Request.Context(), createdBundle.ID, itemParams)
			}
			// Fetch fresh bundle with items
			resp = h.useCase.GetComboBundle(c.Request.Context(), createdBundle.ID)
		}
	}

	c.JSON(resp.StatusCode, resp)
}

// GetComboBundle handles GET /api/combo-bundles/:id
// @Summary      Get combo bundle by ID
// @Description  Retrieves a combo bundle with all its configured items and details
// @Tags         combo-bundles
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string true "Tenant identifier"
// @Param        Authorization header    string true "Bearer token"
// @Param        id            path      int    true "Combo Bundle ID"
// @Success      200           {object}  ComboBundleResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/combo-bundles/{id} [get]
func (h *ComboBundleHandler) GetComboBundle(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid bundle id", nil))
		return
	}

	resp := h.useCase.GetComboBundle(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// ListComboBundles handles GET /api/combo-bundles
// @Summary      List combo bundles
// @Description  List combo bundles for an organization with optional store, price list, and active status filters
// @Tags         combo-bundles
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string true  "Tenant identifier"
// @Param        Authorization   header    string true  "Bearer token"
// @Param        organization_id query     int    true  "Organization ID"
// @Param        store_id        query     int    false "Store ID (optional)"
// @Param        price_list_id   query     int    false "Price List ID (optional)"
// @Param        is_active       query     bool   false "Active status filter (optional)"
// @Success      200             {array}   ComboBundleResponse
// @Failure      400             {object}  ErrorResponse
// @Router       /api/combo-bundles [get]
func (h *ComboBundleHandler) ListComboBundles(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id query param is required", nil))
		return
	}
	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid organization_id", nil))
		return
	}

	var storeID pgtype.Int4
	if s := c.Query("store_id"); s != "" {
		if val, err := strconv.ParseInt(s, 10, 32); err == nil {
			storeID = pgtype.Int4{Int32: int32(val), Valid: true}
		}
	}

	var priceListID pgtype.Int4
	if p := c.Query("price_list_id"); p != "" {
		if val, err := strconv.ParseInt(p, 10, 32); err == nil {
			priceListID = pgtype.Int4{Int32: int32(val), Valid: true}
		}
	}

	var isActive pgtype.Bool
	if a := c.Query("is_active"); a != "" {
		if val, err := strconv.ParseBool(a); err == nil {
			isActive = pgtype.Bool{Bool: val, Valid: true}
		}
	}

	resp := h.useCase.ListComboBundlesByOrg(c.Request.Context(), repository.ListComboBundlesByOrganizationParams{
		OrganizationID: int32(orgID),
		StoreID:        storeID,
		PriceListID:    priceListID,
		IsActive:       isActive,
	})
	c.JSON(resp.StatusCode, resp)
}

// ListStoreComboBundles handles GET /restaurant/stores/:store_id/combo-bundles
// @Summary      List active combo bundles for a store
// @Description  Retrieves active combo bundles valid for a store (used by POS terminals & kiosks)
// @Tags         combo-bundles
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string true "Tenant identifier"
// @Param        Authorization header    string true "Bearer token"
// @Param        store_id      path      int    true "Store ID"
// @Success      200           {array}   ComboBundleResponse
// @Failure      400           {object}  ErrorResponse
// @Router       /api/stores/{store_id}/combo-bundles [get]
func (h *ComboBundleHandler) ListStoreComboBundles(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeIDStr := c.Param("store_id")
	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store id", nil))
		return
	}

	resp := h.useCase.ListComboBundlesByStore(c.Request.Context(), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// UpdateComboBundle handles PUT /api/combo-bundles/:id
// @Summary      Update a combo bundle
// @Description  Updates combo bundle header details (pricing, customer types, schedules)
// @Tags         combo-bundles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                   true "Tenant identifier"
// @Param        Authorization header    string                   true "Bearer token"
// @Param        id            path      int                      true "Combo Bundle ID"
// @Param        bundle        body      UpdateComboBundleRequest true "Updated bundle data"
// @Success      200           {object}  ComboBundleResponse
// @Failure      400           {object}  ErrorResponse
// @Router       /api/combo-bundles/{id} [put]
func (h *ComboBundleHandler) UpdateComboBundle(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid bundle id", nil))
		return
	}

	var req UpdateComboBundleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata json", nil))
		return
	}

	validFrom, err := datePtrFromYMD(req.ValidFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid valid_from date format", nil))
		return
	}

	validTo, err := datePtrFromYMD(req.ValidTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid valid_to date format", nil))
		return
	}

	priceNumeric, err := numericFromString(fmt.Sprintf("%.2f", req.BundlePrice))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid bundle_price", nil))
		return
	}

	customerTypes := req.ApplicableCustomerTypes
	if len(customerTypes) == 0 {
		customerTypes = []string{"retail", "wholesale"}
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	arg := repository.UpdateComboBundleParams{
		ID:                      int32(id),
		StoreID:                 int4Ptr(req.StoreID),
		PriceListID:             int4Ptr(req.PriceListID),
		ApplicableCustomerTypes: customerTypes,
		Code:                    req.Code,
		Name:                    req.Name,
		Description:             textPtr(&req.Description),
		BundlePrice:             priceNumeric,
		BundleType:              textPtr(&req.BundleType),
		IsActive:                pgtype.Bool{Bool: isActive, Valid: true},
		ValidFrom:               validFrom,
		ValidTo:                 validTo,
		DisplayOrder:            pgtype.Int4{Int32: req.DisplayOrder, Valid: true},
		Metadata:                metaBytes,
	}

	resp := h.useCase.UpdateComboBundle(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// ToggleComboBundleActive handles PATCH /api/combo-bundles/:id/toggle
// @Summary      Toggle combo bundle active status
// @Description  Activates or deactivates a combo bundle
// @Tags         combo-bundles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                         true "Tenant identifier"
// @Param        Authorization header    string                         true "Bearer token"
// @Param        id            path      int                            true "Combo Bundle ID"
// @Param        body          body      ToggleComboBundleActiveRequest true "Status"
// @Success      200           {object}  ComboBundleResponse
// @Failure      400           {object}  ErrorResponse
// @Router       /api/combo-bundles/{id}/toggle [patch]
func (h *ComboBundleHandler) ToggleComboBundleActive(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid bundle id", nil))
		return
	}

	var req ToggleComboBundleActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.ToggleComboBundleActive(c.Request.Context(), int32(id), req.IsActive)
	c.JSON(resp.StatusCode, resp)
}

// DeleteComboBundle handles DELETE /api/combo-bundles/:id
// @Summary      Delete a combo bundle
// @Description  Deletes a combo bundle and its associated items
// @Tags         combo-bundles
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string true "Tenant identifier"
// @Param        Authorization header    string true "Bearer token"
// @Param        id            path      int    true "Combo Bundle ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Router       /api/combo-bundles/{id} [delete]
func (h *ComboBundleHandler) DeleteComboBundle(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid bundle id", nil))
		return
	}

	resp := h.useCase.DeleteComboBundle(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// ─── Bundle Item Handlers ────────────────────────────────────────────────────

// AddComboBundleItem handles POST /api/combo-bundles/:id/items
// @Summary      Add item to combo bundle
// @Description  Adds a product or menu item slot to a combo bundle
// @Tags         combo-bundles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                 true "Tenant identifier"
// @Param        Authorization header    string                 true "Bearer token"
// @Param        id            path      int                    true "Combo Bundle ID"
// @Param        item          body      ComboBundleItemRequest true "Item definition"
// @Success      201           {object}  ComboBundleItemResponse
// @Failure      400           {object}  ErrorResponse
// @Router       /api/combo-bundles/{id}/items [post]
func (h *ComboBundleHandler) AddComboBundleItem(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	bundleID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid bundle id", nil))
		return
	}

	var req ComboBundleItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	param, err := toCreateItemParam(int32(bundleID), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.AddComboBundleItem(c.Request.Context(), param)
	c.JSON(resp.StatusCode, resp)
}

// ListComboBundleItems handles GET /api/combo-bundles/:id/items
// @Summary      List items of a combo bundle
// @Description  Lists all items configured for a specific combo bundle
// @Tags         combo-bundles
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string true "Tenant identifier"
// @Param        Authorization header    string true "Bearer token"
// @Param        id            path      int    true "Combo Bundle ID"
// @Success      200           {array}   ComboBundleItemResponse
// @Failure      400           {object}  ErrorResponse
// @Router       /api/combo-bundles/{id}/items [get]
func (h *ComboBundleHandler) ListComboBundleItems(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	bundleID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid bundle id", nil))
		return
	}

	resp := h.useCase.ListComboBundleItems(c.Request.Context(), int32(bundleID))
	c.JSON(resp.StatusCode, resp)
}

// BatchSetComboBundleItems handles PUT /api/combo-bundles/:id/items/batch
// @Summary      Batch set combo bundle items
// @Description  Atomically replaces all items for a combo bundle
// @Tags         combo-bundles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                          true "Tenant identifier"
// @Param        Authorization header    string                          true "Bearer token"
// @Param        id            path      int                             true "Combo Bundle ID"
// @Param        body          body      BatchSetComboBundleItemsRequest true "List of items"
// @Success      200           {array}   ComboBundleItemResponse
// @Failure      400           {object}  ErrorResponse
// @Router       /api/combo-bundles/{id}/items/batch [put]
func (h *ComboBundleHandler) BatchSetComboBundleItems(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	bundleID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid bundle id", nil))
		return
	}

	var req BatchSetComboBundleItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	var items []repository.CreateComboBundleItemParams
	for _, it := range req.Items {
		p, err := toCreateItemParam(int32(bundleID), it)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
			return
		}
		items = append(items, p)
	}

	resp := h.useCase.BatchSetComboBundleItems(c.Request.Context(), int32(bundleID), items)
	c.JSON(resp.StatusCode, resp)
}

// UpdateComboBundleItem handles PUT /api/combo-bundle-items/:item_id
// @Summary      Update a combo bundle item
// @Description  Updates a specific combo bundle item slot
// @Tags         combo-bundles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                 true "Tenant identifier"
// @Param        Authorization header    string                 true "Bearer token"
// @Param        item_id       path      int                    true "Item ID"
// @Param        item          body      ComboBundleItemRequest true "Item data"
// @Success      200           {object}  ComboBundleItemResponse
// @Failure      400           {object}  ErrorResponse
// @Router       /api/combo-bundle-items/{item_id} [put]
func (h *ComboBundleHandler) UpdateComboBundleItem(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	itemIDStr := c.Param("item_id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid item id", nil))
		return
	}

	var req ComboBundleItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}
	qtyNumeric, err := numericFromString(fmt.Sprintf("%.3f", qty))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid quantity", nil))
		return
	}

	var priceOverrideNumeric pgtype.Numeric
	if req.PriceOverride != nil {
		p, err := numericFromString(fmt.Sprintf("%.2f", *req.PriceOverride))
		if err == nil {
			priceOverrideNumeric = p
		}
	}

	metaBytes, _ := bytesFromMap(req.Metadata)
	isRequired := true
	if req.IsRequired != nil {
		isRequired = *req.IsRequired
	}

	itemType := req.ItemType
	if itemType == "" {
		if req.MenuItemID != nil {
			itemType = "menu_item"
		} else {
			itemType = "product"
		}
	}

	resp := h.useCase.UpdateComboBundleItem(c.Request.Context(), repository.UpdateComboBundleItemParams{
		ID:               int32(itemID),
		MenuItemID:       int4Ptr(req.MenuItemID),
		ProductID:        int4Ptr(req.ProductID),
		ProductVariantID: int4Ptr(req.ProductVariantID),
		ItemType:         textPtr(&itemType),
		Quantity:         qtyNumeric,
		IsRequired:       pgtype.Bool{Bool: isRequired, Valid: true},
		GroupTag:         textPtr(&req.GroupTag),
		PriceOverride:    priceOverrideNumeric,
		DisplayOrder:     pgtype.Int4{Int32: req.DisplayOrder, Valid: true},
		Metadata:         metaBytes,
	})
	c.JSON(resp.StatusCode, resp)
}

// DeleteComboBundleItem handles DELETE /api/combo-bundle-items/:item_id
// @Summary      Delete a combo bundle item
// @Description  Removes an item slot from a combo bundle
// @Tags         combo-bundles
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string true "Tenant identifier"
// @Param        Authorization header    string true "Bearer token"
// @Param        item_id       path      int    true "Item ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Router       /api/combo-bundle-items/{item_id} [delete]
func (h *ComboBundleHandler) DeleteComboBundleItem(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	itemIDStr := c.Param("item_id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid item id", nil))
		return
	}

	resp := h.useCase.DeleteComboBundleItem(c.Request.Context(), int32(itemID))
	c.JSON(resp.StatusCode, resp)
}

// ─── Bridge: Create Promotion from Combo Bundle ──────────────────────────────

// CreatePromotionFromCombo handles POST /api/combo-bundles/:id/create-promotion
// @Summary      Create promotion from combo bundle
// @Description  Bridges a sellable combo bundle to the promotion engine as an automatic or coupon-driven bundle promotion
// @Tags         combo-bundles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                       true  "Tenant identifier"
// @Param        Authorization header    string                       true  "Bearer token"
// @Param        id            path      int                          true  "Combo Bundle ID"
// @Param        body          body      CreateBundlePromotionRequest false "Optional coupon code"
// @Success      201           {object}  PromotionResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/combo-bundles/{id}/create-promotion [post]
func (h *ComboBundleHandler) CreatePromotionFromCombo(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	if h.promoUseCase == nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "promotion usecase not initialized", nil))
		return
	}
	h.promoUseCase.SetRepository(repo)

	idStr := c.Param("id")
	bundleID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid bundle id", nil))
		return
	}

	var req CreateBundlePromotionRequest
	_ = c.ShouldBindJSON(&req)

	resp := h.promoUseCase.CreateBundlePromotionFromCombo(c.Request.Context(), int32(bundleID), req.CouponCode)
	c.JSON(resp.StatusCode, resp)
}

// ─── Helper ──────────────────────────────────────────────────────────────────

func toCreateItemParam(bundleID int32, req ComboBundleItemRequest) (repository.CreateComboBundleItemParams, error) {
	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}
	qtyNumeric, err := numericFromString(fmt.Sprintf("%.3f", qty))
	if err != nil {
		return repository.CreateComboBundleItemParams{}, err
	}

	var priceOverrideNumeric pgtype.Numeric
	if req.PriceOverride != nil {
		p, err := numericFromString(fmt.Sprintf("%.2f", *req.PriceOverride))
		if err == nil {
			priceOverrideNumeric = p
		}
	}

	metaBytes, _ := bytesFromMap(req.Metadata)
	isRequired := true
	if req.IsRequired != nil {
		isRequired = *req.IsRequired
	}

	itemType := req.ItemType
	if itemType == "" {
		if req.MenuItemID != nil {
			itemType = "menu_item"
		} else {
			itemType = "product"
		}
	}

	return repository.CreateComboBundleItemParams{
		ComboBundleID:    bundleID,
		MenuItemID:       int4Ptr(req.MenuItemID),
		ProductID:        int4Ptr(req.ProductID),
		ProductVariantID: int4Ptr(req.ProductVariantID),
		ItemType:         textPtr(&itemType),
		Quantity:         qtyNumeric,
		IsRequired:       pgtype.Bool{Bool: isRequired, Valid: true},
		GroupTag:         textPtr(&req.GroupTag),
		PriceOverride:    priceOverrideNumeric,
		DisplayOrder:     pgtype.Int4{Int32: req.DisplayOrder, Valid: true},
		Metadata:         metaBytes,
	}, nil
}
