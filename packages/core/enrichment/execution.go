package enrichment

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ProviderErrorClass is the provider-neutral execution classification used by
// the durable worker. Provider adapters translate their own error types into
// these values.
type ProviderErrorClass string

const (
	ProviderErrorRetryable ProviderErrorClass = "retryable_provider_error"
	ProviderErrorPermanent ProviderErrorClass = "permanent_provider_error"
)

type ProviderError struct {
	Class ProviderErrorClass
	Code  string
	Err   error
}

func (e *ProviderError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return string(e.Class)
	}
	return fmt.Sprintf("%s: %v", e.Class, e.Err)
}

func (e *ProviderError) Unwrap() error { return e.Err }

func ProviderErrorClassOf(err error) ProviderErrorClass {
	var providerErr *ProviderError
	if errors.As(err, &providerErr) {
		return providerErr.Class
	}
	return ""
}

// EnrichmentExecutionSuggestion is the minimum durable row state needed by
// the worker. It deliberately does not expose generated SQLC types.
type EnrichmentExecutionSuggestion struct {
	ID                    int32
	OrganizationID        int32
	ProductID             int32
	SourceItemCode        string
	SourceDataFingerprint string
	Status                SuggestionStatus
	AttemptCount          int
}

type EnrichmentCompletion struct {
	OrganizationID       int32
	ID                   int32
	ProposedBrand        []byte
	ProposedCategory     []byte
	ProposedDescription  []byte
	UnsupportedSemantics []byte
	Provider             string
	Model                string
	ModelVersion         string
}

type EnrichmentRetry struct {
	OrganizationID int32
	ID             int32
	NextAttemptAt  time.Time
	ErrorCode      string
}

type EnrichmentFailure struct {
	OrganizationID int32
	ID             int32
	ErrorCode      string
}

// EnrichmentExecutionStore is the narrow persistence boundary for Stage 2C.
// Every mutating method carries organization ID so tenant isolation remains a
// SQL concern even though the worker scans the master queue.
type EnrichmentExecutionStore interface {
	ListDueEnrichmentSuggestions(context.Context, int, time.Time) ([]EnrichmentExecutionSuggestion, error)
	ClaimEnrichmentSuggestion(context.Context, int32, int32) (EnrichmentExecutionSuggestion, error)
	CompleteEnrichmentSuggestion(context.Context, EnrichmentCompletion) error
	MarkEnrichmentRetryable(context.Context, EnrichmentRetry) error
	MarkEnrichmentFailed(context.Context, EnrichmentFailure) error

	LoadSAPProductEnrichmentSnapshot(context.Context, int32, string) (EnrichmentSourceSnapshot, error)
	ListBrandCandidates(context.Context, int) ([]BrandCandidate, error)
	ListCategoryCandidates(context.Context, int) ([]CategoryCandidate, error)
}

type EnrichmentExecutionConfig struct {
	Interval    time.Duration
	Timeout     time.Duration
	BatchSize   int
	MaxAttempts int
	Now         func() time.Time
}

func (c EnrichmentExecutionConfig) normalized() EnrichmentExecutionConfig {
	if c.Interval <= 0 {
		c.Interval = 30 * time.Second
	}
	if c.Timeout <= 0 {
		c.Timeout = 45 * time.Second
	}
	if c.BatchSize <= 0 {
		c.BatchSize = 5
	}
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 3
	}
	if c.Now == nil {
		c.Now = time.Now
	}
	return c
}
