package utils

import "regexp"

// MatchingChineseCharacters 定义匹配中文字符(连续的中文)
var MatchingChineseCharacters = regexp.MustCompile(`\p{Han}+`)

// MatchingEachChineseCharacters 定义匹配中文字符(每个的中文)
var MatchingEachChineseCharacters = regexp.MustCompile(`\p{Han}`)

// MatchingNonChineseCharacters 定义匹配非中文字符(连续的非中文)
var MatchingNonChineseCharacters = regexp.MustCompile(`[^\p{Han}]+`)

// MatchingNonEachChineseCharacters 定义匹配非中文字符(每个的非中文)
var MatchingNonEachChineseCharacters = regexp.MustCompile(`[^\p{Han}]`)

// ExtractChineseCharacters 提取所有匹配的中文字符(连续的中文)
func ExtractChineseCharacters(s string) []string {
	// 查找所有匹配项
	return MatchingChineseCharacters.FindAllString(s, -1)
}

// ExtractEachChineseCharacters 提取所有匹配的中文字符(每个的中文)
func ExtractEachChineseCharacters(s string) []string {
	// 查找所有匹配项
	return MatchingEachChineseCharacters.FindAllString(s, -1)
}

// ExtractNonChineseCharacters 提取所有匹配的非中文字符(连续的非中文)
func ExtractNonChineseCharacters(s string) []string {
	// 查找所有匹配项
	return MatchingNonChineseCharacters.FindAllString(s, -1)
}

// ExtractNonEachChineseCharacters 提取所有匹配的非中文字符(每个的非中文)
func ExtractNonEachChineseCharacters(s string) []string {
	// 查找所有匹配项
	return MatchingNonEachChineseCharacters.FindAllString(s, -1)
}

// ChineseMobileMatcher 定义匹配手机号码
var ChineseMobileMatcher = regexp.MustCompile(`1\d{10}`)

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
	return ChineseMobileMatcher.ReplaceAllStringFunc(text, ObfuscateMobile)
}

// IsChineseMobile 是否手机号码
func IsChineseMobile(mobileNum string) bool {
	return ChineseMobileMatcher.MatchString(mobileNum)
}
