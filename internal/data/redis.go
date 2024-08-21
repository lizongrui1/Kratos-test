package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisClient struct {
	rdb *redis.Client
	log *log.Helper
}

func NewRedisClient(rdb *redis.Client, logger log.Logger) *RedisClient {
	return &RedisClient{
		rdb: rdb,
		log: log.NewHelper(logger),
	}
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	return val, err
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.rdb.Set(ctx, key, value, expiration).Err()
}
