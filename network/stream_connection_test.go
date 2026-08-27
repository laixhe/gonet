package network

import (
	"net"
	"testing"
	"time"
)

// mockManager 测试用连接管理器实现
type mockManager struct{}

func (m *mockManager) Add(conn IConn) error        { return nil }
func (m *mockManager) Remove(conn IConn)           {}
func (m *mockManager) Close()                      {}
func (m *mockManager) Count() int64                { return 0 }
func (m *mockManager) FindByID(id int64) IConn     { return nil }
func (m *mockManager) FindByUid(uid int64) IConn   { return nil }
func (m *mockManager) KickByID(id int64)           {}
func (m *mockManager) KickByUid(uid int64)         {}
func (m *mockManager) ForEach(fn func(conn IConn)) {}

// TestStreamConnectionWriteTimeout 写超时: 对端不读导致写阻塞, 超过时限断开连接
func TestStreamConnectionWriteTimeout(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	defer clientSide.Close()

	sc := NewStreamConnection(serverSide, &mockManager{}, 1, func(conn IConn, id uint32, data []byte) {},
		time.Second, 5*time.Second, 100*time.Millisecond, 1, "test")
	sc.Start()
	defer sc.Stop()

	// 对端不读(net.Pipe 写需对端读才返回), 写阻塞超过 100ms 应触发连接关闭
	for i := 0; i < 5000; i++ {
		if err := sc.Send(1, []byte("x")); err != nil {
			break // 队列满或连接已关闭
		}
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if sc.IsClosed() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("写超时未触发连接关闭")
}

// TestStreamConnectionGracefulDrain 平滑关闭: Stop 等待正在处理的消息执行完成
func TestStreamConnectionGracefulDrain(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	defer clientSide.Close()

	done := make(chan struct{}, 1)
	sc := NewStreamConnection(serverSide, &mockManager{}, 1, func(conn IConn, id uint32, data []byte) {
		time.Sleep(200 * time.Millisecond) // 模拟慢 handler
		done <- struct{}{}
	}, time.Second, 5*time.Second, 0, 1, "test")
	sc.Start()

	// 客户端发送一条消息(直接写对端)
	clientSide.Write([]byte{0, 0, 0, 1, 0, 0, 0, 1, 120}) // ID=1, Len=1, Data="x"
	time.Sleep(50 * time.Millisecond)                     // 消息已入队并开始处理

	// Stop 同步等待处理完成
	if err := sc.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	// 正在处理的消息应执行完成(未被中断)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop 后正在处理的消息未执行完成")
	}
}
