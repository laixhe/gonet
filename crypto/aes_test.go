package crypto

import (
	"encoding/base64"
	"strings"
	"testing"
)

// ======================== AES GCM 测试 ========================

func TestAesEncryptDecryptGCM(t *testing.T) {
	keySizes := []struct {
		name string
		key  []byte
	}{
		{"AES-128 (16 bytes)", []byte("1234567890123456")},
		{"AES-192 (24 bytes)", []byte("123456789012345678901234")},
		{"AES-256 (32 bytes)", []byte("12345678901234567890123456789012")},
	}

	plainTexts := []struct {
		name string
		data []byte
	}{
		{"normal text", []byte("hello world")},
		{"empty data", []byte{}},
		{"binary data", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}},
		{"long text", []byte(strings.Repeat("a", 1024))},
	}

	for _, ks := range keySizes {
		for _, pt := range plainTexts {
			t.Run(ks.name+"/"+pt.name, func(t *testing.T) {
				cipherText, nonce, err := AesEncryptGCM(ks.key, pt.data)
				if err != nil {
					t.Fatalf("encrypt failed: %v", err)
				}

				decrypted, err := AesDecryptGCM(ks.key, cipherText, nonce)
				if err != nil {
					t.Fatalf("decrypt failed: %v", err)
				}

				if string(decrypted) != string(pt.data) {
					t.Errorf("roundtrip mismatch: got %v, want %v", decrypted, pt.data)
				}
			})
		}
	}
}

func TestAesGCM_InvalidKey(t *testing.T) {
	invalidKeys := [][]byte{
		{},              // empty
		[]byte("123"),   // too short
		[]byte("12345"), // 5 bytes
	}

	for _, key := range invalidKeys {
		_, _, err := AesEncryptGCM(key, []byte("test"))
		if err == nil {
			t.Errorf("expected error for key length %d", len(key))
		}

		_, err = AesDecryptGCM(key, []byte("test"), make([]byte, 12))
		if err == nil {
			t.Errorf("expected error for key length %d in decrypt", len(key))
		}
	}
}

func TestAesGCM_WrongNonce(t *testing.T) {
	key := []byte("1234567890123456")
	plainText := []byte("secret message")

	cipherText, _, err := AesEncryptGCM(key, plainText)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	wrongNonce := make([]byte, 12)
	_, err = AesDecryptGCM(key, cipherText, wrongNonce)
	if err == nil {
		t.Error("expected error with wrong nonce")
	}
}

func TestAesGCM_TamperedCiphertext(t *testing.T) {
	key := []byte("1234567890123456")
	plainText := []byte("secret message")

	cipherText, nonce, err := AesEncryptGCM(key, plainText)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// 篡改密文
	tampered := make([]byte, len(cipherText))
	copy(tampered, cipherText)
	if len(tampered) > 0 {
		tampered[0] ^= 0xFF
	}

	_, err = AesDecryptGCM(key, tampered, nonce)
	if err == nil {
		t.Error("expected error with tampered ciphertext (GCM authentication should fail)")
	}
}

func TestAesGCM_NonceUniqueness(t *testing.T) {
	key := []byte("1234567890123456")
	plainText := []byte("hello")

	cipherText1, nonce1, err := AesEncryptGCM(key, plainText)
	if err != nil {
		t.Fatalf("first encrypt failed: %v", err)
	}
	cipherText2, nonce2, err := AesEncryptGCM(key, plainText)
	if err != nil {
		t.Fatalf("second encrypt failed: %v", err)
	}

	// 两次加密 nonce 应不同
	if string(nonce1) == string(nonce2) {
		t.Error("nonces should be different for each encryption")
	}
	// 两次加密密文应不同（因为 nonce 不同）
	if string(cipherText1) == string(cipherText2) {
		t.Error("ciphertexts should be different with different nonces")
	}
	// 但都应该能正确解密
	d1, _ := AesDecryptGCM(key, cipherText1, nonce1)
	d2, _ := AesDecryptGCM(key, cipherText2, nonce2)
	if string(d1) != string(plainText) || string(d2) != string(plainText) {
		t.Error("both should decrypt to the original plaintext")
	}
}

func TestAesGCM_CiphertextDiffers(t *testing.T) {
	key := []byte("1234567890123456")
	plainText := []byte("hello world")

	cipherText, nonce, err := AesEncryptGCM(key, plainText)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// 密文应与原文不同
	if string(cipherText) == string(plainText) {
		t.Error("ciphertext should differ from plaintext")
	}
	// nonce 应非空
	if len(nonce) == 0 {
		t.Error("nonce should not be empty")
	}
}

// ======================== AES CTR 测试 ========================

func TestAesEncryptDecryptCTR(t *testing.T) {
	key := []byte("1234567890123456")
	plainTexts := []struct {
		name string
		data []byte
	}{
		{"normal text", []byte("hello world")},
		{"empty data", []byte{}},
		{"binary data", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}},
		{"long text", []byte(strings.Repeat("b", 1024))},
	}

	for _, pt := range plainTexts {
		t.Run(pt.name, func(t *testing.T) {
			cipherText, iv, err := AesEncryptCTR(key, pt.data)
			if err != nil {
				t.Fatalf("encrypt failed: %v", err)
			}

			decrypted, err := AesDecryptCTR(key, cipherText, iv)
			if err != nil {
				t.Fatalf("decrypt failed: %v", err)
			}

			if string(decrypted) != string(pt.data) {
				t.Errorf("roundtrip mismatch: got %v, want %v", decrypted, pt.data)
			}
		})
	}
}

func TestAesCTR_InvalidKey(t *testing.T) {
	_, _, err := AesEncryptCTR([]byte("bad"), []byte("test"))
	if err == nil {
		t.Error("expected error for invalid key")
	}
}

func TestAesDecryptCTR_InvalidKey(t *testing.T) {
	_, err := AesDecryptCTR([]byte("bad"), []byte("test"), make([]byte, 16))
	if err == nil {
		t.Error("expected error for invalid key in AesDecryptCTR")
	}
}

func TestAesCTR_WrongIV(t *testing.T) {
	key := []byte("1234567890123456")
	plainText := []byte("secret message")

	cipherText, _, err := AesEncryptCTR(key, plainText)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	wrongIV := make([]byte, 16)
	decrypted, err := AesDecryptCTR(key, cipherText, wrongIV)
	if err != nil {
		t.Fatalf("decrypt should not error with wrong IV: %v", err)
	}
	// 错误 IV 解出来的明文不应等于原始明文
	if string(decrypted) == string(plainText) {
		t.Error("wrong IV should produce different plaintext")
	}
}

func TestAesCTR_CiphertextDiffers(t *testing.T) {
	key := []byte("1234567890123456")
	plainText := []byte("hello world")

	cipherText, iv, err := AesEncryptCTR(key, plainText)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	if string(cipherText) == string(plainText) {
		t.Error("ciphertext should differ from plaintext")
	}
	if len(iv) == 0 {
		t.Error("IV should not be empty")
	}
	if len(cipherText) != len(plainText) {
		t.Errorf("CTR ciphertext length = %d, want %d", len(cipherText), len(plainText))
	}
}

func TestAesCTR_KeySizes(t *testing.T) {
	keys := [][]byte{
		[]byte("123456789012345678901234"),         // AES-192
		[]byte("12345678901234567890123456789012"), // AES-256
	}
	plainText := []byte("test message")

	for i, key := range keys {
		cipherText, iv, err := AesEncryptCTR(key, plainText)
		if err != nil {
			t.Errorf("AES-CTR encrypt with %d-byte key failed: %v", len(key), err)
			continue
		}
		decrypted, err := AesDecryptCTR(key, cipherText, iv)
		if err != nil {
			t.Errorf("AES-CTR decrypt with %d-byte key failed: %v", len(key), err)
			continue
		}
		if string(decrypted) != string(plainText) {
			t.Errorf("AES-CTR roundtrip mismatch with %d-byte key", len(key))
		}
		_ = i
	}
}

// ======================== AES CBC 测试 ========================

func TestAesEncryptDecryptCBC(t *testing.T) {
	key := []byte("1234567890123456")
	plainTexts := []struct {
		name string
		data []byte
	}{
		{"normal text", []byte("hello world")},
		{"empty data", []byte{}},
		{"exact block size (16 bytes)", []byte("1234567890123456")},
		{"binary data", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}},
		{"less than block", []byte("hi")},
		{"long text", []byte(strings.Repeat("c", 1024))},
	}

	for _, pt := range plainTexts {
		t.Run(pt.name, func(t *testing.T) {
			cipherText, iv, err := AesEncryptCBC(key, pt.data)
			if err != nil {
				t.Fatalf("encrypt failed: %v", err)
			}

			decrypted, err := AesDecryptCBC(key, cipherText, iv)
			if err != nil {
				t.Fatalf("decrypt failed: %v", err)
			}

			if string(decrypted) != string(pt.data) {
				t.Errorf("roundtrip mismatch: got %v, want %v", decrypted, pt.data)
			}
		})
	}
}

func TestAesCBC_InvalidKey(t *testing.T) {
	_, _, err := AesEncryptCBC([]byte("bad"), []byte("test"))
	if err == nil {
		t.Error("expected error for invalid key")
	}
}

func TestAesDecryptCBC_InvalidKey(t *testing.T) {
	_, err := AesDecryptCBC([]byte("bad"), []byte("test"), make([]byte, 16))
	if err == nil {
		t.Error("expected error for invalid key in AesDecryptCBC")
	}
}

func TestAesCBC_WrongIV(t *testing.T) {
	key := []byte("1234567890123456")
	plainText := []byte("secret message")

	cipherText, _, err := AesEncryptCBC(key, plainText)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	wrongIV := make([]byte, 16)
	decrypted, err := AesDecryptCBC(key, cipherText, wrongIV)
	if err != nil {
		t.Fatalf("decrypt should not error with wrong IV: %v", err)
	}
	if string(decrypted) == string(plainText) {
		t.Error("wrong IV should produce different plaintext")
	}
}

func TestAesCBC_CiphertextDiffers(t *testing.T) {
	key := []byte("1234567890123456")
	plainText := []byte("hello world")

	cipherText, iv, err := AesEncryptCBC(key, plainText)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	if string(cipherText) == string(plainText) {
		t.Error("ciphertext should differ from plaintext")
	}
	if len(iv) == 0 {
		t.Error("IV should not be empty")
	}
	// CBC 密文长度应 >= 原文（PKCS7 填充）
	if len(cipherText) < len(plainText) {
		t.Errorf("CBC ciphertext length = %d, plaintext = %d, ciphertext should be >= plaintext due to padding", len(cipherText), len(plainText))
	}
}

// ======================== PKCS7 填充测试 ========================

func TestPKCS7Padding(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		blockSize int
		wantLen   int
	}{
		{"needs padding", []byte("hello"), 8, 8},
		{"exact block", []byte("12345678"), 8, 16},
		{"empty data", []byte{}, 8, 8},
		{"large block size", []byte("hi"), 16, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PKCS7Padding(tt.data, tt.blockSize)
			if len(result) != tt.wantLen {
				t.Errorf("PKCS7Padding() length = %v, want %v", len(result), tt.wantLen)
			}
		})
	}
}

func TestPKCS7PaddingUnPadding_Roundtrip(t *testing.T) {
	blockSizes := []int{8, 16, 32}
	dataSizes := []int{0, 1, 7, 8, 9, 15, 16, 17, 31, 32, 33, 100}

	for _, bs := range blockSizes {
		for _, ds := range dataSizes {
			data := make([]byte, ds)
			for i := range data {
				data[i] = byte(i % 256)
			}
			padded := PKCS7Padding(data, bs)
			unpadded := PKCS7UnPadding(padded)
			if string(unpadded) != string(data) {
				t.Errorf("PKCS7 roundtrip failed: blockSize=%d, dataSize=%d", bs, ds)
			}
		}
	}
}

func TestPKCS7UnPadding_Invalid(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []byte
	}{
		{"empty input", []byte{}, []byte{}},
		{"zero padding value", []byte{0x01, 0x02, 0x00}, []byte{0x01, 0x02, 0x00}},
		{"padding larger than length", []byte{0x05, 0x05, 0x05}, []byte{0x05, 0x05, 0x05}},
		{"inconsistent padding bytes", []byte{0x01, 0x02, 0x03}, []byte{0x01, 0x02, 0x03}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PKCS7UnPadding(tt.data)
			if string(result) != string(tt.want) {
				t.Errorf("PKCS7UnPadding() = %v, want %v", result, tt.want)
			}
		})
	}
}

// ======================== AES GCM String 版测试 ========================

func TestAesEncryptDecryptGCMString(t *testing.T) {
	keySizes := []struct {
		name string
		key  string
	}{
		{"AES-128", "1234567890123456"},
		{"AES-192", "123456789012345678901234"},
		{"AES-256", "12345678901234567890123456789012"},
	}

	plainTexts := []string{
		"hello world",
		"",
		"密码测试中文",
		strings.Repeat("x", 1000),
	}

	for _, ks := range keySizes {
		for _, pt := range plainTexts {
			t.Run(ks.name+"/"+pt, func(t *testing.T) {
				cipherBase64, err := AesEncryptGCMString(ks.key, pt)
				if err != nil {
					t.Fatalf("encrypt failed: %v", err)
				}

				// 验证是有效 base64
				_, err = base64.StdEncoding.DecodeString(cipherBase64)
				if err != nil {
					t.Fatalf("output is not valid base64: %v", err)
				}

				decrypted, err := AesDecryptGCMString(ks.key, cipherBase64)
				if err != nil {
					t.Fatalf("decrypt failed: %v", err)
				}

				if decrypted != pt {
					t.Errorf("roundtrip mismatch: got %v, want %v", decrypted, pt)
				}
			})
		}
	}
}

func TestAesDecryptGCMString_InvalidBase64(t *testing.T) {
	_, err := AesDecryptGCMString("1234567890123456", "!!!not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestAesDecryptGCMString_DataTooShort(t *testing.T) {
	key := "1234567890123456"
	// 12 字节 nonce 是 GCM 的最小值，数据长度 < 12 应报错
	shortData := base64.StdEncoding.EncodeToString([]byte{0x01, 0x02})
	_, err := AesDecryptGCMString(key, shortData)
	if err == nil {
		t.Error("expected error for data too short")
	}
}

func TestAesDecryptGCMString_WrongKey(t *testing.T) {
	key1 := "1234567890123456"
	key2 := "6543210987654321"
	plainText := "secret message"

	cipherBase64, err := AesEncryptGCMString(key1, plainText)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	_, err = AesDecryptGCMString(key2, cipherBase64)
	if err == nil {
		t.Error("expected error with wrong key")
	}
}

func TestAesEncryptGCMString_InvalidKey(t *testing.T) {
	_, err := AesEncryptGCMString("bad", "test")
	if err == nil {
		t.Error("expected error for invalid key")
	}
}

func TestAesDecryptGCMString_InvalidKeyLength(t *testing.T) {
	key := "1234567890123456"
	cipherBase64, _ := AesEncryptGCMString(key, "test")

	// 使用无效长度的密钥解密有效 base64 数据
	_, err := AesDecryptGCMString("bad", cipherBase64)
	if err == nil {
		t.Error("expected error for invalid key length in AesDecryptGCMString")
	}
}

func TestAesDecryptGCMString_TamperedData(t *testing.T) {
	key := "1234567890123456"
	cipherBase64, err := AesEncryptGCMString(key, "secret message")
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// 篡改 base64 数据中间某个字符
	raw, _ := base64.StdEncoding.DecodeString(cipherBase64)
	if len(raw) > 12 {
		raw[12] ^= 0xFF // 篡改密文部分（跳过 nonce）
	}
	tamperedBase64 := base64.StdEncoding.EncodeToString(raw)

	_, err = AesDecryptGCMString(key, tamperedBase64)
	if err == nil {
		t.Error("expected error with tampered data (GCM authentication should fail)")
	}
}

func TestAesEncryptDecryptGCMString_NonceUniqueness(t *testing.T) {
	key := "1234567890123456"
	c1, _ := AesEncryptGCMString(key, "hello")
	c2, _ := AesEncryptGCMString(key, "hello")
	if c1 == c2 {
		t.Error("two encryptions should produce different ciphertexts due to random nonce")
	}
}

// ======================== AES CBC String 版测试 ========================

func TestAesEncryptDecryptCBCString(t *testing.T) {
	key := "1234567890123456"
	plainTexts := []string{
		"hello world",
		"",
		"密码测试中文",
		"exact block 16bytes", // exact block size
		strings.Repeat("y", 500),
	}

	for _, pt := range plainTexts {
		t.Run(pt, func(t *testing.T) {
			cipherBase64, err := AesEncryptCBCString(key, pt)
			if err != nil {
				t.Fatalf("encrypt failed: %v", err)
			}

			// 验证是有效 base64
			_, err = base64.StdEncoding.DecodeString(cipherBase64)
			if err != nil {
				t.Fatalf("output is not valid base64: %v", err)
			}

			decrypted, err := AesDecryptCBCString(key, cipherBase64)
			if err != nil {
				t.Fatalf("decrypt failed: %v", err)
			}

			if decrypted != pt {
				t.Errorf("roundtrip mismatch: got %v, want %v", decrypted, pt)
			}
		})
	}
}

func TestAesDecryptCBCString_InvalidBase64(t *testing.T) {
	_, err := AesDecryptCBCString("1234567890123456", "!!!not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestAesDecryptCBCString_DataTooShort(t *testing.T) {
	key := "1234567890123456"
	// CBC 需要至少 blockSize (16) 字节
	shortData := base64.StdEncoding.EncodeToString([]byte{0x01, 0x02})
	_, err := AesDecryptCBCString(key, shortData)
	if err == nil {
		t.Error("expected error for data too short")
	}
}

func TestAesDecryptCBCString_WrongKey(t *testing.T) {
	key1 := "1234567890123456"
	key2 := "6543210987654321"
	plainText := "secret message"

	cipherBase64, err := AesEncryptCBCString(key1, plainText)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// CBC 模式无认证，错误密钥不会报错，但解密结果与原文不同
	decrypted, err := AesDecryptCBCString(key2, cipherBase64)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if decrypted == plainText {
		t.Error("wrong key should produce different plaintext")
	}
}

func TestAesEncryptCBCString_InvalidKey(t *testing.T) {
	_, err := AesEncryptCBCString("bad", "test")
	if err == nil {
		t.Error("expected error for invalid key")
	}
}

func TestAesDecryptCBCString_InvalidKeyLength(t *testing.T) {
	key := "1234567890123456"
	cipherBase64, _ := AesEncryptCBCString(key, "test")

	// 使用无效长度的密钥解密有效 base64 数据
	_, err := AesDecryptCBCString("bad", cipherBase64)
	if err == nil {
		t.Error("expected error for invalid key length in AesDecryptCBCString")
	}
}
