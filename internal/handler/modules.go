package handler

import (
	"net/http"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils" // Assuming your NewResponse is here

	"github.com/gin-gonic/gin"
)

type ModuleHandler struct {
	useCase *usecase.ModuleUseCase
}

func NewModuleHandler(uc *usecase.ModuleUseCase) *ModuleHandler {
	return &ModuleHandler{
		useCase: uc,
	}
}

// getRepositoryFromContext extracts repository from Gin context
func (h *ModuleHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateModule handles POST /modules
// @Summary      Create a new module
// @Description  Create a new module with required name and code
// @Tags         modules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string             true  "Tenant identifier"
// @Param        Authorization header    string             true  "Bearer token"
// @Param        module        body      CreateModuleRequest true  "Module data"
// @Success      201           {object}  ModuleResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/modules [post]
func (h *ModuleHandler) CreateModule(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req struct {
		Name         string  `json:"name" binding:"required"`
		Code         string  `json:"code" binding:"required"`
		Description  *string `json:"description,omitempty"`
		Icon         *string `json:"icon,omitempty"`
		IsActive     bool    `json:"is_active"`
		DisplayOrder int32   `json:"display_order"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	resp := h.useCase.CreateModule(c.Request.Context(), req.Name, req.Code, req.Description, req.Icon, req.IsActive, req.DisplayOrder)
	c.JSON(resp.StatusCode, resp)
}

// GetModule handles GET /modules/:id
// @Summary      Get module by ID
// @Description  Retrieve a specific module by its ID
// @Tags         modules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Module ID"
// @Success      200           {object}  ModuleResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/modules/{id} [get]
func (h *ModuleHandler) GetModule(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetModule(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// GetModuleByCode handles GET /modules/code/:code
// @Summary      Get module by code
// @Description  Retrieve a specific module by its code
// @Tags         modules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        code          path      string  true  "Module code"
// @Success      200           {object}  ModuleResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/modules/code/{code} [get]
func (h *ModuleHandler) GetModuleByCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	code := c.Param("code")
	resp := h.useCase.GetModuleByCode(c.Request.Context(), code)
	c.JSON(resp.StatusCode, resp)
}

// ListModules handles GET /modules
// @Summary      List all modules
// @Description  Retrieve a list of all modules, optionally filtered by active status
// @Tags         modules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        is_active     query     bool    false "Filter by active status"
// @Success      200           {array}   ModuleResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/modules [get]
func (h *ModuleHandler) ListModules(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	isActiveStr := c.Query("is_active")
	var isActive *bool
	if isActiveStr == "true" {
		t := true
		isActive = &t
	} else if isActiveStr == "false" {
		f := false
		isActive = &f
	}

	resp := h.useCase.ListModules(c.Request.Context(), isActive)
	c.JSON(resp.StatusCode, resp)
}

// UpdateModule handles PUT /modules/:id
// @Summary      Update a module
// @Description  Update an existing module by its ID
// @Tags         modules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string              true  "Tenant identifier"
// @Param        Authorization header    string              true  "Bearer token"
// @Param        id            path      string              true  "Module ID"
// @Param        module        body      UpdateModuleRequest true  "Module data"
// @Success      200           {object}  ModuleResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/modules/{id} [put]
func (h *ModuleHandler) UpdateModule(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req struct {
		Name         *string `json:"name,omitempty"`
		Description  *string `json:"description,omitempty"`
		Icon         *string `json:"icon,omitempty"`
		IsActive     *bool   `json:"is_active,omitempty"`
		DisplayOrder *int32  `json:"display_order,omitempty"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	resp := h.useCase.UpdateModule(c.Request.Context(), id, req.Name, req.Description, req.Icon, req.IsActive, req.DisplayOrder)
	c.JSON(resp.StatusCode, resp)
}

// DeleteModule handles DELETE /modules/:id
// @Summary      Delete a module
// @Description  Delete a module by its ID
// @Tags         modules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Module ID"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/modules/{id} [delete]
func (h *ModuleHandler) DeleteModule(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeleteModule(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// GetNavigationHierarchy handles GET /modules/navigation
// @Summary      Get navigation hierarchy
// @Description  Retrieve full navigation hierarchy (modules, menus, submenus)
// @Tags         modules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/modules/navigation [get]
func (h *ModuleHandler) GetNavigationHierarchy(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.GetNavigationHierarchy(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}
