package enrichment

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStrictParserRejectsMalformedTrailingAndUnknownJSON(t *testing.T) {
	request := unresolvedRequest(t)
	valid := responseJSON(t, request.SourceItemCode, validBrandMatch(), validCategoryMatch(), validDescription(), []UnsupportedSemantic{})
	tests := []struct {
		name  string
		raw   []byte
		class ResponseErrorClass
	}{
		{name: "malformed", raw: []byte(`{"source_item_code":`), class: ResponseMalformed},
		{name: "trailing object", raw: append(valid, []byte(` {}`)...), class: ResponseMalformed},
		{name: "unknown top-level field", raw: responseWithExtra(t, valid, "provider", "untrusted"), class: ResponseContractViolation},
		{name: "unknown nested field", raw: responseWithNestedExtra(t, valid, "brand", "unexpected", true), class: ResponseContractViolation},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseEnrichmentResponse(test.raw, request)
			requireResponseClass(t, err, test.class)
		})
	}
}

func TestStrictParserRejectsInvalidActionConfidenceAndCorrelation(t *testing.T) {
	request := unresolvedRequest(t)
	tests := []struct {
		name   string
		mutate func(map[string]any)
		class  ResponseErrorClass
	}{
		{name: "invalid action", mutate: func(body map[string]any) {
			body["brand"].(map[string]any)["action"] = "GUESS"
		}, class: ResponseContractViolation},
		{name: "confidence below zero", mutate: func(body map[string]any) {
			body["description"].(map[string]any)["confidence"] = -0.01
		}, class: ResponseContractViolation},
		{name: "confidence above one", mutate: func(body map[string]any) {
			body["description"].(map[string]any)["confidence"] = 1.01
		}, class: ResponseContractViolation},
		{name: "missing correlation", mutate: func(body map[string]any) {
			delete(body, "source_item_code")
		}, class: ResponseContractViolation},
		{name: "correlation mismatch", mutate: func(body map[string]any) {
			body["source_item_code"] = "OTHER"
		}, class: ResponseCorrelationMismatch},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var body map[string]any
			if err := json.Unmarshal(responseJSON(t, request.SourceItemCode, validBrandMatch(), validCategoryMatch(), validDescription(), []UnsupportedSemantic{}), &body); err != nil {
				t.Fatal(err)
			}
			test.mutate(body)
			raw, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			_, err = ParseEnrichmentResponse(raw, request)
			requireResponseClass(t, err, test.class)
		})
	}
}

func TestStrictBrandValidation(t *testing.T) {
	resolved := resolvedBrandRequest(t)
	for _, test := range []struct {
		name  string
		brand BrandProposal
		class ResponseErrorClass
	}{
		{name: "resolved keep passes", brand: BrandProposal{Action: ActionKeepExisting, Confidence: 0.8}},
		{name: "resolved match fails", brand: BrandProposal{Action: ActionMatchExisting, TargetID: int32Pointer(10), Confidence: 0.8}, class: ResponseContractViolation},
		{name: "resolved new fails", brand: BrandProposal{Action: ActionProposeNew, CanonicalName: "Other", Confidence: 0.8}, class: ResponseContractViolation},
		{name: "resolved no match fails", brand: BrandProposal{Action: ActionNoMatch, Confidence: 0.8}, class: ResponseContractViolation},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw := responseJSON(t, resolved.SourceItemCode, test.brand, validCategoryKeep(), validDescription(), []UnsupportedSemantic{})
			_, err := ParseEnrichmentResponse(raw, resolved)
			if test.class == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				requireResponseClass(t, err, test.class)
			}
		})
	}

	unresolved := unresolvedRequest(t)
	valid := responseJSON(t, unresolved.SourceItemCode, validBrandMatch(), validCategoryMatch(), validDescription(), []UnsupportedSemantic{})
	parsed, err := ParseEnrichmentResponse(valid, unresolved)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Proposals.Brand.TargetID == nil || *parsed.Proposals.Brand.TargetID != 10 || parsed.Proposals.Brand.TargetCode != "PANTENE" || parsed.Proposals.Brand.CanonicalName != "Pantene" {
		t.Fatalf("server canonical candidate was not accepted: %+v", parsed.Proposals.Brand)
	}

	for _, test := range []struct {
		name  string
		brand BrandProposal
		class ResponseErrorClass
	}{
		{name: "unknown id", brand: BrandProposal{Action: ActionMatchExisting, TargetID: int32Pointer(999), Confidence: 0.8}, class: ResponseCandidateMismatch},
		{name: "mismatched id code", brand: BrandProposal{Action: ActionMatchExisting, TargetID: int32Pointer(10), TargetCode: "HIK", Confidence: 0.8}, class: ResponseCandidateMismatch},
		{name: "new with existing id", brand: BrandProposal{Action: ActionProposeNew, TargetID: int32Pointer(10), CanonicalName: "Pantene", Confidence: 0.8}, class: ResponseContractViolation},
		{name: "no match with target", brand: BrandProposal{Action: ActionNoMatch, TargetCode: "PANTENE", Confidence: 0.8}, class: ResponseContractViolation},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw := responseJSON(t, unresolved.SourceItemCode, test.brand, validCategoryMatch(), validDescription(), []UnsupportedSemantic{})
			_, err := ParseEnrichmentResponse(raw, unresolved)
			requireResponseClass(t, err, test.class)
		})
	}
}

func TestStrictBrandProposeNewUsesVerbatimSourceLabel(t *testing.T) {
	const sourceItemName = "شامبو بانتين صحي ونظيف 24*400 مل"
	request := unresolvedRequest(t)
	request.Snapshot.SourceItemName = sourceItemName

	valid := responseJSON(t, request.SourceItemCode, BrandProposal{
		Action:        ActionProposeNew,
		CanonicalName: "  بانتين  ",
		Confidence:    0.8,
	}, validCategoryMatch(), validDescription(), []UnsupportedSemantic{})
	parsed, err := ParseEnrichmentResponse(valid, request)
	if err != nil {
		t.Fatalf("verbatim source-language brand proposal rejected: %v", err)
	}
	if parsed.Proposals.Brand.CanonicalName != "بانتين" {
		t.Fatalf("canonical source label=%q, want %q", parsed.Proposals.Brand.CanonicalName, "بانتين")
	}
	if parsed.Proposals.Brand.TargetID != nil || parsed.Proposals.Brand.TargetCode != "" {
		t.Fatalf("PROPOSE_NEW brand gained a target identity: %+v", parsed.Proposals.Brand)
	}

	translated := responseJSON(t, request.SourceItemCode, BrandProposal{
		Action:        ActionProposeNew,
		CanonicalName: "Pantene",
		Confidence:    0.8,
	}, validCategoryMatch(), validDescription(), []UnsupportedSemantic{})
	_, err = ParseEnrichmentResponse(translated, request)
	requireResponseClass(t, err, ResponseContractViolation)
	if !strings.Contains(err.Error(), "extracted verbatim from source_item_name") {
		t.Fatalf("unexpected source-label validation error: %v", err)
	}
}

func TestStrictCategoryValidationAndPrecedence(t *testing.T) {
	request := unresolvedRequest(t)
	for _, test := range []struct {
		name     string
		category CategoryProposal
		class    ResponseErrorClass
	}{
		{name: "canonical match passes", category: validCategoryMatch()},
		{name: "unknown match fails", category: CategoryProposal{Action: ActionMatchExisting, TargetID: int32Pointer(999), Confidence: 0.8}, class: ResponseCandidateMismatch},
		{name: "new review proposal passes", category: CategoryProposal{Action: ActionProposeNew, CanonicalName: "Hair Care", Confidence: 0.7}},
		{name: "new existing identity fails", category: CategoryProposal{Action: ActionProposeNew, TargetID: int32Pointer(20), CanonicalName: "Hair Care", Confidence: 0.7}, class: ResponseContractViolation},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw := responseJSON(t, request.SourceItemCode, validBrandMatch(), test.category, validDescription(), []UnsupportedSemantic{})
			_, err := ParseEnrichmentResponse(raw, request)
			if test.class == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				requireResponseClass(t, err, test.class)
			}
		})
	}

	populated := populatedCategoryRequest(t)
	for _, category := range []CategoryProposal{
		validCategoryKeep(),
		{Action: ActionMatchExisting, TargetID: int32Pointer(20), Confidence: 0.8},
		{Action: ActionProposeNew, CanonicalName: "Hair Care", Confidence: 0.8},
	} {
		raw := responseJSON(t, populated.SourceItemCode, validBrandNoMatch(), category, validDescription(), []UnsupportedSemantic{})
		_, err := ParseEnrichmentResponse(raw, populated)
		if category.Action == ActionKeepExisting {
			if err != nil {
				t.Fatal(err)
			}
		} else {
			requireResponseClass(t, err, ResponseContractViolation)
		}
	}
}

func TestStrictDescriptionValidation(t *testing.T) {
	missing := unresolvedRequest(t)
	for _, test := range []struct {
		name        string
		description DescriptionProposal
		class       ResponseErrorClass
	}{
		{name: "valid proposal", description: validDescription()},
		{name: "overlong unicode", description: DescriptionProposal{Action: ActionProposeNew, Value: strings.Repeat("ش", MaxDescriptionRunes+1), Confidence: 0.8}, class: ResponseContractViolation},
		{name: "whitespace only", description: DescriptionProposal{Action: ActionProposeNew, Value: "   ", Confidence: 0.8}, class: ResponseContractViolation},
		{name: "match existing rejected", description: DescriptionProposal{Action: ActionMatchExisting, Confidence: 0.8}, class: ResponseContractViolation},
		{name: "no match allowed", description: DescriptionProposal{Action: ActionNoMatch, Confidence: 0.2}},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw := responseJSON(t, missing.SourceItemCode, validBrandMatch(), validCategoryMatch(), test.description, []UnsupportedSemantic{})
			_, err := ParseEnrichmentResponse(raw, missing)
			if test.class == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				requireResponseClass(t, err, test.class)
			}
		})
	}

	populated := populatedCategoryRequest(t)
	populated.Snapshot.Description = "Existing description"
	populated.Gaps = GapsForSnapshot(populated.Snapshot)
	for _, description := range []DescriptionProposal{
		{Action: ActionKeepExisting, Confidence: 0.8},
		{Action: ActionProposeNew, Value: "Replacement", Confidence: 0.8},
	} {
		raw := responseJSON(t, populated.SourceItemCode, validBrandNoMatch(), validCategoryKeep(), description, []UnsupportedSemantic{})
		_, err := ParseEnrichmentResponse(raw, populated)
		if description.Action == ActionKeepExisting {
			if err != nil {
				t.Fatal(err)
			}
		} else {
			requireResponseClass(t, err, ResponseContractViolation)
		}
	}
}

func TestStrictUnsupportedSemanticsAndAuthoritativeProtection(t *testing.T) {
	request := unresolvedRequest(t)
	for _, key := range []string{"anti_dandruff", "size_text", "capacity", "model_number", "resolution_text", "packaging_text"} {
		raw := responseJSON(t, request.SourceItemCode, validBrandMatch(), validCategoryMatch(), validDescription(), []UnsupportedSemantic{{
			SemanticType: "product_evidence", Key: key, Value: json.RawMessage(`"evidence"`), Confidence: 0.8,
		}})
		if _, err := ParseEnrichmentResponse(raw, request); err != nil {
			t.Fatalf("unsupported evidence %q rejected: %v", key, err)
		}
	}
	for _, key := range []string{"productType", "ItemCode", "conversionFactor", "isSellable", "barcode ownership", "warehouse", "supplier"} {
		raw := responseJSON(t, request.SourceItemCode, validBrandMatch(), validCategoryMatch(), validDescription(), []UnsupportedSemantic{{
			SemanticType: "product_evidence", Key: key, Value: json.RawMessage(`"attempt"`), Confidence: 0.8,
		}})
		_, err := ParseEnrichmentResponse(raw, request)
		requireResponseClass(t, err, ResponseProhibitedOutput)
	}
	for _, value := range []string{`{"productType":"electronics"}`, `{"conversionFactor":24}`, `{"isActive":true}`} {
		raw := responseJSON(t, request.SourceItemCode, validBrandMatch(), validCategoryMatch(), validDescription(), []UnsupportedSemantic{{
			SemanticType: "product_evidence", Key: "evidence", Value: json.RawMessage(value), Confidence: 0.8,
		}})
		_, err := ParseEnrichmentResponse(raw, request)
		requireResponseClass(t, err, ResponseProhibitedOutput)
	}
}

func TestCandidateIndexRejectsAmbiguityAndUsesCanonicalIdentity(t *testing.T) {
	if _, err := NewCandidateIndex([]BrandCandidate{{ID: 1, Code: "A", Name: "A"}, {ID: 1, Code: "B", Name: "B"}}, nil); err == nil {
		t.Fatal("expected duplicate candidate ID to be rejected")
	}
	if _, err := NewCandidateIndex(nil, []CategoryCandidate{{ID: 1, Code: "A", Name: "A"}, {ID: 2, Code: "A", Name: "B"}}); err == nil {
		t.Fatal("expected duplicate candidate code to be rejected")
	}
	request := unresolvedRequest(t)
	body := responseMap(t, request.SourceItemCode, BrandProposal{Action: ActionMatchExisting, TargetCode: " PANTENE ", CanonicalName: "Attacker spelling", Confidence: 0.8}, validCategoryMatch(), validDescription(), []UnsupportedSemantic{})
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseEnrichmentResponse(raw, request)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Proposals.Brand.CanonicalName != "Pantene" || parsed.Proposals.Brand.TargetCode != "PANTENE" {
		t.Fatalf("model identity was trusted instead of canonical candidate: %+v", parsed.Proposals.Brand)
	}
}

func TestStrictRealSampleContracts(t *testing.T) {
	t.Run("Arabic Pantene", func(t *testing.T) {
		snapshot := EnrichmentSourceSnapshot{OrganizationID: 1, ProductID: 2, SourceSystem: SourceSystemSAP, SourceItemCode: "INV00006", SourceItemName: "شامبو بانتين صحي ونظيف 24*400 مل", ProductType: ProductTypeStandard, Category: &CategoryIdentity{ID: 50, Code: "BODY", Name: "العناية بالجسم"}}
		request, err := NewEnrichmentRequest(snapshot, []BrandCandidate{{ID: 10, Code: "PANTENE", Name: "Pantene"}}, nil)
		if err != nil {
			t.Fatal(err)
		}
		raw := responseJSON(t, request.SourceItemCode, validBrandMatch(), validCategoryKeep(), DescriptionProposal{Action: ActionProposeNew, Value: "شامبو بانتين صحي ونظيف بحجم 400 مل", Confidence: 0.9}, []UnsupportedSemantic{{SemanticType: "product_kind", Key: "shampoo", Value: json.RawMessage(`true`), Confidence: 0.9}, {SemanticType: "package", Key: "size_text", Value: json.RawMessage(`"400 مل"`), Confidence: 0.9}, {SemanticType: "package", Key: "packaging_text", Value: json.RawMessage(`"24*400 مل"`), Confidence: 0.9}})
		if _, err := ParseEnrichmentResponse(raw, request); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("anti dandruff packaging cannot become conversion", func(t *testing.T) {
		snapshot := EnrichmentSourceSnapshot{OrganizationID: 1, ProductID: 2, SourceSystem: SourceSystemSAP, SourceItemCode: "INV00007", SourceItemName: "شامبو بانتين ضد القشرة 24*400 مل", ProductType: ProductTypeStandard, Category: &CategoryIdentity{ID: 50, Code: "BODY", Name: "العناية بالجسم"}}
		request, err := NewEnrichmentRequest(snapshot, []BrandCandidate{{ID: 10, Code: "PANTENE", Name: "Pantene"}}, nil)
		if err != nil {
			t.Fatal(err)
		}
		raw := responseJSON(t, request.SourceItemCode, validBrandMatch(), validCategoryKeep(), validDescription(), []UnsupportedSemantic{{SemanticType: "product_evidence", Key: "anti_dandruff", Value: json.RawMessage(`true`), Confidence: 0.9}, {SemanticType: "uom_evidence", Key: "conversionFactor", Value: json.RawMessage(`24`), Confidence: 0.9}})
		_, err = ParseEnrichmentResponse(raw, request)
		requireResponseClass(t, err, ResponseProhibitedOutput)
	})

	t.Run("HIKvision fixed asset", func(t *testing.T) {
		snapshot := EnrichmentSourceSnapshot{OrganizationID: 1, ProductID: 2, SourceSystem: SourceSystemSAP, SourceItemCode: "INV00008", SourceItemName: "HIKvision DS 2.8mp (20@208.56)", ProductType: ProductTypeFixedAsset, Category: &CategoryIdentity{ID: 60, Code: "FIXED", Name: "Fixed Asset"}}
		request, err := NewEnrichmentRequest(snapshot, []BrandCandidate{{ID: 11, Code: "HIK", Name: "HIKvision"}}, nil)
		if err != nil {
			t.Fatal(err)
		}
		raw := responseJSON(t, request.SourceItemCode, BrandProposal{Action: ActionMatchExisting, TargetCode: "HIK", Confidence: 0.9}, validCategoryKeep(), validDescription(), []UnsupportedSemantic{{SemanticType: "product_evidence", Key: "resolution_text", Value: json.RawMessage(`"2.8mp"`), Confidence: 0.9}, {SemanticType: "product_evidence", Key: "model_number", Value: json.RawMessage(`"DS 2.8mp"`), Confidence: 0.8}})
		result, err := ParseEnrichmentResponse(raw, request)
		if err != nil {
			t.Fatal(err)
		}
		if request.Snapshot.ProductType != ProductTypeFixedAsset || result.Proposals.Brand.TargetCode != "HIK" {
			t.Fatal("fixed_asset context or brand candidate was altered")
		}
	})

	t.Run("Huawei resolved brand", func(t *testing.T) {
		snapshot := EnrichmentSourceSnapshot{OrganizationID: 1, ProductID: 2, SourceSystem: SourceSystemSAP, SourceItemCode: "INV00009", SourceItemName: "Huawei Computer, LCD DDR4 RAM", ProductType: ProductTypeStandard, Brand: &BrandIdentity{ID: 12, Code: "HUAWEI", Name: "Huawei"}, Category: &CategoryIdentity{ID: 70, Code: "ELECTRONICS", Name: "Electronics"}}
		request, err := NewEnrichmentRequest(snapshot, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		raw := responseJSON(t, request.SourceItemCode, BrandProposal{Action: ActionMatchExisting, TargetID: int32Pointer(12), Confidence: 0.9}, validCategoryKeep(), validDescription(), []UnsupportedSemantic{})
		_, err = ParseEnrichmentResponse(raw, request)
		requireResponseClass(t, err, ResponseContractViolation)
	})

	t.Run("Epson does not invent taxonomy", func(t *testing.T) {
		snapshot := EnrichmentSourceSnapshot{OrganizationID: 1, ProductID: 2, SourceSystem: SourceSystemSAP, SourceItemCode: "INV00010", SourceItemName: "Epson Color Printer", ProductType: ProductTypeStandard, Category: &CategoryIdentity{ID: 71, Code: "OFFICE", Name: "Office"}}
		request, err := NewEnrichmentRequest(snapshot, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		raw := responseJSON(t, request.SourceItemCode, BrandProposal{Action: ActionProposeNew, CanonicalName: "Epson", Confidence: 0.7}, validCategoryKeep(), validDescription(), []UnsupportedSemantic{})
		if _, err := ParseEnrichmentResponse(raw, request); err != nil {
			t.Fatal(err)
		}
	})
}

func TestBuildInferenceInstructionsIsProviderNeutral(t *testing.T) {
	text := BuildInferenceInstructions(unresolvedRequest(t))
	for _, expected := range []string{"exactly one JSON object", "KEEP_EXISTING", "product_type", "conversion factor", "unsupported_semantics", "source language", "verbatim", "never translate", "transliterate", "MATCH_EXISTING instead"} {
		if !strings.Contains(strings.ToLower(text), strings.ToLower(expected)) {
			t.Errorf("instruction contract missing %q", expected)
		}
	}
	for _, forbidden := range []string{"OpenAI", "Anthropic", "temperature", "response_format", "api_key", "tool_choice"} {
		if strings.Contains(strings.ToLower(text), strings.ToLower(forbidden)) {
			t.Errorf("provider-specific instruction leaked %q", forbidden)
		}
	}
}

func unresolvedRequest(t *testing.T) EnrichmentRequest {
	t.Helper()
	snapshot := EnrichmentSourceSnapshot{OrganizationID: 1, ProductID: 2, SourceSystem: SourceSystemSAP, SourceItemCode: "INV00006", SourceItemName: "Sample item", ProductType: ProductTypeStandard}
	request, err := NewEnrichmentRequest(snapshot, []BrandCandidate{{ID: 10, Code: "PANTENE", Name: "Pantene"}, {ID: 11, Code: "HIK", Name: "HIKvision"}}, []CategoryCandidate{{ID: 20, Code: "HAIR", Name: "Hair Care"}})
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func resolvedBrandRequest(t *testing.T) EnrichmentRequest {
	t.Helper()
	snapshot := EnrichmentSourceSnapshot{OrganizationID: 1, ProductID: 2, SourceSystem: SourceSystemSAP, SourceItemCode: "RESOLVED", SourceItemName: "Huawei Computer", ProductType: ProductTypeStandard, Brand: &BrandIdentity{ID: 12, Code: "HUAWEI", Name: "Huawei"}, Category: &CategoryIdentity{ID: 20, Code: "HAIR", Name: "Hair Care"}}
	request, err := NewEnrichmentRequest(snapshot, []BrandCandidate{{ID: 12, Code: "HUAWEI", Name: "Huawei"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func populatedCategoryRequest(t *testing.T) EnrichmentRequest {
	t.Helper()
	snapshot := EnrichmentSourceSnapshot{OrganizationID: 1, ProductID: 2, SourceSystem: SourceSystemSAP, SourceItemCode: "CATEGORY", SourceItemName: "Shampoo", ProductType: ProductTypeStandard, Category: &CategoryIdentity{ID: 20, Code: "BODY", Name: "العناية بالجسم"}}
	request, err := NewEnrichmentRequest(snapshot, nil, []CategoryCandidate{{ID: 20, Code: "HAIR", Name: "Hair Care"}})
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func validBrandMatch() BrandProposal {
	return BrandProposal{Action: ActionMatchExisting, TargetID: int32Pointer(10), TargetCode: "PANTENE", Confidence: 0.9}
}

func validBrandNoMatch() BrandProposal { return BrandProposal{Action: ActionNoMatch, Confidence: 0.3} }

func validCategoryMatch() CategoryProposal {
	return CategoryProposal{Action: ActionMatchExisting, TargetID: int32Pointer(20), TargetCode: "HAIR", Confidence: 0.8}
}

func validCategoryKeep() CategoryProposal {
	return CategoryProposal{Action: ActionKeepExisting, Confidence: 0.8}
}

func validDescription() DescriptionProposal {
	return DescriptionProposal{Action: ActionProposeNew, Value: "Factual catalog description", Confidence: 0.8}
}

func responseJSON(t *testing.T, source string, brand any, category any, description any, unsupported []UnsupportedSemantic) []byte {
	t.Helper()
	raw, err := json.Marshal(responseMap(t, source, brand, category, description, unsupported))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func responseMap(t *testing.T, source string, brand any, category any, description any, unsupported []UnsupportedSemantic) map[string]any {
	t.Helper()
	return map[string]any{"source_item_code": source, "brand": brand, "category": category, "description": description, "unsupported_semantics": unsupported}
}

func responseWithExtra(t *testing.T, raw []byte, key string, value any) []byte {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	body[key] = value
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func responseWithNestedExtra(t *testing.T, raw []byte, proposal string, key string, value any) []byte {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	nested := body[proposal].(map[string]any)
	nested[key] = value
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func requireResponseClass(t *testing.T, err error, want ResponseErrorClass) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s, got nil", want)
	}
	if got := ResponseErrorClassOf(err); got != want {
		t.Fatalf("error class = %q, want %q: %v", got, want, err)
	}
}
