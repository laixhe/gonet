package xlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// SlogContextHandler 包装一个 slog.Handler，在处理日志时自动从 context 中提取元素
type SlogContextHandler struct {
	slog.Handler
	ContextKey string
}

func (ch *SlogContextHandler) Handle(ctx context.Context, record slog.Record) error {
	if ctx != nil {
		if v, ok := ctx.Value(ch.ContextKey).(string); ok && v != "" {
			record.AddAttrs(slog.String(ch.ContextKey, v))
		}
	}
	return ch.Handler.Handle(ctx, record)
}

type SClient struct {
	config  *Config
	writer  io.Writer         // 日志写入接口
	handler *slog.JSONHandler // 配置 slog 格式处理器
	logger  *slog.Logger
}

func (lc *SClient) Writer() io.Writer {
	return lc.writer
}

func (lc *SClient) Handler() *slog.JSONHandler {
	return lc.handler
}

func (lc *SClient) Logger() *slog.Logger {
	return lc.logger
}

func InitSlog(configs ...*Config) *SClient {
	lc := &SClient{}
	if len(configs) == 0 {
		lc.config = &Config{}
	} else {
		lc.config = configs[0]
	}
	if lc.config.Run == RunTypeFile {
		// 日志分割
		lc.writer = &lumberjack.Logger{
			Filename:   lc.config.Path,       // 日志文件路径，为空时默认使用 os.TempDir() + 进程名
			MaxSize:    lc.config.MaxSize,    // 单个日志文件最大（MB）
			MaxBackups: lc.config.MaxBackups, // 保留的旧日志文件数
			MaxAge:     lc.config.MaxAge,     // 保留旧日志文件的最大天数
			LocalTime:  true,                 // 是否使用 本地时间 作为日志文件时间戳，否则使用 UTC 时间
			Compress:   false,                // 是否对 旧日志进行 gzip 压缩
		}
	} else {
		lc.writer = os.Stdout
	}
	options := &slog.HandlerOptions{
		// 显示调用日志的位置
		AddSource: true,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			// 处理时间格式
			if attr.Key == slog.TimeKey {
				attr.Value = slog.StringValue(attr.Value.Time().Format(time.DateTime))
			}
			// 简化调用源信息，只保留文件名和行号
			if attr.Key == slog.SourceKey {
				source, ok := attr.Value.Any().(*slog.Source)
				if ok {
					shortFile := source.File
					shortFileCount := 0
					for i := len(source.File) - 1; i > 0; i-- {
						if shortFileCount <= 2 {
							if source.File[i] == '/' {
								shortFile = source.File[i+1:]
								shortFileCount++
							}
						}
					}
					return slog.String("source", fmt.Sprintf("%s:%d", shortFile, source.Line))
				}
			}
			return attr
		},
	}
	switch lc.config.Level {
	case LevelTypeDebug:
		options.Level = slog.LevelDebug
	case LevelTypeInfo:
		options.Level = slog.LevelInfo
	case LevelTypeWarn:
		options.Level = slog.LevelWarn
	case LevelTypeError:
		options.Level = slog.LevelError
	default:
		options.Level = slog.LevelDebug
	}
	// 配置 JSON 格式处理器
	lc.handler = slog.NewJSONHandler(lc.writer, options)
	lc.logger = slog.New(lc.handler)
	return lc
}
