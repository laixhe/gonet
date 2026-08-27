package oplatform

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"resty.dev/v3"
)

func TestAccessTokenSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sns/oauth2/access_token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("appid") != "appid" || q.Get("secret") != "secret" || q.Get("code") != "code1" || q.Get("grant_type") != "authorization_code" {
			t.Errorf("unexpected query: %v", q)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"access_token":"at1","openid":"openid1","unionid":"union1","refresh_token":"rt1","expires_in":7200,"scope":"snsapi_userinfo"}`))
	}))
	defer srv.Close()

	o := &OpenPlatform{
		config:     &Config{AppID: "appid", Secret: "secret"},
		httpClient: resty.New().SetBaseURL(srv.URL),
	}
	resp, err := o.AccessToken(context.Background(), "code1")
	if err != nil {
		t.Fatalf("AccessToken: %v", err)
	}
	if resp.AccessToken != "at1" || resp.OpenID != "openid1" || resp.UnionID != "union1" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestAccessTokenAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":40029,"errmsg":"invalid code"}`))
	}))
	defer srv.Close()

	o := &OpenPlatform{
		config:     &Config{AppID: "appid", Secret: "secret"},
		httpClient: resty.New().SetBaseURL(srv.URL),
	}
	_, err := o.AccessToken(context.Background(), "bad")
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if apiErr.ErrCode != 40029 {
		t.Fatalf("unexpected errcode: %d", apiErr.ErrCode)
	}
}

func TestNewOpenPlatformPanicOnInvalidConfig(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for invalid config")
		}
	}()
	NewOpenPlatform(&Config{})
}
