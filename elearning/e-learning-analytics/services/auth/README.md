# Auth Service

JWT-based authentication microservice for e-learning platform.

## Architecture

```
auth-service/
├── main.go                     # Entry point, server lifecycle
├── internal/
│   ├── config/
│   │   └── config.go          # Environment configuration
│   ├── handler/
│   │   └── auth.go            # HTTP handlers
│   ├── middleware/
│   │   ├── jwt.go             # JWT validation
│   │   └── logging.go         # Request logging
│   ├── models/
│   │   └── user.go            # Data structures
│   ├── repositories/
│   │   ├── mongo.go           # MongoDB operations
│   │   ├── redis.go           # Redis client
│   │   └── user_repo.go       # Repository interfaces
│   ├── server/
│   │   └── server.go          # HTTP server setup
│   └── services/
│       └── auth.go            # Business logic
├── go.mod                     # Dependencies
└── dockerfile                 # Container build
```

## File Functions

### main.go
**Purpose**: Application bootstrap and graceful shutdown
- `main()` - Loads config, connects to databases, starts server
- **Uses**: `config.NewConfigFromEnv()`, `repositories.NewMongoClient()`, `repositories.NewRedisClient()`, `server.NewServer()`
- **Helps**: Initialize infrastructure and handle shutdown signals

### internal/config/config.go
**Purpose**: Environment-based configuration management
- `NewConfigFromEnv()` - Reads environment variables, validates JWT_SECRET
- `getEnv()` - Gets string env var with fallback
- `getEnvInt()` - Gets integer env var with fallback
- **Helps**: Centralize configuration, enforce required settings

### internal/handler/auth.go
**Purpose**: HTTP request/response handling
- `NewAuthHandler()` - Creates handler with dependencies
- `Register()` - POST /auth/register endpoint
- `Login()` - POST /auth/login endpoint  
- `Logout()` - DELETE /auth/logout endpoint
- `Refresh()` - POST /auth/refresh endpoint
- `HealthHandler()` - GET /health endpoint
- **Uses**: `services.AuthService` methods, `models` validation
- **Helps**: Convert HTTP requests to business operations, handle errors

### internal/middleware/jwt.go
**Purpose**: JWT token validation and context injection
- `RequireAuth()` - Validates Bearer tokens, checks Redis blacklist
- **Uses**: `jwt.Parse()`, `redis.Get()` for blacklist check
- **Helps**: Protect endpoints, inject user context for handlers

### internal/middleware/logging.go
**Purpose**: HTTP request logging
- `RequestLogger()` - Logs method, path, duration
- **Helps**: Monitor API usage, debug requests

### internal/models/user.go
**Purpose**: Data structures and validation
- `User` - Database user model
- `RegisterRequest` - Registration payload with validation
- `LoginRequest` - Login payload with validation
- `Validate()` methods - Input validation using struct tags
- **Uses**: `validator.New()` (package-level instance)
- **Helps**: Ensure data integrity, standardize API contracts

### internal/repositories/mongo.go
**Purpose**: MongoDB database operations
- `NewMongoClient()` - Creates configured MongoDB connection
- `NewUsersRepository()` - Repository with email index
- `Create()` - Insert new user
- `ExistsByEmail()` - Check email uniqueness
- `FindByEmail()` - User lookup for login
- `FindByID()` - User lookup by ID
- `FindUserIDByRefreshToken()` - Token-to-user mapping
- **Uses**: `mongo-driver` for database operations
- **Helps**: Abstract database operations, ensure data consistency

### internal/repositories/redis.go
**Purpose**: Redis cache operations
- `NewRedisClient()` - Creates configured Redis connection with ping test
- **Uses**: `redis.NewClient()`, connection pooling
- **Helps**: Manage refresh tokens and blacklisted JWTs

### internal/services/auth.go
**Purpose**: Authentication business logic
- `NewAuthService()` - Service with repository dependencies
- `Register()` - Create user with hashed password
- `Login()` - Verify credentials, issue JWT + refresh token
- `Logout()` - Blacklist JWT in Redis
- `Refresh()` - Exchange refresh token for new JWT
- `newAccessToken()` - Generate signed JWT with claims
- **Uses**: `bcrypt` for passwords, `jwt` for tokens, `uuid` for IDs
- **Helps**: Implement auth workflows, manage token lifecycle

### internal/server/server.go
**Purpose**: HTTP server configuration and routing
- `NewServer()` - Configure server with middleware and routes
- `Start()` - Start HTTP server (blocking)
- `Shutdown()` - Graceful shutdown with context
- `registerRoutes()` - Wire endpoints to handlers
- **Uses**: `gorilla/mux` for routing, middleware chain
- **Helps**: Organize HTTP layer, enable graceful operations

## Data Flow

1. **Registration**: `handler.Register()` → `services.Register()` → `repositories.Create()`
2. **Login**: `handler.Login()` → `services.Login()` → `repositories.FindByEmail()` + Redis token storage
3. **Protected Routes**: `middleware.RequireAuth()` → JWT validation + Redis blacklist check
4. **Logout**: `handler.Logout()` → `services.Logout()` → Redis blacklist storage
5. **Token Refresh**: `handler.Refresh()` → `services.Refresh()` → Redis lookup + new JWT

## Dependencies

- **JWT**: `golang-jwt/jwt/v5` - Token generation/validation
- **Database**: `mongo-driver` - User persistence
- **Cache**: `redis/go-redis/v9` - Token management
- **Validation**: `go-playground/validator/v10` - Input validation
- **Routing**: `gorilla/mux` - HTTP routing
- **Security**: `golang.org/x/crypto/bcrypt` - Password hashing