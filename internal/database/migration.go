package database

import (
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"gorm.io/gorm"
)

// RunMigrations runs all database migrations using GORM AutoMigrate.
func RunMigrations(db *gorm.DB) error {
	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"").Error; err != nil {
		return fmt.Errorf("failed to enable pgcrypto extension: %w", err)
	}

	// Run migrations for all models
	if err := db.AutoMigrate(
		&models.User{},
		&models.Partner{},
		&models.Product{},
		&models.Quote{},
		&models.Policy{},
		&models.Claim{},
		&models.Subscription{},
		&models.Payment{},
		&models.Invoice{},
		&models.Beneficiary{},
		&models.Coverage{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// DropAll drops all tables (dangerous operation).
func DropAll(db *gorm.DB) error {
	// Drop tables in reverse dependency order
	tables := []interface{}{
		&models.Coverage{},
		&models.Beneficiary{},
		&models.Invoice{},
		&models.Payment{},
		&models.Subscription{},
		&models.Claim{},
		&models.Policy{},
		&models.Quote{},
		&models.Product{},
		&models.Partner{},
		&models.User{},
	}

	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			return fmt.Errorf("failed to drop table: %w", err)
		}
	}

	return nil
}

// Migrate is a method on the Database interface for running migrations.
func (d *Database) Migrate() error {
	return RunMigrations(d.DB)
}
