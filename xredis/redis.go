package xredis

import (
	"context"
	"errors"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"

	"github.com/laixhe/gonet/protocol/gen/config/credis"
	"github.com/laixhe/gonet/xlog"
)

// RedisClient 客户端
type RedisClient struct {
	config *credis.Redis
	client redis.Cmdable
}

// Ping 判断服务是否可用
func (rc *RedisClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := rc.client.Ping(ctx).Err()
	if err != nil {
		return err
	}
	return nil
}

// Client get redis client
func (rc *RedisClient) Client() redis.Cmdable {
	return rc.client
}

// initSingle 单机
func initSingle(config *credis.Redis) redis.Cmdable {
	options := &redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       int(config.DbNum),
	}
	if config.PoolSize > 0 {
		options.PoolSize = int(config.PoolSize)
	}
	if config.MinIdleConn > 0 {
		options.MinIdleConns = int(config.MinIdleConn)
	}
	return redis.NewClient(options)
}

// initSentinel 哨兵主从
// func initSentinel(config *credis.Redis) redis.Cmdable {
// 	options := &redis.FailoverOptions{
// 		MasterName:    config.Master,
// 		SentinelAddrs: strings.Split(config.Addr, ","),
// 		DB:            int(config.DbNum),
// 		Password:      config.Password,
// 	}
// 	if config.PoolSize > 0 {
// 		options.PoolSize = int(config.PoolSize)
// 	}
// 	if config.MinIdleConn > 0 {
// 		options.MinIdleConns = int(config.MinIdleConn)
// 	}
// 	return redis.NewFailoverClient(options)
// }

// initCluster 分布式集群
func initCluster(config *credis.Redis) redis.Cmdable {
	options := &redis.ClusterOptions{
		Addrs:    strings.Split(config.Addr, ","),
		Password: config.Password,
	}
	if config.PoolSize > 0 {
		options.PoolSize = int(config.PoolSize)
	}
	if config.MinIdleConn > 0 {
		options.MinIdleConns = int(config.MinIdleConn)
	}
	return redis.NewClusterClient(options)
}

// connect 连接数据库
func connect(config *credis.Redis) (*RedisClient, error) {
	rc := &RedisClient{
		config: config,
	}
	addrs := strings.Split(config.Addr, ",")
	if len(addrs) == 1 {
		rc.client = initSingle(config) // 单机
	} else {
		rc.client = initCluster(config) // 分布式集群
	}
	err := rc.Ping()
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func Init(config *credis.Redis) (*RedisClient, error) {
	if config == nil {
		return nil, errors.New("redis config as nil")
	}
	if config.Addr == "" {
		return nil, errors.New("redis config addr as empty")
	}
	xlog.Debugf("redis config=%v", config)
	xlog.Debug("redis init...")

	rc, err := connect(config)
	if err != nil {
		return nil, err
	}

	xlog.Debug("redis init ok...")
	return rc, nil
}
