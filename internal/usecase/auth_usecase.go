package usecase

import (
	"context"
	"strconv"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/utils" // Assuming your NewResponse is here

	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	repo *repository.Queries
}

// NewAuthUseCase creates a new auth use case without a repository
// Repository will be injected per request via SetRepository
func NewAuthUseCase() *AuthUseCase {
	return &AuthUseCase{}
}

// SetRepository sets the repository for this request
func (uc *AuthUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// Login authenticates a user and returns a JWT token
func (uc *AuthUseCase) Login(ctx context.Context, userLogin, password string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if userLogin == "" {
		return utils.NewResponse(utils.CodeBadReq, "user_login cannot be empty", nil)
	}

	if password == "" {
		return utils.NewResponse(utils.CodeBadReq, "password cannot be empty", nil)
	}

	// Get user by username
	user, err := uc.repo.GetUserByUsername(ctx, userLogin)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "invalid credentials", nil)
	}

	// Check if user is active
	if !user.IsActive.Bool || !user.IsActive.Valid {
		return utils.NewResponse(utils.CodeError, "user account is inactive", nil)
	}

	// Check if password_hash exists
	if user.PasswordHash == "" {
		return utils.NewResponse(utils.CodeError, "password not set for this user", nil)
	}

	// Compare password with hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return utils.NewResponse(utils.CodeError, "invalid credentials", nil)
	}

	// Generate JWT token - convert user ID from int32 to string
	userIDStr := strconv.FormatInt(int64(user.ID), 10)
	token, err := middleware.GenerateJWTToken(userIDStr, userLogin)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to generate token", nil)
	}

	return utils.NewResponse(utils.CodeOK, "login successful", token)
}
