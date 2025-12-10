package imaging

import (
	"image"
	"image/color"
	"image/draw"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// Transparent 透明
var Transparent = color.RGBA{R: 0, G: 0, B: 0, A: 0}

// White 白色
var White = color.RGBA{R: 255, G: 255, B: 255, A: 255}

// Black 黑色
var Black = color.RGBA{R: 0, G: 0, B: 0, A: 255}

// Red 红色
var Red = color.RGBA{R: 255, G: 0, B: 0, A: 255}

// Green 绿色
var Green = color.RGBA{R: 0, G: 255, B: 0, A: 255}

// Blue 蓝色
var Blue = color.RGBA{R: 0, G: 0, B: 255, A: 255}

// Yellow 黄色
var Yellow = color.RGBA{R: 255, G: 255, B: 0, A: 255}

// Cyan 青色
var Cyan = color.RGBA{R: 0, G: 255, B: 255, A: 255}

// Magenta 品红色（洋红色）
var Magenta = color.RGBA{R: 255, G: 0, B: 255, A: 255}

// Gray 灰色
var Gray = color.RGBA{R: 128, G: 128, B: 128, A: 255}

// Create 创建图像
// 创建一个指定宽度和高度的 RGBA 图像，可选地填充背景颜色
func Create(width, height int, rgba ...color.RGBA) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	// 填充背景
	if len(rgba) > 0 {
		draw.Draw(dst, dst.Bounds(), &image.Uniform{C: rgba[0]}, image.Point{}, draw.Src)
	}
	return dst
}

// Resize 缩放图片到指定尺寸
// 缩放 src 图片到指定的宽度和高度
func Resize(src image.Image, width, height int) *image.RGBA {
	// 创建目标图像
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	// 使用 NearestNeighbor 算法进行缩放
	xdraw.NearestNeighbor.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

// Merge 合并图片
// 合并 src 图片到 dst 图片上指定位置 (x, y)
func Merge(dst draw.Image, src image.Image, x, y int) {
	draw.Draw(dst, dst.Bounds().Add(image.Pt(x, y)), src, src.Bounds().Min, draw.Over)
}

// AddText 添加文字到图片
// 添加文字到 dst 图片上指定位置 (x, y)，使用指定字体、字体大小和字体颜色
func AddText(dst *image.RGBA, text string, x, y int, fontSize float64, fontColor color.RGBA, fontFile []byte) error {
	// 解析字体
	fontFileParse, err := opentype.Parse(fontFile)
	if err != nil {
		return err
	}
	// 创建字体 face
	face, err := opentype.NewFace(fontFileParse, &opentype.FaceOptions{
		Size: fontSize,
		DPI:  72,
	})
	if err != nil {
		return err
	}
	defer face.Close()
	// 创建绘图上下文
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(fontColor),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)},
	}
	// 绘制文字
	d.DrawString(text)
	return nil
}
