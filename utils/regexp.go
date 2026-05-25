package utils

import (
	"encoding/json"
	"regexp"
	"strings"
	"unicode"
)

// MatchingContainNumber 定义匹配包含数字
var MatchingContainNumber = regexp.MustCompile(`\d`) // [0-9]

// IsContainNumber 是否至少包含一个数字
func IsContainNumber(input string) bool {
	return MatchingContainNumber.MatchString(input)
}

// MatchingContainLetter 定义匹配包含字母
var MatchingContainLetter = regexp.MustCompile(`[a-zA-Z]`)

// IsContainLetter 是否至少包含一个字母
func IsContainLetter(str string) bool {
	return MatchingContainLetter.MatchString(str)
}

// MatchingAllLetter 定义匹配全是字母
var MatchingAllLetter = regexp.MustCompile(`^[a-zA-Z]+$`)

// IsAllLetter 是否全是字母
func IsAllLetter(str string) bool {
	return MatchingAllLetter.MatchString(str)
}

// IsAllUpper 是否全是大写字母
func IsAllUpper(str string) bool {
	for _, r := range str {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return str != ""
}

// IsAllLower 是否全是小写字母
func IsAllLower(str string) bool {
	for _, r := range str {
		if !unicode.IsLower(r) {
			return false
		}
	}
	return str != ""
}

// IsAllASCII 是否全是 ASCII 字符
func IsAllASCII(str string) bool {
	for i := 0; i < len(str); i++ {
		if str[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// IsPrintable 是否全是可打印字符组成
func IsPrintable(str string) bool {
	for _, r := range str {
		if !unicode.IsPrint(r) {
			if r == '\n' || r == '\r' || r == '\t' || r == '`' {
				continue
			}
			return false
		}
	}
	return true
}

// IsContainUpper 是否至少包含一个大写字母
func IsContainUpper(str string) bool {
	for _, r := range str {
		if unicode.IsUpper(r) && unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// IsContainLower 是否至少包含一个小写字母
func IsContainLower(str string) bool {
	for _, r := range str {
		if unicode.IsLower(r) && unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

var MatchingBase64 = regexp.MustCompile(`^(?:[A-Za-z0-9+\\/]{4})*(?:[A-Za-z0-9+\\/]{2}==|[A-Za-z0-9+\\/]{3}=|[A-Za-z0-9+\\/]{4})$`)
var MatchingBase64URL = regexp.MustCompile(`^([A-Za-z0-9_-]{4})*([A-Za-z0-9_-]{2}(==)?|[A-Za-z0-9_-]{3}=?)?$`)

// IsBase64 是否为 Base64 编码
func IsBase64(base64 string) bool {
	return MatchingBase64.MatchString(base64)
}

// IsBase64URL 是否为有效的 URL 安全 Base64 编码
func IsBase64URL(v string) bool {
	return MatchingBase64URL.MatchString(v)
}

// MatchingHex 定义匹配十六进制
var MatchingHex = regexp.MustCompile(`^(#|0x|0X)?[0-9a-fA-F]+$`)

// IsHex 是否为十六进制
func IsHex(v string) bool {
	return MatchingHex.MatchString(v)
}

// IsJSON 是否为 JSON 格式
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// IsJWT 是否为 JWT 格式
func IsJWT(v string) bool {
	ss := strings.Split(v, ".")
	if len(ss) != 3 {
		return false
	}
	for _, s := range ss {
		if !IsBase64URL(s) {
			return false
		}
	}
	return true
}
