package xgin

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/laixhe/gonet/protocol/gen/config/cauth"
	"github.com/laixhe/gonet/xerror"
	xginConstant "github.com/laixhe/gonet/xgin/constant"
	"github.com/laixhe/gonet/xjwt"
	"github.com/laixhe/gonet/xlog"
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
					xlog.Debug("jwt",
						zap.String(xginConstant.HeaderRequestID, requestid.Get(c)),
						zap.String("method", c.Request.Method),
						zap.String("path", c.Request.URL.Path),
						zap.String("query", c.Request.URL.RawQuery),
						zap.Any("jwt_claims", claims))
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
					xlog.Debug("jwt",
						zap.String(xginConstant.HeaderRequestID, requestid.Get(c)),
						zap.String("method", c.Request.Method),
						zap.String("path", c.Request.URL.Path),
						zap.String("query", c.Request.URL.RawQuery),
						zap.Any("jwt_claims", claims))
					c.Set(xjwt.AuthorizationClaimsHeaderKey, claims)
				}
			}
		}
		c.Next()
	}
}

// ContextUid 获取用户ID
func ContextUid(c *gin.Context) int {
	value, exists := c.Get(xjwt.AuthorizationClaimsHeaderKey)
	if exists {
		customClaims, is := value.(*xjwt.CustomClaims)
		if is {
			return customClaims.Uid
		}
	}
	return 0
}

// ContextUid64 获取用户ID
func ContextUid64(c *gin.Context) int64 {
	value, exists := c.Get(xjwt.AuthorizationClaimsHeaderKey)
	if exists {
		customClaims, is := value.(*xjwt.CustomClaims)
		if is {
			return int64(customClaims.Uid)
		}
	}
	return 0
}

// ContextID 获取ID
func ContextID(c *gin.Context) string {
	value, exists := c.Get(xjwt.AuthorizationClaimsHeaderKey)
	if exists {
		customClaims, is := value.(*xjwt.CustomClaims)
		if is {
			return customClaims.ID
		}
	}
	return ""
}

// ContextClaims 获取自定义声明类型
func ContextClaims(c *gin.Context) *xjwt.CustomClaims {
	value, exists := c.Get(xjwt.AuthorizationClaimsHeaderKey)
	if exists {
		customClaims, is := value.(*xjwt.CustomClaims)
		if is {
			return customClaims
		}
	}
	return &xjwt.CustomClaims{}
}

// IsLogin 是否登录
func IsLogin(err xerror.IError) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := ContextUid(c)
		if uid > 0 {
			c.Next()
			return
		}
		c.JSON(http.StatusOK, xresponse.Error(err))
		// 返回错误
		c.Abort()
	}
}
