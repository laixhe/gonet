package xgin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/laixhe/gonet/jwt"
)

// UseJwt 创建 JWT 鉴权中间件（可选模式），从 Authorization header 提取 Bearer token 并解析到 claims 中。
// 解析成功时将 claims 存入 gin.Context（key 为 jwt.AuthorizationClaims），后续 handler 可通过 ctx.Get 获取。
// 解析失败或未携带 token 时不阻断请求，仅跳过解析 —— 配合业务层二次校验使用。
// config 为 JWT 解析配置（密钥、算法等），claims 为接收解析结果的结构体实例。
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

// JwtRequired 创建 JWT 鉴权中间件（阻断模式），校验失败时直接返回 401 JSON。
// 适用于必须登录才能访问的接口，不合法的 token 会被拦截。
func JwtRequired(config *jwt.Config, claims jwtv5.Claims) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorization := ctx.Request.Header.Get(jwt.Authorization)
		if authorization == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, AuthorizedError())
			return
		}
		if !strings.HasPrefix(authorization, jwt.Bearer) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, AuthorizedError())
			return
		}
		claimsToken, err := jwt.ParseToken(config, authorization[jwt.BearerLen:], claims)
		if err != nil || claimsToken == nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, AuthorizedError())
			return
		}
		ctx.Set(jwt.AuthorizationClaims, claimsToken)
		ctx.Next()
	}
}
