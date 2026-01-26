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

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "NEMBUS/docs/swagger" // Swagger generated docs
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
func setupRouter(tenantManager *manager.Manager, userUC *usecase.UserUseCase, orgUC *usecase.OrganizationUseCase, authUC *usecase.AuthUseCase, cfg *config.Config) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.Env == "production" || cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	r := gin.Default()

	// Apply logger middleware globally to all routes
	r.Use(middleware.LoggerMiddleware())

	// Swagger documentation endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public Routes (e.g., Health Check)
	r.GET("/health", healthCheck)

	// Dev Routes (only available in development mode)
	if cfg.Env == "development" || cfg.Env == "dev" {
		devHandler := handler.NewDevHandler()
		r.GET("/dev/token", devHandler.GetDevToken)
	}

	// Public Auth Routes (Login - requires tenant but not JWT)
	auth := r.Group("/api/auth")
	auth.Use(middleware.TenantMiddleware(tenantManager)) // Tenant middleware for repository access
	{
		authHandler := handler.NewAuthHandler(authUC)
		auth.POST("/login", authHandler.Login)
	}

	// Tenant-Specific Routes (Wrapped in TenantMiddleware and JWT Auth)
	// These routes require the 'x-tenant-id' header and JWT authentication
	api := r.Group("/api")
	api.Use(middleware.JWTAuthMiddleware())             // JWT authentication first
	api.Use(middleware.TenantMiddleware(tenantManager)) // Then tenant middleware
	{
		// Initialize handlers (they will get repo from context)
		userHandler := handler.NewUserHandler(userUC)
		router.RegisterUserRoutes(api, userHandler)

		organizationHandler := handler.NewOrganizationHandler(orgUC)
		router.RegisterOrganizationRoutes(api, organizationHandler)
	}

	return r
}

// healthCheck handles the health check endpoint
// @Summary      Health check
// @Description  Returns the health status of the API
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string  "status"
// @Router       /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "OK"})
}

func main() {
	// Get environment from command line or default to development
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	// Load configuration based on environment
	cfg := config.LoadConfig(env)
	log.Printf("Starting NEMBUS in %s mode on port %s", cfg.Env, cfg.Port)

	ctx := context.Background()

	// Setup Master Database Connection
	masterPool, masterRepo, err := setupDatabase(ctx, cfg)
	if err != nil {
		log.Fatalf("Unable to connect to Master DB: %v", err)
	}
	defer masterPool.Close()

	// Initialize Tenant Manager
	tenantManager := manager.NewManager(masterRepo)

	// Initialize Use Cases (without repository - will be injected per request)
	userUC := usecase.NewUserUseCase()
	orgUC := usecase.NewOrganizationUseCase()
	authUC := usecase.NewAuthUseCase()

	// Setup Router
	r := setupRouter(tenantManager, userUC, orgUC, authUC, cfg)

	// Start Server
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("failed to run server:", err)
	}
}
