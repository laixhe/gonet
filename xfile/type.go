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
	WEBP = "webp"
	BMP  = "bmp"
	GIF  = "gif"

	// 视频

	MP4  = "mp4"
	MOV  = "mov"
	AVI  = "avi"
	WEBM = "webm"

	// 音频

	MP3 = "mp3"
	WAV = "wav"
	OGG = "ogg"
	AAC = "aac"
)

var fileTypeArray = []string{
	PNG, JPG, JPEG, WEBP, BMP, GIF,
	MP4, MOV, AVI, WEBM,
	MP3, WAV, OGG, AAC,
}
var fileTypeMap = map[string]string{
	PNG:  PNG,
	JPG:  JPG,
	JPEG: JPEG,
	WEBP: WEBP,
	BMP:  BMP,
	GIF:  GIF,

	MP4:  MP4,
	MOV:  MOV,
	AVI:  AVI,
	WEBM: WEBM,

	MP3: MP3,
	WAV: WAV,
	OGG: OGG,
	AAC: AAC,
}
var contentTypeMap = map[string]string{
	PNG:  "image/png",
	JPG:  "image/jpeg",
	JPEG: "image/jpeg",
	WEBP: "image/webp",
	BMP:  "image/bmp",
	GIF:  "image/gif",

	MP4:  "video/mp4",
	MOV:  "video/quicktime",
	AVI:  "video/x-msvideo",
	WEBM: "video/webm",

	MP3: "audio/mpeg",
	WAV: "audio/x-wav",
	OGG: "audio/ogg",
	AAC: "audio/aac",
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
