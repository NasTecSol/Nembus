package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Parse command line flags
	includeMaster := flag.Bool("master", true, "Also run migrations on the master database")
	declarativeMode := flag.Bool("declarative", false, "Use declarative schema sync (atlas schema apply) instead of versioned migrations")
	schemaDir := flag.String("schema", "", "Directory containing Atlas SQL schema files (for declarative mode; default: auto-detected packages/core/db/schema)")
	devDBURL := flag.String("dev-url", "", "Dev database URL for Atlas schema calculation (default: ATLAS_DEV_URL env var or docker://postgres/16/dev)")
	migrationsDir := flag.String("dir", "", "Directory containing Atlas migration files (default: auto-detected packages/core/db/migrations)")
	baselineVer := flag.String("baseline", "", "Baseline migration version. Leave empty (default) to AUTO-DETECT per database: empty DBs and DBs already tracked by Atlas (atlas_schema_revisions) get no baseline; DBs that have schema but no revisions table get the first migration version. Set it explicitly to force the same baseline for every database (e.g. -baseline 20260813124500) — a baseline that does not exist in the migration directory is a hard error.")
	statusOnly := flag.Bool("status", false, "Show status / dry-run without applying")
	hostOverride := flag.String("host-override", "postgres=localhost", "Rewrite the host in database connection strings before connecting (from=to). Set to empty to disable. Use when databases are Docker services (e.g. host=postgres) that only resolve inside the compose network while this runner executes on the host.")
	flag.Parse()

	// Get current working directory for debugging
	cwd, _ := os.Getwd()
	log.Printf("Current working directory: %s\n", cwd)

	// Load environment variables
	envPaths := []string{".env.dev", ".env", "../.env.dev", "../.env", "../../.env.dev", "../../.env"}
	var envLoaded bool
	var loadedPath string
	for _, envPath := range envPaths {
		absPath, _ := filepath.Abs(envPath)
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err == nil {
				envLoaded = true
				loadedPath = absPath
				log.Printf("✓ Loaded env from: %s\n", absPath)
				break
			}
		}
	}

	if !envLoaded {
		if err := godotenv.Load(); err == nil {
			envLoaded = true
			loadedPath, _ = filepath.Abs(".env")
			log.Printf("✓ Loaded .env from: %s\n", loadedPath)
		}
	}

	if !envLoaded {
		log.Println("⚠ Note: .env file not found, using system environment variables")
	}

	// Get master database URL
	masterDBURL := os.Getenv("MASTER_DB_URL")
	if masterDBURL == "" {
		log.Fatal("❌ MASTER_DB_URL is not set. Please set it in .env or as an environment variable.")
	}

	log.Printf("✓ MASTER_DB_URL found (length: %d characters)\n", len(masterDBURL))

	// Normalize to a URL Atlas can consume (pgx keyword DSNs are not
	// accepted by Atlas) and apply the docker-service host override.
	masterURL, err := dsnToURL(masterDBURL, *hostOverride)
	if err != nil {
		log.Fatalf("❌ Invalid MASTER_DB_URL: %v", err)
	}

	var absSchemaPath, devURL string
	var absMigPath, firstMig string

	if *declarativeMode {
		absSchemaPath, err = resolveSchemaDir(*schemaDir)
		if err != nil {
			log.Fatalf("Failed to resolve schema directory: %v", err)
		}
		log.Printf("✓ Declarative mode active. Using schema directory: %s\n", absSchemaPath)
		devURL = resolveDevURL(*devDBURL)
		log.Printf("✓ Using dev database: %s\n", devURL)
	} else {
		// Resolve migrations directory
		migPath := *migrationsDir
		if migPath == "" {
			candidates := []string{
				"../../packages/core/db/migrations",
				"../packages/core/db/migrations",
				"packages/core/db/migrations",
				"./migrations",
				"/app/migrations",
			}
			for _, c := range candidates {
				if info, err := os.Stat(c); err == nil && info.IsDir() {
					migPath = c
					break
				}
			}
			if migPath == "" {
				migPath = "packages/core/db/migrations"
			}
		}

		absMigPath, err = filepath.Abs(migPath)
		if err != nil {
			log.Fatalf("Failed to resolve absolute path for migrations dir: %v", err)
		}
		log.Printf("✓ Using migrations directory: %s\n", absMigPath)

		// First (oldest) migration version — used as the baseline for databases
		// that have schema but no atlas_schema_revisions table.
		firstMig, err = firstMigrationVersion(absMigPath)
		if err != nil {
			log.Fatalf("Failed to determine first migration version: %v", err)
		}
		log.Printf("✓ First migration version (baseline for legacy DBs): %s\n", firstMig)
	}

	atlasBin := resolveAtlasBinary()
	log.Printf("✓ Using Atlas binary: %s\n", atlasBin)

	ctx := context.Background()

	// Connect to master database
	pool, err := pgxpool.New(ctx, masterURL)
	if err != nil {
		log.Fatalf("Unable to connect to master database: %v", err)
	}
	defer pool.Close()

	// 1. Optionally migrate/sync master database
	if *includeMaster {
		log.Println("\n==================================================")
		if *declarativeMode {
			log.Println("⚡ Running Atlas declarative schema sync on Master Database...")
		} else {
			log.Println("⚡ Running Atlas migration on Master Database...")
		}
		log.Println("==================================================")

		if *declarativeMode {
			if err := executeAtlasDeclarative(atlasBin, masterURL, absSchemaPath, devURL, *statusOnly); err != nil {
				log.Fatalf("❌ Master database declarative schema sync failed: %v", err)
			}
		} else {
			baseline, err := resolveBaseline(ctx, masterURL, *baselineVer, firstMig)
			if err != nil {
				log.Fatalf("❌ Failed to inspect master database: %v", err)
			}
			logBaseline(baseline)

			if err := executeAtlas(atlasBin, masterURL, absMigPath, baseline, *statusOnly); err != nil {
				log.Fatalf("❌ Master database migration failed: %v", err)
			}
		}
		log.Println("✅ Master database schema update completed successfully!")
	}

	// 2. Discover all active tenants
	tenants, err := getAllActiveTenants(ctx, pool)
	if err != nil {
		log.Fatalf("Failed to query tenants from master database: %v", err)
	}

	if len(tenants) == 0 {
		log.Println("\nℹ No active tenants found in master database.")
		return
	}

	log.Printf("\nFound %d active tenant(s) to process.\n", len(tenants))

	// 3. Migrate/sync each tenant database
	successCount := 0
	failedCount := 0
	skippedCount := 0

	for _, tenant := range tenants {
		log.Println("\n--------------------------------------------------")
		log.Printf("🚀 Processing tenant: %s (slug: %s)\n", tenant.TenantName, tenant.Slug)
		log.Println("--------------------------------------------------")

		if tenant.DbConnStr == "" {
			log.Printf("❌ Skipping tenant %s: db_conn_str is empty\n", tenant.Slug)
			failedCount++
			continue
		}

		tenantURL, err := dsnToURL(tenant.DbConnStr, *hostOverride)
		if err != nil {
			log.Printf("❌ Failed to parse connection string for tenant %s: %v\n", tenant.Slug, err)
			failedCount++
			continue
		}

		if *declarativeMode {
			err = executeAtlasDeclarative(atlasBin, tenantURL, absSchemaPath, devURL, *statusOnly)
			if err != nil {
				if isMissingDatabase(err) {
					log.Printf("⚠ Skipping tenant %s: its database does not exist on the server (%v).\n", tenant.Slug, err)
					log.Printf("  Create the database or set is_active = false for this tenant in the master DB.\n")
					skippedCount++
					continue
				}
				log.Printf("❌ Failed to apply declarative schema to tenant %s: %v\n", tenant.Slug, err)
				failedCount++
				continue
			}
			log.Printf("✅ Successfully synced declarative schema for tenant: %s\n", tenant.Slug)
			successCount++
		} else {
			baseline, err := resolveBaseline(ctx, tenantURL, *baselineVer, firstMig)
			if err != nil {
				if isMissingDatabase(err) {
					log.Printf("⚠ Skipping tenant %s: its database does not exist on the server (%v).\n", tenant.Slug, err)
					log.Printf("  Create the database or set is_active = false for this tenant in the master DB.\n")
					skippedCount++
					continue
				}
				log.Printf("❌ Failed to inspect tenant %s: %v\n", tenant.Slug, err)
				failedCount++
				continue
			}
			logBaseline(baseline)

			err = executeAtlas(atlasBin, tenantURL, absMigPath, baseline, *statusOnly)
			if err != nil {
				log.Printf("❌ Failed to migrate tenant %s: %v\n", tenant.Slug, err)
				failedCount++
				continue
			}

			log.Printf("✅ Successfully migrated tenant: %s\n", tenant.Slug)
			successCount++
		}
	}

	log.Println("\n==================================================")
	if *declarativeMode {
		log.Println("=== Multi-Tenant Declarative Schema Sync Summary ===")
	} else {
		log.Println("=== Multi-Tenant Atlas Migration Summary ===")
	}
	log.Println("==================================================")
	log.Printf("Successful: %d\n", successCount)
	log.Printf("Skipped:    %d (database missing)\n", skippedCount)
	log.Printf("Failed:     %d\n", failedCount)
	log.Printf("Total:      %d\n", len(tenants))

	if failedCount > 0 {
		os.Exit(1)
	}
}

// firstMigrationVersion returns the oldest versioned migration file
// (YYYYMMDDHHMMSS_*.sql) in dir — used as the baseline for legacy databases.
func firstMigrationVersion(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read migrations dir: %w", err)
	}
	var first string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		name := e.Name()
		if len(name) < 14 {
			continue
		}
		ver := name[:14]
		if _, err := strconv.ParseInt(ver, 10, 64); err != nil {
			continue
		}
		if first == "" || ver < first {
			first = ver
		}
	}
	if first == "" {
		return "", fmt.Errorf("no versioned migration files found in %s", dir)
	}
	return first, nil
}

// resolveBaseline decides the Atlas --baseline for one database:
//   - an explicit -baseline flag always wins;
//   - a database already tracked by Atlas (atlas_schema_revisions table)
//     needs no baseline (apply pending only);
//   - a fresh/empty database needs no baseline (apply everything);
//   - a database that has schema objects but no revisions table (legacy /
//     restored) needs the first migration version as baseline, otherwise
//     Atlas refuses with "connected database is not clean".
func resolveBaseline(ctx context.Context, dbURL, explicit, firstMig string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return "", fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	var hasRevisions bool
	err = pool.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM information_schema.tables
		WHERE table_name = 'atlas_schema_revisions'
		  AND table_schema NOT IN ('pg_catalog', 'information_schema')
	)`).Scan(&hasRevisions)
	if err != nil {
		return "", fmt.Errorf("check atlas_schema_revisions: %w", err)
	}
	if hasRevisions {
		return "", nil
	}

	var objectCount int
	err = pool.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM information_schema.tables    WHERE table_schema NOT IN ('pg_catalog', 'information_schema')) +
		(SELECT count(*) FROM information_schema.views     WHERE table_schema NOT IN ('pg_catalog', 'information_schema')) +
		(SELECT count(*) FROM information_schema.sequences WHERE sequence_schema NOT IN ('pg_catalog', 'information_schema')) +
		(SELECT count(*) FROM information_schema.routines  WHERE routine_schema NOT IN ('pg_catalog', 'information_schema'))`).Scan(&objectCount)
	if err != nil {
		return "", fmt.Errorf("check database emptiness: %w", err)
	}
	if objectCount == 0 {
		return "", nil
	}
	return firstMig, nil
}

// logBaseline prints which baseline strategy is used for the current database.
func logBaseline(baseline string) {
	switch baseline {
	case "":
		log.Println("  ℹ baseline: none (fresh or Atlas-tracked database)")
	default:
		log.Printf("  ℹ baseline: %s (schema present, no revisions table)\n", baseline)
	}
}

var sslModeRe = regexp.MustCompile(`(?i)\bsslmode=([a-z]+)`)

// dsnToURL converts any pgx-accepted connection string (URL or keyword DSN
// like "host=postgres user=x dbname=y") into a postgres:// URL that the Atlas
// CLI can consume, applying the docker-service host override (from=to).
func dsnToURL(connStr, hostOverride string) (string, error) {
	cfg, err := pgx.ParseConfig(connStr)
	if err != nil {
		return "", fmt.Errorf("parse connection string: %w", err)
	}

	host := cfg.Host
	overridden := false
	if hostOverride != "" {
		if from, to, ok := strings.Cut(hostOverride, "="); ok && strings.EqualFold(host, from) {
			host = to
			overridden = true
		}
	}
	if strings.HasPrefix(host, "/") {
		return "", fmt.Errorf("unix-socket hosts are not supported by Atlas URLs (host=%q)", host)
	}

	q := url.Values{}
	if m := sslModeRe.FindStringSubmatch(connStr); len(m) == 2 {
		q.Set("sslmode", m[1])
	} else if overridden {
		// Docker-local connections default to no TLS.
		q.Set("sslmode", "disable")
	}

	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(host, strconv.Itoa(int(cfg.Port))),
		Path:     "/" + cfg.Database,
		RawQuery: q.Encode(),
	}
	return u.String(), nil
}

// isMissingDatabase reports whether the error is PostgreSQL's
// "database ... does not exist" (SQLSTATE 3D000 / invalid_catalog_name).
func isMissingDatabase(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "3D000" {
		return true
	}
	// Fallback for wrapped/non-pgconn errors.
	return strings.Contains(err.Error(), `database "`) && strings.Contains(err.Error(), "does not exist")
}

// resolveAtlasBinary locates the Atlas CLI executable
func resolveAtlasBinary() string {
	if custom := os.Getenv("ATLAS_PATH"); custom != "" {
		return custom
	}
	if path, err := exec.LookPath("atlas"); err == nil {
		return path
	}
	// Common fallback locations on Windows / Linux
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, "go", "bin", "atlas.exe"),
		filepath.Join(home, "go", "bin", "atlas"),
		"/usr/local/bin/atlas",
		"/usr/bin/atlas",
		"C:\\Program Files\\Atlas\\atlas.exe",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return "atlas"
}

// resolveSchemaDir determines the absolute path to the directory containing Atlas SQL schema files.
func resolveSchemaDir(customPath string) (string, error) {
	if customPath != "" {
		return filepath.Abs(customPath)
	}
	candidates := []string{
		"../../packages/core/db/schema",
		"../packages/core/db/schema",
		"packages/core/db/schema",
		"./schema",
		"/app/schema",
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return filepath.Abs(c)
		}
	}
	return filepath.Abs("packages/core/db/schema")
}

// resolveDevURL resolves the Atlas dev database URL used for calculating declarative diffs.
func resolveDevURL(customURL string) string {
	if customURL != "" {
		return customURL
	}
	if env := os.Getenv("ATLAS_DEV_URL"); env != "" {
		return env
	}
	return "docker://postgres/16/dev"
}

// executeAtlasDeclarative invokes Atlas CLI to declaratively sync a database to the schema directory.
func executeAtlasDeclarative(atlasBin, dbURL, schemaDir, devURL string, dryRun bool) error {
	schemaURI := "file://" + filepath.ToSlash(schemaDir)

	args := []string{
		"schema", "apply",
		"--url", dbURL,
		"--to", schemaURI,
		"--dev", devURL,
	}
	if dryRun {
		args = append(args, "--dry-run")
	} else {
		args = append(args, "--auto-approve")
	}

	cmd := exec.Command(atlasBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// executeAtlas invokes the Atlas CLI to apply or inspect migrations
func executeAtlas(atlasBin, dbURL, migrationsDir, baseline string, statusOnly bool) error {
	dirURI := "file://" + filepath.ToSlash(migrationsDir)

	var args []string
	if statusOnly {
		args = []string{"migrate", "status", "--url", dbURL, "--dir", dirURI}
	} else {
		args = []string{"migrate", "apply", "--url", dbURL, "--dir", dirURI}
		if baseline != "" {
			args = append(args, "--baseline", baseline)
		}
	}

	cmd := exec.Command(atlasBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// getAllActiveTenants retrieves all active tenants from the master database
func getAllActiveTenants(ctx context.Context, pool *pgxpool.Pool) ([]repository.Tenant, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	// Check if tenants table exists
	var exists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'tenants')").Scan(&exists)
	if err != nil || !exists {
		return nil, nil
	}

	rows, err := conn.Query(ctx, "SELECT id, tenant_name, slug, db_conn_str, is_active, settings, created_at, updated_at FROM tenants WHERE is_active = true")
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants: %w", err)
	}
	defer rows.Close()

	var tenants []repository.Tenant
	for rows.Next() {
		var tenant repository.Tenant
		err := rows.Scan(
			&tenant.ID,
			&tenant.TenantName,
			&tenant.Slug,
			&tenant.DbConnStr,
			&tenant.IsActive,
			&tenant.Settings,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tenants: %w", err)
	}

	return tenants, nil
}
