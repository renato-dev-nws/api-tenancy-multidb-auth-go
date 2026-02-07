package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/saas-multi-database-api/internal/config"
	"github.com/saas-multi-database-api/internal/utils"
)

// AuthMiddleware validates JWT token and injects user information into context
// Deprecated: Use AdminAuthMiddleware or TenantAuthMiddleware instead
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := utils.ValidateJWT(tokenString, cfg)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Inject user information into context (como UUID)
		c.Set("user_id", claims.UserID)

		c.Next()
	}
}

// AdminAuthMiddleware validates JWT token for Admin API (Control Plane)
// Uses separate JWT secret for security isolation
func AdminAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token with Admin JWT secret
		claims, err := utils.ValidateAdminJWT(tokenString, cfg)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Inject user information into context
		c.Set("user_id", claims.UserID)
		c.Set("api_type", "admin")

		c.Next()
	}
}

// TenantAuthMiddleware validates JWT token for Tenant API (Data Plane)
// Uses separate JWT secret for security isolation
func TenantAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token with Tenant JWT secret
		claims, err := utils.ValidateTenantJWT(tokenString, cfg)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Inject user information into context
		c.Set("user_id", claims.UserID)
		c.Set("api_type", "tenant")

		c.Next()
	}
}
