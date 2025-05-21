package oss

import (
	"context"

	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

// Delete 删除
func (s *SdkAliyunOss) Delete(objectName string) error {
	// 创建删除对象的请求
	request := &ossv2.DeleteObjectRequest{
		Bucket: ossv2.Ptr(s.config.Bucket), // 存储空间名称
		Key:    ossv2.Ptr(objectName),      // 对象名称
	}
	// 执行删除对象的操作并处理结果
	_, err := s.client.DeleteObject(context.TODO(), request)
	if err != nil {
		return err
	}
	return nil
}
