// internal/config/config.go
package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	Port         string        // HTTP port for server to bind to (e.g., "8080")
	MongoURI     string        // MongoDB connection URI
	RedisAddr    string        // Redis address (host:port)
	RedisPassword string       // Redis password
	JWTSecret    string        // HMAC secret for signing JWT tokens
	AccessTTL    time.Duration // access token TTL
	RefreshTTL   time.Duration // refresh token TTL (stored in Redis)
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// NewConfigFromEnv builds a Config reading from environment variables.
// This enforces the 12-factor pattern and supplies reasonable defaults.
func NewConfigFromEnv() (*Config, error) {
	port := getEnv("PORT", "8080")
	mongoURI := getEnv("DB_URI", "mongodb://mongodb:27017/elearn")

	redisAddr := getEnv("REDIS_URL", "redis:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	accessSeconds := getEnvInt("JWT_ACCESS_TTL_SECONDS", 900)  // 15m
	refreshSeconds := getEnvInt("JWT_REFRESH_TTL_SECONDS", 604800) // 7d

	readTimeout := getEnvInt("READ_TIMEOUT_SECONDS", 15)
	writeTimeout := getEnvInt("WRITE_TIMEOUT_SECONDS", 15)
	idleTimeout := getEnvInt("IDLE_TIMEOUT_SECONDS", 60)

	return &Config{
		Port:         port,
		MongoURI:     mongoURI,
		RedisAddr:    redisAddr,
		RedisPassword: redisPassword,
		JWTSecret:    jwtSecret,
		AccessTTL:    time.Duration(accessSeconds) * time.Second,
		RefreshTTL:   time.Duration(refreshSeconds) * time.Second,
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		IdleTimeout:  time.Duration(idleTimeout) * time.Second,
	}, nil
}

// helpers
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
