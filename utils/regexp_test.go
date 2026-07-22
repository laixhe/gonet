package utils

import (
	"encoding/base64"
	"testing"
)

// ======================== IsContainNumber ========================

func TestIsContainNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"with_digit", "abc123", true},
		{"pure_digit", "123", true},
		{"no_digit", "abc", false},
		{"empty", "", false},
		{"chinese_with_digit", "你好123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsContainNumber(tt.input); got != tt.want {
				t.Fatalf("IsContainNumber(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ======================== IsContainLetter ========================

func TestIsContainLetter(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{"with_letter", "abc123", true},
		{"pure_letter", "abc", true},
		{"no_letter", "123", false},
		{"empty", "", false},
		{"chinese_only", "你好", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsContainLetter(tt.str); got != tt.want {
				t.Fatalf("IsContainLetter(%q) = %v, want %v", tt.str, got, tt.want)
			}
		})
	}
}

// ======================== IsAllLetter ========================

func TestIsAllLetter(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{"all_letters", "HelloWorld", true},
		{"single", "A", true},
		{"with_digit", "Hello123", false},
		{"with_space", "Hello World", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAllLetter(tt.str); got != tt.want {
				t.Fatalf("IsAllLetter(%q) = %v, want %v", tt.str, got, tt.want)
			}
		})
	}
}

// ======================== IsAllUpper ========================

func TestIsAllUpper(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{"all_upper", "HELLO", true},
		{"single_upper", "A", true},
		{"mixed", "Hello", false},
		{"with_digit", "HELLO123", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAllUpper(tt.str); got != tt.want {
				t.Fatalf("IsAllUpper(%q) = %v, want %v", tt.str, got, tt.want)
			}
		})
	}
}

// ======================== IsAllLower ========================

func TestIsAllLower(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{"all_lower", "hello", true},
		{"single_lower", "a", true},
		{"mixed", "Hello", false},
		{"with_digit", "hello123", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAllLower(tt.str); got != tt.want {
				t.Fatalf("IsAllLower(%q) = %v, want %v", tt.str, got, tt.want)
			}
		})
	}
}

// ======================== IsAllASCII ========================

func TestIsAllASCII(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{"ascii_text", "Hello World!", true},
		{"empty", "", true},
		{"contains_chinese", "Hello世界", false},
		{"emoji", "Hi😀", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAllASCII(tt.str); got != tt.want {
				t.Fatalf("IsAllASCII(%q) = %v, want %v", tt.str, got, tt.want)
			}
		})
	}
}

// ======================== IsPrintable ========================

func TestIsPrintable(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{"normal_text", "Hello World!", true},
		{"chinese", "你好世界", true},
		{"empty", "", true},
		{"newline", "Hello\nWorld", false},
		{"tab", "Hello\tWorld", false},
		{"carriage_return", "Hello\rWorld", false},
		{"null_char", "Hello\x00World", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPrintable(tt.str); got != tt.want {
				t.Fatalf("IsPrintable(%q) = %v, want %v", tt.str, got, tt.want)
			}
		})
	}
}

// ======================== IsContainUpper ========================

func TestIsContainUpper(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{"all_upper", "HELLO", true},
		{"mixed", "Hello", true},
		{"all_lower", "hello", false},
		{"digit_only", "123", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsContainUpper(tt.str); got != tt.want {
				t.Fatalf("IsContainUpper(%q) = %v, want %v", tt.str, got, tt.want)
			}
		})
	}
}

// ======================== IsContainLower ========================

func TestIsContainLower(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{"all_lower", "hello", true},
		{"mixed", "Hello", true},
		{"all_upper", "HELLO", false},
		{"digit_only", "123", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsContainLower(tt.str); got != tt.want {
				t.Fatalf("IsContainLower(%q) = %v, want %v", tt.str, got, tt.want)
			}
		})
	}
}

// ======================== IsBase64 ========================

func TestIsBase64(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid_std", "SGVsbG8=", true},
		{"valid_short", "aGVsbG8=", true},
		{"empty", "", false},
		{"invalid_char", "SGVsbG8@", false},
		{"no_padding_long", "SGVsbG9Xb3JsZA", false}, // 标准 base64 必须补齐到 4 的倍数
	}

	// 用标准库生成的合法 base64 补充测试
	for _, s := range []string{"hello", "Hello World!", "12345"} {
		encoded := base64.StdEncoding.EncodeToString([]byte(s))
		tests = append(tests, struct {
			name  string
			input string
			want  bool
		}{"stdlib_" + s, encoded, true})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBase64(tt.input); got != tt.want {
				t.Fatalf("IsBase64(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ======================== IsBase64URL ========================

func TestIsBase64URL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid", "SGVsbG8", true},
		{"valid_with_dash", "SGVsbG8-", true},
		{"valid_with_underscore", "SGVsbG8_", true},
		{"invalid_char", "SGVsbG8+", false},
		{"empty", "", true}, // 正则各部分均可选，空串匹配成功
	}

	for _, s := range []string{"", "hello", "Hello World!"} {
		encoded := base64.RawURLEncoding.EncodeToString([]byte(s))
		tests = append(tests, struct {
			name  string
			input string
			want  bool
		}{"stdlib_" + s, encoded, true})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBase64URL(tt.input); got != tt.want {
				t.Fatalf("IsBase64URL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ======================== IsHex ========================

func TestIsHex(t *testing.T) {
	tests := []struct {
		name string
		v    string
		want bool
	}{
		{"hex_lower", "1a2b3c", true},
		{"hex_upper", "1A2B3C", true},
		{"hex_mixed", "1a2B3c", true},
		{"with_hash", "#FF0000", true},
		{"with_0x", "0xFF0000", true},
		{"with_0X", "0XFF0000", true},
		{"invalid", "1g2h3i", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsHex(tt.v); got != tt.want {
				t.Fatalf("IsHex(%q) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

// ======================== IsJSON ========================

func TestIsJSON(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{"object", `{"key":"value"}`, true},
		{"array", `[1,2,3]`, true},
		{"number", `123`, true},
		{"string", `"hello"`, true},
		{"null", `null`, true},
		{"invalid", `{"key":}`, false},
		{"plain_text", `hello`, false},
		{"empty", ``, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsJSON(tt.str); got != tt.want {
				t.Fatalf("IsJSON(%q) = %v, want %v", tt.str, got, tt.want)
			}
		})
	}
}

// ======================== IsJWT ========================

func TestIsJWT(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"123"}`))
	sig := base64.RawURLEncoding.EncodeToString([]byte("secret"))
	validJWT := header + "." + payload + "." + sig

	tests := []struct {
		name string
		v    string
		want bool
	}{
		{"valid_jwt", validJWT, true},
		{"two_parts", "a.b", false},
		{"four_parts", "a.b.c.d", false},
		{"empty", "", false},
		{"no_dot", "abcdef", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsJWT(tt.v); got != tt.want {
				t.Fatalf("IsJWT(%q) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}
