package usecase

import (
	"context"
	"errors"
	"strconv"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"

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
func (uc *AuthUseCase) Login(ctx context.Context, userLogin, password string) (string, error) {
	if uc.repo == nil {
		return "", errors.New("repository not set")
	}

	if userLogin == "" {
		return "", errors.New("user_login cannot be empty")
	}

	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	// Get user by username
	user, err := uc.repo.GetUserByUsername(ctx, userLogin)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive.Bool || !user.IsActive.Valid {
		return "", errors.New("user account is inactive")
	}

	// Check if password_hash exists
	if user.PasswordHash == "" {
		return "", errors.New("password not set for this user")
	}

	// Compare password with hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token - convert user ID from int32 to string
	userIDStr := strconv.FormatInt(int64(user.ID), 10)
	token, err := middleware.GenerateJWTToken(userIDStr, userLogin)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}
