package tcp

import (
	"sync"
	"testing"
	"time"

	"github.com/laixhe/gonet/network"
)

// TestConcurrentClients 并发客户端同时收发消息
func TestConcurrentClients(t *testing.T) {
	const (
		clientCount = 20 // 并发客户端数
		msgCount    = 20 // 每客户端消息数
	)

	s, addr := startTestServer(t, func(s network.IServer) {
		s.Router(100, func(conn network.IConn, data []byte) {
			_ = conn.Send(200, data) // 原样回复
		})
	})
	defer s.Stop()

	var wg sync.WaitGroup
	for i := 0; i < clientCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c := startClient(t, addr)
			defer c.Stop()

			replyCh := make(chan []byte, msgCount)
			c.SetHandler(func(id uint32, data []byte) {
				if id == 200 {
					replyCh <- data
				}
			})

			for j := 0; j < msgCount; j++ {
				msg := []byte{byte(id), byte(j)}
				if err := c.Send(100, msg); err != nil {
					t.Errorf("client %d Send 失败: %v", id, err)
					return
				}
				select {
				case got := <-replyCh:
					if string(got) != string(msg) {
						t.Errorf("client %d 回复不符: %q != %q", id, got, msg)
					}
				case <-time.After(3 * time.Second):
					t.Errorf("client %d 等待回复超时", id)
					return
				}
			}
		}(i)
	}
	wg.Wait()
}

// TestConcurrentSendStop 发送消息的同时并发停止连接
func TestConcurrentSendStop(t *testing.T) {
	s, addr := startTestServer(t, func(s network.IServer) {
		s.Router(100, func(conn network.IConn, data []byte) {
			_ = conn.Send(200, data)
		})
	})
	defer s.Stop()

	c := startClient(t, addr)
	done := make(chan struct{})

	// 发送协程: 持续发送直到停止
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				_ = c.Send(100, []byte("ping"))
			}
		}
	}()

	// 并发调用 Stop, 验证幂等且不 panic
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.Stop()
		}()
	}
	wg.Wait()
	close(done)
}

// TestConcurrentServerStop 客户端活跃时并发停止服务器
func TestConcurrentServerStop(t *testing.T) {
	s, addr := startTestServer(t, func(s network.IServer) {
		s.Router(100, func(conn network.IConn, data []byte) {
			_ = conn.Send(200, data)
		})
	})

	var wg sync.WaitGroup
	ready := make(chan struct{}, 10)
	// 10 个活跃客户端持续收发
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := startClient(t, addr)
			defer c.Stop()
			c.SetHandler(func(id uint32, data []byte) {})
			ready <- struct{}{}
			for j := 0; j < 50; j++ {
				_ = c.Send(100, []byte("ping"))
				time.Sleep(time.Millisecond)
			}
		}()
	}
	// 等全部客户端连上后再并发停止, 避免客户端在服务器停止后仍重试连接
	for i := 0; i < 10; i++ {
		<-ready
	}
	_ = s.Stop()
	wg.Wait()
}
