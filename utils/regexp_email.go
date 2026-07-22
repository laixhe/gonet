package utils

import "regexp"

// MatchingEmail 定义匹配邮箱（大小写不敏感）
var MatchingEmail = regexp.MustCompile(`(?i)^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

// IsEmail 是否为邮箱
func IsEmail(email string) bool {
	return MatchingEmail.MatchString(email)
}
