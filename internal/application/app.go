package application

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/config"
	"github.com/edsonmichaque/bazaruto/internal/database"
	"github.com/edsonmichaque/bazaruto/internal/events/handlers"
	"github.com/edsonmichaque/bazaruto/internal/jobs"
	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/metrics"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/edsonmichaque/bazaruto/internal/tracing"
	"github.com/edsonmichaque/bazaruto/pkg/event"
	"github.com/edsonmichaque/bazaruto/pkg/job"
	"go.uber.org/zap"
)

// Application represents the fully wired application with all dependencies
type Application struct {
	// Core infrastructure
	Config   *config.Config
	Logger   *logger.Logger
	Metrics  *metrics.Metrics
	Tracer   *tracing.Tracer
	Database *database.Database

	// Event system
	EventBus     event.EventBus
	EventService *services.EventService

	// Job system
	JobManager    *job.Manager
	JobDispatcher job.Dispatcher

	// Data stores
	UserStore         store.UserStore
	CustomerStore     store.CustomerStore
	PartnerStore      store.PartnerStore
	ProductStore      store.ProductStore
	QuoteStore        store.QuoteStore
	PolicyStore       store.PolicyStore
	ClaimStore        store.ClaimStore
	SubscriptionStore store.SubscriptionStore
	PaymentStore      store.PaymentStore
	InvoiceStore      store.InvoiceStore
	BeneficiaryStore  store.BeneficiaryStore
	CoverageStore     store.CoverageStore
	WebhookStore      store.WebhookStore

	// Business services
	ProductService         *services.ProductService
	QuoteService           *services.QuoteService
	PolicyService          *services.PolicyService
	ClaimService           *services.ClaimService
	UserService            *services.UserService
	PaymentService         *services.PaymentService
	WebhookService         *services.WebhookService
	FraudDetectionService  *services.FraudDetectionService
	RiskAssessmentService  *services.RiskAssessmentService
	UnderwritingService    *services.UnderwritingService
	PricingEngineService   *services.PricingEngineService
	CommissionService      *services.CommissionService
	ComplianceService      *services.ComplianceService
	PolicyLifecycleService *services.PolicyLifecycleService
	ClaimProcessingService *services.ClaimProcessingService

	// Configuration management
	ConfigManager *config.Manager

	// Event handlers
	UserEventHandlers    []event.EventHandler
	QuoteEventHandlers   []event.EventHandler
	PaymentEventHandlers []event.EventHandler
	PolicyEventHandlers  []event.EventHandler
	ClaimEventHandlers   []event.EventHandler
	WebhookEventHandlers []event.EventHandler

	// Server and worker management
	server         *http.Server
	handler        http.Handler
	serverMu       sync.RWMutex
	workersMu      sync.RWMutex
	workersStarted bool
}

// NewApplication creates a fully wired application with all dependencies
func NewApplication(ctx context.Context, cfg *config.Config) (*Application, error) {
	app := &Application{
		Config: cfg,
	}

	// Initialize core infrastructure
	if err := app.initializeInfrastructure(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Initialize event system
	if err := app.initializeEventSystem(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize event system: %w", err)
	}

	// Initialize job system
	if err := app.initializeJobSystem(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize job system: %w", err)
	}

	// Initialize stores
	if err := app.initializeStores(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize stores: %w", err)
	}

	// Initialize configuration management
	if err := app.initializeConfigManagement(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize config management: %w", err)
	}

	// Initialize business services
	if err := app.initializeBusinessServices(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize business services: %w", err)
	}

	// Initialize event handlers
	if err := app.initializeEventHandlers(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize event handlers: %w", err)
	}

	// Wire event handlers to event bus
	if err := app.wireEventHandlers(ctx); err != nil {
		return nil, fmt.Errorf("failed to wire event handlers: %w", err)
	}

	// Register job types
	if err := app.registerJobTypes(ctx); err != nil {
		return nil, fmt.Errorf("failed to register job types: %w", err)
	}

	app.Logger.Info("Application fully wired and ready")
	return app, nil
}

// initializeInfrastructure initializes core infrastructure components
func (app *Application) initializeInfrastructure(ctx context.Context) error {
	// Initialize logger
	app.Logger = logger.NewLogger(app.Config.LogLevel, app.Config.LogFormat)

	// Initialize metrics
	app.Metrics = metrics.NewMetrics()

	// Initialize tracer if enabled
	if app.Config.Tracing.Enabled {
		tracer, err := tracing.NewTracer(app.Config.Tracing.ServiceName, app.Config.Tracing.Endpoint)
		if err != nil {
			app.Logger.Error("Failed to initialize tracer", zap.Error(err))
			return fmt.Errorf("failed to initialize tracer: %w", err)
		}
		app.Tracer = tracer
	}

	// Connect to database
	db, err := database.Connect(app.Config.DB.DSN, database.DBConfig{
		MaxConnections: app.Config.DB.MaxConnections,
		MinConnections: app.Config.DB.MinConnections,
		ConnectTimeout: app.Config.DB.ConnectTimeout,
		AcquireTimeout: app.Config.DB.AcquireTimeout,
		MaxLifetime:    app.Config.DB.MaxLifetime,
		IdleTimeout:    app.Config.DB.IdleTimeout,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	app.Database = db

	app.Logger.Info("Infrastructure initialized successfully")
	return nil
}

// initializeEventSystem initializes the event system
func (app *Application) initializeEventSystem(ctx context.Context) error {
	// Create event bus
	app.EventBus = event.NewBus()

	// Create event service
	app.EventService = services.NewEventService(app.EventBus, app.Logger)

	app.Logger.Info("Event system initialized successfully")
	return nil
}

// initializeJobSystem initializes the job system
func (app *Application) initializeJobSystem(ctx context.Context) error {
	// Create job manager configuration
	jobConfig := job.ManagerConfig{
		Adapter: job.AdapterTypeMemory, // Use memory adapter for now
		Queues: []string{
			job.QueueMailers,
			job.QueueProcessing,
			job.QueuePayments,
			job.QueueNotifications,
			job.QueueClaims,
		},
		Concurrency:  5,
		PollInterval: 5,
		Timeout:      30,
		MaxRetries:   3,
	}

	// Create job manager
	manager, err := job.NewManager(jobConfig, app.Logger, app.Metrics, app.Tracer)
	if err != nil {
		return fmt.Errorf("failed to create job manager: %w", err)
	}
	app.JobManager = manager
	app.JobDispatcher = *manager.Dispatcher()

	// Note: Job manager will be started via StartWorkers() method

	app.Logger.Info("Job system initialized successfully")
	return nil
}

// initializeStores initializes all data stores
func (app *Application) initializeStores(ctx context.Context) error {
	app.UserStore = store.NewUserStore(app.Database.DB)
	app.CustomerStore = store.NewCustomerStore(app.Database.DB)
	app.PartnerStore = store.NewPartnerStore(app.Database.DB)
	app.ProductStore = store.NewProductStore(app.Database.DB)
	app.QuoteStore = store.NewQuoteStore(app.Database.DB)
	app.PolicyStore = store.NewPolicyStore(app.Database.DB)
	app.ClaimStore = store.NewClaimStore(app.Database.DB)
	app.SubscriptionStore = store.NewSubscriptionStore(app.Database.DB)
	app.PaymentStore = store.NewPaymentStore(app.Database.DB)
	app.InvoiceStore = store.NewInvoiceStore(app.Database.DB)
	app.BeneficiaryStore = store.NewBeneficiaryStore(app.Database.DB)
	app.CoverageStore = store.NewCoverageStore(app.Database.DB)
	app.WebhookStore = store.NewWebhookStore(app.Database.DB)

	app.Logger.Info("Stores initialized successfully")
	return nil
}

// initializeConfigManagement initializes configuration management
func (app *Application) initializeConfigManagement(ctx context.Context) error {
	// Create config manager with default path
	configPath := "business_rules.json"
	app.ConfigManager = config.NewManager(app.Logger, configPath)

	// Load initial configuration
	if err := app.ConfigManager.LoadConfig(ctx); err != nil {
		app.Logger.Warn("Failed to load business rules config, using defaults", zap.Error(err))
	}

	app.Logger.Info("Configuration management initialized successfully")
	return nil
}

// initializeBusinessServices initializes all business services
func (app *Application) initializeBusinessServices(ctx context.Context) error {
	// Basic CRUD services
	app.ProductService = services.NewProductService(app.ProductStore)
	app.QuoteService = services.NewQuoteService(app.QuoteStore)
	app.PolicyService = services.NewPolicyService(app.PolicyStore)
	app.ClaimService = services.NewClaimService(app.ClaimStore, app.PolicyStore)
	app.UserService = services.NewUserService(app.UserStore)
	app.PaymentService = services.NewPaymentService(app.PaymentStore, app.EventService)
	app.WebhookService = services.NewWebhookService(app.WebhookStore)

	// Advanced business services
	app.FraudDetectionService = services.NewFraudDetectionService(
		app.Logger,
		app.ConfigManager,
		app.ClaimStore,
		app.PolicyStore,
		app.CustomerStore,
		app.EventService,
	)

	app.RiskAssessmentService = services.NewRiskAssessmentService(
		app.UserStore,
		app.PolicyStore,
		app.ClaimStore,
	)

	app.UnderwritingService = services.NewUnderwritingService(
		app.UserStore,
		app.PolicyStore,
		app.ClaimStore,
		app.RiskAssessmentService,
		app.FraudDetectionService,
		nil, // PricingEngineService - will be created below
	)

	app.PricingEngineService = services.NewPricingEngineService(
		app.ProductStore,
		app.PolicyStore,
		app.ClaimStore,
		app.UserStore,
	)

	app.CommissionService = services.NewCommissionService(
		app.PartnerStore,
		app.PolicyStore,
		app.PaymentStore,
		nil, // commissionStore - placeholder
	)

	app.ComplianceService = services.NewComplianceService(
		app.UserStore,
		app.PolicyStore,
		app.ClaimStore,
		app.PaymentStore,
	)

	app.PolicyLifecycleService = services.NewPolicyLifecycleService(
		app.Logger,
		app.ConfigManager,
		app.PolicyStore,
		app.PaymentStore,
		app.SubscriptionStore,
		app.UserStore,
		app.EventService,
	)

	app.ClaimProcessingService = services.NewClaimProcessingService(
		app.ClaimStore,
		app.PolicyStore,
		app.UserStore,
		app.FraudDetectionService,
		app.RiskAssessmentService,
		app.EventService,
		app.JobDispatcher,
	)

	// Update underwriting service with pricing service
	app.UnderwritingService = services.NewUnderwritingService(
		app.UserStore,
		app.PolicyStore,
		app.ClaimStore,
		app.RiskAssessmentService,
		app.FraudDetectionService,
		app.PricingEngineService,
	)

	app.Logger.Info("Business services initialized successfully")
	return nil
}

// initializeEventHandlers initializes all event handlers
func (app *Application) initializeEventHandlers(ctx context.Context) error {
	// User event handlers
	app.UserEventHandlers = []event.EventHandler{
		handlers.NewUserRegisteredHandler(app.UserService, app.JobDispatcher, app.Logger),
		handlers.NewUserLoggedInHandler(app.UserService, app.Logger),
	}

	// Quote event handlers
	app.QuoteEventHandlers = []event.EventHandler{
		handlers.NewQuoteCreatedHandler(app.QuoteService, app.JobDispatcher, app.Logger),
		handlers.NewQuoteCalculatedHandler(app.JobDispatcher, app.Logger),
	}

	// Payment event handlers
	app.PaymentEventHandlers = []event.EventHandler{
		handlers.NewPaymentInitiatedHandler(app.PaymentService, app.JobDispatcher, app.Logger),
		handlers.NewPaymentCompletedHandler(app.JobDispatcher, app.Logger),
		handlers.NewPaymentFailedHandler(app.JobDispatcher, app.Logger),
	}

	// Policy event handlers
	app.PolicyEventHandlers = []event.EventHandler{
		handlers.NewPolicyEventHandler(app.JobDispatcher, app.Logger.Logger),
	}

	// Claim event handlers
	app.ClaimEventHandlers = []event.EventHandler{
		handlers.NewClaimSubmittedHandler(app.ClaimService, app.JobDispatcher, app.Logger),
	}

	// Webhook event handlers
	app.WebhookEventHandlers = []event.EventHandler{
		handlers.NewWebhookEventHandler(app.WebhookService, app.JobDispatcher, app.Logger),
	}

	app.Logger.Info("Event handlers initialized successfully")
	return nil
}

// wireEventHandlers wires all event handlers to the event bus
func (app *Application) wireEventHandlers(ctx context.Context) error {
	// Wire user event handlers
	for _, handler := range app.UserEventHandlers {
		if err := app.EventService.SubscribeHandler(handler, "user.registered", "user.logged_in"); err != nil {
			return fmt.Errorf("failed to subscribe user event handler: %w", err)
		}
	}

	// Wire quote event handlers
	for _, handler := range app.QuoteEventHandlers {
		if err := app.EventService.SubscribeHandler(handler, "quote.created", "quote.calculated"); err != nil {
			return fmt.Errorf("failed to subscribe quote event handler: %w", err)
		}
	}

	// Wire payment event handlers
	for _, handler := range app.PaymentEventHandlers {
		if err := app.EventService.SubscribeHandler(handler, "payment.initiated", "payment.completed", "payment.failed"); err != nil {
			return fmt.Errorf("failed to subscribe payment event handler: %w", err)
		}
	}

	// Wire policy event handlers
	for _, handler := range app.PolicyEventHandlers {
		if err := app.EventService.SubscribeHandler(handler, "policy.created", "policy.renewed", "policy.cancelled", "policy.expired", "grace_period.expired", "renewal.reminder"); err != nil {
			return fmt.Errorf("failed to subscribe policy event handler: %w", err)
		}
	}

	// Wire claim event handlers
	for _, handler := range app.ClaimEventHandlers {
		if err := app.EventService.SubscribeHandler(handler, "claim.submitted"); err != nil {
			return fmt.Errorf("failed to subscribe claim event handler: %w", err)
		}
	}

	// Wire webhook event handlers
	for _, handler := range app.WebhookEventHandlers {
		if err := app.EventService.SubscribeHandler(handler, "user.registered", "user.logged_in", "quote.created", "quote.calculated", "payment.initiated", "payment.completed", "payment.failed", "policy.created", "claim.submitted"); err != nil {
			return fmt.Errorf("failed to subscribe webhook event handler: %w", err)
		}
	}

	app.Logger.Info("Event handlers wired successfully")
	return nil
}

// registerJobTypes registers all job types with the job manager
func (app *Application) registerJobTypes(ctx context.Context) error {
	// Email jobs
	app.JobManager.Registry().RegisterJob(&jobs.SendEmailJob{})
	app.JobManager.Registry().RegisterJob(&jobs.WelcomeEmailJob{})

	// PDF jobs
	app.JobManager.Registry().RegisterJob(&jobs.GenerateQuotePDFJob{})

	// Payment jobs
	app.JobManager.Registry().RegisterJob(&jobs.ProcessPaymentJob{})

	// Processing jobs
	app.JobManager.Registry().RegisterJob(&jobs.CalculatePremiumJob{})
	app.JobManager.Registry().RegisterJob(&jobs.FraudDetectionJob{})

	// Notification jobs
	app.JobManager.Registry().RegisterJob(&jobs.PushNotificationJob{})

	app.Logger.Info("Job types registered successfully")
	return nil
}

// SetHandler sets the HTTP handler for the server
func (app *Application) SetHandler(handler http.Handler) {
	app.serverMu.Lock()
	defer app.serverMu.Unlock()

	app.handler = handler
	if app.server != nil {
		app.server.Handler = handler
	}
}

// StartServer starts the HTTP server
func (app *Application) StartServer(ctx context.Context) error {
	app.serverMu.Lock()
	defer app.serverMu.Unlock()

	if app.server != nil {
		return fmt.Errorf("server is already running")
	}

	// Create HTTP server
	app.server = &http.Server{
		Addr:              app.Config.Server.Addr,
		Handler:           app.handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       app.Config.Server.ReadTimeout,
		WriteTimeout:      app.Config.Server.WriteTimeout,
		IdleTimeout:       app.Config.Server.IdleTimeout,
	}

	app.Logger.Info("Starting HTTP server", zap.String("addr", app.Config.Server.Addr))

	// Start server in a goroutine
	go func() {
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	return nil
}

// StopServer stops the HTTP server
func (app *Application) StopServer(ctx context.Context) error {
	app.serverMu.Lock()
	defer app.serverMu.Unlock()

	if app.server == nil {
		return fmt.Errorf("server is not running")
	}

	app.Logger.Info("Stopping HTTP server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Shutdown server
	if err := app.server.Shutdown(shutdownCtx); err != nil {
		app.Logger.Error("Failed to shutdown server gracefully", zap.Error(err))
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	app.server = nil
	app.Logger.Info("HTTP server stopped")
	return nil
}

// StartWorkers starts the job workers
func (app *Application) StartWorkers(ctx context.Context) error {
	app.workersMu.Lock()
	defer app.workersMu.Unlock()

	if app.workersStarted {
		return fmt.Errorf("workers are already started")
	}

	app.Logger.Info("Starting job workers...")

	// Start job manager
	if err := app.JobManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start job manager: %w", err)
	}

	app.workersStarted = true
	app.Logger.Info("Job workers started")
	return nil
}

// StopWorkers stops the job workers
func (app *Application) StopWorkers(ctx context.Context) error {
	app.workersMu.Lock()
	defer app.workersMu.Unlock()

	if !app.workersStarted {
		return fmt.Errorf("workers are not started")
	}

	app.Logger.Info("Stopping job workers...")

	// Stop job manager
	if app.JobManager != nil {
		app.JobManager.Stop()
	}

	app.workersStarted = false
	app.Logger.Info("Job workers stopped")
	return nil
}

// Close gracefully shuts down the application
func (app *Application) Close() error {
	app.Logger.Info("Shutting down application...")

	// Stop server if running
	if app.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := app.StopServer(ctx); err != nil {
			app.Logger.Error("Failed to stop server during shutdown", zap.Error(err))
		}
	}

	// Stop workers if running
	if app.workersStarted {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := app.StopWorkers(ctx); err != nil {
			app.Logger.Error("Failed to stop workers during shutdown", zap.Error(err))
		}
	}

	// Close event service
	if app.EventService != nil {
		if err := app.EventService.Close(); err != nil {
			app.Logger.Error("Failed to close event service", zap.Error(err))
		}
	}

	// Close database
	if app.Database != nil {
		if err := app.Database.Close(); err != nil {
			app.Logger.Error("Failed to close database", zap.Error(err))
		}
	}

	app.Logger.Info("Application shutdown complete")
	return nil
}

// GetService returns a service by name for external access
func (app *Application) GetService(name string) interface{} {
	switch name {
	case "product":
		return app.ProductService
	case "quote":
		return app.QuoteService
	case "policy":
		return app.PolicyService
	case "claim":
		return app.ClaimService
	case "user":
		return app.UserService
	case "payment":
		return app.PaymentService
	case "webhook":
		return app.WebhookService
	case "fraud":
		return app.FraudDetectionService
	case "risk":
		return app.RiskAssessmentService
	case "underwriting":
		return app.UnderwritingService
	case "pricing":
		return app.PricingEngineService
	case "commission":
		return app.CommissionService
	case "compliance":
		return app.ComplianceService
	case "policy_lifecycle":
		return app.PolicyLifecycleService
	case "claim_processing":
		return app.ClaimProcessingService
	case "event":
		return app.EventService
	case "config":
		return app.ConfigManager
	default:
		return nil
	}
}
