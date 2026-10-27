package usecase

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

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

// AssignCashiersToPOSTerminal assigns cashier IDs to a terminal's metadata.
func (uc *PosTerminalsUseCase) AssignCashiersToPOSTerminal(ctx context.Context, terminalID int32, cashierIDs []int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	terminal, err := uc.repo.GetPOSTerminal(ctx, terminalID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "terminal not found", nil)
	}

	meta := make(map[string]interface{})
	if len(terminal.Metadata) > 0 {
		_ = json.Unmarshal(terminal.Metadata, &meta)
	}

	// Merge existing assigned_cashier_ids with incoming cashierIDs to prevent overwriting
	existingSet := make(map[int32]bool)
	if rawIDs, ok := meta["assigned_cashier_ids"]; ok && rawIDs != nil {
		if slice, ok := rawIDs.([]interface{}); ok {
			for _, item := range slice {
				if num, ok := item.(float64); ok {
					existingSet[int32(num)] = true
				}
			}
		}
	}
	for _, cid := range cashierIDs {
		existingSet[cid] = true
	}

	mergedIDs := make([]int32, 0, len(existingSet))
	for cid := range existingSet {
		mergedIDs = append(mergedIDs, cid)
	}

	meta["assigned_cashier_ids"] = mergedIDs
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to marshal metadata", nil)
	}

	updated, err := uc.repo.UpdatePOSTerminal(ctx, repository.UpdatePOSTerminalParams{
		ID:           terminalID,
		TerminalName: terminal.TerminalName,
		DeviceID:     terminal.DeviceID,
		IsActive:     terminal.IsActive,
		Metadata:     metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "cashiers assigned to terminal successfully", updated)
}

// GetCashiersForPOSTerminal returns all cashiers assigned to a terminal via its metadata.
func (uc *PosTerminalsUseCase) GetCashiersForPOSTerminal(ctx context.Context, terminalID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	terminal, err := uc.repo.GetPOSTerminal(ctx, terminalID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "terminal not found", nil)
	}

	if len(terminal.Metadata) == 0 {
		return utils.NewResponse(utils.CodeOK, "no cashiers assigned to terminal", []interface{}{})
	}

	meta := make(map[string]interface{})
	if err := json.Unmarshal(terminal.Metadata, &meta); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to unmarshal metadata", nil)
	}

	rawIDs, ok := meta["assigned_cashier_ids"]
	if !ok || rawIDs == nil {
		return utils.NewResponse(utils.CodeOK, "no cashiers assigned to terminal", []interface{}{})
	}

	var cashierIDs []int32
	switch v := rawIDs.(type) {
	case []interface{}:
		for _, item := range v {
			if num, ok := item.(float64); ok {
				cashierIDs = append(cashierIDs, int32(num))
			}
		}
	}

	if len(cashierIDs) == 0 {
		return utils.NewResponse(utils.CodeOK, "no cashiers assigned to terminal", []interface{}{})
	}

	var cashiers []repository.Cashier
	for _, cid := range cashierIDs {
		c, err := uc.repo.GetCashierByID(ctx, cid)
		if err == nil {
			cashiers = append(cashiers, c)
		}
	}

	return utils.NewResponse(utils.CodeOK, "cashiers fetched successfully", cashiers)
}

// RemoveCashierFromPOSTerminal unassigns a cashier ID from a terminal's metadata.
func (uc *PosTerminalsUseCase) RemoveCashierFromPOSTerminal(ctx context.Context, terminalID int32, cashierID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	terminal, err := uc.repo.GetPOSTerminal(ctx, terminalID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "terminal not found", nil)
	}

	if len(terminal.Metadata) == 0 {
		return utils.NewResponse(utils.CodeOK, "cashier not assigned to terminal", nil)
	}

	meta := make(map[string]interface{})
	if err := json.Unmarshal(terminal.Metadata, &meta); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to unmarshal metadata", nil)
	}

	rawIDs, ok := meta["assigned_cashier_ids"]
	if !ok || rawIDs == nil {
		return utils.NewResponse(utils.CodeOK, "cashier not assigned to terminal", nil)
	}

	var updatedIDs []int32
	removed := false
	switch v := rawIDs.(type) {
	case []interface{}:
		for _, item := range v {
			if num, ok := item.(float64); ok {
				cid := int32(num)
				if cid == cashierID {
					removed = true
				} else {
					updatedIDs = append(updatedIDs, cid)
				}
			}
		}
	}

	if !removed {
		return utils.NewResponse(utils.CodeBadReq, "cashier is not assigned to this terminal", nil)
	}

	meta["assigned_cashier_ids"] = updatedIDs
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to marshal metadata", nil)
	}

	updated, err := uc.repo.UpdatePOSTerminal(ctx, repository.UpdatePOSTerminalParams{
		ID:           terminalID,
		TerminalName: terminal.TerminalName,
		DeviceID:     terminal.DeviceID,
		IsActive:     terminal.IsActive,
		Metadata:     metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier unassigned from terminal successfully", updated)
}
