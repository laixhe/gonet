package xlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// atomicSlogLevel 基于 atomic.Int32 实现 slog.Leveler 接口，无锁并发读写，适合日志热路径场景。
type atomicSlogLevel struct {
	level atomic.Int32
}

func (a *atomicSlogLevel) Level() slog.Level {
	return slog.Level(a.level.Load())
}

func (a *atomicSlogLevel) Set(level slog.Level) {
	a.level.Store(int32(level))
}

// slogLevelFromString 将配置中的日志级别字符串转换为 slog.Level。
func slogLevelFromString(level string) slog.Level {
	switch level {
	case LevelTypeDebug:
		return slog.LevelDebug
	case LevelTypeInfo:
		return slog.LevelInfo
	case LevelTypeWarn:
		return slog.LevelWarn
	case LevelTypeError:
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}

// SlogContextHandler 包装 slog.Handler，在处理日志时自动从 context 中提取指定 key 的值并附加到日志属性中。
type SlogContextHandler struct {
	slog.Handler
	ContextKey string
}

func (ch *SlogContextHandler) Handle(ctx context.Context, record slog.Record) error {
	if ctx != nil {
		if v, ok := ctx.Value(ch.ContextKey).(string); ok && v != "" {
			record = record.Clone()
			record.AddAttrs(slog.String(ch.ContextKey, v))
		}
	}
	return ch.Handler.Handle(ctx, record)
}

// SClient 基于标准库 log/slog 的日志客户端，支持 JSON/Console 格式切换、按大小/时间分割文件、
// 运行时动态切换日志级别和输出目标。
type SClient struct {
	config  *Config
	handler atomic.Value // slog.Handler
	logger  atomic.Value // *slog.Logger

	mu          sync.RWMutex
	writer      io.Writer
	atomicLevel *atomicSlogLevel
}

// loadLogger 加载当前 *slog.Logger 指针，避免每次调用重复类型断言。
func (lc *SClient) loadLogger() *slog.Logger {
	return lc.logger.Load().(*slog.Logger)
}

// loadHandler 加载当前 slog.Handler，避免每次调用重复类型断言。
func (lc *SClient) loadHandler() slog.Handler {
	return lc.handler.Load().(slog.Handler)
}

// Writer 返回当前日志写入目标。
func (lc *SClient) Writer() io.Writer {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.writer
}

// Handler 返回底层 slog.Handler。
func (lc *SClient) Handler() slog.Handler {
	return lc.loadHandler()
}

// SLog 返回底层 *slog.Logger，供需要直接使用 slog API 的场景。
func (lc *SClient) SLog() *slog.Logger {
	return lc.loadLogger()
}

// Sync 刷新日志缓冲区，确保所有缓冲的日志被写入。
func (lc *SClient) Sync() error {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	if w, ok := lc.writer.(interface{ Sync() error }); ok {
		return w.Sync()
	}
	return nil
}

// InitSlog 创建并初始化 SClient。
// 如果不传 config 则使用默认配置（console 模式、debug 级别）。
func InitSlog(configs ...*Config) (*SClient, error) {
	var config *Config
	if len(configs) == 0 {
		config = &Config{}
	} else {
		config = configs[0]
	}
	if err := config.Check(); err != nil {
		return nil, err
	}
	lc := &SClient{
		config:      config,
		atomicLevel: &atomicSlogLevel{},
	}
	lc.atomicLevel.Set(slogLevelFromString(lc.config.Level))

	if lc.config.Run == RunTypeFile {
		lc.writer = &lumberjack.Logger{
			Filename:   lc.config.Path,
			MaxSize:    lc.config.MaxSize,
			MaxBackups: lc.config.MaxBackups,
			MaxAge:     lc.config.MaxAge,
			LocalTime:  lc.config.LocalTime,
			Compress:   lc.config.Compress,
		}
	} else {
		lc.writer = os.Stdout
	}
	lc.rebuildHandler()
	return lc, nil
}

// rebuildHandler 根据当前 writer 和 atomicLevel 重建 handler 与 logger。
func (lc *SClient) rebuildHandler() {
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     lc.atomicLevel,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				attr.Value = slog.StringValue(attr.Value.Time().Format(time.DateTime))
			}
			if attr.Key == slog.SourceKey {
				if source, ok := attr.Value.Any().(*slog.Source); ok {
					return slog.String("source", shortenSource(source))
				}
			}
			return attr
		},
	}
	var baseHandler slog.Handler
	if lc.config.Format == FormatConsole {
		baseHandler = slog.NewTextHandler(lc.writer, opts)
	} else {
		baseHandler = slog.NewJSONHandler(lc.writer, opts)
	}
	h := &SlogContextHandler{
		Handler:    baseHandler,
		ContextKey: lc.config.ContextKey,
	}
	lc.handler.Store(h)
	lc.logger.Store(slog.New(h))
}

// shortenSource 截取源文件路径，只保留最后 3 级目录 + 文件名。
func shortenSource(source *slog.Source) string {
	shortFile := source.File
	count := 0
	for i := len(source.File) - 1; i > 0; i-- {
		if os.IsPathSeparator(source.File[i]) {
			shortFile = source.File[i+1:]
			count++
			if count >= 3 {
				break
			}
		}
	}
	return fmt.Sprintf("%s:%d", shortFile, source.Line)
}

// SetLevel 运行时修改日志级别（立即生效）。
func (lc *SClient) SetLevel(level string) {
	lc.mu.Lock()
	lc.config.Level = level
	lc.mu.Unlock()
	lc.atomicLevel.Set(slogLevelFromString(level))
}

// SetOutput 运行时更换输出目标（立即生效）。
func (lc *SClient) SetOutput(w io.Writer) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.writer = w
	lc.rebuildHandler()
}

// SetFormat 运行时切换日志输出格式（JSON/Console，立即生效）。
// 非法值自动回退为 JSON。
func (lc *SClient) SetFormat(format string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	if format != FormatConsole && format != FormatJSON {
		format = FormatJSON
	}
	lc.config.Format = format
	lc.rebuildHandler()
}

// SetContextKey 运行时更换 SlogContextHandler 提取 context 的 key（立即生效）。
// 用于切换分布式追踪中的 trace key，例如从 "traceId" 改为 "x-trace-id"。
// 仅对 SClient 生效，ZClient 无此概念。
func (lc *SClient) SetContextKey(key string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.config.ContextKey = key
	lc.rebuildHandler()
}

// Level 返回当前日志级别字符串。
func (lc *SClient) Level() string {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.config.Level
}

// ============================================================
// 非格式化方法
// ============================================================

// Debug 输出 Debug 级别日志。
func (lc *SClient) Debug(args ...any) {
	if !lc.loadHandler().Enabled(backgroundCtx, slog.LevelDebug) {
		return
	}
	lc.loadLogger().Debug(fmt.Sprint(args...))
}

// Info 输出 Info 级别日志。
func (lc *SClient) Info(args ...any) {
	if !lc.loadHandler().Enabled(backgroundCtx, slog.LevelInfo) {
		return
	}
	lc.loadLogger().Info(fmt.Sprint(args...))
}

// Warn 输出 Warn 级别日志。
func (lc *SClient) Warn(args ...any) {
	if !lc.loadHandler().Enabled(backgroundCtx, slog.LevelWarn) {
		return
	}
	lc.loadLogger().Warn(fmt.Sprint(args...))
}

// Error 输出 Error 级别日志。
func (lc *SClient) Error(args ...any) {
	if !lc.loadHandler().Enabled(backgroundCtx, slog.LevelError) {
		return
	}
	lc.loadLogger().Error(fmt.Sprint(args...))
}

// ============================================================
// 格式化方法
// ============================================================

// Debugf 格式化输出 Debug 级别日志。
func (lc *SClient) Debugf(template string, args ...any) {
	if !lc.loadHandler().Enabled(backgroundCtx, slog.LevelDebug) {
		return
	}
	lc.loadLogger().Debug(fmt.Sprintf(template, args...))
}

// Infof 格式化输出 Info 级别日志。
func (lc *SClient) Infof(template string, args ...any) {
	if !lc.loadHandler().Enabled(backgroundCtx, slog.LevelInfo) {
		return
	}
	lc.loadLogger().Info(fmt.Sprintf(template, args...))
}

// Warnf 格式化输出 Warn 级别日志。
func (lc *SClient) Warnf(template string, args ...any) {
	if !lc.loadHandler().Enabled(backgroundCtx, slog.LevelWarn) {
		return
	}
	lc.loadLogger().Warn(fmt.Sprintf(template, args...))
}

// Errorf 格式化输出 Error 级别日志。
func (lc *SClient) Errorf(template string, args ...any) {
	if !lc.loadHandler().Enabled(backgroundCtx, slog.LevelError) {
		return
	}
	lc.loadLogger().Error(fmt.Sprintf(template, args...))
}

// Printf 格式化输出 Info 级别日志（等价于 Infof）。
func (lc *SClient) Printf(template string, args ...any) {
	if !lc.loadHandler().Enabled(backgroundCtx, slog.LevelInfo) {
		return
	}
	lc.loadLogger().Info(fmt.Sprintf(template, args...))
}

// ============================================================
// 键值对方法
// ============================================================

// Debugw 以键值对形式输出 Debug 级别日志。
func (lc *SClient) Debugw(msg string, keysAndValues ...any) {
	lc.loadLogger().Debug(msg, keysAndValues...)
}

// Infow 以键值对形式输出 Info 级别日志。
func (lc *SClient) Infow(msg string, keysAndValues ...any) {
	lc.loadLogger().Info(msg, keysAndValues...)
}

// Warnw 以键值对形式输出 Warn 级别日志。
func (lc *SClient) Warnw(msg string, keysAndValues ...any) {
	lc.loadLogger().Warn(msg, keysAndValues...)
}

// Errorw 以键值对形式输出 Error 级别日志。
func (lc *SClient) Errorw(msg string, keysAndValues ...any) {
	lc.loadLogger().Error(msg, keysAndValues...)
}

// ============================================================
// Context 方法
// ============================================================

// DebugContext 带 context 输出 Debug 级别日志。
func (lc *SClient) DebugContext(ctx context.Context, args ...any) {
	if !lc.loadHandler().Enabled(ctx, slog.LevelDebug) {
		return
	}
	lc.loadLogger().DebugContext(ctx, fmt.Sprint(args...))
}

// InfoContext 带 context 输出 Info 级别日志。
func (lc *SClient) InfoContext(ctx context.Context, args ...any) {
	if !lc.loadHandler().Enabled(ctx, slog.LevelInfo) {
		return
	}
	lc.loadLogger().InfoContext(ctx, fmt.Sprint(args...))
}

// WarnContext 带 context 输出 Warn 级别日志。
func (lc *SClient) WarnContext(ctx context.Context, args ...any) {
	if !lc.loadHandler().Enabled(ctx, slog.LevelWarn) {
		return
	}
	lc.loadLogger().WarnContext(ctx, fmt.Sprint(args...))
}

// ErrorContext 带 context 输出 Error 级别日志。
func (lc *SClient) ErrorContext(ctx context.Context, args ...any) {
	if !lc.loadHandler().Enabled(ctx, slog.LevelError) {
		return
	}
	lc.loadLogger().ErrorContext(ctx, fmt.Sprint(args...))
}

// DebugContextf 带 context 格式化输出 Debug 级别日志。
func (lc *SClient) DebugContextf(ctx context.Context, format string, args ...any) {
	if !lc.loadHandler().Enabled(ctx, slog.LevelDebug) {
		return
	}
	lc.loadLogger().DebugContext(ctx, fmt.Sprintf(format, args...))
}

// InfoContextf 带 context 格式化输出 Info 级别日志。
func (lc *SClient) InfoContextf(ctx context.Context, format string, args ...any) {
	if !lc.loadHandler().Enabled(ctx, slog.LevelInfo) {
		return
	}
	lc.loadLogger().InfoContext(ctx, fmt.Sprintf(format, args...))
}

// WarnContextf 带 context 格式化输出 Warn 级别日志。
func (lc *SClient) WarnContextf(ctx context.Context, format string, args ...any) {
	if !lc.loadHandler().Enabled(ctx, slog.LevelWarn) {
		return
	}
	lc.loadLogger().WarnContext(ctx, fmt.Sprintf(format, args...))
}

// ErrorContextf 带 context 格式化输出 Error 级别日志。
func (lc *SClient) ErrorContextf(ctx context.Context, format string, args ...any) {
	if !lc.loadHandler().Enabled(ctx, slog.LevelError) {
		return
	}
	lc.loadLogger().ErrorContext(ctx, fmt.Sprintf(format, args...))
}

// DebugContextw 带 context 以键值对形式输出 Debug 级别日志。
func (lc *SClient) DebugContextw(ctx context.Context, msg string, keysAndValues ...any) {
	lc.loadLogger().DebugContext(ctx, msg, keysAndValues...)
}

// InfoContextw 带 context 以键值对形式输出 Info 级别日志。
func (lc *SClient) InfoContextw(ctx context.Context, msg string, keysAndValues ...any) {
	lc.loadLogger().InfoContext(ctx, msg, keysAndValues...)
}

// WarnContextw 带 context 以键值对形式输出 Warn 级别日志。
func (lc *SClient) WarnContextw(ctx context.Context, msg string, keysAndValues ...any) {
	lc.loadLogger().WarnContext(ctx, msg, keysAndValues...)
}

// ErrorContextw 带 context 以键值对形式输出 Error 级别日志。
func (lc *SClient) ErrorContextw(ctx context.Context, msg string, keysAndValues ...any) {
	lc.loadLogger().ErrorContext(ctx, msg, keysAndValues...)
}

// ============================================================
// Fatal / Panic 方法
// ============================================================

// Fatal 输出 Error 级别日志后调用 os.Exit(1) 终止程序。
func (lc *SClient) Fatal(args ...any) {
	lc.loadLogger().Error(fmt.Sprint(args...))
	os.Exit(1)
}

// Fatalf 格式化输出 Error 级别日志后调用 os.Exit(1) 终止程序。
func (lc *SClient) Fatalf(format string, args ...any) {
	lc.loadLogger().Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// Fatalw 以键值对形式输出 Error 级别日志后调用 os.Exit(1) 终止程序。
func (lc *SClient) Fatalw(msg string, keysAndValues ...any) {
	lc.loadLogger().Error(msg, keysAndValues...)
	os.Exit(1)
}

// Panic 输出 Error 级别日志后触发 panic。
func (lc *SClient) Panic(args ...any) {
	s := fmt.Sprint(args...)
	lc.loadLogger().Error(s)
	panic(s)
}

// Panicf 格式化输出 Error 级别日志后触发 panic。
func (lc *SClient) Panicf(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	lc.loadLogger().Error(s)
	panic(s)
}

// Panicw 以键值对形式输出 Error 级别日志后触发 panic。
func (lc *SClient) Panicw(msg string, keysAndValues ...any) {
	lc.loadLogger().Error(msg, keysAndValues...)
	panic(msg)
}

// ============================================================
// Trace 方法
// ============================================================

// Trace 等价于 Debug，适用于需要 trace 级别语义的场景。
func (lc *SClient) Trace(args ...any) {
	if !lc.loadHandler().Enabled(backgroundCtx, slog.LevelDebug) {
		return
	}
	lc.loadLogger().Debug(fmt.Sprint(args...))
}

// Tracef 等价于 Debugf，格式化输出 trace 日志。
func (lc *SClient) Tracef(format string, args ...any) {
	if !lc.loadHandler().Enabled(backgroundCtx, slog.LevelDebug) {
		return
	}
	lc.loadLogger().Debug(fmt.Sprintf(format, args...))
}

// Tracew 等价于 Debugw，以键值对形式输出 trace 日志。
func (lc *SClient) Tracew(msg string, keysAndValues ...any) {
	lc.loadLogger().Debug(msg, keysAndValues...)
}

// 确保 SClient 实现了 Logger 接口
var _ Logger = (*SClient)(nil)
