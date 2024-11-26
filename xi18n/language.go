package xi18n

// 语言

const (
	ZhCn = "zh-CN" // 中文(简体)
	ZhTw = "zh-TW" // 中文(繁体)
	En   = "en"    // 英文
)

var mapLanguage = map[string]string{
	ZhCn: ZhCn,
	ZhTw: ZhTw,
	En:   En,
}

func LanguageText(language string) string {
	return mapLanguage[language]
}
