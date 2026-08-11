package redis

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

// ==================== Config.Check ====================

func TestConfig_Check_Nil(t *testing.T) {
	var cfg *Config
	err := cfg.Check()
	if err == nil {
		t.Fatal("nil Config 应返回错误")
	}
}

func TestConfig_Check_EmptyAddr(t *testing.T) {
	cfg := &Config{Addr: ""}
	err := cfg.Check()
	if err == nil {
		t.Fatal("空 Addr 应返回错误")
	}
}

func TestConfig_Check_NegativeDbNum(t *testing.T) {
	cfg := &Config{Addr: "127.0.0.1:6379", DbNum: -1}
	err := cfg.Check()
	if err != nil {
		t.Fatalf("合法配置不应返回错误: %v", err)
	}
	if cfg.DbNum != 0 {
		t.Fatalf("负值 DbNum 应归零，实际: %d", cfg.DbNum)
	}
}

func TestConfig_Check_NegativePoolSize(t *testing.T) {
	cfg := &Config{Addr: "127.0.0.1:6379", PoolSize: -5}
	err := cfg.Check()
	if err != nil {
		t.Fatalf("合法配置不应返回错误: %v", err)
	}
	if cfg.PoolSize != 0 {
		t.Fatalf("负值 PoolSize 应归零，实际: %d", cfg.PoolSize)
	}
}

func TestConfig_Check_NegativeMinIdleConn(t *testing.T) {
	cfg := &Config{Addr: "127.0.0.1:6379", MinIdleConn: -3}
	err := cfg.Check()
	if err != nil {
		t.Fatalf("合法配置不应返回错误: %v", err)
	}
	if cfg.MinIdleConn != 0 {
		t.Fatalf("负值 MinIdleConn 应归零，实际: %d", cfg.MinIdleConn)
	}
}

func TestConfig_Check_NegativeDialTimeout(t *testing.T) {
	cfg := &Config{Addr: "127.0.0.1:6379", DialTimeout: -10}
	err := cfg.Check()
	if err != nil {
		t.Fatalf("合法配置不应返回错误: %v", err)
	}
	if cfg.DialTimeout != 0 {
		t.Fatalf("负值 DialTimeout 应归零，实际: %d", cfg.DialTimeout)
	}
}

func TestConfig_Check_NegativeReadTimeout(t *testing.T) {
	cfg := &Config{Addr: "127.0.0.1:6379", ReadTimeout: -2}
	err := cfg.Check()
	if err != nil {
		t.Fatalf("合法配置不应返回错误: %v", err)
	}
	if cfg.ReadTimeout != 0 {
		t.Fatalf("负值 ReadTimeout 应归零，实际: %d", cfg.ReadTimeout)
	}
}

func TestConfig_Check_NegativeWriteTimeout(t *testing.T) {
	cfg := &Config{Addr: "127.0.0.1:6379", WriteTimeout: -1}
	err := cfg.Check()
	if err != nil {
		t.Fatalf("合法配置不应返回错误: %v", err)
	}
	if cfg.WriteTimeout != 0 {
		t.Fatalf("负值 WriteTimeout 应归零，实际: %d", cfg.WriteTimeout)
	}
}

func TestConfig_Check_Valid(t *testing.T) {
	cfg := &Config{
		Addr:         "127.0.0.1:6379",
		DbNum:        1,
		Password:     "secret",
		PoolSize:     50,
		MinIdleConn:  10,
		DialTimeout:  5,
		ReadTimeout:  3,
		WriteTimeout: 3,
	}
	err := cfg.Check()
	if err != nil {
		t.Fatalf("完整配置不应返回错误: %v", err)
	}
	if cfg.DbNum != 1 {
		t.Fatalf("正值 DbNum 不应被修改，实际: %d", cfg.DbNum)
	}
	if cfg.PoolSize != 50 {
		t.Fatalf("正值 PoolSize 不应被修改，实际: %d", cfg.PoolSize)
	}
	if cfg.MinIdleConn != 10 {
		t.Fatalf("正值 MinIdleConn 不应被修改，实际: %d", cfg.MinIdleConn)
	}
	if cfg.DialTimeout != 5 {
		t.Fatalf("正值 DialTimeout 不应被修改，实际: %d", cfg.DialTimeout)
	}
}

func TestConfig_Check_ZeroValues(t *testing.T) {
	cfg := &Config{Addr: "127.0.0.1:6379"}
	err := cfg.Check()
	if err != nil {
		t.Fatalf("零值配置不应返回错误: %v", err)
	}
}

// ==================== initSingle ====================

func TestInitSingle_Basic(t *testing.T) {
	mr := miniredis.RunT(t)

	cfg := &Config{Addr: mr.Addr()}
	client := initSingle(cfg)
	if client == nil {
		t.Fatal("initSingle 不应返回 nil")
	}

	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("Ping 失败: %v", err)
	}
}

func TestInitSingle_WithAllOptions(t *testing.T) {
	mr := miniredis.RunT(t)

	cfg := &Config{
		Addr:         mr.Addr(),
		Password:     "pass",
		DbNum:        0,
		PoolSize:     20,
		MinIdleConn:  5,
		DialTimeout:  10,
		ReadTimeout:  6,
		WriteTimeout: 6,
	}
	client := initSingle(cfg)
	if client == nil {
		t.Fatal("initSingle 不应返回 nil")
	}

	status := client.Set(context.Background(), "foo", "bar", 0)
	if err := status.Err(); err != nil {
		t.Fatalf("Set 失败: %v", err)
	}

	val, err := client.Get(context.Background(), "foo").Result()
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if val != "bar" {
		t.Fatalf("值不匹配: 期望 bar, 实际 %s", val)
	}
}

// ==================== initCluster ====================

func TestInitCluster_ReturnsNonNil(t *testing.T) {
	cfg := &Config{
		Addr:     "127.0.0.1:7000,127.0.0.1:7001",
		Password: "pass",
		PoolSize: 10,
	}
	client := initCluster(cfg)
	if client == nil {
		t.Fatal("initCluster 不应返回 nil")
	}
}

// ==================== connect ====================

func TestConnect_SingleAddr(t *testing.T) {
	mr := miniredis.RunT(t)

	cfg := &Config{Addr: mr.Addr()}
	rc, err := connect(cfg)
	if err != nil {
		t.Fatalf("connect 失败: %v", err)
	}
	if rc == nil {
		t.Fatal("connect 不应返回 nil")
	}
	if rc.client == nil {
		t.Fatal("client 不应为 nil")
	}

	// 验证可以正常执行命令
	err = rc.client.Set(context.Background(), "k", "v", 0).Err()
	if err != nil {
		t.Fatalf("Set 失败: %v", err)
	}
}

// ==================== Init ====================

func TestInit_Success(t *testing.T) {
	mr := miniredis.RunT(t)

	cfg := &Config{
		Addr:        mr.Addr(),
		PoolSize:    10,
		MinIdleConn: 3,
	}
	rc, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init 失败: %v", err)
	}
	if rc == nil {
		t.Fatal("Init 不应返回 nil")
	}

	// Set + Get 完整流程
	if err := rc.Client().Set(context.Background(), "hello", "world", 0).Err(); err != nil {
		t.Fatalf("Set 失败: %v", err)
	}
	val, err := rc.Client().Get(context.Background(), "hello").Result()
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if val != "world" {
		t.Fatalf("值不匹配: 期望 world, 实际 %s", val)
	}
}

func TestInit_InvalidConfig(t *testing.T) {
	// nil config
	_, err := Init(nil)
	if err == nil {
		t.Fatal("nil Config 应返回错误")
	}

	// empty addr
	_, err = Init(&Config{})
	if err == nil {
		t.Fatal("空 Addr 应返回错误")
	}
}

func TestInit_ConnectionFailed(t *testing.T) {
	// 使用一个不存在的地址
	cfg := &Config{Addr: "127.0.0.1:19999"}
	_, err := Init(cfg)
	if err == nil {
		t.Fatal("连接不存在地址应返回错误")
	}
}

// ==================== RClient.Ping ====================

func TestRClient_Ping_Success(t *testing.T) {
	mr := miniredis.RunT(t)

	cfg := &Config{Addr: mr.Addr()}
	rc, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init 失败: %v", err)
	}

	if err := rc.Ping(); err != nil {
		t.Fatalf("Ping 失败: %v", err)
	}
}

func TestRClient_Ping_Fail(t *testing.T) {
	mr := miniredis.RunT(t)

	cfg := &Config{Addr: mr.Addr()}
	rc, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init 失败: %v", err)
	}

	// 关闭 miniredis 模拟服务不可达
	mr.Close()

	err = rc.Ping()
	if err == nil {
		t.Fatal("服务关闭后 Ping 应返回错误")
	}
}

// ==================== RClient.Client ====================

func TestRClient_Client(t *testing.T) {
	mr := miniredis.RunT(t)

	cfg := &Config{Addr: mr.Addr()}
	rc, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init 失败: %v", err)
	}

	client := rc.Client()
	if client == nil {
		t.Fatal("Client() 不应返回 nil")
	}

	// 验证返回的接口可以正常使用
	result := client.Ping(context.Background())
	if result.Err() != nil {
		t.Fatalf("通过 Client() Ping 失败: %v", result.Err())
	}
}

// ==================== RClient.Close ====================

func TestRClient_Close(t *testing.T) {
	mr := miniredis.RunT(t)

	cfg := &Config{Addr: mr.Addr()}
	rc, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init 失败: %v", err)
	}

	if err := rc.Close(); err != nil {
		t.Fatalf("Close 失败: %v", err)
	}

	// 关闭后再 Ping 应该失败
	err = rc.Ping()
	if err == nil {
		t.Fatal("Close 后再 Ping 应返回错误")
	}
}

// ==================== 集成测试: 完整生命周期 ====================

func TestFullLifecycle(t *testing.T) {
	mr := miniredis.RunT(t)

	// 1. 初始化
	cfg := &Config{
		Addr:         mr.Addr(),
		PoolSize:     10,
		MinIdleConn:  2,
		DialTimeout:  5,
		ReadTimeout:  3,
		WriteTimeout: 3,
	}
	rc, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init 失败: %v", err)
	}

	// 2. Ping
	if err := rc.Ping(); err != nil {
		t.Fatalf("Ping 失败: %v", err)
	}

	// 3. 读写
	client := rc.Client()
	err = client.Set(context.Background(), "count", "1", 0).Err()
	if err != nil {
		t.Fatalf("Set 失败: %v", err)
	}
	val, err := client.Get(context.Background(), "count").Result()
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if val != "1" {
		t.Fatalf("值不匹配: 期望 1, 实际 %s", val)
	}

	// 4. Incr
	n, err := client.Incr(context.Background(), "count").Result()
	if err != nil {
		t.Fatalf("Incr 失败: %v", err)
	}
	if n != 2 {
		t.Fatalf("Incr 结果不匹配: 期望 2, 实际 %d", n)
	}

	// 5. TTL
	err = client.Expire(context.Background(), "count", 10*time.Second).Err()
	if err != nil {
		t.Fatalf("Expire 失败: %v", err)
	}
	ttl := client.TTL(context.Background(), "count").Val()
	if ttl <= 0 {
		t.Fatalf("TTL 应大于 0, 实际: %v", ttl)
	}

	// 6. Del
	n, err = client.Del(context.Background(), "count").Result()
	if err != nil {
		t.Fatalf("Del 失败: %v", err)
	}
	if n != 1 {
		t.Fatalf("Del 应删除 1 个 key, 实际: %d", n)
	}

	// 7. 关闭
	if err := rc.Close(); err != nil {
		t.Fatalf("Close 失败: %v", err)
	}
}

// ==================== connect Ping 失败资源释放 ====================

// TestConnect_PingFail_SingleAddr 验证单机模式下 connect Ping 失败时连接池被正确释放
func TestConnect_PingFail_SingleAddr(t *testing.T) {
	// 启动 miniredis 后立即关闭，获得一个不可达的地址
	mr := miniredis.RunT(t)
	addr := mr.Addr()
	mr.Close()

	cfg := &Config{Addr: addr}

	// 先做几次预热调用，排除 runtime 自身 goroutine 波动
	for i := 0; i < 3; i++ {
		connect(cfg)
	}
	time.Sleep(100 * time.Millisecond)

	before := runtime.NumGoroutine()

	const iterations = 30
	for i := 0; i < iterations; i++ {
		rc, err := connect(cfg)
		if err == nil {
			rc.Close()
			t.Fatal("connect 应返回错误（Redis 不可达）")
		}
		if rc != nil {
			t.Fatal("connect 失败时 rc 应为 nil")
		}
	}

	// 等待 goroutine 完全退出
	time.Sleep(200 * time.Millisecond)

	after := runtime.NumGoroutine()
	diff := int64(after) - int64(before)

	if diff > int64(iterations/2) {
		t.Fatalf("可能 goroutine 泄漏: before=%d, after=%d, leaked=%d (预期接近 0)", before, after, diff)
	}
	t.Logf("goroutine diff: %d (before=%d, after=%d)", diff, before, after)
}

// TestConnect_PingFail_ClusterAddr 验证集群模式下 connect Ping 失败时连接池被正确释放
func TestConnect_PingFail_ClusterAddr(t *testing.T) {
	// 集群模式使用多个不可达地址（无需真实 miniredis）
	cfg := &Config{
		Addr: "127.0.0.1:17000,127.0.0.1:17001,127.0.0.1:17002",
	}

	// 预热
	for i := 0; i < 3; i++ {
		connect(cfg)
	}
	time.Sleep(200 * time.Millisecond)

	before := runtime.NumGoroutine()

	const iterations = 20
	for i := 0; i < iterations; i++ {
		rc, err := connect(cfg)
		if err == nil {
			rc.Close()
			t.Fatal("connect 应返回错误（集群不可达）")
		}
		if rc != nil {
			t.Fatal("connect 失败时 rc 应为 nil")
		}
	}

	// 集群模式下 goroutine 退出可能更慢，给予更长等待
	time.Sleep(500 * time.Millisecond)

	after := runtime.NumGoroutine()
	diff := int64(after) - int64(before)

	if diff > int64(iterations/2) {
		t.Fatalf("可能 goroutine 泄漏: before=%d, after=%d, leaked=%d (预期接近 0)", before, after, diff)
	}
	t.Logf("goroutine diff: %d (before=%d, after=%d)", diff, before, after)
}
