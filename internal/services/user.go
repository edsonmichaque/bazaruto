package services

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/events"
	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/google/uuid"
)

// UserService handles business logic for users.
type UserService struct {
	store        store.UserStore
	eventService *EventService
}

// NewUserService creates a new UserService instance.
func NewUserService(store store.UserStore, eventService ...*EventService) *UserService {
	var evtService *EventService
	if len(eventService) > 0 {
		evtService = eventService[0]
	}
	return &UserService{
		store:        store,
		eventService: evtService,
	}
}

// GetUser retrieves a user by ID with business logic validation.
func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}

	user, err := s.store.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email address.
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	user, err := s.store.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// CreateUser creates a new user with business logic validation.
func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	// Validate required fields
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}
	if user.FullName == "" {
		return fmt.Errorf("full name is required")
	}

	// Check if user already exists
	existingUser, err := s.store.FindByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	// Set defaults
	if user.Status == "" {
		user.Status = models.StatusActive
	}

	// Create user in database
	if err := s.store.Create(ctx, user); err != nil {
		return err
	}

	// Publish user registered event
	if s.eventService != nil {
		event := events.NewUserRegisteredEvent(user.ID, user.Email, user.FullName, user.Role)
		if err := s.eventService.PublishEvent(ctx, event); err != nil {
			// Log error but don't fail the user creation
			// In a real system, you might want to use a separate event publishing mechanism
			// or implement compensation logic
		}
	}

	return nil
}

// UpdateUser updates an existing user with business logic validation.
func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	// Validate required fields
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}
	if user.FullName == "" {
		return fmt.Errorf("full name is required")
	}

	return s.store.Update(ctx, user)
}

// DeleteUser soft deletes a user.
func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	return s.store.Delete(ctx, id)
}

// ListUsers retrieves a list of users with optional filtering.
func (s *UserService) ListUsers(ctx context.Context, opts *models.UserListOptions) (*models.ListResponse[*models.User], error) {
	if opts == nil {
		opts = models.NewUserListOptions()
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid list options: %w", err)
	}

	users, err := s.store.List(ctx, opts.GetLimit(), opts.GetOffset())
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.store.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	return models.NewListResponse(users, total, opts.ListOptions), nil
}

// CountUsers returns the total number of users.
func (s *UserService) CountUsers(ctx context.Context) (int64, error) {
	return s.store.Count(ctx)
}

// RecordUserLogin records a user login and publishes an event.
func (s *UserService) RecordUserLogin(ctx context.Context, userID uuid.UUID, email, ipAddress string) error {
	// Publish user logged in event
	if s.eventService != nil {
		event := events.NewUserLoggedInEvent(userID, email, time.Now(), ipAddress)
		if err := s.eventService.PublishEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to publish user login event: %w", err)
		}
	}

	return nil
}
