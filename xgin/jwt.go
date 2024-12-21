package xgin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/laixhe/gonet/protocol/gen/config/cauth"
	"github.com/laixhe/gonet/xerror"
	"github.com/laixhe/gonet/xjwt"
	"github.com/laixhe/gonet/xresponse"
)

// 中间件

// JwtAuth 鉴权
// cjwt 配置
// parseTokenError 错误
func JwtAuth(cjwt *cauth.Jwt, parseTokenError xerror.IError) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get(xjwt.Authorization)
		if len(token) > 0 {
			if strings.HasPrefix(token, xjwt.Bearer) {
				claims, err := xjwt.ParseToken(cjwt, token[xjwt.BearerLen:])
				if err == nil {
					c.Set(xjwt.AuthorizationClaimsHeaderKey, claims)
					c.Next()
					return
				}
			}
		}
		c.JSON(http.StatusOK, xresponse.Error(parseTokenError))
		// 返回错误
		c.Abort()
	}
}

// JwtAuthAuto 自动鉴权
// cjwt 配置
// parseTokenError 错误
func JwtAuthAuto(cjwt *cauth.Jwt, parseTokenError xerror.IError) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get(xjwt.Authorization)
		if len(token) > 0 {
			if strings.HasPrefix(token, xjwt.Bearer) {
				claims, err := xjwt.ParseToken(cjwt, token[xjwt.BearerLen:])
				if err == nil {
					c.Set(xjwt.AuthorizationClaimsHeaderKey, claims)
				}
			}
		}
		c.Next()
	}
}

func ContextUid(c *gin.Context) uint64 {
	value, exists := c.Get(xjwt.AuthorizationClaimsHeaderKey)
	if exists {
		customClaims, is := value.(*xjwt.CustomClaims)
		if is {
			return customClaims.Uid
		}
	}
	return 0
}
