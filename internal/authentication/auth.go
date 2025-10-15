package authentication

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/authorization"
	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/google/uuid"
)

// Service handles all authentication operations
type Service struct {
	userStore       store.UserStore
	jwtService      *JWTService
	passwordService *PasswordService
	mfaService      *MFAService
	authzService    *authorization.Service
}

// NewService creates a new authentication service
func NewService(
	userStore store.UserStore,
	jwtService *JWTService,
	passwordService *PasswordService,
	mfaService *MFAService,
	authzService *authorization.Service,
) *Service {
	return &Service{
		userStore:       userStore,
		jwtService:      jwtService,
		passwordService: passwordService,
		mfaService:      mfaService,
		authzService:    authzService,
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	FullName    string `json:"full_name" validate:"required"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	MFACode  string `json:"mfa_code,omitempty"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	User        *models.User `json:"user"`
	Tokens      *TokenPair   `json:"tokens"`
	RequiresMFA bool         `json:"requires_mfa"`
}

// Register registers a new user
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*models.User, error) {
	// Validate email format
	if err := ValidateEmail(req.Email); err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.userStore.FindByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Validate and hash password
	passwordHash, err := s.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Validate role (simple validation for now)
	role := req.Role
	if role == "" {
		role = "customer" // Default to customer
	}

	// Basic role validation
	validRoles := []string{"admin", "agent", "customer"}
	validRole := false
	for _, valid := range validRoles {
		if role == valid {
			validRole = true
			break
		}
	}
	if !validRole {
		role = "customer" // Default to customer if invalid
	}

	// For now, we'll store empty permissions and let authorization handle it
	permissionsJSON, err := json.Marshal([]string{})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal permissions: %w", err)
	}

	// Generate email verification token
	verifyToken, err := s.generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification token: %w", err)
	}

	verifyExpires := time.Now().Add(24 * time.Hour) // 24 hours

	// Create user
	user := &models.User{
		Email:         req.Email,
		FullName:      req.FullName,
		PhoneNumber:   req.PhoneNumber,
		PasswordHash:  passwordHash,
		Role:          string(role),
		Permissions:   string(permissionsJSON),
		EmailVerified: false,
		Status:        "pending_verification",
		VerifyToken:   verifyToken,
		VerifyExpires: &verifyExpires,
	}

	if err := s.userStore.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Find user by email
	user, err := s.userStore.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	if err := s.passwordService.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if user.Status != "active" {
		return nil, errors.New("account is not active")
	}

	// Check if MFA is required
	if user.MFAEnabled {
		if req.MFACode == "" {
			return &LoginResponse{
				User:        user,
				RequiresMFA: true,
			}, nil
		}

		// Verify MFA code
		if !s.mfaService.VerifyTOTP(user.MFASecret, req.MFACode) {
			return nil, errors.New("invalid MFA code")
		}
	}

	// Parse permissions (for now, just use empty permissions)
	var permissions []string
	if user.Permissions != "" {
		_ = json.Unmarshal([]byte(user.Permissions), &permissions)
	}

	// Generate tokens
	tokens, err := s.jwtService.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		user.Role,
		permissions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userStore.Update(ctx, user); err != nil {
		// Log error but don't fail login
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	return &LoginResponse{
		User:        user,
		Tokens:      tokens,
		RequiresMFA: false,
	}, nil
}

// RefreshToken refreshes an access token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Find user
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID in token")
	}

	user, err := s.userStore.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if user is still active
	if user.Status != "active" {
		return nil, errors.New("user account is not active")
	}

	// Generate new tokens
	return s.jwtService.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		user.Role,
		claims.Permissions,
	)
}

// VerifyEmail verifies a user's email address
func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	user, err := s.userStore.FindByVerifyToken(ctx, token)
	if err != nil {
		return errors.New("invalid verification token")
	}

	// Check if token is expired
	if user.VerifyExpires == nil || time.Now().After(*user.VerifyExpires) {
		return errors.New("verification token has expired")
	}

	// Update user
	user.EmailVerified = true
	user.Status = "active"
	user.VerifyToken = ""
	user.VerifyExpires = nil

	if err := s.userStore.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	return nil
}

// RequestPasswordReset requests a password reset
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userStore.FindByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not
		return nil
	}

	// Generate reset token
	resetToken, err := s.generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	resetExpires := time.Now().Add(1 * time.Hour) // 1 hour

	// Update user
	user.ResetToken = resetToken
	user.ResetExpires = &resetExpires

	if err := s.userStore.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to set reset token: %w", err)
	}

	// TODO: Send email with reset link
	// This would be handled by the email service

	return nil
}

// ResetPassword resets a user's password
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	user, err := s.userStore.FindByResetToken(ctx, token)
	if err != nil {
		return errors.New("invalid reset token")
	}

	// Check if token is expired
	if user.ResetExpires == nil || time.Now().After(*user.ResetExpires) {
		return errors.New("reset token has expired")
	}

	// Hash new password
	passwordHash, err := s.passwordService.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user
	user.PasswordHash = passwordHash
	user.ResetToken = ""
	user.ResetExpires = nil

	if err := s.userStore.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	return nil
}

// ValidateToken validates a JWT token and returns the user
func (s *Service) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID in token")
	}

	user, err := s.userStore.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if user is still active
	if user.Status != "active" {
		return nil, errors.New("user account is not active")
	}

	return user, nil
}

// Logout logs out a user (for now, just a placeholder)
func (s *Service) Logout(ctx context.Context) error {
	// In a real implementation, you might:
	// 1. Add the token to a blacklist
	// 2. Invalidate refresh tokens
	// 3. Clear session data
	// For now, we'll just return nil as JWT tokens are stateless
	return nil
}

// generateToken generates a random token
func (s *Service) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
