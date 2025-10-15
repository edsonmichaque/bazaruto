package authorization

import (
	"encoding/json"
	"net/http"

	"github.com/edsonmichaque/bazaruto/internal/models"
)

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// Middleware provides authorization middleware
type Middleware struct {
	authService *Service
}

// NewMiddleware creates a new authorization middleware
func NewMiddleware(authService *Service) *Middleware {
	return &Middleware{
		authService: authService,
	}
}

// Can middleware checks if a user can perform an action
func (m *Middleware) Can(ability string, resource interface{}) func(http.Handler) http.Handler {
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

			// Check authorization
			allowed, err := m.authService.Can(ctx, user, ability, resource)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "Authorization check failed",
				})
				return
			}

			if !allowed {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "Action not allowed",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Cannot middleware checks if a user cannot perform an action
func (m *Middleware) Cannot(ability string, resource interface{}) func(http.Handler) http.Handler {
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

			// Check authorization
			allowed, err := m.authService.Can(ctx, user, ability, resource)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "Authorization check failed",
				})
				return
			}

			if allowed {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "Action not allowed",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Gate middleware checks if a user can pass through a gate
func (m *Middleware) Gate(gateName string, arguments ...interface{}) func(http.Handler) http.Handler {
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

			// Check gate authorization
			allowed, err := m.authService.AuthorizeGate(ctx, user, gateName, arguments...)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "Gate authorization check failed",
				})
				return
			}

			if !allowed {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "Gate access denied",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
