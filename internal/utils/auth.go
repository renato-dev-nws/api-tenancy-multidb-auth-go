package utils

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/saas-multi-database-api/internal/config"
	"golang.org/x/crypto/bcrypt"
)

// Claims representa os claims do JWT
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// HashPassword cria um hash bcrypt da senha
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash verifica se a senha corresponde ao hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT gera um token JWT para um usuário
// Deprecated: Use GenerateAdminJWT or GenerateTenantJWT instead
func GenerateJWT(userID uuid.UUID, cfg *config.Config) (string, error) {
	expirationTime := time.Now().Add(time.Duration(cfg.JWT.ExpirationHours) * time.Hour)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateAdminJWT gera um token JWT para Admin API (Control Plane)
func GenerateAdminJWT(userID uuid.UUID, cfg *config.Config) (string, error) {
	expirationTime := time.Now().Add(time.Duration(cfg.JWT.ExpirationHours) * time.Hour)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "admin-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.AdminAPI.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateTenantJWT gera um token JWT para Tenant API (Data Plane)
func GenerateTenantJWT(userID uuid.UUID, cfg *config.Config) (string, error) {
	expirationTime := time.Now().Add(time.Duration(cfg.JWT.ExpirationHours) * time.Hour)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "tenant-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.TenantAPI.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT valida um token JWT e retorna os claims
// Deprecated: Use ValidateAdminJWT or ValidateTenantJWT instead
func ValidateJWT(tokenString string, cfg *config.Config) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// ValidateAdminJWT valida um token JWT do Admin API
func ValidateAdminJWT(tokenString string, cfg *config.Config) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.AdminAPI.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Validar issuer
	if claims.Issuer != "admin-api" {
		return nil, fmt.Errorf("invalid token issuer")
	}

	return claims, nil
}

// ValidateTenantJWT valida um token JWT do Tenant API
func ValidateTenantJWT(tokenString string, cfg *config.Config) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.TenantAPI.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Validar issuer
	if claims.Issuer != "tenant-api" {
		return nil, fmt.Errorf("invalid token issuer")
	}

	return claims, nil
}

// GenerateSlug gera um slug único de 11 caracteres (números e letras maiúsculas)
func GenerateSlug() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 11

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// Fallback para timestamp se houver erro
		return fmt.Sprintf("%d", time.Now().UnixNano())[:length]
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[int(b[i])%len(charset)]
	}

	return string(result)
}

// NormalizeSlug normaliza um texto para formato slug (lowercase, sem espaços)
func NormalizeSlug(text string) string {
	// Converter para minúsculas
	slug := strings.ToLower(text)

	// Remover espaços e caracteres especiais, substituir por hífen
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remover caracteres não alfanuméricos (exceto hífen)
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	// Remover hífens duplicados e do início/fim
	slug = result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")

	return slug
}

// NormalizeEmail normaliza um email (lowercase e trim)
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// NormalizeDomainPrefix normaliza um prefixo de domínio
func NormalizeDomainPrefix(prefix string) string {
	return strings.ToLower(strings.TrimSpace(prefix))
}
