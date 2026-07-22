package utils

import (
	"reflect"
	"testing"
)

// ======================== IsContainChinese ========================

func TestIsContainChinese(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"pure_chinese", "你好世界", true},
		{"mixed", "hello世界", true},
		{"no_chinese", "hello world", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsContainChinese(tt.s); got != tt.want {
				t.Fatalf("IsContainChinese(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

// ======================== ExtractChineseCharacters ========================

func TestExtractChineseCharacters(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want []string
	}{
		{"mixed_text", "hello你好world世界", []string{"你好", "世界"}},
		{"continuous_chinese", "你好世界", []string{"你好世界"}},
		{"no_chinese", "hello world", nil},
		{"empty", "", nil},
		{"chinese_split_by_english", "你好a世界", []string{"你好", "世界"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractChineseCharacters(tt.s)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ExtractChineseCharacters(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

// ======================== ExtractEachChineseCharacters ========================

func TestExtractEachChineseCharacters(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want []string
	}{
		{"mixed_text", "hello你好", []string{"你", "好"}},
		{"chinese_only", "你好", []string{"你", "好"}},
		{"no_chinese", "hello", nil},
		{"empty", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractEachChineseCharacters(tt.s)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ExtractEachChineseCharacters(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

// ======================== ExtractNonChineseCharacters ========================

func TestExtractNonChineseCharacters(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want []string
	}{
		{"mixed_text", "hello你好world世界", []string{"hello", "world"}},
		{"non_chinese_only", "hello world", []string{"hello world"}},
		{"chinese_only", "你好世界", nil},
		{"empty", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractNonChineseCharacters(tt.s)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ExtractNonChineseCharacters(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

// ======================== ExtractNonEachChineseCharacters ========================

func TestExtractNonEachChineseCharacters(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want []string
	}{
		{"mixed_text", "hello你好", []string{"h", "e", "l", "l", "o"}},
		{"non_chinese_only", "abc", []string{"a", "b", "c"}},
		{"chinese_only", "你好", nil},
		{"empty", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractNonEachChineseCharacters(tt.s)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ExtractNonEachChineseCharacters(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}
