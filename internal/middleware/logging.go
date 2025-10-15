package middleware

import (
	"net/http"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// LoggingMiddleware creates a logging middleware using zerolog.
func LoggingMiddleware(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Create request logger with context
			reqLogger := logger.WithRequest(r.Method, r.URL.Path, r.UserAgent())

			// Log request
			reqLogger.Info("Request started",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.String("ip", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.String("referer", r.Referer()))

			// Process request
			next.ServeHTTP(ww, r)

			// Calculate duration
			duration := time.Since(start)

			// Log response
			reqLogger.Info("Request completed",
				zap.Int("status", ww.Status()),
				zap.Int("size", ww.BytesWritten()),
				zap.Duration("duration", duration))
		})
	}
}

// RequestIDMiddleware adds a request ID to each request.
func RequestIDMiddleware() func(http.Handler) http.Handler {
	return middleware.RequestID
}

// RealIPMiddleware sets the real IP from headers.
func RealIPMiddleware() func(http.Handler) http.Handler {
	return middleware.RealIP
}

// RecovererMiddleware recovers from panics and logs them.
func RecovererMiddleware(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Panic recovered",
						zap.Any("panic", err),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.String("ip", r.RemoteAddr))

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing.
func CORSMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
			w.Header().Set("Access-Control-Expose-Headers", "Link")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "300")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TimeoutMiddleware sets a timeout for requests.
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return middleware.Timeout(timeout)
}

// CompressMiddleware compresses responses.
func CompressMiddleware() func(http.Handler) http.Handler {
	return middleware.Compress(5)
}

// StripSlashesMiddleware removes trailing slashes from URLs.
func StripSlashesMiddleware() func(http.Handler) http.Handler {
	return middleware.StripSlashes
}

// RedirectSlashesMiddleware redirects URLs with trailing slashes.
func RedirectSlashesMiddleware() func(http.Handler) http.Handler {
	return middleware.RedirectSlashes
}
