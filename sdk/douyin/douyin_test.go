package douyin

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"testing"
)

// TestRsaDecryptPKCS1 验证 PKCS#1 私钥(openssl genrsa 产物)解密
func TestRsaDecryptPKCS1(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	privBase64 := base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(key))

	plain := `{"phoneNumber":"13800138000","purePhoneNumber":"13800138000"}`
	cipher, err := rsa.EncryptPKCS1v15(rand.Reader, &key.PublicKey, []byte(plain))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	d := &Douyin{config: &Config{PrivateKey: privBase64}}
	got, err := d.RsaDecryptByPrivateKeyStr(base64.StdEncoding.EncodeToString(cipher))
	if err != nil {
		t.Fatalf("decrypt pkcs1: %v", err)
	}
	if got != plain {
		t.Fatalf("unexpected plaintext: %q", got)
	}
}

// TestRsaDecryptPKCS8 验证 PKCS#8 私钥(官方 Java 示例格式)解密
func TestRsaDecryptPKCS8(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal pkcs8: %v", err)
	}
	privBase64 := base64.StdEncoding.EncodeToString(der)

	plain := `{"phoneNumber":"13800138000","purePhoneNumber":"13800138000"}`
	cipher, err := rsa.EncryptPKCS1v15(rand.Reader, &key.PublicKey, []byte(plain))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	d := &Douyin{config: &Config{PrivateKey: privBase64}}
	got, err := d.RsaDecryptByPrivateKeyStr(base64.StdEncoding.EncodeToString(cipher))
	if err != nil {
		t.Fatalf("decrypt pkcs8: %v", err)
	}
	if got != plain {
		t.Fatalf("unexpected plaintext: %q", got)
	}
}

func TestRsaDecryptInvalidBase64Key(t *testing.T) {
	d := &Douyin{config: &Config{PrivateKey: "not-a-valid-base64!!!"}}
	if _, err := d.RsaDecryptByPrivateKeyStr("abc"); err == nil {
		t.Fatal("expected error for invalid base64 key")
	}
}

func TestRsaDecryptNonKeyData(t *testing.T) {
	d := &Douyin{config: &Config{PrivateKey: base64.StdEncoding.EncodeToString([]byte("hello"))}}
	if _, err := d.RsaDecryptByPrivateKeyStr("abc"); err == nil {
		t.Fatal("expected error for non-key data")
	}
}
