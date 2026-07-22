package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// SHA256 计算 SHA-256 哈希值（小写十六进制）
// 目前无已知碰撞攻击，是安全场景的推荐哈希算法
func SHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// HmacSha256 计算带密钥的 HMAC-SHA256 消息认证码（小写十六进制）
func HmacSha256(key string, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}
