package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

// ShortageItemInfo records line-item deficit details between Nembus checkout and SAP inventory.
type ShortageItemInfo struct {
	ProductID   int32   `json:"product_id"`
	ProductSKU  string  `json:"product_sku"`
	ProductName string  `json:"product_name"`
	Quantity    float64 `json:"quantity"`
	SAPOnHand   float64 `json:"sap_on_hand"`
}

// SAPInvoiceStockPayload carries held invoice information for stock review.
type SAPInvoiceStockPayload struct {
	TransactionID     int64              `json:"transaction_id"`
	TransactionNumber string             `json:"transaction_number"`
	StoreID           int32              `json:"store_id"`
	WarehouseCode     string             `json:"warehouse_code"`
	TotalAmount       float64            `json:"total_amount"`
	ShortageItems     []ShortageItemInfo `json:"shortage_items"`
	Reason            string             `json:"reason,omitempty"`
}

// SAPInvoicePoster defines the interface for triggering an SAP AR Invoice post.
type SAPInvoicePoster interface {
	PostHeldInvoiceToSAP(ctx context.Context, transactionID int64, force bool) (docEntry int, docNum int, err error)
}

// SAPInvoiceStockAccommodator handles governance decisions for stock-held invoices.
type SAPInvoiceStockAccommodator struct {
	poster SAPInvoicePoster
}

func NewSAPInvoiceStockAccommodator(poster ...SAPInvoicePoster) *SAPInvoiceStockAccommodator {
	var p SAPInvoicePoster
	if len(poster) > 0 {
		p = poster[0]
	}
	return &SAPInvoiceStockAccommodator{poster: p}
}

func (a *SAPInvoiceStockAccommodator) SetPoster(p SAPInvoicePoster) {
	a.poster = p
}

func (a *SAPInvoiceStockAccommodator) Accommodate(ctx context.Context, repo *repository.Queries, req repository.ApprovalRequest, reviewerID int32) error {
	if repo == nil {
		return errors.New("repository is not configured")
	}
	var payload SAPInvoiceStockPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return fmt.Errorf("invalid sap invoice stock payload: %w", err)
	}
	if payload.TransactionID <= 0 {
		return errors.New("transaction_id is required in payload")
	}

	// If poster is available, execute verified post to SAP
	if a.poster != nil {
		docEntry, docNum, err := a.poster.PostHeldInvoiceToSAP(ctx, payload.TransactionID, false)
		if err != nil {
			return fmt.Errorf("failed posting held invoice to SAP: %w", err)
		}
		meta := map[string]any{
			"sap_doc_entry":  docEntry,
			"sap_doc_num":    docNum,
			"sap_sync_state": "synced",
			"sap_posted_at":  time.Now().Format(time.RFC3339),
			"approved_by":    reviewerID,
		}
		rawMeta, _ := json.Marshal(meta)
		_ = repo.UpdatePosTransactionMetadata(ctx, repository.UpdatePosTransactionMetadataParams{
			ID:       int32(payload.TransactionID),
			Metadata: rawMeta,
		})
	} else {
		// Offline / queue approval marking
		meta := map[string]any{
			"sap_sync_state": "approved_pending_upload",
			"approved_by":    reviewerID,
			"approved_at":    time.Now().Format(time.RFC3339),
		}
		rawMeta, _ := json.Marshal(meta)
		_ = repo.UpdatePosTransactionMetadata(ctx, repository.UpdatePosTransactionMetadataParams{
			ID:       int32(payload.TransactionID),
			Metadata: rawMeta,
		})
	}
	return nil
}

func (a *SAPInvoiceStockAccommodator) Reject(ctx context.Context, repo *repository.Queries, req repository.ApprovalRequest, notes string) error {
	if repo == nil {
		return errors.New("repository is not configured")
	}
	var payload SAPInvoiceStockPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return fmt.Errorf("invalid sap invoice stock payload: %w", err)
	}
	if payload.TransactionID <= 0 {
		return errors.New("transaction_id is required in payload")
	}

	// 1. Create internal inventory adjustments for variance
	for _, item := range payload.ShortageItems {
		if item.ProductID > 0 {
			qtyNumeric := pgtype.Numeric{}
			_ = qtyNumeric.Scan(fmt.Sprintf("%.3f", item.Quantity))

			meta := map[string]any{
				"reason":           "SAP stock hold rejected: physical inventory adjustment",
				"transaction_id":   payload.TransactionID,
				"notes":            notes,
				"sap_shortage_qty": item.Quantity - item.SAPOnHand,
			}
			metaBytes, _ := json.Marshal(meta)

			var storeID pgtype.Int4
			if payload.StoreID > 0 {
				storeID = pgtype.Int4{Int32: payload.StoreID, Valid: true}
			}

			_, _ = repo.CreateStockMovement(ctx, repository.CreateStockMovementParams{
				MovementType:  "adjustment",
				ReferenceType: pgtype.Text{String: "approval_rejection", Valid: true},
				ReferenceID:   pgtype.Int4{Int32: int32(req.ID), Valid: true},
				ProductID:     item.ProductID,
				FromStoreID:   storeID,
				Quantity:      qtyNumeric,
				Metadata:      metaBytes,
			})
		}
	}

	// 2. Force-post to SAP if poster configured, otherwise record adjusted state
	if a.poster != nil {
		docEntry, docNum, err := a.poster.PostHeldInvoiceToSAP(ctx, payload.TransactionID, true)
		if err == nil {
			meta := map[string]any{
				"sap_doc_entry":  docEntry,
				"sap_doc_num":    docNum,
				"sap_sync_state": "force_synced_after_adjustment",
				"rejected_notes": notes,
				"sap_posted_at":  time.Now().Format(time.RFC3339),
			}
			rawMeta, _ := json.Marshal(meta)
			_ = repo.UpdatePosTransactionMetadata(ctx, repository.UpdatePosTransactionMetadataParams{
				ID:       int32(payload.TransactionID),
				Metadata: rawMeta,
			})
			return nil
		}
	}

	meta := map[string]any{
		"sap_sync_state": "adjustment_created_pending_force",
		"rejected_notes": notes,
		"adjusted_at":    time.Now().Format(time.RFC3339),
	}
	rawMeta, _ := json.Marshal(meta)
	_ = repo.UpdatePosTransactionMetadata(ctx, repository.UpdatePosTransactionMetadataParams{
		ID:       int32(payload.TransactionID),
		Metadata: rawMeta,
	})
	return nil
}
