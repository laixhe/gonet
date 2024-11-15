package xgin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/laixhe/gonet/proto/gen/config/cauth"
	"github.com/laixhe/gonet/xerror"
	"github.com/laixhe/gonet/xjwt"
	"github.com/laixhe/gonet/xresponse"
)

// 中间件

// JwtAuth 鉴权
func JwtAuth(cjwt *cauth.Jwt) gin.HandlerFunc {
	return func(c *gin.Context) {
		var parseTokenErr error
		token := c.Request.Header.Get(xjwt.Authorization)
		if len(token) > 0 {
			if strings.HasPrefix(token, xjwt.Bearer) {
				claims, err := xjwt.ParseToken(cjwt, token[xjwt.BearerLen:])
				if err == nil {
					c.Set(xjwt.AuthorizationClaimsHeaderKey, claims)
					c.Next()
					return
				}
				parseTokenErr = xerror.AuthInvalidError(err)
			}
		}
		c.JSON(http.StatusOK, xresponse.ResponseError(parseTokenErr))
		// 返回错误
		c.Abort()
	}
}

func ContextUid(c *gin.Context) (uint64, error) {
	value, exists := c.Get(xjwt.AuthorizationClaimsHeaderKey)
	if exists {
		customClaims, is := value.(*xjwt.CustomClaims)
		if is {
			return customClaims.Uid, nil
		}
	}
	return 0, xerror.AuthInvalidError(nil)
}
