package utils

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestRandBool(t *testing.T) {
	trues, falses := 0, 0
	for i := 0; i < 100; i++ {
		if RandBool() {
			trues++
		} else {
			falses++
		}
	}
	if trues == 0 || falses == 0 {
		t.Fatalf("RandBool distribution suspicious: true=%d, false=%d", trues, falses)
	}
}

func TestRandRange(t *testing.T) {
	tests := []struct {
		name string
		min  int
		max  int
	}{
		{"zero_to_five", 0, 5},
		{"equal", 3, 3},
		{"swapped", 5, 2},
		{"negative", -10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 20; i++ {
				v := RandRange(tt.min, tt.max)
				wantMin, wantMax := tt.min, tt.max
				if wantMax < wantMin {
					wantMin, wantMax = wantMax, wantMin
				}
				if v < wantMin || v > wantMax {
					t.Fatalf("RandRange(%d,%d) = %d, out of [%d,%d]", tt.min, tt.max, v, wantMin, wantMax)
				}
			}
		})
	}
}

func TestRandBytes(t *testing.T) {
	tests := []int{0, 1, 4, 16, 64}
	for _, n := range tests {
		data := RandBytes(n)
		if len(data) != n {
			t.Fatalf("RandBytes(%d) length = %d", n, len(data))
		}
		if n >= 8 {
			data2 := RandBytes(n)
			if string(data) == string(data2) {
				t.Log("warning: two RandBytes calls returned same data (extremely unlikely but not impossible)")
			}
		}
	}
	// 验证十六进制输出格式
	data := RandBytes(4)
	hexStr := hex.EncodeToString(data)
	if len(hexStr) != 8 {
		t.Fatalf("hex length expected 8, got %d", len(hexStr))
	}
}

func TestRandNumeral(t *testing.T) {
	tests := []int{0, 1, 5, 10}
	for _, n := range tests {
		s := RandNumeral(n)
		if len(s) != n {
			t.Fatalf("RandNumeral(%d) length = %d", n, len(s))
		}
		for _, c := range s {
			if c < '0' || c > '9' {
				t.Fatalf("RandNumeral contains non-digit: %c", c)
			}
		}
	}
}

func TestRandString(t *testing.T) {
	// 混合大小写 + 数字
	s1 := RandString(20)
	if len(s1) != 20 {
		t.Fatalf("RandString(20) length = %d", len(s1))
	}
	hasDigit := strings.ContainsAny(s1, "0123456789")
	hasUpper := strings.ContainsAny(s1, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLower := strings.ContainsAny(s1, "abcdefghijklmnopqrstuvwxyz")
	if !hasDigit || !hasUpper || !hasLower {
		t.Logf("RandString mixed: digit=%v upper=%v lower=%v (may be rare)", hasDigit, hasUpper, hasLower)
	}

	// 大写 + 数字
	s2 := RandString(20, true)
	if len(s2) != 20 {
		t.Fatalf("RandString(20, true) length = %d", len(s2))
	}
	for _, c := range s2 {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			t.Fatalf("RandString(20, true) contains unexpected char: %c", c)
		}
	}

	// 小写 + 数字
	s3 := RandString(20, false)
	if len(s3) != 20 {
		t.Fatalf("RandString(20, false) length = %d", len(s3))
	}
	for _, c := range s3 {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			t.Fatalf("RandString(20, false) contains unexpected char: %c", c)
		}
	}

	// 零长度
	s0 := RandString(0)
	if s0 != "" {
		t.Fatalf("RandString(0) expected empty, got %q", s0)
	}
}

func TestRandLetter(t *testing.T) {
	// 混合大小写
	s1 := RandLetter(20)
	if len(s1) != 20 {
		t.Fatalf("RandLetter(20) length = %d", len(s1))
	}

	// 大写
	s2 := RandLetter(20, true)
	if len(s2) != 20 {
		t.Fatalf("RandLetter(20, true) length = %d", len(s2))
	}
	for _, c := range s2 {
		if c < 'A' || c > 'Z' {
			t.Fatalf("RandLetter(20, true) contains non-uppercase: %c", c)
		}
	}

	// 小写
	s3 := RandLetter(20, false)
	if len(s3) != 20 {
		t.Fatalf("RandLetter(20, false) length = %d", len(s3))
	}
	for _, c := range s3 {
		if c < 'a' || c > 'z' {
			t.Fatalf("RandLetter(20, false) contains non-lowercase: %c", c)
		}
	}

	// 零长度
	s0 := RandLetter(0)
	if s0 != "" {
		t.Fatalf("RandLetter(0) expected empty, got %q", s0)
	}
}
