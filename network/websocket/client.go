package websocket

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/packet"
)

// Client WebSocket 客户端, 实现 network.IClient
type Client struct {
	mu                sync.Mutex
	conn              *websocket.Conn
	writeMu           sync.Mutex // 写互斥(gorilla 单写限制)
	stop              chan struct{}
	stopOnce          sync.Once
	addr              string        // 目标地址
	reconnect         bool          // 断线自动重连
	reconnectInterval time.Duration // 重连间隔
	path              string        // 连接路径
	heartbeatInterval time.Duration // 心跳发送间隔
	writeTimeout      time.Duration // 写超时时间, 0 不超时
	handler           atomic.Value  // 消息处理, 支持并发设置
}

var _ network.IClient = &Client{}

// NewClient 创建默认配置的 WebSocket 客户端
func NewClient() network.IClient {
	return NewClientWithConfig(DefaultConfig())
}

// NewClientWithConfig 创建指定配置的 WebSocket 客户端, 连接路径与心跳间隔取自配置
func NewClientWithConfig(config Config) network.IClient {
	config.Check()
	return &Client{
		path:              config.Path,
		heartbeatInterval: config.HeartbeatInterval,
		writeTimeout:      config.WriteTimeout,
		reconnect:         true,
		reconnectInterval: time.Second,
	}
}

// SetReconnect 设置断线自动重连
func (c *Client) SetReconnect(reconnect bool) {
	c.reconnect = reconnect
}

// SetReconnectInterval 设置断线重连间隔, 需在 Start 前调用
func (c *Client) SetReconnectInterval(interval time.Duration) {
	c.reconnectInterval = interval
}

// Start 连接并启动, addr 为 host:port 或完整 ws:// 地址(此时 path 不生效)
func (c *Client) Start(addr string) error {
	c.mu.Lock()
	c.stop = make(chan struct{})
	c.addr = addr
	c.mu.Unlock()
	if err := c.dial(addr); err != nil {
		return err
	}
	go c.read()
	go c.heartbeat()
	return nil
}

// dial 建立 WebSocket 连接
func (c *Client) dial(addr string) error {
	wsURL := addr
	if !strings.HasPrefix(addr, "ws://") && !strings.HasPrefix(addr, "wss://") {
		wsURL = "ws://" + addr + c.path
	}
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	return nil
}

// Stop 停止
func (c *Client) Stop() error {
	c.stopOnce.Do(func() {
		c.mu.Lock()
		if c.stop != nil {
			close(c.stop)
		}
		conn := c.conn
		c.mu.Unlock()
		if conn != nil {
			_ = conn.Close()
		}
	})
	return nil
}

// Send 发送消息(WebSocket 二进制帧, 消息体为 packet 协议格式)
func (c *Client) Send(id uint32, data []byte) error {
	packData, err := packet.Pack(packet.NewMessage(id, data))
	if err != nil {
		return err
	}
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return network.ErrConnectionClosed
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	// 写超时保护: 对端不读导致写阻塞时, 超过时限停止
	if c.writeTimeout > 0 {
		_ = conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}
	return conn.WriteMessage(websocket.BinaryMessage, packData)
}

// SetHandler 设置消息处理, 支持在运行中调用
func (c *Client) SetHandler(handler func(id uint32, data []byte)) {
	c.handler.Store(handler)
}

// read 读取消息, gorilla 在读消息时自动响应服务端 ping
func (c *Client) read() {
	for {
		select {
		case <-c.stop:
			return
		default:
		}
		c.mu.Lock()
		conn := c.conn
		c.mu.Unlock()
		if conn == nil {
			return
		}
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			network.Log().Errorf("websocket client read error: %s", err)
			// 置空连接, 断线期间 Send 返回错误
			c.mu.Lock()
			c.conn = nil
			c.mu.Unlock()
			c.reconnectLoop()
			return
		}
		if msgType != websocket.BinaryMessage {
			continue // 忽略文本等其他类型
		}
		msg, err := packet.Unpack(data)
		if err != nil {
			network.Log().Warnf("websocket client unpack error: %s", err)
			continue
		}
		if handler, ok := c.handler.Load().(func(id uint32, data []byte)); ok && handler != nil {
			handler(msg.ID, msg.Data)
		}
	}
}

// reconnectLoop 断线重连
func (c *Client) reconnectLoop() {
	for {
		select {
		case <-c.stop:
			return
		default:
		}
		if !c.reconnect {
			_ = c.Stop()
			return
		}
		if err := c.dial(c.addr); err != nil {
			time.Sleep(c.reconnectInterval)
			continue
		}
		network.Log().Infof("websocket client 重连成功 %s", c.addr)
		go c.read()
		return
	}
}

// heartbeat 心跳发送: 发 ping, 服务端 pong handler 刷新心跳时间
func (c *Client) heartbeat() {
	ticker := time.NewTicker(c.heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			c.mu.Lock()
			conn := c.conn
			c.mu.Unlock()
			if conn == nil {
				continue // 断线重连中, 跳过本轮心跳
			}
			c.writeMu.Lock()
			if c.writeTimeout > 0 {
				_ = conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
			}
			_ = conn.WriteMessage(websocket.PingMessage, nil)
			c.writeMu.Unlock()
		}
	}
}
