package tenant

import (
	"log"
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

// Login authenticates a tenant user and returns last_tenant_id config
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

	// Build base response
	response := tenantModels.LoginResponse{
		Token: token,
	}
	response.User.ID = user.ID
	response.User.Email = user.Email
	response.User.FullName = profile.FullName
	response.Tenants = tenants

	// If user has last_tenant_logged, fetch and return tenant config
	if user.LastTenantLogged != nil && *user.LastTenantLogged != "" {
		urlCode := *user.LastTenantLogged

		// Get tenant by url_code
		tenant, err := h.tenantRepo.GetTenantByURLCode(c.Request.Context(), urlCode)
		if err == nil && tenant.Status == "active" {
			// Verify user still has access to this tenant
			hasAccess := false
			for _, t := range adminTenants {
				if t.URLCode == urlCode {
					hasAccess = true
					break
				}
			}

			if hasAccess {
				// Get tenant profile/layout config first (needed for company name)
				tenantProfile, err := h.tenantRepo.GetTenantProfile(c.Request.Context(), tenant.ID)
				companyName := ""
				if err == nil {
					companyName = tenantProfile.CompanyName
				}

				// Set current tenant
				response.CurrentTenant = &tenantModels.CurrentTenant{
					ID:        tenant.ID,
					URLCode:   tenant.URLCode,
					Subdomain: tenant.Subdomain,
					Name:      companyName,
				}

				// Get tenant features
				features, err := h.tenantRepo.GetTenantFeatures(c.Request.Context(), tenant.ID)
				if err == nil {
					response.Features = features
				} else {
					response.Features = []string{}
				}

				// Get user permissions for this tenant
				permissions, err := h.tenantRepo.GetUserPermissions(c.Request.Context(), user.ID, tenant.ID)
				if err == nil {
					response.Permissions = permissions
				} else {
					response.Permissions = []string{}
				}

				// Set interface config
				if tenantProfile != nil {
					response.Interface = &tenantModels.TenantConfig{
						LogoURL:        tenantProfile.LogoURL,
						CompanyName:    tenantProfile.CompanyName,
						CustomSettings: tenantProfile.CustomSettings,
					}
				} else {
					// Default empty config
					response.Interface = &tenantModels.TenantConfig{
						CustomSettings: make(map[string]interface{}),
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// SwitchTenant troca o tenant ativo do usuário
func (h *TenantAuthHandler) SwitchTenant(c *gin.Context) {
	var req tenantModels.SwitchTenantRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get authenticated user
	userID := c.MustGet("user_id").(uuid.UUID)

	// Get tenant by url_code
	tenant, err := h.tenantRepo.GetTenantByURLCode(c.Request.Context(), req.URLCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}

	// Verify tenant is active
	if tenant.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{"error": "tenant is not active"})
		return
	}

	// Verify user has access to this tenant
	hasAccess, err := h.tenantRepo.CheckUserAccess(c.Request.Context(), userID, tenant.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify access"})
		return
	}

	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this tenant"})
		return
	}

	// Update last_tenant_logged
	if err := h.userRepo.UpdateLastTenantLogged(c.Request.Context(), userID, req.URLCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update last tenant logged"})
		return
	}

	// Get tenant features
	features, err := h.tenantRepo.GetTenantFeatures(c.Request.Context(), tenant.ID)
	if err != nil {
		features = []string{}
	}

	// Get user permissions for this tenant
	permissions, err := h.tenantRepo.GetUserPermissions(c.Request.Context(), userID, tenant.ID)
	if err != nil {
		permissions = []string{}
	}

	// Get tenant profile/layout config
	tenantProfile, err := h.tenantRepo.GetTenantProfile(c.Request.Context(), tenant.ID)
	var config tenantModels.TenantConfig
	var companyName string
	if err == nil {
		config = tenantModels.TenantConfig{
			LogoURL:        tenantProfile.LogoURL,
			CompanyName:    tenantProfile.CompanyName,
			CustomSettings: tenantProfile.CustomSettings,
		}
		companyName = tenantProfile.CompanyName
	} else {
		config = tenantModels.TenantConfig{
			CustomSettings: make(map[string]interface{}),
		}
	}

	// Build response
	response := tenantModels.SwitchTenantResponse{
		Message: "tenant switched successfully",
		CurrentTenant: tenantModels.CurrentTenant{
			ID:        tenant.ID,
			URLCode:   tenant.URLCode,
			Subdomain: tenant.Subdomain,
			Name:      companyName,
		},
		Interface:   config,
		Features:    features,
		Permissions: permissions,
	}

	c.JSON(http.StatusOK, response)
}

// GetMe returns the authenticated tenant user's information
func (h *TenantAuthHandler) GetMe(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	user, err := h.userRepo.GetUserByID(c.Request.Context(), userID)
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

	// Atualizar last_tenant_logged do usuário para o tenant recém-criado
	if err := h.userRepo.UpdateLastTenantLogged(c.Request.Context(), user.ID, tenant.URLCode); err != nil {
		// Não é crítico se falhar, apenas loga
		log.Printf("Warning: Failed to update last_tenant_logged for user %s: %v", user.ID, err)
	}

	// Gerar JWT para o usuário
	token, err := utils.GenerateTenantJWT(user.ID, h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Get tenant features
	features, err := h.tenantRepo.GetTenantFeatures(c.Request.Context(), tenant.ID)
	if err != nil {
		features = []string{}
	}

	// Get user permissions (owner has all permissions)
	permissions, err := h.tenantRepo.GetUserPermissions(c.Request.Context(), user.ID, tenant.ID)
	if err != nil {
		permissions = []string{}
	}

	// Get tenant profile for interface config
	tenantProfile, err := h.tenantRepo.GetTenantProfile(c.Request.Context(), tenant.ID)
	var interfaceConfig tenantModels.TenantConfig
	if err == nil {
		interfaceConfig = tenantModels.TenantConfig{
			LogoURL:        tenantProfile.LogoURL,
			CompanyName:    tenantProfile.CompanyName,
			CustomSettings: tenantProfile.CustomSettings,
		}
	} else {
		interfaceConfig = tenantModels.TenantConfig{
			CustomSettings: make(map[string]interface{}),
		}
	}

	// Retornar resposta completa com configuração do tenant
	response := tenantModels.SubscriptionResponse{
		Token: token,
		CurrentTenant: tenantModels.CurrentTenant{
			ID:        tenant.ID,
			URLCode:   tenant.URLCode,
			Subdomain: tenant.Subdomain,
			Name:      req.CompanyName,
		},
		Interface:   interfaceConfig,
		Features:    features,
		Permissions: permissions,
	}
	response.User.ID = user.ID
	response.User.Email = user.Email
	response.User.FullName = userProfile.FullName

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
