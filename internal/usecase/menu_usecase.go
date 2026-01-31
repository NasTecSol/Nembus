package usecase

import (
	"context"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

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

	return utils.NewResponse(utils.CodeCreated, "menu created successfully", menu)
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

	return utils.NewResponse(utils.CodeOK, "menu fetched successfully", menu)
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

	return utils.NewResponse(utils.CodeOK, "menu fetched successfully", menu)
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

	return utils.NewResponse(utils.CodeOK, "menus fetched successfully", menus)
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

	return utils.NewResponse(utils.CodeOK, "menus fetched successfully", menus)
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

	return utils.NewResponse(utils.CodeOK, "active menus fetched successfully", menus)
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

	return utils.NewResponse(utils.CodeOK, "child menus fetched successfully", menus)
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

	return utils.NewResponse(utils.CodeOK, "menu updated successfully", menu)
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

	return utils.NewResponse(utils.CodeOK, "menu status updated successfully", menu)
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
