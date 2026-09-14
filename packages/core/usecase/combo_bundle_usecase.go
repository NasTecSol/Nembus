package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// ComboBundleOutput is the API response model for a combo bundle.
type ComboBundleOutput struct {
	ID                      int32                        `json:"id"`
	OrganizationID          int32                        `json:"organization_id"`
	StoreID                 *int32                       `json:"store_id,omitempty"`
	PriceListID             *int32                       `json:"price_list_id,omitempty"`
	ApplicableCustomerTypes []string                     `json:"applicable_customer_types"`
	Code                    string                       `json:"code"`
	Name                    string                       `json:"name"`
	Description             string                       `json:"description,omitempty"`
	BundlePrice             string                       `json:"bundle_price"`
	BundleType              string                       `json:"bundle_type"`
	IsActive                bool                         `json:"is_active"`
	ValidFrom               *string                      `json:"valid_from,omitempty"`
	ValidTo                 *string                      `json:"valid_to,omitempty"`
	DisplayOrder            int32                        `json:"display_order"`
	Metadata                json.RawMessage              `json:"metadata"`
	CreatedAt               string                       `json:"created_at"`
	UpdatedAt               string                       `json:"updated_at"`
	Items                   []ComboBundleItemOutput      `json:"items,omitempty"`
}

// ComboBundleItemOutput is the API response model for a combo bundle item.
type ComboBundleItemOutput struct {
	ID               int32           `json:"id"`
	ComboBundleID    int32           `json:"combo_bundle_id"`
	MenuItemID       *int32          `json:"menu_item_id,omitempty"`
	MenuItemName     *string         `json:"menu_item_name,omitempty"`
	MenuItemPrice    *string         `json:"menu_item_price,omitempty"`
	ProductID        *int32          `json:"product_id,omitempty"`
	ProductName      *string         `json:"product_name,omitempty"`
	ProductSKU       *string         `json:"product_sku,omitempty"`
	ProductVariantID *int32          `json:"product_variant_id,omitempty"`
	VariantName      *string         `json:"variant_name,omitempty"`
	VariantSKU       *string         `json:"variant_sku,omitempty"`
	ItemType         string          `json:"item_type"`
	Quantity         string          `json:"quantity"`
	IsRequired       bool            `json:"is_required"`
	GroupTag         string          `json:"group_tag,omitempty"`
	PriceOverride    *string         `json:"price_override,omitempty"`
	DisplayOrder     int32           `json:"display_order"`
	Metadata         json.RawMessage `json:"metadata"`
}

type ComboBundleUseCase struct {
	repo *repository.Queries
}

func NewComboBundleUseCase() *ComboBundleUseCase {
	return &ComboBundleUseCase{}
}

func (uc *ComboBundleUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *ComboBundleUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// ─── Bundle CRUD ─────────────────────────────────────────────────────────────

func (uc *ComboBundleUseCase) CreateComboBundle(ctx context.Context, arg repository.CreateComboBundleParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	bundle, err := uc.repo.CreateComboBundle(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create combo bundle: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "combo bundle created successfully", bundleToOutput(bundle))
}

func (uc *ComboBundleUseCase) GetComboBundle(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	bundle, err := uc.repo.GetComboBundle(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "combo bundle not found", nil)
	}

	out := bundleToOutput(bundle)
	// Fetch detailed items
	items, err := uc.repo.GetComboBundleWithItems(ctx, id)
	if err == nil && len(items) > 0 {
		out.Items = make([]ComboBundleItemOutput, len(items))
		for i, it := range items {
			out.Items[i] = withItemsRowToOutput(it)
		}
	}
	return utils.NewResponse(utils.CodeOK, "combo bundle fetched successfully", out)
}

func (uc *ComboBundleUseCase) GetComboBundleByCode(ctx context.Context, orgID int32, code string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	bundle, err := uc.repo.GetComboBundleByCode(ctx, repository.GetComboBundleByCodeParams{
		OrganizationID: orgID,
		Code:           code,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "combo bundle not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "combo bundle fetched successfully", bundleToOutput(bundle))
}

func (uc *ComboBundleUseCase) ListComboBundlesByOrg(ctx context.Context, arg repository.ListComboBundlesByOrganizationParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	bundles, err := uc.repo.ListComboBundlesByOrganization(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list combo bundles: "+err.Error(), nil)
	}
	outputs := make([]ComboBundleOutput, len(bundles))
	for i, b := range bundles {
		outputs[i] = bundleToOutput(b)
	}
	return utils.NewResponse(utils.CodeOK, "combo bundles fetched successfully", outputs)
}

func (uc *ComboBundleUseCase) ListComboBundlesByStore(ctx context.Context, storeID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	bundles, err := uc.repo.ListComboBundlesByStore(ctx, pgtype.Int4{Int32: storeID, Valid: true})
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list store combo bundles: "+err.Error(), nil)
	}
	outputs := make([]ComboBundleOutput, len(bundles))
	for i, b := range bundles {
		outputs[i] = bundleToOutput(b)
	}
	return utils.NewResponse(utils.CodeOK, "store combo bundles fetched successfully", outputs)
}

func (uc *ComboBundleUseCase) UpdateComboBundle(ctx context.Context, arg repository.UpdateComboBundleParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	bundle, err := uc.repo.UpdateComboBundle(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update combo bundle: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "combo bundle updated successfully", bundleToOutput(bundle))
}

func (uc *ComboBundleUseCase) ToggleComboBundleActive(ctx context.Context, id int32, isActive bool) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	bundle, err := uc.repo.ToggleComboBundleActive(ctx, repository.ToggleComboBundleActiveParams{
		ID:       id,
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to toggle combo bundle status: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "combo bundle status updated successfully", bundleToOutput(bundle))
}

func (uc *ComboBundleUseCase) DeleteComboBundle(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if err := uc.repo.DeleteComboBundle(ctx, id); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to delete combo bundle: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "combo bundle deleted successfully", nil)
}

// ─── Bundle Item Management ──────────────────────────────────────────────────

func (uc *ComboBundleUseCase) AddComboBundleItem(ctx context.Context, arg repository.CreateComboBundleItemParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	item, err := uc.repo.CreateComboBundleItem(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to add item to combo bundle: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "combo bundle item added successfully", itemToOutput(item))
}

func (uc *ComboBundleUseCase) ListComboBundleItems(ctx context.Context, bundleID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	items, err := uc.repo.GetComboBundleWithItems(ctx, bundleID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list combo bundle items: "+err.Error(), nil)
	}
	outputs := make([]ComboBundleItemOutput, len(items))
	for i, it := range items {
		outputs[i] = withItemsRowToOutput(it)
	}
	return utils.NewResponse(utils.CodeOK, "combo bundle items fetched successfully", outputs)
}

func (uc *ComboBundleUseCase) UpdateComboBundleItem(ctx context.Context, arg repository.UpdateComboBundleItemParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	item, err := uc.repo.UpdateComboBundleItem(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update combo bundle item: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "combo bundle item updated successfully", itemToOutput(item))
}

func (uc *ComboBundleUseCase) DeleteComboBundleItem(ctx context.Context, itemID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if err := uc.repo.DeleteComboBundleItem(ctx, itemID); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to delete combo bundle item: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "combo bundle item deleted successfully", nil)
}

// BatchSetComboBundleItems atomically replaces all items of a bundle.
func (uc *ComboBundleUseCase) BatchSetComboBundleItems(ctx context.Context, bundleID int32, items []repository.CreateComboBundleItemParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	// Delete existing items
	if err := uc.repo.DeleteComboBundleItemsByBundle(ctx, bundleID); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to clear existing combo bundle items: "+err.Error(), nil)
	}

	// Insert new items
	for _, it := range items {
		it.ComboBundleID = bundleID
		if _, err := uc.repo.CreateComboBundleItem(ctx, it); err != nil {
			return utils.NewResponse(utils.CodeError, "failed to insert combo bundle item: "+err.Error(), nil)
		}
	}

	return uc.ListComboBundleItems(ctx, bundleID)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func bundleToOutput(b repository.ComboBundle) ComboBundleOutput {
	out := ComboBundleOutput{
		ID:                      b.ID,
		OrganizationID:          b.OrganizationID,
		ApplicableCustomerTypes: b.ApplicableCustomerTypes,
		Code:                    b.Code,
		Name:                    b.Name,
		DisplayOrder:            b.DisplayOrder.Int32,
		Metadata:                b.Metadata,
	}

	if b.StoreID.Valid {
		v := b.StoreID.Int32
		out.StoreID = &v
	}
	if b.PriceListID.Valid {
		v := b.PriceListID.Int32
		out.PriceListID = &v
	}
	if b.Description.Valid {
		out.Description = b.Description.String
	}
	if b.BundleType.Valid {
		out.BundleType = b.BundleType.String
	}
	if b.IsActive.Valid {
		out.IsActive = b.IsActive.Bool
	}
	if b.ValidFrom.Valid {
		s := fmt.Sprintf("%04d-%02d-%02d", b.ValidFrom.Time.Year(), b.ValidFrom.Time.Month(), b.ValidFrom.Time.Day())
		out.ValidFrom = &s
	}
	if b.ValidTo.Valid {
		s := fmt.Sprintf("%04d-%02d-%02d", b.ValidTo.Time.Year(), b.ValidTo.Time.Month(), b.ValidTo.Time.Day())
		out.ValidTo = &s
	}
	if b.CreatedAt.Valid {
		out.CreatedAt = b.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	if b.UpdatedAt.Valid {
		out.UpdatedAt = b.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}

	out.BundlePrice = numericToString(b.BundlePrice)
	return out
}

func itemToOutput(it repository.ComboBundleItem) ComboBundleItemOutput {
	out := ComboBundleItemOutput{
		ID:            it.ID,
		ComboBundleID: it.ComboBundleID,
		DisplayOrder:  it.DisplayOrder.Int32,
		Metadata:      it.Metadata,
	}
	if it.MenuItemID.Valid {
		v := it.MenuItemID.Int32
		out.MenuItemID = &v
	}
	if it.ProductID.Valid {
		v := it.ProductID.Int32
		out.ProductID = &v
	}
	if it.ProductVariantID.Valid {
		v := it.ProductVariantID.Int32
		out.ProductVariantID = &v
	}
	if it.ItemType.Valid {
		out.ItemType = it.ItemType.String
	}
	if it.IsRequired.Valid {
		out.IsRequired = it.IsRequired.Bool
	}
	if it.GroupTag.Valid {
		out.GroupTag = it.GroupTag.String
	}
	if it.PriceOverride.Valid {
		s := numericToString(it.PriceOverride)
		out.PriceOverride = &s
	}
	out.Quantity = numericToString(it.Quantity)
	return out
}

func withItemsRowToOutput(r repository.GetComboBundleWithItemsRow) ComboBundleItemOutput {
	out := ComboBundleItemOutput{
		ID:            r.ItemID,
		ComboBundleID: r.ComboBundleID,
		DisplayOrder:  r.DisplayOrder.Int32,
	}
	if r.MenuItemID.Valid {
		v := r.MenuItemID.Int32
		out.MenuItemID = &v
	}
	if r.MenuItemName.Valid {
		out.MenuItemName = &r.MenuItemName.String
	}
	if r.MenuItemPrice.Valid {
		s := numericToString(r.MenuItemPrice)
		out.MenuItemPrice = &s
	}
	if r.ProductID.Valid {
		v := r.ProductID.Int32
		out.ProductID = &v
	}
	if r.ProductName.Valid {
		out.ProductName = &r.ProductName.String
	}
	if r.ProductSku.Valid {
		out.ProductSKU = &r.ProductSku.String
	}
	if r.ProductVariantID.Valid {
		v := r.ProductVariantID.Int32
		out.ProductVariantID = &v
	}
	if r.VariantName.Valid {
		out.VariantName = &r.VariantName.String
	}
	if r.VariantSku.Valid {
		out.VariantSKU = &r.VariantSku.String
	}
	if r.ItemType.Valid {
		out.ItemType = r.ItemType.String
	}
	if r.IsRequired.Valid {
		out.IsRequired = r.IsRequired.Bool
	}
	if r.GroupTag.Valid {
		out.GroupTag = r.GroupTag.String
	}
	if r.PriceOverride.Valid {
		s := numericToString(r.PriceOverride)
		out.PriceOverride = &s
	}
	out.Quantity = numericToString(r.Quantity)
	return out
}
