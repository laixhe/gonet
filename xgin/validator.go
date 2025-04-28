package xgin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	translator "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"

	"github.com/laixhe/gonet/xi18n"
)

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

// 表单验证

// 全局翻译器
var trans translator.Translator

// ValidatorTranslator 表单验证多语言提示文本
func ValidatorTranslator(language string) (err error) {
	// 修改 gin 框架中 Validator 引擎属性，实现自定制
	if validate, ok := binding.Validator.Engine().(*validator.Validate); ok {
		zhT := zh.New() // 中文
		enT := en.New() // 英文
		// 第一个参数是备用语言，后面参数是应该支持多个语言
		universalTranslator := translator.New(zhT, zhT, enT)
		// language 通常取决于 http 请求 Accept-language
		var is bool
		trans, is = universalTranslator.GetTranslator(language)
		if !is {
			return fmt.Errorf("translator.GetTranslator(%s) failed", language)
		}
		// 注册默认翻译器
		switch language {
		case En:
			err = enTranslations.RegisterDefaultTranslations(validate, trans)
		default:
			err = zhTranslations.RegisterDefaultTranslations(validate, trans)
		}
	}
	return
}

func removeTopStruct(fields map[string]string) map[string]string {
	res := map[string]string{}
	for field, err := range fields {
		res[field[strings.Index(field, ".")+1:]] = err
	}
	return res
}

func TranslatorErrorString(err validator.ValidationErrors) string {
	str := ""
	s := removeTopStruct(err.Translate(trans))
	for _, v := range s {
		str += v + ","
	}
	if str == "" {
		return ""
	}
	return str[:len(str)-1]
}

func TranslatorError(err error) error {
	if err1, is1 := err.(validator.ValidationErrors); is1 {
		return errors.New(TranslatorErrorString(err1))
	}
	return err
}
