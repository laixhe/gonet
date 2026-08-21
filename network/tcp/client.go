package tcp

import (
	"bufio"
	"crypto/tls"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/packet"
)

// Client TCP 客户端
type Client struct {
	mu                sync.Mutex
	conn              net.Conn
	reader            *bufio.Reader
	stop              chan struct{}
	stopOnce          sync.Once
	addr              string
	reconnect         bool          // 断线自动重连
	reconnectInterval time.Duration // 重连间隔
	heartbeatInterval time.Duration // 心跳发送间隔
	tlsConfig         *tls.Config   // TLS 配置
	handler           atomic.Value  // 消息处理, 支持并发设置
}

var _ network.IClient = &Client{}

// NewClient 创建 TCP 客户端
func NewClient() network.IClient {
	return &Client{reconnect: true, reconnectInterval: time.Second, heartbeatInterval: DefaultHeartbeatInterval}
}

// SetReconnect 设置断线自动重连
func (c *Client) SetReconnect(reconnect bool) {
	c.reconnect = reconnect
}

// SetReconnectInterval 设置断线重连间隔, 需在 Start 前调用
func (c *Client) SetReconnectInterval(interval time.Duration) {
	c.reconnectInterval = interval
}

// SetHeartbeatInterval 设置心跳发送间隔, 需在 Start 前调用
func (c *Client) SetHeartbeatInterval(interval time.Duration) {
	if interval > 0 {
		c.heartbeatInterval = interval
	}
}

// SetTLS 设置 TLS 配置, 需在 Start 前调用
func (c *Client) SetTLS(config *tls.Config) {
	c.tlsConfig = config
}

// Start 连接并启动
func (c *Client) Start(addr string) error {
	c.mu.Lock()
	c.addr = addr
	c.stop = make(chan struct{})
	c.mu.Unlock()
	if err := c.dial(); err != nil {
		return err
	}
	go c.read()
	go c.heartbeat()
	return nil
}

// dial 建立 TCP/TLS 连接
func (c *Client) dial() error {
	var conn net.Conn
	var err error
	if c.tlsConfig != nil {
		conn, err = tls.Dial("tcp", c.addr, c.tlsConfig)
	} else {
		conn, err = net.Dial("tcp", c.addr)
	}
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.conn = conn
	c.reader = bufio.NewReader(conn)
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

// Send 发送消息
func (c *Client) Send(id uint32, data []byte) error {
	packData, err := packet.Pack(packet.NewMessage(id, data))
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return network.ErrConnectionClosed
	}
	_, err = c.conn.Write(packData)
	return err
}

// SetHandler 设置消息处理, 支持在运行中调用
func (c *Client) SetHandler(handler func(id uint32, data []byte)) {
	c.handler.Store(handler)
}

// read 读取消息, 断线时自动重连
func (c *Client) read() {
	for {
		select {
		case <-c.stop:
			return
		default:
		}
		c.mu.Lock()
		reader := c.reader
		c.mu.Unlock()
		msg, err := packet.TcpBufRead(reader)
		if err != nil {
			network.Log().Errorf("tcp client read error: %s", err)
			// 置空连接, 断线期间 Send 返回错误, 避免写入失效连接丢失消息
			c.mu.Lock()
			c.conn = nil
			c.reader = nil
			c.mu.Unlock()
			c.reconnectLoop()
			return
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
		if err := c.dial(); err != nil {
			time.Sleep(c.reconnectInterval)
			continue
		}
		network.Log().Infof("tcp client 重连成功 %s", c.addr)
		go c.read()
		return
	}
}

// heartbeat 心跳发送
func (c *Client) heartbeat() {
	ticker := time.NewTicker(c.heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			if err := c.Send(network.MessageIDHeartbeat, nil); err != nil {
				network.Log().Errorf("tcp client heartbeat send error: %s", err)
			}
		}
	}
}
