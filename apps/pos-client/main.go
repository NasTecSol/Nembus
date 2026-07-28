package main

import (
	"context"
	"log"
	"os"

	"github.com/NasTecSol/nembus-client/internal/config"
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/middleware/manager"
	"github.com/NasTecSol/nembus-core/repository"
	router "github.com/NasTecSol/nembus-core/routing"
	"github.com/NasTecSol/nembus-core/usecase"

	_ "github.com/NasTecSol/nembus-client/docs/swagger" // Swagger generated docs

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

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

	// Ensure legacy DB triggers that double-count cashier session balances are dropped
	_, _ = pool.Exec(ctx, "DROP TRIGGER IF EXISTS trg_update_cashier_session_balance ON pos_transactions;")
	_, _ = pool.Exec(ctx, "DROP FUNCTION IF EXISTS update_cashier_session_balance();")

	// Initialize SQLC repository
	queries := repository.New(pool)

	return pool, queries, nil
}

// setupRouter initializes handlers, use cases, middleware, and routes, then returns the configured router
func setupRouter(tenantManager *manager.Manager, masterRepo *repository.Queries, userUC *usecase.UserUseCase, orgUC *usecase.OrganizationUseCase, authUC *usecase.AuthUseCase, moduleUC *usecase.ModuleUseCase, imageUC *usecase.ImageUseCase, navigationUC *usecase.NavigationUseCase, permissionUC *usecase.PermissionUseCase, roleUC *usecase.RoleUseCase, menuUC *usecase.MenuUseCase, submenuUC *usecase.SubmenuUseCase, posUC *usecase.PosUseCase, posPaymentUC *usecase.PosPaymentUseCase, salesReturnUC *usecase.SalesReturnUseCase, posTerminalsUC *usecase.PosTerminalsUseCase, storageLocationsUC *usecase.StorageLocationsUseCase, tenantUC *usecase.TenantUseCase, storesUC *usecase.StoreUseCase, cartUC *usecase.CartUseCase, orderUC *usecase.OrderUseCase, restaurantUC *usecase.RestaurantUseCase, customerUC *usecase.CustomerUseCase, uomUC *usecase.UOMUseCase, priceListsUC *usecase.PriceListsUseCase, taxCategoriesUC *usecase.TaxCategoriesUseCase, cashierSessionUC *usecase.CashierSessionUseCase, brandUC *usecase.BrandUseCase, cashierUC *usecase.CashierUseCase, productBarcodeUC *usecase.ProductBarcodeUseCase, productPricingUC *usecase.ProductPricingUseCase, inventoryStockUC *usecase.InventoryStockUseCase, productVariantUC *usecase.ProductVariantUseCase, promotionUC *usecase.PromotionUseCase, loyaltyUC *usecase.LoyaltyUseCase, productCatalogUC *usecase.ProductCatalogUseCase, printUC *usecase.PrintUseCase, cfg *config.Config) *gin.Engine {
	if cfg.Env == "production" || cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, x-tenant-id, ngrok-skip-browser-warning")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.Use(middleware.LoggerMiddleware())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/health", healthCheck)

	if cfg.Env == "development" || cfg.Env == "dev" {
		devHandler := handler.NewDevHandler()
		r.GET("/dev/token", devHandler.GetDevToken)
	}

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

		// [NEW] Template-based ESC/POS receipt printing
		printHandler := handler.NewPrintHandler(printUC)
		router.RegisterPrintRoutes(api, printHandler)
	}

	return r
}

// Application Version & GitHub release information (populated via -ldflags at build time)
var (
	Version     = "v1.0.0"
	GithubToken = ""
	GithubRepo  = "NasTecSol/Nembus"
)

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
	log.Printf("Starting NEMBUS POS Client %s in %s mode on port %s", Version, cfg.Env, cfg.Port)

	// Wails Setup
	app := NewApp(Version, GithubToken, GithubRepo)

	err := wails.Run(&options.App{
		Title:  "NEMBUS",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal("Error starting Wails:", err)
	}
}
