// Package jwt 提供基于 golang-jwt/jwt/v5 的 JWT 令牌生成与解析功能。
//
// 配置示例（YAML）：
//
//	jwt:
//	  secret_key: abcb9f07bfd4eaf7b8a63d9abc
//	  expire_time: 604800
//	  signing_method: HS256
//
// 使用示例：
//
//	// 初始化配置
//	config := &jwt.Config{
//	    SecretKey:     "your-secret-key",
//	    ExpireTime:    604800,
//	    SigningMethod: jwt.SigningMethodHS256,
//	}
//	if err := config.Check(); err != nil {
//	    log.Fatal(err)
//	}
//
//	// 生成 token
//	claims := &jwt.CustomClaims{
//	    Uid: 123456,
//	    RegisteredClaims: jwtv5.RegisteredClaims{
//	        ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(time.Duration(config.ExpireTime) * time.Second)),
//	    },
//	}
//	tokenString, err := jwt.GenToken(config, claims)
//
//	// 解析 token
//	parsedToken, err := jwt.ParseToken(config, tokenString, &jwt.CustomClaims{})
//	if err != nil {
//	    if errors.Is(err, jwt.ErrTokenExpired) {
//	        // token 已过期
//	    } else if errors.Is(err, jwt.ErrTokenInvalid) {
//	        // token 无效
//	    }
//	}
//	if parsedClaims, ok := parsedToken.Claims.(*jwt.CustomClaims); ok {
//	    fmt.Println(parsedClaims.Uid)
//	}
package jwt

import (
	"errors"
	"fmt"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

var (
	// ErrTokenExpired 令牌已过期，可通过 errors.Is 判断
	ErrTokenExpired = errors.New("token is expired")
	// ErrTokenInvalid 令牌无效（签名错误、格式错误等），可通过 errors.Is 判断
	ErrTokenInvalid = errors.New("token invalid")
)

// JWT 认证相关常量
const (
	Authorization       = "Authorization"       // HTTP 认证头字段名
	Bearer              = "Bearer "             // Bearer 前缀（含尾部空格）
	BearerLen           = 7                     // Bearer 前缀长度，用于截取 token
	AuthorizationClaims = "AuthorizationClaims" // 解析后的 token 在 context 中的 key
)

// JWT 签名方法（签名算法）常量
const (
	SigningMethodHS256 = "HS256" // HMAC-SHA256
	SigningMethodHS384 = "HS384" // HMAC-SHA384
	SigningMethodHS512 = "HS512" // HMAC-SHA512
)

// Config JWT 配置，支持 JSON / YAML / TOML / mapstructure 反序列化。
type Config struct {
	// SecretKey 密钥，用于签名和验证 token。
	SecretKey string `json:"secret_key" mapstructure:"secret_key" toml:"secret_key" yaml:"secret_key"`
	// ExpireTime token 过期时长，单位秒。
	ExpireTime int64 `json:"expire_time" mapstructure:"expire_time" toml:"expire_time" yaml:"expire_time"`
	// SigningMethod 签名算法，可选值：HS256 / HS384 / HS512。
	SigningMethod string `json:"signing_method" mapstructure:"signing_method" toml:"signing_method" yaml:"signing_method"`
}

// Check 校验配置是否合法。secret_key 和 expire_time 不能为空或零值，
// signing_method 必须是 HS256 / HS384 / HS512 之一，否则返回错误。
func (c *Config) Check() error {
	if c == nil {
		return errors.New("没有JWT配置")
	}
	if c.SecretKey == "" {
		return errors.New("没有JWT密钥配置")
	}
	if c.ExpireTime <= 0 {
		return errors.New("没有JWT过期时长配置")
	}
	switch c.SigningMethod {
	case SigningMethodHS256, SigningMethodHS384, SigningMethodHS512:
	default:
		c.SigningMethod = SigningMethodHS256
	}
	return nil
}

// JwtSigningMethod 根据配置返回对应的 HMAC 签名方法。
// 若 SigningMethod 未设置或无效，默认返回 HS256。
func (c *Config) JwtSigningMethod() *jwtv5.SigningMethodHMAC {
	switch c.SigningMethod {
	case SigningMethodHS384:
		return jwtv5.SigningMethodHS384
	case SigningMethodHS512:
		return jwtv5.SigningMethodHS512
	default:
		return jwtv5.SigningMethodHS256
	}
}

// CustomClaims 自定义 JWT 声明，内嵌 jwt.RegisteredClaims 以支持标准字段（exp、iat 等）。
// 可根据业务需要在结构体中添加自定义字段。
type CustomClaims struct {
	// Uid 用户 ID，示例自定义字段。
	Uid int `json:"uid"`
	jwtv5.RegisteredClaims
}

// GenToken 使用指定配置生成 JWT 令牌字符串。
// claims 必须实现 jwt.Claims 接口，通常传入 *CustomClaims。
func GenToken(config *Config, claims jwtv5.Claims) (string, error) {
	token := jwtv5.NewWithClaims(config.JwtSigningMethod(), claims)
	return token.SignedString([]byte(config.SecretKey))
}

// ParseToken 解析并验证 JWT 令牌字符串。
// tokenString 为原始 JWT 字符串，claims 为接收解析结果的结构体指针（如 &CustomClaims{}）。
// 返回值可通过 errors.Is 判断具体错误类型：
//   - ErrTokenExpired：token 已过期
//   - ErrTokenInvalid：token 无效（签名错误、格式错误等），可通过 errors.Unwrap 获取原始错误
func ParseToken(config *Config, tokenString string, claims jwtv5.Claims) (*jwtv5.Token, error) {
	token, err := jwtv5.ParseWithClaims(tokenString, claims, func(token *jwtv5.Token) (interface{}, error) {
		return []byte(config.SecretKey), nil
	})
	if err != nil {
		if errors.Is(err, jwtv5.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %w", ErrTokenInvalid, err)
	}
	if token != nil && token.Valid {
		return token, nil
	}
	return nil, ErrTokenInvalid
}
