package jwt

import (
	"errors"
	"strings"
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// validConfig 返回一个合法的配置，方便各测试复用。
func validConfig() *Config {
	return &Config{
		SecretKey:     "test-secret-key-for-jwt-unit-test",
		ExpireTime:    3600,
		SigningMethod: SigningMethodHS256,
	}
}

// ======================== Config.Check 测试 ========================

func TestConfigCheck(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "没有JWT配置",
		},
		{
			name: "empty SecretKey",
			config: &Config{
				SecretKey:     "",
				ExpireTime:    3600,
				SigningMethod: SigningMethodHS256,
			},
			wantErr: true,
			errMsg:  "没有JWT密钥配置",
		},
		{
			name: "zero ExpireTime",
			config: &Config{
				SecretKey:     "secret",
				ExpireTime:    0,
				SigningMethod: SigningMethodHS256,
			},
			wantErr: true,
			errMsg:  "没有JWT过期时长配置",
		},
		{
			name: "negative ExpireTime",
			config: &Config{
				SecretKey:     "secret",
				ExpireTime:    -1,
				SigningMethod: SigningMethodHS256,
			},
			wantErr: true,
			errMsg:  "没有JWT过期时长配置",
		},
		{
			name: "invalid SigningMethod defaults to HS256",
			config: &Config{
				SecretKey:     "secret",
				ExpireTime:    3600,
				SigningMethod: "INVALID",
			},
			wantErr: false,
		},
		{
			name: "empty SigningMethod defaults to HS256",
			config: &Config{
				SecretKey:     "secret",
				ExpireTime:    3600,
				SigningMethod: "",
			},
			wantErr: false,
		},
		{
			name:    "valid HS256",
			config:  validConfig(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Check()
			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Check() error = %v, want contains %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestConfigCheck_AllValidSigningMethods(t *testing.T) {
	methods := []string{SigningMethodHS256, SigningMethodHS384, SigningMethodHS512}
	for _, m := range methods {
		c := &Config{
			SecretKey:     "secret",
			ExpireTime:    3600,
			SigningMethod: m,
		}
		if err := c.Check(); err != nil {
			t.Errorf("Check() with SigningMethod=%s should pass, got: %v", m, err)
		}
	}
}

// ======================== Config.JwtSigningMethod 测试 ========================

func TestJwtSigningMethod(t *testing.T) {
	tests := []struct {
		name          string
		signingMethod string
		wantAlg       string
	}{
		{"HS256", SigningMethodHS256, "HS256"},
		{"HS384", SigningMethodHS384, "HS384"},
		{"HS512", SigningMethodHS512, "HS512"},
		{"default (empty)", "", "HS256"},
		{"default (unsupported)", "UNKNOWN", "HS256"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{SigningMethod: tt.signingMethod}
			method := c.JwtSigningMethod()
			if method.Alg() != tt.wantAlg {
				t.Errorf("JwtSigningMethod() alg = %s, want %s", method.Alg(), tt.wantAlg)
			}
		})
	}
}

// ======================== GenToken + ParseToken 测试 ========================

func TestGenTokenAndParseToken(t *testing.T) {
	config := validConfig()
	claims := &CustomClaims{
		Uid: 123456,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(time.Duration(config.ExpireTime) * time.Second)),
			IssuedAt:  jwtv5.NewNumericDate(time.Now()),
			Issuer:    "gonet",
		},
	}

	tokenString, err := GenToken(config, claims)
	if err != nil {
		t.Fatalf("GenToken() failed: %v", err)
	}
	if tokenString == "" {
		t.Fatal("GenToken() returned empty string")
	}

	// 验证 JWT 格式 (三段 Base64URL)
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		t.Fatalf("JWT should have 3 parts, got %d", len(parts))
	}

	// 解析并验证
	parsedClaims := &CustomClaims{}
	parsedToken, err := ParseToken(config, tokenString, parsedClaims)
	if err != nil {
		t.Fatalf("ParseToken() failed: %v", err)
	}
	if !parsedToken.Valid {
		t.Error("parsed token should be valid")
	}
	if parsedClaims.Uid != 123456 {
		t.Errorf("parsed Uid = %d, want 123456", parsedClaims.Uid)
	}
	if parsedClaims.Issuer != "gonet" {
		t.Errorf("parsed Issuer = %s, want gonet", parsedClaims.Issuer)
	}
}

func TestGenToken_AllSigningMethods(t *testing.T) {
	methods := []string{SigningMethodHS256, SigningMethodHS384, SigningMethodHS512}
	baseConfig := &Config{
		SecretKey:  "test-key-for-signing-methods",
		ExpireTime: 3600,
	}

	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			config := &Config{
				SecretKey:     baseConfig.SecretKey,
				ExpireTime:    baseConfig.ExpireTime,
				SigningMethod: m,
			}
			claims := &CustomClaims{
				Uid: 100,
				RegisteredClaims: jwtv5.RegisteredClaims{
					ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(time.Hour)),
				},
			}

			tokenString, err := GenToken(config, claims)
			if err != nil {
				t.Fatalf("GenToken() with %s failed: %v", m, err)
			}

			parsedClaims := &CustomClaims{}
			_, err = ParseToken(config, tokenString, parsedClaims)
			if err != nil {
				t.Fatalf("ParseToken() with %s failed: %v", m, err)
			}
			if parsedClaims.Uid != 100 {
				t.Errorf("Uid mismatch for %s", m)
			}
		})
	}
}

// ======================== ParseToken 错误场景测试 ========================

func TestParseToken_ExpiredToken(t *testing.T) {
	config := validConfig()
	// 生成一个已过期的 token
	claims := &CustomClaims{
		Uid: 999,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(-1 * time.Hour)), // 1 小时前过期
		},
	}

	tokenString, err := GenToken(config, claims)
	if err != nil {
		t.Fatalf("GenToken() failed: %v", err)
	}

	_, err = ParseToken(config, tokenString, &CustomClaims{})
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if !errors.Is(err, ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired, got: %v", err)
	}
}

func TestParseToken_WrongSecretKey(t *testing.T) {
	config := validConfig()
	claims := &CustomClaims{
		Uid: 1,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	tokenString, err := GenToken(config, claims)
	if err != nil {
		t.Fatalf("GenToken() failed: %v", err)
	}

	// 用不同的密钥解析
	wrongConfig := &Config{
		SecretKey:     "wrong-secret-key-!@#$",
		ExpireTime:    3600,
		SigningMethod: SigningMethodHS256,
	}
	_, err = ParseToken(wrongConfig, tokenString, &CustomClaims{})
	if err == nil {
		t.Fatal("expected error with wrong secret key")
	}
	if !errors.Is(err, ErrTokenInvalid) {
		t.Errorf("expected ErrTokenInvalid, got: %v", err)
	}
}

func TestParseToken_TamperedToken(t *testing.T) {
	config := validConfig()
	claims := &CustomClaims{
		Uid: 1,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	tokenString, err := GenToken(config, claims)
	if err != nil {
		t.Fatalf("GenToken() failed: %v", err)
	}

	// 篡改 payload 部分（第二段）的最后一个字符
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		t.Fatal("invalid JWT format")
	}
	payloadBytes := []byte(parts[1])
	if len(payloadBytes) > 0 {
		payloadBytes[len(payloadBytes)-1] ^= 0xFF
	}
	parts[1] = string(payloadBytes)
	tamperedToken := strings.Join(parts, ".")

	_, err = ParseToken(config, tamperedToken, &CustomClaims{})
	if err == nil {
		t.Fatal("expected error with tampered token")
	}
	if !errors.Is(err, ErrTokenInvalid) {
		t.Errorf("expected ErrTokenInvalid, got: %v", err)
	}
}

func TestParseToken_InvalidFormat(t *testing.T) {
	config := validConfig()

	tests := []struct {
		name        string
		tokenString string
	}{
		{"empty string", ""},
		{"not a jwt", "not-a-jwt-token"},
		{"single part", "header"},
		{"two parts", "header.payload"},
		{"four parts", "a.b.c.d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseToken(config, tt.tokenString, &CustomClaims{})
			if err == nil {
				t.Errorf("expected error for token: %q", tt.tokenString)
			}
			if !errors.Is(err, ErrTokenInvalid) {
				t.Errorf("expected ErrTokenInvalid, got: %v", err)
			}
		})
	}
}

func TestParseToken_NilConfig(t *testing.T) {
	// 生成一个合法 token
	config := validConfig()
	claims := &CustomClaims{
		Uid: 1,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tokenString, _ := GenToken(config, claims)

	// 用 nil config 解析应触发 panic 或 error
	defer func() {
		if r := recover(); r != nil {
			t.Logf("ParseToken with nil config panicked as expected: %v", r)
		}
	}()
	_, err := ParseToken(nil, tokenString, &CustomClaims{})
	if err != nil {
		t.Logf("ParseToken with nil config returned error: %v", err)
	}
}

// ======================== GenToken 边界测试 ========================

func TestGenToken_EmptyClaimsNoStandardFields(t *testing.T) {
	config := validConfig()
	// 注册声明全部空白
	claims := &CustomClaims{
		Uid: 0,
	}

	tokenString, err := GenToken(config, claims)
	if err != nil {
		t.Fatalf("GenToken() with empty claims failed: %v", err)
	}

	parsedClaims := &CustomClaims{}
	_, err = ParseToken(config, tokenString, parsedClaims)
	if err != nil {
		t.Fatalf("ParseToken() failed: %v", err)
	}
	if parsedClaims.Uid != 0 {
		t.Errorf("Uid = %d, want 0", parsedClaims.Uid)
	}
}

func TestGenToken_LongSecretKey(t *testing.T) {
	config := &Config{
		SecretKey:     strings.Repeat("x", 512),
		ExpireTime:    3600,
		SigningMethod: SigningMethodHS512,
	}
	claims := &CustomClaims{
		Uid: 42,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	tokenString, err := GenToken(config, claims)
	if err != nil {
		t.Fatalf("GenToken() with long key failed: %v", err)
	}

	parsedClaims := &CustomClaims{}
	_, err = ParseToken(config, tokenString, parsedClaims)
	if err != nil {
		t.Fatalf("ParseToken() with long key failed: %v", err)
	}
	if parsedClaims.Uid != 42 {
		t.Errorf("Uid = %d, want 42", parsedClaims.Uid)
	}
}

// ======================== CustomClaims 字段透传测试 ========================

func TestCustomClaims_AllStandardFields(t *testing.T) {
	config := validConfig()
	now := time.Now()
	exp := now.Add(time.Hour)
	nbf := now.Add(-time.Minute)

	claims := &CustomClaims{
		Uid: 777,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(exp),
			IssuedAt:  jwtv5.NewNumericDate(now),
			NotBefore: jwtv5.NewNumericDate(nbf),
			Issuer:    "gonet-test",
			Subject:   "test-subject",
			Audience:  jwtv5.ClaimStrings{"aud1", "aud2"},
			ID:        "jti-12345",
		},
	}

	tokenString, err := GenToken(config, claims)
	if err != nil {
		t.Fatalf("GenToken() failed: %v", err)
	}

	parsedClaims := &CustomClaims{}
	_, err = ParseToken(config, tokenString, parsedClaims)
	if err != nil {
		t.Fatalf("ParseToken() failed: %v", err)
	}

	if parsedClaims.Uid != 777 {
		t.Errorf("Uid = %d, want 777", parsedClaims.Uid)
	}
	if parsedClaims.Issuer != "gonet-test" {
		t.Errorf("Issuer = %s, want gonet-test", parsedClaims.Issuer)
	}
	if parsedClaims.Subject != "test-subject" {
		t.Errorf("Subject = %s, want test-subject", parsedClaims.Subject)
	}
	if parsedClaims.ID != "jti-12345" {
		t.Errorf("ID = %s, want jti-12345", parsedClaims.ID)
	}
	if len(parsedClaims.Audience) != 2 {
		t.Errorf("Audience length = %d, want 2", len(parsedClaims.Audience))
	}
}

// ======================== 常量测试 ========================

func TestConstants(t *testing.T) {
	if Authorization != "Authorization" {
		t.Errorf("Authorization = %s, want Authorization", Authorization)
	}
	if Bearer != "Bearer " {
		t.Errorf("Bearer = %q, want %q", Bearer, "Bearer ")
	}
	if BearerLen != 7 {
		t.Errorf("BearerLen = %d, want 7", BearerLen)
	}
	if AuthorizationClaims != "AuthorizationClaims" {
		t.Errorf("AuthorizationClaims = %s, want AuthorizationClaims", AuthorizationClaims)
	}
	if SigningMethodHS256 != "HS256" {
		t.Errorf("SigningMethodHS256 = %s, want HS256", SigningMethodHS256)
	}
	if SigningMethodHS384 != "HS384" {
		t.Errorf("SigningMethodHS384 = %s, want HS384", SigningMethodHS384)
	}
	if SigningMethodHS512 != "HS512" {
		t.Errorf("SigningMethodHS512 = %s, want HS512", SigningMethodHS512)
	}
}

// ======================== 错误类型测试 ========================

func TestErrorSentinelValues(t *testing.T) {
	// 验证 ErrTokenExpired 和 ErrTokenInvalid 是独立的 sentinel error
	if ErrTokenExpired == ErrTokenInvalid {
		t.Error("ErrTokenExpired and ErrTokenInvalid should be different errors")
	}
	if ErrTokenExpired.Error() != "token is expired" {
		t.Errorf("ErrTokenExpired message = %s, want 'token is expired'", ErrTokenExpired.Error())
	}
	if ErrTokenInvalid.Error() != "token invalid" {
		t.Errorf("ErrTokenInvalid message = %s, want 'token invalid'", ErrTokenInvalid.Error())
	}
}

func TestParseTokenErrorUnwrap(t *testing.T) {
	config := validConfig()
	claims := &CustomClaims{
		Uid: 1,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tokenString, _ := GenToken(config, claims)

	wrongConfig := &Config{
		SecretKey:     "wrong-key",
		ExpireTime:    3600,
		SigningMethod: SigningMethodHS256,
	}
	_, err := ParseToken(wrongConfig, tokenString, &CustomClaims{})

	// ErrTokenInvalid 应包含在错误链中
	if !errors.Is(err, ErrTokenInvalid) {
		t.Error("expected ErrTokenInvalid in error chain")
	}
	// ErrTokenExpired 不应出现在这个错误链中
	if errors.Is(err, ErrTokenExpired) {
		t.Error("should not have ErrTokenExpired for wrong key")
	}
	// 原始 jwtv5 错误应可在错误链中找到
	if !errors.Is(err, jwtv5.ErrSignatureInvalid) {
		t.Error("expected jwtv5.ErrSignatureInvalid in error chain")
	}
}
