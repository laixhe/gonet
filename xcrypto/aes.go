package xcrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

// AES 加密数据块分组长度必须为 128bit(16bytes)
// 密钥长度可以是 128bit(16bytes)、192bit(24bytes)、256bit(32bytes) 中的任意一个

// AesEncryptCBC AES加密，使用 CBC 模式
func AesEncryptCBC(orig string, key string) (string, error) {
	origData := []byte(orig)
	keyData := []byte(key)
	// 分组秘钥
	block, err := aes.NewCipher(keyData)
	if err != nil {
		return "", err
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补全码
	origData = PKCS7Padding(origData, blockSize)
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, keyData[:blockSize])
	// 创建数组
	crypted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(crypted, origData)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

// AesDecryptCBC AES解密，使用 CBC 模式
func AesDecryptCBC(crypted string, key string) (string, error) {
	// 转成字节数组
	cryptedByte, err := base64.StdEncoding.DecodeString(crypted)
	if err != nil {
		return "", err
	}
	keyData := []byte(key)
	// 分组秘钥
	block, err := aes.NewCipher(keyData)
	if err != nil {
		return "", err
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, keyData[:blockSize])
	// 创建数组
	origData := make([]byte, len(cryptedByte))
	// 解密
	blockMode.CryptBlocks(origData, cryptedByte)
	// 去补全码
	origData = PKCS7UnPadding(origData)
	return string(origData), nil
}

// PKCS7Padding 使用 AES 加密文本的时候文本必须定长,即必须是 16,24,32 的整数倍
func PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS7UnPadding 使用 AES 解密文本,解密收删除 padding 的文本
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	if length == 0 {
		return origData
	}
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
