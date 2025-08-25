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
func (oc *OssClient) SetPreSignatureURL(ctx context.Context, fileDir string, fileNames []string, isNotInternal ...bool) ([]FilePreSignatureURL, error) {
	list := make([]FilePreSignatureURL, 0, len(fileNames))
	for _, fileName := range fileNames {
		ext := filepath.Ext(fileName)
		mimeType := mime.TypeByExtension(ext)
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

// GetUrl 获取对象存储URL
func (oc *OssClient) GetUrl(objectName string, isInternal ...bool) string {
	return "https://" + oc.config.Bucket + ".oss-" + oc.config.Region + ".aliyuncs.com/" + objectName
}
