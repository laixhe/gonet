package redis

import (
	"context"
	"errors"
	"strings"
	"time"

	redisv9 "github.com/redis/go-redis/v9"
)

// Redis配置
type Config struct {
	// 连接地址(多个地址是以 , 分割)
	Addr string `json:"addr,omitempty" mapstructure:"addr" toml:"addr" yaml:"addr"`
	// 选择N号数据库
	DbNum int `json:"db_num,omitempty" mapstructure:"db_num" toml:"db_num" yaml:"db_num"`
	// 设置打开数据库连接的最大数量
	Password string `json:"password,omitempty" mapstructure:"password" toml:"password" yaml:"password"`
	// 最大链接数
	PoolSize int `json:"pool_size,omitempty" mapstructure:"pool_size" toml:"pool_size" yaml:"pool_size"`
	// 空闲链接数
	MinIdleConn int `json:"min_idle_conn,omitempty" mapstructure:"min_idle_conn" toml:"min_idle_conn" yaml:"min_idle_conn"`
}

/*
redis:
  # 连接地址(多个地址是以 , 分割)
  addr: 127.0.0.1:6379
  # 选择N号数据库
  db_num: 0
  # 密码
  password:
  # 最大链接数
  pool_size: 100
  # 空闲链接数
  min_idle_conn: 5
*/

// RedisClient 客户端
type RedisClient struct {
	config *Config
	client redisv9.Cmdable
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
func (rc *RedisClient) Client() redisv9.Cmdable {
	return rc.client
}

// initSingle 单机
func initSingle(config *Config) redisv9.Cmdable {
	options := &redisv9.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DbNum,
	}
	if config.PoolSize > 0 {
		options.PoolSize = config.PoolSize
	}
	if config.MinIdleConn > 0 {
		options.MinIdleConns = config.MinIdleConn
	}
	return redisv9.NewClient(options)
}

// initSentinel 哨兵主从
// func initSentinel(config *Config) redisv9.Cmdable {
// 	options := &redisv9.FailoverOptions{
// 		MasterName:    config.Master,
// 		SentinelAddrs: strings.Split(config.Addr, ","),
// 		DB:            config.DbNum,
// 		Password:      config.Password,
// 	}
// 	if config.PoolSize > 0 {
// 		options.PoolSize = config.PoolSize
// 	}
// 	if config.MinIdleConn > 0 {
// 		options.MinIdleConns = config.MinIdleConn
// 	}
// 	return redisv9.NewFailoverClient(options)
// }

// initCluster 分布式集群
func initCluster(config *Config) redisv9.Cmdable {
	options := &redisv9.ClusterOptions{
		Addrs:    strings.Split(config.Addr, ","),
		Password: config.Password,
	}
	if config.PoolSize > 0 {
		options.PoolSize = config.PoolSize
	}
	if config.MinIdleConn > 0 {
		options.MinIdleConns = config.MinIdleConn
	}
	return redisv9.NewClusterClient(options)
}

// connect 连接数据库
func connect(config *Config) (*RedisClient, error) {
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

func Init(config *Config) (*RedisClient, error) {
	if config == nil {
		return nil, errors.New("没有Redis配置")
	}
	if config.Addr == "" {
		return nil, errors.New("没有Redis连接地址配置")
	}
	rc, err := connect(config)
	if err != nil {
		return nil, err
	}
	return rc, nil
}
