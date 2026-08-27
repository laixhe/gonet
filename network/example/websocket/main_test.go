package main

import (
	"testing"
	"time"

	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/websocket"
)

// waitServerAddr 等待服务器就绪并返回实际监听地址(:0 自动分配)
func waitServerAddr(t *testing.T, s network.IServer) string {
	t.Helper()
	for i := 0; i < 100; i++ {
		if addr := s.Addr(); addr != "" {
			return addr
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("服务器未就绪")
	return ""
}

// TestEchoServerClient 端到端: WebSocket 服务器 + 客户端回声往返
func TestEchoServerClient(t *testing.T) {
	s := websocket.NewServer()
	go func() {
		_ = s.Start("127.0.0.1:0")
	}()
	addr := waitServerAddr(t, s)
	defer s.Stop()

	s.Router(msgEcho, func(conn network.IConn, data []byte) {
		_ = conn.Send(msgReply, []byte("echo:"+string(data)))
	})

	client := websocket.NewClient()
	defer client.Stop()
	if err := client.Start(addr); err != nil {
		t.Fatalf("client start: %v", err)
	}
	recv := make(chan string, 4)
	client.SetHandler(func(id uint32, data []byte) {
		recv <- string(data)
	})

	if err := client.Send(msgEcho, []byte("你好")); err != nil {
		t.Fatalf("send: %v", err)
	}
	select {
	case r := <-recv:
		if r != "echo:你好" {
			t.Errorf("reply = %q, want %q", r, "echo:你好")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("客户端未收到回复")
	}
}
