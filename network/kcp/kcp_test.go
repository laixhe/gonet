package kcp

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/packet"
	kcpv5 "github.com/xtaci/kcp-go/v5"
)

const (
	msgEcho  = 100 // 回声请求
	msgReply = 200 // 回声回复
)

// startTestServer 启动 KCP 服务器并等待就绪
func startTestServer(t *testing.T, config Config) (network.IServer, string) {
	t.Helper()
	s := NewServerWithConfig(config)
	go func() {
		_ = s.Start("127.0.0.1:0")
	}()
	for i := 0; i < 100; i++ {
		if addr := s.Addr(); addr != "" {
			return s, addr
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("服务器未就绪")
	return nil, ""
}

// dialKCP 连接 KCP 服务器(FEC/加密参数与服务器一致)
func dialKCP(t *testing.T, addr string, cfg Config) net.Conn {
	t.Helper()
	c, err := kcpv5.DialWithOptions(addr, cfg.Block, cfg.DataShards, cfg.ParityShards)
	if err != nil {
		t.Fatalf("kcp dial: %v", err)
	}
	return c
}

// sendPacket 发送一条消息
func sendPacket(t *testing.T, c net.Conn, id uint32, data []byte) {
	t.Helper()
	packData, err := packet.Pack(packet.NewMessage(id, data))
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	if _, err := c.Write(packData); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// recvPacket 读取一条消息
func recvPacket(t *testing.T, c net.Conn) (*packet.Message, error) {
	t.Helper()
	msg, err := packet.TcpBufRead(bufio.NewReader(c))
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func TestServerClient(t *testing.T) {
	cfg := DefaultConfig()
	s, addr := startTestServer(t, cfg)
	defer s.Stop()

	got := make(chan string, 4)
	s.Router(msgEcho, func(conn network.IConn, data []byte) {
		got <- string(data)
		_ = conn.Send(msgReply, []byte("reply:"+string(data)))
	})

	c := dialKCP(t, addr, cfg)
	defer c.Close()

	// 收发往返
	sendPacket(t, c, msgEcho, []byte("hello"))
	msg, err := recvPacket(t, c)
	if err != nil {
		t.Fatalf("recv: %v", err)
	}
	if msg.ID != msgReply || string(msg.Data) != "reply:hello" {
		t.Errorf("reply = id=%d data=%q, want id=%d data=%q", msg.ID, msg.Data, msgReply, "reply:hello")
	}
	// 服务端确实收到
	select {
	case g := <-got:
		if g != "hello" {
			t.Errorf("server received = %q, want hello", g)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("服务端未收到消息")
	}
}

func TestSetDefaultHandler(t *testing.T) {
	cfg := DefaultConfig()
	s, addr := startTestServer(t, cfg)
	defer s.Stop()

	got := make(chan string, 4)
	s.SetDefaultHandler(func(conn network.IConn, id uint32, data []byte) {
		got <- fmt.Sprintf("%d:%s", id, data)
	})

	c := dialKCP(t, addr, cfg)
	defer c.Close()

	sendPacket(t, c, 999, []byte("unknown"))
	select {
	case g := <-got:
		if g != "999:unknown" {
			t.Errorf("default handler got %q, want %q", g, "999:unknown")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("默认处理器未触发")
	}
}

func TestUidBind(t *testing.T) {
	cfg := DefaultConfig()
	s, addr := startTestServer(t, cfg)
	defer s.Stop()

	uidCh := make(chan int64, 4)
	s.OnConnect(func(conn network.IConn) {
		conn.BindUid(42)
	})
	s.Router(msgEcho, func(conn network.IConn, data []byte) {
		uidCh <- conn.Uid()
	})

	c := dialKCP(t, addr, cfg)
	defer c.Close()
	sendPacket(t, c, msgEcho, []byte("x"))

	select {
	case uid := <-uidCh:
		if uid != 42 {
			t.Errorf("uid = %d, want 42", uid)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("未触发路由")
	}
}

// TestHeartbeatTimeoutKick KCP 无 TCP 的 RST 断线机制, 心跳检测尤为重要
func TestHeartbeatTimeoutKick(t *testing.T) {
	cfg := DefaultConfig()
	cfg.HeartbeatInterval = 100 * time.Millisecond
	cfg.HeartbeatTimeout = 300 * time.Millisecond
	s, addr := startTestServer(t, cfg)
	defer s.Stop()

	disconnected := make(chan struct{}, 4)
	s.OnDisconnect(func(conn network.IConn) {
		disconnected <- struct{}{}
	})

	c := dialKCP(t, addr, cfg)
	defer c.Close()
	// 先发一条消息触发会话建立(kcp-go 懒握手: 服务端收到首个数据包才建会话),
	// 之后静默, 服务端应心跳超时断开
	sendPacket(t, c, 999, []byte("bootstrap"))
	select {
	case <-disconnected:
	case <-time.After(3 * time.Second):
		t.Fatal("静默连接未被心跳超时断开")
	}
}

func TestServerStop(t *testing.T) {
	cfg := DefaultConfig()
	s, addr := startTestServer(t, cfg)
	c := dialKCP(t, addr, cfg)
	defer c.Close()

	s.Stop()
	// 停止后连接应被关闭
	c.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1)
	if _, err := c.Read(buf); err == nil {
		t.Error("服务器停止后连接仍可读")
	}
	// Stop 幂等
	if err := s.Stop(); err != nil {
		t.Errorf("二次 Stop 报错: %v", err)
	}
}

func TestMaxConnections(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxConnections = 1
	s, addr := startTestServer(t, cfg)
	defer s.Stop()

	c1 := dialKCP(t, addr, cfg)
	defer c1.Close()

	// 第二个连接: KCP 握手会成功但服务端应拒绝并关闭
	c2, err := kcpv5.DialWithOptions(addr, cfg.Block, cfg.DataShards, cfg.ParityShards)
	if err != nil {
		t.Fatalf("第二个连接 dial: %v", err)
	}
	defer c2.Close()
	// 向 c2 发消息, 不应收到回复(已被关闭或未注册)
	sendPacket(t, c2, msgEcho, []byte("x"))
	c2.SetReadDeadline(time.Now().Add(1 * time.Second))
	if msg, err := recvPacket(t, c2); err == nil {
		t.Errorf("超限连接仍收到回复: id=%d", msg.ID)
	}
}

// TestClient 通过 kcp.Client 连接服务器做回声往返
func TestClient(t *testing.T) {
	cfg := DefaultConfig()
	s, addr := startTestServer(t, cfg)
	defer s.Stop()

	got := make(chan string, 4)
	s.Router(msgEcho, func(conn network.IConn, data []byte) {
		got <- string(data)
		_ = conn.Send(msgReply, []byte("reply:"+string(data)))
	})

	client := NewClientWithConfig(cfg)
	defer client.Stop()
	if err := client.Start(addr); err != nil {
		t.Fatalf("client start: %v", err)
	}
	recv := make(chan *packet.Message, 4)
	client.SetHandler(func(id uint32, data []byte) {
		recv <- packet.NewMessage(id, data)
	})

	// 客户端 Start 已自动发心跳触发懒握手, 直接发业务消息
	if err := client.Send(msgEcho, []byte("hello")); err != nil {
		t.Fatalf("send: %v", err)
	}
	select {
	case msg := <-recv:
		if msg.ID != msgReply || string(msg.Data) != "reply:hello" {
			t.Errorf("reply = id=%d data=%q, want id=%d data=%q", msg.ID, msg.Data, msgReply, "reply:hello")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("客户端未收到回复")
	}
	select {
	case g := <-got:
		if g != "hello" {
			t.Errorf("server received = %q, want hello", g)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("服务端未收到消息")
	}
}
