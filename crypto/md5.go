package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
)

// MD5 计算 MD5 哈希值（小写十六进制）
// 注意：MD5 存在碰撞攻击，已不再安全，请勿用于密码存储或安全签名场景
func MD5(data string) string {
	md5Data := md5.Sum([]byte(data))
	return hex.EncodeToString(md5Data[:])
}

// HmacMd5 计算带密钥的 HMAC-MD5 消息认证码（小写十六进制）
func HmacMd5(key string, data string) string {
	mac := hmac.New(md5.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}
