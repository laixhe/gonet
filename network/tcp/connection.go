package tcp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/packet"
)

// Connection 用户链接
type Connection struct {
	id                int64            // 连接ID，全局唯一
	uid               int64            // 用户ID
	state             int32            // 连接状态
	conn              net.Conn         // TCP源连接
	connReader        *bufio.Reader    // TCP源连接读缓冲
	rw                *sync.RWMutex    // 读写锁
	chWrite           chan []byte      // 写入队列
	close             chan struct{}    // 关闭信号
	manager           network.IManager // 连接管理器
	lastHeartbeatTime int64            // 上次心跳时间
	remoteAddr        string           // 远程地址
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

func (c *Connection) Clear() {
	if !c.IsClosed() {
		atomic.StoreInt32(&c.state, network.ConnClosed)
		close(c.chWrite)
		close(c.close)
	}
	c.uid = 0
	c.conn = nil
	c.connReader = nil
	c.rw = nil
	c.chWrite = nil
	c.close = nil
	c.manager = nil
	c.lastHeartbeatTime = 0
}

func (c *Connection) init(conn net.Conn, manager network.IManager, id int64) {
	c.id = id
	atomic.StoreInt64(&c.uid, 0)
	atomic.StoreInt32(&c.state, network.ConnOpened)
	c.conn = conn
	c.connReader = bufio.NewReader(conn)
	c.rw = &sync.RWMutex{}
	c.chWrite = make(chan []byte, 4096)
	c.close = make(chan struct{})
	c.manager = manager
	atomic.StoreInt64(&c.lastHeartbeatTime, time.Now().UnixNano())
	c.remoteAddr = conn.RemoteAddr().String()

	go c.read()
	go c.write()

	log.Printf("tcp accept init %d %s", c.id, c.remoteAddr)
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
			log.Printf("tcp read chan close %d %s", c.id, c.remoteAddr)
			return
		default:
			packMessage, err := packet.TcpBufRead(c.connReader)
			if err != nil {
				log.Printf("tcp read error %d %s %s", c.id, c.remoteAddr, err)
				if err1 := c.Stop(); err1 != nil {
					log.Printf("tcp read stop error %d %s %s", c.id, c.remoteAddr, err1)
				}
				return
			}
			// 当前心跳时间
			atomic.StoreInt64(&c.lastHeartbeatTime, time.Now().UnixNano())

			// 发送消息
			c.chWrite <- []byte(string(packMessage.Data) + fmt.Sprintf("--%v", time.Now().UnixMilli()))
			log.Println(string(packMessage.Data))
		}
	}
}

// write 写入消息
func (c *Connection) write() {
	defer func() {
		if err := c.Stop(); err != nil {
			log.Printf("tcp write stop error %d %s %s", c.id, c.remoteAddr, err)
		}
	}()

	for {
		select {
		case <-c.close:
			log.Printf("tcp write chan close %d %s", c.id, c.remoteAddr)
			return
		case data, ok := <-c.chWrite:
			if !ok {
				log.Printf("tcp write chan %d %s", c.id, c.remoteAddr)
				return
			}
			if c.IsClosed() {
				log.Printf("tcp write closed error %d %s", c.id, c.remoteAddr)
				return
			}
			packData, err := packet.Pack(packet.NewMessage(100, data))
			if err != nil {
				log.Printf("tcp write packet message error %d %s %s", c.id, c.remoteAddr, err)
				continue
			}
			_, err = c.conn.Write(packData)
			if err != nil {
				log.Printf("tcp write message error %d %s %s", c.id, c.remoteAddr, err)
				continue
			}
		}
	}
}
