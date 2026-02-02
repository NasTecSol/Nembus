package usecase

import (
	"context"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type RolePermissionInput struct {
	PermissionID int32
	Scope        *string
	Metadata     []byte
}

// RoleUseCase handles business logic for roles.
type RoleUseCase struct {
	repo *repository.Queries
}

// NewRoleUseCase creates a new role use case without a repository.
// Repository will be injected per request via SetRepository.
func NewRoleUseCase() *RoleUseCase {
	return &RoleUseCase{}
}

// SetRepository sets the repository for this request.
func (uc *RoleUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// CreateRole creates a new role.
func (uc *RoleUseCase) CreateRole(
	ctx context.Context,
	name string,
	code string,
	description *string,
	isSystemRole bool,
	isActive bool,
	metadata []byte,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "role name cannot be empty", nil)
	}
	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "role code cannot be empty", nil)
	}

	var descText pgtype.Text
	if description != nil && *description != "" {
		descText = pgtype.Text{String: *description, Valid: true}
	}

	if metadata == nil {
		metadata = []byte("{}")
	}

	role, err := uc.repo.CreateRole(ctx, repository.CreateRoleParams{
		Name:         name,
		Code:         code,
		Description:  descText,
		IsSystemRole: pgtype.Bool{Bool: isSystemRole, Valid: true},
		IsActive:     pgtype.Bool{Bool: isActive, Valid: true},
		Metadata:     metadata,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "role created successfully", role)
}

// GetRole gets a role by ID.
func (uc *RoleUseCase) GetRole(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	role, err := uc.repo.GetRole(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "role fetched successfully", role)
}

// GetRoleByCode gets a role by its code.
func (uc *RoleUseCase) GetRoleByCode(ctx context.Context, code string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "role code cannot be empty", nil)
	}

	role, err := uc.repo.GetRoleByCode(ctx, code)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "role fetched successfully", role)
}

// ListRoles lists all roles.
func (uc *RoleUseCase) ListRoles(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	roles, err := uc.repo.ListRoles(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "roles fetched successfully", roles)
}

// ListActiveRoles lists all active roles.
func (uc *RoleUseCase) ListActiveRoles(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	roles, err := uc.repo.ListActiveRoles(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "active roles fetched successfully", roles)
}

// ListNonSystemRoles lists all non-system roles.
func (uc *RoleUseCase) ListNonSystemRoles(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	roles, err := uc.repo.ListNonSystemRoles(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "non-system roles fetched successfully", roles)
}

// UpdateRole updates an existing role.
func (uc *RoleUseCase) UpdateRole(
	ctx context.Context,
	id int32,
	name string,
	description *string,
	isActive bool,
	metadata []byte,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if id <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid role id", nil)
	}
	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "role name cannot be empty", nil)
	}

	var descText pgtype.Text
	if description != nil && *description != "" {
		descText = pgtype.Text{String: *description, Valid: true}
	}

	if metadata == nil {
		metadata = []byte("{}")
	}

	role, err := uc.repo.UpdateRole(ctx, repository.UpdateRoleParams{
		ID:          id,
		Name:        name,
		Description: descText,
		IsActive:    pgtype.Bool{Bool: isActive, Valid: true},
		Metadata:    metadata,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "role updated successfully", role)
}

// DeleteRole deletes a role by ID.
func (uc *RoleUseCase) DeleteRole(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if id <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid role id", nil)
	}

	if err := uc.repo.DeleteRole(ctx, id); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "role deleted successfully", nil)
}

func (uc *RoleUseCase) AssignPermissionToRole(
	ctx context.Context,
	roleID int32,
	permissions []RolePermissionInput,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if roleID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid role id", nil)
	}

	var results []repository.RolePermission
	for _, p := range permissions {
		if p.PermissionID <= 0 {
			return utils.NewResponse(utils.CodeBadReq, "invalid permission id", nil)
		}

		var scopeText pgtype.Text
		if p.Scope != nil && *p.Scope != "" {
			scopeText = pgtype.Text{String: *p.Scope, Valid: true}
		}

		if p.Metadata == nil {
			p.Metadata = []byte("{}")
		}

		rolePermission, err := uc.repo.AssignPermissionToRole(ctx, repository.AssignPermissionToRoleParams{
			RoleID:       roleID,
			PermissionID: p.PermissionID,
			Scope:        scopeText,
			Metadata:     p.Metadata,
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, err.Error(), nil)
		}

		results = append(results, rolePermission)
	}

	return utils.NewResponse(utils.CodeCreated, "permissions assigned to role successfully", results)
}

// RemovePermissionFromRole removes a permission from a role.
func (uc *RoleUseCase) RemovePermissionFromRole(
	ctx context.Context,
	roleID int32,
	permissionIDs []int32,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if roleID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid role id", nil)
	}
	if len(permissionIDs) == 0 {
		return utils.NewResponse(utils.CodeBadReq, "permission ids are required", nil)
	}

	for _, permID := range permissionIDs {
		if permID <= 0 {
			return utils.NewResponse(utils.CodeBadReq, "invalid permission id", nil)
		}

		err := uc.repo.RemovePermissionFromRole(ctx, repository.RemovePermissionFromRoleParams{
			RoleID:       roleID,
			PermissionID: permID,
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, err.Error(), nil)
		}
	}

	return utils.NewResponse(utils.CodeOK, "permission removed from role successfully", nil)
}

// GetRolePermissions lists permissions for a role.
func (uc *RoleUseCase) GetRolePermissions(ctx context.Context, roleID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if roleID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid role id", nil)
	}

	perms, err := uc.repo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "role permissions fetched successfully", perms)
}

// ToggleRoleActive toggles the active flag for a role.
func (uc *RoleUseCase) ToggleRoleActive(
	ctx context.Context,
	id int32,
	isActive bool,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if id <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid role id", nil)
	}

	role, err := uc.repo.ToggleRoleActive(ctx, repository.ToggleRoleActiveParams{
		ID:       id,
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "role status updated successfully", role)
}

// CheckRoleHasPermission checks if a role has a specific permission.
func (uc *RoleUseCase) CheckRoleHasPermission(
	ctx context.Context,
	roleID int32,
	permissionID int32,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if roleID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid role id", nil)
	}
	if permissionID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid permission id", nil)
	}

	hasPerm, err := uc.repo.CheckRoleHasPermission(ctx, repository.CheckRoleHasPermissionParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "role permission check completed", hasPerm)
}
