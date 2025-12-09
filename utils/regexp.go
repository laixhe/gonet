package utils

import "regexp"

// 定义匹配中文字符
var hanMatching = regexp.MustCompile(`\p{Han}+`)

// ExtractAllHan 提取所有匹配的中文字符
func ExtractAllHan(s string) []string {
	// 查找所有匹配项
	return hanMatching.FindAllString(s, -1)
}

// phoneMatching 定义匹配手机号码
var phoneMatching = regexp.MustCompile(`1\d{10}`)

// ObfuscatePhone 隐藏手机号码中间部分
func ObfuscatePhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

// ReplaceObfuscatePhone 替换字符串中的所有手机号为隐藏格式
func ReplaceObfuscatePhone(text string) string {
	// 替换所有匹配项
	return phoneMatching.ReplaceAllStringFunc(text, ObfuscatePhone)
}
