// WebSocket 服务器消息路由演示
//
// 运行方式:
//
//	cd network && go run ./example/websocket
//
// 演示内容: 服务端监听 WebSocket 升级, 注册回声路由; gorilla 客户端连接并收发消息。
// 浏览器 H5 也可接入: 连接 ws://127.0.0.1:8888/ws, 发送 packet 二进制协议消息。
package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	gorillaws "github.com/gorilla/websocket"
	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/packet"
	"github.com/laixhe/gonet/network/websocket"
	"github.com/laixhe/gonet/xlog"
)

// 业务消息ID
const (
	msgEcho  = 100 // 回声请求
	msgReply = 200 // 回声回复
)

func main() {
	logs, err := xlog.InitSlog(&xlog.Config{Level: xlog.LevelTypeDebug})
	if err != nil {
		fmt.Printf("xlog 初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer logs.Sync()
	network.SetLogger(logs)

	// 服务端: :0 自动分配端口, 监听升级路径 /ws
	s := websocket.NewServer()
	s.Router(msgEcho, func(conn network.IConn, data []byte) {
		network.Log().Infof("[server] 收到来自 %s: %s", "conn", data)
		_ = conn.Send(msgReply, []byte("echo:"+string(data)))
	})
	go func() {
		if err := s.Start("127.0.0.1:0"); err != nil {
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
	network.Log().Infof("[server] 启动监听 ws://%s/ws", serverAddr)
	defer s.Stop()

	// 客户端(gorilla 模拟 H5 浏览器)
	u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/ws"}
	c, _, err := gorillaws.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		network.Log().Errorf("[client] 连接失败: %v", err)
		return
	}
	defer c.Close()
	network.Log().Infof("[client] 已连接 %s", u.String())

	// 后台读循环(自动回 pong 保活)
	go func() {
		for {
			_, data, err := c.ReadMessage()
			if err != nil {
				return
			}
			msg, err := packet.Unpack(data)
			if err != nil {
				continue
			}
			network.Log().Infof("[client] 收到回复: %s", msg.Data)
		}
	}()

	// 发送 3 条消息
	for _, m := range []string{"你好", "WebSocket", "再见"} {
		network.Log().Infof("[client] 发送: %s", m)
		packData, err := packet.Pack(packet.NewMessage(msgEcho, []byte(m)))
		if err != nil {
			network.Log().Errorf("[client] 打包失败: %v", err)
			return
		}
		if err := c.WriteMessage(gorillaws.BinaryMessage, packData); err != nil {
			network.Log().Errorf("[client] 发送失败: %v", err)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}

	time.Sleep(time.Second)
	network.Log().Infof("[demo] WebSocket 演示结束")
}
