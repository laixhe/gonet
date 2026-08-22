package rpc

import (
	"net"
	"sync"

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
	addr   string
	server *grpc.Server
	health *health.Server

	mu sync.Mutex
	l  net.Listener // 实际监听器，Start 后可用
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

// Start 启动服务器
func (s *Server) Start() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", s.addr)
	if err != nil {
		return err
	}
	listener, err := net.Listen(tcpAddr.Network(), tcpAddr.String())
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.l = listener
	s.mu.Unlock()
	return s.server.Serve(listener)
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

// Stop 停止服务器
func (s *Server) Stop() error {
	s.server.GracefulStop()
	return nil
}
