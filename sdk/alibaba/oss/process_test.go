package oss

import (
	"testing"
)

// TestWatermarkResizeProcess 验证水印缩放指令:
// P<=0 不缩放;1~100 正常;越界钳制
func TestWatermarkResizeProcess(t *testing.T) {
	if got := watermarkResizeProcess("wm.png", 0); got != "wm.png" {
		t.Fatalf("expected no resize, got %s", got)
	}
	if got := watermarkResizeProcess("wm.png", -5); got != "wm.png" {
		t.Fatalf("expected no resize for negative, got %s", got)
	}
	if got := watermarkResizeProcess("wm.png", 50); got != "wm.png?x-oss-process=image/resize,P_50" {
		t.Fatalf("unexpected: %s", got)
	}
	if got := watermarkResizeProcess("wm.png", 200); got != "wm.png?x-oss-process=image/resize,P_100" {
		t.Fatalf("expected clamp to 100, got %s", got)
	}
}
