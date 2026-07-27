package handler

import (
	"net/http"
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
)

// UOMHandler handles Unit of Measure and product UOM conversion endpoints.
type UOMHandler struct {
	useCase *usecase.UOMUseCase
}

// NewUOMHandler creates a new UOM handler.
func NewUOMHandler(uc *usecase.UOMUseCase) *UOMHandler {
	return &UOMHandler{useCase: uc}
}

func (h *UOMHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateUnitOfMeasureRequest represents the body for creating a UOM.
type CreateUnitOfMeasureRequest struct {
	Code          string                 `json:"code" binding:"required" example:"PCS"`
	Name          string                 `json:"name" binding:"required" example:"Pieces"`
	UomType       string                 `json:"uom_type" binding:"required" example:"quantity"`
	DecimalPlaces int32                  `json:"decimal_places" binding:"required" example:"0"`
	IsActive      bool                   `json:"is_active" example:"true"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// UpdateUnitOfMeasureRequest represents the body for updating a UOM.
type UpdateUnitOfMeasureRequest struct {
	Name          string                 `json:"name" binding:"required" example:"Pieces"`
	UomType       string                 `json:"uom_type" binding:"required" example:"quantity"`
	DecimalPlaces int32                  `json:"decimal_places" binding:"required" example:"0"`
	IsActive      bool                   `json:"is_active" example:"true"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// CreateProductUOMConversionRequest represents the body for creating a product UOM conversion.
type CreateProductUOMConversionRequest struct {
	FromUomID        int32                  `json:"from_uom_id" binding:"required" example:"1"`
	ToUomID          int32                  `json:"to_uom_id" binding:"required" example:"2"`
	ConversionFactor string                 `json:"conversion_factor" binding:"required" example:"12.0000"`
	IsDefault        bool                   `json:"is_default" example:"false"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// UpdateProductUOMConversionRequest represents the body for updating a product UOM conversion.
type UpdateProductUOMConversionRequest struct {
	ConversionFactor string                 `json:"conversion_factor" binding:"required" example:"12.5000"`
	IsDefault        bool                   `json:"is_default" example:"false"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// CreateUnitOfMeasure handles POST /api/uoms
// @Summary      Create a unit of measure
// @Description  Create a new unit of measure
// @Tags         uoms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                    true  "Tenant identifier"
// @Param        Authorization header    string                    true  "Bearer token"
// @Param        body          body      CreateUnitOfMeasureRequest true  "UOM payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/uoms [post]
func (h *UOMHandler) CreateUnitOfMeasure(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateUnitOfMeasureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.CreateUnitOfMeasure(
		c.Request.Context(),
		req.Code,
		req.Name,
		req.UomType,
		req.DecimalPlaces,
		req.IsActive,
		req.Metadata,
	)
	c.JSON(resp.StatusCode, resp)
}

// GetUnitOfMeasure handles GET /api/uoms/:id
// @Summary      Get unit of measure by ID
// @Description  Retrieve a unit of measure by its ID
// @Tags         uoms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "UOM ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/uoms/{id} [get]
func (h *UOMHandler) GetUnitOfMeasure(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetUnitOfMeasure(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// GetUnitOfMeasureByCode handles GET /api/uoms/code/:code
// @Summary      Get unit of measure by code
// @Description  Retrieve a unit of measure by its unique code
// @Tags         uoms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        code          path      string  true  "UOM code"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/uoms/code/{code} [get]
func (h *UOMHandler) GetUnitOfMeasureByCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	code := c.Param("code")
	resp := h.useCase.GetUnitOfMeasureByCode(c.Request.Context(), code)
	c.JSON(resp.StatusCode, resp)
}

// ListUnitsOfMeasure handles GET /api/uoms
// @Summary      List units of measure
// @Description  List all units of measure
// @Tags         uoms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/uoms [get]
func (h *UOMHandler) ListUnitsOfMeasure(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListUnitsOfMeasure(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListActiveUnitsOfMeasure handles GET /api/uoms/active
// @Summary      List active units of measure
// @Description  List only active units of measure
// @Tags         uoms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/uoms/active [get]
func (h *UOMHandler) ListActiveUnitsOfMeasure(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListActiveUnitsOfMeasure(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListUnitsByType handles GET /api/uoms/by-type?uom_type=
// @Summary      List units of measure by type
// @Description  List units of measure filtered by uom_type
// @Tags         uoms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        uom_type      query     string  true  "UOM type (e.g. quantity, weight)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/uoms/by-type [get]
func (h *UOMHandler) ListUnitsByType(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	uomType := c.Query("uom_type")
	resp := h.useCase.ListUnitsByType(c.Request.Context(), uomType)
	c.JSON(resp.StatusCode, resp)
}

// UpdateUnitOfMeasure handles PUT /api/uoms/:id
// @Summary      Update a unit of measure
// @Description  Update an existing unit of measure by ID
// @Tags         uoms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                     true  "Tenant identifier"
// @Param        Authorization header    string                     true  "Bearer token"
// @Param        id            path      string                     true  "UOM ID"
// @Param        body          body      UpdateUnitOfMeasureRequest true  "UOM payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/uoms/{id} [put]
func (h *UOMHandler) UpdateUnitOfMeasure(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req UpdateUnitOfMeasureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.UpdateUnitOfMeasure(
		c.Request.Context(),
		id,
		req.Name,
		req.UomType,
		req.DecimalPlaces,
		req.IsActive,
		req.Metadata,
	)
	c.JSON(resp.StatusCode, resp)
}

// DeleteUnitOfMeasure handles DELETE /api/uoms/:id
// @Summary      Delete a unit of measure
// @Description  Delete a unit of measure by ID
// @Tags         uoms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "UOM ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/uoms/{id} [delete]
func (h *UOMHandler) DeleteUnitOfMeasure(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeleteUnitOfMeasure(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// CreateProductUOMConversion handles POST /api/products/:product_id/uom-conversions
// @Summary      Create product UOM conversion
// @Description  Create a new UOM conversion for a product
// @Tags         uoms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                          true  "Tenant identifier"
// @Param        Authorization header    string                          true  "Bearer token"
// @Param        product_id    path      int                             true  "Product ID"
// @Param        body          body      CreateProductUOMConversionRequest true  "Conversion payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/products/{product_id}/uom-conversions [post]
func (h *UOMHandler) CreateProductUOMConversion(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productIDStr := c.Param("product_id")
	productID64, err := strconv.ParseInt(productIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil))
		return
	}

	var req CreateProductUOMConversionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.CreateProductUOMConversion(
		c.Request.Context(),
		int32(productID64),
		req.FromUomID,
		req.ToUomID,
		req.ConversionFactor,
		req.IsDefault,
		req.Metadata,
	)
	c.JSON(resp.StatusCode, resp)
}

// ListProductUOMConversions handles GET /api/products/:product_id/uom-conversions
// @Summary      List product UOM conversions
// @Description  List all UOM conversions for a product
// @Tags         uoms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        product_id    path      int     true  "Product ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/products/{product_id}/uom-conversions [get]
func (h *UOMHandler) ListProductUOMConversions(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productIDStr := c.Param("product_id")
	productID64, err := strconv.ParseInt(productIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil))
		return
	}

	resp := h.useCase.ListProductUOMConversions(c.Request.Context(), int32(productID64))
	c.JSON(resp.StatusCode, resp)
}

// GetProductUOMConversion handles GET /api/products/:product_id/uom-conversions/lookup?from_uom_id=&to_uom_id=
// @Summary      Get a specific product UOM conversion
// @Description  Retrieve a conversion row for a product given from_uom_id and to_uom_id
// @Tags         uoms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        product_id    path      int     true  "Product ID"
// @Param        from_uom_id   query     int     true  "From UOM ID"
// @Param        to_uom_id     query     int     true  "To UOM ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/products/{product_id}/uom-conversions/lookup [get]
func (h *UOMHandler) GetProductUOMConversion(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productIDStr := c.Param("product_id")
	productID64, err := strconv.ParseInt(productIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil))
		return
	}

	fromStr := c.Query("from_uom_id")
	toStr := c.Query("to_uom_id")
	fromID64, err := strconv.ParseInt(fromStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid from_uom_id", nil))
		return
	}
	toID64, err := strconv.ParseInt(toStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid to_uom_id", nil))
		return
	}

	resp := h.useCase.GetProductUOMConversion(
		c.Request.Context(),
		int32(productID64),
		int32(fromID64),
		int32(toID64),
	)
	c.JSON(resp.StatusCode, resp)
}

// UpdateProductUOMConversion handles PUT /api/uom-conversions/:id
// @Summary      Update a product UOM conversion
// @Description  Update an existing product UOM conversion by ID
// @Tags         uoms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                           true  "Tenant identifier"
// @Param        Authorization header    string                           true  "Bearer token"
// @Param        id            path      string                           true  "Conversion ID"
// @Param        body          body      UpdateProductUOMConversionRequest true  "Conversion payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/uom-conversions/{id} [put]
func (h *UOMHandler) UpdateProductUOMConversion(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req UpdateProductUOMConversionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.UpdateProductUOMConversion(
		c.Request.Context(),
		id,
		req.ConversionFactor,
		req.IsDefault,
		req.Metadata,
	)
	c.JSON(resp.StatusCode, resp)
}

// DeleteProductUOMConversion handles DELETE /api/uom-conversions/:id
// @Summary      Delete a product UOM conversion
// @Description  Delete a product UOM conversion by ID
// @Tags         uoms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Conversion ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/uom-conversions/{id} [delete]
func (h *UOMHandler) DeleteProductUOMConversion(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeleteProductUOMConversion(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

