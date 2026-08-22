package usecase

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type ProductCategoryUseCase struct {
	repo *repository.Queries
}

func NewProductCategoryUseCase() *ProductCategoryUseCase {
	return &ProductCategoryUseCase{}
}

func (uc *ProductCategoryUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *ProductCategoryUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

type ProductCategoryOutput struct {
	ID               int32           `json:"id"`
	ParentCategoryID *int32          `json:"parent_category_id,omitempty"`
	Name             string          `json:"name"`
	Code             string          `json:"code"`
	Description      string          `json:"description,omitempty"`
	CategoryLevel    int32           `json:"category_level"`
	IsActive         bool            `json:"is_active"`
	Metadata         json.RawMessage `json:"metadata,omitempty"`
	CreatedAt        string          `json:"created_at"`
	UpdatedAt        string          `json:"updated_at"`
}

func categoryToOutput(c repository.ProductCategory) ProductCategoryOutput {
	out := ProductCategoryOutput{
		ID:            c.ID,
		Name:          c.Name,
		Code:          c.Code,
		CategoryLevel: c.CategoryLevel.Int32,
		IsActive:      c.IsActive.Bool,
		Metadata:      utils.BytesToJSONRawMessage(c.Metadata),
		CreatedAt:     utils.FormatTimestamp(c.CreatedAt),
		UpdatedAt:     utils.FormatTimestamp(c.UpdatedAt),
	}

	if c.ParentCategoryID.Valid {
		out.ParentCategoryID = &c.ParentCategoryID.Int32
	}
	if c.Description.Valid {
		out.Description = c.Description.String
	}

	return out
}

func (uc *ProductCategoryUseCase) CreateProductCategory(
	ctx context.Context,
	parentCategoryID *int32,
	name string,
	code string,
	description string,
	level int32,
	isActive bool,
	metadata *json.RawMessage,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	arg := repository.CreateProductCategoryParams{
		Name:          name,
		Code:          code,
		Description:   pgtype.Text{String: description, Valid: description != ""},
		CategoryLevel: pgtype.Int4{Int32: level, Valid: true},
		IsActive:      pgtype.Bool{Bool: isActive, Valid: true},
	}

	if parentCategoryID != nil {
		arg.ParentCategoryID = pgtype.Int4{Int32: *parentCategoryID, Valid: true}
	}

	if metadata != nil {
		arg.Metadata = *metadata
	} else {
		arg.Metadata = []byte("{}")
	}

	category, err := uc.repo.CreateProductCategory(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "category created successfully", categoryToOutput(category))
}

func (uc *ProductCategoryUseCase) GetProductCategory(ctx context.Context, idStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid category id", nil)
	}

	category, err := uc.repo.GetProductCategory(ctx, int32(id))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "category not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "category fetched successfully", categoryToOutput(category))
}

func (uc *ProductCategoryUseCase) GetProductCategoryByCode(ctx context.Context, code string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	category, err := uc.repo.GetProductCategoryByCode(ctx, code)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "category not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "category fetched successfully", categoryToOutput(category))
}

func (uc *ProductCategoryUseCase) ListProductCategories(ctx context.Context, isActive *bool) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	var activeParam pgtype.Bool
	if isActive != nil {
		activeParam = pgtype.Bool{Bool: *isActive, Valid: true}
	}

	categories, err := uc.repo.ListProductCategories(ctx, activeParam)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	output := make([]ProductCategoryOutput, len(categories))
	for i, c := range categories {
		output[i] = categoryToOutput(c)
	}

	return utils.NewResponse(utils.CodeOK, "categories fetched successfully", output)
}

func (uc *ProductCategoryUseCase) ListCategoryChildren(ctx context.Context, parentIDStr string, isActive *bool) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parentID, err := strconv.ParseInt(parentIDStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid parent category id", nil)
	}

	arg := repository.ListCategoryChildrenParams{
		ParentCategoryID: pgtype.Int4{Int32: int32(parentID), Valid: true},
	}

	if isActive != nil {
		arg.IsActive = pgtype.Bool{Bool: *isActive, Valid: true}
	}

	categories, err := uc.repo.ListCategoryChildren(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	output := make([]ProductCategoryOutput, len(categories))
	for i, c := range categories {
		output[i] = categoryToOutput(c)
	}

	return utils.NewResponse(utils.CodeOK, "children fetched successfully", output)
}

func (uc *ProductCategoryUseCase) UpdateProductCategory(
	ctx context.Context,
	idStr string,
	parentCategoryID *int32,
	name *string,
	description *string,
	level *int32,
	isActive *bool,
	metadata *json.RawMessage,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid category id", nil)
	}

	// First get the current category to use for COALESCE replacements
	// Although the SQL uses COALESCE, sqlc.narg works best with pointers or pgtype
	// Our UpdateProductCategory in SQL uses sqlc.narg

	arg := repository.UpdateProductCategoryParams{
		ID: int32(id),
	}

	if parentCategoryID != nil {
		arg.ParentCategoryID = pgtype.Int4{Int32: *parentCategoryID, Valid: true}
	}
	if name != nil {
		arg.Name = pgtype.Text{String: *name, Valid: true}
	}
	if description != nil {
		arg.Description = pgtype.Text{String: *description, Valid: true}
	}
	if level != nil {
		arg.CategoryLevel = pgtype.Int4{Int32: *level, Valid: true}
	}
	if isActive != nil {
		arg.IsActive = pgtype.Bool{Bool: *isActive, Valid: true}
	}
	if metadata != nil {
		arg.Metadata = *metadata
	}

	category, err := uc.repo.UpdateProductCategory(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "category updated successfully", categoryToOutput(category))
}

func (uc *ProductCategoryUseCase) DeleteProductCategory(ctx context.Context, idStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid category id", nil)
	}

	err = uc.repo.DeleteProductCategory(ctx, int32(id))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "category deleted successfully", nil)
}

type GetCategoryHierarchyOutput struct {
	ID               int32           `json:"id"`
	ParentCategoryID *int32          `json:"parent_category_id,omitempty"`
	Name             string          `json:"name"`
	Code             string          `json:"code"`
	Description      string          `json:"description,omitempty"`
	CategoryLevel    int32           `json:"category_level"`
	IsActive         bool            `json:"is_active"`
	Metadata         json.RawMessage `json:"metadata,omitempty"`
	Level            int32           `json:"level"`
	Path             interface{}     `json:"path"`
	FullPath         string          `json:"full_path"`
}

func (uc *ProductCategoryUseCase) GetCategoryHierarchy(ctx context.Context, isActive *bool) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	var activeParam pgtype.Bool
	if isActive != nil {
		activeParam = pgtype.Bool{Bool: *isActive, Valid: true}
	}

	hierarchy, err := uc.repo.GetCategoryHierarchy(ctx, activeParam)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	output := make([]GetCategoryHierarchyOutput, len(hierarchy))
	for i, h := range hierarchy {
		item := GetCategoryHierarchyOutput{
			ID:            h.ID,
			Name:          h.Name,
			Code:          h.Code,
			CategoryLevel: h.CategoryLevel.Int32,
			IsActive:      h.IsActive.Bool,
			Metadata:      utils.BytesToJSONRawMessage(h.Metadata),
			Level:         h.Level,
			Path:          h.Path,
			FullPath:      h.FullPath,
		}

		if h.ParentCategoryID.Valid {
			item.ParentCategoryID = &h.ParentCategoryID.Int32
		}
		if h.Description.Valid {
			item.Description = h.Description.String
		}
		output[i] = item
	}

	return utils.NewResponse(utils.CodeOK, "category hierarchy fetched successfully", output)
}
