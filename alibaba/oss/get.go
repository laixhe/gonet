package oss

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

// Get 简单下载
func (oc *OssClient) Get(ctx context.Context, objectName string) ([]byte, error) {
	// 创建获取对象的请求
	request := &ossv2.GetObjectRequest{
		Bucket: ossv2.Ptr(oc.config.Bucket), // 存储空间名称
		Key:    ossv2.Ptr(objectName),       // 对象名称
	}
	// 执行获取对象的操作并处理结果
	result, err := oc.client.GetObject(ctx, request)
	if err != nil {
		return nil, err
	}
	// 确保在函数结束时关闭响应体
	defer result.Body.Close()
	return io.ReadAll(result.Body)
}

// 公共读或者公共读写图片的信息

type GetInfoValue struct {
	Value string `json:"value"`
}

type GetInfoResponse struct {
	FileSize    GetInfoValue `json:"FileSize"`
	Format      GetInfoValue `json:"Format"`
	ImageHeight GetInfoValue `json:"ImageHeight"`
	ImageWidth  GetInfoValue `json:"ImageWidth"`
}

// GetInfo 获取公共读或者公共读写图片的信息
func (oc *OssClient) GetInfo(ctx context.Context, objectName string) (*GetInfoResponse, error) {
	if strings.HasPrefix(objectName, "http") {
		objectName = strings.Split(objectName, "?")[0]
	} else {
		objectName = oc.GetUrl(objectName)
	}
	resp, err := http.Get(objectName + "?x-oss-process=image/info")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var respBody GetInfoResponse
	err = json.Unmarshal(body, &respBody)
	if err != nil {
		return nil, err
	}
	return &respBody, nil
}
