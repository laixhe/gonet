package imaging

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"sync"

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

// 加载字体
var fontMap = &sync.Map{}

// DecodeBytes 解析图片
func DecodeBytes(data []byte) (img image.Image, format string, err error) {
	return image.Decode(bytes.NewBuffer(data))
}

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
func Resize(img image.Image, width, height int) *image.RGBA {
	// 创建目标图像
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	// 使用 NearestNeighbor 算法进行缩放
	xdraw.NearestNeighbor.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

// Merge 合并图片
// 合并 src 图片到 dst 图片上指定位置 (x, y)
func Merge(dst draw.Image, src image.Image, x, y int) {
	draw.Draw(dst, dst.Bounds().Add(image.Pt(x, y)), src, src.Bounds().Min, draw.Over)
}

// DrawText 绘制文字到图片
func DrawText(img *image.RGBA, face font.Face, fontColor color.Color, text string, x, y int) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(fontColor),
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.I(x),
			Y: fixed.I(y),
		},
	}
	d.DrawString(text)
}

// TextSplit 分割文本为多行
func TextSplit(text string, face font.Face, maxWidth int) []string {
	var lines []string
	var currentLine string

	textRune := []rune(text)
	for _, v := range textRune {
		currentLine += string(v)
		width := TextWidth(currentLine, face)
		if width > maxWidth && currentLine != "" {
			lines = append(lines, currentLine)
			currentLine = ""
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}

// TextWidth 计算文本宽度
func TextWidth(text string, face font.Face) int {
	width := fixed.Int26_6(0)
	for _, r := range []rune(text) {
		advance, ok := face.GlyphAdvance(r)
		if ok {
			width += advance
		}
	}
	return width.Ceil()
}

// FontLoad 加载字体
func FontLoad(fontPath any, fontSize float64, fontDPI float64) (font.Face, error) {
	key := ""
	var fontBytes []byte
	var err error
	// 判断字体类型
	switch fontPath.(type) {
	case string:
		fontPathString := fontPath.(string)
		key = fmt.Sprintf("%s_%f_%f", fontPathString, fontSize, fontDPI)
		//
		faceAny, ok := fontMap.Load(key)
		if ok {
			return faceAny.(font.Face), nil
		}
		//
		fontBytes, err = os.ReadFile(fontPathString)
		if err != nil {
			return nil, err
		}
	case []byte:
		fontBytes = fontPath.([]byte)
		md5Data := md5.Sum(fontBytes)
		key = fmt.Sprintf("%s_%f_%f", hex.EncodeToString(md5Data[:]), fontSize, fontDPI)
		//
		faceAny, ok := fontMap.Load(key)
		if ok {
			return faceAny.(font.Face), nil
		}
	default:
		return nil, errors.New("font path not supported")
	}
	// 解析字体
	fontParse, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	// 创建字体
	face, err := opentype.NewFace(fontParse, &opentype.FaceOptions{
		Size: fontSize,
		DPI:  fontDPI,
	})
	fontMap.Store(key, face)
	return face, nil
}

// AddText 添加文字到图片
// 添加文字到图片上指定位置 (x, y)，使用指定字体、字体大小和字体颜色
func AddText(img *image.RGBA, fontPath any, fontSize float64, fontDPI float64, fontColor color.Color, text string, x, y int) error {
	// 加载字体
	face, err := FontLoad(fontPath, fontSize, fontDPI)
	if err != nil {
		return err
	}
	// 绘制文字到图片
	DrawText(img, face, fontColor, text, x, y)
	return nil
}
