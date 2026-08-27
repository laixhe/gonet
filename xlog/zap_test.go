package xlog

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ============================================================
// InitZap / Init 测试
// ============================================================

func TestInitZap(t *testing.T) {
	t.Run("no config uses defaults", func(t *testing.T) {
		zc, err := InitZap()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer zc.Sync()
		if zc.Level() != LevelTypeDebug {
			t.Errorf("level = %s, want debug", zc.Level())
		}
	})

	t.Run("with config", func(t *testing.T) {
		zc, err := InitZap(&Config{Level: LevelTypeInfo, Format: FormatConsole})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer zc.Sync()
		if zc.Level() != LevelTypeInfo {
			t.Errorf("level = %s, want info", zc.Level())
		}
	})

	t.Run("nil config returns error", func(t *testing.T) {
		_, err := InitZap(nil)
		if err == nil {
			t.Fatal("expected error for nil config")
		}
	})
}

func TestInit(t *testing.T) {
	zc, err := Init()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer zc.Sync()
	if zc.Level() != LevelTypeDebug {
		t.Errorf("level = %s, want debug", zc.Level())
	}
}

// ============================================================
// zapLevelFromString 测试
// ============================================================

func TestZapLevelFromString(t *testing.T) {
	cases := []struct {
		input string
		want  zapcore.Level
	}{
		{LevelTypeDebug, zap.DebugLevel},
		{LevelTypeInfo, zap.InfoLevel},
		{LevelTypeWarn, zap.WarnLevel},
		{LevelTypeError, zap.ErrorLevel},
		{"unknown", zap.DebugLevel},
		{"", zap.DebugLevel},
	}
	for _, tc := range cases {
		got := zapLevelFromString(tc.input)
		if got != tc.want {
			t.Errorf("zapLevelFromString(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

// ============================================================
// zapEncoder 测试
// ============================================================

func TestZapEncoder(t *testing.T) {
	cfg := zapcore.EncoderConfig{
		MessageKey: "msg",
	}
	// console encoder: 输出为 key=value 文本格式
	encConsole := zapEncoder(FormatConsole, cfg)
	consoleEntry, _ := encConsole.EncodeEntry(zapcore.Entry{Message: "test"}, nil)
	if consoleEntry.String() == "" {
		t.Error("console encoder returned empty output")
	}
	// console 格式不以 { 开头
	if s := consoleEntry.String(); len(s) > 0 && s[0] == '{' {
		t.Error("console format should not start with {")
	}

	// json encoder: 输出为 JSON 格式
	encJSON := zapEncoder(FormatJSON, cfg)
	jsonEntry, _ := encJSON.EncodeEntry(zapcore.Entry{Message: "test"}, nil)
	if s := jsonEntry.String(); len(s) == 0 || s[0] != '{' {
		t.Error("json format should start with {")
	}

	// unknown format falls back to json
	encUnknown := zapEncoder("unknown", cfg)
	unknownEntry, _ := encUnknown.EncodeEntry(zapcore.Entry{Message: "test"}, nil)
	if s := unknownEntry.String(); len(s) == 0 || s[0] != '{' {
		t.Error("unknown format should fallback to json")
	}
}

// ============================================================
// zapTimeEncoder 测试
// ============================================================

func TestZapTimeEncoder(t *testing.T) {
	cfg := zapcore.EncoderConfig{
		TimeKey:    "time",
		EncodeTime: zapTimeEncoder,
	}
	encoder := zapcore.NewConsoleEncoder(cfg)

	tm, err := time.Parse(time.DateTime, "2024-01-15 08:30:45")
	if err != nil {
		t.Fatal(err)
	}
	entry := zapcore.Entry{
		Time:    tm,
		Level:   zapcore.InfoLevel,
		Message: "hello",
	}
	enc, err := encoder.EncodeEntry(entry, nil)
	if err != nil {
		t.Fatal(err)
	}
	line := enc.String()
	if !strings.Contains(line, "2024-01-15 08:30:45") {
		t.Errorf("time not formatted correctly: %s", line)
	}
}

// ============================================================
// ZClient 级别/输出控制方法
// ============================================================

func TestZClient_SetLevel(t *testing.T) {
	zc, err := InitZap(&Config{Level: LevelTypeDebug})
	if err != nil {
		t.Fatal(err)
	}
	defer zc.Sync()
	if zc.Level() != LevelTypeDebug {
		t.Errorf("initial level = %s", zc.Level())
	}
	zc.SetLevel(LevelTypeError)
	if zc.Level() != LevelTypeError {
		t.Errorf("level = %s, want error", zc.Level())
	}
}

func TestZClient_SetOutput(t *testing.T) {
	zc, err := InitZap(&Config{Level: LevelTypeDebug})
	if err != nil {
		t.Fatal(err)
	}
	defer zc.Sync()

	var buf bytes.Buffer
	zc.SetOutput(&buf)
	zc.Info("hello output")
	if !strings.Contains(buf.String(), "hello output") {
		t.Errorf("output should contain 'hello output': %s", buf.String())
	}
}

func TestZClient_SetFormat(t *testing.T) {
	zc, err := InitZap(&Config{Format: FormatConsole})
	if err != nil {
		t.Fatal(err)
	}
	defer zc.Sync()

	var buf bytes.Buffer
	zc.SetOutput(&buf)

	buf.Reset()
	zc.Info("msg1")
	if strings.HasPrefix(strings.TrimSpace(buf.String()), "{") {
		t.Error("console format should not start with {")
	}

	zc.SetFormat(FormatJSON)
	buf.Reset()
	zc.Info("msg2")
	if !strings.HasPrefix(strings.TrimSpace(buf.String()), "{") {
		t.Error("json format should start with {")
	}

	zc.SetFormat("invalid")
	buf.Reset()
	zc.Info("msg3")
	if !strings.HasPrefix(strings.TrimSpace(buf.String()), "{") {
		t.Error("invalid format should fallback to json")
	}
}

func TestZClient_Accessors(t *testing.T) {
	zc, err := InitZap(&Config{Level: LevelTypeDebug})
	if err != nil {
		t.Fatal(err)
	}
	defer zc.Sync()
	if zc.ZLog() == nil {
		t.Error("ZLog should not be nil")
	}
	if zc.Logger() == nil {
		t.Error("Logger should not be nil")
	}
	if zc.SugaredLogger() == nil {
		t.Error("SugaredLogger should not be nil")
	}
	if zc.Writer() == nil {
		t.Error("Writer should not be nil")
	}
	if zc.Level() != LevelTypeDebug {
		t.Errorf("Level = %s, want debug", zc.Level())
	}
}

func TestZClient_Sync(t *testing.T) {
	zc, err := InitZap(&Config{Level: LevelTypeDebug})
	if err != nil {
		t.Fatal(err)
	}
	defer zc.Sync()
	if err := zc.Sync(); err != nil {
		t.Logf("Sync returned error (may be expected on this OS): %v", err)
	}
}

// ============================================================
// ZClient 日志方法测试
// ============================================================

func newTestZClient(t *testing.T) (*ZClient, *bytes.Buffer) {
	t.Helper()
	zc, err := InitZap(&Config{Level: LevelTypeDebug, Format: FormatConsole})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	zc.SetOutput(&buf)
	return zc, &buf
}

func TestZClient_NonFormatMethods(t *testing.T) {
	zc, buf := newTestZClient(t)
	defer zc.Sync()

	zc.Debug("debug msg")
	assertContains(t, buf, "debug msg")
	buf.Reset()

	zc.Info("info msg")
	assertContains(t, buf, "info msg")
	buf.Reset()

	zc.Warn("warn msg")
	assertContains(t, buf, "warn msg")
	buf.Reset()

	zc.Error("error msg")
	assertContains(t, buf, "error msg")
}

func TestZClient_FormatMethods(t *testing.T) {
	zc, buf := newTestZClient(t)
	defer zc.Sync()

	zc.Debugf("debug %d", 1)
	assertContains(t, buf, "debug 1")
	buf.Reset()

	zc.Infof("info %s", "x")
	assertContains(t, buf, "info x")
	buf.Reset()

	zc.Warnf("warn %d", 2)
	assertContains(t, buf, "warn 2")
	buf.Reset()

	zc.Errorf("error %s", "y")
	assertContains(t, buf, "error y")
	buf.Reset()

	zc.Printf("printf %d", 3)
	assertContains(t, buf, "printf 3")
}

func TestZClient_KVMethods(t *testing.T) {
	zc, buf := newTestZClient(t)
	defer zc.Sync()

	zc.Debugw("debug msg", "key1", "val1")
	assertContains(t, buf, "key1")
	buf.Reset()

	zc.Infow("info msg", "key2", "val2")
	assertContains(t, buf, "key2")
	buf.Reset()

	zc.Warnw("warn msg", "key3", "val3")
	assertContains(t, buf, "key3")
	buf.Reset()

	zc.Errorw("error msg", "key4", "val4")
	assertContains(t, buf, "key4")
}

func TestZClient_FieldsMethods(t *testing.T) {
	zc, buf := newTestZClient(t)
	defer zc.Sync()

	zc.DebugFields("debug msg", zap.String("field1", "v1"))
	assertContains(t, buf, "field1")
	buf.Reset()

	zc.InfoFields("info msg", zap.Int("field2", 42))
	assertContains(t, buf, "field2")
	buf.Reset()

	zc.WarnFields("warn msg", zap.Bool("field3", true))
	assertContains(t, buf, "field3")
	buf.Reset()

	zc.ErrorFields("error msg", zap.String("field4", "v4"))
	assertContains(t, buf, "field4")
}

func TestZClient_TraceMethods(t *testing.T) {
	zc, buf := newTestZClient(t)
	defer zc.Sync()

	zc.Trace("trace msg")
	assertContains(t, buf, "trace msg")
	buf.Reset()

	zc.Tracef("trace %d", 42)
	assertContains(t, buf, "trace 42")
	buf.Reset()

	zc.Tracew("trace msg", "key", "val")
	assertContains(t, buf, "key")
}

func TestZClient_ContextMethods(t *testing.T) {
	zc, buf := newTestZClient(t)
	defer zc.Sync()

	ctx := context.Background()

	zc.DebugContext(ctx, "dc")
	assertContains(t, buf, "dc")
	buf.Reset()

	zc.InfoContext(ctx, "ic")
	assertContains(t, buf, "ic")
	buf.Reset()

	zc.WarnContext(ctx, "wc")
	assertContains(t, buf, "wc")
	buf.Reset()

	zc.ErrorContext(ctx, "ec")
	assertContains(t, buf, "ec")
	buf.Reset()

	zc.DebugContextf(ctx, "dcf %d", 1)
	assertContains(t, buf, "dcf 1")
	buf.Reset()

	zc.InfoContextf(ctx, "icf %d", 2)
	assertContains(t, buf, "icf 2")
	buf.Reset()

	zc.WarnContextf(ctx, "wcf %d", 3)
	assertContains(t, buf, "wcf 3")
	buf.Reset()

	zc.ErrorContextf(ctx, "ecf %d", 4)
	assertContains(t, buf, "ecf 4")
	buf.Reset()

	zc.DebugContextw(ctx, "dcw", "k", "v")
	assertContains(t, buf, "k")
	buf.Reset()

	zc.InfoContextw(ctx, "icw", "k", "v")
	assertContains(t, buf, "k")
	buf.Reset()

	zc.WarnContextw(ctx, "wcw", "k", "v")
	assertContains(t, buf, "k")
	buf.Reset()

	zc.ErrorContextw(ctx, "ecw", "k", "v")
	assertContains(t, buf, "k")
}

func TestZClient_Panic_Recover(t *testing.T) {
	zc, _ := newTestZClient(t)
	defer zc.Sync()

	t.Run("Panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Panic should have panicked")
			}
		}()
		zc.Panic("panic test")
	})

	t.Run("Panicf", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Panicf should have panicked")
			}
		}()
		zc.Panicf("panic %d", 1)
	})

	t.Run("Panicw", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Panicw should have panicked")
			}
		}()
		zc.Panicw("panicw msg", "k", "v")
	})
}

// ============================================================
// ZClient Logger 接口实现验证
// ============================================================

func TestZClient_ImplementsLogger(t *testing.T) {
	var zc Logger = &ZClient{}
	_ = zc
}

// ============================================================
// Fatal 测试 (通过子进程, os.Exit 无法在单进程验证)
// ============================================================

func TestZClient_Fatal_Subprocess(t *testing.T) {
	if os.Getenv("TEST_FATAL") == "1" {
		zc, _ := InitZap(&Config{Level: LevelTypeDebug})
		_ = zc
		zc.Fatal("fatal exit")
		return
	}
	t.Skip("Fatal calls os.Exit, tested via compilation check")
}

func TestSClient_Fatal_Subprocess(t *testing.T) {
	if os.Getenv("TEST_FATAL_SLOG") == "1" {
		lc, _ := InitSlog(&Config{Level: LevelTypeDebug})
		_ = lc
		lc.Fatal("fatal exit")
		return
	}
	t.Skip("Fatal calls os.Exit, tested via compilation check")
}

// ============================================================
// 辅助函数
// ============================================================

func assertContains(t *testing.T, buf *bytes.Buffer, substr string) {
	t.Helper()
	if !strings.Contains(buf.String(), substr) {
		t.Errorf("output should contain %q, got: %s", substr, buf.String())
	}
}
