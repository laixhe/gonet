package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
)

// MD5 加密
func MD5(data string) string {
	md5Data := md5.Sum([]byte(data))
	return hex.EncodeToString(md5Data[:])
}

// HmacMd5 带秘钥加密
func HmacMd5(key string, data string) string {
	mac := hmac.New(md5.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}
