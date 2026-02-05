package usecase

import (
	"context"
	"encoding/json"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// SubmenuOutput is the response shape for submenu APIs. Metadata is json.RawMessage so JSONB marshals as JSON.
type SubmenuOutput struct {
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

func submenuToOutput(s repository.Submenu) SubmenuOutput {
	return SubmenuOutput{
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

type SubmenuUseCase struct {
	repo *repository.Queries
}

// NewSubmenuUseCase creates a new use case without repository
func NewSubmenuUseCase() *SubmenuUseCase {
	return &SubmenuUseCase{}
}

// SetRepository injects repository per request
func (uc *SubmenuUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// CreateSubmenu creates a new submenu
func (uc *SubmenuUseCase) CreateSubmenu(
	ctx context.Context,
	menuID int32,
	parentSubmenuID *int32,
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
		return utils.NewResponse(utils.CodeBadReq, "submenu name cannot be empty", nil)
	}

	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "submenu code cannot be empty", nil)
	}

	if metadata == nil {
		metadata = []byte("{}")
	}

	submenu, err := uc.repo.CreateSubmenu(ctx, repository.CreateSubmenuParams{
		MenuID:          menuID,
		ParentSubmenuID: toPgInt4(parentSubmenuID),
		Name:            name,
		Code:            code,
		RoutePath:       toPgText(routePath),
		Icon:            toPgText(icon),
		DisplayOrder:    toPgInt4(displayOrder),
		IsActive:        pgtype.Bool{Bool: isActive, Valid: true},
		Metadata:        metadata,
	})

	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "submenu created successfully", submenuToOutput(submenu))
}

// GetSubmenu by ID
func (uc *SubmenuUseCase) GetSubmenu(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	submenu, err := uc.repo.GetSubmenu(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "submenu fetched successfully", submenuToOutput(submenu))
}

// GetSubmenuByCode fetches submenu by menu ID and code
func (uc *SubmenuUseCase) GetSubmenuByCode(ctx context.Context, menuID int32, code string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "submenu code is required", nil)
	}

	submenu, err := uc.repo.GetSubmenuByCode(ctx, repository.GetSubmenuByCodeParams{
		MenuID: menuID,
		Code:   code,
	})

	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "submenu fetched successfully", submenuToOutput(submenu))
}

// ListSubmenus lists all submenus
func (uc *SubmenuUseCase) ListSubmenus(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	submenus, err := uc.repo.ListSubmenus(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, err.Error(), nil)
	}
	out := make([]SubmenuOutput, len(submenus))
	for i := range submenus {
		out[i] = submenuToOutput(submenus[i])
	}
	return utils.NewResponse(utils.CodeOK, "submenus fetched successfully", out)
}

// ListSubmenusByMenu lists submenus under a specific menu
func (uc *SubmenuUseCase) ListSubmenusByMenu(ctx context.Context, menuID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	submenus, err := uc.repo.ListSubmenusByMenu(ctx, menuID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, err.Error(), nil)
	}
	out := make([]SubmenuOutput, len(submenus))
	for i := range submenus {
		out[i] = submenuToOutput(submenus[i])
	}
	return utils.NewResponse(utils.CodeOK, "submenus fetched successfully", out)
}

// ListActiveSubmenusByMenu lists only active submenus under a menu
func (uc *SubmenuUseCase) ListActiveSubmenusByMenu(ctx context.Context, menuID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	submenus, err := uc.repo.ListActiveSubmenusByMenu(ctx, menuID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, err.Error(), nil)
	}
	out := make([]SubmenuOutput, len(submenus))
	for i := range submenus {
		out[i] = submenuToOutput(submenus[i])
	}
	return utils.NewResponse(utils.CodeOK, "active submenus fetched successfully", out)
}

// ListSubmenusByParent lists child submenus by parent submenu
func (uc *SubmenuUseCase) ListSubmenusByParent(ctx context.Context, parentSubmenuID *int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	submenus, err := uc.repo.ListSubmenusByParent(ctx, toPgInt4(parentSubmenuID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]SubmenuOutput, len(submenus))
	for i := range submenus {
		out[i] = submenuToOutput(submenus[i])
	}
	return utils.NewResponse(utils.CodeOK, "child submenus fetched successfully", out)
}

// UpdateSubmenu updates a submenu
func (uc *SubmenuUseCase) UpdateSubmenu(
	ctx context.Context,
	id int32,
	parentSubmenuID *int32,
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
		return utils.NewResponse(utils.CodeBadReq, "submenu id is required", nil)
	}

	if metadata == nil {
		metadata = []byte("{}")
	}

	submenu, err := uc.repo.UpdateSubmenu(ctx, repository.UpdateSubmenuParams{
		ID:              id,
		ParentSubmenuID: toPgInt4(parentSubmenuID),
		Name:            name,
		RoutePath:       toPgText(routePath),
		Icon:            toPgText(icon),
		DisplayOrder:    toPgInt4(displayOrder),
		IsActive:        pgtype.Bool{Bool: isActive, Valid: true},
		Metadata:        metadata,
	})

	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "submenu updated successfully", submenuToOutput(submenu))
}

// ToggleSubmenuActive toggles submenu status
func (uc *SubmenuUseCase) ToggleSubmenuActive(ctx context.Context, id int32, isActive bool) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	submenu, err := uc.repo.ToggleSubmenuActive(ctx, repository.ToggleSubmenuActiveParams{
		ID:       id,
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})

	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "submenu status updated successfully", submenuToOutput(submenu))
}

// DeleteSubmenu deletes a submenu
func (uc *SubmenuUseCase) DeleteSubmenu(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	err := uc.repo.DeleteSubmenu(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "submenu deleted successfully", nil)
}
