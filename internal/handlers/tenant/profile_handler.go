package tenant

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	tenantservice "github.com/saas-multi-database-api/internal/services/tenant"
)

type ProfileHandler struct {
	profileService *tenantservice.ProfileService
}

func NewProfileHandler(profileService *tenantservice.ProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
	}
}

// UploadUserAvatar uploads and processes user avatar (200x200px max)
// POST /api/v1/adm/:url_code/profiles/users/:user_id/avatar
func (h *ProfileHandler) UploadUserAvatar(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
		return
	}

	tenantUUID := c.MustGet("tenant_uuid").(string)
	tenantDBCode := c.MustGet("tenant_db_code").(string)

	// Get file from multipart form
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar file is required"})
		return
	}

	// Upload and process avatar
	result, err := h.profileService.UploadUserAvatar(c.Request.Context(), file, tenantUUID, tenantDBCode, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Avatar uploaded successfully",
		"image":      result,
		"avatar_url": result.PublicURL,
	})
}

// UploadTenantLogo uploads and processes tenant logo
// POST /api/v1/adm/:url_code/profiles/tenants/:tenant_id/logo
func (h *ProfileHandler) UploadTenantLogo(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	tenantUUID := c.MustGet("tenant_uuid").(string)
	tenantDBCode := c.MustGet("tenant_db_code").(string)

	// Get file from multipart form
	file, err := c.FormFile("logo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "logo file is required"})
		return
	}

	// Upload and process logo
	result, err := h.profileService.UploadTenantLogo(c.Request.Context(), file, tenantUUID, tenantDBCode, tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Logo uploaded successfully",
		"image":    result,
		"logo_url": result.PublicURL,
	})
}

// UploadSysUserAvatar uploads and processes system user avatar (admin users)
// POST /api/v1/admin/profiles/sys-users/:sys_user_id/avatar
func (h *ProfileHandler) UploadSysUserAvatar(c *gin.Context) {
	sysUserIDStr := c.Param("sys_user_id")
	sysUserID, err := uuid.Parse(sysUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sys_user_id format"})
		return
	}

	// Get file from multipart form
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar file is required"})
		return
	}

	// Upload and process system user avatar
	result, err := h.profileService.UploadSysUserAvatar(c.Request.Context(), file, sysUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Avatar uploaded successfully",
		"image":      result,
		"avatar_url": result.PublicURL,
	})
}
