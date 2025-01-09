package sdkoss

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/laixhe/gonet/xfile"
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

// PreSignatureURL 生成预签名文件上传url
func PreSignatureURL(fileNames []string) ([]FilePreSignatureURL, error) {
	list := make([]FilePreSignatureURL, 0, len(fileNames))
	for _, fileName := range fileNames {
		ext := strings.ToLower(strings.TrimLeft(filepath.Ext(fileName), "."))
		if !xfile.IsType(ext) {
			return nil, errors.New("file type not exist")
		}
		dir := time.Now().Format("2006/01/02")
		name := xid.New().String()
		dst := dir + "/" + name + "." + ext
		contentType := xfile.GetContentType(ext)
		// 生成 PutObject 的预签名 URL
		result, err := sdkOss.client.Presign(context.TODO(), &oss.PutObjectRequest{
			Bucket:      oss.Ptr(sdkOss.c.Bucket),
			Key:         oss.Ptr(dst),
			ContentType: oss.Ptr(contentType),
		}, oss.PresignExpires(10*time.Minute))
		if err != nil {
			return nil, err
		}
		signUrl := strings.Replace(result.URL, "-internal", "", -1)
		list = append(list, FilePreSignatureURL{
			Url:         dst,
			SignUrl:     signUrl,
			ContentType: contentType,
		})
	}
	return list, nil
}

// GetPreSignatureURL 获取预签名文件上传url
func GetPreSignatureURL(objectName string) string {
	return "https://" + sdkOss.c.Bucket + ".oss-" + sdkOss.c.Region + ".aliyuncs.com/" + objectName
}

// SetObjectACL 设置文件的访问权限
func SetObjectACL(objectName string) error {
	// 创建设置对象 ACL 的请求
	req := &oss.PutObjectAclRequest{
		Bucket: oss.Ptr(sdkOss.c.Bucket), // 存储空间名称
		Key:    oss.Ptr(objectName),      // 对象名称
		Acl:    oss.ObjectACLPublicRead,  // 设置对象的访问权限为私有
	}
	// 执行设置对象 ACL 的操作
	_, err := sdkOss.client.PutObjectAcl(context.TODO(), req)
	if err != nil {
		return err
	}
	return nil
}
