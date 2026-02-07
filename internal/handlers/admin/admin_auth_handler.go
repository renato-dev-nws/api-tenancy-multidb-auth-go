package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saas-multi-database-api/internal/config"
	adminModels "github.com/saas-multi-database-api/internal/models/admin"
	adminRepo "github.com/saas-multi-database-api/internal/repository/admin"
	"github.com/saas-multi-database-api/internal/utils"
)

// AdminAuthHandler handles authentication for SaaS administrators (Control Plane)
type AdminAuthHandler struct {
	sysUserRepo *adminRepo.SysUserRepository
	cfg         *config.Config
}

func NewAdminAuthHandler(sysUserRepo *adminRepo.SysUserRepository, cfg *config.Config) *AdminAuthHandler {
	return &AdminAuthHandler{
		sysUserRepo: sysUserRepo,
		cfg:         cfg,
	}
}

// Register creates a new SaaS administrator
func (h *AdminAuthHandler) Register(c *gin.Context) {
	var req adminModels.AdminRegisterRequest

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

	// Create sys_user
	sysUser, err := h.sysUserRepo.CreateSysUser(c.Request.Context(), req.Email, string(hashedPassword), req.FullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create admin user"})
		return
	}

	// Assign default 'viewer' role (role_id = 4)
	// Query the viewer role ID from sys_roles table
	// For now, skip role assignment in register - admin can assign roles later
	// TODO: Implement dynamic role lookup and assignment

	// Generate Admin JWT
	token, err := utils.GenerateAdminJWT(sysUser.ID, h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Get user roles and permissions
	roles, err := h.sysUserRepo.GetSysUserRoles(c.Request.Context(), sysUser.ID)
	if err != nil {
		roles = []adminModels.SysRole{} // Fallback to empty
	}

	permissions, err := h.sysUserRepo.GetSysUserPermissions(c.Request.Context(), sysUser.ID)
	if err != nil {
		permissions = []adminModels.SysPermission{} // Fallback to empty
	}

	// Build response
	response := adminModels.AdminLoginResponse{
		Token: token,
	}
	response.SysUser.ID = sysUser.ID
	response.SysUser.Email = sysUser.Email
	response.SysUser.FullName = sysUser.FullName

	// Extract role slugs
	response.Roles = make([]string, len(roles))
	for i, role := range roles {
		response.Roles[i] = role.Slug
	}

	// Extract permission slugs
	response.Permissions = make([]string, len(permissions))
	for i, perm := range permissions {
		response.Permissions[i] = perm.Slug
	}

	c.JSON(http.StatusCreated, response)
}

// Login authenticates a SaaS administrator
func (h *AdminAuthHandler) Login(c *gin.Context) {
	var req adminModels.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize email
	req.Email = utils.NormalizeEmail(req.Email)

	// Get sys_user by email
	sysUser, err := h.sysUserRepo.GetSysUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, sysUser.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate Admin JWT
	token, err := utils.GenerateAdminJWT(sysUser.ID, h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Get user roles and permissions
	roles, err := h.sysUserRepo.GetSysUserRoles(c.Request.Context(), sysUser.ID)
	if err != nil {
		roles = []adminModels.SysRole{} // Fallback to empty
	}

	permissions, err := h.sysUserRepo.GetSysUserPermissions(c.Request.Context(), sysUser.ID)
	if err != nil {
		permissions = []adminModels.SysPermission{} // Fallback to empty
	}

	// Build response
	response := adminModels.AdminLoginResponse{
		Token: token,
	}
	response.SysUser.ID = sysUser.ID
	response.SysUser.Email = sysUser.Email
	response.SysUser.FullName = sysUser.FullName

	// Extract role slugs
	response.Roles = make([]string, len(roles))
	for i, role := range roles {
		response.Roles[i] = role.Slug
	}

	// Extract permission slugs
	response.Permissions = make([]string, len(permissions))
	for i, perm := range permissions {
		response.Permissions[i] = perm.Slug
	}

	c.JSON(http.StatusOK, response)
}

// GetMe returns the authenticated SaaS administrator's information
func (h *AdminAuthHandler) GetMe(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	sysUser, err := h.sysUserRepo.GetSysUserByID(c.Request.Context(), mustParseUUID(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin user not found"})
		return
	}

	// Get roles and permissions
	roles, err := h.sysUserRepo.GetSysUserRoles(c.Request.Context(), sysUser.ID)
	if err != nil {
		roles = []adminModels.SysRole{}
	}

	permissions, err := h.sysUserRepo.GetSysUserPermissions(c.Request.Context(), sysUser.ID)
	if err != nil {
		permissions = []adminModels.SysPermission{}
	}

	// Extract slugs
	rolesSlugs := make([]string, len(roles))
	for i, role := range roles {
		rolesSlugs[i] = role.Slug
	}

	permissionsSlugs := make([]string, len(permissions))
	for i, perm := range permissions {
		permissionsSlugs[i] = perm.Slug
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          sysUser.ID,
		"email":       sysUser.Email,
		"full_name":   sysUser.FullName,
		"avatar_url":  sysUser.AvatarURL,
		"status":      sysUser.Status,
		"roles":       rolesSlugs,
		"permissions": permissionsSlugs,
	})
}
