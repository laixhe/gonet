package oss

import (
	"context"

	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

// Get 简单下载
func Get(objectName string, fn func(*ossv2.GetObjectResult) error) error {
	// 创建获取对象的请求
	request := &ossv2.GetObjectRequest{
		Bucket: ossv2.Ptr(sdkOss.c.Bucket), // 存储空间名称
		Key:    ossv2.Ptr(objectName),      // 对象名称
	}
	// 执行获取对象的操作并处理结果
	result, err := sdkOss.client.GetObject(context.TODO(), request)
	if err != nil {
		return err
	}
	// 确保在函数结束时关闭响应体
	defer result.Body.Close()
	//
	return fn(result)
}
