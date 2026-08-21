package websocket

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/packet"
)

// Connection 用户链接
type Connection struct {
	id                int64                                            // 连接ID
	uid               int64                                            // 用户ID
	state             int32                                            // 连接状态
	conn              *websocket.Conn                                  // WebSocket 连接
	writeMu           sync.Mutex                                       // 写互斥(gorilla 单写限制)
	close             chan struct{}                                    // 关闭信号
	manager           network.IManager                                 // 连接管理器
	lastHeartbeatTime int64                                            // 上次心跳时间
	remoteAddr        string                                           // 远程地址
	heartbeatInterval time.Duration                                    // 心跳检测间隔
	heartbeatTimeout  time.Duration                                    // 心跳超时时间
	writeTimeout      time.Duration                                    // 写超时时间, 0 不超时
	stopOnce          sync.Once                                        // 停止幂等
	dispatch          func(conn network.IConn, id uint32, data []byte) // 消息分发
}

var _ network.IConn = &Connection{}

// ID 获取当前连接ID
func (c *Connection) ID() int64 {
	return c.id
}

// Uid 获取用户ID
func (c *Connection) Uid() int64 {
	return atomic.LoadInt64(&c.uid)
}

// BindUid 绑定用户ID
func (c *Connection) BindUid(uid int64) {
	atomic.StoreInt64(&c.uid, uid)
}

// UnbindUid 解绑用户ID
func (c *Connection) UnbindUid() {
	atomic.StoreInt64(&c.uid, 0)
}

// State 获取连接状态
func (c *Connection) State() int32 {
	return atomic.LoadInt32(&c.state)
}

// IsClosed 是否连接关闭
func (c *Connection) IsClosed() bool {
	return atomic.LoadInt32(&c.state) == network.ConnClosed
}

// RemoteAddr 获取远程地址(ip:port)
func (c *Connection) RemoteAddr() string {
	return c.remoteAddr
}

// Send 发送消息(WebSocket 二进制帧, 消息体为 packet 协议格式)
func (c *Connection) Send(id uint32, data []byte) error {
	if c.IsClosed() {
		return network.ErrConnectionClosed
	}
	packData, err := packet.Pack(packet.NewMessage(id, data))
	if err != nil {
		return err
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.IsClosed() {
		return network.ErrConnectionClosed
	}
	// 写超时保护: 对端不读导致写阻塞时, 超过时限断开连接
	if c.writeTimeout > 0 {
		_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}
	return c.conn.WriteMessage(websocket.BinaryMessage, packData)
}

// Stop 停止连接，结束当前连接
func (c *Connection) Stop() error {
	c.stop()
	c.manager.Remove(c)
	return nil
}

// stop 关闭连接底层资源，幂等
func (c *Connection) stop() {
	c.stopOnce.Do(func() {
		atomic.StoreInt32(&c.state, network.ConnClosed)
		close(c.close)
		c.writeMu.Lock()
		_ = c.conn.Close()
		c.writeMu.Unlock()
	})
}

// init 初始化连接字段
func (c *Connection) init(conn *websocket.Conn, manager network.IManager, id int64, dispatch func(network.IConn, uint32, []byte), heartbeatInterval, heartbeatTimeout, writeTimeout time.Duration) {
	c.id = id
	c.uid = 0
	c.state = network.ConnOpened
	c.conn = conn
	c.close = make(chan struct{})
	c.manager = manager
	c.dispatch = dispatch
	c.heartbeatInterval = heartbeatInterval
	c.heartbeatTimeout = heartbeatTimeout
	c.writeTimeout = writeTimeout
	c.lastHeartbeatTime = time.Now().UnixNano()
	c.remoteAddr = conn.RemoteAddr().String()
	// 收到 pong 刷新心跳时间
	conn.SetPongHandler(func(string) error {
		atomic.StoreInt64(&c.lastHeartbeatTime, time.Now().UnixNano())
		return nil
	})
}

// Start 启动连接协程
func (c *Connection) Start() {
	go c.read()
	go c.heartbeat()
	network.Log().Infof("websocket accept init %d %s", c.id, c.remoteAddr)
}

// read 读取消息并分发
func (c *Connection) read() {
	for {
		select {
		case <-c.close:
			return
		default:
			msgType, data, err := c.conn.ReadMessage()
			if err != nil {
				network.Log().Errorf("websocket read error %d %s %s", c.id, c.remoteAddr, err)
				_ = c.Stop()
				return
			}
			if msgType != websocket.BinaryMessage {
				continue // 忽略文本等其他类型
			}
			atomic.StoreInt64(&c.lastHeartbeatTime, time.Now().UnixNano())
			msg, err := packet.Unpack(data)
			if err != nil {
				network.Log().Errorf("websocket unpack error %d %s %s", c.id, c.remoteAddr, err)
				continue
			}
			c.dispatch(c, msg.ID, msg.Data)
		}
	}
}

// heartbeat 心跳检测: 发送 ping, 超时断开连接
func (c *Connection) heartbeat() {
	ticker := time.NewTicker(c.heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.close:
			return
		case now := <-ticker.C:
			// 发送 ping, 对端(浏览器/gorilla 客户端)自动回 pong
			c.writeMu.Lock()
			if c.writeTimeout > 0 {
				_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
			}
			_ = c.conn.WriteMessage(websocket.PingMessage, nil)
			c.writeMu.Unlock()
			last := time.Unix(0, atomic.LoadInt64(&c.lastHeartbeatTime))
			if now.Sub(last) > c.heartbeatTimeout {
				network.Log().Warnf("websocket heartbeat timeout %d %s", c.id, c.remoteAddr)
				_ = c.Stop()
				return
			}
		}
	}
}
