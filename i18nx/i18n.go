package i18nx

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/i18n/gi18n"

	"github.com/laixhe/gonet/proto/gen/enum/eapp"
)

// 头部
const LanguageHeaderKey = "accept-language"

type I18n struct {
	httpHeaderKey string
	i18n          *gi18n.Manager
	zhCnCtx       context.Context
	enCtx         context.Context
}

func NewI18n(httpHeaderKey string) *I18n {
	if httpHeaderKey == "" {
		httpHeaderKey = LanguageHeaderKey
	}
	return &I18n{
		httpHeaderKey: httpHeaderKey,
		i18n:          gi18n.New(),
		zhCnCtx:       gi18n.WithLanguage(context.Background(), eapp.Language_zh_cn.String()),
		enCtx:         gi18n.WithLanguage(context.Background(), eapp.Language_en.String()),
	}
}

func (in *I18n) GinContext(c *gin.Context, s string) string {
	language := c.Request.Header.Get(in.httpHeaderKey)
	if language == eapp.Language_en.String() {
		return in.i18n.Translate(in.enCtx, s)
	}
	return in.i18n.Translate(in.zhCnCtx, s)
}

func (in *I18n) GinContextError(c *gin.Context, s string) error {
	errStr := in.GinContext(c, s)
	return errors.New(errStr)
}
