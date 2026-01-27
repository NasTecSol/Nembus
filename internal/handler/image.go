package handler

import (
	"net/http"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils"

	"github.com/gin-gonic/gin"
)

// ImageHandler holds the use case
type ImageHandler struct {
	useCase *usecase.ImageUseCase
}

// NewImageHandler creates a new handler instance
func NewImageHandler(uc *usecase.ImageUseCase) *ImageHandler {
	return &ImageHandler{
		useCase: uc,
	}
}

// getRepositoryFromContext extracts repository from Gin context
func (h *ImageHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// UploadModuleImage handles POST /api/images/:module
// @Summary      Upload module image
// @Description  Upload an image for a specific module (tenant-based)
// @Tags         images
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header   string  true  "Tenant identifier"
// @Param        Authorization  header   string  true  "Bearer token"
// @Param        module         path     string  true  "Module name"
// @Param        image          formData file    true  "Image file"
// @Success      201  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/images/{module} [post]
func (h *ImageHandler) UploadModuleImage(c *gin.Context) {
	// Get tenant repository
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	// Get tenant ID (from middleware)
	tenantID := c.GetHeader("x-tenant-id") // <-- Read from header
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"tenant_id header is required",
			nil,
		))
		return
	}

	// Module name from path
	moduleName := c.Param("module")
	if moduleName == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"module name is required",
			nil,
		))
		return
	}

	// Get image file
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"image file is required",
			nil,
		))
		return
	}

	// Call usecase (prepare path + validation)
	resp := h.useCase.UploadModuleImage(
		c.Request.Context(),
		moduleName,
		tenantID,
		file,
	)

	if resp.StatusCode != utils.CodeOK {
		c.JSON(resp.StatusCode, resp)
		return
	}

	// Extract path from response
	data := resp.Data.(gin.H)
	filePath := data["path"].(string)

	// Save file (ONLY place where saving happens)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			utils.CodeError,
			"failed to save image",
			nil,
		))
		return
	}

	// Final success response
	c.JSON(http.StatusCreated, utils.NewResponse(
		utils.CodeCreated,
		"module image uploaded successfully",
		gin.H{
			"module": moduleName,
			"path":   filePath,
		},
	))
}
