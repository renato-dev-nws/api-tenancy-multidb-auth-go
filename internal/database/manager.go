package database

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saas-multi-database-api/internal/config"
)

type Manager struct {
	masterPool  *pgxpool.Pool
	adminPool   *pgxpool.Pool
	tenantPools sync.Map // map[string]*pgxpool.Pool
	cfg         *config.Config
	mu          sync.RWMutex
}

var (
	instance *Manager
	once     sync.Once
)

// GetManager returns the singleton database manager instance
func GetManager(cfg *config.Config) *Manager {
	once.Do(func() {
		instance = &Manager{
			cfg: cfg,
		}
	})
	return instance
}

// InitMasterPool initializes the connection pool to the Master DB (Control Plane)
func (m *Manager) InitMasterPool(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.masterPool != nil {
		return nil
	}

	poolConfig, err := pgxpool.ParseConfig(m.cfg.MasterDB.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to parse master db config: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create master pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping master db: %w", err)
	}

	m.masterPool = pool
	log.Println("Master DB pool initialized successfully")
	return nil
}

// InitAdminPool initializes the direct Postgres connection for admin operations
func (m *Manager) InitAdminPool(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.adminPool != nil {
		return nil
	}

	poolConfig, err := pgxpool.ParseConfig(m.cfg.AdminDB.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to parse admin db config: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = 5
	poolConfig.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create admin pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping admin db: %w", err)
	}

	m.adminPool = pool
	log.Println("Admin DB pool initialized successfully")
	return nil
}

// GetMasterPool returns the master database pool
func (m *Manager) GetMasterPool() *pgxpool.Pool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.masterPool
}

// GetAdminPool returns the admin database pool
func (m *Manager) GetAdminPool() *pgxpool.Pool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.adminPool
}

// GetTenantPool retrieves or creates a connection pool for a tenant database
func (m *Manager) GetTenantPool(ctx context.Context, dbCode string) (*pgxpool.Pool, error) {
	// Check if pool already exists
	if pool, ok := m.tenantPools.Load(dbCode); ok {
		return pool.(*pgxpool.Pool), nil
	}

	// Create new pool
	// Substituir hífens por underscores no db_code para nome válido de database (PostgreSQL identifier)
	dbCodeClean := strings.ReplaceAll(dbCode, "-", "_")
	dbName := fmt.Sprintf("db_tenant_%s", dbCodeClean)
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		m.cfg.MasterDB.User,
		m.cfg.MasterDB.Password,
		m.cfg.MasterDB.Host,
		m.cfg.MasterDB.Port,
		dbName,
		m.cfg.MasterDB.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tenant db config: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = 20
	poolConfig.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant pool for %s: %w", dbCode, err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping tenant db %s: %w", dbCode, err)
	}

	// Store the pool
	m.tenantPools.Store(dbCode, pool)
	log.Printf("Tenant DB pool created for: %s", dbCode)

	return pool, nil
}

// CloseTenantPool closes and removes a tenant pool
func (m *Manager) CloseTenantPool(dbCode string) {
	if pool, ok := m.tenantPools.LoadAndDelete(dbCode); ok {
		pool.(*pgxpool.Pool).Close()
		log.Printf("Tenant DB pool closed for: %s", dbCode)
	}
}

// Close closes all database connections
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.masterPool != nil {
		m.masterPool.Close()
		log.Println("Master DB pool closed")
	}

	if m.adminPool != nil {
		m.adminPool.Close()
		log.Println("Admin DB pool closed")
	}

	// Close all tenant pools
	m.tenantPools.Range(func(key, value interface{}) bool {
		value.(*pgxpool.Pool).Close()
		log.Printf("Tenant DB pool closed for: %s", key)
		return true
	})
}
