package usecase

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"NEMBUS/internal/repository"

	"NEMBUS/utils" // Assuming your NewResponse is here

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

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
	return utils.NewResponse(utils.CodeCreated, "user created successfully", user)
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

	return utils.NewResponse(utils.CodeOK, "user fetched successfully", user)
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

	return utils.NewResponse(utils.CodeOK, "users fetched successfully", users)
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

		ownMetadata := map[string]interface{}{
			"scope":    "own",
			"storeids": *storeID,
		}

		metaBytes, _ := json.Marshal(ownMetadata)
		metadata = metaBytes

		log.Printf("[AssignRoleToUser] own scope metadata attached | storeID=%d", *storeID)

	case "all":
		log.Println("[AssignRoleToUser] ALL scope detected (no action yet)")

	case "specific":
		log.Println("[AssignRoleToUser] SPECIFIC scope detected (no action yet)")

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

	return utils.NewResponse(utils.CodeOK, "users fetched successfully", users)
}
