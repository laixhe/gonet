package oss

import (
	"context"
	"io"

	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

// PutFile 简单上传
func (oc *OssClient) Put(ctx context.Context, objectName string, body io.Reader, contentType string, acls ...ossv2.ObjectACLType) error {
	request := &ossv2.PutObjectRequest{
		Bucket: ossv2.Ptr(oc.config.Bucket), // 存储空间名称
		Key:    ossv2.Ptr(objectName),       // 对象名称
		Body:   body,
	}
	if len(acls) > 0 {
		request.Acl = acls[0]
	}
	if contentType != "" {
		request.ContentType = ossv2.Ptr(contentType)
	}
	_, err := oc.client.PutObject(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

// PutFile 简单上传文件
func (oc *OssClient) PutFile(ctx context.Context, objectName string, localFile string, contentType string, acls ...ossv2.ObjectACLType) error {
	request := &ossv2.PutObjectRequest{
		Bucket: ossv2.Ptr(oc.config.Bucket), // 存储空间名称
		Key:    ossv2.Ptr(objectName),       // 对象名称
	}
	if len(acls) > 0 {
		request.Acl = acls[0]
	}
	if contentType != "" {
		request.ContentType = ossv2.Ptr(contentType)
	}
	_, err := oc.client.PutObjectFromFile(ctx, request, localFile)
	if err != nil {
		return err
	}
	return nil
}
