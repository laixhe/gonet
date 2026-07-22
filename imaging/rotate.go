package imaging

import "image"

// Rotate90 顺时针旋转 90°，宽高互换
func Rotate90(img image.Image) *image.RGBA {
	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	// 旋转后宽高互换：原 (x, y) → 新 (srcH-1-y, x)
	dst := image.NewRGBA(image.Rect(0, 0, srcH, srcW))

	// *image.RGBA 快速路径：直接访问像素数组，避免接口装箱开销
	if src, ok := img.(*image.RGBA); ok {
		for y := 0; y < srcH; y++ {
			for x := 0; x < srcW; x++ {
				dst.SetRGBA(srcH-1-y, x, src.RGBAAt(x, y))
			}
		}
		return dst
	}

	for y := 0; y < srcH; y++ {
		for x := 0; x < srcW; x++ {
			dst.Set(srcH-1-y, x, img.At(x, y))
		}
	}
	return dst
}

// Rotate180 旋转 180°，宽高不变
func Rotate180(img image.Image) *image.RGBA {
	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	// 原 (x, y) → 新 (srcW-1-x, srcH-1-y)
	dst := image.NewRGBA(image.Rect(0, 0, srcW, srcH))

	// *image.RGBA 快速路径：直接访问像素数组，避免接口装箱开销
	if src, ok := img.(*image.RGBA); ok {
		for y := 0; y < srcH; y++ {
			for x := 0; x < srcW; x++ {
				dst.SetRGBA(srcW-1-x, srcH-1-y, src.RGBAAt(x, y))
			}
		}
		return dst
	}

	for y := 0; y < srcH; y++ {
		for x := 0; x < srcW; x++ {
			dst.Set(srcW-1-x, srcH-1-y, img.At(x, y))
		}
	}
	return dst
}

// Rotate270 顺时针旋转 270°（等效逆时针 90°），宽高互换
func Rotate270(img image.Image) *image.RGBA {
	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	// 旋转后宽高互换：原 (x, y) → 新 (y, srcW-1-x)
	dst := image.NewRGBA(image.Rect(0, 0, srcH, srcW))

	// *image.RGBA 快速路径：直接访问像素数组，避免接口装箱开销
	if src, ok := img.(*image.RGBA); ok {
		for y := 0; y < srcH; y++ {
			for x := 0; x < srcW; x++ {
				dst.SetRGBA(y, srcW-1-x, src.RGBAAt(x, y))
			}
		}
		return dst
	}

	for y := 0; y < srcH; y++ {
		for x := 0; x < srcW; x++ {
			dst.Set(y, srcW-1-x, img.At(x, y))
		}
	}
	return dst
}
