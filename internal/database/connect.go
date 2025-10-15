package database

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database wraps gorm.DB with additional functionality.
type Database struct {
	*gorm.DB
}

// Connect establishes a connection to the database with the given DSN and configuration.
// Supports both PostgreSQL and SQLite based on the DSN format.
func Connect(dsn string, cfg DBConfig) (*Database, error) {
	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)

	// Configure GORM
	config := &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Determine database type and open connection
	var db *gorm.DB
	var err error

	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		// PostgreSQL connection
		db, err = gorm.Open(postgres.Open(dsn), config)
	} else if strings.HasPrefix(dsn, "file:") || strings.HasSuffix(dsn, ".db") || strings.HasSuffix(dsn, ".sqlite") || strings.HasSuffix(dsn, ".sqlite3") {
		// SQLite connection
		db, err = gorm.Open(sqlite.Open(dsn), config)
	} else {
		// Default to PostgreSQL for backward compatibility
		db, err = gorm.Open(postgres.Open(dsn), config)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxConnections)
	sqlDB.SetMaxIdleConns(cfg.MinConnections)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.IdleTimeout)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{DB: db}, nil
}

// Close closes the database connection.
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// Health checks the database connection health.
func (d *Database) Health() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Ping()
}

// Stats returns database connection statistics.
func (d *Database) Stats() (map[string]interface{}, error) {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}

// DBConfig holds database configuration.
type DBConfig struct {
	MaxConnections int           `mapstructure:"max_connections"`
	MinConnections int           `mapstructure:"min_connections"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
	AcquireTimeout time.Duration `mapstructure:"acquire_timeout"`
	MaxLifetime    time.Duration `mapstructure:"max_lifetime"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
}
