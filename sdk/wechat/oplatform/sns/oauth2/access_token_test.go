package oauth2

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"resty.dev/v3"

	"github.com/laixhe/gonet/sdk/wechat/oplatform/internal/apiutil"
)

func testClient(t *testing.T, srv *httptest.Server) *resty.Client {
	t.Helper()
	return resty.New().SetBaseURL(srv.URL)
}

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

	resp, err := AccessToken(context.Background(), testClient(t, srv), "appid", "secret", "code1")
	if err != nil {
		t.Fatalf("AccessToken: %v", err)
	}
	if resp.AccessToken != "at1" || resp.OpenID != "openid1" || resp.UnionID != "union1" || resp.RefreshToken != "rt1" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestAccessTokenBusinessError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":40029,"errmsg":"invalid code"}`))
	}))
	defer srv.Close()

	_, err := AccessToken(context.Background(), testClient(t, srv), "appid", "secret", "bad")
	var apiErr *apiutil.ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if apiErr.ErrCode != 40029 {
		t.Fatalf("unexpected errcode: %d", apiErr.ErrCode)
	}
}
