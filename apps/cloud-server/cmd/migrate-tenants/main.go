package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Parse command line flags
	includeMaster := flag.Bool("master", true, "Also run migrations on the master database")
	migrationsDir := flag.String("dir", "", "Directory containing Atlas migration files (default: auto-detected packages/core/db/migrations)")
	baselineVer := flag.String("baseline", "", "Baseline migration version. Leave empty (default) to AUTO-DETECT per database: empty DBs and DBs already tracked by Atlas (atlas_schema_revisions) get no baseline; DBs that have schema but no revisions table get the first migration version. Set it explicitly to force the same baseline for every database (e.g. -baseline 20260813124500) — a baseline that does not exist in the migration directory is a hard error.")
	statusOnly := flag.Bool("status", false, "Show migration status without applying")
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

	absMigPath, err := filepath.Abs(migPath)
	if err != nil {
		log.Fatalf("Failed to resolve absolute path for migrations dir: %v", err)
	}
	log.Printf("✓ Using migrations directory: %s\n", absMigPath)

	// First (oldest) migration version — used as the baseline for databases
	// that have schema but no atlas_schema_revisions table.
	firstMig, err := firstMigrationVersion(absMigPath)
	if err != nil {
		log.Fatalf("Failed to determine first migration version: %v", err)
	}
	log.Printf("✓ First migration version (baseline for legacy DBs): %s\n", firstMig)

	atlasBin := resolveAtlasBinary()
	log.Printf("✓ Using Atlas binary: %s\n", atlasBin)

	ctx := context.Background()

	// Connect to master database
	pool, err := pgxpool.New(ctx, masterDBURL)
	if err != nil {
		log.Fatalf("Unable to connect to master database: %v", err)
	}
	defer pool.Close()

	// 1. Optionally migrate master database
	if *includeMaster {
		log.Println("\n==================================================")
		log.Println("⚡ Running Atlas migration on Master Database...")
		log.Println("==================================================")

		baseline, err := resolveBaseline(ctx, masterDBURL, *baselineVer, firstMig)
		if err != nil {
			log.Fatalf("❌ Failed to inspect master database: %v", err)
		}
		logBaseline(baseline)

		if err := executeAtlas(atlasBin, masterDBURL, absMigPath, baseline, *statusOnly); err != nil {
			log.Fatalf("❌ Master database migration failed: %v", err)
		}
		log.Println("✅ Master database migration completed successfully!")
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

	log.Printf("\nFound %d active tenant(s) to migrate.\n", len(tenants))

	// 3. Migrate each tenant database
	successCount := 0
	failedCount := 0

	for _, tenant := range tenants {
		log.Println("\n--------------------------------------------------")
		log.Printf("🚀 Migrating tenant: %s (slug: %s)\n", tenant.TenantName, tenant.Slug)
		log.Println("--------------------------------------------------")

		if tenant.DbConnStr == "" {
			log.Printf("❌ Skipping tenant %s: db_conn_str is empty\n", tenant.Slug)
			failedCount++
			continue
		}

		baseline, err := resolveBaseline(ctx, tenant.DbConnStr, *baselineVer, firstMig)
		if err != nil {
			log.Printf("❌ Failed to inspect tenant %s: %v\n", tenant.Slug, err)
			failedCount++
			continue
		}
		logBaseline(baseline)

		err = executeAtlas(atlasBin, tenant.DbConnStr, absMigPath, baseline, *statusOnly)
		if err != nil {
			log.Printf("❌ Failed to migrate tenant %s: %v\n", tenant.Slug, err)
			failedCount++
			continue
		}

		log.Printf("✅ Successfully migrated tenant: %s\n", tenant.Slug)
		successCount++
	}

	log.Println("\n==================================================")
	log.Println("=== Multi-Tenant Atlas Migration Summary ===")
	log.Println("==================================================")
	log.Printf("Successful: %d\n", successCount)
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
		WHERE table_schema = 'public' AND table_name = 'atlas_schema_revisions'
	)`).Scan(&hasRevisions)
	if err != nil {
		return "", fmt.Errorf("check atlas_schema_revisions: %w", err)
	}
	if hasRevisions {
		return "", nil
	}

	var objectCount int
	err = pool.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM information_schema.tables    WHERE table_schema = 'public') +
		(SELECT count(*) FROM information_schema.views     WHERE table_schema = 'public') +
		(SELECT count(*) FROM information_schema.sequences WHERE sequence_schema = 'public') +
		(SELECT count(*) FROM information_schema.routines  WHERE routine_schema = 'public')`).Scan(&objectCount)
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
