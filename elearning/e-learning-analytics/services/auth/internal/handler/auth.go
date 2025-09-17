// internal/handler/auth.go
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/metrics"
	"auth-service/internal/middleware"
	"auth-service/internal/models"
	"auth-service/internal/services"

	"go.mongodb.org/mongo-driver/mongo"
	"github.com/redis/go-redis/v9"
)

// AuthHandler groups dependencies for authentication endpoints.
type AuthHandler struct {
	svc *services.AuthService
	cfg *config.Config
}

// NewAuthHandler instantiates an AuthHandler wired with repo clients.
func NewAuthHandler(mongoClient *mongo.Client, redisClient *redis.Client, cfg *config.Config) *AuthHandler {
	// Services take clients and config; they wrap repositories internally.
	svc := services.NewAuthService(mongoClient, redisClient, cfg)
	return &AuthHandler{svc: svc, cfg: cfg}
}

// HealthHandler is a shared health endpoint
// @Summary Health check
// @Description Check service health
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "status"
// @Router /health [get]
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "elearn-api"})
}

// Register handles POST /auth/register
// @Summary Register new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration data"
// @Success 201 {object} map[string]string "user_id"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 409 {object} map[string]interface{} "User exists"
// @Failure 500 {object} map[string]interface{} "Internal error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	// input validation
	if err := req.Validate(); err != nil {
		httpError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID, err := h.svc.Register(ctx, req)
	if err != nil {
		if errors.Is(err, services.ErrUserExists) {
			httpError(w, http.StatusConflict, "user_exists", err.Error())
			return
		}
		httpError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}

	// Record successful registration
	metrics.UserRegistrationsTotal.Inc()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"user_id": userID})
}

// Login handles POST /auth/login
// @Summary User login
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} map[string]string "tokens"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Internal error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		httpError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tokens, err := h.svc.Login(ctx, req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			metrics.LoginAttemptsTotal.WithLabelValues("failure").Inc()
			httpError(w, http.StatusUnauthorized, "invalid_credentials", err.Error())
			return
		}
		httpError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}

	// Record successful login
	metrics.LoginAttemptsTotal.WithLabelValues("success").Inc()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tokens)
}

// Logout handles DELETE /auth/logout
// @Summary User logout
// @Description Blacklist current JWT token
// @Tags auth
// @Security BearerAuth
// @Success 204 "No content"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal error"
// @Router /auth/logout [delete]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Expect middleware to attach "jti" in context (JWT middleware)
	jti, ok := r.Context().Value(middleware.JTIKey).(string)
	if !ok || jti == "" {
		httpError(w, http.StatusUnauthorized, "missing_token", "no token jti in context")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.svc.Logout(ctx, jti); err != nil {
		httpError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}

	// Record token blacklist operation
	metrics.TokenOperationsTotal.WithLabelValues("blacklist").Inc()

	w.WriteHeader(http.StatusNoContent)
}

// Refresh handles POST /auth/refresh
// @Summary Refresh JWT token
// @Description Exchange refresh token for new JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "refresh_token"
// @Success 200 {object} map[string]string "new tokens"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Invalid refresh token"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	type refreshReq struct {
		RefreshToken string `json:"refresh_token"`
	}
 // amazonq-ignore-next-line
	var req refreshReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		httpError(w, http.StatusBadRequest, "invalid_request", "refresh_token required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	toks, err := h.svc.Refresh(ctx, req.RefreshToken)
	if err != nil {
		httpError(w, http.StatusUnauthorized, "invalid_refresh", err.Error())
		return
	}

	// Record token refresh operation
	metrics.TokenOperationsTotal.WithLabelValues("refresh").Inc()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toks)
}

// httpError writes a structured JSON error response.
func httpError(w http.ResponseWriter, status int, code string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   code,
		"details": details,
		"code":    status,
	})
}
