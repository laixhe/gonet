package tcp

import (
	"crypto/tls"
	"errors"
	"net"
	"sync"

	"github.com/laixhe/gonet/network"
)

type server struct {
	config              Config        // 服务器配置
	mu                  sync.Mutex    // 保护 listener
	listener            net.Listener  // 监听器
	manager             *manager      // 连接管理器
	stopChan            chan struct{} // 停止信号
	stopOnce            sync.Once     // 停止幂等
	*network.BaseServer               // 路由与连接事件回调
}

var _ network.IServer = &server{}

// NewServer 创建默认配置的 TCP 服务器
func NewServer() network.IServer {
	return NewServerWithConfig(DefaultConfig())
}

// NewServerWithConfig 创建指定配置的 TCP 服务器
func NewServerWithConfig(config Config) network.IServer {
	config.Check()
	s := &server{config: config, BaseServer: network.NewBaseServer()}
	s.manager = newManager(s, config)
	s.SetManager(s.manager)
	s.stopChan = make(chan struct{})
	return s
}

// init 初始化监听器
func (s *server) init(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// 启用 TLS
	if s.config.TLS != nil {
		listener = tls.NewListener(listener, s.config.TLS)
	}

	s.mu.Lock()
	s.listener = listener
	s.mu.Unlock()
	return nil
}

// accept 等待连接
func (s *server) accept() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isStopped() {
				return nil // 服务器已停止
			}
			if errors.Is(err, net.ErrClosed) {
				continue
			}
			var e net.Error
			if errors.As(err, &e) && e.Timeout() {
				continue
			}
			network.Log().Errorf("tcp listener accept error: %s", err)
			continue
		}
		// 创建连接并注册到管理器
		c := &Connection{}
		c.init(conn, s.manager, s.manager.nextID(), s.Dispatch, s.config.HeartbeatInterval, s.config.HeartbeatTimeout, s.config.WriteTimeout, s.config.ProcessWorkers)
		if err := s.manager.Add(c); err != nil {
			network.Log().Errorf("tcp manager add error: %s", err)
		}
	}
}

// isStopped 是否已停止
func (s *server) isStopped() bool {
	select {
	case <-s.stopChan:
		return true
	default:
		return false
	}
}

// Start 启动服务器
func (s *server) Start(addr string) error {
	if err := s.init(addr); err != nil {
		return err
	}
	// 启动期间已被停止, 关闭监听器
	if s.isStopped() {
		s.mu.Lock()
		ln := s.listener
		s.mu.Unlock()
		if ln != nil {
			_ = ln.Close()
		}
		return nil
	}
	return s.accept()
}

// Stop 关闭服务器
func (s *server) Stop() error {
	s.stopOnce.Do(func() {
		close(s.stopChan)
		s.mu.Lock()
		ln := s.listener
		s.mu.Unlock()
		if ln != nil {
			_ = ln.Close()
		}
		s.manager.Close()
	})
	return nil
}

// Addr 返回实际监听地址, Start 前调用返回空串
func (s *server) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

// GetManager 获取连接管理器
func (s *server) GetManager() network.IManager {
	return s.manager
}
