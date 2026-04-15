package cache

import (
	"context"
	"errors"
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
