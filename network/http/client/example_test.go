package client_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/laixhe/gonet/network/http/client"
)

func ExampleClient_Get() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello:" + r.URL.Query().Get("name")))
	}))
	defer srv.Close()

	body, err := client.NewClient(srv.URL).
		Get("/greet").
		SetQueryParam("name", "world").
		Text()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(body)
	// Output: hello:world
}

func ExampleRequest_SetJSON() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v map[string]string
		_ = json.NewDecoder(r.Body).Decode(&v)
		_ = json.NewEncoder(w).Encode(v)
	}))
	defer srv.Close()

	var out map[string]string
	err := client.NewClient(srv.URL).
		Post("/echo").
		SetJSON(map[string]string{"name": "gonet"}).
		JSON(&out)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(out["name"])
	// Output: gonet
}

func ExampleRequest_SetFile() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		b, _ := io.ReadAll(file)
		_, _ = w.Write(b)
	}))
	defer srv.Close()

	body, err := client.NewClient(srv.URL).
		Post("/upload").
		SetFile("file", "a.txt", []byte("hello")).
		Text()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(body)
	// Output: hello
}

func ExampleRequest_Download() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("file-content"))
	}))
	defer srv.Close()

	path := filepath.Join(os.TempDir(), "gonet-example-download.txt")
	defer os.Remove(path)

	if err := client.NewClient(srv.URL).Get("/file").Download(path); err != nil {
		log.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	fmt.Println(string(data))
	// Output: file-content
}

func ExampleRequest_SetRetryCount() {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	body, err := client.NewClient(srv.URL).
		Get("/unstable").
		SetRetryCount(3).
		SetRetryWaitTime(10 * time.Millisecond).
		Text()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(body)
	// Output: ok
}

// apiError 示例错误结构体, 实现 error 接口用于 SetError
type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *apiError) Error() string { return fmt.Sprintf("code=%d %s", e.Code, e.Message) }

func ExampleRequest_SetError() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":1001,"message":"bad request"}`))
	}))
	defer srv.Close()

	var out any
	err := client.NewClient(srv.URL).Get("/error").SetError(&apiError{}).JSON(&out)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("no error")
	// Output: code=1001 bad request
}

func ExampleClient_OnBeforeRequest() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Token")))
	}))
	defer srv.Close()

	c := client.NewClient(srv.URL).OnBeforeRequest(func(req *http.Request) error {
		req.Header.Set("X-Token", "secret")
		return nil
	})

	body, err := c.Get("/api").Text()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(body)
	// Output: secret
}
