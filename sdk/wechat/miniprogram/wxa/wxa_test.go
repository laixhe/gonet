package wxa

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"resty.dev/v3"

	"github.com/laixhe/gonet/sdk/wechat/apiutil"
)

func testClient(t *testing.T, srv *httptest.Server) *resty.Client {
	t.Helper()
	return resty.New().SetBaseURL(srv.URL)
}

func TestGetUserPhoneNumberSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/business/getuserphonenumber" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("access_token") != "tok" {
			t.Errorf("unexpected access_token: %s", r.URL.Query().Get("access_token"))
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"code":"phcode"`) {
			t.Errorf("unexpected body: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		// timestamp 使用数字形态,验证 FlexString 兼容(修复前 string 字段会解析失败)
		_, _ = w.Write([]byte(`{"errcode":0,"phone_info":{"phoneNumber":"13800138000","purePhoneNumber":"13800138000","countryCode":"86","watermark":{"timestamp":1700000000,"appid":"appid"}}}`))
	}))
	defer srv.Close()

	resp, err := GetUserPhoneNumber(context.Background(), testClient(t, srv), "tok", "phcode")
	if err != nil {
		t.Fatalf("GetUserPhoneNumber: %v", err)
	}
	if resp.PhoneInfo.PhoneNumber != "13800138000" || resp.PhoneInfo.CountryCode != "86" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
	if resp.PhoneInfo.Watermark.Timestamp != "1700000000" {
		t.Fatalf("unexpected timestamp: %q", resp.PhoneInfo.Watermark.Timestamp)
	}
}

func TestGenerateSchemeNoHTMLEscape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/generatescheme" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		// & 不能被 JSON 转义成 \u0026,否则微信侧参数解析错误
		if !strings.Contains(string(body), "id=1&age=18") {
			t.Errorf("query was HTML-escaped, body: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"openlink":"weixin://dl/business/?t=abc"}`))
	}))
	defer srv.Close()

	resp, err := GenerateScheme(context.Background(), testClient(t, srv), "tok", &GenerateSchemeRequest{
		JumpWxa: GenerateSchemeJumpWxa{Path: "/pages/index/index", Query: "id=1&age=18"},
	})
	if err != nil {
		t.Fatalf("GenerateScheme: %v", err)
	}
	if resp.OpenLink == "" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestQuerySchemeSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/queryscheme" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"scheme_info":{"appid":"appid","path":"/pages/index/index","query":"id=1","env_version":"release","expire_time":0,"create_time":1700000000},"quota_info":{"remain_visit_quota":1000}}`))
	}))
	defer srv.Close()

	resp, err := QueryScheme(context.Background(), testClient(t, srv), "tok", &QuerySchemeRequest{Scheme: "weixin://dl/business/?t=abc"})
	if err != nil {
		t.Fatalf("QueryScheme: %v", err)
	}
	if resp.SchemeInfo.Path != "/pages/index/index" || resp.QuotaInfo.RemainVisitQuota != 1000 {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestGetWxaCodeUnlimitImage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/getwxacodeunlimit" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"width":1280`) {
			t.Errorf("default width not applied, body: %s", body)
		}
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte{0xff, 0xd8, 0xff})
	}))
	defer srv.Close()

	resp, err := GetWxaCodeUnlimit(context.Background(), testClient(t, srv), "tok", &GetWxaCodeUnlimitRequest{
		Page:  "pages/index/index",
		Scene: "id=1&age=18",
	})
	if err != nil {
		t.Fatalf("GetWxaCodeUnlimit: %v", err)
	}
	if resp.ContentType != "image/jpeg" || len(resp.Buffer) != 3 {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestGetWxaCodeUnlimitError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":40129,"errmsg":"scene invalid"}`))
	}))
	defer srv.Close()

	resp, err := GetWxaCodeUnlimit(context.Background(), testClient(t, srv), "tok", &GetWxaCodeUnlimitRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *apiutil.ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if apiErr.ErrCode != 40129 || resp.ErrCode != 40129 {
		t.Fatalf("unexpected err/response: %v / %+v", err, resp)
	}
}
