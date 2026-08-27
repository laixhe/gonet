package rpc

import (
	"context"
	"encoding/json"
	"errors"
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
	etcdCfg   clientv3.Config
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

// WithRegisterEtcdConfig 覆盖 etcd 客户端配置（TLS、用户名密码认证、超时等）。
// 未设置的字段沿用默认：endpoints 取 NewRegister 参数、DialTimeout 5s。
func WithRegisterEtcdConfig(cfg clientv3.Config) RegisterOption {
	return func(o *registerOptions) {
		o.etcdCfg = cfg
	}
}

// Register 注册服务
type Register struct {
	etcdClient    *clientv3.Client
	leaseID       clientv3.LeaseID                        // 租约ID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse // 租约 KeepAlive 响应 chan
	key           string
	val           string
	addr          string
	weight        int
	weightSet     bool
	metadata      map[string]string
	leaseTtl      int64

	mu     sync.Mutex // 保护 closed / val / weight / metadata / leaseID / keepAliveChan
	closed bool

	keepAliveCancel context.CancelFunc // 自动续租 goroutine 的取消
	keepAliveWG     sync.WaitGroup
}

// NewRegister 注册服务并自动启动租约续租。
// 注册 key：/grpc/{serviceName}/{serverAddr}；value 无附加信息时为裸地址（兼容旧格式），
// 带 weight/metadata 时为 JSON。
func NewRegister(endpoints []string, serviceName, serverAddr string, leaseTtl int64, opts ...RegisterOption) (*Register, error) {
	ro := &registerOptions{}
	for _, opt := range opts {
		opt(ro)
	}

	etcdCfg := ro.etcdCfg
	if len(etcdCfg.Endpoints) == 0 {
		etcdCfg.Endpoints = endpoints
	}
	if etcdCfg.DialTimeout == 0 {
		etcdCfg.DialTimeout = 5 * time.Second
	}
	etcdClient, err := clientv3.New(etcdCfg)
	if err != nil {
		return nil, err
	}

	ser := &Register{
		etcdClient: etcdClient,
		key:        "/" + SchemaName + "/" + serviceName + "/" + serverAddr,
		val:        serviceValue(serverAddr, ro),
		addr:       serverAddr,
		weight:     ro.weight,
		weightSet:  ro.weightSet,
		metadata:   ro.metadata,
		leaseTtl:   leaseTtl,
	}

	// 申请租约并绑定服务地址
	if err = ser.putKeyWithLease(context.Background(), leaseTtl); err != nil {
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

// runKeepAlive 监听续租响应。续租通道异常关闭（租约被撤销/过期、etcd 长时间不可用）
// 时自动重新注册服务（Grant + Put + KeepAlive，带指数退避重试），
// 避免实例因瞬时故障从注册中心永久消失；ctx 取消（Close）时退出。
func (s *Register) runKeepAlive(ctx context.Context) {
	for {
		ch := s.keepAliveChan
		select {
		case _, ok := <-ch:
			if ok {
				continue // 正常心跳
			}
			if s.isClosed() {
				logf("rpc: keepalive stopped (closed)")
				return
			}
			// 通道异常关闭：续租中断，重新注册（内部退避重试，直到成功或 ctx 取消）
			logf("rpc: keepalive channel closed unexpectedly, re-registering")
			if err := s.reRegister(ctx); err != nil {
				logf("rpc: re-register aborted: %v", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// reRegister 在续租中断后重新申请租约并注册服务，内部带指数退避重试，
// 直到成功或 ctx 取消。注册成功后尽力撤销旧租约（若仍有效），避免残留。
func (s *Register) reRegister(ctx context.Context) error {
	oldLease := s.leaseID
	backoff := time.Second
	for {
		if err := s.putKeyWithLease(ctx, s.leaseTtl); err != nil {
			logf("rpc: re-register failed: %v", err)
			select {
			case <-ctx.Done():
				return err
			case <-time.After(backoff):
			}
			backoff = min(backoff*2, 10*time.Second)
			continue
		}
		// 新租约已生效；撤销旧租约（若不同且仍有效）
		if oldLease != 0 && oldLease != s.leaseID {
			if _, err := s.etcdClient.Revoke(ctx, oldLease); err != nil {
				logf("rpc: revoke old lease %d: %v", oldLease, err)
			}
		}
		logf("rpc: re-register success")
		return nil
	}
}

// ListenLease 兼容旧调用：NewRegister 已自动启动续租监听，本方法为 no-op。
//
// Deprecated: 续租已由 NewRegister 自动管理，无需再调用。
func (s *Register) ListenLease(_ context.Context) {}

// putKeyWithLease 设置租约并注册服务
func (s *Register) putKeyWithLease(ctx context.Context, leaseTtl int64) error {
	// 创建租约
	leaseResp, err := s.etcdClient.Grant(ctx, leaseTtl)
	if err != nil {
		return err
	}
	s.mu.Lock()
	val := s.val
	s.mu.Unlock()
	// 注册服务并绑定租约(将服务地址注册到 etcd 中)
	if _, err = s.etcdClient.Put(ctx, s.key, val, clientv3.WithLease(leaseResp.ID)); err != nil {
		return err
	}
	// 建立长连接，定期续租
	leaseRespChan, err := s.etcdClient.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.leaseID = leaseResp.ID
	s.keepAliveChan = leaseRespChan
	s.mu.Unlock()
	return nil
}

// UpdateWeight 热更新实例权重（写回 etcd 注册值，租约与连接不受影响）。
// 灰度发布等场景可在不重启实例的情况下调整分流比例。
func (s *Register) UpdateWeight(weight int) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return errors.New("rpc: register closed")
	}
	newVal := serviceValue(s.addr, &registerOptions{weight: weight, weightSet: true, metadata: s.metadata})
	s.mu.Unlock()
	return s.reput(newVal, func() {
		s.mu.Lock()
		s.weight = weight
		s.weightSet = true
		s.val = newVal
		s.mu.Unlock()
	})
}

// UpdateMetadata 热更新实例元数据（如机房、版本标签，供灰度/路由使用）。
func (s *Register) UpdateMetadata(metadata map[string]string) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return errors.New("rpc: register closed")
	}
	newVal := serviceValue(s.addr, &registerOptions{weight: s.weight, weightSet: s.weightSet, metadata: metadata})
	s.mu.Unlock()
	return s.reput(newVal, func() {
		s.mu.Lock()
		s.metadata = metadata
		s.val = newVal
		s.mu.Unlock()
	})
}

// reput 以当前租约重新写入注册值；成功后才提交本地状态（commit）。
func (s *Register) reput(newVal string, commit func()) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s.mu.Lock()
	_, err := s.etcdClient.Put(ctx, s.key, newVal, clientv3.WithLease(s.leaseID))
	s.mu.Unlock()
	if err != nil {
		return err
	}
	commit()
	return nil
}

// IsRegistered 查询实例当前是否仍注册在 etcd（key 存在且未过期）。
// 可用于启动自检或监控探活；Close 后 etcd 客户端已关闭，调用会返回错误。
func (s *Register) IsRegistered(ctx context.Context) (bool, error) {
	resp, err := s.etcdClient.Get(ctx, s.key)
	if err != nil {
		return false, err
	}
	return len(resp.Kvs) > 0, nil
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
	// 撤销租约并关闭连接（revoke 失败也要关闭 etcd 连接）；
	// revoke 加超时，etcd 不可用时 Close 也不会永久阻塞（租约将随 TTL 自然过期）
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, revokeErr := s.etcdClient.Revoke(ctx, s.leaseID)
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
