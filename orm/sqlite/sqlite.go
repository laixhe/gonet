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

// Schema 模式( PostgreSQL 专用)
func (c *Client) Schema() string {
	return ""
}

// SchemaTableName 模式表名( PostgreSQL 专用)
func (c *Client) SchemaTableName(name string) string {
	return ""
}

func (c *Client) Client() *gorm.DB {
	return c.client
}

func (c *Client) WithContext(ctx context.Context) *gorm.DB {
	return c.client.WithContext(ctx)
}

// GetById 以 id 获取数据
// data 指针传递的结构(表结构)
func (c *Client) GetById(ctx context.Context, id int, data any) error {
	return c.client.WithContext(ctx).Where("id", id).Take(data).Error
}

// GetByField 获取以对应字段条件的数据
// key   要查询的字段名
// value 要查询的字段名的值
// data 指针传递的结构(表结构)
func (c *Client) GetByField(ctx context.Context, key string, value, data any) error {
	return c.client.WithContext(ctx).Where(key, value).Take(data).Error
}

// Save 会保存所有的字段，即使字段是零值
// data 指针传递的结构(表结构)
func (c *Client) Save(ctx context.Context, data any) error {
	return c.client.WithContext(ctx).Save(data).Error
}

// Create 创建数据
// data 指针传递的结构或者数组结构(表结构)
func (c *Client) Create(ctx context.Context, data any) error {
	return c.client.WithContext(ctx).Create(data).Error
}

// Delete 删除数据
// data 指针传递的结构或者数组结构(表结构)（必须包含 id 字段并赋值）
func (c *Client) Delete(ctx context.Context, data any) error {
	return c.client.WithContext(ctx).Delete(data).Error
}

// DeleteById 以 id 删除数据
// model 指针传递的结构(表结构)
func (c *Client) DeleteById(ctx context.Context, id int, model any) error {
	return c.client.WithContext(ctx).Where("id", id).Delete(model).Error
}

// UpdatesById 以 id 修改数据
// model 指针传递的结构(表结构)
// data  修改的数据(表对应的字段)
func (c *Client) UpdatesById(ctx context.Context, id int, model any, data map[string]any) error {
	return c.client.WithContext(ctx).Model(model).Where("id", id).Updates(data).Error
}

// Init 初始化数据库
func Init(config *orm.Config, writer logger.Writer, requestId string) (orm.Client, error) {
	if err := config.Check(); err != nil {
		return nil, err
	}
	if config.Driver != orm.DriverSqlite {
		return nil, errors.New("数据库驱动只支持 sqlite")
	}
	// 自定义配置
	opt := &gorm.Config{
		Logger: config.SetLog(writer, requestId), // 日志等级
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`，不是 `users`
		},
	}
	// 连接数据库
	client, err := gorm.Open(sqliteDriver.Open(config.Dsn), opt)
	if err != nil {
		return nil, err
	}
	// 设置数据库连接
	if err = config.SetDB(client); err != nil {
		return nil, err
	}
	return &Client{
		config: config,
		client: client,
	}, nil
}
