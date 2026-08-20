// Package openaiadapter contains the only OpenAI-specific implementation of
// ProductEnrichmentProvider. The parent enrichment package remains the
// provider-neutral contract and strict Nembus validation boundary.
package openaiadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

const (
	ProviderName = "openai"
	SchemaName   = "nembus_product_enrichment"
)

type responseClient interface {
	New(context.Context, responses.ResponseNewParams, ...option.RequestOption) (*responses.Response, error)
}

type Provider struct {
	client  responseClient
	model   string
	timeout time.Duration
}

func New(apiKey, model string, timeout time.Duration) (*Provider, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("OpenAI enrichment model is required")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("OpenAI enrichment timeout must be positive")
	}
	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithHTTPClient(&http.Client{Timeout: timeout}))
	return &Provider{client: &client.Responses, model: model, timeout: timeout}, nil
}

func newWithClient(client responseClient, model string, timeout time.Duration) *Provider {
	return &Provider{client: client, model: model, timeout: timeout}
}

func (p *Provider) Enrich(ctx context.Context, request enrichment.EnrichmentRequest) (enrichment.EnrichmentResult, error) {
	if p == nil || p.client == nil {
		return enrichment.EnrichmentResult{}, &enrichment.ProviderError{Class: enrichment.ProviderErrorPermanent, Code: "provider_not_configured", Err: fmt.Errorf("OpenAI provider is not configured")}
	}
	if err := request.Validate(); err != nil {
		return enrichment.EnrichmentResult{}, &enrichment.ProviderError{Class: enrichment.ProviderErrorPermanent, Code: "request_contract_failed", Err: err}
	}

	callCtx := ctx
	cancel := func() {}
	if p.timeout > 0 {
		if _, hasDeadline := ctx.Deadline(); !hasDeadline {
			callCtx, cancel = context.WithTimeout(ctx, p.timeout)
		}
	}
	defer cancel()

	payload, err := json.Marshal(safeProviderInput(request))
	if err != nil {
		return enrichment.EnrichmentResult{}, &enrichment.ProviderError{Class: enrichment.ProviderErrorPermanent, Code: "request_encoding_failed", Err: err}
	}
	response, err := p.client.New(callCtx, responses.ResponseNewParams{
		Model:        shared.ResponsesModel(p.model),
		Instructions: openai.String(enrichment.BuildInferenceInstructions(request)),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(string(payload)),
		},
		Store: param.NewOpt(false),
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:   SchemaName,
					Schema: StrictResponseSchema(),
					Strict: param.NewOpt(true),
				},
			},
		},
	})
	if err != nil {
		return enrichment.EnrichmentResult{}, classifyOpenAIError(callCtx, err)
	}
	if response == nil {
		return enrichment.EnrichmentResult{}, &enrichment.ProviderError{Class: enrichment.ProviderErrorRetryable, Code: "empty_provider_response", Err: fmt.Errorf("OpenAI returned an empty response")}
	}

	result, err := enrichment.ParseEnrichmentResponseString(response.OutputText(), request)
	if err != nil {
		// The strict Nembus parser is always applied after SDK extraction. Its
		// response errors are permanent contract failures, never retries.
		return enrichment.EnrichmentResult{}, err
	}
	result.Provider = ProviderName
	result.Model = string(response.Model)
	if result.Model == "" {
		result.Model = p.model
	}
	result.ResponseID = response.ID
	return result, nil
}

type providerInput struct {
	ContractVersion    string                              `json:"contract_version"`
	RequestVersion     string                              `json:"request_version"`
	SourceItemCode     string                              `json:"source_item_code"`
	Snapshot           enrichment.EnrichmentSourceSnapshot `json:"snapshot"`
	Gaps               []enrichment.EnrichmentGap          `json:"gaps"`
	BrandCandidates    []enrichment.BrandCandidate         `json:"brand_candidates,omitempty"`
	CategoryCandidates []enrichment.CategoryCandidate      `json:"category_candidates,omitempty"`
	Policy             enrichment.EnrichmentRequestPolicy  `json:"policy"`
}

func safeProviderInput(request enrichment.EnrichmentRequest) providerInput {
	snapshot := request.Snapshot
	// UoM identity is useful immutable context. Conversion factors are not sent
	// to the provider, so they cannot be mistaken for an inference target.
	snapshot.UOM.Conversions = nil
	return providerInput{
		ContractVersion: request.ContractVersion, RequestVersion: request.RequestVersion,
		SourceItemCode: request.SourceItemCode, Snapshot: snapshot, Gaps: request.Gaps,
		BrandCandidates: request.BrandCandidates, CategoryCandidates: request.CategoryCandidates,
		Policy: request.Policy,
	}
}

func classifyOpenAIError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return ctx.Err()
		}
		return &enrichment.ProviderError{Class: enrichment.ProviderErrorRetryable, Code: "provider_timeout", Err: ctx.Err()}
	}
	var apiErr *responses.Error
	if errors.As(err, &apiErr) {
		code := fmt.Sprintf("http_%d", apiErr.StatusCode)
		class := enrichment.ProviderErrorPermanent
		if apiErr.StatusCode == http.StatusRequestTimeout || apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= 500 {
			class = enrichment.ProviderErrorRetryable
		}
		// Keep SDK response bodies out of logs and durable error state. The
		// normalized status code is sufficient for retry and diagnosis.
		return &enrichment.ProviderError{Class: class, Code: code, Err: fmt.Errorf("OpenAI API returned HTTP status %d", apiErr.StatusCode)}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return &enrichment.ProviderError{Class: enrichment.ProviderErrorRetryable, Code: "provider_timeout", Err: err}
	}
	return &enrichment.ProviderError{Class: enrichment.ProviderErrorRetryable, Code: "provider_network_error", Err: err}
}

// StrictResponseSchema is explicit and intentionally owned by this adapter.
// It mirrors the Stage 2B response DTO and contains no product_type or
// operational mutation fields.
func StrictResponseSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []any{"source_item_code", "brand", "category", "description", "unsupported_semantics"},
		"properties": map[string]any{
			"source_item_code": map[string]any{"type": "string"},
			"brand":            proposalSchema(false),
			"category":         proposalSchema(false),
			"description":      descriptionSchema(),
			"unsupported_semantics": map[string]any{
				"type":  "array",
				"items": unsupportedSemanticSchema(),
			},
		},
	}
}

func proposalSchema(_ bool) map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []any{"action", "target_id", "target_code", "canonical_name", "confidence", "evidence", "explanation"},
		"properties": map[string]any{
			"action":         map[string]any{"type": "string", "enum": []any{"KEEP_EXISTING", "MATCH_EXISTING", "PROPOSE_NEW", "NO_MATCH", "UNSUPPORTED_TARGET"}},
			"target_id":      nullable("integer"),
			"target_code":    nullable("string"),
			"canonical_name": nullable("string"),
			"confidence":     map[string]any{"type": "number", "minimum": 0, "maximum": 1},
			"evidence":       stringArraySchema(),
			"explanation":    nullable("string"),
		},
	}
}

func descriptionSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []any{"action", "value", "confidence", "evidence", "explanation"},
		"properties": map[string]any{
			"action":      map[string]any{"type": "string", "enum": []any{"KEEP_EXISTING", "MATCH_EXISTING", "PROPOSE_NEW", "NO_MATCH", "UNSUPPORTED_TARGET"}},
			"value":       nullable("string"),
			"confidence":  map[string]any{"type": "number", "minimum": 0, "maximum": 1},
			"evidence":    stringArraySchema(),
			"explanation": nullable("string"),
		},
	}
}

func unsupportedSemanticSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []any{"semantic_type", "key", "value", "confidence", "evidence", "explanation"},
		"properties": map[string]any{
			"semantic_type": map[string]any{"type": "string"},
			"key":           map[string]any{"type": "string"},
			"value":         map[string]any{},
			"confidence":    map[string]any{"type": "number", "minimum": 0, "maximum": 1},
			"evidence":      stringArraySchema(),
			"explanation":   nullable("string"),
		},
	}
}

func nullable(kind string) map[string]any { return map[string]any{"type": []any{kind, "null"}} }

func stringArraySchema() map[string]any {
	return map[string]any{"type": "array", "items": map[string]any{"type": "string"}}
}
