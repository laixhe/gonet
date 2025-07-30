package jwt

import (
	"errors"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

/*
jwt:
  # 密钥
  secret_key: 6Kbj0VFeXYMp60lEyiFoVq4UzqX8Z0GSSfnvTh2VuAQn0oHgQNYexU6yYVTk4xf9
  # 过期时长(单位秒)
  expire_time: 604800
  # 签名方法(签名算法) HS256 HS384 HS512
  signing_method: HS256
*/

var (
	ErrTokenExpired = errors.New("token is expired") // 令牌已过期
	ErrTokenInvalid = errors.New("token invalid")    // 令牌无效
)

// JWT 认证头部
const (
	Authorization = "Authorization"
	Bearer        = "Bearer "
	BearerLen     = 7
)

// JWT 签名方法(签名算法)
const (
	SigningMethodHS256 = "HS256"
	SigningMethodHS384 = "HS384"
	SigningMethodHS512 = "HS512"
)

// JWT配置
type Config struct {
	// 密钥
	SecretKey string `json:"secret_key" mapstructure:"secret_key" toml:"secret_key" yaml:"secret_key"`
	// 过期时长(单位秒)
	ExpireTime int64 `json:"expire_time" mapstructure:"expire_time" toml:"expire_time" yaml:"expire_time"`
	// 签名方法(签名算法) HS256 HS384 HS512
	SigningMethod string `json:"signing_method" mapstructure:"signing_method" toml:"signing_method" yaml:"signing_method"`
}

// JwtSigningMethod 获取JWT签名方法
func (c *Config) JwtSigningMethod() *jwtv5.SigningMethodHMAC {
	switch c.SigningMethod {
	case SigningMethodHS256:
		return jwtv5.SigningMethodHS256
	case SigningMethodHS384:
		return jwtv5.SigningMethodHS384
	case SigningMethodHS512:
		return jwtv5.SigningMethodHS512
	default:
		c.SigningMethod = SigningMethodHS256
		return jwtv5.SigningMethodHS256
	}
}

// Checking 检查
func (c *Config) Check() error {
	if c == nil {
		return errors.New("没有JWT配置")
	}
	if c.SecretKey == "" {
		return errors.New("没有JWT密钥配置")
	}
	switch c.SigningMethod {
	case SigningMethodHS256, SigningMethodHS384, SigningMethodHS512:
		break
	default:
		c.SigningMethod = SigningMethodHS256
	}
	return nil
}

// 自定义声明类型
// 内嵌 jwt.RegisteredClaims
// jwt 包自带的 jwt.RegisteredClaims 只包含了官方字段
type CustomClaims struct {
	// 可根据需要自行添加字段
	Uid int `json:"uid"`
	jwtv5.RegisteredClaims
}

func (c CustomClaims) GetUid() int {
	return c.Uid
}

// GenToken 生成JWT
func GenToken(config *Config, claims jwtv5.Claims) (string, error) {
	// 使用指定的签名方法创建签名对象（使用签名算法）
	token := jwtv5.NewWithClaims(config.JwtSigningMethod(), claims)
	// 使用指定的 secret 签名并获得完整的编码后的字符串 token
	return token.SignedString([]byte(config.SecretKey))
}

// ParseToken 解析JWT
func ParseToken(config *Config, tokenString string, claims jwtv5.Claims) (*jwtv5.Token, error) {
	// 如果是自定义 Claim 结构体则需要使用 ParseWithClaims 方法
	token, err := jwtv5.ParseWithClaims(tokenString, claims, func(token *jwtv5.Token) (i interface{}, err error) {
		return []byte(config.SecretKey), nil
	})
	if err != nil {
		return nil, ErrTokenInvalid
	}
	if token != nil && token.Valid {
		return token, nil
	}
	return nil, ErrTokenInvalid
}
