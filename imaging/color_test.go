package imaging

import (
	"image/color"
	"testing"
)

// TestColorConstants 验证所有预定义颜色常量的 RGBA 值是否正确。
// 使用表驱动测试模式：每个子测试名对应常量名，want 是期望的 color.RGBA 值。
func TestColorConstants(t *testing.T) {
	tests := []struct {
		name  string
		color color.RGBA
		want  color.RGBA
	}{
		{"透明(Transparent)", Transparent, color.RGBA{0, 0, 0, 0}},
		{"白色(White)", White, color.RGBA{255, 255, 255, 255}},
		{"黑色(Black)", Black, color.RGBA{0, 0, 0, 255}},
		{"红色(Red)", Red, color.RGBA{255, 0, 0, 255}},
		{"绿色(Green)", Green, color.RGBA{0, 255, 0, 255}},
		{"蓝色(Blue)", Blue, color.RGBA{0, 0, 255, 255}},
		{"黄色(Yellow)", Yellow, color.RGBA{255, 255, 0, 255}},
		{"青色(Cyan)", Cyan, color.RGBA{0, 255, 255, 255}},
		{"洋红(Magenta)", Magenta, color.RGBA{255, 0, 255, 255}},
		{"灰色(Gray)", Gray, color.RGBA{128, 128, 128, 255}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.color != tc.want {
				t.Errorf("got %v, want %v", tc.color, tc.want)
			}
		})
	}
}
