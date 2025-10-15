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

// Router represents the application router with all dependencies
type Router struct {
	chi.Router
	cfg     *config.Config
	db      *database.Database
	logger  *logger.Logger
	metrics *metrics.Metrics
	tracer  *tracing.Tracer

	// Services
	productService *services.ProductService
	quoteService   *services.QuoteService
	policyService  *services.PolicyService
	claimService   *services.ClaimService

	// Handlers
	productHandler *handlers.ProductHandler
	quoteHandler   *handlers.QuoteHandler
	policyHandler  *handlers.PolicyHandler
	claimHandler   *handlers.ClaimHandler
	healthHandler  *handlers.HealthHandler
	versionHandler *handlers.VersionHandler

	// Middleware
	rateLimitEngine *middleware.PolicyEngine
	rateLimitCloser func()
}

// New creates a new Router instance with all dependencies
func New(r chi.Router, cfg *config.Config, db *database.Database, logger *logger.Logger, metrics *metrics.Metrics, tracer *tracing.Tracer) *Router {
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

	return &Router{
		Router:          r,
		cfg:             cfg,
		db:              db,
		logger:          logger,
		metrics:         metrics,
		tracer:          tracer,
		productService:  productService,
		quoteService:    quoteService,
		policyService:   policyService,
		claimService:    claimService,
		productHandler:  productHandler,
		quoteHandler:    quoteHandler,
		policyHandler:   policyHandler,
		claimHandler:    claimHandler,
		healthHandler:   healthHandler,
		versionHandler:  versionHandler,
		rateLimitEngine: rateLimitEngine,
		rateLimitCloser: rateLimitCloser,
	}
}

// RegisterRoutes registers all routes and middleware with the router
func (rt *Router) RegisterRoutes() {
	// Register middleware
	middleware.Register(rt.Router, rt.cfg, rt.logger, rt.metrics, rt.tracer, rt.rateLimitEngine)

	// Register health check endpoint
	rt.Get("/healthz", rt.healthHandler.HealthCheck)

	// Register version endpoint
	rt.Get("/version", rt.versionHandler.GetVersion)

	// Register metrics endpoint
	if rt.cfg.MetricsEnabled {
		rt.Get(rt.cfg.MetricsPath, rt.metrics.Handler().ServeHTTP)
	}

	// Register API routes
	rt.Route("/v1", func(r chi.Router) {
		// Register handler routes
		rt.productHandler.RegisterRoutes(r)
		rt.quoteHandler.RegisterRoutes(r)
		rt.policyHandler.RegisterRoutes(r)
		rt.claimHandler.RegisterRoutes(r)
	})
}

// Close cleans up resources used by the router
func (rt *Router) Close() {
	if rt.rateLimitCloser != nil {
		rt.rateLimitCloser()
	}
}

// GetService returns a service by name for testing or advanced usage
func (rt *Router) GetService(name string) interface{} {
	switch name {
	case "product":
		return rt.productService
	case "quote":
		return rt.quoteService
	case "policy":
		return rt.policyService
	case "claim":
		return rt.claimService
	default:
		return nil
	}
}

// GetHandler returns a handler by name for testing or advanced usage
func (rt *Router) GetHandler(name string) interface{} {
	switch name {
	case "product":
		return rt.productHandler
	case "quote":
		return rt.quoteHandler
	case "policy":
		return rt.policyHandler
	case "claim":
		return rt.claimHandler
	case "health":
		return rt.healthHandler
	case "version":
		return rt.versionHandler
	default:
		return nil
	}
}

// Register is a convenience function that maintains backward compatibility
func Register(r chi.Router, cfg *config.Config, db *database.Database, logger *logger.Logger, metrics *metrics.Metrics, tracer *tracing.Tracer) func() {
	router := New(r, cfg, db, logger, metrics, tracer)
	router.RegisterRoutes()
	return router.Close
}
