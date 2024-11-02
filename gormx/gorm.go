package gormx

import (
	"errors"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/laixhe/gonet/logx"
	"github.com/laixhe/gonet/proto/gen/config/cdb"
)

// Gormx 客户端
type Gormx struct {
	client *gorm.DB
}

// Ping 判断服务是否可用
func (g *Gormx) Ping() error {
	sqlDB, err := g.client.DB()
	if err != nil {
		return err
	}
	// 验证数据库连接是否正常
	return sqlDB.Ping()
}

var db *Gormx

// Client get db client
func Client() *gorm.DB {
	return db.client
}

// Connect 连接数据库
func Connect(c *cdb.DB) (*Gormx, error) {

	defaultLogger := logger.New(NewWriter(log.New(os.Stdout, " ", log.LstdFlags)), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Info,
		Colorful:      true,
	})

	client, err := gorm.Open(mysql.Open(c.Dsn), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),
		Logger: defaultLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	})
	if err != nil {
		return nil, err
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
	return &Gormx{
		client: client,
	}, nil
}

// Init 初始化数据库
func Init(c *cdb.DB) {
	if c == nil {
		panic(errors.New("db config is nil"))
	}
	if c.Dsn == "" {
		panic(errors.New("db config dsn is nil"))
	}
	logx.Debugf("db Config=%v", c)
	logx.Debug("db 开始初始化...")

	var err error
	db, err = Connect(c)
	if err != nil {
		panic(err)
	}

	logx.Debug("db 初始化完毕...")
}
