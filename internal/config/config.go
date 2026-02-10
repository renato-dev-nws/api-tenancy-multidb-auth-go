package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server    ServerConfig
	AdminAPI  APIConfig
	TenantAPI APIConfig
	MasterDB  DatabaseConfig
	AdminDB   DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	App       AppConfig
	Storage   StorageConfig
}

type ServerConfig struct {
	Port    string // Deprecated: use AdminAPI.Port or TenantAPI.Port
	GinMode string
}

type APIConfig struct {
	Port      string
	JWTSecret string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

type AppConfig struct {
	Env string
}

type StorageConfig struct {
	Driver      string
	UploadsPath string
	// AWS S3
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	AWSBucket          string
	// Cloudflare R2
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2AccountID       string
	R2Bucket          string
	R2PublicURL       string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"), // For backward compatibility
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		AdminAPI: APIConfig{
			Port:      getEnv("ADMIN_API_PORT", "8080"),
			JWTSecret: getEnv("ADMIN_JWT_SECRET", "admin-super-secret-jwt-key-change-in-production"),
		},
		TenantAPI: APIConfig{
			Port:      getEnv("TENANT_API_PORT", "8081"),
			JWTSecret: getEnv("TENANT_JWT_SECRET", "tenant-super-secret-jwt-key-change-in-production"),
		},
		MasterDB: DatabaseConfig{
			Host:     getEnv("MASTER_DB_HOST", "localhost"),
			Port:     getEnv("MASTER_DB_PORT", "6432"),
			User:     getEnv("MASTER_DB_USER", "saas_api"),
			Password: getEnv("MASTER_DB_PASSWORD", "saas_api_password"),
			DBName:   getEnv("MASTER_DB_NAME", "master_db"),
			SSLMode:  getEnv("MASTER_DB_SSLMODE", "disable"),
		},
		AdminDB: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "postgres"),
			DBName:   getEnv("POSTGRES_DB", "master_db"),
			SSLMode:  getEnv("MASTER_DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
			ExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		},
		App: AppConfig{
			Env: getEnv("APP_ENV", "development"),
		},
		Storage: StorageConfig{
			Driver:             getEnv("STORAGE_DRIVER", "local"),
			UploadsPath:        getEnv("UPLOADS_PATH", "./uploads"),
			AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
			AWSBucket:          getEnv("AWS_BUCKET", ""),
			R2AccessKeyID:      getEnv("R2_ACCESS_KEY_ID", ""),
			R2SecretAccessKey:  getEnv("R2_SECRET_ACCESS_KEY", ""),
			R2AccountID:        getEnv("R2_ACCOUNT_ID", ""),
			R2Bucket:           getEnv("R2_BUCKET", ""),
			R2PublicURL:        getEnv("R2_PUBLIC_URL", ""),
		},
	}
}

func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
		c.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
