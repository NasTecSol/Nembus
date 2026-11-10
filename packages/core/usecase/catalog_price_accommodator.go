package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

// CatalogPriceChangePayload carries the proposal details for price or cost price changes.
type CatalogPriceChangePayload struct {
	ProductID    int32    `json:"product_id"`
	ProductSKU   string   `json:"product_sku"`
	PriceListID  int32    `json:"price_list_id"`
	UoMID        *int32   `json:"uom_id,omitempty"`
	OldPrice     *float64 `json:"old_price,omitempty"`
	NewPrice     *float64 `json:"new_price,omitempty"`
	OldCostPrice *float64 `json:"old_cost_price,omitempty"`
	NewCostPrice *float64 `json:"new_cost_price,omitempty"`
	Reason       string   `json:"reason,omitempty"`
}

// CatalogPriceAccommodator handles governance decisions for catalog price and cost price revisions.
type CatalogPriceAccommodator struct{}

func NewCatalogPriceAccommodator() *CatalogPriceAccommodator {
	return &CatalogPriceAccommodator{}
}

func (a *CatalogPriceAccommodator) Accommodate(ctx context.Context, repo *repository.Queries, req repository.ApprovalRequest, reviewerID int32) error {
	if repo == nil {
		return errors.New("repository is not configured")
	}
	var payload CatalogPriceChangePayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return fmt.Errorf("invalid catalog price change payload: %w", err)
	}

	if payload.ProductID <= 0 {
		return errors.New("product_id is required in payload")
	}

	// 1. If NewPrice is supplied and >= 0, upsert into product_prices
	if payload.NewPrice != nil && *payload.NewPrice >= 0 {
		if payload.PriceListID <= 0 {
			return errors.New("price_list_id is required for price updates")
		}
		var uom pgtype.Int4
		if payload.UoMID != nil {
			uom = pgtype.Int4{Int32: *payload.UoMID, Valid: true}
		}
		var priceNumeric pgtype.Numeric
		if err := priceNumeric.Scan(fmt.Sprintf("%.2f", *payload.NewPrice)); err != nil {
			return fmt.Errorf("failed scanning new price: %w", err)
		}
		_, err := repo.UpsertProductPrice(ctx, repository.UpsertProductPriceParams{
			ProductID:   payload.ProductID,
			PriceListID: payload.PriceListID,
			UomID:       uom,
			Price:       priceNumeric,
		})
		if err != nil {
			return fmt.Errorf("failed updating product price: %w", err)
		}
	}

	// 2. If NewCostPrice is supplied and >= 0, update products.cost_price
	if payload.NewCostPrice != nil && *payload.NewCostPrice >= 0 {
		var costNumeric pgtype.Numeric
		if err := costNumeric.Scan(fmt.Sprintf("%.4f", *payload.NewCostPrice)); err != nil {
			return fmt.Errorf("failed scanning new cost price: %w", err)
		}
		_, err := repo.UpdateProductCostPrice(ctx, repository.UpdateProductCostPriceParams{
			ID:        payload.ProductID,
			CostPrice: costNumeric,
		})
		if err != nil {
			return fmt.Errorf("failed updating product cost price: %w", err)
		}
	}

	return nil
}

func (a *CatalogPriceAccommodator) Reject(ctx context.Context, repo *repository.Queries, req repository.ApprovalRequest, notes string) error {
	// Active catalog remains unchanged; rejection is recorded in the audit log by ApprovalsUseCase.
	return nil
}
