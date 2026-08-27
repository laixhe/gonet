package rpc

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

const SchemaName = "grpc"

// resolverOnce 保证 grpc:// scheme 在同一进程内只注册一次
var resolverOnce sync.Once

// Discovery 服务发现（resolver.Builder）
// 全局注册一次，只持有 etcd 客户端；每个目标（grpc://serviceName）由 Build 创建独立的 discoveryResolver。
type Discovery struct {
	etcdClient *clientv3.Client
	resolvers  sync.Map // 活跃的 *discoveryResolver → struct{}，Close 时统一回收 watcher
}

// DiscoveryOption 服务发现选项
type DiscoveryOption func(*discoveryOptions)

type discoveryOptions struct {
	etcdCfg clientv3.Config
}

// WithDiscoveryEtcdConfig 覆盖 etcd 客户端配置（TLS、用户名密码认证、超时等）。
// 未设置的字段沿用默认：endpoints 取 NewDiscovery 参数、DialTimeout 取 dialTimeout。
func WithDiscoveryEtcdConfig(cfg clientv3.Config) DiscoveryOption {
	return func(o *discoveryOptions) {
		o.etcdCfg = cfg
	}
}

// NewDiscovery 创建服务发现并注册到 gRPC resolver
// 同一进程内多次调用不会重复注册 scheme；首次创建的实例负责 grpc:// 目标的解析。
func NewDiscovery(etcdAddresses []string, dialTimeout time.Duration, opts ...DiscoveryOption) (*Discovery, error) {
	do := &discoveryOptions{}
	for _, opt := range opts {
		opt(do)
	}
	etcdCfg := do.etcdCfg
	if len(etcdCfg.Endpoints) == 0 {
		etcdCfg.Endpoints = etcdAddresses
	}
	if etcdCfg.DialTimeout == 0 {
		etcdCfg.DialTimeout = dialTimeout
	}
	etcdClient, err := clientv3.New(etcdCfg)
	if err != nil {
		return nil, err
	}
	s := &Discovery{
		etcdClient: etcdClient,
	}
	resolverOnce.Do(func() {
		resolver.Register(s)
	})
	return s, nil
}

// Build 为 grpc://serviceName 构建解析器，grpc.NewClient() 时调用
// 兼容 grpc://serviceName（serviceName 在 URL.Host）与 grpc:///serviceName（在 URL.Path）两种写法。
func (s *Discovery) Build(target resolver.Target, clientConn resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	ctx, cancel := context.WithCancel(context.Background())
	serviceName := target.Endpoint()
	if serviceName == "" {
		serviceName = target.URL.Host
	}
	r := &discoveryResolver{
		etcdClient: s.etcdClient,
		cc:         clientConn,
		prefix:     "/" + target.URL.Scheme + "/" + serviceName + "/",
		cancel:     cancel,
		discovery:  s,
	}
	s.resolvers.Store(r, struct{}{})
	// 初始全量同步：拉取现有实例并推送初始状态；失败则快速失败（不启动 watcher）
	rev, err := r.syncAll(ctx)
	if err != nil {
		s.resolvers.Delete(r)
		cancel()
		return nil, err
	}
	// 从初始同步的 revision 之后开始监听，保证 Get 与 Watch 之间注册的实例不会被漏掉。
	go r.watcher(ctx, rev)
	return r, nil
}

// Scheme return schema
func (s *Discovery) Scheme() string {
	return SchemaName
}

// Close 关闭 etcd 连接并取消所有活跃 watcher，整个 Discovery 不再使用时调用。
// 调用后该 Discovery 的 Build 将失败（etcd 客户端已关闭）。
func (s *Discovery) Close() error {
	s.resolvers.Range(func(k, _ any) bool {
		k.(*discoveryResolver).cancel()
		return true
	})
	return s.etcdClient.Close()
}

// discoveryResolver 单个服务（grpc://serviceName）的解析器，状态按目标隔离
type discoveryResolver struct {
	etcdClient *clientv3.Client
	cc         resolver.ClientConn
	prefix     string
	serverList sync.Map
	cancel     context.CancelFunc
	discovery  *Discovery // 所属 Discovery，Close 时注销自己
}

// ResolveNow 监视(watch)有变化调用
func (r *discoveryResolver) ResolveNow(_ resolver.ResolveNowOptions) {
}

// Close 解析器关闭时调用，取消 watcher 并从 Discovery 注销
func (r *discoveryResolver) Close() {
	if r.discovery != nil {
		r.discovery.resolvers.Delete(r)
	}
	r.cancel()
}

// watcher 监听服务前缀的变化；watch 异常中断（连接断开、revision 被压缩等）时
// 自动全量重同步并重建 watch，保证地址列表不会因瞬时故障而永久失效。
// rev: 从该 revision 之后开始监听，与上次 Get/Watch 衔接（rev+1），不漏事件。
func (r *discoveryResolver) watcher(ctx context.Context, rev int64) {
	backoff := time.Second
	for {
		rch := r.etcdClient.Watch(ctx, r.prefix, clientv3.WithPrefix(), clientv3.WithRev(rev+1), clientv3.WithCreatedNotify())
		lastRev, healthy := r.consumeWatch(ctx, rch)
		if ctx.Err() != nil {
			return // Close 触发，正常退出
		}
		if healthy > 0 {
			backoff = time.Second // watch 正常存活过一段时间，故障视为瞬时，重置退避
		}
		logf("rpc: discovery watch %s interrupted, resyncing", r.prefix)
		newRev, err := r.syncAll(ctx)
		if err != nil {
			logf("rpc: discovery resync %s failed: %v", r.prefix, err)
		} else {
			rev = newRev
		}
		if lastRev > rev {
			rev = lastRev
		}
		if !sleepCtx(ctx, backoff) {
			return
		}
		backoff = min(backoff*2, 30*time.Second)
	}
}

// consumeWatch 消费 watch 响应直到通道关闭（异常中断或 ctx 取消）。
// 返回最后一次响应的 revision 与已处理响应数（healthy=processed>0，
// 用于区分“建立后中断”与“建立即失败”）。
func (r *discoveryResolver) consumeWatch(ctx context.Context, rch clientv3.WatchChan) (int64, int) {
	var rev int64
	var processed int
	for n := range rch {
		processed++
		if n.Header != nil {
			rev = n.Header.Revision
		}
		if n.Canceled || n.Err() != nil {
			// 服务端取消/压缩（如 ErrCompacted）：watch 无法继续，交由 watcher 重同步
			logf("rpc: discovery watch %s canceled: %v", r.prefix, n.Err())
			return rev, processed
		}
		for _, ev := range n.Events {
			switch ev.Type {
			case mvccpb.PUT: // 新增或修改
				addr, attrs := parseServiceValue(string(ev.Kv.Value))
				r.setServiceList(string(ev.Kv.Key), resolver.Address{Addr: addr, Attributes: attrs})
			case mvccpb.DELETE: // 删除
				r.delServiceList(string(ev.Kv.Key))
			}
		}
		// 一个响应内的变更批量推送一次
		_ = r.cc.UpdateState(resolver.State{Addresses: r.getServices()})
	}
	if ctx.Err() != nil {
		return 0, processed // ctx 取消，正常退出
	}
	return rev, processed
}

// syncAll 全量拉取当前前缀下的所有注册实例，替换本地列表并推送一次状态。
// 返回响应 revision，供后续 Watch 衔接，保证 Get 与 Watch 之间的事件不丢。
func (r *discoveryResolver) syncAll(ctx context.Context) (int64, error) {
	resp, err := r.etcdClient.Get(ctx, r.prefix, clientv3.WithPrefix())
	if err != nil {
		return 0, err
	}
	// 清空后按最新快照重建，避免残留已下线实例
	r.serverList.Range(func(k, _ any) bool {
		r.serverList.Delete(k)
		return true
	})
	for _, kv := range resp.Kvs {
		addr, attrs := parseServiceValue(string(kv.Value))
		r.serverList.Store(string(kv.Key), resolver.Address{Addr: addr, Attributes: attrs})
	}
	if err := r.cc.UpdateState(resolver.State{Addresses: r.getServices()}); err != nil {
		return 0, err
	}
	return resp.Header.Revision, nil
}

// sleepCtx 带 ctx 取消的睡眠；ctx 取消时返回 false。
func sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

// parseServiceValue 解析 etcd 注册值：兼容旧格式（裸地址）与新格式（JSON）。
// 返回服务地址与附加属性（weight/metadata），新格式下 Attributes 非 nil。
func parseServiceValue(val string) (string, *attributes.Attributes) {
	var m serviceMeta
	if json.Unmarshal([]byte(val), &m) == nil && m.Addr != "" {
		attrs := attributes.New("weight", m.Weight)
		if len(m.Metadata) > 0 {
			attrs = attrs.WithValue("metadata", m.Metadata)
		}
		return m.Addr, attrs
	}
	return val, nil
}

// setServiceList 新增服务地址
func (r *discoveryResolver) setServiceList(key string, addr resolver.Address) {
	r.serverList.Store(key, addr)
}

// delServiceList 删除服务地址
func (r *discoveryResolver) delServiceList(key string) {
	r.serverList.Delete(key)
}

// getServices 获取服务地址
func (r *discoveryResolver) getServices() []resolver.Address {
	addresses := make([]resolver.Address, 0, 10)
	r.serverList.Range(func(k, v interface{}) bool {
		addresses = append(addresses, v.(resolver.Address))
		return true
	})
	return addresses
}
