package usecase

import (
	"context"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils" // Assuming your NewResponse is here

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
		return utils.NewResponse(utils.CodeNotFound, "invalid credentials", nil)
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

	// Generate JWT token - convert user ID from int32 to string and include organization ID
	userIDStr := strconv.FormatInt(int64(user.ID), 10)
	token, err := middleware.GenerateJWTToken(userIDStr, userLogin, user.OrganizationID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to generate token", nil)
	}

	return utils.NewResponse(utils.CodeOK, "login successful", token)
}

// AuthorizeActionOutput holds the data returned from a successful supervisor override.
type AuthorizeActionOutput struct {
	AuthorizationToken string    `json:"authorization_token"`
	AuthorizedByUserID int32     `json:"authorized_by_user_id"`
	AuthorizedByLogin  string    `json:"authorized_by_login"`
	PermissionCode     string    `json:"permission_code"`
	ExpiresAt          time.Time `json:"expires_at"`
}

// AuthorizeAction implements the supervisor override ("manager authorization") flow.
//
// Steps:
//  1. Authenticate the supervisor by username + password (bcrypt)
//  2. Verify the supervisor has the required permissionCode via role-based permission check
//  3. Mint a short-lived (5-minute) authorization token scoped to that permission
//
// The caller (POS) then sends the returned token in the X-Authorization-Token header
// when invoking the restricted action endpoint (e.g. POST /api/pos/returns).
func (uc *AuthUseCase) AuthorizeAction(ctx context.Context, supervisorLogin, password, permissionCode string, cashierID *int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if supervisorLogin == "" {
		return utils.NewResponse(utils.CodeBadReq, "supervisor_login cannot be empty", nil)
	}
	if password == "" {
		return utils.NewResponse(utils.CodeBadReq, "supervisor_password cannot be empty", nil)
	}
	if permissionCode == "" {
		return utils.NewResponse(utils.CodeBadReq, "permission_code cannot be empty", nil)
	}

	// Step 1: Authenticate supervisor credentials
	supervisor, err := uc.repo.GetUserByUsername(ctx, supervisorLogin)
	if err != nil {
		// Use generic message — do not reveal whether the user exists
		return utils.NewResponse(utils.CodeUnauthorized, "invalid supervisor credentials", nil)
	}

	if !supervisor.IsActive.Bool || !supervisor.IsActive.Valid {
		return utils.NewResponse(utils.CodeUnauthorized, "supervisor account is inactive", nil)
	}

	if supervisor.PasswordHash == "" {
		return utils.NewResponse(utils.CodeUnauthorized, "invalid supervisor credentials", nil)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(supervisor.PasswordHash), []byte(password)); err != nil {
		return utils.NewResponse(utils.CodeUnauthorized, "invalid supervisor credentials", nil)
	}

	// Step 2: Verify the supervisor holds the required permission (via their assigned roles)
	hasPermission, err := uc.repo.CheckUserHasPermission(ctx, repository.CheckUserHasPermissionParams{
		UserID: supervisor.ID,
		Code:   permissionCode,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, "permission check failed: "+err.Error(), nil)
	}
	if !hasPermission {
		return utils.NewResponse(utils.CodeForbidden, "supervisor does not have permission: "+permissionCode, nil)
	}

	// Step 3: Generate short-lived authorization token (5-minute window)
	token, err := middleware.GenerateAuthorizationToken(supervisor.ID, supervisorLogin, permissionCode, cashierID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to generate authorization token", nil)
	}

	expiresAt := time.Now().Add(5 * time.Minute)

	return utils.NewResponse(utils.CodeOK, "action authorized", AuthorizeActionOutput{
		AuthorizationToken: token,
		AuthorizedByUserID: supervisor.ID,
		AuthorizedByLogin:  supervisorLogin,
		PermissionCode:     permissionCode,
		ExpiresAt:          expiresAt,
	})
}
