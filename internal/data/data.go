package data

import (
	"context"
	"gorm.io/driver/mysql"
	"student/internal/conf"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGormDB, NewStudentRepo, NewRdb)

// Data .
type Data struct {
	gormDB *gorm.DB
	rdb    *redis.Client
}

// NewGormDB 初始化 gorm
func NewGormDB(c *conf.Data) (*gorm.DB, error) {
	dsn := c.Database.Source
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetMaxOpenConns(150)
	sqlDB.SetConnMaxLifetime(time.Second * 25)
	return db, nil
}

// NewData .
func NewData(logger log.Logger, db *gorm.DB) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{gormDB: db}, cleanup, nil
}

func NewRdb(c *conf.Data, logger log.Logger) *redis.Client {
	l := log.NewHelper(log.With(logger, "module", "layout-service/data/NewRdb"))

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
