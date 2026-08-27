package orm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	gormLogger "gorm.io/gorm/logger"
)

// captureWriter 捕获日志输出的测试 Writer
type captureWriter struct {
	buf bytes.Buffer
}

func (w *captureWriter) Printf(format string, args ...any) {
	fmt.Fprintf(&w.buf, format, args...)
}

// nopWriter 丢弃日志输出的测试 Writer
type nopWriter struct{}

func (nopWriter) Printf(string, ...any) {}

func newTestLogger(w *captureWriter, level gormLogger.LogLevel) *Logger {
	return NewLogger(w, gormLogger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      level,
	}, "requestId").(*Logger)
}

func TestLoggerLevelGating(t *testing.T) {
	t.Run("Info 级别以下不输出", func(t *testing.T) {
		w := &captureWriter{}
		l := newTestLogger(w, gormLogger.Error)
		l.Info(context.Background(), "info msg")
		if w.buf.Len() != 0 {
			t.Errorf("Error 级别不应输出 Info 日志: %s", w.buf.String())
		}
	})
	t.Run("Info 级别输出", func(t *testing.T) {
		w := &captureWriter{}
		l := newTestLogger(w, gormLogger.Info)
		l.Info(context.Background(), "info msg")
		out := w.buf.String()
		if !strings.Contains(out, "[info]") || !strings.Contains(out, "info msg") {
			t.Errorf("Info 日志输出异常: %s", out)
		}
	})
	t.Run("Warn 级别输出", func(t *testing.T) {
		w := &captureWriter{}
		l := newTestLogger(w, gormLogger.Warn)
		l.Warn(context.Background(), "warn msg")
		out := w.buf.String()
		if !strings.Contains(out, "[warn]") || !strings.Contains(out, "warn msg") {
			t.Errorf("Warn 日志输出异常: %s", out)
		}
	})
	t.Run("Error 级别输出", func(t *testing.T) {
		w := &captureWriter{}
		l := newTestLogger(w, gormLogger.Error)
		l.Error(context.Background(), "err msg")
		out := w.buf.String()
		if !strings.Contains(out, "[error]") || !strings.Contains(out, "err msg") {
			t.Errorf("Error 日志输出异常: %s", out)
		}
	})
}

func TestLoggerRequestId(t *testing.T) {
	t.Run("类型化 key", func(t *testing.T) {
		w := &captureWriter{}
		l := newTestLogger(w, gormLogger.Info)
		ctx := WithRequestId(context.Background(), "typed-123")
		l.Info(ctx, "hello")
		out := w.buf.String()
		if !strings.Contains(out, "orm[requestId: typed-123]") {
			t.Errorf("应使用类型化 key 获取请求 id: %s", out)
		}
	})
	t.Run("字符串 key 兼容", func(t *testing.T) {
		w := &captureWriter{}
		l := newTestLogger(w, gormLogger.Info)
		// 模拟 xfiber/xgin 等使用字符串 key 写入的请求 id
		ctx := context.WithValue(context.Background(), "requestId", "str-456")
		l.Info(ctx, "hello")
		out := w.buf.String()
		if !strings.Contains(out, "orm[requestId: str-456]") {
			t.Errorf("应兼容字符串 key: %s", out)
		}
	})
	t.Run("无请求 id 输出 nil", func(t *testing.T) {
		w := &captureWriter{}
		l := newTestLogger(w, gormLogger.Info)
		l.Info(context.Background(), "hello")
		out := w.buf.String()
		if !strings.Contains(out, "orm[requestId: <nil>]") {
			t.Errorf("无请求 id 时应输出 <nil>: %s", out)
		}
	})
}

func TestLoggerTrace(t *testing.T) {
	t.Run("错误路径", func(t *testing.T) {
		w := &captureWriter{}
		l := newTestLogger(w, gormLogger.Info)
		begin := time.Now()
		l.Trace(context.Background(), begin, func() (string, int64) { return "SELECT 1", -1 }, errors.New("boom"))
		out := w.buf.String()
		// 错误信息打印在 [slow:] 位置
		if !strings.Contains(out, "boom") || !strings.Contains(out, "SELECT 1") || !strings.Contains(out, "[rows: -]") {
			t.Errorf("Trace 错误路径输出异常: %s", out)
		}
	})
	t.Run("慢查询路径", func(t *testing.T) {
		w := &captureWriter{}
		l := newTestLogger(w, gormLogger.Warn)
		begin := time.Now().Add(-time.Second) // 已耗时 1 秒, 超过 200ms 阈值
		l.Trace(context.Background(), begin, func() (string, int64) { return "SELECT 1", 2 }, nil)
		out := w.buf.String()
		if !strings.Contains(out, "SLOW SQL") || !strings.Contains(out, "SELECT 1") {
			t.Errorf("Trace 慢查询输出异常: %s", out)
		}
	})
	t.Run("正常路径", func(t *testing.T) {
		w := &captureWriter{}
		l := newTestLogger(w, gormLogger.Info)
		begin := time.Now()
		l.Trace(context.Background(), begin, func() (string, int64) { return "SELECT 1", 2 }, nil)
		out := w.buf.String()
		if !strings.Contains(out, "[rows: 2]") || !strings.Contains(out, "SELECT 1") {
			t.Errorf("Trace 正常输出异常: %s", out)
		}
	})
	t.Run("Silent 级别不输出", func(t *testing.T) {
		w := &captureWriter{}
		l := newTestLogger(w, gormLogger.Silent)
		l.Trace(context.Background(), time.Now(), func() (string, int64) { return "SELECT 1", 1 }, nil)
		if w.buf.Len() != 0 {
			t.Errorf("Silent 级别不应输出: %s", w.buf.String())
		}
	})
}

func TestLoggerLogMode(t *testing.T) {
	w := &captureWriter{}
	l := newTestLogger(w, gormLogger.Info)
	mode := l.LogMode(gormLogger.Silent).(*Logger)
	if mode.LogLevel != gormLogger.Silent {
		t.Errorf("LogMode 后 LogLevel = %d, want %d", mode.LogLevel, gormLogger.Silent)
	}
	// 原始 logger 不受影响
	if l.LogLevel != gormLogger.Info {
		t.Errorf("原始 LogLevel = %d, want %d", l.LogLevel, gormLogger.Info)
	}
}

func TestGetRequestId(t *testing.T) {
	l := &Logger{RequestId: "requestId"}
	if got := l.getRequestId(context.Background()); got != nil {
		t.Errorf("空 context 应为 nil, got %v", got)
	}
	ctx := WithRequestId(context.Background(), "v1")
	if got := l.getRequestId(ctx); got != "v1" {
		t.Errorf("类型化 key 读取 = %v, want v1", got)
	}
}
