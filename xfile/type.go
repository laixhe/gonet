package xfile

import (
	"strings"
)

// 类型
const (
	// 图片

	PNG  = "png"
	JPG  = "jpg"
	JPEG = "jpeg"
	GIF  = "gif"
	BMP  = "bmp"
	WEBP = "webp"

	// 视频

	MP4  = "mp4"
	MOV  = "mov"
	AVI  = "avi"
	FLV  = "flv"
	WMV  = "wmv"
	MKV  = "mkv"
	WEBM = "webm"

	// 音频

	MP3 = "mp3"
	WAV = "wav"
	OGG = "ogg"
	AAC = "aac"
	M4A = "m4a"
)

var fileTypeArray = []string{
	PNG, JPG, JPEG, GIF, BMP, WEBP,
	MP4, MOV, AVI, FLV, WMV, MKV, WEBM,
	MP3, WAV, OGG, AAC, M4A,
}
var fileTypeMap = map[string]string{
	PNG:  PNG,
	JPG:  JPG,
	JPEG: JPEG,
	GIF:  GIF,
	BMP:  BMP,
	WEBP: WEBP,

	MP4:  MP4,
	MOV:  MOV,
	AVI:  AVI,
	FLV:  FLV,
	WMV:  WMV,
	MKV:  MKV,
	WEBM: WEBM,

	MP3: MP3,
	WAV: WAV,
	OGG: OGG,
	AAC: AAC,
	M4A: M4A,
}
var contentTypeMap = map[string]string{
	PNG:  "image/png",
	JPG:  "image/jpeg",
	JPEG: "image/jpeg",
	GIF:  "image/gif",
	BMP:  "image/bmp",
	WEBP: "image/webp",

	MP4:  "video/mp4",
	MOV:  "video/quicktime",
	AVI:  "video/x-msvideo",
	FLV:  "video/x-flv",
	WMV:  "video/x-ms-wmv",
	MKV:  "video/x-matroska",
	WEBM: "video/webm",

	MP3: "audio/mpeg",
	WAV: "audio/x-wav",
	OGG: "audio/ogg",
	AAC: "audio/aac",
	M4A: "audio/mp4",
}

// GetTypes 所有类型
func GetTypes() []string {
	return fileTypeArray
}

// IsType 判断是否有这个类型
func IsType(str string) bool {
	if str == "" {
		return false
	}
	str = strings.ToLower(str)
	_, ok := fileTypeMap[str]
	return ok
}

// GetContentType 获取文件类型
func GetContentType(str string) string {
	if str == "" {
		return ""
	}
	str = strings.ToLower(str)
	return contentTypeMap[str]
}
