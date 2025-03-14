package xgorm

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/laixhe/gonet/protocol/gen/config/cgorm"
	"github.com/laixhe/gonet/xlog"
)

// GormClient 客户端
type GormClient struct {
	c      *cgorm.Gorm
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

// Schema 模式(postgresql专用)
func (gc *GormClient) Schema() string {
	return gc.c.Schema
}

// SchemaTableName 模式表名(postgresql专用)
func (gc *GormClient) SchemaTableName(name string) string {
	return gc.c.Schema + "." + name
}

// connect 连接数据库
func connect(c *cgorm.Gorm) (*GormClient, error) {
	defaultLogger := logger.New(newWriter(log.New(os.Stdout, " ", log.LstdFlags)), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Info,
		Colorful:      true,
	})

	var client *gorm.DB
	var err error
	if c.Driver == "mysql" {
		client, err = gorm.Open(mysql.Open(c.Dsn), &gorm.Config{
			Logger: defaultLogger,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
			},
		})
		if err != nil {
			return nil, err
		}
	}
	if c.Driver == "postgresql" {
		client, err = gorm.Open(postgres.Open(c.Dsn), &gorm.Config{
			Logger: defaultLogger,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
			},
		})
		if err != nil {
			return nil, err
		}
	}
	if c.Driver == "sqlite" {
		client, err = gorm.Open(sqlite.Open(c.Dsn), &gorm.Config{
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
	if c.MaxIdleCount > 0 {
		// 设置空闲连接池中连接的最大数量
		sqlDB.SetMaxIdleConns(int(c.MaxIdleCount))
	}
	if c.MaxOpenCount > 0 {
		// 设置打开数据库连接的最大数量
		sqlDB.SetMaxOpenConns(int(c.MaxOpenCount))
	}
	if c.MaxLifeTime > 0 {
		// 设置了连接可复用的最大时间(要比数据库设置连接超时时间少)
		sqlDB.SetConnMaxLifetime(time.Duration(c.MaxLifeTime) * time.Second)
	}
	// 验证数据库连接是否正常
	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}
	return &GormClient{
		c:      c,
		client: client,
	}, nil
}

// Init 初始化数据库
func Init(c *cgorm.Gorm) (*GormClient, error) {
	if c == nil {
		return nil, errors.New("gorm config as nil")
	}
	if c.Driver == "" {
		return nil, errors.New("gorm config driver as nil")
	}
	if c.Dsn == "" {
		return nil, errors.New("gorm config dsn as nil")
	}
	if !(c.Driver == "mysql" || c.Driver == "postgresql" || c.Driver == "sqlite") {
		return nil, errors.New("gorm config driver It can only be mysql postgresql sqlite")
	}
	xlog.Debugf("gorm config=%v", c)
	xlog.Debug("gorm init...")

	gc, err := connect(c)
	if err != nil {
		return nil, err
	}

	xlog.Debug("gorm init ok...")
	return gc, nil
}
