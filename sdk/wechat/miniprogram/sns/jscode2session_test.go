package sns

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"resty.dev/v3"

	"github.com/laixhe/gonet/sdk/wechat/apiutil"
)

func testClient(t *testing.T, srv *httptest.Server) *resty.Client {
	t.Helper()
	return resty.New().SetBaseURL(srv.URL)
}

func TestJsCode2SessionSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sns/jscode2session" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("appid") != "appid" || q.Get("secret") != "secret" || q.Get("js_code") != "code1" || q.Get("grant_type") != "authorization_code" {
			t.Errorf("unexpected query: %v", q)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"openid":"openid1","session_key":"sk","unionid":"union1"}`))
	}))
	defer srv.Close()

	resp, err := JsCode2Session(context.Background(), testClient(t, srv), "appid", "secret", "code1")
	if err != nil {
		t.Fatalf("JsCode2Session: %v", err)
	}
	if resp.OpenID != "openid1" || resp.SessionKey != "sk" || resp.UnionID != "union1" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestJsCode2SessionBusinessError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":40029,"errmsg":"invalid code"}`))
	}))
	defer srv.Close()

	_, err := JsCode2Session(context.Background(), testClient(t, srv), "appid", "secret", "bad")
	var apiErr *apiutil.ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if apiErr.ErrCode != 40029 {
		t.Fatalf("unexpected errcode: %d", apiErr.ErrCode)
	}
}
