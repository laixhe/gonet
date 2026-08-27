package mongo

import (
	"context"
	"sync"
	"testing"
	"time"

	mongov2 "go.mongodb.org/mongo-driver/v2/mongo"
	optionsv2 "go.mongodb.org/mongo-driver/v2/mongo/options"
)

// connectTestClient 创建一个连接到本地 MongoDB 的测试客户端。
// 如果本地没有 MongoDB 运行，则跳过当前测试。
func connectTestClient(t *testing.T) *MClient {
	t.Helper()

	opts := optionsv2.Client()
	opts.ApplyURI("mongodb://127.0.0.1:27017")
	opts.SetConnectTimeout(2 * time.Second)
	opts.SetServerSelectionTimeout(2 * time.Second)

	client, err := mongov2.Connect(opts)
	if err != nil {
		t.Skipf("跳过测试：无法连接 MongoDB (%v)", err)
		return nil
	}

	// 快速 ping 确认连接可用
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		t.Skipf("跳过测试：MongoDB ping 失败 (%v)", err)
		return nil
	}

	return &MClient{
		client:          client,
		defaultDatabase: client.Database("test_gonet"),
		databaseMap:     make(map[string]*mongov2.Database),
	}
}

// ======================== Config.Check 测试 ========================

func TestConfig_Check_Nil(t *testing.T) {
	var c *Config
	err := c.Check()
	if err == nil {
		t.Fatal("nil Config 应该返回错误")
	}
	if err.Error() != "没有Mongo配置" {
		t.Fatalf("错误信息不匹配: %s", err.Error())
	}
}

func TestConfig_Check_EmptyUri(t *testing.T) {
	c := &Config{Database: "test"}
	err := c.Check()
	if err == nil {
		t.Fatal("Uri 为空应该返回错误")
	}
}

func TestConfig_Check_EmptyDatabase(t *testing.T) {
	c := &Config{Uri: "mongodb://127.0.0.1:27017"}
	err := c.Check()
	if err == nil {
		t.Fatal("Database 为空应该返回错误")
	}
}

func TestConfig_Check_Valid(t *testing.T) {
	c := &Config{
		Uri:      "mongodb://127.0.0.1:27017",
		Database: "test",
	}
	if err := c.Check(); err != nil {
		t.Fatalf("有效 Config 检查失败: %v", err)
	}
}

// ======================== MClient 方法测试 ========================

func TestMClient_Client(t *testing.T) {
	mc := connectTestClient(t)
	defer mc.Close(context.Background())

	client := mc.Client()
	if client == nil {
		t.Fatal("Client() 返回 nil")
	}
	// 应该返回同一个 client 实例
	if client != mc.client {
		t.Fatal("Client() 返回的不是内部 client")
	}
}

func TestMClient_Collection(t *testing.T) {
	mc := connectTestClient(t)
	defer mc.Close(context.Background())

	col := mc.Collection("users")
	if col == nil {
		t.Fatal("Collection() 返回 nil")
	}

	// 多次调用应返回不同的 Collection 对象（mongo driver 行为）
	col2 := mc.Collection("users")
	if col2 == nil {
		t.Fatal("第二次 Collection() 返回 nil")
	}
}

func TestMClient_Database_Default(t *testing.T) {
	mc := connectTestClient(t)
	defer mc.Close(context.Background())

	db := mc.Database("test_gonet")
	if db == nil {
		t.Fatal("Database() 返回 nil")
	}
}

func TestMClient_Database_Cache(t *testing.T) {
	mc := connectTestClient(t)
	defer mc.Close(context.Background())

	db1 := mc.Database("other_db")
	db2 := mc.Database("other_db")

	if db1 != db2 {
		t.Fatal("Database() 应返回缓存的同一实例")
	}
}

func TestMClient_Database_DifferentNames(t *testing.T) {
	mc := connectTestClient(t)
	defer mc.Close(context.Background())

	db1 := mc.Database("db_a")
	db2 := mc.Database("db_b")

	if db1 == db2 {
		t.Fatal("不同数据库名应返回不同实例")
	}
}

func TestMClient_Ping(t *testing.T) {
	mc := connectTestClient(t)
	defer mc.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := mc.Ping(ctx); err != nil {
		t.Fatalf("Ping 失败: %v", err)
	}
}

func TestMClient_Ping_ContextCanceled(t *testing.T) {
	mc := connectTestClient(t)
	defer mc.Close(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	err := mc.Ping(ctx)
	if err == nil {
		t.Fatal("已取消的 context 应该导致 Ping 失败")
	}
}

func TestMClient_Close(t *testing.T) {
	mc := connectTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := mc.Close(ctx); err != nil {
		t.Fatalf("Close 失败: %v", err)
	}

	// Close 后 Ping 应该失败
	err := mc.Ping(ctx)
	if err == nil {
		t.Fatal("Close 后 Ping 应该失败")
	}
}

func TestMClient_Close_Twice(t *testing.T) {
	mc := connectTestClient(t)

	ctx := context.Background()
	if err := mc.Close(ctx); err != nil {
		t.Fatalf("第一次 Close 失败: %v", err)
	}

	// 第二次 Close 可能返回错误（已断开），但这不应 panic
	_ = mc.Close(ctx)
}

// ======================== Database 并发安全测试 ========================

func TestMClient_Database_Concurrent(t *testing.T) {
	mc := connectTestClient(t)
	defer mc.Close(context.Background())

	const goroutines = 50
	const dbName = "concurrent_db"

	var wg sync.WaitGroup
	wg.Add(goroutines)

	results := make([]*mongov2.Database, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx] = mc.Database(dbName)
		}(i)
	}

	wg.Wait()

	// 所有 goroutine 应拿到同一个 Database 实例
	base := results[0]
	for i := 1; i < goroutines; i++ {
		if results[i] != base {
			t.Fatalf("并发 Database() 返回了不同实例 (index %d)", i)
		}
	}
}

func TestMClient_Database_Concurrent_DifferentNames(t *testing.T) {
	mc := connectTestClient(t)
	defer mc.Close(context.Background())

	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			name := "db_" + string(rune('A'+idx%10))
			db := mc.Database(name)
			if db == nil {
				t.Errorf("Database(%s) 返回 nil", name)
			}
		}(i)
	}

	wg.Wait()
}

// ======================== Init 集成测试 ========================

func TestInit_InvalidConfig_Nil(t *testing.T) {
	_, err := Init(nil)
	if err == nil {
		t.Fatal("nil Config 应该返回错误")
	}
}

func TestInit_InvalidConfig_EmptyUri(t *testing.T) {
	_, err := Init(&Config{Database: "test"})
	if err == nil {
		t.Fatal("空 Uri 应该返回错误")
	}
}

func TestInit_InvalidConfig_EmptyDatabase(t *testing.T) {
	_, err := Init(&Config{Uri: "mongodb://127.0.0.1:27017"})
	if err == nil {
		t.Fatal("空 Database 应该返回错误")
	}
}

func TestInit_ValidConfig(t *testing.T) {
	config := &Config{
		Uri:      "mongodb://127.0.0.1:27017",
		Database: "test_gonet",
	}

	mc, err := Init(config)
	if err != nil {
		t.Skipf("跳过测试：Init 失败 (%v) — 确认本地 MongoDB 已启动", err)
		return
	}
	defer mc.Close(context.Background())

	// 验证 defaultDatabase 使用了正确的数据库名
	col := mc.Collection("test_col")
	if col == nil {
		t.Fatal("Collection() 返回 nil")
	}
}

func TestInit_WithPoolConfig(t *testing.T) {
	config := &Config{
		Uri:             "mongodb://127.0.0.1:27017",
		Database:        "test_gonet",
		MaxPoolSize:     50,
		MinPoolSize:     5,
		MaxConnIdleTime: 120,
	}

	mc, err := Init(config)
	if err != nil {
		t.Skipf("跳过测试：Init 失败 (%v) — 确认本地 MongoDB 已启动", err)
		return
	}
	defer mc.Close(context.Background())

	if mc.config.MaxPoolSize != 50 {
		t.Errorf("MaxPoolSize = %d, want 50", mc.config.MaxPoolSize)
	}
	if mc.config.MaxConnIdleTime != 120 {
		t.Errorf("MaxConnIdleTime = %d, want 120", mc.config.MaxConnIdleTime)
	}
}
