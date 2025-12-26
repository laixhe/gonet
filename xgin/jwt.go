package xgin

import (
	"strings"

	"github.com/gin-gonic/gin"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/laixhe/gonet/jwt"
)

// UseJwt 中间件-JWT
func UseJwt(config *jwt.Config, claims jwtv5.Claims) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorization := ctx.Request.Header.Get(jwt.Authorization)
		if authorization != "" {
			if strings.HasPrefix(authorization, jwt.Bearer) {
				claimsToken, err := jwt.ParseToken(config, authorization[jwt.BearerLen:], claims)
				if err == nil && claimsToken != nil {
					ctx.Set(jwt.AuthorizationClaims, claimsToken)
				}
			}
		}
		ctx.Next()
	}
}
