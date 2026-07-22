package imaging

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	xdraw "golang.org/x/image/draw"
)

// Create 创建指定尺寸的 RGBA 图像，可选背景色填充
func Create(width, height int, rgba ...color.RGBA) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	if len(rgba) > 0 {
		draw.Draw(dst, dst.Bounds(), &image.Uniform{C: rgba[0]}, image.Point{}, draw.Src)
	}
	return dst
}

// Merge 将 src 图片绘制到 dst 图片的指定位置 (x, y)
func Merge(dst draw.Image, src image.Image, x, y int) {
	// 将 src 的边界整体偏移 (x,y)，draw.Draw 内部裁剪会自动处理越界
	r := src.Bounds().Add(image.Pt(x, y))
	draw.Draw(dst, r, src, image.Point{}, draw.Over)
}

// Resize 使用 CatmullRom 算法缩放图片到指定尺寸（不保持宽高比）
func Resize(img image.Image, width, height int) *image.RGBA {
	if width <= 0 || height <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 0, 0))
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), xdraw.Over, nil)
	return dst
}

// ResizeNearestNeighbor 使用最近邻算法缩放图片到指定尺寸（速度快但质量较低）（不保持宽高比）
func ResizeNearestNeighbor(img image.Image, width, height int) *image.RGBA {
	if width <= 0 || height <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 0, 0))
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	xdraw.NearestNeighbor.Scale(dst, dst.Bounds(), img, img.Bounds(), xdraw.Over, nil)
	return dst
}

// ResizeFill 保持宽高比，从原图中心裁剪后缩放到目标尺寸（超出部分居中裁剪）
func ResizeFill(img image.Image, width, height int) *image.RGBA {
	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	// 零尺寸直接返回空图像，避免除零
	if srcW == 0 || srcH == 0 || width == 0 || height == 0 {
		return image.NewRGBA(image.Rect(0, 0, width, height))
	}

	var cropW, cropH, cropX, cropY int

	// 用交叉相乘比较宽高比，避免浮点精度问题
	// srcW/srcH > width/height 等价于 srcW*height > width*srcH
	if srcW*height > width*srcH {
		// 原图更宽：以高度为基准，裁剪左右两侧
		cropH = srcH
		cropW = width * srcH / height
		cropX = (srcW - cropW) / 2 // 居中偏移
	} else {
		// 原图更高：以宽度为基准，裁剪上下两侧
		cropW = srcW
		cropH = height * srcW / width
		cropY = (srcH - cropH) / 2 // 居中偏移
	}

	// 极端宽高比下整数除法可能得到零尺寸，此时无法有效裁剪
	if cropW <= 0 || cropH <= 0 {
		return image.NewRGBA(image.Rect(0, 0, width, height))
	}

	// 直接从原图的裁剪子区域缩放到目标尺寸，省去中间图像
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	srcRect := image.Rect(cropX, cropY, cropX+cropW, cropY+cropH)
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, srcRect, xdraw.Over, nil)
	return dst
}

// blendRGBA 按比例 t 将背景色 bg 混合到源色 src 上，写入 dst 的 (x, y)
// t=0 保留源色不变，t=1 完全替换为背景色，中间值线性过渡
// +0.5 用于四舍五入到最近整数，避免 uint8 直接截断产生的偏差
func blendRGBA(dst *image.RGBA, x, y int, src, bg color.RGBA, t float64) {
	dst.SetRGBA(x, y, color.RGBA{
		R: uint8(float64(src.R)*(1-t) + float64(bg.R)*t + 0.5),
		G: uint8(float64(src.G)*(1-t) + float64(bg.G)*t + 0.5),
		B: uint8(float64(src.B)*(1-t) + float64(bg.B)*t + 0.5),
		A: uint8(float64(src.A)*(1-t) + float64(bg.A)*t + 0.5),
	})
}

// Circle 将图片居中裁剪后转为圆形输出，返回直径×直径的正方形图像，圆外区域为背景色（默认透明）
// 抗锯齿：圆弧边界 ±0.5 像素过渡带，线性混合源色与背景色消除锯齿
//
// 抗锯齿原理：
//   - 以像素中心 (x+0.5, y+0.5) 到圆心的距离 d 为判断依据
//   - r1 = radius-0.5：内边界，d≤r1 为完全在圆内（保留源）
//   - r2 = radius+0.5：外边界，d>r2 为完全在圆外（背景色）
//   - r1<d≤r2：过渡带，按距离比例线性混合源色与背景色
func Circle(img image.Image, diameter int, bgColor ...color.RGBA) *image.RGBA {
	// 先居中裁剪为正方形，保证内容主体在圆内
	square := ResizeFill(img, diameter, diameter)

	bg := Transparent
	if len(bgColor) > 0 {
		bg = bgColor[0]
	}

	dst := image.NewRGBA(image.Rect(0, 0, diameter, diameter))
	radius := float64(diameter) / 2
	r1 := radius - 0.5 // 内边界，d ≤ r1 时完全在圆内
	r1Sq := r1 * r1
	r2 := radius + 0.5 // 外边界，d > r2 时完全在圆外
	r2Sq := r2 * r2

	for y := 0; y < diameter; y++ {
		dy := float64(y) + 0.5 - radius
		dy2 := dy * dy // 提到外层，避免内层每像素重复计算 dy²
		for x := 0; x < diameter; x++ {
			dx := float64(x) + 0.5 - radius
			d2 := dx*dx + dy2 // 像素中心到圆心距离的平方

			if d2 <= r1Sq {
				// 完全在圆内 → 从正方形源图复制像素
				dst.SetRGBA(x, y, square.RGBAAt(x, y))
			} else if d2 > r2Sq {
				// 完全在圆外 → 填充背景色
				dst.SetRGBA(x, y, bg)
			} else {
				// 抗锯齿过渡带：dist∈[-0.5, 0.5] → t∈[0, 1]
				dist := math.Sqrt(d2) - radius
				t := dist + 0.5 // dist=-0.5→t=0(源), dist=0.5→t=1(背景)
				blendRGBA(dst, x, y, square.RGBAAt(x, y), bg, t)
			}
		}
	}
	return dst
}

// RoundCorners 将图片四角裁剪为圆角，保持原图尺寸不变，圆角外侧填充背景色（默认透明）
// radius 为圆角半径（像素）；若超过宽/高的一半则自动钳制；≤0 时返回原图副本
//
// 性能：仅处理四个角区域（O(radius²)），对大图友好；RGBA 输入走 Pix 快速复制路径
// 抗锯齿：圆弧边缘 ±0.5 像素过渡带，通过线性混合源色与背景色消除锯齿
func RoundCorners(img image.Image, radius int, bgColor ...color.RGBA) *image.RGBA {
	srcBounds := img.Bounds()
	w := srcBounds.Dx()
	h := srcBounds.Dy()

	// 零尺寸或非正半径，返回原图副本（不做圆角处理）
	if w == 0 || h == 0 || radius <= 0 {
		dst := image.NewRGBA(srcBounds)
		if src, ok := img.(*image.RGBA); ok {
			copy(dst.Pix, src.Pix)
		} else {
			draw.Draw(dst, srcBounds, img, image.Point{}, draw.Src)
		}
		return dst
	}

	bg := Transparent
	if len(bgColor) > 0 {
		bg = bgColor[0]
	}

	// 限制半径不超过宽/高的一半（避免角区域重叠）
	maxR := w / 2
	if h/2 < maxR {
		maxR = h / 2
	}
	if radius > maxR {
		radius = maxR
	}

	// 先将源图整体复制到目标：RGBA 走 Pix 直接拷贝，其他类型走 draw.Draw
	dst := image.NewRGBA(srcBounds)
	if src, ok := img.(*image.RGBA); ok {
		copy(dst.Pix, src.Pix)
	} else {
		draw.Draw(dst, srcBounds, img, image.Point{}, draw.Src)
	}

	// 抗锯齿参数（与 Circle 相同算法）：
	//   每个角的圆弧圆心在 (角顶点向内偏移 r)，以半径 r 画弧裁剪
	//   过渡带为 r±0.5 像素，共 1 像素宽
	r := float64(radius)
	r1 := r - 0.5 // 内边界
	r1Sq := r1 * r1
	r2 := r + 0.5 // 外边界
	r2Sq := r2 * r2
	wr := float64(w) - r // 右角圆心 x
	hr := float64(h) - r // 下角圆心 y

	// 以下四个循环分别处理左上、右上、左下、右下角区域
	// 每个区域仅遍历 radius×radius 的像素块，而非全图

	// 左上角：像素 (0..r-1, 0..r-1)，圆心 (r, r)
	for y := 0; y < radius; y++ {
		fy := float64(y) + 0.5 - r
		fy2 := fy * fy
		for x := 0; x < radius; x++ {
			fx := float64(x) + 0.5 - r
			d2 := fx*fx + fy2

			if d2 > r2Sq {
				// 完全在圆弧外 → 背景色
				dst.SetRGBA(x, y, bg)
			} else if d2 > r1Sq {
				// 抗锯齿过渡带
				dist := math.Sqrt(d2) - r
				t := dist + 0.5
				blendRGBA(dst, x, y, dst.RGBAAt(x, y), bg, t)
			}
			// d2 <= r1Sq: 完全在圆弧内 → 保留原像素（已在 copy 中设置）
		}
	}
	// 右上角：像素 (w-r..w-1, 0..r-1)，圆心 (w-r, r)
	for y := 0; y < radius; y++ {
		fy := float64(y) + 0.5 - r
		fy2 := fy * fy
		for x := w - radius; x < w; x++ {
			fx := float64(x) + 0.5 - wr
			d2 := fx*fx + fy2

			if d2 > r2Sq {
				dst.SetRGBA(x, y, bg)
			} else if d2 > r1Sq {
				dist := math.Sqrt(d2) - r
				t := dist + 0.5
				blendRGBA(dst, x, y, dst.RGBAAt(x, y), bg, t)
			}
		}
	}
	// 左下角：像素 (0..r-1, h-r..h-1)，圆心 (r, h-r)
	for y := h - radius; y < h; y++ {
		fy := float64(y) + 0.5 - hr
		fy2 := fy * fy
		for x := 0; x < radius; x++ {
			fx := float64(x) + 0.5 - r
			d2 := fx*fx + fy2

			if d2 > r2Sq {
				dst.SetRGBA(x, y, bg)
			} else if d2 > r1Sq {
				dist := math.Sqrt(d2) - r
				t := dist + 0.5
				blendRGBA(dst, x, y, dst.RGBAAt(x, y), bg, t)
			}
		}
	}
	// 右下角：像素 (w-r..w-1, h-r..h-1)，圆心 (w-r, h-r)
	for y := h - radius; y < h; y++ {
		fy := float64(y) + 0.5 - hr
		fy2 := fy * fy
		for x := w - radius; x < w; x++ {
			fx := float64(x) + 0.5 - wr
			d2 := fx*fx + fy2

			if d2 > r2Sq {
				dst.SetRGBA(x, y, bg)
			} else if d2 > r1Sq {
				dist := math.Sqrt(d2) - r
				t := dist + 0.5
				blendRGBA(dst, x, y, dst.RGBAAt(x, y), bg, t)
			}
		}
	}
	return dst
}

// ResizeFit 等比缩放图片到容器内，不裁剪，剩余区域填充背景色（默认透明）
func ResizeFit(img image.Image, width, height int, bgColor ...color.RGBA) *image.RGBA {
	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	if srcW == 0 || srcH == 0 || width == 0 || height == 0 {
		return image.NewRGBA(image.Rect(0, 0, width, height))
	}

	bg := Transparent
	if len(bgColor) > 0 {
		bg = bgColor[0]
	}

	// 交叉相乘判断以哪边为基准缩放（与 ResizeFill 相反：适配不裁剪）
	var scaledW, scaledH int
	if srcW*height > width*srcH {
		// 原图更宽 → 以目标宽度为基准，上下留白
		scaledW = width
		scaledH = srcH * width / srcW
	} else {
		// 原图更高 → 以目标高度为基准，左右留白
		scaledH = height
		scaledW = srcW * height / srcH
	}

	// 极端宽高比下整数除法可能得到零尺寸，此时无法有效缩放
	if scaledW <= 0 || scaledH <= 0 {
		return Create(width, height, bg)
	}

	// 创建背景，将缩放后的图片居中放置
	dst := Create(width, height, bg)
	scaled := Resize(img, scaledW, scaledH)
	x := (width - scaledW) / 2
	y := (height - scaledH) / 2
	Merge(dst, scaled, x, y)
	return dst
}
