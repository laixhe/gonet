package crypto

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestAesDecryptGCM(t *testing.T) {
	key := []byte("aes-gcm-128-bit-") // AES-128 密钥
	plainText := []byte("sensitive data to encrypt 128")
	cipherText, nonce, err := AesEncryptGCM(key, plainText)
	if err != nil {
		fmt.Println("encrypt err", err)
		return
	}

	fmt.Println("decrypted hex nonce: ", hex.EncodeToString(nonce))
	fmt.Println("decrypted hex key: ", hex.EncodeToString(key))
	fmt.Println("decrypted hex: ", hex.EncodeToString(cipherText))
	fmt.Println("decrypted base64: ", base64.StdEncoding.EncodeToString(cipherText))

	decrypted, err := AesDecryptGCM(key, cipherText, nonce)
	if err != nil {
		fmt.Println("decrypt err", err)
		return
	}
	fmt.Println("decrypted: ", string(decrypted))
}

func TestAesDecryptCTR(t *testing.T) {
	key := []byte("aes-ctr-256-bit-key-32bytes-long") // AES-256 密钥
	plainText := []byte("sensitive data to encrypt ctr 256")
	cipherText, vi, err := AesEncryptCTR(key, plainText)
	if err != nil {
		fmt.Println("encrypt err", err)
		return
	}

	fmt.Println("decrypted vi: ", vi)
	fmt.Println("decrypted vi string: ", string(vi))
	fmt.Println("decrypted hex vi: ", hex.EncodeToString(vi))
	fmt.Println("decrypted hex key: ", hex.EncodeToString(key))
	fmt.Println("decrypted hex: ", hex.EncodeToString(cipherText))
	fmt.Println("decrypted base64: ", base64.RawURLEncoding.EncodeToString(cipherText))

	decrypted, err := AesDecryptCTR(key, cipherText, vi)
	if err != nil {
		fmt.Println("decrypt err", err)
		return
	}
	fmt.Println("decrypted: ", string(decrypted))
}

func TestAesDecryptCBC(t *testing.T) {
	key := []byte("aes-cbc-256-bit-key-32bytes-long") // AES-256 密钥
	//key := []byte("aes-cbc-128-bit-") // AES-128 密钥
	plainText := []byte("sensitive data to encrypt 256")
	cipherText, vi, err := AesEncryptCBC(key, plainText)
	if err != nil {
		fmt.Println("encrypt err", err)
		return
	}

	fmt.Println("decrypted hex vi: ", hex.EncodeToString(vi))
	fmt.Println("decrypted hex key: ", hex.EncodeToString(key))
	fmt.Println("decrypted hex: ", hex.EncodeToString(cipherText))
	fmt.Println("decrypted base64: ", base64.StdEncoding.EncodeToString(cipherText))

	decrypted, err := AesDecryptCBC(key, cipherText, vi)
	if err != nil {
		fmt.Println("decrypt err", err)
		return
	}
	fmt.Println("decrypted: ", string(decrypted))
}
