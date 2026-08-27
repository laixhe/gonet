package rpc

import (
	"errors"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// ServerOption 服务端选项
type ServerOption func(*serverOptions)

type serverOptions struct {
	opts []grpc.ServerOption
}

// WithServerOption 追加原始 grpc.ServerOption（TLS、拦截器、codec 等）
func WithServerOption(opts ...grpc.ServerOption) ServerOption {
	return func(o *serverOptions) {
		o.opts = append(o.opts, opts...)
	}
}

type Server struct {
	addr    string
	server  *grpc.Server
	health  *health.Server
	started bool // Start 已调用（防重复启动）

	mu    sync.Mutex
	l     net.Listener // 实际监听器，Start 后可用
	errCh chan error   // ErrChan 返回的 channel（懒创建）
	err   error        // Serve 返回的错误（ErrChan 尚未调用时暂存）
}

// NewServer 创建 gRPC 服务端；certFile/keyFile 同时非空时启用 TLS。
// 自动注册 grpc_health_v1 健康检查服务，可通过 SetHealth 调整状态。
func NewServer(addr, certFile, keyFile string, opts ...ServerOption) (*Server, error) {
	so := &serverOptions{}
	for _, opt := range opts {
		opt(so)
	}
	serverOpts := make([]grpc.ServerOption, 0, 5+len(so.opts))
	if certFile != "" && keyFile != "" {
		cred, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		serverOpts = append(serverOpts, grpc.Creds(cred))
	}
	serverOpts = append(serverOpts, so.opts...)
	s := &Server{
		addr:   addr,
		server: grpc.NewServer(serverOpts...),
		health: health.NewServer(),
	}
	healthpb.RegisterHealthServer(s.server, s.health)
	return s, nil
}

// RegisterService 注册服务
func (s *Server) RegisterService(desc *grpc.ServiceDesc, service any) {
	s.server.RegisterService(desc, service)
}

// SetHealth 设置服务健康状态；service 传空字符串表示整体状态。
// 默认整体状态为 SERVING。设为 NOT_SERVING 后，健康检查客户端（如 LB、注册中心探活）会认为实例不可用。
func (s *Server) SetHealth(service string, status healthpb.HealthCheckResponse_ServingStatus) {
	s.health.SetServingStatus(service, status)
}

// Start 启动服务器；重复调用会返回错误。
func (s *Server) Start() error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return errors.New("rpc: server already started")
	}
	s.started = true
	s.mu.Unlock()

	tcpAddr, err := net.ResolveTCPAddr("tcp", s.addr)
	if err != nil {
		s.markNotStarted()
		return err
	}
	listener, err := net.Listen(tcpAddr.Network(), tcpAddr.String())
	if err != nil {
		s.markNotStarted()
		return err
	}
	s.mu.Lock()
	s.l = listener
	s.mu.Unlock()
	err = s.server.Serve(listener)
	s.recordErr(err)
	return err
}

// markNotStarted 启动准备阶段失败时复位 started，允许重试
func (s *Server) markNotStarted() {
	s.mu.Lock()
	s.started = false
	s.mu.Unlock()
}

// recordErr 记录 Serve 返回的错误；正常停止（ErrServerStopped）记为 nil。
func (s *Server) recordErr(err error) {
	if errors.Is(err, grpc.ErrServerStopped) {
		err = nil // GracefulStop/Stop 触发，属于正常停止
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.err = err
	if s.errCh != nil {
		s.errCh <- err // buffered(1) 且只发送一次，不会阻塞
	}
}

// ErrChan 返回服务终止时的通知 channel：Start 返回后收到一个错误
// （正常停止为 nil；启动/运行期致命错误为对应 error）。
// 适合 go srv.Start() 场景感知致命错误，无需自行包一层 goroutine。
func (s *Server) ErrChan() <-chan error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.errCh == nil {
		s.errCh = make(chan error, 1)
		if s.err != nil {
			s.errCh <- s.err
		}
	}
	return s.errCh
}

// Addr 返回实际监听地址；Start 调用前返回构造时传入的 addr。
// addr 为 127.0.0.1:0 时，Start 后可通过本方法获取实际分配的端口。
func (s *Server) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.l != nil {
		return s.l.Addr().String()
	}
	return s.addr
}

// Stop 停止服务器（GracefulStop，等待存量请求处理完）
func (s *Server) Stop() error {
	s.server.GracefulStop()
	return nil
}

// StopWithTimeout 优雅停止，最多等待 timeout；超时后强制停止（不再等待存量请求）。
// 适合退出流程：给存量请求一个宽限期，卡死的 RPC 不会阻塞进程退出。
func (s *Server) StopWithTimeout(timeout time.Duration) error {
	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
		return nil
	case <-time.After(timeout):
		s.server.Stop()
		<-stopped // 等 GracefulStop 协程返回，避免泄漏
		return nil
	}
}
