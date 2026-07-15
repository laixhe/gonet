package rpc

import (
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Server struct {
	addr   string
	server *grpc.Server
}

func NewServer(addr, certFile, keyFile string) (*Server, error) {
	serverOpts := make([]grpc.ServerOption, 0, 5)
	// serverOpts = append(serverOpts, grpc.UnaryInterceptor(nil))
	if certFile != "" && keyFile != "" {
		cred, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		serverOpts = append(serverOpts, grpc.Creds(cred))
	}
	return &Server{
		addr:   addr,
		server: grpc.NewServer(serverOpts...),
	}, nil
}

// RegisterService 注册服务
func (s *Server) RegisterService(desc *grpc.ServiceDesc, service any) {
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
