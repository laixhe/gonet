package crypto

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
)

// SHA1 计算 SHA-1 哈希值（小写十六进制）
// 注意：SHA-1 存在碰撞攻击（如 SHAttered），请勿用于安全敏感场景，建议使用 SHA-256 或更高版本
func SHA1(data string) string {
	sha1Data := sha1.Sum([]byte(data))
	return hex.EncodeToString(sha1Data[:])
}

// HmacSha1 计算带密钥的 HMAC-SHA1 消息认证码（小写十六进制）
func HmacSha1(key string, data string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}
