package cache

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(redisURL string) (*RedisCache, error) {
	opts, err := redis.ParseURL(strings.TrimSpace(redisURL))
	if err != nil {
		return nil, err
	}
	return &RedisCache{client: redis.NewClient(opts)}, nil
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	if r == nil || r.client == nil {
		return nil, ErrMiss
	}
	v, err := r.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrMiss
	}
	return v, err
}

func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if r == nil || r.client == nil || len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

// guestlistNamespaceKeyTTL bounds lifetime of monotonic counter keys (e.g. guestlist_nsver:{eventID})
// so inactive events do not leave unbounded Redis keys; refreshed on each bump.
const guestlistNamespaceKeyTTL = 30 * 24 * time.Hour

// BumpKey atomically increments a numeric cache key and extends TTL (namespace invalidation counters).
func (r *RedisCache) BumpKey(ctx context.Context, key string) (int64, error) {
	k := strings.TrimSpace(key)
	if r == nil || r.client == nil || k == "" {
		return 0, nil
	}
	pipe := r.client.TxPipeline()
	incr := pipe.Incr(ctx, k)
	pipe.Expire(ctx, k, guestlistNamespaceKeyTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// ReadIntKey reads a numeric key; missing keys are interpreted as zero.
func (r *RedisCache) ReadIntKey(ctx context.Context, key string) (int64, error) {
	if r == nil || r.client == nil || strings.TrimSpace(key) == "" {
		return 0, nil
	}
	raw, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	n, convErr := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if convErr != nil {
		return 0, nil
	}
	return n, nil
}
