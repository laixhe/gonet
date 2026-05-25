package utils

import (
	"regexp"
	"strings"
)

// MatchingEmail 定义匹配邮箱
var MatchingEmail = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

// IsEmail 是否为邮箱
func IsEmail(email string) bool {
	return MatchingEmail.MatchString(strings.ToLower(email))
}
