package crypto

import (
	"strings"
	"testing"
)

func TestMD5(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{"normal string", "hello", "5d41402abc4b2a76b9719d911017c592"},
		{"empty string", "", "d41d8cd98f00b204e9800998ecf8427e"},
		{"numbers", "123456", "e10adc3949ba59abbe56e057f20f883e"},
		{"chinese chars", "你好世界", "65396ee4aad0b4f17aacd1c6112ee364"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MD5(tt.data); got != tt.want {
				t.Errorf("MD5() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMD5_Consistency(t *testing.T) {
	// 同一输入多次调用结果一致
	result := MD5("consistency-test")
	for i := 0; i < 100; i++ {
		if got := MD5("consistency-test"); got != result {
			t.Fatalf("inconsistent MD5 result at iteration %d: %s vs %s", i, got, result)
		}
	}
}

func TestHmacMd5(t *testing.T) {
	tests := []struct {
		name string
		key  string
		data string
		want string
	}{
		{
			name: "standard test vector",
			key:  "key",
			data: "The quick brown fox jumps over the lazy dog",
			want: "80070713463e7749b90c2dc24911e275",
		},
		{
			name: "empty data",
			key:  "secret",
			data: "",
			want: "5c8db03f04cec0f43bcb060023914190",
		},
		{
			name: "empty key and data",
			key:  "",
			data: "",
			want: "74e6f7298a9c2d168935f58c001bad88",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HmacMd5(tt.key, tt.data); got != tt.want {
				t.Errorf("HmacMd5() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHmacMd5_DifferentKeys(t *testing.T) {
	// 不同密钥应产生不同的 HMAC
	data := "same data"
	h1 := HmacMd5("key1", data)
	h2 := HmacMd5("key2", data)
	if h1 == h2 {
		t.Error("different keys should produce different HMACs")
	}
}

func TestMD5_DifferentInputsDifferentResults(t *testing.T) {
	a := MD5("hello")
	b := MD5("world")
	if a == b {
		t.Error("different inputs should produce different results")
	}
	// 验证输出格式：始终是 32 字符小写十六进制
	if len(a) != 32 || len(b) != 32 {
		t.Errorf("MD5 output should be 32 chars, got %d and %d", len(a), len(b))
	}
}

func TestHmacMd5_LongKey(t *testing.T) {
	longKey := strings.Repeat("k", 256)
	data := "test data"
	result := HmacMd5(longKey, data)
	if len(result) != 32 {
		t.Errorf("HMAC-MD5 output should be 32 chars, got %d", len(result))
	}
}
