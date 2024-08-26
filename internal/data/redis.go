package data

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisClient struct {
	rdb *redis.Client
	log *log.Helper
}

type Msg struct {
	Topic     string
	Body      []byte
	Partition int
}

type Handler func(msg *Msg) error

func NewRedisClient(rdb *redis.Client, logger log.Logger) *RedisClient {
	return &RedisClient{
		rdb: rdb,
		log: log.NewHelper(logger),
	}
}

type RedisClientRepo interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	PushMsg(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	PopMsg(ctx context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd
	DeleteMessage(ctx context.Context, key, value string) *redis.IntCmd
	SendMsg(ctx context.Context, msg *Msg) error
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

func (r *RedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return r.rdb.Del(ctx, keys...)
}

func (r *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return r.rdb.RPush(ctx, key, values...)
}

func (r *RedisClient) PushMsg(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return r.rdb.LPush(ctx, key, values...)
}

func (r *RedisClient) PopMsg(ctx context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd {
	return r.rdb.BRPop(ctx, timeout, keys...)
}

func (r *RedisClient) DeleteMessage(ctx context.Context, key, value string) *redis.IntCmd {
	return r.rdb.LRem(ctx, key, 1, value)
}

func (r *RedisClient) SendMsg(ctx context.Context, msg *Msg) error {
	topicPartition := fmt.Sprintf("%s:%d", msg.Topic, msg.Partition)
	return r.PushMsg(ctx, topicPartition, msg.Body).Err()
}
