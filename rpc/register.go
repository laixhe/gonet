package rpc

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// RegisterOption 注册选项
type RegisterOption func(*registerOptions)

type registerOptions struct {
	weight    int
	weightSet bool
	metadata  map[string]string
}

// WithWeight 设置实例权重（写入 etcd 注册信息，供权重感知的负载均衡使用）
func WithWeight(weight int) RegisterOption {
	return func(o *registerOptions) {
		o.weight = weight
		o.weightSet = true
	}
}

// WithMetadata 设置实例附加元数据（如机房、版本等，供灰度发布等场景使用）
func WithMetadata(metadata map[string]string) RegisterOption {
	return func(o *registerOptions) {
		o.metadata = metadata
	}
}

// Register 注册服务
type Register struct {
	etcdClient    *clientv3.Client
	leaseID       clientv3.LeaseID                        // 租约ID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse // 租约 KeepAlive 响应 chan
	key           string
	val           string

	mu     sync.Mutex
	closed bool

	keepAliveCancel context.CancelFunc // 自动续租 goroutine 的取消
	keepAliveWG     sync.WaitGroup
}

// NewRegister 注册服务并自动启动租约续租。
// 注册 key：/grpc/{serviceName}/{serverAddr}；value 无附加信息时为裸地址（兼容旧格式），
// 带 weight/metadata 时为 JSON。
func NewRegister(endpoints []string, serviceName, serverAddr string, leaseTtl int64, opts ...RegisterOption) (*Register, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	ro := &registerOptions{}
	for _, opt := range opts {
		opt(ro)
	}

	ser := &Register{
		etcdClient: etcdClient,
		key:        "/" + SchemaName + "/" + serviceName + "/" + serverAddr,
		val:        serviceValue(serverAddr, ro),
	}

	// 申请租约并绑定服务地址
	if err = ser.putKeyWithLease(leaseTtl); err != nil {
		_ = etcdClient.Close()
		return nil, err
	}
	// 自动启动续租
	ser.startKeepAlive()
	return ser, nil
}

// serviceValue 生成 etcd 注册值：无附加信息时返回裸地址（保持与旧版一致），否则返回 JSON。
func serviceValue(addr string, ro *registerOptions) string {
	if !ro.weightSet && len(ro.metadata) == 0 {
		return addr
	}
	b, _ := json.Marshal(serviceMeta{
		Addr:     addr,
		Weight:   ro.weight,
		Metadata: ro.metadata,
	})
	return string(b)
}

// serviceMeta 注册值的结构化格式
type serviceMeta struct {
	Addr     string            `json:"addr"`
	Weight   int               `json:"weight,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// startKeepAlive 启动自动续租监听 goroutine
func (s *Register) startKeepAlive() {
	ctx, cancel := context.WithCancel(context.Background())
	s.keepAliveCancel = cancel
	s.keepAliveWG.Add(1)
	go func() {
		defer s.keepAliveWG.Done()
		s.runKeepAlive(ctx)
	}()
}

// runKeepAlive 监听续租，续租通道关闭（续租失败）时撤销租约；ctx 取消（Close）时退出
func (s *Register) runKeepAlive(ctx context.Context) {
	for {
		select {
		case _, ok := <-s.keepAliveChan:
			if !ok {
				// 主动 Close 后通道也会关闭，此时无需再撤销租约
				if !s.isClosed() {
					if err := s.revoke(ctx); err != nil {
						log.Println("revoke:", err)
					}
				}
				log.Println("续租通道关闭，停止续租")
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// ListenLease 兼容旧调用：NewRegister 已自动启动续租监听，本方法为 no-op。
//
// Deprecated: 续租已由 NewRegister 自动管理，无需再调用。
func (s *Register) ListenLease(_ context.Context) {}

// putKeyWithLease 设置租约并注册服务
func (s *Register) putKeyWithLease(leaseTtl int64) error {
	// 创建租约
	leaseResp, err := s.etcdClient.Grant(context.Background(), leaseTtl)
	if err != nil {
		return err
	}
	// 注册服务并绑定租约(将服务地址注册到 etcd 中)
	if _, err = s.etcdClient.Put(context.Background(), s.key, s.val, clientv3.WithLease(leaseResp.ID)); err != nil {
		return err
	}
	// 建立长连接，定期续租
	leaseRespChan, err := s.etcdClient.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return err
	}
	s.leaseID = leaseResp.ID
	s.keepAliveChan = leaseRespChan
	return nil
}

// revoke 取消租约
func (s *Register) revoke(ctx context.Context) error {
	_, err := s.etcdClient.Revoke(ctx, s.leaseID)
	return err
}

// Close 注销服务（幂等，可重复调用）；自动停止续租并撤销租约
func (s *Register) Close() error {
	if !s.markClosed() {
		return nil
	}
	// 停止自动续租
	if s.keepAliveCancel != nil {
		s.keepAliveCancel()
	}
	s.keepAliveWG.Wait()
	// 撤销租约并关闭连接（revoke 失败也要关闭 etcd 连接）
	_, revokeErr := s.etcdClient.Revoke(context.Background(), s.leaseID)
	closeErr := s.etcdClient.Close()
	if revokeErr != nil {
		return revokeErr
	}
	return closeErr
}

// markClosed 原子标记已关闭，返回是否由本次调用完成关闭
func (s *Register) markClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	s.closed = true
	return true
}

// isClosed 是否已关闭
func (s *Register) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}
