package tcp

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/laixhe/gonet/network"
)

// manager 服务器连接管理器
type manager struct {
	id                int64         // 连接ID
	total             int64         // 总连接数
	maxConnections    int64         // 最大连接数
	heartbeatInterval time.Duration // 心跳检测间隔
	heartbeatTimeout  time.Duration // 心跳超时时间
	processWorkers    int           // 每连接消息处理 worker 数
	server            *server       // 服务器
	partitions        []*partition  // 用户链接分区管理
}

var _ network.IManager = &manager{}

// partition 用户链接分区
type partition struct {
	rw          sync.RWMutex
	connections map[int64]*Connection
}

func newManager(server *server, config Config) *manager {
	m := &manager{
		server:            server,
		maxConnections:    config.MaxConnections,
		heartbeatInterval: config.HeartbeatInterval,
		heartbeatTimeout:  config.HeartbeatTimeout,
		processWorkers:    config.ProcessWorkers,
		partitions:        make([]*partition, config.Partitions),
	}
	for i := 0; i < config.Partitions; i++ {
		m.partitions[i] = &partition{connections: make(map[int64]*Connection)}
	}
	return m
}

// nextID 分配连接ID
func (m *manager) nextID() int64 {
	return atomic.AddInt64(&m.id, 1)
}

// Add 添加链接
func (m *manager) Add(c network.IConn) error {
	conn, ok := c.(*Connection)
	if !ok {
		return errors.New("tcp manager: invalid connection type")
	}
	// 超过最大连接数直接关闭
	if atomic.LoadInt64(&m.total) >= m.maxConnections {
		_ = conn.Stop()
		return network.ErrTooManyConnection
	}

	// 先注册到分片, 再通知连接建立(含 Uid 绑定等初始化), 最后启动消息协程,
	// 避免 Stop 关闭连接时漏掉刚建立的连接, 也避免消息先于回调被处理
	index := int(conn.ID()) % len(m.partitions)
	m.partitions[index].store(conn)
	atomic.AddInt64(&m.total, 1)
	m.server.NotifyConnect(conn)
	conn.Start()

	return nil
}

// Remove 删除连接
func (m *manager) Remove(c network.IConn) {
	if c == nil {
		return
	}
	index := int(c.ID()) % len(m.partitions)
	if conn, ok := m.partitions[index].delete(c.ID()); ok {
		atomic.AddInt64(&m.total, -1)
		m.server.NotifyDisconnect(conn)
	}
}

// Close 关闭所有连接
func (m *manager) Close() {
	var wg sync.WaitGroup
	wg.Add(len(m.partitions))

	for i := range m.partitions {
		p := m.partitions[i]
		go func() {
			p.close()
			wg.Done()
		}()
	}

	wg.Wait()
}

// Count 当前连接数
func (m *manager) Count() int64 {
	return atomic.LoadInt64(&m.total)
}

// FindByID 按 ID 查找连接, 未找到返回 nil
func (m *manager) FindByID(id int64) network.IConn {
	index := int(id) % len(m.partitions)
	conn := m.partitions[index].findByID(id)
	if conn == nil {
		return nil // 避免 nil *Connection 装箱为非 nil 接口
	}
	return conn
}

// KickByID 按 ID 踢下线(关闭连接)
func (m *manager) KickByID(id int64) {
	if conn := m.FindByID(id); conn != nil {
		_ = conn.Stop()
	}
}

// FindByUid 按 Uid 查找连接, 未找到返回 nil
func (m *manager) FindByUid(uid int64) network.IConn {
	for i := range m.partitions {
		if conn := m.partitions[i].findByUid(uid); conn != nil {
			return conn
		}
	}
	return nil
}

// KickByUid 按 Uid 踢下线(关闭连接)
func (m *manager) KickByUid(uid int64) {
	if conn := m.FindByUid(uid); conn != nil {
		_ = conn.Stop()
	}
}

// ForEach 遍历所有连接
func (m *manager) ForEach(fn func(conn network.IConn)) {
	for i := range m.partitions {
		for _, conn := range m.partitions[i].snapshot() {
			fn(conn)
		}
	}
}

// store 存储连接
func (p *partition) store(conn *Connection) {
	p.rw.Lock()
	p.connections[conn.ID()] = conn
	p.rw.Unlock()
}

// delete 删除连接
func (p *partition) delete(id int64) (*Connection, bool) {
	p.rw.Lock()
	conn, ok := p.connections[id]
	if ok {
		delete(p.connections, id)
	}
	p.rw.Unlock()

	return conn, ok
}

// close 关闭分片内所有连接
func (p *partition) close() {
	for _, conn := range p.snapshot() {
		_ = conn.Stop()
	}
}

// snapshot 获取分片内连接快照
func (p *partition) snapshot() []*Connection {
	p.rw.RLock()
	conns := make([]*Connection, 0, len(p.connections))
	for _, conn := range p.connections {
		conns = append(conns, conn)
	}
	p.rw.RUnlock()
	return conns
}

// findByUid 按 Uid 查找分片内连接
func (p *partition) findByUid(uid int64) *Connection {
	p.rw.RLock()
	defer p.rw.RUnlock()
	for _, conn := range p.connections {
		if conn.Uid() == uid {
			return conn
		}
	}
	return nil
}

// findByID 按 ID 查找分片内连接
func (p *partition) findByID(id int64) *Connection {
	p.rw.RLock()
	defer p.rw.RUnlock()
	return p.connections[id]
}
