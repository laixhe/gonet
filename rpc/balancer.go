package rpc

import (
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/endpointsharding"
	"google.golang.org/grpc/balancer/pickfirst"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/resolver"
)

// WeightedRoundRobinName 权重感知轮询负载均衡策略名。
// 按实例注册的 weight（默认 1）比例轮流分配请求，例如 weight 3 的实例每 4 个请求约收到 3 个。
//
//	c, _ := rpc.NewClient("grpc://svc", rpc.WithLoadBalancingPolicy(rpc.WeightedRoundRobinName))
const WeightedRoundRobinName = "weighted_round_robin"

func init() {
	balancer.Register(weightedRRBuilder{})
}

// weightedRRBuilder 加权轮询 balancer 构造器
type weightedRRBuilder struct{}

func (weightedRRBuilder) Name() string { return WeightedRoundRobinName }

func (weightedRRBuilder) Build(cc balancer.ClientConn, opts balancer.BuildOptions) balancer.Balancer {
	b := &weightedRRBalancer{
		ClientConn: cc,
	}
	// 每个 endpoint 一个 pick_first 子 balancer，由 endpointsharding 统一管理子连接与健康状态
	b.child = endpointsharding.NewBalancer(b, opts, balancer.Get(pickfirst.Name).Build, endpointsharding.Options{})
	return b
}

// weightedRRBalancer 加权轮询：把 endpointsharding 汇报的 READY 子端点按注册权重打包成 picker。
type weightedRRBalancer struct {
	balancer.ClientConn // 拦截子 balancer 的 UpdateState
	child               balancer.Balancer
}

func (b *weightedRRBalancer) UpdateClientConnState(ccs balancer.ClientConnState) error {
	return b.child.UpdateClientConnState(balancer.ClientConnState{
		// 开启客户端健康检查（服务端未实现 grpc_health_v1 时自动降级为不检查）
		ResolverState: pickfirst.EnableHealthListener(ccs.ResolverState),
	})
}

func (b *weightedRRBalancer) ResolverError(err error) {
	b.child.ResolverError(err)
}

// UpdateSubConnState 子连接由 endpointsharding 的 state listener 管理，无需处理
func (b *weightedRRBalancer) UpdateSubConnState(balancer.SubConn, balancer.SubConnState) {}

func (b *weightedRRBalancer) ExitIdle() { b.child.ExitIdle() }

func (b *weightedRRBalancer) Close() { b.child.Close() }

// UpdateState 子 balancer 汇报状态：READY 子端点按权重组装成加权轮询 picker 后转发。
func (b *weightedRRBalancer) UpdateState(state balancer.State) {
	b.ClientConn.UpdateState(balancer.State{
		ConnectivityState: state.ConnectivityState,
		Picker:            newWeightedRRPicker(state.Picker),
	})
}

// newWeightedRRPicker 从 endpointsharding 的 picker 提取各 READY 子端点并按权重构建 picker。
// 无 READY 端点时透传原 picker（由其处理 Connecting/TransientFailure 等错误提示）。
func newWeightedRRPicker(shardingPicker balancer.Picker) balancer.Picker {
	children := endpointsharding.ChildStatesFromPicker(shardingPicker)
	var entries []*swrrEntry
	for _, cs := range children {
		if cs.State.ConnectivityState != connectivity.Ready {
			continue
		}
		w := endpointWeight(cs.Endpoint)
		if w <= 0 {
			w = 1 // 未设置权重时按 1 处理
		}
		entries = append(entries, &swrrEntry{picker: cs.State.Picker, weight: w})
	}
	if len(entries) == 0 {
		return shardingPicker
	}
	return &weightedRRPicker{entries: entries}
}

// endpointWeight 读取 endpoint 首个地址上的 weight 属性（注册时 WithWeight 写入，int 类型）
func endpointWeight(ep resolver.Endpoint) int {
	if len(ep.Addresses) == 0 {
		return 0
	}
	if w, ok := ep.Addresses[0].Attributes.Value("weight").(int); ok {
		return w
	}
	return 0
}

// swrrEntry 平滑加权轮询（SWRR，nginx 同款算法）的单个端点
type swrrEntry struct {
	picker balancer.Picker
	weight int
	cur    int // 当前累计权重
}

// weightedRRPicker 平滑加权轮询 picker：权重 3:1 时输出 A A B A A B...，
// 长时间看比例精确，且不会出现连续打满高权重端点的抖峰。
type weightedRRPicker struct {
	mu      sync.Mutex
	entries []*swrrEntry
}

func (p *weightedRRPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.entries) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var best *swrrEntry
	total := 0
	for _, e := range p.entries {
		e.cur += e.weight
		total += e.weight
		if best == nil || e.cur > best.cur {
			best = e
		}
	}
	best.cur -= total
	return best.picker.Pick(info)
}
