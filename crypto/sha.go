package crypto

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
)

// SHA1 加密
func SHA1(data string) string {
	sha1Data := sha1.Sum([]byte(data))
	return hex.EncodeToString(sha1Data[:])
}

// HmacSha1 带秘钥加密
func HmacSha1(key string, data string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}
