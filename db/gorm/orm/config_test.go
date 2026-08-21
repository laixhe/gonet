package orm

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func TestConfigCheck(t *testing.T) {
	valid := &Config{
		Driver: DriverMysql,
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/test",
	}
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{"合法配置", valid, false},
		{"空配置", nil, true},
		{"空驱动", &Config{Dsn: "dsn"}, true},
		{"空连接地址", &Config{Driver: DriverMysql}, true},
		{"不支持的驱动", &Config{Driver: "oracle", Dsn: "dsn"}, true},
		{"空闲连接池负数", &Config{Driver: DriverMysql, Dsn: "dsn", MaxIdleCount: -1}, true},
		{"打开连接数负数", &Config{Driver: DriverMysql, Dsn: "dsn", MaxOpenCount: -1}, true},
		{"连接复用时间负数", &Config{Driver: DriverMysql, Dsn: "dsn", MaxLifeTime: -1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Check()
			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigSetLog(t *testing.T) {
	t.Run("默认日志级别为 Info", func(t *testing.T) {
		c := &Config{}
		l, ok := c.SetLog(&nopWriter{}, "rid").(*Logger)
		if !ok {
			t.Fatal("SetLog 返回类型应为 *Logger")
		}
		if l.LogLevel != gormLogger.Info {
			t.Errorf("LogLevel = %d, want %d", l.LogLevel, gormLogger.Info)
		}
	})
	t.Run("自定义日志级别", func(t *testing.T) {
		c := &Config{LogLevel: gormLogger.Warn}
		l, _ := c.SetLog(&nopWriter{}, "rid").(*Logger)
		if l.LogLevel != gormLogger.Warn {
			t.Errorf("LogLevel = %d, want %d", l.LogLevel, gormLogger.Warn)
		}
	})
	t.Run("非法日志级别回退默认", func(t *testing.T) {
		c := &Config{LogLevel: gormLogger.LogLevel(9)}
		l, _ := c.SetLog(&nopWriter{}, "rid").(*Logger)
		if l.LogLevel != gormLogger.Info {
			t.Errorf("LogLevel = %d, want %d", l.LogLevel, gormLogger.Info)
		}
	})
	t.Run("ParameterizedQueries 透传并隐藏参数", func(t *testing.T) {
		c := &Config{LogLevel: gormLogger.Info, ParameterizedQueries: true}
		l, _ := c.SetLog(&nopWriter{}, "rid").(*Logger)
		if !l.ParameterizedQueries {
			t.Error("ParameterizedQueries 应为 true")
		}
		sql, params := l.ParamsFilter(context.Background(), "SELECT * FROM user WHERE id = ?", 1)
		if sql != "SELECT * FROM user WHERE id = ?" || params != nil {
			t.Errorf("ParamsFilter = (%q, %v), want (sql, nil)", sql, params)
		}
	})
	t.Run("默认打印 SQL 参数", func(t *testing.T) {
		c := &Config{LogLevel: gormLogger.Info}
		l, _ := c.SetLog(&nopWriter{}, "rid").(*Logger)
		_, params := l.ParamsFilter(context.Background(), "SELECT * FROM user WHERE id = ?", 1)
		if len(params) != 1 {
			t.Errorf("params = %v, want 保留原参数", params)
		}
	})
}

func TestErrorHelpers(t *testing.T) {
	if !IsRecordNotFound(gorm.ErrRecordNotFound) {
		t.Error("IsRecordNotFound 应能识别 gorm.ErrRecordNotFound")
	}
	if IsRecordNotFound(errors.New("other")) {
		t.Error("IsRecordNotFound 不应识别其他错误")
	}
	if !IsNoUpdatedLines(ErrorNoUpdatedLines) {
		t.Error("IsNoUpdatedLines 应能识别 ErrorNoUpdatedLines")
	}
	if !IsNoUpdatedLines(fmt.Errorf("wrap: %w", ErrorNoUpdatedLines)) {
		t.Error("IsNoUpdatedLines 应能识别包装后的 ErrorNoUpdatedLines")
	}
	if IsNoUpdatedLines(errors.New("other")) {
		t.Error("IsNoUpdatedLines 不应识别其他错误")
	}
}
