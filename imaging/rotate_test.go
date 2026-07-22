package imaging

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

// ---------------------------------------------------------------------------
// Rotate90 / Rotate180 / Rotate270 — 基本旋转
// ---------------------------------------------------------------------------

// TestRotate90 验证顺时针旋转 90°：4x2 → 2x4，且像素映射正确（src(0,0)→dst(1,0)）。
func TestRotate90(t *testing.T) {
	src := newTestImage(4, 2)
	dst := Rotate90(src)
	if dst.Bounds().Dx() != 2 || dst.Bounds().Dy() != 4 {
		t.Errorf("size = %v, want 2x4", dst.Bounds())
	}
	if dst.RGBAAt(1, 0) != src.RGBAAt(0, 0) {
		t.Errorf("pixel mapping incorrect: src(0,0) should map to dst(1,0)")
	}
}

// TestRotate90_NonRGBA 验证对非 RGBA 类型（NRGBA）图片的旋转也能正确执行。
func TestRotate90_NonRGBA(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 60), G: uint8(y * 60), B: 128, A: 255})
		}
	}
	dst := Rotate90(src)
	if dst.Bounds().Dx() != 2 || dst.Bounds().Dy() != 4 {
		t.Errorf("size = %v, want 2x4", dst.Bounds())
	}
}

// TestRotate180 验证旋转 180°：宽高不变，像素映射 src(0,0)→dst(3,2)。
func TestRotate180(t *testing.T) {
	src := newTestImage(4, 3)
	dst := Rotate180(src)
	if dst.Bounds().Dx() != 4 || dst.Bounds().Dy() != 3 {
		t.Errorf("size = %v, want 4x3", dst.Bounds())
	}
	if dst.RGBAAt(3, 2) != src.RGBAAt(0, 0) {
		t.Errorf("pixel mapping incorrect: src(0,0) should map to dst(3,2)")
	}
}

// TestRotate180_NonRGBA 验证对 NRGBA 图片旋转 180° 的正确性。
func TestRotate180_NonRGBA(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 4, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 4; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 60), G: uint8(y * 60), B: 128, A: 255})
		}
	}
	dst := Rotate180(src)
	if dst.Bounds().Dx() != 4 || dst.Bounds().Dy() != 3 {
		t.Errorf("size = %v, want 4x3", dst.Bounds())
	}
}

// TestRotate270 验证顺时针旋转 270°（等效逆时针 90°）：4x2 → 2x4，src(0,0)→dst(0,3)。
func TestRotate270(t *testing.T) {
	src := newTestImage(4, 2)
	dst := Rotate270(src)
	if dst.Bounds().Dx() != 2 || dst.Bounds().Dy() != 4 {
		t.Errorf("size = %v, want 2x4", dst.Bounds())
	}
	if dst.RGBAAt(0, 3) != src.RGBAAt(0, 0) {
		t.Errorf("pixel mapping incorrect: src(0,0) should map to dst(0,3)")
	}
}

// TestRotate270_NonRGBA 验证对 NRGBA 图片旋转 270° 的正确性。
func TestRotate270_NonRGBA(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 60), G: uint8(y * 60), B: 128, A: 255})
		}
	}
	dst := Rotate270(src)
	if dst.Bounds().Dx() != 2 || dst.Bounds().Dy() != 4 {
		t.Errorf("size = %v, want 2x4", dst.Bounds())
	}
}

// TestRotate_1x1 验证三种旋转对 1x1 图像均为无操作（宽高不变）。
// 使用表驱动测试统一验证顺时针90度/旋转180度/顺时针270度。
func TestRotate_1x1(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   func(image.Image) *image.RGBA
	}{
		{"顺时针90度", Rotate90},
		{"旋转180度", Rotate180},
		{"顺时针270度", Rotate270},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := newTestImage(1, 1)
			dst := tc.fn(src)
			if dst.Bounds().Dx() != 1 || dst.Bounds().Dy() != 1 {
				t.Errorf("size = %v, want 1x1", dst.Bounds())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// PlaceOverlay — 将水印/Logo 等比缩放后放到背景图指定角落
// ---------------------------------------------------------------------------

// TestPlaceOverlay_TopLeft 验证 overlay 放置在左上角，5px padding 处。
func TestPlaceOverlay_TopLeft(t *testing.T) {
	bg := Create(100, 100, White)
	ov := Create(20, 20, Red)
	// overlayRatio=0.2 → overlay 宽度 20px；paddingRatio=0.05 → 5px padding
	dst := PlaceOverlay(bg, ov, TopLeft, 0.2, 0.05, "", nil, 0, Black, false)
	if dst.RGBAAt(5, 5) != Red {
		t.Errorf("overlay not at expected position: %v", dst.RGBAAt(5, 5))
	}
	if dst.RGBAAt(0, 0) != White {
		t.Errorf("bg pixel overwritten: %v", dst.RGBAAt(0, 0))
	}
}

// TestPlaceOverlay_TopRight 验证 overlay 放在右上角。
func TestPlaceOverlay_TopRight(t *testing.T) {
	bg := Create(100, 100, White)
	ov := Create(20, 20, Red)
	dst := PlaceOverlay(bg, ov, TopRight, 0.2, 0.05, "", nil, 0, Black, false)
	// overlay 在右侧：x = 100 - 20 - 5 = 75
	if dst.RGBAAt(75, 5) != Red {
		t.Errorf("overlay not at expected position: %v", dst.RGBAAt(75, 5))
	}
}

// TestPlaceOverlay_BottomLeft 验证 overlay 放在左下角。
func TestPlaceOverlay_BottomLeft(t *testing.T) {
	bg := Create(100, 100, White)
	ov := Create(20, 20, Red)
	dst := PlaceOverlay(bg, ov, BottomLeft, 0.2, 0.05, "", nil, 0, Black, false)
	// overlay 在底部：y = 100 - 20 - 5 = 75
	if dst.RGBAAt(5, 75) != Red {
		t.Errorf("overlay not at expected position: %v", dst.RGBAAt(5, 75))
	}
}

// TestPlaceOverlay_BottomRight 验证 overlay 放在右下角。
func TestPlaceOverlay_BottomRight(t *testing.T) {
	bg := Create(100, 100, White)
	ov := Create(20, 20, Red)
	dst := PlaceOverlay(bg, ov, BottomRight, 0.2, 0.05, "", nil, 0, Black, false)
	if dst.RGBAAt(75, 75) != Red {
		t.Errorf("overlay not at expected position: %v", dst.RGBAAt(75, 75))
	}
}

// TestPlaceOverlay_EmptyOverlay 验证零尺寸 overlay 时返回原背景副本（RGBA 快速路径）。
func TestPlaceOverlay_EmptyOverlay(t *testing.T) {
	bg := Create(100, 100, White)
	ov := image.NewRGBA(image.Rect(0, 0, 0, 0))
	dst := PlaceOverlay(bg, ov, TopLeft, 0.2, 0.05, "", nil, 0, Black, false)
	if dst.Bounds() != bg.Bounds() {
		t.Errorf("bounds = %v, want %v", dst.Bounds(), bg.Bounds())
	}
	if dst.RGBAAt(0, 0) != White {
		t.Errorf("bg pixel not preserved: %v", dst.RGBAAt(0, 0))
	}
}

// TestPlaceOverlay_EmptyOverlay_NonRGBA 验证零尺寸 overlay + 非 RGBA 背景时走 draw.Draw 回退路径。
func TestPlaceOverlay_EmptyOverlay_NonRGBA(t *testing.T) {
	bg := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	draw.Draw(bg, bg.Bounds(), &image.Uniform{White}, image.Point{}, draw.Src)
	ov := image.NewRGBA(image.Rect(0, 0, 0, 0))
	dst := PlaceOverlay(bg, ov, TopLeft, 0.2, 0.05, "", nil, 0, Black, false)
	if dst.Bounds() != bg.Bounds() {
		t.Errorf("bounds = %v, want %v", dst.Bounds(), bg.Bounds())
	}
	if dst.RGBAAt(0, 0) != White {
		t.Errorf("bg pixel not preserved: %v", dst.RGBAAt(0, 0))
	}
}

// TestPlaceOverlay_Circular 验证圆形 overlay：圆心区域有内容，远角为背景色。
// overlay 位于 padding=5 处，diameter=22（20 + 20/10），覆盖区域为 (5,5)-(27,27)。
// 圆形遮罩使 overlay 方形区域的远角透明，但边缘像素受 ResizeFill 的 CatmullRom 插值影响，
// 这里通过检查背景远角（远离 overlay）和 overlay 圆心区域来验证圆形功能。
func TestPlaceOverlay_Circular(t *testing.T) {
	bg := Create(100, 100, White)
	ov := Create(20, 20, Red)
	dst := PlaceOverlay(bg, ov, TopLeft, 0.2, 0.05, "", nil, 0, Black, true)

	// 背景远角（远离 overlay 区域）应保持白色
	if dst.RGBAAt(0, 0) != White {
		t.Errorf("bg far corner not preserved: %v", dst.RGBAAt(0, 0))
	}

	// overlay 圆心区域（约 (5+11, 5+11) = (16,16)）应在圆形内部，有可见内容
	center := dst.RGBAAt(16, 16)
	if center.A == 0 {
		t.Error("circular overlay center is transparent, expected content")
	}
	// 且圆心应为红色（非白色背景透出）
	if center != Red {
		t.Errorf("circular overlay center = %v, want Red", center)
	}

	// overlay 底部区域（(5, 5+21) = (5,26)）位于覆盖范围的下边缘，
	// 在圆形遮罩下应为透明，背景白色透出
	bottomEdge := dst.RGBAAt(5, 26)
	if bottomEdge != White {
		t.Errorf("circular overlay bottom edge = %v, want White (transparent circle edge)", bottomEdge)
	}
}

// TestPlaceOverlay_NonRGBABackground 验证非 RGBA 类型背景图也能正常叠加。
func TestPlaceOverlay_NonRGBABackground(t *testing.T) {
	bg := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	draw.Draw(bg, bg.Bounds(), &image.Uniform{White}, image.Point{}, draw.Src)

	ov := Create(20, 20, Red)
	dst := PlaceOverlay(bg, ov, TopLeft, 0.2, 0.0, "", nil, 0, Black, false)
	if dst.Bounds().Dx() != 100 {
		t.Errorf("width = %d, want 100", dst.Bounds().Dx())
	}
}

// TestPlaceOverlay_WithTitle 验证带标题文字的 overlay 放置：尺寸应保持背景大小。
func TestPlaceOverlay_WithTitle(t *testing.T) {
	skipIfNoFont(t)
	bg := Create(100, 100, White)
	ov := Create(20, 20, Red)
	dst := PlaceOverlay(bg, ov, BottomRight, 0.2, 0.05, "Hello", testFont, 0.08, Black, false)
	if dst.Bounds().Dx() != 100 || dst.Bounds().Dy() != 100 {
		t.Errorf("size = %v, want 100x100", dst.Bounds())
	}
}

// TestPlaceOverlay_WithTitle_InvalidFont 验证无效字体路径时 PlaceOverlay 静默跳过文字绘制，不应报错。
func TestPlaceOverlay_WithTitle_InvalidFont(t *testing.T) {
	bg := Create(100, 100, White)
	ov := Create(20, 20, Red)
	dst := PlaceOverlay(bg, ov, TopLeft, 0.2, 0.05, "Title", "nonexistent", 0.08, Black, false)
	if dst.Bounds().Dx() != 100 {
		t.Errorf("width = %d, want 100", dst.Bounds().Dx())
	}
}

// TestPlaceOverlay_TitleFontRatioZero 验证 fontRatio=0 时跳过字体加载，overlay 正常放置。
func TestPlaceOverlay_TitleFontRatioZero(t *testing.T) {
	bg := Create(100, 100, White)
	ov := Create(20, 20, Red)
	dst := PlaceOverlay(bg, ov, TopLeft, 0.2, 0.05, "Title", testFont, 0.0, Black, false)
	if dst.Bounds().Dx() != 100 {
		t.Errorf("width = %d, want 100", dst.Bounds().Dx())
	}
	if dst.RGBAAt(5, 5) != Red {
		t.Errorf("overlay not placed: %v", dst.RGBAAt(5, 5))
	}
}

// TestPlaceOverlay_ZeroOverlayRatio 验证 overlayRatio=0 时 overlay 宽度为 0，返回原背景副本。
func TestPlaceOverlay_ZeroOverlayRatio(t *testing.T) {
	bg := Create(100, 100, White)
	ov := Create(20, 20, Red)
	dst := PlaceOverlay(bg, ov, TopLeft, 0.0, 0.05, "", nil, 0, Black, false)
	if dst.Bounds().Dx() != 100 {
		t.Errorf("width = %d, want 100", dst.Bounds().Dx())
	}
	if dst.RGBAAt(0, 0) != White {
		t.Errorf("bg not preserved: %v", dst.RGBAAt(0, 0))
	}
}
