package tcp

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/packet"
)

func TestConnectionHooks(t *testing.T) {
	connected := make(chan int64, 4)
	disconnected := make(chan int64, 4)
	s, addr := startTestServer(t, func(s network.IServer) {
		s.OnConnect(func(conn network.IConn) {
			connected <- conn.ID()
		})
		s.OnDisconnect(func(conn network.IConn) {
			disconnected <- conn.ID()
		})
	})
	defer s.Stop()

	c := startClient(t, addr)

	select {
	case <-connected:
	case <-time.After(3 * time.Second):
		t.Fatal("未触发 OnConnect")
	}
	if err := c.Stop(); err != nil {
		t.Fatalf("client Stop: %v", err)
	}
	select {
	case <-disconnected:
	case <-time.After(3 * time.Second):
		t.Fatal("未触发 OnDisconnect")
	}
}

func TestClientReconnect(t *testing.T) {
	// 事件通道: 用服务端钩子确认断开/重连完成, 避免测试自身与重连时序竞争
	s1Disc := make(chan struct{}, 1)
	s2Conn := make(chan struct{}, 1)

	// 第一台服务器
	s1, addr := startTestServer(t, func(s network.IServer) {
		s.OnDisconnect(func(conn network.IConn) {
			t.Logf("[s1] 连接断开 ID=%d", conn.ID())
			select {
			case s1Disc <- struct{}{}:
			default:
			}
		})
		s.Router(100, func(conn network.IConn, data []byte) {
			_ = conn.Send(200, []byte("reply1:"+string(data)))
		})
	})

	// 客户端: 设置短重连间隔, 需在 Start 前
	c := NewClient().(*Client)
	c.SetReconnectInterval(50 * time.Millisecond)
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
	replyCh := make(chan string, 8)
	c.SetHandler(func(id uint32, data []byte) {
		if id == 200 {
			replyCh <- string(data)
		}
	})

	// 第一轮通信
	if err := c.Send(100, []byte("one")); err != nil {
		t.Fatalf("client Send: %v", err)
	}
	if reply := waitReply(t, replyCh); reply != "reply1:one" {
		t.Errorf("reply = %q, want %q (第一轮应由 server1 回复)", reply, "reply1:one")
	}

	// 关闭 server1, 触发客户端断线重连
	_ = s1.Stop()
	// 等待 server1 断开事件, 确认旧连接已关闭
	select {
	case <-s1Disc:
	case <-time.After(3 * time.Second):
		t.Fatal("server1 未触发连接断开")
	}

	// 同地址启动第二台服务器
	s2 := NewServer()
	s2.OnConnect(func(conn network.IConn) {
		t.Logf("[s2] 连接建立 ID=%d (客户端已重连)", conn.ID())
		select {
		case s2Conn <- struct{}{}:
		default:
		}
	})
	s2.Router(100, func(conn network.IConn, data []byte) {
		_ = conn.Send(200, []byte("reply2:"+string(data)))
	})
	go func() {
		if err := s2.Start(addr); err != nil {
			t.Errorf("server2 Start: %v", err)
		}
	}()
	defer s2.Stop()

	// 等待客户端自动重连到 server2
	select {
	case <-s2Conn:
	case <-time.After(3 * time.Second):
		t.Fatal("客户端未重连到 server2")
	}

	// 第二轮通信: 此时已重连, 消息必然到达 server2
	if err := c.Send(100, []byte("two")); err != nil {
		t.Fatalf("client Send: %v", err)
	}
	if reply := waitReply(t, replyCh); reply != "reply2:two" {
		t.Errorf("reply = %q, want %q (重连后应由 server2 回复)", reply, "reply2:two")
	}
}

func TestServerRejectLargeMessage(t *testing.T) {
	old := packet.MaxMessageLen
	packet.MaxMessageLen = 1024
	defer func() { packet.MaxMessageLen = old }()

	disconnected := make(chan struct{})
	s, addr := startTestServer(t, func(s network.IServer) {
		s.OnDisconnect(func(conn network.IConn) {
			close(disconnected)
		})
	})
	defer s.Stop()

	conn, err := dialRetry(addr, 50)
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	// 构造 DataLen 超限的数据包
	header := make([]byte, 8)
	binary.BigEndian.PutUint32(header[0:4], 1)
	binary.BigEndian.PutUint32(header[4:8], packet.MaxMessageLen+1)
	if _, err := conn.Write(header); err != nil {
		t.Fatalf("写入失败: %v", err)
	}

	// 服务端应拒绝并断开该连接, 不因超大包而崩溃
	select {
	case <-disconnected:
	case <-time.After(3 * time.Second):
		t.Fatal("超长消息未导致连接断开")
	}
}

// waitReply 等待并返回回复
func waitReply(t *testing.T, ch chan string) string {
	t.Helper()
	select {
	case r := <-ch:
		return r
	case <-time.After(3 * time.Second):
		t.Fatal("等待回复超时")
		return ""
	}
}

// TestGracefulDrain 平滑关闭: Stop 时等待正在处理的消息执行完成
func TestGracefulDrain(t *testing.T) {
	done := make(chan string, 4)
	s, addr := startTestServer(t, func(s network.IServer) {
		s.Router(100, func(conn network.IConn, data []byte) {
			time.Sleep(300 * time.Millisecond) // 模拟慢 handler
			done <- string(data)
		})
	})

	conn, err := dialRetry(addr, 50)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	packData, err := packet.Pack(packet.NewMessage(100, []byte("drain")))
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	if _, err := conn.Write(packData); err != nil {
		t.Fatalf("write: %v", err)
	}
	// 消息已入队并开始处理后再关闭
	time.Sleep(100 * time.Millisecond)

	// Stop 同步等待正在处理的消息排空
	if err := s.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	select {
	case msg := <-done:
		if msg != "drain" {
			t.Errorf("done = %q, want drain", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Stop 后正在处理的消息未执行完成")
	}
}
