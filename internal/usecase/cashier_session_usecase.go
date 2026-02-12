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

func (uc *CashierSessionUseCase) CloseCashierSession(ctx context.Context, arg repository.CloseCashierSessionParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	session, err := uc.repo.CloseCashierSession(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to close cashier session", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cashier session closed successfully", session)
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
