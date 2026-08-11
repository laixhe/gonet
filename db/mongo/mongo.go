package mongo

import (
	"context"
	"errors"
	"sync"
	"time"

	mongov2 "go.mongodb.org/mongo-driver/v2/mongo"
	optionsv2 "go.mongodb.org/mongo-driver/v2/mongo/options"
	readprefv2 "go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

/*
mongodb:
  # 连接地址
  uri: mongodb://127.0.0.1:27017
  # 指定数据库
  database: "test"
  # 最大连接的数量
  max_pool_size: 100
  # 最小连接的数量
  min_pool_size: 5
  # 最大连接的空闲时间(设置了连接可复用的最大时间)(单位秒)
  max_conn_idle_time: 300
*/

type Config struct {
	// 连接地址
	Uri string `json:"uri,omitempty" mapstructure:"uri" toml:"uri" yaml:"uri"`
	// 指定数据库
	Database string `json:"database,omitempty" mapstructure:"database" toml:"database" yaml:"database"`
	// 最大连接的数量
	MaxPoolSize uint64 `json:"max_pool_size,omitempty" mapstructure:"max_pool_size" toml:"max_pool_size" yaml:"max_pool_size"`
	// 最小连接的数量
	MinPoolSize uint64 `json:"min_pool_size,omitempty" mapstructure:"min_pool_size" toml:"min_pool_size" yaml:"min_pool_size"`
	// 最大连接的空闲时间(设置了连接可复用的最大时间)(单位秒)
	MaxConnIdleTime int64 `json:"max_conn_idle_time,omitempty" mapstructure:"max_conn_idle_time" toml:"max_conn_idle_time" yaml:"max_conn_idle_time"`
}

// Check 校验配置有效性，任一必填字段为空则返回错误。
func (c *Config) Check() error {
	if c == nil {
		return errors.New("没有Mongo配置")
	}
	if c.Uri == "" {
		return errors.New("没有Mongo连接地址配置")
	}
	if c.Database == "" {
		return errors.New("没有Mongo指定数据库配置")
	}
	return nil
}

// MClient MongoDB 客户端，所有导出方法均为并发安全。
type MClient struct {
	mu              sync.RWMutex
	config          *Config
	client          *mongov2.Client
	defaultDatabase *mongov2.Database            // 默认指定的数据库
	databaseMap     map[string]*mongov2.Database // 选择其他指定的数据库
}

// Ping 通过向 MongoDB 发送 ping 命令判断服务是否可用，可用于健康检查。
func (mc *MClient) Ping(ctx context.Context) error {
	return mc.client.Ping(ctx, readprefv2.Primary())
}

// Close 关闭与 MongoDB 的连接，释放连接池资源。
// 调用后 MClient 不可再使用。
func (mc *MClient) Close(ctx context.Context) error {
	return mc.client.Disconnect(ctx)
}

// Client 返回底层的 mongo.Client，供需要直接操作 driver 的高级场景使用。
func (mc *MClient) Client() *mongov2.Client {
	return mc.client
}

// Database 返回指定名称的数据库实例，结果会被缓存以便复用。
// 该方法并发安全。
func (mc *MClient) Database(name string) *mongov2.Database {
	mc.mu.RLock()
	db, ok := mc.databaseMap[name]
	mc.mu.RUnlock()
	if ok {
		return db
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()
	// 双重检查，避免重复创建
	if db, ok = mc.databaseMap[name]; ok {
		return db
	}
	database := mc.client.Database(name)
	mc.databaseMap[name] = database
	return database
}

// Collection 返回默认数据库下指定名称的集合（表），供 CRUD 操作使用。
func (mc *MClient) Collection(name string) *mongov2.Collection {
	return mc.defaultDatabase.Collection(name)
}

// connect 连接数据库
func connect(config *Config) (*MClient, error) {
	opts := optionsv2.Client()
	opts.ApplyURI(config.Uri)
	// 默认连接超时 10 秒，避免无响应时长时间阻塞
	opts.SetConnectTimeout(10 * time.Second)
	opts.SetServerSelectionTimeout(10 * time.Second)
	// 进行配置
	if config.MaxPoolSize > 0 {
		opts.SetMaxPoolSize(config.MaxPoolSize)
	}
	if config.MinPoolSize > 0 {
		opts.SetMinPoolSize(config.MinPoolSize)
	}
	if config.MaxConnIdleTime > 0 {
		// 最大连接的空闲时间(设置了连接可复用的最大时间)(单位秒)
		opts.SetMaxConnIdleTime(time.Duration(config.MaxConnIdleTime) * time.Second)
	}
	// 链接 mongo 服务
	client, err := mongov2.Connect(opts)
	if err != nil {
		return nil, err
	}
	// 判断服务是否可用
	err = client.Ping(context.Background(), readprefv2.Primary())
	if err != nil {
		return nil, err
	}
	return &MClient{
		config:          config,
		client:          client,
		defaultDatabase: client.Database(config.Database),
		databaseMap:     make(map[string]*mongov2.Database),
	}, nil
}

// Init 初始化数据库
func Init(config *Config) (*MClient, error) {
	if err := config.Check(); err != nil {
		return nil, err
	}
	mc, err := connect(config)
	if err != nil {
		return nil, err
	}
	return mc, nil
}
