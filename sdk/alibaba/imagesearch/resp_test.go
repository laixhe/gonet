package imagesearch

import (
	"strings"
	"testing"

	clientv4 "github.com/alibabacloud-go/imagesearch-20201214/v4/client"
)

// TestAddImageRespErr 覆盖 nil 响应、nil body、成功、业务失败四类场景
func TestAddImageRespErr(t *testing.T) {
	// nil 响应
	if err := addImageRespErr(nil); err == nil {
		t.Fatal("expected error for nil resp")
	}
	// nil body
	resp := &clientv4.AddImageResponse{}
	resp.SetStatusCode(200)
	if err := addImageRespErr(resp); err == nil {
		t.Fatal("expected error for nil body")
	}
	// 成功
	okBody := &clientv4.AddImageResponseBody{}
	okBody.SetCode(0).SetSuccess(true)
	resp.SetBody(okBody)
	if err := addImageRespErr(resp); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	// 业务失败
	failBody := &clientv4.AddImageResponseBody{}
	failBody.SetCode(1001).SetSuccess(false).SetMessage("实例不存在")
	resp.SetBody(failBody)
	err := addImageRespErr(resp)
	if err == nil {
		t.Fatal("expected error for business failure")
	}
	if !strings.Contains(err.Error(), "1001") || !strings.Contains(err.Error(), "实例不存在") {
		t.Fatalf("unexpected err: %v", err)
	}
}

// TestDeleteImageRespErr 覆盖 nil 响应、nil body、成功、业务失败四类场景
func TestDeleteImageRespErr(t *testing.T) {
	if err := deleteImageRespErr(nil); err == nil {
		t.Fatal("expected error for nil resp")
	}
	resp := &clientv4.DeleteImageResponse{}
	resp.SetStatusCode(200)
	if err := deleteImageRespErr(resp); err == nil {
		t.Fatal("expected error for nil body")
	}
	okBody := &clientv4.DeleteImageResponseBody{}
	okBody.SetCode(0).SetSuccess(true)
	resp.SetBody(okBody)
	if err := deleteImageRespErr(resp); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	failBody := &clientv4.DeleteImageResponseBody{}
	failBody.SetCode(500).SetSuccess(false).SetMessage("删除失败")
	resp.SetBody(failBody)
	err := deleteImageRespErr(resp)
	if err == nil {
		t.Fatal("expected error for business failure")
	}
	if !strings.Contains(err.Error(), "500") || !strings.Contains(err.Error(), "删除失败") {
		t.Fatalf("unexpected err: %v", err)
	}
}

// TestSearchImageByPicRespErr 覆盖 nil 响应、nil body、成功、业务失败四类场景
func TestSearchImageByPicRespErr(t *testing.T) {
	if err := searchImageByPicRespErr(nil); err == nil {
		t.Fatal("expected error for nil resp")
	}
	resp := &clientv4.SearchImageByPicResponse{}
	resp.SetStatusCode(200)
	if err := searchImageByPicRespErr(resp); err == nil {
		t.Fatal("expected error for nil body")
	}
	okBody := &clientv4.SearchImageByPicResponseBody{}
	okBody.SetCode(0).SetSuccess(true)
	resp.SetBody(okBody)
	if err := searchImageByPicRespErr(resp); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	failBody := &clientv4.SearchImageByPicResponseBody{}
	failBody.SetCode(101).SetSuccess(false).SetMsg("搜索失败")
	resp.SetBody(failBody)
	err := searchImageByPicRespErr(resp)
	if err == nil {
		t.Fatal("expected error for business failure")
	}
	if !strings.Contains(err.Error(), "101") || !strings.Contains(err.Error(), "搜索失败") {
		t.Fatalf("unexpected err: %v", err)
	}
}
