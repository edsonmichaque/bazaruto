package store

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentStore defines the interface for payment data operations.
type PaymentStore interface {
	CreatePayment(ctx context.Context, payment *models.Payment) error
	GetPayment(ctx context.Context, id uuid.UUID) (*models.Payment, error)
	GetPaymentByNumber(ctx context.Context, paymentNumber string) (*models.Payment, error)
	ListPayments(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, subscriptionID *uuid.UUID, status string, limit, offset int) ([]*models.Payment, error)
	UpdatePayment(ctx context.Context, payment *models.Payment) error
	DeletePayment(ctx context.Context, id uuid.UUID) error
	CountPayments(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, subscriptionID *uuid.UUID, status string) (int64, error)
}

// paymentStore implements PaymentStore interface.
type paymentStore struct {
	db *gorm.DB
}

// NewPaymentStore creates a new PaymentStore instance.
func NewPaymentStore(db *gorm.DB) PaymentStore {
	return &paymentStore{db: db}
}

// CreatePayment creates a new payment.
func (s *paymentStore) CreatePayment(ctx context.Context, payment *models.Payment) error {
	if err := s.db.WithContext(ctx).Create(payment).Error; err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}
	return nil
}

// GetPayment retrieves a payment by ID.
func (s *paymentStore) GetPayment(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	if err := s.db.WithContext(ctx).Preload("User").Preload("Policy").Preload("Subscription").First(&payment, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	return &payment, nil
}

// GetPaymentByNumber retrieves a payment by payment number.
func (s *paymentStore) GetPaymentByNumber(ctx context.Context, paymentNumber string) (*models.Payment, error) {
	var payment models.Payment
	if err := s.db.WithContext(ctx).Preload("User").Preload("Policy").Preload("Subscription").First(&payment, "payment_number = ?", paymentNumber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment by number: %w", err)
	}
	return &payment, nil
}

// ListPayments retrieves a list of payments with optional filtering.
func (s *paymentStore) ListPayments(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, subscriptionID *uuid.UUID, status string, limit, offset int) ([]*models.Payment, error) {
	var payments []*models.Payment
	query := s.db.WithContext(ctx).Model(&models.Payment{}).Preload("User").Preload("Policy").Preload("Subscription")

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

	if err := query.Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}
	return payments, nil
}

// UpdatePayment updates an existing payment.
func (s *paymentStore) UpdatePayment(ctx context.Context, payment *models.Payment) error {
	if err := s.db.WithContext(ctx).Save(payment).Error; err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}
	return nil
}

// DeletePayment soft deletes a payment.
func (s *paymentStore) DeletePayment(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.Payment{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}
	return nil
}

// CountPayments returns the total number of payments with optional filtering.
func (s *paymentStore) CountPayments(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, subscriptionID *uuid.UUID, status string) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&models.Payment{})

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
		return 0, fmt.Errorf("failed to count payments: %w", err)
	}
	return count, nil
}
