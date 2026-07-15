package orm

import (
	"context"

	"gorm.io/gorm"
)

type Client interface {
	// Ping 判断服务是否可用
	Ping() error
	// Client 获取gorm客户端
	Client() *gorm.DB
	// WithContext 获取gorm客户端
	WithContext(ctx context.Context) *gorm.DB

	// GetById 以 id 获取数据
	// model 指针传递的结构(表结构)
	GetById(ctx context.Context, model any, id int) error

	// GetByField 获取以对应字段条件的数据
	// model 指针传递的结构(表结构)
	// key   要查询的字段名
	// value 要查询的字段名的值
	GetByField(ctx context.Context, model any, key string, value any) error
	// LastByField 获取以对应字段条件的数据(最后一条)
	LastByField(ctx context.Context, model any, key string, value any) error
	// FirstByField 获取以对应字段条件的数据(第一条)
	FirstByField(ctx context.Context, model any, key string, value any) error

	// Save 会保存所有的字段，即使字段是零值
	// model 指针传递的结构(表结构)
	Save(ctx context.Context, model any) error

	// Create 创建数据
	// model 指针传递的结构或者数组结构(表结构)
	Create(ctx context.Context, model any) error

	// Delete 删除数据
	// model 指针传递的结构或者数组结构(表结构)（必须包含 id 字段并赋值）
	Delete(ctx context.Context, model any) error

	// DeleteById 以 id 删除数据
	// model 指针传递的结构(表结构)
	DeleteById(ctx context.Context, model any, id int) error

	// Updates 修改数据
	// model  指针传递的结构(表结构)或表名
	// where  查询数据(表对应的字段)
	// update 修改的数据(表对应的字段)
	Updates(ctx context.Context, model any, where map[string]any, update map[string]any) error

	// UpdatesById 以 id 修改数据
	// model  指针传递的结构(表结构)
	// update 修改的数据(表对应的字段)
	UpdatesById(ctx context.Context, model any, id int, update map[string]any) error
}
