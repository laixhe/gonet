package cgibin

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

func TestGetTokenSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/gettoken" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("corpid") != "corp1" || q.Get("corpsecret") != "secret1" {
			t.Errorf("unexpected query: %v", q)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"access_token":"tok1","expires_in":7200}`))
	}))
	defer srv.Close()

	resp, err := GetToken(context.Background(), testClient(t, srv), "corp1", "secret1")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	if resp.AccessToken != "tok1" || resp.ExpiresIn != 7200 {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestGetUserInfoSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/auth/getuserinfo" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("access_token") != "tok" || q.Get("code") != "code1" {
			t.Errorf("unexpected query: %v", q)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"userid":"zhangsan","user_ticket":"tk1","openid":"openid1","external_userid":"ext1"}`))
	}))
	defer srv.Close()

	resp, err := GetUserInfo(context.Background(), testClient(t, srv), "tok", "code1")
	if err != nil {
		t.Fatalf("GetUserInfo: %v", err)
	}
	if resp.UserID != "zhangsan" || resp.UserTicket != "tk1" || resp.ExternalUserID != "ext1" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestGetUserDetailSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/auth/getuserdetail" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"user_ticket":"tk1"`) {
			t.Errorf("unexpected body: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		// gender 使用数字形态,验证 FlexString 兼容(修复前 string 字段会解析失败)
		_, _ = w.Write([]byte(`{"errcode":0,"userid":"zhangsan","gender":1,"mobile":"13800138000"}`))
	}))
	defer srv.Close()

	resp, err := GetUserDetail(context.Background(), testClient(t, srv), "tok", "tk1")
	if err != nil {
		t.Fatalf("GetUserDetail: %v", err)
	}
	if resp.UserID != "zhangsan" || resp.Mobile != "13800138000" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
	if resp.Gender != "1" {
		t.Fatalf("unexpected gender: %q", resp.Gender)
	}
}

func TestUserGetSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/user/get" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("access_token") != "tok" || q.Get("userid") != "zhangsan" {
			t.Errorf("unexpected query: %v", q)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"userid":"zhangsan","name":"张三","gender":1,"mobile":"13800138000","department":[1,2]}`))
	}))
	defer srv.Close()

	resp, err := UserGet(context.Background(), testClient(t, srv), "tok", "zhangsan")
	if err != nil {
		t.Fatalf("UserGet: %v", err)
	}
	if resp.Name != "张三" || len(resp.Department) != 2 {
		t.Fatalf("unexpected resp: %+v", resp)
	}
	if resp.Gender != "1" {
		t.Fatalf("unexpected gender: %q", resp.Gender)
	}
}

func TestUserGetBusinessError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":60111,"errmsg":"user not found"}`))
	}))
	defer srv.Close()

	_, err := UserGet(context.Background(), testClient(t, srv), "tok", "nobody")
	var apiErr *apiutil.ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if apiErr.ErrCode != 60111 {
		t.Fatalf("unexpected errcode: %d", apiErr.ErrCode)
	}
}
