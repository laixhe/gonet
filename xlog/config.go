package xlog

import (
	"errors"
	"slices"
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
	// 堆栈帧数
	CallerSkip int `json:"caller_skip,omitempty" mapstructure:"caller_skip" toml:"caller_skip" yaml:"caller_skip"`
}

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
