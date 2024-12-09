package xgorm

import (
	"fmt"

	"gorm.io/gorm/logger"

	"github.com/laixhe/gonet/protocol/gen/config/clog"
	"github.com/laixhe/gonet/xlog"
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
	if xlog.GetLevel() == clog.LevelType_debug.String() {
		xlog.Debug(fmt.Sprintf(message, data...))
	}
}
