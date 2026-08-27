package xlog

import (
	"context"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// ZClient 基于 uber-go/zap 的日志客户端，支持 JSON/Console 格式切换、按大小/时间分割文件、
// 运行时动态切换日志级别和输出目标。
type ZClient struct {
	config      *Config
	logger      atomic.Pointer[zap.Logger]
	sugarLogger atomic.Pointer[zap.SugaredLogger]

	mu            sync.RWMutex
	writer        zapcore.WriteSyncer
	atomicLevel   zap.AtomicLevel
	encoderConfig zapcore.EncoderConfig
	zapOpts       []zap.Option
}

// ZLog 返回底层 *zap.Logger，供需要直接使用 zap API 的场景。
func (c *ZClient) ZLog() *zap.Logger {
	return c.logger.Load()
}

// Logger 返回底层 *zap.Logger，用于传入 xfiber.New() 等需要 *zap.Logger 的场景。
func (c *ZClient) Logger() *zap.Logger {
	return c.logger.Load()
}

// SugaredLogger 返回底层 *zap.SugaredLogger，供需要直接使用 SugaredLogger API 的场景。
func (c *ZClient) SugaredLogger() *zap.SugaredLogger {
	return c.sugarLogger.Load()
}

// Level 返回当前日志级别字符串。
func (c *ZClient) Level() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.Level
}

// Sync 刷新日志缓冲区。
func (c *ZClient) Sync() error {
	return c.logger.Load().Sync()
}

// Writer 返回当前日志写入目标。
func (c *ZClient) Writer() io.Writer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.writer
}

// InitZap 创建并初始化 ZClient。
// 如果不传 config 则使用默认配置（console 模式、debug 级别）。
func InitZap(configs ...*Config) (*ZClient, error) {
	var config *Config
	if len(configs) == 0 {
		config = &Config{}
	} else {
		config = configs[0]
	}
	if err := config.Check(); err != nil {
		return nil, err
	}
	client := &ZClient{
		config: config,
	}
	if config.Run == RunTypeFile {
		client.writer = zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.Path,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			LocalTime:  config.LocalTime,
			Compress:   config.Compress,
		})
	} else {
		client.writer = zapcore.AddSync(os.Stdout)
	}
	zapInit(client)
	return client, nil
}

// zapInit 初始化 zap logger 的编码器和核心配置。
func zapInit(c *ZClient) {
	c.atomicLevel = zap.NewAtomicLevel()
	c.atomicLevel.SetLevel(zapLevelFromString(c.config.Level))

	c.encoderConfig = zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		CallerKey:      "source",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	c.zapOpts = []zap.Option{
		zap.Development(),
		zap.AddCaller(),
		zap.AddCallerSkip(c.config.CallerSkip + 1),
	}

	logger := zap.New(
		zapcore.NewCore(
			zapEncoder(c.config.Format, c.encoderConfig),
			c.writer,
			c.atomicLevel,
		),
		c.zapOpts...,
	)
	c.logger.Store(logger)
	c.sugarLogger.Store(logger.Sugar())
}

// zapEncoder 根据配置返回对应的 zap 编码器。
func zapEncoder(format string, cfg zapcore.EncoderConfig) zapcore.Encoder {
	if format == FormatConsole {
		return zapcore.NewConsoleEncoder(cfg)
	}
	return zapcore.NewJSONEncoder(cfg)
}

// zapLevelFromString 将配置中的日志级别字符串转换为 zapcore.Level。
func zapLevelFromString(level string) zapcore.Level {
	switch level {
	case LevelTypeDebug:
		return zap.DebugLevel
	case LevelTypeInfo:
		return zap.InfoLevel
	case LevelTypeWarn:
		return zap.WarnLevel
	case LevelTypeError:
		return zap.ErrorLevel
	default:
		return zap.DebugLevel
	}
}

// zapTimeEncoder 格式化时间为 "2006-01-02 15:04:05" 格式。
func zapTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.DateTime))
}

// SetLevel 运行时修改日志级别（立即生效）。
func (c *ZClient) SetLevel(level string) {
	c.mu.Lock()
	c.config.Level = level
	c.mu.Unlock()
	c.atomicLevel.SetLevel(zapLevelFromString(level))
}

// SetOutput 运行时更换输出目标（立即生效）。
func (c *ZClient) SetOutput(w io.Writer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writer = zapcore.AddSync(w)
	logger := zap.New(
		zapcore.NewCore(
			zapEncoder(c.config.Format, c.encoderConfig),
			c.writer,
			c.atomicLevel,
		),
		c.zapOpts...,
	)
	c.logger.Store(logger)
	c.sugarLogger.Store(logger.Sugar())
}

// SetFormat 运行时切换日志输出格式（JSON/Console，立即生效）。
// 非法值自动回退为 JSON。
func (c *ZClient) SetFormat(format string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if format != FormatConsole && format != FormatJSON {
		format = FormatJSON
	}
	c.config.Format = format
	logger := zap.New(
		zapcore.NewCore(
			zapEncoder(c.config.Format, c.encoderConfig),
			c.writer,
			c.atomicLevel,
		),
		c.zapOpts...,
	)
	c.logger.Store(logger)
	c.sugarLogger.Store(logger.Sugar())
}

// DebugFields 使用 zap.Field 输出 Debug 级别日志。
func (c *ZClient) DebugFields(msg string, args ...zap.Field) {
	c.logger.Load().Debug(msg, args...)
}

// InfoFields 使用 zap.Field 输出 Info 级别日志。
func (c *ZClient) InfoFields(msg string, args ...zap.Field) {
	c.logger.Load().Info(msg, args...)
}

// WarnFields 使用 zap.Field 输出 Warn 级别日志。
func (c *ZClient) WarnFields(msg string, args ...zap.Field) {
	c.logger.Load().Warn(msg, args...)
}

// ErrorFields 使用 zap.Field 输出 Error 级别日志。
func (c *ZClient) ErrorFields(msg string, args ...zap.Field) {
	c.logger.Load().Error(msg, args...)
}

// ============================================================
// 非格式化方法
// ============================================================

// Debug 输出 Debug 级别日志。
func (c *ZClient) Debug(args ...any) {
	c.sugarLogger.Load().Debug(args...)
}

// Info 输出 Info 级别日志。
func (c *ZClient) Info(args ...any) {
	c.sugarLogger.Load().Info(args...)
}

// Warn 输出 Warn 级别日志。
func (c *ZClient) Warn(args ...any) {
	c.sugarLogger.Load().Warn(args...)
}

// Error 输出 Error 级别日志。
func (c *ZClient) Error(args ...any) {
	c.sugarLogger.Load().Error(args...)
}

// ============================================================
// 格式化方法
// ============================================================

// Debugf 格式化输出 Debug 级别日志。
func (c *ZClient) Debugf(template string, args ...any) {
	c.sugarLogger.Load().Debugf(template, args...)
}

// Infof 格式化输出 Info 级别日志。
func (c *ZClient) Infof(template string, args ...any) {
	c.sugarLogger.Load().Infof(template, args...)
}

// Warnf 格式化输出 Warn 级别日志。
func (c *ZClient) Warnf(template string, args ...any) {
	c.sugarLogger.Load().Warnf(template, args...)
}

// Errorf 格式化输出 Error 级别日志。
func (c *ZClient) Errorf(template string, args ...any) {
	c.sugarLogger.Load().Errorf(template, args...)
}

// Printf 格式化输出 Info 级别日志（等价于 Infof）。
func (c *ZClient) Printf(template string, args ...any) {
	c.sugarLogger.Load().Infof(template, args...)
}

// ============================================================
// 键值对方法
// ============================================================

// Debugw 以键值对形式输出 Debug 级别日志。
func (c *ZClient) Debugw(msg string, keysAndValues ...any) {
	c.sugarLogger.Load().Debugw(msg, keysAndValues...)
}

// Infow 以键值对形式输出 Info 级别日志。
func (c *ZClient) Infow(msg string, keysAndValues ...any) {
	c.sugarLogger.Load().Infow(msg, keysAndValues...)
}

// Warnw 以键值对形式输出 Warn 级别日志。
func (c *ZClient) Warnw(msg string, keysAndValues ...any) {
	c.sugarLogger.Load().Warnw(msg, keysAndValues...)
}

// Errorw 以键值对形式输出 Error 级别日志。
func (c *ZClient) Errorw(msg string, keysAndValues ...any) {
	c.sugarLogger.Load().Errorw(msg, keysAndValues...)
}

// ============================================================
// Context 方法
// 注意：zap 原生不支持 context，ctx 参数仅用于保持接口一致性，实际被忽略。
// ============================================================

// DebugContext 带 context 输出 Debug 级别日志（ctx 被忽略）。
func (c *ZClient) DebugContext(ctx context.Context, args ...any) {
	c.sugarLogger.Load().Debug(args...)
}

// InfoContext 带 context 输出 Info 级别日志（ctx 被忽略）。
func (c *ZClient) InfoContext(ctx context.Context, args ...any) {
	c.sugarLogger.Load().Info(args...)
}

// WarnContext 带 context 输出 Warn 级别日志（ctx 被忽略）。
func (c *ZClient) WarnContext(ctx context.Context, args ...any) {
	c.sugarLogger.Load().Warn(args...)
}

// ErrorContext 带 context 输出 Error 级别日志（ctx 被忽略）。
func (c *ZClient) ErrorContext(ctx context.Context, args ...any) {
	c.sugarLogger.Load().Error(args...)
}

// DebugContextf 带 context 格式化输出 Debug 级别日志（ctx 被忽略）。
func (c *ZClient) DebugContextf(ctx context.Context, format string, args ...any) {
	c.sugarLogger.Load().Debugf(format, args...)
}

// InfoContextf 带 context 格式化输出 Info 级别日志（ctx 被忽略）。
func (c *ZClient) InfoContextf(ctx context.Context, format string, args ...any) {
	c.sugarLogger.Load().Infof(format, args...)
}

// WarnContextf 带 context 格式化输出 Warn 级别日志（ctx 被忽略）。
func (c *ZClient) WarnContextf(ctx context.Context, format string, args ...any) {
	c.sugarLogger.Load().Warnf(format, args...)
}

// ErrorContextf 带 context 格式化输出 Error 级别日志（ctx 被忽略）。
func (c *ZClient) ErrorContextf(ctx context.Context, format string, args ...any) {
	c.sugarLogger.Load().Errorf(format, args...)
}

// DebugContextw 带 context 以键值对形式输出 Debug 级别日志（ctx 被忽略）。
func (c *ZClient) DebugContextw(ctx context.Context, msg string, keysAndValues ...any) {
	c.sugarLogger.Load().Debugw(msg, keysAndValues...)
}

// InfoContextw 带 context 以键值对形式输出 Info 级别日志（ctx 被忽略）。
func (c *ZClient) InfoContextw(ctx context.Context, msg string, keysAndValues ...any) {
	c.sugarLogger.Load().Infow(msg, keysAndValues...)
}

// WarnContextw 带 context 以键值对形式输出 Warn 级别日志（ctx 被忽略）。
func (c *ZClient) WarnContextw(ctx context.Context, msg string, keysAndValues ...any) {
	c.sugarLogger.Load().Warnw(msg, keysAndValues...)
}

// ErrorContextw 带 context 以键值对形式输出 Error 级别日志（ctx 被忽略）。
func (c *ZClient) ErrorContextw(ctx context.Context, msg string, keysAndValues ...any) {
	c.sugarLogger.Load().Errorw(msg, keysAndValues...)
}

// ============================================================
// Fatal / Panic 方法
// ============================================================

// Fatal 输出 Fatal 级别日志后调用 os.Exit(1) 终止程序。
func (c *ZClient) Fatal(args ...any) {
	c.sugarLogger.Load().Fatal(args...)
}

// Fatalf 格式化输出 Fatal 级别日志后调用 os.Exit(1) 终止程序。
func (c *ZClient) Fatalf(format string, args ...any) {
	c.sugarLogger.Load().Fatalf(format, args...)
}

// Fatalw 以键值对形式输出 Fatal 级别日志后调用 os.Exit(1) 终止程序。
func (c *ZClient) Fatalw(msg string, keysAndValues ...any) {
	c.sugarLogger.Load().Fatalw(msg, keysAndValues...)
}

// Panic 输出 Panic 级别日志后触发 panic。
func (c *ZClient) Panic(args ...any) {
	c.sugarLogger.Load().Panic(args...)
}

// Panicf 格式化输出 Panic 级别日志后触发 panic。
func (c *ZClient) Panicf(format string, args ...any) {
	c.sugarLogger.Load().Panicf(format, args...)
}

// Panicw 以键值对形式输出 Panic 级别日志后触发 panic。
func (c *ZClient) Panicw(msg string, keysAndValues ...any) {
	c.sugarLogger.Load().Panicw(msg, keysAndValues...)
}

// ============================================================
// Trace 方法
// ============================================================

// Trace 等价于 Debug，适用于需要 trace 级别语义的场景。
func (c *ZClient) Trace(args ...any) {
	c.sugarLogger.Load().Debug(args...)
}

// Tracef 等价于 Debugf，格式化输出 trace 日志。
func (c *ZClient) Tracef(format string, args ...any) {
	c.sugarLogger.Load().Debugf(format, args...)
}

// Tracew 等价于 Debugw，以键值对形式输出 trace 日志。
func (c *ZClient) Tracew(msg string, keysAndValues ...any) {
	c.sugarLogger.Load().Debugw(msg, keysAndValues...)
}

// Init 创建并初始化 ZClient（默认使用 zap 后端，与 xfiber 兼容）。
// 如果不传 config 则使用默认配置（console 模式、debug 级别）。
func Init(configs ...*Config) (*ZClient, error) {
	return InitZap(configs...)
}

// 确保 ZClient 实现了 Logger 接口
var _ Logger = (*ZClient)(nil)
