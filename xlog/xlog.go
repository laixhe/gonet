package xlog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"time"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

/*
log:
  # 日志模式 console file
  run: console
  # 日志文件路径
  path: logs.log
  # 日志级别 debug info warn error
  level: debug
  # 每个日志文件保存大小 20M
  max_size: 20
  # 保留 N 个备份
  max_backups: 20
  # 保留 N 天
  max_age: 7
*/

const (
	RunTypeConsole = "console" // 终端
	RunTypeFile    = "file"    // 文件
)

const (
	LevelTypeDebug = "debug"
	LevelTypeInfo  = "info"
	LevelTypeWarn  = "warn"
	LevelTypeError = "error"
)

type Config struct {
	// 日志模式 console file
	Run string `json:"run,omitempty" mapstructure:"run" toml:"run" yaml:"run"`
	// 日志文件路径
	Path string `json:"path,omitempty" mapstructure:"path" toml:"path" yaml:"path"`
	// 日志级别 debug info warn error
	Level string `json:"level,omitempty" mapstructure:"level" toml:"level" yaml:"level"`
	// 单个日志文件最大（MB）
	MaxSize int `json:"max_size,omitempty" mapstructure:"max_size" toml:"max_size" yaml:"max_size"`
	// 保留的旧日志文件数
	MaxBackups int `json:"max_backups,omitempty" mapstructure:"max_backups" toml:"max_backups" yaml:"max_backups"`
	// 保留旧日志文件的最大天数
	MaxAge int `json:"max_age,omitempty" mapstructure:"max_age" toml:"max_age" yaml:"max_age"`
}

// Check 检查
func (c *Config) Check() error {
	if c == nil {
		return errors.New("没有日志配置")
	}
	if c.Run == "" {
		c.Run = RunTypeConsole
	}
	if c.Run == RunTypeFile {
		if c.Path == "" {
			c.Path = "logs.log"
		}
	}
	if !slices.Contains([]string{LevelTypeDebug, LevelTypeInfo, LevelTypeWarn, LevelTypeError}, c.Level) {
		c.Level = LevelTypeDebug
	}
	if c.MaxSize <= 0 {
		c.MaxSize = 3
	}
	if c.MaxBackups <= 0 {
		c.MaxBackups = 3
	}
	if c.MaxAge <= 0 {
		c.MaxAge = 3
	}
	return nil
}

// ContextHandler 包装一个 slog.Handler，在处理日志时自动从 context 中提取元素
type ContextHandler struct {
	slog.Handler
	ContextKey string
}

func (ch *ContextHandler) Handle(ctx context.Context, record slog.Record) error {
	if ctx != nil {
		if v, ok := ctx.Value(ch.ContextKey).(string); ok && v != "" {
			record.AddAttrs(slog.String(ch.ContextKey, v))
		}
	}
	return ch.Handler.Handle(ctx, record)
}

type LClient struct {
	config  *Config
	writer  io.Writer         // 日志写入接口
	handler *slog.JSONHandler // 配置 slog 格式处理器
}

func (lc *LClient) Writer() io.Writer {
	return lc.writer
}

func (lc *LClient) Handler() *slog.JSONHandler {
	return lc.handler
}

func Init(configs ...*Config) *LClient {
	lc := &LClient{}
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
	return lc
}
