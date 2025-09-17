// internal/middleware/logging.go
package middleware

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"auth-service/internal/metrics"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger logs HTTP requests and records Prometheus metrics
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Call the next handler
		next.ServeHTTP(wrapped, r)
		
		duration := time.Since(start)
		statusCode := strconv.Itoa(wrapped.statusCode)
		
		// Normalize endpoint for metrics (avoid high cardinality)
		endpoint := normalizeEndpoint(r.URL.Path)
		
		// Record metrics
		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, endpoint, statusCode).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, endpoint).Observe(duration.Seconds())
		
		// Log request details
		log.Printf("%s %s %s %v", r.Method, r.URL.Path, statusCode, duration)
	})
}

// normalizeEndpoint converts dynamic paths to static labels for metrics
func normalizeEndpoint(path string) string {
	switch {
	case path == "/health":
		return "/health"
	case path == "/metrics":
		return "/metrics"
	case path == "/auth/login":
		return "/auth/login"
	case path == "/auth/register":
		return "/auth/register"
	case path == "/auth/logout":
		return "/auth/logout"
	case path == "/auth/refresh":
		return "/auth/refresh"
	default:
		if strings.HasPrefix(path, "/swagger/") {
			return "/swagger/*"
		}
		return "/unknown"
	}
}