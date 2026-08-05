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

// UomPackagingTemplateHandler handles packaging template and level endpoints.
type UomPackagingTemplateHandler struct {
	useCase *usecase.UomPackagingTemplateUseCase
}

// NewUomPackagingTemplateHandler creates a new packaging template handler.
func NewUomPackagingTemplateHandler(uc *usecase.UomPackagingTemplateUseCase) *UomPackagingTemplateHandler {
	return &UomPackagingTemplateHandler{useCase: uc}
}

func (h *UomPackagingTemplateHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateUomPackagingTemplateRequest represents the body for creating a template.
type CreateUomPackagingTemplateRequest struct {
	OrganizationID int32  `json:"organization_id" binding:"required" example:"1"`
	Name           string `json:"name" binding:"required" example:"Bottle Pack"`
	Code           string `json:"code" binding:"required" example:"BTL_PK"`
	IsActive       bool   `json:"is_active" example:"true"`
}

// UpdateUomPackagingTemplateRequest represents the body for updating a template.
type UpdateUomPackagingTemplateRequest struct {
	Name     string `json:"name" binding:"required" example:"Bottle Pack Updated"`
	Code     string `json:"code" binding:"required" example:"BTL_PK_V2"`
	IsActive bool   `json:"is_active" example:"true"`
}

// CreateUomPackagingTemplateLevelRequest represents the body for creating a level.
type CreateUomPackagingTemplateLevelRequest struct {
	LevelOrder int32  `json:"level_order" binding:"required" example:"1"`
	UomID      int32  `json:"uom_id" binding:"required" example:"10"`
	Multiplier string `json:"multiplier" binding:"required" example:"12.00"`
}

// UpdateUomPackagingTemplateLevelRequest represents the body for updating a level.
type UpdateUomPackagingTemplateLevelRequest struct {
	LevelOrder int32  `json:"level_order" binding:"required" example:"1"`
	UomID      int32  `json:"uom_id" binding:"required" example:"10"`
	Multiplier string `json:"multiplier" binding:"required" example:"24.00"`
}

// CreateTemplate handles POST /api/uom-packaging-templates
// @Summary      Create a packaging template
// @Description  Create a new UOM packaging template
// @Tags         uom-packaging-templates
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  CreateUomPackagingTemplateRequest true  "Template payload"
// @Success      201   {object}  SuccessResponse
// @Failure      400   {object}  ErrorResponse
// @Router       /api/uom-packaging-templates [post]
func (h *UomPackagingTemplateHandler) CreateTemplate(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateUomPackagingTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.CreateTemplate(c.Request.Context(), req.OrganizationID, req.Name, req.Code, req.IsActive)
	c.JSON(resp.StatusCode, resp)
}

// GetTemplate handles GET /api/uom-packaging-templates/:id
// @Summary      Get packaging template by ID
// @Description  Retrieve a packaging template by its ID
// @Tags         uom-packaging-templates
// @Param        id  path  string  true  "Template ID"
// @Success      200 {object}  SuccessResponse
// @Router       /api/uom-packaging-templates/{id} [get]
func (h *UomPackagingTemplateHandler) GetTemplate(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetTemplate(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// ListTemplates handles GET /api/uom-packaging-templates
// @Summary      List packaging templates
// @Description  List templates for an organization
// @Tags         uom-packaging-templates
// @Param        organization_id  query  int  true  "Organization ID"
// @Success      200 {object}  SuccessResponse
// @Router       /api/uom-packaging-templates [get]
func (h *UomPackagingTemplateHandler) ListTemplates(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid organization_id", nil))
		return
	}

	resp := h.useCase.ListTemplates(c.Request.Context(), int32(orgID))
	c.JSON(resp.StatusCode, resp)
}

// UpdateTemplate handles PUT /api/uom-packaging-templates/:id
// @Summary      Update a packaging template
// @Tags         uom-packaging-templates
// @Param        id    path  string                             true  "Template ID"
// @Param        body  body  UpdateUomPackagingTemplateRequest  true  "Template payload"
// @Success      200   {object}  SuccessResponse
// @Router       /api/uom-packaging-templates/{id} [put]
func (h *UomPackagingTemplateHandler) UpdateTemplate(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	var req UpdateUomPackagingTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.UpdateTemplate(c.Request.Context(), id, req.Name, req.Code, req.IsActive)
	c.JSON(resp.StatusCode, resp)
}

// DeleteTemplate handles DELETE /api/uom-packaging-templates/:id
// @Summary      Delete a packaging template
// @Tags         uom-packaging-templates
// @Param        id  path  string  true  "Template ID"
// @Success      200 {object}  SuccessResponse
// @Router       /api/uom-packaging-templates/{id} [delete]
func (h *UomPackagingTemplateHandler) DeleteTemplate(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeleteTemplate(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// CreateLevel handles POST /api/uom-packaging-templates/:id/levels
// @Summary      Create a template level
// @Tags         uom-packaging-templates
// @Param        id    path  int                                     true  "Template ID"
// @Param        body  body  CreateUomPackagingTemplateLevelRequest  true  "Level payload"
// @Success      201   {object}  SuccessResponse
// @Router       /api/uom-packaging-templates/{id}/levels [post]
func (h *UomPackagingTemplateHandler) CreateLevel(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid template_id", nil))
		return
	}

	var req CreateUomPackagingTemplateLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.CreateLevel(c.Request.Context(), int32(id), req.LevelOrder, req.UomID, req.Multiplier)
	c.JSON(resp.StatusCode, resp)
}

// ListLevels handles GET /api/uom-packaging-templates/:id/levels
// @Summary      List levels for a template
// @Tags         uom-packaging-templates
// @Param        id  path  int  true  "Template ID"
// @Success      200 {object}  SuccessResponse
// @Router       /api/uom-packaging-templates/{id}/levels [get]
func (h *UomPackagingTemplateHandler) ListLevels(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid template_id", nil))
		return
	}

	resp := h.useCase.ListLevels(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// UpdateLevel handles PUT /api/uom-packaging-template-levels/:level_id
// @Summary      Update a template level
// @Tags         uom-packaging-templates
// @Param        level_id  path  string                                  true  "Level ID"
// @Param        body      body  UpdateUomPackagingTemplateLevelRequest  true  "Level payload"
// @Success      200       {object}  SuccessResponse
// @Router       /api/uom-packaging-template-levels/{level_id} [put]
func (h *UomPackagingTemplateHandler) UpdateLevel(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	levelID := c.Param("level_id")
	var req UpdateUomPackagingTemplateLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.UpdateLevel(c.Request.Context(), levelID, req.LevelOrder, req.UomID, req.Multiplier)
	c.JSON(resp.StatusCode, resp)
}

// DeleteLevel handles DELETE /api/uom-packaging-template-levels/:level_id
// @Summary      Delete a template level
// @Tags         uom-packaging-templates
// @Param        level_id  path  string  true  "Level ID"
// @Success      200       {object}  SuccessResponse
// @Router       /api/uom-packaging-template-levels/{level_id} [delete]
func (h *UomPackagingTemplateHandler) DeleteLevel(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	levelID := c.Param("level_id")
	resp := h.useCase.DeleteLevel(c.Request.Context(), levelID)
	c.JSON(resp.StatusCode, resp)
}

// GetTemplateWithLevels handles GET /api/uom-packaging-templates/:id/full
// @Summary      Get packaging template with all its levels
// @Tags         uom-packaging-templates
// @Param        id  path  string  true  "Template ID"
// @Success      200 {object}  SuccessResponse
// @Router       /api/uom-packaging-templates/{id}/full [get]
func (h *UomPackagingTemplateHandler) GetTemplateWithLevels(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetTemplateWithLevels(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// CreateTemplatePipelineRequest represents the payload for bulk pipeline creation.
type CreateTemplatePipelineRequest struct {
	BaseUomCode      string `json:"base_uom_code" binding:"required" example:"KG"`
	BaseUomName      string `json:"base_uom_name" binding:"required" example:"Kilogram"`
	BaseUomType      string `json:"base_uom_type" binding:"required" example:"packaging"`
	BaseUomDecimals  int32  `json:"base_uom_decimals" example:"0"`
	BaseUomActive    bool   `json:"base_uom_active" example:"true"`
	OrganizationID   int32  `json:"organization_id" binding:"required" example:"1"`
	TemplateName     string `json:"template_name" binding:"required" example:"Template A"`
	TemplateCode     string `json:"template_code" binding:"required" example:"1-24-12"`
	TemplateActive   bool   `json:"template_active" example:"true"`
	Tier1Multiplier  string `json:"tier1_multiplier" binding:"required" example:"1.000000"`
	Tier2UomCode     string `json:"tier2_uom_code" binding:"required" example:"PKT"`
	Tier2Multiplier  string `json:"tier2_multiplier" binding:"required" example:"24.000000"`
	Tier3UomCode     string `json:"tier3_uom_code" binding:"required" example:"CAN"`
	Tier3Multiplier  string `json:"tier3_multiplier" binding:"required" example:"12.000000"`
}

// CreateTemplatePipeline handles POST /api/uom-packaging-templates/pipeline
// @Summary      Create UOM packaging template pipeline
// @Description  Create a template, its base UOM, and its level hierarchy in a single query execution
// @Tags         uom-packaging-templates
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  CreateTemplatePipelineRequest true  "Pipeline payload"
// @Success      201   {object}  SuccessResponse
// @Failure      400   {object}  ErrorResponse
// @Router       /api/uom-packaging-templates/pipeline [post]
func (h *UomPackagingTemplateHandler) CreateTemplatePipeline(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateTemplatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.CreateTemplatePipeline(c.Request.Context(), usecase.CreateUomPackagingTemplatePipelineInput{
		BaseUomCode:     req.BaseUomCode,
		BaseUomName:     req.BaseUomName,
		BaseUomType:     req.BaseUomType,
		BaseUomDecimals: req.BaseUomDecimals,
		BaseUomActive:   req.BaseUomActive,
		OrganizationID:  req.OrganizationID,
		TemplateName:    req.TemplateName,
		TemplateCode:    req.TemplateCode,
		TemplateActive:  req.TemplateActive,
		Tier1Multiplier: req.Tier1Multiplier,
		Tier2UomCode:    req.Tier2UomCode,
		Tier2Multiplier: req.Tier2Multiplier,
		Tier3UomCode:    req.Tier3UomCode,
		Tier3Multiplier: req.Tier3Multiplier,
	})
	c.JSON(resp.StatusCode, resp)
}

// GetTemplatesByUomID handles GET /api/uom-packaging-templates/by-uom/:uom_id
// @Summary      Get packaging templates by UOM ID
// @Description  Retrieve all templates and their levels that use the specified UOM ID
// @Tags         uom-packaging-templates
// @Param        uom_id  path  string  true  "UOM ID"
// @Success      200 {object}  SuccessResponse
// @Router       /api/uom-packaging-templates/by-uom/{uom_id} [get]
func (h *UomPackagingTemplateHandler) GetTemplatesByUomID(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	uomID := c.Param("uom_id")
	resp := h.useCase.GetTemplatesByUomID(c.Request.Context(), uomID)
	c.JSON(resp.StatusCode, resp)
}

