package tcp

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/packet"
)

// Connection 用户链接
type Connection struct {
	id                int64            // 连接ID，全局唯一
	uid               int64            // 用户ID
	state             int32            // 连接状态
	conn              net.Conn         // TCP源连接
	connReader        *bufio.Reader    // TCP源连接读缓冲
	rw                sync.RWMutex     // 读写锁
	chWrite           chan []byte      // 写入队列
	close             chan struct{}    // 关闭信号
	manager           network.IManager // 连接管理器
	lastHeartbeatTime int64            // 上次心跳时间
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

func (c *Connection) init(conn net.Conn, manager network.IManager, id int64) {
	c.id = id
	atomic.StoreInt64(&c.uid, 0)
	atomic.StoreInt32(&c.state, network.ConnOpened)
	c.conn = conn
	c.connReader = bufio.NewReader(conn)
	c.chWrite = make(chan []byte, 4096)
	c.close = make(chan struct{})
	c.manager = manager
	atomic.StoreInt64(&c.lastHeartbeatTime, time.Now().UnixNano())

	go c.read()
	go c.write()

	// TODO: log
	fmt.Println("tcp accept init", conn.RemoteAddr(), c.id)
}

// Stop 停止连接，结束当前连接
func (c *Connection) Stop() error {
	if c.IsClosed() {
		return nil
	}
	c.rw.Lock()

	atomic.StoreInt32(&c.state, network.ConnClosed)
	conn := c.conn
	c.conn = nil
	c.connReader = nil
	close(c.chWrite)
	close(c.close)

	c.rw.Unlock()
	// 删除连接
	c.manager.Remove(conn)
	return conn.Close()
}

// read 读取消息
func (c *Connection) read() {
	for {
		select {
		case <-c.close:
			// TODO: log
			fmt.Println("tcp read chan close", c.conn.RemoteAddr(), c.id)
			return
		default:
			packMessage, err := packet.TcpBufRead(c.connReader)
			if err != nil {
				// TODO: log
				fmt.Println("tcp read error", c.conn.RemoteAddr(), c.id, err)
				if err1 := c.Stop(); err1 != nil {
					// TODO: log
					fmt.Println("tcp read stop error", c.conn.RemoteAddr(), c.id, err1)
				}
				return
			}

			// 当前心跳时间
			atomic.StoreInt64(&c.lastHeartbeatTime, time.Now().UnixNano())

			c.chWrite <- []byte(string(packMessage.Data) + fmt.Sprintf("--%v", time.Now().UnixMilli()))
			fmt.Println(string(packMessage.Data))
		}
	}
}

// write 写入消息
func (c *Connection) write() {
	defer c.Stop()

	for {
		select {
		case <-c.close:
			// TODO: log
			fmt.Println("tcp write chan close", c.conn.RemoteAddr(), c.id)
			return
		case data, ok := <-c.chWrite:
			if !ok {
				// TODO: log
				fmt.Println("tcp write chan", c.conn.RemoteAddr(), c.id)
				return
			}
			if c.IsClosed() {
				// TODO: log
				fmt.Println("tcp write closed error", c.conn.RemoteAddr(), c.id)
				return
			}
			packData, err := packet.Pack(packet.NewMessage(100, data))
			if err != nil {
				// TODO: log
				fmt.Println("tcp write packet message error", c.conn.RemoteAddr(), c.id, err)
				continue
			}
			_, err = c.conn.Write(packData)
			if err != nil {
				// TODO: log
				fmt.Println("tcp write message error", c.conn.RemoteAddr(), c.id, err)
				continue
			}
		}
	}
}
