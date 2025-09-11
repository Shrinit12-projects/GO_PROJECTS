// internal/middleware/logging.go
package middleware

import (
	"log"
	"net/http"
	"time"
)

// RequestLogger logs HTTP requests with basic information
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Call the next handler
		next.ServeHTTP(w, r)
		
		// Log request details
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}