package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/jackc/pgx/v5/pgtype"
)

// ProductEnrichmentStore adapts existing generated repository queries to the
// provider-neutral Stage 2A enqueue contract. It never calls a provider and
// never mutates products or taxonomy records.
type ProductEnrichmentStore struct {
	queries *Queries
}

func NewProductEnrichmentStore(queries *Queries) *ProductEnrichmentStore {
	return &ProductEnrichmentStore{queries: queries}
}

func (s *ProductEnrichmentStore) LoadSAPProductEnrichmentSnapshot(ctx context.Context, organizationID int32, sourceItemCode string) (enrichment.EnrichmentSourceSnapshot, error) {
	if s == nil || s.queries == nil {
		return enrichment.EnrichmentSourceSnapshot{}, fmt.Errorf("product enrichment repository is not configured")
	}
	product, err := s.queries.GetProductBySKU(ctx, GetProductBySKUParams{
		OrganizationID: organizationID,
		Sku:            sourceItemCode,
	})
	if err != nil {
		return enrichment.EnrichmentSourceSnapshot{}, fmt.Errorf("get committed product: %w", err)
	}

	snapshot := enrichment.EnrichmentSourceSnapshot{
		OrganizationID: product.OrganizationID,
		ProductID:      product.ID,
		SourceSystem:   enrichment.SourceSystemSAP,
		SourceItemCode: product.Sku,
		SourceItemName: product.Name,
		ProductType:    enrichment.ProductType(pgText(product.ProductType)),
		UOM:            enrichment.UOMContext{},
	}
	if product.Description.Valid {
		snapshot.Description = product.Description.String
	}

	if product.BrandID.Valid && product.BrandID.Int32 > 0 {
		brand, err := s.queries.GetBrandByID(ctx, product.BrandID.Int32)
		if err != nil {
			return enrichment.EnrichmentSourceSnapshot{}, fmt.Errorf("get structured product brand: %w", err)
		}
		snapshot.Brand = &enrichment.BrandIdentity{ID: brand.ID, Code: brand.Code, Name: brand.Name}
	}

	if product.CategoryID.Valid && product.CategoryID.Int32 > 0 {
		category, err := s.loadCategoryIdentity(ctx, product.CategoryID.Int32)
		if err != nil {
			return enrichment.EnrichmentSourceSnapshot{}, fmt.Errorf("get structured product category: %w", err)
		}
		snapshot.Category = category
	}

	if product.BaseUomID.Valid && product.BaseUomID.Int32 > 0 {
		uom, err := s.queries.GetUnitOfMeasure(ctx, product.BaseUomID.Int32)
		if err != nil {
			return enrichment.EnrichmentSourceSnapshot{}, fmt.Errorf("get base product UoM: %w", err)
		}
		snapshot.UOM.Base = &enrichment.UOMIdentity{ID: uom.ID, Code: uom.Code, Name: uom.Name}
	}

	conversions, err := s.queries.ListProductUOMConversions(ctx, product.ID)
	if err != nil {
		return enrichment.EnrichmentSourceSnapshot{}, fmt.Errorf("list product UoM conversions: %w", err)
	}
	for _, conversion := range conversions {
		from, err := s.queries.GetUnitOfMeasure(ctx, conversion.FromUomID)
		if err != nil {
			return enrichment.EnrichmentSourceSnapshot{}, fmt.Errorf("get conversion source UoM: %w", err)
		}
		to, err := s.queries.GetUnitOfMeasure(ctx, conversion.ToUomID)
		if err != nil {
			return enrichment.EnrichmentSourceSnapshot{}, fmt.Errorf("get conversion target UoM: %w", err)
		}
		factor, err := conversionFactorString(conversion.ConversionFactor)
		if err != nil {
			return enrichment.EnrichmentSourceSnapshot{}, fmt.Errorf("format product UoM conversion factor: %w", err)
		}
		snapshot.UOM.Conversions = append(snapshot.UOM.Conversions, enrichment.UOMConversionContext{
			From:             enrichment.UOMIdentity{ID: from.ID, Code: from.Code, Name: from.Name},
			To:               enrichment.UOMIdentity{ID: to.ID, Code: to.Code, Name: to.Name},
			ConversionFactor: factor,
			IsDefault:        conversion.IsDefault.Valid && conversion.IsDefault.Bool,
		})
	}

	return snapshot, nil
}

func (s *ProductEnrichmentStore) CreateOrGetPendingSuggestion(ctx context.Context, input enrichment.PendingSuggestionInput) (enrichment.PendingSuggestion, error) {
	if s == nil || s.queries == nil {
		return enrichment.PendingSuggestion{}, fmt.Errorf("product enrichment repository is not configured")
	}
	suggestion, err := s.queries.CreateOrGetProductEnrichmentSuggestion(ctx, CreateOrGetProductEnrichmentSuggestionParams{
		OrganizationID:        input.OrganizationID,
		ProductID:             input.ProductID,
		SourceItemCode:        input.SourceItemCode,
		SourceItemName:        input.SourceItemName,
		SourceDataFingerprint: input.SourceDataFingerprint,
		ContractVersion:       input.ContractVersion,
		StructuredCurrent:     input.StructuredCurrent,
		ProposedBrand:         nil,
		ProposedCategory:      nil,
		ProposedDescription:   nil,
		UnsupportedSemantics:  nil,
		Provider:              pgtype.Text{},
		Model:                 pgtype.Text{},
		ModelVersion:          pgtype.Text{},
	})
	if err != nil {
		return enrichment.PendingSuggestion{}, err
	}
	return enrichment.PendingSuggestion{ID: suggestion.ID, Status: enrichment.SuggestionStatus(suggestion.Status)}, nil
}

func (s *ProductEnrichmentStore) ListBrandCandidates(ctx context.Context, limit int) ([]enrichment.BrandCandidate, error) {
	if s == nil || s.queries == nil {
		return nil, fmt.Errorf("product enrichment repository is not configured")
	}
	brands, err := s.queries.ListActiveBrands(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = enrichment.DefaultCandidateLimit
	}
	if limit > len(brands) {
		limit = len(brands)
	}
	candidates := make([]enrichment.BrandCandidate, 0, limit)
	for _, brand := range brands[:limit] {
		candidates = append(candidates, enrichment.BrandCandidate{ID: brand.ID, Code: brand.Code, Name: brand.Name})
	}
	return candidates, nil
}

func (s *ProductEnrichmentStore) ListCategoryCandidates(ctx context.Context, limit int) ([]enrichment.CategoryCandidate, error) {
	if s == nil || s.queries == nil {
		return nil, fmt.Errorf("product enrichment repository is not configured")
	}
	categories, err := s.queries.GetCategoryHierarchy(ctx, true)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = enrichment.DefaultCandidateLimit
	}
	if limit > len(categories) {
		limit = len(categories)
	}
	candidates := make([]enrichment.CategoryCandidate, 0, limit)
	for _, category := range categories[:limit] {
		path := []string(nil)
		if strings.TrimSpace(category.FullPath) != "" {
			path = strings.Split(category.FullPath, " > ")
		}
		candidates = append(candidates, enrichment.CategoryCandidate{
			ID:   category.ID,
			Code: category.Code,
			Name: category.Name,
			Path: path,
		})
	}
	return candidates, nil
}

func (s *ProductEnrichmentStore) loadCategoryIdentity(ctx context.Context, categoryID int32) (*enrichment.CategoryIdentity, error) {
	currentID := categoryID
	visited := map[int32]struct{}{}
	var names []string
	var identity *enrichment.CategoryIdentity
	for currentID > 0 {
		if _, exists := visited[currentID]; exists {
			return nil, fmt.Errorf("category hierarchy cycle at id %d", currentID)
		}
		visited[currentID] = struct{}{}
		category, err := s.queries.GetProductCategory(ctx, currentID)
		if err != nil {
			return nil, err
		}
		if identity == nil {
			identity = &enrichment.CategoryIdentity{ID: category.ID, Code: category.Code, Name: category.Name}
		}
		names = append([]string{category.Name}, names...)
		if !category.ParentCategoryID.Valid {
			break
		}
		currentID = category.ParentCategoryID.Int32
	}
	identity.Path = names
	return identity, nil
}

func conversionFactorString(value pgtype.Numeric) (string, error) {
	encoded, err := value.Value()
	if err != nil {
		return "", err
	}
	if encoded == nil {
		return "", nil
	}
	return fmt.Sprint(encoded), nil
}

func pgText(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

var _ enrichment.EnrichmentStore = (*ProductEnrichmentStore)(nil)
var _ enrichment.CandidateDictionary = (*ProductEnrichmentStore)(nil)
