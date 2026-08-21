package tcp

import (
	"bufio"
	"net"
	"testing"

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

	conn, err := dialRetry(addr, 50)
	if err != nil {
		b.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)

	packData, err := packet.Pack(packet.NewMessage(100, []byte("hello")))
	if err != nil {
		b.Fatalf("pack: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := conn.Write(packData); err != nil {
			b.Fatalf("write: %v", err)
		}
		if _, err := packet.TcpBufRead(reader); err != nil {
			b.Fatalf("read: %v", err)
		}
	}
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "roundtrips/s")
}
