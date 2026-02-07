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

// Tenant API - Data Plane
// Responsável por: autenticação de usuários de tenant, operações CRUD dentro dos tenants
func main() {
	log.Println("Starting Tenant API (Data Plane)...")

	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize database manager
	dbManager := database.GetManager(cfg)

	ctx := context.Background()

	// Initialize Master DB pool (READ-ONLY for tenant metadata and validation)
	if err := dbManager.InitMasterPool(ctx); err != nil {
		log.Fatalf("Failed to initialize master DB pool: %v", err)
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
	authHandler := handlers.NewTenantAuthHandler(userRepo, tenantRepo, cfg)
	tenantHandler := handlers.NewTenantHandler(nil) // TenantService not needed for Data Plane

	// Setup router
	router := setupTenantRouter(cfg, dbManager, redisClient, authHandler, tenantHandler, tenantRepo)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.TenantAPI.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Tenant API listening on port %s", cfg.TenantAPI.Port)
		log.Printf("Security: Tenant JWT secret isolated, rate limiting enabled")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Tenant API...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Tenant API forced to shutdown: %v", err)
	}

	// Close database connections
	dbManager.Close()
	redisClient.Close()

	log.Println("Tenant API exited")
}

func setupTenantRouter(
	cfg *config.Config,
	dbManager *database.Manager,
	redisClient *cache.Client,
	authHandler *handlers.TenantAuthHandler,
	tenantHandler *handlers.TenantHandler,
	tenantRepo *repository.TenantRepository,
) *gin.Engine {
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "tenant-api"})
	})

	// Public routes (tenant user authentication)
	public := router.Group("/api/v1")
	{
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/login", authHandler.Login)
	}

	// Protected tenant user routes (requires tenant JWT)
	protected := router.Group("/api/v1")
	protected.Use(middleware.TenantAuthMiddleware(cfg))
	{
		protected.GET("/auth/me", authHandler.GetMe)
		protected.GET("/tenants", tenantHandler.ListMyTenants)
	}

	// Tenant-scoped routes (authentication + tenant resolution required)
	tenant := router.Group("/api/v1/:url_code")
	tenant.Use(middleware.TenantAuthMiddleware(cfg))
	tenant.Use(middleware.TenantMiddleware(dbManager, redisClient, tenantRepo))
	{
		// Auth endpoint to update last_tenant_logged
		tenant.POST("/auth/login-to-tenant", authHandler.LoginToTenant)

		// Tenant configuration endpoint for frontend
		tenant.GET("/config", tenantHandler.GetConfig)

		// Products routes (requires 'products' feature)
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

			products.GET("/:id", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "get product"})
			})

			products.PUT("/:id", middleware.RequirePermission("update_product"), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "update product"})
			})

			products.DELETE("/:id", middleware.RequirePermission("delete_product"), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "delete product"})
			})
		}

		// Services routes (requires 'services' feature)
		services := tenant.Group("/services")
		services.Use(middleware.RequireFeature("services"))
		{
			services.GET("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "list services"})
			})

			services.POST("", middleware.RequirePermission("create_service"), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "create service"})
			})

			services.GET("/:id", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "get service"})
			})

			services.PUT("/:id", middleware.RequirePermission("update_service"), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "update service"})
			})

			services.DELETE("/:id", middleware.RequirePermission("delete_service"), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "delete service"})
			})
		}

		// Customers routes (always available)
		customers := tenant.Group("/customers")
		{
			customers.GET("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "list customers"})
			})

			customers.POST("", middleware.RequirePermission("create_customer"), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "create customer"})
			})

			customers.GET("/:id", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "get customer"})
			})

			customers.PUT("/:id", middleware.RequirePermission("update_customer"), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "update customer"})
			})

			customers.DELETE("/:id", middleware.RequirePermission("delete_customer"), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "delete customer"})
			})
		}

		// Orders routes (always available)
		orders := tenant.Group("/orders")
		{
			orders.GET("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "list orders"})
			})

			orders.POST("", middleware.RequirePermission("create_order"), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "create order"})
			})

			orders.GET("/:id", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "get order"})
			})

			orders.PUT("/:id", middleware.RequirePermission("update_order"), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "update order"})
			})

			orders.DELETE("/:id", middleware.RequirePermission("delete_order"), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "delete order"})
			})
		}
	}

	return router
}
