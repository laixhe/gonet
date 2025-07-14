package log

import (
	"errors"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

// 日志配置
type Config struct {
	// 日志模式 console file
	Run string `json:"run,omitempty" mapstructure:"run" toml:"run" yaml:"run"`
	// 日志文件路径
	Path string `json:"path,omitempty" mapstructure:"path" toml:"path" yaml:"path"`
	// 日志级别 debug info warn error
	Level string `json:"level,omitempty" mapstructure:"level" toml:"level" yaml:"level"`
	// 每个日志文件保存大小 *M
	MaxSize int `json:"max_size,omitempty" mapstructure:"max_size" toml:"max_size" yaml:"max_size"`
	// 保留 N 个备份
	MaxBackups int `json:"max_backups,omitempty" mapstructure:"max_backups" toml:"max_backups" yaml:"max_backups"`
	// 保留 N 天
	MaxAge int `json:"max_age,omitempty" mapstructure:"max_age" toml:"max_age" yaml:"max_age"`
	// 堆栈帧数
	CallerSkip int `json:"caller_skip,omitempty" mapstructure:"caller_skip" toml:"caller_skip" yaml:"caller_skip"`
}

// Checking 检查
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
	if !(c.Level == LevelTypeDebug ||
		c.Level == LevelTypeInfo ||
		c.Level == LevelTypeWarn ||
		c.Level == LevelTypeError) {
		c.Level = LevelTypeDebug
	}
	return nil
}

type ZapClient struct {
	config      *Config
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
}

func (zc *ZapClient) Logger() *zap.Logger {
	return zc.logger
}

func (zc *ZapClient) SugaredLogger() *zap.SugaredLogger {
	return zc.sugarLogger
}

func (zc *ZapClient) Level() string {
	return zc.config.Level
}

// Init 初始日志
func Init(config *Config) (*ZapClient, error) {
	if err := config.Check(); err != nil {
		return nil, err
	}
	client := &ZapClient{
		config: config,
	}
	if config.Run == RunTypeFile {
		initSizeFile(client)
	} else {
		initConsole(client)
	}
	return client, nil
}

// initSizeFile 初始 zap 日志，按大小切割和备份个数、文件有效期
func initSizeFile(zc *ZapClient) {
	// 日志分割
	hook := &lumberjack.Logger{
		Filename:   zc.config.Path, // 日志文件路径，默认 os.TempDir()
		MaxSize:    zc.config.MaxSize,
		MaxBackups: zc.config.MaxBackups,
		MaxAge:     zc.config.MaxAge,
		Compress:   false,
	}
	// 打印到文件
	write := zapcore.AddSync(hook)
	// 初始 zap 日志
	zapInit(zc, write)
}

// initConsole 初始 zap 日志，输出到终端
func initConsole(zc *ZapClient) {
	// 打印到控制台
	write := zapcore.AddSync(os.Stdout)
	// 初始 zap 日志
	zapInit(zc, write)
}

// zapInit 初始化 zap 基本信息
// write       文件描述符
// serviceName 服务名
// logLevel    日志级别
// callerSkip 提升的堆栈帧数，0=当前函数，1=上一层函数，....
func zapInit(zc *ZapClient, write zapcore.WriteSyncer) {
	// 设置日志级别
	var level zapcore.Level
	switch zc.config.Level {
	case LevelTypeDebug:
		level = zap.DebugLevel
	case LevelTypeInfo:
		level = zap.InfoLevel
	case LevelTypeWarn:
		level = zap.WarnLevel
	case LevelTypeError:
		level = zap.ErrorLevel
	default:
		level = zap.DebugLevel
	}
	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "call",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     zapTimeEncoder,                 // 日志时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, // 执行消耗的时间转化成浮点型的秒
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 短路径编码器,格式化调用堆栈
		EncodeName:     zapcore.FullNameEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		write,
		level,
	)
	// 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()
	// 提升打印的堆栈帧数
	addCallerSkip := zap.AddCallerSkip(zc.config.CallerSkip + 1)
	// 开启文件及行号
	development := zap.Development()
	// 添加字段-服务器名称
	//filed := zap.Fields(zap.String("service", serviceName))
	// 构造日志
	//client.logger = zap.New(core, caller, addCallerSkip, development, filed)
	zc.logger = zap.New(core, caller, addCallerSkip, development)
	zc.sugarLogger = zc.logger.Sugar()
}

// zapTimeEncoder 日志时间格式
func zapTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.DateTime))
}

// Debug 调试
func (zc *ZapClient) Debug(msg string, args ...zap.Field) {
	zc.logger.Debug(msg, args...)
}

// Debugf 调试
func (zc *ZapClient) Debugf(template string, args ...interface{}) {
	zc.sugarLogger.Debugf(template, args...)
}

// Info 信息
func (zc *ZapClient) Info(msg string, args ...zap.Field) {
	zc.logger.Info(msg, args...)
}

// Infof 信息
func (zc *ZapClient) Infof(template string, args ...interface{}) {
	zc.sugarLogger.Infof(template, args...)
}

// Warn 警告
func (zc *ZapClient) Warn(msg string, args ...zap.Field) {
	zc.logger.Warn(msg, args...)
}

// Warnf 警告
func (zc *ZapClient) Warnf(template string, args ...interface{}) {
	zc.sugarLogger.Warnf(template, args...)
}

// Error 错误
func (zc *ZapClient) Error(msg string, args ...zap.Field) {
	zc.logger.Error(msg, args...)
}

// Errorf 错误
func (zc *ZapClient) Errorf(template string, args ...interface{}) {
	zc.sugarLogger.Errorf(template, args...)
}

// Print 打印
func (zc *ZapClient) Printf(template string, args ...interface{}) {
	zc.logger.Debug(fmt.Sprintf(template, args...))
}
