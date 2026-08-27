package cgibin

import (
	"context"
	"encoding/json"
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

func TestStableTokenSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/stable_token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body["appid"] != "appid" || body["secret"] != "secret" {
			t.Errorf("unexpected body: %v", body)
		}
		if body["force_refresh"] != false {
			t.Errorf("expected force_refresh=false, got %v", body["force_refresh"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"access_token":"tok123","expires_in":7200}`))
	}))
	defer srv.Close()

	resp, err := StableToken(context.Background(), testClient(t, srv), "appid", "secret", false)
	if err != nil {
		t.Fatalf("StableToken: %v", err)
	}
	if resp.AccessToken != "tok123" || resp.ExpiresIn != 7200 {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestTokenQueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("appid") != "appid" || q.Get("secret") != "secret" || q.Get("grant_type") != "client_credential" {
			t.Errorf("unexpected query: %v", q)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"access_token":"tok123","expires_in":7200}`))
	}))
	defer srv.Close()

	resp, err := Token(context.Background(), testClient(t, srv), "appid", "secret")
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if resp.AccessToken != "tok123" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestStableTokenBusinessError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":40164,"errmsg":"invalid ip"}`))
	}))
	defer srv.Close()

	_, err := StableToken(context.Background(), testClient(t, srv), "appid", "secret", false)
	var apiErr *apiutil.ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if apiErr.ErrCode != 40164 || apiErr.ErrMsg != "invalid ip" {
		t.Fatalf("unexpected err: %+v", apiErr)
	}
}
