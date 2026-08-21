// TCP 服务端/客户端心跳检测与消息路由演示
//
// 运行方式:
//
//	cd network && go run ./example/tcp            # 自动演示(约 10 秒)
//	cd network && go run ./example/tcp -interactive  # 交互模式, 从标准输入发送消息
//
// 演示内容:
//  1. 消息路由: 登录/聊天消息按 ID 分发, 服务端回复与广播
//  2. 心跳检测: 正常客户端自动发心跳保持存活, 静默连接被服务端超时断开
package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/tcp"
	"github.com/laixhe/gonet/xlog"
)

// 业务消息ID (框架心跳消息 network.MessageIDHeartbeat=1 已占用)
const (
	msgLogin      = 100 // 登录
	msgChat       = 101 // 聊天
	msgUnknown    = 999 // 未注册路由的消息
	msgServerPush = 200 // 服务端单发
	msgBroadcast  = 201 // 服务端广播
)

// sessions 应用层会话表, 记录已登录连接
type sessions struct {
	mu    sync.Mutex
	conns map[int64]network.IConn // 连接ID -> 连接
	names map[int64]string        // 连接ID -> 用户名
}

func newSessions() *sessions {
	return &sessions{
		conns: make(map[int64]network.IConn),
		names: make(map[int64]string),
	}
}

// add 注册会话, name 为空时不覆盖已有用户名
func (s *sessions) add(conn network.IConn, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conns[conn.ID()] = conn
	if name != "" {
		s.names[conn.ID()] = name
	}
}

// remove 移除会话
func (s *sessions) remove(id int64) {
	s.mu.Lock()
	delete(s.conns, id)
	delete(s.names, id)
	s.mu.Unlock()
}

// name 获取连接用户名
func (s *sessions) name(id int64) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n, ok := s.names[id]; ok {
		return n
	}
	return fmt.Sprintf("conn-%d", id)
}

// broadcast 广播消息给所有会话
func (s *sessions) broadcast(id uint32, data []byte) {
	s.mu.Lock()
	conns := make([]network.IConn, 0, len(s.conns))
	for _, c := range s.conns {
		conns = append(conns, c)
	}
	s.mu.Unlock()
	for _, c := range conns {
		_ = c.Send(id, data)
	}
}

// bindRoutes 注册登录/聊天路由
func bindRoutes(s network.IServer, ss *sessions) {
	// 未注册消息的默认处理
	s.SetDefaultHandler(func(conn network.IConn, id uint32, data []byte) {
		// 跳过框架心跳消息, 避免日志噪音
		if id == network.MessageIDHeartbeat {
			return
		}
		network.Log().Warnf("[server] 收到未注册消息 ID=%d, data=%s", id, data)
	})

	// 登录: 注册会话并欢迎
	s.Router(msgLogin, func(conn network.IConn, data []byte) {
		name := string(data)
		ss.add(conn, name)
		_ = conn.Send(msgServerPush, []byte(fmt.Sprintf("欢迎 %s, 你的连接ID=%d", name, conn.ID())))
		network.Log().Infof("[server] %s 登录, 连接ID=%d", name, conn.ID())
	})

	// 聊天: 广播给所有会话
	s.Router(msgChat, func(conn network.IConn, data []byte) {
		msg := fmt.Sprintf("%s: %s", ss.name(conn.ID()), string(data))
		network.Log().Infof("[server] 广播 -> %s", msg)
		ss.broadcast(msgBroadcast, []byte(msg))
	})
}

// runServer 运行服务端
func runServer(s network.IServer, addr string, done <-chan struct{}) {
	ss := newSessions()
	// 连接建立/断开事件: 维护会话表
	s.OnConnect(func(conn network.IConn) {
		ss.add(conn, "")
		network.Log().Infof("[server] 连接建立 ID=%d", conn.ID())
	})
	s.OnDisconnect(func(conn network.IConn) {
		ss.remove(conn.ID())
		network.Log().Infof("[server] 连接断开 ID=%d", conn.ID())
	})
	bindRoutes(s, ss)

	go func() {
		if err := s.Start(addr); err != nil {
			network.Log().Errorf("[server] 已停止: %v", err)
		}
	}()
	network.Log().Infof("[server] 启动监听 %s", addr)

	<-done
	s.Stop()
	network.Log().Infof("[server] 优雅关闭")
}

// runClient 运行一个客户端
func runClient(addr, name string, stop <-chan struct{}) {
	c := tcp.NewClient()
	c.SetHandler(func(id uint32, data []byte) {
		switch id {
		case msgServerPush:
			network.Log().Infof("[%s] 服务端单发: %s", name, data)
		case msgBroadcast:
			network.Log().Infof("[%s] 收到广播: %s", name, data)
		}
	})
	if err := c.Start(addr); err != nil {
		network.Log().Errorf("[%s] 连接失败: %v", name, err)
		return
	}
	defer c.Stop()
	network.Log().Infof("[%s] 已连接, 发送登录", name)
	if err := c.Send(msgLogin, []byte(name)); err != nil {
		network.Log().Errorf("[%s] 登录发送失败: %v", name, err)
	}

	// alice 模拟发送聊天消息和一条未注册路由的消息
	if name == "alice" {
		go func() {
			for _, m := range []string{"大家好", "有人吗?"} {
				time.Sleep(500 * time.Millisecond)
				network.Log().Infof("[alice] 发送聊天: %s", m)
				if err := c.Send(msgChat, []byte(m)); err != nil {
					network.Log().Errorf("[alice] 发送失败: %v", err)
					return
				}
			}
			time.Sleep(500 * time.Millisecond)
			_ = c.Send(msgUnknown, []byte("没人处理我"))
			network.Log().Infof("[alice] 已发送未注册消息 ID=%d, 服务端将丢弃且无回复", msgUnknown)
		}()
	}

	<-stop
	network.Log().Infof("[%s] 断开连接", name)
}

// runSilentConn 建立一条不发任何消息的原始连接, 演示服务端心跳超时断开
func runSilentConn(addr string, stop <-chan struct{}) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		network.Log().Errorf("[silent] 连接失败: %v", err)
		return
	}
	defer conn.Close()
	network.Log().Infof("[silent] 已建立原始连接, 不发任何消息, 等待服务端心跳超时断开...")
	<-stop
}

// runAutoDemo 自动演示
func runAutoDemo(addr string) {
	stopClients := make(chan struct{})
	done := make(chan struct{})

	// 演示可配置化: 自定义最大连接数与分片数
	s := tcp.NewServerWithConfig(tcp.Config{MaxConnections: 100, Partitions: 16})
	go runServer(s, addr, done)
	time.Sleep(200 * time.Millisecond) // 等待服务端就绪

	go runClient(addr, "alice", stopClients)
	go runClient(addr, "bob", stopClients)
	go runSilentConn(addr, stopClients)

	network.Log().Infof("=== 演示开始(运行约 10 秒) ===")
	network.Log().Infof("1. alice/bob 登录并聊天 -> 演示消息路由与广播")
	network.Log().Infof("2. alice/bob 每 2s 自动发送心跳, 连接保持存活")
	network.Log().Infof("3. silent 连接不发消息, 约 5s 后被服务端心跳检测断开")
	time.Sleep(10 * time.Second)

	close(stopClients)
	time.Sleep(300 * time.Millisecond)
	close(done)
	time.Sleep(300 * time.Millisecond)
	network.Log().Infof("=== 演示结束 ===")
}

// runInteractive 交互模式: 从标准输入发送聊天消息
func runInteractive(addr string) {
	done := make(chan struct{})
	s := tcp.NewServer()
	go runServer(s, addr, done)
	defer s.Stop()
	time.Sleep(200 * time.Millisecond)

	c := tcp.NewClient()
	c.SetHandler(func(id uint32, data []byte) {
		switch id {
		case msgBroadcast:
			network.Log().Infof("[me] 收到广播: %s", data)
		case msgServerPush:
			network.Log().Infof("[me] %s", data)
		}
	})
	if err := c.Start(addr); err != nil {
		network.Log().Errorf("客户端连接失败: %v", err)
		return
	}
	defer c.Stop()
	if err := c.Send(msgLogin, []byte("me")); err != nil {
		network.Log().Errorf("登录发送失败: %v", err)
	}

	fmt.Println("交互模式: 输入消息回车发送, 输入 quit 退出")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "quit" || line == "exit" {
			break
		}
		if err := c.Send(msgChat, []byte(line)); err != nil {
			network.Log().Errorf("发送失败: %v", err)
		}
	}
	close(done)
}

func main() {
	interactive := flag.Bool("interactive", false, "交互模式: 从标准输入发送聊天消息")
	flag.Parse()

	// 初始化 xlog 并注入 network 框架日志
	logs, err := xlog.InitSlog(&xlog.Config{Level: xlog.LevelTypeDebug})
	if err != nil {
		fmt.Printf("xlog 初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer logs.Sync()
	network.SetLogger(logs)

	// 调小心跳参数加速演示: 客户端每 2s 发心跳, 服务端 5s 无消息则断开
	tcp.DefaultHeartbeatInterval = 2 * time.Second
	tcp.DefaultHeartbeatTimeout = 5 * time.Second

	addr := "127.0.0.1:9999"
	if *interactive {
		runInteractive(addr)
		return
	}
	runAutoDemo(addr)
}
