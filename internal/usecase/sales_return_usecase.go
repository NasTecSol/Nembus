package usecase

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type SalesReturnUseCase struct {
	repo *repository.Queries
}

func NewSalesReturnUseCase() *SalesReturnUseCase {
	return &SalesReturnUseCase{}
}

func (uc *SalesReturnUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *SalesReturnUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

type ProcessReturnInput struct {
	StoreID               int32
	CashierID             *int32
	SessionID             *int32
	OriginalTransactionID *int32
	CustomerID            *int32
	ReturnReason          string
	Subtotal              string
	TaxAmount             string
	TotalRefundAmount     string
	RefundMethod          string
	RefundReference       string
	Lines                 []ProcessReturnLineInput
}

type ProcessReturnLineInput struct {
	ProductID        int32
	ProductVariantID *int32
	OriginalLineID   *int32
	Quantity         string
	UnitPrice        string
	RefundAmount     string
	ReturnToStock    bool
	SerialNumber     *string
	BatchNumber      *string
	Condition        string
}

func (uc *SalesReturnUseCase) ProcessSalesReturn(ctx context.Context, in ProcessReturnInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	// 1. Prepare Return Header
	returnNumber := fmt.Sprintf("RET-%d", time.Now().Unix())

	subtotal, _ := uc.repo.ParseNumeric(ctx, in.Subtotal)
	taxAmount, _ := uc.repo.ParseNumeric(ctx, in.TaxAmount)
	totalRefund, _ := uc.repo.ParseNumeric(ctx, in.TotalRefundAmount)

	headerParams := repository.CreateSalesReturnParams{
		ReturnNumber:      returnNumber,
		StoreID:           in.StoreID,
		ReturnDate:        pgtype.Timestamp{Time: time.Now(), Valid: true},
		ReturnReason:      pgtype.Text{String: in.ReturnReason, Valid: in.ReturnReason != ""},
		Status:            pgtype.Text{String: "completed", Valid: true},
		Subtotal:          subtotal,
		TaxAmount:         taxAmount,
		TotalRefundAmount: totalRefund,
		RefundMethod:      pgtype.Text{String: in.RefundMethod, Valid: in.RefundMethod != ""},
		RefundReference:   pgtype.Text{String: in.RefundReference, Valid: in.RefundReference != ""},
		Metadata:          nil,
	}

	if in.CashierID != nil {
		headerParams.CashierID = pgtype.Int4{Int32: *in.CashierID, Valid: true}
	}
	if in.SessionID != nil {
		headerParams.CashierSessionID = pgtype.Int4{Int32: *in.SessionID, Valid: true}
	}
	if in.OriginalTransactionID != nil {
		headerParams.OriginalTransactionID = pgtype.Int4{Int32: *in.OriginalTransactionID, Valid: true}
	}
	if in.CustomerID != nil {
		headerParams.CustomerID = pgtype.Int4{Int32: *in.CustomerID, Valid: true}
	}

	salesReturn, err := uc.repo.CreateSalesReturn(ctx, headerParams)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create return header: "+err.Error(), nil)
	}

	// 2. Prepare and Create Return Lines
	for i, l := range in.Lines {
		qty, _ := uc.repo.ParseNumeric(ctx, l.Quantity)
		uPrice, _ := uc.repo.ParseNumeric(ctx, l.UnitPrice)
		refAmount, _ := uc.repo.ParseNumeric(ctx, l.RefundAmount)

		lineParams := repository.CreateSalesReturnLineParams{
			ReturnID:      salesReturn.ID,
			ProductID:     l.ProductID,
			Quantity:      qty,
			UnitPrice:     uPrice,
			RefundAmount:  refAmount,
			ReturnToStock: pgtype.Bool{Bool: l.ReturnToStock, Valid: true},
			Condition:     pgtype.Text{String: l.Condition, Valid: l.Condition != ""},
			LineNumber:    pgtype.Int4{Int32: int32(i + 1), Valid: true},
		}

		if l.ProductVariantID != nil {
			lineParams.ProductVariantID = pgtype.Int4{Int32: *l.ProductVariantID, Valid: true}
		}
		if l.OriginalLineID != nil {
			lineParams.OriginalLineID = pgtype.Int4{Int32: *l.OriginalLineID, Valid: true}
		}
		if l.SerialNumber != nil {
			lineParams.SerialNumber = pgtype.Text{String: *l.SerialNumber, Valid: true}
		}
		if l.BatchNumber != nil {
			lineParams.BatchNumber = pgtype.Text{String: *l.BatchNumber, Valid: true}
		}

		_, err = uc.repo.CreateSalesReturnLine(ctx, lineParams)
		if err != nil {
			fmt.Printf("Warning: failed to create return line %d: %s\n", i+1, err.Error())
		}
	}

	// 3. Update Cashier Session Balance (Negative delta for refund)
	if in.SessionID != nil {
		// Use negative value for refund
		negRefund := totalRefund
		if negRefund.Int != nil {
			negRefund.Int = new(big.Int).Neg(negRefund.Int)
		}

		err = uc.repo.UpdateSessionExpectedBalance(ctx, repository.UpdateSessionExpectedBalanceParams{
			ID:              *in.SessionID,
			ExpectedBalance: negRefund,
		})
		if err != nil {
			fmt.Printf("Error: failed to update session balance for refund: %s\n", err.Error())
		}
	}

	return utils.NewResponse(utils.CodeCreated, "sales return processed successfully", salesReturn)
}
