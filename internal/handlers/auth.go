package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/edsonmichaque/bazaruto/internal/authentication"
	"github.com/go-chi/chi/v5"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService *authentication.Service
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *authentication.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRoutes registers authentication routes
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.RefreshToken)
		r.Post("/logout", h.Logout)
		r.Post("/verify-email", h.VerifyEmail)
		r.Post("/reset-password", h.RequestPasswordReset)
		r.Post("/reset-password/confirm", h.ResetPassword)
		r.Route("/mfa", func(r chi.Router) {
			r.Post("/enroll", h.EnrollMFA)
			r.Post("/verify", h.VerifyMFA)
		})
	})
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	FullName    string `json:"full_name" validate:"required"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONIgnoreError(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Validate request
	if err := validateRegisterRequest(&req); err != nil {
		WriteJSONIgnoreError(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Register user
	user, err := h.authService.Register(r.Context(), &authentication.RegisterRequest{
		Email:       req.Email,
		Password:    req.Password,
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		Role:        req.Role,
	})
	if err != nil {
		WriteJSONIgnoreError(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	WriteJSONIgnoreError(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully. Please check your email for verification.",
		"user":    user,
	})
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	MFACode  string `json:"mfa_code,omitempty"`
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONIgnoreError(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Validate request
	if err := validateLoginRequest(&req); err != nil {
		WriteJSONIgnoreError(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Login user
	response, err := h.authService.Login(r.Context(), &authentication.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
		MFACode:  req.MFACode,
	})
	if err != nil {
		WriteJSONIgnoreError(w, http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
		return
	}

	WriteJSONIgnoreError(w, http.StatusOK, response)
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONIgnoreError(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Refresh token
	tokens, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		WriteJSONIgnoreError(w, http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
		return
	}

	WriteJSONIgnoreError(w, http.StatusOK, tokens)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Logout user
	if err := h.authService.Logout(r.Context()); err != nil {
		WriteJSONIgnoreError(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to logout",
		})
		return
	}

	WriteJSONIgnoreError(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// VerifyEmailRequest represents an email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONIgnoreError(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Verify email
	if err := h.authService.VerifyEmail(r.Context(), req.Token); err != nil {
		WriteJSONIgnoreError(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	WriteJSONIgnoreError(w, http.StatusOK, map[string]string{
		"message": "Email verified successfully",
	})
}

// RequestPasswordResetRequest represents a password reset request
type RequestPasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// RequestPasswordReset handles password reset requests
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req RequestPasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONIgnoreError(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Request password reset
	if err := h.authService.RequestPasswordReset(r.Context(), req.Email); err != nil {
		WriteJSONIgnoreError(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to request password reset",
		})
		return
	}

	WriteJSONIgnoreError(w, http.StatusOK, map[string]string{
		"message": "Password reset email sent",
	})
}

// ResetPasswordRequest represents a password reset confirmation request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ResetPassword handles password reset confirmation
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONIgnoreError(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Reset password
	if err := h.authService.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		WriteJSONIgnoreError(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	WriteJSONIgnoreError(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
}

// EnrollMFARequest represents an MFA enrollment request
type EnrollMFARequest struct {
	UserID string `json:"user_id" validate:"required"`
}

// EnrollMFA handles MFA enrollment
func (h *AuthHandler) EnrollMFA(w http.ResponseWriter, r *http.Request) {
	var req EnrollMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONIgnoreError(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// This would implement MFA enrollment
	WriteJSONIgnoreError(w, http.StatusOK, map[string]string{
		"message": "MFA enrollment not implemented yet",
	})
}

// VerifyMFARequest represents an MFA verification request
type VerifyMFARequest struct {
	Code string `json:"code" validate:"required"`
}

// VerifyMFA handles MFA verification
func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	var req VerifyMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONIgnoreError(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// This would implement MFA verification
	WriteJSONIgnoreError(w, http.StatusOK, map[string]string{
		"message": "MFA verification not implemented yet",
	})
}

// Validation functions

func validateRegisterRequest(req *RegisterRequest) error {
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if req.FullName == "" {
		return errors.New("full name is required")
	}
	return nil
}

func validateLoginRequest(req *LoginRequest) error {
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	return nil
}
