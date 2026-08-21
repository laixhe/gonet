// Package network 提供 TCP/KCP/WebSocket/UDP 网络服务器与客户端的抽象接口与实现,
// 包含二进制消息编解码、连接管理与消息路由。
//
// 接口层定义 IServer/IClient/IConn/IManager, tcp/kcp/websocket 子包实现该接口体系
// (UDP 无连接, 为独立 API); 消息协议位于 packet 子包; HTTP 客户端位于 http/client 子包;
// Router/BaseServer/StreamConnection/MapManager 为 tcp/kcp/websocket 共用的公共组件。
package network

import (
	"log"
	"sync"
)

// Logger 日志接口, 方法签名与 xlog.Logger 兼容, 可通过 SetLogger 注入 xlog 实例
type Logger interface {
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
	Sync() error
}

// stdLogger 默认日志器, 输出到标准日志
type stdLogger struct{}

func (stdLogger) Debug(args ...any)            { log.Print(args...) }
func (stdLogger) Info(args ...any)             { log.Print(args...) }
func (stdLogger) Warn(args ...any)             { log.Print(args...) }
func (stdLogger) Error(args ...any)            { log.Print(args...) }
func (stdLogger) Debugf(f string, args ...any) { log.Printf(f, args...) }
func (stdLogger) Infof(f string, args ...any)  { log.Printf(f, args...) }
func (stdLogger) Warnf(f string, args ...any)  { log.Printf(f, args...) }
func (stdLogger) Errorf(f string, args ...any) { log.Printf(f, args...) }
func (stdLogger) Debugw(msg string, kv ...any) { log.Print(msg, kv) }
func (stdLogger) Infow(msg string, kv ...any)  { log.Print(msg, kv) }
func (stdLogger) Warnw(msg string, kv ...any)  { log.Print(msg, kv) }
func (stdLogger) Errorw(msg string, kv ...any) { log.Print(msg, kv) }
func (stdLogger) Sync() error                  { return nil }

// logger 包级日志器, 读写锁保护支持运行时替换
var (
	loggerMu sync.RWMutex
	logger   Logger = &stdLogger{}
)

// Log 获取当前日志器
func Log() Logger {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	return logger
}

// SetLogger 设置日志器, 需在服务器/客户端启动前调用
func SetLogger(l Logger) {
	if l == nil {
		return
	}
	loggerMu.Lock()
	logger = l
	loggerMu.Unlock()
}
