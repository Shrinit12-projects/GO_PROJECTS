// internal/repositories/redis.go
package repositories

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient returns a connected Redis client with a ping check.
func NewRedisClient(ctx context.Context, addr, password string) (*redis.Client, error) {
	opt := &redis.Options{
		Addr:         addr,
		Password:     password,
		PoolSize:     50,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
	client := redis.NewClient(opt)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}
