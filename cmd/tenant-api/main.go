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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saas-multi-database-api/internal/cache"
	"github.com/saas-multi-database-api/internal/config"
	"github.com/saas-multi-database-api/internal/database"
	tenantHandlers "github.com/saas-multi-database-api/internal/handlers/tenant"
	"github.com/saas-multi-database-api/internal/middleware"
	adminModels "github.com/saas-multi-database-api/internal/models/admin"
	adminRepo "github.com/saas-multi-database-api/internal/repository/admin"
	tenantImageRepo "github.com/saas-multi-database-api/internal/repository/tenant"
	adminService "github.com/saas-multi-database-api/internal/services/admin"
	tenantImageService "github.com/saas-multi-database-api/internal/services/tenant"
	"github.com/saas-multi-database-api/internal/storage"
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
	userRepo := adminRepo.NewUserRepository(dbManager.GetMasterPool())
	tenantRepoMaster := adminRepo.NewTenantRepository(dbManager.GetMasterPool())
	planRepo := adminRepo.NewPlanRepository(dbManager.GetMasterPool())

	// Initialize services
	tenantServiceAdmin := adminService.NewTenantService(tenantRepoMaster, userRepo, redisClient.Client, dbManager.GetMasterPool())
	planService := adminService.NewPlanService(planRepo, redisClient.Client)

	// Initialize storage driver
	storageDriver, err := storage.NewStorageDriver(&storage.Config{
		Driver:             cfg.Storage.Driver,
		UploadsPath:        cfg.Storage.UploadsPath,
		AWSAccessKeyID:     cfg.Storage.AWSAccessKeyID,
		AWSSecretAccessKey: cfg.Storage.AWSSecretAccessKey,
		AWSRegion:          cfg.Storage.AWSRegion,
		AWSBucket:          cfg.Storage.AWSBucket,
		R2AccessKeyID:      cfg.Storage.R2AccessKeyID,
		R2SecretAccessKey:  cfg.Storage.R2SecretAccessKey,
		R2AccountID:        cfg.Storage.R2AccountID,
		R2Bucket:           cfg.Storage.R2Bucket,
		R2PublicURL:        cfg.Storage.R2PublicURL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize storage driver: %v", err)
	}

	// Initialize handlers
	authHandler := tenantHandlers.NewTenantAuthHandler(userRepo, tenantRepoMaster, tenantServiceAdmin, cfg)
	productHandler := tenantHandlers.NewProductHandler()
	serviceHandler := tenantHandlers.NewServiceHandler()
	settingHandler := tenantHandlers.NewSettingHandler()

	// Setup router
	router := setupTenantRouter(cfg, dbManager, redisClient, authHandler, productHandler, serviceHandler, settingHandler, tenantRepoMaster, tenantServiceAdmin, storageDriver, planService)

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
	authHandler *tenantHandlers.TenantAuthHandler,
	productHandler *tenantHandlers.ProductHandler,
	serviceHandler *tenantHandlers.ServiceHandler,
	settingHandler *tenantHandlers.SettingHandler,
	tenantRepo *adminRepo.TenantRepository,
	tenantService *adminService.TenantService,
	storageDriver storage.StorageDriver,
	planService *adminService.PlanService,
) *gin.Engine {
	router := gin.Default()

	// CORS middleware for frontend development
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173", "http://localhost:5174", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Serve static files from uploads directory
	router.Static("/uploads", cfg.Storage.UploadsPath)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "tenant-api"})
	})

	// Public routes (tenant user authentication)
	public := router.Group("/api/v1")
	{
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/login", authHandler.Login)
		public.POST("/subscription", authHandler.Subscribe) // Nova rota de assinatura

		// Public endpoint to list plans (for registration) - with Redis cache
		public.GET("/plans", func(c *gin.Context) {
			planResponses, err := planService.GetAllPlansWithCache(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plans", "details": err.Error()})
				return
			}

			c.JSON(http.StatusOK, adminModels.PlanListResponse{
				Plans: planResponses,
				Total: len(planResponses),
			})
		})
	}

	// Protected tenant user routes (requires tenant JWT)
	protected := router.Group("/api/v1")
	protected.Use(middleware.TenantAuthMiddleware(cfg))
	{
		protected.GET("/auth/me", authHandler.GetMe)
		protected.POST("/auth/switch-tenant", authHandler.SwitchTenant) // Nova rota de troca de tenant
		protected.GET("/tenants", func(c *gin.Context) {
			userID := c.MustGet("user_id").(uuid.UUID)
			tenants, _ := tenantService.ListUserTenants(c.Request.Context(), userID)
			c.JSON(http.StatusOK, tenants)
		})
	}

	// Tenant-scoped routes (authentication + tenant resolution required)
	tenant := router.Group("/api/v1/:url_code")
	tenant.Use(middleware.TenantAuthMiddleware(cfg))
	tenant.Use(middleware.TenantMiddleware(dbManager, redisClient, tenantRepo))
	{
		// Tenant configuration endpoint for frontend
		tenant.GET("/config", func(c *gin.Context) {
			features := c.MustGet("features").([]string)
			permissions := c.MustGet("permissions").([]string)
			tenantIDStr := c.MustGet("tenant_id").(string)
			tenantID, _ := uuid.Parse(tenantIDStr)

			// Get tenant profile for layout configuration
			profile, err := tenantRepo.GetTenantProfile(c.Request.Context(), tenantID)
			if err != nil {
				// If no profile found, use empty config
				profile = &adminModels.TenantProfile{
					CustomSettings: make(map[string]interface{}),
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"features":    features,
				"permissions": permissions,
				"config": gin.H{
					"logo_url":        profile.LogoURL,
					"company_name":    profile.CompanyName,
					"custom_settings": profile.CustomSettings,
				},
			})
		})

		// Products routes (requires 'products' feature)
		products := tenant.Group("/products")
		products.Use(middleware.RequireFeature("products"))
		{
			products.GET("", productHandler.List)
			products.POST("", middleware.RequirePermission("create_product"), productHandler.Create)
			products.GET("/:id", productHandler.GetByID)
			products.PUT("/:id", middleware.RequirePermission("update_product"), productHandler.Update)
			products.DELETE("/:id", middleware.RequirePermission("delete_product"), productHandler.Delete)
		}

		// Services routes (requires 'services' feature)
		services := tenant.Group("/services")
		services.Use(middleware.RequireFeature("services"))
		{
			services.GET("", serviceHandler.List)
			services.POST("", middleware.RequirePermission("create_service"), serviceHandler.Create)
			services.GET("/:id", serviceHandler.GetByID)
			services.PUT("/:id", middleware.RequirePermission("update_service"), serviceHandler.Update)
			services.DELETE("/:id", middleware.RequirePermission("delete_service"), serviceHandler.Delete)
		}

		// Settings routes (always available for reading, manage_settings for editing)
		settings := tenant.Group("/settings")
		{
			settings.GET("", settingHandler.List)
			settings.GET("/:key", settingHandler.GetByKey)
			settings.PUT("/:key", middleware.RequirePermission("setg_m"), settingHandler.Update)
		}

		// Profile routes (avatar and logo uploads)
		profiles := tenant.Group("/profiles")
		{
			// User avatar upload
			profiles.POST("/users/:user_id/avatar", func(c *gin.Context) {
				tenantPool := c.MustGet("tenant_pool").(*pgxpool.Pool)
				imageRepo := tenantImageRepo.NewImageRepository(tenantPool)
				profileService := tenantImageService.NewProfileService(imageRepo, storageDriver)
				profileHandler := tenantHandlers.NewProfileHandler(profileService)
				profileHandler.UploadUserAvatar(c)
			})

			// Tenant logo upload
			profiles.POST("/tenants/:tenant_id/logo", func(c *gin.Context) {
				tenantPool := c.MustGet("tenant_pool").(*pgxpool.Pool)
				imageRepo := tenantImageRepo.NewImageRepository(tenantPool)
				profileService := tenantImageService.NewProfileService(imageRepo, storageDriver)
				profileHandler := tenantHandlers.NewProfileHandler(profileService)
				profileHandler.UploadTenantLogo(c)
			})
		}

		// Images routes (polymorphic - works with products, services, etc.)
		// Permissions: Uses product/service permissions (update_product, update_service, delete_product, delete_service)
		images := tenant.Group("/images")
		{
			// Handler will be initialized per request with tenant pool
			// Upload requires update permission for product OR service
			images.POST("", middleware.RequireAnyPermission("prod_u", "serv_u"), func(c *gin.Context) {
				// Get tenant pool from context
				tenantPool := c.MustGet("tenant_pool").(*pgxpool.Pool)
				imageRepo := tenantImageRepo.NewImageRepository(tenantPool)
				uploadService := tenantImageService.NewUploadService(imageRepo, storageDriver, redisClient)
				imageHandler := tenantHandlers.NewImageHandler(imageRepo, uploadService)
				imageHandler.UploadImages(c)
			})

			images.GET("", func(c *gin.Context) {
				tenantPool := c.MustGet("tenant_pool").(*pgxpool.Pool)
				imageRepo := tenantImageRepo.NewImageRepository(tenantPool)
				uploadService := tenantImageService.NewUploadService(imageRepo, storageDriver, redisClient)
				imageHandler := tenantHandlers.NewImageHandler(imageRepo, uploadService)
				imageHandler.ListImages(c)
			})

			images.GET("/:id", func(c *gin.Context) {
				tenantPool := c.MustGet("tenant_pool").(*pgxpool.Pool)
				imageRepo := tenantImageRepo.NewImageRepository(tenantPool)
				uploadService := tenantImageService.NewUploadService(imageRepo, storageDriver, redisClient)
				imageHandler := tenantHandlers.NewImageHandler(imageRepo, uploadService)
				imageHandler.GetImage(c)
			})

			// Update metadata requires update permission for product OR service
			images.PUT("/:id", middleware.RequireAnyPermission("prod_u", "serv_u"), func(c *gin.Context) {
				tenantPool := c.MustGet("tenant_pool").(*pgxpool.Pool)
				imageRepo := tenantImageRepo.NewImageRepository(tenantPool)
				uploadService := tenantImageService.NewUploadService(imageRepo, storageDriver, redisClient)
				imageHandler := tenantHandlers.NewImageHandler(imageRepo, uploadService)
				imageHandler.UpdateImage(c)
			})

			// Delete requires delete permission for product OR service
			images.DELETE("/:id", middleware.RequireAnyPermission("prod_d", "serv_d"), func(c *gin.Context) {
				tenantPool := c.MustGet("tenant_pool").(*pgxpool.Pool)
				imageRepo := tenantImageRepo.NewImageRepository(tenantPool)
				uploadService := tenantImageService.NewUploadService(imageRepo, storageDriver, redisClient)
				imageHandler := tenantHandlers.NewImageHandler(imageRepo, uploadService)
				imageHandler.DeleteImage(c)
			})
		}
	}

	return router
}
