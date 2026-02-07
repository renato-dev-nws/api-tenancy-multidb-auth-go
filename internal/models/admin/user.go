package admin

import (
	"time"

	"github.com/google/uuid"
)

// User representa um usuário da plataforma (Data Plane)
// Armazenado no Master DB mas usado pelos tenants
type User struct {
	ID               uuid.UUID `json:"id"`
	Email            string    `json:"email"`
	PasswordHash     string    `json:"-"`
	LastTenantLogged *string   `json:"last_tenant_logged,omitempty"` // Nullable
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// UserProfile contém dados adicionais do usuário
type UserProfile struct {
	UserID    uuid.UUID `json:"user_id"`
	FullName  string    `json:"full_name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserTenant representa um tenant ao qual o usuário tem acesso
type UserTenant struct {
	ID        uuid.UUID `json:"id"`
	URLCode   string    `json:"url_code"`  // For admin panel
	Subdomain string    `json:"subdomain"` // For public site
	Name      string    `json:"name,omitempty"`
	Role      string    `json:"role,omitempty"`
}
