package handler

import (
	"net/http"
	"strconv"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils"

	"github.com/gin-gonic/gin"
)

// PosTerminalsHandler holds the POS terminals use case.
type PosTerminalsHandler struct {
	useCase *usecase.PosTerminalsUseCase
}

// NewPosTerminalsHandler creates a new POS terminals handler.
func NewPosTerminalsHandler(uc *usecase.PosTerminalsUseCase) *PosTerminalsHandler {
	return &PosTerminalsHandler{useCase: uc}
}

func (h *PosTerminalsHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreatePOSTerminal handles POST /api/pos/terminals
// @Summary      Create POS terminal
// @Description  Creates a new POS terminal for a store
// @Tags         pos-terminals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        body          body      CreatePOSTerminalRequest  true  "Terminal payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/terminals [post]
func (h *PosTerminalsHandler) CreatePOSTerminal(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreatePOSTerminalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	input := &usecase.CreatePOSTerminalInput{
		StoreID:      req.StoreID,
		TerminalCode: req.TerminalCode,
		TerminalName: req.TerminalName,
		DeviceID:     req.DeviceID,
		IsActive:     req.IsActive,
		Metadata:     nil,
	}
	resp := h.useCase.CreatePOSTerminal(c.Request.Context(), input)
	c.JSON(resp.StatusCode, resp)
}

// GetPOSTerminal handles GET /api/pos/terminals/:id
// @Summary      Get POS terminal by ID
// @Description  Returns a single POS terminal by id
// @Tags         pos-terminals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Terminal ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/terminals/{id} [get]
func (h *PosTerminalsHandler) GetPOSTerminal(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid terminal id", nil))
		return
	}
	resp := h.useCase.GetPOSTerminal(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// GetPOSTerminalByCode handles GET /api/pos/stores/:store_id/terminals/code/:code
// @Summary      Get POS terminal by code
// @Description  Returns a POS terminal by store ID and terminal code
// @Tags         pos-terminals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      path      int     true  "Store ID"
// @Param        code          path      string  true  "Terminal code"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/stores/{store_id}/terminals/code/{code} [get]
func (h *PosTerminalsHandler) GetPOSTerminalByCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeID, err := strconv.ParseInt(c.Param("store_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "terminal code is required", nil))
		return
	}
	resp := h.useCase.GetPOSTerminalByCode(c.Request.Context(), int32(storeID), code)
	c.JSON(resp.StatusCode, resp)
}

// ListPOSTerminals handles GET /api/pos/terminals
// @Summary      List all POS terminals
// @Description  Returns all POS terminals
// @Tags         pos-terminals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/terminals [get]
func (h *PosTerminalsHandler) ListPOSTerminals(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListPOSTerminals(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListPOSTerminalsByStore handles GET /api/pos/stores/:store_id/terminals
// @Summary      List POS terminals by store
// @Description  Returns all POS terminals for a store
// @Tags         pos-terminals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      path      int     true  "Store ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/stores/{store_id}/terminals [get]
func (h *PosTerminalsHandler) ListPOSTerminalsByStore(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeID, err := strconv.ParseInt(c.Param("store_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}
	resp := h.useCase.ListPOSTerminalsByStore(c.Request.Context(), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// ListActivePOSTerminalsByStore handles GET /api/pos/stores/:store_id/terminals/active
// @Summary      List active POS terminals by store
// @Description  Returns active POS terminals for a store
// @Tags         pos-terminals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      path      int     true  "Store ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/stores/{store_id}/terminals/active [get]
func (h *PosTerminalsHandler) ListActivePOSTerminalsByStore(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeID, err := strconv.ParseInt(c.Param("store_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}
	resp := h.useCase.ListActivePOSTerminalsByStore(c.Request.Context(), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// UpdatePOSTerminal handles PUT /api/pos/terminals/:id
// @Summary      Update POS terminal
// @Description  Updates an existing POS terminal
// @Tags         pos-terminals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Terminal ID"
// @Param        body          body      UpdatePOSTerminalRequest  true  "Update payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/terminals/{id} [put]
func (h *PosTerminalsHandler) UpdatePOSTerminal(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid terminal id", nil))
		return
	}
	var req UpdatePOSTerminalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}
	input := &usecase.UpdatePOSTerminalInput{
		ID:           int32(id),
		TerminalName: req.TerminalName,
		DeviceID:     req.DeviceID,
		IsActive:     req.IsActive,
		Metadata:     nil,
	}
	resp := h.useCase.UpdatePOSTerminal(c.Request.Context(), input)
	c.JSON(resp.StatusCode, resp)
}

// DeletePOSTerminal handles DELETE /api/pos/terminals/:id
// @Summary      Delete POS terminal
// @Description  Deletes a POS terminal by id
// @Tags         pos-terminals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Terminal ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/terminals/{id} [delete]
func (h *PosTerminalsHandler) DeletePOSTerminal(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid terminal id", nil))
		return
	}
	resp := h.useCase.DeletePOSTerminal(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// TogglePOSTerminalActive handles PATCH /api/pos/terminals/:id/active
// @Summary      Toggle POS terminal active state
// @Description  Sets the active state of a POS terminal
// @Tags         pos-terminals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Terminal ID"
// @Param        body          body      TogglePOSTerminalActiveRequest  true  "Active state"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/terminals/{id}/active [patch]
func (h *PosTerminalsHandler) TogglePOSTerminalActive(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid terminal id", nil))
		return
	}
	var req TogglePOSTerminalActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}
	resp := h.useCase.TogglePOSTerminalActive(c.Request.Context(), int32(id), req.IsActive)
	c.JSON(resp.StatusCode, resp)
}
