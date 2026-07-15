package xlog

import (
	"log/slog"
	"testing"
)

func TestInIt(t *testing.T) {
	l := Init()
	slog.SetDefault(l.Logger())

	slog.Info("hello world...")
	l.Logger().Info("hello world")
}
