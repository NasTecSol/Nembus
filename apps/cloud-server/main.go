package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/NasTecSol/nembus-core/config"
	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/NasTecSol/nembus-core/enrichment/openaiadapter"
	"github.com/NasTecSol/nembus-core/grpc/backuppb"
	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/middleware/manager"
	"github.com/NasTecSol/nembus-core/repository"
	router "github.com/NasTecSol/nembus-core/routing"
	"github.com/NasTecSol/nembus-core/usecase"
	grpcbackup "github.com/NasTecSol/nembus-server/internal/grpc"
	cloudzatca "github.com/NasTecSol/nembus-server/internal/zatca"

	_ "github.com/NasTecSol/nembus-server/docs/swagger" // Swagger generated docs

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// @title           NEMBUS API
// @version         1.0
// @description     NEMBUS Backend API - Nasar Entity-driven Modular Business Unified System
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  MIT
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// setupDatabase initializes and returns the master database connection pool and repository
func setupDatabase(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, *repository.Queries, error) {
	if cfg.MasterDBURL == "" {
		log.Fatal("MASTER_DB_URL is not set")
	}

	pool, err := pgxpool.New(ctx, cfg.MasterDBURL)
	if err != nil {
		return nil, nil, err
	}
	// Initialize SQLC repository
	queries := repository.New(pool)

	return pool, queries, nil
}

func setupRouter(tenantManager *manager.Manager, masterRepo *repository.Queries, userUC *usecase.UserUseCase, orgUC *usecase.OrganizationUseCase, authUC *usecase.AuthUseCase, moduleUC *usecase.ModuleUseCase, imageUC *usecase.ImageUseCase, navigationUC *usecase.NavigationUseCase, permissionUC *usecase.PermissionUseCase, roleUC *usecase.RoleUseCase, menuUC *usecase.MenuUseCase, submenuUC *usecase.SubmenuUseCase, posUC *usecase.PosUseCase, posPaymentUC *usecase.PosPaymentUseCase, salesReturnUC *usecase.SalesReturnUseCase, posTerminalsUC *usecase.PosTerminalsUseCase, storageLocationsUC *usecase.StorageLocationsUseCase, tenantUC *usecase.TenantUseCase, storesUC *usecase.StoreUseCase, cartUC *usecase.CartUseCase, orderUC *usecase.OrderUseCase, restaurantUC *usecase.RestaurantUseCase, customerUC *usecase.CustomerUseCase, uomUC *usecase.UOMUseCase, uomPackagingTemplateUC *usecase.UomPackagingTemplateUseCase, priceListsUC *usecase.PriceListsUseCase, taxCategoriesUC *usecase.TaxCategoriesUseCase, cashierSessionUC *usecase.CashierSessionUseCase, brandUC *usecase.BrandUseCase, cashierUC *usecase.CashierUseCase, productBarcodeUC *usecase.ProductBarcodeUseCase, productPricingUC *usecase.ProductPricingUseCase, inventoryStockUC *usecase.InventoryStockUseCase, stockMovementsUC *usecase.StockMovementsUseCase, productVariantUC *usecase.ProductVariantUseCase, promotionUC *usecase.PromotionUseCase, loyaltyUC *usecase.LoyaltyUseCase, productCatalogUC *usecase.ProductCatalogUseCase, productCategoryUC *usecase.ProductCategoryUseCase, backupUC *usecase.BackupUseCase, transferRequestsUC *usecase.TransferRequestsUseCase, goodsReceiptNotesUC *usecase.GoodsReceiptNotesUseCase, sapMigrationUC *usecase.SAPMigrationUseCase, cfg *config.Config) *gin.Engine {
	if cfg.Env == "production" || cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, x-tenant-id, ngrok-skip-browser-warning")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.Use(middleware.LoggerMiddleware())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/health", healthCheck)

	devHandler := handler.NewDevHandler()
	r.GET("/api/dev/token", devHandler.GetDevToken)

	auth := r.Group("/api/auth")
	auth.Use(middleware.TenantMiddleware(tenantManager))
	{
		authHandler := handler.NewAuthHandler(authUC)
		auth.POST("/login", authHandler.Login)
	}

	publicTenants := r.Group("/api/tenants")
	publicTenants.Use(middleware.MasterRepositoryMiddleware(masterRepo))
	{
		tenantHandler := handler.NewTenantHandler(tenantUC)
		publicTenants.GET("/:slug", tenantHandler.GetTenantBySlug)
		publicTenants.GET("/active", tenantHandler.ListActiveTenants)
	}

	api := r.Group("/api")
	api.Use(middleware.JWTAuthMiddleware())
	api.Use(middleware.TenantBindingMiddleware())
	api.Use(middleware.TenantMiddleware(tenantManager))
	{
		userHandler := handler.NewUserHandler(userUC)
		router.RegisterUserRoutes(api, userHandler)
		moduleHandler := handler.NewModuleHandler(moduleUC)
		router.RegisterModuleRoutes(api, moduleHandler)
		imageHandler := handler.NewImageHandler(imageUC)
		router.RegisterImageRoutes(api, imageHandler)
		organizationHandler := handler.NewOrganizationHandler(orgUC)
		router.RegisterOrganizationRoutes(api, organizationHandler)
		navigationHandler := handler.NewNavigationHandler(navigationUC, roleUC, userUC)
		router.RegisterNavigationRoutes(api, navigationHandler)
		permissionHandler := handler.NewPermissionHandler(permissionUC)
		router.RegisterPermissionRoutes(api, permissionHandler)
		roleHandler := handler.NewRoleHandler(roleUC)
		router.RegisterRoleRoutes(api, roleHandler)
		menuHandler := handler.NewMenuHandler(menuUC)
		router.RegisterMenuRoutes(api, menuHandler)
		submenuHandler := handler.NewSubmenuHandler(submenuUC)
		router.RegisterSubmenuRoutes(api, submenuHandler)
		posHandler := handler.NewPosHandler(posUC, posPaymentUC)
		router.RegisterPosRoutes(api, posHandler)
		salesReturnHandler := handler.NewSalesReturnHandler(salesReturnUC)
		router.RegisterSalesReturnRoutes(api, salesReturnHandler)
		posTerminalsHandler := handler.NewPosTerminalsHandler(posTerminalsUC)
		router.RegisterPosTerminalsRoutes(api, posTerminalsHandler)
		storageLocationsHandler := handler.NewStorageLocationsHandler(storageLocationsUC)
		router.RegisterStorageLocationsRoutes(api, storageLocationsHandler)
		tenantHandler := handler.NewTenantHandler(tenantUC)
		router.RegisterTenantRoutes(api, tenantHandler)
		storeHandler := handler.NewStoreHandler(storesUC)
		router.RegisterStoreRoutes(api, storeHandler)
		priceListsHandler := handler.NewPriceListsHandler(priceListsUC)
		router.RegisterPriceListRoutes(api, priceListsHandler)
		taxCategoriesHandler := handler.NewTaxCategoriesHandler(taxCategoriesUC)
		router.RegisterTaxCategoryRoutes(api, taxCategoriesHandler)
		uomHandler := handler.NewUOMHandler(uomUC)
		router.RegisterUOMRoutes(api, uomHandler)
		uomPackagingTemplateHandler := handler.NewUomPackagingTemplateHandler(uomPackagingTemplateUC)
		router.RegisterUomPackagingTemplateRoutes(api, uomPackagingTemplateHandler)
		restaurantHandler := handler.NewRestaurantHandler(restaurantUC)
		router.RegisterRestaurantRoutes(api, restaurantHandler)
		customerHandler := handler.NewCustomerHandler(customerUC)
		router.RegisterCustomerRoutes(api, customerHandler)

		// [NEW] Cart and Order Modules
		cartHandler := handler.NewCartHandler(cartUC)
		router.RegisterCartRoutes(api, cartHandler)
		orderHandler := handler.NewOrderHandler(orderUC)
		router.RegisterOrderRoutes(api, orderHandler)

		cashierSessionHandler := handler.NewCashierSessionHandler(cashierSessionUC)
		router.RegisterCashierSessionRoutes(api, cashierSessionHandler)

		productBarcodeHandler := handler.NewProductBarcodeHandler(productBarcodeUC)
		router.RegisterProductBarcodeRoutes(api, productBarcodeHandler)
		// Brand and Cashier routes
		brandHandler := handler.NewBrandHandler(brandUC)
		router.RegisterBrandRoutes(api, brandHandler)
		cashierHandler := handler.NewCashierHandler(cashierUC)
		router.RegisterCashierRoutes(api, cashierHandler)
		productPricingHandler := handler.NewProductPricingHandler(productPricingUC)
		router.RegisterProductPricingRoutes(api, productPricingHandler)
		inventoryStockHandler := handler.NewInventoryStockHandler(inventoryStockUC)
		router.RegisterInventoryStockRoutes(api, inventoryStockHandler)
		stockMovementsHandler := handler.NewStockMovementHandler(stockMovementsUC)
		router.RegisterStockMovementRoutes(api, stockMovementsHandler)
		productVariantHandler := handler.NewProductVariantHandler(productVariantUC)
		router.RegisterProductVariantRoutes(api, productVariantHandler)

		// [NEW] Promotions module
		promotionHandler := handler.NewPromotionHandler(promotionUC)
		router.RegisterPromotionRoutes(api, promotionHandler)

		// [NEW] Loyalty Rules module
		loyaltyHandler := handler.NewLoyaltyHandler(loyaltyUC)
		router.RegisterLoyaltyRoutes(api, loyaltyHandler)

		// [NEW] Product Catalog module (admin: products + embedded variants)
		productCatalogHandler := handler.NewProductCatalogHandler(productCatalogUC)
		router.RegisterProductCatalogRoutes(api, productCatalogHandler)

		productCategoryHandler := handler.NewProductCategoryHandler(productCategoryUC)
		router.RegisterProductCategoryRoutes(api, productCategoryHandler)

		// [NEW] Backup management (status, grpc-info, server-side trigger)
		backupHandler := handler.NewBackupHandler(backupUC)
		router.RegisterBackupRoutes(api, backupHandler)

		// [NEW] M2M Token management
		m2mHandler := handler.NewM2MHandler()
		router.RegisterM2MRoutes(api, m2mHandler)

		// [NEW] Warehouse Logistics (Transfer Requests & Goods Receipt Notes)
		transferRequestsHandler := handler.NewTransferRequestsHandler(transferRequestsUC)
		router.RegisterTransferRequestRoutes(api, transferRequestsHandler)

		goodsReceiptNotesHandler := handler.NewGoodsReceiptNotesHandler(goodsReceiptNotesUC)
		router.RegisterGoodsReceiptNoteRoutes(api, goodsReceiptNotesHandler)

		// ZATCA Phase 2 + Sync Routes
		zatcaCfg := &usecase.ZatcaConfig{
			Enabled:  cfg.ZatcaEnabled,
			Env:      cfg.ZatcaEnv,
			BaseURL:  cfg.ZatcaBaseURL,
			OrgVATID: cfg.ZatcaOrgVATID,
		}
		zatcaUC := usecase.NewZatcaUseCase(zatcaCfg)
		zatcaUC.SetRepository(masterRepo)
		zatcaHandler := handler.NewZatcaHandler(zatcaUC)
		router.RegisterZatcaRoutes(api, zatcaHandler)

		// [NEW] SAP B1 Migration routes
		sapMigrationHandler := handler.NewSAPMigrationHandler(sapMigrationUC)
		router.RegisterSAPMigrationRoutes(api, sapMigrationHandler)
	}

	// Also support /api/v1/migration directly for migration agents
	apiV1 := r.Group("/api/v1")
	{
		sapMigrationHandler := handler.NewSAPMigrationHandler(sapMigrationUC)
		router.RegisterSAPMigrationRoutes(apiV1, sapMigrationHandler)
	}

	return r
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "OK"})
}

func main() {
	env := "development"
	if len(os.Args) > 1 {
		arg := os.Args[1]
		switch arg {
		case "stg":
			env = "stg"
		case "dev":
			env = "dev"
		case "prod":
			env = "prod"
		}
	} else if os.Getenv("ENV") != "" {
		env = os.Getenv("ENV")
	}

	cfg := config.LoadConfig(env)
	if err := cfg.ValidateEnrichmentConfig(); err != nil {
		log.Fatalf("Invalid enrichment configuration: %v", err)
	}
	log.Printf("Starting NEMBUS in %s mode on port %s", cfg.Env, cfg.Port)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	masterPool, masterRepo, err := setupDatabase(ctx, cfg)
	if err != nil {
		log.Fatalf("Unable to connect to Master DB: %v", err)
	}
	defer masterPool.Close()

	tenantManager := manager.NewManager(masterRepo)

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
	posTerminalsUC := usecase.NewPosTerminalsUseCase()
	storageLocationsUC := usecase.NewStorageLocationsUseCase()
	tenantUC := usecase.NewTenantUseCase()
	storesUC := usecase.NewStoreUseCase()
	restaurantUC := usecase.NewRestaurantUseCase()
	customerUC := usecase.NewCustomerUseCase()
	uomUC := usecase.NewUOMUseCase()
	uomPackagingTemplateUC := usecase.NewUomPackagingTemplateUseCase()
	priceListsUC := usecase.NewPriceListsUseCase()
	taxCategoriesUC := usecase.NewTaxCategoriesUseCase()

	// [NEW]
	cartUC := usecase.NewCartUseCase()
	orderUC := usecase.NewOrderUseCase()
	salesReturnUC := usecase.NewSalesReturnUseCase()
	cashierSessionUC := usecase.NewCashierSessionUseCase()
	productBarcodeUC := usecase.NewProductBarcodeUseCase()
	brandUC := usecase.NewBrandUseCase()
	cashierUC := usecase.NewCashierUseCase()
	productPricingUC := usecase.NewProductPricingUseCase()
	inventoryStockUC := usecase.NewInventoryStockUseCase()
	stockMovementsUC := usecase.NewStockMovementsUseCase()
	productVariantUC := usecase.NewProductVariantUseCase()

	promotionUC := usecase.NewPromotionUseCase()
	loyaltyUC := usecase.NewLoyaltyUseCase()
	productCatalogUC := usecase.NewProductCatalogUseCase()
	productCategoryUC := usecase.NewProductCategoryUseCase()

	transferRequestsUC := usecase.NewTransferRequestsUseCase()
	goodsReceiptNotesUC := usecase.NewGoodsReceiptNotesUseCase()

	// BackupUseCase talks to the local gRPC backup server (same process)
	backupUC := usecase.NewBackupUseCase(tenantManager, "localhost:"+cfg.GRPCPort)

	// ZATCA Phase 2 Cloud Service & Reporting Worker
	zatcaCfg := &usecase.ZatcaConfig{
		Enabled:  cfg.ZatcaEnabled,
		Env:      cfg.ZatcaEnv,
		BaseURL:  cfg.ZatcaBaseURL,
		OrgVATID: cfg.ZatcaOrgVATID,
	}
	zatcaSvc := cloudzatca.NewService(masterRepo, zatcaCfg)
	if cfg.ZatcaEnabled {
		zatcaSvc.StartReportingWorker(ctx)
	}

	// SAP B1 Migration
	sapMigrationUC := usecase.NewSAPMigrationUseCase(masterPool)
	enrichmentStore := repository.NewProductEnrichmentStore(masterRepo)
	enrichmentCoordinator := enrichment.NewProductEnrichmentCoordinator(enrichmentStore)
	sapMigrationUC.SetProductEnrichmentCoordinator(enrichmentCoordinator)
	if cfg.EnrichmentEnabled {
		provider, err := openaiadapter.New(cfg.OpenAIAPIKey, cfg.OpenAIEnrichmentModel, cfg.OpenAIEnrichmentTimeout)
		if err != nil {
			log.Fatalf("Unable to configure OpenAI enrichment provider: %v", err)
		}
		worker := enrichment.NewEnrichmentWorker(enrichmentStore, provider, enrichment.EnrichmentExecutionConfig{
			Interval: cfg.EnrichmentWorkerInterval, Timeout: cfg.OpenAIEnrichmentTimeout,
			BatchSize: cfg.EnrichmentBatchSize, MaxAttempts: cfg.EnrichmentMaxRetries,
		}, log.Default())
		worker.Start(ctx)
		log.Printf("Product enrichment worker started (provider=%s model=%s)", cfg.EnrichmentProvider, cfg.OpenAIEnrichmentModel)
	}

	// Setup Router
	r := setupRouter(tenantManager, masterRepo, userUC, orgUC, authUC, moduleUC, imageUC, navigationUC, permissionUC, roleUC, menuUC, submenuUC, posUC, posPaymentUC, salesReturnUC, posTerminalsUC, storageLocationsUC, tenantUC, storesUC, cartUC, orderUC, restaurantUC, customerUC, uomUC, uomPackagingTemplateUC, priceListsUC, taxCategoriesUC, cashierSessionUC, brandUC, cashierUC, productBarcodeUC, productPricingUC, inventoryStockUC, stockMovementsUC, productVariantUC, promotionUC, loyaltyUC, productCatalogUC, productCategoryUC, backupUC, transferRequestsUC, goodsReceiptNotesUC, sapMigrationUC, cfg)
	// Serve the images folder under /images URL path
	r.Static("/images", "./images") // <-- this makes /images/* accessible

	// ── gRPC Backup Service ────────────────────────────────────────────────
	go func() {
		grpcAddr := ":" + cfg.GRPCPort
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Fatalf("gRPC: failed to listen on %s: %v", grpcAddr, err)
		}
		grpcServer := grpc.NewServer()
		backupSrv := grpcbackup.NewBackupServer(tenantManager, cfg.JWTSecret, cfg.PGDumpPath)
		backuppb.RegisterBackupServiceServer(grpcServer, backupSrv)

		syncSrv := grpcbackup.NewSyncServer(tenantManager, masterPool)
		syncpb.RegisterSyncServiceServer(grpcServer, syncSrv)

		// Enable reflection so grpcurl/Postman can discover services at runtime
		reflection.Register(grpcServer)
		log.Printf("✅ gRPC Backup & Sync Service listening on %s (reflection enabled)", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC: serve error: %v", err)
		}
	}()

	// Start HTTP Server
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("failed to run server:", err)
	}
}
