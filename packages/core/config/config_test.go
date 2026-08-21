package config

import "testing"

func TestValidateEnrichmentConfigProviderSpecificSecrets(t *testing.T) {
	base := Config{
		EnrichmentEnabled:        true,
		EnrichmentProvider:       "deepseek",
		DeepSeekEnrichmentModel:  "deepseek-v4-flash",
		OpenAIEnrichmentTimeout:  1,
		EnrichmentWorkerInterval: 1,
		EnrichmentBatchSize:      1,
		EnrichmentMaxRetries:     1,
	}
	if err := base.ValidateEnrichmentConfig(); err == nil {
		t.Fatal("deepseek key should be required")
	}
	base.DeepSeekAPIKey = "test-key"
	if err := base.ValidateEnrichmentConfig(); err != nil {
		t.Fatalf("configured deepseek should validate: %v", err)
	}

	base.EnrichmentEnabled = false
	base.DeepSeekAPIKey = ""
	if err := base.ValidateEnrichmentConfig(); err != nil {
		t.Fatalf("disabled enrichment should not require a key: %v", err)
	}
}

func TestValidateEnrichmentConfigPreservesOpenAIAndRejectsUnknown(t *testing.T) {
	openAI := Config{
		EnrichmentEnabled:        true,
		EnrichmentProvider:       "openai",
		OpenAIAPIKey:             "test-key",
		OpenAIEnrichmentModel:    "test-model",
		OpenAIEnrichmentTimeout:  1,
		EnrichmentWorkerInterval: 1,
		EnrichmentBatchSize:      1,
		EnrichmentMaxRetries:     1,
	}
	if err := openAI.ValidateEnrichmentConfig(); err != nil {
		t.Fatalf("existing OpenAI configuration changed: %v", err)
	}

	openAI.EnrichmentProvider = "unknown"
	if err := openAI.ValidateEnrichmentConfig(); err == nil {
		t.Fatal("unknown provider should be rejected")
	}
}
