package usecase

import (
	"context"
	"fmt"
	"strconv"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type PaymentTermsUseCase struct {
	repo *repository.Queries
}

func NewPaymentTermsUseCase() *PaymentTermsUseCase {
	return &PaymentTermsUseCase{}
}

func (uc *PaymentTermsUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// Helper to convert pgtype.Numeric to *string
func pgNumericToStringPointer(n pgtype.Numeric) *string {
	if !n.Valid {
		return nil
	}
	val, err := n.Value()
	if err != nil || val == nil {
		return nil
	}
	s := fmt.Sprintf("%v", val)
	return &s
}

// Helper to convert repository.PaymentTerm to Output
func paymentTermToOutput(pt repository.PaymentTerm) PaymentTermOutput {
	var discountDays *int32
	if pt.DiscountDays.Valid {
		discountDays = &pt.DiscountDays.Int32
	}

	return PaymentTermOutput{
		ID:                 pt.ID,
		OrganizationID:     pt.OrganizationID,
		Code:               pt.Code,
		Name:               pt.Name,
		DueDays:            pt.DueDays,
		DiscountDays:       discountDays,
		DiscountPercentage: pgNumericToStringPointer(pt.DiscountPercentage),
		LateFeePercentage:  pgNumericToStringPointer(pt.LateFeePercentage),
		IsActive:           pt.IsActive.Bool,
		CreatedAt:          utils.FormatTimestamp(pt.CreatedAt),
	}
}

type PaymentTermOutput struct {
	ID                 int32   `json:"id"`
	OrganizationID     int32   `json:"organization_id"`
	Code               string  `json:"code"`
	Name               string  `json:"name"`
	DueDays            int32   `json:"due_days"`
	DiscountDays       *int32  `json:"discount_days,omitempty"`
	DiscountPercentage *string `json:"discount_percentage,omitempty"`
	LateFeePercentage  *string `json:"late_fee_percentage,omitempty"`
	IsActive           bool    `json:"is_active"`
	CreatedAt          string  `json:"created_at"`
}

type CreatePaymentTermInput struct {
	OrganizationID     int32    `json:"organization_id"`
	Code               string   `json:"code"`
	Name               string   `json:"name"`
	DueDays            int32    `json:"due_days"`
	DiscountDays       *int32   `json:"discount_days"`
	DiscountPercentage *float64 `json:"discount_percentage"`
	LateFeePercentage  *float64 `json:"late_fee_percentage"`
	IsActive           bool     `json:"is_active"`
}

type UpdatePaymentTermInput struct {
	Code               *string  `json:"code"`
	Name               *string  `json:"name"`
	DueDays            *int32   `json:"due_days"`
	DiscountDays       *int32   `json:"discount_days"`
	DiscountPercentage *float64 `json:"discount_percentage"`
	LateFeePercentage  *float64 `json:"late_fee_percentage"`
	IsActive           *bool    `json:"is_active"`
}

func (uc *PaymentTermsUseCase) CreatePaymentTerm(ctx context.Context, input CreatePaymentTermInput) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if input.Code == "" {
		return utils.NewResponse(utils.CodeBadReq, "code cannot be empty", nil)
	}
	if input.Name == "" {
		return utils.NewResponse(utils.CodeBadReq, "name cannot be empty", nil)
	}
	if input.DueDays < 0 {
		return utils.NewResponse(utils.CodeBadReq, "due_days cannot be negative", nil)
	}

	// Check duplicates
	_, err := uc.repo.GetPaymentTermByCode(ctx, repository.GetPaymentTermByCodeParams{
		OrganizationID: input.OrganizationID,
		Code:           input.Code,
	})
	if err == nil {
		return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("payment term with code '%s' already exists", input.Code), nil)
	}

	// Call repository
	pt, err := uc.repo.CreatePaymentTerm(ctx, repository.CreatePaymentTermParams{
		OrganizationID:     input.OrganizationID,
		Code:               input.Code,
		Name:               input.Name,
		DueDays:            input.DueDays,
		DiscountDays:       utils.Int32ToPgInt4(input.DiscountDays),
		DiscountPercentage: utils.Float64PointerToPgNumeric(input.DiscountPercentage),
		LateFeePercentage:  utils.Float64PointerToPgNumeric(input.LateFeePercentage),
		IsActive:           pgtype.Bool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "payment term created successfully", paymentTermToOutput(pt))
}

func (uc *PaymentTermsUseCase) GetPaymentTerm(ctx context.Context, idStr string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid ID", nil)
	}

	pt, err := uc.repo.GetPaymentTerm(ctx, int32(id))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "payment term not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "payment term fetched successfully", paymentTermToOutput(pt))
}

func (uc *PaymentTermsUseCase) ListPaymentTerms(ctx context.Context, orgID int32, limit, offset int32, isActive *bool) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var activeParam pgtype.Bool
	if isActive != nil {
		activeParam = pgtype.Bool{Bool: *isActive, Valid: true}
	} else {
		activeParam = pgtype.Bool{Valid: false}
	}

	pts, err := uc.repo.ListPaymentTerms(ctx, repository.ListPaymentTermsParams{
		OrganizationID: orgID,
		Limit:          limit,
		Offset:         offset,
		IsActive:       activeParam,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	count, err := uc.repo.CountPaymentTerms(ctx, repository.CountPaymentTermsParams{
		OrganizationID: orgID,
		IsActive:       activeParam,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	outputs := make([]PaymentTermOutput, len(pts))
	for i, pt := range pts {
		outputs[i] = paymentTermToOutput(pt)
	}

	return utils.NewResponse(utils.CodeOK, "payment terms fetched successfully", map[string]interface{}{
		"items": outputs,
		"total": count,
	})
}

func (uc *PaymentTermsUseCase) UpdatePaymentTerm(ctx context.Context, idStr string, input UpdatePaymentTermInput) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid ID", nil)
	}

	// Fetch existing to get organization_id for unique code checks
	existing, err := uc.repo.GetPaymentTerm(ctx, int32(id))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "payment term not found", nil)
	}

	if input.Code != nil && *input.Code != existing.Code {
		_, err := uc.repo.GetPaymentTermByCode(ctx, repository.GetPaymentTermByCodeParams{
			OrganizationID: existing.OrganizationID,
			Code:           *input.Code,
		})
		if err == nil {
			return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("payment term with code '%s' already exists", *input.Code), nil)
		}
	}

	var codeParam pgtype.Text
	if input.Code != nil {
		codeParam = pgtype.Text{String: *input.Code, Valid: true}
	}

	var nameParam pgtype.Text
	if input.Name != nil {
		nameParam = pgtype.Text{String: *input.Name, Valid: true}
	}

	var dueDaysParam pgtype.Int4
	if input.DueDays != nil {
		dueDaysParam = pgtype.Int4{Int32: *input.DueDays, Valid: true}
	}

	var discountDaysParam pgtype.Int4
	if input.DiscountDays != nil {
		discountDaysParam = pgtype.Int4{Int32: *input.DiscountDays, Valid: true}
	}

	var discountPctParam pgtype.Numeric
	if input.DiscountPercentage != nil {
		discountPctParam = utils.Float64PointerToPgNumeric(input.DiscountPercentage)
	}

	var lateFeePctParam pgtype.Numeric
	if input.LateFeePercentage != nil {
		lateFeePctParam = utils.Float64PointerToPgNumeric(input.LateFeePercentage)
	}

	var isActiveParam pgtype.Bool
	if input.IsActive != nil {
		isActiveParam = pgtype.Bool{Bool: *input.IsActive, Valid: true}
	}

	updated, err := uc.repo.UpdatePaymentTerm(ctx, repository.UpdatePaymentTermParams{
		ID:                 int32(id),
		Code:               codeParam,
		Name:               nameParam,
		DueDays:            dueDaysParam,
		DiscountDays:       discountDaysParam,
		DiscountPercentage: discountPctParam,
		LateFeePercentage:  lateFeePctParam,
		IsActive:           isActiveParam,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "payment term updated successfully", paymentTermToOutput(updated))
}

func (uc *PaymentTermsUseCase) DeletePaymentTerm(ctx context.Context, idStr string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid ID", nil)
	}

	// Verify existence
	_, err = uc.repo.GetPaymentTerm(ctx, int32(id))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "payment term not found", nil)
	}

	err = uc.repo.DeletePaymentTerm(ctx, int32(id))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "payment term deleted successfully", nil)
}
