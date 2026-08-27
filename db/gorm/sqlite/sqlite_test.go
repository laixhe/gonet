package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"github.com/laixhe/gonet/db/gorm/orm"
)

// user 测试用模型
// 启用 SingularTable 后对应表名为 user
type user struct {
	ID   uint `gorm:"primaryKey"`
	Name string
	Age  int
}

// nopWriter 丢弃 gorm 日志
type nopWriter struct{}

func (nopWriter) Printf(string, ...any) {}

// newTestClient 初始化一个临时文件 SQLite 客户端并建表
func newTestClient(t *testing.T) orm.Client {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "test.db")
	client, err := Init(&orm.Config{
		Driver:   orm.DriverSqlite,
		Dsn:      dsn,
		LogLevel: gormLogger.Silent,
	}, nopWriter{}, "test")
	if err != nil {
		t.Fatalf("Init 失败: %v", err)
	}
	// 关闭连接池, 释放临时数据库文件
	t.Cleanup(func() {
		if sqlDB, err := client.Client().DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	// 创建测试表
	if err = client.Client().AutoMigrate(&user{}); err != nil {
		t.Fatalf("AutoMigrate 失败: %v", err)
	}
	return client
}

func TestInit(t *testing.T) {
	t.Run("驱动不匹配", func(t *testing.T) {
		_, err := Init(&orm.Config{Driver: orm.DriverMysql, Dsn: "x"}, nopWriter{}, "test")
		if err == nil || !strings.Contains(err.Error(), "数据库驱动只支持 sqlite") {
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
		_, err := Init(&orm.Config{Driver: orm.DriverSqlite}, nopWriter{}, "test")
		if err == nil {
			t.Error("缺少 Dsn 应返回错误")
		}
	})
	t.Run("接口实现断言", func(t *testing.T) {
		var _ orm.Client = (*Client)(nil)
	})
}

func TestPing(t *testing.T) {
	client := newTestClient(t)
	if err := client.Ping(); err != nil {
		t.Errorf("Ping 失败: %v", err)
	}
	if client.Client() == nil {
		t.Error("Client() 不应返回 nil")
	}
	if client.WithContext(context.Background()) == nil {
		t.Error("WithContext 不应返回 nil")
	}
}

func TestCreateAndQuery(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// 建立数据: alice + bob + carol 共 3 行
	alice := &user{Name: "alice", Age: 20}
	if err := client.Create(ctx, alice); err != nil {
		t.Fatalf("Create 失败: %v", err)
	}
	if alice.ID == 0 {
		t.Error("Create 后应回填自增 ID")
	}

	t.Run("Create 批量", func(t *testing.T) {
		users := []user{{Name: "bob", Age: 21}, {Name: "carol", Age: 22}}
		if err := client.Create(ctx, &users); err != nil {
			t.Fatalf("Create 批量失败: %v", err)
		}
		if users[0].ID == 0 || users[1].ID == 0 {
			t.Error("批量 Create 后应回填自增 ID")
		}
	})

	t.Run("GetById", func(t *testing.T) {
		got := &user{}
		if err := client.GetById(ctx, got, int(alice.ID)); err != nil {
			t.Fatalf("GetById 失败: %v", err)
		}
		if got.Name != "alice" {
			t.Errorf("Name = %q, want alice", got.Name)
		}
	})
	t.Run("GetById 未找到", func(t *testing.T) {
		got := &user{}
		err := client.GetById(ctx, got, 999999)
		if !orm.IsRecordNotFound(err) {
			t.Errorf("err = %v, want 记录未找到", err)
		}
	})
	t.Run("GetByField", func(t *testing.T) {
		got := &user{}
		if err := client.GetByField(ctx, got, "name", "alice"); err != nil {
			t.Fatalf("GetByField 失败: %v", err)
		}
		if got.Age != 20 {
			t.Errorf("Age = %d, want 20", got.Age)
		}
	})
	t.Run("GetByWhere", func(t *testing.T) {
		got := &user{}
		if err := client.GetByWhere(ctx, got, map[string]any{"name": "alice", "age": 20}); err != nil {
			t.Fatalf("GetByWhere 失败: %v", err)
		}
		if got.Name != "alice" {
			t.Errorf("Name = %q, want alice", got.Name)
		}
	})
	t.Run("FirstByField / LastByField", func(t *testing.T) {
		first, last := &user{}, &user{}
		if err := client.FirstByField(ctx, first, "name", "alice"); err != nil {
			t.Fatalf("FirstByField 失败: %v", err)
		}
		if err := client.LastByField(ctx, last, "name", "alice"); err != nil {
			t.Fatalf("LastByField 失败: %v", err)
		}
		if first.ID != last.ID {
			t.Errorf("单条数据 First(ID=%d) 与 Last(ID=%d) 应一致", first.ID, last.ID)
		}
	})
	t.Run("Count", func(t *testing.T) {
		n, err := client.Count(ctx, &user{}, map[string]any{"name": "alice"})
		if err != nil {
			t.Fatalf("Count 失败: %v", err)
		}
		if n != 1 {
			t.Errorf("Count = %d, want 1", n)
		}
	})
	t.Run("Count 全部", func(t *testing.T) {
		n, err := client.Count(ctx, &user{}, nil)
		if err != nil {
			t.Fatalf("Count 失败: %v", err)
		}
		if n != 3 {
			t.Errorf("Count = %d, want 3", n)
		}
	})
}

func TestSave(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	u := &user{Name: "dave", Age: 30}
	if err := client.Create(ctx, u); err != nil {
		t.Fatalf("Create 失败: %v", err)
	}
	// Save 会保存所有字段, 包括零值
	u.Name = "dave-updated"
	u.Age = 0
	if err := client.Save(ctx, u); err != nil {
		t.Fatalf("Save 失败: %v", err)
	}
	got := &user{}
	if err := client.GetById(ctx, got, int(u.ID)); err != nil {
		t.Fatalf("GetById 失败: %v", err)
	}
	if got.Name != "dave-updated" || got.Age != 0 {
		t.Errorf("Save 后 = %+v, want Name=dave-updated Age=0", got)
	}
}

func TestUpdates(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	u := &user{Name: "eve", Age: 40}
	if err := client.Create(ctx, u); err != nil {
		t.Fatalf("Create 失败: %v", err)
	}

	t.Run("更新成功", func(t *testing.T) {
		err := client.Updates(ctx, &user{}, map[string]any{"id": u.ID}, map[string]any{"name": "eve-updated", "age": 41})
		if err != nil {
			t.Fatalf("Updates 失败: %v", err)
		}
		got := &user{}
		_ = client.GetById(ctx, got, int(u.ID))
		if got.Name != "eve-updated" || got.Age != 41 {
			t.Errorf("更新后 = %+v", got)
		}
	})
	t.Run("无匹配行返回 ErrorNoUpdatedLines", func(t *testing.T) {
		err := client.Updates(ctx, &user{}, map[string]any{"id": 999999}, map[string]any{"name": "x"})
		if !orm.IsNoUpdatedLines(err) {
			t.Errorf("err = %v, want ErrorNoUpdatedLines", err)
		}
	})
	t.Run("空 update 返回错误", func(t *testing.T) {
		err := client.Updates(ctx, &user{}, map[string]any{"id": u.ID}, map[string]any{})
		if err == nil {
			t.Error("空 update 应返回错误")
		}
	})
}

func TestUpdatesById(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	u := &user{Name: "frank", Age: 50}
	if err := client.Create(ctx, u); err != nil {
		t.Fatalf("Create 失败: %v", err)
	}
	if err := client.UpdatesById(ctx, &user{}, int(u.ID), map[string]any{"age": 51}); err != nil {
		t.Fatalf("UpdatesById 失败: %v", err)
	}
	got := &user{}
	_ = client.GetById(ctx, got, int(u.ID))
	if got.Age != 51 {
		t.Errorf("Age = %d, want 51", got.Age)
	}
	// 无匹配 id
	err := client.UpdatesById(ctx, &user{}, 999999, map[string]any{"age": 1})
	if !orm.IsNoUpdatedLines(err) {
		t.Errorf("err = %v, want ErrorNoUpdatedLines", err)
	}
}

func TestDelete(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	u := &user{Name: "grace", Age: 60}
	if err := client.Create(ctx, u); err != nil {
		t.Fatalf("Create 失败: %v", err)
	}
	t.Run("DeleteById", func(t *testing.T) {
		if err := client.DeleteById(ctx, &user{}, int(u.ID)); err != nil {
			t.Fatalf("DeleteById 失败: %v", err)
		}
		got := &user{}
		if err := client.GetById(ctx, got, int(u.ID)); !orm.IsRecordNotFound(err) {
			t.Errorf("删除后应查不到, err = %v", err)
		}
	})
	t.Run("Delete", func(t *testing.T) {
		u2 := &user{Name: "heidi", Age: 61}
		if err := client.Create(ctx, u2); err != nil {
			t.Fatalf("Create 失败: %v", err)
		}
		if err := client.Delete(ctx, u2); err != nil {
			t.Fatalf("Delete 失败: %v", err)
		}
		got := &user{}
		if err := client.GetById(ctx, got, int(u2.ID)); !orm.IsRecordNotFound(err) {
			t.Errorf("删除后应查不到, err = %v", err)
		}
	})
}

func TestTransaction(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	t.Run("提交", func(t *testing.T) {
		err := client.Transaction(ctx, func(tx *gorm.DB) error {
			return tx.Create(&user{Name: "tx-commit", Age: 1}).Error
		})
		if err != nil {
			t.Fatalf("Transaction 失败: %v", err)
		}
		n, _ := client.Count(ctx, &user{}, map[string]any{"name": "tx-commit"})
		if n != 1 {
			t.Errorf("提交后 Count = %d, want 1", n)
		}
	})
	t.Run("回滚", func(t *testing.T) {
		err := client.Transaction(ctx, func(tx *gorm.DB) error {
			if err := tx.Create(&user{Name: "tx-rollback", Age: 2}).Error; err != nil {
				return err
			}
			return errors.New("rollback")
		})
		if err == nil || !strings.Contains(err.Error(), "rollback") {
			t.Fatalf("Transaction 应返回回滚错误, got %v", err)
		}
		n, _ := client.Count(ctx, &user{}, map[string]any{"name": "tx-rollback"})
		if n != 0 {
			t.Errorf("回滚后 Count = %d, want 0", n)
		}
	})
}
