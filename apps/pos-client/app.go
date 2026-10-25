package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	clientbackup "github.com/NasTecSol/nembus-client/client"
	"github.com/NasTecSol/nembus-client/internal/config"
	"github.com/NasTecSol/nembus-client/internal/db"
	"github.com/NasTecSol/nembus-client/internal/payment/geidea"
	"github.com/NasTecSol/nembus-client/internal/printing"
	"github.com/NasTecSol/nembus-client/internal/sync"
	"github.com/NasTecSol/nembus-client/internal/updater"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/middleware/manager"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed migrations/*.sql
var migrations embed.FS

// App struct
type App struct {
	ctx           context.Context
	dbManager     *db.DBManager
	masterRepo    *repository.Queries
	masterPool    *pgxpool.Pool
	tenantManager *manager.Manager
	syncService   *sync.SyncService
	geideaClient  *geidea.Client
	cfg           *config.Config
	version       string
	ghToken       string
	ghRepo        string
}

// NewApp creates a new App application struct
func NewApp(version, ghToken, ghRepo string) *App {
	return &App{
		version: version,
		ghToken: ghToken,
		ghRepo:  ghRepo,
	}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Clean up legacy executable backups from previous update cycles
	updater.CleanupOldExecutables()

	if a.IsAppSetup() {
		// Auto-start DB if setup is already complete
		// Using default credentials as these are local to the machine
		go func() {
			log.Println("Auto-starting local database and backend server...")
			if res := a.StartDatabase("postgres", "password", "nembus", 5433); res != "Success" {
				log.Printf("[ERROR] Failed to start database and backend server: %s", res)
			} else {
				log.Printf("[SUCCESS] Database and backend Gin server started on port %s", a.cfg.Port)
			}
		}()
	} else {
		log.Println("[INFO] App setup is not completed yet. Setup wizard is required before backend API starts.")
	}
}

// GetAppVersion returns current application version
func (a *App) GetAppVersion() string {
	if a.version == "" {
		return "v1.0.0"
	}
	return a.version
}

// CheckForUpdates queries GitHub releases to detect newer versions.
// In development mode (wails dev), the version is the hardcoded default "v1.0.0"
// because -ldflags are not injected, which would always trigger a false-positive
// update notification. Skip the check entirely in that case.
func (a *App) CheckForUpdates() (*updater.UpdateInfo, error) {
	if a.version == "" {
		log.Println("[UPDATER] Skipping update check — running in development mode (default version)")
		return &updater.UpdateInfo{
			HasUpdate:      false,
			CurrentVersion: a.GetAppVersion(),
		}, nil
	}

	token := a.ghToken
	if token == "" && a.cfg != nil {
		token = a.cfg.GithubToken
	}
	repo := a.ghRepo
	if repo == "" && a.cfg != nil {
		repo = a.cfg.GithubRepo
	}
	return updater.CheckForUpdate(a.GetAppVersion(), repo, token)
}

// ApplyUpdate downloads release asset and restarts application
func (a *App) ApplyUpdate(downloadURL string, assetID int64) error {
	token := a.ghToken
	if token == "" && a.cfg != nil {
		token = a.cfg.GithubToken
	}
	repo := a.ghRepo
	if repo == "" && a.cfg != nil {
		repo = a.cfg.GithubRepo
	}
	return updater.ApplyUpdate(downloadURL, assetID, repo, token)
}

// domReady is called after front-end resources have been loaded
func (a App) domReady(ctx context.Context) {
}

// beforeClose is called when the application is about to quit
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	if a.dbManager != nil {
		_ = a.dbManager.Stop()
	}
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	if a.syncService != nil {
		a.syncService.Stop()
	}
	if a.dbManager != nil {
		_ = a.dbManager.Stop()
	}
}

// StartDatabase starts the embedded database with the given configuration and then the backend
func (a *App) StartDatabase(username, password, database string, port uint32) string {
	a.cfg = config.LoadConfig("development") // Default or tailored
	a.cfg.MasterDBURL = fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable",
		username, password, port, database)

	dbCfg := db.Config{
		Username: username,
		Password: password,
		Database: database,
		Port:     port,
	}
	a.dbManager = db.NewDBManager(dbCfg)
	if err := a.dbManager.Start(); err != nil {
		return fmt.Sprintf("Error starting database: %v", err)
	}

	// Initialize master repo
	ctx := context.Background()

	// Run migrations before connecting with pgxpool. Preserve local databases
	// on failure: a storage or migration error must never trigger a data reset.
	if err := a.migrate(a.cfg.MasterDBURL); err != nil {
		return fmt.Sprintf("Error running migrations (local data preserved): %v", db.WithStorageHint(err))
	}

	pool, repo, err := setupDatabase(ctx, a.cfg)
	if err != nil {
		return fmt.Sprintf("Error connecting to DB: %v", err)
	}
	a.masterPool = pool
	a.masterRepo = repo

	// Now start the Gin server in background
	go a.runBackend(pool)

	// Resume sync if needed
	a.InitializeSync()

	return "Success"
}

func (a *App) migrate(dbURL string) error {
	sqlDB, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer sqlDB.Close()

	// Wait for the database to be ready and not in recovery mode (e.g. after a bulk restore or WAL replay)
	deadline := time.Now().Add(30 * time.Second)
	for {
		var inRecovery bool
		err := sqlDB.QueryRow("SELECT pg_is_in_recovery()").Scan(&inRecovery)
		if err == nil && !inRecovery {
			var databaseName string
			if err := sqlDB.QueryRow("SELECT current_database()").Scan(&databaseName); err != nil {
				return fmt.Errorf("failed to identify migration database: %w", err)
			}
			log.Printf("Database [%s] is ready for migrations", databaseName)
			break
		}
		if time.Now().After(deadline) {
			if err != nil {
				return fmt.Errorf("database not ready after 30s: %w", err)
			}
			return fmt.Errorf("database still in recovery mode after 30s")
		}
		log.Printf("Waiting for database to finish recovery / accept connections... (status: %v)", err)
		time.Sleep(500 * time.Millisecond)
	}

	// If the database was restored from a Cloud backup, the base schema tables (e.g. organizations)
	// already exist, but goose_db_version may not be initialized. Mark initial baseline as applied
	// so Goose skips the initial schema and proceeds to apply POS extensions.
	var baseSchemaExists bool
	if err := sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' AND table_name = 'organizations'
		);
	`).Scan(&baseSchemaExists); err != nil {
		return fmt.Errorf("failed to check restored schema: %w", err)
	}

	if baseSchemaExists {
		if _, err := sqlDB.Exec(`
			CREATE TABLE IF NOT EXISTS goose_db_version (
				id serial PRIMARY KEY,
				version_id bigint NOT NULL,
				is_applied boolean NOT NULL,
				tstamp timestamp NULL default now()
			);
			INSERT INTO goose_db_version (version_id, is_applied)
			SELECT 20260918111505, true
			WHERE NOT EXISTS (SELECT 1 FROM goose_db_version WHERE version_id = 20260918111505);

			DELETE FROM goose_db_version WHERE version_id IN (
				20260813124500, 20260821060910, 20260829102802,
				20260829112654, 20260916085642, 20260917120000,
				20260918140000, 20261001153500
			);
		`); err != nil {
			return fmt.Errorf("failed to initialize restored migration baseline: %w", db.WithStorageHint(err))
		}
	}

	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %v", err)
	}

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", db.WithStorageHint(err))
	}

	// Ensure legacy DB triggers that double-count cashier session balances or inventory are dropped
	for _, statement := range []string{
		"DROP TRIGGER IF EXISTS trg_update_cashier_session_balance ON pos_transactions;",
		"DROP TRIGGER IF EXISTS trg_deduct_inventory_on_pos_transaction ON pos_transaction_lines;",
		"DROP FUNCTION IF EXISTS update_cashier_session_balance();",
		"ALTER TABLE IF EXISTS transfer_requests ADD COLUMN IF NOT EXISTS is_stock_reserved BOOLEAN NOT NULL DEFAULT false;",
	} {
		if _, err := sqlDB.Exec(statement); err != nil {
			return fmt.Errorf("failed to finalize POS schema: %w", db.WithStorageHint(err))
		}
	}

	return nil
}

// FetchCloudTenants fetches available tenants from the cloud URL
func (a *App) FetchCloudTenants(slug string) interface{} {
	url := fmt.Sprintf("%s/api/tenants/%s", a.cfg.CloudURL, slug)
	log.Printf("Fetching cloud tenant from: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Sprintf("Error creating request: %v", err)
	}

	req.Header.Set("x-tenant-id", "masterdb")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("FetchCloudTenants Error: %v", err)
		return fmt.Sprintf("Error fetching tenant: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Cloud API returned status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Error: Cloud API returned status %d", resp.StatusCode)
	}

	var result repository.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Sprintf("Error decoding response: %v", err)
	}

	return result.Data
}

// CloneTenant clones a specific tenant and its master data from the cloud
func (a *App) CloneTenant(slug string) string {
	if a.masterRepo == nil {
		return "Error: Local database not started"
	}

	emitProgress := func(phase string, percent int, bytes int64, speed string, estSecs int, msg string) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "clone:progress", map[string]interface{}{
				"phase":             phase,
				"percent":           percent,
				"bytes_transferred": bytes,
				"speed":             speed,
				"estimated_seconds": estSecs,
				"message":           msg,
			})
		}
	}

	emitProgress("connecting", 5, 0, "", 30, "Connecting to cloud server...")

	previousSync := a.syncService
	if previousSync != nil {
		log.Printf("CloneTenant: stopping background sync during backup and restore")
		previousSync.Stop()
		defer func() {
			if a.syncService == previousSync {
				a.StartSyncService(previousSync.TenantSlug())
			}
		}()
	}

	// 1. Fetch Tenant details from cloud
	emitProgress("metadata", 12, 0, "", 28, fmt.Sprintf("Fetching tenant details for [%s]...", slug))
	url := fmt.Sprintf("%s/api/tenants/%s", a.cfg.CloudURL, slug)
	log.Printf("Fetching tenant details for [%s] from: %s", slug, url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Sprintf("Error creating request: %v", err)
	}

	req.Header.Set("x-tenant-id", "masterdb")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("CloneTenant Fetch Error: %v", err)
		return fmt.Sprintf("Error connecting to cloud: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("CloneTenant Cloud API returned status: %d", resp.StatusCode)

	var result repository.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Sprintf("Error decoding tenant data: %v", err)
	}

	type cloudTenant struct {
		ID         string          `json:"id"`
		TenantName string          `json:"tenant_name"`
		Slug       string          `json:"slug"`
		DbConnStr  string          `json:"db_conn_str"`
		IsActive   bool            `json:"is_active"`
		Settings   json.RawMessage `json:"settings"`
	}

	tenantData, err := json.Marshal(result.Data)
	if err != nil {
		return fmt.Sprintf("Error processing tenant data: %v", err)
	}

	var tenant cloudTenant
	if err := json.Unmarshal(tenantData, &tenant); err != nil {
		log.Printf("CloneTenant: JSON unmarshal error: %v — raw data: %s", err, string(tenantData))
		return fmt.Sprintf("Error parsing tenant data: %v", err)
	}

	// 2. Create the tenant database locally
	emitProgress("creating_db", 20, 0, "", 25, fmt.Sprintf("Preparing database schema for %s...", tenant.TenantName))
	tenantDBName := slug
	_, err = a.masterPool.Exec(a.ctx, fmt.Sprintf(`CREATE DATABASE "%s" WITH ENCODING 'UTF8'`, tenantDBName))
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Sprintf("Error creating tenant database: %v", db.WithStorageHint(err))
		}
		log.Printf("Database %s already exists, proceeding...", tenantDBName)
	}

	// 3. Download tenant backup via gRPC and restore it directly into the NEW tenant DB.
	home, _ := os.UserHomeDir()
	backupDir := filepath.Join(home, ".nembus", "backups", tenant.Slug)
	backupToken := a.cfg.BackupAuthToken

	tenantDBURL := fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable",
		a.dbManager.GetConfig().Username,
		a.dbManager.GetConfig().Password,
		a.dbManager.GetConfig().Port,
		tenantDBName,
	)

	log.Printf("CloneTenant: downloading backup for [%s] via gRPC from %s", tenant.Slug, a.cfg.GRPCAddr)
	backupPath, err := clientbackup.DownloadBackupWithProgress(
		a.cfg.GRPCAddr,
		tenant.Slug,
		backupToken,
		backupDir,
		tenantDBURL,
		func(stage string, percent int, bytesTransferred int64, speed string, estimatedSecs int, message string) {
			emitProgress(stage, percent, bytesTransferred, speed, estimatedSecs, message)
		},
	)
	if err != nil {
		log.Printf("CloneTenant Backup Error: %v", err)
		return fmt.Sprintf("Error downloading/restoring tenant backup: %v", db.WithStorageHint(err))
	}
	log.Printf("Tenant backup restored from: %s", backupPath)

	// 4. Run POS extensions migrations directly on the restored tenant database
	emitProgress("migrating", 88, 0, "", 5, "Applying offline tables and local triggers...")
	log.Printf("CloneTenant: applying POS migrations to tenant DB [%s]", tenant.Slug)
	if err := a.migrate(tenantDBURL); err != nil {
		log.Printf("Error applying POS migrations to tenant DB: %v", err)
		return fmt.Sprintf("Error applying POS migrations to tenant database: %v", err)
	}
	log.Printf("POS migrations applied successfully to tenant DB [%s]", tenant.Slug)

	// 5. Ensure the local tenant record is correct in the master DB.
	emitProgress("finalizing", 96, 0, "", 2, "Finalizing local configuration...")
	upsertSQL := `
		INSERT INTO tenants (id, tenant_name, slug, db_conn_str, is_active, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (slug) DO UPDATE SET
			tenant_name = EXCLUDED.tenant_name,
			db_conn_str = EXCLUDED.db_conn_str,
			is_active = EXCLUDED.is_active,
			settings = EXCLUDED.settings,
			updated_at = NOW();
	`
	_, err = a.masterPool.Exec(a.ctx, upsertSQL,
		tenant.ID, tenant.TenantName, tenant.Slug, tenantDBURL, true, tenant.Settings)
	if err != nil {
		log.Printf("Error upserting tenant record: %v", err)
		return fmt.Sprintf("Error finalizing local tenant record: %v", err)
	}

	if a.tenantManager != nil {
		a.tenantManager.EvictPool(tenant.Slug)
	}

	emitProgress("complete", 100, 0, "", 0, "Organization cloned successfully!")

	// 6. Start sync service for this tenant
	a.StartSyncService(tenant.Slug)

	return "Success"
}

// StartSyncService starts the background synchronization worker
func (a *App) StartSyncService(slug string) {
	if a.syncService != nil {
		a.syncService.Stop()
	}
	if a.masterPool == nil {
		log.Println("Warning: Cannot start sync service without master pool")
		return
	}

	bgCtx := context.Background()
	if a.tenantManager == nil && a.masterRepo != nil {
		a.tenantManager = manager.NewManager(a.masterRepo)
	}
	if a.tenantManager == nil {
		log.Printf("[SyncService] Cannot resolve tenant database for [%s]", slug)
		return
	}
	tenantPool, err := a.tenantManager.GetPool(bgCtx, slug)
	if err != nil || tenantPool == nil {
		log.Printf("[SyncService] Cannot resolve tenant database for [%s]: %v", slug, err)
		return
	}
	var configured bool
	deviceCtx, cancelDevice := context.WithTimeout(bgCtx, 15*time.Second)
	defer cancelDevice()
	if err := tenantPool.QueryRow(deviceCtx, `SELECT COALESCE((SELECT store_id FROM local_device_config ORDER BY id LIMIT 1),0) > 0`).Scan(&configured); err != nil {
		log.Printf("[SyncService] Cannot check device configuration: %v", err)
	} else if !configured {
		if raw := a.LoadDeviceConfig(); raw != "" {
			identity, err := sync.ParseDeviceIdentity(raw)
			if err == nil {
				err = sync.SaveDeviceIdentity(deviceCtx, tenantPool, identity)
			}
			if err != nil {
				log.Printf("[SyncService] Cannot restore saved device configuration: %v", err)
			} else {
				log.Printf("[SyncService] Restored local device configuration for store %d", identity.StoreID)
			}
		}
	}

	grpcAddr := ""
	if a.cfg != nil {
		grpcAddr = a.cfg.GRPCAddr
		if grpcAddr == "" {
			grpcAddr = a.cfg.CloudURL
		}
	}
	a.syncService = sync.NewSyncService(bgCtx, tenantPool, grpcAddr, slug)

	// Reset any sync_queue items that previously hit max_retries (e.g. due to
	// transient cloud errors or FK constraint issues that have since been fixed)
	// so they are retried on the next sync cycle.
	if count, err := a.syncService.ResetFailedItems(); err != nil {
		log.Printf("[SyncService] Warning: could not reset failed sync_queue items: %v", err)
	} else if count > 0 {
		log.Printf("[SyncService] Reset %d previously-failed sync_queue items to pending — will retry shortly", count)
	}
	a.syncService.Start()
}

// InitializeSync resumes synchronization if a tenant is already setup
func (a *App) InitializeSync() {
	if !a.IsAppSetup() {
		return
	}

	// Fetch the first active tenant to resume sync
	// Since it's a desktop app, we usually have only one tenant

	// Actually, let's just query the local master DB for the tenant slug
	if a.masterRepo != nil {
		ctx := context.Background()
		t, err := a.masterRepo.ListActiveTenants(ctx)
		if err == nil && len(t) > 0 {
			a.StartSyncService(t[0].Slug)
		}
	}
}

func (a *App) runBackend(masterPool *pgxpool.Pool) {
	// reuse the pool
	a.tenantManager = manager.NewManager(a.masterRepo)
	tenantManager := a.tenantManager

	// Initialize Use Cases
	userUC := usecase.NewUserUseCase()
	orgUC := usecase.NewOrganizationUseCase()
	authUC := usecase.NewAuthUseCase()
	moduleUC := usecase.NewModuleUseCase()
	imageUC := usecase.NewImageUseCase()
	navigationUC := usecase.NewNavigationUseCase()
	permissionUC := usecase.NewPermissionUseCase()
	roleUC := usecase.NewRoleUseCase()
	menuUC := usecase.NewMenuUseCase()
	submenuUC := usecase.NewSubmenuUseCase()
	posUC := usecase.NewPosUseCase()
	posPaymentUC := usecase.NewPosPaymentUseCase()
	salesReturnUC := usecase.NewSalesReturnUseCase()
	posTerminalsUC := usecase.NewPosTerminalsUseCase()
	storageLocationsUC := usecase.NewStorageLocationsUseCase()
	tenantUC := usecase.NewTenantUseCase()
	storesUC := usecase.NewStoreUseCase()
	cartUC := usecase.NewCartUseCase()
	orderUC := usecase.NewOrderUseCase()
	restaurantUC := usecase.NewRestaurantUseCase()
	customerUC := usecase.NewCustomerUseCase()
	businessPartnerUC := usecase.NewBusinessPartnerUseCase()
	bpPriceContractUC := usecase.NewBPPriceContractUseCase()
	uomUC := usecase.NewUOMUseCase()
	priceListsUC := usecase.NewPriceListsUseCase()
	taxCategoriesUC := usecase.NewTaxCategoriesUseCase()
	cashierSessionUC := usecase.NewCashierSessionUseCase()
	brandUC := usecase.NewBrandUseCase()
	cashierUC := usecase.NewCashierUseCase()
	productBarcodeUC := usecase.NewProductBarcodeUseCase()
	productPricingUC := usecase.NewProductPricingUseCase()
	inventoryStockUC := usecase.NewInventoryStockUseCase()
	productVariantUC := usecase.NewProductVariantUseCase()
	promotionUC := usecase.NewPromotionUseCase()
	loyaltyUC := usecase.NewLoyaltyUseCase()
	productCatalogUC := usecase.NewProductCatalogUseCase()
	printUC := usecase.NewPrintUseCase()
	paymentTermsUC := usecase.NewPaymentTermsUseCase()
	purchaseOrdersUC := usecase.NewPurchaseOrdersUseCase()
	invoiceUC := usecase.NewInvoiceUseCase()

	r := setupRouter(tenantManager, a.masterRepo, userUC, orgUC, authUC, moduleUC, imageUC, navigationUC, permissionUC, roleUC, menuUC, submenuUC, posUC, posPaymentUC, salesReturnUC, posTerminalsUC, storageLocationsUC, tenantUC, storesUC, cartUC, orderUC, invoiceUC, restaurantUC, customerUC, uomUC, priceListsUC, taxCategoriesUC, cashierSessionUC, brandUC, cashierUC, productBarcodeUC, productPricingUC, inventoryStockUC, productVariantUC, promotionUC, loyaltyUC, productCatalogUC, printUC, businessPartnerUC, bpPriceContractUC, paymentTermsUC, purchaseOrdersUC, a.cfg)
	r.Static("/images", "./images")

	log.Printf("Starting Gin HTTP server on port %s (Swagger: http://localhost:%s/swagger/index.html)", a.cfg.Port, a.cfg.Port)
	if err := r.Run(":" + a.cfg.Port); err != nil {
		log.Printf("failed to run server: %v", err)
	}
}

// CreateInitialTenant creates the first tenant in the master database
func (a *App) CreateInitialTenant(name, slug string) string {
	if a.masterRepo == nil {
		return "Error: Database not started"
	}

	uc := usecase.NewTenantUseCase()
	uc.SetRepository(a.masterRepo)

	params := repository.CreateTenantParams{
		TenantName: name,
		Slug:       slug,
		DbConnStr:  a.cfg.MasterDBURL, // Use the same DB for the initial tenant
		IsActive:   pgtype.Bool{Bool: true, Valid: true},
		Settings:   []byte("{}"),
	}

	resp := uc.CreateTenant(a.ctx, params)
	if resp.StatusCode != 201 {
		return fmt.Sprintf("Error: %s", resp.Message)
	}

	return "Success"
}

// CreateInitialOrganization creates the first organization for the tenant
func (a *App) CreateInitialOrganization(name, code string) string {
	if a.masterRepo == nil {
		return "Error: Database not started"
	}

	uc := usecase.NewOrganizationUseCase()
	uc.SetRepository(a.masterRepo)

	resp := uc.CreateOrganization(a.ctx, name, code, nil, nil, nil, nil, true)
	if resp.StatusCode != 201 {
		return fmt.Sprintf("Error: %s", resp.Message)
	}

	return "Success"
}

// CreateInitialAdmin creates the first administrator user
func (a *App) CreateInitialAdmin(firstName, lastName, username, email, password string) string {
	if a.masterRepo == nil {
		return "Error: Database not started"
	}

	uc := usecase.NewUserUseCase()
	uc.SetRepository(a.masterRepo)

	resp := uc.CreateUser(a.ctx, firstName, lastName, username, email, true, &password, nil, nil)
	if resp.StatusCode != 201 {
		return fmt.Sprintf("Error: %s", resp.Message)
	}

	// Write marker file to indicate setup is complete
	home, _ := os.UserHomeDir()
	markerPath := filepath.Join(home, ".nembus", ".setup_done")
	_ = os.MkdirAll(filepath.Dir(markerPath), 0755)
	_ = os.WriteFile(markerPath, []byte("done"), 0644)

	return "Success"
}

// IsAppSetup checks if the application's database has already been initialized
func (a *App) IsAppSetup() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	markerPath := filepath.Join(home, ".nembus", ".setup_done")
	_, err = os.Stat(markerPath)
	return err == nil
}

func (a *App) GetDBStatus() string {
	if a.dbManager == nil {
		return "Not Started"
	}
	return a.dbManager.GetConnectionString()
}

// IsBackendReady returns true once the embedded database and the Gin HTTP
// server have both finished initialising.  The frontend calls this on startup
// so it can wait before making any API requests.
func (a *App) IsBackendReady() bool {
	return a.masterPool != nil
}

// LoadDeviceConfig reads the persisted device configuration from disk.
// Returns an empty string if no config has been saved yet (first run).
func (a *App) LoadDeviceConfig() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("LoadDeviceConfig: cannot get home dir: %v", err)
		return ""
	}
	path := filepath.Join(home, ".nembus", "device_config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		// No config yet — normal on first launch
		return ""
	}
	return string(data)
}

// SaveDeviceConfig writes the device configuration to disk so it persists
// across Wails restarts and clean builds, and updates local_device_config in the tenant DB.
func (a *App) SaveDeviceConfig(configJSON string) string {
	identity, identityErr := sync.ParseDeviceIdentity(configJSON)
	if identityErr != nil && !errors.Is(identityErr, sync.ErrDeviceConfigIncomplete) {
		return fmt.Sprintf("Error: %v", identityErr)
	}
	if identityErr == nil {
		if a.masterRepo == nil {
			return "Error: local database is not ready to save device configuration"
		}
		if a.tenantManager == nil {
			a.tenantManager = manager.NewManager(a.masterRepo)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		var slug string
		if a.syncService != nil {
			slug = a.syncService.TenantSlug()
		} else {
			tenants, err := a.masterRepo.ListActiveTenants(ctx)
			if err != nil {
				return fmt.Sprintf("Error: cannot resolve configured tenant: %v", err)
			}
			if len(tenants) != 1 {
				return "Error: device configuration requires one active local tenant"
			}
			slug = tenants[0].Slug
		}
		pool, err := a.tenantManager.GetPool(ctx, slug)
		if err != nil {
			return fmt.Sprintf("Error: cannot open tenant database: %v", err)
		}
		if err := sync.SaveDeviceIdentity(ctx, pool, identity); err != nil {
			log.Printf("SaveDeviceConfig: database save failed: %v", err)
			return fmt.Sprintf("Error: %v", err)
		}
		log.Printf("SaveDeviceConfig: saved local_device_config for tenant %s, store %d", slug, identity.StoreID)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Sprintf("Error: cannot get home dir: %v", err)
	}
	dir := filepath.Join(home, ".nembus")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Sprintf("Error: cannot create dir: %v", err)
	}
	path := filepath.Join(dir, "device_config.json")
	if err := os.WriteFile(path, []byte(configJSON), 0644); err != nil {
		return fmt.Sprintf("Error: cannot write config: %v", err)
	}
	log.Printf("SaveDeviceConfig: config saved to %s", path)

	if identityErr == nil {
		markerPath := filepath.Join(dir, ".setup_done")
		if err := os.WriteFile(markerPath, []byte("done"), 0644); err != nil {
			return fmt.Sprintf("Error: cannot save setup marker: %v", err)
		}
		if a.syncService != nil {
			a.syncService.SyncNow()
		}
	}

	return "OK"
}

// GetDevToken generates a valid development JWT token for local testing and Postman/Wails calls.
func (a *App) GetDevToken() string {
	token, err := middleware.GenerateJWTToken("1", "admin")
	if err != nil {
		log.Printf("GetDevToken error: %v", err)
		return ""
	}
	return token
}

// GeideaConfigure configures or updates the terminal parameters (TCP or Serial COM).
func (a *App) GeideaConfigure(cfg geidea.TerminalConfig) string {
	if a.geideaClient == nil {
		a.geideaClient = geidea.NewClient(cfg)
	} else {
		a.geideaClient.UpdateConfig(cfg)
	}
	log.Printf("[GEIDEA] Terminal configured: mode=%s, ip=%s:%s, com=%s, baud=%d",
		cfg.Mode, cfg.IP, cfg.Port, cfg.SerialPort, cfg.BaudRate)
	return "OK"
}

// GeideaCheckStatus tests connectivity to the physical terminal.
func (a *App) GeideaCheckStatus() (bool, error) {
	if a.geideaClient == nil {
		return false, fmt.Errorf("geidea terminal is not configured")
	}
	return a.geideaClient.CheckStatus(a.ctx)
}

// GeideaPurchase initiates a card purchase with live status event streaming to the frontend.
func (a *App) GeideaPurchase(req geidea.PurchaseRequest) (*geidea.TransactionResult, error) {
	if a.geideaClient == nil {
		return nil, fmt.Errorf("geidea terminal is not configured")
	}
	return a.geideaClient.Purchase(a.ctx, req, func(event geidea.TerminalStatusEvent) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "geidea:status", event)
		}
	})
}

// GeideaRefund initiates a refund using original transaction RRN.
func (a *App) GeideaRefund(req geidea.RefundRequest) (*geidea.TransactionResult, error) {
	if a.geideaClient == nil {
		return nil, fmt.Errorf("geidea terminal is not configured")
	}
	return a.geideaClient.Refund(a.ctx, req, func(event geidea.TerminalStatusEvent) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "geidea:status", event)
		}
	})
}

// GeideaReconcile triggers End of Day settlement and reconciliation with SAMA.
func (a *App) GeideaReconcile() (*geidea.ReconciliationResult, error) {
	if a.geideaClient == nil {
		return nil, fmt.Errorf("geidea terminal is not configured")
	}
	return a.geideaClient.Reconcile(a.ctx, func(event geidea.TerminalStatusEvent) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "geidea:status", event)
		}
	})
}

// GeideaCancel sends the abort signal to cancel an in-flight transaction.
func (a *App) GeideaCancel() error {
	if a.geideaClient == nil {
		return fmt.Errorf("geidea terminal is not configured")
	}
	return a.geideaClient.Cancel(a.ctx)
}

// GeideaRepeatLast queries the terminal for the last transaction result.
func (a *App) GeideaRepeatLast() (*geidea.TransactionResult, error) {
	if a.geideaClient == nil {
		return nil, fmt.Errorf("geidea terminal is not configured")
	}
	return a.geideaClient.RepeatLast(a.ctx)
}

// StartGeideaPayment initiates the payment through Wails runtime with status streaming (backward-compatible).
func (a *App) StartGeideaPayment(terminalID int, amount float64, ecrRef string) (string, error) {
	if a.geideaClient == nil {
		a.geideaClient = geidea.NewClient(geidea.TerminalConfig{
			Mode: "network",
			IP:   "127.0.0.1",
			Port: "80",
		})
	}
	res, err := a.GeideaPurchase(geidea.PurchaseRequest{
		Amount:        amount,
		InvoiceNumber: ecrRef,
	})
	if err != nil {
		return "", err
	}
	resJSON, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(resJSON), nil
}

// PrintMadaReceipt builds an ESC/POS formatted mada transaction receipt and sends it to the printer.
func (a *App) PrintMadaReceipt(data printing.MadaReceiptData, cfg printing.PrinterConfig) error {
	rawBytes := printing.BuildMadaReceipt(data)
	return printing.Print(cfg, rawBytes)
}


