# network 示例程序

演示 gonet network 包的 TCP / UDP / WebSocket 网络功能，包含完整的服务端、客户端与对应单元测试。

## 前置要求

- Go 1.26+
- xlog 通过 go.mod 的 replace 指向本地模块
- WebSocket 示例依赖 `github.com/gorilla/websocket`（已声明在 go.mod）
- KCP 示例依赖 `github.com/xtaci/kcp-go/v5`（已声明在 go.mod）

## 目录结构

```
example/
├── tcp/
│   ├── main.go      # TCP 示例(自动演示 + 交互模式)
│   └── main_test.go # TCP 示例单元测试
├── udp/
│   ├── main.go      # UDP 示例(回声服务器)
│   └── main_test.go # UDP 示例单元测试
├── websocket/
│   └── main.go      # WebSocket 示例(回声服务器)
├── kcp/
│   └── main.go      # KCP 示例(回声服务器, FEC+快速模式)
└── README.md
```

## TCP 示例

演示：消息路由与广播、服务端单发、心跳保活与超时踢线、未注册消息的默认处理。

```bash
cd network

# 自动演示(运行约 10 秒, 固定端口 127.0.0.1:9999)
go run ./example/tcp
```

自动演示包含 3 个客户端：

| 客户端 | 行为 | 演示点 |
|---|---|---|
| alice | 登录后发送聊天消息 | 消息路由 + 服务端单发 + 广播 |
| bob | 登录后接收聊天广播 | 广播接收 |
| silent | 建立连接后不发任何消息 | 服务端心跳检测, 约 5 秒后被断开 |

```bash
# 交互模式: 从标准输入发送聊天消息, 输入 quit 或 exit 退出
go run ./example/tcp -interactive
```

交互模式中启动一个服务端和 `me` 客户端，输入内容将以聊天消息广播给自己。

## UDP 示例

演示：UDP 无连接收发、消息路由、定向回复（`Reply`）。服务端通过 `:0` 自动分配端口。

```bash
cd network
go run ./example/udp
```

运行后客户端发送 3 条消息，服务端回声回复 `echo:<消息>`。

## WebSocket 示例

演示：WebSocket 升级握手、消息路由、回声回复，与 tcp/udp 共用同一套 `network.IServer` 接口（路由/心跳/钩子）。

```bash
cd network
go run ./example/websocket
```

运行后 gorilla 客户端连接 `ws://127.0.0.1:<port>/ws` 发送 3 条消息，服务端回声回复。H5 浏览器可用原生 `WebSocket` 连接同一地址，发送 packet 二进制协议消息。

## KCP 示例

演示：KCP(UDP 可靠传输) 服务器、消息路由、回声回复，默认启用 FEC(10+3) 前向纠错与快速模式，与 tcp/websocket 共用同一套 `network.IServer` 接口。

```bash
cd network
go run ./example/kcp
```

运行后 kcp 客户端连接 `kcp://127.0.0.1:<port>` 发送 3 条消息，服务端回声回复。注意 kcp-go 为懒握手，客户端发送第一条消息后服务端才建立会话。

## 运行单元测试

示例测试随库一起纳入全量测试，也可单独运行：

```bash
cd network

# 单独运行示例测试(含数据竞争检测)
go test -race ./example/tcp ./example/udp

# 或直接跑全模块测试(自动包含示例与全部协议实现)
go test -race ./...
```

测试覆盖：会话表增删与广播、TCP 登录欢迎与聊天广播端到端、UDP 回声往返、WebSocket 心跳/路由、KCP 弱网传输。

## 消息 ID 约定

| ID | 含义 | 方向 |
|---|---|---|
| 1 | 框架心跳消息 | 客户端 → 服务端 |
| 100 | 登录 | 客户端 → 服务端 |
| 101 | 聊天 | 客户端 → 服务端 |
| 200 | 服务端单发 | 服务端 → 客户端 |
| 201 | 服务端广播 | 服务端 → 客户端 |
| 999 | 未注册路由的消息(演示默认处理) | 客户端 → 服务端 |
