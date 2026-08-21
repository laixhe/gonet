package main

import (
	"strings"
	"testing"
	"time"

	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/tcp"
)

// mockConn 模拟 IConn, 记录 Send 数据
type mockConn struct {
	id   int64
	sent []byte
}

func (m *mockConn) ID() int64      { return m.id }
func (m *mockConn) Uid() int64     { return 0 }
func (m *mockConn) BindUid(int64)  {}
func (m *mockConn) UnbindUid()     {}
func (m *mockConn) State() int32   { return 0 }
func (m *mockConn) IsClosed() bool { return false }
func (m *mockConn) RemoteAddr() string {
	return "127.0.0.1:0"
}
func (m *mockConn) Stop() error { return nil }
func (m *mockConn) Send(id uint32, data []byte) error {
	m.sent = append(m.sent, data...)
	return nil
}

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

// connectClient 连接并登录, 返回客户端
func connectClient(t *testing.T, addr, name string, handler func(id uint32, data []byte)) network.IClient {
	t.Helper()
	c := tcp.NewClient()
	if handler != nil {
		c.SetHandler(handler)
	}
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
	if err := c.Send(msgLogin, []byte(name)); err != nil {
		t.Fatalf("登录发送失败: %v", err)
	}
	return c
}

func TestSessionsAddRemove(t *testing.T) {
	ss := newSessions()
	ss.add(&mockConn{id: 1}, "alice")
	if name := ss.name(1); name != "alice" {
		t.Errorf("name = %q, want alice", name)
	}
	// 未登录连接显示默认名
	if name := ss.name(99); name != "conn-99" {
		t.Errorf("name = %q, want conn-99", name)
	}
	ss.remove(1)
	if _, ok := ss.conns[1]; ok {
		t.Error("remove 后连接仍存在")
	}
}

func TestSessionsBroadcast(t *testing.T) {
	ss := newSessions()
	c1 := &mockConn{id: 1}
	c2 := &mockConn{id: 2}
	ss.add(c1, "alice")
	ss.add(c2, "bob")
	ss.broadcast(msgBroadcast, []byte("hello"))
	if string(c1.sent) != "hello" || string(c2.sent) != "hello" {
		t.Errorf("广播未到达所有连接: c1=%q c2=%q", c1.sent, c2.sent)
	}
}

func TestServerClientChat(t *testing.T) {
	done := make(chan struct{})
	s := tcp.NewServer()
	go runServer(s, "127.0.0.1:0", done)
	addr := waitServerAddr(t, s)
	defer func() {
		close(done)
		s.Stop()
	}()

	alicePush := make(chan string, 4)
	aliceBroadcast := make(chan string, 4)
	alice := connectClient(t, addr, "alice", func(id uint32, data []byte) {
		switch id {
		case msgServerPush:
			alicePush <- string(data)
		case msgBroadcast:
			aliceBroadcast <- string(data)
		}
	})
	defer alice.Stop()

	bobPush := make(chan string, 4)
	bobBroadcast := make(chan string, 4)
	bob := connectClient(t, addr, "bob", func(id uint32, data []byte) {
		switch id {
		case msgServerPush:
			bobPush <- string(data)
		case msgBroadcast:
			bobBroadcast <- string(data)
		}
	})
	defer bob.Stop()

	// 等 bob 登录完成(收到欢迎)后再发聊天, 避免 bob 错过广播
	select {
	case <-bobPush:
	case <-time.After(3 * time.Second):
		t.Fatal("bob 未完成登录")
	}

	// alice 应收到登录欢迎
	select {
	case msg := <-alicePush:
		if !strings.Contains(msg, "欢迎 alice") {
			t.Errorf("欢迎消息 = %q, 应包含 欢迎 alice", msg)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("未收到登录欢迎")
	}

	// alice 发送聊天, bob 应收到广播
	if err := alice.Send(msgChat, []byte("大家好")); err != nil {
		t.Fatalf("alice Send: %v", err)
	}
	select {
	case msg := <-bobBroadcast:
		if msg != "alice: 大家好" {
			t.Errorf("bob 收到广播 = %q, want %q", msg, "alice: 大家好")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("bob 未收到广播")
	}
}
