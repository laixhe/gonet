package tcp

import (
	"errors"
	"log"
	"net"

	"github.com/laixhe/gonet/network"
)

type server struct {
	listener *net.TCPListener // 监听器
	manager  network.IManager // 连接管理器
}

var _ network.IServer = &server{}

func NewServer() network.IServer {
	s := &server{}
	s.manager = newManager(s)
	return s
}

// init 初始化 TCP 服务器
func (s *server) init(addr string) error {
	// 获取一个 TCP 的 Addr
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	// 监听服务器地址
	listener, err := net.ListenTCP(tcpAddr.Network(), tcpAddr)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	s.listener = listener
	return nil
}

// accept 等待连接
func (s *server) accept() error {
	// 主协程，循环阻塞待用户链接
	for {
		// 阻塞等待客户端建立连接请求
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Printf("tcp listener accept closed error: %s", err)
				continue
			}
			var e net.Error
			if errors.As(err, &e) && e.Timeout() {
				log.Printf("tcp listener accept timeout error: %s", err)
				continue
			}
			log.Printf("tcp listener accept error: %s", err)
			continue
		}
		// 处理用户链接
		_ = s.manager.Add(conn)
	}
}

// Start 启动服务器
func (s *server) Start(addr string) error {
	if err := s.init(addr); err != nil {
		return err
	}
	return s.accept()
}

// Stop 关闭服务器
func (s *server) Stop() error {
	return nil
}

// GetManager 获取连接管理器
func (s *server) GetManager() network.IManager {
	return s.manager
}

// RouterPath 路由路径
func (s *server) RouterPath(path string) {
}
