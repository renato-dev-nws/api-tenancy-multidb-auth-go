package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saas-multi-database-api/internal/config"
	"github.com/saas-multi-database-api/internal/models"
	"github.com/saas-multi-database-api/internal/repository"
	"github.com/saas-multi-database-api/internal/utils"
)

type AuthHandler struct {
	userRepo   *repository.UserRepository
	tenantRepo *repository.TenantRepository
	cfg        *config.Config
	apiType    string // "admin" or "tenant"
}

func NewAuthHandler(userRepo *repository.UserRepository, tenantRepo *repository.TenantRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		cfg:        cfg,
		apiType:    "tenant", // default for backward compatibility
	}
}

func NewAdminAuthHandler(userRepo *repository.UserRepository, tenantRepo *repository.TenantRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		cfg:        cfg,
		apiType:    "admin",
	}
}

func NewTenantAuthHandler(userRepo *repository.UserRepository, tenantRepo *repository.TenantRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		cfg:        cfg,
		apiType:    "tenant",
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize email
	req.Email = utils.NormalizeEmail(req.Email)

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Create user
	user, err := h.userRepo.CreateUser(c.Request.Context(), req.Email, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Create user profile
	if err := h.userRepo.CreateUserProfile(c.Request.Context(), user.ID, req.FullName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user profile"})
		return
	}

	// Generate token based on API type
	var token string
	if h.apiType == "admin" {
		token, err = utils.GenerateAdminJWT(user.ID, h.cfg)
	} else {
		token, err = utils.GenerateTenantJWT(user.ID, h.cfg)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Get user profile
	profile, err := h.userRepo.GetUserProfile(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user profile"})
		return
	}

	response := models.LoginResponse{
		Token: token,
	}
	response.User.ID = user.ID
	response.User.Email = user.Email
	response.User.FullName = profile.FullName
	response.Tenants = []models.UserTenant{} // Novo usuário não tem tenant ainda

	c.JSON(http.StatusCreated, response)
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize email
	req.Email = utils.NormalizeEmail(req.Email)

	// Get user by email
	user, err := h.userRepo.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate token based on API type
	var token string
	if h.apiType == "admin" {
		token, err = utils.GenerateAdminJWT(user.ID, h.cfg)
	} else {
		token, err = utils.GenerateTenantJWT(user.ID, h.cfg)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Get user profile
	profile, err := h.userRepo.GetUserProfile(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user profile"})
		return
	}

	// Get user tenants
	tenants, err := h.tenantRepo.GetUserTenants(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user tenants"})
		return
	}
	if tenants == nil {
		tenants = []models.UserTenant{}
	}

	response := models.LoginResponse{
		Token: token,
	}
	response.User.ID = user.ID
	response.User.Email = user.Email
	response.User.FullName = profile.FullName
	response.Tenants = tenants

	c.JSON(http.StatusOK, response)
}

// GetMe returns the authenticated user's information
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	user, err := h.userRepo.GetUserByID(c.Request.Context(), mustParseUUID(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	profile, err := h.userRepo.GetUserProfile(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        user.ID,
		"email":     user.Email,
		"full_name": profile.FullName,
	})
}
