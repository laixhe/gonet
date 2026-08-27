package utils

import "regexp"

// MatchingChineseMobile 定义匹配手机号码
var MatchingChineseMobile = regexp.MustCompile(`1\d{10}`)

// ObfuscateMobile 隐藏手机号码中间部分
func ObfuscateMobile(mobileNum string) string {
	if len(mobileNum) != 11 {
		return mobileNum
	}
	return mobileNum[:3] + "****" + mobileNum[7:]
}

// ReplaceObfuscateMobile 替换字符串中的所有手机号为隐藏格式
func ReplaceObfuscateMobile(text string) string {
	// 替换所有匹配项
	return MatchingChineseMobile.ReplaceAllStringFunc(text, ObfuscateMobile)
}

// IsChineseMobile 是否包含手机号码（子串匹配，非全串匹配；若要严格校验整串请自行加 ^$ 锚点）
func IsChineseMobile(mobileNum string) bool {
	return MatchingChineseMobile.MatchString(mobileNum)
}
