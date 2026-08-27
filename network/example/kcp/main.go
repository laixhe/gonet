// KCP 服务器消息路由演示
//
// 运行方式:
//
//	cd network && go run ./example/kcp
//
// 演示内容: KCP(UDP 可靠传输) 服务端监听, 注册回声路由; kcp 客户端连接并收发消息。
// 注意: kcp-go 为懒握手, 客户端发送第一条消息后服务端才建立会话。
package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/kcp"
	"github.com/laixhe/gonet/network/packet"
	"github.com/laixhe/gonet/xlog"
	kcpv5 "github.com/xtaci/kcp-go/v5"
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

	// 服务端: 默认配置(启用 FEC 10+3 与快速模式), :0 自动分配端口
	cfg := kcp.DefaultConfig()
	s := kcp.NewServerWithConfig(cfg)
	s.Router(msgEcho, func(conn network.IConn, data []byte) {
		network.Log().Infof("[server] 连接ID=%d 收到: %s", conn.ID(), data)
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
	network.Log().Infof("[server] 启动监听 kcp://%s (FEC=%d+%d, NoDelay=%v)", serverAddr, cfg.DataShards, cfg.ParityShards, cfg.NoDelay)
	defer s.Stop()

	// 客户端
	c, err := kcpv5.DialWithOptions(serverAddr, cfg.Block, cfg.DataShards, cfg.ParityShards)
	if err != nil {
		network.Log().Errorf("[client] 连接失败: %v", err)
		return
	}
	defer c.Close()
	network.Log().Infof("[client] 已连接 kcp://%s", serverAddr)

	// 发送 3 条消息
	for _, m := range []string{"你好", "KCP", "再见"} {
		network.Log().Infof("[client] 发送: %s", m)
		packData, err := packet.Pack(packet.NewMessage(msgEcho, []byte(m)))
		if err != nil {
			network.Log().Errorf("[client] 打包失败: %v", err)
			return
		}
		if _, err := c.Write(packData); err != nil {
			network.Log().Errorf("[client] 发送失败: %v", err)
			return
		}
		// 读取回复
		msg, err := packet.TcpBufRead(bufio.NewReader(c))
		if err != nil {
			network.Log().Errorf("[client] 读取失败: %v", err)
			return
		}
		network.Log().Infof("[client] 收到回复: %s", msg.Data)
		time.Sleep(500 * time.Millisecond)
	}

	network.Log().Infof("[demo] KCP 演示结束")
}
