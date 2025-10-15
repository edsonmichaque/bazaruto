package store

import (
	"context"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/google/uuid"
)

// CustomerStore defines the interface for customer data operations.
type CustomerStore interface {
	// Create creates a new customer.
	Create(ctx context.Context, customer *models.Customer) error

	// GetByID retrieves a customer by ID.
	GetByID(ctx context.Context, id uuid.UUID) (*models.Customer, error)

	// GetByUserID retrieves a customer by user ID.
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Customer, error)

	// GetByCustomerNumber retrieves a customer by customer number.
	GetByCustomerNumber(ctx context.Context, customerNumber string) (*models.Customer, error)

	// GetByEmail retrieves a customer by email.
	GetByEmail(ctx context.Context, email string) (*models.Customer, error)

	// Update updates an existing customer.
	Update(ctx context.Context, customer *models.Customer) error

	// Delete soft deletes a customer.
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves a list of customers with filtering and pagination.
	List(ctx context.Context, opts *models.CustomerListOptions) (*models.ListResponse[*models.Customer], error)

	// Count returns the total number of customers with optional filtering.
	Count(ctx context.Context, opts *models.CustomerListOptions) (int64, error)

	// GetRiskFactors retrieves active risk factors for a customer.
	GetRiskFactors(ctx context.Context, customerID uuid.UUID) ([]models.CustomerRiskFactor, error)

	// AddRiskFactor adds a risk factor to a customer.
	AddRiskFactor(ctx context.Context, customerID uuid.UUID, riskFactor *models.CustomerRiskFactor) error

	// UpdateRiskFactor updates a customer's risk factor.
	UpdateRiskFactor(ctx context.Context, customerID uuid.UUID, riskFactor *models.CustomerRiskFactor) error

	// RemoveRiskFactor removes a risk factor from a customer.
	RemoveRiskFactor(ctx context.Context, customerID uuid.UUID, riskFactorID uuid.UUID) error

	// GetAddresses retrieves addresses for a customer.
	GetAddresses(ctx context.Context, customerID uuid.UUID) ([]models.CustomerAddress, error)

	// AddAddress adds an address to a customer.
	AddAddress(ctx context.Context, customerID uuid.UUID, address *models.CustomerAddress) error

	// UpdateAddress updates a customer's address.
	UpdateAddress(ctx context.Context, customerID uuid.UUID, address *models.CustomerAddress) error

	// RemoveAddress removes an address from a customer.
	RemoveAddress(ctx context.Context, customerID uuid.UUID, addressID uuid.UUID) error

	// GetDocuments retrieves documents for a customer.
	GetDocuments(ctx context.Context, customerID uuid.UUID) ([]models.CustomerDocument, error)

	// AddDocument adds a document to a customer.
	AddDocument(ctx context.Context, customerID uuid.UUID, document *models.CustomerDocument) error

	// UpdateDocument updates a customer's document.
	UpdateDocument(ctx context.Context, customerID uuid.UUID, document *models.CustomerDocument) error

	// RemoveDocument removes a document from a customer.
	RemoveDocument(ctx context.Context, customerID uuid.UUID, documentID uuid.UUID) error

	// UpdateKYCStatus updates a customer's KYC status.
	UpdateKYCStatus(ctx context.Context, customerID uuid.UUID, status string) error

	// UpdateAMLStatus updates a customer's AML status.
	UpdateAMLStatus(ctx context.Context, customerID uuid.UUID, status string) error

	// UpdateRiskProfile updates a customer's risk profile.
	UpdateRiskProfile(ctx context.Context, customerID uuid.UUID, riskProfile string) error

	// UpdateCustomerTier updates a customer's tier.
	UpdateCustomerTier(ctx context.Context, customerID uuid.UUID, tier string) error

	// GetCustomersByRiskProfile retrieves customers by risk profile.
	GetCustomersByRiskProfile(ctx context.Context, riskProfile string, limit, offset int) ([]*models.Customer, error)

	// GetCustomersByTier retrieves customers by tier.
	GetCustomersByTier(ctx context.Context, tier string, limit, offset int) ([]*models.Customer, error)

	// GetHighRiskCustomers retrieves high-risk customers.
	GetHighRiskCustomers(ctx context.Context, limit, offset int) ([]*models.Customer, error)

	// GetCustomersRequiringKYC retrieves customers requiring KYC verification.
	GetCustomersRequiringKYC(ctx context.Context, limit, offset int) ([]*models.Customer, error)

	// GetCustomersRequiringAML retrieves customers requiring AML review.
	GetCustomersRequiringAML(ctx context.Context, limit, offset int) ([]*models.Customer, error)
}
