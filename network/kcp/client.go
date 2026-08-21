package kcp

import (
	"bufio"
	"sync"
	"sync/atomic"
	"time"

	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/packet"
	kcpv5 "github.com/xtaci/kcp-go/v5"
)

// Client KCP 客户端, 实现 network.IClient, FEC 分片与加密参数需与服务端一致
type Client struct {
	mu                sync.Mutex
	session           *kcpv5.UDPSession
	stop              chan struct{}
	stopOnce          sync.Once
	block             kcpv5.BlockCrypt // 加密块
	dataShards        int              // FEC 数据分片数
	parityShards      int              // FEC 冗余分片数
	noDelay           bool             // 快速模式
	mtu               int              // MTU
	heartbeatInterval time.Duration    // 心跳发送间隔
	handler           atomic.Value     // 消息处理, 支持并发设置
}

var _ network.IClient = &Client{}

// NewClient 创建默认配置的 KCP 客户端
func NewClient() network.IClient {
	return NewClientWithConfig(DefaultConfig())
}

// NewClientWithConfig 创建指定配置的 KCP 客户端, FEC 分片与加密需与服务端一致
func NewClientWithConfig(config Config) network.IClient {
	config.Check()
	return &Client{
		block:             config.Block,
		dataShards:        config.DataShards,
		parityShards:      config.ParityShards,
		noDelay:           config.NoDelay,
		mtu:               config.Mtu,
		heartbeatInterval: config.HeartbeatInterval,
	}
}

// Start 连接并启动
func (c *Client) Start(addr string) error {
	c.mu.Lock()
	c.stop = make(chan struct{})
	c.mu.Unlock()
	if err := c.dial(addr); err != nil {
		return err
	}
	// KCP 为懒握手: 服务端收到首个数据包才建立会话, 主动发一次心跳触发
	if err := c.Send(network.MessageIDHeartbeat, nil); err != nil {
		return err
	}
	go c.read()
	go c.heartbeat()
	return nil
}

// dial 建立 KCP 会话
func (c *Client) dial(addr string) error {
	session, err := kcpv5.DialWithOptions(addr, c.block, c.dataShards, c.parityShards)
	if err != nil {
		return err
	}
	if c.noDelay {
		session.SetNoDelay(1, 10, 2, 1)
	}
	if c.mtu > 0 {
		session.SetMtu(c.mtu)
	}
	c.mu.Lock()
	c.session = session
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
		session := c.session
		c.mu.Unlock()
		if session != nil {
			_ = session.Close()
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
	if c.session == nil {
		return network.ErrConnectionClosed
	}
	_, err = c.session.Write(packData)
	return err
}

// SetHandler 设置消息处理, 支持在运行中调用
func (c *Client) SetHandler(handler func(id uint32, data []byte)) {
	c.handler.Store(handler)
}

// read 读取消息
func (c *Client) read() {
	reader := bufio.NewReader(c.session)
	for {
		select {
		case <-c.stop:
			return
		default:
		}
		msg, err := packet.TcpBufRead(reader)
		if err != nil {
			network.Log().Errorf("kcp client read error: %s", err)
			_ = c.Stop()
			return
		}
		if handler, ok := c.handler.Load().(func(id uint32, data []byte)); ok && handler != nil {
			handler(msg.ID, msg.Data)
		}
	}
}

// heartbeat 心跳发送, 同时维持对端会话活跃
func (c *Client) heartbeat() {
	ticker := time.NewTicker(c.heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			if err := c.Send(network.MessageIDHeartbeat, nil); err != nil {
				network.Log().Errorf("kcp client heartbeat send error: %s", err)
			}
		}
	}
}
