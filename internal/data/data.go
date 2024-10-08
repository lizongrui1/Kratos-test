package data

import (
	"context"
	"student/internal/conf"
	"time"

	"gorm.io/driver/mysql"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewStudentRepo, NewRdb, NewRedisClient)

// Data .
type Data struct {
	db *gorm.DB
}

func (d *Data) DB() *gorm.DB {
	return d.db
}

// NewRdb 初始化 redis
func NewRdb(c *conf.Data, logger log.Logger) *redis.Client {
	l := log.NewHelper(log.With(logger, "module", "data/NewRdb"))

	opts := &redis.Options{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		DB:           int(c.Redis.Db),
		DialTimeout:  c.Redis.DialTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
	}

	rdb := redis.NewClient(opts)
	l.Infof("redis client created, opts = #%+v", opts)
	// Enable tracing instrumentation.
	if err := redisotel.InstrumentTracing(rdb); err != nil {
		panic(err)
	}
	// Enable metrics instrumentation.
	if err := redisotel.InstrumentMetrics(rdb); err != nil {
		panic(err)
	}
	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		l.Errorf("failed to ping redis: %w", err)
		return nil
	}
	//if err := rdb.Close(); err != nil {
	//	l.Errorf("failed to close connection to redis: %v", err)
	//}

	return rdb
}

// NewData 初始化 Data
func NewData(logger log.Logger, c *conf.Data) (*Data, func(), error) {
	dsn := c.Database.Source
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetMaxOpenConns(150)
	sqlDB.SetConnMaxLifetime(time.Second * 25)

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		sqlDB, _ := db.DB()
		if err := sqlDB.Close(); err != nil {
			log.NewHelper(logger).Errorf("failed to close db: %v", err)
		}
	}
	return &Data{db: db}, cleanup, nil
}
