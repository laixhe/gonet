package mongo

import (
	"context"
	"errors"
	"time"

	mongov2 "go.mongodb.org/mongo-driver/v2/mongo"
	optionsv2 "go.mongodb.org/mongo-driver/v2/mongo/options"
	readprefv2 "go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// 数据库配置
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

// MongoClient 客户端
type MongoClient struct {
	config          *Config
	client          *mongov2.Client
	defaultDatabase *mongov2.Database            // 默认指定的数据库
	databaseMap     map[string]*mongov2.Database // 选择其他指定的数据库
}

// Ping 判断服务是否可用
func (mc *MongoClient) Ping() error {
	return mc.client.Ping(context.Background(), readprefv2.Primary())
}

// Client get mongo client
func (mc *MongoClient) Client() *mongov2.Client {
	return mc.client
}

// Database 指定数据库
func (mc *MongoClient) Database(name string) *mongov2.Database {
	loadDatabase, ok := mc.databaseMap[name]
	if ok {
		return loadDatabase
	}
	database := mc.client.Database(name)
	mc.databaseMap[name] = database
	return database
}

// Collection 选择集合(表)
func (mc *MongoClient) Collection(name string) *mongov2.Collection {
	return mc.defaultDatabase.Collection(name)
}

// connect 连接数据库
func connect(config *Config) (*MongoClient, error) {
	opts := optionsv2.Client()
	opts.ApplyURI(config.Uri)
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
	return &MongoClient{
		config:          config,
		client:          client,
		defaultDatabase: client.Database(config.Database),
		databaseMap:     make(map[string]*mongov2.Database),
	}, nil
}

// Init 初始化数据库
func Init(config *Config) (*MongoClient, error) {
	if config == nil {
		return nil, errors.New("没有Mongo配置")
	}
	if config.Uri == "" {
		return nil, errors.New("没有Mongo连接地址配置")
	}
	mc, err := connect(config)
	if err != nil {
		return nil, err
	}
	return mc, nil
}
