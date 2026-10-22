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

// GetZReport calculates a session report without changing session state.
// @Summary Calculate a cashier session Z-report
// @Description Recalculates from existing records. finalized is true only for a closed session with a closing time. Requires user store access. Does not close a session or save a snapshot.
// @Tags reports
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param cashier_session_id query int true "Cashier session ID"
// @Param store_id query int false "Must match the session store"
// @Param pos_terminal_id query int false "Must match the session terminal"
// @Param cashier_id query int false "Must match the session cashier"
// @Success 200 {object} repository.Response{data=usecase.ZReport}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/reports/z-report [get]
func (h *CashierSessionHandler) GetZReport(c *gin.Context) {
	// Never cache an on-demand financial report.
	c.Header("Cache-Control", "no-store")
	orgID, ok := middleware.GetOrganizationIDFromContext(c)
	if !ok || orgID <= 0 {
		c.JSON(http.StatusUnauthorized, utils.NewResponse(utils.CodeUnauthorized, "organization claim missing", nil))
		return
	}
	userClaim, ok := middleware.GetUserIDFromContext(c)
	userID, err := strconv.ParseInt(userClaim, 10, 32)
	if !ok || err != nil || userID <= 0 {
		c.JSON(http.StatusUnauthorized, utils.NewResponse(utils.CodeUnauthorized, "valid user claim required", nil))
		return
	}
	filters := repository.GetCashierSessionZReportParams{OrganizationID: orgID, UserID: int32(userID)}
	for _, field := range []struct {
		name string
		dest *int32
	}{
		{"cashier_session_id", &filters.CashierSessionID},
		{"store_id", &filters.StoreID},
		{"pos_terminal_id", &filters.PosTerminalID},
		{"cashier_id", &filters.CashierID},
	} {
		values, exists := c.Request.URL.Query()[field.name]
		if !exists && field.name != "cashier_session_id" {
			continue
		}
		if len(values) != 1 {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, field.name+" must be provided exactly once as a positive integer", nil))
			return
		}
		value, err := strconv.ParseInt(values[0], 10, 32)
		if err != nil || value <= 0 {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid "+field.name+": expected a positive integer", nil))
			return
		}
		*field.dest = int32(value)
	}
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	// Request-scoped repository avoids mutating the shared handler/use case.
	resp := usecase.CalculateCashierSessionZReport(c.Request.Context(), repo, filters)
	c.JSON(resp.StatusCode, resp)
}
