package main

import (
	"context"
	"embed"
	"fmt"
	"log"

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

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist/browser
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

//go:embed migrations/*.sql
var migrations embed.FS

// setupDatabase initializes and returns the master database connection pool and repository
func setupDatabase(ctx context.Context, dbURL string) (*pgxpool.Pool, *repository.Queries, error) {
	if dbURL == "" {
		return nil, nil, fmt.Errorf("database URL is empty")
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, nil, err
	}
	// Initialize SQLC repository
	queries := repository.New(pool)

	return pool, queries, nil
}

// setupRouter initializes handlers, use cases, middleware, and routes, then returns the configured router
func setupRouter(tenantManager *manager.Manager, userUC *usecase.UserUseCase, orgUC *usecase.OrganizationUseCase, authUC *usecase.AuthUseCase, moduleUC *usecase.ModuleUseCase, imageUC *usecase.ImageUseCase, navigationUC *usecase.NavigationUseCase, permissionUC *usecase.PermissionUseCase, roleUC *usecase.RoleUseCase, menuUC *usecase.MenuUseCase, submenuUC *usecase.SubmenuUseCase, posUC *usecase.PosUseCase, posTerminalsUC *usecase.PosTerminalsUseCase, storageLocationsUC *usecase.StorageLocationsUseCase, tenantUC *usecase.TenantUseCase, storesUC *usecase.StoreUseCase, restaurantUC *usecase.RestaurantUseCase, cfg *config.Config) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.Env == "production" || cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, x-tenant-id")
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

	api := r.Group("/api")
	api.Use(middleware.JWTAuthMiddleware())
	api.Use(middleware.TenantMiddleware(tenantManager))
	{
		router.RegisterUserRoutes(api, handler.NewUserHandler(userUC))
		router.RegisterModuleRoutes(api, handler.NewModuleHandler(moduleUC))
		router.RegisterImageRoutes(api, handler.NewImageHandler(imageUC))
		router.RegisterOrganizationRoutes(api, handler.NewOrganizationHandler(orgUC))
		router.RegisterNavigationRoutes(api, handler.NewNavigationHandler(navigationUC, roleUC, userUC))
		router.RegisterPermissionRoutes(api, handler.NewPermissionHandler(permissionUC))
		router.RegisterRoleRoutes(api, handler.NewRoleHandler(roleUC))
		router.RegisterMenuRoutes(api, handler.NewMenuHandler(menuUC))
		router.RegisterSubmenuRoutes(api, handler.NewSubmenuHandler(submenuUC))
		router.RegisterPosRoutes(api, handler.NewPosHandler(posUC))
		router.RegisterPosTerminalsRoutes(api, handler.NewPosTerminalsHandler(posTerminalsUC))
		router.RegisterStorageLocationsRoutes(api, handler.NewStorageLocationsHandler(storageLocationsUC))
		router.RegisterTenantRoutes(api, handler.NewTenantHandler(tenantUC))
		router.RegisterStoreRoutes(api, handler.NewStoreHandler(storesUC))
		router.RegisterRestaurantRoutes(api, handler.NewRestaurantHandler(restaurantUC))
	}

	return r
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "OK"})
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "NEMBUS",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
