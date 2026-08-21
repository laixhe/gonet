package orm

import (
	"context"
	"errors"
	"fmt"
	"time"

	gormLogger "gorm.io/gorm/logger"
	gormUtils "gorm.io/gorm/utils"
)

// contextKey 自定义 context key 类型, 避免直接使用字符串 key 与其他库冲突
type contextKey string

// CtxRequestId 请求 id 在 context 中的 key
// 通过 WithRequestId 写入, 日志器中通过该 key 读取
const CtxRequestId contextKey = "requestId"

// WithRequestId 将请求 id 写入 context
func WithRequestId(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, CtxRequestId, requestId)
}

type Logger struct {
	RequestId string
	gormLogger.Writer
	gormLogger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

func NewLogger(writer gormLogger.Writer, config gormLogger.Config, requestId string) gormLogger.Interface {
	return &Logger{
		RequestId:    requestId,
		Writer:       writer,
		Config:       config,
		infoStr:      "[file: %s] [info] ",
		warnStr:      "[file: %s] [warn] ",
		errStr:       "[file: %s] [error] ",
		traceStr:     "[file: %s] [time: %.3fms] [rows: %v] [sql: %s]",
		traceWarnStr: "[file: %s] [slow: %s] [time: %.3fms] [rows: %v] [sql: %s]",
		traceErrStr:  "[file: %s] [slow: %s] [time: %.3fms] [rows: %v] [sql: %s]",
	}
}

// getRequestId 从 context 中获取请求 id
// 优先使用 Init 传入的字符串 key(兼容 xfiber/xgin 等现有集成), 再使用类型化 key CtxRequestId
func (l *Logger) getRequestId(ctx context.Context) any {
	if v := ctx.Value(l.RequestId); v != nil {
		return v
	}
	return ctx.Value(CtxRequestId)
}

func (l *Logger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *Logger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormLogger.Info {
		l.Printf("orm["+l.RequestId+": %v] "+l.infoStr+msg, append([]any{l.getRequestId(ctx), gormUtils.FileWithLineNum()}, data...)...)
	}
}

func (l *Logger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormLogger.Warn {
		l.Printf("orm["+l.RequestId+": %v] "+l.warnStr+msg, append([]any{l.getRequestId(ctx), gormUtils.FileWithLineNum()}, data...)...)
	}
}

func (l *Logger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormLogger.Error {
		l.Printf("orm["+l.RequestId+": %v] "+l.errStr+msg, append([]any{l.getRequestId(ctx), gormUtils.FileWithLineNum()}, data...)...)
	}
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormLogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormLogger.Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.Printf("orm["+l.RequestId+": %v] "+l.traceErrStr, l.getRequestId(ctx), gormUtils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf("orm["+l.RequestId+": %v] "+l.traceErrStr, l.getRequestId(ctx), gormUtils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormLogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Printf("orm["+l.RequestId+": %v] "+l.traceWarnStr, l.getRequestId(ctx), gormUtils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf("orm["+l.RequestId+": %v] "+l.traceWarnStr, l.getRequestId(ctx), gormUtils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == gormLogger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.Printf("orm["+l.RequestId+": %v] "+l.traceStr, l.getRequestId(ctx), gormUtils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf("orm["+l.RequestId+": %v] "+l.traceStr, l.getRequestId(ctx), gormUtils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

func (l *Logger) ParamsFilter(ctx context.Context, sql string, params ...any) (string, []any) {
	if l.Config.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}
