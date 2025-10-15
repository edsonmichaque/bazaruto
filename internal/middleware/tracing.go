package middleware

import (
	"net/http"

	"github.com/edsonmichaque/bazaruto/internal/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

// TracingMiddleware creates a tracing middleware using OpenTelemetry.
func TracingMiddleware(tracer *tracing.Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract trace context from headers
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Start span
			_ = r.Method + " " + r.URL.Path
			spanCtx, span := tracer.StartHTTPRequestSpan(ctx, r.Method, r.URL.Path)
			defer span.End()

			// Add request attributes
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.scheme", r.URL.Scheme),
				attribute.String("http.host", r.Host),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("http.request_id", r.Header.Get("X-Request-ID")),
			)

			// Add client IP
			if clientIP := r.Header.Get("X-Forwarded-For"); clientIP != "" {
				span.SetAttributes(attribute.String("http.client_ip", clientIP))
			} else if clientIP := r.Header.Get("X-Real-IP"); clientIP != "" {
				span.SetAttributes(attribute.String("http.client_ip", clientIP))
			} else {
				span.SetAttributes(attribute.String("http.client_ip", r.RemoteAddr))
			}

			// Create response writer wrapper to capture status code
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(ww, r.WithContext(spanCtx))

			// Add response attributes
			span.SetAttributes(
				attribute.Int("http.status_code", ww.statusCode),
				attribute.Int64("http.response_size", ww.size),
			)

			// Set span status
			if ww.statusCode >= 400 {
				span.SetStatus(codes.Error, http.StatusText(ww.statusCode))
			} else {
				span.SetStatus(codes.Ok, "")
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and size.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size.
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += int64(n)
	return n, err
}
