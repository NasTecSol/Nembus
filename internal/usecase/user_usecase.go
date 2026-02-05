package usecase

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

// UserOutput is the response shape for user APIs. Metadata is json.RawMessage so JSONB marshals as JSON.
// PasswordHash is never serialized to API responses.
type UserOutput struct {
	ID             int32            `json:"id"`
	OrganizationID int32            `json:"organization_id"`
	Username       string           `json:"username"`
	Email          string           `json:"email"`
	PasswordHash   string           `json:"-"` // never expose in API
	FirstName      pgtype.Text      `json:"first_name"`
	LastName       pgtype.Text      `json:"last_name"`
	EmployeeCode   pgtype.Text      `json:"employee_code"`
	IsActive       pgtype.Bool      `json:"is_active"`
	Metadata       json.RawMessage  `json:"metadata"`
	CreatedAt      pgtype.Timestamp `json:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at"`
}

func userToOutput(u repository.User) UserOutput {
	return UserOutput{
		ID:             u.ID,
		OrganizationID: u.OrganizationID,
		Username:       u.Username,
		Email:          u.Email,
		PasswordHash:   "", // never sent (json:"-")
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		EmployeeCode:   u.EmployeeCode,
		IsActive:       u.IsActive,
		Metadata:       utils.BytesToJSONRawMessage(u.Metadata),
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

type UserUseCase struct {
	repo *repository.Queries
}

// NewUserUseCase creates a new use case without a repository
// Repository will be injected per request via SetRepository
func NewUserUseCase() *UserUseCase {
	return &UserUseCase{}
}

// SetRepository sets the repository for this request
func (uc *UserUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// getOrganizationID gets the first active organization from the tenant database
// Since each tenant database is isolated, we assume there's one organization
func (uc *UserUseCase) getOrganizationID(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	// Get the first active organization
	orgs, err := uc.repo.ListOrganizations(ctx, repository.ListOrganizationsParams{
		Limit:    1,
		Offset:   0,
		IsActive: pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	if len(orgs) == 0 {
		return utils.NewResponse(utils.CodeNotFound, "no active organization found in tenant database", nil)
	}
	return utils.NewResponse(utils.CodeOK, "organization found successfully", orgs[0].ID)
}

// CreateUser creates a new user
func (uc *UserUseCase) CreateUser(ctx context.Context, firstName, lastName, username, email string, isActive bool, password *string, employeeCode *string) *repository.Response {
	if uc.repo == nil {
		//return repository.User{}, errors.New("repository not set")
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if firstName == "" {
		//return repository.User{}, errors.New("first name cannot be empty")
		return utils.NewResponse(utils.CodeBadReq, "first name cannot be empty", nil)
	}
	if username == "" {
		//return repository.User{}, errors.New("username cannot be empty")
		return utils.NewResponse(utils.CodeBadReq, "username cannot be empty", nil)
	}
	if email == "" {
		//return repository.User{}, errors.New("email cannot be empty")
		return utils.NewResponse(utils.CodeBadReq, "email cannot be empty", nil)
	}

	// Get organization ID
	orgResp := uc.getOrganizationID(ctx)
	if orgResp.StatusCode != utils.CodeOK {
		return orgResp
	}
	orgID := orgResp.Data.(int32)

	// Prepare password_hash
	var passwordHash string
	if password != nil && *password != "" {
		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			//return repository.User{}, errors.New("failed to hash password")
			return utils.NewResponse(utils.CodeError, "failed to hash password", nil)
		}
		passwordHash = string(hashedPassword)
	} else {
		passwordHash = "" // Empty password hash if not provided
	}

	// Prepare optional fields
	var firstNameText pgtype.Text
	if firstName != "" {
		firstNameText = pgtype.Text{String: firstName, Valid: true}
	}

	var lastNameText pgtype.Text
	if lastName != "" {
		lastNameText = pgtype.Text{String: lastName, Valid: true}
	}

	var employeeCodeText pgtype.Text
	if employeeCode != nil && *employeeCode != "" {
		employeeCodeText = pgtype.Text{String: *employeeCode, Valid: true}
	}

	// return uc.repo.CreateUser(ctx, repository.CreateUserParams{
	// 	OrganizationID: orgID,
	// 	Username:       username,
	// 	Email:          email,
	// 	PasswordHash:   passwordHash,
	// 	FirstName:      firstNameText,
	// 	LastName:       lastNameText,
	// 	EmployeeCode:   employeeCodeText,
	// 	IsActive:       pgtype.Bool{Bool: isActive, Valid: true},
	// 	Metadata:       []byte("{}"),
	// })
	user, err := uc.repo.CreateUser(ctx, repository.CreateUserParams{
		OrganizationID: orgID,
		Username:       username,
		Email:          email,
		PasswordHash:   passwordHash,
		FirstName:      firstNameText,
		LastName:       lastNameText,
		EmployeeCode:   employeeCodeText,
		IsActive:       pgtype.Bool{Bool: isActive, Valid: true},
		Metadata:       []byte("{}"),
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	// Return success response
	return utils.NewResponse(utils.CodeCreated, "user created successfully", userToOutput(user))
}

// GetUser gets a user by ID
func (uc *UserUseCase) GetUser(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	// Parse ID as int32
	userID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "invalid user ID", nil)
	}

	user, err := uc.repo.GetUser(ctx, int32(userID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "user fetched successfully", userToOutput(user))
}

// ListUsers lists all users for the organization
func (uc *UserUseCase) ListUsers(ctx context.Context, limit, offset int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	// Get organization ID
	orgResp := uc.getOrganizationID(ctx)
	if orgResp.StatusCode != utils.CodeOK {
		return orgResp
	}
	orgID := orgResp.Data.(int32)

	// Set default limit if not provided
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	users, err := uc.repo.ListUsers(ctx, repository.ListUsersParams{
		OrganizationID: orgID,
		Limit:          limit,
		Offset:         offset,
		IsActive:       pgtype.Bool{Valid: false}, // Don't filter by is_active
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]UserOutput, len(users))
	for i := range users {
		out[i] = userToOutput(users[i])
	}
	return utils.NewResponse(utils.CodeOK, "users fetched successfully", out)
}

func decodeJSONMetadata(b []byte) (map[string]interface{}, error) {
	if len(b) == 0 {
		return map[string]interface{}{}, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// AssignRoleToUser assigns a role to a user
func (uc *UserUseCase) AssignRoleToUser(
	ctx context.Context,
	userID int32,
	roleID int32,
	storeID *int32, // 👈 optional
	metadata []byte,
) *repository.Response {

	log.Printf("[AssignRoleToUser] start | userID=%d roleID=%d storeID=%v", userID, roleID, storeID)

	// 1. Repo check
	if uc.repo == nil {
		log.Println("[AssignRoleToUser] repository not set")
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	// 2. Validation
	if userID <= 0 {
		log.Println("[AssignRoleToUser] invalid user id")
		return utils.NewResponse(utils.CodeBadReq, "invalid user id", nil)
	}

	if roleID <= 0 {
		log.Println("[AssignRoleToUser] invalid role id")
		return utils.NewResponse(utils.CodeBadReq, "invalid role id", nil)
	}

	if metadata == nil {
		log.Println("[AssignRoleToUser] metadata is nil, defaulting to {}")
		metadata = []byte("{}")
	}

	// 3. Check if user already has role
	hasRole, err := uc.repo.CheckUserHasRole(ctx, repository.CheckUserHasRoleParams{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		log.Printf("[AssignRoleToUser] failed to check user role | err=%v", err)
		return utils.NewResponse(utils.CodeError, "failed to check user role", nil)
	}

	if hasRole {
		log.Printf("[AssignRoleToUser] user already has role | userID=%d roleID=%d", userID, roleID)
		return utils.NewResponse(
			utils.CodeBadReq,
			"user already has this role",
			nil,
		)
	}

	// 4. Assign role
	log.Printf("[AssignRoleToUser] assigning role | userID=%d roleID=%d", userID, roleID)
	userRole, err := uc.repo.AssignRoleToUser(ctx, repository.AssignRoleToUserParams{
		UserID:   userID,
		RoleID:   roleID,
		Metadata: metadata,
	})
	if err != nil {
		log.Printf("[AssignRoleToUser] failed to assign role | err=%v", err)
		return utils.NewResponse(utils.CodeBadReq, "failed to assign role", nil)
	}

	// After Assigning Role → assign store access
	log.Println("[AssignRoleToUser] role assigned, fetching role metadata")

	// 3. Fetch role
	role, err := uc.repo.GetRole(ctx, roleID)
	if err != nil {
		log.Printf("[AssignRoleToUser] failed to fetch role | roleID=%d err=%v", roleID, err)
		return utils.NewResponse(utils.CodeBadReq, "failed to fetch role", nil)
	}

	// 4. Decode role metadata (BASE64 → JSON)
	roleMetadata, err := decodeJSONMetadata(role.Metadata)
	if err != nil {
		log.Printf("[AssignRoleToUser] failed to decode role metadata | err=%v", err)
		return utils.NewResponse(utils.CodeBadReq, "invalid role metadata", nil)
	}

	scope, ok := roleMetadata["scope"].(string)
	if !ok || scope == "" {
		log.Println("[AssignRoleToUser] role scope missing")
		return utils.NewResponse(utils.CodeError, "role scope missing", nil)
	}

	log.Printf("[AssignRoleToUser] role scope detected | scope=%s", scope)

	switch scope {

	case "own":
		log.Println("[AssignRoleToUser] processing OWN scope")

		if storeID == nil || *storeID <= 0 {
			log.Println("[AssignRoleToUser] store_id missing for own scope")
			return utils.NewResponse(
				utils.CodeBadReq,
				"store_id is required for own scope",
				nil,
			)
		}

		ownMetadata := map[string]interface{}{}

		metaBytes, _ := json.Marshal(ownMetadata)
		metadata = metaBytes
		// 🔹 Grant store access immediately
		resp := uc.GrantStoreAccess(ctx, userID, *storeID, true, metadata)
		if resp.StatusCode != utils.CodeCreated {
			log.Printf("[AssignRoleToUser] failed to grant store access | err=%v", resp.Message)
			return utils.NewResponse(utils.CodeBadReq, "failed to grant store access", nil)
		}

		log.Printf("[AssignRoleToUser] own scope metadata attached | storeID=%d", *storeID)

	case "all":
		log.Println("[AssignRoleToUser] ALL scope detected")

		// 1️⃣ Get organization ID (same logic as StoreUseCase)
		orgResp := uc.getOrganizationID(ctx)
		if orgResp.StatusCode != utils.CodeOK {
			return orgResp
		}
		orgID := orgResp.Data.(int32)

		bodyStoreID := int32(0)
		if storeID != nil && *storeID > 0 {
			bodyStoreID = *storeID
		}

		// 2️⃣ Fetch ALL stores (no pagination, no filters)
		stores, err := uc.repo.ListStores(ctx, repository.ListStoresParams{
			OrganizationID: orgID,
			Limit:          1000, // 0 = no limit (sqlc style)
			Offset:         0,
			IsActive:       pgtype.Bool{}, // not filtering
			StoreType:      pgtype.Text{},
		})
		if err != nil {
			log.Printf("[AssignRoleToUser] failed to fetch stores | err=%v", err)
			return utils.NewResponse(
				utils.CodeError,
				"failed to fetch stores",
				nil,
			)
		}

		// 3️⃣ Grant secondary access to all stores except primary
		for _, store := range stores {
			if store.ID == bodyStoreID {
				continue // skip primary store for now
			}

			secondaryMetadata := map[string]interface{}{}
			metaBytes, _ := json.Marshal(secondaryMetadata)

			log.Printf(
				"[AssignRoleToUser] granting access (secondary) | userID=%d storeID=%d",
				userID,
				store.ID,
			)

			resp := uc.GrantStoreAccess(ctx, userID, store.ID, false, metaBytes)
			if resp.StatusCode != utils.CodeCreated {
				log.Printf(
					"[AssignRoleToUser] failed to grant access | storeID=%d err=%v",
					store.ID,
					resp.Message,
				)
				return utils.NewResponse(
					utils.CodeBadReq,
					"failed to grant access to one of the stores",
					nil,
				)
			}
		}

		// 4️⃣ Grant PRIMARY store from request body
		if bodyStoreID > 0 {
			primaryMetadata := map[string]interface{}{
				"scope":    "all",
				"store_id": bodyStoreID,
			}
			metaBytes, _ := json.Marshal(primaryMetadata)

			log.Printf(
				"[AssignRoleToUser] granting access (primary) | userID=%d storeID=%d",
				userID,
				bodyStoreID,
			)

			resp := uc.GrantStoreAccess(ctx, userID, bodyStoreID, true, metaBytes)
			if resp.StatusCode != utils.CodeCreated {
				log.Printf(
					"[AssignRoleToUser] failed to grant primary store access | err=%v",
					resp.Message,
				)
				return utils.NewResponse(
					utils.CodeBadReq,
					"failed to grant access to the primary store",
					nil,
				)
			}
		}

	case "specific":
		log.Println("[AssignRoleToUser] SPECIFIC scope detected")

		bodyStoreID := int32(0)
		if storeID != nil && *storeID > 0 {
			bodyStoreID = *storeID
		}

		// 1️⃣ Grant access for all stores in role metadata, skip request body store
		storeIDsRaw, ok := roleMetadata["ids"]
		if ok {
			storeIDsInterface, ok := storeIDsRaw.([]interface{})
			if ok {
				for _, sid := range storeIDsInterface {
					idFloat, ok := sid.(float64)
					if !ok {
						log.Printf("[AssignRoleToUser] invalid store ID type in role metadata: %v", sid)
						continue
					}
					sidInt := int32(idFloat)

					// skip if it's the same as the request body primary store
					if sidInt == bodyStoreID {
						continue
					}

					// Store-specific metadata
					storeMetadata := map[string]interface{}{}
					metaBytes, _ := json.Marshal(storeMetadata)

					log.Printf("[AssignRoleToUser] granting access (secondary) | userID=%d storeID=%d metadata=%s", userID, sidInt, string(metaBytes))

					resp := uc.GrantStoreAccess(ctx, userID, sidInt, false, metaBytes)
					if resp.StatusCode != utils.CodeCreated {
						log.Printf("[AssignRoleToUser] failed to grant access to store %d | err=%v", sidInt, resp.Message)
						return utils.NewResponse(
							utils.CodeBadReq,
							"failed to grant access to one of the specific stores",
							nil,
						)
					}
				}
			}
		}

		// 2️⃣ Grant access to the store ID from request body as PRIMARY
		if bodyStoreID > 0 {
			primaryMetadata := map[string]interface{}{}
			metaBytes, _ := json.Marshal(primaryMetadata)

			log.Printf("[AssignRoleToUser] granting access (primary) | userID=%d storeID=%d metadata=%s", userID, bodyStoreID, string(metaBytes))

			resp := uc.GrantStoreAccess(ctx, userID, bodyStoreID, true, metaBytes)
			if resp.StatusCode != utils.CodeCreated {
				log.Printf("[AssignRoleToUser] failed to grant primary store access | err=%v", resp.Message)
				return utils.NewResponse(
					utils.CodeBadReq,
					"failed to grant access to the primary store",
					nil,
				)
			}
		}

	default:
		log.Printf("[AssignRoleToUser] invalid role scope | scope=%s", scope)
		return utils.NewResponse(
			utils.CodeBadReq,
			"invalid role scope",
			nil,
		)
	}

	log.Printf("[AssignRoleToUser] success | userID=%d roleID=%d scope=%s", userID, roleID, scope)

	// 5. Success
	return utils.NewResponse(
		utils.CodeCreated,
		"role assigned successfully",
		userRole,
	)
}

// GetUsersByRole fetches all active users assigned to a specific role
func (uc *UserUseCase) GetUsersByRole(ctx context.Context, roleID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	users, err := uc.repo.GetUsersWithRole(ctx, roleID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]UserOutput, len(users))
	for i := range users {
		out[i] = userToOutput(users[i])
	}
	return utils.NewResponse(utils.CodeOK, "users fetched successfully", out)
}

// UpdateUser updates user details
func (uc *UserUseCase) UpdateUser(
	ctx context.Context,
	id int32,
	email, firstName, lastName, employeeCode *string,
	isActive *bool,
	metadata []byte,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	var emailText, firstNameText, lastNameText, employeeCodeText pgtype.Text
	if email != nil {
		emailText = pgtype.Text{String: *email, Valid: true}
	}
	if firstName != nil {
		firstNameText = pgtype.Text{String: *firstName, Valid: true}
	}
	if lastName != nil {
		lastNameText = pgtype.Text{String: *lastName, Valid: true}
	}
	if employeeCode != nil {
		employeeCodeText = pgtype.Text{String: *employeeCode, Valid: true}
	}

	var isActivePG pgtype.Bool
	if isActive != nil {
		isActivePG = pgtype.Bool{Bool: *isActive, Valid: true}
	} else {
		isActivePG = pgtype.Bool{Valid: false}
	}

	user, err := uc.repo.UpdateUser(ctx, repository.UpdateUserParams{
		Email:        emailText,
		FirstName:    firstNameText,
		LastName:     lastNameText,
		EmployeeCode: employeeCodeText,
		IsActive:     isActivePG,
		Metadata:     metadata,
		ID:           id,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "user updated successfully", user)
}

// UpdateUserPassword updates a user's password
func (uc *UserUseCase) UpdateUserPassword(ctx context.Context, userID int32, newPassword string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to hash password", nil)
	}

	row, err := uc.repo.UpdateUserPassword(ctx, repository.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: string(hashed),
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "password updated successfully", row)
}

// GrantStoreAccess grants a user access to a store
func (uc *UserUseCase) GrantStoreAccess(
	ctx context.Context,
	userID, storeID int32,
	isPrimary bool,
	metadata []byte,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	access, err := uc.repo.GrantStoreAccessToUser(ctx, repository.GrantStoreAccessToUserParams{
		UserID:    userID,
		StoreID:   storeID,
		IsPrimary: pgtype.Bool{Bool: isPrimary, Valid: true},
		Metadata:  metadata,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "store access granted successfully", access)
}

// SetUserPrimaryStore unsets other primary stores for a user
func (uc *UserUseCase) SetUserPrimaryStore(ctx context.Context, userID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if err := uc.repo.SetUserPrimaryStore(ctx, userID); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "user primary store unset successfully", nil)
}

// RevokeRole revokes a specific role from a user
func (uc *UserUseCase) RevokeRole(ctx context.Context, userID, roleID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if err := uc.repo.RevokeRoleFromUser(ctx, repository.RevokeRoleFromUserParams{
		UserID: userID,
		RoleID: roleID,
	}); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "role revoked successfully", nil)
}

// RevokeAllRoles revokes all roles from a user
func (uc *UserUseCase) RevokeAllRoles(ctx context.Context, userID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if err := uc.repo.RevokeAllRolesFromUser(ctx, userID); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "all roles revoked successfully", nil)
}

// RevokeStoreAccess revokes a user's access to a store
func (uc *UserUseCase) RevokeStoreAccess(ctx context.Context, userID, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if err := uc.repo.RevokeStoreAccessFromUser(ctx, repository.RevokeStoreAccessFromUserParams{
		UserID:  userID,
		StoreID: storeID,
	}); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "store access revoked successfully", nil)
}

// RevokeAllStoreAccess revokes all store access from a user
func (uc *UserUseCase) RevokeAllStoreAccess(ctx context.Context, userID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if err := uc.repo.RevokeAllStoreAccessFromUser(ctx, userID); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "all store access revoked successfully", nil)
}

// SearchUsers searches users by term
func (uc *UserUseCase) SearchUsers(ctx context.Context, searchTerm string, limit, offset int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	orgResp := uc.getOrganizationID(ctx)
	if orgResp.StatusCode != utils.CodeOK {
		return orgResp
	}
	orgID := orgResp.Data.(int32)

	users, err := uc.repo.SearchUsers(ctx, repository.SearchUsersParams{
		OrganizationID: orgID,
		Column2:        pgtype.Text{String: searchTerm, Valid: true},
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "users fetched successfully", users)
}

// GetUserWithDetails fetches user with roles and stores
func (uc *UserUseCase) GetUserWithDetails(ctx context.Context, userID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	user, err := uc.repo.GetUserWithDetails(ctx, userID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "user fetched with details successfully", user)
}

// ListUsersWithDetails lists users with roles and stores
func (uc *UserUseCase) ListUsersWithDetails(ctx context.Context, limit, offset int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	orgResp := uc.getOrganizationID(ctx)
	if orgResp.StatusCode != utils.CodeOK {
		return orgResp
	}
	orgID := orgResp.Data.(int32)

	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	users, err := uc.repo.ListUsersWithDetails(ctx, repository.ListUsersWithDetailsParams{
		OrganizationID: orgID,
		Limit:          limit,
		Offset:         offset,
		IsActive:       pgtype.Bool{Valid: false},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "users fetched with details successfully", users)
}

// GetStoreUsers fetches all users for a store
func (uc *UserUseCase) GetStoreUsers(ctx context.Context, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	users, err := uc.repo.GetStoreUsers(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "users fetched for store successfully", users)
}

// GetUserPrimaryStore fetches user's primary store
func (uc *UserUseCase) GetUserPrimaryStore(ctx context.Context, userID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	store, err := uc.repo.GetUserPrimaryStore(ctx, userID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "user primary store fetched successfully", store)
}

// GetUserStores fetches all stores for a user
func (uc *UserUseCase) GetUserStores(ctx context.Context, userID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	stores, err := uc.repo.GetUserStores(ctx, userID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "user stores fetched successfully", stores)
}
