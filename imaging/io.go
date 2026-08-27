package imaging

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/HugoSmits86/nativewebp"
	_ "golang.org/x/image/webp" // 注册 WEBP 解码器，使 image.Decode 支持 WEBP
)

// DecodeBytes 从字节切片解码图片，返回图片对象、格式和错误
func DecodeBytes(data []byte) (img image.Image, format string, err error) {
	return image.Decode(bytes.NewReader(data))
}

// Open 从文件路径加载图片，返回图片对象、格式和错误
func Open(filePath string) (image.Image, string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", err
	}
	return DecodeBytes(data)
}

// Encode 将 RGBA 图片编码为 PNG/JPEG/WEBP 字节数据，jpegQuality 仅对 JPEG 有效（默认 85）
func Encode(img *image.RGBA, format string, jpegQuality ...int) ([]byte, error) {
	if img == nil || img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
		return nil, errors.New("invalid image: nil or zero size")
	}
	var buf bytes.Buffer
	switch strings.ToLower(format) {
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
	case "jpeg", "jpg":
		q := 85
		if len(jpegQuality) > 0 {
			q = jpegQuality[0]
		}
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: q}); err != nil {
			return nil, err
		}
	case "webp":
		if err := nativewebp.Encode(&buf, img, &nativewebp.Options{
			CompressionLevel: nativewebp.DefaultCompression,
		}); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported format: " + format + " (use png, jpeg or webp)")
	}
	return buf.Bytes(), nil
}

// EncodeFile 将 RGBA 图片编码并写入文件，根据后缀自动选择格式
func EncodeFile(img *image.RGBA, filePath string, jpegQuality ...int) error {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".png":
		return EncodeFilePNG(img, filePath)
	case ".jpeg", ".jpg":
		return EncodeFileJPEG(img, filePath, jpegQuality...)
	case ".webp":
		return EncodeFileWEBP(img, filePath)
	default:
		return errors.New("unsupported file extension: " + ext + " (use .png, .jpg or .webp)")
	}
}

// EncodeFilePNG 将 RGBA 图片编码并写入 PNG 文件
func EncodeFilePNG(img *image.RGBA, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// EncodeFileJPEG 将 RGBA 图片编码并写入 JPEG 文件，jpegQuality 默认 85
func EncodeFileJPEG(img *image.RGBA, filePath string, jpegQuality ...int) error {
	q := 85
	if len(jpegQuality) > 0 {
		q = jpegQuality[0]
	}
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: q})
}

// EncodeFileWEBP 将 RGBA 图片编码并写入 WEBP 文件
func EncodeFileWEBP(img *image.RGBA, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return nativewebp.Encode(f, img, &nativewebp.Options{
		CompressionLevel: nativewebp.DefaultCompression,
	})
}
