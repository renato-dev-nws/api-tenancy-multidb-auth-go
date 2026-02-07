package admin

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saas-multi-database-api/internal/models/admin"
)

type SysUserRepository struct {
	pool *pgxpool.Pool
}

func NewSysUserRepository(pool *pgxpool.Pool) *SysUserRepository {
	return &SysUserRepository{pool: pool}
}

// CreateSysUser cria um novo administrador do sistema
func (r *SysUserRepository) CreateSysUser(ctx context.Context, email, passwordHash, fullName string) (*admin.SysUser, error) {
	query := `
		INSERT INTO sys_users (email, password_hash, full_name, status)
		VALUES ($1, $2, $3, 'active')
		RETURNING id, email, full_name, avatar_url, status, created_at, updated_at
	`

	var user admin.SysUser
	err := r.pool.QueryRow(ctx, query, email, passwordHash, fullName).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.AvatarURL,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetSysUserByEmail busca administrador por email
func (r *SysUserRepository) GetSysUserByEmail(ctx context.Context, email string) (*admin.SysUser, error) {
	query := `
		SELECT id, email, password_hash, full_name, avatar_url, status, created_at, updated_at
		FROM sys_users
		WHERE email = $1 AND status = 'active'
	`

	var user admin.SysUser
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.AvatarURL,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetSysUserByID busca administrador por ID
func (r *SysUserRepository) GetSysUserByID(ctx context.Context, id uuid.UUID) (*admin.SysUser, error) {
	query := `
		SELECT id, email, password_hash, full_name, avatar_url, status, created_at, updated_at
		FROM sys_users
		WHERE id = $1 AND status = 'active'
	`

	var user admin.SysUser
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.AvatarURL,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetSysUserRoles retorna as roles de um sys_user
func (r *SysUserRepository) GetSysUserRoles(ctx context.Context, sysUserID uuid.UUID) ([]admin.SysRole, error) {
	query := `
		SELECT sr.id, sr.name, sr.slug, sr.description, sr.created_at, sr.updated_at
		FROM sys_roles sr
		INNER JOIN sys_user_roles sur ON sr.id = sur.sys_role_id
		WHERE sur.sys_user_id = $1
		ORDER BY sr.name
	`

	rows, err := r.pool.Query(ctx, query, sysUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []admin.SysRole
	for rows.Next() {
		var role admin.SysRole
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Slug,
			&role.Description,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

// GetSysUserPermissions retorna as permissões de um sys_user (através de suas roles)
func (r *SysUserRepository) GetSysUserPermissions(ctx context.Context, sysUserID uuid.UUID) ([]admin.SysPermission, error) {
	query := `
		SELECT DISTINCT sp.id, sp.name, sp.slug, sp.description, sp.created_at, sp.updated_at
		FROM sys_permissions sp
		INNER JOIN sys_role_permissions srp ON sp.id = srp.sys_permission_id
		INNER JOIN sys_user_roles sur ON srp.sys_role_id = sur.sys_role_id
		WHERE sur.sys_user_id = $1
		ORDER BY sp.name
	`

	rows, err := r.pool.Query(ctx, query, sysUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []admin.SysPermission
	for rows.Next() {
		var perm admin.SysPermission
		err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.Slug,
			&perm.Description,
			&perm.CreatedAt,
			&perm.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	return permissions, rows.Err()
}

// AssignRoleToSysUser atribui uma role a um sys_user
func (r *SysUserRepository) AssignRoleToSysUser(ctx context.Context, sysUserID, roleID uuid.UUID) error {
	query := `
		INSERT INTO sys_user_roles (sys_user_id, sys_role_id)
		VALUES ($1, $2)
		ON CONFLICT (sys_user_id, sys_role_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, sysUserID, roleID)
	return err
}

// GetAllSysUsers lista todos os administradores do sistema
func (r *SysUserRepository) GetAllSysUsers(ctx context.Context) ([]admin.SysUser, error) {
	query := `
		SELECT id, email, full_name, avatar_url, status, created_at, updated_at
		FROM sys_users
		ORDER BY full_name
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []admin.SysUser
	for rows.Next() {
		var user admin.SysUser
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FullName,
			&user.AvatarURL,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// UpdateSysUser atualiza dados de um administrador
func (r *SysUserRepository) UpdateSysUser(ctx context.Context, id uuid.UUID, email, fullName string, avatarURL *string, status string) (*admin.SysUser, error) {
	query := `
		UPDATE sys_users
		SET email = $2, full_name = $3, avatar_url = $4, status = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, full_name, avatar_url, status, created_at, updated_at
	`

	var user admin.SysUser
	err := r.pool.QueryRow(ctx, query, id, email, fullName, avatarURL, status).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.AvatarURL,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CheckEmailExists verifica se email já existe (exceto o próprio user)
func (r *SysUserRepository) CheckEmailExists(ctx context.Context, email string, excludeUserID *uuid.UUID) (bool, error) {
	var query string
	var args []interface{}

	if excludeUserID != nil {
		query = `SELECT EXISTS(SELECT 1 FROM sys_users WHERE email = $1 AND id != $2)`
		args = []interface{}{email, *excludeUserID}
	} else {
		query = `SELECT EXISTS(SELECT 1 FROM sys_users WHERE email = $1)`
		args = []interface{}{email}
	}

	var exists bool
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	return exists, err
}

// RemoveAllRolesFromSysUser remove todas as roles de um sys_user
func (r *SysUserRepository) RemoveAllRolesFromSysUser(ctx context.Context, sysUserID uuid.UUID) error {
	query := `DELETE FROM sys_user_roles WHERE sys_user_id = $1`
	_, err := r.pool.Exec(ctx, query, sysUserID)
	return err
}

// DeleteSysUser deleta permanentemente um sys_user (soft delete via status também é possível)
func (r *SysUserRepository) DeleteSysUser(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sys_users WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
