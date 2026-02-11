package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	tenantservice "github.com/saas-multi-database-api/internal/services/tenant"
)

type AdminProfileHandler struct {
	profileService *tenantservice.ProfileService
}

func NewAdminProfileHandler(profileService *tenantservice.ProfileService) *AdminProfileHandler {
	return &AdminProfileHandler{
		profileService: profileService,
	}
}

// UploadSysUserAvatar uploads and processes system user avatar (admin API)
// POST /api/v1/admin/profiles/sys-users/:sys_user_id/avatar
func (h *AdminProfileHandler) UploadSysUserAvatar(c *gin.Context) {
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
