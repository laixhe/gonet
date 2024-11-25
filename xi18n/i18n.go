package xi18n

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/i18n/gi18n"
)

// 头部

const LanguageHeaderKey = "accept-language"

type I18n struct {
	httpHeaderKey string
	i18n          *gi18n.Manager
	zhCtx         context.Context
	enCtx         context.Context
}

func NewI18n(httpHeaderKey string) *I18n {
	if httpHeaderKey == "" {
		httpHeaderKey = LanguageHeaderKey
	}
	return &I18n{
		httpHeaderKey: httpHeaderKey,
		i18n:          gi18n.New(),
		zhCtx:         gi18n.WithLanguage(context.Background(), ZhCn),
		enCtx:         gi18n.WithLanguage(context.Background(), En),
	}
}

func (in *I18n) GinContext(c *gin.Context, s string) string {
	language := c.Request.Header.Get(in.httpHeaderKey)
	languageText := LanguageText(language)
	if languageText == En {
		return in.i18n.Translate(in.enCtx, s)
	}
	return in.i18n.Translate(in.zhCtx, s)
}

func (in *I18n) GinContextError(c *gin.Context, s string) error {
	errStr := in.GinContext(c, s)
	return errors.New(errStr)
}

func (in *I18n) ToLanguage(c *gin.Context) string {
	language := c.Request.Header.Get(in.httpHeaderKey)
	languageText := LanguageText(language)
	if languageText == "" {
		return ZhCn
	}
	return languageText
}
