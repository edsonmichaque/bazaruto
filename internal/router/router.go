package router

import (
	"github.com/edsonmichaque/bazaruto/internal/config"
	"github.com/edsonmichaque/bazaruto/internal/database"
	"github.com/edsonmichaque/bazaruto/internal/handlers"
	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/metrics"
	"github.com/edsonmichaque/bazaruto/internal/middleware"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/edsonmichaque/bazaruto/internal/tracing"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Register registers all routes and middleware with the router.
func Register(r chi.Router, cfg *config.Config, db *database.Database, logger *logger.Logger, metrics *metrics.Metrics, tracer *tracing.Tracer) func() {
	// Create stores
	stores := store.NewStores(db.DB)

	// Create services
	productService := services.NewProductService(stores.Products)
	quoteService := services.NewQuoteService(stores.Quotes)
	policyService := services.NewPolicyService(stores.Policies)
	claimService := services.NewClaimService(stores.Claims, stores.Policies)

	// Create handlers
	productHandler := handlers.NewProductHandler(productService)
	quoteHandler := handlers.NewQuoteHandler(quoteService)
	policyHandler := handlers.NewPolicyHandler(policyService)
	claimHandler := handlers.NewClaimHandler(claimService)
	healthHandler := handlers.NewHealthHandler(db)
	versionHandler := handlers.NewVersionHandler()

	// Create rate limiting engine
	var rateLimitEngine *middleware.PolicyEngine
	var rateLimitCloser func()
	if cfg.Rate.Enabled {
		engine, closer, err := middleware.BuildPolicyEngine(cfg.Rate, cfg.Redis)
		if err != nil {
			logger.Error("Failed to create rate limiting engine", zap.Error(err))
		} else {
			rateLimitEngine = engine
			rateLimitCloser = closer
		}
	}

	// Register middleware
	middleware.Register(r, cfg, logger, metrics, tracer, rateLimitEngine)

	// Register health check endpoint
	r.Get("/healthz", healthHandler.HealthCheck)

	// Register version endpoint
	r.Get("/version", versionHandler.GetVersion)

	// Register metrics endpoint
	if cfg.MetricsEnabled {
		r.Get(cfg.MetricsPath, metrics.Handler().ServeHTTP)
	}

	// Register API routes
	r.Route("/v1", func(r chi.Router) {
		// Register handler routes
		productHandler.RegisterRoutes(r)
		quoteHandler.RegisterRoutes(r)
		policyHandler.RegisterRoutes(r)
		claimHandler.RegisterRoutes(r)
	})

	// Return cleanup function
	return func() {
		if rateLimitCloser != nil {
			rateLimitCloser()
		}
	}
}
