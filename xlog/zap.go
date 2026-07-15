package xlog

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type ZClient struct {
	config      *Config
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
}

func (c *ZClient) Logger() *zap.Logger {
	return c.logger
}

func (c *ZClient) SugaredLogger() *zap.SugaredLogger {
	return c.sugarLogger
}

func (c *ZClient) Level() string {
	return c.config.Level
}

func InitZap(config *Config) (*ZClient, error) {
	if err := config.Check(); err != nil {
		return nil, err
	}
	client := &ZClient{
		config: config,
	}
	if config.Run == RunTypeFile {
		initZapSizeFile(client)
	} else {
		initZapConsole(client)
	}
	return client, nil
}

// initZapSizeFile 初始文件日志，按大小切割和备份个数、文件有效期
func initZapSizeFile(c *ZClient) {
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

// initZapConsole 初始终端日志，输出到终端
func initZapConsole(c *ZClient) {
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
func zapInit(c *ZClient, write zapcore.WriteSyncer) {
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

func (c *ZClient) Debug(msg string, args ...zap.Field) {
	c.logger.Debug(msg, args...)
}

func (c *ZClient) Debugf(template string, args ...interface{}) {
	c.sugarLogger.Debugf(template, args...)
}

func (c *ZClient) Info(msg string, args ...zap.Field) {
	c.logger.Info(msg, args...)
}

func (c *ZClient) Infof(template string, args ...interface{}) {
	c.sugarLogger.Infof(template, args...)
}

func (c *ZClient) Warn(msg string, args ...zap.Field) {
	c.logger.Warn(msg, args...)
}

func (c *ZClient) Warnf(template string, args ...interface{}) {
	c.sugarLogger.Warnf(template, args...)
}

func (c *ZClient) Error(msg string, args ...zap.Field) {
	c.logger.Error(msg, args...)
}

func (c *ZClient) Errorf(template string, args ...interface{}) {
	c.sugarLogger.Errorf(template, args...)
}

func (c *ZClient) Printf(template string, args ...interface{}) {
	c.sugarLogger.Debugf(template, args...)
}
