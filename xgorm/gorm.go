package xgorm

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/laixhe/gonet/protocol/gen/config/cgorm"
	"github.com/laixhe/gonet/protocol/gen/config/clog"
	"github.com/laixhe/gonet/xgin"
	"github.com/laixhe/gonet/xgin/constant"
	"github.com/laixhe/gonet/xlog"
)

// GormClient 客户端
type GormClient struct {
	config *cgorm.Gorm
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

func (gc *GormClient) WithGinContext(c *gin.Context) *gorm.DB {
	ctx := context.WithValue(c.Request.Context(), constant.HeaderRequestID, xgin.GetRequestID(c))
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
func connect(config *cgorm.Gorm) (*GormClient, error) {
	logLevel := logger.Info
	if xlog.GetLevel() == clog.LevelType_warn.String() {
		logLevel = logger.Warn
	}
	if xlog.GetLevel() == clog.LevelType_error.String() {
		logLevel = logger.Error
	}
	defaultLogger := NewDBlogger(NewDBWriter(), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logLevel,
		Colorful:      true,
	})

	var client *gorm.DB
	var err error
	if config.Driver == "mysql" {
		client, err = gorm.Open(mysql.Open(config.Dsn), &gorm.Config{
			Logger: defaultLogger,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
			},
		})
		if err != nil {
			return nil, err
		}
	}
	if config.Driver == "postgresql" {
		client, err = gorm.Open(postgres.Open(config.Dsn), &gorm.Config{
			Logger: defaultLogger,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
			},
		})
		if err != nil {
			return nil, err
		}
	}
	if config.Driver == "sqlite" {
		client, err = gorm.Open(sqlite.Open(config.Dsn), &gorm.Config{
			Logger: defaultLogger,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
			},
		})
		if err != nil {
			return nil, err
		}
	}
	if client == nil {
		return nil, errors.New("gorm config driver It can only be mysql postgresql sqlite")
	}

	sqlDB, err := client.DB()
	if err != nil {
		return nil, err
	}
	if config.MaxIdleCount > 0 {
		// 设置空闲连接池中连接的最大数量
		sqlDB.SetMaxIdleConns(int(config.MaxIdleCount))
	}
	if config.MaxOpenCount > 0 {
		// 设置打开数据库连接的最大数量
		sqlDB.SetMaxOpenConns(int(config.MaxOpenCount))
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
func Init(config *cgorm.Gorm) (*GormClient, error) {
	if config == nil {
		return nil, errors.New("gorm config as nil")
	}
	if config.Driver == "" {
		return nil, errors.New("gorm config driver as nil")
	}
	if config.Dsn == "" {
		return nil, errors.New("gorm config dsn as nil")
	}
	if !(config.Driver == "mysql" || config.Driver == "postgresql" || config.Driver == "sqlite") {
		return nil, errors.New("gorm config driver It can only be mysql postgresql sqlite")
	}
	xlog.Debugf("gorm config=%v", config)
	xlog.Debug("gorm init...")

	gc, err := connect(config)
	if err != nil {
		return nil, err
	}

	xlog.Debug("gorm init ok...")
	return gc, nil
}
