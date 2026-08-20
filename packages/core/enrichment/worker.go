package enrichment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"time"
)

// EnrichmentWorker executes durable suggestions sequentially. Sequential
// processing is intentional for the first provider rollout: it bounds model
// concurrency and keeps shutdown behavior straightforward.
type EnrichmentWorker struct {
	store    EnrichmentExecutionStore
	provider ProductEnrichmentProvider
	config   EnrichmentExecutionConfig
	logger   *log.Logger
}

func NewEnrichmentWorker(store EnrichmentExecutionStore, provider ProductEnrichmentProvider, config EnrichmentExecutionConfig, logger *log.Logger) *EnrichmentWorker {
	if logger == nil {
		logger = log.Default()
	}
	return &EnrichmentWorker{store: store, provider: provider, config: config.normalized(), logger: logger}
}

// Start returns immediately and stops on context cancellation. It does not
// start when the provider or store is absent; SAP synchronization is allowed
// to operate without enrichment.
func (w *EnrichmentWorker) Start(ctx context.Context) {
	if w == nil || w.store == nil || w.provider == nil {
		return
	}
	go func() {
		_ = w.RunOnce(ctx)
		ticker := time.NewTicker(w.config.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := w.RunOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
					w.logger.Printf("product enrichment worker batch error: %v", err)
				}
			}
		}
	}()
}

// RunOnce is exported for deterministic tests and operational probes. A
// record-level failure is persisted and does not abort the remaining batch.
func (w *EnrichmentWorker) RunOnce(ctx context.Context) error {
	if w == nil || w.store == nil || w.provider == nil {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	now := w.config.Now()
	rows, err := w.store.ListDueEnrichmentSuggestions(ctx, w.config.BatchSize, now)
	if err != nil {
		return fmt.Errorf("list due enrichment suggestions: %w", err)
	}
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := w.processOne(ctx, row); err != nil && !errors.Is(err, context.Canceled) {
			w.logger.Printf("product enrichment suggestion %d failed to process: %v", row.ID, err)
		}
	}
	return nil
}

func (w *EnrichmentWorker) processOne(ctx context.Context, queued EnrichmentExecutionSuggestion) error {
	claimed, err := w.store.ClaimEnrichmentSuggestion(ctx, queued.OrganizationID, queued.ID)
	if err != nil {
		return err
	}

	snapshot, err := w.store.LoadSAPProductEnrichmentSnapshot(ctx, claimed.OrganizationID, claimed.SourceItemCode)
	if err != nil {
		return w.failOrRetry(ctx, claimed, "source_load_failed", err)
	}
	if snapshot.ProductID != claimed.ProductID || snapshot.OrganizationID != claimed.OrganizationID {
		return w.fail(ctx, claimed, "source_correlation_mismatch", fmt.Errorf("committed product identity changed"))
	}

	fingerprint, err := FingerprintSnapshot(snapshot)
	if err != nil {
		return w.fail(ctx, claimed, "source_fingerprint_failed", err)
	}
	if fingerprint != claimed.SourceDataFingerprint {
		decision := EvaluateEligibility(EnrichmentEligibilityInput{
			SourceSystem: snapshot.SourceSystem, OrganizationID: snapshot.OrganizationID, ProductID: snapshot.ProductID,
			SourceItemCode: snapshot.SourceItemCode, SourceItemName: snapshot.SourceItemName,
			ProductType: snapshot.ProductType, Brand: snapshot.Brand, Category: snapshot.Category, Description: snapshot.Description,
		})
		if !decision.Eligible {
			return w.fail(ctx, claimed, "stale_source_no_gap", fmt.Errorf("source changed and no current enrichment gap remains"))
		}
		return w.fail(ctx, claimed, "stale_source", fmt.Errorf("source fingerprint changed before execution"))
	}

	decision := EvaluateEligibility(EnrichmentEligibilityInput{
		SourceSystem: snapshot.SourceSystem, OrganizationID: snapshot.OrganizationID, ProductID: snapshot.ProductID,
		SourceItemCode: snapshot.SourceItemCode, SourceItemName: snapshot.SourceItemName,
		ProductType: snapshot.ProductType, Brand: snapshot.Brand, Category: snapshot.Category, Description: snapshot.Description,
	})
	if !decision.Eligible {
		return w.fail(ctx, claimed, "no_enrichment_gap", fmt.Errorf("current source is no longer eligible"))
	}

	brands, err := w.store.ListBrandCandidates(ctx, DefaultCandidateLimit)
	if err != nil {
		return w.failOrRetry(ctx, claimed, "brand_candidates_failed", err)
	}
	categories, err := w.store.ListCategoryCandidates(ctx, DefaultCandidateLimit)
	if err != nil {
		return w.failOrRetry(ctx, claimed, "category_candidates_failed", err)
	}
	request, err := NewEnrichmentRequest(snapshot, brands, categories)
	if err != nil {
		return w.fail(ctx, claimed, "request_contract_failed", err)
	}

	callCtx, cancel := context.WithTimeout(ctx, w.config.Timeout)
	result, err := w.provider.Enrich(callCtx, request)
	cancel()
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return w.handleProviderError(ctx, claimed, err)
	}
	if err := validateResultMetadata(result, request); err != nil {
		return w.fail(ctx, claimed, "provider_result_metadata_failed", err)
	}

	brand, err := json.Marshal(result.Proposals.Brand)
	if err != nil {
		return w.fail(ctx, claimed, "proposal_encode_failed", err)
	}
	category, err := json.Marshal(result.Proposals.Category)
	if err != nil {
		return w.fail(ctx, claimed, "proposal_encode_failed", err)
	}
	description, err := json.Marshal(result.Proposals.Description)
	if err != nil {
		return w.fail(ctx, claimed, "proposal_encode_failed", err)
	}
	unsupported, err := json.Marshal(result.Proposals.UnsupportedSemantics)
	if err != nil {
		return w.fail(ctx, claimed, "proposal_encode_failed", err)
	}
	return w.store.CompleteEnrichmentSuggestion(ctx, EnrichmentCompletion{
		OrganizationID: claimed.OrganizationID, ID: claimed.ID,
		ProposedBrand: brand, ProposedCategory: category, ProposedDescription: description, UnsupportedSemantics: unsupported,
		Provider: result.Provider, Model: result.Model, ModelVersion: result.ModelVersion,
	})
}

func validateResultMetadata(result EnrichmentResult, request EnrichmentRequest) error {
	if result.SourceItemCode != request.SourceItemCode {
		return fmt.Errorf("provider result source_item_code does not match request")
	}
	if result.Provider == "" || result.Model == "" {
		return fmt.Errorf("trusted provider metadata is incomplete")
	}
	return nil
}

func (w *EnrichmentWorker) handleProviderError(ctx context.Context, row EnrichmentExecutionSuggestion, err error) error {
	if ResponseErrorClassOf(err) != "" || ProviderErrorClassOf(err) == ProviderErrorPermanent {
		return w.fail(ctx, row, errorCode(err), err)
	}
	if ProviderErrorClassOf(err) == ProviderErrorRetryable || errors.Is(err, context.DeadlineExceeded) {
		return w.failOrRetry(ctx, row, errorCode(err), err)
	}
	return w.fail(ctx, row, errorCode(err), err)
}

func (w *EnrichmentWorker) failOrRetry(ctx context.Context, row EnrichmentExecutionSuggestion, code string, err error) error {
	if row.AttemptCount >= w.config.MaxAttempts {
		return w.fail(ctx, row, code, err)
	}
	delay := retryBackoff(row.AttemptCount)
	return w.store.MarkEnrichmentRetryable(ctx, EnrichmentRetry{OrganizationID: row.OrganizationID, ID: row.ID, NextAttemptAt: w.config.Now().Add(delay), ErrorCode: code})
}

func (w *EnrichmentWorker) fail(ctx context.Context, row EnrichmentExecutionSuggestion, code string, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return w.store.MarkEnrichmentFailed(ctx, EnrichmentFailure{OrganizationID: row.OrganizationID, ID: row.ID, ErrorCode: code})
}

func retryBackoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 6 {
		attempt = 6
	}
	return time.Duration(math.Pow(2, float64(attempt-1))) * time.Minute
}

func errorCode(err error) string {
	var providerErr *ProviderError
	if errors.As(err, &providerErr) && providerErr.Code != "" {
		return providerErr.Code
	}
	if class := ResponseErrorClassOf(err); class != "" {
		return string(class)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "provider_timeout"
	}
	return "execution_failed"
}
