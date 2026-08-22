package rpc

import (
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// defaultLBPolicy 默认负载均衡策略：round_robin 让多实例真正分流
const defaultLBPolicy = "round_robin"

// ClientOption 客户端选项
type ClientOption func(*clientOptions)

type clientOptions struct {
	opts     []grpc.DialOption
	lbPolicy string // "" 表示使用默认 round_robin
}

// WithClientOption 追加原始 grpc.DialOption（TLS、拦截器、codec 等）
func WithClientOption(opts ...grpc.DialOption) ClientOption {
	return func(o *clientOptions) {
		o.opts = append(o.opts, opts...)
	}
}

// WithLoadBalancingPolicy 设置负载均衡策略，默认 round_robin（pick_first / round_robin 等）
func WithLoadBalancingPolicy(policy string) ClientOption {
	return func(o *clientOptions) {
		o.lbPolicy = policy
	}
}

type Client struct {
	conn *grpc.ClientConn
}

// NewClient 创建明文 gRPC 客户端；addr 支持 grpc://serviceName 走服务发现
func NewClient(addr string, opts ...ClientOption) (*Client, error) {
	return newClient(addr, nil, opts...)
}

// NewClientTLS 创建 TLS gRPC 客户端，与服务端 NewServer(addr, certFile, keyFile) 对称。
// serverName 用于校验服务端证书（一般为服务端证书的 SAN/CN），不能为空，
// 否则 tls 握手时会 panic（"either ServerName or InsecureSkipVerify must be specified"）。
func NewClientTLS(addr, certFile, serverName string, opts ...ClientOption) (*Client, error) {
	if serverName == "" {
		return nil, errors.New("rpc: NewClientTLS: serverName must not be empty (used to verify server certificate)")
	}
	cred, err := credentials.NewClientTLSFromFile(certFile, serverName)
	if err != nil {
		return nil, err
	}
	return newClient(addr, cred, opts...)
}

func newClient(addr string, creds credentials.TransportCredentials, opts ...ClientOption) (*Client, error) {
	co := &clientOptions{}
	for _, opt := range opts {
		opt(co)
	}
	if creds == nil {
		creds = insecure.NewCredentials()
	}
	lb := co.lbPolicy
	if lb == "" {
		lb = defaultLBPolicy
	}
	dialOpts := make([]grpc.DialOption, 0, 5+len(co.opts))
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	dialOpts = append(dialOpts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy":%q}`, lb)))
	dialOpts = append(dialOpts, co.opts...)
	conn, err := grpc.NewClient(addr, dialOpts...)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}

// Conn 返回底层连接，供生成的客户端 stub（grpc.ClientConnInterface）使用
func (c *Client) Conn() grpc.ClientConnInterface {
	return c.conn
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.conn.Close()
}
