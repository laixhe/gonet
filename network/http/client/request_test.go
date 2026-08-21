package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestText(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello:" + r.URL.Query().Get("name")))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	body, err := c.Get("/greet").SetQueryParam("name", "world").Text()
	if err != nil {
		t.Fatalf("Text 失败: %v", err)
	}
	if body != "hello:world" {
		t.Errorf("body = %q, want %q", body, "hello:world")
	}
}
