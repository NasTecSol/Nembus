package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/backuppb"
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// resolveEndpoints parses any cloud URL or gRPC address and produces both
// a valid gRPC address (host:50051) and an absolute HTTP/HTTPS base URL.
func resolveEndpoints(cloudURL string) (grpcAddr string, httpBaseURL string) {
	trimmed := strings.TrimSpace(cloudURL)
	if trimmed == "" {
		return "nembus.nashrms.com:50051", "https://nembus.nashrms.com"
	}

	var host string
	var scheme = "https"
	var customHTTPPort string

	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		if u, err := url.Parse(trimmed); err == nil && u.Hostname() != "" {
			host = u.Hostname()
			scheme = u.Scheme
			port := u.Port()
			if port != "" && port != "80" && port != "443" && port != "50051" {
				customHTTPPort = port
			}
		} else {
			host = "nembus.nashrms.com"
		}
	} else {
		parts := strings.Split(trimmed, ":")
		host = parts[0]
		if host == "localhost" || host == "127.0.0.1" || strings.HasPrefix(host, "192.168.") || strings.HasPrefix(host, "10.") {
			scheme = "http"
			if len(parts) > 1 && parts[1] != "50051" {
				customHTTPPort = parts[1]
			}
		}
	}

	if strings.Contains(trimmed, ":50051") || !strings.Contains(trimmed, ":") {
		grpcAddr = fmt.Sprintf("%s:50051", host)
	} else if strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "http") {
		grpcAddr = trimmed
	} else {
		grpcAddr = fmt.Sprintf("%s:50051", host)
	}

	if customHTTPPort != "" {
		httpBaseURL = fmt.Sprintf("%s://%s:%s", scheme, host, customHTTPPort)
	} else {
		httpBaseURL = fmt.Sprintf("%s://%s", scheme, host)
	}

	return grpcAddr, httpBaseURL
}

// SyncTenantBackupData connects to the gRPC BackupService and populates SQLite,
// streaming directly into a temporary file and falling back seamlessly to Cloud REST sync.
func SyncTenantBackupData(tenantSlug, cloudURL, sqlitePath string) map[string]interface{} {
	cleanSlug := strings.TrimSpace(tenantSlug)
	if cleanSlug == "" || strings.EqualFold(cleanSlug, "masterdb") {
		return map[string]interface{}{
			"success": false,
			"error":   "Cannot clone database: a valid tenant slug must be specified (cannot clone masterdb)",
		}
	}
	if strings.TrimSpace(sqlitePath) == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "Local SQLite database path is empty. Call InitNembusMobile before cloning.",
		}
	}

	grpcAddr, httpBaseURL := resolveEndpoints(cloudURL)
	log.Printf("🚀 [Database Clone] Starting clone for tenant '%s' (gRPC: %s, HTTP: %s)", cleanSlug, grpcAddr, httpBaseURL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Direct gRPC dial with 15s timeout so mobile networks (cellular/Wi-Fi) have time to handshake
	dialCtx, dialCancel := context.WithTimeout(ctx, 15*time.Second)
	defer dialCancel()

	conn, err := grpc.DialContext(
		dialCtx,
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Printf("⚠️ [Database Clone] gRPC connection to %s failed: %v. Falling back to HTTP sync (%s)...", grpcAddr, err, httpBaseURL)
		return syncTenantMasterDataHTTP(cleanSlug, httpBaseURL, sqlitePath)
	}
	defer conn.Close()

	// If gRPC connection was established, stream the exact tenant database to temporary file
	if conn != nil {
		client := backuppb.NewBackupServiceClient(conn)
		stream, err := client.StreamBackup(ctx, &backuppb.BackupRequest{
			TenantSlug: cleanSlug,
			AuthToken:  "pos-mobile-bootstrap-token",
			Compressed: false,
		})
		if err == nil {
			tmpDumpFile := sqlitePath + ".pgdump.tmp"
			f, createErr := os.Create(tmpDumpFile)
			if createErr == nil {
				var totalBytes uint64
				streamFailed := false

				for {
					chunk, err := stream.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Printf("⚠️ [Database Clone] Stream recv error: %v", err)
						streamFailed = true
						break
					}

					if len(chunk.Data) > 0 {
						if _, werr := f.Write(chunk.Data); werr != nil {
							log.Printf("⚠️ [Database Clone] Stream write error: %v", werr)
							streamFailed = true
							break
						}
						totalBytes += uint64(len(chunk.Data))
					}

					if chunk.IsLast {
						break
					}
				}
				_ = f.Close()

				if !streamFailed && totalBytes > 0 {
					db, err := sql.Open("sqlite3", sqlitePath)
					if err == nil {
						defer db.Close()
						insertedCount, err := applyPgDumpFileToSQLite(db, tmpDumpFile)
						_ = os.Remove(tmpDumpFile)

						if err == nil && insertedCount > 0 {
							log.Printf("🎉 [Database Clone Complete] gRPC stream successfully restored %d records into SQLite", insertedCount)
							return map[string]interface{}{
								"success": true,
								"data": map[string]interface{}{
									"tenant_slug":      cleanSlug,
									"bytes_synced":     totalBytes,
									"records_synced":   insertedCount,
									"connected_target": grpcAddr,
									"method":           "grpc_stream",
									"status":           "completed",
								},
							}
						}
					}
				}
				_ = os.Remove(tmpDumpFile)
			}
		}
	}

	// Fallback to Cloud Tenant Master Data HTTP Sync Engine
	log.Printf("ℹ️ [Database Clone] gRPC stream not completed. Falling back to HTTP sync (%s)...", httpBaseURL)
	return syncTenantMasterDataHTTP(cleanSlug, httpBaseURL, sqlitePath)
}

func applyPgDumpFileToSQLite(db *sql.DB, filePath string) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return applyPgDumpReaderToSQLite(db, f)
}

func applyPgDumpToSQLite(db *sql.DB, dumpContent string) (int, error) {
	return applyPgDumpReaderToSQLite(db, strings.NewReader(dumpContent))
}

func applyPgDumpReaderToSQLite(db *sql.DB, reader io.Reader) (int, error) {
	_, _ = db.Exec("PRAGMA foreign_keys = OFF;")

	scanner := bufio.NewScanner(reader)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 50*1024*1024)

	var inCopy bool
	var copyTable string
	var copyCols []string
	var colIndices []int
	var validCols []string
	var insertStmt *sql.Stmt
	var tx *sql.Tx
	totalInserted := 0
	tableInserted := 0

	copyRegexWithCols := regexp.MustCompile(`(?i)COPY\s+([a-zA-Z0-9_."]+)\s*\(([^)]+)\)\s+FROM\s+stdin`)
	copyRegexNoCols := regexp.MustCompile(`(?i)COPY\s+([a-zA-Z0-9_."]+)\s+FROM\s+stdin`)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if inCopy {
			if trimmed == "\\." {
				inCopy = false
				if insertStmt != nil {
					_ = insertStmt.Close()
					insertStmt = nil
				}
				if tx != nil {
					_ = tx.Commit()
					tx = nil
				}
				log.Printf("✅ [Database Clone] Cloned table '%s': %d rows inserted", copyTable, tableInserted)
				tableInserted = 0
				continue
			}

			fields := strings.Split(line, "\t")
			if len(fields) == 0 || insertStmt == nil {
				continue
			}

			// Map values to valid columns
			vals := make([]interface{}, 0, len(validCols))
			for _, idx := range colIndices {
				if idx >= len(fields) {
					vals = append(vals, nil)
					continue
				}
				f := fields[idx]
				if f == "\\N" {
					vals = append(vals, nil)
				} else if f == "t" {
					vals = append(vals, 1)
				} else if f == "f" {
					vals = append(vals, 0)
				} else {
					unescaped := strings.ReplaceAll(f, "\\\\", "\\")
					unescaped = strings.ReplaceAll(unescaped, "\\n", "\n")
					unescaped = strings.ReplaceAll(unescaped, "\\r", "\r")
					unescaped = strings.ReplaceAll(unescaped, "\\t", "\t")
					vals = append(vals, unescaped)
				}
			}

			if len(vals) == len(validCols) {
				if _, err := insertStmt.Exec(vals...); err == nil {
					totalInserted++
					tableInserted++
				} else {
					log.Printf("⚠️ [Database Clone Error] Row insert into '%s' failed: %v", copyTable, err)
				}
			}
			continue
		}

		if strings.HasPrefix(strings.ToUpper(trimmed), "COPY ") && strings.Contains(strings.ToUpper(trimmed), "FROM STDIN") {
			inCopy = true
			copyCols = nil
			colIndices = nil
			validCols = nil
			tableInserted = 0
			if insertStmt != nil {
				_ = insertStmt.Close()
				insertStmt = nil
			}
			if tx != nil {
				_ = tx.Commit()
				tx = nil
			}

			matches := copyRegexWithCols.FindStringSubmatch(trimmed)
			if len(matches) >= 3 {
				rawTable := matches[1]
				rawTable = strings.ReplaceAll(rawTable, `"public".`, "")
				rawTable = strings.ReplaceAll(rawTable, `public.`, "")
				rawTable = strings.ReplaceAll(rawTable, `"`, "")
				copyTable = strings.TrimSpace(rawTable)

				rawCols := matches[2]
				cols := strings.Split(rawCols, ",")
				for _, c := range cols {
					colName := strings.TrimSpace(c)
					colName = strings.ReplaceAll(colName, `"`, "")
					copyCols = append(copyCols, colName)
				}
			} else {
				mNoCols := copyRegexNoCols.FindStringSubmatch(trimmed)
				if len(mNoCols) >= 2 {
					rawTable := mNoCols[1]
					rawTable = strings.ReplaceAll(rawTable, `"public".`, "")
					rawTable = strings.ReplaceAll(rawTable, `public.`, "")
					rawTable = strings.ReplaceAll(rawTable, `"`, "")
					copyTable = strings.TrimSpace(rawTable)
				}
			}

			if copyTable == "" {
				continue
			}

			// Query SQLite PRAGMA table_info to get the real columns for this table
			sqliteColsMap := make(map[string]bool)
			var sqliteColsList []string
			rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(\"%s\")", copyTable))
			if err == nil {
				for rows.Next() {
					var cid int
					var name, ctype string
					var notnull, pk int
					var dfltValue sql.NullString
					_ = rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
					sqliteColsMap[strings.ToLower(name)] = true
					sqliteColsList = append(sqliteColsList, name)
				}
				rows.Close()
			}

			if len(sqliteColsMap) == 0 {
				log.Printf("ℹ️ [Database Clone] Table '%s' does not exist in SQLite, skipping copy", copyTable)
				continue
			}

			if len(copyCols) == 0 {
				copyCols = sqliteColsList
			}

			// Find intersection of Postgres dump columns and SQLite table columns
			for i, c := range copyCols {
				if sqliteColsMap[strings.ToLower(c)] {
					validCols = append(validCols, fmt.Sprintf(`"%s"`, c))
					colIndices = append(colIndices, i)
				}
			}

			if len(validCols) > 0 {
				var err error
				tx, err = db.Begin()
				if err != nil {
					log.Printf("❌ [Database Clone Error] Begin tx failed for '%s': %v", copyTable, err)
					continue
				}

				placeholders := make([]string, len(validCols))
				for i := range placeholders {
					placeholders[i] = "?"
				}
				query := fmt.Sprintf("INSERT OR REPLACE INTO \"%s\" (%s) VALUES (%s)",
					copyTable,
					strings.Join(validCols, ", "),
					strings.Join(placeholders, ", "),
				)
				stmt, err := tx.Prepare(query)
				if err == nil {
					insertStmt = stmt
				} else {
					log.Printf("❌ [Database Clone Error] Prepare failed for '%s': %v (query: %s)", copyTable, err, query)
					_ = tx.Rollback()
					tx = nil
				}
			}
			continue
		}

		if strings.HasPrefix(strings.ToUpper(trimmed), "INSERT INTO ") {
			cleanInsert := strings.ReplaceAll(trimmed, `"public".`, "")
			cleanInsert = strings.ReplaceAll(cleanInsert, `public.`, "")
			if _, err := db.Exec(cleanInsert); err == nil {
				totalInserted++
			}
			continue
		}
	}

	if tx != nil {
		_ = tx.Commit()
	}
	_, _ = db.Exec("PRAGMA foreign_keys = ON;")
	log.Printf("🎉 [Database Clone Complete] Total records restored into local SQLite: %d", totalInserted)
	return totalInserted, nil
}

// fetchCloudAuthToken retrieves a valid Bearer token from the cloud server.
func fetchCloudAuthToken(client *http.Client, baseURL string) string {
	devTokenURL := fmt.Sprintf("%s/api/dev/token", baseURL)
	req, err := http.NewRequest("GET", devTokenURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("⚠️ [Token Fetch] GET %s error: %v", devTokenURL, err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var res map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&res); err == nil {
			if token, ok := res["token"].(string); ok && token != "" {
				log.Printf("🔑 [Token Fetch] Successfully acquired Bearer token from %s", devTokenURL)
				return token
			}
		}
	}
	return ""
}

// extractRecords extracts a list of map records from diverse REST API response shapes.
func extractRecords(respData map[string]interface{}) []map[string]interface{} {
	var recs []map[string]interface{}
	dataVal, exists := respData["data"]
	if !exists || dataVal == nil {
		return recs
	}

	switch d := dataVal.(type) {
	case []interface{}:
		for _, item := range d {
			if m, ok := item.(map[string]interface{}); ok {
				recs = append(recs, m)
			}
		}
	case map[string]interface{}:
		if innerList, ok := d["data"].([]interface{}); ok {
			for _, item := range innerList {
				if m, ok := item.(map[string]interface{}); ok {
					recs = append(recs, m)
				}
			}
		} else {
			recs = append(recs, d)
		}
	}
	return recs
}

// syncTenantMasterDataHTTP syncs ALL tenant tables and master data entities directly from the cloud REST APIs.
func syncTenantMasterDataHTTP(tenantSlug, cloudURL, sqlitePath string) map[string]interface{} {
	if strings.TrimSpace(sqlitePath) == "" {
		return map[string]interface{}{"success": false, "error": "Local SQLite database path is empty"}
	}

	cleanSlug := strings.TrimSpace(tenantSlug)
	db, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		return map[string]interface{}{"success": false, "error": fmt.Sprintf("Failed to open SQLite db: %v", err)}
	}
	defer db.Close()

	_, _ = db.Exec("PRAGMA foreign_keys = OFF;")
	_, _ = db.Exec("PRAGMA synchronous = NORMAL;")
	_, _ = db.Exec("PRAGMA journal_mode = WAL;")

	_, baseURL := resolveEndpoints(cloudURL)
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	totalRecords := 0
	var firstErr error
	var errMu sync.Mutex
	recordErr := func(err error) {
		if err != nil {
			errMu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			errMu.Unlock()
		}
	}

	// 1. Fetch Cloud Auth Token
	authToken := fetchCloudAuthToken(httpClient, baseURL)

	// 2. Fetch & Store Tenant Record
	tenantURL := fmt.Sprintf("%s/api/tenants/%s", baseURL, cleanSlug)
	if tResp, err := httpGetJSON(httpClient, tenantURL, cleanSlug, authToken); err == nil {
		if tData, ok := tResp["data"].(map[string]interface{}); ok {
			inserted := insertRecords(db, "tenants", []map[string]interface{}{tData})
			totalRecords += inserted
			log.Printf("📦 [HTTP Clone] 'tenants': cloned %d records", inserted)
		}
	} else {
		recordErr(err)
	}

	// 3. Fetch Organizations and extract organization_id
	var orgID int
	orgURL := fmt.Sprintf("%s/api/organizations", baseURL)
	if oResp, err := httpGetJSON(httpClient, orgURL, cleanSlug, authToken); err == nil {
		orgRecs := extractRecords(oResp)
		if len(orgRecs) > 0 {
			inserted := insertRecords(db, "organizations", orgRecs)
			totalRecords += inserted
			log.Printf("📦 [HTTP Clone] 'organizations': cloned %d records", inserted)
			if idVal, ok := orgRecs[0]["id"].(float64); ok {
				orgID = int(idVal)
			}
		}
	} else {
		recordErr(err)
	}
	if orgID == 0 {
		orgID = 1
	}

	// 4. Comprehensive Master Data & RBAC Resource Endpoints to clone
	type syncResource struct {
		endpoint  string
		tableName string
	}

	resources := []syncResource{
		{endpoint: "/api/stores", tableName: "stores"},
		{endpoint: "/api/storage-locations", tableName: "storage_locations"},
		{endpoint: "/api/roles", tableName: "roles"},
		{endpoint: "/api/permissions", tableName: "permissions"},
		{endpoint: "/api/modules", tableName: "modules"},
		{endpoint: "/api/menus", tableName: "menus"},
		{endpoint: "/api/submenus", tableName: "submenus"},
		{endpoint: "/api/product-categories", tableName: "product_categories"},
		{endpoint: "/api/brands", tableName: "brands"},
		{endpoint: "/api/uoms", tableName: "units_of_measure"},
		{endpoint: "/api/tax-categories", tableName: "tax_categories"},
		{endpoint: "/api/price-lists", tableName: "price_lists"},
		{endpoint: fmt.Sprintf("/api/products/catalog?organization_id=%d", orgID), tableName: "products"},
		{endpoint: "/api/product-prices", tableName: "product_prices"},
		{endpoint: "/api/product-barcodes", tableName: "product_barcodes"},
		{endpoint: "/api/product-variants", tableName: "product_variants"},
		{endpoint: "/api/inventory-stock", tableName: "inventory_stock"},
		{endpoint: "/api/stock-movements", tableName: "stock_movements"},
		{endpoint: "/api/stock-counts", tableName: "stock_counts"},
		{endpoint: "/api/stock-count-lines", tableName: "stock_count_lines"},
		{endpoint: fmt.Sprintf("/api/customers?organization_id=%d", orgID), tableName: "customers"},
		{endpoint: fmt.Sprintf("/api/business-partners?organization_id=%d", orgID), tableName: "business_partners"},
		{endpoint: "/api/cashiers", tableName: "cashiers"},
		{endpoint: "/api/pos/terminals", tableName: "pos_terminals"},
		{endpoint: "/api/pos/categories", tableName: "product_categories"},
		{endpoint: "/api/promotions", tableName: "promotions"},
		{endpoint: "/api/loyalty", tableName: "loyalty_rules"},
		{endpoint: "/api/users", tableName: "users"},
	}

	type fetchResult struct {
		tableName string
		records   []map[string]interface{}
	}

	resChan := make(chan syncResource, len(resources))
	for _, r := range resources {
		resChan <- r
	}
	close(resChan)

	resultsChan := make(chan fetchResult, len(resources))
	var wg sync.WaitGroup
	workers := 6
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for res := range resChan {
				resURL := fmt.Sprintf("%s%s", baseURL, res.endpoint)
				respData, err := httpGetJSON(httpClient, resURL, cleanSlug, authToken)
				if err != nil {
					recordErr(err)
					continue
				}

				recs := extractRecords(respData)
				if len(recs) > 0 {
					resultsChan <- fetchResult{tableName: res.tableName, records: recs}
				}
			}
		}()
	}

	wg.Wait()
	close(resultsChan)

	// Batch insert results into local SQLite
	for fr := range resultsChan {
		inserted := insertRecords(db, fr.tableName, fr.records)
		totalRecords += inserted
		log.Printf("📦 [HTTP Clone] '%s': cloned %d records", fr.tableName, inserted)
	}

	// 5. User Details sync for user_roles and user_store_access
	userDetailsURL := fmt.Sprintf("%s/api/users/details", baseURL)
	if uResp, err := httpGetJSON(httpClient, userDetailsURL, cleanSlug, authToken); err == nil {
		if uList, ok := uResp["data"].([]interface{}); ok {
			for _, uItem := range uList {
				if uMap, ok := uItem.(map[string]interface{}); ok {
					userID := uMap["id"]
					if userID == nil {
						continue
					}

					// Insert user_roles
					if roles, ok := uMap["roles"].([]interface{}); ok {
						for _, r := range roles {
							if rMap, ok := r.(map[string]interface{}); ok {
								roleID := rMap["id"]
								if roleID != nil {
									totalRecords += insertRecords(db, "user_roles", []map[string]interface{}{{
										"user_id": userID,
										"role_id": roleID,
									}})
								}
							}
						}
					}

					// Insert user_store_access
					if stores, ok := uMap["stores"].([]interface{}); ok {
						for _, s := range stores {
							if sMap, ok := s.(map[string]interface{}); ok {
								storeID := sMap["id"]
								if storeID != nil {
									totalRecords += insertRecords(db, "user_store_access", []map[string]interface{}{{
										"user_id":    userID,
										"store_id":   storeID,
										"is_primary": sMap["is_primary"],
									}})
								}
							}
						}
					}
				}
			}
		}
	}

	// 6. Role Permissions Mapping Sync
	var roleIDs []int
	if rRows, err := db.Query("SELECT id FROM roles"); err == nil {
		for rRows.Next() {
			var rid int
			_ = rRows.Scan(&rid)
			roleIDs = append(roleIDs, rid)
		}
		rRows.Close()
	}

	for _, rid := range roleIDs {
		rPermURL := fmt.Sprintf("%s/api/permissions/role/%d", baseURL, rid)
		if rpResp, err := httpGetJSON(httpClient, rPermURL, cleanSlug, authToken); err == nil {
			rpList := extractRecords(rpResp)
			for _, rpMap := range rpList {
				permID := rpMap["permission_id"]
				if permID == nil {
					permID = rpMap["id"]
				}
				if permID != nil {
					totalRecords += insertRecords(db, "role_permissions", []map[string]interface{}{{
						"role_id":       rid,
						"permission_id": permID,
					}})
				}
			}
		}
	}

	// 7. Menu Permissions Mapping Sync
	var menuIDs []int
	if mRows, err := db.Query("SELECT id FROM menus"); err == nil {
		for mRows.Next() {
			var mid int
			_ = mRows.Scan(&mid)
			menuIDs = append(menuIDs, mid)
		}
		mRows.Close()
	}

	for _, mid := range menuIDs {
		mPermURL := fmt.Sprintf("%s/api/permissions/menu/%d", baseURL, mid)
		if mpResp, err := httpGetJSON(httpClient, mPermURL, cleanSlug, authToken); err == nil {
			mpList := extractRecords(mpResp)
			for _, mpMap := range mpList {
				permID := mpMap["permission_id"]
				if permID == nil {
					permID = mpMap["id"]
				}
				if permID != nil {
					totalRecords += insertRecords(db, "menu_permissions", []map[string]interface{}{{
						"menu_id":       mid,
						"permission_id": permID,
					}})
				}
			}
		}
	}

	// 8. Submenu Permissions Mapping Sync
	var submenuIDs []int
	if smRows, err := db.Query("SELECT id FROM submenus"); err == nil {
		for smRows.Next() {
			var smid int
			_ = smRows.Scan(&smid)
			submenuIDs = append(submenuIDs, smid)
		}
		smRows.Close()
	}

	for _, smid := range submenuIDs {
		smPermURL := fmt.Sprintf("%s/api/permissions/submenu/%d", baseURL, smid)
		if smpResp, err := httpGetJSON(httpClient, smPermURL, cleanSlug, authToken); err == nil {
			smpList := extractRecords(smpResp)
			for _, smpMap := range smpList {
				permID := smpMap["permission_id"]
				if permID == nil {
					permID = smpMap["id"]
				}
				if permID != nil {
					totalRecords += insertRecords(db, "submenu_permissions", []map[string]interface{}{{
						"submenu_id":    smid,
						"permission_id": permID,
					}})
				}
			}
		}
	}

	_, _ = db.Exec("PRAGMA foreign_keys = ON;")

	if totalRecords == 0 {
		errMsg := fmt.Sprintf("Database clone failed for tenant '%s': 0 records could be retrieved from %s (verify tenant slug and connection)", cleanSlug, baseURL)
		if firstErr != nil {
			errMsg = fmt.Sprintf("Database clone failed for tenant '%s': 0 records could be retrieved from %s (%v)", cleanSlug, baseURL, firstErr)
		}
		return map[string]interface{}{
			"success": false,
			"error":   errMsg,
		}
	}

	log.Printf("🎉 [Database Clone] Full master data sync completed for '%s': %d total records cloned", cleanSlug, totalRecords)

	return map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"tenant_slug":    cleanSlug,
			"records_synced": totalRecords,
			"method":         "full_master_data_sync",
			"status":         "completed",
		},
	}
}

func httpGetJSON(client *http.Client, targetURL, tenantSlug, authToken string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	cleanSlug := strings.TrimSpace(tenantSlug)
	req.Header.Set("Accept", "application/json")
	if cleanSlug != "" && !strings.EqualFold(cleanSlug, "masterdb") {
		req.Header.Set("x-tenant-id", cleanSlug)
		req.Header.Set("x-tenant-slug", cleanSlug)
		req.Header.Set("X-Tenant-ID", cleanSlug)
		req.Header.Set("X-Tenant-Slug", cleanSlug)
	}
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("⚠️ [HTTP Error] GET %s: %v", targetURL, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		statusErr := fmt.Errorf("HTTP %d for %s", resp.StatusCode, targetURL)
		log.Printf("⚠️ [HTTP Error] %v", statusErr)
		return nil, statusErr
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("⚠️ [HTTP JSON Error] GET %s: %v", targetURL, err)
		return nil, err
	}
	return result, nil
}

func insertRecords(db *sql.DB, tableName string, records []map[string]interface{}) int {
	if len(records) == 0 {
		return 0
	}

	// Query SQLite PRAGMA table_info to get the real columns for this table
	sqliteColsMap := make(map[string]bool)
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(\"%s\")", tableName))
	if err == nil {
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dfltValue sql.NullString
			_ = rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
			sqliteColsMap[strings.ToLower(name)] = true
		}
		rows.Close()
	}

	if len(sqliteColsMap) == 0 {
		return 0
	}

	tx, err := db.Begin()
	if err != nil {
		return 0
	}
	inserted := 0
	for _, rec := range records {
		var cols []string
		var placeholders []string
		var vals []interface{}

		for k, v := range rec {
			if !sqliteColsMap[strings.ToLower(k)] {
				continue
			}

			// Skip nested object arrays if they are sub-relations (e.g. variants array inside product)
			switch val := v.(type) {
			case []interface{}:
				continue
			case map[string]interface{}:
				b, _ := json.Marshal(val)
				cols = append(cols, fmt.Sprintf(`"%s"`, k))
				placeholders = append(placeholders, "?")
				vals = append(vals, string(b))
			default:
				cols = append(cols, fmt.Sprintf(`"%s"`, k))
				placeholders = append(placeholders, "?")
				vals = append(vals, val)
			}
		}

		if len(cols) == 0 {
			continue
		}

		query := fmt.Sprintf("INSERT OR REPLACE INTO \"%s\" (%s) VALUES (%s)",
			tableName,
			strings.Join(cols, ", "),
			strings.Join(placeholders, ", "),
		)

		if _, err := tx.Exec(query, vals...); err == nil {
			inserted++
		}
	}
	_ = tx.Commit()
	return inserted
}

// scanWithCoercion maps SQLite dynamic types to pgtype, sqlc, and standard Go models.
func scanWithCoercion(scanFn func(dest ...any) error, dest ...any) error {
	rawValues := make([]interface{}, len(dest))
	for i := range dest {
		var v interface{}
		rawValues[i] = &v
	}

	if err := scanFn(rawValues...); err != nil {
		return err
	}

	for i, d := range dest {
		val := *(rawValues[i].(*interface{}))
		if val == nil {
			continue
		}

		switch target := d.(type) {
		case *pgtype.Bool:
			switch v := val.(type) {
			case int64:
				target.Bool = v != 0
				target.Valid = true
			case bool:
				target.Bool = v
				target.Valid = true
			case string:
				target.Bool = v == "1" || v == "true" || v == "t"
				target.Valid = true
			default:
				_ = target.Scan(val)
			}
		case *pgtype.Text:
			switch v := val.(type) {
			case string:
				target.String = v
				target.Valid = true
			case []byte:
				target.String = string(v)
				target.Valid = true
			default:
				_ = target.Scan(val)
			}
		case *pgtype.Timestamp:
			switch v := val.(type) {
			case time.Time:
				target.Time = v
				target.Valid = true
			case string:
				if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
					target.Time = t
					target.Valid = true
				} else if t, err := time.Parse(time.RFC3339, v); err == nil {
					target.Time = t
					target.Valid = true
				} else {
					target.Valid = false
				}
			default:
				_ = target.Scan(val)
			}
		case *json.RawMessage:
			switch v := val.(type) {
			case string:
				*target = []byte(v)
			case []byte:
				*target = v
			default:
				b, _ := json.Marshal(v)
				*target = b
			}
		case *pgtype.Int4:
			switch v := val.(type) {
			case int64:
				target.Int32 = int32(v)
				target.Valid = true
			case int:
				target.Int32 = int32(v)
				target.Valid = true
			case int32:
				target.Int32 = v
				target.Valid = true
			default:
				_ = target.Scan(val)
			}
		case *pgtype.Int8:
			switch v := val.(type) {
			case int64:
				target.Int64 = v
				target.Valid = true
			case int:
				target.Int64 = int64(v)
				target.Valid = true
			default:
				_ = target.Scan(val)
			}
		case *pgtype.Numeric:
			switch v := val.(type) {
			case float64:
				_ = target.Scan(fmt.Sprintf("%f", v))
			case int64:
				_ = target.Scan(fmt.Sprintf("%d", v))
			case string:
				_ = target.Scan(v)
			case []byte:
				_ = target.Scan(string(v))
			default:
				_ = target.Scan(val)
			}
		case *int32:
			switch v := val.(type) {
			case int64:
				*target = int32(v)
			case int:
				*target = int32(v)
			}
		case *uuid.UUID:
			switch v := val.(type) {
			case string:
				if u, err := uuid.Parse(v); err == nil {
					*target = u
				}
			case []byte:
				if u, err := uuid.ParseBytes(v); err == nil {
					*target = u
				} else if u, err := uuid.FromBytes(v); err == nil {
					*target = u
				}
			default:
				if scanner, ok := d.(sql.Scanner); ok {
					_ = scanner.Scan(val)
				}
			}
		case *uuid.NullUUID:
			switch v := val.(type) {
			case string:
				if u, err := uuid.Parse(v); err == nil {
					target.UUID = u
					target.Valid = true
				}
			case []byte:
				if u, err := uuid.ParseBytes(v); err == nil {
					target.UUID = u
					target.Valid = true
				}
			default:
				if scanner, ok := d.(sql.Scanner); ok {
					_ = scanner.Scan(val)
				}
			}
		case *string:
			if s, ok := val.(string); ok {
				*target = s
			} else if b, ok := val.([]byte); ok {
				*target = string(b)
			}
		default:
			if scanner, ok := d.(sql.Scanner); ok {
				_ = scanner.Scan(val)
			}
		}
	}
	return nil
}

// sqliteRow adapts SQLite single-row scan results.
type sqliteRow struct {
	row *sql.Row
}

func (r *sqliteRow) Scan(dest ...interface{}) error {
	return scanWithCoercion(r.row.Scan, dest...)
}

// sqliteRows adapts SQLite multi-row scan results to pgx.Rows.
type sqliteRows struct {
	rows *sql.Rows
	err  error
}

func (r *sqliteRows) Close() {
	if r.rows != nil {
		r.rows.Close()
	}
}

func (r *sqliteRows) Err() error {
	if r.err != nil {
		return r.err
	}
	if r.rows != nil {
		return r.rows.Err()
	}
	return nil
}

func (r *sqliteRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (r *sqliteRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (r *sqliteRows) Next() bool {
	if r.rows != nil {
		return r.rows.Next()
	}
	return false
}

func (r *sqliteRows) Scan(dest ...any) error {
	if r.rows == nil {
		return sql.ErrNoRows
	}
	return scanWithCoercion(r.rows.Scan, dest...)
}

func (r *sqliteRows) Values() ([]any, error) {
	return nil, nil
}

func (r *sqliteRows) RawValues() [][]byte {
	return nil
}

func (r *sqliteRows) Conn() *pgx.Conn {
	return nil
}

var (
	dollarParamRegex = regexp.MustCompile(`\$(\d+)`)
	nowFuncRegex     = regexp.MustCompile(`(?i)\bNOW\(\)`)
	pgCastRegex      = regexp.MustCompile(`::[a-zA-Z0-9_]+(\[\])?`)
	ilikeRegex       = regexp.MustCompile(`(?i)\bILIKE\b`)
)

func adaptQueryForSQLite(query string) string {
	q := pgCastRegex.ReplaceAllString(query, "")
	q = nowFuncRegex.ReplaceAllString(q, "CURRENT_TIMESTAMP")
	q = ilikeRegex.ReplaceAllString(q, "LIKE")
	q = dollarParamRegex.ReplaceAllString(q, "?$1")
	return q
}

func coerceArgsForSQLite(args []interface{}) []interface{} {
	out := make([]interface{}, len(args))
	for i, arg := range args {
		if arg == nil {
			out[i] = nil
			continue
		}
		switch v := arg.(type) {
		case uuid.UUID:
			out[i] = v.String()
		case [16]byte:
			u, err := uuid.FromBytes(v[:])
			if err == nil {
				out[i] = u.String()
			} else {
				out[i] = string(v[:])
			}
		case *uuid.UUID:
			if v != nil {
				out[i] = v.String()
			} else {
				out[i] = nil
			}
		case uuid.NullUUID:
			if v.Valid {
				out[i] = v.UUID.String()
			} else {
				out[i] = nil
			}
		case pgtype.Text:
			if v.Valid {
				out[i] = v.String
			} else {
				out[i] = nil
			}
		case pgtype.Numeric:
			if v.Valid {
				if val, err := v.Value(); err == nil {
					out[i] = val
				} else {
					out[i] = nil
				}
			} else {
				out[i] = nil
			}
		case pgtype.Bool:
			if v.Valid {
				if v.Bool {
					out[i] = 1
				} else {
					out[i] = 0
				}
			} else {
				out[i] = nil
			}
		default:
			out[i] = arg
		}
	}
	return out
}

// sqliteDBTX adapts *sql.DB to repository.DBTX interface so core repository queries and handlers execute seamlessly against SQLite.
type sqliteDBTX struct {
	db *sql.DB
}

func (s *sqliteDBTX) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	cleanQuery := adaptQueryForSQLite(query)
	coercedArgs := coerceArgsForSQLite(args)
	res, err := s.db.ExecContext(ctx, cleanQuery, coercedArgs...)
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	n, _ := res.RowsAffected()
	return pgconn.NewCommandTag(fmt.Sprintf("UPDATE %d", n)), nil
}

func (s *sqliteDBTX) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	cleanQuery := adaptQueryForSQLite(query)
	coercedArgs := coerceArgsForSQLite(args)
	return &sqliteRow{row: s.db.QueryRowContext(ctx, cleanQuery, coercedArgs...)}
}

func (s *sqliteDBTX) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	cleanQuery := adaptQueryForSQLite(query)
	coercedArgs := coerceArgsForSQLite(args)
	rows, err := s.db.QueryContext(ctx, cleanQuery, coercedArgs...)
	if err != nil {
		return nil, err
	}
	return &sqliteRows{rows: rows}, nil
}

func (s *sqliteDBTX) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

// DispatchHandler forwards UI requests directly to existing packages/core/handler functions.
func DispatchHandler(handlerName, action string, payload []byte) map[string]interface{} {
	dbPathMu.RLock()
	activeDB := globalDBPath
	dbPathMu.RUnlock()

	if activeDB == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "Local SQLite database not initialized",
		}
	}

	db, err := sql.Open("sqlite3", activeDB)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to open SQLite database: %v", err),
		}
	}
	defer db.Close()

	// Initialize repository with SQLite adapter
	repo := repository.New(&sqliteDBTX{db: db})

	gin.SetMode(gin.ReleaseMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("POST", "/api/"+handlerName+"/"+action, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req.WithContext(context.WithValue(req.Context(), middleware.RepoKey, repo))

	var payloadMap map[string]interface{}
	if len(payload) > 0 {
		_ = json.Unmarshal(payload, &payloadMap)
	}

	getParam := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := payloadMap[k]; ok && v != nil {
				return fmt.Sprintf("%v", v)
			}
		}
		return ""
	}

	// Set common URL/Path parameters for Gin handlers
	if uid := getParam("user_id", "userId", "id"); uid != "" {
		c.Params = append(c.Params, gin.Param{Key: "user_id", Value: uid})
		c.Params = append(c.Params, gin.Param{Key: "id", Value: uid})
	}
	if rid := getParam("role_id", "roleId"); rid != "" {
		c.Params = append(c.Params, gin.Param{Key: "role_id", Value: rid})
	}
	if rcode := getParam("role_code", "roleCode", "code"); rcode != "" {
		c.Params = append(c.Params, gin.Param{Key: "role_code", Value: rcode})
		c.Params = append(c.Params, gin.Param{Key: "code", Value: rcode})
	}
	if mid := getParam("menu_id", "menuId"); mid != "" {
		c.Params = append(c.Params, gin.Param{Key: "menu_id", Value: mid})
	}
	if modid := getParam("module_id", "moduleId"); modid != "" {
		c.Params = append(c.Params, gin.Param{Key: "module_id", Value: modid})
	}
	if sid := getParam("store_id", "storeId"); sid != "" {
		c.Params = append(c.Params, gin.Param{Key: "store_id", Value: sid})
	}
	if cid := getParam("category_id", "categoryId"); cid != "" {
		c.Params = append(c.Params, gin.Param{Key: "category_id", Value: cid})
		if getParam("id") == "" {
			c.Params = append(c.Params, gin.Param{Key: "id", Value: cid})
		}
	}
	if plid := getParam("price_list_id", "priceListId"); plid != "" {
		c.Params = append(c.Params, gin.Param{Key: "price_list_id", Value: plid})
		c.Params = append(c.Params, gin.Param{Key: "id", Value: plid})
	}
	if pid := getParam("product_id", "productId"); pid != "" {
		c.Params = append(c.Params, gin.Param{Key: "product_id", Value: pid})
	}
	if orgid := getParam("organization_id", "org_id", "orgId"); orgid != "" {
		c.Params = append(c.Params, gin.Param{Key: "organization_id", Value: orgid})
		c.Params = append(c.Params, gin.Param{Key: "org_id", Value: orgid})
	}
	if custid := getParam("customer_id", "customerId"); custid != "" {
		c.Params = append(c.Params, gin.Param{Key: "customer_id", Value: custid})
	}
	if cartId := getParam("cart_id", "cartId"); cartId != "" {
		c.Params = append(c.Params, gin.Param{Key: "cart_id", Value: cartId})
		c.Params = append(c.Params, gin.Param{Key: "id", Value: cartId})
	}
	if itemId := getParam("item_id", "itemId"); itemId != "" {
		c.Params = append(c.Params, gin.Param{Key: "item_id", Value: itemId})
		if getParam("id") == "" {
			c.Params = append(c.Params, gin.Param{Key: "id", Value: itemId})
		}
	}
	if modifId := getParam("modifier_id", "modifierId"); modifId != "" {
		c.Params = append(c.Params, gin.Param{Key: "modifier_id", Value: modifId})
		if getParam("id") == "" {
			c.Params = append(c.Params, gin.Param{Key: "id", Value: modifId})
		}
	}
	if orderId := getParam("order_id", "orderId"); orderId != "" {
		c.Params = append(c.Params, gin.Param{Key: "order_id", Value: orderId})
		c.Params = append(c.Params, gin.Param{Key: "id", Value: orderId})
	}
	if orderNum := getParam("order_number", "orderNumber"); orderNum != "" {
		c.Params = append(c.Params, gin.Param{Key: "order_number", Value: orderNum})
	}
	if csid := getParam("cashier_session_id", "cashierSessionId", "session_id", "sessionId"); csid != "" {
		c.Params = append(c.Params, gin.Param{Key: "cashier_session_id", Value: csid})
		c.Params = append(c.Params, gin.Param{Key: "session_id", Value: csid})
		c.Params = append(c.Params, gin.Param{Key: "id", Value: csid})
	}
	if cashId := getParam("cashier_id", "cashierId"); cashId != "" {
		c.Params = append(c.Params, gin.Param{Key: "cashier_id", Value: cashId})
	}
	if posTermId := getParam("pos_terminal_id", "posTerminalId"); posTermId != "" {
		c.Params = append(c.Params, gin.Param{Key: "pos_terminal_id", Value: posTermId})
	}
	if countId := getParam("stock_count_id", "stockCountId", "count_id", "countId"); countId != "" {
		c.Params = append(c.Params, gin.Param{Key: "stock_count_id", Value: countId})
		c.Params = append(c.Params, gin.Param{Key: "id", Value: countId})
	}
	if lineId := getParam("stock_count_line_id", "line_id", "lineId"); lineId != "" {
		c.Params = append(c.Params, gin.Param{Key: "line_id", Value: lineId})
	}
	if tid := getParam("table_id", "tableId"); tid != "" {
		c.Params = append(c.Params, gin.Param{Key: "table_id", Value: tid})
		if getParam("id") == "" {
			c.Params = append(c.Params, gin.Param{Key: "id", Value: tid})
		}
	}
	if oiId := getParam("order_item_id", "orderItemId"); oiId != "" {
		c.Params = append(c.Params, gin.Param{Key: "order_item_id", Value: oiId})
		if getParam("id") == "" {
			c.Params = append(c.Params, gin.Param{Key: "id", Value: oiId})
		}
	}
	if recId := getParam("recipe_id", "recipeId"); recId != "" {
		c.Params = append(c.Params, gin.Param{Key: "recipe_id", Value: recId})
		if getParam("id") == "" {
			c.Params = append(c.Params, gin.Param{Key: "id", Value: recId})
		}
	}
	if riId := getParam("recipe_ingredient_id", "recipeIngredientId"); riId != "" {
		c.Params = append(c.Params, gin.Param{Key: "recipe_ingredient_id", Value: riId})
		if getParam("id") == "" {
			c.Params = append(c.Params, gin.Param{Key: "id", Value: riId})
		}
	}
	if wId := getParam("waste_log_id", "wasteLogId"); wId != "" {
		c.Params = append(c.Params, gin.Param{Key: "waste_log_id", Value: wId})
		if getParam("id") == "" {
			c.Params = append(c.Params, gin.Param{Key: "id", Value: wId})
		}
	}

	// Forward query string parameters from payloadMap
	q := req.URL.Query()
	for k, v := range payloadMap {
		if v != nil {
			q.Set(k, fmt.Sprintf("%v", v))
		}
	}
	req.URL.RawQuery = q.Encode()

	switch handlerName {
	case "auth":
		authHandler := handler.NewAuthHandler(usecase.NewAuthUseCase())
		switch action {
		case "login":
			authHandler.Login(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on auth handler", action),
			}
		}

	case "permission":
		permHandler := handler.NewPermissionHandler(usecase.NewPermissionUseCase())
		switch action {
		case "getUserAccessibleMenus":
			log.Printf("[Nembus Core Bridge] Executing PermissionHandler.GetUserAccessibleMenus for user_id=%s", getParam("user_id", "userId", "id"))
			permHandler.GetUserAccessibleMenus(c)
		case "getUserAccessibleModules":
			permHandler.GetUserAccessibleModules(c)
		case "getUserAccessibleSubmenus":
			log.Printf("[Nembus Core Bridge] Executing PermissionHandler.GetUserAccessibleSubmenus for user_id=%s", getParam("user_id", "userId", "id"))
			permHandler.GetUserAccessibleSubmenus(c)
		case "getUserPermissions":
			permHandler.GetUserPermissions(c)
		case "getMenuPermissions":
			permHandler.GetMenuPermissions(c)
		case "getRolePermissionsWithScope":
			permHandler.GetRolePermissionsWithScope(c)
		case "diagnoseUserRBAC":
			diagnostics := diagnoseUserRBAC(db, getParam("user_id", "userId", "id"))
			c.JSON(200, gin.H{
				"success": true,
				"message": "RBAC diagnostics completed",
				"data":    diagnostics,
			})
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on permission handler", action),
			}
		}

	case "navigation":
		navHandler := handler.NewNavigationHandler(usecase.NewNavigationUseCase(), usecase.NewRoleUseCase(), usecase.NewUserUseCase())
		switch action {
		case "getUserNavigation":
			navHandler.GetUserNavigation(c)
		case "getNavigationByRoleCode":
			navHandler.GetNavigationByRoleCodeWithUserCounts(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on navigation handler", action),
			}
		}

	case "role":
		roleHandler := handler.NewRoleHandler(usecase.NewRoleUseCase())
		switch action {
		case "getRole":
			roleHandler.GetRole(c)
		case "getRoleByCode":
			roleHandler.GetRoleByCode(c)
		case "listRoles":
			roleHandler.ListRoles(c)
		case "getRolePermissions":
			roleHandler.GetRolePermissions(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on role handler", action),
			}
		}

	case "menu":
		menuHandler := handler.NewMenuHandler(usecase.NewMenuUseCase())
		switch action {
		case "getMenu":
			menuHandler.GetMenu(c)
		case "listMenus":
			menuHandler.ListMenus(c)
		case "listMenusByModule":
			menuHandler.ListMenusByModule(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on menu handler", action),
			}
		}

	case "submenu":
		submenuHandler := handler.NewSubmenuHandler(usecase.NewSubmenuUseCase())
		switch action {
		case "listSubmenusByMenu":
			submenuHandler.ListSubmenusByMenu(c)
		case "listActiveSubmenusByMenu":
			submenuHandler.ListActiveSubmenusByMenu(c)
		case "getSubmenu":
			submenuHandler.GetSubmenu(c)
		case "getSubmenuByCode":
			submenuHandler.GetSubmenuByCode(c)
		case "listSubmenus":
			submenuHandler.ListSubmenus(c)
		case "listSubmenusByParent":
			submenuHandler.ListSubmenusByParent(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on submenu handler", action),
			}
		}

	case "customer", "customers":
		custHandler := handler.NewCustomerHandler(usecase.NewCustomerUseCase())
		switch action {
		case "listCustomers", "listAllCustomers":
			custHandler.ListCustomers(c)
		case "listActiveCustomers":
			custHandler.ListActiveCustomers(c)
		case "getCustomer", "getCustomerByID":
			custHandler.GetCustomerByID(c)
		case "getCustomerByCode":
			custHandler.GetCustomerByCode(c)
		case "searchCustomers":
			custHandler.SearchCustomers(c)
		case "createCustomer":
			custHandler.CreateCustomer(c)
		case "updateCustomer":
			custHandler.UpdateCustomer(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on customer handler", action),
			}
		}

	case "cart", "carts":
		cartHandler := handler.NewCartHandler(usecase.NewCartUseCase())
		switch action {
		case "createCart":
			cartHandler.CreateCart(c)
		case "createNewCart":
			cartHandler.CreateNewCart(c)
		case "getCart":
			cartHandler.GetCart(c)
		case "getCartByNumber":
			cartHandler.GetCartByNumber(c)
		case "getActiveCartByCustomer":
			cartHandler.GetActiveCartByCustomer(c)
		case "addToCart", "addItemToCart":
			cartHandler.AddToCart(c)
		case "listCartItems":
			cartHandler.ListCartItems(c)
		case "getCartItem":
			cartHandler.GetCartItem(c)
		case "updateCartItemQuantity":
			cartHandler.UpdateCartItemQuantity(c)
		case "deleteCartItem", "removeItemFromCart":
			cartHandler.DeleteCartItem(c)
		case "clearCartItems", "clearCart":
			cartHandler.ClearCartItems(c)
		case "convertToOrder", "convertCartToOrder":
			cartHandler.ConvertToOrder(c)
		case "getCartTotals":
			cartHandler.GetCartTotals(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on cart handler", action),
			}
		}

	case "order", "orders", "sales_order", "sales_orders":
		orderHandler := handler.NewOrderHandler(usecase.NewOrderUseCase())
		switch action {
		case "createOrder", "createSalesOrder":
			orderHandler.CreateOrder(c)
		case "getOrder", "getSalesOrder":
			orderHandler.GetOrder(c)
		case "listOrders", "listSalesOrders":
			orderHandler.ListOrders(c)
		case "getOrderByNumber":
			orderHandler.GetOrderByNumber(c)
		case "updateOrderStatus":
			orderHandler.UpdateOrderStatus(c)
		case "updateOrderPaymentStatus", "updatePaymentStatus":
			orderHandler.UpdateOrderPaymentStatus(c)
		case "updateOrderFulfillmentStatus", "updateFulfillmentStatus":
			orderHandler.UpdateOrderFulfillmentStatus(c)
		case "createOrderLine", "addOrderLine":
			orderHandler.CreateOrderLine(c)
		case "listOrderLines":
			orderHandler.ListOrderLines(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on order handler", action),
			}
		}

	case "pos", "point_of_sale":
		posHandler := handler.NewPosHandler(usecase.NewPosUseCase(), usecase.NewPosPaymentUseCase())
		switch action {
		case "listProducts":
			posHandler.ListProducts(c)
		case "getProductsByCategory":
			posHandler.GetProductsByCategory(c)
		case "searchProduct":
			posHandler.SearchProduct(c)
		case "getCategories":
			posHandler.GetCategories(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on pos handler", action),
			}
		}

	case "stores":
		storeHandler := handler.NewStoreHandler(usecase.NewStoreUseCase())
		switch action {
		case "getStore":
			storeHandler.GetStore(c)
		case "listStores":
			storeHandler.ListStores(c)
		case "listPOSEnabledStores":
			storeHandler.ListPOSEnabledStores(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on stores handler", action),
			}
		}

	case "pos_terminals", "terminals", "pos_terminal":
		terminalHandler := handler.NewPosTerminalsHandler(usecase.NewPosTerminalsUseCase())
		switch action {
		case "listPOSTerminalsByStore", "listTerminalsByStore":
			terminalHandler.ListPOSTerminalsByStore(c)
		case "listActivePOSTerminalsByStore":
			terminalHandler.ListActivePOSTerminalsByStore(c)
		case "listPOSTerminals", "listTerminals":
			terminalHandler.ListPOSTerminals(c)
		case "getPOSTerminal", "getTerminal":
			terminalHandler.GetPOSTerminal(c)
		case "createPOSTerminal", "createTerminal":
			terminalHandler.CreatePOSTerminal(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on pos_terminals handler", action),
			}
		}

	case "inventory_stock":
		invHandler := handler.NewInventoryStockHandler(usecase.NewInventoryStockUseCase())
		switch action {
		case "listInventoryStockByStore":
			invHandler.ListInventoryStockByStore(c)
		case "getInventoryStock":
			invHandler.GetInventoryStock(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on inventory_stock handler", action),
			}
		}

	case "stock_counts", "stock-counts", "stock_count":
		stockCountHandler := handler.NewStockCountsHandler(usecase.NewStockCountsUseCase())
		switch action {
		case "listStockCounts", "list":
			stockCountHandler.ListStockCounts(c)
		case "getStockCount", "get":
			stockCountHandler.GetStockCount(c)
		case "createStockCount", "create":
			stockCountHandler.CreateStockCount(c)
		case "updateStockCount", "update":
			stockCountHandler.UpdateStockCount(c)
		case "deleteStockCount", "delete":
			stockCountHandler.DeleteStockCount(c)
		case "startStockCount", "start":
			stockCountHandler.StartStockCount(c)
		case "completeStockCount", "complete":
			stockCountHandler.CompleteStockCount(c)
		case "approveStockCount", "approve":
			stockCountHandler.ApproveStockCount(c)
		case "reconcileStockCount", "reconcile":
			stockCountHandler.ReconcileStockCount(c)
		case "getStockCountSummary", "summary":
			stockCountHandler.GetStockCountSummary(c)
		case "listStockCountLines", "lines":
			stockCountHandler.ListStockCountLines(c)
		case "addStockCountLine", "addLine":
			stockCountHandler.AddStockCountLine(c)
		case "updateStockCountLine", "updateLine":
			stockCountHandler.UpdateStockCountLine(c)
		case "deleteStockCountLine", "deleteLine":
			stockCountHandler.DeleteStockCountLine(c)
		case "bulkUpdateStockCountLines", "bulkUpdateLines":
			stockCountHandler.BulkUpdateStockCountLines(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on stock_counts handler", action),
			}
		}

	case "user":
		userHandler := handler.NewUserHandler(usecase.NewUserUseCase())
		switch action {
		case "getUser":
			userHandler.GetUser(c)
		case "getUserPrimaryStore":
			userHandler.GetUserPrimaryStore(c)
		case "getUserStores":
			userHandler.GetUserStores(c)
		case "getUsersByRole":
			userHandler.GetUsersByRole(c)
		case "getUserWithDetails":
			userHandler.GetUserWithDetails(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on user handler", action),
			}
		}

	case "price_lists":
		priceListsHandler := handler.NewPriceListsHandler(usecase.NewPriceListsUseCase())
		switch action {
		case "getDefaultPriceList":
			priceListsHandler.GetDefaultPriceList(c)
		case "listPriceLists":
			priceListsHandler.ListPriceLists(c)
		case "listActivePriceLists":
			priceListsHandler.ListActivePriceLists(c)
		case "getPriceList":
			priceListsHandler.GetPriceList(c)
		default:
			return map[string]interface{}{"success": false, "error": fmt.Sprintf("Action '%s' not supported on price_lists handler", action)}
		}

	case "product_pricing":
		pricingHandler := handler.NewProductPricingHandler(usecase.NewProductPricingUseCase())
		switch action {
		case "listPricesByPriceList":
			pricingHandler.ListPricesByPriceList(c)
		case "listProductPrices":
			pricingHandler.ListProductPrices(c)
		case "getProductPrice":
			pricingHandler.GetProductPrice(c)
		case "getProductWithPricing":
			pricingHandler.GetProductWithPricing(c)
		case "getEffectivePrice":
			pricingHandler.GetEffectivePrice(c)
		case "searchProductsWithPrices":
			pricingHandler.SearchProductsWithPrices(c)
		default:
			return map[string]interface{}{"success": false, "error": fmt.Sprintf("Action '%s' not supported on product_pricing handler", action)}
		}

	case "product_catalog":
		catalogHandler := handler.NewProductCatalogHandler(usecase.NewProductCatalogUseCase())
		switch action {
		case "listProductsWithVariants":
			catalogHandler.ListProductsWithVariants(c)
		case "getMasterProductCatalog":
			catalogHandler.GetMasterProductCatalog(c)
		default:
			return map[string]interface{}{"success": false, "error": fmt.Sprintf("Action '%s' not supported on product_catalog handler", action)}
		}

	case "sync", "outbox":
		tenantSlug := "qitaf"
		cloudURL := "nembus.nashrms.com:50051"
		var storeID int32 = 1

		var syncReq struct {
			TenantSlug    string                 `json:"tenant_slug"`
			CloudURL      string                 `json:"cloud_url"`
			StoreID       int32                  `json:"store_id"`
			EntityType    string                 `json:"entity_type"`
			EntityID      string                 `json:"entity_id"`
			Action        string                 `json:"action"`
			Payload       map[string]interface{} `json:"payload"`
			Priority      int                    `json:"priority"`
			CorrelationID string                 `json:"correlation_id"`
		}
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &syncReq)
			if syncReq.TenantSlug != "" {
				tenantSlug = syncReq.TenantSlug
			}
			if syncReq.CloudURL != "" {
				cloudURL = syncReq.CloudURL
			}
			if syncReq.StoreID > 0 {
				storeID = syncReq.StoreID
			}
		}

		syncSvc := NewSyncService(context.Background(), db, cloudURL, tenantSlug, storeID)
		switch action {
		case "syncNow", "performFullSync":
			return syncSvc.PerformFullSync()
		case "drainOutbox", "pushOutbox":
			return syncSvc.DrainOutboxGRPC()
		case "fetchDelta", "pullDelta":
			return syncSvc.FetchDeltaGRPC()
		case "getSyncStatus", "getStatus":
			return map[string]interface{}{
				"success": true,
				"data":    syncSvc.GetSyncQueueStatus(),
			}
		case "enqueueOutbox", "enqueue":
			err := EnqueueOutboxRecord(db, syncReq.EntityType, syncReq.EntityID, syncReq.Action, syncReq.Payload, syncReq.Priority, syncReq.CorrelationID)
			return map[string]interface{}{
				"success": err == nil,
				"error":   fmt.Sprintf("%v", err),
			}
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on sync handler", action),
			}
		}

	case "cashier_session", "cashier_sessions":
		cashierSessionHandler := handler.NewCashierSessionHandler(usecase.NewCashierSessionUseCase())
		switch action {
		case "open", "openSession", "openCashierSession":
			cashierSessionHandler.OpenCashierSession(c)
		case "close", "closeSession", "closeCashierSession":
			cashierSessionHandler.CloseCashierSession(c)
		case "getActive", "getActiveSession", "getActiveCashierSession":
			cashierSessionHandler.GetActiveCashierSession(c)
		case "get", "getSession", "getSessionByID":
			cashierSessionHandler.GetSessionByID(c)
		case "summary", "getSessionSummary":
			cashierSessionHandler.GetSessionSummary(c)
		case "list", "getCashierSessions", "listSessions":
			cashierSessionHandler.GetCashierSessions(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on cashier_session handler", action),
			}
		}

	case "cashier", "cashiers":
		cashierHandler := handler.NewCashierHandler(usecase.NewCashierUseCase())
		switch action {
		case "createCashier":
			cashierHandler.CreateCashier(c)
		case "createCashierWithDefaults":
			cashierHandler.CreateCashierWithDefaults(c)
		case "getCashierByID":
			cashierHandler.GetCashierByID(c)
		case "getCashierByCode":
			cashierHandler.GetCashierByCode(c)
		case "getCashierByUserID":
			cashierHandler.GetCashierByUserID(c)
		case "listAllCashiers":
			cashierHandler.ListAllCashiers(c)
		case "listActiveCashiers":
			cashierHandler.ListActiveCashiers(c)
		case "listCashiersByStore":
			cashierHandler.ListCashiersByStore(c)
		case "listActiveCashiersByStore":
			cashierHandler.ListActiveCashiersByStore(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on cashier handler", action),
			}
		}

	case "restaurant", "cafe":
		restaurantHandler := handler.NewRestaurantHandler(usecase.NewRestaurantUseCase())
		switch action {
		case "listTables":
			restaurantHandler.ListTables(c)
		case "getTable":
			restaurantHandler.GetTable(c)
		case "createTable":
			restaurantHandler.CreateTable(c)
		case "updateTable":
			restaurantHandler.UpdateTable(c)
		case "deleteTable":
			restaurantHandler.DeleteTable(c)
		case "listMenuCategories":
			restaurantHandler.ListMenuCategories(c)
		case "getMenuCategory":
			restaurantHandler.GetMenuCategory(c)
		case "createMenuCategory":
			restaurantHandler.CreateMenuCategory(c)
		case "updateMenuCategory":
			restaurantHandler.UpdateMenuCategory(c)
		case "deleteMenuCategory":
			restaurantHandler.DeleteMenuCategory(c)
		case "listMenuItems":
			restaurantHandler.ListMenuItems(c)
		case "getMenuItem":
			restaurantHandler.GetMenuItem(c)
		case "getMenuItemFullDetails":
			restaurantHandler.GetMenuItemFullDetails(c)
		case "createMenuItem":
			restaurantHandler.CreateMenuItem(c)
		case "updateMenuItem":
			restaurantHandler.UpdateMenuItem(c)
		case "deleteMenuItem":
			restaurantHandler.DeleteMenuItem(c)
		case "listComboComponents":
			restaurantHandler.ListComboComponents(c)
		case "listModifiers":
			restaurantHandler.ListModifiers(c)
		case "listStoreModifiers":
			restaurantHandler.ListStoreModifiers(c)
		case "getModifier":
			restaurantHandler.GetModifier(c)
		case "createModifier":
			restaurantHandler.CreateModifier(c)
		case "updateModifier":
			restaurantHandler.UpdateModifier(c)
		case "deleteModifier":
			restaurantHandler.DeleteModifier(c)
		case "getFullMenu", "getMenu":
			restaurantHandler.GetFullMenu(c)
		case "getKdsOrders":
			restaurantHandler.GetKdsOrders(c)
		case "createOrder":
			restaurantHandler.CreateOrder(c)
		case "updateOrder":
			restaurantHandler.UpdateOrder(c)
		case "deleteOrder":
			restaurantHandler.DeleteOrder(c)
		case "listOrders", "listBills":
			restaurantHandler.ListOrders(c)
		case "listStoreOrders":
			restaurantHandler.ListStoreOrders(c)
		case "getOrder":
			restaurantHandler.GetOrder(c)
		case "updateOrderStatus":
			restaurantHandler.UpdateOrderStatus(c)
		case "settleOrder":
			restaurantHandler.SettleOrder(c)
		case "createOnlineOrder":
			restaurantHandler.CreateOnlineOrder(c)
		case "getOrderItem":
			restaurantHandler.GetOrderItem(c)
		case "updateOrderItem":
			restaurantHandler.UpdateOrderItem(c)
		case "deleteOrderItem":
			restaurantHandler.DeleteOrderItem(c)
		case "createMenuModifierGroup":
			restaurantHandler.CreateMenuModifierGroup(c)
		case "getMenuModifierGroup":
			restaurantHandler.GetMenuModifierGroup(c)
		case "listMenuModifierGroupsByStore":
			restaurantHandler.ListMenuModifierGroupsByStore(c)
		case "updateMenuModifierGroup":
			restaurantHandler.UpdateMenuModifierGroup(c)
		case "deleteMenuModifierGroup":
			restaurantHandler.DeleteMenuModifierGroup(c)
		case "createRestaurantPromotion":
			restaurantHandler.CreateRestaurantPromotion(c)
		case "getRestaurantPromotion":
			restaurantHandler.GetRestaurantPromotion(c)
		case "getRestaurantPromotionByCode":
			restaurantHandler.GetRestaurantPromotionByCode(c)
		case "listActiveRestaurantPromotions":
			restaurantHandler.ListActiveRestaurantPromotions(c)
		case "listAllRestaurantPromotions":
			restaurantHandler.ListAllRestaurantPromotions(c)
		case "updateRestaurantPromotion":
			restaurantHandler.UpdateRestaurantPromotion(c)
		case "updateRestaurantPromotionStatus":
			restaurantHandler.UpdateRestaurantPromotionStatus(c)
		case "deleteRestaurantPromotion":
			restaurantHandler.DeleteRestaurantPromotion(c)
		case "createRecipe":
			restaurantHandler.CreateRecipe(c)
		case "getRecipe":
			restaurantHandler.GetRecipe(c)
		case "updateRecipe":
			restaurantHandler.UpdateRecipe(c)
		case "deleteRecipe":
			restaurantHandler.DeleteRecipe(c)
		case "addRecipeIngredient":
			restaurantHandler.AddRecipeIngredient(c)
		case "listRecipes":
			restaurantHandler.ListRecipes(c)
		case "getRecipeIngredient":
			restaurantHandler.GetRecipeIngredient(c)
		case "updateRecipeIngredient":
			restaurantHandler.UpdateRecipeIngredient(c)
		case "deleteRecipeIngredient":
			restaurantHandler.DeleteRecipeIngredient(c)
		case "createWasteLog":
			restaurantHandler.CreateWasteLog(c)
		case "getWasteLog":
			restaurantHandler.GetWasteLog(c)
		case "updateWasteLog":
			restaurantHandler.UpdateWasteLog(c)
		case "deleteWasteLog":
			restaurantHandler.DeleteWasteLog(c)
		case "getWasteReport":
			restaurantHandler.GetWasteReport(c)
		case "createKioskSession":
			restaurantHandler.CreateKioskSession(c)
		case "updateKioskSession":
			restaurantHandler.UpdateKioskSession(c)
		case "deleteKioskSession":
			restaurantHandler.DeleteKioskSession(c)
		case "getKioskSession":
			restaurantHandler.GetKioskSession(c)
		case "getKioskSessionByID":
			restaurantHandler.GetKioskSessionByID(c)
		default:
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Action '%s' not supported on restaurant handler", action),
			}
		}

	default:
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Handler '%s' not registered in Go core engine", handlerName),
		}
	}

	var res map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		return map[string]interface{}{
			"success": w.Code >= 200 && w.Code < 300,
			"raw":     w.Body.String(),
		}
	}

	res["success"] = w.Code >= 200 && w.Code < 300
	return res
}

// diagnoseUserRBAC performs a comprehensive SQLite check to see why menus/submenus are present or missing.
func diagnoseUserRBAC(db *sql.DB, userIDStr string) map[string]interface{} {
	diag := make(map[string]interface{})
	diag["user_id"] = userIDStr

	var userCount, roleCount, permCount, menuCount, submenuCount, userRoleCount, rolePermCount, menuPermCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM roles").Scan(&roleCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM permissions").Scan(&permCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM menus").Scan(&menuCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM submenus").Scan(&submenuCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM user_roles").Scan(&userRoleCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM role_permissions").Scan(&rolePermCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM menu_permissions").Scan(&menuPermCount)

	diag["total_users"] = userCount
	diag["total_roles"] = roleCount
	diag["total_permissions"] = permCount
	diag["total_menus"] = menuCount
	diag["total_submenus"] = submenuCount
	diag["total_user_roles"] = userRoleCount
	diag["total_role_permissions"] = rolePermCount
	diag["total_menu_permissions"] = menuPermCount

	if userIDStr != "" {
		rows, err := db.Query("SELECT ur.role_id, r.name, r.code FROM user_roles ur LEFT JOIN roles r ON ur.role_id = r.id WHERE ur.user_id = ?", userIDStr)
		if err == nil {
			var roles []map[string]interface{}
			for rows.Next() {
				var rid int
				var rname, rcode sql.NullString
				_ = rows.Scan(&rid, &rname, &rcode)
				roles = append(roles, map[string]interface{}{
					"role_id":   rid,
					"role_name": rname.String,
					"role_code": rcode.String,
				})
			}
			rows.Close()
			diag["user_assigned_roles"] = roles
		}

		var directMenuCount int
		_ = db.QueryRow(`
			SELECT COUNT(DISTINCT mn.id) FROM menus mn
			INNER JOIN menu_permissions mnp ON mn.id = mnp.menu_id
			INNER JOIN role_permissions rp ON mnp.permission_id = rp.permission_id
			INNER JOIN user_roles ur ON rp.role_id = ur.role_id
			WHERE ur.user_id = ? AND mn.is_active = 1
		`, userIDStr).Scan(&directMenuCount)
		diag["accessible_menus_count"] = directMenuCount

		var directSubmenuCount int
		_ = db.QueryRow(`
			SELECT COUNT(DISTINCT sm.id) FROM submenus sm
			INNER JOIN submenu_permissions smp ON sm.id = smp.submenu_id
			INNER JOIN role_permissions rp ON smp.permission_id = rp.permission_id
			INNER JOIN user_roles ur ON rp.role_id = ur.role_id
			WHERE ur.user_id = ? AND sm.is_active = 1
		`, userIDStr).Scan(&directSubmenuCount)
		diag["accessible_submenus_count"] = directSubmenuCount
	}

	log.Printf("[Nembus RBAC Diagnostics] user_id=%s users=%d roles=%d perms=%d menus=%d submenus=%d user_roles=%d role_perms=%d menu_perms=%d accessible_menus=%v",
		userIDStr, userCount, roleCount, permCount, menuCount, submenuCount, userRoleCount, rolePermCount, menuPermCount, diag["accessible_menus_count"])

	return diag
}
