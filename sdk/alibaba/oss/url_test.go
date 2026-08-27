package oss

import (
	"testing"
)

func TestGetUrlDefault(t *testing.T) {
	oc := &OClient{config: &Config{Bucket: "test", Region: "cn-shenzhen"}}
	if got := oc.GetUrl("a.jpg"); got != "https://test.oss-cn-shenzhen.aliyuncs.com/a.jpg" {
		t.Fatalf("unexpected url: %s", got)
	}
	if got := oc.GetUrl("a.jpg", true); got != "https://test.oss-cn-shenzhen-internal.aliyuncs.com/a.jpg" {
		t.Fatalf("unexpected internal url: %s", got)
	}
}

func TestGetUrlWithEndpoint(t *testing.T) {
	// 标准 OSS endpoint
	oc := &OClient{config: &Config{Bucket: "test", Endpoint: "https://oss-cn-shenzhen.aliyuncs.com"}}
	if got := oc.GetUrl("a.jpg"); got != "https://test.oss-cn-shenzhen.aliyuncs.com/a.jpg" {
		t.Fatalf("unexpected url: %s", got)
	}
	if got := oc.GetUrl("a.jpg", true); got != "https://test.oss-cn-shenzhen-internal.aliyuncs.com/a.jpg" {
		t.Fatalf("unexpected internal url: %s", got)
	}
	// 自定义 CDN 域名 + http 协议
	oc2 := &OClient{config: &Config{Bucket: "test", Endpoint: "http://cdn.example.com"}}
	if got := oc2.GetUrl("a.jpg"); got != "http://test.cdn.example.com/a.jpg" {
		t.Fatalf("unexpected cdn url: %s", got)
	}
}

func TestParseEndpoint(t *testing.T) {
	cases := []struct {
		endpoint string
		scheme   string
		host     string
	}{
		{"https://oss-cn-shenzhen.aliyuncs.com", "https", "oss-cn-shenzhen.aliyuncs.com"},
		{"https://oss-cn-shenzhen.aliyuncs.com/", "https", "oss-cn-shenzhen.aliyuncs.com"},
		{"http://cdn.example.com", "http", "cdn.example.com"},
		{"oss-cn-shenzhen.aliyuncs.com", "https", "oss-cn-shenzhen.aliyuncs.com"},
	}
	for _, c := range cases {
		scheme, host := parseEndpoint(c.endpoint)
		if scheme != c.scheme || host != c.host {
			t.Fatalf("parseEndpoint(%q) = %q %q, want %q %q", c.endpoint, scheme, host, c.scheme, c.host)
		}
	}
}

func TestContentTypeByExt(t *testing.T) {
	cases := map[string]string{
		".jpg":  "image/jpeg",
		".JPG":  "image/jpeg", // 大小写不敏感
		".png":  "image/png",
		".webp": "image/webp",
		".pdf":  "application/pdf",
	}
	for ext, want := range cases {
		if got := contentTypeByExt(ext); got != want {
			t.Fatalf("contentTypeByExt(%q) = %q, want %q", ext, got, want)
		}
	}
	// 未知扩展名不 panic,回退系统 mime 表(可能为空)
	_ = contentTypeByExt(".zzz_unknown")
}
