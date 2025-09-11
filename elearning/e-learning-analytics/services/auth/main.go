// Auth Service API
//
// JWT-based authentication microservice for e-learning platform
//
// @title Auth Service API
// @version 1.0
// @description JWT-based authentication microservice
// @host localhost:8080
// @BasePath /
// @schemes http https
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/repositories"
	"auth-service/internal/server"

	_ "auth-service/docs" // swagger docs
)

func main() {
	// ─────────────────────────────────────────────
	// Load configuration (env vars)
	// ─────────────────────────────────────────────
	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		log.Fatalf("failed loading config: %v", err)
	}

	// ─────────────────────────────────────────────
	// Initialize shared infrastructure (Mongo, Redis)
	// ─────────────────────────────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := repositories.NewMongoClient(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("mongo connect error: %v", err)
	}
	redisClient, err := repositories.NewRedisClient(ctx, cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		_ = mongoClient.Disconnect(ctx)
		log.Fatalf("redis connect error: %v", err)
	}

	// Gracefully close on exit
	defer func() {
		shutdownCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		_ = mongoClient.Disconnect(shutdownCtx)
		_ = redisClient.Close()
	}()

	// ─────────────────────────────────────────────
	// Create and start HTTP server
	// ─────────────────────────────────────────────
	srv := server.NewServer(cfg, mongoClient, redisClient)

	// run server in background goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("server start failed: %v", err)
		}
	}()

	// ─────────────────────────────────────────────
	// Wait for termination signal and graceful shutdown
	// ─────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("shutdown initiated...")
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	} else {
		log.Println("server stopped gracefully")
	}
}
