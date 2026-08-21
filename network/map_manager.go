package network

import (
	"sync"
	"sync/atomic"
)

// MapManager 基于 map 的连接管理器, 适合连接数规模中等、锁竞争不激烈的服务器(如 websocket/kcp), 实现 IManager
type MapManager struct {
	mu           sync.RWMutex
	conns        map[int64]IConn
	id           int64            // 连接ID
	total        int64            // 总连接数
	maxConn      int64            // 最大连接数
	StartFn      func(conn IConn) // 启动连接协程
	ConnectFn    func(conn IConn) // 连接建立回调
	DisconnectFn func(conn IConn) // 连接断开回调
}

var _ IManager = (*MapManager)(nil)

// NewMapManager 创建连接管理器
func NewMapManager(maxConn int64, start, connect, disconnect func(IConn)) *MapManager {
	return &MapManager{
		conns:        make(map[int64]IConn),
		maxConn:      maxConn,
		StartFn:      start,
		ConnectFn:    connect,
		DisconnectFn: disconnect,
	}
}

// NextID 分配连接ID
func (m *MapManager) NextID() int64 {
	return atomic.AddInt64(&m.id, 1)
}

// Add 添加链接
func (m *MapManager) Add(conn IConn) error {
	if atomic.LoadInt64(&m.total) >= m.maxConn {
		_ = conn.Stop()
		return ErrTooManyConnection
	}
	m.mu.Lock()
	m.conns[conn.ID()] = conn
	m.mu.Unlock()
	atomic.AddInt64(&m.total, 1)
	// 先通知连接建立(含 Uid 绑定等初始化), 再启动消息协程, 避免消息先于回调被处理
	if m.ConnectFn != nil {
		m.ConnectFn(conn)
	}
	if m.StartFn != nil {
		m.StartFn(conn)
	}
	return nil
}

// Remove 删除连接
func (m *MapManager) Remove(conn IConn) {
	if conn == nil {
		return
	}
	m.mu.Lock()
	_, ok := m.conns[conn.ID()]
	if ok {
		delete(m.conns, conn.ID())
	}
	m.mu.Unlock()
	if ok {
		atomic.AddInt64(&m.total, -1)
		if m.DisconnectFn != nil {
			m.DisconnectFn(conn)
		}
	}
}

// Close 关闭所有连接
func (m *MapManager) Close() {
	m.mu.RLock()
	conns := make([]IConn, 0, len(m.conns))
	for _, conn := range m.conns {
		conns = append(conns, conn)
	}
	m.mu.RUnlock()
	for _, conn := range conns {
		_ = conn.Stop()
	}
}

// Count 当前连接数
func (m *MapManager) Count() int64 {
	return atomic.LoadInt64(&m.total)
}

// FindByID 按 ID 查找连接, 未找到返回 nil
func (m *MapManager) FindByID(id int64) IConn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conns[id]
}

// KickByID 按 ID 踢下线(关闭连接)
func (m *MapManager) KickByID(id int64) {
	if conn := m.FindByID(id); conn != nil {
		_ = conn.Stop()
	}
}

// FindByUid 按 Uid 查找连接, 未找到返回 nil
func (m *MapManager) FindByUid(uid int64) IConn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, conn := range m.conns {
		if conn.Uid() == uid {
			return conn
		}
	}
	return nil
}

// KickByUid 按 Uid 踢下线(关闭连接)
func (m *MapManager) KickByUid(uid int64) {
	if conn := m.FindByUid(uid); conn != nil {
		_ = conn.Stop()
	}
}

// ForEach 遍历所有连接
func (m *MapManager) ForEach(fn func(conn IConn)) {
	m.mu.RLock()
	conns := make([]IConn, 0, len(m.conns))
	for _, conn := range m.conns {
		conns = append(conns, conn)
	}
	m.mu.RUnlock()
	for _, conn := range conns {
		fn(conn)
	}
}
