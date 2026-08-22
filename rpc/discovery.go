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
}

// NewDiscovery 创建服务发现并注册到 gRPC resolver
// 同一进程内多次调用不会重复注册 scheme；首次创建的实例负责 grpc:// 目标的解析。
func NewDiscovery(etcdAddresses []string, dialTimeout time.Duration) (*Discovery, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdAddresses,
		DialTimeout: dialTimeout,
	})
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
	}
	// 根据前缀批量获取现有的key
	resp, err := s.etcdClient.Get(ctx, r.prefix, clientv3.WithPrefix())
	if err != nil {
		cancel()
		return nil, err
	}
	for _, kv := range resp.Kvs {
		addr, attrs := parseServiceValue(string(kv.Value))
		r.serverList.Store(string(kv.Key), resolver.Address{Addr: addr, Attributes: attrs})
	}
	// 初始地址一次性推送
	if err = r.cc.UpdateState(resolver.State{Addresses: r.getServices()}); err != nil {
		cancel()
		return nil, err
	}
	// 监听服务地址列表的变化；从 Get 的 revision 之后开始监听，
	// 保证 Get 与 Watch 之间注册的实例不会被漏掉。
	go r.watcher(ctx, resp.Header.Revision)
	return r, nil
}

// Scheme return schema
func (s *Discovery) Scheme() string {
	return SchemaName
}

// Close 关闭 etcd 连接，整个 Discovery 不再使用时调用
func (s *Discovery) Close() error {
	return s.etcdClient.Close()
}

// discoveryResolver 单个服务（grpc://serviceName）的解析器，状态按目标隔离
type discoveryResolver struct {
	etcdClient *clientv3.Client
	cc         resolver.ClientConn
	prefix     string
	serverList sync.Map
	cancel     context.CancelFunc
}

// ResolveNow 监视(watch)有变化调用
func (r *discoveryResolver) ResolveNow(_ resolver.ResolveNowOptions) {
}

// Close 解析器关闭时调用，取消 watcher
func (r *discoveryResolver) Close() {
	r.cancel()
}

// watcher 监听前缀，ctx 取消时退出。
// rev: 从该 revision 之后开始监听，与 Build 里的初始 Get 衔接（rev+1），不漏事件。
// WithCreatedNotify: 让服务端在 watch 建立时先发一个 created 响应，
// 确认监听已建立（否则空前缀下永远不会有初始响应）。
func (r *discoveryResolver) watcher(ctx context.Context, rev int64) {
	rch := r.etcdClient.Watch(ctx, r.prefix, clientv3.WithPrefix(), clientv3.WithRev(rev+1), clientv3.WithCreatedNotify())
	for n := range rch {
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
