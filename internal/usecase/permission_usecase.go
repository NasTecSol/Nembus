package usecase

import (
	"context"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"
)

type PermissionUseCase struct {
	repo *repository.Queries
}

// NewPermissionUseCase creates a new permission use case without a repository
// Repository will be injected per request via SetRepository
func NewPermissionUseCase() *PermissionUseCase {
	return &PermissionUseCase{}
}

// SetRepository sets the repository for this request
func (uc *PermissionUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// CheckUserSubmenuPermission checks if a user has access to a submenu by submenu code
func (uc *PermissionUseCase) CheckUserSubmenuPermission(ctx context.Context, userID int32, submenuCode string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if submenuCode == "" {
		return utils.NewResponse(utils.CodeBadReq, "submenu code cannot be empty", nil)
	}

	// Check if user has access to submenu by code directly
	hasAccess, err := uc.repo.CheckUserHasSubmenuAccessByCode(ctx, repository.CheckUserHasSubmenuAccessByCodeParams{
		UserID: userID,
		Code:   submenuCode,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "permission check completed", map[string]interface{}{
		"has_access":    hasAccess,
		"submenu_code":  submenuCode,
		"user_id":       userID,
	})
}
