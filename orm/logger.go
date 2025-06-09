package orm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type OrmLogger struct {
	Requestid string
	logger.Writer
	logger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

func NewOrmLogger(writer logger.Writer, config logger.Config, requestid string) logger.Interface {
	return &OrmLogger{
		Requestid:    requestid,
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

func (l *OrmLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

func (l *OrmLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.Printf("orm["+l.Requestid+": %v] "+l.infoStr+msg, append([]interface{}{ctx.Value(l.Requestid), utils.FileWithLineNum()}, data...)...)
	}
}

func (l *OrmLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.Printf("orm["+l.Requestid+": %v] "+l.warnStr+msg, append([]interface{}{ctx.Value(l.Requestid), utils.FileWithLineNum()}, data...)...)
	}
}

func (l *OrmLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.Printf("orm["+l.Requestid+": %v] "+l.errStr+msg, append([]interface{}{ctx.Value(l.Requestid), utils.FileWithLineNum()}, data...)...)
	}
}

func (l *OrmLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.Printf("orm["+l.Requestid+": %v] "+l.traceErrStr, ctx.Value(l.Requestid), utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf("orm["+l.Requestid+": %v] "+l.traceErrStr, ctx.Value(l.Requestid), utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Printf("orm["+l.Requestid+": %v] "+l.traceWarnStr, ctx.Value(l.Requestid), utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf("orm["+l.Requestid+": %v] "+l.traceWarnStr, ctx.Value(l.Requestid), utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.Printf("orm["+l.Requestid+": %v] "+l.traceStr, ctx.Value(l.Requestid), utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf("orm["+l.Requestid+": %v] "+l.traceStr, ctx.Value(l.Requestid), utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

func (l *OrmLogger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.Config.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}
