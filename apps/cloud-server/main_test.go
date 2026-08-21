package main

import (
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/config"
	"github.com/NasTecSol/nembus-core/enrichment/deepseekadapter"
	"github.com/NasTecSol/nembus-core/enrichment/openaiadapter"
)

func TestNewProductEnrichmentProviderSelectsConfiguredAdapter(t *testing.T) {
	deepSeekConfig := &config.Config{
		EnrichmentProvider:      "deepseek",
		DeepSeekAPIKey:          "test-key",
		DeepSeekBaseURL:         "https://api.deepseek.com",
		DeepSeekEnrichmentModel: "test-model",
		OpenAIEnrichmentTimeout: time.Second,
	}
	provider, err := newProductEnrichmentProvider(deepSeekConfig)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := provider.(*deepseekadapter.Provider); !ok {
		t.Fatalf("provider type=%T, want DeepSeek adapter", provider)
	}

	openAIConfig := &config.Config{
		EnrichmentProvider:      "openai",
		OpenAIAPIKey:            "test-key",
		OpenAIEnrichmentModel:   "test-model",
		OpenAIEnrichmentTimeout: time.Second,
	}
	provider, err = newProductEnrichmentProvider(openAIConfig)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := provider.(*openaiadapter.Provider); !ok {
		t.Fatalf("provider type=%T, want OpenAI adapter", provider)
	}
}

func TestNewProductEnrichmentProviderRejectsUnknownProvider(t *testing.T) {
	_, err := newProductEnrichmentProvider(&config.Config{EnrichmentProvider: "unknown"})
	if err == nil {
		t.Fatal("unknown provider should fail safely")
	}
}
