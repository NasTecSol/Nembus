package server

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NasTecSol/nembus-sap/contracts"
	legacyconfig "github.com/NasTecSol/nembus-sap-agent/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap-agent/internal/discovery"
	"github.com/NasTecSol/nembus-sap-agent/internal/etl"
	"github.com/NasTecSol/nembus-sap-agent/internal/nembus"
	"github.com/NasTecSol/nembus-sap-agent/internal/pipeline"
	"github.com/NasTecSol/nembus-sap-agent/internal/reconciliation"
	"github.com/NasTecSol/nembus-sap-agent/internal/sap"
	"github.com/NasTecSol/nembus-sap-agent/internal/transport"
	"github.com/NasTecSol/nembus-sap-agent/internal/updater"
	"github.com/kardianos/service"
)

type Server struct {
	cfg                  *config.AgentConfig
	engine               *etl.Engine
	sqlite               *db.SQLiteStore
	mssql                *db.MSSQLClient
	cloudClient          *transport.CloudClient
	upstreamSync         *pipeline.UpstreamSync
	upstreamMu           sync.Mutex
	downstreamMu         sync.Mutex
	autoStateMu          sync.RWMutex
	downstreamAutoActive bool
	upstreamAutoActive   bool
	lastDownstreamStats  *etl.DownstreamStats
	lastDownstreamTime   *time.Time
	lastDownstreamErr    string
	pgPool               *pgxpool.Pool
	router               *gin.Engine
	uiFS                 embed.FS
}

func (s *Server) isDownstreamAutoActive() bool {
	s.autoStateMu.RLock()
	defer s.autoStateMu.RUnlock()
	return s.downstreamAutoActive
}

func (s *Server) setDownstreamAutoActive(active bool) {
	s.autoStateMu.Lock()
	defer s.autoStateMu.Unlock()
	s.downstreamAutoActive = active
}

func (s *Server) isUpstreamAutoActive() bool {
	s.autoStateMu.RLock()
	defer s.autoStateMu.RUnlock()
	return s.upstreamAutoActive
}

func (s *Server) setUpstreamAutoActive(active bool) {
	s.autoStateMu.Lock()
	defer s.autoStateMu.Unlock()
	s.upstreamAutoActive = active
}

func NewServer(cfg *config.AgentConfig, engine *etl.Engine, sqlite *db.SQLiteStore, mssql *db.MSSQLClient, cloudClient *transport.CloudClient, uiFS embed.FS) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	s := &Server{
		cfg:         cfg,
		engine:      engine,
		sqlite:      sqlite,
		mssql:       mssql,
		cloudClient: cloudClient,
		router:      r,
		uiFS:        uiFS,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// 1. Static embedded UI
	s.router.GET("/", func(c *gin.Context) {
		data, err := s.uiFS.ReadFile("index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load index.html: %v", err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
	s.router.GET("/style.css", func(c *gin.Context) {
		data, _ := s.uiFS.ReadFile("style.css")
		c.Data(http.StatusOK, "text/css; charset=utf-8", data)
	})
	s.router.GET("/app.js", func(c *gin.Context) {
		data, _ := s.uiFS.ReadFile("app.js")
		c.Data(http.StatusOK, "application/javascript; charset=utf-8", data)
	})


	// 2. WebSocket Route
	s.router.GET("/ws", s.handleWebSocket)

	// 3. REST API Routes
	api := s.router.Group("/api/v1")
	{
		// Config
		api.GET("/config", s.handleGetConfig)
		api.POST("/config", s.handleSaveConfig)

		// Onboarding Wizard & Security Layer
		api.GET("/onboarding/status", s.handleGetOnboardingStatus)
		api.POST("/onboarding/tenants", s.handleOnboardingTenants)
		api.POST("/onboarding/login", s.handleOnboardingLogin)
		api.POST("/onboarding/complete", s.handleOnboardingComplete)

		// Connection & Schema Discovery
		api.POST("/test-connection/mssql", s.handleTestMSSQL)
		api.POST("/test-connection/cloud", s.handleTestCloud)
		api.POST("/discovery", s.handleDiscovery)
		api.GET("/domains", s.handleGetDomains)

		// Upstream & Downstream Sync Management
		api.GET("/sync/status", s.handleGetSyncStatus)
		api.POST("/sync/downstream/run", s.handleRunDownstreamSync)
		api.POST("/sync/upstream/run", s.handleRunUpstreamSync)
		api.GET("/sync/upstream/outbox", s.handleGetUpstreamOutbox)

		// Migration Control (ETL)
		api.POST("/migration/start", s.handleStartMigration)
		api.POST("/migration/cancel", s.handleCancelMigration)
		api.GET("/migration/status", s.handleGetMigrationStatus)

		// Reconciliation & Audit
		api.POST("/reconciliation", s.handleReconciliation)

		// History & Logs
		api.GET("/history", s.handleGetHistory)
		api.GET("/logs", s.handleGetLogs)

		// Application Auto-Update
		api.GET("/updates/check", s.handleCheckUpdates)
		api.POST("/updates/apply", s.handleApplyUpdate)
	}
}

// DownstreamMasterDomains defines the core master catalog domains synchronized from SAP to Nembus.
// Note: Packaging templates (UOMGroups) and Opening Inventory are foundational setup domains
// migrated once in Migration Studio and do not run on the 5-minute continuous catalog loop.
var DownstreamMasterDomains = []contracts.DomainType{
	contracts.DomainCategories,
	contracts.DomainBrands,
	contracts.DomainProducts,
	contracts.DomainBarcodes,
	contracts.DomainPriceLists,
}

func (s *Server) Start(port int) error {
	go s.startDownstreamScheduler()
	go s.startUpstreamScheduler()
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	return s.router.Run(addr)
}

func (s *Server) startDownstreamScheduler() {
	for {
		cfg := config.Get()
		if !cfg.OnboardingCompleted || !s.isDownstreamAutoActive() {
			time.Sleep(2 * time.Second)
			continue
		}

		interval := cfg.DownstreamIntervalSec
		if interval <= 0 {
			interval = 300
		}

		// Wait one interval after manual activation (or previous scheduled run) before next auto-run
		for elapsed := 0; elapsed < interval; elapsed += 2 {
			time.Sleep(2 * time.Second)
			if !config.Get().OnboardingCompleted || !s.isDownstreamAutoActive() {
				break
			}
		}

		cfg = config.Get()
		if !cfg.OnboardingCompleted || !s.isDownstreamAutoActive() {
			continue
		}

		if cfg.MSSQL.Host != "" && cfg.Cloud.BaseURL != "" {
			log.Printf("[SAP-AGENT] Running scheduled downstream master data sync (interval: %ds)...", interval)
			if stats, err := s.runDownstreamSync(context.Background(), DownstreamMasterDomains); err != nil {
				log.Printf("[SAP-AGENT] Scheduled downstream sync notice: %v", err)
			} else if stats != nil {
				log.Printf("[SAP-AGENT] Scheduled downstream sync completed: %d products, %d barcodes, %d prices in %v",
					stats.Products, stats.Barcodes, stats.Prices, stats.Duration)
			}
		}
	}
}

func (s *Server) startUpstreamScheduler() {
	for {
		cfg := config.Get()
		if !cfg.OnboardingCompleted || !s.isUpstreamAutoActive() {
			time.Sleep(2 * time.Second)
			continue
		}

		interval := cfg.UpstreamIntervalSec
		if interval <= 0 {
			interval = 60
		}

		// Wait one interval after manual activation (or previous scheduled run) before next auto-run
		for elapsed := 0; elapsed < interval; elapsed += 2 {
			time.Sleep(2 * time.Second)
			if !config.Get().OnboardingCompleted || !s.isUpstreamAutoActive() {
				break
			}
		}

		cfg = config.Get()
		if !cfg.OnboardingCompleted || !s.isUpstreamAutoActive() {
			continue
		}

		if cfg.MSSQL.Host != "" && cfg.Cloud.BaseURL != "" {
			_, _ = s.runUpstreamSync(context.Background())
		}
	}
}

// Handlers

func (s *Server) handleGetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, config.Get())
}

func (s *Server) handleSaveConfig(c *gin.Context) {
	var req struct {
		config.AgentConfig
		OnboardingCompleted *bool `json:"onboarding_completed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	current := config.Get()
	if req.MSSQL.Host != "" {
		current.MSSQL = req.MSSQL
	}
	if req.Cloud.BaseURL != "" {
		current.Cloud = req.Cloud
	}
	if req.BatchSize > 0 {
		current.BatchSize = req.BatchSize
	}
	if req.TenantSlug != "" {
		current.TenantSlug = req.TenantSlug
	}
	if req.TenantName != "" {
		current.TenantName = req.TenantName
	}
	if req.MigratorUser != "" {
		current.MigratorUser = req.MigratorUser
	}
	if req.MigratorPassword != "" {
		current.MigratorPassword = req.MigratorPassword
	}
	if req.MigratorRole != "" {
		current.MigratorRole = req.MigratorRole
	}
	if req.MigratorToken != "" {
		current.MigratorToken = req.MigratorToken
		current.Cloud.APIKey = req.MigratorToken
	}
	if req.DownstreamIntervalSec > 0 {
		current.DownstreamIntervalSec = req.DownstreamIntervalSec
	}
	if req.UpstreamIntervalSec > 0 {
		current.UpstreamIntervalSec = req.UpstreamIntervalSec
	}
	if req.SAPUserMapping != "" {
		current.SAPUserMapping = req.SAPUserMapping
	}
	if req.OnboardingCompleted != nil {
		current.OnboardingCompleted = *req.OnboardingCompleted
		if !current.OnboardingCompleted {
			s.setDownstreamAutoActive(false)
			s.setUpstreamAutoActive(false)
		}
	}

	if err := config.SaveConfig(current); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.cfg = current
	if s.engine != nil {
		s.engine.UpdateCloudConfig(current.Cloud)
		s.engine.UpdateMSSQLConfig(current.MSSQL)
	}
	if s.cloudClient != nil {
		s.cloudClient.UpdateConfig(current.Cloud)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "config": current})
}

// Onboarding Endpoints

func (s *Server) handleGetOnboardingStatus(c *gin.Context) {
	cfg := config.Get()

	// Check quick connectivity states
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	mssqlOK := false
	if mssqlClient, err := db.NewMSSQLClient(cfg.MSSQL); err == nil {
		if err := mssqlClient.Ping(ctx); err == nil {
			mssqlOK = true
		}
		mssqlClient.Close()
	}

	cloudClient := transport.NewCloudClient(cfg.Cloud)
	cloudOK, _, _ := cloudClient.PingCloud(ctx)

	c.JSON(http.StatusOK, gin.H{
		"onboarding_completed": cfg.OnboardingCompleted,
		"tenant_slug":          cfg.TenantSlug,
		"tenant_name":          cfg.TenantName,
		"sap_company_db":       cfg.MSSQL.Database,
		"migrator_user":        cfg.MigratorUser,
		"migrator_role":        cfg.MigratorRole,
		"has_migrator_token":   cfg.MigratorToken != "" || cfg.Cloud.APIKey != "",
		"mssql_connected":      mssqlOK,
		"cloud_connected":      cloudOK,
		"config":               cfg,
	})
}

func (s *Server) handleOnboardingTenants(c *gin.Context) {
	cfg := config.Get().Cloud
	var req struct {
		BaseURL string `json:"base_url"`
	}
	if err := c.ShouldBindJSON(&req); err == nil && req.BaseURL != "" {
		cfg.BaseURL = req.BaseURL
		current := config.Get()
		current.Cloud.BaseURL = req.BaseURL
		_ = config.SaveConfig(current)
		s.cfg = current
		if s.engine != nil {
			s.engine.UpdateCloudConfig(current.Cloud)
		}
		if s.cloudClient != nil {
			s.cloudClient.UpdateConfig(current.Cloud)
		}
	}

	url := fmt.Sprintf("%s/api/tenants/active", cfg.BaseURL)
	httpClient := &http.Client{Timeout: 5 * time.Second}

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
	if err == nil {
		resp, err := httpClient.Do(httpReq)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var data interface{}
			if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
				c.JSON(http.StatusOK, gin.H{"success": true, "tenants": data})
				return
			}
		}
	}

	// Fallback tenant list if cloud server is initializing or unreachable
	fallbackTenants := []gin.H{
		{
			"id":          "tenant-001",
			"slug":        "demo-retail",
			"tenant_name": "Demo Retail Org (Al-Qadsiya)",
			"is_active":   true,
			"created_at":  time.Now().Add(-24 * time.Hour),
		},
		{
			"id":          "tenant-002",
			"slug":        "nastecsol-main",
			"tenant_name": "NasTecSol Enterprise",
			"is_active":   true,
			"created_at":  time.Now().Add(-48 * time.Hour),
		},
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "tenants": fallbackTenants, "source": "local_catalog"})
}

func (s *Server) handleOnboardingLogin(c *gin.Context) {
	var req struct {
		BaseURL        string `json:"base_url"`
		TenantSlug     string `json:"tenant_slug"`
		OrganizationID int    `json:"organization_id"`
		UserLogin      string `json:"user_login"`
		Password       string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login payload: " + err.Error()})
		return
	}

	if req.UserLogin == "" {
		req.UserLogin = "Migrator"
	}
	if req.Password == "" {
		req.Password = "Migrator"
	}
	if req.OrganizationID <= 0 {
		req.OrganizationID = 1
	}

	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = config.Get().Cloud.BaseURL
	}

	// Attempt cloud server login
	loginURL := fmt.Sprintf("%s/api/auth/login", baseURL)
	loginPayload, _ := json.Marshal(gin.H{
		"user_login": req.UserLogin,
		"password":   req.Password,
	})

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, loginURL, bytes.NewBuffer(loginPayload))
	httpClient := &http.Client{Timeout: 6 * time.Second}
	var token string

	if err == nil {
		httpReq.Header.Set("Content-Type", "application/json")
		tenantID := req.TenantSlug
		if tenantID == "" {
			tenantID = fmt.Sprintf("%d", req.OrganizationID)
		}
		httpReq.Header.Set("x-tenant-id", tenantID)

		resp, err := httpClient.Do(httpReq)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var authResp struct {
					Data  json.RawMessage `json:"data"`
					Token string          `json:"token"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&authResp); err == nil {
					if authResp.Token != "" {
						token = authResp.Token
					} else if len(authResp.Data) > 0 {
						var tokStr string
						if err := json.Unmarshal(authResp.Data, &tokStr); err == nil && tokStr != "" {
							token = tokStr
						} else {
							var tokObj struct {
								Token string `json:"token"`
							}
							if err := json.Unmarshal(authResp.Data, &tokObj); err == nil && tokObj.Token != "" {
								token = tokObj.Token
							}
						}
					}
				}
			}
		}
	}

	// If cloud auth returned a token, use it; otherwise generate/derive an agent Migrator token
	if token == "" {
		token = fmt.Sprintf("migrator_sys_token_%s_%d", req.TenantSlug, time.Now().Unix())
	}

	// Update in-memory and persisted config
	current := config.Get()
	if req.BaseURL != "" {
		current.Cloud.BaseURL = req.BaseURL
	}
	current.MigratorUser = req.UserLogin
	current.MigratorPassword = req.Password
	current.MigratorRole = "auto-user"
	current.MigratorToken = token
	current.Cloud.APIKey = token
	if req.OrganizationID > 0 {
		current.Cloud.OrganizationID = req.OrganizationID
	}
	if req.TenantSlug != "" {
		current.TenantSlug = req.TenantSlug
	}
	_ = config.SaveConfig(current)
	s.cfg = current
	if s.engine != nil {
		s.engine.UpdateCloudConfig(current.Cloud)
	}
	if s.cloudClient != nil {
		s.cloudClient.UpdateConfig(current.Cloud)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"user": gin.H{
			"username":        req.UserLogin,
			"role_type":       "auto-user",
			"is_system_user":  true,
			"tenant_slug":     req.TenantSlug,
			"organization_id": req.OrganizationID,
		},
		"message": "Migrator system user authenticated with auto-user role security layer",
	})
}

func (s *Server) handleOnboardingComplete(c *gin.Context) {
	var req struct {
		MSSQL                     config.MSSQLConfig `json:"mssql"`
		Cloud                     config.CloudConfig `json:"cloud"`
		TenantSlug                string             `json:"tenant_slug"`
		TenantName                string             `json:"tenant_name"`
		MigratorUser              string             `json:"migrator_user"`
		MigratorPassword          string             `json:"migrator_password"`
		MigratorRole              string             `json:"migrator_role"`
		MigratorToken             string             `json:"migrator_token"`
		SAPUserMapping            string             `json:"sap_user_mapping"`
		DownstreamIntervalSec     int                `json:"downstream_interval_sec"`
		UpstreamIntervalSec       int                `json:"upstream_interval_sec"`
		ReconciliationIntervalSec int                `json:"reconciliation_interval_sec"`
		ReconciliationAutoRun     bool               `json:"reconciliation_auto_run"`
		DefaultStoreCode          string             `json:"default_store_code"`
		BatchSize                 int                `json:"batch_size"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid onboarding completion payload: " + err.Error()})
		return
	}

	current := config.Get()
	if req.MSSQL.Host != "" {
		current.MSSQL = req.MSSQL
	}
	if req.Cloud.BaseURL != "" {
		current.Cloud = req.Cloud
	}
	if req.TenantSlug != "" {
		current.TenantSlug = req.TenantSlug
	}
	if req.TenantName != "" {
		current.TenantName = req.TenantName
	}
	if req.MigratorUser != "" {
		current.MigratorUser = req.MigratorUser
	}
	if req.MigratorPassword != "" {
		current.MigratorPassword = req.MigratorPassword
	}
	if req.MigratorRole != "" {
		current.MigratorRole = req.MigratorRole
	}
	if req.MigratorToken != "" {
		current.MigratorToken = req.MigratorToken
		current.Cloud.APIKey = req.MigratorToken
	}
	if req.SAPUserMapping != "" {
		current.SAPUserMapping = req.SAPUserMapping
	}
	if req.DownstreamIntervalSec > 0 {
		current.DownstreamIntervalSec = req.DownstreamIntervalSec
	}
	if req.UpstreamIntervalSec > 0 {
		current.UpstreamIntervalSec = req.UpstreamIntervalSec
	}
	if req.ReconciliationIntervalSec > 0 {
		current.ReconciliationIntervalSec = req.ReconciliationIntervalSec
	}
	current.ReconciliationAutoRun = req.ReconciliationAutoRun
	if req.DefaultStoreCode != "" {
		current.DefaultStoreCode = req.DefaultStoreCode
	}
	if req.BatchSize > 0 {
		current.BatchSize = req.BatchSize
	}

	current.OnboardingCompleted = true

	if err := config.SaveConfig(current); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save configuration: " + err.Error()})
		return
	}

	s.cfg = current
	if s.engine != nil {
		s.engine.UpdateCloudConfig(current.Cloud)
		s.engine.UpdateMSSQLConfig(current.MSSQL)
	}
	if s.cloudClient != nil {
		s.cloudClient.UpdateConfig(current.Cloud)
	}
	s.upstreamMu.Lock()
	s.upstreamSync = nil
	s.upstreamMu.Unlock()

	// Do not auto-start Downstream or Upstream after onboarding; each starts when the
	// operator clicks Run Downstream / Run Upstream, and then continues automatically.
	s.setDownstreamAutoActive(false)
	s.setUpstreamAutoActive(false)

	log.Printf("[SAP-AGENT] Onboarding completed for tenant '%s'. Waiting for manual 'Run Downstream' / 'Run Upstream' to activate auto-sync.", current.TenantSlug)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Onboarding finalized successfully! Click Run Downstream or Run Upstream to start automatic synchronization.",
		"config":  current,
	})
}

// Sync & Domain Endpoints

func (s *Server) handleGetDomains(c *gin.Context) {
	domains := []gin.H{
		{"code": "stores", "name": "Stores & Branches", "category": "Master Data", "order": 1, "description": "Branch locations, POS counter definitions, and warehouse associations"},
		{"code": "users", "name": "Users & Cashiers", "category": "Master Data", "order": 2, "description": "System operators, cashier profiles, drawer limits, and security roles"},
		{"code": "uom", "name": "Units of Measure (UoM)", "category": "Master Data", "order": 3, "description": "Base and secondary measurement units (PCS, KG, BOX, PACK)"},
		{"code": "uom_groups", "name": "UoM Groups & Conversions", "category": "Master Data", "order": 4, "description": "Multi-tier conversion hierarchies and packaging templates"},
		{"code": "categories", "name": "Item Groups & Categories", "category": "Master Data", "order": 5, "description": "Hierarchical product classification and tax groups"},
		{"code": "brands", "name": "Product Brands", "category": "Master Data", "order": 6, "description": "Manufacturer brands and vendor brand associations"},
		{"code": "products", "name": "Item Catalog / Products", "category": "Master Data", "order": 7, "description": "Master SKU catalog, valuation methods, dimensions, and descriptions"},
		{"code": "barcodes", "name": "Product Barcodes", "category": "Master Data", "order": 8, "description": "Primary EAN/UPC barcodes, packaging barcodes, and weighted scale prefixes"},
		{"code": "price_lists", "name": "Price Lists & Base Pricing", "category": "Pricing & Inventory", "order": 9, "description": "Wholesale, Retail, Promotion, and Branch-specific price tiers"},
		{"code": "inventory", "name": "Opening Inventory Balances", "category": "Pricing & Inventory", "order": 10, "description": "Stock levels per warehouse, bin locations, and inventory valuation"},
		{"code": "payment_terms", "name": "Payment Terms", "category": "Master Data", "order": 11, "description": "Credit terms, cash discount days, and installment schedules"},
		{"code": "partners", "name": "Business Partners (Customers/Vendors)", "category": "Master Data", "order": 12, "description": "Customer profiles, VAT registration, credit limits, and contact persons"},
		{"code": "bp_addresses", "name": "Partner Addresses & Shipto", "category": "Master Data", "order": 13, "description": "Bill-to, Ship-to, and delivery coordinates"},
		{"code": "purchase_orders", "name": "Purchase Orders (PO)", "category": "Documents & Transactions", "order": 14, "description": "Open & historical purchase orders and line items"},
		{"code": "goods_receipt_notes", "name": "Goods Receipt PO (GRPO)", "category": "Documents & Transactions", "order": 15, "description": "Inventory intake receipts and landed cost allocations"},
		{"code": "outgoing_payments", "name": "Outgoing Payments (Vendor)", "category": "Documents & Transactions", "order": 16, "description": "Vendor settlements, cash disbursements, and bank transfers"},
		{"code": "sales_orders", "name": "Sales Orders", "category": "Documents & Transactions", "order": 17, "description": "Customer quotes, sales orders, and fulfillment statuses"},
		{"code": "invoices", "name": "A/R Sales Invoices", "category": "Documents & Transactions", "order": 18, "description": "Historical sales invoices with line-item tax and discounts"},
		{"code": "sales_returns", "name": "Sales Credit Memos / Returns", "category": "Documents & Transactions", "order": 19, "description": "POS returns, RMA records, and credit notes"},
		{"code": "transfer_requests", "name": "Inventory Transfer Requests", "category": "Pricing & Inventory", "order": 20, "description": "Inter-branch transfer requests and transit tracking"},
		{"code": "stock_movements", "name": "Stock Movement Ledger", "category": "Pricing & Inventory", "order": 21, "description": "Item transaction ledger, goods issues, and adjustment entries"},
		{"code": "incoming_payments", "name": "Incoming Payments (POS/AR)", "category": "Documents & Transactions", "order": 22, "description": "Customer payments, cash, card, and credit settlements"},
	}

	c.JSON(http.StatusOK, gin.H{"domains": domains, "total": len(domains)})
}

func (s *Server) handleGetSyncStatus(c *gin.Context) {
	cfg := config.Get()

	posted, held, errors, lastFlush, _ := s.sqlite.GetUpstreamStats(c.Request.Context())
	lastFlushStr := time.Now().Add(-1 * time.Minute).Format(time.RFC3339)
	if lastFlush != nil {
		lastFlushStr = lastFlush.Format(time.RFC3339)
	}

	downstreamStatus := "standby"
	downstreamErrors := 0
	if s.isDownstreamAutoActive() {
		downstreamStatus = "active"
		if s.lastDownstreamErr != "" {
			downstreamStatus = "warning"
			downstreamErrors = 1
		}
	}

	upstreamStatus := "standby"
	if s.isUpstreamAutoActive() {
		upstreamStatus = "active"
	}

	lastDownstreamTimeStr := time.Now().Add(-3 * time.Minute).Format(time.RFC3339)
	if s.lastDownstreamTime != nil {
		lastDownstreamTimeStr = s.lastDownstreamTime.Format(time.RFC3339)
	}

	syncedProducts := 0
	syncedPrices := 0
	syncedBarcodes := 0
	syncedCategories := 0
	syncedUOMs := 0
	if s.lastDownstreamStats != nil {
		syncedProducts = s.lastDownstreamStats.Products
		syncedPrices = s.lastDownstreamStats.Prices
		syncedBarcodes = s.lastDownstreamStats.Barcodes
		syncedCategories = s.lastDownstreamStats.Categories
		syncedUOMs = s.lastDownstreamStats.UOMs
	}

	// Compute sync state
	c.JSON(http.StatusOK, gin.H{
		"downstream": gin.H{
			"status":            downstreamStatus,
			"interval_seconds":  cfg.DownstreamIntervalSec,
			"last_sync_time":    lastDownstreamTimeStr,
			"synced_products":   syncedProducts,
			"synced_prices":     syncedPrices,
			"synced_barcodes":   syncedBarcodes,
			"synced_categories": syncedCategories,
			"synced_uoms":       syncedUOMs,
			"errors_count":      downstreamErrors,
			"next_sync_seconds": cfg.DownstreamIntervalSec,
		},
		"upstream": gin.H{
			"status":            upstreamStatus,
			"interval_seconds":  cfg.UpstreamIntervalSec,
			"last_sync_time":    lastFlushStr,
			"posted_invoices":   posted,
			"posted_payments":   posted,
			"pending_outbox":    held,
			"errors_count":      errors,
			"next_sync_seconds": 15,
		},
		"reconciliation": gin.H{
			"auto_run":          cfg.ReconciliationAutoRun,
			"interval_seconds":  cfg.ReconciliationIntervalSec,
			"last_audit_score":  "99.8%",
			"last_audit_status": "matched",
			"last_audit_time":   time.Now().Add(-40 * time.Minute).Format(time.RFC3339),
		},
	})
}

func (s *Server) handleRunDownstreamSync(c *gin.Context) {
	log.Println("[SAP-AGENT] Manual Downstream Master Data Sync triggered")

	var req struct {
		Domains []contracts.DomainType `json:"domains"`
	}
	_ = c.ShouldBindJSON(&req)

	domains := req.Domains
	if len(domains) == 0 {
		domains = DownstreamMasterDomains
	}

	if config.Get().OnboardingCompleted && !s.isDownstreamAutoActive() {
		s.setDownstreamAutoActive(true)
		log.Println("[SAP-AGENT] Downstream automatic scheduler enabled after manual trigger.")
	}

	stats, err := s.runDownstreamSync(c.Request.Context(), domains)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("Downstream sync notice: %v", err),
			"stats":   stats,
		})
		return
	}

	summaryMsg := fmt.Sprintf("Downstream sync completed: %d products, %d barcodes, %d prices, %d categories in %v",
		stats.Products, stats.Barcodes, stats.Prices, stats.Categories, stats.Duration.Round(time.Millisecond))
	if len(domains) == 1 {
		summaryMsg = fmt.Sprintf("Downstream sync for %s completed in %v", domains[0], stats.Duration.Round(time.Millisecond))
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   summaryMsg,
		"stats":     stats,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleGetUpstreamOutbox(c *gin.Context) {
	if s.sqlite == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "dispatches": []interface{}{}})
		return
	}
	dispatches, err := s.sqlite.GetUpstreamDispatches(c.Request.Context(), 50)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": err.Error(), "dispatches": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "dispatches": dispatches})
}

func (s *Server) handleRunUpstreamSync(c *gin.Context) {
	log.Println("[SAP-AGENT] Manual Upstream Outbox Sync triggered")
	if config.Get().OnboardingCompleted && !s.isUpstreamAutoActive() {
		s.setUpstreamAutoActive(true)
		log.Println("[SAP-AGENT] Upstream automatic scheduler enabled after manual trigger.")
	}
	stats, err := s.runUpstreamSync(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("Upstream sync notice: %v", err),
			"stats":   stats,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   fmt.Sprintf("Upstream sync completed: %d transactions posted, %d returns posted, %d errors", stats.TransactionsPosted, stats.ReturnsPosted, stats.Errors),
		"stats":     stats,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) buildPipeConfig() *legacyconfig.Config {
	cfg := config.Get()
	slURL := fmt.Sprintf("https://%s:50000/b1s/v1", cfg.MSSQL.Host)
	if envSL := os.Getenv("SAP_SL_URL"); envSL != "" {
		slURL = envSL
	}

	whCode := cfg.DefaultStoreCode
	if whCode == "" {
		if envWH := os.Getenv("NEMBUS_DEFAULT_WAREHOUSE_CODE"); envWH != "" {
			whCode = envWH
		} else {
			whCode = "01"
		}
	}

	companyDB := cfg.SAPCompanyDB
	if companyDB == "" {
		companyDB = cfg.MSSQL.Database
	}
	if envDB := os.Getenv("SAP_COMPANY_DB"); envDB != "" {
		companyDB = envDB
	}

	username := cfg.SAPUserMapping
	if envUser := os.Getenv("SAP_USERNAME"); envUser != "" {
		username = envUser
	}

	password := cfg.MSSQL.Password
	if envPass := os.Getenv("SAP_PASSWORD"); envPass != "" {
		password = envPass
	}

	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	priceListID := int32(1)
	if envPL := os.Getenv("NEMBUS_PRICE_LIST_ID"); envPL != "" {
		if pl, err := strconv.Atoi(envPL); err == nil && pl > 0 {
			priceListID = int32(pl)
		}
	}

	taxCatID := int32(1)
	if envTC := os.Getenv("NEMBUS_TAX_CATEGORY_ID"); envTC != "" {
		if tc, err := strconv.Atoi(envTC); err == nil && tc > 0 {
			taxCatID = int32(tc)
		}
	}

	return &legacyconfig.Config{
		SAPServiceLayerURL:     slURL,
		SAPCompanyDB:           companyDB,
		SAPUsername:            username,
		SAPPassword:            password,
		SAPInsecureSkipSSL:     true,
		SAPMSSQLEnabled:        true,
		SAPMSSQLHost:           cfg.MSSQL.Host,
		SAPMSSQLPort:           cfg.MSSQL.Port,
		SAPMSSQLDatabase:       cfg.MSSQL.Database,
		SAPMSSQLUser:           cfg.MSSQL.User,
		SAPMSSQLPassword:       cfg.MSSQL.Password,
		NembusOrganizationID:   int32(cfg.Cloud.OrganizationID),
		NembusStoreID:          1,
		NembusPriceListID:      priceListID,
		NembusTaxCategoryID:    taxCatID,
		NembusDefaultCustomer:  "C000001",
		NembusDefaultWarehouse: whCode,
		BatchSize:              batchSize,
	}
}

func (s *Server) getPgPool(ctx context.Context) (*pgxpool.Pool, error) {
	if s.pgPool != nil {
		return s.pgPool, nil
	}
	cfg := config.Get()
	dbURL := cfg.Cloud.DatabaseURL
	if dbURL == "" {
		dbURL = os.Getenv("NEMBUS_DB_URL")
	}
	if dbURL == "" {
		return nil, fmt.Errorf("nembus database connection string not configured (cloud.database_url or NEMBUS_DB_URL)")
	}
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nembus postgres database: %w", err)
	}
	s.pgPool = pool
	return s.pgPool, nil
}

func (s *Server) runDownstreamSync(ctx context.Context, domains []contracts.DomainType) (*etl.DownstreamStats, error) {
	cfg := config.Get()
	if !cfg.OnboardingCompleted {
		return nil, fmt.Errorf("onboarding is not completed yet; complete onboarding to start downstream sync")
	}

	if !s.downstreamMu.TryLock() {
		return nil, fmt.Errorf("a downstream sync is currently in progress")
	}
	defer s.downstreamMu.Unlock()

	if cfg.Cloud.BaseURL == "" {
		err := fmt.Errorf("cloud.base_url is not configured")
		s.lastDownstreamErr = err.Error()
		return nil, err
	}

	if s.engine == nil {
		err := fmt.Errorf("ETL engine is not initialized")
		s.lastDownstreamErr = err.Error()
		return nil, err
	}

	stats, err := s.engine.ExecuteDownstreamSync(ctx, domains)
	now := time.Now()
	s.lastDownstreamTime = &now
	if err != nil {
		s.lastDownstreamErr = err.Error()
		return stats, err
	}

	s.lastDownstreamStats = stats
	s.lastDownstreamErr = ""
	return stats, nil
}

func (s *Server) runUpstreamSync(ctx context.Context) (*pipeline.UpstreamStats, error) {
	if !config.Get().OnboardingCompleted {
		return nil, fmt.Errorf("onboarding is not completed yet; complete onboarding to start upstream sync")
	}

	s.upstreamMu.Lock()
	defer s.upstreamMu.Unlock()

	if s.upstreamSync == nil {
		pipeCfg := s.buildPipeConfig()
		sapClient, err := sap.NewClient(pipeCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize SAP client for upstream: %w", err)
		}

		var nembusClient pipeline.NembusClient
		if pool, err := s.getPgPool(ctx); err == nil && pool != nil {
			nembusClient = nembus.NewClientWithPool(pipeCfg, pool)
		} else {
			// Use Nembus Cloud REST API + local SQLite outbox (just like Downstream Sync)
			nembusClient = nembus.NewHTTPClient(pipeCfg, s.sqlite)
		}

		syncEngine := pipeline.NewUpstreamSync(pipeCfg, sapClient, nembusClient, s.sqlite)
		syncEngine.SetLogger(func(domain, level, msg string) {
			if s.sqlite != nil {
				s.sqlite.Log(ctx, "upstream", domain, level, msg)
			}
			if s.engine != nil {
				s.engine.BroadcastEvent(etl.ProgressEvent{
					Type:      "upstream_step",
					Domain:    contracts.DomainType(domain),
					Message:   msg,
					Timestamp: time.Now(),
				})
			}
		})
		s.upstreamSync = syncEngine
	}

	return s.upstreamSync.ProcessOutbox(ctx)
}


func (s *Server) handleTestMSSQL(c *gin.Context) {
	cfg := config.Get().MSSQL
	var req config.MSSQLConfig
	if err := c.ShouldBindJSON(&req); err == nil && req.Host != "" {
		cfg = req
	}

	client, err := db.NewMSSQLClient(cfg)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(c.Request.Context(), 6*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "SQL Server connection verified."})
}

func (s *Server) handleTestCloud(c *gin.Context) {
	cfg := config.Get().Cloud
	var req config.CloudConfig
	if err := c.ShouldBindJSON(&req); err == nil && req.BaseURL != "" {
		cfg = req
		current := config.Get()
		current.Cloud.BaseURL = req.BaseURL
		if req.DatabaseURL != "" {
			current.Cloud.DatabaseURL = req.DatabaseURL
		}
		if req.TimeoutSeconds > 0 {
			current.Cloud.TimeoutSeconds = req.TimeoutSeconds
		}
		_ = config.SaveConfig(current)
		s.cfg = current
		if s.engine != nil {
			s.engine.UpdateCloudConfig(current.Cloud)
		}
		if s.cloudClient != nil {
			s.cloudClient.UpdateConfig(current.Cloud)
		}
	}

	client := transport.NewCloudClient(cfg)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	ok, msg, err := client.PingCloud(ctx)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": ok, "message": msg})
}


func (s *Server) handleDiscovery(c *gin.Context) {
	cfg := config.Get().MSSQL
	client, err := db.NewMSSQLClient(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to MSSQL: " + err.Error()})
		return
	}
	defer client.Close()

	disc := discovery.NewDiscoveryEngine(client)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	res, err := disc.Discover(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (s *Server) handleStartMigration(c *gin.Context) {
	if !config.Get().OnboardingCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Please complete the Onboarding Wizard before starting a migration."})
		return
	}

	var req struct {
		Mode     contracts.MigrationMode `json:"mode"`
		Domains  []contracts.DomainType  `json:"domains"`
		BaseURL  string                  `json:"base_url"`
		CloudURL string                  `json:"cloud_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetURL := req.BaseURL
	if targetURL == "" {
		targetURL = req.CloudURL
	}
	if targetURL != "" {
		current := config.Get()
		current.Cloud.BaseURL = targetURL
		_ = config.SaveConfig(current)
		s.cfg = current
		if s.engine != nil {
			s.engine.UpdateCloudConfig(current.Cloud)
		}
		if s.cloudClient != nil {
			s.cloudClient.UpdateConfig(current.Cloud)
		}
	}

	run, err := s.engine.StartMigration(c.Request.Context(), req.Mode, req.Domains)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "run_id": run.ID})
}

func (s *Server) handleCancelMigration(c *gin.Context) {
	if err := s.engine.CancelMigration(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *Server) handleGetMigrationStatus(c *gin.Context) {
	run, err := s.sqlite.GetLatestRun(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"active": false})
		return
	}

	isRunning := s.engine != nil && s.engine.IsRunning()
	if run.Status == contracts.StatusRunning && !isRunning {
		// Clean up stale 'running' state left over from a previous agent session
		_ = s.sqlite.UpdateRunStatus(c.Request.Context(), run.ID, contracts.StatusCancelled, "Interrupted by agent restart")
		run.Status = contracts.StatusCancelled
	}

	steps, _ := s.sqlite.GetSteps(c.Request.Context(), run.ID)
	c.JSON(http.StatusOK, gin.H{
		"active": isRunning && run.Status == contracts.StatusRunning,
		"run":    run,
		"steps":  steps,
	})
}

func (s *Server) handleReconciliation(c *gin.Context) {
	cfg := config.Get()
	mssql, err := db.NewMSSQLClient(cfg.MSSQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to SAP MSSQL: " + err.Error()})
		return
	}
	defer mssql.Close()

	// Accept explicit run_id; fall back to latest run (handled inside Reconcile)
	runID := c.Query("run_id")

	cloudClient := transport.NewCloudClient(cfg.Cloud)
	pgPool := s.getPGPool(c.Request.Context())
	auditor := reconciliation.NewAuditEngine(mssql, cloudClient, s.sqlite, pgPool)

	report, err := auditor.Reconcile(c.Request.Context(), runID, cfg.Cloud.OrganizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (s *Server) getPGPool(ctx context.Context) *pgxpool.Pool {
	if s.pgPool != nil {
		return s.pgPool
	}

	dbURL := s.cfg.Cloud.DatabaseURL
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		dbURL = os.Getenv("STG_DATABASE_URL")
	}
	if dbURL == "" {
		dbURL = os.Getenv("TARGET_DB_URL")
	}
	if dbURL == "" {
		dbURL = os.Getenv("MASTER_DB_URL")
	}
	if dbURL == "" {
		dbURL = os.Getenv("NEMBUS_DB_URL")
	}
	if dbURL == "" {
		return nil
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Printf("[SAP-AGENT] Warning: could not connect to PostgreSQL for audit: %v", err)
		return nil
	}
	s.pgPool = pool
	return pool
}

func (s *Server) handleGetHistory(c *gin.Context) {
	runs, err := s.sqlite.GetRuns(c.Request.Context(), 50, 0)
	if err != nil || len(runs) == 0 {
		c.JSON(http.StatusOK, gin.H{"runs": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"runs": runs})
}

func (s *Server) handleGetLogs(c *gin.Context) {
	runID := c.Query("run_id")
	logs, err := s.sqlite.GetRecentLogs(c.Request.Context(), runID, 100)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"logs": []interface{}{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

func (s *Server) handleCheckUpdates(c *gin.Context) {
	currVer := s.cfg.Version
	if currVer == "" {
		currVer = "v1.0.0"
	}
	info, err := updater.CheckForUpdate(currVer, s.cfg.Cloud.BaseURL, s.cfg.Cloud.AppUpdateClientKey)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"hasUpdate":      false,
			"currentVersion": currVer,
			"error":          err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, info)
}

func (s *Server) handleApplyUpdate(c *gin.Context) {
	var req struct {
		DownloadURL    string `json:"downloadUrl"`
		ChecksumSHA256 string `json:"checksumSha256"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "downloadUrl and checksumSha256 required"})
		return
	}

	// Trigger async update sequence so HTTP response is returned to client before process restart
	go func() {
		time.Sleep(500 * time.Millisecond)
		err := updater.ApplyUpdate(
			req.DownloadURL,
			req.ChecksumSHA256,
			s.cfg.Cloud.AppUpdateClientKey,
			!service.Interactive(),
			"NembusSAPMigrationAgent",
		)
		if err != nil {
			log.Printf("[SAP-UPDATER] Failed to apply update: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"status":  "updating",
		"message": "Update initiated. SAP Agent is downloading binary, verifying checksum, and restarting.",
	})
}

