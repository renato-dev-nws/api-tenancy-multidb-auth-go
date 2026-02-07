package tenant

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saas-multi-database-api/internal/config"
	adminModels "github.com/saas-multi-database-api/internal/models/admin"
	tenantModels "github.com/saas-multi-database-api/internal/models/tenant"
	adminRepo "github.com/saas-multi-database-api/internal/repository/admin"
	adminService "github.com/saas-multi-database-api/internal/services/admin"
	"github.com/saas-multi-database-api/internal/utils"
)

// TenantAuthHandler handles authentication for tenant users (Data Plane)
type TenantAuthHandler struct {
	userRepo      *adminRepo.UserRepository
	tenantRepo    *adminRepo.TenantRepository
	tenantService *adminService.TenantService
	cfg           *config.Config
}

func NewTenantAuthHandler(userRepo *adminRepo.UserRepository, tenantRepo *adminRepo.TenantRepository, tenantService *adminService.TenantService, cfg *config.Config) *TenantAuthHandler {
	return &TenantAuthHandler{
		userRepo:      userRepo,
		tenantRepo:    tenantRepo,
		tenantService: tenantService,
		cfg:           cfg,
	}
}

// Register creates a new tenant user
func (h *TenantAuthHandler) Register(c *gin.Context) {
	var req tenantModels.RegisterRequest

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

	response := tenantModels.TenantLoginResponse{
		Token: token,
	}
	response.User.ID = user.ID
	response.User.Email = user.Email
	response.User.FullName = profile.FullName
	response.Tenants = []tenantModels.UserTenant{} // New user has no tenants yet
	response.LastTenantLogged = nil                // No tenant logged yet

	c.JSON(http.StatusCreated, response)
}

// Login authenticates a tenant user and updates last_tenant_logged
func (h *TenantAuthHandler) Login(c *gin.Context) {
	var req tenantModels.LoginRequest

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
	adminTenants, err := h.tenantRepo.GetUserTenants(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user tenants"})
		return
	}
	tenants := convertUserTenants(adminTenants)
	if tenants == nil {
		tenants = []tenantModels.UserTenant{}
	}

	// Build response
	response := tenantModels.TenantLoginResponse{
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

	// Get tenant_id from context (injected by TenantMiddleware)
	tenantIDStr, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant context not found"})
		return
	}
	tenantID := mustParseUUID(tenantIDStr.(string))

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

	// Get tenant features and user permissions (injected by TenantMiddleware)
	features := c.MustGet("features").([]string)
	permissions := c.MustGet("permissions").([]string)

	// Get tenant profile for layout configuration
	tenantProfile, err := h.tenantRepo.GetTenantProfile(c.Request.Context(), tenantID)
	if err != nil {
		// If no profile found, use empty config (not a critical error)
		tenantProfile = &adminModels.TenantProfile{
			CustomSettings: make(map[string]interface{}),
		}
	}

	// Return complete configuration for frontend
	response := tenantModels.LoginToTenantResponse{
		Message:          "login successful",
		LastTenantLogged: urlCode,
		Features:         features,
		Permissions:      permissions,
		Config: tenantModels.TenantConfig{
			LogoURL:        tenantProfile.LogoURL,
			CompanyName:    tenantProfile.CompanyName,
			CustomSettings: tenantProfile.CustomSettings,
		},
	}

	c.JSON(http.StatusOK, response)
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
	adminTenants, err := h.tenantRepo.GetUserTenants(c.Request.Context(), user.ID)
	if err != nil {
		adminTenants = []adminModels.UserTenant{}
	}
	tenants := convertUserTenants(adminTenants)

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
	var req tenantModels.SubscriptionRequest

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
	err = h.userRepo.CreateUserProfile(c.Request.Context(), user.ID, req.FullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user profile"})
		return
	}

	// Get user profile for response
	userProfile, err := h.userRepo.GetUserProfile(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user profile"})
		return
	}

	// Criar tenant com o usuário como owner
	tenantReq := adminService.CreateTenantRequest{
		Name:         req.Name,
		Subdomain:    req.Subdomain, // User-chosen subdomain for public site (joao.meusaas.app)
		URLCode:      "",            // Auto-generate admin URL code (ex: FR34JJO390G)
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
	response := tenantModels.SubscriptionResponse{
		Token: token,
	}
	response.User.ID = user.ID
	response.User.Email = user.Email
	response.User.FullName = userProfile.FullName
	response.Tenant.ID = tenant.ID
	response.Tenant.URLCode = tenant.URLCode
	response.Tenant.Subdomain = tenant.Subdomain
	response.Tenant.Name = req.Name // Use request name as tenant hasn't got a name field

	c.JSON(http.StatusCreated, response)
}

// Helper function to parse UUID
func mustParseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}

// Helper function to convert admin.UserTenant to tenantModels.UserTenant
func convertUserTenants(adminTenants []adminModels.UserTenant) []tenantModels.UserTenant {
	result := make([]tenantModels.UserTenant, len(adminTenants))
	for i, at := range adminTenants {
		result[i] = tenantModels.UserTenant{
			ID:        at.ID,
			URLCode:   at.URLCode,
			Subdomain: at.Subdomain,
			Name:      at.Name,
			Role:      at.Role,
		}
	}
	return result
}
