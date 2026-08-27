package network

import "sync"

// BaseServer 服务器公共实现, 提供 IServer 中路由、连接事件回调与广播部分, tcp/kcp/websocket 服务器嵌入复用
type BaseServer struct {
	mu           sync.Mutex
	onConnect    func(IConn)
	onDisconnect func(IConn)
	router       *Router
	manager      IManager
}

// NewBaseServer 创建服务器公共实现
func NewBaseServer() *BaseServer {
	return &BaseServer{router: NewRouter()}
}

// SetManager 设置连接管理器, 供 Broadcast 使用, 需在构造服务器时调用
func (s *BaseServer) SetManager(m IManager) {
	s.manager = m
}

// OnConnect 设置连接建立回调, 支持运行中注册
func (s *BaseServer) OnConnect(handler func(conn IConn)) {
	s.mu.Lock()
	s.onConnect = handler
	s.mu.Unlock()
}

// OnDisconnect 设置连接断开回调, 支持运行中注册
func (s *BaseServer) OnDisconnect(handler func(conn IConn)) {
	s.mu.Lock()
	s.onDisconnect = handler
	s.mu.Unlock()
}

// NotifyConnect 通知连接建立
func (s *BaseServer) NotifyConnect(conn IConn) {
	s.mu.Lock()
	h := s.onConnect
	s.mu.Unlock()
	if h != nil {
		h(conn)
	}
}

// NotifyDisconnect 通知连接断开
func (s *BaseServer) NotifyDisconnect(conn IConn) {
	s.mu.Lock()
	h := s.onDisconnect
	s.mu.Unlock()
	if h != nil {
		h(conn)
	}
}

// Router 注册消息路由
func (s *BaseServer) Router(id uint32, handler RouterHandler) {
	s.router.Register(id, handler)
}

// SetDefaultHandler 设置默认消息处理, 未注册的消息走该处理器
func (s *BaseServer) SetDefaultHandler(handler func(conn IConn, id uint32, data []byte)) {
	s.router.SetDefault(handler)
}

// Dispatch 分发消息
func (s *BaseServer) Dispatch(conn IConn, id uint32, data []byte) {
	s.router.Dispatch(conn, id, data)
}

// Broadcast 广播消息给所有连接
func (s *BaseServer) Broadcast(id uint32, data []byte) {
	s.broadcast(id, data, 0, false)
}

// BroadcastExclude 广播消息给所有连接, 排除指定 Uid(未绑定 Uid 的连接视为 Uid 0)
func (s *BaseServer) BroadcastExclude(id uint32, data []byte, excludeUid int64) {
	s.broadcast(id, data, excludeUid, true)
}

// broadcast 遍历连接发送消息, 排除模式下跳过指定 Uid
func (s *BaseServer) broadcast(id uint32, data []byte, excludeUid int64, exclude bool) {
	if s.manager == nil {
		return
	}
	s.manager.ForEach(func(conn IConn) {
		if exclude && conn.Uid() == excludeUid {
			return
		}
		_ = conn.Send(id, data)
	})
}
