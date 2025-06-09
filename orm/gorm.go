package orm

import (
	"context"
	"errors"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// 数据库驱动名称
const (
	DriverMysql      = "mysql"
	DriverPostgresql = "postgresql"
	DriverSqlite     = "sqlite"
)

// 数据库配置
type Config struct {
	// 驱动名称 mysql postgresql sqlite
	Driver string `json:"driver" mapstructure:"driver" toml:"driver" yaml:"driver"`
	// 连接地址
	Dsn string `json:"dsn" mapstructure:"dsn" toml:"dsn" yaml:"dsn"`
	// 模式(postgresql专用)
	Schema string `json:"schema" mapstructure:"schema" toml:"schema" yaml:"schema"`
	// 设置空闲连接池中连接的最大数量
	MaxIdleCount int `json:"max_idle_count" mapstructure:"max_idle_count" toml:"max_idle_count" yaml:"max_idle_count"`
	// 设置打开数据库连接的最大数量
	MaxOpenCount int `json:"max_open_count" mapstructure:"max_open_count" toml:"max_open_count" yaml:"max_open_count"`
	// 设置了连接可复用的最大时间(要比数据库设置连接超时时间少)(单位秒)
	MaxLifeTime int `json:"max_life_time" mapstructure:"max_life_time" toml:"max_life_time" yaml:"max_life_time"`
	// 设置日志级别 1=Silent 2=Error 3=Warn 4=Info (默认 4)
	LogLevel logger.LogLevel `json:"log_level" mapstructure:"log_level" toml:"log_level" yaml:"log_level"`
}

/*
orm:
  # 驱动名称 mysql postgresql sqlite
  driver: mysql
  # 连接地址
  dsn: root:123456@tcp(127.0.0.1:3306)/webapi?charset=utf8mb4&parseTime=True&loc=Local
  # 设置空闲连接池中连接的最大数量
  max_idle_count: 10
  # 设置打开数据库连接的最大数量
  max_open_count: 100
  # 设置了连接可复用的最大时间(要比数据库设置连接超时时间少)(单位秒)
  max_life_time: 300
  # 设置日志级别 1=Silent 2=Error 3=Warn 4=Info
  log_level: 4
*/

// GormClient 客户端
type GormClient struct {
	config *Config
	client *gorm.DB
}

// Ping 判断服务是否可用
func (gc *GormClient) Ping() error {
	sqlDB, err := gc.client.DB()
	if err != nil {
		return err
	}
	// 验证数据库连接是否正常
	return sqlDB.Ping()
}

// Client get gorm client
func (gc *GormClient) Client() *gorm.DB {
	return gc.client
}

func (gc *GormClient) WithContext(ctx context.Context) *gorm.DB {
	return gc.client.WithContext(ctx)
}

// Schema 模式(postgresql专用)
func (gc *GormClient) Schema() string {
	return gc.config.Schema
}

// SchemaTableName 模式表名(postgresql专用)
func (gc *GormClient) SchemaTableName(name string) string {
	return gc.config.Schema + "." + name
}

// connect 连接数据库
func connect(config *Config, log logger.Interface) (*GormClient, error) {
	opt := &gorm.Config{
		Logger: log,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	}
	var client *gorm.DB
	var err error
	if config.Driver == DriverMysql {
		client, err = gorm.Open(mysql.Open(config.Dsn), opt)
		if err != nil {
			return nil, err
		}
	}
	if config.Driver == DriverPostgresql {
		client, err = gorm.Open(postgres.Open(config.Dsn), opt)
		if err != nil {
			return nil, err
		}
	}
	if config.Driver == DriverSqlite {
		client, err = gorm.Open(sqlite.Open(config.Dsn), opt)
		if err != nil {
			return nil, err
		}
	}
	sqlDB, err := client.DB()
	if err != nil {
		return nil, err
	}
	if config.MaxIdleCount > 0 {
		// 设置空闲连接池中连接的最大数量
		sqlDB.SetMaxIdleConns(config.MaxIdleCount)
	}
	if config.MaxOpenCount > 0 {
		// 设置打开数据库连接的最大数量
		sqlDB.SetMaxOpenConns(config.MaxOpenCount)
	}
	if config.MaxLifeTime > 0 {
		// 设置了连接可复用的最大时间(要比数据库设置连接超时时间少)
		sqlDB.SetConnMaxLifetime(time.Duration(config.MaxLifeTime) * time.Second)
	}
	// 验证数据库连接是否正常
	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}
	return &GormClient{
		config: config,
		client: client,
	}, nil
}

// Init 初始化数据库
func Init(config *Config, log logger.Interface) (*GormClient, error) {
	if config == nil {
		return nil, errors.New("没有数据库配置")
	}
	if config.Driver == "" {
		return nil, errors.New("没有数据库驱动名称配置")
	}
	if config.Dsn == "" {
		return nil, errors.New("没有数据库连接地址配置")
	}
	if !(config.Driver == DriverMysql || config.Driver == DriverPostgresql || config.Driver == DriverSqlite) {
		return nil, errors.New("数据库驱动名称支持 mysql postgresql sqlite 三种")
	}
	// 连接数据库
	gc, err := connect(config, log)
	if err != nil {
		return nil, err
	}
	return gc, nil
}
