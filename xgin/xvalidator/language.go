package xvalidator

import "github.com/laixhe/gonet/xi18n"

// 语言

const (
	Zh = "zh" // 中文(简体)
	En = "en" // 英文
)

var mapLanguage = map[string]string{
	Zh: Zh,
	En: En,
}

func LanguageText(language string) string {
	return mapLanguage[language]
}

func I18nToLanguage(language string) string {
	switch language {
	case xi18n.ZhCn:
		return Zh
	case xi18n.En:
		return En
	default:
		return Zh
	}
}
