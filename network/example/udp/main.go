// UDP 服务器/客户端消息路由演示
//
// 运行方式:
//
//	cd network && go run ./example/udp
//
// 演示内容: 服务端注册回声路由, 客户端发送消息并接收回复。
// 注意: UDP 无连接且不可靠, 极端情况下可能丢包。
package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/udp"
	"github.com/laixhe/gonet/xlog"
)

// 消息ID
const (
	msgEcho  = 100 // 回声请求
	msgReply = 200 // 回声回复
)

// newEchoServer 创建回声服务器, 返回服务器与监听地址(可能为 127.0.0.1:0 自动分配)
func newEchoServer(addr string) (*udp.Server, string) {
	s := udp.NewServer()
	s.Router(msgEcho, func(addr *net.UDPAddr, data []byte) {
		network.Log().Infof("[server] 收到来自 %s: %s", addr, data)
		_ = s.Reply(addr, msgReply, []byte("echo:"+string(data)))
	})
	go func() {
		if err := s.Start(addr); err != nil {
			network.Log().Errorf("[server] 已停止: %v", err)
		}
	}()
	// 等待监听器就绪
	var serverAddr string
	for i := 0; i < 100; i++ {
		if serverAddr = s.Addr(); serverAddr != "" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	return s, serverAddr
}

func main() {
	logs, err := xlog.InitSlog(&xlog.Config{Level: xlog.LevelTypeDebug})
	if err != nil {
		fmt.Printf("xlog 初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer logs.Sync()
	network.SetLogger(logs)

	// 服务端: :0 自动分配端口
	s, serverAddr := newEchoServer("127.0.0.1:0")
	network.Log().Infof("[server] 启动监听 %s", serverAddr)
	defer s.Stop()

	// 客户端
	c := udp.NewClient()
	done := make(chan struct{}, 3)
	c.SetHandler(func(id uint32, data []byte) {
		network.Log().Infof("[client] 收到回复: %s", data)
		done <- struct{}{}
	})
	if err := c.Start(serverAddr); err != nil {
		network.Log().Errorf("[client] 连接失败: %v", err)
		return
	}
	defer c.Stop()
	network.Log().Infof("[client] 已连接 %s", serverAddr)

	// 发送 3 条消息
	for _, m := range []string{"你好", "UDP", "再见"} {
		network.Log().Infof("[client] 发送: %s", m)
		if err := c.Send(msgEcho, []byte(m)); err != nil {
			network.Log().Errorf("[client] 发送失败: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// 等待回复
	timeout := time.After(3 * time.Second)
	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-timeout:
			network.Log().Warnf("[demo] 等待回复超时(UDP 可能丢包)")
			goto end
		}
	}
end:
	network.Log().Infof("[demo] UDP 演示结束")
}
