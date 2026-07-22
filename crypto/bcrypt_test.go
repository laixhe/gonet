package crypto

import (
	"strings"
	"testing"
)

func TestBcryptPasswordHash(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"normal password", "mypassword123", false},
		{"empty password", "", false},
		{"long password", strings.Repeat("a", 72), false},
		{"special chars", "密码!@#$%^&*()", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := BcryptPasswordHash(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("BcryptPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if hash == "" {
					t.Error("expected non-empty hash")
				}
				if !strings.HasPrefix(hash, "$2a$") {
					t.Errorf("hash should start with $2a$, got prefix: %s", hash[:8])
				}
			}
		})
	}
}

func TestBcryptPasswordCheck(t *testing.T) {
	password := "testpassword"
	hash, err := BcryptPasswordHash(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{"correct password", password, hash, true},
		{"wrong password", "wrongpassword", hash, false},
		{"empty password", "", hash, false},
		{"empty hash", password, "", false},
		{"both empty", "", "", false},
		{"invalid hash format", password, "invalid-hash", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BcryptPasswordCheck(tt.password, tt.hash); got != tt.want {
				t.Errorf("BcryptPasswordCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBcryptPasswordHash_UniqueHash(t *testing.T) {
	// 同一密码两次哈希结果应不同（因为随机盐）
	hash1, err := BcryptPasswordHash("samepassword")
	if err != nil {
		t.Fatalf("first hash failed: %v", err)
	}
	hash2, err := BcryptPasswordHash("samepassword")
	if err != nil {
		t.Fatalf("second hash failed: %v", err)
	}
	if hash1 == hash2 {
		t.Error("same password should produce different hashes due to random salt")
	}
	// 两个哈希都应该能通过校验
	if !BcryptPasswordCheck("samepassword", hash1) {
		t.Error("hash1 should pass check")
	}
	if !BcryptPasswordCheck("samepassword", hash2) {
		t.Error("hash2 should pass check")
	}
}

func TestBcryptPasswordHash_ExceedsMaxLength(t *testing.T) {
	// bcrypt 内部限制密码最大 72 字节，超过应报错
	_, err := BcryptPasswordHash(strings.Repeat("a", 73))
	if err == nil {
		t.Error("expected error for password exceeding 72 bytes")
	}
}

func TestBcryptPasswordCheck_TamperedHash(t *testing.T) {
	password := "mypassword"
	hash, err := BcryptPasswordHash(password)
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}

	// 篡改哈希值中间某个字符
	hashBytes := []byte(hash)
	if len(hashBytes) > 10 {
		hashBytes[10] ^= 0xFF
	}
	if BcryptPasswordCheck(password, string(hashBytes)) {
		t.Error("tampered hash should fail verification")
	}
}
