package usecase

import (
	"context"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"
)

type NavigationUseCase struct {
	repo *repository.Queries
}

// NewNavigationUseCase creates a new navigation use case without a repository
// Repository will be injected per request via SetRepository
func NewNavigationUseCase() *NavigationUseCase {
	return &NavigationUseCase{}
}

// SetRepository sets the repository for this request
func (uc *NavigationUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// GetUserNavigation fetches complete navigation structure for a user
func (uc *NavigationUseCase) GetUserNavigation(ctx context.Context, userID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	navigation, err := uc.repo.GetCompleteNavigationByUserID(ctx, userID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "navigation fetched successfully", navigation)
}
