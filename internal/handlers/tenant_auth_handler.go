package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saas-multi-database-api/internal/config"
	"github.com/saas-multi-database-api/internal/models"
	"github.com/saas-multi-database-api/internal/repository"
	"github.com/saas-multi-database-api/internal/services"
	"github.com/saas-multi-database-api/internal/utils"
)

// TenantAuthHandler handles authentication for tenant users (Data Plane)
type TenantAuthHandler struct {
	userRepo      *repository.UserRepository
	tenantRepo    *repository.TenantRepository
	tenantService *services.TenantService
	cfg           *config.Config
}

func NewTenantAuthHandler(userRepo *repository.UserRepository, tenantRepo *repository.TenantRepository, tenantService *services.TenantService, cfg *config.Config) *TenantAuthHandler {
	return &TenantAuthHandler{
		userRepo:      userRepo,
		tenantRepo:    tenantRepo,
		tenantService: tenantService,
		cfg:           cfg,
	}
}

// Register creates a new tenant user
func (h *TenantAuthHandler) Register(c *gin.Context) {
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

	// Generate Tenant JWT
	token, err := utils.GenerateTenantJWT(user.ID, h.cfg)
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

	response := models.TenantLoginResponse{
		Token: token,
	}
	response.User.ID = user.ID
	response.User.Email = user.Email
	response.User.FullName = profile.FullName
	response.Tenants = []models.UserTenant{} // New user has no tenants yet
	response.LastTenantLogged = nil          // No tenant logged yet

	c.JSON(http.StatusCreated, response)
}

// Login authenticates a tenant user and updates last_tenant_logged
func (h *TenantAuthHandler) Login(c *gin.Context) {
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

	// Generate Tenant JWT
	token, err := utils.GenerateTenantJWT(user.ID, h.cfg)
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

	// Build response
	response := models.TenantLoginResponse{
		Token: token,
	}
	response.User.ID = user.ID
	response.User.Email = user.Email
	response.User.FullName = profile.FullName
	response.Tenants = tenants
	response.LastTenantLogged = user.LastTenantLogged

	c.JSON(http.StatusOK, response)
}

// LoginToTenant authenticates user and updates last_tenant_logged field
// This endpoint is called when user selects a tenant from the list
func (h *TenantAuthHandler) LoginToTenant(c *gin.Context) {
	urlCode := c.Param("url_code")

	// Validate that user is authenticated
	userID := c.MustGet("user_id").(string)
	parsedUserID := mustParseUUID(userID)

	// Verify user has access to this tenant
	tenants, err := h.tenantRepo.GetUserTenants(c.Request.Context(), parsedUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user tenants"})
		return
	}

	// Check if url_code exists in user's tenants
	hasAccess := false
	for _, t := range tenants {
		if t.URLCode == urlCode {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this tenant"})
		return
	}

	// Update last_tenant_logged
	if err := h.userRepo.UpdateLastTenantLogged(c.Request.Context(), parsedUserID, urlCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update last tenant logged"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "last tenant logged updated",
		"last_tenant_logged": urlCode,
	})
}

// GetMe returns the authenticated tenant user's information
func (h *TenantAuthHandler) GetMe(c *gin.Context) {
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

	// Get user tenants
	tenants, err := h.tenantRepo.GetUserTenants(c.Request.Context(), user.ID)
	if err != nil {
		tenants = []models.UserTenant{}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                 user.ID,
		"email":              user.Email,
		"full_name":          profile.FullName,
		"last_tenant_logged": user.LastTenantLogged,
		"tenants":            tenants,
	})
}

// Subscribe cria um novo assinante com usuário e tenant simultaneamente
func (h *TenantAuthHandler) Subscribe(c *gin.Context) {
	var req models.SubscriptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos", "details": err.Error()})
		return
	}

	// Normalize email
	req.Email = utils.NormalizeEmail(req.Email)

	// Se não for empresa, usar Name como CompanyName
	if !req.IsCompany && req.CompanyName == "" {
		req.CompanyName = req.Name
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Criar usuário
	user, err := h.userRepo.CreateUser(c.Request.Context(), req.Email, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user", "details": err.Error()})
		return
	}

	// Criar perfil do usuário
	err = h.userRepo.CreateUserProfile(c.Request.Context(), user.ID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user profile"})
		return
	}

	// Criar tenant com o usuário como owner
	tenantReq := services.CreateTenantRequest{
		Name:         req.Name,
		URLCode:      req.Subdomain,
		OwnerID:      &user.ID,
		PlanID:       req.PlanID,
		BillingCycle: req.BillingCycle,
		CompanyName:  req.CompanyName,
		IsCompany:    req.IsCompany,
		CustomDomain: req.CustomDomain,
	}

	tenant, err := h.tenantService.CreateTenant(c.Request.Context(), tenantReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tenant", "details": err.Error()})
		return
	}

	// Gerar JWT para o usuário
	token, err := utils.GenerateTenantJWT(user.ID, h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Retornar resposta completa
	response := models.SubscriptionResponse{
		Token: token,
		User: models.User{
			ID:               user.ID,
			Email:            user.Email,
			LastTenantLogged: user.LastTenantLogged,
			CreatedAt:        user.CreatedAt,
			UpdatedAt:        user.UpdatedAt,
		},
		Tenant: *tenant,
	}

	c.JSON(http.StatusCreated, response)
}
