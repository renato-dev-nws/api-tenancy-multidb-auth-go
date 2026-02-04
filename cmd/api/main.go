package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saas-multi-database-api/internal/cache"
	"github.com/saas-multi-database-api/internal/config"
	"github.com/saas-multi-database-api/internal/database"
	"github.com/saas-multi-database-api/internal/handlers"
	"github.com/saas-multi-database-api/internal/middleware"
	"github.com/saas-multi-database-api/internal/repository"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize database manager
	dbManager := database.GetManager(cfg)

	ctx := context.Background()

	// Initialize Master DB pool
	if err := dbManager.InitMasterPool(ctx); err != nil {
		log.Fatalf("Failed to initialize master DB pool: %v", err)
	}

	// Initialize Admin DB pool (for migrations and admin operations)
	if err := dbManager.InitAdminPool(ctx); err != nil {
		log.Fatalf("Failed to initialize admin DB pool: %v", err)
	}

	// Initialize Redis client
	redisClient, err := cache.NewClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis client: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbManager.GetMasterPool())
	tenantRepo := repository.NewTenantRepository(dbManager.GetMasterPool())

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, tenantRepo, cfg)
	tenantHandler := handlers.NewTenantHandler(tenantRepo)

	// Setup router
	router := setupRouter(cfg, dbManager, redisClient, authHandler, tenantHandler, tenantRepo)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close database connections
	dbManager.Close()
	redisClient.Close()

	log.Println("Server exited")
}

func setupRouter(
	cfg *config.Config,
	dbManager *database.Manager,
	redisClient *cache.Client,
	authHandler *handlers.AuthHandler,
	tenantHandler *handlers.TenantHandler,
	tenantRepo *repository.TenantRepository,
) *gin.Engine {
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public routes (no authentication required)
	public := router.Group("/api/v1")
	{
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/login", authHandler.Login)
	}

	// Protected routes (authentication required)
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		protected.GET("/auth/me", authHandler.GetMe)
	}

	// Tenant-scoped routes (authentication + tenant resolution required)
	tenant := router.Group("/api/v1/adm/:url_code")
	tenant.Use(middleware.AuthMiddleware(cfg))
	tenant.Use(middleware.TenantMiddleware(dbManager, redisClient, tenantRepo))
	{
		// Tenant configuration endpoint for frontend
		tenant.GET("/config", tenantHandler.GetConfig)

		// Example: Products routes (requires 'products' feature)
		products := tenant.Group("/products")
		products.Use(middleware.RequireFeature("products"))
		{
			products.GET("", func(c *gin.Context) {
				// Example handler - would query tenant database
				c.JSON(http.StatusOK, gin.H{"message": "list products"})
			})

			products.POST("", middleware.RequirePermission("create_product"), func(c *gin.Context) {
				// Example handler - would create product in tenant database
				c.JSON(http.StatusOK, gin.H{"message": "create product"})
			})
		}

		// Example: Services routes (requires 'services' feature)
		services := tenant.Group("/services")
		services.Use(middleware.RequireFeature("services"))
		{
			services.GET("", func(c *gin.Context) {
				// Example handler - would query tenant database
				c.JSON(http.StatusOK, gin.H{"message": "list services"})
			})

			services.POST("", middleware.RequirePermission("create_service"), func(c *gin.Context) {
				// Example handler - would create service in tenant database
				c.JSON(http.StatusOK, gin.H{"message": "create service"})
			})
		}
	}

	return router
}
