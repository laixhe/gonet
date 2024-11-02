package tcp

import (
	"net"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/laixhe/gonet/network"
)

// manager 服务器连接管理器
type manager struct {
	id         int64           // 连接ID
	total      int64           // 总连接数
	server     network.IServer // 服务器
	pool       sync.Pool       // 用户链接对象池
	partitions []*partition    // 用户链接分区管理
}

var _ network.IManager = &manager{}

// partition 用户链接分区
type partition struct {
	rw          sync.RWMutex
	connections map[net.Conn]*Connection
}

func newManager(server *server) network.IManager {
	m := &manager{
		server:     server,
		pool:       sync.Pool{New: func() interface{} { return &Connection{} }},
		partitions: make([]*partition, 100),
	}
	for i := 0; i < 100; i++ {
		m.partitions[i] = &partition{connections: make(map[net.Conn]*Connection)}
	}
	return m
}

// Add 添加链接
func (m *manager) Add(c net.Conn) error {
	// 检查是否超过最大连接数
	if atomic.LoadInt64(&m.total) >= 1000 {
		return network.ErrTooManyConnection
	}

	id := atomic.AddInt64(&m.id, 1)
	conn := m.pool.Get().(*Connection)
	conn.init(c, m, id)
	index := int(reflect.ValueOf(c).Pointer()) % len(m.partitions)
	m.partitions[index].store(c, conn)
	atomic.AddInt64(&m.total, 1)

	return nil
}

// Remove 删除连接
func (m *manager) Remove(c net.Conn) {
	index := int(reflect.ValueOf(c).Pointer()) % len(m.partitions)
	if conn, ok := m.partitions[index].delete(c); ok {
		conn.Clear()
		m.pool.Put(conn)
		atomic.AddInt64(&m.total, -1)
	}
}

// Close 关闭连接
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

// 存储连接
func (p *partition) store(c net.Conn, conn *Connection) {
	p.rw.Lock()
	p.connections[c] = conn
	p.rw.Unlock()
}

// 加载连接
func (p *partition) load(c net.Conn) (*Connection, bool) {
	p.rw.RLock()
	conn, ok := p.connections[c]
	p.rw.RUnlock()

	return conn, ok
}

// 删除连接
func (p *partition) delete(c net.Conn) (*Connection, bool) {
	p.rw.Lock()
	conn, ok := p.connections[c]
	if ok {
		delete(p.connections, c)
	}
	p.rw.Unlock()

	return conn, ok
}

// 关闭该分片内的所有连接
func (p *partition) close() {
	for _, conn := range p.connections {
		_ = conn.Stop()
	}
}
