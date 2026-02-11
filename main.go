package main

import (
	"context"
	"log"
	"os"

	"NEMBUS/internal/config"
	"NEMBUS/internal/handler"
	"NEMBUS/internal/middleware"
	"NEMBUS/internal/middleware/manager"
	"NEMBUS/internal/repository"
	router "NEMBUS/internal/routing"
	"NEMBUS/internal/usecase"

	_ "NEMBUS/docs/swagger" // Swagger generated docs

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

// setupRouter initializes handlers, use cases, middleware, and routes, then returns the configured router
func setupRouter(tenantManager *manager.Manager, masterRepo *repository.Queries, userUC *usecase.UserUseCase, orgUC *usecase.OrganizationUseCase, authUC *usecase.AuthUseCase, moduleUC *usecase.ModuleUseCase, imageUC *usecase.ImageUseCase, navigationUC *usecase.NavigationUseCase, permissionUC *usecase.PermissionUseCase, roleUC *usecase.RoleUseCase, menuUC *usecase.MenuUseCase, submenuUC *usecase.SubmenuUseCase, posUC *usecase.PosUseCase, posTerminalsUC *usecase.PosTerminalsUseCase, storageLocationsUC *usecase.StorageLocationsUseCase, tenantUC *usecase.TenantUseCase, storesUC *usecase.StoreUseCase, cartUC *usecase.CartUseCase, orderUC *usecase.OrderUseCase, restaurantUC *usecase.RestaurantUseCase, uomUC *usecase.UOMUseCase, priceListsUC *usecase.PriceListsUseCase, taxCategoriesUC *usecase.TaxCategoriesUseCase, cfg *config.Config) *gin.Engine {
	if cfg.Env == "production" || cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
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
		posHandler := handler.NewPosHandler(posUC)
		router.RegisterPosRoutes(api, posHandler)
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

		// [NEW] Cart and Order Modules
		cartHandler := handler.NewCartHandler(cartUC)
		router.RegisterCartRoutes(api, cartHandler)
		orderHandler := handler.NewOrderHandler(orderUC)
		router.RegisterOrderRoutes(api, orderHandler)
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
	log.Printf("Starting NEMBUS in %s mode on port %s", cfg.Env, cfg.Port)

	ctx := context.Background()
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
	posTerminalsUC := usecase.NewPosTerminalsUseCase()
	storageLocationsUC := usecase.NewStorageLocationsUseCase()
	tenantUC := usecase.NewTenantUseCase()
	storesUC := usecase.NewStoreUseCase()
	restaurantUC := usecase.NewRestaurantUseCase()
	uomUC := usecase.NewUOMUseCase()
	priceListsUC := usecase.NewPriceListsUseCase()
	taxCategoriesUC := usecase.NewTaxCategoriesUseCase()

	// [NEW]
	cartUC := usecase.NewCartUseCase()
	orderUC := usecase.NewOrderUseCase()

	// Setup Router
	r := setupRouter(tenantManager, masterRepo, userUC, orgUC, authUC, moduleUC, imageUC, navigationUC, permissionUC, roleUC, menuUC, submenuUC, posUC, posTerminalsUC, storageLocationsUC, tenantUC, storesUC, cartUC, orderUC, restaurantUC, uomUC, priceListsUC, taxCategoriesUC, cfg)
	// Serve the images folder under /images URL path
	r.Static("/images", "./images") // <-- this makes /images/* accessible

	// Start Server
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("failed to run server:", err)
	}
}
