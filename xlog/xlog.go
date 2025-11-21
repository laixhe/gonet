package xlog

import (
	"errors"
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

type LogClient struct {
	config      *Config
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
}

func (c *LogClient) Logger() *zap.Logger {
	return c.logger
}

func (c *LogClient) SugaredLogger() *zap.SugaredLogger {
	return c.sugarLogger
}

func (c *LogClient) Level() string {
	return c.config.Level
}

// Init 初始日志
func Init(config *Config) (*LogClient, error) {
	if err := config.Check(); err != nil {
		return nil, err
	}
	client := &LogClient{
		config: config,
	}
	if config.Run == RunTypeFile {
		initSizeFile(client)
	} else {
		initConsole(client)
	}
	return client, nil
}

// initSizeFile 初始文件日志，按大小切割和备份个数、文件有效期
func initSizeFile(c *LogClient) {
	// 日志分割
	hook := &lumberjack.Logger{
		Filename:   c.config.Path, // 日志文件路径，默认 os.TempDir()
		MaxSize:    c.config.MaxSize,
		MaxBackups: c.config.MaxBackups,
		MaxAge:     c.config.MaxAge,
		Compress:   false,
	}
	// 打印到文件
	write := zapcore.AddSync(hook)
	// 初始日志
	zapInit(c, write)
}

// initConsole 初始终端日志，输出到终端
func initConsole(c *LogClient) {
	// 打印到控制台
	write := zapcore.AddSync(os.Stdout)
	// 初始日志
	zapInit(c, write)
}

// zapInit 初始化 zap 基本信息
// write       文件描述符
// serviceName 服务名
// logLevel    日志级别
// callerSkip 提升的堆栈帧数，0=当前函数，1=上一层函数，....
func zapInit(c *LogClient, write zapcore.WriteSyncer) {
	// 设置日志级别
	var level zapcore.Level
	switch c.config.Level {
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
	zapOptions := make([]zap.Option, 0, 5)
	// 开启开发模式，堆栈跟踪
	zapOptions = append(zapOptions, zap.Development())
	// 开启文件及行号
	zapOptions = append(zapOptions, zap.AddCaller())
	// 提升打印的堆栈帧数
	zapOptions = append(zapOptions, zap.AddCallerSkip(c.config.CallerSkip+1))
	// 添加字段-服务器名称
	// zapOptions = append(zapOptions, zap.Fields(zap.String("service", serviceName)))
	// 构造日志
	//client.logger = zap.New(core, zapOptions...)
	c.logger = zap.New(core, zapOptions...)
	c.sugarLogger = c.logger.Sugar()
}

// zapTimeEncoder 日志时间格式
func zapTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.DateTime))
}

// Debug 调试
func (c *LogClient) Debug(msg string, args ...zap.Field) {
	c.logger.Debug(msg, args...)
}

// Debugf 调试
func (c *LogClient) Debugf(template string, args ...interface{}) {
	c.sugarLogger.Debugf(template, args...)
}

// Info 信息
func (c *LogClient) Info(msg string, args ...zap.Field) {
	c.logger.Info(msg, args...)
}

// Infof 信息
func (c *LogClient) Infof(template string, args ...interface{}) {
	c.sugarLogger.Infof(template, args...)
}

// Warn 警告
func (c *LogClient) Warn(msg string, args ...zap.Field) {
	c.logger.Warn(msg, args...)
}

// Warnf 警告
func (c *LogClient) Warnf(template string, args ...interface{}) {
	c.sugarLogger.Warnf(template, args...)
}

// Error 错误
func (c *LogClient) Error(msg string, args ...zap.Field) {
	c.logger.Error(msg, args...)
}

// Errorf 错误
func (c *LogClient) Errorf(template string, args ...interface{}) {
	c.sugarLogger.Errorf(template, args...)
}

// Print 打印
func (c *LogClient) Printf(template string, args ...interface{}) {
	c.sugarLogger.Debugf(template, args...)
}
