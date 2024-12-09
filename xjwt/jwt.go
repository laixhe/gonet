package xjwt

import (
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/laixhe/gonet/protocol/gen/config/cauth"
)

var (
	ErrTokenExpired = errors.New("token is expired") // 令牌已过期
	ErrTokenInvalid = errors.New("token invalid")    // 令牌无效
)

// Authorization
const (
	Authorization = "Authorization"
	Bearer        = "Bearer "
	BearerLen     = 7
)

// 头部

const AuthorizationClaimsHeaderKey = "AuthorizationClaims"

// Checking 检查
func Checking(cjwt *cauth.Jwt) error {
	if cjwt == nil {
		return errors.New("config jwt is nil")
	}
	if cjwt.SecretKey == "" {
		return errors.New("config jwt secret_key is empty")
	}
	return nil
}

// CustomClaims 自定义声明类型 并内嵌 jwt.RegisteredClaims
// jwt 包自带的 jwt.RegisteredClaims 只包含了官方字段
type CustomClaims struct {
	// 可根据需要自行添加字段
	Uid uint64 `json:"uid"`
	jwt.RegisteredClaims
}

// GenToken 生成JWT
func GenToken(cjwt *cauth.Jwt, uid uint64, id string) (string, error) {
	claims := CustomClaims{
		Uid: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ID: id,
		},
	}

	nowTime := time.Now()
	// 过期时间
	if cjwt.GetExpireTime() > 0 {
		claims.ExpiresAt = jwt.NewNumericDate(nowTime.Add(time.Duration(cjwt.GetExpireTime()) * time.Second))
	}
	// 发布时间（创建时间）
	claims.IssuedAt = jwt.NewNumericDate(nowTime)
	// 生效时间
	claims.NotBefore = jwt.NewNumericDate(nowTime)

	// 使用指定的签名方法创建签名对象（使用 HS256 签名算法）
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 使用指定的 secret 签名并获得完整的编码后的字符串 token
	return token.SignedString([]byte(cjwt.SecretKey))
}

// ParseToken 解析JWT
func ParseToken(cjwt *cauth.Jwt, tokenString string) (*CustomClaims, error) {
	// 如果是自定义 Claim 结构体则需要使用 ParseWithClaims 方法
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return []byte(cjwt.SecretKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, ErrTokenInvalid
		}
		return nil, ErrTokenInvalid
	}
	if token != nil {
		// 对 token 对象中的 Claim 进行类型断言
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid { // 校验 token
			return claims, nil
		}
	}
	return nil, ErrTokenInvalid
}
