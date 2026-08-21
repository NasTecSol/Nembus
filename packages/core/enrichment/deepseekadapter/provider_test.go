package deepseekadapter

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/enrichment"
)

func TestProviderUsesConfiguredEndpointModelJSONModeAndSafeInput(t *testing.T) {
	request := adapterRequest(t)
	var gotRequest struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		ResponseFormat struct {
			Type string `json:"type"`
		} `json:"response_format"`
		Thinking        map[string]json.RawMessage `json:"thinking"`
		MaxTokens       int                        `json:"max_tokens"`
		Stream          bool                       `json:"stream"`
		ReasoningEffort json.RawMessage            `json:"reasoning_effort"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/custom/chat/completions" {
			t.Errorf("path=%s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-deepseek-key" {
			t.Errorf("authorization header=%q", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(body, &gotRequest); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"ds-test","model":"deepseek-v4-flash","choices":[{"finish_reason":"stop","message":{"content":`+strconvQuote(validResponse(request.SourceItemCode))+`}}]}`)
	}))
	defer server.Close()

	provider, err := New("test-deepseek-key", server.URL+"/custom", "configured-model", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	result, err := provider.Enrich(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Provider != ProviderName || result.Model != "deepseek-v4-flash" || result.ResponseID != "ds-test" {
		t.Fatalf("trusted metadata=%+v", result)
	}
	if gotRequest.Model != "configured-model" || gotRequest.ResponseFormat.Type != "json_object" || gotRequest.MaxTokens != 2048 || gotRequest.Stream {
		t.Fatalf("request configuration=%+v", gotRequest)
	}
	if len(gotRequest.Thinking) != 1 {
		t.Fatalf("thinking configuration=%v, want exactly {type: disabled}", gotRequest.Thinking)
	}
	var thinkingType string
	if err := json.Unmarshal(gotRequest.Thinking["type"], &thinkingType); err != nil || thinkingType != "disabled" {
		t.Fatalf("thinking.type=%q, want disabled", thinkingType)
	}
	if len(gotRequest.ReasoningEffort) != 0 {
		t.Fatalf("reasoning_effort should not be sent: %s", gotRequest.ReasoningEffort)
	}
	if len(gotRequest.Messages) != 2 || gotRequest.Messages[0].Role != "system" || gotRequest.Messages[1].Role != "user" {
		t.Fatalf("messages=%+v", gotRequest.Messages)
	}
	if !strings.Contains(gotRequest.Messages[0].Content, "Return JSON") || !strings.Contains(gotRequest.Messages[0].Content, "source_item_code") {
		t.Fatalf("JSON output instruction missing: %s", gotRequest.Messages[0].Content)
	}
	if strings.Contains(gotRequest.Messages[1].Content, "conversion_factor") || strings.Contains(gotRequest.Messages[1].Content, "inventory") || strings.Contains(gotRequest.Messages[1].Content, "price") || strings.Contains(gotRequest.Messages[1].Content, "supplier") {
		t.Fatalf("unsafe provider input leaked: %s", gotRequest.Messages[1].Content)
	}
}

func TestDeepSeekInstructionsStateExactStage2BShapeAndActionRules(t *testing.T) {
	instructions := buildJSONInstructions()
	for _, expected := range []string{
		"brand and category are objects with required action and confidence members",
		"description is an object with required action and confidence members",
		"canonical content member is value",
		"unsupported_semantics is an array of objects with required semantic_type, key, value, and confidence members",
		"Evidence is optional. Omit an evidence member when no evidence is available.",
		"Whenever evidence is present on brand, category, description, or an unsupported_semantics item, it MUST be a JSON array of strings.",
		`"evidence":["first source fact","second source fact"]`,
		`"evidence":["single source fact"]`,
		`NEVER return a scalar string such as "evidence":"single source fact"`,
		"resolved structured brand requires KEEP_EXISTING",
		"unresolved brand requires MATCH_EXISTING against a supplied exact candidate, PROPOSE_NEW, or NO_MATCH",
		"populated structured category requires KEEP_EXISTING",
		"missing category requires MATCH_EXISTING against a supplied exact candidate, PROPOSE_NEW, or NO_MATCH",
		"existing description requires KEEP_EXISTING",
		"missing description requires PROPOSE_NEW with value or NO_MATCH",
		"Product type remains immutable context",
	} {
		if !strings.Contains(instructions, expected) {
			t.Errorf("instructions missing %q:\n%s", expected, instructions)
		}
	}
	for _, forbidden := range []string{"description.text", "description_text"} {
		if strings.Contains(instructions, forbidden) {
			t.Errorf("instructions contain noncanonical alias %q", forbidden)
		}
	}
}

func TestProviderAcceptsEvidenceArraysAndOmission(t *testing.T) {
	request := adapterRequest(t)
	for _, test := range []struct {
		name string
		raw  string
	}{
		{name: "single evidence string in arrays", raw: strings.ReplaceAll(validResponse(request.SourceItemCode), `"evidence":[]`, `"evidence":["fact"]`)},
		{name: "multiple evidence strings in arrays", raw: strings.ReplaceAll(validResponse(request.SourceItemCode), `"evidence":[]`, `"evidence":["first fact","second fact"]`)},
		{name: "optional evidence omitted", raw: strings.ReplaceAll(validResponse(request.SourceItemCode), `,"evidence":[]`, "")},
	} {
		t.Run(test.name, func(t *testing.T) {
			provider, err := newWithClient(fakeHTTPClient{body: `{"id":"ds-test","model":"model","choices":[{"finish_reason":"stop","message":{"content":` + strconvQuote(test.raw) + `}}]}`}, "https://deepseek.test", "test-key", "model", time.Second)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := provider.Enrich(context.Background(), request); err != nil {
				t.Fatalf("canonical evidence response rejected: %v", err)
			}
		})
	}
}

func TestProviderStrictParserAndBusinessRules(t *testing.T) {
	request := adapterRequest(t)
	for _, test := range []struct {
		name  string
		raw   string
		class enrichment.ResponseErrorClass
	}{
		{name: "malformed", raw: "not-json", class: enrichment.ResponseMalformed},
		{name: "unknown field", raw: strings.Replace(validResponse(request.SourceItemCode), "}", `,"unknown":true}`, 1), class: enrichment.ResponseContractViolation},
		{name: "noncanonical description text", raw: strings.Replace(validResponse(request.SourceItemCode), `"value":"Pantene shampoo 400ml"`, `"text":"Pantene shampoo 400ml"`, 1), class: enrichment.ResponseContractViolation},
		{name: "brand scalar evidence", raw: responseWithScalarEvidence(validResponse(request.SourceItemCode), "brand"), class: enrichment.ResponseContractViolation},
		{name: "category scalar evidence", raw: responseWithScalarEvidence(validResponse(request.SourceItemCode), "category"), class: enrichment.ResponseContractViolation},
		{name: "description scalar evidence", raw: responseWithScalarEvidence(validResponse(request.SourceItemCode), "description"), class: enrichment.ResponseContractViolation},
		{name: "unsupported semantic scalar evidence", raw: responseWithScalarEvidence(validResponse(request.SourceItemCode), "unsupported_semantics"), class: enrichment.ResponseContractViolation},
		{name: "prohibited product type", raw: strings.Replace(validResponse(request.SourceItemCode), `"unsupported_semantics":[]`, `"unsupported_semantics":[{"semantic_type":"productType","key":"product_type","value":"fixed_asset","confidence":0.9,"evidence":[],"explanation":null}]`, 1), class: enrichment.ResponseProhibitedOutput},
		{name: "UoM conversion", raw: strings.Replace(validResponse(request.SourceItemCode), `"unsupported_semantics":[]`, `"unsupported_semantics":[{"semantic_type":"packaging","key":"conversion_factor","value":24,"confidence":0.9,"evidence":[],"explanation":null}]`, 1), class: enrichment.ResponseProhibitedOutput},
		{name: "category override", raw: strings.Replace(validResponse(request.SourceItemCode), `"category":{"action":"NO_MATCH","target_id":null,"target_code":"","canonical_name":"","confidence":0.2`, `"category":{"action":"PROPOSE_NEW","target_id":null,"target_code":"new","canonical_name":"New","confidence":0.9`, 1), class: enrichment.ResponseContractViolation},
		{name: "correlation mismatch", raw: validResponse("OTHER"), class: enrichment.ResponseCorrelationMismatch},
	} {
		t.Run(test.name, func(t *testing.T) {
			provider, err := newWithClient(fakeHTTPClient{body: `{"id":"ds-test","model":"configured-model","choices":[{"finish_reason":"stop","message":{"content":` + strconvQuote(test.raw) + `}}]}`}, "https://deepseek.test", "test-key", "configured-model", time.Second)
			if err != nil {
				t.Fatal(err)
			}
			_, err = provider.Enrich(context.Background(), request)
			if got := enrichment.ResponseErrorClassOf(err); got != test.class {
				t.Fatalf("response class=%s want %s err=%v", got, test.class, err)
			}
		})
	}
}

func TestProviderRejectsInvalidKeepExistingActionsForUnresolvedFields(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(string) string
	}{
		{name: "unresolved brand", mutate: func(raw string) string {
			return strings.Replace(raw, `"brand":{"action":"MATCH_EXISTING"`, `"brand":{"action":"KEEP_EXISTING"`, 1)
		}},
		{name: "missing category", mutate: func(raw string) string {
			return strings.Replace(raw, `"category":{"action":"NO_MATCH"`, `"category":{"action":"KEEP_EXISTING"`, 1)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := adapterRequest(t)
			raw := test.mutate(validResponse(request.SourceItemCode))
			provider, err := newWithClient(fakeHTTPClient{body: `{"id":"ds-test","model":"model","choices":[{"finish_reason":"stop","message":{"content":` + strconvQuote(raw) + `}}]}`}, "https://deepseek.test", "test-key", "model", time.Second)
			if err != nil {
				t.Fatal(err)
			}
			_, err = provider.Enrich(context.Background(), request)
			if got := enrichment.ResponseErrorClassOf(err); got != enrichment.ResponseContractViolation {
				t.Fatalf("response class=%s want %s err=%v", got, enrichment.ResponseContractViolation, err)
			}
		})
	}
}

func TestProviderClassifiesHTTPAndTransportFailures(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		class  enrichment.ProviderErrorClass
	}{
		{name: "unauthorized", status: http.StatusUnauthorized, class: enrichment.ProviderErrorPermanent},
		{name: "rate limit", status: http.StatusTooManyRequests, class: enrichment.ProviderErrorRetryable},
		{name: "server", status: http.StatusInternalServerError, class: enrichment.ProviderErrorRetryable},
	} {
		t.Run(test.name, func(t *testing.T) {
			provider, err := newWithClient(fakeHTTPClient{status: test.status}, "https://deepseek.test", "test-key", "model", time.Second)
			if err != nil {
				t.Fatal(err)
			}
			_, err = provider.Enrich(context.Background(), adapterRequest(t))
			if got := enrichment.ProviderErrorClassOf(err); got != test.class {
				t.Fatalf("provider class=%s want %s", got, test.class)
			}
		})
	}

	for _, test := range []struct {
		name string
		err  error
		code string
	}{
		{name: "timeout", err: context.DeadlineExceeded, code: "provider_timeout"},
		{name: "connection", err: errors.New("connection refused"), code: "provider_network_error"},
	} {
		t.Run(test.name, func(t *testing.T) {
			provider, err := newWithClient(fakeHTTPClient{err: test.err}, "https://deepseek.test", "test-key", "model", time.Second)
			if err != nil {
				t.Fatal(err)
			}
			_, err = provider.Enrich(context.Background(), adapterRequest(t))
			providerErr := new(enrichment.ProviderError)
			if !errors.As(err, &providerErr) || providerErr.Code != test.code {
				t.Fatalf("error=%v code=%s", err, test.code)
			}
		})
	}
}

func TestProviderRejectsEmptyContentAndInvalidConfiguration(t *testing.T) {
	provider, err := newWithClient(fakeHTTPClient{body: `{"id":"ds-test","model":"model","choices":[]}`}, "https://deepseek.test", "test-key", "model", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_, err = provider.Enrich(context.Background(), adapterRequest(t))
	if got := enrichment.ResponseErrorClassOf(err); got != enrichment.ResponseMalformed {
		t.Fatalf("empty choices class=%s err=%v", got, err)
	}
	if _, err := New("", "", "model", time.Second); err == nil {
		t.Fatal("missing key should fail")
	}
	if _, err := New("key", "not-a-url", "model", time.Second); err == nil {
		t.Fatal("invalid base URL should fail")
	}
}

func TestProviderPreservesFixedAssetProductTypeForRepresentativeNames(t *testing.T) {
	for _, name := range []string{"HIKvision camera", "Huawei network device", "Epson printer"} {
		t.Run(name, func(t *testing.T) {
			request := adapterRequest(t)
			request.Snapshot.SourceItemName = name
			request.Snapshot.ProductType = enrichment.ProductTypeFixedAsset
			provider, err := newWithClient(fakeHTTPClient{body: `{"id":"ds-test","model":"model","choices":[{"finish_reason":"stop","message":{"content":` + strconvQuote(validResponse(request.SourceItemCode)) + `}}]}`}, "https://deepseek.test", "test-key", "model", time.Second)
			if err != nil {
				t.Fatal(err)
			}
			result, err := provider.Enrich(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if result.Proposals.Brand == nil || result.Proposals.Category == nil || result.Proposals.Description == nil {
				t.Fatal("provider result did not pass the provider-neutral proposal contract")
			}
		})
	}
}

func TestProviderRejectsStructuredBrandAndCategoryOverrides(t *testing.T) {
	request := adapterRequest(t)
	request.Snapshot.Brand = &enrichment.BrandIdentity{ID: 10, Code: "PANTENE", Name: "Pantene"}
	request.Snapshot.Category = &enrichment.CategoryIdentity{ID: 20, Code: "CAT-SHAMPOO", Name: "Shampoo"}
	request.Gaps = enrichment.GapsForSnapshot(request.Snapshot)
	// The model response attempts to replace both resolved structured values.
	raw := validResponse(request.SourceItemCode)
	raw = strings.Replace(raw, `"brand":{"action":"MATCH_EXISTING","target_id":10,"target_code":"PANTENE","canonical_name":"","confidence":0.9`, `"brand":{"action":"PROPOSE_NEW","target_id":null,"target_code":"new","canonical_name":"New Brand","confidence":0.9`, 1)
	raw = strings.Replace(raw, `"category":{"action":"NO_MATCH","target_id":null,"target_code":"","canonical_name":"","confidence":0.2`, `"category":{"action":"PROPOSE_NEW","target_id":null,"target_code":"new","canonical_name":"New Category","confidence":0.9`, 1)
	provider, err := newWithClient(fakeHTTPClient{body: `{"id":"ds-test","model":"model","choices":[{"finish_reason":"stop","message":{"content":` + strconvQuote(raw) + `}}]}`}, "https://deepseek.test", "test-key", "model", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_, err = provider.Enrich(context.Background(), request)
	if got := enrichment.ResponseErrorClassOf(err); got != enrichment.ResponseContractViolation {
		t.Fatalf("structured precedence class=%s err=%v", got, err)
	}
}

type fakeHTTPClient struct {
	status int
	body   string
	err    error
}

func (f fakeHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	if f.err != nil {
		return nil, f.err
	}
	status := f.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(f.body))}, nil
}

func validResponse(sourceItemCode string) string {
	return `{"source_item_code":` + strconvQuote(sourceItemCode) + `,"brand":{"action":"MATCH_EXISTING","target_id":10,"target_code":"PANTENE","canonical_name":"","confidence":0.9,"evidence":[],"explanation":null},"category":{"action":"NO_MATCH","target_id":null,"target_code":"","canonical_name":"","confidence":0.2,"evidence":[],"explanation":null},"description":{"action":"PROPOSE_NEW","value":"Pantene shampoo 400ml","confidence":0.8,"evidence":[],"explanation":null},"unsupported_semantics":[]}`
}

func responseWithScalarEvidence(raw, field string) string {
	switch field {
	case "brand":
		return strings.Replace(raw, `"brand":{"action":"MATCH_EXISTING","target_id":10,"target_code":"PANTENE","canonical_name":"","confidence":0.9,"evidence":[]`, `"brand":{"action":"MATCH_EXISTING","target_id":10,"target_code":"PANTENE","canonical_name":"","confidence":0.9,"evidence":"fact"`, 1)
	case "category":
		return strings.Replace(raw, `"category":{"action":"NO_MATCH","target_id":null,"target_code":"","canonical_name":"","confidence":0.2,"evidence":[]`, `"category":{"action":"NO_MATCH","target_id":null,"target_code":"","canonical_name":"","confidence":0.2,"evidence":"fact"`, 1)
	case "description":
		return strings.Replace(raw, `"description":{"action":"PROPOSE_NEW","value":"Pantene shampoo 400ml","confidence":0.8,"evidence":[]`, `"description":{"action":"PROPOSE_NEW","value":"Pantene shampoo 400ml","confidence":0.8,"evidence":"fact"`, 1)
	case "unsupported_semantics":
		return strings.Replace(raw, `"unsupported_semantics":[]`, `"unsupported_semantics":[{"semantic_type":"attribute","key":"size","value":"400 ml","confidence":0.9,"evidence":"fact","explanation":null}]`, 1)
	default:
		return raw
	}
}

func strconvQuote(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func adapterRequest(t *testing.T) enrichment.EnrichmentRequest {
	t.Helper()
	snapshot := enrichment.EnrichmentSourceSnapshot{
		OrganizationID: 2,
		ProductID:      3,
		SourceSystem:   enrichment.SourceSystemSAP,
		SourceItemCode: "INV00006",
		SourceItemName: "Pantene shampoo 24*400 ml",
		ProductType:    enrichment.ProductTypeStandard,
		UOM:            enrichment.UOMContext{Conversions: []enrichment.UOMConversionContext{{ConversionFactor: "24"}}},
	}
	request, err := enrichment.NewEnrichmentRequest(snapshot, []enrichment.BrandCandidate{{ID: 10, Code: "PANTENE", Name: "Pantene"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return request
}
