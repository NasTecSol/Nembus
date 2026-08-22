// local-apply-service-diagnostic invokes the current tenant-local enrichment
// application service once so its returned error is visible before any HTTP
// handler maps it to a generic response. It must be run only through its
// companion PowerShell script.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	diagnosticDatabaseURLEnv = "NEMBUS_DIRECT_APPLY_DIAGNOSTIC_DATABASE_URL"
	diagnosticSuggestionID  = int32(7)
	diagnosticActorUsername = "e2e_enrichment_reviewer"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	databaseURL, err := localDiagnosticDatabaseURL()
	if err != nil {
		fmt.Printf("DIAGNOSTIC_CONFIGURATION=FAILED\nCONFIGURATION_ERROR=%s\n", sanitizeError(err))
		return
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Printf("DIAGNOSTIC_CONFIGURATION=FAILED\nCONFIGURATION_ERROR=%s\n", sanitizeError(err))
		return
	}
	defer pool.Close()

	// This mirrors the actual tenant middleware/handler composition: a
	// repository is built from the tenant pool, then passed to the real store
	// and real application service. There is intentionally no master fallback.
	queries := repository.New(pool)
	actor, err := queries.GetUserByUsername(ctx, diagnosticActorUsername)
	if err != nil {
		fmt.Printf("DIAGNOSTIC_ACTOR_CONTEXT=FAILED\nACTOR_CONTEXT_ERROR=%s\n", sanitizeError(err))
		return
	}
	if actor.ID <= 0 || actor.OrganizationID <= 0 || (actor.IsActive.Valid && !actor.IsActive.Bool) {
		fmt.Println("DIAGNOSTIC_ACTOR_CONTEXT=FAILED")
		fmt.Println("ACTOR_CONTEXT_ERROR=the disposable E2E reviewer is not an active user with positive tenant identifiers")
		return
	}

	store := repository.NewProductEnrichmentApplicationStore(queries, pool)
	service := enrichment.NewProductEnrichmentApplicationService(store)

	// This is the diagnostic's only application call. Its arguments match the
	// handler's normal call shape: authenticated user's organization, path
	// suggestion ID, and authenticated user ID.
	result, err := service.ApplyApprovedSuggestion(ctx, actor.OrganizationID, diagnosticSuggestionID, actor.ID)
	if err != nil {
		printApplyFailure(err)
		return
	}

	// A non-idempotent successful return means the real service inserted its
	// audit row and committed the same transaction; no separate diagnostic SQL
	// is issued.
	fmt.Println("DIRECT_APPLY_RESULT=PASS")
	fmt.Printf("DIRECT_APPLY_SUGGESTION_ID=%d\n", result.SuggestionID)
	fmt.Printf("DIRECT_APPLY_PRODUCT_ID=%d\n", result.ProductID)
	fmt.Printf("DIRECT_APPLY_STATUS=%s\n", result.Status)
	fmt.Printf("DIRECT_APPLY_ALREADY_APPLIED=%t\n", result.AlreadyApplied)
	fmt.Printf("DIRECT_APPLY_CHANGED_FIELDS=%s\n", strings.Join(result.ChangedFields, ","))
	if result.AppliedAt == nil {
		fmt.Println("DIRECT_APPLY_APPLIED_AT=")
	} else {
		fmt.Printf("DIRECT_APPLY_APPLIED_AT=%s\n", result.AppliedAt.UTC().Format(time.RFC3339Nano))
	}
	if result.AlreadyApplied {
		fmt.Println("DIRECT_APPLY_AUDIT_RESULT=NO_NEW_AUDIT_ALREADY_APPLIED")
	} else {
		fmt.Println("DIRECT_APPLY_AUDIT_RESULT=COMMITTED_BY_REAL_APPLICATION_SERVICE")
	}
}

func localDiagnosticDatabaseURL() (string, error) {
	rawURL := os.Getenv(diagnosticDatabaseURLEnv)
	if rawURL == "" {
		return "", fmt.Errorf("%s is required", diagnosticDatabaseURLEnv)
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse diagnostic database URL: %w", err)
	}
	if (u.Scheme != "postgres" && u.Scheme != "postgresql") ||
		u.Hostname() != "127.0.0.1" ||
		u.Port() != "5432" ||
		u.User == nil ||
		u.User.Username() != "nembus_e2e_user" ||
		u.EscapedPath() != "/nembus_e2e_tenant" ||
		u.Query().Get("sslmode") != "disable" {
		return "", fmt.Errorf("diagnostic database target must be 127.0.0.1:5432 / nembus_e2e_tenant as nembus_e2e_user with sslmode=disable")
	}
	return rawURL, nil
}

func printApplyFailure(err error) {
	layer, function, query := classifyApplyFailure(err)
	fmt.Println("DIRECT_APPLY_RESULT=FAILED")
	fmt.Printf("DIRECT_APPLY_ERROR=%s\n", sanitizeError(err))
	fmt.Printf("DIRECT_APPLY_ERROR_TYPE=%T\n", err)

	pgErrorIndex := 0
	for index, current := 0, err; current != nil; index, current = index+1, errors.Unwrap(current) {
		fmt.Printf("ERROR_CHAIN_%d=%s\n", index, sanitizeError(current))
		if pgErr, ok := current.(*pgconn.PgError); ok {
			pgErrorIndex++
			printPostgresError(pgErr, pgErrorIndex)
		}
	}
	fmt.Printf("APPLY_FAILURE_LAYER=%s\n", layer)
	fmt.Printf("FAILING_FUNCTION=%s\n", function)
	fmt.Printf("FAILING_QUERY=%s\n", query)
}

func printPostgresError(pgErr *pgconn.PgError, index int) {
	prefix := "PG_"
	if index > 1 {
		prefix = fmt.Sprintf("PG_%d_", index)
	}
	fmt.Printf("%sSQLSTATE=%s\n", prefix, sanitizeText(pgErr.Code))
	fmt.Printf("%sMESSAGE=%s\n", prefix, sanitizeText(pgErr.Message))
	fmt.Printf("%sDETAIL=%s\n", prefix, sanitizeText(pgErr.Detail))
	fmt.Printf("%sCONSTRAINT=%s\n", prefix, sanitizeText(pgErr.ConstraintName))
	fmt.Printf("%sTABLE=%s\n", prefix, sanitizeText(pgErr.TableName))
	fmt.Printf("%sCOLUMN=%s\n", prefix, sanitizeText(pgErr.ColumnName))
}

func classifyApplyFailure(err error) (layer, function, query string) {
	function, query = "UNKNOWN", "UNKNOWN"

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		function, query = sourceLocation(err)
		switch pgErr.Code {
		case "23503":
			return "FOREIGN_KEY", function, query
		case "23502":
			return "NOT_NULL", function, query
		}
	}

	message := strings.ToLower(sanitizeError(err))
	function, query = sourceLocation(err)
	if strings.Contains(message, "cannot scan") || strings.Contains(message, "scan error") {
		return "SQL_SCAN", function, query
	}

	var applicationErr *enrichment.ApplicationError
	if !errors.As(err, &applicationErr) {
		return "UNKNOWN", function, query
	}

	switch applicationErr.Message {
	case "lock enrichment suggestion":
		return "SUGGESTION_LOCK", function, query
	case "lock current enrichment product":
		return "PRODUCT_LOCK", function, query
	case "enrichment suggestion source is stale":
		return "STALE_FINGERPRINT_RECHECK", function, query
	case "revalidate approved brand target":
		return "BRAND_REVALIDATION", function, query
	case "revalidate approved category target":
		return "CATEGORY_REVALIDATION", function, query
	case "apply narrow product enrichment fields":
		return "PRODUCT_UPDATE", function, query
	case "mark enrichment suggestion applied":
		return "MARK_APPLIED", function, query
	case "write enrichment application audit":
		return "AUDIT_INSERT", function, query
	case "commit enrichment application", "commit already-applied result":
		return "TRANSACTION_COMMIT", function, query
	case "begin application transaction":
		return "OTHER_TRANSACTION_BEGIN", function, query
	case "fingerprint current enrichment product":
		return "OTHER_FINGERPRINT", function, query
	case "current product source context is no longer valid", "current product source identity changed", "current product source context changed":
		return "OTHER_SOURCE_CONTEXT_REVALIDATION", function, query
	}
	return "OTHER_APPLICATION_VALIDATION", function, query
}

func sourceLocation(err error) (function, query string) {
	function, query = "UNKNOWN", "UNKNOWN"
	var applicationErr *enrichment.ApplicationError
	if !errors.As(err, &applicationErr) {
		return function, query
	}
	switch applicationErr.Message {
	case "begin application transaction":
		return "BeginProductEnrichmentApplication", "UNKNOWN"
	case "lock enrichment suggestion":
		return "LockProductEnrichmentSuggestion", "LockProductEnrichmentSuggestionForApplication"
	case "lock current enrichment product":
		return "LockProductEnrichmentSnapshot", "LockProductForEnrichmentApplication"
	case "revalidate approved brand target":
		return "ResolveBrandApplicationTarget", "UNKNOWN"
	case "revalidate approved category target":
		return "ResolveCategoryApplicationTarget", "UNKNOWN"
	case "apply narrow product enrichment fields":
		return "ApplyProductEnrichmentFields", "ApplyProductEnrichmentFields"
	case "mark enrichment suggestion applied":
		return "MarkProductEnrichmentSuggestionApplied", "MarkProductEnrichmentSuggestionApplied"
	case "write enrichment application audit":
		return "InsertProductEnrichmentApplicationAudit", "InsertProductEnrichmentReviewAudit"
	case "commit enrichment application", "commit already-applied result":
		return "Commit", "UNKNOWN"
	case "fingerprint current enrichment product", "enrichment suggestion source is stale":
		return "FingerprintSnapshot", "UNKNOWN"
	case "current product source context is no longer valid", "current product source identity changed", "current product source context changed":
		return "validateApplicationSource", "UNKNOWN"
	}
	return function, query
}

var (
	postgresURLPattern = regexp.MustCompile(`(?i)postgres(?:ql)?://[^\s@]+@`)
	passwordPattern    = regexp.MustCompile(`(?i)(password|pwd)=([^\s&;]+)`)
)

func sanitizeError(err error) string {
	if err == nil {
		return ""
	}
	return sanitizeText(err.Error())
}

func sanitizeText(value string) string {
	value = postgresURLPattern.ReplaceAllString(value, "postgres://***@")
	value = passwordPattern.ReplaceAllString(value, "${1}=***")
	return strings.Join(strings.Fields(value), " ")
}
