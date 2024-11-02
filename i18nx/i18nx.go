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

var (
	i18n = gi18n.New()

	CtxCn = context.Background()
	CtxEn = context.Background()
)

func Init() {
	i18n.SetLanguage(eapp.Language_zh_cn.String())
	CtxCn = gi18n.WithLanguage(context.Background(), eapp.Language_zh_cn.String())
	CtxEn = gi18n.WithLanguage(context.Background(), eapp.Language_en.String())
}

func FromContext(c *gin.Context, s string) string {
	language := c.Request.Header.Get(LanguageHeaderKey)
	if language == eapp.Language_en.String() {
		return i18n.Translate(CtxEn, s)
	}
	return i18n.Translate(CtxCn, s)
}

func FromContextError(c *gin.Context, s string) error {
	errStr := FromContext(c, s)
	return errors.New(errStr)
}
