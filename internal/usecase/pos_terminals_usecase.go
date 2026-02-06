package usecase

import (
	"context"
	"strings"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// PosTerminalsUseCase handles POS terminal business logic.
type PosTerminalsUseCase struct {
	repo *repository.Queries
}

// NewPosTerminalsUseCase creates a new POS terminals use case.
func NewPosTerminalsUseCase() *PosTerminalsUseCase {
	return &PosTerminalsUseCase{}
}

// SetRepository sets the repository (called per-request from handler with tenant repo).
func (uc *PosTerminalsUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// CreatePOSTerminalInput is the input for CreatePOSTerminal.
type CreatePOSTerminalInput struct {
	StoreID      int32
	TerminalCode string
	TerminalName *string
	DeviceID     *string
	IsActive     *bool
	Metadata     []byte
}

// CreatePOSTerminal creates a POS terminal for a store.
func (uc *PosTerminalsUseCase) CreatePOSTerminal(ctx context.Context, in *CreatePOSTerminalInput) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetStore(ctx, in.StoreID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "store not found", nil)
	}
	code := strings.TrimSpace(in.TerminalCode)
	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "terminal_code is required", nil)
	}
	_, err = uc.repo.GetPOSTerminalByCode(ctx, repository.GetPOSTerminalByCodeParams{
		StoreID:      in.StoreID,
		TerminalCode: code,
	})
	if err == nil {
		return utils.NewResponse(utils.CodeBadReq, "terminal with this code already exists for store", nil)
	}
	params := repository.CreatePOSTerminalParams{
		StoreID:      in.StoreID,
		TerminalCode: code,
		TerminalName: pgtype.Text{},
		DeviceID:     pgtype.Text{},
		IsActive:     pgtype.Bool{Bool: true, Valid: true},
		Metadata:     in.Metadata,
	}
	if in.TerminalName != nil {
		params.TerminalName = pgtype.Text{String: *in.TerminalName, Valid: true}
	}
	if in.DeviceID != nil {
		params.DeviceID = pgtype.Text{String: *in.DeviceID, Valid: true}
	}
	if in.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *in.IsActive, Valid: true}
	}
	terminal, err := uc.repo.CreatePOSTerminal(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "terminal created", terminal)
}

// GetPOSTerminal returns a single POS terminal by ID.
func (uc *PosTerminalsUseCase) GetPOSTerminal(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	terminal, err := uc.repo.GetPOSTerminal(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "terminal not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "terminal fetched successfully", terminal)
}

// GetPOSTerminalByCode returns a POS terminal by store ID and terminal code.
func (uc *PosTerminalsUseCase) GetPOSTerminalByCode(ctx context.Context, storeID int32, terminalCode string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "store not found", nil)
	}
	terminal, err := uc.repo.GetPOSTerminalByCode(ctx, repository.GetPOSTerminalByCodeParams{
		StoreID:      storeID,
		TerminalCode: strings.TrimSpace(terminalCode),
	})
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "terminal not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "terminal fetched successfully", terminal)
}

// ListPOSTerminals returns all POS terminals.
func (uc *PosTerminalsUseCase) ListPOSTerminals(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	terminals, err := uc.repo.ListPOSTerminals(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "terminals fetched successfully", terminals)
}

// ListPOSTerminalsByStore returns all POS terminals for a store.
func (uc *PosTerminalsUseCase) ListPOSTerminalsByStore(ctx context.Context, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "store not found", nil)
	}
	terminals, err := uc.repo.ListPOSTerminalsByStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "terminals fetched successfully", terminals)
}

// ListActivePOSTerminalsByStore returns active POS terminals for a store.
func (uc *PosTerminalsUseCase) ListActivePOSTerminalsByStore(ctx context.Context, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "store not found", nil)
	}
	terminals, err := uc.repo.ListActivePOSTerminalsByStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "terminals fetched successfully", terminals)
}

// UpdatePOSTerminalInput is the input for UpdatePOSTerminal.
type UpdatePOSTerminalInput struct {
	ID           int32
	TerminalName *string
	DeviceID     *string
	IsActive     *bool
	Metadata     []byte
}

// UpdatePOSTerminal updates a POS terminal.
func (uc *PosTerminalsUseCase) UpdatePOSTerminal(ctx context.Context, in *UpdatePOSTerminalInput) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetPOSTerminal(ctx, in.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "terminal not found", nil)
	}
	params := repository.UpdatePOSTerminalParams{
		ID:           in.ID,
		TerminalName: pgtype.Text{},
		DeviceID:     pgtype.Text{},
		IsActive:     pgtype.Bool{},
		Metadata:     in.Metadata,
	}
	if in.TerminalName != nil {
		params.TerminalName = pgtype.Text{String: *in.TerminalName, Valid: true}
	}
	if in.DeviceID != nil {
		params.DeviceID = pgtype.Text{String: *in.DeviceID, Valid: true}
	}
	if in.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *in.IsActive, Valid: true}
	}
	terminal, err := uc.repo.UpdatePOSTerminal(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "terminal updated", terminal)
}

// DeletePOSTerminal deletes a POS terminal.
func (uc *PosTerminalsUseCase) DeletePOSTerminal(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetPOSTerminal(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "terminal not found", nil)
	}
	if err := uc.repo.DeletePOSTerminal(ctx, id); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "terminal deleted", nil)
}

// TogglePOSTerminalActive sets the active state of a POS terminal.
func (uc *PosTerminalsUseCase) TogglePOSTerminalActive(ctx context.Context, id int32, isActive bool) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetPOSTerminal(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "terminal not found", nil)
	}
	terminal, err := uc.repo.TogglePOSTerminalActive(ctx, repository.TogglePOSTerminalActiveParams{
		ID:       id,
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "terminal updated", terminal)
}
