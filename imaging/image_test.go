package imaging

import (
	"image"
	"image/color"
	"testing"
)

// ---------------------------------------------------------------------------
// Create — 创建指定尺寸的 RGBA 图像
// ---------------------------------------------------------------------------

// TestCreate_NoBackground 验证不传背景色时创建的是完全透明图像（alpha=0）。
func TestCreate_NoBackground(t *testing.T) {
	dst := Create(10, 20)
	if dst.Bounds().Dx() != 10 || dst.Bounds().Dy() != 20 {
		t.Errorf("size = %v, want 10x20", dst.Bounds())
	}
	c := dst.RGBAAt(0, 0)
	if c.A != 0 {
		t.Errorf("default alpha = %d, want 0 (transparent)", c.A)
	}
}

// TestCreate_WithBackground 验证传入背景色后所有像素都是该颜色。
func TestCreate_WithBackground(t *testing.T) {
	dst := Create(10, 10, Red)
	c := dst.RGBAAt(0, 0)
	if c != Red {
		t.Errorf("pixel = %v, want %v", c, Red)
	}
}

// TestCreate_ZeroSize 验证 0x0 尺寸能正常创建（边界情况）。
func TestCreate_ZeroSize(t *testing.T) {
	dst := Create(0, 0)
	if dst.Bounds().Dx() != 0 || dst.Bounds().Dy() != 0 {
		t.Errorf("size = %v, want 0x0", dst.Bounds())
	}
}

// ---------------------------------------------------------------------------
// Merge — 将小图叠加到大图上
// ---------------------------------------------------------------------------

// TestMerge 验证将 5x5 红色小图合并到 20x20 白色背景的中心位置。
func TestMerge(t *testing.T) {
	dst := Create(20, 20, White)
	src := Create(5, 5, Red)
	Merge(dst, src, 5, 5)
	// 叠加区域中心应为红色
	if dst.RGBAAt(7, 7) != Red {
		t.Errorf("center not red: %v", dst.RGBAAt(7, 7))
	}
	// 未覆盖的角落应为白色
	if dst.RGBAAt(1, 1) != White {
		t.Errorf("corner not white: %v", dst.RGBAAt(1, 1))
	}
}

// TestMerge_NegativeOffset 验证负偏移时自动裁剪：超出目标边界的部分不绘制。
func TestMerge_NegativeOffset(t *testing.T) {
	dst := Create(10, 10, White)
	src := Create(10, 10, Red)
	Merge(dst, src, -5, -5)
	// src 向右下偏移 -5 后，可见区域为 (0,0)-(5,5)
	if dst.RGBAAt(4, 4) != Red {
		t.Errorf("pixel(4,4) = %v, want Red (inside visible area)", dst.RGBAAt(4, 4))
	}
	// (5,5) 在裁剪区域之外，应保持白色
	if dst.RGBAAt(5, 5) != White {
		t.Errorf("pixel(5,5) = %v, want White (outside clipped area)", dst.RGBAAt(5, 5))
	}
}

// TestMerge_BeyondBounds 验证偏移超出目标边界时自动裁剪超出部分。
func TestMerge_BeyondBounds(t *testing.T) {
	dst := Create(10, 10, White)
	src := Create(10, 10, Red)
	Merge(dst, src, 5, 5)
	// 左上角未覆盖区域保持白色
	if dst.RGBAAt(1, 1) != White {
		t.Errorf("pixel(1,1) = %v, want White", dst.RGBAAt(1, 1))
	}
	// 覆盖区域内为红色
	if dst.RGBAAt(6, 6) != Red {
		t.Errorf("pixel(6,6) = %v, want Red", dst.RGBAAt(6, 6))
	}
}

// ---------------------------------------------------------------------------
// Resize — CatmullRom 算法缩放（不保持宽高比）
// ---------------------------------------------------------------------------

// TestResize_Downscale 验证从 20x20 缩小到 10x5（不等比例）。
func TestResize_Downscale(t *testing.T) {
	src := newTestImage(20, 20)
	dst := Resize(src, 10, 5)
	if dst.Bounds().Dx() != 10 || dst.Bounds().Dy() != 5 {
		t.Errorf("size = %v, want 10x5", dst.Bounds())
	}
}

// TestResize_Upscale 验证从 4x4 放大到 20x20。
func TestResize_Upscale(t *testing.T) {
	src := newTestImage(4, 4)
	dst := Resize(src, 20, 20)
	if dst.Bounds().Dx() != 20 || dst.Bounds().Dy() != 20 {
		t.Errorf("size = %v, want 20x20", dst.Bounds())
	}
}

// TestResize_ZeroSize 验证目标尺寸为 0 时返回空图像。
func TestResize_ZeroSize(t *testing.T) {
	src := newTestImage(10, 10)
	dst := Resize(src, 0, 0)
	if dst.Bounds().Dx() != 0 || dst.Bounds().Dy() != 0 {
		t.Errorf("size = %v, want 0x0", dst.Bounds())
	}
}

// TestResize_NegativeSize 验证负数尺寸应返回 0x0 空图像，而非产生畸形 rect。
// 使用表驱动测试覆盖三种情况：宽高均为负、仅宽为负、仅高为负。
func TestResize_NegativeSize(t *testing.T) {
	src := newTestImage(10, 10)
	for _, tc := range []struct {
		name string
		w, h int
	}{
		{"宽高均为负", -1, -1},
		{"仅宽为负", -1, 10},
		{"仅高为负", 10, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dst := Resize(src, tc.w, tc.h)
			if dst.Bounds().Dx() != 0 || dst.Bounds().Dy() != 0 {
				t.Errorf("Resize(%d, %d): bounds=%v, want 0x0", tc.w, tc.h, dst.Bounds())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ResizeNearestNeighbor — 最近邻算法缩放（速度优先）
// ---------------------------------------------------------------------------

// TestResizeNearestNeighbor_Downscale 验证最近邻缩小：20x20 → 10x5。
func TestResizeNearestNeighbor_Downscale(t *testing.T) {
	src := newTestImage(20, 20)
	dst := ResizeNearestNeighbor(src, 10, 5)
	if dst.Bounds().Dx() != 10 || dst.Bounds().Dy() != 5 {
		t.Errorf("size = %v, want 10x5", dst.Bounds())
	}
}

// TestResizeNearestNeighbor_Upscale 验证最近邻放大：4x4 → 20x20。
func TestResizeNearestNeighbor_Upscale(t *testing.T) {
	src := newTestImage(4, 4)
	dst := ResizeNearestNeighbor(src, 20, 20)
	if dst.Bounds().Dx() != 20 || dst.Bounds().Dy() != 20 {
		t.Errorf("size = %v, want 20x20", dst.Bounds())
	}
}

// TestResizeNearestNeighbor_ZeroSize 验证目标尺寸为 0 时返回空图像。
func TestResizeNearestNeighbor_ZeroSize(t *testing.T) {
	src := newTestImage(10, 10)
	dst := ResizeNearestNeighbor(src, 0, 0)
	if dst.Bounds().Dx() != 0 || dst.Bounds().Dy() != 0 {
		t.Errorf("size = %v, want 0x0", dst.Bounds())
	}
}

// TestResizeNearestNeighbor_NegativeSize 验证负数尺寸返回空图像。
func TestResizeNearestNeighbor_NegativeSize(t *testing.T) {
	src := newTestImage(10, 10)
	dst := ResizeNearestNeighbor(src, -1, -1)
	if dst.Bounds().Dx() != 0 || dst.Bounds().Dy() != 0 {
		t.Errorf("size = %v, want 0x0", dst.Bounds())
	}
}

// ---------------------------------------------------------------------------
// ResizeFill — 保持宽高比居中裁剪后缩放
// ---------------------------------------------------------------------------

// TestResizeFill_LandscapeSource 验证横图（100x50）填充到正方形（50x50）时上下被裁剪。
func TestResizeFill_LandscapeSource(t *testing.T) {
	src := newTestImage(100, 50)
	dst := ResizeFill(src, 50, 50)
	if dst.Bounds().Dx() != 50 || dst.Bounds().Dy() != 50 {
		t.Errorf("size = %v, want 50x50", dst.Bounds())
	}
}

// TestResizeFill_PortraitSource 验证竖图（50x100）填充到正方形时左右被裁剪。
func TestResizeFill_PortraitSource(t *testing.T) {
	src := newTestImage(50, 100)
	dst := ResizeFill(src, 50, 50)
	if dst.Bounds().Dx() != 50 || dst.Bounds().Dy() != 50 {
		t.Errorf("size = %v, want 50x50", dst.Bounds())
	}
}

// TestResizeFill_SameRatio 验证宽高比相同时正常缩放，无裁剪。
func TestResizeFill_SameRatio(t *testing.T) {
	src := newTestImage(20, 20)
	dst := ResizeFill(src, 10, 10)
	if dst.Bounds().Dx() != 10 || dst.Bounds().Dy() != 10 {
		t.Errorf("size = %v, want 10x10", dst.Bounds())
	}
}

// TestResizeFill_Zero 验证源图或目标尺寸为 0 时返回空图像，不触发除零。
// 使用表驱动测试覆盖四种零尺寸场景。
func TestResizeFill_Zero(t *testing.T) {
	for _, tc := range []struct {
		name string
		srcW int
		srcH int
		dstW int
		dstH int
	}{
		{"源图全零", 0, 0, 10, 10},
		{"目标全零", 10, 10, 0, 0},
		{"源图宽为零", 0, 10, 10, 10},
		{"源图高为零", 10, 0, 10, 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := image.NewRGBA(image.Rect(0, 0, tc.srcW, tc.srcH))
			dst := ResizeFill(src, tc.dstW, tc.dstH)
			if dst.Bounds().Dx() != tc.dstW {
				t.Errorf("width = %d", dst.Bounds().Dx())
			}
		})
	}
}

// TestResizeFill_ExtremeRatio 验证极端宽高比（2x1→1x100、1x2→100x1）不会因整数除法导致零尺寸裁剪区域。
func TestResizeFill_ExtremeRatio(t *testing.T) {
	// 极端横图：2x1 填充到 1x100
	src := newTestImage(2, 1)
	dst := ResizeFill(src, 1, 100)
	if dst.Bounds().Dx() != 1 || dst.Bounds().Dy() != 100 {
		t.Errorf("size = %v, want 1x100", dst.Bounds())
	}

	// 极端竖图：1x2 填充到 100x1
	src2 := newTestImage(1, 2)
	dst2 := ResizeFill(src2, 100, 1)
	if dst2.Bounds().Dx() != 100 || dst2.Bounds().Dy() != 1 {
		t.Errorf("size = %v, want 100x1", dst2.Bounds())
	}
}

// ---------------------------------------------------------------------------
// Circle — 圆形裁剪
// ---------------------------------------------------------------------------

// TestCircle 验证圆形裁剪：四角透明，中心有内容。
func TestCircle(t *testing.T) {
	src := newTestImage(20, 20)
	dst := Circle(src, 10)
	if dst.Bounds().Dx() != 10 || dst.Bounds().Dy() != 10 {
		t.Errorf("size = %v, want 10x10", dst.Bounds())
	}
	// 四角应在圆外，默认透明
	c := dst.RGBAAt(0, 0)
	if c.A != 0 {
		t.Errorf("corner alpha = %d, want 0 (transparent outside circle)", c.A)
	}
	// 圆心附近应有内容
	center := dst.RGBAAt(5, 5)
	if center.A == 0 {
		t.Error("center is transparent, expected content inside circle")
	}
}

// TestCircle_BGCustom 验证自定义圆形背景色：四角应为指定颜色。
func TestCircle_BGCustom(t *testing.T) {
	src := newTestImage(20, 20)
	dst := Circle(src, 10, Red)
	c := dst.RGBAAt(0, 0)
	if c != Red {
		t.Errorf("corner = %v, want Red (custom background)", c)
	}
}

// ---------------------------------------------------------------------------
// RoundCorners — 圆角裁剪
// ---------------------------------------------------------------------------

// TestRoundCorners_Basic 验证圆角裁剪：四角透明，中心区域保持原色。
func TestRoundCorners_Basic(t *testing.T) {
	src := newTestImage(20, 20)
	dst := RoundCorners(src, 5)

	// 尺寸不变
	if dst.Bounds().Dx() != 20 || dst.Bounds().Dy() != 20 {
		t.Errorf("size = %v, want 20x20", dst.Bounds())
	}

	// 左上角 (0,0) 在圆角外侧，应透明
	c := dst.RGBAAt(0, 0)
	if c.A != 0 {
		t.Errorf("corner (0,0) alpha = %d, want 0 (transparent)", c.A)
	}

	// 中心区域 (10,10) 远离四角，应保持原色
	center := dst.RGBAAt(10, 10)
	if center.A == 0 {
		t.Error("center is transparent, expected content")
	}

	// 顶部中间 (10,0) 不在角区域内，应保持原色
	topMid := dst.RGBAAt(10, 0)
	if topMid.A == 0 {
		t.Error("top-middle is transparent, expected content (not in corner)")
	}
}

// TestRoundCorners_CustomBG 验证自定义圆角背景色：圆角外侧应为指定颜色。
func TestRoundCorners_CustomBG(t *testing.T) {
	src := newTestImage(20, 20)
	dst := RoundCorners(src, 8, Red)

	// 四角应为红色背景
	c := dst.RGBAAt(0, 0)
	if c != Red {
		t.Errorf("corner = %v, want Red (custom background)", c)
	}

	// 中心区域不受影响
	center := dst.RGBAAt(10, 10)
	if center == Red {
		t.Error("center should not be background color")
	}
}

// TestRoundCorners_ZeroRadius 验证 radius=0 时返回原图（无圆角效果）。
func TestRoundCorners_ZeroRadius(t *testing.T) {
	src := newTestImage(10, 10)
	dst := RoundCorners(src, 0)
	if dst.Bounds() != src.Bounds() {
		t.Errorf("bounds = %v, want %v", dst.Bounds(), src.Bounds())
	}
	// 像素应完全一致
	if dst.RGBAAt(0, 0) != src.RGBAAt(0, 0) {
		t.Error("zero radius should return unmodified image")
	}
}

// TestRoundCorners_NegativeRadius 验证负数半径返回原图。
func TestRoundCorners_NegativeRadius(t *testing.T) {
	src := newTestImage(10, 10)
	dst := RoundCorners(src, -5)
	if dst.RGBAAt(0, 0) != src.RGBAAt(0, 0) {
		t.Error("negative radius should return unmodified image")
	}
}

// TestRoundCorners_ZeroSize 验证零尺寸图像不会 panic。
func TestRoundCorners_ZeroSize(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 0, 0))
	dst := RoundCorners(src, 10)
	if dst.Bounds().Dx() != 0 || dst.Bounds().Dy() != 0 {
		t.Errorf("size = %v, want 0x0", dst.Bounds())
	}
}

// TestRoundCorners_MaxRadius 验证 radius 超过图片尺寸一半时自动钳制，不 panic。
func TestRoundCorners_MaxRadius(t *testing.T) {
	src := newTestImage(20, 20)
	// radius=100 远超 20/2=10，应自动钳制为 10
	dst := RoundCorners(src, 100)
	if dst.Bounds().Dx() != 20 || dst.Bounds().Dy() != 20 {
		t.Errorf("size = %v, want 20x20", dst.Bounds())
	}
	// 半径被钳制为 10，四角仍应为透明（圆角覆盖整个半图）
	c := dst.RGBAAt(0, 0)
	if c.A != 0 {
		t.Errorf("corner alpha = %d, want 0 (clamped radius should round corners)", c.A)
	}
}

// TestRoundCorners_NonRGBA 验证对非 RGBA 类型图片的圆角处理。
func TestRoundCorners_NonRGBA(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 12), G: uint8(y * 12), B: 128, A: 255})
		}
	}
	dst := RoundCorners(src, 5)
	if dst.Bounds().Dx() != 20 || dst.Bounds().Dy() != 20 {
		t.Errorf("size = %v, want 20x20", dst.Bounds())
	}
	// 角部应透明
	if dst.RGBAAt(0, 0).A != 0 {
		t.Error("NRGBA corner should be transparent after round corners")
	}
}

// TestRoundCorners_AntiAliasing 验证抗锯齿：圆弧边缘像素应有中间 alpha 值（非 0 非 255）。
func TestRoundCorners_AntiAliasing(t *testing.T) {
	src := newTestImage(20, 20)
	dst := RoundCorners(src, 10)

	// 扫描左上角区域，确认存在抗锯齿像素（alpha 介于 0 和 255 之间）
	foundAA := false
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			a := dst.RGBAAt(x, y).A
			if a > 0 && a < 255 {
				foundAA = true
				break
			}
		}
	}
	if !foundAA {
		t.Error("no anti-aliased pixels found near arc edge; expected smooth blending")
	}

	// 靠近圆弧外部的像素 alpha 应较小（偏背景），靠近内部的应较大（偏源图）
	// 像素 (2,2) 在圆弧外侧，应接近透明
	outerA := dst.RGBAAt(2, 2).A
	if outerA > 128 {
		t.Errorf("pixel (2,2) near outer arc edge: alpha=%d, want <=128 (closer to bg)", outerA)
	}

	// 像素 (3,3) 在圆弧内侧，应接近不透明
	innerA := dst.RGBAAt(3, 3).A
	if innerA < 128 {
		t.Errorf("pixel (3,3) near inner arc edge: alpha=%d, want >=128 (closer to source)", innerA)
	}
}

// ---------------------------------------------------------------------------
// ResizeFit — 等比缩放适配容器（不裁剪，留白填充）
// ---------------------------------------------------------------------------

// TestResizeFit_LandscapeSource 验证横图（100x50）适配到正方形（50x50）时上下留白。
func TestResizeFit_LandscapeSource(t *testing.T) {
	src := newTestImage(100, 50)
	dst := ResizeFit(src, 50, 50)
	if dst.Bounds().Dx() != 50 || dst.Bounds().Dy() != 50 {
		t.Errorf("size = %v, want 50x50", dst.Bounds())
	}
}

// TestResizeFit_PortraitSource 验证竖图（50x100）适配到正方形时左右留白。
func TestResizeFit_PortraitSource(t *testing.T) {
	src := newTestImage(50, 100)
	dst := ResizeFit(src, 50, 50)
	if dst.Bounds().Dx() != 50 || dst.Bounds().Dy() != 50 {
		t.Errorf("size = %v, want 50x50", dst.Bounds())
	}
}

// TestResizeFit_CustomBG 验证竖图适配到横容器时，留白区域为自定义背景色。
func TestResizeFit_CustomBG(t *testing.T) {
	src := newTestImage(10, 20)
	dst := ResizeFit(src, 40, 10, Red)
	// 留白区域应为红色背景
	if dst.RGBAAt(0, 0) != Red {
		t.Errorf("corner = %v, want Red (padding area)", dst.RGBAAt(0, 0))
	}
}

// TestResizeFit_Zero 验证源图为 0x0 时返回指定尺寸的背景图像。
func TestResizeFit_Zero(t *testing.T) {
	src := newTestImage(0, 0)
	dst := ResizeFit(src, 10, 10)
	if dst.Bounds().Dx() != 10 || dst.Bounds().Dy() != 10 {
		t.Errorf("size = %v, want 10x10", dst.Bounds())
	}
}

// TestResizeFit_ZeroDimensions 验证目标宽或高为 0 时仍返回正确尺寸。
func TestResizeFit_ZeroDimensions(t *testing.T) {
	src := newTestImage(10, 10)
	for _, tc := range []struct {
		name string
		dstW int
		dstH int
	}{
		{"目标宽为零", 0, 10},
		{"目标高为零", 10, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dst := ResizeFit(src, tc.dstW, tc.dstH)
			if dst.Bounds().Dx() != tc.dstW || dst.Bounds().Dy() != tc.dstH {
				t.Errorf("size = %v, want %dx%d", dst.Bounds(), tc.dstW, tc.dstH)
			}
		})
	}
}

// TestResizeFit_ExtremeRatio 验证极端宽高比（100x1→1x1、1x100→1x1）不会因整数除法导致崩溃。
func TestResizeFit_ExtremeRatio(t *testing.T) {
	// 极宽图缩放为极小正方形：应返回纯背景色
	src := newTestImage(100, 1)
	dst := ResizeFit(src, 1, 1, Red)
	if dst.Bounds().Dx() != 1 || dst.Bounds().Dy() != 1 {
		t.Errorf("size = %v, want 1x1", dst.Bounds())
	}
	if dst.RGBAAt(0, 0) != Red {
		t.Errorf("pixel = %v, want Red (bg only)", dst.RGBAAt(0, 0))
	}

	// 极高图缩放为极小正方形
	src2 := newTestImage(1, 100)
	dst2 := ResizeFit(src2, 1, 1)
	if dst2.Bounds().Dx() != 1 || dst2.Bounds().Dy() != 1 {
		t.Errorf("size = %v, want 1x1", dst2.Bounds())
	}
}
