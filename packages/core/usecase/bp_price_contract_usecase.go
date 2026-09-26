package usecase

import (
	"context"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// BPPriceContractOutput represents a business partner price contract with JSON-friendly fields.
type BPPriceContractOutput struct {
	ID                 int32   `json:"id"`
	OrganizationID     int32   `json:"organization_id"`
	PartnerID          int32   `json:"partner_id"`
	ProductID          int32   `json:"product_id"`
	ProductVariantID   *int32  `json:"product_variant_id,omitempty"`
	UOMID              *int32  `json:"uom_id,omitempty"`
	ContractPrice      string  `json:"contract_price"`
	DiscountPercentage string  `json:"discount_percentage"`
	DiscountType       *string `json:"discount_type,omitempty"`
	DiscountAmount     *string `json:"discount_amount,omitempty"`
	MinQuantity        string  `json:"min_quantity"`
	ValidFrom          *string `json:"valid_from,omitempty"`
	ValidTo            *string `json:"valid_to,omitempty"`
	IsActive           bool    `json:"is_active"`
	Notes              *string `json:"notes,omitempty"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
	PartnerName        *string `json:"partner_name,omitempty"`
	PartnerCode        *string `json:"partner_code,omitempty"`
	ProductName        *string `json:"product_name,omitempty"`
	ProductSku         *string `json:"product_sku,omitempty"`
	VariantName        *string `json:"variant_name,omitempty"`
	VariantSku         *string `json:"variant_sku,omitempty"`
	UOMName            *string `json:"uom_name,omitempty"`
	UOMCode            *string `json:"uom_code,omitempty"`
}

func bpPriceContractRowToOutput(row repository.GetBPPriceContractRow) BPPriceContractOutput {
	out := BPPriceContractOutput{
		ID:                 row.ID,
		OrganizationID:     row.OrganizationID,
		PartnerID:          row.PartnerID,
		ProductID:          row.ProductID,
		ContractPrice:      numericToString(row.ContractPrice),
		DiscountPercentage: numericToString(row.DiscountPercentage),
		MinQuantity:        numericToString(row.MinQuantity),
		IsActive:           row.IsActive.Bool,
		CreatedAt:          utils.FormatTimestamp(row.CreatedAt),
		UpdatedAt:          utils.FormatTimestamp(row.UpdatedAt),
		PartnerName:        &row.PartnerName,
		PartnerCode:        &row.PartnerCode,
		ProductName:        &row.ProductName,
		ProductSku:         &row.ProductSku,
	}

	if row.DiscountType.Valid {
		dt := row.DiscountType.String
		out.DiscountType = &dt
	}
	if row.DiscountAmount.Valid {
		da := numericToString(row.DiscountAmount)
		out.DiscountAmount = &da
	}

	if row.ProductVariantID.Valid {
		vID := row.ProductVariantID.Int32
		out.ProductVariantID = &vID
	}
	if row.UomID.Valid {
		uID := row.UomID.Int32
		out.UOMID = &uID
	}
	if row.ValidFrom.Valid {
		vf := utils.FormatDate(row.ValidFrom)
		out.ValidFrom = &vf
	}
	if row.ValidTo.Valid {
		vt := utils.FormatDate(row.ValidTo)
		out.ValidTo = &vt
	}
	if row.Notes.Valid {
		out.Notes = &row.Notes.String
	}
	if row.VariantName.Valid {
		out.VariantName = &row.VariantName.String
	}
	if row.VariantSku.Valid {
		out.VariantSku = &row.VariantSku.String
	}
	if row.UomName.Valid {
		out.UOMName = &row.UomName.String
	}
	if row.UomCode.Valid {
		out.UOMCode = &row.UomCode.String
	}

	return out
}

func bpPriceContractListRowToOutput(row repository.ListBPPriceContractsRow) BPPriceContractOutput {
	out := BPPriceContractOutput{
		ID:                 row.ID,
		OrganizationID:     row.OrganizationID,
		PartnerID:          row.PartnerID,
		ProductID:          row.ProductID,
		ContractPrice:      numericToString(row.ContractPrice),
		DiscountPercentage: numericToString(row.DiscountPercentage),
		MinQuantity:        numericToString(row.MinQuantity),
		IsActive:           row.IsActive.Bool,
		CreatedAt:          utils.FormatTimestamp(row.CreatedAt),
		UpdatedAt:          utils.FormatTimestamp(row.UpdatedAt),
		PartnerName:        &row.PartnerName,
		PartnerCode:        &row.PartnerCode,
		ProductName:        &row.ProductName,
		ProductSku:         &row.ProductSku,
	}

	if row.DiscountType.Valid {
		dt := row.DiscountType.String
		out.DiscountType = &dt
	}
	if row.DiscountAmount.Valid {
		da := numericToString(row.DiscountAmount)
		out.DiscountAmount = &da
	}

	if row.ProductVariantID.Valid {
		vID := row.ProductVariantID.Int32
		out.ProductVariantID = &vID
	}
	if row.UomID.Valid {
		uID := row.UomID.Int32
		out.UOMID = &uID
	}
	if row.ValidFrom.Valid {
		vf := utils.FormatDate(row.ValidFrom)
		out.ValidFrom = &vf
	}
	if row.ValidTo.Valid {
		vt := utils.FormatDate(row.ValidTo)
		out.ValidTo = &vt
	}
	if row.Notes.Valid {
		out.Notes = &row.Notes.String
	}
	if row.VariantName.Valid {
		out.VariantName = &row.VariantName.String
	}
	if row.VariantSku.Valid {
		out.VariantSku = &row.VariantSku.String
	}
	if row.UomName.Valid {
		out.UOMName = &row.UomName.String
	}
	if row.UomCode.Valid {
		out.UOMCode = &row.UomCode.String
	}

	return out
}

func bpPriceContractPartnerRowToOutput(row repository.ListBPPriceContractsByPartnerRow) BPPriceContractOutput {
	out := BPPriceContractOutput{
		ID:                 row.ID,
		OrganizationID:     row.OrganizationID,
		PartnerID:          row.PartnerID,
		ProductID:          row.ProductID,
		ContractPrice:      numericToString(row.ContractPrice),
		DiscountPercentage: numericToString(row.DiscountPercentage),
		MinQuantity:        numericToString(row.MinQuantity),
		IsActive:           row.IsActive.Bool,
		CreatedAt:          utils.FormatTimestamp(row.CreatedAt),
		UpdatedAt:          utils.FormatTimestamp(row.UpdatedAt),
		PartnerName:        &row.PartnerName,
		PartnerCode:        &row.PartnerCode,
		ProductName:        &row.ProductName,
		ProductSku:         &row.ProductSku,
	}

	if row.DiscountType.Valid {
		dt := row.DiscountType.String
		out.DiscountType = &dt
	}
	if row.DiscountAmount.Valid {
		da := numericToString(row.DiscountAmount)
		out.DiscountAmount = &da
	}

	if row.ProductVariantID.Valid {
		vID := row.ProductVariantID.Int32
		out.ProductVariantID = &vID
	}
	if row.UomID.Valid {
		uID := row.UomID.Int32
		out.UOMID = &uID
	}
	if row.ValidFrom.Valid {
		vf := utils.FormatDate(row.ValidFrom)
		out.ValidFrom = &vf
	}
	if row.ValidTo.Valid {
		vt := utils.FormatDate(row.ValidTo)
		out.ValidTo = &vt
	}
	if row.Notes.Valid {
		out.Notes = &row.Notes.String
	}
	if row.VariantName.Valid {
		out.VariantName = &row.VariantName.String
	}
	if row.VariantSku.Valid {
		out.VariantSku = &row.VariantSku.String
	}
	if row.UomName.Valid {
		out.UOMName = &row.UomName.String
	}
	if row.UomCode.Valid {
		out.UOMCode = &row.UomCode.String
	}

	return out
}

func bpPriceContractEffectiveRowToOutput(row repository.GetEffectiveBPPriceContractRow) BPPriceContractOutput {
	out := BPPriceContractOutput{
		ID:                 row.ID,
		OrganizationID:     row.OrganizationID,
		PartnerID:          row.PartnerID,
		ProductID:          row.ProductID,
		ContractPrice:      numericToString(row.ContractPrice),
		DiscountPercentage: numericToString(row.DiscountPercentage),
		MinQuantity:        numericToString(row.MinQuantity),
		IsActive:           row.IsActive.Bool,
		CreatedAt:          utils.FormatTimestamp(row.CreatedAt),
		UpdatedAt:          utils.FormatTimestamp(row.UpdatedAt),
		PartnerName:        &row.PartnerName,
		PartnerCode:        &row.PartnerCode,
		ProductName:        &row.ProductName,
		ProductSku:         &row.ProductSku,
	}

	if row.DiscountType.Valid {
		dt := row.DiscountType.String
		out.DiscountType = &dt
	}
	if row.DiscountAmount.Valid {
		da := numericToString(row.DiscountAmount)
		out.DiscountAmount = &da
	}

	if row.ProductVariantID.Valid {
		vID := row.ProductVariantID.Int32
		out.ProductVariantID = &vID
	}
	if row.UomID.Valid {
		uID := row.UomID.Int32
		out.UOMID = &uID
	}
	if row.ValidFrom.Valid {
		vf := utils.FormatDate(row.ValidFrom)
		out.ValidFrom = &vf
	}
	if row.ValidTo.Valid {
		vt := utils.FormatDate(row.ValidTo)
		out.ValidTo = &vt
	}
	if row.Notes.Valid {
		out.Notes = &row.Notes.String
	}
	if row.VariantName.Valid {
		out.VariantName = &row.VariantName.String
	}
	if row.VariantSku.Valid {
		out.VariantSku = &row.VariantSku.String
	}
	if row.UomName.Valid {
		out.UOMName = &row.UomName.String
	}
	if row.UomCode.Valid {
		out.UOMCode = &row.UomCode.String
	}

	return out
}

func bpPriceContractRawToOutput(c repository.BpPriceContract) BPPriceContractOutput {
	out := BPPriceContractOutput{
		ID:                 c.ID,
		OrganizationID:     c.OrganizationID,
		PartnerID:          c.PartnerID,
		ProductID:          c.ProductID,
		ContractPrice:      numericToString(c.ContractPrice),
		DiscountPercentage: numericToString(c.DiscountPercentage),
		MinQuantity:        numericToString(c.MinQuantity),
		IsActive:           c.IsActive.Bool,
		CreatedAt:          utils.FormatTimestamp(c.CreatedAt),
		UpdatedAt:          utils.FormatTimestamp(c.UpdatedAt),
	}

	if c.DiscountType.Valid {
		dt := c.DiscountType.String
		out.DiscountType = &dt
	}
	if c.DiscountAmount.Valid {
		da := numericToString(c.DiscountAmount)
		out.DiscountAmount = &da
	}

	if c.ProductVariantID.Valid {
		vID := c.ProductVariantID.Int32
		out.ProductVariantID = &vID
	}
	if c.UomID.Valid {
		uID := c.UomID.Int32
		out.UOMID = &uID
	}
	if c.ValidFrom.Valid {
		vf := utils.FormatDate(c.ValidFrom)
		out.ValidFrom = &vf
	}
	if c.ValidTo.Valid {
		vt := utils.FormatDate(c.ValidTo)
		out.ValidTo = &vt
	}
	if c.Notes.Valid {
		out.Notes = &c.Notes.String
	}

	return out
}

type CreateBPPriceContractInput struct {
	OrganizationID     int32    `json:"organization_id"`
	PartnerID          int32    `json:"partner_id"`
	ProductID          int32    `json:"product_id"`
	ProductVariantID   *int32   `json:"product_variant_id"`
	UomID              *int32   `json:"uom_id"`
	ContractPrice      float64  `json:"contract_price"`
	DiscountPercentage *float64 `json:"discount_percentage"`
	DiscountType       *string  `json:"discount_type"`
	DiscountAmount     *float64 `json:"discount_amount"`
	MinQuantity        *float64 `json:"min_quantity"`
	ValidFrom          *string  `json:"valid_from"`
	ValidTo            *string  `json:"valid_to"`
	IsActive           *bool    `json:"is_active"`
	Notes              *string  `json:"notes"`
}

type UpdateBPPriceContractInput struct {
	UomID              *int32   `json:"uom_id"`
	ContractPrice      *float64 `json:"contract_price"`
	DiscountPercentage *float64 `json:"discount_percentage"`
	DiscountType       *string  `json:"discount_type"`
	DiscountAmount     *float64 `json:"discount_amount"`
	MinQuantity        *float64 `json:"min_quantity"`
	ValidFrom          *string  `json:"valid_from"`
	ValidTo            *string  `json:"valid_to"`
	IsActive           *bool    `json:"is_active"`
	Notes              *string  `json:"notes"`
}

type BPPriceContractUseCase struct {
	repo *repository.Queries
}

func NewBPPriceContractUseCase() *BPPriceContractUseCase {
	return &BPPriceContractUseCase{}
}

func (uc *BPPriceContractUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *BPPriceContractUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// CalculateNetPrice calculates net unit price after applying contract discount (percentage or fixed amount).
func CalculateNetPrice(contractPrice float64, discountType string, discountPercentage float64, discountAmount float64) float64 {
	if discountType == "fixed" {
		net := contractPrice - discountAmount
		if net < 0 {
			return 0
		}
		return net
	}
	// Percentage discount
	net := contractPrice * (1.0 - (discountPercentage / 100.0))
	if net < 0 {
		return 0
	}
	return net
}

func (uc *BPPriceContractUseCase) CreateBPPriceContract(ctx context.Context, input CreateBPPriceContractInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	if input.OrganizationID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}
	if input.PartnerID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "partner_id is required", nil)
	}
	if input.ProductID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "product_id is required", nil)
	}
	if input.ContractPrice < 0 {
		return utils.NewResponse(utils.CodeBadReq, "contract_price must be non-negative", nil)
	}

	discountType := "percentage"
	if input.DiscountType != nil && *input.DiscountType != "" {
		discountType = *input.DiscountType
	}
	if discountType != "percentage" && discountType != "fixed" {
		return utils.NewResponse(utils.CodeBadReq, "discount_type must be either 'percentage' or 'fixed'", nil)
	}

	discountAmount := 0.00
	if input.DiscountAmount != nil {
		discountAmount = *input.DiscountAmount
	}
	if discountAmount < 0 {
		return utils.NewResponse(utils.CodeBadReq, "discount_amount must be non-negative", nil)
	}

	if discountType == "fixed" && discountAmount > input.ContractPrice {
		return utils.NewResponse(utils.CodeBadReq, "fixed discount amount cannot exceed contract price", nil)
	}

	// Check if unique contract already exists
	var pvID pgtype.Int4
	if input.ProductVariantID != nil && *input.ProductVariantID > 0 {
		pvID = pgtype.Int4{Int32: *input.ProductVariantID, Valid: true}
	}

	var uomID pgtype.Int4
	if input.UomID != nil && *input.UomID > 0 {
		uomID = pgtype.Int4{Int32: *input.UomID, Valid: true}
	}

	existing, err := uc.repo.GetBPPriceContractByUnique(ctx, repository.GetBPPriceContractByUniqueParams{
		PartnerID:        input.PartnerID,
		ProductID:        input.ProductID,
		ProductVariantID: pvID,
		UomID:            uomID,
	})
	if err == nil && existing.ID > 0 {
		return utils.NewResponse(utils.CodeBadReq, "a price contract already exists for this business partner, product, variant, and uom combination", nil)
	}

	var fromDate pgtype.Date
	if input.ValidFrom != nil && *input.ValidFrom != "" {
		parsed, err := time.Parse("2006-01-02", *input.ValidFrom)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid valid_from date format, expected YYYY-MM-DD", nil)
		}
		fromDate = pgtype.Date{Time: parsed, Valid: true}
	}

	var toDate pgtype.Date
	if input.ValidTo != nil && *input.ValidTo != "" {
		parsed, err := time.Parse("2006-01-02", *input.ValidTo)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid valid_to date format, expected YYYY-MM-DD", nil)
		}
		toDate = pgtype.Date{Time: parsed, Valid: true}
	}

	discount := 0.00
	if input.DiscountPercentage != nil {
		discount = *input.DiscountPercentage
	}

	minQty := 1.000
	if input.MinQuantity != nil && *input.MinQuantity > 0 {
		minQty = *input.MinQuantity
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	var notesText pgtype.Text
	if input.Notes != nil {
		notesText = pgtype.Text{String: *input.Notes, Valid: true}
	}

	created, err := uc.repo.CreateBPPriceContract(ctx, repository.CreateBPPriceContractParams{
		OrganizationID:     input.OrganizationID,
		PartnerID:          input.PartnerID,
		ProductID:          input.ProductID,
		ProductVariantID:   pvID,
		UomID:              uomID,
		ContractPrice:      utils.Float64ToPgNumeric(input.ContractPrice),
		DiscountPercentage: utils.Float64ToPgNumeric(discount),
		DiscountType:       pgtype.Text{String: discountType, Valid: true},
		DiscountAmount:     utils.Float64ToPgNumeric(discountAmount),
		MinQuantity:        utils.Float64ToPgNumeric(minQty),
		ValidFrom:          fromDate,
		ValidTo:            toDate,
		IsActive:           pgtype.Bool{Bool: isActive, Valid: true},
		Notes:              notesText,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	// Fetch full details with join names
	fullRow, err := uc.repo.GetBPPriceContract(ctx, created.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeCreated, "price contract created successfully", bpPriceContractRawToOutput(created))
	}

	return utils.NewResponse(utils.CodeCreated, "price contract created successfully", bpPriceContractRowToOutput(fullRow))
}

func (uc *BPPriceContractUseCase) GetBPPriceContract(ctx context.Context, idStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid price contract ID", nil)
	}

	row, err := uc.repo.GetBPPriceContract(ctx, int32(id))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "price contract not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "price contract fetched successfully", bpPriceContractRowToOutput(row))
}

func (uc *BPPriceContractUseCase) ListBPPriceContracts(ctx context.Context, orgIDStr string, partnerIDStr string, productIDStr string, isActiveStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil || orgID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid or missing organization_id", nil)
	}

	var partnerID pgtype.Int4
	if partnerIDStr != "" {
		if pID, err := strconv.ParseInt(partnerIDStr, 10, 32); err == nil && pID > 0 {
			partnerID = pgtype.Int4{Int32: int32(pID), Valid: true}
		}
	}

	var productID pgtype.Int4
	if productIDStr != "" {
		if prID, err := strconv.ParseInt(productIDStr, 10, 32); err == nil && prID > 0 {
			productID = pgtype.Int4{Int32: int32(prID), Valid: true}
		}
	}

	var isActive pgtype.Bool
	if isActiveStr != "" {
		if act, err := strconv.ParseBool(isActiveStr); err == nil {
			isActive = pgtype.Bool{Bool: act, Valid: true}
		}
	}

	rows, err := uc.repo.ListBPPriceContracts(ctx, repository.ListBPPriceContractsParams{
		OrganizationID: int32(orgID),
		PartnerID:      partnerID,
		ProductID:      productID,
		IsActive:       isActive,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	outputs := make([]BPPriceContractOutput, len(rows))
	for i, r := range rows {
		outputs[i] = bpPriceContractListRowToOutput(r)
	}

	return utils.NewResponse(utils.CodeOK, "price contracts listed successfully", outputs)
}

func (uc *BPPriceContractUseCase) ListBPPriceContractsByPartner(ctx context.Context, partnerIDStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	partnerID, err := strconv.ParseInt(partnerIDStr, 10, 32)
	if err != nil || partnerID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid or missing partner_id", nil)
	}

	rows, err := uc.repo.ListBPPriceContractsByPartner(ctx, int32(partnerID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	outputs := make([]BPPriceContractOutput, len(rows))
	for i, r := range rows {
		outputs[i] = bpPriceContractPartnerRowToOutput(r)
	}

	return utils.NewResponse(utils.CodeOK, "partner price contracts listed successfully", outputs)
}

func (uc *BPPriceContractUseCase) GetEffectiveBPPriceContract(ctx context.Context, partnerIDStr string, productIDStr string, variantIDStr string, uomIDStr string, quantityStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	partnerID, err := strconv.ParseInt(partnerIDStr, 10, 32)
	if err != nil || partnerID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid or missing partner_id", nil)
	}

	productID, err := strconv.ParseInt(productIDStr, 10, 32)
	if err != nil || productID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid or missing product_id", nil)
	}

	var variantID pgtype.Int4
	if variantIDStr != "" {
		if vID, err := strconv.ParseInt(variantIDStr, 10, 32); err == nil && vID > 0 {
			variantID = pgtype.Int4{Int32: int32(vID), Valid: true}
		}
	}

	var uomID pgtype.Int4
	if uomIDStr != "" {
		if uID, err := strconv.ParseInt(uomIDStr, 10, 32); err == nil && uID > 0 {
			uomID = pgtype.Int4{Int32: int32(uID), Valid: true}
		}
	}

	qty := 1.0
	if quantityStr != "" {
		if q, err := strconv.ParseFloat(quantityStr, 64); err == nil && q > 0 {
			qty = q
		}
	}

	row, err := uc.repo.GetEffectiveBPPriceContract(ctx, repository.GetEffectiveBPPriceContractParams{
		PartnerID:        int32(partnerID),
		ProductID:        int32(productID),
		ProductVariantID: variantID,
		UomID:            uomID,
		Quantity:         utils.Float64ToPgNumeric(qty),
	})
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "no effective price contract found for the specified criteria", nil)
	}

	return utils.NewResponse(utils.CodeOK, "effective price contract fetched successfully", bpPriceContractEffectiveRowToOutput(row))
}

func (uc *BPPriceContractUseCase) UpdateBPPriceContract(ctx context.Context, idStr string, input UpdateBPPriceContractInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid price contract ID", nil)
	}

	existing, err := uc.repo.GetBPPriceContractRaw(ctx, int32(id))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "price contract not found", nil)
	}

	uomID := existing.UomID
	if input.UomID != nil {
		if *input.UomID <= 0 {
			uomID = pgtype.Int4{Valid: false}
		} else {
			uomID = pgtype.Int4{Int32: *input.UomID, Valid: true}
		}
	}

	contractPrice := existing.ContractPrice
	contractPriceFloat := 0.0
	if cpVal, err := existing.ContractPrice.Float64Value(); err == nil && cpVal.Valid {
		contractPriceFloat = cpVal.Float64
	}
	if input.ContractPrice != nil {
		if *input.ContractPrice < 0 {
			return utils.NewResponse(utils.CodeBadReq, "contract_price must be non-negative", nil)
		}
		contractPriceFloat = *input.ContractPrice
		contractPrice = utils.Float64ToPgNumeric(*input.ContractPrice)
	}

	discountPercentage := existing.DiscountPercentage
	if input.DiscountPercentage != nil {
		discountPercentage = utils.Float64ToPgNumeric(*input.DiscountPercentage)
	}

	discountTypeStr := "percentage"
	if existing.DiscountType.Valid && existing.DiscountType.String != "" {
		discountTypeStr = existing.DiscountType.String
	}
	if input.DiscountType != nil && *input.DiscountType != "" {
		discountTypeStr = *input.DiscountType
	}
	if discountTypeStr != "percentage" && discountTypeStr != "fixed" {
		return utils.NewResponse(utils.CodeBadReq, "discount_type must be either 'percentage' or 'fixed'", nil)
	}

	discountAmountFloat := 0.0
	if daVal, err := existing.DiscountAmount.Float64Value(); err == nil && daVal.Valid {
		discountAmountFloat = daVal.Float64
	}
	if input.DiscountAmount != nil {
		discountAmountFloat = *input.DiscountAmount
	}
	if discountAmountFloat < 0 {
		return utils.NewResponse(utils.CodeBadReq, "discount_amount must be non-negative", nil)
	}

	if discountTypeStr == "fixed" && discountAmountFloat > contractPriceFloat {
		return utils.NewResponse(utils.CodeBadReq, "fixed discount amount cannot exceed contract price", nil)
	}

	discountType := pgtype.Text{String: discountTypeStr, Valid: true}
	discountAmount := utils.Float64ToPgNumeric(discountAmountFloat)

	minQuantity := existing.MinQuantity
	if input.MinQuantity != nil {
		minQuantity = utils.Float64ToPgNumeric(*input.MinQuantity)
	}

	validFrom := existing.ValidFrom
	if input.ValidFrom != nil {
		if *input.ValidFrom == "" {
			validFrom = pgtype.Date{Valid: false}
		} else {
			parsed, err := time.Parse("2006-01-02", *input.ValidFrom)
			if err != nil {
				return utils.NewResponse(utils.CodeBadReq, "invalid valid_from date format, expected YYYY-MM-DD", nil)
			}
			validFrom = pgtype.Date{Time: parsed, Valid: true}
		}
	}

	validTo := existing.ValidTo
	if input.ValidTo != nil {
		if *input.ValidTo == "" {
			validTo = pgtype.Date{Valid: false}
		} else {
			parsed, err := time.Parse("2006-01-02", *input.ValidTo)
			if err != nil {
				return utils.NewResponse(utils.CodeBadReq, "invalid valid_to date format, expected YYYY-MM-DD", nil)
			}
			validTo = pgtype.Date{Time: parsed, Valid: true}
		}
	}

	isActive := existing.IsActive
	if input.IsActive != nil {
		isActive = pgtype.Bool{Bool: *input.IsActive, Valid: true}
	}

	notes := existing.Notes
	if input.Notes != nil {
		notes = pgtype.Text{String: *input.Notes, Valid: true}
	}

	updated, err := uc.repo.UpdateBPPriceContract(ctx, repository.UpdateBPPriceContractParams{
		ID:                 int32(id),
		UomID:              uomID,
		ContractPrice:      contractPrice,
		DiscountPercentage: discountPercentage,
		DiscountType:       discountType,
		DiscountAmount:     discountAmount,
		MinQuantity:        minQuantity,
		ValidFrom:          validFrom,
		ValidTo:            validTo,
		IsActive:           isActive,
		Notes:              notes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	fullRow, err := uc.repo.GetBPPriceContract(ctx, updated.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeOK, "price contract updated successfully", bpPriceContractRawToOutput(updated))
	}

	return utils.NewResponse(utils.CodeOK, "price contract updated successfully", bpPriceContractRowToOutput(fullRow))
}

func (uc *BPPriceContractUseCase) DeleteBPPriceContract(ctx context.Context, idStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid price contract ID", nil)
	}

	_, err = uc.repo.GetBPPriceContractRaw(ctx, int32(id))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "price contract not found", nil)
	}

	err = uc.repo.DeleteBPPriceContract(ctx, int32(id))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "price contract deleted successfully", nil)
}

func (uc *BPPriceContractUseCase) ToggleBPPriceContractActive(ctx context.Context, idStr string, isActive bool) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid price contract ID", nil)
	}

	_, err = uc.repo.GetBPPriceContractRaw(ctx, int32(id))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "price contract not found", nil)
	}

	updated, err := uc.repo.ToggleBPPriceContractActive(ctx, repository.ToggleBPPriceContractActiveParams{
		ID:       int32(id),
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	fullRow, err := uc.repo.GetBPPriceContract(ctx, updated.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeOK, "price contract status updated successfully", bpPriceContractRawToOutput(updated))
	}

	return utils.NewResponse(utils.CodeOK, "price contract status updated successfully", bpPriceContractRowToOutput(fullRow))
}
