package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/NasTecSol/nembus-core/repository"

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
		log.Fatal("Failed to initialize master DB connection")
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
	fmt.Printf("  DB Connection: %s\n", sanitizeConnectionString(dbConnStr))

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
			fmt.Printf("Failed to initialize tenant DB connection\n")
		} else {
			fmt.Printf("✓ Successfully connected to tenant database\n")
			tenantPool.Close()
		}
	}
}

const redactedConnectionString = "<redacted>"

// sanitizeConnectionString returns only non-sensitive, deliberately reconstructed
// metadata for recognized PostgreSQL URL DSNs. Unsupported or malformed input is
// redacted in full so a failed sanitization can never fall back to the raw DSN.
func sanitizeConnectionString(connStr string) string {
	raw := strings.TrimSpace(connStr)
	if raw == "" {
		return redactedConnectionString
	}

	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.Host == "" || parsed.Opaque != "" || parsed.Fragment != "" {
		return redactedConnectionString
	}

	host := parsed.Hostname()
	if host == "" || !safeDiagnosticHost(host) {
		return redactedConnectionString
	}

	port := parsed.Port()
	if port != "" {
		portNumber, err := strconv.Atoi(port)
		if err != nil || portNumber < 1 || portNumber > 65535 {
			return redactedConnectionString
		}
	}

	metadata := "host=" + host
	if port != "" {
		metadata += " port=" + port
	}
	return metadata + " credentials=" + redactedConnectionString
}

func safeDiagnosticHost(host string) bool {
	for _, r := range host {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '-' || r == ':' {
			continue
		}
		return false
	}
	return true
}
