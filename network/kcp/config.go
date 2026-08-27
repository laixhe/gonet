// Package kcp 提供基于 KCP(UDP 可靠传输) 的服务器与客户端实现,
// 与 tcp/websocket 共用同一套 IServer 接口体系, 支持消息路由、心跳检测与连接事件钩子。
// KCP 在弱网(高丢包/高延迟)环境下比 TCP 传输更快, 适合游戏/实时对战场景。
package kcp

import (
	"time"

	kcpv5 "github.com/xtaci/kcp-go/v5"
)

var (
	// DefaultHeartbeatInterval 默认心跳发送/检测间隔
	DefaultHeartbeatInterval = 15 * time.Second
	// DefaultHeartbeatTimeout 默认心跳超时时间, 超过该时间未收到任何消息则断开连接
	DefaultHeartbeatTimeout = 30 * time.Second
)

// Config 服务器配置
type Config struct {
	// 最大连接数
	MaxConnections int64
	// FEC 数据分片数, 0 禁用 FEC 前向纠错
	DataShards int
	// FEC 冗余分片数, 0 禁用 FEC 前向纠错
	ParityShards int
	// 加密块, nil 不加密; 可用 Key 便捷生成
	Block kcpv5.BlockCrypt
	// 加密密钥, 设置后自动生成 AES 加密块(Block 优先)
	Key []byte
	// MTU, 0 使用默认 1400
	Mtu int
	// 快速模式(NoDelay), 默认关闭
	NoDelay bool
	// 心跳检测间隔, 0 使用默认
	HeartbeatInterval time.Duration
	// 心跳超时时间, 超过该时间未收到任何消息则断开连接, 0 使用默认
	HeartbeatTimeout time.Duration
	// 写超时时间, 对端不读导致写阻塞超过该时间则断开连接, 0 不超时(默认)
	WriteTimeout time.Duration
	// 每连接消息处理 worker 数, 大于 1 时同连接消息可并发处理(可能乱序), 0 使用默认
	ProcessWorkers int
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		MaxConnections:    1000,
		DataShards:        10,
		ParityShards:      3,
		Mtu:               1400,
		NoDelay:           true,
		HeartbeatInterval: DefaultHeartbeatInterval,
		HeartbeatTimeout:  DefaultHeartbeatTimeout,
		ProcessWorkers:    1,
	}
}

// Check 校验并补全默认值
func (c *Config) Check() {
	if c.MaxConnections <= 0 {
		c.MaxConnections = 1000
	}
	if c.Mtu <= 0 {
		c.Mtu = 1400
	}
	if c.HeartbeatInterval <= 0 {
		c.HeartbeatInterval = DefaultHeartbeatInterval
	}
	if c.HeartbeatTimeout <= 0 {
		c.HeartbeatTimeout = DefaultHeartbeatTimeout
	}
	if c.ProcessWorkers <= 0 {
		c.ProcessWorkers = 1
	}
	// 用 Key 便捷生成 AES 加密块
	if c.Block == nil && len(c.Key) > 0 {
		if block, err := kcpv5.NewAESBlockCrypt(c.Key); err == nil {
			c.Block = block
		}
	}
}
