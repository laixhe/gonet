package oss

import (
	"context"

	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

// SetObjectACL 设置文件的访问权限
func (oc *OssClient) SetObjectACL(ctx context.Context, objectName string, acls ...ossv2.ObjectACLType) error {
	acl := ossv2.ObjectACLPublicRead // 设置对象的访问权限为：公共读
	if len(acls) == 0 {
		acl = acls[0]
	}
	req := &ossv2.PutObjectAclRequest{
		Bucket: ossv2.Ptr(oc.config.Bucket), // 存储空间名称
		Key:    ossv2.Ptr(objectName),       // 对象名称
		Acl:    acl,
	}
	_, err := oc.client.PutObjectAcl(ctx, req)
	if err != nil {
		return err
	}
	return nil
}
