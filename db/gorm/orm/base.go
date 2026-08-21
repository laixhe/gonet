package orm

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// BaseClient 数据库客户端通用实现
// 各数据库驱动(mysql/postgresql/sqlite)通过嵌入该结构体复用通用的 CRUD 方法,
// 避免三份驱动中维护完全相同的重复代码
type BaseClient struct {
	config *Config
	client *gorm.DB
}

// NewBaseClient 创建基础客户端
func NewBaseClient(config *Config, client *gorm.DB) *BaseClient {
	return &BaseClient{
		config: config,
		client: client,
	}
}

// Ping 判断服务是否可用
func (c *BaseClient) Ping() error {
	db, err := c.client.DB()
	if err != nil {
		return err
	}
	// 验证数据库连接是否正常
	return db.Ping()
}

// Client 获取 gorm 客户端
func (c *BaseClient) Client() *gorm.DB {
	return c.client
}

// WithContext 获取 gorm 客户端
func (c *BaseClient) WithContext(ctx context.Context) *gorm.DB {
	return c.client.WithContext(ctx)
}

// Transaction 在事务中执行回调函数
// fc 回调函数中返回错误会回滚事务, 返回 nil 提交事务
func (c *BaseClient) Transaction(ctx context.Context, fc func(tx *gorm.DB) error) error {
	return c.client.WithContext(ctx).Transaction(fc)
}

// GetById 以 id 获取数据
// model 指针传递的结构(表结构)
func (c *BaseClient) GetById(ctx context.Context, model any, id int) error {
	return c.client.WithContext(ctx).Where("id", id).Take(model).Error
}

// GetByField 获取以对应字段条件的数据
// model 指针传递的结构(表结构)
// key   要查询的字段名
// value 要查询的字段名的值
func (c *BaseClient) GetByField(ctx context.Context, model any, key string, value any) error {
	return c.client.WithContext(ctx).Where(key, value).Take(model).Error
}

// GetByWhere 获取以对应条件的数据
// model 指针传递的结构(表结构)
// where 查询数据(表对应的字段)
func (c *BaseClient) GetByWhere(ctx context.Context, model any, where map[string]any) error {
	return c.client.WithContext(ctx).Where(where).Take(model).Error
}

// LastByField 获取以对应字段条件的数据(最后一条)
func (c *BaseClient) LastByField(ctx context.Context, model any, key string, value any) error {
	return c.client.WithContext(ctx).Where(key, value).Last(model).Error
}

// FirstByField 获取以对应字段条件的数据(第一条)
func (c *BaseClient) FirstByField(ctx context.Context, model any, key string, value any) error {
	return c.client.WithContext(ctx).Where(key, value).First(model).Error
}

// Save 会保存所有的字段，即使字段是零值
// model 指针传递的结构(表结构)
func (c *BaseClient) Save(ctx context.Context, model any) error {
	return c.client.WithContext(ctx).Save(model).Error
}

// Create 创建数据
// model 指针传递的结构或者数组结构(表结构)
func (c *BaseClient) Create(ctx context.Context, model any) error {
	return c.client.WithContext(ctx).Create(model).Error
}

// Delete 删除数据
// model 指针传递的结构或者数组结构(表结构)（必须包含 id 字段并赋值）
func (c *BaseClient) Delete(ctx context.Context, model any) error {
	return c.client.WithContext(ctx).Delete(model).Error
}

// DeleteById 以 id 删除数据
// model 指针传递的结构(表结构)
func (c *BaseClient) DeleteById(ctx context.Context, model any, id int) error {
	return c.client.WithContext(ctx).Where("id", id).Delete(model).Error
}

// Updates 修改数据
// model  指针传递的结构(表结构)或表名
// where  查询数据(表对应的字段)
// update 修改的数据(表对应的字段)
// 未匹配到任何行时返回 ErrorNoUpdatedLines
func (c *BaseClient) Updates(ctx context.Context, model any, where map[string]any, update map[string]any) error {
	if len(update) == 0 {
		return errors.New("update 修改的数据不能为空")
	}
	db := c.client.WithContext(ctx).Model(model).Where(where).Updates(update)
	if db.Error != nil {
		return db.Error
	}
	// 没有匹配到任何行
	if db.RowsAffected == 0 {
		return ErrorNoUpdatedLines
	}
	return nil
}

// UpdatesById 以 id 修改数据
// model  指针传递的结构(表结构)
// update 修改的数据(表对应的字段)
// 未匹配到任何行时返回 ErrorNoUpdatedLines
func (c *BaseClient) UpdatesById(ctx context.Context, model any, id int, update map[string]any) error {
	if len(update) == 0 {
		return errors.New("update 修改的数据不能为空")
	}
	db := c.client.WithContext(ctx).Model(model).Where("id", id).Updates(update)
	if db.Error != nil {
		return db.Error
	}
	// 没有匹配到任何行
	if db.RowsAffected == 0 {
		return ErrorNoUpdatedLines
	}
	return nil
}

// Count 统计数量
// model 指针传递的结构(表结构)或表名
// where 查询数据(表对应的字段)
func (c *BaseClient) Count(ctx context.Context, model any, where map[string]any) (int64, error) {
	var count int64
	db := c.client.WithContext(ctx).Model(model).Where(where).Count(&count)
	return count, db.Error
}
