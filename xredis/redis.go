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
	c      *credis.Redis
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
func initSingle(c *credis.Redis) redis.Cmdable {
	options := &redis.Options{
		Addr:     c.Addr,
		Password: c.Password,
		DB:       int(c.DbNum),
	}
	if c.PoolSize > 0 {
		options.PoolSize = int(c.PoolSize)
	}
	if c.MinIdleConn > 0 {
		options.MinIdleConns = int(c.MinIdleConn)
	}
	return redis.NewClient(options)
}

// initSentinel 哨兵主从
//func initSentinel(c *credis.Redis) redis.Cmdable {
//	options := &redis.FailoverOptions{
//		MasterName:    c.Master,
//		SentinelAddrs: strings.Split(c.Addr, ","),
//		DB:            int(c.DbNum),
//		Password:      c.Password,
//	}
//	if c.PoolSize > 0 {
//		options.PoolSize = int(c.PoolSize)
//	}
//	if c.MinIdleConn > 0 {
//		options.MinIdleConns = int(c.MinIdleConn)
//	}
//	return redis.NewFailoverClient(options)
//}

// initCluster 分布式集群
func initCluster(c *credis.Redis) redis.Cmdable {
	options := &redis.ClusterOptions{
		Addrs:    strings.Split(c.Addr, ","),
		Password: c.Password,
	}
	if c.PoolSize > 0 {
		options.PoolSize = int(c.PoolSize)
	}
	if c.MinIdleConn > 0 {
		options.MinIdleConns = int(c.MinIdleConn)
	}
	return redis.NewClusterClient(options)
}

// connect 连接数据库
func connect(c *credis.Redis) (*RedisClient, error) {
	rc := &RedisClient{
		c: c,
	}

	addrs := strings.Split(c.Addr, ",")
	if len(addrs) == 1 {
		rc.client = initSingle(c) // 单机
	} else {
		rc.client = initCluster(c) // 分布式集群
	}
	err := rc.Ping()
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func Init(c *credis.Redis) (*RedisClient, error) {
	if c == nil {
		return nil, errors.New("redis config as nil")
	}
	if c.Addr == "" {
		return nil, errors.New("redis config addr as empty")
	}
	xlog.Debugf("redis config=%v", c)
	xlog.Debug("redis init...")

	rc, err := connect(c)
	if err != nil {
		return nil, err
	}

	xlog.Debug("redis init ok...")
	return rc, nil
}
