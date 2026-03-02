package usecase

import (
	"context"
	"encoding/json"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// PermissionOutput is the response shape for permission APIs. Metadata is json.RawMessage so JSONB marshals as JSON.
type PermissionOutput struct {
	ID          int32            `json:"id"`
	Name        string           `json:"name"`
	Code        string           `json:"code"`
	Description pgtype.Text      `json:"description"`
	Metadata    json.RawMessage  `json:"metadata"`
	CreatedAt   pgtype.Timestamp `json:"created_at"`
}

// PermissionWithScopeOutput adds scope to PermissionOutput for role/user permission queries.
type PermissionWithScopeOutput struct {
	PermissionOutput
	Scope pgtype.Text `json:"scope"`
}

func permissionToOutput(p repository.Permission) PermissionOutput {
	return PermissionOutput{
		ID:          p.ID,
		Name:        p.Name,
		Code:        p.Code,
		Description: p.Description,
		Metadata:    utils.BytesToJSONRawMessage(p.Metadata),
		CreatedAt:   p.CreatedAt,
	}
}

func rolePermissionWithScopeToOutput(row repository.GetRolePermissionsWithScopeRow) PermissionWithScopeOutput {
	return PermissionWithScopeOutput{
		PermissionOutput: PermissionOutput{
			ID:          row.ID,
			Name:        row.Name,
			Code:        row.Code,
			Description: row.Description,
			Metadata:    utils.BytesToJSONRawMessage(row.Metadata),
			CreatedAt:   row.CreatedAt,
		},
		Scope: row.Scope,
	}
}

func userPermissionWithScopeToOutput(row repository.GetUserPermissionsWithScopeRow) PermissionWithScopeOutput {
	return PermissionWithScopeOutput{
		PermissionOutput: PermissionOutput{
			ID:          row.ID,
			Name:        row.Name,
			Code:        row.Code,
			Description: row.Description,
			Metadata:    utils.BytesToJSONRawMessage(row.Metadata),
			CreatedAt:   row.CreatedAt,
		},
		Scope: row.Scope,
	}
}

// ModuleOutputForPermission is used for GetUserAccessibleModules response. Metadata as json.RawMessage.
type ModuleOutputForPermission struct {
	ID           int32            `json:"id"`
	Name         string           `json:"name"`
	Code         string           `json:"code"`
	Description  pgtype.Text      `json:"description"`
	Icon         pgtype.Text      `json:"icon"`
	IsActive     pgtype.Bool      `json:"is_active"`
	DisplayOrder pgtype.Int4      `json:"display_order"`
	Metadata     json.RawMessage  `json:"metadata"`
	CreatedAt    pgtype.Timestamp `json:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at"`
}

func moduleToOutputForPermission(m repository.Module) ModuleOutputForPermission {
	return ModuleOutputForPermission{
		ID:           m.ID,
		Name:         m.Name,
		Code:         m.Code,
		Description:  m.Description,
		Icon:         m.Icon,
		IsActive:     m.IsActive,
		DisplayOrder: m.DisplayOrder,
		Metadata:     utils.BytesToJSONRawMessage(m.Metadata),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// MenuOutputForPermission is used for GetUserAccessibleMenus response.
type MenuOutputForPermission struct {
	ID           int32            `json:"id"`
	ModuleID     int32            `json:"module_id"`
	ParentMenuID pgtype.Int4      `json:"parent_menu_id"`
	Name         string           `json:"name"`
	Code         string           `json:"code"`
	RoutePath    pgtype.Text      `json:"route_path"`
	Icon         pgtype.Text      `json:"icon"`
	DisplayOrder pgtype.Int4      `json:"display_order"`
	IsActive     pgtype.Bool      `json:"is_active"`
	Metadata     json.RawMessage  `json:"metadata"`
	CreatedAt    pgtype.Timestamp `json:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at"`
}

func menuToOutputForPermission(m repository.Menu) MenuOutputForPermission {
	return MenuOutputForPermission{
		ID:           m.ID,
		ModuleID:     m.ModuleID,
		ParentMenuID: m.ParentMenuID,
		Name:         m.Name,
		Code:         m.Code,
		RoutePath:    m.RoutePath,
		Icon:         m.Icon,
		DisplayOrder: m.DisplayOrder,
		IsActive:     m.IsActive,
		Metadata:     utils.BytesToJSONRawMessage(m.Metadata),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// SubmenuOutputForPermission is used for GetUserAccessibleSubmenus response.
type SubmenuOutputForPermission struct {
	ID              int32            `json:"id"`
	MenuID          int32            `json:"menu_id"`
	ParentSubmenuID pgtype.Int4      `json:"parent_submenu_id"`
	Name            string           `json:"name"`
	Code            string           `json:"code"`
	RoutePath       pgtype.Text      `json:"route_path"`
	Icon            pgtype.Text      `json:"icon"`
	DisplayOrder    pgtype.Int4      `json:"display_order"`
	IsActive        pgtype.Bool      `json:"is_active"`
	Metadata        json.RawMessage  `json:"metadata"`
	CreatedAt       pgtype.Timestamp `json:"created_at"`
	UpdatedAt       pgtype.Timestamp `json:"updated_at"`
}

func submenuToOutputForPermission(s repository.Submenu) SubmenuOutputForPermission {
	return SubmenuOutputForPermission{
		ID:              s.ID,
		MenuID:          s.MenuID,
		ParentSubmenuID: s.ParentSubmenuID,
		Name:            s.Name,
		Code:            s.Code,
		RoutePath:       s.RoutePath,
		Icon:            s.Icon,
		DisplayOrder:    s.DisplayOrder,
		IsActive:        s.IsActive,
		Metadata:        utils.BytesToJSONRawMessage(s.Metadata),
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

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

// CreatePermission creates a new permission
func (uc *PermissionUseCase) CreatePermission(ctx context.Context, name, code string, description *string, metadata []byte) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "permission name cannot be empty", nil)
	}
	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "permission code cannot be empty", nil)
	}
	var descText pgtype.Text
	if description != nil && *description != "" {
		descText = pgtype.Text{String: *description, Valid: true}
	}
	var meta []byte
	if len(metadata) == 0 {
		meta = []byte("{}")
	} else {
		meta = metadata
	}
	p, err := uc.repo.CreatePermission(ctx, repository.CreatePermissionParams{
		Name:        name,
		Code:        code,
		Description: descText,
		Metadata:    meta,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "permission created successfully", permissionToOutput(p))
}

// GetPermission returns a permission by ID
func (uc *PermissionUseCase) GetPermission(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	p, err := uc.repo.GetPermission(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "permission fetched successfully", permissionToOutput(p))
}

// GetPermissionByCode returns a permission by code
func (uc *PermissionUseCase) GetPermissionByCode(ctx context.Context, code string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "permission code is required", nil)
	}
	p, err := uc.repo.GetPermissionByCode(ctx, code)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "permission fetched successfully", permissionToOutput(p))
}

// ListPermissions returns paginated permissions
func (uc *PermissionUseCase) ListPermissions(ctx context.Context, limit, offset int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if limit <= 0 {
		limit = 50
	}
	perms, err := uc.repo.ListPermissions(ctx, repository.ListPermissionsParams{Limit: limit, Offset: offset})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]PermissionOutput, len(perms))
	for i := range perms {
		out[i] = permissionToOutput(perms[i])
	}
	return utils.NewResponse(utils.CodeOK, "permissions fetched successfully", out)
}

// UpdatePermission updates an existing permission
func (uc *PermissionUseCase) UpdatePermission(ctx context.Context, id int32, name, description *string, metadata []byte) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	params := repository.UpdatePermissionParams{ID: id}
	if name != nil {
		params.Name = pgtype.Text{String: *name, Valid: true}
	}
	if description != nil {
		params.Description = pgtype.Text{String: *description, Valid: true}
	}
	if len(metadata) > 0 {
		params.Metadata = metadata
	}
	p, err := uc.repo.UpdatePermission(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "permission updated successfully", permissionToOutput(p))
}

// DeletePermission deletes a permission
func (uc *PermissionUseCase) DeletePermission(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if err := uc.repo.DeletePermission(ctx, id); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "permission deleted successfully", nil)
}

// AssignPermissionToMenu assigns a permission to a menu
func (uc *PermissionUseCase) AssignPermissionToMenu(ctx context.Context, menuID, permissionID int32, metadata []byte) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	var meta []byte
	if len(metadata) == 0 {
		meta = []byte("{}")
	} else {
		meta = metadata
	}
	mp, err := uc.repo.AssignPermissionToMenu(ctx, repository.AssignPermissionToMenuParams{
		MenuID:       menuID,
		PermissionID: permissionID,
		Metadata:     meta,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "permission assigned to menu successfully", mp)
}

// GetMenuPermissions returns permissions for a menu
func (uc *PermissionUseCase) GetMenuPermissions(ctx context.Context, menuID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	perms, err := uc.repo.GetMenuPermissions(ctx, menuID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]PermissionOutput, len(perms))
	for i := range perms {
		out[i] = permissionToOutput(perms[i])
	}
	return utils.NewResponse(utils.CodeOK, "menu permissions fetched successfully", out)
}

// RevokePermissionFromMenu revokes a permission from a menu
func (uc *PermissionUseCase) RevokePermissionFromMenu(ctx context.Context, menuID, permissionID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if err := uc.repo.RevokePermissionFromMenu(ctx, repository.RevokePermissionFromMenuParams{
		MenuID:       menuID,
		PermissionID: permissionID,
	}); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "permission revoked from menu successfully", nil)
}

// AssignPermissionToModule assigns a permission to a module
func (uc *PermissionUseCase) AssignPermissionToModule(ctx context.Context, moduleID, permissionID int32, metadata []byte) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	var meta []byte
	if len(metadata) == 0 {
		meta = []byte("{}")
	} else {
		meta = metadata
	}
	mp, err := uc.repo.AssignPermissionToModule(ctx, repository.AssignPermissionToModuleParams{
		ModuleID:     moduleID,
		PermissionID: permissionID,
		Metadata:     meta,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "permission assigned to module successfully", mp)
}

// GetModulePermissions returns permissions for a module
func (uc *PermissionUseCase) GetModulePermissions(ctx context.Context, moduleID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	perms, err := uc.repo.GetModulePermissions(ctx, moduleID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]PermissionOutput, len(perms))
	for i := range perms {
		out[i] = permissionToOutput(perms[i])
	}
	return utils.NewResponse(utils.CodeOK, "module permissions fetched successfully", out)
}

// RevokePermissionFromModule revokes a permission from a module
func (uc *PermissionUseCase) RevokePermissionFromModule(ctx context.Context, moduleID, permissionID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if err := uc.repo.RevokePermissionFromModule(ctx, repository.RevokePermissionFromModuleParams{
		ModuleID:     moduleID,
		PermissionID: permissionID,
	}); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "permission revoked from module successfully", nil)
}

// AssignPermissionToSubmenu assigns a permission to a submenu
func (uc *PermissionUseCase) AssignPermissionToSubmenu(ctx context.Context, submenuID, permissionID int32, metadata []byte) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	var meta []byte
	if len(metadata) == 0 {
		meta = []byte("{}")
	} else {
		meta = metadata
	}
	sp, err := uc.repo.AssignPermissionToSubmenu(ctx, repository.AssignPermissionToSubmenuParams{
		SubmenuID:    submenuID,
		PermissionID: permissionID,
		Metadata:     meta,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "permission assigned to submenu successfully", sp)
}

// GetSubmenuPermissions returns permissions for a submenu
func (uc *PermissionUseCase) GetSubmenuPermissions(ctx context.Context, submenuID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	perms, err := uc.repo.GetSubmenuPermissions(ctx, submenuID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]PermissionOutput, len(perms))
	for i := range perms {
		out[i] = permissionToOutput(perms[i])
	}
	return utils.NewResponse(utils.CodeOK, "submenu permissions fetched successfully", out)
}

// RevokePermissionFromSubmenu revokes a permission from a submenu
func (uc *PermissionUseCase) RevokePermissionFromSubmenu(ctx context.Context, submenuID, permissionID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if err := uc.repo.RevokePermissionFromSubmenu(ctx, repository.RevokePermissionFromSubmenuParams{
		SubmenuID:    submenuID,
		PermissionID: permissionID,
	}); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "permission revoked from submenu successfully", nil)
}

// GetRolePermissionsWithScope returns permissions for a role with scope
func (uc *PermissionUseCase) GetRolePermissionsWithScope(ctx context.Context, roleID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	rows, err := uc.repo.GetRolePermissionsWithScope(ctx, roleID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]PermissionWithScopeOutput, len(rows))
	for i := range rows {
		out[i] = rolePermissionWithScopeToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "role permissions with scope fetched successfully", out)
}

// RevokePermissionFromRole revokes a permission from a role
func (uc *PermissionUseCase) RevokePermissionFromRole(ctx context.Context, roleID, permissionID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if err := uc.repo.RevokePermissionFromRole(ctx, repository.RevokePermissionFromRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	}); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "permission revoked from role successfully", nil)
}

// UpdateRolePermissionScope updates the scope of a permission on a role
func (uc *PermissionUseCase) UpdateRolePermissionScope(ctx context.Context, roleID, permissionID int32, scope *string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	var scopeText pgtype.Text
	if scope != nil {
		scopeText = pgtype.Text{String: *scope, Valid: true}
	}
	rp, err := uc.repo.UpdateRolePermissionScope(ctx, repository.UpdateRolePermissionScopeParams{
		RoleID:       roleID,
		PermissionID: permissionID,
		Scope:        scopeText,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "role permission scope updated successfully", rp)
}

// GetUserPermissions returns permissions for a user (via roles)
func (uc *PermissionUseCase) GetUserPermissions(ctx context.Context, userID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	perms, err := uc.repo.GetUserPermissions(ctx, userID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]PermissionOutput, len(perms))
	for i := range perms {
		out[i] = permissionToOutput(perms[i])
	}
	return utils.NewResponse(utils.CodeOK, "user permissions fetched successfully", out)
}

// GetUserPermissionsWithScope returns user permissions with scope
func (uc *PermissionUseCase) GetUserPermissionsWithScope(ctx context.Context, userID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	rows, err := uc.repo.GetUserPermissionsWithScope(ctx, userID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]PermissionWithScopeOutput, len(rows))
	for i := range rows {
		out[i] = userPermissionWithScopeToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "user permissions with scope fetched successfully", out)
}

// CheckUserHasPermission checks if a user has a specific permission by code
func (uc *PermissionUseCase) CheckUserHasPermission(ctx context.Context, userID int32, code string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "permission code is required", nil)
	}
	has, err := uc.repo.CheckUserHasPermission(ctx, repository.CheckUserHasPermissionParams{
		UserID: userID,
		Code:   code,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "permission check completed", map[string]interface{}{
		"has_permission": has,
		"permission_code": code,
		"user_id":        userID,
	})
}

// GetUserAccessibleModules returns modules accessible to a user
func (uc *PermissionUseCase) GetUserAccessibleModules(ctx context.Context, userID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	modules, err := uc.repo.GetUserAccessibleModules(ctx, userID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]ModuleOutputForPermission, len(modules))
	for i := range modules {
		out[i] = moduleToOutputForPermission(modules[i])
	}
	return utils.NewResponse(utils.CodeOK, "user accessible modules fetched successfully", out)
}

// GetUserAccessibleMenus returns menus accessible to a user
func (uc *PermissionUseCase) GetUserAccessibleMenus(ctx context.Context, userID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	menus, err := uc.repo.GetUserAccessibleMenus(ctx, userID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]MenuOutputForPermission, len(menus))
	for i := range menus {
		out[i] = menuToOutputForPermission(menus[i])
	}
	return utils.NewResponse(utils.CodeOK, "user accessible menus fetched successfully", out)
}

// GetUserAccessibleSubmenus returns submenus accessible to a user
func (uc *PermissionUseCase) GetUserAccessibleSubmenus(ctx context.Context, userID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	submenus, err := uc.repo.GetUserAccessibleSubmenus(ctx, userID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]SubmenuOutputForPermission, len(submenus))
	for i := range submenus {
		out[i] = submenuToOutputForPermission(submenus[i])
	}
	return utils.NewResponse(utils.CodeOK, "user accessible submenus fetched successfully", out)
}

// CheckUserSubmenuPermission checks if a user has access to a submenu by submenu code
func (uc *PermissionUseCase) CheckUserSubmenuPermission(ctx context.Context, userID int32, submenuCode string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if submenuCode == "" {
		return utils.NewResponse(utils.CodeBadReq, "submenu code cannot be empty", nil)
	}

	hasAccess, err := uc.repo.CheckUserHasSubmenuAccessByCode(ctx, repository.CheckUserHasSubmenuAccessByCodeParams{
		UserID: userID,
		Code:   submenuCode,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "permission check completed", map[string]interface{}{
		"has_access":   hasAccess,
		"submenu_code": submenuCode,
		"user_id":      userID,
	})
}
