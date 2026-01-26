package usecase

import (
	"context"
	"errors"
	"strconv"

	"NEMBUS/internal/repository"

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
func (uc *UserUseCase) getOrganizationID(ctx context.Context) (int32, error) {
	if uc.repo == nil {
		return 0, errors.New("repository not set")
	}

	// Get the first active organization
	orgs, err := uc.repo.ListOrganizations(ctx, repository.ListOrganizationsParams{
		Limit:  1,
		Offset: 0,
		IsActive: pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		return 0, err
	}
	if len(orgs) == 0 {
		return 0, errors.New("no active organization found in tenant database")
	}
	return orgs[0].ID, nil
}

// CreateUser creates a new user
func (uc *UserUseCase) CreateUser(ctx context.Context, firstName, lastName, username, email string, isActive bool, password *string, employeeCode *string) (repository.User, error) {
	if uc.repo == nil {
		return repository.User{}, errors.New("repository not set")
	}
	if firstName == "" {
		return repository.User{}, errors.New("first name cannot be empty")
	}
	if username == "" {
		return repository.User{}, errors.New("username cannot be empty")
	}
	if email == "" {
		return repository.User{}, errors.New("email cannot be empty")
	}

	// Get organization ID
	orgID, err := uc.getOrganizationID(ctx)
	if err != nil {
		return repository.User{}, err
	}

	// Prepare password_hash
	var passwordHash string
	if password != nil && *password != "" {
		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			return repository.User{}, errors.New("failed to hash password")
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

	return uc.repo.CreateUser(ctx, repository.CreateUserParams{
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
}

// GetUser gets a user by ID
func (uc *UserUseCase) GetUser(ctx context.Context, id string) (repository.User, error) {
	if uc.repo == nil {
		return repository.User{}, errors.New("repository not set")
	}

	// Parse ID as int32
	userID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return repository.User{}, errors.New("invalid user ID")
	}

	return uc.repo.GetUser(ctx, int32(userID))
}

// ListUsers lists all users for the organization
func (uc *UserUseCase) ListUsers(ctx context.Context, limit, offset int32) ([]repository.User, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not set")
	}

	// Get organization ID
	orgID, err := uc.getOrganizationID(ctx)
	if err != nil {
		return nil, err
	}

	// Set default limit if not provided
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return uc.repo.ListUsers(ctx, repository.ListUsersParams{
		OrganizationID: orgID,
		Limit:          limit,
		Offset:         offset,
		IsActive:       pgtype.Bool{Valid: false}, // Don't filter by is_active
	})
}
