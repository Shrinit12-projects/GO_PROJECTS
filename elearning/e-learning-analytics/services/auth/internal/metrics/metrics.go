// internal/metrics/metrics.go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_service_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_service_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Auth Business Metrics
	LoginAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_service_login_attempts_total",
			Help: "Total number of login attempts",
		},
		[]string{"result"}, // success, failure
	)

	UserRegistrationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_service_user_registrations_total",
			Help: "Total number of user registrations",
		},
	)

	TokenOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_service_token_operations_total",
			Help: "Total number of token operations",
		},
		[]string{"operation"}, // refresh, blacklist, validate
	)

	JWTValidationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "auth_service_jwt_validation_duration_seconds",
			Help:    "JWT validation duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
	)

	// Infrastructure Metrics
	RedisOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_service_redis_operations_total",
			Help: "Total number of Redis operations",
		},
		[]string{"operation", "result"}, // get/set/del, hit/miss/error
	)

	MongoOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_service_mongo_operations_total",
			Help: "Total number of MongoDB operations",
		},
		[]string{"operation", "collection", "result"}, // find/insert/update, users, success/error
	)

	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_service_active_connections",
			Help: "Number of active connections",
		},
	)
)