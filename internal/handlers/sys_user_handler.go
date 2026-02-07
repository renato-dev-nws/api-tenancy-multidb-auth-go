package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saas-multi-database-api/internal/models"
	"github.com/saas-multi-database-api/internal/repository"
	"github.com/saas-multi-database-api/internal/utils"
)

type SysUserHandler struct {
	sysUserRepo *repository.SysUserRepository
}

func NewSysUserHandler(sysUserRepo *repository.SysUserRepository) *SysUserHandler {
	return &SysUserHandler{
		sysUserRepo: sysUserRepo,
	}
}

// GetAllSysUsers lista todos os administradores
// GET /api/v1/admin/sys-users
func (h *SysUserHandler) GetAllSysUsers(c *gin.Context) {
	users, err := h.sysUserRepo.GetAllSysUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sys users", "details": err.Error()})
		return
	}

	// Para cada user, buscar suas roles
	var userResponses []models.SysUserResponse
	for _, user := range users {
		roles, err := h.sysUserRepo.GetSysUserRoles(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user roles"})
			return
		}

		var roleNames []string
		for _, role := range roles {
			roleNames = append(roleNames, role.Name)
		}

		userResponses = append(userResponses, models.SysUserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FullName:  user.FullName,
			AvatarURL: user.AvatarURL,
			Status:    user.Status,
			Roles:     roleNames,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, models.SysUserListResponse{
		Users: userResponses,
		Total: len(userResponses),
	})
}

// GetSysUserByID retorna um administrador específico
// GET /api/v1/admin/sys-users/:id
func (h *SysUserHandler) GetSysUserByID(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := h.sysUserRepo.GetSysUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sys user not found"})
		return
	}

	roles, err := h.sysUserRepo.GetSysUserRoles(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user roles"})
		return
	}

	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	c.JSON(http.StatusOK, models.SysUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		AvatarURL: user.AvatarURL,
		Status:    user.Status,
		Roles:     roleNames,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

// CreateSysUser cria um novo administrador
// POST /api/v1/admin/sys-users
func (h *SysUserHandler) CreateSysUser(c *gin.Context) {
	var req models.CreateSysUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Normalizar email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Verificar se email já existe
	emailExists, err := h.sysUserRepo.CheckEmailExists(c.Request.Context(), req.Email, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check email"})
		return
	}
	if emailExists {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}

	// Hash da senha
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Criar usuário
	user, err := h.sysUserRepo.CreateSysUser(c.Request.Context(), req.Email, passwordHash, req.FullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create sys user", "details": err.Error()})
		return
	}

	// TODO: Atribuir roles se fornecidas (sys_roles.id é SERIAL/int, precisa ajustar AssignRoleToSysUser)

	c.JSON(http.StatusCreated, models.SysUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		AvatarURL: user.AvatarURL,
		Status:    user.Status,
		Roles:     []string{},
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

// UpdateSysUser atualiza um administrador existente
// PUT /api/v1/admin/sys-users/:id
func (h *SysUserHandler) UpdateSysUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req models.UpdateSysUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Normalizar email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Verificar se email já existe (exceto para este usuário)
	emailExists, err := h.sysUserRepo.CheckEmailExists(c.Request.Context(), req.Email, &userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check email"})
		return
	}
	if emailExists {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}

	// Atualizar usuário
	user, err := h.sysUserRepo.UpdateSysUser(c.Request.Context(), userID, req.Email, req.FullName, req.AvatarURL, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update sys user", "details": err.Error()})
		return
	}

	// Atualizar roles se fornecidas
	if req.RoleIDs != nil {
		// Remover todas as roles existentes
		if err := h.sysUserRepo.RemoveAllRolesFromSysUser(c.Request.Context(), userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove existing roles"})
			return
		}

		// Atribuir novas roles
		// (mesma questão do CreateSysUser sobre int vs UUID)
	}

	roles, err := h.sysUserRepo.GetSysUserRoles(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user roles"})
		return
	}

	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	c.JSON(http.StatusOK, models.SysUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		AvatarURL: user.AvatarURL,
		Status:    user.Status,
		Roles:     roleNames,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

// DeleteSysUser deleta um administrador
// DELETE /api/v1/admin/sys-users/:id
func (h *SysUserHandler) DeleteSysUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Verificar se não está tentando deletar a si mesmo
	currentUserID := c.MustGet("user_id").(uuid.UUID)
	if userID == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete your own account"})
		return
	}

	if err := h.sysUserRepo.DeleteSysUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete sys user", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "sys user deleted successfully"})
}
