package usecase

import (
	"context"
	"errors"
	"strconv"

	"github.com/NasTecSol/nembus-core/repository"
)

// GoodsReceiptNoteApprovalAccommodator completes or rejects an already-created
// GRN. It is registered only for the goods_receipt_note approval domain.
type GoodsReceiptNoteApprovalAccommodator struct{}

func NewGoodsReceiptNoteApprovalAccommodator() *GoodsReceiptNoteApprovalAccommodator {
	return &GoodsReceiptNoteApprovalAccommodator{}
}

func (a *GoodsReceiptNoteApprovalAccommodator) grnID(req repository.ApprovalRequest) (int32, error) {
	if req.Action != "post" || !req.EntityID.Valid {
		return 0, errors.New("approval request is not linked to a GRN posting")
	}
	id, err := strconv.ParseInt(req.EntityID.String, 10, 32)
	if err != nil || id <= 0 {
		return 0, errors.New("approval request contains an invalid GRN id")
	}
	return int32(id), nil
}

func (a *GoodsReceiptNoteApprovalAccommodator) Accommodate(ctx context.Context, repo *repository.Queries, req repository.ApprovalRequest, _ int32) error {
	id, err := a.grnID(req)
	if err != nil {
		return err
	}
	grn, err := repo.GetGoodsReceiptNote(ctx, id)
	if err != nil {
		return errors.New("linked GRN not found")
	}
	if grn.OrganizationID != req.OrganizationID || !grn.Status.Valid || grn.Status.String != "pending_approval" {
		return errors.New("linked GRN is not pending approval in this organization")
	}
	result, err := repo.CallProcessGoodsReceipt(ctx, id)
	if err != nil {
		return err
	}
	if !result.Success {
		return errors.New(result.Message)
	}
	return nil
}

func (a *GoodsReceiptNoteApprovalAccommodator) Reject(ctx context.Context, repo *repository.Queries, req repository.ApprovalRequest, _ string) error {
	id, err := a.grnID(req)
	if err != nil {
		return err
	}
	grn, err := repo.GetGoodsReceiptNote(ctx, id)
	if err != nil {
		return errors.New("linked GRN not found")
	}
	if grn.OrganizationID != req.OrganizationID {
		return errors.New("linked GRN belongs to another organization")
	}
	rows, err := repo.RejectGoodsReceiptNoteApproval(ctx, id)
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("linked GRN is no longer pending approval")
	}
	return nil
}
