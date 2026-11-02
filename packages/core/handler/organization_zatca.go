package handler

import (
	"errors"
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (h *OrganizationHandler) organizationZatcaTarget(c *gin.Context) (*repository.Queries, int32, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		c.JSON(400, utils.NewResponse(400, "invalid organization id", nil))
		return nil, 0, false
	}
	if caller, ok := middleware.GetOrganizationIDFromContext(c); ok && caller > 0 && caller != int32(id) {
		c.JSON(403, utils.NewResponse(403, "organization does not match authenticated organization", nil))
		return nil, 0, false
	}
	repo := h.getRepositoryFromContext(c)
	return repo, int32(id), repo != nil
}

// @Summary      Read organization ZATCA enablement
// @Description  Reads organizations.metadata.zatca.enabled. Missing values mean disabled; environment-file flags are ignored.
// @Tags         zatca, organizations
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        id path int true "Organization ID"
// @Success      200 {object} ZatcaOrganizationSettingResponse
// @Failure      400 {object} ZatcaAPIError
// @Failure      403 {object} ZatcaAPIError
// @Failure      404 {object} ZatcaAPIError
// @Failure      500 {object} ZatcaAPIError
// @Failure      401 {object} ErrorResponse
// @Router       /api/organizations/{id}/zatca [get]
func (h *OrganizationHandler) GetOrganizationZatca(c *gin.Context) {
	repo, id, ok := h.organizationZatcaTarget(c)
	if !ok {
		return
	}
	enabled, err := usecase.OrganizationZatcaEnabledByID(c.Request.Context(), repo, id)
	if err != nil {
		status := 500
		if errors.Is(err, pgx.ErrNoRows) {
			status = 404
		}
		c.JSON(status, utils.NewResponse(status, err.Error(), nil))
		return
	}
	c.JSON(200, utils.NewResponse(200, "Organization ZATCA setting", gin.H{"organization_id": id, "enabled": enabled, "source": "organizations.metadata.zatca.enabled"}))
}

// @Summary      Enable or disable ZATCA for an organization
// @Description  Atomically updates only metadata.zatca.enabled, preserving seller details and unrelated metadata. Both true and false are valid. Disabling pauses ZATCA operations without deleting credentials or signed documents; changes apply without a restart.
// @Tags         zatca, organizations
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        id path int true "Organization ID"
// @Accept       json
// @Param        body body ZatcaOrganizationSettingRequest true "Organization-wide switch; false is valid"
// @Success      200 {object} ZatcaOrganizationSettingResponse
// @Failure      400 {object} ZatcaAPIError
// @Failure      403 {object} ZatcaAPIError
// @Failure      404 {object} ZatcaAPIError
// @Failure      422 {object} ZatcaAPIError
// @Failure      500 {object} ZatcaAPIError
// @Failure      401 {object} ErrorResponse
// @Router       /api/organizations/{id}/zatca [put]
func (h *OrganizationHandler) UpdateOrganizationZatca(c *gin.Context) {
	repo, id, ok := h.organizationZatcaTarget(c)
	if !ok {
		return
	}
	var request ZatcaOrganizationSettingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, utils.NewResponse(400, "enabled must be a JSON boolean", nil))
		return
	}
	// Reject malformed existing metadata instead of replacing it or losing keys.
	if _, err := usecase.OrganizationZatcaEnabledByID(c.Request.Context(), repo, id); err != nil {
		status := 422
		if errors.Is(err, pgx.ErrNoRows) {
			status = 404
		}
		c.JSON(status, utils.NewResponse(status, err.Error(), nil))
		return
	}
	if _, err := repo.SetOrganizationZatcaEnabled(c.Request.Context(), id, *request.Enabled); err != nil {
		status := 500
		if errors.Is(err, pgx.ErrNoRows) {
			status = 404
		}
		c.JSON(status, utils.NewResponse(status, "failed to update organization ZATCA setting", nil))
		return
	}
	c.JSON(200, utils.NewResponse(200, "Organization ZATCA setting updated", gin.H{"organization_id": id, "enabled": *request.Enabled, "source": "organizations.metadata.zatca.enabled"}))
}
