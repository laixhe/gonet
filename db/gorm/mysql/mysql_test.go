package mysql

import (
	"strings"
	"testing"

	"github.com/laixhe/gonet/db/gorm/orm"
)

// nopWriter 丢弃 gorm 日志
type nopWriter struct{}

func (nopWriter) Printf(string, ...any) {}

func TestInit(t *testing.T) {
	t.Run("接口实现断言", func(t *testing.T) {
		var _ orm.Client = (*Client)(nil)
	})
	t.Run("驱动不匹配", func(t *testing.T) {
		_, err := Init(&orm.Config{Driver: orm.DriverSqlite, Dsn: "x"}, nopWriter{}, "test")
		if err == nil || !strings.Contains(err.Error(), "数据库驱动只支持 mysql") {
			t.Errorf("err = %v, want 驱动不匹配错误", err)
		}
	})
	t.Run("空配置", func(t *testing.T) {
		_, err := Init(nil, nopWriter{}, "test")
		if err == nil {
			t.Error("nil 配置应返回错误")
		}
	})
	t.Run("配置校验失败", func(t *testing.T) {
		_, err := Init(&orm.Config{Driver: orm.DriverMysql}, nopWriter{}, "test")
		if err == nil {
			t.Error("缺少 Dsn 应返回错误")
		}
	})
	t.Run("非法 DSN 连接失败", func(t *testing.T) {
		// 非法的 DSN 在解析阶段即失败, 不发起真实网络连接
		_, err := Init(&orm.Config{Driver: orm.DriverMysql, Dsn: "bad-dsn"}, nopWriter{}, "test")
		if err == nil {
			t.Error("非法 DSN 应返回错误")
		}
	})
}
