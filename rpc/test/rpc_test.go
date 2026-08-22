// Package rpc_test 是 github.com/laixhe/gonet/rpc 的集成测试（独立测试模块）。
// 使用 embedded etcd，无需外部依赖；仅通过公开 API 驱动，兼作使用示例。
package rpc_test

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
	"google.golang.org/grpc/status"

	"github.com/laixhe/gonet/rpc"
)

// TestMain 启动一个共享的 embedded etcd，供所有测试使用。
var etcdEndpoints []string

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	clientPort, peerPort := freePort(), freePort()
	dir, err := os.MkdirTemp("", "gonet-etcd-*")
	if err != nil {
		fmt.Println("etcd temp dir:", err)
		return 1
	}
	cfg := embed.NewConfig()
	cfg.Dir = dir
	cfg.ListenClientUrls = []url.URL{{Scheme: "http", Host: fmt.Sprintf("127.0.0.1:%d", clientPort)}}
	cfg.AdvertiseClientUrls = []url.URL{{Scheme: "http", Host: fmt.Sprintf("127.0.0.1:%d", clientPort)}}
	cfg.ListenPeerUrls = []url.URL{{Scheme: "http", Host: fmt.Sprintf("127.0.0.1:%d", peerPort)}}
	cfg.AdvertisePeerUrls = []url.URL{{Scheme: "http", Host: fmt.Sprintf("127.0.0.1:%d", peerPort)}}
	cfg.InitialCluster = cfg.InitialClusterFromName(cfg.Name)
	cfg.LogOutputs = []string{filepath.Join(dir, "etcd.log")}
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		fmt.Println("start embedded etcd:", err)
		return 1
	}
	select {
	case <-e.Server.ReadyNotify():
	case <-time.After(20 * time.Second):
		e.Close()
		fmt.Println("embedded etcd start timeout")
		return 1
	}
	etcdEndpoints = []string{fmt.Sprintf("http://127.0.0.1:%d", clientPort)}
	code := m.Run()
	e.Close()
	_ = os.RemoveAll(dir)
	return code
}

func freePort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

// fakeClientConn 记录 UpdateState 调用，用于直接驱动 Discovery.Build 的断言。
type fakeClientConn struct {
	mu     sync.Mutex
	states []resolver.State
}

func (f *fakeClientConn) UpdateState(state resolver.State) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.states = append(f.states, state)
	return nil
}

func (f *fakeClientConn) ReportError(error) {}

func (f *fakeClientConn) NewAddress(addresses []resolver.Address) {
	_ = f.UpdateState(resolver.State{Addresses: addresses})
}

func (f *fakeClientConn) ParseServiceConfig(string) *serviceconfig.ParseResult {
	return nil
}

func (f *fakeClientConn) lastState() resolver.State {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.states) == 0 {
		return resolver.State{}
	}
	return f.states[len(f.states)-1]
}

// waitAddresses 轮询直到最后一次 UpdateState 的地址数等于 want。
func (f *fakeClientConn) waitAddresses(t *testing.T, want int) []resolver.Address {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if st := f.lastState(); len(st.Addresses) == want {
			return st.Addresses
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %d addresses, last state: %+v", want, f.lastState())
	return nil
}

// TestRegisterAndDiscover 覆盖：注册 → 初始发现 → 实例下线(DELETE) → 新实例上线(PUT)。
// 顺序设计成天然无竞态：两个实例都在 Build 之前注册（走初始 Get）；
// 下线用 DELETE 事件验证 watch 已存活；之后的新实例 PUT 必然被同一 watch 送达。
func TestRegisterAndDiscover(t *testing.T) {
	const svc = "unit-service"
	const addr1 = "127.0.0.1:10001"
	const addr2 = "127.0.0.1:10002"
	const addr3 = "127.0.0.1:10003"

	reg1, err := rpc.NewRegister(etcdEndpoints, svc, addr1, 10)
	if err != nil {
		t.Fatalf("NewRegister(addr1): %v", err)
	}
	t.Cleanup(func() { _ = reg1.Close() })

	reg2, err := rpc.NewRegister(etcdEndpoints, svc, addr2, 10)
	if err != nil {
		t.Fatalf("NewRegister(addr2): %v", err)
	}
	t.Cleanup(func() { _ = reg2.Close() })

	// 共享 Discovery：进程内只创建一个（grpc:// scheme 只注册一次），不会在测试中被 Close
	disc, err := sharedDiscovery()
	if err != nil {
		t.Fatalf("NewDiscovery: %v", err)
	}

	cc := &fakeClientConn{}
	// 规范写法 grpc:///svc：serviceName 在 URL.Path，走 target.Endpoint()
	target := resolver.Target{URL: url.URL{Scheme: rpc.SchemaName, Path: "/" + svc}}
	r, err := disc.Build(target, cc, resolver.BuildOptions{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	t.Cleanup(r.Close)

	// 初始能发现已注册的两个实例（无选项注册 → 裸地址，不带 Attributes）
	addrs := cc.waitAddresses(t, 2)
	got := map[string]bool{}
	for _, a := range addrs {
		got[a.Addr] = true
		if a.Attributes != nil {
			t.Fatalf("plain registration should have no attributes, got %v", a.Attributes)
		}
	}
	if !got[addr1] || !got[addr2] {
		t.Fatalf("initial addrs = %v, want %q and %q", addrs, addr1, addr2)
	}

	// 一个实例下线（注销）→ watch DELETE → 1 个地址
	_ = reg2.Close()
	addrs = cc.waitAddresses(t, 1)
	if got := addrs[0].Addr; got != addr1 {
		t.Fatalf("after dereg, addr = %q, want %q", got, addr1)
	}

	// 新实例上线 → watch PUT → 2 个地址（此时 watch 已被 DELETE 事件证明存活）
	reg3, err := rpc.NewRegister(etcdEndpoints, svc, addr3, 10)
	if err != nil {
		t.Fatalf("NewRegister(addr3): %v", err)
	}
	t.Cleanup(func() { _ = reg3.Close() })
	addrs = cc.waitAddresses(t, 2)
	got = map[string]bool{}
	for _, a := range addrs {
		got[a.Addr] = true
	}
	if !got[addr1] || !got[addr3] {
		t.Fatalf("after add, addrs = %v, want %q and %q", addrs, addr1, addr3)
	}
}

// TestDiscoverImmediateRegister Build 后立即注册的实例也不能漏掉。
// 回归测试：watcher 必须从初始 Get 的 revision 之后开始监听（WithRev），
// 否则 Get 与 Watch 之间注册的实例会永久丢失。
func TestDiscoverImmediateRegister(t *testing.T) {
	const svc = "immediate-service"
	const addr = "127.0.0.1:10001"

	disc, err := sharedDiscovery()
	if err != nil {
		t.Fatalf("NewDiscovery: %v", err)
	}
	cc := &fakeClientConn{}
	target := resolver.Target{URL: url.URL{Scheme: rpc.SchemaName, Path: "/" + svc}}
	r, err := disc.Build(target, cc, resolver.BuildOptions{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	t.Cleanup(r.Close)

	// Build 返回后立即注册（其 PUT 的 revision 必然大于初始 Get 的 revision）
	reg, err := rpc.NewRegister(etcdEndpoints, svc, addr, 10)
	if err != nil {
		t.Fatalf("NewRegister: %v", err)
	}
	t.Cleanup(func() { _ = reg.Close() })

	addrs := cc.waitAddresses(t, 1)
	if got := addrs[0].Addr; got != addr {
		t.Fatalf("addr = %q, want %q", got, addr)
	}
}

// TestRegisterCloseIdempotent Close 可重复调用。
func TestRegisterCloseIdempotent(t *testing.T) {
	reg, err := rpc.NewRegister(etcdEndpoints, "close-service", "127.0.0.1:10001", 10)
	if err != nil {
		t.Fatalf("NewRegister: %v", err)
	}
	if err := reg.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := reg.Close(); err != nil {
		t.Fatalf("second Close should be no-op, got: %v", err)
	}
}

// TestAutoKeepAlive NewRegister 自动续租：不手动续租，超过 TTL 后注册仍有效；Close 后注销。
func TestAutoKeepAlive(t *testing.T) {
	const svc = "keepalive-service"
	const addr = "127.0.0.1:10001"

	reg, err := rpc.NewRegister(etcdEndpoints, svc, addr, 2) // 2s 租约
	if err != nil {
		t.Fatalf("NewRegister: %v", err)
	}

	cli, err := clientv3.New(clientv3.Config{Endpoints: etcdEndpoints, DialTimeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("etcd client: %v", err)
	}
	t.Cleanup(func() { _ = cli.Close() })
	key := "/" + rpc.SchemaName + "/" + svc + "/" + addr

	// 不手动启动任何续租监听，等待超过 2 个租约周期，key 应仍存在（自动续租生效）
	time.Sleep(4 * time.Second)
	resp, err := cli.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(resp.Kvs) == 0 {
		t.Fatal("registration expired: auto keepalive not working")
	}

	// Close 后租约撤销，key 应被删除
	if err := reg.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	resp, err = cli.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("Get after Close: %v", err)
	}
	if len(resp.Kvs) != 0 {
		t.Fatal("registration still exists after Close")
	}
}

// TestRegisterWithMetadata 带 weight/metadata 注册的实例，发现时 Address 应带 Attributes。
func TestRegisterWithMetadata(t *testing.T) {
	const svc = "meta-service"
	const addr = "127.0.0.1:10001"

	reg, err := rpc.NewRegister(etcdEndpoints, svc, addr, 10,
		rpc.WithWeight(10), rpc.WithMetadata(map[string]string{"zone": "a"}))
	if err != nil {
		t.Fatalf("NewRegister: %v", err)
	}
	t.Cleanup(func() { _ = reg.Close() })

	disc, err := sharedDiscovery()
	if err != nil {
		t.Fatalf("NewDiscovery: %v", err)
	}
	cc := &fakeClientConn{}
	target := resolver.Target{URL: url.URL{Scheme: rpc.SchemaName, Path: "/" + svc}}
	r, err := disc.Build(target, cc, resolver.BuildOptions{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	t.Cleanup(r.Close)

	addrs := cc.waitAddresses(t, 1)
	a := addrs[0]
	if a.Addr != addr {
		t.Fatalf("addr = %q, want %q", a.Addr, addr)
	}
	if a.Attributes == nil {
		t.Fatal("expected attributes for metadata registration")
	}
	if w, ok := a.Attributes.Value("weight").(int); !ok || w != 10 {
		t.Fatalf("weight = %v, want 10", a.Attributes.Value("weight"))
	}
	md, ok := a.Attributes.Value("metadata").(map[string]string)
	if !ok || md["zone"] != "a" {
		t.Fatalf("metadata = %v, want zone=a", a.Attributes.Value("metadata"))
	}
}

// ---------- 端到端：raw codec + 极简 gRPC 服务 ----------

// rawCodec 把 []byte 直接当消息体，避免测试里定义 proto 文件。
type rawCodec struct{}

func (rawCodec) Marshal(v any) ([]byte, error) {
	b, ok := v.([]byte)
	if !ok {
		return nil, fmt.Errorf("rawCodec: want []byte, got %T", v)
	}
	return b, nil
}

func (rawCodec) Unmarshal(data []byte, v any) error {
	p, ok := v.(*[]byte)
	if !ok {
		return fmt.Errorf("rawCodec: want *[]byte, got %T", v)
	}
	*p = data
	return nil
}

func (rawCodec) Name() string { return "raw" }

type pingService interface{}

func pingHandler(_ any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new([]byte)
	if err := dec(in); err != nil {
		return nil, err
	}
	handler := func(_ context.Context, req any) (any, error) {
		return []byte("pong:" + string(req.([]byte))), nil
	}
	if interceptor == nil {
		return handler(ctx, *in)
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Ping/Ping"}
	return interceptor(ctx, *in, info, handler)
}

var testPingServiceDesc = &grpc.ServiceDesc{
	ServiceName: "test.Ping",
	HandlerType: (*pingService)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Ping", Handler: pingHandler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "test.proto",
}

// 共享的全局 Discovery：进程内第一个 NewDiscovery 注册了 grpc:// scheme，
// 所有测试共用且不 Close，保证注册的 builder 一直可用。
var (
	sharedDiscOnce sync.Once
	sharedDisc     *rpc.Discovery
	sharedDiscErr  error
)

func sharedDiscovery() (*rpc.Discovery, error) {
	sharedDiscOnce.Do(func() {
		sharedDisc, sharedDiscErr = rpc.NewDiscovery(etcdEndpoints, 5*time.Second)
	})
	return sharedDisc, sharedDiscErr
}

// waitServerAddr 等待服务端 Start 完成并返回实际监听地址（支持 :0 随机端口）。
func waitServerAddr(t *testing.T, s *rpc.Server) string {
	t.Helper()
	var addr string
	eventually(t, 5*time.Second, func() error {
		addr = s.Addr()
		if addr == "" || addr == "127.0.0.1:0" {
			return fmt.Errorf("server not started yet")
		}
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err != nil {
			return err
		}
		_ = conn.Close()
		return nil
	})
	return addr
}

func newPingServer(t *testing.T) *rpc.Server {
	t.Helper()
	s, err := rpc.NewServer("127.0.0.1:0", "", "", rpc.WithServerOption(grpc.ForceServerCodec(rawCodec{})))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	s.RegisterService(testPingServiceDesc, struct{}{})
	go func() { _ = s.Start() }()
	waitServerAddr(t, s)
	return s
}

func newRegister(t *testing.T, svc, addr string) *rpc.Register {
	t.Helper()
	reg, err := rpc.NewRegister(etcdEndpoints, svc, addr, 10)
	if err != nil {
		t.Fatalf("NewRegister(%s): %v", addr, err)
	}
	return reg
}

// eventually 轮询 fn 直到成功或超时。
func eventually(t *testing.T, timeout time.Duration, fn func() error) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if err := fn(); err == nil {
			return
		} else {
			lastErr = err
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("eventually failed: %v", lastErr)
}

// eventuallyPing 轮询调用 Ping，直到成功返回 want。
func eventuallyPing(t *testing.T, c *rpc.Client, want string) {
	t.Helper()
	eventually(t, 15*time.Second, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var reply []byte
		err := c.Conn().Invoke(ctx, "/test.Ping/Ping", []byte("hi"), &reply)
		if err != nil {
			return err
		}
		if string(reply) != want {
			return fmt.Errorf("reply = %q, want %q", reply, want)
		}
		return nil
	})
}

// panicHandler 故意 panic，用于验证 RecoveryUnaryServerInterceptor。
// 注意：按 gRPC 拦截器模型，panic 必须发生在传给 interceptor 的 handler 内部，
// 拦截器才会包住它（生成代码里就是这种结构）。
func panicHandler(_ any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	var in []byte
	if err := dec(&in); err != nil {
		return nil, err
	}
	handler := func(_ context.Context, _ any) (any, error) {
		panic("boom: " + string(in))
	}
	if interceptor == nil {
		return handler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Panic/Panic"}
	return interceptor(ctx, in, info, handler)
}

var testPanicServiceDesc = &grpc.ServiceDesc{
	ServiceName: "test.Panic",
	HandlerType: (*pingService)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Panic", Handler: panicHandler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "test.proto",
}

// TestRecoveryInterceptor handler panic 应转成 codes.Internal，服务进程不崩溃。
func TestRecoveryInterceptor(t *testing.T) {
	const svc = "panic-service"

	s, err := rpc.NewServer("127.0.0.1:0", "", "", rpc.WithServerOption(
		grpc.ChainUnaryInterceptor(rpc.RecoveryUnaryServerInterceptor()),
		grpc.ForceServerCodec(rawCodec{}),
	))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	s.RegisterService(testPanicServiceDesc, struct{}{})
	go func() { _ = s.Start() }()
	t.Cleanup(func() { _ = s.Stop() })
	waitServerAddr(t, s)

	reg := newRegister(t, svc, s.Addr())
	t.Cleanup(func() { _ = reg.Close() })
	if _, err := sharedDiscovery(); err != nil {
		t.Fatalf("NewDiscovery: %v", err)
	}

	c, err := rpc.NewClient("grpc://"+svc, rpc.WithClientOption(grpc.WithDefaultCallOptions(grpc.ForceCodec(rawCodec{}))))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })

	eventually(t, 10*time.Second, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var reply []byte
		err := c.Conn().Invoke(ctx, "/test.Panic/Panic", []byte("boom"), &reply)
		if status.Code(err) != codes.Internal {
			return fmt.Errorf("want codes.Internal, got %v (err=%v)", status.Code(err), err)
		}
		return nil
	})
}

// TestEndToEndDiscoverCallAndFailover 真实 gRPC 服务 + grpc:// 发现拨号 + 故障转移。
func TestEndToEndDiscoverCallAndFailover(t *testing.T) {
	const svc = "ping-service"

	s1 := newPingServer(t)
	s2 := newPingServer(t)
	t.Cleanup(func() { _ = s1.Stop() })
	t.Cleanup(func() { _ = s2.Stop() })

	reg1 := newRegister(t, svc, s1.Addr())
	reg2 := newRegister(t, svc, s2.Addr())
	t.Cleanup(func() { _ = reg1.Close() })
	t.Cleanup(func() { _ = reg2.Close() })

	if _, err := sharedDiscovery(); err != nil {
		t.Fatalf("NewDiscovery: %v", err)
	}

	c, err := rpc.NewClient("grpc://"+svc, rpc.WithClientOption(grpc.WithDefaultCallOptions(grpc.ForceCodec(rawCodec{}))))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })

	// 能正常调用
	eventuallyPing(t, c, "pong:hi")

	// 停掉 s1 并注销 → 客户端应 failover 到 s2
	_ = s1.Stop()
	_ = reg1.Close()
	eventuallyPing(t, c, "pong:hi")
}

// TestHealthCheck 服务端自动注册 grpc_health_v1，SetHealth 可切换状态。
func TestHealthCheck(t *testing.T) {
	const svc = "health-service"

	s, err := rpc.NewServer("127.0.0.1:0", "", "")
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	go func() { _ = s.Start() }()
	t.Cleanup(func() { _ = s.Stop() })
	waitServerAddr(t, s)

	reg := newRegister(t, svc, s.Addr())
	t.Cleanup(func() { _ = reg.Close() })
	if _, err := sharedDiscovery(); err != nil {
		t.Fatalf("NewDiscovery: %v", err)
	}

	c, err := rpc.NewClient("grpc://" + svc)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })

	check := func() healthpb.HealthCheckResponse_ServingStatus {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		resp := &healthpb.HealthCheckResponse{}
		err := c.Conn().Invoke(ctx, "/grpc.health.v1.Health/Check", &healthpb.HealthCheckRequest{}, resp)
		if err != nil {
			t.Fatalf("Health/Check: %v", err)
		}
		return resp.Status
	}

	// 默认 SERVING
	eventually(t, 10*time.Second, func() error {
		if st := check(); st != healthpb.HealthCheckResponse_SERVING {
			return fmt.Errorf("status = %v, want SERVING", st)
		}
		return nil
	})

	// 置为 NOT_SERVING 后应反映到健康检查
	s.SetHealth("", healthpb.HealthCheckResponse_NOT_SERVING)
	if st := check(); st != healthpb.HealthCheckResponse_NOT_SERVING {
		t.Fatalf("status = %v, want NOT_SERVING", st)
	}
}
