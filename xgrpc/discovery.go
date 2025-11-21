package xgrpc

import (
	"context"
	"sync"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

const SchemaName = "grpc"

// Discovery 服务发现
type Discovery struct {
	etcdClient *clientv3.Client
	grpcClient resolver.ClientConn
	serverList sync.Map
}

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
	resolver.Register(s)
	return s, nil
}

// Build 构建解析器 grpc.NewClient() 时调用
func (s *Discovery) Build(target resolver.Target, clientConn resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	s.grpcClient = clientConn
	prefix := "/" + target.URL.Scheme + "/" + target.Endpoint() + "/"
	// 根据前缀获取现有的key
	resp, err := s.etcdClient.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	for k := range resp.Kvs {
		_ = s.SetServiceList(string(resp.Kvs[k].Key), string(resp.Kvs[k].Value))
	}
	err = s.grpcClient.UpdateState(resolver.State{Addresses: s.getServices()})
	if err != nil {
		return nil, err
	}
	// 监听服务地址列表的变化
	go s.watcher(prefix)
	return s, nil
}

// ResolveNow 监视(watch)有变化调用
func (s *Discovery) ResolveNow(rn resolver.ResolveNowOptions) {
}

// Scheme return schema
func (s *Discovery) Scheme() string {
	return SchemaName
}

// Close 解析器关闭时调用
func (s *Discovery) Close() {
	_ = s.etcdClient.Close()
}

// watcher 监听前缀
func (s *Discovery) watcher(keyPrefix string) {
	rch := s.etcdClient.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	for n := range rch {
		for k := range n.Events {
			switch n.Events[k].Type {
			case mvccpb.PUT: // 新增或修改
				_ = s.SetServiceList(string(n.Events[k].Kv.Key), string(n.Events[k].Kv.Value))
			case mvccpb.DELETE: // 删除
				_ = s.DelServiceList(string(n.Events[k].Kv.Key))
			}
		}
	}
}

// SetServiceList 新增服务地址
func (s *Discovery) SetServiceList(key, val string) error {
	s.serverList.Store(key, resolver.Address{Addr: val})
	return s.grpcClient.UpdateState(resolver.State{Addresses: s.getServices()})
}

// DelServiceList 删除服务地址
func (s *Discovery) DelServiceList(key string) error {
	s.serverList.Delete(key)
	return s.grpcClient.UpdateState(resolver.State{Addresses: s.getServices()})
}

// GetServices 获取服务地址
func (s *Discovery) getServices() []resolver.Address {
	addresses := make([]resolver.Address, 0, 10)
	s.serverList.Range(func(k, v interface{}) bool {
		addresses = append(addresses, v.(resolver.Address))
		return true
	})
	return addresses
}
