// Package deepseekadapter contains the DeepSeek-specific Chat Completions
// transport for the provider-neutral product enrichment contract.
package deepseekadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/enrichment"
)

const (
	ProviderName     = "deepseek"
	DefaultBaseURL   = "https://api.deepseek.com"
	DefaultModel     = "deepseek-v4-flash"
	chatCompletions  = "/chat/completions"
	maxResponseBytes = 2 << 20
)

type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type Provider struct {
	client  httpDoer
	baseURL string
	apiKey  string
	model   string
	timeout time.Duration
}

func New(apiKey, baseURL, model string, timeout time.Duration) (*Provider, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("DeepSeek API key is required")
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("DeepSeek enrichment model is required")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("DeepSeek enrichment timeout must be positive")
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultBaseURL
	}
	return newWithClient(&http.Client{Timeout: timeout}, baseURL, apiKey, model, timeout)
}

func newWithClient(client httpDoer, baseURL, apiKey, model string, timeout time.Duration) (*Provider, error) {
	if client == nil {
		return nil, fmt.Errorf("DeepSeek HTTP client is required")
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("DeepSeek API key is required")
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("DeepSeek enrichment model is required")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("DeepSeek enrichment timeout must be positive")
	}
	normalizedURL, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	return &Provider{
		client:  client,
		baseURL: normalizedURL,
		apiKey:  apiKey,
		model:   strings.TrimSpace(model),
		timeout: timeout,
	}, nil
}

func normalizeBaseURL(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("DeepSeek base URL must be an absolute HTTP(S) URL")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("DeepSeek base URL must not contain credentials, query parameters, or fragments")
	}
	return strings.TrimRight(parsed.String(), "/"), nil
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
	// UoM identity is immutable context. Conversion factors are deliberately
	// removed so they cannot be inferred or proposed by the provider.
	snapshot.UOM.Conversions = nil
	return providerInput{
		ContractVersion:    request.ContractVersion,
		RequestVersion:     request.RequestVersion,
		SourceItemCode:     request.SourceItemCode,
		Snapshot:           snapshot,
		Gaps:               request.Gaps,
		BrandCandidates:    request.BrandCandidates,
		CategoryCandidates: request.CategoryCandidates,
		Policy:             request.Policy,
	}
}

type chatCompletionRequest struct {
	Model          string         `json:"model"`
	Messages       []chatMessage  `json:"messages"`
	ResponseFormat responseFormat `json:"response_format"`
	Thinking       thinkingConfig `json:"thinking"`
	MaxTokens      int            `json:"max_tokens"`
	Temperature    float64        `json:"temperature"`
	Stream         bool           `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type thinkingConfig struct {
	Type string `json:"type"`
}

type chatCompletionResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (p *Provider) Enrich(ctx context.Context, request enrichment.EnrichmentRequest) (enrichment.EnrichmentResult, error) {
	if p == nil || p.client == nil {
		return enrichment.EnrichmentResult{}, &enrichment.ProviderError{
			Class: enrichment.ProviderErrorPermanent,
			Code:  "provider_not_configured",
			Err:   fmt.Errorf("DeepSeek provider is not configured"),
		}
	}
	if err := request.Validate(); err != nil {
		return enrichment.EnrichmentResult{}, &enrichment.ProviderError{
			Class: enrichment.ProviderErrorPermanent,
			Code:  "request_contract_failed",
			Err:   err,
		}
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
		return enrichment.EnrichmentResult{}, &enrichment.ProviderError{
			Class: enrichment.ProviderErrorPermanent,
			Code:  "request_encoding_failed",
			Err:   err,
		}
	}
	body, err := json.Marshal(chatCompletionRequest{
		Model: p.model,
		Messages: []chatMessage{
			{Role: "system", Content: buildJSONInstructions()},
			{Role: "user", Content: string(payload)},
		},
		ResponseFormat: responseFormat{Type: "json_object"},
		Thinking:       thinkingConfig{Type: "disabled"},
		MaxTokens:      2048,
		Temperature:    0,
		Stream:         false,
	})
	if err != nil {
		return enrichment.EnrichmentResult{}, &enrichment.ProviderError{
			Class: enrichment.ProviderErrorPermanent,
			Code:  "request_encoding_failed",
			Err:   err,
		}
	}

	endpoint := p.baseURL + chatCompletions
	httpRequest, err := http.NewRequestWithContext(callCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return enrichment.EnrichmentResult{}, &enrichment.ProviderError{
			Class: enrichment.ProviderErrorPermanent,
			Code:  "request_configuration_failed",
			Err:   err,
		}
	}
	httpRequest.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")

	response, err := p.client.Do(httpRequest)
	if err != nil {
		return enrichment.EnrichmentResult{}, classifyDeepSeekError(callCtx, err)
	}
	if response == nil {
		return enrichment.EnrichmentResult{}, responseMalformed("DeepSeek returned an empty HTTP response")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return enrichment.EnrichmentResult{}, classifyHTTPStatus(response.StatusCode)
	}

	raw, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return enrichment.EnrichmentResult{}, &enrichment.ProviderError{
			Class: enrichment.ProviderErrorRetryable,
			Code:  "provider_response_read_failed",
			Err:   err,
		}
	}
	if len(raw) > maxResponseBytes {
		return enrichment.EnrichmentResult{}, responseMalformed("DeepSeek response exceeded the bounded response size")
	}

	var completion chatCompletionResponse
	if err := json.Unmarshal(raw, &completion); err != nil {
		return enrichment.EnrichmentResult{}, responseMalformed("DeepSeek response envelope is not valid JSON")
	}
	if len(completion.Choices) == 0 {
		return enrichment.EnrichmentResult{}, responseMalformed("DeepSeek response contained no choices")
	}
	choice := completion.Choices[0]
	if choice.FinishReason == "length" {
		return enrichment.EnrichmentResult{}, responseMalformed("DeepSeek response was truncated")
	}
	content := strings.TrimSpace(choice.Message.Content)
	if content == "" {
		return enrichment.EnrichmentResult{}, responseMalformed("DeepSeek returned empty content")
	}

	result, err := enrichment.ParseEnrichmentResponseString(content, request)
	if err != nil {
		// Stage 2B remains the strict parser and candidate/prohibited-target
		// validation boundary. No provider response repair occurs here.
		return enrichment.EnrichmentResult{}, err
	}
	result.Provider = ProviderName
	result.Model = strings.TrimSpace(completion.Model)
	if result.Model == "" {
		result.Model = p.model
	}
	result.ResponseID = completion.ID
	return result, nil
}

func buildJSONInstructions() string {
	return enrichment.StrictInferenceInstructions + `

Return JSON, not markdown. The top-level object must contain exactly these keys: source_item_code, brand, category, description, unsupported_semantics.

Use exactly this Stage 2B proposal shape and these JSON member names:
- brand and category are objects with required action and confidence members. They may also use only target_id, target_code, canonical_name, evidence, and explanation.
- description is an object with required action and confidence members. Its canonical content member is value; do not invent any alternate member for the description text. It may also use only value, evidence, and explanation.
- unsupported_semantics is an array of objects with required semantic_type, key, value, and confidence members. It may also use evidence and explanation. These entries are informational evidence only.
- source_item_code is the exact source correlation value supplied in the request.

Evidence is optional. Omit an evidence member when no evidence is available. Whenever evidence is present on brand, category, description, or an unsupported_semantics item, it MUST be a JSON array of strings. For example, use "evidence":["first source fact","second source fact"] or, for one fact, "evidence":["single source fact"]. NEVER return a scalar string such as "evidence":"single source fact"; even one evidence item must be a one-element array.

Omit optional members that do not apply to the selected action; do not invent aliases. For KEEP_EXISTING and NO_MATCH, omit target_id, target_code, canonical_name, and description value. For MATCH_EXISTING brand/category, use only an exact target_id and/or target_code from the supplied canonical candidate list; never use a model-authored name as identity. For PROPOSE_NEW brand/category, provide canonical_name for review only and omit target_id and target_code. For PROPOSE_NEW description, provide a non-empty string in value. Description does not use MATCH_EXISTING.

Action conditions are strict: resolved structured brand requires KEEP_EXISTING; unresolved brand requires MATCH_EXISTING against a supplied exact candidate, PROPOSE_NEW, or NO_MATCH, and must not use KEEP_EXISTING. populated structured category requires KEEP_EXISTING; missing category requires MATCH_EXISTING against a supplied exact candidate, PROPOSE_NEW, or NO_MATCH, and must not use KEEP_EXISTING. existing description requires KEEP_EXISTING; missing description requires PROPOSE_NEW with value or NO_MATCH. Unsupported semantics are informational only. Product type remains immutable context and must never be proposed.`
}

func classifyHTTPStatus(status int) error {
	class := enrichment.ProviderErrorPermanent
	if status == http.StatusRequestTimeout || status == http.StatusTooManyRequests || status >= 500 {
		class = enrichment.ProviderErrorRetryable
	}
	return &enrichment.ProviderError{
		Class: class,
		Code:  fmt.Sprintf("http_%d", status),
		Err:   fmt.Errorf("DeepSeek API returned HTTP status %d", status),
	}
}

func classifyDeepSeekError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return ctx.Err()
		}
		return &enrichment.ProviderError{Class: enrichment.ProviderErrorRetryable, Code: "provider_timeout", Err: ctx.Err()}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return &enrichment.ProviderError{Class: enrichment.ProviderErrorRetryable, Code: "provider_timeout", Err: err}
	}
	var networkErr net.Error
	if errors.As(err, &networkErr) && networkErr.Timeout() {
		return &enrichment.ProviderError{Class: enrichment.ProviderErrorRetryable, Code: "provider_timeout", Err: err}
	}
	return &enrichment.ProviderError{Class: enrichment.ProviderErrorRetryable, Code: "provider_network_error", Err: err}
}

func responseMalformed(message string) error {
	return &enrichment.ResponseError{Class: enrichment.ResponseMalformed, Err: fmt.Errorf("%s", message)}
}
