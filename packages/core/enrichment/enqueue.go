package enrichment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	SourceSystemSAP           = "SAP"
	EnrichmentContractVersion = "sap-product-enrichment-v1"
	EnrichmentRequestVersion  = "sap-product-enrichment-request-v1"
	MaxDescriptionRunes       = 500
	DefaultCandidateLimit     = 100
)

// ProductType is immutable structured product context. It is never a
// proposal target.
type ProductType string

const (
	ProductTypeStandard     ProductType = "standard"
	ProductTypeRawMaterial  ProductType = "raw_material"
	ProductTypeFixedAsset   ProductType = "fixed_asset"
	ProductTypeFinishedGood ProductType = "finished_good"
)

func (t ProductType) Valid() bool {
	switch t {
	case ProductTypeStandard, ProductTypeRawMaterial, ProductTypeFixedAsset, ProductTypeFinishedGood:
		return true
	default:
		return false
	}
}

type EnrichmentGap string

const (
	GapMissingBrand       EnrichmentGap = "missing_brand"
	GapMissingDescription EnrichmentGap = "missing_description"
	GapMissingCategory    EnrichmentGap = "missing_category"
)

type EligibilityReason string

const (
	ReasonNonSAPSource        EligibilityReason = "non_sap_source"
	ReasonMissingOrganization EligibilityReason = "missing_organization"
	ReasonMissingProduct      EligibilityReason = "missing_product"
	ReasonMissingItemCode     EligibilityReason = "missing_item_code"
	ReasonMissingItemName     EligibilityReason = "missing_item_name"
	ReasonInvalidProductType  EligibilityReason = "invalid_product_type"
	ReasonNoEnrichmentGaps    EligibilityReason = "no_enrichment_gaps"
)

// BrandIdentity and CategoryIdentity contain only structured taxonomy
// identity useful for precedence checks and future provider context.
type BrandIdentity struct {
	ID   int32  `json:"id,omitempty"`
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

func (b *BrandIdentity) Resolved() bool {
	return b != nil && (b.ID > 0 || strings.TrimSpace(b.Code) != "")
}

type CategoryIdentity struct {
	ID   int32    `json:"id,omitempty"`
	Code string   `json:"code,omitempty"`
	Name string   `json:"name,omitempty"`
	Path []string `json:"path,omitempty"`
}

func (c *CategoryIdentity) Resolved() bool {
	return c != nil && (c.ID > 0 || strings.TrimSpace(c.Code) != "" || strings.TrimSpace(c.Name) != "")
}

type UOMIdentity struct {
	ID   int32  `json:"id,omitempty"`
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

type UOMConversionContext struct {
	From             UOMIdentity `json:"from"`
	To               UOMIdentity `json:"to"`
	ConversionFactor string      `json:"conversion_factor"`
	IsDefault        bool        `json:"is_default"`
}

type UOMContext struct {
	Base        *UOMIdentity           `json:"base,omitempty"`
	Purchase    *UOMIdentity           `json:"purchase,omitempty"`
	Sales       *UOMIdentity           `json:"sales,omitempty"`
	Conversions []UOMConversionContext `json:"conversions,omitempty"`
}

// EnrichmentSourceSnapshot is the allowlisted, deterministic source context.
// OrganizationID and ProductID are internal correlation values and are never
// serialized into provider input or structured_current.
type EnrichmentSourceSnapshot struct {
	OrganizationID int32             `json:"-"`
	ProductID      int32             `json:"-"`
	SourceSystem   string            `json:"source_system"`
	SourceItemCode string            `json:"source_item_code"`
	SourceItemName string            `json:"source_item_name"`
	Description    string            `json:"description,omitempty"`
	ProductType    ProductType       `json:"product_type"`
	Brand          *BrandIdentity    `json:"brand,omitempty"`
	Category       *CategoryIdentity `json:"category,omitempty"`
	UOM            UOMContext        `json:"uom"`
}

type EnrichmentEligibilityInput struct {
	SourceSystem   string
	OrganizationID int32
	ProductID      int32
	SourceItemCode string
	SourceItemName string
	ProductType    ProductType
	Brand          *BrandIdentity
	Category       *CategoryIdentity
	Description    string
}

type EligibilityDecision struct {
	Eligible bool
	Rejected bool
	Reasons  []EligibilityReason
	Gaps     []EnrichmentGap
}

// EvaluateEligibility is pure and only considers deterministic source state.
func EvaluateEligibility(input EnrichmentEligibilityInput) EligibilityDecision {
	decision := EligibilityDecision{}
	if input.SourceSystem != SourceSystemSAP {
		decision.Reasons = append(decision.Reasons, ReasonNonSAPSource)
	}
	if input.OrganizationID <= 0 {
		decision.Reasons = append(decision.Reasons, ReasonMissingOrganization)
	}
	if input.ProductID <= 0 {
		decision.Reasons = append(decision.Reasons, ReasonMissingProduct)
	}
	if strings.TrimSpace(input.SourceItemCode) == "" {
		decision.Reasons = append(decision.Reasons, ReasonMissingItemCode)
	}
	if strings.TrimSpace(input.SourceItemName) == "" {
		decision.Reasons = append(decision.Reasons, ReasonMissingItemName)
	}
	if !input.ProductType.Valid() {
		decision.Reasons = append(decision.Reasons, ReasonInvalidProductType)
		decision.Rejected = true
	}

	if !input.Brand.Resolved() {
		decision.Gaps = append(decision.Gaps, GapMissingBrand)
	}
	if strings.TrimSpace(input.Description) == "" {
		decision.Gaps = append(decision.Gaps, GapMissingDescription)
	}
	if !input.Category.Resolved() {
		decision.Gaps = append(decision.Gaps, GapMissingCategory)
	}

	if len(decision.Reasons) == 0 && len(decision.Gaps) == 0 {
		decision.Reasons = append(decision.Reasons, ReasonNoEnrichmentGaps)
	}
	decision.Eligible = len(decision.Reasons) == 0 && len(decision.Gaps) > 0
	return decision
}

// BrandCandidate and CategoryCandidate are bounded canonical dictionary
// records. They intentionally omit database metadata and operational fields.
type BrandCandidate struct {
	ID   int32  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type CategoryCandidate struct {
	ID   int32    `json:"id"`
	Code string   `json:"code"`
	Name string   `json:"name"`
	Path []string `json:"path,omitempty"`
}

type CandidateDictionary interface {
	ListBrandCandidates(context.Context, int) ([]BrandCandidate, error)
	ListCategoryCandidates(context.Context, int) ([]CategoryCandidate, error)
}

type EnrichmentRequest struct {
	// OrganizationID and ProductID are internal correlation values. They are
	// deliberately excluded from provider-facing JSON.
	OrganizationID     int32                    `json:"-"`
	ProductID          int32                    `json:"-"`
	ContractVersion    string                   `json:"contract_version"`
	RequestVersion     string                   `json:"request_version"`
	SourceItemCode     string                   `json:"source_item_code"`
	Snapshot           EnrichmentSourceSnapshot `json:"snapshot"`
	Gaps               []EnrichmentGap          `json:"gaps"`
	BrandCandidates    []BrandCandidate         `json:"brand_candidates,omitempty"`
	CategoryCandidates []CategoryCandidate      `json:"category_candidates,omitempty"`
	Policy             EnrichmentRequestPolicy  `json:"policy"`
}

// EnrichmentRequestPolicy is provider-neutral contract information. It makes
// the safety boundary explicit without exposing provider parameters.
type EnrichmentRequestPolicy struct {
	MaxDescriptionRunes              int  `json:"max_description_runes"`
	SourceLanguagePreserving         bool `json:"source_language_preserving"`
	StructuredDataAuthoritative      bool `json:"structured_data_authoritative"`
	PopulatedCategoryKeepExisting    bool `json:"populated_category_keep_existing"`
	ResolvedBrandKeepExisting        bool `json:"resolved_brand_keep_existing"`
	ProductTypeImmutable             bool `json:"product_type_immutable"`
	OperationalFieldsProhibited      bool `json:"operational_fields_prohibited"`
	UnsupportedSemanticsEvidenceOnly bool `json:"unsupported_semantics_evidence_only"`
}

type EnrichmentResult struct {
	SourceItemCode string
	Proposals      ProposalSet
	// Trusted adapter metadata. These values are attached by the server-side
	// provider implementation, never accepted from model JSON.
	Provider     string
	Model        string
	ModelVersion string
	ResponseID   string
}

// ProductEnrichmentProvider is a future adapter contract only. Stage 2A does
// not provide or invoke an implementation.
type ProductEnrichmentProvider interface {
	Enrich(context.Context, EnrichmentRequest) (EnrichmentResult, error)
}

type PendingSuggestionInput struct {
	OrganizationID        int32
	ProductID             int32
	SourceItemCode        string
	SourceItemName        string
	SourceDataFingerprint string
	ContractVersion       string
	StructuredCurrent     json.RawMessage
}

type PendingSuggestion struct {
	ID     int32
	Status SuggestionStatus
}

type EnrichmentStore interface {
	LoadSAPProductEnrichmentSnapshot(context.Context, int32, string) (EnrichmentSourceSnapshot, error)
	CreateOrGetPendingSuggestion(context.Context, PendingSuggestionInput) (PendingSuggestion, error)
}

type EnrichmentEnqueuer interface {
	EnqueueSAPProduct(context.Context, int32, string) (EnqueueResult, error)
}

type EnqueueResult struct {
	Decision    EligibilityDecision
	Snapshot    EnrichmentSourceSnapshot
	Fingerprint string
	Suggestion  PendingSuggestion
}

type ProductEnrichmentCoordinator struct {
	store           EnrichmentStore
	contractVersion string
}

func NewProductEnrichmentCoordinator(store EnrichmentStore) *ProductEnrichmentCoordinator {
	return &ProductEnrichmentCoordinator{
		store:           store,
		contractVersion: EnrichmentContractVersion,
	}
}

func (c *ProductEnrichmentCoordinator) EnqueueSAPProduct(ctx context.Context, organizationID int32, sourceItemCode string) (EnqueueResult, error) {
	if c == nil || c.store == nil {
		return EnqueueResult{}, fmt.Errorf("product enrichment coordinator is not configured")
	}
	snapshot, err := c.store.LoadSAPProductEnrichmentSnapshot(ctx, organizationID, sourceItemCode)
	if err != nil {
		return EnqueueResult{}, fmt.Errorf("load SAP product enrichment snapshot: %w", err)
	}
	decision := EvaluateEligibility(EnrichmentEligibilityInput{
		SourceSystem:   snapshot.SourceSystem,
		OrganizationID: snapshot.OrganizationID,
		ProductID:      snapshot.ProductID,
		SourceItemCode: snapshot.SourceItemCode,
		SourceItemName: snapshot.SourceItemName,
		ProductType:    snapshot.ProductType,
		Brand:          snapshot.Brand,
		Category:       snapshot.Category,
		Description:    snapshot.Description,
	})
	result := EnqueueResult{Decision: decision, Snapshot: snapshot}
	if !decision.Eligible {
		return result, nil
	}

	fingerprint, err := FingerprintSnapshot(snapshot)
	if err != nil {
		return result, fmt.Errorf("fingerprint SAP product enrichment source: %w", err)
	}
	structuredCurrent, err := StructuredCurrent(snapshot)
	if err != nil {
		return result, fmt.Errorf("marshal structured current enrichment context: %w", err)
	}
	suggestion, err := c.store.CreateOrGetPendingSuggestion(ctx, PendingSuggestionInput{
		OrganizationID:        snapshot.OrganizationID,
		ProductID:             snapshot.ProductID,
		SourceItemCode:        snapshot.SourceItemCode,
		SourceItemName:        snapshot.SourceItemName,
		SourceDataFingerprint: fingerprint,
		ContractVersion:       c.contractVersion,
		StructuredCurrent:     structuredCurrent,
	})
	if err != nil {
		return result, fmt.Errorf("create or get pending product enrichment suggestion: %w", err)
	}
	result.Fingerprint = fingerprint
	result.Suggestion = suggestion
	return result, nil
}

var _ EnrichmentEnqueuer = (*ProductEnrichmentCoordinator)(nil)

type structuredCurrent struct {
	SourceSystem   string            `json:"source_system"`
	SourceItemCode string            `json:"source_item_code"`
	SourceItemName string            `json:"source_item_name"`
	Description    string            `json:"description,omitempty"`
	ProductType    ProductType       `json:"product_type"`
	Brand          *BrandIdentity    `json:"brand,omitempty"`
	Category       *CategoryIdentity `json:"category,omitempty"`
	UOM            UOMContext        `json:"uom"`
}

func StructuredCurrent(snapshot EnrichmentSourceSnapshot) (json.RawMessage, error) {
	return json.Marshal(structuredCurrent{
		SourceSystem:   snapshot.SourceSystem,
		SourceItemCode: snapshot.SourceItemCode,
		SourceItemName: snapshot.SourceItemName,
		Description:    snapshot.Description,
		ProductType:    snapshot.ProductType,
		Brand:          snapshot.Brand,
		Category:       snapshot.Category,
		UOM:            normalizeUOM(snapshot.UOM),
	})
}

type fingerprintSource struct {
	ContractVersion string            `json:"contract_version"`
	SourceSystem    string            `json:"source_system"`
	SourceItemCode  string            `json:"source_item_code"`
	SourceItemName  string            `json:"source_item_name"`
	Description     string            `json:"description"`
	ProductType     ProductType       `json:"product_type"`
	Brand           *BrandIdentity    `json:"brand,omitempty"`
	Category        *CategoryIdentity `json:"category,omitempty"`
	UOM             UOMContext        `json:"uom"`
}

// FingerprintSnapshot hashes only enrichment-relevant structured source state.
// JSON structs and sorted conversion pairs make the representation stable;
// operational values, timestamps, candidates, and provider state are absent.
func FingerprintSnapshot(snapshot EnrichmentSourceSnapshot) (string, error) {
	canonical := fingerprintSource{
		ContractVersion: EnrichmentContractVersion,
		SourceSystem:    snapshot.SourceSystem,
		SourceItemCode:  snapshot.SourceItemCode,
		SourceItemName:  snapshot.SourceItemName,
		Description:     snapshot.Description,
		ProductType:     snapshot.ProductType,
		Brand:           snapshot.Brand,
		Category:        snapshot.Category,
		UOM:             normalizeUOM(snapshot.UOM),
	}
	data, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:]), nil
}

func normalizeUOM(uom UOMContext) UOMContext {
	copyUOM := uom
	copyUOM.Conversions = append([]UOMConversionContext(nil), uom.Conversions...)
	sort.Slice(copyUOM.Conversions, func(i, j int) bool {
		left, right := copyUOM.Conversions[i], copyUOM.Conversions[j]
		if left.From.Code != right.From.Code {
			return left.From.Code < right.From.Code
		}
		if left.To.Code != right.To.Code {
			return left.To.Code < right.To.Code
		}
		if left.ConversionFactor != right.ConversionFactor {
			return left.ConversionFactor < right.ConversionFactor
		}
		return !left.IsDefault && right.IsDefault
	})
	return copyUOM
}

func NormalizeProposedDescription(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", fmt.Errorf("description must not be empty")
	}
	if !utf8.ValidString(normalized) {
		return "", fmt.Errorf("description must be valid UTF-8")
	}
	if utf8.RuneCountInString(normalized) > MaxDescriptionRunes {
		return "", fmt.Errorf("description must be at most %d Unicode characters", MaxDescriptionRunes)
	}
	return normalized, nil
}
