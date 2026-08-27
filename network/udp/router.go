// Package udp 提供基于 UDP 的服务器与客户端实现。
//
// 与 TCP 的区别:
//   - 无连接: 服务器通过数据报的来源地址识别对端, 回复需指定目标地址
//   - 保留消息边界: 一个数据报即一条消息, 无需粘包/拆包
//   - 不可靠: 可能丢包/乱序, 可靠性由业务层保证
package udp

import (
	"net"
	"sync"
)

// Handler UDP 消息处理函数
type Handler func(addr *net.UDPAddr, data []byte)

// DefaultHandler 默认消息处理函数, 未注册的消息走该处理器
type DefaultHandler func(addr *net.UDPAddr, id uint32, data []byte)

// router UDP 消息路由
type router struct {
	mu       sync.RWMutex
	handlers map[uint32]Handler
	defaultH DefaultHandler
}

func newRouter() *router {
	return &router{handlers: make(map[uint32]Handler)}
}

// Register 注册消息路由
func (r *router) Register(id uint32, handler Handler) {
	r.mu.Lock()
	r.handlers[id] = handler
	r.mu.Unlock()
}

// SetDefault 设置默认消息处理, 未注册的消息走该处理器
func (r *router) SetDefault(handler DefaultHandler) {
	r.mu.Lock()
	r.defaultH = handler
	r.mu.Unlock()
}

// dispatch 分发消息
func (r *router) dispatch(addr *net.UDPAddr, id uint32, data []byte) {
	r.mu.RLock()
	h := r.handlers[id]
	if h == nil && r.defaultH != nil {
		r.defaultH(addr, id, data)
		r.mu.RUnlock()
		return
	}
	r.mu.RUnlock()
	if h != nil {
		h(addr, data)
	}
}
