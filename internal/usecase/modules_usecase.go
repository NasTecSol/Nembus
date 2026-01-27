package usecase

import (
	"context"
	"strconv"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type ModuleUseCase struct {
	repo *repository.Queries
}

// NewModuleUseCase creates a new use case without repository
func NewModuleUseCase() *ModuleUseCase {
	return &ModuleUseCase{}
}

// SetRepository sets repository per request
func (uc *ModuleUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// MODULE CREATION USECASE
func (uc *ModuleUseCase) CreateModule(
	ctx context.Context,
	name string,
	code string,
	description *string,
	icon *string,
	isActive bool,
	displayOrder int32,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "module name cannot be empty", nil)
	}

	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "module code cannot be empty", nil)
	}

	var descText pgtype.Text
	if description != nil && *description != "" {
		descText = pgtype.Text{String: *description, Valid: true}
	}

	var iconText pgtype.Text
	if icon != nil && *icon != "" {
		iconText = pgtype.Text{String: *icon, Valid: true}
	}

	module, err := uc.repo.CreateModule(ctx, repository.CreateModuleParams{
		Name:         name,
		Code:         code,
		Description:  descText,
		Icon:         iconText,
		IsActive:     pgtype.Bool{Bool: isActive, Valid: true},
		DisplayOrder: pgtype.Int4{Int32: displayOrder, Valid: true},
		Metadata:     []byte("{}"),
	})

	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "module created successfully", module)
}

// MODULE FETCH USECASE BY ID
func (uc *ModuleUseCase) GetModule(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	moduleID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid module id", nil)
	}

	module, err := uc.repo.GetModule(ctx, int32(moduleID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "module fetched successfully", module)
}

// MODULE FETCH USECASE BY CODE
func (uc *ModuleUseCase) GetModuleByCode(ctx context.Context, code string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "module code is required", nil)
	}

	module, err := uc.repo.GetModuleByCode(ctx, code)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "module fetched successfully", module)
}

// MODULE LISTING USECASE
func (uc *ModuleUseCase) ListModules(
	ctx context.Context,
	isActive *bool,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	var active pgtype.Bool
	if isActive != nil {
		active = pgtype.Bool{Bool: *isActive, Valid: true}
	} else {
		active = pgtype.Bool{Valid: false} // no filter
	}

	modules, err := uc.repo.ListModules(ctx, active)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "modules fetched successfully", modules)
}

// MODULE UPDATE USECASE
func (uc *ModuleUseCase) UpdateModule(
	ctx context.Context,
	id string,
	name *string,
	description *string,
	icon *string,
	isActive *bool,
	displayOrder *int32,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	moduleID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid module id", nil)
	}

	params := repository.UpdateModuleParams{
		ID: int32(moduleID),
	}

	if name != nil {
		params.Name = pgtype.Text{String: *name, Valid: true}
	}

	if description != nil {
		params.Description = pgtype.Text{String: *description, Valid: true}
	}

	if icon != nil {
		params.Icon = pgtype.Text{String: *icon, Valid: true}
	}

	if isActive != nil {
		params.IsActive = pgtype.Bool{Bool: *isActive, Valid: true}
	}

	if displayOrder != nil {
		params.DisplayOrder = pgtype.Int4{Int32: *displayOrder, Valid: true}
	}

	module, err := uc.repo.UpdateModule(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "module updated successfully", module)
}

// MODULE DELETION USECASE
func (uc *ModuleUseCase) DeleteModule(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	moduleID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid module id", nil)
	}

	err = uc.repo.DeleteModule(ctx, int32(moduleID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "module deleted successfully", nil)
}

// GET FULL NAVIGATION HIERARCHY USECASE
func (uc *ModuleUseCase) GetNavigationHierarchy(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	data, err := uc.repo.GetFullNavigationHierarchy(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "navigation hierarchy fetched successfully", data)
}
