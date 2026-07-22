// Package xlog 提供基于 zap 和 slog 的统一日志客户端，支持 JSON/Console 输出格式、终端/文件输出、日志轮转和运行时切换级别。
//
// 快速开始（zap 后端）：
//
//	logs, _ := xlog.InitZap(&xlog.Config{Level: xlog.LevelTypeDebug})
//	defer logs.Sync()
//
// 快速开始（slog 后端）：
//
//	logs, _ := xlog.InitSlog(&xlog.Config{Level: xlog.LevelTypeDebug})
//	defer logs.Sync()
package xlog

import (
	"context"
	"errors"
)

// backgroundCtx 缓存 context.Background()，避免每次日志调用时重复创建。
var backgroundCtx = context.Background()

const (
	RunTypeConsole = "console" // 终端输出
	RunTypeFile    = "file"    // 文件输出
)

const (
	// FormatJSON JSON 格式输出
	FormatJSON = "json"
	// FormatConsole 控制台友好格式输出
	FormatConsole = "console"
)

const (
	// LevelTypeDebug debug 级别
	LevelTypeDebug = "debug"
	// LevelTypeInfo info 级别
	LevelTypeInfo = "info"
	// LevelTypeWarn warn 级别
	LevelTypeWarn = "warn"
	// LevelTypeError error 级别
	LevelTypeError = "error"
)

// validLevels 合法日志级别集合
var validLevels = map[string]bool{
	LevelTypeDebug: true,
	LevelTypeInfo:  true,
	LevelTypeWarn:  true,
	LevelTypeError: true,
}

// validFormats 合法日志格式集合
var validFormats = map[string]bool{
	FormatJSON:    true,
	FormatConsole: true,
}

// validRuns 合法日志运行模式集合
var validRuns = map[string]bool{
	RunTypeConsole: true,
	RunTypeFile:    true,
}

// Logger 统一日志接口，SClient 和 ZClient 均实现该接口。
//
// 该接口的方法签名与以下框架的日志接口兼容（无直接依赖）：
//   - Fiber (v2/v3): 与 log.CommonLogger 签名一致，可直接用 log.SetLogger(client) 集成
//   - Gin: 可通过 gin.DefaultWriter = client.Writer() 集成
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

// Config 日志配置。
//
// 支持的 YAML 配置格式：
//
//	log:
//	  run: console          # 日志模式：console（终端）或 file（文件）
//	  format: json          # 输出格式：json 或 console（默认 json）
//	  path: logs.log        # 日志文件路径（仅 file 模式生效）
//	  level: debug          # 日志级别：debug / info / warn / error
//	  max_size: 20          # 单个日志文件最大大小（MB）
//	  max_backups: 20       # 保留的旧日志文件数
//	  max_age: 7            # 保留旧日志文件的最大天数
type Config struct {
	// 日志模式：console 或 file
	Run string `json:"run,omitempty" mapstructure:"run" toml:"run" yaml:"run"`
	// 输出格式：json 或 console（默认 json）
	Format string `json:"format,omitempty" mapstructure:"format" toml:"format" yaml:"format"`
	// 日志文件路径（仅 file 模式生效）
	Path string `json:"path,omitempty" mapstructure:"path" toml:"path" yaml:"path"`
	// 日志级别：debug / info / warn / error
	Level string `json:"level,omitempty" mapstructure:"level" toml:"level" yaml:"level"`
	// 单个日志文件最大大小（MB）
	MaxSize int `json:"max_size,omitempty" mapstructure:"max_size" toml:"max_size" yaml:"max_size"`
	// 保留的旧日志文件数
	MaxBackups int `json:"max_backups,omitempty" mapstructure:"max_backups" toml:"max_backups" yaml:"max_backups"`
	// 保留旧日志文件的最大天数
	MaxAge int `json:"max_age,omitempty" mapstructure:"max_age" toml:"max_age" yaml:"max_age"`
	// 调用者跳过帧数（用于调整 caller 信息）
	CallerSkip int `json:"caller_skip,omitempty" mapstructure:"caller_skip" toml:"caller_skip" yaml:"caller_skip"`
	// 是否对旧日志文件进行 gzip 压缩
	Compress bool `json:"compress,omitempty" mapstructure:"compress" toml:"compress" yaml:"compress"`
	// 是否使用本地时间作为日志文件时间戳
	LocalTime bool `json:"local_time,omitempty" mapstructure:"local_time" toml:"local_time" yaml:"local_time"`
	// slog context 提取 key（仅对 SClient 生效）
	ContextKey string `json:"context_key,omitempty" mapstructure:"context_key" toml:"context_key" yaml:"context_key"`
}

// Check 校验并补全配置的默认值。
func (c *Config) Check() error {
	if c == nil {
		return errors.New("没有日志配置")
	}
	if c.Run == "" {
		c.Run = RunTypeConsole
	}
	if !validRuns[c.Run] {
		c.Run = RunTypeConsole
	}
	if c.Run == RunTypeFile {
		if c.Path == "" {
			c.Path = "logs.log"
		}
	}
	if !validLevels[c.Level] {
		c.Level = LevelTypeDebug
	}
	if !validFormats[c.Format] {
		c.Format = FormatJSON
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
