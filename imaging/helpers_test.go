package imaging

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// testFont 是字体相关测试使用的系统字体路径。
// 测试仅在字体文件存在时运行，不存在时自动跳过。
const testFont = `C:\Windows\Fonts\arial.ttf`

// skipIfNoFont 检查系统字体是否存在，不存在则跳过当前测试。
// 应在所有依赖字体的测试函数开头调用。
func skipIfNoFont(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(testFont); os.IsNotExist(err) {
		t.Skip("font file not found, skipping font-dependent test")
	}
}

// newTestImage 创建指定宽高的 RGBA 测试图像。
// 每个像素的颜色根据坐标计算：R = x*25, G = y*25, B = (x+y)*12，
// 生成一个可区分的彩色渐变图案，便于验证像素级变换是否正确。
func newTestImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x * 25), G: uint8(y * 25), B: uint8((x + y) * 12), A: 255})
		}
	}
	return img
}

// encodePNGBytes 将 RGBA 图像编码为 PNG 字节数据。
// 用于生成测试用的二进制图像数据，不写入磁盘。
// 编码失败时直接 panic（仅测试辅助，失败意味着环境异常）。
func encodePNGBytes(img *image.RGBA) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// tempFilePath 在测试临时目录中创建文件路径。
// 临时目录在测试结束后由 Go 测试框架自动清理。
func tempFilePath(t *testing.T, ext string) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "test"+ext)
}
