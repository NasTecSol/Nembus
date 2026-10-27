<#
.SYNOPSIS
    Tests a PostgreSQL / Supabase connection string.
#>
param(
    [Parameter(Mandatory = $true)]
    [string]$DbUrl
)

$ErrorActionPreference = "Stop"

# Ensure sslmode=require if connecting to Supabase
if ($DbUrl -match "supabase" -and $DbUrl -notmatch "sslmode=") {
    if ($DbUrl -contains "?") {
        $DbUrl += "&sslmode=require"
    } else {
        $DbUrl += "?sslmode=require"
    }
}

Write-Host "Testing database connection..." -ForegroundColor Cyan

$RepoRoot = Split-Path -Parent $PSScriptRoot
$CloudServerDir = Join-Path $RepoRoot "apps\cloud-server"

# Run a quick Go command to verify connectivity & check Atlas revisions table
$testScript = @"
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	connStr := os.Args[1]
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("❌ Failed to parse/create pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("❌ Ping failed: %v", err)
	}
	fmt.Println("✅ Connection successful!")

	var version string
	if err := pool.QueryRow(ctx, "SELECT version()").Scan(&version); err == nil {
		fmt.Printf("   PostgreSQL Version: %s\n", version)
	}

	var hasAtlas bool
	_ = pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'atlas_schema_revisions')").Scan(&hasAtlas)
	if hasAtlas {
		fmt.Println("   Atlas Revisions: Found existing atlas_schema_revisions table.")
	} else {
		fmt.Println("   Atlas Revisions: Fresh database (no atlas_schema_revisions table yet).")
	}

	var tableCount int
	_ = pool.QueryRow(ctx, "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)
	fmt.Printf("   Public Tables: %d table(s) found.\n", tableCount)
}
"@

$tmpGoFile = Join-Path $env:TEMP "test_db_conn.go"
Set-Content -Path $tmpGoFile -Value $testScript -Encoding UTF8

Push-Location $CloudServerDir
try {
    go run $tmpGoFile $DbUrl
} finally {
    Pop-Location
    if (Test-Path $tmpGoFile) {
        Remove-Item $tmpGoFile -Force
    }
}
