// local-category-store-diagnostic is a deliberately narrow, local-only
// diagnostic for ProductEnrichmentStore.ListCategoryCandidates. It is not part
// of an application package and must be run only through its companion script.
package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

const diagnosticDatabaseURLEnv = "NEMBUS_CATEGORY_DIAGNOSTIC_DATABASE_URL"

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	databaseURL, err := localDiagnosticDatabaseURL()
	if err != nil {
		fmt.Printf("DIAGNOSTIC_CONFIGURATION=FAILED\nCONFIGURATION_ERROR=%s\n", sanitizeError(err))
		return
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Printf("SQLC_GET_CATEGORY_HIERARCHY=FAILED\nSQLC_ERROR=%s\nCATEGORY_FAILURE_LAYER=CONNECTION_CONFIGURATION\n", sanitizeError(err))
		return
	}
	defer pool.Close()

	queries := repository.New(pool)
	rows, err := queries.GetCategoryHierarchy(ctx, true)
	if err != nil {
		fmt.Printf("SQLC_GET_CATEGORY_HIERARCHY=FAILED\nSQLC_ERROR=%s\n", sanitizeError(err))
		printCategoryRowTypes()
		fmt.Println("CATEGORY_FAILURE_LAYER=SQLC_SCAN")
		return
	}
	fmt.Printf("SQLC_GET_CATEGORY_HIERARCHY=PASS\nSQLC_CATEGORY_ROWS=%d\n", len(rows))

	store := repository.NewProductEnrichmentStore(queries)
	candidates, err := store.ListCategoryCandidates(ctx, enrichment.DefaultCandidateLimit)
	if err != nil {
		fmt.Printf("STORE_LIST_CATEGORY_CANDIDATES=FAILED\nSTORE_ERROR=%s\nCATEGORY_FAILURE_LAYER=STORE_QUERY\n", sanitizeError(err))
		return
	}
	fmt.Printf("STORE_LIST_CATEGORY_CANDIDATES=PASS\nSTORE_CATEGORY_CANDIDATES=%d\nCATEGORY_FAILURE_LAYER=NONE\n", len(candidates))
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
	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return "", fmt.Errorf("diagnostic database scheme must be postgres or postgresql")
	}
	if u.Hostname() != "127.0.0.1" || u.Port() != "5432" || u.User == nil || u.User.Username() != "nembus_e2e_user" || u.EscapedPath() != "/nembus_e2e_tenant" || u.Query().Get("sslmode") != "disable" {
		return "", fmt.Errorf("diagnostic database target must be 127.0.0.1:5432 / nembus_e2e_tenant as nembus_e2e_user with sslmode=disable")
	}
	return rawURL, nil
}

func printCategoryRowTypes() {
	fmt.Println("SQLC_CATEGORY_ROW_TYPES=id=int32,parent_category_id=pgtype.Int4,name=string,code=string,description=pgtype.Text,category_level=pgtype.Int4,is_active=pgtype.Bool,metadata=json.RawMessage,full_path=string")
}

var (
	postgresURLPattern = regexp.MustCompile(`(?i)postgres(?:ql)?://[^\s@]+@`)
	passwordPattern    = regexp.MustCompile(`(?i)(password=)[^\s&;]+`)
)

func sanitizeError(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	message = postgresURLPattern.ReplaceAllString(message, "postgres://***@")
	message = passwordPattern.ReplaceAllString(message, "${1}***")
	return strings.TrimSpace(message)
}
