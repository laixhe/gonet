package xgrpc

import (
	"net"

	goGrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Server struct {
	addr   string
	server *goGrpc.Server
}

func NewServer(addr, certFile, keyFile string) (*Server, error) {
	serverOpts := make([]goGrpc.ServerOption, 0, 5)
	// serverOpts = append(serverOpts, goGrpc.UnaryInterceptor(nil))
	if certFile != "" && keyFile != "" {
		cred, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		serverOpts = append(serverOpts, goGrpc.Creds(cred))
	}
	return &Server{
		addr:   addr,
		server: goGrpc.NewServer(serverOpts...),
	}, nil
}

// RegisterService 注册服务
func (s *Server) RegisterService(desc *goGrpc.ServiceDesc, service any) {
	s.server.RegisterService(desc, service)
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
	return s.server.Serve(listener)
}

// Stop 停止服务器
func (s *Server) Stop() error {
	s.server.GracefulStop()
	return nil
}
