package usecase

import (
	"context"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/gin-gonic/gin"
)

type ImageUseCase struct {
	repo *repository.Queries
}

// NewImageUseCase creates a new Image use case
func NewImageUseCase() *ImageUseCase {
	return &ImageUseCase{}
}

// SetRepository sets the repository for tenant-specific operations
func (uc *ImageUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// UploadModuleImage saves image under:
// images/<tenant>/<module_name>/
func (uc *ImageUseCase) UploadModuleImage(
	ctx context.Context,
	moduleName string,
	tenantID string,
	file *multipart.FileHeader,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	// 🔽 Normalize module name
	moduleName = strings.ToLower(moduleName)

	// Folder: images/<tenant>/<module_name>/
	folderPath := filepath.Join(
		"images",
		tenantID,
		moduleName,
	)

	// Create directory if not exists
	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create image directory", nil)
	}

	// Final file path
	filePath := filepath.Join(folderPath, file.Filename)
	// Convert Windows backslashes to forward slashes for URL
	urlPath := strings.ReplaceAll(filePath, "\\", "/")

	// ❗ DO NOT SAVE FILE HERE (handler will do it)

	return utils.NewResponse(
		utils.CodeOK,
		"image path prepared successfully",
		gin.H{
			"module": moduleName,
			"path":   urlPath,
		},
	)
}
