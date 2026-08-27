package rpc

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"
)

// defaultLBPolicy 默认负载均衡策略：round_robin 让多实例真正分流
const defaultLBPolicy = "round_robin"

// defaultKeepalive 默认客户端 keepalive：空闲时也发送 ping（PermitWithoutStream），
// 网络半开/对端进程崩溃时能及时感知连接死亡（默认 gRPC 不发送 keepalive）。
// 可通过 WithClientOption 传入 grpc.WithKeepaliveParams 覆盖（后传入者优先生效）。
var defaultKeepalive = keepalive.ClientParameters{
	Time:                10 * time.Second,
	Timeout:             3 * time.Second,
	PermitWithoutStream: true,
}

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
	// 友好提示：grpc:// 目标依赖进程内已注册的 Discovery（resolverOnce 只注册一次）
	if u, err := url.Parse(addr); err == nil && u.Scheme == SchemaName && resolver.Get(SchemaName) == nil {
		return nil, fmt.Errorf("rpc: target %q uses %q scheme but no Discovery registered; call rpc.Init or rpc.NewDiscovery first", addr, SchemaName)
	}
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
	dialOpts := make([]grpc.DialOption, 0, 6+len(co.opts))
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	dialOpts = append(dialOpts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy":%q}`, lb)))
	dialOpts = append(dialOpts, grpc.WithKeepaliveParams(defaultKeepalive))
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

// Invoke 直接发起一次一元 RPC（无需生成 stub 即可调用）。
// 例：c.Invoke(ctx, "/pb.Greeter/SayHello", req, &resp)
func (c *Client) Invoke(ctx context.Context, method string, args, reply any, opts ...grpc.CallOption) error {
	return c.conn.Invoke(ctx, method, args, reply, opts...)
}

// WaitForReady 等待连接进入 READY 状态（或 ctx 取消/超时）。
// 会主动触发连接建立（gRPC 默认惰性连接，首个 RPC 前不会自动拨号），
// 常用于进程启动时等待服务发现完成后再发起首批请求。
func (c *Client) WaitForReady(ctx context.Context) error {
	c.conn.Connect() // 幂等：已连接/连接中时为 no-op
	for {
		st := c.conn.GetState()
		if st == connectivity.Ready {
			return nil
		}
		if !c.conn.WaitForStateChange(ctx, st) {
			return ctx.Err()
		}
	}
}

// State 返回当前连接状态（connectivity.State）：
// Idle / Connecting / Ready / TransientFailure / Shutdown。
func (c *Client) State() connectivity.State {
	return c.conn.GetState()
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.conn.Close()
}
