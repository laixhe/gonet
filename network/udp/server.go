package udp

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/laixhe/gonet/network"
	"github.com/laixhe/gonet/network/packet"
)

// udpMaxPacket 单个 UDP 数据报最大长度
const udpMaxPacket = 65535

// udpBufferSize UDP 收发缓冲区大小, 支撑大包传输
const udpBufferSize = 1 << 20 // 1MB

// Server UDP 服务器
type Server struct {
	mu        sync.Mutex // 保护 conn
	conn      *net.UDPConn
	router    *router
	stopChan  chan struct{}
	stopOnce  sync.Once
	rateLimit int                   // 单地址每秒消息数上限, 0 不限
	rates     map[string]*rateEntry // 按来源地址限流
	rateMu    sync.Mutex
}

// rateEntry 单地址限流窗口
type rateEntry struct {
	windowStart time.Time
	count       int
}

// NewServer 创建 UDP 服务器
func NewServer() *Server {
	return &Server{router: newRouter(), stopChan: make(chan struct{})}
}

// SetRateLimit 设置单地址每秒消息数上限, 超过上限的消息丢弃, 0 不限(默认), 需在 Start 前调用
func (s *Server) SetRateLimit(limit int) {
	s.rateMu.Lock()
	defer s.rateMu.Unlock()
	s.rateLimit = limit
	if limit > 0 && s.rates == nil {
		s.rates = make(map[string]*rateEntry)
	}
}

// allow 判断来源地址是否允许接收, 超过每秒上限返回 false
func (s *Server) allow(addr string, now time.Time) bool {
	if s.rateLimit <= 0 {
		return true
	}
	s.rateMu.Lock()
	defer s.rateMu.Unlock()
	e := s.rates[addr]
	if e == nil {
		e = &rateEntry{windowStart: now}
		s.rates[addr] = e
	}
	// 窗口滑动: 超过 1 秒重置计数
	if now.Sub(e.windowStart) >= time.Second {
		e.windowStart = now
		e.count = 0
	}
	e.count++
	// 惰性清理过期窗口, 防止来源地址无限增长
	if len(s.rates) > 10000 {
		for k, v := range s.rates {
			if now.Sub(v.windowStart) >= 2*time.Second {
				delete(s.rates, k)
			}
		}
	}
	return e.count <= s.rateLimit
}

// Router 注册消息路由
func (s *Server) Router(id uint32, handler Handler) {
	s.router.Register(id, handler)
}

// SetDefaultHandler 设置默认消息处理, 未注册的消息走该处理器
func (s *Server) SetDefaultHandler(handler DefaultHandler) {
	s.router.SetDefault(handler)
}

// Start 启动服务器, 阻塞运行, 停止后返回 nil
func (s *Server) Start(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	// 默认接收缓冲区可能小于大包尺寸导致丢包, 显式调大
	_ = conn.SetReadBuffer(udpBufferSize)
	_ = conn.SetWriteBuffer(udpBufferSize)
	s.mu.Lock()
	s.conn = conn
	s.mu.Unlock()
	// 启动期间已被停止, 关闭监听
	select {
	case <-s.stopChan:
		_ = conn.Close()
		return nil
	default:
	}
	return s.readLoop()
}

// Stop 关闭服务器
func (s *Server) Stop() error {
	s.stopOnce.Do(func() {
		close(s.stopChan)
		s.mu.Lock()
		conn := s.conn
		s.mu.Unlock()
		if conn != nil {
			_ = conn.Close()
		}
	})
	return nil
}

// Addr 返回实际监听地址, Start 前调用返回空串
func (s *Server) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == nil {
		return ""
	}
	return s.conn.LocalAddr().String()
}

// Reply 向指定地址发送消息
func (s *Server) Reply(addr *net.UDPAddr, id uint32, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == nil {
		return ErrClosed
	}
	packData, err := packet.Pack(packet.NewMessage(id, data))
	if err != nil {
		return err
	}
	_, err = s.conn.WriteToUDP(packData, addr)
	return err
}

// readLoop 读取并处理数据报
func (s *Server) readLoop() error {
	buf := make([]byte, udpMaxPacket)
	for {
		n, addr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-s.stopChan:
				return nil // 服务器已停止
			default:
			}
			if errors.Is(err, net.ErrClosed) {
				continue
			}
			network.Log().Errorf("udp read error: %s", err)
			continue
		}
		msg, err := packet.Unpack(buf[:n])
		if err != nil {
			network.Log().Warnf("udp unpack error from %s: %s", addr, err)
			continue
		}
		// 限流: 超过单地址每秒上限的消息丢弃
		if !s.allow(addr.String(), time.Now()) {
			network.Log().Warnf("udp rate limit exceeded from %s", addr)
			continue
		}
		s.router.dispatch(addr, msg.ID, msg.Data)
	}
}
