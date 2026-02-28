package orm

import (
	"context"

	"gorm.io/gorm"
)

type Client interface {
	// Ping 判断服务是否可用
	Ping() error
	// Schema 模式( PostgreSQL 专用)
	Schema() string
	// SchemaTableName 模式表名( PostgreSQL 专用)
	SchemaTableName(name string) string
	// Client 获取gorm客户端
	Client() *gorm.DB
	// WithContext 获取gorm客户端
	WithContext(ctx context.Context) *gorm.DB

	// GetById 以 id 获取数据
	// data 指针传递的结构(表结构)
	GetById(ctx context.Context, id int, data any) error

	// GetByField 获取以对应字段条件的数据
	// key   要查询的字段名
	// value 要查询的字段名的值
	// data 指针传递的结构(表结构)
	GetByField(ctx context.Context, key string, value, data any) error

	// Save 会保存所有的字段，即使字段是零值
	// data 指针传递的结构(表结构)
	Save(ctx context.Context, data any) error

	// Create 创建数据
	// data 指针传递的结构或者数组结构(表结构)
	Create(ctx context.Context, data any) error

	// Delete 删除数据
	// data 指针传递的结构或者数组结构(表结构)（必须包含 id 字段并赋值）
	Delete(ctx context.Context, data any) error

	// DeleteById 以 id 删除数据
	// model 指针传递的结构(表结构)
	DeleteById(ctx context.Context, id int, model any) error

	// UpdatesById 以 id 修改数据
	// model 指针传递的结构(表结构)
	// data  修改的数据(表对应的字段)
	UpdatesById(ctx context.Context, id int, model any, data map[string]any) error
}
