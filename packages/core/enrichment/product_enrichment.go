// Package enrichment contains the provider-neutral Stage 1 product-enrichment
// contract. It describes reviewable proposals only; it does not create or
// mutate products, brands, categories, or any other master data.
package enrichment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// ProposalAction is the only action vocabulary accepted by enrichment
// proposals. PROPOSE_NEW is a review proposal and never authorizes creation.
type ProposalAction string

const (
	ActionKeepExisting      ProposalAction = "KEEP_EXISTING"
	ActionMatchExisting     ProposalAction = "MATCH_EXISTING"
	ActionProposeNew        ProposalAction = "PROPOSE_NEW"
	ActionNoMatch           ProposalAction = "NO_MATCH"
	ActionUnsupportedTarget ProposalAction = "UNSUPPORTED_TARGET"
)

const (
	minConfidence = 0.0
	maxConfidence = 1.0
)

// SuggestionStatus is the durable Stage 1 inference/review lifecycle.
type SuggestionStatus string

const (
	SuggestionStatusPending    SuggestionStatus = "pending"
	SuggestionStatusProcessing SuggestionStatus = "processing"
	SuggestionStatusInReview   SuggestionStatus = "in_review"
	SuggestionStatusApproved   SuggestionStatus = "approved"
	SuggestionStatusRejected   SuggestionStatus = "rejected"
	SuggestionStatusRetryable  SuggestionStatus = "retryable"
	SuggestionStatusFailed     SuggestionStatus = "failed"
	SuggestionStatusApplied    SuggestionStatus = "applied"
)

// Valid reports whether the status is one of the architect-approved lifecycle
// values persisted by product_enrichment_suggestions.
func (s SuggestionStatus) Valid() bool {
	switch s {
	case SuggestionStatusPending,
		SuggestionStatusProcessing,
		SuggestionStatusInReview,
		SuggestionStatusApproved,
		SuggestionStatusRejected,
		SuggestionStatusRetryable,
		SuggestionStatusFailed,
		SuggestionStatusApplied:
		return true
	default:
		return false
	}
}

// CanTransition mirrors the SQL state predicates for the Stage 1 lifecycle.
// It is a provider-neutral domain guard; SQL remains the concurrency boundary.
func CanTransition(from, to SuggestionStatus) bool {
	switch from {
	case SuggestionStatusPending:
		return to == SuggestionStatusProcessing
	case SuggestionStatusProcessing:
		return to == SuggestionStatusInReview || to == SuggestionStatusRetryable || to == SuggestionStatusFailed
	case SuggestionStatusInReview:
		return to == SuggestionStatusApproved || to == SuggestionStatusRejected
	case SuggestionStatusApproved:
		return to == SuggestionStatusApplied
	case SuggestionStatusRetryable:
		return to == SuggestionStatusProcessing
	default:
		return false
	}
}

// BrandProposal is a reviewable brand suggestion. A MATCH_EXISTING proposal
// must carry a canonical existing brand ID or code. PROPOSE_NEW has no target
// identifier because it cannot imply automatic brand creation.
type BrandProposal struct {
	Action        ProposalAction `json:"action"`
	TargetID      *int32         `json:"target_id,omitempty"`
	TargetCode    string         `json:"target_code,omitempty"`
	CanonicalName string         `json:"canonical_name,omitempty"`
	Confidence    float64        `json:"confidence"`
	Evidence      []string       `json:"evidence,omitempty"`
	Explanation   string         `json:"explanation,omitempty"`
}

// CategoryProposal is a reviewable category suggestion. If structured_current
// already has a valid category, the only accepted action is KEEP_EXISTING;
// AI must not replace a populated structured category.
type CategoryProposal struct {
	Action        ProposalAction `json:"action"`
	TargetID      *int32         `json:"target_id,omitempty"`
	TargetCode    string         `json:"target_code,omitempty"`
	CanonicalName string         `json:"canonical_name,omitempty"`
	Confidence    float64        `json:"confidence"`
	Evidence      []string       `json:"evidence,omitempty"`
	Explanation   string         `json:"explanation,omitempty"`
}

// DescriptionProposal is a reviewable description suggestion. Descriptions
// have no existing target identity, so MATCH_EXISTING is not applicable.
type DescriptionProposal struct {
	Action      ProposalAction `json:"action"`
	Value       string         `json:"value,omitempty"`
	Confidence  float64        `json:"confidence"`
	Evidence    []string       `json:"evidence,omitempty"`
	Explanation string         `json:"explanation,omitempty"`
}

// UnsupportedSemantic preserves useful extracted information that has no
// approved Nembus destination yet, such as capacity, dimensions, or a family
// hint. Value is intentionally JSON rather than a product-column mapping.
type UnsupportedSemantic struct {
	SemanticType string          `json:"semantic_type"`
	Key          string          `json:"key"`
	Value        json.RawMessage `json:"value"`
	Confidence   float64         `json:"confidence"`
	Evidence     []string        `json:"evidence,omitempty"`
	Explanation  string          `json:"explanation,omitempty"`
}

// ProposalSet is the provider-neutral payload persisted in the proposed JSONB
// columns. It has no product_type field by design.
type ProposalSet struct {
	Brand                *BrandProposal        `json:"brand,omitempty"`
	Category             *CategoryProposal     `json:"category,omitempty"`
	Description          *DescriptionProposal  `json:"description,omitempty"`
	UnsupportedSemantics []UnsupportedSemantic `json:"unsupported_semantics,omitempty"`
}

// Validate checks the complete reviewable proposal contract against the
// current structured product data. structuredCurrent must be valid JSON.
func (p ProposalSet) Validate(structuredCurrent json.RawMessage) error {
	if len(bytes.TrimSpace(structuredCurrent)) == 0 || !json.Valid(structuredCurrent) {
		return fmt.Errorf("structured_current must be valid JSON")
	}

	categoryPresent, err := HasValidStructuredCategory(structuredCurrent)
	if err != nil {
		return err
	}

	brandPresent, err := HasValidStructuredBrand(structuredCurrent)
	if err != nil {
		return err
	}

	if p.Brand != nil {
		if err := p.Brand.validate(brandPresent); err != nil {
			return fmt.Errorf("brand proposal: %w", err)
		}
	}
	if p.Category != nil {
		if err := p.Category.validate(categoryPresent); err != nil {
			return fmt.Errorf("category proposal: %w", err)
		}
	}
	if p.Description != nil {
		if err := p.Description.validate(); err != nil {
			return fmt.Errorf("description proposal: %w", err)
		}
	}
	for i := range p.UnsupportedSemantics {
		if err := p.UnsupportedSemantics[i].validate(); err != nil {
			return fmt.Errorf("unsupported semantic %d: %w", i, err)
		}
	}
	if err := rejectProductTypeFields(p); err != nil {
		return err
	}
	return nil
}

func (p BrandProposal) validate(brandPresent bool) error {
	if err := validateAction(p.Action); err != nil {
		return err
	}
	if err := validateConfidence(p.Confidence); err != nil {
		return err
	}
	if brandPresent && p.Action != ActionKeepExisting {
		return fmt.Errorf("a populated structured brand may only record KEEP_EXISTING")
	}

	switch p.Action {
	case ActionMatchExisting:
		if !hasExistingTarget(p.TargetID, p.TargetCode) {
			return fmt.Errorf("MATCH_EXISTING requires a canonical target_id or target_code")
		}
	case ActionProposeNew:
		if p.TargetID != nil || strings.TrimSpace(p.TargetCode) != "" {
			return fmt.Errorf("PROPOSE_NEW cannot contain an existing target")
		}
		if strings.TrimSpace(p.CanonicalName) == "" {
			return fmt.Errorf("PROPOSE_NEW requires a canonical_name for review")
		}
	default:
		if p.TargetID != nil || strings.TrimSpace(p.TargetCode) != "" {
			return fmt.Errorf("%s cannot contain an existing target", p.Action)
		}
	}
	return nil
}

func (p CategoryProposal) validate(categoryPresent bool) error {
	if err := validateAction(p.Action); err != nil {
		return err
	}
	if err := validateConfidence(p.Confidence); err != nil {
		return err
	}
	if categoryPresent && p.Action != ActionKeepExisting {
		return fmt.Errorf("a populated structured category may only record KEEP_EXISTING")
	}

	switch p.Action {
	case ActionMatchExisting:
		if !hasExistingTarget(p.TargetID, p.TargetCode) {
			return fmt.Errorf("MATCH_EXISTING requires a canonical target_id or target_code")
		}
	case ActionProposeNew:
		if p.TargetID != nil || strings.TrimSpace(p.TargetCode) != "" {
			return fmt.Errorf("PROPOSE_NEW cannot contain an existing target")
		}
		if strings.TrimSpace(p.CanonicalName) == "" {
			return fmt.Errorf("PROPOSE_NEW requires a canonical_name for review")
		}
	default:
		if p.TargetID != nil || strings.TrimSpace(p.TargetCode) != "" {
			return fmt.Errorf("%s cannot contain an existing target", p.Action)
		}
	}
	return nil
}

func (p DescriptionProposal) validate() error {
	if err := validateAction(p.Action); err != nil {
		return err
	}
	if err := validateConfidence(p.Confidence); err != nil {
		return err
	}

	switch p.Action {
	case ActionProposeNew:
		if _, err := NormalizeProposedDescription(p.Value); err != nil {
			return fmt.Errorf("PROPOSE_NEW requires a valid catalog description for review: %w", err)
		}
	case ActionMatchExisting:
		return fmt.Errorf("MATCH_EXISTING is not applicable to descriptions")
	default:
		if strings.TrimSpace(p.Value) != "" {
			return fmt.Errorf("%s cannot contain a description value", p.Action)
		}
	}
	return nil
}

func (s UnsupportedSemantic) validate() error {
	if strings.TrimSpace(s.SemanticType) == "" || strings.TrimSpace(s.Key) == "" {
		return fmt.Errorf("semantic_type and key are required")
	}
	if err := validateConfidence(s.Confidence); err != nil {
		return err
	}
	if len(bytes.TrimSpace(s.Value)) == 0 || !json.Valid(s.Value) || bytes.Equal(bytes.TrimSpace(s.Value), []byte("null")) {
		return fmt.Errorf("value must be non-null valid JSON")
	}
	if prohibitedEnrichmentTarget(s.SemanticType) || prohibitedEnrichmentTarget(s.Key) {
		return fmt.Errorf("unsupported semantic cannot target an SAP-authoritative field: %q/%q", s.SemanticType, s.Key)
	}
	if key, found := findProhibitedJSONKey(s.Value); found {
		return fmt.Errorf("unsupported semantic value cannot contain SAP-authoritative field %q", key)
	}
	return nil
}

func validateAction(action ProposalAction) error {
	switch action {
	case ActionKeepExisting, ActionMatchExisting, ActionProposeNew, ActionNoMatch, ActionUnsupportedTarget:
		return nil
	default:
		return fmt.Errorf("invalid proposal action %q", action)
	}
}

func validateConfidence(confidence float64) error {
	if math.IsNaN(confidence) || math.IsInf(confidence, 0) || confidence < minConfidence || confidence > maxConfidence {
		return fmt.Errorf("confidence must be between %.1f and %.1f", minConfidence, maxConfidence)
	}
	return nil
}

func hasExistingTarget(targetID *int32, targetCode string) bool {
	return (targetID != nil && *targetID > 0) || strings.TrimSpace(targetCode) != ""
}

func rejectProductTypeFields(p ProposalSet) error {
	b, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal proposals: %w", err)
	}
	return rejectJSONKeys(b, map[string]struct{}{"product_type": {}})
}

func rejectJSONKeys(data []byte, forbidden map[string]struct{}) error {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("invalid proposal JSON: %w", err)
	}
	if key, found := findForbiddenKey(value, forbidden); found {
		return fmt.Errorf("proposal cannot contain field %q", key)
	}
	return nil
}

func findForbiddenKey(value any, forbidden map[string]struct{}) (string, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if isForbiddenKey(key, forbidden) {
				return key, true
			}
			if key, found := findForbiddenKey(child, forbidden); found {
				return key, true
			}
		}
	case []any:
		for _, child := range typed {
			if key, found := findForbiddenKey(child, forbidden); found {
				return key, true
			}
		}
	}
	return "", false
}

func isForbiddenKey(key string, forbidden map[string]struct{}) bool {
	normalizedKey := normalizeEnrichmentTarget(key)
	for candidate := range forbidden {
		if normalizedKey == normalizeEnrichmentTarget(candidate) {
			return true
		}
	}
	return false
}

// HasValidStructuredCategory recognizes the structured category forms used by
// the review contract. A non-empty category object/value, category_id,
// category_code, or category_name counts as already populated.
func HasValidStructuredCategory(structuredCurrent json.RawMessage) (bool, error) {
	var document map[string]json.RawMessage
	if err := json.Unmarshal(structuredCurrent, &document); err != nil {
		return false, fmt.Errorf("structured_current must be a JSON object: %w", err)
	}
	for _, key := range []string{"category", "structured_category"} {
		if raw, ok := document[key]; ok && hasMeaningfulStructuredValue(raw) {
			return true, nil
		}
	}
	for _, key := range []string{"category_id", "category_code", "category_name"} {
		if raw, ok := document[key]; ok && hasMeaningfulStructuredValue(raw) {
			return true, nil
		}
	}
	return false, nil
}

// HasValidStructuredBrand recognizes only resolved structured brand identity.
// A brand name without an existing ID or canonical code remains unresolved.
func HasValidStructuredBrand(structuredCurrent json.RawMessage) (bool, error) {
	var document map[string]json.RawMessage
	if err := json.Unmarshal(structuredCurrent, &document); err != nil {
		return false, fmt.Errorf("structured_current must be a JSON object: %w", err)
	}

	if raw, ok := document["brand_id"]; ok && hasStructuredBrandID(raw) {
		return true, nil
	}
	for _, key := range []string{"brand_code", "canonical_brand_code"} {
		if raw, ok := document[key]; ok && hasStructuredBrandCode(raw) {
			return true, nil
		}
	}
	for _, key := range []string{"brand", "structured_brand"} {
		if raw, ok := document[key]; ok && hasStructuredBrandIdentity(raw) {
			return true, nil
		}
	}
	return false, nil
}

func hasStructuredBrandIdentity(raw json.RawMessage) bool {
	var value map[string]json.RawMessage
	if err := json.Unmarshal(raw, &value); err != nil {
		return false
	}
	for _, key := range []string{"id", "brand_id"} {
		if candidate, ok := value[key]; ok && hasStructuredBrandID(candidate) {
			return true
		}
	}
	for _, key := range []string{"code", "brand_code", "canonical_brand_code"} {
		if candidate, ok := value[key]; ok && hasStructuredBrandCode(candidate) {
			return true
		}
	}
	return false
}

func hasStructuredBrandID(raw json.RawMessage) bool {
	var value float64
	if err := json.Unmarshal(raw, &value); err != nil {
		return false
	}
	return value > 0
}

func hasStructuredBrandCode(raw json.RawMessage) bool {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return false
	}
	return strings.TrimSpace(value) != ""
}

func hasMeaningfulStructuredValue(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) || bytes.Equal(trimmed, []byte(`""`)) {
		return false
	}
	var value any
	if json.Unmarshal(trimmed, &value) != nil {
		return false
	}
	switch typed := value.(type) {
	case float64:
		return typed > 0
	case string:
		return strings.TrimSpace(typed) != ""
	case map[string]any:
		for _, key := range []string{"id", "code", "name", "value"} {
			if candidate, ok := typed[key]; ok {
				if rawCandidate, err := json.Marshal(candidate); err == nil && hasMeaningfulStructuredValue(rawCandidate) {
					return true
				}
			}
		}
		return false
	default:
		return true
	}
}

var prohibitedTargets = map[string]struct{}{
	"sku": {}, "skucode": {}, "itemcode": {}, "sourceitemcode": {}, "sapitemcode": {}, "sapdocumentidentity": {}, "sapdocumentid": {}, "sapdocentry": {}, "docentry": {}, "documentid": {}, "documentidentity": {},
	"barcode": {}, "barcodes": {}, "barcodeownership": {},
	"inventory": {}, "trackinventory": {}, "inventoryquantity": {}, "stock": {}, "stockquantity": {}, "warehousequantity": {}, "warehouseinventory": {}, "inventorywarehouse": {},
	"warehouse": {}, "warehouseid": {}, "store": {}, "storeid": {}, "storeinventory": {}, "inventorystore": {},
	"price": {}, "prices": {}, "pricing": {}, "tax": {}, "taxrate": {}, "taxrates": {}, "taxcategory": {},
	"uom": {}, "uomid": {}, "unitofmeasure": {}, "unitofmeasureid": {}, "uomconversion": {}, "uomconversionfactor": {}, "uomconversionfactors": {}, "unitofmeasureconversion": {}, "unitofmeasureconversionfactor": {}, "conversionfactor": {}, "conversionfactors": {}, "conversion": {},
	"supplier": {}, "supplierid": {}, "suppliercode": {}, "supplieridentity": {}, "active": {}, "isactive": {},
	"sellable": {}, "issellable": {}, "purchasable": {}, "ispurchasable": {},
	"producttype": {},
}

func prohibitedEnrichmentTarget(value string) bool {
	_, found := prohibitedTargets[normalizeEnrichmentTarget(value)]
	return found
}

// normalizeEnrichmentTarget maps casing and common separators to one
// deterministic semantic key. It intentionally does not perform fuzzy
// matching or infer unrelated concepts.
func normalizeEnrichmentTarget(value string) string {
	var normalized strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			normalized.WriteRune(r)
		}
	}
	return normalized.String()
}
