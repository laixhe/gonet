package tcp

import (
	"net"
	"testing"
	"time"

	"github.com/laixhe/gonet/network"
)

// startTestServer 启动测试服务器并返回地址
func startTestServer(t *testing.T, setup func(s network.IServer)) (network.IServer, string) {
	t.Helper()
	s := NewServer()
	if setup != nil {
		setup(s)
	}
	// 先占用端口获取可用地址
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("获取可用端口失败: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	go func() {
		if err := s.Start(addr); err != nil {
			t.Errorf("server Start: %v", err)
		}
	}()
	return s, addr
}

// dialRetry 重试拨号, 等待服务器监听器就绪
func dialRetry(addr string, attempts int) (net.Conn, error) {
	var err error
	for i := 0; i < attempts; i++ {
		var conn net.Conn
		conn, err = net.Dial("tcp", addr)
		if err == nil {
			return conn, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil, err
}

// startClient 重试启动客户端, 等待服务器监听器就绪
func startClient(t *testing.T, addr string) *Client {
	t.Helper()
	c := NewClient().(*Client)
	for i := 0; i < 50; i++ {
		if err := c.Start(addr); err == nil {
			return c
		} else {
			time.Sleep(10 * time.Millisecond)
		}
	}
	t.Fatalf("client Start 失败: %s", addr)
	return nil
}

func TestServerClient(t *testing.T) {
	s, addr := startTestServer(t, func(s network.IServer) {
		s.Router(100, func(conn network.IConn, data []byte) {
			_ = conn.Send(200, []byte("reply:"+string(data)))
		})
	})
	defer s.Stop()

	c := startClient(t, addr)
	defer c.Stop()
	replyCh := make(chan string, 1)
	c.SetHandler(func(id uint32, data []byte) {
		if id == 200 {
			replyCh <- string(data)
		}
	})

	if err := c.Send(100, []byte("world")); err != nil {
		t.Fatalf("client Send: %v", err)
	}

	select {
	case reply := <-replyCh:
		if reply != "reply:world" {
			t.Errorf("reply = %q, want %q", reply, "reply:world")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("等待回复超时")
	}
}

func TestServerUid(t *testing.T) {
	s, addr := startTestServer(t, func(s network.IServer) {
		s.Router(100, func(conn network.IConn, data []byte) {
			conn.BindUid(12345)
			if conn.Uid() != 12345 {
				t.Errorf("Uid = %d, want 12345", conn.Uid())
			}
			if conn.ID() == 0 {
				t.Error("连接 ID 不应为 0")
			}
			if conn.IsClosed() {
				t.Error("连接不应为关闭状态")
			}
			_ = conn.Send(200, []byte("ok"))
		})
	})
	defer s.Stop()

	c := startClient(t, addr)
	defer c.Stop()
	done := make(chan struct{})
	c.SetHandler(func(id uint32, data []byte) {
		if id == 200 {
			close(done)
		}
	})

	if err := c.Send(100, nil); err != nil {
		t.Fatalf("client Send: %v", err)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("等待回复超时")
	}
}

func TestServerStop(t *testing.T) {
	s, addr := startTestServer(t, nil)
	// 等待服务器就绪
	conn, err := dialRetry(addr, 50)
	if err != nil {
		t.Fatalf("等待服务器就绪失败: %v", err)
	}
	_ = conn.Close()

	if err := s.Stop(); err != nil {
		t.Fatalf("server Stop: %v", err)
	}
	// 监听器已关闭, 再次连接应失败
	if conn, err := net.Dial("tcp", addr); err == nil {
		_ = conn.Close()
		t.Error("Stop 后仍可连接")
	}
}

func TestClientStop(t *testing.T) {
	s, addr := startTestServer(t, nil)
	defer s.Stop()

	c := startClient(t, addr)
	if err := c.Stop(); err != nil {
		t.Fatalf("client Stop: %v", err)
	}
	// Stop 幂等
	if err := c.Stop(); err != nil {
		t.Fatalf("client Stop 幂等失败: %v", err)
	}
}
