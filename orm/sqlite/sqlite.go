package sqlite

import (
	"context"
	"errors"

	sqliteDriver "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/laixhe/gonet/orm/orm"
)

type Client struct {
	config *orm.Config
	client *gorm.DB
}

// Ping 判断服务是否可用
func (c *Client) Ping() error {
	db, err := c.client.DB()
	if err != nil {
		return err
	}
	// 验证数据库连接是否正常
	return db.Ping()
}

func (c *Client) Client() *gorm.DB {
	return c.client
}

func (c *Client) WithContext(ctx context.Context) *gorm.DB {
	return c.client.WithContext(ctx)
}

// Schema 模式(postgresql专用)
func (c *Client) Schema() string {
	return ""
}

// SchemaTableName 模式表名(postgresql专用)
func (c *Client) SchemaTableName(name string) string {
	return ""
}

// Init 初始化数据库
func Init(config *orm.Config, writer logger.Writer, requestid string) (orm.OrmClient, error) {
	if err := config.Check(); err != nil {
		return nil, err
	}
	if config.Driver != orm.DriverSqlite {
		return nil, errors.New("数据库驱动只支持 sqlite")
	}
	// 自定义配置
	opt := &gorm.Config{
		Logger: config.SetLog(writer, requestid), // 日志等级
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	}
	// 连接数据库
	client, err := gorm.Open(sqliteDriver.Open(config.Dsn), opt)
	if err != nil {
		return nil, err
	}
	// 设置数据库连接
	if err := config.SetDB(client); err != nil {
		return nil, err
	}
	return &Client{
		config: config,
		client: client,
	}, nil
}
