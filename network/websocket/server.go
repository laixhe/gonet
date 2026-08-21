package websocket

import (
	"errors"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/laixhe/gonet/network"
)

// server WebSocket 服务器, 实现 network.IServer
type server struct {
	config              Config              // 服务器配置
	mu                  sync.Mutex          // 保护 listener/httpServer
	listener            net.Listener        // 监听器
	httpServer          *http.Server        // HTTP 服务器
	manager             *network.MapManager // 连接管理器
	upgrader            websocket.Upgrader  // 升级器
	stopChan            chan struct{}       // 停止信号
	stopOnce            sync.Once           // 停止幂等
	*network.BaseServer                     // 路由与连接事件回调
}

var _ network.IServer = &server{}

// NewServer 创建默认配置的 WebSocket 服务器
func NewServer() network.IServer {
	return NewServerWithConfig(DefaultConfig())
}

// NewServerWithConfig 创建指定配置的 WebSocket 服务器
func NewServerWithConfig(config Config) network.IServer {
	config.Check()
	checkOrigin := config.CheckOrigin
	if checkOrigin == nil {
		checkOrigin = func(*http.Request) bool { return true }
	}
	s := &server{config: config, BaseServer: network.NewBaseServer()}
	s.manager = network.NewMapManager(config.MaxConnections,
		func(c network.IConn) { c.(*Connection).Start() },
		s.BaseServer.NotifyConnect,
		s.BaseServer.NotifyDisconnect,
	)
	s.SetManager(s.manager)
	s.stopChan = make(chan struct{})
	s.upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin:     checkOrigin,
	}
	return s
}

// handleWS 处理 WebSocket 升级
func (s *server) handleWS(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != s.config.Path {
		http.NotFound(w, r)
		return
	}
	raw, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		network.Log().Warnf("websocket upgrade error: %s", err)
		return
	}
	c := &Connection{}
	c.init(raw, s.manager, s.manager.NextID(), s.Dispatch, s.config.HeartbeatInterval, s.config.HeartbeatTimeout, s.config.WriteTimeout)
	if err := s.manager.Add(c); err != nil {
		network.Log().Errorf("websocket manager add error: %s", err)
	}
}

// Start 启动服务器, 阻塞运行直到 Stop
func (s *server) Start(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.listener = ln
	s.httpServer = &http.Server{Handler: http.HandlerFunc(s.handleWS), TLSConfig: s.config.TLS}
	s.mu.Unlock()

	errCh := make(chan error, 1)
	go func() {
		if s.config.TLS != nil {
			// wss: 使用配置中的证书
			errCh <- s.httpServer.ServeTLS(ln, "", "")
		} else {
			errCh <- s.httpServer.Serve(ln)
		}
	}()

	select {
	case <-s.stopChan:
		s.mu.Lock()
		hs := s.httpServer
		s.mu.Unlock()
		if hs != nil {
			_ = hs.Close()
		}
		return nil
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

// Stop 关闭服务器
func (s *server) Stop() error {
	s.stopOnce.Do(func() {
		close(s.stopChan)
		s.manager.Close()
		s.mu.Lock()
		ln := s.listener
		hs := s.httpServer
		s.mu.Unlock()
		if ln != nil {
			_ = ln.Close()
		}
		if hs != nil {
			_ = hs.Close()
		}
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
