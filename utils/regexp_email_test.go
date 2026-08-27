package utils

import (
	"testing"
)

func TestIsEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid_simple", "test@example.com", true},
		{"valid_subdomain", "test@mail.example.com", true},
		{"valid_uppercase", "TEST@EXAMPLE.COM", true},
		{"valid_mixed_case", "Test@Example.Com", true},
		{"no_at", "testexample.com", false},
		{"no_domain_extension", "test@example", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmail(tt.email); got != tt.want {
				t.Fatalf("IsEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}
