package usecase

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/biometrics"
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

type BiometricLoginOutput struct {
	Token          string `json:"token"`
	UserID         int32  `json:"user_id"`
	UserLogin      string `json:"user_login"`
	OrganizationID int32  `json:"organization_id"`
	MatchedFinger  int16  `json:"matched_finger_index"`
}

// LoginWithBiometric authenticates a user via fingerprint in either 1:1 mode (when userLogin is provided)
// or 1:N Touch-Only mode (when userLogin is empty).
func (uc *AuthUseCase) LoginWithBiometric(ctx context.Context, fmdBase64 string, userLogin string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	probeFMD, err := biometrics.ValidateAndDecodeFMD(fmdBase64)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, err.Error(), nil)
	}
	if !biometrics.IsEngineAvailable() {
		return utils.NewResponse(utils.CodeError, biometrics.ErrEngineUnavailable.Error(), nil)
	}

	var matchedUserID int32
	var matchedUsername string
	var matchedOrgID int32
	var matchedFinger int16

	if userLogin != "" {
		// 1:1 Verification Mode
		user, err := uc.repo.GetUserByUsername(ctx, userLogin)
		if err != nil {
			return utils.NewResponse(utils.CodeUnauthorized, "invalid biometric credentials", nil)
		}
		if !user.IsActive.Bool || !user.IsActive.Valid {
			return utils.NewResponse(utils.CodeUnauthorized, "user account is inactive", nil)
		}
		templates, err := uc.repo.GetActiveUserBiometricTemplatesWithFMD(ctx, user.ID)
		if err != nil || len(templates) == 0 {
			return utils.NewResponse(utils.CodeUnauthorized, "no biometric fingerprints enrolled for this user", nil)
		}
		matched := false
		for _, t := range templates {
			candBytes, decErr := biometrics.ValidateAndDecodeFMD(t.FmdBase64)
			if decErr != nil {
				continue
			}
			_, ok, cmpErr := biometrics.Compare(probeFMD, candBytes, biometrics.DefaultMaxFalseMatchScore)
			if cmpErr == nil && ok {
				matched = true
				matchedUserID = user.ID
				matchedUsername = user.Username
				matchedOrgID = user.OrganizationID
				matchedFinger = t.FingerIndex
				break
			}
		}
		if !matched {
			return utils.NewResponse(utils.CodeUnauthorized, "fingerprint verification did not match", nil)
		}
	} else {
		// 1:N Touch-Only Identification Mode
		allTemplates, err := uc.repo.ListAllActiveBiometricTemplates(ctx)
		if err != nil || len(allTemplates) == 0 {
			return utils.NewResponse(utils.CodeUnauthorized, "no enrolled biometric fingerprints found", nil)
		}
		candidates := make([][]byte, 0, len(allTemplates))
		validRows := make([]repository.ListAllActiveBiometricTemplatesRow, 0, len(allTemplates))
		for _, r := range allTemplates {
			if b, decErr := biometrics.ValidateAndDecodeFMD(r.FmdBase64); decErr == nil {
				candidates = append(candidates, b)
				validRows = append(validRows, r)
			}
		}
		bestIdx, _, idErr := biometrics.Identify(probeFMD, candidates, biometrics.DefaultMaxFalseMatchScore)
		if idErr != nil || bestIdx < 0 || bestIdx >= len(validRows) {
			return utils.NewResponse(utils.CodeUnauthorized, "fingerprint did not match any active user", nil)
		}
		row := validRows[bestIdx]
		matchedUserID = row.UserID
		matchedUsername = row.Username
		matchedOrgID = row.OrganizationID
		matchedFinger = row.FingerIndex
	}

	userIDStr := strconv.FormatInt(int64(matchedUserID), 10)
	token, err := middleware.GenerateJWTToken(userIDStr, matchedUsername, matchedOrgID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to generate token", nil)
	}

	log.Printf("[BIOMETRICS] Biometric login succeeded for user_id=%d (%s) via finger=%d",
		matchedUserID, matchedUsername, matchedFinger)

	return utils.NewResponse(utils.CodeOK, "biometric login successful", BiometricLoginOutput{
		Token:          token,
		UserID:         matchedUserID,
		UserLogin:      matchedUsername,
		OrganizationID: matchedOrgID,
		MatchedFinger:  matchedFinger,
	})
}

// AuthorizeActionWithBiometric authenticates a supervisor via fingerprint (1:1 if supervisorLogin is provided,
// or 1:N touch-only across enrolled users) and verifies the matched user holds permissionCode before minting
// the 5-minute X-Authorization-Token.
func (uc *AuthUseCase) AuthorizeActionWithBiometric(
	ctx context.Context,
	fmdBase64 string,
	supervisorLogin string,
	permissionCode string,
	cashierID *int32,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if permissionCode == "" {
		return utils.NewResponse(utils.CodeBadReq, "permission_code cannot be empty", nil)
	}
	probeFMD, err := biometrics.ValidateAndDecodeFMD(fmdBase64)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, err.Error(), nil)
	}
	if !biometrics.IsEngineAvailable() {
		return utils.NewResponse(utils.CodeError, biometrics.ErrEngineUnavailable.Error(), nil)
	}

	var matchedSupervisorID int32
	var matchedSupervisorLogin string

	if supervisorLogin != "" {
		// 1:1 Supervisor Verification
		supervisor, err := uc.repo.GetUserByUsername(ctx, supervisorLogin)
		if err != nil || !supervisor.IsActive.Bool || !supervisor.IsActive.Valid {
			return utils.NewResponse(utils.CodeUnauthorized, "invalid supervisor biometric credentials", nil)
		}
		templates, err := uc.repo.GetActiveUserBiometricTemplatesWithFMD(ctx, supervisor.ID)
		if err != nil || len(templates) == 0 {
			return utils.NewResponse(utils.CodeUnauthorized, "supervisor has no enrolled biometric templates", nil)
		}
		matched := false
		for _, t := range templates {
			candBytes, decErr := biometrics.ValidateAndDecodeFMD(t.FmdBase64)
			if decErr != nil {
				continue
			}
			_, ok, cmpErr := biometrics.Compare(probeFMD, candBytes, biometrics.DefaultMaxFalseMatchScore)
			if cmpErr == nil && ok {
				matched = true
				matchedSupervisorID = supervisor.ID
				matchedSupervisorLogin = supervisor.Username
				break
			}
		}
		if !matched {
			return utils.NewResponse(utils.CodeUnauthorized, "supervisor fingerprint did not match", nil)
		}
	} else {
		// 1:N Touch-Only Supervisor Identification
		allTemplates, err := uc.repo.ListAllActiveBiometricTemplates(ctx)
		if err != nil || len(allTemplates) == 0 {
			return utils.NewResponse(utils.CodeUnauthorized, "no enrolled supervisor fingerprints found", nil)
		}
		candidates := make([][]byte, 0, len(allTemplates))
		validRows := make([]repository.ListAllActiveBiometricTemplatesRow, 0, len(allTemplates))
		for _, r := range allTemplates {
			if b, decErr := biometrics.ValidateAndDecodeFMD(r.FmdBase64); decErr == nil {
				candidates = append(candidates, b)
				validRows = append(validRows, r)
			}
		}
		bestIdx, _, idErr := biometrics.Identify(probeFMD, candidates, biometrics.DefaultMaxFalseMatchScore)
		if idErr != nil || bestIdx < 0 || bestIdx >= len(validRows) {
			return utils.NewResponse(utils.CodeUnauthorized, "fingerprint did not match any enrolled supervisor", nil)
		}
		row := validRows[bestIdx]
		matchedSupervisorID = row.UserID
		matchedSupervisorLogin = row.Username
	}

	// Verify the matched supervisor holds the required permission
	hasPermission, err := uc.repo.CheckUserHasPermission(ctx, repository.CheckUserHasPermissionParams{
		UserID: matchedSupervisorID,
		Code:   permissionCode,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, "permission check failed: "+err.Error(), nil)
	}
	if !hasPermission {
		return utils.NewResponse(utils.CodeForbidden, "matched user ("+matchedSupervisorLogin+") does not have permission: "+permissionCode, nil)
	}

	token, err := middleware.GenerateAuthorizationToken(matchedSupervisorID, matchedSupervisorLogin, permissionCode, cashierID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to generate authorization token", nil)
	}

	expiresAt := time.Now().Add(5 * time.Minute)
	log.Printf("[BIOMETRICS] Supervisor override authorized by %s (user_id=%d) for permission=%s",
		matchedSupervisorLogin, matchedSupervisorID, permissionCode)

	return utils.NewResponse(utils.CodeOK, "action authorized via biometrics", AuthorizeActionOutput{
		AuthorizationToken: token,
		AuthorizedByUserID: matchedSupervisorID,
		AuthorizedByLogin:  matchedSupervisorLogin,
		PermissionCode:     permissionCode,
		ExpiresAt:          expiresAt,
	})
}
