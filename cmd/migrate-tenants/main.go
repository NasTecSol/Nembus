package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"NEMBUS/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Parse command line flags
	down := flag.Bool("down", false, "Rollback migrations instead of applying them")
	migrationsDir := flag.String("dir", "./migrations", "Directory containing migration files")
	flag.Parse()

	// Get current working directory for debugging
	cwd, _ := os.Getwd()
	log.Printf("Current working directory: %s\n", cwd)

	// Load environment variables from project root
	// Try multiple locations: current dir, parent dir (if running from cmd/), and explicit .env
	envPaths := []string{".env", "../.env", "../../.env"}
	var envLoaded bool
	var loadedPath string
	for _, envPath := range envPaths {
		absPath, _ := filepath.Abs(envPath)
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err == nil {
				envLoaded = true
				loadedPath = absPath
				log.Printf("✓ Loaded .env from: %s\n", absPath)
				break
			}
		}
	}

	// Also try loading from current working directory (default behavior)
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
		log.Fatal("❌ MASTER_DB_URL is not set. Please:\n" +
			"  1. Create a .env file in the project root with: MASTER_DB_URL=postgresql://...\n" +
			"  2. Or set it as an environment variable: export MASTER_DB_URL=...")
	}

	log.Printf("✓ MASTER_DB_URL found (length: %d characters)\n", len(masterDBURL))

	ctx := context.Background()

	// Connect to master database
	pool, err := pgxpool.New(ctx, masterDBURL)
	if err != nil {
		log.Fatalf("Unable to connect to master database: %v", err)
	}
	defer pool.Close()

	// Get all active tenants directly from the pool
	tenants, err := getAllActiveTenants(ctx, pool)
	if err != nil {
		log.Fatalf("Failed to get tenants: %v", err)
	}

	if len(tenants) == 0 {
		log.Println("No active tenants found")
		return
	}

	log.Printf("Found %d active tenant(s)\n", len(tenants))

	// Get absolute path to migrations directory
	migrationsPath, err := filepath.Abs(*migrationsDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for migrations directory: %v", err)
	}

	// Run migrations for each tenant
	successCount := 0
	failedCount := 0

	for _, tenant := range tenants {
		log.Printf("\n--- Migrating tenant: %s (slug: %s) ---\n", tenant.TenantName, tenant.Slug)

		action := "up"
		if *down {
			action = "down"
		}

		err := runMigrations(tenant.DbConnStr, migrationsPath, action)
		if err != nil {
			log.Printf("❌ Failed to migrate tenant %s: %v\n", tenant.Slug, err)
			failedCount++
			continue
		}

		log.Printf("✅ Successfully migrated tenant: %s\n", tenant.Slug)
		successCount++
	}

	log.Printf("\n=== Migration Summary ===\n")
	log.Printf("Successful: %d\n", successCount)
	log.Printf("Failed: %d\n", failedCount)
	log.Printf("Total: %d\n", len(tenants))

	if failedCount > 0 {
		os.Exit(1)
	}
}

// getAllActiveTenants retrieves all active tenants from the master database
func getAllActiveTenants(ctx context.Context, pool *pgxpool.Pool) ([]repository.Tenant, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

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

// runMigrations executes goose migrations on a tenant database
func runMigrations(dbConnStr, migrationsDir, action string) error {
	// Set environment variables for goose
	os.Setenv("GOOSE_DRIVER", "postgres")
	os.Setenv("GOOSE_DBSTRING", dbConnStr)

	// Build goose command
	var cmd *exec.Cmd
	if action == "down" {
		cmd = exec.Command("goose", "-dir", migrationsDir, "down")
	} else {
		cmd = exec.Command("goose", "-dir", migrationsDir, "up")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
