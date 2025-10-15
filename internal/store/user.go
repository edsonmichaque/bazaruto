package store

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserStore defines the interface for user data operations.
type UserStore interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByVerifyToken(ctx context.Context, token string) (*models.User, error)
	FindByResetToken(ctx context.Context, token string) (*models.User, error)
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

// userStore implements UserStore interface.
type userStore struct {
	db *gorm.DB
}

// NewUserStore creates a new UserStore instance.
func NewUserStore(db *gorm.DB) UserStore {
	return &userStore{db: db}
}

// Create creates a new user.
func (s *userStore) Create(ctx context.Context, user *models.User) error {
	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUser retrieves a user by ID.
func (s *userStore) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email.
func (s *userStore) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// FindByVerifyToken retrieves a user by verification token.
func (s *userStore) FindByVerifyToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "verify_token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found for verification token")
		}
		return nil, fmt.Errorf("failed to get user by verification token: %w", err)
	}
	return &user, nil
}

// FindByResetToken retrieves a user by reset token.
func (s *userStore) FindByResetToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "reset_token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found for reset token")
		}
		return nil, fmt.Errorf("failed to get user by reset token: %w", err)
	}
	return &user, nil
}

// List retrieves a list of users with pagination.
func (s *userStore) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	var users []*models.User
	query := s.db.WithContext(ctx).Model(&models.User{})

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// UpdateUser updates an existing user.
func (s *userStore) Update(ctx context.Context, user *models.User) error {
	if err := s.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// DeleteUser soft deletes a user.
func (s *userStore) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// Count returns the total number of users.
func (s *userStore) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}
