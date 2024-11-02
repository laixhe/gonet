package ginx

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/laixhe/gonet/errorx"
	"github.com/laixhe/gonet/jwtx"
	"github.com/laixhe/gonet/proto/gen/config/cauth"
	"github.com/laixhe/gonet/responsex"
)

// 中间件

// JwtAuth 鉴权
func JwtAuth(cjwt *cauth.Jwt) gin.HandlerFunc {
	return func(c *gin.Context) {
		var parseTokenErr error
		token := c.Request.Header.Get(jwtx.Authorization)
		if len(token) > 0 {
			if strings.HasPrefix(token, jwtx.Bearer) {
				claims, err := jwtx.ParseToken(cjwt, token[jwtx.BearerLen:])
				if err == nil {
					c.Set(jwtx.AuthorizationClaimsHeaderKey, claims)
					c.Next()
					return
				}
				parseTokenErr = errorx.AuthInvalidError(err)
			}
		}
		c.JSON(http.StatusOK, responsex.ResponseError(parseTokenErr))
		// 返回错误
		c.Abort()
	}
}

func ContextUid(c *gin.Context) (uint64, error) {
	value, exists := c.Get(jwtx.AuthorizationClaimsHeaderKey)
	if exists {
		customClaims, is := value.(*jwtx.CustomClaims)
		if is {
			return customClaims.Uid, nil
		}
	}
	return 0, errorx.AuthInvalidError(nil)
}
