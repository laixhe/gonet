package oss

import (
	"context"

	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

// Delete 删除单个文件
func (oc *OssClient) Delete(ctx context.Context, objectName string) error {
	request := &ossv2.DeleteObjectRequest{
		Bucket: ossv2.Ptr(oc.config.Bucket), // 存储空间名称
		Key:    ossv2.Ptr(objectName),       // 对象名称
	}
	_, err := oc.client.DeleteObject(ctx, request)
	if err != nil {
		return err
	}
	return nil
}
