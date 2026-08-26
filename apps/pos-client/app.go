package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	clientbackup "github.com/NasTecSol/nembus-client/client"
	"github.com/NasTecSol/nembus-client/internal/config"
	"github.com/NasTecSol/nembus-client/internal/db"
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
)

//go:embed migrations/*.sql
var migrations embed.FS

// App struct
type App struct {
	ctx         context.Context
	dbManager   *db.DBManager
	masterRepo  *repository.Queries
	masterPool  *pgxpool.Pool
	syncService *sync.SyncService
	cfg         *config.Config
	version     string
	ghToken     string
	ghRepo      string
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

	// Run migrations before connecting with pgxpool.
	// If migration fails (e.g. due to a stale/partial previous install), wipe the
	// data directory and retry once from a clean DB.
	if err := a.migrate(a.cfg.MasterDBURL); err != nil {
		log.Printf("Migration failed: %v — wiping data dir and retrying from scratch", err)

		// Stop the embedded DB so we can delete the data dir.
		_ = a.dbManager.Stop()

		// Wipe the data directory.
		home, _ := os.UserHomeDir()
		dataPath := filepath.Join(home, ".nembus", "data")
		if removeErr := os.RemoveAll(dataPath); removeErr != nil {
			return fmt.Sprintf("Error cleaning data dir: %v", removeErr)
		}

		// Also remove the setup marker so the wizard starts fresh.
		_ = os.Remove(filepath.Join(home, ".nembus", ".setup_done"))

		// Start a fresh embedded DB.
		a.dbManager = db.NewDBManager(dbCfg)
		if startErr := a.dbManager.Start(); startErr != nil {
			return fmt.Sprintf("Error restarting DB after wipe: %v", startErr)
		}

		// Retry migration on the clean DB.
		if retryErr := a.migrate(a.cfg.MasterDBURL); retryErr != nil {
			return fmt.Sprintf("Error running migrations (after reset): %v", retryErr)
		}
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

	// If the database was restored from a Cloud backup, the base schema tables (e.g. organizations)
	// already exist, but goose_db_version may not be initialized. Mark initial baseline as applied
	// so Goose skips 20260813124500.sql and proceeds to apply POS extensions.
	var baseSchemaExists bool
	_ = sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' AND table_name = 'organizations'
		);
	`).Scan(&baseSchemaExists)

	if baseSchemaExists {
		_, _ = sqlDB.Exec(`
			CREATE TABLE IF NOT EXISTS goose_db_version (
				id serial PRIMARY KEY,
				version_id bigint NOT NULL,
				is_applied boolean NOT NULL,
				tstamp timestamp NULL default now()
			);
			INSERT INTO goose_db_version (version_id, is_applied)
			SELECT 20260813124500, true
			WHERE NOT EXISTS (SELECT 1 FROM goose_db_version WHERE version_id = 20260813124500);
		`)
	}

	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %v", err)
	}

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	// Ensure legacy DB triggers that double-count cashier session balances are dropped
	_, _ = sqlDB.Exec("DROP TRIGGER IF EXISTS trg_update_cashier_session_balance ON pos_transactions;")
	_, _ = sqlDB.Exec("DROP FUNCTION IF EXISTS update_cashier_session_balance();")

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

	// 1. Fetch Tenant details from cloud
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

	// cloudTenant is a plain struct for parsing the cloud API's JSON response.
	// repository.Tenant uses pgtype.* types which require special JSON nesting
	// and cannot be decoded directly from plain REST JSON.
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
	tenantDBName := slug
	// Check if database exists, create if not
	// We use standard db/sql for this as creating databases cannot run in a transaction block which pgx pool might implicitly use or log.
	// Actually masterPool.Exec without a transaction is fine, but CREATE DATABASE cannot be executed inside a transaction block.
	// Let's drop it if it exists (for a clean clone), but we must disconnect first. Since it's a new setup, it probably doesn't exist,
	// but to be safe we try to create it and ignore "already exists" errors.
	_, err = a.masterPool.Exec(a.ctx, fmt.Sprintf(`CREATE DATABASE "%s" WITH ENCODING 'UTF8'`, tenantDBName))
	if err != nil {
		// Ignore error if database already exists
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Sprintf("Error creating tenant database: %v", err)
		}
		log.Printf("Database %s already exists, proceeding...", tenantDBName)
	}

	// 3. Download tenant backup via gRPC and restore it directly into the NEW tenant DB.
	home, _ := os.UserHomeDir()
	backupDir := filepath.Join(home, ".nembus", "backups", tenant.Slug)
	backupToken := a.cfg.BackupAuthToken

	// Construct the tenant-specific DB URL
	// We extract username, password, port from a.cfg.MasterDBURL
	// Since we know StartDatabase uses postgres / password / nembus / port
	tenantDBURL := fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable",
		a.dbManager.GetConfig().Username,
		a.dbManager.GetConfig().Password,
		a.dbManager.GetConfig().Port,
		tenantDBName,
	)

	log.Printf("CloneTenant: downloading backup for [%s] via gRPC from %s", tenant.Slug, a.cfg.GRPCAddr)
	backupPath, err := clientbackup.DownloadBackup(
		a.cfg.GRPCAddr,
		tenant.Slug,
		backupToken,
		backupDir,
		tenantDBURL, // restore directly into the tenant's database
	)
	if err != nil {
		log.Printf("CloneTenant Backup Error: %v", err)
		return fmt.Sprintf("Error downloading/restoring tenant backup: %v", err)
	}
	log.Printf("Tenant backup restored from: %s", backupPath)

	// 4. Run POS extensions migrations directly on the restored tenant database
	log.Printf("CloneTenant: applying POS migrations to tenant DB [%s]", tenant.Slug)
	if err := a.migrate(tenantDBURL); err != nil {
		log.Printf("Error applying POS migrations to tenant DB: %v", err)
		return fmt.Sprintf("Error applying POS migrations to tenant database: %v", err)
	}
	log.Printf("POS migrations applied successfully to tenant DB [%s]", tenant.Slug)

	// 5. Ensure the local tenant record is correct in the master DB.
	// We upsert it to make sure the slug exists and points to our local tenant DB URL.
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

	// 6. Initialize sync (will use the newly created tenant)
	a.InitializeSync()

	// 7. Finalize setup
	markerPath := filepath.Join(home, ".nembus", ".setup_done")
	_ = os.MkdirAll(filepath.Dir(markerPath), 0755)
	_ = os.WriteFile(markerPath, []byte("done"), 0644)

	return "Success"
}

// StartSyncService starts the background synchronization worker
func (a *App) StartSyncService(slug string) {
	if a.masterPool == nil {
		log.Println("Warning: Cannot start sync service without master pool")
		return
	}

	bgCtx := context.Background()
	tenantPool := a.masterPool
	if a.masterRepo != nil {
		tenantManager := manager.NewManager(a.masterRepo)
		pool, err := tenantManager.GetPool(bgCtx, slug)
		if err == nil && pool != nil {
			tenantPool = pool
		} else {
			log.Printf("[SyncService] Could not resolve pool for tenant [%s]: %v (falling back to masterPool)", slug, err)
		}
	}

	a.syncService = sync.NewSyncService(bgCtx, tenantPool, a.cfg.CloudURL, slug)
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
	tenantManager := manager.NewManager(a.masterRepo)

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

	r := setupRouter(tenantManager, a.masterRepo, userUC, orgUC, authUC, moduleUC, imageUC, navigationUC, permissionUC, roleUC, menuUC, submenuUC, posUC, posPaymentUC, salesReturnUC, posTerminalsUC, storageLocationsUC, tenantUC, storesUC, cartUC, orderUC, restaurantUC, customerUC, uomUC, priceListsUC, taxCategoriesUC, cashierSessionUC, brandUC, cashierUC, productBarcodeUC, productPricingUC, inventoryStockUC, productVariantUC, promotionUC, loyaltyUC, productCatalogUC, printUC, businessPartnerUC, paymentTermsUC, a.cfg)
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
// across Wails restarts and clean builds.
func (a *App) SaveDeviceConfig(configJSON string) string {
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
