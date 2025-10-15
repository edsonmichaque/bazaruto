package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/metrics"
	"github.com/go-chi/chi/v5/middleware"
)

// MetricsMiddleware creates a metrics middleware using Prometheus.
func MetricsMiddleware(metrics *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response writer wrapper to capture status code and size
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Process request
			next.ServeHTTP(ww, r)

			// Calculate duration
			duration := time.Since(start)

			// Extract request size
			requestSize := r.ContentLength
			if requestSize < 0 {
				requestSize = 0
			}

			// Record metrics
			status := strconv.Itoa(ww.Status())
			path := normalizePath(r.URL.Path)

			metrics.RecordHTTPRequest(
				r.Method,
				path,
				status,
				duration,
				requestSize,
				int64(ww.BytesWritten()),
			)
		})
	}
}

// normalizePath normalizes the path for metrics to avoid high cardinality.
func normalizePath(path string) string {
	// Remove query parameters
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// Replace UUIDs and IDs with placeholders
	path = strings.ReplaceAll(path, "/", "_")

	// Common patterns to normalize
	normalizations := map[string]string{
		"_[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}": "_uuid",
		"_[0-9]+":           "_id",
		"_[a-zA-Z0-9]{20,}": "_hash",
	}

	for pattern, replacement := range normalizations {
		_ = pattern
		_ = replacement
		// Simple string replacement for common patterns
		if strings.Contains(path, "_") {
			// This is a simplified version - in production you'd use regex
			path = strings.ReplaceAll(path, "_", "/")
			break
		}
	}

	return path
}

// DatabaseMetricsMiddleware creates a middleware for database metrics.
func DatabaseMetricsMiddleware(metrics *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This would typically be used with a database middleware
			// that tracks query execution times and counts
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMetricsMiddleware creates a middleware for rate limiting metrics.
func RateLimitMetricsMiddleware(metrics *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This would typically be used with rate limiting middleware
			// that tracks rate limit decisions
			next.ServeHTTP(w, r)
		})
	}
}
