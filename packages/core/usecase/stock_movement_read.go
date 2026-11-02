package usecase

import (
	"encoding/json"
	"strconv"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

// StockMovementReadOutput supports UUID references in the general and reference-type lists.
// Embedding preserves the existing movement fields; the reference overrides the integer field.
type StockMovementReadOutput struct {
	StockMovementOutput
	ReferenceID json.RawMessage `json:"reference_id"`
}

type StockMovementReadListResponse struct {
	Data       []StockMovementReadOutput `json:"data"`
	TotalCount int64                     `json:"total_count"`
	Page       int32                     `json:"page"`
	Limit      int32                     `json:"limit"`
	TotalPages int32                     `json:"total_pages"`
}

func stockMovementReadToOutput(m repository.ListStockMovementsRow) StockMovementReadOutput {
	return StockMovementReadOutput{
		StockMovementOutput: stockMovementBasicToOutput(repository.StockMovement{
			ID:               m.ID,
			MovementType:     m.MovementType,
			ReferenceType:    m.ReferenceType,
			ProductID:        m.ProductID,
			ProductVariantID: m.ProductVariantID,
			FromStoreID:      m.FromStoreID,
			ToStoreID:        m.ToStoreID,
			FromLocationID:   m.FromLocationID,
			ToLocationID:     m.ToLocationID,
			Quantity:         m.Quantity,
			UomID:            m.UomID,
			BatchNumber:      m.BatchNumber,
			SerialNumber:     m.SerialNumber,
			MovementDate:     m.MovementDate,
			PostedBy:         m.PostedBy,
			Status:           m.Status,
			CostPerUnit:      m.CostPerUnit,
			TotalValue:       m.TotalValue,
			Metadata:         m.Metadata,
			CreatedAt:        m.CreatedAt,
			UpdatedAt:        m.UpdatedAt,
		}),
		ReferenceID: stockMovementReadReferenceJSON(pgtype.Text{String: m.ReferenceID, Valid: m.ReferenceID != ""}),
	}
}

func stockMovementReadReferenceJSON(reference pgtype.Text) json.RawMessage {
	if !reference.Valid {
		return json.RawMessage("null")
	}
	// Keep existing integer references as JSON numbers for current API consumers.
	if numeric, err := strconv.ParseInt(reference.String, 10, 32); err == nil {
		return json.RawMessage(strconv.FormatInt(numeric, 10))
	}
	encoded, _ := json.Marshal(reference.String)
	return encoded
}
