// Package cache is the Redis adapter: client construction, health checking
// and a fixed-window rate limiter. Hot telemetry caching joins in Phase 5.
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/config"
)

// New builds the Redis client and verifies connectivity once at boot.
func New(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("redis connect: %w", err)
	}
	return client, nil
}

// HealthChecker implements application/health.Checker for Redis.
type HealthChecker struct{ client *redis.Client }

func NewHealthChecker(client *redis.Client) *HealthChecker { return &HealthChecker{client: client} }

func (h *HealthChecker) Name() string { return "redis" }

func (h *HealthChecker) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return h.client.Ping(ctx).Err()
}

// RateLimiter is a fixed-window counter limiter (INCR + EXPIRE). Simple,
// O(1), good enough for auth brute-force protection; a sliding-window or
// token-bucket variant can swap in behind the same method.
type RateLimiter struct{ client *redis.Client }

func NewRateLimiter(client *redis.Client) *RateLimiter { return &RateLimiter{client: client} }

// Allow reports whether the action identified by key may proceed under
// `limit` events per `window`.
func (l *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	pipe := l.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}
	return incr.Val() <= int64(limit), nil
}
