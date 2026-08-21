package main

import (
	"testing"
	"time"

	"github.com/laixhe/gonet/network/udp"
)

func TestEchoServerClient(t *testing.T) {
	s, addr := newEchoServer("127.0.0.1:0")
	defer s.Stop()
	if addr == "" {
		t.Fatal("服务器未就绪")
	}

	c := udp.NewClient()
	replyCh := make(chan string, 4)
	c.SetHandler(func(id uint32, data []byte) {
		if id == msgReply {
			replyCh <- string(data)
		}
	})
	if err := c.Start(addr); err != nil {
		t.Fatalf("client Start: %v", err)
	}
	defer c.Stop()

	// 发送消息并验证回声
	for _, m := range []string{"你好", "UDP", "再见"} {
		if err := c.Send(msgEcho, []byte(m)); err != nil {
			t.Fatalf("Send %q 失败: %v", m, err)
		}
		select {
		case reply := <-replyCh:
			if reply != "echo:"+m {
				t.Errorf("reply = %q, want %q", reply, "echo:"+m)
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("等待 %q 回声超时", m)
		}
	}
}
