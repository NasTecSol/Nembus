package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/NasTecSol/nembus-client/utils"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/printing"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/gin-gonic/gin"
)

type ReceiptTemplateHandler struct{}

type ReceiptTemplateSaveRequest struct {
	Name     string                         `json:"name" example:"Standard receipt"`
	Type     string                         `json:"type" example:"receipt"`
	Language string                         `json:"language" example:"en"`
	Config   printing.ReceiptTemplateConfig `json:"config"`
	IsActive *bool                          `json:"is_active,omitempty" example:"true"`
}

type ReceiptTemplateDefaultRequest struct {
	TemplateID int64 `json:"template_id" example:"12"`
}

type StoreReceiptTemplateDefaultRequest struct {
	TemplateID *int64 `json:"template_id" example:"12"` // null means inherit the organization default
	Type       string `json:"type" example:"receipt"`
	Language   string `json:"language" example:"en"`
}

func NewReceiptTemplateHandler() *ReceiptTemplateHandler { return &ReceiptTemplateHandler{} }
func receiptTemplateRepo(c *gin.Context) (*repository.Queries, bool) {
	r, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		return nil, false
	}
	return r, true
}
func receiptOrgID(c *gin.Context) (int32, bool) {
	n, e := strconv.ParseInt(c.Param("id"), 10, 32)
	if e != nil || n <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid organization_id is required"})
		return 0, false
	}
	return int32(n), true
}
func receiptTemplateError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

// List returns active and inactive templates owned by an organization.
// @Summary List organization receipt templates
// @Tags receipt-templates
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param id path int true "Organization ID"
// @Success 200 {object} repository.Response
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/organizations/{id}/receipt-templates [get]
func (h *ReceiptTemplateHandler) List(c *gin.Context) {
	r, ok := receiptTemplateRepo(c)
	if !ok {
		return
	}
	org, ok := receiptOrgID(c)
	if !ok {
		return
	}
	v, err := r.ListReceiptTemplates(c, org)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, utils.NewResponse(utils.CodeOK, "receipt templates", v))
}

// Create adds a receipt template to an organization.
// @Summary Create an organization receipt template
// @Tags receipt-templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param id path int true "Organization ID"
// @Param body body ReceiptTemplateSaveRequest true "Template name, type, language, layout config, and active state"
// @Success 200 {object} repository.Response
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/organizations/{id}/receipt-templates [post]
func (h *ReceiptTemplateHandler) Create(c *gin.Context) { h.save(c, 0) }

// Update changes a template without changing its organization.
// @Summary Update an organization receipt template
// @Tags receipt-templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param id path int true "Organization ID"
// @Param template_id path int true "Receipt template ID"
// @Param body body ReceiptTemplateSaveRequest true "Updated template name, type, language, layout config, and active state"
// @Success 200 {object} repository.Response
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/organizations/{id}/receipt-templates/{template_id} [put]
func (h *ReceiptTemplateHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("template_id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid template_id is required"})
		return
	}
	h.save(c, id)
}

func (h *ReceiptTemplateHandler) save(c *gin.Context, id int64) {
	r, ok := receiptTemplateRepo(c)
	if !ok {
		return
	}
	org, ok := receiptOrgID(c)
	if !ok {
		return
	}
	var body ReceiptTemplateSaveRequest
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "invalid template request"})
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	body.Type = strings.TrimSpace(body.Type)
	body.Language = strings.TrimSpace(body.Language)
	body.Type = strings.ToLower(body.Type)
	body.Language = strings.ToLower(body.Language)
	if body.Type == "" {
		body.Type = "receipt"
	}
	if body.Language == "" {
		body.Language = "en"
	}
	if body.Name == "" {
		c.JSON(400, gin.H{"error": "name and template config are required"})
		return
	}
	configJSON, marshalErr := json.Marshal(body.Config)
	if marshalErr != nil {
		receiptTemplateError(c, marshalErr)
		return
	}
	cfg, err := printing.ParseReceiptTemplateConfig(configJSON)
	if err != nil {
		receiptTemplateError(c, err)
		return
	}
	normalized, _ := json.Marshal(cfg)
	active := true
	if body.IsActive != nil {
		active = *body.IsActive
	}
	v, err := r.SaveReceiptTemplate(c, org, id, body.Name, body.Type, body.Language, normalized, active)
	if err != nil {
		receiptTemplateError(c, err)
		return
	}
	c.JSON(200, utils.NewResponse(utils.CodeOK, "receipt template saved", v))
}

// SetOrganizationDefault sets the fallback template inherited by stores.
// @Summary Set organization receipt template default
// @Tags receipt-templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param id path int true "Organization ID"
// @Param body body ReceiptTemplateDefaultRequest true "Active template to use as organization default"
// @Success 200 {object} repository.Response
// @Failure 400 {object} ErrorResponse
// @Router /api/organizations/{id}/receipt-template-default [put]
func (h *ReceiptTemplateHandler) SetOrganizationDefault(c *gin.Context) {
	r, ok := receiptTemplateRepo(c)
	if !ok {
		return
	}
	org, ok := receiptOrgID(c)
	if !ok {
		return
	}
	var b struct {
		TemplateID int64 `json:"template_id"`
	}
	if c.ShouldBindJSON(&b) != nil || b.TemplateID <= 0 {
		c.JSON(400, gin.H{"error": "template_id is required"})
		return
	}
	if err := r.SetOrganizationReceiptDefault(c, org, b.TemplateID); err != nil {
		receiptTemplateError(c, err)
		return
	}
	c.JSON(200, utils.NewResponse(utils.CodeOK, "organization receipt default saved", nil))
}

// SetStoreDefault sets a store override; template_id null clears the override and inherits the organization default.
// @Summary Set or inherit a store receipt template default
// @Tags receipt-templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param id path int true "Store ID"
// @Param body body StoreReceiptTemplateDefaultRequest true "Template ID, or null with type/language to inherit"
// @Success 200 {object} repository.Response
// @Failure 400 {object} ErrorResponse
// @Router /api/stores/{id}/receipt-template-default [put]
func (h *ReceiptTemplateHandler) SetStoreDefault(c *gin.Context) {
	r, ok := receiptTemplateRepo(c)
	if !ok {
		return
	}
	store, e := strconv.ParseInt(c.Param("id"), 10, 64)
	if e != nil || store <= 0 {
		c.JSON(400, gin.H{"error": "valid store_id is required"})
		return
	}
	var b struct {
		TemplateID *int64 `json:"template_id"`
		Type       string `json:"type"`
		Language   string `json:"language"`
	}
	if c.ShouldBindJSON(&b) != nil || (b.TemplateID != nil && *b.TemplateID <= 0) {
		c.JSON(400, gin.H{"error": "template_id must be a positive ID or null to inherit the organization default"})
		return
	}
	b.Type = strings.ToLower(strings.TrimSpace(b.Type))
	b.Language = strings.ToLower(strings.TrimSpace(b.Language))
	if err := r.SetStoreReceiptDefault(c, store, b.TemplateID, b.Type, b.Language); err != nil {
		receiptTemplateError(c, err)
		return
	}
	c.JSON(200, utils.NewResponse(utils.CodeOK, "store receipt default saved", nil))
}

// GetStoreDefault returns the effective template and whether it comes from the store or organization.
// @Summary Get effective store receipt template
// @Tags receipt-templates
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param id path int true "Store ID"
// @Param type query string false "Template type (defaults to receipt)"
// @Param language query string false "Template language (defaults to en)"
// @Success 200 {object} repository.Response
// @Failure 404 {object} ErrorResponse
// @Router /api/stores/{id}/receipt-template-default [get]
func (h *ReceiptTemplateHandler) GetStoreDefault(c *gin.Context) {
	r, ok := receiptTemplateRepo(c)
	if !ok {
		return
	}
	store, e := strconv.ParseInt(c.Param("id"), 10, 64)
	if e != nil || store <= 0 {
		c.JSON(400, gin.H{"error": "valid store_id is required"})
		return
	}
	orgID, err := r.GetStoreOrganizationID(c, store)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
		return
	}
	templateType := strings.TrimSpace(c.Query("type"))
	if templateType == "" {
		templateType = "receipt"
	}
	language := strings.TrimSpace(c.Query("language"))
	if language == "" {
		language = "en"
	}
	v, source, err := r.ResolveReceiptTemplate(c, orgID, store, templateType, language)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "receipt template default not found"})
		return
	}
	c.JSON(200, utils.NewResponse(utils.CodeOK, "effective receipt template", gin.H{"template": v, "source": source}))
}
