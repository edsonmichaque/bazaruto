package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/edsonmichaque/bazaruto/internal/config"
	"github.com/edsonmichaque/bazaruto/internal/database"
	"github.com/edsonmichaque/bazaruto/internal/models"
)

func newAdminCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Administrative operations",
		Long: `Administrative commands for Bazaruto.
These commands handle seeding data and maintenance operations.`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "seed",
			Short: "Seed default admin data",
			Long: `Seed the database with default administrative data.
This includes creating default users, partners, and other essential data.`,
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := config.Load()
				if err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}

				db, err := database.Connect(cfg.DB.DSN, database.DBConfig{
					MaxConnections: cfg.DB.MaxConnections,
					MinConnections: cfg.DB.MinConnections,
					ConnectTimeout: cfg.DB.ConnectTimeout,
					AcquireTimeout: cfg.DB.AcquireTimeout,
					MaxLifetime:    cfg.DB.MaxLifetime,
					IdleTimeout:    cfg.DB.IdleTimeout,
				})
				if err != nil {
					return fmt.Errorf("failed to connect to database: %w", err)
				}
				defer func() { _ = db.Close() }()

				cmd.Println("Seeding admin user...")

				// Create admin user
				admin := models.User{
					Email:        "admin@bazaruto.local",
					FullName:     "Administrator",
					PasswordHash: "placeholder-hash", // In production, this should be properly hashed
					Status:       models.StatusActive,
				}

				if err := db.DB.WithContext(context.Background()).Create(&admin).Error; err != nil {
					return fmt.Errorf("failed to seed admin user: %w", err)
				}

				cmd.Println("Seeding default partner...")

				// Create default partner
				partner := models.Partner{
					Name:           "Default Insurance Partner",
					Description:    "Default insurance partner for the marketplace",
					Website:        "https://example.com",
					Email:          "partner@example.com",
					PhoneNumber:    "+1-555-0123",
					LicenseNumber:  "LIC-001",
					Status:         models.StatusActive,
					CommissionRate: 0.1, // 10% commission
				}

				if err := db.DB.WithContext(context.Background()).Create(&partner).Error; err != nil {
					return fmt.Errorf("failed to seed default partner: %w", err)
				}

				cmd.Println("Seeding sample products...")

				// Create sample products
				products := []models.Product{
					{
						Name:           "Basic Health Insurance",
						Description:    "Basic health insurance coverage",
						Category:       "health",
						PartnerID:      partner.ID,
						BasePrice:      100.0,
						Currency:       models.CurrencyUSD,
						CoverageAmount: 10000.0,
						CoveragePeriod: 365, // 1 year
						Deductible:     500.0,
						Status:         models.StatusActive,
					},
					{
						Name:           "Premium Health Insurance",
						Description:    "Premium health insurance coverage with extended benefits",
						Category:       "health",
						PartnerID:      partner.ID,
						BasePrice:      200.0,
						Currency:       models.CurrencyUSD,
						CoverageAmount: 50000.0,
						CoveragePeriod: 365, // 1 year
						Deductible:     250.0,
						Status:         models.StatusActive,
					},
					{
						Name:           "Auto Insurance",
						Description:    "Comprehensive auto insurance coverage",
						Category:       "auto",
						PartnerID:      partner.ID,
						BasePrice:      150.0,
						Currency:       models.CurrencyUSD,
						CoverageAmount: 25000.0,
						CoveragePeriod: 365, // 1 year
						Deductible:     1000.0,
						Status:         models.StatusActive,
					},
				}

				for _, product := range products {
					if err := db.DB.WithContext(context.Background()).Create(&product).Error; err != nil {
						return fmt.Errorf("failed to seed product %s: %w", product.Name, err)
					}
				}

				cmd.Println("Admin data seeded successfully.")
				return nil
			},
		},
	)

	return cmd
}
