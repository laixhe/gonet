# gonet/rpc

基于 **gRPC + etcd** 的轻量微服务 RPC 框架：服务注册、服务发现、RPC 调用。

## 架构

```
服务A(提供方)                     etcd                     服务B(消费方)
NewRegister(...) ──Grant/Put──▶  /grpc/{svc}/{addr}        NewClient("grpc://svc")
KeepAlive 续租  ────────────────▶  (租约心跳)                │
                                 ◀── Get(/grpc/{svc}/) ────  Build → UpdateState
                                 ◀── Watch 监听变更 ────────  watcher 实时同步实例
服务A 下线/崩溃 → 租约过期/Revoke → DELETE → 消费方自动摘除该实例
```

- 注册 key 格式：`/grpc/{serviceName}/{serverAddr}`。value 默认是裸地址；带权重/元数据时为 JSON `{"addr": ..., "weight": ..., "metadata": {...}}`（两种格式发现端都兼容）。
- 租约（lease）即心跳：`NewRegister` **自动续租**，续租失败或主动 `Close()` 都会撤销租约，etcd 立即删除注册，消费方 watch 到 DELETE 后自动摘除，实现故障转移。
- 服务发现实现为 gRPC 自定义 resolver（scheme `grpc`），客户端用 `grpc://serviceName` 拨号即可透明完成寻址与故障转移。

## 快速开始

### 1. 服务端 + 注册

```go
import "github.com/laixhe/gonet/rpc"

// 启动 gRPC 服务（certFile/keyFile 同时非空则启用 TLS；自动注册 grpc_health_v1 健康检查）
srv, err := rpc.NewServer("0.0.0.0:50051", "", "")
if err != nil {
    log.Fatal(err)
}
srv.RegisterService(&pb.Greeter_ServiceDesc, &greeterServer{})
go func() {
    if err := srv.Start(); err != nil {
        log.Fatal(err)
    }
}()

// 注册到 etcd（leaseTtl 秒租约；自动续租，无需手动启动续租监听）
reg, err := rpc.NewRegister([]string{"http://127.0.0.1:2379"}, "greeter", "127.0.0.1:50051", 10)
if err != nil {
    log.Fatal(err)
}
defer reg.Close() // 优雅下线：撤销租约，消费方立即摘除
```

### 2. 客户端（服务发现）

```go
// 服务发现（进程内首次创建时注册 grpc:// scheme，只需创建一次）
disc, err := rpc.NewDiscovery([]string{"http://127.0.0.1:2379"}, 5*time.Second)
if err != nil {
    log.Fatal(err)
}
defer disc.Close()

// 用 grpc://serviceName 拨号，自动发现 + 故障转移
c, err := rpc.NewClient("grpc://greeter")
if err != nil {
    log.Fatal(err)
}
defer c.Close()

// 生成的 stub 通过 c.Conn() 使用
client := pb.NewGreeterClient(c.Conn())
resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "world"})
```

### 3. 优雅退出

```go
// 收到信号后：先注销（撤租约）→ 停服务 → 关客户端
sig := make(chan os.Signal, 1)
signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
<-sig

if err := reg.Close(); err != nil {   // 1. 撤销 etcd 租约，消费方立即摘除
    log.Printf("deregister: %v", err)
}
_ = srv.Stop()                        // 2. GracefulStop，等存量请求处理完
_ = c.Close()                         // 3. 关闭客户端
```

### 4. TLS

服务端与客户端对称使用：

```go
// 服务端
srv, _ := rpc.NewServer("0.0.0.0:50051", "server.crt", "server.key")

// 客户端（serverName 为服务端证书的 SAN/CN）
c, _ := rpc.NewClientTLS("grpc://greeter", "ca.crt", "greeter.example.com")
```

### 5. 负载均衡

默认 `round_robin`，多实例自动分流；可用选项覆盖：

```go
c, _ := rpc.NewClient("grpc://greeter", rpc.WithLoadBalancingPolicy("pick_first"))
```

### 6. 注册元数据（权重 / 附加信息）

注册时可携带权重与自定义元数据（写入 etcd 的 JSON value，供权重感知的负载均衡、灰度发布等使用）：

```go
reg, err := rpc.NewRegister(endpoints, "greeter", "127.0.0.1:50051", 10,
    rpc.WithWeight(10),
    rpc.WithMetadata(map[string]string{"zone": "cn-east-1", "version": "v2"}),
)
```

消费端通过 `resolver.Address.Attributes` 读取：

```go
// 自定义 resolver 或 balancer 中：
addr.Attributes.Value("weight")   // int
addr.Attributes.Value("metadata") // map[string]string
```

### 7. 健康检查

服务端自动注册 `grpc_health_v1`，默认整体状态 `SERVING`；实例不可用时置为 `NOT_SERVING`，供 LB / 注册中心探活：

```go
// 服务不可用（如依赖的后端挂了）
srv.SetHealth("", healthpb.HealthCheckResponse_NOT_SERVING)
// 恢复
srv.SetHealth("", healthpb.HealthCheckResponse_SERVING)
```

客户端也可通过 `/grpc.health.v1.Health/Check` 主动探活。

### 8. 拦截器与可观测性

内置 panic 恢复与访问日志拦截器：

```go
srv, err := rpc.NewServer(addr, "", "", rpc.WithServerOption(
    grpc.ChainUnaryInterceptor(
        rpc.RecoveryUnaryServerInterceptor(), // panic → codes.Internal，进程不崩
        rpc.LoggingUnaryServerInterceptor(),  // 打印 method / code / 耗时
    ),
))
```

Prometheus / OpenTelemetry 等中间件不内置（避免膨胀库依赖），通过选项透传接入，例如：

```go
import (
    "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/..."
    grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
)

srv, _ := rpc.NewServer(addr, "", "", rpc.WithServerOption(
    grpc.ChainUnaryInterceptor(grpcprom.NewServerMetrics().UnaryServerInterceptor()),
))
```

其它 gRPC 选项（自定义 codec、流式拦截器、限流等）均可通过 `rpc.WithServerOption` / `rpc.WithClientOption` 透传。

## 注意事项

- **服务发现单例**：`grpc://` scheme 在进程内只注册一次，**首次创建的 `Discovery`** 负责所有 `grpc://` 目标的解析，请使用同一个 etcd 集群地址创建。
- **多实例隔离**：每个 `grpc://serviceName` 解析器状态独立，互不干扰。
- **自动续租**：`NewRegister` 已自动启动租约续租，`Close()` 自动停止；`ListenLease` 保留仅为兼容旧调用（no-op）。
- 目标地址支持两种写法：`grpc://svc`（serviceName 在 Host）与 `grpc:///svc`（规范写法，在 Path）。
- `Server.Addr()` 可获取实际监听地址（`127.0.0.1:0` 随机端口场景）。
- 测试基于 embedded etcd，无需外部依赖即可 `go test ./...`。

## 测试

集成测试在独立测试模块 `rpc/test`（embedded etcd，无需外部依赖）：

```bash
cd rpc/test
go test ./... -race
```

覆盖：注册→发现→下线→上线的闭环、自动续租、真实 gRPC 调用、故障转移、幂等 Close、权重/元数据注册、panic 恢复、健康检查。

> 测试依赖（`go.etcd.io/etcd/server/v3` 及其依赖树）只存在于 `rpc/test/go.mod`，不污染库模块 `rpc/go.mod`。
