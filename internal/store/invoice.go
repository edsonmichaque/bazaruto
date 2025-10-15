package store

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InvoiceStore defines the interface for invoice data operations.
type InvoiceStore interface {
	CreateInvoice(ctx context.Context, invoice *models.Invoice) error
	GetInvoice(ctx context.Context, id uuid.UUID) (*models.Invoice, error)
	GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*models.Invoice, error)
	ListInvoices(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, subscriptionID *uuid.UUID, status string, limit, offset int) ([]*models.Invoice, error)
	UpdateInvoice(ctx context.Context, invoice *models.Invoice) error
	DeleteInvoice(ctx context.Context, id uuid.UUID) error
	CountInvoices(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, subscriptionID *uuid.UUID, status string) (int64, error)
}

// invoiceStore implements InvoiceStore interface.
type invoiceStore struct {
	db *gorm.DB
}

// NewInvoiceStore creates a new InvoiceStore instance.
func NewInvoiceStore(db *gorm.DB) InvoiceStore {
	return &invoiceStore{db: db}
}

// CreateInvoice creates a new invoice.
func (s *invoiceStore) CreateInvoice(ctx context.Context, invoice *models.Invoice) error {
	if err := s.db.WithContext(ctx).Create(invoice).Error; err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}
	return nil
}

// GetInvoice retrieves an invoice by ID.
func (s *invoiceStore) GetInvoice(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := s.db.WithContext(ctx).Preload("User").Preload("Policy").Preload("Subscription").Preload("Payment").First(&invoice, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return &invoice, nil
}

// GetInvoiceByNumber retrieves an invoice by invoice number.
func (s *invoiceStore) GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := s.db.WithContext(ctx).Preload("User").Preload("Policy").Preload("Subscription").Preload("Payment").First(&invoice, "invoice_number = ?", invoiceNumber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice by number: %w", err)
	}
	return &invoice, nil
}

// ListInvoices retrieves a list of invoices with optional filtering.
func (s *invoiceStore) ListInvoices(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, subscriptionID *uuid.UUID, status string, limit, offset int) ([]*models.Invoice, error) {
	var invoices []*models.Invoice
	query := s.db.WithContext(ctx).Model(&models.Invoice{}).Preload("User").Preload("Policy").Preload("Subscription")

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if policyID != nil {
		query = query.Where("policy_id = ?", *policyID)
	}
	if subscriptionID != nil {
		query = query.Where("subscription_id = ?", *subscriptionID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}
	return invoices, nil
}

// UpdateInvoice updates an existing invoice.
func (s *invoiceStore) UpdateInvoice(ctx context.Context, invoice *models.Invoice) error {
	if err := s.db.WithContext(ctx).Save(invoice).Error; err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}
	return nil
}

// DeleteInvoice soft deletes an invoice.
func (s *invoiceStore) DeleteInvoice(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.Invoice{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}
	return nil
}

// CountInvoices returns the total number of invoices with optional filtering.
func (s *invoiceStore) CountInvoices(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, subscriptionID *uuid.UUID, status string) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&models.Invoice{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if policyID != nil {
		query = query.Where("policy_id = ?", *policyID)
	}
	if subscriptionID != nil {
		query = query.Where("subscription_id = ?", *subscriptionID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count invoices: %w", err)
	}
	return count, nil
}
