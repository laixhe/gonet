package imaging

import (
	"image"
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// DecodeBytes / Open — 图片解码与文件读取
// ---------------------------------------------------------------------------

// TestDecodeBytes 验证从 PNG 字节数据解码图片：格式应为 "png"，尺寸应保持。
func TestDecodeBytes(t *testing.T) {
	img := newTestImage(10, 10)
	pngData := encodePNGBytes(img)

	decoded, format, err := DecodeBytes(pngData)
	if err != nil {
		t.Fatalf("DecodeBytes err: %v", err)
	}
	if format != "png" {
		t.Errorf("format = %q, want %q", format, "png")
	}
	if decoded.Bounds().Dx() != 10 || decoded.Bounds().Dy() != 10 {
		t.Errorf("size = %v, want 10x10", decoded.Bounds())
	}
}

// TestDecodeBytes_Invalid 验证解码空数据时应返回错误。
func TestDecodeBytes_Invalid(t *testing.T) {
	_, _, err := DecodeBytes([]byte{})
	if err == nil {
		t.Error("expected error for empty data")
	}
}

// TestOpen 验证从文件路径读取 PNG 图片：格式应为 "png"，尺寸应正确。
func TestOpen(t *testing.T) {
	img := newTestImage(10, 10)
	path := tempFilePath(t, ".png")
	if err := EncodeFilePNG(img, path); err != nil {
		t.Fatalf("setup: %v", err)
	}

	decoded, format, err := Open(path)
	if err != nil {
		t.Fatalf("Open err: %v", err)
	}
	if format != "png" {
		t.Errorf("format = %q, want png", format)
	}
	if decoded.Bounds().Dx() != 10 {
		t.Errorf("width = %d, want 10", decoded.Bounds().Dx())
	}
}

// TestOpen_NotFound 验证读取不存在的文件时应返回错误。
func TestOpen_NotFound(t *testing.T) {
	_, _, err := Open(filepath.Join(t.TempDir(), "nonexistent.png"))
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// ---------------------------------------------------------------------------
// Encode — 图片编码为字节数据
// ---------------------------------------------------------------------------

// TestEncode_PNG 验证将 RGBA 图片编码为 PNG 格式，再解码回来尺寸应一致。
func TestEncode_PNG(t *testing.T) {
	img := newTestImage(10, 10)
	data, err := Encode(img, "png")
	if err != nil {
		t.Fatalf("Encode PNG: %v", err)
	}
	if len(data) == 0 {
		t.Error("empty encoded data")
	}
	// 解码回 PNG 验证数据有效
	decoded, _, err := DecodeBytes(data)
	if err != nil {
		t.Errorf("decode back: %v", err)
	}
	if decoded.Bounds().Dx() != 10 {
		t.Errorf("width = %d", decoded.Bounds().Dx())
	}
}

// TestEncode_JPEG 验证 JPEG 编码不报错且产出非空数据。
func TestEncode_JPEG(t *testing.T) {
	img := newTestImage(10, 10)
	data, err := Encode(img, "jpeg")
	if err != nil {
		t.Fatalf("Encode JPEG: %v", err)
	}
	if len(data) == 0 {
		t.Error("empty encoded data")
	}
}

// TestEncode_JPEG_Quality 验证 JPEG 编码质量参数的有效性。
// 注意：10x10 小图由于 JPEG 文件头开销，不同质量的数据大小差异不可靠，仅验证编码成功。
func TestEncode_JPEG_Quality(t *testing.T) {
	img := newTestImage(10, 10)

	// 质量参数 10 和默认 85 都应编码成功
	for _, q := range []int{10, 85} {
		data, err := Encode(img, "jpg", q)
		if err != nil {
			t.Fatalf("Encode with quality=%d: %v", q, err)
		}
		if len(data) == 0 {
			t.Errorf("empty data for quality=%d", q)
		}
	}

	// 负数质量会被 JPEG 编码器内部钳制，不应报错
	dataNeg, err := Encode(img, "jpg", -1)
	if err != nil {
		t.Fatalf("Encode with -1 quality: %v", err)
	}
	if len(dataNeg) == 0 {
		t.Error("empty data for negative quality")
	}
}

// TestEncode_WEBP 验证 WEBP 编码不报错且产出非空数据。
func TestEncode_WEBP(t *testing.T) {
	img := newTestImage(10, 10)
	data, err := Encode(img, "webp")
	if err != nil {
		t.Fatalf("Encode WEBP: %v", err)
	}
	if len(data) == 0 {
		t.Error("empty encoded data")
	}
}

// TestEncode_Unsupported 验证不支持的格式（如 gif）应返回错误。
func TestEncode_Unsupported(t *testing.T) {
	img := newTestImage(10, 10)
	_, err := Encode(img, "gif")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

// TestEncode_UppercaseFormats 验证 Encode 对大写/混合大小写格式名（如 "PNG"、"Webp"）的处理能力。
func TestEncode_UppercaseFormats(t *testing.T) {
	img := newTestImage(10, 10)
	for _, f := range []string{"PNG", "JPEG", "JPG", "WEBP", "Webp"} {
		t.Run("格式"+f, func(t *testing.T) {
			data, err := Encode(img, f)
			if err != nil {
				t.Errorf("Encode(%q): %v", f, err)
			}
			if len(data) == 0 {
				t.Errorf("Encode(%q): empty data", f)
			}
		})
	}
}

// TestEncode_NilImage 验证传入 nil 图片时应返回错误。
func TestEncode_NilImage(t *testing.T) {
	_, err := Encode(nil, "png")
	if err == nil {
		t.Error("expected error for nil image")
	}
}

// TestEncode_EmptyImage 验证 0x0 空图像在所有格式下均应返回错误。
func TestEncode_EmptyImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 0, 0))
	for _, f := range []string{"png", "jpeg", "webp"} {
		t.Run(f+"格式", func(t *testing.T) {
			_, err := Encode(img, f)
			if err == nil {
				t.Errorf("Encode(0x0, %s): expected error", f)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// EncodeFile / EncodeFilePNG / EncodeFileJPEG / EncodeFileWEBP — 图片写入文件
// ---------------------------------------------------------------------------

// TestEncodeFile 验证 EncodeFile 根据后缀自动选择格式写入 .png/.jpg/.webp 文件。
func TestEncodeFile(t *testing.T) {
	img := newTestImage(10, 10)

	for _, ext := range []string{".png", ".jpg", ".webp"} {
		t.Run(ext+"后缀", func(t *testing.T) {
			path := tempFilePath(t, ext)
			if err := EncodeFile(img, path); err != nil {
				t.Fatalf("EncodeFile %s: %v", ext, err)
			}
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Error("file not created")
			}
		})
	}
}

// TestEncodeFile_InvalidExt 验证不支持的扩展名（如 .bmp）应返回错误。
func TestEncodeFile_InvalidExt(t *testing.T) {
	img := newTestImage(10, 10)
	err := EncodeFile(img, tempFilePath(t, ".bmp"))
	if err == nil {
		t.Error("expected error for unsupported extension")
	}
}

// TestEncodeFile_UppercaseExts 验证大写扩展名（如 .PNG、.WEBP）同样生效。
func TestEncodeFile_UppercaseExts(t *testing.T) {
	img := newTestImage(10, 10)
	for _, ext := range []string{".PNG", ".JPEG", ".JPG", ".WEBP"} {
		t.Run(ext+"大写后缀", func(t *testing.T) {
			path := tempFilePath(t, ext)
			if err := EncodeFile(img, path); err != nil {
				t.Fatalf("EncodeFile %s: %v", ext, err)
			}
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Error("file not created")
			}
		})
	}
}

// TestEncodeFileJPEG_DefaultQuality 验证 EncodeFileJPEG 使用默认质量 85 写入文件并能读回。
func TestEncodeFileJPEG_DefaultQuality(t *testing.T) {
	img := newTestImage(10, 10)
	path := tempFilePath(t, ".jpg")
	if err := EncodeFileJPEG(img, path); err != nil {
		t.Fatalf("EncodeFileJPEG default: %v", err)
	}
	_, _, err := Open(path)
	if err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
}

// TestEncodeFileJPEG_InvalidPath 验证写入不存在的目录路径应返回错误。
func TestEncodeFileJPEG_InvalidPath(t *testing.T) {
	img := newTestImage(10, 10)
	err := EncodeFileJPEG(img, filepath.Join(t.TempDir(), "nonexistent", "test.jpg"))
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

// TestEncodeFileWEBP_InvalidPath 验证 WEBP 写入不存在的目录路径应返回错误。
func TestEncodeFileWEBP_InvalidPath(t *testing.T) {
	img := newTestImage(10, 10)
	err := EncodeFileWEBP(img, filepath.Join(t.TempDir(), "nonexistent", "test.webp"))
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

// TestEncodeFilePNG 验证 EncodeFilePNG 写入文件后能用 Open 读回且尺寸一致。
func TestEncodeFilePNG(t *testing.T) {
	img := newTestImage(10, 10)
	path := tempFilePath(t, ".png")
	if err := EncodeFilePNG(img, path); err != nil {
		t.Fatalf("EncodeFilePNG: %v", err)
	}
	// 写入后立即读回验证（往返测试）
	decoded, _, err := Open(path)
	if err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
	if decoded.Bounds().Dx() != 10 {
		t.Errorf("width = %d", decoded.Bounds().Dx())
	}
}

// TestEncodeFilePNG_InvalidPath 验证 PNG 写入不存在的目录路径应返回错误。
func TestEncodeFilePNG_InvalidPath(t *testing.T) {
	img := newTestImage(10, 10)
	err := EncodeFilePNG(img, filepath.Join(t.TempDir(), "nonexistent", "test.png"))
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

// TestEncodeFileJPEG_CustomQuality 验证 EncodeFileJPEG 自定义质量参数写入后能正确读回。
func TestEncodeFileJPEG_CustomQuality(t *testing.T) {
	img := newTestImage(10, 10)
	path := tempFilePath(t, ".jpg")
	if err := EncodeFileJPEG(img, path, 50); err != nil {
		t.Fatalf("EncodeFileJPEG: %v", err)
	}
	// 验证文件存在且可解码
	decoded, _, err := Open(path)
	if err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
	if decoded.Bounds().Dx() != 10 {
		t.Errorf("width = %d", decoded.Bounds().Dx())
	}
}

// TestEncodeFileWEBP 验证 EncodeFileWEBP 写入后能正确读回。
func TestEncodeFileWEBP(t *testing.T) {
	img := newTestImage(10, 10)
	path := tempFilePath(t, ".webp")
	if err := EncodeFileWEBP(img, path); err != nil {
		t.Fatalf("EncodeFileWEBP: %v", err)
	}
	decoded, _, err := Open(path)
	if err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
	if decoded.Bounds().Dx() != 10 {
		t.Errorf("width = %d", decoded.Bounds().Dx())
	}
}

// ---------------------------------------------------------------------------
// Round-trip — 编解码往返集成测试
// ---------------------------------------------------------------------------

// TestRoundtrip_PNG 验证 PNG 编码→写文件→读文件→解码的完整往返流程。
func TestRoundtrip_PNG(t *testing.T) {
	src := newTestImage(10, 10)
	path := tempFilePath(t, ".png")
	if err := EncodeFilePNG(src, path); err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, format, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if format != "png" {
		t.Errorf("format = %q", format)
	}
	if decoded.Bounds() != src.Bounds() {
		t.Errorf("bounds = %v, want %v", decoded.Bounds(), src.Bounds())
	}
}

// TestRoundtrip_JPEG 验证 JPEG 编码→写文件→读文件的往返流程无错误。
func TestRoundtrip_JPEG(t *testing.T) {
	src := newTestImage(10, 10)
	path := tempFilePath(t, ".jpg")
	if err := EncodeFileJPEG(src, path); err != nil {
		t.Fatalf("encode: %v", err)
	}
	_, _, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
}

// TestRoundtrip_WEBP 验证 WEBP 编码→写文件→读文件的往返流程无错误。
func TestRoundtrip_WEBP(t *testing.T) {
	src := newTestImage(10, 10)
	path := tempFilePath(t, ".webp")
	if err := EncodeFileWEBP(src, path); err != nil {
		t.Fatalf("encode: %v", err)
	}
	_, _, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
}
