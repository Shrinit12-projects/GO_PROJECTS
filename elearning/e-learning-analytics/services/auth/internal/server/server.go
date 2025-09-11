// internal/server/server.go
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"auth-service/internal/config"
	"auth-service/internal/handler"
	"auth-service/internal/middleware"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/redis/go-redis/v9"
)

// Server encapsulates http.Server and references to infra clients.
type Server struct {
	cfg    *config.Config
	mux    *mux.Router
	server *http.Server

	Mongo *mongo.Client
	Redis *redis.Client
}

// NewServer creates a fully wired server instance (routes + middleware).
func NewServer(cfg *config.Config, mongoClient *mongo.Client, redisClient *redis.Client) *Server {
	r := mux.NewRouter()

	// Global middleware
	r.Use(middleware.RequestLogger) // request logging
	// CORS, recovery, rate-limiting middleware can be added here

	s := &Server{
		cfg:    cfg,
		mux:    r,
		Mongo:  mongoClient,
		Redis:  redisClient,
	}

	// Register handlers (grouped)
	s.registerRoutes()

	// HTTP server
	svr := &http.Server{
		Handler:      r,
		Addr:         ":" + cfg.Port,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	s.server = svr
	return s
}

// Start launches HTTP server (blocking) — returns error if ListenAndServe fails.
func (s *Server) Start() error {
	log.Printf("starting HTTP server on %s\n", s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server with context deadline.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("shutting down HTTP server...")
	return s.server.Shutdown(ctx)
}

// registerRoutes wires endpoints to handlers and middlewares.
func (s *Server) registerRoutes() {
	// health
	s.mux.HandleFunc("/health", handler.HealthHandler).Methods("GET")

	// swagger
	s.mux.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// auth group
	authH := handler.NewAuthHandler(s.Mongo, s.Redis, s.cfg)
	authRouter := s.mux.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/register", authH.Register).Methods("POST")
	authRouter.HandleFunc("/login", authH.Login).Methods("POST")
	authRouter.HandleFunc("/logout", middleware.RequireAuth(s.cfg, s.Redis, authH.Logout)).Methods("DELETE")
	authRouter.HandleFunc("/refresh", authH.Refresh).Methods("POST")

	// other services (course/progress/analytics) will be registered similarly
	// e.g. s.mux.PathPrefix("/courses").HandlerFunc(...)
	_ = fmt.Sprintf // keep import usages tidy if you expand later
}
