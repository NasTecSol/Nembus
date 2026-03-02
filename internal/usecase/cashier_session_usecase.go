package usecase

import (
	"context"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"
)

type CashierSessionUseCase struct {
	repo *repository.Queries
}

func NewCashierSessionUseCase() *CashierSessionUseCase {
	return &CashierSessionUseCase{}
}

func (uc *CashierSessionUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *CashierSessionUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

func (uc *CashierSessionUseCase) OpenCashierSession(ctx context.Context, arg repository.OpenCashierSessionParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	session, err := uc.repo.OpenCashierSession(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to open cashier session", err.Error())
	}
	return utils.NewResponse(utils.CodeCreated, "cashier session opened successfully", session)
}

func (uc *CashierSessionUseCase) GetActiveCashierSession(ctx context.Context, cashierID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	session, err := uc.repo.GetActiveCashierSession(ctx, cashierID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "active cashier session not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "active cashier session fetched successfully", session)
}

func (uc *CashierSessionUseCase) GetSessionByID(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	session, err := uc.repo.GetSessionByID(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cashier session not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "session fetched successfully", session)
}

func (uc *CashierSessionUseCase) CloseCashierSession(ctx context.Context, arg repository.CloseCashierSessionParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	// Ensure session exists and is open before closing
	session, err := uc.repo.GetSessionByID(ctx, arg.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cashier session not found", nil)
	}
	if session.Status.Valid && session.Status.String != "open" {
		return utils.NewResponse(utils.CodeBadReq, "session is not open", nil)
	}
	// Reconcile: use physical closing_balance and compute variance = closing_balance - expected_balance
	row, err := uc.repo.CloseCashierSessionReconcile(ctx, repository.CloseCashierSessionReconcileParams{
		ID:             arg.ID,
		ClosingBalance: arg.ClosingBalance,
		Column3:        arg.Column5,
		Column4:        arg.Column6,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to close cashier session", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cashier session closed successfully", row)
}

func (uc *CashierSessionUseCase) GetSessionSummary(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	summary, err := uc.repo.GetSessionSummary(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "session summary not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "session summary fetched successfully", summary)
}
