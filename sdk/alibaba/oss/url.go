package oss

import (
	"context"
	"mime"
	"path/filepath"
	"strings"
	"time"

	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/rs/xid"
)

// FilePreSignatureURL 生成预签名文件上传url
type FilePreSignatureURL struct {
	// 文件地址
	Url string
	// 上传url
	SignUrl string
	// 类型
	ContentType string
}

// SetPreSignatureURL 生成预签名文件上传url
func (oc *OClient) SetPreSignatureURL(ctx context.Context, fileDir string, fileNames []string, isNotInternal ...bool) ([]FilePreSignatureURL, error) {
	list := make([]FilePreSignatureURL, 0, len(fileNames))
	for _, fileName := range fileNames {
		ext := filepath.Ext(fileName)
		mimeType := contentTypeByExt(ext)
		// if mimeType == "" {
		// 	mimeType = "application/octet-stream"
		// }
		name := xid.New().String()
		dst := fileDir + "/" + name
		if ext != "" {
			dst = dst + ext
		}
		req := &ossv2.PutObjectRequest{
			Bucket: ossv2.Ptr(oc.config.Bucket),
			Key:    ossv2.Ptr(dst),
		}
		if mimeType != "" {
			req.ContentType = ossv2.Ptr(mimeType)
		}
		// 生成 PutObject 的预签名 URL
		result, err := oc.client.Presign(ctx, req, ossv2.PresignExpires(10*time.Minute))
		if err != nil {
			return nil, err
		}
		if len(isNotInternal) > 0 && isNotInternal[0] {
			result.URL = strings.Replace(result.URL, "-internal", "", -1)
		}
		list = append(list, FilePreSignatureURL{
			Url:         dst,
			SignUrl:     result.URL,
			ContentType: mimeType,
		})
	}
	return list, nil
}

// commonMimeTypes 常用类型 Content-Type 兜底表。
// 直接查表避免依赖系统 mime 表(Windows 查注册表、Linux 查内置表,结果可能不一致)
var commonMimeTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".bmp":  "image/bmp",
	".svg":  "image/svg+xml",
	".ico":  "image/x-icon",
	".pdf":  "application/pdf",
	".mp4":  "video/mp4",
	".mp3":  "audio/mpeg",
	".zip":  "application/zip",
}

// contentTypeByExt 根据扩展名获取 Content-Type,优先查内置表,查不到再回退系统 mime 表
func contentTypeByExt(ext string) string {
	if mimeType, ok := commonMimeTypes[strings.ToLower(ext)]; ok {
		return mimeType
	}
	return mime.TypeByExtension(ext)
}

// parseEndpoint 解析 endpoint 的协议与域名(去掉路径与末尾斜杠),未显式指定协议时默认 https
func parseEndpoint(endpoint string) (scheme, host string) {
	scheme = "https"
	switch {
	case strings.HasPrefix(endpoint, "http://"):
		scheme = "http"
		endpoint = strings.TrimPrefix(endpoint, "http://")
	case strings.HasPrefix(endpoint, "https://"):
		endpoint = strings.TrimPrefix(endpoint, "https://")
	}
	if idx := strings.Index(endpoint, "/"); idx >= 0 {
		endpoint = endpoint[:idx]
	}
	return scheme, endpoint
}

// GetUrl 获取对象存储URL
// 优先基于 Config.Endpoint 生成(虚拟主机风格: {协议}://{bucket}.{endpoint域名}/{object}),
// 未配置 Endpoint 时回退为标准 OSS 地域域名
func (oc *OClient) GetUrl(objectName string, isInternal ...bool) string {
	internal := len(isInternal) > 0 && isInternal[0]
	if oc.config.Endpoint == "" {
		if internal {
			return "https://" + oc.config.Bucket + ".oss-" + oc.config.Region + "-internal.aliyuncs.com/" + objectName
		}
		return "https://" + oc.config.Bucket + ".oss-" + oc.config.Region + ".aliyuncs.com/" + objectName
	}
	scheme, host := parseEndpoint(oc.config.Endpoint)
	if internal {
		host = strings.Replace(host, ".aliyuncs.com", "-internal.aliyuncs.com", 1)
	}
	return scheme + "://" + oc.config.Bucket + "." + host + "/" + objectName
}
