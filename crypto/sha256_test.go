package crypto

import (
	"strings"
	"testing"
)

func TestSHA256(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{"normal string", "hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"empty string", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"numbers", "123456", "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SHA256(tt.data); got != tt.want {
				t.Errorf("SHA256() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSHA256_Consistency(t *testing.T) {
	result := SHA256("consistency-test")
	for i := 0; i < 100; i++ {
		if got := SHA256("consistency-test"); got != result {
			t.Fatalf("inconsistent SHA256 result at iteration %d", i)
		}
	}
}

func TestHmacSha256(t *testing.T) {
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
			want: "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8",
		},
		{
			name: "empty data",
			key:  "secret",
			data: "",
			want: "f9e66e179b6747ae54108f82f8ade8b3c25d76fd30afde6c395822c530196169",
		},
		{
			name: "empty key and data",
			key:  "",
			data: "",
			want: "b613679a0814d9ec772f95d778c35fc5ff1697c493715653c6c712144292c5ad",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HmacSha256(tt.key, tt.data); got != tt.want {
				t.Errorf("HmacSha256() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSHA256_DifferentInputsDifferentResults(t *testing.T) {
	a := SHA256("hello")
	b := SHA256("world")
	if a == b {
		t.Error("different inputs should produce different results")
	}
	if len(a) != 64 || len(b) != 64 {
		t.Errorf("SHA256 output should be 64 chars, got %d and %d", len(a), len(b))
	}
}

func TestHmacSha256_LongKey(t *testing.T) {
	longKey := strings.Repeat("k", 256)
	result := HmacSha256(longKey, "test data")
	if len(result) != 64 {
		t.Errorf("HMAC-SHA256 output should be 64 chars, got %d", len(result))
	}
}
