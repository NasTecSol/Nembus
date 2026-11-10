package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type RFQListFilter struct {
	OrganizationID int32
	Status         string
	Timeline       string
	Search         string
	HasResponse    string
	SupplierType   string
	Page           int32
	Limit          int32
}

type RFQListResult struct {
	Items      []repository.ListRFQsRow `json:"items"`
	Total      int64                    `json:"total"`
	Page       int32                    `json:"page"`
	Limit      int32                    `json:"limit"`
	TotalPages int64                    `json:"total_pages"`
}

type RFQUseCase struct{}

func NewRFQUseCase() *RFQUseCase { return &RFQUseCase{} }

type RFQItemInput struct {
	ProductID        int32    `json:"product_id" example:"1"`
	ProductVariantID *int32   `json:"product_variant_id"`
	ItemName         string   `json:"item_name" example:"ALMARAI-MILK-FW-1L"`
	ItemDescription  string   `json:"item_description" example:"ALMARAI-MILK-FW-1L lumbar support"`
	Quantity         float64  `json:"quantity" example:"20"`
	UomID            int32    `json:"uom_id" example:"3"`
	TargetPrice      *float64 `json:"target_price" example:"25000"`
	Notes            string   `json:"notes" example:"Black preferred"`
}

type CreateRFQInput struct {
	OrganizationID   int32          `json:"organization_id" binding:"required" example:"1"`
	VendorName       string         `json:"vendor_name" binding:"required" example:"Example Vendor Co."`
	Title            string         `json:"title" example:"500 ALMARAI-MILK-FW-1L SAUDIYA BRANCH"`
	Description      string         `json:"description" example:"Please quote the requested ALMARAI-MILK-FW-1L for delivery next month."`
	StoreID          *int32         `json:"store_id" example:"1"`
	InvitedPartnerID *int32         `json:"invited_partner_id"`
	Items            []RFQItemInput `json:"items"`
}

func nullableInt4(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

func (uc *RFQUseCase) Create(ctx context.Context, repo *repository.Queries, orgID int32, userID *int32, in CreateRFQInput) *repository.Response {
	in.VendorName = strings.TrimSpace(in.VendorName)
	in.Title = strings.TrimSpace(in.Title)
	if orgID <= 0 || in.OrganizationID != orgID {
		return utils.NewResponse(utils.CodeBadReq, "organization_id must match the authenticated organization", nil)
	}
	if in.VendorName == "" || in.Title == "" || len(in.Items) == 0 {
		return utils.NewResponse(utils.CodeBadReq, "vendor_name, title, and at least one item are required", nil)
	}
	for i := range in.Items {
		in.Items[i].ItemName = strings.TrimSpace(in.Items[i].ItemName)
		if in.Items[i].ItemName == "" || in.Items[i].ProductID <= 0 || in.Items[i].UomID <= 0 || in.Items[i].Quantity <= 0 {
			return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("item %d requires product_id, uom_id, name, and positive quantity", i+1), nil)
		}
	}
	itemJSON, err := json.Marshal(in.Items)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid item data", nil)
	}
	internalToken, err := uuid.NewRandom()
	if err != nil {
		return utils.NewResponse(utils.CodeError, "could not generate RFQ token", nil)
	}
	numberID, err := uuid.NewRandom()
	if err != nil {
		return utils.NewResponse(utils.CodeError, "could not generate RFQ number", nil)
	}
	now := time.Now()
	creator := nullableInt4(userID)
	row, err := repo.CreateRFQWithItems(ctx, repository.CreateRFQWithItemsParams{
		OrganizationID: orgID, StoreID: nullableInt4(in.StoreID),
		RfqNumber: "RFQ-" + now.Format("20060102") + "-" + strings.ToUpper(numberID.String()[:8]),
		Title:     in.Title, Description: pgtype.Text{String: strings.TrimSpace(in.Description), Valid: strings.TrimSpace(in.Description) != ""},
		VendorName: in.VendorName,
		Token:      internalToken.String(), TokenValidFrom: pgtype.Timestamp{Time: now, Valid: true},
		TokenExpiresAt:   pgtype.Timestamp{Time: now.Add(7 * 24 * time.Hour), Valid: true},
		InvitedPartnerID: nullableInt4(in.InvitedPartnerID), CreatedByUserID: creator,
		Column12: itemJSON,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return utils.NewResponse(utils.CodeBadReq, "each product must belong to this organization and use its base or configured product UOM", nil)
		}
		return utils.NewResponse(utils.CodeError, "failed to create RFQ", nil)
	}
	return utils.NewResponse(201, "RFQ draft created", row)
}

func (uc *RFQUseCase) Publish(ctx context.Context, repo *repository.Queries, orgID int32, id uuid.UUID, expiresAt time.Time) *repository.Response {
	if !expiresAt.After(time.Now()) {
		return utils.NewResponse(utils.CodeBadReq, "expires_at must be in the future", nil)
	}
	token, err := uuid.NewRandom()
	if err != nil {
		return utils.NewResponse(utils.CodeError, "could not generate public token", nil)
	}
	row, err := repo.PublishRFQ(ctx, repository.PublishRFQParams{ID: id, OrganizationID: orgID, Token: token.String(), TokenExpiresAt: pgtype.Timestamp{Time: expiresAt, Valid: true}})
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "RFQ not found or is not a draft", nil)
	}
	return utils.NewResponse(200, "RFQ published", row)
}

func (uc *RFQUseCase) Close(ctx context.Context, repo *repository.Queries, orgID int32, id uuid.UUID) *repository.Response {
	row, err := repo.CloseRFQ(ctx, repository.CloseRFQParams{ID: id, OrganizationID: orgID})
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "RFQ not found or is not published", nil)
	}
	return utils.NewResponse(200, "RFQ closed", row)
}

type RFQQuoteItemInput struct {
	RFQItemID            string  `json:"rfq_item_id" example:"362d4816-773b-4f0c-9b0d-f7660b5a4e2f"`
	ProductID            int32   `json:"product_id" example:"1"`
	UomID                int32   `json:"uom_id" example:"3"`
	Quantity             float64 `json:"quantity" example:"20"`
	UnitPrice            float64 `json:"unit_price" example:"22000"`
	TaxAmount            float64 `json:"tax_amount" example:"1"`
	DeliveryLeadTimeDays *int32  `json:"delivery_lead_time_days" example:"21"`
	Notes                string  `json:"notes" example:"Price per Liter"`
}

type SubmitRFQResponseInput struct {
	SupplierName  string              `json:"supplier_name" example:"Example Supply Co."`
	SupplierEmail string              `json:"supplier_email" example:"quotes@example.com"`
	SupplierPhone string              `json:"supplier_phone" example:"+92-300-1234567"`
	ValidUntil    string              `json:"valid_until" example:"2026-11-30"`
	Notes         string              `json:"notes" example:"We can deliver in two shipments."`
	Items         []RFQQuoteItemInput `json:"items"`
}

func (uc *RFQUseCase) Submit(ctx context.Context, repo *repository.Queries, rfq repository.GetRFQByTokenRow, in SubmitRFQResponseInput) *repository.Response {
	if rfq.Status != "published" {
		return utils.NewResponse(utils.CodeBadReq, "this RFQ is not accepting responses", nil)
	}
	in.SupplierName = strings.TrimSpace(in.SupplierName)
	in.SupplierEmail = strings.TrimSpace(in.SupplierEmail)
	if in.SupplierName == "" || in.SupplierEmail == "" || len(in.Items) == 0 {
		return utils.NewResponse(utils.CodeBadReq, "supplier name, email, and item quotes are required", nil)
	}
	rfqItems, err := repo.ListRFQItems(ctx, rfq.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to load RFQ items", nil)
	}
	type requestedProduct struct{ productID, uomID int32 }
	validItems := make(map[string]requestedProduct, len(rfqItems))
	for _, item := range rfqItems {
		if item.ProductID.Valid && item.UomID.Valid {
			validItems[item.ID.String()] = requestedProduct{productID: item.ProductID.Int32, uomID: item.UomID.Int32}
		}
	}
	if len(validItems) != len(rfqItems) {
		return utils.NewResponse(utils.CodeBadReq, "each requested item must have a product and UOM", nil)
	}
	seen := make(map[string]bool, len(in.Items))
	var total float64
	for _, item := range in.Items {
		id, parseErr := uuid.Parse(item.RFQItemID)
		requested := validItems[id.String()]
		if parseErr != nil || requested.productID == 0 || requested.productID != item.ProductID || requested.uomID != item.UomID || seen[id.String()] || item.Quantity <= 0 || item.UnitPrice < 0 || item.TaxAmount < 0 {
			return utils.NewResponse(utils.CodeBadReq, "each quote must include the matching product_id, uom_id, and RFQ item, once, with nonnegative price and tax and positive quantity", nil)
		}
		seen[id.String()] = true
		total += item.Quantity*item.UnitPrice + item.TaxAmount
	}
	if len(seen) != len(validItems) {
		return utils.NewResponse(utils.CodeBadReq, "a quote is required for every requested item", nil)
	}
	var validUntil pgtype.Date
	if in.ValidUntil != "" {
		parsed, parseErr := time.Parse("2006-01-02", in.ValidUntil)
		if parseErr != nil {
			return utils.NewResponse(utils.CodeBadReq, "valid_until must use YYYY-MM-DD", nil)
		}
		validUntil = pgtype.Date{Time: parsed, Valid: true}
	}
	itemsJSON, err := json.Marshal(in.Items)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid quote item data", nil)
	}
	var totalAmount pgtype.Numeric
	if err := totalAmount.Scan(fmt.Sprintf("%.4f", total)); err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid quote amount", nil)
	}
	row, err := repo.SubmitRFQResponse(ctx, repository.SubmitRFQResponseParams{
		RfqID: rfq.ID, OrganizationID: rfq.OrganizationID, SupplierName: in.SupplierName,
		Lower: in.SupplierEmail, SupplierPhone: pgtype.Text{String: strings.TrimSpace(in.SupplierPhone), Valid: strings.TrimSpace(in.SupplierPhone) != ""},
		TotalQuotedAmount: totalAmount, ValidUntil: validUntil,
		Notes: pgtype.Text{String: strings.TrimSpace(in.Notes), Valid: strings.TrimSpace(in.Notes) != ""}, Column9: itemsJSON,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to submit supplier response", nil)
	}
	return utils.NewResponse(201, "supplier response submitted", row)
}

// List returns RFQs and computed token timeline/submission information. It never updates RFQ status.
func (uc *RFQUseCase) List(ctx context.Context, repo *repository.Queries, filter RFQListFilter) *repository.Response {
	if filter.OrganizationID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}
	args := repository.ListRFQsParams{
		OrganizationID: filter.OrganizationID, Column2: filter.Status, Column3: filter.Timeline,
		Column4: filter.Search, Column5: filter.HasResponse, Column6: filter.SupplierType,
		Limit: filter.Limit, Offset: (filter.Page - 1) * filter.Limit,
	}
	items, err := repo.ListRFQs(ctx, args)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list RFQs", nil)
	}
	total, err := repo.CountRFQs(ctx, repository.CountRFQsParams{
		OrganizationID: filter.OrganizationID, Column2: filter.Status, Column3: filter.Timeline,
		Column4: filter.Search, Column5: filter.HasResponse, Column6: filter.SupplierType,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to count RFQs", nil)
	}
	return utils.NewResponse(200, "RFQs retrieved successfully", RFQListResult{
		Items: items, Total: total, Page: filter.Page, Limit: filter.Limit,
		TotalPages: int64(math.Ceil(float64(total) / float64(filter.Limit))),
	})
}
