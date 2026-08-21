package postgres

import (
	"errors"

	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/laixhe/gonet/db/gorm/orm"
)

type Client struct {
	*orm.BaseClient
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
	return &Client{BaseClient: orm.NewBaseClient(config, client)}, nil
}
