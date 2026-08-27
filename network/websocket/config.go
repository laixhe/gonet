// Package websocket 提供基于 WebSocket 的服务器与客户端实现, 与 tcp 共用同一套接口体系
// (服务器 IServer, 客户端 IClient), 支持消息路由、心跳检测、连接事件钩子与连接管理,
// 方便 H5/Web 客户端接入同一后端。
package websocket

import (
	"crypto/tls"
	"net/http"
	"time"
)

var (
	// DefaultHeartbeatInterval 默认心跳发送/检测间隔
	DefaultHeartbeatInterval = 15 * time.Second
	// DefaultHeartbeatTimeout 默认心跳超时时间, 超过该时间未收到任何消息(pong/业务消息)则断开连接
	DefaultHeartbeatTimeout = 30 * time.Second
)

// Config 服务器配置
type Config struct {
	// 最大连接数
	MaxConnections int64
	// 升级路径, 为空默认 "/ws"
	Path string
	// 心跳检测间隔, 0 使用默认
	HeartbeatInterval time.Duration
	// 心跳超时时间, 超过该时间未收到任何消息则断开连接, 0 使用默认
	HeartbeatTimeout time.Duration
	// 写超时时间, 对端不读导致写阻塞超过该时间则断开连接, 0 不超时(默认)
	WriteTimeout time.Duration
	// TLS 配置, 非空时以 wss 启动(配置需含证书)
	TLS *tls.Config
	// 跨域校验, 为空默认允许所有来源(游戏/IM 服务器通常需接受 H5 跨域连接)
	CheckOrigin func(r *http.Request) bool
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		MaxConnections:    1000,
		Path:              "/ws",
		HeartbeatInterval: DefaultHeartbeatInterval,
		HeartbeatTimeout:  DefaultHeartbeatTimeout,
	}
}

// Check 校验并补全默认值
func (c *Config) Check() {
	if c.MaxConnections <= 0 {
		c.MaxConnections = 1000
	}
	if c.Path == "" {
		c.Path = "/ws"
	}
	if c.HeartbeatInterval <= 0 {
		c.HeartbeatInterval = DefaultHeartbeatInterval
	}
	if c.HeartbeatTimeout <= 0 {
		c.HeartbeatTimeout = DefaultHeartbeatTimeout
	}
}
