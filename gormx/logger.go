package gormx

import (
	"fmt"

	"gorm.io/gorm/logger"

	"github.com/laixhe/gonet/logx"
	"github.com/laixhe/gonet/proto/gen/config/clog"
)

type Writer struct {
	logger.Writer
}

// newWriter writer 构造函数
func newWriter(w logger.Writer) *Writer {
	return &Writer{Writer: w}
}

// Printf 格式化打印日志
func (w *Writer) Printf(message string, data ...interface{}) {
	if logx.GetLevel() == clog.LevelType_debug.String() {
		logx.Debug(fmt.Sprintf(message, data...))
	}
}
