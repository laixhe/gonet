package tcp

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/packet"
)

const (
	msgAll     = 300 // 全员广播
	msgExclude = 301 // 排除广播
)

// TestManagerQueryAndBroadcast 连接管理查询(Count/FindByUid/KickByUid)与广播(全员/排除)
func TestManagerQueryAndBroadcast(t *testing.T) {
	connected := make(chan struct{}, 4)
	disconnected := make(chan struct{}, 4)
	var uidSeq int64
	s, addr := startTestServer(t, func(s network.IServer) {
		s.OnConnect(func(conn network.IConn) {
			if conn.RemoteAddr() == "" {
				t.Errorf("RemoteAddr 为空")
			}
			conn.BindUid(atomic.AddInt64(&uidSeq, 1) + 100) // 101, 102
			connected <- struct{}{}
		})
		s.OnDisconnect(func(conn network.IConn) {
			disconnected <- struct{}{}
		})
	})
	defer s.Stop()

	c1 := startClient(t, addr)
	defer c1.Stop()
	c2 := startClient(t, addr)
	defer c2.Stop()

	// 等待两个连接建立并绑定 Uid
	for i := 0; i < 2; i++ {
		select {
		case <-connected:
		case <-time.After(3 * time.Second):
			t.Fatal("连接未建立")
		}
	}

	m := s.GetManager()
	if m.Count() != 2 {
		t.Fatalf("Count = %d, want 2", m.Count())
	}
	// 按 Uid 查找
	c1Conn := m.FindByUid(101)
	if c1Conn == nil || c1Conn.Uid() != 101 {
		t.Fatalf("FindByUid(101) = %v, want Uid 101", c1Conn)
	}
	if m.FindByUid(999) != nil {
		t.Error("FindByUid 未绑定的 Uid 应返回 nil")
	}
	// 按 ID 查找
	if got := m.FindByID(c1Conn.ID()); got != c1Conn {
		t.Error("FindByID 应返回同一连接")
	}
	if got := m.FindByID(99999); got != nil {
		t.Errorf("FindByID 不存在的 ID 应返回 nil, got ID=%d Uid=%d", got.ID(), got.Uid())
	}

	// 收集两端消息
	c1Got := make(chan *packet.Message, 8)
	c1.SetHandler(func(id uint32, data []byte) { c1Got <- packet.NewMessage(id, data) })
	c2Got := make(chan *packet.Message, 8)
	c2.SetHandler(func(id uint32, data []byte) { c2Got <- packet.NewMessage(id, data) })

	// 全员广播: 两端都应收到
	s.Broadcast(msgAll, []byte("all"))
	if msg := <-c1Got; msg.ID != msgAll {
		t.Errorf("c1 收到 id=%d, want %d", msg.ID, msgAll)
	}
	if msg := <-c2Got; msg.ID != msgAll {
		t.Errorf("c2 收到 id=%d, want %d", msg.ID, msgAll)
	}

	// 排除广播: 排除 Uid 101, 只有 c2(Uid 102) 收到
	s.BroadcastExclude(msgExclude, []byte("to-bob"), 101)
	if msg := <-c2Got; msg.ID != msgExclude || string(msg.Data) != "to-bob" {
		t.Errorf("c2 收到 id=%d data=%q", msg.ID, msg.Data)
	}
	select {
	case msg := <-c1Got:
		t.Fatalf("c1 收到排除广播 id=%d", msg.ID)
	case <-time.After(200 * time.Millisecond):
	}

	// 踢下线 Uid 102
	m.KickByUid(102)
	select {
	case <-disconnected:
	case <-time.After(3 * time.Second):
		t.Fatal("KickByUid 未触发断开")
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && m.Count() != 1 {
		time.Sleep(10 * time.Millisecond)
	}
	if m.Count() != 1 {
		t.Fatalf("踢线后 Count = %d, want 1", m.Count())
	}

	// 按 ID 踢掉剩余连接(Uid 101)
	m.KickByID(c1Conn.ID())
	select {
	case <-disconnected:
	case <-time.After(3 * time.Second):
		t.Fatal("KickByID 未触发断开")
	}
	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && m.Count() != 0 {
		time.Sleep(10 * time.Millisecond)
	}
	if m.Count() != 0 {
		t.Fatalf("KickByID 后 Count = %d, want 0", m.Count())
	}
}
