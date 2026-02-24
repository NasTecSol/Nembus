package usecase

import (
	"context"
	"fmt"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// PosPaymentUseCase handles POS payment CRUD and listing.
type PosPaymentUseCase struct {
	repo *repository.Queries
}

// NewPosPaymentUseCase creates a new POS payment use case.
func NewPosPaymentUseCase() *PosPaymentUseCase {
	return &PosPaymentUseCase{}
}

// SetRepository sets the repository (called per-request from handler with tenant repo).
func (uc *PosPaymentUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// CreatePosPaymentInput is the input for CreatePayment.
type CreatePosPaymentInput struct {
	TransactionID    int32
	PaymentMethod    string
	PaymentGateway   *string
	Amount           pgtype.Numeric
	PaymentReference *string
	ReferenceNumber  *string
	Metadata         []byte
}

// CreatePayment creates a POS payment and optionally updates cashier session expected balance.
func (uc *PosPaymentUseCase) CreatePayment(ctx context.Context, in *CreatePosPaymentInput) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if in.PaymentMethod == "" {
		return utils.NewResponse(utils.CodeBadReq, "payment_method is required", nil)
	}
	_, err := uc.repo.GetPosTransaction(ctx, in.TransactionID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "transaction not found", nil)
	}
	params := repository.CreatePosPaymentParams{
		TransactionID:    in.TransactionID,
		PaymentMethod:    in.PaymentMethod,
		PaymentGateway:   pgtype.Text{},
		Amount:           in.Amount,
		PaymentReference: pgtype.Text{},
		ReferenceNumber:  pgtype.Text{},
		Metadata:         in.Metadata,
	}
	if in.PaymentGateway != nil {
		params.PaymentGateway = pgtype.Text{String: *in.PaymentGateway, Valid: true}
	}
	if in.PaymentReference != nil {
		params.PaymentReference = pgtype.Text{String: *in.PaymentReference, Valid: true}
	}
	if in.ReferenceNumber != nil {
		params.ReferenceNumber = pgtype.Text{String: *in.ReferenceNumber, Valid: true}
	}
	payment, err := uc.repo.CreatePosPayment(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create payment: "+err.Error(), nil)
	}
	// Update cashier session expected balance
	txn, _ := uc.repo.GetPosTransaction(ctx, in.TransactionID)
	if txn.CashierSessionID != 0 {
		_ = uc.repo.UpdateSessionExpectedBalance(ctx, repository.UpdateSessionExpectedBalanceParams{
			ID:              txn.CashierSessionID,
			ExpectedBalance: in.Amount,
		})
	}
	return utils.NewResponse(utils.CodeCreated, "payment created", payment)
}

// GetPayment returns a single POS payment by ID.
func (uc *PosPaymentUseCase) GetPayment(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	payment, err := uc.repo.GetPosPayment(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "payment not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "payment fetched", payment)
}

// ListPaymentsForTransaction returns all payments for a POS transaction (full rows).
func (uc *PosPaymentUseCase) ListPaymentsForTransaction(ctx context.Context, transactionID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetPosTransaction(ctx, transactionID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "transaction not found", nil)
	}
	rows, err := uc.repo.GetPaymentsForTransactionFull(ctx, transactionID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "payments fetched", rows)
}

// GetPaymentSummary returns total paid and payment breakdown for a transaction.
func (uc *PosPaymentUseCase) GetPaymentSummary(ctx context.Context, transactionID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	row, err := uc.repo.GetTransactionPaymentSummary(ctx, transactionID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "payment summary fetched", row)
}

// UpdatePosPaymentInput is the input for UpdatePayment.
type UpdatePosPaymentInput struct {
	ID               int32
	PaymentMethod    string
	PaymentGateway   *string
	Amount           pgtype.Numeric
	PaymentReference *string
	ReferenceNumber  *string
	Metadata         []byte
}

// UpdatePayment updates a POS payment by ID.
func (uc *PosPaymentUseCase) UpdatePayment(ctx context.Context, in *UpdatePosPaymentInput) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetPosPayment(ctx, in.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "payment not found", nil)
	}
	params := repository.UpdatePosPaymentParams{
		ID:               in.ID,
		PaymentMethod:    in.PaymentMethod,
		PaymentGateway:   pgtype.Text{},
		Amount:           in.Amount,
		PaymentReference: pgtype.Text{},
		ReferenceNumber:  pgtype.Text{},
		Metadata:         in.Metadata,
	}
	if in.PaymentGateway != nil {
		params.PaymentGateway = pgtype.Text{String: *in.PaymentGateway, Valid: true}
	}
	if in.PaymentReference != nil {
		params.PaymentReference = pgtype.Text{String: *in.PaymentReference, Valid: true}
	}
	if in.ReferenceNumber != nil {
		params.ReferenceNumber = pgtype.Text{String: *in.ReferenceNumber, Valid: true}
	}
	payment, err := uc.repo.UpdatePosPayment(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update payment: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "payment updated", payment)
}

// DeletePayment deletes a POS payment by ID.
func (uc *PosPaymentUseCase) DeletePayment(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetPosPayment(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "payment not found", nil)
	}
	err = uc.repo.DeletePosPayment(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to delete payment: %v", err), nil)
	}
	return utils.NewResponse(utils.CodeOK, "payment deleted", map[string]interface{}{"id": id})
}
