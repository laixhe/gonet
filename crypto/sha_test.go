package crypto

import (
	"strings"
	"testing"
)

func TestSHA1(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{"normal string", "hello", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{"empty string", "", "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
		{"numbers", "123456", "7c4a8d09ca3762af61e59520943dc26494f8941b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SHA1(tt.data); got != tt.want {
				t.Errorf("SHA1() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSHA1_Consistency(t *testing.T) {
	result := SHA1("consistency-test")
	for i := 0; i < 100; i++ {
		if got := SHA1("consistency-test"); got != result {
			t.Fatalf("inconsistent SHA1 result at iteration %d", i)
		}
	}
}

func TestHmacSha1(t *testing.T) {
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
			want: "de7c9b85b8b78aa6bc8a7a36f70a90701c9db4d9",
		},
		{
			name: "empty data",
			key:  "secret",
			data: "",
			want: "25af6174a0fcecc4d346680a72b7ce644b9a88e8",
		},
		{
			name: "empty key and data",
			key:  "",
			data: "",
			want: "fbdb1d1b18aa6c08324b7d64b71fb76370690e1d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HmacSha1(tt.key, tt.data); got != tt.want {
				t.Errorf("HmacSha1() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSHA1_DifferentInputsDifferentResults(t *testing.T) {
	a := SHA1("hello")
	b := SHA1("world")
	if a == b {
		t.Error("different inputs should produce different results")
	}
	if len(a) != 40 || len(b) != 40 {
		t.Errorf("SHA1 output should be 40 chars, got %d and %d", len(a), len(b))
	}
}

func TestHmacSha1_LongKey(t *testing.T) {
	longKey := strings.Repeat("k", 256)
	result := HmacSha1(longKey, "test data")
	if len(result) != 40 {
		t.Errorf("HMAC-SHA1 output should be 40 chars, got %d", len(result))
	}
}
