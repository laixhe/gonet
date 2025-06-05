package xgorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"

	"github.com/laixhe/gonet/protocol/gen/config/clog"
	"github.com/laixhe/gonet/xgin/constant"
	"github.com/laixhe/gonet/xlog"
)

type DBWriter struct {
}

func NewDBWriter() *DBWriter {
	return &DBWriter{}
}

// Printf 格式化打印日志
func (w *DBWriter) Printf(message string, data ...interface{}) {
	if xlog.GetLevel() == clog.LevelType_debug.String() {
		xlog.Debug(fmt.Sprintf(message, data...))
	}
}

//

type DBlogger struct {
	logger.Writer
	logger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

func NewDBlogger(writer logger.Writer, config logger.Config) logger.Interface {
	return &DBlogger{
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

func (l *DBlogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

func (l *DBlogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.Printf("orm["+constant.HeaderRequestID+": %v] "+l.infoStr+msg, append([]interface{}{ctx.Value(constant.HeaderRequestID), utils.FileWithLineNum()}, data...)...)
	}
}

func (l *DBlogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.Printf("orm["+constant.HeaderRequestID+": %v] "+l.warnStr+msg, append([]interface{}{ctx.Value(constant.HeaderRequestID), utils.FileWithLineNum()}, data...)...)
	}
}

// Error print error messages
func (l *DBlogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.Printf("orm["+constant.HeaderRequestID+": %v] "+l.errStr+msg, append([]interface{}{ctx.Value(constant.HeaderRequestID), utils.FileWithLineNum()}, data...)...)
	}
}

func (l *DBlogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.Printf("orm["+constant.HeaderRequestID+": %v] "+l.traceErrStr, ctx.Value(constant.HeaderRequestID), utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf("orm["+constant.HeaderRequestID+": %v] "+l.traceErrStr, ctx.Value(constant.HeaderRequestID), utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Printf("orm["+constant.HeaderRequestID+": %v] "+l.traceWarnStr, ctx.Value(constant.HeaderRequestID), utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf("orm["+constant.HeaderRequestID+": %v] "+l.traceWarnStr, ctx.Value(constant.HeaderRequestID), utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.Printf("orm["+constant.HeaderRequestID+": %v] "+l.traceStr, ctx.Value(constant.HeaderRequestID), utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf("orm["+constant.HeaderRequestID+": %v] "+l.traceStr, ctx.Value(constant.HeaderRequestID), utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

func (l *DBlogger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.Config.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}
