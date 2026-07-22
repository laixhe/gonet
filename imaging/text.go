package imaging

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// fontMap 字体缓存
var fontMap = &sync.Map{}

// DrawText 在图片指定位置 (x, y) 绘制文字
func DrawText(img *image.RGBA, face font.Face, fontColor color.Color, text string, x, y int) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(fontColor),
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.I(x), // fixed.I 将整型像素坐标转为 26.6 定点数
			Y: fixed.I(y),
		},
	}
	d.DrawString(text)
}

// TextSplit 按最大像素宽度将文本分割为多行
func TextSplit(text string, face font.Face, maxWidth int) []string {
	var lines []string
	var currentLine []rune
	var currentWidth fixed.Int26_6 // 当前行累计像素宽度

	for _, r := range text {
		advance, hasWidth := face.GlyphAdvance(r)
		newWidth := currentWidth + advance

		// 加入当前字符后超宽，则换行
		if hasWidth && newWidth.Ceil() > maxWidth && len(currentLine) > 0 {
			lines = append(lines, string(currentLine))
			currentLine = []rune{r} // 溢出字符开启新行
			currentWidth = advance
		} else {
			// 宽度未超，或字符无宽度度量（字库缺失），继续累加
			currentLine = append(currentLine, r)
			if hasWidth {
				currentWidth = newWidth
			}
		}
	}
	// 最后一行
	if len(currentLine) > 0 {
		lines = append(lines, string(currentLine))
	}
	return lines
}

// TextWidth 计算文本在指定字体下的像素宽度
func TextWidth(text string, face font.Face) int {
	width := fixed.Int26_6(0) // 26.6 定点数：整数部分 26 位，小数部分 6 位
	for _, r := range text {
		advance, ok := face.GlyphAdvance(r)
		if ok {
			width += advance
		}
		// ok == false 表示字库中无此字形，不计入宽度
	}
	return width.Ceil() // 向上取整为整像素
}

// FontLoad 加载并缓存字体，支持字体文件路径 (string) 或字体数据 ([]byte)
func FontLoad(fontPath any, fontSize float64, fontDPI float64) (font.Face, error) {
	key := ""
	var fontBytes []byte
	var err error
	switch v := fontPath.(type) {
	case string:
		// 文件路径作为缓存 key
		key = fmt.Sprintf("%s_%f_%f", v, fontSize, fontDPI)

		faceAny, ok := fontMap.Load(key)
		if ok {
			return faceAny.(font.Face), nil
		}

		fontBytes, err = os.ReadFile(v)
		if err != nil {
			return nil, err
		}
	case []byte:
		fontBytes = v
		// 字体数据通过 MD5 生成缓存 key
		md5Data := md5.Sum(fontBytes)
		key = fmt.Sprintf("%s_%f_%f", hex.EncodeToString(md5Data[:]), fontSize, fontDPI)

		faceAny, ok := fontMap.Load(key)
		if ok {
			return faceAny.(font.Face), nil
		}
	default:
		return nil, errors.New("font path not supported")
	}

	fontParse, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	face, err := opentype.NewFace(fontParse, &opentype.FaceOptions{
		Size: fontSize,
		DPI:  fontDPI,
	})
	if err != nil {
		return nil, err
	}

	// LoadOrStore 保证并发安全：如果另一个 goroutine 已缓存，使用它的结果
	actual, loaded := fontMap.LoadOrStore(key, face)
	if loaded {
		return actual.(font.Face), nil
	}
	return face, nil
}

// AddText 加载字体并在图片指定位置 (x, y) 绘制文字
func AddText(img *image.RGBA, fontPath any, fontSize float64, fontDPI float64, fontColor color.Color, text string, x, y int) error {
	face, err := FontLoad(fontPath, fontSize, fontDPI)
	if err != nil {
		return err
	}
	DrawText(img, face, fontColor, text, x, y)
	return nil
}

// AddTextToImage 在图片上方或下方添加文字，水平居中，返回含文字的新图片
// 文字超过图片宽度 80% 时自动换行
// position: 文字位置，"top" 或 "bottom"
// gap: 文字与图片之间的间距，单位像素
// margin: 文字外侧留白，单位像素（top 时在上方，bottom 时在下方）
// bgColor: 新图片背景色（可选，默认白色）
func AddTextToImage(img image.Image, position string, fontPath any, fontSize float64, fontDPI float64, fontColor color.Color, text string, gap int, margin int, bgColor ...color.RGBA) (*image.RGBA, error) {
	bg := White
	if len(bgColor) > 0 {
		bg = bgColor[0]
	}

	face, err := FontLoad(fontPath, fontSize, fontDPI)
	if err != nil {
		return nil, err
	}

	imgW := img.Bounds().Dx()
	imgH := img.Bounds().Dy()
	lineH := face.Metrics().Height.Ceil()

	// 文字超过图片宽度 80% 时自动换行
	maxLineW := int(float64(imgW) * 0.8)
	lines := []string{text}
	if TextWidth(text, face) > maxLineW {
		lines = TextSplit(text, face, maxLineW)
	}

	// 计算最长行宽度（用于水平居中）
	maxTextW := 0
	for _, l := range lines {
		w := TextWidth(l, face)
		if w > maxTextW {
			maxTextW = w
		}
	}

	canvasW := imgW
	if maxTextW > canvasW {
		canvasW = maxTextW
	}

	textBlockH := lineH * len(lines)
	canvasH := imgH + gap + textBlockH + margin

	canvas := Create(canvasW, canvasH, bg)
	imgX := (canvasW - imgW) / 2

	switch position {
	case "top":
		// 文字在上，图片在下：上部留白 margin
		for i, l := range lines {
			tx := (canvasW - TextWidth(l, face)) / 2
			ty := margin + face.Metrics().Ascent.Ceil() + i*lineH
			DrawText(canvas, face, fontColor, l, tx, ty)
		}
		Merge(canvas, img, imgX, margin+textBlockH+gap)
	case "bottom":
		// 图片在上，文字在下：下部留白 margin
		Merge(canvas, img, imgX, 0)
		for i, l := range lines {
			tx := (canvasW - TextWidth(l, face)) / 2
			ty := imgH + gap + face.Metrics().Ascent.Ceil() + i*lineH
			DrawText(canvas, face, fontColor, l, tx, ty)
		}
	default:
		return nil, errors.New("position must be \"top\" or \"bottom\"")
	}

	return canvas, nil
}
