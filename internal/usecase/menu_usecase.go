package usecase

import (
	"context"
	"encoding/json"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// MenuOutput is the response shape for menu APIs. Metadata is json.RawMessage so JSONB marshals as JSON.
type MenuOutput struct {
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

func menuToOutput(m repository.Menu) MenuOutput {
	return MenuOutput{
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

type MenuUseCase struct {
	repo *repository.Queries
}

// NewMenuUseCase creates a new use case without repository
func NewMenuUseCase() *MenuUseCase {
	return &MenuUseCase{}
}

// SetRepository injects repository per request
func (uc *MenuUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *MenuUseCase) CreateMenu(
	ctx context.Context,
	moduleID int32,
	parentMenuID *int32,
	name string,
	code string,
	routePath *string,
	icon *string,
	displayOrder *int32,
	isActive bool,
	metadata []byte,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "menu name cannot be empty", nil)
	}

	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "menu code cannot be empty", nil)
	}
	if metadata == nil {
		metadata = []byte("{}")
	}

	menu, err := uc.repo.CreateMenu(ctx, repository.CreateMenuParams{
		ModuleID:     moduleID,
		ParentMenuID: toPgInt4(parentMenuID),
		Name:         name,
		Code:         code,
		RoutePath:    toPgText(routePath),
		Icon:         toPgText(icon),
		DisplayOrder: toPgInt4(displayOrder),
		IsActive:     pgtype.Bool{Bool: isActive, Valid: true},
		Metadata:     metadata,
	})

	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "menu created successfully", menuToOutput(menu))
}

// 🔍 Get Menu by ID
func (uc *MenuUseCase) GetMenu(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	menu, err := uc.repo.GetMenu(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "menu fetched successfully", menuToOutput(menu))
}

// Get Menu by Code
func (uc *MenuUseCase) GetMenuByCode(
	ctx context.Context,
	moduleID int32,
	code string,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "menu code is required", nil)
	}

	menu, err := uc.repo.GetMenuByCode(ctx, repository.GetMenuByCodeParams{
		ModuleID: moduleID,
		Code:     code,
	})

	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "menu fetched successfully", menuToOutput(menu))
}

// 📋 List Menus
func (uc *MenuUseCase) ListMenus(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	menus, err := uc.repo.ListMenus(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, err.Error(), nil)
	}
	out := make([]MenuOutput, len(menus))
	for i := range menus {
		out[i] = menuToOutput(menus[i])
	}
	return utils.NewResponse(utils.CodeOK, "menus fetched successfully", out)
}

// 📦 List Menus by Module
func (uc *MenuUseCase) ListMenusByModule(
	ctx context.Context,
	moduleID int32,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	menus, err := uc.repo.ListMenusByModule(ctx, moduleID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, err.Error(), nil)
	}
	out := make([]MenuOutput, len(menus))
	for i := range menus {
		out[i] = menuToOutput(menus[i])
	}
	return utils.NewResponse(utils.CodeOK, "menus fetched successfully", out)
}

// 🟢 List Active Menus by Module
func (uc *MenuUseCase) ListActiveMenusByModule(
	ctx context.Context,
	moduleID int32,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	menus, err := uc.repo.ListActiveMenusByModule(ctx, moduleID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, err.Error(), nil)
	}
	out := make([]MenuOutput, len(menus))
	for i := range menus {
		out[i] = menuToOutput(menus[i])
	}
	return utils.NewResponse(utils.CodeOK, "active menus fetched successfully", out)
}

// 🌳 List Menus by Parent
func (uc *MenuUseCase) ListMenusByParent(
	ctx context.Context,
	parentMenuID *int32,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	menus, err := uc.repo.ListMenusByParent(ctx, toPgInt4(parentMenuID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]MenuOutput, len(menus))
	for i := range menus {
		out[i] = menuToOutput(menus[i])
	}
	return utils.NewResponse(utils.CodeOK, "child menus fetched successfully", out)
}

// ✏️ Update Menu
func (uc *MenuUseCase) UpdateMenu(
	ctx context.Context,
	id int32,
	parentMenuID *int32,
	name string,
	routePath *string,
	icon *string,
	displayOrder *int32,
	isActive bool,
	metadata []byte,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if id == 0 {
		return utils.NewResponse(utils.CodeBadReq, "menu id is required", nil)
	}

	menu, err := uc.repo.UpdateMenu(ctx, repository.UpdateMenuParams{
		ID:           id,
		ParentMenuID: toPgInt4(parentMenuID),
		Name:         name,
		RoutePath:    toPgText(routePath),
		Icon:         toPgText(icon),
		DisplayOrder: toPgInt4(displayOrder),
		IsActive:     pgtype.Bool{Bool: isActive, Valid: true},
		Metadata:     metadata,
	})

	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "menu updated successfully", menuToOutput(menu))
}

// 🔁 Toggle Menu Active
func (uc *MenuUseCase) ToggleMenuActive(
	ctx context.Context,
	id int32,
	isActive bool,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	menu, err := uc.repo.ToggleMenuActive(ctx, repository.ToggleMenuActiveParams{
		ID:       id,
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})

	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "menu status updated successfully", menuToOutput(menu))
}

// 🗑 Delete Menu
func (uc *MenuUseCase) DeleteMenu(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	err := uc.repo.DeleteMenu(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "menu deleted successfully", nil)
}

func toPgInt4(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

func toPgText(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *v, Valid: true}
}
