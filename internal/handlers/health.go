package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/database"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	db *database.Database
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *database.Database) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// HealthCheck handles GET /healthz.
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	_, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)

	// Check database health
	if h.db != nil {
		if err := h.db.Health(); err != nil {
			services["database"] = "unhealthy"
			_ = writeHealthResponse(w, "unhealthy", services)
			return
		}
		services["database"] = "healthy"
	} else {
		services["database"] = "not configured"
	}

	// All services are healthy
	_ = writeHealthResponse(w, "healthy", services)
}
