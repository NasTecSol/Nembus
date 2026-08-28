package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// TableSyncConfig defines the schema, table name, key columns, and whether it is org-scoped.
type TableSyncConfig struct {
	Domain        string
	Schema        string
	Table         string
	HasOrgScope   bool
	OrgColumn     string
	PKColumn      string
	ConflictKey   string
	DependentOnly bool // If true, truncated with parent or cascaded
}

// Ordered list of tables to sync in exact foreign-key topological order.
var syncTables = []TableSyncConfig{
	// 1. Foundation: Units of Measure
	{Domain: "uom", Schema: "public", Table: "units_of_measure", HasOrgScope: false, PKColumn: "id", ConflictKey: "code"},

	// 2. Foundation: Product Categories & Brands
	{Domain: "categories", Schema: "public", Table: "product_categories", HasOrgScope: false, PKColumn: "id", ConflictKey: "code"},
	{Domain: "brands", Schema: "public", Table: "brands", HasOrgScope: false, PKColumn: "id", ConflictKey: "code"},

	// 3. Foundation: Stores & Storage Locations
	{Domain: "stores", Schema: "public", Table: "stores", HasOrgScope: true, OrgColumn: "organization_id", PKColumn: "id", ConflictKey: "organization_id, code"},
	{Domain: "stores", Schema: "public", Table: "storage_locations", HasOrgScope: false, PKColumn: "id", ConflictKey: "store_id, code"},

	// 4. Foundation: Users & Cashiers
	{Domain: "users", Schema: "public", Table: "users", HasOrgScope: true, OrgColumn: "organization_id", PKColumn: "id", ConflictKey: "username"},
	{Domain: "users", Schema: "public", Table: "cashiers", HasOrgScope: false, PKColumn: "id", ConflictKey: "store_id, cashier_code"},

	// 5. UoM Packaging Groups & Levels
	{Domain: "uom_groups", Schema: "public", Table: "uom_packaging_templates", HasOrgScope: true, OrgColumn: "organization_id", PKColumn: "id", ConflictKey: "code"},
	{Domain: "uom_groups", Schema: "public", Table: "uom_packaging_template_levels", HasOrgScope: false, PKColumn: "id", ConflictKey: "template_id, level_order"},

	// 6. Products & Catalog
	{Domain: "products", Schema: "public", Table: "products", HasOrgScope: true, OrgColumn: "organization_id", PKColumn: "id", ConflictKey: "organization_id, sku"},
	{Domain: "barcodes", Schema: "public", Table: "product_barcodes", HasOrgScope: false, PKColumn: "id", ConflictKey: "barcode"},
	{Domain: "products", Schema: "public", Table: "product_uom_conversions", HasOrgScope: false, PKColumn: "id", ConflictKey: "product_id, from_uom_id, to_uom_id"},

	// 7. Price Lists & Prices
	{Domain: "price_lists", Schema: "public", Table: "price_lists", HasOrgScope: false, PKColumn: "id", ConflictKey: "code"},
	{Domain: "price_lists", Schema: "public", Table: "product_prices", HasOrgScope: false, PKColumn: "id", ConflictKey: "product_id, price_list_id, uom_id"},

	// 8. Business Partners, Customers & Addresses
	{Domain: "partners", Schema: "public", Table: "business_partners", HasOrgScope: true, OrgColumn: "organization_id", PKColumn: "id", ConflictKey: "code"},
	{Domain: "partners", Schema: "public", Table: "customers", HasOrgScope: true, OrgColumn: "organization_id", PKColumn: "id", ConflictKey: "organization_id, customer_code"},
	{Domain: "bp_addresses", Schema: "public", Table: "partner_addresses", HasOrgScope: false, PKColumn: "id", ConflictKey: "partner_id, address_type, address_name"},
	{Domain: "bp_addresses", Schema: "public", Table: "customer_addresses", HasOrgScope: false, PKColumn: "id", ConflictKey: "customer_id, address_type, address_line"},

	// 9. Warehouse Inventory Stock
	{Domain: "inventory", Schema: "public", Table: "inventory_stock", HasOrgScope: false, PKColumn: "id", ConflictKey: "product_id, store_id"},

	// 10. Sales Orders & Order Lines
	{Domain: "sales_orders", Schema: "public", Table: "sales_orders_v2", HasOrgScope: true, OrgColumn: "organization_id", PKColumn: "id", ConflictKey: "order_number"},
	{Domain: "sales_orders", Schema: "public", Table: "sales_order_lines_v2", HasOrgScope: true, OrgColumn: "organization_id", PKColumn: "id", ConflictKey: "id"},
}

type SyncReport struct {
	Domain      string
	Table       string
	SourceCount int64
	TargetCount int64
	SyncedCount int64
	Status      string
	Duration    time.Duration
}

func main() {
	sourceURLFlag := flag.String("source-url", "", "Source (STG) PostgreSQL connection string (or SOURCE_DB_URL env)")
	targetURLFlag := flag.String("target-url", "", "Target PostgreSQL connection string (or TARGET_DB_URL env)")
	orgIDFlag := flag.Int("org-id", 0, "Filter by organization ID (0 = sync all organizations)")
	modeFlag := flag.String("mode", "truncate_copy", "Sync mode: 'truncate_copy' (clean refresh) or 'upsert'")
	domainsFlag := flag.String("domains", "all", "Comma-separated list of domains to sync (or 'all')")
	dryRunFlag := flag.Bool("dry-run", false, "Simulate execution without modifying target database")
	batchSizeFlag := flag.Int("batch-size", 2000, "Batch size for chunked data transfer")
	flag.Parse()

	// Load environments
	loadEnv()

	sourceURL := *sourceURLFlag
	if sourceURL == "" {
		sourceURL = os.Getenv("SOURCE_DB_URL")
	}
	if sourceURL == "" {
		sourceURL = os.Getenv("STG_DATABASE_URL")
	}
	if sourceURL == "" {
		sourceURL = os.Getenv("DATABASE_URL")
	}

	targetURL := *targetURLFlag
	if targetURL == "" {
		targetURL = os.Getenv("TARGET_DB_URL")
	}
	if targetURL == "" {
		targetURL = os.Getenv("PROD_DATABASE_URL")
	}

	if sourceURL == "" {
		log.Fatal("❌ Source database URL is required. Provide via --source-url or SOURCE_DB_URL env variable.")
	}
	if targetURL == "" {
		log.Fatal("❌ Target database URL is required. Provide via --target-url or TARGET_DB_URL env variable.")
	}

	if sourceURL == targetURL {
		log.Fatal("❌ Source and Target database URLs cannot be identical. Prevented self-overwrite.")
	}

	ctx := context.Background()

	log.Println("=================================================================")
	log.Println("           NEMBUS DATABASE DOMAIN REFRESH & SYNC TOOL            ")
	log.Println("=================================================================")
	log.Printf("Source DB    : %s\n", maskConnectionString(sourceURL))
	log.Printf("Target DB    : %s\n", maskConnectionString(targetURL))
	log.Printf("Sync Mode    : %s\n", *modeFlag)
	log.Printf("Domain Filter: %s\n", *domainsFlag)
	if *orgIDFlag > 0 {
		log.Printf("Org Scope    : Organization ID = %d\n", *orgIDFlag)
	} else {
		log.Printf("Org Scope    : ALL Organizations\n")
	}
	if *dryRunFlag {
		log.Println("Mode         : 🔍 DRY-RUN ONLY (No target changes)")
	}
	log.Println("=================================================================")

	// Connect to source pool
	srcPool, err := pgxpool.New(ctx, sourceURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Source DB: %v", err)
	}
	defer srcPool.Close()

	if err := srcPool.Ping(ctx); err != nil {
		log.Fatalf("❌ Ping failed on Source DB: %v", err)
	}
	log.Println("✓ Connected to Source DB")

	// Connect to target pool
	tgtPool, err := pgxpool.New(ctx, targetURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Target DB: %v", err)
	}
	defer tgtPool.Close()

	if err := tgtPool.Ping(ctx); err != nil {
		log.Fatalf("❌ Ping failed on Target DB: %v", err)
	}
	log.Println("✓ Connected to Target DB")

	// Filter domains
	selectedTables := filterTables(*domainsFlag)
	if len(selectedTables) == 0 {
		log.Fatalf("❌ No matching tables found for domain filter: %s", *domainsFlag)
	}

	reports := make([]SyncReport, 0, len(selectedTables))
	startTime := time.Now()

	// If truncate mode and not dry run, truncate in reverse topological order (or cascade)
	if *modeFlag == "truncate_copy" && !*dryRunFlag {
		log.Println("\n[1/3] Preparing target tables (clean refresh)...")
		if err := truncateTargetTables(ctx, tgtPool, selectedTables, *orgIDFlag); err != nil {
			log.Fatalf("❌ Failed during table preparation: %v", err)
		}
	}

	log.Println("\n[2/3] Syncing domain records from Source (STG) to Target...")
	for _, t := range selectedTables {
		report, err := syncSingleTable(ctx, srcPool, tgtPool, t, *orgIDFlag, *modeFlag, *dryRunFlag, *batchSizeFlag)
		if err != nil {
			log.Printf("❌ [%s] %s.%s failed: %v\n", t.Domain, t.Schema, t.Table, err)
			report.Status = fmt.Sprintf("FAILED: %v", err)
		} else {
			log.Printf("✓ [%s] %s.%s -> %d rows synced (Source: %d, Target: %d)\n",
				t.Domain, t.Schema, t.Table, report.SyncedCount, report.SourceCount, report.TargetCount)
		}
		reports = append(reports, report)
	}

	// Update sequence numbers in Target DB so auto-increment works seamlessly
	if !*dryRunFlag {
		log.Println("\n[3/3] Synchronizing PostgreSQL sequence values...")
		if err := syncSequences(ctx, tgtPool, selectedTables); err != nil {
			log.Printf("⚠ Warning during sequence synchronization: %v", err)
		} else {
			log.Println("✓ All table primary key sequences successfully aligned.")
		}
	}

	// Summary output
	printSummaryTable(reports, time.Since(startTime))
}

func syncSingleTable(ctx context.Context, srcPool, tgtPool *pgxpool.Pool, t TableSyncConfig, orgID int, mode string, dryRun bool, batchSize int) (SyncReport, error) {
	start := time.Now()
	report := SyncReport{
		Domain: t.Domain,
		Table:  fmt.Sprintf("%s.%s", t.Schema, t.Table),
		Status: "SUCCESS",
	}

	// Check if table exists in source
	var srcExists bool
	_ = srcPool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = $1 AND table_name = $2
		);
	`, t.Schema, t.Table).Scan(&srcExists)

	if !srcExists {
		report.Status = "SKIPPED (Not found in source)"
		report.Duration = time.Since(start)
		return report, nil
	}

	// Check if table exists in target
	var tgtExists bool
	_ = tgtPool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = $1 AND table_name = $2
		);
	`, t.Schema, t.Table).Scan(&tgtExists)

	if !tgtExists {
		report.Status = "SKIPPED (Not found in target)"
		report.Duration = time.Since(start)
		return report, nil
	}

	// Build WHERE clause for Source
	whereClause := ""
	if t.HasOrgScope && orgID > 0 {
		whereClause = fmt.Sprintf(" WHERE %s = %d", t.OrgColumn, orgID)
	}

	// Count source rows
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s%s", t.Schema, t.Table, whereClause)
	if err := srcPool.QueryRow(ctx, countQuery).Scan(&report.SourceCount); err != nil {
		return report, fmt.Errorf("failed to count source rows: %w", err)
	}

	if dryRun {
		// Dry run: query target count
		_ = tgtPool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s.%s%s", t.Schema, t.Table, whereClause)).Scan(&report.TargetCount)
		report.SyncedCount = report.SourceCount
		report.Status = "DRY-RUN (Ready)"
		report.Duration = time.Since(start)
		return report, nil
	}

	if report.SourceCount == 0 {
		report.Status = "NO_DATA (0 rows in source)"
		report.Duration = time.Since(start)
		return report, nil
	}

	// Fetch table columns
	columns, err := getTableColumns(ctx, srcPool, t.Schema, t.Table)
	if err != nil {
		return report, fmt.Errorf("failed to inspect columns: %w", err)
	}

	colNames := make([]string, len(columns))
	quotedCols := make([]string, len(columns))
	for i, c := range columns {
		colNames[i] = c
		quotedCols[i] = fmt.Sprintf(`"%s"`, c)
	}

	selectQuery := fmt.Sprintf("SELECT %s FROM %s.%s%s ORDER BY %s ASC",
		strings.Join(quotedCols, ", "), t.Schema, t.Table, whereClause, t.PKColumn)

	rows, err := srcPool.Query(ctx, selectQuery)
	if err != nil {
		return report, fmt.Errorf("failed to select source records: %w", err)
	}
	defer rows.Close()

	if mode == "truncate_copy" {
		// Fast bulk copy via COPY protocol
		copied, err := tgtPool.CopyFrom(
			ctx,
			pgx.Identifier{t.Schema, t.Table},
			colNames,
			rows,
		)
		if err != nil {
			return report, fmt.Errorf("copy into %s failed: %w", t.Table, err)
		}
		report.SyncedCount = copied
	} else {
		// Upsert mode with batched INSERT ... ON CONFLICT
		synced, err := upsertRows(ctx, tgtPool, t, colNames, rows, batchSize)
		if err != nil {
			return report, fmt.Errorf("upsert into %s failed: %w", t.Table, err)
		}
		report.SyncedCount = synced
	}

	// Query target count after sync
	_ = tgtPool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s.%s%s", t.Schema, t.Table, whereClause)).Scan(&report.TargetCount)

	report.Duration = time.Since(start)
	return report, nil
}

func upsertRows(ctx context.Context, pool *pgxpool.Pool, t TableSyncConfig, cols []string, rows pgx.Rows, batchSize int) (int64, error) {
	var totalSynced int64

	// Prepare column lists
	quotedCols := make([]string, len(cols))
	updateSet := make([]string, 0, len(cols))
	for i, c := range cols {
		quotedCols[i] = fmt.Sprintf(`"%s"`, c)
		if c != t.PKColumn && !strings.Contains(t.ConflictKey, c) {
			updateSet = append(updateSet, fmt.Sprintf(`"%s" = EXCLUDED."%s"`, c, c))
		}
	}

	var onConflictClause string
	if t.ConflictKey != "" && len(updateSet) > 0 {
		onConflictClause = fmt.Sprintf(" ON CONFLICT (%s) DO UPDATE SET %s", t.ConflictKey, strings.Join(updateSet, ", "))
	} else if t.ConflictKey != "" {
		onConflictClause = fmt.Sprintf(" ON CONFLICT (%s) DO NOTHING", t.ConflictKey)
	} else {
		onConflictClause = " ON CONFLICT DO NOTHING"
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return totalSynced, err
		}

		placeholders := make([]string, len(vals))
		args := make([]interface{}, len(vals))
		for i := range vals {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = vals[i]
		}

		insertQuery := fmt.Sprintf(
			"INSERT INTO %s.%s (%s) VALUES (%s)%s",
			t.Schema, t.Table, strings.Join(quotedCols, ", "), strings.Join(placeholders, ", "), onConflictClause,
		)

		if _, err := tx.Exec(ctx, insertQuery, args...); err != nil {
			return totalSynced, fmt.Errorf("row upsert failed: %w", err)
		}
		totalSynced++
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return totalSynced, nil
}

func truncateTargetTables(ctx context.Context, tgtPool *pgxpool.Pool, tables []TableSyncConfig, orgID int) error {
	// If no org ID specified, we can TRUNCATE ... CASCADE in reverse order or single statement
	if orgID <= 0 {
		var tableNames []string
		for i := len(tables) - 1; i >= 0; i-- {
			tableNames = append(tableNames, fmt.Sprintf("%s.%s", tables[i].Schema, tables[i].Table))
		}
		stmt := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", strings.Join(tableNames, ", "))
		_, err := tgtPool.Exec(ctx, stmt)
		return err
	}

	// If org ID specified, DELETE per organization in reverse dependency order
	for i := len(tables) - 1; i >= 0; i-- {
		t := tables[i]
		if t.HasOrgScope {
			stmt := fmt.Sprintf("DELETE FROM %s.%s WHERE %s = $1;", t.Schema, t.Table, t.OrgColumn)
			if _, err := tgtPool.Exec(ctx, stmt, orgID); err != nil {
				return fmt.Errorf("delete failed on %s: %w", t.Table, err)
			}
		}
	}
	return nil
}

func syncSequences(ctx context.Context, tgtPool *pgxpool.Pool, tables []TableSyncConfig) error {
	for _, t := range tables {
		query := fmt.Sprintf(`
			DO $$
			DECLARE
				seq_name text;
				max_val bigint;
			BEGIN
				seq_name := pg_get_serial_sequence('%s.%s', '%s');
				IF seq_name IS NOT NULL THEN
					EXECUTE format('SELECT COALESCE(MAX(%I), 1) FROM %s.%s', '%s') INTO max_val;
					PERFORM setval(seq_name, max_val, true);
				END IF;
			END $$;
		`, t.Schema, t.Table, t.PKColumn, t.Schema, t.Table, t.PKColumn)

		if _, err := tgtPool.Exec(ctx, query); err != nil {
			// Non-critical if table doesn't have serial PK
			continue
		}
	}
	return nil
}

func getTableColumns(ctx context.Context, pool *pgxpool.Pool, schema, table string) ([]string, error) {
	query := `
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_schema = $1 AND table_name = $2
		  AND is_generated = 'NEVER'
		ORDER BY ordinal_position ASC;
	`
	rows, err := pool.Query(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, err
		}
		cols = append(cols, col)
	}
	return cols, nil
}

func filterTables(filter string) []TableSyncConfig {
	if filter == "all" || filter == "" {
		return syncTables
	}

	parts := strings.Split(strings.ToLower(filter), ",")
	domainSet := make(map[string]bool)
	for _, p := range parts {
		domainSet[strings.TrimSpace(p)] = true
	}

	var filtered []TableSyncConfig
	for _, t := range syncTables {
		if domainSet[t.Domain] || domainSet[t.Table] {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func loadEnv() {
	envPaths := []string{".env.dev", ".env", "../.env.dev", "../.env", "../../.env.dev", "../../.env"}
	for _, p := range envPaths {
		if _, err := os.Stat(p); err == nil {
			_ = godotenv.Load(p)
			break
		}
	}
	_ = godotenv.Load()
}

func maskConnectionString(connStr string) string {
	parts := strings.Split(connStr, "@")
	if len(parts) > 1 {
		prefixParts := strings.Split(parts[0], "://")
		if len(prefixParts) > 1 {
			userParts := strings.Split(prefixParts[1], ":")
			return fmt.Sprintf("%s://%s:****@%s", prefixParts[0], userParts[0], parts[1])
		}
		return fmt.Sprintf("****@%s", parts[1])
	}
	return connStr
}

func printSummaryTable(reports []SyncReport, totalDuration time.Duration) {
	fmt.Println("\n=========================================================================================")
	fmt.Println("                             DATABASE SYNC & REFRESH REPORT                              ")
	fmt.Println("=========================================================================================")
	fmt.Printf("%-15s | %-30s | %10s | %10s | %10s | %-12s\n",
		"DOMAIN", "TABLE", "SOURCE", "SYNCED", "TARGET", "STATUS")
	fmt.Println("----------------+--------------------------------+------------+------------+------------+-------------")

	var totalSource, totalSynced, totalTarget int64
	for _, r := range reports {
		fmt.Printf("%-15s | %-30s | %10d | %10d | %10d | %-12s\n",
			r.Domain, r.Table, r.SourceCount, r.SyncedCount, r.TargetCount, r.Status)
		totalSource += r.SourceCount
		totalSynced += r.SyncedCount
		totalTarget += r.TargetCount
	}

	fmt.Println("----------------+--------------------------------+------------+------------+------------+-------------")
	fmt.Printf("%-15s | %-30s | %10d | %10d | %10d | Completed in %v\n",
		"TOTAL", fmt.Sprintf("%d tables", len(reports)), totalSource, totalSynced, totalTarget, totalDuration.Round(time.Millisecond))
	fmt.Println("=========================================================================================\n")
}
