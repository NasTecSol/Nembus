package enrichment

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ResponseErrorClass identifies permanent provider-response validation
// failures. Network, rate-limit, timeout, and retry policy errors belong to a
// future provider/worker stage.
type ResponseErrorClass string

const (
	ResponseMalformed           ResponseErrorClass = "malformed_response"
	ResponseContractViolation   ResponseErrorClass = "contract_violation"
	ResponseCandidateMismatch   ResponseErrorClass = "candidate_mismatch"
	ResponseProhibitedOutput    ResponseErrorClass = "prohibited_output"
	ResponseCorrelationMismatch ResponseErrorClass = "correlation_mismatch"
)

// ResponseError is intentionally small so callers can classify a permanent
// parser/validation failure without depending on provider-specific errors.
type ResponseError struct {
	Class ResponseErrorClass
	Err   error
}

func (e *ResponseError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return string(e.Class)
	}
	return fmt.Sprintf("%s: %v", e.Class, e.Err)
}

func (e *ResponseError) Unwrap() error { return e.Err }

// ResponseErrorClassOf returns the parser classification, if err is one of
// the provider-neutral response errors.
func ResponseErrorClassOf(err error) ResponseErrorClass {
	var responseErr *ResponseError
	if errors.As(err, &responseErr) {
		return responseErr.Class
	}
	return ""
}

type responseWire struct {
	SourceItemCode       json.RawMessage `json:"source_item_code"`
	Brand                json.RawMessage `json:"brand"`
	Category             json.RawMessage `json:"category"`
	Description          json.RawMessage `json:"description"`
	UnsupportedSemantics json.RawMessage `json:"unsupported_semantics"`
}

// ParseEnrichmentResponse parses exactly the provider-neutral model object.
// Provider adapters must extract structured model content before calling this
// function; wrapper formats and provider metadata are not accepted here.
func ParseEnrichmentResponse(raw []byte, request EnrichmentRequest) (EnrichmentResult, error) {
	if err := request.Validate(); err != nil {
		return EnrichmentResult{}, responseError(ResponseContractViolation, "invalid enrichment request: %w", err)
	}
	if !json.Valid(raw) {
		return EnrichmentResult{}, responseError(ResponseMalformed, "response is not valid JSON")
	}

	var wire responseWire
	if err := decodeStrictObject(raw, &wire, "source_item_code", "brand", "category", "description", "unsupported_semantics"); err != nil {
		return EnrichmentResult{}, responseError(ResponseContractViolation, "response object: %w", err)
	}

	correlation, err := decodeRequiredString(wire.SourceItemCode, "source_item_code")
	if err != nil {
		return EnrichmentResult{}, responseError(ResponseContractViolation, "%v", err)
	}
	correlation = strings.TrimSpace(correlation)
	if correlation == "" {
		return EnrichmentResult{}, responseError(ResponseContractViolation, "source_item_code is required")
	}
	if correlation != strings.TrimSpace(request.SourceItemCode) {
		return EnrichmentResult{}, responseError(ResponseCorrelationMismatch, "source_item_code %q does not match request %q", correlation, request.SourceItemCode)
	}

	brand, err := decodeProposal[BrandProposal](wire.Brand, "brand", "action", "confidence")
	if err != nil {
		return EnrichmentResult{}, responseError(ResponseContractViolation, "%v", err)
	}
	category, err := decodeProposal[CategoryProposal](wire.Category, "category", "action", "confidence")
	if err != nil {
		return EnrichmentResult{}, responseError(ResponseContractViolation, "%v", err)
	}
	description, err := decodeProposal[DescriptionProposal](wire.Description, "description", "action", "confidence")
	if err != nil {
		return EnrichmentResult{}, responseError(ResponseContractViolation, "%v", err)
	}
	unsupported, err := decodeUnsupportedSemantics(wire.UnsupportedSemantics)
	if err != nil {
		return EnrichmentResult{}, responseError(ResponseContractViolation, "%v", err)
	}

	proposals := ProposalSet{
		Brand:                &brand,
		Category:             &category,
		Description:          &description,
		UnsupportedSemantics: unsupported,
	}
	normalizeResponseText(&proposals)
	if err := validateResponseAgainstRequest(&proposals, request); err != nil {
		return EnrichmentResult{}, err
	}
	return EnrichmentResult{SourceItemCode: correlation, Proposals: proposals}, nil
}

// ParseEnrichmentResponseString is the string convenience form of the strict
// byte parser. It performs no provider-specific extraction or repair.
func ParseEnrichmentResponseString(raw string, request EnrichmentRequest) (EnrichmentResult, error) {
	return ParseEnrichmentResponse([]byte(raw), request)
}

func responseError(class ResponseErrorClass, format string, args ...any) error {
	return &ResponseError{Class: class, Err: fmt.Errorf(format, args...)}
}

func decodeStrictObject(raw []byte, destination any, required ...string) error {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' || !json.Valid(trimmed) {
		return fmt.Errorf("expected one JSON object")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil || fields == nil {
		return fmt.Errorf("expected one JSON object")
	}
	for _, field := range required {
		value, present := fields[field]
		if !present || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("%s is required", field)
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value is not allowed")
		}
		return err
	}
	return nil
}

func decodeRequiredString(raw json.RawMessage, field string) (string, error) {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return "", fmt.Errorf("%s is required", field)
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("%s must be a string: %w", field, err)
	}
	return value, nil
}

func decodeProposal[T any](raw json.RawMessage, field string, required ...string) (T, error) {
	var proposal T
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return proposal, fmt.Errorf("%s is required", field)
	}
	if err := decodeStrictObject(raw, &proposal, required...); err != nil {
		return proposal, fmt.Errorf("%s: %w", field, err)
	}
	return proposal, nil
}

func decodeUnsupportedSemantics(raw json.RawMessage) ([]UnsupportedSemantic, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '[' || bytes.Equal(trimmed, []byte("null")) {
		return nil, fmt.Errorf("unsupported_semantics must be a JSON array")
	}
	var elements []json.RawMessage
	if err := json.Unmarshal(trimmed, &elements); err != nil {
		return nil, fmt.Errorf("unsupported_semantics must be a JSON array: %w", err)
	}
	result := make([]UnsupportedSemantic, 0, len(elements))
	for i, element := range elements {
		var semantic UnsupportedSemantic
		if err := decodeStrictObject(element, &semantic, "semantic_type", "key", "value", "confidence"); err != nil {
			return nil, fmt.Errorf("unsupported_semantics[%d]: %w", i, err)
		}
		result = append(result, semantic)
	}
	return result, nil
}

func normalizeResponseText(proposals *ProposalSet) {
	if proposals == nil {
		return
	}
	if proposals.Brand != nil {
		proposals.Brand.TargetCode = strings.TrimSpace(proposals.Brand.TargetCode)
		proposals.Brand.CanonicalName = strings.TrimSpace(proposals.Brand.CanonicalName)
		normalizeEvidence(proposals.Brand.Evidence)
		proposals.Brand.Explanation = strings.TrimSpace(proposals.Brand.Explanation)
	}
	if proposals.Category != nil {
		proposals.Category.TargetCode = strings.TrimSpace(proposals.Category.TargetCode)
		proposals.Category.CanonicalName = strings.TrimSpace(proposals.Category.CanonicalName)
		normalizeEvidence(proposals.Category.Evidence)
		proposals.Category.Explanation = strings.TrimSpace(proposals.Category.Explanation)
	}
	if proposals.Description != nil {
		proposals.Description.Value = strings.TrimSpace(proposals.Description.Value)
		normalizeEvidence(proposals.Description.Evidence)
		proposals.Description.Explanation = strings.TrimSpace(proposals.Description.Explanation)
	}
	for i := range proposals.UnsupportedSemantics {
		proposals.UnsupportedSemantics[i].SemanticType = strings.TrimSpace(proposals.UnsupportedSemantics[i].SemanticType)
		proposals.UnsupportedSemantics[i].Key = strings.TrimSpace(proposals.UnsupportedSemantics[i].Key)
		normalizeEvidence(proposals.UnsupportedSemantics[i].Evidence)
		proposals.UnsupportedSemantics[i].Explanation = strings.TrimSpace(proposals.UnsupportedSemantics[i].Explanation)
	}
}

func normalizeEvidence(evidence []string) {
	for i := range evidence {
		evidence[i] = strings.TrimSpace(evidence[i])
	}
}

func validateResponseAgainstRequest(proposals *ProposalSet, request EnrichmentRequest) error {
	if proposals == nil || proposals.Brand == nil || proposals.Category == nil || proposals.Description == nil {
		return responseError(ResponseContractViolation, "brand, category, and description proposals are required")
	}

	structuredCurrent, err := StructuredCurrent(request.Snapshot)
	if err != nil {
		return responseError(ResponseContractViolation, "marshal request structured context: %v", err)
	}
	for i := range proposals.UnsupportedSemantics {
		if key, found := findProhibitedJSONKey(proposals.UnsupportedSemantics[i].Value); found {
			return responseError(ResponseProhibitedOutput, "unsupported semantic contains prohibited authoritative field %q", key)
		}
		if prohibitedEnrichmentTarget(proposals.UnsupportedSemantics[i].SemanticType) || prohibitedEnrichmentTarget(proposals.UnsupportedSemantics[i].Key) {
			return responseError(ResponseProhibitedOutput, "unsupported semantic targets an SAP-authoritative field")
		}
	}
	if err := proposals.Validate(structuredCurrent); err != nil {
		return responseError(ResponseContractViolation, "%v", err)
	}

	index, err := NewCandidateIndex(request.BrandCandidates, request.CategoryCandidates)
	if err != nil {
		return responseError(ResponseCandidateMismatch, "%v", err)
	}
	if err := validateBrandResponse(proposals.Brand, request.Snapshot.Brand, index); err != nil {
		return err
	}
	if err := validateCategoryResponse(proposals.Category, request.Snapshot.Category, index); err != nil {
		return err
	}
	if err := validateDescriptionResponse(proposals.Description, request.Snapshot.Description); err != nil {
		return err
	}
	return nil
}

func validateBrandResponse(proposal *BrandProposal, current *BrandIdentity, index *CandidateIndex) error {
	if current != nil && current.Resolved() {
		if proposal.Action != ActionKeepExisting {
			return responseError(ResponseContractViolation, "resolved structured brand requires KEEP_EXISTING")
		}
		return rejectIdentityText(proposal.CanonicalName, "KEEP_EXISTING brand")
	}

	switch proposal.Action {
	case ActionMatchExisting:
		candidate, err := index.MatchBrand(proposal.TargetID, proposal.TargetCode)
		if err != nil {
			return responseError(ResponseCandidateMismatch, "%v", err)
		}
		// Model canonical names are not identity. The server candidate is the
		// only accepted identity in the validated result.
		proposal.TargetID = int32Pointer(candidate.ID)
		proposal.TargetCode = candidate.Code
		proposal.CanonicalName = candidate.Name
	case ActionProposeNew:
		if strings.TrimSpace(proposal.CanonicalName) == "" {
			return responseError(ResponseContractViolation, "PROPOSE_NEW brand requires a proposed canonical_name")
		}
	case ActionNoMatch:
		if err := rejectIdentityText(proposal.CanonicalName, "NO_MATCH brand"); err != nil {
			return err
		}
	default:
		return responseError(ResponseContractViolation, "unresolved structured brand requires MATCH_EXISTING, PROPOSE_NEW, or NO_MATCH")
	}
	return nil
}

func validateCategoryResponse(proposal *CategoryProposal, current *CategoryIdentity, index *CandidateIndex) error {
	if current != nil && current.Resolved() {
		if proposal.Action != ActionKeepExisting {
			return responseError(ResponseContractViolation, "populated structured category requires KEEP_EXISTING")
		}
		return rejectIdentityText(proposal.CanonicalName, "KEEP_EXISTING category")
	}

	switch proposal.Action {
	case ActionMatchExisting:
		candidate, err := index.MatchCategory(proposal.TargetID, proposal.TargetCode)
		if err != nil {
			return responseError(ResponseCandidateMismatch, "%v", err)
		}
		proposal.TargetID = int32Pointer(candidate.ID)
		proposal.TargetCode = candidate.Code
		proposal.CanonicalName = candidate.Name
	case ActionProposeNew:
		if strings.TrimSpace(proposal.CanonicalName) == "" {
			return responseError(ResponseContractViolation, "PROPOSE_NEW category requires a proposed canonical_name")
		}
	case ActionNoMatch:
		if err := rejectIdentityText(proposal.CanonicalName, "NO_MATCH category"); err != nil {
			return err
		}
	default:
		return responseError(ResponseContractViolation, "unresolved category requires MATCH_EXISTING, PROPOSE_NEW, or NO_MATCH")
	}
	return nil
}

func validateDescriptionResponse(proposal *DescriptionProposal, current string) error {
	if strings.TrimSpace(current) != "" {
		if proposal.Action != ActionKeepExisting {
			return responseError(ResponseContractViolation, "populated description requires KEEP_EXISTING")
		}
		if strings.TrimSpace(proposal.Value) != "" {
			return responseError(ResponseContractViolation, "KEEP_EXISTING description cannot contain a replacement value")
		}
		return nil
	}

	switch proposal.Action {
	case ActionProposeNew:
		value, err := NormalizeProposedDescription(proposal.Value)
		if err != nil {
			return responseError(ResponseContractViolation, "%v", err)
		}
		proposal.Value = value
	case ActionNoMatch:
		if strings.TrimSpace(proposal.Value) != "" {
			return responseError(ResponseContractViolation, "NO_MATCH description cannot contain a value")
		}
	default:
		return responseError(ResponseContractViolation, "missing description requires PROPOSE_NEW or NO_MATCH")
	}
	return nil
}

func rejectIdentityText(value, context string) error {
	if strings.TrimSpace(value) != "" {
		return responseError(ResponseContractViolation, "%s cannot contain canonical identity text", context)
	}
	return nil
}

func int32Pointer(value int32) *int32 { return &value }

// CandidateIndex is a deterministic server-owned lookup. It never performs
// fuzzy name matching and never queries a database using model-authored text.
type CandidateIndex struct {
	brandsByID       map[int32]BrandCandidate
	brandsByCode     map[string]BrandCandidate
	categoriesByID   map[int32]CategoryCandidate
	categoriesByCode map[string]CategoryCandidate
}

func NewCandidateIndex(brands []BrandCandidate, categories []CategoryCandidate) (*CandidateIndex, error) {
	index := &CandidateIndex{
		brandsByID:       make(map[int32]BrandCandidate, len(brands)),
		brandsByCode:     make(map[string]BrandCandidate, len(brands)),
		categoriesByID:   make(map[int32]CategoryCandidate, len(categories)),
		categoriesByCode: make(map[string]CategoryCandidate, len(categories)),
	}
	for _, candidate := range brands {
		candidate.Code = strings.TrimSpace(candidate.Code)
		candidate.Name = strings.TrimSpace(candidate.Name)
		if candidate.ID <= 0 || candidate.Code == "" || candidate.Name == "" {
			return nil, fmt.Errorf("invalid brand candidate identity")
		}
		if _, exists := index.brandsByID[candidate.ID]; exists {
			return nil, fmt.Errorf("ambiguous duplicate brand candidate id %d", candidate.ID)
		}
		if _, exists := index.brandsByCode[candidate.Code]; exists {
			return nil, fmt.Errorf("ambiguous duplicate brand candidate code %q", candidate.Code)
		}
		index.brandsByID[candidate.ID] = candidate
		index.brandsByCode[candidate.Code] = candidate
	}
	for _, candidate := range categories {
		candidate.Code = strings.TrimSpace(candidate.Code)
		candidate.Name = strings.TrimSpace(candidate.Name)
		if candidate.ID <= 0 || candidate.Code == "" || candidate.Name == "" {
			return nil, fmt.Errorf("invalid category candidate identity")
		}
		if _, exists := index.categoriesByID[candidate.ID]; exists {
			return nil, fmt.Errorf("ambiguous duplicate category candidate id %d", candidate.ID)
		}
		if _, exists := index.categoriesByCode[candidate.Code]; exists {
			return nil, fmt.Errorf("ambiguous duplicate category candidate code %q", candidate.Code)
		}
		index.categoriesByID[candidate.ID] = candidate
		index.categoriesByCode[candidate.Code] = candidate
	}
	return index, nil
}

func (i *CandidateIndex) MatchBrand(id *int32, code string) (BrandCandidate, error) {
	if i == nil {
		return BrandCandidate{}, fmt.Errorf("candidate index is not configured")
	}
	code = strings.TrimSpace(code)
	if id == nil && code == "" {
		return BrandCandidate{}, fmt.Errorf("MATCH_EXISTING brand requires target_id or target_code")
	}
	var byID BrandCandidate
	if id != nil {
		candidate, ok := i.brandsByID[*id]
		if !ok {
			return BrandCandidate{}, fmt.Errorf("unknown brand target_id %d", *id)
		}
		byID = candidate
	}
	if code != "" {
		candidate, ok := i.brandsByCode[code]
		if !ok {
			return BrandCandidate{}, fmt.Errorf("unknown brand target_code %q", code)
		}
		if id != nil && candidate.ID != byID.ID {
			return BrandCandidate{}, fmt.Errorf("brand target_id and target_code refer to different candidates")
		}
		return candidate, nil
	}
	return byID, nil
}

func (i *CandidateIndex) MatchCategory(id *int32, code string) (CategoryCandidate, error) {
	if i == nil {
		return CategoryCandidate{}, fmt.Errorf("candidate index is not configured")
	}
	code = strings.TrimSpace(code)
	if id == nil && code == "" {
		return CategoryCandidate{}, fmt.Errorf("MATCH_EXISTING category requires target_id or target_code")
	}
	var byID CategoryCandidate
	if id != nil {
		candidate, ok := i.categoriesByID[*id]
		if !ok {
			return CategoryCandidate{}, fmt.Errorf("unknown category target_id %d", *id)
		}
		byID = candidate
	}
	if code != "" {
		candidate, ok := i.categoriesByCode[code]
		if !ok {
			return CategoryCandidate{}, fmt.Errorf("unknown category target_code %q", code)
		}
		if id != nil && candidate.ID != byID.ID {
			return CategoryCandidate{}, fmt.Errorf("category target_id and target_code refer to different candidates")
		}
		return candidate, nil
	}
	return byID, nil
}

// GapsForSnapshot derives the exact deterministic Stage 2A inference gaps.
func GapsForSnapshot(snapshot EnrichmentSourceSnapshot) []EnrichmentGap {
	var gaps []EnrichmentGap
	if snapshot.Brand == nil || !snapshot.Brand.Resolved() {
		gaps = append(gaps, GapMissingBrand)
	}
	if strings.TrimSpace(snapshot.Description) == "" {
		gaps = append(gaps, GapMissingDescription)
	}
	if snapshot.Category == nil || !snapshot.Category.Resolved() {
		gaps = append(gaps, GapMissingCategory)
	}
	return gaps
}

// NewEnrichmentRequest builds a provider-safe request from the already
// allowlisted Stage 2A snapshot and bounded server candidate dictionaries.
func NewEnrichmentRequest(snapshot EnrichmentSourceSnapshot, brands []BrandCandidate, categories []CategoryCandidate) (EnrichmentRequest, error) {
	request := EnrichmentRequest{
		OrganizationID:     snapshot.OrganizationID,
		ProductID:          snapshot.ProductID,
		ContractVersion:    EnrichmentContractVersion,
		RequestVersion:     EnrichmentRequestVersion,
		SourceItemCode:     strings.TrimSpace(snapshot.SourceItemCode),
		Snapshot:           snapshot,
		Gaps:               GapsForSnapshot(snapshot),
		BrandCandidates:    brands,
		CategoryCandidates: categories,
		Policy: EnrichmentRequestPolicy{
			MaxDescriptionRunes:              MaxDescriptionRunes,
			SourceLanguagePreserving:         true,
			StructuredDataAuthoritative:      true,
			PopulatedCategoryKeepExisting:    true,
			ResolvedBrandKeepExisting:        true,
			ProductTypeImmutable:             true,
			OperationalFieldsProhibited:      true,
			UnsupportedSemanticsEvidenceOnly: true,
		},
	}
	if err := request.Validate(); err != nil {
		return EnrichmentRequest{}, err
	}
	return request, nil
}

func (r EnrichmentRequest) Validate() error {
	if r.OrganizationID <= 0 || r.ProductID <= 0 {
		return fmt.Errorf("organization_id and product_id correlation values are required")
	}
	if r.ContractVersion != EnrichmentContractVersion || r.RequestVersion != EnrichmentRequestVersion {
		return fmt.Errorf("unsupported enrichment contract or request version")
	}
	if strings.TrimSpace(r.SourceItemCode) == "" || strings.TrimSpace(r.Snapshot.SourceItemCode) == "" || strings.TrimSpace(r.SourceItemCode) != strings.TrimSpace(r.Snapshot.SourceItemCode) {
		return fmt.Errorf("source_item_code correlation is required and must match the snapshot")
	}
	if r.Snapshot.SourceSystem != SourceSystemSAP || !r.Snapshot.ProductType.Valid() {
		return fmt.Errorf("request snapshot must contain SAP and an approved immutable product_type")
	}
	if len(r.BrandCandidates) > DefaultCandidateLimit || len(r.CategoryCandidates) > DefaultCandidateLimit {
		return fmt.Errorf("candidate dictionaries exceed the bounded request limit")
	}
	wantGaps := GapsForSnapshot(r.Snapshot)
	if len(r.Gaps) != len(wantGaps) {
		return fmt.Errorf("request gaps do not match deterministic snapshot eligibility")
	}
	for i := range wantGaps {
		if r.Gaps[i] != wantGaps[i] {
			return fmt.Errorf("request gap %q does not match deterministic snapshot eligibility", r.Gaps[i])
		}
	}
	if r.Policy.MaxDescriptionRunes != MaxDescriptionRunes || !r.Policy.SourceLanguagePreserving || !r.Policy.StructuredDataAuthoritative || !r.Policy.PopulatedCategoryKeepExisting || !r.Policy.ResolvedBrandKeepExisting || !r.Policy.ProductTypeImmutable || !r.Policy.OperationalFieldsProhibited || !r.Policy.UnsupportedSemanticsEvidenceOnly {
		return fmt.Errorf("request policy does not match the strict enrichment contract")
	}
	if _, err := NewCandidateIndex(r.BrandCandidates, r.CategoryCandidates); err != nil {
		return err
	}
	return nil
}

// findProhibitedJSONKey rejects authoritative concepts anywhere inside an
// unsupported semantic value. The value itself remains otherwise free-form
// informational JSON.
func findProhibitedJSONKey(raw json.RawMessage) (string, bool) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", false
	}
	var walk func(any) (string, bool)
	walk = func(value any) (string, bool) {
		switch typed := value.(type) {
		case map[string]any:
			for key, child := range typed {
				if prohibitedEnrichmentTarget(key) {
					return key, true
				}
				if key, found := walk(child); found {
					return key, true
				}
			}
		case []any:
			for _, child := range typed {
				if key, found := walk(child); found {
					return key, true
				}
			}
		}
		return "", false
	}
	return walk(value)
}
