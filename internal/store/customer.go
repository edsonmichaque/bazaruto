package store

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

// customerStore implements CustomerStore interface.
type customerStore struct {
	db *gorm.DB
}

// NewCustomerStore creates a new CustomerStore instance.
func NewCustomerStore(db *gorm.DB) CustomerStore {
	return &customerStore{db: db}
}

// Create creates a new customer.
func (s *customerStore) Create(ctx context.Context, customer *models.Customer) error {
	if err := s.db.WithContext(ctx).Create(customer).Error; err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}
	return nil
}

// GetByID retrieves a customer by ID.
func (s *customerStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	var customer models.Customer
	if err := s.db.WithContext(ctx).First(&customer, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	return &customer, nil
}

// GetByUserID retrieves a customer by user ID.
func (s *customerStore) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Customer, error) {
	var customer models.Customer
	if err := s.db.WithContext(ctx).First(&customer, "user_id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer by user ID: %w", err)
	}
	return &customer, nil
}

// GetByCustomerNumber retrieves a customer by customer number.
func (s *customerStore) GetByCustomerNumber(ctx context.Context, customerNumber string) (*models.Customer, error) {
	var customer models.Customer
	if err := s.db.WithContext(ctx).First(&customer, "customer_number = ?", customerNumber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer by number: %w", err)
	}
	return &customer, nil
}

// GetByEmail retrieves a customer by email.
func (s *customerStore) GetByEmail(ctx context.Context, email string) (*models.Customer, error) {
	var customer models.Customer
	if err := s.db.WithContext(ctx).First(&customer, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}
	return &customer, nil
}

// Update updates an existing customer.
func (s *customerStore) Update(ctx context.Context, customer *models.Customer) error {
	if err := s.db.WithContext(ctx).Save(customer).Error; err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}
	return nil
}

// Delete soft deletes a customer.
func (s *customerStore) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.Customer{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}
	return nil
}

// List retrieves a list of customers with filtering and pagination.
func (s *customerStore) List(ctx context.Context, opts *models.CustomerListOptions) (*models.ListResponse[*models.Customer], error) {
	var customers []*models.Customer
	var total int64

	query := s.db.WithContext(ctx).Model(&models.Customer{})

	// Apply filters
	if opts != nil {
		if opts.UserID != nil {
			query = query.Where("user_id = ?", *opts.UserID)
		}
		if opts.RiskProfile != "" {
			query = query.Where("risk_profile = ?", opts.RiskProfile)
		}
		if opts.CustomerTier != "" {
			query = query.Where("tier = ?", opts.CustomerTier)
		}
		if opts.KYCStatus != "" {
			query = query.Where("kyc_status = ?", opts.KYCStatus)
		}
		if opts.AMLStatus != "" {
			query = query.Where("aml_status = ?", opts.AMLStatus)
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count customers: %w", err)
	}

	// Apply pagination
	limit := 20
	offset := 0
	if opts != nil {
		limit = opts.GetLimit()
		offset = opts.GetOffset()
	}

	// Execute query
	if err := query.Limit(limit).Offset(offset).Find(&customers).Error; err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	return &models.ListResponse[*models.Customer]{
		Items:   customers,
		Total:   total,
		Page:    (offset / limit) + 1,
		PerPage: limit,
	}, nil
}

// Count returns the total number of customers with optional filtering.
func (s *customerStore) Count(ctx context.Context, opts *models.CustomerListOptions) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&models.Customer{})

	// Apply filters
	if opts != nil {
		if opts.UserID != nil {
			query = query.Where("user_id = ?", *opts.UserID)
		}
		if opts.RiskProfile != "" {
			query = query.Where("risk_profile = ?", opts.RiskProfile)
		}
		if opts.CustomerTier != "" {
			query = query.Where("tier = ?", opts.CustomerTier)
		}
		if opts.KYCStatus != "" {
			query = query.Where("kyc_status = ?", opts.KYCStatus)
		}
		if opts.AMLStatus != "" {
			query = query.Where("aml_status = ?", opts.AMLStatus)
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count customers: %w", err)
	}

	return count, nil
}

// Placeholder implementations for remaining methods
func (s *customerStore) GetRiskFactors(ctx context.Context, customerID uuid.UUID) ([]models.CustomerRiskFactor, error) {
	return []models.CustomerRiskFactor{}, nil
}

func (s *customerStore) AddRiskFactor(ctx context.Context, customerID uuid.UUID, riskFactor *models.CustomerRiskFactor) error {
	return nil
}

func (s *customerStore) UpdateRiskFactor(ctx context.Context, customerID uuid.UUID, riskFactor *models.CustomerRiskFactor) error {
	return nil
}

func (s *customerStore) RemoveRiskFactor(ctx context.Context, customerID uuid.UUID, riskFactorID uuid.UUID) error {
	return nil
}

func (s *customerStore) GetAddresses(ctx context.Context, customerID uuid.UUID) ([]models.CustomerAddress, error) {
	return []models.CustomerAddress{}, nil
}

func (s *customerStore) AddAddress(ctx context.Context, customerID uuid.UUID, address *models.CustomerAddress) error {
	return nil
}

func (s *customerStore) UpdateAddress(ctx context.Context, customerID uuid.UUID, address *models.CustomerAddress) error {
	return nil
}

func (s *customerStore) RemoveAddress(ctx context.Context, customerID uuid.UUID, addressID uuid.UUID) error {
	return nil
}

func (s *customerStore) GetDocuments(ctx context.Context, customerID uuid.UUID) ([]models.CustomerDocument, error) {
	return []models.CustomerDocument{}, nil
}

func (s *customerStore) AddDocument(ctx context.Context, customerID uuid.UUID, document *models.CustomerDocument) error {
	return nil
}

func (s *customerStore) UpdateDocument(ctx context.Context, customerID uuid.UUID, document *models.CustomerDocument) error {
	return nil
}

func (s *customerStore) RemoveDocument(ctx context.Context, customerID uuid.UUID, documentID uuid.UUID) error {
	return nil
}

func (s *customerStore) UpdateKYCStatus(ctx context.Context, customerID uuid.UUID, status string) error {
	return s.db.WithContext(ctx).Model(&models.Customer{}).Where("id = ?", customerID).Update("kyc_status", status).Error
}

func (s *customerStore) UpdateAMLStatus(ctx context.Context, customerID uuid.UUID, status string) error {
	return s.db.WithContext(ctx).Model(&models.Customer{}).Where("id = ?", customerID).Update("aml_status", status).Error
}

func (s *customerStore) UpdateRiskProfile(ctx context.Context, customerID uuid.UUID, riskProfile string) error {
	return s.db.WithContext(ctx).Model(&models.Customer{}).Where("id = ?", customerID).Update("risk_profile", riskProfile).Error
}

func (s *customerStore) UpdateCustomerTier(ctx context.Context, customerID uuid.UUID, tier string) error {
	return s.db.WithContext(ctx).Model(&models.Customer{}).Where("id = ?", customerID).Update("tier", tier).Error
}

func (s *customerStore) GetCustomersByRiskProfile(ctx context.Context, riskProfile string, limit, offset int) ([]*models.Customer, error) {
	var customers []*models.Customer
	if err := s.db.WithContext(ctx).Where("risk_profile = ?", riskProfile).Limit(limit).Offset(offset).Find(&customers).Error; err != nil {
		return nil, fmt.Errorf("failed to get customers by risk profile: %w", err)
	}
	return customers, nil
}

func (s *customerStore) GetCustomersByTier(ctx context.Context, tier string, limit, offset int) ([]*models.Customer, error) {
	var customers []*models.Customer
	if err := s.db.WithContext(ctx).Where("tier = ?", tier).Limit(limit).Offset(offset).Find(&customers).Error; err != nil {
		return nil, fmt.Errorf("failed to get customers by tier: %w", err)
	}
	return customers, nil
}

func (s *customerStore) GetHighRiskCustomers(ctx context.Context, limit, offset int) ([]*models.Customer, error) {
	var customers []*models.Customer
	if err := s.db.WithContext(ctx).Where("risk_profile = ?", "high").Limit(limit).Offset(offset).Find(&customers).Error; err != nil {
		return nil, fmt.Errorf("failed to get high-risk customers: %w", err)
	}
	return customers, nil
}

func (s *customerStore) GetCustomersRequiringKYC(ctx context.Context, limit, offset int) ([]*models.Customer, error) {
	var customers []*models.Customer
	if err := s.db.WithContext(ctx).Where("kyc_status = ?", "pending").Limit(limit).Offset(offset).Find(&customers).Error; err != nil {
		return nil, fmt.Errorf("failed to get customers requiring KYC: %w", err)
	}
	return customers, nil
}

func (s *customerStore) GetCustomersRequiringAML(ctx context.Context, limit, offset int) ([]*models.Customer, error) {
	var customers []*models.Customer
	if err := s.db.WithContext(ctx).Where("aml_status = ?", "pending").Limit(limit).Offset(offset).Find(&customers).Error; err != nil {
		return nil, fmt.Errorf("failed to get customers requiring AML: %w", err)
	}
	return customers, nil
}
