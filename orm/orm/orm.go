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
	// Schema 模式(postgresql专用)
	Schema() string
	// SchemaTableName 模式表名(postgresql专用)
	SchemaTableName(name string) string
}
