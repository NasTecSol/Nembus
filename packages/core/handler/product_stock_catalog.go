package handler

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
)

// GetMasterProductStockCatalog returns a paginated catalog classified by scoped available stock.
// @Summary      Master product catalog by stock level
// @Description  Returns inventory-tracked products with full catalog details. Stock is summed across variants and inventory rows within the selected store/location scope. Low means strictly below the threshold; non_low means at or above it. Only scoped inventory is embedded. Without a scope, products without inventory have zero stock; with a scope, products must have an inventory row in that scope.
// @Tags         products
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id          header string true  "Tenant identifier"
// @Param        organization_id      query  int    true  "Organization ID"
// @Param        low_stock_threshold  query  number false "Low stock threshold (default 20, non-negative)"
// @Param        stock_status         query  string false "Stock filter (default low)" Enums(low, non_low, all)
// @Param        category_id          query  int    false "Category ID (omit or 0 for all)"
// @Param        store_id             query  int    false "Store ID (omit or 0 for all)"
// @Param        storage_location_id  query  int    false "Storage location ID (omit or 0 for all)"
// @Param        limit                query  int    false "Page size (default 100)"
// @Param        offset               query  int    false "Page offset (default 0)"
// @Success      200 {object} SuccessResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/products/master-catalog/stock [get]
func (h *ProductCatalogHandler) GetMasterProductStockCatalog(c *gin.Context) {
	params, err := parseProductStockCatalogParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	// Keep this request's tenant repository local to the request.
	uc := usecase.NewProductCatalogUseCase()
	uc.SetRepository(repo)
	resp := uc.GetMasterProductStockCatalog(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

func parseProductStockCatalogParams(c *gin.Context) (repository.GetMasterProductStockCatalogParams, error) {
	var params repository.GetMasterProductStockCatalogParams
	for _, field := range []struct {
		name     string
		dest     *int32
		fallback string
		minimum  int64
	}{
		{"organization_id", &params.OrganizationID, "", 1},
		{"category_id", &params.CategoryID, "0", 0},
		{"store_id", &params.StoreID, "0", 0},
		{"storage_location_id", &params.StorageLocationID, "0", 0},
		{"limit", &params.Limit, "100", 1},
		{"offset", &params.Offset, "0", 0},
	} {
		value, err := strconv.ParseInt(c.DefaultQuery(field.name, field.fallback), 10, 32)
		if err != nil || value < field.minimum {
			return params, fmt.Errorf("invalid or missing %s", field.name)
		}
		*field.dest = int32(value)
	}
	params.StockStatus = c.DefaultQuery("stock_status", "low")
	if params.StockStatus != "low" && params.StockStatus != "non_low" && params.StockStatus != "all" {
		return params, fmt.Errorf("stock_status must be low, non_low, or all")
	}
	threshold, err := strconv.ParseFloat(c.DefaultQuery("low_stock_threshold", "20"), 64)
	if err != nil || math.IsNaN(threshold) || math.IsInf(threshold, 0) || threshold < 0 {
		return params, fmt.Errorf("low_stock_threshold must be a finite non-negative number")
	}
	if err := params.LowStockThreshold.Scan(strconv.FormatFloat(threshold, 'f', -1, 64)); err != nil {
		return params, fmt.Errorf("invalid low_stock_threshold")
	}
	return params, nil
}
