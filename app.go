package main

import (
	"NEMBUS/internal/config"
	"NEMBUS/internal/db"
	"NEMBUS/internal/middleware/manager"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/sync"
	"NEMBUS/internal/usecase"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// App struct
type App struct {
	ctx         context.Context
	dbManager   *db.DBManager
	masterRepo  *repository.Queries
	masterPool  *pgxpool.Pool
	syncService *sync.SyncService
	cfg         *config.Config
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if a.IsAppSetup() {
		// Auto-start DB if setup is already complete
		// Using default credentials as these are local to the machine
		go a.StartDatabase("postgres", "password", "nembus", 5432)
	}
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
		return fmt.Sprintf("Error: %v", err)
	}

	// Initialize master repo
	ctx := context.Background()

	// Run migrations before connecting with pgxpool
	if err := a.migrate(a.cfg.MasterDBURL); err != nil {
		return fmt.Sprintf("Error running migrations: %v", err)
	}

	pool, repo, err := setupDatabase(ctx, a.cfg.MasterDBURL)
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
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %v", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	return nil
}

// FetchCloudTenants fetches available tenants from the cloud URL
func (a *App) FetchCloudTenants() interface{} {
	url := fmt.Sprintf("%s/api/tenants", a.cfg.CloudURL)
	log.Printf("Fetching cloud tenants from: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Sprintf("Error creating request: %v", err)
	}

	req.Header.Set("x-tenant-id", "masterdb")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("FetchCloudTenants Error: %v", err)
		return fmt.Sprintf("Error fetching tenants: %v", err)
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

	// Data is a single repository.Tenant
	tenantData, err := json.Marshal(result.Data)
	if err != nil {
		return "Error processing tenant data"
	}

	var tenant repository.Tenant
	if err := json.Unmarshal(tenantData, &tenant); err != nil {
		return "Error parsing tenant data"
	}

	// 2. Create tenant locally in master DB
	params := repository.CreateTenantParams{
		TenantName: tenant.TenantName,
		Slug:       tenant.Slug,
		DbConnStr:  a.cfg.MasterDBURL, // Desktop always uses local master DB
		IsActive:   pgtype.Bool{Bool: true, Valid: true},
		Settings:   tenant.Settings,
	}

	uc := usecase.NewTenantUseCase()
	uc.SetRepository(a.masterRepo)
	createResp := uc.CreateTenant(a.ctx, params)
	if createResp.StatusCode != 201 {
		return fmt.Sprintf("Error creating local tenant: %s", createResp.Message)
	}

	// 3. Clone Master Data (Organizations, Users, etc.)
	cloner := sync.NewTenantCloner(a.ctx, a.masterRepo, a.cfg.CloudURL)
	if err := cloner.CloneMasterData(tenant.Slug); err != nil {
		return fmt.Sprintf("Error cloning master data: %v", err)
	}

	// 4. Start Sync Service
	a.StartSyncService(tenant.Slug)

	// 5. Finalize setup
	home, _ := os.UserHomeDir()
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
	a.syncService = sync.NewSyncService(a.ctx, a.masterPool, a.cfg.CloudURL, slug)
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
	posTerminalsUC := usecase.NewPosTerminalsUseCase()
	storageLocationsUC := usecase.NewStorageLocationsUseCase()
	tenantUC := usecase.NewTenantUseCase()
	storesUC := usecase.NewStoreUseCase()
	restaurantUC := usecase.NewRestaurantUseCase()

	r := setupRouter(tenantManager, userUC, orgUC, authUC, moduleUC, imageUC, navigationUC, permissionUC, roleUC, menuUC, submenuUC, posUC, posTerminalsUC, storageLocationsUC, tenantUC, storesUC, restaurantUC, a.cfg)
	r.Static("/images", "./images")

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

	resp := uc.CreateUser(a.ctx, firstName, lastName, username, email, true, &password, nil)
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
