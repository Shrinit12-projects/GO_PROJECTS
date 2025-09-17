// internal/middleware/jwt.go
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/metrics"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// Context key types for type safety
type contextKey string

const (
	UserIDKey contextKey = "sub"
	JTIKey    contextKey = "jti"
)

// RequireAuth is a middleware constructor that verifies JWT and checks Redis blacklist.
// On success it sets "sub" and "jti" in the request context for handlers to consume.
func RequireAuth(cfg *config.Config, redisClient *redis.Client, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			metrics.JWTValidationDuration.Observe(time.Since(start).Seconds())
			metrics.TokenOperationsTotal.WithLabelValues("validate").Inc()
		}()

		auth := r.Header.Get("Authorization")
		if auth == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing_authorization"})
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_auth_header"})
			return
		}
		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Only HMAC is supported in this scaffold
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenMalformed
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_token"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_claims"})
			return
		}
		// check blacklisted jti
		jti, _ := claims["jti"].(string)
		if jti != "" {
			if exists, _ := redisClient.Get(r.Context(), "blacklisted_tokens:"+jti).Result(); exists != "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "token_revoked"})
				return
			}
		}
		// attach claims to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims["sub"])
		ctx = context.WithValue(ctx, JTIKey, jti)
		next(w, r.WithContext(ctx))
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
