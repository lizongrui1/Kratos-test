package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"student/internal/biz"
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

func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	return val, err
}

func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

func (r *StudentRepo) GetStuById(ctx context.Context, id int32) (*biz.Student, error) {
	var student *biz.Student
	r.log.WithContext(ctx).Infof("biz.GetStuById: %d", id)
	cacheKey := "student:" + fmt.Sprint(id)
	val, err := r.rdb.Get(ctx, cacheKey)
	if err == nil && val != "" {
		if err := json.Unmarshal([]byte(val), &student); err == nil {
			r.log.WithContext(ctx).Infof("biz.GetStuById - Cache Hit: %v", student)
			return student, nil
		}
		r.log.WithContext(ctx).Errorf("failed to unmarshal student from redis: %v", err)
	}
	// 缓存未命中或解析失败，从数据库获取数据
	student, err = biz.StudentRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	// 数据库查询成功，将结果缓存到 Redis
	data, err := json.Marshal(student)
	if err != nil {
		r.log.WithContext(ctx).Errorf("failed to marshal student to json: %v", err)
	} else {
		err = r.rdb.Set(ctx, cacheKey, data, time.Minute*5)
		if err != nil {
			r.log.WithContext(ctx).Errorf("failed to set student data to redis: %v", err)
		}
	}
	return student, nil
}
