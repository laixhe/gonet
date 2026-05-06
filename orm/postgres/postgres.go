package postgres

import (
	"context"
	"errors"

	postgresDriver "gorm.io/driver/postgres"
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

// GetById 以 id 获取数据
// model 指针传递的结构(表结构)
func (c *Client) GetById(ctx context.Context, model any, id int) error {
	return c.client.WithContext(ctx).Where("id", id).Take(model).Error
}

// GetByField 获取以对应字段条件的数据
// model 指针传递的结构(表结构)
// key   要查询的字段名
// value 要查询的字段名的值
func (c *Client) GetByField(ctx context.Context, model any, key string, value any) error {
	return c.client.WithContext(ctx).Where(key, value).Take(model).Error
}

// LastByField 获取以对应字段条件的数据(最后一条)
func (c *Client) LastByField(ctx context.Context, model any, key string, value any) error {
	return c.client.WithContext(ctx).Where(key, value).Last(model).Error
}

// FirstByField 获取以对应字段条件的数据(第一条)
func (c *Client) FirstByField(ctx context.Context, model any, key string, value any) error {
	return c.client.WithContext(ctx).Where(key, value).First(model).Error
}

// Save 会保存所有的字段，即使字段是零值
// model 指针传递的结构(表结构)
func (c *Client) Save(ctx context.Context, model any) error {
	return c.client.WithContext(ctx).Save(model).Error
}

// Create 创建数据
// model 指针传递的结构或者数组结构(表结构)
func (c *Client) Create(ctx context.Context, model any) error {
	return c.client.WithContext(ctx).Create(model).Error
}

// Delete 删除数据
// model 指针传递的结构或者数组结构(表结构)（必须包含 id 字段并赋值）
func (c *Client) Delete(ctx context.Context, model any) error {
	return c.client.WithContext(ctx).Delete(model).Error
}

// DeleteById 以 id 删除数据
// model 指针传递的结构(表结构)
func (c *Client) DeleteById(ctx context.Context, model any, id int) error {
	return c.client.WithContext(ctx).Where("id", id).Delete(model).Error
}

// Updates 修改数据
// model  指针传递的结构(表结构)或表名
// where  查询数据(表对应的字段)
// update 修改的数据(表对应的字段)
func (c *Client) Updates(ctx context.Context, model any, where map[string]any, update map[string]any) error {
	return c.client.WithContext(ctx).Model(model).Where(where).Updates(update).Error
}

// UpdatesById 以 id 修改数据
// model  指针传递的结构(表结构)
// update 修改的数据(表对应的字段)
func (c *Client) UpdatesById(ctx context.Context, model any, id int, update map[string]any) error {
	return c.client.WithContext(ctx).Model(model).Where("id", id).Updates(update).Error
}

// Init 初始化数据库
func Init(config *orm.Config, writer logger.Writer, requestId string) (orm.Client, error) {
	if err := config.Check(); err != nil {
		return nil, err
	}
	if config.Driver != orm.DriverPostgresql {
		return nil, errors.New("数据库驱动只支持 postgresql")
	}
	// 自定义配置
	opt := &gorm.Config{
		Logger: config.SetLog(writer, requestId), // 日志等级
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`，不是 `users`
		},
	}
	// 连接数据库
	client, err := gorm.Open(postgresDriver.Open(config.Dsn), opt)
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
