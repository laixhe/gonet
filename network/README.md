# network 网络模块

提供 TCP / KCP / WebSocket / UDP 四种网络协议的服务器与客户端实现，统一二进制消息协议、连接管理、消息路由与心跳检测。前三种协议共享同一套 `IServer`/`IClient` 接口体系（UDP 无连接，为独立 API）。

## 包结构

| 包 | 说明 |
|---|---|
| `tcp` | TCP 长连接服务器/客户端：分片锁连接管理、worker 池、心跳、断线重连、TLS |
| `kcp` | KCP(UDP 可靠传输) 服务器/客户端：FEC 前向纠错、AES 加密、快速模式，弱网场景比 TCP 快 |
| `websocket` | WebSocket 服务器/客户端：协议级 ping/pong 心跳，H5/Web 客户端接入 |
| `udp` | UDP 无连接服务器/客户端：按来源地址识别对端，支持每地址限流 |
| `packet` | 二进制消息协议 `[ID:4字节][DataLen:4字节][Data]`，含大包防护 |
| `http/client` | 链式 API 的 HTTP 客户端 |
| `header` | HTTP 头、MIME 类型、平台/语言常量 |
| `example` | 各协议的可运行示例与端到端测试 |

根包提供接口定义与公共组件：

- 接口：`IServer` / `IClient` / `IConn` / `IManager`
- 公共组件：`Router` 消息路由、`BaseServer` 服务器基座（路由 + 连接事件 + 广播）、`StreamConnection` 通用流式连接（tcp/kcp 共用）、`MapManager` 通用连接管理器（kcp/websocket 共用）
- 日志：`Logger` 接口 + `SetLogger`（与 xlog 兼容）
- 错误：`ErrConnectionClosed`、`ErrTooManyConnection`、`ErrConnectionHanged` 等

## 协议选型

| 场景 | 推荐 | 说明 |
|---|---|---|
| 通用长连接、实时通信 | `tcp` | 可靠、简单、生态好 |
| 弱网（高丢包/高延迟）、游戏/实时对战 | `kcp` | 比 TCP 更快，需业务容忍少量重传开销 |
| H5/Web 前端直连 | `websocket` | 浏览器原生支持，与 tcp 共用接口体系 |
| 高吞吐、可容忍丢包（日志、指标、广播） | `udp` | 无连接，消息边界天然保留 |
| 服务端间 HTTP 调用 | `http/client` | 链式 API + 连接池 |

## 快速上手（TCP 示例）

```go
// 服务器
s := tcp.NewServer()
s.Router(100, func(conn network.IConn, data []byte) {
    _ = conn.Send(200, []byte("reply:"+string(data)))
})
s.OnConnect(func(conn network.IConn) { conn.BindUid(42) })
go s.Start("127.0.0.1:8888")
defer s.Stop()

// 客户端
c := tcp.NewClient()
c.SetHandler(func(id uint32, data []byte) { /* 处理消息 */ })
_ = c.Start("127.0.0.1:8888")
_ = c.Send(100, []byte("hello"))
```

其他协议用法一致：把 `tcp.NewServer()` 换成 `kcp.NewServer()` / `websocket.NewServer()` 即可，客户端同理（UDP 除外，见 `example/udp`）。

## 广播与连接管理

服务器接口内置广播能力，可向所有连接或排除指定 Uid 广播：

```go
s.Broadcast(300, []byte("全体公告"))
s.BroadcastExclude(301, []byte("有人说话"), 42) // 排除 Uid=42

m := s.GetManager()
m.Count()                  // 当前连接数
m.FindByUid(42)            // 按 Uid 查找连接
m.KickByUid(42)            // 顶号踢线下线
m.ForEach(func(c network.IConn) { /* 遍历 */ })
```

## 配置要点

- **心跳**：`HeartbeatInterval`/`HeartbeatTimeout`，超过超时时间未收到任何消息断开连接；tcp/kcp 客户端自动发业务心跳，websocket 用协议级 ping/pong。
- **worker 池**：`ProcessWorkers` 控制每连接消息处理协程数，大于 1 时同连接消息可并发处理（可能乱序）。
- **写超时**：`WriteTimeout` 防止对端不读导致写阻塞，0 不超时。
- **平滑关闭**：`Stop()` 会等待已入队消息处理完再关闭连接。
- **KCP**：`DataShards`/`ParityShards` 配 FEC、`Key` 配加密（客户端须与服务端一致）；kcp-go 为懒握手，客户端连接后自动发心跳触发服务端建立会话。

## 测试

```bash
go test -count=1 -race -v ./...
```

四个协议示例均带端到端测试：`example/tcp`、`example/udp`、`example/kcp`、`example/websocket`。
