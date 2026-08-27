// Package redis 对 go-redis 的封装，支持单机和集群模式，提供统一的初始化和连接管理。
//
// 使用示例：
//
//	cfg := &redis.Config{
//		Addr:     "127.0.0.1:6379",
//		Password: "",
//		DbNum:    0,
//		PoolSize: 100,
//	}
//	rc, err := redis.Init(cfg)
//	if err != nil {
//		panic(err)
//	}
//	defer rc.Close()
//
//	// 通过 Client() 获取 go-redis 的 Cmdable 接口，调用所有标准 Redis 命令
//	rc.Client().Set(context.Background(), "key", "value", 0)
//	val, err := rc.Client().Get(context.Background(), "key").Result()
package redis

import (
	"context"
	"errors"
	"strings"
	"time"

	redisv9 "github.com/redis/go-redis/v9"
)

/*
redis:
  # 连接地址(多个地址是以 , 分割)
  addr: 127.0.0.1:6379
  # 选择N号数据库
  db_num: 0
  # 密码
  password:
  # 最大连接数
  pool_size: 100
  # 空闲连接数
  min_idle_conn: 5
  # 连接超时(秒)
  dial_timeout: 5
  # 读超时(秒)
  read_timeout: 3
  # 写超时(秒)
  write_timeout: 3
*/

// Config Redis 连接配置
type Config struct {
	// 连接地址(多个地址是以 , 分割)
	Addr string `json:"addr,omitempty" mapstructure:"addr" toml:"addr" yaml:"addr"`
	// 选择N号数据库
	DbNum int `json:"db_num,omitempty" mapstructure:"db_num" toml:"db_num" yaml:"db_num"`
	// 密码
	Password string `json:"password,omitempty" mapstructure:"password" toml:"password" yaml:"password"`
	// 最大连接数
	PoolSize int `json:"pool_size,omitempty" mapstructure:"pool_size" toml:"pool_size" yaml:"pool_size"`
	// 空闲连接数
	MinIdleConn int `json:"min_idle_conn,omitempty" mapstructure:"min_idle_conn" toml:"min_idle_conn" yaml:"min_idle_conn"`
	// 连接超时(秒)，为0时使用 go-redis 默认值(5s)
	DialTimeout int `json:"dial_timeout,omitempty" mapstructure:"dial_timeout" toml:"dial_timeout" yaml:"dial_timeout"`
	// 读超时(秒)，为0时使用 go-redis 默认值(3s)
	ReadTimeout int `json:"read_timeout,omitempty" mapstructure:"read_timeout" toml:"read_timeout" yaml:"read_timeout"`
	// 写超时(秒)，为0时使用 go-redis 默认值(3s)
	WriteTimeout int `json:"write_timeout,omitempty" mapstructure:"write_timeout" toml:"write_timeout" yaml:"write_timeout"`
}

// Check 校验配置参数合法性，对负值字段自动归零
func (c *Config) Check() error {
	if c == nil {
		return errors.New("没有Redis配置")
	}
	if c.Addr == "" {
		return errors.New("没有Redis连接地址配置")
	}
	if c.DbNum < 0 {
		c.DbNum = 0
	}
	if c.PoolSize < 0 {
		c.PoolSize = 0
	}
	if c.MinIdleConn < 0 {
		c.MinIdleConn = 0
	}
	if c.DialTimeout < 0 {
		c.DialTimeout = 0
	}
	if c.ReadTimeout < 0 {
		c.ReadTimeout = 0
	}
	if c.WriteTimeout < 0 {
		c.WriteTimeout = 0
	}
	return nil
}

// RClient Redis 客户端，封装 go-redis 连接，提供 Ping 检测和优雅关闭
type RClient struct {
	config *Config
	client redisv9.Cmdable
}

// Ping 通过发送 PING 命令检测 Redis 服务是否可用
func (rc *RClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return rc.client.Ping(ctx).Err()
}

// Client 返回 go-redis 的 Cmdable 接口，可直接调用所有标准 Redis 命令
func (rc *RClient) Client() redisv9.Cmdable {
	return rc.client
}

// Close 关闭 Redis 客户端连接，释放资源
func (rc *RClient) Close() error {
	if closer, ok := rc.client.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

// initSingle 创建单机模式的 Redis 客户端
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
	if config.DialTimeout > 0 {
		options.DialTimeout = time.Duration(config.DialTimeout) * time.Second
	}
	if config.ReadTimeout > 0 {
		options.ReadTimeout = time.Duration(config.ReadTimeout) * time.Second
	}
	if config.WriteTimeout > 0 {
		options.WriteTimeout = time.Duration(config.WriteTimeout) * time.Second
	}
	return redisv9.NewClient(options)
}

// TODO: 哨兵主从模式，待补充 MasterName 配置项后启用
// initSentinel 创建哨兵主从模式的 Redis 客户端
//
//	func initSentinel(config *Config) redisv9.Cmdable {
//		options := &redisv9.FailoverOptions{
//			MasterName:    config.MasterName,
//			SentinelAddrs: strings.Split(config.Addr, ","),
//			DB:            config.DbNum,
//			Password:      config.Password,
//		}
//		if config.PoolSize > 0 {
//			options.PoolSize = config.PoolSize
//		}
//		if config.MinIdleConn > 0 {
//			options.MinIdleConns = config.MinIdleConn
//		}
//		return redisv9.NewFailoverClient(options)
//	}

// initCluster 创建分布式集群模式的 Redis 客户端
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
	if config.DialTimeout > 0 {
		options.DialTimeout = time.Duration(config.DialTimeout) * time.Second
	}
	if config.ReadTimeout > 0 {
		options.ReadTimeout = time.Duration(config.ReadTimeout) * time.Second
	}
	if config.WriteTimeout > 0 {
		options.WriteTimeout = time.Duration(config.WriteTimeout) * time.Second
	}
	return redisv9.NewClusterClient(options)
}

// connect 根据地址数量自动选择单机或集群模式创建连接
func connect(config *Config) (*RClient, error) {
	rc := &RClient{
		config: config,
	}
	addrs := strings.Split(config.Addr, ",")
	if len(addrs) == 1 {
		rc.client = initSingle(config) // 单机模式
	} else {
		rc.client = initCluster(config) // 分布式集群模式
	}
	if err := rc.Ping(); err != nil {
		rc.Close() // Ping 失败时释放已创建的连接池资源
		return nil, err
	}
	return rc, nil
}

// Init 初始化 Redis 客户端。根据 Addr 的地址数量自动选择单机或集群模式，
// 返回 RClient 供后续操作使用。
func Init(config *Config) (*RClient, error) {
	if err := config.Check(); err != nil {
		return nil, err
	}
	rc, err := connect(config)
	if err != nil {
		return nil, err
	}
	return rc, nil
}
