package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"os"
	"student/internal/conf"
	"testing"
)

// 创建用于测试的 Redis 配置
func getTestConf() *conf.Data {
	return &conf.Data{
		Redis: &conf.Data_Redis{
			Addr:     "localhost:6379", // 替换为你的 Redis 地址
			Password: "",               // 如果有密码，填写你的密码
			Db:       0,                // 使用的数据库
			//DialTimeout:  time.Second * 5,   // 连接超时时间
			//ReadTimeout:  time.Second * 3,   // 读取超时时间
			//WriteTimeout: time.Second * 3,   // 写入超时时间
		},
	}
}

func TestRedisConnection(t *testing.T) {
	// 加载测试配置
	confData := getTestConf()

	// 创建 logger
	logger := log.NewStdLogger(os.Stdout)

	// 调用 NewRdb 来初始化 Redis 客户端
	redisClient := NewRdb(confData, logger)

	// 检查 Redis 客户端是否成功连接
	err := redisClient.Ping(context.Background()).Err()
	assert.NoError(t, err, "Redis should connect without error")
}
