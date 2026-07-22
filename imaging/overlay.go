package imaging

import (
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
)

// Corner 角落位置
type Corner int

const (
	TopLeft     Corner = iota // 左上
	TopRight                  // 右上
	BottomLeft                // 左下
	BottomRight               // 右下
)

// PlaceOverlay 将 overlay 等比缩放后放到背景图指定角落
// corner: 放置位置，可选 TopLeft / TopRight / BottomLeft / BottomRight
// overlayRatio: overlay 宽度占背景宽度的比例，如 0.25 表示 25%
// paddingRatio: 边距占背景宽度的比例，如 0.03 表示 3%
// title 为空时不绘制文字；fontPath/fontRatio/fontColor 仅在 title 非空时有效
// fontPath: 字体文件路径 (string) 或字体数据 ([]byte)
// fontRatio: 文字大小占 overlay 宽度的比例，如 0.08 表示 8%（自适应背景尺寸）
// circular: 是否将 overlay 裁剪为圆形
func PlaceOverlay(background image.Image, overlay image.Image, corner Corner, overlayRatio, paddingRatio float64, title string, fontPath any, fontRatio float64, fontColor color.Color, circular bool) *image.RGBA {
	srcBounds := overlay.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	bgW := background.Bounds().Dx()
	bgH := background.Bounds().Dy()

	overlayWidth := int(float64(bgW) * overlayRatio)
	padding := int(float64(bgW) * paddingRatio)

	// 无有效内容时直接返回背景副本
	if srcW == 0 || srcH == 0 || overlayWidth <= 0 {
		dst := image.NewRGBA(background.Bounds())
		if bg, ok := background.(*image.RGBA); ok {
			copy(dst.Pix, bg.Pix)
		} else {
			draw.Draw(dst, dst.Bounds(), background, image.Point{}, draw.Src)
		}
		return dst
	}

	// 等比缩放 overlay，可选圆形裁剪
	overlayHeight := overlayWidth * srcH / srcW
	resized := Resize(overlay, overlayWidth, overlayHeight)
	if circular {
		diameter := overlayWidth + overlayWidth/10 // 放大 10%，避免圆形裁切内容边缘
		resized = Circle(resized, diameter)
		overlayHeight = diameter
		overlayWidth = diameter // 同步更新宽度，保证后续定位和文字居中正确
	}

	// 加载字体并计算文字高度（title 非空时，字体大小随 overlay 宽度自适应）
	var textFace font.Face
	fontSize := float64(overlayWidth) * fontRatio
	totalHeight := overlayHeight
	if title != "" && fontRatio > 0 {
		face, err := FontLoad(fontPath, fontSize, 72)
		if err == nil {
			textFace = face
			textGap := int(fontSize * 0.3) // 文字与二维码间距随字体缩放
			totalHeight += textFace.Metrics().Height.Ceil() + textGap
		}
	}

	// 根据角落位置计算坐标（totalHeight 包含文字高度）
	var blockX, blockY int
	switch corner {
	case TopLeft:
		blockX, blockY = padding, padding
	case TopRight:
		blockX = bgW - overlayWidth - padding
		blockY = padding
	case BottomLeft:
		blockX = padding
		blockY = bgH - totalHeight - padding
	case BottomRight:
		blockX = bgW - overlayWidth - padding
		blockY = bgH - totalHeight - padding
	}

	// 复制背景并放置覆盖图
	dst := image.NewRGBA(background.Bounds())
	// *image.RGBA 快速路径：直接拷贝像素缓冲区
	if bg, ok := background.(*image.RGBA); ok {
		copy(dst.Pix, bg.Pix)
	} else {
		draw.Draw(dst, dst.Bounds(), background, image.Point{}, draw.Src)
	}
	Merge(dst, resized, blockX, blockY)

	// 绘制标题文字（居中显示在二维码下方）
	if textFace != nil {
		textW := TextWidth(title, textFace)
		textX := blockX + (overlayWidth-textW)/2
		textGap := int(fontSize * 0.3)
		textY := blockY + overlayHeight + textFace.Metrics().Ascent.Ceil() + textGap
		DrawText(dst, textFace, fontColor, title, textX, textY)
	}

	return dst
}
