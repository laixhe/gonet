package utils

import "regexp"

// MatchingContainChinese 定义匹配包含中文
var MatchingContainChinese = regexp.MustCompile("[\u4e00-\u9fa5]")

// MatchingChineseCharacters 定义匹配中文字符(连续的中文)
var MatchingChineseCharacters = regexp.MustCompile(`\p{Han}+`)

// MatchingEachChineseCharacters 定义匹配中文字符(每个的中文)
var MatchingEachChineseCharacters = regexp.MustCompile(`\p{Han}`)

// MatchingNonChineseCharacters 定义匹配非中文字符(连续的非中文)
var MatchingNonChineseCharacters = regexp.MustCompile(`[^\p{Han}]+`)

// MatchingNonEachChineseCharacters 定义匹配非中文字符(每个的非中文)
var MatchingNonEachChineseCharacters = regexp.MustCompile(`[^\p{Han}]`)

// IsContainChinese 是否包含中文
func IsContainChinese(s string) bool {
	return MatchingContainChinese.MatchString(s)
}

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
