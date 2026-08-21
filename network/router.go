package network

import "sync"

// Router 消息路由, tcp/kcp/websocket 服务器共用
type Router struct {
	mu             sync.RWMutex
	handlers       map[uint32]RouterHandler
	defaultHandler func(conn IConn, id uint32, data []byte)
}

// NewRouter 创建消息路由
func NewRouter() *Router {
	return &Router{handlers: make(map[uint32]RouterHandler)}
}

// Register 注册消息路由
func (r *Router) Register(id uint32, handler RouterHandler) {
	r.mu.Lock()
	r.handlers[id] = handler
	r.mu.Unlock()
}

// SetDefault 设置默认消息处理, 未注册的消息走该处理器
func (r *Router) SetDefault(handler func(conn IConn, id uint32, data []byte)) {
	r.mu.Lock()
	r.defaultHandler = handler
	r.mu.Unlock()
}

// Dispatch 分发消息
func (r *Router) Dispatch(conn IConn, id uint32, data []byte) {
	r.mu.RLock()
	h := r.handlers[id]
	if h == nil && r.defaultHandler != nil {
		r.defaultHandler(conn, id, data)
		r.mu.RUnlock()
		return
	}
	r.mu.RUnlock()
	if h != nil {
		h(conn, data)
	}
}
