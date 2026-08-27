package udp

import (
	"bytes"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

// startTestServer 启动测试服务器(:0 自动分配端口)并返回实际地址
func startTestServer(t *testing.T, setup func(s *Server)) (*Server, string) {
	t.Helper()
	s := NewServer()
	if setup != nil {
		setup(s)
	}
	go func() {
		if err := s.Start("127.0.0.1:0"); err != nil {
			t.Errorf("server Start: %v", err)
		}
	}()
	// 等待监听器就绪
	var addr string
	for i := 0; i < 100; i++ {
		if addr = s.Addr(); addr != "" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if addr == "" {
		t.Fatal("服务器未就绪")
	}
	return s, addr
}

// startClient 重试启动客户端(等待服务端就绪, UDP 无连接故拨号恒成功)
func startClient(t *testing.T, addr string) *Client {
	t.Helper()
	c := NewClient()
	if err := c.Start(addr); err != nil {
		t.Fatalf("client Start: %v", err)
	}
	return c
}

func TestServerClient(t *testing.T) {
	s, addr := startTestServer(t, func(s *Server) {
		s.Router(100, func(addr *net.UDPAddr, data []byte) {
			_ = s.Reply(addr, 200, []byte("reply:"+string(data)))
		})
	})
	defer s.Stop()

	c := startClient(t, addr)
	defer c.Stop()
	replyCh := make(chan string, 4)
	c.SetHandler(func(id uint32, data []byte) {
		if id == 200 {
			replyCh <- string(data)
		}
	})

	if err := c.Send(100, []byte("hello")); err != nil {
		t.Fatalf("client Send: %v", err)
	}
	select {
	case reply := <-replyCh:
		if reply != "reply:hello" {
			t.Errorf("reply = %q, want %q", reply, "reply:hello")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("等待回复超时")
	}
}

func TestSetDefaultHandler(t *testing.T) {
	unhandled := make(chan uint32, 4)
	s, addr := startTestServer(t, func(s *Server) {
		s.SetDefaultHandler(func(addr *net.UDPAddr, id uint32, data []byte) {
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

func TestLargePacket(t *testing.T) {
	// 大包(16KB, 远超常见小包)应完整收发
	s, addr := startTestServer(t, func(s *Server) {
		s.Router(100, func(addr *net.UDPAddr, data []byte) {
			_ = s.Reply(addr, 200, data) // 原样回传
		})
	})
	defer s.Stop()

	c := startClient(t, addr)
	defer c.Stop()
	replyCh := make(chan []byte, 4)
	c.SetHandler(func(id uint32, data []byte) {
		if id == 200 {
			replyCh <- data
		}
	})

	payload := bytes.Repeat([]byte("x"), 16*1024)
	if err := c.Send(100, payload); err != nil {
		t.Fatalf("client Send: %v", err)
	}
	select {
	case got := <-replyCh:
		if !bytes.Equal(got, payload) {
			t.Errorf("大包回复长度不符: got %d, want %d", len(got), len(payload))
		}
	case <-time.After(3 * time.Second):
		t.Fatal("等待大包回复超时")
	}
}

func TestServerStop(t *testing.T) {
	s, addr := startTestServer(t, nil)
	if err := s.Stop(); err != nil {
		t.Fatalf("server Stop: %v", err)
	}
	// Stop 幂等
	if err := s.Stop(); err != nil {
		t.Fatalf("server Stop 幂等失败: %v", err)
	}
	// 停止后 Reply 返回错误
	udpAddr, _ := net.ResolveUDPAddr("udp", addr)
	if err := s.Reply(udpAddr, 1, []byte("x")); err == nil {
		t.Error("Stop 后 Reply 应返回错误")
	}
}

// TestRateLimit 限流: 超过单地址每秒上限的消息被丢弃
func TestRateLimit(t *testing.T) {
	s, addr := startTestServer(t, func(s *Server) {
		s.SetRateLimit(2)
		s.Router(100, func(addr *net.UDPAddr, data []byte) {
			_ = s.Reply(addr, 200, []byte("ok"))
		})
	})
	defer s.Stop()

	c := startClient(t, addr)
	defer c.Stop()
	var count atomic.Int32
	c.SetHandler(func(id uint32, data []byte) {
		if id == 200 {
			count.Add(1)
		}
	})

	// 连发 5 条, 限流 2/s, 应只收到 2 条回复
	for i := 0; i < 5; i++ {
		if err := c.Send(100, []byte("x")); err != nil {
			t.Fatalf("client Send: %v", err)
		}
	}
	// 等待回复到达(最多 3 条)或超时
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && count.Load() < 3 {
		time.Sleep(10 * time.Millisecond)
	}
	if n := count.Load(); n != 2 {
		t.Fatalf("收到回复 = %d, want 2(限流生效)", n)
	}
}
