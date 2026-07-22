package xlog

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

// testWriter 实现 io.Writer 和 Sync，用于捕获日志输出。
type testWriter struct {
	buf bytes.Buffer
}

func (w *testWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *testWriter) Sync() error {
	return nil
}

// ============================================================
// InitSlog 测试
// ============================================================

func TestInitSlog(t *testing.T) {
	t.Run("no config uses defaults", func(t *testing.T) {
		lc, err := InitSlog()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer lc.Sync()
		if lc.Level() != LevelTypeDebug {
			t.Errorf("level = %s, want debug", lc.Level())
		}
	})

	t.Run("with config", func(t *testing.T) {
		lc, err := InitSlog(&Config{Level: LevelTypeInfo, Format: FormatConsole})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer lc.Sync()
		if lc.Level() != LevelTypeInfo {
			t.Errorf("level = %s, want info", lc.Level())
		}
	})

	t.Run("nil config returns error", func(t *testing.T) {
		_, err := InitSlog(nil)
		if err == nil {
			t.Fatal("expected error for nil config")
		}
	})
}

// ============================================================
// atomicSlogLevel 测试
// ============================================================

func TestAtomicSlogLevel(t *testing.T) {
	a := &atomicSlogLevel{}
	if a.Level() != slog.Level(0) {
		t.Errorf("initial level = %v, want 0", a.Level())
	}
	a.Set(slog.LevelWarn)
	if a.Level() != slog.LevelWarn {
		t.Errorf("level = %v, want warn", a.Level())
	}
}

// ============================================================
// slogLevelFromString 测试
// ============================================================

func TestSlogLevelFromString(t *testing.T) {
	cases := []struct {
		input string
		want  slog.Level
	}{
		{LevelTypeDebug, slog.LevelDebug},
		{LevelTypeInfo, slog.LevelInfo},
		{LevelTypeWarn, slog.LevelWarn},
		{LevelTypeError, slog.LevelError},
		{"unknown", slog.LevelDebug},
		{"", slog.LevelDebug},
	}
	for _, tc := range cases {
		got := slogLevelFromString(tc.input)
		if got != tc.want {
			t.Errorf("slogLevelFromString(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

// ============================================================
// shortenSource 测试
// ============================================================

func TestShortenSource(t *testing.T) {
	// 少于3级目录，返回完整路径
	t.Run("short path", func(t *testing.T) {
		src := &slog.Source{File: "a/b/file.go", Line: 42}
		result := shortenSource(src)
		if !strings.Contains(result, "42") {
			t.Errorf("result should contain line: %s", result)
		}
	})

	// 超过3级目录，截取最后3级（即文件名+2级父目录）
	t.Run("long path truncated", func(t *testing.T) {
		src := &slog.Source{File: "/root/a/b/c/d/e/file.go", Line: 10}
		result := shortenSource(src)
		// 应该保留最后3个路径段：e/file.go （3段 = 2级目录+文件名）
		expected := "d/e/file.go:10"
		if result != expected {
			t.Errorf("shortenSource = %q, want %q", result, expected)
		}
	})
}

// ============================================================
// SlogContextHandler 测试
// ============================================================

func TestSlogContextHandler(t *testing.T) {
	tw := &testWriter{}
	baseHandler := slog.NewTextHandler(tw, &slog.HandlerOptions{Level: slog.LevelDebug})
	ch := &SlogContextHandler{
		Handler:    baseHandler,
		ContextKey: "traceId",
	}

	logger := slog.New(ch)

	t.Run("context with value", func(t *testing.T) {
		tw.buf.Reset()
		ctx := context.WithValue(context.Background(), "traceId", "abc123")
		logger.InfoContext(ctx, "test message")
		output := tw.buf.String()
		if !strings.Contains(output, "traceId=abc123") {
			t.Errorf("output should contain traceId: %s", output)
		}
	})

	t.Run("context without value", func(t *testing.T) {
		tw.buf.Reset()
		logger.InfoContext(context.Background(), "test message")
		output := tw.buf.String()
		if strings.Contains(output, "traceId") {
			t.Errorf("output should not contain traceId: %s", output)
		}
	})

	t.Run("context with empty value", func(t *testing.T) {
		tw.buf.Reset()
		ctx := context.WithValue(context.Background(), "traceId", "")
		logger.InfoContext(ctx, "test message")
		output := tw.buf.String()
		if strings.Contains(output, "traceId") {
			t.Errorf("output should not contain traceId for empty value: %s", output)
		}
	})

	t.Run("nil context", func(t *testing.T) {
		tw.buf.Reset()
		logger.InfoContext(nil, "test message") //nolint:staticcheck
		// 不应 panic，正常输出日志
		if !strings.Contains(tw.buf.String(), "test message") {
			t.Error("should log message even with nil context")
		}
	})
}

// ============================================================
// SClient 级别/输出控制方法
// ============================================================

func TestSClient_SetLevel(t *testing.T) {
	lc, err := InitSlog(&Config{Level: LevelTypeDebug})
	if err != nil {
		t.Fatal(err)
	}
	defer lc.Sync()
	if lc.Level() != LevelTypeDebug {
		t.Errorf("initial level = %s", lc.Level())
	}
	lc.SetLevel(LevelTypeError)
	if lc.Level() != LevelTypeError {
		t.Errorf("level = %s, want error", lc.Level())
	}
	lc.SetLevel("invalid")
	if lc.Level() != "invalid" {
		t.Errorf("config level = %s, want invalid", lc.Level())
	}
}

func TestSClient_SetOutput(t *testing.T) {
	lc, err := InitSlog(&Config{Level: LevelTypeDebug})
	if err != nil {
		t.Fatal(err)
	}
	defer lc.Sync()
	tw := &testWriter{}
	lc.SetOutput(tw)
	lc.Info("hello")
	if !strings.Contains(tw.buf.String(), "hello") {
		t.Errorf("output should contain 'hello': %s", tw.buf.String())
	}
}

func TestSClient_SetFormat(t *testing.T) {
	lc, err := InitSlog(&Config{Format: FormatConsole})
	if err != nil {
		t.Fatal(err)
	}
	defer lc.Sync()
	tw := &testWriter{}
	lc.SetOutput(tw)

	// 默认 console 输出
	tw.buf.Reset()
	lc.Info("msg1")
	if strings.HasPrefix(strings.TrimSpace(tw.buf.String()), "{") {
		t.Error("console format should not start with {")
	}

	// 切换为 JSON
	lc.SetFormat(FormatJSON)
	tw.buf.Reset()
	lc.Info("msg2")
	if !strings.HasPrefix(strings.TrimSpace(tw.buf.String()), "{") {
		t.Error("json format should start with {")
	}

	// 非法格式回退 JSON
	lc.SetFormat("invalid")
	tw.buf.Reset()
	lc.Info("msg3")
	if !strings.HasPrefix(strings.TrimSpace(tw.buf.String()), "{") {
		t.Error("invalid format should fallback to json")
	}
}

func TestSClient_SetContextKey(t *testing.T) {
	lc, err := InitSlog(&Config{Level: LevelTypeDebug, Format: FormatConsole, ContextKey: "traceId"})
	if err != nil {
		t.Fatal(err)
	}
	defer lc.Sync()
	tw := &testWriter{}
	lc.SetOutput(tw)

	lc.SetContextKey("newTraceKey")
	ctx := context.WithValue(context.Background(), "newTraceKey", "val123")
	lc.InfoContext(ctx, "test")
	if !strings.Contains(tw.buf.String(), "newTraceKey=val123") {
		t.Errorf("output should contain newTraceKey: %s", tw.buf.String())
	}
}

func TestSClient_Handler_SLog_Writer(t *testing.T) {
	lc, err := InitSlog(&Config{Level: LevelTypeDebug})
	if err != nil {
		t.Fatal(err)
	}
	defer lc.Sync()
	if lc.Handler() == nil {
		t.Error("Handler should not be nil")
	}
	if lc.SLog() == nil {
		t.Error("SLog should not be nil")
	}
	if lc.Writer() == nil {
		t.Error("Writer should not be nil")
	}
}

func TestSClient_Sync(t *testing.T) {
	lc, err := InitSlog(&Config{Level: LevelTypeDebug})
	if err != nil {
		t.Fatal(err)
	}
	defer lc.Sync()
	// Sync 应该返回 nil（stdout 不支持 sync，但 testWriter 支持）
	tw := &testWriter{}
	lc.SetOutput(tw)
	if err := lc.Sync(); err != nil {
		t.Errorf("Sync returned error: %v", err)
	}
}

// ============================================================
// SClient 日志方法测试
// ============================================================

func newTestSClient(t *testing.T) (*SClient, *testWriter) {
	t.Helper()
	lc, err := InitSlog(&Config{Level: LevelTypeDebug, Format: FormatConsole})
	if err != nil {
		t.Fatal(err)
	}
	tw := &testWriter{}
	lc.SetOutput(tw)
	return lc, tw
}

func TestSClient_NonFormatMethods(t *testing.T) {
	lc, tw := newTestSClient(t)
	defer lc.Sync()

	lc.Debug("debug msg")
	if !strings.Contains(tw.buf.String(), "debug msg") {
		t.Error("Debug failed")
	}
	tw.buf.Reset()

	lc.Info("info msg")
	if !strings.Contains(tw.buf.String(), "info msg") {
		t.Error("Info failed")
	}
	tw.buf.Reset()

	lc.Warn("warn msg")
	if !strings.Contains(tw.buf.String(), "warn msg") {
		t.Error("Warn failed")
	}
	tw.buf.Reset()

	lc.Error("error msg")
	if !strings.Contains(tw.buf.String(), "error msg") {
		t.Error("Error failed")
	}
}

func TestSClient_FormatMethods(t *testing.T) {
	lc, tw := newTestSClient(t)
	defer lc.Sync()

	lc.Debugf("debug %d", 1)
	if !strings.Contains(tw.buf.String(), "debug 1") {
		t.Error("Debugf failed")
	}
	tw.buf.Reset()

	lc.Infof("info %s", "x")
	if !strings.Contains(tw.buf.String(), "info x") {
		t.Error("Infof failed")
	}
	tw.buf.Reset()

	lc.Warnf("warn %d", 2)
	if !strings.Contains(tw.buf.String(), "warn 2") {
		t.Error("Warnf failed")
	}
	tw.buf.Reset()

	lc.Errorf("error %s", "y")
	if !strings.Contains(tw.buf.String(), "error y") {
		t.Error("Errorf failed")
	}
	tw.buf.Reset()

	lc.Printf("printf %d", 3)
	if !strings.Contains(tw.buf.String(), "printf 3") {
		t.Error("Printf failed")
	}
}

func TestSClient_KVMethods(t *testing.T) {
	lc, tw := newTestSClient(t)
	defer lc.Sync()

	lc.Debugw("debug msg", "key1", "val1")
	if !strings.Contains(tw.buf.String(), "key1=val1") {
		t.Error("Debugw failed")
	}
	tw.buf.Reset()

	lc.Infow("info msg", "key2", "val2")
	if !strings.Contains(tw.buf.String(), "key2=val2") {
		t.Error("Infow failed")
	}
	tw.buf.Reset()

	lc.Warnw("warn msg", "key3", "val3")
	if !strings.Contains(tw.buf.String(), "key3=val3") {
		t.Error("Warnw failed")
	}
	tw.buf.Reset()

	lc.Errorw("error msg", "key4", "val4")
	if !strings.Contains(tw.buf.String(), "key4=val4") {
		t.Error("Errorw failed")
	}
}

func TestSClient_TraceMethods(t *testing.T) {
	lc, tw := newTestSClient(t)
	defer lc.Sync()

	lc.Trace("trace msg")
	if !strings.Contains(tw.buf.String(), "trace msg") {
		t.Error("Trace failed")
	}
	tw.buf.Reset()

	lc.Tracef("trace %d", 42)
	if !strings.Contains(tw.buf.String(), "trace 42") {
		t.Error("Tracef failed")
	}
	tw.buf.Reset()

	lc.Tracew("trace msg", "key", "val")
	if !strings.Contains(tw.buf.String(), "key=val") {
		t.Error("Tracew failed")
	}
}

func TestSClient_ContextMethods(t *testing.T) {
	lc, err := InitSlog(&Config{Level: LevelTypeDebug, Format: FormatConsole, ContextKey: "ctxKey"})
	if err != nil {
		t.Fatal(err)
	}
	defer lc.Sync()
	tw := &testWriter{}
	lc.SetOutput(tw)
	ctx := context.WithValue(context.Background(), "ctxKey", "ctxVal")

	// DebugContext
	lc.DebugContext(ctx, "d")
	if !strings.Contains(tw.buf.String(), "ctxKey=ctxVal") {
		t.Error("DebugContext failed")
	}
	tw.buf.Reset()

	// InfoContext
	lc.InfoContext(ctx, "i")
	if !strings.Contains(tw.buf.String(), "ctxKey=ctxVal") {
		t.Error("InfoContext failed")
	}
	tw.buf.Reset()

	// WarnContext
	lc.WarnContext(ctx, "w")
	if !strings.Contains(tw.buf.String(), "ctxKey=ctxVal") {
		t.Error("WarnContext failed")
	}
	tw.buf.Reset()

	// ErrorContext
	lc.ErrorContext(ctx, "e")
	if !strings.Contains(tw.buf.String(), "ctxKey=ctxVal") {
		t.Error("ErrorContext failed")
	}
	tw.buf.Reset()

	// DebugContextf
	lc.DebugContextf(ctx, "dd %s", "x")
	if !strings.Contains(tw.buf.String(), "dd x") {
		t.Error("DebugContextf failed")
	}
	tw.buf.Reset()

	// InfoContextf
	lc.InfoContextf(ctx, "ii %s", "y")
	if !strings.Contains(tw.buf.String(), "ii y") {
		t.Error("InfoContextf failed")
	}
	tw.buf.Reset()

	// WarnContextf
	lc.WarnContextf(ctx, "ww %s", "z")
	if !strings.Contains(tw.buf.String(), "ww z") {
		t.Error("WarnContextf failed")
	}
	tw.buf.Reset()

	// ErrorContextf
	lc.ErrorContextf(ctx, "ee %s", "a")
	if !strings.Contains(tw.buf.String(), "ee a") {
		t.Error("ErrorContextf failed")
	}
	tw.buf.Reset()

	// DebugContextw
	lc.DebugContextw(ctx, "dw", "k", "v")
	if !strings.Contains(tw.buf.String(), "k=v") {
		t.Error("DebugContextw failed")
	}
	tw.buf.Reset()

	// InfoContextw
	lc.InfoContextw(ctx, "iw", "k", "v")
	if !strings.Contains(tw.buf.String(), "k=v") {
		t.Error("InfoContextw failed")
	}
	tw.buf.Reset()

	// WarnContextw
	lc.WarnContextw(ctx, "ww", "k", "v")
	if !strings.Contains(tw.buf.String(), "k=v") {
		t.Error("WarnContextw failed")
	}
	tw.buf.Reset()

	// ErrorContextw
	lc.ErrorContextw(ctx, "ew", "k", "v")
	if !strings.Contains(tw.buf.String(), "k=v") {
		t.Error("ErrorContextw failed")
	}
}

func TestSClient_Panic_Recover(t *testing.T) {
	lc, err := InitSlog(&Config{Level: LevelTypeDebug, Format: FormatConsole})
	if err != nil {
		t.Fatal(err)
	}
	defer lc.Sync()
	tw := &testWriter{}
	lc.SetOutput(tw)

	t.Run("Panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("Panic should have panicked")
			}
			if !strings.Contains(tw.buf.String(), "panic test") {
				t.Errorf("output should contain 'panic test': %s", tw.buf.String())
			}
		}()
		lc.Panic("panic test")
	})

	tw.buf.Reset()
	t.Run("Panicf", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("Panicf should have panicked")
			}
			if !strings.Contains(tw.buf.String(), "panic 42") {
				t.Errorf("output should contain 'panic 42': %s", tw.buf.String())
			}
		}()
		lc.Panicf("panic %d", 42)
	})

	tw.buf.Reset()
	t.Run("Panicw", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("Panicw should have panicked")
			}
			if !strings.Contains(tw.buf.String(), "panicw msg") {
				t.Errorf("output should contain 'panicw msg': %s", tw.buf.String())
			}
		}()
		lc.Panicw("panicw msg", "key", "val")
	})
}

// ============================================================
// SClient 级别过滤测试（日志被忽略的场景）
// ============================================================

func TestSClient_LevelFiltering(t *testing.T) {
	lc, err := InitSlog(&Config{Level: LevelTypeError, Format: FormatConsole})
	if err != nil {
		t.Fatal(err)
	}
	defer lc.Sync()
	tw := &testWriter{}
	lc.SetOutput(tw)

	// Debug / Info / Warn 级别的消息应该被过滤
	lc.Debug("should not appear")
	lc.Debugf("should not appear %d", 1)
	lc.Debugw("should not appear", "k", "v")
	lc.Trace("should not appear")
	lc.Tracef("should not appear %d", 2)
	lc.Tracew("should not appear", "k", "v")
	lc.Info("should not appear")
	lc.Infof("should not appear %d", 3)
	lc.Infow("should not appear", "k", "v")
	lc.Warn("should not appear")
	lc.Warnf("should not appear %d", 4)
	lc.Warnw("should not appear", "k", "v")

	output := tw.buf.String()
	if output != "" {
		t.Errorf("output should be empty when level is Error, got: %s", output)
	}
}

// ============================================================
// Logger 接口实现验证
// ============================================================

func TestSClient_ImplementsLogger(t *testing.T) {
	var lc Logger = &SClient{}
	_ = lc
}
