package utils

import (
	"fmt"
	"testing"
)

func TestIsChineseMobile(t *testing.T) {
	tests := []struct {
		name   string
		mobile string
		want   bool
	}{
		{"valid", "13800138000", true},
		{"valid_189", "18912345678", true},
		{"too_short", "1380013800", false},
		{"too_long_12_digits", "138001380000", true}, // 正则不含 ^$ 锚点，前11位即匹配成功
		{"empty", "", false},
		{"has_letter", "13800a38000", false},
		{"starts_with_2", "23800138000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsChineseMobile(tt.mobile)
			if got != tt.want {
				t.Fatalf("IsChineseMobile(%q) = %v, want %v", tt.mobile, got, tt.want)
			}
		})
	}
}

func TestObfuscateMobile(t *testing.T) {
	tests := []struct {
		name   string
		mobile string
		want   string
	}{
		{"normal_11_digit", "13800138000", "138****8000"},
		{"another_11_digit", "18912345678", "189****5678"},
		{"non_11_digit_short", "1380013800", "1380013800"},
		{"non_11_digit_long", "138001380000", "138001380000"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ObfuscateMobile(tt.mobile)
			if got != tt.want {
				t.Fatalf("ObfuscateMobile(%q) = %q, want %q", tt.mobile, got, tt.want)
			}
		})
	}
}

func TestReplaceObfuscateMobile(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "single_mobile",
			text: "请联系 13800138000",
			want: "请联系 138****8000",
		},
		{
			name: "multiple_mobiles",
			text: "13800138000 或 18912345678",
			want: "138****8000 或 189****5678",
		},
		{
			name: "no_mobile",
			text: "没有手机号",
			want: "没有手机号",
		},
		{
			name: "empty",
			text: "",
			want: "",
		},
		{
			name: "mobile_no_space",
			text: "13800138000",
			want: "138****8000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceObfuscateMobile(tt.text)
			if got != tt.want {
				t.Fatalf("ReplaceObfuscateMobile(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

// ExampleIsChineseMobile 演示 IsChineseMobile 的子串匹配行为
func ExampleIsChineseMobile() {
	fmt.Println(IsChineseMobile("13800138000"))       // 标准 11 位手机号
	fmt.Println(IsChineseMobile("abc13800138000xyz")) // 包含手机号的文本 — 注意：子串匹配，仍返回 true
	fmt.Println(IsChineseMobile("12345678901"))       // 以 1 开头的 11 位数字
	fmt.Println(IsChineseMobile("23800138000"))       // 不以 1 开头
	// Output:
	// true
	// true
	// true
	// false
}

// ExampleObfuscateMobile 演示手机号脱敏
func ExampleObfuscateMobile() {
	fmt.Println(ObfuscateMobile("13800138000"))
	fmt.Println(ObfuscateMobile("1380013800"))
	// Output:
	// 138****8000
	// 1380013800
}
