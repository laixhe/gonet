package orm

import (
	"context"
	"errors"
	"fmt"
	"time"

	gormLogger "gorm.io/gorm/logger"
	gormUtils "gorm.io/gorm/utils"
)

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

func (l *Logger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *Logger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormLogger.Info {
		l.Printf("orm["+l.RequestId+": %v] "+l.infoStr+msg, append([]any{ctx.Value(l.RequestId), gormUtils.FileWithLineNum()}, data...)...)
	}
}

func (l *Logger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormLogger.Warn {
		l.Printf("orm["+l.RequestId+": %v] "+l.warnStr+msg, append([]any{ctx.Value(l.RequestId), gormUtils.FileWithLineNum()}, data...)...)
	}
}

func (l *Logger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormLogger.Error {
		l.Printf("orm["+l.RequestId+": %v] "+l.errStr+msg, append([]any{ctx.Value(l.RequestId), gormUtils.FileWithLineNum()}, data...)...)
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
			l.Printf("orm["+l.RequestId+": %v] "+l.traceErrStr, ctx.Value(l.RequestId), gormUtils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf("orm["+l.RequestId+": %v] "+l.traceErrStr, ctx.Value(l.RequestId), gormUtils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormLogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Printf("orm["+l.RequestId+": %v] "+l.traceWarnStr, ctx.Value(l.RequestId), gormUtils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf("orm["+l.RequestId+": %v] "+l.traceWarnStr, ctx.Value(l.RequestId), gormUtils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == gormLogger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.Printf("orm["+l.RequestId+": %v] "+l.traceStr, ctx.Value(l.RequestId), gormUtils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf("orm["+l.RequestId+": %v] "+l.traceStr, ctx.Value(l.RequestId), gormUtils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

func (l *Logger) ParamsFilter(ctx context.Context, sql string, params ...any) (string, []any) {
	if l.Config.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}
