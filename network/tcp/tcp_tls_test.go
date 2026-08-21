package tcp

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/laixhe/gonet/network"
)

// newTestTLSConfig 生成测试用自签 TLS 配置, 返回服务端/客户端配置
func newTestTLSConfig(t *testing.T) (*tls.Config, *tls.Config) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("生成私钥失败: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("生成证书失败: %v", err)
	}
	cert := tls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv}
	return &tls.Config{Certificates: []tls.Certificate{cert}}, &tls.Config{InsecureSkipVerify: true}
}

// probeAddr 探测一个可用端口
func probeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("获取可用端口失败: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

func TestServerClientTLS(t *testing.T) {
	serverCfg, clientCfg := newTestTLSConfig(t)
	addr := probeAddr(t)

	s := NewServerWithConfig(Config{TLS: serverCfg})
	s.Router(100, func(conn network.IConn, data []byte) {
		_ = conn.Send(200, []byte("tls:"+string(data)))
	})
	go func() {
		if err := s.Start(addr); err != nil {
			t.Errorf("server Start: %v", err)
		}
	}()
	defer s.Stop()

	c := NewClient().(*Client)
	c.SetTLS(clientCfg)
	// 等待服务器就绪
	var err error
	for i := 0; i < 50; i++ {
		if err = c.Start(addr); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("client Start: %v", err)
	}
	defer c.Stop()

	replyCh := make(chan string, 1)
	c.SetHandler(func(id uint32, data []byte) {
		if id == 200 {
			replyCh <- string(data)
		}
	})
	if err := c.Send(100, []byte("hi")); err != nil {
		t.Fatalf("client Send: %v", err)
	}

	select {
	case reply := <-replyCh:
		if reply != "tls:hi" {
			t.Errorf("reply = %q, want %q", reply, "tls:hi")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("TLS 通信超时")
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := Config{}
	cfg.Check()
	if cfg.MaxConnections != 1000 || cfg.Partitions != 100 {
		t.Errorf("默认配置补全错误: %+v", cfg)
	}
	cfg2 := Config{MaxConnections: 10, Partitions: 4}
	cfg2.Check()
	if cfg2.MaxConnections != 10 || cfg2.Partitions != 4 {
		t.Errorf("自定义配置错误: %+v", cfg2)
	}
}

func TestMaxConnections(t *testing.T) {
	addr := probeAddr(t)
	connCh := make(chan network.IConn, 4)

	s := NewServerWithConfig(Config{MaxConnections: 1})
	s.OnConnect(func(conn network.IConn) {
		connCh <- conn
	})
	go func() {
		if err := s.Start(addr); err != nil {
			t.Errorf("server Start: %v", err)
		}
	}()
	defer s.Stop()

	// 第一个连接被接受
	c1 := startClient(t, addr)
	defer c1.Stop()
	select {
	case <-connCh:
	case <-time.After(3 * time.Second):
		t.Fatal("第一个连接未被接受")
	}

	// 第二个连接应被服务端拒绝并关闭
	c2 := NewClient().(*Client)
	c2.SetReconnect(false)
	if err := c2.Start(addr); err != nil {
		t.Fatalf("client2 Start: %v", err)
	}
	defer c2.Stop()

	// 服务端不应触发第二次 OnConnect
	select {
	case <-connCh:
		t.Error("超过上限的连接不应被接受")
	case <-time.After(300 * time.Millisecond):
	}
}

func TestSetDefaultHandler(t *testing.T) {
	unhandled := make(chan uint32, 4)
	s, addr := startTestServer(t, func(s network.IServer) {
		s.SetDefaultHandler(func(conn network.IConn, id uint32, data []byte) {
			unhandled <- id
		})
	})
	defer s.Stop()

	c := startClient(t, addr)
	defer c.Stop()
	if err := c.Send(999, []byte("x")); err != nil {
		t.Fatalf("client Send: %v", err)
	}
	select {
	case id := <-unhandled:
		if id != 999 {
			t.Errorf("未注册消息 ID = %d, want 999", id)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("默认处理器未被调用")
	}
}

func TestHooksRuntimeRegistration(t *testing.T) {
	connected := make(chan struct{}, 4)
	s, addr := startTestServer(t, nil)
	// 运行中注册 OnConnect
	s.OnConnect(func(conn network.IConn) {
		connected <- struct{}{}
	})

	c := startClient(t, addr)
	defer c.Stop()
	select {
	case <-connected:
	case <-time.After(3 * time.Second):
		t.Fatal("运行中注册的 OnConnect 未触发")
	}
	defer s.Stop()
}

func TestHeartbeatConfig(t *testing.T) {
	cfg := Config{HeartbeatInterval: 2 * time.Second, HeartbeatTimeout: 5 * time.Second}
	cfg.Check()
	if cfg.HeartbeatInterval != 2*time.Second || cfg.HeartbeatTimeout != 5*time.Second {
		t.Errorf("心跳配置错误: %+v", cfg)
	}
	cfg2 := Config{}
	cfg2.Check()
	if cfg2.HeartbeatInterval != DefaultHeartbeatInterval || cfg2.HeartbeatTimeout != DefaultHeartbeatTimeout {
		t.Errorf("心跳默认值错误: %+v", cfg2)
	}
}

func TestHeartbeatTimeoutKick(t *testing.T) {
	// 服务端配置 300ms 心跳超时, 静默连接应被断开
	addr := probeAddr(t)
	disconnected := make(chan struct{}, 1)
	s := NewServerWithConfig(Config{
		HeartbeatInterval: 100 * time.Millisecond,
		HeartbeatTimeout:  300 * time.Millisecond,
	})
	s.OnDisconnect(func(conn network.IConn) {
		disconnected <- struct{}{}
	})
	go func() {
		if err := s.Start(addr); err != nil {
			t.Errorf("server Start: %v", err)
		}
	}()
	defer s.Stop()

	// 静默原始连接, 不发任何消息
	conn, err := dialRetry(addr, 50)
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	select {
	case <-disconnected:
	case <-time.After(2 * time.Second):
		t.Fatal("心跳超时未断开静默连接")
	}
}

// TestProcessWorkers 验证 worker 池: 单 worker 串行, 多 worker 并发
func TestProcessWorkers(t *testing.T) {
	const (
		msgCount  = 3
		handlerMS = 50 * time.Millisecond // 每条消息处理耗时
	)

	run := func(t *testing.T, workers int, expectParallel bool) {
		addr := probeAddr(t)
		s := NewServerWithConfig(Config{ProcessWorkers: workers})
		done := make(chan struct{}, msgCount)
		start := time.Now()
		s.Router(100, func(conn network.IConn, data []byte) {
			time.Sleep(handlerMS)
			done <- struct{}{}
		})
		go func() {
			if err := s.Start(addr); err != nil {
				t.Errorf("server Start: %v", err)
			}
		}()
		defer s.Stop()

		c := startClient(t, addr)
		defer c.Stop()
		for i := 0; i < msgCount; i++ {
			if err := c.Send(100, []byte("x")); err != nil {
				t.Fatalf("client Send: %v", err)
			}
		}
		for i := 0; i < msgCount; i++ {
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				t.Fatal("消息处理超时")
			}
		}
		elapsed := time.Since(start)
		// 串行约 150ms, 并行约 50ms; 以 140ms 为分界
		if expectParallel && elapsed >= 140*time.Millisecond {
			t.Errorf("多 worker 并行处理耗时 %v, 应 < 140ms", elapsed)
		}
		if !expectParallel && elapsed < 140*time.Millisecond {
			t.Errorf("单 worker 串行处理耗时 %v, 应 >= 140ms", elapsed)
		}
	}

	t.Run("单worker串行", func(t *testing.T) { run(t, 1, false) })
	t.Run("多worker并行", func(t *testing.T) { run(t, 3, true) })
}
