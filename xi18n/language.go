package xi18n

// 语言

const ZhCn = "zh-CN" // 中文(简体)
const ZhTw = "zh-TW" // 中文(繁体)
const En = "en"      // 英文

var mapLanguage = map[string]string{
	ZhCn: ZhCn,
	ZhTw: ZhTw,
	En:   En,
}

func LanguageText(language string) string {
	return mapLanguage[language]
}
