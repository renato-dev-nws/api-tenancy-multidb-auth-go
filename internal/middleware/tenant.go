package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/saas-multi-database-api/internal/cache"
	"github.com/saas-multi-database-api/internal/database"
	adminRepo "github.com/saas-multi-database-api/internal/repository/admin"
)

// TenantMiddleware resolves tenant from URL code and injects context
func TenantMiddleware(dbManager *database.Manager, redisClient *cache.Client, tenantRepo *adminRepo.TenantRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract url_code from route parameter
		urlCode := c.Param("url_code")
		if urlCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url_code parameter required"})
			c.Abort()
			return
		}

		// Get user ID from context (set by AuthMiddleware)
		userIDStr, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			c.Abort()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		// Step 1: Try to get db_code from Redis cache
		dbCode, err := redisClient.GetDBCode(ctx, urlCode)
		if err != nil && err != redis.Nil {
			log.Printf("Redis error: %v", err)
		}

		// Step 2: If not in cache, query from database
		var tenant *adminRepo.Tenant
		if dbCode == "" {
			// Query tenant from Master DB using the repository
			tenantModel, err := tenantRepo.GetTenantByURLCode(ctx, urlCode)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
				c.Abort()
				return
			}

			tenant = &adminRepo.Tenant{
				ID:     tenantModel.ID,
				DBCode: tenantModel.DBCode,
				Status: string(tenantModel.Status),
			}

			dbCode = tenantModel.DBCode.String()

			// Cache the db_code for future requests (24 hour expiration)
			if err := redisClient.SetDBCode(ctx, urlCode, dbCode, 24*time.Hour); err != nil {
				log.Printf("Failed to cache db_code: %v", err)
			}
		} else {
			// If we got db_code from cache, we still need tenant details
			tenantModel, err := tenantRepo.GetTenantByURLCode(ctx, urlCode)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
				c.Abort()
				return
			}

			tenant = &adminRepo.Tenant{
				ID:     tenantModel.ID,
				DBCode: tenantModel.DBCode,
				Status: string(tenantModel.Status),
			}
		}

		// Step 3: Verify tenant is active
		if tenant.Status != "active" {
			c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("tenant is %s", tenant.Status)})
			c.Abort()
			return
		}

		// Step 4: Verify user has access to this tenant
		hasAccess, err := tenantRepo.CheckUserAccess(ctx, userID, tenant.ID)
		if err != nil {
			log.Printf("Error checking user access: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify access"})
			c.Abort()
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this tenant"})
			c.Abort()
			return
		}

		// Step 5: Get tenant features from plan
		features, err := tenantRepo.GetTenantFeatures(ctx, tenant.ID)
		if err != nil {
			log.Printf("Error getting tenant features: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get tenant features"})
			c.Abort()
			return
		}

		// Step 6: Get user permissions for this tenant
		permissions, err := tenantRepo.GetUserPermissions(ctx, userID, tenant.ID)
		if err != nil {
			log.Printf("Error getting user permissions: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user permissions"})
			c.Abort()
			return
		}

		// Step 7: Get or create tenant database pool
		tenantPool, err := dbManager.GetTenantPool(ctx, dbCode)
		if err != nil {
			log.Printf("Error getting tenant pool: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to tenant database"})
			c.Abort()
			return
		}

		// Step 8: Inject data into context
		c.Set("tenant_id", tenant.ID.String())
		c.Set("tenant_pool", tenantPool)
		c.Set("features", features)
		c.Set("permissions", permissions)

		log.Printf("Tenant resolved: %s (DB: %s) | User: %s | Features: %v | Permissions: %v",
			urlCode, dbCode, userID, features, permissions)

		c.Next()
	}
}

// RequireFeature middleware checks if a specific feature is enabled for the tenant
func RequireFeature(featureSlug string) gin.HandlerFunc {
	return func(c *gin.Context) {
		features, exists := c.Get("features")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "features not found in context"})
			c.Abort()
			return
		}

		featureList := features.([]string)
		hasFeature := false
		for _, f := range featureList {
			if f == featureSlug {
				hasFeature = true
				break
			}
		}

		if !hasFeature {
			c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("feature '%s' is not enabled", featureSlug)})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission middleware checks if user has a specific permission
func RequirePermission(permissionSlug string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "permissions not found in context"})
			c.Abort()
			return
		}

		permissionList := permissions.([]string)
		hasPermission := false
		for _, p := range permissionList {
			if p == permissionSlug {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("permission '%s' required", permissionSlug)})
			c.Abort()
			return
		}

		c.Next()
	}
}
