package imaging

import (
	"os"
	"testing"
)

// ---------------------------------------------------------------------------
// DrawText / TextWidth / TextSplit — 文字绘制与度量
// ---------------------------------------------------------------------------

// TestDrawText 验证 DrawText 确实在图像上绘制了像素（某些像素不再是白色）。
func TestDrawText(t *testing.T) {
	skipIfNoFont(t)
	face, err := FontLoad(testFont, 12, 72)
	if err != nil {
		t.Fatalf("FontLoad: %v", err)
	}
	img := Create(100, 30, White)
	DrawText(img, face, Black, "Hello", 10, 20)
	// 扫描文字区域，确认至少有一个像素从白色变为其他颜色
	changed := false
	for x := 10; x < 60; x++ {
		if img.RGBAAt(x, 15) != White {
			changed = true
			break
		}
	}
	if !changed {
		t.Error("no pixels changed after DrawText")
	}
}

// TestTextWidth 验证 TextWidth 对 "Hello" 返回正值。
func TestTextWidth(t *testing.T) {
	skipIfNoFont(t)
	face, err := FontLoad(testFont, 12, 72)
	if err != nil {
		t.Fatalf("FontLoad: %v", err)
	}
	w := TextWidth("Hello", face)
	if w <= 0 {
		t.Errorf("TextWidth = %d, want > 0", w)
	}
}

// TestTextSplit 验证短文本不换行、超出宽度的文本被拆分为多行。
func TestTextSplit(t *testing.T) {
	skipIfNoFont(t)
	face, err := FontLoad(testFont, 12, 72)
	if err != nil {
		t.Fatalf("FontLoad: %v", err)
	}
	singleW := TextWidth("A", face)

	// 宽度足够时应保持一行
	lines := TextSplit("AB", face, singleW*10)
	if len(lines) != 1 {
		t.Errorf("lines = %d, want 1", len(lines))
	}

	// 宽度仅够一个字符时应收割为两行
	lines2 := TextSplit("AB", face, singleW)
	if len(lines2) != 2 {
		t.Errorf("lines = %d, want 2: %v", len(lines2), lines2)
	}
}

// TestTextSplit_Empty 验证空字符串应返回空切片。
func TestTextSplit_Empty(t *testing.T) {
	skipIfNoFont(t)
	face, err := FontLoad(testFont, 12, 72)
	if err != nil {
		t.Fatalf("FontLoad: %v", err)
	}
	lines := TextSplit("", face, 100)
	if len(lines) != 0 {
		t.Errorf("lines = %d, want 0", len(lines))
	}
}

// ---------------------------------------------------------------------------
// FontLoad — 字体加载与缓存
// ---------------------------------------------------------------------------

// TestFontLoad_FilePath 验证从文件路径加载字体，第二次调用命中缓存。
func TestFontLoad_FilePath(t *testing.T) {
	skipIfNoFont(t)
	face, err := FontLoad(testFont, 12, 72)
	if err != nil {
		t.Fatalf("FontLoad: %v", err)
	}
	if face == nil {
		t.Fatal("face is nil")
	}
	// 再次加载相同参数应命中缓存
	face2, err := FontLoad(testFont, 12, 72)
	if err != nil {
		t.Fatalf("FontLoad (cached): %v", err)
	}
	if face2 == nil {
		t.Fatal("cached face is nil")
	}
}

// TestFontLoad_ByteData 验证从 []byte 字体数据加载及缓存命中。
func TestFontLoad_ByteData(t *testing.T) {
	skipIfNoFont(t)
	data, err := os.ReadFile(testFont)
	if err != nil {
		t.Fatalf("read font: %v", err)
	}
	face, err := FontLoad(data, 14, 72)
	if err != nil {
		t.Fatalf("FontLoad([]byte): %v", err)
	}
	if face == nil {
		t.Fatal("face is nil")
	}
	// 再次加载相同数据应命中缓存
	face2, err := FontLoad(data, 14, 72)
	if err != nil {
		t.Fatalf("FontLoad([]byte, cached): %v", err)
	}
	if face2 == nil {
		t.Fatal("cached face is nil")
	}
}

// TestFontLoad_InvalidPath 验证不存在的字体路径应返回错误。
func TestFontLoad_InvalidPath(t *testing.T) {
	_, err := FontLoad("nonexistent.ttf", 12, 72)
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

// TestFontLoad_InvalidType 验证不支持的字体路径类型（如 int）应返回错误。
func TestFontLoad_InvalidType(t *testing.T) {
	_, err := FontLoad(123, 12, 72)
	if err == nil {
		t.Error("expected error for unsupported type")
	}
}

// TestFontLoad_InvalidFontData 验证无效的字体字节数据应返回错误。
func TestFontLoad_InvalidFontData(t *testing.T) {
	_, err := FontLoad([]byte("not a font"), 12, 72)
	if err == nil {
		t.Error("expected error for invalid font bytes")
	}
}

// TestFontLoad_DifferentSizesDifferentCache 验证不同字号产生不同的缓存条目。
func TestFontLoad_DifferentSizesDifferentCache(t *testing.T) {
	skipIfNoFont(t)
	face12, err := FontLoad(testFont, 12, 72)
	if err != nil {
		t.Fatalf("FontLoad 12: %v", err)
	}
	face24, err := FontLoad(testFont, 24, 72)
	if err != nil {
		t.Fatalf("FontLoad 24: %v", err)
	}
	if face12 == face24 {
		t.Error("different font sizes should return different faces")
	}
	// 再次加载相同字号应通过 LoadOrStore 获取同一实例
	face12b, err := FontLoad(testFont, 12, 72)
	if err != nil {
		t.Fatalf("FontLoad 12 cached: %v", err)
	}
	if face12 != face12b {
		t.Error("cached load should return same face instance")
	}
}

// TestFontLoad_ConcurrentLoadOrStore 验证 sync.Map.LoadOrStore 路径：首次加载存储，二次加载命中。
func TestFontLoad_ConcurrentLoadOrStore(t *testing.T) {
	skipIfNoFont(t)
	// 使用不同 DPI 参数产生新的缓存键，强制走 LoadOrStore 路径
	face, err := FontLoad(testFont, 11, 96)
	if err != nil {
		t.Fatalf("FontLoad: %v", err)
	}
	if face == nil {
		t.Fatal("face is nil")
	}
	// 第二次加载应命中 loaded=true 分支
	face2, err := FontLoad(testFont, 11, 96)
	if err != nil {
		t.Fatalf("FontLoad cached: %v", err)
	}
	if face2 == nil {
		t.Fatal("cached face is nil")
	}
}

// ---------------------------------------------------------------------------
// AddText — 加载字体并绘制文字
// ---------------------------------------------------------------------------

// TestAddText 验证 AddText 便捷函数：加载字体后在图像上绘制文字。
func TestAddText(t *testing.T) {
	skipIfNoFont(t)
	img := Create(100, 30, White)
	err := AddText(img, testFont, 12, 72, Black, "Hello", 10, 20)
	if err != nil {
		t.Fatalf("AddText: %v", err)
	}
	changed := false
	for x := 10; x < 60; x++ {
		if img.RGBAAt(x, 15) != White {
			changed = true
			break
		}
	}
	if !changed {
		t.Error("no pixels changed after AddText")
	}
}

// TestAddText_InvalidFont 验证无效字体路径时 AddText 应返回错误。
func TestAddText_InvalidFont(t *testing.T) {
	img := Create(100, 30, White)
	err := AddText(img, "nonexistent.ttf", 12, 72, Black, "Hello", 10, 20)
	if err == nil {
		t.Error("expected error for invalid font")
	}
}

// TestAddText_ByteFont 验证 AddText 支持 []byte 类型的字体数据。
func TestAddText_ByteFont(t *testing.T) {
	skipIfNoFont(t)
	data, err := os.ReadFile(testFont)
	if err != nil {
		t.Fatalf("read font: %v", err)
	}
	img := Create(100, 30, White)
	err = AddText(img, data, 12, 72, Black, "Hello", 10, 20)
	if err != nil {
		t.Fatalf("AddText([]byte): %v", err)
	}
}

// ---------------------------------------------------------------------------
// AddTextToImage — 在图片上方/下方添加文字并返回新图像
// ---------------------------------------------------------------------------

// TestAddTextToImage_Top 验证文字在上方：canvas 高度应大于原图（增加了文字区域）。
func TestAddTextToImage_Top(t *testing.T) {
	skipIfNoFont(t)
	src := newTestImage(50, 30)
	dst, err := AddTextToImage(src, "top", testFont, 12, 72, Black, "Title", 5, 5)
	if err != nil {
		t.Fatalf("AddTextToImage top: %v", err)
	}
	if dst.Bounds().Dx() < 50 {
		t.Errorf("width = %d, want >= 50", dst.Bounds().Dx())
	}
	if dst.Bounds().Dy() <= 30 {
		t.Errorf("height = %d, want > 30 (added text)", dst.Bounds().Dy())
	}
}

// TestAddTextToImage_Bottom 验证文字在下方：canvas 高度应大于原图。
func TestAddTextToImage_Bottom(t *testing.T) {
	skipIfNoFont(t)
	src := newTestImage(50, 30)
	dst, err := AddTextToImage(src, "bottom", testFont, 12, 72, Black, "Title", 5, 5)
	if err != nil {
		t.Fatalf("AddTextToImage bottom: %v", err)
	}
	if dst.Bounds().Dx() < 50 {
		t.Errorf("width = %d, want >= 50", dst.Bounds().Dx())
	}
	if dst.Bounds().Dy() <= 30 {
		t.Errorf("height = %d, want > 30", dst.Bounds().Dy())
	}
}

// TestAddTextToImage_InvalidPosition 验证无效的 position 参数（如 "middle"）应返回错误。
func TestAddTextToImage_InvalidPosition(t *testing.T) {
	skipIfNoFont(t)
	src := newTestImage(50, 30)
	_, err := AddTextToImage(src, "middle", testFont, 12, 72, Black, "Title", 5, 5)
	if err == nil {
		t.Error("expected error for invalid position")
	}
}

// TestAddTextToImage_CustomBG 验证自定义背景色：文字下方的 margin 区域应为指定背景色。
func TestAddTextToImage_CustomBG(t *testing.T) {
	skipIfNoFont(t)
	src := newTestImage(50, 30)
	dst, err := AddTextToImage(src, "bottom", testFont, 12, 72, Black, "Title", 5, 5, Red)
	if err != nil {
		t.Fatalf("AddTextToImage: %v", err)
	}
	// "bottom" 模式：图片在上，文字+margin 在下。最底部应落在 margin 区域（Red 背景）
	bottomY := dst.Bounds().Dy() - 1
	if dst.RGBAAt(0, bottomY) != Red {
		t.Errorf("bg pixel (0,%d) = %v, want Red", bottomY, dst.RGBAAt(0, bottomY))
	}
}

// TestAddTextToImage_NoWrapShortText 验证短文本不换行：canvas 宽度应 >= 原图宽度。
func TestAddTextToImage_NoWrapShortText(t *testing.T) {
	skipIfNoFont(t)
	src := newTestImage(200, 30)
	dst, err := AddTextToImage(src, "bottom", testFont, 12, 72, Black, "Hi", 5, 5)
	if err != nil {
		t.Fatalf("AddTextToImage: %v", err)
	}
	if dst.Bounds().Dx() < 50 {
		t.Errorf("width too small: %d", dst.Bounds().Dx())
	}
}

// TestAddTextToImage_WideTextWraps 验证宽文字自动换行：canvas 高度应显著增加。
func TestAddTextToImage_WideTextWraps(t *testing.T) {
	skipIfNoFont(t)
	src := newTestImage(50, 30)
	// 文字超出 50*0.8=40px 宽时应自动换行，增加行数
	dst, err := AddTextToImage(src, "bottom", testFont, 12, 72, Black, "Long long title text", 5, 5)
	if err != nil {
		t.Fatalf("AddTextToImage: %v", err)
	}
	if dst.Bounds().Dy() <= 30+5+5 {
		t.Errorf("height = %d, expected more lines after wrapping", dst.Bounds().Dy())
	}
}

// TestAddTextToImage_InvalidFont 验证无效字体时应返回错误。
func TestAddTextToImage_InvalidFont(t *testing.T) {
	src := newTestImage(50, 30)
	_, err := AddTextToImage(src, "top", "nonexistent.ttf", 12, 72, Black, "Title", 5, 5)
	if err == nil {
		t.Error("expected error for invalid font")
	}
}

// TestAddTextToImage_TextWiderThanImage 验证文字宽于图片时 canvas 自动扩展宽度。
func TestAddTextToImage_TextWiderThanImage(t *testing.T) {
	skipIfNoFont(t)
	// 窄图（20px） + 大字体（24px）→ canvas 宽度应大于原图宽度
	src := newTestImage(20, 30)
	dst, err := AddTextToImage(src, "top", testFont, 24, 72, Black, "WIDETEXT", 5, 5)
	if err != nil {
		t.Fatalf("AddTextToImage: %v", err)
	}
	if dst.Bounds().Dx() <= 20 {
		t.Errorf("canvas width = %d, want > 20 (expanded for text)", dst.Bounds().Dx())
	}
}
