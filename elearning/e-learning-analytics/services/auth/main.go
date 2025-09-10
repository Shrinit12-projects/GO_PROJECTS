// main.go - Auth Service Stub

package main

import (
	// "fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// HealthHandler -> simple readiness probe
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok","service":"auth-service"}`))
}

func main() {
	// ────────────────────────────────
	// Config from environment (12-factor style)
	// ────────────────────────────────
	port := getEnv("PORT", "8080")

	// ────────────────────────────────
	// Router setup
	// ────────────────────────────────
	r := mux.NewRouter()
	r.HandleFunc("/health", HealthHandler).Methods("GET")

	// ────────────────────────────────
	// HTTP server setup
	// ────────────────────────────────
	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 Auth Service running on port %s\n", port)

	// ────────────────────────────────
	// Start server (blocking call)
	// ────────────────────────────────
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ failed to start server: %v", err)
	}
}

// getEnv -> helper for env var with default
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
