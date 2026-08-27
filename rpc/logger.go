package rpc

import (
	"log"
	"sync/atomic"
)

// logger 是 rpc 模块的内部日志器，默认标准库 log，可用 SetLogger 替换。
// 用 atomic 保证并发读安全（watcher/续租 goroutine 与用户线程可能同时访问）。
var logger atomic.Pointer[log.Logger]

func init() {
	logger.Store(log.Default())
}

// SetLogger 替换 rpc 模块内部日志器（默认标准库 log）。
// 可接入项目自己的日志体系（如 slog/zap 的 writer）；并发安全，可随时调用。
func SetLogger(l *log.Logger) {
	if l == nil {
		l = log.Default()
	}
	logger.Store(l)
}

// logf 经内部日志器输出
func logf(format string, args ...any) {
	logger.Load().Printf(format, args...)
}
