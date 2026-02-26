package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// 对称加密
// AES 是分组密码算法
// 支持 128bit(16bytes)、192bit(24bytes)、256bit(32bytes) 位密钥长度（对应 AES-128、AES-192、AES-256）
// 分组长度固定为 128bit(16bytes)
// GCM 认证加密模式（AEAD），提供完整性校验和抗重放攻击，推荐首选（内置支持）
// CBC 分组密码模式，需要初始化向量（IV），填充常用 PKCS#7，但不提供认证（内置支持）
// CTR 流模式，将分组转化为流加密，支持并行计算，适合高速场景（内置支持）
// ECB 不使用向量（IV），相同明文生成相同密文，存在严重安全缺陷（禁止使用）（需手动实现，不推荐）

// AesEncryptGCM 使用 GCM 模式的 AES 加密
// key 密钥，长度要 16/24/32 字节分别对应 AES-128/192/256
// plainText 内容文本
// 返回加密内容、随机生成、错误
func AesEncryptGCM(key, plainText []byte) ([]byte, []byte, error) {
	// 分组秘钥(密文块)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	// 随机生成（每次加密唯一）
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	// 进行加密
	ciphertext := gcm.Seal(nil, nonce, plainText, nil)
	return ciphertext, nonce, nil
}

// AesDecryptGCM 使用 GCM 模式的 AES 解密
// key 密钥，长度要 16/24/32 字节分别对应 AES-128/192/256
// cipherText 加密内容
// nonce 随机
func AesDecryptGCM(key, cipherText, nonce []byte) ([]byte, error) {
	// 分组秘钥
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	// 进行解密
	return gcm.Open(nil, nonce, cipherText, nil)
}

// AesEncryptCTR 使用 CTR 模式的 AES 加密
// key 密钥，长度要 16/24/32 字节分别对应 AES-128/192/256
// plainText 内容文本
// 返回加密内容、随机生成向量、错误
func AesEncryptCTR(key, plainText []byte) ([]byte, []byte, error) {
	// 分组秘钥(密文块)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	// 随机生成向量（每次加密唯一）
	iv := make([]byte, block.BlockSize())
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, err
	}
	// 创建加密内容(存放)
	cipherText := make([]byte, len(plainText))
	// 创建加密流
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(cipherText, plainText)
	return cipherText, iv, nil
}

// AesDecryptCTR 使用 CTR 模式的 AES 解密
// key 密钥，长度要 16/24/32 字节分别对应 AES-128/192/256
// cipherText 加密内容
// iv 随机向量
func AesDecryptCTR(key, cipherText, iv []byte) ([]byte, error) {
	// 分组秘钥
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 创建解密内容(存放)
	plainText := make([]byte, len(cipherText))
	// 创建加密流
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plainText, cipherText)
	return plainText, nil
}

// AesEncryptCBC 使用 CBC 模式的 AES 加密
// key 密钥，长度要 16/24/32 字节分别对应 AES-128/192/256
// plainText 内容文本
func AesEncryptCBC(key, plainText []byte) ([]byte, []byte, error) {
	// 分组秘钥
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补全码
	plainTextPadding := PKCS7Padding(plainText, blockSize)
	// 创建加密内容(存放)
	cipherText := make([]byte, len(plainTextPadding))
	// 生成随机向量
	iv := make([]byte, blockSize)
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, err
	}
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, iv)
	// 进行加密
	blockMode.CryptBlocks(cipherText, plainTextPadding)
	return cipherText, iv, nil
}

// AesDecryptCBC 使用 CBC 模式的 AES 解密
// key 密钥，长度要 16/24/32 字节分别对应 AES-128/192/256
// cipherText 加密内容
// vi 向量
func AesDecryptCBC(key, cipherText, vi []byte) ([]byte, error) {
	// 分组秘钥
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 创建解密内容(存放)
	plainTextPadding := make([]byte, len(cipherText))
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, vi)
	// 进行解密
	blockMode.CryptBlocks(plainTextPadding, cipherText)
	// 去补全码
	plainText := PKCS7UnPadding(plainTextPadding)
	return plainText, nil
}

// PKCS7Padding 填充（加密文本的时候文本必须定长,即必须是 16,24,32 的整数倍）
func PKCS7Padding(plainText []byte, blocksize int) []byte {
	padding := blocksize - len(plainText)%blocksize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plainText, paddingText...)
}

// PKCS7UnPadding 去填充
func PKCS7UnPadding(cipherText []byte) []byte {
	length := len(cipherText)
	if length == 0 {
		return cipherText
	}
	unPadding := int(cipherText[length-1])
	return cipherText[:(length - unPadding)]
}
