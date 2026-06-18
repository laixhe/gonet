package xlog

import (
	"log/slog"
	"testing"
)

func TestInIt(t *testing.T) {
	l := Init()
	slog.SetDefault(slog.New(l.Handler()))

	slog.Info("hello world")
}
