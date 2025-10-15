package authentication

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/edsonmichaque/bazaruto/internal/models"
)

// Context keys for storing user information
type contextKey string

const (
	userContextKey   contextKey = "user"
	userIDContextKey contextKey = "user_id"
)

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	authService *Service
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService *Service) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Authenticate middleware ensures the user is authenticated
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Authorization header is required",
			})
			return
		}

		token, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid authorization header format",
			})
			return
		}

		// Validate token and get user
		user, err := m.authService.ValidateToken(ctx, token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid or expired token",
			})
			return
		}

		// Add user to context
		ctx = context.WithValue(ctx, userContextKey, user)
		ctx = context.WithValue(ctx, userIDContextKey, user.ID.String())

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware ensures the user has one of the required roles
func (m *AuthMiddleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			user, ok := ctx.Value("user").(*models.User)
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "Authentication required",
				})
				return
			}

			// Check if user has required role
			hasRole := false
			for _, role := range roles {
				if user.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "Insufficient permissions",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission middleware ensures the user has the required permissions
func (m *AuthMiddleware) RequirePermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			user, ok := ctx.Value("user").(*models.User)
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "Authentication required",
				})
				return
			}

			// For now, just check if user is admin for any permission
			// Complex permission logic should use the authorization package
			if user.Role != "admin" {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "Insufficient permissions",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext extracts the authenticated user from context
func GetUserFromContext(ctx context.Context) (*models.User, error) {
	user, ok := ctx.Value("user").(*models.User)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}

// GetUserIDFromContext extracts the user ID from context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return "", errors.New("user ID not found in context")
	}
	return userID, nil
}
