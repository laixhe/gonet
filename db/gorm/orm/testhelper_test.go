package orm

import (
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// testModel 测试用模型
type testModel struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

// fakeDialector 测试用伪数据库驱动, 仅用于构造可执行语句构建的 gorm.DB, 不做真实连接
type fakeDialector struct{}

func (fakeDialector) Name() string                                        { return "fake" }
func (fakeDialector) Initialize(*gorm.DB) error                           { return nil }
func (fakeDialector) Migrator(*gorm.DB) gorm.Migrator                     { return nil }
func (fakeDialector) DataTypeOf(*schema.Field) string                     { return "" }
func (fakeDialector) DefaultValueOf(*schema.Field) clause.Expression      { return nil }
func (fakeDialector) BindVarTo(w clause.Writer, _ *gorm.Statement, _ any) { w.WriteString("?") }
func (fakeDialector) QuoteTo(w clause.Writer, str string)                 { w.WriteString(str) }
func (fakeDialector) Explain(sql string, _ ...any) string                 { return sql }

// newTestDB 构造一个不依赖真实数据库连接的 gorm.DB, 用于测试语句构建类方法
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(fakeDialector{}, &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open 失败: %v", err)
	}
	return db
}
