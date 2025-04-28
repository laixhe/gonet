package xmongo

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/laixhe/gonet/protocol/gen/config/cmongodb"
	"github.com/laixhe/gonet/xlog"
)

// MongoClient 客户端
type MongoClient struct {
	config          *cmongodb.MongoDB
	client          *mongo.Client
	defaultDatabase *mongo.Database // 默认指定的数据库
	databaseMap     *sync.Map       // 选择其他指定的数据库
}

// Ping 判断服务是否可用
func (mc *MongoClient) Ping() error {
	return mc.client.Ping(context.Background(), readpref.Primary())
}

// Client get mongo client
func (mc *MongoClient) Client() *mongo.Client {
	return mc.client
}

// Database 指定数据库
func (mc *MongoClient) Database(name string, opts ...*options.DatabaseOptions) *mongo.Database {
	loadDatabase, ok := mc.databaseMap.Load(name)
	if ok {
		return loadDatabase.(*mongo.Database)
	}
	database := mc.client.Database(name)
	mc.databaseMap.Store(name, database)
	return database
}

// Collection 选择集合(表)
func (mc *MongoClient) Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection {
	return mc.defaultDatabase.Collection(name, opts...)
}

// connect 连接数据库
func connect(config *cmongodb.MongoDB) (*MongoClient, error) {
	opts := options.Client()
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
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	// 判断服务是否可用
	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		return nil, err
	}

	return &MongoClient{
		config:          config,
		client:          client,
		defaultDatabase: client.Database(config.Database),
		databaseMap:     &sync.Map{},
	}, nil
}

// Init 初始化数据库
func Init(config *cmongodb.MongoDB) (*MongoClient, error) {
	if config == nil {
		return nil, errors.New("mongo config as nil")
	}
	if config.Uri == "" {
		return nil, errors.New("mongo config uri as empty")
	}
	xlog.Debugf("mongo config=%v", config)
	xlog.Debug("mongo init...")

	mc, err := connect(config)
	if err != nil {
		return nil, err
	}

	xlog.Debug("mongo init ok...")
	return mc, nil
}
