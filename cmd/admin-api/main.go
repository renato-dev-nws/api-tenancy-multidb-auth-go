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
	adminHandlers "github.com/saas-multi-database-api/internal/handlers/admin"
	"github.com/saas-multi-database-api/internal/middleware"
	adminRepo "github.com/saas-multi-database-api/internal/repository/admin"
	adminService "github.com/saas-multi-database-api/internal/services/admin"
)

// Admin API - Control Plane
// Responsável por: autenticação de admins, gestão de tenants, planos e features
func main() {
	log.Println("Starting Admin API (Control Plane)...")

	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize database manager
	dbManager := database.GetManager(cfg)

	ctx := context.Background()

	// Initialize Master DB pool (READ/WRITE for Control Plane)
	if err := dbManager.InitMasterPool(ctx); err != nil {
		log.Fatalf("Failed to initialize master DB pool: %v", err)
	}

	// Initialize Admin DB pool (for CREATE DATABASE operations via Worker)
	if err := dbManager.InitAdminPool(ctx); err != nil {
		log.Fatalf("Failed to initialize admin DB pool: %v", err)
	}

	// Initialize Redis client
	redisClient, err := cache.NewClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis client: %v", err)
	}

	// Initialize repositories
	sysUserRepo := adminRepo.NewSysUserRepository(dbManager.GetMasterPool())
	userRepo := adminRepo.NewUserRepository(dbManager.GetMasterPool())
	tenantRepo := adminRepo.NewTenantRepository(dbManager.GetMasterPool())
	planRepo := adminRepo.NewPlanRepository(dbManager.GetMasterPool())
	featureRepo := adminRepo.NewFeatureRepository(dbManager.GetMasterPool())

	// Initialize services
	tenantService := adminService.NewTenantService(tenantRepo, userRepo, redisClient.Client, dbManager.GetMasterPool())

	// Initialize handlers (Admin API uses SysUserRepository)
	authHandler := adminHandlers.NewAdminAuthHandler(sysUserRepo, cfg)
	tenantHandler := adminHandlers.NewTenantHandler(tenantService)
	planHandler := adminHandlers.NewPlanHandler(planRepo)
	featureHandler := adminHandlers.NewFeatureHandler(featureRepo)
	sysUserHandler := adminHandlers.NewSysUserHandler(sysUserRepo)

	// Setup router
	router := setupAdminRouter(cfg, authHandler, tenantHandler, planHandler, featureHandler, sysUserHandler)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.AdminAPI.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Admin API listening on port %s", cfg.AdminAPI.Port)
		log.Printf("Security: Admin JWT secret isolated, IP whitelist recommended")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Admin API...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Admin API forced to shutdown: %v", err)
	}

	// Close database connections
	dbManager.Close()
	redisClient.Close()

	log.Println("Admin API exited")
}

func setupAdminRouter(
	cfg *config.Config,
	authHandler *adminHandlers.AdminAuthHandler,
	tenantHandler *adminHandlers.TenantHandler,
	planHandler *adminHandlers.PlanHandler,
	featureHandler *adminHandlers.FeatureHandler,
	sysUserHandler *adminHandlers.SysUserHandler,
) *gin.Engine {
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "admin-api"})
	})

	// Public routes (admin registration/login)
	public := router.Group("/api/v1/admin")
	{
		public.POST("/register", authHandler.Register)
		public.POST("/login", authHandler.Login)
	}

	// Protected admin routes (requires admin JWT with AdminAuthMiddleware)
	protected := router.Group("/api/v1/admin")
	protected.Use(middleware.AdminAuthMiddleware(cfg))
	{
		protected.GET("/me", authHandler.GetMe)

		// Tenant Management (Control Plane)
		protected.POST("/tenants", tenantHandler.CreateTenant)
		protected.GET("/tenants", tenantHandler.ListMyTenants)
		protected.GET("/tenants/:tenant_id", tenantHandler.GetTenant)
		protected.PUT("/tenants/:tenant_id", tenantHandler.UpdateTenant)
		protected.DELETE("/tenants/:tenant_id", tenantHandler.DeleteTenant)

		// Plan Management
		protected.GET("/plans", planHandler.GetAllPlans)
		protected.GET("/plans/:id", planHandler.GetPlanByID)
		protected.POST("/plans", planHandler.CreatePlan)
		protected.PUT("/plans/:id", planHandler.UpdatePlan)
		protected.DELETE("/plans/:id", planHandler.DeletePlan)

		// Feature Management
		protected.GET("/features", featureHandler.GetAllFeatures)
		protected.GET("/features/:id", featureHandler.GetFeatureByID)
		protected.POST("/features", featureHandler.CreateFeature)
		protected.PUT("/features/:id", featureHandler.UpdateFeature)
		protected.DELETE("/features/:id", featureHandler.DeleteFeature)

		// SysUser Management
		protected.GET("/sys-users", sysUserHandler.GetAllSysUsers)
		protected.GET("/sys-users/:id", sysUserHandler.GetSysUserByID)
		protected.POST("/sys-users", sysUserHandler.CreateSysUser)
		protected.PUT("/sys-users/:id", sysUserHandler.UpdateSysUser)
		protected.DELETE("/sys-users/:id", sysUserHandler.DeleteSysUser)

		// Feature Management (future)
		// protected.POST("/features", featureHandler.CreateFeature)
		// protected.GET("/features", featureHandler.ListFeatures)

		// Analytics & Monitoring (future)
		// protected.GET("/analytics/tenants", analyticsHandler.GetTenantStats)
		// protected.GET("/analytics/revenue", analyticsHandler.GetRevenueStats)
	}

	return router
}
