package xlog

import (
	"testing"
)

func TestConfig_Check(t *testing.T) {
	t.Run("nil config return error", func(t *testing.T) {
		var c *Config
		err := c.Check()
		if err == nil {
			t.Fatal("expected error for nil config")
		}
	})

	t.Run("empty config fills defaults", func(t *testing.T) {
		c := &Config{}
		err := c.Check()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Run != RunTypeConsole {
			t.Errorf("default Run = %s, want %s", c.Run, RunTypeConsole)
		}
		if c.Level != LevelTypeDebug {
			t.Errorf("default Level = %s, want %s", c.Level, LevelTypeDebug)
		}
		if c.Format != FormatJSON {
			t.Errorf("default Format = %s, want %s", c.Format, FormatJSON)
		}
		if c.MaxSize != 3 {
			t.Errorf("default MaxSize = %d, want 3", c.MaxSize)
		}
		if c.MaxBackups != 3 {
			t.Errorf("default MaxBackups = %d, want 3", c.MaxBackups)
		}
		if c.MaxAge != 3 {
			t.Errorf("default MaxAge = %d, want 3", c.MaxAge)
		}
	})

	t.Run("file mode without path fills default path", func(t *testing.T) {
		c := &Config{Run: RunTypeFile}
		err := c.Check()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Path != "logs.log" {
			t.Errorf("default Path = %s, want logs.log", c.Path)
		}
	})

	t.Run("file mode with path keeps path", func(t *testing.T) {
		c := &Config{Run: RunTypeFile, Path: "/var/log/app.log"}
		err := c.Check()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Path != "/var/log/app.log" {
			t.Errorf("Path = %s, want /var/log/app.log", c.Path)
		}
	})

	t.Run("invalid Run defaults to console", func(t *testing.T) {
		c := &Config{Run: "invalid"}
		err := c.Check()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Run != RunTypeConsole {
			t.Errorf("Run = %s, want %s", c.Run, RunTypeConsole)
		}
	})

	t.Run("invalid Level defaults to debug", func(t *testing.T) {
		c := &Config{Level: "trace"}
		err := c.Check()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Level != LevelTypeDebug {
			t.Errorf("Level = %s, want %s", c.Level, LevelTypeDebug)
		}
	})

	t.Run("invalid Format defaults to json", func(t *testing.T) {
		c := &Config{Format: "text"}
		err := c.Check()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Format != FormatJSON {
			t.Errorf("Format = %s, want %s", c.Format, FormatJSON)
		}
	})

	t.Run("all levels accepted", func(t *testing.T) {
		for _, level := range []string{LevelTypeDebug, LevelTypeInfo, LevelTypeWarn, LevelTypeError} {
			c := &Config{Level: level}
			err := c.Check()
			if err != nil {
				t.Errorf("unexpected error for level %s: %v", level, err)
			}
			if c.Level != level {
				t.Errorf("Level = %s, want %s", c.Level, level)
			}
		}
	})

	t.Run("all formats accepted", func(t *testing.T) {
		for _, format := range []string{FormatJSON, FormatConsole} {
			c := &Config{Format: format}
			err := c.Check()
			if err != nil {
				t.Errorf("unexpected error for format %s: %v", format, err)
			}
			if c.Format != format {
				t.Errorf("Format = %s, want %s", c.Format, format)
			}
		}
	})

	t.Run("custom MaxSize/MaxBackups/MaxAge kept", func(t *testing.T) {
		c := &Config{MaxSize: 100, MaxBackups: 50, MaxAge: 30}
		err := c.Check()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.MaxSize != 100 {
			t.Errorf("MaxSize = %d, want 100", c.MaxSize)
		}
		if c.MaxBackups != 50 {
			t.Errorf("MaxBackups = %d, want 50", c.MaxBackups)
		}
		if c.MaxAge != 30 {
			t.Errorf("MaxAge = %d, want 30", c.MaxAge)
		}
	})
}

func TestValidConstants(t *testing.T) {
	t.Run("RunTypeConsole in validRuns", func(t *testing.T) {
		if !validRuns[RunTypeConsole] {
			t.Error("RunTypeConsole should be valid")
		}
	})
	t.Run("RunTypeFile in validRuns", func(t *testing.T) {
		if !validRuns[RunTypeFile] {
			t.Error("RunTypeFile should be valid")
		}
	})
	t.Run("FormatJSON in validFormats", func(t *testing.T) {
		if !validFormats[FormatJSON] {
			t.Error("FormatJSON should be valid")
		}
	})
	t.Run("FormatConsole in validFormats", func(t *testing.T) {
		if !validFormats[FormatConsole] {
			t.Error("FormatConsole should be valid")
		}
	})
	t.Run("all levels in validLevels", func(t *testing.T) {
		for _, level := range []string{LevelTypeDebug, LevelTypeInfo, LevelTypeWarn, LevelTypeError} {
			if !validLevels[level] {
				t.Errorf("%s should be valid", level)
			}
		}
	})
}
