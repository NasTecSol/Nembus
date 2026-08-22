// local-stage2b-contract-diagnostic is a deliberately narrow, read-only
// operator diagnostic. It replays one already-failed disposable-E2E
// suggestion through the current DeepSeek adapter and Stage 2B parser without
// starting the worker or changing the suggestion lifecycle.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/config"
	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/NasTecSol/nembus-core/enrichment/deepseekadapter"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	diagnosticDBHost = "127.0.0.1"
	diagnosticDBPort = 5432
	diagnosticDBName = "nembus_e2e_tenant"
	diagnosticDBUser = "nembus_e2e_user"

	dbPasswordEnvironment = "NEMBUS_STAGE2B_DB_PASSWORD"
	apiKeyEnvironment     = "NEMBUS_STAGE2B_DEEPSEEK_API_KEY"
)

// The direct query exists only because there is no generated repository API
// for this precise operational selection. Its product join proves that the
// source identity belongs to the selected disposable tenant-local product.
const selectFailedE2ESuggestion = `
SELECT s.id, s.organization_id, s.product_id, s.source_item_code
FROM product_enrichment_suggestions AS s
JOIN products AS p
  ON p.id = s.product_id
 AND p.organization_id = s.organization_id
WHERE s.source_item_code LIKE 'E2E-PANTENE-%'
  AND s.status = 'failed'
  AND s.last_error_code = 'contract_violation'
  AND p.sku = s.source_item_code
ORDER BY s.updated_at DESC NULLS LAST, s.id DESC
LIMIT 1`

type selectedSuggestion struct {
	ID             int32
	OrganizationID int32
	ProductID      int32
	SourceItemCode string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	dbPassword := requiredEnvironment(dbPasswordEnvironment)
	apiKey := requiredEnvironment(apiKeyEnvironment)
	pool, err := openReadOnlyE2EPool(ctx, dbPassword)
	if err != nil {
		fatal("DATABASE_CONNECTION_FAILED", err, dbPassword, apiKey)
	}
	defer pool.Close()

	selected, err := findFailedE2ESuggestion(ctx, pool)
	if errors.Is(err, pgx.ErrNoRows) {
		fatalMessage("NO_MATCHING_FAILED_E2E_SUGGESTION")
	}
	if err != nil {
		fatal("SUGGESTION_SELECTION_FAILED", err, dbPassword, apiKey)
	}

	queries := repository.New(pool)
	stored, err := queries.GetProductEnrichmentSuggestionByID(ctx, repository.GetProductEnrichmentSuggestionByIDParams{
		OrganizationID: selected.OrganizationID,
		ID:             selected.ID,
	})
	if err != nil {
		fatal("SUGGESTION_RELOAD_FAILED", err, dbPassword, apiKey)
	}
	if err := validateSelectedSuggestion(selected, stored); err != nil {
		fatal("SUGGESTION_IDENTITY_INVALID", err, dbPassword, apiKey)
	}
	printSelection(selected)

	store := repository.NewProductEnrichmentStore(queries)
	snapshot, err := store.LoadSAPProductEnrichmentSnapshot(ctx, selected.OrganizationID, selected.SourceItemCode)
	if err != nil {
		fatal("SOURCE_SNAPSHOT_FAILED", err, dbPassword, apiKey)
	}
	if snapshot.OrganizationID != selected.OrganizationID || snapshot.ProductID != selected.ProductID || strings.TrimSpace(snapshot.SourceItemCode) != selected.SourceItemCode {
		fatalMessage("SOURCE_CORRELATION_MISMATCH")
	}
	if err := validateExpectedE2EContract(snapshot); err != nil {
		fatal("E2E_CONTRACT_EXPECTATION_FAILED", err, dbPassword, apiKey)
	}
	brands, err := store.ListBrandCandidates(ctx, enrichment.DefaultCandidateLimit)
	if err != nil {
		fatal("BRAND_CANDIDATES_FAILED", err, dbPassword, apiKey)
	}
	categories, err := store.ListCategoryCandidates(ctx, enrichment.DefaultCandidateLimit)
	if err != nil {
		fatal("CATEGORY_CANDIDATES_FAILED", err, dbPassword, apiKey)
	}
	request, err := enrichment.NewEnrichmentRequest(snapshot, brands, categories)
	if err != nil {
		fatal("PROVIDER_REQUEST_INVALID", err, dbPassword, apiKey)
	}
	if err := request.Validate(); err != nil {
		fatal("PROVIDER_REQUEST_INVALID", err, dbPassword, apiKey)
	}
	fmt.Println("E2E_CONTRACT_EXPECTATION=PASS")
	fmt.Println("PROVIDER_REQUEST_RECONSTRUCTED=PASS")

	// These are the current cloud-server configuration semantics: the config
	// supplies the DeepSeek base URL/model and the shared enrichment timeout;
	// the API key is deliberately the runtime prompt, never config fallback.
	cfg := config.LoadConfig("diagnostic")
	provider, err := deepseekadapter.New(apiKey, cfg.DeepSeekBaseURL, cfg.DeepSeekEnrichmentModel, cfg.OpenAIEnrichmentTimeout)
	if err != nil {
		fatal("DEEPSEEK_PROVIDER_CONFIGURATION_FAILED", err, dbPassword, apiKey)
	}

	// Provider.Enrich has one HTTP Do call and no retry loop. It is also the
	// real adapter boundary that extracts the response envelope and invokes
	// enrichment.ParseEnrichmentResponse; no local prompt or parser is copied.
	_, err = provider.Enrich(ctx, request)
	if code := providerErrorCode(err); code == "request_encoding_failed" || code == "request_configuration_failed" || code == "provider_not_configured" {
		fatal("DEEPSEEK_CALL_NOT_MADE", err, dbPassword, apiKey)
	}
	fmt.Println("DEEPSEEK_CALL_COUNT=1")
	if err == nil {
		fmt.Println("STAGE2B_PARSE_RESULT=PASS")
		fmt.Println("DATABASE_MUTATION=false")
		return
	}

	class := enrichment.ResponseErrorClassOf(err)
	if class == "" {
		fmt.Println("STAGE2B_PARSE_RESULT=FAILED")
		fmt.Printf("STAGE2B_ERROR_CLASS=%s\n", providerErrorClass(err))
		fmt.Printf("STAGE2B_REJECTION=%s\n", sanitize(err.Error(), dbPassword, apiKey))
		fmt.Println("DATABASE_MUTATION=false")
		return
	}

	message := sanitize(err.Error(), dbPassword, apiKey)
	fmt.Println("STAGE2B_PARSE_RESULT=FAILED")
	fmt.Printf("STAGE2B_ERROR_CLASS=%s\n", class)
	fmt.Printf("STAGE2B_REJECTION=%s\n", message)
	if field := rejectedField(message); field != "" {
		fmt.Printf("STAGE2B_REJECTED_FIELD=%s\n", field)
	}
	fmt.Println("DATABASE_MUTATION=false")
}

func openReadOnlyE2EPool(ctx context.Context, password string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable", diagnosticDBHost, diagnosticDBPort, diagnosticDBName, diagnosticDBUser))
	if err != nil {
		return nil, err
	}
	poolConfig.ConnConfig.Password = password
	poolConfig.ConnConfig.RuntimeParams["default_transaction_read_only"] = "on"
	poolConfig.MaxConns = 1
	return pgxpool.NewWithConfig(ctx, poolConfig)
}

func findFailedE2ESuggestion(ctx context.Context, pool *pgxpool.Pool) (selectedSuggestion, error) {
	var selected selectedSuggestion
	err := pool.QueryRow(ctx, selectFailedE2ESuggestion).Scan(
		&selected.ID,
		&selected.OrganizationID,
		&selected.ProductID,
		&selected.SourceItemCode,
	)
	return selected, err
}

func validateSelectedSuggestion(selected selectedSuggestion, stored repository.ProductEnrichmentSuggestion) error {
	if selected.ID <= 0 || selected.OrganizationID <= 0 || selected.ProductID <= 0 {
		return fmt.Errorf("selected identifiers must be positive")
	}
	if !strings.HasPrefix(selected.SourceItemCode, "E2E-PANTENE-") {
		return fmt.Errorf("selected source_item_code is outside the E2E fixture")
	}
	if stored.ID != selected.ID || stored.OrganizationID != selected.OrganizationID || stored.ProductID != selected.ProductID || stored.SourceItemCode != selected.SourceItemCode {
		return fmt.Errorf("stored suggestion identity does not match the selected row")
	}
	if stored.Status != string(enrichment.SuggestionStatusFailed) || !stored.LastErrorCode.Valid || stored.LastErrorCode.String != string(enrichment.ResponseContractViolation) {
		return fmt.Errorf("stored suggestion no longer has the required failed contract_violation state")
	}
	if stored.ContractVersion != enrichment.EnrichmentContractVersion {
		return fmt.Errorf("stored suggestion contract version does not match the current contract")
	}
	return nil
}

func validateExpectedE2EContract(snapshot enrichment.EnrichmentSourceSnapshot) error {
	if snapshot.ProductType != enrichment.ProductTypeStandard {
		return fmt.Errorf("E2E fixture product_type must be standard")
	}
	if snapshot.Brand != nil && snapshot.Brand.Resolved() {
		return fmt.Errorf("E2E fixture must have unresolved structured brand")
	}
	if snapshot.Category == nil || !snapshot.Category.Resolved() {
		return fmt.Errorf("E2E fixture must retain a populated structured category")
	}
	if strings.TrimSpace(snapshot.Description) != "" {
		return fmt.Errorf("E2E fixture description must be missing")
	}
	wantGaps := []enrichment.EnrichmentGap{enrichment.GapMissingBrand, enrichment.GapMissingDescription}
	gaps := enrichment.GapsForSnapshot(snapshot)
	if len(gaps) != len(wantGaps) {
		return fmt.Errorf("E2E fixture has unexpected enrichment gaps")
	}
	for i := range wantGaps {
		if gaps[i] != wantGaps[i] {
			return fmt.Errorf("E2E fixture has unexpected enrichment gap %q", gaps[i])
		}
	}
	return nil
}

func printSelection(selected selectedSuggestion) {
	fmt.Printf("SELECTED_SUGGESTION_ID=%d\n", selected.ID)
	fmt.Printf("SELECTED_SOURCE_ITEM_CODE=%s\n", selected.SourceItemCode)
}

func requiredEnvironment(name string) string {
	value := os.Getenv(name)
	if strings.TrimSpace(value) == "" {
		fatalMessage("REQUIRED_RUNTIME_SECRET_MISSING=" + name)
	}
	return value
}

func providerErrorClass(err error) string {
	if code := providerErrorCode(err); code != "" {
		return code
	}
	return "provider_or_transport_error"
}

func providerErrorCode(err error) string {
	var providerErr *enrichment.ProviderError
	if errors.As(err, &providerErr) {
		return providerErr.Code
	}
	return ""
}

func sanitize(value string, secrets ...string) string {
	for _, secret := range secrets {
		if secret != "" {
			value = strings.ReplaceAll(value, secret, "[REDACTED]")
		}
	}
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 800 {
		return value[:800] + "…"
	}
	return value
}

var (
	unknownFieldPattern = regexp.MustCompile(`unknown field "([^"]+)"`)
	pathPattern         = regexp.MustCompile(`(?:^|: )((?:brand|category|description|unsupported_semantics)(?:\[[0-9]+\])?(?:\.[a-z_]+)?)`)
)

func rejectedField(message string) string {
	if match := unknownFieldPattern.FindStringSubmatch(message); len(match) == 2 {
		return match[1]
	}
	if match := pathPattern.FindStringSubmatch(message); len(match) == 2 {
		return match[1]
	}
	return ""
}

func fatal(label string, err error, secrets ...string) {
	fmt.Printf("%s=%s\n", label, sanitize(err.Error(), secrets...))
	fmt.Println("DATABASE_MUTATION=false")
	os.Exit(1)
}

func fatalMessage(message string) {
	fmt.Println(message)
	fmt.Println("DATABASE_MUTATION=false")
	os.Exit(1)
}
