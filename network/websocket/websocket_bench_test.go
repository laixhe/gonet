package websocket

import (
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/packet"
)

// BenchmarkEcho 回声往返性能: 客户端发送一条消息并等待回复
func BenchmarkEcho(b *testing.B) {
	s := NewServer()
	s.Router(100, func(conn network.IConn, data []byte) {
		_ = conn.Send(200, data)
	})
	// 获取可用端口(:0 自动分配)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	go func() {
		_ = s.Start(addr)
	}()
	defer s.Stop()

	// 等待监听器就绪
	for i := 0; i < 100; i++ {
		if s.Addr() != "" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		b.Fatalf("websocket dial: %v", err)
	}
	defer conn.Close()

	packData, err := packet.Pack(packet.NewMessage(100, []byte("hello")))
	if err != nil {
		b.Fatalf("pack: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := conn.WriteMessage(websocket.BinaryMessage, packData); err != nil {
			b.Fatalf("write: %v", err)
		}
		if _, _, err := conn.ReadMessage(); err != nil {
			b.Fatalf("read: %v", err)
		}
	}
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "roundtrips/s")
}
