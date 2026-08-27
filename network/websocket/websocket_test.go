package websocket

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/packet"
)

const (
	msgEcho  = 100 // 回声请求
	msgReply = 200 // 回声回复
)

// startTestServer 启动 WebSocket 服务器并等待就绪
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

// dialWS 连接 WebSocket 服务器
func dialWS(t *testing.T, addr string) *websocket.Conn {
	t.Helper()
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("websocket dial: %v", err)
	}
	return c
}

// sendPacket 发送一条消息
func sendPacket(t *testing.T, c *websocket.Conn, id uint32, data []byte) {
	t.Helper()
	packData, err := packet.Pack(packet.NewMessage(id, data))
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	if err := c.WriteMessage(websocket.BinaryMessage, packData); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// recvPacket 读取一条消息
func recvPacket(t *testing.T, c *websocket.Conn) (*packet.Message, error) {
	t.Helper()
	_, data, err := c.ReadMessage()
	if err != nil {
		return nil, err
	}
	msg, err := packet.Unpack(data)
	if err != nil {
		t.Fatalf("unpack: %v", err)
	}
	return msg, nil
}

func TestServerClient(t *testing.T) {
	s, addr := startTestServer(t, DefaultConfig())
	defer s.Stop()

	replyCh := make(chan *packet.Message, 4)
	s.Router(msgEcho, func(conn network.IConn, data []byte) {
		_ = conn.Send(msgReply, []byte("reply:"+string(data)))
		replyCh <- packet.NewMessage(0, data) // 记录服务端收到的数据
	})

	c := dialWS(t, addr)
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
	case m := <-replyCh:
		if string(m.Data) != "hello" {
			t.Errorf("server received = %q, want hello", m.Data)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("服务端未收到消息")
	}
}

func TestSetDefaultHandler(t *testing.T) {
	s, addr := startTestServer(t, DefaultConfig())
	defer s.Stop()

	got := make(chan string, 4)
	s.SetDefaultHandler(func(conn network.IConn, id uint32, data []byte) {
		got <- fmt.Sprintf("%d:%s", id, data)
	})

	c := dialWS(t, addr)
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
	s, addr := startTestServer(t, DefaultConfig())
	defer s.Stop()

	uidCh := make(chan int64, 4)
	s.OnConnect(func(conn network.IConn) {
		conn.BindUid(42)
	})
	s.Router(msgEcho, func(conn network.IConn, data []byte) {
		uidCh <- conn.Uid()
	})

	c := dialWS(t, addr)
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

// TestHeartbeatTimeoutKick 服务端心跳超时踢线
// 客户端通过拦截 ping 不回 pong 模拟静默连接
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

	c := dialWS(t, addr)
	defer c.Close()
	// 拦截 ping, 不回 pong
	c.SetPingHandler(func(string) error { return nil })

	select {
	case <-disconnected:
	case <-time.After(3 * time.Second):
		t.Fatal("静默连接未被心跳超时断开")
	}
}

func TestHeartbeatKeepAlive(t *testing.T) {
	cfg := DefaultConfig()
	cfg.HeartbeatInterval = 100 * time.Millisecond
	cfg.HeartbeatTimeout = 400 * time.Millisecond
	s, addr := startTestServer(t, cfg)
	defer s.Stop()

	disconnected := make(chan struct{}, 4)
	s.OnDisconnect(func(conn network.IConn) {
		disconnected <- struct{}{}
	})

	c := dialWS(t, addr)
	defer c.Close()
	// 后台读循环: gorilla 仅在读消息时处理控制帧并自动回 pong, 模拟真实客户端行为
	go func() {
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				return
			}
		}
	}()

	// 存活超过超时时间(400ms)仍未被踢, 说明 pong 保活生效
	time.Sleep(600 * time.Millisecond)
	select {
	case <-disconnected:
		t.Fatal("正常 pong 保活的连接被误踢")
	default:
	}
}

func TestMaxConnections(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxConnections = 1
	s, addr := startTestServer(t, cfg)
	defer s.Stop()

	c1 := dialWS(t, addr)
	defer c1.Close()
	// 第二个连接应被拒绝(连接数上限)
	c2, _, err := websocket.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err == nil {
		c2.Close()
		// 升级成功但可能随后被关闭, 稍等确认
		time.Sleep(200 * time.Millisecond)
		if _, _, err := c2.ReadMessage(); err == nil {
			t.Error("超过最大连接数仍可正常通信")
		}
	}
}

func TestServerStop(t *testing.T) {
	s, addr := startTestServer(t, DefaultConfig())
	c := dialWS(t, addr)
	defer c.Close()

	s.Stop()
	// 停止后连接应被关闭
	c.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, _, err := c.ReadMessage(); err == nil {
		t.Error("服务器停止后连接仍可读")
	}
	// Stop 幂等
	if err := s.Stop(); err != nil {
		t.Errorf("二次 Stop 报错: %v", err)
	}
}

func TestRejectWrongPath(t *testing.T) {
	s, addr := startTestServer(t, DefaultConfig())
	defer s.Stop()

	u := url.URL{Scheme: "ws", Host: addr, Path: "/wrong"}
	_, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err == nil {
		t.Fatal("错误路径仍升级成功")
	}
	if resp != nil && resp.StatusCode != 404 {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

// TestClient 通过 websocket.Client 连接服务器做回声往返
func TestClient(t *testing.T) {
	s, addr := startTestServer(t, DefaultConfig())
	defer s.Stop()

	s.Router(msgEcho, func(conn network.IConn, data []byte) {
		_ = conn.Send(msgReply, []byte("reply:"+string(data)))
	})

	client := NewClient()
	defer client.Stop()
	if err := client.Start(addr); err != nil {
		t.Fatalf("client start: %v", err)
	}
	recv := make(chan *packet.Message, 4)
	client.SetHandler(func(id uint32, data []byte) {
		recv <- packet.NewMessage(id, data)
	})

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
}

// TestClientKeepAlive 客户端 ping 心跳保活, 存活超过超时时间不被踢线
func TestClientKeepAlive(t *testing.T) {
	cfg := DefaultConfig()
	cfg.HeartbeatInterval = 100 * time.Millisecond
	cfg.HeartbeatTimeout = 400 * time.Millisecond
	s, addr := startTestServer(t, cfg)
	defer s.Stop()

	disconnected := make(chan struct{}, 4)
	s.OnDisconnect(func(conn network.IConn) {
		disconnected <- struct{}{}
	})

	client := NewClientWithConfig(cfg)
	defer client.Stop()
	if err := client.Start(addr); err != nil {
		t.Fatalf("client start: %v", err)
	}

	// 客户端定时 ping, 服务端 pong handler 刷新心跳, 存活超过超时时间(400ms)仍不被踢
	time.Sleep(600 * time.Millisecond)
	select {
	case <-disconnected:
		t.Fatal("正常 ping 保活的客户端被误踢")
	default:
	}
}

// newTestCert 生成测试用自签名证书
func newTestCert(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("生成证书失败: %v", err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

// TestServerTLS wss: 服务器启用 TLS, 客户端以 wss 连接并完成回声往返
func TestServerTLS(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TLS = &tls.Config{Certificates: []tls.Certificate{newTestCert(t)}}
	s, addr := startTestServer(t, cfg)
	defer s.Stop()

	s.Router(msgEcho, func(conn network.IConn, data []byte) {
		_ = conn.Send(msgReply, []byte("reply:"+string(data)))
	})

	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	u := url.URL{Scheme: "wss", Host: addr, Path: "/ws"}
	c, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("wss dial: %v", err)
	}
	defer c.Close()

	sendPacket(t, c, msgEcho, []byte("hi"))
	msg, err := recvPacket(t, c)
	if err != nil {
		t.Fatalf("recv: %v", err)
	}
	if msg.ID != msgReply || string(msg.Data) != "reply:hi" {
		t.Errorf("reply = id=%d data=%q", msg.ID, msg.Data)
	}
}

// TestClientReconnect 断线重连: server1 停止后客户端自动重连到同地址的 server2
func TestClientReconnect(t *testing.T) {
	// 获取固定端口供两代服务器复用
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	s1Conn := make(chan struct{}, 4)
	s2Conn := make(chan struct{}, 4)

	// 第一台服务器
	s1 := NewServerWithConfig(DefaultConfig())
	s1.OnConnect(func(conn network.IConn) {
		select {
		case s1Conn <- struct{}{}:
		default:
		}
	})
	go func() {
		_ = s1.Start(addr)
	}()

	// 客户端: 短重连间隔, 需在 Start 前设置
	client := NewClient().(*Client)
	client.SetReconnectInterval(50 * time.Millisecond)
	for i := 0; i < 50; i++ {
		if err := client.Start(addr); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	defer client.Stop()
	select {
	case <-s1Conn:
	case <-time.After(3 * time.Second):
		t.Fatal("连接 server1 失败")
	}

	// 停止 server1, 触发客户端断线重连
	_ = s1.Stop()

	// 同地址启动 server2
	s2 := NewServerWithConfig(DefaultConfig())
	s2.OnConnect(func(conn network.IConn) {
		select {
		case s2Conn <- struct{}{}:
		default:
		}
	})
	go func() {
		_ = s2.Start(addr)
	}()
	defer s2.Stop()

	// 客户端应自动重连到 server2
	select {
	case <-s2Conn:
	case <-time.After(3 * time.Second):
		t.Fatal("客户端未自动重连到 server2")
	}
}
