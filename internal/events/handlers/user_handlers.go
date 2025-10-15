package handlers

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/edsonmichaque/bazaruto/internal/events"
	"github.com/edsonmichaque/bazaruto/internal/jobs"
	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/edsonmichaque/bazaruto/pkg/event"
	"github.com/edsonmichaque/bazaruto/pkg/job"
)

// UserRegisteredHandler handles user registration events.
type UserRegisteredHandler struct {
	userService *services.UserService
	dispatcher  job.Dispatcher
	logger      *logger.Logger
}

// NewUserRegisteredHandler creates a new user registered event handler.
func NewUserRegisteredHandler(userService *services.UserService, dispatcher job.Dispatcher, logger *logger.Logger) *UserRegisteredHandler {
	return &UserRegisteredHandler{
		userService: userService,
		dispatcher:  dispatcher,
		logger:      logger,
	}
}

// Handle processes a user registered event.
func (h *UserRegisteredHandler) Handle(ctx context.Context, event event.Event) error {
	userEvent, ok := event.(*events.UserRegisteredEvent)
	if !ok {
		return fmt.Errorf("expected UserRegisteredEvent, got %T", event)
	}

	h.logger.Info("Processing user registered event",
		zap.String("user_id", userEvent.UserID.String()),
		zap.String("email", userEvent.Email))

	// Dispatch welcome email job
	welcomeJob := &jobs.WelcomeEmailJob{
		UserID:      userEvent.UserID,
		UserService: h.userService,
	}

	if err := h.dispatcher.PerformWithContext(ctx, welcomeJob); err != nil {
		h.logger.Error("Failed to dispatch welcome email job",
			zap.Error(err),
			zap.String("user_id", userEvent.UserID.String()))
		return fmt.Errorf("failed to dispatch welcome email job: %w", err)
	}

	h.logger.Info("Welcome email job dispatched successfully",
		zap.String("user_id", userEvent.UserID.String()))

	return nil
}

// CanHandle returns true if this handler can process the given event type.
func (h *UserRegisteredHandler) CanHandle(eventType string) bool {
	return eventType == "user.registered"
}

// HandlerName returns a unique name for this handler.
func (h *UserRegisteredHandler) HandlerName() string {
	return "user_registered_handler"
}

// UserLoggedInHandler handles user login events.
type UserLoggedInHandler struct {
	userService *services.UserService
	logger      *logger.Logger
}

// NewUserLoggedInHandler creates a new user logged in event handler.
func NewUserLoggedInHandler(userService *services.UserService, logger *logger.Logger) *UserLoggedInHandler {
	return &UserLoggedInHandler{
		userService: userService,
		logger:      logger,
	}
}

// Handle processes a user logged in event.
func (h *UserLoggedInHandler) Handle(ctx context.Context, event event.Event) error {
	userEvent, ok := event.(*events.UserLoggedInEvent)
	if !ok {
		return fmt.Errorf("expected UserLoggedInEvent, got %T", event)
	}

	h.logger.Info("Processing user logged in event",
		zap.String("user_id", userEvent.UserID.String()),
		zap.String("email", userEvent.Email),
		zap.String("ip_address", userEvent.IPAddress))

	// Update user's last login time
	user, err := h.userService.GetUser(ctx, userEvent.UserID)
	if err != nil {
		h.logger.Error("Failed to fetch user for login update",
			zap.Error(err),
			zap.String("user_id", userEvent.UserID.String()))
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	// Update last login time
	user.LastLoginAt = &userEvent.LoginTime

	if err := h.userService.UpdateUser(ctx, user); err != nil {
		h.logger.Error("Failed to update user last login time",
			zap.Error(err),
			zap.String("user_id", userEvent.UserID.String()))
		return fmt.Errorf("failed to update user: %w", err)
	}

	h.logger.Info("User last login time updated successfully",
		zap.String("user_id", userEvent.UserID.String()))

	return nil
}

// CanHandle returns true if this handler can process the given event type.
func (h *UserLoggedInHandler) CanHandle(eventType string) bool {
	return eventType == "user.logged_in"
}

// HandlerName returns a unique name for this handler.
func (h *UserLoggedInHandler) HandlerName() string {
	return "user_logged_in_handler"
}
