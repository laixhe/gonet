package network

import (
	"bufio"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/laixhe/gonet/network/packet"
)

// StreamConnection 基于 net.Conn 流式连接的通用实现,
// 提供消息收发(写队列)、worker 并发处理与心跳检测, tcp/kcp 服务器共用。
type StreamConnection struct {
	IDVal             int64                                    // 连接ID
	UidVal            int64                                    // 用户ID
	StateVal          int32                                    // 连接状态
	Conn              net.Conn                                 // 底层连接
	ConnReader        *bufio.Reader                            // 读缓冲
	Rw                *sync.RWMutex                            // 读写锁
	ChWrite           chan []byte                              // 写入队列
	ChProcess         chan packet.Message                      // 消息处理队列
	CloseChan         chan struct{}                            // 关闭信号
	Manager           IManager                                 // 连接管理器
	LastHeartbeatTime int64                                    // 上次心跳时间
	RemoteAddrVal     string                                   // 远程地址
	HeartbeatInterval time.Duration                            // 心跳检测间隔
	HeartbeatTimeout  time.Duration                            // 心跳超时时间
	WriteTimeout      time.Duration                            // 写超时时间, 写阻塞超过该时间则断开连接, 0 不超时
	WorkerCount       int                                      // 消息处理 worker 数
	StopOnce          sync.Once                                // 停止幂等
	DispatchFn        func(conn IConn, id uint32, data []byte) // 消息分发
	LogPrefix         string                                   // 日志前缀
	processMu         sync.Mutex                               // 保护 processWg, 避免 Start 的 Add 与 stop 的 Wait 并发
	processWg         sync.WaitGroup                           // 消息处理协程, 平滑关闭等待排空
}

// NewStreamConnection 创建流式连接
func NewStreamConnection(conn net.Conn, manager IManager, id int64, dispatch func(IConn, uint32, []byte), heartbeatInterval, heartbeatTimeout, writeTimeout time.Duration, workerCount int, logPrefix string) *StreamConnection {
	return &StreamConnection{
		IDVal:             id,
		StateVal:          ConnOpened,
		Conn:              conn,
		ConnReader:        bufio.NewReader(conn),
		Rw:                &sync.RWMutex{},
		ChWrite:           make(chan []byte, 4096),
		ChProcess:         make(chan packet.Message, 1024),
		CloseChan:         make(chan struct{}),
		Manager:           manager,
		HeartbeatInterval: heartbeatInterval,
		HeartbeatTimeout:  heartbeatTimeout,
		WriteTimeout:      writeTimeout,
		WorkerCount:       workerCount,
		DispatchFn:        dispatch,
		LogPrefix:         logPrefix,
		LastHeartbeatTime: time.Now().UnixNano(),
		RemoteAddrVal:     conn.RemoteAddr().String(),
	}
}

// ID 获取当前连接ID
func (c *StreamConnection) ID() int64 {
	return c.IDVal
}

// Uid 获取用户ID
func (c *StreamConnection) Uid() int64 {
	return atomic.LoadInt64(&c.UidVal)
}

// BindUid 绑定用户ID
func (c *StreamConnection) BindUid(uid int64) {
	atomic.StoreInt64(&c.UidVal, uid)
}

// UnbindUid 解绑用户ID
func (c *StreamConnection) UnbindUid() {
	atomic.StoreInt64(&c.UidVal, 0)
}

// State 获取连接状态
func (c *StreamConnection) State() int32 {
	return atomic.LoadInt32(&c.StateVal)
}

// IsClosed 是否连接关闭
func (c *StreamConnection) IsClosed() bool {
	return atomic.LoadInt32(&c.StateVal) == ConnClosed
}

// RemoteAddr 获取远程地址(ip:port)
func (c *StreamConnection) RemoteAddr() string {
	return c.RemoteAddrVal
}

// Send 发送消息
func (c *StreamConnection) Send(id uint32, data []byte) error {
	packData, err := packet.Pack(packet.NewMessage(id, data))
	if err != nil {
		return err
	}
	c.Rw.RLock()
	defer c.Rw.RUnlock()
	if c.IsClosed() {
		return ErrConnectionClosed
	}
	select {
	case c.ChWrite <- packData:
		return nil
	default:
		// 写入队列已满, 对端消费过慢, 标记连接挂起
		atomic.CompareAndSwapInt32(&c.StateVal, ConnOpened, ConnHanged)
		return ErrConnectionHanged
	}
}

// Stop 停止连接，结束当前连接
func (c *StreamConnection) Stop() error {
	c.stop()
	c.Manager.Remove(c)
	return nil
}

// stop 关闭连接底层资源，幂等
func (c *StreamConnection) stop() {
	c.StopOnce.Do(func() {
		atomic.StoreInt32(&c.StateVal, ConnClosed)
		c.Rw.Lock()
		close(c.CloseChan)
		c.Rw.Unlock()
		// 平滑关闭: 等待消息处理协程排空已入队消息, 再关闭写通道与底层连接
		c.processMu.Lock()
		c.processWg.Wait()
		c.processMu.Unlock()
		c.Rw.Lock()
		close(c.ChWrite)
		conn := c.Conn
		c.Rw.Unlock()
		if conn != nil {
			_ = conn.Close()
		}
	})
}

// Start 启动连接协程
func (c *StreamConnection) Start() {
	go c.read()
	go c.write()
	// 多个 worker 并发处理消息, 慢 handler 不阻塞同连接其他消息(可能乱序)
	c.processMu.Lock()
	for i := 0; i < c.WorkerCount; i++ {
		c.processWg.Add(1)
		go c.process()
	}
	c.processMu.Unlock()
	go c.heartbeat()

	Log().Infof("%s accept init %d %s", c.LogPrefix, c.IDVal, c.RemoteAddrVal)
}

// read 读取消息
func (c *StreamConnection) read() {
	for {
		select {
		case <-c.CloseChan:
			return
		default:
			packMessage, err := packet.TcpBufRead(c.ConnReader)
			if err != nil {
				Log().Errorf("%s read error %d %s %s", c.LogPrefix, c.IDVal, c.RemoteAddrVal, err)
				_ = c.Stop()
				return
			}
			atomic.StoreInt64(&c.LastHeartbeatTime, time.Now().UnixNano())
			// 入队处理, 队列满时阻塞形成背压
			select {
			case <-c.CloseChan:
				return
			case c.ChProcess <- *packMessage:
			}
		}
	}
}

// process 处理消息
func (c *StreamConnection) process() {
	defer c.processWg.Done()
	for {
		select {
		case <-c.CloseChan:
			// 平滑关闭: 排空已入队消息后退出
			for {
				select {
				case msg := <-c.ChProcess:
					c.DispatchFn(c, msg.ID, msg.Data)
				default:
					return
				}
			}
		case msg := <-c.ChProcess:
			c.DispatchFn(c, msg.ID, msg.Data)
		}
	}
}

// write 写入消息
func (c *StreamConnection) write() {
	for {
		select {
		case <-c.CloseChan:
			return
		case data, ok := <-c.ChWrite:
			if !ok {
				return
			}
			// 写超时保护: 对端不读导致写阻塞时, 超过时限断开连接
			if c.WriteTimeout > 0 {
				_ = c.Conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
			}
			if _, err := c.Conn.Write(data); err != nil {
				Log().Errorf("%s write error %d %s %s", c.LogPrefix, c.IDVal, c.RemoteAddrVal, err)
				_ = c.Stop()
				return
			}
			// 写入队列已清空, 连接恢复正常
			if len(c.ChWrite) == 0 {
				atomic.CompareAndSwapInt32(&c.StateVal, ConnHanged, ConnOpened)
			}
		}
	}
}

// heartbeat 心跳检测，超时断开连接
func (c *StreamConnection) heartbeat() {
	ticker := time.NewTicker(c.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.CloseChan:
			return
		case now := <-ticker.C:
			last := time.Unix(0, atomic.LoadInt64(&c.LastHeartbeatTime))
			if now.Sub(last) > c.HeartbeatTimeout {
				Log().Warnf("%s heartbeat timeout %d %s", c.LogPrefix, c.IDVal, c.RemoteAddrVal)
				_ = c.Stop()
				return
			}
		}
	}
}
