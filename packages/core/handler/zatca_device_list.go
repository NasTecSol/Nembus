package handler

import (
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
)

// ListZatcaDevices lists saved device registrations for backoffice inspection.
// @Summary List registered ZATCA devices by store and terminal
// @Description Returns saved ZATCA device registrations for the requested store within the authenticated user's organization and tenant. Optionally filter by POS terminal; both filters use AND. Includes terminal name/code, device serial, environment, active/revoked flags, credential expiry and reporting readiness. Includes incomplete, inactive, expired and revoked registrations so backoffice users can inspect them; reporting_ready is true only when organization ZATCA is enabled and valid reporting credentials are active. Empty or unmatched stores/terminals return devices=[] and total=0. Ordered by terminal ID then newest device ID. Available even when organization ZATCA is disabled. Reads saved state only; never calls ZATCA or returns certificates, private keys, secrets or raw metadata. Requires a user JWT with an organization_id claim.
// @Tags zatca
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param store_id query int true "Store ID" minimum(1) maximum(2147483647)
// @Param pos_terminal_id query int false "POS terminal ID within the requested store" minimum(1) maximum(2147483647)
// @Success 200 {object} ZatcaDeviceListAPIResponse
// @Failure 400 {object} ZatcaAPIError
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ZatcaAPIError
// @Router /api/zatca/devices [get]
func (h *ZatcaHandler) ListZatcaDevices(c *gin.Context) {
	var storeID, terminalID int32
	for _, field := range []struct {
		name     string
		dest     *int32
		required bool
	}{
		{"store_id", &storeID, true}, {"pos_terminal_id", &terminalID, false},
	} {
		values, exists := c.Request.URL.Query()[field.name]
		if !exists && !field.required {
			continue
		}
		if len(values) != 1 {
			c.JSON(400, utils.NewResponse(400, field.name+" must be provided exactly once as a positive integer", nil))
			return
		}
		id, err := strconv.ParseInt(values[0], 10, 32)
		if err != nil || id <= 0 {
			c.JSON(400, utils.NewResponse(400, field.name+" must be a positive integer no greater than 2147483647", nil))
			return
		}
		*field.dest = int32(id)
	}
	orgID, ok := middleware.GetOrganizationIDFromContext(c)
	if !ok || orgID <= 0 {
		c.JSON(401, utils.NewResponse(401, "organization claim missing", nil))
		return
	}
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok || repo == nil {
		c.JSON(500, utils.NewResponse(500, "tenant repository is required", nil))
		return
	}
	devices, err := repo.ListRegisteredZatcaDevices(c.Request.Context(), orgID, storeID, terminalID)
	if err != nil {
		c.JSON(500, utils.NewResponse(500, "failed to load registered ZATCA devices", nil))
		return
	}
	c.JSON(200, utils.NewResponse(200, "Registered ZATCA devices", gin.H{
		"devices": devices, "total": len(devices), "store_id": storeID,
	}))
}
