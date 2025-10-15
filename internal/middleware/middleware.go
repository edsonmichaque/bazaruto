package middleware

import (
	"time"

	"github.com/edsonmichaque/bazaruto/internal/config"
	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/metrics"
	"github.com/edsonmichaque/bazaruto/internal/tracing"
	"github.com/go-chi/chi/v5"
)

// Register registers all middleware with the router.
func Register(r chi.Router, cfg *config.Config, logger *logger.Logger, metrics *metrics.Metrics, tracer *tracing.Tracer, rateLimitEngine *PolicyEngine) {
	// Recovery middleware (should be first)
	r.Use(RecovererMiddleware(logger))

	// Request ID middleware
	r.Use(RequestIDMiddleware())

	// Real IP middleware
	r.Use(RealIPMiddleware())

	// CORS middleware
	r.Use(CORSMiddleware())

	// Timeout middleware
	r.Use(TimeoutMiddleware(30 * time.Second))

	// Compress middleware
	r.Use(CompressMiddleware())

	// Strip slashes middleware
	r.Use(StripSlashesMiddleware())

	// Logging middleware
	r.Use(LoggingMiddleware(logger))

	// Tracing middleware
	if cfg.Tracing.Enabled {
		r.Use(TracingMiddleware(tracer))
	}

	// Metrics middleware
	if cfg.MetricsEnabled {
		r.Use(MetricsMiddleware(metrics))
	}

	// Rate limiting middleware
	if cfg.Rate.Enabled && rateLimitEngine != nil {
		r.Use(RateLimitByPolicy(rateLimitEngine))
	}
}
