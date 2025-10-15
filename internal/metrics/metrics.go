package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics for the application.
type Metrics struct {
	registry *prometheus.Registry

	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestSize     *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec

	// Database metrics
	DatabaseConnectionsActive prometheus.Gauge
	DatabaseConnectionsIdle   prometheus.Gauge
	DatabaseQueriesTotal      *prometheus.CounterVec
	DatabaseQueryDuration     *prometheus.HistogramVec

	// Rate limiting metrics
	RateLimitRequestsTotal   *prometheus.CounterVec
	RateLimitRequestsAllowed *prometheus.CounterVec
	RateLimitRequestsDenied  *prometheus.CounterVec

	// Business metrics
	ProductsCreated   prometheus.Counter
	QuotesGenerated   prometheus.Counter
	PoliciesIssued    prometheus.Counter
	ClaimsSubmitted   prometheus.Counter
	PaymentsProcessed prometheus.Counter

	// Job metrics
	JobStarted   *prometheus.CounterVec
	JobCompleted *prometheus.CounterVec
	JobFailed    *prometheus.CounterVec
	JobDuration  *prometheus.HistogramVec
}

// NewMetrics creates a new metrics instance.
func NewMetrics() *Metrics {
	registry := prometheus.NewRegistry()

	// Register default collectors
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	metrics := &Metrics{
		registry: registry,
	}

	// Initialize HTTP metrics
	metrics.HTTPRequestsTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	metrics.HTTPRequestDuration = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	metrics.HTTPRequestSize = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	metrics.HTTPResponseSize = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path", "status"},
	)

	// Initialize database metrics
	metrics.DatabaseConnectionsActive = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_active",
			Help: "Number of active database connections",
		},
	)

	metrics.DatabaseConnectionsIdle = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	metrics.DatabaseQueriesTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	metrics.DatabaseQueryDuration = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Initialize rate limiting metrics
	metrics.RateLimitRequestsTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_requests_total",
			Help: "Total number of rate limit requests",
		},
		[]string{"policy", "key"},
	)

	metrics.RateLimitRequestsAllowed = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_requests_allowed",
			Help: "Number of rate limit requests allowed",
		},
		[]string{"policy", "key"},
	)

	metrics.RateLimitRequestsDenied = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_requests_denied",
			Help: "Number of rate limit requests denied",
		},
		[]string{"policy", "key"},
	)

	// Initialize business metrics
	metrics.ProductsCreated = promauto.With(registry).NewCounter(
		prometheus.CounterOpts{
			Name: "products_created_total",
			Help: "Total number of products created",
		},
	)

	metrics.QuotesGenerated = promauto.With(registry).NewCounter(
		prometheus.CounterOpts{
			Name: "quotes_generated_total",
			Help: "Total number of quotes generated",
		},
	)

	metrics.PoliciesIssued = promauto.With(registry).NewCounter(
		prometheus.CounterOpts{
			Name: "policies_issued_total",
			Help: "Total number of policies issued",
		},
	)

	metrics.ClaimsSubmitted = promauto.With(registry).NewCounter(
		prometheus.CounterOpts{
			Name: "claims_submitted_total",
			Help: "Total number of claims submitted",
		},
	)

	metrics.PaymentsProcessed = promauto.With(registry).NewCounter(
		prometheus.CounterOpts{
			Name: "payments_processed_total",
			Help: "Total number of payments processed",
		},
	)

	// Initialize job metrics
	metrics.JobStarted = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_started_total",
			Help: "Total number of jobs started",
		},
		[]string{"queue", "type"},
	)

	metrics.JobCompleted = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_completed_total",
			Help: "Total number of jobs completed",
		},
		[]string{"queue", "type"},
	)

	metrics.JobFailed = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_failed_total",
			Help: "Total number of jobs failed",
		},
		[]string{"queue", "type"},
	)

	metrics.JobDuration = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "job_duration_seconds",
			Help:    "Job execution duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"queue", "type"},
	)

	return metrics
}

// Handler returns the Prometheus metrics handler.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// RecordHTTPRequest records HTTP request metrics.
func (m *Metrics) RecordHTTPRequest(method, path, status string, duration time.Duration, requestSize, responseSize int64) {
	statusCode := status
	m.HTTPRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path, statusCode).Observe(duration.Seconds())
	m.HTTPRequestSize.WithLabelValues(method, path).Observe(float64(requestSize))
	m.HTTPResponseSize.WithLabelValues(method, path, statusCode).Observe(float64(responseSize))
}

// RecordDatabaseQuery records database query metrics.
func (m *Metrics) RecordDatabaseQuery(operation, table string, duration time.Duration) {
	m.DatabaseQueriesTotal.WithLabelValues(operation, table).Inc()
	m.DatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordRateLimitRequest records rate limiting metrics.
func (m *Metrics) RecordRateLimitRequest(policy, key string, allowed bool) {
	m.RateLimitRequestsTotal.WithLabelValues(policy, key).Inc()
	if allowed {
		m.RateLimitRequestsAllowed.WithLabelValues(policy, key).Inc()
	} else {
		m.RateLimitRequestsDenied.WithLabelValues(policy, key).Inc()
	}
}

// UpdateDatabaseConnections updates database connection metrics.
func (m *Metrics) UpdateDatabaseConnections(active, idle int) {
	m.DatabaseConnectionsActive.Set(float64(active))
	m.DatabaseConnectionsIdle.Set(float64(idle))
}

// IncrementProductCreated increments the products created counter.
func (m *Metrics) IncrementProductCreated() {
	m.ProductsCreated.Inc()
}

// IncrementQuoteGenerated increments the quotes generated counter.
func (m *Metrics) IncrementQuoteGenerated() {
	m.QuotesGenerated.Inc()
}

// IncrementPolicyIssued increments the policies issued counter.
func (m *Metrics) IncrementPolicyIssued() {
	m.PoliciesIssued.Inc()
}

// IncrementClaimSubmitted increments the claims submitted counter.
func (m *Metrics) IncrementClaimSubmitted() {
	m.ClaimsSubmitted.Inc()
}

// IncrementPaymentProcessed increments the payments processed counter.
func (m *Metrics) IncrementPaymentProcessed() {
	m.PaymentsProcessed.Inc()
}

// StartMetricsServer starts the metrics server.
func (m *Metrics) StartMetricsServer(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", m.Handler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	return server.ListenAndServe()
}
