// Package tcp 提供基于 TCP 的长连接服务器与客户端实现, 包含连接管理、消息路由、心跳检测、断线重连与 TLS 支持。
package tcp

import (
	"crypto/tls"
	"time"
)

// Config 服务器配置
type Config struct {
	// 最大连接数
	MaxConnections int64
	// 连接分片数, 用于降低锁竞争
	Partitions int
	// TLS 配置, 非空时启用 TLS 监听
	TLS *tls.Config
	// 每连接消息处理 worker 数, 大于 1 时同连接消息可并发处理(可能乱序), 0 使用默认
	ProcessWorkers int
	// 心跳检测间隔, 0 使用默认
	HeartbeatInterval time.Duration
	// 心跳超时时间, 超过该时间未收到任何消息则断开连接, 0 使用默认
	HeartbeatTimeout time.Duration
	// 写超时时间, 对端不读导致写阻塞超过该时间则断开连接, 0 不超时(默认)
	WriteTimeout time.Duration
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		MaxConnections:    1000,
		Partitions:        100,
		ProcessWorkers:    1,
		HeartbeatInterval: DefaultHeartbeatInterval,
		HeartbeatTimeout:  DefaultHeartbeatTimeout,
	}
}

// Check 校验并补全默认值
func (c *Config) Check() {
	if c.MaxConnections <= 0 {
		c.MaxConnections = 1000
	}
	if c.Partitions <= 0 {
		c.Partitions = 100
	}
	if c.ProcessWorkers <= 0 {
		c.ProcessWorkers = 1
	}
	if c.HeartbeatInterval <= 0 {
		c.HeartbeatInterval = DefaultHeartbeatInterval
	}
	if c.HeartbeatTimeout <= 0 {
		c.HeartbeatTimeout = DefaultHeartbeatTimeout
	}
}
