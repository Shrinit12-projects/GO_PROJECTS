// internal/repositories/instrumented_redis.go
package repositories

import (
	"context"
	"time"

	"auth-service/internal/metrics"

	"github.com/redis/go-redis/v9"
)

// InstrumentedRedisClient wraps redis.Client with metrics
type InstrumentedRedisClient struct {
	*redis.Client
}

// NewInstrumentedRedisClient wraps Redis client with Prometheus metrics
func NewInstrumentedRedisClient(client *redis.Client) *InstrumentedRedisClient {
	return &InstrumentedRedisClient{Client: client}
}

// Get wraps redis Get with metrics
func (c *InstrumentedRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := c.Client.Get(ctx, key)
	
	result := "hit"
	if cmd.Err() == redis.Nil {
		result = "miss"
	} else if cmd.Err() != nil {
		result = "error"
	}
	
	metrics.RedisOperationsTotal.WithLabelValues("get", result).Inc()
	return cmd
}

// Set wraps redis Set with metrics
func (c *InstrumentedRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := c.Client.Set(ctx, key, value, expiration)
	
	result := "success"
	if cmd.Err() != nil {
		result = "error"
	}
	
	metrics.RedisOperationsTotal.WithLabelValues("set", result).Inc()
	return cmd
}

// Del wraps redis Del with metrics
func (c *InstrumentedRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := c.Client.Del(ctx, keys...)
	
	result := "success"
	if cmd.Err() != nil {
		result = "error"
	}
	
	metrics.RedisOperationsTotal.WithLabelValues("del", result).Inc()
	return cmd
}