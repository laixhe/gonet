package udp

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/packet"
)

// Client UDP 客户端
type Client struct {
	conn     *net.UDPConn
	stop     chan struct{}
	stopOnce sync.Once
	handler  atomic.Value // func(id uint32, data []byte)
}

// NewClient 创建 UDP 客户端
func NewClient() *Client {
	return &Client{}
}

// Start 连接并启动接收, addr 为目标服务器地址
func (c *Client) Start(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}
	// 默认接收缓冲区可能小于大包尺寸导致丢包, 显式调大
	_ = conn.SetReadBuffer(udpBufferSize)
	_ = conn.SetWriteBuffer(udpBufferSize)
	c.conn = conn
	c.stop = make(chan struct{})
	go c.read()
	return nil
}

// Stop 停止
func (c *Client) Stop() error {
	c.stopOnce.Do(func() {
		if c.stop != nil {
			close(c.stop)
		}
		if c.conn != nil {
			_ = c.conn.Close()
		}
	})
	return nil
}

// Send 发送消息
func (c *Client) Send(id uint32, data []byte) error {
	if c.conn == nil {
		return ErrClosed
	}
	packData, err := packet.Pack(packet.NewMessage(id, data))
	if err != nil {
		return err
	}
	_, err = c.conn.Write(packData)
	return err
}

// SetHandler 设置消息处理, 支持运行中调用
func (c *Client) SetHandler(handler func(id uint32, data []byte)) {
	c.handler.Store(handler)
}

// read 读取并处理数据报
func (c *Client) read() {
	buf := make([]byte, udpMaxPacket)
	for {
		select {
		case <-c.stop:
			return
		default:
		}
		n, err := c.conn.Read(buf)
		if err != nil {
			network.Log().Errorf("udp client read error: %s", err)
			_ = c.Stop()
			return
		}
		msg, err := packet.Unpack(buf[:n])
		if err != nil {
			network.Log().Warnf("udp client unpack error: %s", err)
			continue
		}
		if handler, ok := c.handler.Load().(func(id uint32, data []byte)); ok && handler != nil {
			handler(msg.ID, msg.Data)
		}
	}
}
