package sdkoss

import (
	"context"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

// Delete 删除
func Delete(objectName string) error {
	// 创建删除对象的请求
	request := &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(sdkOss.c.Bucket), // 存储空间名称
		Key:    oss.Ptr(objectName),      // 对象名称
	}
	// 执行删除对象的操作并处理结果
	_, err := sdkOss.client.DeleteObject(context.TODO(), request)
	if err != nil {
		return err
	}
	return nil
}
