package usecase

import (
	"context"
	"encoding/json"
	"strconv"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// ModuleOutput is the response shape for module APIs. Metadata is json.RawMessage so JSONB marshals as JSON.
type ModuleOutput struct {
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

func moduleToOutput(m repository.Module) ModuleOutput {
	return ModuleOutput{
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

	return utils.NewResponse(utils.CodeCreated, "module created successfully", moduleToOutput(module))
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

	return utils.NewResponse(utils.CodeOK, "module fetched successfully", moduleToOutput(module))
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

	return utils.NewResponse(utils.CodeOK, "module fetched successfully", moduleToOutput(module))
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
	out := make([]ModuleOutput, len(modules))
	for i := range modules {
		out[i] = moduleToOutput(modules[i])
	}
	return utils.NewResponse(utils.CodeOK, "modules fetched successfully", out)
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

	return utils.NewResponse(utils.CodeOK, "module updated successfully", moduleToOutput(module))
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
