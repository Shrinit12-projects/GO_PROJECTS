// internal/services/auth.go
package services

import (
	"context"
	"errors"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/models"
	"auth-service/internal/repositories"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

var (
	// exported sentinel errors for handler decisions
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// AuthService contains authentication related business logic.
type AuthService struct {
	usersRepo *repositories.UsersRepository
	redis     *redis.Client
	cfg       *config.Config
}

// NewAuthService constructs AuthService with required deps.
func NewAuthService(mongoClient *mongo.Client, redisClient *redis.Client, cfg *config.Config) *AuthService {
	return &AuthService{
		usersRepo: repositories.NewUsersRepository(mongoClient),
		redis:     redisClient,
		cfg:       cfg,
	}
}

// Register creates a new user (hashed password) and returns the new user id.
func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (string, error) {
	// Check existing
	exists, err := s.usersRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", ErrUserExists
	}
	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	user := models.User{
		Email:    req.Email,
		Password: string(hash),
		Role:     "student",
		Created:  time.Now().UTC(),
	}
	return s.usersRepo.Create(ctx, user)
}

// Login verifies credentials and issues access + refresh tokens.
// Refresh token is stored in Redis keyed by user id (simple approach).
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (map[string]string, error) {
	user, err := s.usersRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	// verify password
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return nil, ErrInvalidCredentials
	}
	// create tokens
	jti := uuid.NewString()
	accessToken, err := s.newAccessToken(user.ID, user.Role, jti)
	if err != nil {
		return nil, err
	}
	refreshToken := uuid.NewString() // opaque token
	// store refresh token in Redis: refresh_tokens:{user_id} -> refreshToken
	if err := s.redis.Set(ctx, "refresh_tokens:"+user.ID, refreshToken, s.cfg.RefreshTTL).Err(); err != nil {
		return nil, err
	}
	return map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "bearer",
		"expires_in":    s.cfg.AccessTTL.String(),
	}, nil
}

// Logout blacklists the token jti in Redis with TTL = access token remaining duration.
// For simplicity, we store with AccessTTL.
func (s *AuthService) Logout(ctx context.Context, jti string) error {
	return s.redis.Set(ctx, "blacklisted_tokens:"+jti, time.Now().UTC().String(), s.cfg.AccessTTL).Err()
}

// Refresh exchanges a refresh token for a new access token.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (map[string]string, error) {
	// find which user has this refresh token (simple scan not ideal in prod)
	// For this scaffold we assume client sends refresh token and user id; but to keep API simple
	// we scan a single keyspace per-user in production we would store refresh token keyed by value or store mapping.
	// Here we'll iterate over users collection to find matching refresh token - *not* good at scale but OK for scaffold.
	uid, err := s.usersRepo.FindUserIDByRefreshToken(ctx, s.redis, refreshToken)
	if err != nil {
		return nil, err
	}
	user, err := s.usersRepo.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	jti := uuid.NewString()
	accessToken, err := s.newAccessToken(user.ID, user.Role, jti)
	if err != nil {
		return nil, err
	}
	// rotate refresh token
	newRefresh := uuid.NewString()
	if err := s.redis.Set(ctx, "refresh_tokens:"+user.ID, newRefresh, s.cfg.RefreshTTL).Err(); err != nil {
		return nil, err
	}
	return map[string]string{
		"access_token":  accessToken,
		"refresh_token": newRefresh,
		"token_type":    "bearer",
		"expires_in":    s.cfg.AccessTTL.String(),
	}, nil
}

// internal: create signed JWT
func (s *AuthService) newAccessToken(userID, role, jti string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"jti":  jti,
		"exp":  time.Now().Add(s.cfg.AccessTTL).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
