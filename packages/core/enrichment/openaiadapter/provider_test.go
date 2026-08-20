package openaiadapter

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

func TestProviderUsesConfiguredModelStrictSchemaAndSafeInput(t *testing.T) {
	client := &fakeResponseClient{response: responseWithText(`{"source_item_code":"INV00006","brand":{"action":"MATCH_EXISTING","target_id":10,"target_code":"PANTENE","canonical_name":"","confidence":0.9,"evidence":[],"explanation":null},"category":{"action":"NO_MATCH","target_id":null,"target_code":null,"canonical_name":null,"confidence":0.2,"evidence":[],"explanation":null},"description":{"action":"PROPOSE_NEW","value":"شامبو بانتين","confidence":0.8,"evidence":[],"explanation":null},"unsupported_semantics":[]}`)}
	provider := newWithClient(client, "future-model", time.Second)
	request := adapterRequest(t)
	result, err := provider.Enrich(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Provider != ProviderName || result.Model != "future-model" || client.params.Model != "future-model" {
		t.Fatalf("configured/trusted model metadata was not preserved: result=%+v", result)
	}
	payload, err := json.Marshal(client.params)
	if err != nil {
		t.Fatal(err)
	}
	serialized := string(payload)
	if !strings.Contains(client.params.Instructions.Value, "Structured SAP values are authoritative") || !strings.Contains(client.params.Input.OfString.Value, "PANTENE") {
		t.Fatalf("instructions or candidates missing from request: %s", serialized)
	}
	inputPayload := client.params.Input.OfString.Value
	if strings.Contains(inputPayload, "conversion_factor") || strings.Contains(inputPayload, "inventory") || strings.Contains(inputPayload, "price") || strings.Contains(inputPayload, "supplier") {
		t.Fatalf("operational/UoM conversion data leaked into provider input: %s", inputPayload)
	}
	format := client.params.Text.Format
	if format.OfJSONSchema == nil || format.OfJSONSchema.Strict.Value != true {
		t.Fatalf("strict structured output schema was not requested")
	}
	if format.OfJSONSchema.Schema["additionalProperties"] != false {
		t.Fatalf("root schema is not closed")
	}
}

func TestProviderClassifiesTypedHTTPFailures(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		class  enrichment.ProviderErrorClass
	}{
		{name: "rate limit", status: 429, class: enrichment.ProviderErrorRetryable},
		{name: "server", status: 503, class: enrichment.ProviderErrorRetryable},
		{name: "auth", status: 401, class: enrichment.ProviderErrorPermanent},
		{name: "bad request", status: 400, class: enrichment.ProviderErrorPermanent},
	} {
		t.Run(test.name, func(t *testing.T) {
			provider := newWithClient(&fakeResponseClient{err: &responses.Error{StatusCode: test.status}}, "model", time.Second)
			_, err := provider.Enrich(context.Background(), adapterRequest(t))
			if got := enrichment.ProviderErrorClassOf(err); got != test.class {
				t.Fatalf("class=%s want %s", got, test.class)
			}
		})
	}
}

func TestProviderTimeoutAndCancellation(t *testing.T) {
	provider := newWithClient(&fakeResponseClient{err: context.DeadlineExceeded}, "model", time.Second)
	_, err := provider.Enrich(context.Background(), adapterRequest(t))
	if got := enrichment.ProviderErrorClassOf(err); got != enrichment.ProviderErrorRetryable {
		t.Fatalf("timeout class=%s", got)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = provider.Enrich(ctx, adapterRequest(t))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation should be preserved, err=%v", err)
	}
}

func TestProviderRoutesMalformedAndCandidateMismatchThroughStrictParser(t *testing.T) {
	for _, raw := range []string{
		`{"source_item_code":"INV00006"}`,
		`{"source_item_code":"INV00006","brand":{"action":"MATCH_EXISTING","target_id":999,"target_code":null,"canonical_name":null,"confidence":0.9,"evidence":[],"explanation":null},"category":{"action":"NO_MATCH","target_id":null,"target_code":null,"canonical_name":null,"confidence":0.2,"evidence":[],"explanation":null},"description":{"action":"NO_MATCH","value":null,"confidence":0.2,"evidence":[],"explanation":null},"unsupported_semantics":[]}`,
	} {
		provider := newWithClient(&fakeResponseClient{response: responseWithText(raw)}, "model", time.Second)
		_, err := provider.Enrich(context.Background(), adapterRequest(t))
		if enrichment.ResponseErrorClassOf(err) == "" {
			t.Fatalf("strict parser error was not preserved: %v", err)
		}
	}
}

type fakeResponseClient struct {
	params   responses.ResponseNewParams
	response *responses.Response
	err      error
}

func (f *fakeResponseClient) New(_ context.Context, params responses.ResponseNewParams, _ ...option.RequestOption) (*responses.Response, error) {
	f.params = params
	return f.response, f.err
}

func responseWithText(text string) *responses.Response {
	return &responses.Response{ID: "resp_test", Model: "future-model", Output: []responses.ResponseOutputItemUnion{{Content: []responses.ResponseOutputMessageContentUnion{{Type: "output_text", Text: text}}}}}
}

func adapterRequest(t *testing.T) enrichment.EnrichmentRequest {
	t.Helper()
	snapshot := enrichment.EnrichmentSourceSnapshot{OrganizationID: 2, ProductID: 3, SourceSystem: enrichment.SourceSystemSAP, SourceItemCode: "INV00006", SourceItemName: "شامبو بانتين 24*400 مل", ProductType: enrichment.ProductTypeStandard, UOM: enrichment.UOMContext{Conversions: []enrichment.UOMConversionContext{{ConversionFactor: "24"}}}}
	request, err := enrichment.NewEnrichmentRequest(snapshot, []enrichment.BrandCandidate{{ID: 10, Code: "PANTENE", Name: "Pantene"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return request
}
