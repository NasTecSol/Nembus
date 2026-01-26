package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"NEMBUS/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	slug := flag.String("slug", "", "Tenant slug to check")
	flag.Parse()

	if *slug == "" {
		log.Fatal("Usage: go run cmd/check-tenant/main.go -slug <tenant-slug>")
	}

	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found")
	}

	masterDBURL := os.Getenv("MASTER_DB_URL")
	if masterDBURL == "" {
		log.Fatal("MASTER_DB_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, masterDBURL)
	if err != nil {
		log.Fatalf("Failed to connect to master DB: %v", err)
	}
	defer pool.Close()

	queries := repository.New(pool)

	// Check if tenant exists (any status)
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("Failed to acquire connection: %v", err)
	}
	defer conn.Release()

	var tenantName, dbConnStr string
	var isActive *bool
	var exists bool

	err = conn.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM tenants WHERE slug = $1), tenant_name, db_conn_str, is_active FROM tenants WHERE slug = $1",
		*slug).Scan(&exists, &tenantName, &dbConnStr, &isActive)

	if err != nil {
		log.Fatalf("Error querying tenant: %v", err)
	}

	fmt.Printf("\n=== Tenant Check for slug: '%s' ===\n", *slug)
	if !exists {
		fmt.Printf("❌ Tenant NOT FOUND\n")
		fmt.Printf("\nAvailable tenants:\n")
		rows, _ := conn.Query(ctx, "SELECT slug, tenant_name, is_active FROM tenants ORDER BY slug")
		for rows.Next() {
			var s, n string
			var a *bool
			rows.Scan(&s, &n, &a)
			active := "inactive"
			if a != nil && *a {
				active = "active"
			}
			fmt.Printf("  - %s (%s) - %s\n", s, n, active)
		}
		rows.Close()
		return
	}

	fmt.Printf("✓ Tenant EXISTS\n")
	fmt.Printf("  Name: %s\n", tenantName)
	if isActive == nil {
		fmt.Printf("  Status: ❌ is_active is NULL\n")
	} else if *isActive {
		fmt.Printf("  Status: ✓ ACTIVE\n")
	} else {
		fmt.Printf("  Status: ❌ INACTIVE (is_active = false)\n")
	}
	fmt.Printf("  DB Connection: %s\n", maskPassword(dbConnStr))

	// Try to get using GetTenantBySlug (only returns active)
	tenant, err := queries.GetTenantBySlug(ctx, *slug)
	if err != nil {
		fmt.Printf("\n❌ GetTenantBySlug failed: %v\n", err)
		fmt.Printf("   This means the tenant won't be accessible via middleware.\n")
	} else {
		fmt.Printf("\n✓ GetTenantBySlug SUCCESS\n")
		fmt.Printf("  Tenant ID: %s\n", tenant.ID)
	}

	// Try to connect to tenant database
	if dbConnStr != "" {
		fmt.Printf("\nTesting tenant database connection...\n")
		tenantPool, err := pgxpool.New(ctx, dbConnStr)
		if err != nil {
			fmt.Printf("❌ Failed to connect: %v\n", err)
		} else {
			fmt.Printf("✓ Successfully connected to tenant database\n")
			tenantPool.Close()
		}
	}
}

func maskPassword(connStr string) string {
	// Simple password masking for display
	// In a real connection string like: postgres://user:pass@host:port/db
	// This is a simple approach - you might want to use url.Parse for better handling
	if len(connStr) > 50 {
		return connStr[:30] + "..." + connStr[len(connStr)-20:]
	}
	return connStr
}
