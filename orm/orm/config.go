package orm

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/*
orm:
  # 驱动名称 mysql postgresql sqlite
  driver: mysql
  # 连接地址
  dsn: root:123456@tcp(127.0.0.1:3306)/webapi?charset=utf8mb4&parseTime=True&loc=Local
  # 设置空闲连接池中连接的最大数量
  max_idle_count: 10
  # 设置打开数据库连接的最大数量
  max_open_count: 100
  # 设置了连接可复用的最大时间(要比数据库设置连接超时时间少)(单位秒)
  max_life_time: 300
  # 设置日志级别 1=Silent 2=Error 3=Warn 4=Info
  log_level: 4
*/

// 数据库驱动名称
const (
	DriverMysql      = "mysql"
	DriverPostgresql = "postgresql"
	DriverSqlite     = "sqlite"
)

// 数据库配置
type Config struct {
	// 驱动名称 mysql postgresql sqlite
	Driver string `json:"driver" mapstructure:"driver" toml:"driver" yaml:"driver"`
	// 连接地址
	Dsn string `json:"dsn" mapstructure:"dsn" toml:"dsn" yaml:"dsn"`
	// 模式(postgresql专用)
	Schema string `json:"schema" mapstructure:"schema" toml:"schema" yaml:"schema"`
	// 设置空闲连接池中连接的最大数量
	MaxIdleCount int `json:"max_idle_count" mapstructure:"max_idle_count" toml:"max_idle_count" yaml:"max_idle_count"`
	// 设置打开数据库连接的最大数量
	MaxOpenCount int `json:"max_open_count" mapstructure:"max_open_count" toml:"max_open_count" yaml:"max_open_count"`
	// 设置了连接可复用的最大时间(要比数据库设置连接超时时间少)(单位秒)
	MaxLifeTime int `json:"max_life_time" mapstructure:"max_life_time" toml:"max_life_time" yaml:"max_life_time"`
	// 设置日志级别 1=Silent 2=Error 3=Warn 4=Info (默认 4)
	LogLevel logger.LogLevel `json:"log_level" mapstructure:"log_level" toml:"log_level" yaml:"log_level"`
}

func (c *Config) Check() error {
	if c == nil {
		return errors.New("数据库配置不能为空")
	}
	if c.Driver == "" {
		return errors.New("没有数据库驱动名称配置")
	}
	if c.Dsn == "" {
		return errors.New("没有数据库连接地址配置")
	}
	if !(c.Driver == DriverMysql || c.Driver == DriverPostgresql || c.Driver == DriverSqlite) {
		return errors.New("数据库驱动只支持 mysql postgresql sqlite")
	}
	return nil
}

func (c *Config) SetLog(writer logger.Writer, requestid string) logger.Interface {
	logLevel := logger.Info
	if c.LogLevel == logger.Silent ||
		c.LogLevel == logger.Error ||
		c.LogLevel == logger.Warn ||
		c.LogLevel == logger.Info {
		logLevel = c.LogLevel
	}
	return NewOrmLogger(writer,
		logger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      logLevel,
		}, requestid)
}

// SetDB 设置数据库连接
func (c *Config) SetDB(client *gorm.DB) error {
	db, err := client.DB()
	if err != nil {
		return err
	}
	if c.MaxIdleCount > 0 {
		// 设置空闲连接池中连接的最大数量
		db.SetMaxIdleConns(c.MaxIdleCount)
	}
	if c.MaxOpenCount > 0 {
		// 设置打开数据库连接的最大数量
		db.SetMaxOpenConns(c.MaxOpenCount)
	}
	if c.MaxLifeTime > 0 {
		// 设置了连接可复用的最大时间(要比数据库设置连接超时时间少)
		db.SetConnMaxLifetime(time.Duration(c.MaxLifeTime) * time.Second)
	}
	// 验证数据库连接是否正常
	err = db.Ping()
	if err != nil {
		return err
	}
	return nil
}
